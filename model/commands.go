package model

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/zhuangsirui/binpacker"
	"io"
	"log"
	"strings"
)

type SqliType uint16

const (
	SQ_COMMAND    SqliType = 1
	SQ_PREPARE             = 2
	SQ_CURNAME             = 3
	SQ_ID                  = 4
	SQ_BIND                = 5
	SQ_OPEN                = 6
	SQ_EXECUTE             = 7
	SQ_DESCRIBE            = 8
	SQ_NFETCH              = 9
	SQ_CLOSE               = 10
	SQ_RELEASE             = 11
	SQ_EOT                 = 12
	SQ_ERR                 = 13
	SQ_TUPLE               = 14
	SQ_DONE                = 15
	SQ_CMMTWORK            = 19
	SQ_NDESCRIBE           = 22
	SQ_BEGIN               = 35
	SQ_DBOPEN              = 36
	SQ_WANTDONE            = 49
	SQ_COST                = 55
	SQ_EXIT                = 56
	SQ_INFO                = 81
	SQ_INSERTDONE          = 94
	SQ_XACTSTAT            = 99
	SQ_RETTYPE             = 100
	SQ_AUTOFREE            = 108
	SQ_CIDESCRIBE          = 124
	SQ_IDESCRIBE           = 125
	SQ_PROTOCOLS           = 126
)

var (
	SqliTypeStringMap = map[SqliType]string{
		SQ_COMMAND:    "SQ_COMMAND",
		SQ_PREPARE:    "SQ_PREPARE",
		SQ_CURNAME:    "SQ_CURNAME",
		SQ_ID:         "SQ_ID",
		SQ_BIND:       "SQ_BIND",
		SQ_OPEN:       "SQ_OPEN",
		SQ_EXECUTE:    "SQ_EXECUTE",
		SQ_DESCRIBE:   "SQ_DESCRIBE",
		SQ_NFETCH:     "SQ_NFETCH",
		SQ_CLOSE:      "SQ_CLOSE",
		SQ_RELEASE:    "SQ_RELEASE",
		SQ_EOT:        "SQ_EOT",
		SQ_ERR:        "SQ_ERR",
		SQ_TUPLE:      "SQ_TUPLE",
		SQ_DONE:       "SQ_DONE",
		SQ_NDESCRIBE:  "SQ_NDESCRIBE",
		SQ_DBOPEN:     "SQ_DBOPEN",
		SQ_WANTDONE:   "SQ_WANTDONE",
		SQ_COST:       "SQ_COST",
		SQ_EXIT:       "SQ_EXIT",
		SQ_INFO:       "SQ_INFO",
		SQ_INSERTDONE: "SQ_INSERTDONE",
		SQ_RETTYPE:    "SQ_RETTYPE",
		SQ_AUTOFREE:   "SQ_AUTOFREE",
		SQ_PROTOCOLS:  "SQ_PROTOCOLS",
		SQ_CMMTWORK:   "SQ_CMMTWORK",
		SQ_BEGIN:      "SQ_BEGIN",
		SQ_XACTSTAT:   "SQ_XACTSTAT",
		SQ_CIDESCRIBE: "SQ_CIDESCRIBE",
		SQ_IDESCRIBE:  "SQ_IDESCRIBE",
	}

	StringSqliTypeMap = map[string]SqliType{
		"SQ_COMMAND":    SQ_COMMAND,
		"SQ_PREPARE":    SQ_PREPARE,
		"SQ_CURNAME":    SQ_CURNAME,
		"SQ_ID":         SQ_ID,
		"SQ_BIND":       SQ_BIND,
		"SQ_OPEN":       SQ_OPEN,
		"SQ_EXECUTE":    SQ_EXECUTE,
		"SQ_DESCRIBE":   SQ_DESCRIBE,
		"SQ_NFETCH":     SQ_NFETCH,
		"SQ_CLOSE":      SQ_CLOSE,
		"SQ_RELEASE":    SQ_RELEASE,
		"SQ_EOT":        SQ_EOT,
		"SQ_ERR":        SQ_ERR,
		"SQ_TUPLE":      SQ_TUPLE,
		"SQ_DONE":       SQ_DONE,
		"SQ_NDESCRIBE":  SQ_NDESCRIBE,
		"SQ_DBOPEN":     SQ_DBOPEN,
		"SQ_WANTDONE":   SQ_WANTDONE,
		"SQ_COST":       SQ_COST,
		"SQ_EXIT":       SQ_EXIT,
		"SQ_INFO":       SQ_INFO,
		"SQ_INSERTDONE": SQ_INSERTDONE,
		"SQ_RETTYPE":    SQ_RETTYPE,
		"SQ_AUTOFREE":   SQ_AUTOFREE,
		"SQ_PROTOCOLS":  SQ_PROTOCOLS,
		"SQ_CMMTWORK":   SQ_CMMTWORK,
		"SQ_BEGIN":      SQ_BEGIN,
		"SQ_XACTSTAT":   SQ_XACTSTAT,
		"SQ_CIDESCRIBE": SQ_CIDESCRIBE,
		"SQ_IDESCRIBE":  SQ_IDESCRIBE,
	}
)

func (t SqliType) String() string {
	str, _ := SqliTypeStringMap[t]
	return str
}

func ParseSqliType(str string) SqliType {
	t, _ := StringSqliTypeMap[str]
	return t
}

type TupleValue interface {
	PackTupleValue(writer io.Writer) error
	UnpackTupleValue(reader io.Reader) error
	Size() int64
}

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
	Unpack(r io.Reader) error
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

// SqliCmd SQ_COMMAND 1
type SqliCmd struct {
	Sql string
}

