package consensus

import (
	"errors"

	bolt "github.com/rivine/bbolt"
	"github.com/threefoldtech/rivine/build"
	"github.com/threefoldtech/rivine/modules"
	"github.com/threefoldtech/rivine/pkg/encoding/siabin"
	"github.com/threefoldtech/rivine/types"
)

var (
	errOutputAlreadyMature = errors.New("delayed coin output is already in the matured outputs set")
	errPayoutsAlreadyPaid  = errors.New("payouts are already in the consensus set")
)

// applyMinerPayouts adds a block's miner payouts to the consensus set as
// delayed coin outputs.
func (cs *ConsensusSet) applyMinerPayouts(tx *bolt.Tx, pb *processedBlock) {
	for i := range pb.Block.MinerPayouts {
		mpid := pb.Block.MinerPayoutID(uint64(i))
		dscod := modules.DelayedCoinOutputDiff{
			Direction: modules.DiffApply,
			ID:        mpid,
			CoinOutput: types.CoinOutput{
				Value: pb.Block.MinerPayouts[i].Value,
				Condition: types.UnlockConditionProxy{
					Condition: types.NewUnlockHashCondition(pb.Block.MinerPayouts[i].UnlockHash),
				},
			},
			MaturityHeight: pb.Height + cs.chainCts.MaturityDelay,
		}
		pb.DelayedCoinOutputDiffs = append(pb.DelayedCoinOutputDiffs, dscod)
		commitDelayedCoinOutputDiff(tx, dscod, modules.DiffApply)
	}
}

// applyMaturedCoinOutputs goes through the list of coin outputs that
// have matured and adds them to the consensus set. This also updates the block
// node diff set.
func applyMaturedCoinOutputs(tx *bolt.Tx, pb *processedBlock) {
	// Iterate through the list of delayed coin outputs. Sometimes boltdb
	// has trouble if you delete elements in a bucket while iterating through
	// the bucket (and sometimes not - nondeterministic), so all of the
	// elements are collected into an array and then deleted after the bucket
	// scan is complete.
	pbHeightBytes, err := siabin.Marshal(pb.Height)
	if err != nil {
		build.Severe(err)
	}
	bucketID := append(prefixDCO, pbHeightBytes...)
	var scods []modules.CoinOutputDiff
	var dscods []modules.DelayedCoinOutputDiff
	bucket := tx.Bucket(bucketID)
	if bucket == nil {
		// No delayed coin output bucket for this height
		return
	}
	dbErr := bucket.ForEach(func(idBytes, scoBytes []byte) error {
		// Decode the key-value pair into an id and a coin output.
		var id types.CoinOutputID
		var sco types.CoinOutput
		copy(id[:], idBytes)
		encErr := siabin.Unmarshal(scoBytes, &sco)
		if encErr != nil {
			build.Severe(encErr)
		}

		// Sanity check - the output should not already be in coinOuptuts.
		if isCoinOutput(tx, id) {
			build.Severe(errOutputAlreadyMature)
		}

		// Add the output to the ConsensusSet and record the diff in the
		// blockNode.
		scod := modules.CoinOutputDiff{
			Direction:  modules.DiffApply,
			ID:         id,
			CoinOutput: sco,
		}
		scods = append(scods, scod)

		// Create the dscod and add it to the list of dscods that should be
		// deleted.
		dscod := modules.DelayedCoinOutputDiff{
			Direction:      modules.DiffRevert,
			ID:             id,
			CoinOutput:     sco,
			MaturityHeight: pb.Height,
		}
		dscods = append(dscods, dscod)
		return nil
	})
	if dbErr != nil {
		build.Severe(dbErr)
	}
	for _, scod := range scods {
		pb.CoinOutputDiffs = append(pb.CoinOutputDiffs, scod)
		commitCoinOutputDiff(tx, scod, modules.DiffApply)
	}
	for _, dscod := range dscods {
		pb.DelayedCoinOutputDiffs = append(pb.DelayedCoinOutputDiffs, dscod)
		commitDelayedCoinOutputDiff(tx, dscod, modules.DiffApply)
	}
	deleteDCOBucket(tx, pb.Height)
}

// applyMaintenance applies block-level alterations to the consensus set.
// Maintenance is applied after all of the transactions for the block have been
// applied.
func (cs *ConsensusSet) applyMaintenance(tx *bolt.Tx, pb *processedBlock) {
	cs.applyMinerPayouts(tx, pb)
	applyMaturedCoinOutputs(tx, pb)
}
