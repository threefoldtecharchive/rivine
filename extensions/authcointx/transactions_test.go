package authcointx

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/threefoldtech/rivine/crypto"
	"github.com/threefoldtech/rivine/pkg/encoding/siabin"
	"github.com/threefoldtech/rivine/types"
)

func TestJSONExampleAuthAddressUpdateTransaction(t *testing.T) {
	const authAddressUpdateTxVersion types.TransactionVersion = 176
	types.RegisterTransactionVersion(authAddressUpdateTxVersion, AuthAddressUpdateTransactionController{TransactionVersion: authAddressUpdateTxVersion})
	defer types.RegisterTransactionVersion(authAddressUpdateTxVersion, nil)

	const jsonEncodedExample = `{
	"version": 176, 
	"data": {	
		"nonce": "FoAiO8vN2eU=",
		"authaddresses": [
			"0112210f9efa5441ab705226b0628679ed190eb4588b662991747ea3809d93932c7b41cbe4b732",
			"01450aeb140c58012cb4afb48e068f976272fefa44ffe0991a8a4350a3687558d66c8fc753c37e"
		],
		"deauthaddresses": [
			"019e9b6f2d43a44046b62836ce8d75c935ff66cbba1e624b3e9755b98ac176a08dac5267b2c8ee"
		],
		"arbitrarydata": "dGVzdC4uLiAxLCAyLi4uIDM=",
		"authfulfillment": {
			"type": 1,
			"data": {
				"publickey": "ed25519:d285f92d6d449d9abb27f4c6cf82713cec0696d62b8c123f1627e054dc6d7780",
				"signature": "bdf023fbe7e0efec584d254b111655e1c2f81b9488943c3a712b91d9ad3a140cb0949a8868c5f72e08ccded337b79479114bdb4ed05f94dfddb359e1a6124602"
			}
		}
	}
}`

	var tx types.Transaction
	err := json.Unmarshal([]byte(jsonEncodedExample), &tx)
	if err != nil {
		t.Fatal(err)
	}
	b, err := json.Marshal(tx)
	if err != nil {
		t.Fatal(err)
	}
	output := string(b)
	buffer := bytes.NewBuffer(nil)
	err = json.Compact(buffer, []byte(jsonEncodedExample))
	if err != nil {
		t.Fatal(err)
	}
	expectedOutput := buffer.String()
	if expectedOutput != output {
		t.Fatal(expectedOutput, "!=", output)
	}
}

func TestBinaryExampleAuthAddressUpdateTransaction(t *testing.T) {
	const authAddressUpdateTxVersion types.TransactionVersion = 176
	types.RegisterTransactionVersion(authAddressUpdateTxVersion, AuthAddressUpdateTransactionController{TransactionVersion: authAddressUpdateTxVersion})
	defer types.RegisterTransactionVersion(authAddressUpdateTxVersion, nil)

	const hexEncodedExample = `b0d8c18020e5cfb2da020101f68299b26a89efdb4351a61c3a062321d23edbc1399c8499947c1313375609020101f68299b26a89efdb4351a61c3a062321d23edbc1399c8499947c131337560900014401336f56368308d77b186e3dab7f8b09f18b4012a823593bb7bde09bebfa1f89820000`

	b, err := hex.DecodeString(hexEncodedExample)
	if err != nil {
		t.Fatal(err)
	}
	var tx types.Transaction
	err = siabin.Unmarshal(b, &tx)
	if err != nil {
		t.Fatal(err)
	}

	b, err = siabin.Marshal(tx)
	if err != nil {
		t.Fatal(err)
	}
	output := hex.EncodeToString(b)
	if hexEncodedExample != output {
		t.Fatal(hexEncodedExample, "!=", output)
	}
}

