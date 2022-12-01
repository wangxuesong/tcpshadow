package bridge

import (
	"net"
	"sync"
	"testing"

	pgproto "github.com/jackc/pgproto3"
	"github.com/stretchr/testify/assert"
	"github.com/wangxuesong/tcpshadow/pkg/services"
)

func TestConnect(t *testing.T) {
	address := "127.0.0.1:11030"
	clientConn, err := net.ResolveTCPAddr("tcp4", address)
	assert.Nil(t, err)
	listener, err := net.ListenTCP("tcp4", clientConn)
	assert.Nil(t, err)
	defer listener.Close()

	front, backend := net.Pipe()
	config := &services.ProxyConfig{
		ClientId:      "test",
		Front:         backend,
		ServerAddress: address,
		DeleteChan:    nil,
		ProtocolType:  "pg",
		Monitor:       nil,
	}
	bridge := NewBridgeService(config, 0)
	go func() {
		err = bridge.Run()
		assert.Nil(t, err)
	}()
	var wg sync.WaitGroup
	wg.Add(1)
	defer bridge.Close(&wg)

	msg := &pgproto.StartupMessage{
		ProtocolVersion: 196608,
		Parameters: map[string]string{
			"DateStyle":          "ISO",
			"TimeZone":           "Asia/Shanghai",
			"client_encoding":    "UTF8",
			"database":           "postgres",
			"extra_float_digits": "2",
			"user":               "postgres",
		},
	}
	buf := msg.Encode(nil)
	_, err = front.Write(buf)
	assert.Nil(t, err)

	{
		parse, err := pgproto.NewFrontend(pgproto.NewChunkReader(front), nil)
		assert.Nil(t, err)
		msg, err := parse.Receive()
		assert.Nil(t, err)
		assert.IsType(t, &pgproto.Authentication{}, msg)
		msg, err = parse.Receive()
		assert.Nil(t, err)
		assert.IsType(t, &pgproto.ParameterStatus{}, msg)
		msg, err = parse.Receive()
		assert.Nil(t, err)
		assert.IsType(t, &pgproto.BackendKeyData{}, msg)
		msg, err = parse.Receive()
		assert.Nil(t, err)
		assert.IsType(t, &pgproto.ReadyForQuery{}, msg)
	}
}
