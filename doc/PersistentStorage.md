# Persistent Storage

A Rivine Daemon (`rivined`) stores data on persistent storage (e.g. a HDD/SDD),
it does so using a library which itself calls the OS-provided System Calls, to read and write data to file(s).

Most of the data is managed using [bbolt][bbolt], a minimalistic, performant and embedded key-value storage.
We use a fork of the etcd-io organization (which is the fork of itself of the original repo by Ben Johnson).
The other kind of data is stored directly as JSON-encoded bytes.

Each (standard) module (of the daemon) defines exactly how to store data and what to use.
All of them however use either `bbolt` or a hybrid JSON-encoded file
(more about this in the [Persist Utilities](#persist-utilities) chapter),
there are no exceptions. There is also a `persist` root package, which defines several primitives used by
all modules in order to store the persistent data. Prior to going over what data
is stored by each module (and how it is stored), it might be useful to know
what primitives are made available to each module.

## Persist Utilities

In the [`/persist`](https://godoc.org/github.com/rivine/rivine/persist) package one can find
common utilities used by the different daemon modules in order to store data to persistent memory.

In it you'll find:

* functions to [Load](https://godoc.org/github.com/rivine/rivine/persist#LoadJSON) and
  [Save](https://godoc.org/github.com/rivine/rivine/persist#SaveJSON) JSON files;
  * Note that prior to writing the given JSON object it will write out 3 lines, each containing a JSON-encoded string:
    1. Metadata header;
    2. Metadata version (the version the persistent data stored in that file changed last,
       given our historical connections with Sia it could be both the Sia version as well as the Rivine version,
       with the latter only being the case when we introduced the change ourselves);
    3. Checksum of the data;
  * as a result the file content, taken as a whole, is not a valid JSON object,
    but rather a collection of four valid JSON objects
    (one per line for the first 3 objects, and the rest of the lines for the fourth object).
* [the Logger struct type](https://godoc.org/github.com/rivine/rivine/persist#Logger) type
  (using the [std Logger type](https://godoc.org/log#Logger)
  with an [OS file](https://godoc.org/os#File) as the [Writer](https://godoc.org/io#Writer)),
  used by all modules for logging purposes;
* [the Metadata struct type](https://godoc.org/github.com/rivine/rivine/persist#Metadata),
  used by modules to store the Header (name) and Version of a module,
  useful to ensure version (backwards) compatibility;
* [the BoltDatabase struct type](https://godoc.org/github.com/rivine/rivine/persist#BoltDatabase),
  wrapping around the default [bbolt.DB](https://godoc.org/github.com/rivine/bbolt#DB),
  adding the default integration of [the earlier mentioned Metadata struct type](https://godoc.org/github.com/rivine/rivine/persist#Metadata), as a way to identify each DB file by a name and version;
* some other tiny utility functions...

### Modules

Each module —defined in the Rivine library— stores data to persistent storage:

- it keeps track of the state, such that an instance can pick up where it left of last time it ran using that data;
- it ensures that the entire state (based on the blockchain state) does not have to be kept in volatile memory.

Most if this data is managed using [bbolt][bbolt], using our [wrapped `BoltDatabase` struct type](https://godoc.org/github.com/rivine/rivine/persist#BoltDatabase), but some modules might also choose
to store data to persistent memory directly as a JSON-encoded OS file.

When navigating through the files stored on your file system by a Rivine daemon
you might also encounter files with the "_temp" suffix.
These are files used as a backup, should the file on the actual location have an issue somehow.

#### BlockCreator

All persistent data of the BlockCreator module is stored on the file system
within the `<root_dir>/<network>/blockcreator` directory.

The BlockCreator module stores all persistent data in a single JSON-encoded OS file.

> `blockcreator.json`

```
BlockCreatorDir Settings
0.0.1
```

Used to keep track of its state and required when subscribing to the ConsensusSet.

Contains:
* `RecentChange`: the last known [`ConsensusChangeID`](https://godoc.org/github.com/rivine/rivine/modules#ConsensusChangeID);
* `Height`: the last known [`BlockHeight`](https://godoc.org/github.com/rivine/rivine/types#BlockHeight);
* `ParentID`: the last known ID of the parent (meaning the [`BlockID`](https://godoc.org/github.com/rivine/rivine/types#BlockHeight) of the block prior to the current Block);

> `blockcreator.log`

The (appended) text file used for logging purposes.
Each new log —triggered within the daemon's BlockCreator module— will append a line (of logging info) to that file.

#### ConsensusSet

All persistent data of the ConsensusSet module is stored on the file system
within the `<root_dir>/<network>/consensus` directory.

The ConsensusSet module stores all persistent data in a single OS file, managed by [bbolt][bbolt].

> `consensus.db`

```
Consensus Set Database
1.0.5
```

Used to persist all state and data of the Consensus Module.
The correctness of this data at all times is important,
as it is assumed to be correct by other modules when received,
regardless if it is served hot from the blockchain or cold from disk.

Contains:
* bucket `"BlockHeight"`:
  * used to store the current block height (binary encoded, 8 bytes);
  * information is also cached in volatile memory of a running consensus-enabled daemon;
* bucket `"BlockMap"`:
  * used to map blocks to their
    [identifier](https://godoc.org/github.com/rivine/rivine/types#BlockID),
    storing the blocks themselves as [processed blocks](https://github.com/rivine/rivine/blob/master/modules/consensus/processedblock.go),
    a type which also stores the depth, height, child target and diffs alongside
    [the block itself](https://godoc.org/github.com/rivine/rivine/types#Block);
    * Note that this includes blocks that are not currently in the
      consensus set, and blocks that may not have been fully validated yet.
* bucket `"BlockPath"`:
  * maps all [block heights](https://godoc.org/github.com/rivine/rivine/types#BlockHeight) to
    the [identifier of the block stored at that height](https://godoc.org/github.com/rivine/rivine/types#BlockID);
    * Note that this includes blocks only at the current path,
      meaning blocks for the according-to-this-consensus-longest (block) chain;
* bucket `"Consistency"`:
  * a byte representing whether the ConsensusDB is inconsistent (1) or not (0);
  * used to ensure the data stored in this db can be trusted;
* bucket `"CoinOutputs"`:
  * stores all _unused_ [coin outputs](https://godoc.org/github.com/rivine/rivine/types#CoinOutput) using their
    [coin output identifier](https://godoc.org/github.com/rivine/rivine/types#CoinOutputID);
  * used to validate transactions that attempt to spend coin outputs;
* bucket `"BlockStakeOutputs"`:
  * stores all _unused_
    [block stake outputs](https://godoc.org/github.com/rivine/rivine/types#BlockStakeOutput) using their
    [block stake output identifier](https://godoc.org/github.com/rivine/rivine/types#BlockStakeOutputID);
  * used to validate transactions that attempt to spend block stake outputs;
* bucket `"TransactionIDMap"`:
  * maps all [Tx short identifiers](https://godoc.org/github.com/rivine/rivine/types#TransactionShortID) to
    their [regular Tx identifier](https://godoc.org/github.com/rivine/rivine/types#TransactionID);
  * used by other modules to be able to fetch a Tx using its (unique) short identifier;
* bucket `"dco_<height>"`:
    * used to keep track of all delayed coin outputs;
    * stores all _delayed_ [coin outputs](https://godoc.org/github.com/rivine/rivine/types#CoinOutput)
      on [the block height](https://godoc.org/github.com/rivine/rivine/types#BlockHeight)
      as identified by the bucket using their [coin output identifier](https://godoc.org/github.com/rivine/rivine/types#CoinOutputID);

> `consensus.log`

The (appended) text file used for logging purposes.
Each new log —triggered within the daemon's ConsensusSet module— will append a line (of logging info) to that file.

#### Explorer

All persistent data of the Explorer module is stored on the file system
within the `<root_dir>/<network>/explorer` directory.

The Explorer module stores all persistent data in a single OS file, managed by [bbolt][bbolt].

> `explorer.db`

```
Sia Explorer
1.0.8
```

Used to persist all data indexed by the Explorer module.
Pretty much all data is indexed and stored (multiple times),
allowing it to swiftly return any requested blockchain-data, directly from disk.

Contains:
* bucket `"BlockFacts"`:
  * maps all [block identifiers](https://godoc.org/github.com/rivine/rivine/types#BlockID) to
    their [facts (a bunch of statistics about the consensus set as they were at a specific block)](https://godoc.org/github.com/rivine/rivine/modules#BlockFacts);
* bucket `"BlocksDifficulty"`: no longer used;
* bucket `"BlockTargets"`:
  * maps all [block identifiers](https://godoc.org/github.com/rivine/rivine/types#BlockID) to
    their [block targets](https://godoc.org/github.com/rivine/rivine/types#Target);
* bucket `"Internal"`: stores the internal state of the Explorer Module:
  * `BlockHeight`: the last known block height;
  * `RecentChange`: used to store the last known
    [`ConsensusChangeID`](https://godoc.org/github.com/rivine/rivine/modules#ConsensusChangeID),
    needed for subscribing to the ConsensusSet;
* bucket `"CoinOutputIDs"`:
  * maps all [coin output identifiers](https://godoc.org/github.com/rivine/rivine/types#CoinOutputID) to
    the [identifier of the transaction they are part of](https://godoc.org/github.com/rivine/rivine/types#TransactionID);
* bucket `"CoinOutputs"`:
  * stores all [coin outputs](https://godoc.org/github.com/rivine/rivine/types#CoinOutput) using their
    [coin output identifier](https://godoc.org/github.com/rivine/rivine/types#CoinOutputID);
* bucket `"BlockStakeOutputIDs"`:
  * maps all [block stake output identifiers](https://godoc.org/github.com/rivine/rivine/types#BlockStakeOutputID) to
    the [identifier of the transaction they are part of](https://godoc.org/github.com/rivine/rivine/types#TransactionID);
* bucket `"TransactionIDs"`:
  * maps all [transaction identifiers](https://godoc.org/github.com/rivine/rivine/types#TransactionID) to
    the [height of the block they are part of](https://godoc.org/github.com/rivine/rivine/types#BlockHeight);
* bucket `"UnlockHashes"`:
  * contains a bucket for each unlock hash (e.g. wallet addresses, contracts, ...);
  * each internal bucket contains all transaction identifiers the unlock hash is referenced by;
* bucket `"WalletAddressToMultiSigAddressMapping"`:
  * used to map (single-signature) wallet addresses to all the multi-signature addresses they are part of;
  * each (single-signature) wallet address has is a key for a separate internal bucket;
  * each internal wallet bucket contains one bucket per (multi-signature) wallet address it is part off;
  * each multi-signature wallet bucket (within each single-signature wallet bucket) contains the list of transaction identifiers
    the multi-signature wallet is referenced by;

#### Gateway

All persistent data of the Gateway module is stored on the file system
within the `<root_dir>/<network>/gateway` directory.

The Gateway module stores all persistent data (the nodes list) JSON-encoded OS file.

> `nodes.json`

```
Sia Node List
1.3.0
```

Used to keep track of all known inbound/outbound nodes, to which the
gateway has been previously connected. When a gateway module spawns,
it attempts to first connect to the bootstrap nodes, if any, and it will
than attempt to fill its node list using outbound nodes of this list.
As a last resort it will try to connect to the inbound nodes of this list.
In between all this it is also possible that it will try to connect
to nodes given via other peers or that tried to made a connection to the gateway.

Contains a list of node entries, with each entry containing:
* `netaddress`: the public network address (including port) the node was reachable on, last connection with it;
* `wasoutboundpeer`: identifies whether or not it is an outbound peer;

> `gateway.log`

The (appended) text file used for logging purposes.
Each new log —triggered within the daemon's Gateway module— will append a line (of logging info) to that file.

#### TransactionPool

All persistent data of the TransactionPool module is stored on the file system
within the `<root_dir>/<network>/transactionpool` directory.

The TransactionPool module stores all persistent data in a single OS file, managed by [bbolt][bbolt].

> `transactionpool.db`

```
Sia Transaction Pool DB
0.6.0
```

Used to make the Transaction Module's state persistent.

Contains:
* bucket `"ConfirmedTransactions"`: contains the identifiers of all confirmed transactions
  (confirmed as in part of a block created on the blockchain);
* bucket `"RecentConsensusChange"`: used to store the last known
  [`ConsensusChangeID`](https://godoc.org/github.com/rivine/rivine/modules#ConsensusChangeID),
  needed for subscribing to the ConsensusSet;

#### Wallet

All persistent data of the Gateway module is stored on the file system
within the `<root_dir>/<network>/gateway` directory.

The Gateway module stores all persistent data in two JSON-encoded OS files.

> `wallet.json`

```
Wallet Database
0.4.0
```

Used to store all information that identifies and configures the wallet,
which also includes the encrypted seed.

Contains:
* `UID`: a crypto-random unique 32-byte ID, as to prevent leaking information, should the same key be used for multiple wallets;  
* `EncryptionVerification`: 32 bytes of pure zeros, encrypted and used to verify the encryption;
* `PrimarySeedFile`: the primary seed, stored in the format used for backups of seeds;
* `PrimarySeedProgress`: progress counter, identifying how many keys have already been generated for the primary seed, within this wallet;
* `AuxiliarySeedFiles`: extra (non-primary) seeds, loaded using the `wallet load` command or REST API;
* `UnseededKeys`: currently not supported by Rivine and always `null`
  (an issue is open about this: <https://github.com/rivine/rivine/issues/159>).

> Currently the Seed is always encrypted, using the key defined when creating the daemon-hosted wallet,
> there are however plans to supported unencrypted wallets, meaning the seed would also be stored
> unencrypted. This is a feature that should be used in secure and isolated environments only.
> See <https://github.com/rivine/rivine/issues/345> for more information.

> `<chain_name> Wallet Encrypted Backup Seed - <crypto_random_suffix>.seed`

```
Wallet Seed
0.4.0
```

Used to store a backup of the seed on disk, encrypted for security.

Contains:
* `UID`: a crypto-random unique 32-byte ID, as to prevent leaking information, should the same key be used for multiple wallets;  
* `EncryptionVerification`: 32 bytes of pure zeros, encrypted and used to verify the encryption;
* `Seed`: the backed up seed, encrypted for security;

> Currently the Seed is always encrypted, using the key defined when creating the daemon-hosted wallet,
> there are however plans to supported unencrypted wallets, meaning the seed would also be stored
> unencrypted. This is a feature that should be used in secure and isolated environments only.
> See <https://github.com/rivine/rivine/issues/345> for more information.

> `wallet.log`

The (appended) text file used for logging purposes.
Each new log —triggered within the daemon's Wallet module— will append a line (of logging info) to that file.

[bbolt]: https://github.com/rivine/bbolt
