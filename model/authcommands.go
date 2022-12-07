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
	header []Header
	body   []Body
}

func (au *AuthRequest) Pack() ([]byte, error) {
	buffer := new(bytes.Buffer)
	packer := binpacker.NewPacker(binary.BigEndian, buffer)
	packer.PushUint16(uint16(len(au.header) + len(au.body)))
	for _, c := range au.header {
		packer.PushUint8(c.noname1)
		packer.PushUint8(c.noname2)
		packer.PushUint16(c.noname3)
	}
	for _, b := range au.body {
		packer.PushUint16(b.noname1)
		packer.PushUint16(b.noname2)
		packer.PushUint32(b.noname3)
		packer.PushInt16(int16(b.ieeemlength))
		packer.PushBytes([]byte(b.ieeem))
		packer.PushUint8(b.non)
		packer.PushInt16(int16(b.noname4))
		packer.PushBytes([]byte(b.sqlexec))
		packer.PushUint16(b.versionlength)
		packer.PushBytes([]byte(b.version))
		packer.PushUint8(b.noname5)
		packer.PushInt16(int16(b.numberlength))
		packer.PushUint8(b.rds)
		packer.PushUint8(b.noname6)
		packer.PushUint16(b.sqlilength)
		packer.PushBytes([]byte(b.sqli))
		packer.PushUint8(b.noname7)
		packer.PushUint32(b.noname8)
		packer.PushUint32(b.noname9)
		packer.PushUint32(b.noname10)
		packer.PushUint16(b.noname11)
		packer.PushUint16(b.clientnamelength)
		packer.PushBytes([]byte(b.clientname))
		packer.PushUint8(b.noname12)
		packer.PushUint16(b.passwordlength)
		packer.PushBytes([]byte(b.password))
		packer.PushUint8(b.noname13)
		packer.PushBytes([]byte(b.noname14))
		packer.PushUint32(b.noname15)
		packer.PushBytes([]byte(b.tlitcp))
		packer.PushUint32(b.noname16)
		packer.PushUint16(b.noname17)
		packer.PushUint16(b.asf)
		packer.PushUint32(b.noname18)
		packer.PushUint16(b.servernamelength)
		packer.PushBytes([]byte(b.servername))
		packer.PushUint8(b.noname19)
		packer.PushUint16(b.noname20)
		packer.PushUint16(b.noname21)
		packer.PushUint16(b.noname22)
		packer.PushUint16(b.noname23)
		packer.PushUint16(b.noname24)
		packer.PushUint16(b.noname25)
		packer.PushUint16(b.noname26)
		packer.PushUint16(b.dbpathlength)
		packer.PushBytes([]byte(b.dbpath))
		packer.PushUint8(b.noname27)
		packer.PushUint16(b.dbpathattributelength)
		packer.PushBytes([]byte(b.dbpathattribute))
		packer.PushUint8(b.noname28)
		packer.PushUint16(b.noname29)
		packer.PushUint32(b.noname30)
		packer.PushUint32(b.noname31)
		packer.PushUint16(b.hostnamelength)
		packer.PushBytes([]byte(b.noname32))
		packer.PushUint8(b.noname33)
		packer.PushUint16(b.noname34)
		packer.PushUint16(b.directorylength)
		packer.PushBytes([]byte(b.directory))
		packer.PushUint8(b.noname35)
		packer.PushUint16(b.noname36)
		packer.PushUint16(b.appnamelengthall)
		packer.PushUint32(b.noname37)
		packer.PushUint32(b.noname38)
		packer.PushUint16(b.appnamelength)
		packer.PushBytes([]byte(b.appname))
		packer.PushUint8(b.noname39)
		packer.PushUint16(b.asceot)
	}

	return buffer.Bytes(), packer.Error()
}

