package consensus

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/threefoldtech/rivine/crypto"
	"github.com/threefoldtech/rivine/modules"
	"github.com/threefoldtech/rivine/types"
)

var (
	errBadMinerPayouts            = errors.New("miner payout sum does not equal block subsidy")
	errEarlyTimestamp             = errors.New("block timestamp is too early")
	errExtremeFutureTimestamp     = errors.New("block timestamp too far in future, discarded")
	errFutureTimestamp            = errors.New("block timestamp too far in future, but saved for later use")
	errLargeBlock                 = errors.New("block is too large to be accepted")
	errBlockStakeAgeNotMet        = errors.New("The unspent blockstake (not at index 0 in transaction) is not aged enough")
	errBlockStakeNotRespent       = errors.New("The block stake used to generate block should be respent")
	errPOBSBlockIndexDoesNotExist = errors.New("POBS blockheight index points to unexisting block")
)

// blockValidator validates a Block against a set of block validity rules.
type blockValidator interface {
	// ValidateBlock validates a block against a minimum timestamp, a block
	// target, and a block height.
	ValidateBlock(types.Block, types.Timestamp, types.Target, types.BlockHeight) error
}

// stdBlockValidator is the standard implementation of blockValidator.
type stdBlockValidator struct {
	// clock is a Clock interface that indicates the current system time.
	clock types.Clock

	// marshaler encodes and decodes between objects and byte slices.
	marshaler marshaler
	cs        *ConsensusSet
}

// newBlockValidator creates a new stdBlockValidator with default settings.
func newBlockValidator(consensusSet *ConsensusSet) stdBlockValidator {
	return stdBlockValidator{
		clock:     types.StdClock{},
		marshaler: stdMarshaler{},
		cs:        consensusSet,
	}
}

// checkTarget returns true if the block's ID meets the given target.
func checkTarget(b types.Block, target types.Target, value types.Currency, height types.BlockHeight, cs *ConsensusSet) bool {

	stakemodifier := cs.CalculateStakeModifier(height, b, cs.chainCts.StakeModifierDelay)

	// Calculate the hash for the given unspent output and timestamp

	pobshash, err := crypto.HashAll(stakemodifier.Bytes(), b.POBSOutput.BlockHeight, b.POBSOutput.TransactionIndex, b.POBSOutput.OutputIndex, b.Timestamp)
	if err != nil {
		return false
	}
	// Check if it meets the difficulty
	pobshashvalue := big.NewInt(0).SetBytes(pobshash[:])
	pobshashvalue.Div(pobshashvalue, value.Big()) //TODO rivine : this div can be mul on the other side of the compare

	if pobshashvalue.Cmp(target.Int()) == -1 {
		return true
	}
	return false
}

