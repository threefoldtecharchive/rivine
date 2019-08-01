package consensus

import (
	"errors"
	"fmt"

	bolt "github.com/rivine/bbolt"
	"github.com/threefoldtech/rivine/build"
	"github.com/threefoldtech/rivine/modules"
	"github.com/threefoldtech/rivine/pkg/encoding/siabin"
	"github.com/threefoldtech/rivine/types"
)

var (
	errDiffsNotGenerated   = errors.New("applying diff set before generating errors")
	errInvalidSuccessor    = errors.New("generating diffs for a block that's an invalid successsor to the current block")
	errWrongAppliedDiffSet = errors.New("applying a diff set that isn't the current block")
	errWrongRevertDiffSet  = errors.New("reverting a diff set that isn't the current block")
)

// commitDiffSetSanity performs a series of sanity checks before committing a
// diff set.
func commitDiffSetSanity(tx *bolt.Tx, pb *processedBlock, dir modules.DiffDirection) {
	// This function is purely sanity checks.
	if !build.DEBUG {
		return
	}

	// Diffs should have already been generated for this node.
	if !pb.DiffsGenerated {
		build.Critical(errDiffsNotGenerated)
	}

	// Current node must be the input node's parent if applying, and
	// current node must be the input node if reverting.
	if dir == modules.DiffApply {
		parent, err := getBlockMap(tx, pb.Block.ParentID)
		if err != nil {
			build.Severe(err)
		}
		if parent.Block.ID() != currentBlockID(tx) {
			build.Critical(errWrongAppliedDiffSet)
		}
	} else {
		if pb.Block.ID() != currentBlockID(tx) {
			build.Critical(errWrongRevertDiffSet)
		}
	}
}

// commitCoinOutputDiff applies or reverts a SiacoinOutputDiff.
func commitCoinOutputDiff(tx *bolt.Tx, scod modules.CoinOutputDiff, dir modules.DiffDirection) {
	if scod.Direction == dir {
		addCoinOutput(tx, scod.ID, scod.CoinOutput)
	} else {
		removeCoinOutput(tx, scod.ID)
	}
}

// commitBlockStakeOutputDiff applies or reverts a Siafund output diff.
func commitBlockStakeOutputDiff(tx *bolt.Tx, sfod modules.BlockStakeOutputDiff, dir modules.DiffDirection) {
	if sfod.Direction == dir {
		addBlockStakeOutput(tx, sfod.ID, sfod.BlockStakeOutput)
	} else {
		removeBlockStakeOutput(tx, sfod.ID)
	}
}

// commitTxIDMapDiff applies or reverts a transaction ID mapping diff
func commitTxIDMapDiff(tx *bolt.Tx, tidmod modules.TransactionIDDiff, dir modules.DiffDirection) {
	if tidmod.Direction == dir {
		addTxnIDMapping(tx, tidmod.LongID, tidmod.ShortID)
	} else {
		removeTxnIDMapping(tx, tidmod.LongID)
	}
}

// commitDelayedCoinOutputDiff applies or reverts a delayedCoinOutputDiff.
func commitDelayedCoinOutputDiff(tx *bolt.Tx, dscod modules.DelayedCoinOutputDiff, dir modules.DiffDirection) {
	if dscod.Direction == dir {
		addDCO(tx, dscod.MaturityHeight, dscod.ID, dscod.CoinOutput)
	} else {
		removeDCO(tx, dscod.MaturityHeight, dscod.ID)
	}
}

