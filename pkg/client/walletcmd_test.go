package client

import (
	"bytes"
	"encoding/hex"
	"testing"

	"github.com/threefoldtech/rivine/crypto"
	"github.com/threefoldtech/rivine/types"
)

func TestParsePairedOutputs(t *testing.T) {
	// utility funcs
	hbs := func(str string) []byte { // hexStr -> byte slice
		bs, err := hex.DecodeString(str)
		if err != nil {
			t.Fatal(err)
		}
		return bs
	}
	hs := func(str string) (hash crypto.Hash) { // hbs -> crypto.Hash
		copy(hash[:], hbs(str))
		return
	}

	testCases := []struct {
		Arguments     []string
		ExpectedPairs []outputPair
	}{
		{nil, nil}, // error, not enough
		{[]string{"01746677df456546d93729066dd88514e2009930f3eebac3c93d43c88a108f8f9aa9e7c6f58893"}, nil}, // error not enough
		{
			[]string{"01746677df456546d93729066dd88514e2009930f3eebac3c93d43c88a108f8f9aa9e7c6f58893", "42"},
			[]outputPair{{
				Value: types.NewCurrency64(42000000000),
				Condition: types.NewCondition(types.NewUnlockHashCondition(types.UnlockHash{
					Type: types.UnlockTypePubKey,
					Hash: hs("746677df456546d93729066dd88514e2009930f3eebac3c93d43c88a108f8f9a"),
				})),
			}},
		}, // no error
		{
			[]string{"01746677df456546d93729066dd88514e2009930f3eebac3c93d43c88a108f8f9aa9e7c6f58893", "zz"},
			nil,
		}, // error, invalid (currency) value
		{
			[]string{
				"01746677df456546d93729066dd88514e2009930f3eebac3c93d43c88a108f8f9aa9e7c6f58893", "42",
				"01ad4f73417476f8b8350298681dd0fa8640baa53a91915417b1dd8103d118b543c992e6fba1c4",
			},
			nil,
		}, // error, arguments length needs to be even (paired)
		{
			[]string{
				"01746677df456546d93729066dd88514e2009930f3eebac3c93d43c88a108f8f9aa9e7c6f58893", "42",
				"01ad4f73417476f8b8350298681dd0fa8640baa53a91915417b1dd8103d118b543c992e6fba1c4", "10000",
			},
			[]outputPair{
				{
					Value: types.NewCurrency64(42000000000),
					Condition: types.NewCondition(types.NewUnlockHashCondition(types.UnlockHash{
						Type: types.UnlockTypePubKey,
						Hash: hs("746677df456546d93729066dd88514e2009930f3eebac3c93d43c88a108f8f9a"),
					})),
				},
				{
					Value: types.NewCurrency64(10000000000000),
					Condition: types.NewCondition(types.NewUnlockHashCondition(types.UnlockHash{
						Type: types.UnlockTypePubKey,
						Hash: hs("ad4f73417476f8b8350298681dd0fa8640baa53a91915417b1dd8103d118b543"),
					})),
				},
			},
		}, // no error
		{
			[]string{`{"type":1,"data":{"unlockhash":"01746677df456546d93729066dd88514e2009930f3eebac3c93d43c88a108f8f9aa9e7c6f58893"}}`, "4.2"},
			[]outputPair{{
				Value: types.NewCurrency64(4200000000),
				Condition: types.NewCondition(types.NewUnlockHashCondition(types.UnlockHash{
					Type: types.UnlockTypePubKey,
					Hash: hs("746677df456546d93729066dd88514e2009930f3eebac3c93d43c88a108f8f9a"),
				})),
			}},
		}, // no error, more explicit version of first non-error example
		{
			[]string{`{"type":3,"data":{"locktime":42,"condition":{"type":1,"data":{"unlockhash":"01746677df456546d93729066dd88514e2009930f3eebac3c93d43c88a108f8f9aa9e7c6f58893"}}}}`, "42"},
			[]outputPair{{
				Value: types.NewCurrency64(42000000000),
				Condition: types.NewCondition(types.NewTimeLockCondition(42, types.NewUnlockHashCondition(types.UnlockHash{
					Type: types.UnlockTypePubKey,
					Hash: hs("746677df456546d93729066dd88514e2009930f3eebac3c93d43c88a108f8f9a"),
				}))),
			}},
		}, // no error, more explicit version of first non-error example, combined with a timelock
		{
			[]string{
				`01ad4f73417476f8b8350298681dd0fa8640baa53a91915417b1dd8103d118b543c992e6fba1c4`, "1000",
				`{"type":3,"data":{"locktime":42,"condition":{"type":1,"data":{"unlockhash":"01746677df456546d93729066dd88514e2009930f3eebac3c93d43c88a108f8f9aa9e7c6f58893"}}}}`, "90000.50",
				`01746677df456546d93729066dd88514e2009930f3eebac3c93d43c88a108f8f9aa9e7c6f58893`, "200",
				`{
					"type": 4,
					"data": {
						"unlockhashes": [
							"01746677df456546d93729066dd88514e2009930f3eebac3c93d43c88a108f8f9aa9e7c6f58893",
							"01ad4f73417476f8b8350298681dd0fa8640baa53a91915417b1dd8103d118b543c992e6fba1c4"
						],
						"minimumsignaturecount": 1
					}
				}`, "12.345",
				`{}`, "100",
			},
			[]outputPair{
				{
					Value: types.NewCurrency64(1000000000000),
					Condition: types.NewCondition(types.NewUnlockHashCondition(types.UnlockHash{
						Type: types.UnlockTypePubKey,
						Hash: hs("ad4f73417476f8b8350298681dd0fa8640baa53a91915417b1dd8103d118b543"),
					})),
				},
				{
					Value: types.NewCurrency64(90000500000000),
					Condition: types.NewCondition(types.NewTimeLockCondition(42, types.NewUnlockHashCondition(types.UnlockHash{
						Type: types.UnlockTypePubKey,
						Hash: hs("746677df456546d93729066dd88514e2009930f3eebac3c93d43c88a108f8f9a"),
					}))),
				},
				{
					Value: types.NewCurrency64(200000000000),
					Condition: types.NewCondition(types.NewUnlockHashCondition(types.UnlockHash{
						Type: types.UnlockTypePubKey,
						Hash: hs("746677df456546d93729066dd88514e2009930f3eebac3c93d43c88a108f8f9a"),
					})),
				},
				{
					Value: types.NewCurrency64(12345000000),
					Condition: types.NewCondition(types.NewMultiSignatureCondition(types.UnlockHashSlice{
						types.UnlockHash{
							Type: types.UnlockTypePubKey,
							Hash: hs("746677df456546d93729066dd88514e2009930f3eebac3c93d43c88a108f8f9a"),
						},
						types.UnlockHash{
							Type: types.UnlockTypePubKey,
							Hash: hs("ad4f73417476f8b8350298681dd0fa8640baa53a91915417b1dd8103d118b543"),
						},
					}, 1)),
				},
				{
					Value:     types.NewCurrency64(100000000000),
					Condition: types.NewCondition(&types.NilCondition{}),
				},
			},
		}, // no error, a more complex example
	}
	for idx, testCase := range testCases {
		pairs, err := parsePairedOutputs(testCase.Arguments, createDefaultCurrencyConvertor().ParseCoinString)
		if len(testCase.ExpectedPairs) == 0 {
			// expecting error
			if err == nil {
				t.Error("expected error, but received none for idx: ", idx)
			}
			continue
		}
		if err != nil {
			t.Errorf("received unexpected error for idx %d: %v", idx, err)
		}
		for i, pair := range pairs {
			if !pair.Value.Equals(testCase.ExpectedPairs[i].Value) {
				t.Errorf("unexpected pair (currency) value for pair %d/#%d: '%v' != '%v'",
					idx, i, testCase.ExpectedPairs[i].Value, pair.Value)
			}
			bsA, err := pair.Condition.MarshalJSON()
			if err != nil {
				t.Errorf("failed to JSON encode output pair %d/#%d's condition: %v", idx, i, err)
			}
			bsB, err := testCase.ExpectedPairs[i].Condition.MarshalJSON()
			if err != nil {
				t.Errorf("failed to JSON encode expected pair %d/#%d's condition: %v", idx, i, err)
			}
			if !bytes.Equal(bsA, bsB) {
				t.Errorf("unexpected pair (currency) value for pair %d/#%d: '%s' != '%s'",
					idx, i, string(bsA), string(bsB))
			}
		}
	}
}

func createDefaultCurrencyConvertor() CurrencyConvertor {
	bchainInfo := types.DefaultBlockchainInfo()
	return NewCurrencyConvertor(types.DefaultCurrencyUnits(), bchainInfo.CoinUnit)
}
