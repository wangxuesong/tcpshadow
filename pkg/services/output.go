package services

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/davecgh/go-spew/spew"
	"github.com/jackc/pgproto3"
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
		monitor chan Context
		outputs []Output

		wg   sync.WaitGroup
		done chan struct{}
	}

	OutputConfig struct {
		Monitor        chan Context
		Outputs        []OutputType
		Filename       string
		ProtocolType   string
		IsPrintPackage bool
	}

	Output interface {
		Write(context Context) error
		Close(wg *sync.WaitGroup)
	}

	ConsoleOutput struct {
	}

	sqliConsoleOutput struct {
		ConsoleOutput
		counts map[int]int
	}

	pgConsoleOutput struct {
		ConsoleOutput
		counts map[int]int

		serverParser *model.PgServerParser
		clientParser *model.PgClientParser
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
			var output Output = &ConsoleOutput{}
			if config.IsPrintPackage {
				switch config.ProtocolType {
				case "sqli":
					output = &sqliConsoleOutput{
						counts: make(map[int]int),
					}
					break
				case "pg":
					output = &pgConsoleOutput{
						counts:       make(map[int]int),
						serverParser: model.NewPgServerParser(),
						clientParser: model.NewPgClientParser(),
					}
					break
				}
			}
			s.outputs = append(s.outputs, output)
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

func (s *OutputService) Run() error {
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

	return nil
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

func (o *ConsoleOutput) Write(ctx Context) error {
	if ctx.Data().Forward == model.ServerToClient {
		o.dump(YELLOW, ctx)
	} else {
		o.dump(GREEN, ctx)
	}
	return nil
}

func (o *ConsoleOutput) Close(wg *sync.WaitGroup) {
	defer wg.Done()
}

func (o *ConsoleOutput) sDump(a interface{}) string {
	return o.printConfig().Sdump(a)
}

func (o *ConsoleOutput) dump(color string, ctx Context) {
	consolePrint(color, fmt.Sprintf("[%d] %s", ctx.SessionId(), o.sDump(ctx.Data())))
}

func (o *ConsoleOutput) printConfig() *spew.ConfigState {
	scs := spew.NewDefaultConfig()
	scs.Indent = "    "
	return scs
}

func (o *sqliConsoleOutput) Write(ctx Context) error {
	if _, ok := o.counts[ctx.SessionId()]; !ok {
		o.counts[ctx.SessionId()] = 0
	}
	color := RED
	if ctx.Data().Forward == model.ServerToClient {
		color = YELLOW
	} else {
		color = GREEN
	}

	if o.counts[ctx.SessionId()] < 2 {
		o.dump(color, ctx)
	} else {
		reader := bytes.NewReader(ctx.Data().Buffer)
		cmds, err := model.UnpackSqliTransmission(reader)
		if err != nil {
			o.dump(RED, ctx)
		} else {
			printPackage(color, ctx, o.sDump(cmds))
		}

	}

	o.counts[ctx.SessionId()]++
	return nil
}

func (o *pgConsoleOutput) Write(ctx Context) error {
	if _, ok := o.counts[ctx.SessionId()]; !ok {
		o.counts[ctx.SessionId()] = 0
	}
	color := RED
	if ctx.Data().Forward == model.ServerToClient {
		color = YELLOW
	} else {
		color = GREEN
	}

	if o.counts[ctx.SessionId()] == 0 {
		backend, err := pgproto3.NewBackend(
			pgproto3.NewChunkReader(bytes.NewReader(ctx.Data().Buffer)),
			nil)
		msg, err := backend.ReceiveStartupMessage()
		if err != nil {
			o.dump(color, ctx)
		} else {
			buf, err := json.MarshalIndent(msg, "", "  ")
			if err != nil {
				o.dump(color, ctx)
			} else {
				printPackage(color, ctx, string(buf))
			}
		}
	} else {
		if ctx.Data().Forward == model.ServerToClient {
			o.serverParser.Append(ctx.Data().Buffer)
			msg, err := o.serverParser.ParseMessage()
			if err != nil {
				o.dump(color, ctx)
			} else {
				for _, i := range msg {
					buf, err := json.MarshalIndent(i, "", "  ")
					if err != nil {
						o.dump(color, ctx)
					} else {
						printPackage(color, ctx, string(buf))
					}
				}
			}
		} else {
			o.clientParser.Append(ctx.Data().Buffer)
			msg, err := o.clientParser.ParseMessage()
			if err != nil {
				o.dump(color, ctx)
			} else {
				for _, i := range msg {
					buf, err := json.MarshalIndent(i, "", "  ")
					if err != nil {
						o.dump(color, ctx)
					} else {
						printPackage(color, ctx, string(buf))
					}
				}
			}
		}
	}

	o.counts[ctx.SessionId()]++
	return nil
}

func printPackage(color string, ctx Context, str string) {
	consolePrint(color, fmt.Sprintf("[%d] %s %s", ctx.SessionId(), ctx.Data().Forward, str))
}

func consolePrint(color string, str string) {
	scs := spew.NewDefaultConfig()
	scs.Indent = "    "
	_, _ = scs.Print(color)
	_, _ = scs.Print(str)
	_, _ = scs.Println(CLEAR)
}

func (o *FileOutput) Write(ctx Context) error {
	var header struct {
		Index   uint16
		Forward uint8
		Length  int64
	}
	header.Index = uint16(o.index)
	header.Forward = uint8(ctx.Data().Forward)
	header.Length = int64(len(ctx.Data().Buffer))
	buf := new(bytes.Buffer)
	_ = binary.Write(buf, binary.LittleEndian, header)
	_, _ = o.file.Write(buf.Bytes())
	_, _ = o.file.Write(ctx.Data().Buffer)
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
