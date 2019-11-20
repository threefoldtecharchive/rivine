package basedb

import (
	"bytes"

	mp "github.com/vmihailenco/msgpack"
)

func MsgpackMarshal(value interface{}) ([]byte, error) {
	var buf bytes.Buffer
	enc := mp.NewEncoder(&buf).UseCompactEncoding(true)
	err := enc.Encode(value)
	return buf.Bytes(), err
}

func MsgpackUnmarshal(b []byte, value interface{}) error {
	dec := mp.NewDecoder(bytes.NewReader(b))
	return dec.Decode(value)
}
