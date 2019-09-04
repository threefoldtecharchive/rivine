# Auth Coin Transfer Extension

The auth coin transfer extension provides the ability to ensure only authorized addresses can receive/send coins.
Only the condition authorized to do so is able to define which addresses are authorized and which aren't.
The condition that defines who is authorized to do so can also only changed by the one(s) owning the current
condition authorized to do this and the address authorization.

> :warning: This extension also overwrites the standard Rivine `0x01` transaction controller,
> and disabled `0x00` transactions by allowing it to be decoded but not validated.

> :warning: This extensions adds a rule that is applied to all transaction versions.
> Using this extensions means that coin inputs and coin outputs can only be sent if:
> all parties are authorized, or if only one address is involved with just a single refund coin output.

An initial condition is passed to the extensions which will be used as the authorized condition until it is changed.

It is recommended to use a multisignature condition as the authorized mint condition.

```raw
                                                     +
                                                     +
                                                                          (B) Coin Transfer
                                                     +                 +----------------------+
                                                     +                 |                      |
                                                                       | +--+                 |
                                                  +-----+              | +-----------+        |
                     (C) Deauthorize              |     |   Block      | | Wallet    |        |
                   +---------------------+     +->+     |   N+2        | | for       |        |
              +----+   AddressA          +-----+  +--+--+              | | Address A |        |
+---------+   |    +---------------------+           |                 | +-----+-----+        |
|Authority+---+                                   +--+--+  +-----------+       |              |
+---+-----+                                       |     +<-+           |       |  Transfer    |
    |                                             |     |   Block      |       |  X Coins     |
    |                                             +--+--+   N+1        | +--+  v              |
    |                                                |                 | +-----+-----+        |
    |                (A) Authorize                +--+--+              | | Wallet    |        |
    |               +--------------------+        |     |   Block      | | for       |        |
    +---------------+ AddressA, AddressB +------->+     |   N          | | Address B |        |
                    +--------------------+        +--+--+              | +-----------+        |
                                                     |                 |                      |
                                                     +                 +----------------------+

                                                     +
                                                     +
```

- An address is by default not authorized;
- Only addresses that are authorized `(A)` by the authority can send or receive tokens;
  - Two addresses can transfer tokens between one another `(B)` as both the sender and receiver are authorized;
  - This implies that tokens linked to an unauthorized address are locked until it is authorized again;
- Addresses that are authorized can be deauthorized again `(C)` (locking any funds still on the address);

## Transactions

### Auth Address Update Transactions

These transactions are used to authorize and deauthorize addresses. Only addresses that are authorized at the transactions' blockheight
can receive (coin outputs targetting the authorized address) or send (coin inputs coming from the authorized address) coins.

#### JSON Encoding an Auth Address Update Transaction

```javascript
{
	// 0xB0, an example version number of an  Auth Address Update Transaction
	"version": 176, // the decimal representation of the above example version number
	// Auth Address Update Transaction Data
	"data": {
		// crypto-random 8-byte array (base64-encoded to a string) to ensure
		// the uniqueness of this transaction's ID
		"nonce": "FoAiO8vN2eU=",
		// it is not required to define auth addresses,
		// but an Auth Address Update transaction requires at least one auth address
		// or one deauth address to be defined
		"authaddresses": [
			"0112210f9efa5441ab705226b0628679ed190eb4588b662991747ea3809d93932c7b41cbe4b732",
			"01450aeb140c58012cb4afb48e068f976272fefa44ffe0991a8a4350a3687558d66c8fc753c37e",
		],
		"deauthaddresses": [
			"019e9b6f2d43a44046b62836ce8d75c935ff66cbba1e624b3e9755b98ac176a08dac5267b2c8ee",
		],
		// arbitrary data, can contain anything as long as it
		// fits within 83 bytes, and is optional
		"arbitrarydata": "dGVzdC4uLiAxLCAyLi4uIDM="
		// fulfillment which fulfills the AuthCondition,
		// can be any type of fulfillment as long as it is
		// valid AND fulfills the AuthCondition
		"authfulfillment": {
			"type": 1,
			"data": {
				"publickey": "ed25519:d285f92d6d449d9abb27f4c6cf82713cec0696d62b8c123f1627e054dc6d7780",
				"signature": "bdf023fbe7e0efec584d254b111655e1c2f81b9488943c3a712b91d9ad3a140cb0949a8868c5f72e08ccded337b79479114bdb4ed05f94dfddb359e1a6124602"
			}
		},
	}
}
```

