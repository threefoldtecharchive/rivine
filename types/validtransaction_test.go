package types

import (
	"math"
	"testing"
)

// TestTransactionFitsInABlock probes the fitsInABlock method of the
// Transaction type.
func TestTransactionFitsInABlock(t *testing.T) {
	blockSizeLimit := DefaultChainConstants().BlockSizeLimit
	data := make([]byte, blockSizeLimit/2)
	txn := Transaction{ArbitraryData: data}
	err := txn.fitsInABlock(blockSizeLimit)
	if err != nil {
		t.Error(err)
	}
	data = make([]byte, blockSizeLimit)
	txn.ArbitraryData = data
	err = txn.fitsInABlock(blockSizeLimit)
	if err != ErrTransactionTooLarge {
		t.Error(err)
	}
}

// TestTransactionFollowsMinimumValues probes the followsMinimumValues method
// of the Transaction type.
func TestTransactionFollowsMinimumValues(t *testing.T) {
	// Start with a transaction that follows all of minimum-values rules.
	txn := Transaction{
		CoinOutputs:       []CoinOutput{{Value: NewCurrency64(1)}},
		BlockStakeOutputs: []BlockStakeOutput{{Value: NewCurrency64(1)}},
		MinerFees:         []Currency{NewCurrency64(1)},
	}
	err := txn.followsMinimumValues()
	if err != nil {
		t.Error(err)
	}

	// Try a zero value for each type.
	txn.CoinOutputs[0].Value = ZeroCurrency
	err = txn.followsMinimumValues()
	if err != ErrZeroOutput {
		t.Error(err)
	}
	txn.CoinOutputs[0].Value = NewCurrency64(1)
	txn.BlockStakeOutputs[0].Value = ZeroCurrency
	err = txn.followsMinimumValues()
	if err != ErrZeroOutput {
		t.Error(err)
	}
	txn.BlockStakeOutputs[0].Value = NewCurrency64(1)
	txn.MinerFees[0] = ZeroCurrency
	err = txn.followsMinimumValues()
	if err != ErrZeroMinerFee {
		t.Error(err)
	}
	txn.MinerFees[0] = NewCurrency64(1)

}

// TestTransactionNoRepeats probes the noRepeats method of the Transaction
// type.
func TestTransactionNoRepeats(t *testing.T) {
	// Try a transaction all the repeatable types but no conflicts.
	txn := Transaction{
		CoinInputs:       []CoinInput{{}},
		BlockStakeInputs: []BlockStakeInput{{}},
	}

	// Try a transaction double spending a siacoin output.
	txn.CoinInputs = append(txn.CoinInputs, CoinInput{})
	err := txn.noRepeats()
	if err != ErrDoubleSpend {
		t.Error(err)
	}
	txn.CoinInputs = txn.CoinInputs[:1]

	// Try a transaction double spending a siafund output.
	txn.BlockStakeInputs = append(txn.BlockStakeInputs, BlockStakeInput{})
	err = txn.noRepeats()
	if err != ErrDoubleSpend {
		t.Error(err)
	}
	txn.BlockStakeInputs = txn.BlockStakeInputs[:1]
}

func TestUnknownTransactionValidation(t *testing.T) {
	cts := DefaultChainConstants()

	// Build a working unknown transaction.
	txn := Transaction{Extension: unknownTransactionExtension{}}

	// validation of unknown transactions should always succeed,
	// as no validation is applied here
	err := txn.ValidateTransaction(TransactionValidationContext{
		CurrentBlockHeight: 0,
		BlockSizeLimit:     cts.BlockSizeLimit,
	})
	if err != nil {
		t.Errorf("expected no error, but received: %v", err)
	}
	err = txn.ValidateTransaction(TransactionValidationContext{
		CurrentBlockHeight: math.MaxUint64,
		BlockSizeLimit:     0,
	})
	if err != nil {
		t.Errorf("expected no error, but received: %v", err)
	}
}

// TestLegacyTransactionValidation probes the validation logic of the
// Transaction type.
func TestLegacyTransactionValidation(t *testing.T) {
	cts := DefaultChainConstants()

	// Build a working transaction.
	var txn Transaction
	err := txn.ValidateTransaction(TransactionValidationContext{
		CurrentBlockHeight: 0,
		BlockSizeLimit:     cts.BlockSizeLimit,
	})
	if err != nil {
		t.Error(err)
	}

	// Violate fitsInABlock.
	data := make([]byte, cts.BlockSizeLimit)
	txn.ArbitraryData = data
	err = txn.ValidateTransaction(TransactionValidationContext{
		CurrentBlockHeight: 0,
		BlockSizeLimit:     cts.BlockSizeLimit,
	})
	if err == nil {
		t.Error("failed to trigger fitsInABlock error")
	}
	txn.ArbitraryData = nil

	// Violate noRepeats
	txn.CoinInputs = []CoinInput{{}, {}}
	err = txn.ValidateTransaction(TransactionValidationContext{
		CurrentBlockHeight: 0,
		BlockSizeLimit:     cts.BlockSizeLimit,
	})
	if err == nil {
		t.Error("failed to trigger noRepeats error")
	}
	txn.CoinInputs = nil

	// Violate followsMinimumValues
	txn.CoinOutputs = []CoinOutput{{}}
	err = txn.ValidateTransaction(TransactionValidationContext{
		CurrentBlockHeight: 0,
		BlockSizeLimit:     cts.BlockSizeLimit,
	})
	if err == nil {
		t.Error("failed to trigger followsMinimumValues error")
	}
}
