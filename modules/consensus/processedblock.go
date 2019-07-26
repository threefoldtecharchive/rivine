package consensus

import (
	"fmt"
	"math/big"

	bolt "github.com/rivine/bbolt"
	"github.com/threefoldtech/rivine/build"
	"github.com/threefoldtech/rivine/crypto"
	"github.com/threefoldtech/rivine/modules"
	"github.com/threefoldtech/rivine/pkg/encoding/siabin"
	"github.com/threefoldtech/rivine/types"
)

// SurpassThreshold is a percentage that dictates how much heavier a competing
// chain has to be before the node will switch to mining on that chain. This is
// not a consensus rule. This percentage is only applied to the most recent
// block, not the entire chain; see blockNode.heavierThan.
//
// If no threshold were in place, it would be possible to manipulate a block's
// timestamp to produce a sufficiently heavier block.
var SurpassThreshold = big.NewRat(20, 100)

// processedBlock is a copy/rename of blockNode, with the pointers to
// other blockNodes replaced with block ID's, and all the fields
// exported, so that a block node can be marshalled
type processedBlock struct {
	Block       types.Block
	Height      types.BlockHeight
	Depth       types.Target
	ChildTarget types.Target

	DiffsGenerated         bool
	CoinOutputDiffs        []modules.CoinOutputDiff
	BlockStakeOutputDiffs  []modules.BlockStakeOutputDiff
	DelayedCoinOutputDiffs []modules.DelayedCoinOutputDiff
	TxIDDiffs              []modules.TransactionIDDiff

	ConsensusChecksum crypto.Hash
}

// heavierThan returns true if the blockNode is sufficiently heavier than
// 'cmp'. 'cmp' is expected to be the current block node. "Sufficient" means
// that the weight of 'bn' exceeds the weight of 'cmp' by:
//		(the target of 'cmp' * 'Surpass Threshold')
func (pb *processedBlock) heavierThan(cmp *processedBlock, rootDepth types.Target) bool {
	requirement := cmp.Depth.AddDifficulties(cmp.ChildTarget.MulDifficulty(SurpassThreshold, rootDepth), rootDepth)
	return requirement.Cmp(pb.Depth) > 0 // Inversed, because the smaller target is actually heavier.
}

// childDepth returns the depth of a blockNode's child nodes. The depth is the
// "sum" of the current depth and current difficulty. See target.Add for more
// detailed information.
func (pb *processedBlock) childDepth(rootDepth types.Target) types.Target {
	return pb.Depth.AddDifficulties(pb.ChildTarget, rootDepth)
}

// targetAdjustmentBase returns the magnitude that the target should be
// adjusted by before a clamp is applied.
func (cs *ConsensusSet) targetAdjustmentBase(blockMap *bolt.Bucket, pb *processedBlock) *big.Rat {
	// Grab the block that was generated 'TargetWindow' blocks prior to the
	// parent. If there are not 'TargetWindow' blocks yet, stop at the genesis
	// block.
	var windowSize types.BlockHeight
	parent := pb.Block.ParentID
	current := pb.Block.ID()
	for windowSize = 0; windowSize < cs.chainCts.TargetWindow && parent != (types.BlockID{}); windowSize++ {
		current = parent
		parentID, _ := pb.Block.UnmarshalBlockHeadersParentIDAndTS(blockMap.Get(parent[:]))
		parent = parentID
	}
	_, timestamp := pb.Block.UnmarshalBlockHeadersParentIDAndTS(blockMap.Get(current[:]))

	// The target of a child is determined by the amount of time that has
	// passed between the generation of its immediate parent and its
	// TargetWindow'th parent. The expected amount of seconds to have passed is
	// TargetWindow*BlockFrequency. The target is adjusted in proportion to how
	// time has passed vs. the expected amount of time to have passed.
	//
	// The target is converted to a big.Rat to provide infinite precision
	// during the calculation. The big.Rat is just the int representation of a
	// target.
	timePassed := pb.Block.Timestamp - timestamp
	expectedTimePassed := cs.chainCts.BlockFrequency * windowSize
	return big.NewRat(int64(timePassed), int64(expectedTimePassed))
}

// clampTargetAdjustment returns a clamped version of the base adjustment
// value. The clamp keeps the maximum adjustment to ~7x every 2000 blocks. This
// ensures that raising and lowering the difficulty requires a minimum amount
// of total work, which prevents certain classes of difficulty adjusting
// attacks.
func (cs *ConsensusSet) clampTargetAdjustment(base *big.Rat) *big.Rat {
	if base.Cmp(cs.chainCts.MaxAdjustmentUp) > 0 {
		return cs.chainCts.MaxAdjustmentUp
	} else if base.Cmp(cs.chainCts.MaxAdjustmentDown) < 0 {
		return cs.chainCts.MaxAdjustmentDown
	}
	return base
}

// setChildTarget computes the target of a blockNode's child. All children of a node
// have the same target.
func (cs *ConsensusSet) setChildTarget(blockMap *bolt.Bucket, pb *processedBlock) {
	// Fetch the parent block.
	var parent processedBlock
	parentBytes := blockMap.Get(pb.Block.ParentID[:])
	err := siabin.Unmarshal(parentBytes, &parent)
	if err != nil {
		build.Severe(err)
	}

	if pb.Height%(cs.chainCts.TargetWindow/2) != 0 {
		pb.ChildTarget = parent.ChildTarget
		return
	}
	adjustment := cs.clampTargetAdjustment(cs.targetAdjustmentBase(blockMap, pb))
	adjustedRatTarget := new(big.Rat).Mul(parent.ChildTarget.Rat(), adjustment)
	pb.ChildTarget = types.RatToTarget(
		adjustedRatTarget, cs.chainCts.RootDepth)
}

// newChild creates a blockNode from a block and adds it to the parent's set of
// children. The new node is also returned. It necessarily modifies the database
func (cs *ConsensusSet) newChild(tx *bolt.Tx, pb *processedBlock, b types.Block) *processedBlock {
	// Create the child node.
	childID := b.ID()
	child := &processedBlock{
		Block:  b,
		Height: pb.Height + 1,
		Depth:  pb.childDepth(cs.chainCts.RootDepth),
	}
	blockMap := tx.Bucket(BlockMap)
	cs.setChildTarget(blockMap, child)
	childBytes, err := siabin.Marshal(*child)
	if err != nil {
		build.Severe(fmt.Errorf("failed to (siabin) marshal child processed block: %v", err))
	}
	err = blockMap.Put(childID[:], childBytes)
	if err != nil {
		build.Severe(err)
	}
	return child
}
