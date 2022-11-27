package supervisor

import (
	"log"
	"sync"

	"github.com/wangxuesong/tcpshadow/pkg/services"
)

type (
	Supervisor struct {
		config         *Config
		serviceBuilder services.ServiceBuilder

		listener   *services.Listener
		done       chan struct{}
		listenChan chan services.ListenChannel
		deleteChan chan string
		wg         sync.WaitGroup
		services   map[string]services.Service
		output     *services.OutputService
	}

	Config struct {
		ListenAddress string
		ServerAddress string

		OutputFile   string
		ProtocolType string

		IsPrintPackage bool
	}
)

func NewSupervisor(config *Config, builder services.ServiceBuilder) *Supervisor {
	return &Supervisor{
		config:         config,
		serviceBuilder: builder,
		done:           make(chan struct{}),
		listenChan:     make(chan services.ListenChannel),
		deleteChan:     make(chan string),
		services:       make(map[string]services.Service),
	}
}

func (s *Supervisor) Close(wg *sync.WaitGroup) {
	defer wg.Done()

	// 关闭 Listener
	wg.Add(1)
	s.listener.Close(wg)

	// 关闭 proxy
	wg.Add(len(s.services))
	for _, proxy := range s.services {
		proxy.Close(wg)
	}

	// 关闭 output
	wg.Add(1)
	s.output.Close(wg)

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
		index := 0
		monitor := make(chan *services.Context)
		config := services.OutputConfig{
			Monitor:        monitor,
			Outputs:        []services.OutputType{services.Console, services.File},
			Filename:       s.config.OutputFile,
			ProtocolType:   s.config.ProtocolType,
			IsPrintPackage: s.config.IsPrintPackage,
		}
		s.output = services.NewOutputService(config)
		go s.output.Run()
		for {
			select {
			case <-s.done:
				log.Println("stopping supervisor")
				return
			case conn := <-s.listenChan:
				clientId := conn.Conn().RemoteAddr().String()
				config := &services.ProxyConfig{
					ClientId:      clientId,
					DeleteChan:    s.deleteChan,
					Front:         conn.Conn(),
					ServerAddress: s.config.ServerAddress,
					ProtocolType:  s.config.ProtocolType,
					Monitor:       monitor,
				}
				s.services[clientId] = s.serviceBuilder(config, index)
				index += 1
				err := s.services[clientId].Run()
				if err != nil {
					log.Println(err)
					s.deleteClient(clientId)
				}
			case id := <-s.deleteChan:
				s.deleteClient(id)
			default:
			}
		}
	}()
}

func (s *Supervisor) deleteClient(clientId string) {
	if service, ok := s.services[clientId]; ok {
		delete(s.services, clientId)
		go func() {
			wg := sync.WaitGroup{}
			wg.Add(1)
			service.Close(&wg)
		}()
	}
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
