package main

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	mnemonics "gitlab.com/NebulousLabs/entropy-mnemonics"

	"github.com/threefoldtech/rivine/crypto"
	"github.com/threefoldtech/rivine/modules"
)

func main() {
	printUsage := func() {
		cmd := "seedmig"
		if len(os.Args) > 0 {
			cmd = os.Args[0]
		}
		fmt.Fprintf(os.Stderr, "USAGE: %s <old_29_words_seed>\n", cmd)
		os.Exit(1)
	}

	if len(os.Args) != 2 {
		printUsage()
	}

	inputPhrase := os.Args[1]
	phrase := mnemonics.Phrase(strings.Split(inputPhrase, " "))
	inputEntropy, err := mnemonics.FromPhrase(phrase, mnemonics.English)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error while interpreting given phrase: %v\n", err)
		printUsage()
	}

	fmt.Printf("Input phrase (= seed): %s\n", inputPhrase)
	fmt.Printf("Input entropy: %x\n", inputEntropy)
	fmt.Println()

	const (
		SeedSize         = crypto.EntropySize
		SeedChecksumSize = 6
		TotalSeedSize    = SeedSize + SeedChecksumSize
	)
	if TotalSeedSize != len(inputEntropy) {
		fmt.Fprintf(os.Stderr, "Incorrect input entropy length: expected length %d, but input entropy has length %d\n", TotalSeedSize, len(inputEntropy))
		os.Exit(1)
	}

	var seed modules.Seed
	copy(seed[:], inputEntropy[:])

	fullChecksum, err := crypto.HashObject(seed)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error while hashing seed: %v\n", err)
		os.Exit(1)
	}
	if len(inputEntropy) != TotalSeedSize || !bytes.Equal(fullChecksum[:SeedChecksumSize], inputEntropy[SeedSize:]) {
		fmt.Fprintln(os.Stderr, "Error while verifying seed: seed failed checksum verification")
		os.Exit(1)
	}

	outputPhrase, err := modules.NewMnemonic(seed)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error while converting input entropy to rivine-standard (BIP-39) mnemonic (= seed): %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✓ Input phrase and entropy checksum-verified")
	fmt.Println()

	fmt.Printf("Output phrase (= seed): %s\n", outputPhrase)
}
