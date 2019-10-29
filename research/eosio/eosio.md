# eosio

## Consensus

Permissioned blockchain using a delegated proof of stake algorithm.

## Smart contracts

Written in C++ and compiled to wasm while other compilers exist for rust and solidity although they are not officially supported.

### Verification

Verifying a smart contract can not be done without the source code being published and having a reproducible build( while needing the constructor arguments) and it is still hard to read the code and find flaws.

Smart contracts can simply be upgraded by the publisher.

## Scalability

Eos procceses transactions by using 21 block propagators, which speeds up the process significantly. If one blockchain is not enough, other chains can be created to increase the number of transactions per second.

## Blochain interoperability

- Relays: contracts on both chains, much like our erc-20 bridge.
- atomic swaps possibility

Due to the fast validation on eos, this is much faster then on traditional blockchains.

## Private chains

Using the eos.io toolkit, it is easy to set up private chains.

## Controversy

A lot of discussion is going on about the centralization of Block propagators (90% in China) and some questionable actions they took(freezing of accounts).

Even a sentiment of calling the entire EOS ecosystem a Scam.

Block.one, the company behind eos managed to raise a huge amount during their ICO.
Yet they fail to deliver on some promises like withdrawing the announced decentralized facebook project.

## Questions

- While it seems that the platform is created for making private chains, what does the public EOS token serve?
  - "A Developer needs to hold eos tokens, not spend them in order to use network resources and to build and run dApps".
