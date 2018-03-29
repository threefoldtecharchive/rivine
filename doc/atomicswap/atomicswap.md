# Cross chain atomic swapping

# Theory

A cross-chain swap is a trade between two users of different cryptocurrencies. For example, one party may send Rivine-based tokens to a second party's Rivine-based address, while the second party would send Bitcoin to the first party's Bitcoin address. However, as the blockchains are unrelated and transactions cannot be reversed, this provides no protection against one of the parties not honoring their end of the deal. One common solution to this problem is to introduce a mutually-trusted third party for escrow. An atomic cross-chain swap solves this problem without the need for a third party. On top of that it achieves waterproof validation without introducing the problems and complexities introduced by a escrow-based validation system.

Atomic swaps involve each party paying into a double-contract transaction, one contract for each blockchain. The contracts contain an output that is spendable by either party, but the rules required for redemption are different for each party involved. The validation of the swapping process happens by the parties themselves. The entire process is completely secure and trustless. It is made such that both parties either walk away with the money of the other, or get both their invested money back. It is not possible at any point that one party walks away with all money, or that money gets in a permanent deadlock situation disallowing any of the two parties to touch it.

# Example

Let's assume Bob wants to buy 9876 Rivine based tokens from Alice for 0.1234BTC, and they agree on both this exchange price and the fact that they want to trade using an atomic swap.

Bob creates a bitcoin address and Alice creates a Rivine address.

Bob initiates the swap, he generates a (32-byte) secret and hashes it
using the SHA256 algorithm, resulting in a 32-byte hashed secret.

Bob now creates a swap transaction, as a smart contract, and publishes it on the bitcoin chain, it has 0.1234BTC as an output and the output can be redeemed (used as input) using either 1 of the following conditions:
- timeout has passed (48hours) and claimed by Bob's refund address;
- the money is claimed by Alice's registered address, prior to that 48 hour deadline, by providing the secret received from bob on a side channel (e.g. an encrypted chat/messenging service);

It should be noted that at any given point, only (and only one) one party can claim the money, given their conditions are fulfilled. This means Alice can claim the bitcoin if she has the secret. Should Alice fail to claim the locked BTC before the 48 hours deadline, Bob can reclaim it whenever he wants.
 
Once Bob created the contract, using a 48 hour timeout and correct hashed secret, he has to send the transaction ID, of the created bitcoin transaction, to Alice. At this point he should only share the transaction ID, NOT the secret. This is important as otherwise Alice can claim the money already, without having to her promised Rivine-based tokens.

Upon the receival of the transaction ID, over the used side channel, Alice can validate (=audit) if everything is agreed. Meaning that she can validate if the Bitcoin transaction exists and whethet it contains the expected contract with all information being correct. After validating and agreeing what has happened so far, she will have to create a similar contract, but on the Rivine-based blockchain, not on Bitcoin. She will use a 24 hour timeout, not a 48 hour timeout, and use the same hashed secret as used by Bob in the first contract. This transaction has 9876 Rivine-based tokens as an output and the output can be redeemed (used as input) using either 1 of the following conditions:
- timeout has passed (24 hours) and claimed by Alice's refund address;
- the money is claimed by Bob's registered address, prior to that 24 hour deadline, by providing the secret he choose himself when creating the Bitcoin contract;

This means that Bob has to claim the Rivine tokens prior to the 24 hours deadline. If he fails to do so, Alice can reclaim her tokens whenever she wants. In order for Bob to claim the tokens, he has to use and as such disclose the secret choosen by him. Having made that secret public, by claiming his tokens, Alice can use this secret as well to claim her promised bitcoins in the bitcoin contract created by Bob. It should be noted however that if Alice does not claim this money prior to the 48 hour deadline, she will not be able to receive that money, and Bob will have both the BTC as well as the Rivine-based tokens. When both timeouts are configured reasonable however, it should mean that Alice has at least 24 hours to do so, and usually even closer to 48 hours.

The magic of the atomic swap lies in the fact that the same secret is used to claim the tokens in both swap transactions (contracts) but it is not disclosed in the contracts because only the hash of the secret is used there. Therefore only the hashed secret is public, while the secret should remain private up until the point that Bob discloses it, by claiming the Rivine-based tokens. The moment Bob claims those, he discloses the secret and Alice still has renough time to claim the bitcoin because the timeout of the first contract is longer than the one of the second contract

