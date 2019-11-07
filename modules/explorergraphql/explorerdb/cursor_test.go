package explorerdb

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"
)

func TestCursorCycle(t *testing.T) {
	testCases := []struct {
		Value             interface{}
		ExpectedGQLString string
	}{
		{
			true,
			`c3`,
		},
		{
			42,
			`d3000000000000002a`,
		},
		{
			struct {
				Data interface{}
			}{nil},
			`81a444617461c0`,
		},
		{
			struct {
				Hash StormHash
			}{StormHash{}},
			`81a448617368c4200000000000000000000000000000000000000000000000000000000000000000`,
		},
		{
			StormHash{},
			`c4200000000000000000000000000000000000000000000000000000000000000000`,
		},
		{
			struct {
				Foo  uint64
				Bar  string
				Flag bool
			}{42, "answer", true},
			`83a3466f6fcf000000000000002aa3426172a6616e73776572a4466c6167c3`,
		},
	}
	for idx, testCase := range testCases {
		testCase := testCase
		t.Run(fmt.Sprintf("test #%d", idx+1), func(t *testing.T) {
			// create a new cursor from the given input value (value -> bytes)
			c, err := NewCursor(testCase.Value)
			if err != nil {
				t.Fatal(err)
			}
			// GQL marshal our cursor (bytes -> string -> quoted string)
			buf := bytes.NewBuffer(nil)
			c.MarshalGQL(buf)
			b := buf.Bytes()
			b = b[1 : len(b)-1] // remove quotes, as the GQL library also removes them prior to the decoding process on our side (quoted string -> string)
			s := string(b)
			if s != testCase.ExpectedGQLString {
				t.Errorf("unexpected GQL string: expected %q, not %q", testCase.ExpectedGQLString, s)
			}
			// unmarshal cursor, as string or bytes
			// (even though it seems that the gql lib always decodes our types as strings)
			// (string/bytes (hex) -> cursor (bytes))
			var (
				c1 Cursor
				c2 Cursor
			)
			err = c1.UnmarshalGQL(s)
			if err != nil {
				t.Errorf("failed to unmarshal GQL string as cursor: %v", err)
			} else if !reflect.DeepEqual(c, c1) {
				t.Errorf("unmarshalled GQL string as cursor, but it is unexpected: expected %v, not %v", c, c1)
			}
			err = c2.UnmarshalGQL(b)
			if err != nil {
				t.Errorf("failed to unmarshal GQL bytes as cursor: %v", err)
			} else if !reflect.DeepEqual(c, c2) {
				t.Errorf("unmarshalled GQL bytes as cursor, but it is unexpected: expected %v, not %v", c, c2)
			}
			// unpack value (bytes -> value)
			refValue := reflect.New(reflect.TypeOf(testCase.Value))
			err = c.UnpackValue(refValue.Interface())
			if err != nil {
				t.Errorf("failed to unpack cursor as value: %v", err)
			} else {
				value := refValue.Elem().Interface()
				if !reflect.DeepEqual(testCase.Value, value) {
					t.Errorf("unpacked cursor as value, but it is unexpected: expected %v, not %v", testCase.Value, value)
				}
			}
		})
	}
}
