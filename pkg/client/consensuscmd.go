package client

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"time"

	"github.com/spf13/cobra"

	"github.com/rivine/rivine/api"
	"github.com/rivine/rivine/types"
)

var (
	consensusCmd = &cobra.Command{
		Use:   "consensus",
		Short: "Print the current state of consensus",
		Long:  "Print the current state of consensus such as current block, block height, and target.",
		Run:   wrap(Consensuscmd),
	}

	consensusTransactionCmd = &cobra.Command{
		Use:   "transaction <shortID>",
		Short: "Get an existing transaction",
		Long:  "Get an existing transaction from the blockchain, using its given shortID.",
		Run:   wrap(Consensustransactioncmd),
	}
)

// Consensuscmd is the handler for the command `rivinec consensus`.
// Prints the current state of consensus.
func Consensuscmd() {
	var cg api.ConsensusGET
	err := GetAPI("/consensus", &cg)
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
Progress (estimated): %.f%%
`, YesNo(cg.Synced), cg.Height, estimatedProgress)
	}
}

// EstimatedHeightAt returns the estimated block height for the given time.
// Block height is estimated by calculating the minutes since a known block in
// the past and dividing by 10 minutes (the block time).
func EstimatedHeightAt(t time.Time) types.BlockHeight {
	// Calc the amount of time passed since the Genesis block
	genesisTime := time.Unix(int64(types.GenesisTimestamp), 0)
	diff := t.Sub(genesisTime)
	// Estimated number of blocks, at a rate of 10/minute
	estimatedHeight := (diff.Minutes() / 10)
	return types.BlockHeight(estimatedHeight)
}

// Consensustransactioncmd is the handler for the command `rivinec consensus transaction`.
// Prints the transaction found for the given shortID.
func Consensustransactioncmd(shortID string) {
	if !consensusShortIDRegexp.MatchString(shortID) {
		Die("invalid shortID: ", shortID)
	}

	var txn api.ConsensusGetTransaction
	err := GetAPI("/consensus/transactions/"+shortID, &txn)
	if err != nil {
		Die("failed to get transaction: ", err, "; shortID: ", shortID)
	}

	encoder := json.NewEncoder(os.Stdout)
	err = encoder.Encode(txn)
	if err != nil {
		Die("failed to encode transaction: ", err, "; shortID: ", shortID)
	}
}

var (
	consensusShortIDRegexp = regexp.MustCompile("^[0-9]{1,8}$")
)
