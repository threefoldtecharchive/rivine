package consensus

import (
	"errors"

	bolt "github.com/rivine/bbolt"
	"github.com/threefoldtech/rivine/build"
	"github.com/threefoldtech/rivine/modules"
	"github.com/threefoldtech/rivine/types"
)

var (
	errExternalRevert = errors.New("cannot revert to block outside of current path")
)

// backtrackToCurrentPath traces backwards from 'pb' until it reaches a block
// in the ConsensusSet's current path (the "common parent"). It returns the
// (inclusive) set of blocks between the common parent and 'pb', starting from
// the former.
func backtrackToCurrentPath(tx *bolt.Tx, pb *processedBlock) []*processedBlock {
	path := []*processedBlock{pb}
	for {
		// Error is not checked in production code - an error can only indicate
		// that pb.Height > blockHeight(tx).
		currentPathID, err := getPath(tx, pb.Height)
		if currentPathID == pb.Block.ID() {
			break
		}
		// Sanity check - an error should only indicate that pb.Height >
		// blockHeight(tx).
		if err != nil && pb.Height <= blockHeight(tx) {
			build.Severe(err)
		}

		// Prepend the next block to the list of blocks leading from the
		// current path to the input block.
		pb, err = getBlockMap(tx, pb.Block.ParentID)
		if err != nil {
			build.Severe(err)
		}
		path = append([]*processedBlock{pb}, path...)
	}
	return path
}

// revertToBlock will revert blocks from the ConsensusSet's current path until
// 'pb' is the current block. Blocks are returned in the order that they were
// reverted.  'pb' is not reverted.
func (cs *ConsensusSet) revertToBlock(tx *bolt.Tx, pb *processedBlock) (revertedBlocks []*processedBlock) {
	// Sanity check - make sure that pb is in the current path.
	currentPathID, err := getPath(tx, pb.Height)
	if err != nil || currentPathID != pb.Block.ID() {
		build.Severe(errExternalRevert)
	}

	// Rewind blocks until 'pb' is the current block.
	for currentBlockID(tx) != pb.Block.ID() {
		block := currentProcessedBlock(tx)
		err = cs.rewindBlock(tx, block)
		if err != nil {
			build.Severe(err)
		}
		revertedBlocks = append(revertedBlocks, block)

		// Sanity check - after removing a block, check that the consensus set
		// has maintained consistency.
		if build.Release == "testing" {
			cs.checkConsistency(tx)
		} else {
			cs.maybeCheckConsistency(tx)
		}
	}
	return revertedBlocks
}

// applyUntilBlock will successively apply the blocks between the consensus
// set's current path and 'pb'.
func (cs *ConsensusSet) applyUntilBlock(tx *bolt.Tx, pb *processedBlock) (appliedBlocks []*processedBlock, err error) {
	// Backtrack to the common parent of 'bn' and current path and then apply the new blocks.
	newPath := backtrackToCurrentPath(tx, pb)
	for _, block := range newPath[1:] {
		// If the diffs for this block have already been generated, apply diffs
		// directly instead of generating them. This is much faster.
		if block.DiffsGenerated {
			err := cs.forwardBlock(tx, block)
			if err != nil {
				return nil, err
			}
		} else {
			err := cs.generateAndApplyDiff(tx, block)
			if err != nil {
				// Mark the block as invalid.
				cs.dosBlocks[block.Block.ID()] = struct{}{}
				return nil, err
			}
		}
		appliedBlocks = append(appliedBlocks, block)

		// Sanity check - after applying a block, check that the consensus set
		// has maintained consistency.
		if build.Release == "testing" {
			cs.checkConsistency(tx)
		} else {
			cs.maybeCheckConsistency(tx)
		}
	}
	return appliedBlocks, nil
}

