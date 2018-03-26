# Transaction

The main purpose of a transaction is to transfer coins and/or block stakes between addresses.
For each coin that is spend (registered as coin output), there must be one or multiple
(coin) inputs backing it up. These inputs must be previously-registered outputs,
which haven't been used as input yet. The same goes for block stakes, another kind of asset.

In order to create/send a transaction, one has to pay a transaction fee. These are registered
as part of a transaction and labeled as "Miner Fees". It is important to note that the total sum
of coin inputs MUST be equal to the total sum of coin outs and miner fees combined.
Block stakes are a seperate asset, where the total sum of block stake inputs must
simply be equal to the total sum of block stake outputs.

As each output is backed by one or multiple inputs, it is not uncommon to have a too big
amount of input registered. If so, it is the convention to simply register the extra
amount as another output, but this time addressed to your own wallet.
This works very similar to change money you get in a supermarket by paying too much.

Besides the inputs, outputs and minerfees, one can also optionally attach some arbtirary binary
data to a transaction, which isn't of any meaning to the blockchain itself.

## Relevant Source Files

For those interested, this document explains logic
implemented in the Golang reference Rivine implementation, and covers following source files:

+ [/types/transaction.go](/types/transaction.go)
+ [/types/inputlock.go](/types/inputlock.go)
+ [/types/unlockhash.go](/types/unlockhash.go)
+ [/types/signatures.go](/types/signatures.go)
+ [/types/currency.go](/types/currency.go)
+ [/types/timestamp.go](/types/timestamp.go)

The master version of the (public) Golang documentation for
the module of these files can be found at: https://godoc.org/github.com/rivine/rivine/types

## Send coins

Send coins from 0x78e6... to 0x1d64..

// TODO: Diagram

When spending money you need to add a coininput in a transaction, info needed in coininput is:
1. ParentID is the ID of the unspend coinoutput (UTXO)
2. Recalculate the unlockconditions that generates the unlockhash (normally they are standard and can be reconstructed)
3. Use the corresponding private key to sign a transactionSignature.
4. The money that can be spent in the transaction is found in the corresponding coinoutput. ex. Value = 2000

Per transaction: The sum of all values in coinoutput should be less than the sum of all "unlocked" values in the coininput. The difference is minerfee for the the block generator.

## Arbitrary data

Arbitrary Data can be of any size. There is however a size limit on a Transaction and a block.
Keep in mind that the fee is depending on the size of a transaction, a blockcreator can ignore to add a transaction with small fees for a lot of small transactions with a summed up bigger fee (opportune).

Arbitrary data can be used to make verifiable announcements, or to have other
protocols sit on top of Rivine. The arbitrary data can also be used for soft
forks, and for protocol relevant information. Any arbitrary data is allowed by
the consensus.

## Double Spend Rules

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

## Format

As always, there are 2 formats. The binary format and the text/json format.
The binary format is used as part of the rivine protocol, when encoding transactions between gateways,
while the text (JSON) encoding is used for storage as well as part of the daemon's REST API (e.g. `/transactionpool/transactions`).

### Text Format

The text format is done using JSON encoding,
where binary values (byte slices) are usually encoded in a hex-format,
except for arbitrary data, which is base64-encoded.

The text (JSON) format can be explained best using an example:

