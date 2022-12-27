package bridge

import (
	"bytes"
	"fmt"
	"github.com/jackc/pgproto3"
	"github.com/wangxuesong/tcpshadow/model"
	"github.com/wangxuesong/tcpshadow/pkg/services"
)

type ConnectFilter struct {
}

func NewConnectFilter() *ConnectFilter {
	return &ConnectFilter{}
}

func (c *ConnectFilter) Handle(ctx services.Context) error {
	buff := bytes.NewBuffer(ctx.Data().Buffer)
	if ctx.Data().Forward == model.ClientToServer {
		context, ok := ctx.(*Context)
		if !ok {
			return fmt.Errorf("unknown context type: %T", ctx)
		}

		if context.requests == nil {
			backend, err := pgproto3.NewBackend(pgproto3.NewChunkReader(buff), nil)
			if err != nil {
				return err
			}

			msg, err := backend.ReceiveStartupMessage()
			if err != nil {
				return err
			}
			user := msg.Parameters["user"]
			_ = msg.Parameters["password"]

			context.requests = []model.PgCommand{msg}

			// TODO: send auth package to 8s
			authrequest, err := (&model.AuthRequest{
				Length:           433,
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
				Clientname:       user,
				Passwordlength:   33,
				Password:         "HmQOYC1ZfTYt+vlXUhkn3w==",
				Noname12:         "ol",
				Noname13:         61,
				Tlitcp:           "tlitcp",
				Noname14:         1,
				Noname15:         104,
				Asf:              11,
				Noname16:         3,
				Servernamelength: 12,
				Servername:       "gbaseserver",
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
				Hostnamelength:   16,
				Noname27:         "MM-202201031507",
				Noname28:         0,
				Directorylength:  21,
				Directory:        "E:\\JDBCTest\\JDBCTest",
				Noname29:         116,
				Appnamelengthall: 80,
				Noname30:         0,
				Noname31:         0,
				Appnamelength:    70,
				Appname:          "/E:/JDBCTest/JDBCTest/lib/gbasedbtjdbc_3.3.0_2.jarConnectionTest/Test",
				Asceot:           127,
			}).Pack()
			ctx.Data().Buffer = authrequest
			context.backend.Write(ctx.Data().Buffer)
			ctx.SetMetaData("ConnectStage", "ConnectProtocol")
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
		case "ConnectInfo":
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
		case "ConnectDbOpen":
			//TODO: receive SqliEot
			_, err := model.UnpackSqliTransmission(reader)

			//TODO: send SqliDBOpen
			deopen := &model.SqliDBOpen{
				DBName: "t",
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

		//v, err = context.MetaData("ConnectStage")
		//if err != nil {
		//	return err
		//}
		//stage, ok = v.(string)
		//if !ok {
		//	return fmt.Errorf("mistake metadata type: %T", v)
		//}
		//if ok && stage == EndStage {
		//	context.state = QueryState
		//}
	}

	return nil
}
