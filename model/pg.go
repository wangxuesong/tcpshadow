package model

import (
	"bytes"
	"encoding/binary"
	pgproto "github.com/jackc/pgproto3"
	"io"
)

//type PgCommand interface {
//	Command() uint16
//	Pack() ([]byte, error)
//	Unpack(r io.Reader) error
//}

type PgCommand pgproto.Message

type PgTransmission []PgCommand

func (t *PgTransmission) Pack() ([]byte, error) {
	temp := new(bytes.Buffer)
	for _, cmd := range *t {
		buf := cmd.Encode(nil)
		temp.Write(buf)
	}
	buffer := new(bytes.Buffer)
	buffer.Write(temp.Bytes())
	return buffer.Bytes(), nil
}

func (t *PgTransmission) Append(cmd PgCommand) {
	*t = append(*t, cmd)
}

func UnpackPgCommand(reader io.ReadSeeker) (PgCommand, error) {
	var cmd byte
	pos, err := reader.Seek(0, io.SeekCurrent)
	if err != nil {
		return nil, err
	}
	err = binary.Read(reader, binary.BigEndian, &cmd)
	if err != nil {
		return nil, err
	}
	_, _ = reader.Seek(pos, io.SeekStart)

	switch cmd {
	//case 1:
	//	command := &SqliCmd{}
	//	err = command.Unpack(reader)
	//	if err != nil {
	//		reader.Seek(pos, io.SeekStart)
	//		return nil, err
	//	}
	//	return command, nil
	//case 2:
	//	command := &SqliPrepare{}
	//	err = command.Unpack(reader)
	//	if err != nil {
	//		reader.Seek(pos, io.SeekStart)
	//		return nil, err
	//	}
	//	return command, nil
	//case 3:
	//	command := &SqliCurName{}
	//	err = command.Unpack(reader)
	//	if err != nil {
	//		reader.Seek(pos, io.SeekStart)
	//		return nil, err
	//	}
	//	return command, nil
	//case 4:
	//	command := &SqliID{}
	//	err = command.Unpack(reader)
	//	if err != nil {
	//		reader.Seek(pos, io.SeekStart)
	//		return nil, err
	//	}
	//	return command, nil
	//case 5:
	//	command := &SqliBind{}
	//	err = command.Unpack(reader)
	//	if err != nil {
	//		reader.Seek(pos, io.SeekStart)
	//		return nil, err
	//	}
	//	return command, nil
	//case 6:
	//	command := &SqliOpen{}
	//	return command, nil
	//case 7:
	//	command := &SqliExecute{}
	//	return command, nil
	//case 8:
	//	command := &SqliDescribe{}
	//	err = command.Unpack(reader)
	//	if err != nil {
	//		reader.Seek(pos, io.SeekStart)
	//		return nil, err
	//	}
	//	return command, nil
	//case 9:
	//	command := &SqliNFetch{}
	//	err = command.Unpack(reader)
	//	if err != nil {
	//		reader.Seek(pos, io.SeekStart)
	//		return nil, err
	//	}
	//	return command, nil
	//case 10:
	//	command := &SqliClose{}
	//	return command, nil
	//case 11:
	//	command := &SqliRelease{}
	//	return command, nil
	//case 12:
	//	command := &SqliEot{}
	//	return command, nil
	//case 13:
	//	command := &SqliErr{}
	//	err = command.Unpack(reader)
	//	if err != nil {
	//		reader.Seek(pos, io.SeekStart)
	//		return nil, err
	//	}
	//	return command, nil
	//case 14:
	//	command := &SqliTuple{}
	//	err = command.Unpack(reader)
	//	if err != nil {
	//		reader.Seek(pos, io.SeekStart)
	//		return nil, err
	//	}
	//	return command, nil
	//case 15:
	//	command := &SqliDone{}
	//	err = command.Unpack(reader)
	//	if err != nil {
	//		reader.Seek(pos, io.SeekStart)
	//		return nil, err
	//	}
	//	return command, nil
	//case 22:
	//	command := &SqliNDescribe{}
	//	return command, nil
	//case 36:
	//	command := &SqliDBOpen{}
	//	err = command.Unpack(reader)
	//	if err != nil {
	//		reader.Seek(pos, io.SeekStart)
	//		return nil, err
	//	}
	//	return command, nil
	//case 49:
	//	command := &SqliWantDone{}
	//	return command, nil
	//case 56:
	//	command := &SqliExit{}
	//	return command, nil
	//case 81:
	//	command := &SqliInfo{}
	//	err = command.Unpack(reader)
	//	if err != nil {
	//		reader.Seek(pos, io.SeekStart)
	//		return nil, err
	//	}
	//	return command, nil
	//case 82:
	//	command, err := pgproto.ParseAuthenticationRequest(reader)
	//	if err != nil {
	//		_, _ = reader.Seek(pos, io.SeekStart)
	//		return nil, err
	//	}
	//	return command, nil
	//case 94:
	//	command := &SqliInsertDone{}
	//	err = command.Unpack(reader)
	//	if err != nil {
	//		reader.Seek(pos, io.SeekStart)
	//		return nil, err
	//	}
	//	return command, nil
	//case 100:
	//	command := &SqliRetType{}
	//	err = command.Unpack(reader)
	//	if err != nil {
	//		reader.Seek(pos, io.SeekStart)
	//		return nil, err
	//	}
	//	return command, nil
	//case 108:
	//	command := &SqliAutoFree{}
	//	return command, nil
	//case 126:
	//	command := &SqliProtocols{}
	//	err = command.Unpack(reader)
	//	if err != nil {
	//		reader.Seek(pos, io.SeekStart)
	//		return nil, err
	//	}
	//	return command, nil
	default:
		return nil, UnknownCommandError(uint16(cmd))
	}
}

func ParseServerMessage(reader io.Reader) (PgTransmission, error) {
	backend, err := pgproto.NewFrontend(pgproto.NewChunkReader(reader), nil)
	if err != nil {
		return nil, err
	}
	trans := make(PgTransmission, 0, 5)
	for {
		cmd, err := backend.Receive()
		if err == io.EOF {
			return trans, nil
		}
		if err != nil {
			return nil, err
		}
		trans.Append(cmd)
	}
}

func ParseClientMessage(reader io.Reader) (PgTransmission, error) {
	frontend, err := pgproto.NewBackend(pgproto.NewChunkReader(reader), nil)
	if err != nil {
		return nil, err
	}
	trans := make(PgTransmission, 0, 5)
	for {
		cmd, err := frontend.Receive()
		if err == io.EOF {
			return trans, nil
		}
		if err != nil {
			return nil, err
		}
		trans.Append(cmd)
	}
}

func UnpackPgTransmission(reader io.ReadSeeker) (PgTransmission, error) {
	trans := make(PgTransmission, 0, 5)
	for {
		cmd, err := UnpackPgCommand(reader)
		if err == io.EOF {
			return trans, nil
		}
		if err != nil {
			return nil, err
		}
		trans.Append(cmd)
	}
}
