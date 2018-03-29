package main

import (
	"github.com/rivine/rivine/pkg/daemon"
)

// main establishes a set of commands and flags using the cobra package.
func main() {
	// Use the default daemon configuration
	cfg := daemon.DefaultConfig()
	// setup & start the daemon
	daemon.SetupDefaultDaemon(cfg)
}
