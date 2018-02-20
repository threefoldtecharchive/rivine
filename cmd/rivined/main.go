package main

import "github.com/rivine/rivine/rivined"

// main establishes a set of commands and flags using the cobra package.
func main() {

	cfg := rivined.DefaultConfig()
	rivined.SetupDefaultDaemon(cfg)
}
