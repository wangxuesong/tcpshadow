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
			Length: 313,
			Context: []model.Context{{
				Noname1:   2,
				Noname2:   0,
				Noname222: 0,
				Noname3:   100,
				Noname4:   101,
				Noname5:   0,
				Noname6:   6,
				Noname7:   108,
				Noname8:   "000000000000",
				Noname9:   32,
				Noname10:  "GBase server00000000000000.FC4G1",
				Noname11:  35,
				Noname12:  "SOftware serial number 0000000 0000",
				Noname13:  18,
				Noname14:  "Ol_gbasedbt1210_8 mber AAA#B000000",
				Noname15:  316,
				Noname16:  0,
				Noname17:  0,
				Noname18:  0,
				Noname19:  0,
				Noname20:  0,
				Noname21:  "000000000000000000000000",
				Noname22:  102,
				Noname23:  "000000",
				Noname24:  0,
				Noname25:  0,
				Noname26:  0,
				Noname27:  0,
			}},
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
		buffer, err := (&model.SqliDescribe{
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
		}).Pack()
		assert.Nil(t, err)
		buffer, err = (&model.SqliDone{
			Warning:  0,
			Rows:     0,
			RowID:    0,
			SerialID: 0,
		}).Pack()
		assert.Nil(t, err)
		buffer, err = (&model.SqliCost{
			EstimatedRows: 1,
			EstimatedIO:   2,
		}).Pack()
		assert.Nil(t, err)
		buffer, err = (&model.SqliEot{}).Pack()
		assert.Nil(t, err)
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
