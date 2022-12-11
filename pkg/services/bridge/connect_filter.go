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

		context, ok := ctx.(*Context)
		if !ok {
			return fmt.Errorf("unknown context type: %T", ctx)
		}
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
			Clientname:       "gbasedbt",
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
	}

	if ctx.Data().Forward == model.ServerToClient {
		//TODO: receive from 8s
		reader := bytes.NewReader(ctx.Data().Buffer)
		authreponse := &model.AuthResponse{}
		authreponse.Unpack(reader)

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

		context, ok := ctx.(*Context)
		if !ok {
			return fmt.Errorf("unknown context type: %T", ctx)
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

		context.state = QueryState
	}

	return nil
}
