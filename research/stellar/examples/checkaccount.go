package main

//go get github.com/segmentio/go-loggly
import (
	"fmt"
	"log"

	"github.com/stellar/go/clients/horizon"

	"github.com/pelletier/go-toml"
	"github.com/stellar/go/keypair"
)

func main() {
	config, err := toml.LoadFile("config.toml")
	if err != nil {
		log.Fatal(err)
	}
	seed := config.Get("default.seed").(string)
	newPK, err := keypair.Parse(seed)
	pair, _ := newPK.(*keypair.Full)
	log.Println("Seed:", pair.Seed())
	log.Println("Address:", pair.Address())
	account, err := horizon.DefaultTestNetClient.LoadAccount(pair.Address())
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Balances for account:", pair.Address())

	for _, balance := range account.Balances {
		log.Println(balance)
	}
}
