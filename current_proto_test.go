package main

import (
	"testing"

	"github.com/emicklei/proto"
	"github.com/stretchr/testify/assert"
)

func TestNewProtoMesssageFromMessage(t *testing.T) {
	t.Parallel()

	var testCases = []struct {
		testName string
		given    *proto.Message
		expected *ProtoMessage
	}{
		{
			testName: "empty message",
			given:    &proto.Message{},
			expected: &ProtoMessage{
				currMaxNum: 0,
				fields:     make(map[string]int),
			},
		},
		{
			testName: "populated message - one normal field",
			given: &proto.Message{
				Elements: []proto.Visitee{
					&proto.NormalField{
						Field: &proto.Field{
							Name:     "field1",
							Sequence: 1,
						},
					},
				},
			},
			expected: &ProtoMessage{
				currMaxNum: 1,
				fields: map[string]int{
					"field1": 1,
				},
			},
		},
		{
			testName: "populated message - one map field",
			given: &proto.Message{
				Elements: []proto.Visitee{
					&proto.MapField{
						Field: &proto.Field{
							Name:     "field1",
							Sequence: 1,
						},
					},
				},
			},
			expected: &ProtoMessage{
				currMaxNum: 1,
				fields: map[string]int{
					"field1": 1,
				},
			},
		},
		{
			testName: "populated message - multiple fields",
			given: &proto.Message{
				Elements: []proto.Visitee{
					&proto.NormalField{
						Field: &proto.Field{
							Name:     "field1",
							Sequence: 1,
						},
					},
					&proto.MapField{
						Field: &proto.Field{
							Name:     "field3",
							Sequence: 3,
						},
					},
					&proto.NormalField{
						Field: &proto.Field{
							Name:     "field2",
							Sequence: 2,
						},
					},
				},
			},
			expected: &ProtoMessage{
				currMaxNum: 3,
				fields: map[string]int{
					"field1": 1,
					"field2": 2,
					"field3": 3,
				},
			},
		},
		{
			testName: "populated message - multiple fields; one non-accepted type",
			given: &proto.Message{
				Elements: []proto.Visitee{
					&proto.NormalField{
						Field: &proto.Field{
							Name:     "field1",
							Sequence: 1,
						},
					},
					&proto.EnumField{
						Name:    "zzzzz",
						Integer: 4,
					},
					&proto.MapField{
						Field: &proto.Field{
							Name:     "field3",
							Sequence: 3,
						},
					},
					&proto.NormalField{
						Field: &proto.Field{
							Name:     "field2",
							Sequence: 2,
						},
					},
				},
			},
			expected: &ProtoMessage{
				currMaxNum: 3,
				fields: map[string]int{
					"field1": 1,
					"field2": 2,
					"field3": 3,
				},
			},
		},
	}

	for _, testCase := range testCases {
		result := NewProtoMesssageFromMessage(testCase.given)
		assert.Equal(t, testCase.expected, result, testCase.testName)
	}
}

