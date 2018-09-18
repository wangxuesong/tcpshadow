package main

import (
	"fmt"
	"net"
)

func main() {
	fmt.Println("Start....")

	clientConn, err := net.Listen("tcp4", ":9088")
	if err != nil {
		panic(err)
	}

	for {
		conn, err := clientConn.Accept()
		if err != nil {
			panic(err)
		}

		go connectServer(conn)
	}
}
func connectServer(client net.Conn) {
	if client == nil {
		return
	}

	fmt.Println("Connect to server")

	server, err := net.Dial("tcp4", "info:9088")
	if err != nil {
		panic(err)
	}

	go readFromServer(client, server)

	buf := make([]byte, 16384)
	for {
		// fmt.Println("Read from client")
		cnt, err := client.Read(buf)
		if err != nil {
			panic(err)
		}
		fmt.Printf("C->S: %#v\n", buf[:cnt])
		server.Write(buf[:cnt])
	}
}
func readFromServer(client net.Conn, server net.Conn) {
	buf := make([]byte, 16384)
	for {
		// fmt.Println("Read from server")
		cnt, err := server.Read(buf)
		if err != nil {
			panic(err)
		}
		fmt.Printf("S->C: %#v\n", buf[:cnt])
		client.Write(buf[:cnt])
	}
}
