package supervisor

import (
	"log"
	"sync"

	"github.com/wangxuesong/tcpshadow/services"
)

type (
	Supervisor struct {
		config *Config

		listener   *services.Listener
		done       chan struct{}
		listenChan chan services.ListenChannel
		wg         sync.WaitGroup
		proxies    map[string]*services.ProxyService
	}

	Config struct {
		ListenAddress string
		ServerAddress string

		OutputFile   string
		ProtocolType string

		IsPrintPackage bool
	}
)

func NewSupervisor(config *Config) *Supervisor {
	return &Supervisor{
		config:     config,
		done:       make(chan struct{}),
		listenChan: make(chan services.ListenChannel),
		proxies:    make(map[string]*services.ProxyService),
	}
}

func (s *Supervisor) Close(wg *sync.WaitGroup) {
	defer wg.Done()

	// 关闭 Listener
	wg.Add(1)
	s.listener.Close(wg)

	// 关闭 proxy
	wg.Add(len(s.proxies))
	for _, proxy := range s.proxies {
		proxy.Close(wg)
	}

	// 关闭 Supervisor
	close(s.done)
	s.wg.Wait()
}

func (s *Supervisor) Serve() {
	s.listener = services.NewListener(s.config.ListenAddress, s.listenChan)
	s.run()
}

func (s *Supervisor) run() {
	// 启动 Listener
	err := s.listener.Run()
	if err != nil {
		log.Println(err)
		return
	}
	// 处理事件
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		for {
			select {
			case <-s.done:
				log.Println("stopping supervisor")
				return
			case conn := <-s.listenChan:
				client := conn.Conn().RemoteAddr().String()
				config := &services.ProxyConfig{
					Front:         conn.Conn(),
					ServerAddress: s.config.ServerAddress,
					ProtocolType:  s.config.ProtocolType,
				}
				s.proxies[client] = services.NewProxyService(config)
				err := s.proxies[client].Run()
				if err != nil {
					log.Println(err)
					wg := sync.WaitGroup{}
					s.proxies[client].Close(&wg)
					delete(s.proxies, client)
				}
			default:
			}
		}
	}()
}

func DefaultConfig() *Config {
	return &Config{
		ServerAddress: "",
		ProtocolType:  "sqli",
	}
}

func (c *Config) Validate() error {
	return nil
}
