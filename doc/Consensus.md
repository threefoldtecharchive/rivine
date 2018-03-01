Consensus Rules
===============

This document is meant to provide a good high level overview of the Sia
cryptosystem, but does not fully explain all of the small details. The most
accurate explanation of the consensus rules is the consensus package (and all
dependencies).

This document will be more understandable if you have a general understanding
of proof of work blockchains, and does not try to build up from first
principles.

Cryptographic Algorithms
------------------------

Rivine uses cryptographic hashing and cryptographic signing, each of which has
many potentially secure algorithms that can be used. We acknowledge our
inexperience, and that we have chosen these algorithms not because of our own
confidence in their properties, but because other people seem confident in
their properties.

For hashing, our primary goal is hashing speed on consumer hardware, including
phones and other low power devices.

For signing, our primary goal is verification speed. A secondary goal is an
algorithm that supports HD keys. A tertiary goal is an algorithm that supports
threshold signatures.

#### Hashing: blake2b

  [blake2b](http://en.wikipedia.org/wiki/BLAKE_%28hash_function%29#BLAKE2 "Wiki page") has been chosen as a hashing algorithm because it is fast, it has had
  substantial review, and it is invulnerable to length extension attacks.

#### Signatures: variable type signatures

  Each public key will have an specifier (a 16 byte array) and a byte slice
  containing an encoding of the public key. The specifier will tell the
  signature verification which signing algorithm to use when verifying a
  signature. Each signature will be a byte slice, the encoding can be
  determined by looking at the specifier of the corresponding public key.

  This method allows new signature types to be easily added to the currency in
  a way that does not invalidate existing outputs and keys. Adding a new
  signature type requires a hard fork, but allows easy protection against
  cryptographic breaks, and easy migration to new cryptography if there are any
  breakthroughs in areas like verification speed, ring signatures, etc.

  Allowed algorithms:

  ed25519: The specifier must match the string "ed25519". The public key
  must be encoded into 32 bytes. Signatures and public keys will need to
  follow the ed25519 specification. More information can be found at
  ed25519.cr.yp.to

  entropy: The specifier must match the string "entropy". The signature will
  always be invalid. This provides a way to add entropy buffers to
  SpendCondition objects to protect low entropy information, while being able
  to prove that the entropy buffers are invalid public keys.

  There are plans to also add ECDSA secp256k1 and Schnorr secp256k1. New
  signing algorithms can be added to Sia through a soft fork, because
  unrecognized algorithm types are always considered to have valid signatures.

Currency
--------

The Sia cryptosystem has two types of currency. The first is the Siacoin.
Siacoins are generated every block and distributed to the miners. These miners
can then use the siacoins to fund file contracts, or can send the siacoins to
other parties. The siacoin is represented by an infinite precision unsigned
integer.

The second currency in the Sia cryptosystem is the Siafund, which is a special
asset limited to 10,000 indivisible units. Each time a file contract payout is
made, 3.9% of the payout is put into the siafund pool. The number of siacoins
in the siafund pool must always be divisible by 10,000; the number of coins
taken from the payout is rounded down to the nearest 10,000. The siafund is
also represented by an infinite precision unsigned integer.

Siafund owners can collect the siacoins in the siafund pool. For every 10,000
siacoins added to the siafund pool, a siafund owner can withdraw 1 siacoin.
Approx. 8790 siafunds are owned by Nebulous Inc. The remaining siafunds are
owned by early backers of the Sia project.

There are future plans to enable sidechain compatibility with Sia. This would
allow other currencies such as Bitcoin to be spent in all the same places that
the Siacoin can be spent.

Marshaling
-----------

Many of the Sia types need to be hashed at some point, which requires having a
consistent algorithm for marshaling types into a set of bytes that can be
hashed. The following rules are used for hashing:

 - Integers are little-endian, and are always encoded as 8 bytes.
 - Bools are encoded as one byte, where zero is false and one is true.
 - Variable length types such as strings are prefaced by 8 bytes containing
   their length.
 - Arrays and structs are encoded as their individual elements concatenated
   together. The ordering of the struct is determined by the struct definition.
   There is only one way to encode each struct.
 - The Currency type (an infinite precision integer) is encoded in big endian
   using as many bytes as necessary to represent the underlying number. As it
   is a variable length type, it is prefixed by 8 bytes containing the length.

Block Size
----------

The maximum block size is 2e6 bytes. There is no limit on transaction size,
though it must fit inside of the block. Most miners enforce a size limit of
16e3 bytes per transaction.

Block Timestamps
----------------

Each block has a minimum allowed timestamp. The minimum timestamp is found by
taking the median timestamp of the previous 11 blocks. If there are not 11
previous blocks, the genesis timestamp is used repeatedly.

Blocks will be rejected if they are timestamped more than three hours in the
future, but can be accepted again once enough time has passed.

Block ID
--------

The ID of a block is derived using:
	Hash(Parent Block ID + 64 bit Nonce + Block Merkle Root)

The block Merkle root is obtained by creating a Merkle tree whose leaves are
the hash of the timestamp, the hashes of the miner outputs (one leaf per miner
output), and the hashes of the transactions (one leaf per transaction).

Block Target
------------

For a block to be valid, the id of the block must be below a certain target.
The target is adjusted once every 500 blocks, and it is adjusted by looking at
the timestamps of the previous 1000 blocks. The expected amount of time passed
between the most recent block and the 1000th previous block is 10e3 minutes. If
more time has passed, the target is lowered. If less time has passed, the
target is increased. Each adjustment can adjust the target by up to 2.5x.

The target is changed in proportion to the difference in time (If the time was
half of what was expected, the new target is 1/2 the old target). There is a
clamp on the adjustment. In one block, the target cannot adjust upwards by more
more than 1001/1000, and cannot adjust downwards by more than 999/1000.

The new target is calculated using (expected time passed in seconds) / (actual
time passed in seconds) * (current target). The division and multiplication
should be done using infinite precision, and the result should be truncated.

If there are not 1000 blocks, the genesis timestamp is used for comparison.
The expected time is (10 minutes * block height).

Block Subsidy
-------------

The coinbase for a block is (300,000 - height) * 10^24, with a minimum of
30,000 \* 10^24. Any miner fees get added to the coinbase to create the block
subsidy. The block subsidy is then given to multiple outputs, called the miner
payouts. The total value of the miner payouts must equal the block subsidy.

The ids of the outputs created by the miner payouts is determined by taking the
block id and concatenating the index of the payout that the output corresponds
to.

The outputs created by the block subsidy cannot be spent for 50 blocks, and are
not considered a part of the consensus set until 50 blocks have transpired.
This limitation is in place because a simple blockchain reorganization is
enough to invalidate the output; double spend attacks and false spend attacks
are much easier to execute.

Transactions
------------

A Transaction is composed of the following:

- Siacoin Inputs
- Siacoin Outputs
- BlockStake Inputs
- BlockStake Outputs
- Miner Fees
- Arbitrary Data
- Transaction Signatures

The sum of all the siacoin inputs must equal the sum of all the miner fees and
siacoin outputs. There can be no leftovers. The sum
of all BlockStake inputs must equal the sum of all BlockStake outputs.

Several objects have unlock hashes. An unlock hash is the Merkle root of the
'unlock conditions' object. The unlock conditions contain a timelock, a number
of required signatures, and a set of public keys that can be used during
signing.

The Merkle root of the unlock condition objects is formed by taking the Merkle
root of a tree whose leaves are the timelock, the public keys (one leaf per
key), and the number of signatures. This ordering is chosen specifically
because the timelock and the number of signatures are low entropy. By using
random data as the first and last public key, you can make it safe to reveal
any of the public keys without revealing the low entropy items.

The unlock conditions cannot be satisfied until enough signatures have
provided, and until the height of the blockchain is at least equal to the value
of the timelock.

The unlock conditions contains a set of public keys which can each be used only
once when providing signatures. The same public key can be listed twice, which
means that it can be used twice. The number of required signatures indicates
how many public keys must be used to validate the input. If required signatures
is '0', the input is effectively 'anyone can spend'. If the required signature
count is greater than the number of public keys, the input is unspendable.
There must be exactly enough signatures. For example, if there are 3 public
keys and only two required signatures, then only two signatures can be included
into the transaction.

Siacoin Inputs
--------------

Each input spends an output. The output being spent must exist in the consensus
set. The 'value' field of the output indicates how many siacoins must be used
in the outputs of the transaction. Valid outputs are miner fees, siacoin
outputs, and contract payouts.

Siacoin Outputs
---------------

Siacoin outputs contain a value and an unlock hash (also called a coin
address). The unlock hash is the Merkle root of the spend conditions that must
be met to spend the output.

BlockStake Inputs
-----------------

A blockstake input works similar to a siacoin input. It contains the id of a
blockstake output being spent, and the unlock conditions required to spend the
output.

BlockStake Outputs
------------------

Like siacoin outputs, blockstake outputs contain a value and an unlock hash. The
value indicates the number of blockstakes that are put into the output, and the
unlock hash is the Merkle root of the unlock conditions object which allows the
output to be spent.

Miner Fees
----------


Arbitrary Data
--------------

Arbitrary data is a set of data that is ignored by consensus. In the future, it
may be used for soft forks, paired with 'anyone can spend' transactions. In the
meantime, it is an easy way for third party applications to make use of the
siacoin blockchain.

Transaction Signatures
----------------------

Each signature points to a single public key index in a single unlock
conditions object. No two signatures can point to the same public key index for
the same set of unlock conditions.

Each signature also contains a timelock, and is not valid until the blockchain
has reached a height equal to the timelock height.

Signatures also have a 'covered fields' object, which indicates which parts of
the transaction get included in the signature. There is a 'whole transaction'
flag, which indicates that every part of the transaction except for the
signatures gets included, which eliminates any malleability outside of the
signatures. The signatures can also be individually included, to enforce that
your signature is only valid if certain other signatures are present.

If the 'whole transaction' is not set, all fields need to be added manually,
and additional parties can add new fields, meaning the transaction will be
malleable. This does however allow other parties to add additional inputs,
fees, etc. after you have signed the transaction without invalidating your
signature. If the whole transaction flag is set, all other elements in the
covered fields object must be empty except for the signatures field.

The covered fields object contains a slice of indexes for each element of the
transaction (siacoin inputs, minting fees, etc.). The slice must be sorted, and
there can be no repeated elements.

Entirely nonmalleable transactions can be achieved by setting the 'whole
transaction' flag and then providing the last signature, including every other
signature in your signature. Because no frivolous signatures are allowed, the
transaction cannot be changed without your signature being invalidated.
