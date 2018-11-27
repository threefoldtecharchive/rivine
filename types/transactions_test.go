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

	"github.com/threefoldtech/rivine/build"
	"github.com/threefoldtech/rivine/crypto"
)

// TestTransactionIDs probes all of the ID functions of the Transaction type.
func TestIDs(t *testing.T) {
	// Create every type of ID using empty fields.
	txn := Transaction{
		Version:           TestnetChainConstants().DefaultTransactionVersion,
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
		Version: TestnetChainConstants().DefaultTransactionVersion,
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

func TestTransactionVersionSiaMarshaling(t *testing.T) {
	testCases := []struct {
		Version           TransactionVersion
		HexEncodedVersion string
	}{
		{TransactionVersionZero, "00"},
		{TransactionVersionOne, "01"},
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

func TestTransactionVersionRivineMarshaling(t *testing.T) {
	testCases := []struct {
		Version           TransactionVersion
		HexEncodedVersion string
	}{
		{TransactionVersionZero, "00"},
		{TransactionVersionOne, "01"},
	}
	for idx, testCase := range testCases {
		buf := bytes.NewBuffer(nil)
		err := testCase.Version.MarshalRivine(buf)
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
		err = txver.UnmarshalRivine(buf)
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
		// v0 @ v1.0.2
		{
			"0001000000000000002200000000000000000000000000000000000000000000000000000000000022013800000000000000656432353531390000000000000000002000000000000000ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff4000000000000000ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff00000000000000000000000000000000000000000000000001000000000000000100000000000000010000000000000000",
			Transaction{
				Version: TransactionVersionZero,
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
			"0001000000000000002200000000000000000000000000000000000000000000000000000000000022013800000000000000656432353531390000000000000000002000000000000000ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff4000000000000000ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff020000000000000001000000000000000201cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc01000000000000000301dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd0000000000000000000000000000000001000000000000000100000000000000010000000000000000",
			Transaction{
				Version: TransactionVersionZero,
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
							Type: UnlockTypePubKey,
							Hash: hs("dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd"),
						})),
					},
				},
				MinerFees: []Currency{NewCurrency64(1)},
			},
		},
		{
			"0002000000000000002200000000000000000000000000000000000000000000000000000000000022013800000000000000656432353531390000000000000000002000000000000000ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff4000000000000000ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff3300000000000000000000000000000000000000000000000000000000000033026a00000000000000011234567891234567891234567891234567891234567891234567891234567891016363636363636363636363636363636363636363636363636363636363636363bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb07edb85a00000000a000000000000000656432353531390000000000000000002000000000000000abababababababababababababababababababababababababababababababab4000000000000000dededededededededededededededededededededededededededededededededededededededededededededededededededededededededededededededededabadabadabadabadabadabadabadabadabadabadabadabadabadabadabadaba020000000000000001000000000000000201cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc01000000000000000302dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd01000000000000004400000000000000000000000000000000000000000000000000000000000044013800000000000000656432353531390000000000000000002000000000000000eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee4000000000000000eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee010000000000000001000000000000002a01abcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcd010000000000000001000000000000000102000000000000003432",
			Transaction{
				Version: TransactionVersionZero,
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
				ArbitraryData: ArbitraryData{Data: []byte("42")},
			},
		},
		// v0 @ v1.0.3
		{
			"00000000000000000001000000000000000800000000000000016345785d8a000001fc8714235d549f890f35e52d745b9eeeee34926f96c4b9ef1689832f338d93490000000000000000010000000000000002000000000000000bb801fc8714235d549f890f35e52d745b9eeeee34926f96c4b9ef1689832f338d934900000000000000000000000000000000",
			Transaction{
				Version:    TransactionVersionZero,
				CoinInputs: []CoinInput{},
				CoinOutputs: []CoinOutput{
					{
						Value:     NewCurrency64(100000000000000000),
						Condition: NewCondition(NewUnlockHashCondition(unlockHashFromHex("01fc8714235d549f890f35e52d745b9eeeee34926f96c4b9ef1689832f338d9349b453898f7e51"))),
					},
				},
				BlockStakeOutputs: []BlockStakeOutput{
					{
						Value:     NewCurrency64(3000),
						Condition: NewCondition(NewUnlockHashCondition(unlockHashFromHex("01fc8714235d549f890f35e52d745b9eeeee34926f96c4b9ef1689832f338d9349b453898f7e51"))),
					},
				},
				MinerFees: nil,
			},
		},
		// v1 @ v1.0.4
		{
			"01e20000000000000001000000000000002200000000000000000000000000000000000000000000000000000000000022018000000000000000656432353531390000000000000000002000000000000000ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff4000000000000000ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff00000000000000000000000000000000000000000000000001000000000000000100000000000000010000000000000000",
			Transaction{
				Version: TransactionVersionOne,
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
			"01f40000000000000001000000000000001100000000000000000000000000000000000000000000000000000000000011018000000000000000656432353531390000000000000000002000000000000000ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff4000000000000000ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff01000000000000000100000000000000090000000000000000000000000000000000000000000000000001000000000000000100000000000000030000000000000000",
			Transaction{
				Version: TransactionVersionOne,
				CoinInputs: []CoinInput{
					{
						ParentID: CoinOutputID(hs("1100000000000000000000000000000000000000000000000000000000000011")),
						Fulfillment: NewFulfillment(&SingleSignatureFulfillment{
							PublicKey: PublicKey{
								Algorithm: SignatureAlgoEd25519,
								Key:       hbs("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"),
							},
							Signature: hbs("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"),
						}),
					},
				},
				CoinOutputs: []CoinOutput{
					{
						Value:     NewCurrency64(9),
						Condition: NewCondition(&NilCondition{}), // `nil` would be functionally equal, but it will give a non-deep-equal result
					},
				},
				MinerFees: []Currency{NewCurrency64(3)},
			},
		},
		{
			"01480100000000000001000000000000002200000000000000000000000000000000000000000000000000000000000022018000000000000000656432353531390000000000000000002000000000000000ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff4000000000000000ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff020000000000000001000000000000000201210000000000000001cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc01000000000000000301210000000000000001dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd0000000000000000000000000000000001000000000000000100000000000000010000000000000000",
			Transaction{
				Version: TransactionVersionOne,
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
							Type: UnlockTypePubKey,
							Hash: hs("dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd"),
						})),
					},
				},
				MinerFees: []Currency{NewCurrency64(1)},
			},
		},
		{
			"019e0400000000000003000000000000002200000000000000000000000000000000000000000000000000000000000022018000000000000000656432353531390000000000000000002000000000000000ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff4000000000000000ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff330000000000000000000000000000000000000000000000000000000000003302a000000000000000656432353531390000000000000000002000000000000000abababababababababababababababababababababababababababababababab4000000000000000dededededededededededededededededededededededededededededededededededededededededededededededededededededededededededededededededabadabadabadabadabadabadabadabadabadabadabadabadabadabadabadaba4400000000000000000000000000000000000000000000000000000000000044020a01000000000000011234567891234567891234567891234567891234567891234567891234567891016363636363636363636363636363636363636363636363636363636363636363bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb07edb85a00000000656432353531390000000000000000002000000000000000abababababababababababababababababababababababababababababababab4000000000000000dededededededededededededededededededededededededededededededededededededededededededededededededededededededededededededededededabadabadabadabadabadabadabadabadabadabadabadabadabadabadabadaba030000000000000001000000000000000201210000000000000001cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc01000000000000000301210000000000000002dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd010000000000000004026a00000000000000011234567891234567891234567891234567891234567891234567891234567891016363636363636363636363636363636363636363636363636363636363636363bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb07edb85a0000000001000000000000004400000000000000000000000000000000000000000000000000000000000044018000000000000000656432353531390000000000000000002000000000000000eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee4000000000000000eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee010000000000000001000000000000002a01210000000000000001abcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcd010000000000000001000000000000000102000000000000003432",
			Transaction{
				Version: TransactionVersionOne,
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
						Fulfillment: NewFulfillment(&anyAtomicSwapFulfillment{
							atomicSwapFulfillment: &AtomicSwapFulfillment{
								PublicKey: PublicKey{
									Algorithm: SignatureAlgoEd25519,
									Key:       hbs("abababababababababababababababababababababababababababababababab"),
								},
								Signature: hbs("dededededededededededededededededededededededededededededededededededededededededededededededededededededededededededededededede"),
								Secret:    AtomicSwapSecret(hs("dabadabadabadabadabadabadabadabadabadabadabadabadabadabadabadaba")),
							},
						}),
					},
					{
						ParentID: CoinOutputID(hs("4400000000000000000000000000000000000000000000000000000000000044")),
						Fulfillment: NewFulfillment(&anyAtomicSwapFulfillment{
							atomicSwapFulfillment: &LegacyAtomicSwapFulfillment{
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
							},
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
					{
						Value: NewCurrency64(4),
						Condition: NewCondition(&AtomicSwapCondition{
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
						}),
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
				ArbitraryData: ArbitraryData{Data: []byte("42")},
			},
		},
		{
			"01fd0000000000000001000000000000001100000000000000000000000000000000000000000000000000000000000011018000000000000000656432353531390000000000000000002000000000000000ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff4000000000000000ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff01000000000000000100000000000000090309000000000000002a00000000000000000000000000000000000000000000000001000000000000000100000000000000030000000000000000",
			Transaction{
				Version: TransactionVersionOne,
				CoinInputs: []CoinInput{
					{
						ParentID: CoinOutputID(hs("1100000000000000000000000000000000000000000000000000000000000011")),
						Fulfillment: NewFulfillment(&SingleSignatureFulfillment{
							PublicKey: PublicKey{
								Algorithm: SignatureAlgoEd25519,
								Key:       hbs("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"),
							},
							Signature: hbs("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"),
						}),
					},
				},
				CoinOutputs: []CoinOutput{
					{
						Value: NewCurrency64(9),
						Condition: NewCondition(&TimeLockCondition{
							LockTime:  42,
							Condition: &NilCondition{}, // `nil` would be functionally equal, but it will give a non-deep-equal result
						}),
					},
				},
				MinerFees: []Currency{NewCurrency64(3)},
			},
		},
		{
			"011e0100000000000001000000000000001100000000000000000000000000000000000000000000000000000000000011018000000000000000656432353531390000000000000000002000000000000000ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff4000000000000000ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff0100000000000000010000000000000009032a000000000000002a000000000000000101abcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcd0000000000000000000000000000000001000000000000000100000000000000030000000000000000",
			Transaction{
				Version: TransactionVersionOne,
				CoinInputs: []CoinInput{
					{
						ParentID: CoinOutputID(hs("1100000000000000000000000000000000000000000000000000000000000011")),
						Fulfillment: NewFulfillment(&SingleSignatureFulfillment{
							PublicKey: PublicKey{
								Algorithm: SignatureAlgoEd25519,
								Key:       hbs("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"),
							},
							Signature: hbs("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"),
						}),
					},
				},
				CoinOutputs: []CoinOutput{
					{
						Value: NewCurrency64(9),
						Condition: NewCondition(&TimeLockCondition{
							LockTime: 42,
							Condition: NewUnlockHashCondition(UnlockHash{
								Type: UnlockTypePubKey,
								Hash: hs("abcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcd"),
							}),
						}),
					},
				},
				MinerFees: []Currency{NewCurrency64(3)},
			},
		},
		{
			"018101000000000000010000000000000022000000000000000000000000000000000000000000000000000000000000220388000000000000000100000000000000656432353531390000000000000000002000000000000000ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff4000000000000000ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff020000000000000001000000000000000201210000000000000001cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc0100000000000000030452000000000000000200000000000000020000000000000001dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd01bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb0000000000000000000000000000000001000000000000000100000000000000010000000000000000",
			Transaction{
				Version: TransactionVersionOne,
				CoinInputs: []CoinInput{
					{
						ParentID: CoinOutputID(hs("2200000000000000000000000000000000000000000000000000000000000022")),
						Fulfillment: NewFulfillment(&MultiSignatureFulfillment{
							Pairs: []PublicKeySignaturePair{
								{
									PublicKey: PublicKey{
										Algorithm: SignatureAlgoEd25519,
										Key:       hbs("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"),
									},
									Signature: hbs("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"),
								},
							},
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
						Condition: NewCondition(NewMultiSignatureCondition(UnlockHashSlice{
							{
								Type: UnlockTypePubKey,
								Hash: hs("dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd"),
							},
							{
								Type: UnlockTypePubKey,
								Hash: hs("bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"),
							},
						}, 2)),
					},
				},
				MinerFees: []Currency{NewCurrency64(1)},
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
		// v0 @ v1.0.2
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
				Version: TransactionVersionZero,
				CoinInputs: []CoinInput{
					{
						ParentID: CoinOutputID(hs("0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")),
						Fulfillment: NewFulfillment(&SingleSignatureFulfillment{
							PublicKey: PublicKey{
								Algorithm: SignatureAlgoEd25519,
								Key:       hbs("def123def123def123def123def123def123def123def123def123def123def1"),
							},
							Signature: hbs("ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef"),
						}),
					},
				},
				MinerFees:     []Currency{NewCurrency64(1)},
				ArbitraryData: ArbitraryData{Data: []byte("Hello, World!")},
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
				Version: TransactionVersionZero,
				CoinInputs: []CoinInput{
					{
						ParentID: CoinOutputID(hs("abcdef012345abcdef012345abcdef012345abcdef012345abcdef012345abcd")),
						Fulfillment: NewFulfillment(&SingleSignatureFulfillment{
							PublicKey: PublicKey{
								Algorithm: SignatureAlgoEd25519,
								Key:       hbs("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"),
							},
							Signature: hbs("abcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefab"),
						}),
					},
					{
						ParentID: CoinOutputID(hs("012345defabc012345defabc012345defabc012345defabc012345defabc0123")),
						Fulfillment: NewFulfillment(&LegacyAtomicSwapFulfillment{
							Sender: UnlockHash{
								Type: UnlockTypePubKey,
								Hash: hs("654f96b317efe5fd6cd8ba1a394dce7b6ebe8c9621d6c44cbe3c8f1b58ce632a"),
							},
							Receiver: UnlockHash{
								Type: UnlockTypePubKey,
								Hash: hs("e89843e4b8231a01ba18b254d530110364432aafab8206bea72e5a20eaa55f70"),
							},
							HashedSecret: AtomicSwapHashedSecret(hs("abc543defabc543defabc543defabc543defabc543defabc543defabc543defa")),
							TimeLock:     1522068743,
							PublicKey: PublicKey{
								Algorithm: SignatureAlgoEd25519,
								Key:       hbs("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"),
							},
							Signature: hbs("abcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefab"),
							Secret:    AtomicSwapSecret(hs("def789def789def789def789def789dedef789def789def789def789def789de")),
						}),
					},
				},
				CoinOutputs: []CoinOutput{
					{
						Value: NewCurrency64(3),
						Condition: NewCondition(NewUnlockHashCondition(UnlockHash{
							Type: UnlockTypePubKey,
							Hash: hs("42e9458e348598111b0bc19bda18e45835605db9f4620616d752220ae8605ce0"),
						})),
					},
					{
						Value: NewCurrency64(5),
						Condition: NewCondition(NewUnlockHashCondition(UnlockHash{
							Type: UnlockTypePubKey,
							Hash: hs("a6a6c5584b2bfbd08738996cd7930831f958b9a5ed1595525236e861c1a0dc35"),
						})),
					},
					{
						Value: NewCurrency64(8),
						Condition: NewCondition(NewUnlockHashCondition(UnlockHash{
							Type: UnlockTypeAtomicSwap,
							Hash: hs("a24c97c80eeac111aa4bcbb0ac8ffc364fa9b22da10d3054778d2332f68b365e"),
						})),
					},
				},
				BlockStakeInputs: []BlockStakeInput{
					{
						ParentID: BlockStakeOutputID(hs("dfd23dfd23dfd23dfd23dfd23dfd23dfd23dfd23dfd23dfd23dfd23dfd23dfde")),
						Fulfillment: NewFulfillment(&SingleSignatureFulfillment{
							PublicKey: PublicKey{
								Algorithm: SignatureAlgoEd25519,
								Key:       hbs("ef1234ef1234ef1234ef1234ef1234ef1234ef1234ef1234ef1234ef1234ef12"),
							},
							Signature: hbs("01234def01234def01234def01234def01234def01234def01234def01234def01234def01234def01234def01234def01234def01234def01234def01234def"),
						}),
					},
				},
				BlockStakeOutputs: []BlockStakeOutput{
					{
						Value: NewCurrency64(4),
						Condition: NewCondition(NewUnlockHashCondition(UnlockHash{
							Type: 100,
							Hash: hs("53402d094ed0f336950c4be0feec37167aaaaf8b974d265900e49ab22773584c"),
						})),
					},
					{
						Value: NewCurrency64(2),
						Condition: NewCondition(NewUnlockHashCondition(UnlockHash{
							Type: 42,
							Hash: hs("b39baa9a58319fa47f78ed542a733a7198d106caeabf0a231b91ea3e4e222ffd"),
						})),
					},
				},
				MinerFees: []Currency{
					NewCurrency64(1), NewCurrency64(2), NewCurrency64(3),
				},
				ArbitraryData: ArbitraryData{Data: []byte("data")},
			},
		},
		// v0 @ v1.0.3
		{
			`{
	"version": 0,
	"data": {
		"coininputs": [],
		"coinoutputs": [{
			"value": "100000000000000000",
			"unlockhash": "01fc8714235d549f890f35e52d745b9eeeee34926f96c4b9ef1689832f338d9349b453898f7e51"
		}],
		"blockstakeoutputs": [{
			"value": "3000",
			"unlockhash": "01fc8714235d549f890f35e52d745b9eeeee34926f96c4b9ef1689832f338d9349b453898f7e51"
		}],
		"minerfees": null
	}
}`,
			Transaction{
				Version:    TransactionVersionZero,
				CoinInputs: []CoinInput{},
				CoinOutputs: []CoinOutput{
					{
						Value: NewCurrency64(100000000000000000),
						Condition: NewCondition(NewUnlockHashCondition(
							unlockHashFromHex("01fc8714235d549f890f35e52d745b9eeeee34926f96c4b9ef1689832f338d9349b453898f7e51"))),
					},
				},
				BlockStakeOutputs: []BlockStakeOutput{
					{
						Value: NewCurrency64(3000),
						Condition: NewCondition(NewUnlockHashCondition(
							unlockHashFromHex("01fc8714235d549f890f35e52d745b9eeeee34926f96c4b9ef1689832f338d9349b453898f7e51"))),
					},
				},
				MinerFees: nil,
			},
		},
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
			Transaction{
				Version: TransactionVersionOne,
				CoinInputs: []CoinInput{
					{
						ParentID: CoinOutputID(hs("1100000000000000000000000000000000000000000000000000000000000011")),
						Fulfillment: NewFulfillment(&SingleSignatureFulfillment{
							PublicKey: PublicKey{
								Algorithm: SignatureAlgoEd25519,
								Key:       hbs("def123def123def123def123def123def123def123def123def123def123def1"),
							},
							Signature: hbs("ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef"),
						}),
					},
				},
				CoinOutputs: []CoinOutput{
					{
						Value:     NewCurrency64(9),
						Condition: NewCondition(&NilCondition{}),
					},
				},
				MinerFees: []Currency{NewCurrency64(3)},
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
			Transaction{
				Version: TransactionVersionOne,
				CoinInputs: []CoinInput{
					{
						ParentID: CoinOutputID(hs("0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")),
						Fulfillment: NewFulfillment(&SingleSignatureFulfillment{
							PublicKey: PublicKey{
								Algorithm: SignatureAlgoEd25519,
								Key:       hbs("def123def123def123def123def123def123def123def123def123def123def1"),
							},
							Signature: hbs("ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef"),
						}),
					},
				},
				MinerFees:     []Currency{NewCurrency64(1)},
				ArbitraryData: ArbitraryData{Data: []byte("Hello, World!")},
			},
		},
		{
			`{
	"version": 1,
	"data": {
		"coininputs": [
			{
				"parentid": "abcdef012345abcdef012345abcdef012345abcdef012345abcdef012345abcd",
				"fulfillment": {
					"type": 1,
					"data": {
						"publickey": "ed25519:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
						"signature": "abcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefab"
					}
				}
			},
			{
				"parentid": "012345defabc012345defabc012345defabc012345defabc012345defabc0123",
				"fulfillment": {
					"type": 2,
					"data": {
						"publickey": "ed25519:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
						"signature": "abcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefab",
						"secret": "def789def789def789def789def789dedef789def789def789def789def789de"
					}
				}
			},
			{
				"parentid": "045645defabc012345defabc012345defabc012345defabc012345defabc0123",
				"fulfillment": {
					"type": 2,
					"data": {
						"sender": "01654f96b317efe5fd6cd8ba1a394dce7b6ebe8c9621d6c44cbe3c8f1b58ce632a3216de71b23b",
						"receiver": "01e89843e4b8231a01ba18b254d530110364432aafab8206bea72e5a20eaa55f70b1ccc65e2105",
						"hashedsecret": "abc543defabc543defabc543defabc543defabc543defabc543defabc543defa",
						"timelock": 1522068743,
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
				"condition": {
					"type": 1,
					"data": {
						"unlockhash": "0142e9458e348598111b0bc19bda18e45835605db9f4620616d752220ae8605ce0df815fd7570e"
					}
				}
			},
			{
				"value": "5",
				"condition": {
					"type": 1,
					"data": {
						"unlockhash": "01a6a6c5584b2bfbd08738996cd7930831f958b9a5ed1595525236e861c1a0dc353bdcf54be7d8"
					}
				}
			},
			{
				"value": "8",
				"condition": {
					"type": 1,
					"data": {
						"unlockhash": "02a24c97c80eeac111aa4bcbb0ac8ffc364fa9b22da10d3054778d2332f68b365e5e5af8e71541"
					}
				}
			},
			{
				"value": "13",
				"condition": {
					"type": 2,
					"data": {
						"sender": "01654f96b317efe5fd6cd8ba1a394dce7b6ebe8c9621d6c44cbe3c8f1b58ce632a3216de71b23b",
						"receiver": "01e89843e4b8231a01ba18b254d530110364432aafab8206bea72e5a20eaa55f70b1ccc65e2105",
						"hashedsecret": "abc543defabc543defabc543defabc543defabc543defabc543defabc543defa",
						"timelock": 1522068743
					}
				}
			}
		],
		"blockstakeinputs": [
			{
				"parentid": "dfd23dfd23dfd23dfd23dfd23dfd23dfd23dfd23dfd23dfd23dfd23dfd23dfde",
				"fulfillment": {
					"type": 1,
					"data": {
						"publickey": "ed25519:ef1234ef1234ef1234ef1234ef1234ef1234ef1234ef1234ef1234ef1234ef12",
						"signature": "01234def01234def01234def01234def01234def01234def01234def01234def01234def01234def01234def01234def01234def01234def01234def01234def"
					}
				}
			}
		],
		"blockstakeoutputs": [
			{
				"value": "4",
				"condition": {
					"type": 1,
					"data": {
						"unlockhash": "6453402d094ed0f336950c4be0feec37167aaaaf8b974d265900e49ab22773584cfe96393b1360"
					}
				}
			},
			{
				"value": "2",
				"condition": {
					"type": 1,
					"data": {
						"unlockhash": "2ab39baa9a58319fa47f78ed542a733a7198d106caeabf0a231b91ea3e4e222ffd8b27c861beff"
					}
				}
			}
		],
		"minerfees": ["1", "2", "3"],
		"arbitrarydata": "ZGF0YQ=="
	}
}`,
			Transaction{
				Version: TransactionVersionOne,
				CoinInputs: []CoinInput{
					{
						ParentID: CoinOutputID(hs("abcdef012345abcdef012345abcdef012345abcdef012345abcdef012345abcd")),
						Fulfillment: NewFulfillment(&SingleSignatureFulfillment{
							PublicKey: PublicKey{
								Algorithm: SignatureAlgoEd25519,
								Key:       hbs("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"),
							},
							Signature: hbs("abcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefab"),
						}),
					},
					{
						ParentID: CoinOutputID(hs("012345defabc012345defabc012345defabc012345defabc012345defabc0123")),
						Fulfillment: NewFulfillment(&anyAtomicSwapFulfillment{
							atomicSwapFulfillment: &AtomicSwapFulfillment{
								PublicKey: PublicKey{
									Algorithm: SignatureAlgoEd25519,
									Key:       hbs("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"),
								},
								Signature: hbs("abcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefab"),
								Secret:    AtomicSwapSecret(hs("def789def789def789def789def789dedef789def789def789def789def789de")),
							},
						}),
					},
					{
						ParentID: CoinOutputID(hs("045645defabc012345defabc012345defabc012345defabc012345defabc0123")),
						Fulfillment: NewFulfillment(&anyAtomicSwapFulfillment{
							atomicSwapFulfillment: &LegacyAtomicSwapFulfillment{
								Sender: UnlockHash{
									Type: UnlockTypePubKey,
									Hash: hs("654f96b317efe5fd6cd8ba1a394dce7b6ebe8c9621d6c44cbe3c8f1b58ce632a"),
								},
								Receiver: UnlockHash{
									Type: UnlockTypePubKey,
									Hash: hs("e89843e4b8231a01ba18b254d530110364432aafab8206bea72e5a20eaa55f70"),
								},
								HashedSecret: AtomicSwapHashedSecret(hs("abc543defabc543defabc543defabc543defabc543defabc543defabc543defa")),
								TimeLock:     1522068743,
								PublicKey: PublicKey{
									Algorithm: SignatureAlgoEd25519,
									Key:       hbs("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"),
								},
								Signature: hbs("abcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefab"),
								Secret:    AtomicSwapSecret(hs("def789def789def789def789def789dedef789def789def789def789def789de")),
							},
						}),
					},
				},
				CoinOutputs: []CoinOutput{
					{
						Value: NewCurrency64(3),
						Condition: NewCondition(NewUnlockHashCondition(UnlockHash{
							Type: UnlockTypePubKey,
							Hash: hs("42e9458e348598111b0bc19bda18e45835605db9f4620616d752220ae8605ce0"),
						})),
					},
					{
						Value: NewCurrency64(5),
						Condition: NewCondition(NewUnlockHashCondition(UnlockHash{
							Type: UnlockTypePubKey,
							Hash: hs("a6a6c5584b2bfbd08738996cd7930831f958b9a5ed1595525236e861c1a0dc35"),
						})),
					},
					{
						Value: NewCurrency64(8),
						Condition: NewCondition(NewUnlockHashCondition(UnlockHash{
							Type: UnlockTypeAtomicSwap,
							Hash: hs("a24c97c80eeac111aa4bcbb0ac8ffc364fa9b22da10d3054778d2332f68b365e"),
						})),
					},
					{
						Value: NewCurrency64(13),
						Condition: NewCondition(&AtomicSwapCondition{
							Sender: UnlockHash{
								Type: UnlockTypePubKey,
								Hash: hs("654f96b317efe5fd6cd8ba1a394dce7b6ebe8c9621d6c44cbe3c8f1b58ce632a"),
							},
							Receiver: UnlockHash{
								Type: UnlockTypePubKey,
								Hash: hs("e89843e4b8231a01ba18b254d530110364432aafab8206bea72e5a20eaa55f70"),
							},
							HashedSecret: AtomicSwapHashedSecret(hs("abc543defabc543defabc543defabc543defabc543defabc543defabc543defa")),
							TimeLock:     1522068743,
						}),
					},
				},
				BlockStakeInputs: []BlockStakeInput{
					{
						ParentID: BlockStakeOutputID(hs("dfd23dfd23dfd23dfd23dfd23dfd23dfd23dfd23dfd23dfd23dfd23dfd23dfde")),
						Fulfillment: NewFulfillment(&SingleSignatureFulfillment{
							PublicKey: PublicKey{
								Algorithm: SignatureAlgoEd25519,
								Key:       hbs("ef1234ef1234ef1234ef1234ef1234ef1234ef1234ef1234ef1234ef1234ef12"),
							},
							Signature: hbs("01234def01234def01234def01234def01234def01234def01234def01234def01234def01234def01234def01234def01234def01234def01234def01234def"),
						}),
					},
				},
				BlockStakeOutputs: []BlockStakeOutput{
					{
						Value: NewCurrency64(4),
						Condition: NewCondition(NewUnlockHashCondition(UnlockHash{
							Type: 100,
							Hash: hs("53402d094ed0f336950c4be0feec37167aaaaf8b974d265900e49ab22773584c"),
						})),
					},
					{
						Value: NewCurrency64(2),
						Condition: NewCondition(NewUnlockHashCondition(UnlockHash{
							Type: 42,
							Hash: hs("b39baa9a58319fa47f78ed542a733a7198d106caeabf0a231b91ea3e4e222ffd"),
						})),
					},
				},
				MinerFees: []Currency{
					NewCurrency64(1), NewCurrency64(2), NewCurrency64(3),
				},
				ArbitraryData: ArbitraryData{Data: []byte("data")},
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
			Transaction{
				Version: TransactionVersionOne,
				CoinInputs: []CoinInput{
					{
						ParentID: CoinOutputID(hs("1100000000000000000000000000000000000000000000000000000000000011")),
						Fulfillment: NewFulfillment(&SingleSignatureFulfillment{
							PublicKey: PublicKey{
								Algorithm: SignatureAlgoEd25519,
								Key:       hbs("def123def123def123def123def123def123def123def123def123def123def1"),
							},
							Signature: hbs("ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef"),
						}),
					},
				},
				CoinOutputs: []CoinOutput{
					{
						Value: NewCurrency64(9),
						Condition: NewCondition(&TimeLockCondition{
							LockTime:  42,
							Condition: &NilCondition{},
						}),
					},
				},
				MinerFees: []Currency{NewCurrency64(3)},
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
			Transaction{
				Version: TransactionVersionOne,
				CoinInputs: []CoinInput{
					{
						ParentID: CoinOutputID(hs("1100000000000000000000000000000000000000000000000000000000000011")),
						Fulfillment: NewFulfillment(&SingleSignatureFulfillment{
							PublicKey: PublicKey{
								Algorithm: SignatureAlgoEd25519,
								Key:       hbs("def123def123def123def123def123def123def123def123def123def123def1"),
							},
							Signature: hbs("ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef"),
						}),
					},
				},
				CoinOutputs: []CoinOutput{
					{
						Value: NewCurrency64(9),
						Condition: NewCondition(&TimeLockCondition{
							LockTime: 42,
							Condition: &UnlockHashCondition{
								TargetUnlockHash: UnlockHash{
									Type: UnlockTypePubKey,
									Hash: hs("e89843e4b8231a01ba18b254d530110364432aafab8206bea72e5a20eaa55f70"),
								},
							},
						}),
					},
				},
				MinerFees: []Currency{NewCurrency64(3)},
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
			Transaction{
				Version: TransactionVersionOne,
				CoinInputs: []CoinInput{
					{
						ParentID: CoinOutputID(hs("1100000000000000000000000000000000000000000000000000000000000011")),
						Fulfillment: NewFulfillment(&SingleSignatureFulfillment{
							PublicKey: PublicKey{
								Algorithm: SignatureAlgoEd25519,
								Key:       hbs("def123def123def123def123def123def123def123def123def123def123def1"),
							},
							Signature: hbs("ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef"),
						}),
					},
				},
				CoinOutputs: []CoinOutput{
					{
						Value: NewCurrency64(9),
						Condition: NewCondition(&MultiSignatureCondition{
							MinimumSignatureCount: 2,
							UnlockHashes: UnlockHashSlice{
								UnlockHash{
									Type: UnlockTypePubKey,
									Hash: hs("e89843e4b8231a01ba18b254d530110364432aafab8206bea72e5a20eaa55f70"),
								},
								UnlockHash{
									Type: UnlockTypePubKey,
									Hash: hs("a6a6c5584b2bfbd08738996cd7930831f958b9a5ed1595525236e861c1a0dc35"),
								},
							},
						}),
					},
				},
				MinerFees: []Currency{NewCurrency64(3)},
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
			Transaction{
				Version: TransactionVersionOne,
				CoinInputs: []CoinInput{
					{
						ParentID: CoinOutputID(hs("1100000000000000000000000000000000000000000000000000000000000011")),
						Fulfillment: NewFulfillment(&MultiSignatureFulfillment{
							Pairs: []PublicKeySignaturePair{
								{
									PublicKey: PublicKey{
										Algorithm: SignatureAlgoEd25519,
										Key:       hbs("def123def123def123def123def123def123def123def123def123def123def1"),
									},
									Signature: hbs("ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef"),
								},
								{
									PublicKey: PublicKey{
										Algorithm: SignatureAlgoEd25519,
										Key:       hbs("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"),
									},
									Signature: hbs("abcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefab"),
								},
							},
						}),
					},
				},
				CoinOutputs: []CoinOutput{
					{
						Value: NewCurrency64(9),
						Condition: NewCondition(&TimeLockCondition{
							LockTime: 42,
							Condition: &UnlockHashCondition{
								TargetUnlockHash: UnlockHash{
									Type: UnlockTypePubKey,
									Hash: hs("e89843e4b8231a01ba18b254d530110364432aafab8206bea72e5a20eaa55f70"),
								},
							},
						}),
					},
				},
				MinerFees: []Currency{NewCurrency64(3)},
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

func TestTransactionWithUnknownVersionJSONEncoding(t *testing.T) {
	const str = `{"version":42,"data":"aGVsbG8sIHdvcmxk"}`
	var txn Transaction
	err := json.Unmarshal([]byte(str), &txn)
	if err == nil {
		t.Fatal("txn with unknown version shouldn't be able to be decoded")
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
	testCases := []string{
		// transactions with unknown transaction versions
		`2a170000000000000048656c6c6f2c20526177205472616e73616374696f6e21`,
	}
	for idx, inputHexTxn := range testCases {
		// sanity check to ensure our hex is valid
		encodedTx, err := hex.DecodeString(inputHexTxn)
		if err != nil {
			t.Error(idx, err)
			continue
		}
		var tx Transaction
		err = tx.UnmarshalSia(bytes.NewReader(encodedTx))
		if err == nil {
			t.Error(idx, "expected error, but none received")
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

// unit test for https://github.com/threefoldfoundation/tfchain/issues/134
func TestIDComputationCompatibleWithLegacyIDs_TFChain_Issue134(t *testing.T) {
	type IDIndexPair struct {
		Index    uint64
		ParentID string
	}
	testCases := []struct {
		JSONEncodedTransaction string
		HexTransactionID       string
		CoinParents            []IDIndexPair
		BlockStakeParents      []IDIndexPair
	}{
		{
			`{
	"version": 1,
	"data": {
		"coininputs": null,
		"blockstakeinputs": [{
			"parentid": "782e4819d6e199856ba1bff3def5d7cc37ae2a0dabecb05359d6072156190d68",
			"fulfillment": {
				"type": 1,
				"data": {
					"publickey": "ed25519:95990ca3774de81309932302f74dfe9e540d6c29ca5cb9ee06e999ad46586737",
					"signature": "70be2115b82a54170c94bf4788e2a6dd154a081f61e97999c2d9fcc64c41e7df2e8a8d4f82a57a04a1247b9badcb6bffbd238e9a6761dd59e5fef7ff6df0fc01"
				}
			}
		}],
		"blockstakeoutputs": [{
			"value": "99",
			"condition": {
				"type": 1,
				"data": {
					"unlockhash": "01fdf10836c119186f1d21666ae2f7dc62d6ecc46b5f41449c3ee68aea62337dad917808e46799"
				}
			}
		}],
		"minerfees": null
	}
}`,
			"af2e02682fb8423e4c9060d93b847b7e26b2a2b90f999b12a116605fa5358195",
			nil,
			[]IDIndexPair{
				{0, "107a0a18a511cf1a6d4a2ac1f1a3626a7aa6004ef7aa21ac684b63b95594c6c1"},
			},
		},
		{
			`{
	"version": 1,
	"data": {
		"coininputs": [{
			"parentid": "5b907d6e4d34cdd825484d2f9f14445377fb8b4f8cab356a390a7fe4833a3085",
			"fulfillment": {
				"type": 1,
				"data": {
					"publickey": "ed25519:bd5e0e345d5939f5f9eb330084c7f0ffb8fc7fc5bdb07a94c304620eb4e2d99a",
					"signature": "55dace7ccbc9cdd23220a8ef3ec09e84ce5c5acc202c5f270ea0948743ebf52135f3936ef7477170b4f9e0fe141a61d8312ab31afbf926a162982247e5d2720a"
				}
			}
		}],
		"coinoutputs": [{
			"value": "1000000000",
			"condition": {
				"type": 1,
				"data": {
					"unlockhash": "010009a2b6a482da73204ccc586f6fab5504a1a69c0d316cdf828a476ae7c91c9137fd6f1b62bb"
				}
			}
		}, {
			"value": "8900000000",
			"condition": {
				"type": 1,
				"data": {
					"unlockhash": "01b81f9e02d6be3a7de8440365a7c799e07dedf2ccba26fd5476c304e036b87c1ab716558ce816"
				}
			}
		}],
		"minerfees": ["100000000"]
	}
}`,
			"943778f4e2ecb5b2186e05c989b7c08e9d0199d419509725e166ac7f221600c7",
			[]IDIndexPair{
				{1, "733b266749141bad0900d7f9fe69dd511a98f7a6f4c78e66aa41af7f55650a15"},
			},
			nil,
		},
	}
	for idx, testCase := range testCases {
		var txn Transaction
		err := txn.UnmarshalJSON([]byte(testCase.JSONEncodedTransaction))
		if err != nil {
			t.Error(idx, "failed to decode JSON txn", err)
			continue
		}
		var txnID TransactionID
		err = txnID.LoadString(testCase.HexTransactionID)
		if err != nil {
			t.Error(idx, "failed to decode hex encoded txnID", err)
			continue
		}
		// test transaction ID
		if txn.Version == TransactionVersionZero {
			// ensure our legacy ID is correct for v0 transactions
			legacyID := txn.LegacyID()
			if legacyID != txnID {
				t.Error(idx, "LegacyID", legacyID, "!=", txnID)
			}
		}
		// sanity check, to ensure our expected ID is actually correct
		v106ID := txn.V106ID()
		if v106ID != txnID {
			t.Error(idx, "V106ID", v106ID, "!=", txnID)
		}
		// ensure our computed ID is correct
		transactionID := txn.ID()
		if transactionID != txnID {
			t.Error(idx, "TransactionID", transactionID, "!=", txnID)
		}
		// test coin parent IDs
		for udx, parent := range testCase.CoinParents {
			var parentID CoinOutputID
			err = parentID.LoadString(parent.ParentID)
			if err != nil {
				t.Error(idx, udx, "failed to decode coin parentID", err)
				continue
			}
			if txn.Version == TransactionVersionZero {
				// ensure our legacy ID is correct for v0 transactions
				legacyOutputID := txn.LegacyCoinOutputID(parent.Index)
				if legacyOutputID != parentID {
					t.Error(idx, udx, "LegacyCoinOutputID", legacyOutputID, "!=", parentID)
				}
			}
			// sanity check, to ensure our expected ID is actually correct
			v106OutputID := txn.V106CoinOutputID(parent.Index)
			if v106OutputID != parentID {
				t.Error(idx, udx, "V106CoinOutputID", v106OutputID, "!=", parentID)
			}
			// ensure our computed ID is correct
			outputID := txn.CoinOutputID(parent.Index)
			if outputID != parentID {
				t.Error(idx, udx, "CoinOutputID", outputID, "!=", parentID)
			}
		}
		// test block stake parent IDs
		for udx, parent := range testCase.BlockStakeParents {
			var parentID BlockStakeOutputID
			err = parentID.LoadString(parent.ParentID)
			if err != nil {
				t.Error(idx, udx, "failed to decode block stake parentID", err)
				continue
			}
			if txn.Version == TransactionVersionZero {
				// ensure our legacy ID is correct for v0 transactions
				legacyOutputID := txn.LegacyBlockStakeOutputID(parent.Index)
				if legacyOutputID != parentID {
					t.Error(idx, udx, "LegacyBlockStakeOutputID", legacyOutputID, "!=", parentID)
				}
			}
			// sanity check, to ensure our expected ID is actually correct
			v106OutputID := txn.V106BlockstakeOutputID(parent.Index)
			if v106OutputID != parentID {
				t.Error(idx, udx, "V106BlockstakeOutputID", v106OutputID, "!=", parentID)
			}
			// ensure our computed ID is correct
			outputID := txn.BlockStakeOutputID(parent.Index)
			if outputID != parentID {
				t.Error(idx, udx, "BlockStakeOutputID", outputID, "!=", parentID)
			}
		}
	}
}

// ID returns the id of a transaction, which is taken by marshalling all of the
// fields except for the signatures and taking the hash of the result.
func (t Transaction) LegacyID() TransactionID {
	ltd, err := newLegacyTransactionDataFromTransaction(t)
	if err != nil {
		panic(err)
	}
	return TransactionID(crypto.HashAll(
		ltd.CoinInputs,
		ltd.CoinOutputs,
		ltd.BlockStakeInputs,
		ltd.BlockStakeOutputs,
		ltd.MinerFees,
		ltd.ArbitraryData,
	))
}

// CoinOutputID returns the ID of a coin output at the given index,
// which is calculated by hashing the concatenation of the CoinOutput
// Specifier, all of the fields in the transaction (except the signatures),
// and output index.
func (t Transaction) LegacyCoinOutputID(i uint64) CoinOutputID {
	ltd, err := newLegacyTransactionDataFromTransaction(t)
	if err != nil {
		panic(err)
	}
	return CoinOutputID(crypto.HashAll(
		SpecifierCoinOutput,
		ltd.CoinInputs,
		ltd.CoinOutputs,
		ltd.BlockStakeInputs,
		ltd.BlockStakeOutputs,
		ltd.MinerFees,
		ltd.ArbitraryData,
		i,
	))
}

// sanity checks for TestIDComputationCompatibleWithLegacyIDs_TFChain_Issue134
func (t Transaction) V106ID() (id TransactionID) {
	if t.Version == TransactionVersionZero {
		ltd, err := newLegacyTransactionDataFromTransaction(t)
		if build.DEBUG && err != nil {
			panic(err)
		}
		// the legacy version does not include the transaction version
		// as part of the crypto hash
		return TransactionID(crypto.HashObject(ltd))
	}
	h := crypto.NewHash()
	t.MarshalSia(h)
	h.Sum(id[:0])
	return
}
func (t Transaction) V106CoinOutputID(i uint64) CoinOutputID {
	if t.Version == TransactionVersionZero {
		ltd, err := newLegacyTransactionDataFromTransaction(t)
		if build.DEBUG && err != nil {
			panic(err)
		}
		// the legacy version does not include the transaction version
		// as part of the crypto hash
		return CoinOutputID(crypto.HashAll(
			SpecifierCoinOutput,
			ltd,
			i,
		))
	}
	return CoinOutputID(crypto.HashAll(
		SpecifierCoinOutput,
		t,
		i,
	))
}
func (t Transaction) V106BlockstakeOutputID(i uint64) BlockStakeOutputID {
	if t.Version == TransactionVersionZero {
		ltd, err := newLegacyTransactionDataFromTransaction(t)
		if build.DEBUG && err != nil {
			panic(err)
		}
		// the legacy version does not include the transaction version
		// as part of the crypto hash
		return BlockStakeOutputID(crypto.HashAll(
			SpecifierBlockStakeOutput,
			ltd,
			i,
		))
	}
	return BlockStakeOutputID(crypto.HashAll(
		SpecifierBlockStakeOutput,
		t,
		i,
	))
}

// BlockStakeOutputID returns the ID of a BlockStakeOutput at the given index, which
// is calculated by hashing the concatenation of the BlockStakeOutput Specifier,
// all of the fields in the transaction (except the signatures), and output
// index.
func (t Transaction) LegacyBlockStakeOutputID(i uint64) BlockStakeOutputID {

	ltd, err := newLegacyTransactionDataFromTransaction(t)
	if err != nil {
		panic(err)
	}
	return BlockStakeOutputID(crypto.HashAll(
		SpecifierBlockStakeOutput,
		ltd.CoinInputs,
		ltd.CoinOutputs,
		ltd.BlockStakeInputs,
		ltd.BlockStakeOutputs,
		ltd.MinerFees,
		ltd.ArbitraryData,
		i,
	))
}

func TestIsValidTransactionVersion(t *testing.T) {
	minVersion, maxVersion := TransactionVersion(0), TransactionVersion(1)
	for v := minVersion; v <= maxVersion; v++ {
		err := v.IsValidTransactionVersion()
		if err != nil {
			t.Error("unexpected invalid version", v, err)
		}
	}
	err := (maxVersion + 1).IsValidTransactionVersion()
	if err == nil {
		t.Error("nknown version should be valid, while it is not:", maxVersion+1)
	}
}
