package consensus

import (
	"context"
	"testing"

	bolt "github.com/rivine/bbolt"
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
func (plugin *testPlugin) ApplyBlock(block modules.ConsensusBlock, bucket *persist.LazyBoltBucket) error {
	return nil
}
func (plugin *testPlugin) RevertBlock(block modules.ConsensusBlock, bucket *persist.LazyBoltBucket) error {
	return nil
}

func (plugin *testPlugin) ApplyBlockHeader(block modules.ConsensusBlockHeader, bucket *persist.LazyBoltBucket) error {
	return nil
}
func (plugin *testPlugin) RevertBlockHeader(block modules.ConsensusBlockHeader, bucket *persist.LazyBoltBucket) error {
	return nil
}

// Apply the transaction to the plugin.
// An error should be returned in case something went wrong.
func (plugin *testPlugin) ApplyTransaction(txn modules.ConsensusTransaction, bucket *persist.LazyBoltBucket) error {
	return nil
}

// Revert the transaction from the plugin.
// An error should be returned in case something went wrong.
func (plugin *testPlugin) RevertTransaction(txn modules.ConsensusTransaction, bucket *persist.LazyBoltBucket) error {
	return nil
}

// TransactionValidatorFunctions allows the plugin to provide validation rules for all transaction versions it mapped to
func (plugin *testPlugin) TransactionValidatorVersionFunctionMapping() map[types.TransactionVersion][]modules.PluginTransactionValidationFunction {
	return nil
}

// TransactionValidators allows the plugin to provide validation rules for all transactions versions it wants
func (plugin *testPlugin) TransactionValidators() []modules.PluginTransactionValidationFunction {
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

	if !plugin.closeCalled {
		t.Fatal("Closing the consensus should call close on the registered plugins")
	}
}
