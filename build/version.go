package build

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
)

// Parse attempts to create a version based on a given string
func Parse(raw string) (ver ProtocolVersion, err error) {
	parts := versionReg.FindStringSubmatch(raw)
	if len(parts) != 5 {
		err = InvalidVersionError(raw)
		return
	}

	// because of the regexp we can be sure that this will always succeed
	// an empty string automatically results in 0, even though it also returns an err
	major, _ := strconv.ParseUint(parts[1], 10, 8)
	minor, _ := strconv.ParseUint(parts[2], 10, 8)
	patch, _ := strconv.ParseUint(parts[3], 10, 8)

	ver = NewPrereleaseVersion(uint8(major), uint8(minor), uint8(patch), parts[4])
	return
}

// MustParse creates a version based on a given string,
// panics in case the given string is invalid
func MustParse(raw string) ProtocolVersion {
	version, err := Parse(raw)
	if err != nil {
		panic(err)
	}
	return version
}

// NewVersion creates a new protocol version
func NewVersion(major, minor, patch uint8) ProtocolVersion {
	return NewPrereleaseVersion(major, minor, patch, "")
}

// NewPrereleaseVersion creates a new protocol prerelease version
func NewPrereleaseVersion(major, minor, patch uint8, prerelease string) ProtocolVersion {
	var v ProtocolVersion
	v.Version = (uint32(major) << 24) | (uint32(minor) << 16) | (uint32(patch) << 8)
	copy(v.Prerelease[:], prerelease[:])
	return v
}

// ProtocolVersion defines the protocol version that a node uses
type ProtocolVersion struct {
	Version    uint32  // Semantic Versioning
	Prerelease [8]byte // Any pre-release tag (eg. 'alpha' or the git commit hash)
}

// InvalidVersionError indicates a protocol version is invalid.
type InvalidVersionError string

// Error implements the error interface for InvalidVersionError.
func (e InvalidVersionError) Error() string {
	if len(e) == 0 {
		return "invalid version: <nil>"
	}

	return "invalid version: " + string(e)
}

// Compare returns an integer comparing this version with another version.
func (pv *ProtocolVersion) Compare(other ProtocolVersion) int {
	if pv.Version < other.Version {
		return -1
	} else if pv.Version > other.Version {
		return 1
	}

	isAPrerelease := pv.Prerelease != nilPreRelease
	isBPrerelease := other.Prerelease != nilPreRelease

	if !isAPrerelease && isBPrerelease {
		return 1
	} else if isAPrerelease && !isBPrerelease {
		return -1
	}

	// NOTE: if the 2 versions both have a prerelease defined we count them as equal,
	//       regardless if those prerelease values are different or not

	return 0
}

// String returns the string version of this ProtocolVersion
func (pv *ProtocolVersion) String() string {
	str := fmt.Sprintf("%d.%d.%d",
		(pv.Version>>24)&0xFF, // major
		(pv.Version>>16)&0xFF, // minor
		(pv.Version>>8)&0xFF,  // patch
	)

	// optional prerelease
	if pv.Prerelease != nilPreRelease {
		str += "-" + string(pv.Prerelease[:])
	}

	return str
}

// MarshalJSON implements json.Marshaler.MarshalJSON
func (pv ProtocolVersion) MarshalJSON() ([]byte, error) {
	return json.Marshal(pv.String())
}

// UnmarshalJSON implements json.Unmarshaler.UnmarshalJSON
func (pv *ProtocolVersion) UnmarshalJSON(b []byte) error {
	var raw string
	if err := json.Unmarshal(b, &raw); err != nil {
		return InvalidVersionError(string(b))
	}

	result, err := Parse(raw)
	if err != nil {
		return err
	}

	pv.Version = result.Version
	pv.Prerelease = result.Prerelease
	return nil
}

var (
	nilPreRelease = [8]byte{0, 0, 0, 0, 0, 0, 0, 0}
)

var (
	// rawVersion used to generate rivine's protocol version
	rawVersion = "v1.0.5"
	// Version is the current version of rivined.
	Version ProtocolVersion
)

const (
	// EncodedVersionLength is the static length of a sia-encoded ProtocolVersion.
	EncodedVersionLength = 16 // sizeof(uint32==64) + sizeof([8]uint8)
)

const versionRe = `^v?(0{0,2}[0-9]|[0-1]?[0-9]{1,2}|2[0-4][0-9]|25[0-5])(?:\.(0{0,2}[0-9]|[0-1]?[0-9]{1,2}|2[0-4][0-9]|25[0-5]))?(?:\.(0{0,2}[0-9]|[0-1]?[0-9]{1,2}|2[0-4][0-9]|25[0-5]))?(?:-(.+?))?$`

// contains the regexp for all valid versions
var versionReg = regexp.MustCompile(versionRe)

func init() {
	Version = MustParse(rawVersion)
}
