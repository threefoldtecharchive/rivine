package types

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/rivine/rivine/crypto"
	"github.com/rivine/rivine/encoding"
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

	// UnlockTypeSingleSignature provides the standard and most simple unlock type.
	// In it the sender gives the public key of the intendend receiver.
	// The receiver can redeem the relevant locked input by providing a signature
	// which proofs the ownership of the private key linked to the known public key.
	UnlockTypeSingleSignature

	// UnlockTypeAtomicSwap provides a more advanced unlocker,
	// which allows for a more advanced InputLock,
	// where before the TimeLock expired, the output can only go to the receiver,
	// who has to give the secret in order to do so. After the InputLock,
	// the output can only be claimed by the sender, with no deadline in this phas
	UnlockTypeAtomicSwap
)

// NewUnlockHash creates a new unlock hash
func NewUnlockHash(t UnlockType, h crypto.Hash) UnlockHash {
	return UnlockHash{
		Type: t,
		Hash: h,
	}
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

// MarshalSia implements SiaMarshaler.MarshalSia
func (uh UnlockHash) MarshalSia(w io.Writer) error {
	return encoding.NewEncoder(w).EncodeAll(uh.Type, uh.Hash)
}

// UnmarshalSia implements SiaUnmarshaler.UnmarshalSia
func (uh *UnlockHash) UnmarshalSia(r io.Reader) error {
	return encoding.NewDecoder(r).DecodeAll(&uh.Type, &uh.Hash)
}

var (
	_ encoding.SiaMarshaler   = UnlockType(0)
	_ encoding.SiaMarshaler   = UnlockHash{}
	_ encoding.SiaUnmarshaler = (*UnlockType)(nil)
	_ encoding.SiaUnmarshaler = (*UnlockHash)(nil)
)

// MarshalJSON is implemented on the unlock hash to always produce a hex string
// upon marshalling.
func (uh UnlockHash) MarshalJSON() ([]byte, error) {
	return json.Marshal(uh.String())
}

// UnmarshalJSON is implemented on the unlock hash to recover an unlock hash
// that has been encoded to a hex string.
func (uh *UnlockHash) UnmarshalJSON(b []byte) error {
	// Check the length of b.
	// total length is 39, 1 byte for the (unlock) type,
	// 32 for the hash itself and 6 for the (partial) checksum of the hash.
	// This amount gets multiplied by 2, as the unlock hash is hex encoded,
	// and on top of that we require 2 extra bytes for the double quote characters.
	// wrapping the hex-encoded unlockhash string, as it is a JSON-string.
	if len(b) != (1+crypto.HashSize+UnlockHashChecksumSize)*2+2 {
		return ErrUnlockHashWrongLen
	}
	return uh.LoadString(string(b[1 : len(b)-1]))
}

// String returns the hex representation of the unlock hash as a string - this
// includes a checksum.
func (uh UnlockHash) String() string {
	uhChecksum := crypto.HashObject(uh.Hash)
	return fmt.Sprintf("%02x%x%x",
		uh.Type, uh.Hash[:], uhChecksum[:UnlockHashChecksumSize])
}

// LoadString loads a hex representation (including checksum)
// of an unlock hash into an unlock hash object.
// An error is returned if the string is invalid or
// fails the checksum.
func (uh *UnlockHash) LoadString(strUH string) error {
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
	var byteUnlockHash []byte
	var checksum []byte
	_, err = fmt.Sscanf(strUH[2:2+crypto.HashSize*2], "%x", &byteUnlockHash)
	if err != nil {
		return err
	}

	// Decode and verify the checksum.
	_, err = fmt.Sscanf(strUH[2+crypto.HashSize*2:], "%x", &checksum)
	if err != nil {
		return err
	}
	expectedChecksum := crypto.HashBytes(byteUnlockHash)
	if !bytes.Equal(expectedChecksum[:UnlockHashChecksumSize], checksum) {
		return ErrInvalidUnlockHashChecksum
	}

	uh.Type = ut
	copy(uh.Hash[:], byteUnlockHash[:])
	return nil
}

// Len implements the Len method of sort.Interface.
func (uhs UnlockHashSlice) Len() int {
	return len(uhs)
}

// Less implements the Less method of sort.Interface.
func (uhs UnlockHashSlice) Less(i, j int) bool {
	if uhs[i].Type == uhs[j].Type {
		return bytes.Compare(uhs[i].Hash[:], uhs[j].Hash[:]) < 0
	}
	return uhs[i].Type < uhs[j].Type
}

// Swap implements the Swap method of sort.Interface.
func (uhs UnlockHashSlice) Swap(i, j int) {
	uhs[i], uhs[j] = uhs[j], uhs[i]
}
