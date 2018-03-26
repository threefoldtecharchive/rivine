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

```
+---------+----------------------------------+-----------------------------+
|  type   |    32-byte hex-formatted hash    | 6-byte hex-encoded checksum |
|         |                                  |                             | description
| 2 bytes |            64 bytes              |           12 bytes          | + byte length
+----+----+---+-------------------------+----+----+-------------------+----+
| 0  | 1  | 2 |           ...           | 66 | 67 |        ...        | 77 | byte positions
+----+----+---+-------------------------+----+----+-------------------+----+
```

As you can see an unlock hash consists out of a type, a hash and a checksum of that hash.
The entire unlock hash is hex-formatted, which explains why the actual unlock hash size
is doubled from 39 bytes to 78 bytes. Let's go over all parts of an unlock hash in detail.

#### binary encoding

```
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

+ Single Signature (`0x01` -> `"01"`): the unlock hash identifies a wallet address (a public key linked to a wallet);
+ Atomic Swap Contract (`0x02` -> `"02"`): the unlock hash identifies an atomic swap contract between two addresses;

The default (input) lock type is `00` and should not be used for anything,
as it identifies a `nil` input lock type, and is only expected for nil inputs.

Any other (of the 253 available) input lock types can be used for soft-forks (e.g. future updates),
and will be able to be used safely already, but the network (of blockc reators)
won't use them for block creation until the transaction's types become known.

### hash

The hash has no predefined format, and the generation of which is defined by the (input) lock type.
A fixed length of 32 bytes (64 bytes in hex-encoded format) is the only requirement of the hash.

> Sidenote: in the standard reference implementation, written in Golang,
> the standard  `encoding/hex` lib is used for hex-encoding/decoding.
> It writes every byte (value range `0`-`255`) into
> its 2-byte hex representation `"00"`-`"ff"`.

For more information about the generation of these hash parts,
I kindly refer you to the more extensive [transactions.md docs](transaction.md#unlock-hashs-hash).

### checksum

The last part of a unlock hash is the checksum of the hash (which was the previous part).
Blake2b is used to generate a 256-bit checksum from the hash, of which the first 6 bytes are used
as checksum for this unlock hash. As the checksum is also hex-encoded its byte size doubles to 12.
