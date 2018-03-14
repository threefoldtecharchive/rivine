# Cross chain atomic swapping


# Theory

A cross-chain swap is a trade between two users of different cryptocurrencies. For example, one party may send Rivine based tokens to a second party's Rivine address, while the second party would send Bitcoin to the first party's Bitcoin address. However, as the blockchains are unrelated and transactions can not be reversed, this provides no protection against one of the parties never honoring their end of the trade. One common solution to this problem is to introduce a mutually-trusted third party for escrow. An atomic cross-chain swap solves this problem without the need for a third party.

Atomic swaps involve each party paying into a contract transaction, one contract for each blockchain. The contracts contain an output that is spendable by either party, but the rules required for redemption are different for each party involved.

# Example

Let's assume Bob wants to buy 100 Rivine based tokens from someone Alice for 1BTC.

Bob creates a bitcoin address and Alice creates a Rivine address.

Bob -> create bitcoin address 

Alice -> create Rivine address 

Bob initiates the swap, he generates a secret and hashes it as well

Bob-> create secret and hashsecret = sha256(secret)

Bob now creates a swap transaction and publishes it on the bitcoin chain, it has 1 bitcoin as an output and the output can be redeemed( used as input) using either 1 of the following conditions:
- timeout has passed ( 48hours) and claimed by the Bob refund address
- the secret is given that hashes to the hashsecret created by Bob and claimed by Alice's address

This means Alice can claim the bitcoin if she has the secret and if the atomic swap process fails, Bob can always reclaim it's btc after the timeout.
 
 Now Bob sends this contract and the transaction id of this transaction on the bitcoin chain to Alice, making sure he does not share the secret of course.

 Now Alice validates if everything is as agreed (=audit)after which She creates a similar transaction on the Rivine chain but with a timeout for refund of only24 hours and she uses the same hashscecret as the first contract for Bob to claim the tokens.
 This transaction has 100 tokens as an output and the output can be redeemed( used as input) using either 1 of the following conditions:
- timeout has passed ( 24hours) and claimed by the sellers refund address
- the secret is given that hashes to the hashsecret Bob created (= same one as used in the bitcoin swap transaction) and claimed by the buyers's address

This means Bob can claim the Rivine tokens using his secret and if the atomic swap process fails, Alice  can always reclaim her tokens after the timeout.

In order for Bob to claim the tokens, he has to use and as such disclose the secret.
Now Alice can use this secret to claim the bitcoin.

The magic of the atomic swap lies in the fact that the same secret is used to claim the tokens in both swap transactions but it is not disclosed in the contracts because only the hash of the secret is used there. The moment Bob claims the tokens, he discloses the secret and Alice still has renough time to claim the bitcoin because the timeout of the first contract is longewr than the one of the second contract

Of course, either Bob or Alice can be the initiator or the participant.


## References

Rivine atomic swaps are an implementation of [Decred atomic swaps](https://github.com/decred/atomicswap).