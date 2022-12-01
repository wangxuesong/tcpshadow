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
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/wangxuesong/tcpshadow/pkg/services/bridge"
	"github.com/wangxuesong/tcpshadow/pkg/supervisor"
)

type bridgeCommand struct {
	command *cobra.Command
	Config  *supervisor.Config
}

// bridgeCmd represents the capture command
var bridgeCmd = &bridgeCommand{
	Config: config,
	command: &cobra.Command{
		Use:          "bridge",
		Short:        "",
		Long:         ``,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(config)
		},
	},
}

func run(config *supervisor.Config) error {
	err := config.Validate()
	if err != nil {
		return err
	}

	super := supervisor.NewSupervisor(config, bridge.NewBridgeService)
	go super.Serve()

	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	log.Println(<-ch)

	wg := &sync.WaitGroup{}
	wg.Add(1)
	super.Close(wg)
	wg.Wait()

	return nil
}

func init() {
	rootCmd.AddCommand(bridgeCmd.command)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	config := bridgeCmd.Config
	bridgeCmd.command.Flags().StringVarP(&config.ServerAddress, "server", "s", "", "server address")
	_ = bridgeCmd.command.MarkFlagRequired("server")
	bridgeCmd.command.Flags().StringVarP(&config.ListenAddress, "listen", "l", "", "listen address")
	_ = bridgeCmd.command.MarkFlagRequired("listen")
	bridgeCmd.command.Flags().StringVarP(&config.OutputFile, "output", "o", "", "output file")
	_ = bridgeCmd.command.MarkFlagRequired("output")
	bridgeCmd.command.Flags().BoolVarP(&config.IsPrintPackage, "printPackage", "p", false, "printPackage package")
	bridgeCmd.command.Flags().StringVarP(&config.ProtocolType, "type", "t", "sqli", "protocol type")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// bridgeCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
