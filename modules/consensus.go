package modules

import (
	"errors"
	"math/big"

	"github.com/rivine/rivine/crypto"
	"github.com/rivine/rivine/types"
)

const (
	// ConsensusDir is the name of the directory used for all of the consensus
	// persistence files.
	ConsensusDir = "consensus"

	// DiffApply indicates that a diff is being applied to the consensus set.
	DiffApply DiffDirection = true

	// DiffRevert indicates that a diff is being reverted from the consensus
	// set.
	DiffRevert DiffDirection = false
)

var (
	// ConsensusChangeBeginning is a special consensus change id that tells the
	// consensus set to provide all consensus changes starting from the very
	// first diff, which includes the genesis block diff.
	ConsensusChangeBeginning = ConsensusChangeID{}

	// ConsensusChangeRecent is a special consensus change id that tells the
	// consensus set to provide the most recent consensus change, instead of
	// starting from a specific value (which may not be known to the caller).
	ConsensusChangeRecent = ConsensusChangeID{1}

	// ErrBlockKnown is an error indicating that a block is already in the
	// database.
	ErrBlockKnown = errors.New("block already present in database")

	// ErrBlockUnsolved indicates that a block did not meet the required POS
	// target.
	ErrBlockUnsolved = errors.New("block does not meet target")

	// ErrInvalidConsensusChangeID indicates that ConsensusSetPersistSubscribe
	// was called with a consensus change id that is not recognized. Most
	// commonly, this means that the consensus set was deleted or replaced and
	// now the module attempting the subscription has desynchronized. This error
	// should be handled by the module, and not reported to the user.
	ErrInvalidConsensusChangeID = errors.New("consensus subscription has invalid id - files are inconsistent")

	// ErrNonExtendingBlock indicates that a block is valid but does not result
	// in a fork that is the heaviest known fork - the consensus set has not
	// changed as a result of seeing the block.
	ErrNonExtendingBlock = errors.New("block does not extend the longest fork")
)

