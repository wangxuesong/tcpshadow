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
	"github.com/spf13/cobra"
	"github.com/wangxuesong/tcpshadow/model"
	"log"
	"net"
)

// serverCmd represents the server command
var clientCmd = &cobra.Command{
	Use:   "client",
	Short: "Client request",
	Long:  `Client request`,
	Run:   runClientCommand,
}

func runClientCommand(cmd *cobra.Command, args []string) {
	//fmt.Println("server called")
	datas := ReadPackages()

	log.Println("Start...")
	log.Println("Connect to server")
	server, err := net.Dial("tcp", serverAddress)
	if err != nil {
		panic(err)
	}

	for _, d := range datas {
		switch d.Forward {
		case model.ClientToServer:
			_, err := server.Write(d.Buffer)
			if err != nil {
				panic(err)
			}
		case model.ServerToClient:
			buf := make([]byte, 16384)
			_, err := server.Read(buf)
			if err != nil {
				panic(err)
			}
			printPack(YELLOW, d)
		}
	}
}

func init() {
	rootCmd.AddCommand(clientCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// serverCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// serverCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	clientCmd.Flags().StringVarP(&serverAddress, "address", "l", "", "server address")
	clientCmd.MarkFlagRequired("address")
	clientCmd.Flags().StringVarP(&inputFile, "input", "i", "", "input file")
	clientCmd.MarkFlagRequired("input")
}
