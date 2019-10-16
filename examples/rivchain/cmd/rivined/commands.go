package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/threefoldtech/rivine/pkg/cli"
	"github.com/threefoldtech/rivine/pkg/daemon"
	"github.com/threefoldtech/rivine/profile"

	"github.com/bgentry/speakeasy"
	"github.com/spf13/cobra"
)

type commands struct {
	cfg           ExtendedDaemonConfig
	moduleSetFlag daemon.ModuleSetFlag
}

func (cmds *commands) rootCommand(*cobra.Command, []string) {
	var err error

	// Silently append a subdirectory for storage with the name of the network so we don't create conflicts
	cmds.cfg.RootPersistentDir = filepath.Join(cmds.cfg.RootPersistentDir, cmds.cfg.BlockchainInfo.NetworkName)

	// Check if we require an api password
	if cmds.cfg.AuthenticateAPI {
		// if not set, ask one now
		if cmds.cfg.APIPassword == "" {
			// Prompt user for API password.
			cmds.cfg.APIPassword, err = speakeasy.Ask("Enter API password: ")
			if err != nil {
				cli.DieWithError("failed to ask for API password", err)
			}
		}
		if cmds.cfg.APIPassword == "" {
			cli.DieWithError("failed to configure daemon", errors.New("password cannot be blank"))
		}
	} else {
		// If authenticateAPI is not set, explicitly set the password to the empty string.
		// This way the api server maintains consistency with the authenticateAPI var, even if apiPassword is set (possibly by mistake)
		cmds.cfg.APIPassword = ""
	}

	// Process the config variables, cleaning up slightly invalid values
	cmds.cfg.Config = daemon.ProcessConfig(cmds.cfg.Config)

	// Create the profiling directory if profiling is enabled.
	if cmds.cfg.Profile {
		go profile.StartContinuousProfile(cmds.cfg.ProfileDir, cmds.cfg.BlockchainInfo, cmds.cfg.VerboseLogging)
	}

	// run daemon
	err = runDaemon(cmds.cfg, cmds.moduleSetFlag.ModuleIdentifiers())
	if err != nil {
		cli.DieWithError("daemon failed", err)
	}
}

func (cmds *commands) versionCommand(*cobra.Command, []string) {
	var postfix string
	switch cmds.cfg.BlockchainInfo.NetworkName {

	case "devnet":

		postfix = "-devnet"

	case "standard":

		// no special prefix needed for standard-type networks

	case "testnet":

		postfix = "-testnet"

	default:
		postfix = "-???"
	}
	fmt.Printf("%s Daemon v%s%s\n",
		strings.Title(cmds.cfg.BlockchainInfo.Name),
		cmds.cfg.BlockchainInfo.ChainVersion.String(), postfix)
	fmt.Println("Rivine Protocol v" + cmds.cfg.BlockchainInfo.ProtocolVersion.String())

	fmt.Println()
	fmt.Printf("Go Version   v%s\r\n", runtime.Version()[2:])
	fmt.Printf("GOOS         %s\r\n", runtime.GOOS)
	fmt.Printf("GOARCH       %s\r\n", runtime.GOARCH)
}

func (cmds *commands) modulesCommand(*cobra.Command, []string) {
	err := cmds.moduleSetFlag.WriteDescription(os.Stdout)
	if err != nil {
		cli.DieWithError("failed to write usage of the modules flag", err)
	}
}
