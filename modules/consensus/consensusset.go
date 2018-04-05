package consensus

// All changes to the consenuss set are made via diffs, specifically by calling
// a commitDiff function. This means that future modifications (such as
// replacing in-memory versions of the utxo set with on-disk versions of the
// utxo set) should be relatively easy to verify for correctness. Modifying the
// commitDiff functions will be sufficient.

import (
	"errors"

	"github.com/rivine/rivine/encoding"
	"github.com/rivine/rivine/modules"
	"github.com/rivine/rivine/persist"
	"github.com/rivine/rivine/sync"
	"github.com/rivine/rivine/types"

	"github.com/NebulousLabs/demotemutex"
	"github.com/rivine/bbolt"
)

var (
	errNilGateway = errors.New("cannot have a nil gateway as input")
)

// marshaler marshals objects into byte slices and unmarshals byte
// slices into objects.
type marshaler interface {
	Marshal(interface{}) []byte
	Unmarshal([]byte, interface{}) error
}
type stdMarshaler struct{}

func (stdMarshaler) Marshal(v interface{}) []byte            { return encoding.Marshal(v) }
func (stdMarshaler) Unmarshal(b []byte, v interface{}) error { return encoding.Unmarshal(b, v) }

// The ConsensusSet is the object responsible for tracking the current status
// of the blockchain. Broadly speaking, it is responsible for maintaining
// consensus.  It accepts blocks and constructs a blockchain, forking when
// necessary.
type ConsensusSet struct {
	// The gateway manages peer connections and keeps the consensus set
	// synchronized to the rest of the network.
	gateway modules.Gateway

	// The block root contains the genesis block.
	blockRoot processedBlock

	// Subscribers to the consensus set will receive a changelog every time
	// there is an update to the consensus set. At initialization, they receive
	// all changes that they are missing.
	//
	// Memory: A consensus set typically has fewer than 10 subscribers, and
	// subscription typically happens entirely at startup. This slice is
	// unlikely to grow beyond 1kb, and cannot by manipulated by an attacker as
	// the function of adding a subscriber should not be exposed.
	subscribers []modules.ConsensusSetSubscriber

	// dosBlocks are blocks that are invalid, but the invalidity is only
	// discoverable during an expensive step of validation. These blocks are
	// recorded to eliminate a DoS vector where an expensive-to-validate block
	// is submitted to the consensus set repeatedly.
	//
	// TODO: dosBlocks needs to be moved into the database, and if there's some
	// reason it can't be in THE database, it should be in a separate database.
	// dosBlocks is an unbounded map that an attacker can manipulate, though
	// iirc manipulations are expensive, to the tune of creating a blockchain
	// PoW per DoS block (though the attacker could conceivably build off of
	// the genesis block, meaning the PoW is not very expensive.
	dosBlocks map[types.BlockID]struct{}

	// checkingConsistency is a bool indicating whether or not a consistency
	// check is in progress. The consistency check logic call itself, resulting
	// in infinite loops. This bool prevents that while still allowing for full
	// granularity consistency checks. Previously, consistency checks were only
	// performed after a full reorg, but now they are performed after every
	// block.
	checkingConsistency bool

	// synced is true if initial blockchain download has finished. It indicates
	// whether the consensus set is synced with the network.
	synced bool

	// Interfaces to abstract the dependencies of the ConsensusSet.
	marshaler       marshaler
	blockRuleHelper blockRuleHelper
	blockValidator  blockValidator

	// Utilities
	db         *persist.BoltDatabase
	log        *persist.Logger
	mu         demotemutex.DemoteMutex
	persistDir string
	tg         sync.ThreadGroup

	bcInfo                 types.BlockchainInfo
	chainCts               types.ChainConstants
	genesisBlockStakeCount types.Currency
}

