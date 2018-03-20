package types

// signatures.go contains all of the types and functions related to creating
// and verifying transaction signatures. There are a lot of rules surrounding
// the correct use of signatures. Signatures can cover part or all of a
// transaction, can be multiple different algorithms, and must satify a field
// called 'UnlockConditions'.

import (
	"encoding/hex"
	"errors"
	"fmt"
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
		Algorithm Specifier `json:"algorithm"`
		Key       []byte    `json:"key"`
	}

	// A TransactionSignature is a signature that is included in the transaction.
	// The signature should correspond to a public key in one of the
	// UnlockConditions of the transaction. This key is specified first by
	// 'ParentID', which specifies the UnlockConditions, and then
	// 'PublicKeyIndex', which indicates the key in the UnlockConditions. There
	// are three types that use UnlockConditions: SiacoinInputs, SiafundInputs,
	// and FileContractTerminations. Each of these types also references a
	// ParentID, and this is the hash that 'ParentID' must match. The 'Timelock'
	// prevents the signature from being used until a certain height.
	// 'CoveredFields' indicates which parts of the transaction are being signed;
	// see CoveredFields.
	TransactionSignature struct {
		ParentID       crypto.Hash `json:"parentid"`
		PublicKeyIndex uint64      `json:"publickeyindex"`
		Timelock       BlockHeight `json:"timelock"`
		Signature      []byte      `json:"signature"`
	}

	// UnlockConditions are a set of conditions which must be met to execute
	// certain actions, such as spending a SiacoinOutput or terminating a
	// FileContract.
	//
	// The simplest requirement is that the block containing the UnlockConditions
	// must have a height >= 'Timelock'.
	//
	// 'PublicKeys' specifies the set of keys that can be used to satisfy the
	// UnlockConditions; of these, at least 'SignaturesRequired' unique keys must sign
	// the transaction. The keys that do not need to use the same cryptographic
	// algorithm.
	//
	// If 'SignaturesRequired' == 0, the UnlockConditions are effectively "anyone can
	// unlock." If 'SignaturesRequired' > len('PublicKeys'), then the UnlockConditions
	// cannot be fulfilled under any circumstances.
	UnlockConditions struct {
		PublicKeys         []SiaPublicKey `json:"publickeys"`
		SignaturesRequired uint64         `json:"signaturesrequired"`
	}

	// Each input has a list of public keys and a required number of signatures.
	// inputSignatures keeps track of which public keys have been used and how many
	// more signatures are needed.
	inputSignatures struct {
		remainingSignatures uint64
		possibleKeys        []SiaPublicKey
		usedKeys            map[uint64]struct{}
		index               int
	}
)

// Ed25519PublicKey returns pk as a SiaPublicKey, denoting its algorithm as
// Ed25519.
func Ed25519PublicKey(pk crypto.PublicKey) SiaPublicKey {
	return SiaPublicKey{
		Algorithm: SignatureEd25519,
		Key:       pk[:],
	}
}

// UnlockHash calculates the root hash of a Merkle tree of the
// UnlockConditions object. The leaves of this tree are formed by taking the
// hash of the timelock, the hash of the public keys (one leaf each), and the
// hash of the number of signatures. The keys are put in the middle because
// Timelock and SignaturesRequired are both low entropy fields; they can bee
// protected by having random public keys next to them.
func (uc UnlockConditions) UnlockHash() UnlockHash {
	tree := crypto.NewTree()
	for i := range uc.PublicKeys {
		tree.PushObject(uc.PublicKeys[i])
	}
	tree.PushObject(uc.SignaturesRequired)
	return UnlockHash(tree.Root())
}

