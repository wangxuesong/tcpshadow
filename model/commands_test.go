package model

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"io"
	"testing"
)

func TestSqTuple_Pack_Unpack(t *testing.T) {
	expect := []byte{0, 14, 0, 0, 0, 0, 0, 15, 0, 11, 0, 0, 0, 0, 8, 115, 119, 101, 101, 116, 104, 117, 105, 0}

	intvalue := &SmallIntTupleValue{Value: 11}
	varcharvalue := &LVarcharTupleValue{Value: "sweethui"}
	tuple := SqliTuple{
		Warnings: 0,
	}
	tuple.Values = append(tuple.Values, intvalue)
	tuple.Values = append(tuple.Values, varcharvalue)
	actual, err := tuple.Pack()
	assert.NoError(t, err)
	assert.NotNil(t, actual)
	assert.Equal(t, expect, actual)

	buffer := bytes.NewReader(expect)
	cmd, err := UnpackSqliCommand(buffer)
	assert.NoError(t, err)
	assert.IsType(t, &SqliTuple{}, cmd)
	tupleCmd, ok := cmd.(*SqliTuple)
	assert.True(t, ok)
	fields := make([]SqliField, 0, 2)
	field1 := SqliField{
		FieldIndex:     0,
		ColumnStartPos: 0,
		ColumnType:     1,
		Length:         4,
		Name:           "a",
	}
	fields = append(fields, field1)
	field2 := SqliField{
		FieldIndex:     2,
		ColumnStartPos: 4,
		ColumnType:     43,
		Length:         50,
		Name:           "b",
	}
	fields = append(fields, field2)
	tupleCmd.SetMetaData(fields)
	tupleCmd.UnpackValues()
	assert.Equal(t, tuple.Warnings, tupleCmd.Warnings)
	assert.Equal(t, tuple.Values, tupleCmd.Values)
	pos, err := buffer.Seek(0, io.SeekCurrent)
	assert.NoError(t, err)
	assert.EqualValues(t, pos, len(expect))
}

func TestSqliDescribe_Pack_Unpack(t *testing.T) {
	expect := []byte{0, 8, 0, 2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 56, 0, 3,
		0, 0, 0, 6, 0, 0, 0, 0, 0, 0, 0, 0, 0, 2, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 4, 0, 0, 0, 2, 0, 0, 0, 4, 0, 13, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 50,
		0, 0, 0, 4, 0, 0, 0, 56, 0, 41, 0, 0, 0, 11, 0, 8,
		103, 98, 97, 115, 101, 100, 98, 116, 0, 4, 99, 108,
		111, 98, 0, 1, 0, 4, 0, 0, 0, 0, 0, 0, 0, 72, 97, 0,
		98, 0, 99, 0}
	desc := &SqliDescribe{
		StatementType: 2,
		StatementID:   0,
		EstimatedCost: 0,
		TupleSize:     56,
	}
	field1 := SqliField{
		FieldIndex:     0,
		ColumnStartPos: 0,
		ColumnType:     2,
		Length:         4,
		Name:           "a",
	}
	desc.AppendFields(field1)
	field2 := SqliField{
		FieldIndex:     2,
		ColumnStartPos: 4,
		ColumnType:     13,
		Length:         50,
		Name:           "b",
	}
	desc.AppendFields(field2)
	field3 := SqliField{
		FieldIndex:              4,
		ColumnStartPos:          56,
		ColumnType:              41,
		ColumnExtendedBuiltinId: 11,
		OwnerName:               "gbasedbt",
		ExtendedName:            "clob",
		Reference:               1,
		Alignment:               4,
		Length:                  72,
		Name:                    "c",
	}
	desc.AppendFields(field3)

	buf, err := desc.Pack()
	assert.NoError(t, err)
	assert.Equal(t, expect, buf)

	buffer := bytes.NewReader(expect)
	cmd, err := UnpackSqliCommand(buffer)
	assert.NoError(t, err)
	assert.Equal(t, desc, cmd)
	pos, err := buffer.Seek(0, io.SeekCurrent)
	assert.NoError(t, err)
	assert.EqualValues(t, pos, len(expect))
}