```javascript
{
	"version": 0, // version ID, byte, for now always `0`, optional
	"data": { // actual transaction, required
		"coininputs": [ // coin inputs, at least 1 element required
			{
				// coin output ID, crypto hash (blake2b, 256 bit), required, hex-encoded
				"parentid": "abcdef012345abcdef012345abcdef012345abcdef012345abcdef012345abcd",
				// input lock, required
				"unlocker": {
					"type": 1, 	// input lock, byte, supported range: [1,2], unknown range: [3,255], required
											// `1` = SingleSignature, `2` = AtomicSwap
					"condition": { // unlock condition, required, format dependend upon sibling "type" property
						// public key, required, format: `<algorithmSpecifier>:<key>`, where <key> is hex-encoded
						// and which byte-size is fixed but dependend upon the <algorithmSpecifier>,
						// <algorithmSpecifier> can currently only be `"ed25519"`
						"publickey": "ed25519:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"
					},
					"fulfillment": { 	// unlock fulfillment, fulfills the unlock condition, required,
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
					"type": 2, 	// input lock, byte, supported range: [1,2], unknown range: [3,255], required
											// `1` = SingleSignature, `2` = AtomicSwap
					"condition": { // unlock condition, required, format dependend upon sibling "type" property
						// sender's unlock hash, required, hex-encoded, fixed-size
						"sender": "010123456789012345678901234567890101234567890123456789012345678901dec8f8544d34",
						// receiver's unlock hash, required, hex-encoded, fixed-size
						"receiver": "01abc0123abc0123abc0123abc0123abc0abc0123abc0123abc0123abc0123abc0efb39211ea2a",
						// hashed secret, fixed size, 32 bytes, hex-encoded, sha256(secret)
						"hashedsecret": "abc543defabc543defabc543defabc543defabc543defabc543defabc543defa",
						// time lock, unix epoch timestamp (in seconds), 64-bit unsigned integer, required
						"timelock": 1522068743
					},
					"fulfillment": { 	// unlock fulfillment, fulfills the unlock condition, required,
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
				// output's unlock hash, specifier output will belocked to that hash, could represent wallet
				"unlockhash": "010123456789012345678901234567890101234567890123456789012345678901dec8f8544d34"
			},
			{
				"value": "5", // currency value (in smallest unit), big integer as string, required, positive
				// output's unlock hash, specifier output will belocked to that hash, could represent wallet
				"unlockhash": "01abc0123abc0123abc0123abc0123abc0abc0123abc0123abc0123abc0123abc0efb39211ea2a"
			},
			{
				"value": "8", // currency value (in smallest unit), big integer as string, required, positive
				// output's unlock hash, specifier output will belocked to that hash, could represent wallet
				"unlockhash": "02abc0123abc0123abc0123abc0123abc0abc0123abc0123abc0123abc0123abc0efb39211ea2a"
			}
		],
		"blockstakeinputs": [ // block stake inputs, optional
			{
				// blockstake output ID, crypto hash (blake2b, 256 bit), required, hex-encoded
				"parentid": "dfd23dfd23dfd23dfd23dfd23dfd23dfd23dfd23dfd23dfd23dfd23dfd23dfde",
				// input lock, required
				"unlocker": {
					"type": 1, 	// input lock, byte, supported range: [1,2], unknown range: [3,255], required
											// `1` = SingleSignature, `2` = AtomicSwap
					"condition": { // unlock condition, required, format dependend upon sibling "type" property
						// public key, required, format: `<algorithmSpecifier>:<key>`, where <key> is hex-encoded
						// and which byte-size is fixed but dependend upon the <algorithmSpecifier>,
						// <algorithmSpecifier> can currently only be `"ed25519"`
						"publickey": "ed25519:ef1234ef1234ef1234ef1234ef1234ef1234ef1234ef1234ef1234ef1234ef12"
					},
					"fulfillment": { 	// unlock fulfillment, fulfills the unlock condition, required,
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
					"type": 42, // input lock, byte, supported range: [1,2], unknown range: [3,255], required
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
				// output's unlock hash, specifier output will belocked to that hash, could represent wallet
				"unlockhash": "2a0123456789012345678901234567890101234567890123456789012345678901dec8f8544d34"
			},
			{
				"value": "2", // currency value (in smallest unit), big integer as string, required, positive
				// output's unlock hash, specifier output will belocked to that hash, could represent wallet
				"unlockhash": "18abc0123abc0123abc0123abc0123abc0abc0123abc0123abc0123abc0123abc0efb39211ea2a"
			}
		],
		// miner fees, list of currency values, at least 1 required
		"minerfees": ["1", "2", "3"],
		// arbitrary data, optional, base64 encoded
		"arbitrarydata": "ZGF0YQ==" // represents "data" in base64 encoding
	}
}
```

The details of the text (JSON) encoding of an unlock hash are described in
[/doc/transactions/unlockhash.md](/doc/transactions/unlockhash.md).

