// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ed25519

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"crypto/rand"
	"encoding/hex"
	"io"
	"os"
	"strings"
	"testing"
)

func TestSignVerify(t *testing.T) {
	secretKey, publicKey := GenerateKey([32]byte{})
	message := []byte("test message")

	sig := Sign(secretKey, message)
	if !Verify(publicKey, message, sig) {
		t.Errorf("valid signature rejected")
	}

	wrongMessage := []byte("wrong message")
	if Verify(publicKey, wrongMessage, sig) {
		t.Errorf("signature of different message accepted")
	}
}

func TestGolden(t *testing.T) {
	// sign.input.gz is a selection of test cases from
	// http://ed25519.cr.yp.to/python/sign.input
	testDataZ, err := os.Open("testdata.gz")
	if err != nil {
		t.Fatal(err)
	}
	defer testDataZ.Close()
	testData, err := gzip.NewReader(testDataZ)
	if err != nil {
		t.Fatal(err)
	}
	defer testData.Close()

	in := bufio.NewReaderSize(testData, 1<<12)
	lineNo := 0
	for {
		lineNo++
		lineBytes, isPrefix, err := in.ReadLine()
		if isPrefix {
			t.Fatal("bufio buffer too small")
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			t.Fatalf("error reading test data: %s", err)
		}

		line := string(lineBytes)
		parts := strings.Split(line, ":")
		if len(parts) != 5 {
			t.Fatalf("bad number of parts on line %d", lineNo)
		}

		privBytes, _ := hex.DecodeString(parts[0])
		pubKeyBytes, _ := hex.DecodeString(parts[1])
		msg, _ := hex.DecodeString(parts[2])
		sig, _ := hex.DecodeString(parts[3])
		// The signatures in the test vectors also include the message
		// at the end, but we just want R and S.
		sig = sig[:SignatureSize]

		if l := len(pubKeyBytes); l != PublicKeySize {
			t.Fatalf("bad public key length on line %d: got %d bytes", lineNo, l)
		}

		priv := new([SecretKeySize]byte)
		pub := new([PublicKeySize]byte)
		copy(priv[:], privBytes)
		copy(priv[32:], pubKeyBytes)
		copy(pub[:], pubKeyBytes)

		sig2 := Sign(priv, msg)
		if !bytes.Equal(sig, sig2[:]) {
			t.Errorf("different signature result on line %d: %x vs %x", lineNo, sig, sig2)
		}
		if !Verify(pub, msg, sig2) {
			t.Errorf("signature failed to verify on line %d", lineNo)
		}
	}
}

func BenchmarkGenerateKeys(b *testing.B) {
	var entropy [32]byte
	for i := 0; i < b.N; i++ {
		_, err := rand.Read(entropy[:])
		if err != nil {
			panic(err)
		}
		_, _ = GenerateKey(entropy)
	}
}

func BenchmarkSign(b *testing.B) {
	var entropy [32]byte
	_, err := rand.Read(entropy[:])
	if err != nil {
		panic(err)
	}
	sk, _ := GenerateKey(entropy)
	var secret [64]byte
	copy(secret[:], sk[:])

	b.ResetTimer()
	message := make([]byte, 64)
	for i := 0; i < b.N; i++ {
		_, err := rand.Read(message)
		if err != nil {
			panic(err)
		}

		_ = Sign(sk, message)
	}
}

func BenchmarkVerify(b *testing.B) {
	var entropy [32]byte
	_, err := rand.Read(entropy[:])
	if err != nil {
		panic(err)
	}
	sk, pk := GenerateKey(entropy)

	b.ResetTimer()
	message := make([]byte, 64)
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		_, err := rand.Read(message)
		if err != nil {
			panic(err)
		}

		sig := Sign(sk, message)
		b.StartTimer()
		ver := Verify(pk, message, sig)
		if !ver {
			panic(err)
		}
	}
}
