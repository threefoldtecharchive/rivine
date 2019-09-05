package main

//go get github.com/segmentio/go-loggly
import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/pelletier/go-toml"
	"github.com/stellar/go/clients/horizonclient"
	"github.com/stellar/go/keypair"
	"github.com/stellar/go/network"
	"github.com/stellar/go/protocols/horizon"
	"github.com/stellar/go/txnbuild"
)

func main() {
	var accountname string
	var sourceAccountName string
	var fundUsingFriendbot bool
	flag.StringVar(&accountname, "name", "", " The name of the account to fund (required)")
	flag.StringVar(&sourceAccountName, "source", "default", " The name of the account to fund from")
	flag.BoolVar(&fundUsingFriendbot, "friendbot", false, "Fund the account through friendbot instead of from source")
	flag.Parse()
	if accountname == "" {
		flag.Usage()
		log.Fatalln("Name is a required parameter")
	}

	pair, err := getKeyPairFromConfig(accountname)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Seed:", pair.Seed())
	log.Println("Address:", pair.Address())

	if fundUsingFriendbot {
		err = fundThroughFriendbot(pair.Address())
		if err != nil {
			log.Fatal(err)
		}
		return
	}
	// Load the account to fund from
	sourcePair, err := getKeyPairFromConfig(sourceAccountName)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Funding from address:", sourcePair.Address())

	//get the funding account details
	sourceAccount, err := getAccountDetails(sourcePair.Address())
	if err != nil {
		log.Fatal(err)
	}
	//Create the `createaccount` transaction
	op := txnbuild.CreateAccount{
		Destination: pair.Address(),
		Amount:      "10",
	}

	tx := txnbuild.Transaction{
		SourceAccount: &sourceAccount,
		Operations:    []txnbuild.Operation{&op},
		Timebounds:    txnbuild.NewInfiniteTimeout(), // Use a real timeout in production!
		Network:       network.TestNetworkPassphrase,
	}

	txe, err := tx.BuildSignEncode(sourcePair)
	if err != nil {
		log.Fatal(err)
	}
	client := horizonclient.DefaultTestNetClient
	txSuccess, err := client.SubmitTransactionXDR(txe)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(txSuccess.TransactionSuccessToString())
}

func getAccountDetails(address string) (account horizon.Account, err error) {
	client := horizonclient.DefaultTestNetClient
	ar := horizonclient.AccountRequest{AccountID: address}
	account, err = client.AccountDetail(ar)
	return
}

func getKeyPairFromConfig(accountname string) (pair *keypair.Full, err error) {
	config, err := toml.LoadFile("config.toml")

	if err != nil {
		return
	}
	seed := config.Get(accountname + ".seed")
	if seed == nil {
		err = fmt.Errorf("account %s not found", accountname)
		return
	}
	newPK, err := keypair.Parse(seed.(string))
	pair, _ = newPK.(*keypair.Full)
	return
}

func fundThroughFriendbot(address string) (err error) {

	resp, err := http.Get("https://friendbot.stellar.org/?addr=" + address)
	if err != nil {
		return
	}
	log.Println("Friendbot response status:", resp.Status)
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	fmt.Println(string(body))
	return
}
