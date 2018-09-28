package model

import (
	"bytes"
	"encoding/binary"
	"github.com/stretchr/testify/assert"
	"github.com/zhuangsirui/binpacker"
	"testing"
)

func TestSqTuple_Unpack(t *testing.T) {
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