package client

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/rivine/rivine/api"
	"github.com/rivine/rivine/pkg/cli"
	"github.com/spf13/cobra"
)

var (
	exploreCmd = &cobra.Command{
		Use:   "explore",
		Short: "Explore the blockchain",
		Long:  "Explore the blockchain using the daemon's explorer module.",
	}

	exploreBlockCmd = &cobra.Command{
		Use:   "block <height>",
		Short: "Explore a block on the blockchain",
		Long:  "Explore a block on the blockchain, using its ID.",
		Run:   Wrap(exploreblockcmd),
	}

	exploreHashCmd = &cobra.Command{
		Use:   "hash <unlockhash>|<transactionID>|<blockID>|<outputID>",
		Short: "Explore an item on the blockchain",
		Long:  "Explore an item on the blockchain, using its hash or ID.",
		Run:   Wrap(explorehashcmd),
	}
)

var (
	exploreBlockcfg struct {
		EncodingType cli.EncodingType
		BlockOnly    bool
	}
	exploreHashcfg struct {
		EncodingType cli.EncodingType
	}
)

// exploreblockcmd is the handler for the command `rivinec explore block`,
// explores a block on the blockchain, by looking it up by its height,
// and printing either all info, or just the raw block itself.
func exploreblockcmd(blockHeightStr string) {
	// get the block on the given height, using the daemon's explorer module
	var resp api.ExplorerBlockGET
	err := _DefaultClient.httpClient.GetAPI("/explorer/blocks/"+blockHeightStr, &resp)
	if err != nil {
		Die(fmt.Sprintf("Could not get a block on height %q: %v", blockHeightStr, err))
	}

	// define the value to print
	value := interface{}(resp.Block)
	if exploreBlockcfg.BlockOnly {
		value = resp.Block.RawBlock
	}

	// print depending on the encoding type
	switch exploreBlockcfg.EncodingType {
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
func explorehashcmd(hash string) {
	// get the block on the given height, using the daemon's explorer module
	var resp api.ExplorerHashGET
	err := _DefaultClient.httpClient.GetAPI("/explorer/hashes/"+hash, &resp)
	if err != nil {
		Die(fmt.Sprintf("Could not get an item using the hash %q: %v", hash, err))
	}

	// print depending on the encoding type
	switch exploreBlockcfg.EncodingType {
	case cli.EncodingTypeJSON:
		json.NewEncoder(os.Stdout).Encode(resp)
	default:
		e := json.NewEncoder(os.Stdout)
		e.SetIndent("", "  ")
		e.Encode(resp)
	}
}

func init() {
	exploreBlockCmd.Flags().Var(
		cli.NewEncodingTypeFlag(0, &exploreBlockcfg.EncodingType, cli.EncodingTypeHuman|cli.EncodingTypeJSON), "encoding",
		cli.EncodingTypeFlagDescription(cli.EncodingTypeHuman|cli.EncodingTypeJSON))
	exploreBlockCmd.Flags().BoolVar(
		&exploreBlockcfg.BlockOnly, "block-only", false, "print the raw block only")

	exploreHashCmd.Flags().Var(
		cli.NewEncodingTypeFlag(0, &exploreHashcfg.EncodingType, cli.EncodingTypeHuman|cli.EncodingTypeJSON), "encoding",
		cli.EncodingTypeFlagDescription(cli.EncodingTypeHuman|cli.EncodingTypeJSON))
}
