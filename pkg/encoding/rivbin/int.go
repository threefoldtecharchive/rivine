package rivbin

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
)

// MarshalBool writes a bool value as a single byte value.
func MarshalBool(w io.Writer, b bool) error {
	if b {
		return MarshalUint8(w, 1)
	}
	return MarshalUint8(w, 0)
}

// UnmarshalBool reads a bool value as a single byte value.
func UnmarshalBool(r io.Reader) (bool, error) {
	x, err := UnmarshalUint8(r)
	if err != nil {
		return false, fmt.Errorf("UnmarshalBool: %v", err)
	}
	switch x {
	case 1:
		return true, nil
	case 0:
		return false, nil
	default:
		return false, fmt.Errorf("UnmarshalBool: invalid bool value %v", x)
	}
}

// MarshalUint8 writes an uint8 value as a single byte value.
func MarshalUint8(w io.Writer, x uint8) error {
	n, err := w.Write([]byte{byte(x)})
	if err != nil {
		return err
	}
	if n != 1 {
		return io.ErrShortWrite
	}
	return nil
}

// UnmarshalUint8 reads a single byte and casts its to an uint8 value.
func UnmarshalUint8(r io.Reader) (uint8, error) {
	var b [1]byte
	n, err := r.Read(b[:])
	if err != nil {
		return 0, err
	}
	if n != 1 {
		return 0, io.ErrUnexpectedEOF
	}
	return uint8(b[0]), nil
}

// MarshalUint16 marshals an uint16 value as a 2-byte little-endian value.
func MarshalUint16(w io.Writer, x uint16) error {
	var b [2]byte
	binary.LittleEndian.PutUint16(b[:], x)
	n, err := w.Write(b[:])
	if err != nil {
		return err
	}
	if n != 2 {
		return io.ErrShortWrite
	}
	return nil
}

// UnmarshalUint16 unmarshals a 2-byte little-endian value as an uint16 value.
func UnmarshalUint16(r io.Reader) (uint16, error) {
	var b [2]byte
	n, err := r.Read(b[:])
	if err != nil {
		return 0, err
	}
	if n != 2 {
		return 0, io.ErrUnexpectedEOF
	}
	return binary.LittleEndian.Uint16(b[:]), nil
}

// MarshalUint24 marshals an uint24 value as a 3-byte little-endian value.
func MarshalUint24(w io.Writer, x uint32) error {
	const (
		limit = math.MaxUint32 >> 8
	)
	if x > limit {
		return errors.New("an Uint24 integer cannot be bigger than (2^24)-1")
	}
	var b [4]byte
	binary.LittleEndian.PutUint32(b[:], x)
	n, err := w.Write(b[:3])
	if err != nil {
		return err
	}
	if n != 3 {
		return io.ErrShortWrite
	}
	return nil
}

// UnmarshalUint24 unmarshals a 3-byte little-endian value as an uint32 value.
func UnmarshalUint24(r io.Reader) (uint32, error) {
	var b [4]byte
	n, err := r.Read(b[:3])
	if err != nil {
		return 0, err
	}
	if n != 3 {
		return 0, io.ErrUnexpectedEOF
	}
	return binary.LittleEndian.Uint32(b[:]), nil
}

// MarshalUint32 marshals an uint32 value as a 4-byte little-endian value.
func MarshalUint32(w io.Writer, x uint32) error {
	var b [4]byte
	binary.LittleEndian.PutUint32(b[:], x)
	n, err := w.Write(b[:])
	if err != nil {
		return err
	}
	if n != 4 {
		return io.ErrShortWrite
	}
	return nil
}

// UnmarshalUint32 unmarshals a 4-byte little-endian value as an uint32 value.
func UnmarshalUint32(r io.Reader) (uint32, error) {
	var b [4]byte
	n, err := r.Read(b[:])
	if err != nil {
		return 0, err
	}
	if n != 4 {
		return 0, io.ErrUnexpectedEOF
	}
	return binary.LittleEndian.Uint32(b[:]), nil
}

// MarshalUint64 marshals an uint64 value as an 8-byte little-endian value.
func MarshalUint64(w io.Writer, x uint64) error {
	var b [8]byte
	binary.LittleEndian.PutUint64(b[:], x)
	n, err := w.Write(b[:])
	if err != nil {
		return err
	}
	if n != 8 {
		return io.ErrShortWrite
	}
	return nil
}

// UnmarshalUint64 unmarshals an 8-byte little-endian value as an uint64 value.
func UnmarshalUint64(r io.Reader) (uint64, error) {
	var b [8]byte
	n, err := r.Read(b[:])
	if err != nil {
		return 0, err
	}
	if n != 8 {
		return 0, io.ErrUnexpectedEOF
	}
	return binary.LittleEndian.Uint64(b[:]), nil
}
