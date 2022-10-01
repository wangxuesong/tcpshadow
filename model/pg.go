package model

import (
	"bytes"
	pgproto "github.com/jackc/pgproto3"
	"io"
)

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

type PgServerParser struct {
	parser *pgproto.Frontend
	buffer *bytes.Buffer
}

func NewPgServerParser() *PgServerParser {
	buffer := bytes.NewBuffer(nil)
	frontend, _ := pgproto.NewFrontend(pgproto.NewChunkReader(buffer), nil)
	return &PgServerParser{
		parser: frontend,
		buffer: buffer,
	}
}

func (s *PgServerParser) Append(data []byte) (int, error) {
	return s.buffer.Write(data)
}

func (s *PgServerParser) ParseMessage() (PgTransmission, error) {
	trans := make(PgTransmission, 0, 5)
	for {
		cmd, err := s.parser.Receive()
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			return trans, nil
		}
		if err != nil {
			return nil, err
		}
		trans.Append(cmd)
	}
}

type PgClientParser struct {
	parser *pgproto.Backend
	buffer *bytes.Buffer
}

func NewPgClientParser() *PgClientParser {
	buffer := bytes.NewBuffer(nil)
	backend, _ := pgproto.NewBackend(pgproto.NewChunkReader(buffer), nil)
	return &PgClientParser{
		parser: backend,
		buffer: buffer,
	}
}

func (s *PgClientParser) Append(data []byte) (int, error) {
	return s.buffer.Write(data)
}

func (s *PgClientParser) ParseMessage() (PgTransmission, error) {
	trans := make(PgTransmission, 0, 5)
	for {
		cmd, err := s.parser.Receive()
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			return trans, nil
		}
		if err != nil {
			return nil, err
		}
		trans.Append(cmd)
	}
}
