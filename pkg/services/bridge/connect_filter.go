package bridge

import (
	"bytes"
	"fmt"
	//"net"

	"github.com/jackc/pgproto3"
	"github.com/wangxuesong/tcpshadow/model"
	"github.com/wangxuesong/tcpshadow/pkg/services"
)

type (
	ConnectHandler interface {
		Handle(filter *ConnectFilter, ctx services.Context) error
	}

	ConnectFilter struct {
		handler ConnectHandler
	}

	startMessageHandler struct {
	}

	authResponseHandler struct {
	}

	protocolHandler struct {
	}

	infoHandler struct {
	}

	dbOpenHandler struct {
	}

	connectSet struct {
	}
)

func NewConnectFilter() *ConnectFilter {
	return &ConnectFilter{
		handler: &startMessageHandler{},
	}
}

func (c *ConnectFilter) Handle(ctx services.Context) error {
	return c.handler.Handle(c, ctx)
}

func (c *ConnectFilter) SetHandler(handler ConnectHandler) {
	c.handler = handler
}

func (h *startMessageHandler) Handle(filter *ConnectFilter, ctx services.Context) error {
	ClientToServre("start", ctx, "stratupmessage") //收
	authrequest := &model.AuthRequest{
		Noname1:    1,
		Noname2:    60,
		Noname3:    0,
		Noname4:    100,
		Noname5:    101,
		Noname6:    61,
		Ieeem:      "IEEEM",
		Noname7:    108,
		Sqlexec:    "sqlexec",
		Version:    "",
		Rds:        "RDS#R000000",
		Sqli:       "sqli",
		Noname8:    316,
		Noname9:    0,
		Noname10:   0,
		Noname11:   1,
		Clientname: "",
		Password:   "",
		Noname12:   "ol",
		Noname13:   61,
		Tlitcp:     "tlitcp",
		Noname14:   1,
		Noname15:   104,
		Asf:        11,
		Noname16:   3,
		Servername: "",
		Noname17:   0,
		Noname18:   0,
		Noname19:   0,
		Noname20:   0,
		Noname21:   0,
		Noname22:   106,
		Noname23:   6,
		Dpath: []model.DPath{{
			Dbpath:          "DBPATH",
			Dbpathattribute: ".",
		}, {
			Dbpath:          "CLNT_PAM_CAPABLE",
			Dbpathattribute: "1",
		}, {
			Dbpath:          "DBDATE",
			Dbpathattribute: "Y4MD-",
		}, {
			Dbpath:          "IFX_UPDDESC",
			Dbpathattribute: "1",
		}, {
			Dbpath:          "SQLMODE",
			Dbpathattribute: "gbase",
		}, {
			Dbpath:          "NODEFDAC",
			Dbpathattribute: "no",
		}},
		Noname24:         107,
		Noname25:         0,
		Noname26:         0,
		Longthreadid:     1,
		Noname27:         "bogon",
		Noname28:         0,
		Directory:        "E:\\JDBCTest\\JDBCTest",
		Noname29:         116,
		Appnamelengthall: 111,
		Noname30:         0,
		Noname31:         0,
		Appname:          "/E:/JDBCTest/JDBCTest/lib/gbasedbtjdbc_3.3.0_2.jarConnectionTest/Test",
		Asceot:           127,
	}
	tRequestBuilder := model.RequestBuilder{
		*authrequest,
	}
	authrequest = tRequestBuilder.BuildClientname("gbasedbt").
		BuildPassword("HmQOYC1ZfTYt+vlXUhkn3w==").
		BuildGbaseS("gbaseserver", 80).Create()
	pack, err := authrequest.Pack()
	if err != nil {
		return fmt.Errorf("failed to pack AuthRequest, err: %T", err)
	}
	SendPackage("backend", pack, ctx) //发
	ctx.SetMetaData("ConnectStage", "ConnectProtocol")
	filter.SetHandler(&authResponseHandler{})
	return nil
}

