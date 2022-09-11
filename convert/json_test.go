package convert

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wangxuesong/tcpshadow/model"
)

func TestConvert_SqliCommand_Json(t *testing.T) {
	tests := []struct {
		name       string
		json       []byte
		cmdType    interface{}
		assertFunc func(*testing.T, model.SqliCommand)
	}{{
		name:    "SqliPrepare",
		json:    []byte(`{"type":"SQ_PREPARE","sqli":{"QMarks":0,"Sql":"select * from test;"}}`),
		cmdType: &model.SqliPrepare{},
		assertFunc: func(t *testing.T, cmd model.SqliCommand) {
			prepare, _ := cmd.(*model.SqliPrepare)
			assert.Equal(t, "select * from test;", prepare.Sql)
		},
	}, {
		name:    "SqliEot",
		json:    []byte(`{"type":"SQ_EOT","sqli":{}}`),
		cmdType: &model.SqliEot{},
		assertFunc: func(t *testing.T, cmd model.SqliCommand) {
			_, ok := cmd.(*model.SqliEot)
			require.True(t, ok)
		},
	}}

	for _, test := range tests {
		cmd, err := UnmarshalCommand(test.json)
		require.NoError(t, err)
		assert.IsType(t, test.cmdType, cmd)
		test.assertFunc(t, cmd)

		buf, err := json.Marshal(newJsonSqliCommand(cmd))
		require.NoError(t, err)
		assert.Equal(t, test.json, buf, fmt.Sprintf("Test Failed: %s", test.name))
	}
}
