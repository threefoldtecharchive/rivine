package client

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	stopCmd = &cobra.Command{
		Use:   "stop",
		Short: fmt.Sprintf("Stop the %s daemon", _DefaultClient.name),
		Long:  fmt.Sprintf("Stop the %s daemon.", _DefaultClient.name),
		Run:   Wrap(stopcmd),
	}

	updateCmd = &cobra.Command{
		Use:   "update",
		Short: fmt.Sprintf("Update %s", _DefaultClient.name),
		Long:  fmt.Sprintf("Check for (and/or download) available updates for %s.", _DefaultClient.name),
		Run:   Wrap(updatecmd),
	}

	updateCheckCmd = &cobra.Command{
		Use:   "check",
		Short: "Check for available updates",
		Long:  "Check for available updates.",
		Run:   Wrap(updatecheckcmd),
	}
)

type updateInfo struct {
	Available bool   `json:"available"`
	Version   string `json:"version"`
}

// Stopcmd is the handler for the command `siac stop`.
// Stops the daemon.
func stopcmd() {
	err := _DefaultClient.httpClient.Post("/daemon/stop", "")
	if err != nil {
		Die("Could not stop daemon:", err)
	}
	fmt.Printf("%s daemon stopped.\n", _DefaultClient.name)
}

func updatecmd() {
	var update updateInfo
	err := _DefaultClient.httpClient.GetAPI("/daemon/update", &update)
	if err != nil {
		fmt.Println("Could not check for update:", err)
		return
	}
	if !update.Available {
		fmt.Println("Already up to date.")
		return
	}

	err = _DefaultClient.httpClient.Post("/daemon/update", "")
	if err != nil {
		fmt.Println("Could not apply update:", err)
		return
	}
	fmt.Printf("Updated to version %s! Restart %s now.\n", update.Version, _DefaultClient.name)
}

func updatecheckcmd() {
	var update updateInfo
	err := _DefaultClient.httpClient.GetAPI("/daemon/update", &update)
	if err != nil {
		fmt.Println("Could not check for update:", err)
		return
	}
	if update.Available {
		fmt.Printf("A new release (v%s) is available! Run '%s update' to install it.\n", update.Version, _DefaultClient.name)
	} else {
		fmt.Println("Up to date.")
	}
}
