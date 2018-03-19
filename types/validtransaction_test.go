package types

import (
	"testing"
)

// TestTransactionFitsInABlock probes the fitsInABlock method of the
// Transaction type.
func TestTransactionFitsInABlock(t *testing.T) {
	// Try a transaction that will fit in a block, followed by one that won't.
	SetConstants(DefaultChainConstants())
	data := make([]byte, cts.BlockSizeLimit/2)
	txn := Transaction{ArbitraryData: [][]byte{data}}
	err := txn.fitsInABlock()
	if err != nil {
		t.Error(err)
	}
	data = make([]byte, cts.BlockSizeLimit)
	txn.ArbitraryData[0] = data
	err = txn.fitsInABlock()
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

// TestValudUnlockConditions probes the validUnlockConditions function.
func TestValidUnlockConditions(t *testing.T) {
	// The only thing to check is the timelock.
	uc := UnlockConditions{Timelock: 3}
	err := validUnlockConditions(uc, 2)
	if err != ErrTimelockNotSatisfied {
		t.Error(err)
	}
	err = validUnlockConditions(uc, 3)
	if err != nil {
		t.Error(err)
	}
	err = validUnlockConditions(uc, 4)
	if err != nil {
		t.Error(err)
	}
}

// TestTransactionValidUnlockConditions probes the validUnlockConditions method
// of the transaction type.
func TestTransactionValidUnlockConditions(t *testing.T) {
	// Create a transaction with each type of valid unlock condition.
	txn := Transaction{
		CoinInputs: []CoinInput{
			{UnlockConditions: UnlockConditions{Timelock: 3}},
		},
		BlockStakeInputs: []BlockStakeInput{
			{UnlockConditions: UnlockConditions{Timelock: 3}},
		},
	}
	err := txn.validUnlockConditions(4)
	if err != nil {
		t.Error(err)
	}

	// Try with illegal conditions in the siacoin inputs.
	txn.CoinInputs[0].UnlockConditions.Timelock = 5
	err = txn.validUnlockConditions(4)
	if err == nil {
		t.Error(err)
	}
	txn.CoinInputs[0].UnlockConditions.Timelock = 3

	// Try with illegal conditions in the siafund inputs.
	txn.BlockStakeInputs[0].UnlockConditions.Timelock = 5
	err = txn.validUnlockConditions(4)
	if err == nil {
		t.Error(err)
	}
	txn.BlockStakeInputs[0].UnlockConditions.Timelock = 3
}

// TestTransactionStandaloneValid probes the StandaloneValid method of the
// Transaction type.
func TestTransactionStandaloneValid(t *testing.T) {
	// Build a working transaction.
	var txn Transaction
	err := txn.StandaloneValid(0)
	if err != nil {
		t.Error(err)
	}

	// Violate fitsInABlock.
	data := make([]byte, cts.BlockSizeLimit)
	txn.ArbitraryData = [][]byte{data}
	err = txn.StandaloneValid(0)
	if err == nil {
		t.Error("failed to trigger fitsInABlock error")
	}
	txn.ArbitraryData = nil

	// Violate noRepeats
	txn.CoinInputs = []CoinInput{{}, {}}
	err = txn.StandaloneValid(0)
	if err == nil {
		t.Error("failed to trigger noRepeats error")
	}
	txn.CoinInputs = nil

	// Violate followsMinimumValues
	txn.CoinOutputs = []CoinOutput{{}}
	err = txn.StandaloneValid(0)
	if err == nil {
		t.Error("failed to trigger followsMinimumValues error")
	}
	txn.CoinOutputs = nil

	// Violate validUnlockConditions
	txn.CoinInputs = []CoinInput{{}}
	txn.CoinInputs[0].UnlockConditions.Timelock = 1
	err = txn.StandaloneValid(0)
	if err == nil {
		t.Error("failed to trigger validUnlockConditions error")
	}
	txn.CoinInputs = nil

	// Violate validSignatures
	txn.TransactionSignatures = []TransactionSignature{{}}
	err = txn.StandaloneValid(0)
	if err == nil {
		t.Error("failed to trigger validSignatures error")
	}
	txn.TransactionSignatures = nil
}
