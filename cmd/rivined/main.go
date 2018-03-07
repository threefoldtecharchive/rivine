package main

import "github.com/rivine/rivine/pkg/daemon"

// main establishes a set of commands and flags using the cobra package.
func main() {
	// Daemon name defaults to rivine if it isn't set, but do so anyway just to make sure
	daemon.DaemonName = "rivine"
	cfg := daemon.DefaultConfig()
	daemon.SetupDefaultDaemon(cfg)
}
