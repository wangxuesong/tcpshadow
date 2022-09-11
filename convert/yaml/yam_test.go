package yaml

import (
	"fmt"
	yaml2 "gopkg.in/yaml.v3"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wangxuesong/tcpshadow/model"
)

func TestConvert_SqliCommand_Yaml(t *testing.T) {
	tests := []struct {
		name       string
		yaml       []byte
		cmdType    interface{}
		assertFunc func(*testing.T, model.SqliCommand)
	}{{
		name: "SqliPrepare",
		yaml: []byte(`type: SQ_PREPARE
sqli:
    QMarks: 0
    Sql: select * from test;
`),
		cmdType: &model.SqliPrepare{},
		assertFunc: func(t *testing.T, cmd model.SqliCommand) {
			prepare, _ := cmd.(*model.SqliPrepare)
			assert.Equal(t, "select * from test;", prepare.Sql)
		},
	}, {
		name: "SqliEot",
		yaml: []byte(`type: SQ_EOT
sqli: {}
`),
		cmdType: &model.SqliEot{},
		assertFunc: func(t *testing.T, cmd model.SqliCommand) {
			_, ok := cmd.(*model.SqliEot)
			require.True(t, ok)
		},
	}}

	for _, test := range tests {
		cmd, err := UnmarshalCommand(test.yaml)
		require.NoError(t, err)
		assert.IsType(t, test.cmdType, cmd, fmt.Sprintf("Test Failed: %s", test.name))
		test.assertFunc(t, cmd)

		buf, err := yaml2.Marshal(newYamlSqliCommand(cmd))
		require.NoError(t, err)
		assert.Equal(t, test.yaml, buf, fmt.Sprintf("Test Failed: %s", test.name))
	}
}