Of course, either Bob or Alice can be the initiator or the participant.

## Technical details of the example

secret: 32 bytes

hashsecret: Sha256( secret)

To Start bitcoin core qt in server mode on testnet: 
` ./Bitcoin-Qt  -testnet -server -rpcuser=user -rpcpassword=pass-rpcport=18332`


Alice creates a new bitcoin  address (as of bitcoin core 0.16, make sure to specify the 'legacy' address type since we need a p2pkh address): 
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
You can check the transaction [on a bitcoin testnet blockexplorer](https://testnet.blockexplorer.com/tx/afbc4dc719d9f79a9413945c92752bef644c618a4362fc8e8be0764a1b888e10) where you can see that 0.1234 BTC is sent to2NAfLwhThYzB1kGYVxmjYqS98yXF8JBVc3v (= the contract script hash) being a [p2sh](https://en.bitcoin.it/wiki/Pay_to_script_hash) address in the bitcoin testnet. 

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

Bob sends Alice the contract and the contract transaction. Alice should now verify if
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

### Participate

Alice trusts the contract and so she participates in the atomic swap by paying the tokens into a Rivine contract using the same secret hash. The refund transaction can not be sent until the locktime expires, but should be saved in case a refund is necessary.

```
$rivinec atomicswap --testnet participate bb6e12437c6fecbe83f5bf3724ced0369c01e166364cc320adf166125a8b6e2c756ada1be3f6 98765 2891f924fde4cc3c43af0d501a9fb52acb47b9a2e650c16ef0abb0a02c0ed988
abcdsdsfsfsefsese02033r0sdfsf03a # coinOutputID
``` 

The above command will create a transaction with `98765` as the CoinOutput, and an atomic script very similar to the bitcoinscript earlier. The receiver is registered under public key `2891f924fde4cc3c43af0d501a9fb52acb47b9a2e650c16ef0abb0a02c0ed988` and the receiver will have to also proof the ownership of the secret that can get hash into the hashed_secret `\x86\x10\t\xecMY\x9f\xab\x1f@\xab\xc7no\x89\x88\f\xffX3ǜT\x8c\x99\xf9\x04_\x19\x1c\xd9\v`.

Alice now informs Bob that the threefold contract transaction has been created  and provides him with the contract details.

### audit rivine contract

Just as Alice had to audit Bob's contract, Bob now has to do the same with Alice's contract before withdrawing. 
Bob verifies :
- the needed coinOutput (ID: `abcdsdsfsfsefsese02033r0sdfsf03a`) exists and has a script attached
- the amount of tokens (coins) defined in the coinOutput is correct
- the attached script is correct
- the locktime, hashed secret (`2891f924fde4cc3c43af0d501a9fb52acb47b9a2e650c16ef0abb0a02c0ed988`) and public key (wallet addr), defined in the attached script, are correct
```
$rivinec atomicswap --testnet audit abcdsdsfsfsefsese02033r0sdfsf03a \x86\x10\t\xecMY\x9f\xab\x1f@\xab\xc7no\x89\x88\f\xffX3ǜT\x8c\x99\xf9\x04_\x19\x1c\xd9\v
```

WARNING:
The audit should also ensure that the given `coinOutput` has not already been used as a `coinInput`.

### redeem tokens

Now that both Bob and Alice have paid into their respective contracts, Bob may withdraw from the rivine contract. This step involves publishing a transaction which reveals the secret to Alice, allowing her to withdraw from the Bitcoin contract.

```
$rivinec atomicswap --testnet redeem abcdsdsfsfsefsese02033r0sdfsf03a \x86\x10\t\xecMY\x9f\xab\x1f@\xab\xc7no\x89\x88\f\xffX3ǜT\x8c\x99\xf9\x04_\x19\x1c\xd9\v 01234567890123456789012345678901
```

Note that redeeming tokens will firt validate (audit) the used rivine contract (unless specified otherwise), and thus the redeemer can in fact skip the manual audit step. `--no-audit` can be used in case no extra `audit` is required.

### redeem bitcoins

Now that Bob has withdrawn from the rivine contract and revealed the secret, Alice must extract the secret from this redemption transaction. Alice may watch a block explorer to see when the rivine contract output was spent and look up the redeeming transaction.
Alice can now extract the secret:
```
$$ rivatomicswap --testnet extractsecret 000000000118d94f38b8532bfe78bda0d0848a7965bdfbe6e88476896f01318717bc7e1aa50100000000ffffffff01885ef5050000000000001976a9149551ab760ba64b7e573f54d34c53506676e8145888ace6dabb590000000001ffffffffffffffff00000000ffffffffe0483045022100a1a3b37a67f3ed5d6445a0312e825299b54d91a09e0d1b59b5c0a8baa7c0642102201a0d53e9efe7db8dc47210b446fde6425be82761252ff0ebe620efc183788d86012103395a4a3c8c96ef5e5af6fd80ae42486b5d3d860bf3b41dafc415354de8c7ad80203e0b064c97247732a3b345ce7b2a835d928623cb2871c26db4c2539a38e61a16514c5163a61429c36b8dd380e0426bdc1d834e74a630bfd5d1118876a9149ee19833332a04d2be97b5c99c970191221c070c6704e6dabb59b17576a914b0ec0640c89cf803b8fdbd6e0183c354f71748c46888ac 29c36b8dd380e0426bdc1d834e74a630bfd5d111
dsadadadadad322r3rsfsdf123r2fsdf
Secret: d685a0b8aacf03f024c84092b4b951d1e54f7747a72cd21ed091f16996502a8e
```

With the secret known (extracted from the coinInput with ID `dsadadadadad322r3rsfsdf123r2fsdf`), Alice may redeem from Bob's Bitcoin contract:
```
$ btcatomicswap --testnet --rpcuser=user --rpcpass=pass redeem 6382012088a8202891f924fde4cc3c43af0d501a9fb52acb47b9a2e650c16ef0abb0a02c0ed9888876a9140db229d573c1ca5042f1f6f8d95b0e48dd30f54c670418daac5ab17576a914dbb79258a0200feeef593cc753e3c0c21757a1306888ac 0200000000010106a85a242263768b81554b0ccab63ec124146014338ec962ef0197a1043b867e0c00000017160014ad430d1ed266d7130b5207a9bef00f8030947c3ffeffffff02204bbc000000000017a914bf09ed70b0c505d750f333bad8ca0520e48370fb87104772000000000017a914616aac51e9c239f6f85fe6db12249f264dd5ff2987024730440220441058ac56f1db678f955610adc264d7982da25dc7079a36c9022bc72827311c0220209d552d7e5ac04cef5ab615edea6c7836e7276ae7dbe0c2950eb44c781c7c7b01210335c272b2cfbd3c0a02bb38bfd859e4a44175545eb12ce93e7a2bac690068f92f00000000 d685a0b8aacf03f024c84092b4b951d1e54f7747a72cd21ed091f16996502a8e
Redeem fee: 0.00002499 BTC (0.00007713 BTC/kB)

Redeem transaction (71775d49f8032a7e326b9ca04a3a2ba2f5661a877a187e1346cd21ac55e43910):
0200000001108e881b4a76e08b8efc62438a614c64ef2b75925c9413949af7d919c74dbcaf00000000ef4730440220500da27a6a46f99f7b96fc83c49f9b4207aae3433e971d9d21eb17267e565a5702204841ed1db53763384f661505fe230b876d2bce2c1b785fc457236602fc9a9b36012102696019f19198a3bbc4b774b81de7468e77301d5f836953d83831c2589ed19cbd20d685a0b8aacf03f024c84092b4b951d1e54f7747a72cd21ed091f16996502a8e514c616382012088a8202891f924fde4cc3c43af0d501a9fb52acb47b9a2e650c16ef0abb0a02c0ed9888876a9140db229d573c1ca5042f1f6f8d95b0e48dd30f54c670418daac5ab17576a914dbb79258a0200feeef593cc753e3c0c21757a1306888acffffffff015d41bc00000000001976a914930c8178dec2519a18e0609e766fa881acab582d88ac18daac5a

Publish redeem transaction? [y/N] y
Published redeem transaction (71775d49f8032a7e326b9ca04a3a2ba2f5661a877a187e1346cd21ac55e43910)
```
decoding the redeem transaction gives 
```
{
  "txid": "71775d49f8032a7e326b9ca04a3a2ba2f5661a877a187e1346cd21ac55e43910",
  "hash": "71775d49f8032a7e326b9ca04a3a2ba2f5661a877a187e1346cd21ac55e43910",
  "version": 2,
  "size": 324,
  "vsize": 324,
  "locktime": 1521277464,
  "vin": [
    {
      "txid": "afbc4dc719d9f79a9413945c92752bef644c618a4362fc8e8be0764a1b888e10",
      "vout": 0,
      "scriptSig": {
        "asm": "30440220500da27a6a46f99f7b96fc83c49f9b4207aae3433e971d9d21eb17267e565a5702204841ed1db53763384f661505fe230b876d2bce2c1b785fc457236602fc9a9b36[ALL] 02696019f19198a3bbc4b774b81de7468e77301d5f836953d83831c2589ed19cbd d685a0b8aacf03f024c84092b4b951d1e54f7747a72cd21ed091f16996502a8e 1 6382012088a8202891f924fde4cc3c43af0d501a9fb52acb47b9a2e650c16ef0abb0a02c0ed9888876a9140db229d573c1ca5042f1f6f8d95b0e48dd30f54c670418daac5ab17576a914dbb79258a0200feeef593cc753e3c0c21757a1306888ac",
        "hex": "4730440220500da27a6a46f99f7b96fc83c49f9b4207aae3433e971d9d21eb17267e565a5702204841ed1db53763384f661505fe230b876d2bce2c1b785fc457236602fc9a9b36012102696019f19198a3bbc4b774b81de7468e77301d5f836953d83831c2589ed19cbd20d685a0b8aacf03f024c84092b4b951d1e54f7747a72cd21ed091f16996502a8e514c616382012088a8202891f924fde4cc3c43af0d501a9fb52acb47b9a2e650c16ef0abb0a02c0ed9888876a9140db229d573c1ca5042f1f6f8d95b0e48dd30f54c670418daac5ab17576a914dbb79258a0200feeef593cc753e3c0c21757a1306888ac"
      },
      "sequence": 4294967295
    }
  ],
  "vout": [
    {
      "value": 0.12337501,
      "n": 0,
      "scriptPubKey": {
        "asm": "OP_DUP OP_HASH160 930c8178dec2519a18e0609e766fa881acab582d OP_EQUALVERIFY OP_CHECKSIG",
        "hex": "76a914930c8178dec2519a18e0609e766fa881acab582d88ac",
        "reqSigs": 1,
        "type": "pubkeyhash",
        "addresses": [
          "mtvUcAgzLLAWfjPPxkCu5vy7B67GmKJRuo"
        ]
      }
    }
  ]
}
```
In the script signature to unlock the input, you can recognize the secret `d685a0b8aacf03f024c84092b4b951d1e54f7747a72cd21ed091f16996502a8e` and the transaction script `4730440220500da27a6a46f99f7b96fc83c49f9b4207aae3433e971d9d21eb17267e565a5702204841ed1db53763384f661505fe230b876d2bce2c1b785fc457236602fc9a9b36012102696019f19198a3bbc4b774b81de7468e77301d5f836953d83831c2589ed19cbd20d685a0b8aacf03f024c84092b4b951d1e54f7747a72cd21ed091f16996502a8e514c616382012088a8202891f924fde4cc3c43af0d501a9fb52acb47b9a2e650c16ef0abb0a02c0ed9888876a9140db229d573c1ca5042f1f6f8d95b0e48dd30f54c670418daac5ab17576a914dbb79258a0200feeef593cc753e3c0c21757a1306888ac`"
This transaction can be verified [on a bitcoin testnet blockexplorer](https://testnet.blockexplorer.com/tx/71775d49f8032a7e326b9ca04a3a2ba2f5661a877a187e1346cd21ac55e43910) .
The cross-chain atomic swap is now completed and successful.

## References

Rivine atomic swaps are an implementation of [Decred atomic swaps](https://github.com/decred/atomicswap).

[Bitcoin scripts and opcodes](https://en.bitcoin.it/wiki/Script)