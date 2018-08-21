package daemon

import "github.com/rivine/rivine/build"

// Version defines the version of a daemon.
type Version struct {
	ChainVersion    build.ProtocolVersion `json:"version"`
	ProtocolVersion build.ProtocolVersion `json:"protocolVersion"`
}
