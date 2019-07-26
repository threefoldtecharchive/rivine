package rivbin

import (
	"fmt"
	"io"
)

// WriteDataSlice writes a length-prefixed byte slice to w.
func WriteDataSlice(w io.Writer, data []byte) error {
	dataLength := len(data)
	err := encodeSliceLength(w, dataLength)
	if err != nil {
		return err
	}
	n, err := w.Write(data)
	if err != nil {
		return err
	}
	if n != dataLength {
		err = io.ErrShortWrite
	}
	return err
}

// ReadDataSlice reads a byte slice length of maximum 4 bytes followed by the number of bytes
// specified in the (length) prefix. The operation is aborted if the prefix exceeds a
// specified maximum length.
func ReadDataSlice(r io.Reader, maxLen int) ([]byte, error) {
	dataLen, err := decodeSliceLength(r)
	if err != nil {
		return nil, err
	}
	if dataLen > maxLen {
		return nil, fmt.Errorf("length %d exceeds maxLen of %d", dataLen, maxLen)
	}
	// read dataLen bytes
	data := make([]byte, dataLen)
	_, err = io.ReadFull(r, data)
	return data, err
}

// WriteObject writes a length-prefixed object to w.
func WriteObject(w io.Writer, v interface{}) error {
	b, err := Marshal(v)
	if err != nil {
		return err
	}
	return WriteDataSlice(w, b)
}

// ReadObject reads and decodes a length-prefixed and marshalled object.
func ReadObject(r io.Reader, obj interface{}, maxLen int) error {
	data, err := ReadDataSlice(r, maxLen)
	if err != nil {
		return err
	}
	return Unmarshal(data, obj)
}
