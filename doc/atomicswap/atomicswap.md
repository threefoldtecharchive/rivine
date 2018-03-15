# Cross chain atomic swapping


# Theory

A cross-chain swap is a trade between two users of different cryptocurrencies. For example, one party may send Rivine based tokens to a second party's Rivine address, while the second party would send Bitcoin to the first party's Bitcoin address. However, as the blockchains are unrelated and transactions can not be reversed, this provides no protection against one of the parties never honoring their end of the trade. One common solution to this problem is to introduce a mutually-trusted third party for escrow. An atomic cross-chain swap solves this problem without the need for a third party.

Atomic swaps involve each party paying into a contract transaction, one contract for each blockchain. The contracts contain an output that is spendable by either party, but the rules required for redemption are different for each party involved.

# Example

Let's assume Bob wants to buy 9876 Rivine based tokens from someone Alice for 0.1234BTC.

Bob creates a bitcoin address and Alice creates a Rivine address.

Bob -> create bitcoin address 

Alice -> create Rivine address 

Bob initiates the swap, he generates a secret and hashes it as well

Bob-> create secret and hashsecret = sha256(secret)

Bob now creates a swap transaction and publishes it on the bitcoin chain, it has 0.1234 bitcoin as an output and the output can be redeemed( used as input) using either 1 of the following conditions:
- timeout has passed ( 48hours) and claimed by the Bob refund address
- the secret is given that hashes to the hashsecret created by Bob and claimed by Alice's address

This means Alice can claim the bitcoin if she has the secret and if the atomic swap process fails, Bob can always reclaim it's btc after the timeout.
 
 Now Bob sends this contract and the transaction id of this transaction on the bitcoin chain to Alice, making sure he does not share the secret of course.

 Now Alice validates if everything is as agreed (=audit)after which She creates a similar transaction on the Rivine chain but with a timeout for refund of only 24 hours and she uses the same hashscecret as the first contract for Bob to claim the tokens.
 This transaction has 9876 tokens as an output and the output can be redeemed( used as input) using either 1 of the following conditions:
- timeout has passed ( 24hours) and claimed by the sellers refund address
- the secret is given that hashes to the hashsecret Bob created (= same one as used in the bitcoin swap transaction) and claimed by the buyers's address

This means Bob can claim the Rivine tokens using his secret and if the atomic swap process fails, Alice  can always reclaim her tokens after the timeout.

In order for Bob to claim the tokens, he has to use and as such disclose the secret.
Now Alice can use this secret to claim the bitcoins.

The magic of the atomic swap lies in the fact that the same secret is used to claim the tokens in both swap transactions but it is not disclosed in the contracts because only the hash of the secret is used there. The moment Bob claims the tokens, he discloses the secret and Alice still has renough time to claim the bitcoin because the timeout of the first contract is longer than the one of the second contract

Of course, either Bob or Alice can be the initiator or the participant.

## Technical details of the example

secret: 32 bytes

hashsecret: Sha256( secret)

To Start bitcoin core qt in server mode on testnet: 
` ./Bitcoin-Qt  -testnet -server -rpcuser=user -rpcpassword=pass-rpcport=18332`


Alice creates a new bitcoin address (as of bitcoin core 0.16, make sure to specify the 'legacy' address type): 
```￼
getnewaddress "" legacy
mgmNWZN29WeFz3X4Na8thcbLB12JA5vj9j
```
Bob creates a new rivine address : 
```
rivinec wallet address
Created new address: bb6e12437c6fecbe83f5bf3724ced0369c01e166364cc320adf166125a8b6e2c756ada1be3f6
```

### initiate step
Bob initiates the process by using btcatomicswap to pay 0.1234BTC into the Bitcoin contract using Alice's Bitcoin address, sending the contract transaction, and sharing the secret hash (not the secret), contract, and contract transaction with B. The refund transaction can not be sent until the locktime expires, but should be saved in case a refund is necessary.

