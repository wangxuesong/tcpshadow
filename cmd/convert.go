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
	"io"
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/wangxuesong/tcpshadow/convert/json"
)

var (
	fromFormat string
	toFormat   string
)

// convertCmd represents the convert command
var convertCmd = &cobra.Command{
	Use:   "convert INPUT",
	Short: "Convert between capture files and other file formats",
	Long:  `Convert between capture files and other file formats`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if fromFormat == "" && toFormat == "" {
			log.Fatal("Error: please set flag from or flag to")
		}

		if fromFormat != "" && toFormat != "" {
			log.Fatal("Error: ")
		}

		inputFile = args[0]

		if fromFormat != "" {
			err := convertFrom(inputFile, outputFile, "json")
			if err != nil {
				panic(err)
			}
			return
		}

		if toFormat != "" {
			err := convertTo(inputFile, outputFile, "json")
			if err != nil {
				panic(err)
			}
			return
		}

	},
}

func convertFrom(source, target, _ string) error {
	file, err := os.Open(source)
	if err != nil {
		return err
	}
	defer file.Close()
	jsonStr, err := io.ReadAll(file)
	output, err := json.UnmarshalSqliSavePackages(string(jsonStr))
	if err != nil {
		return err
	}

	if target != "" {
		file, err := os.Create(target)
		if err != nil {
			return err
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
			return err
		}
		_ = file.Sync()
	} else {
		fmt.Println(output)
	}

	return nil
}

func convertTo(source, target, _ string) error {
	packages := ReadPackages(source)
	jsonStr, err := json.MarshalSqliSavePackages(packages)
	if err != nil {
		return err
	}

	if target != "" {
		file, err := os.Create(target)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = file.WriteString(jsonStr)
		if err != nil {
			return err
		}
	} else {
		fmt.Println(jsonStr)
	}

	return nil
}

func init() {
	rootCmd.AddCommand(convertCmd)

	convertCmd.Flags().StringVarP(&fromFormat, "from", "f", "", "from FORMAT")
	convertCmd.Flags().StringVarP(&toFormat, "to", "t", "", "to FORMAT")
	convertCmd.Flags().StringVarP(&outputFile, "output", "o", "", "output file")
	_ = convertCmd.MarkFlagRequired("output")
}
