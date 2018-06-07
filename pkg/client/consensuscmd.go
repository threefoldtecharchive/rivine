package client

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/rivine/rivine/api"
	"github.com/rivine/rivine/encoding"
	"github.com/rivine/rivine/pkg/cli"
	"github.com/rivine/rivine/types"
	"github.com/spf13/cobra"
)

var (
	consensusCmd = &cobra.Command{
		Use:   "consensus",
		Short: "Print the current state of consensus",
		Long:  "Print the current state of consensus such as current block, block height, and target.",
		Run:   Wrap(consensuscmd),
	}

	consensusTransactionCmd = &cobra.Command{
		Use:   "transaction <shortID>|<longID>",
		Short: "Get an existing transaction",
		Long:  "Get an existing transaction from the blockchain, using its given shortID or longID.",
		Run:   Wrap(consensustransactioncmd),
	}
)

var (
	consensusTransactioncfg struct {
		EncodingType cli.EncodingType
	}
)

// Consensuscmd is the handler for the command `rivinec consensus`.
// Prints the current state of consensus.
func consensuscmd() {
	var cg api.ConsensusGET
	err := _DefaultClient.httpClient.GetAPI("/consensus", &cg)
	if err != nil {
		Die("Could not get current consensus state:", err)
	}
	if cg.Synced {
		fmt.Printf(`Synced: %v
Block:  %v
Height: %v
Target: %v
`, YesNo(cg.Synced), cg.CurrentBlock, cg.Height, cg.Target)
	} else {
		estimatedHeight := EstimatedHeightAt(time.Now())
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
func EstimatedHeightAt(t time.Time) types.BlockHeight {
	if _GenesisBlockTimestamp == 0 {
		panic("GenesisBlockTimestamp is undefined")
	}
	return estimatedHeightBetween(int64(_GenesisBlockTimestamp), t.Unix(), _BlockFrequencyInSeconds)
}

func estimatedHeightBetween(from, to, blockFrequency int64) types.BlockHeight {
	lifetimeInSeconds := to - from
	if lifetimeInSeconds < blockFrequency {
		return 0
	}
	estimatedHeight := float64(lifetimeInSeconds) / float64(blockFrequency)
	return types.BlockHeight(estimatedHeight + 0.5) // round to the nearest block
}

// consensustransactioncmd is the handler for the command `rivinec consensus transaction`.
// Prints the transaction found for the given id. If the ID is a long transaction ID, it also
// prints the short transaction ID for future reference
func consensustransactioncmd(id string) {
	var txn api.ConsensusGetTransaction

	err := _DefaultClient.httpClient.GetAPI("/consensus/transactions/"+id, &txn)
	if err != nil {
		Die("failed to get transaction:", err, "; ID:", id)
	}

	var encode func(interface{}) error
	switch consensusTransactioncfg.EncodingType {
	case cli.EncodingTypeHuman:
		e := json.NewEncoder(os.Stdout)
		e.SetIndent("", "  ")
		encode = e.Encode
	case cli.EncodingTypeJSON:
		encode = json.NewEncoder(os.Stdout).Encode
	case cli.EncodingTypeHex:
		encode = func(v interface{}) error {
			b := encoding.Marshal(v)
			fmt.Println(hex.EncodeToString(b))
			return nil
		}
	}

	err = encode(txn)
	if err != nil {
		Die("failed to encode transaction:", err, "; ID:", id)
	}
}

func init() {
	consensusTransactionCmd.Flags().Var(
		cli.NewEncodingTypeFlag(0, &consensusTransactioncfg.EncodingType, 0), "encoding",
		cli.EncodingTypeFlagDescription(0))
}
