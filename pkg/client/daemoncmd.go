package client

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	stopCmd = &cobra.Command{
		Use:   "stop",
		Short: fmt.Sprintf("Stop the %s daemon", ClientName),
		Long:  fmt.Sprintf("Stop the %s daemon.", ClientName),
		Run:   wrap(Stopcmd),
	}

	updateCmd = &cobra.Command{
		Use:   "update",
		Short: fmt.Sprintf("Update %s", ClientName),
		Long:  fmt.Sprintf("Check for (and/or download) available updates for %s.", ClientName),
		Run:   wrap(Updatecmd),
	}

	updateCheckCmd = &cobra.Command{
		Use:   "check",
		Short: "Check for available updates",
		Long:  "Check for available updates.",
		Run:   wrap(Updatecheckcmd),
	}
)

type updateInfo struct {
	Available bool   `json:"available"`
	Version   string `json:"version"`
}

// Stopcmd is the handler for the command `siac stop`.
// Stops the daemon.
func Stopcmd() {
	err := Post("/daemon/stop", "")
	if err != nil {
		Die("Could not stop daemon:", err)
	}
	fmt.Printf("%s daemon stopped.\n", ClientName)
}

func Updatecmd() {
	var update updateInfo
	err := GetAPI("/daemon/update", &update)
	if err != nil {
		fmt.Println("Could not check for update:", err)
		return
	}
	if !update.Available {
		fmt.Println("Already up to date.")
		return
	}

	err = Post("/daemon/update", "")
	if err != nil {
		fmt.Println("Could not apply update:", err)
		return
	}
	fmt.Printf("Updated to version %s! Restart %s now.\n", update.Version, ClientName)
}

func Updatecheckcmd() {
	var update updateInfo
	err := GetAPI("/daemon/update", &update)
	if err != nil {
		fmt.Println("Could not check for update:", err)
		return
	}
	if update.Available {
		fmt.Printf("A new release (v%s) is available! Run '%s update' to install it.\n", update.Version, ClientName)
	} else {
		fmt.Println("Up to date.")
	}
}
