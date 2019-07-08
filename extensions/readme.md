# Rivine extensions

Gathered here are a set of extensions that implementing chains can reuse to add functionality or use them as inspiration to implement their own.

## Description

An extension can be registered to the consensus, it's a subscriber which can process consensus changes. When a block gets applied or reverted by the consensus, extensions are notified of this change. Most commonly an extension will process applied/reverted transactions that reside in a block. Every extensions should have their own transaction types which can be recognised by the consensus. This way an extension can know which transaction type it wants to process. If an extension has commands that can be executed by the client it should also provide client functionality, this is also applied for an exposed API.

## Available extensions

- [minting extension](./minting/readme.md)
- [auth coin transactions extension](./authcointx/README.md)
- [ERC20 extension](https://github.com/threefoldtech/rivine-extension-erc20/blob/master/README.md)

## Examples

Lets take [the minting extension's plugin](./minting/minting.go) as an example.
A plugin (extension) is a struct that needs to be defined with for example following properties:

```golang
type (
	// Plugin is a struct defines the minting plugin
	Plugin struct {
		genesisMintCondition 				types.UnlockConditionProxy
		minterDefinitionTransactionVersion 	types.TransactionVersion
		storage              				modules.PluginViewStorage
		unregisterCallback  				modules.PluginUnregisterCallback
	}
)
```

A plugin **has to** implement following interface:

```golang
// A ConsensusSetPlugin is an object that receives updates to the consensus set
	// every time there is a change in consensus. The difference with a ConsensusSetSubscriber
	// is that it stores its data within the database of the consensus set.
	ConsensusSetPlugin interface {
		// Initialize the bucket, could be creating it, migrating it,
		// or simply checking that all is as expected.
		// An error should be returned in case something went wrong.
		// metadata is nil in case the plugin wasn't registered prior to this attempt.
		// This method will be called while registering the plugin.
		InitPlugin(metadata *persist.Metadata, bucket *bolt.Bucket, ps PluginViewStorage, cb PluginUnregisterCallback) (persist.Metadata, error)

		// Apply the transaction to the plugin.
		// An error should be returned in case something went wrong.
		ApplyBlock(block types.Block, height types.BlockHeight, bucket *persist.LazyBoltBucket) error
		// Revert the block from the plugin.
		// An error should be returned in case something went wrong.
		RevertBlock(block types.Block, height types.BlockHeight, bucket *persist.LazyBoltBucket) error

		// Close releases any resources helt by the plugin like the PluginViewStorage
		Close() error
	}
```

### InitPlugin

InitPlugin inits the plugin with some metadata, a metadata bucket, a pluginViewStorage and a PluginUnregisteredCallback.
It creates the buckets that are designated to the plugin and optionally stores some default values.

* Metadata should consist with a header and a version. See [persist.Metatdata](https://godoc.org/github.com/threefoldtech/rivine/persist#Metadata).
* A Metadata bucket is the bucket where the metadata is stored.
* PluginViewStorage abstract the way we View whats inside an plugin's bucket.
* UnregisterCallback unregisters the plugin from the consensus when the consensus is closed.

### ApplyBlock

ApplyBlocks applies blocks processed by the consensus. In this method the extension has access to the entire applied block and can do a range of operations on it. For an example implementation look at [ApplyBlockMintingPlugin](https://github.com/threefoldtech/rivine/blob/75993ba4f451b08b970e593ba6c3e99d5fb492e9/extensions/minting/minting.go#L73). In this method the minting plugin looks for `Transactions` that applies to itself and stores these transactions data in the designated buckets.


### RevertBlock

RevertBlock reverts blocks processed by the consensus. In this method the extension has access to the entire reverted block and can do a range of operations on it. For an example implementation look at [ApplyBlockMintingPlugin](https://github.com/threefoldtech/rivine/blob/75993ba4f451b08b970e593ba6c3e99d5fb492e9/extensions/minting/minting.go#L105). In this method the minting plugin looks for `Transactions` that applies to itself and reverts these transactions data from the designated buckets.

### Other methods that are useful

* `NewPlugin()` creates a new plugin and registers the transaction types.
* `Close()` closes the plugin when the consensus is closed. This should release any rsources and call  `p.storage.close()`.
