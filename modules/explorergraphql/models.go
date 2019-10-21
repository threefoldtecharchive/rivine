package explorergraphql

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"strconv"

	"github.com/99designs/gqlgen/graphql"
)

type (
	ReferencePoint uint64
	LockTime       uint64
	BinaryData     []byte
	ByteVersion    byte
	BigInt         struct {
		*big.Int
	}
)

// MarshalGQL implements the graphql.Marshaler interface
func (rp ReferencePoint) MarshalGQL(w io.Writer) {
	io.WriteString(w, strconv.FormatUint(uint64(rp), 10))
}

// UnmarshalGQL implements the graphql.Unmarshaler interface
func (rp *ReferencePoint) UnmarshalGQL(v interface{}) error {
	x, err := unmarshalUint64(v)
	if err != nil {
		return err
	}
	*rp = ReferencePoint(x)
	return nil
}

// MarshalGQL implements the graphql.Marshaler interface
func (lt LockTime) MarshalGQL(w io.Writer) {
	io.WriteString(w, strconv.FormatUint(uint64(lt), 10))
}

// UnmarshalGQL implements the graphql.Unmarshaler interface
func (lt *LockTime) UnmarshalGQL(v interface{}) error {
	x, err := unmarshalUint64(v)
	if err != nil {
		return err
	}
	*lt = LockTime(x)
	return nil
}

// MarshalGQL implements the graphql.Marshaler interface
func (bd BinaryData) MarshalGQL(w io.Writer) {
	io.WriteString(w, base64.StdEncoding.EncodeToString([]byte(bd)))
}

// UnmarshalGQL implements the graphql.Unmarshaler interface
func (bd *BinaryData) UnmarshalGQL(v interface{}) error {
	b, err := unmarshalByteSlice(v)
	if err != nil {
		return err
	}
	*bd = BinaryData(b)
	return nil
}

// MarshalGQL implements the graphql.Marshaler interface
func (bv ByteVersion) MarshalGQL(w io.Writer) {
	io.WriteString(w, strconv.FormatUint(uint64(bv), 10))
}

// UnmarshalGQL implements the graphql.Unmarshaler interface
func (bv *ByteVersion) UnmarshalGQL(v interface{}) error {
	x, err := unmarshalUint8(v)
	if err != nil {
		return err
	}
	*bv = ByteVersion(x)
	return nil
}

// MarshalGQL implements the graphql.Marshaler interface
func (bi BigInt) MarshalGQL(w io.Writer) {
	if bi.Int == nil {
		io.WriteString(w, new(big.Int).String())
	}
	io.WriteString(w, bi.String())
}

// UnmarshalGQL implements the graphql.Unmarshaler interface
func (bi *BigInt) UnmarshalGQL(v interface{}) error {
	s, err := graphql.UnmarshalString(v)
	if err != nil {
		return err
	}
	bi.Int = new(big.Int)
	_, ok := bi.SetString(s, 10)
	if !ok {
		return fmt.Errorf("failed to convert %v (as str: '%s') to *big.Int", v, s)
	}
	return nil
}

func unmarshalUint8(v interface{}) (uint8, error) {
	switch v := v.(type) {
	case string:
		x, err := strconv.ParseUint(v, 10, 8)
		return uint8(x), err
	case uint8:
		return v, nil
	case json.Number:
		x, err := strconv.ParseUint(string(v), 10, 8)
		return uint8(x), err
	default:
		return 0, fmt.Errorf("%T is not an uint8", v)
	}
}

func unmarshalUint64(v interface{}) (uint64, error) {
	switch v := v.(type) {
	case string:
		return strconv.ParseUint(v, 10, 64)
	case uint:
		return uint64(v), nil
	case uint64:
		return v, nil
	case json.Number:
		return strconv.ParseUint(string(v), 10, 64)
	default:
		return 0, fmt.Errorf("%T is not an uint", v)
	}
}

func unmarshalByteSlice(v interface{}) ([]byte, error) {
	switch v := v.(type) {
	case string:
		return base64.RawStdEncoding.DecodeString(v)
	case []byte:
		return v, nil
	default:
		return nil, fmt.Errorf("%T is not a []byte", v)
	}
}