func (au *AuthRequest) Unpack(r io.Reader) error {
	unpacker := binpacker.NewUnpacker(binary.BigEndian, r)
	for _, a := range au.header {
		unpacker.FetchUint16(&a.Length)
		unpacker.FetchUint8(&a.noname1)
		unpacker.FetchUint8(&a.noname2)
		unpacker.FetchUint16(&a.noname3)
	}
	for _, c := range au.body {
		unpacker.FetchUint16(&c.noname1)
		unpacker.FetchUint16(&c.noname2)
		unpacker.FetchUint32(&c.noname3)
		unpacker.FetchUint16(&c.ieeemlength)
		unpacker.FetchString(uint64(c.ieeemlength)+1, &c.ieeem)
		unpacker.FetchByte(&c.non)
		unpacker.FetchUint16(&c.noname4)
		unpacker.FetchString(12, &c.sqlexec)
		unpacker.FetchUint16(&c.versionlength)
		unpacker.FetchString(uint64(c.versionlength)+1, &c.version)
		unpacker.FetchByte(&c.noname5)
		unpacker.FetchUint16(&c.numberlength)
		unpacker.FetchByte(&c.rds)
		unpacker.FetchByte(&c.noname6)
		unpacker.FetchUint16(&c.sqlilength)
		unpacker.FetchString(uint64(c.sqlilength)+1, &c.sqli)
		unpacker.FetchByte(&c.noname7)
		unpacker.FetchUint32(&c.noname8)
		unpacker.FetchUint32(&c.noname9)
		unpacker.FetchUint32(&c.noname10)
		unpacker.FetchUint16(&c.noname11)
		unpacker.FetchUint16(&c.clientnamelength)
		unpacker.FetchString(uint64(c.clientnamelength)+1, &c.clientname)
		unpacker.FetchByte(&c.noname12)
		unpacker.FetchUint16(&c.passwordlength)
		unpacker.FetchString(uint64(c.passwordlength), &c.password)
		unpacker.FetchByte(&c.noname13)
		unpacker.FetchString(8, &c.noname14)
		unpacker.FetchUint32(&c.noname15)
		unpacker.FetchString(8, &c.tlitcp)
		unpacker.FetchUint32(&c.noname16)
		unpacker.FetchUint16(&c.noname17)
		unpacker.FetchUint16(&c.asf)
		unpacker.FetchUint32(&c.noname18)
		unpacker.FetchUint16(&c.servernamelength)
		unpacker.FetchString(uint64(c.servernamelength), &c.servername)
		unpacker.FetchByte(&c.noname19)
		unpacker.FetchUint16(&c.noname20)
		unpacker.FetchUint16(&c.noname21)
		unpacker.FetchUint16(&c.noname22)
		unpacker.FetchUint16(&c.noname23)
		unpacker.FetchUint16(&c.noname24)
		unpacker.FetchUint16(&c.noname25)
		unpacker.FetchUint16(&c.noname26)
		unpacker.FetchUint16(&c.dbpathlength)
		unpacker.FetchString(uint64(c.dbpathlength), &c.dbpath)
		unpacker.FetchByte(&c.noname27)
		unpacker.FetchUint16(&c.dbpathattributelength)
		unpacker.FetchString(uint64(c.dbpathattributelength), &c.dbpathattribute)
		unpacker.FetchByte(&c.noname28)
		unpacker.FetchUint16(&c.noname29)
		unpacker.FetchUint32(&c.noname30)
		unpacker.FetchUint32(&c.noname31)
		unpacker.FetchUint32(&c.longthreadid)
		unpacker.FetchUint16(&c.hostnamelength)
		unpacker.FetchString(uint64(c.hostnamelength), &c.noname32)
		unpacker.FetchUint8(&c.noname33)
		unpacker.FetchUint16(&c.noname34)
		unpacker.FetchUint16(&c.directorylength)
		unpacker.FetchString(uint64(c.directorylength), &c.directory)
		unpacker.FetchByte(&c.noname35)
		unpacker.FetchUint16(&c.noname36)
		unpacker.FetchUint16(&c.appnamelengthall)
		unpacker.FetchUint32(&c.noname37)
		unpacker.FetchUint32(&c.noname38)
		unpacker.FetchUint16(&c.appnamelength)
		unpacker.FetchString(uint64(c.appnamelength), &c.appname)
		unpacker.FetchByte(&c.noname39)
		unpacker.FetchUint16(&c.asceot)
	}

	return unpacker.Error()
}

type Header struct {
	Length  uint16
	noname1 uint8
	noname2 uint8
	noname3 uint16
}
type Body struct {
	noname1               uint16
	noname2               uint16
	noname3               uint32
	ieeemlength           uint16
	ieeem                 string
	non                   uint8
	noname4               uint16
	sqlexec               string
	versionlength         uint16
	version               string
	noname5               uint8
	numberlength          uint16
	rds                   uint8
	noname6               uint8
	sqlilength            uint16
	sqli                  string
	noname7               uint8
	noname8               uint32
	noname9               uint32
	noname10              uint32
	noname11              uint16
	clientnamelength      uint16
	clientname            string
	noname12              uint8
	passwordlength        uint16
	password              string
	noname13              uint8
	noname14              string
	noname15              uint32
	tlitcp                string
	noname16              uint32
	noname17              uint16
	asf                   uint16
	noname18              uint32
	servernamelength      uint16
	servername            string
	noname19              uint8
	noname20              uint16
	noname21              uint16
	noname22              uint16
	noname23              uint16
	noname24              uint16
	noname25              uint16
	noname26              uint16
	dbpathlength          uint16
	dbpath                string
	noname27              uint8
	dbpathattributelength uint16
	dbpathattribute       string
	noname28              uint8
	noname29              uint16
	noname30              uint32
	noname31              uint32
	longthreadid          uint32
	hostnamelength        uint16
	noname32              string
	noname33              uint8
	noname34              uint16
	directorylength       uint16
	directory             string
	noname35              uint8
	noname36              uint16
	appnamelengthall      uint16
	noname37              uint32
	noname38              uint32
	appnamelength         uint16
	appname               string
	noname39              uint8
	asceot                uint16
}

