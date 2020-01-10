# Issuing a custom token

Links:

- [Stellar Guide Custom asset Guide](https://www.stellar.org/developers/guides/walkthroughs/custom-assets.html)

In this walkthrough, we create a *BTC* asset on the Stellar testnet.

## 1 Create a source account

```sh
node account/create-key-pair.js
node account/fund-account.js --address={address}
```

## 2 Create the issuing account

```sh
node account/create-key-pair.js
node account/fund-account.js --address={address}
```

## 3 Create the distribution account

```sh
node account/create-key-pair.js
node account/fund-account.js --address={address}
```

## 4 Create a trustline between the issuing account and the distribution account

It requires

- an asset code
- trust limit: max tokens

> The trust limit parameter limits the number of tokens the distribution account will be able to hold at once. It is recommended to either make this number larger than the total number of tokens expected to be available on the network or set it to be the maximum value (a total of max int64 stroops) that an account can hold.

`node create-trustline.js --sourceKey={sourceKeyDistributionAccount} --issuerAddress={issuerAddress} --asset={assetCode} [-limit {limit}]`

## 5 Token creation

The Issuing account creates tokens and sends them to the distributing account.

`node account/transfer.js --sourceKey={issuerKey} --destinationAddress={distributoraddress} --amount=100 --asset=BTC:{issueraddress}`