# Cardano concept

[Cardano website](https://www.cardano.org)

## Concept

Call themselves the 3rd genetion blockchain with

* Bitcoin being generation 1: a decentralized payment system
* Ethereum starts generation 2: Possibility of scripted contracts

### Principles

* Scalability
  * TPS
  * Bandwith
  * Data
* Interoperability
* Sustainability

* Peer review
* High assurance code

### Consensus

Ouroboros: Proof Of Stake consensus algorithm

Creates epochs each with a slot leader by election. Cheap to construct a block, one can run multiple blockchains in parallel.

### Scalability

Ouroboros Proof Of Stake consensus protocol

#### Bandwith

RINA: **TODO**

#### Data

Not everyone needs all the data.

Pruning, Subscriptions, compression, Partionioning

Sidechains

#### TPS

The consensus, bandwith and data solutions allow high TPS

### Interoperability

Not "1 system to rule them all".

Concepts in the traditional world:

* Metadata
* Attribution (part of the metadata,from to who)
* Compliance
  * KyC
  * AML
  * ATF

One should be able to opt in on these features.

### Sustainability

* Built in funding system (treasury) with democratic voting for the spending.
* Change by improvement proposals instead of forks

## Technology

Written in Haskell
