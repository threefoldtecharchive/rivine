package main

import (
	"github.com/threefoldtech/rivine/modules"
	"github.com/threefoldtech/rivine/pkg/daemon"
)

// ExtendedDaemonConfig contains all configurable variables for rivine.
type ExtendedDaemonConfig struct {
	daemon.Config

	BootstrapPeers []modules.NetAddress
}
