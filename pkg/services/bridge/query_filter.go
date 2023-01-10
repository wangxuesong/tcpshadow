package bridge

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/zhuangsirui/binpacker"
	"strings"

	"github.com/jackc/pgproto3"
	"github.com/wangxuesong/tcpshadow/model"
	"github.com/wangxuesong/tcpshadow/pkg/services"
)

type (
	QueryHandler interface {
		Handle(filter *QueryFilter, ctx services.Context) error
	}

	QueryFilter struct {
		handler QueryHandler
	}

	parseHandler struct {
	}

	prepareHandler struct {
	}

	cidescribeHandler struct {
	}

	executeHandler struct {
	}

	releaseHandler struct {
	}

	cidescribeSelectHandler struct {
	}

	openHandler struct {
	}

	nfetchdoneHandler struct {
	}
)

var tuple [1024]model.SqliTuple
var tupleNumber int
var fieldsname [1024]string
var data [1024]pgproto3.DataRow
var idnumber int16
var bindcondition int
var bindtype []uint32
var datebindint uint16
var datebindchar string
var idesfields []model.Sqlifields

func NewQueryFilter() *QueryFilter {
	return &QueryFilter{
		handler: &parseHandler{},
	}
}

func (f *QueryFilter) Handle(ctx services.Context) error {
	return f.handler.Handle(f, ctx)
}

func (c *QueryFilter) SetHandler(handler QueryHandler) {
	c.handler = handler
}

func (h *parseHandler) Handle(filter *QueryFilter, ctx services.Context) error {
	if ctx.Data().Forward != model.ClientToServer {
		return fmt.Errorf(" The error direction of the message is %T", model.ServerToClient)
	}

	context, ok := ctx.(*Context)
	if !ok {
		return fmt.Errorf("unknown context type: %T", ctx)
	}

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
	sql = strings.ReplaceAll(sql, "$1", "?")
	err = ctx.SetMetaData("Condition", sql[:6])
	bindtype = parse.ParameterOIDs
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
	err = ctx.SetMetaData("Bindcondition", len(bind.ParameterFormatCodes))
	binddata := bind.Parameters
	for c, t := range bindtype {
		if t == 23 {
			r := bytes.NewReader(binddata[c])
			unpacker := binpacker.NewUnpacker(binary.BigEndian, r)
			var pad uint16
			unpacker.FetchUint16(&pad).FetchUint16(&datebindint)
		} else {
			r := bytes.NewReader(binddata[c])
			unpacker := binpacker.NewUnpacker(binary.BigEndian, r)
			unpacker.FetchString(uint64(len(binddata[c])), &datebindchar)
		}
	}
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

	context.requests = msg

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
	filter.SetHandler(&prepareHandler{})
	return nil
}

func (h *prepareHandler) Handle(filter *QueryFilter, ctx services.Context) error {
	if ctx.Data().Forward != model.ServerToClient {
		return fmt.Errorf(" The error direction of the message is %T", model.ClientToServer)
	}

	context, ok := ctx.(*Context)
	if !ok {
		return fmt.Errorf("unknown context type: %T", ctx)
	}

	buffer := bytes.NewReader(ctx.Data().Buffer)
	//TODO: receive Sqli
	msgs, err := model.UnpackSqliTransmission(buffer)
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
	idnumber = int16(describe.StatementID)
	_ = describe.EstimatedCost
	_ = describe.CountOfFields
	_ = describe.StringTable
	for i := 0; i < len(describe.Fields); i++ {
		fieldsname[i] = describe.Fields[i].Name
	}

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

	v, err := context.MetaData("Condition")
	if v.(string) == "select" || v.(string) == "SELECT" {
		err = ctx.SetMetaData("QueryStage", "QuerySelectIDescribeDone")
		filter.SetHandler(&cidescribeSelectHandler{})
		if err != nil {
			return err
		}
	} else {
		err = ctx.SetMetaData("QueryStage", "QueryIDescribeDone")
		filter.SetHandler(&cidescribeHandler{})
		if err != nil {
			return err
		}
	}
	return nil
}

