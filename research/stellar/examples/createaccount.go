package main

// In order to run this example install the dependencies:
// ```sh
// go get -d github.com/stellar/go
// go get -d github.com/stellar/go-xdr/xdr3
// go get github.com/lib/pq
// go get github.com/pelletier/go-toml
// ```

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/pelletier/go-toml"
	"github.com/stellar/go/keypair"
)

func main() {
	pair, err := keypair.Random()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Seed:", pair.Seed())
	log.Println("Address:", pair.Address())

	config, err := toml.Load("[default]\nseed=\"" + pair.Seed() + "\"")
	if err != nil {
		log.Fatal(err)
	}
	f, err := os.Create("config.toml")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	_, err = config.WriteTo(f)
	if err != nil {
		log.Fatal(err)
	}
	address := pair.Address()
	resp, err := http.Get("https://friendbot.stellar.org/?addr=" + address)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Friendbot response status:", resp.Status)
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(body))

}
