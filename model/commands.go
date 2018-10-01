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

type SqliCommand interface {
	Command() uint16
	Pack() ([]byte, error)
}

type SqliTransmission []SqliCommand

func (t *SqliTransmission) Pack() ([]byte, error) {
	source := []byte{0, 8, 0, 2, 0, 0, 0, 0, 0, 0, 0, 55, 0, 2, 0, 0, 0, 4, 0, 0, 0, 0, 0, 0, 0, 0, 0, 2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 4, 0, 0, 0, 2, 0, 0, 0, 4, 0, 13, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 50, 97, 0, 98, 0, 0, 15, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 55, 0, 0, 0, 1, 0, 0, 0, 1, 0, 12}
	temp := new(bytes.Buffer)
	for i, cmd := range *t {
		if i == 0 {
			continue
		}
		buf, err := cmd.Pack()
		if err != nil {
			return nil, err
		}
		temp.Write(buf)
	}
	buffer := new(bytes.Buffer)
	buffer.Write(source[:len(source)-temp.Len()])
	buffer.Write(temp.Bytes())
	return buffer.Bytes(), nil
}

//SqliTuple SQ_TUPLE 14
type SqliTuple struct {
	Warnings uint16
	Value    TupleValue
}

func (sq *SqliTuple) Command() uint16 {
	return 14
}

func (sq *SqliTuple) Pack() ([]byte, error) {
	buffer := new(bytes.Buffer)
	packer := binpacker.NewPacker(binary.BigEndian, buffer)
	packer.PushUint16(sq.Command())
	packer.PushUint16(sq.Warnings)
	sq.Value.PackTupleValue(buffer)
	return buffer.Bytes(), nil
}

func CreateDescribeTransmission() (SqliTransmission, error) {
	trans := make([]SqliCommand, 0, 4)
	trans = append(trans, &SqliDescribe{})
	trans = append(trans, &SqliDone{})
	trans = append(trans, &SqliCost{EstimatedRows:1, EstimatedIO:1})
	trans = append(trans, &SqliEot{})
	return trans, nil
}

type SqliDescribe struct {
}

func (*SqliDescribe) Command() uint16 {
	return 8
}

func (*SqliDescribe) Pack() ([]byte, error) {
	panic("implement me")
}

//SqliDone SQ_DONE 15
type SqliDone struct {
	Warning  int16
	Rows     uint32
	RowID    uint32
	SerialID uint32
}

func (*SqliDone) Command() uint16 {
	return 15
}

func (sq *SqliDone) Pack() ([]byte, error) {
	buffer := new(bytes.Buffer)
	packer := binpacker.NewPacker(binary.BigEndian, buffer)
	packer.PushUint16(sq.Command())
	packer.PushInt16(sq.Warning).
		PushUint32(sq.Rows).
		PushUint32(sq.RowID).
		PushUint32(sq.SerialID)
	return buffer.Bytes(), packer.Error()
}

//SqliCost SQ_COST 55
type SqliCost struct {
	EstimatedRows uint32
	EstimatedIO   uint32
}

func (sq *SqliCost) Command() uint16 {
	return 55
}

func (sq *SqliCost) Pack() ([]byte, error) {
	buffer := new(bytes.Buffer)
	packer := binpacker.NewPacker(binary.BigEndian, buffer)
	packer.PushUint16(sq.Command())
	packer.PushUint32(sq.EstimatedRows).PushUint32(sq.EstimatedIO)
	return buffer.Bytes(), packer.Error()
}

//SqliEot SQ_EOT 12
type SqliEot struct {
}

func (*SqliEot) Command() uint16 {
	return 12
}

func (*SqliEot) Pack() ([]byte, error) {
	return []byte{0, 12}, nil
}

type SmallIntTupleValue struct {
	length uint32
	Value  int16
}

func NewSmallIntTuple(warn uint16, value int16) SqliTuple {
	tvalue := SmallIntTupleValue{length: 2, Value: value}
	return SqliTuple{Warnings: warn, Value: &tvalue}
}

func (v *SmallIntTupleValue) PackTupleValue(writer io.Writer) error {
	packer := binpacker.NewPacker(binary.BigEndian, writer)
	packer.PushUint32(v.length)
	packer.PushInt16(v.Value)
	return packer.Error()
}
