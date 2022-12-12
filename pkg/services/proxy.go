package services

import (
	"io"
	"log"
	"net"
	"sync"
	"time"

	"github.com/wangxuesong/tcpshadow/model"
)

type (
	ProxyService struct {
		index   int
		front   net.Conn
		backend net.Conn
		done    chan struct{}
		wg      sync.WaitGroup

		config *ProxyConfig
	}
	ProxyConfig struct {
		ClientId      string
		Front         net.Conn
		ServerAddress string
		DeleteChan    chan string

		ProtocolType string
		Monitor      chan Context
	}
	ProxyContext struct {
		data      *model.Data
		sessionId int
	}
)

func NewProxyService(config *ProxyConfig, index int) Service {
	return &ProxyService{
		index:  index,
		front:  config.Front,
		done:   make(chan struct{}),
		config: config,
	}
}

func (s *ProxyService) Run() error {
	defer s.wg.Done()
	s.wg.Add(1)

	log.Printf("[%d] %s %s\n", s.index, s.front.RemoteAddr(), "connected")

	server, err := net.Dial("tcp", s.config.ServerAddress)
	if err != nil {
		s.front.Close()
		return err
	}
	log.Printf("[%d]Success connected to the server: %s\n", s.index, s.config.ServerAddress)
	client := s.front

	go s.proxyData(client, server, model.ClientToServer, s.config.Monitor)
	go s.proxyData(server, client, model.ServerToClient, s.config.Monitor)

	return nil
}

func (s *ProxyService) Close(wg *sync.WaitGroup) {
	defer wg.Done()
	close(s.done)
	s.wg.Wait()
}

func (s *ProxyService) proxyData(src net.Conn, dest net.Conn, forward model.DataForward, monitor chan Context) {
	defer src.Close()
	defer s.wg.Done()
	s.wg.Add(1)
	for {
		select {
		case <-s.done:
			log.Printf("[%d]disconnecting %s\n", s.index, src.RemoteAddr())
			return
		default:
		}
		_ = src.SetDeadline(time.Now().Add(1e9))
		buf := make([]byte, 16384)
		cnt, err := src.Read(buf)
		if nil != err {
			if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
				continue
			}
			if err == io.EOF {
				log.Printf("[%d]disconnecting %s\n", s.index, src.RemoteAddr())
				s.config.DeleteChan <- s.config.ClientId
				return
			}
			log.Println(err)
			return
		}
		go func() {
			data := model.Data{Forward: forward, Buffer: buf[:cnt]}
			context := NewProxyContext(s.index, &data)
			monitor <- context
		}()

		if _, err := dest.Write(buf[:cnt]); nil != err {
			log.Println(err)
			return
		}
	}
}

func NewProxyContext(id int, data *model.Data) Context {
	return &ProxyContext{
		data:      data,
		sessionId: id,
	}
}

func (p *ProxyContext) Data() *model.Data {
	return p.data
}

func (p *ProxyContext) SetData(d *model.Data) error {
	p.data = d
	return nil
}

func (p *ProxyContext) SessionId() int {
	return p.sessionId
}

func (p *ProxyContext) SetSessionId(id int) error {
	p.sessionId = id
	return nil
}

func (p *ProxyContext) MetaData(key string) (interface{}, error) {
	//TODO implement me
	panic("implement me")
}

func (p *ProxyContext) SetMetaData(key string, value interface{}) error {
	//TODO implement me
	panic("implement me")
}
