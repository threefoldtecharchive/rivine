# monitor

monitor is a prototype for an ethereum oracle. It can perform the following functions:

- run as a regular node in the ethereum network, in light client mode
- have the ability to call functions on a specific contract
- be able to listen for and pickup (custom) events on a specific contract

## building

Run `go build` in this directory

## code

The [contract.go](./contract.go) file is generated from the [example contract](../testcontract.sol) using the `abigen` tool.
A deployed contract is available on the `Rinkeby` testnet. The prototype is configured to use this contract.

## Running

The prototype runs in `light mode`. The first time it is run, it will take a while for the chain to fully sync (seemingly roughly an hour).
An account is expected to run this example. One can be created using the default `geth` binary, and then imported. To import an account, pass the
`--account.json $ACCOUNTFILE` flag. This will import the account into the oracles keystore. You must also use the `--account.pass $ACCOUNTPASS` flag
with the password which was used to encrypt the account. After the first time, the account remains loaded (unless the keystore dir is removed/cleared), and
only the password needs to be provided.

The prototype will print information about any transaction it picks up being run on the contract.

If the `--trasnferTo $HEXADDR` is given, the prototype will try to send 10 tokens to this address. Note that this means that your account needs
at least 10 of the given ERC20 tokens + an amount of eth to pay the gas fee.