package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/threefoldtech/rivine/pkg/cli"
	"github.com/threefoldtech/rivine/types"
	"github.com/spf13/cobra"
)

func createMergeCmd(*CommandLineClient) *cobra.Command {
	mergeCmd := new(mergeCmd)

	// create root merge command and all subs
	var (
		rootCmd = &cobra.Command{
			Use:   "merge",
			Short: "merge transaction inputs",
			// Run field is not set, as the create command itself is not a valid command.
			// A subcommand must be provided.
		}
		mergeTxCmd = &cobra.Command{
			Use:   "transactions <txnjson1> <txnjson2> [txnjsonN...]",
			Short: "Merge compatible input fulfillments",
			Long: `Merge the compatible input fulfillments from two or more transactions together.
Currently only multisignature inputs can be merged.

Duplicate signatures are only deleted if there are more signatures for a public key given,
than that the condition defines for that public key's unlock hash.
`,
			Args: cobra.MinimumNArgs(2),
			Run:  mergeCmd.mergeTransactions,
		}
	)
	rootCmd.AddCommand(mergeTxCmd)

	// return root command
	return rootCmd
}

type mergeCmd struct{}

func (mergeCmd *mergeCmd) mergeTransactions(cmd *cobra.Command, args []string) {
	var masterTxn transactionInputs
	err := json.NewDecoder(bytes.NewBufferString(args[0])).Decode(&masterTxn)
	if err != nil {
		cli.Die("failed to decode transaction first transaction:", err)
	}

	// compare the master txn against all other txns,
	// assuming the first transaction is the correct one
	for idx, arg := range args[1:] {
		var (
			txnIndex = idx + 2
			otherTxn transactionInputs
		)

		err = json.NewDecoder(bytes.NewBufferString(arg)).Decode(&otherTxn)
		if err != nil {
			cli.Die(fmt.Sprintf("failed to decode transaction #%d: %v", txnIndex, err))
		}

		err = compareNonMergeableTransactionData(masterTxn, otherTxn)
		if err != nil {
			cli.Die(fmt.Sprintf("transaction #%d cannot be merged into the previous transaction(s): %v", txnIndex, err))
		}

		for i := range masterTxn.Data.CoinInputs {
			inputIndex := i + 1

			if bytes.Compare(masterTxn.Data.CoinInputs[i].ParentID[:], otherTxn.Data.CoinInputs[i].ParentID[:]) != 0 {
				cli.Die(fmt.Sprintf("transaction #%d has a different coin input at index %v", txnIndex, inputIndex))
			}

			err = compareAndMergeFulfillmentsIfNeeded(
				&masterTxn.Data.CoinInputs[i].Fulfillment, otherTxn.Data.CoinInputs[i].Fulfillment)
			if err != nil {
				cli.Die(fmt.Sprintf(
					"failed to compare and/or merge fulfillment of coin input #%d in transaction #%d: %v",
					inputIndex, txnIndex, err))
			}
		}
		for i := range masterTxn.Data.BlockStakeInputs {
			inputIndex := i + 1

			if bytes.Compare(masterTxn.Data.BlockStakeInputs[i].ParentID[:], otherTxn.Data.BlockStakeInputs[i].ParentID[:]) != 0 {
				cli.Die(fmt.Sprintf("transaction #%d has a different block stake input at index %v", txnIndex, inputIndex))
			}

			err = compareAndMergeFulfillmentsIfNeeded(
				&masterTxn.Data.BlockStakeInputs[i].Fulfillment, otherTxn.Data.BlockStakeInputs[i].Fulfillment)
			if err != nil {
				cli.Die(fmt.Sprintf(
					"failed to compare and/or merge fulfillment of blockstake input #%d in transaction #%d: %v",
					inputIndex, txnIndex, err))
			}
		}
	}

	// encode the merged result back as JSON
	json.NewEncoder(os.Stdout).Encode(masterTxn)
}

