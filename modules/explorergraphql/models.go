package explorergraphql

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"strconv"

	"github.com/99designs/gqlgen/graphql"
	"github.com/threefoldtech/rivine/crypto"
	"github.com/threefoldtech/rivine/types"
)

// Custom Scalar Types

type (
	LockTime    uint64
	ObjectID    []byte
	BinaryData  []byte
	Signature   []byte
	ByteVersion byte
	BigInt      struct {
		*big.Int
	}
)

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
	marshalByteSlice(w, bd[:])
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
func (id ObjectID) MarshalGQL(w io.Writer) {
	marshalByteSlice(w, id[:])
}

// UnmarshalGQL implements the graphql.Unmarshaler interface
func (id *ObjectID) UnmarshalGQL(v interface{}) error {
	b, err := unmarshalByteSlice(v)
	if err != nil {
		return err
	}
	const unlockHashDecodedSize = 39
	switch lb := len(b); lb {
	case crypto.HashSize:
		// nothing to do
	case unlockHashDecodedSize:
		str := hex.EncodeToString(b)
		var uh types.UnlockHash
		err = uh.LoadString(str)
		if err != nil {
			return fmt.Errorf("invalid objectID: it has the size of an unlockhash but is invalid: %v", err)
		}
		// valid
	default:
		return fmt.Errorf("%d is an invalid binary size for an object ID", lb)
	}
	*id = ObjectID(b)
	return nil
}

// MarshalGQL implements the graphql.Marshaler interface
func (sig Signature) MarshalGQL(w io.Writer) {
	marshalByteSlice(w, sig[:])
}

// UnmarshalGQL implements the graphql.Unmarshaler interface
func (sig *Signature) UnmarshalGQL(v interface{}) error {
	b, err := unmarshalByteSlice(v)
	if err != nil {
		return err
	}
	*sig = Signature(b)
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
	switch v := v.(type) {
	case string:
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
	case uint:
		bi.fromInt64(int64(v))
		return nil
	case uint64:
		bi.fromInt64(int64(v))
		return nil
	case int64:
		bi.fromInt64(v)
		return nil
	case int:
		bi.fromInt64(int64(v))
		return nil
	case json.Number:
		x, err := v.Int64()
		if err != nil {
			return err
		}
		bi.fromInt64(x)
		return nil
	default:
		return fmt.Errorf("%T is not a valid BigInt", v)
	}
}

func (bi *BigInt) fromInt64(x int64) {
	bi.Int = big.NewInt(x)
}

// custom third-party (Rivine) scalar types

func MarshalBlockHeight(bh types.BlockHeight) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		io.WriteString(w, strconv.FormatUint(uint64(bh), 10))
	})
}

func UnmarshalBlockHeight(v interface{}) (types.BlockHeight, error) {
	x, err := unmarshalUint64(v)
	if err != nil {
		return 0, err
	}
	return types.BlockHeight(x), nil
}

func MarshalTimestamp(ts types.Timestamp) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		io.WriteString(w, strconv.FormatUint(uint64(ts), 10))
	})
}

func UnmarshalTimestamp(v interface{}) (types.Timestamp, error) {
	x, err := unmarshalUint64(v)
	if err != nil {
		return 0, err
	}
	return types.Timestamp(x), nil
}

func MarshalHash(hash crypto.Hash) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		marshalStringer(w, &hash)
	})
}

func UnmarshalHash(v interface{}) (crypto.Hash, error) {
	var h crypto.Hash
	err := unmarshalStringer(v, &h)
	if err != nil {
		return crypto.Hash{}, err
	}
	return h, nil
}

func MarshalUnlockHash(uh types.UnlockHash) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		marshalStringer(w, &uh)
	})
}

func UnmarshalUnlockHash(v interface{}) (types.UnlockHash, error) {
	var uh types.UnlockHash
	err := unmarshalStringer(v, &uh)
	if err != nil {
		return types.UnlockHash{}, err
	}
	return uh, nil
}

func MarshalPublicKey(pk types.PublicKey) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		marshalStringer(w, &pk)
	})
}

func UnmarshalPublicKey(v interface{}) (types.PublicKey, error) {
	var pk types.PublicKey
	err := unmarshalStringer(v, &pk)
	if err != nil {
		return types.PublicKey{}, err
	}
	return pk, nil
}

// scalar helpers

func unmarshalUint8(v interface{}) (uint8, error) {
	switch v := v.(type) {
	case string:
		x, err := strconv.ParseUint(v, 10, 8)
		return uint8(x), err
	case uint8:
		return v, nil
	case int64:
		if v < 0 || v > 255 {
			return 0, fmt.Errorf("%d is out of range: cannot be casted to an uint8", v)
		}
		return uint8(v), nil
	case int:
		if v < 0 || v > 255 {
			return 0, fmt.Errorf("%d is out of range: cannot be casted to an uint8", v)
		}
		return uint8(v), nil
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
	case int64:
		if v < 0 {
			return 0, fmt.Errorf("%d is out of range: cannot be casted to an uint64 as it is negative", v)
		}
		return uint64(v), nil
	case int:
		if v < 0 {
			return 0, fmt.Errorf("%d is out of range: cannot be casted to an uint64 as it is negative", v)
		}
		return uint64(v), nil
	case json.Number:
		return strconv.ParseUint(string(v), 10, 64)
	default:
		return 0, fmt.Errorf("%T is not an uint", v)
	}
}

type stringer interface {
	String() string
	LoadString(string) error
}

func marshalStringer(w io.Writer, s stringer) {
	graphql.MarshalString(s.String()).MarshalGQL(w)
}

func unmarshalStringer(v interface{}, s stringer) error {
	switch v := v.(type) {
	case string:
		return s.LoadString(v)
	default:
		return fmt.Errorf("%T is not a string", v)
	}
}

func marshalByteSlice(w io.Writer, b []byte) {
	graphql.MarshalString(hex.EncodeToString(b)).MarshalGQL(w)
}

func unmarshalByteSlice(v interface{}) ([]byte, error) {
	switch v := v.(type) {
	case string:
		return hex.DecodeString(v)
	case []byte:
		length := len(v)
		b := make([]byte, length)
		_, err := hex.Decode(b[:], v[:])
		return b, err
	default:
		return nil, fmt.Errorf("%T is not a []byte", v)
	}
}