// ValidateBlock validates a block against a minimum timestamp, a block target,
// and a block height. Returns nil if the block is valid and an appropriate
// error otherwise.
func (bv stdBlockValidator) ValidateBlock(b types.Block, minTimestamp types.Timestamp, target types.Target, height types.BlockHeight) error {
	bv.cs.log.Debugf("[SBV] Validating new block for height %d\n", height)
	// Check that the timestamp is not too far in the past to be acceptable.
	if minTimestamp > b.Timestamp {
		return errEarlyTimestamp
	}

	//In what block (transaction) is unspent block stake generated for this POBS
	ubsu := b.POBSOutput

	// We now need to retrieve the block to check if the blockstake used by the POBS protocol are indeed in the right block and transaction,
	// as indicated by the POBSOutput field in the new block. This field conveniently includes a block height for us. But the problem is that
	// only the "blockmap" bucket has every block available, and those are keyed by their ID, which we don't have. Using the BlockAtHeight
	// cs function returns the block at the given height, but that is the block from the active chain, while the block we are validating
	// could potentially be extending a currently inactive fork. Since the blocks are only keyed by height from the active chain, this convenient
	// approach won't work here. The only way to retrieve the correct block, is to lookup the parent of the validating block, to backtrack in the fork,
	// until the block at the given blockheight is retrieved.
	//
	// But counting back from the current block to the given block is an expensive action, as it requires n database lookup, where n is the amount of blocks
	// since the blockstake was last used. In normal operation of the chain (which we assume), most of the time we should be on the active chain. Which means
	// that we can get the targeted block in 2 lookups in the "blockatheight" helper function. If we have, for example, 100 active block creators, with equal
	// blockstake distribution, then every node should create a block (thus using their blockstake) every 100 blocks. And thuse we would need to always traverse
	// about 100 blocks for the check. So given normal operation, it is much less intensive to first do the check assuming the active chain, and then only do the
	// "correct" check should the previous have failed.

	var valueofblockstakeoutput types.Currency
	spent := false
	blockatheight, exist := bv.cs.BlockAtHeight(ubsu.BlockHeight)
	//Check that unspent block stake used is spent
	if exist {
		// Check bounds
		if ubsu.TransactionIndex < uint64(len(blockatheight.Transactions)) && ubsu.OutputIndex < uint64(len(blockatheight.Transactions[ubsu.TransactionIndex].BlockStakeOutputs)) {
			bsoid := blockatheight.Transactions[ubsu.TransactionIndex].BlockStakeOutputID(ubsu.OutputIndex)
			valueofblockstakeoutput = blockatheight.Transactions[ubsu.TransactionIndex].BlockStakeOutputs[ubsu.OutputIndex].Value
			for _, tr := range b.Transactions {
				for _, bsi := range tr.BlockStakeInputs {
					if bsi.ParentID == bsoid {
						spent = true
					}
				}
			}
		}
	}

	// If the "quick" check in the active fork has failed, try going back from the submitted block in a possible inactive fork
	if !spent {
		blockatheight, _ = bv.cs.FindParentBlock(b, height-ubsu.BlockHeight)
		if ubsu.TransactionIndex < uint64(len(blockatheight.Transactions)) && ubsu.OutputIndex < uint64(len(blockatheight.Transactions[ubsu.TransactionIndex].BlockStakeOutputs)) {
			for _, tr := range b.Transactions {
				for _, bsi := range tr.BlockStakeInputs {
					if blockatheight.Transactions[ubsu.TransactionIndex].BlockStakeOutputID(ubsu.OutputIndex) == bsi.ParentID {
						bv.cs.log.Debugf("[SBV] Confirmed blockstake respend from an inactive fork, ubsu in block %d, new block at height %d\n", ubsu.BlockHeight, height)
						valueofblockstakeoutput = blockatheight.Transactions[ubsu.TransactionIndex].BlockStakeOutputs[ubsu.OutputIndex].Value
						spent = true
					}
				}
			}
		}
	}

	// If we still didn't find a valid respend, it's just not there
	if !spent {
		return errBlockStakeNotRespent
	}

	// Check that the target of the new block is sufficient.
	if !checkTarget(b, target, valueofblockstakeoutput, height, bv.cs) {
		return modules.ErrBlockUnsolved
	}

	// If the index of the unspent block stake output is not the first transaction
	// with the first index, then block stake can only be used to solve blocks
	// after its aging is older than types.BlockStakeAging (more than 1 day)
	if ubsu.TransactionIndex != 0 || ubsu.OutputIndex != 0 {
		BlockStakeAge := blockatheight.Header().Timestamp + types.Timestamp(bv.cs.chainCts.BlockStakeAging)
		if BlockStakeAge > types.Timestamp(b.Header().Timestamp) {
			return errBlockStakeAgeNotMet
		}
	}

	// Check that the block is below the size limit.
	bb, err := bv.marshaler.Marshal(b)
	if err != nil {
		return fmt.Errorf("failed to marshal block: %v", err)
	}
	if uint64(len(bb)) > bv.cs.chainCts.BlockSizeLimit {
		return errLargeBlock
	}

	// Check if the block is in the extreme future. We make a distinction between
	// future and extreme future because there is an assumption that by the time
	// the extreme future arrives, this block will no longer be a part of the
	// longest fork because it will have been ignored by all of the miners.
	if b.Timestamp > bv.clock.Now()+bv.cs.chainCts.ExtremeFutureThreshold {
		return errExtremeFutureTimestamp
	}

	// Verify that the miner payouts are valid.
	if !bv.checkMinerPayouts(b) {
		return errBadMinerPayouts
	}

	// Check if the block is in the near future, but too far to be acceptable.
	// This is the last check because it's an expensive check, and not worth
	// performing if the payouts are incorrect.
	if b.Timestamp > bv.clock.Now()+bv.cs.chainCts.FutureThreshold {
		return errFutureTimestamp
	}
	return nil
}

// checkMinerPayouts checks a block creator payouts to the block's subsidy and
// returns true if they are equal.
func (bv stdBlockValidator) checkMinerPayouts(b types.Block) bool {
	var sumBC, sumTFP types.Currency
	// Add up the payouts and check that all values are legal.
	txFeeUnlockHash := bv.cs.chainCts.TransactionFeeCondition.UnlockHash()
	for _, payout := range b.MinerPayouts {
		if payout.Value.IsZero() {
			return false
		}
		if payout.UnlockHash.Cmp(txFeeUnlockHash) == 0 {
			sumTFP = sumTFP.Add(payout.Value) // payout is for tx fee beneficiary
		} else {
			sumBC = sumBC.Add(payout.Value) // payout is for bc
		}
	}
	// ensure tx fee beneficiary has no payouts, should it not be given
	totalMinerFees := b.CalculateTotalMinerFees()
	if bv.cs.chainCts.TransactionFeeCondition.ConditionType() == types.ConditionTypeNil {
		if !sumTFP.IsZero() {
			return false // no beneficiary is given, so it should have no payouts
		}
	}
	// also take into account any custom miner fee payouts in total miner fees sum
	var (
		err error
		mps []types.MinerPayout
	)
	for _, txn := range b.Transactions {
		mps, err = txn.CustomMinerPayouts()
		if err != nil {
			// ignore here, as the block creator does so as well,
			// but do log as an error
			bv.cs.log.Printf("error occured while fetching custom miner payouts from txn v%v: %v", txn.Version, err)
			continue
		}
		for _, mp := range mps {
			totalMinerFees = totalMinerFees.Add(mp.Value)
		}
	}
	// ensure total sum is correct
	return totalMinerFees.Add(bv.cs.chainCts.BlockCreatorFee).Equals(sumBC.Add(sumTFP))
}
