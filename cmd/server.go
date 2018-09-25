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
	"encoding/binary"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/wangxuesong/tcpshadow/model"
	"io"
	"net"
	"os"
)

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Server response",
	Long: `Server response`,
	Run: func(cmd *cobra.Command, args []string) {
		//fmt.Println("server called")
		datas := ReadPackages()
		clientConn, err := net.Listen("tcp4", listenAddress)
		if err != nil {
			panic(err)
		}
		client, err := clientConn.Accept()
		if err != nil {
			panic(err)
		}

		for _, d := range datas {
			switch d.Forward {
			case model.ClientToServer:
				buf := make([]byte, 16384)
				_, err := client.Read(buf)
				if err != nil {
					panic(err)
				}
				if d.Number == 0 {
					continue
				}
			case model.ServerToClient:
				client.Write(d.Buffer)
			}
		}
	},
}

var (
	inputFile string
)

func init() {
	rootCmd.AddCommand(serverCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// serverCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// serverCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	serverCmd.Flags().StringVarP(&listenAddress, "listen", "l", "", "listen address")
	serverCmd.MarkFlagRequired("listen")
	serverCmd.Flags().StringVarP(&inputFile, "input", "i", "", "input file")
	serverCmd.MarkFlagRequired("input")
}
func ReadPackages() []model.SavePackage {
	file, err := os.Open(inputFile)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	packages := make([]model.SavePackage, 0, 30)

	for {
		var header struct {
			Index   uint16
			Forward uint8
			Length  uint32
		}
		err := binary.Read(file, binary.LittleEndian, &header)
		if err == io.EOF {
			return packages
		}
		if err != nil {
			panic(err)
		}
		fmt.Println(header)
		index := int(header.Index)
		forward := model.DataForward(header.Forward)
		length := int(header.Length)
		buf := make([]byte, length)
		data := model.SavePackage{
			Number:  index,
			Forward: forward,
			Length:  length,
			Buffer:  buf,
		}
		count, err := file.Read(data.Buffer)
		fmt.Println(index, forward, length, count)
		fmt.Println(data)
		if err != nil {
			return packages
		}
		packages = append(packages, data)
	}
}

