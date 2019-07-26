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
	"io"
	"strings"

	"github.com/threefoldtech/rivine/build"
	"github.com/threefoldtech/rivine/crypto"
	"github.com/threefoldtech/rivine/pkg/encoding/rivbin"
	"github.com/threefoldtech/rivine/pkg/encoding/siabin"
)

// SignatureAlgoType identifies a signature algorithm as a single byte.
type SignatureAlgoType uint8

const (
	// SignatureAlgoNil identifies a nil SignatureAlgoType value.
	SignatureAlgoNil SignatureAlgoType = iota
	// SignatureAlgoEd25519 identifies the Ed25519 signature Algorithm,
	// the default (and only) algorithm supported by this chain.
	SignatureAlgoEd25519
)

// These Specifiers enumerate the string versions of the types of signatures that are recognized
// by this implementation. see Consensus.md for more details.
var (
	SignatureAlgoNilSpecifier     = Specifier{}
	SignatureAlgoEd25519Specifier = Specifier{'e', 'd', '2', '5', '5', '1', '9'}
)

func (sat SignatureAlgoType) String() string {
	return sat.Specifier().String()
}

// Specifier returns the specifier linked to this Signature Algorithm Type,
// returns the SignatureAlgoNilSpecifier if the algorithm type is unknown.
func (sat SignatureAlgoType) Specifier() Specifier {
	switch sat {
	case SignatureAlgoEd25519:
		return SignatureAlgoEd25519Specifier
	default:
		return SignatureAlgoNilSpecifier
	}
}

// LoadString loads the stringified algo type as its single byte representation.
func (sat *SignatureAlgoType) LoadString(str string) error {
	switch str {
	case SignatureAlgoEd25519Specifier.String():
		*sat = SignatureAlgoEd25519
	case SignatureAlgoNilSpecifier.String():
		*sat = SignatureAlgoNil
	default:
		return fmt.Errorf("unknown SignatureAlgoType string: %s", str)
	}
	return nil
}

// LoadSpecifier loads the algorithm type in specifier-format.
func (sat *SignatureAlgoType) LoadSpecifier(specifier Specifier) error {
	switch specifier {
	case SignatureAlgoEd25519Specifier:
		*sat = SignatureAlgoEd25519
	case SignatureAlgoNilSpecifier:
		*sat = SignatureAlgoNil
	default:
		return fmt.Errorf("unknown SignatureAlgoType specifier: %s", specifier.String())
	}
	return nil
}

// Signature-related errors
var (
	//ErrFrivolousSignature        = errors.New("transaction contains a frivolous signature")
	//ErrInvalidPubKeyIndex        = errors.New("transaction contains a signature that points to a nonexistent public key")
	ErrInvalidUnlockHashChecksum = errors.New("provided unlock hash has an invalid checksum")
	//ErrMissingSignatures         = errors.New("transaction has inputs with missing signatures")
	//ErrPrematureSignature        = errors.New("timelock on signature has not expired")
	//ErrPublicKeyOveruse          = errors.New("public key was used multiple times while signing transaction")
	//ErrSortedUniqueViolation     = errors.New("sorted unique violation")
	ErrUnlockHashWrongLen = errors.New("marshalled unlock hash is the wrong length")
)

type (
	// A PublicKey is a public key prefixed by a Specifier. The Specifier
	// indicates the algorithm used for signing and verification. Unrecognized
	// algorithms will always verify, which allows new algorithms to be added to
	// the protocol via a soft-fork.
	PublicKey struct {
		Algorithm SignatureAlgoType
		Key       ByteSlice
	}

	// ByteSlice defines any kind of raw binary value,
	// in-memory defined as a byte slice,
	// and JSON-encoded in hexadecimal form.
	ByteSlice []byte
)

// Ed25519PublicKey returns pk as a PublicKey, denoting its algorithm as
// Ed25519.
func Ed25519PublicKey(pk crypto.PublicKey) PublicKey {
	return PublicKey{
		Algorithm: SignatureAlgoEd25519,
		Key:       pk[:],
	}
}

