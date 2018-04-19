package types

import (
	"bytes"
	"reflect"
	"strings"
	"testing"

	"github.com/NebulousLabs/fastrand"
	"github.com/rivine/rivine/crypto"
)

// TestEd25519PublicKey tests the Ed25519PublicKey function.
func TestEd25519PublicKey(t *testing.T) {
	_, pk := crypto.GenerateKeyPair()
	spk := Ed25519PublicKey(pk)
	if spk.Algorithm != SignatureEd25519 {
		t.Error("Ed25519PublicKey created key with wrong algorithm specifier:", spk.Algorithm)
	}
	if !bytes.Equal(spk.Key, pk[:]) {
		t.Error("Ed25519PublicKey created key with wrong data")
	}
}

// TestSigHash runs the SigHash function of the transaction type.
func TestSigHash(t *testing.T) {
	txn := Transaction{
		Version:           DefaultChainConstants().DefaultTransactionVersion,
		CoinInputs:        []CoinInput{{}},
		CoinOutputs:       []CoinOutput{{}},
		BlockStakeInputs:  []BlockStakeInput{{}},
		BlockStakeOutputs: []BlockStakeOutput{{}},
		MinerFees:         []Currency{{}},
		ArbitraryData:     []byte{'o', 't'},
	}

	_ = txn.InputSigHash(0)
}

// TestSortedUnique probes the sortedUnique function.
func TestSortedUnique(t *testing.T) {
	su := []uint64{3, 5, 6, 8, 12}
	if !sortedUnique(su, 13) {
		t.Error("sortedUnique rejected a valid array")
	}
	if sortedUnique(su, 12) {
		t.Error("sortedUnique accepted an invalid max")
	}
	if sortedUnique(su, 11) {
		t.Error("sortedUnique accepted an invalid max")
	}

	unsorted := []uint64{3, 5, 3}
	if sortedUnique(unsorted, 6) {
		t.Error("sortedUnique accepted an unsorted array")
	}

	repeats := []uint64{2, 4, 4, 7}
	if sortedUnique(repeats, 8) {
		t.Error("sortedUnique accepted an array with repeats")
	}

	bothFlaws := []uint64{2, 3, 4, 5, 6, 6, 4}
	if sortedUnique(bothFlaws, 7) {
		t.Error("Sorted unique accetped array with multiple flaws")
	}
}

func TestByteSliceStringify(t *testing.T) {
	testCases := []struct {
		ByteSlice ByteSlice
		String    string
	}{
		{ByteSlice{}, ""},
		{ByteSlice{0}, "00"},
		{ByteSlice{42}, "2a"},
		{ByteSlice{255}, "ff"},
		{ByteSlice{0, 255, 0}, "00ff00"},
		{ByteSlice{1, 2, 3}, "010203"},
		{ByteSlice{2, 5, 5}, "020505"},
		{ByteSlice{0, 0, 0, 0}, "00000000"},
	}
	for index, testCase := range testCases {
		str := testCase.ByteSlice.String()
		if str != testCase.String {
			t.Errorf("stringification went wrong: %q != %q", str, testCase.String)
		}
		var bs ByteSlice
		err := bs.LoadString(str)
		if err != nil {
			t.Errorf("destringification of #%d went wrong: %v", index, err)
		}
		if !reflect.DeepEqual(bs, testCase.ByteSlice) {
			t.Errorf("destringification of #%d went wrong: %v != %v", index, bs, testCase.ByteSlice)
		}
	}
}

// TestSiaPublicKeyLoadString checks that the LoadString method is the proper
// inverse of the String() method, also checks that there are no stupid panics
// or severe errors.
func TestSiaPublicKeyLoadString(t *testing.T) {
	spk := SiaPublicKey{
		Algorithm: SignatureEd25519,
		Key:       fastrand.Bytes(32),
	}

	spkString := spk.String()
	var loadedSPK SiaPublicKey
	loadedSPK.LoadString(spkString)
	if !bytes.Equal(loadedSPK.Algorithm[:], spk.Algorithm[:]) {
		t.Error("SiaPublicKey is not loading correctly")
	}
	if !bytes.Equal(loadedSPK.Key, spk.Key) {
		t.Log(loadedSPK.Key, spk.Key)
		t.Error("SiaPublicKey is not loading correctly")
	}

	// Try loading crappy strings.
	parts := strings.Split(spkString, ":")
	spk.LoadString(parts[0])
	spk.LoadString(parts[0][1:])
	spk.LoadString(parts[0][:1])
	spk.LoadString(parts[1])
	spk.LoadString(parts[1][1:])
	spk.LoadString(parts[1][:1])
	spk.LoadString(parts[0] + parts[1])

}

// TestSiaPublicKeyString does a quick check to verify that the String method
// on the SiaPublicKey is producing the expected output.
func TestSiaPublicKeyString(t *testing.T) {
	spk := SiaPublicKey{
		Algorithm: SignatureEd25519,
		Key:       make([]byte, 32),
	}

	if spk.String() != "ed25519:0000000000000000000000000000000000000000000000000000000000000000" {
		t.Error("got wrong value for spk.String():", spk.String())
	}
}
