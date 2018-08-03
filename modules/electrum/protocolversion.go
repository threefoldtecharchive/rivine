package electrum

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type (

	// ProtocolVersion is a specific version of the electrum protocol.
	// The version is denoted by a major and minor number, and optionally, a rivision
	ProtocolVersion struct {
		major    uint8
		minor    uint8
		revision uint8
	}

	// ProtocolArgument is the input for the `server.version` call's `protocol_version`
	// argument. As per the spec, it is either a single version, or an array
	// [protocol_min, protocol_max].
	ProtocolArgument struct {
		protocolMin ProtocolVersion
		protocolMax ProtocolVersion
	}
)

// String implements the stringer interface
func (epv *ProtocolVersion) String() string {
	rawVersion := fmt.Sprintf("%v.%v", epv.major, epv.minor)
	if epv.revision != 0 {
		rawVersion = fmt.Sprintf("%v.%v", rawVersion, epv.revision)
	}
	return rawVersion
}

// MarshalJSON produces a valid json value for the protocol version
func (epv ProtocolVersion) MarshalJSON() ([]byte, error) {
	return json.Marshal(epv.String())
}

// UnmarshalJSON unmarshals a json value to a protocol version
func (epv *ProtocolVersion) UnmarshalJSON(data []byte) error {
	var rawVersion string
	if err := json.Unmarshal(data, &rawVersion); err != nil {
		return err
	}
	if err := epv.ParseRawVersion(rawVersion); err != nil {
		return err
	}
	return nil
}

// ParseRawVersion tries to parse a version string. If parsing succeeds (no error
// is returned), the version information is updated.
func (epv *ProtocolVersion) ParseRawVersion(rawVersion string) error {
	segments := strings.Split(rawVersion, ".")
	if len(segments) < 2 || len(segments) > 3 {
		return errors.New("Invalid amount of segments")
	}
	major, err := strconv.ParseUint(segments[0], 10, 8)
	if err != nil {
		return err
	}
	minor, err := strconv.ParseUint(segments[1], 10, 8)
	if err != nil {
		return err
	}
	// if there is no revision number we're done here
	if len(segments) == 2 {
		epv.major = uint8(major)
		epv.minor = uint8(minor)
		return nil
	}
	revision, err := strconv.ParseUint(segments[2], 10, 8)
	if err != nil {
		return err
	}
	epv.major = uint8(major)
	epv.minor = uint8(minor)
	epv.revision = uint8(revision)
	return nil
}

// MarshalJSON produces a valid json value for the protocol argument
func (pa ProtocolArgument) MarshalJSON() ([]byte, error) {
	if pa.protocolMin == pa.protocolMax {
		return json.Marshal(pa.protocolMin)
	}
	return json.Marshal([]ProtocolVersion{pa.protocolMin, pa.protocolMax})
}

// UnmarshalJSON unmarshals a json value to a protocol argument
func (pa *ProtocolArgument) UnmarshalJSON(data []byte) error {
	var pv ProtocolVersion
	if err := json.Unmarshal(data, &pv); err != nil {
		var versions []ProtocolVersion
		if err = json.Unmarshal(data, &versions); err != nil {
			return err
		}
		if len(versions) != 2 {
			return errors.New("Invalid amount of protocol versions")
		}
		*pa = ProtocolArgument{
			protocolMin: versions[0],
			protocolMax: versions[1],
		}
		return nil
	}
	*pa = ProtocolArgument{
		protocolMin: pv,
		protocolMax: pv,
	}
	return nil
}
