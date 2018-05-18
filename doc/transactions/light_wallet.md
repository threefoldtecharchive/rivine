# Light Wallet

In the context of Rivine, we consider a light wallet to be a wallet which is managed using a light client.
A light client has no local daemon running and instead relies on a remote daemon.
This daemon could be one publicly available, such as standard explorer nodes,
but it could also be a private one.

These types of wallets are called light, because they don't do any validation or
other heavy-lifting processing. Such wallets also do not store all blocks,
and instead just process, and optionally store, the data that is relevant to the wallet.
Making these types of wallets light in terms of both processing as well as storage,
making them ideal for mobile wallets. It does however mean that trust is to be placed
in the remote daemon(s) used.

This document is to be used as a reference document,
explaining all you need in order to add support for a Rivine wallet,
to your light client.

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

> (!) Your wallet won't work if your public-private key pairs
> aren't [Ed25519][ed25519]-compatible !!!

In Rivine we use [BIP-39][bip39] to generate 24-word mnemonics from a seed,
and go back from such a mnemonic to such a seed. The 24-word count mentioned before,
means we expect random seeds of 32 bytes. This because of the assumption that it is easier
for humans to communicate and remember words, rather than 32 random hexadecimal characters.

Your wallet does not need to support [BIP-39][bip39], nor do your seeds need to be 32 bytes
in order to be compatible with the Rivine Protocol. We do however recommend that you
to support both these optional requirements, as to allow your users to use other wallet
implementations should they wish to do so.

## Creating Transactions

Creating a transaction is done so by assembling a transaction, encoding it in the JSON format and
passing it as the body in a POST call to the remote daemon using its REST API:

```plain
POST <daemon_addr>/transactionpool/transactions
```

> Note that this does mean that the remote daemon has to have the `Transaction Pool` module (`t`) enabled.
> See the CLI daemon's `modules` command for more information.

In order for your wallet to support this feature, you'll have to ensure that you can:

