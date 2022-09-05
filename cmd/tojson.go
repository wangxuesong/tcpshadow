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
	"github.com/spf13/cobra"
	"github.com/wangxuesong/tcpshadow/model"
	"io"
	"log"
	"os"
)

// toJsonCmd represents the tojson command
var toJsonCmd = &cobra.Command{
	Use:   "tojson",
	Short: "convter capture file to json",
	Long:  `convter capture file to json`,
	Run: func(cmd *cobra.Command, args []string) {
		packages := ReadPackages()
		jsonStr, err := marshalSqliSavePackages(packages)
		if err != nil {
			log.Fatal(err)
		}

		if outputFile != "" {
			file, err := os.Create(outputFile)
			if err != nil {
				panic(err)
			}
			defer file.Close()

			_, err = file.WriteString(jsonStr)
			if err != nil {
				panic(err)
			}
		} else {
			fmt.Println(jsonStr)
		}
	},
}

type packageForJson struct {
	Name        string               `json:"name""`
	AuthPackage [2]model.SavePackage `json:"auth,omitempty"`
	SqliPackage []sqliPackage        `json:"package,omitempty"`
}

type jsonSqliCommand struct {
	Type              int `json:"type"`
	model.SqliCommand `json:"sqli"`
}

type sqliPackage struct {
	Number  int               `json:"number"`
	Forward model.DataForward `json:"forward"`
	Command []jsonSqliCommand `json:"command"`
}

func marshalSqliPackage(packages []model.SavePackage) ([]sqliPackage, error) {
	if len(packages) == 0 {
		return nil, nil
	}

	var sqliPack []sqliPackage
	for _, input := range packages {
		output := sqliPackage{
			Number:  input.Number,
			Forward: input.Forward,
		}
		buffer := bytes.NewReader(input.Buffer)
		for {
			sqli, err := model.UnpackSqliCommand(buffer)
			if err == io.EOF {
				break
			}
			if err != nil {
				return nil, err
			}
			var cmd jsonSqliCommand
			cmd.Type = int(sqli.Command())
			cmd.SqliCommand = sqli
			output.Command = append(output.Command, cmd)
		}
		sqliPack = append(sqliPack, output)
	}

	return sqliPack, nil
}

func marshalSqliSavePackages(packages []model.SavePackage) (string, error) {
	pack := packageForJson{Name: "test"}
	pack.AuthPackage[0] = packages[0]
	pack.AuthPackage[1] = packages[1]
	sqliPackages, err := marshalSqliPackage(packages[2:])
	pack.SqliPackage = sqliPackages

	buf, err := json.MarshalIndent(pack, "", "  ")
	if err != nil {
		return "", err
	}

	tmp, _ := unmarshalSqliSavePackages(string(buf))
	if tmp == nil {

	}
	return string(buf), nil
}

func unmarshalCommand[T model.SqliCommand](str json.RawMessage) (model.SqliCommand, error) {
	var cmd T
	err := json.Unmarshal(str, &cmd)
	if err != nil {
		return nil, err
	}
	return cmd, nil
}

