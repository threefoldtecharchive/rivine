package types

import (
	"encoding/binary"
	"io"
)

// An encHelper provides convenience methods and reduces allocations during
// encoding. All of its methods become no-ops after the encHelper encounters a
// Write error.
type encHelper struct {
	w   io.Writer
	buf []byte
	err error
}

// reset reslices e's internal buffer to have a length of 0.
func (e *encHelper) reset() {
	e.buf = e.buf[:0]
}

// append appends a byte to e's internal buffer.
func (e *encHelper) append(b byte) {
	if e.err != nil {
		return
	}
	e.buf = append(e.buf, b)
}

// flush writes e's internal buffer to the underlying io.Writer.
func (e *encHelper) flush() (int, error) {
	if e.err != nil {
		return 0, e.err
	}
	n, err := e.w.Write(e.buf)
	if e.err == nil {
		e.err = err
	}
	return n, e.err
}

// Write implements the io.Writer interface.
func (e *encHelper) Write(p []byte) (int, error) {
	if e.err != nil {
		return 0, e.err
	}
	e.buf = append(e.buf[:0], p...)
	return e.flush()
}

// WriteUint64 writes a uint64 value to the underlying io.Writer.
func (e *encHelper) WriteUint64(u uint64) {
	if e.err != nil {
		return
	}
	e.buf = e.buf[:8]
	binary.LittleEndian.PutUint64(e.buf, u)
	e.flush()
}

// WriteUint64 writes an int value to the underlying io.Writer.
func (e *encHelper) WriteInt(i int) {
	e.WriteUint64(uint64(i))
}

// WriteUint64 writes p to the underlying io.Writer, prefixed by its length.
func (e *encHelper) WritePrefix(p []byte) {
	e.WriteInt(len(p))
	e.Write(p)
}

// Err returns the first non-nil error encountered by e.
func (e *encHelper) Err() error {
	return e.err
}

// encoder converts w to an encHelper. If w's underlying type is already
// *encHelper, it is returned; otherwise, a new encHelper is allocated.
func encoder(w io.Writer) *encHelper {
	if e, ok := w.(*encHelper); ok {
		return e
	}
	return &encHelper{
		w:   w,
		buf: make([]byte, 64), // large enough for everything but ArbitraryData
	}
}