// commitNodeDiffs commits all of the diffs in a block node.
func commitNodeDiffs(tx *bolt.Tx, pb *processedBlock, dir modules.DiffDirection) {
	if dir == modules.DiffApply {
		for _, scod := range pb.CoinOutputDiffs {
			commitCoinOutputDiff(tx, scod, dir)
		}
		for _, sfod := range pb.BlockStakeOutputDiffs {
			commitBlockStakeOutputDiff(tx, sfod, dir)
		}
		for _, dscod := range pb.DelayedCoinOutputDiffs {
			commitDelayedCoinOutputDiff(tx, dscod, dir)
		}
		for _, txIDd := range pb.TxIDDiffs {
			commitTxIDMapDiff(tx, txIDd, dir)
		}
	} else {
		for i := len(pb.CoinOutputDiffs) - 1; i >= 0; i-- {
			commitCoinOutputDiff(tx, pb.CoinOutputDiffs[i], dir)
		}
		for i := len(pb.BlockStakeOutputDiffs) - 1; i >= 0; i-- {
			commitBlockStakeOutputDiff(tx, pb.BlockStakeOutputDiffs[i], dir)
		}
		for i := len(pb.DelayedCoinOutputDiffs) - 1; i >= 0; i-- {
			commitDelayedCoinOutputDiff(tx, pb.DelayedCoinOutputDiffs[i], dir)
		}
		for i := len(pb.TxIDDiffs) - 1; i >= 0; i-- {
			commitTxIDMapDiff(tx, pb.TxIDDiffs[i], dir)
		}
	}
}

// updateCurrentPath updates the current path after applying a diff set.
func updateCurrentPath(tx *bolt.Tx, pb *processedBlock, dir modules.DiffDirection) {
	// Update the current path.
	if dir == modules.DiffApply {
		pushPath(tx, pb.Block.ID())
	} else {
		popPath(tx)
	}
}

// commitDiffSet applies or reverts the diffs in a blockNode.
func commitDiffSet(tx *bolt.Tx, pb *processedBlock, dir modules.DiffDirection) {
	// Sanity checks - there are a few so they were moved to another function.
	if build.DEBUG {
		commitDiffSetSanity(tx, pb, dir)
	}

	commitNodeDiffs(tx, pb, dir)
	updateCurrentPath(tx, pb, dir)
}

// generateAndApplyDiff will verify the block and then integrate it into the
// consensus state. These two actions must happen at the same time because
// transactions are allowed to depend on each other. We can't be sure that a
// transaction is valid unless we have applied all of the previous transactions
// in the block, which means we need to apply while we verify.
func (cs *ConsensusSet) generateAndApplyDiff(tx *bolt.Tx, pb *processedBlock) error {
	// Sanity check - the block being applied should have the current block as
	// a parent.
	if pb.Block.ParentID != currentBlockID(tx) {
		build.Critical(errInvalidSuccessor)
	}

	// Create the bucket to hold all of the delayed siacoin outputs created by
	// transactions this block. Needs to happen before any transactions are
	// applied.
	createDCOBucket(tx, pb.Height+cs.chainCts.MaturityDelay)

	// Validate and apply each transaction in the block. They cannot be
	// validated all at once because some transactions may not be valid until
	// previous transactions have been applied.
	for idx, txn := range pb.Block.Transactions {
		cTxn := modules.ConsensusTransaction{
			Transaction:            txn,
			SpentCoinOutputs:       make(map[types.CoinOutputID]types.CoinOutput),
			SpentBlockStakeOutputs: make(map[types.BlockStakeOutputID]types.BlockStakeOutput),
		}
		var err error
		for _, ci := range txn.CoinInputs {
			cTxn.SpentCoinOutputs[ci.ParentID], err = getCoinOutput(tx, ci.ParentID)
			if err != nil {
				return fmt.Errorf("failed to find coin input %s as unspent coin output in current consensus state: %v", ci.ParentID.String(), err)
			}
		}
		for _, bsi := range txn.BlockStakeInputs {
			cTxn.SpentBlockStakeOutputs[bsi.ParentID], err = getBlockStakeOutput(tx, bsi.ParentID)
			if err != nil {
				return fmt.Errorf("failed to find block stake input %s as unspent block stake output in current consensus state: %v", bsi.ParentID.String(), err)
			}
		}

		err = cs.validTransaction(tx, cTxn, types.TransactionValidationConstants{
			BlockSizeLimit:         cs.chainCts.BlockSizeLimit,
			ArbitraryDataSizeLimit: cs.chainCts.ArbitraryDataSizeLimit,
			MinimumMinerFee:        cs.chainCts.MinimumTransactionFee,
		}, pb.Height, pb.Block.Timestamp, cs.isBlockCreatingTx(idx, pb.Block))
		if err != nil {
			cs.log.Printf("WARN: block %v cannot be applied: tx %v is invalid: %v",
				pb.Block.ID(), txn.ID(), err)
			return err
		}
		applyTransaction(tx, pb, txn)

		// apply the transaction for each of the plugins
		for name, plugin := range cs.plugins {
			bucket := cs.bucketForPlugin(tx, name)
			err := plugin.ApplyTransaction(cTxn, pb.Height, bucket)
			if err != nil {
				return err
			}
		}
	}

	// After all of the transactions have been applied, 'maintenance' is
	// applied on the block. This includes adding any outputs that have reached
	// maturity, applying any contracts with missed storage proofs, and adding
	// the miner payouts to the list of delayed outputs.
	cs.applyMaintenance(tx, pb)

	// DiffsGenerated are only set to true after the block has been fully
	// validated and integrated. This is required to prevent later blocks from
	// being accepted on top of an invalid block - if the consensus set ever
	// forks over an invalid block, 'DiffsGenerated' will be set to 'false',
	// requiring validation to occur again. when 'DiffsGenerated' is set to
	// true, validation is skipped, therefore the flag should only be set to
	// true on fully validated blocks.
	pb.DiffsGenerated = true

	// Add the block to the current path and block map.
	bid := pb.Block.ID()
	blockMap := tx.Bucket(BlockMap)
	updateCurrentPath(tx, pb, modules.DiffApply)

	// Sanity check preparation - set the consensus hash at this height so that
	// during reverting a check can be performed to assure consistency when
	// adding and removing blocks. Must happen after the block is added to the
	// path.
	if build.DEBUG {
		pb.ConsensusChecksum = consensusChecksum(tx)
	}

	pbb, err := siabin.Marshal(*pb)
	if err != nil {
		return fmt.Errorf("failed to (siabin) marshal processed block: %v", err)
	}
	return blockMap.Put(bid[:], pbb)
}

