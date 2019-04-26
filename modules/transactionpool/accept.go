package transactionpool

// TODO: It seems like the transaction pool is not properly detecting conflicts
// between a file contract revision and a file contract.

import (
	"errors"

	"github.com/threefoldtech/rivine/crypto"
	"github.com/threefoldtech/rivine/modules"
	"github.com/threefoldtech/rivine/pkg/encoding/siabin"
	"github.com/threefoldtech/rivine/types"

	"github.com/rivine/bbolt"
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

// relatedObjectIDs determines all of the object ids related to a transaction.
func relatedObjectIDs(ts []types.Transaction) []ObjectID {
	oidMap := make(map[ObjectID]struct{})
	for _, t := range ts {
		for _, sci := range t.CoinInputs {
			oidMap[ObjectID(sci.ParentID)] = struct{}{}
		}
		for i := range t.CoinOutputs {
			oidMap[ObjectID(t.CoinOutputID(uint64(i)))] = struct{}{}
		}
		for _, sfi := range t.BlockStakeInputs {
			oidMap[ObjectID(sfi.ParentID)] = struct{}{}
		}
		for i := range t.BlockStakeOutputs {
			oidMap[ObjectID(t.BlockStakeOutputID(uint64(i)))] = struct{}{}
		}
	}

	var oids []ObjectID
	for oid := range oidMap {
		oids = append(oids, oid)
	}
	return oids
}

// validateTransactionSetComposition checks if the transaction set
// is valid given the state of the pool.
func (tp *TransactionPool) validateTransactionSetComposition(ts []types.Transaction) error {
	// Check that the transaction set is not already known.
	setID := TransactionSetID(crypto.HashObject(ts))
	_, exists := tp.transactionSets[setID]
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
	err := tp.ValidateTransactionSet(ts)
	if err != nil {
		return err
	}
	return nil
}

// handleConflicts detects whether the conflicts in the transaction pool are
// legal children of the new transaction pool set or not.
func (tp *TransactionPool) handleConflicts(ts []types.Transaction, conflicts []TransactionSetID) error {
	// Create a list of all the transaction ids that compose the set of
	// conflicts.
	conflictMap := make(map[types.TransactionID]TransactionSetID)
	for _, conflict := range conflicts {
		conflictSet := tp.transactionSets[conflict]
		for _, conflictTxn := range conflictSet {
			conflictMap[conflictTxn.ID()] = conflict
		}
	}

	// Discard all duplicate transactions from the input transaction set.
	var dedupSet []types.Transaction
	for _, t := range ts {
		_, exists := conflictMap[t.ID()]
		if exists {
			continue
		}
		dedupSet = append(dedupSet, t)
	}
	if len(dedupSet) == 0 {
		return modules.ErrDuplicateTransactionSet
	}
	// If transactions were pruned, it's possible that the set of
	// dependencies/conflicts has also reduced. To minimize computational load
	// on the consensus set, we want to prune out all of the conflicts that are
	// no longer relevant. As an example, consider the transaction set {A}, the
	// set {B}, and the new set {A, C}, where C is dependent on B. {A} and {B}
	// are both conflicts, but after deduplication {A} is no longer a conflict.
	// This is recursive, but it is guaranteed to run only once as the first
	// deduplication is guaranteed to be complete.
	if len(dedupSet) < len(ts) {
		oids := relatedObjectIDs(dedupSet)
		var conflicts []TransactionSetID
		for _, oid := range oids {
			conflict, exists := tp.knownObjects[oid]
			if exists {
				conflicts = append(conflicts, conflict)
			}
		}
		return tp.handleConflicts(dedupSet, conflicts)
	}

	// Merge all of the conflict sets with the input set (input set goes last
	// to preserve dependency ordering), and see if the set as a whole is both
	// small enough to be legal and valid as a set. If no, return an error. If
	// yes, add the new set to the pool, and eliminate the old set. The output
	// diff objects can be repeated, (no need to remove those). Just need to
	// remove the conflicts from tp.transactionSets.
	var superset []types.Transaction
	supersetMap := make(map[TransactionSetID]struct{})
	for _, conflict := range conflictMap {
		supersetMap[conflict] = struct{}{}
	}
	for conflict := range supersetMap {
		superset = append(superset, tp.transactionSets[conflict]...)
	}
	superset = append(superset, dedupSet...)

	// Validates the composition of the transaction set, including fees and
	// IsStandard rules (this is a new set, the rules must be rechecked).
	err := tp.validateTransactionSetComposition(superset)
	if err != nil {
		return err
	}

	// Check that the transaction set is valid.
	cc, err := tp.consensusSet.TryTransactionSet(superset)
	if err != nil {
		return modules.NewConsensusConflict(err.Error())
	}

	// Remove the conflicts from the transaction pool. The diffs do not need to
	// be removed, they will be overwritten later in the function.
	for _, conflict := range conflictMap {
		conflictSet := tp.transactionSets[conflict]
		tp.transactionListSize -= len(siabin.Marshal(conflictSet))
		delete(tp.transactionSets, conflict)
		delete(tp.transactionSetDiffs, conflict)
	}

	// Add the transaction set to the pool.
	setID := TransactionSetID(crypto.HashObject(superset))
	tp.transactionSets[setID] = superset
	for _, diff := range cc.CoinOutputDiffs {
		tp.knownObjects[ObjectID(diff.ID)] = setID
	}
	for _, diff := range cc.BlockStakeOutputDiffs {
		tp.knownObjects[ObjectID(diff.ID)] = setID
	}
	tp.transactionSetDiffs[setID] = cc
	tp.transactionListSize += len(siabin.Marshal(superset))
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

	setID := TransactionSetID(crypto.HashObject(ts))

	// Validate the composition of the transaction set
	tp.log.Debug("validation transaction set %v composition", setID)
	err = tp.validateTransactionSetComposition(ts)
	if err != nil {
		tp.log.Debug("Transaction set %v composition invalid: %v", setID, err)
		return err
	}

	// Check for conflicts with other transactions, which would indicate a
	// double-spend. Legal children of a transaction set will also trigger the
	// conflict-detector.
	oids := relatedObjectIDs(ts)
	var conflicts []TransactionSetID
	for _, oid := range oids {
		conflict, exists := tp.knownObjects[oid]
		if exists {
			conflicts = append(conflicts, conflict)
		}
	}
	if len(conflicts) > 0 {
		tp.log.Debug("Handling conflicts in transaction set %v", setID)
		return tp.handleConflicts(ts, conflicts)
	}
	tp.log.Debug("Trying out transaction set %v against current consensus", setID)
	cc, err := tp.consensusSet.TryTransactionSet(ts)
	if err != nil {
		tp.log.Debug("Transaction set %v has conflict with current consensus", setID)
		return modules.NewConsensusConflict(err.Error())
	}

	// Add the transaction set to the pool.
	tp.transactionSets[setID] = ts
	for _, oid := range oids {
		tp.knownObjects[oid] = setID
	}
	tp.log.Println("Accepted transaction set %v in pool", setID)
	// remember when the transaction was added
	tp.broadcastCache.add(setID, tp.consensusSet.Height())
	tp.transactionSetDiffs[setID] = cc
	tp.transactionListSize += len(siabin.Marshal(ts))
	return nil
}

// AcceptTransaction adds a transaction to the unconfirmed set of
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
	tp.log.Debug("Relaying transaction set %v to peers", TransactionSetID(crypto.HashObject(ts)))
	go tp.gateway.Broadcast("RelayTransactionSet", ts, tp.gateway.Peers())
	tp.updateSubscribersTransactions()
	return nil
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

func (tp *TransactionPool) transactionMinFee() types.Currency {
	return tp.chainCts.MinimumTransactionFee
}