// SignatureHash returns the hash of all fields in a transaction,
// relevant to a Tx sig.
func (t Transaction) SignatureHash(extraObjects ...interface{}) (crypto.Hash, error) {
	controller, exists := _RegisteredTransactionVersions[t.Version]
	if !exists {
		return crypto.Hash{}, ErrUnknownTransactionType
	}
	if hasher, ok := controller.(TransactionSignatureHasher); ok {
		// if extension implements TransactionSignatureHasher,
		// use it here to sign the input with it
		return hasher.SignatureHash(t, extraObjects...)
	}

	h := crypto.NewHash()
	enc := siabin.NewEncoder(h)

	enc.Encode(t.Version)
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
		build.Severe(fmt.Sprintf("unexpected condition %[1]v (%[1]T) encountered", uc))
		return NilUnlockHash
	}
	return uhc.TargetUnlockHash
}

func legacyUnlockHashFromFulfillment(uf UnlockFulfillment) UnlockHash {
	switch tuf := uf.(type) {
	case *SingleSignatureFulfillment:
		pkb, _ := siabin.Marshal(tuf.PublicKey)
		h, _ := crypto.HashObject(pkb)
		return NewUnlockHash(UnlockTypePubKey, h)
	case *LegacyAtomicSwapFulfillment:
		b, _ := siabin.MarshalAll(tuf.Sender, tuf.Receiver, tuf.HashedSecret, tuf.TimeLock)
		h, _ := crypto.HashObject(b)
		return NewUnlockHash(UnlockTypeAtomicSwap, h)
	default:
		build.Severe(fmt.Sprintf("unexpected fulfillment %[1]v (%[1]T) encountered", uf))
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

// MarshalSia implements SiaMarshaler.MarshalSia
func (pk PublicKey) MarshalSia(w io.Writer) error {
	return siabin.NewEncoder(w).EncodeAll(
		pk.Algorithm.Specifier(),
		pk.Key,
	)
}

// UnmarshalSia implements SiaUnmarshaler.UnmarshalSia
func (pk *PublicKey) UnmarshalSia(r io.Reader) error {
	// decode the algorithm type, required to know
	// what length of byte slice to expect
	var algoSpecifier Specifier
	err := siabin.NewDecoder(r).DecodeAll(&algoSpecifier, &pk.Key)
	if err != nil {
		return err
	}
	return pk.Algorithm.LoadSpecifier(algoSpecifier)
}

// MarshalRivine implements RivineMarshaler.MarshalRivine
func (pk PublicKey) MarshalRivine(w io.Writer) error {
	err := rivbin.NewEncoder(w).Encode(pk.Algorithm)
	if err != nil || pk.Algorithm == SignatureAlgoNil {
		return err // nil if pk.Algorithm == SignatureAlgoNil
	}
	l, err := w.Write([]byte(pk.Key))
	if err != nil {
		return err
	}
	if l != len(pk.Key) {
		return io.ErrShortWrite
	}
	return nil
}

// UnmarshalRivine implements RivineUnmarshaler.UnmarshalRivine
func (pk *PublicKey) UnmarshalRivine(r io.Reader) error {
	// decode the algorithm type, required to know
	// what length of byte slice to expect
	err := rivbin.NewDecoder(r).Decode(&pk.Algorithm)
	if err != nil {
		return err
	}
	// create the expected sized byte slice, depending on the algorithm type
	switch pk.Algorithm {
	case SignatureAlgoEd25519:
		pk.Key = make(ByteSlice, crypto.PublicKeySize)
	case SignatureAlgoNil:
		pk.Key = nil
	default:
		return fmt.Errorf("unknown SignatureAlgoType %d", pk.Algorithm)
	}
	// read byte slice
	_, err = io.ReadFull(r, pk.Key[:])
	return err
}

// LoadString is the inverse of PublicKey.String().
func (pk *PublicKey) LoadString(s string) error {
	parts := strings.SplitN(s, ":", 2)
	if len(parts) != 2 {
		return errors.New("invalid public key string")
	}
	err := pk.Key.LoadString(parts[1])
	if err != nil {
		return err
	}
	return pk.Algorithm.LoadString(parts[0])
}

// String defines how to print a PublicKey - hex is used to keep things
// compact during logging. The key type prefix and lack of a checksum help to
// separate it from a sia address.
func (pk *PublicKey) String() string {
	return pk.Algorithm.String() + ":" + pk.Key.String()
}

// MarshalJSON marshals a byte slice as a hex string.
func (pk PublicKey) MarshalJSON() ([]byte, error) {
	return json.Marshal(pk.String())
}

// UnmarshalJSON decodes the json string of the byte slice.
func (pk *PublicKey) UnmarshalJSON(b []byte) error {
	var str string
	if err := json.Unmarshal(b, &str); err != nil {
		return err
	}
	return pk.LoadString(str)
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
