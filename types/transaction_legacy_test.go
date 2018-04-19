package types

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/rivine/rivine/crypto"
	"github.com/rivine/rivine/encoding"
)

func TestLegacyTransactionInputLockProxyBinaryEncoding(t *testing.T) {
	testCases := []string{
		`013800000000000000656432353531390000000000000000002000000000000000ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff4000000000000000ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff`,
		`026a00000000000000011234567891234567891234567891234567891234567891234567891234567891016363636363636363636363636363636363636363636363636363636363636363bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb07edb85a00000000a000000000000000656432353531390000000000000000002000000000000000abababababababababababababababababababababababababababababababab4000000000000000dededededededededededededededededededededededededededededededededededededededededededededededededededededededededededededededededabadabadabadabadabadabadabadabadabadabadabadabadabadabadabadaba`,
	}
	for testIndex, testCase := range testCases {
		binaryInput, err := hex.DecodeString(testCase)
		if err != nil {
			t.Error(testIndex, err)
			continue
		}

		var proxy legacyTransactionInputLockProxy
		err = proxy.UnmarshalSia(bytes.NewReader(binaryInput))
		if err != nil {
			t.Error(testIndex, err)
			continue
		}
		buf := bytes.NewBuffer(nil)
		err = proxy.MarshalSia(buf)
		if err != nil {
			t.Error(testIndex, err)
			continue
		}

		output := hex.EncodeToString(buf.Bytes())
		if output != testCase {
			t.Error(testIndex, output, "!=", testCase)
		}
	}
}

func TestLegacyTransactionInputLockProxyJSONEncoding(t *testing.T) {
	testCases := []string{
		`{
	"type": 1,
	"condition": {
		"publickey": "ed25519:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"
	},
	"fulfillment": {
		"signature": "abcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefab"
	}
}`, `{
	"type": 2,
	"condition": {
		"sender": "01654f96b317efe5fd6cd8ba1a394dce7b6ebe8c9621d6c44cbe3c8f1b58ce632a3216de71b23b",
		"receiver": "01e89843e4b8231a01ba18b254d530110364432aafab8206bea72e5a20eaa55f70b1ccc65e2105",
		"hashedsecret": "abc543defabc543defabc543defabc543defabc543defabc543defabc543defa",
		"timelock": 1522068743
	},
	"fulfillment": {
		"publickey": "ed25519:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
		"signature": "abcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefab",
		"secret": "def789def789def789def789def789dedef789def789def789def789def789de"
	}
}`,
	}
	for testIndex, testCase := range testCases {
		var proxy legacyTransactionInputLockProxy
		err := proxy.UnmarshalJSON([]byte(testCase))
		if err != nil {
			t.Error(testIndex, err)
			continue
		}
		buf := bytes.NewBuffer(nil)
		encoder := json.NewEncoder(buf)
		encoder.SetIndent("", "\t")
		err = encoder.Encode(proxy)
		if err != nil {
			t.Error(testIndex, err)
			continue
		}

		output := strings.TrimSpace(string(buf.Bytes()))
		if output != testCase {
			t.Error(testIndex, "'", output, "'!='", testCase, "'")
		}
	}
}

func TestLegacyTransactionBinaryEncoding(t *testing.T) {
	testCases := []string{
		`0001000000000000002200000000000000000000000000000000000000000000000000000000000022013800000000000000656432353531390000000000000000002000000000000000ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff4000000000000000ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff00000000000000000000000000000000000000000000000001000000000000000100000000000000010000000000000000`,
		`0002000000000000002200000000000000000000000000000000000000000000000000000000000022013800000000000000656432353531390000000000000000002000000000000000ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff4000000000000000ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff3300000000000000000000000000000000000000000000000000000000000033026a00000000000000011234567891234567891234567891234567891234567891234567891234567891016363636363636363636363636363636363636363636363636363636363636363bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb07edb85a00000000a000000000000000656432353531390000000000000000002000000000000000abababababababababababababababababababababababababababababababab4000000000000000dededededededededededededededededededededededededededededededededededededededededededededededededededededededededededededededededabadabadabadabadabadabadabadabadabadabadabadabadabadabadabadaba020000000000000001000000000000000201cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc01000000000000000302dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd01000000000000004400000000000000000000000000000000000000000000000000000000000044013800000000000000656432353531390000000000000000002000000000000000eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee4000000000000000eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee010000000000000001000000000000002a01abcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcd010000000000000001000000000000000102000000000000003432`,
	}
	for testIndex, testCase := range testCases {
		binaryInput, err := hex.DecodeString(testCase)
		if err != nil {
			t.Error(testIndex, err)
			continue
		}

		var txn legacyTransaction
		err = encoding.Unmarshal(binaryInput, &txn)
		if err != nil {
			t.Error(testIndex, err)
			continue
		}
		b := encoding.Marshal(txn)

		output := hex.EncodeToString(b)
		if output != testCase {
			t.Error(testIndex, output, "!=", testCase)
		}
	}
}

