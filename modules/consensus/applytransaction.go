package consensus

// applytransaction.go handles applying a transaction to the consensus set.
// There is an assumption that the transaction has already been verified.

import (
	"fmt"

	bolt "github.com/rivine/bbolt"
	"github.com/threefoldtech/rivine/build"
	"github.com/threefoldtech/rivine/modules"
	"github.com/threefoldtech/rivine/types"
)

// applyCoinInputs takes all of the coin inputs in a transaction and
// applies them to the state, updating the diffs in the processed block.
func applyCoinInputs(tx *bolt.Tx, pb *processedBlock, t types.Transaction) {
	// Remove all coin inputs from the unspent siacoin outputs list.
	for _, sci := range t.CoinInputs {
		sco, err := getCoinOutput(tx, sci.ParentID)
		if err != nil {
			build.Severe(fmt.Errorf("%s, coininput parentid: %s", err.Error(), sci.ParentID))
		}
		scod := modules.CoinOutputDiff{
			Direction:  modules.DiffRevert,
			ID:         sci.ParentID,
			CoinOutput: sco,
		}
		pb.CoinOutputDiffs = append(pb.CoinOutputDiffs, scod)
		commitCoinOutputDiff(tx, scod, modules.DiffApply)
	}
}

// applyCoinOutputs takes all of the coin outputs in a transaction and
// applies them to the state, updating the diffs in the processed block.
func applyCoinOutputs(tx *bolt.Tx, pb *processedBlock, t types.Transaction) {
	// Add all siacoin outputs to the unspent siacoin outputs list.
	for i, sco := range t.CoinOutputs {
		scoid := t.CoinOutputID(uint64(i))
		scod := modules.CoinOutputDiff{
			Direction:  modules.DiffApply,
			ID:         scoid,
			CoinOutput: sco,
		}
		pb.CoinOutputDiffs = append(pb.CoinOutputDiffs, scod)
		commitCoinOutputDiff(tx, scod, modules.DiffApply)
	}
}

// applyBlockStakeInputs takes all of the siafund inputs in a transaction and
// applies them to the state, updating the diffs in the processed block.
func applyBlockStakeInputs(tx *bolt.Tx, pb *processedBlock, t types.Transaction) {
	for _, sfi := range t.BlockStakeInputs {
		// Calculate the volume of siacoins to put in the claim output.
		sfo, err := getBlockStakeOutput(tx, sfi.ParentID)
		if err != nil {
			build.Severe(err)
		}

		// Create the siafund output diff and remove the output from the
		// consensus set.
		sfod := modules.BlockStakeOutputDiff{
			Direction:        modules.DiffRevert,
			ID:               sfi.ParentID,
			BlockStakeOutput: sfo,
		}
		pb.BlockStakeOutputDiffs = append(pb.BlockStakeOutputDiffs, sfod)
		commitBlockStakeOutputDiff(tx, sfod, modules.DiffApply)
	}
}

// applyBlockStakeOutput applies a siafund output to the consensus set.
func applyBlockStakeOutputs(tx *bolt.Tx, pb *processedBlock, t types.Transaction) {
	for i, sfo := range t.BlockStakeOutputs {
		sfoid := t.BlockStakeOutputID(uint64(i))
		sfod := modules.BlockStakeOutputDiff{
			Direction:        modules.DiffApply,
			ID:               sfoid,
			BlockStakeOutput: sfo,
		}
		pb.BlockStakeOutputDiffs = append(pb.BlockStakeOutputDiffs, sfod)
		commitBlockStakeOutputDiff(tx, sfod, modules.DiffApply)
	}
}

// applyTransactionIDMapping applies a transaction id mapping to the consensus set
func applyTransactionIDMapping(tx *bolt.Tx, pb *processedBlock, t types.Transaction) {
	tidmod := modules.TransactionIDDiff{
		Direction: modules.DiffApply,
		LongID:    t.ID(),
		ShortID:   types.NewTransactionShortID(pb.Height, uint16(len(pb.TxIDDiffs))),
	}
	pb.TxIDDiffs = append(pb.TxIDDiffs, tidmod)
	commitTxIDMapDiff(tx, tidmod, modules.DiffApply)
}

// applyTransaction applies the contents of a transaction to the ConsensusSet.
// This produces a set of diffs, which are stored in the blockNode containing
// the transaction. No verification is done by this function.
func applyTransaction(tx *bolt.Tx, pb *processedBlock, t types.Transaction) {
	applyCoinInputs(tx, pb, t)
	applyCoinOutputs(tx, pb, t)
	applyBlockStakeInputs(tx, pb, t)
	applyBlockStakeOutputs(tx, pb, t)
	applyTransactionIDMapping(tx, pb, t)
}
