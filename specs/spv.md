 # SPV, Simplified Payment Verification

 ## Problem
 The  thin client  implementation through the explorers requires trust in the explorers  even though a thin client can call multiple. 
There is also no fast way for thin clients to properly validate if transactions are really included in the blockchain. 
 
 Clients expose quite some information to the explorers as well ( which addresses the hold and the transactions they are making).
 
## Proposed solution
  A more secure thin client implementation like [bitcoin SPV](https://bitcoin.org/en/developer-guide#simplified-payment-verification-spv)  combined with bloomfilters would be better.
