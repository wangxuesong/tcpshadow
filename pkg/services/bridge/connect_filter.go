package bridge

import (
	"bytes"
	"fmt"

	"github.com/jackc/pgproto3"
	"github.com/wangxuesong/tcpshadow/model"
	"github.com/wangxuesong/tcpshadow/pkg/services"
)

type (
	ConnectHandler interface {
		Handle(filter *ConnectFilter, ctx services.Context) error
	}

	ConnectFilter struct {
		handler ConnectHandler
	}

	startMessageHandler struct {
	}

	authResponseHandler struct {
	}

	protocolHandler struct {
	}

	infoHandler struct {
	}

	dbOpenHandler struct {
	}

	connectDone struct {
	}

	connectSet struct {
	}
)

func NewConnectFilter() *ConnectFilter {
	return &ConnectFilter{
		handler: &startMessageHandler{},
	}
}

func (c *ConnectFilter) Handle(ctx services.Context) error {
	_ = bytes.NewBuffer(ctx.Data().Buffer)
	if ctx.Data().Forward == model.ClientToServer {
		context, ok := ctx.(*Context)
		if !ok {
			return fmt.Errorf("unknown context type: %T", ctx)
		}

		if context.requests == nil {
			c.handler.Handle(c, ctx)
		} else {
			parser := model.NewPgClientParser()
			_, err := parser.Append(ctx.Data().Buffer)
			if err != nil {
				return err
			}
			msg, err := parser.ParseMessage()
			if err != nil {
				return err
			}
			buf := make([]byte, 2048)
			//parse
			buf = msg[0].Encode(nil)
			parse := &pgproto3.Parse{}
			err = parse.Decode(buf[5:])
			if err != nil {
				return err
			}
			query := parse.Query
			tag := query[:3]
			set := query[4:15]

			parsecomplete := &pgproto3.ParseComplete{}
			bindcomplete := &pgproto3.BindComplete{}
			commandcomplete := &pgproto3.CommandComplete{CommandTag: tag}
			readyforquery := &pgproto3.ReadyForQuery{TxStatus: 'T'}
			context.responses = []model.PgCommand{parsecomplete, bindcomplete, commandcomplete, readyforquery}
			buf, err = context.responses.Pack()
			if err != nil {
				return err
			}
			context.front.Write(buf)
			if set != "application" {
				err = ctx.SetMetaData("ConnectStage", "ConnectSet")
				if err != nil {
					return err
				}
			} else {
				err = ctx.SetMetaData("ConnectStage", EndStage)
				if err != nil {
					return err
				}
			}
		}
		v, err := context.MetaData("ConnectStage")
		if err != nil {
			return err
		}
		stage, ok := v.(string)
		if !ok {
			return fmt.Errorf("mistake metadata type: %T", v)
		}
		if ok && stage == EndStage {
			context.state = QueryState
		}
	}

	if ctx.Data().Forward == model.ServerToClient {
		context, ok := ctx.(*Context)
		if !ok {
			return fmt.Errorf("unknown context type: %T", ctx)
		}

		//TODO: receive from 8s
		v, err := context.MetaData("ConnectStage")
		if err != nil {
			return err
		}
		stage, ok := v.(string)
		if !ok {
			return fmt.Errorf("mistake metadata type: %T", v)
		}
		reader := bytes.NewReader(ctx.Data().Buffer)
		switch stage {
		case "ConnectProtocol":
			c.handler.Handle(c, ctx)
		case "ConnectInfo":
			c.handler.Handle(c, ctx)
		case "ConnectDbOpen":
			c.handler.Handle(c, ctx)
		case "ConnectDone":
			//TODO: receive SqliDone
			_, err := model.UnpackSqliTransmission(reader)

			// send Authentication to front
			auth := &pgproto3.Authentication{
				Type:               pgproto3.AuthTypeOk,
				Salt:               [4]byte{0},
				SASLAuthMechanisms: nil,
				SASLData:           nil,
			}
			status := &pgproto3.ParameterStatus{Name: "server_version", Value: "9.5"}
			key := &pgproto3.BackendKeyData{
				ProcessID: 881103,
				SecretKey: 1569992916,
			}
			ready := &pgproto3.ReadyForQuery{
				TxStatus: 73,
			}

			context.responses = []model.PgCommand{auth, status, key, ready}
			buf, err := context.responses.Pack()
			if err != nil {
				return err
			}

			_, err = context.front.Write(buf)
			if err != nil {
				return err
			}
			//ctx.SetMetaData("ConnectStage", EndStage)
			ctx.SetMetaData("ConnectStage", "ConnectSet")
		}
	}

	return nil
}

