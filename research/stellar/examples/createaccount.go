package main

// In order to run this example install the dependencies:
// ```sh
// go get -d github.com/stellar/go
// go get -d github.com/stellar/go-xdr/xdr3
// go get github.com/lib/pq
// go get github.com/pelletier/go-toml
// ```

import (
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

	pair, err := keypair.Random()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Seed:", pair.Seed())
	log.Println("Address:", pair.Address())
	var config *toml.Tree

	config, err = toml.LoadFile("config.toml")
	if err != nil {
		if os.IsNotExist(err) {
			config, err = toml.Load("")
		}
		if err != nil {
			log.Fatal(err)
		}
	}

	if config.HasPath([]string{accountname}) {
		log.Fatal("Account already exists")
	}
	config.Set(accountname+".seed", pair.Seed())

	f, err := os.Create("config.toml")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	_, err = config.WriteTo(f)
	if err != nil {
		log.Fatal(err)
	}

}
