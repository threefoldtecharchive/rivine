package bip39

const (
	decBits  = 8  // Number of bits in the origin
	encBits  = 11 // Number of bits in the destination
	set11bit = 0x400
	set8bit  = 0x80
)

func encode11(src []byte) []int {
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

	return ret
}

func decode11(src []int) []byte {
	var ret = make([]byte, 0, len(src))
	var bits int
	var b8 byte

	for _, v := range src {
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
	return ret
}
