package types

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/threefoldtech/rivine/build"
	"github.com/threefoldtech/rivine/crypto"
	"github.com/threefoldtech/rivine/pkg/encoding/rivbin"
	"github.com/threefoldtech/rivine/pkg/encoding/siabin"
)

// unlockhash.go contains the unlockhash alias along with usability methods
// such as String and an implementation of sort.Interface.

const (
	// UnlockHashChecksumSize is the size of the checksum used to verify
	// human-readable addresses. It is not a crypytographically secure
	// checksum, it's merely intended to prevent typos. 6 is chosen because it
	// brings the total size of the address to 38 bytes, leaving 2 bytes for
	// potential version additions in the future.
	UnlockHashChecksumSize = 6
)

type (
	// UnlockType defines the type of
	// an unlock condition-fulfillment pair.
	UnlockType byte

	// An UnlockHash is a specially constructed hash of the UnlockConditions type.
	// "Locked" values can be unlocked by providing the UnlockConditions that hash
	// to a given UnlockHash. See SpendConditions.UnlockHash for details on how the
	// UnlockHash is constructed.
	UnlockHash struct {
		Type UnlockType
		Hash crypto.Hash
	}

	// UnlockHashSlice defines an optionally sorted
	// slice of unlock hashes.
	UnlockHashSlice []UnlockHash
)

const (
	// UnlockTypeNil defines a nil (empty) Input Lock and is the default.
	UnlockTypeNil UnlockType = iota

	// UnlockTypePubKey provides the standard and most simple unlock type.
	// In it the sender gives the public key of the intendend receiver.
	// The receiver can redeem the relevant locked input by providing a signature
	// which proofs the ownership of the private key linked to the known public key.
	UnlockTypePubKey

	// UnlockTypeAtomicSwap provides the unlocking of a more advanced condition,
	// where before the TimeLock expired, the output can only go to the receiver,
	// who has to give the secret in order to do so. After the TimeLock,
	// the output can only be claimed by the sender, with no deadline in this phase.
	UnlockTypeAtomicSwap

	// UnlockTypeMultiSig provides a condition in which the receiving party
	// consists of multiple (possibly separate) identities. The output can only
	// be spent after at least the specified amount of identities have agreed,
	// by means of providing their signature.
	UnlockTypeMultiSig
)

var (
	NilUnlockHash     UnlockHash
	UnknownUnlockHash = UnlockHash{
		Type: UnlockTypeNil,
		Hash: crypto.Hash{
			255, 255, 255, 255, 255, 255, 255, 255,
			255, 255, 255, 255, 255, 255, 255, 255,
			255, 255, 255, 255, 255, 255, 255, 255,
			255, 255, 255, 255, 255, 255, 255, 255,
		},
	}
)

// NewEd25519PubKeyUnlockHash creates a new unlock hash of type UnlockTypePubKey,
// using a given public key which is assumed to be used in combination with the Ed25519 algorithm.
func NewEd25519PubKeyUnlockHash(pk crypto.PublicKey) (UnlockHash, error) {
	return NewPubKeyUnlockHash(Ed25519PublicKey(pk))
}

// NewPubKeyUnlockHash creates a new unlock hash of type UnlockTypePubKey,
// using a given Sia-standard Public key.
func NewPubKeyUnlockHash(pk PublicKey) (UnlockHash, error) {
	pkb, err := siabin.Marshal(pk)
	if err != nil {
		return UnlockHash{}, err
	}
	h, err := crypto.HashObject(pkb)
	if err != nil {
		return UnlockHash{}, err
	}
	return UnlockHash{
		Type: UnlockTypePubKey,
		Hash: h,
	}, nil
}

// NewUnlockHash creates a new unlock hash
func NewUnlockHash(t UnlockType, h crypto.Hash) UnlockHash {
	return UnlockHash{
		Type: t,
		Hash: h,
	}
}

func unlockHashFromHex(hstr string) (uh UnlockHash) {
	err := uh.LoadString(hstr)
	if err != nil {
		build.Critical(fmt.Sprintf("func unlockHashFromHex(%s) failed: %v", hstr, err))
	}
	return
}

// MarshalSia implements SiaMarshaler.MarshalSia
func (t UnlockType) MarshalSia(w io.Writer) error {
	_, err := w.Write([]byte{byte(t)})
	return err
}

// UnmarshalSia implements SiaUnmarshaler.UnmarshalSia
func (t *UnlockType) UnmarshalSia(r io.Reader) error {
	var bt [1]byte
	_, err := io.ReadFull(r, bt[:])
	*t = UnlockType(bt[0])
	return err
}

// MarshalRivine implements RivineMarshaler.MarshalRivine
func (t UnlockType) MarshalRivine(w io.Writer) error {
	return rivbin.MarshalUint8(w, uint8(t))
}

// UnmarshalRivine implements RivineUnmarshaler.UnmarshalRivine
func (t *UnlockType) UnmarshalRivine(r io.Reader) error {
	x, err := rivbin.UnmarshalUint8(r)
	if err != nil {
		return err
	}
	*t = UnlockType(x)
	return nil
}

// MarshalSia implements SiaMarshaler.MarshalSia
func (uh UnlockHash) MarshalSia(w io.Writer) error {
	return siabin.NewEncoder(w).EncodeAll(uh.Type, uh.Hash)
}

// UnmarshalSia implements SiaUnmarshaler.UnmarshalSia
func (uh *UnlockHash) UnmarshalSia(r io.Reader) error {
	return siabin.NewDecoder(r).DecodeAll(&uh.Type, &uh.Hash)
}

