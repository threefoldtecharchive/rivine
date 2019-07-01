package client

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/threefoldtech/rivine/pkg/cli"
	client "github.com/threefoldtech/rivine/pkg/client"
	"github.com/threefoldtech/rivine/pkg/encoding/siabin"
	types "github.com/threefoldtech/rivine/types"
)

func CreateExploreCmd(client *client.CommandLineClient) {
	createExploreCmd(client, client.ExploreCmd, NewPluginExplorerClient(client))
}

func CreateConsensusCmd(client *client.CommandLineClient) {
	createExploreCmd(client, client.ConsensusCmd, NewPluginConsensusClient(client))
}

func createExploreCmd(client *client.CommandLineClient, rootCmd *cobra.Command, pluginClient *PluginClient) {
	exploreCmd := &exploreCmd{
		cli:          client,
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
			Run: exploreCmd.getMintCondition,
		}
	)

	getMintConditionCmd.Flags().Var(
		cli.NewEncodingTypeFlag(0, &exploreCmd.getMintConditionCfg.EncodingType, 0), "encoding",
		cli.EncodingTypeFlagDescription(0))

	// Add getMintConditionCmd to the ExploreCmd
	rootCmd.AddCommand(getMintConditionCmd)
}

type exploreCmd struct {
	cli                 *client.CommandLineClient
	pluginClient        *PluginClient
	getMintConditionCfg struct {
		EncodingType cli.EncodingType
	}
}

func (explorerSubCmds *exploreCmd) getMintCondition(cmd *cobra.Command, args []string) {
	pluginReader := NewPluginExplorerClient(explorerSubCmds.cli)

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
	switch explorerSubCmds.getMintConditionCfg.EncodingType {
	case cli.EncodingTypeHuman:
		e := json.NewEncoder(os.Stdout)
		e.SetIndent("", "  ")
		encode = e.Encode
	case cli.EncodingTypeJSON:
		encode = json.NewEncoder(os.Stdout).Encode
	case cli.EncodingTypeHex:
		encode = func(v interface{}) error {
			b := siabin.Marshal(v)
			fmt.Println(hex.EncodeToString(b))
			return nil
		}
	}
	err = encode(mintCondition)
	if err != nil {
		cli.DieWithError("failed to encode mint condition", err)
	}
}
