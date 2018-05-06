package types

import (
	"testing"
)

// TestTransactionFitsInABlock_V0 probes the fitsInABlock method of the
// Transaction type.
func TestTransactionFitsInABlock_V0(t *testing.T) {
	blockSizeLimit := DefaultChainConstants().BlockSizeLimit
	data := make([]byte, blockSizeLimit/2)
	txn := Transaction{
		Version:       TransactionVersionZero,
		ArbitraryData: data}
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

// TestTransactionFitsInABlock_Vd probes the fitsInABlock method of the
// Transaction type.
func TestTransactionFitsInABlock_Vd(t *testing.T) {
	blockSizeLimit := DefaultChainConstants().BlockSizeLimit
	data := make([]byte, blockSizeLimit/2)
	txn := Transaction{
		Version:       DefaultChainConstants().DefaultTransactionVersion,
		ArbitraryData: data}
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

// TestTransactionFollowsMinimumValues_V0 probes the followsMinimumValues method
// of the Transaction type.
func TestTransactionFollowsMinimumValues_V0(t *testing.T) {
	// Start with a transaction that follows all of minimum-values rules.
	txn := Transaction{
		Version:           TransactionVersionZero,
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

// TestTransactionFollowsMinimumValues_Vd probes the followsMinimumValues method
// of the Transaction type.
func TestTransactionFollowsMinimumValues_Vd(t *testing.T) {
	// Start with a transaction that follows all of minimum-values rules.
	txn := Transaction{
		Version:           DefaultChainConstants().DefaultTransactionVersion,
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

// TestTransactionNoRepeats_V0 probes the noRepeats method of the Transaction
// type.
func TestTransactionNoRepeats_V0(t *testing.T) {
	// Try a transaction all the repeatable types but no conflicts.
	txn := Transaction{
		Version:          TransactionVersionZero,
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

// TestTransactionNoRepeats_Vd probes the noRepeats method of the Transaction
// type.
func TestTransactionNoRepeats_Vd(t *testing.T) {
	// Try a transaction all the repeatable types but no conflicts.
	txn := Transaction{
		Version:          DefaultChainConstants().DefaultTransactionVersion,
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
	txn := Transaction{Version: 42}

	// validation of unknown transactions should always succeed,
	// as no validation is applied here
	err := txn.ValidateTransaction(cts.BlockSizeLimit)
	if err != nil {
		t.Errorf("expected no error, but received: %v", err)
	}
	err = txn.ValidateTransaction(0)
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
	err := txn.ValidateTransaction(cts.BlockSizeLimit)
	if err != nil {
		t.Error(err)
	}

	// Violate fitsInABlock.
	data := make([]byte, cts.BlockSizeLimit)
	txn.ArbitraryData = data
	err = txn.ValidateTransaction(cts.BlockSizeLimit)
	if err == nil {
		t.Error("failed to trigger fitsInABlock error")
	}
	txn.ArbitraryData = nil

	// Violate noRepeats
	txn.CoinInputs = []CoinInput{{}, {}}
	err = txn.ValidateTransaction(cts.BlockSizeLimit)
	if err == nil {
		t.Error("failed to trigger noRepeats error")
	}
	txn.CoinInputs = nil

	// Violate followsMinimumValues
	txn.CoinOutputs = []CoinOutput{{}}
	err = txn.ValidateTransaction(cts.BlockSizeLimit)
	if err == nil {
		t.Error("failed to trigger followsMinimumValues error")
	}
}
