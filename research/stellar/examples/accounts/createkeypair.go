package main

// In order to run this example install the dependencies:
// ```sh
// go get -d github.com/stellar/go
// go get -d github.com/stellar/go-xdr/xdr3
// go get github.com/lib/pq
// go get github.com/pelletier/go-toml
// ```

import (
	"errors"
	"flag"
	"log"
	"os"

	"github.com/pelletier/go-toml"
	"github.com/stellar/go/keypair"
)

func main() {
	var accountname string
	flag.StringVar(&accountname, "name", "default", " The name of the account to create")
	flag.Parse()

	pair, err := generateKeyPair(accountname)
	if err != nil {
		log.Fatal(err)
	}
	err = saveSeed(accountname, pair.Seed())
	if err != nil {
		log.Fatal(err)
	}

}
func generateKeyPair(accountname string) (pair *keypair.Full, err error) {

	pair, err = keypair.Random()
	if err != nil {
		return
	}

	log.Println("Seed:", pair.Seed())
	log.Println("Address:", pair.Address())
	return
}

func saveSeed(accountname string, seed string) (err error) {

	var config *toml.Tree

	config, err = toml.LoadFile("config.toml")
	if err != nil {
		if os.IsNotExist(err) {
			config, err = toml.Load("")
		}
		if err != nil {
			return
		}
	}

	if config.HasPath([]string{accountname}) {
		return errors.New("Account already exists")
	}

	config.Set(accountname+".seed", seed)

	f, err := os.Create("config.toml")
	if err != nil {
		return
	}
	defer f.Close()
	_, err = config.WriteTo(f)
	return
}
