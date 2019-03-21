package client

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/threefoldtech/rivine/pkg/api"
	"github.com/threefoldtech/rivine/pkg/cli"
	"github.com/threefoldtech/rivine/pkg/encoding/siabin"
	types "github.com/threefoldtech/rivine/types"
)

func createExploreCmd(client *CommandLineClient) *cobra.Command {
	exploreCmd := &exploreCmd{cli: client}

	// create root explore command and all subs
	var (
		rootCmd = &cobra.Command{
			Use:   "explore",
			Short: "Explore the blockchain",
			Long:  "Explore the blockchain using the daemon's explorer module.",
		}
		blockCmd = &cobra.Command{
			Use:   "block <height>",
			Short: "Explore a block on the blockchain",
			Long:  "Explore a block on the blockchain, using its ID.",
			Run:   Wrap(exploreCmd.blockCmd),
		}
		hashCmd = &cobra.Command{
			Use:   "hash <unlockhash>|<transactionID>|<blockID>|<outputID>",
			Short: "Explore an item on the blockchain",
			Long:  "Explore an item on the blockchain, using its hash or ID.",
			Run:   Wrap(exploreCmd.hashCmd),
		}
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
	rootCmd.AddCommand(blockCmd, hashCmd, getMintConditionCmd)

	// create flags
	blockCmd.Flags().Var(
		cli.NewEncodingTypeFlag(0, &exploreCmd.blockCfg.EncodingType, cli.EncodingTypeHuman|cli.EncodingTypeJSON|cli.EncodingTypeHex), "encoding",
		cli.EncodingTypeFlagDescription(cli.EncodingTypeHuman|cli.EncodingTypeJSON|cli.EncodingTypeHex))
	blockCmd.Flags().BoolVar(
		&exploreCmd.blockCfg.BlockOnly, "block-only", false, "print the raw block only")

	hashCmd.Flags().Var(
		cli.NewEncodingTypeFlag(0, &exploreCmd.hashCfg.EncodingType, cli.EncodingTypeHuman|cli.EncodingTypeJSON), "encoding",
		cli.EncodingTypeFlagDescription(cli.EncodingTypeHuman|cli.EncodingTypeJSON))
	hashCmd.Flags().Uint64Var(
		&exploreCmd.hashCfg.MinHeight, "min-height", 0,
		"when looking up the transactions linked to an unlockhash, only show transactions since a given height")

	getMintConditionCmd.Flags().Var(
		cli.NewEncodingTypeFlag(0, &exploreCmd.getMintConditionCfg.EncodingType, 0), "encoding",
		cli.EncodingTypeFlagDescription(0))

	// return root command
	return rootCmd
}

type exploreCmd struct {
	cli      *CommandLineClient
	blockCfg struct {
		EncodingType cli.EncodingType
		BlockOnly    bool
	}
	hashCfg struct {
		EncodingType cli.EncodingType
		MinHeight    uint64
	}
	getMintConditionCfg struct {
		EncodingType cli.EncodingType
	}
}

// blockCmd is the handler for the command `rivinec explore block`,
// explores a block on the blockchain, by looking it up by its height,
// and printing either all info, or just the raw block itself.
func (explorerSubCmds *exploreCmd) blockCmd(blockHeightStr string) {
	// get the block on the given height, using the daemon's explorer module
	var resp api.ExplorerBlockGET
	err := explorerSubCmds.cli.GetAPI("/explorer/blocks/"+blockHeightStr, &resp)
	if err != nil {
		cli.Die(fmt.Sprintf("Could not get a block on height %q: %v", blockHeightStr, err))
	}

	// define the value to print
	value := interface{}(resp.Block)
	if explorerSubCmds.blockCfg.BlockOnly {
		value = resp.Block.RawBlock
	}

	// print depending on the encoding type
	switch explorerSubCmds.blockCfg.EncodingType {
	case cli.EncodingTypeHex:
		enc := siabin.NewEncoder(hex.NewEncoder(os.Stdout))
		enc.Encode(value)
		fmt.Println()
	case cli.EncodingTypeJSON:
		json.NewEncoder(os.Stdout).Encode(value)
	default:
		e := json.NewEncoder(os.Stdout)
		e.SetIndent("", "  ")
		e.Encode(value)
	}
}

// explorehashcmd is the handler for the command `rivinec explore hash`,
// explores an item on the blockchain, by looking it up by its hash,
// and printing all info it receives back for that hash
func (explorerSubCmds *exploreCmd) hashCmd(hash string) {
	// get the block on the given height, using the daemon's explorer module
	var resp api.ExplorerHashGET
	url := "/explorer/hashes/" + hash
	if explorerSubCmds.hashCfg.MinHeight > 0 {
		url += fmt.Sprintf("?minheight=%d", explorerSubCmds.hashCfg.MinHeight)
	}
	err := explorerSubCmds.cli.GetAPI(url, &resp)
	if err != nil {
		cli.Die(fmt.Sprintf("Could not get an item using the hash %q: %v", hash, err))
	}

	// print depending on the encoding type
	switch explorerSubCmds.hashCfg.EncodingType {
	case cli.EncodingTypeJSON:
		json.NewEncoder(os.Stdout).Encode(resp)
	default:
		e := json.NewEncoder(os.Stdout)
		e.SetIndent("", "  ")
		e.Encode(resp)
	}
}

func (explorerSubCmds *exploreCmd) getMintCondition(cmd *cobra.Command, args []string) {
	txDBReader := NewTransactionDBExplorerClient(explorerSubCmds.cli)

	var (
		mintCondition types.UnlockConditionProxy
		err           error
	)

	switch len(args) {
	case 0:
		// get active mint condition for the latest block height
		mintCondition, err = txDBReader.GetActiveMintCondition()
		if err != nil {
			cli.DieWithError("failed to get the active mint condition", err)
		}

	case 1:
		// get active mint condition for a given block height
		height, err := strconv.ParseUint(args[0], 10, 64)
		if err != nil {
			cmd.UsageFunc()
			cli.DieWithError("invalid block height given", err)
		}
		mintCondition, err = txDBReader.GetMintConditionAt(types.BlockHeight(height))
		if err != nil {
			cli.DieWithError("failed to get the mint condition at the given block height", err)
		}

	default:
		cmd.UsageFunc()
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
