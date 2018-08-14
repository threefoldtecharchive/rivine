package main

import (
	"fmt"
	"os"

	"github.com/rivine/rivine/modules"
	"github.com/rivine/rivine/pkg/client"
	"github.com/rivine/rivine/types"
)

func main() {
	// create command line client
	bchainInfo := types.DefaultBlockchainInfo()
	cli, err := client.NewCommandLineClient("", bchainInfo.Name)
	if err != nil {
		panic(err)
	}
	// define preRunE, as to ensure we go to a default config should it be required
	cli.PreRunE = func(cfg *client.Config) (*client.Config, error) {
		if cfg != nil {
			return cfg, nil
		}
		chainConstants := types.DefaultChainConstants()
		daemonConstants := modules.NewDaemonConstants(bchainInfo, chainConstants)
		vcfg := client.ConfigFromDaemonConstants(daemonConstants)
		return &vcfg, nil
	}
	// start cli
	if err := cli.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "client exited with an error: ", err)
		// Since no commands return errors (all commands set Command.Run instead of
		// Command.RunE), Command.Execute() should only return an error on an
		// invalid command or flag. Therefore Command.Usage() was called (assuming
		// Command.SilenceUsage is false) and we should exit with exitCodeUsage.
		os.Exit(client.ExitCodeUsage)
	}
}
