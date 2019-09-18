# Rivine.io
May 02, 2017

Version 0.1

## Overview

A blockchain, as introduced by Bitcoin[1], is a technological concept that allows consensus between an unknown number of parties about the timestamp and validity of transactions in a decentralized peer2peer network without externally trusted parties.
While the application of the Bitcoin blockchain is limited to transferring Bitcoins (and registering small amounts of data),  this technology can be leveraged to provide much more exciting functionality like smart contracts that execute on the chain without external parties governing what is true and what is not. Legal agreements can be converted to such a digital contract in a way they are verifiable by participants and executed according to the digital rules, one can not just choose to ignore such a contract.
Another problem is the number of transactions per second that can be processed in such a global chain and that private forks are not secure.

## Goals

* **Green:** Traditional Proof Of Work (POW) requires a lot of mining power to secure the chain.[2]
* **Secure:** Blockchains are a solution to the “double spending problem”. They give a timestamp to published transactions by having a broad consensus over which transactions are valid and included in the blockchain. By distributing the power to participate in the consensus, failing/hacked nodes or malicious participants do no longer pose a threat. POW requires to have a tremendous amount of mining power to be secure and Traditional Proof Of Stake (POS) is not perfect either[3].
* **Scalable:** Blockchains do not scale very well, the amount of transactions per second is limited. We need a way to overcome this limitation.
* **Smart contracts:** A token transfer transaction is the simplest kind of contract, it just states that some inputs are being converted to some (spendable) outputs. We need more advanced contracts to be supported like making reservations for computing capacity where a consuming party might be able to stop the contract if the provider fails to deliver.
* **Micropayments support:** Performing micropayments on-chain does not make sense, it costs a lot on transaction fees and increases the chain size while these micropayments often occur between the same parties. The average block time is 10 minutes and since someone should wait for a couple of blocks to ensure the payment will not be reversed, making small payments for services someone else delivers continuously becomes a slow process.
* **No dependency on external tokens:** When leveraging an existing chain, the token used for transaction fees is the base currency of that chain. This means consumers interacting with the smart contracts or people transferring tokens need to be aware of the underlying base token and make sure they acquire this base token to interact with the overlay logic. Pretty annoying and difficult to explain.
* **Private:** It should be possible to create private chains while not giving in on security.
* **Lightweight client support:** A chain can grow significantly in size and it might take a while to synchronize initially. Devices that are not constantly connected to the peer2peer network or do not have the required storage need to be able to interact with the chain as well. Lightweight clients are wallets that contain private keys, can view transactions and interact with the blockchain without having a copy of the blockchain locally or participate in the peer2peer protocol directly. A secure way to enable lightweight clients that guarantees correctness and reliability is required.
* **Notary:** It should be possible to register arbitrary data on the chain.
* **Hierarchical deterministic wallets:** Addresses are generated in a known fashion rather than randomly so some clients can be used on multiple devices without the risk of losing funds. Users can conveniently create a single backup of the seed in a human-readable format that will last the life of the wallet, without the worry of this backup becoming stale.
Additionally, there is a complete separation of private and public key creation for greater security and convenience. In this model, a server can be set up to only know the Master Public Key of a particular deterministic wallet. This allows the server to create as many public keys as is necessary for receiving funds, but a compromise of the MPK will not allow an attacker to spend from the wallet.
* **Authorized addresses:** Most blockchain implementations do not care about the addresses used, it is a hash-derivation of the public key from a random byte string used as a private key. When token ownership is required to be gone through accepted KYC (Know Your Customer) procedures, this is not an option. It should be possible to pass on the authorizations according to the hierarchical deterministic wallet structure.
* **Simple:** Too many components, layers, abstractions, and integrations only cause bugs.

## Problems with existing implementations

