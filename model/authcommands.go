package model

import (
	"bytes"
	"encoding/binary"
	"github.com/zhuangsirui/binpacker"
	"io"
)

type AuthCommand interface {
	Pack() ([]byte, error)
	Unpack(r io.Reader) error
}

// request
type AuthRequest struct {
	Header []Header
	Body   []Body
}

func (au *AuthRequest) Pack() ([]byte, error) {
	buffer := new(bytes.Buffer)
	packer := binpacker.NewPacker(binary.BigEndian, buffer)
	packer.PushUint16(uint16(len(au.Header) + len(au.Body)))
	for _, c := range au.Header {
		packer.PushUint8(c.Noname1)
		packer.PushUint8(c.Noname2)
		packer.PushUint16(c.Noname3)
	}
	for _, b := range au.Body {
		packer.PushUint16(b.Noname1)
		packer.PushUint16(b.Noname2)
		packer.PushUint32(b.Noname3)
		packer.PushInt16(int16(b.Ieeemlength))
		packer.PushBytes([]byte(b.Ieeem))
		packer.PushUint8(b.Non)
		packer.PushInt16(int16(b.Noname4))
		packer.PushBytes([]byte(b.Sqlexec))
		packer.PushUint16(b.Versionlength)
		packer.PushBytes([]byte(b.Version))
		packer.PushUint8(b.Noname5)
		packer.PushInt16(int16(b.Numberlength))
		packer.PushBytes([]byte(b.Rds))
		packer.PushUint8(b.Noname6)
		packer.PushUint16(b.Sqlilength)
		packer.PushBytes([]byte(b.Sqli))
		packer.PushUint8(b.Noname7)
		packer.PushUint32(b.Noname8)
		packer.PushUint32(b.Noname9)
		packer.PushUint32(b.Noname10)
		packer.PushUint16(b.Noname11)
		packer.PushUint16(b.Clientnamelength)
		packer.PushBytes([]byte(b.Clientname))
		packer.PushUint8(b.Noname12)
		packer.PushUint16(b.Passwordlength)
		packer.PushBytes([]byte(b.Password))
		packer.PushUint8(b.Noname13)
		packer.PushBytes([]byte(b.Noname14))
		packer.PushUint32(b.Noname15)
		packer.PushBytes([]byte(b.Tlitcp))
		packer.PushUint32(b.Noname16)
		packer.PushUint16(b.Noname17)
		packer.PushUint16(b.Asf)
		packer.PushUint32(b.Noname18)
		packer.PushUint16(b.Servernamelength)
		packer.PushBytes([]byte(b.Servername))
		packer.PushUint8(b.Noname19)
		packer.PushUint16(b.Noname20)
		packer.PushUint16(b.Noname21)
		packer.PushUint16(b.Noname22)
		packer.PushUint16(b.Noname23)
		packer.PushUint16(b.Noname24)
		packer.PushUint16(b.Noname25)
		packer.PushUint16(b.Noname26)
		packer.PushUint16(b.Dbpathlength)
		packer.PushBytes([]byte(b.Dbpath))
		packer.PushUint8(b.Noname27)
		packer.PushUint16(b.Dbpathattributelength)
		packer.PushBytes([]byte(b.Dbpathattribute))
		packer.PushUint8(b.Noname28)
		packer.PushUint16(b.Noname29)
		packer.PushUint32(b.Noname30)
		packer.PushUint32(b.Noname31)
		packer.PushUint16(b.Hostnamelength)
		packer.PushBytes([]byte(b.Noname32))
		packer.PushUint8(b.Noname33)
		packer.PushUint16(b.Noname34)
		packer.PushUint16(b.Directorylength)
		packer.PushBytes([]byte(b.Directory))
		packer.PushUint8(b.Noname35)
		packer.PushUint16(b.Noname36)
		packer.PushUint16(b.Appnamelengthall)
		packer.PushUint32(b.Noname37)
		packer.PushUint32(b.Noname38)
		packer.PushUint16(b.Appnamelength)
		packer.PushBytes([]byte(b.Appname))
		packer.PushUint8(b.Noname39)
		packer.PushUint16(b.Asceot)
	}

	return buffer.Bytes(), packer.Error()
}

