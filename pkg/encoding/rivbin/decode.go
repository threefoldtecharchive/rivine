package rivbin

import (
	"bytes"
	"errors"
	"io"
	"reflect"
)

type (
	// A RivineUnmarshaler can read and decode itself from a stream.
	RivineUnmarshaler interface {
		UnmarshalRivine(io.Reader) error
	}
)

// Unmarshal decodes the encoded value b and stores it in v, which must be a
// pointer. The decoding rules are the inverse of those specified in the
// package docstring for marshaling.
func Unmarshal(b []byte, v interface{}) error {
	r := bytes.NewBuffer(b)
	return NewDecoder(r).Decode(v)
}

// UnmarshalAll decodes the encoded values in b and stores them in vs, which
// must be pointers.
func UnmarshalAll(b []byte, vs ...interface{}) error {
	dec := NewDecoder(bytes.NewBuffer(b))
	return dec.DecodeAll(vs...)
}

// A Decoder reads and decodes values from an input stream.
type Decoder struct {
	r io.Reader
}

// NewDecoder returns a new decoder that reads from r.
func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{r}
}

// Decode reads the next encoded value from its input stream and stores it in
// v, which must be a pointer. The decoding rules are the inverse of those
// specified in the package docstring.
func (d *Decoder) Decode(v interface{}) error {
	// v must be a pointer
	pval := reflect.ValueOf(v)
	if pval.Kind() != reflect.Ptr || pval.IsNil() {
		return errBadPointer
	}

	return d.decode(pval.Elem())
}

var errBadPointer = errors.New("cannot decode into invalid pointer")

// DecodeAll decodes a variable number of arguments.
func (d *Decoder) DecodeAll(vs ...interface{}) error {
	var err error
	for _, v := range vs {
		err = d.Decode(v)
		if err != nil {
			return err
		}
	}
	return nil
}

// decode reads the next encoded value from its input stream and stores it in
// val. The decoding rules are the inverse of those specified in the package
// docstring.
func (d *Decoder) decode(val reflect.Value) error {
	// check for RivineUnmarshaler interface first
	if val.CanAddr() && val.Addr().CanInterface() {
		ival := val.Addr().Interface()
		if u, ok := ival.(RivineUnmarshaler); ok {
			err := u.UnmarshalRivine(d.r)
			return err
		}
	}

	switch val.Kind() {
	case reflect.Ptr:
		isDefined, err := UnmarshalBool(d.r)
		if err != nil || !isDefined {
			return err // nil in case !isDefined
		}
		// make sure we aren't decoding into nil
		if val.IsNil() {
			val.Set(reflect.New(val.Type().Elem()))
		}
		// decode the actual value
		return d.decode(val.Elem())

	case reflect.Bool:
		b, err := UnmarshalBool(d.r)
		if err != nil {
			return err
		}
		val.SetBool(b)
		return nil

	case reflect.Uint8:
		x, err := UnmarshalUint8(d.r)
		if err != nil {
			return err
		}
		val.SetUint(uint64(x))
		return nil

	case reflect.Uint32:
		x, err := UnmarshalUint32(d.r)
		if err != nil {
			return err
		}
		val.SetUint(uint64(x))
		return nil

	case reflect.Int:
		x, err := UnmarshalUint64(d.r)
		if err != nil {
			return err
		}
		val.SetInt(int64(x))
		return nil

	case reflect.Int64:
		x, err := UnmarshalUint64(d.r)
		if err != nil {
			return err
		}
		val.SetInt(int64(x))
		return nil

	case reflect.Uint64:
		x, err := UnmarshalUint64(d.r)
		if err != nil {
			return err
		}
		val.SetUint(x)
		return nil

	case reflect.Uint:
		x, err := UnmarshalUint64(d.r)
		if err != nil {
			return err
		}
		val.SetUint(x)
		return nil

	case reflect.Int32:
		x, err := UnmarshalUint32(d.r)
		if err != nil {
			return err
		}
		val.SetInt(int64(int32(x)))
		return nil

	case reflect.Uint16:
		x, err := UnmarshalUint16(d.r)
		if err != nil {
			return err
		}
		val.SetUint(uint64(x))
		return nil

	case reflect.Int16:
		x, err := UnmarshalUint16(d.r)
		if err != nil {
			return err
		}
		val.SetInt(int64(int16(x)))
		return nil

	case reflect.Int8:
		x, err := UnmarshalUint8(d.r)
		if err != nil {
			return err
		}
		val.SetInt(int64(int8(x)))
		return nil

	case reflect.String: // very similar to byte slices
		strLen, err := decodeSliceLength(d.r) // length is capped by the decodeSliceLength Func
		if err != nil {
			return err
		}
		b, err := d.readN(strLen)
		if err != nil {
			return err
		}
		val.SetString(string(b))
		return nil

	case reflect.Slice:
		// slices are variable length, but otherwise the same as arrays.
		// just have to allocate them first, then we can fallthrough to the array logic.
		sliceLen, err := decodeSliceLength(d.r) // length is capped by the decodeSliceLength Func
		if err != nil || sliceLen == 0 {
			return err // nil in case sliceLen==0
		}
		val.Set(reflect.MakeSlice(val.Type(), sliceLen, sliceLen))
		fallthrough
	case reflect.Array:
		// special case for byte arrays (e.g. hashes)
		if val.Type().Elem().Kind() == reflect.Uint8 {
			// convert val to a slice and read into it directly
			b := val.Slice(0, val.Len())
			_, err := io.ReadFull(d.r, b.Bytes())
			if err != nil {
				return err
			}
			return nil
		}
		// arrays are unmarshalled by sequentially unmarshalling their elements
		var err error
		for i := 0; i < val.Len(); i++ {
			err = d.decode(val.Index(i))
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
			err = d.decode(val.Field(i))
			if err != nil {
				return err
			}
		}
		return nil

	default:
		return errors.New("unknown type")
	}
}

// readN reads n bytes and panics if the read fails.
func (d *Decoder) readN(n int) ([]byte, error) {
	if buf, ok := d.r.(*bytes.Buffer); ok {
		b := buf.Next(n)
		if len(b) != n {
			return nil, io.ErrUnexpectedEOF
		}
		return b, nil
	}
	b := make([]byte, n)
	_, err := io.ReadFull(d.r, b)
	if err != nil {
		return nil, err
	}
	return b, nil
}
