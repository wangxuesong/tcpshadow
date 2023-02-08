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

	prepareDescribeHandler struct {
	}

	cidesbatchHandler struct {
	}
)

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
	//buff := bytes.NewBuffer(ctx.Data().Buffer)
	context, ok := ctx.(*Context)
	if !ok {
		return fmt.Errorf("unknown context type: %T", ctx)
	}
	parser := model.NewPgClientParser()
	_, err := parser.Append(ctx.Data().Buffer)
	if err != nil {
		return fmt.Errorf("failed to append parser buffer, err: %T", err)
	}
	msg, err := parser.ParseMessage()
	if err != nil {
		return fmt.Errorf("failed to unpack ParseMessage, err: %T", err)
	}
	err = context.SetMetaData("parse", msg)
	buf := make([]byte, 2048)
	b, err := ctx.MetaData("parse")
	if err != nil {
		return err
	}
	buf = b.(model.PgTransmission)[1].Encode(nil)
	r := bytes.NewReader(buf[:1])
	unpacker := binpacker.NewUnpacker(binary.BigEndian, r)
	var mestype byte
	unpacker.FetchByte(&mestype)
	switch mestype {
	case 66:
		h.HandleHandle1(filter, ctx)
	case 68:
		h.HandleHandle2(filter, ctx)
	}

	return nil
}

