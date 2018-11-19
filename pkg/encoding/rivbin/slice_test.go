package rivbin

import (
	"bytes"
	"io"
	"math"
	"reflect"
	"testing"
)

func TestMarshalUnmarshalTinySlices(t *testing.T) {
	// utility funcs
	/*hbs := func(str string) []byte { // hexStr -> byte slice
		bs, _ := hex.DecodeString(str)
		return bs
	}hs := func(str string) (hash crypto.Hash) { // hbs -> crypto.Hash
		copy(hash[:], hbs(str))
		return
	}*/

	testCases := []struct {
		Value interface{}
		Valid bool
	}{
		// test nil slices (including an empty string)
		{
			"",
			true,
		},
		{
			[]byte{},
			true,
		},
		{
			[]interface{}{},
			true,
		},
		// test all strings
		{
			"Hello, World!",
			true,
		},
		{
			"读万卷书不如行万里路。",
			true,
		},
		{
			"abcdefghijklmnopqrstuvwxyz01234567890ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz" +
				"01234567890ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz01234567890ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
				"abcdefghijklmnopqrstuvwxyz01234567890ABCDEFGHIJKLMNOPQRSTUVWXYZabc",
			true,
		},
		{
			"abcdefghijklmnopqrstuvwxyz01234567890ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz" +
				"01234567890ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz01234567890ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
				"abcdefghijklmnopqrstuvwxyz01234567890ABCDEFGHIJKLMNOPQRSTUVWXYZabcd",
			false, // overflow, len(x)>255
		},
		// test primitive slices
		{
			[]bool{true, false, true, true, false, false, true, true, true, false, true, false},
			true,
		},
		{
			make([]bool, 256, 256),
			false,
		},
		{
			[]byte{0, 1, 2, 'a', 'b'},
			true,
		},
		{
			make([]byte, 256, 256),
			false,
		},
		{
			[]uint64{0, math.MaxUint64, 42, 1000},
			true,
		},
	}
	for idx, testCase := range testCases {
		b := bytes.NewBuffer(nil)

		// marshal
		err := MarshalTinySlice(b, testCase.Value)
		if !testCase.Valid {
			if err == nil {
				t.Errorf("expected error for testCase %d ('%v'), but received none", idx, testCase.Value)
			}
			continue // continue either way, as we do not expected it to be valid
		}
		if err != nil {
			t.Errorf("unexpected Marshal error for testCase %d ('%v'): %v", idx, testCase.Value, err)
			continue
		}

		// write extra content, as to ensure nothing more is read, than should be read
		n, err := b.Write([]byte("test"))
		if err != nil {
			t.Fatal(idx, err)
		}
		if n != 4 {
			t.Fatal(idx, io.ErrShortWrite)
		}

		// allocate new value
		val := reflect.ValueOf(testCase.Value)
		switch k := val.Kind(); k {
		case reflect.Slice:
			l := val.Len()
			sv := reflect.MakeSlice(val.Type(), l, l)
			val = reflect.New(sv.Type())
			val.Elem().Set(sv)
		case reflect.String:
			val = reflect.New(val.Type())
		default:
			t.Errorf("UnmarshalTinySlice: non-slice type %s (kind: %s) is not supported",
				val.Type().String(), k.String())
			continue
		}

		// unmarshal (valid values only)
		err = UnmarshalTinySlice(b, val.Interface())
		if err != nil {
			t.Errorf("unexpected Unmarshal error for testCase %d (iptrVal: %s/%s) ('%v'): %v",
				idx, val.Type().String(), val.Kind().String(), testCase.Value, err)
			continue
		}

		// compare unmarshalled value with original value
		if ival := val.Elem().Interface(); !reflect.DeepEqual(testCase.Value, ival) {
			t.Errorf("unexpected unmarshalled value: '%v' != '%v'", testCase.Value, ival)
		}

		// ensure that the unmarshal process has read the exact amount of bytes
		remaining := string(b.Bytes())
		if remaining != "test" {
			t.Error(idx, "read more or less than it should have, unexpected remaining:", remaining, "!= test")
		}
	}
}

