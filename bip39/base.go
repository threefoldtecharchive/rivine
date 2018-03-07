// Copyright GreenITGlobe/ThreeFold 2018, Licensed under the MIT Licence

package bip39

import "errors"

const (
	decBits  = 8  // Number of bits in the origin
	encBits  = 11 // Number of bits in the destination
	set11bit = 0x400
	set8bit  = 0x80
)

var (
	errModulo       = errors.New("The number of bits must be divisible by 32")
	errInvalidValue = errors.New("Value has more than 11 significant bits")
)

// encode11 takes a slice of bytes (8bits) and returns a slice of ints (11bits)
func encode11(src []byte) ([]int, error) {
	if (len(src) % 4) != 0 {
		return nil, errModulo
	}

	var ret = make([]int, 0, len(src))

	var bits int
	var b11 uint16

	for _, v := range src {
		for i := 0; i < decBits; i++ {
			bits++
			b11 >>= 1
			if byte(v&0x1) == 1 {
				b11 |= set11bit
			}
			v >>= 1

			if bits == encBits {
				bits = 0
				ret = append(ret, int(b11))
				b11 = 0
			}
		}
	}
	b11 >>= uint(encBits - bits)
	ret = append(ret, int(b11))

	return ret, nil
}

// decode11 takes a slice of ints (11bits) and returns a slice of bytes (8bits)
func decode11(src []int) ([]byte, error) {
	var ret = make([]byte, 0, len(src))
	var bits int
	var b8 byte

	for _, v := range src {
		if v >= 2^11 {
			return nil, errInvalidValue
		}
		for i := 0; i < encBits; i++ {
			bits++
			b8 >>= 1
			if byte(v&0x1) == 1 {
				b8 |= set8bit
			}

			v >>= 1
			if bits == decBits {
				ret = append(ret, b8)
				b8 = 0
				bits = 0
			}
		}
	}
	return ret, nil
}
