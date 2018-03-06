package bip39

import (
	"bytes"
	"unsafe"
)

// Todo:
//       Independence of constants
// Check size of source
// LittleEndian ? Not a problem as we only process bytes not ints

const (
	srcBits    = 8     // Number of bits in the origin
	dstBits    = 11    // Number of bits in the destination
	dstBitMask = 0x7FF // dstBits all Up to be used as AND mask
)

type base struct {
	// bits :   number of bits usables on bufptr
	// missing: number of bits that need to be replaces
	// read : number of bits readed that are disposable
	// bits and missing are redundant, but makes clear what's going on
	bits, missing, read uint
	// slice with the values converted
	ret []int
	// buffer : a uint32 addressed as a slice
	// bufptr : pointer to a uint32 stored on a slice
	buffer []byte
	bufptr *uint32
	// Reader with the data to be converted
	rd *bytes.Reader
	// Flag to show no more data available to convert
	exit bool
}

// encode11 given a slice of bytes, returns a slice of 11 bit ints
// the bits on both slices are the same, they difer only on the size
func encode11(src []byte) ([]int, error) {
	var ret = make([]int, 0, 1+len(src)*srcBits/dstBits)
	var sz uint32          // Reference to retrieve the number of bytes in a uint32
	sz = unsafe.Sizeof(sz) // Need to know the size of uint32
	var buffer = make([]byte, sz)
	var bufptr = (*uint32)(unsafe.Pointer(&buffer[0]))
	var rd = bytes.NewReader(srt)

	b11 = base{
		bits:   uint(sz),
		ret:    ret,
		buffer: buffer,
		bufptr: bufptr,
		rd:     rd}

	for i := range buffer {
		buffer[1], _ = rd.ReadByte()
	}

	for {
		b11.read11()
		b11.remove()
		if b11.add8() {
			break
		}
	}

	// there are no more bits to extract but the ones in buffer (24 bits)
	b11.read11() // read 11 bits
	b11.remove()
	b11.read11() // read 11 remaining bits, not all significants
	return b11.ret, nil
}

// add8 adds a byte from the reader to the buffer
// activate exit flag when there are no more data in the reader
func (b *base) add8() bool {
	if b.bits == 24 {
		c, err := b.rd.ReadByte()
		b.buffer[3] = c
		if err != nil {
			b.exit = true
			b.remove()
			return false
		}
		b.bits += srcBits
		missing -= srcBits
	}
	return true
}

// read11 reads 11 bits from the buffer and add it to the ret slice
func (b *base) read11() {
	if b.read == 0 {
		c := int(*b.bufptr) & dstBitMask
		b.ret = append(b.ret, c)
		b.read = 11
	}
	return
}

// remove, calcs the amount of bits to remove from the buffer
func (b *base) remove() {
	if b.read > 0 { // Do nothing if there are unread bits
		var minb = b.read                        // maximum number of bits to remove
		var maxb = (len(b.buffer) - 1) * srcBits // min num of bits to keep
		if !b.exit {                             // limit remove to 8 bits max
			if min > srcBits {
				minb = srcBits
			}

			if b.bits-min < maxb {
				minb = b.bits - maxb
			}
		}
		b.shiftr(minb) // remove the unwanted bits
	}
}

// shiftr, shifts the buffer to the right, removing effectively sr bits
func (b *base) shiftr(sr uint) {
	*b.bufptr >>= sr
	b.read -= sr // Update buffer control counters
	b.missing += sr
	b.bits -= sr
}
