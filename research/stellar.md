# Stellar
## concept
All assets in stellar are credit, issued by anchors.

When you hold assets in Stellar, you’re actually holding credit from a particular issuer that promises to return the real world value ( for example USD) when you return it.

License wise, this might pose a problem

## Distributed Exchange
Assets are linked to the issuing Anchor. An order can only be executed if both buyer and seller have a trustline to the issuing party.

Supports asset conversion up to 6 hops, but the whole payment is atomic.

**Assets must be present as credit on the stellar network.** 
This works perfectly if your assets only exist on the stellar network as credit or if users interact through centralized parties holding their assets( like bitcoin bank or traditional exchange).

_Question is how a credit issuer that acts as an exchange for it's users can control/recuperate free floating real world assets like BTCor TFT that stay in control of the user._

Fees are in XLM.
## random thoughts

Not completely decentralized,for example sending with kyc, requires knowledge and trust of the counterparty, also 
Very good solution for traditional centralized banks that want to collaborate:
https://www.stellar.org/wp-content/uploads/2016/08/Sending-Payment-Flow-Detailed.jpg

However, the protocol includes who can hold assets: https://www.stellar.org/developers/guides/concepts/assets.html#controlling-asset-holders

### Wallets
A variety of wallets is available, each with their own features or focus, it makes it a bit hard though as a beginner.


## Stellar Smart Contracts (SSC)
Not as flexible as for example Ethereum but this might be a good thing.

compositions of transactions that are connected and executed using various constraints:
- Multisignature
- Batching/Atomicity
- Sequence
- Time Bounds

**Seems sufficient for crowdfunding** or other basic financial operations or agreements.

**Atomic swaps** can be implemented since stellar supports multisig, [SHA256 hash](https://www.stellar.org/developers/guides/concepts/multi-sig.html#hashx) and timebounds.

## Ico's, crowdfunding, ...
As mentioned above, the stellar platform is suited for for crowdfunding or other basic financial operations or agreements

Possible options to issue credit:
- Already be on the stellar platform with trusted credit.
- Accept fiat or cryptocurrencies directly and issue credit on the stellar platform

We would have to provide a small wallet if people do not already have one.