func TestLegacyTransactionJSONEncoding(t *testing.T) {
	testCases := []string{
		`{
	"version": 0,
	"data": {
		"coininputs": [
			{
				"parentid": "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
				"unlocker": {
					"type": 1,
					"condition": {
						"publickey": "ed25519:def123def123def123def123def123def123def123def123def123def123def1"
					},
					"fulfillment": {
						"signature": "ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef"
					}
				}
			}
		],
		"minerfees": [
			"1"
		],
		"arbitrarydata": "SGVsbG8sIFdvcmxkIQ=="
	}
}`, `{
	"version": 0,
	"data": {
		"coininputs": [
			{
				"parentid": "abcdef012345abcdef012345abcdef012345abcdef012345abcdef012345abcd",
				"unlocker": {
					"type": 1,
					"condition": {
						"publickey": "ed25519:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"
					},
					"fulfillment": {
						"signature": "abcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefab"
					}
				}
			},
			{
				"parentid": "012345defabc012345defabc012345defabc012345defabc012345defabc0123",
				"unlocker": {
					"type": 2,
					"condition": {
						"sender": "01654f96b317efe5fd6cd8ba1a394dce7b6ebe8c9621d6c44cbe3c8f1b58ce632a3216de71b23b",
						"receiver": "01e89843e4b8231a01ba18b254d530110364432aafab8206bea72e5a20eaa55f70b1ccc65e2105",
						"hashedsecret": "abc543defabc543defabc543defabc543defabc543defabc543defabc543defa",
						"timelock": 1522068743
					},
					"fulfillment": {
						"publickey": "ed25519:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
						"signature": "abcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefab",
						"secret": "def789def789def789def789def789dedef789def789def789def789def789de"
					}
				}
			}
		],
		"coinoutputs": [
			{
				"value": "3",
				"unlockhash": "0142e9458e348598111b0bc19bda18e45835605db9f4620616d752220ae8605ce0df815fd7570e"
			},
			{
				"value": "5",
				"unlockhash": "01a6a6c5584b2bfbd08738996cd7930831f958b9a5ed1595525236e861c1a0dc353bdcf54be7d8"
			},
			{
				"value": "8",
				"unlockhash": "02a24c97c80eeac111aa4bcbb0ac8ffc364fa9b22da10d3054778d2332f68b365e5e5af8e71541"
			}
		],
		"blockstakeinputs": [
			{
				"parentid": "dfd23dfd23dfd23dfd23dfd23dfd23dfd23dfd23dfd23dfd23dfd23dfd23dfde",
				"unlocker": {
					"type": 1,
					"condition": {
						"publickey": "ed25519:ef1234ef1234ef1234ef1234ef1234ef1234ef1234ef1234ef1234ef1234ef12"
					},
					"fulfillment": {
						"signature": "01234def01234def01234def01234def01234def01234def01234def01234def01234def01234def01234def01234def01234def01234def01234def01234def"
					}
				}
			}
		],
		"blockstakeoutputs": [
			{
				"value": "4",
				"unlockhash": "6453402d094ed0f336950c4be0feec37167aaaaf8b974d265900e49ab22773584cfe96393b1360"
			},
			{
				"value": "2",
				"unlockhash": "2ab39baa9a58319fa47f78ed542a733a7198d106caeabf0a231b91ea3e4e222ffd8b27c861beff"
			}
		],
		"minerfees": [
			"1",
			"2",
			"3"
		],
		"arbitrarydata": "ZGF0YQ=="
	}
}`,
	}
	for testIndex, testCase := range testCases {
		var txn legacyTransaction
		err := json.Unmarshal([]byte(testCase), &txn)
		if err != nil {
			t.Error(testIndex, err)
			continue
		}
		buf := bytes.NewBuffer(nil)
		encoder := json.NewEncoder(buf)
		encoder.SetIndent("", "\t")
		err = encoder.Encode(txn)
		if err != nil {
			t.Error(testIndex, err)
			continue
		}

		output := strings.TrimSpace(string(buf.Bytes()))
		if output != testCase {
			t.Error(testIndex, "'", output, "'!='", testCase, "'")
		}
	}
}

