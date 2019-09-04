package types

// validtransaction.go has functions for checking whether a transaction is
// valid outside of the context of a consensus set. This means checking the
// size of the transaction, the content of the signatures, and a large set of
// other rules that are inherent to how a transaction should be constructed.

import (
	"errors"
	"fmt"

	"github.com/threefoldtech/rivine/pkg/encoding/siabin"
)

// various errors that can be returned as result of a specific transaction validation
var (
	ErrDoubleSpend                   = errors.New("transaction uses a parent object twice")
	ErrNonZeroRevision               = errors.New("new file contract has a nonzero revision number")
	ErrTransactionTooLarge           = errors.New("transaction is too large to fit in a block")
	ErrTooSmallMinerFee              = errors.New("transaction has a too small miner fee")
	ErrZeroOutput                    = errors.New("transaction cannot have an output or payout that has zero value")
	ErrArbitraryDataTooLarge         = errors.New("arbitrary data is too large to fit in a transaction")
	ErrCoinInputOutputMismatch       = errors.New("coin inputs do not equal coin outputs for transaction")
	ErrBlockStakeInputOutputMismatch = errors.New("blockstake inputs do not equal blockstake outputs for transaction")
	ErrMissingMinerFee               = errors.New("transaction does not specify any miner fees")
)

// TransactionFitsInABlock checks if the transaction is likely to fit in a block.
// Currently there is no limitation on transaction size other than it must fit
// in a block.
func TransactionFitsInABlock(t Transaction, blockSizeLimit uint64) error {
	// Check that the transaction will fit inside of a block, leaving 5kb for
	// overhead.
	tb, err := siabin.Marshal(t)
	if err != nil {
		return fmt.Errorf("failed to (siabin) marshal transaction: %v", err)
	}
	if uint64(len(tb)) > blockSizeLimit-5e3 {
		return ErrTransactionTooLarge
	}
	return nil
}

// TransactionFollowsMinimumValues checks that all outputs adhere to the rules for the
// minimum allowed values
func TransactionFollowsMinimumValues(t Transaction, minimumMinerFee Currency, IsBlockCreatingTx bool) error {
	for _, sco := range t.CoinOutputs {
		if sco.Value.IsZero() {
			return ErrZeroOutput
		}
	}
	for _, bso := range t.BlockStakeOutputs {
		if bso.Value.IsZero() {
			return ErrZeroOutput
		}
	}
	for _, fee := range t.MinerFees {
		if fee.Cmp(minimumMinerFee) == -1 {
			return ErrTooSmallMinerFee
		}
	}
	// also reject abscent miner fees is the minimum fee is non zero
	// block creation transactions are exempt from this rule
	if !IsBlockCreatingTx && len(t.MinerFees) == 0 && !minimumMinerFee.IsZero() {
		return ErrMissingMinerFee
	}
	return nil
}

// ArbitraryDataFits checks if an arbtirary data first within a given size limit.
func ArbitraryDataFits(arbitraryData []byte, sizeLimit uint64) error {
	if uint64(len(arbitraryData)) > sizeLimit {
		return ErrArbitraryDataTooLarge
	}
	return nil
}

// ValidateNoDoubleSpendsWithinTransaction validates that no output has been spend twice,
// within the given transaction. NOTE that this is a local test only,
// and does not guarantee that an output isn't already spend in another transaction.
func ValidateNoDoubleSpendsWithinTransaction(t Transaction) (err error) {
	spendCoins := make(map[CoinOutputID]struct{})
	for _, ci := range t.CoinInputs {
		if _, found := spendCoins[ci.ParentID]; found {
			err = ErrDoubleSpend
			return
		}
		spendCoins[ci.ParentID] = struct{}{}
	}

	spendBlockStakes := make(map[BlockStakeOutputID]struct{})
	for _, bsi := range t.BlockStakeInputs {
		if _, found := spendBlockStakes[bsi.ParentID]; found {
			err = ErrDoubleSpend
			return
		}
		spendBlockStakes[bsi.ParentID] = struct{}{}
	}

	return
}
