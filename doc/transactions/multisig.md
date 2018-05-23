# Multisig

Since the 1.0.6 Release, Rivine supports Multisig, in the form of
[MultiSignatureConditions][multisigout] and [MultiSignatureFulfillments][multisigin].

Meaning you can spend an output and lock it using a [MultiSignatureCondition][multisigout],
targetting multiple unlockhashes (addresses), and spend such an (unspent) output
by collecting the signature-public-key pairs from you and your partners in a [MultiSignatureFulfillment][multisigin].

As part of a [MultiSignatureCondition][multisigout], you specify all addresses (unlockhashes)
which are allowed to create a signature, as well as the amount of signatures required.

Meaning you could create a [MultiSignatureCondition][multisigout] which defines 5 addresses, but of which only 3 out of 5 are required,
just as well as that you can create a [MultiSignatureCondition][multisigout] which defines 2 addresses of which both are required.

## Use Cases

Multisig allows you to create all kinds of wallets:

+ Shared Wallet, where at least one signature is required, multiple people own it, but not nececarly all people need to agree to complete a spendature of an unspent output;
+ Partner wallet, where everyone who owns it has to agree on a spendature of an unspent output;
+ Consensus Account, where more than the majority of owners has to agree to complete a spendature of an unspent output;
+ Split account, where only half of the owners has to agree to complete a spendature of an unspent output;

A party agrees by providing its signature of a transaction as part of a [MultiSignatureFulfillment][multisigin], as to be able to spend an unspent output.

A joint bank account would be an example of a Shared Wallet, as both you and your partner can spend money from that bank account, without needing the approval of the other.

Note that even though we talk about wallets here, it doesn't have to mean that you need to have a seperate wallet application for each such wallet.
It is perfectly acceptable, and even normal, to include all these different types of wallets, together with the personal (classic) wallet, in a single wallet.

## Flow Examples

Now that we know what multisig is all about, it is time to go over some examples as to show how this feature works exactly.

### Partner Wallet with 2 owners

In this example we have Bob and Alice, who own a partner wallet together.

This wallet is made up of the public keys of both Bob and Alice:

+ Bob uses public key `ed25519:c889b47f99912ae02d7a50c8a46ea116b2756ba79b41e5d62585691060fc5114`, which hashes to the unlock hash (address): `01e3c6f29e75b03351b3d72ad49ee427aa60d32a6d06f6b1c2d60e73c9438e7b927dd7817ad27c`
+ Alice uses public key `ed25519:c12abbf9f99b90c0bae737eeded8f6ae5c9a4312c4636d5ddeebf11705ce9498`, which hashes to the unlock hash (address): `01cc1872da1c5b2f6bc02fead5f660992477b7c3d7133c75746b7adeec72bdda5c9149cb36e34a`

Anyone who wishes to send coins to this partner wallet of Alice and Bob would do so by defining following coin output:

```javascript
{
    "value": 1000, // value in the lowest-unit of coins
    "condition": {
        "type": 4,
        "data": {
            "unlockhashes": [ // order does not matter
                "01e3c6f29e75b03351b3d72ad49ee427aa60d32a6d06f6b1c2d60e73c9438e7b927dd7817ad27c", // unlock hash (address) owned by Bob
                "01cc1872da1c5b2f6bc02fead5f660992477b7c3d7133c75746b7adeec72bdda5c9149cb36e34a", // unlock hash (address) owned by Alice
            ]
            "minimumsignaturecount": 2
        }
    }
}
```