func TestLegacyTransactionToTransaction(t *testing.T) {
	// utility funcs
	hbs := func(str string) []byte { // hexStr -> byte slice
		bs, _ := hex.DecodeString(str)
		return bs
	}
	hs := func(str string) (hash crypto.Hash) { // hbs -> crypto.Hash
		copy(hash[:], hbs(str))
		return
	}

	testCases := []struct {
		EncodedTransaction  string
		ExpectedTransaction Transaction
	}{
		{
			`0001000000000000002200000000000000000000000000000000000000000000000000000000000022013800000000000000656432353531390000000000000000002000000000000000ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff4000000000000000ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff00000000000000000000000000000000000000000000000001000000000000000100000000000000010000000000000000`,
			Transaction{
				Version: TransactionVersionZero,
				CoinInputs: []CoinInput{
					{
						ParentID: CoinOutputID(hs("2200000000000000000000000000000000000000000000000000000000000022")),
						Fulfillment: NewFulfillment(&SingleSignatureFulfillment{
							PublicKey: SiaPublicKey{
								Algorithm: SignatureEd25519,
								Key:       hbs("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"),
							},
							Signature: hbs("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"),
						}),
					},
				},
				MinerFees: []Currency{NewCurrency64(1)},
			},
		},
		{
			`0002000000000000002200000000000000000000000000000000000000000000000000000000000022013800000000000000656432353531390000000000000000002000000000000000ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff4000000000000000ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff3300000000000000000000000000000000000000000000000000000000000033026a00000000000000011234567891234567891234567891234567891234567891234567891234567891016363636363636363636363636363636363636363636363636363636363636363bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb07edb85a00000000a000000000000000656432353531390000000000000000002000000000000000abababababababababababababababababababababababababababababababab4000000000000000dededededededededededededededededededededededededededededededededededededededededededededededededededededededededededededededededabadabadabadabadabadabadabadabadabadabadabadabadabadabadabadaba020000000000000001000000000000000201cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc01000000000000000302dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd01000000000000004400000000000000000000000000000000000000000000000000000000000044013800000000000000656432353531390000000000000000002000000000000000eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee4000000000000000eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee010000000000000001000000000000002a01abcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcd010000000000000001000000000000000102000000000000003432`,
			Transaction{
				Version: TransactionVersionZero,
				CoinInputs: []CoinInput{
					{
						ParentID: CoinOutputID(hs("2200000000000000000000000000000000000000000000000000000000000022")),
						Fulfillment: NewFulfillment(&SingleSignatureFulfillment{
							PublicKey: SiaPublicKey{
								Algorithm: SignatureEd25519,
								Key:       hbs("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"),
							},
							Signature: hbs("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"),
						}),
					},
					{
						ParentID: CoinOutputID(hs("3300000000000000000000000000000000000000000000000000000000000033")),
						Fulfillment: NewFulfillment(&LegacyAtomicSwapFulfillment{
							Sender: UnlockHash{
								Type: UnlockTypePubKey,
								Hash: hs("1234567891234567891234567891234567891234567891234567891234567891"),
							},
							Receiver: UnlockHash{
								Type: UnlockTypePubKey,
								Hash: hs("6363636363636363636363636363636363636363636363636363636363636363"),
							},
							HashedSecret: AtomicSwapHashedSecret(hs("bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb")),
							TimeLock:     1522068743,
							PublicKey: SiaPublicKey{
								Algorithm: SignatureEd25519,
								Key:       hbs("abababababababababababababababababababababababababababababababab"),
							},
							Signature: hbs("dededededededededededededededededededededededededededededededededededededededededededededededededededededededededededededededede"),
							Secret:    AtomicSwapSecret(hs("dabadabadabadabadabadabadabadabadabadabadabadabadabadabadabadaba")),
						}),
					},
				},
				CoinOutputs: []CoinOutput{
					{
						Value: NewCurrency64(2),
						Condition: NewCondition(NewUnlockHashCondition(UnlockHash{
							Type: UnlockTypePubKey,
							Hash: hs("cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"),
						})),
					},
					{
						Value: NewCurrency64(3),
						Condition: NewCondition(NewUnlockHashCondition(UnlockHash{
							Type: UnlockTypeAtomicSwap,
							Hash: hs("dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd"),
						})),
					},
				},
				BlockStakeInputs: []BlockStakeInput{
					{
						ParentID: BlockStakeOutputID(hs("4400000000000000000000000000000000000000000000000000000000000044")),
						Fulfillment: NewFulfillment(&SingleSignatureFulfillment{
							PublicKey: SiaPublicKey{
								Algorithm: SignatureEd25519,
								Key:       hbs("eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee"),
							},
							Signature: hbs("eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee"),
						}),
					},
				},
				BlockStakeOutputs: []BlockStakeOutput{
					{
						Value: NewCurrency64(42),
						Condition: NewCondition(NewUnlockHashCondition(UnlockHash{
							Type: UnlockTypePubKey,
							Hash: hs("abcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcd"),
						})),
					},
				},
				MinerFees:     []Currency{NewCurrency64(1)},
				ArbitraryData: []byte("42"),
			},
		},
	}
	for testIndex, testCase := range testCases {
		binaryInput, err := hex.DecodeString(testCase.EncodedTransaction)
		if err != nil {
			t.Error(testIndex, err)
			continue
		}

		var lt legacyTransaction
		err = encoding.Unmarshal(binaryInput, &lt)
		if err != nil {
			t.Error(testIndex, err)
			continue
		}

		txn := lt.Transaction()
		if !reflect.DeepEqual(testCase.ExpectedTransaction, txn) {
			t.Error(testIndex, testCase.ExpectedTransaction, "!=", txn)
		}
	}
}

