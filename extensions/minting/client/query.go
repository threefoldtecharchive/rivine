package client

import (
	"encoding/json"
	"os"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/threefoldtech/rivine/pkg/cli"
	client "github.com/threefoldtech/rivine/pkg/client"
	types "github.com/threefoldtech/rivine/types"
)

//CreateExploreCmd adds the explorer clisubcommands for the minting plugin
func CreateExploreCmd(cli *client.CommandLineClient) {
	bc, err := client.NewBaseClientFromCommandLineClient(cli)
	if err != nil {
		panic(err)
	}
	createCmd(cli.ExploreCmd, NewPluginExplorerClient(bc))
}

//CreateConsensusCmd adds the consensus cli subcommands for the minting plugin
func CreateConsensusCmd(cli *client.CommandLineClient) {
	bc, err := client.NewBaseClientFromCommandLineClient(cli)
	if err != nil {
		panic(err)
	}
	createCmd(cli.ConsensusCmd, NewPluginConsensusClient(bc))
}

func createCmd(rootCmd *cobra.Command, pluginClient *PluginClient) {
	subCmds := &subCmd{
		pluginClient: pluginClient,
	}

	// create root explore command and all subs
	var (
		getMintConditionCmd = &cobra.Command{
			Use:   "mintcondition [height]",
			Short: "Get the active mint condition",
			Long: `Get the active mint condition,
either the one active for the current block height,
or the one for the given block height.
`,
			Run: subCmds.getMintCondition,
		}
	)

	getMintConditionCmd.Flags().Var(
		cli.NewEncodingTypeFlag(cli.EncodingTypeHuman, &subCmds.getMintConditionCfg.EncodingType, cli.EncodingTypeJSON|cli.EncodingTypeHuman), "encoding",
		cli.EncodingTypeFlagDescription(cli.EncodingTypeJSON|cli.EncodingTypeHuman))

	// Add getMintConditionCmd to the ExploreCmd
	rootCmd.AddCommand(getMintConditionCmd)
}

type subCmd struct {
	pluginClient        *PluginClient
	getMintConditionCfg struct {
		EncodingType cli.EncodingType
	}
}

func (subCmds *subCmd) getMintCondition(cmd *cobra.Command, args []string) {
	pluginReader := subCmds.pluginClient

	var (
		mintCondition types.UnlockConditionProxy
		err           error
	)

	switch len(args) {
	case 0:
		// get active mint condition for the latest block height
		mintCondition, err = pluginReader.GetActiveMintCondition()
		if err != nil {
			cli.DieWithError("failed to get the active mint condition", err)
		}

	case 1:
		// get active mint condition for a given block height
		height, err := strconv.ParseUint(args[0], 10, 64)
		if err != nil {
			cmd.UsageFunc()(cmd)
			cli.DieWithError("invalid block height given", err)
		}
		mintCondition, err = pluginReader.GetMintConditionAt(types.BlockHeight(height))
		if err != nil {
			cli.DieWithError("failed to get the mint condition at the given block height", err)
		}

	default:
		cmd.UsageFunc()(cmd)
		cli.Die("Invalid amount of arguments. One optional pos argument can be given, a valid block height.")
	}

	// encode depending on the encoding flag
	var encode func(interface{}) error
	switch subCmds.getMintConditionCfg.EncodingType {
	case cli.EncodingTypeHuman:
		e := json.NewEncoder(os.Stdout)
		e.SetIndent("", "  ")
		encode = e.Encode
	case cli.EncodingTypeJSON:
		encode = json.NewEncoder(os.Stdout).Encode
	}
	err = encode(mintCondition)
	if err != nil {
		cli.DieWithError("failed to encode mint condition", err)
	}
}