type (
	// ConsensusChangeID is the id of a consensus change.
	ConsensusChangeID crypto.Hash

	// A DiffDirection indicates the "direction" of a diff, either applied or
	// reverted. A bool is used to restrict the value to these two possibilities.
	DiffDirection bool

	// A ConsensusSetSubscriber is an object that receives updates to the consensus
	// set every time there is a change in consensus.
	ConsensusSetSubscriber interface {
		// ProcessConsensusChange sends a consensus update to a module through
		// a function call. Updates will always be sent in the correct order.
		// There may not be any reverted blocks, but there will always be
		// applied blocks.
		ProcessConsensusChange(ConsensusChange)
	}

	// A ConsensusChange enumerates a set of changes that occurred to the consensus set.
	ConsensusChange struct {
		// ID is a unique id for the consensus change derived from the reverted
		// and applied blocks.
		ID ConsensusChangeID

		// RevertedBlocks is the list of blocks that were reverted by the change.
		// The reverted blocks were always all reverted before the applied blocks
		// were applied. The revered blocks are presented in the order that they
		// were reverted.
		RevertedBlocks []types.Block

		// AppliedBlocks is the list of blocks that were applied by the change. The
		// applied blocks are always all applied after all the reverted blocks were
		// reverted. The applied blocks are presented in the order that they were
		// applied.
		AppliedBlocks []types.Block

		// CoinOutputDiffs contains the set of coin diffs that were applied
		// to the consensus set in the recent change. The direction for the set of
		// diffs is 'DiffApply'.
		CoinOutputDiffs []CoinOutputDiff

		// BlockStakeOutputDiffs contains the set of blockstake diffs that were applied
		// to the consensus set in the recent change. The direction for the set of
		// diffs is 'DiffApply'.
		BlockStakeOutputDiffs []BlockStakeOutputDiff

		// ChildTarget defines the target of any block that would be the child
		// of the block most recently appended to the consensus set.
		ChildTarget types.Target

		// MinimumValidChildTimestamp defines the minimum allowed timestamp for
		// any block that is the child of the block most recently appended to
		// the consensus set.
		MinimumValidChildTimestamp types.Timestamp

		// Synced indicates whether or not the ConsensusSet is synced with its
		// peers.
		Synced bool
	}

	// A CoinOutputDiff indicates the addition or removal of a CoinOutput in
	// the consensus set.
	CoinOutputDiff struct {
		Direction  DiffDirection
		ID         types.CoinOutputID
		CoinOutput types.CoinOutput
	}

	// A BlockStakeOutputDiff indicates the addition or removal of a BlockStakeOutput in
	// the consensus set.
	BlockStakeOutputDiff struct {
		Direction        DiffDirection
		ID               types.BlockStakeOutputID
		BlockStakeOutput types.BlockStakeOutput
	}

	// A DelayedCoinOutputDiff indicates the introduction of a coin output
	// that cannot be spent until after maturing for 144 blocks. When the output
	// has matured, a CoinOutputDiff will be provided.
	DelayedCoinOutputDiff struct {
		Direction      DiffDirection
		ID             types.CoinOutputID
		CoinOutput     types.CoinOutput
		MaturityHeight types.BlockHeight
	}

	// TransactionIDDiff represents the addition or removal of a mapping between a transactions
	// long ID ands its short ID
	TransactionIDDiff struct {
		Direction DiffDirection
		LongID    types.TransactionID
		ShortID   types.TransactionShortID
	}

	// A ConsensusSet accepts blocks and builds an understanding of network
	// consensus.
	ConsensusSet interface {
		// AcceptBlock adds a block to consensus. An error will be returned if the
		// block is invalid, has been seen before, is an orphan, or doesn't
		// contribute to the heaviest fork known to the consensus set. If the block
		// does not become the head of the heaviest known fork but is otherwise
		// valid, it will be remembered by the consensus set but an error will
		// still be returned.
		AcceptBlock(types.Block) error

		// BlockAtHeight returns the block found at the input height, with a
		// bool to indicate whether that block exists.
		BlockAtHeight(types.BlockHeight) (types.Block, bool)

		// BlockHeightOfBlock returns the blockheight of a given block, with a
		// bool to indicate whether that block exists.
		BlockHeightOfBlock(types.Block) (types.BlockHeight, bool)

		// TransactionAtShortID allows you fetch a transaction from a block within
		// the blockchain, using a given shortID.
		// If that transaction does not exist, false is returned.
		TransactionAtShortID(shortID types.TransactionShortID) (types.Transaction, bool)

		// TransactionAtID allows you to fetch a transaction from a block within
		// the blockchain, using a given transaction ID. If that transaction
		// does not exist, false is returned
		TransactionAtID(types.TransactionID) (types.Transaction, types.TransactionShortID, bool)

		// FindParentBlock finds the parent of a block at the given depth. It guarantees that
		// the correct parent block is found, even if the block is not on the longest fork.
		FindParentBlock(b types.Block, depth types.BlockHeight) (block types.Block, exists bool)

		// ChildTarget returns the target required to extend the current heaviest
		// fork. This function is typically used by miners looking to extend the
		// heaviest fork.
		ChildTarget(types.BlockID) (types.Target, bool)

		// Close will shut down the consensus set, giving the module enough time to
		// run any required closing routines.
		Close() error

		// ConsensusSetSubscribe adds a subscriber to the list of subscribers
		// and gives them every consensus change that has occurred since the
		// change with the provided id. There are a few special cases,
		// described by the ConsensusChangeX variables in this package.
		ConsensusSetSubscribe(ConsensusSetSubscriber, ConsensusChangeID) error

		// CurrentBlock returns the latest block in the heaviest known
		// blockchain.
		CurrentBlock() types.Block

		// Flush will cause the consensus set to finish all in-progress
		// routines.
		Flush() error

		// Height returns the current height of consensus.
		Height() types.BlockHeight

		// Synced returns true if the consensus set is synced with the network.
		Synced() bool

		// InCurrentPath returns true if the block id presented is found in the
		// current path, false otherwise.
		InCurrentPath(types.BlockID) bool

		// MinimumValidChildTimestamp returns the earliest timestamp that is
		// valid on the current longest fork according to the consensus set. This is
		// a required piece of information for the miner, who could otherwise be at
		// risk of mining invalid blocks.
		MinimumValidChildTimestamp(types.BlockID) (types.Timestamp, bool)

		// CalculateStakeModifier calculates the stakemodifier from the blockchain.
		CalculateStakeModifier(height types.BlockHeight, block types.Block, delay types.BlockHeight) *big.Int

		// TryTransactionSet checks whether the transaction set would be valid if
		// it were added in the next block. A consensus change is returned
		// detailing the diffs that would result from the application of the
		// transaction.
		TryTransactionSet([]types.Transaction) (ConsensusChange, error)

		// Unsubscribe removes a subscriber from the list of subscribers,
		// allowing for garbage collection and rescanning. If the subscriber is
		// not found in the subscriber database, no action is taken.
		Unsubscribe(ConsensusSetSubscriber)
	}
)

// Append takes to ConsensusChange objects and adds all of their diffs together.
//
// NOTE: It is possible for diffs to overlap or be inconsistent. This function
// should only be used with consecutive or disjoint consensus change objects.
func (cc ConsensusChange) Append(cc2 ConsensusChange) ConsensusChange {
	return ConsensusChange{
		RevertedBlocks:        append(cc.RevertedBlocks, cc2.RevertedBlocks...),
		AppliedBlocks:         append(cc.AppliedBlocks, cc2.AppliedBlocks...),
		CoinOutputDiffs:       append(cc.CoinOutputDiffs, cc2.CoinOutputDiffs...),
		BlockStakeOutputDiffs: append(cc.BlockStakeOutputDiffs, cc2.BlockStakeOutputDiffs...),
	}
}