An option would be to use an existing, public chain like to implement the required goals. One can argue that even though it is POW, it’s still green since the mining power is there anyway. We would, however, contribute to transaction fees and stimulate more mining capacity to be deployed.
While leveraging an existing public chain would be the simplest thing to do, the base currency would be the one of that chain. Using Ethereum and creating the extra functionality in smart contracts there would still require Ethereum to be used as gas when interacting with smart contracts. Next to having a limit of 20 transactions/second for the entire global Ethereum network, syncing and having a copy of the entire Ethereum chain while we are only interested in the interactions with our contracts is quite an overhead, especially during the initial block download.
Forking such a general platform is no option either, especially not if they are POW based like Ethereum. It’s not green and it would require a tremendous amount  of mining power to secure such a fork.

General-purpose blockchain platforms are complex. Smart contracts are written in higher level languages like solidity, compiled to some sort of bytecode and executed in sandboxes. Mistakes can happen everywhere, even with people that know the system very well like the well known DAO hack proved[4].
Some of the goals like authorized addresses, hierarchical deterministic wallets, notary, micropayments... are simple to accomplish when entering the blockchain code itself but would require custom wallet implementations that handle both the more complex blockchain and the smart contract interaction levels.

## Concepts

* **Basic principles and concepts:** Rivine is a Bitcoin variant and builds upon the principles and concepts of Bitcoin for transaction validity and token transfer. A good read is the original Bitcoin whitepaper by Satoshi Nakamoto[5].
* **Proof Of BlockStake (POS):** In a rivine network, two digital assets exist, normal coins and BlockStakes. BlockStakes are digital tokens as well and can be transferred to normal rivine addresses. The founder of a chain distributes these BlockStakes to other people and or nodes. The POS algorithm in Rivine can only be performed by BlockStakes, the normal coins can not participate to enlarge the chain, which does not mean they do not validate incoming transactions and blocks themselves off course. A node containing BlockStakes and that participates in the Proof Of BlockStake (POBS) protocol is called a Block Creator Node (BCN).  
You can think of BlockStakes as being the shares of the blockchain. Since shareholders are the owners of a company (and if real-world assets are linked, are liable if something goes wrong), they collaboratively control what eventually ends up in the chain or not according to the defined rules. These BlockStakes must be well distributed so a hack of systems containing BlockStakes does not allow an attacker to do a 51% attack.
* **Multiple chains:** Because the POBS consensus protocol does not require massive amounts of distributed mining power to secure it, starting special purpose chains does not pose a security threat.
* **Off-chain payment channels:** Within a specific contract, micropayments can be exchanged to reflect the used amount of computing resources for example. While the contracts themselves are registered on the chain, the updates and micropayments themselves are not. The payment updates are negotiated off-chain and only the updated version that both parties agreed upon can be posted back on the chain, effectively executing the combined payments. This allows for a lot of transactions while not consuming the transactions/second limitations of a normal blockchain.
* **Contracts with locked collateral:** When creating smart contracts, collateral can be locked on it, it’s a guarantee that the parties are agreeing to keep their promise since they will lose the collateral if they don’t. In this case, the collateral is burned, no one benefits from it since that would be an incentive for people that can become the beneficiary of collateral to try to make a party unable to fulfill its promise (like a DOS attack on a provider for example). Collateral is formed in normal coins, blockstakes are not eligible for collateral deposits.
* **Gateways and offline transaction signing:** To support lightweight clients like mobile phones, a full node can expose a gateway to the consensus that also accepts posting fully signed transactions back to the network. A lightweight client can connect to multiple gateways it chooses itself to validate the data they return. It can post a signed transaction multiple times to different gateways but it will end up uniquely in the chain.

## Specifications

### Proof Of BlockStake

#### Green

In traditional POW, the (simplified) formula to create a block is HASH(merkletree(tx_1..tx_n)+nonce+previous_block_id) < difficulty. Miners constantly modify the nonce, calculate the hash and check if it is smaller than the difficulty at that time. The more compute power and electricity you spend, the more chances you have to find a block.
In POBS, the variables hashed are chosen in a way that there is no randomness and people that are participating don’t win anything by trying to manipulate them. Since the search space is limited to 1 calculation per unspent blockstake output per second, the processing power required is neglectable.

