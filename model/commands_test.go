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
	tupleCmd.Values = append(tupleCmd.Values, &SmallIntTupleValue{})
	tupleCmd.Values = append(tupleCmd.Values, &LVarcharTupleValue{})
	reader := bytes.NewReader(tupleCmd.tupleBytes)
	tupleCmd.unpackValues(reader)
	//assert.Equal(t, &tuple, cmd)
	assert.Equal(t, tuple.Warnings, tupleCmd.Warnings)
	assert.Equal(t, tuple.Values, tupleCmd.Values)
	pos, err := buffer.Seek(0, io.SeekCurrent)
	assert.NoError(t, err)
	assert.EqualValues(t, pos, len(expect))
}

func TestSqliDescribe_Pack_Unpack(t *testing.T) {
	expect := []byte{0, 8, 0, 2, 0, 0, 0, 0, 0, 0, 0, 55, 0, 2, 0, 0, 0, 4, 0, 0, 0, 0, 0, 0, 0, 0, 0, 2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 4, 0, 0, 0, 2, 0, 0, 0, 4, 0, 13, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 50, 97, 0, 98, 0}
	desc := &SqliDescribe{
		StatementType: 2,
		StatementID:   0,
		EstimatedCost: 0,
		TupleSize:     55,
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
	expect := []byte{0, 8, 0, 2, 0, 0, 0, 0, 0, 0, 0, 55, 0, 2, 0, 0, 0, 4, 0, 0, 0, 0, 0, 0, 0, 0, 0, 2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 4, 0, 0, 0, 2, 0, 0, 0, 4, 0, 13, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 50, 97, 0, 98, 0, 0, 15, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 55, 0, 0, 0, 1, 0, 0, 0, 1, 0, 12}
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

func TestSqliInfo_Pack(t *testing.T) {
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
