package bridge

import (
	"fmt"

	"github.com/jackc/pgproto3"
	"github.com/wangxuesong/tcpshadow/model"
	"github.com/wangxuesong/tcpshadow/pkg/services"
)

type QueryFilter struct {
}

func NewQueryFilter() *QueryFilter {
	return &QueryFilter{}
}

func (f *QueryFilter) Handle(ctx services.Context) error {
	if ctx.Data().Forward == model.ClientToServer {
		parser := model.NewPgClientParser()
		parser.Append(ctx.Data().Buffer)
		msg, err := parser.ParseMessage()
		if err != nil {
			return err
		}

		context, ok := ctx.(*Context)
		context.requests = msg
		if !ok {
			return fmt.Errorf("unknown context type: %T", ctx)
		}

		// TODO: send auth package to 8s
	}

	if ctx.Data().Forward == model.ServerToClient {
		//TODO: receive from 8s

		//TODO: 生成查询对应的 Pg 包
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
