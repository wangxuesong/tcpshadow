package model

import (
	"bytes"
	"encoding/binary"
	"github.com/stretchr/testify/assert"
	"github.com/zhuangsirui/binpacker"
	"testing"
)

func TestSqTuple_Pack(t *testing.T) {
	expect := new(bytes.Buffer)
	packer := binpacker.NewPacker(binary.BigEndian, expect)
	packer.PushUint16(14)
	packer.PushUint16(0)
	packer.PushUint32(2)
	packer.PushUint16(11)

	sq := NewSmallIntTuple(0, 11)
	actual, err := sq.Pack()
	assert.NoError(t, err)
	assert.NotNil(t, actual)
	assert.Equal(t, expect.Bytes(), actual)
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

func TestSqliDone_Pack(t *testing.T) {
	done := SqliDone{}
	buf, err := done.Pack()
	expect := []byte{0, 15, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	assert.NoError(t, err)
	assert.Equal(t, expect, buf)
}

func TestSqliCost_Pack(t *testing.T) {
	cost := SqliCost{EstimatedRows:1, EstimatedIO:1}
	buf, err := cost.Pack()
	expect := []byte{0, 55, 0, 0, 0, 1, 0, 0, 0, 1}
	assert.NoError(t, err)
	assert.Equal(t, expect, buf)
}

func TestSqliTransmission_Pack(t *testing.T) {
	expect := []byte{0, 8, 0, 2, 0, 0, 0, 0, 0, 0, 0, 55, 0, 2, 0, 0, 0, 4, 0, 0, 0, 0, 0, 0, 0, 0, 0, 2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 4, 0, 0, 0, 2, 0, 0, 0, 4, 0, 13, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 50, 97, 0, 98, 0, 0, 15, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 55, 0, 0, 0, 1, 0, 0, 0, 1, 0, 12}
	size := len(expect)
	trans, err := CreateDescribeTransmission()
	assert.NoError(t, err)
	assert.Equal(t, 4, len(trans))
	assert.EqualValues(t,  8, trans[0].Command()) // SQ_DESCRIBE
	assert.EqualValues(t, 15, trans[1].Command()) // SQ_DONE
	assert.EqualValues(t, 55, trans[2].Command()) // SQ_COST
	assert.EqualValues(t, 12, trans[3].Command()) // SQ_EOT
	buf, err:=trans.Pack()
	assert.NoError(t, err)
	assert.Equal(t, expect,buf)
	buf, err = trans[3].Pack()
	assert.NoError(t, err)
	assert.Equal(t, expect[size -2:], buf)
}