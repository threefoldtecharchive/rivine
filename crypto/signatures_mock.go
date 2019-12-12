package crypto

import (
	"bytes"
	"crypto/rand"
	"io"

	"golang.org/x/crypto/ed25519"
)

type (
	// keyDeriver contains all of the dependencies for a signature key
	// generator. The dependencies are separated to enable mocking.
	keyDeriver interface {
		deriveKeyPair([EntropySize]byte) (SecretKey, PublicKey)
	}
)

var (
	// stdKeyGen is a signature generator that can be used to generate random
	// and deterministic keys for signing objects.
	stdKeyGen = sigKeyGen{entropySource: rand.Reader, keyDeriver: &stdKeyDeriver{}}
)

// sigKeyGen contains a set of dependencies that are used to build out the core
// logic for generating keys in Sia.
type sigKeyGen struct {
	entropySource io.Reader
	keyDeriver    keyDeriver
}

// stdKeyDeriver implements the keyDeriver dependency for the sigKeyGen.
type stdKeyDeriver struct{}

func (skd *stdKeyDeriver) deriveKeyPair(entropy [EntropySize]byte) (sk SecretKey, pk PublicKey) {
	epk, esk, _ := ed25519.GenerateKey(bytes.NewReader(entropy[:]))
	copy(pk[:], epk[:])
	copy(sk[:], esk[:])
	return
}
