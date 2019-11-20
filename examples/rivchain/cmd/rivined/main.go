package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/threefoldtech/rivine/examples/rivchain/pkg/config"
	"github.com/threefoldtech/rivine/pkg/cli"
	"github.com/threefoldtech/rivine/pkg/daemon"
)

func main() {
	var cmds commands
	// load default config to start with
	cmds.cfg.Config = DefaultConfig()
	cmds.cfg.BlockchainInfo = config.GetBlockchainInfo()

	// load default config flag
	cmds.moduleSetFlag = daemon.DefaultModuleSetFlag()

	// create the root command and add the flags to the root command
	rootCommand := &cobra.Command{
		Use: os.Args[0],
		Short: strings.Title(cmds.cfg.BlockchainInfo.Name) + " Daemon v" +
			cmds.cfg.BlockchainInfo.ChainVersion.String(),
		Long: strings.Title(cmds.cfg.BlockchainInfo.Name) + " Daemon v" +
			cmds.cfg.BlockchainInfo.ChainVersion.String(),
		Run: cmds.rootCommand,
	}
	cmds.cfg.RegisterAsFlags(rootCommand.Flags())
	// also add our modules as a flag
	cmds.moduleSetFlag.RegisterFlag(rootCommand.Flags(), fmt.Sprintf("%s modules", os.Args[0]))

	// register a custom explorerGraphQL BCDB Flag
	rootCommand.Flags().StringVar(
		&cmds.cfg.ExplorerBCDBAddress, "explorer-bcdb", "",
		"when GraphQL Explorer enabled, use the BCDB as backend reachable at this address instead of an embedded StormDB")

	// create the other commands
	rootCommand.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Long: "Print version information about the " +
			strings.Title(cmds.cfg.BlockchainInfo.Name) + " Daemon",
		Run: cmds.versionCommand,
	})

	rootCommand.AddCommand(&cobra.Command{
		Use:   "modules",
		Short: "List available modules for use with -M, --modules flag",
		Long:  "List available modules for use with -M, --modules flag and their uses",
		Run:   cmds.modulesCommand,
	})

	// Parse cmdline flags, overwriting both the default values and the config
	// file values.
	if err := rootCommand.Execute(); err != nil {
		// Since no commands return errors (all commands set Command.Run instead of
		// Command.RunE), Command.Execute() should only return an error on an
		// invalid command or flag. Therefore Command.Usage() was called (assuming
		// Command.SilenceUsage is false) and we should exit with exitCodeUsage.
		os.Exit(cli.ExitCodeUsage)
	}
}
