package types

import (
	"testing"

	"github.com/rivine/rivine/crypto"
	"github.com/rivine/rivine/encoding"
)

// TestBlockHeader checks that BlockHeader returns the correct value, and that
// the hash is consistent with the old method for obtaining the hash.
// TODO: Nonce has been removed in favour of POBS: https://github.com/rivine/rivine/commit/8bac48bb38776bbef9ef22956c3b9ae301e25334#diff-fd289e47592d409909487becb9d38925
func TestBlockHeader(t *testing.T) {
	var b Block
	b.ParentID[1] = 1
	b.POBSOutput = BlockStakeOutputIndexes{BlockHeight: 1, TransactionIndex: 1, OutputIndex: 0}
	b.Timestamp = 3
	b.MinerPayouts = []CoinOutput{{Value: NewCurrency64(4)}}
	b.Transactions = []Transaction{
		{
			Version:       DefaultChainConstants().DefaultTransactionVersion,
			ArbitraryData: []byte{'5'},
		},
	}

	id1 := b.ID()
	id2 := BlockID(crypto.HashBytes(encoding.Marshal(b.Header())))
	id3 := BlockID(crypto.HashAll(
		b.ParentID,
		b.POBSOutput,
		b.Timestamp,
		b.MerkleRoot(),
	))

	if id1 != id2 || id2 != id3 || id3 != id1 {
		t.Error("Methods for getting block id don't return the same results:", id1, id2, id3)
	}
}

// TestBlockID probes the ID function of the block type.
func TestBlockID(t *testing.T) {
	cts := DefaultChainConstants()

	// Create a bunch of different blocks and check that all of them have
	// unique ids.
	var b Block
	var ids []BlockID

	ids = append(ids, b.ID())
	b.ParentID[0] = 1
	ids = append(ids, b.ID())

	b.POBSOutput = BlockStakeOutputIndexes{BlockHeight: 1, TransactionIndex: 1, OutputIndex: 0}
	ids = append(ids, b.ID())
	b.Timestamp = CurrentTimestamp()
	ids = append(ids, b.ID())
	b.MinerPayouts = append(b.MinerPayouts, CoinOutput{Value: cts.BlockCreatorFee})
	ids = append(ids, b.ID())
	b.MinerPayouts = append(b.MinerPayouts, CoinOutput{Value: cts.BlockCreatorFee})
	ids = append(ids, b.ID())
	b.Transactions = append(b.Transactions, Transaction{
		Version:   DefaultChainConstants().DefaultTransactionVersion,
		MinerFees: []Currency{cts.BlockCreatorFee},
	})
	ids = append(ids, b.ID())
	b.Transactions = append(b.Transactions, Transaction{
		Version:   DefaultChainConstants().DefaultTransactionVersion,
		MinerFees: []Currency{cts.BlockCreatorFee},
	})
	ids = append(ids, b.ID())

	knownIDs := make(map[BlockID]struct{})
	for i, id := range ids {
		_, exists := knownIDs[id]
		if exists {
			t.Error("id repeat for index", i)
		}
		knownIDs[id] = struct{}{}
	}
}

// TestHeaderID probes the ID function of the BlockHeader type.
// TODO: CaluclateCoinbase has been removed in https://github.com/rivine/rivine/commit/8675b2afff5f200fe6c7d3fca7c21811e65f446a#diff-fd289e47592d409909487becb9d38925
// TODO: Nonce has been removed in favour of POBS: https://github.com/rivine/rivine/commit/8bac48bb38776bbef9ef22956c3b9ae301e25334#diff-fd289e47592d409909487becb9d38925
func TestHeaderID(t *testing.T) {
	cts := DefaultChainConstants()

	// Create a bunch of different blocks and check that all of them have
	// unique ids.
	var blocks []Block
	var b Block

	blocks = append(blocks, b)
	b.ParentID[0] = 1
	blocks = append(blocks, b)
	b.POBSOutput = BlockStakeOutputIndexes{BlockHeight: 1, TransactionIndex: 1, OutputIndex: 0}
	blocks = append(blocks, b)
	b.Timestamp = CurrentTimestamp()
	blocks = append(blocks, b)
	b.MinerPayouts = append(b.MinerPayouts, CoinOutput{Value: cts.BlockCreatorFee})
	blocks = append(blocks, b)
	b.MinerPayouts = append(b.MinerPayouts, CoinOutput{Value: cts.BlockCreatorFee})
	blocks = append(blocks, b)
	b.Transactions = append(b.Transactions, Transaction{
		Version:   DefaultChainConstants().DefaultTransactionVersion,
		MinerFees: []Currency{cts.BlockCreatorFee},
	})
	blocks = append(blocks, b)
	b.Transactions = append(b.Transactions, Transaction{
		Version:   DefaultChainConstants().DefaultTransactionVersion,
		MinerFees: []Currency{cts.BlockCreatorFee},
	})
	blocks = append(blocks, b)

	knownIDs := make(map[BlockID]struct{})
	for i, block := range blocks {
		blockID := block.ID()
		headerID := block.Header().ID()
		if blockID != headerID {
			t.Error("headerID does not match blockID for index", i)
		}
		_, exists := knownIDs[headerID]
		if exists {
			t.Error("id repeat for index", i)
		}
		knownIDs[headerID] = struct{}{}
	}
}

