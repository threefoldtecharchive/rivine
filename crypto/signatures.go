package crypto

import (
	"bytes"
	"errors"
	"io"

	"github.com/NebulousLabs/fastrand"
	"github.com/threefoldtech/rivine/pkg/encoding/siabin"

	"golang.org/x/crypto/ed25519"
)

const (
	// EntropySize defines the amount of entropy necessary to do secure
	// cryptographic operations, in bytes.
	EntropySize = 32

	// PublicKeySize defines the size of public keys in bytes.
	PublicKeySize = ed25519.PublicKeySize

	// SecretKeySize defines the size of secret keys in bytes.
	SecretKeySize = ed25519.PrivateKeySize

	// SignatureSize defines the size of signatures in bytes.
	SignatureSize = ed25519.SignatureSize
)

var (
	// ErrInvalidSignature is returned if a signature is provided that does not
	// match the data and public key.
	ErrInvalidSignature = errors.New("invalid signature")

	// ErrSecretNilKey is returned if a secret key, which is used, is nil
	ErrSecretNilKey = errors.New("secret key is nil")

	// ErrPublicNilKey is returned if a public key, which is used, is nil
	ErrPublicNilKey = errors.New("public key is nil")
)

type (
	// PublicKey is an object that can be used to verify signatures.
	PublicKey [PublicKeySize]byte

	// SecretKey can be used to sign data for the corresponding public key.
	SecretKey [SecretKeySize]byte

	// Signature proves that data was signed by the owner of a particular
	// public key's corresponding secret key.
	Signature [SignatureSize]byte
)

var (
	// nilSecretKey defines a pure nil secret key
	nilSecretKey = SecretKey{}
	// nilPublicKey defines a pure nil public key
	nilPublicKey = PublicKey{}
)

// PublicKey returns the public key that corresponds to a secret key.
func (sk SecretKey) PublicKey() (pk PublicKey) {
	copy(pk[:], sk[SecretKeySize-PublicKeySize:])
	return
}

// IsNil returns true if this secret key equals to a pure zero array.
func (sk SecretKey) IsNil() bool {
	return sk == nilSecretKey
}

// IsNil returns true if this public key equals to a pure zero array.
func (pk PublicKey) IsNil() bool {
	return pk == nilPublicKey
}

// GenerateKeyPair creates a public-secret keypair that can be used to sign and verify
// messages.
func GenerateKeyPair() (sk SecretKey, pk PublicKey) {
	// no error possible when using fastrand.Reader
	epk, esk, _ := ed25519.GenerateKey(fastrand.Reader)
	copy(sk[:], esk)
	copy(pk[:], epk)
	return
}

// GenerateKeyPairDeterministic generates keys deterministically using the input
// entropy. The input entropy must be 32 bytes in length.
func GenerateKeyPairDeterministic(entropy [EntropySize]byte) (sk SecretKey, pk PublicKey) {
	// no error possible when using bytes.Reader
	epk, esk, _ := ed25519.GenerateKey(bytes.NewReader(entropy[:]))
	copy(sk[:], esk)
	copy(pk[:], epk)
	return
}

// ReadSignedObject reads a length-prefixed object prefixed by its signature,
// and verifies the signature.
func ReadSignedObject(r io.Reader, obj interface{}, maxLen uint64, pk PublicKey) error {
	// read the signature
	var sig Signature
	err := siabin.NewDecoder(r).Decode(&sig)
	if err != nil {
		return err
	}
	// read the encoded object
	encObj, err := siabin.ReadPrefix(r, maxLen)
	if err != nil {
		return err
	}
	// verify the signature
	if err := VerifyHash(HashBytes(encObj), pk, sig); err != nil {
		return err
	}
	// decode the object
	return siabin.Unmarshal(encObj, obj)
}

// SignHash signs a message using a secret key.
func SignHash(data Hash, sk SecretKey) (sig Signature) {
	copy(sig[:], ed25519.Sign(sk[:], data[:]))
	return
}

// VerifyHash uses a public key and input data to verify a signature.
func VerifyHash(data Hash, pk PublicKey, sig Signature) error {
	verifies := ed25519.Verify(pk[:], data[:], sig[:])
	if !verifies {
		return ErrInvalidSignature
	}
	return nil
}

// WriteSignedObject writes a length-prefixed object prefixed by its signature.
func WriteSignedObject(w io.Writer, obj interface{}, sk SecretKey) error {
	objBytes, err := siabin.Marshal(obj)
	if err != nil {
		return err
	}
	sig := SignHash(HashBytes(objBytes), sk)
	return siabin.NewEncoder(w).EncodeAll(sig, objBytes)
}