func (h *cidescribeHandler) Handle(filter *QueryFilter, ctx services.Context) error {
	if ctx.Data().Forward != model.ServerToClient {
		return fmt.Errorf(" The error direction of the message is %T", model.ClientToServer)
	}

	context, ok := ctx.(*Context)
	if !ok {
		return fmt.Errorf("unknown context type: %T", ctx)
	}

	buffer := bytes.NewReader(ctx.Data().Buffer)

	msgs, err := model.UnpackSqliTransmission(buffer)
	if err != nil {
		return err
	}
	buf := make([]byte, 1024)

	//idescribe
	buf, err = msgs[0].Pack()
	if err != nil {
		return err
	}
	reader := bytes.NewReader(buf[2:])
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
		ID: idnumber,
	}
	bind := &model.SqliBind{
		Columns: []model.BindColumn{},
	}
	if bindcondition != 0 {
		for _, c := range bindtype {
			var bindType int16
			bindint := &model.BindColumnInt{}
			bindchar := &model.BindColumnChar{}
			switch c {
			case 23:
				bindType = 2
				bindint = &model.BindColumnInt{
					Type:      bindType,
					Indicator: 0,
					Precision: 2560,
					Data:      datebindint,
				}
				bind.Columns = []model.BindColumn{*bindint}
			case 1043:
				bindType = 0
				bindchar = &model.BindColumnChar{
					Type:      bindType,
					Indicator: 0,
					Precision: 0,
					Data:      datebindchar,
				}
				bind.Columns = []model.BindColumn{*bindchar}
			}
		}
	}
	execute := &model.SqliExecute{}
	eot := &model.SqliEot{}
	var transmission model.SqliTransmission
	v, err := context.MetaData("Bindcondition")
	if v != 0 {
		transmission = []model.SqliCommand{id, bind, execute, eot}
	} else {
		transmission = []model.SqliCommand{id, execute, eot}
	}
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
	filter.SetHandler(&executeHandler{})
	if err != nil {
		return err
	}
	return nil
}

func (h *executeHandler) Handle(filter *QueryFilter, ctx services.Context) error {
	if ctx.Data().Forward != model.ServerToClient {
		return fmt.Errorf(" The error direction of the message is %T", model.ClientToServer)
	}

	context, ok := ctx.(*Context)
	if !ok {
		return fmt.Errorf("unknown context type: %T", ctx)
	}

	buffer := bytes.NewReader(ctx.Data().Buffer)

	msgs, err := model.UnpackSqliTransmission(buffer)
	if err != nil {
		return err
	}
	buf := make([]byte, 1024)

	//insertdone
	buf, err = msgs[0].Pack()
	if err != nil {
		return err
	}
	reader := bytes.NewReader(buf[2:])
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
		ID: idnumber,
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
	filter.SetHandler(&releaseHandler{})
	return nil
}

func (h *releaseHandler) Handle(filter *QueryFilter, ctx services.Context) error {
	if ctx.Data().Forward != model.ServerToClient {
		return fmt.Errorf(" The error direction of the message is %T", model.ClientToServer)
	}

	context, ok := ctx.(*Context)
	if !ok {
		return fmt.Errorf("unknown context type: %T", ctx)
	}

	buffer := bytes.NewReader(ctx.Data().Buffer)

	_, err := model.UnpackSqliTransmission(buffer)
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
	buff, err := context.responses.Pack()
	if err != nil {
		return err
	}

	_, err = context.front.Write(buff)
	if err != nil {
		return err
	}

	err = ctx.SetMetaData("QueryStage", EndStage)
	if err != nil {
		return err
	}

	v, err := context.MetaData("QueryStage")
	if err != nil {
		return err
	}
	stage, ok := v.(string)
	if ok && stage == EndStage {
		context.state = QueryState
	}
	return nil
}

func (h *cidescribeSelectHandler) Handle(filter *QueryFilter, ctx services.Context) error {
	if ctx.Data().Forward != model.ServerToClient {
		return fmt.Errorf(" The error direction of the message is %T", model.ClientToServer)
	}

	context, ok := ctx.(*Context)
	if !ok {
		return fmt.Errorf("unknown context type: %T", ctx)
	}

	buffer := bytes.NewReader(ctx.Data().Buffer)

	msgs, err := model.UnpackSqliTransmission(buffer)
	if err != nil {
		return err
	}
	buf := make([]byte, 1024)

	//idescribe
	buf, err = msgs[0].Pack()
	if err != nil {
		return err
	}
	reader := bytes.NewReader(buf[2:])
	idescribe := &model.SqliIdescribe{}
	err = idescribe.Unpack(reader)
	if err != nil {
		return err
	}
	_ = idescribe.Inputfields
	idesfields = idescribe.Fields
	_ = idescribe.Fields

	//eot
	_ = msgs[1]

	//TODO: send Sqli
	id := &model.SqliID{
		ID: idnumber,
	}
	curname := &model.SqliCurName{
		CurName: "_ifxc0000000000000",
	}
	bind := &model.SqliBind{
		Columns: []model.BindColumn{},
	}
	if bindcondition != 0 {
		for _, c := range bindtype {
			var bindType int16
			bindint := &model.BindColumnInt{}
			bindchar := &model.BindColumnChar{}
			switch c {
			case 23:
				bindType = 2
				bindint = &model.BindColumnInt{
					Type:      bindType,
					Indicator: 0,
					Precision: 2560,
					Data:      datebindint,
				}
				bind.Columns = []model.BindColumn{*bindint}
			case 1043:
				bindType = 0
				bindchar = &model.BindColumnChar{
					Type:      bindType,
					Indicator: 0,
					Precision: 0,
					Data:      datebindchar,
				}
				bind.Columns = []model.BindColumn{*bindchar}
			}
		}
	}
	open := &model.SqliOpen{}
	eot := &model.SqliEot{}
	var transmission model.SqliTransmission
	v, err := context.MetaData("Bindcondition")
	if v != 0 {
		transmission = []model.SqliCommand{id, curname, bind, open, eot}
	} else {
		transmission = []model.SqliCommand{id, curname, open, eot}
	}
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
	filter.SetHandler(&openHandler{})
	return nil
}

