package types

import (
	"testing"

	"github.com/threefoldtech/rivine/pkg/encoding/siabin"
)

// BenchmarkEncodeEmptyBlock benchmarks encoding an empty block.
//
// i5-4670K, 9a90f86: 48 MB/s
func BenchmarkEncodeBlock(b *testing.B) {
	var block Block
	b.SetBytes(int64(len(siabin.Marshal(block))))
	for i := 0; i < b.N; i++ {
		siabin.Marshal(block)
	}
}

// BenchmarkDecodeEmptyBlock benchmarks decoding an empty block.
//
// i7-4770,  b0b162d: 38 MB/s
// i5-4670K, 9a90f86: 55 MB/s
func BenchmarkDecodeEmptyBlock(b *testing.B) {
	var block Block
	encodedBlock := siabin.Marshal(block)
	b.SetBytes(int64(len(encodedBlock)))
	for i := 0; i < b.N; i++ {
		err := siabin.Unmarshal(encodedBlock, &block)
		if err != nil {
			b.Fatal(err)
		}
	}
}