func (h *prepareHandler) Handle(filter *QueryFilter, ctx services.Context) error {
	if ctx.Data().Forward != model.ServerToClient {
		return fmt.Errorf(" The error direction of the message is %T", model.ClientToServer)
	}
	reader := bytes.NewReader(ctx.Data().Buffer)
	context, ok := ctx.(*Context)
	if !ok {
		return fmt.Errorf("unknown context type: %T", ctx)
	}
	mges, err := model.UnpackSqliTransmission(reader)
	if err != nil {
		return fmt.Errorf("failed to unpack SqliInfo response packages, err: %T", err)
	}
	context.SetMetaData("sqliprepare", mges)

	//TODO: receive Sqli
	buf := make([]byte, 1024)
	b, err := ctx.MetaData("sqliprepare")

	//describe
	describe := b.(model.SqliTransmission)[0].(model.SqliCommand).(*model.SqliDescribe)
	field := describe.Fields
	err = ctx.SetMetaData("Idnumber", int16(describe.StatementID))
	var fieldsname [1024]string
	for i := 0; i < len(field); i++ {
		fieldsname[i] = field[i].Name
	}
	err = ctx.SetMetaData("Fieldsname", fieldsname)

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
	context.backend.Write(ctx.Data().Buffer)

	v, err := ctx.MetaData("Condition")
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
	reader := bytes.NewReader(ctx.Data().Buffer)
	context, ok := ctx.(*Context)
	if !ok {
		return fmt.Errorf("unknown context type: %T", ctx)
	}
	mges, err := model.UnpackSqliTransmission(reader)
	if err != nil {
		return fmt.Errorf("failed to unpack SqliInfo response packages, err: %T", err)
	}
	context.SetMetaData("sqlicidescribe", mges)
	buf := make([]byte, 1024)
	groupcount, err := ctx.MetaData("Groupcount")
	_, err = ctx.MetaData("sqlicidescribe")

	//TODO: send Sqli
	b, err := ctx.MetaData("Idnumber")
	id := &model.SqliID{
		ID: b.(int16),
	}
	bind := &model.SqliBind{
		Columns: []model.BindColumn{},
	}
	v, err := ctx.MetaData("Bindcondition")
	if v != 0 {
		a, err := ctx.MetaData("Bindtype")
		if err != nil {
			return err
		}
		for _, c := range a.([]uint32) {
			var bindType int16
			bindint := &model.BindColumnInt{}
			bindchar := &model.BindColumnChar{}
			if err != nil {
				return err
			}
			switch c {
			case 23:
				t, err := ctx.MetaData("Datebindint")
				if err != nil {
					return err
				}
				bindType = 2
				bindint = &model.BindColumnInt{
					Type:      bindType,
					Indicator: 0,
					Precision: 2560,
					Data:      t.(uint16),
				}
				bind.Columns = []model.BindColumn{*bindint}
			case 1043:
				p, err := ctx.MetaData("Datebindchar")
				if err != nil {
					return err
				}
				bindType = 0
				bindchar = &model.BindColumnChar{
					Type:      bindType,
					Indicator: 0,
					Precision: 0,
					Data:      p.(string),
				}
				bind.Columns = []model.BindColumn{*bindchar}
			}
		}
	}
	execute := &model.SqliExecute{}
	eot := &model.SqliEot{}
	var beloop []model.SqliCommand
	for i := 0; i < groupcount.(int); i++ {
		beloop = append(beloop, bind, execute)
	}

	var transmission model.SqliTransmission
	if v != 0 {
		transmission = []model.SqliCommand{id}
		for _, v := range beloop {
			transmission = append(transmission, v)
		}
		transmission = append(transmission, eot)
	} else {
		transmission = []model.SqliCommand{id, execute, eot}
	}
	buf, err = transmission.Pack()
	if err != nil {
		return err
	}
	ctx.Data().Buffer = buf
	context.backend.Write(ctx.Data().Buffer)
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
	reader := bytes.NewReader(ctx.Data().Buffer)
	context, ok := ctx.(*Context)
	if !ok {
		return fmt.Errorf("unknown context type: %T", ctx)
	}
	mges, err := model.UnpackSqliTransmission(reader)
	if err != nil {
		return fmt.Errorf("failed to unpack SqliInfo response packages, err: %T", err)
	}
	context.SetMetaData("sqliexecute", mges)
	buf := make([]byte, 1024)

	//TODO: send Sqli
	b, err := ctx.MetaData("Idnumber")
	id := &model.SqliID{
		ID: b.(int16),
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
	context.backend.Write(ctx.Data().Buffer)
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
	reader := bytes.NewReader(ctx.Data().Buffer)
	context, ok := ctx.(*Context)
	if !ok {
		return fmt.Errorf("unknown context type: %T", ctx)
	}
	mges, err := model.UnpackSqliTransmission(reader)
	if err != nil {
		return fmt.Errorf("failed to unpack SqliInfo response packages, err: %T", err)
	}
	context.SetMetaData("sqlirelease", mges)

	groupcount, err := context.MetaData("Groupcount")
	if err != nil {
		return err
	}

	//TODO: send pg
	p := &pgproto3.ParseComplete{}
	b := &pgproto3.BindComplete{}
	n := &pgproto3.NoData{}
	c := &pgproto3.CommandComplete{CommandTag: "INSERT 0 1"}
	re := &pgproto3.ReadyForQuery{TxStatus: 'I'}
	var beloop []model.PgCommand
	for i := 0; i < groupcount.(int); i++ {
		beloop = append(beloop, b, n, c)
	}
	if groupcount.(int) > 1 {
		for _, v := range beloop {
			context.responses = append(context.responses, v)
		}
		context.responses = append(context.responses, re)
	} else {
		context.responses = []model.PgCommand{p, b, n, c, re}
	}
	buff, err := context.responses.Pack()
	if err != nil {
		return err
	}

	ctx.Data().Buffer = buff
	context.front.Write(ctx.Data().Buffer)

	err = ctx.SetMetaData("QueryStage", EndStage)
	if err != nil {
		return err
	}
	filter.SetHandler(&parseHandler{})

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
	reader := bytes.NewReader(ctx.Data().Buffer)
	context, ok := ctx.(*Context)
	if !ok {
		return fmt.Errorf("unknown context type: %T", ctx)
	}
	mges, err := model.UnpackSqliTransmission(reader)
	if err != nil {
		return fmt.Errorf("failed to unpack SqliInfo response packages, err: %T", err)
	}
	context.SetMetaData("sqlicidescribeSelect", mges)
	m, err := ctx.MetaData("sqlicidescribeSelect")
	//idescribe
	idescribe := m.(model.SqliTransmission)[0].(model.SqliCommand).(*model.SqliIdescribe)
	if err != nil {
		return err
	}
	err = ctx.SetMetaData("Idesfields", idescribe.Fields)

	//TODO: send Sqli
	b, err := ctx.MetaData("Idnumber")
	id := &model.SqliID{
		ID: b.(int16),
	}
	curname := &model.SqliCurName{
		CurName: "_ifxc0000000000000",
	}
	bind := &model.SqliBind{
		Columns: []model.BindColumn{},
	}
	v, err := ctx.MetaData("Bindcondition")
	if v != 0 {
		a, err := ctx.MetaData("Bindtype")
		if err != nil {
			return err
		}
		for _, c := range a.([]uint32) {
			var bindType int16
			bindint := &model.BindColumnInt{}
			bindchar := &model.BindColumnChar{}
			if err != nil {
				return err
			}
			switch c {
			case 23:
				t, err := ctx.MetaData("Datebindint")
				if err != nil {
					return err
				}
				bindType = 2
				bindint = &model.BindColumnInt{
					Type:      bindType,
					Indicator: 0,
					Precision: 2560,
					Data:      t.(uint16),
				}
				bind.Columns = []model.BindColumn{*bindint}
			case 1043:
				p, err := ctx.MetaData("Datebindchar")
				if err != nil {
					return err
				}
				bindType = 0
				bindchar = &model.BindColumnChar{
					Type:      bindType,
					Indicator: 0,
					Precision: 0,
					Data:      p.(string),
				}
				bind.Columns = []model.BindColumn{*bindchar}
			}
		}
	}
	open := &model.SqliOpen{}
	eot := &model.SqliEot{}
	var transmission model.SqliTransmission
	if v != 0 {
		transmission = []model.SqliCommand{id, curname, bind, open, eot}
	} else {
		transmission = []model.SqliCommand{id, curname, open, eot}
	}
	var buf []byte
	buf, err = transmission.Pack()
	if err != nil {
		return err
	}
	ctx.Data().Buffer = buf
	context.backend.Write(ctx.Data().Buffer)
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
	reader := bytes.NewReader(ctx.Data().Buffer)
	context, ok := ctx.(*Context)
	if !ok {
		return fmt.Errorf("unknown context type: %T", ctx)
	}
	mges, err := model.UnpackSqliTransmission(reader)
	if err != nil {
		return fmt.Errorf("failed to unpack SqliInfo response packages, err: %T", err)
	}
	context.SetMetaData("sqliopen", mges)

	//TODO: send Sqli
	b, err := ctx.MetaData("Idnumber")
	id := &model.SqliID{
		ID: b.(int16),
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
	context.backend.Write(ctx.Data().Buffer)
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
	reader := bytes.NewReader(ctx.Data().Buffer)
	context, ok := ctx.(*Context)
	if !ok {
		return fmt.Errorf("unknown context type: %T", ctx)
	}
	mges, err := model.UnpackSqliTransmission(reader)
	if err != nil {
		return fmt.Errorf("failed to unpack SqliInfo response packages, err: %T", err)
	}
	context.SetMetaData("sqlinfetchdone", mges)
	n, err := ctx.MetaData("sqlinfetchdone")
	err = ctx.SetMetaData("TupleNumber", len(n.(model.SqliTransmission))-3)
	//tuple
	var tuple [1024]*model.SqliTuple
	t, err := ctx.MetaData("TupleNumber")
	for i := 0; i < t.(int); i++ {
		tuple[i] = n.(model.SqliTransmission)[i].(model.SqliCommand).(*model.SqliTuple)
	}
	err = ctx.SetMetaData("Tuple", tuple)

	//TODO: send pg
	buff := (&pgproto3.ParseComplete{}).Encode(nil)
	buff = (&pgproto3.BindComplete{}).Encode(buff)
	m, err := ctx.MetaData("Fieldsname")
	buff = (&pgproto3.RowDescription{Fields: []pgproto3.FieldDescription{
		{
			Name:                 m.([1024]string)[0],
			TableOID:             40963,
			TableAttributeNumber: 1,
			DataTypeOID:          23,
			DataTypeSize:         4,
			TypeModifier:         -1,
			Format:               0,
		},
	}}).Encode(buff)

	response := [][]byte{[]byte("1")}
	var data [1024]pgproto3.DataRow
	for i := 0; i < 3; i++ {
		data[i].Values = response
	}
	err = ctx.SetMetaData("Data", data)
	p, err := ctx.MetaData("Data")
	row := p.([1024]pgproto3.DataRow)
	buff = row[0].Encode(buff)
	buff = row[1].Encode(buff)
	buff = row[2].Encode(buff)
	buff = (&pgproto3.CommandComplete{CommandTag: "SELECT 1"}).Encode(buff)
	buff = (&pgproto3.ReadyForQuery{TxStatus: 'I'}).Encode(buff)
	ctx.Data().Buffer = buff
	context.front.Write(ctx.Data().Buffer)
	err = ctx.SetMetaData("QueryStage", EndStage)
	if err != nil {
		return err
	}
	filter.SetHandler(&parseHandler{})

	v, err := ctx.MetaData("QueryStage")
	if err != nil {
		return err
	}
	stage, ok := v.(string)
	if ok && stage == EndStage {
		context.state = QueryState
	}
	return nil
}

func (h *parseHandler) HandleHandle1(filter *QueryFilter, ctx services.Context) error {
	context, ok := ctx.(*Context)
	if !ok {
		return fmt.Errorf("unknown context type: %T", ctx)
	}
	b, err := ctx.MetaData("parse")
	if err != nil {
		return err
	}
	//parse
	parse := b.(model.PgTransmission)[0].(model.PgCommand).(*pgproto3.Parse)
	sql := parse.Query
	sql = strings.ReplaceAll(sql, "$1", "?")
	ctx.SetMetaData("Condition", sql[:6])
	ctx.SetMetaData("Bindtype", parse.ParameterOIDs)
	paranumber := len(parse.ParameterOIDs)

	//bind
	bind := b.(model.PgTransmission)[1].(model.PgCommand).(*pgproto3.Bind)
	err = ctx.SetMetaData("Bindcondition", len(bind.ParameterFormatCodes))
	binddata := bind.Parameters
	a, err := context.MetaData("Bindtype")
	var datebindint uint16
	var datebindchar string
	for c, t := range a.([]uint32) {
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
	err = ctx.SetMetaData("Datebindint", datebindint)
	if err != nil {
		return err
	}
	err = ctx.SetMetaData("Datebindchar", datebindchar)
	if err != nil {
		return err
	}

	context.requests = b.(model.PgTransmission)

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
	context.backend.Write(ctx.Data().Buffer)
	err = ctx.SetMetaData("QueryStage", "QueryPrepareDone")
	if err != nil {
		return err
	}
	filter.SetHandler(&prepareHandler{})
	err = ctx.SetMetaData("Groupcount", 1)
	return nil
}

func (h *parseHandler) HandleHandle2(filter *QueryFilter, ctx services.Context) error {
	context, ok := ctx.(*Context)
	if !ok {
		return fmt.Errorf("unknown context type: %T", ctx)
	}
	b, err := ctx.MetaData("parse")
	if err != nil {
		return err
	}

	//parse
	parse := b.(model.PgTransmission)[0].(model.PgCommand).(*pgproto3.Parse)
	sql := parse.Query
	sql = strings.ReplaceAll(sql, "$1", "?")
	err = ctx.SetMetaData("Condition", sql[:6])
	err = ctx.SetMetaData("Bindtype", parse.ParameterOIDs)
	paranumber := len(parse.ParameterOIDs)

	context.requests = b.(model.PgTransmission)

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
	context.backend.Write(ctx.Data().Buffer)
	err = ctx.SetMetaData("QueryStage", "QueryPrepareDone")
	if err != nil {
		return err
	}
	filter.SetHandler(&prepareDescribeHandler{}) //TODO 这个地方要改啊
	return nil
}

func (h *prepareDescribeHandler) Handle(filter *QueryFilter, ctx services.Context) error {
	if ctx.Data().Forward != model.ServerToClient {
		return fmt.Errorf(" The error direction of the message is %T", model.ClientToServer)
	}
	reader := bytes.NewReader(ctx.Data().Buffer)
	context, ok := ctx.(*Context)
	if !ok {
		return fmt.Errorf("unknown context type: %T", ctx)
	}
	mges, err := model.UnpackSqliTransmission(reader)
	if err != nil {
		return fmt.Errorf("failed to unpack SqliInfo response packages, err: %T", err)
	}
	context.SetMetaData("sqliprepareDescribe", mges)
	b, err := ctx.MetaData("sqliprepareDescribe")
	if err != nil {
		return err
	}
	//describe
	describe := b.(model.SqliTransmission)[0].(model.SqliCommand).(*model.SqliDescribe)
	err = ctx.SetMetaData("Idnumber", int16(describe.StatementID))
	var fieldsname [1024]string
	for i := 0; i < len(describe.Fields); i++ {
		fieldsname[i] = describe.Fields[i].Name
	}
	err = ctx.SetMetaData("Fieldsname", fieldsname)

	buff := (&pgproto3.ParseComplete{}).Encode(nil)
	buff = (&pgproto3.ParameterDescription{
		ParameterOIDs: []uint32{23},
	}).Encode(buff)
	buff = (&pgproto3.NoData{}).Encode(buff)
	buff = (&pgproto3.ReadyForQuery{TxStatus: 'I'}).Encode(buff)
	ctx.Data().Buffer = buff
	context.front.Write(ctx.Data().Buffer)

	err = ctx.SetMetaData("QueryStage", "QueryPDescribeDone")
	filter.SetHandler(&cidesbatchHandler{})

	return nil
}

func (h *cidesbatchHandler) Handle(filter *QueryFilter, ctx services.Context) error {
	if ctx.Data().Forward != model.ClientToServer {
		return fmt.Errorf(" The error direction of the message is %T", model.ServerToClient)
	}
	//buff := bytes.NewBuffer(ctx.Data().Buffer)
	context, ok := ctx.(*Context)
	if !ok {
		return fmt.Errorf("unknown context type: %T", ctx)
	}
	parser := model.NewPgClientParser()
	_, err := parser.Append(ctx.Data().Buffer)
	if err != nil {
		return fmt.Errorf("failed to append parser buffer, err: %T", err)
	}
	msg, err := parser.ParseMessage()
	if err != nil {
		return fmt.Errorf("failed to unpack ParseMessage, err: %T", err)
	}
	err = context.SetMetaData("cidesbatch", msg)
	b, err := ctx.MetaData("cidesbatch")
	if err != nil {
		return err
	}
	groupcount := (len(b.(model.PgTransmission)) - 1) / 3
	err = ctx.SetMetaData("Groupcount", groupcount)
	if err != nil {
		return err
	}

	for j := 0; j < len(b.(model.PgTransmission))-1; j++ {
		for i := 0; i < groupcount; i++ {
			//bind
			pgbind := b.(model.PgTransmission)[j].(model.PgCommand).(*pgproto3.Bind)
			err = ctx.SetMetaData("Bindcondition", len(pgbind.ParameterFormatCodes))
			binddata := pgbind.Parameters
			a, err := ctx.MetaData("Bindtype")
			var datebindint uint16
			var datebindchar string
			for c, t := range a.([]uint32) {
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
			err = ctx.SetMetaData("Datebindint", datebindint)
			if err != nil {
				return err
			}
			err = ctx.SetMetaData("Datebindchar", datebindchar)
			if err != nil {
				return err
			}
			j += 3
		}
	}

	id := &model.SqliID{
		ID: 0,
	}
	cidescribe := &model.SqliCIdescribe{}
	eot := &model.SqliEot{}
	var transmission model.SqliTransmission
	transmission = []model.SqliCommand{id, cidescribe, eot}
	buff, err := transmission.Pack()
	ctx.Data().Buffer = buff
	context.backend.Write(ctx.Data().Buffer)
	err = ctx.SetMetaData("QueryStage", "QueryExecuteDone")
	filter.SetHandler(&executeHandler{})
	if err != nil {
		return err
	}

	v, err := ctx.MetaData("Condition")
	if v.(string) == "select" || v.(string) == "SELECT" {
		err = ctx.SetMetaData("QueryStage", "QueryOpen")
		filter.SetHandler(&cidescribeSelectHandler{})
		if err != nil {
			return err
		}
	} else {
		err = ctx.SetMetaData("QueryStage", "QueryExecute")
		filter.SetHandler(&cidescribeHandler{})
		if err != nil {
			return err
		}
	}

	return nil
}
