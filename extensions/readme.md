# Rivine extensions

Gathered here are a set of extensions that implementing chains can reuse to add functionality or use them as inspiration to implement their own.

## Description

An extension can be registered to the consensus, when a block is created or reverted some transactions might be applied or reverted. Every extensions should have their own transaction types which can be recognised by the consensus. If an extension has commands that can be executed by the client it should also provide some client functionality.

## Examples

Lets take the minting extension as an example. A plugin (extensions) is a struct with following properties:

```golang
type (
	// Plugin is a struct defines the minting plugin
	Plugin struct {
		genesisMintCondition types.UnlockConditionProxy
		storage              modules.PluginViewStorage
		unregisterCallback   modules.PluginUnregisterCallback
	}
)
```

A plugin should implement following interface:

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
	}
```

### InitPlugin

InitPlugin inits the plugin with some metadata, a metadata bucket, a pluginViewStorage and a PluginUnregisteredCallback.
It creates the buckets that are designated to the plugin and optionally stores some default values.

* Metadata should consist with a header and a version. See `persist.Metadata`.
* Metadata bucket is the bucket where the metadata is stored
* PluginViewStorage abstract the way we View whats inside an plugin's bucket.
* UnregisterCallback unregisters the plugin from the consensus when the consensus is closed

### ApplyBlock

ApplyBlocks applies transactions that are in applied blocks processed by the consensus. In this method a plugins looks for `Transactions` that applies to itself and stores these transactions data in the designated buckets. For an example implementation look at [ApplyBlockMintingPlugin](https://github.com/threefoldtech/rivine/blob/75993ba4f451b08b970e593ba6c3e99d5fb492e9/extensions/minting/minting.go#L73)


### RevertBlock

RevertBlock reverts transactions that are in reverted blocks processed by the consensus. In this method a plugins looks for `Transactions` that applies to itself and reverts these transactions data from the designated buckets. For an example implementation look at [ApplyBlockMintingPlugin](https://github.com/threefoldtech/rivine/blob/75993ba4f451b08b970e593ba6c3e99d5fb492e9/extensions/minting/minting.go#L105)

### Other methods that are usefull

* `NewPlugin()` creates a new plugin and registers the transaction types.
* `Close()` closes the plugin when the consensus is closed. This should the `unregisterCallback` and `p.storage.close()`.


## Example usage of minting plugin

How to use the minting plugin can be found [here](/extensions/minting/readme.md)