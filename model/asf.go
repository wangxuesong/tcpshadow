package model

import (
	"bytes"
	"encoding/binary"
	"github.com/zhuangsirui/binpacker"
	"io"
)

type TupleValue interface {
	PackTupleValue(writer io.Writer) (error)
}

type SqTuple struct {
	Warnings uint16
	Value    TupleValue
}

type SmallIntTupleValue struct {
	length uint32
	Value  int16
}

func NewSmallIntTuple(warn uint16, value int16) SqTuple {
	tvalue := SmallIntTupleValue{length:2, Value:value}
	return SqTuple{Warnings:warn, Value:&tvalue}
}

func (sq *SqTuple) command() uint16 {
	return 14
}

func (sq *SqTuple) Pack() ([]byte, error) {
	buffer := new(bytes.Buffer)
	packer := binpacker.NewPacker(binary.BigEndian, buffer)
	packer.PushUint16(sq.command())
	packer.PushUint16(sq.Warnings)
	sq.Value.PackTupleValue(buffer)
	return buffer.Bytes(), nil
}

func (v *SmallIntTupleValue) PackTupleValue(writer io.Writer) error {
	packer := binpacker.NewPacker(binary.BigEndian, writer)
	packer.PushUint32(v.length)
	packer.PushInt16(v.Value)
	return nil
}
