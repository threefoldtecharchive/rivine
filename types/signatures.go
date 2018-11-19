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
	"fmt"
	"strings"

	"github.com/threefoldtech/rivine/build"
	"github.com/threefoldtech/rivine/crypto"
	"github.com/threefoldtech/rivine/pkg/encoding/siabin"
)

var (
	// These Specifiers enumerate the types of signatures that are recognized
	// by this implementation. see Consensus.md for more details.
	SignatureEd25519 = Specifier{'e', 'd', '2', '5', '5', '1', '9'}

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
func (t Transaction) InputSigHash(inputIndex uint64, extraObjects ...interface{}) (crypto.Hash, error) {
	controller, exists := _RegisteredTransactionVersions[t.Version]
	if !exists {
		return crypto.Hash{}, ErrUnknownTransactionType
	}
	if hasher, ok := controller.(InputSigHasher); ok {
		// if extension implements InputSigHasher,
		// use it here to sign the input with it
		return hasher.InputSigHash(t, inputIndex, extraObjects...)
	}

	h := crypto.NewHash()
	enc := siabin.NewEncoder(h)

	enc.EncodeAll(
		t.Version,
		inputIndex,
	)

	if len(extraObjects) > 0 {
		enc.EncodeAll(extraObjects...)
	}
	enc.Encode(len(t.CoinInputs))
	for _, ci := range t.CoinInputs {
		enc.Encode(ci.ParentID)
	}
	enc.Encode(t.CoinOutputs)
	enc.Encode(len(t.BlockStakeInputs))
	for _, bsi := range t.BlockStakeInputs {
		enc.Encode(bsi.ParentID)
	}
	enc.EncodeAll(
		t.BlockStakeOutputs,
		t.MinerFees,
		t.ArbitraryData,
	)

	var hash crypto.Hash
	h.Sum(hash[:0])
	return hash, nil
}

func (t Transaction) legacyInputSigHash(inputIndex uint64, extraObjects ...interface{}) crypto.Hash {
	h := crypto.NewHash()
	enc := siabin.NewEncoder(h)

	enc.Encode(inputIndex)
	if len(extraObjects) > 0 {
		enc.EncodeAll(extraObjects...)
	}
	for _, ci := range t.CoinInputs {
		enc.EncodeAll(ci.ParentID, legacyUnlockHashFromFulfillment(ci.Fulfillment.Fulfillment))
	}
	// legacy transactions encoded unlock hashes in pure form
	enc.Encode(len(t.CoinOutputs))
	for _, co := range t.CoinOutputs {
		enc.EncodeAll(
			co.Value,
			legacyUnlockHashCondition(co.Condition.Condition),
		)
	}
	for _, bsi := range t.BlockStakeInputs {
		enc.EncodeAll(bsi.ParentID, legacyUnlockHashFromFulfillment(bsi.Fulfillment.Fulfillment))
	}
	// legacy transactions encoded unlock hashes in pure form
	enc.Encode(len(t.BlockStakeOutputs))
	for _, bso := range t.BlockStakeOutputs {
		enc.EncodeAll(
			bso.Value,
			legacyUnlockHashCondition(bso.Condition.Condition),
		)
	}
	enc.EncodeAll(
		t.MinerFees,
		t.ArbitraryData,
	)

	var hash crypto.Hash
	h.Sum(hash[:0])
	return hash
}

func legacyUnlockHashCondition(uc UnlockCondition) UnlockHash {
	uhc, ok := uc.(*UnlockHashCondition)
	if !ok {
		if build.DEBUG {
			panic(fmt.Sprintf("unexpected condition %[1]v (%[1]T) encountered", uc))
		}
		return NilUnlockHash
	}
	return uhc.TargetUnlockHash
}

func legacyUnlockHashFromFulfillment(uf UnlockFulfillment) UnlockHash {
	switch tuf := uf.(type) {
	case *SingleSignatureFulfillment:
		return NewUnlockHash(UnlockTypePubKey,
			crypto.HashObject(siabin.Marshal(tuf.PublicKey)))
	case *LegacyAtomicSwapFulfillment:
		return NewUnlockHash(UnlockTypeAtomicSwap,
			crypto.HashObject(siabin.MarshalAll(
				tuf.Sender, tuf.Receiver, tuf.HashedSecret, tuf.TimeLock)))
	default:
		if build.DEBUG {
			panic(fmt.Sprintf("unexpected fulfillment %[1]v (%[1]T) encountered", uf))
		}
		return NilUnlockHash
	}
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
