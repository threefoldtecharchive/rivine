package consensus

import (
	"errors"

	"github.com/rivine/rivine/build"
	"github.com/rivine/rivine/encoding"
	"github.com/rivine/rivine/modules"
	"github.com/rivine/rivine/types"

	"github.com/NebulousLabs/bolt"
)

var (
	errMissingFileContract = errors.New("storage proof submitted for non existing file contract")
	errOutputAlreadyMature = errors.New("delayed siacoin output is already in the matured outputs set")
	errPayoutsAlreadyPaid  = errors.New("payouts are already in the consensus set")
	errStorageProofTiming  = errors.New("missed proof triggered for file contract that is not expiring")
)

// applyMinerPayouts adds a block's miner payouts to the consensus set as
// delayed siacoin outputs.
func applyMinerPayouts(tx *bolt.Tx, pb *processedBlock) {
	for i := range pb.Block.MinerPayouts {
		mpid := pb.Block.MinerPayoutID(uint64(i))
		dscod := modules.DelayedSiacoinOutputDiff{
			Direction:      modules.DiffApply,
			ID:             mpid,
			SiacoinOutput:  pb.Block.MinerPayouts[i],
			MaturityHeight: pb.Height + types.MaturityDelay,
		}
		pb.DelayedSiacoinOutputDiffs = append(pb.DelayedSiacoinOutputDiffs, dscod)
		commitDelayedSiacoinOutputDiff(tx, dscod, modules.DiffApply)
	}
}

// applyMaturedSiacoinOutputs goes through the list of siacoin outputs that
// have matured and adds them to the consensus set. This also updates the block
// node diff set.
func applyMaturedSiacoinOutputs(tx *bolt.Tx, pb *processedBlock) {
	// Skip this step if the blockchain is not old enough to have maturing
	// outputs.
	if pb.Height < types.MaturityDelay {
		return
	}

	// Iterate through the list of delayed siacoin outputs. Sometimes boltdb
	// has trouble if you delete elements in a bucket while iterating through
	// the bucket (and sometimes not - nondeterministic), so all of the
	// elements are collected into an array and then deleted after the bucket
	// scan is complete.
	bucketID := append(prefixDSCO, encoding.Marshal(pb.Height)...)
	var scods []modules.SiacoinOutputDiff
	var dscods []modules.DelayedSiacoinOutputDiff
	dbErr := tx.Bucket(bucketID).ForEach(func(idBytes, scoBytes []byte) error {
		// Decode the key-value pair into an id and a siacoin output.
		var id types.SiacoinOutputID
		var sco types.SiacoinOutput
		copy(id[:], idBytes)
		encErr := encoding.Unmarshal(scoBytes, &sco)
		if build.DEBUG && encErr != nil {
			panic(encErr)
		}

		// Sanity check - the output should not already be in siacoinOuptuts.
		if build.DEBUG && isSiacoinOutput(tx, id) {
			panic(errOutputAlreadyMature)
		}

		// Add the output to the ConsensusSet and record the diff in the
		// blockNode.
		scod := modules.SiacoinOutputDiff{
			Direction:     modules.DiffApply,
			ID:            id,
			SiacoinOutput: sco,
		}
		scods = append(scods, scod)

		// Create the dscod and add it to the list of dscods that should be
		// deleted.
		dscod := modules.DelayedSiacoinOutputDiff{
			Direction:      modules.DiffRevert,
			ID:             id,
			SiacoinOutput:  sco,
			MaturityHeight: pb.Height,
		}
		dscods = append(dscods, dscod)
		return nil
	})
	if build.DEBUG && dbErr != nil {
		panic(dbErr)
	}
	for _, scod := range scods {
		pb.SiacoinOutputDiffs = append(pb.SiacoinOutputDiffs, scod)
		commitSiacoinOutputDiff(tx, scod, modules.DiffApply)
	}
	for _, dscod := range dscods {
		pb.DelayedSiacoinOutputDiffs = append(pb.DelayedSiacoinOutputDiffs, dscod)
		commitDelayedSiacoinOutputDiff(tx, dscod, modules.DiffApply)
	}
	deleteDSCOBucket(tx, pb.Height)
}

// applyMaintenance applies block-level alterations to the consensus set.
// Maintenance is applied after all of the transcations for the block have been
// applied.
func applyMaintenance(tx *bolt.Tx, pb *processedBlock) {
	applyMinerPayouts(tx, pb)
	applyMaturedSiacoinOutputs(tx, pb)
}