func TestSqliDone_Pack_Unpack(t *testing.T) {
	done := &SqliDone{}
	buf, err := done.Pack()
	expect := []byte{0, 15, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	assert.NoError(t, err)
	assert.Equal(t, expect, buf)

	buffer := bytes.NewReader(expect)
	cmd, err := UnpackSqliCommand(buffer)
	assert.NoError(t, err)
	assert.Equal(t, done, cmd)
	pos, err := buffer.Seek(0, io.SeekCurrent)
	assert.NoError(t, err)
	assert.EqualValues(t, pos, len(expect))
}

func TestSqliCost_Pack_Unpack(t *testing.T) {
	cost := &SqliCost{EstimatedRows: 1, EstimatedIO: 1}
	buf, err := cost.Pack()
	expect := []byte{0, 55, 0, 0, 0, 1, 0, 0, 0, 1}
	assert.NoError(t, err)
	assert.Equal(t, expect, buf)

	buffer := bytes.NewReader(expect)
	cmd, err := UnpackSqliCommand(buffer)
	assert.NoError(t, err)
	assert.Equal(t, cost, cmd)
	pos, err := buffer.Seek(0, io.SeekCurrent)
	assert.NoError(t, err)
	assert.EqualValues(t, pos, len(expect))
}

func TestSqliTransmission_Pack_Unpack(t *testing.T) {
	expect := []byte{0, 8, 0, 2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 55, 0, 2, 0, 0, 0, 4, 0, 0, 0, 0, 0, 0, 0, 0, 0, 2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 4, 0, 0, 0, 2, 0, 0, 0, 4, 0, 13, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 50, 97, 0, 98, 0, 0, 15, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 55, 0, 0, 0, 1, 0, 0, 0, 1, 0, 12}
	size := len(expect)
	trans, err := NewDescribeTransmission()
	assert.NoError(t, err)
	assert.Equal(t, 4, len(trans))
	assert.EqualValues(t, 8, trans[0].Command())  // SQ_DESCRIBE
	assert.EqualValues(t, 15, trans[1].Command()) // SQ_DONE
	assert.EqualValues(t, 55, trans[2].Command()) // SQ_COST
	assert.EqualValues(t, 12, trans[3].Command()) // SQ_EOT
	buf, err := trans.Pack()
	assert.NoError(t, err)
	assert.Equal(t, expect, buf)
	buf, err = trans[3].Pack()
	assert.NoError(t, err)
	assert.Equal(t, expect[size-2:], buf)

	buffer := bytes.NewReader(expect)
	cmds, err := UnpackSqliTransmission(buffer)
	assert.NoError(t, err)
	assert.Equal(t, trans, cmds)
	pos, err := buffer.Seek(0, io.SeekCurrent)
	assert.NoError(t, err)
	assert.EqualValues(t, pos, len(expect))
}

func TestSqliPrepare_Pack_Unpack(t *testing.T) {
	expect := []byte{0, 2, 0, 0, 0, 0, 0, 29, 115, 101, 108, 101, 99, 116, 32, 42, 32, 102, 114, 111, 109, 32, 115, 116, 111, 114, 101, 115, 95, 100, 101, 109, 111, 58, 116, 97, 98, 0}
	prepare := &SqliPrepare{QMarks: 0, Sql: "select * from stores_demo:tab"}
	buf, err := prepare.Pack()
	assert.NoError(t, err)
	assert.Equal(t, expect, buf)

	buffer := bytes.NewReader(expect)
	cmd, err := UnpackSqliCommand(buffer)
	assert.NoError(t, err)
	assert.Equal(t, prepare, cmd)
	pos, err := buffer.Seek(0, io.SeekCurrent)
	assert.NoError(t, err)
	assert.EqualValues(t, pos, len(expect))
}

func TestSqliProtocols_Pack_Unpack(t *testing.T) {
	expect := []byte{0, 126, 0, 9, 189, 190, 159, 254, 127, 183, 255, 239, 240, 0}
	protocol := &SqliProtocols{Protocol: []byte{189, 190, 159, 254, 127, 183, 255, 239, 240}}
	buf, err := protocol.Pack()
	assert.NoError(t, err)
	assert.Equal(t, expect, buf)

	buffer := bytes.NewReader(expect)
	cmd, err := UnpackSqliCommand(buffer)
	assert.NoError(t, err)
	assert.Equal(t, protocol, cmd)
	pos, err := buffer.Seek(0, io.SeekCurrent)
	assert.NoError(t, err)
	assert.EqualValues(t, pos, len(expect))
}

func TestSqliInfo_Pack_Unpack(t *testing.T) {
	expect := []byte{0, 81, 0, 6, 0, 38, 0, 12, 0, 4, 0, 6, 68, 66, 84, 69, 77, 80, 0, 4, 47, 116, 109, 112, 0, 11, 83, 85, 66, 81, 67, 65, 67, 72, 69, 83, 90, 0, 0, 2, 49, 48, 0, 0, 0, 0}
	info := &SqliInfo{MessageType: 6, Length: 38, InfoEnv: InfoEnv{}}
	info.InfoEnv.NameLength = 12
	info.InfoEnv.ValueLength = 4
	info.InfoEnv.Env = make(map[string]string)
	info.InfoEnv.Env["DBTEMP"] = "/tmp"
	info.InfoEnv.Env["SUBQCACHESZ"] = "10"
	buf, err := info.Pack()
	assert.NoError(t, err)
	assert.Equal(t, expect, buf)

	buffer := bytes.NewReader(expect)
	cmd, err := UnpackSqliCommand(buffer)
	assert.NoError(t, err)
	assert.Equal(t, info, cmd)
	pos, err := buffer.Seek(0, io.SeekCurrent)
	assert.NoError(t, err)
	assert.EqualValues(t, pos, len(expect))

}

func TestSqliExit_Pack_Unpack(t *testing.T) {
	exit := &SqliExit{}
	buf, err := exit.Pack()
	expect := []byte{0, 56}
	assert.NoError(t, err)
	assert.Equal(t, expect, buf)

	buffer := bytes.NewReader(expect)
	cmd, err := UnpackSqliCommand(buffer)
	assert.NoError(t, err)
	assert.Equal(t, exit, cmd)
	pos, err := buffer.Seek(0, io.SeekCurrent)
	assert.NoError(t, err)
	assert.EqualValues(t, pos, len(expect))
}

func TestSqliNDescribe_Pack_Unpack(t *testing.T) {
	nDesc := &SqliNDescribe{}
	buf, err := nDesc.Pack()
	expect := []byte{0, 22}
	assert.NoError(t, err)
	assert.Equal(t, expect, buf)

	buffer := bytes.NewReader(expect)
	cmd, err := UnpackSqliCommand(buffer)
	assert.NoError(t, err)
	assert.Equal(t, nDesc, cmd)
	pos, err := buffer.Seek(0, io.SeekCurrent)
	assert.NoError(t, err)
	assert.EqualValues(t, pos, len(expect))
}

func TestSqliWantDone_Pack_Unpack(t *testing.T) {
	wantDone := &SqliWantDone{}
	buf, err := wantDone.Pack()
	expect := []byte{0, 49}
	assert.NoError(t, err)
	assert.Equal(t, expect, buf)

	buffer := bytes.NewReader(expect)
	cmd, err := UnpackSqliCommand(buffer)
	assert.NoError(t, err)
	assert.Equal(t, wantDone, cmd)
	pos, err := buffer.Seek(0, io.SeekCurrent)
	assert.NoError(t, err)
	assert.EqualValues(t, pos, len(expect))
}

func TestSqliDBOpen_Pack_Unpack(t *testing.T) {
	dbOpen := &SqliDBOpen{DBName: "test", Foo: 0}
	buf, err := dbOpen.Pack()
	expect := []byte{00, 0x24, 0x00, 0x04, 0x74, 0x65, 0x73, 0x74, 0x00, 0x00}
	assert.NoError(t, err)
	assert.Equal(t, expect, buf)

	buffer := bytes.NewReader(expect)
	cmd, err := UnpackSqliCommand(buffer)
	assert.NoError(t, err)
	assert.Equal(t, dbOpen, cmd)
	pos, err := buffer.Seek(0, io.SeekCurrent)
	assert.NoError(t, err)
	assert.EqualValues(t, pos, len(expect))
}

func TestSqliID_Pack_Unpack(t *testing.T) {
	dbOpen := &SqliID{}
	buf, err := dbOpen.Pack()
	expect := []byte{0, 4, 00, 00}
	assert.NoError(t, err)
	assert.Equal(t, expect, buf)

	buffer := bytes.NewReader(expect)
	cmd, err := UnpackSqliCommand(buffer)
	assert.NoError(t, err)
	assert.Equal(t, dbOpen, cmd)
	pos, err := buffer.Seek(0, io.SeekCurrent)
	assert.NoError(t, err)
	assert.EqualValues(t, pos, len(expect))
}

func TestSqliRelease_Pack_Unpack(t *testing.T) {
	release := &SqliRelease{}
	buf, err := release.Pack()
	expect := []byte{0, 11}
	assert.NoError(t, err)
	assert.Equal(t, expect, buf)

	buffer := bytes.NewReader(expect)
	cmd, err := UnpackSqliCommand(buffer)
	assert.NoError(t, err)
	assert.Equal(t, release, cmd)
	pos, err := buffer.Seek(0, io.SeekCurrent)
	assert.NoError(t, err)
	assert.EqualValues(t, pos, len(expect))
}

func TestSqliClose_Pack_Unpack(t *testing.T) {
	sqliClose := &SqliClose{}
	buf, err := sqliClose.Pack()
	expect := []byte{0, 10}
	assert.NoError(t, err)
	assert.Equal(t, expect, buf)

	buffer := bytes.NewReader(expect)
	cmd, err := UnpackSqliCommand(buffer)
	assert.NoError(t, err)
	assert.Equal(t, sqliClose, cmd)
	pos, err := buffer.Seek(0, io.SeekCurrent)
	assert.NoError(t, err)
	assert.EqualValues(t, pos, len(expect))
}

func TestSqliOpen_Pack_Unpack(t *testing.T) {
	sqliOpen := &SqliOpen{}
	buf, err := sqliOpen.Pack()
	expect := []byte{0, 6}
	assert.NoError(t, err)
	assert.Equal(t, expect, buf)

	buffer := bytes.NewReader(expect)
	cmd, err := UnpackSqliCommand(buffer)
	assert.NoError(t, err)
	assert.Equal(t, sqliOpen, cmd)
	pos, err := buffer.Seek(0, io.SeekCurrent)
	assert.NoError(t, err)
	assert.EqualValues(t, pos, len(expect))
}

func TestSqliCurName_Pack_Unpack(t *testing.T) {
	curName := &SqliCurName{CurName: "_ifxc0000000000008"}
	buf, err := curName.Pack()
	expect := []byte{0, 3, 0, 18, 95, 105, 102, 120, 99, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 56}
	assert.NoError(t, err)
	assert.Equal(t, expect, buf)

	buffer := bytes.NewReader(expect)
	cmd, err := UnpackSqliCommand(buffer)
	assert.NoError(t, err)
	assert.Equal(t, curName, cmd)
	pos, err := buffer.Seek(0, io.SeekCurrent)
	assert.NoError(t, err)
	assert.EqualValues(t, pos, len(expect))
}

func TestSqliNFetch_Pack_Unpack(t *testing.T) {
	sqliNFetch := &SqliNFetch{TupleBufferSize: 4096, FetchArraySize: 0}
	buf, err := sqliNFetch.Pack()
	expect := []byte{0, 9, 0, 0, 16, 0, 0, 0}
	assert.NoError(t, err)
	assert.Equal(t, expect, buf)

	buffer := bytes.NewReader(expect)
	cmd, err := UnpackSqliCommand(buffer)
	assert.NoError(t, err)
	assert.Equal(t, sqliNFetch, cmd)
	pos, err := buffer.Seek(0, io.SeekCurrent)
	assert.NoError(t, err)
	assert.EqualValues(t, pos, len(expect))
}

func TestSqliCmd_Pack_Unpack(t *testing.T) {
	sqliCmd := &SqliCmd{Sql: "insert into x values (22, 'sweet');"}
	buf, err := sqliCmd.Pack()
	expect := []byte{0, 1, 0, 0, 0, 0, 0, 35, 105, 110, 115, 101, 114, 116, 32, 105, 110, 116, 111, 32, 120, 32, 118, 97, 108, 117, 101, 115, 32, 40, 50, 50, 44, 32, 39, 115, 119, 101, 101, 116, 39, 41, 59, 0}
	assert.NoError(t, err)
	assert.Equal(t, expect, buf)

	buffer := bytes.NewReader(expect)
	cmd, err := UnpackSqliCommand(buffer)
	assert.NoError(t, err)
	assert.Equal(t, sqliCmd, cmd)
	pos, err := buffer.Seek(0, io.SeekCurrent)
	assert.NoError(t, err)
	assert.EqualValues(t, pos, len(expect))
}

func TestSqliExecute_Pack_Unpack(t *testing.T) {
	sqliExecute := &SqliExecute{}
	buf, err := sqliExecute.Pack()
	expect := []byte{0, 7}
	assert.NoError(t, err)
	assert.Equal(t, expect, buf)

	buffer := bytes.NewReader(expect)
	cmd, err := UnpackSqliCommand(buffer)
	assert.NoError(t, err)
	assert.Equal(t, sqliExecute, cmd)
	pos, err := buffer.Seek(0, io.SeekCurrent)
	assert.NoError(t, err)
	assert.EqualValues(t, pos, len(expect))
}

func TestSqliInsertDone_Pack_Unpack(t *testing.T) {
	sqliInsertDone := &SqliInsertDone{Serial8: 0, BigSerial: 0}
	buf, err := sqliInsertDone.Pack()
	expect := []byte{0, 94, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	assert.NoError(t, err)
	assert.Equal(t, expect, buf)

	buffer := bytes.NewReader(expect)
	cmd, err := UnpackSqliCommand(buffer)
	assert.NoError(t, err)
	assert.Equal(t, sqliInsertDone, cmd)
	pos, err := buffer.Seek(0, io.SeekCurrent)
	assert.NoError(t, err)
	assert.EqualValues(t, pos, len(expect))
}

func TestSqliRetType_Pack_Unpack(t *testing.T) {
	sqliRetType := &SqliRetType{Direction: 1, Columns: make([]ColumnType, 0)}
	column := ColumnType{Type: 13, Length: 128}
	sqliRetType.Columns = append(sqliRetType.Columns, column)
	buf, err := sqliRetType.Pack()
	expect := []byte{0, 100, 0, 1, 0, 1, 0, 13, 0, 0, 0, 128}
	assert.NoError(t, err)
	assert.Equal(t, expect, buf)

	buffer := bytes.NewReader(expect)
	cmd, err := UnpackSqliCommand(buffer)
	assert.NoError(t, err)
	assert.Equal(t, sqliRetType, cmd)
	pos, err := buffer.Seek(0, io.SeekCurrent)
	assert.NoError(t, err)
	assert.EqualValues(t, pos, len(expect))

	sqliRetType.Columns[0].Type = 44
	sqliRetType.Columns[0].Length = 72
	sqliRetType.Columns[0].OwnerName = "gbasedbt"
	sqliRetType.Columns[0].ExtTypeName = "clob"
	expect = []byte{0, 100, 0, 1, 0, 1, 0, 44, 0, 8, 103, 98, 97, 115, 101, 100, 98, 116, 0, 4, 99, 108, 111, 98, 0, 0, 0, 72}
	buf, err = sqliRetType.Pack()
	assert.NoError(t, err)
	assert.Equal(t, expect, buf)

	buffer = bytes.NewReader(expect)
	cmd, err = UnpackSqliCommand(buffer)
	assert.NoError(t, err)
	assert.Equal(t, sqliRetType, cmd)
	pos, err = buffer.Seek(0, io.SeekCurrent)
	assert.NoError(t, err)
	assert.EqualValues(t, pos, len(expect))
}

func TestSqliErr_Pack_Unpack(t *testing.T) {
	sqliErr := &SqliErr{SQLCode: -206, RSAMError: -111, Offset: 0xc, SQLerrm: "x"}
	buf, err := sqliErr.Pack()
	expect := []byte{0, 13, 255, 50, 255, 145, 0, 0, 0, 12, 0, 1, 120, 0}
	assert.NoError(t, err)
	assert.Equal(t, expect, buf)

	buffer := bytes.NewReader(expect)
	cmd, err := UnpackSqliCommand(buffer)
	assert.NoError(t, err)
	assert.Equal(t, sqliErr, cmd)
	pos, err := buffer.Seek(0, io.SeekCurrent)
	assert.NoError(t, err)
	assert.EqualValues(t, pos, len(expect))
}

func TestSqliBind_Pack_Unpack(t *testing.T) {
	sqliBind := &SqliBind{}
	col := BindColumnInt{Type: 2, Indicator: 0, Precision: 2560, Data: 11}
	sqliBind.Columns = append(sqliBind.Columns, col)
	buf, err := sqliBind.Pack()
	expect := []byte{0, 5, 0, 1, 0, 2, 0, 0, 10, 0, 0, 0, 0, 11}
	assert.NoError(t, err)
	assert.Equal(t, expect, buf)

	buffer := bytes.NewReader(expect)
	cmd, err := UnpackSqliCommand(buffer)
	assert.NoError(t, err)
	assert.Equal(t, sqliBind, cmd)
	pos, err := buffer.Seek(0, io.SeekCurrent)
	assert.NoError(t, err)
	assert.EqualValues(t, pos, len(expect))

	expect = []byte{0, 5, 0, 1, 0, 0, 0, 0, 0, 0, 0, 9, 122, 104, 97, 110, 103, 115, 97, 110, 0x31, 0}
	sqliBind1 := &SqliBind{}
	col1 := BindColumnChar{Type: 0, Indicator: 0, Precision: 0, Data: "zhangsan1"}
	sqliBind1.Columns = append(sqliBind1.Columns, col1)
	//buf, err = sqliBind1.Pack()
	//assert.NoError(t, err)
	//assert.Equal(t, expect, buf)

	buffer = bytes.NewReader(expect)
	cmd, err = UnpackSqliCommand(buffer)
	assert.NoError(t, err)
	assert.Equal(t, sqliBind1, cmd)
	pos, err = buffer.Seek(0, io.SeekCurrent)
	assert.NoError(t, err)
	assert.EqualValues(t, pos, len(expect))
}

func TestSqliAutoFree_Pack_Unpack(t *testing.T) {
	sqliAutoFree := &SqliAutoFree{}
	buf, err := sqliAutoFree.Pack()
	expect := []byte{0, 108}
	assert.NoError(t, err)
	assert.Equal(t, expect, buf)

	buffer := bytes.NewReader(expect)
	cmd, err := UnpackSqliCommand(buffer)
	assert.NoError(t, err)
	assert.Equal(t, sqliAutoFree, cmd)
	pos, err := buffer.Seek(0, io.SeekCurrent)
	assert.NoError(t, err)
	assert.EqualValues(t, pos, len(expect))
}

func TestSqliCmmtwork_Pack_Unpack(t *testing.T) {
	Cmmtwork := &SqliCmmtwork{}
	buf, err := Cmmtwork.Pack()
	expect := []byte{0, 19}
	assert.NoError(t, err)
	assert.Equal(t, expect, buf)

	buffer := bytes.NewReader(expect)
	cmd, err := UnpackSqliCommand(buffer)
	assert.NoError(t, err)
	assert.Equal(t, Cmmtwork, cmd)
	pos, err := buffer.Seek(0, io.SeekCurrent)
	assert.NoError(t, err)
	assert.EqualValues(t, pos, len(expect))
}

func TestSqliBegin_Pack_Unpack(t *testing.T) {
	begin := &SqliBegin{}
	buf, err := begin.Pack()
	expect := []byte{0, 35}
	assert.NoError(t, err)
	assert.Equal(t, expect, buf)

	buffer := bytes.NewReader(expect)
	cmd, err := UnpackSqliCommand(buffer)
	assert.NoError(t, err)
	assert.Equal(t, begin, cmd)
	pos, err := buffer.Seek(0, io.SeekCurrent)
	assert.NoError(t, err)
	assert.EqualValues(t, pos, len(expect))
}

func TestSqliXActstat_Pack_Unpack(t *testing.T) {
	xactxtat := &SqliXActstat{}
	buf, err := xactxtat.Pack()
	expect := []byte{0, 99, 0, 0, 0, 0, 0, 0}
	assert.NoError(t, err)
	assert.Equal(t, expect, buf)

	buffer := bytes.NewReader(expect)
	cmd, err := UnpackSqliCommand(buffer)
	assert.NoError(t, err)
	assert.Equal(t, xactxtat, cmd)
	pos, err := buffer.Seek(0, io.SeekCurrent)
	assert.NoError(t, err)
	assert.EqualValues(t, pos, len(expect))
}

func TestSqliCIdescribe_Pack_Unpack(t *testing.T) {
	cidescribe := &SqliCIdescribe{}
	buf, err := cidescribe.Pack()
	expect := []byte{0, 124}
	assert.NoError(t, err)
	assert.Equal(t, expect, buf)

	buffer := bytes.NewReader(expect)
	cmd, err := UnpackSqliCommand(buffer)
	assert.NoError(t, err)
	assert.Equal(t, cidescribe, cmd)
	pos, err := buffer.Seek(0, io.SeekCurrent)
	assert.NoError(t, err)
	assert.EqualValues(t, pos, len(expect))
}

func TestSqliIdescribe_Pack_Unpack(t *testing.T) {
	idescribe := &SqliIdescribe{Fields: nil}
	idescribe.Fields = append(idescribe.Fields)
	buf, err := idescribe.Pack()
	expect := []byte{0, 125, 0, 0, 0, 0}
	assert.NoError(t, err)
	assert.Equal(t, expect, buf)

	buffer := bytes.NewReader(expect)
	cmd, err := UnpackSqliCommand(buffer)
	assert.NoError(t, err)
	assert.Equal(t, idescribe, cmd)
	pos, err := buffer.Seek(0, io.SeekCurrent)
	assert.NoError(t, err)
	assert.EqualValues(t, pos, len(expect)-2)
}
