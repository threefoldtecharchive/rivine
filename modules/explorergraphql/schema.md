# Schema Types

<details>
  <summary><strong>Table of Contents</strong></summary>

  * [Query](#query)
  * [Objects](#objects)
    * [AtomicSwapCondition](#atomicswapcondition)
    * [AtomicSwapContract](#atomicswapcontract)
    * [AtomicSwapFulfillment](#atomicswapfulfillment)
    * [AuthAddressUpdateTransaction](#authaddressupdatetransaction)
    * [AuthConditionUpdateTransaction](#authconditionupdatetransaction)
    * [Balance](#balance)
    * [Block](#block)
    * [BlockChainSnapshotFacts](#blockchainsnapshotfacts)
    * [BlockFacts](#blockfacts)
    * [BlockHeader](#blockheader)
    * [BlockPayout](#blockpayout)
    * [ChainAggregatedData](#chainaggregateddata)
    * [ChainConstants](#chainconstants)
    * [ChainFacts](#chainfacts)
    * [FreeForAllWallet](#freeforallwallet)
    * [Input](#input)
    * [LockTimeCondition](#locktimecondition)
    * [MintCoinCreationTransaction](#mintcoincreationtransaction)
    * [MintCoinDestructionTransaction](#mintcoindestructiontransaction)
    * [MintConditionDefinitionTransaction](#mintconditiondefinitiontransaction)
    * [MultiSignatureCondition](#multisignaturecondition)
    * [MultiSignatureFulfillment](#multisignaturefulfillment)
    * [MultiSignatureWallet](#multisignaturewallet)
    * [NilCondition](#nilcondition)
    * [Output](#output)
    * [PublicKeySignaturePair](#publickeysignaturepair)
    * [ResponseBlocks](#responseblocks)
    * [SingleSignatureFulfillment](#singlesignaturefulfillment)
    * [SingleSignatureWallet](#singlesignaturewallet)
    * [StandardTransaction](#standardtransaction)
    * [TransactionFeePayout](#transactionfeepayout)
    * [TransactionParentInfo](#transactionparentinfo)
    * [UnlockHashCondition](#unlockhashcondition)
    * [UnlockHashPublicKeyPair](#unlockhashpublickeypair)
  * [Inputs](#inputs)
    * [BigIntFilter](#bigintfilter)
    * [BinaryDataFilter](#binarydatafilter)
    * [BlockPositionOperators](#blockpositionoperators)
    * [BlockPositionRange](#blockpositionrange)
    * [BlocksFilter](#blocksfilter)
    * [IntFilter](#intfilter)
    * [TimestampOperators](#timestampoperators)
    * [TimestampRange](#timestamprange)
    * [TransactionsFilter](#transactionsfilter)
  * [Enums](#enums)
    * [BlockPayoutType](#blockpayouttype)
    * [LockType](#locktype)
    * [OutputType](#outputtype)
  * [Scalars](#scalars)
    * [BigInt](#bigint)
    * [BinaryData](#binarydata)
    * [BlockHeight](#blockheight)
    * [Boolean](#boolean)
    * [ByteVersion](#byteversion)
    * [Cursor](#cursor)
    * [Hash](#hash)
    * [Int](#int)
    * [LockTime](#locktime)
    * [ObjectID](#objectid)
    * [PublicKey](#publickey)
    * [Signature](#signature)
    * [String](#string)
    * [Timestamp](#timestamp)
    * [UnlockHash](#unlockhash)
  * [Interfaces](#interfaces)
    * [Transaction](#transaction)
    * [UnlockCondition](#unlockcondition)
    * [UnlockFulfillment](#unlockfulfillment)
    * [Wallet](#wallet)

</details>

## Query (QueryRoot)
<table>
<thead>
<tr>
<th align="left">Field</th>
<th align="right">Argument</th>
<th align="left">Type</th>
<th align="left">Description</th>
</tr>
</thead>
<tbody>
<tr>
<td colspan="2" valign="top"><strong>chain</strong></td>
<td valign="top"><a href="#chainfacts">ChainFacts</a></td>
<td>

Query the chain facts, for constant, aggregated and contemporary data.
Constant data allows you to learn more about the network configuration,
aggregated data allows you have a quick overview of the amount of coins
and blocks available (locked or unlocked), and the contemporary data
allows you to know the latest block the chain is on (and thus
also the chain time and height).

</td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>object</strong></td>
<td valign="top"><a href="#object">Object</a></td>
<td>

Query an object, which can be a wallet, contract, block, transaction or output.
If no identifier is given the latest block will be returned.

</td>
</tr>
<tr>
<td colspan="2" align="right" valign="top">id</td>
<td valign="top"><a href="#objectid">ObjectID</a></td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>block</strong></td>
<td valign="top"><a href="#block">Block</a></td>
<td>

Query a block by identifier.
If no Hash is given, the latest block will be
returned.

</td>
</tr>
<tr>
<td colspan="2" align="right" valign="top">id</td>
<td valign="top"><a href="#hash">Hash</a></td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>blockAt</strong></td>
<td valign="top"><a href="#block">Block</a></td>
<td>

Query a block by position.
If no position is given, the latest block will be returned.
Use a negative position to choose a block starting from the latest block,
for example -2 for the 2nd last block, and -1 for the last block.
The genesis block is at position 0 and the numbering goes upwards from there.

</td>
</tr>
<tr>
<td colspan="2" align="right" valign="top">position</td>
<td valign="top"><a href="#int">Int</a></td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>blocks</strong></td>
<td valign="top"><a href="#responseblocks">ResponseBlocks</a></td>
<td>

Query multiple blocks, optionally giving a filter.
If no filter is given the first blocks at the start are given.
This query uses pagination and will have a server-defined upper limit
of items to return maximum. In case more items can be returned,
follow-up call(s) have to be made, using the returned Cursor
as a FilterSinceCursor.

</td>
</tr>
<tr>
<td colspan="2" align="right" valign="top">filter</td>
<td valign="top"><a href="#blocksfilter">BlocksFilter</a></td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>transaction</strong></td>
<td valign="top"><a href="#transaction">Transaction</a></td>
<td>

Query a transaction by identifier.
For transaction version specific information the desired
types have to be checked using the union-on GraphQL selection in your query.
Please consult the Transaction interface implementations to know
which versions are available and how they are typed in this API.

</td>
</tr>
<tr>
<td colspan="2" align="right" valign="top">id</td>
<td valign="top"><a href="#hash">Hash</a>!</td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>output</strong></td>
<td valign="top"><a href="#output">Output</a></td>
<td>

Query an output by identifier.
An input can be queried as well by querying the output,
the input data will be available as spenditure data on the returned output.
Please consult the `OutputType` enum to know what Outputs are supported
by this API.

</td>
</tr>
<tr>
<td colspan="2" align="right" valign="top">id</td>
<td valign="top"><a href="#hash">Hash</a>!</td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>wallet</strong></td>
<td valign="top"><a href="#wallet">Wallet</a></td>
<td>

Query a wallet by its address (a more human friendly name for unlockhash).
Please consult the Wallet implementations t oknow
which wallet types are available in this API and how they are typed.

</td>
</tr>
<tr>
<td colspan="2" align="right" valign="top">unlockhash</td>
<td valign="top"><a href="#unlockhash">UnlockHash</a>!</td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>contract</strong></td>
<td valign="top"><a href="#contract">Contract</a></td>
<td>

Query a contract by its address (a more human friendly name for unlockhash).
Please consult the Contract union type to know what contracts are available
and how they are typed in this API.

</td>
</tr>
<tr>
<td colspan="2" align="right" valign="top">unlockhash</td>
<td valign="top"><a href="#unlockhash">UnlockHash</a>!</td>
<td></td>
</tr>
</tbody>
</table>

## Objects

### AtomicSwapCondition

<table>
<thead>
<tr>
<th align="left">Field</th>
<th align="right">Argument</th>
<th align="left">Type</th>
<th align="left">Description</th>
</tr>
</thead>
<tbody>
<tr>
<td colspan="2" valign="top"><strong>Version</strong></td>
<td valign="top"><a href="#byteversion">ByteVersion</a>!</td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>UnlockHash</strong></td>
<td valign="top"><a href="#unlockhash">UnlockHash</a>!</td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>Sender</strong></td>
<td valign="top"><a href="#unlockhashpublickeypair">UnlockHashPublicKeyPair</a>!</td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>Receiver</strong></td>
<td valign="top"><a href="#unlockhashpublickeypair">UnlockHashPublicKeyPair</a>!</td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>HashedSecret</strong></td>
<td valign="top"><a href="#binarydata">BinaryData</a>!</td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>TimeLock</strong></td>
<td valign="top"><a href="#locktime">LockTime</a>!</td>
<td></td>
</tr>
</tbody>
</table>

### AtomicSwapContract

A contract used by Rivine chains to start and complete an atomic swap.
See https://github.com/threefoldtech/rivine/blob/master/doc/atomicswap/atomicswap.md for more
information about atomic swaps.

In short it is in generally used to safely swap coins between two different blockchains,
without requiring trust in the other party.

<table>
<thead>
<tr>
<th align="left">Field</th>
<th align="right">Argument</th>
<th align="left">Type</th>
<th align="left">Description</th>
</tr>
</thead>
<tbody>
<tr>
<td colspan="2" valign="top"><strong>UnlockHash</strong></td>
<td valign="top"><a href="#unlockhash">UnlockHash</a>!</td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>ContractCondition</strong></td>
<td valign="top"><a href="#atomicswapcondition">AtomicSwapCondition</a>!</td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>ContractFulfillment</strong></td>
<td valign="top"><a href="#atomicswapfulfillment">AtomicSwapFulfillment</a></td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>ContractValue</strong></td>
<td valign="top"><a href="#bigint">BigInt</a>!</td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>Transactions</strong></td>
<td valign="top">[<a href="#transaction">Transaction</a>!]</td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>CoinOutput</strong></td>
<td valign="top"><a href="#output">Output</a></td>
<td></td>
</tr>
</tbody>
</table>

### AtomicSwapFulfillment

<table>
<thead>
<tr>
<th align="left">Field</th>
<th align="right">Argument</th>
<th align="left">Type</th>
<th align="left">Description</th>
</tr>
</thead>
<tbody>
<tr>
<td colspan="2" valign="top"><strong>Version</strong></td>
<td valign="top"><a href="#byteversion">ByteVersion</a>!</td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>ParentCondition</strong></td>
<td valign="top"><a href="#unlockcondition">UnlockCondition</a></td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>PublicKey</strong></td>
<td valign="top"><a href="#publickey">PublicKey</a>!</td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>Signature</strong></td>
<td valign="top"><a href="#signature">Signature</a>!</td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>Secret</strong></td>
<td valign="top"><a href="#binarydata">BinaryData</a></td>
<td></td>
</tr>
</tbody>
</table>

### AuthAddressUpdateTransaction

A transaction that allows the auth power to update
the authentication of an address.

<table>
<thead>
<tr>
<th align="left">Field</th>
<th align="right">Argument</th>
<th align="left">Type</th>
<th align="left">Description</th>
</tr>
</thead>
<tbody>
<tr>
<td colspan="2" valign="top"><strong>ID</strong></td>
<td valign="top"><a href="#hash">Hash</a>!</td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>Version</strong></td>
<td valign="top"><a href="#byteversion">ByteVersion</a>!</td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>ParentBlock</strong></td>
<td valign="top"><a href="#transactionparentinfo">TransactionParentInfo</a></td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>CoinInputs</strong></td>
<td valign="top">[<a href="#input">Input</a>!]</td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>CoinOutputs</strong></td>
<td valign="top">[<a href="#output">Output</a>!]</td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>Nonce</strong></td>
<td valign="top"><a href="#binarydata">BinaryData</a>!</td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>AuthAddresses</strong></td>
<td valign="top">[<a href="#unlockhashpublickeypair">UnlockHashPublicKeyPair</a>]</td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>DeauthAddresses</strong></td>
<td valign="top">[<a href="#unlockhashpublickeypair">UnlockHashPublicKeyPair</a>]</td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>AuthFulfillment</strong></td>
<td valign="top"><a href="#unlockfulfillment">UnlockFulfillment</a>!</td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>FeePayouts</strong></td>
<td valign="top">[<a href="#transactionfeepayout">TransactionFeePayout</a>!]</td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>ArbitraryData</strong></td>
<td valign="top"><a href="#binarydata">BinaryData</a></td>
<td></td>
</tr>
</tbody>
</table>

### AuthConditionUpdateTransaction

A transaction that allows the auth power to update
the condiiton, defining who is the auth power.

<table>
<thead>
<tr>
<th align="left">Field</th>
<th align="right">Argument</th>
<th align="left">Type</th>
<th align="left">Description</th>
</tr>
</thead>
<tbody>
<tr>
<td colspan="2" valign="top"><strong>ID</strong></td>
<td valign="top"><a href="#hash">Hash</a>!</td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>Version</strong></td>
<td valign="top"><a href="#byteversion">ByteVersion</a>!</td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>ParentBlock</strong></td>
<td valign="top"><a href="#transactionparentinfo">TransactionParentInfo</a></td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>CoinInputs</strong></td>
<td valign="top">[<a href="#input">Input</a>!]</td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>CoinOutputs</strong></td>
<td valign="top">[<a href="#output">Output</a>!]</td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>Nonce</strong></td>
<td valign="top"><a href="#binarydata">BinaryData</a>!</td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>AuthFulfillment</strong></td>
<td valign="top"><a href="#unlockfulfillment">UnlockFulfillment</a>!</td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>NewAuthCondition</strong></td>
<td valign="top"><a href="#unlockcondition">UnlockCondition</a>!</td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>FeePayouts</strong></td>
<td valign="top">[<a href="#transactionfeepayout">TransactionFeePayout</a>!]</td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>ArbitraryData</strong></td>
<td valign="top"><a href="#binarydata">BinaryData</a></td>
<td></td>
</tr>
</tbody>
</table>

### Balance

The balance contains aggregated asset values,
and is updated for each block that affect's a wallet's
coin or block stake balance.

<table>
<thead>
<tr>
<th align="left">Field</th>
<th align="right">Argument</th>
<th align="left">Type</th>
<th align="left">Description</th>
</tr>
</thead>
<tbody>
<tr>
<td colspan="2" valign="top"><strong>Unlocked</strong></td>
<td valign="top"><a href="#bigint">BigInt</a>!</td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>Locked</strong></td>
<td valign="top"><a href="#bigint">BigInt</a>!</td>
<td></td>
</tr>
</tbody>
</table>

### Block

The API of the block's data view.

<table>
<thead>
<tr>
<th align="left">Field</th>
<th align="right">Argument</th>
<th align="left">Type</th>
<th align="left">Description</th>
</tr>
</thead>
<tbody>
<tr>
<td colspan="2" valign="top"><strong>Header</strong></td>
<td valign="top"><a href="#blockheader">BlockHeader</a>!</td>
<td>

Data for this block,
such as its Identifier, height and timestamp.

</td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>Facts</strong></td>
<td valign="top"><a href="#blockfacts">BlockFacts</a></td>
<td>

Facts for this block such as its difficulty and target used,
but also a snapshot of the aggregated chain data at the
chain state on this block.

Queried in a lazy manner, fetching only as much data as
required for the query.

</td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>Transactions</strong></td>
<td valign="top">[<a href="#transaction">Transaction</a>!]</td>
<td>

The transactions part of this block.

Queried in a lazy manner, fetching only as much data as
required for the query.

</td>
</tr>
<tr>
<td colspan="2" align="right" valign="top">filter</td>
<td valign="top"><a href="#transactionsfilter">TransactionsFilter</a></td>
<td></td>
</tr>
</tbody>
</table>

### BlockChainSnapshotFacts

The API of the chainshots facts collected for a block.

<table>
<thead>
<tr>
<th align="left">Field</th>
<th align="right">Argument</th>
<th align="left">Type</th>
<th align="left">Description</th>
</tr>
</thead>
<tbody>
<tr>
<td colspan="2" valign="top"><strong>TotalCoins</strong></td>
<td valign="top"><a href="#bigint">BigInt</a></td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>TotalLockedCoins</strong></td>
<td valign="top"><a href="#bigint">BigInt</a></td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>TotalBlockStakes</strong></td>
<td valign="top"><a href="#bigint">BigInt</a></td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>TotalLockedBlockStakes</strong></td>
<td valign="top"><a href="#bigint">BigInt</a></td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>EstimatedActiveBlockStakes</strong></td>
<td valign="top"><a href="#bigint">BigInt</a></td>
<td></td>
</tr>
</tbody>
</table>

### BlockFacts

The API of the facts collected for a block.

<table>
<thead>
<tr>
<th align="left">Field</th>
<th align="right">Argument</th>
<th align="left">Type</th>
<th align="left">Description</th>
</tr>
</thead>
<tbody>
<tr>
<td colspan="2" valign="top"><strong>Difficulty</strong></td>
<td valign="top"><a href="#bigint">BigInt</a></td>
<td>

The difficulty used, in the consensus algorithm,
for creating this block.

</td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>Target</strong></td>
<td valign="top"><a href="#hash">Hash</a></td>
<td>

The target hash used, in the consensus algorithm,
for creating this block.

</td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>ChainSnapshot</strong></td>
<td valign="top"><a href="#blockchainsnapshotfacts">BlockChainSnapshotFacts</a></td>
<td>

The aggregated chain data as a snapshot taken,
after this fact's block was applied.

</td>
</tr>
</tbody>
</table>

### BlockHeader

The API for the block-specific "header" data of a block.
Containing information such as the ID, the ID of its parent block,
block time and height as well as (miner) payout information.
The Parent and Child block can be queried recursively as well.

<table>
<thead>
<tr>
<th align="left">Field</th>
<th align="right">Argument</th>
<th align="left">Type</th>
<th align="left">Description</th>
</tr>
</thead>
<tbody>
<tr>
<td colspan="2" valign="top"><strong>ID</strong></td>
<td valign="top"><a href="#hash">Hash</a>!</td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>ParentID</strong></td>
<td valign="top"><a href="#hash">Hash</a></td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>Parent</strong></td>
<td valign="top"><a href="#block">Block</a></td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>Child</strong></td>
<td valign="top"><a href="#block">Block</a></td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>BlockTime</strong></td>
<td valign="top"><a href="#timestamp">Timestamp</a></td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>BlockHeight</strong></td>
<td valign="top"><a href="#blockheight">BlockHeight</a></td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>Payouts</strong></td>
<td valign="top">[<a href="#blockpayout">BlockPayout</a>!]</td>
<td></td>
</tr>
<tr>
<td colspan="2" align="right" valign="top">type</td>
<td valign="top"><a href="#blockpayouttype">BlockPayoutType</a></td>
<td></td>
</tr>
</tbody>
</table>

### BlockPayout

<table>
<thead>
<tr>
<th align="left">Field</th>
<th align="right">Argument</th>
<th align="left">Type</th>
<th align="left">Description</th>
</tr>
</thead>
<tbody>
<tr>
<td colspan="2" valign="top"><strong>Output</strong></td>
<td valign="top"><a href="#output">Output</a>!</td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>Type</strong></td>
<td valign="top"><a href="#blockpayouttype">BlockPayoutType</a></td>
<td></td>
</tr>
</tbody>
</table>

### ChainAggregatedData

The aggregated chain data,
updated for every block is that is applied and reverted.

<table>
<thead>
<tr>
<th align="left">Field</th>
<th align="right">Argument</th>
<th align="left">Type</th>
<th align="left">Description</th>
</tr>
</thead>
<tbody>
<tr>
<td colspan="2" valign="top"><strong>TotalCoins</strong></td>
<td valign="top"><a href="#bigint">BigInt</a></td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>TotalLockedCoins</strong></td>
<td valign="top"><a href="#bigint">BigInt</a></td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>TotalBlockStakes</strong></td>
<td valign="top"><a href="#bigint">BigInt</a></td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>TotalLockedBlockStakes</strong></td>
<td valign="top"><a href="#bigint">BigInt</a></td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>EstimatedActiveBlockStakes</strong></td>
<td valign="top"><a href="#bigint">BigInt</a></td>
<td></td>
</tr>
</tbody>
</table>

### ChainConstants

ChainConstants collect all constant information known about
a chain network and exposed via this API.

<table>
<thead>
<tr>
<th align="left">Field</th>
<th align="right">Argument</th>
<th align="left">Type</th>
<th align="left">Description</th>
</tr>
</thead>
<tbody>
<tr>
<td colspan="2" valign="top"><strong>Name</strong></td>
<td valign="top"><a href="#string">String</a>!</td>
<td>

The name of the chain that this explorer is connected to.

</td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>NetworkName</strong></td>
<td valign="top"><a href="#string">String</a>!</td>
<td>

The name of the network,
usually one of `"standard"`, `"testnet"`, "`devnet"`.
The name of a network is not restricted to these values however.

</td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>CoinUnit</strong></td>
<td valign="top"><a href="#string">String</a>!</td>
<td>

The unit of coins.
For the Threefold Chain this is for example `"TFT"`.

</td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>CoinPecision</strong></td>
<td valign="top"><a href="#int">Int</a>!</td>
<td>

The amount of decimals that the coins can be expressed in.
The coin values are always exposed as the lowered unit,
see the `BigInt` type for more information about the encoding
within the context of this API.

If for example the CoinPrecision is `2`,
than a currency value of `"104"` is actually `1.04`.

</td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>ChainVersion</strong></td>
<td valign="top"><a href="#string">String</a>!</td>
<td>

The source code version this daemon is compiled on.

</td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>DefaultTransactionVersion</strong></td>
<td valign="top"><a href="#byteversion">ByteVersion</a>!</td>
<td>

The transaction version that clients should use as the default
transaction version.

</td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>GatewayProtocolVersion</strong></td>
<td valign="top"><a href="#string">String</a>!</td>
<td>

The gateway Protocol Version used by this daemon's gateway module.

</td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>ConsensusPlugins</strong></td>
<td valign="top">[<a href="#string">String</a>!]</td>
<td>

ConsensusPlugins provide you with the names of all plugins used by
the consensus of this network's daemons and thus allows you to know
what extra features might be available for this network.

</td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>GenesisTimestamp</strong></td>
<td valign="top"><a href="#timestamp">Timestamp</a>!</td>
<td>

The (Unix Epoch, seconds) timestamp of the first block,
the so called genesis block.

</td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>BlockSizeLimitInBytes</strong></td>
<td valign="top"><a href="#int">Int</a>!</td>
<td>

Defines the maximum size a block is allowed to be, in bytes.

</td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>AverageBlockCreationTimeInSeconds</strong></td>
<td valign="top"><a href="#int">Int</a>!</td>
<td>

The average block creation time in seconds the consensus algorithm
aims to achieve. It does not mean that it will take exatly this amount of seconds for a
new block to be created, nor it is an upper limit. You will however notice that the average
block creation time of a sufficient amount of sequential blocks does come close
to this average block creation time.

</td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>GenesisTotalBlockStakes</strong></td>
<td valign="top"><a href="#bigint">BigInt</a>!</td>
<td>

The total amount of block stakes available at the creation of this blockchain.
As blockchains can currently not create new block stakes it is also the final amount of block stakes on the chain.

</td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>BlockStakeAging</strong></td>
<td valign="top"><a href="#int">Int</a>!</td>
<td>

Defines how many blocks a block stake have to age
prior to being able to use block stakes for creating blocks.
The age is calculated by computing the height when the stakes were
transfered until the current block height.
When transfering stakes to yourself as part of a block creation,
the constant required aging concept (using the amount as defined here) does not apply.

</td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>BlockCreatorFee</strong></td>
<td valign="top"><a href="#bigint">BigInt</a></td>
<td>

The fee that a block creator recieves for the creation of a block.
Can be null in case the chain does not award fees for such creations,
a possibility for private chains where all nodes are owned by one organisation.

</td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>MinimumTransactionFee</strong></td>
<td valign="top"><a href="#bigint">BigInt</a></td>
<td>

The minimum fee that has to be spent by a wallet in order to make a coin or block stake transfer.
The fee does not apply for block creation transactions.
Can be null in case the network does not require transaction fees.

</td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>TransactionFeeBeneficiary</strong></td>
<td valign="top"><a href="#unlockcondition">UnlockCondition</a></td>
<td>

Some networks collect all transaction fees in a single wallet,
if this is the case it will be available as condition in this field, for query purposes.

</td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>PayoutMaturityDelay</strong></td>
<td valign="top"><a href="#blockheight">BlockHeight</a>!</td>
<td>

This delay, in block amount, defines how long a miner payout (e.g. block creator or transaction fee)
is locked prior to being spendable.

</td>
</tr>
</tbody>
</table>

### ChainFacts

ChainFacts collects facts about the queried chain.

<table>
<thead>
<tr>
<th align="left">Field</th>
<th align="right">Argument</th>
<th align="left">Type</th>
<th align="left">Description</th>
</tr>
</thead>
<tbody>
<tr>
<td colspan="2" valign="top"><strong>Constants</strong></td>
<td valign="top"><a href="#chainconstants">ChainConstants</a>!</td>
<td>

Constants collects all constant (static) data known about the chain,
and is provided by the daemon network configuration.

</td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>LastBlock</strong></td>
<td valign="top"><a href="#block">Block</a>!</td>
<td>

LastBlock allows you to look up the last block,
saving you a second query, in case you need it, to look up that block,
even if it is just the height or timestamp.

</td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>Aggregated</strong></td>
<td valign="top"><a href="#chainaggregateddata">ChainAggregatedData</a></td>
<td>

Contains all aggregated global data,
updated for this chain for every applied and reverted block.

</td>
</tr>
</tbody>
</table>

### FreeForAllWallet

<table>
<thead>
<tr>
<th align="left">Field</th>
<th align="right">Argument</th>
<th align="left">Type</th>
<th align="left">Description</th>
</tr>
</thead>
<tbody>
<tr>
<td colspan="2" valign="top"><strong>UnlockHash</strong></td>
<td valign="top"><a href="#unlockhash">UnlockHash</a>!</td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>CoinBalance</strong></td>
<td valign="top"><a href="#balance">Balance</a></td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>BlockStakeBalance</strong></td>
<td valign="top"><a href="#balance">Balance</a></td>
<td></td>
</tr>
</tbody>
</table>

### Input

Specific Output information for an `Output` used as `Input`
in a transaction. Within the context of this API
is this type only used for Transactions.

When looking up an `Output` it is always returned as the `Output`
type, optionally containing `ChildInput` in case it was also
used as an `Input` in some transaction already.

<table>
<thead>
<tr>
<th align="left">Field</th>
<th align="right">Argument</th>
<th align="left">Type</th>
<th align="left">Description</th>
</tr>
</thead>
<tbody>
<tr>
<td colspan="2" valign="top"><strong>ID</strong></td>
<td valign="top"><a href="#hash">Hash</a>!</td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>Type</strong></td>
<td valign="top"><a href="#outputtype">OutputType</a></td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>Value</strong></td>
<td valign="top"><a href="#bigint">BigInt</a>!</td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>Fulfillment</strong></td>
<td valign="top"><a href="#unlockfulfillment">UnlockFulfillment</a>!</td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>ParentOutput</strong></td>
<td valign="top"><a href="#output">Output</a></td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>ParentTransaction</strong></td>
<td valign="top"><a href="#transaction">Transaction</a></td>
<td></td>
</tr>
</tbody>
</table>

### LockTimeCondition

<table>
<thead>
<tr>
<th align="left">Field</th>
<th align="right">Argument</th>
<th align="left">Type</th>
<th align="left">Description</th>
</tr>
</thead>
<tbody>
<tr>
<td colspan="2" valign="top"><strong>Version</strong></td>
<td valign="top"><a href="#byteversion">ByteVersion</a>!</td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>UnlockHash</strong></td>
<td valign="top"><a href="#unlockhash">UnlockHash</a></td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>LockValue</strong></td>
<td valign="top"><a href="#locktime">LockTime</a>!</td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>LockType</strong></td>
<td valign="top"><a href="#locktype">LockType</a>!</td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>Condition</strong></td>
<td valign="top"><a href="#unlockcondition">UnlockCondition</a>!</td>
<td></td>
</tr>
</tbody>
</table>

### MintCoinCreationTransaction

The transaction used to mint tokens,
a transaction that can only be done by "who"
owns the currently active Minter condition.

See `MintConditionDefinitionTransaction` for more information.

<table>
<thead>
<tr>
<th align="left">Field</th>
<th align="right">Argument</th>
<th align="left">Type</th>
<th align="left">Description</th>
</tr>
</thead>
<tbody>
<tr>
<td colspan="2" valign="top"><strong>ID</strong></td>
<td valign="top"><a href="#hash">Hash</a>!</td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>Version</strong></td>
<td valign="top"><a href="#byteversion">ByteVersion</a>!</td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>ParentBlock</strong></td>
<td valign="top"><a href="#transactionparentinfo">TransactionParentInfo</a></td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>Nonce</strong></td>
<td valign="top"><a href="#binarydata">BinaryData</a>!</td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>MintFulfillment</strong></td>
<td valign="top"><a href="#unlockfulfillment">UnlockFulfillment</a>!</td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>CoinInputs</strong></td>
<td valign="top">[<a href="#input">Input</a>!]</td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>CoinOutputs</strong></td>
<td valign="top">[<a href="#output">Output</a>!]!</td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>FeePayouts</strong></td>
<td valign="top">[<a href="#transactionfeepayout">TransactionFeePayout</a>!]</td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>ArbitraryData</strong></td>
<td valign="top"><a href="#binarydata">BinaryData</a></td>
<td></td>
</tr>
</tbody>
</table>

### MintCoinDestructionTransaction

A transaction that allows you to burn coins,
meaning that the value of the spent coin input(s)
are (partly) sent to no one. Or in other words,
the combined value sum of the fee payouts and coin outputs
will be smaller than the value sum of coin inputs.

<table>
<thead>
<tr>
<th align="left">Field</th>
<th align="right">Argument</th>
<th align="left">Type</th>
<th align="left">Description</th>
</tr>
</thead>
<tbody>
<tr>
<td colspan="2" valign="top"><strong>ID</strong></td>
<td valign="top"><a href="#hash">Hash</a>!</td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>Version</strong></td>
<td valign="top"><a href="#byteversion">ByteVersion</a>!</td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>ParentBlock</strong></td>
<td valign="top"><a href="#transactionparentinfo">TransactionParentInfo</a></td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>CoinInputs</strong></td>
<td valign="top">[<a href="#input">Input</a>!]!</td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>CoinOutputs</strong></td>
<td valign="top">[<a href="#output">Output</a>!]</td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>FeePayouts</strong></td>
<td valign="top">[<a href="#transactionfeepayout">TransactionFeePayout</a>!]</td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>ArbitraryData</strong></td>
<td valign="top"><a href="#binarydata">BinaryData</a></td>
<td></td>
</tr>
</tbody>
</table>

### MintConditionDefinitionTransaction

The transaction used to redefine the Minter condition,
defining "who" can mint (= create) new coins,
as well as who can redefine the Minter condition once this
transaction is applied.

As long as no MintConditionDefinitionTransaction has been created
the condition as defined in the network configuration,
the so called genesis Mint Condition, is used.

The currently active Minter condition (or the one active at a certain height)
cannot yet be queried with the GraphQL API.

<table>
<thead>
<tr>
<th align="left">Field</th>
<th align="right">Argument</th>
<th align="left">Type</th>
<th align="left">Description</th>
</tr>
</thead>
<tbody>
<tr>
<td colspan="2" valign="top"><strong>ID</strong></td>
<td valign="top"><a href="#hash">Hash</a>!</td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>Version</strong></td>
<td valign="top"><a href="#byteversion">ByteVersion</a>!</td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>ParentBlock</strong></td>
<td valign="top"><a href="#transactionparentinfo">TransactionParentInfo</a></td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>Nonce</strong></td>
<td valign="top"><a href="#binarydata">BinaryData</a>!</td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>MintFulfillment</strong></td>
<td valign="top"><a href="#unlockfulfillment">UnlockFulfillment</a>!</td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>NewMintCondition</strong></td>
<td valign="top"><a href="#unlockcondition">UnlockCondition</a>!</td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>CoinInputs</strong></td>
<td valign="top">[<a href="#input">Input</a>!]</td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>CoinOutputs</strong></td>
<td valign="top">[<a href="#output">Output</a>!]</td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>FeePayouts</strong></td>
<td valign="top">[<a href="#transactionfeepayout">TransactionFeePayout</a>!]</td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>ArbitraryData</strong></td>
<td valign="top"><a href="#binarydata">BinaryData</a></td>
<td></td>
</tr>
</tbody>
</table>

### MultiSignatureCondition

<table>
<thead>
<tr>
<th align="left">Field</th>
<th align="right">Argument</th>
<th align="left">Type</th>
<th align="left">Description</th>
</tr>
</thead>
<tbody>
<tr>
<td colspan="2" valign="top"><strong>Version</strong></td>
<td valign="top"><a href="#byteversion">ByteVersion</a>!</td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>UnlockHash</strong></td>
<td valign="top"><a href="#unlockhash">UnlockHash</a>!</td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>Owners</strong></td>
<td valign="top">[<a href="#unlockhashpublickeypair">UnlockHashPublicKeyPair</a>]!</td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>RequiredSignatureCount</strong></td>
<td valign="top"><a href="#int">Int</a>!</td>
<td></td>
</tr>
</tbody>
</table>

### MultiSignatureFulfillment

<table>
<thead>
<tr>
<th align="left">Field</th>
<th align="right">Argument</th>
<th align="left">Type</th>
<th align="left">Description</th>
</tr>
</thead>
<tbody>
<tr>
<td colspan="2" valign="top"><strong>Version</strong></td>
<td valign="top"><a href="#byteversion">ByteVersion</a>!</td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>ParentCondition</strong></td>
<td valign="top"><a href="#unlockcondition">UnlockCondition</a></td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>Pairs</strong></td>
<td valign="top">[<a href="#publickeysignaturepair">PublicKeySignaturePair</a>!]!</td>
<td></td>
</tr>
</tbody>
</table>

### MultiSignatureWallet

<table>
<thead>
<tr>
<th align="left">Field</th>
<th align="right">Argument</th>
<th align="left">Type</th>
<th align="left">Description</th>
</tr>
</thead>
<tbody>
<tr>
<td colspan="2" valign="top"><strong>UnlockHash</strong></td>
<td valign="top"><a href="#unlockhash">UnlockHash</a>!</td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>CoinBalance</strong></td>
<td valign="top"><a href="#balance">Balance</a></td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>BlockStakeBalance</strong></td>
<td valign="top"><a href="#balance">Balance</a></td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>Owners</strong></td>
<td valign="top">[<a href="#unlockhashpublickeypair">UnlockHashPublicKeyPair</a>!]</td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>RequiredSignatureCount</strong></td>
<td valign="top"><a href="#int">Int</a></td>
<td></td>
</tr>
</tbody>
</table>

### NilCondition

<table>
<thead>
<tr>
<th align="left">Field</th>
<th align="right">Argument</th>
<th align="left">Type</th>
<th align="left">Description</th>
</tr>
</thead>
<tbody>
<tr>
<td colspan="2" valign="top"><strong>Version</strong></td>
<td valign="top"><a href="#byteversion">ByteVersion</a>!</td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>UnlockHash</strong></td>
<td valign="top"><a href="#unlockhash">UnlockHash</a>!</td>
<td></td>
</tr>
</tbody>
</table>

### Output

The API for an output, as queried directly or as a (sub) field of another query.
The `ChildInput` can be queried to know when (see the `Parent`) it was spent,
as well as who spent it (see the `Fulfillment` of the `ChildInput`).

<table>
<thead>
<tr>
<th align="left">Field</th>
<th align="right">Argument</th>
<th align="left">Type</th>
<th align="left">Description</th>
</tr>
</thead>
<tbody>
<tr>
<td colspan="2" valign="top"><strong>ID</strong></td>
<td valign="top"><a href="#hash">Hash</a>!</td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>Type</strong></td>
<td valign="top"><a href="#outputtype">OutputType</a></td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>Value</strong></td>
<td valign="top"><a href="#bigint">BigInt</a>!</td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>Condition</strong></td>
<td valign="top"><a href="#unlockcondition">UnlockCondition</a>!</td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>ChildInput</strong></td>
<td valign="top"><a href="#input">Input</a></td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>ParentID</strong></td>
<td valign="top"><a href="#hash">Hash</a>!</td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>Parent</strong></td>
<td valign="top"><a href="#outputparent">OutputParent</a>!</td>
<td></td>
</tr>
</tbody>
</table>

### PublicKeySignaturePair

<table>
<thead>
<tr>
<th align="left">Field</th>
<th align="right">Argument</th>
<th align="left">Type</th>
<th align="left">Description</th>
</tr>
</thead>
<tbody>
<tr>
<td colspan="2" valign="top"><strong>PublicKey</strong></td>
<td valign="top"><a href="#publickey">PublicKey</a>!</td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>Signature</strong></td>
<td valign="top"><a href="#signature">Signature</a>!</td>
<td></td>
</tr>
</tbody>
</table>

### ResponseBlocks

Response type for the blocks query.

<table>
<thead>
<tr>
<th align="left">Field</th>
<th align="right">Argument</th>
<th align="left">Type</th>
<th align="left">Description</th>
</tr>
</thead>
<tbody>
<tr>
<td colspan="2" valign="top"><strong>Blocks</strong></td>
<td valign="top">[<a href="#block">Block</a>!]</td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>NextCursor</strong></td>
<td valign="top"><a href="#cursor">Cursor</a></td>
<td>

In case all items could not be returned within a single call,
this cursor can be used for a follow-up blocks query call.

</td>
</tr>
</tbody>
</table>

### SingleSignatureFulfillment

<table>
<thead>
<tr>
<th align="left">Field</th>
<th align="right">Argument</th>
<th align="left">Type</th>
<th align="left">Description</th>
</tr>
</thead>
<tbody>
<tr>
<td colspan="2" valign="top"><strong>Version</strong></td>
<td valign="top"><a href="#byteversion">ByteVersion</a>!</td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>ParentCondition</strong></td>
<td valign="top"><a href="#unlockcondition">UnlockCondition</a></td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>PublicKey</strong></td>
<td valign="top"><a href="#publickey">PublicKey</a>!</td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>Signature</strong></td>
<td valign="top"><a href="#signature">Signature</a>!</td>
<td></td>
</tr>
</tbody>
</table>

### SingleSignatureWallet

<table>
<thead>
<tr>
<th align="left">Field</th>
<th align="right">Argument</th>
<th align="left">Type</th>
<th align="left">Description</th>
</tr>
</thead>
<tbody>
<tr>
<td colspan="2" valign="top"><strong>UnlockHash</strong></td>
<td valign="top"><a href="#unlockhash">UnlockHash</a>!</td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>CoinBalance</strong></td>
<td valign="top"><a href="#balance">Balance</a></td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>BlockStakeBalance</strong></td>
<td valign="top"><a href="#balance">Balance</a></td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>PublicKey</strong></td>
<td valign="top"><a href="#publickey">PublicKey</a></td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>MultiSignatureWallets</strong></td>
<td valign="top">[<a href="#multisignaturewallet">MultiSignatureWallet</a>!]</td>
<td></td>
</tr>
</tbody>
</table>

### StandardTransaction

The standard transaction used for all regular transactions,
most commonly used for coin transfers and block creation transactions.

<table>
<thead>
<tr>
<th align="left">Field</th>
<th align="right">Argument</th>
<th align="left">Type</th>
<th align="left">Description</th>
</tr>
</thead>
<tbody>
<tr>
<td colspan="2" valign="top"><strong>ID</strong></td>
<td valign="top"><a href="#hash">Hash</a>!</td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>Version</strong></td>
<td valign="top"><a href="#byteversion">ByteVersion</a>!</td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>ParentBlock</strong></td>
<td valign="top"><a href="#transactionparentinfo">TransactionParentInfo</a></td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>CoinInputs</strong></td>
<td valign="top">[<a href="#input">Input</a>!]</td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>CoinOutputs</strong></td>
<td valign="top">[<a href="#output">Output</a>!]</td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>BlockStakeInputs</strong></td>
<td valign="top">[<a href="#input">Input</a>!]</td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>BlockStakeOutputs</strong></td>
<td valign="top">[<a href="#output">Output</a>!]</td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>FeePayouts</strong></td>
<td valign="top">[<a href="#transactionfeepayout">TransactionFeePayout</a>!]</td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>ArbitraryData</strong></td>
<td valign="top"><a href="#binarydata">BinaryData</a></td>
<td></td>
</tr>
</tbody>
</table>

### TransactionFeePayout

<table>
<thead>
<tr>
<th align="left">Field</th>
<th align="right">Argument</th>
<th align="left">Type</th>
<th align="left">Description</th>
</tr>
</thead>
<tbody>
<tr>
<td colspan="2" valign="top"><strong>BlockPayout</strong></td>
<td valign="top"><a href="#blockpayout">BlockPayout</a></td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>Value</strong></td>
<td valign="top"><a href="#bigint">BigInt</a>!</td>
<td></td>
</tr>
</tbody>
</table>

### TransactionParentInfo

The information about the created block that contains this transaction.

<table>
<thead>
<tr>
<th align="left">Field</th>
<th align="right">Argument</th>
<th align="left">Type</th>
<th align="left">Description</th>
</tr>
</thead>
<tbody>
<tr>
<td colspan="2" valign="top"><strong>ID</strong></td>
<td valign="top"><a href="#hash">Hash</a>!</td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>ParentID</strong></td>
<td valign="top"><a href="#hash">Hash</a></td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>Height</strong></td>
<td valign="top"><a href="#blockheight">BlockHeight</a></td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>Timestamp</strong></td>
<td valign="top"><a href="#timestamp">Timestamp</a></td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>TransactionOrder</strong></td>
<td valign="top"><a href="#int">Int</a></td>
<td>

The static order (index) of this transaction
as defined by the (parent) block that contains this transaction.

</td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>SiblingTransactions</strong></td>
<td valign="top">[<a href="#transaction">Transaction</a>!]</td>
<td>

All transactions found in the (parent) block,
excluding this transaction.

</td>
</tr>
<tr>
<td colspan="2" align="right" valign="top">filter</td>
<td valign="top"><a href="#transactionsfilter">TransactionsFilter</a></td>
<td></td>
</tr>
</tbody>
</table>

### UnlockHashCondition

<table>
<thead>
<tr>
<th align="left">Field</th>
<th align="right">Argument</th>
<th align="left">Type</th>
<th align="left">Description</th>
</tr>
</thead>
<tbody>
<tr>
<td colspan="2" valign="top"><strong>Version</strong></td>
<td valign="top"><a href="#byteversion">ByteVersion</a>!</td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>UnlockHash</strong></td>
<td valign="top"><a href="#unlockhash">UnlockHash</a>!</td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>PublicKey</strong></td>
<td valign="top"><a href="#publickey">PublicKey</a></td>
<td></td>
</tr>
</tbody>
</table>

### UnlockHashPublicKeyPair

Each `01` prefixed `UnlockHash` (wallet address) is linked to a `PublicKey`.
If it is known, and thus exposed on the chain at some point,
it will be stored, and can be queried using the `PublicKey` field.
That field will be `null` in case the `PublicKey` is not (yet) known.

<table>
<thead>
<tr>
<th align="left">Field</th>
<th align="right">Argument</th>
<th align="left">Type</th>
<th align="left">Description</th>
</tr>
</thead>
<tbody>
<tr>
<td colspan="2" valign="top"><strong>UnlockHash</strong></td>
<td valign="top"><a href="#unlockhash">UnlockHash</a>!</td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>PublicKey</strong></td>
<td valign="top"><a href="#publickey">PublicKey</a></td>
<td></td>
</tr>
</tbody>
</table>

## Inputs

### BigIntFilter

Filter based on a big integer based on one of these options.

NOTE that these options should really be a Union, not an input composition type.
Once the RFC https://github.com/graphql/graphql-spec/blob/master/rfcs/InputUnion.md
is accepted and implemented by the implementations (including the one used by us),
we could use it here.

<table>
<thead>
<tr>
<th colspan="2" align="left">Field</th>
<th align="left">Type</th>
<th align="left">Description</th>
</tr>
</thead>
<tbody>
<tr>
<td colspan="2" valign="top"><strong>LessThan</strong></td>
<td valign="top"><a href="#bigint">BigInt</a></td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>LessThanOrEqualTo</strong></td>
<td valign="top"><a href="#bigint">BigInt</a></td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>EqualTo</strong></td>
<td valign="top"><a href="#bigint">BigInt</a></td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>GreaterThanOrEqualTo</strong></td>
<td valign="top"><a href="#bigint">BigInt</a></td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>GreaterThan</strong></td>
<td valign="top"><a href="#bigint">BigInt</a></td>
<td></td>
</tr>
</tbody>
</table>

### BinaryDataFilter

Filter based on binary data based on one of these options.

NOTE that these options should really be a Union, not an input composition type.
Once the RFC https://github.com/graphql/graphql-spec/blob/master/rfcs/InputUnion.md
is accepted and implemented by the implementations (including the one used by us),
we could use it here.

<table>
<thead>
<tr>
<th colspan="2" align="left">Field</th>
<th align="left">Type</th>
<th align="left">Description</th>
</tr>
</thead>
<tbody>
<tr>
<td colspan="2" valign="top"><strong>StartsWith</strong></td>
<td valign="top"><a href="#binarydata">BinaryData</a></td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>Contains</strong></td>
<td valign="top"><a href="#binarydata">BinaryData</a></td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>EndsWith</strong></td>
<td valign="top"><a href="#binarydata">BinaryData</a></td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>EqualTo</strong></td>
<td valign="top"><a href="#binarydata">BinaryData</a></td>
<td></td>
</tr>
</tbody>
</table>

### BlockPositionOperators

A poor man's input Union, allowing you to filter on block positions,
by defining what the upper- or lower limit is.
Or the inclusive range of blocks including and between (one or) two positions.
If no fields are given it is seen as a nil operator. No more then one field can be defined.

This really should be a Union, this is however not (yet) supported by the official GraphQL
specification. We probably will have to break this API later, as there is an active RFC working on supporting use cases like this.

<table>
<thead>
<tr>
<th colspan="2" align="left">Field</th>
<th align="left">Type</th>
<th align="left">Description</th>
</tr>
</thead>
<tbody>
<tr>
<td colspan="2" valign="top"><strong>Before</strong></td>
<td valign="top"><a href="#int">Int</a></td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>After</strong></td>
<td valign="top"><a href="#int">Int</a></td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>Between</strong></td>
<td valign="top"><a href="#blockpositionrange">BlockPositionRange</a></td>
<td></td>
</tr>
</tbody>
</table>

### BlockPositionRange

An inclusive range of block positions, with one or two points to be defined.
A range with no fields defined is equal to a nil range.

<table>
<thead>
<tr>
<th colspan="2" align="left">Field</th>
<th align="left">Type</th>
<th align="left">Description</th>
</tr>
</thead>
<tbody>
<tr>
<td colspan="2" valign="top"><strong>Start</strong></td>
<td valign="top"><a href="#int">Int</a></td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>End</strong></td>
<td valign="top"><a href="#int">Int</a></td>
<td></td>
</tr>
</tbody>
</table>

### BlocksFilter

All possible filters that can be used to query for a list of blocks.
Multiple filters can be combined. It is also valid that none are given.

<table>
<thead>
<tr>
<th colspan="2" align="left">Field</th>
<th align="left">Type</th>
<th align="left">Description</th>
</tr>
</thead>
<tbody>
<tr>
<td colspan="2" valign="top"><strong>Height</strong></td>
<td valign="top"><a href="#blockpositionoperators">BlockPositionOperators</a></td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>Timestamp</strong></td>
<td valign="top"><a href="#timestampoperators">TimestampOperators</a></td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>TransactionLength</strong></td>
<td valign="top"><a href="#intfilter">IntFilter</a></td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>Limit</strong></td>
<td valign="top"><a href="#int">Int</a></td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>Cursor</strong></td>
<td valign="top"><a href="#cursor">Cursor</a></td>
<td>

A cursor that allows the blocks query to pick up from a state previously left off.
When this cursor is defined, you should define the same filters as used last time,
even though this is not enforced. The Limit filter is an exception to this.

</td>
</tr>
</tbody>
</table>

### IntFilter

Filter based on an integer based on one of these options.

NOTE that these options should really be a Union, not an input composition type.
Once the RFC https://github.com/graphql/graphql-spec/blob/master/rfcs/InputUnion.md
is accepted and implemented by the implementations (including the one used by us),
we could use it here.

<table>
<thead>
<tr>
<th colspan="2" align="left">Field</th>
<th align="left">Type</th>
<th align="left">Description</th>
</tr>
</thead>
<tbody>
<tr>
<td colspan="2" valign="top"><strong>LessThan</strong></td>
<td valign="top"><a href="#int">Int</a></td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>LessThanOrEqualTo</strong></td>
<td valign="top"><a href="#int">Int</a></td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>EqualTo</strong></td>
<td valign="top"><a href="#int">Int</a></td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>GreaterThanOrEqualTo</strong></td>
<td valign="top"><a href="#int">Int</a></td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>GreaterThan</strong></td>
<td valign="top"><a href="#int">Int</a></td>
<td></td>
</tr>
</tbody>
</table>

### TimestampOperators

A poor man's input Union, allowing you to filter on timestamps,
by defining what the upper- or lower limit is.
Or the inclusive range of blocks including and between (one or) two timestamps.
If no fields are given it is seen as a nil operator. No more then one field can be defined.

This really should be a Union, this is however not (yet) supported by the official GraphQL
specification. We probably will have to break this API later, as there is an active RFC working on supporting use cases like this.

<table>
<thead>
<tr>
<th colspan="2" align="left">Field</th>
<th align="left">Type</th>
<th align="left">Description</th>
</tr>
</thead>
<tbody>
<tr>
<td colspan="2" valign="top"><strong>Before</strong></td>
<td valign="top"><a href="#timestamp">Timestamp</a></td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>After</strong></td>
<td valign="top"><a href="#timestamp">Timestamp</a></td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>Between</strong></td>
<td valign="top"><a href="#timestamprange">TimestampRange</a></td>
<td></td>
</tr>
</tbody>
</table>

### TimestampRange

An inclusive range of timestamps, with one or two points to be defined.
A range with no fields defined is equal to a nil range.

<table>
<thead>
<tr>
<th colspan="2" align="left">Field</th>
<th align="left">Type</th>
<th align="left">Description</th>
</tr>
</thead>
<tbody>
<tr>
<td colspan="2" valign="top"><strong>Start</strong></td>
<td valign="top"><a href="#timestamp">Timestamp</a></td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>End</strong></td>
<td valign="top"><a href="#timestamp">Timestamp</a></td>
<td></td>
</tr>
</tbody>
</table>

### TransactionsFilter

All possible filters that can be used to query for a list of transactions.
Multiple filters can be combined. It is also valid that none are given.

<table>
<thead>
<tr>
<th colspan="2" align="left">Field</th>
<th align="left">Type</th>
<th align="left">Description</th>
</tr>
</thead>
<tbody>
<tr>
<td colspan="2" valign="top"><strong>Versions</strong></td>
<td valign="top">[<a href="#byteversion">ByteVersion</a>!]</td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>ArbitraryData</strong></td>
<td valign="top"><a href="#binarydatafilter">BinaryDataFilter</a></td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>CoinInputValue</strong></td>
<td valign="top"><a href="#bigintfilter">BigIntFilter</a></td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>CoinOutputValue</strong></td>
<td valign="top"><a href="#bigintfilter">BigIntFilter</a></td>
<td></td>
</tr>
</tbody>
</table>

## Enums

### BlockPayoutType

The different types of Payouts one can find in a block (header).

<table>
<thead>
<th align="left">Value</th>
<th align="left">Description</th>
</thead>
<tbody>
<tr>
<td valign="top"><strong>BLOCK_REWARD</strong></td>
<td></td>
</tr>
<tr>
<td valign="top"><strong>TRANSACTION_FEE</strong></td>
<td></td>
</tr>
</tbody>
</table>

### LockType

<table>
<thead>
<th align="left">Value</th>
<th align="left">Description</th>
</thead>
<tbody>
<tr>
<td valign="top"><strong>BLOCK_HEIGHT</strong></td>
<td></td>
</tr>
<tr>
<td valign="top"><strong>TIMESTAMP</strong></td>
<td></td>
</tr>
</tbody>
</table>

### OutputType

The different types of `Output` possible within the context of this API.

<table>
<thead>
<th align="left">Value</th>
<th align="left">Description</th>
</thead>
<tbody>
<tr>
<td valign="top"><strong>COIN</strong></td>
<td></td>
</tr>
<tr>
<td valign="top"><strong>BLOCK_STAKE</strong></td>
<td></td>
</tr>
<tr>
<td valign="top"><strong>BLOCK_CREATION_REWARD</strong></td>
<td></td>
</tr>
<tr>
<td valign="top"><strong>TRANSACTION_FEE</strong></td>
<td></td>
</tr>
</tbody>
</table>

## Scalars

### BigInt

BigInt represents an unbound integer type.
It is decimal (base 10) encoded as a string.
Within the context of this API it is used for currency values (coins and block stakes).

### BinaryData

BinaryData is the scalar type used by this API as the go-to
binary byte-slice type. It is always hex-encoded within the context of this API.

### BlockHeight

BlockHeight is implemented as an unsigned 64-bit integer,
and represents the height of a block, stargting at 0.

### Boolean

The `Boolean` scalar type represents `true` or `false`.

### ByteVersion

ByteVersion is a generic type, an unsigned 8-bit integer (equivalent to a single byte),
used for any place where we use such Versions.
Examples of such versions are `Transaction` and `UnlockCondition` versions.

### Cursor

A hex-encoded MsgPack-based cursor,
allowing to continue a query from where
you started.

### Hash

Hash represents a crypto (blake2b_256) `Hash` (a byte array of fixed length 32),
and is used as the identifier for blocks, transactions and outputs.
Within the context of this API it is always hex-encoded.

### Int

The `Int` scalar type represents non-fractional signed whole numeric values. Int can represent values between -(2^31) and 2^31 - 1.

### LockTime

LockTime is perhaps an unfortunate name and is is used for places where either a `BlockHeight`
or `Timestamp` is used. It is named like this for historic reasons, as a block height can also be seen as some kind of time unit.

### ObjectID

ObjectID is the generic identifier type for an object.
Can be a crypto (blake2b_256) `Hash` (for block-, transaction-, and output identifiers),
or an `UnlockHash` (for any kind of wallet or contract).

### PublicKey

PublicKey is the go-to type of this API for cryptographic public keys.
Within the context of this API it is encoded as a 2 part string, separated by a colon,
with the first part identifying the cryptographic algorithm and the second part the hex-encoded
public key (currently always 32 bytes, or 64 characters hex-encoded, at the moment).
At the moment only one cryptographic (signature) algorithm is supported.
This algorithm is ED25119 and identifier as `ed25519`.
Please see documentation such as
https://github.com/threefoldtech/rivine/blob/master/doc/transactions/transaction.md#json-encoding-of-a-singlesignaturefulfillment
for an example of such an encoded public key. Technical details can be found on
https://godoc.org/github.com/threefoldtech/rivine/types#PublicKey

### Signature

Signature is the go-to type of this API for cryptographic signatures.
It is hex-encoded and can be seen by implementations of this API as a raw byte slice,
its content identified by the algorithm defined in the linked public key.
Please consult the documentation in the `PublicKey` type for more information about public keys
within the context of this API.

### String

The `String` scalar type represents textual data, represented as UTF-8 character sequences. The String type is most often used by GraphQL to represent free-form human-readable text.

### Timestamp

Timestamp is implemented as an unsigned 64-bit integer,
and represents the UNIX Epoch Timestamp (in seconds).

### UnlockHash

UnlockHash represents an address of a wallet or contract,
and can be stored as a fixed array of 32 bytes (the hash part) as well as one byte for the type.
Within the context of this API it is always hex-encoded,
where the first 2 characters represent the type byte, the next 64 characters represent
the hex-encoded hash and the last 12 characters represent the hex-encoded 6-byte checksum.
Please consult https://github.com/threefoldtech/rivine/blob/master/doc/transactions/unlockhash.md#textstring-encoding
for more information about the full details of the encoding used for this type within the context of this API.


## Interfaces


### Transaction

<table>
<thead>
<tr>
<th align="left">Field</th>
<th align="right">Argument</th>
<th align="left">Type</th>
<th align="left">Description</th>
</tr>
</thead>
<tbody>
<tr>
<td colspan="2" valign="top"><strong>ID</strong></td>
<td valign="top"><a href="#hash">Hash</a>!</td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>Version</strong></td>
<td valign="top"><a href="#byteversion">ByteVersion</a>!</td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>ParentBlock</strong></td>
<td valign="top"><a href="#transactionparentinfo">TransactionParentInfo</a></td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>CoinInputs</strong></td>
<td valign="top">[<a href="#input">Input</a>!]</td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>CoinOutputs</strong></td>
<td valign="top">[<a href="#output">Output</a>!]</td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>FeePayouts</strong></td>
<td valign="top">[<a href="#transactionfeepayout">TransactionFeePayout</a>!]</td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>ArbitraryData</strong></td>
<td valign="top"><a href="#binarydata">BinaryData</a></td>
<td></td>
</tr>
</tbody>
</table>

### UnlockCondition

An unlock condition is used to define "who"
can spent a coin or block stake output.

See the different `UnlockCondition` implementations
to kow the different conditions that are possible.

<table>
<thead>
<tr>
<th align="left">Field</th>
<th align="right">Argument</th>
<th align="left">Type</th>
<th align="left">Description</th>
</tr>
</thead>
<tbody>
<tr>
<td colspan="2" valign="top"><strong>Version</strong></td>
<td valign="top"><a href="#byteversion">ByteVersion</a>!</td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>UnlockHash</strong></td>
<td valign="top"><a href="#unlockhash">UnlockHash</a></td>
<td></td>
</tr>
</tbody>
</table>

### UnlockFulfillment

<table>
<thead>
<tr>
<th align="left">Field</th>
<th align="right">Argument</th>
<th align="left">Type</th>
<th align="left">Description</th>
</tr>
</thead>
<tbody>
<tr>
<td colspan="2" valign="top"><strong>Version</strong></td>
<td valign="top"><a href="#byteversion">ByteVersion</a>!</td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>ParentCondition</strong></td>
<td valign="top"><a href="#unlockcondition">UnlockCondition</a></td>
<td></td>
</tr>
</tbody>
</table>

### Wallet

A wallet is identified by an `UnlockHash` and can be sent
coins and block stakes to, as well as spent those
coins and block stakes received. In practise
it is nothing more than a storage of a private/public key pair
a public key which can be converted to (and exposed as) an UnlockHash
to look up its balance, in the case of a non-full client wallet.

See the Wallet implementations for the different wallets that are possible.

<table>
<thead>
<tr>
<th align="left">Field</th>
<th align="right">Argument</th>
<th align="left">Type</th>
<th align="left">Description</th>
</tr>
</thead>
<tbody>
<tr>
<td colspan="2" valign="top"><strong>UnlockHash</strong></td>
<td valign="top"><a href="#unlockhash">UnlockHash</a>!</td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>CoinBalance</strong></td>
<td valign="top"><a href="#balance">Balance</a></td>
<td></td>
</tr>
<tr>
<td colspan="2" valign="top"><strong>BlockStakeBalance</strong></td>
<td valign="top"><a href="#balance">Balance</a></td>
<td></td>
</tr>
</tbody>
</table>