You can learn more about atomic swaps in
[/doc/atomicswap/atomicswap.md](/doc/atomicswap/atomicswap.md).

Please read [the output identifiers section](#output-identifiers) to
learn more about how coin- and block stake output identifiers are generated.

Notes for soft-forks:

+ When `"version"` defines an unknown version (anything other than `0` at the moment), the data should simply define the transaction as a base64 encoding of the binary encoding format of a transaction;
+ When an input's `"type"` defines an unknown unlock type (anything more than `2` at the moment), the condition and fulfillment should be the base64 encoding of the binary encoded format of those conditions and fulfillments;

Another, this time very minimalistic, example:

```javascript
{
	"version": 0, // version ID, byte, for now always `0`, optional
	"data": { // actual transaction, required
		"coininputs": [ // coin inputs, at least 1 element required
			{
				// coin output ID, crypto hash (blake2b, 256 bit), required, hex-encoded
				"parentid": "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
				// input lock, required
				"unlocker": {
					"type": 1, 	// input lock, byte, supported range: [1,2], unknown range: [3,255], required
											// `1` = SingleSignature, `2` = AtomicSwap, other = unknown
					"condition": { // unlock condition, required, format dependend upon sibling "type" property
						// public key, required, format: `<algorithmSpecifier>:<key>`, where <key> is hex-encoded
						// and which byte-size is fixed but dependend upon the <algorithmSpecifier>,
						// <algorithmSpecifier> can currently only be `"ed25519"`
						"publickey": "ed25519:def123def123def123def123def123def123def123def123def123def123def1"
					},
					"fulfillment": { 	// unlock fulfillment, fulfills the unlock condition, required,
														// format dependend upon sibling "type" property
						// signature, byte-size is fixed but dependend upon the <algorithmSpecifier>,
						// when <algorithmSpecifier> equals `"ed25519"` the byte size is 64, required, hex-encoded
						"signature": "ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef"
					}
				}
			}
		],
		// miner fees, list of currency values, at least 1 required
		"minerfees": ["1"],
		// arbitrary data, optional, base64 encoded
		"arbitrarydata": "SGVsbG8sIFdvcmxkIQ=="  // represents "Hello, World!" in base64 encoding
	}
}`
```

### Binary Format

While the exact details of a transaction (binary encoding) depend upon which parts are used,
and what types of inputs and outputs are used, all transactions do follow the same structure.

All parts of a transaction, and the serialized transaction as combined serialialized parts,
are encoded using a binary little-endian based encoding algorithm.
While some of it is repeated here as part of the detailed format explanation,
it would still be helpful to read the detailed information that describes how this encoding algorithm
works at [/doc/Encoding.md](/doc/Encoding.md).

All transactions are always serialized in following order:

```
+---------+------+------+------------+------------+-----------+-----------+
| version | coin | coin | blockstake | blockstake | minerfees | arbitrary |
|         | in   | out  | inputs     | outputs    |           | data      |
+---------+------+------+------------+------------+-----------+-----------+
```

Where version is always a single byte with as value `0x00` (until we support a new format), and where
coin outputs, blockstake inputs, blockstake outputs and arbitrary data are all optional.

As coin outputs, blockstake inputs and blockstake outputs are encoded as slices of data,
they'll be encoded as a nil slice `0x0000000000000000` when not given, and thus still take 8 bytes each.

Version, coin inputs and minerfees are required and will therefore always (have to) be defined.

#### outputs

The format of coin outputs and blockstake outputs are the same:

```
+----------+----------+-----+----------+
| length N | output#1 | ... | output#N |
+----------+----------+-----+----------+
| 8 bytes  | where each output is      |
|          | 87 bytes or more          |
```

Where each output is encoded as:

```
+--------------------------------------+-------------+
| currency (big integer)               | unlock hash |
+--------------------------------------+-------------+
| length N     | big integer (N bytes) | (33 bytes)  |
| (8 bytes)    |   absolute value &    |             |
|              |   big endian          |             |
```

As you can see a currency is at least 9 bytes,
due to the fact that we do not accept zero-value currencies.
The length is always encoded as a 64-bit unsigned integer.

An example of an encoded currency is `0x000000000000000101`, which stands for `1`.

The details of the binary encoding of an unlock hash are described in
[/doc/transactions/unlockhash.md](/doc/transactions/unlockhash.md).

#### Miner Fees

Miner fees are encoded as a slice of currencies:

```
+----------+-------------+-----+-------------+
| length N | currency #0 | ... | currency #N |
+----------+-------------+-----+-------------+
| 8 bytes  | where each currency             |
|          | is 9 bytes or more              |
```

Details of encoding of slices and currencies have been
explained earlier in this document.

#### Arbitrary Data

As arbitrary data is an optional slice of bytes and encoded as follows:

```
+----------+-----------------+
| length N | data byte slice |
+----------+-----------------+
| 8 bytes  | 0 or more bytes |
```

Therefore it follows that if no arbitrary data is attached to the transaction,
it will be encoded as the nil slice `0x0000000000000000`. Should one want to attach `"42"`
as arbitrary data to a transaction, it will be encoded as `0x00000000000000023432`.

#### inputs

Blockstake inputs are optional, but coin inputs are required. The latter is explained
due to the fact that minerfees are paid with coin inputs. Both are encoded in the same format however:

```
+-----------+------------------+
| output ID | input lock       |
+-----------+------------------+
| 32 bytes  | X bytes          |
```

The encoding details of output ID is discused in [the others section](#others).

Where the input lock is encoded as:

```
+--------+-----------+-------------+
| type   | condition | fulfillment |
+--------+-----------+-------------+
| 1 byte | Y bytes   | Z bytes     |
```

Both condition and fulfillment are encoded as a raw binary (byte) slice,
and therefore, as you should know by now, are prefixed by the length (a 8 bytes unsigned integer).

While the encoding of the condtion and fulfillment slices are dependent upon the type,
indicated by the byte prefix, the unlock hash (defined in the output defined by the earlier given output ID)
can be computed without having to decode the different parts of the input lock:

```
hash = blake2b_checksum256(binary_encoded_condition)
unlockHash = single_byte_unlock_type + hash
```

Please read about the binary encoding of the unlockhash at
[/doc/transactions/unlockhash.md](/doc/transactions/unlockhash.md)
should you not understand the formula given above.

As the condition and fulfillment parts are encoded as byte slices,
their encoded form will be at least 9 bytes, and usually a lot more.

The only currently known types are:

+ `0x01`: Single Signature
+ `0x02`: Atomic Swap Contract

Other types can be added by means of a soft-fork,
but as long as block creator nodes on the network do not support these types,
they'll not be able to be added to the blockchain as part of a block.

##### Single Signature

When the type (prefix) byte of an input lock equals to `0x01`,
it should be assumed that a single-signature input lock is used for the unspend output.
This also means that the first 2 bytes of the unlockhash of that output should equal `"01"`.

The conditon, equals the receiver's public key, and is encoded as:

```
+---------------------+------------+
| algorithm specifier | key        |
+---------------------+------------+
| 16 bytes            | 8+N bytes  |
```

The public key is encoded as a slice,
which length (N) depends upon the algorithm used.

The algorithm is identified by the 16 bytes algorithm specifier,
and currently the only known/supported algorithm is `"ed25519\0\0\0\0\0\0\0\0\0"`,
in which case the public key size is 32 bytes (`N=32`), and therefore the condition
will have a byte (slice) size of 56 bytes (`16+8+32`) for ed25519 signatures.

The fulfillment is encoded as:

```
+-----------+
| signature |
+-----------+
| X bytes   |
```

Where the exact byte size will depend on the signature algorithm used,
indicated by the public key's specifier.

There is currently only one known specifier, `"ed25519\0\0\0\0\0\0\0\0\0"`,
which signature size is 64 bytes, generated using a private key of 64 bytes.

In this only known/supported signature algorithm (`ed25519`),
the signature is computed/signed using `ed25519_sign(message, privateKey)`,
where message is `blake2b_sum256(encoded_object`), where `encoded_object` is formatted as:

```
+-------------+-------------+--------------+-------------------+
| input index | coin inputs | coin outputs | blockstake inputs | ...
+-------------+-------------+--------------+-------------------+
| 8 bytes     |             |              |                   |

    +--------------------+------------+----------------+
... | blockstake outputs | miner fees | arbitrary data |
    +--------------------+------------+----------------+
    |                    |            | 8+ bytes       |
```

The encoding of coin/blockstake outputs, miner fees and arbtirary data
have already been discussed earlier in this document. Input index is encoded as
a 64-bit unsigned integer and represents the sequential position the input has
within the coin/blockstake input slice.

For this encoding only, each coin- and block stake input is encoded as:

```
+----------+-------------+
| ParentID | unlock hash |
+----------+-------------+
| 32 bytes | 33 bytes    |
```

The ParentID is an output identifier, as discussed in [the output identifier section](#output-identifiers). 

As usual, the details of binary encoding of an unlock hash can
be read at [/doc/transactions/unlockhash.md](/doc/transactions/unlockhash.md).

##### Atomic Swap Contract

When the type (prefix) byte of an input lock equals to `0x02`,
it should be assumed that a single-signature input lock is used for the unspend output.
This also means that the first 2 bytes of the unlockhash of that output should equal `"02"`.

The conditon is encoded as:

```
+---------------------+----------------------+---------------+-----------+
| sender unlock hash  | receiver unlock hash | hashed secret | time lock |
+---------------------+----------------------+---------------+-----------+
| 33 bytes            | 33 bytes             | 32 bytes      | 8 bytes   |
```

As usual, the details of binary encoding of an unlock hash can
be read at [/doc/transactions/unlockhash.md](/doc/transactions/unlockhash.md).

The hashed secret is 32 bytes, and is a sha256 crypto hash with as input the 32-byte secret. The timelock is a Unix timestamp in seconds, encoded as a 64 bit unsigned integer.

```
+---------------------+------------+-----------+----------+
| algorithm specifier | public key | signature | secret   |
+---------------------+------------+-----------+----------+
| 16 bytes            | 8+N bytes  | 8+M bytes | 32 bytes |
```

The public key (algo specifier + public key) has already been discussed before.
Just to be clear: the specifier is for now always assumed to be `"ed25519\0\0\0\0\0\0\0\0\0"`, than the key itself is encoded. However as that is just a byte slice, we first encode the length (N) as an 64-bit unsigned integer, and only than N bytes for the key itself. for `ed25519` the public key size (N) is always 32.

The signature depends upon the algorithm specifier as well, and is for `ed25519` 64 bytes, meaning that `64` will be encoded as a 64-bit unsigned integer, and than the 64 bytes of the `ed25519` signature itself.

The secret has no specific format and only has a byte-length of 32.

The signature is always generated using the formula:

```
signature = crypto_sign_algo(message: blake2b_256(encoded_object), private_key)
```

What crypto sign(ature) algorithm is used,
depends upon the algorithm specifier of the to be used
public key (given as part of the same fulfillment).
Similarly, the size of the private_key is fixed but
also depends upon that same algorithm specifier.

The `encoded_object` is encoded as follows:

```
+-------------+-------------+--------------+-------------------+
| input index | coin inputs | coin outputs | blockstake inputs | ...
+-------------+-------------+--------------+-------------------+
| 8 bytes     |             |              |                   |

    +--------------------+------------+----------------+------------+----------+
... | blockstake outputs | miner fees | arbitrary data | public key | secret   |
    +--------------------+------------+----------------+------------+----------+
    |                    |            | 8+ bytes       |            | 32 bytes |
```

Input index, the zero-index indicating the position of the coin/blockstake input within the transaction's coin/blockstake input slice, is an integer and thus as usual encoded as an unsigned 64 bit integer.
Earlier in this document is already explained how coin outputs, blockstake outputs,
miner fees and arbitrary data are encoded.

he 32 bytes of the secret are simply copied, and no further encoding is applied.
The secret part isn't used as part of the atomic swap's condition,
in case the public key is owned by the sender, or in other words,
the locked output is used as a refund input

Coin inputs and block stake inputs are in this case encoded in the following way:

```
+------------------+-------------------+
| parent output ID | input unlock hash |
+------------------+-------------------+
| 32 bytes         | 33 bytes          |
```

Parent output ID is a 32 byte hash, computed as a blake2b 256bit checksum.
The details of binary encoding of an (input) unlock hash can
be read at [/doc/transactions/unlockhash.md](/doc/transactions/unlockhash.md).

The encoding details of (parent) output ID is discused in [the others section](#others).

The encoding of the public key has already been discussed earlier, but just to repeat:

```
+---------------------+------------+
| algorithm specifier | key        |
+---------------------+------------+
| 16 bytes            | 8+N bytes  |
```

The public key is encoded as a slice,
which length (N) depends upon the algorithm used.

The algorithm is identified by the 16 bytes algorithm specifier,
and currently the only known/supported algorithm is `"ed25519\0\0\0\0\0\0\0\0\0"`,
in which case the public key size is 32 bytes (`N=32`), and therefore the condition
will have a byte (slice) size of 56 bytes (`16+8+32`) for ed25519 signatures.

#### others

In this section we'll discuss the details/format/encoding of anything else relevant.

##### Output Identifiers

All output identifiers are computed the same way,
with the only difference the (prefix) 16 byte specifier used.
Possible output identifiers are coinOutputID and blockstakeOutputID.
With that said, here is how an output ID is formatted:

```
+----------+
| hash     |
+----------+
| 32 bytes |
```

That's it! Just a 32 byte (blake2b, 256 bit) crypto hash. Simple.

But what is used as input message for the blake2b_256 hash?
Glad you asked. The message is binary encoded and formatted as:

```
+-----------+-------------+--------------+-------------------+--------------------+
| specifier | coin inputs | coin outputs | blockstake inputs | blockstake outputs | ...
+-----------+-------------+--------------+-------------------+--------------------+
| 16 bytes  |             |              |                   |                    |

    +------------+----------------+--------------+
... | miner fees | arbitrary data | output index |
    +------------+----------------+--------------+
    |            | 8+ bytes       | 8 bytes      |
```

The encoding of coin/blockstake inputs/outputs, miner fees and arbtirary data
have already been discussed earlier in this document. Output index is encoded as
a 64-bit unsigned integer and represents the sequential position the output has
within the coin/blockstake output slice.

Specifier defines what kind of outputID it is and is one of following:

+ `"coin output\0\0\0\0\0"`: coin output ID
+ `"blstake output\0\0"`: blockstake output ID

##### Unlock hash's hash

As can be read in [/doc/transactions/unlockhash.md](/doc/transactions/unlockhash.md),
an unlock hash text encoding consists out of 3 parts: type, hash and checksum (of hash).

But how is that hash generated? Well, using the blake2b 256-bit algorithm,
we compute a 32-byte crypto hash, using the binary encoding of
the relevant input's unlock condition as hash input.

How the unlock condition is binary encoded depends upon the unlock type used
(and defined in both the input as well as the output,
where it is the prefix of the unlock hash in the latter case).
For unknown types the unlock condition is opaque and binary, and thus nothing has to be done if so.

When the unlock type is of a known type however, the enoding will depend on this type.

You can read more about how each unlock type's condition is binary encoded
in [the inputs section](#inputs).

### examples

In this example we'll explain a couple of example binary encoded transactions
in detail, byte by byte, as to summarize everything discussed in this document so far.

All examples are also unit tested as part of this Golang reference implementation of Rivine,
as to help ensuring that the documentation stays in sync with the code.
You can find the unit tests for all examples listed below
as `TestTransactionEncodingDocExamples` in [/types/transactions_test.go](/types/transactions_test.go).

#### transaction with one single signature coin input and no outputs

Encoded transaction in hex format:

```
0001000000000000002200000000000000000000000000000000000000000000000000000000000022013800000000000000656432353531390000000000000000002000000000000000ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff4000000000000000ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff00000000000000000000000000000000000000000000000001000000000000000100000000000000010000000000000000
```

Breakdown:

+ `00` (version One, `0x00`)
+ `0100000000000000` (coin inputs, length `1` in little endian, 8 bytes):
  + `2200000000000000000000000000000000000000000000000000000000000022` (coinOutputID, 32 bytes, crypto Hash)
  + `01` (input lock type, SingleSignature (`0x01`)):
    + `3800000000000000` (condition length: `48` in little endian, 8 bytes):
      + `65643235353139000000000000000000` (public key, algorithm specifier `ed25519`)
      + `2000000000000000` (public key, key length `32`)
        + `ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff` (public key, key itself, 32 bytes)
    + `4000000000000000` (fulfillment length: `64`, in little endian, 8 bytes):
      + `ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff` (ed25519 signature, 64 bytes)
+ `0000000000000000` (coin outputs, length `0`, in little endian, 8 bytes)
+ `0000000000000000` (blockstake inputs, length `0`, in little endian, 8 bytes)
+ `0000000000000000` (blockstake outputs, length `0`, in little endian, 8 bytes)
+ `0100000000000000` (miner fees, length `1`, in little endian, 8 bytes):
  + `0100000000000000` (currency, length `1`, in little endian, 8 bytes):
    + `01` (currency value `1`, in big endian, 1 byte)
+ `0000000000000000` (arbitrary data, length `0`, in little endian, 8 bytes)


#### transaction with one signle signature coin input and a couple of coin outputs

```
0001000000000000002200000000000000000000000000000000000000000000000000000000000022013800000000000000656432353531390000000000000000002000000000000000ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff4000000000000000ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff 0200000000000000 0100000000000000 02 01 cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc01000000000000000301dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd0000000000000000000000000000000001000000000000000100000000000000010000000000000000
```

Breakdown:

+ `00` (version One, `0x00`)
+ `0100000000000000` (coin inputs, length `1` in little endian, 8 bytes):
  + `2200000000000000000000000000000000000000000000000000000000000022` (coinOutputID, 32 bytes, crypto Hash)
  + `01` (input lock type, SingleSignature (`0x01`)):
    + `3800000000000000` (condition length: `48` in little endian, 8 bytes):
      + `65643235353139000000000000000000` (public key, algorithm specifier `ed25519`)
      + `2000000000000000` (public key, key length `32`)
        + `ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff` (public key, key itself, 32 bytes)
    + `4000000000000000` (fulfillment length: `64`, in little endian, 8 bytes):
      + `ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff` (ed25519 signature, 64 bytes)
+ `0200000000000000` (coin outputs, length `2`, in little endian, 8 bytes):
  + first:
    + `0100000000000000` (currency, length `1`, in little endian, 8 bytes):
      + `02` (currency, value `2`, in big endian)
    + `01` (unlock hash, unlock type `0x01`, single byte):
      + `cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc` (unlock hash, crypto hash, 32 bytes)
  + second:
    + `0100000000000000` (currency, length `1`, in little endian, 8 bytes):
      + `03` (currency, value `3`, in big endian)
    + `01` (unlock hash, unlock type `0x01`, single byte):
      + `dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd` (unlock hash, crypto hash, 32 bytes)
+ `0000000000000000` (blockstake inputs, length `0`, in little endian, 8 bytes)
+ `0000000000000000` (blockstake outputs, length `0`, in little endian, 8 bytes)
+ `0100000000000000` (miner fees, length `1`, in little endian, 8 bytes):
  + `0100000000000000` (currency, length `1`, in little endian, 8 bytes):
    + `01` (currency, value `1`, in big endian)
+ `0000000000000000` (arbitrary data, length `0`, in little endian, 8 bytes)

#### complete transaction multiple coin/blockstake inputs and outputs, as well as arbitrary data

```
0002000000000000002200000000000000000000000000000000000000000000000000000000000022013800000000000000656432353531390000000000000000002000000000000000ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff4000000000000000ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff3300000000000000000000000000000000000000000000000000000000000033026a00000000000000011234567891234567891234567891234567891234567891234567891234567891016363636363636363636363636363636363636363636363636363636363636363bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb07edb85a00000000a000000000000000656432353531390000000000000000002000000000000000abababababababababababababababababababababababababababababababab4000000000000000dededededededededededededededededededededededededededededededededededededededededededededededededededededededededededededededededabadabadabadabadabadabadabadabadabadabadabadabadabadabadabadaba020000000000000001000000000000000201cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc01000000000000000302dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd01000000000000004400000000000000000000000000000000000000000000000000000000000044013800000000000000656432353531390000000000000000002000000000000000eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee4000000000000000eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee010000000000000001000000000000002a01abcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcd010000000000000001000000000000000102000000000000003432
```


Breakdown:

+ `00` (version One, `0x00`)
+ `0200000000000000` (coin inputs, length `2` in little endian, 8 bytes):
  + first:
    + `2200000000000000000000000000000000000000000000000000000000000022` (coinOutputID, 32 bytes, crypto Hash)
    + `01` (input lock type, SingleSignature (`0x01`)):
      + `3800000000000000` (condition length: `48` in little endian, 8 bytes):
        + `65643235353139000000000000000000` (public key, algorithm specifier `ed25519`)
        + `2000000000000000` (public key, key length `32`)
          + `ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff` (public key, key itself, 32 bytes)
      + `4000000000000000` (fulfillment length: `64`, in little endian, 8 bytes):
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
      + `a000000000000000` (fulfillment length: `160`, in little endian, 8 bytes):
        + `65643235353139000000000000000000` (public key, algorithm specifier `ed25519`)
        + `2000000000000000` (public key, key length `32`)
          + `abababababababababababababababababababababababababababababababab` (public key, key itself, 32 bytes)
        + `4000000000000000` (ed25519 signature length `64` in little endian, 8 bytes):
          + `dededededededededededededededededededededededededededededededededededededededededededededededededededededededededededededededede` (ed25519 signature, 64 bytes)
       + `dabadabadabadabadabadabadabadabadabadabadabadabadabadabadabadaba` (secret, 32 bytes, fixed size)
+ `0200000000000000` (coin outputs, length `2`, in little endian, 8 bytes):
  + first:
    + `0100000000000000` (currency, length `1`, in little endian, 8 bytes):
      + `02` (currency, value `2`, in big endian)
    + `01` (unlock hash, unlock type `0x01`, single byte):
      + `cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc` (unlock hash, crypto hash, 32 bytes)
  + second:
    + `0100000000000000` (currency, length `1`, in little endian, 8 bytes):
      + `03` (currency, value `3`, in big endian)
    + `01` (unlock hash, unlock type `0x01`, single byte):
      + `dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd` (unlock hash, crypto hash, 32 bytes)
+ `0100000000000000` (blockstake inputs, length `1`, in little endian, 8 bytes):
  + first:
    + `4400000000000000000000000000000000000000000000000000000000000044` (blockStakeOutputID, crypto Hash, 32 bytes)
    + `01` (input lock type, SingleSignature (`0x01`)):
      + `3800000000000000` (condition length: `48` in little endian, 8 bytes):
        + `65643235353139000000000000000000` (public key, algorithm specifier `ed25519`)
        + `2000000000000000` (public key, key length `32`)
          + `eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee` (public key, key itself, 32 bytes)
      + `4000000000000000` (fulfillment length: `64`, in little endian, 8 bytes):
        + `eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee` (ed25519 signature, 64 bytes)
+ `0100000000000000` (blockstake outputs, length `1`, in little endian, 8 bytes):
  + first:
    + `0100000000000000` (currency, length `1`, in little endian, 8 bytes):
     + `2a` (currency, value `42`, in big endian)
    + `01` (unlock hash, unlock type `0x01`, SingleSignature):
      + `abcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcd` (unlock hash, crypto hash, 32 bytes)
+ `0100000000000000` (miner fees, length `1`, in little endian, 8 bytes):
  + `0100000000000000` (currency, length `1`, in little endian, 8 bytes):
    + `01` (currency, value `1`, in big endian)
+ `0200000000000000` (arbitrary data, length `2`, in little endian, 8 bytes):
  + `3432` (arbitrary data `"42"`)
