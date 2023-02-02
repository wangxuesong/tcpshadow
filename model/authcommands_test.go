package model

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAuthRequest_Pack(t *testing.T) {
	authrequest := &AuthRequest{
		Noname1:  1,
		Noname2:  60,
		Noname3:  0,
		Noname4:  100,
		Noname5:  101,
		Noname6:  61,
		Ieeem:    "IEEEM",
		Noname7:  108,
		Sqlexec:  "sqlexec",
		Version:  "9.280",
		Rds:      "RDS#R000000",
		Sqli:     "sqli",
		Noname8:  316,
		Noname9:  0,
		Noname10: 0,
		Noname11: 1,
		//Clientname: "gbasedbt",
		//Password:   "HmQOYC1ZfTYt+vlXUhkn3w==",
		Noname12: "ol",
		Noname13: 61,
		Tlitcp:   "tlitcp",
		Noname14: 1,
		Noname15: 104,
		Asf:      11,
		Noname16: 3,
		//Servername: "gbaseserver",
		Noname17: 0,
		Noname18: 0,
		Noname19: 0,
		Noname20: 0,
		Noname21: 0,
		Noname22: 106,
		Noname23: 6,
		Dpath: []DPath{{
			Dbpath:          "DBPATH",
			Dbpathattribute: ".",
		}, {
			Dbpath:          "CLNT_PAM_CAPABLE",
			Dbpathattribute: "1",
		}, {
			Dbpath:          "DBDATE",
			Dbpathattribute: "Y4MD-",
		}, {
			Dbpath:          "IFX_UPDDESC",
			Dbpathattribute: "1",
		}, {
			Dbpath:          "SQLMODE",
			Dbpathattribute: "gbase",
		}, {
			Dbpath:          "NODEFDAC",
			Dbpathattribute: "no",
		}},
		Noname24:         107,
		Noname25:         0,
		Noname26:         0,
		Longthreadid:     1,
		Noname27:         "MM-202201031507",
		Noname28:         0,
		Directory:        "E:\\JDBCTest\\JDBCTest",
		Noname29:         116,
		Appnamelengthall: 80,
		Noname30:         0,
		Noname31:         0,
		Appname:          "/E:/JDBCTest/JDBCTest/lib/gbasedbtjdbc_3.3.0_2.jarConnectionTest/Test",
		Asceot:           127,
	}

	tRequestBuilder := RequestBuilder{
		Aauthrequest: *authrequest,
	}
	authrequest = tRequestBuilder.BuildClientname("gbasedbt").
		BuildPassword("HmQOYC1ZfTYt+vlXUhkn3w==").
		BuildGbaseS("gbaseserver", 80).Create()

	pack, err := authrequest.Pack()
	assert.Nil(t, err)
	buffer := []byte{1, 177, 1, 60, 0, 0, 0, 100, 0, 101, 0, 0, 0, 61, 0, 6, 73, 69, 69, 69, 77, 0, 0, 108, 115, 113, 108, 101, 120, 101, 99, 0, 0, 0, 0, 0, 0, 6, 57, 46, 50, 56, 48, 0, 0, 12, 82, 68, 83, 35, 82,
		48, 48, 48, 48, 48, 48, 0, 0, 5, 115, 113, 108, 105, 0, 0, 0, 1, 60, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 9, 103, 98, 97, 115, 101, 100, 98, 116, 0, 0, 33, 1, 1, 1, 1, 1, 1, 1, 1, 72, 109, 81, 79, 89, 67, 49, 90,
		102, 84, 89, 116, 43, 118, 108, 88, 85, 104, 107, 110, 51, 119, 61, 61, 0, 111, 108, 0, 0, 0, 0, 0, 0, 0, 0, 0, 61, 116, 108, 105, 116, 99, 112, 0, 0, 0, 0, 0, 1, 0, 104, 0, 11, 0, 0, 0, 3, 0, 12, 103,
		98, 97, 115, 101, 115, 101, 114, 118, 101, 114, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 106, 0, 6, 0, 7, 68, 66, 80, 65, 84, 72, 0, 0, 2, 46, 0, 0, 17, 67, 76, 78, 84, 95, 80, 65, 77, 95, 67, 65, 80, 65, 66, 76,
		69, 0, 0, 2, 49, 0, 0, 7, 68, 66, 68, 65, 84, 69, 0, 0, 6, 89, 52, 77, 68, 45, 0, 0, 12, 73, 70, 88, 95, 85, 80, 68, 68, 69, 83, 67, 0, 0, 2, 49, 0, 0, 8, 83, 81, 76, 77, 79, 68, 69, 0, 0, 6, 103, 98, 97,
		115, 101, 0, 0, 9, 78, 79, 68, 69, 70, 68, 65, 67, 0, 0, 3, 110, 111, 0, 0, 107, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 16, 77, 77, 45, 50, 48, 50, 50, 48, 49, 48, 51, 49, 53, 48, 55, 0, 0, 0, 0, 21, 69, 58, 92,
		74, 68, 66, 67, 84, 101, 115, 116, 92, 74, 68, 66, 67, 84, 101, 115, 116, 0, 0, 116, 0, 80, 0, 0, 0, 0, 0, 0, 0, 0, 0, 70, 47, 69, 58, 47, 74, 68, 66, 67, 84, 101, 115, 116, 47, 74, 68, 66, 67, 84, 101,
		115, 116, 47, 108, 105, 98, 47, 103, 98, 97, 115, 101, 100, 98, 116, 106, 100, 98, 99, 95, 51, 46, 51, 46, 48, 95, 50, 46, 106, 97, 114, 67, 111, 110, 110, 101, 99, 116, 105, 111, 110, 84, 101, 115, 116,
		47, 84, 101, 115, 116, 0, 0, 127}
	assert.Equal(t, buffer, pack)
}

