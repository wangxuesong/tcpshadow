package services

import (
	"github.com/wangxuesong/tcpshadow/model"
	"io"
	"log"
	"net"
	"sync"
	"time"
)

type (
	ProxyService struct {
		front   net.Conn
		backend net.Conn
		done    chan struct{}
		wg      sync.WaitGroup

		config *ProxyConfig
	}

	ProxyConfig struct {
		Front         net.Conn
		ServerAddress string

		ProtocolType string
		Monitor      chan model.Data
	}
)

func NewProxyService(config *ProxyConfig) *ProxyService {
	return &ProxyService{
		front:  config.Front,
		done:   make(chan struct{}),
		config: config,
	}
}

func (s *ProxyService) Run() error {
	defer s.wg.Done()
	s.wg.Add(1)

	server, err := net.Dial("tcp", s.config.ServerAddress)
	if err != nil {
		return err
	}
	log.Println("Success connected to the server:", s.config.ServerAddress)
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

func (s *ProxyService) proxyData(src net.Conn, dest net.Conn, forward model.DataForward, monitor chan model.Data) {
	defer src.Close()
	defer s.wg.Done()
	s.wg.Add(1)
	for {
		select {
		case <-s.done:
			log.Println("disconnecting", src.RemoteAddr())
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
				log.Println("disconnecting", src.RemoteAddr())
				//s.done <- true
				return
			}
			log.Println(err)
			return
		}
		go func() {
			//data := model.Data{Forward: forward, Buffer: buf[:cnt]}
			//monitor <- data
		}()

		if _, err := dest.Write(buf[:cnt]); nil != err {
			log.Println(err)
			return
		}
	}
}
