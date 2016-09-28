// Copyright 2013 The Go Authors. All rights reserved.  Use of this source code
// is governed by a BSD-style license that can be found in the LICENSE file.

// Package ed25519 implements the Ed25519 signature algorithm. See
// http://ed25519.cr.yp.to/.
package ed25519

// This code is a port of the public domain, "ref10" implementation of ed25519
// from SUPERCOP.

import (
	"bytes"
	"crypto/sha512"
)

const (
	// EntropySize is the number of bytes used as input to GenerateKey.
	EntropySize = 32

	// PublicKeySize is the size of the public key in bytes.
	PublicKeySize = 32

	// SecretKeySize is the size of the secret key in bytes.
	SecretKeySize = 64

	// SignatureSize is size of the signature in bytes.
	SignatureSize = 64
)

type (
	// PublicKey is used to verify signatures.
	PublicKey *[PublicKeySize]byte

	// SecretKey is used to sign messages.
	SecretKey *[SecretKeySize]byte

	// Signature is used to authenticate a message.
	Signature *[SignatureSize]byte
)

// GenerateKey generates a public/secret key pair using randomness from rand.
func GenerateKey(entropy [EntropySize]byte) (sk SecretKey, pk PublicKey) {
	sk = new([SecretKeySize]byte)
	pk = new([PublicKeySize]byte)
	copy(sk[:32], entropy[:])

	h := sha512.New()
	h.Write(sk[:32])
	digest := h.Sum(nil)
	digest[0] &= 248
	digest[31] &= 127
	digest[31] |= 64

	A := geScalarMultBase(digest[:32])
	A.ToBytes(sk[32:])

	copy(pk[:], sk[32:])
	return sk, pk
}

// Sign signs the message with secretKey and returns a signature.
func Sign(sk SecretKey, message []byte) (sig Signature) {
	sig = new([SignatureSize]byte)

	h := sha512.New()
	h.Write(sk[:32])

	digest := h.Sum(nil)
	digest[0] &= 248
	digest[31] &= 63
	digest[31] |= 64

	h.Reset()
	h.Write(digest[32:])
	h.Write(message)
	messageDigest := h.Sum(nil)

	messageDigestReduced := scReduce(messageDigest)
	R := geScalarMultBase(messageDigestReduced[:])
	R.ToBytes(sig[:32])

	h.Reset()
	h.Write(sig[:32])
	h.Write(sk[32:])
	h.Write(message)
	hramDigest := h.Sum(nil)
	hramDigestReduced := scReduce(hramDigest)
	scMulAdd(sig[32:], hramDigestReduced, digest[:32], messageDigestReduced)
	return sig
}

// Verify returns true iff sig is a valid signature of message by publicKey.
func Verify(pk PublicKey, message []byte, sig Signature) bool {
	if sig[63]&224 != 0 {
		return false
	}
	var A extendedGroupElement
	if !A.FromBytes(pk[:]) {
		return false
	}

	h := sha512.New()
	h.Write(sig[:32])
	h.Write(pk[:])
	h.Write(message)
	digest := h.Sum(nil)

	hReduced := scReduce(digest[:])
	R := geDoubleScalarMultVartime(hReduced, &A, sig[32:])

	var checkR [32]byte
	R.ToBytes(&checkR)
	return bytes.Equal(sig[:32], checkR[:])
}