func (*SqliCmd) Command() uint16 {
	return uint16(SQ_COMMAND)
}

func (sq *SqliCmd) Pack() ([]byte, error) {
	buffer := new(bytes.Buffer)
	packer := binpacker.NewPacker(binary.BigEndian, buffer)
	packer.PushUint16(sq.Command()).PushUint16(0)
	packer.PushUint32(uint32(len(sq.Sql)))
	packer.PushBytes([]byte(sq.Sql))
	if len(sq.Sql)%2 == 1 {
		packer.PushByte(0)
	}

	return buffer.Bytes(), packer.Error()
}

func (sq *SqliCmd) Unpack(r io.Reader) error {
	unpacker := binpacker.NewUnpacker(binary.BigEndian, r)
	var size uint32
	var foo uint16
	unpacker.FetchUint16(&foo).FetchUint32(&size).
		FetchString(uint64(size), &sq.Sql)
	if size%2 == 1 {
		var tmp byte
		unpacker.FetchByte(&tmp)
	}
	return unpacker.Error()
}

// SqliPrepare SQ_PREPARE 2
type SqliPrepare struct {
	QMarks uint16 `yaml:"QMarks"`
	Sql    string `yaml:"Sql"`
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
	if size%2 == 1 {
		var tmp byte
		unpacker.FetchByte(&tmp)
	}
	return unpacker.Error()
}

// SqliCurName SQ_CURNAME 3
type SqliCurName struct {
	CurName string
}

func (*SqliCurName) Command() uint16 {
	return 3
}

func (sq *SqliCurName) Pack() ([]byte, error) {
	buffer := new(bytes.Buffer)
	packer := binpacker.NewPacker(binary.BigEndian, buffer)
	packer.PushUint16(sq.Command())
	packer.PushUint16(uint16(len(sq.CurName)))
	packer.PushBytes([]byte(sq.CurName))
	if len(sq.CurName)%2 == 1 {
		packer.PushByte(0)
	}

	return buffer.Bytes(), packer.Error()
}

func (sq *SqliCurName) Unpack(r io.Reader) error {
	unpacker := binpacker.NewUnpacker(binary.BigEndian, r)
	var size uint16
	unpacker.FetchUint16(&size).
		FetchString(uint64(size), &sq.CurName)
	if size%2 == 1 {
		var tmp byte
		unpacker.FetchByte(&tmp)
	}
	return unpacker.Error()
}

// SqliID SQ_ID 4
type SqliID struct {
	ID int16
}

func (sq *SqliID) Command() uint16 {
	return 4
}

func (sq *SqliID) Pack() ([]byte, error) {
	buffer := new(bytes.Buffer)
	packer := binpacker.NewPacker(binary.BigEndian, buffer)
	packer.PushUint16(sq.Command())
	packer.PushInt16(sq.ID)

	return buffer.Bytes(), packer.Error()
}

func (sq *SqliID) Unpack(r io.Reader) error {
	unpacker := binpacker.NewUnpacker(binary.BigEndian, r)
	unpacker.FetchInt16(&sq.ID)
	return unpacker.Error()
}

// SqliBind SQ_BIND 5
type SqliBind struct {
	Columns []BindColumn
}

func (sq *SqliBind) Command() uint16 {
	return 5
}

func (sq *SqliBind) Pack() ([]byte, error) {
	buffer := new(bytes.Buffer)
	packer := binpacker.NewPacker(binary.BigEndian, buffer)
	packer.PushUint16(sq.Command())
	size := int16(len(sq.Columns))
	packer.PushInt16(size)

	for _, cc := range sq.Columns {
		switch cc.ColumnType() {
		case 2:
			var c BindColumnInt
			c = cc.(BindColumnInt)
			var precisionLow, precisionHigh uint16
			precisionHigh = uint16(c.Precision >> 16)
			precisionLow = uint16(c.Precision)
			packer.PushInt16(c.Type).
				PushInt16(c.Indicator).
				PushUint16(precisionLow).PushUint16(precisionHigh).
				PushUint16(c.Data)
		}
	}

	return buffer.Bytes(), packer.Error()
}

func (sq *SqliBind) Unpack(r io.Reader) error {
	unpacker := binpacker.NewUnpacker(binary.BigEndian, r)
	var size int16
	unpacker.FetchInt16(&size)
	sq.Columns = make([]BindColumn, 0, size)
	for i := 0; i < int(size); i++ {
		var colType int16
		unpacker.FetchInt16(&colType)
		switch colType {
		case 0:
			col := BindColumnChar{Type: colType}
			var precsionLow uint16
			unpacker.FetchInt16(&col.Indicator).
				FetchUint16(&precsionLow) //.FetchUint16(&precsionHigh)
			var count uint16
			unpacker.FetchUint16(&count).
				FetchString(uint64(count), &col.Data)
			if count%2 == 1 {
				var pad byte
				unpacker.FetchByte(&pad)
			}
			col.Precision = uint32(precsionLow)
			sq.Columns = append(sq.Columns, col)
		case 2:
			col := BindColumnInt{Type: colType}
			var precsionLow, precsionHigh uint16
			unpacker.FetchInt16(&col.Indicator).
				FetchUint16(&precsionLow).FetchUint16(&precsionHigh).
				FetchUint16(&col.Data)
			col.Precision = uint32(precsionHigh)<<16 + uint32(precsionLow)
			sq.Columns = append(sq.Columns, col)
		}
	}
	return unpacker.Error()
}

type BindColumn interface {
	ColumnType() int16
}

type BindColumnChar struct {
	Type      int16
	Indicator int16
	Precision uint32
	Data      string
}

func (c BindColumnChar) ColumnType() int16 {
	return 0
}

