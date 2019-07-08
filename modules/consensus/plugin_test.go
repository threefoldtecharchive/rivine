package consensus

import (
	"context"
	"testing"

	bolt "github.com/rivine/bbolt"
	"github.com/stretchr/testify/assert"
	"github.com/threefoldtech/rivine/modules"
	"github.com/threefoldtech/rivine/persist"
	"github.com/threefoldtech/rivine/types"
)

type testPlugin struct {
	closeCalled bool

	storage modules.PluginViewStorage
}

func (plugin *testPlugin) Close() error {
	plugin.storage.Close()
	plugin.closeCalled = true
	return nil
}
func (plugin *testPlugin) InitPlugin(metadata *persist.Metadata, bucket *bolt.Bucket, ps modules.PluginViewStorage, cb modules.PluginUnregisterCallback) (persist.Metadata, error) {
	plugin.storage = ps
	metadata = &persist.Metadata{
		Version: "1.0.0.0",
		Header:  "testPlugin",
	}
	return *metadata, nil
}
func (plugin *testPlugin) ApplyBlock(block types.Block, height types.BlockHeight, bucket *persist.LazyBoltBucket) error {
	return nil
}
func (plugin *testPlugin) RevertBlock(block types.Block, height types.BlockHeight, bucket *persist.LazyBoltBucket) error {
	return nil
}

// Apply the transaction to the plugin.
// An error should be returned in case something went wrong.
func (plugin *testPlugin) ApplyTransaction(txn types.Transaction, block types.Block, height types.BlockHeight, bucket *persist.LazyBoltBucket) error {
	return nil
}

// Revert the transaction from the plugin.
// An error should be returned in case something went wrong.
func (plugin *testPlugin) RevertTransaction(txn types.Transaction, block types.Block, height types.BlockHeight, bucket *persist.LazyBoltBucket) error {
	return nil
}

func TestPluginCloseCalled(t *testing.T) {
	cst, err := blankConsensusSetTester("testconsensus")
	if err != nil {
		t.Fatal(err)
	}
	var plugin testPlugin

	cst.cs.RegisterPlugin(context.Background(), "testplugin", &plugin)
	cst.cs.Close()
	assert.True(t, plugin.closeCalled, "Closing the consensus should call close on the registered plugins")
}
