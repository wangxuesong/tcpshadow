package bridge

import (
	"bytes"
	pgproto "github.com/jackc/pgproto3"
	"github.com/stretchr/testify/assert"
	"github.com/wangxuesong/tcpshadow/model"
	"github.com/wangxuesong/tcpshadow/pkg/services"
	"net"
	"sync"
	"testing"
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
		//read SqliProtocols
		c, err = gbBackend.Read(buf)
		buff := buf[:c]
		re := bytes.NewReader(buff)
		msgs, err := model.UnpackSqliTransmission(re)
		assert.Nil(t, err)
		assert.IsType(t, &model.SqliProtocols{}, msgs[0])
		assert.IsType(t, &model.SqliEot{}, msgs[1])
		assert.Nil(t, err)
		assert.True(t, c > 0)
		//read SqliInfo
		c, err = gbBackend.Read(buf)
		buff = buf[:c]
		re = bytes.NewReader(buff)
		msgs, err = model.UnpackSqliTransmission(re)
		assert.Nil(t, err)
		assert.IsType(t, &model.SqliInfo{}, msgs[0])
		assert.IsType(t, &model.SqliEot{}, msgs[1])
		assert.Nil(t, err)
		assert.True(t, c > 0)
		//read SqliDBOpen
		c, err = gbBackend.Read(buf)
		assert.Nil(t, err)
		assert.True(t, c > 0)
		buff = buf[:c]
		re = bytes.NewReader(buff)
		msgs, err = model.UnpackSqliTransmission(re)
		assert.Nil(t, err)
		assert.IsType(t, &model.SqliDBOpen{}, msgs[0])
		assert.IsType(t, &model.SqliEot{}, msgs[1])
	}()

	go func() {
		err := filter.Handle(ctx)
		assert.Nil(t, err)
		// AuthResponse
		buff, err := (&model.AuthResponse{
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
		assert.Nil(t, err)
		ctx.SetData(&model.Data{
			Forward: model.ServerToClient,
			Buffer:  buff,
		})
		err = filter.Handle(ctx)
		assert.Nil(t, err)
		// SqliProtocols
		protocol := []byte{0, 126, 0, 9, 189, 190, 159, 254, 127, 183, 255, 239, 240, 0}
		eot, err := (&model.SqliEot{}).Pack()
		assert.Nil(t, err)
		for _, c := range eot {
			protocol = append(protocol, c)
		}
		ctx.SetData(&model.Data{
			Forward: model.ServerToClient,
			Buffer:  protocol,
		})
		err = filter.Handle(ctx)
		assert.Nil(t, err)
		// SqliInfo
		ctx.SetData(&model.Data{
			Forward: model.ServerToClient,
			Buffer:  eot,
		})
		err = filter.Handle(ctx)
		assert.Nil(t, err)
		// SqliDBOpen
		done := &model.SqliDone{
			Warning:  21,
			Rows:     0,
			RowID:    0,
			SerialID: 0,
		}
		cost := &model.SqliCost{
			EstimatedRows: 1,
			EstimatedIO:   1,
		}
		eott := &model.SqliEot{}
		var transmission model.SqliTransmission
		transmission = []model.SqliCommand{done, cost, eott}
		buf, err = transmission.Pack()
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

	go func() {
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
		assert.IsType(t, &pgproto.CommandComplete{}, msg)
		msg, err = parse.Receive()
		assert.Nil(t, err)
		assert.IsType(t, &pgproto.ReadyForQuery{}, msg)

		parse, err = pgproto.NewFrontend(pgproto.NewChunkReader(pgFront), nil)
		assert.Nil(t, err)
		msg, err = parse.Receive()
		assert.Nil(t, err)
		assert.IsType(t, &pgproto.ParseComplete{}, msg)
		msg, err = parse.Receive()
		assert.Nil(t, err)
		assert.IsType(t, &pgproto.BindComplete{}, msg)
		msg, err = parse.Receive()
		assert.Nil(t, err)
		assert.IsType(t, &pgproto.CommandComplete{}, msg)
		msg, err = parse.Receive()
		assert.Nil(t, err)
		assert.IsType(t, &pgproto.ReadyForQuery{}, msg)
	}()

	{
		buffer := (&pgproto.Parse{
			Name:          "",
			Query:         "SET extra_float_digits = 3",
			ParameterOIDs: nil,
		}).Encode(nil)
		buffer = (&pgproto.Bind{
			DestinationPortal:    "",
			PreparedStatement:    "",
			ParameterFormatCodes: nil,
			Parameters:           nil,
			ResultFormatCodes:    nil,
		}).Encode(buffer)
		buffer = (&pgproto.Execute{
			Portal:  "",
			MaxRows: 1,
		}).Encode(buffer)
		buffer = (&pgproto.Sync{}).Encode(buffer)
		ctx.SetData(&model.Data{
			Forward: model.ClientToServer,
			Buffer:  buffer,
		})

		buffer = (&pgproto.Parse{
			Name:          "",
			Query:         "SET application_name = 'PostgreSQL JDBC Driver'",
			ParameterOIDs: nil,
		}).Encode(nil)
		buffer = (&pgproto.Bind{
			DestinationPortal:    "",
			PreparedStatement:    "",
			ParameterFormatCodes: nil,
			Parameters:           nil,
			ResultFormatCodes:    nil,
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
	}

	assert.True(t, server_passed)
}

func TestQueryFilter_Handle_INSERT(t *testing.T) {
	pgFront, pgBackend := net.Pipe()
	gbFront, gbBackend := net.Pipe()
	filter := NewQueryFilter()
	ctx := &Context{
		sessionId: 0,
		front:     pgBackend,
		backend:   gbFront,
		state:     QueryState,
		metadata:  make(map[string]interface{}),
	}
	buffer := (&pgproto.Parse{
		Name:          "",
		Query:         "insert into t values (1)",
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

	//server_passed := false
	// 8s Server
	go func() {
		//defer func() { server_passed = true }()
		buf := make([]byte, 1024)
		c, err := gbBackend.Read(buf)
		assert.Nil(t, err)
		assert.True(t, c > 0)
		buff := buf[:c]
		readseeker := bytes.NewReader(buff)
		msgs, err := model.UnpackSqliTransmission(readseeker)
		assert.Nil(t, err)
		assert.Equal(t, &model.SqliPrepare{
			QMarks: 0,
			Sql:    "insert into t values (1)",
		}, msgs[0])
		assert.IsType(t, &model.SqliNDescribe{}, msgs[1])
		assert.IsType(t, &model.SqliWantDone{}, msgs[2])
		assert.IsType(t, &model.SqliEot{}, msgs[3])

		c, err = gbBackend.Read(buf)
		assert.Nil(t, err)
		assert.True(t, c > 0)
		buff = buf[:c]
		readseeker = bytes.NewReader(buff)
		msgs, err = model.UnpackSqliTransmission(readseeker)
		assert.Nil(t, err)
		assert.IsType(t, &model.SqliID{}, msgs[0])
		assert.IsType(t, &model.SqliCIdescribe{}, msgs[1])
		assert.IsType(t, &model.SqliEot{}, msgs[2])

		c, err = gbBackend.Read(buf)
		assert.Nil(t, err)
		assert.True(t, c > 0)
		buff = buf[:c]
		readseeker = bytes.NewReader(buff)
		msgs, err = model.UnpackSqliTransmission(readseeker)
		assert.Nil(t, err)
		assert.IsType(t, &model.SqliID{}, msgs[0])
		//assert.IsType(t, &model.SqliBind{}, msgs[1])
		assert.IsType(t, &model.SqliExecute{}, msgs[1])
		assert.IsType(t, &model.SqliEot{}, msgs[2])

		c, err = gbBackend.Read(buf)
		assert.Nil(t, err)
		assert.True(t, c > 0)
		buff = buf[:c]
		readseeker = bytes.NewReader(buff)
		msgs, err = model.UnpackSqliTransmission(readseeker)
		assert.Nil(t, err)
		assert.IsType(t, &model.SqliID{}, msgs[0])
		assert.IsType(t, &model.SqliRelease{}, msgs[1])
		assert.IsType(t, &model.SqliEot{}, msgs[2])

	}()

	go func() {
		//defer func() { server_passed = true }()
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
		done := &model.SqliDone{
			Warning:  0,
			Rows:     0,
			RowID:    0,
			SerialID: 0,
		}
		cost := &model.SqliCost{
			EstimatedRows: 1,
			EstimatedIO:   2,
		}
		eot := &model.SqliEot{}
		var transmission model.SqliTransmission
		transmission = []model.SqliCommand{describe, done, cost, eot}
		buffer, err := transmission.Pack()
		assert.Nil(t, err)
		//TODO: 将 buffer 改成正式的 sqli 数据
		ctx.SetData(&model.Data{
			Forward: model.ServerToClient,
			Buffer:  buffer,
		})
		err = filter.Handle(ctx)
		assert.Nil(t, err)

		idescribe := &model.SqliIdescribe{
			Inputfields: 2,
			Fields: []model.Sqlifields{{
				Type:                 2,
				ExtendID:             0,
				OwnerNameLength:      0,
				ExtendTypeNameLength: 0,
				PassByReferenceFlag:  0,
				Alignment:            0,
				SourceType:           0,
				Length:               4,
			}, {
				Type:                 2,
				ExtendID:             0,
				OwnerNameLength:      0,
				ExtendTypeNameLength: 0,
				PassByReferenceFlag:  0,
				Alignment:            0,
				SourceType:           0,
				Length:               4,
			}},
		}
		transmission = []model.SqliCommand{idescribe, eot}
		buffer, err = transmission.Pack()
		assert.Nil(t, err)
		//TODO: 将 buffer 改成正式的 sqli 数据
		ctx.SetData(&model.Data{
			Forward: model.ServerToClient,
			Buffer:  buffer,
		})
		err = filter.Handle(ctx)
		assert.Nil(t, err)

		insertdone := &model.SqliInsertDone{
			Serial8:   1,
			BigSerial: 2,
		}
		done = &model.SqliDone{
			Warning:  0,
			Rows:     0,
			RowID:    0,
			SerialID: 0,
		}
		cost = &model.SqliCost{
			EstimatedRows: 1,
			EstimatedIO:   2,
		}
		transmission = []model.SqliCommand{insertdone, done, cost, eot}
		buffer, err = transmission.Pack()
		assert.Nil(t, err)
		//TODO: 将 buffer 改成正式的 sqli 数据
		ctx.SetData(&model.Data{
			Forward: model.ServerToClient,
			Buffer:  buffer,
		})
		err = filter.Handle(ctx)
		assert.Nil(t, err)

		transmission = []model.SqliCommand{eot}
		buffer, err = transmission.Pack()
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
		assert.IsType(t, &pgproto.NoData{}, msg)
		//msg, err = parse.Receive()
		//assert.Nil(t, err)
		//assert.IsType(t, &pgproto.RowDescription{}, msg)
		//msg, err = parse.Receive()
		//assert.Nil(t, err)
		//assert.IsType(t, &pgproto.DataRow{}, msg)
		msg, err = parse.Receive()
		assert.Nil(t, err)
		assert.IsType(t, &pgproto.CommandComplete{}, msg)
		msg, err = parse.Receive()
		assert.Nil(t, err)
		assert.IsType(t, &pgproto.ReadyForQuery{}, msg)
	}
	//assert.True(t, server_passed)
}

func TestQueryFilter_Handle_SELECT(t *testing.T) {
	pgFront, pgBackend := net.Pipe()
	gbFront, gbBackend := net.Pipe()
	filter := NewQueryFilter()
	ctx := &Context{
		sessionId: 0,
		front:     pgBackend,
		backend:   gbFront,
		state:     QueryState,
		metadata:  make(map[string]interface{}),
	}
	buffer := (&pgproto.Parse{
		Name:          "",
		Query:         "select * from t",
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

	//server_passed := false
	// 8s Server
	go func() {
		//defer func() { server_passed = true }()
		buf := make([]byte, 1024)
		c, err := gbBackend.Read(buf)
		assert.Nil(t, err)
		assert.True(t, c > 0)
		buff := buf[:c]
		readseeker := bytes.NewReader(buff)
		msgs, err := model.UnpackSqliTransmission(readseeker)
		assert.Nil(t, err)
		assert.Equal(t, &model.SqliPrepare{
			QMarks: 0,
			Sql:    "select * from t",
		}, msgs[0])
		assert.IsType(t, &model.SqliNDescribe{}, msgs[1])
		assert.IsType(t, &model.SqliWantDone{}, msgs[2])
		assert.IsType(t, &model.SqliEot{}, msgs[3])

		c, err = gbBackend.Read(buf)
		assert.Nil(t, err)
		assert.True(t, c > 0)
		buff = buf[:c]
		readseeker = bytes.NewReader(buff)
		msgs, err = model.UnpackSqliTransmission(readseeker)
		assert.Nil(t, err)
		assert.IsType(t, &model.SqliID{}, msgs[0])
		assert.IsType(t, &model.SqliCIdescribe{}, msgs[1])
		assert.IsType(t, &model.SqliEot{}, msgs[2])

		c, err = gbBackend.Read(buf)
		assert.Nil(t, err)
		assert.True(t, c > 0)
		buff = buf[:c]
		readseeker = bytes.NewReader(buff)
		msgs, err = model.UnpackSqliTransmission(readseeker)
		assert.Nil(t, err)
		assert.IsType(t, &model.SqliID{}, msgs[0])
		//assert.IsType(t, &model.SqliBind{}, msgs[1])
		assert.IsType(t, &model.SqliCurName{}, msgs[1])
		assert.IsType(t, &model.SqliOpen{}, msgs[2])
		assert.IsType(t, &model.SqliEot{}, msgs[3])

		c, err = gbBackend.Read(buf)
		assert.Nil(t, err)
		assert.True(t, c > 0)
		buff = buf[:c]
		readseeker = bytes.NewReader(buff)
		msgs, err = model.UnpackSqliTransmission(readseeker)
		assert.Nil(t, err)
		assert.IsType(t, &model.SqliID{}, msgs[0])
		assert.IsType(t, &model.SqliRetType{}, msgs[1])
		assert.IsType(t, &model.SqliNFetch{}, msgs[2])
		assert.IsType(t, &model.SqliEot{}, msgs[3])

	}()

	go func() {
		//defer func() { server_passed = true }()
		err := filter.Handle(ctx)
		assert.Nil(t, err)

		// gbase response
		describe := &model.SqliDescribe{
			StatementType: 2,
			StatementID:   0,
			EstimatedCost: 0,
			TupleSize:     4,
			CountOfFields: 1,
			StringTable:   3,
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
			},
			},
		}
		done := &model.SqliDone{
			Warning:  0,
			Rows:     0,
			RowID:    0,
			SerialID: 0,
		}
		cost := &model.SqliCost{
			EstimatedRows: 32,
			EstimatedIO:   2,
		}
		eot := &model.SqliEot{}
		var transmission model.SqliTransmission
		transmission = []model.SqliCommand{describe, done, cost, eot}
		buffer, err := transmission.Pack()
		assert.Nil(t, err)
		//TODO: 将 buffer 改成正式的 sqli 数据
		ctx.SetData(&model.Data{
			Forward: model.ServerToClient,
			Buffer:  buffer,
		})
		err = filter.Handle(ctx)
		assert.Nil(t, err)

		idescribe := &model.SqliIdescribe{
			Inputfields: 2,
			Fields: []model.Sqlifields{{
				Type:                 2,
				ExtendID:             0,
				OwnerNameLength:      0,
				ExtendTypeNameLength: 0,
				PassByReferenceFlag:  0,
				Alignment:            0,
				SourceType:           0,
				Length:               4,
			}, {
				Type:                 2,
				ExtendID:             0,
				OwnerNameLength:      0,
				ExtendTypeNameLength: 0,
				PassByReferenceFlag:  0,
				Alignment:            0,
				SourceType:           0,
				Length:               4,
			}},
		}
		transmission = []model.SqliCommand{idescribe, eot}
		buffer, err = transmission.Pack()
		assert.Nil(t, err)
		//TODO: 将 buffer 改成正式的 sqli 数据
		ctx.SetData(&model.Data{
			Forward: model.ServerToClient,
			Buffer:  buffer,
		})
		err = filter.Handle(ctx)
		assert.Nil(t, err)

		transmission = []model.SqliCommand{eot}
		buffer, err = transmission.Pack()
		assert.Nil(t, err)
		//TODO: 将 buffer 改成正式的 sqli 数据
		ctx.SetData(&model.Data{
			Forward: model.ServerToClient,
			Buffer:  buffer,
		})
		err = filter.Handle(ctx)
		assert.Nil(t, err)

		tuple := &model.SqliTuple{
			Warnings: 0,
			Size:     4,
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
			},
			},
		}
		done = &model.SqliDone{
			Warning:  0,
			Rows:     3,
			RowID:    259,
			SerialID: 0,
		}
		cost = &model.SqliCost{
			EstimatedRows: 32,
			EstimatedIO:   2,
		}
		transmission = []model.SqliCommand{tuple, tuple, tuple, done, cost, eot}
		buffer, err = transmission.Pack()
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
		//msg, err = parse.Receive()
		//assert.Nil(t, err)
		//assert.IsType(t, &pgproto.NoData{}, msg)
		msg, err = parse.Receive()
		assert.Nil(t, err)
		assert.IsType(t, &pgproto.RowDescription{}, msg)
		for i := 0; i < 3; i++ {
			msg, err = parse.Receive()
			assert.Nil(t, err)
			assert.IsType(t, &pgproto.DataRow{}, msg)
		}
		msg, err = parse.Receive()
		assert.Nil(t, err)
		assert.IsType(t, &pgproto.CommandComplete{}, msg)
		msg, err = parse.Receive()
		assert.Nil(t, err)
		assert.IsType(t, &pgproto.ReadyForQuery{}, msg)
	}
	//assert.True(t, server_passed)
}
