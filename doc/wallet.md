# Wallet

## Key Pairs

A wallet is identified by one or multiple private keys. Each private key is linked to a public key.
The public key is used to generate a wallet address, and thus a wallet can have multiple addresses,
as many as there are private-public key pairs.

All private keys of a single wallet are generated using the same seed,
that seed should be backed up by the user as to be able to recover the wallet
at any time, or even have multiple wallets using the same seed.

> In the standard Rivine CLI Wallet a [blake2b][blake2b] checksum is generated using
> the seed and the key (integral) index as input.
> This checksum (32 bytes) is used as the entropy for the generation of
> a [Ed25519][ed25519] private-public key pair.

The [Ed25519][ed25519] signature algorithm is the only algorithm currently supported by
the Rivine Protocol, for signatures provides as part of an input (spending an unspent output).
This means that the public keys of your wallet have to have a size of 32 bytes,
and the private keys (from which the public keys are derived)
have to have a size of 64 bytes. The produced signatures will have a size of 64 bytes as well.

### private key generation

The default Rivine wallet is a deterministical wallet meaning it derives keys from a single starting point known as a seed. The seed allows a user to easily back up and restore a wallet without needing any other information.
In Rivine we use [BIP39](bip39) with the default [bip39 spec wordlist](https://github.com/bitcoin/bips/tree/master/bip-0039) to generate 24-word mnemonics from a seed and go back from such a mnemonic to such a seed.
The 24-word count mentioned before,
means we expect random seeds of 32 bytes. This because of the assumption that it is easier
for humans to communicate and remember words, rather than 32 random hexadecimal characters
Seeds are  serialized into human-readable words in a seed phrase or mnemonic using [BIP39]
A wallet allows you to identify yourself by means of addresses,
using your private key(s) as proof of ownership.

## addresses and unlockhashes
An address is represented in Rivine as an unlock hash.
But just so we're clear, not all unlock hashes are wallet addresses.
Each unlock hash is prefixed with 2 bytes, indicating the type of hash (address) it represents.
Hence, wallet addresses are simply a type/category of possible addresses.
You can read more about unlock hashes (including the binary encoding in detail) in
[/doc/transactions/unlockhash.md](/doc/transactions/unlockhash.md).

```
                                                --------------------+
                                                unlockType (1 byte) |
 (ed25519)                                         +                |
publicKey *-----> binary *-----> blake2b *------> hash (32 bytes)   +----> hex_encoded
    *            encoding      256-bit hash        +                |         = unlock hash
    |                                           first 6 bytes of    |         (wallet address)
    |                                           blake2b_hash(hash)  |         78 bytes (2 * 1+32+6)
    |                                           --------------------+
    *
privateKey
 (ed25519) 
```

So an address is actually an unlock hash, which contains a hash, uniquely identifying you.
For a wallet address this (blake2b_256 crypto) hash is your public key in binary encoded format.
You can read more about the binary encoded format of public keys and anything
else used for encoding as part of transactions in
[/doc/transactions/transaction.md](/doc/transactions/transaction.md).

The private key is linked to the public key, but only the public key is publicly known.
Only the owner of the unlock hash (and thus the public key) should know and possess the private key.
The private key is used to make signatures, claiming tokens (mostly by using them),
which are targetted at the linked public key.

A wallet can have multiple keys and thus also have multiple addresses,
just like the wallet in your pockets might have multiple bank cards.

Spending coins, means actually spending unspend outputs.
Meaning that all coins you spend, are coins that have been somehow send to you earlier.
Sending in this context means that the output's unlockhash is one owned by you in your wallet.
When you spend those coins, you'll have to proof ownership by giving both your public key,
as well as a signature, signed/made using your private key.

Verification of ownership happens in 2 steps:

+ your signature can be verified, using your transaction's data
  and given public key and signature.
+ as the unlock hash contains the crypto hash of the public key,
  it can be verified that the given public key is the correct one.

Combining these 2, makes it easy and cheap for anyone to verify the ownership
and valid spending of unspend outputs.

A final note on the signature algorithm. While the rivine blockchain protocol allows
for any signature algorithm, we only support ed25519 for now.
This is important to take into account when developing your own (light) clients,
as your wallet will have to use the ed25519 algo as well,
in order to be able to sign and verify transactions.
 

[bip39]: https://github.com/bitcoin/bips/blob/master/bip-0039.mediawiki
[Ed25519]: https://tools.ietf.org/html/rfc8032#section-5.1
[blake2b]: https://blake2.net