Anyone could also send, using the same exact output, but requiring just one signature by defining `"minimumsignaturecount": 1`.
This would still send it to a wallet owned by both Alice and Bob, but to a shared wallet, not the partner wallet.
The key difference being, as explained in [the previous section](#use-cases), that in the latter wallet type Bob and Alice have to approve together, while in the other one not.

> See the [MultiSignatureFulfillment JSON-encoding documentation][multisigin] to see how the fulfillment would look like, used to fulfill such a [MultiSignatureCondition][multisigout].

All unspent (coin) outputs submitted to the blockchain in this way, will be both visible to Bob and Alice.

Let's say they want to spend such an output:

1. A transaction will have to be created by either one of them, where the desired outputID is used as parentID of an input, spend by defining the desired outputs;
2. That raw transaction will have to be signed by both partners, signed as in adding the signature paired with the correct public key as part of the [MultiSignatureFulfillment][multisigin];
3. The raw transaction has to be signed, and completed, submitted to the transaction pool of a daemon, such that it can be propagated over the network and submitted as part of a block ASAP;

There are multiple ways to achieve this, but one simple flow could be that:

+ (1) Bob creates the transaction:
  + defining an input, where the ID of the unspent output is used as the parentID and an empty [MultiSignatureFulfillment][multisigin] is used as the fulfillment;
  + defining the agreed upon output(s);
  + defining the required Miner Fee, and ensuring that `sum(inputs)=sum(outputs)+sum(minerfees)`;
  + defining any optional arbitrary data, should this be desired by both parties;
+ (2) Bob signs this raw transaction using his wallet;
+ (3) Bob sends the prepared and signed raw transaction to Alice;
+ (4) Alice signs the prepared transaction, already signed by Bob, as well;
+ (5) Alice submits the transaction to the transaction pool of a daemon (`POST /transactionpool/transaction`);

Note that step (5) in this flow can also be replaced by
the action of Alice sending the transaction where both she and bob signed back to Bob,
such that Bob submits that transaction to the transaction pool of a daemon instead.

> Important: In the above flow, Bob signs the raw transaction prior to sending it to Alice. As a result, the transaction cannot be altered, given that a signature of Bob is required and thus any change to the transaction would make his signature invalid. By signing prior to sending it to Alice, he can be sure that the received transactions, co-signed by Alice, will not be altered by her (or a malicious 3rd party, frequently known as Eve).

Using the `rivinec` binary CLI client,
the 5-step flow from above could be represented in bash as follows:

```bash
# create a transaction and sign it
$ TXN="$(rivinec create cointransaction \
    97495f5c40d392046bd45c27acc860c6a93581930a735e0990a1e42a05cbe55e \
    01907fef3ba1c3905021ae2d1486adf9bc8721821229a8a858f567b7303a26dfba454db47fa71d 1000 | \
    rivinec wallet sign)"

# send signed raw transaction to our partner
# so that they can sign the same way:
$ rivinec wallet sign "$TXN"

# either they submit it, or they send it back to us so we submit the transaction,
# whatever the case, it will be submitted using following command:
$ rivinec wallet send "$TXN"
```

### Partner wallet with 3 owners

When a wallet has 3 owners, of which all 3 have to agree, the 5-step flow described in [the previous example](#partner-wallet-with-2-owners) could be exended to more than 2 parties as well, and thus this example.

Let's say we have Bob, Alice and Carlos who own such a partner wallet. And let's say that Bob and Alice use the same public keys as in our previous example, while Carlos uses the public key `ed25519:d1caa9f04349121769ae96f65b722e58d5d85fc0f57dfd636ccddd6118c62b62` which hashes to `018b17bb8a31d94d26a1202f1c3c07bbcad61164731d38b500b08b1218126791ad97ca568727ee`. This would mean that people can send coins to this example partner wallet by giving a coin output such as:

```javascript
{
    "value": 1000, // value in the lowest-unit of coins
    "condition": {
        "type": 4,
        "data": {
            "unlockhashes": [ // order does not matter
                "01e3c6f29e75b03351b3d72ad49ee427aa60d32a6d06f6b1c2d60e73c9438e7b927dd7817ad27c", // unlock hash (address) owned by Bob
                "01cc1872da1c5b2f6bc02fead5f660992477b7c3d7133c75746b7adeec72bdda5c9149cb36e34a", // unlock hash (address) owned by Alice
                "018b17bb8a31d94d26a1202f1c3c07bbcad61164731d38b500b08b1218126791ad97ca568727ee", // unlock hash (address) owned by Carlos
            ]
            "minimumsignaturecount": 3
        }
    }
}
```

And we can extent the simple 5-step flow used in [the previous example](#partner-wallet-with-2-owners) to a 7-step flow:

+ (1) Bob creates the transaction:
  + defining an input, where the ID of the unspent output is used as the parentID and an empty [MultiSignatureFulfillment][multisigin] is used as the fulfillment;
  + defining the agreed upon output(s);
  + defining the required Miner Fee, and ensuring that `sum(inputs)=sum(outputs)+sum(minerfees)`;
  + defining any optional arbitrary data, should this be desired by both parties;
+ (2) Bob signs this raw transaction using his wallet;
+ (3) Bob sends the prepared and signed raw transaction to Alice;
+ (4) Alice signs the prepared transaction, already signed by Bob, as well;
+ (5) Alice sends the prepared raw transaction, signed by both Bob and Alice, to Carlos;
+ (6) Carlos signs the prepared transaction, signed by both Bob and Alice, as well;
+ (7) Carlos submits the transaction to the transaction pool of a daemon (`POST /transactionpool/transaction`);

Just as in the 5-step flow described [the previous example](#partner-wallet-with-2-owners), it doesn't have to be Carlos who sends/submits the transaction, this can be done by Bob and Alice as well. In fact, it can be done by anyone, as long as all required signatures are part of that transaction already.

Should it be desired for whatever reason, we can also turn this serial 7-step flow into a parallel flow:

+ (1) Bob creates the transaction:
  + defining an input, where the ID of the unspent output is used as the parentID and an empty [MultiSignatureFulfillment][multisigin] is used as the fulfillment;
  + defining the agreed upon output(s);
  + defining the required Miner Fee, and ensuring that `sum(inputs)=sum(outputs)+sum(minerfees)`;
  + defining any optional arbitrary data, should this be desired by both parties;
+ (2) Bob signs this raw transaction using his wallet;
+ (3) Bob sends the prepared and signed raw transaction to both Alice and Carlos;
+ (3) Alice and Carlos sign the transaction and send the transaction, signed by them and Bob, back to Bob;
+ (4) Bob merges the input(s) of the 2 signed transactions (one transaction signed by Bob+Alice and one transaction signed by Bob+Carlos), received back from Alice and Carlos, into a single transaction;
+ (5) Bob submits it to the transaction pool of a daemon;

> Important: In the above flow, Bob signs the raw transaction prior to sending it to Alice and Carlos. As a result, the transaction cannot be altered, given that a signature of Bob is required and thus any change to the transaction would make his signature invalid. By signing prior to sending it to Alice and Carlos, he can be sure that the received transactions, signed by them, will not be altered by them (or a malicious 3rd party, frequently known as Eve).

Using the `rivinec` binary CLI client,
this parallel flow from above could be represented in bash as follows:

```bash
# create a transaction
$ TXN="$(rivinec create cointransaction \
    97495f5c40d392046bd45c27acc860c6a93581930a735e0990a1e42a05cbe55e \
    01907fef3ba1c3905021ae2d1486adf9bc8721821229a8a858f567b7303a26dfba454db47fa71d 1000 | \
    rivinec wallet sign)"

# send unsigned raw transaction to Alice and Carlos,
# so they can sign it each and send the signed transaction back to Bob:
$ TXN_ALICE="$(rivinec wallet sign "$TXN")" | sendto Bob
$ TXN_CARLOS="$(rivinec wallet sign "$TXN")" | sendto Bob
# (sendto is a fictional command)

# bob merges the transactions and submits the end result
$ rivinec merge inputs "$TXN_ALICE" "$TXN_CARLOS" | rivinec wallet send
```

[multisigout]: /doc/transactions/transaction.md#json-encoding-of-a-multisignaturecondition
[multisigin]: /doc/transactions/transaction.md#json-encoding-of-a-multisignaturefulfillment
