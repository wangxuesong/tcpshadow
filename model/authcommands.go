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
	Noname1          uint8
	Noname2          uint8
	Noname3          uint16
	Noname4          uint16
	Noname5          uint16
	Noname6          uint32
	Ieeem            string
	Noname7          uint16
	Sqlexec          string
	Version          string
	Rds              string
	Sqli             string
	Noname8          uint32
	Noname9          uint32
	Noname10         uint32
	Noname11         uint16
	Clientname       string
	Password         string
	Noname12         string
	Noname13         uint32
	Tlitcp           string
	Noname14         uint32
	Noname15         uint16
	Asf              uint16
	Noname16         uint32
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
	Noname27         string
	Noname28         uint16
	Directory        string
	Noname29         uint16
	Appnamelengthall uint16
	Noname30         uint32
	Noname31         uint32
	Appname          string
	Asceot           uint16
}

type AuthRequestBuilder interface {
	BuildVersion()
	BuildClientname()
	BuildPassword()
	BuildServername()
	BuildDirectory()
	BuildAppname()
	CreateActor() *AuthRequest
}

type RequestBuilder struct {
	Aauthrequest AuthRequest
}

func (a *RequestBuilder) BuildVersion() {
	a.Aauthrequest.Version = "9.280"
}

func (a *RequestBuilder) BuildClientname() {
	a.Aauthrequest.Clientname = "gbasedbt"
}

func (a *RequestBuilder) BuildPassword() {
	a.Aauthrequest.Password = "HmQOYC1ZfTYt+vlXUhkn3w=="
}

func (a *RequestBuilder) BuildServername() {
	a.Aauthrequest.Servername = "gbaseserver"
}

func (a *RequestBuilder) BuildDirectory() {
	a.Aauthrequest.Directory = "E:\\JDBCTest\\JDBCTest"
}

func (a *RequestBuilder) BuildAppname() {
	a.Aauthrequest.Appname = "/E:/JDBCTest/JDBCTest/lib/gbasedbtjdbc_3.3.0_2.jarConnectionTest/Test"
}

func (a RequestBuilder) CreateActor() *AuthRequest {
	return &a.Aauthrequest
}

type AuthRequestController struct {
	authRequestBuilder AuthRequestBuilder
}

func (ac *AuthRequestController) Construct() *AuthRequest {
	ac.authRequestBuilder.BuildVersion()
	ac.authRequestBuilder.BuildClientname()
	ac.authRequestBuilder.BuildPassword()
	ac.authRequestBuilder.BuildServername()
	ac.authRequestBuilder.BuildDirectory()
	ac.authRequestBuilder.BuildAppname()
	return ac.authRequestBuilder.CreateActor()
}

func NewAuthRequestController(ab AuthRequestBuilder) *AuthRequestController {
	return &AuthRequestController{
		authRequestBuilder: ab,
	}
}

