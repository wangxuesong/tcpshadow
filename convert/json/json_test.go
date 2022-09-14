package json

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
	}, {
		name:    "SqliNDescribe",
		json:    []byte(`{"type":"SQ_NDESCRIBE","sqli":{}}`),
		cmdType: &model.SqliNDescribe{},
		assertFunc: func(t *testing.T, cmd model.SqliCommand) {
			_, ok := cmd.(*model.SqliNDescribe)
			require.True(t, ok)
		},
	}, {
		name:    "SqliWantDone",
		json:    []byte(`{"type":"SQ_WANTDONE","sqli":{}}`),
		cmdType: &model.SqliWantDone{},
		assertFunc: func(t *testing.T, cmd model.SqliCommand) {
			_, ok := cmd.(*model.SqliWantDone)
			require.True(t, ok)
		},
	}, {
		name:    "SqliProtocols",
		json:    []byte(`{"type":"SQ_PROTOCOLS","sqli":{"Protocol":"//x//DyMqpcQ"}}`),
		cmdType: &model.SqliProtocols{},
		assertFunc: func(t *testing.T, cmd model.SqliCommand) {
			protocols, _ := cmd.(*model.SqliProtocols)
			assert.Equal(t, []byte{255, 252, 127, 252, 60, 140, 170, 151, 16}, protocols.Protocol)
		},
	}, {
		name:    "SqliInfo",
		json:    []byte(`{"type":"SQ_INFO","sqli":{"MessageType":6,"Length":38,"InfoEnv":{"NameLength":12,"ValueLength":4,"Env":{"DBTEMP":"/tmp","SUBQCACHESZ":"10"}}}}`),
		cmdType: &model.SqliInfo{},
		assertFunc: func(t *testing.T, cmd model.SqliCommand) {
			sqli, _ := cmd.(*model.SqliInfo)
			assert.Equal(t, int16(38), sqli.Length)
		},
	}, {
		name:    "SqliDBOpen",
		json:    []byte(`{"type":"SQ_DBOPEN","sqli":{"DBName":"dfe","Foo":0}}`),
		cmdType: &model.SqliDBOpen{},
		assertFunc: func(t *testing.T, cmd model.SqliCommand) {
			sqli, _ := cmd.(*model.SqliDBOpen)
			assert.Equal(t, "dfe", sqli.DBName)
		},
	}, {
		name:    "SqliDone",
		json:    []byte(`{"type":"SQ_DONE","sqli":{"Warning":21,"Rows":0,"RowID":0,"SerialID":0}}`),
		cmdType: &model.SqliDone{},
		assertFunc: func(t *testing.T, cmd model.SqliCommand) {
			sqli, _ := cmd.(*model.SqliDone)
			assert.Equal(t, int16(21), sqli.Warning)
		},
	}, {
		name:    "SqliCost",
		json:    []byte(`{"type":"SQ_COST","sqli":{"EstimatedRows":1,"EstimatedIO":1}}`),
		cmdType: &model.SqliCost{},
		assertFunc: func(t *testing.T, cmd model.SqliCommand) {
			sqli, _ := cmd.(*model.SqliCost)
			assert.Equal(t, uint32(1), sqli.EstimatedRows)
		},
	}, {
		name:    "SqliCost",
		json:    []byte(`{"type":"SQ_DESCRIBE","sqli":{"StatementType":2,"StatementID":0,"EstimatedCost":0,"TupleSize":130,"CountOfFields":1,"StringTable":5,"Fields":[{"FieldIndex":0,"ColumnStartPos":0,"ColumnType":13,"ColumnExtendedBuiltinId":0,"OwnerName":0,"ExtendedName":0,"Reference":0,"Alignment":0,"SourceType":0,"Length":128,"Name":"site"}]}}`),
		cmdType: &model.SqliDescribe{},
		assertFunc: func(t *testing.T, cmd model.SqliCommand) {
			sqli, _ := cmd.(*model.SqliDescribe)
			assert.Equal(t, uint16(2), sqli.StatementType)
			assert.Equal(t, uint32(130), sqli.TupleSize)
			assert.Equal(t, 1, len(sqli.Fields))
		},
	}, {
		name:    "SqliExit",
		json:    []byte(`{"type":"SQ_EXIT","sqli":{}}`),
		cmdType: &model.SqliExit{},
		assertFunc: func(t *testing.T, cmd model.SqliCommand) {
			_, ok := cmd.(*model.SqliExit)
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
		assert.Equal(t, string(test.json), string(buf), fmt.Sprintf("Test Failed: %s", test.name))
	}
}

