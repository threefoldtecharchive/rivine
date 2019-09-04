# Rivine minting extension

The minting extensions provides the ability to mint tokens without any prior backing (see: create coins).
Only the condition authorized to do so is able to mint new tokens or change the condition that's authorized to do this from then on.

An initial condition is passed to the extensions which will be used as the authorized condition until it is changed.

It is recommended to use a multisignature condition as the authorized mint condition.

## Usage

### Daemon

After creating the consensus module and registering the HTTP handlers you can use following snippet:

```golang
// This is an unlockcondition for testing purposes.
uhString := "01fbaa166912244082784a28ec8756d5d2126f789ed25d93074b2afa05f984e32c0ae996e398e5"
var uh types.UnlockHash
if err := uh.LoadString(uhString); err != nil {
    panic(err)
}
condition := types.NewUnlockHashCondition(uh)

/* any condition that is defined types.unlockconditions.go can be passed to NewMintingPlugin
    for example this can be one of following:
    * singlesignature condition, which is an unlockhash
    * timelock condition
    * multisignature condition
*/

// define the transaction versions for the 2 extra transactions possible
const (
	// can be any unique transaction version >= 128
	minterDefinitionTxVersion = iota + 128
	coinCreationTxVersion
	coinDestructionTxVersion
)

// Pass the condition the NewMintingPlugin
plugin := minting.NewMintingPlugin(types.NewCondition(condition), minterDefinitionTxVersion, coinCreationTxVersion, &minting.MintingPluginOptions{
	CoinDestructionTransactionVersion: coinDestructionTxVersion, // if equals 0 -> disabled
	UseLegacySiaEncoding: false, // false by default, usng rivine encoding, otherwise it uses sia encoding if true
	RequireMinerFees: false, // false by default, otherwise it adds and requires miner fees if true
})

err = cs.RegisterPlugin(context.BackGround(),"minting", plugin)
if err != nil {
	plugin.Close() // Close it since it can still hold resources
    return err
}
```

This will register the minting plugin on the consensus.

### Client

After creating the command line client you can use following snippet

```golang
import (
    mintingcli "github.com/threefoldtech/rivine/extensions/minting/client"
)

// Will create the explore mintcondition command
// * rivinec explore mintcondition [height] [flags]
mintingcli.CreateExploreCmd(cliClient)

// define the transaction versions for the 2 extra transactions possible
const (
	// can be any unique transaction version >= 128
	minterDefinitionTxVersion = iota + 128
	coinCreationTxVersion
	coinDestructionTxVersion
)

// Will create the createCoinTransaction and createMinterDefinitionTransaction command
// * rivinec wallet create minterdefinitiontransaction
// * rivinec wallet create coincreationtransaction
mintingcli.CreateWalletCmds(cliClient, minterDefinitionTxVersion, coinCreationTxVersion, &mintingcli.WalletCmdsOpts{
	CoinDestructionTxVersion: coinDestructionTxVersion, // tx version of the coin destruction tx
	RequireMinerFees: false, // false by default, otherwise it adds and requires miner fees if true
})

mintingReader := mintingcli.NewPluginExplorerClient(cliClient)

// Register the transaction types
types.RegisterTransactionVersion(minterDefinitionTxVersion, minting.MinterDefinitionTransactionController{
	MintConditionGetter: mintingReader,
	TransactionVersion: minterDefinitionTxVersion,
})
types.RegisterTransactionVersion(coinCreationTxVersion, minting.CoinCreationTransactionController{
	MintConditionGetter: mintingReader,
	TransactionVersion: coinCreationTxVersion,
})
types.RegisterTransactionVersion(coinDestructionTxVersion, minting.CoinDestructionTransactionController{
	MintConditionGetter: mintingReader,
	TransactionVersion: coinDestructionTxVersion,
})
```

## Transactions

### Minter Definition Transactions

Minter Definition Transactions are used to redefine the creators of coins (AKA minters). These transactions can only be created by the Coin Creators. The (previously-defined) mint condition —meaning the mint condition active at the height of the (to be) created Minter Definition Transaction— defines who the coin creators are and thus who can redefine who the coin creators are to become. A mint condition can be any of the following conditions:

* (1) An [UnlockHash Condition][rivine-condition-uh]: it is a single person (or multiple people owning the same private key for the same wallet);
* (2) A [MultiSignature Condition][rivine-condition-multisig]: it is a multi signature wallet, most likely meaning multiple people owning different private keys that all have a certain degree of control in the same multi signature wallet, allowing them to come to a consensus on the creation of coins and redefinition of who the coin creators are (to be);
* (3) An [TimeLocked Condition][rivine-condition-tl]: the minting powers (creation of coins and redefinition of the coin creators) are locked until a certain (unix epoch) timestamp or block height. Once this height or time is reached, the internal condition (an [UnlockHash Condition][rivine-condition-uh] (1) or a [MultiSignature Condition][rivine-condition-multisig]) (2) is the condition that defines who can create coins and redefine who the coin creators are to be. Prior to this timestamp or block height no-one can create coins or redefine who the coin creators are to be, not even the ones defined by the internal condition of the currently active [TimeLocked Condition][rivine-condition-tl].

The Coin Creation transactions defines 4 fields:

* `mintfulfillment`: the fulfillment which has to fulfill the consensus-defined MintCondition, just the same as that a Coin Input's fulfillment has to fulfill the condition of the Coin Output it is about to spend;
* `mintcondition`: the condition which will become the new mint condition (that has to be fulfilled in order to create coins and redefine the mint condition, in other words the condition that defines who the coin creators are) once the transaction is part of a created block and until there is a newer block with an accepted mint condition;
* `minerfees`: defines the transaction fee(s) (works the same as in regular transactions);
* `arbitrarydata`: describes the reason for creating these coins;
* `nonce`: a crypto-random 8-byte array, used to ensure the uniqueness of this transaction's ID;

The Genesis Mint Condition is hardcoded.

In practice a MultiSignature Condition will always be used as MintCondition,
this is however not a consensus-defined requirement, as discussed earlier.

#### JSON Encoding a Minter Definition Transaction

```javascript
{
	// 0x80, an example version number of a Minter Definition Transaction
	"version": 128, // the decimal representation of the above example version number
	// Coin Creation Transaction Data
	"data": {
		// crypto-random 8-byte array (base64-encoded to a string) to ensure
		// the uniqueness of this transaction's ID
		"nonce": "FoAiO8vN2eU=",
		// fulfillment which fulfills the MintCondition,
		// can be any type of fulfillment as long as it is
		// valid AND fulfills the MintCondition
		"mintfulfillment": {
			"type": 1,
			"data": {
				"publickey": "ed25519:d285f92d6d449d9abb27f4c6cf82713cec0696d62b8c123f1627e054dc6d7780",
				"signature": "bdf023fbe7e0efec584d254b111655e1c2f81b9488943c3a712b91d9ad3a140cb0949a8868c5f72e08ccded337b79479114bdb4ed05f94dfddb359e1a6124602"
			}
		},
		// condition which will become the new MintCondition
		// once the transaction is part of a created block and
		// until there is a newer block with another accepted MintCondition
		"mintcondition": {
			"type": 1,
			"data": {
				"unlockhash": "01e78fd5af261e49643dba489b29566db53fa6e195fa0e6aad4430d4f06ce88b73e047fe6a0703"
			}
		},
		// the transaction fees to be paid, also paid in
		// newly created) coins, rather than inputs
		"minerfees": ["1000000000"],
		// arbitrary data, can contain anything as long as it
		// fits within 83 bytes, but is in practice used
		// to link the capacity added/created
		// with as a consequence the creation of
		// these transaction and its coin outputs
		"arbitrarydata": "dGVzdC4uLiAxLCAyLi4uIDM="
	}
}
```

See the [Rivine documentation about JSON-encoding of v1 Transactions][rivine-tx-v1] for more information about the primitive data types used, as well as the meaning and encoding of the different types of unlock conditions and fulfillments.

#### Binary Encoding a Minter Definition Transaction