func (au *AuthRequest) Pack() ([]byte, error) {
	var pad byte
	buffer := new(bytes.Buffer)
	packer := binpacker.NewPacker(binary.BigEndian, buffer)
	length := 141 + len(au.Dpath)*6 + len(au.Ieeem) + len(au.Sqlexec) + len(au.Version) +
		len(au.Rds) + len(au.Sqli) + len(au.Clientname) + len(au.Password) + len(au.Noname12) + len(au.Tlitcp) +
		len(au.Servername) + len(au.Noname27) + len(au.Directory) + len(au.Appname)
	for _, c := range au.Dpath {
		length = length + len(c.Dbpath) + len(c.Dbpathattribute)
	}
	packer.PushUint16(uint16(length))
	packer.PushUint8(au.Noname1)
	packer.PushUint8(au.Noname2)
	packer.PushUint16(au.Noname3)
	packer.PushUint16(au.Noname4)
	packer.PushUint16(au.Noname5)
	packer.PushUint32(au.Noname6)
	packer.PushUint16(uint16(len(au.Ieeem) + 1))
	packer.PushBytes([]byte(au.Ieeem))
	packer.PushByte(pad)
	packer.PushUint16(au.Noname7)
	packer.PushBytes([]byte(au.Sqlexec))
	for i := 0; i < 5; i++ {
		packer.PushByte(pad)
	}
	packer.PushUint16(uint16(len(au.Version) + 1))
	packer.PushBytes([]byte(au.Version))
	packer.PushByte(pad)
	packer.PushUint16(uint16(len(au.Rds) + 1))
	packer.PushBytes([]byte(au.Rds))
	packer.PushByte(pad)
	packer.PushUint16(uint16(len(au.Sqli) + 1))
	packer.PushBytes([]byte(au.Sqli))
	packer.PushByte(pad)
	packer.PushUint32(au.Noname8)
	packer.PushUint32(au.Noname9)
	packer.PushUint32(au.Noname10)
	packer.PushUint16(au.Noname11)
	packer.PushUint16(uint16(len(au.Clientname) + 1))
	packer.PushBytes([]byte(au.Clientname))
	packer.PushByte(pad)
	packer.PushUint16(uint16(len(au.Password) + 9))
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
	packer.PushUint16(uint16(len(au.Servername) + 1))
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
		packer.PushUint16(uint16(len(c.Dbpath) + 1))
		packer.PushBytes([]byte(c.Dbpath))
		packer.PushByte(pad)
		packer.PushUint16(uint16(len(c.Dbpathattribute) + 1))
		packer.PushBytes([]byte(c.Dbpathattribute))
		packer.PushByte(pad)
	}
	packer.PushUint16(au.Noname24)
	packer.PushUint32(au.Noname25)
	packer.PushUint32(au.Noname26)
	packer.PushUint32(au.Longthreadid)
	packer.PushUint16(uint16(len(au.Noname27) + 1))
	packer.PushBytes([]byte(au.Noname27))
	packer.PushByte(pad)
	packer.PushUint16(au.Noname28)
	packer.PushUint16(uint16(len(au.Directory) + 1))
	packer.PushBytes([]byte(au.Directory))
	packer.PushByte(pad)
	packer.PushUint16(au.Noname29)
	packer.PushUint16(au.Appnamelengthall)
	packer.PushUint32(au.Noname30)
	packer.PushUint32(au.Noname31)
	packer.PushUint16(uint16(len(au.Appname) + 1))
	packer.PushBytes([]byte(au.Appname))
	packer.PushByte(pad)
	packer.PushUint16(au.Asceot)

	return buffer.Bytes(), packer.Error()
}

