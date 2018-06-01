# Transaction

The main purpose of a transaction is to transfer coins and/or block stakes between addresses.
For each coin that is spend (registered as coin output), there must be one or multiple
(coin) inputs backing it up. These inputs must be previously-registered outputs,
which haven't been used as input yet. The same goes for block stakes, another kind of asset.

In order to create/send a transaction, one has to pay a transaction fee. These are registered
as part of a transaction and labeled as "Miner Fees". It is important to note that the total sum
of coin inputs MUST be equal to the total sum of coin outs and miner fees combined.
Block stakes are a separate asset, where the total sum of block stake inputs must
simply be equal to the total sum of block stake outputs.

As each output is backed by one or multiple inputs, it is not uncommon to have a too big
amount of input registered. If so, it is the convention to simply register the extra
amount as another output, but this time addressed to your own wallet.
This works very similar to change money you get in a supermarket by paying too much.

Each transaction needs to have sufficient Miner Fees as well as coin input(s) to back these fees up.
The rest of the requirements and structure of a transaction depends upon the version of that transaction.

## Versions

Each transaction has a version, which is to be decoded as the very first step.
Knowing the version, it can be deduced how to decode the rest of the data, if possible at all.

At the time of writing there are only two versions, `0x00` (0) and `0x01` (1).
Version 1 deprecates version 0, which is now considered legacy.
While version 0 is still accepted, it is no longer recommended.

Versions do however not always need to replace previous versions.
One other use case of versions could be to provide the option to have alternative
transaction structures, requiring their own requirements, validation and encoding.

