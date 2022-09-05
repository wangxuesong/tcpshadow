package model

import (
	"encoding/binary"
	"github.com/zhuangsirui/binpacker"
	"io"
	"strings"
)

type TupleValues []TupleValue

type CharTupleValue struct {
	Length uint32
	Value  string
}

func (v *CharTupleValue) PackTupleValue(writer io.Writer) error {
	buffer := []byte(v.Value + strings.Repeat(" ", int(v.Length)-len(v.Value)))
	_, err := writer.Write(buffer)
	return err
}

func (v *CharTupleValue) UnpackTupleValue(reader io.Reader) error {
	unpacker := binpacker.NewUnpacker(binary.BigEndian, reader)
	unpacker.FetchString(uint64(v.Length), &v.Value)
	return unpacker.Error()
}

func (v *CharTupleValue) Size() int64 {
	return int64(v.Length)
}

type SmallIntTupleValue struct {
	Value int16
}

func (v *SmallIntTupleValue) PackTupleValue(writer io.Writer) error {
	packer := binpacker.NewPacker(binary.BigEndian, writer)
	packer.PushInt16(v.Value)
	return packer.Error()
}

func (v *SmallIntTupleValue) UnpackTupleValue(reader io.Reader) error {
	unpacker := binpacker.NewUnpacker(binary.BigEndian, reader)
	unpacker.FetchInt16(&v.Value)
	return unpacker.Error()
}

func (v *SmallIntTupleValue) Size() int64 {
	return 2
}

type IntTupleValue struct {
	Value int32
}

func (v *IntTupleValue) PackTupleValue(writer io.Writer) error {
	packer := binpacker.NewPacker(binary.BigEndian, writer)
	packer.PushInt32(v.Value)
	return packer.Error()
}

func (v *IntTupleValue) UnpackTupleValue(reader io.Reader) error {
	unpacker := binpacker.NewUnpacker(binary.BigEndian, reader)
	unpacker.FetchInt32(&v.Value)
	return unpacker.Error()
}

func (v *IntTupleValue) Size() int64 {
	return 4
}

type VarcharTupleValue struct {
	Value string
}

func (v *VarcharTupleValue) PackTupleValue(writer io.Writer) error {
	packer := binpacker.NewPacker(binary.BigEndian, writer)
	packer.PushUint16(uint16(len(v.Value)))
	packer.PushBytes([]byte(v.Value))
	return packer.Error()
}

func (v *VarcharTupleValue) UnpackTupleValue(reader io.Reader) error {
	unpacker := binpacker.NewUnpacker(binary.BigEndian, reader)
	size, err := unpacker.ShiftUint16()
	if err != nil {
		return err
	}
	unpacker.FetchString(uint64(size), &v.Value)
	return unpacker.Error()
}

func (v *VarcharTupleValue) Size() int64 {
	return int64(len(v.Value)) + 1
}

type LVarcharTupleValue struct {
	Value string
}

func (v *LVarcharTupleValue) PackTupleValue(writer io.Writer) error {
	packer := binpacker.NewPacker(binary.BigEndian, writer)
	packer.PushByte(0) // 长度最高字节,暂时只支持4字节，协议支持5字节
	packer.PushUint32(uint32(len(v.Value)))
	packer.PushBytes([]byte(v.Value))
	return packer.Error()
}

func (v *LVarcharTupleValue) UnpackTupleValue(reader io.Reader) error {
	unpacker := binpacker.NewUnpacker(binary.BigEndian, reader)
	_, err := unpacker.ShiftByte()
	if err != nil {
		return err
	}
	size, err := unpacker.ShiftUint32()
	if err != nil {
		return err
	}
	unpacker.FetchString(uint64(size), &v.Value)
	return unpacker.Error()
}

func (v *LVarcharTupleValue) Size() int64 {
	return int64(len(v.Value)) + 5
}