func TestAuthRequest_Unpack(t *testing.T) {
	buffer := []byte{1, 177, 1, 60, 0, 0, 0, 100, 0, 101, 0, 0, 0, 61, 0, 6, 73, 69, 69, 69, 77, 0, 0, 108, 115, 113, 108, 101, 120, 101, 99, 0, 0, 0, 0, 0, 0, 6, 57, 46, 50, 56, 48, 0, 0, 12, 82, 68, 83, 35, 82,
		48, 48, 48, 48, 48, 48, 0, 0, 5, 115, 113, 108, 105, 0, 0, 0, 1, 60, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 9, 103, 98, 97, 115, 101, 100, 98, 116, 0, 0, 33, 1, 1, 1, 1, 1, 1, 1, 1, 72, 109, 81, 79, 89, 67, 49, 90,
		102, 84, 89, 116, 43, 118, 108, 88, 85, 104, 107, 110, 51, 119, 61, 61, 0, 111, 108, 0, 0, 0, 0, 0, 0, 0, 0, 0, 61, 116, 108, 105, 116, 99, 112, 0, 0, 0, 0, 0, 1, 0, 104, 0, 11, 0, 0, 0, 3, 0, 12, 103,
		98, 97, 115, 101, 115, 101, 114, 118, 101, 114, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 106, 0, 6, 0, 7, 68, 66, 80, 65, 84, 72, 0, 0, 2, 46, 0, 0, 17, 67, 76, 78, 84, 95, 80, 65, 77, 95, 67, 65, 80, 65, 66, 76,
		69, 0, 0, 2, 49, 0, 0, 7, 68, 66, 68, 65, 84, 69, 0, 0, 6, 89, 52, 77, 68, 45, 0, 0, 12, 73, 70, 88, 95, 85, 80, 68, 68, 69, 83, 67, 0, 0, 2, 49, 0, 0, 8, 83, 81, 76, 77, 79, 68, 69, 0, 0, 6, 103, 98, 97,
		115, 101, 0, 0, 9, 78, 79, 68, 69, 70, 68, 65, 67, 0, 0, 3, 110, 111, 0, 0, 107, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 16, 77, 77, 45, 50, 48, 50, 50, 48, 49, 48, 51, 49, 53, 48, 55, 0, 0, 0, 0, 21, 69, 58, 92,
		74, 68, 66, 67, 84, 101, 115, 116, 92, 74, 68, 66, 67, 84, 101, 115, 116, 0, 0, 116, 0, 80, 0, 0, 0, 0, 0, 0, 0, 0, 0, 70, 47, 69, 58, 47, 74, 68, 66, 67, 84, 101, 115, 116, 47, 74, 68, 66, 67, 84, 101,
		115, 116, 47, 108, 105, 98, 47, 103, 98, 97, 115, 101, 100, 98, 116, 106, 100, 98, 99, 95, 51, 46, 51, 46, 48, 95, 50, 46, 106, 97, 114, 67, 111, 110, 110, 101, 99, 116, 105, 111, 110, 84, 101, 115, 116,
		47, 84, 101, 115, 116, 0, 0, 127}
	reader := bytes.NewReader(buffer)
	authrequest := &AuthRequest{}
	err := authrequest.Unpack(reader)
	assert.Nil(t, err)
	expect := &AuthRequest{
		Noname1:    1,
		Noname2:    60,
		Noname3:    0,
		Noname4:    100,
		Noname5:    101,
		Noname6:    61,
		Ieeem:      "IEEEM",
		Noname7:    108,
		Sqlexec:    "sqlexec",
		Version:    "9.280",
		Rds:        "RDS#R000000",
		Sqli:       "sqli",
		Noname8:    316,
		Noname9:    0,
		Noname10:   0,
		Noname11:   1,
		Clientname: "gbasedbt",
		Password:   "HmQOYC1ZfTYt+vlXUhkn3w==",
		Noname12:   "ol",
		Noname13:   61,
		Tlitcp:     "tlitcp",
		Noname14:   1,
		Noname15:   104,
		Asf:        11,
		Noname16:   3,
		Servername: "gbaseserver",
		Noname17:   0,
		Noname18:   0,
		Noname19:   0,
		Noname20:   0,
		Noname21:   0,
		Noname22:   106,
		Noname23:   6,
		Dpath: []DPath{{
			Dbpath:          "DBPATH",
			Dbpathattribute: ".",
		}, {
			Dbpath:          "CLNT_PAM_CAPABLE",
			Dbpathattribute: "1",
		}, {
			Dbpath:          "DBDATE",
			Dbpathattribute: "Y4MD-",
		}, {
			Dbpath:          "IFX_UPDDESC",
			Dbpathattribute: "1",
		}, {
			Dbpath:          "SQLMODE",
			Dbpathattribute: "gbase",
		}, {
			Dbpath:          "NODEFDAC",
			Dbpathattribute: "no",
		}},
		Noname24:         107,
		Noname25:         0,
		Noname26:         0,
		Longthreadid:     1,
		Noname27:         "MM-202201031507",
		Noname28:         0,
		Directory:        "E:\\JDBCTest\\JDBCTest",
		Noname29:         116,
		Appnamelengthall: 80,
		Noname30:         0,
		Noname31:         0,
		Appname:          "/E:/JDBCTest/JDBCTest/lib/gbasedbtjdbc_3.3.0_2.jarConnectionTest/Test",
		Asceot:           127,
	}
	assert.Equal(t, authrequest, expect)
}