func TestJSONExampleAuthConditionUpdateTransaction(t *testing.T) {
	const authConditionUpdateTxVersion types.TransactionVersion = 177
	types.RegisterTransactionVersion(authConditionUpdateTxVersion, AuthConditionUpdateTransactionController{TransactionVersion: authConditionUpdateTxVersion})
	defer types.RegisterTransactionVersion(authConditionUpdateTxVersion, nil)

	const jsonEncodedExample = `{
	"version": 177, 
	"data": {
		"nonce": "1oQFzIwsLs8=",
		"arbitrarydata": "dGVzdC4uLiAxLCAyLi4uIDM=",
		"authcondition": {
			"type": 1,
			"data": {
				"unlockhash": "01e78fd5af261e49643dba489b29566db53fa6e195fa0e6aad4430d4f06ce88b73e047fe6a0703"
			}
		},
		"authfulfillment": {
			"type": 1,
			"data": {
				"publickey": "ed25519:d285f92d6d449d9abb27f4c6cf82713cec0696d62b8c123f1627e054dc6d7780",
				"signature": "ad59389329ed01c5ee14ce25ae38634c2b3ef694a2bdfa714f73b175f979ba6613025f9123d68c0f11e8f0a7114833c0aab4c8596d4c31671ec8a73923f02305"
			}
		}
	}
}`

	var tx types.Transaction
	err := json.Unmarshal([]byte(jsonEncodedExample), &tx)
	if err != nil {
		t.Fatal(err)
	}
	b, err := json.Marshal(tx)
	if err != nil {
		t.Fatal(err)
	}
	output := string(b)
	buffer := bytes.NewBuffer(nil)
	err = json.Compact(buffer, []byte(jsonEncodedExample))
	if err != nil {
		t.Fatal(err)
	}
	expectedOutput := buffer.String()
	if expectedOutput != output {
		t.Fatal(expectedOutput, "!=", output)
	}
}

func TestBinaryExampleAuthConditionUpdateTransaction(t *testing.T) {
	const authConditionUpdateTxVersion types.TransactionVersion = 177
	types.RegisterTransactionVersion(authConditionUpdateTxVersion, AuthConditionUpdateTransactionController{TransactionVersion: authConditionUpdateTxVersion})
	defer types.RegisterTransactionVersion(authConditionUpdateTxVersion, nil)

	const hexEncodedExample = `b16b177d2b892fd85f0001420101f68299b26a89efdb4351a61c3a062321d23edbc1399c8499947c1313375609014401336f56368308d77b186e3dab7f8b09f18b4012a823593bb7bde09bebfa1f89820000`

	b, err := hex.DecodeString(hexEncodedExample)
	if err != nil {
		t.Fatal(err)
	}
	var tx types.Transaction
	err = siabin.Unmarshal(b, &tx)
	if err != nil {
		t.Fatal(err)
	}

	b, err = siabin.Marshal(tx)
	if err != nil {
		t.Fatal(err)
	}
	output := hex.EncodeToString(b)
	if hexEncodedExample != output {
		t.Fatal(hexEncodedExample, "!=", output)
	}
}

