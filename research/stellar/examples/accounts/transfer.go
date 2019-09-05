package main

import (
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/stellar/go/network"
	"github.com/stellar/go/txnbuild"

	"github.com/stellar/go/clients/horizon"
	"github.com/stellar/go/clients/horizonclient"
	"github.com/threefoldtech/rivine/research/stellar/examples/config"
)

func main() {
	var accountname string
	flag.StringVar(&accountname, "from", "default", "The name of the account to make the payment from")
	var destination string
	var assetString string
	flag.StringVar(&destination, "destination", "", "Destination address")
	flag.StringVar(&assetString, "asset", "", "The asset to transfer in case of non native XLM, format: `code:issuer`")
	var amount string
	flag.StringVar(&amount, "amount", "10", "The amount of topkens to transfer")
	flag.Parse()

	pair, err := config.GetKeyPairFromConfig(accountname)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("From Address", pair.Address())
	sourceAccount, err := getAccountDetails(pair.Address())
	if err != nil {
		log.Fatal(err)
	}

	var asset txnbuild.Asset
	if assetString == "" {
		asset = txnbuild.NativeAsset{}
	} else {
		assetparts := strings.SplitN(assetString, ":", 2)
		if len(assetparts) != 2 {
			log.Fatalln("Invalid asset format")
		}
		asset = txnbuild.CreditAsset{
			Code:   assetparts[0],
			Issuer: assetparts[1],
		}
	}

	payment := txnbuild.Payment{
		Destination: destination,
		Amount:      amount,
		Asset:       asset,
	}

	tx := txnbuild.Transaction{
		SourceAccount: &sourceAccount,
		Operations:    []txnbuild.Operation{&payment},
		Timebounds:    txnbuild.NewInfiniteTimeout(), // Use a real timeout in production!
		Network:       network.TestNetworkPassphrase,
	}

	txe, err := tx.BuildSignEncode(pair)
	if err != nil {
		panic(err)
	}

	// And finally, send it off to Stellar!
	resp, err := horizon.DefaultTestNetClient.SubmitTransaction(txe)
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

func getAccountDetails(address string) (account horizon.Account, err error) {
	client := horizonclient.DefaultTestNetClient
	ar := horizonclient.AccountRequest{AccountID: address}
	account, err = client.AccountDetail(ar)
	return
}
