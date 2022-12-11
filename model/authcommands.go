package model

import (
	"bytes"
	"encoding/binary"
	//"fmt"
	"github.com/zhuangsirui/binpacker"
	"io"
)

type AuthCommand interface {
	Pack() ([]byte, error)
	Unpack(r io.Reader) error
}

// request
type AuthRequest struct {
	Length  uint16
	Noname1 uint8
	Noname2 uint8
	Noname3 uint16

	Noname4          uint16
	Noname5          uint16
	Noname6          uint32
	Ieeemlength      uint16
	Ieeem            string
	Noname7          uint16
	Sqlexec          string
	Versionlength    uint16
	Version          string
	Numberlength     uint16
	Rds              string
	Sqlilength       uint16
	Sqli             string
	Noname8          uint32
	Noname9          uint32
	Noname10         uint32
	Noname11         uint16
	Clientnamelength uint16
	Clientname       string
	Passwordlength   uint16
	Password         string
	Noname12         string
	Noname13         uint32
	Tlitcp           string
	Noname14         uint32
	Noname15         uint16
	Asf              uint16
	Noname16         uint32
	Servernamelength uint16
	Servername       string
	Noname17         uint16
	Noname18         uint16
	Noname19         uint16
	Noname20         uint16
	Noname21         uint16
	Noname22         uint16
	Noname23         uint16
	Dpath            []DPath
	Noname24         uint16
	Noname25         uint32
	Noname26         uint32
	Longthreadid     uint32
	Hostnamelength   uint16
	Noname27         string
	Noname28         uint16
	Directorylength  uint16
	Directory        string
	Noname29         uint16
	Appnamelengthall uint16
	Noname30         uint32
	Noname31         uint32
	Appnamelength    uint16
	Appname          string
	Asceot           uint16
}

func (au *AuthRequest) Pack() ([]byte, error) {
	var pad byte
	buffer := new(bytes.Buffer)
	packer := binpacker.NewPacker(binary.BigEndian, buffer)
	packer.PushUint16(au.Length)
	packer.PushUint8(au.Noname1)
	packer.PushUint8(au.Noname2)
	packer.PushUint16(au.Noname3)
	packer.PushUint16(au.Noname4)
	packer.PushUint16(au.Noname5)
	packer.PushUint32(au.Noname6)
	packer.PushUint16(au.Ieeemlength)
	packer.PushBytes([]byte(au.Ieeem))
	packer.PushByte(pad)
	packer.PushUint16(au.Noname7)
	packer.PushBytes([]byte(au.Sqlexec))
	for i := 0; i < 5; i++ {
		packer.PushByte(pad)
	}
	packer.PushUint16(au.Versionlength)
	packer.PushBytes([]byte(au.Version))
	packer.PushByte(pad)
	packer.PushUint16(au.Numberlength)
	packer.PushBytes([]byte(au.Rds))
	packer.PushByte(pad)
	packer.PushUint16(au.Sqlilength)
	packer.PushBytes([]byte(au.Sqli))
	packer.PushByte(pad)
	packer.PushUint32(au.Noname8)
	packer.PushUint32(au.Noname9)
	packer.PushUint32(au.Noname10)
	packer.PushUint16(au.Noname11)
	packer.PushUint16(au.Clientnamelength)
	packer.PushBytes([]byte(au.Clientname))
	packer.PushByte(pad)
	packer.PushUint16(au.Passwordlength)
	for i := 1; i <= 8; i++ {
		packer.PushUint8(1)
	}
	packer.PushBytes([]byte(au.Password))
	packer.PushByte(pad)
	packer.PushBytes([]byte(au.Noname12))
	for i := 0; i < 6; i++ {
		packer.PushByte(pad)
	}
	packer.PushUint32(au.Noname13)
	packer.PushBytes([]byte(au.Tlitcp))
	packer.PushByte(pad)
	packer.PushByte(pad)
	packer.PushUint32(au.Noname14)
	packer.PushUint16(au.Noname15)
	packer.PushUint16(au.Asf)
	packer.PushUint32(au.Noname16)
	packer.PushUint16(au.Servernamelength)
	packer.PushBytes([]byte(au.Servername))
	packer.PushByte(pad)
	packer.PushUint16(au.Noname17)
	packer.PushUint16(au.Noname18)
	packer.PushUint16(au.Noname19)
	packer.PushUint16(au.Noname20)
	packer.PushUint16(au.Noname21)
	packer.PushUint16(au.Noname22)
	packer.PushUint16(au.Noname23)
	for _, c := range au.Dpath {
		packer.PushUint16(c.Dbpathlength)
		packer.PushBytes([]byte(c.Dbpath))
		packer.PushByte(pad)
		packer.PushUint16(c.Dbpathattributelength)
		packer.PushBytes([]byte(c.Dbpathattribute))
		packer.PushByte(pad)
	}
	packer.PushUint16(au.Noname24)
	packer.PushUint32(au.Noname25)
	packer.PushUint32(au.Noname26)
	packer.PushUint32(au.Longthreadid)
	packer.PushUint16(au.Hostnamelength)
	packer.PushBytes([]byte(au.Noname27))
	packer.PushByte(pad)
	packer.PushUint16(au.Noname28)
	packer.PushUint16(au.Directorylength)
	packer.PushBytes([]byte(au.Directory))
	packer.PushByte(pad)
	packer.PushUint16(au.Noname29)
	packer.PushUint16(au.Appnamelengthall)
	packer.PushUint32(au.Noname30)
	packer.PushUint32(au.Noname31)
	packer.PushUint16(au.Appnamelength)
	packer.PushBytes([]byte(au.Appname))
	packer.PushByte(pad)
	packer.PushUint16(au.Asceot)

	return buffer.Bytes(), packer.Error()
}

