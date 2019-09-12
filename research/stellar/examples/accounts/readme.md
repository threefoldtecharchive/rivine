# Stellar examples

- createkeypair.go

Creates a new key pair,being a seed and an address a seed and  saves the seed to `config.toml`.

To run:
`go run createkeypair.go`

- activateaccount.go

Registers an address as an  account and funds it, either from an existing account or through friendbot.

- checkaccount.go

Loads the accounts from the saved `config.toml` and checks the balance.

To run:
`go run checkaccount.go`

- transfer.go

Tranfers an amount from an account to a destination address


