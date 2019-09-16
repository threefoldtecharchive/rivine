# UnlockCondition

Coins and blockstakes active on the chain (see: not destroyed somehow) are bound to an unlock condition.
Simply put, the condition defines who can spend coins/blockstakes by providing the matching fulfillment.


## Types of unlock conditions

In the Go implementation all unlock conditions adhere to the [UnlockCondition](https://godoc.org/github.com/threefoldtech/rivine/types#UnlockCondition) interface.

### UnlockhashCondition

An [UnlockHashCondition](https://godoc.org/github.com/threefoldtech/rivine/types#UnlockHashCondition) specifies the target unlockhash that can spend the
paired coins/blockstakes. You can read more about unlockhashes in the [./unclockhash.md](./unlockhash.md) documentation.

Only [the Public Key Unlock Hash](./unlockhash.md#the-public-key-unlock-hash) is allowed to be used for this condition,
meaning that assets paired with an [UnlockHashCondition](https://godoc.org/github.com/threefoldtech/rivine/types#UnlockHashCondition) can be spend by
the owner of the matching private (ED25519) signature key.

This is the most common type of unlock condition, and when we talk about wallets we are really talking about
some piece of software/hardware in possession of this private key, paired to the public key of which the unlockhash (synonym for address) is generated.

### MultiSignatureCondition

A [MultiSignatureCondition](https://godoc.org/github.com/threefoldtech/rivine/types#MultiSignatureCondition) specifies the target unlockhashes that can spend the
paired coins/blockstakes on the condition that they collectively sign with a sufficient amount of paired private keys.

The minimum amount of signatures required are specified as part of the condition and together with the unlockhashes authorized to sign.
When fulfilling the condition, in order to spend the assets, each unlockhash can only provide one signature.
It is allowed that more signatures are given than required.


### TimeLockCondition

A [TimeLockCondition](https://godoc.org/github.com/threefoldtech/rivine/types#TimeLockCondition) is a wrapping condition,
using internally either an [UnlockhashCondition](#UnlockhashCondition) or a [MultiSignatureCondition](#MultiSignatureCondition).

Prior to being able to fulfill the internal condition, a certain time or blockheight has to be reached on the active chain as specified.

### AtomicSwapCondition

An [AtomicSwapCondition](https://godoc.org/github.com/threefoldtech/rivine/types#AtomicSwapCondition) is the creation of an atomic swap contract.

See [../atomicswap/atomicswap.md](../atomicswap/atomicswap.md) and [../atomicswap/technical details.md](../atomicswap/technical%20details.md) for more information.
