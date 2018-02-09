package types

import (
	"testing"

	"github.com/rivine/rivine/crypto"
)

// TestTransactionIDs probes all of the ID functions of the Transaction type.
func TestIDs(t *testing.T) {
	// Create every type of ID using empty fields.
	txn := Transaction{
		CoinOutputs:       []CoinOutput{{}},
		BlockStakeOutputs: []BlockStakeOutput{{}},
	}
	tid := txn.ID()
	scoid := txn.CoinOutputID(0)
	sfoid := txn.BlockStakeOutputID(0)

	// Put all of the ids into a slice.
	var ids []crypto.Hash
	ids = append(ids,
		crypto.Hash(tid),
		crypto.Hash(scoid),
		crypto.Hash(sfoid),
	)

	// Check that each id is unique.
	knownIDs := make(map[crypto.Hash]struct{})
	for i, id := range ids {
		_, exists := knownIDs[id]
		if exists {
			t.Error("id repeat for index", i)
		}
		knownIDs[id] = struct{}{}
	}
}

// TestTransactionCoinOutputSum probes the CoinOutputSum method of the
// Transaction type.
func TestTransactionCoinOutputSum(t *testing.T) {
	// Create a transaction with all types of coin outputs.
	txn := Transaction{
		CoinOutputs: []CoinOutput{
			{Value: NewCurrency64(1)},
			{Value: NewCurrency64(20)},
		},
		MinerFees: []Currency{
			NewCurrency64(50000),
			NewCurrency64(600000),
		},
	}
	if txn.CoinOutputSum().Cmp(NewCurrency64(650021)) != 0 {
		t.Error("wrong coin output sum was calculated, got:", txn.CoinOutputSum())
	}
}
