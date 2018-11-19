package types

// difficulty.go defines the difficulty type and implements a few helper functions for
// manipulating the difficulty type.

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"math/big"

	"github.com/threefoldtech/rivine/pkg/encoding/rivbin"

	"github.com/threefoldtech/rivine/build"
	"github.com/threefoldtech/rivine/pkg/encoding/siabin"
)

type (
	// A Difficulty represents a number in number of blockstake times time in seconds
	// Normally the difficulty is the number of active blockstake times the
	// BlockFrequency. ex. If the number of active blockstake grows, the
	// difficulty will also increase to maintain the same BlockFrequency.
	Difficulty struct {
		i big.Int
	}
)

var (
	// ErrNegativeDifficulty is the error that is returned if performing an
	// operation results in a negative difficulty.
	ErrNegativeDifficulty = errors.New("negative difficulty not allowed")
)

// Big returns the value of c as a *big.Int. Importantly, it does not provide
// access to the c's internal big.Int object, only a copy.
func (x Difficulty) Big() *big.Int {
	return new(big.Int).Set(&x.i)
}

// NewDifficulty creates a Difficulty value from a big.Int. Undefined behavior
// occurs if a negative input is used.
func NewDifficulty(b *big.Int) (d Difficulty) {
	if b.Sign() < 0 {
		build.Critical(ErrNegativeDifficulty)
	} else {
		d.i = *b
	}
	return
}

// Div64 returns a new Difficulty value c = x / y.
func (x Difficulty) Div64(y uint64) (c Difficulty) {
	c.i.Div(&x.i, new(big.Int).SetUint64(y))
	return
}

// Cmp compares two Difficulty values. The return value follows the convention
// of math/big.
func (x Difficulty) Cmp(y Difficulty) int {
	return x.i.Cmp(&y.i)
}

// MarshalJSON implements the json.Marshaler interface.
func (c Difficulty) MarshalJSON() ([]byte, error) {
	// Must enclosed the value in quotes; otherwise JS will convert it to a
	// double and lose precision.
	return []byte(`"` + c.String() + `"`), nil
}

// UnmarshalJSON implements the json.Unmarshaler interface. An error is
// returned if a negative number is provided.
func (c *Difficulty) UnmarshalJSON(b []byte) error {
	// UnmarshalJSON does not expect quotes
	b = bytes.Trim(b, `"`)
	err := c.i.UnmarshalJSON(b)
	if err != nil {
		return err
	}
	if c.i.Sign() < 0 {
		c.i = *big.NewInt(0)
		return ErrNegativeDifficulty
	}
	return nil
}

// MarshalSia implements the siabin.SiaMarshaler interface. It writes the
// byte-slice representation of the Difficulty's internal big.Int to w. Note
// that as the bytes of the big.Int correspond to the absolute value of the
// integer, there is no way to marshal a negative Difficulty.
func (c Difficulty) MarshalSia(w io.Writer) error {
	return siabin.WritePrefix(w, c.i.Bytes())
}

// UnmarshalSia implements the siabin.SiaUnmarshaler interface.
func (c *Difficulty) UnmarshalSia(r io.Reader) error {
	b, err := siabin.ReadPrefix(r, 256)
	if err != nil {
		return err
	}
	c.i.SetBytes(b)
	return nil
}

// MarshalRivine implements the rivbin.RivineMarshaler interface. It writes the
// byte-slice representation of the Difficulty's internal big.Int to w. Note
// that as the bytes of the big.Int correspond to the absolute value of the
// integer, there is no way to marshal a negative Difficulty.
func (c Difficulty) MarshalRivine(w io.Writer) error {
	return rivbin.WriteDataSlice(w, c.i.Bytes())
}

// UnmarshalRivine implements the rivbin.RivineMarshaler interface.
func (c *Difficulty) UnmarshalRivine(r io.Reader) error {
	b, err := rivbin.ReadDataSlice(r, 256)
	if err != nil {
		return err
	}
	c.i.SetBytes(b)
	return nil
}

// String implements the fmt.Stringer interface.
func (c Difficulty) String() string {
	return c.i.String()
}

// Scan implements the fmt.Scanner interface, allowing Difficulty values to be
// scanned from text.
func (c *Difficulty) Scan(s fmt.ScanState, ch rune) error {
	err := c.i.Scan(s, ch)
	if err != nil {
		return err
	}
	if c.i.Sign() < 0 {
		return ErrNegativeDifficulty
	}
	return nil
}
