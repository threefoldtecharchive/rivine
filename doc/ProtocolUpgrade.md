# Protocol Upgrade

Rivine is a Blockchain protocol, allowing you to implement your own `PoS` blockchain.

The protocol defines:

+ The JSON encoding of blocks, transactions (within blocks), outputs and inputs (within transactions) and any other JSON-only data;
+ The Binary encoding of blocks, transactions (within blocks), outputs and inputs (within transactions);
+ The P2P protocol (handshake and RPC methods) used for the communication between daemon's gateway in a Peer-to-Peer (P2P) fashion;
+ Consensus Rules;
+ Allowed transactions versions, as well as how each version is encoded and behaves;
+ Allowed coin/blockstake output types, as well as how each type is encoded and behaves;
+ Allowed coin/blockstake input types, as well as how each type is encoded and behaves;
+ Allowed Signature Algorithms and how it functions;

When a blockchain is launched, it does so with all this defined as the protocol.
It could be that the blockchain follows the Rivine Protocol as-is,
or it could be that the blockchain has specific modifications made to the protocol (e.g. more standard transaction versions).

## Version/Type Upgrades

Examples of version/type upgrades are:

+ Adding a new transaction version to a live blockchain;
+ Adding a new unlock condition type to a live blockchain;
+ Adding a new unlock fulfillment type to a live blockchain;
+ Adding a new signature algorithm to a live blockchain;
+ Increasing the range of possible fulfillments that can fulfill one or multiple unlock conditions in a live blockchain;

It should be noted that all these examples increase the set of possible transactions,
meaning all already created transactions will (and must) remain to be valid.

This is important because if your upgraded protocol would be only a subset of the previous protocol,
it would mean you might get a fork in the blockchain, to a recent or very early version of the chain, on a (much) lower block height.
This is something you want to avoid at all costs, which is why it is important that
you avoid a decrease of possible transaction varients in any protocol upgrade.

Any protocol upgrade must be done with great care:

+ Ensure the design is sound and has a long enough lifetime;
+ Ensure the implementation is well tested;
+ Ensure that all possible transactions from past protocol versions, and especially
  all transactions versions already created on your live blockchain, are still valid within the context of your new upgraded protocol;

Furthermore, you should ensure that you have the backing of sufficient block creators,
such that the majority of (active) block stakes will apply the protocol upgrade.
All nodes with the upgraded protocol will be on their own shorter/losing chain, should this not be the case,
from the moment that a feature newly added in this protocol upgrade is used in a transaction.

Blocks have no versions and remain static in format, no matter how many protocol upgrades,
their intenrals however might change, specifically increase the range of possibilities.

Forks will happen during a protocol upgrade, but if all has been prepared well,
your blockchain should settle to a single truth fairly soon, once again.
