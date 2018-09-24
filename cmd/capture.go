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
	"errors"
	"fmt"
	"io"
	"net"
	"os"

	"github.com/wangxuesong/tcpshadow/model"

	"github.com/spf13/cobra"
)

// captureCmd represents the capture command
var captureCmd = &cobra.Command{
	Use:          "capture",
	Short:        "Capture tcp data",
	Long:         `Capture tcp data`,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("capture called")
		capture()
		return errors.New("abc")
	},
}

var (
	serverAddress string
	listenAddress string
	outputFile    string
)

func init() {
	rootCmd.AddCommand(captureCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	captureCmd.Flags().StringVarP(&serverAddress, "server", "s", "", "server address")
	captureCmd.MarkFlagRequired("server")
	captureCmd.Flags().StringVarP(&listenAddress, "listen", "l", "", "listen address")
	captureCmd.MarkFlagRequired("listen")
	captureCmd.Flags().StringVarP(&outputFile, "output", "o", "", "output file")
	captureCmd.MarkFlagRequired("output")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// captureCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
func capture() {
	fmt.Println("Start....")
	clientConn, err := net.Listen("tcp4", listenAddress)
	if err != nil {
		panic(err)
	}
	for {
		client, err := clientConn.Accept()
		if err != nil {
			panic(err)
		}

		server, err := connectServer(client)
		if err != nil {
			panic(err)
		}

		monitor := make(chan model.Data)

		go func() {
			file, err := os.OpenFile(outputFile, os.O_CREATE|os.O_RDWR, 0644)
			if err != nil {
				fmt.Printf("Can't open file with %s\n", err)
			}
			defer file.Close()

			index := 0
			for {
				select {
				case d := <-monitor:
					fmt.Printf("%s: %#v\n", d.Forward, d.Buffer)
					var header struct {
						Index   uint16
						Forward uint8
						Length  uint32
					}
					header.Index = uint16(index)
					header.Forward = uint8(d.Forward)
					header.Length = uint32(len(d.Buffer))
					buf := new(bytes.Buffer)
					//binary.Write(buf, binary.LittleEndian, uint16(index))
					//binary.Write(buf, binary.LittleEndian, uint8(d.Forward))
					binary.Write(buf, binary.LittleEndian, header)
					file.Write(buf.Bytes())
					file.Write(d.Buffer)
					file.Sync()
				}
				index++
			}
		}()

		go proxyData(client, server, model.ClientToServer, monitor)
		go proxyData(server, client, model.ServerToClient, monitor)
	}
}

func connectServer(client net.Conn) (net.Conn, error) {
	if client == nil {
		return nil, errors.New("Client nil")
	}

	fmt.Println("Connect to server")

	server, err := net.Dial("tcp4", serverAddress)
	return server, err
}

func proxyData(src net.Conn, dest net.Conn, forward model.DataForward, monitor chan model.Data) {
	buf := make([]byte, 16384)
	for {
		cnt, err := src.Read(buf)
		if err == io.EOF {
			return
		}
		if err != nil {
			panic(err)
		}
		data := model.Data{Forward: forward, Buffer: buf[:cnt]}
		monitor <- data
		dest.Write(buf[:cnt])
	}
}