// rewindBlock rewinds a single block from the consensus set. This method assumes that pb is the current top op the chain, i.e. the active fork
func (cs *ConsensusSet) rewindBlock(tx *bolt.Tx, pb *processedBlock) error {
	cs.log.Debugf("[CS] rewinding block %d\n", pb.Height)
	createDCOBucket(tx, pb.Height)
	commitDiffSet(tx, pb, modules.DiffRevert)
	deleteDCOBucket(tx, pb.Height+cs.chainCts.MaturityDelay)
	return cs.rewindBlockForPlugins(tx, pb)
}

func (cs *ConsensusSet) rewindBlockForPlugins(tx *bolt.Tx, pb *processedBlock) error {
	cBlock := modules.ConsensusBlock{
		Block:                  pb.Block,
		SpentCoinOutputs:       make(map[types.CoinOutputID]types.CoinOutput),
		SpentBlockStakeOutputs: make(map[types.BlockStakeOutputID]types.BlockStakeOutput),
	}
	for _, diff := range pb.CoinOutputDiffs {
		cBlock.SpentCoinOutputs[diff.ID] = diff.CoinOutput
	}
	for _, diff := range pb.DelayedCoinOutputDiffs {
		cBlock.SpentCoinOutputs[diff.ID] = diff.CoinOutput
	}
	for _, diff := range pb.BlockStakeOutputDiffs {
		cBlock.SpentBlockStakeOutputs[diff.ID] = diff.BlockStakeOutput
	}

	for name, plugin := range cs.plugins {
		bucket := cs.bucketForPlugin(tx, name)
		err := plugin.RevertBlock(cBlock, pb.Height, bucket)
		if err != nil {
			return err
		}
	}
	return nil
}

// forwardBlock adds a single block to the chain. It assumes that pb is the block at "currentHeight + 1"
func (cs *ConsensusSet) forwardBlock(tx *bolt.Tx, pb *processedBlock) error {
	cs.log.Debugf("[CS] reapplying block %d\n", pb.Height)
	createDCOBucket(tx, pb.Height+cs.chainCts.MaturityDelay)
	commitDiffSet(tx, pb, modules.DiffApply)
	deleteDCOBucket(tx, pb.Height)
	return cs.forwardBlockForPlugins(tx, pb)
}

func (cs *ConsensusSet) forwardBlockForPlugins(tx *bolt.Tx, pb *processedBlock) error {
	cBlock := modules.ConsensusBlock{
		Block:                  pb.Block,
		SpentCoinOutputs:       make(map[types.CoinOutputID]types.CoinOutput),
		SpentBlockStakeOutputs: make(map[types.BlockStakeOutputID]types.BlockStakeOutput),
	}
	for _, diff := range pb.CoinOutputDiffs {
		cBlock.SpentCoinOutputs[diff.ID] = diff.CoinOutput
	}
	for _, diff := range pb.DelayedCoinOutputDiffs {
		cBlock.SpentCoinOutputs[diff.ID] = diff.CoinOutput
	}
	for _, diff := range pb.BlockStakeOutputDiffs {
		cBlock.SpentBlockStakeOutputs[diff.ID] = diff.BlockStakeOutput
	}

	for name, plugin := range cs.plugins {
		bucket := cs.bucketForPlugin(tx, name)
		err := plugin.ApplyBlock(cBlock, pb.Height, bucket)
		if err != nil {
			return err
		}
	}
	return nil
}

// forkBlockchain will move the consensus set onto the 'newBlock' fork. An
// error will be returned if any of the blocks applied in the transition are
// found to be invalid. forkBlockchain is atomic; the ConsensusSet is only
// updated if the function returns nil.
func (cs *ConsensusSet) forkBlockchain(tx *bolt.Tx, newBlock *processedBlock) (revertedBlocks, appliedBlocks []*processedBlock, err error) {
	commonParent := backtrackToCurrentPath(tx, newBlock)[0]
	revertedBlocks = cs.revertToBlock(tx, commonParent)
	appliedBlocks, err = cs.applyUntilBlock(tx, newBlock)
	if err != nil {
		return nil, nil, err
	}
	return revertedBlocks, appliedBlocks, nil
}
