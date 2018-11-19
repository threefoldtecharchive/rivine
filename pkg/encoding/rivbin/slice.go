package rivbin

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"reflect"
)

const (
	// MaxSliceSize refers to the maximum size slice could have. Limited
	// to 5 MB.
	MaxSliceSize = 5e6 // 5 MB
)

var (
	// ErrSliceTooLarge is an error when encoded slice is too large.
	ErrSliceTooLarge = errors.New("encoded slice is too large")
)

// MarshalTinySlice allows the marshaling of tiny slices,
// meaning slices with a length of maximum 255 (elements).
// All elements are encoded in the order given, using the encoder of this package.
//
// Note that this specialised marshal function should only be used when
// also in complete control of the unmarshal process, otherwise it is better
// to use the regular dynamlic slice encoding, using the normal encoding process of this package.
//
// Supported types: `[]byte`, `[]x` (where x can be of any type), `string`
func MarshalTinySlice(w io.Writer, v interface{}) error {
	val := reflect.ValueOf(v)
	switch k := val.Kind(); k {
	case reflect.Slice:
		l := val.Len()
		if l > math.MaxUint8 {
			return fmt.Errorf("a tiny slice can have a maximum of %d elements", math.MaxUint8)
		}
		// slices are variable length, so prepend the length
		err := MarshalUint8(w, uint8(l))
		if err != nil {
			return err
		}
		if l == 0 {
			// if length is 0, we are done
			return nil
		}
		// special case for byte slices
		if val.Type().Elem().Kind() == reflect.Uint8 {
			// if the array is addressable, we can optimize a bit here
			if val.CanAddr() {
				return marshalBytes(w, val.Slice(0, val.Len()).Bytes())
			}
			// otherwise we have to copy into a newly allocated slice
			slice := reflect.MakeSlice(reflect.SliceOf(val.Type().Elem()), val.Len(), val.Len())
			reflect.Copy(slice, val)
			return marshalBytes(w, slice.Bytes())
		}
		// create an encoder and encode all slice elements in the regular Sia-encoding way
		e := NewEncoder(w)
		// normal slices are encoded by sequentially encoding their elements
		for i := 0; i < l; i++ {
			err = e.Encode(val.Index(i).Interface())
			if err != nil {
				return err
			}
		}
		return nil

	case reflect.String:
		return MarshalTinySlice(w, []byte(val.String()))

	default:
		return fmt.Errorf("MarshalTinySlice: non-slice type %s (kind: %s) is not supported",
			val.Type().String(), k.String())
	}
}

// UnmarshalTinySlice allows the unmarshaling of tiny slices,
// meaning slices with a length of maximum 255 (elements).
// All elements are decoded in the order given, using the decoder of this package.
// Note that this can only decode values that were encoded
// using an equivalent algorithm as implemented in `MarshalTinySlice`.
//
// Supported types: `*[]byte`, `*[]x` (where x can be of any type), `*string`
func UnmarshalTinySlice(r io.Reader, v interface{}) error {
	// v must be a pointer
	pval := reflect.ValueOf(v)
	if pval.Kind() != reflect.Ptr || pval.IsNil() {
		return errors.New("cannot unmarshal tiny slice into invalid pointer")
	}
	val := pval.Elem()
	switch k := val.Kind(); k {
	case reflect.Slice:
		// slices are variable length, have to allocate them first,
		// for that we need to read the 1 byte length prefix
		sliceLen, err := UnmarshalUint8(r)
		if err != nil {
			return err
		}

		// sanity-check the sliceLen, otherwise you can crash a peer by making
		// them allocate a massive slice
		if sliceLen > math.MaxUint8 || uint64(sliceLen)*uint64(val.Type().Elem().Size()) > MaxSliceSize {
			return ErrSliceTooLarge
		}

		if sliceLen == 0 {
			val.Set(reflect.MakeSlice(val.Type(), 0, 0))
			return nil
		}
		val.Set(reflect.MakeSlice(val.Type(), int(sliceLen), int(sliceLen)))

		// special case for byte slices
		if val.Type().Elem().Kind() == reflect.Uint8 {
			// convert val to a slice and read into it directly
			b := val.Slice(0, val.Len())
			_, err := io.ReadFull(r, b.Bytes()) // n (1st return param) is already checked by io.ReadFull
			return err
		}
		// create regular sia decoder for the last part
		d := NewDecoder(r)
		// slices are unmarshalled by sequentially unmarshalling their elements
		for i := 0; i < val.Len(); i++ {
			err := d.Decode(val.Index(i).Addr().Interface())
			if err != nil {
				return fmt.Errorf("UnmarshalTinySlice failed to unmarshal element %d: %v", i, err)
			}
		}
		return nil

	case reflect.String:
		var b []byte
		err := UnmarshalTinySlice(r, &b)
		if err != nil {
			return err
		}
		val.SetString(string(b))
		return nil

	default:
		return fmt.Errorf("UnmarshalTinySlice: non-slice type %s (kind: %s) is not supported",
			val.Type().String(), k.String())
	}
}