+ encode transactions in JSON format (see [/doc/transactions/transaction.md#json-encoding][jsonenc]);
+ create a signature for each input (see [the Signing Transactions chapter](#signing-transactions));
+ get all unspend outputs, as to be able to fund your transaction (see [the Getting Transactions chapter](#getting-transactions));

When your (light) wallet supports this feature, you'll users will be able to:

+ send coins;
+ send block stakes;
+ put arbitrary data on the blockchain;

If the POST call to that REST endpoint returns `200 OK`, it should mean that it was added to the transaction pool,
it does however not mean that it will actually end up on the blockchain, certainly not immediately.

### Signing Transactions

As discussed before, signatures are created using the [Ed25519][ed25519] signature algorithm.
This algorithm requires a hash as input. The hash (generated using [the blake2b 256 bit algorithm][blake2b])
uses as input the binary encoding of most of the properties of a transaction.

Therefore in order to be able to sign transactions you'll have to be able to:

+ Create a hash using [the blake2b (256 bit) algorithm][blake2b] (libraries are usually available to do this for you);
+ Binary-encode all necessary properties (see [/doc/transactions/transaction.md#binary-encoding][binenc]);
+ Convert raw binary data into a hex-encoded string (required as the [JSON-encoded][jsonenc] form of a signature is a hex-encoded string);

One or multiple signatures are required per fulfillment, which becomes an input when paired with a ParentID (the ID of an unspent output).
These fields are always labeled as `signature` (in singular or plural form). See the [JSON-encoding][jsonenc] for more information.

Without the ability to sign transactions, your wallet will not be able to [create transactions](#Creating Transactions).

### Example

You can find example code, labeled "offline transaction",
in [/doc/examples/offline_transaction/offline_transaction.go](/doc/examples/offline_transaction/offline_transaction.go).

This example demonstrates how one can prepare a transaction manually,
sign all inputs and finally create it using the POST call to `/transactionpool/transactions`.

It is the only example implementation bundled with this codebase.
As it is written in Golang and makes use of the Rivine library code,
it does mean that not everything in this example is implemented from scratch,
so you'll have to read parts of the Rivine codebase as well
if you want to understand this offline transaction example in its full scope.

## Getting Transactions

A wallet (identified by a seed), can have one or multiple addresses. That we already know.
Getting all transactions linked to one address can be done using the REST API of the remote daemon:

```plain
GET <daemon_addr>/explorer/hashes/<address>
```

> Note that this does mean that the remote daemon has to have the `Explorer` module (`e`) enabled.
> See the CLI daemon's `modules` command for more information.

It will give you response using the following JSON structure:

```javascript
{
    "hashtype": "unlockhash",   // should always be the value "unlockhash"
    "block": block,             // block structure can be ignored
    /////////////////////////////////////////////////////////////////////////////////////
    "blocks": [explorerBlock1, explorerBlock2, ...],
    /////////////////////////////////////////////////////////////////////////////////////
    "transaction": explorerTxn, // explorer transaction can be ignored as well
    /////////////////////////////////////////////////////////////////////////////////////
    "transactions": [explorerTxn1, explorerTxn2, ...],
}

// where each explorerTxnN is structured as follows:
{
    "id": id,                     // the id of the txn
    "height": uint64,             // height of the block this txn belongs to
    "parent": string,             // id of the block this txn belongs to
    /////////////////////////////////////////////////////////////////////////////////////
    "rawtransaction": txn,        // SEE /doc/transactions/transaction.md#json-encoding
                                  // to know how this transaction is encoded
    /////////////////////////////////////////////////////////////////////////////////////
    "hextransaction": hextxn,     // can be ignored for this purpose
    "coininputoutputs": io,       // can be ignored for this purpose
    /////////////////////////////////////////////////////////////////////////////////////
    // contains all ids of the coin outputs as found in the rawtransaction (using the same order).
    // if the outputid is the parentID of a coin input of any of the returned transactions,
    // it should be assumed the coin output is already spend, which can only happen once.
    "coinoutputids": [id1, id2, ...],
    /////////////////////////////////////////////////////////////////////////////////////
    "blockstakeinputoutputs": io, // can be ignored for this purpose
    /////////////////////////////////////////////////////////////////////////////////////
    // contains all ids of the block stake outputs as found in the rawtransaction (using the same order).
    // if the outputid is the parentID of a block stake input of any of the returned transactions,
    // it should be assumed the block stake output is already spend, which can only happen once.
    "blockstakeoutputids": ids,
}
// and where each explorerBlockN is structured as follows:
{
    "minerpayoutids": ids,        // can be ignored for this purpose
    "transactions": txns,         // can be ignored for this purpose
    /////////////////////////////////////////////////////////////////////////////////////
    "rawblock": block,            // SEE /doc/transactions/transaction.md#json-encoding
                                  // to know how this block is encoded
    /////////////////////////////////////////////////////////////////////////////////////
    "hexblock": string,           // can be ignored for this purpose
    // ... block facts properties, which also can be ignored for this purpose
}
```

Should your wallet support multiple addresses per seed,
you'll need to merge the results for all addresses together,
in order to find the complete information for that wallet.

### Getting Unconfirmed Transactions

When a transaction isn't part of a block yet,
it should exist in the transaction pool of the remote daemon.
The transaction pool is the pool of transactions which the daemon uses
as input in order to create a block. This pool can be filled both by the daemon
itself (using its REST API), as well as by peers (other daemons) it is connected to.

You can get all transactions in the transaction pool of a remote daemon using its REST API:

```plain
GET <daemon_addr>/transactionpool/transactions
```

> Note that this does mean that the remote daemon has to have the `Transaction Pool` module (`t`) enabled.
> See the CLI daemon's `modules` command for more information.

It will give you response using the following JSON structure:

```javascript
{
    // SEE /doc/transactions/transaction.md#json-encoding
    // to know how each txnN transaction is encoded
    "transactions": [txn1, txn2],
}
```

You'll have to filter the outputs and inputs using the address(es) of the wallet you're checking for,
let's call that wallet A. Each output targeted for wallet A means your unconfirmed balance gets increased,
while each input funded by wallet A (which can be known by tracing the parentID of each input)
means your unconfirmed balance gets decreased.

If you aim to be efficient in resources (including time) you should already have a mapping
of all unspent outputs by the time you call this function,
this way you can easily check if the input's parentID belongs to any of the unspent outputs of wallet A.

[bip39]: https://github.com/bitcoin/bips/blob/master/bip-0039.mediawiki
[Ed25519]: https://tools.ietf.org/html/rfc8032#section-5.1
[blake2b]: https://blake2.net

[jsonenc]: /doc/transactions/transaction.md#json-encoding
[binenc]: /doc/transactions/transaction.md#binary-encoding
