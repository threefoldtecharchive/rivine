# Stellar account examples

## basic Stellar account examples

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

## Transfer assets

- transfer.go

Transfers an amount from an account to a destination address

## Rivine key conversion

Rivine uses default ed25519 keys, meaning a private key of 64 bytes from a 32 byte entropy (is actually the private key) and a public key size of 32 bytes.
Rivine hashes the public key along with the key alorithm to create an unlockhash. The address is then formed by concatenating the type, the hash and a checksum.

Stellar also uses uses default ed25519 keys.
A Stellar seed is just a base32 encoded concatatantion of a versionbyte, a 32byte private key and a checksum.
An address is the rawseed used to create an ed25519 keypair after which the versionbyte is concatenated with the public key and a checksum and base32 encoded.

The same 32 bytes can be used to create Rivine and Stellar keypairs.

It is possible to go from a Stellar account address to a Rivine "01"-address if they are created using the same private key (ed25519 entropy) without knowing the private keybut not the reverse.

Example code proving the above can be found in `convert.go`
