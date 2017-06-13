# Transactions

## Generate address from master seed

![Address Generation](addressgeneration.png)

## Send coins

Send coins from 0x78e6... to 0x1d64..

![Send coins](send.png)

When spending money you need to add a coininput in a transaction, info needed in coininput is:
1. ParentID is the ID of the unspend coinoutput (UTXO)
2. Recalculate the unlockconditions that generates the unlockhash (normally they are standard and can be reconstructed)
3. Use the corresponding private key to sign a transactionSignature.
4. The money that can be spent in the transaction is found in the corresponding coinoutput. ex. Value = 2000

Per transaction: The sum of all values in coinoutput should be less than the sum of all "unlocked" values in the coininput. The difference is minerfee for the the block generator.

## Arbitrary data

![Arbitrary data](arbitrarydata.png)

Arbitrary Data can be of any size. There is however a size limit on a Transaction and a block.
Keep in mind that the minerfee is depending on the size of a transaction, a blockcreator can ignore to add a transaction with small fees for a lot of small transactions with a summed up bigger fee (opportune).
