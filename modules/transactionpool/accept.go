package transactionpool

// TODO: It seems like the transaction pool is not properly detecting conflicts
// between a file contract revision and a file contract.

import (
	"errors"
	"fmt"

	"github.com/threefoldtech/rivine/crypto"
	"github.com/threefoldtech/rivine/modules"
	"github.com/threefoldtech/rivine/pkg/encoding/siabin"
	"github.com/threefoldtech/rivine/types"

	bolt "github.com/rivine/bbolt"
)

// TODO: Add a priority structure that will allow the transaction pool to
// fill up beyond the size of a single block, without being subject to
// manipulation.

var (
	errObjectConflict      = errors.New("transaction set conflicts with an existing transaction set")
	errFullTransactionPool = errors.New("transaction pool cannot accept more transactions")
	errLowMinerFees        = errors.New("transaction set needs more miner fees to be accepted")
	errEmptySet            = errors.New("transaction set is empty")
)

// validateTransactionSetComposition checks if the transaction set
// is valid given the state of the pool.
func (tp *TransactionPool) validateTransactionSetComposition(ts []types.Transaction) error {
	// Check that the transaction set is not already known.
	tsH, err := crypto.HashObject(ts)
	if err != nil {
		return err
	}
	setID := TransactionSetID(tsH)
	_, exists := tp.transactionSetByID(setID)
	if exists {
		return modules.ErrDuplicateTransactionSet
	}

	if tp.transactionListSize > tp.chainCts.TransactionPool.PoolSizeLimit {
		return errFullTransactionPool
	}

	// TODO: There is no DoS prevention mechanism in place to prevent repeated
	// expensive verifications of invalid transactions that are created on the
	// fly.

	// Validates that the transaction set fits within the
	// chain (network) defined byte size limit, when binary encoded.
	// It also validates that the the transaction itself does not exceed a transaction pool defined
	// network (chain) constant, again on a byte level. On top of these transaction pool specific rules,
	// it also validates that the transaction is valid according to the consensus,
	// meaning all properties are standard, known and valid. It does however check this within the context
	// that the validation code knows the transaction is still unconfirmed, and thus not yet part of a created block.
	err = tp.ValidateTransactionSetSize(ts)
	if err != nil {
		return err
	}
	return nil
}

// acceptTransactionSet verifies that a transaction set is allowed to be in the
// transaction pool, and then adds it to the transaction pool.
func (tp *TransactionPool) acceptTransactionSet(ts []types.Transaction) error {
	tp.log.Debug("Trying to accept transaction set")
	if len(ts) == 0 {
		tp.log.Debug("Attempted to accept empty transaction set")
		return errEmptySet
	}

	// Remove all transactions that have been confirmed in the transaction set.
	err := tp.db.Update(func(tx *bolt.Tx) error {
		oldTS := ts
		ts = []types.Transaction{}
		for _, txn := range oldTS {
			if !tp.transactionConfirmed(tx, txn.ID()) {
				ts = append(ts, txn)
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	// If no transactions remain, return a duplicate error.
	if len(ts) == 0 {
		tp.log.Debug("Transaction set could not be accepted: all transactions in set are duplicates")
		return modules.ErrDuplicateTransactionSet
	}

	tsh, err := crypto.HashObject(ts)
	if err != nil {
		return err
	}
	setID := TransactionSetID(tsh)

	// Validate the composition of the transaction set
	tp.log.Debug(fmt.Sprintf("validation transaction set %v composition", crypto.Hash(setID).String()))
	err = tp.validateTransactionSetComposition(ts)
	if err != nil {
		tp.log.Debug(fmt.Sprintf("Transaction set %v composition invalid: %v", crypto.Hash(setID).String(), err))
		return err
	}

	// Validate the new set in context of all other sets
	txns := tp.transactionList()
	txns = append(txns, ts...)

	tp.log.Debug(fmt.Sprintf("Trying out transaction set %v against current consensus and txpool state", crypto.Hash(setID).String()))
	cc, err := tp.consensusSet.TryTransactionSet(txns)
	if err != nil {
		tp.log.Debug(fmt.Sprintf("Transaction set %v has conflict with current consensus", crypto.Hash(setID).String()))
		return err
	}

	// Add the transaction set to the pool.
	tp.transactionSetMapping[setID] = len(tp.transactionSets)
	tp.transactionSets = append(tp.transactionSets, poolTransactionSet{
		ID:           setID,
		Transactions: ts,
	})
	tp.log.Println(fmt.Sprintf("Accepted transaction set %v in pool", crypto.Hash(setID).String()))
	// remember when the transaction was added
	tp.broadcastCache.add(setID, tp.consensusSet.Height())
	tp.transactionSetDiffs[setID] = cc
	tsBytes, err := siabin.Marshal(ts)
	if err != nil {
		return fmt.Errorf("failed to (siabin) marshal transaction set: %v", err)
	}
	tp.transactionListSize += len(tsBytes)
	return nil
}

// AcceptTransactionSet adds a transaction to the unconfirmed set of
// transactions. If the transaction is accepted, it will be relayed to
// connected peers.
func (tp *TransactionPool) AcceptTransactionSet(ts []types.Transaction) error {
	tp.mu.Lock()
	defer tp.mu.Unlock()

	err := tp.acceptTransactionSet(ts)
	if err != nil {
		return err
	}

	// Notify subscribers and broadcast the transaction set.
	tsh, _ := crypto.HashObject(ts)
	tp.log.Debug(fmt.Sprintf("Relaying transaction set %v to peers", tsh))
	go tp.gateway.Broadcast("RelayTransactionSet", ts, tp.gateway.Peers())
	return tp.updateSubscribersTransactions()
}

// relayTransactionSet is an RPC that accepts a transaction set from a peer. If
// the accept is successful, the transaction will be relayed to the gateway's
// other peers.
func (tp *TransactionPool) relayTransactionSet(conn modules.PeerConn) error {
	tp.log.Debug("Received transaction set from peer")
	var ts []types.Transaction
	err := siabin.ReadObject(conn, &ts, tp.chainCts.BlockSizeLimit)
	if err != nil {
		return err
	}
	return tp.AcceptTransactionSet(ts)
}

func (tp *TransactionPool) transactionSetByID(id TransactionSetID) (poolTransactionSet, bool) {
	index, ok := tp.transactionSetMapping[id]
	if !ok {
		return poolTransactionSet{}, false
	}
	return tp.transactionSets[index], true
}
