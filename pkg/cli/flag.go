package cli

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/pflag"
	"github.com/threefoldtech/rivine/build"
	"github.com/threefoldtech/rivine/modules"
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

type (
	// EncodingTypeFlag is a utility flag which can be used to
	// expose an encoding type as an optionally masked flag.
	EncodingTypeFlag struct {
		et   *EncodingType
		mask EncodingType
	}
	// EncodingType defines an enum, to represent all encoding options.
	EncodingType uint8

	// ArbitraryDataFlag is used to pass arbitrary data to transactions
	ArbitraryDataFlag struct {
		arbitraryData *[]byte
	}
)

const (
	// EncodingTypeHuman encodes the output in a human-optimized format,
	// a format which can chance at any point and should be used only
	// for human readers.
	EncodingTypeHuman EncodingType = 1 << iota
	// EncodingTypeJSON encodes the output as a minified JSON string,
	// and has a format which is promised to be backwards-compatible,
	// to be used for automation purposes.
	EncodingTypeJSON
	// EncodingTypeHex encodes the output using the internal binary encoder,
	// and encoding that binary output using the std hex encoder,
	// resulting in a hex-encoded string.
	EncodingTypeHex
)

// defaultEncodingTypeMask returns a mask which allows all possible encoding types.
func defaultEncodingTypeMask() EncodingType {
	return EncodingTypeHuman | EncodingTypeJSON | EncodingTypeHex
}

// NewEncodingTypeFlag returns a new EncodingTypeFlag,
// referencing an encoding type value, defaulting it to a default value,
// and optionally allowing you to mask
func NewEncodingTypeFlag(def EncodingType, ref *EncodingType, mask EncodingType) EncodingTypeFlag {
	if ref == nil {
		build.Critical("no encoding type reference given")
	}
	if def == 0 {
		// default to human encoding
		def = EncodingTypeHuman
	}
	if mask == 0 { // default to all options
		mask = defaultEncodingTypeMask()
	}
	*ref = def
	// sanity checks
	if mask&def == 0 {
		build.Critical(fmt.Sprintf("given default encoding type %d is not covered by given encoding type mask %b", def, mask))
	} else if def != EncodingTypeHuman && def != EncodingTypeJSON && def != EncodingTypeHex {
		build.Critical(fmt.Sprintf("given default encoding type %d is not a valid encoding type", def))
	}
	return EncodingTypeFlag{
		et:   ref,
		mask: mask,
	}
}

// String implements pflag.Value.String,
// returning the selected enum option as a lower-case string.
func (e EncodingTypeFlag) String() string {
	switch *e.et {
	case EncodingTypeJSON:
		return "json"
	case EncodingTypeHex:
		return "hex"
	default:
		return "human"
	}
}

// Set implements pflag.Value.Set,
// only the options as defind by the mask are allowed,
// and the given string is interpreted in a case insensitive manner.
func (e EncodingTypeFlag) Set(s string) error {
	switch strings.ToLower(s) {
	case "json":
		if e.mask&EncodingTypeJSON == 0 {
			return errors.New("this command does not suppport JSON encoding")
		}
		*e.et = EncodingTypeJSON
	case "hex":
		if e.mask&EncodingTypeHex == 0 {
			return errors.New("this command does not suppport Binary-Hex encoding")
		}
		*e.et = EncodingTypeHex
	case "human":
		if e.mask&EncodingTypeHuman == 0 {
			return errors.New("this command does not suppport Human-Format encoding")
		}
		*e.et = EncodingTypeHuman
	default:
		return fmt.Errorf("%q is not a valid EncodingType", s)
	}
	return nil
}

// Type implements pflag.Value.Type
func (e EncodingTypeFlag) Type() string {
	return "EncodingType"
}

// EncodingTypeFlagDescription returns a description for an encoding type flag,
// optionally given an encoding type mask (0 means all encoding types are allowed).
func EncodingTypeFlagDescription(mask EncodingType) string {
	if mask == 0 { // default to all options
		mask = defaultEncodingTypeMask()
	}

	var options []string
	if mask&EncodingTypeJSON != 0 {
		options = append(options, "json")
	}
	if mask&EncodingTypeHex != 0 {
		options = append(options, "hex")
	}
	if mask&EncodingTypeHuman != 0 {
		options = append(options, "human")
	}
	return "enum flag to define how to encode the output, options: " + strings.Join(options, "|")
}

