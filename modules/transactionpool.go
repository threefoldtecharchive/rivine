package modules

import (
	"errors"

	"github.com/rivine/rivine/encoding"
	"github.com/rivine/rivine/types"
)

var (
	// ErrDuplicateTransactionSet is the error that gets returned if a
	// duplicate transaction set is given to the transaction pool.
	ErrDuplicateTransactionSet = errors.New("transaction set contains only duplicate transactions")

	// ErrLargeTransaction is the error that gets returned if a transaction
	// provided to the transaction pool is larger than what is allowed by the
	// IsStandard rules.
	ErrLargeTransaction = errors.New("transaction is too large for this transaction pool")

	// ErrLargeTransactionSet is the error that gets returned if a transaction
	// set given to the transaction pool is larger than the limit placed by the
	// IsStandard rules of the transaction pool.
	ErrLargeTransactionSet = errors.New("transaction set is too large for this transaction pool")

	// ErrInvalidArbPrefix is the error that gets returned if a transaction is
	// submitted to the transaction pool which contains a prefix that is not
	// recognized. This helps prevent miners on old versions from mining
	// potentially illegal transactions in the event of a soft-fork.
	ErrInvalidArbPrefix = errors.New("transaction contains non-standard arbitrary data")

	// ErrTransactionNotFound is returned in case no transaction could be found
	// in the transaction pool for a specific ID.
	ErrTransactionNotFound = errors.New("transaction not found")

	// TransactionPoolDir is the name of the directory that is used to store
	// the transaction pool's persistent data.
	TransactionPoolDir = "transactionpool"
)

// A TransactionPoolSubscriber receives updates about the confirmed and
// unconfirmed set from the transaction pool. Generally, there is no need to
// subscribe to both the consensus set and the transaction pool.
type TransactionPoolSubscriber interface {
	// ReceiveTransactionPoolUpdate notifies subscribers of a change to the
	// consensus set and/or unconfirmed set, and includes the consensus change
	// that would result if all of the transactions made it into a block.
	ReceiveUpdatedUnconfirmedTransactions([]types.Transaction, ConsensusChange)
}

// A TransactionPool manages unconfirmed transactions.
type TransactionPool interface {
	// AcceptTransactionSet accepts a set of potentially interdependent
	// transactions.
	AcceptTransactionSet([]types.Transaction) error

	// Close is necessary for clean shutdown (e.g. during testing).
	Close() error

	// FeeEstimation returns an estimation for how high the transaction fee
	// needs to be per byte. The minimum recommended targets getting accepted
	// in ~3 blocks, and the maximum recommended targets getting accepted
	// immediately. Taking the average has a moderate chance of being accepted
	// within one block. The minimum has a strong chance of getting accepted
	// within 10 blocks.
	FeeEstimation() (minimumRecommended, maximumRecommended types.Currency)

	// PurgeTransactionPool is a temporary function available to the miner. In
	// the event that a miner mines an unacceptable block, the transaction pool
	// will be purged to clear out the transaction pool and get rid of the
	// illegal transaction. This should never happen, however there are bugs
	// that make this condition necessary.
	PurgeTransactionPool()

	// TransactionList returns a list of all transactions in the transaction
	// pool. The transactions are provided in an order that can acceptably be
	// put into a block.
	TransactionList() []types.Transaction

	// Transaction returns the transaction with the given ID from the transaction pool.
	// If no transaction for that ID is found ErrNotFound is returned.
	Transaction(id types.TransactionID) (types.Transaction, error)

	// TransactionPoolSubscribe adds a subscriber to the transaction pool.
	// Subscribers will receive all consensus set changes as well as
	// transaction pool changes, and should not subscribe to both.
	TransactionPoolSubscribe(TransactionPoolSubscriber)

	// Unsubscribe removes a subscriber from the transaction pool.
	// This is necessary for clean shutdown of the miner.
	Unsubscribe(TransactionPoolSubscriber)
}

// ConsensusConflict implements the error interface, and indicates that a
// transaction was rejected due to being incompatible with the current
// consensus set, meaning either a double spend or a consensus rule violation -
// it is unlikely that the transaction will ever be valid.
type ConsensusConflict string

// NewConsensusConflict returns a consensus conflict, which implements the
// error interface.
func NewConsensusConflict(s string) ConsensusConflict {
	return ConsensusConflict("consensus conflict: " + s)
}

// Error implements the error interface, turning the consensus conflict into an
// acceptable error type.
func (cc ConsensusConflict) Error() string {
	return string(cc)
}

// CalculateFee returns the fee-per-byte of a transaction set.
func CalculateFee(ts []types.Transaction) types.Currency {
	var sum types.Currency
	for _, t := range ts {
		for _, fee := range t.MinerFees {
			sum = sum.Add(fee)
		}
	}
	size := len(encoding.Marshal(ts))
	return sum.Div64(uint64(size))
}