func TestAuthResponse_Pack(t *testing.T) {
	authresponse := &AuthResponse{
		Noname1:    2,
		Noname2:    15376,
		Noname3:    0,
		Noname4:    100,
		Noname5:    101,
		Noname6:    61,
		IEEEI:      "IEEEI",
		Noname7:    108,
		Srvinfx:    "srvinfx",
		Version:    "GBase Server Version 9.56.FC4G1TL",
		Software:   "Software Serial Number AAA#B000000",
		Clientname: "gbaseserver",
		Noname8:    316,
		Noname9:    0,
		Noname10:   0,
		Noname11:   0,
		Noname12:   0,
		Noname13:   0,
		Noname14:   "on",
		Noname15:   "=soctcp",
		Noname16:   102,
		Noname17:   0,
		Noname18:   0,
		Noname19:   20,
		Noname20:   0,
		Noname21:   107,
		Noname22:   2958,
		Noname23:   872,
		Noname24:   13312,
		Path1:      "/dev/pts/0",
		Path2:      "/home/gbasedbt",
		Noname25:   110,
		Noname26:   4,
		Noname27:   0,
		Noname28:   0,
		Noname29:   116,
		Noname30:   43,
		Noname31:   0,
		Noname32:   1001,
		Noname33:   0,
		Noname34:   1001,
		Path3:      "/home/zhangyaru/gbase/bin/oninit",
		Asceot:     127,
	}
	pack, err := authresponse.Pack()
	assert.Nil(t, err)
	buffer := []byte{1, 31, 2, 60, 16, 0, 0, 100, 0, 101, 0, 0, 0, 61, 0, 6, 73, 69, 69, 69, 73, 0, 0, 108, 115, 114, 118, 105, 110, 102, 120, 0, 0, 0, 0, 0, 0, 34, 71, 66, 97, 115, 101, 32, 83, 101, 114, 118,
		101, 114, 32, 86, 101, 114, 115, 105, 111, 110, 32, 57, 46, 53, 54, 46, 70, 67, 52, 71, 49, 84, 76, 0, 0, 35, 83, 111, 102, 116, 119, 97, 114, 101, 32, 83, 101, 114, 105, 97, 108, 32, 78, 117, 109,
		98, 101, 114, 32, 65, 65, 65, 35, 66, 48, 48, 48, 48, 48, 48, 0, 0, 12, 103, 98, 97, 115, 101, 115, 101, 114, 118, 101, 114, 0, 0, 0, 1, 60, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 111, 110, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 61, 115, 111, 99, 116, 99, 112, 0, 0, 0, 0, 0, 0, 0, 102, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 20, 0, 0, 0, 107, 0, 0, 0, 0, 0, 0, 11, 142, 0, 0, 0, 0, 0, 3, 104, 52, 0, 0, 11, 47, 100, 101, 118, 47,
		112, 116, 115, 47, 48, 0, 0, 15, 47, 104, 111, 109, 101, 47, 103, 98, 97, 115, 101, 100, 98, 116, 0, 0, 110, 0, 4, 0, 0, 0, 0, 0, 116, 0, 43, 0, 0, 3, 233, 0, 0, 3, 233, 0, 33, 47, 104, 111, 109, 101,
		47, 122, 104, 97, 110, 103, 121, 97, 114, 117, 47, 103, 98, 97, 115, 101, 47, 98, 105, 110, 47, 111, 110, 105, 110, 105, 116, 0, 0, 127}
	assert.Equal(t, buffer, pack)
}

