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
		context.requests = []model.PgCommand{msg}
		if !ok {
			return fmt.Errorf("unknown context type: %T", ctx)
		}

		// TODO: send auth package to 8s
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
