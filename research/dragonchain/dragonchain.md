# Dragonchain

Detailed [Architecture](https://dragonchain-core-docs.dragonchain.com/latest/overview/architecture.html)

## Verification and consensus

Multi level verification( Business, validation, network validation, notary, checkpointing)is  an interesting concept for security, it does make it ( unnecessary ? ) complex.

## Smart contracts

Unlike traditional smart contracts, Dragonchain smart contracts are simply code that can be executed as:

- Hardcoded in a forked codebase
- Configured (with smart contract code deployed on the system)
- Delivered via blockchain (admin would send multi-signed transaction with smart contract, start date, etc.)

In other words, execute whatever code and have the business verification consensus level have it trusted. The most common way is to execute your code in a docker container.

This is pretty blackboxed.

OpenFaas and docker Registry are needed for smart contract deployment and execution.

## Data distribution

Only headers are distributed in the consensus process, if a node requires more data, if will have to get it from the owning node.

## Interoperability

- Possible to wrap foreign cryptocurrency transactions within a private blockchain transaction.
- Checkpointing

## Storage

Redis (+ cache LRU redis and Redisearch for indexing).

## Blockchain as a service

Combines the above features in a cloud hosted model, paid with "Dragons", the token of Dragonchain wich at first sight looks like an ERC-20 token.

**The console requires you to have an understanding of the dragonchain concepts, is not for novice users.**

Starts with $99/month for a L1 managed chain on a single node.
The out of the box smart contracts are limited:

- A single currency
- Ethereum and Bitcoin Watcher
- Ethereum and Bitcoin publisher

Not really clear what the pricing of custom contracts ( docker containers) is.