// ArbitraryDataFlagVar defines a ArbitraryData flag with specified name and usage string.
// The argument s points to an array of bytes variable in which to store the validated values of the flags.
// The value of each argument will not try to be separated by comma, each value has to be defined as a separate flag (using the same name).
func ArbitraryDataFlagVar(f *pflag.FlagSet, s *[]byte, name string, usage string) {
	f.Var(&ArbitraryDataFlag{arbitraryData: s}, name, usage)
}

// ArbitraryDataFlagVarP defines a ArbitraryData flag with specified name, shorthand and usage string.
// The argument s points to an array of bytes variable in which to store the validated values of the flags.
// The value of each argument will not try to be separated by comma, each value has to be defined as a separate flag (using the same name or shorthand).
func ArbitraryDataFlagVarP(f *pflag.FlagSet, s *[]byte, name, shorthand string, usage string) {
	f.VarP(&ArbitraryDataFlag{arbitraryData: s}, name, shorthand, usage)
}

// Set implements pflag.Value.Set
func (flag *ArbitraryDataFlag) Set(str string) error {
	if out, err := strconv.Unquote(`"` + str + `"`); err == nil {
		*flag.arbitraryData = []byte(out)
	} else {
		*flag.arbitraryData = []byte(str)
	}
	return nil
}

// Type implements pflag.Value.Type
func (flag *ArbitraryDataFlag) Type() string {
	return "ArbitraryDataFlag"
}

// String implements pflag.Value.String
func (flag *ArbitraryDataFlag) String() string {
	byteArray := *flag.arbitraryData
	if len(byteArray) == 0 {
		return ""
	}

	s := strconv.Quote(string(byteArray[:]))
	return strings.Trim(s, "\"")
}

// NetAddressArrayFlagVar defines a []modules.NetAddress flag with specified name and usage string.
// The argument s points to a []modules.NetAddress variable in which to store the validated values of the flags.
// The value of each argument will not try to be separated by comma, each value has to be defined as a separate flag (using the same name).
func NetAddressArrayFlagVar(f *pflag.FlagSet, s *[]modules.NetAddress, name string, usage string) {
	f.Var(&netAddressArray{array: s}, name, usage)
}

// NetAddressArrayFlagVarP defines a []modules.NetAddress flag with specified name, shorthand and usage string.
// The argument s points to a []modules.NetAddress variable in which to store the validated values of the flags.
// The value of each argument will not try to be separated by comma, each value has to be defined as a separate flag (using the same name or shorthand).
func NetAddressArrayFlagVarP(f *pflag.FlagSet, s *[]modules.NetAddress, name, shorthand string, usage string) {
	f.VarP(&netAddressArray{array: s}, name, shorthand, usage)
}

type netAddressArray struct {
	array   *[]modules.NetAddress
	changed bool
}

// Set implements pflag.Value.Set
func (flag *netAddressArray) Set(val string) error {
	if !flag.changed {
		*flag.array = make([]modules.NetAddress, 0)
		flag.changed = true
	}
	na := modules.NetAddress(val)
	err := na.IsStdValid()
	if err != nil {
		return fmt.Errorf("invalid network address %v: %v", val, err)
	}
	*flag.array = append(*flag.array, na)
	return nil
}

// Type implements pflag.Value.Type
func (flag *netAddressArray) Type() string {
	return "NetAddressArray"
}

// String implements pflag.Value.String
func (flag *netAddressArray) String() string {
	if flag.array == nil || len(*flag.array) == 0 {
		return ""
	}
	var str string
	for _, na := range *flag.array {
		str += string(na) + ","
	}
	return str[:len(str)-1]
}

var computeTimeNow = func() time.Time {
	return time.Now()
}

var (
	_ pflag.Value = (*LockTimeFlag)(nil)
	_ pflag.Value = StringLoaderFlag{}
)
