package main

import (
	"bytes"
	"flag"
	"fmt"
	"strings"

	"github.com/stellar/go/keypair"

	"github.com/threefoldtech/rivine/crypto"
	"github.com/threefoldtech/rivine/modules"
	"github.com/threefoldtech/rivine/types"
	bip39 "github.com/tyler-smith/go-bip39"
	"golang.org/x/crypto/ed25519"
)

func main() {
	var showStellarSecrets bool
	var amountOfAddresses int
	flag.BoolVar(&showStellarSecrets, "secrets", false, "Print  the Stellar secrets")
	flag.IntVar(&amountOfAddresses, "amount", 2525, "amount of addresses to generate from the seed")
	flag.Parse()
	seedmnemonic := strings.Join(flag.Args(), " ")

	seedFromMnemonic, err := bip39.EntropyFromMnemonic(seedmnemonic)
	if err != nil {
		panic("ERROR generating seed from mnemonic")
	}
	seed := modules.Seed{}
	_ = copy(seed[:], seedFromMnemonic)
	for i := 0; i < amountOfAddresses; i++ {
		h, err := crypto.HashAll(seed, i)
		if err != nil {
			panic("ERROR hashing the seed and the index")
		}

		//Rivine
		rawrivinePublicKey, _, err := ed25519.GenerateKey(bytes.NewReader(h[:]))
		if err != nil {
			panic("ERROR generating the ed25519 key")
		}
		var rivinePublicKey crypto.PublicKey
		copy(rivinePublicKey[:], rawrivinePublicKey)
		rivineUnlockHash, err := types.NewEd25519PubKeyUnlockHash(rivinePublicKey)
		if err != nil {
			panic("ERROR generating the rivine unlockhash")
		}

		// Stellar
		stellarKeypair, err := keypair.FromRawSeed(h)
		if err != nil {
			panic("ERROR generating the stellar keypair ")
		}
		stellarSecret := ""
		if showStellarSecrets {
			stellarSecret = stellarKeypair.Seed()
		}
		stellarAddress := stellarKeypair.Address()
		fmt.Println(rivineUnlockHash.String(), " ", stellarAddress, " ", stellarSecret)
	}
}