// TestBlockCalculateTotalMinerFees probes the CalculateTotalMinerFees function of the block
// type.
func TestBlockCalculateTotalMinerFees(t *testing.T) {
	// All tests are done at height = 0.
	var coinbase Currency

	// Calculate the total miner fees on a block with 0 fees at height 0. Result should
	// be 300,000.
	var b Block
	if b.CalculateTotalMinerFees().Cmp(coinbase) != 0 {
		t.Error("total miner fees is miscalculated for an empty block")
	}

	// Calculate when there is a fee in a transcation.
	expected := coinbase.Add(NewCurrency64(123))
	txn := Transaction{
		Version:   DefaultChainConstants().DefaultTransactionVersion,
		MinerFees: []Currency{NewCurrency64(123)},
	}
	b.Transactions = append(b.Transactions, txn)
	if b.CalculateTotalMinerFees().Cmp(expected) != 0 {
		t.Error("total miner fees is miscalculated for a block with a single transaction")
	}

	// Add a single no-fee transaction and check again.
	txn = Transaction{
		Version:       DefaultChainConstants().DefaultTransactionVersion,
		ArbitraryData: []byte{'6'},
	}
	b.Transactions = append(b.Transactions, txn)
	if b.CalculateTotalMinerFees().Cmp(expected) != 0 {
		t.Error("total miner fees is miscalculated with empty transactions.")
	}

	// Add a transaction with multiple fees.
	expected = expected.Add(NewCurrency64(1 + 2 + 3))
	txn = Transaction{
		Version: DefaultChainConstants().DefaultTransactionVersion,
		MinerFees: []Currency{
			NewCurrency64(1),
			NewCurrency64(2),
			NewCurrency64(3),
		},
	}
	b.Transactions = append(b.Transactions, txn)
	if b.CalculateTotalMinerFees().Cmp(expected) != 0 {
		t.Error("total miner fees is miscalculated for a block with a single transaction")
	}

	// Add an empty transaction to the beginning.
	txn = Transaction{
		Version:       DefaultChainConstants().DefaultTransactionVersion,
		ArbitraryData: []byte{'7'},
	}
	b.Transactions = append([]Transaction{txn}, b.Transactions...)
	if b.CalculateTotalMinerFees().Cmp(expected) != 0 {
		t.Error("total miner fees is miscalculated with empty transactions.")
	}
}

// TestBlockMinerPayoutID probes the MinerPayout function of the block type.
// TODO: CaluclateCoinbase has been removed in https://github.com/rivine/rivine/commit/8675b2afff5f200fe6c7d3fca7c21811e65f446a#diff-fd289e47592d409909487becb9d38925
func TestBlockMinerPayoutID(t *testing.T) {
	cts := DefaultChainConstants()

	// Create a block with 2 miner payouts, and check that each payout has a
	// different id, and that the id is dependent on the block id.
	var ids []CoinOutputID
	b := Block{
		MinerPayouts: []CoinOutput{
			{Value: cts.BlockCreatorFee},
			{Value: cts.BlockCreatorFee},
		},
	}
	ids = append(ids, b.MinerPayoutID(1), b.MinerPayoutID(2))
	b.ParentID[0] = 1
	ids = append(ids, b.MinerPayoutID(1), b.MinerPayoutID(2))

	knownIDs := make(map[CoinOutputID]struct{})
	for i, id := range ids {
		_, exists := knownIDs[id]
		if exists {
			t.Error("id repeat for index", i)
		}
		knownIDs[id] = struct{}{}
	}
}

// TestBlockEncodes probes the MarshalSia and UnmarshalSia methods of the
// Block type.
func TestBlockEncoding(t *testing.T) {
	cts := DefaultChainConstants()

	b := Block{
		MinerPayouts: []CoinOutput{
			{Value: cts.BlockCreatorFee},
			{Value: cts.BlockCreatorFee},
		},
	}
	var decB Block
	err := encoding.Unmarshal(encoding.Marshal(b), &decB)
	if err != nil {
		t.Fatal(err)
	}
	if len(decB.MinerPayouts) != len(b.MinerPayouts) ||
		decB.MinerPayouts[0].Value.Cmp(b.MinerPayouts[0].Value) != 0 ||
		decB.MinerPayouts[1].Value.Cmp(b.MinerPayouts[1].Value) != 0 {
		t.Fatal("block changed after encode/decode:", b, decB)
	}
}
