package types

// signatures.go contains all of the types and functions related to creating
// and verifying transaction signatures. There are a lot of rules surrounding
// the correct use of signatures. Signatures can cover part or all of a
// transaction, can be multiple different algorithms, and must satify a field
// called 'UnlockConditions'.

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"

	"github.com/rivine/rivine/crypto"
	"github.com/rivine/rivine/encoding"
)

var (
	// These Specifiers enumerate the types of signatures that are recognized
	// by this implementation. If a signature's type is unrecognized, the
	// signature is treated as valid. Signatures using the special "entropy"
	// type are always treated as invalid; see Consensus.md for more details.
	SignatureEntropy = Specifier{'e', 'n', 't', 'r', 'o', 'p', 'y'}
	SignatureEd25519 = Specifier{'e', 'd', '2', '5', '5', '1', '9'}

	ErrEntropyKey                = errors.New("transaction tries to sign an entproy public key")
	ErrFrivolousSignature        = errors.New("transaction contains a frivolous signature")
	ErrInvalidPubKeyIndex        = errors.New("transaction contains a signature that points to a nonexistent public key")
	ErrInvalidUnlockHashChecksum = errors.New("provided unlock hash has an invalid checksum")
	ErrMissingSignatures         = errors.New("transaction has inputs with missing signatures")
	ErrPrematureSignature        = errors.New("timelock on signature has not expired")
	ErrPublicKeyOveruse          = errors.New("public key was used multiple times while signing transaction")
	ErrSortedUniqueViolation     = errors.New("sorted unique violation")
	ErrUnlockHashWrongLen        = errors.New("marshalled unlock hash is the wrong length")
)

type (
	// A SiaPublicKey is a public key prefixed by a Specifier. The Specifier
	// indicates the algorithm used for signing and verification. Unrecognized
	// algorithms will always verify, which allows new algorithms to be added to
	// the protocol via a soft-fork.
	SiaPublicKey struct {
		Algorithm Specifier
		Key       ByteSlice
	}

	// ByteSlice defines any kind of raw binary value,
	// in-memory defined as a byte slice,
	// and JSON-encoded in hexadecimal form.
	ByteSlice []byte
)

// Ed25519PublicKey returns pk as a SiaPublicKey, denoting its algorithm as
// Ed25519.
func Ed25519PublicKey(pk crypto.PublicKey) SiaPublicKey {
	return SiaPublicKey{
		Algorithm: SignatureEd25519,
		Key:       pk[:],
	}
}

// InputSigHash returns the hash of all fields in a transaction,
// relevant to an input sig.
func (t Transaction) InputSigHash(inputIndex uint64, extraObjects ...interface{}) (hash crypto.Hash) {
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

// sortedUnique checks that 'elems' is sorted, contains no repeats, and that no
// element is larger than or equal to 'max'.
func sortedUnique(elems []uint64, max int) bool {
	if len(elems) == 0 {
		return true
	}

	biggest := elems[0]
	for _, elem := range elems[1:] {
		if elem <= biggest {
			return false
		}
		biggest = elem
	}
	if biggest >= uint64(max) {
		return false
	}
	return true
}

// validSignatures checks the validaty of all signatures in a transaction.
func (t *Transaction) validSignatures(currentHeight BlockHeight) (err error) {
	spendCoins := make(map[CoinOutputID]struct{})
	for index, ci := range t.CoinInputs {
		if _, found := spendCoins[ci.ParentID]; found {
			err = ErrDoubleSpend
			return
		}
		spendCoins[ci.ParentID] = struct{}{}
		err = ci.Unlocker.Unlock(uint64(index), *t)
		if err != nil {
			return
		}
	}

	spendBlockStakes := make(map[BlockStakeOutputID]struct{})
	for index, bsi := range t.BlockStakeInputs {
		if _, found := spendBlockStakes[bsi.ParentID]; found {
			err = ErrDoubleSpend
			return
		}
		spendBlockStakes[bsi.ParentID] = struct{}{}
		err = bsi.Unlocker.Unlock(uint64(index), *t)
		if err != nil {
			return
		}
	}

	return
}

// LoadString is the inverse of SiaPublicKey.String().
func (spk *SiaPublicKey) LoadString(s string) error {
	parts := strings.SplitN(s, ":", 2)
	if len(parts) != 2 {
		return errors.New("invalid public key string")
	}
	err := spk.Key.LoadString(parts[1])
	if err != nil {
		return err
	}
	copy(spk.Algorithm[:], []byte(parts[0]))
	return nil
}

// String defines how to print a SiaPublicKey - hex is used to keep things
// compact during logging. The key type prefix and lack of a checksum help to
// separate it from a sia address.
func (spk *SiaPublicKey) String() string {
	return spk.Algorithm.String() + ":" + spk.Key.String()
}

// MarshalJSON marshals a byte slice as a hex string.
func (spk SiaPublicKey) MarshalJSON() ([]byte, error) {
	return json.Marshal(spk.String())
}

// UnmarshalJSON decodes the json string of the byte slice.
func (spk *SiaPublicKey) UnmarshalJSON(b []byte) error {
	var str string
	if err := json.Unmarshal(b, &str); err != nil {
		return err
	}
	return spk.LoadString(str)
}

// String turns this byte slice into a hex-formatted string.
func (bs ByteSlice) String() string {
	return hex.EncodeToString([]byte(bs))
}

// LoadString loads a  byte slice from a hex-formatted string.
func (bs *ByteSlice) LoadString(str string) error {
	b, err := hex.DecodeString(str)
	if err != nil {
		return err
	}
	*bs = ByteSlice(b)
	return nil
}

// MarshalJSON marshals a byte slice as a hex string.
func (bs ByteSlice) MarshalJSON() ([]byte, error) {
	return json.Marshal(bs.String())
}

// UnmarshalJSON decodes the json string of the byte slice.
func (bs *ByteSlice) UnmarshalJSON(b []byte) error {
	var str string
	if err := json.Unmarshal(b, &str); err != nil {
		return err
	}
	return bs.LoadString(str)
}

var (
	_ json.Marshaler   = ByteSlice{}
	_ json.Unmarshaler = (*ByteSlice)(nil)
)