func (au *AuthRequest) Unpack(r io.Reader) error {
	unpacker := binpacker.NewUnpacker(binary.BigEndian, r)
	for _, a := range au.Header {
		unpacker.FetchUint16(&a.Length)
		unpacker.FetchUint8(&a.Noname1)
		unpacker.FetchUint8(&a.Noname2)
		unpacker.FetchUint16(&a.Noname3)
	}
	for _, c := range au.Body {
		unpacker.FetchUint16(&c.Noname1)
		unpacker.FetchUint16(&c.Noname2)
		unpacker.FetchUint32(&c.Noname3)
		unpacker.FetchUint16(&c.Ieeemlength)
		unpacker.FetchString(uint64(c.Ieeemlength)+1, &c.Ieeem)
		unpacker.FetchByte(&c.Non)
		unpacker.FetchUint16(&c.Noname4)
		unpacker.FetchString(12, &c.Sqlexec)
		unpacker.FetchUint16(&c.Versionlength)
		unpacker.FetchString(uint64(c.Versionlength)+1, &c.Version)
		unpacker.FetchByte(&c.Noname5)
		unpacker.FetchUint16(&c.Numberlength)
		unpacker.FetchString(11, &c.Rds)
		unpacker.FetchByte(&c.Noname6)
		unpacker.FetchUint16(&c.Sqlilength)
		unpacker.FetchString(uint64(c.Sqlilength)+1, &c.Sqli)
		unpacker.FetchByte(&c.Noname7)
		unpacker.FetchUint32(&c.Noname8)
		unpacker.FetchUint32(&c.Noname9)
		unpacker.FetchUint32(&c.Noname10)
		unpacker.FetchUint16(&c.Noname11)
		unpacker.FetchUint16(&c.Clientnamelength)
		unpacker.FetchString(uint64(c.Clientnamelength)+1, &c.Clientname)
		unpacker.FetchByte(&c.Noname12)
		unpacker.FetchUint16(&c.Passwordlength)
		unpacker.FetchString(uint64(c.Passwordlength), &c.Password)
		unpacker.FetchByte(&c.Noname13)
		unpacker.FetchString(8, &c.Noname14)
		unpacker.FetchUint32(&c.Noname15)
		unpacker.FetchString(8, &c.Tlitcp)
		unpacker.FetchUint32(&c.Noname16)
		unpacker.FetchUint16(&c.Noname17)
		unpacker.FetchUint16(&c.Asf)
		unpacker.FetchUint32(&c.Noname18)
		unpacker.FetchUint16(&c.Servernamelength)
		unpacker.FetchString(uint64(c.Servernamelength), &c.Servername)
		unpacker.FetchByte(&c.Noname19)
		unpacker.FetchUint16(&c.Noname20)
		unpacker.FetchUint16(&c.Noname21)
		unpacker.FetchUint16(&c.Noname22)
		unpacker.FetchUint16(&c.Noname23)
		unpacker.FetchUint16(&c.Noname24)
		unpacker.FetchUint16(&c.Noname25)
		unpacker.FetchUint16(&c.Noname26)
		unpacker.FetchUint16(&c.Dbpathlength)
		unpacker.FetchString(uint64(c.Dbpathlength), &c.Dbpath)
		unpacker.FetchByte(&c.Noname27)
		unpacker.FetchUint16(&c.Dbpathattributelength)
		unpacker.FetchString(uint64(c.Dbpathattributelength), &c.Dbpathattribute)
		unpacker.FetchByte(&c.Noname28)
		unpacker.FetchUint16(&c.Noname29)
		unpacker.FetchUint32(&c.Noname30)
		unpacker.FetchUint32(&c.Noname31)
		unpacker.FetchUint32(&c.Longthreadid)
		unpacker.FetchUint16(&c.Hostnamelength)
		unpacker.FetchString(uint64(c.Hostnamelength), &c.Noname32)
		unpacker.FetchUint8(&c.Noname33)
		unpacker.FetchUint16(&c.Noname34)
		unpacker.FetchUint16(&c.Directorylength)
		unpacker.FetchString(uint64(c.Directorylength), &c.Directory)
		unpacker.FetchByte(&c.Noname35)
		unpacker.FetchUint16(&c.Noname36)
		unpacker.FetchUint16(&c.Appnamelengthall)
		unpacker.FetchUint32(&c.Noname37)
		unpacker.FetchUint32(&c.Noname38)
		unpacker.FetchUint16(&c.Appnamelength)
		unpacker.FetchString(uint64(c.Appnamelength), &c.Appname)
		unpacker.FetchByte(&c.Noname39)
		unpacker.FetchUint16(&c.Asceot)
	}

	return unpacker.Error()
}

type Header struct {
	Length  uint16
	Noname1 uint8
	Noname2 uint8
	Noname3 uint16
}
type Body struct {
	Noname1               uint16
	Noname2               uint16
	Noname3               uint32
	Ieeemlength           uint16
	Ieeem                 string
	Non                   uint8
	Noname4               uint16
	Sqlexec               string
	Versionlength         uint16
	Version               string
	Noname5               uint8
	Numberlength          uint16
	Rds                   string
	Noname6               uint8
	Sqlilength            uint16
	Sqli                  string
	Noname7               uint8
	Noname8               uint32
	Noname9               uint32
	Noname10              uint32
	Noname11              uint16
	Clientnamelength      uint16
	Clientname            string
	Noname12              uint8
	Passwordlength        uint16
	Password              string
	Noname13              uint8
	Noname14              string
	Noname15              uint32
	Tlitcp                string
	Noname16              uint32
	Noname17              uint16
	Asf                   uint16
	Noname18              uint32
	Servernamelength      uint16
	Servername            string
	Noname19              uint8
	Noname20              uint16
	Noname21              uint16
	Noname22              uint16
	Noname23              uint16
	Noname24              uint16
	Noname25              uint16
	Noname26              uint16
	Dbpathlength          uint16
	Dbpath                string
	Noname27              uint8
	Dbpathattributelength uint16
	Dbpathattribute       string
	Noname28              uint8
	Noname29              uint16
	Noname30              uint32
	Noname31              uint32
	Longthreadid          uint32
	Hostnamelength        uint16
	Noname32              string
	Noname33              uint8
	Noname34              uint16
	Directorylength       uint16
	Directory             string
	Noname35              uint8
	Noname36              uint16
	Appnamelengthall      uint16
	Noname37              uint32
	Noname38              uint32
	Appnamelength         uint16
	Appname               string
	Noname39              uint8
	Asceot                uint16
}

