package bridge

import (
	"net"
	"sync"
	"testing"

	pgproto "github.com/jackc/pgproto3"
	"github.com/stretchr/testify/assert"
	"github.com/wangxuesong/tcpshadow/model"
	"github.com/wangxuesong/tcpshadow/pkg/services"
)

func TestConnect(t *testing.T) {
	address := "127.0.0.1:11030"
	clientConn, err := net.ResolveTCPAddr("tcp4", address)
	assert.Nil(t, err)
	listener, err := net.ListenTCP("tcp4", clientConn)
	assert.Nil(t, err)

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
	defer listener.Close()

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

	// 8s Server
	{
		conn, err := listener.AcceptTCP()
		assert.Nil(t, err)
		//TODO: 增加发送给 8s 登录包之后取消以下注释
		buf := make([]byte, 1024)
		c, err := conn.Read(buf)
		assert.Nil(t, err)
		assert.True(t, c > 0)

		// send auth ack
		conn.Write([]byte{88, 11, 03}) //TODO: 将临时数据替换成正式数据
	}

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

func TestConnectFilter_Handle(t *testing.T) {
	pgFront, pgBackend := net.Pipe()
	gbFront, gbBackend := net.Pipe()
	filter := NewConnectFilter()
	ctx := &Context{
		sessionId: 0,
		front:     pgBackend,
		backend:   gbFront,
		state:     ConnectState,
	}
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
	ctx.SetData(&model.Data{
		Forward: model.ClientToServer,
		Buffer:  buf,
	})

	// 8s Server
	go func() {
		buf := make([]byte, 1024)
		gbBackend.Read(buf)
	}()

	go func() {
		err := filter.Handle(ctx)
		assert.Nil(t, err)
		ctx.SetData(&model.Data{
			Forward: model.ServerToClient,
			Buffer:  buf,
		})
		err = filter.Handle(ctx)
		assert.Nil(t, err)
	}()

	{
		parse, err := pgproto.NewFrontend(pgproto.NewChunkReader(pgFront), nil)
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

func TestQueryFilter_Handle(t *testing.T) {
	pgFront, pgBackend := net.Pipe()
	gbFront, gbBackend := net.Pipe()
	filter := NewQueryFilter()
	ctx := &Context{
		sessionId: 0,
		front:     pgBackend,
		backend:   gbFront,
		state:     QueryState,
	}
	buffer := (&pgproto.Parse{
		Name:          "",
		Query:         "selet * from test",
		ParameterOIDs: nil,
	}).Encode(nil)
	buffer = (&pgproto.Bind{
		DestinationPortal:    "",
		PreparedStatement:    "",
		ParameterFormatCodes: nil,
		Parameters:           nil,
		ResultFormatCodes:    nil,
	}).Encode(buffer)
	buffer = (&pgproto.Describe{
		ObjectType: 'P',
		Name:       "",
	}).Encode(buffer)
	buffer = (&pgproto.Execute{
		Portal:  "",
		MaxRows: 0,
	}).Encode(buffer)
	buffer = (&pgproto.Sync{}).Encode(buffer)
	ctx.SetData(&model.Data{
		Forward: model.ClientToServer,
		Buffer:  buffer,
	})
	// 8s Server
	go func() {
		buf := make([]byte, 1024)
		gbBackend.Read(buf)
	}()

	go func() {
		err := filter.Handle(ctx)
		assert.Nil(t, err)

		// gbase response
		//TODO: 将 buffer 改成正式的 sqli 数据
		ctx.SetData(&model.Data{
			Forward: model.ServerToClient,
			Buffer:  buffer,
		})
		err = filter.Handle(ctx)
		assert.Nil(t, err)
	}()

	{
		parse, err := pgproto.NewFrontend(pgproto.NewChunkReader(pgFront), nil)
		assert.Nil(t, err)

		msg, err := parse.Receive()
		assert.Nil(t, err)
		assert.IsType(t, &pgproto.ParseComplete{}, msg)
		msg, err = parse.Receive()
		assert.Nil(t, err)
		assert.IsType(t, &pgproto.BindComplete{}, msg)
		msg, err = parse.Receive()
		assert.Nil(t, err)
		assert.IsType(t, &pgproto.RowDescription{}, msg)
		msg, err = parse.Receive()
		assert.Nil(t, err)
		assert.IsType(t, &pgproto.DataRow{}, msg)
		msg, err = parse.Receive()
		assert.Nil(t, err)
		assert.IsType(t, &pgproto.CommandComplete{}, msg)
		msg, err = parse.Receive()
		assert.Nil(t, err)
		assert.IsType(t, &pgproto.ReadyForQuery{}, msg)
	}
}