func TestAuthResponse_Unpack(t *testing.T) {
	buffer := []byte{1, 31, 2, 60, 16, 0, 0, 100, 0, 101, 0, 0, 0, 61, 0, 6, 73, 69, 69, 69, 73, 0, 0, 108, 115, 114, 118, 105, 110, 102, 120, 0, 0, 0, 0, 0, 0, 34, 71, 66, 97, 115, 101, 32, 83, 101, 114, 118,
		101, 114, 32, 86, 101, 114, 115, 105, 111, 110, 32, 57, 46, 53, 54, 46, 70, 67, 52, 71, 49, 84, 76, 0, 0, 35, 83, 111, 102, 116, 119, 97, 114, 101, 32, 83, 101, 114, 105, 97, 108, 32, 78, 117, 109,
		98, 101, 114, 32, 65, 65, 65, 35, 66, 48, 48, 48, 48, 48, 48, 0, 0, 12, 103, 98, 97, 115, 101, 115, 101, 114, 118, 101, 114, 0, 0, 0, 1, 60, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 111, 110, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 61, 115, 111, 99, 116, 99, 112, 0, 0, 0, 0, 0, 0, 0, 102, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 20, 0, 0, 0, 107, 0, 0, 0, 0, 0, 0, 11, 142, 0, 0, 0, 0, 0, 3, 104, 52, 0, 0, 11, 47, 100, 101, 118, 47,
		112, 116, 115, 47, 48, 0, 0, 15, 47, 104, 111, 109, 101, 47, 103, 98, 97, 115, 101, 100, 98, 116, 0, 0, 110, 0, 4, 0, 0, 0, 0, 0, 116, 0, 43, 0, 0, 3, 233, 0, 0, 3, 233, 0, 33, 47, 104, 111, 109, 101,
		47, 122, 104, 97, 110, 103, 121, 97, 114, 117, 47, 103, 98, 97, 115, 101, 47, 98, 105, 110, 47, 111, 110, 105, 110, 105, 116, 0, 0, 127}
	reader := bytes.NewReader(buffer)
	authresponse := &AuthResponse{}
	err := authresponse.Unpack(reader)
	assert.Nil(t, err)
	expect := &AuthResponse{
		Noname1:    2,
		Noname2:    15376,
		Noname3:    0,
		Noname4:    100,
		Noname5:    101,
		Noname6:    61,
		IEEEI:      "IEEEI",
		Noname7:    108,
		Srvinfx:    "srvinfx",
		Version:    "GBase Server Version 9.56.FC4G1TL",
		Software:   "Software Serial Number AAA#B000000",
		Clientname: "gbaseserver",
		Noname8:    316,
		Noname9:    0,
		Noname10:   0,
		Noname11:   0,
		Noname12:   0,
		Noname13:   0,
		Noname14:   "on",
		Noname15:   "=soctcp",
		Noname16:   102,
		Noname17:   0,
		Noname18:   0,
		Noname19:   20,
		Noname20:   0,
		Noname21:   107,
		Noname22:   2958,
		Noname23:   872,
		Noname24:   13312,
		Path1:      "/dev/pts/0",
		Path2:      "/home/gbasedbt",
		Noname25:   110,
		Noname26:   4,
		Noname27:   0,
		Noname28:   0,
		Noname29:   116,
		Noname30:   43,
		Noname31:   0,
		Noname32:   1001,
		Noname33:   0,
		Noname34:   1001,
		Path3:      "/home/zhangyaru/gbase/bin/oninit",
		Asceot:     127,
	}
	assert.Equal(t, authresponse, expect)
}
