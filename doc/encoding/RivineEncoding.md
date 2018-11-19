# Rivine Binary Encoding

The main goal of the rivine (binary) encoding library is to achieve the smallest byte footprint for encoded content.
This encoding library is heavily inspired upon the initial Sia (binary) encoding library.

## Standard Encoding

All integers are little-endian, encoded as unsigned integers, but the amount of types depend on the exact integral type:

| byte size | types |
| - | - |
| 1 | uint8, int8 |
| 2 | uint16, int16 |
| 3 | uint24<sup>(1)</sup> |
| 4 | uint32, int32 |
| 8 | uint64, int64, uint, int |

> (1) `uint24` is not a standard type, but the tfchain encoding lib does allow to encode uint32 integers that fit in 3 bytes, as 3 bytes.

Booleans are encoded as a single byte, `0x00` for `False` and `0x01` for `True`.

Nil pointers are equivalent to "False", i.e. a single zero byte. Valid pointers are represented by a "True" byte (0x01) followed by the encoding of the dereferenced value.

Variable-length types, such as strings and slices, are represented by a length prefix followed by the encoded value. Strings are encoded as their literal UTF-8 bytes. Slices are encoded as the concatenation of their encoded elements. The length prefix can be one, two, three or four bytes:

| byte size | inclusive size range |
| - | - |
| 1 | 0 - 127 |
| 2 | 128 - 16 383 |
| 3 | 16 384 - 2 097 151 |
| 4 | 2 097 152 - 536 870 911 |

This implies that variable-length types cannot have a size greater than `536 870 911`,
which to be fair is a very big limit for blockchain purposes. Perhaps too big of a limit already,
as it is expected that for most purposes the slice length will fit in a single byte, and the extreme cases in 2 bytes.

Maps are not supported; attempting to encode a map will cause Marshal to panic. This is because their elements are not ordered in a consistent way, and it is imperative that this encoding scheme be deterministic. To encode a map, either convert it to a slice of structs, or define a MarshalSia method (see below).

Arrays and structs are simply the concatenation of their encoded elements (no length prefix is required here as the size is fixed). Byte slices are not subject to the 8-byte integer rule; they are encoded as their literal representation, one byte per byte.

All struct fields must be exported. The ordering of struct fields is determined by their type definition.

Finally, if a type implements the `RivineMarshaler` interface, its `MarshalRivine` method will be used to encode the type. Similarly, if a type implements the `RivineUnmarshaler` interface, its `UnmarshalRivine` method will be used to decode the type. Note that unless a type implements both interfaces, it must conform to the spec above. Otherwise, it may encode and decode itself however desired. This may be an attractive option where speed is critical, since it allows for more compact representations, and bypasses the use of reflection.
