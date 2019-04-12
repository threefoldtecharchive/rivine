package minting

import (
	"errors"
	"fmt"

	"github.com/threefoldtech/rivine/modules"
	"github.com/threefoldtech/rivine/persist"
	"github.com/threefoldtech/rivine/pkg/encoding/siabin"
	"github.com/threefoldtech/rivine/types"

	bolt "github.com/rivine/bbolt"
)

const (
	pluginDBVersion = "1.0.0.0"
	pluginDBHeader  = "mintingPlugin"
)

var (
	bucketMintConditions = []byte("mintconditions")
)

type (
	// Plugin is a struct defines the minting plugin
	Plugin struct {
		genesisMintCondition types.UnlockConditionProxy
		storage              modules.PluginViewStorage
		unregisterCallback   modules.PluginUnregisterCallback
	}
)

// New creates a new Plugin with a genesisMintCondition
func NewMintingPlugin(genesisMintCondition types.UnlockConditionProxy) *Plugin {
	p := &Plugin{
		genesisMintCondition: genesisMintCondition,
	}
	types.RegisterTransactionVersion(TransactionVersionMinterDefinition, MinterDefinitionTransactionController{MintConditionGetter: p})
	types.RegisterTransactionVersion(TransactionVersionCoinCreation, CoinCreationTransactionController{MintConditionGetter: p})
	return p
}

// InitPlugin initializes the Bucket for the first time
func (p *Plugin) InitPlugin(metadata *persist.Metadata, bucket *bolt.Bucket, storage modules.PluginViewStorage, unregisterCallback modules.PluginUnregisterCallback) (persist.Metadata, error) {
	p.storage = storage
	p.unregisterCallback = unregisterCallback
	if metadata == nil {
		mintingBucket := bucket.Bucket([]byte(bucketMintConditions))
		if mintingBucket == nil {
			var err error
			mintingBucket, err = bucket.CreateBucket([]byte(bucketMintConditions))
			if err != nil {
				return persist.Metadata{}, fmt.Errorf("failed to create mintcondition bucket: %v", err)
			}
		}

		mintcond := siabin.Marshal(p.genesisMintCondition)
		err := mintingBucket.Put(EncodeBlockheight(0), mintcond)
		if err != nil {
			return persist.Metadata{}, fmt.Errorf("failed to store genesis mint condition: %v", err)
		}
		metadata = &persist.Metadata{
			Version: pluginDBVersion,
			Header:  pluginDBHeader,
		}
	} else if metadata.Version != pluginDBVersion {
		return persist.Metadata{}, errors.New("There is only 1 version of this plugin, version mismatch")
	}
	return *metadata, nil
}

// ApplyBlock applies a block's minting transactions to the minting bucket.
func (p *Plugin) ApplyBlock(block types.Block, height types.BlockHeight, bucket *persist.LazyBoltBucket) error {
	if bucket == nil {
		return errors.New("minting bucket does not exist")
	}
	mintingBucket, err := bucket.Bucket([]byte(bucketMintConditions))
	if err != nil {
		return errors.New("mintcondition bucket does not exist")
	}
	for i := range block.Transactions {
		rtx := &block.Transactions[i]
		if rtx.Version == types.TransactionVersionOne {
			continue // ignore most common Tx
		}
		// check the version and handle the ones we care about
		switch rtx.Version {
		case TransactionVersionMinterDefinition:
			mdtx, err := MinterDefinitionTransactionFromTransaction(*rtx)
			if err != nil {
				return fmt.Errorf("unexpected error while unpacking the minter def. tx type: %v" + err.Error())
			}
			err = mintingBucket.Put(EncodeBlockheight(height), siabin.Marshal(mdtx.MintCondition))
			if err != nil {
				return fmt.Errorf(
					"failed to put mint condition for block height %d: %v",
					height, err)
			}
		}
	}
	return nil
}

// RevertBlock reverts a block's minting transaction from the minting bucket
func (p *Plugin) RevertBlock(block types.Block, height types.BlockHeight, bucket *persist.LazyBoltBucket) error {
	if bucket == nil {
		return errors.New("mint conditions bucket does not exist")
	}
	mintingBucket, err := bucket.Bucket([]byte(bucketMintConditions))
	if err != nil {
		return errors.New("mintcondition bucket does not exist")
	}
	// collect all one-per-block mint conditions
	for i := range block.Transactions {
		rtx := &block.Transactions[i]
		if rtx.Version == types.TransactionVersionOne {
			continue // ignore most common Tx
		}

		// check the version and handle the ones we care about
		switch rtx.Version {
		case TransactionVersionMinterDefinition:
			err := mintingBucket.Delete(EncodeBlockheight(height))
			if err != nil {
				return fmt.Errorf(
					"failed to delete mint condition for block height %d: %v",
					height, err)
			}
			return nil
		}
	}
	return nil
}

// GetActiveMintCondition implements types.MintConditionGetter.GetActiveMintCondition
func (p *Plugin) GetActiveMintCondition() (types.UnlockConditionProxy, error) {
	var mintCondition types.UnlockConditionProxy
	err := p.storage.View(func(bucket *bolt.Bucket) error {
		var b []byte
		mintingBucket := bucket.Bucket([]byte(bucketMintConditions))
		// return the last cursor
		cursor := mintingBucket.Cursor()

		var k []byte
		k, b = cursor.Last()
		if len(k) == 0 {
			return errors.New("no matching mint condition could be found")
		}

		err := siabin.Unmarshal(b, &mintCondition)
		if err != nil {
			return fmt.Errorf("failed to decode found mint condition: %v", err)
		}
		return nil
	})

	return mintCondition, err
}

// GetMintConditionAt implements types.MintConditionGetter.GetMintConditionAt
func (p *Plugin) GetMintConditionAt(height types.BlockHeight) (types.UnlockConditionProxy, error) {
	var mintCondition types.UnlockConditionProxy
	var b []byte
	err := p.storage.View(func(bucket *bolt.Bucket) error {
		mintingBucket := bucket.Bucket([]byte(bucketMintConditions))
		cursor := mintingBucket.Cursor()
		var k []byte
		k, b = cursor.Seek(EncodeBlockheight(height))
		if len(k) == 0 {
			// could be that we're past the last key, let's try the last key first
			k, b = cursor.Last()
			if len(k) == 0 {
				return errors.New("corrupt transaction DB: no matching mint condition could be found")
			}
			return nil
		}
		foundHeight := DecodeBlockheight(k)
		if foundHeight <= height {
			return nil
		}
		k, b = cursor.Prev()
		if len(k) == 0 {
			return errors.New("corrupt transaction DB: no matching mint condition could be found")
		}
		return nil
	})
	if err != nil {
		return types.UnlockConditionProxy{}, err
	}

	err = siabin.Unmarshal(b, &mintCondition)
	if err != nil {
		return types.UnlockConditionProxy{}, fmt.Errorf("corrupt transaction DB: failed to decode found mint condition: %v", err)
	}

	return mintCondition, nil
}

// Close unregisters the plugin from the consensus
func (p *Plugin) Close() error {
	p.unregisterCallback(p)
	return p.storage.Close()
}
