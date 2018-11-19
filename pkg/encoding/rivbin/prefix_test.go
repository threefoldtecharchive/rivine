package rivbin

import (
	"bytes"
	"io"
	"testing"
)

// badReader/Writer used to test error handling

type badReader struct{}

func (br *badReader) Read([]byte) (int, error) { return 0, io.EOF }

type badWriter struct{}

func (bw *badWriter) Write([]byte) (int, error) { return 0, nil }

func TestReadDataSlice(t *testing.T) {
	b := new(bytes.Buffer)

	// standard
	encodeSliceLength(b, 3)
	b.Write([]byte("foo"))
	data, err := ReadDataSlice(b, 3)
	if err != nil {
		t.Error(err)
	} else if string(data) != "foo" {
		t.Errorf("expected foo, got %s", data)
	}

	// 0-length
	encodeSliceLength(b, 0)
	_, err = ReadDataSlice(b, 0)
	if err != nil {
		t.Error(err)
	}

	// empty
	b.Write([]byte{})
	_, err = ReadDataSlice(b, 3)
	if err != io.EOF {
		t.Error("expected EOF, got", err)
	}

	// less than 4 bytes
	encodeSliceLength(b, 3)
	b.Write([]byte{1, 2})
	_, err = ReadDataSlice(b, 3)
	if err != io.ErrUnexpectedEOF {
		t.Error("expected unexpected EOF, got", err)
	}

	// exceed maxLen
	encodeSliceLength(b, 3)
	_, err = ReadDataSlice(b, 3)
	if err == nil {
		t.Error("expected maxLen error, got", err)
	}

	// no data after length prefix
	encodeSliceLength(b, 3)
	_, err = ReadDataSlice(b, 3)
	if err != io.EOF {
		t.Error("expected EOF, got", err)
	}
}

func TestReadObject(t *testing.T) {
	b := new(bytes.Buffer)
	var obj string

	// standard
	encodeSliceLength(b, 4)
	encodeSliceLength(b, 3)
	b.Write([]byte("foo"))
	err := ReadObject(b, &obj, 11)
	if err != nil {
		t.Error(err)
	} else if obj != "foo" {
		t.Errorf("expected foo, got %s", obj)
	}

	// empty
	b.Write([]byte{})
	err = ReadObject(b, &obj, 0)
	if err != io.EOF {
		t.Error("expected EOF, got", err)
	}
}

func TestWritePrefix(t *testing.T) {
	b := new(bytes.Buffer)
	expectedBuffer := new(bytes.Buffer)

	// standard
	err := WriteDataSlice(b, []byte("foo"))
	encodeSliceLength(expectedBuffer, 3)
	expectedBuffer.Write([]byte("foo"))
	if err != nil {
		t.Error(err)
	} else if !bytes.Equal(b.Bytes(), expectedBuffer.Bytes()) {
		t.Errorf("WritePrefix wrote wrong data: expected %v, got %v",
			b.Bytes(), expectedBuffer.Bytes())
	}

	// badWriter (returns nil error, but doesn't write anything)
	bw := new(badWriter)
	err = WriteDataSlice(bw, []byte("foo"))
	if err != io.ErrShortWrite {
		t.Error("expected ErrShortWrite, got", err)
	}
}

func TestWriteObject(t *testing.T) {
	b := new(bytes.Buffer)
	expectedBuffer := new(bytes.Buffer)

	// standard
	err := WriteObject(b, "foo")
	encodeSliceLength(expectedBuffer, 4)
	encodeSliceLength(expectedBuffer, 3)
	expectedBuffer.Write([]byte("foo"))
	if err != nil {
		t.Error(err)
	} else if !bytes.Equal(b.Bytes(), expectedBuffer.Bytes()) {
		t.Errorf("WritePrefix wrote wrong data: expected %v, got %v",
			b.Bytes(), expectedBuffer.Bytes())
	}

	// badWriter
	bw := new(badWriter)
	err = WriteObject(bw, "foo")
	if err != io.ErrShortWrite {
		t.Error("expected ErrShortWrite, got", err)
	}
}

func TestReadWritePrefix(t *testing.T) {
	b := new(bytes.Buffer)

	// WritePrefix -> ReadPrefix
	data := []byte("foo")
	err := WriteDataSlice(b, data)
	if err != nil {
		t.Fatal(err)
	}
	rdata, err := ReadDataSlice(b, 100)
	if err != nil {
		t.Error(err)
	} else if !bytes.Equal(rdata, data) {
		t.Errorf("read/write mismatch: wrote %s, read %s", data, rdata)
	}

	// WriteObject -> ReadObject
	obj := "bar"
	err = WriteObject(b, obj)
	if err != nil {
		t.Fatal(err)
	}
	var robj string
	err = ReadObject(b, &robj, 100)
	if err != nil {
		t.Error(err)
	} else if robj != obj {
		t.Errorf("read/write mismatch: wrote %s, read %s", obj, robj)
	}
}
