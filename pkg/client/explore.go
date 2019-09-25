package client

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/threefoldtech/rivine/pkg/api"
	"github.com/threefoldtech/rivine/pkg/cli"
	"github.com/threefoldtech/rivine/pkg/encoding/siabin"
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
	)
	rootCmd.AddCommand(blockCmd, hashCmd)

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
}

// blockCmd is the handler for the command `rivinec explore block`,
// explores a block on the blockchain, by looking it up by its height,
// and printing either all info, or just the raw block itself.
func (cmd *exploreCmd) blockCmd(blockHeightStr string) {
	// get the block on the given height, using the daemon's explorer module
	var resp api.ExplorerBlockGET
	err := cmd.cli.GetWithResponse("/explorer/blocks/"+blockHeightStr, &resp)
	if err != nil {
		cli.Die(fmt.Sprintf("Could not get a block on height %q: %v", blockHeightStr, err))
	}

	// define the value to print
	value := interface{}(resp.Block)
	if cmd.blockCfg.BlockOnly {
		value = resp.Block.RawBlock
	}

	// print depending on the encoding type
	switch cmd.blockCfg.EncodingType {
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
func (cmd *exploreCmd) hashCmd(hash string) {
	// get the block on the given height, using the daemon's explorer module
	var resp api.ExplorerHashGET
	url := "/explorer/hashes/" + hash
	if cmd.hashCfg.MinHeight > 0 {
		url += fmt.Sprintf("?minheight=%d", cmd.hashCfg.MinHeight)
	}
	err := cmd.cli.GetWithResponse(url, &resp)
	if err != nil {
		cli.Die(fmt.Sprintf("Could not get an item using the hash %q: %v", hash, err))
	}

	// print depending on the encoding type
	switch cmd.hashCfg.EncodingType {
	case cli.EncodingTypeJSON:
		json.NewEncoder(os.Stdout).Encode(resp)
	default:
		e := json.NewEncoder(os.Stdout)
		e.SetIndent("", "  ")
		e.Encode(resp)
	}
}
