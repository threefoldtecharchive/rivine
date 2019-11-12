package explorerdb

import (
	"encoding/hex"
	"fmt"
	"io"

	"github.com/99designs/gqlgen/graphql"
)

// Cursor is the hex-encoding distribution format
// of a MsgPack-encoded opaque cursor value.
// It contains all the data a query operation requires
// to continue a master query issued by the user
// that has too many results to be returned at once.
type Cursor struct {
	data []byte
}

// NewCursor creates a cursor from a MsgPack-encodable value.
// See `Cursor` for more information.
func NewCursor(value interface{}) (Cursor, error) {
	b, err := msgpackMarshal(value)
	if err != nil {
		return Cursor{}, fmt.Errorf("failed to create a new cursor: failed to MsgPack encode the given value: %v", err)
	}
	return Cursor{data: b}, nil
}

// UnpackValue allows you to MsgPack-decode the cursor as
// an in-memory MsgPack-decodable value.
func (c Cursor) UnpackValue(value interface{}) error {
	return msgpackUnmarshal(c.data, value)
}

// String formats the MsgPack-encoded cursor as a hex-encoded string.
func (c Cursor) String() string {
	return hex.EncodeToString(c.data)
}

// FromString hex-decodes a string as a raw byte slice.
// The actual cursor value still has to be MsgPack-decoded to a value,
// using the (Cursor).UnpackValue method.
func (c *Cursor) FromString(s string) error {
	if len(s) == 0 {
		c.data = nil
		return nil
	}
	b, err := hex.DecodeString(s)
	if err != nil {
		return err
	}
	c.data = b
	return nil
}

// MarshalGQL implements the graphql.Marshaler interface
func (c Cursor) MarshalGQL(w io.Writer) {
	graphql.MarshalString(c.String()).MarshalGQL(w)
}

// UnmarshalGQL implements the graphql.Unmarshaler interface
func (c *Cursor) UnmarshalGQL(v interface{}) error {
	switch v := v.(type) {
	case string:
		b, err := hex.DecodeString(v)
		if err != nil {
			return fmt.Errorf("failed to UnmarshalGQL cursor: failed to hex-decode string '%s': %v", v, err)
		}
		c.data = b
		return nil
	case []byte:
		length := len(v)
		c.data = make([]byte, length/2)
		_, err := hex.Decode(c.data[:], v[:])
		if err != nil {
			return fmt.Errorf("failed to UnmarshalGQL cursor: failed to hex-decode bytes: %v", err)
		}
		return nil
	default:
		return fmt.Errorf("failed to UnmarshalGQL cursor: %T is not a string or byte slice", v)
	}
}
