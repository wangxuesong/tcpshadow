package main

import (
	"fmt"
	"net"
	"testing"

	"github.com/wangxuesong/tcpshadow/cmd"
	"github.com/wangxuesong/tcpshadow/model"

	"github.com/stretchr/testify/assert"
)

func TestServer(t *testing.T) {
	datas := cmd.ReadPackages()
	// count := len(datas)
	// index := 0
	// nextServerResponse := func(index int) SavePackage {
	// 	for i := index; i < count; i++ {
	// 		if datas[i].Forward == ServerToClient {
	// 			return datas[i]
	// 		}
	// 	}
	// 	panic("abc")
	// }
	// nextClientResponse := func(index int) SavePackage {
	// 	for i := index; i < count; i++ {
	// 		if datas[i].Forward == ClientToServer {
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
		switch d.Forward {
		case model.ClientToServer:
			buf := make([]byte, 16384)
			cnt, err := client.Read(buf)
			if err != nil {
				panic(err)
			}
			if d.Number == 0 {
				continue
			}
			assert.ElementsMatch(t, d.Buffer, buf[:cnt], func() string {
				return fmt.Sprintf("acutaled: %#v\n", d.Buffer)
			}())
		case model.ServerToClient:
			client.Write(d.Buffer)
		}
	}

}
