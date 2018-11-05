package main

import (
	"bytes"
	"fmt"
	"log"
	"os"

	"github.com/threefoldtech/rivine/crypto"
	"github.com/threefoldtech/rivine/types"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatal("not enough arguments given, expected one: hex(unlock_hash)")
	}

	uhStr := os.Args[1]
	var uh types.UnlockHash
	err := uh.LoadString(uhStr)
	if err == nil {
		log.Fatal("Given unlock hash is already in the new correct format")
	}

	uh, err = LoadOldUnlockHashString(uhStr)
	if err != nil {
		log.Fatal("failed to load given unlock hash using the old checksum way:", err)
	}

	// print the new unlock hash
	fmt.Println(uh.String())
}

// LoadOldUnlockHashString tries to load the unlock hash
// using the old validation way
func LoadOldUnlockHashString(strUH string) (types.UnlockHash, error) {
	// Check the length of strUH.
	// total length is 39, 1 byte for the (unlock) type,
	// 32 for the hash itself and 6 for the (partial) checksum of the hash.
	// This amount gets multiplied by 2, as the unlock hash is hex encoded.
	if len(strUH) != (1+crypto.HashSize+types.UnlockHashChecksumSize)*2 {
		return types.UnlockHash{}, types.ErrUnlockHashWrongLen
	}

	// decode the unlock type
	var ut types.UnlockType
	_, err := fmt.Sscanf(strUH[:2], "%02x", &ut)
	if err != nil {
		return types.UnlockHash{}, err
	}

	// Decode the unlock hash.
	var unlockHashBytes []byte
	_, err = fmt.Sscanf(strUH[2:2+crypto.HashSize*2], "%x", &unlockHashBytes)
	if err != nil {
		return types.UnlockHash{}, err
	}

	// Decode and verify the checksum.
	var checksum []byte
	_, err = fmt.Sscanf(strUH[2+crypto.HashSize*2:], "%x", &checksum)
	if err != nil {
		return types.UnlockHash{}, err
	}
	expectedChecksum := crypto.HashBytes(unlockHashBytes)
	if !bytes.Equal(expectedChecksum[:types.UnlockHashChecksumSize], checksum) {
		return types.UnlockHash{}, types.ErrInvalidUnlockHashChecksum
	}

	// return unlock hash
	uh := types.UnlockHash{Type: ut}
	copy(uh.Hash[:], unlockHashBytes[:])
	return uh, nil
}
