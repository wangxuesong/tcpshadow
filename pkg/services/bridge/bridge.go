package bridge

import (
	pgproto "github.com/jackc/pgproto3"
	"io"
	"log"
	"net"
	"sync"
	"time"

	"github.com/wangxuesong/tcpshadow/pkg/services"
)

type (
	BridgeService struct {
		index   int
		front   net.Conn
		backend net.Conn
		done    chan struct{}
		wg      sync.WaitGroup

		config *services.ProxyConfig
	}

	Handler interface {
		Handle(ctx *services.Context) error
	}
)

func NewBridgeService(config *services.ProxyConfig, index int) services.Service {
	return &BridgeService{
		index:  index,
		front:  config.Front,
		done:   make(chan struct{}),
		config: config,
	}
}

func (b *BridgeService) Run() error {
	defer b.wg.Done()
	b.wg.Add(1)

	log.Printf("[%d] %s %s\n", b.index, b.front.RemoteAddr(), "connected")

	var err error
	b.backend, err = net.Dial("tcp", b.config.ServerAddress)
	if err != nil {
		return err
	}
	log.Printf("[%d]Success connected to the server: %s\n", b.index, b.config.ServerAddress)

	err = b.handleConn()
	if err != nil {
		return err
	}

	backendReadChan := make(chan []byte)
	backendSendChan := make(chan []byte)
	frontReadChan := make(chan []byte)
	frontSendChan := make(chan []byte)

	// backend receiver
	go func() {
		for {
			select {
			case <-b.done:
				log.Printf("[%d]disconnecting %s\n", b.index, b.backend.RemoteAddr())
				return
			default:
			}
			_ = b.backend.SetDeadline(time.Now().Add(1e9))
			buf := make([]byte, 16384)
			cnt, err := b.backend.Read(buf)
			if nil != err {
				if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
					continue
				}
				if err == io.EOF {
					log.Printf("[%d]disconnecting %s\n", b.index, b.backend.RemoteAddr())
					b.config.DeleteChan <- b.config.ClientId
					return
				}
				log.Println(err)
			}
			backendReadChan <- buf[:cnt]
		}
	}()

	// backend sender
	go func() {
		for {
			select {
			case <-b.done:
				return
			case buf := <-backendSendChan:
				b.backend.Write(buf)
			}
		}
	}()

	// front receiver
	go func() {
		for {
			_ = b.front.SetDeadline(time.Now().Add(1e9))
			buf := make([]byte, 16384)
			cnt, err := b.front.Read(buf)
			if nil != err {
				if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
					continue
				}
				if err == io.EOF {
					log.Printf("[%d]disconnecting %s\n", b.index, b.front.RemoteAddr())
					b.config.DeleteChan <- b.config.ClientId
					return
				}
				log.Println(err)
			}
			frontReadChan <- buf[:cnt]
		}
	}()

	// front sender
	go func() {
		for {
			select {
			case <-b.done:
				return
			case buf := <-frontSendChan:
				b.front.Write(buf)
			}
		}
	}()

	return nil
}

func (b *BridgeService) Close(wg *sync.WaitGroup) {
	defer wg.Done()
	close(b.done)
}

func (b *BridgeService) handleConn() error {
	backend, err := pgproto.NewBackend(pgproto.NewChunkReader(b.front), b.front)
	if err != nil {
		return err
	}

	msg, err := backend.ReceiveStartupMessage()
	if err != nil {
		return err
	}

	_ = msg.Parameters["user"]

	auth := &pgproto.Authentication{
		Type:               pgproto.AuthTypeOk,
		Salt:               [4]byte{0},
		SASLAuthMechanisms: nil,
		SASLData:           nil,
	}
	buf := auth.Encode(nil)

	status := &pgproto.ParameterStatus{Name: "server_version", Value: "9.5"}
	buf = status.Encode(buf)

	key := &pgproto.BackendKeyData{
		ProcessID: 881103,
		SecretKey: 1569992916,
	}
	buf = key.Encode(buf)

	ready := &pgproto.ReadyForQuery{
		TxStatus: 73,
	}
	buf = ready.Encode(buf)
	_, err = b.front.Write(buf)
	if err != nil {
		return err
	}

	return nil
}
