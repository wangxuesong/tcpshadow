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
	"fmt"
	"github.com/spf13/cobra"
	"io"
	"log"
	"os"
)

// fromJsonCmd represents the fromJson command
var fromJsonCmd = &cobra.Command{
	Use:   "fromjson",
	Short: "convert json to capture file",
	Long:  `convert json to capture file`,
	Run: func(cmd *cobra.Command, args []string) {
		//packages := ReadPackages()
		file, err := os.Open(inputFile)
		if err != nil {
			panic(err)
		}
		defer file.Close()
		jsonStr, err := io.ReadAll(file)
		output, err := unmarshalSqliSavePackages(string(jsonStr))
		if err != nil {
			log.Fatal(err)
		}

		if outputFile != "" {
			file, err := os.Create(outputFile)
			if err != nil {
				panic(err)
			}
			defer file.Close()

			buf := new(bytes.Buffer)
			for _, savePackage := range output {
				var header struct {
					Index   uint16
					Forward uint8
					Length  int64
				}
				header.Index = uint16(savePackage.Number)
				header.Forward = uint8(savePackage.Forward)
				header.Length = int64(savePackage.Length)
				_ = binary.Write(buf, binary.LittleEndian, header)
				buf.Write(savePackage.Buffer)
			}
			_, err = file.Write(buf.Bytes())
			if err != nil {
				panic(err)
			}
			_ = file.Sync()
		} else {
			fmt.Println(output)
		}
	},
}

func init() {
	rootCmd.AddCommand(fromJsonCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// toJsonCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	fromJsonCmd.Flags().StringVarP(&inputFile, "input", "i", "", "input file")
	fromJsonCmd.MarkFlagRequired("input")
	fromJsonCmd.Flags().StringVarP(&outputFile, "output", "o", "", "output file")
	fromJsonCmd.MarkFlagRequired("output")
	//toJsonCmd.Flags().BoolVarP(&color, "color", "c", false, "print with color")
	//toJsonCmd.Flags().BoolVarP(&raw, "raw", "r", false, "show raw")
	//toJsonCmd.Flags().BoolVarP(&printPackage, "parse", "p", false, "parse package")
}