// response
type AuthResponse struct {
	length  uint16
	context []Context
}

func (au *AuthResponse) Pack() ([]byte, error) {
	buffer := new(bytes.Buffer)
	packer := binpacker.NewPacker(binary.BigEndian, buffer)
	packer.PushUint16(au.length)
	for _, c := range au.context {
		packer.PushByte(c.noname1)
		packer.PushUint16(c.noname2)
		packer.PushByte(c.noname222)
		packer.PushUint16(c.noname3)
		packer.PushUint16(c.noname4)
		packer.PushUint32(c.noname5)
		packer.PushUint16(c.noname6)
		packer.PushUint16(c.noname7)
		packer.PushBytes([]byte(c.noname8))
		packer.PushUint16(c.noname9)
		packer.PushBytes([]byte(c.noname10))
		packer.PushUint16(c.noname11)
		packer.PushBytes([]byte(c.noname12))
		packer.PushUint16(c.noname13)
		packer.PushBytes([]byte(c.noname14))
		packer.PushUint32(c.noname15)
		packer.PushUint32(c.noname16)
		packer.PushUint32(c.noname17)
		packer.PushUint16(c.noname18)
		packer.PushUint16(c.noname19)
		packer.PushUint16(c.noname20)
		packer.PushBytes([]byte(c.noname21))
		packer.PushUint16(c.noname22)
		packer.PushBytes([]byte(c.noname23))
		packer.PushUint16(c.noname24)
		packer.PushUint16(c.noname25)
		packer.PushUint16(c.noname26)
		packer.PushUint16(c.noname27)
	}

	return buffer.Bytes(), packer.Error()
}

func (au *AuthResponse) Unpack(r io.Reader) error {
	unpacker := binpacker.NewUnpacker(binary.BigEndian, r)
	unpacker.FetchUint16(&au.length)
	for _, c := range au.context {
		unpacker.FetchByte(&c.noname1)
		unpacker.FetchUint16(&c.noname2)
		unpacker.FetchByte(&c.noname222)
		unpacker.FetchUint16(&c.noname3)
		unpacker.FetchUint16(&c.noname4)
		unpacker.FetchUint32(&c.noname5)
		unpacker.FetchUint16(&c.noname6)
		unpacker.FetchUint16(&c.noname7)
		unpacker.FetchString(12, &c.noname8)
		unpacker.FetchUint16(&c.noname9)
		unpacker.FetchString(32, &c.noname10)
		unpacker.FetchUint16(&c.noname11)
		unpacker.FetchString(35, &c.noname12)
		unpacker.FetchUint16(&c.noname13)
		unpacker.FetchString(18, &c.noname14)
		unpacker.FetchUint32(&c.noname15)
		unpacker.FetchUint32(&c.noname16)
		unpacker.FetchUint32(&c.noname17)
		unpacker.FetchUint16(&c.noname18)
		unpacker.FetchUint16(&c.noname19)
		unpacker.FetchUint16(&c.noname20)
		unpacker.FetchString(24, &c.noname21)
		unpacker.FetchUint16(&c.noname22)
		unpacker.FetchString(6, &c.noname23)
		unpacker.FetchUint16(&c.noname24)
		unpacker.FetchUint16(&c.noname25)
		unpacker.FetchUint16(&c.noname26)
		unpacker.FetchUint16(&c.noname27)
	}

	return unpacker.Error()
}

type Context struct {
	noname1   uint8
	noname2   uint16
	noname222 uint8
	noname3   uint16
	noname4   uint16
	noname5   uint32
	noname6   uint16
	noname7   uint16
	noname8   string
	noname9   uint16
	noname10  string
	noname11  uint16
	noname12  string
	noname13  uint16
	noname14  string
	noname15  uint32
	noname16  uint32
	noname17  uint32
	noname18  uint16
	noname19  uint16
	noname20  uint16
	noname21  string
	noname22  uint16
	noname23  string
	noname24  uint16
	noname25  uint16
	noname26  uint16
	noname27  uint16
}