func TestProtoMessage_GetFieldNum(t *testing.T) {
	t.Parallel()

	var testCases = []struct {
		testName             string
		givenProtoMessage    *ProtoMessage
		givenFieldName       string
		expected             int
		expectedProtoMessage *ProtoMessage
	}{
		{
			testName:          "message fields not instantiated",
			givenProtoMessage: &ProtoMessage{},
			givenFieldName:    "givenFieldName",
			expected:          1,
			expectedProtoMessage: &ProtoMessage{
				currMaxNum: 1,
				fields: map[string]int{
					"givenFieldName": 1,
				},
			},
		},
		{
			testName: "empty fields map",
			givenProtoMessage: &ProtoMessage{
				currMaxNum: 0,
				fields:     map[string]int{},
			},
			givenFieldName: "givenFieldName",
			expected:       1,
			expectedProtoMessage: &ProtoMessage{
				currMaxNum: 1,
				fields: map[string]int{
					"givenFieldName": 1,
				},
			},
		},
		{
			testName: "field exists",
			givenProtoMessage: &ProtoMessage{
				currMaxNum:  5,
				droppedNums: []int{1, 2, 3, 4},
				fields: map[string]int{
					"givenFieldName": 5,
				},
			},
			givenFieldName: "givenFieldName",
			expected:       5,
			expectedProtoMessage: &ProtoMessage{
				currMaxNum:  5,
				droppedNums: []int{1, 2, 3, 4},
				fields: map[string]int{
					"givenFieldName": 5,
				},
			},
		},
		{
			testName: "field DNE - has dropped field",
			givenProtoMessage: &ProtoMessage{
				currMaxNum:  5,
				droppedNums: []int{1, 2, 3, 4},
				fields: map[string]int{
					"givenFieldName2": 5,
				},
			},
			givenFieldName: "givenFieldName",
			expected:       1,
			expectedProtoMessage: &ProtoMessage{
				currMaxNum:  5,
				droppedNums: []int{2, 3, 4},
				fields: map[string]int{
					"givenFieldName":  1,
					"givenFieldName2": 5,
				},
			},
		},
		{
			testName: "field DNE - no dropped",
			givenProtoMessage: &ProtoMessage{
				currMaxNum: 1,
				fields: map[string]int{
					"givenFieldName": 1,
				},
			},
			givenFieldName: "givenFieldName2",
			expected:       2,
			expectedProtoMessage: &ProtoMessage{
				currMaxNum: 2,
				fields: map[string]int{
					"givenFieldName":  1,
					"givenFieldName2": 2,
				},
			},
		},
	}

	for _, testCase := range testCases {
		tmpResultProtoMessage := *testCase.givenProtoMessage
		resultProtoMessage := &tmpResultProtoMessage
		result := resultProtoMessage.GetFieldNum(testCase.givenFieldName)
		assert.Equal(t, testCase.expected, result, testCase.testName)
		assert.Equal(t, testCase.expectedProtoMessage, resultProtoMessage, testCase.testName)
	}
}

func TestProtoMessage_RemoveFieldNum(t *testing.T) {
	t.Parallel()

	var testCases = []struct {
		testName             string
		givenProtoMessage    *ProtoMessage
		givenFieldName       string
		expectedProtoMessage *ProtoMessage
	}{
		{
			testName:             "message fields not instantiated",
			givenProtoMessage:    &ProtoMessage{},
			givenFieldName:       "givenFieldName",
			expectedProtoMessage: &ProtoMessage{},
		},
		{
			testName: "empty fields map",
			givenProtoMessage: &ProtoMessage{
				currMaxNum: 0,
				fields:     map[string]int{},
			},
			givenFieldName: "givenFieldName",
			expectedProtoMessage: &ProtoMessage{
				currMaxNum: 0,
				fields:     map[string]int{},
			},
		},
		{
			testName: "field exists - only element in map - No current dropped",
			givenProtoMessage: &ProtoMessage{
				currMaxNum: 1,
				fields: map[string]int{
					"givenFieldName": 1,
				},
			},
			givenFieldName: "givenFieldName",
			expectedProtoMessage: &ProtoMessage{
				currMaxNum:  1,
				droppedNums: []int{1},
				fields:      map[string]int{},
			},
		},
		{
			testName: "field exists - other elements in map - No current dropped",
			givenProtoMessage: &ProtoMessage{
				currMaxNum: 3,
				fields: map[string]int{
					"givenFieldName":  1,
					"givenFieldName2": 2,
					"givenFieldName3": 3,
				},
			},
			givenFieldName: "givenFieldName2",
			expectedProtoMessage: &ProtoMessage{
				currMaxNum:  3,
				droppedNums: []int{2},
				fields: map[string]int{
					"givenFieldName":  1,
					"givenFieldName3": 3,
				},
			},
		},
		{
			testName: "field exists - only element in map - has current dropped",
			givenProtoMessage: &ProtoMessage{
				currMaxNum:  2,
				droppedNums: []int{1},
				fields: map[string]int{
					"givenFieldName": 2,
				},
			},
			givenFieldName: "givenFieldName",
			expectedProtoMessage: &ProtoMessage{
				currMaxNum:  2,
				droppedNums: []int{1, 2},
				fields:      map[string]int{},
			},
		},
		{
			testName: "field exists - other elements in map - has current dropped",
			givenProtoMessage: &ProtoMessage{
				currMaxNum:  4,
				droppedNums: []int{3},
				fields: map[string]int{
					"givenFieldName":  1,
					"givenFieldName2": 2,
					"givenFieldName4": 4,
				},
			},
			givenFieldName: "givenFieldName2",
			expectedProtoMessage: &ProtoMessage{
				currMaxNum:  4,
				droppedNums: []int{3, 2},
				fields: map[string]int{
					"givenFieldName":  1,
					"givenFieldName4": 4,
				},
			},
		},
	}

	for _, testCase := range testCases {
		tmpResultProtoMessage := *testCase.givenProtoMessage
		resultProtoMessage := &tmpResultProtoMessage
		resultProtoMessage.RemoveFieldNum(testCase.givenFieldName)
		assert.Equal(t, testCase.expectedProtoMessage, resultProtoMessage, testCase.testName)
	}
}

