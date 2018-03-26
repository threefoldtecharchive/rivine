package types

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"reflect"
	"strings"
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

// TestSpecifierMarshaling tests the marshaling methods of the specifier
// type.
func TestSpecifierMarshaling(t *testing.T) {
	s1 := SpecifierBlockStakeOutput
	b, err := json.Marshal(s1)
	if err != nil {
		t.Fatal(err)
	}
	var s2 Specifier
	err = json.Unmarshal(b, &s2)
	if err != nil {
		t.Fatal(err)
	} else if s2 != s1 {
		t.Fatal("mismatch:", s1, s2)
	}

	// invalid json
	x := 3
	b, _ = json.Marshal(x)
	err = json.Unmarshal(b, &s2)
	if err == nil {
		t.Fatal("Unmarshal should have failed")
	}
}

func TestTransactionShortID(t *testing.T) {
	testCases := []struct {
		Height       BlockHeight
		TxSequenceID uint16
		ShortTxID    TransactionShortID
	}{
		// nil/default/zero, and also minimum
		{0, 0, 0},
		// the maximum possible value
		{1125899906842623, 16383, 18446744073709551615},
		// some other examples
		{1, 2, 16386},
		{0, 16383, 16383},
		{1125899906842623, 0, 18446744073709535232},
		{42, 13, 688141},
	}
	for _, testCase := range testCases {
		shortID := NewTransactionShortID(testCase.Height, testCase.TxSequenceID)
		if shortID != testCase.ShortTxID {
			t.Errorf("shortID (%v) != %v", shortID, testCase.ShortTxID)
		}
		if bh := shortID.BlockHeight(); bh != testCase.Height {
			t.Errorf("block height (%v) != %v", bh, testCase.Height)
		}
		if tsid := shortID.TransactionSequenceIndex(); tsid != testCase.TxSequenceID {
			t.Errorf("transaction seq ID (%v) != %v", tsid, testCase.TxSequenceID)
		}
	}
}

func TestTransactionVersionMarshaling(t *testing.T) {
	testCases := []struct {
		Version           TransactionVersion
		HexEncodedVersion string
	}{
		{TransactionVersionOne, "00"},
	}
	for idx, testCase := range testCases {
		buf := bytes.NewBuffer(nil)
		err := testCase.Version.MarshalSia(buf)
		if err != nil {
			t.Error(idx, err)
			continue
		}

		hexEncodedVersion := hex.EncodeToString(buf.Bytes())
		if hexEncodedVersion != testCase.HexEncodedVersion {
			t.Errorf("#%d: %q != %q", idx, hexEncodedVersion, testCase.HexEncodedVersion)
			continue
		}

		var txver TransactionVersion
		err = txver.UnmarshalSia(buf)
		if err != nil {
			t.Error(idx, err)
			continue
		}
		if txver != testCase.Version {
			t.Errorf("#%d: %v != %v", idx, txver, testCase.Version)
		}
	}
}