See the [Rivine documentation about JSON-encoding of v1 Transactions][rivine-tx-v1] for more information about the primitive data types used, as well as the meaning and encoding of the different types of unlock conditions and fulfillments.

#### Binary Encoding an Auth Address Update Transaction

The binary encoding of an Auth Address UpdateTransaction uses the Rivine encoding package. In order to understand the binary encoding of such a transaction, please see [the Rivine encoding documentation][rivine-encoding] and [Rivine binary encoding of a v1 transaction](/doc/transactions/transaction.md#binary-encoding-of-v1-transactions) in order to understand how an Auth Address Update Transaction is binary encoded. That documentation also contains [an in-detail documented example of a binary encoding v1 transaction](/doc/transactions/transaction.md#example-of-a-binary-encoded-v1-transaction).

The same transaction that was shown as an example of a JSON-encoded Auth Address Update Transaction, can be represented in a hexadecimal string —when binary encoded— as:

```raw
b01680223bcbcdd9e5040112210f9efa5441ab705226b0628679ed190eb4588b662991747ea3809d93932c01450aeb140c58012cb4afb48e068f976272fefa44ffe0991a8a4350a3687558d602019e9b6f2d43a44046b62836ce8d75c935ff66cbba1e624b3e9755b98ac176a08d22746573742e2e2e20312c20322e2e2e203301c401d285f92d6d449d9abb27f4c6cf82713cec0696d62b8c123f1627e054dc6d778080bdf023fbe7e0efec584d254b111655e1c2f81b9488943c3a712b91d9ad3a140cb0949a8868c5f72e08ccded337b79479114bdb4ed05f94dfddb359e1a6124602
```

#### Signing an Auth Address Update Transaction

It is assumed that the reader of this chapter has already
read [Rivine's Introduction to Signing Transactions][rivine-signing-into] and all its referenced content.

In order to sign this type of transaction, you first need to compute the hash,
which is used as message, which we'll than to create a signature using the Ed25519 algorithm.

Computing that hash can be represented by following pseudo code:

```plain
blake2b_256_hash(SiaBinaryEncoding(
  - transactionVersion: 1 byte
  - specifier: 16 bytes, hardcoded to "auth addr update"
  - nonce: 8 bytes
  - slice of auth addresses
  - slice of deauth addresses
  - arbitraryData: 8 bytes length + n bytes
)) : 32 bytes fixed-size crypto hash
```

### Auth Condition Update Transactions

These transactions are used to change the condition authorized to use these transactions

An Auth condition can be any of the following conditions:

* (1) An [UnlockHash Condition][rivine-condition-uh]: it is a single person (or multiple people owning the same private key for the same wallet);
* (2) A [MultiSignature Condition][rivine-condition-multisig]: it is a multi signature wallet, most likely meaning multiple people owning different private keys that all have a certain degree of control in the same multi signature wallet, allowing them to come to a consensus on the creation of coins and redefinition of who the coin creators are (to be);
* (3) An [TimeLocked Condition][rivine-condition-tl]: the minting powers (creation of coins and redefinition of the coin creators) are locked until a certain (unix epoch) timestamp or block height. Once this height or time is reached, the internal condition (an [UnlockHash Condition][rivine-condition-uh] (1) or a [MultiSignature Condition][rivine-condition-multisig]) (2) is the condition that defines who can create coins and redefine who the coin creators are to be. Prior to this timestamp or block height no-one can create coins or redefine who the coin creators are to be, not even the ones defined by the internal condition of the currently active [TimeLocked Condition][rivine-condition-tl].
as well as to authorize and deauthorize addresses that can (no longer) send/receive tokens.

The Genesis Auth Condition is hardcoded.

In practice a MultiSignature Condition will always be used as AuthCondition,
this is however not a consensus-defined requirement, as discussed earlier.

#### JSON Encoding an Auth Condition Update Transaction

```javascript
{
	// 0xB1, an example version number of an Auth Condition Update Transaction
	"version": 177, // the decimal representation of the above example version number
	// Auth Condition Update Transaction Data
	"data": {
		// crypto-random 8-byte array (base64-encoded to a string) to ensure
		// the uniqueness of this transaction's ID
		"nonce": "1oQFzIwsLs8=",
		// arbitrary data, can contain anything as long as it
		// fits within 83 bytes, and is optional.
		"arbitrarydata": "dGVzdC4uLiAxLCAyLi4uIDM="
		// condition which will become the new AuthCondition
		// once the transaction is part of a created block and
		// until there is a newer block with another accepted AuthCondition
		"authcondition": {
			"type": 1,
			"data": {
				"unlockhash": "01e78fd5af261e49643dba489b29566db53fa6e195fa0e6aad4430d4f06ce88b73e047fe6a0703"
			}
		},
		// fulfillment which fulfills the current active AuthCondition,
		// can be any type of fulfillment as long as it is
		// valid AND fulfills the current active AuthCondition
		"authfulfillment": {
			"type": 1,
			"data": {
				"publickey": "ed25519:d285f92d6d449d9abb27f4c6cf82713cec0696d62b8c123f1627e054dc6d7780",
				"signature": "ad59389329ed01c5ee14ce25ae38634c2b3ef694a2bdfa714f73b175f979ba6613025f9123d68c0f11e8f0a7114833c0aab4c8596d4c31671ec8a73923f02305"
			}
		},
	}
}
```

See the [Rivine documentation about JSON-encoding of v1 Transactions][rivine-tx-v1] for more information about the primitive data types used, as well as the meaning and encoding of the different types of unlock conditions and fulfillments.

#### Binary Encoding an Auth Condition Update Transaction

The binary encoding of an Auth Condition Update Transaction uses the Rivine encoding package. In order to understand the binary encoding of such a transaction, please see [the Rivine encoding documentation][rivine-encoding] and [Rivine binary encoding of a v1 transaction](/doc/transactions/transaction.md#binary-encoding-of-v1-transactions) in order to understand how an Auth Condition Update Transaction binary encoded. That documentation also contains [an in-detail documented example of a binary encoding v1 transaction](/doc/transactions/transaction.md#example-of-a-binary-encoded-v1-transaction).

The same transaction that was shown as an example of a JSON-encoded Auth Condition Update Transaction, can be represented in a hexadecimal string —when binary encoded— as:

```raw
b1d68405cc8c2c2ecf22746573742e2e2e20312c20322e2e2e2033014201e78fd5af261e49643dba489b29566db53fa6e195fa0e6aad4430d4f06ce88b7301c401d285f92d6d449d9abb27f4c6cf82713cec0696d62b8c123f1627e054dc6d778080ad59389329ed01c5ee14ce25ae38634c2b3ef694a2bdfa714f73b175f979ba6613025f9123d68c0f11e8f0a7114833c0aab4c8596d4c31671ec8a73923f02305
```

#### Signing an Auth Condition Update Transaction

It is assumed that the reader of this chapter has already
read [Rivine's Introduction to Signing Transactions][rivine-signing-into] and all its referenced content.

In order to sign an auth condition update transaction, you first need to compute the hash,
which is used as message, which we'll than to create a signature using the Ed25519 algorithm.

Computing that hash can be represented by following pseudo code:

```plain
blake2b_256_hash(SiaBinaryEncoding(
  - transactionVersion: 1 byte
  - specifier: 16 bytes, hardcoded to "auth cond update"
  - nonce: 8 bytes
  - auth condition
  - arbitraryData: 8 bytes length + n bytes
)) : 32 bytes fixed-size crypto hash
```
