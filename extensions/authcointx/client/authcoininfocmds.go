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
	"github.com/threefoldtech/rivine/pkg/encoding/rivbin"
	types "github.com/threefoldtech/rivine/types"
)

// CreateExploreAuthCoinInfoCmd creates and attached the multi-use
// auth-coin-info fetch command to the root explore cmd.
func CreateExploreAuthCoinInfoCmd(cli *client.CommandLineClient) error {
	bc, err := client.NewLazyBaseClientFromCommandLineClient(cli)
	if err != nil {
		return err
	}
	createAuthCoinCmd(cli, NewPluginExplorerClient(bc), cli.ExploreCmd)
	return nil
}

// CreateConsensusAuthCoinInfoCmd creates and attached the multi-use
// auth-coin-info fetch command to the root consensus cmd.
func CreateConsensusAuthCoinInfoCmd(cli *client.CommandLineClient) error {
	bc, err := client.NewLazyBaseClientFromCommandLineClient(cli)
	if err != nil {
		return err
	}
	createAuthCoinCmd(cli, NewPluginConsensusClient(bc), cli.ConsensusCmd)
	return nil
}

func createAuthCoinCmd(client *client.CommandLineClient, pluginClient *PluginClient, rootCmd *cobra.Command) {
	authCoinCmd := &authCoinCmd{
		cli:          client,
		pluginClient: pluginClient,
	}

	// create root explore command and all subs
	var (
		getAuthInfo = &cobra.Command{
			Use:   "authcoin [condition|<address>] [height]",
			Short: "Get the active auth condition or address auth state",
			Long: `Get the active auth condition or address auth state.
If condition is used as a keyword the auth condition is looked up,
otherwise if an address is given instead the auth state for an address is fetched.
When a blockheight is specified it is done for a specific blockheight in mind,
otherwise it will look for the most recent condition/state known by the contacted daemon.
`,
			Args: cobra.RangeArgs(1, 2),
			Run:  authCoinCmd.getAuthInfo,
		}
	)

	getAuthInfo.Flags().Var(
		cli.NewEncodingTypeFlag(0, &authCoinCmd.getAuthInfoCfg.EncodingType, 0), "encoding",
		cli.EncodingTypeFlagDescription(0))

	// Add getAuthInfo to the Root Cmd
	rootCmd.AddCommand(getAuthInfo)
}

type authCoinCmd struct {
	cli            *client.CommandLineClient
	pluginClient   *PluginClient
	getAuthInfoCfg struct {
		EncodingType cli.EncodingType
	}
}

func (ac *authCoinCmd) getAuthInfo(cmd *cobra.Command, args []string) {
	if args[0] == "condition" {
		ac.getAuthCondition(cmd, args[1:])
	} else {
		var uh types.UnlockHash
		err := uh.LoadString(args[0])
		if err != nil {
			cmd.UsageFunc()(cmd)
			cli.DieWithError("invalid address cannot be authorized", err)
		}
		ac.getAddressAuthState(cmd, uh, args[1:])
	}
}

func (ac *authCoinCmd) getAuthCondition(cmd *cobra.Command, args []string) {
	var (
		err           error
		authCondition types.UnlockConditionProxy
	)

	switch len(args) {
	case 0:
		// get active auth condition for the latest block height
		authCondition, err = ac.pluginClient.GetActiveAuthCondition()
		if err != nil {
			cli.DieWithError("failed to get the active auth condition", err)
		}

	case 1:
		// get active auth condition for a given block height
		height, err := strconv.ParseUint(args[0], 10, 64)
		if err != nil {
			cmd.UsageFunc()(cmd)
			cli.DieWithError("invalid block height given", err)
		}
		authCondition, err = ac.pluginClient.GetAuthConditionAt(types.BlockHeight(height))
		if err != nil {
			cli.DieWithError("failed to get the auth condition at the given block height", err)
		}

	default:
		panic(fmt.Sprint("BUG: unexpected amount of args at this point", args))
	}

	// encode depending on the encoding flag
	var encode func(interface{}) error
	switch ac.getAuthInfoCfg.EncodingType {
	case cli.EncodingTypeHuman:
		e := json.NewEncoder(os.Stdout)
		e.SetIndent("", "  ")
		encode = e.Encode
	case cli.EncodingTypeJSON:
		encode = json.NewEncoder(os.Stdout).Encode
	case cli.EncodingTypeHex:
		encode = func(v interface{}) error {
			b, err := rivbin.Marshal(v)
			if err == nil {
				fmt.Println(hex.EncodeToString(b))
			}
			return err
		}
	}
	err = encode(map[string]interface{}{
		"condition": authCondition,
	})
	if err != nil {
		cli.DieWithError("failed to encode auth condition", err)
	}
}

func (ac *authCoinCmd) getAddressAuthState(cmd *cobra.Command, address types.UnlockHash, args []string) {
	var (
		err       error
		authState bool
	)

	switch len(args) {
	case 0:
		// get address auth state for the given address at the latest block height
		authState, err = ac.pluginClient.GetAddresAuthStateNow(address)
		if err != nil {
			cli.DieWithError("failed to fetch auth state for given address", err)
		}

	case 1:
		// get address auth state for the given address at the given block height
		height, err := strconv.ParseUint(args[0], 10, 64)
		if err != nil {
			cmd.UsageFunc()(cmd)
			cli.DieWithError("invalid block height given", err)
		}
		authState, err = ac.pluginClient.GetAddressAuthStateAt(types.BlockHeight(height), address)
		if err != nil {
			cli.DieWithError("failed to fetch auth state for given address", err)
		}

	default:
		panic(fmt.Sprint("BUG: unexpected amount of args at this point", args))
	}

	// encode depending on the encoding flag
	var encode func(interface{}) error
	switch ac.getAuthInfoCfg.EncodingType {
	case cli.EncodingTypeHuman:
		e := json.NewEncoder(os.Stdout)
		e.SetIndent("", "  ")
		encode = e.Encode
	case cli.EncodingTypeJSON:
		encode = json.NewEncoder(os.Stdout).Encode
	case cli.EncodingTypeHex:
		encode = func(v interface{}) error {
			b, err := rivbin.Marshal(v)
			if err == nil {
				fmt.Println(hex.EncodeToString(b))
			}
			return err
		}
	}
	err = encode(map[string]interface{}{
		"address": address.String(),
		"auth":    authState,
	})
	if err != nil {
		cli.DieWithError("failed to encode address auth state", err)
	}
}
