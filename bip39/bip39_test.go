// Copyright GreenITGlobe/ThreeFold 2018, Licensed under the MIT Licence
package bip39

import (
	"encoding/hex"
	"testing"
)

const (
// sampleWords = []string{"hola"}
)

var (
	sampleHex = []byte("fc1ad7569fc5d1abe9971690d9bf8b05")
)

func isError(t *testing.T, err error) {
	if err != nil {
		t.Error(err)
	}
}

func TestCoder11(t *testing.T) {
	// Check Encoding and Decoding
	var src = make([]byte, hex.DecodedLen(len(sampleHex)))
	hex.Decode(src, sampleHex)
	resp, err := encode11(src)
	isError(t, err)
	orig, err := decode11(resp)
	isError(t, err)
	if string(orig) != string(src) {
		t.Error("Coder not working properly")
	}

	// Check bad length
	src = src[1:]
	_, err = encode11(src)
	if err != errModulo {
		t.Error("Bad encoding Length accepted")
	}
}

func TestPhrases(t *testing.T) {
	var src = make([]byte, hex.DecodedLen(len(sampleHex)))
	hex.Decode(src, sampleHex)
	for did := range bibliotheque {
		t.Log("testing ", did)
		phrases, err := ToPhrase(src, did)
		isError(t, err)
		if len(phrases) != 15 {
			t.Error("ToPhrase don't work right")
		}

		orig, err := FromPhrase(phrases, did)
		isError(t, err)
		// hex.Encode(dst, orig)
		if string(orig) != string(src) {
			t.Error("Phraser not working properly")
		}
	}
}
