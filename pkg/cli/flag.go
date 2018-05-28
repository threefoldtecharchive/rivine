package cli

import (
	"strconv"
	"time"

	"github.com/spf13/pflag"
)

const (
	// DateOnlyLayout allows you to define a timestamp using a date only,
	// where GMT is assumed as the timezone: DD/MM/YYYY TZN
	DateOnlyLayout = "02/01/2006 MST"
)

// LockTimeFlag defines LockTime as a flag,
// as to give the user several ways to define the lock time,
// such that for example the user isn't required to define it in unix epoch time.
type LockTimeFlag struct {
	lockTime uint64
	rawFlag  string
}

// String implements pflag.Value.String,
// printing this LockTime either as a timestamp in DateOnlyLayout or RFC822 layout,
// a duration or as an uint64.
func (f *LockTimeFlag) String() string {
	return f.rawFlag
}

// Set implements pflag.Value.Set,
// which parses the given string either as a timestamp in DateOnlyLayout or RFC822 layout,
// a duration or as an uint64.
func (f *LockTimeFlag) Set(s string) error {
	f.rawFlag = s
	if t, err := time.Parse(DateOnlyLayout, s); err == nil {
		// epoch unix (block) time
		f.lockTime = uint64(t.Unix())
		return nil
	}
	if t, err := time.Parse(time.RFC822, s); err == nil {
		// epoch unix (block) time
		f.lockTime = uint64(t.Unix())
		return nil
	}
	if d, err := time.ParseDuration(s); err == nil {
		// epoch unix (block) time
		f.lockTime = uint64(computeTimeNow().Add(d).Unix())
		return nil
	}
	// epoch unix (block) time or block height
	x, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return err
	}
	f.lockTime = x
	return nil
}

// Type implements pflag.Value.Type
func (f *LockTimeFlag) Type() string {
	return "LockTime"
}

// LockTime returns the internal lock time of this LockTime flag
func (f *LockTimeFlag) LockTime() uint64 {
	return f.lockTime
}

type (
	// StringLoaderFlag defines a utility type,
	// allowing any StringLoader to be turned into a pflag-interface compatible type.
	StringLoaderFlag struct {
		StringLoader
	}
	// StringLoader defines the interface of a type (t),
	// which can be loaded from a string,
	// as well as being turned back into a string,
	// such that `t.LoadString(str) == nil && t.String() == str`.
	StringLoader interface {
		LoadString(string) error
		String() string
	}
)

// Set implements pflag.Value.Set,
// which parses the string using the StringLoader's LoadString method.
func (s StringLoaderFlag) Set(str string) error {
	return s.LoadString(str)
}

// Type implements pflag.Value.Type
func (s StringLoaderFlag) Type() string {
	return "StringLoader"
}

var computeTimeNow = func() time.Time {
	return time.Now()
}

var (
	_ pflag.Value = (*LockTimeFlag)(nil)
	_ pflag.Value = StringLoaderFlag{}
)
