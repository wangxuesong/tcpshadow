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

		// TODO: send sqli package to 8s
		context.backend.Write(ctx.Data().Buffer)
	}

	if ctx.Data().Forward == model.ServerToClient {
		//TODO: receive from 8s

		//TODO: 生成查询对应的 Pg 包
		p := &pgproto3.ParseComplete{}
		b := &pgproto3.BindComplete{}
		r := &pgproto3.RowDescription{Fields: []pgproto3.FieldDescription{
			{
				Name:                 "id",
				TableOID:             40963,
				TableAttributeNumber: 1,
				DataTypeOID:          23,
				DataTypeSize:         4,
				TypeModifier:         -1,
				Format:               0,
			},
		}}
		context, ok := ctx.(*Context)
		response := [][]byte{[]byte("response")}
		d := &pgproto3.DataRow{Values: response}
		c := &pgproto3.CommandComplete{CommandTag: "SELECT 1"}
		re := &pgproto3.ReadyForQuery{TxStatus: 'I'}
		if !ok {
			return fmt.Errorf("unknown context type: %T", ctx)
		}
		context.responses = []model.PgCommand{p, b, r, d, c, re}
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
