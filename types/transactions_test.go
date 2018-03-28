package types

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"reflect"
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
