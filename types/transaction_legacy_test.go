package types

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/threefoldtech/rivine/crypto"
	"github.com/threefoldtech/rivine/pkg/encoding/siabin"
)

func TestLegacyTransactionInputLockProxyBinarySiaEncoding(t *testing.T) {
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

func TestLegacyTransactionInputLockProxyBinaryRivineEncoding(t *testing.T) {
	testCases := []string{
		`014201ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff80ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff`,
		`02d4011234567891234567891234567891234567891234567891234567891234567891016363636363636363636363636363636363636363636363636363636363636363bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb07edb85a00000000090201abababababababababababababababababababababababababababababababab80dededededededededededededededededededededededededededededededededededededededededededededededededededededededededededededededededabadabadabadabadabadabadabadabadabadabadabadabadabadabadabadaba`,
	}
	for testIndex, testCase := range testCases {
		binaryInput, err := hex.DecodeString(testCase)
		if err != nil {
			t.Error(testIndex, err)
			continue
		}

		var proxy legacyTransactionInputLockProxy
		err = proxy.UnmarshalRivine(bytes.NewReader(binaryInput))
		if err != nil {
			t.Error(testIndex, err)
			continue
		}
		buf := bytes.NewBuffer(nil)
		err = proxy.MarshalRivine(buf)
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
		`01000000000000002200000000000000000000000000000000000000000000000000000000000022013800000000000000656432353531390000000000000000002000000000000000ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff4000000000000000ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff00000000000000000000000000000000000000000000000001000000000000000100000000000000010000000000000000`,
		`02000000000000002200000000000000000000000000000000000000000000000000000000000022013800000000000000656432353531390000000000000000002000000000000000ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff4000000000000000ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff3300000000000000000000000000000000000000000000000000000000000033026a00000000000000011234567891234567891234567891234567891234567891234567891234567891016363636363636363636363636363636363636363636363636363636363636363bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb07edb85a00000000a000000000000000656432353531390000000000000000002000000000000000abababababababababababababababababababababababababababababababab4000000000000000dededededededededededededededededededededededededededededededededededededededededededededededededededededededededededededededededabadabadabadabadabadabadabadabadabadabadabadabadabadabadabadaba020000000000000001000000000000000201cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc01000000000000000302dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd01000000000000004400000000000000000000000000000000000000000000000000000000000044013800000000000000656432353531390000000000000000002000000000000000eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee4000000000000000eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee010000000000000001000000000000002a01abcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcd010000000000000001000000000000000102000000000000003432`,
	}
	for testIndex, testCase := range testCases {
		binaryInput, err := hex.DecodeString(testCase)
		if err != nil {
			t.Error(testIndex, err)
			continue
		}

		var ltd legacyTransactionData
		err = siabin.Unmarshal(binaryInput, &ltd)
		if err != nil {
			t.Error(testIndex, err)
			continue
		}
		b := siabin.Marshal(ltd)

		output := hex.EncodeToString(b)
		if output != testCase {
			t.Error(testIndex, output, "!=", testCase)
		}
	}
}

func TestLegacyTransactionJSONEncoding(t *testing.T) {
	testCases := []string{
		`{
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
}`, `{
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
}`,
	}
	for testIndex, testCase := range testCases {
		var ltd legacyTransactionData
		err := json.Unmarshal([]byte(testCase), &ltd)
		if err != nil {
			t.Error(testIndex, err)
			continue
		}
		buf := bytes.NewBuffer(nil)
		encoder := json.NewEncoder(buf)
		encoder.SetIndent("", "\t")
		err = encoder.Encode(ltd)
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
		EncodedTransactionData  string
		ExpectedTransactionData TransactionData
	}{
		{
			`01000000000000002200000000000000000000000000000000000000000000000000000000000022013800000000000000656432353531390000000000000000002000000000000000ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff4000000000000000ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff00000000000000000000000000000000000000000000000001000000000000000100000000000000010000000000000000`,
			TransactionData{
				CoinInputs: []CoinInput{
					{
						ParentID: CoinOutputID(hs("2200000000000000000000000000000000000000000000000000000000000022")),
						Fulfillment: NewFulfillment(&SingleSignatureFulfillment{
							PublicKey: PublicKey{
								Algorithm: SignatureAlgoEd25519,
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
			`02000000000000002200000000000000000000000000000000000000000000000000000000000022013800000000000000656432353531390000000000000000002000000000000000ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff4000000000000000ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff3300000000000000000000000000000000000000000000000000000000000033026a00000000000000011234567891234567891234567891234567891234567891234567891234567891016363636363636363636363636363636363636363636363636363636363636363bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb07edb85a00000000a000000000000000656432353531390000000000000000002000000000000000abababababababababababababababababababababababababababababababab4000000000000000dededededededededededededededededededededededededededededededededededededededededededededededededededededededededededededededededabadabadabadabadabadabadabadabadabadabadabadabadabadabadabadaba020000000000000001000000000000000201cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc01000000000000000302dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd01000000000000004400000000000000000000000000000000000000000000000000000000000044013800000000000000656432353531390000000000000000002000000000000000eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee4000000000000000eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee010000000000000001000000000000002a01abcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcd010000000000000001000000000000000102000000000000003432`,
			TransactionData{
				CoinInputs: []CoinInput{
					{
						ParentID: CoinOutputID(hs("2200000000000000000000000000000000000000000000000000000000000022")),
						Fulfillment: NewFulfillment(&SingleSignatureFulfillment{
							PublicKey: PublicKey{
								Algorithm: SignatureAlgoEd25519,
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
							PublicKey: PublicKey{
								Algorithm: SignatureAlgoEd25519,
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
							PublicKey: PublicKey{
								Algorithm: SignatureAlgoEd25519,
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
		binaryInput, err := hex.DecodeString(testCase.EncodedTransactionData)
		if err != nil {
			t.Error(testIndex, err)
			continue
		}

		var ltd legacyTransactionData
		err = siabin.Unmarshal(binaryInput, &ltd)
		if err != nil {
			t.Error(testIndex, err)
			continue
		}
		data := ltd.TransactionData()

		if !reflect.DeepEqual(testCase.ExpectedTransactionData, data) {
			t.Error(testIndex, testCase.ExpectedTransactionData, "!=", data)
		}
	}
}

// TestLegacyTransactionSignatures tests that the signatures,
// generated for input's fulfillments, are 100% compatible
// with the signatures computed in older versions, for
// transaction v0 (legacy) input(s) (locks).
// Added as part of https://github.com/threefoldtech/rivine/issues/312
func TestLegacyTransactionSignatures(t *testing.T) {
	// utility funcs
	hbs := func(str string) []byte { // hexStr -> byte slice
		bs, _ := hex.DecodeString(str)
		return bs
	}
	hs := func(str string) (hash crypto.Hash) { // hbs -> crypto.Hash
		copy(hash[:], hbs(str))
		return
	}

	var entropy [crypto.EntropySize]byte
	b, err := hex.DecodeString("5722127ae85b120d15d5998d7591a603f4cbe2ac47a69b4d16bd7c90e21aeedd")
	if err != nil {
		t.Fatal(err)
	}
	copy(entropy[:], b[:])
	sk, rpk := crypto.GenerateKeyPairDeterministic(entropy)
	pk := Ed25519PublicKey(rpk)

	var secret AtomicSwapSecret
	b, err = hex.DecodeString("8a1ed1737c0a84598bbf429ba351e03b5f4e00d7ab5d8635da832957f14bb272")
	if err != nil {
		t.Fatal(err)
	}
	copy(secret[:], b[:])

	testCases := []struct {
		Transaction                  Transaction
		ExpectedCoinSignatures       []ByteSlice
		ExpectedBlockStakeSignatures []ByteSlice
	}{
		{
			Transaction{
				Version:     0,
				CoinInputs:  nil,
				CoinOutputs: nil,
				BlockStakeInputs: []BlockStakeInput{
					{
						ParentID: BlockStakeOutputID(hs("0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")),
						Fulfillment: UnlockFulfillmentProxy{
							Fulfillment: NewSingleSignatureFulfillment(pk),
						},
					},
				},
				BlockStakeOutputs: []BlockStakeOutput{
					{
						Value: NewCurrency64(42),
						Condition: UnlockConditionProxy{
							Condition: NewUnlockHashCondition(UnlockHash{
								Type: UnlockTypePubKey,
								Hash: hs("a6a6c5584b2bfbd08738996cd7930831f958b9a5ed1595525236e861c1a0dc35"),
							}),
						},
					},
				},
				MinerFees:     nil,
				ArbitraryData: nil,
			},
			nil,
			[]ByteSlice{hbs("73695bac6d27e70f4885c151cf499c8c15a9fb3bef493070d265b015b997a76e7e578f52f25fab0dc3f860f331b735a3abd78fe803e7d21bc8ae0389d9e52606")},
		},
		{
			Transaction{
				Version: 0,
				CoinInputs: []CoinInput{
					{
						ParentID: CoinOutputID(hs("dfd23dfd23dfd23dfd23dfd23dfd23dfd23dfd23dfd23dfd23dfd23dfd23dfde")),
						Fulfillment: UnlockFulfillmentProxy{
							Fulfillment: NewSingleSignatureFulfillment(pk),
						},
					},
				},
				CoinOutputs: []CoinOutput{
					{
						Value: NewCurrency64(42),
						Condition: UnlockConditionProxy{
							Condition: NewUnlockHashCondition(UnlockHash{
								Type: UnlockTypePubKey,
								Hash: hs("b39baa9a58319fa47f78ed542a733a7198d106caeabf0a231b91ea3e4e222ffd"),
							}),
						},
					},
				},
				BlockStakeInputs:  nil,
				BlockStakeOutputs: nil,
				MinerFees:         []Currency{NewCurrency64(1000)},
				ArbitraryData:     nil,
			},
			[]ByteSlice{hbs("e73b4a2d8d6ab62e56b84dc8345e59dbf77a0f406cd31face93bc93757d8277a88431ec7d8e2a1b44c1cc2f8bb45107aabf689464cd26a20fdf43532984de304")},
			nil,
		},
		{
			Transaction{
				Version: 0,
				CoinInputs: []CoinInput{
					{
						ParentID: CoinOutputID(hs("dfd23dfd23dfd23dfd23dfd23dfd23dfd23dfd23dfd23dfd23dfd23dfd23dfde")),
						Fulfillment: UnlockFulfillmentProxy{
							Fulfillment: NewSingleSignatureFulfillment(pk),
						},
					},
				},
				CoinOutputs: []CoinOutput{
					{
						Value: NewCurrency64(42),
						Condition: UnlockConditionProxy{
							Condition: NewUnlockHashCondition(UnlockHash{
								Type: UnlockTypePubKey,
								Hash: hs("b39baa9a58319fa47f78ed542a733a7198d106caeabf0a231b91ea3e4e222ffd"),
							}),
						},
					},
				},
				BlockStakeInputs:  nil,
				BlockStakeOutputs: nil,
				MinerFees:         []Currency{NewCurrency64(1000)},
				ArbitraryData:     []byte("Hello, World!"),
			},
			[]ByteSlice{hbs("93655e18ba4d1e582983eba728e181e70db62edce8e48cd6d42611847411961705e816700233b70971314cf99042f6fc0e00c0f5111f9ef900975d11b9b6e603")},
			nil,
		},
		{
			Transaction{
				Version: 0,
				CoinInputs: []CoinInput{
					{
						ParentID: CoinOutputID(hs("0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")),
						Fulfillment: UnlockFulfillmentProxy{
							Fulfillment: NewSingleSignatureFulfillment(pk),
						},
					},
				},
				CoinOutputs: []CoinOutput{
					{
						Value: NewCurrency64(42),
						Condition: UnlockConditionProxy{
							Condition: NewUnlockHashCondition(UnlockHash{
								Type: UnlockTypePubKey,
								Hash: hs("def123def123def123def123def123def123def123def123def123def123def1"),
							}),
						},
					},
				},
				BlockStakeInputs: []BlockStakeInput{
					{
						ParentID: BlockStakeOutputID(hs("abcdef012345abcdef012345abcdef012345abcdef012345abcdef012345abcd")),
						Fulfillment: UnlockFulfillmentProxy{
							Fulfillment: NewSingleSignatureFulfillment(pk),
						},
					},
				},
				BlockStakeOutputs: []BlockStakeOutput{
					{
						Value: NewCurrency64(42),
						Condition: UnlockConditionProxy{
							Condition: NewUnlockHashCondition(UnlockHash{
								Type: UnlockTypePubKey,
								Hash: hs("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"),
							}),
						},
					},
				},
				MinerFees:     []Currency{NewCurrency64(9129)},
				ArbitraryData: nil,
			},
			[]ByteSlice{hbs("061a85329cb5a6de6b6df4dcb6886682353d04f94039237232f1c3c50398631a0c06e96d1f529ae0ea204d0b5f3d3c3b615958b79417fb40a2557d38d8b97501")},
			[]ByteSlice{hbs("061a85329cb5a6de6b6df4dcb6886682353d04f94039237232f1c3c50398631a0c06e96d1f529ae0ea204d0b5f3d3c3b615958b79417fb40a2557d38d8b97501")},
		},
		{
			Transaction{
				Version: 0,
				CoinInputs: []CoinInput{
					{
						ParentID: CoinOutputID(hs("dfd23dfd23dfd23dfd23dfd23dfd23dfd23dfd23dfd23dfd23dfd23dfd23dfde")),
						Fulfillment: UnlockFulfillmentProxy{
							Fulfillment: &LegacyAtomicSwapFulfillment{
								Sender: UnlockHash{
									Type: UnlockTypePubKey,
									Hash: hs("a6a6c5584b2bfbd08738996cd7930831f958b9a5ed1595525236e861c1a0dc35"),
								},
								Receiver:     NewPubKeyUnlockHash(pk),
								HashedSecret: NewAtomicSwapHashedSecret(secret),
								TimeLock:     1525549854,
								PublicKey:    pk,
								Secret:       secret,
							},
						},
					},
				},
				CoinOutputs: []CoinOutput{
					{
						Value: NewCurrency64(42),
						Condition: UnlockConditionProxy{
							Condition: NewUnlockHashCondition(UnlockHash{
								Type: UnlockTypeAtomicSwap,
								Hash: hs("b39baa9a58319fa47f78ed542a733a7198d106caeabf0a231b91ea3e4e222ffd"),
							}),
						},
					},
				},
				BlockStakeInputs:  nil,
				BlockStakeOutputs: nil,
				MinerFees:         []Currency{NewCurrency64(500)},
				ArbitraryData:     nil,
			},
			[]ByteSlice{hbs("df316d4f33ee4e1a35db6ac8129b103109b558016c267546d710d6d15ba578ef20c5bcecc7c8b9d010cc78a5843e28f5cc3433f7bbe027d8313cce7d51413d05")},
			nil,
		},
		{
			Transaction{
				Version: 0,
				CoinInputs: []CoinInput{
					{
						ParentID: CoinOutputID(hs("dfd23dfd23dfd23dfd23dfd23dfd23dfd23dfd23dfd23dfd23dfd23dfd23dfde")),
						Fulfillment: UnlockFulfillmentProxy{
							Fulfillment: &LegacyAtomicSwapFulfillment{
								Sender: UnlockHash{
									Type: UnlockTypePubKey,
									Hash: hs("a6a6c5584b2bfbd08738996cd7930831f958b9a5ed1595525236e861c1a0dc35"),
								},
								Receiver:     NewPubKeyUnlockHash(pk),
								HashedSecret: NewAtomicSwapHashedSecret(secret),
								TimeLock:     1525549854,
								PublicKey:    pk,
								Secret:       secret,
							},
						},
					},
				},
				CoinOutputs: []CoinOutput{
					{
						Value: NewCurrency64(42),
						Condition: UnlockConditionProxy{
							Condition: NewUnlockHashCondition(UnlockHash{
								Type: UnlockTypeAtomicSwap,
								Hash: hs("b39baa9a58319fa47f78ed542a733a7198d106caeabf0a231b91ea3e4e222ffd"),
							}),
						},
					},
				},
				BlockStakeInputs: []BlockStakeInput{
					{
						ParentID: BlockStakeOutputID(hs("0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")),
						Fulfillment: UnlockFulfillmentProxy{
							Fulfillment: NewSingleSignatureFulfillment(pk),
						},
					},
				},
				BlockStakeOutputs: []BlockStakeOutput{
					{
						Value: NewCurrency64(42),
						Condition: UnlockConditionProxy{
							Condition: NewUnlockHashCondition(UnlockHash{
								Type: UnlockTypePubKey,
								Hash: hs("a6a6c5584b2bfbd08738996cd7930831f958b9a5ed1595525236e861c1a0dc353bdcf54be7d8"),
							}),
						},
					},
				},
				MinerFees:     []Currency{NewCurrency64(1111)},
				ArbitraryData: nil,
			},
			[]ByteSlice{hbs("a373bafeaa8fe3ac8514f4c250d33fab17c12c577cf658eef45703f58ec5c57fe191580e3f3d2575e3a01fffc738a8e46e1e30d24c4422f1c92c46aaafb3bb01")},
			[]ByteSlice{hbs("762eacab9b953d69042a8e45e28c80d6d0600c78b577e0e84e099a618ea20d6f5dd211457d5967feb979e54b44764fdfb3831c9fcea2a0de8a90bf454fe4250a")},
		},
	}
	for idx, testCase := range testCases {
		for i, ci := range testCase.Transaction.CoinInputs {
			err := ci.Fulfillment.Sign(FulfillmentSignContext{
				ExtraObjects: []interface{}{uint64(i)},
				Transaction:  testCase.Transaction,
				Key:          sk,
			})
			if err != nil {
				t.Error(idx, i, err)
			}
			switch tf := ci.Fulfillment.Fulfillment.(type) {
			case *SingleSignatureFulfillment:
				if bytes.Compare(tf.Signature[:], testCase.ExpectedCoinSignatures[i][:]) != 0 {
					t.Error(idx, "coin", i,
						hex.EncodeToString(tf.Signature[:]), "!=",
						hex.EncodeToString(testCase.ExpectedCoinSignatures[i][:]))
				}
			case *LegacyAtomicSwapFulfillment:
				if bytes.Compare(tf.Signature[:], testCase.ExpectedCoinSignatures[i][:]) != 0 {
					t.Error(idx, "coin", i,
						hex.EncodeToString(tf.Signature[:]), "!=",
						hex.EncodeToString(testCase.ExpectedCoinSignatures[i][:]))
				}
			}
		}
		for i, bsi := range testCase.Transaction.BlockStakeInputs {
			err := bsi.Fulfillment.Sign(FulfillmentSignContext{
				ExtraObjects: []interface{}{uint64(i)},
				Transaction:  testCase.Transaction,
				Key:          sk,
			})
			if err != nil {
				t.Error(idx, i, err)
			}
			switch tf := bsi.Fulfillment.Fulfillment.(type) {
			case *SingleSignatureFulfillment:
				if bytes.Compare(tf.Signature[:], testCase.ExpectedBlockStakeSignatures[i][:]) != 0 {
					t.Error(idx, "block stake", i,
						hex.EncodeToString(tf.Signature[:]), "!=",
						hex.EncodeToString(testCase.ExpectedBlockStakeSignatures[i][:]))
				}
			case *LegacyAtomicSwapFulfillment:
				if bytes.Compare(tf.Signature[:], testCase.ExpectedBlockStakeSignatures[i][:]) != 0 {
					t.Error(idx, "block stake", i,
						hex.EncodeToString(tf.Signature[:]), "!=",
						hex.EncodeToString(testCase.ExpectedBlockStakeSignatures[i][:]))
				}
			}
		}
	}
}

func TestLegacyTransactionDataBiDirectional(t *testing.T) {
	testCases := []string{
		`01000000000000002200000000000000000000000000000000000000000000000000000000000022013800000000000000656432353531390000000000000000002000000000000000ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff4000000000000000ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff00000000000000000000000000000000000000000000000001000000000000000100000000000000010000000000000000`,
		`02000000000000002200000000000000000000000000000000000000000000000000000000000022013800000000000000656432353531390000000000000000002000000000000000ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff4000000000000000ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff3300000000000000000000000000000000000000000000000000000000000033026a00000000000000011234567891234567891234567891234567891234567891234567891234567891016363636363636363636363636363636363636363636363636363636363636363bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb07edb85a00000000a000000000000000656432353531390000000000000000002000000000000000abababababababababababababababababababababababababababababababab4000000000000000dededededededededededededededededededededededededededededededededededededededededededededededededededededededededededededededededabadabadabadabadabadabadabadabadabadabadabadabadabadabadabadaba020000000000000001000000000000000201cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc01000000000000000302dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd01000000000000004400000000000000000000000000000000000000000000000000000000000044013800000000000000656432353531390000000000000000002000000000000000eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee4000000000000000eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee010000000000000001000000000000002a01abcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcd010000000000000001000000000000000102000000000000003432`,
	}
	for testIndex, testCase := range testCases {
		binaryInput, err := hex.DecodeString(testCase)
		if err != nil {
			t.Error(testIndex, err)
			continue
		}

		var ltd legacyTransactionData
		err = siabin.Unmarshal(binaryInput, &ltd)
		if err != nil {
			t.Error(testIndex, err)
			continue
		}

		oltd, err := newLegacyTransactionData(ltd.TransactionData())
		if err != nil {
			t.Error(testIndex, err)
			continue
		}

		if !reflect.DeepEqual(ltd, oltd) {
			t.Error(testIndex, ltd, "!=", oltd)
		}
	}
}
