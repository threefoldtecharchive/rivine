package daemon

import "github.com/threefoldtech/rivine/build"

// Version defines the version of a daemon.
type Version struct {
	ChainVersion    build.ProtocolVersion `json:"version"`
	ProtocolVersion build.ProtocolVersion `json:"protocolVersion"`
}