type BindColumnInt struct {
	Type      int16
	Indicator int16
	Precision uint32
	Data      uint16
}

func (c BindColumnInt) ColumnType() int16 {
	return 2
}

// SqliOpen SQ_OPEN 6
type SqliOpen struct {
}

func (*SqliOpen) Command() uint16 {
	return 6
}

func (*SqliOpen) Pack() ([]byte, error) {
	return []byte{0, 6}, nil
}

func (*SqliOpen) Unpack(r io.Reader) error {
	panic("implement me")
}

// SqliExecute SQ_EXECUTE 7
type SqliExecute struct {
}

func (*SqliExecute) Command() uint16 {
	return 7
}

func (*SqliExecute) Pack() ([]byte, error) {
	return []byte{0, 7}, nil
}

func (*SqliExecute) Unpack(r io.Reader) error {
	panic("implement me")
}

// SqliDescribe SQ_DESCRIBE 8
type SqliDescribe struct {
	StatementType uint16
	StatementID   uint16
	EstimatedCost uint32
	TupleSize     uint32
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
		PushUint32(sq.TupleSize).
		PushUint16(sq.CountOfFields).
		PushUint32(sq.StringTable)

	fieldsBuf.WriteTo(buffer)
	stringTableBuf.WriteTo(buffer)

	return buffer.Bytes(), packer.Error()
}