func unmarshalSqliSavePackages(jsonString string) ([]model.SavePackage, error) {
	buf := bytes.NewBufferString(jsonString)
	type unmarshalSqliPackage struct {
		Number  int               `json:"number"`
		Forward model.DataForward `json:"forward"`
		Type    int               `json:"type"`
		Command []json.RawMessage `json:"command"`
	}
	type packageForUnmarshalJson struct {
		Name        string                 `json:"name""`
		AuthPackage [2]model.SavePackage   `json:"auth,omitempty"`
		SqliPackage []unmarshalSqliPackage `json:"package,omitempty"`
	}

	var anyJson packageForUnmarshalJson
	json.Unmarshal(buf.Bytes(), &anyJson)

	result := make([]model.SavePackage, 0)
	for _, p := range anyJson.AuthPackage {
		result = append(result, p)
	}

	type commandType struct {
		Type int `json:"type"`
		Sqli json.RawMessage
	}
	for _, p := range anyJson.SqliPackage {
		savePackage := model.SavePackage{
			Number:  p.Number,
			Forward: p.Forward,
			Length:  0,
			Buffer:  nil,
		}

		var trans model.SqliTransmission
		for _, message := range p.Command {
			var t commandType
			err := json.Unmarshal(message, &t)
			if err != nil {
				return nil, err
			}

			switch t.Type {
			case 2:
				cmd, err := unmarshalCommand[*model.SqliPrepare](t.Sqli)
				if err != nil {
					return nil, err
				}
				trans.Append(cmd)
			case 3:
				cmd, err := unmarshalCommand[*model.SqliCurName](t.Sqli)
				if err != nil {
					return nil, err
				}
				trans.Append(cmd)
			case 4:
				cmd, err := unmarshalCommand[*model.SqliID](t.Sqli)
				if err != nil {
					return nil, err
				}
				trans.Append(cmd)
			case 6:
				cmd, err := unmarshalCommand[*model.SqliOpen](t.Sqli)
				if err != nil {
					return nil, err
				}
				trans.Append(cmd)
			case 8:
				cmd, err := unmarshalCommand[*model.SqliDescribe](t.Sqli)
				if err != nil {
					return nil, err
				}
				trans.Append(cmd)
			case 9:
				cmd, err := unmarshalCommand[*model.SqliNFetch](t.Sqli)
				if err != nil {
					return nil, err
				}
				trans.Append(cmd)
			case 10:
				cmd, err := unmarshalCommand[*model.SqliClose](t.Sqli)
				if err != nil {
					return nil, err
				}
				trans.Append(cmd)
			case 11:
				cmd, err := unmarshalCommand[*model.SqliRelease](t.Sqli)
				if err != nil {
					return nil, err
				}
				trans.Append(cmd)
			case 12:
				cmd, err := unmarshalCommand[*model.SqliEot](t.Sqli)
				if err != nil {
					return nil, err
				}
				trans.Append(cmd)
			case 14:
				cmd, err := unmarshalCommand[*model.SqliTuple](t.Sqli)
				if err != nil {
					return nil, err
				}
				trans.Append(cmd)
			case 15:
				cmd, err := unmarshalCommand[*model.SqliDone](t.Sqli)
				if err != nil {
					return nil, err
				}
				trans.Append(cmd)
			case 22:
				cmd, err := unmarshalCommand[*model.SqliNDescribe](t.Sqli)
				if err != nil {
					return nil, err
				}
				trans.Append(cmd)
			case 36:
				cmd, err := unmarshalCommand[*model.SqliDBOpen](t.Sqli)
				if err != nil {
					return nil, err
				}
				trans.Append(cmd)
			case 49:
				cmd, err := unmarshalCommand[*model.SqliWantDone](t.Sqli)
				if err != nil {
					return nil, err
				}
				trans.Append(cmd)
			case 55:
				cmd, err := unmarshalCommand[*model.SqliCost](t.Sqli)
				if err != nil {
					return nil, err
				}
				trans.Append(cmd)
			case 56:
				cmd, err := unmarshalCommand[*model.SqliExit](t.Sqli)
				if err != nil {
					return nil, err
				}
				trans.Append(cmd)
			case 81:
				cmd, err := unmarshalCommand[*model.SqliInfo](t.Sqli)
				if err != nil {
					return nil, err
				}
				trans.Append(cmd)
			case 100:
				cmd, err := unmarshalCommand[*model.SqliRetType](t.Sqli)
				if err != nil {
					return nil, err
				}
				trans.Append(cmd)
			case 126:
				cmd, err := unmarshalCommand[*model.SqliProtocols](t.Sqli)
				if err != nil {
					return nil, err
				}
				trans.Append(cmd)
			default:
				log.Println(t.Type)
			}
		}

		buf, err := trans.Pack()
		if err != nil {
			return nil, err
		}
		savePackage.Buffer = buf
		savePackage.Length = len(buf)

		result = append(result, savePackage)
	}

	return result, nil
}

func init() {
	rootCmd.AddCommand(toJsonCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// toJsonCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	toJsonCmd.Flags().StringVarP(&inputFile, "input", "i", "", "input file")
	toJsonCmd.MarkFlagRequired("input")
	toJsonCmd.Flags().StringVarP(&outputFile, "output", "o", "", "output file")
	//toJsonCmd.MarkFlagRequired("output")
	//toJsonCmd.Flags().BoolVarP(&color, "color", "c", false, "print with color")
	//toJsonCmd.Flags().BoolVarP(&raw, "raw", "r", false, "show raw")
	//toJsonCmd.Flags().BoolVarP(&printPackage, "parse", "p", false, "parse package")
}
