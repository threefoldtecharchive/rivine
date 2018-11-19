package types

// validtransaction.go has functions for checking whether a transaction is
// valid outside of the context of a consensus set. This means checking the
// size of the transaction, the content of the signatures, and a large set of
// other rules that are inherent to how a transaction should be constructed.

import (
	"errors"

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
)

// MissingCoinOutputError is returned in case a non-existing coin output is spend by a Tx.
type MissingCoinOutputError struct {
	ID CoinOutputID
}

func (err MissingCoinOutputError) Error() string {
	return "transaction spends a nonexisting coin output" + err.ID.String()
}

// MissingBlockStakeOutputError is returned in case a non-existing blockstake output is spend by a Tx.
type MissingBlockStakeOutputError struct {
	ID BlockStakeOutputID
}

func (err MissingBlockStakeOutputError) Error() string {
	return "transaction spends a nonexisting blockstake output" + err.ID.String()
}

// TransactionFitsInABlock checks if the transaction is likely to fit in a block.
// Currently there is no limitation on transaction size other than it must fit
// in a block.
func TransactionFitsInABlock(t Transaction, blockSizeLimit uint64) error {
	// Check that the transaction will fit inside of a block, leaving 5kb for
	// overhead.
	if uint64(len(siabin.Marshal(t))) > blockSizeLimit-5e3 {
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

// DefaultCoinOutputValidation contains the default coin output
// (within the context of a transaction) validation logic.
func DefaultCoinOutputValidation(t Transaction, ctx FundValidationContext, coinInputs map[CoinOutputID]CoinOutput) (err error) {
	var inputSum Currency
	for index, sci := range t.CoinInputs {
		sco, ok := coinInputs[sci.ParentID]
		if !ok {
			return MissingCoinOutputError{ID: sci.ParentID}
		}
		// check if the referenced output's condition has been fulfilled
		err = sco.Condition.Fulfill(sci.Fulfillment, FulfillContext{
			InputIndex:  uint64(index),
			BlockHeight: ctx.BlockHeight,
			BlockTime:   ctx.BlockTime,
			Transaction: t,
		})
		if err != nil {
			return
		}
		inputSum = inputSum.Add(sco.Value)
	}
	if !inputSum.Equals(t.CoinOutputSum()) {
		return ErrCoinInputOutputMismatch
	}
	return nil
}

// DefaultBlockStakeOutputValidation contains the default blockstkae output
// (within the context of a transaction) validation logic.
func DefaultBlockStakeOutputValidation(t Transaction, ctx FundValidationContext, blockStakeInputs map[BlockStakeOutputID]BlockStakeOutput) (err error) {
	var inputSum Currency
	for index, bsi := range t.BlockStakeInputs {
		bso, ok := blockStakeInputs[bsi.ParentID]
		if !ok {
			return MissingBlockStakeOutputError{ID: bsi.ParentID}
		}
		// check if the referenced output's condition has been fulfilled
		err = bso.Condition.Fulfill(bsi.Fulfillment, FulfillContext{
			InputIndex:  uint64(index),
			BlockHeight: ctx.BlockHeight,
			BlockTime:   ctx.BlockTime,
			Transaction: t,
		})
		if err != nil {
			return
		}
		inputSum = inputSum.Add(bso.Value)
	}
	var blockstakeOutputSum Currency
	for _, bso := range t.BlockStakeOutputs {
		blockstakeOutputSum = blockstakeOutputSum.Add(bso.Value)
	}
	if !inputSum.Equals(blockstakeOutputSum) {
		return ErrBlockStakeInputOutputMismatch
	}
	return nil
}