The binary encoding of a Minter Definition Transaction uses the Rivine encoding package. In order to understand the binary encoding of such a transaction, please see [the Rivine encoding documentation][rivine-encoding] and [Rivine binary encoding of a v1 transaction](https://github.com/threefoldtech/rivine/blob/master/doc/transactions/transaction.md#binary-encoding-of-v1-transactions) in order to understand how a Minter Definition Transaction is binary encoded. That documentation also contains [an in-detail documented example of a binary encoding v1 transaction](https://github.com/threefoldtech/rivine/blob/master/doc/transactions/transaction.md#example-of-a-binary-encoded-v1-transaction).

The same transaction that was shown as an example of a JSON-encoded Minter Definition Transaction, can be represented in a hexadecimal string —when binary encoded— as:

```raw
801680223bcbcdd9e5018000000000000000656432353531390000000000000000002000000000000000d285f92d6d449d9abb27f4c6cf82713cec0696d62b8c123f1627e054dc6d77804000000000000000bdf023fbe7e0efec584d254b111655e1c2f81b9488943c3a712b91d9ad3a140cb0949a8868c5f72e08ccded337b79479114bdb4ed05f94dfddb359e1a612460201210000000000000001e78fd5af261e49643dba489b29566db53fa6e195fa0e6aad4430d4f06ce88b73010000000000000004000000000000003b9aca00180000000000000061206d696e74657220646566696e6974696f6e2074657374
```

#### Signing a Minter Definition Transaction

It is assumed that the reader of this chapter has already
read [Rivine's Introduction to Signing Transactions][rivine-signing-into] and all its referenced content.

In order to sign a v1 transaction, you first need to compute the hash,
which is used as message, which we'll than to create a signature using the Ed25519 algorithm.

Computing that hash can be represented by following pseudo code:

```plain
blake2b_256_hash(RivineBinaryEncoding(
  - transactionVersion: 1 byte
  - specifier: 16 bytes, hardcoded to "minter defin tx\0"
  - nonce: 8 bytes
  - binaryEncoding(mintCondition)
  - length(minerFees): int64 (8 bytes, little endian)
  for each minerFee:
    - fee: Currency (8 bytes length + n bytes, little endian encoded)
  - arbitraryData: 8 bytes length + n bytes
)) : 32 bytes fixed-size crypto hash
```

### Coin Creation Transactions

Coin Creation Transactions are used for the creation of new coins. These transactions can only be created by the Coin Creators (also called minters). The Mint Condition defines who the coin creators are. If it is an [UnlockHash Condition][rivine-condition-uh] it is a single person, while it will be a [MultiSignature Condition][rivine-condition-multisig] in case there are multiple coin creators that have to come to a consensus.

The Coin Creation transactions defines 4 fields:

* `mintfulfillment`: the fulfillment which has to fulfill the consensus-defined MintCondition, just the same as that a Coin Input's fulfillment has to fulfill the condition of the Coin Output it is about to spend;
* `coinoutputs`: defines coin outputs, the destination of coins (works the same as in regular transactions);
* `minerfees`: defines the transaction fee(s) (works the same as in regular transactions);
* `arbitrarydata`: describes the capacity that is created/added, creating these coins as a result;
* `nonce`: a crypto-random 8-byte array, used to ensure the uniqueness of this transaction's ID;

The Genesis Mint Condition is hardcoded.

In practice a MultiSignature Condition will always be used as MintCondition,
this is however not a consensus-defined requirement. You can ready more about this in the chapter on
[Minter Definition Transactions](#minter-definition-transactions).

#### JSON Encoding a Coin Creation Transaction

```javascript
{
	// 0x81, an example version number of a Coin Creation Transaction
	"version": 129, // the decimal representation of the above example version number
	// Coin Creation Transaction Data
	"data": {
		// crypto-random 8-byte array (base64-encoded to a string) to ensure
		// the uniqueness of this transaction's ID
		"nonce": "1oQFzIwsLs8=",
		// fulfillment which fulfills the MintCondition,
		// can be any type of fulfillment as long as it is
		// valid AND fulfills the MintCondition
		"mintfulfillment": {
			"type": 1,
			"data": {
				"publickey": "ed25519:d285f92d6d449d9abb27f4c6cf82713cec0696d62b8c123f1627e054dc6d7780",
				"signature": "ad59389329ed01c5ee14ce25ae38634c2b3ef694a2bdfa714f73b175f979ba6613025f9123d68c0f11e8f0a7114833c0aab4c8596d4c31671ec8a73923f02305"
			}
		},
		// defines the recipients (as conditions) who are to
		// receive the paired (newly created) coin values
		"coinoutputs": [{
			"value": "500000000000000",
			"condition": {
				"type": 1,
				"data": {
					"unlockhash": "01e3cbc41bd3cdfec9e01a6be46a35099ba0e1e1b793904fce6aa5a444496c6d815f5e3e981ccf"
				}
			}
		}],
		// the transaction fees to be paid, also paid in
		// newly created) coins, rather than inputs
		"minerfees": ["1000000000"],
		// arbitrary data, can contain anything as long as it
		// fits within 83 bytes, but is in practice used
		// to link the capacity added/created
		// with as a consequence the creation of
		// these transaction and its coin outputs
		"arbitrarydata": "dGVzdC4uLiAxLCAyLi4uIDM="
	}
}
```

See the [Rivine documentation about JSON-encoding of v1 Transactions][rivine-tx-v1] for more information about the primitive data types used, as well as the meaning and encoding of the different types of unlock conditions and fulfillments.

#### Binary Encoding a Coin Creation Transaction

The binary encoding of a Coin Creation Transaction uses the Rivine encoding package. In order to understand the binary encoding of such a transaction, please see [the Rivine encoding documentation][rivine-encoding] and [Rivine binary encoding of a v1 transaction](https://github.com/threefoldtech/rivine/blob/master/doc/transactions/transaction.md#binary-encoding-of-v1-transactions) in order to understand how a Coin Creation Transaction is binary encoded. That documentation also contains [an in-detail documented example of a binary encoding v1 transaction](https://github.com/threefoldtech/rivine/blob/master/doc/transactions/transaction.md#example-of-a-binary-encoded-v1-transaction).

The same transaction that was shown as an example of a JSON-encoded Coin Creation Transaction, can be represented in a hexadecimal string —when binary encoded— as:

```raw
8133a6432220334946018000000000000000656432353531390000000000000000002000000000000000d285f92d6d449d9abb27f4c6cf82713cec0696d62b8c123f1627e054dc6d77804000000000000000a074b976556d6ea2e4ae8d51fbbb5ec99099f11918201abfa31cf80d415c8d5bdfda5a32d9cc167067b6b798e80c6c1a45f6fd9e0f01ac09053e767b15d310050100000000000000070000000000000001c6bf5263400001210000000000000001e78fd5af261e49643dba489b29566db53fa6e195fa0e6aad4430d4f06ce88b73010000000000000004000000000000003b9aca0012000000000000006d6f6e65792066726f6d2074686520736b79
```

#### Signing a Coin Creation Transaction

It is assumed that the reader of this chapter has already
read [Rivine's Introduction to Signing Transactions][rivine-signing-into] and all its referenced content.

In order to sign a v1 transaction, you first need to compute the hash,
which is used as message, which we'll than to create a signature using the Ed25519 algorithm.

Computing that hash can be represented by following pseudo code:

```plain
blake2b_256_hash(RivineBinaryEncoding(
  - transactionVersion: 1 byte
  - specifier: 16 bytes, hardcoded to "coin mint tx\0\0\0\0"
  - nonce: 8 bytes
  - length(coinOutputs): int64 (8 bytes, little endian)
  for each coinOutput:
    - value: Currency (8 bytes length + n bytes, little endian encoded)
    - binaryEncoding(condition)
  - length(minerFees): int64 (8 bytes, little endian)
  for each minerFee:
    - fee: Currency (8 bytes length + n bytes, little endian encoded)
  - arbitraryData: 8 bytes length + n bytes
)) : 32 bytes fixed-size crypto hash
```

### Coin Destruction Transactions

Coin Destruction Transactions are used for the destruction of existing coins.

The Coin Destruction transactions defines 4 fields:

* `coininputs`: the fulfillment which has to fulfill the consensus-defined MintCondition, just the same as that a Coin Input's fulfillment has to fulfill the condition of the Coin Output it is about to spend;
* `refundcoinoutput`: defines coin outputs, the destination of coins (works the same as in regular transactions);
* `minerfees`: defines the transaction fee(s) (works the same as in regular transactions);
* `arbitrarydata`: describes the capacity that is created/added, creating these coins as a result;

#### JSON Encoding a Coin Destruction Transaction

```javascript
{
	// 0x82, an example version number of a Coin Destruction Transaction
	"version": 130, // the decimal representation of the above example version number
	// Coin Creation Transaction Data
	"data": {
		// defines the coin outputs to be burned
		"coininputs": [
			{
				"parentid": "1100000000000000000000000000000000000000000000000000000000000011",
				"fulfillment": {
					"type": 1,
					"data": {
						"publickey": "ed25519:def123def123def123def123def123def123def123def123def123def123def1",
						"signature": "ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef"
					}
				}
			}
		],
		// defines the optonal refund coin output
		"refundcoinoutput": {
			"value": "500000000000000",
			"condition": {
				"type": 1,
				"data": {
					"unlockhash": "01e3cbc41bd3cdfec9e01a6be46a35099ba0e1e1b793904fce6aa5a444496c6d815f5e3e981ccf"
				}
			}
		},
		// the transaction fees to be paid, small partion of coins that are not burned
		"minerfees": ["1000000000"],
		// arbitrary data, can contain anything as long as it
		// fits within 83 bytes
		"arbitrarydata": "dGVzdC4uLiAxLCAyLi4uIDM="
	}
}
```

See the [Rivine documentation about JSON-encoding of v1 Transactions][rivine-tx-v1] for more information about the primitive data types used, as well as the meaning and encoding of the different types of unlock conditions and fulfillments.

#### Binary Encoding a Coin Destruction Transaction

The binary encoding of a Coin Destruction Transaction uses the Rivine encoding package. In order to understand the binary encoding of such a transaction, please see [the Rivine encoding documentation][rivine-encoding] and [Rivine binary encoding of a v1 transaction](https://github.com/threefoldtech/rivine/blob/master/doc/transactions/transaction.md#binary-encoding-of-v1-transactions) in order to understand how a Coin Destruction Transaction is binary encoded. That documentation also contains [an in-detail documented example of a binary encoding v1 transaction](https://github.com/threefoldtech/rivine/blob/master/doc/transactions/transaction.md#example-of-a-binary-encoded-v1-transaction).

The same transaction that was shown as an example of a JSON-encoded Coin Destruction Transaction, can be represented in a hexadecimal string —when binary encoded— as:

```raw
8133a6432220334946018000000000000000656432353531390000000000000000002000000000000000d285f92d6d449d9abb27f4c6cf82713cec0696d62b8c123f1627e054dc6d77804000000000000000a074b976556d6ea2e4ae8d51fbbb5ec99099f11918201abfa31cf80d415c8d5bdfda5a32d9cc167067b6b798e80c6c1a45f6fd9e0f01ac09053e767b15d310050100000000000000070000000000000001c6bf5263400001210000000000000001e78fd5af261e49643dba489b29566db53fa6e195fa0e6aad4430d4f06ce88b73010000000000000004000000000000003b9aca0012000000000000006d6f6e65792066726f6d2074686520736b79
```

#### Signing a Coin Destruction Transaction

It is assumed that the reader of this chapter has already
read [Rivine's Introduction to Signing Transactions][rivine-signing-into] and all its referenced content.

In order to sign a v1 transaction, you first need to compute the hash,
which is used as message, which we'll than to create a signature using the Ed25519 algorithm.

Computing that hash can be represented by following pseudo code:

```plain
blake2b_256_hash(RivineBinaryEncoding(
  - transactionVersion: 1 byte
  - specifier: 16 bytes, hardcoded to "coin destruct tx\0"
  - nonce: 8 bytes
  - all coin inputs
  - refund pointer defined or not: 1 byte
    if refund output defined:
	- value: Currency (8 bytes length + n bytes, little endian encoded)
	- binaryEncoding(condition)
  - length(minerFees): int64 (8 bytes, little endian)
  for each minerFee:
    - fee: Currency (8 bytes length + n bytes, little endian encoded)
  - arbitraryData: 8 bytes length + n bytes
)) : 32 bytes fixed-size crypto hash
```