func TestAuthStandardTransactionEncodingDocExamples(t *testing.T) {
	// utility funcs
	hbs := func(str string) []byte { // hexStr -> byte slice
		bs, _ := hex.DecodeString(str)
		return bs
	}
	hs := func(str string) (hash crypto.Hash) { // hbs -> crypto.Hash
		copy(hash[:], hbs(str))
		return
	}

	// examples
	examples := []struct {
		HexEncoding string
		ExpectedTx  types.Transaction
	}{
		// v1 @ v1.0.4
		{
			"01e20000000000000001000000000000002200000000000000000000000000000000000000000000000000000000000022018000000000000000656432353531390000000000000000002000000000000000ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff4000000000000000ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff00000000000000000000000000000000000000000000000001000000000000000100000000000000010000000000000000",
			types.Transaction{
				Version: types.TransactionVersionOne,
				CoinInputs: []types.CoinInput{
					{
						ParentID: types.CoinOutputID(hs("2200000000000000000000000000000000000000000000000000000000000022")),
						Fulfillment: types.NewFulfillment(&types.SingleSignatureFulfillment{
							PublicKey: types.PublicKey{
								Algorithm: types.SignatureAlgoEd25519,
								Key:       hbs("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"),
							},
							Signature: hbs("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"),
						}),
					},
				},
				MinerFees: []types.Currency{types.NewCurrency64(1)},
			},
		},
		{
			"01f40000000000000001000000000000001100000000000000000000000000000000000000000000000000000000000011018000000000000000656432353531390000000000000000002000000000000000ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff4000000000000000ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff01000000000000000100000000000000090000000000000000000000000000000000000000000000000001000000000000000100000000000000030000000000000000",
			types.Transaction{
				Version: types.TransactionVersionOne,
				CoinInputs: []types.CoinInput{
					{
						ParentID: types.CoinOutputID(hs("1100000000000000000000000000000000000000000000000000000000000011")),
						Fulfillment: types.NewFulfillment(&types.SingleSignatureFulfillment{
							PublicKey: types.PublicKey{
								Algorithm: types.SignatureAlgoEd25519,
								Key:       hbs("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"),
							},
							Signature: hbs("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"),
						}),
					},
				},
				CoinOutputs: []types.CoinOutput{
					{
						Value:     types.NewCurrency64(9),
						Condition: types.NewCondition(&types.NilCondition{}), // `nil` would be functionally equal, but it will give a non-deep-equal result
					},
				},
				MinerFees: []types.Currency{types.NewCurrency64(3)},
			},
		},
		{
			"01480100000000000001000000000000002200000000000000000000000000000000000000000000000000000000000022018000000000000000656432353531390000000000000000002000000000000000ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff4000000000000000ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff020000000000000001000000000000000201210000000000000001cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc01000000000000000301210000000000000001dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd0000000000000000000000000000000001000000000000000100000000000000010000000000000000",
			types.Transaction{
				Version: types.TransactionVersionOne,
				CoinInputs: []types.CoinInput{
					{
						ParentID: types.CoinOutputID(hs("2200000000000000000000000000000000000000000000000000000000000022")),
						Fulfillment: types.NewFulfillment(&types.SingleSignatureFulfillment{
							PublicKey: types.PublicKey{
								Algorithm: types.SignatureAlgoEd25519,
								Key:       hbs("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"),
							},
							Signature: hbs("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"),
						}),
					},
				},
				CoinOutputs: []types.CoinOutput{
					{
						Value: types.NewCurrency64(2),
						Condition: types.NewCondition(types.NewUnlockHashCondition(types.UnlockHash{
							Type: types.UnlockTypePubKey,
							Hash: hs("cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"),
						})),
					},
					{
						Value: types.NewCurrency64(3),
						Condition: types.NewCondition(types.NewUnlockHashCondition(types.UnlockHash{
							Type: types.UnlockTypePubKey,
							Hash: hs("dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd"),
						})),
					},
				},
				MinerFees: []types.Currency{types.NewCurrency64(1)},
			},
		},
		{
			"01fd0000000000000001000000000000001100000000000000000000000000000000000000000000000000000000000011018000000000000000656432353531390000000000000000002000000000000000ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff4000000000000000ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff01000000000000000100000000000000090309000000000000002a00000000000000000000000000000000000000000000000001000000000000000100000000000000030000000000000000",
			types.Transaction{
				Version: types.TransactionVersionOne,
				CoinInputs: []types.CoinInput{
					{
						ParentID: types.CoinOutputID(hs("1100000000000000000000000000000000000000000000000000000000000011")),
						Fulfillment: types.NewFulfillment(&types.SingleSignatureFulfillment{
							PublicKey: types.PublicKey{
								Algorithm: types.SignatureAlgoEd25519,
								Key:       hbs("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"),
							},
							Signature: hbs("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"),
						}),
					},
				},
				CoinOutputs: []types.CoinOutput{
					{
						Value: types.NewCurrency64(9),
						Condition: types.NewCondition(&types.TimeLockCondition{
							LockTime:  42,
							Condition: &types.NilCondition{}, // `nil` would be functionally equal, but it will give a non-deep-equal result
						}),
					},
				},
				MinerFees: []types.Currency{types.NewCurrency64(3)},
			},
		},
		{
			"011e0100000000000001000000000000001100000000000000000000000000000000000000000000000000000000000011018000000000000000656432353531390000000000000000002000000000000000ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff4000000000000000ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff0100000000000000010000000000000009032a000000000000002a000000000000000101abcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcd0000000000000000000000000000000001000000000000000100000000000000030000000000000000",
			types.Transaction{
				Version: types.TransactionVersionOne,
				CoinInputs: []types.CoinInput{
					{
						ParentID: types.CoinOutputID(hs("1100000000000000000000000000000000000000000000000000000000000011")),
						Fulfillment: types.NewFulfillment(&types.SingleSignatureFulfillment{
							PublicKey: types.PublicKey{
								Algorithm: types.SignatureAlgoEd25519,
								Key:       hbs("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"),
							},
							Signature: hbs("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"),
						}),
					},
				},
				CoinOutputs: []types.CoinOutput{
					{
						Value: types.NewCurrency64(9),
						Condition: types.NewCondition(&types.TimeLockCondition{
							LockTime: 42,
							Condition: types.NewUnlockHashCondition(types.UnlockHash{
								Type: types.UnlockTypePubKey,
								Hash: hs("abcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcd"),
							}),
						}),
					},
				},
				MinerFees: []types.Currency{types.NewCurrency64(3)},
			},
		},
		{
			"018101000000000000010000000000000022000000000000000000000000000000000000000000000000000000000000220388000000000000000100000000000000656432353531390000000000000000002000000000000000ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff4000000000000000ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff020000000000000001000000000000000201210000000000000001cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc0100000000000000030452000000000000000200000000000000020000000000000001dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd01bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb0000000000000000000000000000000001000000000000000100000000000000010000000000000000",
			types.Transaction{
				Version: types.TransactionVersionOne,
				CoinInputs: []types.CoinInput{
					{
						ParentID: types.CoinOutputID(hs("2200000000000000000000000000000000000000000000000000000000000022")),
						Fulfillment: types.NewFulfillment(&types.MultiSignatureFulfillment{
							Pairs: []types.PublicKeySignaturePair{
								{
									PublicKey: types.PublicKey{
										Algorithm: types.SignatureAlgoEd25519,
										Key:       hbs("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"),
									},
									Signature: hbs("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"),
								},
							},
						}),
					},
				},
				CoinOutputs: []types.CoinOutput{
					{
						Value: types.NewCurrency64(2),
						Condition: types.NewCondition(types.NewUnlockHashCondition(types.UnlockHash{
							Type: types.UnlockTypePubKey,
							Hash: hs("cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"),
						})),
					},
					{
						Value: types.NewCurrency64(3),
						Condition: types.NewCondition(types.NewMultiSignatureCondition(types.UnlockHashSlice{
							{
								Type: types.UnlockTypePubKey,
								Hash: hs("dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd"),
							},
							{
								Type: types.UnlockTypePubKey,
								Hash: hs("bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"),
							},
						}, 2)),
					},
				},
				MinerFees: []types.Currency{types.NewCurrency64(1)},
			},
		},
	}
	for idx, example := range examples {
		encodedTx, err := hex.DecodeString(example.HexEncoding)
		if err != nil {
			t.Error(idx, err)
			continue
		}

		var tx types.Transaction
		err = tx.UnmarshalSia(bytes.NewReader(encodedTx))
		if err != nil {
			t.Error(idx, err)
			continue
		}

		jms := func(v interface{}) string {
			bs, err := json.Marshal(v)
			if err != nil {
				t.Error(err)
			}
			return string(bs)
		}

		if !reflect.DeepEqual(example.ExpectedTx, tx) {
			t.Errorf("wrong tx hex decoding of example #%d: %v != %v", idx, jms(example.ExpectedTx), jms(tx))
			// go through each input/output to compare
			for cidx, ci := range example.ExpectedTx.CoinInputs {
				t1 := fmt.Sprintf("%T", ci.Fulfillment.Fulfillment)
				t2 := fmt.Sprintf("%T", tx.CoinInputs[cidx].Fulfillment.Fulfillment)
				if t1 != t2 {
					t.Error(idx, "coin input #", cidx, ":", t1, "!=", t2)
				}
			}
			for codx, co := range example.ExpectedTx.CoinOutputs {
				t1 := fmt.Sprintf("%T", co.Condition.Condition)
				t2 := fmt.Sprintf("%T", tx.CoinOutputs[codx].Condition.Condition)
				if t1 != t2 {
					t.Error(idx, "coin output #", codx, ":", t1, "!=", t2)
				}
			}
			for bsidx, bsi := range example.ExpectedTx.BlockStakeInputs {
				t1 := fmt.Sprintf("%T", bsi.Fulfillment.Fulfillment)
				t2 := fmt.Sprintf("%T", tx.BlockStakeInputs[bsidx].Fulfillment.Fulfillment)
				if t1 != t2 {
					t.Error(idx, "coin input #", bsidx, ":", t1, "!=", t2)
				}
			}
			for bsodx, bso := range example.ExpectedTx.BlockStakeOutputs {
				t1 := fmt.Sprintf("%T", bso.Condition.Condition)
				t2 := fmt.Sprintf("%T", tx.BlockStakeOutputs[bsodx].Condition.Condition)
				if t1 != t2 {
					t.Error(idx, "coin output #", bsodx, ":", t1, "!=", t2)
				}
			}
		}

		buf := bytes.NewBuffer(nil)
		err = tx.MarshalSia(buf)
		if err != nil {
			t.Error(idx, err)
			continue
		}

		if hexEncoding := hex.EncodeToString(buf.Bytes()); example.HexEncoding != hexEncoding {
			t.Errorf("wrong hex encoding of example #%d: %q != %q", idx, example.HexEncoding, hexEncoding)
		}
	}
}

