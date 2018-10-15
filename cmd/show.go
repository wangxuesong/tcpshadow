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
	"github.com/davecgh/go-spew/spew"
	"github.com/wangxuesong/tcpshadow/model"

	"github.com/spf13/cobra"
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
		packages := ReadPackages()
		scs := spew.NewDefaultConfig()
		scs.Indent = "    "
		var desc *model.SqliDescribe
		printPack := func(colorStr string, savePackage model.SavePackage) {
			clearStr := CLEAR
			warningStr := RED
			if !color {
				colorStr = ""
				clearStr = ""
				warningStr = ""
			}
			scs.Print(colorStr)
			if raw {
				scs.Printf("%v\n", savePackage)
			} else {
				if print && savePackage.Number >= 2 {
					reader := bytes.NewReader(savePackage.Buffer)
					cmds, err := model.UnpackSqliTransmission(reader)
					if cmds[0].Command() == 8 {
						desc = cmds[0].(*model.SqliDescribe)
					}
					if cmds[0].Command() == new(model.SqliTuple).Command() {
						if desc != nil {
							tuple := cmds[0].(*model.SqliTuple)
							tuple.SetMetaData(desc.Fields)
							tuple.UnpackValues()
						}
					}
					if err != nil {
						scs.Print(warningStr)
						scs.Println(err)
						scs.Dump(savePackage)
					} else {
						scs.Dump(cmds)
					}
				} else {
					scs.Dump(savePackage)
				}
			}
			scs.Println(clearStr)
		}

		for _, pack := range packages {
			switch pack.Forward {
			case model.ClientToServer:
				printPack(GREEN, pack)
			case model.ServerToClient:
				printPack(YELLOW, pack)
			}
		}
	},
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
	showCmd.MarkFlagRequired("input")
	showCmd.Flags().BoolVarP(&color, "color", "c", false, "print with color")
	showCmd.Flags().BoolVarP(&raw, "raw", "r", false, "show raw")
	showCmd.Flags().BoolVarP(&print, "parse", "p", false, "parse package")
}
