package model

import (
	"bytes"
	"encoding/binary"
	"github.com/zhuangsirui/binpacker"
	"io"
)

type TupleValue interface {
	PackTupleValue(writer io.Writer) error
	Size() int64
}

type TupleValues []TupleValue

func (tv *TupleValues) PackTupleValue(writer io.Writer) error {
	for _, v := range *tv {
		err := v.PackTupleValue(writer)
		if err != nil {
			return err
		}
	}
	return nil
}

type SqliCommand interface {
	Command() uint16
	Pack() ([]byte, error)
}

type SqliTransmission []SqliCommand

func (t *SqliTransmission) Pack() ([]byte, error) {
	temp := new(bytes.Buffer)
	for _, cmd := range *t {
		buf, err := cmd.Pack()
		if err != nil {
			return nil, err
		}
		temp.Write(buf)
	}
	buffer := new(bytes.Buffer)
	buffer.Write(temp.Bytes())
	return buffer.Bytes(), nil
}
func (t *SqliTransmission) Append(cmd SqliCommand) {
	*t = append(*t, cmd)
}

//SqliPrepare SQ_PREPARE 2
type SqliPrepare struct {
	QMarks uint16
	Sql    string
}

func (*SqliPrepare) Command() uint16 {
	return 2
}

func (sq *SqliPrepare) Pack() ([]byte, error) {
	buffer := new(bytes.Buffer)
	packer := binpacker.NewPacker(binary.BigEndian, buffer)
	packer.PushUint16(sq.Command())
	packer.PushUint16(sq.QMarks)
	packer.PushUint32(uint32(len(sq.Sql)))
	packer.PushBytes([]byte(sq.Sql))
	if len(sq.Sql)%2 == 1 {
		packer.PushByte(0)
	}

	return buffer.Bytes(), packer.Error()
}
func (sq *SqliPrepare) Unpack(r io.Reader) error {
	unpacker := binpacker.NewUnpacker(binary.BigEndian, r)
	var size uint32
	unpacker.FetchUint16(&sq.QMarks).
		FetchUint32(&size).
		FetchString(uint64(size), &sq.Sql)
	if size %2 == 1 {
		var tmp byte
		unpacker.FetchByte(&tmp)
	}
	return unpacker.Error()
}

//SqliTuple SQ_TUPLE 14
type SqliTuple struct {
	Warnings uint16
	Values   TupleValues
}

func (sq *SqliTuple) Command() uint16 {
	return 14
}

func (sq *SqliTuple) Pack() ([]byte, error) {
	buffer := new(bytes.Buffer)
	packer := binpacker.NewPacker(binary.BigEndian, buffer)
	packer.PushUint16(sq.Command())
	packer.PushUint16(sq.Warnings)
	var sum int64 = 0
	for _, v := range sq.Values {
		sum += v.Size()
	}
	packer.PushUint32(uint32(sum))
	valuesBuf := new(bytes.Buffer)
	sq.Values.PackTupleValue(valuesBuf)
	valuesBuf.WriteTo(buffer)
	if sum%2 == 1 {
		packer.PushByte(0) // Pad
	}
	return buffer.Bytes(), nil
}