func TestConvert_SavePackage_Json(t *testing.T) {
	tests := []struct {
		name       string
		json       string
		count      int
		assertFunc func(t *testing.T, packages []model.SavePackage, name string)
	}{{
		name:  "empty",
		json:  `{"name":"test"}`,
		count: 2,
		assertFunc: func(t *testing.T, packages []model.SavePackage, _ string) {
			assert.Equal(t, 0, packages[0].Length)
			assert.Equal(t, 0, packages[1].Length)
		},
	}, {
		name: "auth",
		json: `{
  "name": "test",
  "auth": [
    {
      "number": 0,
      "forward": 0,
      "length": 489,
      "buffer": "AekBPAAAAGQAZQAAAD0ABklFRUVNAABsc3FsZXhlYwAAAAAAAAY5LjI4MAAADFJEUyNSMDAwMDAwAAAFc3FsaQAAAAE8AAAAAAAAAAAAAQAJZ2Jhc2VkYnQAACEBAQEBAQEBAWx2WnB4YmdNRnd4OGpycGVlaWNRRVE9PQBvbAAAAAAAAAAAAD10bGl0Y3AAAAAAAAEAaAALAAAAAwAOb2xfZ2Jhc2VkYnRfMQAAAAAAAAAAAAAAAGoABgAHREJQQVRIAAACLgAAB0RCREFURQAABlk0TUQtAAAMSUZYX1VQRERFU0MAAAIxAAARQ0xOVF9QQU1fQ0FQQUJMRQAAAjEAAAhTUUxNT0RFAAAGZ2Jhc2UAAAlOT0RFRkRBQwAAA25vAABrAAAAAAAAAAAAAAABABFNYXJ0aW5Qcm8tODMubGFuAAAAACsvVXNlcnMvbWFydGluL3Byb2plY3RzLzhzcHJvamVjdHMvSkRCQ1Rlc3QAAHQAbwAAAAAAAAAAAGUvVXNlcnMvbWFydGluL3Byb2plY3RzLzhzcHJvamVjdHMvSkRCQ1Rlc3QvbGliL2diYXNlZGJ0amRiY18zLjMuMF8yLmphckNvbm5lY3Rpb25UZXN0L0Nvbm5lY3Rpb25UZXN0AAB/"
    },
    {
      "number": 1,
      "forward": 1,
      "length": 267,
      "buffer": "AQsCPBAAAGQAZQAAAD0ABklFRUVJAABsc3J2aW5meAAAAAAAACNHQmFzZSBTZXJ2ZXIgVmVyc2lvbiA5LjU2LkZDNEcxQUVFAAAjU29mdHdhcmUgU2VyaWFsIE51bWJlciBBQUEjQjAwMDAwMAAADm9sX2diYXNlZGJ0XzEAAAABPAAAAAAAAAAAAAAAAAAAb24AAAAAAAAAAAA9c29jdGNwAAAAAAAAAGYAAAAAAAAAAAAAABQAAABrAAAAAAAAAAwAAAAAAA1kOTAzOGRhNWE4MjQAAAAAAi8AAG4ABAAAAAAAdAAiAAAD6AAAAAAAGC9vcHQvZ2Jhc2U4cy9iaW4vb25pbml0AAB/"
    }
  ]
}`,
		count: 2,
		assertFunc: func(t *testing.T, packages []model.SavePackage, _ string) {
			assert.Equal(t, model.ClientToServer, packages[0].Forward)
			assert.Equal(t, 489, packages[0].Length)
			assert.Equal(t, model.ServerToClient, packages[1].Forward)
			assert.Equal(t, 267, packages[1].Length)
		},
	}, {
		name: "protocol",
		json: `{
  "name": "test",
  "auth": [
    {
      "number": 0,
      "forward": 0,
      "length": 489,
      "buffer": "AekBPAAAAGQAZQAAAD0ABklFRUVNAABsc3FsZXhlYwAAAAAAAAY5LjI4MAAADFJEUyNSMDAwMDAwAAAFc3FsaQAAAAE8AAAAAAAAAAAAAQAJZ2Jhc2VkYnQAACEBAQEBAQEBAWx2WnB4YmdNRnd4OGpycGVlaWNRRVE9PQBvbAAAAAAAAAAAAD10bGl0Y3AAAAAAAAEAaAALAAAAAwAOb2xfZ2Jhc2VkYnRfMQAAAAAAAAAAAAAAAGoABgAHREJQQVRIAAACLgAAB0RCREFURQAABlk0TUQtAAAMSUZYX1VQRERFU0MAAAIxAAARQ0xOVF9QQU1fQ0FQQUJMRQAAAjEAAAhTUUxNT0RFAAAGZ2Jhc2UAAAlOT0RFRkRBQwAAA25vAABrAAAAAAAAAAAAAAABABFNYXJ0aW5Qcm8tODMubGFuAAAAACsvVXNlcnMvbWFydGluL3Byb2plY3RzLzhzcHJvamVjdHMvSkRCQ1Rlc3QAAHQAbwAAAAAAAAAAAGUvVXNlcnMvbWFydGluL3Byb2plY3RzLzhzcHJvamVjdHMvSkRCQ1Rlc3QvbGliL2diYXNlZGJ0amRiY18zLjMuMF8yLmphckNvbm5lY3Rpb25UZXN0L0Nvbm5lY3Rpb25UZXN0AAB/"
    },
    {
      "number": 1,
      "forward": 1,
      "length": 267,
      "buffer": "AQsCPBAAAGQAZQAAAD0ABklFRUVJAABsc3J2aW5meAAAAAAAACNHQmFzZSBTZXJ2ZXIgVmVyc2lvbiA5LjU2LkZDNEcxQUVFAAAjU29mdHdhcmUgU2VyaWFsIE51bWJlciBBQUEjQjAwMDAwMAAADm9sX2diYXNlZGJ0XzEAAAABPAAAAAAAAAAAAAAAAAAAb24AAAAAAAAAAAA9c29jdGNwAAAAAAAAAGYAAAAAAAAAAAAAABQAAABrAAAAAAAAAAwAAAAAAA1kOTAzOGRhNWE4MjQAAAAAAi8AAG4ABAAAAAAAdAAiAAAD6AAAAAAAGC9vcHQvZ2Jhc2U4cy9iaW4vb25pbml0AAB/"
    }
  ],
  "package": [
    {
      "number": 2,
      "forward": 0,
      "command": [
        {
          "type": "SQ_PROTOCOLS",
          "sqli": {
            "Protocol": "//x//DyMqpcQ"
          }
        },
        {
          "type": "SQ_EOT",
          "sqli": {}
        }
      ]
    }
  ]
}`,
		count: 3,
		assertFunc: func(t *testing.T, packages []model.SavePackage, _ string) {
			assert.Equal(t, model.ClientToServer, packages[2].Forward)
			assert.Equal(t, 16, packages[2].Length)
			bytes := []byte{0, 126, 0, 9, 255, 252, 127, 252, 60, 140, 170, 151, 16, 0, 0, 12}
			assert.Equal(t, bytes, packages[2].Buffer)
		},
	}}

	for _, test := range tests {
		pkg, err := UnmarshalSqliSavePackages(test.json)
		require.NoError(t, err)
		assert.IsType(t, []model.SavePackage{}, pkg)
		assert.Equal(t, test.count, len(pkg))
		test.assertFunc(t, pkg, test.name)
	}
}
