package modules

import (
	"testing"

	"github.com/threefoldtech/rivine/pkg/encoding/siabin"
	"github.com/threefoldtech/rivine/types"
)

// TestCalculateFee checks that the CalculateFee function is correctly tallying
// the number of fees in a transaction set.
func TestCalculateFee(t *testing.T) {
	t.Parallel()

	// Try calculating the fees on a nil transaction set.
	if CalculateFee(nil).Cmp(types.ZeroCurrency) != 0 {
		t.Error("CalculateFee is incorrectly handling nil input")
	}
	// Try a single transaction with no fees.
	txnSet := []types.Transaction{{}}
	if CalculateFee(txnSet).Cmp(types.ZeroCurrency) != 0 {
		t.Error("CalculateFee is not correctly calculating the fees on an empty transaction set")
	}
	cst := types.TestnetChainConstants()

	// Try a non-empty transaction.
	txnSet = []types.Transaction{{
		Version: cst.DefaultTransactionVersion,
		CoinOutputs: []types.CoinOutput{{
			Value: types.NewCurrency64(253e9),
		}},
	}}
	if CalculateFee(txnSet).Cmp(types.ZeroCurrency) != 0 {
		t.Error("CalculateFee is not correctly calculating the fees on a non-empty transaction set")
	}

	// Try a transaction set with a single miner fee.
	baseFee := types.NewCurrency64(12e3)
	txnSet = []types.Transaction{{
		Version: cst.DefaultTransactionVersion,
		MinerFees: []types.Currency{
			baseFee,
		},
	}}
	setLen := uint64(len(siabin.Marshal(txnSet)))
	expectedFee := baseFee.Div64(setLen)
	if CalculateFee(txnSet).Cmp(expectedFee) != 0 {
		t.Error("CalculateFee doesn't seem to be calculating the correct transaction fee")
	}

	// Try a transaction set with multiple transactions and multiple fees per
	// transaction.
	fee1 := types.NewCurrency64(1e6)
	fee2 := types.NewCurrency64(2e6)
	fee3 := types.NewCurrency64(3e6)
	fee4 := types.NewCurrency64(4e6)
	txnSet = []types.Transaction{
		{
			Version: cst.DefaultTransactionVersion,
			MinerFees: []types.Currency{
				fee1,
				fee2,
			},
		},
		{
			Version: cst.DefaultTransactionVersion,
			MinerFees: []types.Currency{
				fee3,
				fee4,
			},
		},
	}
	currencyLen := types.NewCurrency64(uint64(len(siabin.Marshal(txnSet))))
	multiExpectedFee := fee1.Add(fee2).Add(fee3).Add(fee4).Div(currencyLen)
	if CalculateFee(txnSet).Cmp(multiExpectedFee) != 0 {
		t.Error("got the wrong fee for a multi transaction set")
	}
}
