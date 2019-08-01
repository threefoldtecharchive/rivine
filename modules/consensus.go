package modules

import (
	"context"
	"errors"
	"math/big"

	"github.com/threefoldtech/rivine/crypto"
	"github.com/threefoldtech/rivine/persist"
	"github.com/threefoldtech/rivine/types"

	bolt "github.com/rivine/bbolt"
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

	// A ConsensusSetPlugin is an object that receives updates to the consensus set
	// every time there is a change in consensus. The difference with a ConsensusSetSubscriber
	// is that it stores its data within the database of the consensus set.
	ConsensusSetPlugin interface {
		// Initialize the bucket, could be creating it, migrating it,
		// or simply checking that all is as expected.
		// An error should be returned in case something went wrong.
		// metadata is nil in case the plugin wasn't registered prior to this attempt.
		// This method will be called while registering the plugin.
		InitPlugin(metadata *persist.Metadata, bucket *bolt.Bucket, ps PluginViewStorage, cb PluginUnregisterCallback) (persist.Metadata, error)

		// Apply the block to the plugin.
		// An error should be returned in case something went wrong.
		ApplyBlock(block ConsensusBlock, height types.BlockHeight, bucket *persist.LazyBoltBucket) error
		// Revert the block from the plugin.
		// An error should be returned in case something went wrong.
		RevertBlock(block ConsensusBlock, height types.BlockHeight, bucket *persist.LazyBoltBucket) error

		// Apply the transaction to the plugin.
		// An error should be returned in case something went wrong.
		ApplyTransaction(txn ConsensusTransaction, height types.BlockHeight, bucket *persist.LazyBoltBucket) error
		// Revert the transaction from the plugin.
		// An error should be returned in case something went wrong.
		RevertTransaction(txn ConsensusTransaction, height types.BlockHeight, bucket *persist.LazyBoltBucket) error

		// TransactionValidatorFunctions allows the plugin to provide validation rules for all transaction versions it mapped to
		TransactionValidatorVersionFunctionMapping() map[types.TransactionVersion][]PluginTransactionValidationFunction

		// TransactionValidators allows the plugin to provide validation rules for all transactions versions it wants
		TransactionValidators() []PluginTransactionValidationFunction

		// Close releases any resources helt by the plugin like the PluginViewStorage
		Close() error
	}

	// PluginUnregisterCallback allows plugins to unregister
	PluginUnregisterCallback func(plugin ConsensusSetPlugin)

	// A PluginStorage
	PluginViewStorage interface {
		View(callback func(bucket *bolt.Bucket) error) error
		Close() error
	}

	// PluginTransactionValidationFunction is the signature of a validator function that
	// can be used to provide plugin-driven transaction validation, provided by (and linked to) a plugin.
	PluginTransactionValidationFunction func(tx ConsensusTransaction, ctx types.TransactionValidationContext, bucket *persist.LazyBoltBucket) error

	// TransactionValidationFunction is the signature of a validator function that
	// can be used to provide validation rules for transactions.
	TransactionValidationFunction func(tx ConsensusTransaction, ctx types.TransactionValidationContext) error

	// ConsensusBlock is the block type as exposed by the consensus module,
	// allowing you to easily find the spend coin and blockstake outputs,
	// for any of the by-the-block defined inputs
	ConsensusBlock struct {
		types.Block

		SpentCoinOutputs       map[types.CoinOutputID]types.CoinOutput
		SpentBlockStakeOutputs map[types.BlockStakeOutputID]types.BlockStakeOutput
	}

	// ConsensusTransaction is the transaction type as exposed by the consensus module
	// allowing you to easily find the spend coin and blockstake outputs,
	// for any of the by-the-transaction defined inputs
	ConsensusTransaction struct {
		types.Transaction

		SpentCoinOutputs       map[types.CoinOutputID]types.CoinOutput
		SpentBlockStakeOutputs map[types.BlockStakeOutputID]types.BlockStakeOutput
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
		// Start the consensusset
		// this function starts a new goroutine and then returns immediately
		Start()

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
		ConsensusSetSubscribe(ConsensusSetSubscriber, ConsensusChangeID, <-chan struct{}) error

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

		// GetCoinOutput takes a coin output ID and returns the appropriate coin output
		GetCoinOutput(types.CoinOutputID) (types.CoinOutput, error)

		// GetBlockStakeOutput takes a blockstake output ID and returns the appropriate blockstake output
		GetBlockStakeOutput(types.BlockStakeOutputID) (types.BlockStakeOutput, error)

		// RegisterPlugin takes in a name and plugin and registers this plugin on the consensus
		// When the plugin is registered, all unprocessed changes are synchronously sent to the plugin
		// unless the passed context is cancelled
		RegisterPlugin(ctx context.Context, name string, plugin ConsensusSetPlugin) error

		// UnregisterPlugin takes in a name and plugin and unregisters this plugin off the consensus
		UnregisterPlugin(name string, plugin ConsensusSetPlugin)
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
