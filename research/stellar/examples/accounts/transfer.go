package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/pelletier/go-toml"
	"github.com/stellar/go/build"
	"github.com/stellar/go/clients/horizon"
	"github.com/stellar/go/keypair"
)

func main() {
	var accountname string
	flag.StringVar(&accountname, "name", "default", "The name of the account to create")
	var destination string
	flag.StringVar(&destination, "destination", "", "Destination address")

	flag.Parse()
	config, err := toml.LoadFile("config.toml")

	if err != nil {
		log.Fatal(err)
	}
	seed := config.Get(accountname + ".seed")
	if seed == nil {
		log.Fatal("No such account")
	}
	if seed == nil {
		log.Fatal("No such account")
	}
	newPK, err := keypair.Parse(seed.(string))
	pair, _ := newPK.(*keypair.Full)

	log.Println("From Address", pair.Address())

	tx, err := build.Transaction(
		build.TestNetwork,
		build.SourceAccount{pair.Seed()},
		build.AutoSequence{horizon.DefaultTestNetClient},
		build.Payment(
			build.Destination{destination},
			build.NativeAmount{"10"},
		),
	)

	if err != nil {
		panic(err)
	}

	// Sign the transaction to prove you are actually the person sending it.
	txe, err := tx.Sign(pair.Seed())
	if err != nil {
		panic(err)
	}

	txeB64, err := txe.Base64()
	if err != nil {
		panic(err)
	}

	// And finally, send it off to Stellar!
	resp, err := horizon.DefaultTestNetClient.SubmitTransaction(txeB64)
	if err != nil {
		he := err.(*horizon.Error)
		log.Println(he.Problem.Detail)
		resultcodes := he.ResultCodes
		log.Println(resultcodes)
		for _, ex := range he.Problem.Extras {
			log.Printf("%s\n", ex)
		}
		panic(err)

	}

	fmt.Println("Successful Transaction:")
	fmt.Println("Ledger:", resp.Ledger)
	fmt.Println("Hash:", resp.Hash)
}
