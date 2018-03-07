// Copyright GreenITGlobe/ThreeFold 2018, Licensed under the MIT Licence

package bip39

import (
<<<<<<< HEAD
	"crypto/sha256"
=======
>>>>>>> 53156a661ae8bc5176a7770d6991a8578d6794bb
	"errors"
	"strings"
)

// Package bip39 converts human readable phrase to slice of bytes and backwards
// Human readable string can be worded in English, German, Japanese, Chinesse and Spanish
//
// Any 128 to 256 bit hash, can be converted to a sequence of words that are easy to
// read, and to transmit.
//
// Words a chosen from a 2048 dicctionary of carefully picked words to be identifies by the
// first 4 letter and to avoid confusion with other similar ones.

// Missing examples

// ToDo: adding the CRC check

type (
	// Phrase is the human readable version of a random []byte. Most typically,
	// a phrase is displayed to the user using the String method.
	Phrase []string
)

var (
	errEmptyInput        = errors.New("input is empty not valid for conversion")
	errUnknownDictionary = errors.New("language not recognized")
	errUnknownWord       = errors.New("word not found in dictionary for given language")
	errBadCrc            = errors.New("Bad CRC, contents does not validate")
)

// FromPhrase converts a []phrase to the embedded []byte codified in the phrase.
func FromPhrase(p Phrase, did DictionaryID) ([]byte, error) {
	if len(p) == 0 {
		return nil, errEmptyInput
	}

	//  check the CRC and remove it from answer
	data, err := phraseToBytes(p, did)
	if err != nil {
		return nil, err
	}

	var entropy = data[:len(data)-4]
	var crc = data[len(data)-4:]
	if crcToInt32(crc) != crcToInt32(calcCrc(entropy)) {
		return nil, errBadCrc
	}

	return entropy, nil
}

// calcCrc return the first 32 bits of SHA256(entropy)
func calcCrc(entropy []byte) []byte {
	sha := sha256.Sum256(entropy)
	return sha[:4]
}

func crcToInt32(crc []byte) uint32 {
	var ret uint32
	var exp uint32 = 1

	for i := uint32(0); i < 4; i++ {
		ret += exp * uint32(crc[i])
		exp *= 256
	}
	return ret
}

// phraseToBytes returns the 11bit []int codified in the phrase
func phraseToBytes(p Phrase, did DictionaryID) ([]byte, error) {
	var src = make([]int, 0, len(p))
	var err error
	var value int

	for _, word := range p {
		value, err = searchDic(word, did)
		if err != nil {
			return nil, err
		}
		src = append(src, value)
	}

	return decode11(src)
}

// searchDic returns the index of a given word
func searchDic(word string, did DictionaryID) (int, error) {
	var i, j, mid int
	var dictionary = bibliotheque[did]

	// Dichotomic search
	for i, j = 0, DictionarySize; i != j && j-i > 1; mid = (i + j) / 2 {
		switch {
		case dictionary[mid] > word:
			j = mid
		case dictionary[mid] < word:
			i = mid
		default:
			return mid, nil
		}
	}
	return 0, errUnknownWord
}

// FromString converts an input string into the []byte codified in the string
func FromString(str string, did DictionaryID) ([]byte, error) {
	phrase := Phrase(strings.Split(str, " "))
	return FromPhrase(phrase, did)
}

// ToPhrase converts a []byte to a human-friendly words
func ToPhrase(src []byte, did DictionaryID) (Phrase, error) {
	if len(src) == 0 {
		return nil, errEmptyInput
	}

	// Add the CRC
	crc := calcCrc(src)
	src = append(src, crc...)
	enc11, err := encode11(src)
	if err != nil {
		return nil, err
	}

	return encodePhrase(enc11, did), nil
}

// encodePhrase converts 11bit ints into words in a phrase
func encodePhrase(enc11 []int, did DictionaryID) Phrase {
	var p = make(Phrase, 0, len(enc11))
	var dictionary = bibliotheque[did]

	for _, v := range enc11 {
		p = append(p, dictionary[v])
	}
	return p
}

// String concatenates a phrase words into a single string.
func (p Phrase) String() string {
	return strings.Join(p, " ")
}
