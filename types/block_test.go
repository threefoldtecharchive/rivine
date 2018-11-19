package types

import (
	"bytes"
	"encoding/hex"
	"testing"

	"github.com/threefoldtech/rivine/crypto"
	"github.com/threefoldtech/rivine/pkg/encoding/rivbin"
	"github.com/threefoldtech/rivine/pkg/encoding/siabin"
)

// TestBlockHeader checks that BlockHeader returns the correct value, and that
// the hash is consistent with the old method for obtaining the hash.
// TODO: Nonce has been removed in favour of POBS: https://github.com/threefoldtech/rivine/commit/8bac48bb38776bbef9ef22956c3b9ae301e25334#diff-fd289e47592d409909487becb9d38925
func TestBlockHeader(t *testing.T) {
	var b Block
	b.ParentID[1] = 1
	b.POBSOutput = BlockStakeOutputIndexes{BlockHeight: 1, TransactionIndex: 1, OutputIndex: 0}
	b.Timestamp = 3
	b.MinerPayouts = []MinerPayout{{Value: NewCurrency64(4)}}
	b.Transactions = []Transaction{
		{
			Version:       TestnetChainConstants().DefaultTransactionVersion,
			ArbitraryData: []byte{'5'},
		},
	}

	id1 := b.ID()
	id2 := BlockID(crypto.HashBytes(siabin.Marshal(b.Header())))
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
	cts := TestnetChainConstants()

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
	b.MinerPayouts = append(b.MinerPayouts, MinerPayout{Value: cts.BlockCreatorFee})
	ids = append(ids, b.ID())
	b.MinerPayouts = append(b.MinerPayouts, MinerPayout{Value: cts.BlockCreatorFee})
	ids = append(ids, b.ID())
	b.Transactions = append(b.Transactions, Transaction{
		Version:   TestnetChainConstants().DefaultTransactionVersion,
		MinerFees: []Currency{cts.BlockCreatorFee},
	})
	ids = append(ids, b.ID())
	b.Transactions = append(b.Transactions, Transaction{
		Version:   TestnetChainConstants().DefaultTransactionVersion,
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
// TODO: CaluclateCoinbase has been removed in https://github.com/threefoldtech/rivine/commit/8675b2afff5f200fe6c7d3fca7c21811e65f446a#diff-fd289e47592d409909487becb9d38925
// TODO: Nonce has been removed in favour of POBS: https://github.com/threefoldtech/rivine/commit/8bac48bb38776bbef9ef22956c3b9ae301e25334#diff-fd289e47592d409909487becb9d38925
func TestHeaderID(t *testing.T) {
	cts := TestnetChainConstants()

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
	b.MinerPayouts = append(b.MinerPayouts, MinerPayout{Value: cts.BlockCreatorFee})
	blocks = append(blocks, b)
	b.MinerPayouts = append(b.MinerPayouts, MinerPayout{Value: cts.BlockCreatorFee})
	blocks = append(blocks, b)
	b.Transactions = append(b.Transactions, Transaction{
		Version:   TestnetChainConstants().DefaultTransactionVersion,
		MinerFees: []Currency{cts.BlockCreatorFee},
	})
	blocks = append(blocks, b)
	b.Transactions = append(b.Transactions, Transaction{
		Version:   TestnetChainConstants().DefaultTransactionVersion,
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
		Version:   TestnetChainConstants().DefaultTransactionVersion,
		MinerFees: []Currency{NewCurrency64(123)},
	}
	b.Transactions = append(b.Transactions, txn)
	if b.CalculateTotalMinerFees().Cmp(expected) != 0 {
		t.Error("total miner fees is miscalculated for a block with a single transaction")
	}

	// Add a single no-fee transaction and check again.
	txn = Transaction{
		Version:       TestnetChainConstants().DefaultTransactionVersion,
		ArbitraryData: []byte{'6'},
	}
	b.Transactions = append(b.Transactions, txn)
	if b.CalculateTotalMinerFees().Cmp(expected) != 0 {
		t.Error("total miner fees is miscalculated with empty transactions.")
	}

	// Add a transaction with multiple fees.
	expected = expected.Add(NewCurrency64(1 + 2 + 3))
	txn = Transaction{
		Version: TestnetChainConstants().DefaultTransactionVersion,
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
		Version:       TestnetChainConstants().DefaultTransactionVersion,
		ArbitraryData: []byte{'7'},
	}
	b.Transactions = append([]Transaction{txn}, b.Transactions...)
	if b.CalculateTotalMinerFees().Cmp(expected) != 0 {
		t.Error("total miner fees is miscalculated with empty transactions.")
	}
}

// TestBlockMinerPayoutID probes the MinerPayout function of the block type.
// TODO: CaluclateCoinbase has been removed in https://github.com/threefoldtech/rivine/commit/8675b2afff5f200fe6c7d3fca7c21811e65f446a#diff-fd289e47592d409909487becb9d38925
func TestBlockMinerPayoutID(t *testing.T) {
	cts := TestnetChainConstants()

	// Create a block with 2 miner payouts, and check that each payout has a
	// different id, and that the id is dependent on the block id.
	var ids []CoinOutputID
	b := Block{
		MinerPayouts: []MinerPayout{
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

// TestBlockSiaEncoding probes the MarshalSia and UnmarshalSia methods of the
// Block type.
func TestBlockSiaEncoding(t *testing.T) {
	cts := TestnetChainConstants()

	b := Block{
		MinerPayouts: []MinerPayout{
			{Value: cts.BlockCreatorFee},
			{Value: cts.BlockCreatorFee},
		},
	}
	var decB Block
	err := siabin.Unmarshal(siabin.Marshal(b), &decB)
	if err != nil {
		t.Fatal(err)
	}
	if len(decB.MinerPayouts) != len(b.MinerPayouts) ||
		decB.MinerPayouts[0].Value.Cmp(b.MinerPayouts[0].Value) != 0 ||
		decB.MinerPayouts[1].Value.Cmp(b.MinerPayouts[1].Value) != 0 {
		t.Fatal("block changed after encode/decode:", b, decB)
	}
}

// TestBlockRivineEncoding probes the MarshalRivine and UnmarshalRivine methods of the
// Block type.
func TestBlockRivineEncoding(t *testing.T) {
	cts := TestnetChainConstants()

	b := Block{
		MinerPayouts: []MinerPayout{
			{Value: cts.BlockCreatorFee},
			{Value: cts.BlockCreatorFee},
		},
	}
	var decB Block
	err := rivbin.Unmarshal(rivbin.Marshal(b), &decB)
	if err != nil {
		t.Fatal(err)
	}
	if len(decB.MinerPayouts) != len(b.MinerPayouts) ||
		decB.MinerPayouts[0].Value.Cmp(b.MinerPayouts[0].Value) != 0 ||
		decB.MinerPayouts[1].Value.Cmp(b.MinerPayouts[1].Value) != 0 {
		t.Fatal("block changed after encode/decode:", b, decB)
	}
}

// TestBlockIDAfterFixForBug302 ensures that the block ID is correct after all the condition/fulfillment changes
// part of issue https://github.com/threefoldtech/rivine/issues/302
func TestBlockIDAfterFixForBug302(t *testing.T) { // utility funcs
	hbs := func(str string) []byte { // hexStr -> byte slice
		bs, err := hex.DecodeString(str)
		if err != nil {
			panic(err)
		}
		return bs
	}
	hs := func(str string) (hash crypto.Hash) { // hbs -> crypto.Hash
		copy(hash[:], hbs(str))
		return
	}

	testCases := []struct {
		BlockID BlockID
		Block   Block
	}{
		{
			blockIDFromHex("76e55acb89d1a16514a74123160b79d9917995d87e2176668afcb1a9df53bd1d"),
			Block{
				ParentID:  blockIDFromHex("0000000000000000000000000000000000000000000000000000000000000000"),
				Timestamp: 1519200000,
				POBSOutput: BlockStakeOutputIndexes{
					BlockHeight:      0,
					TransactionIndex: 0,
					OutputIndex:      0,
				},
				MinerPayouts: nil,
				Transactions: []Transaction{
					{
						Version:    0,
						CoinInputs: nil,
						CoinOutputs: []CoinOutput{
							{
								Value: NewCurrency64(100000000000000000),
								Condition: UnlockConditionProxy{
									Condition: NewUnlockHashCondition(unlockHashFromHex(
										"01fc8714235d549f890f35e52d745b9eeeee34926f96c4b9ef1689832f338d9349b453898f7e51")),
								},
							},
						},
						BlockStakeInputs: nil,
						BlockStakeOutputs: []BlockStakeOutput{
							{
								Value: NewCurrency64(3000),
								Condition: UnlockConditionProxy{
									Condition: NewUnlockHashCondition(unlockHashFromHex(
										"01fc8714235d549f890f35e52d745b9eeeee34926f96c4b9ef1689832f338d9349b453898f7e51")),
								},
							},
						},
						MinerFees:     nil,
						ArbitraryData: nil,
					},
				},
			},
		},
		{
			blockIDFromHex("0624c4830353c83e75683ce47683d0acc11216e528cd46c87350c2516a24d743"),
			Block{
				ParentID:  blockIDFromHex("76e55acb89d1a16514a74123160b79d9917995d87e2176668afcb1a9df53bd1d"),
				Timestamp: 1522792547,
				POBSOutput: BlockStakeOutputIndexes{
					BlockHeight:      0,
					TransactionIndex: 0,
					OutputIndex:      0,
				},
				MinerPayouts: []MinerPayout{
					{
						Value:      NewCurrency64(10000000000),
						UnlockHash: unlockHashFromHex("01fc8714235d549f890f35e52d745b9eeeee34926f96c4b9ef1689832f338d9349b453898f7e51"),
					},
				},
				Transactions: []Transaction{
					{
						Version:     0,
						CoinInputs:  nil,
						CoinOutputs: nil,
						BlockStakeInputs: []BlockStakeInput{
							{
								ParentID: BlockStakeOutputID(hs("4cd0ec4f270ac5fe50ec3a4008ff652754400aff02af1caa8d1b5889ffa61292")),
								Fulfillment: UnlockFulfillmentProxy{
									Fulfillment: &SingleSignatureFulfillment{
										PublicKey: SiaPublicKey{
											Algorithm: SignatureEd25519,
											Key:       hbs("47c54e33cdfa770f2180af660d27881d1dbc544b37a4233390af603162c7433d"),
										},
										Signature: hbs("a17970cbf50fe1f9929f2dec6c60994f2ea9e6c30af1ae84e5a496171d86584c944853edab70b15a423d64857d62f96b21fa263c338c966dd9a1e4bdf6725d04"),
									},
								},
							},
						},
						BlockStakeOutputs: []BlockStakeOutput{
							{
								Value: NewCurrency64(3000),
								Condition: UnlockConditionProxy{
									Condition: NewUnlockHashCondition(unlockHashFromHex(
										"01fc8714235d549f890f35e52d745b9eeeee34926f96c4b9ef1689832f338d9349b453898f7e51")),
								},
							},
						},
						MinerFees:     nil,
						ArbitraryData: nil,
					},
				},
			},
		},
	}
	for idx, testCase := range testCases {
		blockID := testCase.Block.ID()
		if bytes.Compare(testCase.BlockID[:], blockID[:]) != 0 {
			t.Error(idx, testCase.BlockID, "!=", blockID)
		}
	}
}

func TestDecodeLegacyBlockAfterBug305(t *testing.T) {
	testCases := []string{
		// tfchain standard - block height 3
		// > https://explorer.threefoldtoken.com/explorer/blocks/3
		// {
		// 	"parentid": "026e4b5c1d54d6237ec7246e789ee34a63ae98540bb63971c2e7f19904f77758",
		// 	"timestamp": 1524169049,
		// 	"pobsindexes": {
		// 		"BlockHeight": 2,
		// 		"TransactionIndex": 0,
		// 		"OutputIndex": 0
		// 	},
		// 	"minerpayouts": [{
		// 		"value": "1000000000",
		// 		"unlockhash": "01ad4f73417476f8b8350298681dd0fa8640baa53a91915417b1dd8103d118b543c992e6fba1c4"
		// 	}],
		// 	"transactions": [{
		// 		"version": 0,
		// 		"data": {
		// 			"coininputs": null,
		// 			"blockstakeinputs": [{
		// 				"parentid": "441ac4150342765cb73c6656613ee2aaa79173cffa8b19fda3eadc5aadc3d682",
		// 				"unlocker": {
		// 					"type": 1,
		// 					"condition": {
		// 						"publickey": "ed25519:b5662caa078efd42b25f3ab10768b55fd0607ed8cb8e3c44f3b26df1d17ef934"
		// 					},
		// 					"fulfillment": {
		// 						"signature": "b30700ee21a314e0838e460649c2519ab58fe28413298bc5ee8feb2af0284f46e71eb24453a8f8c89c200735ee9100c7f859999dfe5dfc8be7964545d05e7008"
		// 					}
		// 				}
		// 			}],
		// 			"blockstakeoutputs": [{
		// 				"value": "100",
		// 				"unlockhash": "01ad4f73417476f8b8350298681dd0fa8640baa53a91915417b1dd8103d118b543c992e6fba1c4"
		// 			}],
		// 			"minerfees": null
		// 		}
		// 	}]
		// }
		`026e4b5c1d54d6237ec7246e789ee34a63ae98540bb63971c2e7f19904f7775859f9d85a00000000020000000000000000000000000000000000000000000000010000000000000004000000000000003b9aca0001ad4f73417476f8b8350298681dd0fa8640baa53a91915417b1dd8103d118b543010000000000000000000000000000000000000000000000000100000000000000441ac4150342765cb73c6656613ee2aaa79173cffa8b19fda3eadc5aadc3d682013800000000000000656432353531390000000000000000002000000000000000b5662caa078efd42b25f3ab10768b55fd0607ed8cb8e3c44f3b26df1d17ef9344000000000000000b30700ee21a314e0838e460649c2519ab58fe28413298bc5ee8feb2af0284f46e71eb24453a8f8c89c200735ee9100c7f859999dfe5dfc8be7964545d05e7008010000000000000001000000000000006401ad4f73417476f8b8350298681dd0fa8640baa53a91915417b1dd8103d118b54300000000000000000000000000000000`,
		// tfchain standard - block height 5
		// > https://explorer.threefoldtoken.com/explorer/blocks/5
		// {
		// 	"parentid": "d8db0191654e0ed1798879dd63f519f9054e9400341b40a93190dfe9ee5787d4",
		// 	"timestamp": 1524169770,
		// 	"pobsindexes": {
		// 		"BlockHeight": 4,
		// 		"TransactionIndex": 0,
		// 		"OutputIndex": 0
		// 	},
		// 	"minerpayouts": [{
		// 		"value": "1000000000",
		// 		"unlockhash": "01ad4f73417476f8b8350298681dd0fa8640baa53a91915417b1dd8103d118b543c992e6fba1c4"
		// 	}],
		// 	"transactions": [{
		// 		"version": 0,
		// 		"data": {
		// 			"coininputs": null,
		// 			"blockstakeinputs": [{
		// 				"parentid": "940a089ea20801e87faf8b70e710728c28443261cbf8dc00b33c76c5a21ec7bf",
		// 				"unlocker": {
		// 					"type": 1,
		// 					"condition": {
		// 						"publickey": "ed25519:b5662caa078efd42b25f3ab10768b55fd0607ed8cb8e3c44f3b26df1d17ef934"
		// 					},
		// 					"fulfillment": {
		// 						"signature": "c4ef3c95ad062c363ec3cc66ff67a296b54e02ea5b97781239ff75cbac8ae6e7e0e71b6d7f59f990488c44e80c5461eedee8cc7aed1cd0f22a055a1564f48607"
		// 					}
		// 				}
		// 			}],
		// 			"blockstakeoutputs": [{
		// 				"value": "100",
		// 				"unlockhash": "01ad4f73417476f8b8350298681dd0fa8640baa53a91915417b1dd8103d118b543c992e6fba1c4"
		// 			}],
		// 			"minerfees": null
		// 		}
		// 	}]
		// }
		`d8db0191654e0ed1798879dd63f519f9054e9400341b40a93190dfe9ee5787d42afcd85a00000000040000000000000000000000000000000000000000000000010000000000000004000000000000003b9aca0001ad4f73417476f8b8350298681dd0fa8640baa53a91915417b1dd8103d118b543010000000000000000000000000000000000000000000000000100000000000000940a089ea20801e87faf8b70e710728c28443261cbf8dc00b33c76c5a21ec7bf013800000000000000656432353531390000000000000000002000000000000000b5662caa078efd42b25f3ab10768b55fd0607ed8cb8e3c44f3b26df1d17ef9344000000000000000c4ef3c95ad062c363ec3cc66ff67a296b54e02ea5b97781239ff75cbac8ae6e7e0e71b6d7f59f990488c44e80c5461eedee8cc7aed1cd0f22a055a1564f48607010000000000000001000000000000006401ad4f73417476f8b8350298681dd0fa8640baa53a91915417b1dd8103d118b54300000000000000000000000000000000`,
		// tfchain standard - block height 42
		// > https://explorer.threefoldtoken.com/explorer/blocks/42
		// {
		// 	"parentid": "fb6f2fd3491fc4f9ffbbbdf9d2728f6c93307064748a323c4ea412c3f7411ca6",
		// 	"timestamp": 1524216449,
		// 	"pobsindexes": {
		// 		"BlockHeight": 41,
		// 		"TransactionIndex": 0,
		// 		"OutputIndex": 0
		// 	},
		// 	"minerpayouts": [{
		// 		"value": "1000000000",
		// 		"unlockhash": "01746677df456546d93729066dd88514e2009930f3eebac3c93d43c88a108f8f9aa9e7c6f58893"
		// 	}],
		// 	"transactions": [{
		// 		"version": 0,
		// 		"data": {
		// 			"coininputs": null,
		// 			"blockstakeinputs": [{
		// 				"parentid": "b33661b26a27cb24c87cb3f39e2dc91522fd070428d1ba97da7704a8ddd705b5",
		// 				"unlocker": {
		// 					"type": 1,
		// 					"condition": {
		// 						"publickey": "ed25519:8d368f6c457f1f7f49f4cb32636c1d34197c046f5398ea6661b0b4ecfe36a3cd"
		// 					},
		// 					"fulfillment": {
		// 						"signature": "e72783399189c5c1f2518bb4f81b91d2e0ccf4d2cd7b010ad8f75ef7beec74d80b5d822a6848b4765c1866cbc0acbbb1fb3b003b58a19c9f1191bb7f22bc1d0f"
		// 					}
		// 				}
		// 			}],
		// 			"blockstakeoutputs": [{
		// 				"value": "390",
		// 				"unlockhash": "01746677df456546d93729066dd88514e2009930f3eebac3c93d43c88a108f8f9aa9e7c6f58893"
		// 			}],
		// 			"minerfees": null
		// 		}
		// 	}]
		// }
		`fb6f2fd3491fc4f9ffbbbdf9d2728f6c93307064748a323c4ea412c3f7411ca681b2d95a00000000290000000000000000000000000000000000000000000000010000000000000004000000000000003b9aca0001746677df456546d93729066dd88514e2009930f3eebac3c93d43c88a108f8f9a010000000000000000000000000000000000000000000000000100000000000000b33661b26a27cb24c87cb3f39e2dc91522fd070428d1ba97da7704a8ddd705b50138000000000000006564323535313900000000000000000020000000000000008d368f6c457f1f7f49f4cb32636c1d34197c046f5398ea6661b0b4ecfe36a3cd4000000000000000e72783399189c5c1f2518bb4f81b91d2e0ccf4d2cd7b010ad8f75ef7beec74d80b5d822a6848b4765c1866cbc0acbbb1fb3b003b58a19c9f1191bb7f22bc1d0f01000000000000000200000000000000018601746677df456546d93729066dd88514e2009930f3eebac3c93d43c88a108f8f9a00000000000000000000000000000000`,
		// tfchain standard - block height 7535
		// > https://explorer.threefoldtoken.com/explorer/blocks/7535
		// {
		// 	"parentid": "a8176b8937312ae1b897945ea9305b68e1f19ba3310835b1dd3f4d968b12d076",
		// 	"timestamp": 1525101295,
		// 	"pobsindexes": {
		// 		"BlockHeight": 7504,
		// 		"TransactionIndex": 0,
		// 		"OutputIndex": 0
		// 	},
		// 	"minerpayouts": [{
		// 		"value": "1000000000",
		// 		"unlockhash": "01d82a5845555245610dfbebfe9a876f0892277114859dd6bcaf7c57d5f6667b430bb26a23eee3"
		// 	}],
		// 	"transactions": [{
		// 		"version": 0,
		// 		"data": {
		// 			"coininputs": null,
		// 			"blockstakeinputs": [{
		// 				"parentid": "c9fab10d20d746bcca39603b30f41c5d5610d86387e6a88966e5fcac8a2a078e",
		// 				"unlocker": {
		// 					"type": 1,
		// 					"condition": {
		// 						"publickey": "ed25519:f9118670ce11881d9c679c34f66e92380a085f460725d6bdb250fbb6472c5437"
		// 					},
		// 					"fulfillment": {
		// 						"signature": "1741d07a462e6b6596acbc07efecdebe1ae806afc03589e51300f4ec81986b75c18a53385c769269fbf3bd48a95f0d73a5b6fa45573f690c2ed8014f5e0bc805"
		// 					}
		// 				}
		// 			}],
		// 			"blockstakeoutputs": [{
		// 				"value": "1",
		// 				"unlockhash": "01d82a5845555245610dfbebfe9a876f0892277114859dd6bcaf7c57d5f6667b430bb26a23eee3"
		// 			}],
		// 			"minerfees": null
		// 		}
		// 	}]
		// }
		`a8176b8937312ae1b897945ea9305b68e1f19ba3310835b1dd3f4d968b12d076ef32e75a00000000501d00000000000000000000000000000000000000000000010000000000000004000000000000003b9aca0001d82a5845555245610dfbebfe9a876f0892277114859dd6bcaf7c57d5f6667b43010000000000000000000000000000000000000000000000000100000000000000c9fab10d20d746bcca39603b30f41c5d5610d86387e6a88966e5fcac8a2a078e013800000000000000656432353531390000000000000000002000000000000000f9118670ce11881d9c679c34f66e92380a085f460725d6bdb250fbb6472c543740000000000000001741d07a462e6b6596acbc07efecdebe1ae806afc03589e51300f4ec81986b75c18a53385c769269fbf3bd48a95f0d73a5b6fa45573f690c2ed8014f5e0bc805010000000000000001000000000000000101d82a5845555245610dfbebfe9a876f0892277114859dd6bcaf7c57d5f6667b4300000000000000000000000000000000`,
		// tfchain standard - block height 8143
		// > https://explorer.threefoldtoken.com/explorer/blocks/8143
		// {
		// 	"parentid": "12bb1fda9736c394aca93d1fcfe208c8808ccdc0acb8c34d488cd62a2ef1ee99",
		// 	"timestamp": 1525242351,
		// 	"pobsindexes": {
		// 		"BlockHeight": 8142,
		// 		"TransactionIndex": 0,
		// 		"OutputIndex": 0
		// 	},
		// 	"minerpayouts": [{
		// 		"value": "1000000000",
		// 		"unlockhash": "01746677df456546d93729066dd88514e2009930f3eebac3c93d43c88a108f8f9aa9e7c6f58893"
		// 	}, {
		// 		"value": "100000000",
		// 		"unlockhash": "017267221ef1947bb18506e390f1f9446b995acfb6d08d8e39508bb974d9830b8cb8fdca788e34"
		// 	}],
		// 	"transactions": [{
		// 		"version": 0,
		// 		"data": {
		// 			"coininputs": null,
		// 			"blockstakeinputs": [{
		// 				"parentid": "9c38904ce13b6437a1c45121ad7c5f34ce744eccab43b4a7276b2c69eaf562c0",
		// 				"unlocker": {
		// 					"type": 1,
		// 					"condition": {
		// 						"publickey": "ed25519:8d368f6c457f1f7f49f4cb32636c1d34197c046f5398ea6661b0b4ecfe36a3cd"
		// 					},
		// 					"fulfillment": {
		// 						"signature": "88c3dfdd5afb9337dfe53878f8666491d6f0bda5b6e2804cea5d1e86f8eb920c432e33d4d6dd382b86e69d5bc245f19651977cc7e98f0d6ee9f78358c04d5705"
		// 					}
		// 				}
		// 			}],
		// 			"blockstakeoutputs": [{
		// 				"value": "390",
		// 				"unlockhash": "01746677df456546d93729066dd88514e2009930f3eebac3c93d43c88a108f8f9aa9e7c6f58893"
		// 			}],
		// 			"minerfees": null
		// 		}
		// 	}, {
		// 		"version": 0,
		// 		"data": {
		// 			"coininputs": [{
		// 				"parentid": "27d30e2ec9814588323362132b0d9d73c19d9f7269e6ec7cbef340db390c7fe3",
		// 				"unlocker": {
		// 					"type": 1,
		// 					"condition": {
		// 						"publickey": "ed25519:fd66e536490d7e834fd1e655fb16d2b5ddb20c41369bc3fc613da7d0360cc13c"
		// 					},
		// 					"fulfillment": {
		// 						"signature": "120b95c41fb420777af37a8658ec5f71c3138312dcef6ae7a251ea5e2bf1c7133a3c094661d8fa0614507b5c6f1ad04fb1299bd4539509658b1a1a0f9ca9c30e"
		// 					}
		// 				}
		// 			}],
		// 			"coinoutputs": [{
		// 				"value": "12000000000000",
		// 				"unlockhash": "01922d994ede1ce5f90b39e73b1a03dcd4a4b7f4746e7dad602a3c8d6df190c923e6998ddbfb2f"
		// 			}, {
		// 				"value": "12000000000000",
		// 				"unlockhash": "01980da1071fb4f1095bb785bc4bd52832fc47222a772345ab41ac4d49d685500ef809e55960a8"
		// 			}, {
		// 				"value": "1000000000000000",
		// 				"unlockhash": "01fad166445b2fee3a2e32dc28bc5f7c7e27fbfa3786db53ba337f5bd9218f549f49ef6d06a917"
		// 			}, {
		// 				"value": "250000000000000",
		// 				"unlockhash": "01f6fbe425a9535b5d794b4b2ca7486ec3a76aec1cc816e20662fc2fb55d0d740e72cfb72ddb11"
		// 			}, {
		// 				"value": "685228263400000000",
		// 				"unlockhash": "016232383c2791b5f87722a3140f7a8774d691df8dc2d0ca84c3973c2a995d514a4ffefc521b6c"
		// 			}],
		// 			"minerfees": ["100000000"]
		// 		}
		// 	}]
		// }
		`12bb1fda9736c394aca93d1fcfe208c8808ccdc0acb8c34d488cd62a2ef1ee99ef59e95a00000000ce1f00000000000000000000000000000000000000000000020000000000000004000000000000003b9aca0001746677df456546d93729066dd88514e2009930f3eebac3c93d43c88a108f8f9a040000000000000005f5e100017267221ef1947bb18506e390f1f9446b995acfb6d08d8e39508bb974d9830b8c0200000000000000000000000000000000000000000000000001000000000000009c38904ce13b6437a1c45121ad7c5f34ce744eccab43b4a7276b2c69eaf562c00138000000000000006564323535313900000000000000000020000000000000008d368f6c457f1f7f49f4cb32636c1d34197c046f5398ea6661b0b4ecfe36a3cd400000000000000088c3dfdd5afb9337dfe53878f8666491d6f0bda5b6e2804cea5d1e86f8eb920c432e33d4d6dd382b86e69d5bc245f19651977cc7e98f0d6ee9f78358c04d570501000000000000000200000000000000018601746677df456546d93729066dd88514e2009930f3eebac3c93d43c88a108f8f9a0000000000000000000000000000000000010000000000000027d30e2ec9814588323362132b0d9d73c19d9f7269e6ec7cbef340db390c7fe3013800000000000000656432353531390000000000000000002000000000000000fd66e536490d7e834fd1e655fb16d2b5ddb20c41369bc3fc613da7d0360cc13c4000000000000000120b95c41fb420777af37a8658ec5f71c3138312dcef6ae7a251ea5e2bf1c7133a3c094661d8fa0614507b5c6f1ad04fb1299bd4539509658b1a1a0f9ca9c30e050000000000000006000000000000000ae9f7bcc00001922d994ede1ce5f90b39e73b1a03dcd4a4b7f4746e7dad602a3c8d6df190c92306000000000000000ae9f7bcc00001980da1071fb4f1095bb785bc4bd52832fc47222a772345ab41ac4d49d685500e0700000000000000038d7ea4c6800001fad166445b2fee3a2e32dc28bc5f7c7e27fbfa3786db53ba337f5bd9218f549f0600000000000000e35fa931a00001f6fbe425a9535b5d794b4b2ca7486ec3a76aec1cc816e20662fc2fb55d0d740e080000000000000009826b799e03ca00016232383c2791b5f87722a3140f7a8774d691df8dc2d0ca84c3973c2a995d514a000000000000000000000000000000000100000000000000040000000000000005f5e1000000000000000000`,
	}
	for idx, testCase := range testCases {
		b, err := hex.DecodeString(testCase)
		if err != nil {
			t.Error(idx, err)
			continue
		}
		var block Block
		err = siabin.Unmarshal(b, &block)
		if err != nil {
			t.Error(idx, err)
		}
	}
}

func blockIDFromHex(s string) (id BlockID) {
	b, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	if len(b) != len(id) {
		panic("wrong length")
	}
	copy(id[:], b[:])
	return
}
