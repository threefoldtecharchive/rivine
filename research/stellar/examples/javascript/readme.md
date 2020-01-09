# Using the javascript examples

First of all, install the dependencies

`npm install`

<!-- We consider an account is made up of 2 things: a secret key and an address. -->

## 1 Creating a keypair

This creates an `account`.

```sh
node account/create-key-pair.js
```

## 2 Funding an address

Funding an address of a generated keypair through the Stellar Friendbot.

```sh
node account/fund-account.js --address={address}
```

## 3 Checking balance of an address

Checks the balance of an address.

```sh
node account/check-account.js --address={address}
```

## 4 Activating (Funding) another account

Folowing funds another account, this account must exist in order to work.

Sourcekey is the secret key from which account another account will be funded.

```sh
node account/activate-account.js --sourceKey={address} --destinationAddress={destinationAddress}
```

## 5 Transfering assets to another account

Folowing transfers funds to another account, this account must exist in order to work.

Sourcekey is the secret key from which account funds will be transfered.

Amount must be a positive number, with a maximum of 7 digits after the decimal point.

Is the asset flag is provided in format: `code:issuer` then this asset will be transfered, if not provided the native XLM asset is used.

```sh
node account/transfer.js --sourceKey={address} --destinationAddress={destinationAddress} --amount={amount} --asset={code:issuer}
```

