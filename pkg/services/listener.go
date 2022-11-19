package services

import (
	"log"
	"net"
	"sync"
	"time"
)

type (
	Listener struct {
		address    string
		done       chan struct{}
		listenChan chan ListenChannel
		wg         sync.WaitGroup
	}

	ListenChannel interface {
		Address() string
		Conn() net.Conn
	}

	listenChannel struct {
		address string
		conn    net.Conn
	}
)

func NewListener(address string, listenChan chan ListenChannel) *Listener {
	return &Listener{
		address:    address,
		done:       make(chan struct{}),
		listenChan: listenChan,
	}
}

func (l *Listener) Run() error {
	log.Println("Start....")
	clientConn, err := net.ResolveTCPAddr("tcp4", l.address)
	if nil != err {
		return err
	}
	listener, err := net.ListenTCP("tcp4", clientConn)
	if nil != err {
		return err
	}
	log.Println("listening on", listener.Addr())

	l.wg.Add(1)
	go func() {
		defer l.wg.Done()
		for {
			select {
			case <-l.done:
				log.Println("stopping listening on", listener.Addr())
				_ = listener.Close()
				return
			default:
			}

			_ = listener.SetDeadline(time.Now().Add(1e9))
			conn, err := listener.AcceptTCP()
			if err != nil {
				if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
					continue
				}
				log.Println(err)
			}
			l.listenChan <- &listenChannel{address: l.address, conn: conn}
		}
	}()
	return nil
}

func (l *Listener) Close(wg *sync.WaitGroup) {
	defer wg.Done()
	close(l.done)
	l.wg.Wait()
}

func (c *listenChannel) Address() string {
	return c.address
}
func (c *listenChannel) Conn() net.Conn {
	return c.conn
}
