package main

import (
	"fmt"
	"os"

	"github.com/threefoldtech/rivine/build"
	"github.com/threefoldtech/rivine/modules"
	"github.com/threefoldtech/rivine/pkg/cli"
	"github.com/threefoldtech/rivine/pkg/client"
	"github.com/threefoldtech/rivine/pkg/daemon"
	"github.com/threefoldtech/rivine/types"

	rivtypes "github.com/threefoldtech/rivine/cmd/rivinec/types"

	"github.com/threefoldtech/rivine/extensions/minting"
	mintingcli "github.com/threefoldtech/rivine/extensions/minting/client"

	"github.com/threefoldtech/rivine/extensions/authcointx"
	authcoincli "github.com/threefoldtech/rivine/extensions/authcointx/client"
)

func main() {
	// create command line client
	bchainInfo := types.DefaultBlockchainInfo()
	cliClient, err := client.NewCommandLineClient("", bchainInfo.Name, daemon.RivineUserAgent, nil)
	if err != nil {
		build.Critical(err)
	}

	// register minting extension commands
	mintingcli.CreateConsensusCmd(cliClient)
	mintingcli.CreateWalletCmds(cliClient,
		rivtypes.TransactionVersionMinterDefinition,
		rivtypes.TransactionVersionCoinCreation,
		&mintingcli.WalletCmdsOpts{
			CoinDestructionTxVersion: rivtypes.TransactionVersionCoinDestruction,
			RequireMinerFees:         false,
		})
	mintingcli.CreateExploreCmd(cliClient)

	// register authcoin extension commands
	authcoincli.CreateExploreAuthCoinInfoCmd(cliClient)
	authcoincli.CreateWalletCmds(
		cliClient,
		rivtypes.TransactionVersionAuthConditionUpdate,
		rivtypes.TransactionVersionAuthAddressUpdate,
	)

	// define preRunE, as to ensure we go to a default config should it be required
	cliClient.PreRunE = func(cfg *client.Config) (*client.Config, error) {
		if cfg == nil {
			chainConstants := types.StandardnetChainConstants()
			daemonConstants := modules.NewDaemonConstants(bchainInfo, chainConstants, nil)
			newCfg := client.ConfigFromDaemonConstants(daemonConstants)
			cfg = &newCfg
		}

		if !(cfg.NetworkName == "standard" || cfg.NetworkName == "devnet" || cfg.NetworkName == "testnet") {
			return nil, fmt.Errorf("Netork name %q not recognized", cfg.NetworkName)
		}

		// creating minting plugin client
		baseClient, err := client.NewBaseClientFromCommandLineClient(cliClient)
		if err != nil {
			return nil, fmt.Errorf("failed to create minting plugin client: %v", err)
		}
		mintingPluginClient := mintingcli.NewPluginConsensusClient(baseClient)

		// register minting transaction versions
		types.RegisterTransactionVersion(rivtypes.TransactionVersionMinterDefinition, &minting.MinterDefinitionTransactionController{
			MintingMinerFeeBaseTransactionController: minting.MintingMinerFeeBaseTransactionController{
				MintingBaseTransactionController: minting.MintingBaseTransactionController{
					UseLegacySiaEncoding: false,
				},
				RequireMinerFees: false,
			},
			MintConditionGetter: mintingPluginClient,
			TransactionVersion:  rivtypes.TransactionVersionMinterDefinition,
		})
		types.RegisterTransactionVersion(rivtypes.TransactionVersionCoinCreation, &minting.CoinCreationTransactionController{
			MintingMinerFeeBaseTransactionController: minting.MintingMinerFeeBaseTransactionController{
				MintingBaseTransactionController: minting.MintingBaseTransactionController{
					UseLegacySiaEncoding: false,
				},
				RequireMinerFees: false,
			},
			MintConditionGetter: mintingPluginClient,
			TransactionVersion:  rivtypes.TransactionVersionCoinCreation,
		})
		types.RegisterTransactionVersion(rivtypes.TransactionVersionCoinDestruction, &minting.CoinDestructionTransactionController{
			MintingBaseTransactionController: minting.MintingBaseTransactionController{
				UseLegacySiaEncoding: false,
			},
			TransactionVersion: rivtypes.TransactionVersionCoinDestruction,
		})

		// create authcoin plugin client
		authCoinPluginClient := authcoincli.NewPluginConsensusClient(baseClient)

		// register auth coin transaction versions
		types.RegisterTransactionVersion(rivtypes.TransactionVersionAuthAddressUpdate, &authcointx.AuthAddressUpdateTransactionController{
			AuthInfoGetter:     authCoinPluginClient,
			TransactionVersion: rivtypes.TransactionVersionAuthAddressUpdate,
		})
		types.RegisterTransactionVersion(rivtypes.TransactionVersionAuthConditionUpdate, &authcointx.AuthConditionUpdateTransactionController{
			AuthInfoGetter:     authCoinPluginClient,
			TransactionVersion: rivtypes.TransactionVersionAuthConditionUpdate,
		})

		return cfg, nil
	}
	// start cli
	if err := cliClient.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "client exited with an error: ", err)
		// Since no commands return errors (all commands set Command.Run instead of
		// Command.RunE), Command.Execute() should only return an error on an
		// invalid command or flag. Therefore Command.Usage() was called (assuming
		// Command.SilenceUsage is false) and we should exit with exitCodeUsage.
		os.Exit(cli.ExitCodeUsage)
	}
}
