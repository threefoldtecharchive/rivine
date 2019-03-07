package blockcreator

import (
	"errors"
	"fmt"
	"sync"

	"github.com/threefoldtech/rivine/build"
	"github.com/threefoldtech/rivine/modules"
	"github.com/threefoldtech/rivine/persist"
	rivinesync "github.com/threefoldtech/rivine/sync"
	"github.com/threefoldtech/rivine/types"
)

// BlockCreator participates in the Proof Of Block Stake protocol for creating new blocks
type BlockCreator struct {
	// Module dependencies
	cs     modules.ConsensusSet
	tpool  modules.TransactionPool
	wallet modules.Wallet

	bcInfo    types.BlockchainInfo
	chainCts  types.ChainConstants
	genesisID types.BlockID

	// Cache the synced state of the consensus set to avoid unnecessarily locking it
	csSynced bool

	unsolvedBlock *types.Block

	log        *persist.Logger
	mu         sync.RWMutex
	persist    persistence
	persistDir string
	// tg signals the BlockCreator's goroutines to shut down and blocks until all
	// goroutines have exited before returning from Close().
	tg rivinesync.ThreadGroup
}

// startupRescan will rescan the blockchain in the event that the block creator
// persistence layer has become desynchronized from the consensus persistence
// layer. This might happen if a user replaces any of the folders with backups
// or deletes any of the folders.
func (b *BlockCreator) startupRescan() error {
	// Reset all of the variables that have relevance to the consensus set. The
	// operations are wrapped by an anonymous function so that the locking can
	// be handled using a defer statement.
	err := func() error {
		b.mu.Lock()
		defer b.mu.Unlock()

		b.log.Println("Performing a block creator rescan.")
		b.persist.RecentChange = modules.ConsensusChangeBeginning
		b.persist.Height = 0
		b.persist.ParentID = types.BlockID{}
		return b.save()
	}()
	if err != nil {
		return err
	}

	// Subscribe to the consensus set. This is a blocking call that will not
	// return until the block creator has fully caught up to the current block.
	err = b.cs.ConsensusSetSubscribe(b, modules.ConsensusChangeBeginning, b.tg.StopChan())
	if err != nil {
		return err
	}
	b.tg.OnStop(func() {
		b.cs.Unsubscribe(b)
	})
	return nil
}

// New returns a block creator that is collaborating in the pobs protocol.
func New(cs modules.ConsensusSet, tpool modules.TransactionPool, w modules.Wallet, persistDir string, bcInfo types.BlockchainInfo, chainCts types.ChainConstants, verboseLogging bool) (*BlockCreator, error) {
	// Create the block creator and its dependencies.
	if cs == nil {
		return nil, errors.New("A consensset is required to create a block creator")
	}
	if tpool == nil {
		return nil, errors.New("A transaction pool is required to create a block creator")
	}
	if w == nil {
		return nil, errors.New("A wallet is required to create a block creator")
	}

	// Assemble the block creator.
	b := &BlockCreator{
		cs:     cs,
		tpool:  tpool,
		wallet: w,

		bcInfo:    bcInfo,
		chainCts:  chainCts,
		genesisID: chainCts.GenesisBlockID(),

		unsolvedBlock: &types.Block{},

		persistDir: persistDir,
	}

	err := b.initPersist(verboseLogging)
	if err != nil {
		return nil, errors.New("block creator persistence startup failed: " + err.Error())
	}

	b.unsolvedBlock.ParentID = b.persist.ParentID

	err = b.cs.ConsensusSetSubscribe(b, b.persist.RecentChange, b.tg.StopChan())
	if err == modules.ErrInvalidConsensusChangeID {
		// Perform a rescan of the consensus set if the change id is not found.
		// The id will only be not found if there has been desynchronization
		// between the block creator and the consensus package.
		err = b.startupRescan()
		if err != nil {
			return nil, errors.New("block creator startup failed - rescanning failed: " + err.Error())
		}
	} else if err != nil {
		return nil, errors.New("block creator subscription failed: " + err.Error())
	}
	b.tg.OnStop(func() {
		b.cs.Unsubscribe(b)
	})

	b.tpool.TransactionPoolSubscribe(b)
	b.tg.OnStop(func() {
		b.tpool.Unsubscribe(b)
	})

	// Save after synchronizing with consensus
	// In case we exit while the consensus set is syncing, it is possible that one
	// of the callbacks is still modifying the block creator persistent data, which
	// would result in a data race. So grab a readlock here to avoid issues
	b.mu.RLock()
	err = b.save()
	b.mu.RUnlock()
	if err != nil {
		return nil, errors.New("block creator could not save during startup: " + err.Error())
	}

	// Start the proof of block stake protocol
	go b.SolveBlocks()

	return b, nil
}

// Close terminates all ongoing processes involving the block creator, enabling garbage
// collection.
func (b *BlockCreator) Close() error {
	if err := b.tg.Stop(); err != nil {
		return err
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	b.cs.Unsubscribe(b)

	var errs []error
	if err := b.saveSync(); err != nil {
		errs = append(errs, fmt.Errorf("save failed: %v", err))
	}
	if err := b.log.Close(); err != nil {
		errs = append(errs, fmt.Errorf("log.Close failed: %v", err))
	}
	return build.JoinErrors(errs, "; ")
}
