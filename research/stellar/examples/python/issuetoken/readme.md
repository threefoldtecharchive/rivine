# Issuing a custom token

Links:

- [Stellar Guide Custom asset Guide](https://www.stellar.org/developers/guides/walkthroughs/custom-assets.html)

In this walkthrough, we create a *BTC* asset on the Stellar testnet.

## 1 Create a source account

```sh
python3 ../account/create-keypair.py
python3 ../account/fund-account.py --address <address>
```

## 2 Create the issuing account

```sh
python3 ../account/create-key-pair.py
python3 ../account/fund-account.py --address <address>
```

## 3 Create the distribution account

```sh
python3 ../account/create-key-pair.py
python3 ../account/fund-account.py --address <address>
```

## 4 Create a trustline between the issuing account and the distribution account

It requires

- an asset code
- trust limit: max tokens

> The trust limit parameter limits the number of tokens the distribution account will be able to hold at once. It is recommended to either make this number larger than the total number of tokens expected to be available on the network or set it to be the maximum value (a total of max int64 stroops) that an account can hold.

`python3 create-trustline.py --sourcekey <sourcekeyDistributionAccount> --issueraddress <issueraddress> --asset <assetCode> [-limit <limit>]`

## 5 Token creation

The Issuing account creates tokens and sends them to the distributing account.

`python3 ../account/transfer.py --sourcekey <issuerKey> --destinationAddress <distributoraddress> --amount=100 --asset=BTC:<issueraddress>`