func (c *ConnectFilter) SetHandler(handler ConnectHandler) {
	c.handler = handler
}

func (h *startMessageHandler) Handle(filter *ConnectFilter, ctx services.Context) error {
	if ctx.Data().Forward != model.ClientToServer {
		panic("implement me")
	}

	buff := bytes.NewBuffer(ctx.Data().Buffer)
	context, ok := ctx.(*Context)
	if !ok {
		return fmt.Errorf("unknown context type: %T", ctx)
	}

	backend, err := pgproto3.NewBackend(pgproto3.NewChunkReader(buff), nil)
	if err != nil {
		return err
	}

	msg, err := backend.ReceiveStartupMessage()
	if err != nil {
		return err
	}
	_ = msg.Parameters["user"]
	_ = msg.Parameters["password"]

	context.requests = []model.PgCommand{msg}

	authrequest, err := (&model.AuthRequest{
		Length:           0x1de,
		Noname1:          1,
		Noname2:          60,
		Noname3:          0,
		Noname4:          100,
		Noname5:          101,
		Noname6:          61,
		Ieeemlength:      6,
		Ieeem:            "IEEEM",
		Noname7:          108,
		Sqlexec:          "sqlexec",
		Versionlength:    6,
		Version:          "9.280",
		Numberlength:     12,
		Rds:              "RDS#R000000",
		Sqlilength:       5,
		Sqli:             "sqli",
		Noname8:          316,
		Noname9:          0,
		Noname10:         0,
		Noname11:         1,
		Clientnamelength: 9,
		Clientname:       "gbasedbt",
		Passwordlength:   33,
		Password:         "lvZpxbgMFwx8jrpeeicQEQ==",
		Noname12:         "ol",
		Noname13:         61,
		Tlitcp:           "tlitcp",
		Noname14:         1,
		Noname15:         104,
		Asf:              11,
		Noname16:         3,
		Servernamelength: 14,
		Servername:       "ol_gbasedbt_1",
		Noname17:         0,
		Noname18:         0,
		Noname19:         0,
		Noname20:         0,
		Noname21:         0,
		Noname22:         106,
		Noname23:         6,
		Dpath: []model.DPath{{
			Dbpathlength:          7,
			Dbpath:                "DBPATH",
			Dbpathattributelength: 2,
			Dbpathattribute:       ".",
		}, {
			Dbpathlength:          17,
			Dbpath:                "CLNT_PAM_CAPABLE",
			Dbpathattributelength: 2,
			Dbpathattribute:       "1",
		}, {
			Dbpathlength:          7,
			Dbpath:                "DBDATE",
			Dbpathattributelength: 6,
			Dbpathattribute:       "Y4MD-",
		}, {
			Dbpathlength:          12,
			Dbpath:                "IFX_UPDDESC",
			Dbpathattributelength: 2,
			Dbpathattribute:       "1",
		}, {
			Dbpathlength:          8,
			Dbpath:                "SQLMODE",
			Dbpathattributelength: 6,
			Dbpathattribute:       "gbase",
		}, {
			Dbpathlength:          9,
			Dbpath:                "NODEFDAC",
			Dbpathattributelength: 3,
			Dbpathattribute:       "no",
		}},
		Noname24:         107,
		Noname25:         0,
		Noname26:         0,
		Longthreadid:     1,
		Hostnamelength:   6,
		Noname27:         "bogon",
		Noname28:         0,
		Directorylength:  43,
		Directory:        "/Users/martin/projects/8sprojects/JDBCTest",
		Noname29:         116,
		Appnamelengthall: 111,
		Noname30:         0,
		Noname31:         0,
		Appnamelength:    101,
		Appname:          "/Users/martin/projects/8sprojects/JDBCTest/lib/gbasedbtjdbc_3.3.0_2.jarConnectionTest/ConnectionTest",
		Asceot:           127,
	}).Pack()
	ctx.Data().Buffer = authrequest
	context.backend.Write(ctx.Data().Buffer)
	ctx.SetMetaData("ConnectStage", "ConnectProtocol")
	filter.SetHandler(&authResponseHandler{})
	return nil
}

