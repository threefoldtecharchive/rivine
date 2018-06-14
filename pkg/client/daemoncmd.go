package client

import (
	"fmt"
)

// Stopcmd is the handler for the command `siac stop`.
// Stops the daemon.
func stopcmd() {
	err := _DefaultClient.httpClient.Post("/daemon/stop", "")
	if err != nil {
		Die("Could not stop daemon:", err)
	}
	cfg := _ConfigStorage.Config()
	fmt.Printf("%s daemon stopped.\n", cfg.ChainName)
}
