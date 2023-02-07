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
	ClientToServre("pg", ctx, "parse")
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
	ServerToClient("sqli", ctx, "sqliprepare")
	//TODO: receive Sqli
	buf := make([]byte, 1024)
	b, err := ctx.MetaData("sqliprepare")
	//describe
	buf, err = b.(model.SqliTransmission)[0].Pack()
	if err != nil {
		return err
	}
	reader := bytes.NewReader(buf[2:])
	describe := &model.SqliDescribe{}
	err = describe.Unpack(reader)
	if err != nil {
		return err
	}
	err = ctx.SetMetaData("Idnumber", int16(describe.StatementID))
	var fieldsname [1024]string
	for i := 0; i < len(describe.Fields); i++ {
		fieldsname[i] = describe.Fields[i].Name
	}
	err = ctx.SetMetaData("Fieldsname", fieldsname)

	//done
	buf, err = b.(model.SqliTransmission)[1].Pack()
	if err != nil {
		return err
	}
	reader = bytes.NewReader(buf[2:])
	done := &model.SqliDone{}
	err = done.Unpack(reader)
	if err != nil {
		return err
	}

	//cost
	buf, err = b.(model.SqliTransmission)[2].Pack()
	if err != nil {
		return err
	}
	reader = bytes.NewReader(buf[2:])
	cost := &model.SqliCost{}
	err = cost.Unpack(reader)
	if err != nil {
		return err
	}

	//eot
	buf, err = b.(model.SqliTransmission)[3].Pack()

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
	SendPackage("backend", buf, ctx)

	context, ok := ctx.(*Context)
	if !ok {
		return fmt.Errorf("unknown context type: %T", ctx)
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
	ServerToClient("sqli", ctx, "sqlicidescribe")
	buf := make([]byte, 1024)
	context, ok := ctx.(*Context)
	if !ok {
		return fmt.Errorf("unknown context type: %T", ctx)
	}
	groupcount, err := context.MetaData("Groupcount")
	m, err := ctx.MetaData("sqlicidescribe")

	//idescribe
	buf, err = m.(model.SqliTransmission)[0].Pack()
	if err != nil {
		return err
	}
	reader := bytes.NewReader(buf[2:])
	idescribe := &model.SqliIdescribe{}
	err = idescribe.Unpack(reader)
	if err != nil {
		return err
	}

	//TODO: send Sqli
	b, err := context.MetaData("Idnumber")
	id := &model.SqliID{
		ID: b.(int16),
	}
	bind := &model.SqliBind{
		Columns: []model.BindColumn{},
	}
	v, err := context.MetaData("Bindcondition")
	if v != 0 {
		a, err := context.MetaData("Bindtype")
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
				t, err := context.MetaData("Datebindint")
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
				p, err := context.MetaData("Datebindchar")
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
	SendPackage("backend", buf, ctx)
	err = ctx.SetMetaData("QueryStage", "QueryExecuteDone")
	filter.SetHandler(&executeHandler{})
	if err != nil {
		return err
	}
	return nil
}

func (h *executeHandler) Handle(filter *QueryFilter, ctx services.Context) error {
	context, ok := ctx.(*Context)
	if !ok {
		return fmt.Errorf("unknown context type: %T", ctx)
	}
	groupcount, err := context.MetaData("Groupcount")
	ServerToClient("sqli", ctx, "sqliexecute")
	buf := make([]byte, 1024)
	m, err := ctx.MetaData("sqliexecute")
	for j := 0; j < len(m.(model.SqliTransmission))-1; j++ {
		for i := 0; i < groupcount.(int); i++ {
			//insertdone
			buf, err = m.(model.SqliTransmission)[0].Pack()
			if err != nil {
				return err
			}
			reader := bytes.NewReader(buf[2:])
			insertdone := &model.SqliInsertDone{}
			err = insertdone.Unpack(reader)
			if err != nil {
				return err
			}
			j++

			//done
			buf, err = m.(model.SqliTransmission)[1].Pack()
			if err != nil {
				return err
			}
			reader = bytes.NewReader(buf[2:])
			done := &model.SqliDone{}
			err = done.Unpack(reader)
			if err != nil {
				return err
			}
			j++

			//cost
			buf, err = m.(model.SqliTransmission)[2].Pack()
			if err != nil {
				return err
			}
			reader = bytes.NewReader(buf[2:])
			cost := &model.SqliCost{}
			err = cost.Unpack(reader)
			if err != nil {
				return err
			}
			j++
		}
	}

	//TODO: send Sqli
	b, err := context.MetaData("Idnumber")
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
	SendPackage("backend", buf, ctx)
	err = ctx.SetMetaData("QueryStage", "QueryRelease")
	if err != nil {
		return err
	}
	filter.SetHandler(&releaseHandler{})
	return nil
}

func (h *releaseHandler) Handle(filter *QueryFilter, ctx services.Context) error {
	ServerToClient("sqli", ctx, "sqlirelease")
	context, ok := ctx.(*Context)
	if !ok {
		return fmt.Errorf("unknown context type: %T", ctx)
	}
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

	SendPackage("front", buff, ctx)

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
	ServerToClient("sqli", ctx, "sqlicidescribeSelect")
	context, ok := ctx.(*Context)
	if !ok {
		return fmt.Errorf("unknown context type: %T", ctx)
	}
	buf := make([]byte, 1024)
	m, err := ctx.MetaData("sqlicidescribeSelect")
	//idescribe
	buf, err = m.(model.SqliTransmission)[0].Pack()
	if err != nil {
		return err
	}
	reader := bytes.NewReader(buf[2:])
	idescribe := &model.SqliIdescribe{}
	err = idescribe.Unpack(reader)
	if err != nil {
		return err
	}
	err = ctx.SetMetaData("Idesfields", idescribe.Fields)

	//TODO: send Sqli
	b, err := context.MetaData("Idnumber")
	id := &model.SqliID{
		ID: b.(int16),
	}
	curname := &model.SqliCurName{
		CurName: "_ifxc0000000000000",
	}
	bind := &model.SqliBind{
		Columns: []model.BindColumn{},
	}
	v, err := context.MetaData("Bindcondition")
	if v != 0 {
		a, err := context.MetaData("Bindtype")
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
				t, err := context.MetaData("Datebindint")
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
				p, err := context.MetaData("Datebindchar")
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
	buf, err = transmission.Pack()
	if err != nil {
		return err
	}
	SendPackage("backend", buf, ctx)
	err = ctx.SetMetaData("QueryStage", "QueryOpen")
	if err != nil {
		return err
	}
	filter.SetHandler(&openHandler{})
	return nil
}

func (h *openHandler) Handle(filter *QueryFilter, ctx services.Context) error {
	ServerToClient("sqli", ctx, "sqliopen")
	context, ok := ctx.(*Context)
	if !ok {
		return fmt.Errorf("unknown context type: %T", ctx)
	}

	//TODO: send Sqli
	b, err := context.MetaData("Idnumber")
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
	SendPackage("backend", buff, ctx)
	err = ctx.SetMetaData("QueryStage", "QueryDone")
	if err != nil {
		return err
	}
	filter.SetHandler(&nfetchdoneHandler{})
	return nil
}

func (h *nfetchdoneHandler) Handle(filter *QueryFilter, ctx services.Context) error {
	ServerToClient("sqli", ctx, "sqlinfetchdone")
	context, ok := ctx.(*Context)
	if !ok {
		return fmt.Errorf("unknown context type: %T", ctx)
	}
	n, err := ctx.MetaData("sqlinfetchdone")
	buf := make([]byte, 1024)
	err = ctx.SetMetaData("TupleNumber", len(n.(model.SqliTransmission))-3)

	//tuple
	var tuple [1024]model.SqliTuple
	t, err := context.MetaData("TupleNumber")
	for i := 0; i < t.(int); i++ {
		buf, err = n.(model.SqliTransmission)[i].Pack()
		if err != nil {
			return err
		}
		reader := bytes.NewReader(buf)
		err = tuple[i].Unpack(reader)
		if err != nil {
			return err
		}
	}
	err = ctx.SetMetaData("Tuple", tuple)

	//done
	buf, err = n.(model.SqliTransmission)[t.(int)].Pack()
	if err != nil {
		return err
	}
	reader := bytes.NewReader(buf[2:])
	done := &model.SqliDone{}
	err = done.Unpack(reader)
	if err != nil {
		return err
	}

	//cost
	buf, err = n.(model.SqliTransmission)[t.(int)+1].Pack()
	if err != nil {
		return err
	}
	reader = bytes.NewReader(buf[2:])
	cost := &model.SqliCost{}
	err = cost.Unpack(reader)
	if err != nil {
		return err
	}

	//TODO: send pg
	buff := (&pgproto3.ParseComplete{}).Encode(nil)
	buff = (&pgproto3.BindComplete{}).Encode(buff)
	m, err := context.MetaData("Fieldsname")
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
	p, err := context.MetaData("Data")
	row := p.([1024]pgproto3.DataRow)
	buff = row[0].Encode(buff)
	buff = row[1].Encode(buff)
	buff = row[2].Encode(buff)
	buff = (&pgproto3.CommandComplete{CommandTag: "SELECT 1"}).Encode(buff)
	buff = (&pgproto3.ReadyForQuery{TxStatus: 'I'}).Encode(buff)
	SendPackage("front", buff, ctx)
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

func (h *parseHandler) HandleHandle1(filter *QueryFilter, ctx services.Context) error {
	context, ok := ctx.(*Context)
	if !ok {
		return fmt.Errorf("unknown context type: %T", ctx)
	}
	buf := make([]byte, 2048)
	b, err := ctx.MetaData("parse")
	if err != nil {
		return err
	}
	//parse
	buf = b.(model.PgTransmission)[0].Encode(nil)
	parse := &pgproto3.Parse{}
	err = parse.Decode(buf[5:])
	if err != nil {
		return err
	}
	sql := parse.Query
	sql = strings.ReplaceAll(sql, "$1", "?")
	err = ctx.SetMetaData("Condition", sql[:6])
	err = ctx.SetMetaData("Bindtype", parse.ParameterOIDs)
	paranumber := len(parse.ParameterOIDs)

	//bind
	buf = b.(model.PgTransmission)[1].Encode(nil)
	bind := &pgproto3.Bind{}
	err = bind.Decode(buf[5:])
	if err != nil {
		return err
	}
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

	//describe
	buf = b.(model.PgTransmission)[2].Encode(nil)
	describe := &pgproto3.Describe{}
	err = describe.Decode(buf[5:])
	if err != nil {
		return err
	}

	//execute
	buf = b.(model.PgTransmission)[3].Encode(nil)
	execute := &pgproto3.Execute{}
	err = execute.Decode(buf[5:])
	if err != nil {
		return err
	}

	//sync
	buf = b.(model.PgTransmission)[4].Encode(nil)
	sync := &pgproto3.Sync{}
	err = sync.Decode(buf[5:])
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
	SendPackage("backend", buffer, ctx)
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
	buf := make([]byte, 2048)
	b, err := ctx.MetaData("parse")
	if err != nil {
		return err
	}

	//parse
	buf = b.(model.PgTransmission)[0].Encode(nil)
	parse := &pgproto3.Parse{}
	err = parse.Decode(buf[5:])
	if err != nil {
		return err
	}
	sql := parse.Query
	sql = strings.ReplaceAll(sql, "$1", "?")
	err = ctx.SetMetaData("Condition", sql[:6])
	err = ctx.SetMetaData("Bindtype", parse.ParameterOIDs)
	paranumber := len(parse.ParameterOIDs)

	//describe
	buf = b.(model.PgTransmission)[1].Encode(nil)
	describe := &pgproto3.Describe{}
	err = describe.Decode(buf[5:])
	if err != nil {
		return err
	}
	_ = describe.ObjectType
	_ = describe.Name

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
	SendPackage("backend", buffer, ctx)
	err = ctx.SetMetaData("QueryStage", "QueryPrepareDone")
	if err != nil {
		return err
	}
	filter.SetHandler(&prepareDescribeHandler{}) //TODO 这个地方要改啊
	return nil
}

func (h *prepareDescribeHandler) Handle(filter *QueryFilter, ctx services.Context) error {
	ServerToClient("sqli", ctx, "sqliprepareDescribe")
	buf := make([]byte, 1024)
	b, err := ctx.MetaData("sqliprepareDescribe")
	//describe
	buf, err = b.(model.SqliTransmission)[0].Pack()
	if err != nil {
		return err
	}
	reader := bytes.NewReader(buf[2:])
	describe := &model.SqliDescribe{}
	err = describe.Unpack(reader)
	if err != nil {
		return err
	}
	err = ctx.SetMetaData("Idnumber", int16(describe.StatementID))
	var fieldsname [1024]string
	for i := 0; i < len(describe.Fields); i++ {
		fieldsname[i] = describe.Fields[i].Name
	}
	err = ctx.SetMetaData("Fieldsname", fieldsname)

	//done
	buf, err = b.(model.SqliTransmission)[1].Pack()
	if err != nil {
		return err
	}
	reader = bytes.NewReader(buf[2:])
	done := &model.SqliDone{}
	err = done.Unpack(reader)
	if err != nil {
		return err
	}

	//cost
	buf, err = b.(model.SqliTransmission)[2].Pack()
	if err != nil {
		return err
	}
	reader = bytes.NewReader(buf[2:])
	cost := &model.SqliCost{}
	err = cost.Unpack(reader)
	if err != nil {
		return err
	}

	buff := (&pgproto3.ParseComplete{}).Encode(nil)
	buff = (&pgproto3.ParameterDescription{
		ParameterOIDs: []uint32{23},
	}).Encode(buff)
	buff = (&pgproto3.NoData{}).Encode(buff)
	buff = (&pgproto3.ReadyForQuery{TxStatus: 'I'}).Encode(buff)
	SendPackage("front", buff, ctx)

	err = ctx.SetMetaData("QueryStage", "QueryPDescribeDone")
	filter.SetHandler(&cidesbatchHandler{})

	return nil
}

func (h *cidesbatchHandler) Handle(filter *QueryFilter, ctx services.Context) error {
	context, ok := ctx.(*Context)
	if !ok {
		return fmt.Errorf("unknown context type: %T", ctx)
	}
	ClientToServre("pg", ctx, "cidesbatch")
	buf := make([]byte, 1024)
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
			buf = b.(model.PgTransmission)[j].Encode(nil)
			pgbind := &pgproto3.Bind{}
			err = pgbind.Decode(buf[5:])
			if err != nil {
				return err
			}
			err = ctx.SetMetaData("Bindcondition", len(pgbind.ParameterFormatCodes))
			binddata := pgbind.Parameters
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
			j++

			//describe
			buf = b.(model.PgTransmission)[j].Encode(nil)
			describe := &pgproto3.Describe{}
			err = describe.Decode(buf[5:])
			if err != nil {
				return err
			}
			j++

			//execute
			buf = b.(model.PgTransmission)[j].Encode(nil)
			pgexecute := &pgproto3.Execute{}
			err = pgexecute.Decode(buf[5:])
			if err != nil {
				return err
			}
			j++
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
	SendPackage("backend", buff, ctx)
	err = ctx.SetMetaData("QueryStage", "QueryExecuteDone")
	filter.SetHandler(&executeHandler{})
	if err != nil {
		return err
	}

	v, err := context.MetaData("Condition")
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
