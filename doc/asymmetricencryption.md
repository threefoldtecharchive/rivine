# Using Rivine keys for asymmetric encryption

 As explained in the [wallet documentation](wallet.md), Rivine uses [Ed25519][ed25519] private-public key pairs. 

 Ed25519 is widely used for signing and  X25519 for asymmetric encryption. Derivation from from both private and public keys exist and is considered safe so that the same key pair can be used both for authenticated encryption (crypto_box) and for signatures (crypto_sign).

 A wallet address is basically a [blake2b_256](blake2b) hash of the ed25519public key in binary encoded format. This means that public key derivation to X25519 using an address is not possible since the public key is hashed to generate it.
 As such, one can not simply take a receivers address and use that to encrypt a message that only the receiver can decrypt.


[Ed25519]: https://tools.ietf.org/html/rfc8032#section-5.1
[blake2b]: https://blake2.net