// New returns a new ConsensusSet, containing at least the genesis block. If
// there is an existing block database present in the persist directory, it
// will be loaded.
func New(gateway modules.Gateway, bootstrap bool, persistDir string, bcInfo types.BlockchainInfo, chainCts types.ChainConstants) (*ConsensusSet, error) {
	// Check for nil dependencies.
	if gateway == nil {
		return nil, errNilGateway
	}

	genesisBlock := chainCts.GenesisBlock()
	// Create the ConsensusSet object.
	cs := &ConsensusSet{
		gateway: gateway,

		blockRoot: processedBlock{
			Block:       genesisBlock,
			ChildTarget: chainCts.RootTarget(),
			Depth:       chainCts.RootDepth,

			DiffsGenerated: true,
		},

		dosBlocks: make(map[types.BlockID]struct{}),

		marshaler:       stdMarshaler{},
		blockRuleHelper: stdBlockRuleHelper{chainCts: chainCts},

		persistDir: persistDir,

		bcInfo:                 bcInfo,
		chainCts:               chainCts,
		genesisBlockStakeCount: chainCts.GenesisBlockStakeCount(),
	}

	cs.blockValidator = newBlockValidator(cs)

	// Create the diffs for the genesis blockstake outputs.
	for i, siafundOutput := range genesisBlock.Transactions[0].BlockStakeOutputs {
		sfid := genesisBlock.Transactions[0].BlockStakeOutputID(uint64(i))
		sfod := modules.BlockStakeOutputDiff{
			Direction:        modules.DiffApply,
			ID:               sfid,
			BlockStakeOutput: siafundOutput,
		}
		cs.blockRoot.BlockStakeOutputDiffs = append(cs.blockRoot.BlockStakeOutputDiffs, sfod)
	}
	// Create the diffs for the genesis coin outputs.
	for i, coinOutput := range genesisBlock.Transactions[0].CoinOutputs {
		sfid := genesisBlock.Transactions[0].CoinOutputID(uint64(i))
		cod := modules.CoinOutputDiff{
			Direction:  modules.DiffApply,
			ID:         sfid,
			CoinOutput: coinOutput,
		}
		cs.blockRoot.CoinOutputDiffs = append(cs.blockRoot.CoinOutputDiffs, cod)
	}

	// Initialize the consensus persistence structures.
	err := cs.initPersist()
	if err != nil {
		return nil, err
	}

	go func() {
		// Sync with the network. Don't sync if we are testing because
		// typically we don't have any mock peers to synchronize with in
		// testing.
		if bootstrap {
			// We are in a virgin goroutine right now, so calling the threaded
			// function without a goroutine is okay.
			err = cs.threadedInitialBlockchainDownload()
			if err != nil {
				return
			}
		}

		// threadedInitialBlockchainDownload will release the threadgroup 'Add'
		// it was holding, so another needs to be grabbed to finish off this
		// goroutine.
		err = cs.tg.Add()
		if err != nil {
			return
		}
		defer cs.tg.Done()

		// Register RPCs
		gateway.RegisterRPC("SendBlocks", cs.rpcSendBlocks)
		gateway.RegisterRPC("RelayHeader", cs.threadedRPCRelayHeader)
		gateway.RegisterRPC("SendBlk", cs.rpcSendBlk)
		gateway.RegisterConnectCall("SendBlocks", cs.threadedReceiveBlocks)
		cs.tg.OnStop(func() {
			cs.gateway.UnregisterRPC("SendBlocks")
			cs.gateway.UnregisterRPC("RelayHeader")
			cs.gateway.UnregisterRPC("SendBlk")
			cs.gateway.UnregisterConnectCall("SendBlocks")
		})

		// Mark that we are synced with the network.
		cs.mu.Lock()
		cs.synced = true
		cs.mu.Unlock()
	}()

	return cs, nil
}

// BlockAtHeight returns the block at a given height.
func (cs *ConsensusSet) BlockAtHeight(height types.BlockHeight) (block types.Block, exists bool) {
	_ = cs.db.View(func(tx *bolt.Tx) error {
		id, err := getPath(tx, height)
		if err != nil {
			return err
		}
		pb, err := getBlockMap(tx, id)
		if err != nil {
			return err
		}
		block = pb.Block
		exists = true
		return nil
	})
	return block, exists
}

// BlockHeightOfBlock returns the blockheight given a block.
func (cs *ConsensusSet) BlockHeightOfBlock(block types.Block) (height types.BlockHeight, exists bool) {
	_ = cs.db.View(func(tx *bolt.Tx) error {
		pb, err := getBlockMap(tx, block.ID())
		if err != nil {
			return err
		}
		height = pb.Height
		exists = true
		return nil
	})
	return height, exists
}

// TransactionAtShortID allows you fetch a transaction from a block within the blockchain,
// using a given shortID.  If that transaction does not exist, false is returned.
func (cs *ConsensusSet) TransactionAtShortID(shortID types.TransactionShortID) (types.Transaction, bool) {
	height := shortID.BlockHeight()
	block, found := cs.BlockAtHeight(height)
	if !found {
		return types.Transaction{}, false
	}

	txSeqID := int(shortID.TransactionSequenceIndex())
	if len(block.Transactions) <= txSeqID {
		return types.Transaction{}, false
	}

	return block.Transactions[txSeqID], true
}

