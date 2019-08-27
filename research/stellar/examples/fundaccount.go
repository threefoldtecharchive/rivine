package main

//go get github.com/segmentio/go-loggly
import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/pelletier/go-toml"
	"github.com/stellar/go/keypair"
)

func main() {
	var accountname string
	flag.StringVar(&accountname, "name", "default", " The name of the account to create")
	flag.Parse()
	config, err := toml.LoadFile("config.toml")

	if err != nil {
		log.Fatal(err)
	}
	seed := config.Get(accountname + ".seed")
	if seed == nil {
		log.Fatal("No such account")
	}
	newPK, err := keypair.Parse(seed.(string))
	pair, _ := newPK.(*keypair.Full)
	log.Println("Seed:", pair.Seed())
	log.Println("Address:", pair.Address())

	resp, err := http.Get("https://friendbot.stellar.org/?addr=" + pair.Address())
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
