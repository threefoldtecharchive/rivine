package transactionpool

import (
	"errors"
	"fmt"

	"github.com/NebulousLabs/demotemutex"

	"github.com/threefoldtech/rivine/build"
	"github.com/threefoldtech/rivine/crypto"
	"github.com/threefoldtech/rivine/modules"
	"github.com/threefoldtech/rivine/persist"
	"github.com/threefoldtech/rivine/types"
)

const (
	dbFilename = "transactionpool.db"
)

var (
	dbMetadata = persist.Metadata{
		Header:  "Sia Transaction Pool DB",
		Version: "0.6.0",
	}

	errNilCS      = errors.New("transaction pool cannot initialize with a nil consensus set")
	errNilGateway = errors.New("transaction pool cannot initialize with a nil gateway")
)

type (
	// TransactionSetID is the hash of a transaction set.
	TransactionSetID crypto.Hash

	poolTransactionSet struct {
		ID           TransactionSetID
		Transactions []types.Transaction
	}

	// The TransactionPool tracks incoming transactions, accepting them or
	// rejecting them based on internal criteria such as fees and unconfirmed
	// double spends.
	TransactionPool struct {
		// Dependencies of the transaction pool.
		consensusSet modules.ConsensusSet
		gateway      modules.Gateway

		// transactionSetDiffs map form a transaction set id to the set of
		// diffs that resulted from the transaction set.
		transactionSets       []poolTransactionSet
		transactionSetMapping map[TransactionSetID]int
		transactionSetDiffs   map[TransactionSetID]modules.ConsensusChange
		transactionListSize   int
		// TODO: Write a consistency check comparing transactionSets,
		// transactionSetDiffs.
		//
		// TODO: Write a consistency check making sure that all unconfirmedIDs
		// point to the right place, and that all UnconfirmedIDs are accounted for.

		// The consensus change index tracks how many consensus changes have
		// been sent to the transaction pool. When a new subscriber joins the
		// transaction pool, all prior consensus changes are sent to the new
		// subscriber.
		subscribers []modules.TransactionPoolSubscriber

		// broadcastCache keeps track of all transaction sets currently in the pool.
		broadcastCache transactionCache

		// Utilities.
		db         *persist.BoltDatabase
		mu         demotemutex.DemoteMutex
		persistDir string

		log *persist.Logger

		bcInfo   types.BlockchainInfo
		chainCts types.ChainConstants
	}
)

// New creates a transaction pool that is ready to receive transactions.
func New(cs modules.ConsensusSet, g modules.Gateway, persistDir string, bcInfo types.BlockchainInfo, chainCts types.ChainConstants, verbose bool) (*TransactionPool, error) {
	// Check that the input modules are non-nil.
	if cs == nil {
		return nil, errNilCS
	}
	if g == nil {
		return nil, errNilGateway
	}

	// Initialize a transaction pool.
	tp := &TransactionPool{
		consensusSet: cs,
		gateway:      g,

		transactionSets:       make([]poolTransactionSet, 0),
		transactionSetMapping: make(map[TransactionSetID]int),
		transactionSetDiffs:   make(map[TransactionSetID]modules.ConsensusChange),

		broadcastCache: newTransactionCache(),

		persistDir: persistDir,

		bcInfo:   bcInfo,
		chainCts: chainCts,
	}

	// Open the tpool database.
	err := tp.initPersist(verbose)
	if err != nil {
		return nil, err
	}

	// Register RPCs
	g.RegisterRPC("RelayTransactionSet", tp.relayTransactionSet)

	return tp, nil
}

func (tp *TransactionPool) Close() error {
	tp.gateway.UnregisterRPC("RelayTransactionSet")
	tp.consensusSet.Unsubscribe(tp)

	var errs []error
	if err := tp.db.Close(); err != nil {
		errs = append(errs, fmt.Errorf("db.Close failed: %v", err))
	}
	if err := tp.log.Close(); err != nil {
		errs = append(errs, fmt.Errorf("log.Close failed: %v", err))
	}

	return build.JoinErrors(errs, "; ")
}

// TransactionList returns a list of all transactions in the transaction pool.
// The transactions are provided in an order that can acceptably be put into a
// block.
func (tp *TransactionPool) TransactionList() []types.Transaction {
	tp.mu.RLock()
	defer tp.mu.RUnlock()
	return tp.transactionList()
}
func (tp *TransactionPool) transactionList() []types.Transaction {
	var txns []types.Transaction
	for _, tSet := range tp.transactionSets {
		txns = append(txns, tSet.Transactions...)
	}
	return txns
}

// Transaction implements TransactionPool.Transaction
func (tp *TransactionPool) Transaction(id types.TransactionID) (types.Transaction, error) {
	tp.mu.RLock()
	defer tp.mu.RUnlock()
	for _, tSet := range tp.transactionSets {
		for _, txn := range tSet.Transactions {
			if id == txn.ID() {
				return txn, nil
			}
		}
	}
	return types.Transaction{}, modules.ErrTransactionNotFound
}
