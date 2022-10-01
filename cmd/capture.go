// Copyright © 2018 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/jackc/pgproto3"
	"github.com/spf13/cobra"
	"github.com/wangxuesong/tcpshadow/model"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// captureCmd represents the capture command
var captureCmd = &cobra.Command{
	Use:          "capture",
	Short:        "Capture tcp data",
	Long:         `Capture tcp data`,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		//fmt.Println("capture called")
		capture()
		return nil
	},
}

var (
	serverAddress string
	listenAddress string
	outputFile    string
	printPackage  bool
	protocolType  string
)

func init() {
	rootCmd.AddCommand(captureCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	captureCmd.Flags().StringVarP(&serverAddress, "server", "s", "", "server address")
	_ = captureCmd.MarkFlagRequired("server")
	captureCmd.Flags().StringVarP(&listenAddress, "listen", "l", "", "listen address")
	_ = captureCmd.MarkFlagRequired("listen")
	captureCmd.Flags().StringVarP(&outputFile, "output", "o", "", "output file")
	_ = captureCmd.MarkFlagRequired("output")
	captureCmd.Flags().BoolVarP(&printPackage, "printPackage", "p", false, "printPackage package")
	captureCmd.Flags().StringVarP(&protocolType, "type", "t", "sqli", "protocol type")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// captureCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

type Service struct {
	done      chan bool
	waitGroup *sync.WaitGroup
}

// NewService Make a new Service.
func NewService() *Service {
	s := &Service{
		done:      make(chan bool),
		waitGroup: &sync.WaitGroup{},
	}
	//s.waitGroup.Add(1)
	return s
}

// Stop the service by closing the service's channel.  Block until the service
// is really stopped.
func (s *Service) Stop() {
	close(s.done)
	s.waitGroup.Wait()
}

// Serve Accept connections and spawn a goroutine to serve each one.
// Stop listening if anything is received on the service's channel.
func (s *Service) Serve(listener *net.TCPListener) {
	defer s.waitGroup.Done()
	for {
		select {
		case <-s.done:
			log.Println("stopping listening on", listener.Addr())
			_ = listener.Close()
			return
		default:
		}
		_ = listener.SetDeadline(time.Now().Add(1e9))
		client, err := listener.AcceptTCP()
		if nil != err {
			if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
				continue
			}
			log.Println(err)
		}
		log.Println(client.RemoteAddr(), "connected")
		server, err := connectServer(client)
		if err != nil {
			panic(err)
		}

		s.waitGroup.Add(1)
		monitor := make(chan model.Data)

		go func() {
			file, err := os.OpenFile(outputFile, os.O_CREATE|os.O_RDWR, 0644)
			if err != nil {
				log.Printf("Can't open file with %s\n", err)
			}
			defer file.Close()
			if err := file.Truncate(0); err != nil {
				panic(err)
			}

			serverParser := model.NewPgServerParser()
			clientParser := model.NewPgClientParser()
			index := 0
			for {
				select {
				case d := <-monitor:
					var header struct {
						Index   uint16
						Forward uint8
						Length  int64
					}
					header.Index = uint16(index)
					header.Forward = uint8(d.Forward)
					header.Length = int64(len(d.Buffer))
					buf := new(bytes.Buffer)
					_ = binary.Write(buf, binary.LittleEndian, header)
					_, _ = file.Write(buf.Bytes())
					_, _ = file.Write(d.Buffer)
					_ = file.Sync()
					if printPackage {
						scs := spew.NewDefaultConfig()
						scs.Indent = "    "
						msgDump := spew.NewDefaultConfig()
						msgDump.DisableMethods = true
						msgDump.DisablePointerAddresses = true
						msgDump.DisablePointerMethods = true
						data := model.SavePackage{
							Number:  int(header.Index),
							Forward: model.DataForward(header.Forward),
							Length:  int(header.Length),
							Buffer:  d.Buffer,
						}

						colorStr := GREEN
						clearStr := CLEAR
						if d.Forward == model.ServerToClient {
							colorStr = YELLOW
						}
						warningStr := RED
						_, _ = scs.Print(colorStr)
						switch protocolType {
						case "pg":
							//scs.Dump(data)
							if data.Number < 1 {
								if data.Number == 0 {
									backend, err := pgproto3.NewBackend(
										pgproto3.NewChunkReader(bytes.NewReader(data.Buffer)),
										nil)
									msg, err := backend.ReceiveStartupMessage()
									//err := msg.Decode(data.Buffer)
									//msg, err := pgproto3.ParseStartupMessage(bytes.NewReader(data.Buffer))
									if err != nil {
										scs.Dump(data)
									} else {
										buf, err := json.MarshalIndent(msg, "", "  ")
										if err != nil {
											msgDump.Dump(msg)
										} else {
											fmt.Println("C->S", string(buf))
										}
									}
									break
								}
							} else {
								if data.Forward == model.ServerToClient {
									serverParser.Append(data.Buffer)
									msg, err := serverParser.ParseMessage()
									if err != nil {
										scs.Dump(data)
									} else {
										for _, i := range msg {
											buf, err := json.MarshalIndent(i, "", "  ")
											if err != nil {
												msgDump.Dump(i)
											} else {
												fmt.Println("S->C", string(buf))
											}
										}
									}
								} else {
									clientParser.Append(data.Buffer)
									msg, err := clientParser.ParseMessage()
									if err != nil {
										scs.Dump(data)
									} else {
										for _, i := range msg {
											buf, err := json.MarshalIndent(i, "", "  ")
											if err != nil {
												msgDump.Dump(i)
											} else {
												fmt.Println("C->S", string(buf))
											}
										}
									}
								}
							}
						case "sqli":
							if data.Number < 2 {
								scs.Dump(data)
							} else {
								reader := bytes.NewReader(data.Buffer)
								cmds, err := model.UnpackSqliTransmission(reader)
								if err != nil {
									_, _ = scs.Print(warningStr)
									_, _ = scs.Println(err)
									scs.Dump(data)
									_, _ = scs.Println(clearStr)
									index++
									continue
								}
								scs.Dump(cmds)
							}
						}
						_, _ = scs.Println(clearStr)
					}
				case <-s.done:
					_ = file.Close()
					return
				}
				index++
			}
		}()
		go s.proxyData(client, server, model.ClientToServer, monitor)
		go s.proxyData(server, client, model.ServerToClient, monitor)
	}
}

func capture() {
	log.Println("Start....")
	clientConn, err := net.ResolveTCPAddr("tcp4", listenAddress)
	if nil != err {
		log.Fatalln(err)
	}
	listener, err := net.ListenTCP("tcp4", clientConn)
	if nil != err {
		log.Fatalln(err)
	}
	log.Println("listening on", listener.Addr())

	// Make a new service and send it into the background.
	service := NewService()
	go service.Serve(listener)

	// Handle SIGINT and SIGTERM.
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	log.Println(<-ch)

	// Stop the service gracefully.
	service.Stop()
}

func connectServer(client net.Conn) (net.Conn, error) {
	if client == nil {
		return nil, errors.New("client nil")
	}

	log.Println("Connect to server")

	server, err := net.Dial("tcp", serverAddress)
	return server, err
}

func (s *Service) proxyData(src net.Conn, dest net.Conn, forward model.DataForward, monitor chan model.Data) {
	defer src.Close()
	defer s.waitGroup.Done()
	//buf := make([]byte, 16384)
	s.waitGroup.Add(1)
	for {
		select {
		case <-s.done:
			log.Println("disconnecting", src.RemoteAddr())
			return
		default:
		}
		_ = src.SetDeadline(time.Now().Add(1e9))
		buf := make([]byte, 16384)
		cnt, err := src.Read(buf)
		if nil != err {
			if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
				continue
			}
			if err == io.EOF {
				log.Println("disconnecting", src.RemoteAddr())
				//s.done <- true
				return
			}
			log.Println(err)
			return
		}
		data := model.Data{Forward: forward, Buffer: buf[:cnt]}
		monitor <- data
		if _, err := dest.Write(buf[:cnt]); nil != err {
			log.Println(err)
			return
		}
	}
}
