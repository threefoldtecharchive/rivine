# Cross chain atomic swapping

Rivine implementation of [Decred atomic swaps](https://github.com/decred/atomicswap).

# Theory

A cross-chain swap is a trade between two users of different cryptocurrencies. For example, one party may send Rivine coins to a second party's Rivine address, while the second party would send Bitcoin to the first party's Bitcoin address. However, as the blockchains are unrelated and transactions can not be reversed, this provides no protection against one of the parties never honoring their end of the trade. One common solution to this problem is to introduce a mutually-trusted third party for escrow. An atomic cross-chain swap solves this problem without the need for a third party.

Atomic swaps involve each party paying into a contract transaction, one contract for each blockchain. The contracts contain an output that is spendable by either party, but the rules required for redemption are different for each party involved.

# Example

Let's assume someone( from now one called the buyer or initiator) wants to buy 100 Rivine based tokens from someone else ( the seller or participant) for 1BTC

The seller creates a bitcoin address and the buyer creates a Rivine address: 
Seller -> create bitcoin address 
buyer -> create Rivine address 

Let the buyer initiate the swap, he generates a secret and hashes it as well
buyer-> create secret
hashsecret = sha256(secret)

The buyer now creates a swap transaction and publishes it on the bitcoin chain, it has 1btc as an output and the output can be redeemed( used as input) using 1 of the following conditions:
- timeout has passed ( 48hours) and claimed by the buyers refund address
- the secret is given that hashes to the hashsecret and claimed by the seller's address

If the atomic swap process fails, the buyer can always reclaim it's btc after the timeout.
 Now the buyer sends this contract and the transaction id of this transaction on the bitcoin chain to seller, making sure he does not share the secret off course.

 Now the seller validates if everything is as agreed (=audit) The seller now creates a similar transaction on the Rivine chain but with a timeout for refund of only24 hours.
 The transaction has 100 tokens as an output and the output can be redeemed( used as input) using 1 of the following conditions:
- timeout has passed ( 24hours) and claimed by the sellers refund address
- the secret is given that hashes to the hashsecret and claimed by the buyers's address

In order for the buyer to claim the tokens, he has to use and disclose the secret.
Now the seller can use the secret to claim the 1btc.

Off course, either party can be the initiator or the participant.