// isBlockCreatingTx checks if a transaction at a given index in the block is considered to
// be a block creating transaction within the block. A block creating transaction is a
// transaction which respends a single blockstake output for the proof of blockstake protocol.
// The blockstake input and output MUST be the only input and output in the transaction.
// Furthermore, the blockstake input used in the transaction MUST spend the blockstake
// output indicated in the POBSOutput field in the block header
func (cs *ConsensusSet) isBlockCreatingTx(txIdx int, block types.Block) bool {

	// First check if there is only 1 blockstake input and output

	if len(block.Transactions[txIdx].BlockStakeInputs) != 1 {
		return false
	}

	if len(block.Transactions[txIdx].BlockStakeOutputs) != 1 {
		return false
	}

	if len(block.Transactions[txIdx].CoinInputs) != 0 {
		return false
	}

	if len(block.Transactions[txIdx].CoinOutputs) != 0 {
		return false
	}

	// this might be block creation transaction, so do a more expensive check to see
	// if we indeed spend the output indicated in the block header

	// Note that we can't expect the block to extend the active fork, if we do
	// we won't be able to recover forks
	//
	// Also, the block under validation can not be expected to be in the cs yet,
	// so we first load its parent, then the height of the parent, then work
	// from there
	parent, exists := cs.FindParentBlock(block, 1)
	if !exists {
		// can this happen?
		return false
	}
	height, exists := cs.BlockHeightOfBlock(parent)
	if !exists {
		return false
	}
	// height points at our parent block right now, so increment by one to
	// account for that
	height++
	refBlock, exists := cs.FindParentBlock(block, height-block.Header().POBSOutput.BlockHeight)
	if !exists {
		return false
	}
	if block.Header().POBSOutput.TransactionIndex >= uint64(len(refBlock.Transactions)) ||
		block.Header().POBSOutput.OutputIndex >= uint64(len(refBlock.Transactions[block.Header().POBSOutput.TransactionIndex].BlockStakeOutputs)) {
		return false
	}

	t := refBlock.Transactions[block.Header().POBSOutput.TransactionIndex]

	if uint64(len(t.BlockStakeOutputs)) > block.Header().POBSOutput.OutputIndex &&
		t.BlockStakeOutputID(block.Header().POBSOutput.OutputIndex) ==
			block.Transactions[txIdx].BlockStakeInputs[0].ParentID {
		return true
	}

	return false
}