func TestTransactionEncodingDocExamples(t *testing.T) {
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
		ExpectedTx  Transaction
	}{
		{
			"0001000000000000002200000000000000000000000000000000000000000000000000000000000022013800000000000000656432353531390000000000000000002000000000000000ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff4000000000000000ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff00000000000000000000000000000000000000000000000001000000000000000100000000000000010000000000000000",
			Transaction{
				CoinInputs: []CoinInput{
					{
						ParentID: CoinOutputID(hs("2200000000000000000000000000000000000000000000000000000000000022")),
						Unlocker: InputLockProxy{
							t: UnlockTypeSingleSignature,
							il: &SingleSignatureInputLock{
								PublicKey: SiaPublicKey{
									Algorithm: SignatureEd25519,
									Key:       hbs("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"),
								},
								Signature: hbs("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"),
							},
						},
					},
				},
				MinerFees: []Currency{NewCurrency64(1)},
			},
		},
		{
			"0001000000000000002200000000000000000000000000000000000000000000000000000000000022013800000000000000656432353531390000000000000000002000000000000000ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff4000000000000000ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff020000000000000001000000000000000201cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc01000000000000000301dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd0000000000000000000000000000000001000000000000000100000000000000010000000000000000",
			Transaction{
				CoinInputs: []CoinInput{
					{
						ParentID: CoinOutputID(hs("2200000000000000000000000000000000000000000000000000000000000022")),
						Unlocker: InputLockProxy{
							t: UnlockTypeSingleSignature,
							il: &SingleSignatureInputLock{
								PublicKey: SiaPublicKey{
									Algorithm: SignatureEd25519,
									Key:       hbs("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"),
								},
								Signature: hbs("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"),
							},
						},
					},
				},
				CoinOutputs: []CoinOutput{
					{
						Value: NewCurrency64(2),
						UnlockHash: UnlockHash{
							Type: UnlockTypeSingleSignature,
							Hash: hs("cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"),
						},
					},
					{
						Value: NewCurrency64(3),
						UnlockHash: UnlockHash{
							Type: UnlockTypeSingleSignature,
							Hash: hs("dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd"),
						},
					},
				},
				MinerFees: []Currency{NewCurrency64(1)},
			},
		},
		{
			"0002000000000000002200000000000000000000000000000000000000000000000000000000000022013800000000000000656432353531390000000000000000002000000000000000ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff4000000000000000ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff3300000000000000000000000000000000000000000000000000000000000033026a00000000000000011234567891234567891234567891234567891234567891234567891234567891016363636363636363636363636363636363636363636363636363636363636363bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb07edb85a00000000a000000000000000656432353531390000000000000000002000000000000000abababababababababababababababababababababababababababababababab4000000000000000dededededededededededededededededededededededededededededededededededededededededededededededededededededededededededededededededabadabadabadabadabadabadabadabadabadabadabadabadabadabadabadaba020000000000000001000000000000000201cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc01000000000000000302dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd01000000000000004400000000000000000000000000000000000000000000000000000000000044013800000000000000656432353531390000000000000000002000000000000000eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee4000000000000000eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee010000000000000001000000000000002a01abcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcd010000000000000001000000000000000102000000000000003432",
			Transaction{
				CoinInputs: []CoinInput{
					{
						ParentID: CoinOutputID(hs("2200000000000000000000000000000000000000000000000000000000000022")),
						Unlocker: InputLockProxy{
							t: UnlockTypeSingleSignature,
							il: &SingleSignatureInputLock{
								PublicKey: SiaPublicKey{
									Algorithm: SignatureEd25519,
									Key:       hbs("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"),
								},
								Signature: hbs("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"),
							},
						},
					},
					{
						ParentID: CoinOutputID(hs("3300000000000000000000000000000000000000000000000000000000000033")),
						Unlocker: InputLockProxy{
							t: UnlockTypeAtomicSwap,
							il: &AtomicSwapInputLock{
								Sender: UnlockHash{
									Type: UnlockTypeSingleSignature,
									Hash: hs("1234567891234567891234567891234567891234567891234567891234567891"),
								},
								Receiver: UnlockHash{
									Type: UnlockTypeSingleSignature,
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
							},
						},
					},
				},
				CoinOutputs: []CoinOutput{
					{
						Value: NewCurrency64(2),
						UnlockHash: UnlockHash{
							Type: UnlockTypeSingleSignature,
							Hash: hs("cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"),
						},
					},
					{
						Value: NewCurrency64(3),
						UnlockHash: UnlockHash{
							Type: UnlockTypeAtomicSwap,
							Hash: hs("dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd"),
						},
					},
				},
				BlockStakeInputs: []BlockStakeInput{
					{
						ParentID: BlockStakeOutputID(hs("4400000000000000000000000000000000000000000000000000000000000044")),
						Unlocker: InputLockProxy{
							t: UnlockTypeSingleSignature,
							il: &SingleSignatureInputLock{
								PublicKey: SiaPublicKey{
									Algorithm: SignatureEd25519,
									Key:       hbs("eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee"),
								},
								Signature: hbs("eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee"),
							},
						},
					},
				},
				BlockStakeOutputs: []BlockStakeOutput{
					{
						Value: NewCurrency64(42),
						UnlockHash: UnlockHash{
							Type: UnlockTypeSingleSignature,
							Hash: hs("abcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcd"),
						},
					},
				},
				MinerFees:     []Currency{NewCurrency64(1)},
				ArbitraryData: []byte("42"),
			},
		},
	}
	for idx, example := range examples {
		encodedTx, err := hex.DecodeString(example.HexEncoding)
		if err != nil {
			t.Error(idx, err)
			continue
		}

		var tx Transaction
		err = tx.UnmarshalSia(bytes.NewReader(encodedTx))
		if err != nil {
			t.Error(idx, err)
			continue
		}

		jms := func(v interface{}) string {
			bs, _ := json.Marshal(v)
			return string(bs)
		}

		if !reflect.DeepEqual(example.ExpectedTx, tx) {
			t.Errorf("wrong tx hex decoding of example #%d: %v != %v", idx, jms(example.ExpectedTx), jms(tx))
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

func TestTransactionJSONEncodingExamples(t *testing.T) {
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
		ExpectedTx  Transaction
	}{
		{
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
		"minerfees": ["1"],
		"arbitrarydata": "SGVsbG8sIFdvcmxkIQ=="
	}
}`,
			Transaction{
				CoinInputs: []CoinInput{
					{
						ParentID: CoinOutputID(hs("0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")),
						Unlocker: InputLockProxy{
							t: UnlockTypeSingleSignature,
							il: &SingleSignatureInputLock{
								PublicKey: SiaPublicKey{
									Algorithm: SignatureEd25519,
									Key:       hbs("def123def123def123def123def123def123def123def123def123def123def1"),
								},
								Signature: hbs("ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef"),
							},
						},
					},
				},
				MinerFees:     []Currency{NewCurrency64(1)},
				ArbitraryData: []byte("Hello, World!"),
			},
		},
		{
			`{
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
						"sender": "010123456789012345678901234567890101234567890123456789012345678901dec8f8544d34",
						"receiver": "01abc0123abc0123abc0123abc0123abc0abc0123abc0123abc0123abc0123abc0efb39211ea2a",
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
				"unlockhash": "010123456789012345678901234567890101234567890123456789012345678901dec8f8544d34"
			},
			{
				"value": "5",
				"unlockhash": "01abc0123abc0123abc0123abc0123abc0abc0123abc0123abc0123abc0123abc0efb39211ea2a"
			},
			{
				"value": "8",
				"unlockhash": "02abc0123abc0123abc0123abc0123abc0abc0123abc0123abc0123abc0123abc0efb39211ea2a"
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
			},
			{
				"parentid": "fed42fed42fed42fed42fed42fed42fed42fed42fed42fed42fed42fed42fed4",
				"unlocker": {
					"type": 42,
					"condition": "Y29uZGl0aW9u",
					"fulfillment": "ZnVsZmlsbG1lbnQ="
				}
			}
		],
		"blockstakeoutputs": [
			{
				"value": "4",
				"unlockhash": "2a0123456789012345678901234567890101234567890123456789012345678901dec8f8544d34"
			},
			{
				"value": "2",
				"unlockhash": "18abc0123abc0123abc0123abc0123abc0abc0123abc0123abc0123abc0123abc0efb39211ea2a"
			}
		],
		"minerfees": ["1", "2", "3"],
		"arbitrarydata": "ZGF0YQ=="
	}
}`,
			Transaction{
				CoinInputs: []CoinInput{
					{
						ParentID: CoinOutputID(hs("abcdef012345abcdef012345abcdef012345abcdef012345abcdef012345abcd")),
						Unlocker: InputLockProxy{
							t: UnlockTypeSingleSignature,
							il: &SingleSignatureInputLock{
								PublicKey: SiaPublicKey{
									Algorithm: SignatureEd25519,
									Key:       hbs("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"),
								},
								Signature: hbs("abcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefab"),
							},
						},
					},
					{
						ParentID: CoinOutputID(hs("012345defabc012345defabc012345defabc012345defabc012345defabc0123")),
						Unlocker: InputLockProxy{
							t: UnlockTypeAtomicSwap,
							il: &AtomicSwapInputLock{
								Sender: UnlockHash{
									Type: UnlockTypeSingleSignature,
									Hash: hs("0123456789012345678901234567890101234567890123456789012345678901"),
								},
								Receiver: UnlockHash{
									Type: UnlockTypeSingleSignature,
									Hash: hs("abc0123abc0123abc0123abc0123abc0abc0123abc0123abc0123abc0123abc0"),
								},
								HashedSecret: AtomicSwapHashedSecret(hs("abc543defabc543defabc543defabc543defabc543defabc543defabc543defa")),
								TimeLock:     1522068743,
								PublicKey: SiaPublicKey{
									Algorithm: SignatureEd25519,
									Key:       hbs("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"),
								},
								Signature: hbs("abcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefab"),
								Secret:    AtomicSwapSecret(hs("def789def789def789def789def789dedef789def789def789def789def789de")),
							},
						},
					},
				},
				CoinOutputs: []CoinOutput{
					{
						Value: NewCurrency64(3),
						UnlockHash: UnlockHash{
							Type: UnlockTypeSingleSignature,
							Hash: hs("0123456789012345678901234567890101234567890123456789012345678901"),
						},
					},
					{
						Value: NewCurrency64(5),
						UnlockHash: UnlockHash{
							Type: UnlockTypeSingleSignature,
							Hash: hs("abc0123abc0123abc0123abc0123abc0abc0123abc0123abc0123abc0123abc0"),
						},
					},
					{
						Value: NewCurrency64(8),
						UnlockHash: UnlockHash{
							Type: UnlockTypeAtomicSwap,
							Hash: hs("abc0123abc0123abc0123abc0123abc0abc0123abc0123abc0123abc0123abc0"),
						},
					},
				},
				BlockStakeInputs: []BlockStakeInput{
					{
						ParentID: BlockStakeOutputID(hs("dfd23dfd23dfd23dfd23dfd23dfd23dfd23dfd23dfd23dfd23dfd23dfd23dfde")),
						Unlocker: InputLockProxy{
							t: UnlockTypeSingleSignature,
							il: &SingleSignatureInputLock{
								PublicKey: SiaPublicKey{
									Algorithm: SignatureEd25519,
									Key:       hbs("ef1234ef1234ef1234ef1234ef1234ef1234ef1234ef1234ef1234ef1234ef12"),
								},
								Signature: hbs("01234def01234def01234def01234def01234def01234def01234def01234def01234def01234def01234def01234def01234def01234def01234def01234def"),
							},
						},
					},
					{
						ParentID: BlockStakeOutputID(hs("fed42fed42fed42fed42fed42fed42fed42fed42fed42fed42fed42fed42fed4")),
						Unlocker: InputLockProxy{
							t: 42,
							il: &UnknownInputLock{
								Condition:   []byte("condition"),
								Fulfillment: []byte("fulfillment"),
							},
						},
					},
				},
				BlockStakeOutputs: []BlockStakeOutput{
					{
						Value: NewCurrency64(4),
						UnlockHash: UnlockHash{
							Type: 42,
							Hash: hs("0123456789012345678901234567890101234567890123456789012345678901"),
						},
					},
					{
						Value: NewCurrency64(2),
						UnlockHash: UnlockHash{
							Type: 24,
							Hash: hs("abc0123abc0123abc0123abc0123abc0abc0123abc0123abc0123abc0123abc0"),
						},
					},
				},
				MinerFees: []Currency{
					NewCurrency64(1), NewCurrency64(2), NewCurrency64(3),
				},
				ArbitraryData: []byte("data"),
			},
		},
	}

	for idx, example := range examples {
		var tx Transaction
		err := json.Unmarshal([]byte(example.JSONEncoded), &tx)
		if err != nil {
			t.Error(idx, err)
			continue
		}
		if !reflect.DeepEqual(example.ExpectedTx, tx) {
			t.Errorf("#%d: %v != %v", idx, example.ExpectedTx, tx)
			continue
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