func NewDescribeTransmission() (SqliTransmission, error) {
	trans := SqliTransmission{}
	desc := SqliDescribe{
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
	trans.Append(&desc)
	trans.Append(&SqliDone{})
	trans.Append(&SqliCost{EstimatedRows: 1, EstimatedIO: 1})
	trans.Append(&SqliEot{})
	return trans, nil
}

//SqliDescribe SQ_DESCRIBE 8
type SqliDescribe struct {
	StatementType uint16
	StatementID   uint16
	EstimatedCost uint32
	TupleSize     uint16
	CountOfFields uint16
	StringTable   uint32
	Fields        []SqliField
}

func (*SqliDescribe) Command() uint16 {
	return 8
}

func (sq *SqliDescribe) Pack() ([]byte, error) {
	fieldsBuf := new(bytes.Buffer)
	err := sq.packFields(fieldsBuf)
	if err != nil {
		return nil, err
	}
	stringTableBuf := new(bytes.Buffer)
	strTable, err := sq.packStringTable(stringTableBuf)
	if err != nil {
		return nil, err
	}
	sq.CountOfFields = uint16(len(sq.Fields))
	sq.StringTable = strTable
	buffer := new(bytes.Buffer)
	packer := binpacker.NewPacker(binary.BigEndian, buffer)
	packer.PushUint16(sq.Command()).
		PushUint16(sq.StatementType).
		PushUint16(sq.StatementID).
		PushUint32(sq.EstimatedCost).
		PushUint16(sq.TupleSize).
		PushUint16(sq.CountOfFields).
		PushUint32(sq.StringTable)

	fieldsBuf.WriteTo(buffer)
	stringTableBuf.WriteTo(buffer)

	return buffer.Bytes(), packer.Error()
}

func (sq *SqliDescribe) packStringTable(writer io.Writer) (uint32, error) {
	packer := binpacker.NewPacker(binary.BigEndian, writer)
	count := 0
	for _, f := range sq.Fields {
		packer.PushString(f.Name)
		packer.PushByte(0)
		count += len(f.Name) + 1
	}
	if count%2 != 0 {
		packer.PushByte(0)
	}

	return uint32(count), packer.Error()
}

func (sq *SqliDescribe) packFields(writer io.Writer) error {
	packer := binpacker.NewPacker(binary.BigEndian, writer)
	for _, f := range sq.Fields {
		packer.PushUint32(f.FieldIndex).
			PushUint32(f.ColumnStartPos).
			PushUint16(f.ColumnType).
			PushUint32(f.ColumnExtendedBuiltinId).
			PushUint16(f.OwnerName).
			PushUint16(f.ExtendedName).
			PushUint16(f.Reference).
			PushUint16(f.Alignment).
			PushUint32(f.SourceType).
			PushUint32(f.Length)
	}

	return packer.Error()
}
func (sq *SqliDescribe) AppendFields(field SqliField) {
	sq.Fields = append(sq.Fields, field)
}

type SqliField struct {
	FieldIndex              uint32
	ColumnStartPos          uint32
	ColumnType              uint16
	ColumnExtendedBuiltinId uint32
	OwnerName               uint16
	ExtendedName            uint16
	Reference               uint16
	Alignment               uint16
	SourceType              uint32
	Length                  uint32
	Name                    string
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

func (v *SmallIntTupleValue) Size() int64 {
	return 2
}

func NewSmallIntTuple(warn uint16, value int16) SqliTuple {
	tvalue := SmallIntTupleValue{length: 2, Value: value}
	return SqliTuple{Warnings: warn, Values: []TupleValue{&tvalue}}
}

func (v *SmallIntTupleValue) PackTupleValue(writer io.Writer) error {
	packer := binpacker.NewPacker(binary.BigEndian, writer)
	//packer.PushUint32(v.length)
	packer.PushInt16(v.Value)
	return packer.Error()
}

type LVarcharTupleValue struct {
	Value string
}

func (v *LVarcharTupleValue) Size() int64 {
	return int64(len(v.Value)) + 5
}

func (v *LVarcharTupleValue) PackTupleValue(writer io.Writer) error {
	packer := binpacker.NewPacker(binary.BigEndian, writer)
	packer.PushByte(0) // 长度最高字节,暂时只支持4字节，协议支持5字节
	packer.PushUint32(uint32(len(v.Value)))
	packer.PushBytes([]byte(v.Value))
	return packer.Error()
}

func UnpackSqliCommand(reader io.ReadSeeker) (SqliCommand, error) {
	var cmd uint16
	pos, err := reader.Seek(0, io.SeekCurrent)
	if err != nil {
		return nil, err
	}
	err = binary.Read(reader, binary.BigEndian, &cmd)
	if err != nil {
		reader.Seek(pos, io.SeekStart)
		return nil, err
	}

	switch cmd {
	case 2:
		command := &SqliPrepare{}
		err = command.Unpack(reader)
		if err != nil {
			reader.Seek(pos, io.SeekStart)
			return nil, err
		}
		return command, nil
	default:
		panic(cmd)
	}
}