// compareNonMergeableTransactionData ensures that all the non-mergeable data of transactions
// is equal, this means that anything except for fulfillments which are mergeable, has to be equal.
func compareNonMergeableTransactionData(masterTxn, otherTxn transactionInputs) error {
	if masterTxn.Version != otherTxn.Version {
		return errors.New("transaction version is different")
	}
	if bytes.Compare(masterTxn.Data.CoinOutputs, otherTxn.Data.CoinOutputs) != 0 {
		return errors.New("coin outputs are different")
	}
	if bytes.Compare(masterTxn.Data.MinerFees, otherTxn.Data.MinerFees) != 0 {
		return errors.New("miner fees are different")
	}
	if bytes.Compare(masterTxn.Data.BlockStakeOutputs, otherTxn.Data.BlockStakeOutputs) != 0 {
		return errors.New("blockstake outputs are different")
	}
	if bytes.Compare(masterTxn.Data.ArbitraryData, otherTxn.Data.ArbitraryData) != 0 {
		return errors.New("arbitrary data different")
	}
	if len(masterTxn.Data.CoinInputs) != len(otherTxn.Data.CoinInputs) {
		return errors.New("coin input length is different")
	}
	if len(masterTxn.Data.BlockStakeInputs) != len(otherTxn.Data.BlockStakeInputs) {
		return errors.New("blockstake input length is different")
	}
	return nil
}

// compareAndMergeFulfillmentsIfNeeded ensures that non-mergeable fulfillments are equal,
// and that MultiSigFulfillments (only ones that are mergeable) can be merged.
func compareAndMergeFulfillmentsIfNeeded(masterFulfillment *types.UnlockFulfillmentProxy, otherFulfillment types.UnlockFulfillmentProxy) error {
	masterFT := masterFulfillment.FulfillmentType()
	otherFT := otherFulfillment.FulfillmentType()
	if masterFT != otherFT {
		return errors.New("different fulfillment type")
	}
	if masterFT != types.FulfillmentTypeMultiSignature {
		// if it isn't of the multisig type, the 2 input's fulfillment types have to be equal
		if !masterFulfillment.Equal(otherFulfillment) {
			return errors.New("different non-mergable fulfillment data")
		}
		return nil
	}
	// inputs both have a FulfillmentTypeMultisignature, time to merge, if possible
	ff1, ok := masterFulfillment.Fulfillment.(*types.MultiSignatureFulfillment)
	if !ok {
		// Shouldn't happen
		return fmt.Errorf("unexpected fulfillment type %T for master transaction", masterFulfillment.Fulfillment)
	}
	ff2, ok := otherFulfillment.Fulfillment.(*types.MultiSignatureFulfillment)
	if !ok {
		// Shouldn't happen
		return fmt.Errorf("unexpected fulfillment type %T for other transaction", otherFulfillment.Fulfillment)
	}
	// remove duplicate pairs,
	// we'll assume that each transaction is signed to the fully extend possible for that owner,
	// such that we can assume that all pairs which already exist in the master transaction,
	// can be seen as real duplicates, and thus removed
	for _, pksp := range ff1.Pairs {
		for i := 0; i < len(ff2.Pairs); i++ {
			newPksp := ff2.Pairs[i]
			// Check for equality. If the pair matches, remove it so we don't add it
			if pksp.PublicKey.Algorithm == newPksp.PublicKey.Algorithm &&
				bytes.Compare(pksp.PublicKey.Key, newPksp.PublicKey.Key) == 0 &&
				bytes.Compare(pksp.Signature, newPksp.Signature) == 0 {
				ff2.Pairs = append(ff2.Pairs[:i], ff2.Pairs[i+1:]...)
				break
			}
		}
	}
	ff1.Pairs = append(ff1.Pairs, ff2.Pairs...)
	return nil
}

type (
	// transactionInputs hold the transaction version and the transaction data. Only
	// the inputs are fully decoded
	transactionInputs struct {
		Version types.TransactionVersion `json:"version"`
		Data    transactionInputData     `json:"data"`
	}

	// transactionInputData are all relevant fields of a transaction, with only the coin
	// and blockstake inputs properly decoded. This allows for easy equality checking of
	// other fields
	transactionInputData struct {
		CoinInputs        []types.CoinInput       `json:"coininputs"`
		CoinOutputs       json.RawMessage         `json:"coinoutputs,omitempty"`
		BlockStakeInputs  []types.BlockStakeInput `json:"blockstakeinputs,omitempty"`
		BlockStakeOutputs json.RawMessage         `json:"blockstakeoutputs,omitempty"`
		MinerFees         json.RawMessage         `json:"minerfees,omitempty"`
		ArbitraryData     json.RawMessage         `json:"arbitrarydata,omitempty"`
	}
)
