package modules

import (
	"bytes"
	"testing"
)

// TestSeedMnemonicFunctions tests that
// seed = InitialSeedFromMnemonic(NewMnemonic(seed))
func TestSeedMnemonicFunctions(t *testing.T) {
	testCases := []Seed{
		{},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
		{
			255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
			255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
		},
		{
			1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23,
			24, 25, 26, 27, 28, 29, 30, 31,
		},
		{
			1, 4, 2, 3, 6, 4, 5, 8, 6, 7, 10, 8, 9, 12, 10, 11, 14, 12, 13, 16, 14, 15,
			18, 16, 17, 20, 18, 19, 22, 20, 21, 24,
		},
	}
	for _, initialSeed := range testCases {
		mnemonic, err := NewMnemonic(initialSeed)
		if err != nil {
			t.Errorf("failed to create mnemonic: %v", err)
			continue
		}
		if mnemonic == "" {
			t.Errorf("mnemonic created using i.seed %v is empty", initialSeed)
			continue
		}

		seed, err := InitialSeedFromMnemonic(mnemonic)
		if err != nil {
			t.Errorf("failed to recover seed from mnemonic %q: %v", mnemonic, err)
			continue
		}
		if bytes.Compare(initialSeed[:], seed[:]) != 0 {
			t.Errorf("out %v != %v in", seed, initialSeed)
		}
	}
}
