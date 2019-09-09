# Issuing a custom token

Stellar Guide : https://www.stellar.org/developers/guides/walkthroughs/custom-assets.html

## 1 Create a source account

```
go run ../accounts/createkeypair.go -name source
go run ../accounts/activateaccount.go -name source -friendbot
```


## 2 Create the issuing account

```
go run ../accounts/createkeypair.go -name issuer
go run ../accounts/activateaccount.go -name issuer -source source
```

## 3 Create the distribution account

```
go run ../accounts/createkeypair.go -name distributor
go run ../accounts/activateaccount.go -name distributor -source source
```

## 4 Create a trustline between the issuing account and the distribution account

It requires
- an asset code
- trust limit: max tokens

> The trust limit parameter limits the number of tokens the distribution account will be able to hold at once. It is recommended to either make this number larger than the total number of tokens expected to be available on the network or set it to be the maximum value (a total of max int64 stroops) that an account can hold.

`go run createtrustline.go  -source distributor`

## 5 Token creation

The Issuing account creates tokens and sends them to the distributing account.
`go run ../accounts/transfer.go -from issuer -destination <distributoraddress> -asset ROBTEST:<issueraddress>`

