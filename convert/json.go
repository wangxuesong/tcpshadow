package convert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"

	"github.com/wangxuesong/tcpshadow/model"
)

type packageForJson struct {
	Name        string               `json:"name"`
	AuthPackage [2]model.SavePackage `json:"auth,omitempty"`
	SqliPackage []sqliPackage        `json:"package,omitempty"`
}

type jsonSqliCommand struct {
	Type              string `json:"type"`
	model.SqliCommand `json:"sqli"`
}

type sqliPackage struct {
	Number  int               `json:"number"`
	Forward model.DataForward `json:"forward"`
	Command []jsonSqliCommand `json:"command"`
}

func MarshalSqliPackage(packages []model.SavePackage) ([]sqliPackage, error) {
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
			cmd := newJsonSqliCommand(sqli)
			output.Command = append(output.Command, cmd)
		}
		sqliPack = append(sqliPack, output)
	}

	return sqliPack, nil
}

func newJsonSqliCommand(sqli model.SqliCommand) jsonSqliCommand {
	return jsonSqliCommand{
		Type:        model.SqliType(sqli.Command()).String(),
		SqliCommand: sqli,
	}
}

func MarshalSqliSavePackages(packages []model.SavePackage) (string, error) {
	pack := packageForJson{Name: "test"}
	pack.AuthPackage[0] = packages[0]
	pack.AuthPackage[1] = packages[1]
	sqliPackages, err := MarshalSqliPackage(packages[2:])
	pack.SqliPackage = sqliPackages

	buf, err := json.MarshalIndent(pack, "", "  ")
	if err != nil {
		return "", err
	}

	return string(buf), nil
}

func unmarshal[T model.SqliCommand](str json.RawMessage) (model.SqliCommand, error) {
	var cmd T
	err := json.Unmarshal(str, &cmd)
	if err != nil {
		return nil, err
	}
	return cmd, nil
}

func UnmarshalSqliSavePackages(jsonString string) ([]model.SavePackage, error) {
	buf := bytes.NewBufferString(jsonString)
	type unmarshalSqliPackage struct {
		Number  int               `json:"number"`
		Forward model.DataForward `json:"forward"`
		Type    string            `json:"type"`
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

	for _, p := range anyJson.SqliPackage {
		savePackage := model.SavePackage{
			Number:  p.Number,
			Forward: p.Forward,
			Length:  0,
			Buffer:  nil,
		}

		var trans model.SqliTransmission
		for _, message := range p.Command {
			command, err := UnmarshalCommand(message)
			if err != nil {
				return nil, err
			}
			trans.Append(command)
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

func UnmarshalCommand(message []byte) (command model.SqliCommand, err error) {
	type commandType struct {
		Type string `json:"type"`
		Sqli json.RawMessage
	}
	var t commandType
	err = json.Unmarshal(message, &t)
	if err != nil {
		return nil, err
	}

	switch model.ParseSqliType(t.Type) {
	case model.SQ_COMMAND:
		cmd, err := unmarshal[*model.SqliCmd](t.Sqli)
		if err != nil {
			return nil, err
		}
		command = cmd
	case 2:
		cmd, err := unmarshal[*model.SqliPrepare](t.Sqli)
		if err != nil {
			return nil, err
		}
		command = cmd
	case 3:
		cmd, err := unmarshal[*model.SqliCurName](t.Sqli)
		if err != nil {
			return nil, err
		}
		command = cmd
	case 4:
		cmd, err := unmarshal[*model.SqliID](t.Sqli)
		if err != nil {
			return nil, err
		}
		command = cmd
	case 6:
		cmd, err := unmarshal[*model.SqliOpen](t.Sqli)
		if err != nil {
			return nil, err
		}
		command = cmd
	case model.SQ_EXECUTE:
		cmd, err := unmarshal[*model.SqliExecute](t.Sqli)
		if err != nil {
			return nil, err
		}
		command = cmd
	case 8:
		cmd, err := unmarshal[*model.SqliDescribe](t.Sqli)
		if err != nil {
			return nil, err
		}
		command = cmd
	case 9:
		cmd, err := unmarshal[*model.SqliNFetch](t.Sqli)
		if err != nil {
			return nil, err
		}
		command = cmd
	case 10:
		cmd, err := unmarshal[*model.SqliClose](t.Sqli)
		if err != nil {
			return nil, err
		}
		command = cmd
	case 11:
		cmd, err := unmarshal[*model.SqliRelease](t.Sqli)
		if err != nil {
			return nil, err
		}
		command = cmd
	case 12:
		cmd, err := unmarshal[*model.SqliEot](t.Sqli)
		if err != nil {
			return nil, err
		}
		command = cmd
	case 14:
		cmd, err := unmarshal[*model.SqliTuple](t.Sqli)
		if err != nil {
			return nil, err
		}
		command = cmd
	case 15:
		cmd, err := unmarshal[*model.SqliDone](t.Sqli)
		if err != nil {
			return nil, err
		}
		command = cmd
	case 22:
		cmd, err := unmarshal[*model.SqliNDescribe](t.Sqli)
		if err != nil {
			return nil, err
		}
		command = cmd
	case 36:
		cmd, err := unmarshal[*model.SqliDBOpen](t.Sqli)
		if err != nil {
			return nil, err
		}
		command = cmd
	case 49:
		cmd, err := unmarshal[*model.SqliWantDone](t.Sqli)
		if err != nil {
			return nil, err
		}
		command = cmd
	case 55:
		cmd, err := unmarshal[*model.SqliCost](t.Sqli)
		if err != nil {
			return nil, err
		}
		command = cmd
	case 56:
		cmd, err := unmarshal[*model.SqliExit](t.Sqli)
		if err != nil {
			return nil, err
		}
		command = cmd
	case 81:
		cmd, err := unmarshal[*model.SqliInfo](t.Sqli)
		if err != nil {
			return nil, err
		}
		command = cmd
	case model.SQ_INSERTDONE:
		cmd, err := unmarshal[*model.SqliInsertDone](t.Sqli)
		if err != nil {
			return nil, err
		}
		command = cmd
	case 100:
		cmd, err := unmarshal[*model.SqliRetType](t.Sqli)
		if err != nil {
			return nil, err
		}
		command = cmd
	case 126:
		cmd, err := unmarshal[*model.SqliProtocols](t.Sqli)
		if err != nil {
			return nil, err
		}
		command = cmd
	default:
		log.Println(t.Type)
		return nil, fmt.Errorf("unknown error type: %s", t.Type)
	}

	return command, nil
}
