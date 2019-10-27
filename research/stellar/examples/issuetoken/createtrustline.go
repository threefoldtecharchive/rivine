package main

//go get github.com/segmentio/go-loggly
import (
	"flag"
	"log"

	"github.com/stellar/go/clients/horizonclient"
	"github.com/stellar/go/network"
	"github.com/stellar/go/protocols/horizon"
	"github.com/stellar/go/txnbuild"
	"github.com/threefoldtech/rivine/research/stellar/examples/config"
)

func main() {
	var issuerAddress string
	var sourceAccountName string
	var assetCode string
	flag.StringVar(&issuerAddress, "issuer", "issuer", " The adress of the issuing account")
	flag.StringVar(&sourceAccountName, "source", "", " The name of the account to create the trustline")
	flag.StringVar(&assetCode, "asset", "", "The asset code")

	flag.Parse()
	if sourceAccountName == "" {
		flag.Usage()
		log.Fatalln("source is a required parameter")
	}
	if issuerAddress == "" {
		flag.Usage()
		log.Fatalln("issuer is a required parameter")
	}
	if assetCode == "" {
		flag.Usage()
		log.Fatalln("asset is a required parameter")
	}

	sourcePair, err := config.GetKeyPairFromConfig(sourceAccountName)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Seed:", sourcePair.Seed())
	log.Println("Address:", sourcePair.Address())

	//get the source account details
	sourceAccount, err := getAccountDetails(sourcePair.Address())
	if err != nil {
		log.Fatal(err)
	}

	op := txnbuild.ChangeTrust{
		Line:  txnbuild.CreditAsset{Code: assetCode, Issuer: issuerAddress},
		Limit: "10000",
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
	transactionID, err := tx.HashHex()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Transaction ID: ", transactionID)
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