// TransactionAtID allows you to fetch a transaction from a block within the blockchain,
// using a given transaction ID. If that transaction does not exist, false is returned
func (cs *ConsensusSet) TransactionAtID(id types.TransactionID) (types.Transaction, types.TransactionShortID, bool) {
	var txnShortID types.TransactionShortID
	var exists bool
	_ = cs.db.View(func(tx *bolt.Tx) error {
		shortID, err := getTransactionShortID(tx, id)
		if err != nil {
			return err
		}
		txnShortID = shortID
		return nil
	})

	txn, exists := cs.TransactionAtShortID(txnShortID)
	return txn, txnShortID, exists
}

// ChildTarget returns the target for the child of a block.
func (cs *ConsensusSet) ChildTarget(id types.BlockID) (target types.Target, exists bool) {
	// A call to a closed database can cause undefined behavior.
	err := cs.tg.Add()
	if err != nil {
		return types.Target{}, false
	}
	defer cs.tg.Done()

	_ = cs.db.View(func(tx *bolt.Tx) error {
		pb, err := getBlockMap(tx, id)
		if err != nil {
			return err
		}
		target = pb.ChildTarget
		exists = true
		return nil
	})
	return target, exists
}

// Close safely closes the block database.
func (cs *ConsensusSet) Close() error {
	return cs.tg.Stop()
}

// managedCurrentBlock returns the latest block in the heaviest known blockchain.
func (cs *ConsensusSet) managedCurrentBlock() (block types.Block) {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	_ = cs.db.View(func(tx *bolt.Tx) error {
		pb := currentProcessedBlock(tx)
		block = pb.Block
		return nil
	})
	return block
}

// CurrentBlock returns the latest block in the heaviest known blockchain.
func (cs *ConsensusSet) CurrentBlock() (block types.Block) {
	// A call to a closed database can cause undefined behavior.
	err := cs.tg.Add()
	if err != nil {
		return types.Block{}
	}
	defer cs.tg.Done()
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	_ = cs.db.View(func(tx *bolt.Tx) error {
		pb := currentProcessedBlock(tx)
		block = pb.Block
		return nil
	})
	return block
}

// Flush will block until the consensus set has finished all in-progress
// routines.
func (cs *ConsensusSet) Flush() error {
	return cs.tg.Flush()
}

// Height returns the height of the consensus set.
func (cs *ConsensusSet) Height() (height types.BlockHeight) {
	// A call to a closed database can cause undefined behavior.
	err := cs.tg.Add()
	if err != nil {
		return 0
	}
	defer cs.tg.Done()

	_ = cs.db.View(func(tx *bolt.Tx) error {
		height = blockHeight(tx)
		return nil
	})
	return height
}

// InCurrentPath returns true if the block presented is in the current path,
// false otherwise.
func (cs *ConsensusSet) InCurrentPath(id types.BlockID) (inPath bool) {
	// A call to a closed database can cause undefined behavior.
	err := cs.tg.Add()
	if err != nil {
		return false
	}
	defer cs.tg.Done()

	_ = cs.db.View(func(tx *bolt.Tx) error {
		pb, err := getBlockMap(tx, id)
		if err != nil {
			inPath = false
			return nil
		}
		pathID, err := getPath(tx, pb.Height)
		if err != nil {
			inPath = false
			return nil
		}
		inPath = pathID == id
		return nil
	})
	return inPath
}

// MinimumValidChildTimestamp returns the earliest timestamp that the next block
// can have in order for it to be considered valid.
func (cs *ConsensusSet) MinimumValidChildTimestamp(id types.BlockID) (timestamp types.Timestamp, exists bool) {
	// A call to a closed database can cause undefined behavior.
	err := cs.tg.Add()
	if err != nil {
		return 0, false
	}
	defer cs.tg.Done()

	// Error is not checked because it does not matter.
	_ = cs.db.View(func(tx *bolt.Tx) error {
		pb, err := getBlockMap(tx, id)
		if err != nil {
			return err
		}
		timestamp = cs.blockRuleHelper.minimumValidChildTimestamp(tx.Bucket(BlockMap), pb)
		exists = true
		return nil
	})
	return timestamp, exists
}

var (
	_ modules.ConsensusSet = (*ConsensusSet)(nil)
)
