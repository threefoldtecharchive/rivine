# Unlock Hash

An unlock Hash identifies the recipient of an output.
Because of this, an unlock is often called "address".

It should be noted however that the recipient is not always a wallet address.
Therefore each unlock hash is prefixed with two bytes,
as to indicate the type of unlock hash.

This document will explain the format of the unlock hash in serialized form.

Please check out [the official Golang implementation](/types/unlockhash.go) for
more detailed and in-depth technical information.

## Format

### encoding

An unlock hash has a different representation, depending if its encoded in text format (string),
or in a raw binary format (byte slice). For JSON marshalling and print/debugging purposes the text format is used,
while the binary encoding is used when used as part of a transaction in the Rivine protocol.

#### text/string encoding

The format of an address can be visually represented as:

```plain
+---------+----------------------------------+-----------------------------+
|  type   |    32-byte hex-formatted hash    | 6-byte hex-encoded checksum | description
|         |                                  |                             |
| 2 bytes |            64 bytes              |           12 bytes          | byte length
+----+----+----+------------------------+----+----+-------------------+----+
| 0  | 01 | 02 |           ...          | 66 | 67 |        ...        | 77 | byte positions
+----+----+----+------------------------+----+----+-------------------+----+
```

As you can see an unlock hash consists out of a type, a hash and a checksum of the type and hash.
The entire unlock hash is hex-formatted, which explains why the actual unlock hash size
is doubled from 39 bytes to 78 bytes. Let's go over all parts of an unlock hash in detail.

#### binary encoding

```plain
+--------+----------+
| type   | hash     |
+--------+----------+
| 1 byte | 32 bytes |
```

The hash is a 32 bytes cryptographic hash,
but how it was generated depends upon the unlock type used.

### type

The type of an unlock hash defines the type of input lock.
Meaning that it defines what kind of condition is used and consequently,
what kind of fulfillment is required. It essentially identifies the requirements
in order to be able to spend the given output as a future input.

Currently there are only 2 known (input) lock types:

+ Public Key (`0x01` -> `"01"`): the unlock hash identifies a wallet address (a public key linked to a wallet);
+ Atomic Swap Contract (`0x02` -> `"02"`): the unlock hash identifies an atomic swap contract between two addresses;

> NOTE: Atomic Swap Contract Unlock Hashes are no longer used (by default) as output conditions since v1 transactions.
> They are however still used as to identify such an output by some modules, such as the explorer,
> as to identify the outputs within a bucket.

The default (input) lock type is `00` and should not be used for anything,
as it identifies a `nil` input lock type, and is only expected for nil inputs.

Any other (of the 253 available) input lock types can be used in future updates,
but remain unusable until the majority of active block stake owners accepts them.

### hash

The hash has no predefined format, and the generation of which is defined by the (input) lock type.
A fixed length of 32 bytes (64 bytes in hex-encoded format) is the only requirement of the hash.

> Sidenote: in the standard reference implementation, written in Golang,
> the standard  `encoding/hex` lib is used for hex-encoding/decoding.
> It writes every byte (value range `0`-`255`) into
> its 2-byte hex representation `"00"`-`"ff"`.

But how is that hash generated? Well, using the blake2b 256-bit algorithm,
we compute a 32-byte crypto hash, using the binary encoding of
the relevant input's unlock condition as hash input.

How the unlock hash is binary encoded depends upon the unlock type used.
When the unlock type is of a known type however, the enoding will depend on this type.

#### Public Key Unlock Hash

A public key (`0x01`) unlock hash's hash is computed as follows:

```plain
blake2b_256(binaryEncoding(publicKey))
```

Where the binary encoded layout of a public key is as follows:

```plain
+----------------------+-----------------+------------------------+
|     fixed-size       |  length N of    |                        |
|  array, indicating   |   public key    |       public key       | description
|  the signature algo  |  which follows  |                        |
|                      |                 |                        |
|       16 bytes       |     8 bytes     |        N bytes         | byte length
+----+----+-------+----+----+------------+----+----+-------+------+
| 00 | 01 |  ...  | 15 | 16 |  ...  | 23 | 24 | 25 |  ...  | 23+N | byte positions
+----+----+-------+----+----+------------+----+----+-------+------+
```

> Implemented in the official/reference Golang implementation
> as the `NewPubKeyUnlockHash` function in [/types/unlockhash.go](/types/unlockhash.go).
>
> Documentation of this function, and reference to its source,
> is available at [https://godoc.org/github.com/rivine/rivine/types#NewPubKeyUnlockHash](https://godoc.org/github.com/rivine/rivine/types#NewPubKeyUnlockHash).

#### Atomic Swap Unlock Hash

A Atomic Swap (`0x02`) unlock hash's hash,
remaining active for legacy reasons only, is computed as follows:

```plain
blake2b_256(binaryEncoding(atomicSwapCondition))
```

Where the binary encoded layout of an atomic swap's condition is as follows:

```plain
+----------------+----------------+---------------+----------------+
|    sender      |    receiver    |    hashed     |    timelock    |
|   unlockhash   |   unlockhash   |    secret     |     uint64     | description
|                |                |               |                |
|    33 bytes    |    33 bytes    |   32 bytes    |     8 bytes    | byte length
+----+------+----+----+------+----+----+-----+----+----+-----+-----+
| 00 | ...  | 32 | 33 | ...  | 65 | 66 | ... | 97 | 98 | ... | 105 | byte positions
+----+------+----+----+------+----+----+-----+----+----+-----+-----+
```

> Implemented in the official/reference Golang implementation
> as the `AtomicSwapCondition`'s `UnlockHash` method in [/types/unlockcondition.go](/types/unlockcondition.go).
>
> Documentation of this function, and reference to its source,
> is available at [https://godoc.org/github.com/rivine/rivine/types#NewPubKeyUnlockHash](https://godoc.org/github.com/rivine/rivine/types#NewPubKeyUnlockHash).

### checksum

When encoding the unlockhash in text/string format,
the last part of a unlock hash is the checksum of the type and hash (which were the previous 2 parts).
When encoding the unlockhash in binary format, no checksum is generated as part of the encoding.

The blake2b algorithm is used to generate a 256-bit checksum from the hash, of which the first 6 bytes are used
as checksum for this unlock hash. As the checksum is also hex-encoded its byte size doubles to 12.

> ```plain
> checksum := first_6_bytes(blake2b_256(type, hash))
> ```
> > where type is one byte, and hash is a fixed-size byte array of 32 bytes
> > returned checksum is 32 bytes (as 256 bit version of the blake2b hash algo is used)

Meaning that the input message used to generate a 256-bit checksum with Blake2b is 33 bytes,
one byte of the (unlock type) and 32 bytes for the hash itself. The output (checksum) is 32 bytes.

See [the text/string encoding section](#text/string-encoding) for more information about
how this checksum is used as part of the text encoding.
