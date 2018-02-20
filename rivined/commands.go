package rivined

import (
	"fmt"
	"os"
	"strings"

	"github.com/rivine/rivine/build"
	"github.com/rivine/rivine/profile"
	"github.com/spf13/cobra"
)

// exit codes
// inspired by sysexits.h
const (
	exitCodeGeneral = 1  // Not in sysexits.h, but is standard practice.
	exitCodeUsage   = 64 // EX_USAGE in sysexits.h
)

// DaemonName sets the name of the daemon, used in some help messages
var (
	DaemonName = "rivine"
)

// die prints its arguments to stderr, then exits the program with the default
// error code.
func die(args ...interface{}) {
	fmt.Fprintln(os.Stderr, args...)
	os.Exit(exitCodeGeneral)
}

// startDaemonCmd is a passthrough function for startDaemon.
func startDaemonCmd(cmd *cobra.Command, _ []string) {
	// Create the profiling directory if profiling is enabled.
	if globalConfig.Rivined.Profile || build.DEBUG {
		go profile.StartContinuousProfile(globalConfig.Rivined.ProfileDir)
	}

	// Start siad. startDaemon will only return when it is shutting down.
	err := startDaemon(globalConfig)
	if err != nil {
		die(err)
	}
}

// versionCmd is a cobra command that prints the version of siad.
func versionCmd(*cobra.Command, []string) {
	switch build.Release {
	case "dev":
		fmt.Println(fmt.Sprintf("%s Daemon v", strings.Title(DaemonName)) + build.Version.String() + "-dev")
	case "standard":
		fmt.Println(fmt.Sprintf("%s Daemon v", strings.Title(DaemonName)) + build.Version.String())
	case "testing":
		fmt.Println(fmt.Sprintf("%s Daemon v", strings.Title(DaemonName)) + build.Version.String() + "-testing")
	default:
		fmt.Println(fmt.Sprintf("%s Daemon v", strings.Title(DaemonName)) + build.Version.String() + "-???")
	}
}

// modulesCmd is a cobra command that prints help info about modules.
func modulesCmd(*cobra.Command, []string) {
	fmt.Printf(`Use the -M or --modules flag to only run specific modules. Modules are
independent components of Sia. This flag should only be used by developers or
people who want to reduce overhead from unused modules. Modules are specified by
their first letter. If the -M or --modules flag is not specified the default
modules are run. The default modules are:
	gateway, consensus set, transaction pool, wallet, block creator
This is equivalent to:
	%[1]sd -M cgtwb
Below is a list of all the modules available.

Gateway (g):
	The gateway maintains a peer to peer connection to the network and
	enables other modules to perform RPC calls on peers.
	The gateway is required by all other modules.
	Example:
		%[1]sd -M g
Consensus Set (c):
	The consensus set manages everything related to consensus and keeps the
	blockchain in sync with the rest of the network.
	The consensus set requires the gateway.
	Example:
		%[1]sd -M gc
Transaction Pool (t):
	The transaction pool manages unconfirmed transactions.
	The transaction pool requires the consensus set.

	Example:
		%[1]sd -M gct
Wallet (w):
	The wallet stores and manages coins and blockstakes.
	The wallet requires the consensus set and transaction pool.
	Example:
		%[1]sd -M gctw
BlockCreator (b):
	The block creator participates in the proof of block stake protocol
	for creating new blocks. BlockStakes are required to participate.
	The block creator requires the consensus set, transaction pool and wallet.
	Example:
		%[1]sd -M gctwb
Explorer (e):
	The explorer provides statistics about the blockchain and can be
	queried for information about specific transactions or other objects on
	the blockchain.
	The explorer requires the consensus set.
	Example:
		%[1]sd -M gce
`, DaemonName)
}

// SetupDefaultDaemon sets up and starts a default daemon. The chain options and constants
// need to be configured prior to this. This function does not return untill the daemon is stopped
func SetupDefaultDaemon(cfg RivinedCfg) {
	root := &cobra.Command{
		Use:   os.Args[0],
		Short: strings.Title(DaemonName) + " Daemon v" + build.Version.String(),
		Long:  strings.Title(DaemonName) + " Daemon v" + build.Version.String(),
		Run:   startDaemonCmd,
	}

	root.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Long:  "Print version information about the " + strings.Title(DaemonName) + " Daemon",
		Run:   versionCmd,
	})

	root.AddCommand(&cobra.Command{
		Use:   "modules",
		Short: "List available modules for use with -M, --modules flag",
		Long:  "List available modules for use with -M, --modules flag and their uses",
		Run:   modulesCmd,
	})

	// Set default values, which have the lowest priority.
	root.Flags().StringVarP(&globalConfig.Rivined.RequiredUserAgent, "agent", "", cfg.RequiredUserAgent, "required substring for the user agent")
	root.Flags().StringVarP(&globalConfig.Rivined.ProfileDir, "profile-directory", "", cfg.ProfileDir, "location of the profiling directory")
	root.Flags().StringVarP(&globalConfig.Rivined.APIaddr, "api-addr", "", cfg.APIaddr, "which host:port the API server listens on")
	root.Flags().StringVarP(&globalConfig.Rivined.RivineDir, DaemonName+"-directory", "d", cfg.RivineDir, "location of the "+DaemonName+" directory")
	root.Flags().BoolVarP(&globalConfig.Rivined.NoBootstrap, "no-bootstrap", "", cfg.NoBootstrap, "disable bootstrapping on this run")
	root.Flags().BoolVarP(&globalConfig.Rivined.Profile, "profile", "", cfg.Profile, "enable profiling")
	root.Flags().StringVarP(&globalConfig.Rivined.RPCaddr, "rpc-addr", "", cfg.RPCaddr, "which port the gateway listens on")
	root.Flags().StringVarP(&globalConfig.Rivined.Modules, "modules", "M", cfg.Modules, fmt.Sprintf("enabled modules, see '%sd modules' for more info", DaemonName))
	root.Flags().BoolVarP(&globalConfig.Rivined.AuthenticateAPI, "authenticate-api", "", cfg.AuthenticateAPI, "enable API password protection")
	root.Flags().BoolVarP(&globalConfig.Rivined.AllowAPIBind, "disable-api-security", "", cfg.AllowAPIBind, fmt.Sprintf("allow %sd to listen on a non-localhost address (DANGEROUS)", DaemonName))

	// Parse cmdline flags, overwriting both the default values and the config
	// file values.
	if err := root.Execute(); err != nil {
		// Since no commands return errors (all commands set Command.Run instead of
		// Command.RunE), Command.Execute() should only return an error on an
		// invalid command or flag. Therefore Command.Usage() was called (assuming
		// Command.SilenceUsage is false) and we should exit with exitCodeUsage.
		os.Exit(exitCodeUsage)
	}
}
