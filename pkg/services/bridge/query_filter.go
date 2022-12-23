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

var condition string

func NewQueryFilter() *QueryFilter {
	return &QueryFilter{}
}

func (f *QueryFilter) Handle(ctx services.Context) error {

	if ctx.Data().Forward == model.ClientToServer {
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
		sql := parse.Query
		condition = sql[:6]
		paranumber := len(parse.ParameterOIDs)
		//bind
		buf = msg[1].Encode(nil)
		bind := &pgproto3.Bind{}
		err = bind.Decode(buf[5:])
		if err != nil {
			return err
		}
		_ = bind.DestinationPortal
		_ = bind.PreparedStatement
		_ = bind.ParameterFormatCodes
		_ = bind.Parameters
		_ = bind.ResultFormatCodes
		//describe
		buf = msg[2].Encode(nil)
		describe := &pgproto3.Describe{}
		err = describe.Decode(buf[5:])
		if err != nil {
			return err
		}
		_ = describe.ObjectType
		_ = describe.Name
		//execute
		buf = msg[3].Encode(nil)
		execute := &pgproto3.Execute{}
		err = execute.Decode(buf[5:])
		if err != nil {
			return err
		}
		_ = execute.Portal
		_ = execute.MaxRows
		//sync
		buf = msg[4].Encode(nil)
		sync := &pgproto3.Sync{}
		err = sync.Decode(buf[5:])
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
			QMarks: uint16(paranumber),
			Sql:    sql,
		}
		ndescribe := &model.SqliNDescribe{}
		wantdone := &model.SqliWantDone{}
		eot := &model.SqliEot{}
		var transmission model.SqliTransmission
		transmission = []model.SqliCommand{prepare, ndescribe, wantdone, eot}
		buffer, err := transmission.Pack()
		ctx.Data().Buffer = buffer
		_, err = context.backend.Write(ctx.Data().Buffer)
		if err != nil {
			return err
		}
		err = ctx.SetMetaData("QueryStage", "QueryPrepareDone")
		if err != nil {
			return err
		}
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
			msgs, err := model.UnpackSqliTransmission(reader)
			if err != nil {
				return err
			}
			buf := make([]byte, 1024)

			//describe
			buf, err = msgs[0].Pack()
			if err != nil {
				return err
			}
			reader := bytes.NewReader(buf[2:])
			describe := &model.SqliDescribe{}
			err = describe.Unpack(reader)
			if err != nil {
				return err
			}
			_ = describe.StringTable
			_ = describe.StatementType
			_ = describe.StatementID
			_ = describe.EstimatedCost
			_ = describe.TupleSize
			_ = describe.CountOfFields
			_ = describe.StringTable
			_ = describe.Fields

			//done
			buf, err = msgs[1].Pack()
			if err != nil {
				return err
			}
			reader = bytes.NewReader(buf[2:])
			done := &model.SqliDone{}
			err = done.Unpack(reader)
			if err != nil {
				return err
			}
			_ = done.Rows

			//cost
			buf, err = msgs[2].Pack()
			if err != nil {
				return err
			}
			reader = bytes.NewReader(buf[2:])
			cost := &model.SqliCost{}
			err = cost.Unpack(reader)
			if err != nil {
				return err
			}
			_ = cost.EstimatedRows

			//eot
			buf, err = msgs[3].Pack()

			//TODO: send Sqli
			id := &model.SqliID{
				ID: 0,
			}
			cidescribe := &model.SqliCIdescribe{}
			eot := &model.SqliEot{}
			var transmission model.SqliTransmission
			transmission = []model.SqliCommand{id, cidescribe, eot}
			buf, err = transmission.Pack()
			if err != nil {
				return err
			}
			ctx.Data().Buffer = buf
			_, err = context.backend.Write(ctx.Data().Buffer)
			if err != nil {
				return err
			}

			if condition != "select" {
				err = ctx.SetMetaData("QueryStage", "QueryIDescribeDone")
				if err != nil {
					return err
				}
			} else {
				err = ctx.SetMetaData("QueryStage", "QuerySelectIDescribeDone")
				if err != nil {
					return err
				}
			}
		case "QuerySelectIDescribeDone":
			//TODO: receive Sqli
			msgs, err := model.UnpackSqliTransmission(reader)
			if err != nil {
				return err
			}
			buf := make([]byte, 1024)

			//idescribe
			buf, err = msgs[0].Pack()
			if err != nil {
				return err
			}
			reader = bytes.NewReader(buf[2:])
			idescribe := &model.SqliIdescribe{}
			err = idescribe.Unpack(reader)
			if err != nil {
				return err
			}
			_ = idescribe.Inputfields
			_ = idescribe.Fields

			//eot
			_ = msgs[1]

			//TODO: send Sqli
			id := &model.SqliID{
				ID: 0,
			}
			curname := &model.SqliCurName{
				CurName: "_ifxc0000000000000",
			}
			//bind := &model.SqliBind{
			//	Columns: nil,
			//}
			open := &model.SqliOpen{}
			eot := &model.SqliEot{}
			var transmission model.SqliTransmission
			transmission = []model.SqliCommand{id, curname, open, eot}
			buf, err = transmission.Pack()
			if err != nil {
				return err
			}
			ctx.Data().Buffer = buf
			_, err = context.backend.Write(ctx.Data().Buffer)
			if err != nil {
				return err
			}
			err = ctx.SetMetaData("QueryStage", "QueryOpen")
			if err != nil {
				return err
			}
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
			_, err = context.backend.Write(ctx.Data().Buffer)
			if err != nil {
				return err
			}
			err = ctx.SetMetaData("QueryStage", "QueryDone")
			if err != nil {
				return err
			}
		case "QueryIDescribeDone":
			//TODO: receive Sqli
			msgs, err := model.UnpackSqliTransmission(reader)
			if err != nil {
				return err
			}
			buf := make([]byte, 1024)

			//idescribe
			buf, err = msgs[0].Pack()
			if err != nil {
				return err
			}
			reader = bytes.NewReader(buf[2:])
			idescribe := &model.SqliIdescribe{}
			err = idescribe.Unpack(reader)
			if err != nil {
				return err
			}
			_ = idescribe.Inputfields
			_ = idescribe.Fields

			//eot
			_ = msgs[1]

			//TODO: send Sqli
			id := &model.SqliID{
				ID: 0,
			}
			//bind := &model.SqliBind{
			//	Columns: []model.BindColumn{},
			//}
			execute := &model.SqliExecute{}
			eot := &model.SqliEot{}
			var transmission model.SqliTransmission
			transmission = []model.SqliCommand{id, execute, eot}
			buf, err = transmission.Pack()
			if err != nil {
				return err
			}
			ctx.Data().Buffer = buf
			_, err = context.backend.Write(ctx.Data().Buffer)
			if err != nil {
				return err
			}
			err = ctx.SetMetaData("QueryStage", "QueryExecuteDone")
			if err != nil {
				return err
			}
		case "QueryExecuteDone":
			//TODO: receive Sqli
			msgs, err := model.UnpackSqliTransmission(reader)
			if err != nil {
				return err
			}
			buf := make([]byte, 1024)

			//insertdone
			buf, err = msgs[0].Pack()
			if err != nil {
				return err
			}
			reader = bytes.NewReader(buf[2:])
			insertdone := &model.SqliInsertDone{}
			err = insertdone.Unpack(reader)
			if err != nil {
				return err
			}
			_ = insertdone.Serial8
			_ = insertdone.BigSerial

			//done
			buf, err = msgs[1].Pack()
			if err != nil {
				return err
			}
			reader = bytes.NewReader(buf[2:])
			done := &model.SqliDone{}
			err = done.Unpack(reader)
			if err != nil {
				return err
			}
			_ = done.Rows

			//cost
			buf, err = msgs[2].Pack()
			if err != nil {
				return err
			}
			reader = bytes.NewReader(buf[2:])
			cost := &model.SqliCost{}
			err = cost.Unpack(reader)
			if err != nil {
				return err
			}
			_ = cost.EstimatedRows

			//eot
			_ = msgs[3]

			//TODO: send Sqli
			id := &model.SqliID{
				ID: 0,
			}
			release := &model.SqliRelease{}
			eot := &model.SqliEot{}
			var transmission model.SqliTransmission
			transmission = []model.SqliCommand{id, release, eot}
			buf, err = transmission.Pack()
			if err != nil {
				return err
			}
			ctx.Data().Buffer = buf
			_, err = context.backend.Write(ctx.Data().Buffer)
			if err != nil {
				return err
			}
			err = ctx.SetMetaData("QueryStage", "QueryRelease")
			if err != nil {
				return err
			}
		case "QueryDone":
			//TODO: receive Sqli
			msgs, err := model.UnpackSqliTransmission(reader)
			if err != nil {
				return err
			}
			buf := make([]byte, 1024)
			tupleNumber := len(msgs) - 3

			//tuple
			var tuple [3]model.SqliTuple
			for i := 0; i < tupleNumber; i++ {
				buf, err = msgs[i].Pack()
				if err != nil {
					return err
				}
				reader = bytes.NewReader(buf[2:])
				err = tuple[i].Unpack(reader)
				if err != nil {
					return err
				}
				_ = tuple[i].Warnings
				_ = tuple[i].Size
				_ = tuple[i].TupleBytes
				_ = tuple[i].Values
				_ = tuple[i].Fields
			}

			//done
			buf, err = msgs[tupleNumber].Pack()
			if err != nil {
				return err
			}
			reader = bytes.NewReader(buf[2:])
			done := &model.SqliDone{}
			err = done.Unpack(reader)
			if err != nil {
				return err
			}
			_ = done.Rows

			//cost
			buf, err = msgs[tupleNumber+1].Pack()
			if err != nil {
				return err
			}
			reader = bytes.NewReader(buf[2:])
			cost := &model.SqliCost{}
			err = cost.Unpack(reader)
			if err != nil {
				return err
			}
			_ = cost.EstimatedRows

			//eot
			_ = msgs[tupleNumber+2]

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
			buf, err = context.responses.Pack()
			if err != nil {
				return err
			}

			_, err = context.front.Write(buf)
			if err != nil {
				return err
			}
			err = ctx.SetMetaData("QueryStage", EndStage)
			if err != nil {
				return err
			}

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
			err = ctx.SetMetaData("QueryStage", EndStage)
			if err != nil {
				return err
			}

			if ok && stage == EndStage {
				context.state = QueryState
			}
		}
	}
	return nil
}
