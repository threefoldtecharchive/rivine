package client

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/threefoldtech/rivine/build"
	"github.com/threefoldtech/rivine/pkg/api"
	"github.com/threefoldtech/rivine/pkg/cli"
	"github.com/threefoldtech/rivine/pkg/encoding/siabin"
	"github.com/threefoldtech/rivine/types"
)

func createConsensusCmd(client *CommandLineClient) (*consensusCmd, *cobra.Command) {
	consensusCmd := &consensusCmd{cli: client}

	// create root consensus command and all subs
	var (
		rootCmd = &cobra.Command{
			Use:   "consensus",
			Short: "Print the current state of consensus",
			Long:  "Print the current state of consensus such as current block, block height, and target.",
			Run:   Wrap(consensusCmd.rootCmd),
		}
		transactionCmd = &cobra.Command{
			Use:   "transaction <shortID>|<longID>",
			Short: "Get an existing transaction",
			Long:  "Get an existing transaction from the blockchain, using its given shortID or longID.",
			Run:   Wrap(consensusCmd.transactionCmd),
		}
	)
	rootCmd.AddCommand(transactionCmd)

	// create flags
	transactionCmd.Flags().Var(
		cli.NewEncodingTypeFlag(0, &consensusCmd.transactionCfg.EncodingType, 0), "encoding",
		cli.EncodingTypeFlagDescription(0))

	// return root command
	return consensusCmd, rootCmd
}

type consensusCmd struct {
	cli            *CommandLineClient
	transactionCfg struct {
		EncodingType cli.EncodingType
	}
}

// rootCmd is the handler for the command `rivinec consensus`.
// Prints the current state of consensus.
func (consensusCmd *consensusCmd) rootCmd() {
	var cg api.ConsensusGET
	err := consensusCmd.cli.GetWithResponse("/consensus", &cg)
	if err != nil {
		cli.Die("Could not get current consensus state:", err)
	}
	if cg.Synced {
		fmt.Printf(`Synced: %v
Block:  %v
Height: %v
Target: %v
`, YesNo(cg.Synced), cg.CurrentBlock, cg.Height, cg.Target)
	} else {
		estimatedHeight := consensusCmd.estimatedHeightAt(time.Now())
		estimatedProgress := float64(cg.Height) / float64(estimatedHeight) * 100
		if estimatedProgress > 99 {
			estimatedProgress = 99
		}
		fmt.Printf(`Synced: %v
Height: %v
Progress (estimated): %.2f%%
`, YesNo(cg.Synced), cg.Height, estimatedProgress)
	}
}

// EstimatedHeightAt returns the estimated block height for the given time.
// Block height is estimated by calculating the minutes since a known block in
// the past and dividing by 10 minutes (the block time).
func (consensusCmd *consensusCmd) estimatedHeightAt(t time.Time) types.BlockHeight {
	if consensusCmd.cli.Config.GenesisBlockTimestamp == 0 {
		build.Critical("GenesisBlockTimestamp is undefined")
	}
	return estimatedHeightBetween(
		int64(consensusCmd.cli.Config.GenesisBlockTimestamp),
		t.Unix(),
		consensusCmd.cli.Config.BlockFrequencyInSeconds,
	)
}

func estimatedHeightBetween(from, to, blockFrequency int64) types.BlockHeight {
	lifetimeInSeconds := to - from
	if lifetimeInSeconds < blockFrequency {
		return 0
	}
	estimatedHeight := float64(lifetimeInSeconds) / float64(blockFrequency)
	return types.BlockHeight(estimatedHeight + 0.5) // round to the nearest block
}

// transactionCmd is the handler for the command `rivinec consensus transaction`.
// Prints the transaction found for the given id. If the ID is a long transaction ID, it also
// prints the short transaction ID for future reference
func (consensusCmd *consensusCmd) transactionCmd(id string) {
	var txn api.ConsensusGetTransaction

	err := consensusCmd.cli.GetWithResponse("/consensus/transactions/"+id, &txn)
	if err != nil {
		cli.Die("failed to get transaction:", err, "; ID:", id)
	}

	var encode func(interface{}) error
	switch consensusCmd.transactionCfg.EncodingType {
	case cli.EncodingTypeHuman:
		e := json.NewEncoder(os.Stdout)
		e.SetIndent("", "  ")
		encode = e.Encode
	case cli.EncodingTypeJSON:
		encode = json.NewEncoder(os.Stdout).Encode
	case cli.EncodingTypeHex:
		encode = func(v interface{}) error {
			b, err := siabin.Marshal(v)
			if err == nil {
				fmt.Println(hex.EncodeToString(b))
			}
			return err
		}
	}

	err = encode(txn)
	if err != nil {
		cli.Die("failed to encode transaction:", err, "; ID:", id)
	}
}
