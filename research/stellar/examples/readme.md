# Stellar examples

- createaccount.go

Creates a new account, saves the seed to `config.toml` and funds it on the testnet using friendbot.


To run:
`go run createaccount.go`

- checkaccount.go

Loads the  accounts from the saved `config.toml` and checks the balance.

To run:
`go run checkaccount.go`
- transfer.go

Tranfers an amount from an account to a destination address