func (sq *SqliDescribe) Unpack(r io.Reader) error {
	unpacker := binpacker.NewUnpacker(binary.BigEndian, r)
	unpacker.FetchUint16(&sq.StatementType).
		FetchUint16(&sq.StatementID).
		FetchUint32(&sq.EstimatedCost).
		FetchUint32(&sq.TupleSize).
		FetchUint16(&sq.CountOfFields).
		FetchUint32(&sq.StringTable)
	err := unpacker.Error()
	if err != nil {
		return err
	}
	err = sq.unpackFields(r)
	if err != nil {
		return err
	}
	err = sq.unpackStringTable(r)
	if err != nil {
		return err
	}
	return unpacker.Error()
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

func (sq *SqliDescribe) unpackStringTable(r io.Reader) error {
	unpacker := binpacker.NewUnpacker(binary.BigEndian, r)
	var temp string
	unpacker.FetchString(uint64(sq.StringTable), &temp)
	if sq.StringTable%2 == 1 {
		var b byte
		unpacker.FetchByte(&b)
	}
	names := strings.Split(temp, string("\000"))
	if len(names)-1 != int(sq.CountOfFields) {
		return errors.New("unpack string table error")
	}

	for i := range sq.Fields {
		sq.Fields[i].Name = names[i]
	}
	return nil
}

func (sq *SqliDescribe) packFields(writer io.Writer) error {
	packer := binpacker.NewPacker(binary.BigEndian, writer)
	for _, f := range sq.Fields {
		packer.PushUint32(f.FieldIndex).
			PushUint32(f.ColumnStartPos).
			PushUint16(f.ColumnType).
			PushUint32(f.ColumnExtendedBuiltinId)
		if len(f.OwnerName) > 0 {
			packer.PushUint16(uint16(len(f.OwnerName)))
			packer.PushBytes([]byte(f.OwnerName))
			if len(f.OwnerName)%2 == 1 {
				packer.PushByte(0)
			}
		} else {
			packer.PushUint16(0)
		}
		if len(f.ExtendedName) > 0 {
			packer.PushUint16(uint16(len(f.ExtendedName)))
			packer.PushBytes([]byte(f.ExtendedName))
			if len(f.ExtendedName)%2 == 1 {
				packer.PushByte(0)
			}
		} else {
			packer.PushUint16(0)
		}
		packer.PushUint16(f.Reference).
			PushUint16(f.Alignment).
			PushUint32(f.SourceType).
			PushUint32(f.Length)
	}

	return packer.Error()
}

func (sq *SqliDescribe) unpackFields(r io.Reader) error {
	unpacker := binpacker.NewUnpacker(binary.BigEndian, r)

	for i := uint16(0); i < sq.CountOfFields; i++ {
		f := &SqliField{}
		unpacker.FetchUint32(&f.FieldIndex).
			FetchUint32(&f.ColumnStartPos).
			FetchUint16(&f.ColumnType).
			FetchUint32(&f.ColumnExtendedBuiltinId)
		var length uint16
		unpacker.FetchUint16(&length)
		if length > 0 {
			unpacker.FetchString(uint64(length), &f.OwnerName)
			if length%2 == 1 {
				var tmp byte
				unpacker.FetchByte(&tmp)
			}
		}
		unpacker.FetchUint16(&length)
		if length > 0 {
			unpacker.FetchString(uint64(length), &f.ExtendedName)
			if length%2 == 1 {
				var tmp byte
				unpacker.FetchByte(&tmp)
			}
		}
		unpacker.FetchUint16(&f.Reference).
			FetchUint16(&f.Alignment).
			FetchUint32(&f.SourceType).
			FetchUint32(&f.Length)

		sq.AppendFields(*f)
	}

	return unpacker.Error()
}

func (sq *SqliDescribe) AppendFields(field SqliField) {
	sq.Fields = append(sq.Fields, field)
}

type SqliField struct {
	FieldIndex              uint32
	ColumnStartPos          uint32
	ColumnType              uint16
	ColumnExtendedBuiltinId uint32
	OwnerName               string
	ExtendedName            string
	Reference               uint16
	Alignment               uint16
	SourceType              uint32
	Length                  uint32
	Name                    string
}

// SqliNFetch SQ_NFETCH 9
type SqliNFetch struct {
	TupleBufferSize uint32
	FetchArraySize  uint16
}

func (*SqliNFetch) Command() uint16 {
	return 9
}

func (sq *SqliNFetch) Pack() ([]byte, error) {
	buffer := new(bytes.Buffer)
	packer := binpacker.NewPacker(binary.BigEndian, buffer)
	packer.PushUint16(sq.Command()).
		PushUint32(sq.TupleBufferSize).
		PushUint16(sq.FetchArraySize)

	return buffer.Bytes(), packer.Error()
}

func (sq *SqliNFetch) Unpack(r io.Reader) error {
	unpacker := binpacker.NewUnpacker(binary.BigEndian, r)
	unpacker.FetchUint32(&sq.TupleBufferSize).
		FetchUint16(&sq.FetchArraySize)
	return unpacker.Error()
}

// SqliClose SQ_CLOSE 10
type SqliClose struct {
}

func (*SqliClose) Command() uint16 {
	return 10
}

func (*SqliClose) Pack() ([]byte, error) {
	return []byte{0, 10}, nil
}

func (*SqliClose) Unpack(r io.Reader) error {
	panic("implement me")
}

// SqliRelease SQ_RELEASE 11
type SqliRelease struct {
}

func (*SqliRelease) Command() uint16 {
	return 11
}

func (*SqliRelease) Pack() ([]byte, error) {
	return []byte{0, 11}, nil
}

func (*SqliRelease) Unpack(r io.Reader) error {
	panic("implement me")
}

// SqliEot SQ_EOT 12
type SqliEot struct {
}

func (*SqliEot) Command() uint16 {
	return 12
}

func (*SqliEot) Pack() ([]byte, error) {
	return []byte{0, 12}, nil
}

func (*SqliEot) Unpack(r io.Reader) error {
	panic("implement me")
}

// SqliErr SQ_ERR 13
type SqliErr struct {
	SQLCode   int16
	RSAMError int16
	Offset    uint32
	SQLerrm   string
}

func (*SqliErr) Command() uint16 {
	return 13
}

func (sq *SqliErr) Pack() ([]byte, error) {
	buffer := new(bytes.Buffer)
	packer := binpacker.NewPacker(binary.BigEndian, buffer)
	packer.PushUint16(sq.Command())
	size := len(sq.SQLerrm)
	packer.PushInt16(sq.SQLCode).
		PushInt16(sq.RSAMError).
		PushUint32(sq.Offset).
		PushUint16(uint16(size)).
		PushString(sq.SQLerrm)
	if size%2 == 1 {
		packer.PushByte(0)
	}
	return buffer.Bytes(), packer.Error()
}

func (sq *SqliErr) Unpack(r io.Reader) error {
	var size uint16
	unpacker := binpacker.NewUnpacker(binary.BigEndian, r)
	unpacker.FetchInt16(&sq.SQLCode).
		FetchInt16(&sq.RSAMError).
		FetchUint32(&sq.Offset).
		FetchUint16(&size).
		FetchString(uint64(size), &sq.SQLerrm)
	if size%2 == 1 {
		var b byte
		unpacker.FetchByte(&b)
	}
	return unpacker.Error()
}

// SqliTuple SQ_TUPLE 14
type SqliTuple struct {
	Warnings   uint16
	size       uint32
	tupleBytes []byte
	Values     TupleValues
	fields     []SqliField
}

func (sq *SqliTuple) Command() uint16 {
	return 14
}

func (sq *SqliTuple) Pack() ([]byte, error) {
	buffer := new(bytes.Buffer)
	packer := binpacker.NewPacker(binary.BigEndian, buffer)
	packer.PushUint16(sq.Command())
	packer.PushUint16(sq.Warnings)

	if sq.tupleBytes == nil || len(sq.tupleBytes) == 0 {
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
	} else {
		packer.PushUint32(sq.size)
		packer.PushBytes(sq.tupleBytes)
		if sq.size%2 == 1 {
			packer.PushByte(0)
		}
	}
	return buffer.Bytes(), nil
}

func (sq *SqliTuple) Unpack(r io.Reader) error {
	unpacker := binpacker.NewUnpacker(binary.BigEndian, r)
	unpacker.FetchUint16(&sq.Warnings).FetchUint32(&sq.size)
	err := unpacker.Error()
	if err != nil {
		return err
	}
	sq.tupleBytes = make([]byte, sq.size)
	_, err = r.Read(sq.tupleBytes)
	if err != nil {
		return err
	}
	if sq.size%2 == 1 {
		var pad byte
		unpacker.FetchByte(&pad) // Pad
	}
	return unpacker.Error()
}

func (sq *SqliTuple) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Warnings   uint16
		Size       uint32
		TupleBytes []byte
		Values     TupleValues `json:"values,omitempty"`
		Fields     []SqliField `json:"fields,omitempty"`
	}{
		Warnings:   sq.Warnings,
		Size:       sq.size,
		TupleBytes: sq.tupleBytes,
		Values:     sq.Values,
		Fields:     sq.fields,
	})
}

func (sq *SqliTuple) UnmarshalJSON(data []byte) (err error) {
	var tuple struct {
		Warnings   uint16
		Size       uint32
		TupleBytes []byte
		Values     TupleValues `json:"values,omitempty"`
		Fields     []SqliField `json:"fields,omitempty"`
	}
	err = json.Unmarshal(data, &tuple)

	sq.Warnings = tuple.Warnings
	sq.size = tuple.Size
	sq.tupleBytes = tuple.TupleBytes
	sq.Values = tuple.Values
	sq.fields = tuple.Fields

	return
}

var ERRSQLITUPLENOMETADATA = errors.New("can't find metadata")

