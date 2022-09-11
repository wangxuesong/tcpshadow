package yaml

import (
	"fmt"
	yaml2 "gopkg.in/yaml.v3"
	"log"

	"github.com/wangxuesong/tcpshadow/model"
)

type RawMessage struct {
	unmarshal func(interface{}) error
}

func (msg *RawMessage) UnmarshalYAML(unmarshal func(interface{}) error) error {
	msg.unmarshal = unmarshal
	return nil
}

func (msg *RawMessage) Unmarshal(v interface{}) error {
	return msg.unmarshal(v)
}

type yamlSqliCommand struct {
	Type              string `yaml:"type"`
	model.SqliCommand `yaml:"sqli"`
}

func newYamlSqliCommand(sqli model.SqliCommand) yamlSqliCommand {
	return yamlSqliCommand{
		Type:        model.SqliType(sqli.Command()).String(),
		SqliCommand: sqli,
	}
}

func UnmarshalCommand(yaml []byte) (command model.SqliCommand, err error) {
	type commandType struct {
		Type string     `yaml:"type"`
		Sqli RawMessage `yaml:"sqli"`
	}

	var t commandType
	err = yaml2.Unmarshal(yaml, &t)
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

func unmarshal[T model.SqliCommand](msg RawMessage) (model.SqliCommand, error) {
	var cmd T
	err := msg.Unmarshal(&cmd)
	if err != nil {
		return nil, err
	}
	return cmd, nil
}
