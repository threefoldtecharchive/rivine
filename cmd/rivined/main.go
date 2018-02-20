package main

import "github.com/rivine/rivine/rivined"

// main establishes a set of commands and flags using the cobra package.
func main() {
	// Daemon name defaults to rivine if it isn't set, but do so anyway just to make sure
	rivined.DaemonName = "rivine"
	cfg := rivined.DefaultConfig()
	rivined.SetupDefaultDaemon(cfg)
}
