package rivbin

import (
	"encoding/hex"
	"reflect"
	"testing"
)

// test to guarantee that issue Rivine/#477 is fixed, and remains fixed
func TestPointerDecodeValue_Issue477(t *testing.T) {
	type _TrippleByte [3]byte

	testCases := []struct {
		HexInput string
		Value    interface{}
	}{
		{`00`, (*bool)(nil)},
		{`00`, (*uint8)(nil)},
		{`00`, (*uint64)(nil)},
		{`00`, (*_TrippleByte)(nil)},
		{`00`, (*string)(nil)},
		{`0100`, (*bool)(nil)},
		{`0101`, (*bool)(nil)},
		{`0102`, (*uint8)(nil)},
		{`014020`, (*uint16)(nil)},
		{`010203040506070809`, (*uint64)(nil)},
		{`01124356`, (*_TrippleByte)(nil)},
	}
	for idx, testCase := range testCases {
		b, err := hex.DecodeString(testCase.HexInput)
		if err != nil {
			t.Error(idx, "hex.DecodeString", err)
			continue
		}

		reference := reflect.New(reflect.TypeOf(testCase.Value))
		err = Unmarshal(b, reference.Interface())
		if err != nil {
			t.Error(idx, "Unmarshal", err)
			continue
		}

		value := reflect.Indirect(reference)
		b = Marshal(value.Interface())
		output := hex.EncodeToString(b)
		if output != testCase.HexInput {
			t.Error(idx, "Marshal", output, "!=", testCase.HexInput)
		}
	}
}
