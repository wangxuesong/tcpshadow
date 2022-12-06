package bridge

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"

	pgproto "github.com/jackc/pgproto3"
	"github.com/wangxuesong/tcpshadow/model"
	"github.com/wangxuesong/tcpshadow/pkg/services"
)

type (
	BridgeService struct {
		index   int
		front   net.Conn
		backend net.Conn
		done    chan struct{}
		wg      sync.WaitGroup

		currentCtx *Context

		filters map[SessionState]Handler

		config *services.ProxyConfig
	}

	Handler interface {
		Handle(ctx services.Context) error
	}

	SessionState string

	Context struct {
		data      *model.Data
		sessionId int
		front     net.Conn
		backend   net.Conn
		state     SessionState
		requests  model.PgTransmission
		responses model.PgTransmission
	}

	ConnectFilter struct {
		//bridge *BridgeService
	}
)

const (
	ConnectState SessionState = "Connect"
	QueryState   SessionState = "Query"
)

func NewBridgeService(config *services.ProxyConfig, index int) services.Service {
	b := &BridgeService{
		index:   index,
		front:   config.Front,
		done:    make(chan struct{}),
		filters: make(map[SessionState]Handler),
		config:  config,
	}
	b.filters[ConnectState] = NewConnectFilter()
	return b
}

func (b *BridgeService) Run() error {
	defer b.wg.Done()
	b.wg.Add(1)

	log.Printf("[%d] %s %s\n", b.index, b.front.RemoteAddr(), "connected")

	var err error
	b.backend, err = net.Dial("tcp", b.config.ServerAddress)
	if err != nil {
		b.front.Close()
		return err
	}
	log.Printf("[%d]Success connected to the server: %s\n", b.index, b.config.ServerAddress)

	//err = b.handleConn()
	//if err != nil {
	//	return err
	//}

	ctx := &Context{
		sessionId: b.index,
		front:     b.front,
		backend:   b.backend,
		state:     ConnectState,
		requests:  nil,
		responses: nil,
	}
	b.currentCtx = ctx

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
			select {
			case <-b.done:
				log.Printf("[%d]disconnecting %s\n", b.index, b.front.RemoteAddr())
				return
			default:
			}
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

	go b.frontReadLoop(frontReadChan)
	return nil
}

func (b *BridgeService) Close(wg *sync.WaitGroup) {
	defer wg.Done()
	close(b.done)
	b.wg.Wait()
}

func (b *BridgeService) handleConn(ctx services.Context) error {
	buff := bytes.NewBuffer(ctx.Data().Buffer)
	backend, err := pgproto.NewBackend(pgproto.NewChunkReader(buff), nil)
	if err != nil {
		return err
	}

	msg, err := backend.ReceiveStartupMessage()
	if err != nil {
		return err
	}

	context, ok := ctx.(*Context)
	if !ok {
		return fmt.Errorf("unknown context type: %T", ctx)
	}

	context.requests = []model.PgCommand{msg}

	_ = msg.Parameters["user"]

	auth := &pgproto.Authentication{
		Type:               pgproto.AuthTypeOk,
		Salt:               [4]byte{0},
		SASLAuthMechanisms: nil,
		SASLData:           nil,
	}
	status := &pgproto.ParameterStatus{Name: "server_version", Value: "9.5"}
	key := &pgproto.BackendKeyData{
		ProcessID: 881103,
		SecretKey: 1569992916,
	}
	ready := &pgproto.ReadyForQuery{
		TxStatus: 73,
	}

	context.responses = []model.PgCommand{auth, status, key, ready}
	buf, err := context.responses.Pack()
	if err != nil {
		return err
	}

	_, err = b.front.Write(buf)
	if err != nil {
		return err
	}

	return nil
}

func (b *BridgeService) frontReadLoop(readChan chan []byte) {
	for {
		select {
		case <-b.done:
			return
		case buf := <-readChan:
			data := &model.Data{
				Forward: model.ClientToServer,
				Buffer:  buf,
			}
			b.currentCtx.data = data
			err := b.filters[b.currentCtx.state].Handle(b.currentCtx)
			//err := b.handleConn(b.currentCtx)
			if err != nil {
				return
			}
		}
	}
}

func (c *Context) Data() *model.Data {
	return c.data
}

func (c *Context) SetData(d *model.Data) error {
	c.data = d
	return nil
}

func (c *Context) SessionId() int {
	return c.sessionId
}

func (c *Context) SetSessionId(id int) error {
	c.sessionId = id
	return nil
}

func NewConnectFilter() *ConnectFilter {
	return &ConnectFilter{}
}

func (c *ConnectFilter) Handle(ctx services.Context) error {
	//return c.bridge.handleConn(ctx)
	buff := bytes.NewBuffer(ctx.Data().Buffer)
	backend, err := pgproto.NewBackend(pgproto.NewChunkReader(buff), nil)
	if err != nil {
		return err
	}

	msg, err := backend.ReceiveStartupMessage()
	if err != nil {
		return err
	}

	context, ok := ctx.(*Context)
	if !ok {
		return fmt.Errorf("unknown context type: %T", ctx)
	}

	context.requests = []model.PgCommand{msg}

	_ = msg.Parameters["user"]

	auth := &pgproto.Authentication{
		Type:               pgproto.AuthTypeOk,
		Salt:               [4]byte{0},
		SASLAuthMechanisms: nil,
		SASLData:           nil,
	}
	status := &pgproto.ParameterStatus{Name: "server_version", Value: "9.5"}
	key := &pgproto.BackendKeyData{
		ProcessID: 881103,
		SecretKey: 1569992916,
	}
	ready := &pgproto.ReadyForQuery{
		TxStatus: 73,
	}

	context.responses = []model.PgCommand{auth, status, key, ready}
	buf, err := context.responses.Pack()
	if err != nil {
		return err
	}

	_, err = context.front.Write(buf)
	if err != nil {
		return err
	}

	context.state = QueryState
	return nil
}