func TestProtoMessageMap_GetFieldNum(t *testing.T) {
	t.Parallel()

	var testCases = []struct {
		testName                string
		givenProtoMessageMap    ProtoMessageMap
		givenMessageName        string
		givenFieldName          string
		expected                int
		expectedProtoMessageMap ProtoMessageMap
	}{
		{
			testName:             "map empty",
			givenProtoMessageMap: make(ProtoMessageMap),
			givenMessageName:     "givenMessageName",
			givenFieldName:       "givenFieldName",
			expected:             1,
			expectedProtoMessageMap: ProtoMessageMap{
				"givenMessageName": &ProtoMessage{
					currMaxNum: 1,
					fields: map[string]int{
						"givenFieldName": 1,
					},
				},
			},
		},
		{
			testName: "map doesn't have message key",
			givenProtoMessageMap: ProtoMessageMap{
				"givenMessageName2": &ProtoMessage{
					currMaxNum: 1,
					fields: map[string]int{
						"givenFieldName": 1,
					},
				},
			},
			givenMessageName: "givenMessageName",
			givenFieldName:   "givenFieldName",
			expected:         1,
			expectedProtoMessageMap: ProtoMessageMap{
				"givenMessageName2": &ProtoMessage{
					currMaxNum: 1,
					fields: map[string]int{
						"givenFieldName": 1,
					},
				},
				"givenMessageName": &ProtoMessage{
					currMaxNum: 1,
					fields: map[string]int{
						"givenFieldName": 1,
					},
				},
			},
		},
		{
			testName: "map doesn't have field key",
			givenProtoMessageMap: ProtoMessageMap{
				"givenMessageName": &ProtoMessage{
					currMaxNum: 1,
					fields: map[string]int{
						"givenFieldName": 1,
					},
				},
			},
			givenMessageName: "givenMessageName",
			givenFieldName:   "givenFieldName2",
			expected:         2,
			expectedProtoMessageMap: ProtoMessageMap{
				"givenMessageName": &ProtoMessage{
					currMaxNum: 2,
					fields: map[string]int{
						"givenFieldName":  1,
						"givenFieldName2": 2,
					},
				},
			},
		},
		{
			testName: "map has field key",
			givenProtoMessageMap: ProtoMessageMap{
				"givenMessageName": &ProtoMessage{
					currMaxNum: 1,
					fields: map[string]int{
						"givenFieldName": 1,
					},
				},
			},
			givenMessageName: "givenMessageName",
			givenFieldName:   "givenFieldName",
			expected:         1,
			expectedProtoMessageMap: ProtoMessageMap{
				"givenMessageName": &ProtoMessage{
					currMaxNum: 1,
					fields: map[string]int{
						"givenFieldName": 1,
					},
				},
			},
		},
	}

	for _, testCase := range testCases {
		resultProtoMessageMap := make(ProtoMessageMap, len(testCase.givenProtoMessageMap))
		for k, _ := range testCase.givenProtoMessageMap {
			resultProtoMessageMap[k] = testCase.givenProtoMessageMap[k]
		}
		result := resultProtoMessageMap.GetFieldNum(testCase.givenMessageName, testCase.givenFieldName)
		assert.Equal(t, testCase.expected, result, testCase.testName)
		assert.Equal(t, testCase.expectedProtoMessageMap, resultProtoMessageMap, testCase.testName)
	}
}

