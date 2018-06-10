package types

// validtransaction.go has functions for checking whether a transaction is
// valid outside of the context of a consensus set. This means checking the
// size of the transaction, the content of the signatures, and a large set of
// other rules that are inherent to how a transaction should be constructed.

import (
	"errors"

	"github.com/rivine/rivine/encoding"
)

var (
	ErrDoubleSpend           = errors.New("transaction uses a parent object twice")
	ErrNonZeroRevision       = errors.New("new file contract has a nonzero revision number")
	ErrTransactionTooLarge   = errors.New("transaction is too large to fit in a block")
	ErrTooSmallMinerFee      = errors.New("transaction has a too small miner fee")
	ErrZeroOutput            = errors.New("transaction cannot have an output or payout that has zero value")
	ErrArbitraryDataTooLarge = errors.New("arbitrary data is too large to fit in a transaction")
)

// TransactionFitsInABlock checks if the transaction is likely to fit in a block.
// Currently there is no limitation on transaction size other than it must fit
// in a block.
func TransactionFitsInABlock(t Transaction, blockSizeLimit uint64) error {
	// Check that the transaction will fit inside of a block, leaving 5kb for
	// overhead.
	if uint64(len(encoding.Marshal(t))) > blockSizeLimit-5e3 {
		return ErrTransactionTooLarge
	}
	return nil
}

// TransactionFollowsMinimumValues checks that all outputs adhere to the rules for the
// minimum allowed values
func TransactionFollowsMinimumValues(t Transaction, minimumMinerFee Currency) error {
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
	return nil
}

// NoRepeatsInTransaction checks that a transaction does not spend multiple outputs twice,
// submit two valid storage proofs for the same file contract, etc. We
// frivolously check that a file contract termination and storage proof don't
// act on the same file contract. There is very little overhead for doing so,
// and the check is only frivolous because of the current rule that file
// contract terminations are not valid after the proof window opens.
func NoRepeatsInTransaction(t Transaction) error {
	// Check that there are no repeat instances of coin outputs, storage
	// proofs, contract terminations, or siafund outputs.
	coinInputs := make(map[CoinOutputID]struct{})
	for _, sci := range t.CoinInputs {
		_, exists := coinInputs[sci.ParentID]
		if exists {
			return ErrDoubleSpend
		}
		coinInputs[sci.ParentID] = struct{}{}
	}
	blockstakeInputs := make(map[BlockStakeOutputID]struct{})
	for _, bsi := range t.BlockStakeInputs {
		_, exists := blockstakeInputs[bsi.ParentID]
		if exists {
			return ErrDoubleSpend
		}
		blockstakeInputs[bsi.ParentID] = struct{}{}
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

// DefaultTransactionValidation contains the default transaction validation logic,
// ensuring that within
func DefaultTransactionValidation(t Transaction, ctx ValidationContext, constants TransactionValidationConstants) (err error) {
	err = TransactionFitsInABlock(t, constants.BlockSizeLimit)
	if err != nil {
		return
	}
	err = ArbitraryDataFits(t.ArbitraryData, constants.ArbitraryDataSizeLimit)
	if err != nil {
		return
	}
	err = NoRepeatsInTransaction(t)
	if err != nil {
		return
	}
	err = TransactionFollowsMinimumValues(t, constants.MinimumMinerFee)
	if err != nil {
		return
	}
	err = ValidateNoDoubleSpendsWithinTransaction(t)
	if err != nil {
		return
	}
	// check if all condtions are standard
	for _, sco := range t.CoinOutputs {
		err = sco.Condition.IsStandardCondition(ctx)
		if err != nil {
			return err
		}
	}
	for _, sfo := range t.BlockStakeOutputs {
		err = sfo.Condition.IsStandardCondition(ctx)
		if err != nil {
			return err
		}
	}
	// check if all fulfillments are standard
	for _, sci := range t.CoinInputs {
		err = sci.Fulfillment.IsStandardFulfillment(ctx)
		if err != nil {
			return err
		}
	}
	for _, sfi := range t.BlockStakeInputs {
		err = sfi.Fulfillment.IsStandardFulfillment(ctx)
		if err != nil {
			return err
		}
	}
	// transaction is valid, according to this local check
	return nil
}