func (au *AuthRequest) Unpack(r io.Reader) error {
	var pad byte
	var size uint16
	unpacker := binpacker.NewUnpacker(binary.BigEndian, r)
	unpacker.FetchUint16(&size)
	unpacker.FetchUint8(&au.Noname1)
	unpacker.FetchUint8(&au.Noname2)
	unpacker.FetchUint16(&au.Noname3)

	unpacker.FetchUint16(&au.Noname4)
	unpacker.FetchUint16(&au.Noname5)
	unpacker.FetchUint32(&au.Noname6)
	unpacker.FetchUint16(&size)
	unpacker.FetchString(uint64(size-1), &au.Ieeem)
	unpacker.FetchByte(&pad)
	unpacker.FetchUint16(&au.Noname7)
	unpacker.FetchString(7, &au.Sqlexec)
	for i := 0; i < 5; i++ {
		unpacker.FetchByte(&pad)
	}
	unpacker.FetchUint16(&size)
	unpacker.FetchString(uint64(size-1), &au.Version)
	unpacker.FetchByte(&pad)
	unpacker.FetchUint16(&size)
	unpacker.FetchString(uint64(size-1), &au.Rds)
	unpacker.FetchByte(&pad)
	unpacker.FetchUint16(&size)
	unpacker.FetchString(uint64(size-1), &au.Sqli)
	unpacker.FetchByte(&pad)
	unpacker.FetchUint32(&au.Noname8)
	unpacker.FetchUint32(&au.Noname9)
	unpacker.FetchUint32(&au.Noname10)
	unpacker.FetchUint16(&au.Noname11)
	unpacker.FetchUint16(&size)
	unpacker.FetchString(uint64(size-1), &au.Clientname)
	unpacker.FetchByte(&pad)
	unpacker.FetchUint16(&size)
	for i := 1; i <= 8; i++ {
		unpacker.FetchByte(&pad)
	}
	unpacker.FetchString(uint64(size-9), &au.Password)
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
	unpacker.FetchUint16(&size)
	unpacker.FetchString(uint64(size-1), &au.Servername)
	unpacker.FetchByte(&pad)
	unpacker.FetchUint16(&au.Noname17)
	unpacker.FetchUint16(&au.Noname18)
	unpacker.FetchUint16(&au.Noname19)
	unpacker.FetchUint16(&au.Noname20)
	unpacker.FetchUint16(&au.Noname21)
	unpacker.FetchUint16(&au.Noname22)
	unpacker.FetchUint16(&au.Noname23)
	for i := 0; i < 6; i++ {
		var c DPath
		var sizee uint16
		unpacker.FetchUint16(&sizee)
		unpacker.FetchString(uint64(sizee-1), &c.Dbpath)
		unpacker.FetchByte(&pad)
		unpacker.FetchUint16(&sizee)
		unpacker.FetchString(uint64(sizee-1), &c.Dbpathattribute)
		unpacker.FetchByte(&pad)
		au.Dpath = append(au.Dpath, c)
	}
	unpacker.FetchUint16(&au.Noname24)
	unpacker.FetchUint32(&au.Noname25)
	unpacker.FetchUint32(&au.Noname26)
	unpacker.FetchUint32(&au.Longthreadid)
	unpacker.FetchUint16(&size)
	unpacker.FetchString(uint64(size-1), &au.Noname27)
	unpacker.FetchByte(&pad)
	unpacker.FetchUint16(&au.Noname28)
	unpacker.FetchUint16(&size)
	unpacker.FetchString(uint64(size-1), &au.Directory)
	unpacker.FetchByte(&pad)
	unpacker.FetchUint16(&au.Noname29)
	unpacker.FetchUint16(&au.Appnamelengthall)
	unpacker.FetchUint32(&au.Noname30)
	unpacker.FetchUint32(&au.Noname31)
	unpacker.FetchUint16(&size)
	unpacker.FetchString(uint64(size-1), &au.Appname)
	unpacker.FetchByte(&pad)
	unpacker.FetchUint16(&au.Asceot)

	return unpacker.Error()
}

type DPath struct {
	Dbpath          string
	Dbpathattribute string
}

// response
type AuthResponse struct {
	Noname1    uint8
	Noname2    uint16
	Noname3    uint8
	Noname4    uint16
	Noname5    uint16
	Noname6    uint32
	IEEEI      string
	Noname7    uint16
	Srvinfx    string
	Version    string
	Software   string
	Clientname string
	Noname8    uint32
	Noname9    uint32
	Noname10   uint32
	Noname11   uint16
	Noname12   uint16
	Noname13   uint16
	Noname14   string
	Noname15   string
	Noname16   uint16
	Noname17   uint16
	Noname18   uint16
	Noname19   uint16
	Noname20   uint16
	Noname21   uint16
	Noname22   uint16
	Noname23   uint16
	Noname24   uint16
	Path1      string
	Path2      string
	Noname25   uint16
	Noname26   uint16
	Noname27   uint16
	Noname28   uint16
	Noname29   uint16
	Noname30   uint16
	Noname31   uint16
	Noname32   uint16
	Noname33   uint16
	Noname34   uint16
	Path3      string
	Asceot     uint16
}

