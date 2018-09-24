package main

import (
	"fmt"
	capture2 "github.com/wangxuesong/tcpshadow/model"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestServer(t *testing.T) {
	datas := readPackages()
	// count := len(datas)
	// index := 0
	// nextServerResponse := func(index int) SavePackage {
	// 	for i := index; i < count; i++ {
	// 		if datas[i].Forware == ServerToClient {
	// 			return datas[i]
	// 		}
	// 	}
	// 	panic("abc")
	// }
	// nextClientResponse := func(index int) SavePackage {
	// 	for i := index; i < count; i++ {
	// 		if datas[i].Forware == ClientToServer {
	// 			return datas[i]
	// 		}
	// 	}
	// 	panic("abc")
	// }

	clientConn, err := net.Listen("tcp4", ":9088")
	if err != nil {
		panic(err)
	}
	client, err := clientConn.Accept()
	if err != nil {
		panic(err)
	}

	for _, d := range datas {
		switch d.Forware {
		case capture2.ClientToServer:
			buf := make([]byte, 16384)
			cnt, err := client.Read(buf)
			if err != nil {
				panic(err)
			}
			if d.number == 0 {
				continue
			}
			assert.ElementsMatch(t, d.Buffer, buf[:cnt], func() string {
				return fmt.Sprintf("acutaled: %#v\n", d.Buffer)
			}())
		case capture2.ServerToClient:
			client.Write(d.Buffer)
		}
	}

}
