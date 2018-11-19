package rivbin

import (
	"bytes"
	"encoding/hex"
	"testing"
)

func TestUint24BinaryEncodingUnmarshalMarshalExample(t *testing.T) {
	const hexStr = `7af905`
	b, err := hex.DecodeString(hexStr)
	if err != nil {
		t.Fatal(err)
	}
	u32, err := UnmarshalUint24(bytes.NewReader(b))
	if err != nil {
		t.Fatal(err)
	}
	buffer := bytes.NewBuffer(nil)
	err = MarshalUint24(buffer, u32)
	if err != nil {
		t.Fatal(err)
	}
	str := hex.EncodeToString(buffer.Bytes())
	if str != hexStr {
		t.Fatal("unexpected hex result", str)
	}
}