// SigHash returns the hash of all the fields in a transaction.
func (t Transaction) SigHash(i int) (hash crypto.Hash) {
	h := crypto.NewHash()
	enc := encoding.NewEncoder(h)
	enc.EncodeAll(
		t.CoinInputs,
		t.CoinOutputs,
		t.BlockStakeInputs,
		t.BlockStakeOutputs,
		t.MinerFees,
		t.ArbitraryData,
		t.TransactionSignatures[i].ParentID,
		t.TransactionSignatures[i].PublicKeyIndex,
		t.TransactionSignatures[i].Timelock,
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
func (t *Transaction) validSignatures(currentHeight BlockHeight) error {
	// Create the inputSignatures object for each input.
	sigMap := make(map[crypto.Hash]*inputSignatures)
	for i, input := range t.CoinInputs {
		id := crypto.Hash(input.ParentID)
		_, exists := sigMap[id]
		if exists {
			return ErrDoubleSpend
		}

		sigMap[id] = &inputSignatures{
			remainingSignatures: input.UnlockConditions.SignaturesRequired,
			possibleKeys:        input.UnlockConditions.PublicKeys,
			usedKeys:            make(map[uint64]struct{}),
			index:               i,
		}
	}
	for i, input := range t.BlockStakeInputs {
		id := crypto.Hash(input.ParentID)
		_, exists := sigMap[id]
		if exists {
			return ErrDoubleSpend
		}

		sigMap[id] = &inputSignatures{
			remainingSignatures: input.UnlockConditions.SignaturesRequired,
			possibleKeys:        input.UnlockConditions.PublicKeys,
			usedKeys:            make(map[uint64]struct{}),
			index:               i,
		}
	}

	// Check all of the signatures for validity.
	for i, sig := range t.TransactionSignatures {
		// Check that sig corresponds to an entry in sigMap.
		inSig, exists := sigMap[crypto.Hash(sig.ParentID)]
		if !exists || inSig.remainingSignatures == 0 {
			return ErrFrivolousSignature
		}
		// Check that sig's key hasn't already been used.
		_, exists = inSig.usedKeys[sig.PublicKeyIndex]
		if exists {
			return ErrPublicKeyOveruse
		}
		// Check that the public key index refers to an existing public key.
		if sig.PublicKeyIndex >= uint64(len(inSig.possibleKeys)) {
			return ErrInvalidPubKeyIndex
		}
		// Check that the timelock has expired.
		if sig.Timelock > currentHeight {
			return ErrPrematureSignature
		}

		// Check that the signature verifies. Multiple signature schemes are
		// supported.
		publicKey := inSig.possibleKeys[sig.PublicKeyIndex]
		switch publicKey.Algorithm {
		case SignatureEntropy:
			// Entropy cannot ever be used to sign a transaction.
			return ErrEntropyKey

		case SignatureEd25519:
			// Decode the public key and signature.
			var edPK crypto.PublicKey
			err := encoding.Unmarshal([]byte(publicKey.Key), &edPK)
			if err != nil {
				return err
			}
			var edSig [crypto.SignatureSize]byte
			err = encoding.Unmarshal([]byte(sig.Signature), &edSig)
			if err != nil {
				return err
			}
			cryptoSig := crypto.Signature(edSig)

			sigHash := t.SigHash(i)
			err = crypto.VerifyHash(sigHash, edPK, cryptoSig)
			if err != nil {
				return err
			}

		default:
			// If the identifier is not recognized, assume that the signature
			// is valid. This allows more signature types to be added via soft
			// forking.
		}

		inSig.usedKeys[sig.PublicKeyIndex] = struct{}{}
		inSig.remainingSignatures--
	}

	// Check that all inputs have been sufficiently signed.
	for _, reqSigs := range sigMap {
		if reqSigs.remainingSignatures != 0 {
			return ErrMissingSignatures
		}
	}

	return nil
}

// LoadString is the inverse of SiaPublicKey.String().
func (spk *SiaPublicKey) LoadString(s string) {
	parts := strings.Split(s, ":")
	if len(parts) != 2 {
		return
	}
	var err error
	spk.Key, err = hex.DecodeString(parts[1])
	if err != nil {
		spk.Key = nil
		return
	}
	copy(spk.Algorithm[:], []byte(parts[0]))
}

// String defines how to print a SiaPublicKey - hex is used to keep things
// compact during logging. The key type prefix and lack of a checksum help to
// separate it from a sia address.
func (spk *SiaPublicKey) String() string {
	return spk.Algorithm.String() + ":" + fmt.Sprintf("%x", spk.Key)
}