// standard tests copied and modified from Rivine's types/transactions.test

func TestAuthStandardTransactionJSONEncodingExamples(t *testing.T) {
	// utility funcs
	hbs := func(str string) []byte { // hexStr -> byte slice
		bs, _ := hex.DecodeString(str)
		return bs
	}
	hs := func(str string) (hash crypto.Hash) { // hbs -> crypto.Hash
		copy(hash[:], hbs(str))
		return
	}

	// examples
	examples := []struct {
		JSONEncoded string
		ExpectedTx  types.Transaction
	}{
		// v1 @ v1.0.4
		{
			`{
	"version": 1,
	"data": {
		"coininputs": [
			{
				"parentid": "1100000000000000000000000000000000000000000000000000000000000011",
				"fulfillment": {
					"type": 1,
					"data": {
						"publickey": "ed25519:def123def123def123def123def123def123def123def123def123def123def1",
						"signature": "ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef"
					}
				}
			}
		],
		"coinoutputs": [
			{
				"value": "9",
				"condition": {}
			}
		],
		"minerfees": [
			"3"
		]
	}
}`,
			types.Transaction{
				Version: types.TransactionVersionOne,
				CoinInputs: []types.CoinInput{
					{
						ParentID: types.CoinOutputID(hs("1100000000000000000000000000000000000000000000000000000000000011")),
						Fulfillment: types.NewFulfillment(&types.SingleSignatureFulfillment{
							PublicKey: types.PublicKey{
								Algorithm: types.SignatureAlgoEd25519,
								Key:       hbs("def123def123def123def123def123def123def123def123def123def123def1"),
							},
							Signature: hbs("ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef"),
						}),
					},
				},
				CoinOutputs: []types.CoinOutput{
					{
						Value:     types.NewCurrency64(9),
						Condition: types.NewCondition(&types.NilCondition{}),
					},
				},
				MinerFees: []types.Currency{types.NewCurrency64(3)},
			},
		},
		{
			`{
	"version": 1,
	"data": {
		"coininputs": [
			{
				"parentid": "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
				"fulfillment": {
					"type": 1,
					"data": {
						"publickey": "ed25519:def123def123def123def123def123def123def123def123def123def123def1",
						"signature": "ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef"
					}
				}
			}
		],
		"minerfees": ["1"],
		"arbitrarydata": "SGVsbG8sIFdvcmxkIQ=="
	}
}`,
			types.Transaction{
				Version: types.TransactionVersionOne,
				CoinInputs: []types.CoinInput{
					{
						ParentID: types.CoinOutputID(hs("0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")),
						Fulfillment: types.NewFulfillment(&types.SingleSignatureFulfillment{
							PublicKey: types.PublicKey{
								Algorithm: types.SignatureAlgoEd25519,
								Key:       hbs("def123def123def123def123def123def123def123def123def123def123def1"),
							},
							Signature: hbs("ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef"),
						}),
					},
				},
				MinerFees:     []types.Currency{types.NewCurrency64(1)},
				ArbitraryData: []byte("Hello, World!"),
			},
		},
		{
			`{
	"version": 1,
	"data": {
		"coininputs": [
			{
				"parentid": "1100000000000000000000000000000000000000000000000000000000000011",
				"fulfillment": {
					"type": 1,
					"data": {
						"publickey": "ed25519:def123def123def123def123def123def123def123def123def123def123def1",
						"signature": "ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef"
					}
				}
			}
		],
		"coinoutputs": [
			{
				"value": "9",
				"condition": {
					"type": 3,
					"data": {
						"locktime": 42,
						"condition": {}
					}
				}
			}
		],
		"minerfees": [
			"3"
		]
	}
}`,
			types.Transaction{
				Version: types.TransactionVersionOne,
				CoinInputs: []types.CoinInput{
					{
						ParentID: types.CoinOutputID(hs("1100000000000000000000000000000000000000000000000000000000000011")),
						Fulfillment: types.NewFulfillment(&types.SingleSignatureFulfillment{
							PublicKey: types.PublicKey{
								Algorithm: types.SignatureAlgoEd25519,
								Key:       hbs("def123def123def123def123def123def123def123def123def123def123def1"),
							},
							Signature: hbs("ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef"),
						}),
					},
				},
				CoinOutputs: []types.CoinOutput{
					{
						Value: types.NewCurrency64(9),
						Condition: types.NewCondition(&types.TimeLockCondition{
							LockTime:  42,
							Condition: &types.NilCondition{},
						}),
					},
				},
				MinerFees: []types.Currency{types.NewCurrency64(3)},
			},
		},
		{
			`{
	"version": 1,
	"data": {
		"coininputs": [
			{
				"parentid": "1100000000000000000000000000000000000000000000000000000000000011",
				"fulfillment": {
					"type": 1,
					"data": {
						"publickey": "ed25519:def123def123def123def123def123def123def123def123def123def123def1",
						"signature": "ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef"
					}
				}
			}
		],
		"coinoutputs": [
			{
				"value": "9",
				"condition": {
					"type": 3,
					"data": {
						"locktime": 42,
						"condition": {
							"type": 1,
							"data": {
								"unlockhash": "01e89843e4b8231a01ba18b254d530110364432aafab8206bea72e5a20eaa55f70b1ccc65e2105"
							}
						}
					}
				}
			}
		],
		"minerfees": [
			"3"
		]
	}
}`,
			types.Transaction{
				Version: types.TransactionVersionOne,
				CoinInputs: []types.CoinInput{
					{
						ParentID: types.CoinOutputID(hs("1100000000000000000000000000000000000000000000000000000000000011")),
						Fulfillment: types.NewFulfillment(&types.SingleSignatureFulfillment{
							PublicKey: types.PublicKey{
								Algorithm: types.SignatureAlgoEd25519,
								Key:       hbs("def123def123def123def123def123def123def123def123def123def123def1"),
							},
							Signature: hbs("ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef"),
						}),
					},
				},
				CoinOutputs: []types.CoinOutput{
					{
						Value: types.NewCurrency64(9),
						Condition: types.NewCondition(&types.TimeLockCondition{
							LockTime: 42,
							Condition: &types.UnlockHashCondition{
								TargetUnlockHash: types.UnlockHash{
									Type: types.UnlockTypePubKey,
									Hash: hs("e89843e4b8231a01ba18b254d530110364432aafab8206bea72e5a20eaa55f70"),
								},
							},
						}),
					},
				},
				MinerFees: []types.Currency{types.NewCurrency64(3)},
			},
		},
		{
			`{
	"version": 1,
	"data": {
		"coininputs": [
			{
				"parentid": "1100000000000000000000000000000000000000000000000000000000000011",
				"fulfillment": {
					"type": 1,
					"data": {
						"publickey": "ed25519:def123def123def123def123def123def123def123def123def123def123def1",
						"signature": "ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef"
					}
				}
			}
		],
		"coinoutputs": [
			{
				"value": "9",
				"condition": {
					"type": 4,
					"data": {
						"unlockhashes": [
							"01e89843e4b8231a01ba18b254d530110364432aafab8206bea72e5a20eaa55f70b1ccc65e2105",
							"01a6a6c5584b2bfbd08738996cd7930831f958b9a5ed1595525236e861c1a0dc353bdcf54be7d8"
						],
						"minimumsignaturecount": 2
					}
				}
			}
		],
		"minerfees": [
			"3"
		]
	}
}`,
			types.Transaction{
				Version: types.TransactionVersionOne,
				CoinInputs: []types.CoinInput{
					{
						ParentID: types.CoinOutputID(hs("1100000000000000000000000000000000000000000000000000000000000011")),
						Fulfillment: types.NewFulfillment(&types.SingleSignatureFulfillment{
							PublicKey: types.PublicKey{
								Algorithm: types.SignatureAlgoEd25519,
								Key:       hbs("def123def123def123def123def123def123def123def123def123def123def1"),
							},
							Signature: hbs("ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef"),
						}),
					},
				},
				CoinOutputs: []types.CoinOutput{
					{
						Value: types.NewCurrency64(9),
						Condition: types.NewCondition(&types.MultiSignatureCondition{
							MinimumSignatureCount: 2,
							UnlockHashes: types.UnlockHashSlice{
								types.UnlockHash{
									Type: types.UnlockTypePubKey,
									Hash: hs("e89843e4b8231a01ba18b254d530110364432aafab8206bea72e5a20eaa55f70"),
								},
								types.UnlockHash{
									Type: types.UnlockTypePubKey,
									Hash: hs("a6a6c5584b2bfbd08738996cd7930831f958b9a5ed1595525236e861c1a0dc35"),
								},
							},
						}),
					},
				},
				MinerFees: []types.Currency{types.NewCurrency64(3)},
			},
		},
		{
			`{
	"version": 1,
	"data": {
		"coininputs": [
			{
				"parentid": "1100000000000000000000000000000000000000000000000000000000000011",
				"fulfillment": {
					"type": 3,
					"data": {
						"pairs": [
							{
								"publickey": "ed25519:def123def123def123def123def123def123def123def123def123def123def1",
								"signature": "ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef"
							},
							{
								"publickey": "ed25519:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
								"signature": "abcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefab"
							}
						]
					}
				}
			}
		],
		"coinoutputs": [
			{
				"value": "9",
				"condition": {
					"type": 3,
					"data": {
						"locktime": 42,
						"condition": {
							"type": 1,
							"data": {
								"unlockhash": "01e89843e4b8231a01ba18b254d530110364432aafab8206bea72e5a20eaa55f70b1ccc65e2105"
							}
						}
					}
				}
			}
		],
		"minerfees": [
			"3"
		]
	}
}`,
			types.Transaction{
				Version: types.TransactionVersionOne,
				CoinInputs: []types.CoinInput{
					{
						ParentID: types.CoinOutputID(hs("1100000000000000000000000000000000000000000000000000000000000011")),
						Fulfillment: types.NewFulfillment(&types.MultiSignatureFulfillment{
							Pairs: []types.PublicKeySignaturePair{
								{
									PublicKey: types.PublicKey{
										Algorithm: types.SignatureAlgoEd25519,
										Key:       hbs("def123def123def123def123def123def123def123def123def123def123def1"),
									},
									Signature: hbs("ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef"),
								},
								{
									PublicKey: types.PublicKey{
										Algorithm: types.SignatureAlgoEd25519,
										Key:       hbs("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"),
									},
									Signature: hbs("abcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefab"),
								},
							},
						}),
					},
				},
				CoinOutputs: []types.CoinOutput{
					{
						Value: types.NewCurrency64(9),
						Condition: types.NewCondition(&types.TimeLockCondition{
							LockTime: 42,
							Condition: &types.UnlockHashCondition{
								TargetUnlockHash: types.UnlockHash{
									Type: types.UnlockTypePubKey,
									Hash: hs("e89843e4b8231a01ba18b254d530110364432aafab8206bea72e5a20eaa55f70"),
								},
							},
						}),
					},
				},
				MinerFees: []types.Currency{types.NewCurrency64(3)},
			},
		},
	}

	for idx, example := range examples {
		var tx types.Transaction
		err := json.Unmarshal([]byte(example.JSONEncoded), &tx)
		if err != nil {
			t.Error(idx, err)
			continue
		}
		if !reflect.DeepEqual(example.ExpectedTx, tx) {
			t.Errorf("#%d: %v != %v", idx, example.ExpectedTx, tx)
			// go through each input/output to compare
			for cidx, ci := range example.ExpectedTx.CoinInputs {
				t1 := fmt.Sprintf("%T", ci.Fulfillment.Fulfillment)
				t2 := fmt.Sprintf("%T", tx.CoinInputs[cidx].Fulfillment.Fulfillment)
				if t1 != t2 {
					t.Error(idx, "coin input #", cidx, ":", t1, "!=", t2)
				}
			}
			for codx, co := range example.ExpectedTx.CoinOutputs {
				t1 := fmt.Sprintf("%T", co.Condition.Condition)
				t2 := fmt.Sprintf("%T", tx.CoinOutputs[codx].Condition.Condition)
				if t1 != t2 {
					t.Error(idx, "coin output #", codx, ":", t1, "!=", t2)
				}
			}
			for bsidx, bsi := range example.ExpectedTx.BlockStakeInputs {
				t1 := fmt.Sprintf("%T", bsi.Fulfillment.Fulfillment)
				t2 := fmt.Sprintf("%T", tx.BlockStakeInputs[bsidx].Fulfillment.Fulfillment)
				if t1 != t2 {
					t.Error(idx, "coin input #", bsidx, ":", t1, "!=", t2)
				}
			}
			for bsodx, bso := range example.ExpectedTx.BlockStakeOutputs {
				t1 := fmt.Sprintf("%T", bso.Condition.Condition)
				t2 := fmt.Sprintf("%T", tx.BlockStakeOutputs[bsodx].Condition.Condition)
				if t1 != t2 {
					t.Error(idx, "coin output #", bsodx, ":", t1, "!=", t2)
				}
			}
		}
		b, err := json.Marshal(tx)
		if err != nil {
			t.Error(idx, err)
		}
		jsonEncoded := string(b)
		expectedJSONEncoded := strings.NewReplacer(" ", "", "\t", "", "\n", "").Replace(example.JSONEncoded)
		if expectedJSONEncoded != jsonEncoded {
			t.Errorf("#%d: %v != %v", idx, expectedJSONEncoded, jsonEncoded)
		}
	}
}