func (h *openHandler) Handle(filter *QueryFilter, ctx services.Context) error {
	if ctx.Data().Forward != model.ServerToClient {
		return fmt.Errorf(" The error direction of the message is %T", model.ClientToServer)
	}

	context, ok := ctx.(*Context)
	if !ok {
		return fmt.Errorf("unknown context type: %T", ctx)
	}

	buffer := bytes.NewReader(ctx.Data().Buffer)

	_, err := model.UnpackSqliTransmission(buffer)
	if err != nil {
		return err
	}

	//TODO: send Sqli
	id := &model.SqliID{
		ID: idnumber,
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
	buff, err := transmission.Pack()
	if err != nil {
		return err
	}
	ctx.Data().Buffer = buff
	_, err = context.backend.Write(ctx.Data().Buffer)
	if err != nil {
		return err
	}
	err = ctx.SetMetaData("QueryStage", "QueryDone")
	if err != nil {
		return err
	}
	filter.SetHandler(&nfetchdoneHandler{})
	return nil
}

func (h *nfetchdoneHandler) Handle(filter *QueryFilter, ctx services.Context) error {
	if ctx.Data().Forward != model.ServerToClient {
		return fmt.Errorf(" The error direction of the message is %T", model.ClientToServer)
	}

	context, ok := ctx.(*Context)
	if !ok {
		return fmt.Errorf("unknown context type: %T", ctx)
	}

	buffer := bytes.NewReader(ctx.Data().Buffer)

	msgs, err := model.UnpackSqliTransmission(buffer)
	if err != nil {
		return err
	}
	buf := make([]byte, 1024)
	tupleNumber = len(msgs) - 3

	//tuple
	for i := 0; i < tupleNumber; i++ {
		buf, err = msgs[i].Pack()
		if err != nil {
			return err
		}
		reader := bytes.NewReader(buf)
		err = tuple[i].Unpack(reader)
		if err != nil {
			return err
		}
		//_ = tuple[i].Warnings
		//_ = tuple[i].Size
		//_ = tuple[i].TupleBytes
		//_ = tuple[i].Values
		//_ = tuple[i].Fields
	}

	//done
	buf, err = msgs[tupleNumber].Pack()
	if err != nil {
		return err
	}
	reader := bytes.NewReader(buf[2:])
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
			Name:                 fieldsname[0],
			TableOID:             40963,
			TableAttributeNumber: 1,
			DataTypeOID:          23,
			DataTypeSize:         4,
			TypeModifier:         -1,
			Format:               0,
		},
	}}

	response := [][]byte{[]byte("1")}
	for i := 0; i < 3; i++ {
		data[i].Values = response
	}
	c := &pgproto3.CommandComplete{CommandTag: "SELECT 1"}
	re := &pgproto3.ReadyForQuery{TxStatus: 'I'}
	context.responses = []model.PgCommand{p, b, r, &data[0], &data[1], &data[2], c, re}
	buff, err := context.responses.Pack()
	if err != nil {
		return err
	}

	_, err = context.front.Write(buff)
	if err != nil {
		return err
	}
	err = ctx.SetMetaData("QueryStage", EndStage)
	if err != nil {
		return err
	}

	v, err := context.MetaData("QueryStage")
	if err != nil {
		return err
	}
	stage, ok := v.(string)
	if ok && stage == EndStage {
		context.state = QueryState
	}
	return nil
}
