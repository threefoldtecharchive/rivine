package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"

	"github.com/rivine/rivine/types"
	"github.com/spf13/cobra"
)

func createMergeCommands() {
	mergeCmd = &cobra.Command{
		Use:   "merge",
		Short: "merge transaction inputs",
		// Run field is not set, as the create command itself is not a valid command.
		// A subcommand must be provided.
	}

	mergeTransactionCmd = &cobra.Command{
		Use:   "transaction <txnjson1> <txnjson2>",
		Short: "Merge compatible input fulfillments",
		Long:  "Merge the compatible input fulfillments from the 2 transactions together. Currently only multisignature inputs can be merged.",
		Args:  cobra.ExactArgs(2),
		Run:   Wrap(walletmergetransactions),
	}
}

var (
	mergeCmd            *cobra.Command
	mergeTransactionCmd *cobra.Command
)

func walletmergetransactions(txn1 string, txn2 string) {
	buff1 := bytes.NewBufferString(txn1)
	buff2 := bytes.NewBufferString(txn2)

	var tx1, tx2 types.TransactionInputs
	err := json.NewDecoder(buff1).Decode(&tx1)
	if err != nil {
		Die("Failed to decode transaction 1:", err)
	}
	err = json.NewDecoder(buff2).Decode(&tx2)
	if err != nil {
		Die("Failed to decode transaction 2:", err)
	}

	if tx1.Version != tx2.Version {
		Die("Transaction versions don't match")
	}
	if bytes.Compare(tx1.Data.CoinOutputs, tx2.Data.CoinOutputs) != 0 {
		Die("Transaction coin outputs do not match")
	}
	if bytes.Compare(tx1.Data.BlockStakeOutputs, tx2.Data.BlockStakeOutputs) != 0 {
		Die("Transaction block stake outputs do not match")
	}
	if bytes.Compare(tx1.Data.ArbitraryData, tx2.Data.ArbitraryData) != 0 {
		Die("Transaction arbitrary data does not match")
	}
	if len(tx1.Data.CoinInputs) != len(tx2.Data.CoinInputs) {
		Die("Transactions have a different amount of coin inputs")
	}
	if len(tx1.Data.BlockStakeInputs) != len(tx2.Data.BlockStakeInputs) {
		Die("Transactions have a different amount of blockstake inputs")
	}
	for i := range tx1.Data.CoinInputs {
		if bytes.Compare(tx1.Data.CoinInputs[i].ParentID[:], tx2.Data.CoinInputs[i].ParentID[:]) != 0 {
			Die(fmt.Sprintf("Transactions have a different coin input at index %v", i))
		}
		if tx1.Data.CoinInputs[i].Fulfillment.FulfillmentType() == types.FulfillmentTypeMultiSignature &&
			tx2.Data.CoinInputs[i].Fulfillment.FulfillmentType() == types.FulfillmentTypeMultiSignature {
			ff1, ok := tx1.Data.CoinInputs[i].Fulfillment.Fulfillment.(*types.MultiSignatureFulfillment)
			if !ok {
				// Shouldn't happen
				Die("Failed to assert multisignature fulfillment")
			}
			ff2, ok := tx2.Data.CoinInputs[i].Fulfillment.Fulfillment.(*types.MultiSignatureFulfillment)
			if !ok {
				// Shouldn't happen
				Die("Failed to assert multisignature fulfillment")
			}
			newPairs := make([]types.PublicKeySignaturePair, len(ff2.Pairs))
			copy(newPairs, ff2.Pairs)
			for _, pksp := range ff1.Pairs {
				for i, newPksp := range newPairs {
					// Check for equality. If the pair matches, remove it so we don't add it
					if pksp.PublicKey.Algorithm == newPksp.PublicKey.Algorithm &&
						bytes.Compare(pksp.PublicKey.Key, newPksp.PublicKey.Key) == 0 &&
						bytes.Compare(pksp.Signature, newPksp.Signature) == 0 {

						newPairs = append(newPairs[:i], newPairs[i+1:]...)
						break
					}
				}
			}
			ff1.Pairs = append(ff1.Pairs, newPairs...)
		}
	}
	for i := range tx1.Data.BlockStakeInputs {
		if bytes.Compare(tx1.Data.BlockStakeInputs[i].ParentID[:], tx2.Data.BlockStakeInputs[i].ParentID[:]) != 0 {
			Die(fmt.Sprintf("Transactions have a different blockstake input input at index %v", i))
		}
		if tx1.Data.BlockStakeInputs[i].Fulfillment.FulfillmentType() == types.FulfillmentTypeMultiSignature &&
			tx2.Data.BlockStakeInputs[i].Fulfillment.FulfillmentType() == types.FulfillmentTypeMultiSignature {
			ff1, ok := tx1.Data.BlockStakeInputs[i].Fulfillment.Fulfillment.(*types.MultiSignatureFulfillment)
			if !ok {
				// Shouldn't happen
				Die("Failed to assert multisignature fulfillment")
			}
			ff2, ok := tx2.Data.BlockStakeInputs[i].Fulfillment.Fulfillment.(*types.MultiSignatureFulfillment)
			if !ok {
				// Shouldn't happen
				Die("Failed to assert multisignature fulfillment")
			}
			newPairs := make([]types.PublicKeySignaturePair, len(ff2.Pairs))
			copy(newPairs, ff2.Pairs)
			for _, pksp := range ff1.Pairs {
				for i, newPksp := range newPairs {
					// Check for equality. If the pair matches, remove it so we don't add it
					if pksp.PublicKey.Algorithm == newPksp.PublicKey.Algorithm &&
						bytes.Compare(pksp.PublicKey.Key, newPksp.PublicKey.Key) == 0 &&
						bytes.Compare(pksp.Signature, newPksp.Signature) == 0 {

						newPairs = append(newPairs[:i], newPairs[i+1:]...)
						break
					}
				}
			}
			ff1.Pairs = append(ff1.Pairs, newPairs...)
		}
	}
	json.NewEncoder(os.Stdout).Encode(tx1)
}
