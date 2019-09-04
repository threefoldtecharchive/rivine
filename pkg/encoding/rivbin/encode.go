package rivbin

import (
	"bytes"
	"fmt"
	"io"
	"reflect"

	"github.com/threefoldtech/rivine/build"
)

type (
	// A RivineMarshaler can encode and write itself to a stream.
	RivineMarshaler interface {
		MarshalRivine(io.Writer) error
	}
)

// Marshal returns the encoding of v. For encoding details, see the package
// docstring.
func Marshal(v interface{}) ([]byte, error) {
	b := new(bytes.Buffer)
	err := NewEncoder(b).Encode(v) // no error possible when using a bytes.Buffer
	if err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

// MarshalAll encodes all of its inputs and returns their concatenation.
func MarshalAll(vs ...interface{}) ([]byte, error) {
	b := new(bytes.Buffer)
	enc := NewEncoder(b)
	// Error from EncodeAll is ignored as encoding cannot fail when writing
	// to a bytes.Buffer.
	err := enc.EncodeAll(vs...)
	if err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

// Encoder writes objects to an output stream.
//
// A modified and improved Binary Encoder based upon the siabin Encoder,
// found in the <github.com/threefoldtech/rivine/pkg/encoding/siabin> package.
type Encoder struct {
	w io.Writer
}

// NewEncoder returns a new encoder that writes to w.
func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{w}
}

// Encode writes the encoding of v to the stream. For encoding details, see
// the package docstring.
func (e *Encoder) Encode(v interface{}) error {
	return e.encode(reflect.ValueOf(v))
}

// EncodeAll encodes a variable number of arguments.
func (e *Encoder) EncodeAll(vs ...interface{}) error {
	for _, v := range vs {
		if err := e.Encode(v); err != nil {
			return err
		}
	}
	return nil
}

// write catches instances where short writes do not return an error.
func (e *Encoder) write(p []byte) error {
	n, err := e.w.Write(p)
	if n != len(p) && err == nil {
		return io.ErrShortWrite
	}
	return err
}

// Encode writes the encoding of val to the stream. For encoding details, see
// the package docstring.
func (e *Encoder) encode(val reflect.Value) error {
	// check for RivineMarshaler interface first
	if val.CanInterface() {
		ival := val.Interface()
		if m, ok := ival.(RivineMarshaler); ok {
			return m.MarshalRivine(e.w)
		}
	}

	switch val.Kind() {
	case reflect.Ptr:
		isDefined := !val.IsNil()
		if err := MarshalBool(e.w, isDefined); err != nil || !isDefined {
			return err // nil in case !isDefined
		}
		return e.encode(val.Elem())

	case reflect.Bool:
		return MarshalBool(e.w, val.Bool())

	case reflect.Uint8:
		return MarshalUint8(e.w, uint8(val.Uint()))
	case reflect.Uint32:
		return MarshalUint32(e.w, uint32(val.Uint()))
	case reflect.Int:
		return MarshalUint64(e.w, uint64(val.Int()))
	case reflect.Int64:
		return MarshalUint64(e.w, uint64(val.Int()))
	case reflect.Uint64:
		return MarshalUint64(e.w, val.Uint())
	case reflect.Uint:
		return MarshalUint64(e.w, val.Uint())
	case reflect.Int32:
		return MarshalUint32(e.w, uint32(val.Int()))
	case reflect.Uint16:
		return MarshalUint16(e.w, uint16(val.Uint()))
	case reflect.Int16:
		return MarshalUint16(e.w, uint16(val.Int()))
	case reflect.Int8:
		return MarshalUint8(e.w, uint8(val.Int()))

	case reflect.String: // very similar to a byte slice
		length := val.Len()
		if err := encodeSliceLength(e.w, length); err != nil || length == 0 {
			return err // nil in case length == 0
		}
		return e.write([]byte(val.String()))

	case reflect.Slice:
		// slices are variable length, so prepend the length and then fallthrough to array logic
		length := val.Len()
		if err := encodeSliceLength(e.w, length); err != nil || length == 0 {
			return err // nil in case length == 0
		}
		fallthrough
	case reflect.Array:
		// special case for byte arrays
		if val.Type().Elem().Kind() == reflect.Uint8 {
			// if the array is addressable, we can optimize a bit here
			if val.CanAddr() {
				return e.write(val.Slice(0, val.Len()).Bytes())
			}
			// otherwise we have to copy into a newly allocated slice
			slice := reflect.MakeSlice(reflect.SliceOf(val.Type().Elem()), val.Len(), val.Len())
			reflect.Copy(slice, val)
			return e.write(slice.Bytes())
		}
		// normal slices/arrays are encoded by sequentially encoding their elements
		var err error
		for i := 0; i < val.Len(); i++ {
			err = e.encode(val.Index(i))
			if err != nil {
				return err
			}
		}
		return nil

	case reflect.Struct:
		var err error
		for i := 0; i < val.NumField(); i++ {
			if isFieldHidden(val, i) {
				continue // ignore
			}
			err = e.encode(val.Field(i))
			if err != nil {
				return err
			}
		}
		return nil

	default:
		// Marshalling should never fail. If it panics, you're doing something wrong,
		// like trying to encode a map or an unexported struct field.
		errf := fmt.Errorf("error while trying to marshal unsupported type %s/%s",
			val.Type().String(), val.Kind().String())
		build.Critical(errf)
		return errf
	}
}
