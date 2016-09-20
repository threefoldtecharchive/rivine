package consensus

// applytransaction.go handles applying a transaction to the consensus set.
// There is an assumption that the transaction has already been verified.

import (
	"github.com/rivine/rivine/build"
	"github.com/rivine/rivine/modules"
	"github.com/rivine/rivine/types"

	"github.com/NebulousLabs/bolt"
)

// applySiacoinInputs takes all of the siacoin inputs in a transaction and
// applies them to the state, updating the diffs in the processed block.
func applySiacoinInputs(tx *bolt.Tx, pb *processedBlock, t types.Transaction) {
	// Remove all siacoin inputs from the unspent siacoin outputs list.
	for _, sci := range t.SiacoinInputs {
		sco, err := getSiacoinOutput(tx, sci.ParentID)
		if build.DEBUG && err != nil {
			panic(err)
		}
		scod := modules.SiacoinOutputDiff{
			Direction:     modules.DiffRevert,
			ID:            sci.ParentID,
			SiacoinOutput: sco,
		}
		pb.SiacoinOutputDiffs = append(pb.SiacoinOutputDiffs, scod)
		commitSiacoinOutputDiff(tx, scod, modules.DiffApply)
	}
}

// applySiacoinOutputs takes all of the siacoin outputs in a transaction and
// applies them to the state, updating the diffs in the processed block.
func applySiacoinOutputs(tx *bolt.Tx, pb *processedBlock, t types.Transaction) {
	// Add all siacoin outputs to the unspent siacoin outputs list.
	for i, sco := range t.SiacoinOutputs {
		scoid := t.SiacoinOutputID(uint64(i))
		scod := modules.SiacoinOutputDiff{
			Direction:     modules.DiffApply,
			ID:            scoid,
			SiacoinOutput: sco,
		}
		pb.SiacoinOutputDiffs = append(pb.SiacoinOutputDiffs, scod)
		commitSiacoinOutputDiff(tx, scod, modules.DiffApply)
	}
}

// applyTxSiafundInputs takes all of the siafund inputs in a transaction and
// applies them to the state, updating the diffs in the processed block.
func applySiafundInputs(tx *bolt.Tx, pb *processedBlock, t types.Transaction) {
	for _, sfi := range t.SiafundInputs {
		// Calculate the volume of siacoins to put in the claim output.
		sfo, err := getSiafundOutput(tx, sfi.ParentID)
		if build.DEBUG && err != nil {
			panic(err)
		}

		// Create the siafund output diff and remove the output from the
		// consensus set.
		sfod := modules.SiafundOutputDiff{
			Direction:     modules.DiffRevert,
			ID:            sfi.ParentID,
			SiafundOutput: sfo,
		}
		pb.SiafundOutputDiffs = append(pb.SiafundOutputDiffs, sfod)
		commitSiafundOutputDiff(tx, sfod, modules.DiffApply)
	}
}

// applySiafundOutput applies a siafund output to the consensus set.
func applySiafundOutputs(tx *bolt.Tx, pb *processedBlock, t types.Transaction) {
	for i, sfo := range t.SiafundOutputs {
		sfoid := t.SiafundOutputID(uint64(i))
		sfod := modules.SiafundOutputDiff{
			Direction:     modules.DiffApply,
			ID:            sfoid,
			SiafundOutput: sfo,
		}
		pb.SiafundOutputDiffs = append(pb.SiafundOutputDiffs, sfod)
		commitSiafundOutputDiff(tx, sfod, modules.DiffApply)
	}
}

// applyTransaction applies the contents of a transaction to the ConsensusSet.
// This produces a set of diffs, which are stored in the blockNode containing
// the transaction. No verification is done by this function.
func applyTransaction(tx *bolt.Tx, pb *processedBlock, t types.Transaction) {
	applySiacoinInputs(tx, pb, t)
	applySiacoinOutputs(tx, pb, t)
	applySiafundInputs(tx, pb, t)
	applySiafundOutputs(tx, pb, t)
}
