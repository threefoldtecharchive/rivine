# Using the python examples

First of all, create a virtual environment

`python3 -m venv stellar`

Activate the environment

`source stellar/bin/activate`

Install dependencies

`pip3 install -r requirements.txt`

## Basic account examples
There is a more advanced example of [issuing a custom asset on testnet](issuetoken/readme.md)

### 1 Creating a keypair

This creates an `account`.

```sh
python3 account/create-key-pair.py
```

### 2 Funding an address

Funding an address of a generated keypair through the Stellar Friendbot.

```sh
python3 account/fund-account.py --address <address>
```

### 3 Checking balance of an address

Checks the balance of an address.

```sh
python3 account/check-account.py --address <address>
```

### 4 Activating (Funding) another account

Folowing funds another account, this account must exist in order to work.

Sourcekey is the secret key from which account another account will be funded.

```sh
python3 account/activate-account.py --sourcekey <address> --destinationaddress <destinationAddress>
```

### 5 Transfering assets to another account

Folowing transfers funds to another account, this account must exist in order to work.

Sourcekey is the secret key from which account funds will be transfered.

Amount must be a positive number, with a maximum of 7 digits after the decimal point.

Is the asset flag is provided in format: `code:issuer` then this asset will be transfered, if not provided the native XLM asset is used.

```sh
python3 account/transfer.py --sourceKey <address> --destinationAddress <destinationAddress> --amount <amount> --asset <code:issuer>
```
