# Data replication

In order to replicate data stored on the chain to an external database, the `datastore` module is provided. In order to make use of it, the `d` flag must be passed to the modules string (`-M` flag) when starting the rivine daemon. The `datastore` module depends on the `consensus` module, which in turn depends on the `gateway` module. An minimal setup for example would be: 

```bash
rivined -M gcd
```

This will set up a rivine daemon which acts as a full node (it has a full copy of the consensus set, and will relay incomming blocks to its peers), while at the same time offer the ability to replicate data to an external database. Data is given an `UID` when it gets replicated on the external database. This `UID` is the sequential order of appearance, local to the namespace (see below).

## Saving data

Data is saved on the blockchain by adding it as `arbitrary data` in a  transaction. Although there is no hard requirement to the structure of such arbitrary data, the `datastore` module expects it to be of the form: `specifier|namespace|actualdata`. By default, a specifier has a length of 16 bytes, and a namespace has a length of 4 bytes. The specifier is used by the transactionpool and consensus set to validate arbitrary data. Data which does not start with a validated specifier causes the transaction to be rejected. As such, the datastore currently ignores the specifier altoghether. The only whitelisted specifier is currently `NonSia` (padded with 0-value bytes to a length of 16 bytes).

In order to allow people to store data on the blockchain, a command is exposed in the `client` which creates a minimal transaction to the provided address, while paying the miner fee, and adding the provided data. This command silently adds a specifier (the specifier being `NonSia`, as to not have the transaction rejected by the `transactionpool` for containing invalid data). The command is as follows:

```bash
rivinec wallet registerdata [destination address] [data]
```

The destination address can be any valid address (inlcuding one of your own addresses). Keeping in mind the structure the data should have, as well as the fact that the client adds the specifier by itself, we only need to add the namespace to the data. For example, if we want to write "testdata" to the namespace "1111", we can do so with the following command:

```bash
rivinec wallet registerdata 53d6162d18db2dffbe7f9dd9fcbce1c34... 1111testdata
```

## Database support

Currently only support for `Redis` is implemented (or rather, the `Redis protocol`). In order to be usable by the `datastore`, a database must implement the following operations according to the `redis protocol`:

- `PING`
- `HSET`
- `HGETALL`
- `HDEL`
- `SUBSCRIBE`
- `UNSUBSCRIBE`

## Managing data replication

By default, data is not replicated. In order to replicate data for a namespace, the `datastore` must be instructed to do so. As there are no http endpoints to manage this yet, the client can't do this. The only way is through the redis instance itself. When the `datastore` starts, it subsribes to the `replication` channel, and listens for messages. These messages have the following format: `command:namespace[:starttime]`. There are currently 2 commands: `subscribe` and `unsubscribe`. The starttime is a unix timestamp. It indicates that data stored in blocks before this time should be ignored (though the `UID` is still incremented). Starttime is optional, and has no effect when using the `unsubscribe` command. Additional fields (by adding additional colons) are ignored, but they don't cause the parsing of the command to fail. 

When the `subscribe` command is received, data replication starts for the namespace. All data from the start of the chain is replicated (unless a starttime is specified, in which replication can start at a certain point). When the replication reaches the point of the current consensus, data is replicated in real time as it is received by the consensus set. The `unsubscribe` command stops replication, but replicated data is not removed. Unless `unsubscribe` is used, replication settings for a namespace are persistent accross daemon restarts.

When using the `redis-cli`, subscribing to a namespace (for example namespace `1111` used earlier) is as simple as:

```
publish replication subscribe:1111
```

Which will start to replicate all data written to namespace `1111` since the start of the chain, as well as keeping track of new data written after the command was given. Should we only wish to replicate recent data, the following command can be used:

```
publish replication subscribe:1111:1521022000
```

In order to stop replicating the data, use the following command:

```
publish replication unsubscribe:1111
```

Bear in mind that already replicated data is not removed when this is done