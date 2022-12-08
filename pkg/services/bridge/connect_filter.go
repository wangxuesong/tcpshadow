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
		authcommand, err := (&model.AuthRequest{
			Header: []model.Header{{
				Length:  253,
				Noname1: 1,
				Noname2: 60,
				Noname3: 0,
			}},
			Body: []model.Body{{
				Noname1:               100,
				Noname2:               101,
				Noname3:               61,
				Ieeemlength:           6,
				Ieeem:                 "IEEEM",
				Non:                   0,
				Noname4:               108,
				Sqlexec:               "Sqlexec     ",
				Versionlength:         6,
				Version:               "9.280",
				Noname5:               0,
				Numberlength:          12,
				Rds:                   "RDS#R      ",
				Noname6:               0,
				Sqlilength:            5,
				Sqli:                  "sqli",
				Noname7:               0,
				Noname8:               316,
				Noname9:               0,
				Noname10:              0,
				Noname11:              1,
				Clientnamelength:      9,
				Clientname:            "gbasedbt",
				Noname12:              0,
				Passwordlength:        7,
				Password:              "111111",
				Noname13:              0,
				Noname14:              "o1000000",
				Noname15:              61,
				Tlitcp:                "Tlitcp00",
				Noname16:              1,
				Noname17:              104,
				Asf:                   11,
				Noname18:              3,
				Servernamelength:      18,
				Servername:            "Ol_gbasedbt1210_8",
				Noname19:              0,
				Noname20:              0,
				Noname21:              0,
				Noname22:              0,
				Noname23:              0,
				Noname24:              0,
				Noname25:              106,
				Noname26:              6,
				Dbpathlength:          7,
				Dbpath:                "dbpath",
				Noname27:              0,
				Dbpathattributelength: 1,
				Dbpathattribute:       "",
				Noname28:              0,
				Noname29:              107,
				Noname30:              0,
				Noname31:              0,
				Longthreadid:          1,
				Hostnamelength:        9,
				Noname32:              "8t_vm_30",
				Noname33:              0,
				Noname34:              0,
				Directorylength:       22,
				Directory:             "/home/zhangyaru/gbase",
				Noname35:              0,
				Noname36:              116,
				Appnamelengthall:      18,
				Noname37:              0,
				Noname38:              0,
				Appnamelength:         8,
				Appname:               "XXXXXXX",
				Noname39:              0,
				Asceot:                127,
			}},
		}).Pack()
		ctx.Data().Buffer = authcommand
		context.backend.Write(ctx.Data().Buffer)
	}

	if ctx.Data().Forward == model.ServerToClient {
		//TODO: receive from 8s
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