```
$btcatomicswap --testnet --rpcuser=user --rpcpass=pass initiate mgmNWZN29WeFz3X4Na8thcbLB12JA5vj9j 0.1234
Secret:      d685a0b8aacf03f024c84092b4b951d1e54f7747a72cd21ed091f16996502a8e
Secret hash: 2891f924fde4cc3c43af0d501a9fb52acb47b9a2e650c16ef0abb0a02c0ed988

Contract fee: 0.00000166 BTC (0.00000672 BTC/kB)
Refund fee:   0.00000297 BTC (0.00001021 BTC/kB)

Contract (2NAfLwhThYzB1kGYVxmjYqS98yXF8JBVc3v):
6382012088a8202891f924fde4cc3c43af0d501a9fb52acb47b9a2e650c16ef0abb0a02c0ed9888876a9140db229d573c1ca5042f1f6f8d95b0e48dd30f54c670418daac5ab17576a914dbb79258a0200feeef593cc753e3c0c21757a1306888ac

Contract transaction (afbc4dc719d9f79a9413945c92752bef644c618a4362fc8e8be0764a1b888e10):
0200000000010106a85a242263768b81554b0ccab63ec124146014338ec962ef0197a1043b867e0c00000017160014ad430d1ed266d7130b5207a9bef00f8030947c3ffeffffff02204bbc000000000017a914bf09ed70b0c505d750f333bad8ca0520e48370fb87104772000000000017a914616aac51e9c239f6f85fe6db12249f264dd5ff2987024730440220441058ac56f1db678f955610adc264d7982da25dc7079a36c9022bc72827311c0220209d552d7e5ac04cef5ab615edea6c7836e7276ae7dbe0c2950eb44c781c7c7b01210335c272b2cfbd3c0a02bb38bfd859e4a44175545eb12ce93e7a2bac690068f92f00000000

Refund transaction (028bcba1f15cc1ae6df1bd55881e8fdde2e36a863798bee71c969053d392ab90):
0200000001108e881b4a76e08b8efc62438a614c64ef2b75925c9413949af7d919c74dbcaf00000000ce47304402207382af301fc6fea74131929235696894daadba674576ebc991790ff8bbe54f4d02200c5f55bab13d0b336528bb6dc3e51b993a30f2071618b7e1298df582ff913cce01210272a09bdefb4d12536fbf2f782166daceef3113bf06898a0f46b9066b3de9e449004c616382012088a8202891f924fde4cc3c43af0d501a9fb52acb47b9a2e650c16ef0abb0a02c0ed9888876a9140db229d573c1ca5042f1f6f8d95b0e48dd30f54c670418daac5ab17576a914dbb79258a0200feeef593cc753e3c0c21757a1306888ac0000000001f749bc00000000001976a914328e89e0e93df379e593deb10ee9efafca53e08988ac18daac5a

Publish contract transaction? [y/N] y
Published contract transaction (afbc4dc719d9f79a9413945c92752bef644c618a4362fc8e8be0764a1b888e10)

``` 
You can check the transaction [on the bitcoin testnet blockexplorer](https://testnet.blockexplorer.com/tx/afbc4dc719d9f79a9413945c92752bef644c618a4362fc8e8be0764a1b888e10) where you can see that 0.1234 BTC is sent to2NAfLwhThYzB1kGYVxmjYqS98yXF8JBVc3v (= the contract script hash) being a [p2sh](https://en.bitcoin.it/wiki/Pay_to_script_hash) address in the bitcoin testnet. 

Decoding the contract( in the debug console or using the bitcoin-cli):
```￼
decodescript 6382012088a8202891f924fde4cc3c43af0d501a9fb52acb47b9a2e650c16ef0abb0a02c0ed9888876a9140db229d573c1ca5042f1f6f8d95b0e48dd30f54c670418daac5ab17576a914dbb79258a0200feeef593cc753e3c0c21757a1306888ac
￼
{
  "asm": "OP_IF OP_SIZE 32 OP_EQUALVERIFY OP_SHA256 2891f924fde4cc3c43af0d501a9fb52acb47b9a2e650c16ef0abb0a02c0ed988 OP_EQUALVERIFY OP_DUP OP_HASH160 0db229d573c1ca5042f1f6f8d95b0e48dd30f54c OP_ELSE 1521277464 OP_CHECKLOCKTIMEVERIFY OP_DROP OP_DUP OP_HASH160 dbb79258a0200feeef593cc753e3c0c21757a130 OP_ENDIF OP_EQUALVERIFY OP_CHECKSIG",
  "type": "nonstandard",
  "p2sh": "2NAfLwhThYzB1kGYVxmjYqS98yXF8JBVc3v"
}
```
Lets explain this script:
```
OP_IF   // top of Stack: secret
    OP_SIZE 32 OP_EQUALVERIFY  //length of the secret is 32 bytes
     OP_SHA256 2891f924fde4cc3c43af0d501a9fb52acb47b9a2e650c16ef0abb0a02c0ed988 OP_EQUALVERIFY // Sha256hash of the secret = hashsecret 
     OP_DUP OP_HASH160 0db229d573c1ca5042f1f6f8d95b0e48dd30f54c // combined with OP_EQUALVERIFY OP_CHECKSIG, checks if  Alice claims the output 
 OP_ELSE    //top of stack: False
    1521277464  // 48hours timestamp 
    OP_CHECKLOCKTIMEVERIFY  // check if the 48 hours have passed 
    OP_DROP    //pop the 48hours timestamp from the stack
    OP_DUP OP_HASH160 dbb79258a0200feeef593cc753e3c0c21757a130 // combined with OP_EQUALVERIFY OP_CHECKSIG, checks if Bob claims the output 
 OP_ENDIF 
 OP_EQUALVERIFY OP_CHECKSIG
```

 ### audit contract

Bob sendsAlice the contract and the contract transaction. Alice should now  verify if
- the script is correct 
- the locktime is far enough in the future
- the amount is correct
- she is the recipient 

 ```
$ btcatomicswap --testnet auditcontract 6382012088a8202891f924fde4cc3c43af0d501a9fb52acb47b9a2e650c16ef0abb0a02c0ed9888876a9140db229d573c1ca5042f1f6f8d95b0e48dd30f54c670418daac5ab17576a914dbb79258a0200feeef593cc753e3c0c21757a1306888ac 0200000000010106a85a242263768b81554b0ccab63ec124146014338ec962ef0197a1043b867e0c00000017160014ad430d1ed266d7130b5207a9bef00f8030947c3ffeffffff02204bbc000000000017a914bf09ed70b0c505d750f333bad8ca0520e48370fb87104772000000000017a914616aac51e9c239f6f85fe6db12249f264dd5ff2987024730440220441058ac56f1db678f955610adc264d7982da25dc7079a36c9022bc72827311c0220209d552d7e5ac04cef5ab615edea6c7836e7276ae7dbe0c2950eb44c781c7c7b01210335c272b2cfbd3c0a02bb38bfd859e4a44175545eb12ce93e7a2bac690068f92f00000000
Contract address:        2NAfLwhThYzB1kGYVxmjYqS98yXF8JBVc3v
Contract value:          0.1234 BTC
Recipient address:       mgmNWZN29WeFz3X4Na8thcbLB12JA5vj9j
Author's refund address: n1YiBr87yEeiiKnL29d5ZEBkNW9SGVNWXs

Secret hash: 2891f924fde4cc3c43af0d501a9fb52acb47b9a2e650c16ef0abb0a02c0ed988

Locktime: 2018-03-17 09:04:24 +0000 UTC
Locktime reached in 45h32m24s
```

WARNING:
A check on the blockchain should be done as the auditcontract does not do that so an already spent output could have been used as an input. Checking if the contract has been mined in a block should suffice
## References

Rivine atomic swaps are an implementation of [Decred atomic swaps](https://github.com/decred/atomicswap).

[Bitcoin scripts and opcodes](https://en.bitcoin.it/wiki/Script)