Such alternative transactions are however up to blockchains to be implemented using the Rivine protocol,
(using the [`RegisterTransactionVersion`](https://godoc.org/github.com/rivine/rivine/types#RegisterTransactionVersion))
as Rivine keeps it at the v0 and v1 transactions for now,
which are only to be used for coin/blockstake transfers, optionally wth some (limited) Arbitrary Data attached to it.

## Relevant Source Files

For those interested, this document explains logic
implemented in the Golang reference Rivine implementation, and covers following source files:

+ [/types/transactions.go](/types/transactions.go)
+ [/types/unlockcondition.go](/types/unlockcondition.go)
+ [/types/unlockhash.go](/types/unlockhash.go)
+ [/types/signatures.go](/types/signatures.go)
+ [/types/currency.go](/types/currency.go)
+ [/types/timestamp.go](/types/timestamp.go)

The master version of the (public) Golang documentation for
the module of these files can be found at: <https://godoc.org/github.com/rivine/rivine/types>

> Should you have Rivine cloned onto a Golang-enabled machine available to you,
> you can render the same godoc information for whatever version you wish (after checking it out using `git checkout`),
> by running `godoc -http=:6060` in your terminal, after which you should be able to browse to the go documentation
> of your locally checked out Rivine version at: <http://localhost:6060/pkg/github.com/rivine/rivine/>,
> making the types module documentation available at: <http://localhost:6060/pkg/github.com/rivine/rivine/types>

## Index

+ Logic and rules of normal transactions:
  + [Send coins](#send-coins): explains a bit about the process of sending coins
  + [Arbitrary data](#arbitrary-data): explains a bit about what arbitrary data is and its limits
  + [Double Spend Rules](#double-spend-rules): explains a bit about how double spending is prevented
+ [JSON Encoding](#json-encoding):
  + [Introduction to JSON encoding](#introduction-to-json-encoding): why json encoding, when is it used
  + [JSON Encoding of Types](#json-encoding-of-types): explains how all parts of a transactions are JSON encoded
  + [JSON encoding of a v1 Transaction](#json-encoding-of-v1-transactions): a full and detailed example of a JSON-encoded v1 transaction
  + [JSON encoding of a v0 Transaction](#json-encoding-of-v0-transactions): a full and detailed example of a JSON-encoded v0 transaction
+ [Binary Encoding](#binary-encoding):
  + [Introduction to Binary Encoding](#introduction-to-binary-encoding): why binary encoding, when is it used
  + [Binary Encoding of Types](#binary-encoding-of-types): explains how all parts of a transactions are binary encoded
  + [Binary encoding of a v1 Transaction](#binary-encoding-of-v1-transactions): a full and detailed example of a binary-encoded v1 transaction
  + [Binary encoding of a v0 Transaction](#binary-encoding-of-v0-transactions): a full and detailed example of a binary-encoded v0 transaction
+ [Signing Transactions](#signing-transactions):
  + [Introduction to Signing Transactions](#introduction-to-signing-transaction)
  + [Signing a v1 Transaction](#signing-a-v1-transaction): explains how inputs are signed in v1 transactions
  + [Signing a v0 Transaction](#signing-a-v0-transaction): explains how inputs are signed in v0 transactions

## Logic and rules of normal transactions

This list is not an exclusive list, and some rules might be missing.
Regarding the logic and rules of normal transactions, please only
use the Golang codebase of Rivine as the only true documentation
about logic and rules of normal transactions.

### Arbitrary data

Arbitrary Data can be of any size. There is however a size limit on a Transaction and a block.
Keep in mind that the fee is depending on the size of a transaction, a blockcreator can ignore to add a transaction with small fees for a lot of small transactions with a summed up bigger fee (opportune).

Arbitrary data can be used to make verifiable announcements, or to have other
protocols sit on top of Rivine. The arbitrary data can also be used for soft
forks, and for protocol relevant information. Any arbitrary data is allowed by
the consensus.

### Double Spend Rules

When two conflicting transactions are seen, the first transaction is the only
one that is kept. If the blockchain reorganizes, the transaction that is kept
is the transaction that was most recently in the blockchain. This is to
discourage double spending, and enforce that the first transaction seen is the
one that should be kept by the network. Other conflicts are thrown out.

Transactions are currently included into blocks using a first-come first-serve
algorithm. Eventually, transactions will be rejected if the fee does not meet a
certain minimum. For the near future, there are no plans to prioritize
transactions with substantially higher fees. Other mining software may take
alternative approaches.

## JSON Encoding

### Introduction to JSON encoding

JSON encoding is used as the encoding protocol for any data that leaves and enters the daemon via the REST API.

A Transaction, no matter the version, is always JSON-encoded as follows:

```javascript
{
    "version": byte, // byte identifying the (transaction) version
    "data": data // format of data depends upon the (transaction) version
}
```

A Transaction is given as part of a body, which is always JSON-encoded as follows:

```javascript
{
    "parentid": blockID, // ID of the previous block, the nil-block if this block is the genesis block
    "timestamp": uint64, // creation Unix Epoch timestamp of the block
    "pobsindexes": { // indices and values required for the POBS protocol.
        "BlockHeight": blockHeight,
        "TransactionIndex": uint64,
        "OutputIndex": uint64,
    },
    "minerpayouts": [
        {
            "value": currency,
            "unlockhash": unlockHash, // see /doc/transactions/unlockhash.md to learn how this is encoded
        },
    ],
    "transactions": [txn1, ...] // where each txn1 is encoded as discussed earlier
}
```

### JSON Encoding of Types

The encoding of primitives such as integers, booleans and strings behave just the same
as in any other JSON encoding you might have seen before.

A currency (coins/block stakes), represented in memory as a Big Integer, is JSON-encoded into a string, representing
the number in the human-readable 10-base format you're used to. (e.g. `1234` is represented as `"1234"`).

Arbitrary data, which in memory is a raw byte slice, is base64-encoded into a string when part of a JSON structure.

Any kind of Identifier type (e.g. CoinOutputID][conid], [BlockStakeOutputID][bsoid], OutputID][outid], [TransactionID][txnid] and [BlockID][blkid]) is hex encoded into a string with a static length of 64 characters, when part of a JSON structure.

Pointer types are only encoded when not nil, in which case their value type is used to encode (except when [the `json.Marshaler` interface](https://godoc.org/encoding/json#Marshaler) (in the Rivine Golang lib) is supported by that pointer type).

Types are always encoded using custom logic if [the `json.Marshaler` interface]((https://godoc.org/encoding/json#Marshaler))
is supported (in the Rivine Golang lib).

Properties of structures are only encoded if they are marked as public
(in golang a property is public if it starts with a capital letter).
The exception is again if the structure type implements [the `json.Marshaler` interface](https://godoc.org/encoding/json#Marshaler).

Arrays and slices are encoded as JSON arrays, where the elements are than encoded one by one depending upon their type.

Anything that uses the `ByteSlice` Golang type, is also encoded as a hex-encoded string.
Example of a `ByteSlice` are: signatures, secrets, keys and hashed secrets.

When decoding JSON-encoded data, it is assumed that you know the structure, and it will use your provided value's type in order to know how to decode it. The exception to this rule again (when using the Rivine Go Library) is that if the given value's type implements [the `json.Unmarshaler` interface](https://godoc.org/encoding/json#Unmarshaler), it is decoded using the custom defined logic.

When you know these rules (and exceptions) it should be trivial to deduce how any given type is JSON-encoded and decoded,
given that you know how the type is defined and which interfaces it implements.

### JSON encoding of v1 Transactions

Before we show a full example of a v1 transaction in JSON-encoded form,
with some possible input and output types,
let's see how inputs and outputs are JSON-encoded, as the encoding of those are different
in v1 transactions compared to v0 transactions.

Understanding the following [inputs](#json-encoding-of-inputs-in-v1-transactions) and [outputs](#json-encoding-of-outputs-in-v1-transactions) chapters,
together with [the JSON Encoding of Types chapter](#json-encoding-of-types),
should make you able to understand the [Example of a JSON-encoded v1 Transaction](#example-of-a-json-encoded-v1-transaction),
without a sweat.

#### JSON Encoding of Inputs in v1 Transactions

Block stake- and coin- inputs are optional, and both are encoded in the same format:

```javascript
{
    // coin/blockstake output ID, crypto hash (blake2b, 256 bit), required, hex-encoded,
    // linking the output that is planned to be spend by this input
    "parentid": "dfd23dfd23dfd23dfd23dfd23dfd23dfd23dfd23dfd23dfd23dfd23dfd23dfde",
    "fulfillment": {
        "type": byte,   // fulfillment type, byte, supported range: [1,2], required
                        // `1` = SingleSignature, `2` = AtomicSwap
        "data": data, // structure of the fulfillment depends upon the sibling type byte
    },
}
```

The `fulfillment`'s `type` defines the JSON-encoded format of the `fulfillment`'s `data` property.

##### JSON Encoding of a SingleSignatureFulfillment

When the `fulfillment`'s `type` equals `1`, it indicates a SingleSignature Fulfillment,
with as consequence that the fulfillment will have the following JSON-encoded format:

```javascript
{
    "type": 1, // indicates a SingleSignature Fulfillment
    "data": {
        // public key, required, format: `<algorithmSpecifier>:<key>`, where <key> is hex-encoded
        // and which byte-size is fixed but dependend upon the <algorithmSpecifier>,
        // <algorithmSpecifier> can currently only be `"ed25519"`
        "publickey": "ed25519:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
        // signature, byte-size is fixed but dependend upon the <algorithmSpecifier>,
        // when <algorithmSpecifier> equals `"ed25519"` the byte size is 64, required, hex-encoded
        "signature": "abcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefab"
    }
}
```

##### JSON Encoding of an AtomicSwapFulfillment

When the `fulfillment`'s `type` equals `2`, it indicates an AtomicSwap Fulfillment,
with as consequence that the fulfillment will usually have the following JSON-encoded format:

```javascript
{
    "type": 2, // indicates an AtomicSwap Fulfillment (in the new/v1 format)
    "data": {
        // public key, required, format: `<algorithmSpecifier>:<key>`, where <key> is hex-encoded
        // and which byte-size is fixed but dependent upon the <algorithmSpecifier>,
        // <algorithmSpecifier> can currently only be `"ed25519"`
        "publickey": "ed25519:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
        // signature, byte-size is fixed but dependend upon the <algorithmSpecifier>,
        // when <algorithmSpecifier> equals `"ed25519"` the byte size is 64, required, hex-encoded
        "signature": "abcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefab",
        // secret, fixed size, 32 bytes, hex-encoded
        // optional, and doesn't have to be given if this is a refund rather than a claim
        "secret": "def789def789def789def789def789dedef789def789def789def789def789de"
    }
}
```

It can however also have the following legacy/v0 JSON-encoded format:

```javascript
{
    "type": 2, // indicates an AtomicSwap Fulfillment (in the new/v1 format)
    "data": {
        // sender's unlock hash, required, hex-encoded, fixed-size
        "sender": "01654f96b317efe5fd6cd8ba1a394dce7b6ebe8c9621d6c44cbe3c8f1b58ce632a3216de71b23b",
        // receiver's unlock hash, required, hex-encoded, fixed-size
        "receiver": "01e89843e4b8231a01ba18b254d530110364432aafab8206bea72e5a20eaa55f70b1ccc65e2105",
        // hashed secret, fixed size, 32 bytes, hex-encoded, sha256(secret)
        "hashedsecret": "abc543defabc543defabc543defabc543defabc543defabc543defabc543defa",
        // time lock, unix epoch timestamp (in seconds), 64-bit unsigned integer, required
        "timelock": 1522068743,
        // public key, required, format: `<algorithmSpecifier>:<key>`, where <key> is hex-encoded
        // and which byte-size is fixed but dependent upon the <algorithmSpecifier>,
        // <algorithmSpecifier> can currently only be `"ed25519"`
        "publickey": "ed25519:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
        // signature, byte-size is fixed but dependend upon the <algorithmSpecifier>,
        // when <algorithmSpecifier> equals `"ed25519"` the byte size is 64, required, hex-encoded
        "signature": "abcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefab",
        // secret, fixed size, 32 bytes, hex-encoded
        // optional, and doesn't have to be given if this is a refund rather than a claim
        "secret": "def789def789def789def789def789dedef789def789def789def789def789de"
}
```

Where the legacy/v0 format is meant to fulfill legacy UnlockHash-based conditions (ConditionType `1` with UnlockType `2`).
While that format can also fulfill the new/v1 AtomicSwapCondition (ConditionType `2`),
when fulfilling that condition it makes more sense to use the new/v1 format, as to avoid info-duplication.

##### JSON Encoding of a MultiSignatureFulfillment

The FulfillmentTypeMultiSignature (`3`) identifies a MultiSignatureFulfillment
and is json-encoded in the following format:

```javascript
{
    "type": 3, // indicates a MultiSignatureFulfillment
    "data": {
        // on a local level, at least one pair is required,
        // but in order to fulfill a MultiSignatureCondition (ConditionType `4`),
        // it needs to contain at least as many pairs as defined by the `condition.minimumsignaturecount`.
        "pairs": [
            {
                // public key, required, format: `<algorithmSpecifier>:<key>`, where <key> is hex-encoded
                // and which byte-size is fixed but dependent upon the <algorithmSpecifier>,
                // <algorithmSpecifier> can currently only be `"ed25519"`
                "publickey": "ed25519:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
                // signature, byte-size is fixed but dependend upon the <algorithmSpecifier>,
                // when <algorithmSpecifier> equals `"ed25519"` the byte size is 64, required, hex-encoded
                "signature": "abcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefab"
            },
            {
                // public key, required, format: `<algorithmSpecifier>:<key>`, where <key> is hex-encoded
                // and which byte-size is fixed but dependent upon the <algorithmSpecifier>,
                // <algorithmSpecifier> can currently only be `"ed25519"`
                "publickey": "ed25519:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
                // signature, byte-size is fixed but dependend upon the <algorithmSpecifier>,
                // when <algorithmSpecifier> equals `"ed25519"` the byte size is 64, required, hex-encoded
                "signature": "abcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefab"
            }
        ]
    }
}
```

The public key of each listed pair in the MultiSignatureFulfillment,
should be listed in the unlockhashes listed as part of the MultiSignatureCondition this fulfillment is to fulfill.
That is to be said, if the public key is turned into a PubKeyUnlockHash, the resulting unlockhash should be present in the condition's unlockhashes.

#### JSON Encoding of Outputs in v1 Transactions

Block stake- and coin- outputs are optional, and both are encoded in the same format:

```javascript
{
    // currency value (in smallest unit), big integer as string, required, positive
    "value": "10000",
    "condition": {
        "type": byte,   // condition type, byte, supported range: [0,3], required
                        // `0` = NilCondition, `1` = UnlockHashCondition,
                        // `2` = AtomicSwapCondition, `3` = TimeLockCondition
                        // `4` = MultiSignatureCondition
        "data": data, // structure of the fulfillment depends upon the sibling type byte
    },
}
```

The `condition`'s `type` defines the JSON-encoded format of the `condition`'s `data` property.

##### JSON Encoding of a NilCondition

The ConditionTypeNil (`0`) identifies a NilCondition,
is the default condition type and always has the following format:

```javascript
{
    "type": 0, // indicates a NilCondition
    "data": {},
}
```

As it is the default condition, you could however also just represent it
as an empty JSON struct:

```javascript
{}
```

An output using a NilCondition can be spend anyone who owns a wallet address,
simply by signing the input with a single signature, using a private key of choice.

##### JSON Encoding of an UnlockHashCondition

The ConditionTypeUnlockHash (`1`) identifies an UnlockHashCondition
and is json-encoded in the following format:

```javascript
{
    "type": 1, // indicates an UnlockHashCondition
    "data": {
        // output's unlock hash, specifier output will be locked to that hash
        // the `0x01` prefix in this instance indicates that it is a PubKey Unlock Hash (a wallet address)
        "unlockhash": "01a6a6c5584b2bfbd08738996cd7930831f958b9a5ed1595525236e861c1a0dc353bdcf54be7d8"
    }
}
```

The details of the JSON-encoding of an unlock hash are described in
[/doc/transactions/unlockhash.md](/doc/transactions/unlockhash.md).

Such condition can be fulfilled by proving the ownership of the private key,
which is linked to the public key from which this unlock hash is derived.

The signature, given as part of the fulfillment, is the proof.
See [the Signing Transactions chapter](#signing-transactions) for more information.

##### JSON Encoding of an AtomicSwapCondition

The ConditionTypeAtomicSwap (`2`) identifies an AtomicSwapCondition
and is json-encoded in the following format:

```javascript
{
    "type": 2, // indicates an AtomicSwapCondition
    "data": {
        // sender's unlock hash, required, hex-encoded, fixed-size
        "sender": "01654f96b317efe5fd6cd8ba1a394dce7b6ebe8c9621d6c44cbe3c8f1b58ce632a3216de71b23b",
        // receiver's unlock hash, required, hex-encoded, fixed-size
        "receiver": "01e89843e4b8231a01ba18b254d530110364432aafab8206bea72e5a20eaa55f70b1ccc65e2105",
        // hashed secret, fixed size, 32 bytes, hex-encoded, sha256(secret)
        "hashedsecret": "abc543defabc543defabc543defabc543defabc543defabc543defabc543defa",
        // time lock, unix epoch timestamp (in seconds), 64-bit unsigned integer, required
        "timelock": 1522068743
    }
}
```

How such condition can be fulfilled depends upon the height or timestamp of the last block in the chain.
If the contract is active, the condition can be fulfilled by proving the ownership of the private key,
which is linked to the public key from which the receiver unlock hash is derived. Otherwise the condition
can be fulfilled by proving the ownership of the private key, which is linked to the public key from which
the sender unlock hash is derived.

The signature, given as part of the fulfillment, is the proof.
See [the Signing Transactions chapter](#signing-transactions) for more information.

##### JSON Encoding of a TimeLockCondition

The ConditionTypeTimeLock (`3`) identifies a TimeLockCondition
and is json-encoded in the following format:

```javascript
{
    "type": 3, // indicates a TimeLockCondition
    "data": {
        // locktime identifies the height or timestamp until which this output is locked,
        // meaning that as long as the last block's height/timestamp is less than this value,
        // the output cannot be spend.
        //
        // If the locktime is less then 500 milion it is to be assumed to be identifying a block height,
        // otherwise it identifies a unix epoch timestamp in seconds;
        "locktime": 500000000,
        // internal condition this TimeLockCondition wraps around,
        // meaning that in order to fulfill this TimeLockCondition,
        // this internal condition will have to be explicitly fulfilled on top
        // of the implicit locktime fulfillment.
        "condition": {
            // Supported types are:
            //   + UnlockHashCondition (ConditionType `1` with a UnlockType of `1`)
            //   + MultiSignatureCondition (ConditionType `4`)
            "type": conditionType,
            "data": data, // data depends upon the ConditionType
                          // defined in the sibling type property
        }
    }
}
```

Such condition is to be fulfilled in 2 parts:

+ The first part is done implicitly, by ensuring the last block height or time is less than the height/time specified as LockTime.
+ If the first part has been fulfilled, the internal condition has to be fulfilled explicitly, by giving a fulfillment which is able to fulfill the internal condition, which is part of the TimeLockCondition and encoded as the very last thing;

The constant which defines whether a LockTime is a Block Height or a Unix Epoch Timestamp in seconds,
is `LockTimeMinTimestampValue` and is documented in <https://godoc.org/github.com/rivine/rivine/types#pkg-constants>.

##### JSON Encoding of a MultiSignatureCondition

The ConditionTypeMultiSignature (`4`) identifies a MultiSignatureCondition
and is json-encoded in the following format:

```javascript
{
    "type": 4, // indicates a MultiSignatureCondition
    "data": {
        // lists all unlock hashes which are authorised to
        // spend this output by signing off
        "unlockhashes": [ // at least one unlockhash is required, this array cannot be empty or non-defined
            "01e89843e4b8231a01ba18b254d530110364432aafab8206bea72e5a20eaa55f70b1ccc65e2105",
            "01a6a6c5584b2bfbd08738996cd7930831f958b9a5ed1595525236e861c1a0dc353bdcf54be7d8"
        ],
        // defines the amount of signatures required in order to spend this output,
        // note that this number must be at least one, and it cannot be greater
        // than the total amount of unlockhashes listed in the sibliging "unlockhashes" property
        "minimumsignaturecount": 2 // can be anything as long as it is `0 > n >= len(unlockhashes)`
    }
}
```

Such condition can only be fulfilled by a `MultiSignatureFulfillment` (FulfillmentType `3`),
and the condition is fulfilled in 3 steps:

1. First it is ensured that the minimumsignaturecount property is valid,
   and that there are enough key-signature pairs present in the fulfillment
2. Then is ensured that all public keys listed in the fulfillment are authorized to do so,
   checking this is as easy as ensuring that the `PubKeyUnlockHash` of each public key, is present
   in the `unlockhashes` list of this condition
3. Finally all signatures are checked against the paired public key and the given transaction,
   within the given Fulfillment Context;

#### Example of a JSON-encoded v1 Transaction

The JSON encoding of a v1 Transaction can be explained best using an example:

```javascript
{
    "version": 1, // version ID, byte, always `1` for v1 transaction, required as this value is `0` by default
    "data": { // actual transaction data, required
        "coininputs": [ // coin inputs, optional
            {
                // coin output ID, crypto hash (blake2b, 256 bit), required, hex-encoded
                "parentid": "abcdef012345abcdef012345abcdef012345abcdef012345abcdef012345abcd",
                // fulfillment, used to fulfill the condition of the output identified by the above parentID
                "fulfillment": {
                    // fulfillment type, byte, supported range: [1,2], required
                    // `1` = SingleSignature, `2` = AtomicSwap
                    "type": 1,
                    "data": {   // actual fulfillment data, requirement depends upon sibling type property,
                                // in this instance it represents a SingleSignature Fulfillment
                        // public key, required, format: `<algorithmSpecifier>:<key>`, where <key> is hex-encoded
                        // and which byte-size is fixed but dependend upon the <algorithmSpecifier>,
                        // <algorithmSpecifier> can currently only be `"ed25519"`
                        "publickey": "ed25519:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
                        // signature, byte-size is fixed but dependend upon the <algorithmSpecifier>,
                        // when <algorithmSpecifier> equals `"ed25519"` the byte size is 64, required, hex-encoded
                        "signature": "abcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefab"
                    }
                }
            },
            {
                // coin output ID, crypto hash (blake2b, 256 bit), required, hex-encoded
                "parentid": "012345defabc012345defabc012345defabc012345defabc012345defabc0123",
                // fulfillment, used to fulfill the condition of the output identified by the above parentID
                "fulfillment": {
                    // fulfillment type, byte, supported range: [1,2], required
                    // `1` = SingleSignature, `2` = AtomicSwap
                    "type": 2,
                    "data": {   // actual fulfillment data, requirement depends upon sibling type property,
                                // in this instance it represents an AtomicSwap Fulfillment in the new/v1 format
                        // public key, required, format: `<algorithmSpecifier>:<key>`, where <key> is hex-encoded
                        // and which byte-size is fixed but dependend upon the <algorithmSpecifier>,
                        // <algorithmSpecifier> can currently only be `"ed25519"`
                        "publickey": "ed25519:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
                        // signature, byte-size is fixed but dependend upon the <algorithmSpecifier>,
                        // when <algorithmSpecifier> equals `"ed25519"` the byte size is 64, required, hex-encoded
                        "signature": "abcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefab",
                        // secret, fixed size, 32 bytes, hex-encoded
                        "secret": "def789def789def789def789def789dedef789def789def789def789def789de"
                    }
                }
            },
            {
                // coin output ID, crypto hash (blake2b, 256 bit), required, hex-encoded
                "parentid": "045645defabc012345defabc012345defabc012345defabc012345defabc0123",
                // fulfillment, used to fulfill the condition of the output identified by the above parentID
                "fulfillment": {
                    // fulfillment type, byte, supported range: [1,2], required
                    // `1` = SingleSignature, `2` = AtomicSwap
                    "type": 2,
                    "data": {   // actual fulfillment data, requirement depends upon sibling type property,
                                // in this instance it represents an AtomicSwap Fulfillment in the legacy/v0 format
                        // sender's unlock hash, required, hex-encoded, fixed-size
                        "sender": "01654f96b317efe5fd6cd8ba1a394dce7b6ebe8c9621d6c44cbe3c8f1b58ce632a3216de71b23b",
                        // receiver's unlock hash, required, hex-encoded, fixed-size
                        "receiver": "01e89843e4b8231a01ba18b254d530110364432aafab8206bea72e5a20eaa55f70b1ccc65e2105",
                        // hashed secret, fixed size, 32 bytes, hex-encoded, sha256(secret)
                        "hashedsecret": "abc543defabc543defabc543defabc543defabc543defabc543defabc543defa",
                        // time lock, unix epoch timestamp (in seconds), 64-bit unsigned integer, required
                        "timelock": 1522068743,
                        // public key, required, format: `<algorithmSpecifier>:<key>`, where <key> is hex-encoded
                        // and which byte-size is fixed but dependend upon the <algorithmSpecifier>,
                        // <algorithmSpecifier> can currently only be `"ed25519"`
                        "publickey": "ed25519:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
                        // signature, byte-size is fixed but dependend upon the <algorithmSpecifier>,
                        // when <algorithmSpecifier> equals `"ed25519"` the byte size is 64, required, hex-encoded
                        "signature": "abcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefab",
                        // secret, fixed size, 32 bytes, hex-encoded
                        "secret": "def789def789def789def789def789dedef789def789def789def789def789de"
                    }
                }
            }
        ],
        "coinoutputs": [ // coin outputs, optional
            {
                "value": "3", // currency value (in smallest unit), big integer as string, required, positive
                // condition, used to define what condition has to be fulfilled in order to spend this output
                "condition": {
                    // condition type, byte, supported range: [0,3], required
                    // `0` = NilCondition, `1` = UnlockHashCondition, `2` = AtomicSwapCondition
                    // `3` = TimeLockCondition
                    "type": 1,
                    "data": {    // actual condition data, requirement depends upon sibling type property,
                                // in this instance it represents an UnlockHashCondition
                        // output's unlock hash, specifier output will be locked to that hash
                        // the `0x01` prefix in this instance indicates that it is a PubKey Unlock Hash (a wallet address)
                        "unlockhash": "0142e9458e348598111b0bc19bda18e45835605db9f4620616d752220ae8605ce0df815fd7570e"
                    }
                }
            },
            {
                "value": "5", // currency value (in smallest unit), big integer as string, required, positive
                // condition, used to define what condition has to be fulfilled in order to spend this output
                "condition": {
                    // condition type, byte, supported range: [0,3], required
                    // `0` = NilCondition, `1` = UnlockHashCondition, `2` = AtomicSwapCondition
                    // `3` = TimeLockCondition
                    "type": 1,
                    "data": {   // actual condition data, requirement depends upon sibling type property,
                                // in this instance it represents an UnlockHashCondition
                        // output's unlock hash, specifier output will be locked to that hash
                        // the `0x01` prefix in this instance indicates that it is a PubKey Unlock Hash (a wallet address)
                        "unlockhash": "01a6a6c5584b2bfbd08738996cd7930831f958b9a5ed1595525236e861c1a0dc353bdcf54be7d8"
                    }
                }
            },
            {
                "value": "8", // currency value (in smallest unit), big integer as string, required, positive
                // condition, used to define what condition has to be fulfilled in order to spend this output
                "condition": {
                    // condition type, byte, supported range: [0,3], required
                    // `0` = NilCondition, `1` = UnlockHashCondition, `2` = AtomicSwapCondition
                    // `3` = TimeLockCondition
                    "type": 1,
                    "data": {   // actual condition data, requirement depends upon sibling type property,
                                // in this instance it represents an UnlockHashCondition
                        "unlockhash": "02a24c97c80eeac111aa4bcbb0ac8ffc364fa9b22da10d3054778d2332f68b365e5e5af8e71541"
                    }
                }
            },
            {
                "value": "13", // currency value (in smallest unit), big integer as string, required, positive
                // condition, used to define what condition has to be fulfilled in order to spend this output
                "condition": {
                    // condition type, byte, supported range: [0,3], required
                    // `0` = NilCondition, `1` = UnlockHashCondition, `2` = AtomicSwapCondition
                    // `3` = TimeLockCondition
                    "type": 2,
                    "data": {   // actual condition data, requirement depends upon sibling type property,
                                // in this instance it represents an AtomicSwapCondition
                        // sender's unlock hash, required, hex-encoded, fixed-size
                        "sender": "01654f96b317efe5fd6cd8ba1a394dce7b6ebe8c9621d6c44cbe3c8f1b58ce632a3216de71b23b",
                        // receiver's unlock hash, required, hex-encoded, fixed-size
                        "receiver": "01e89843e4b8231a01ba18b254d530110364432aafab8206bea72e5a20eaa55f70b1ccc65e2105",
                        // hashed secret, fixed size, 32 bytes, hex-encoded, sha256(secret)
                        "hashedsecret": "abc543defabc543defabc543defabc543defabc543defabc543defabc543defa",
                        // time lock, unix epoch timestamp (in seconds), 64-bit unsigned integer, required
                        "timelock": 1522068743
                    }
                }
            }
        ],
        "blockstakeinputs": [ // block stake inputs, optional
            {
                // block stake output ID, crypto hash (blake2b, 256 bit), required, hex-encoded
                "parentid": "dfd23dfd23dfd23dfd23dfd23dfd23dfd23dfd23dfd23dfd23dfd23dfd23dfde",
                // fulfillment, used to fulfill the condition of the output identified by the above parentID
                "fulfillment": {
                    // fulfillment type, byte, supported range: [1,2], required
                    // `1` = SingleSignature, `2` = AtomicSwap
                    "type": 1,
                    "data": {   // actual fulfillment data, requirement depends upon sibling type property,
                                // in this instance it represents a SingleSignature Fulfillment
                        // public key, required, format: `<algorithmSpecifier>:<key>`, where <key> is hex-encoded
                        // and which byte-size is fixed but dependend upon the <algorithmSpecifier>,
                        // <algorithmSpecifier> can currently only be `"ed25519"`
                        "publickey": "ed25519:ef1234ef1234ef1234ef1234ef1234ef1234ef1234ef1234ef1234ef1234ef12",
                        // signature, byte-size is fixed but dependend upon the <algorithmSpecifier>,
                        // when <algorithmSpecifier> equals `"ed25519"` the byte size is 64, required, hex-encoded
                        "signature": "01234def01234def01234def01234def01234def01234def01234def01234def01234def01234def01234def01234def01234def01234def01234def01234def"
                    }
                }
            }
        ],
        "blockstakeoutputs": [ // block stake outputs, optional
            {
                "value": "4", // currency value (in smallest unit), big integer as string, required, positive
                // condition, used to define what condition has to be fulfilled in order to spend this output
                "condition": {
                    // condition type, byte, supported range: [0,3], required
                    // `0` = NilCondition, `1` = UnlockHashCondition, `2` = AtomicSwapCondition
                    // `3` = TimeLockCondition
                    "type": 1,
                    "data": {   // actual condition data, requirement depends upon sibling type property,
                                // in this instance it represents an UnlockHashCondition
                        // output's unlock hash, specifier output will be locked to that hash
                        // the `0x01` prefix in this instance indicates that it is a PubKey Unlock Hash (a wallet address)
                        "unlockhash": "01a6a6c5584b2bfbd08738996cd7930831f958b9a5ed1595525236e861c1a0dc353bdcf54be7d8"
                    }
                }
            }
        ],
        // miner fees, list of currency values, at least 1 required
        "minerfees": ["1", "2", "3"],
        // arbitrary data, optional, base64 encoded
        "arbitrarydata": "ZGF0YQ=="
    }
}
```

### JSON encoding of v0 Transactions

Before we show a full example of a v0 transaction in JSON-encoded form,
with some possible input and output types,
let's see how inputs and outputs are JSON-encoded, as the encoding of those are different
in v0 transactions compared to v1 transactions.

Understanding the following [inputs](#json-encoding-of-inputs-in-v0-transactions) and [outputs](#json-encoding-of-outputs-in-v0-transactions) chapters,
together with [the JSON Encoding of Types chapter](#json-encoding-of-types),
should make you able to understand the [Example of a JSON-encoded v0 Transaction](#example-of-a-json-encoded-v0-transaction),
without a sweat.

#### JSON encoding of Inputs in v0 Transactions

Block stake- and coin- inputs are optional, and both are encoded in the same format:

```javascript
{
    // coin/blockstake output ID, crypto hash (blake2b, 256 bit), required, hex-encoded,
    // linking the output that is planned to be spend by this input
    "parentid": "dfd23dfd23dfd23dfd23dfd23dfd23dfd23dfd23dfd23dfd23dfd23dfd23dfde",
    "unlocker": {
        "type": byte,    // input lock, byte, supported range: [1,2], required
                        // `1` = SingleSignature, `2` = AtomicSwap
        "condition": condition, // structure of the condition depends upon the sibling type byte
        "fulfillment": fulfillment, // structure of the fulfillment depends upon the sibling type byte
    },
}
```

The `unlocker`'s `type` defines the JSON-encoded format of the `condition` and `fulfillment`.

##### JSON Encoding of a SingleSignature InputLock

When the `unlocker`'s `type` equals `1`, it indicates a SingleSignature InputLock,
with as consequence that the unlocker will have the following JSON-encoded format:

```javascript
{
    "type": 1, // indicates a SingleSignature Input Lock
    "condition": {
        // public key, required, format: `<algorithmSpecifier>:<key>`, where <key> is hex-encoded
        // and which byte-size is fixed but dependend upon the <algorithmSpecifier>,
        // <algorithmSpecifier> can currently only be `"ed25519"`
        "publickey": "ed25519:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"
    },
    "fulfillment": {
        // signature, byte-size is fixed but dependend upon the <algorithmSpecifier>,
        // when <algorithmSpecifier> equals `"ed25519"` the byte size is 64, required, hex-encoded
        "signature": "abcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefab"
    }
}
```

##### JSON Encoding of an AtomicSwap InputLock

When the `unlocker`'s `type` equals `2`, it indicates an AtomicSwap InputLock,
with as consequence that the unlocker will have the following JSON-encoded format:

```javascript
{
    "type": 2, // indicates an AtomicSwap Input Lock
    "condition": {
        // sender's unlock hash, required, hex-encoded, fixed-size
        "sender": "0101234567890123456789012345678901012345678901234567890123456789018a50e31447b8",
        // receiver's unlock hash, required, hex-encoded, fixed-size
        "receiver": "01abc0123abc0123abc0123abc0123abc0abc0123abc0123abc0123abc0123abc057382370c8d9",
        // hashed secret, fixed size, 32 bytes, hex-encoded, sha256(secret)
        "hashedsecret": "abc543defabc543defabc543defabc543defabc543defabc543defabc543defa",
        // time lock, unix epoch timestamp (in seconds), 64-bit unsigned integer, required
        "timelock": 1522068743
    },
    "fulfillment": {
        // public key, required, format: `<algorithmSpecifier>:<key>`, where <key> is hex-encoded
        // and which byte-size is fixed but dependend upon the <algorithmSpecifier>,
        // <algorithmSpecifier> can currently only be `"ed25519"`
        "publickey": "ed25519:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
        // signature, byte-size is fixed but dependend upon the <algorithmSpecifier>,
        // when <algorithmSpecifier> equals `"ed25519"` the byte size is 64, required, hex-encoded
        "signature": "abcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefab",
        // secret, fixed size, 32 bytes, hex-encoded,
        // optional, and doesn't have to be given if this is a refund rather than a claim
        "secret": "def789def789def789def789def789dedef789def789def789def789def789de"
    }
}
```

#### JSON Encoding of Outputs in v0 Transactions

The JSON-encoded format of coin- and blockstake- outputs are always the same in v0 transactions:

```javascript
{
    // currency value (in smallest unit), big integer as string, required, positive
    "value": "10000",

    // output's unlock hash, specifier output will be locked to that hash, could represent wallet,
    // unlock type identified by the first hex-decoded byte in the unlock hash,
    // which is `0x01` in this example, telling us that this is
    // a PubKeyUnlockHash (and thus indeed a wallet address).
    //
    // See /doc/transactions/unlockhash.md to learn more about unlock hashes
    // and the JSON-encoding of the different types of unlock hashes
    "unlockhash": "01abc0123abc0123abc0123abc0123abc0abc0123abc0123abc0123abc0123abc057382370c8d9",
}
```

#### Example of a JSON-encoded v0 Transaction

The JSON encoding of a v0 Transaction can be explained best using an example:

```javascript
{
    "version": 0, // version ID, byte, always `0` for v0 transaction, optional
    "data": { // actual transaction data, required
        "coininputs": [ // coin inputs, optional
            {
                // coin output ID, crypto hash (blake2b, 256 bit), required, hex-encoded
                "parentid": "abcdef012345abcdef012345abcdef012345abcdef012345abcdef012345abcd",
                // input lock, required
                "unlocker": {
                    "type": 1,  // input lock, byte, supported range: [1,2], required
                                // `1` = SingleSignature, `2` = AtomicSwap
                    "condition": { // unlock condition, required, format dependend upon sibling "type" property
                        // public key, required, format: `<algorithmSpecifier>:<key>`, where <key> is hex-encoded
                        // and which byte-size is fixed but dependend upon the <algorithmSpecifier>,
                        // <algorithmSpecifier> can currently only be `"ed25519"`
                        "publickey": "ed25519:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"
                    },
                    "fulfillment": {    // unlock fulfillment, fulfills the unlock condition, required,
                                        // format dependend upon sibling "type" property
                        // signature, byte-size is fixed but dependend upon the <algorithmSpecifier>,
                        // when <algorithmSpecifier> equals `"ed25519"` the byte size is 64, required, hex-encoded
                        "signature": "abcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefab"
                    }
                }
            },
            {
                // coin output ID, crypto hash (blake2b, 256 bit), required, hex-encoded
                "parentid": "012345defabc012345defabc012345defabc012345defabc012345defabc0123",
                // input lock, required
                "unlocker": {
                    "type": 2,  // input lock, byte, supported range: [1,2], required
                                // `1` = SingleSignature, `2` = AtomicSwap
                    "condition": { // unlock condition, required, format dependend upon sibling "type" property
                        // sender's unlock hash, required, hex-encoded, fixed-size
                        "sender": "0101234567890123456789012345678901012345678901234567890123456789018a50e31447b8",
                        // receiver's unlock hash, required, hex-encoded, fixed-size
                        "receiver": "01abc0123abc0123abc0123abc0123abc0abc0123abc0123abc0123abc0123abc057382370c8d9",
                        // hashed secret, fixed size, 32 bytes, hex-encoded, sha256(secret)
                        "hashedsecret": "abc543defabc543defabc543defabc543defabc543defabc543defabc543defa",
                        // time lock, unix epoch timestamp (in seconds), 64-bit unsigned integer, required
                        "timelock": 1522068743
                    },
                    "fulfillment": {    // unlock fulfillment, fulfills the unlock condition, required,
                                        // format dependend upon sibling "type" property
                        // public key, required, format: `<algorithmSpecifier>:<key>`, where <key> is hex-encoded
                        // and which byte-size is fixed but dependend upon the <algorithmSpecifier>,
                        // <algorithmSpecifier> can currently only be `"ed25519"`
                        "publickey": "ed25519:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
                        // signature, byte-size is fixed but dependend upon the <algorithmSpecifier>,
                        // when <algorithmSpecifier> equals `"ed25519"` the byte size is 64, required, hex-encoded
                        "signature": "abcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefab",
                        // secret, fixed size, 32 bytes, hex-encoded
                        "secret": "def789def789def789def789def789dedef789def789def789def789def789de"
                    }
                }
            }
        ],
        "coinoutputs": [ // coin outputs, optional
            {
                "value": "3", // currency value (in smallest unit), big integer as string, required, positive
                // output's unlock hash, specifier output will be locked to that hash, could represent wallet
                "unlockhash": "0101234567890123456789012345678901012345678901234567890123456789018a50e31447b8"
            },
            {
                "value": "5", // currency value (in smallest unit), big integer as string, required, positive
                // output's unlock hash, specifier output will be locked to that hash, could represent wallet
                "unlockhash": "01abc0123abc0123abc0123abc0123abc0abc0123abc0123abc0123abc0123abc057382370c8d9"
            },
            {
                "value": "8", // currency value (in smallest unit), big integer as string, required, positive
                // output's unlock hash, specifier output will be locked to that hash, could represent wallet
                "unlockhash": "02abc0123abc0123abc0123abc0123abc0abc0123abc0123abc0123abc0123abc0ee7715ac0e0e"
            }
        ],
        "blockstakeinputs": [ // block stake inputs, optional
            {
                // blockstake output ID, crypto hash (blake2b, 256 bit), required, hex-encoded
                "parentid": "dfd23dfd23dfd23dfd23dfd23dfd23dfd23dfd23dfd23dfd23dfd23dfd23dfde",
                // input lock, required
                "unlocker": {
                    "type": 1,  // input lock, byte, supported range: [1,2], required
                                // `1` = SingleSignature, `2` = AtomicSwap
                    "condition": { // unlock condition, required, format dependend upon sibling "type" property
                        // public key, required, format: `<algorithmSpecifier>:<key>`, where <key> is hex-encoded
                        // and which byte-size is fixed but dependend upon the <algorithmSpecifier>,
                        // <algorithmSpecifier> can currently only be `"ed25519"`
                        "publickey": "ed25519:ef1234ef1234ef1234ef1234ef1234ef1234ef1234ef1234ef1234ef1234ef12"
                    },
                    "fulfillment": {    // unlock fulfillment, fulfills the unlock condition, required,
                                        // format dependend upon sibling "type" property
                        // public key, required, format: `<algorithmSpecifier>:<key>`, where <key> is hex-encoded
                        // and which byte-size is fixed but dependend upon the <algorithmSpecifier>,
                        // <algorithmSpecifier> can currently only be `"ed25519"`
                        "signature": "01234def01234def01234def01234def01234def01234def01234def01234def01234def01234def01234def01234def01234def01234def01234def01234def"
                    }
                }
            },
            {
                // blockstake output ID, crypto hash (blake2b, 256 bit), required, hex-encoded
                "parentid": "fed42fed42fed42fed42fed42fed42fed42fed42fed42fed42fed42fed42fed4",
                "unlocker": {
                    "type": 42, // input lock, byte, supported range: [1,2], required
                                // `1` = SingleSignature, `2` = AtomicSwap, other = unknown
                    // unlock condition, required,
                    // base64 encoding of the input lock condition in binary encoding format,
                    "condition": "Y29uZGl0aW9u",
                    // unlock fulfillment, required,
                    // base64 encoding of the input fulfillment condition in binary encoding format,
                    "fulfillment": "ZnVsZmlsbG1lbnQ="
                }
            }
        ],
        "blockstakeoutputs": [ // block stake outputs, optional
            {
                "value": "4", // currency value (in smallest unit), big integer as string, required, positive
                // output's unlock hash, specifier output will be locked to that hash, could represent wallet
                "unlockhash": "2a0123456789012345678901234567890101234567890123456789012345678901baa780c3d6f2"
            },
            {
                "value": "2", // currency value (in smallest unit), big integer as string, required, positive
                // output's unlock hash, specifier output will be locked to that hash, could represent wallet
                "unlockhash": "18abc0123abc0123abc0123abc0123abc0abc0123abc0123abc0123abc0123abc09c8b0a79cf2b"
            }
        ],
        // miner fees, list of currency values, at least 1 required
        "minerfees": ["1", "2", "3"],
        // arbitrary data, optional, base64 encoded
        "arbitrarydata": "ZGF0YQ==" // represents "data" in base64 encoding
    }
}
```

The details of the JSON-encoding of an unlock hash are described in
[/doc/transactions/unlockhash.md](/doc/transactions/unlockhash.md).

You can learn more about atomic swaps in
[/doc/atomicswap/atomicswap.md](/doc/atomicswap/atomicswap.md).

## Binary Encoding

### Introduction to Binary encoding

Binary encoding is used for the following purposes:

+ persistent storage of blockchain data by the daemon's modules (e.g. consensus): see the `/modules` package and all its subpackages to see this in its full detail;
+ blockchain data and any other data exchanged between peers, using the `gateway` module: see [/doc/RPC.md](/doc/RPC.md) for more information;
+ signature creation: see [the Signing Transactions chapter](#signing-transactions) for more information;

The first byte of every binary-encoded transaction always indicates the version of the transaction.
For example the `0` byte indicates the legacy v0 transaction, while the `1` byte indicates the v1 transactons
(which deprecate those v0 transactions).

Blocks have no version, instead a block always has a fixed format, no matter the transaction version(s) used within that block.

### Binary Encoding of Types

The encoding of Primitive types is explained in full in [/doc/Encoding.md](/doc/Encoding.md).
With Primitive Types we mean integers, booleans, strings, arrays, slices and structures.
In order to save you some time, it can however be summarized as:

+ Booleans are encoded as either the `1` or `0` byte;
+ All integers are encoded using [Little Endian Order][litend] (that includes lengths);
+ All signed integers are encoded as 64-bit signed integers;
+ All unsigned integers are encoded as 64-bit unsigned integers;
+ Nil pointers are encoded as the `0` byte, while non-nil pointers are prefixed with the `1` byte, and encoded using its value type;
+ Arrays are simply encoded element by element, if it's a byte array that means simply all bytes are written as is;
+ Slices (dynamic arrays) are encoded the same way as Arrays, except that they're prefixed with a 64-bit signed integer indicating the length of the slice;
+ Structures are encoded by encoding the public properties one by one, in the order as defined;

The only exception to these rules (when using the Rivine Go Library) is if a type implements [the `SiaMarshaler` interface](https://godoc.org/github.com/rivine/rivine/encoding#SiaMarshaler), it is encoded using the custom defined logic.

When decoding binary encoded data, it is assumed that you know the structure, and it will use your provided value's type in order to know how to decode it. The exception to this rule again (when using the Rivine Go Library) is that if the given value's type implements [the `SiaUnmarshaler` interface](https://godoc.org/github.com/rivine/rivine/encoding#SiaUnmarshaler), it is decoded using the custom defined logic.

When you know these rules (and exceptions) it should be trivial to deduce how any given type is encoded,
given that you know how the type is defined and which interfaces it implements.

A standard transaction (`v0` and `v1` transactions) as we know it has the following structure:

```plain
+---------+------+------+------------+------------+-----------+-----------+
| version | coin | coin | blockstake | blockstake | minerfees | arbitrary |
|         | in   | out  | inputs     | outputs    |           | data      |
+---------+------+------+------------+------------+-----------+-----------+
```

We've already seen in the introduction that (transaction) versions are encoded as a single byte.

Miner fees are encoded as a slice of currencies:

```plain
+----------+-------------+-----+-------------+
| length N | currency #0 | ... | currency #N |
+----------+-------------+-----+-------------+
| 8 bytes  | where each currency             |
|          | is 9 bytes or more              |
```

A currency (coins or block stakes) are encoded as:

```plain
+----------+-----------------------------------------------+
| length N | absolute value encoded using Big Endian order |
+----------+-----------------------------------------------+
| 8 bytes  | N bytes                                       |
```

Arbitrary data is an optional slice of bytes and encoded as follows:

```plain
+----------+---------------------+
| length N |   data byte slice   |
+----------+---------------------+
| 8 bytes  | N bytes (0 or more) |
```

Therefore it follows that if no arbitrary data is given,
it would be encoded as `0x0000000000000000`.

Coin- and blockstake- inputs have a different structure depending upon
if they are defined as part of a v0- or v1- transaction.
In both versions however it contains a ParentID.

A parentID has either the [CoinOutputID][conid] or [BlockStakeOutputID][bsoid] type,
depending on whether is a coin- or a blockstake- input.
These 2 types are just some of many identifier types in the Rivine protocol.
Other examples of identifier types are [OutputID][outid], [TransactionID][txnid] and [BlockID][blkid].
All of these identifiers types are simply alias of the `crypto.Hash` type,
which is a 32-byte cryptographic hash, generated using the [blake2b 256-bit algorithm][blake2b].

As all these identifier types are defined as a static 32-byte array,
its encoding is simply the bytes written as-is:

```plain
+-----------+
| id (hash) |
+-----------+
| 32 bytes  |
```

The binary encoding of public keys and unlock hashes and so on is explained in
[/doc/transactions/unlockhash.md](/doc/transactions/unlockhash.md).

Coin- and blockstake- outputs have a different structure depending upon
if they are defined as part of a v0- or v1- transaction.
In both versions however it contains a Value, typed as a Currency,
which we already know how to binary-encode.

[conid]: https://godoc.org/github.com/rivine/rivine/types#CoinOutputID
[bsoid]: https://godoc.org/github.com/rivine/rivine/types#BlockStakeOutputID
[outid]: https://godoc.org/github.com/rivine/rivine/types#OutputID
[txnid]: https://godoc.org/github.com/rivine/rivine/types#TransactionID
[blkid]: https://godoc.org/github.com/rivine/rivine/types#BlockID

### Binary Encoding of v1 Transactions

Before we show and explain a full example of a v1 Transaction with some possible input and output types,
let's see how inputs and outputs are binary-encoded, as the encoding of those are different in
v1 transactions compared to v0 transactions.

Understanding the following [inputs](#binary-encoding-of-inputs-in-v1-transactions) and [output](#binary-encoding-of-outputs-in-v1-transactions) chapters,
together with [the Binary Encoding of Types chapter](#the-binary-encoding-of-types),
should make you able to understand the [Example of a binary-encoded v1 transaction](#example-of-a-binary-encoded-v1-transaction),
without a sweat.

#### Binary Encoding of Inputs in v1 Transactions

Block stake- and coin- inputs are optional, and both are encoded in the same format:

```plain
+------------+--------------------+--------------------+--------------------+
| value      | unlock fulfillment | unlock fulfillment | unlock fulfillment |
| (currency) | type               | data length N      | data               |
+------------+--------------------+--------------------+--------------------+
| X bytes    | 1 byte             | 8 bytes            | N bytes            |
```

The value is of type Currency and expressed as a Big Integer. The length depends upon how big it is.
You can learn how a currency is binary-encoded in [the Binary Encoding of Types chapter](#the-binary-encoding-of-types).

The format of the unlock fulfillment data, and consequently its byte length,
depends upon the unlock fulfillment type.

##### Binary Encoding of a SingleSignatureFulfillment

The FulfillmentTypeSingleSignature (`0x01`) identifies a SingleSignatureFulfillment
and has following format:

```plain
+------------+-----------+
| PublicKey  | Signature |
+------------+-----------+
| 24+N bytes | M bytes   |
```

Please read [/doc/transactions/unlockhash.md#public-key-unlock-hash](/doc/transactions/unlockhash.md#public-key-unlock-hash),
if you want to know how a public key is encoded.

The length of a signature depends upon the signature algorithm used.
The type of (signature) algorithm is identified by the Algorithm Specifier, encoded as part of the Public Key.

This fulfillment is used to fulfill an UnlockHashCondition (ConditionType `0x01`),
where the UnlockHash given as part of that condition is of UnlockType `0x01` (implying a PublicKey UnlockHash).
It is also used to fulfill a TimeLockCondition (ConditionType `0x03`), where the internal condition,
given as part of the TimeLockCondition is the same kind of UnlockHashCondition as we just discussed.

##### Binary Encoding of an AtomicSwapFulfillment

The FulfillmentTypeAtomicSwap (`0x02`) identifies an AtomicSwapFulfillment
and can have 2 formats. It can have the new/v1 format, as well as the legacy (v0) format.

The new/v1 format can fulfill only an AtomicSwapCondition (ConditionType `0x02`), and is formatted as follows:

```plain
+------------+-----------+----------+
| PublicKey  | Signature | Secret   |
+------------+-----------+----------+
| 24+N bytes | M bytes   | 32 bytes |
```

The legacy/v0 format can fulfill an AtomicSwapCondition (ConditionType `0x02`) as well,
but is usually given in order to fulfill an UnlockHashCondition (ConditionType `0x01`),
where the UnlockHash given as part of that condition is of UnlockType Atomic Swap (`0x02`).
This legacy format is formatted as follows:

```plain
+---------------------+----------------------+---------------+-----------+------------+--------------+--------------+
| sender unlock hash  | receiver unlock hash | hashed secret | time lock | PublicKey  | Signature    | Secret       |
|                     |                      | (byte array)  | (uint64)  |            | (byte slice) | (byte array) |
+---------------------+----------------------+---------------+-----------+------------+--------------+--------------+
| 33 bytes            | 33 bytes             | 32 bytes      | 8 bytes   | 24+N bytes | M bytes      | 32 bytes     |
```

Please read [/doc/transactions/unlockhash.md#public-key-unlock-hash](/doc/transactions/unlockhash.md#public-key-unlock-hash),
if you want to know how a public key is encoded.

The length of a signature depends upon the signature algorithm used.
The type of (signature) algorithm is identified by the Algorithm Specifier, encoded as part of the Public Key.

##### Binary Encoding of a MultiSignatureFulfillment

The FulfillmentTypeMultiSignature (`0x03`) identifies a MultiSignatureFulfillment
and has following format:

```plain
+---------------------+------------------+-----+------------------+
| pubkey-signature    | pubkey-signature | ... | pubkey-signature |
| pair slice length N | pair #1          |     | pair #N          |
+---------------------+------------------+-----+------------------+
| 8 bytes             | M bytes          |     | K bytes          |
```

Where each pubkey-signature pair is binary-encoded as:

```plain
+------------+-----------+
| PublicKey  | Signature |
+------------+-----------+
| 24+N bytes | M bytes   |
```

Please read [/doc/transactions/unlockhash.md#public-key-unlock-hash](/doc/transactions/unlockhash.md#public-key-unlock-hash),
if you want to know how a public key is encoded.

The length of a signature depends upon the signature algorithm used.
The type of (signature) algorithm is identified by the Algorithm Specifier, encoded as part of the Public Key.

The MultiSignatureFulfillment is used to fulfill a MultiSignatureCondition  (ConditionType `0x04`) only.

#### Binary Encoding of Outputs in v1 Transactions

Block stake- and coin- outputs are optional, and both are encoded in the same format:

```plain
+-----------+------------------+------------------+------------------+
| outputID  | unlock condition | unlock condition | unlock condition |
|           | type             | data length N    | data             |
+-----------+------------------+------------------+------------------+
| 32 bytes  | 1 byte           | 8 bytes          | N bytes          |
```

The format of the unlock condition data, and consequently its byte length,
depends upon the unlock condition type.

##### Binary Encoding of a NilCondition

The ConditionTypeNil (`0x00`) identifies a NilCondition
and has no format, given that its data length is `0x0000000000000000` (`0`).

An output using a NilCondition can be spend anyone who owns a wallet address,
simply by signing the input with a single signature, using a private key of choice.

##### Binary Encoding of an UnlockHashCondition

The ConditionTypeUnlockHash (`0x01`) identifies an UnlockHashCondition,
has always a length of `0x2100000000000000` (`33`) and is simply an UnlockHash encoded in binary form.

The details of the binary encoding of an unlock hash are described in
[/doc/transactions/unlockhash.md](/doc/transactions/unlockhash.md).

Such condition can be fulfilled by proving the ownership of the private key,
which is linked to the public key from which this unlock hash is derived.

The signature, given as part of the fulfillment, is the proof.
See [the Signing Transactions chapter](#signing-transactions) for more information.

##### Binary Encoding of an AtomicSwapCondition

The ConditionTypeAtomicSwap (`0x02`) identifies an AtomicSwapCondition,
has always a length of `0x6a00000000000000` (`106`) and has following format:

```plain
+---------------------+----------------------+---------------+-----------+
| sender unlock hash  | receiver unlock hash | hashed secret | time lock |
|                     |                      | (byte array)  | (uint64)  |
+---------------------+----------------------+---------------+-----------+
| 33 bytes            | 33 bytes             | 32 bytes      | 8 bytes   |
```

How such condition can be fulfilled depends upon the height or timestamp of the last block in the chain.
If the contract is active, the condition can be fulfilled by proving the ownership of the private key,
which is linked to the public key from which the receiver unlock hash is derived. Otherwise the condition
can be fulfilled by proving the ownerhship of the private key, which is linked to the public key from which
the sender unlock hash is derived.

The signature, given as part of the fulfillment, is the proof.
See [the Signing Transactions chapter](#signing-transactions) for more information.

##### Binary Encoding of a TimeLockCondition

The ConditionTypeTimeLock (`0x03`) identifies a TimeLockCondition
and has following format:

```plain
+----------+------------------+------------------+------------------+
| LockTime | unlock condition | unlock condition | unlock condition |
| (uint64) | type             | data length N    | data             |
+----------+------------------+------------------+------------------+
| 33 bytes | 1 byte           | 8 bytes          | N bytes          |
```

Such condition is to be fulfilled in 2 parts:

+ The first part is done implicitly, by ensuring the last block height or time is less than the height/time specified as LockTime. If the LockTime is less then 500 milion it is to be assumed to be identifying a block height, otherwise it identifies a unix epoch timestamp in seconds;
+ If the first part has been fulfilled, the internal condition has to be fulfilled explicitly, by giving a fulfillment which is able to fulfill the internal condition, which is part of the TimeLockCondition and encoded as the very last thing;

The constant which defines whether a LockTime is a Block Height or a Unix Epoch Timestamp in seconds,
is `LockTimeMinTimestampValue` and is documented in <https://godoc.org/github.com/rivine/rivine/types#pkg-constants>.

##### Binary Encoding of a MultiSignatureCondition

The ConditionTypeMultiSignature (`0x04`) identifies a MultiSignatureCondition
and has following format:

```plain
+-------------------+------------------+----------------+-----+----------------+
| minimum signature | unlockhash slice | unlock hash #1 | ... | unlock hash #N |
| count             | length N         |                |     |                |
+-------------------+------------------+----------------+-----+----------------+
| 8 bytes           | 8bytes           | 33 bytes       |     | 33 bytes       |
```

Please read [/doc/transactions/unlockhash.md#public-key-unlock-hash](/doc/transactions/unlockhash.md#public-key-unlock-hash),
if you want to know how a public key is encoded.

Such condition can only be fulfilled by a `MultiSignatureFulfillment` (FulfillmentType `0x03`),
and the condition is fulfilled in 3 steps:

1. First it is ensured that the minimumsignaturecount property is valid,
   and that there are enough key-signature pairs present in the fulfillment
2. Then is ensured that all public keys listed in the fulfillment are authorized to do so,
   checking this is as easy as ensuring that the `PubKeyUnlockHash` of each public key, is present
   in the `unlockhashes` list of this condition
3. Finally all signatures are checked against the paired public key and the given transaction,
   within the given Fulfillment Context;

#### Example of a binary-encoded v1 transaction

Complete v1 transaction using multiple coin/blockstake inputs and outputs, as well as arbitrary data:

```plain
019e0400000000000003000000000000002200000000000000000000000000000000000000000000000000000000000022018000000000000000656432353531390000000000000000002000000000000000ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff4000000000000000ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff330000000000000000000000000000000000000000000000000000000000003302a000000000000000656432353531390000000000000000002000000000000000abababababababababababababababababababababababababababababababab4000000000000000dededededededededededededededededededededededededededededededededededededededededededededededededededededededededededededededededabadabadabadabadabadabadabadabadabadabadabadabadabadabadabadaba4400000000000000000000000000000000000000000000000000000000000044020a01000000000000011234567891234567891234567891234567891234567891234567891234567891016363636363636363636363636363636363636363636363636363636363636363bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb07edb85a00000000656432353531390000000000000000002000000000000000abababababababababababababababababababababababababababababababab4000000000000000dededededededededededededededededededededededededededededededededededededededededededededededededededededededededededededededededabadabadabadabadabadabadabadabadabadabadabadabadabadabadabadaba030000000000000001000000000000000201210000000000000001cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc01000000000000000301210000000000000002dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd010000000000000004026a00000000000000011234567891234567891234567891234567891234567891234567891234567891016363636363636363636363636363636363636363636363636363636363636363bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb07edb85a0000000001000000000000004400000000000000000000000000000000000000000000000000000000000044018000000000000000656432353531390000000000000000002000000000000000eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee4000000000000000eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee010000000000000001000000000000002a01210000000000000001abcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcd010000000000000001000000000000000102000000000000003432
```

> Note that not all possible inputs, outputs and variants are shown. See the Inputs- and Outputs chapter from above,
> to see the binary encoding of all those possible inputs and outputs.

Breakdown:

+ `01`: Version One (`0x01`)
+ `9e04000000000000`: length (`1182`) of the transaction data in binary-encoded form
+ `0300000000000000`: amount of coin inputs (`3`)
  + Coin Input #1:
    + `2200000000000000000000000000000000000000000000000000000000000022`: parentID (an ID of an unspend coin output)
    + `01`: Condition Type: UnlockHashCondition
    + `8000000000000000`: length (`128`) of the UnlockHashCondition in binary-encoded form
      + `65643235353139000000000000000000`: public key's algorithm specifier (identifying the Ed25519 signature algorithm)
      + `2000000000000000`: public key (byte) length (`32`)
      + `ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff`: public key (32 bytes)
      + `4000000000000000`: signature (byte) length (`64`)
      + `ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff`: signature (64 bytes)
  + Coin Input #2:
    + `3300000000000000000000000000000000000000000000000000000000000033`: parentID (an ID of an unspend coin output)
    + `02`: Condition Type: AtomicSwapFulfillment (could be in legacy format or the new one, but this one is using the new (v1) format)
    + `a000000000000000`: length (`160`) of the AtomicSwapFulfillment in binary-encoded form
      + `65643235353139000000000000000000`: public key's algorithm specifier (identifying the Ed25519 signature algorithm)
      + `2000000000000000`: public key (byte) length (`32`)
      + `abababababababababababababababababababababababababababababababab`: public key (32 bytes)
      + `4000000000000000`: signature (byte) length (`64`)
      + `dededededededededededededededededededededededededededededededededededededededededededededededededededededededededededededededede`: signature (64 bytes)
      + `dabadabadabadabadabadabadabadabadabadabadabadabadabadabadabadaba`: secret (32 bytes)
  + Coin Input #3:
    + `4400000000000000000000000000000000000000000000000000000000000044`: parentID (an ID of an unspend coin output)
    + `02`: Condition Type: AtomicSwapFulfillment (could be in legacy format or the new one, but this one is using the legacy (v0) format)
    + `0a01000000000000`: length (`266`) of the AtomicSwapFulfillment in legacy-binary-encoded form
      + `011234567891234567891234567891234567891234567891234567891234567891`: unlock hash of sender (`0x01` prefix indicated a PubKeyUnlockHash, implying a wallet address)
      + `016363636363636363636363636363636363636363636363636363636363636363`: unlock hash of receiver (`0x01` prefix indicated a PubKeyUnlockHash, implying a wallet address)
      + `bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb`: hashed secret (sha256 checksum of secret) (32 bytes)
      + `07edb85a00000000`: TimeLock, unix epoch seconds timestamp (`1522068743`) on which the atomic swap contract expires
      + `65643235353139000000000000000000`: public key's algorithm specifier (identifying the Ed25519 signature algorithm)
      + `2000000000000000`: public key (byte) length (`32`)
      + `abababababababababababababababababababababababababababababababab`: public key (32 bytes)
      + `4000000000000000`: signature (byte) length (`64`)
      + `dededededededededededededededededededededededededededededededededededededededededededededededededededededededededededededededededabadabadabadabadabadabadabadabadabadabadabadabadabadabadabadaba`: signature (64 bytes)
+ `0300000000000000`: amount of coin outputs (`3`) (does not have to equal the amount of coin outputs, just so you know)
  + Coin Output #1:
    + `0100000000000000`: Currency length
    + `02`: Currency Value (`2`) in smallest Coin Unit
    + `01`: UnlockConditionType, UnlockHash
    + `2100000000000000`: byte length (`33`) of UnlockHashCondition in binary-encoded form
      + `01cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc`: unlockhash of recipient (`0x01` prefix indicated a PubKeyUnlockHash, implying a wallet address)
  + Coin Output #2:
    + `0100000000000000`: Currency length
    + `03`: Currency Value (`3`) in smallest Coin Unit
    + `01`: UnlockConditionType, UnlockHash
    + `2100000000000000`: byte length (`33`) of UnlockHashCondition in binary-encoded form
      + `02dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd`: unlockhash of recipient (`0x02` prefix indicated a AtomicSwapUnlockHash)
  + Coin Output #3:
    + `0100000000000000`: Currency length
    + `04`: Currency Value (`4`) in smallest Coin Unit
    + `02`: UnlockConditionType, AtomicSwap
    + `6a00000000000000`: byte length (`106`) of AtomicSwapCondition in binary-encoded form
      + `011234567891234567891234567891234567891234567891234567891234567891`: unlock hash of sender (`0x01` prefix indicated a PubKeyUnlockHash, implying a wallet address)
      + `016363636363636363636363636363636363636363636363636363636363636363`: unlock hash of receiver (`0x01` prefix indicates a PubKeyUnlockHash, implying a wallet address)
      + `bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb`: hashed secret (sha256 checksum of secret)
      + `07edb85a00000000`: TimeLock, unix epoch seconds timestamp (`1522068743`) on which the atomic swap contract expires
+ `0100000000000000`: amount of block stake inputs (`1`)
  + Block Stake Input #1:
    + `4400000000000000000000000000000000000000000000000000000000000044`: parentID (an ID of an unspend block stake output)
    + `01`: Condition Type: UnlockHashCondition
    + `8000000000000000`: length (`128`) of the UnlockHashCondition in binary-encoded form
      + `65643235353139000000000000000000`: public key's algorithm specifier (identifying the Ed25519 signature algorithm)
      + `2000000000000000`: public key (byte) length (`32`)
      + `eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee`: public key (32 bytes)
      + `4000000000000000`: signature (byte) length (`64`)
      + `eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee`: signature (64 bytes)
+ `0100000000000000`: amount of block stake outputs (`1`)
  + Block Stake Output #1:
    + `0100000000000000`: Currency length
    + `2a`: Currency Value (`42`) expressed in units of block stakes
    + `01`: UnlockConditionType, UnlockHash
    + `2100000000000000`: byte length of UnlockHashCondition in binary-encoded form
      + `01abcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcd`: unlockhash of recipient (`0x01` prefix indicated a PubKeyUnlockHash, implying a wallet address)
+ `0100000000000000`: amount of miner fees (`1`)
  + Miner Fee #1:
    + `0100000000000000`: Currency length
    + `01`: Currency Value (`1`) in smallest Coin Unit
+ `0200000000000000`: Arbitrary data byte length (`2`)
  + `3432`: Arbitrary Data (`"42"`)

### Binary Encoding of v0 Transactions

Before we show and explain a full example of a v0 Transaction with all possible input and output types,
let's see how inputs and outputs are binary-encoded, as the encoding of those are different in
v0 transactions compared to v1 transactions.

Understanding the following [inputs](#binary-encoding-of-inputs-in-v0-transactions) and [outputs](#binary-encoding-of-outputs-in-v0-transactions) chapters,
toghether with [the Binary Encoding of Types chapter](#the-binary-encoding-of-types),
should make you able to understand the [Example of a binary-encoded v0 transaction](#example-of-a-binary-encoded-v0-transaction),
without a sweat.

#### Binary Encoding of Inputs in v0 Transactions

Block stake- and coin- inputs are optional, and both are encoded in the same format:

```plain
+-----------+------------------+
| outputID  | input lock       |
+-----------+------------------+
| 32 bytes  | X bytes          |
```

Where the input lock is encoded as:

```plain
+--------+-----------+-------------+
| type   | condition | fulfillment |
+--------+-----------+-------------+
| 1 byte | Y bytes   | Z bytes     |
```

Both condition and fulfillment are encoded as a raw binary (byte) slice,
and therefore, as you should know by now, are prefixed by the length.

While the encoding of the condtion and fulfillment slices are dependent upon the type,
indicated by the byte prefix, the unlock hash (defined in the output defined by the earlier given output ID)
can be computed without having to decode the different parts of the input lock:

```plain
hash = blake2b_checksum256(binary_encoded_condition)
unlockHash = single_byte_unlock_type + hash
```

Please read about the binary encoding of the unlockhash at
[/doc/transactions/unlockhash.md](/doc/transactions/unlockhash.md)
should you not understand the formula given above.

As the condition and fulfillment parts are encoded as byte slices,
their encoded form will be at least 8 bytes, and usually a lot more.

The only supported types are:

+ `0x01`: Single Signature
+ `0x02`: Atomic Swap Contract

As v0 transactions are legacy, no other types will ever be added.

##### Single Signature

When the type (prefix) byte of an input lock equals to `0x01`,
it should be assumed that a single-signature input lock is used for the unspend output.
This also means that the first byte of the unlockhash of that output should equal `0x01`.

The conditon, equals the receiver's public key.
See [/doc/transactions/unlockhash.md#public-key-unlock-hash](/doc/transactions/unlockhash.md#public-key-unlock-hash)
to know how a public key is encoded.

The fulfillment is encoded as:

```plain
+-----------+
| signature |
+-----------+
| X bytes   |
```

Where the exact byte size will depend on the signature algorithm used,
indicated by the public key's (algorithm) specifier.

There is currently only one known specifier, `"ed25519\0\0\0\0\0\0\0\0\0"`,
which signature size is 64 bytes, generated using a private key of 64 bytes.

See [the Signing a V0 Transactions chapter](#signing-a-v0-transaction) to know how this signature is created.

##### Atomic Swap Contract

When the type (prefix) byte of an input lock equals to `0x02`,
it should be assumed that a atomic-swap input lock is used for the unspend output.
This also means that the first byte of the unlockhash of that output should equal `0x02`.

The conditon is encoded as:

```plain
+---------------------+----------------------+---------------+-----------+
| sender unlock hash  | receiver unlock hash | hashed secret | time lock |
+---------------------+----------------------+---------------+-----------+
| 33 bytes            | 33 bytes             | 32 bytes      | 8 bytes   |
```

As usual, the details of binary encoding of an unlock hash can
be read at [/doc/transactions/unlockhash.md](/doc/transactions/unlockhash.md).

The hashed secret is 32 bytes, and is a sha256 crypto hash with as input the 32-byte secret.
The timelock is a Unix timestamp in seconds, encoded as a 64 bit unsigned integer.

The fulfillment is encoded as:

```plain
+---------------------+------------+-----------+----------+
| algorithm specifier | public key | signature | secret   |
+---------------------+------------+-----------+----------+
| 16 bytes            | 8+N bytes  | 8+M bytes | 32 bytes |
```

The encoding of a public key is explained in
[/doc/transactions/unlockhash.md#public-key-unlock-hash](/doc/transactions/unlockhash.md#public-key-unlock-hash).

The signature depends upon the algorithm specifier as well,
and is 64 bytes when using the `ed25519` signature algorithm used.
What signature algorithm is used is defined by the given Public Key.

The secret has no specific format and has a fixed byte-length of 32,
and is assumed to be random in a crypto-secure manner.

See [the Signing a V0 Transactions chapter](#signing-a-v0-transaction) to know how this signature is created.

#### Binary Encoding of Outputs in v0 Transactions

The format of coin- and blockstake- outputs are the same:

```plain
+----------+----------+-----+----------+
| length N | output#1 | ... | output#N |
+----------+----------+-----+----------+
| 8 bytes  | where each output is      |
|          | 87 bytes or more          |
```

Where each output is encoded as:

```plain
+--------------------------------------+-------------+
| currency (big integer)               | unlock hash |
+--------------------------------------+-------------+
| length N     | big integer (N bytes) | (33 bytes)  |
| (8 bytes)    |   absolute value &    |             |
|              |   big endian          |             |
```

The details of the binary encoding of an unlock hash are described in
[/doc/transactions/unlockhash.md](/doc/transactions/unlockhash.md).

#### Example of a binary-encoded v0 transaction

Complete v0 transaction using multiple coin/blockstake inputs and outputs, as well as arbitrary data:

```plain
0002000000000000002200000000000000000000000000000000000000000000000000000000000022013800000000000000656432353531390000000000000000002000000000000000ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff4000000000000000ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff3300000000000000000000000000000000000000000000000000000000000033026a00000000000000011234567891234567891234567891234567891234567891234567891234567891016363636363636363636363636363636363636363636363636363636363636363bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb07edb85a00000000a000000000000000656432353531390000000000000000002000000000000000abababababababababababababababababababababababababababababababab4000000000000000dededededededededededededededededededededededededededededededededededededededededededededededededededededededededededededededededabadabadabadabadabadabadabadabadabadabadabadabadabadabadabadaba020000000000000001000000000000000201cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc01000000000000000302dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd01000000000000004400000000000000000000000000000000000000000000000000000000000044013800000000000000656432353531390000000000000000002000000000000000eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee4000000000000000eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee010000000000000001000000000000002a01abcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcd010000000000000001000000000000000102000000000000003432
```

Breakdown:

+ `00` (version Zero, `0x00`)
+ `0200000000000000` (coin inputs, length `2` in little endian, 8 bytes):
  + first:
    + `2200000000000000000000000000000000000000000000000000000000000022` (coinOutputID, 32 bytes, crypto Hash)
    + `01` (input lock type, SingleSignature (`0x01`)):
      + `3800000000000000` (condition length: `48` in little endian, 8 bytes):
        + `65643235353139000000000000000000` (public key, algorithm specifier `ed25519`)
        + `2000000000000000` (public key, key length `32`)
          + `ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff` (public key, key itself, 32 bytes)
      + `4000000000000000` (fulfillment length: `64`, 8 bytes):
        + `ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff` (ed25519 signature, 64 bytes)
  + second:
    + `3300000000000000000000000000000000000000000000000000000000000033` (coinOutputID, 32 bytes, crypto Hash)
    + `02` (input lock type, AtomicSwapContract (`0x02`)):
      + `6a00000000000000` (condition length: `106` in little endian, 8 bytes):
        + sender:
          + `01` (unlock hash, unlock type, SingleSignature (`0x01`))
          + `1234567891234567891234567891234567891234567891234567891234567891` (unlock hash, crypto hash, 32 bytes)
        + receiver:
          + `01` (unlock hash, unlock type, SingleSignature (`0x01`))
          + `6363636363636363636363636363636363636363636363636363636363636363` (unlock hash, crypto hash, 32 bytes)
        + `bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb` (hashed secret, 32 bytes, sha256_checksum)
        + `07edb85a00000000` (timelock, unix epoch time in seconds, 1522068743)
      + `a000000000000000` (fulfillment length: `160`, 8 bytes):
        + `65643235353139000000000000000000` (public key, algorithm specifier `ed25519`)
        + `2000000000000000` (public key, key length `32`)
          + `abababababababababababababababababababababababababababababababab` (public key, key itself, 32 bytes)
        + `4000000000000000` (ed25519 signature length `64` in little endian, 8 bytes):
          + `dededededededededededededededededededededededededededededededededededededededededededededededededededededededededededededededede` (ed25519 signature, 64 bytes)
        + `dabadabadabadabadabadabadabadabadabadabadabadabadabadabadabadaba` (secret, 32 bytes, fixed size)
+ `0200000000000000` (coin outputs, length `2`):
  + first:
    + `0100000000000000` (currency, length `1`):
      + `02` (currency, value `2`, in big endian)
    + `01` (unlock hash, unlock type `0x01`, single byte):
      + `cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc` (unlock hash, crypto hash, 32 bytes)
  + second:
    + `0100000000000000` (currency, length `1`):
      + `03` (currency, value `3`, in big endian)
    + `01` (unlock hash, unlock type `0x01`, single byte):
      + `dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd` (unlock hash, crypto hash, 32 bytes)
+ `0100000000000000` (blockstake inputs, length `1`):
  + first:
    + `4400000000000000000000000000000000000000000000000000000000000044` (blockStakeOutputID, crypto Hash, 32 bytes)
    + `01` (input lock type, SingleSignature (`0x01`)):
      + `3800000000000000` (condition length: `48` in little endian, 8 bytes):
        + `65643235353139000000000000000000` (public key, algorithm specifier `ed25519`)
        + `2000000000000000` (public key, key length `32`)
          + `eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee` (public key, key itself, 32 bytes)
      + `4000000000000000` (fulfillment length: `64`):
        + `eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee` (ed25519 signature, 64 bytes)
+ `0100000000000000` (blockstake outputs, length `1`):
  + first:
    + `0100000000000000` (currency, length `1`):
      + `2a` (currency, value `42`, in big endian)
    + `01` (unlock hash, unlock type `0x01`, SingleSignature):
      + `abcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcd` (unlock hash, crypto hash, 32 bytes)
+ `0100000000000000` (miner fees, length `1`):
  + `0100000000000000` (currency, length `1`):
    + `01` (currency, value `1`, in big endian)
+ `0200000000000000` (arbitrary data, length `2`):
  + `3432` (arbitrary data `"42"`)

## Signing Transactions

### Introduction to Signing Transactions

In this chapter we'll assume that signatures are created using the Ed25519 algorithm,
as that is for now the only signature algorithm supported by the Rivine protocol.

The signature gets created using a private key, which for the Ed25519 algorithm is a fixed 64-byte key,
and a byte-slice message, which for the Rivine Protocol is, unless otherwise noted, a 32-byte crytographic hash,
which is a hash created using the [blake2b 256-bit algorithm][blake2b].

Therefore in order to understand signing, you'll need to know how the binary encoding works in the Rivine Protocol.
Reading, and understanding, [the Binary Encoding chapter](#binary-encoding) and all its subchapters is therefore really important.

Verification of a signature is done in 2 steps:

+ First the cryptographic hash is created, the same one that was used as input for the creation of that signature;
+ Using the verification Ed25519 algorithm, the signature is checked, using the newly computed hash (which is assumed to be used as input of the to be check signature), against the public key (a 32-byte sized key for the Ed25519 algorithm) that is linked to the private key that is assumed to be used;

### Signing a v1 Transaction

In order to sign a v1 transaction, you first need to compute the hash,
which is used as message, which we'll than to create a signature using the Ed25519 algorithm.

Computing that hash can be represented by following pseudo code:

```plain
blake2b_256_hash(BinaryEncoding(
  - transactionVersion: byte,
  - inputIndex: int64 (8 bytes, little endian),
  extraObjects:
    if atomicSwap:
      SiaPublicKey:
        - Algorithm: 16 bytes fixed-size array
        - Key: 8 bytes length + n bytes
    if atomicSwap as claimer (owner of receiver pub key):
      - Secret: 32 bytes fixed-size array
  - length(coinInputs): int64 (8 bytes, little endian)
  for each coinInput:
    - parentID: 32 bytes fixed-size array
  - length(coinOutputs): int64 (8 bytes, little endian)
  for each coinOutput:
    - value: Currency (8 bytes length + n bytes, little endian encoded)
    - binaryEncoding(condition)
  - length(blockStakeInputs): int64 (8 bytes, little endian)
  for each blockStakeInput:
    - parentID: 32 bytes fixed-size array
  - length(blockStakeOutputs): int64 (8 bytes, little endian)
  for each blockStakeOutput:
    - value: Currency (8 bytes length + n bytes, little endian encoded)
    - binaryEncoding(condition)
  - length(minerFees): int64 (8 bytes, little endian)
  for each minerFee:
    - fee: Currency (8 bytes length + n bytes, little endian encoded)
  - arbitraryData: 8 bytes length + n bytes
)) : 32 bytes fixed-size crypto hash
```

BinaryEncoding encodes all bytes as described above,
concatenating all output bytes together, in order as given.
The total output byte slice (result of encoding and concatenation)
is used as the input to make the crypto hashsum using [the blake2b 256bit algorithm][blake2b],
resulting in a fixed 32-byte array.

The binary encoding of an output condition depends upon the unlock condition type:

+ ConditionTypeNil: 0x000000000000000000 (type 0 as 1 byte + length 0 as 8 bytes)
+ ConditionTypeUnlockHash:
  + 0x01: 1, 1 byte (type)
  + 0x2100000000000000: 21, 8 bytes (length of unlockHash)
  + unlockHash: 33 bytes fixed-size array
+ ConditionTypeAtomicSwap:
  + 0x02: 2, 1 byte (type)
  + 0x6a00000000000000: 106, 8 bytes (length of following condition properties)
  + sender (unlockHash): 33 bytes fixed-size array
  + receiver (unlockHash): 33 bytes fixed-size array
  + hashedSecret: 32 bytes fixed-size array
  + timeLock: uint64 (8 bytes, little endian)
+ ConditionTypeTimeLock:
  + 0x03: 3, 1 byte (type)
  + length(conditionProperties): int64 (8 bytes, little endian)
  + lockTime: uint64 (8 bytes, little endian)
  + binaryEncoding(condition)

A TimeLock wraps around another condition.
For now the only valid condition type that can be used as the internal
condition of a TimeLock is a `ConditionTypeUnlockHash`.
A future version however might also allow other internal condition,
if so, this document will be updated as well to clarify that.

See [the Binary Encoding of v1 Transactions chapter](#binary-encoding-of-v1-transactions) for more information.

### Signing a v0 Transaction

In order to sign a v0 transaction, you first need to compute the hash,
which is used as message, which we'll than to create a signature using the Ed25519 algorithm.

Computing that hash can be represented by following pseudo code:

```plain
blake2b_256_hash(BinaryEncoding(
  - inputIndex: int64 (8 bytes, little endian),
  extraObjects:
    if atomicSwap:
      SiaPublicKey:
        - Algorithm: 16 bytes fixed-size array
        - Key: 8 bytes length + n bytes
    if atomicSwap as claimer (owner of receiver pub key):
      - Secret: 32 bytes fixed-size array
  for each coinInput:
    - parentID: 32 bytes fixed-size array
    - unlockHash: 33 bytes fixed-size array
  - length(coinOutputs): int64 (8 bytes, little endian)
  for each coinOutput:
    - value: Currency (8 bytes length + n bytes, little endian encoded)
    - unlockHash: 33 bytes fixed-size array
  for each blockStakeInput:
    - parentID: 32 bytes fixed-size array
    - unlockHash: 33 bytes fixed-size array
  - length(blockStakeOutputs): int64 (8 bytes, little endian)
  for each blockStakeOutput:
    - value: Currency (8 bytes length + n bytes, little endian encoded)
    - unlockHash: 33 bytes fixed-size array
  - length(minerFees): int64 (8 bytes, little endian)
  for each minerFee:
    - fee: Currency (8 bytes length + n bytes, little endian encoded)
  - arbitraryData: 8 bytes length + n bytes
)) : 32 bytes fixed-size crypto hash
```

BinaryEncoding encodes all bytes as described above,
concatenating all output bytes together, in order as given.
The total output byte slice (result of encoding and concatenation)
is used as the input to make the crypto hashsum using [the blake2b 256bit algorithm][blake2b],
resulting in a fixed 32-byte array.

See [the Binary Encoding of v0 Transactions chapter](#binary-encoding-of-v0-transactions) for more information.

[litend]: https://en.wikipedia.org/wiki/Endianness#Little-endian
[blake2b]: http://blake2.net