func TestMarshalUnmarshalEmptyTinyStringSlice(t *testing.T) {
	v := "foo"
	err := UnmarshalTinySlice(bytes.NewBuffer([]byte{0}), &v)
	if err != nil {
		t.Fatal(err)
	}
	if v != "" {
		t.Fatal("unexpected value of 'v', it should be empty due to a nil-reset, but it is instead:", v)
	}
}

func TestMarshalUnmarshalEmptyTinySlice(t *testing.T) {
	v := []int{4, 2}
	err := UnmarshalTinySlice(bytes.NewBuffer([]byte{0}), &v)
	if err != nil {
		t.Fatal(err)
	}
	if len(v) != 0 || cap(v) != 0 {
		t.Fatal("unexpected value of 'v', it should be empty (and have 0 capacity) due to a nil-reset, but it is instead:", v)
	}
}

func TestInvalidUnmarshalTinySlicesWithInvalidInput(t *testing.T) {
	var bs []bool
	err := UnmarshalTinySlice(bytes.NewBuffer([]byte{3, 1, 1}), &bs)
	if err == nil {
		t.Errorf("unexpected []bool unmarshal into: %v", bs)
	}

	var uint16s []uint16
	err = UnmarshalTinySlice(bytes.NewBuffer([]byte{1, 1}), &uint16s)
	if err == nil {
		t.Errorf("unexpected []uint16 unmarshal into: %v", uint16s)
	}
	err = UnmarshalTinySlice(bytes.NewBuffer([]byte{2, 1, 1, 1}), &uint16s)
	if err == nil {
		t.Errorf("unexpected []uint16 unmarshal into: %v", uint16s)
	}
}

func TestValidUnmarshalTinySliceWithOptimizedEncodedIntegerValues(t *testing.T) {
	var uint32s []uint32
	err := UnmarshalTinySlice(bytes.NewBuffer([]byte{2, 1, 1, 1, 1, 1, 1, 1, 1}), &uint32s)
	if err != nil {
		t.Errorf("unexpected []uint32 unmarshal error: %v", err)
	}

	var uint16s []uint16
	err = UnmarshalTinySlice(bytes.NewBuffer([]byte{2, 1, 1, 1, 1}), &uint16s)
	if err != nil {
		t.Errorf("unexpected []uint16 unmarshal error: %v", err)
	}

	var uint8s []uint8
	err = UnmarshalTinySlice(bytes.NewBuffer([]byte{2, 1, 1}), &uint8s)
	if err != nil {
		t.Errorf("unexpected []uint8 unmarshal error: %v", err)
	}
}

func TestEncodeDecodeSliceLength(t *testing.T) {
	testCases := []struct {
		Value      int
		ByteLength int
	}{
		// one byte
		{0, 1},
		{1, 1},
		{42, 1},
		{(1 << 2), 1},
		{(1 << 5), 1},
		{(1 << 6), 1},
		// two bytes
		{(1 << 7), 2},
		{(1 << 8), 2},
		{15999, 2},
		{(1 << 12), 2},
		{(1 << 14) - 1, 2},
		// three bytes
		{(1 << 14), 3},
		{(1 << 15), 3},
		{(1 << 18), 3},
		{2000000, 3},
		{(1 << 19), 3},
		{(1 << 21) - 1, 3},
		// four bytes
		{(1 << 21), 4},
		{(1 << 22), 4},
		{(1 << 24), 4},
		{(1 << 25), 4},
		{(1 << 27), 4},
		{(1 << 28) - 1, 4},
		{(1 << 29) - 1, 4},
	}
	for idx, testCase := range testCases {
		buffer := bytes.NewBuffer(nil)
		err := encodeSliceLength(buffer, testCase.Value)
		if err != nil {
			t.Error(idx+1, "failed to encode slice length", testCase.Value, "error:", err)
			continue
		}
		if buffer.Len() != testCase.ByteLength {
			t.Error(idx+1, testCase.Value, "unexpected encoded byte length", buffer.Len(), "!=", testCase.ByteLength)
			// keep testing this testCase, as it doesn't have to mean the other test(s) fail as well
		}
		value, err := decodeSliceLength(buffer)
		if err != nil {
			t.Error(idx+1, "failed to decode slice length", testCase.Value, "error:", err)
			continue
		}
		if value != testCase.Value {
			t.Error(idx+1, "unexpected decoded slice length", value, "!=", testCase.Value)
		}
	}
}
