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
	"io"
	"log"
	"net"
	"sync"
	"time"
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

	number := 0
	waitGroup := &sync.WaitGroup{}
	done := make(chan bool)
	defer close(done)
	proxy := func(src net.Conn, forward model.DataForward, monitor chan model.Data) {
		defer src.Close()
		defer waitGroup.Done()
		//buf := make([]byte, 16384)
		waitGroup.Add(1)
		for {
			select {
			case <-done:
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
		}
	}

	color = true
	log.Println("Start...")
	log.Println("Connect to server")
	server, err := net.Dial("tcp", serverAddress)
	if err != nil {
		panic(err)
	}

	clientDataChan := make(chan model.SavePackage)
	clientNextChan := make(chan struct{})

	serverDataChan := make(chan model.Data)

	go func() {
		for {
			select {
			case d := <-clientDataChan:
				_, err := server.Write(d.Buffer)
				if err != nil {
					panic(err)
				}
				printPack(GREEN, d)
			case d := <-serverDataChan:
				data := model.SavePackage{
					Number:  number,
					Forward: d.Forward,
					Length:  len(d.Buffer),
					Buffer:  d.Buffer,
				}
				number++
				printPack(YELLOW, data)
				clientNextChan <- struct{}{}
			}
		}
	}()

	go proxy(server, model.ServerToClient, serverDataChan)

	for _, d := range datas {
		switch d.Forward {
		case model.ClientToServer:
			//_, err := server.Write(d.Buffer)
			//if err != nil {
			//	panic(err)
			//}
			//printPack(GREEN, d)
			clientDataChan <- d
			number++
			<-clientNextChan
		case model.ServerToClient:
			continue
			buf := make([]byte, 16384)
			_, err := server.Read(buf)
			if err != nil {
				panic(err)
			}
			data := model.SavePackage{
				Number:  113,
				Forward: model.ServerToClient,
				Length:  len(buf),
				Buffer:  buf,
			}
			printPack(YELLOW, data)

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
	clientCmd.Flags().BoolVarP(&printPackage, "parse", "p", false, "parse package")
}