func TestProtoMessageMap_RemoveFieldNum(t *testing.T) {
	t.Parallel()

	var testCases = []struct {
		testName                string
		givenProtoMessageMap    ProtoMessageMap
		givenMessageName        string
		givenFieldName          string
		expectedProtoMessageMap ProtoMessageMap
	}{
		{
			testName:                "map empty",
			givenProtoMessageMap:    make(ProtoMessageMap),
			givenMessageName:        "givenMessageName",
			givenFieldName:          "givenFieldName",
			expectedProtoMessageMap: make(ProtoMessageMap),
		},
		{
			testName: "map doesn't have message key",
			givenProtoMessageMap: ProtoMessageMap{
				"givenMessageName2": &ProtoMessage{
					currMaxNum: 1,
					fields: map[string]int{
						"givenFieldName": 1,
					},
				},
			},
			givenMessageName: "givenMessageName",
			givenFieldName:   "givenFieldName",
			expectedProtoMessageMap: ProtoMessageMap{
				"givenMessageName2": &ProtoMessage{
					currMaxNum: 1,
					fields: map[string]int{
						"givenFieldName": 1,
					},
				},
			},
		},
		{
			testName: "map doesn't have field key",
			givenProtoMessageMap: ProtoMessageMap{
				"givenMessageName": &ProtoMessage{
					currMaxNum: 1,
					fields: map[string]int{
						"givenFieldName": 1,
					},
				},
			},
			givenMessageName: "givenMessageName",
			givenFieldName:   "givenFieldName2",
			expectedProtoMessageMap: ProtoMessageMap{
				"givenMessageName": &ProtoMessage{
					currMaxNum: 1,
					fields: map[string]int{
						"givenFieldName": 1,
					},
				},
			},
		},
		{
			testName: "map has field key - no existing drops - no other elements",
			givenProtoMessageMap: ProtoMessageMap{
				"givenMessageName": &ProtoMessage{
					currMaxNum: 1,
					fields: map[string]int{
						"givenFieldName": 1,
					},
				},
			},
			givenMessageName: "givenMessageName",
			givenFieldName:   "givenFieldName",
			expectedProtoMessageMap: ProtoMessageMap{
				"givenMessageName": &ProtoMessage{
					currMaxNum:  1,
					droppedNums: []int{1},
					fields:      map[string]int{},
				},
			},
		},
		{
			testName: "map has field key - no existing drops - has other elements",
			givenProtoMessageMap: ProtoMessageMap{
				"givenMessageName": &ProtoMessage{
					currMaxNum: 2,
					fields: map[string]int{
						"givenFieldName":  1,
						"givenFieldName2": 2,
					},
				},
			},
			givenMessageName: "givenMessageName",
			givenFieldName:   "givenFieldName",
			expectedProtoMessageMap: ProtoMessageMap{
				"givenMessageName": &ProtoMessage{
					currMaxNum:  2,
					droppedNums: []int{1},
					fields: map[string]int{
						"givenFieldName2": 2,
					},
				},
			},
		},
		{
			testName: "map has field key - has existing drops - no other elements",
			givenProtoMessageMap: ProtoMessageMap{
				"givenMessageName": &ProtoMessage{
					currMaxNum:  2,
					droppedNums: []int{1},
					fields: map[string]int{
						"givenFieldName": 2,
					},
				},
			},
			givenMessageName: "givenMessageName",
			givenFieldName:   "givenFieldName",
			expectedProtoMessageMap: ProtoMessageMap{
				"givenMessageName": &ProtoMessage{
					currMaxNum:  2,
					droppedNums: []int{1, 2},
					fields:      map[string]int{},
				},
			},
		},
		{
			testName: "map has field key - no existing drops - has other elements",
			givenProtoMessageMap: ProtoMessageMap{
				"givenMessageName": &ProtoMessage{
					currMaxNum:  3,
					droppedNums: []int{2},
					fields: map[string]int{
						"givenFieldName":  1,
						"givenFieldName3": 3,
					},
				},
			},
			givenMessageName: "givenMessageName",
			givenFieldName:   "givenFieldName",
			expectedProtoMessageMap: ProtoMessageMap{
				"givenMessageName": &ProtoMessage{
					currMaxNum:  3,
					droppedNums: []int{2, 1},
					fields: map[string]int{
						"givenFieldName3": 3,
					},
				},
			},
		},
	}

	for _, testCase := range testCases {
		resultProtoMessageMap := make(ProtoMessageMap, len(testCase.givenProtoMessageMap))
		for k, _ := range testCase.givenProtoMessageMap {
			resultProtoMessageMap[k] = testCase.givenProtoMessageMap[k]
		}
		resultProtoMessageMap.RemoveFieldNum(testCase.givenMessageName, testCase.givenFieldName)
		assert.Equal(t, testCase.expectedProtoMessageMap, resultProtoMessageMap, testCase.testName)
	}
}
