package types

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/rivine/rivine/crypto"
	"github.com/rivine/rivine/encoding"
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
		{TransactionVersionZero, "00"},
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
									Hash: hs("654f96b317efe5fd6cd8ba1a394dce7b6ebe8c9621d6c44cbe3c8f1b58ce632a"),
								},
								Receiver: UnlockHash{
									Type: UnlockTypeSingleSignature,
									Hash: hs("e89843e4b8231a01ba18b254d530110364432aafab8206bea72e5a20eaa55f70"),
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
							Hash: hs("42e9458e348598111b0bc19bda18e45835605db9f4620616d752220ae8605ce0"),
						},
					},
					{
						Value: NewCurrency64(5),
						UnlockHash: UnlockHash{
							Type: UnlockTypeSingleSignature,
							Hash: hs("a6a6c5584b2bfbd08738996cd7930831f958b9a5ed1595525236e861c1a0dc35"),
						},
					},
					{
						Value: NewCurrency64(8),
						UnlockHash: UnlockHash{
							Type: UnlockTypeAtomicSwap,
							Hash: hs("a24c97c80eeac111aa4bcbb0ac8ffc364fa9b22da10d3054778d2332f68b365e"),
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
				},
				BlockStakeOutputs: []BlockStakeOutput{
					{
						Value: NewCurrency64(4),
						UnlockHash: UnlockHash{
							Type: 100,
							Hash: hs("53402d094ed0f336950c4be0feec37167aaaaf8b974d265900e49ab22773584c"),
						},
					},
					{
						Value: NewCurrency64(2),
						UnlockHash: UnlockHash{
							Type: 42,
							Hash: hs("b39baa9a58319fa47f78ed542a733a7198d106caeabf0a231b91ea3e4e222ffd"),
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
func TestOutputIDStringJSONEncoding(t *testing.T) {
	testIDStringJSONEncoding(t, func(h *[crypto.HashSize]byte) idEncoder {
		if h == nil {
			return new(OutputID)
		}
		return (*OutputID)(h)
	})
}

func TestTransactionIDStringJSONEncoding(t *testing.T) {
	testIDStringJSONEncoding(t, func(h *[crypto.HashSize]byte) idEncoder {
		if h == nil {
			return new(TransactionID)
		}
		return (*TransactionID)(h)
	})
}

func TestCoinOutputIDStringJSONEncoding(t *testing.T) {
	testIDStringJSONEncoding(t, func(h *[crypto.HashSize]byte) idEncoder {
		if h == nil {
			return new(CoinOutputID)
		}
		return (*CoinOutputID)(h)
	})
}

func TestBlockStakeOutputIDStringJSONEncoding(t *testing.T) {
	testIDStringJSONEncoding(t, func(h *[crypto.HashSize]byte) idEncoder {
		if h == nil {
			return new(BlockStakeOutputID)
		}
		return (*BlockStakeOutputID)(h)
	})
}

func testIDStringJSONEncoding(t *testing.T, f func(*[crypto.HashSize]byte) idEncoder) {
	for i := 0; i < 128; i++ {
		// generate fresh id
		var s [crypto.HashSize]byte
		rand.Read(s[:])
		inputID := f(&s)

		// test string encoding
		str := inputID.String()
		if len(str) == 0 {
			t.Errorf("#%d: string length is 0", i)
			continue
		}
		outputID := f(nil)
		err := outputID.LoadString(str)
		if err != nil {
			t.Error(i, err)
			continue
		}
		if !reflect.DeepEqual(inputID, outputID) {
			t.Errorf("%d => %v != %v", i, inputID, outputID)
		}

		// test JSON encoding
		bs, err := inputID.MarshalJSON()
		if err != nil {
			t.Error(i, err)
			continue
		}
		outputID = f(nil)
		err = outputID.UnmarshalJSON(bs)
		if err != nil {
			t.Error(i, err)
			continue
		}
		if !reflect.DeepEqual(inputID, outputID) {
			t.Errorf("%d => %v != %v", i, inputID, outputID)
		}
	}
}

type idEncoder interface {
	json.Marshaler
	json.Unmarshaler

	fmt.Stringer
	LoadString(string) error
}

// legacy transactions (version 0x00)
var legacyHexTestCases = []string{
	`0001000000000000002200000000000000000000000000000000000000000000000000000000000022013800000000000000656432353531390000000000000000002000000000000000ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff4000000000000000ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff00000000000000000000000000000000000000000000000001000000000000000100000000000000010000000000000000`,
	`0002000000000000002200000000000000000000000000000000000000000000000000000000000022013800000000000000656432353531390000000000000000002000000000000000ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff4000000000000000ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff3300000000000000000000000000000000000000000000000000000000000033026a00000000000000011234567891234567891234567891234567891234567891234567891234567891016363636363636363636363636363636363636363636363636363636363636363bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb07edb85a00000000a000000000000000656432353531390000000000000000002000000000000000abababababababababababababababababababababababababababababababab4000000000000000dededededededededededededededededededededededededededededededededededededededededededededededededededededededededededededededededabadabadabadabadabadabadabadabadabadabadabadabadabadabadabadaba020000000000000001000000000000000201cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc01000000000000000302dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd01000000000000004400000000000000000000000000000000000000000000000000000000000044013800000000000000656432353531390000000000000000002000000000000000eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee4000000000000000eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee010000000000000001000000000000002a01abcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcd010000000000000001000000000000000102000000000000003432`,
}

func TestUnknownVersionBinaryEncoding(t *testing.T) {
	testCases := append(legacyHexTestCases,
		// transactions with unknown transaction versions
		`2a170000000000000048656c6c6f2c20526177205472616e73616374696f6e21`,
	)
	for idx, inputHexTxn := range testCases {
		// sanity check to ensure our hex is valid
		encodedTx, err := hex.DecodeString(inputHexTxn)
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
		// change tx version number to something unknown
		origVersion := tx.Version
		// ensure our origVersion does not equal 255 already, as this would fuck it all
		if origVersion == 255 {
			t.Error(idx, "transaction version should not be 255 for this test")
			continue
		}
		tx.Version = 255

		// serialize it once again
		buf := bytes.NewBuffer(nil)
		err = tx.MarshalSia(buf)
		if err != nil {
			t.Error(idx, err)
			continue
		}
		inputTxBytes := make([]byte, len(buf.Bytes()))
		copy(inputTxBytes, buf.Bytes())

		compareOffset := 0
		if origVersion == TransactionVersionZero {
			// extra offset needed for this step
			compareOffset = 8
		}
		// should equal the same bytes, except for the modified version number
		if bytes.Compare(encodedTx[1:], inputTxBytes[1+compareOffset:]) != 0 {
			t.Errorf("#%d: %v != %v", idx, encodedTx[1:], inputTxBytes[1+compareOffset:])
			continue
		}
		if inputTxBytes[0] != 255 {
			t.Error(idx, "unexpected version number", inputTxBytes[0])
			continue
		}

		tx = Transaction{} // clean any state
		// now decode it again, should be done using unknown decoder
		err = tx.UnmarshalSia(bytes.NewReader(inputTxBytes))
		if err != nil {
			t.Error(idx, err)
			continue
		}
		// validate that our tx version is still 255, the hacked version number
		if tx.Version != 255 {
			t.Error(idx, "invalid/unexpected transaction version", tx.Version)
			continue
		}
		// validate that our ext now equals the unknown extension
		if ext, ok := tx.Extension.(unknownTransactionExtension); !ok {
			t.Errorf("#%d: invalid/unexpected txn extension: %[2]v (%[2]T)", idx, tx.Extension)
			continue
		} else {
			offset := 9
			if origVersion == TransactionVersionZero {
				// in legacy versions there was no length encoded, thus no need to offset by 9
				offset = 1
			}
			// validate that our raw data equals our expexted binary encoding
			if bytes.Compare(encodedTx[offset:], ext.rawData) != 0 {
				t.Errorf("#%d: %v != %v", idx, encodedTx[1:], ext.rawData)
				continue
			}
		}

		// now change version again, and serialize it,
		// should be once again fine, and should still be serialized using our unknown encoder
		tx.Version = origVersion
		buf = bytes.NewBuffer(nil)
		err = tx.MarshalSia(buf)
		if err != nil {
			t.Error(idx, err)
			continue
		}

		outputBytes := buf.Bytes()
		// again, if we were on the legacy version,
		// we'll have to remove the length now, to come to the same hex string
		if origVersion == TransactionVersionZero {
			outputBytes = append(outputBytes[:1], outputBytes[9:]...)
		}

		// now turn it into a hex string again, and compare with the original input case, should be equal
		outputHexTxn := hex.EncodeToString(outputBytes)
		if outputHexTxn != inputHexTxn {
			t.Error(idx, outputHexTxn, "!=", inputHexTxn)
		}
	}
}

// legacy test to ensure we're compatible with the old transaction ID computation logic
// as that logic has changed since issue/feature #201
func TestIDComputationCompatibleWithLegacyIDs(t *testing.T) {
	for idx, inputHexTxn := range legacyHexTestCases {
		// sanity check to ensure our hex is valid
		encodedTx, err := hex.DecodeString(inputHexTxn)
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

		// compare ID, CoinOutputID and BlockStakeOutputID
		// these should be equal
		idA, idB := tx.ID(), tx.LegacyID()
		if bytes.Compare(idA[:], idB[:]) != 0 {
			t.Error(idx, idA, "!=", idB)
			continue
		}
		coinOutputIDA, coinOutputIDB := tx.CoinOutputID(42), tx.LegacyCoinOutputID(42)
		if bytes.Compare(coinOutputIDA[:], coinOutputIDB[:]) != 0 {
			t.Error(idx, coinOutputIDA, "!=", coinOutputIDB)
			continue
		}
		blockStakeOutputIDA, blockStakeOutputIDB := tx.BlockStakeOutputID(42), tx.LegacyBlockStakeOutputID(42)
		if bytes.Compare(blockStakeOutputIDA[:], blockStakeOutputIDB[:]) != 0 {
			t.Error(idx, blockStakeOutputIDA, "!=", blockStakeOutputIDB)
		}

		// now change it to something else than 0x00, but still without a custom encoder/decoder,
		// this should give it a very new ID
		tx.Version = TransactionVersionZero + 1
		// compare ID, CoinOutputID and BlockStakeOutputID
		// these should now be different
		idA, idB = tx.ID(), tx.LegacyID()
		if bytes.Compare(idA[:], idB[:]) == 0 {
			t.Error(idx, idA, "==", idB)
			continue
		}
		coinOutputIDA, coinOutputIDB = tx.CoinOutputID(42), tx.LegacyCoinOutputID(42)
		if bytes.Compare(coinOutputIDA[:], coinOutputIDB[:]) == 0 {
			t.Error(idx, coinOutputIDA, "==", coinOutputIDB)
			continue
		}
		blockStakeOutputIDA, blockStakeOutputIDB = tx.BlockStakeOutputID(42), tx.LegacyBlockStakeOutputID(42)
		if bytes.Compare(blockStakeOutputIDA[:], blockStakeOutputIDB[:]) == 0 {
			t.Error(idx, blockStakeOutputIDA, "==", blockStakeOutputIDB)
		}
	}
}

// ID returns the id of a transaction, which is taken by marshalling all of the
// fields except for the signatures and taking the hash of the result.
func (t Transaction) LegacyID() TransactionID {
	return TransactionID(crypto.HashAll(
		t.CoinInputs,
		t.CoinOutputs,
		t.BlockStakeInputs,
		t.BlockStakeOutputs,
		t.MinerFees,
		t.ArbitraryData,
	))
}

// CoinOutputID returns the ID of a coin output at the given index,
// which is calculated by hashing the concatenation of the CoinOutput
// Specifier, all of the fields in the transaction (except the signatures),
// and output index.
func (t Transaction) LegacyCoinOutputID(i uint64) CoinOutputID {
	return CoinOutputID(crypto.HashAll(
		SpecifierCoinOutput,
		t.CoinInputs,
		t.CoinOutputs,
		t.BlockStakeInputs,
		t.BlockStakeOutputs,
		t.MinerFees,
		t.ArbitraryData,
		i,
	))
}

// BlockStakeOutputID returns the ID of a BlockStakeOutput at the given index, which
// is calculated by hashing the concatenation of the BlockStakeOutput Specifier,
// all of the fields in the transaction (except the signatures), and output
// index.
func (t Transaction) LegacyBlockStakeOutputID(i uint64) BlockStakeOutputID {
	return BlockStakeOutputID(crypto.HashAll(
		SpecifierBlockStakeOutput,
		t.CoinInputs,
		t.CoinOutputs,
		t.BlockStakeInputs,
		t.BlockStakeOutputs,
		t.MinerFees,
		t.ArbitraryData,
		i,
	))
}

func TestInputSigHashComputationCompatibleWithLegacyInputSigHash(t *testing.T) {
	for idx, inputHexTxn := range legacyHexTestCases {
		// sanity check to ensure our hex is valid
		encodedTx, err := hex.DecodeString(inputHexTxn)
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

		// compare inputSigHash
		inputSigHashA, inputSigHashB :=
			tx.InputSigHash(42, "foo"),
			tx.LegacyInputSigHash(42, "foo")
		if bytes.Compare(inputSigHashA[:], inputSigHashB[:]) != 0 {
			t.Error(idx, inputSigHashA, "!=", inputSigHashB)
			continue
		}

		// now change it to something else than 0x00, but still without a custom inputSigHasher,
		// even so it should give a different input signature hash, regardless of
		tx.Version = TransactionVersionZero + 1
		// compare ID, CoinOutputID and BlockStakeOutputID
		// these should now be different
		inputSigHashA, inputSigHashB =
			tx.InputSigHash(42, "foo"),
			tx.LegacyInputSigHash(42, "foo")
		if bytes.Compare(inputSigHashA[:], inputSigHashB[:]) == 0 {
			t.Error(idx, inputSigHashA, "==", inputSigHashB)
		}
	}
}

func (t Transaction) LegacyInputSigHash(inputIndex uint64, extraObjects ...interface{}) (hash crypto.Hash) {
	h := crypto.NewHash()
	enc := encoding.NewEncoder(h)

	enc.Encode(inputIndex)
	if len(extraObjects) > 0 {
		enc.EncodeAll(extraObjects...)
	}
	for _, ci := range t.CoinInputs {
		enc.EncodeAll(ci.ParentID, ci.Unlocker.UnlockHash())
	}
	enc.Encode(t.CoinOutputs)
	for _, bsi := range t.BlockStakeInputs {
		enc.EncodeAll(bsi.ParentID, bsi.Unlocker.UnlockHash())
	}
	enc.EncodeAll(
		t.BlockStakeOutputs,
		t.MinerFees,
		t.ArbitraryData,
	)

	h.Sum(hash[:0])
	return
}
