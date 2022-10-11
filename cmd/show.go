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
	"encoding/json"
	"fmt"

	"github.com/davecgh/go-spew/spew"
	"github.com/jackc/pgproto3"
	"github.com/spf13/cobra"
	"github.com/wangxuesong/tcpshadow/model"
)

var (
	WHITE  = "\033[37m"
	CYAN   = "\033[36m"
	PURPLE = "\033[95m"
	BLUE   = "\033[94m"
	GREEN  = "\033[92m"
	YELLOW = "\033[93m"
	RED    = "\033[91m"
	CLEAR  = "\033[0m"

	UNDERLINE            = "\033[4m"
	UNDERLINE_WHITE      = "\033[4m\033[37m"
	UNDERLINE_BOLD_WHITE = "\033[4m\033[1m\033[37m"

	BOLD        = "\033[1m"
	BOLD_CYAN   = "\033[1m\033[36m"
	BOLD_PURPLE = "\033[1m\033[95m"
	BOLD_YELLOW = "\033[1m\033[93m"
	BOLD_RED    = "\033[1m\033[91m"

	INVERSE_WHITE  = "\033[7m\033[37m"
	INVERSE_PURPLE = "\033[7m\033[95m"
	INVERSE_BLUE   = "\033[7m\033[94m"
	INVERSE_GREEN  = "\033[7m\033[92m"
	INVERSE_YELLOW = "\033[7m\033[93m"
	INVERSE_RED    = "\033[7m\033[91m"

	color bool
	raw   bool
)

// showCmd represents the show command
var showCmd = &cobra.Command{
	Use:   "show",
	Short: "show capture file",
	Long:  `show capture file`,
	Run: func(cmd *cobra.Command, args []string) {
		switch protocolType {
		case "sqli":
			showSqli()
		case "pg":
			showPg()
		}
	},
}

func showPg() {
	serverParser := model.NewPgServerParser()
	clientParser := model.NewPgClientParser()
	packages := ReadPackages(inputFile)
	scs := spew.NewDefaultConfig()
	scs.Indent = "    "
	for _, pack := range packages {
		switch pack.Forward {
		case model.ClientToServer:
			if pack.Number == 0 {
				backend, err := pgproto3.NewBackend(
					pgproto3.NewChunkReader(bytes.NewReader(pack.Buffer)),
					nil)
				msg, err := backend.ReceiveStartupMessage()
				if err != nil {
					printError(scs, RED, err)
					continue
				}
				buf, err := json.MarshalIndent(msg, "", "  ")
				if err != nil {
					printError(scs, RED, err)
					continue
				} else {
					_, _ = scs.Print(GREEN)
					_, _ = scs.Print(fmt.Sprintln("C->S", string(buf)))
					_, _ = scs.Print(CLEAR)
				}
				continue
			}
			printPgPack(GREEN, pack, clientParser)
		case model.ServerToClient:
			printPgPack(YELLOW, pack, serverParser)
		}
	}
}

func printPgPack(colorStr string, savePackage model.SavePackage, parser model.MessageParser) {
	scs := spew.NewDefaultConfig()
	scs.Indent = "    "
	clearStr := CLEAR
	warningStr := RED
	if !color {
		colorStr = ""
		clearStr = ""
		warningStr = ""
	}
	_, _ = scs.Print(colorStr)
	if raw {
		_, _ = scs.Printf("%v\n", savePackage)
	} else {
		if printPackage {
			_, _ = parser.Append(savePackage.Buffer)
			msg, err := parser.ParseMessage()
			if err != nil {
				printError(scs, warningStr, err)
				//scs.Dump(savePackage)
			} else {
				for _, i := range msg {
					buf, err := json.MarshalIndent(i, "", "  ")
					if err == nil {
						fmt.Println("S->C", string(buf))
					} else {

					}
				}
			}
		} else {
			scs.Dump(savePackage)
		}
	}
	_, _ = scs.Println(clearStr)
}

func printError(scs *spew.ConfigState, color string, err error) {
	_, _ = scs.Print(color)
	_, _ = scs.Print(err)
}

func showSqli() {
	packages := ReadPackages(inputFile)
	scs := spew.NewDefaultConfig()
	scs.Indent = "    "
	for _, pack := range packages {
		switch pack.Forward {
		case model.ClientToServer:
			printSqliPack(GREEN, pack)
		case model.ServerToClient:
			printSqliPack(YELLOW, pack)
		}
	}
}

func printSqliPack(colorStr string, savePackage model.SavePackage) {
	scs := spew.NewDefaultConfig()
	scs.Indent = "    "
	clearStr := CLEAR
	warningStr := RED
	if !color {
		colorStr = ""
		clearStr = ""
		warningStr = ""
	}
	_, _ = scs.Print(colorStr)
	if raw {
		_, _ = scs.Printf("%v\n", savePackage)
	} else {
		var desc *model.SqliDescribe
		if printPackage && savePackage.Number >= 2 {
			reader := bytes.NewReader(savePackage.Buffer)
			cmds, err := model.UnpackSqliTransmission(reader)
			if err == nil {
				if cmds[0].Command() == 8 {
					desc = cmds[0].(*model.SqliDescribe)
				}
				if cmds[0].Command() == new(model.SqliTuple).Command() {
					if desc != nil {
						tuple := cmds[0].(*model.SqliTuple)
						tuple.SetMetaData(desc.Fields)
						_ = tuple.UnpackValues()
					}
				}
			}
			if err != nil {
				_, _ = scs.Print(warningStr)
				_, _ = scs.Println(err)
				scs.Dump(savePackage)
			} else {
				scs.Dump(cmds)
			}
		} else {
			scs.Dump(savePackage)
		}
	}
	_, _ = scs.Println(clearStr)
}

func init() {
	rootCmd.AddCommand(showCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// showCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	showCmd.Flags().StringVarP(&inputFile, "input", "i", "", "input file")
	_ = showCmd.MarkFlagRequired("input")
	showCmd.Flags().BoolVarP(&color, "color", "c", false, "print with color")
	showCmd.Flags().BoolVarP(&raw, "raw", "r", false, "show raw")
	showCmd.Flags().BoolVarP(&printPackage, "parse", "p", false, "parse package")
	showCmd.Flags().StringVarP(&protocolType, "type", "t", "sqli", "protocol type")
}
