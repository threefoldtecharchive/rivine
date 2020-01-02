# Output selection

While many strategies exist to fund a transaction, each with their (dis)advantages, the reference implementation chooses a conservative one in order to prevent the disability to fund a transaction while enough funds are available.

The reason is that other strategies might split the outputs in many very small ones so they become unusable to fund a larger transaction.

## algorithm

First collect all available outputs.

Secondly try to fund using the lowest (up to the highest) confirmed outputs available.
If we reach the limit and still didn't fund enough, try to replace from smallest to highest with the remaining available unconfirmed outputs.

Lastly if we still couldn't fund, or perhaps we never reached the limit, we'll try to
fund using unconfirmed, starting from highest to lowest, and overwriting as long as we have higher ones available to overwrite lower ones with.

If this last step failed than funding is simply not possible.

### Maximum number of coin inputs per transaction

If you consult the code or documentation you'll be able to conclude that a v1 transaction
with no coin inputs, one coin output, no arbitrary data, no block stake inputs or outputs and one miner fee (minimum),
has a size of 358 bytes when Sia-binary encoded (the encoding used for v1 transactions).

Each added coin input will add 169 bytes to the transaction (it's as linear as this, given the length of coin inputs is a static 8 bytes,
already included in the 358 bytes mentioned earlier).

The block size limit is in tfchain (and all other rivine chains) 2e6 bytes, so there won't be an issue there.

In the transaction pool module however there is another constant adhered by that module,
which limits the max size a single transaction is allowed to have. This is 16e3 in tfchain (and all other rivine chains).

We can therefore conclude that for a merge transaction we can put
⌊(16e3 - 358) / 169⌋ = 92 coin inputs per transaction.

If we have two coin outputs (because we also have a refund coin output) we can instead only have up to:

⌊(16e3 - 409) / 169⌋ = 92 coin inputs per transaction.

This consistency is good, as it means we can as a measurement already add a check in our high level wallets
that checks that there are no more than x coin inputs per transaction, irrelevant if there are one or two coin outputs,
(more coin outputs are no used in regular high level wallet transactions, and less neither.

They will however also have to take into account the byte size of the arbitrary data, if for example 83 bytes are used for
arbitrary data (the maximum) than we can have:

⌊(16e3 - 492) / 169⌋ = 91 coin inputs per transaction. It is however simple enough to check this extra size on the fly.

17 bytes have to be added in case a LockTime is used for the first coin output.