func (au *AuthResponse) Pack() ([]byte, error) {
	var pad byte
	buffer := new(bytes.Buffer)
	packer := binpacker.NewPacker(binary.BigEndian, buffer)
	packer.PushUint16(uint16(len(au.IEEEI) + len(au.Srvinfx) + len(au.Version) + len(au.Software) +
		len(au.Clientname) + len(au.Noname14) + len(au.Noname15) + len(au.Path1) + len(au.Path2) +
		len(au.Path3) + 132))
	packer.PushByte(au.Noname1)
	packer.PushUint16(au.Noname2)
	packer.PushByte(au.Noname3)
	packer.PushUint16(au.Noname4)
	packer.PushUint16(au.Noname5)
	packer.PushUint32(au.Noname6)
	packer.PushUint16(uint16(len(au.IEEEI) + 1))
	packer.PushBytes([]byte(au.IEEEI))
	packer.PushByte(pad)
	packer.PushUint16(au.Noname7)
	packer.PushBytes([]byte(au.Srvinfx))
	for i := 0; i < 5; i++ {
		packer.PushByte(pad)
	}
	packer.PushUint16(uint16(len(au.Version) + 1))
	packer.PushBytes([]byte(au.Version))
	packer.PushByte(pad)
	packer.PushUint16(uint16(len(au.Software) + 1))
	packer.PushBytes([]byte(au.Software))
	packer.PushByte(pad)
	packer.PushUint16(uint16(len(au.Clientname) + 1))
	packer.PushBytes([]byte(au.Clientname))
	packer.PushByte(pad)
	packer.PushUint32(au.Noname8)
	packer.PushUint32(au.Noname9)
	packer.PushUint32(au.Noname10)
	packer.PushUint16(au.Noname11)
	packer.PushUint16(au.Noname12)
	packer.PushUint16(au.Noname13)
	packer.PushBytes([]byte(au.Noname14))
	for i := 0; i < 9; i++ {
		packer.PushByte(pad)
	}
	packer.PushBytes([]byte(au.Noname15))
	for i := 0; i < 6; i++ {
		packer.PushByte(pad)
	}
	packer.PushUint16(au.Noname16)
	for i := 0; i < 6; i++ {
		packer.PushByte(pad)
	}
	packer.PushUint16(au.Noname17)
	packer.PushUint16(au.Noname18)
	packer.PushUint16(au.Noname19)
	packer.PushUint16(au.Noname20)
	packer.PushUint16(au.Noname21)
	for i := 0; i < 6; i++ {
		packer.PushByte(pad)
	}
	packer.PushUint16(au.Noname22)
	for i := 0; i < 5; i++ {
		packer.PushByte(pad)
	}
	packer.PushUint16(au.Noname23)
	packer.PushUint16(au.Noname24)
	packer.PushUint16(uint16(len(au.Path1) + 1))
	packer.PushBytes([]byte(au.Path1))
	packer.PushByte(pad)
	packer.PushUint16(uint16(len(au.Path2) + 1))
	packer.PushBytes([]byte(au.Path2))
	packer.PushByte(pad)
	packer.PushUint16(au.Noname25)
	packer.PushUint16(au.Noname26)
	packer.PushUint16(au.Noname27)
	packer.PushUint16(au.Noname28)
	packer.PushUint16(au.Noname29)
	packer.PushUint16(au.Noname30)
	packer.PushUint16(au.Noname31)
	packer.PushUint16(au.Noname32)
	packer.PushUint16(au.Noname33)
	packer.PushUint16(au.Noname34)
	packer.PushUint16(uint16(len(au.Path3) + 1))
	packer.PushBytes([]byte(au.Path3))
	packer.PushByte(pad)
	packer.PushUint16(au.Asceot)

	return buffer.Bytes(), packer.Error()
}

