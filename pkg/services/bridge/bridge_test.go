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
		response, err := (&model.AuthResponse{
			Length:           287,
			Noname1:          2,
			Noname2:          15376,
			Noname3:          0,
			Noname4:          100,
			Noname5:          101,
			Noname6:          61,
			IEEEIlength:      6,
			IEEEI:            "IEEEI",
			Noname7:          108,
			Srvinfx:          "srvinfx",
			Versionlength:    34,
			Version:          "GBase Server Version 9.56.FC4G1TL",
			Softwarelength:   35,
			Software:         "Software Serial Number AAA#B000000",
			Clientnamelength: 12,
			Clientname:       "gbaseserver",
			Noname8:          316,
			Noname9:          0,
			Noname10:         0,
			Noname11:         0,
			Noname12:         0,
			Noname13:         0,
			Noname14:         "on",
			Noname15:         "=soctcp",
			Noname16:         102,
			Noname17:         0,
			Noname18:         0,
			Noname19:         20,
			Noname20:         0,
			Noname21:         107,
			Noname22:         2958,
			Noname23:         872,
			Noname24:         13312,
			Path1length:      11,
			Path1:            "/dev/pts/0",
			Path2length:      15,
			Path2:            "/home/gbasedbt",
			Noname25:         110,
			Noname26:         4,
			Noname27:         0,
			Noname28:         0,
			Noname29:         116,
			Noname30:         43,
			Noname31:         0,
			Noname32:         1001,
			Noname33:         0,
			Noname34:         1001,
			Path3length:      33,
			Path3:            "/home/zhangyaru/gbase/bin/oninit",
			Asceot:           127,
		}).Pack()
		conn.Write(response) //TODO: 将临时数据替换成正式数据
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
		metadata:  make(map[string]interface{}),
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

	server_passed := false
	// 8s Server
	go func() {
		defer func() { server_passed = true }()
		buf := make([]byte, 1024)
		// read AuthRequest
		c, err := gbBackend.Read(buf)
		assert.Nil(t, err)
		assert.True(t, c > 0)
		// read SqliProtocols
		//c, err = gbBackend.Read(buf)
		//assert.Nil(t, err)
		//assert.True(t, c > 0)
	}()

	go func() {
		err := filter.Handle(ctx)
		assert.Nil(t, err)
		// AuthResponse
		ctx.SetData(&model.Data{
			Forward: model.ServerToClient,
			Buffer:  buf,
		})
		err = filter.Handle(ctx)
		assert.Nil(t, err)
		// SqliProtocols
		ctx.SetData(&model.Data{
			Forward: model.ServerToClient,
			Buffer:  buf,
		})
		err = filter.Handle(ctx)
		assert.Nil(t, err)
		// SqliInfo
		ctx.SetData(&model.Data{
			Forward: model.ServerToClient,
			Buffer:  buf,
		})
		err = filter.Handle(ctx)
		assert.Nil(t, err)
		// SqliDBOpen
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

	assert.True(t, server_passed)
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
		c, err := gbBackend.Read(buf)
		assert.Nil(t, err)
		assert.True(t, c > 0)
		//readseeker := bytes.NewReader(buf)
		//msgs, err := model.UnpackSqliTransmission(readseeker)
		//assert.Nil(t, nil)
		//assert.IsType(t, &model.SqliPrepare{}, msgs[1:2])
		//assert.IsType(t, &model.SqliNDescribe{}, msgs[2:3])
		//assert.IsType(t, &model.SqliWantDone{}, msgs[3:4])
		//assert.IsType(t, &model.SqliEot{}, msgs[4:5])
	}()

	go func() {
		err := filter.Handle(ctx)
		assert.Nil(t, err)

		// gbase response
		describe := &model.SqliDescribe{
			StatementType: 2,
			StatementID:   0,
			EstimatedCost: 0,
			TupleSize:     8,
			CountOfFields: 2,
			StringTable:   8,
			Fields: []model.SqliField{{
				FieldIndex:              0,
				ColumnStartPos:          0,
				ColumnType:              2,
				ColumnExtendedBuiltinId: 0,
				OwnerName:               "",
				ExtendedName:            "",
				Reference:               0,
				Alignment:               0,
				SourceType:              0,
				Length:                  4,
				Name:                    "id",
			}, {
				FieldIndex:              3,
				ColumnStartPos:          4,
				ColumnType:              2,
				ColumnExtendedBuiltinId: 0,
				OwnerName:               "",
				ExtendedName:            "",
				Reference:               0,
				Alignment:               0,
				SourceType:              0,
				Length:                  4,
				Name:                    "code",
			},
			},
		}
		assert.Nil(t, err)
		done := &model.SqliDone{
			Warning:  0,
			Rows:     0,
			RowID:    0,
			SerialID: 0,
		}
		assert.Nil(t, err)
		cost := &model.SqliCost{
			EstimatedRows: 1,
			EstimatedIO:   2,
		}
		assert.Nil(t, err)
		eot := &model.SqliEot{}
		assert.Nil(t, err)
		var transmission model.SqliTransmission
		transmission = []model.SqliCommand{describe, done, cost, eot}
		buffer, err := transmission.Pack()
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