func (au *AuthRequest) Unpack(r io.Reader) error {
	var pad byte
	unpacker := binpacker.NewUnpacker(binary.BigEndian, r)
	unpacker.FetchUint16(&au.Length)
	unpacker.FetchUint8(&au.Noname1)
	unpacker.FetchUint8(&au.Noname2)
	unpacker.FetchUint16(&au.Noname3)

	unpacker.FetchUint16(&au.Noname4)
	unpacker.FetchUint16(&au.Noname5)
	unpacker.FetchUint32(&au.Noname6)
	unpacker.FetchUint16(&au.Ieeemlength)
	unpacker.FetchString(uint64(au.Ieeemlength-1), &au.Ieeem)
	unpacker.FetchByte(&pad)
	unpacker.FetchUint16(&au.Noname7)
	unpacker.FetchString(7, &au.Sqlexec)
	for i := 0; i < 5; i++ {
		unpacker.FetchByte(&pad)
	}
	unpacker.FetchUint16(&au.Versionlength)
	unpacker.FetchString(uint64(au.Versionlength-1), &au.Version)
	unpacker.FetchByte(&pad)
	unpacker.FetchUint16(&au.Numberlength)
	unpacker.FetchString(11, &au.Rds)
	unpacker.FetchByte(&pad)
	unpacker.FetchUint16(&au.Sqlilength)
	unpacker.FetchString(uint64(au.Sqlilength-1), &au.Sqli)
	unpacker.FetchByte(&pad)
	unpacker.FetchUint32(&au.Noname8)
	unpacker.FetchUint32(&au.Noname9)
	unpacker.FetchUint32(&au.Noname10)
	unpacker.FetchUint16(&au.Noname11)
	unpacker.FetchUint16(&au.Clientnamelength)
	unpacker.FetchString(uint64(au.Clientnamelength-1), &au.Clientname)
	unpacker.FetchByte(&pad)
	unpacker.FetchUint16(&au.Passwordlength)
	for i := 1; i <= 8; i++ {
		unpacker.FetchByte(&pad)
	}
	unpacker.FetchString(uint64(au.Passwordlength-9), &au.Password)
	unpacker.FetchByte(&pad)
	unpacker.FetchString(2, &au.Noname12)
	for i := 0; i < 6; i++ {
		unpacker.FetchByte(&pad)
	}
	unpacker.FetchUint32(&au.Noname13)
	unpacker.FetchString(6, &au.Tlitcp)
	unpacker.FetchByte(&pad)
	unpacker.FetchByte(&pad)
	unpacker.FetchUint32(&au.Noname14)
	unpacker.FetchUint16(&au.Noname15)
	unpacker.FetchUint16(&au.Asf)
	unpacker.FetchUint32(&au.Noname16)
	unpacker.FetchUint16(&au.Servernamelength)
	unpacker.FetchString(uint64(au.Servernamelength-1), &au.Servername)
	unpacker.FetchByte(&pad)
	unpacker.FetchUint16(&au.Noname17)
	unpacker.FetchUint16(&au.Noname18)
	unpacker.FetchUint16(&au.Noname19)
	unpacker.FetchUint16(&au.Noname20)
	unpacker.FetchUint16(&au.Noname21)
	unpacker.FetchUint16(&au.Noname22)
	unpacker.FetchUint16(&au.Noname23)
	for i := 1; i < 6; i++ {
		var c DPath
		unpacker.FetchUint16(&c.Dbpathlength)
		unpacker.FetchString(uint64(c.Dbpathlength-1), &c.Dbpath)
		unpacker.FetchByte(&pad)
		unpacker.FetchUint16(&c.Dbpathattributelength)
		unpacker.FetchString(uint64(c.Dbpathattributelength-1), &c.Dbpathattribute)
		unpacker.FetchByte(&pad)
		au.Dpath = append(au.Dpath, c)
	}
	unpacker.FetchUint16(&au.Noname24)
	unpacker.FetchUint32(&au.Noname25)
	unpacker.FetchUint32(&au.Noname26)
	unpacker.FetchUint32(&au.Longthreadid)
	unpacker.FetchUint16(&au.Hostnamelength)
	unpacker.FetchString(uint64(au.Hostnamelength-1), &au.Noname27)
	unpacker.FetchByte(&pad)
	unpacker.FetchUint16(&au.Noname28)
	unpacker.FetchUint16(&au.Directorylength)
	unpacker.FetchString(uint64(au.Directorylength-1), &au.Directory)
	unpacker.FetchByte(&pad)
	unpacker.FetchUint16(&au.Noname29)
	unpacker.FetchUint16(&au.Appnamelengthall)
	unpacker.FetchUint32(&au.Noname30)
	unpacker.FetchUint32(&au.Noname31)
	unpacker.FetchUint16(&au.Appnamelength)
	unpacker.FetchString(uint64(au.Appnamelength-1), &au.Appname)
	unpacker.FetchByte(&pad)
	unpacker.FetchUint16(&au.Asceot)

	return unpacker.Error()
}

type DPath struct {
	Dbpathlength          uint16
	Dbpath                string
	Dbpathattributelength uint16
	Dbpathattribute       string
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

//func UnpackAuthRequest(reader io.ReadSeeker) (AuthCommand, error) {
//	error := "UnpackauthrequestErr"
//	pos, err := reader.Seek(0, io.SeekCurrent)
//	if err != nil {
//		return nil, err
//	}
//	err = binary.Read(reader, binary.BigEndian, &error)
//	if err != nil{
//		reader.Seek(pos, io.SeekStart)
//		return nil, err
//	}
//
//}
