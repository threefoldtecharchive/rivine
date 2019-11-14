# Beyond basic atomic swaps

## Problems
While atomic swaps are great and enable trustless trading, they do have some drawbacks.

### 1. Current toolset requires wallets with full nodes 

 
### 2. slow
It can take minutes for a trade to complete.

### 3. Transaction fees per chain can be quite high.

 ### 4 Only penaltie for  malicious users bailing out of the swap is the transaction fee cost
 This  is not only annoying for other users but can also cost the other party transaction fees and lock their funds for a while. A naive orderbook implementation can easily be attacked this way.
## Solutions

### 1 Eliminate the need to download the blockchain.

The atomic swap tools should be able to use thin clients. Rivine currently has no commandline thin client. For bitcoin, we could use [Electrum](https://electrum.org/).

Issues:

- Generic commandline Rivine thin client: https://github.com/threefoldtech/rivine/issues/378

Jumpscale has thin client atomic swap support for Rivine based chains like tfchain.

Thin client support for other currencies are available for Bitcoin and Stellar.

### 2 Payment channels and a lightning network

## References

- [Bitcoin Simplified Payment Verification](https://bitcoin.org/en/developer-guide#simplified-payment-verification-spv)
