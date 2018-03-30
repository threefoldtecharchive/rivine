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
	fmt.Printf("%s daemon stopped.\n", _DefaultClient.name)
}