func encodeSliceLength(w io.Writer, length int) error {
	const (
		inclusiveUpperLimitOneByte   = math.MaxUint8 >> 1
		inclusiveUpperLimitTwoBytes  = math.MaxUint16 >> 2
		inclusiveUpperLimithreeBytes = math.MaxUint32 >> 11
		inclusiveUpperLimitFourBytes = math.MaxUint32 >> 3
	)
	switch {
	// 0xxx xxxx
	case length <= inclusiveUpperLimitOneByte:
		return MarshalUint8(w, uint8(length<<1)) // &0b0
	// 10xx xxxx xxxx xxxx
	case length <= inclusiveUpperLimitTwoBytes:
		return MarshalUint16(w, uint16(1)|uint16(length<<2)) // &0b10
	// 110x xxxx xxxx xxxx xxxx xxxx
	case length <= inclusiveUpperLimithreeBytes:
		return MarshalUint24(w, uint32(3)|uint32(length<<3)) // &0b110
	// 111x xxxx xxxx xxxx xxxx xxxx xxxx xxxx
	case length <= inclusiveUpperLimitFourBytes:
		return MarshalUint32(w, uint32(7)|uint32(length<<3)) // &0b111
	// error: overflow
	default:
		return fmt.Errorf(
			"slice length encode overflow: a length of %d is the maximum supported sice length",
			inclusiveUpperLimitFourBytes)
	}
}

func decodeSliceLength(r io.Reader) (int, error) {
	b := make([]byte, 1)
	_, err := io.ReadFull(r, b[:])
	if err != nil {
		return 0, err
	}
	switch {
	// 0xxx xxxx
	case b[0]&1 == 0:
		return int(b[0] >> 1), nil
	// 10xx xxxx xxxx xxxx
	case b[0]&3 == 1:
		b = append(b, 0)
		if _, err = io.ReadFull(r, b[1:2]); err != nil {
			return 0, err
		}
		return int(binary.LittleEndian.Uint16(b[:]) >> 2), nil
	// 110x xxxx xxxx xxxx xxxx xxxx
	case b[0]&7 == 3:
		b = append(b, 0, 0, 0)
		if _, err = io.ReadFull(r, b[1:3]); err != nil {
			return 0, err
		}
		return int(binary.LittleEndian.Uint32(b[:]) >> 3), nil
	// 111x xxxx xxxx xxxx xxxx xxxx xxxx xxxx
	case b[0]&7 == 7:
		b = append(b, 0, 0, 0)
		if _, err = io.ReadFull(r, b[1:4]); err != nil {
			return 0, err
		}
		return int(binary.LittleEndian.Uint32(b[:]) >> 3), nil
	// error: invalid prefix byte (must mean the combo of first 3 bits must be one of {0b001, 0b010, 0b011})
	default:
		return 0, err
	}
}

func marshalBytes(w io.Writer, p []byte) error {
	n, err := w.Write(p)
	if n != len(p) && err == nil {
		return io.ErrShortWrite
	}
	return err
}
