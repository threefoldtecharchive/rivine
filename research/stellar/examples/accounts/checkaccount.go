package main

//go get github.com/segmentio/go-loggly
import (
	"flag"
	"log"

	"github.com/stellar/go/clients/horizon"

	"github.com/pelletier/go-toml"
	"github.com/stellar/go/keypair"
)

func main() {
	var accountname string
	flag.StringVar(&accountname, "name", "", " The name of the account to check, if not given, checks the blances of all accounts")
	flag.Parse()
	config, err := toml.LoadFile("config.toml")

	if err != nil {
		log.Fatal(err)
	}
	accountnames := []string{accountname}
	if accountname == "" {
		accountnames = config.Keys()
	}
	for _, accountname = range accountnames {
		log.Println("Accountname:", accountname)
		seed := config.Get(accountname + ".seed")
		if seed == nil {
			log.Println("No such account")
			continue
		}
		newPK, err := keypair.Parse(seed.(string))
		pair, _ := newPK.(*keypair.Full)
		log.Println("Seed:", pair.Seed())
		log.Println("Address:", pair.Address())
		account, err := horizon.DefaultTestNetClient.LoadAccount(pair.Address())
		if err != nil {
			log.Println(err)
			continue
		}

		log.Println("Balances for account:", pair.Address())

		for _, balance := range account.Balances {
			log.Println(balance)
		}
	}
}