func (h *authResponseHandler) Handle(filter *ConnectFilter, ctx services.Context) error {
	if ctx.Data().Forward != model.ServerToClient {
		panic("implement me")
	}

	context, ok := ctx.(*Context)
	if !ok {
		return fmt.Errorf("unknown context type: %T", ctx)
	}

	reader := bytes.NewReader(ctx.Data().Buffer)
	authreponse := &model.AuthResponse{}
	authreponse.Unpack(reader)

	//TODO:send protocols
	protocol := &model.SqliProtocols{
		Protocol: []byte{0xff, 0xfc, 0x7f, 0xfc, 0x3c, 0x8c, 0xaa, 0x97, 0x10},
	}
	eot := &model.SqliEot{}
	var transmission model.SqliTransmission
	transmission = []model.SqliCommand{protocol, eot}
	buf, err := transmission.Pack()
	if err != nil {
		return err
	}
	ctx.Data().Buffer = buf
	context.backend.Write(ctx.Data().Buffer)
	ctx.SetMetaData("ConnectStage", "ConnectInfo")
	filter.SetHandler(&protocolHandler{})
	return nil
}

func (h *protocolHandler) Handle(filter *ConnectFilter, ctx services.Context) error {
	if ctx.Data().Forward != model.ServerToClient {
		panic("implement me")
	}

	context, ok := ctx.(*Context)
	if !ok {
		return fmt.Errorf("unknown context type: %T", ctx)
	}

	reader := bytes.NewReader(ctx.Data().Buffer)
	//TODO: receive SqliProtocols
	_, err := model.UnpackSqliTransmission(reader)

	//TODO: send SqliInfo
	info := &model.SqliInfo{
		MessageType: 6,
		Length:      38,
		InfoEnv: model.InfoEnv{
			NameLength:  12,
			ValueLength: 4,
			Env: map[string]string{
				"DBTEMP":      "/tmp",
				"SUBQCACHESZ": "10",
			},
		},
	}
	eot := &model.SqliEot{}
	var transmission model.SqliTransmission
	transmission = []model.SqliCommand{info, eot}
	buf, err := transmission.Pack()
	if err != nil {
		return err
	}
	ctx.Data().Buffer = buf
	context.backend.Write(ctx.Data().Buffer)
	ctx.SetMetaData("ConnectStage", "ConnectDbOpen")
	filter.SetHandler(&infoHandler{})
	return nil
}

func (h *infoHandler) Handle(filter *ConnectFilter, ctx services.Context) error {
	if ctx.Data().Forward != model.ServerToClient {
		panic("implement me")
	}

	context, ok := ctx.(*Context)
	if !ok {
		return fmt.Errorf("unknown context type: %T", ctx)
	}

	reader := bytes.NewReader(ctx.Data().Buffer)
	//TODO: receive SqliEot
	_, err := model.UnpackSqliTransmission(reader)

	//TODO: send SqliDBOpen
	deopen := &model.SqliDBOpen{
		DBName: "dfe",
		Foo:    0,
	}
	eot := &model.SqliEot{}
	var transmission model.SqliTransmission
	transmission = []model.SqliCommand{deopen, eot}
	buf, err := transmission.Pack()
	if err != nil {
		return err
	}
	ctx.Data().Buffer = buf
	context.backend.Write(ctx.Data().Buffer)
	ctx.SetMetaData("ConnectStage", "ConnectDone")
	filter.SetHandler(&dbOpenHandler{})
	return nil
}

func (h *dbOpenHandler) Handle(filter *ConnectFilter, ctx services.Context) error {
	//TODO implement me
	panic("implement me")
}

func (h *connectDone) Handle(filter *ConnectFilter, ctx services.Context) error {
	//TODO implement me
	panic("implement me")
}

func (h *connectSet) Handle(filter *ConnectFilter, ctx services.Context) error {
	//TODO implement me
	panic("implement me")
}
