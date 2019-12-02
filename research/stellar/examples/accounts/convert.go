package main

import (
	"bytes"
	"crypto/rand"
	"fmt"

	"github.com/stellar/go/keypair"

	"github.com/threefoldtech/rivine/crypto"
	"github.com/threefoldtech/rivine/modules"
	"github.com/threefoldtech/rivine/types"
	bip39 "github.com/tyler-smith/go-bip39"
	"golang.org/x/crypto/ed25519"
)

func main() {

	//Generic (rivine code used but not needed)
	seed := modules.Seed{}
	if _, err := rand.Read(seed[:]); err != nil {
		panic("ERROR reading from rand")
	}
	seedphrase, err := bip39.NewMnemonic(seed[:])
	if err != nil {
		panic("ERROR generating mnemonic from seed")
	}
	fmt.Println("Generated seed:", seedphrase)
	h, err := crypto.HashAll(seed, 1)
	if err != nil {
		panic("ERROR hashing the seed and the index")
	}
	fmt.Printf("Rivine and stellar private key: %x\n", h)
	//Rivine
	rawrivinePublicKey, rawrivineSecretKey, err := ed25519.GenerateKey(bytes.NewReader(h[:]))
	if err != nil {
		panic("ERROR generating the ed25519 key")
	}
	fmt.Printf("Generated rivine ed25519 private key (len=%d): %x\n", len(rawrivineSecretKey), rawrivineSecretKey)
	var rivinePublicKey crypto.PublicKey
	copy(rivinePublicKey[:], rawrivinePublicKey)
	rivineUnlockHash, err := types.NewEd25519PubKeyUnlockHash(rivinePublicKey)
	if err != nil {
		panic("ERROR generating the rivine unlockhash")
	}
	fmt.Println("Generated address:", rivineUnlockHash.String())

	// Stellar
	stellarKeypair, err := keypair.FromRawSeed(h)
	if err != nil {
		panic("ERROR generating the stellar keypair ")
	}
	fmt.Println("Stellar Seed:", stellarKeypair.Seed())
	fmt.Println("Stellar address:", stellarKeypair.Address())

}
