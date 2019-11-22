# Stellar

## concept

All assets in stellar are credit, issued by anchors.

When you hold assets in Stellar, you’re actually holding credit from a particular issuer that promises to return the real world value ( for example USD) when you return it.

License wise, this might pose a problem.

## Consensus

[Stellar Consensus protocol](https://medium.com/stellar-development-foundation/on-worldwide-consensus-359e9eb3e949)

## Distributed Exchange

Assets are linked to the issuing Anchor. An order can only be executed if both buyer and seller have a trustline to the issuing party.

Supports asset conversion up to 6 hops, but the whole payment is atomic.

**Assets must be present as credit on the stellar network.**

This works perfectly if your assets only exist on the stellar network as credit or if users interact through centralized parties holding their assets( like bitcoin bank or traditional exchange).

_Question is how a credit issuer that acts as an exchange for it's users can control/recuperate free floating real world assets like BTCor TFT that stay in control of the user._

Fees are in XLM.

## Stellar accounts

Just creating a keypair and an address like is common blockchains is not sufficient to receive funds, a transfer will fail. An account has to be explicitely created and funded with at least 1 XLM.

### Wallets

A variety of wallets is available, each with their own features or focus, it makes it a bit hard though as a beginner, especially since you need to know the concept of trustlines.

## Stellar Smart Contracts (SSC)

Not as flexible as for example Ethereum but this might be a good thing.

compositions of transactions that are connected and executed using various constraints:

- Multisignature
- Batching/Atomicity
- Sequence
- Time Bounds

**Seems sufficient for crowdfunding** or other basic financial operations or agreements.

**Atomic swaps** can be implemented since stellar supports multisig, [SHA256 hash](https://www.stellar.org/developers/guides/concepts/multi-sig.html#hashx) and timebounds.

## Custom assets

It is easy to create custom assets on the Stellar network as the [issuetoken example](./issuetoken/readme.md) shows.

### Token creation reasons

Stellar transactions can contain a memotext up to 28 bytes of ASCII/UTF-8 which is not sufficient to hold a sha256 encoded hash.

There is a Memohash field MemoHash which is a hash representing a reference to another transaction.

This can be used however to insert the hash of other documents, a concept known as [stellar attachments](https://www.stellar.org/developers/guides/attachment.html).

## Company accounts and multisig

[Stellar allows multiple signatures for custom asset issuer accounts or company accounts](https://www.stellar.org/developers/guides/concepts/multi-sig.html).

## Multiple addresses and anonimity

Stellar is not meant for anonimity.
You can have multiple registered accounts, each with their own balances, requiring XLM to perform transactions.

Payments from multiple addresses are not possible(as a workaround, one can merge the accounts first).

## Federation servers

This allows easy addresses like  bob@yourdomain.com.

## Ico's, crowdfunding

As mentioned above, the stellar platform is suited for for crowdfunding or other basic financial operations or agreements

Possible options to issue credit:

- Already be on the stellar platform with trusted credit.
- Accept fiat or cryptocurrencies directly and issue credit on the stellar platform

We would have to provide a small wallet if people do not already have one.

## Market making

A bifrost server exists that will automatically exchange the received BTC or ETH for your custom token. Usefull for market making and ICO's.

There is also Kelp, a free, customizable, open-source trading bot for the Stellar universal marketplace.

## Signature schemes

Stellar currently uses the ed25519 signature scheme which is the same as Rivine currently.

## Storage for Core and Horizon

PostgreSQL

## Scalability

Due to the Stellar Consensus protocol, very fast settlement is achieved (seconds). Stellar itself claims 1500 TPS
while people claim to reach 10 000 TPS easily on Google cloud platform without optimizations.

Private/public stellar networks can easily be created seperate from the  main Stellar network.

Validating nodes do not have to archive the entire history while a full validator or archiver does consume  terrabytes for the full public stellar network.

## random thoughts

[Not completely decentralized,for example sending with kyc, requires knowledge and trust of the counterparty.
Very good solution for [traditional centralized banks that want to collaborate](
https://www.stellar.org/wp-content/uploads/2016/08/Sending-Payment-Flow-Detailed.jpg).

However, [the protocol includes who can hold assets](https://www.stellar.org/developers/guides/concepts/assets.html#controlling-asset-holders).

A very extensive toolset for managing custom tokens and ICO's.