func TestLegacyTransactionBiDirectional(t *testing.T) {
	testCases := []string{
		`0001000000000000002200000000000000000000000000000000000000000000000000000000000022013800000000000000656432353531390000000000000000002000000000000000ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff4000000000000000ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff00000000000000000000000000000000000000000000000001000000000000000100000000000000010000000000000000`,
		`0002000000000000002200000000000000000000000000000000000000000000000000000000000022013800000000000000656432353531390000000000000000002000000000000000ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff4000000000000000ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff3300000000000000000000000000000000000000000000000000000000000033026a00000000000000011234567891234567891234567891234567891234567891234567891234567891016363636363636363636363636363636363636363636363636363636363636363bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb07edb85a00000000a000000000000000656432353531390000000000000000002000000000000000abababababababababababababababababababababababababababababababab4000000000000000dededededededededededededededededededededededededededededededededededededededededededededededededededededededededededededededededabadabadabadabadabadabadabadabadabadabadabadabadabadabadabadaba020000000000000001000000000000000201cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc01000000000000000302dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd01000000000000004400000000000000000000000000000000000000000000000000000000000044013800000000000000656432353531390000000000000000002000000000000000eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee4000000000000000eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee010000000000000001000000000000002a01abcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcd010000000000000001000000000000000102000000000000003432`,
	}
	for testIndex, testCase := range testCases {
		binaryInput, err := hex.DecodeString(testCase)
		if err != nil {
			t.Error(testIndex, err)
			continue
		}

		var txn legacyTransaction
		err = encoding.Unmarshal(binaryInput, &txn)
		if err != nil {
			t.Error(testIndex, err)
			continue
		}

		otxn, err := newLegacyTransaction(txn.Transaction())
		if err != nil {
			t.Error(testIndex, err)
			continue
		}

		if !reflect.DeepEqual(txn, otxn) {
			t.Error(testIndex, txn, "!=", otxn)
		}
	}
}