// MarshalRivine implements RivineMarshaler.MarshalRivine
func (uh UnlockHash) MarshalRivine(w io.Writer) error {
	return rivbin.NewEncoder(w).EncodeAll(uh.Type, uh.Hash)
}

// UnmarshalRivine implements RivineUnmarshaler.UnmarshalRivine
func (uh *UnlockHash) UnmarshalRivine(r io.Reader) error {
	return rivbin.NewDecoder(r).DecodeAll(&uh.Type, &uh.Hash)
}

// Cmp compares returns an integer comparing two unlock hashes lexicographically.
// The result will be 0 if a==b, -1 if a < b, and +1 if a > b.
// A nil argument is equivalent to an empty slice.
func (uh UnlockHash) Cmp(other UnlockHash) int {
	if uh.Type < other.Type {
		return -1
	}
	if uh.Type > other.Type {
		return 1
	}
	return bytes.Compare(uh.Hash[:], other.Hash[:])
}

// TODO: unit test (UnlockHash).Cmp

var (
	_ siabin.SiaMarshaler   = UnlockType(0)
	_ siabin.SiaMarshaler   = UnlockHash{}
	_ siabin.SiaUnmarshaler = (*UnlockType)(nil)
	_ siabin.SiaUnmarshaler = (*UnlockHash)(nil)

	_ rivbin.RivineMarshaler   = UnlockType(0)
	_ rivbin.RivineMarshaler   = UnlockHash{}
	_ rivbin.RivineUnmarshaler = (*UnlockType)(nil)
	_ rivbin.RivineUnmarshaler = (*UnlockHash)(nil)
)

// MarshalJSON is implemented on the unlock hash to always produce a hex string
// upon marshalling.
func (uh UnlockHash) MarshalJSON() ([]byte, error) {
	return json.Marshal(uh.String())
}

// UnmarshalJSON is implemented on the unlock hash to recover an unlock hash
// that has been encoded to a hex string.
func (uh *UnlockHash) UnmarshalJSON(b []byte) error {
	// load the json bytes as a raw string first
	var str string
	err := json.Unmarshal(b, &str)
	if err != nil {
		return err
	}
	// piggy-back the actual unlockhash decoding onto the LoadString method
	return uh.LoadString(str)
}

// String returns the hex representation of the unlock hash as a string - this
// includes a checksum.
func (uh UnlockHash) String() string {
	if uh.Type == 0 {
		return "" // nil unlock hash
	}

	uhChecksum, _ := crypto.HashAll(uh.Type, uh.Hash)
	return fmt.Sprintf("%02x%x%x",
		uh.Type, uh.Hash[:], uhChecksum[:UnlockHashChecksumSize])
}

// LoadString loads a hex representation (including checksum)
// of an unlock hash into an unlock hash object.
// An error is returned if the string is invalid or
// fails the checksum.
func (uh *UnlockHash) LoadString(strUH string) error {
	if strUH == "" {
		// an empty string is considered to be a(n) empty/nil unlock hash
		*uh = NilUnlockHash
		return nil
	}

	// Check the length of strUH.
	// total length is 39, 1 byte for the (unlock) type,
	// 32 for the hash itself and 6 for the (partial) checksum of the hash.
	// This amount gets multiplied by 2, as the unlock hash is hex encoded.
	if len(strUH) != (1+crypto.HashSize+UnlockHashChecksumSize)*2 {
		return ErrUnlockHashWrongLen
	}

	// decode the unlock type
	var ut UnlockType
	_, err := fmt.Sscanf(strUH[:2], "%02x", &ut)
	if err != nil {
		return err
	}

	// Decode the unlock hash.
	var unlockHashBytes []byte
	_, err = fmt.Sscanf(strUH[2:2+crypto.HashSize*2], "%x", &unlockHashBytes)
	if err != nil {
		return err
	}
	var unlockHash crypto.Hash
	copy(unlockHash[:], unlockHashBytes[:])

	if ut != UnlockTypeNil {
		// Decode and verify the checksum.
		var checksum []byte
		_, err = fmt.Sscanf(strUH[2+crypto.HashSize*2:], "%x", &checksum)
		if err != nil {
			return err
		}
		expectedChecksum, err := crypto.HashAll(ut, unlockHash)
		if err != nil {
			return err
		}
		if !bytes.Equal(expectedChecksum[:UnlockHashChecksumSize], checksum) {
			return ErrInvalidUnlockHashChecksum
		}
	} else {
		if unlockHash != NilUnlockHash.Hash {
			return fmt.Errorf("unexpected crypto hash for UnlockTypeNil: " + NilUnlockHash.Hash.String())
		}
		// Decode and verify the checksum.
		var checksum []byte
		_, err = fmt.Sscanf(strUH[2+crypto.HashSize*2:], "%x", &checksum)
		if err != nil {
			return err
		}
		if !bytes.Equal(make([]byte, UnlockHashChecksumSize), checksum) {
			return ErrInvalidUnlockHashChecksum
		}
	}

	uh.Type = ut
	uh.Hash = unlockHash
	return nil
}

// Len implements the Len method of sort.Interface.
func (uhs UnlockHashSlice) Len() int {
	return len(uhs)
}

// Less implements the Less method of sort.Interface.
func (uhs UnlockHashSlice) Less(i, j int) bool {
	return uhs[i].Cmp(uhs[j]) < 0
}

// Swap implements the Swap method of sort.Interface.
func (uhs UnlockHashSlice) Swap(i, j int) {
	uhs[i], uhs[j] = uhs[j], uhs[i]
}