// response
type AuthResponse struct {
	Length  uint16
	Context []Context
}

func (au *AuthResponse) Pack() ([]byte, error) {
	buffer := new(bytes.Buffer)
	packer := binpacker.NewPacker(binary.BigEndian, buffer)
	packer.PushUint16(au.Length)
	for _, c := range au.Context {
		packer.PushByte(c.Noname1)
		packer.PushUint16(c.Noname2)
		packer.PushByte(c.Noname222)
		packer.PushUint16(c.Noname3)
		packer.PushUint16(c.Noname4)
		packer.PushUint32(c.Noname5)
		packer.PushUint16(c.Noname6)
		packer.PushUint16(c.Noname7)
		packer.PushBytes([]byte(c.Noname8))
		packer.PushUint16(c.Noname9)
		packer.PushBytes([]byte(c.Noname10))
		packer.PushUint16(c.Noname11)
		packer.PushBytes([]byte(c.Noname12))
		packer.PushUint16(c.Noname13)
		packer.PushBytes([]byte(c.Noname14))
		packer.PushUint32(c.Noname15)
		packer.PushUint32(c.Noname16)
		packer.PushUint32(c.Noname17)
		packer.PushUint16(c.Noname18)
		packer.PushUint16(c.Noname19)
		packer.PushUint16(c.Noname20)
		packer.PushBytes([]byte(c.Noname21))
		packer.PushUint16(c.Noname22)
		packer.PushBytes([]byte(c.Noname23))
		packer.PushUint16(c.Noname24)
		packer.PushUint16(c.Noname25)
		packer.PushUint16(c.Noname26)
		packer.PushUint16(c.Noname27)
	}

	return buffer.Bytes(), packer.Error()
}

func (au *AuthResponse) Unpack(r io.Reader) error {
	unpacker := binpacker.NewUnpacker(binary.BigEndian, r)
	unpacker.FetchUint16(&au.Length)
	for _, c := range au.Context {
		unpacker.FetchByte(&c.Noname1)
		unpacker.FetchUint16(&c.Noname2)
		unpacker.FetchByte(&c.Noname222)
		unpacker.FetchUint16(&c.Noname3)
		unpacker.FetchUint16(&c.Noname4)
		unpacker.FetchUint32(&c.Noname5)
		unpacker.FetchUint16(&c.Noname6)
		unpacker.FetchUint16(&c.Noname7)
		unpacker.FetchString(12, &c.Noname8)
		unpacker.FetchUint16(&c.Noname9)
		unpacker.FetchString(32, &c.Noname10)
		unpacker.FetchUint16(&c.Noname11)
		unpacker.FetchString(35, &c.Noname12)
		unpacker.FetchUint16(&c.Noname13)
		unpacker.FetchString(18, &c.Noname14)
		unpacker.FetchUint32(&c.Noname15)
		unpacker.FetchUint32(&c.Noname16)
		unpacker.FetchUint32(&c.Noname17)
		unpacker.FetchUint16(&c.Noname18)
		unpacker.FetchUint16(&c.Noname19)
		unpacker.FetchUint16(&c.Noname20)
		unpacker.FetchString(24, &c.Noname21)
		unpacker.FetchUint16(&c.Noname22)
		unpacker.FetchString(6, &c.Noname23)
		unpacker.FetchUint16(&c.Noname24)
		unpacker.FetchUint16(&c.Noname25)
		unpacker.FetchUint16(&c.Noname26)
		unpacker.FetchUint16(&c.Noname27)
	}

	return unpacker.Error()
}

type Context struct {
	Noname1   uint8
	Noname2   uint16
	Noname222 uint8
	Noname3   uint16
	Noname4   uint16
	Noname5   uint32
	Noname6   uint16
	Noname7   uint16
	Noname8   string
	Noname9   uint16
	Noname10  string
	Noname11  uint16
	Noname12  string
	Noname13  uint16
	Noname14  string
	Noname15  uint32
	Noname16  uint32
	Noname17  uint32
	Noname18  uint16
	Noname19  uint16
	Noname20  uint16
	Noname21  string
	Noname22  uint16
	Noname23  string
	Noname24  uint16
	Noname25  uint16
	Noname26  uint16
	Noname27  uint16
}
