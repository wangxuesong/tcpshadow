package model

import (
	"bytes"
	"encoding/binary"
	"github.com/stretchr/testify/assert"
	"github.com/zhuangsirui/binpacker"
	"testing"
)

func TestSqTuple_Pack(t *testing.T) {
	expect := []byte{0, 14, 0, 0, 0, 0, 0, 15, 0, 11, 0, 0, 0, 0, 8, 115, 119, 101, 101, 116, 104, 117, 105, 0}

	intvalue := &SmallIntTupleValue{Value:11}
	varcharvalue := &LVarcharTupleValue{Value:"sweethui"}
	sq := SqliTuple{
		Warnings:0,
	}
	sq.Values = append(sq.Values, intvalue)
	sq.Values = append(sq.Values, varcharvalue)
	actual, err := sq.Pack()
	assert.NoError(t, err)
	assert.NotNil(t, actual)
	assert.Equal(t, expect, actual)
}

func TestSqTuple_Unpack(t *testing.T) {
	buf := []byte{0, 0, 0, 0, 0, 2, 0, 11}
	buffer := bytes.NewBuffer(buf)
	unpacker := binpacker.NewUnpacker(binary.BigEndian, buffer)
	tuple := &SqliTuple{}
	smallint := &SmallIntTupleValue{}
	unpacker.FetchUint16(&tuple.Warnings).
		FetchUint32(&smallint.length).FetchInt16(&smallint.Value)
	assert.NoError(t, unpacker.Error())
}

func TestSqliDescribe_Pack(t *testing.T) {
	expect := []byte{0, 8, 0, 2, 0, 0, 0, 0, 0, 0, 0, 55, 0, 2, 0, 0, 0, 4, 0, 0, 0, 0, 0, 0, 0, 0, 0, 2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 4, 0, 0, 0, 2, 0, 0, 0, 4, 0, 13, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 50, 97, 0, 98, 0}
	desc := SqliDescribe{
		StatementType: 2,
		StatementID:   0,
		EstimatedCost: 0,
		TupleSize:     55,
		//CountOfFields: 2,
		//StringTable:   4,
	}
	field1 := SqliField{
		FieldIndex:     0,
		ColumnStartPos: 0,
		ColumnType:     2,
		Length:         4,
		Name:           "a",
	}
	desc.Fields = append(desc.Fields, field1)
	field2 := SqliField{
		FieldIndex:     2,
		ColumnStartPos: 4,
		ColumnType:     13,
		Length:         50,
		Name:           "b",
	}
	desc.Fields = append(desc.Fields, field2)

	buf, err := desc.Pack()
	assert.NoError(t, err)
	assert.Equal(t, expect, buf)
}

func TestSqliDone_Pack(t *testing.T) {
	done := SqliDone{}
	buf, err := done.Pack()
	expect := []byte{0, 15, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	assert.NoError(t, err)
	assert.Equal(t, expect, buf)
}

func TestSqliCost_Pack(t *testing.T) {
	cost := SqliCost{EstimatedRows: 1, EstimatedIO: 1}
	buf, err := cost.Pack()
	expect := []byte{0, 55, 0, 0, 0, 1, 0, 0, 0, 1}
	assert.NoError(t, err)
	assert.Equal(t, expect, buf)
}

func TestSqliTransmission_Pack(t *testing.T) {
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
}
