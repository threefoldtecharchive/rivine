package main

import (
	"bytes"
	"crypto/rand"
	"fmt"

	"github.com/stellar/go/keypair"
	"github.com/stellar/go/strkey"

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
	fmt.Printf("Private key (1st one generated from the seed): %x\n", h[:])
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
	fmt.Println("Generated address:", rivineUnlockHash.String())

	// Stellar
	stellarKeypair, err := keypair.FromRawSeed(h)
	if err != nil {
		panic("ERROR generating the stellar keypair ")
	}
	stellarSeed := stellarKeypair.Seed()
	fmt.Println("Stellar Seed:", stellarSeed)
	stellarAddress := stellarKeypair.Address()
	fmt.Println("Stellar address:", stellarAddress)

	//Back to Rivine
	decodedStellarPK := strkey.MustDecode(strkey.VersionByteAccountID, stellarAddress)
	copy(rivinePublicKey[:], decodedStellarPK)

	rivineFromstellarUnlockHash, err := types.NewEd25519PubKeyUnlockHash(rivinePublicKey)
	if err != nil {
		panic("ERROR generating the rivine unlockhash from the decoded stellar address")
	}
	fmt.Println("Rivine address from Stellar address:", rivineFromstellarUnlockHash.String())
}