func (h *authResponseHandler) Handle(filter *ConnectFilter, ctx services.Context) error {
	ServerToClient("authresponse", ctx, "authresponse")
	//TODO:send protocols
	protocol := &model.SqliProtocols{
		Protocol: []byte{0xff, 0xfc, 0x7f, 0xfc, 0x3c, 0x8c, 0xaa, 0x97, 0x10},
	}
	eot := &model.SqliEot{}
	var transmission model.SqliTransmission
	transmission = []model.SqliCommand{protocol, eot}
	buf, err := transmission.Pack()
	if err != nil {
		return fmt.Errorf("failed to pack SqliProtocols transmission packages, err: %T", err)
	}
	SendPackage("backend", buf, ctx)
	ctx.SetMetaData("ConnectStage", "ConnectInfo")
	filter.SetHandler(&protocolHandler{})
	return nil
}

func (h *protocolHandler) Handle(filter *ConnectFilter, ctx services.Context) error {
	ServerToClient("sqli", ctx, "sqliinfo")
	//TODO: send SqliInfo
	info := &model.SqliInfo{
		MessageType: 6,
		Length:      38,
		InfoEnv: model.InfoEnv{
			NameLength:  12,
			ValueLength: 4,
			Env: map[string]string{
				"DBTEMP":      "/tmp",
				"SUBQCACHESZ": "10",
			},
		},
	}
	eot := &model.SqliEot{}
	var transmission model.SqliTransmission
	transmission = []model.SqliCommand{info, eot}
	buf, err := transmission.Pack()
	if err != nil {
		return fmt.Errorf("failed to pack SqliInfo transmission packages, err: %T", err)
	}
	SendPackage("backend", buf, ctx)
	ctx.SetMetaData("ConnectStage", "ConnectDbOpen")
	filter.SetHandler(&infoHandler{})
	return nil
}

func (h *infoHandler) Handle(filter *ConnectFilter, ctx services.Context) error {
	ServerToClient("sqli", ctx, "sqlieot")
	//TODO: send SqliDBOpen
	deopen := &model.SqliDBOpen{
		DBName: "dfe",
		Foo:    0,
	}
	eot := &model.SqliEot{}
	var transmission model.SqliTransmission
	transmission = []model.SqliCommand{deopen, eot}
	buf, err := transmission.Pack()
	if err != nil {
		return fmt.Errorf("failed to pack SqliDBOpen transmission packages, err: %T", err)
	}
	SendPackage("backend", buf, ctx)
	ctx.SetMetaData("ConnectStage", "ConnectDone")
	filter.SetHandler(&dbOpenHandler{})
	return nil
}

func (h *dbOpenHandler) Handle(filter *ConnectFilter, ctx services.Context) error {
	ServerToClient("sqli", ctx, "sqlidone")
	// TODO:send Authentication to front
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
	SendPackage("front", buf, ctx)
	ctx.SetMetaData("ConnectStage", "ConnectSet")
	filter.SetHandler(&connectSet{})
	return nil
}