#### Protocol

The hash function used is a 32-byte BLAKE2b hash. To compare the hash with the difficulty it is interpreted as a big-endian unsigned integer.
Maturity of blockstakes

Unspent BlockStake Outputs not on index 0 (transferred blockstakes) need to mature for 144 blocks before they can participate in the POBS protocol. This removes the benefit of manipulating the UTXO index variable in the POBS hash and pre-calculating better future blockstake creation chances.

#### Maturity of collected transaction fees

Coins received through the collection of transaction fees during block creation can't be spent until the block has 144 confirmations. Transactions that try to spend a block creation fee output before this will be rejected.
The reason for this is that sometimes the blockchain forks, valid blocks become invalid, and the block creation reward in those blocks is lost. That's just an unavoidable part of how blockchains work and it can sometimes happen even when there is no one attacking the network. If there was no maturation time, then whenever a fork happened, everyone who received coins that were collected on an unlucky fork (possibly through many intermediaries) would have their coins disappear, even without any sort of double-spend or other attacks. On long forks, thousands of people could find coins disappearing from their wallets, even though there is no one attacking them and they had no reason to be suspicious of the money they were receiving. For example, without a maturation time, a block creator might deposit 25 coins into an EWallet, and if I withdraw money from a completely unrelated account on the same EWallet, my withdrawn money might just disappear if there is a fork and I'm unlucky enough to withdraw coins that have been "tainted" by the block creator's now-invalid coins. Therefore, this sort of taint tends to "infect" transactions, far more than 25 coins per block would be affected. Each invalidated block could cause transactions collectively worth hundreds of coins to be reversed. The maturation time makes it impossible for anyone to lose coins by accident like this as long as a fork doesn't last longer than 144 blocks.

#### Difficulty

The hash in the POBS protocol results in a 256-bit integer so there are 2^256 possible combinations. If we want 1 block to be created every 10 minutes on average, this means that 1 hash should match every 10*60 seconds. The chance of having a match is also multiplied by the number of blockstakes you have available so for the starting difficulty this means it should be divided by the total number of blockstakes in the system.
Difficulty is adjusted every 50 blocks to compensate for the fact that not every blockstake available is always participating in the POBS protocol.

### Notary

Every transaction can contain arbitrary data. No validation is performed on the data itself. The maximum size of this arbitrary data is 1KiB. The minimal transaction fee depends on the size of the transaction so adding arbitrary data increases the required fee.

### Off-chain payment channels

Rivine provides the possibility to register a payment channel on the chain and the micropayments are agreed upon between the different parties directly. If one of the parties decides to execute the entire payment, it is pushed to the chain where it is executed. Only the last version of the payment contract is valid and will be accepted.
### Smart contracts with locked collateral

### Gateways and offline transaction signing

Every full node can run a public gateway. It indexes the entire chain and provides a public API to access it. Next to this “explorer” app, it accepts entire transactions and puts them on the network.

#### Smart gateways

Having a fixed number of publicly available gateways is easy but not secure, dynamic nor fair. It would be better if anyone could host a public gateway and everyone can discover it. To keep this secure, the gateways create smart contracts with locked collateral. The incentive to host such a gateway is that they can charge for providing the information and for putting transactions on the network. Proof of malicious gateways would make them lose the collateral.

[1] https://bitcoin.org/bitcoin.pdf  
[2] https://www.coindesk.com/carbon-footprint-bitcoin  
[3] https://download.wpsoftware.net/bitcoin/pos.pdf  
[4] https://www.wired.com/2016/06/50-million-hack-just-showed-dao-human/  
[5] https://bitcoin.org/bitcoin.pdf  
