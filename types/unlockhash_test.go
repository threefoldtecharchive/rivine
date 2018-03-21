package types

import (
	"bytes"
	"encoding/json"
	"reflect"
	"sort"
	"testing"

	"github.com/rivine/rivine/crypto"
)

// TestUnlockHashJSONMarshalling checks that when an unlock hash is marshalled
// and unmarshalled using JSON, the result is what is expected.
func TestUnlockHashJSONMarshalling(t *testing.T) {
	_, pk := crypto.GenerateKeyPair()
	uc := NewSingleSignatureInputLock(Ed25519PublicKey(pk))
	uh := uc.UnlockHash()

	// Marshal the unlock hash.
	marUH, err := json.Marshal(uh)
	if err != nil {
		t.Fatal(err)
	}

	// Unmarshal the unlock hash and compare to the original.
	var umarUH UnlockHash
	err = json.Unmarshal(marUH, &umarUH)
	if err != nil {
		t.Fatal(err)
	}
	if umarUH != uh {
		t.Error("Marshalled and unmarshalled unlock hash are not equivalent")
	}

	// Corrupt the checksum.
	old := marUH[36]
	if old < 102 {
		marUH[36]++
	} else {
		marUH[36] = 97
	}
	err = umarUH.UnmarshalJSON(marUH)
	if err != ErrInvalidUnlockHashChecksum {
		t.Error("expecting an invalid checksum:", err)
	}
	marUH[36] = old

	// Try an input that's not correct hex.
	marUH[7] += 100
	err = umarUH.UnmarshalJSON(marUH)
	if err == nil {
		t.Error("Expecting error after corrupting input")
	}
	marUH[7] -= 100

	// Try an input of the wrong length.
	err = (&umarUH).UnmarshalJSON(marUH[2:])
	if err != ErrUnlockHashWrongLen {
		t.Error("Got wrong error:", err)
	}
}

// TestUnlockHashStringMarshalling checks that when an unlock hash is
// marshalled and unmarshalled using String and LoadString, the result is what
// is expected.
func TestUnlockHashStringMarshalling(t *testing.T) {
	_, pk := crypto.GenerateKeyPair()
	uc := NewSingleSignatureInputLock(Ed25519PublicKey(pk))
	uh := uc.UnlockHash()

	// Marshal the unlock hash.
	marUH := uh.String()

	// Unmarshal the unlock hash and compare to the original.
	var umarUH UnlockHash
	err := umarUH.LoadString(marUH)
	if err != nil {
		t.Fatal(err)
	}
	if umarUH != uh {
		t.Error("Marshalled and unmarshalled unlock hash are not equivalent")
	}

	// Corrupt the checksum.
	byteMarUH := []byte(marUH)
	byteMarUH[36]++
	err = umarUH.LoadString(string(byteMarUH))
	if err != ErrInvalidUnlockHashChecksum {
		t.Error("expecting an invalid checksum:", err)
	}
	byteMarUH[36]--

	// Try an input that's not correct hex.
	byteMarUH[7] += 100
	err = umarUH.LoadString(string(byteMarUH))
	if err == nil {
		t.Error("Expecting error after corrupting input")
	}
	byteMarUH[7] -= 100

	// Try an input of the wrong length.
	err = umarUH.LoadString(string(byteMarUH[2:]))
	if err != ErrUnlockHashWrongLen {
		t.Error("Got wrong error:", err)
	}
}

// TestUnlockHashSliceSorting checks that the sort method correctly sorts
// unlock hash slices.
func TestUnlockHashSliceSorting(t *testing.T) {
	// To test that byte-order is done correctly, a semi-random second byte is
	// used that is equal to the first byte * 23 % 7
	uhs := UnlockHashSlice{
		UnlockHash{0, crypto.Hash{4, 1}},
		UnlockHash{0, crypto.Hash{0, 0}},
		UnlockHash{1, crypto.Hash{1, 2}},
		UnlockHash{0, crypto.Hash{2, 4}},
		UnlockHash{0, crypto.Hash{3, 6}},
		UnlockHash{1, crypto.Hash{0, 0}},
		UnlockHash{0, crypto.Hash{1, 2}},
		UnlockHash{1, crypto.Hash{2, 4}},
		UnlockHash{1, crypto.Hash{4, 1}},
		UnlockHash{1, crypto.Hash{3, 6}},
	}
	sort.Sort(uhs)
	for i := byte(0); i < 5; i++ {
		if uhs[i] != (UnlockHash{0, crypto.Hash{i, (i * 23) % 7}}) {
			t.Error("sorting failed on index", i, uhs[i])
		}
		if uhs[i+5] != (UnlockHash{1, crypto.Hash{i, (i * 23) % 7}}) {
			t.Error("sorting failed on index", i+5, uhs[i+5])
		}
	}
}

func TestUnlockHashSiaMarshalling(t *testing.T) {
	testCases := []struct {
		UnlockHash UnlockHash
		Expected   []byte
	}{
		{
			UnlockHash{0, crypto.Hash{}},
			[]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		},
		{
			UnlockHash{4, crypto.Hash{2}},
			[]byte{4, 2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		},
		{
			UnlockHash{1, crypto.Hash{4, 2}},
			[]byte{1, 4, 2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		},
		{
			UnlockHash{1, crypto.Hash{2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3}},
			[]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3},
		},
	}
	for idx, testCase := range testCases {
		buf := bytes.NewBuffer(nil)
		err := testCase.UnlockHash.MarshalSia(buf)
		if err != nil {
			t.Errorf("error sia-marshalling #%d: %v", idx, err)
		}
		if bytes.Compare(testCase.Expected, buf.Bytes()) != 0 {
			t.Errorf("unexpected marshalled form of the unlock hash: (%v) != (%v)",
				testCase.Expected, buf.Bytes())
		}
		var uh UnlockHash
		err = uh.UnmarshalSia(buf)
		if err != nil {
			t.Errorf("error sia-unmarshalling #%d: %v", idx, err)
		}
		if !reflect.DeepEqual(testCase.UnlockHash, uh) {
			t.Errorf("unexpected unmarshalled form of the unlock hash: (%v) != (%v)",
				testCase.UnlockHash, uh)

		}
	}
}