func (h *connectSet) Handle(filter *ConnectFilter, ctx services.Context) error {
	ClientToServre("pg", ctx, "pbes")
	buf := make([]byte, 2048)
	//parse
	b, err := ctx.MetaData("pbes")
	buf = b.(model.PgTransmission)[0].Encode(nil)
	parse := &pgproto3.Parse{}
	parse.Decode(buf[5:])
	query := parse.Query
	tag := query[:3]
	set := query[4:15]

	parsecomplete := &pgproto3.ParseComplete{}
	bindcomplete := &pgproto3.BindComplete{}
	commandcomplete := &pgproto3.CommandComplete{CommandTag: tag}
	readyforquery := &pgproto3.ReadyForQuery{TxStatus: 'T'}
	context, ok := ctx.(*Context)
	if !ok {
		return fmt.Errorf("unknown context type: %T", ctx)
	}
	context.responses = []model.PgCommand{parsecomplete, bindcomplete, commandcomplete, readyforquery}
	buf, err = context.responses.Pack()
	if err != nil {
		return fmt.Errorf("failed to pack PgCompleteMessage, err: %T", err)
	}
	context.front.Write(buf)
	if set != "application" {
		err = ctx.SetMetaData("ConnectStage", "ConnectSet")
		if err != nil {
			return fmt.Errorf("failed to SetMetaData ConnectSet, err: %T", err)
		}
		filter.SetHandler(&connectSet{})
		return nil
	} else {
		err = ctx.SetMetaData("ConnectStage", EndStage)
		if err != nil {
			return fmt.Errorf("failed to SetMetaData EndStage, err: %T", err)
		}
		filter.SetHandler(&connectSet{})
	}

	v, err := context.MetaData("ConnectStage")
	if err != nil {
		return fmt.Errorf("failed to MetaData ConnectStage, err: %T", err)
	}
	stage, ok := v.(string)
	if !ok {
		return fmt.Errorf("mistake metadata type: %T", v)
	}
	if ok && stage == EndStage {
		context.state = QueryState
	}
	return nil
}

func ClientToServre(name string, ctx services.Context, tag string) error {
	if ctx.Data().Forward != model.ClientToServer {
		return fmt.Errorf(" The error direction of the message is %T", model.ServerToClient)
	}
	buff := bytes.NewBuffer(ctx.Data().Buffer)
	context, ok := ctx.(*Context)
	if !ok {
		return fmt.Errorf("unknown context type: %T", ctx)
	}

	switch name {
	case "start":
		backend, err := pgproto3.NewBackend(pgproto3.NewChunkReader(buff), nil)
		if err != nil {
			return fmt.Errorf("failed to generate NewBackend, err: %T", err)
		}
		msg, err := backend.ReceiveStartupMessage()
		if err != nil {
			return fmt.Errorf("failed to receive StartupMessage, err: %T", err)
		}
		err = context.SetMetaData(tag, msg)
		context.requests = []model.PgCommand{msg}
	case "pg":
		parser := model.NewPgClientParser()
		_, err := parser.Append(ctx.Data().Buffer)
		if err != nil {
			return fmt.Errorf("failed to append parser buffer, err: %T", err)
		}
		msg, err := parser.ParseMessage()
		if err != nil {
			return fmt.Errorf("failed to unpack ParseMessage, err: %T", err)
		}
		err = context.SetMetaData(tag, msg)
	}
	return nil
}

func ServerToClient(name string, ctx services.Context, tag string) error {
	if ctx.Data().Forward != model.ServerToClient {
		return fmt.Errorf(" The error direction of the message is %T", model.ClientToServer)
	}
	reader := bytes.NewReader(ctx.Data().Buffer)
	context, ok := ctx.(*Context)
	if !ok {
		return fmt.Errorf("unknown context type: %T", ctx)
	}
	switch name {
	case "authresponse":
		authreponse := &model.AuthResponse{}
		authreponse.Unpack(reader)
	case "authrequest":
		authrequest := &model.AuthRequest{}
		authrequest.Unpack(reader)
	case "sqli":
		mges, err := model.UnpackSqliTransmission(reader)
		if err != nil {
			return fmt.Errorf("failed to unpack SqliInfo response packages, err: %T", err)
		}
		context.SetMetaData(tag, mges)
	default:
		return nil
	}

	return nil
}

func SendPackage(dir string, pack []byte, ctx services.Context) error {
	context, ok := ctx.(*Context)
	if !ok {
		return fmt.Errorf("unknown context type: %T", ctx)
	}
	ctx.Data().Buffer = pack
	switch dir {
	case "backend":
		context.backend.Write(ctx.Data().Buffer)
	case "front":
		context.front.Write(ctx.Data().Buffer)
	default:
		return nil
	}
	return nil
}