func (au *AuthResponse) Unpack(r io.Reader) error {
	var pad byte
	var size uint16
	unpacker := binpacker.NewUnpacker(binary.BigEndian, r)
	unpacker.FetchUint16(&size)
	unpacker.FetchByte(&au.Noname1)
	unpacker.FetchUint16(&au.Noname2)
	unpacker.FetchByte(&au.Noname3)
	unpacker.FetchUint16(&au.Noname4)
	unpacker.FetchUint16(&au.Noname5)
	unpacker.FetchUint32(&au.Noname6)
	unpacker.FetchUint16(&size)
	unpacker.FetchString(uint64(size-1), &au.IEEEI)
	unpacker.FetchByte(&pad)
	unpacker.FetchUint16(&au.Noname7)
	unpacker.FetchString(7, &au.Srvinfx)
	for i := 0; i < 5; i++ {
		unpacker.FetchByte(&pad)
	}
	unpacker.FetchUint16(&size)
	unpacker.FetchString(uint64(size-1), &au.Version)
	unpacker.FetchByte(&pad)
	unpacker.FetchUint16(&size)
	unpacker.FetchString(uint64(size-1), &au.Software)
	unpacker.FetchByte(&pad)
	unpacker.FetchUint16(&size)
	unpacker.FetchString(uint64(size-1), &au.Clientname)
	unpacker.FetchByte(&pad)
	unpacker.FetchUint32(&au.Noname8)
	unpacker.FetchUint32(&au.Noname9)
	unpacker.FetchUint32(&au.Noname10)
	unpacker.FetchUint16(&au.Noname11)
	unpacker.FetchUint16(&au.Noname12)
	unpacker.FetchUint16(&au.Noname13)
	unpacker.FetchString(2, &au.Noname14)
	for i := 0; i < 9; i++ {
		unpacker.FetchByte(&pad)
	}
	unpacker.FetchString(7, &au.Noname15)
	for i := 0; i < 6; i++ {
		unpacker.FetchByte(&pad)
	}
	unpacker.FetchUint16(&au.Noname16)
	for i := 0; i < 6; i++ {
		unpacker.FetchByte(&pad)
	}
	unpacker.FetchUint16(&au.Noname17)
	unpacker.FetchUint16(&au.Noname18)
	unpacker.FetchUint16(&au.Noname19)
	unpacker.FetchUint16(&au.Noname20)
	unpacker.FetchUint16(&au.Noname21)
	for i := 0; i < 6; i++ {
		unpacker.FetchByte(&pad)
	}
	unpacker.FetchUint16(&au.Noname22)
	for i := 0; i < 5; i++ {
		unpacker.FetchByte(&pad)
	}
	unpacker.FetchUint16(&au.Noname23)
	unpacker.FetchUint16(&au.Noname24)
	unpacker.FetchUint16(&size)
	unpacker.FetchString(uint64(size-1), &au.Path1)
	unpacker.FetchByte(&pad)
	unpacker.FetchUint16(&size)
	unpacker.FetchString(uint64(size-1), &au.Path2)
	unpacker.FetchByte(&pad)
	unpacker.FetchUint16(&au.Noname25)
	unpacker.FetchUint16(&au.Noname26)
	unpacker.FetchUint16(&au.Noname27)
	unpacker.FetchUint16(&au.Noname28)
	unpacker.FetchUint16(&au.Noname29)
	unpacker.FetchUint16(&au.Noname30)
	unpacker.FetchUint16(&au.Noname31)
	unpacker.FetchUint16(&au.Noname32)
	unpacker.FetchUint16(&au.Noname33)
	unpacker.FetchUint16(&au.Noname34)
	unpacker.FetchUint16(&size)
	unpacker.FetchString(uint64(size-1), &au.Path3)
	unpacker.FetchByte(&pad)
	unpacker.FetchUint16(&au.Asceot)

	return unpacker.Error()
}