func (sq *SqliTuple) UnpackValues() error {
	if sq.fields == nil {
		return ERRSQLITUPLENOMETADATA
	}
	sq.Values = nil
	sq.Values = make([]TupleValue, 0, len(sq.fields))
	for _, f := range sq.fields {
		switch f.ColumnType & 0xff {
		case 0:
			sq.Values = append(sq.Values, &CharTupleValue{Length: f.Length})
		case 1:
			sq.Values = append(sq.Values, &SmallIntTupleValue{})
		case 2:
			sq.Values = append(sq.Values, &IntTupleValue{})
		case 13:
			sq.Values = append(sq.Values, &VarcharTupleValue{})
		case 43:
			sq.Values = append(sq.Values, &LVarcharTupleValue{})
		default:
			log.Printf("unknown data type: %d\n", f.ColumnType)
			return errors.New("unknown data type")
		}
	}
	r := bytes.NewReader(sq.tupleBytes)
	for _, v := range sq.Values {
		err := v.UnpackTupleValue(r)
		if err != nil {
			return err
		}
	}

	return nil
}
func (sq *SqliTuple) SetMetaData(fields []SqliField) {
	sq.fields = fields
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

// SqliDone SQ_DONE 15
type SqliDone struct {
	Warning  int16
	Rows     uint32
	RowID    uint32
	SerialID uint32
}

func (*SqliDone) Command() uint16 {
	return SQ_DONE
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

func (sq *SqliDone) Unpack(r io.Reader) error {
	unpacker := binpacker.NewUnpacker(binary.BigEndian, r)
	unpacker.FetchInt16(&sq.Warning).
		FetchUint32(&sq.Rows).
		FetchUint32(&sq.RowID).
		FetchUint32(&sq.SerialID)
	return unpacker.Error()
}

// SqliCmmtwork SQ_CMMTWORK 19
type SqliCmmtwork struct {
}

func (*SqliCmmtwork) Command() uint16 {
	return 19
}

func (*SqliCmmtwork) Pack() ([]byte, error) {
	return []byte{0, 19}, nil
}

func (*SqliCmmtwork) Unpack(r io.Reader) error {
	panic("implement me")
}

// SqliNDescribe SQ_NDESCRIBE 22
type SqliNDescribe struct {
}

func (*SqliNDescribe) Command() uint16 {
	return SQ_NDESCRIBE
}

func (*SqliNDescribe) Pack() ([]byte, error) {
	return []byte{0, 22}, nil
}

func (*SqliNDescribe) Unpack(r io.Reader) error {
	return nil
}

// SqliBegin SQ_BEGIN 35
type SqliBegin struct {
}

func (*SqliBegin) Command() uint16 {
	return 35
}

func (*SqliBegin) Pack() ([]byte, error) {
	return []byte{0, 35}, nil
}

func (*SqliBegin) Unpack(r io.Reader) error {
	panic("implement me")
}

// SqliDBOpen SQ_DBOPEN 36
type SqliDBOpen struct {
	DBName string
	Foo    int16 // TODO: Naming
}

func (sq *SqliDBOpen) Command() uint16 {
	return SQ_DBOPEN
}

func (sq *SqliDBOpen) Pack() ([]byte, error) {
	buffer := new(bytes.Buffer)
	packer := binpacker.NewPacker(binary.BigEndian, buffer)
	packer.PushUint16(sq.Command())
	packer.PushInt16(int16(len(sq.DBName))).
		PushString(sq.DBName)
	if len(sq.DBName)%2 == 1 {
		packer.PushByte(0)
	}

	packer.PushInt16(sq.Foo)
	return buffer.Bytes(), packer.Error()
}

func (sq *SqliDBOpen) Unpack(r io.Reader) error {
	unpacker := binpacker.NewUnpacker(binary.BigEndian, r)
	var size uint16
	unpacker.
		FetchUint16(&size).
		FetchString(uint64(size), &sq.DBName)
	if size%2 == 1 {
		var tmp byte
		unpacker.FetchByte(&tmp)
	}
	unpacker.FetchInt16(&sq.Foo)
	return unpacker.Error()
}

// SqliWantDone SQ_WANTDONE 49
type SqliWantDone struct {
}

func (*SqliWantDone) Command() uint16 {
	return SQ_WANTDONE
}

func (*SqliWantDone) Pack() ([]byte, error) {
	return []byte{0, 49}, nil
}

func (*SqliWantDone) Unpack(r io.Reader) error {
	return nil
}

// SqliCost SQ_COST 55
type SqliCost struct {
	EstimatedRows uint32
	EstimatedIO   uint32
}

func (sq *SqliCost) Command() uint16 {
	return SQ_COST
}

func (sq *SqliCost) Pack() ([]byte, error) {
	buffer := new(bytes.Buffer)
	packer := binpacker.NewPacker(binary.BigEndian, buffer)
	packer.PushUint16(sq.Command())
	packer.PushUint32(sq.EstimatedRows).PushUint32(sq.EstimatedIO)
	return buffer.Bytes(), packer.Error()
}

func (sq *SqliCost) Unpack(r io.Reader) error {
	unpacker := binpacker.NewUnpacker(binary.BigEndian, r)
	unpacker.FetchUint32(&sq.EstimatedRows).
		FetchUint32(&sq.EstimatedIO)
	return unpacker.Error()
}

// SqliExit SQ_EXIT 56
type SqliExit struct {
}

func (*SqliExit) Command() uint16 {
	return 56
}

func (*SqliExit) Pack() ([]byte, error) {
	return []byte{0, 56}, nil
}

func (*SqliExit) Unpack(r io.Reader) error {
	return nil
}

// SqliInfo SQ_INFO 81
type SqliInfo struct {
	MessageType int16
	Length      int16
	InfoEnv     InfoEnv
}

func (sq *SqliInfo) Command() uint16 {
	return SQ_INFO
}

func (sq *SqliInfo) Pack() ([]byte, error) {
	buffer := new(bytes.Buffer)
	packer := binpacker.NewPacker(binary.BigEndian, buffer)
	packer.PushUint16(sq.Command()).
		PushInt16(sq.MessageType).PushInt16(sq.Length)
	env, err := sq.InfoEnv.Pack()
	if err != nil {
		return nil, err
	}
	buffer.Write(env)

	packer.PushInt16(0)
	packer.PushInt16(0)
	return buffer.Bytes(), packer.Error()
}

func (sq *SqliInfo) Unpack(r io.Reader) error {
	unpacker := binpacker.NewUnpacker(binary.BigEndian, r)
	unpacker.FetchInt16(&sq.MessageType).
		FetchInt16(&sq.Length)
	err := sq.InfoEnv.Unpack(r)
	if err != nil {
		return err
	}
	var temp int16
	unpacker.FetchInt16(&temp)
	return unpacker.Error()
}

// SqliInsertDone SQ_INSERTDONE 94
type SqliInsertDone struct {
	Serial8   int64
	BigSerial uint64
}

func (sq *SqliInsertDone) Command() uint16 {
	return 94
}

func (sq *SqliInsertDone) Pack() ([]byte, error) {
	buffer := new(bytes.Buffer)
	packer := binpacker.NewPacker(binary.BigEndian, buffer)
	var sign uint16
	var serial uint64
	if sq.Serial8 < 0 {
		sign = 0
		serial = uint64(-sq.Serial8)
	} else {
		sign = 1
		serial = uint64(sq.Serial8)
	}
	packer.PushUint16(sq.Command()).PushUint16(sign).PushUint64(serial).PushUint64(sq.BigSerial)
	return buffer.Bytes(), packer.Error()
}

func (sq *SqliInsertDone) Unpack(r io.Reader) error {
	unpacker := binpacker.NewUnpacker(binary.BigEndian, r)
	var sign uint16
	var serial uint64
	unpacker.FetchUint16(&sign).
		FetchUint64(&serial).FetchUint64(&sq.BigSerial)
	if sign == 0 {
		sq.Serial8 = -int64(serial)
	} else {
		sq.Serial8 = int64(serial)
	}
	return unpacker.Error()
}

type InfoEnv struct {
	NameLength  int16
	ValueLength int16
	Env         map[string]string
}

func (env *InfoEnv) Pack() ([]byte, error) {
	buffer := new(bytes.Buffer)
	packer := binpacker.NewPacker(binary.BigEndian, buffer)
	packer.PushInt16(env.NameLength).PushInt16(env.ValueLength)
	for k, v := range env.Env {
		packer.PushInt16(int16(len(k))).PushString(k)
		if len(k)%2 == 1 {
			packer.PushByte(0) // Pad
		}
		packer.PushInt16(int16(len(v))).PushString(v)
		if len(v)%2 == 1 {
			packer.PushByte(0) // Pad
		}

	}
	return buffer.Bytes(), packer.Error()
}

func (env *InfoEnv) Unpack(r io.Reader) error {
	if env.Env == nil {
		env.Env = make(map[string]string)
	}
	unpacker := binpacker.NewUnpacker(binary.BigEndian, r)
	unpacker.FetchInt16(&env.NameLength).FetchInt16(&env.ValueLength)
	for size, err := unpacker.ShiftInt16(); size != 0; {
		if err != nil {
			return err
		}
		var k, v string
		unpacker.FetchString(uint64(size), &k)
		if size%2 == 1 {
			unpacker.ShiftByte() // Pad
		}
		unpacker.FetchInt16(&size).FetchString(uint64(size), &v)
		if size%2 == 1 {
			unpacker.ShiftByte() // Pad
		}
		if unpacker.Error() != nil {
			return unpacker.Error()
		}
		env.Env[k] = v
		size, err = unpacker.ShiftInt16()
		if err != nil {
			return err
		}
	}
	return nil
}

// SqliXActstat SQ_XACTSTAT 99
type SqliXActstat struct {
	Event    int16
	NewLevel int16
	OldLevel int16
}

func (sq *SqliXActstat) Command() uint16 {
	return 99
}

func (sq *SqliXActstat) Pack() ([]byte, error) {
	buffer := new(bytes.Buffer)
	packer := binpacker.NewPacker(binary.BigEndian, buffer)
	packer.PushUint16(sq.Command()).
		PushInt16(sq.Event).
		PushInt16(sq.NewLevel).
		PushInt16(sq.OldLevel)

	return buffer.Bytes(), packer.Error()
}

func (sq *SqliXActstat) Unpack(r io.Reader) error {
	unpacker := binpacker.NewUnpacker(binary.BigEndian, r)
	unpacker.FetchInt16(&sq.Event).
		FetchInt16(&sq.NewLevel).
		FetchInt16(&sq.OldLevel)
	return unpacker.Error()
}

// SqliRetType SQ_RETTYPE 100
type SqliRetType struct {
	Direction uint16
	Columns   []ColumnType
}

func (sq *SqliRetType) Command() uint16 {
	return 100
}

func (sq *SqliRetType) Pack() ([]byte, error) {
	buffer := new(bytes.Buffer)
	packer := binpacker.NewPacker(binary.BigEndian, buffer)
	packer.PushUint16(sq.Command()).PushUint16(sq.Direction).
		PushUint16(uint16(len(sq.Columns)))
	for _, c := range sq.Columns {
		packer.PushUint16(c.Type)
		if c.Type == 43 /* lvarchar */ || c.Type == 44 /* clob */ || c.Type == 45 /* sendrecv */ {
			packer.PushUint16(uint16(len(c.OwnerName)))
			packer.PushBytes([]byte(c.OwnerName))
			if len(c.OwnerName)%2 == 1 {
				packer.PushByte(0)
			}
			packer.PushUint16(uint16(len(c.ExtTypeName)))
			packer.PushBytes([]byte(c.ExtTypeName))
			if len(c.ExtTypeName)%2 == 1 {
				packer.PushByte(0)
			}
		}
		packer.PushUint32(c.Length)
	}
	return buffer.Bytes(), packer.Error()
}

func (sq *SqliRetType) Unpack(r io.Reader) error {
	unpacker := binpacker.NewUnpacker(binary.BigEndian, r)
	var count uint16
	unpacker.FetchUint16(&sq.Direction).
		FetchUint16(&count)
	sq.Columns = make([]ColumnType, 0, count)
	for i := 0; i < int(count); i++ {
		var c ColumnType
		unpacker.FetchUint16(&c.Type)
		if c.Type == 43 /* lvarchar */ || c.Type == 44 /* clob */ || c.Type == 45 /* sendrecv */ {
			var length uint16
			var name string
			// owner name
			unpacker.FetchUint16(&length)
			unpacker.FetchString(uint64(length), &name)
			c.OwnerName = name
			if length%2 == 1 {
				var tmp byte
				unpacker.FetchByte(&tmp)
			}
			// extended type name
			unpacker.FetchUint16(&length)
			unpacker.FetchString(uint64(length), &name)
			c.ExtTypeName = name
			if length%2 == 1 {
				var tmp byte
				unpacker.FetchByte(&tmp)
			}
		}
		unpacker.FetchUint32(&c.Length)
		sq.Columns = append(sq.Columns, c)
	}
	return unpacker.Error()
}

type ColumnType struct {
	Type        uint16
	Length      uint32
	OwnerName   string
	ExtTypeName string
}

// SqliAutoFree SQ_AUTOFREE 108
type SqliAutoFree struct {
}

func (*SqliAutoFree) Command() uint16 {
	return 108
}

func (*SqliAutoFree) Pack() ([]byte, error) {
	return []byte{0, 108}, nil
}

func (*SqliAutoFree) Unpack(r io.Reader) error {
	panic("implement me")
}

// SqliCIdescribe SQ_CIDESCEIBE
type SqliCIdescribe struct {
}

func (*SqliCIdescribe) Command() uint16 {
	return 124
}

func (*SqliCIdescribe) Pack() ([]byte, error) {
	return []byte{0, 124}, nil
}

func (*SqliCIdescribe) Unpack(r io.Reader) error {
	panic("implement me")
}

// SqliIdescribe SQ_IDESCEIBE
type SqliIdescribe struct {
	Inputfields uint16
	Fields      []Sqlifields
}

func (sq *SqliIdescribe) Command() uint16 {
	return 125
}

func (sq *SqliIdescribe) Pack() ([]byte, error) {
	buffer := new(bytes.Buffer)
	packer := binpacker.NewPacker(binary.BigEndian, buffer)
	packer.PushUint16(sq.Command()).PushUint16(sq.Inputfields).PushUint16(uint16(len(sq.Fields)))
	for _, c := range sq.Fields {
		packer.PushUint16(c.Type)
		packer.PushUint32(c.ExtendID)
		packer.PushUint16(c.OwnerNameLength)
		packer.PushUint16(c.ExtendTypeNameLength)
		packer.PushUint16(c.PassByReferenceFlag)
		packer.PushUint16(c.Alignment)
		packer.PushUint32(c.SourceType)
		packer.PushUint32(c.Length)
	}
	return buffer.Bytes(), packer.Error()
}

func (sq *SqliIdescribe) Unpack(r io.Reader) error {
	unpacker := binpacker.NewUnpacker(binary.BigEndian, r)
	var count uint16
	unpacker.FetchUint16(&count)
	unpacker.FetchUint16(&sq.Inputfields)
	sq.Fields = make([]Sqlifields, 0, count)
	for i := 0; i < int(count); i++ {
		var c Sqlifields
		unpacker.FetchUint16(&c.Type)
		unpacker.FetchUint32(&c.ExtendID)
		unpacker.FetchUint16(&c.OwnerNameLength)
		unpacker.FetchUint16(&c.ExtendTypeNameLength)
		unpacker.FetchUint16(&c.PassByReferenceFlag)
		unpacker.FetchUint16(&c.Alignment)
		unpacker.FetchUint32(&c.SourceType)
		unpacker.FetchUint32(&c.Length)
		sq.Fields = append(sq.Fields, c)
	}
	return unpacker.Error()
}

type Sqlifields struct {
	Type                 uint16
	ExtendID             uint32
	OwnerNameLength      uint16
	ExtendTypeNameLength uint16
	PassByReferenceFlag  uint16
	Alignment            uint16
	SourceType           uint32
	Length               uint32
}

// SqliProtocols SQ_PROTOCOLS 126
type SqliProtocols struct {
	Protocol []byte
}

func (sq *SqliProtocols) Command() uint16 {
	return SQ_PROTOCOLS
}

func (sq *SqliProtocols) Pack() ([]byte, error) {
	buffer := new(bytes.Buffer)
	packer := binpacker.NewPacker(binary.BigEndian, buffer)
	packer.PushUint16(sq.Command()).
		PushUint16(uint16(len(sq.Protocol))).
		PushBytes(sq.Protocol)
	if len(sq.Protocol)%2 == 1 {
		packer.PushByte(0)
	}
	return buffer.Bytes(), packer.Error()
}

func (sq *SqliProtocols) Unpack(r io.Reader) error {
	unpacker := binpacker.NewUnpacker(binary.BigEndian, r)
	var size uint16
	unpacker.FetchUint16(&size)
	err := unpacker.Error()
	if err != nil {
		return err
	}
	sq.Protocol = make([]byte, size)
	_, err = r.Read(sq.Protocol)
	if err != nil {
		return err
	}
	if size%2 == 1 {
		var pad byte
		unpacker.FetchByte(&pad) // Pad
	}
	return unpacker.Error()
}

func UnpackSqliCommand(reader io.ReadSeeker) (SqliCommand, error) {
	var cmd SqliType
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
	case SQ_COMMAND:
		command := &SqliCmd{}
		err = command.Unpack(reader)
		if err != nil {
			reader.Seek(pos, io.SeekStart)
			return nil, err
		}
		return command, nil
	case 2:
		command := &SqliPrepare{}
		err = command.Unpack(reader)
		if err != nil {
			reader.Seek(pos, io.SeekStart)
			return nil, err
		}
		return command, nil
	case 3:
		command := &SqliCurName{}
		err = command.Unpack(reader)
		if err != nil {
			reader.Seek(pos, io.SeekStart)
			return nil, err
		}
		return command, nil
	case 4:
		command := &SqliID{}
		err = command.Unpack(reader)
		if err != nil {
			reader.Seek(pos, io.SeekStart)
			return nil, err
		}
		return command, nil
	case 5:
		command := &SqliBind{}
		err = command.Unpack(reader)
		if err != nil {
			reader.Seek(pos, io.SeekStart)
			return nil, err
		}
		return command, nil
	case 6:
		command := &SqliOpen{}
		return command, nil
	case 7:
		command := &SqliExecute{}
		return command, nil
	case 8:
		command := &SqliDescribe{}
		err = command.Unpack(reader)
		if err != nil {
			reader.Seek(pos, io.SeekStart)
			return nil, err
		}
		return command, nil
	case 9:
		command := &SqliNFetch{}
		err = command.Unpack(reader)
		if err != nil {
			reader.Seek(pos, io.SeekStart)
			return nil, err
		}
		return command, nil
	case 10:
		command := &SqliClose{}
		return command, nil
	case 11:
		command := &SqliRelease{}
		return command, nil
	case 12:
		command := &SqliEot{}
		return command, nil
	case 13:
		command := &SqliErr{}
		err = command.Unpack(reader)
		if err != nil {
			reader.Seek(pos, io.SeekStart)
			return nil, err
		}
		return command, nil
	case 14:
		command := &SqliTuple{}
		err = command.Unpack(reader)
		if err != nil {
			reader.Seek(pos, io.SeekStart)
			return nil, err
		}
		return command, nil
	case SQ_DONE:
		command := &SqliDone{}
		err = command.Unpack(reader)
		if err != nil {
			reader.Seek(pos, io.SeekStart)
			return nil, err
		}
		return command, nil
	case 19:
		command := &SqliCmmtwork{}
		return command, nil
	case SQ_NDESCRIBE:
		command := &SqliNDescribe{}
		return command, nil
	case 35:
		command := &SqliBegin{}
		return command, nil
	case SQ_DBOPEN:
		command := &SqliDBOpen{}
		err = command.Unpack(reader)
		if err != nil {
			reader.Seek(pos, io.SeekStart)
			return nil, err
		}
		return command, nil
	case SQ_WANTDONE:
		command := &SqliWantDone{}
		return command, nil
	case SQ_COST:
		command := &SqliCost{}
		err = command.Unpack(reader)
		if err != nil {
			reader.Seek(pos, io.SeekStart)
			return nil, err
		}
		return command, nil
	case 56:
		command := &SqliExit{}
		return command, nil
	case SQ_INFO:
		command := &SqliInfo{}
		err = command.Unpack(reader)
		if err != nil {
			reader.Seek(pos, io.SeekStart)
			return nil, err
		}
		return command, nil
	case 94:
		command := &SqliInsertDone{}
		err = command.Unpack(reader)
		if err != nil {
			reader.Seek(pos, io.SeekStart)
			return nil, err
		}
		return command, nil
	case 99:
		command := &SqliXActstat{}
		err = command.Unpack(reader)
		if err != nil {
			reader.Seek(pos, io.SeekStart)
			return nil, err
		}
		return command, nil
	case 100:
		command := &SqliRetType{}
		err = command.Unpack(reader)
		if err != nil {
			reader.Seek(pos, io.SeekStart)
			return nil, err
		}
		return command, nil
	case 108:
		command := &SqliAutoFree{}
		return command, nil
	case 124:
		command := &SqliCIdescribe{}
		return command, nil
	case 125:
		command := &SqliIdescribe{}
		err = command.Unpack(reader)
		if err != nil {
			reader.Seek(pos, io.SeekStart)
			return nil, err
		}
		return command, nil
	case SQ_PROTOCOLS:
		command := &SqliProtocols{}
		err = command.Unpack(reader)
		if err != nil {
			reader.Seek(pos, io.SeekStart)
			return nil, err
		}
		return command, nil
	default:
		return nil, UnknownCommandError(uint16(cmd))
	}
}

func UnknownCommandError(cmd uint16) error {
	return errors.New(fmt.Sprintf("unknown command: %d", cmd))
}

func UnpackSqliTransmission(reader io.ReadSeeker) (SqliTransmission, error) {
	trans := make(SqliTransmission, 0, 5)
	for {
		cmd, err := UnpackSqliCommand(reader)
		if err == io.EOF {
			return trans, nil
		}
		if err != nil {
			return nil, err
		}
		trans.Append(cmd)
	}
	return trans, nil
}
