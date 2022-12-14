package bridge

import (
	"bytes"
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
		prepare := &model.SqliPrepare{
			QMarks: 0,
			Sql:    "select * from test",
		}
		ndescribe := &model.SqliNDescribe{}
		wantdone := &model.SqliWantDone{}
		eot := &model.SqliEot{}
		var transmission model.SqliTransmission
		transmission = []model.SqliCommand{prepare, ndescribe, wantdone, eot}
		buffer, err := transmission.Pack()
		ctx.Data().Buffer = buffer
		context.backend.Write(ctx.Data().Buffer)
		ctx.SetMetaData("QueryStage", "QueryPrepareDone")
	}

	if ctx.Data().Forward == model.ServerToClient {
		//TODO: receive from 8s
		context, ok := ctx.(*Context)
		if !ok {
			return fmt.Errorf("unknown context type: %T", ctx)
		}

		//TODO: receive from 8s
		v, err := context.MetaData("QueryStage")
		if err != nil {
			return err
		}
		stage, ok := v.(string)
		if !ok {
			return fmt.Errorf("mistake metadata type: %T", v)
		}
		reader := bytes.NewReader(ctx.Data().Buffer)
		switch stage {
		case "QueryPrepareDone":
			//TODO: receive Sqli
			_, err := model.UnpackSqliTransmission(reader)
			if err != nil {
				return err
			}

			//TODO: send Sqli
			id := &model.SqliID{
				ID: 0,
			}
			cidescribe := &model.SqliCIdescribe{}
			eot := &model.SqliEot{}
			var transmission model.SqliTransmission
			transmission = []model.SqliCommand{id, cidescribe, eot}
			buf, err := transmission.Pack()
			if err != nil {
				return err
			}
			ctx.Data().Buffer = buf
			context.backend.Write(ctx.Data().Buffer)
			ctx.SetMetaData("QueryStage", "QueryIDescribeDone")
		case "":
			//TODO: receive Sqli
			_, err := model.UnpackSqliTransmission(reader)
			if err != nil {
				return err
			}

			//TODO: send Sqli
			id := &model.SqliID{
				ID: 0,
			}
			curname := &model.SqliCurName{
				CurName: "_ifxc0000000000000",
			}
			bind := &model.SqliBind{
				Columns: nil,
			}
			open := &model.SqliOpen{}
			eot := &model.SqliEot{}
			var transmission model.SqliTransmission
			transmission = []model.SqliCommand{id, curname, bind, open, eot}
			buf, err := transmission.Pack()
			if err != nil {
				return err
			}
			ctx.Data().Buffer = buf
			context.backend.Write(ctx.Data().Buffer)
			ctx.SetMetaData("QueryStage", "QueryOpen")
		case "QueryOpen":
			//TODO: receive Sqli
			_, err := model.UnpackSqliTransmission(reader)
			if err != nil {
				return err
			}

			//TODO: send Sqli
			id := &model.SqliID{
				ID: 0,
			}
			rettype := &model.SqliRetType{
				Direction: 1,
				Columns: []model.ColumnType{{
					Type:        2,
					Length:      4,
					OwnerName:   "",
					ExtTypeName: "",
				}, {
					Type:        2,
					Length:      4,
					OwnerName:   "",
					ExtTypeName: "",
				}, {
					Type:        13,
					Length:      80,
					OwnerName:   "",
					ExtTypeName: "",
				}},
			}
			nfetch := &model.SqliNFetch{
				TupleBufferSize: 4096,
				FetchArraySize:  0,
			}
			eot := &model.SqliEot{}
			var transmission model.SqliTransmission
			transmission = []model.SqliCommand{id, rettype, nfetch, eot}
			buf, err := transmission.Pack()
			if err != nil {
				return err
			}
			ctx.Data().Buffer = buf
			context.backend.Write(ctx.Data().Buffer)
			ctx.SetMetaData("QueryStage", "QueryDone")
		case "QueryIDescribeDone":
			//TODO: receive Sqli
			_, err := model.UnpackSqliTransmission(reader)
			if err != nil {
				return err
			}

			//TODO: send Sqli
			id := &model.SqliID{
				ID: 0,
			}
			bind := &model.SqliBind{
				Columns: []model.BindColumn{},
			}
			execute := &model.SqliExecute{}
			eot := &model.SqliEot{}
			var transmission model.SqliTransmission
			transmission = []model.SqliCommand{id, bind, execute, eot}
			buf, err := transmission.Pack()
			if err != nil {
				return err
			}
			ctx.Data().Buffer = buf
			context.backend.Write(ctx.Data().Buffer)
			ctx.SetMetaData("QueryStage", "QueryExecuteDone")
		case "QueryExecuteDone":
			//TODO: receive Sqli
			_, err := model.UnpackSqliTransmission(reader)
			if err != nil {
				return err
			}

			//TODO: send Sqli
			id := &model.SqliID{
				ID: 0,
			}
			release := &model.SqliRelease{}
			eot := &model.SqliEot{}
			var transmission model.SqliTransmission
			transmission = []model.SqliCommand{id, release, eot}
			buf, err := transmission.Pack()
			if err != nil {
				return err
			}
			ctx.Data().Buffer = buf
			context.backend.Write(ctx.Data().Buffer)
			ctx.SetMetaData("QueryStage", "QueryRelease")
		case "QueryDone":
			//TODO: receive Sqli
			_, err := model.UnpackSqliTransmission(reader)
			if err != nil {
				return err
			}

			//TODO: send pg
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
			response := [][]byte{[]byte("response")}
			d := &pgproto3.DataRow{Values: response}
			c := &pgproto3.CommandComplete{CommandTag: "SELECT 1"}
			re := &pgproto3.ReadyForQuery{TxStatus: 'I'}
			context.responses = []model.PgCommand{p, b, r, d, c, re}
			buf, err := context.responses.Pack()
			if err != nil {
				return err
			}

			_, err = context.front.Write(buf)
			if err != nil {
				return err
			}
			ctx.SetMetaData("QueryStage", EndStage)

			if ok && stage == EndStage {
				context.state = QueryState
			}
		case "QueryRelease":
			//TODO: receive Sqli
			_, err := model.UnpackSqliTransmission(reader)
			if err != nil {
				return err
			}

			//TODO: send pg
			p := &pgproto3.ParseComplete{}
			b := &pgproto3.BindComplete{}
			n := &pgproto3.NoData{}
			c := &pgproto3.CommandComplete{CommandTag: "INSERT 0 1"}
			re := &pgproto3.ReadyForQuery{TxStatus: 'I'}
			context.responses = []model.PgCommand{p, b, n, c, re}
			buf, err := context.responses.Pack()
			if err != nil {
				return err
			}

			_, err = context.front.Write(buf)
			if err != nil {
				return err
			}
			ctx.SetMetaData("QueryStage", EndStage)

			if ok && stage == EndStage {
				context.state = QueryState
			}
		}
	}
	return nil
}
