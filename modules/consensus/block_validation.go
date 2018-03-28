package consensus

import (
	"errors"
	"math/big"

	"github.com/rivine/rivine/crypto"
	"github.com/rivine/rivine/modules"
	"github.com/rivine/rivine/types"
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
func checkTarget(b types.Block, target types.Target, value types.Currency, cs *ConsensusSet) bool {

	stakemodifier := cs.CalculateStakeModifier(cs.Height() + 1)

	// Calculate the hash for the given unspent output and timestamp

	pobshash := crypto.HashAll(stakemodifier.Bytes(), b.POBSOutput.BlockHeight, b.POBSOutput.TransactionIndex, b.POBSOutput.OutputIndex, b.Timestamp)
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
	// Check that the timestamp is not too far in the past to be acceptable.
	if minTimestamp > b.Timestamp {
		return errEarlyTimestamp
	}

	//In what block (transaction) is unspent block stake generated for this POBS
	ubsu := b.POBSOutput

	blockatheight, exist := bv.cs.BlockAtHeight(ubsu.BlockHeight)
	if !exist {
		return errPOBSBlockIndexDoesNotExist
	}
	bsoid := blockatheight.Transactions[ubsu.TransactionIndex].BlockStakeOutputID(ubsu.OutputIndex)
	valueofblockstakeoutput := blockatheight.Transactions[ubsu.TransactionIndex].BlockStakeOutputs[ubsu.OutputIndex].Value

	spent := 0
	//Check that unspent block stake used is spent
	for _, tr := range b.Transactions {
		for _, bsi := range tr.BlockStakeInputs {
			if bsi.ParentID == bsoid {
				spent = 1
			}
		}
	}
	if spent == 0 {
		return errBlockStakeNotRespent
	}

	// Check that the target of the new block is sufficient.
	if !checkTarget(b, target, valueofblockstakeoutput, bv.cs) {
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
	if uint64(len(bv.marshaler.Marshal(b))) > bv.cs.chainCts.BlockSizeLimit {
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
	// Add up the payouts and check that all values are legal.
	var payoutSum types.Currency
	for _, payout := range b.MinerPayouts {
		if payout.Value.IsZero() {
			return false
		}
		payoutSum = payoutSum.Add(payout.Value)
	}
	return b.CalculateSubsidy(bv.cs.chainCts.BlockCreatorFee).Equals(payoutSum)
}
