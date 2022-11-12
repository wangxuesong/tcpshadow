package services

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/davecgh/go-spew/spew"
	"github.com/wangxuesong/tcpshadow/model"
)

type OutputType string

const (
	Console OutputType = "console"
	File    OutputType = "file"
)

const (
	WHITE  = "\033[37m"
	CYAN   = "\033[36m"
	PURPLE = "\033[95m"
	BLUE   = "\033[94m"
	GREEN  = "\033[92m"
	YELLOW = "\033[93m"
	RED    = "\033[91m"
	CLEAR  = "\033[0m"
)

type (
	OutputService struct {
		monitor chan *Context
		outputs []Output

		wg   sync.WaitGroup
		done chan struct{}
	}

	OutputConfig struct {
		Monitor  chan *Context
		Outputs  []OutputType
		Filename string
	}

	Output interface {
		Write(context *Context) error
		Close(wg *sync.WaitGroup)
	}

	ConsoleOutput struct {
	}

	FileOutput struct {
		filename string
		file     *os.File
		index    uint16
	}
)

func NewOutputService(config OutputConfig) *OutputService {
	s := &OutputService{
		monitor: config.Monitor,
		done:    make(chan struct{}),
	}

	for _, o := range config.Outputs {
		switch o {
		case Console:
			s.outputs = append(s.outputs, &ConsoleOutput{})
		case File:
			s.outputs = append(s.outputs, newFileOutput(config.Filename))
		}
	}

	return s
}

func newFileOutput(filename string) *FileOutput {
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil
	}
	file.Truncate(0)
	return &FileOutput{
		filename: filename,
		file:     file,
		index:    0,
	}
}

func (s *OutputService) Run() {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		for {
			select {
			case <-s.done:
				log.Println("stopping output services")
				return
			case ctx := <-s.monitor:
				for _, o := range s.outputs {
					o.Write(ctx)
				}
			}
		}
	}()
}

func (s *OutputService) Close(wg *sync.WaitGroup) {
	defer wg.Done()
	close(s.done)
	for _, o := range s.outputs {
		s.wg.Add(1)
		o.Close(&s.wg)
	}
	s.wg.Wait()
}

func (o *ConsoleOutput) Write(ctx *Context) error {
	if ctx.Data.Forward == model.ServerToClient {
		o.dump(YELLOW, ctx)
	} else {
		o.dump(GREEN, ctx)
	}
	return nil
}

func (o *ConsoleOutput) dump(color string, ctx *Context) {
	scs := spew.NewDefaultConfig()
	scs.Indent = "    "
	_, _ = scs.Print(color)
	str := fmt.Sprintf("[%d] %s", ctx.SessionId, scs.Sdump(*ctx.Data))
	_, _ = scs.Print(str)
	_, _ = scs.Println(CLEAR)
}

func (o *ConsoleOutput) Close(wg *sync.WaitGroup) {
	defer wg.Done()
}

func (o *FileOutput) Write(ctx *Context) error {
	var header struct {
		Index   uint16
		Forward uint8
		Length  int64
	}
	header.Index = uint16(o.index)
	header.Forward = uint8(ctx.Data.Forward)
	header.Length = int64(len(ctx.Data.Buffer))
	buf := new(bytes.Buffer)
	_ = binary.Write(buf, binary.LittleEndian, header)
	_, _ = o.file.Write(buf.Bytes())
	_, _ = o.file.Write(ctx.Data.Buffer)
	_ = o.file.Sync()
	o.index++
	return nil
}

func (o *FileOutput) Close(wg *sync.WaitGroup) {
	defer wg.Done()
	err := o.file.Close()
	if err != nil {
		panic(err)
	}
}
