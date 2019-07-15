package minting

import (
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/threefoldtech/rivine/modules"
	"github.com/threefoldtech/rivine/persist"
	"github.com/threefoldtech/rivine/pkg/encoding/rivbin"
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
		genesisMintCondition               types.UnlockConditionProxy
		minterDefinitionTransactionVersion types.TransactionVersion
		storage                            modules.PluginViewStorage
		unregisterCallback                 modules.PluginUnregisterCallback

		binMarshal   func(v interface{}) []byte
		binUnmarshal func(b []byte, v interface{}) error
	}
)

type (
	// PluginOptions allows optional parameters to be defined for the minting plugin.
	PluginOptions struct {
		CoinDestructionTransactionVersion types.TransactionVersion
		UseLegacySiaEncoding              bool
	}
)

// NewMintingPlugin creates a new Plugin with a genesisMintCondition and correct transaction versions
func NewMintingPlugin(genesisMintCondition types.UnlockConditionProxy, minterDefinitionTransactionVersion, coinCreationTransactionVersion types.TransactionVersion, opts *PluginOptions) *Plugin {
	p := &Plugin{
		genesisMintCondition:               genesisMintCondition,
		minterDefinitionTransactionVersion: minterDefinitionTransactionVersion,
	}
	types.RegisterTransactionVersion(minterDefinitionTransactionVersion, MinterDefinitionTransactionController{
		MintConditionGetter: p,
		TransactionVersion:  minterDefinitionTransactionVersion,
	})
	types.RegisterTransactionVersion(coinCreationTransactionVersion, CoinCreationTransactionController{
		MintConditionGetter: p,
		TransactionVersion:  coinCreationTransactionVersion,
	})
	var legacyEncoding bool
	if opts != nil {
		if opts.CoinDestructionTransactionVersion > 0 {
			types.RegisterTransactionVersion(opts.CoinDestructionTransactionVersion, CoinDestructionTransactionController{
				TransactionVersion: opts.CoinDestructionTransactionVersion,
			})
		}
		legacyEncoding = opts.UseLegacySiaEncoding
	}
	if legacyEncoding {
		p.binMarshal = siabin.Marshal
		p.binUnmarshal = siabin.Unmarshal
	} else {
		p.binMarshal = rivbin.Marshal
		p.binUnmarshal = rivbin.Unmarshal
	}
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

		mintcond := p.binMarshal(p.genesisMintCondition)
		err := mintingBucket.Put(encodeBlockheight(0), mintcond)
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
	var err error
	for _, txn := range block.Transactions {
		err = p.ApplyTransaction(txn, block, height, bucket)
		if err != nil {
			return err
		}
	}
	return nil
}

// ApplyTransaction applies a minting transactions to the minting bucket.
func (p *Plugin) ApplyTransaction(txn types.Transaction, block types.Block, height types.BlockHeight, bucket *persist.LazyBoltBucket) error {
	if bucket == nil {
		return errors.New("minting bucket does not exist")
	}
	var (
		mintingBucket *bolt.Bucket
	)
	// check the version and handle the ones we care about
	switch txn.Version {
	case p.minterDefinitionTransactionVersion:
		mdtx, err := MinterDefinitionTransactionFromTransaction(txn, p.minterDefinitionTransactionVersion)
		if err != nil {
			return fmt.Errorf("unexpected error while unpacking the minter def. tx type: %v" + err.Error())
		}
		if mintingBucket == nil {
			mintingBucket, err = bucket.Bucket([]byte(bucketMintConditions))
			if err != nil {
				return errors.New("mintcondition bucket does not exist")
			}
		}
		err = mintingBucket.Put(encodeBlockheight(height), p.binMarshal(mdtx.MintCondition))
		if err != nil {
			return fmt.Errorf(
				"failed to put mint condition for block height %d: %v",
				height, err)
		}
	}
	return nil
}

// RevertBlock reverts a block's minting transaction from the minting bucket
func (p *Plugin) RevertBlock(block types.Block, height types.BlockHeight, bucket *persist.LazyBoltBucket) error {
	if bucket == nil {
		return errors.New("mint conditions bucket does not exist")
	}
	// collect all one-per-block mint conditions
	var err error
	for _, txn := range block.Transactions {
		err = p.RevertTransaction(txn, block, height, bucket)
		if err != nil {
			return err
		}
	}
	return nil
}

// RevertTransaction reverts a minting transactions to the minting bucket.
func (p *Plugin) RevertTransaction(txn types.Transaction, block types.Block, height types.BlockHeight, bucket *persist.LazyBoltBucket) error {
	if bucket == nil {
		return errors.New("minting bucket does not exist")
	}
	var (
		err           error
		mintingBucket *bolt.Bucket
	)
	// check the version and handle the ones we care about
	switch txn.Version {
	case p.minterDefinitionTransactionVersion:
		if mintingBucket == nil {
			mintingBucket, err = bucket.Bucket([]byte(bucketMintConditions))
			if err != nil {
				return errors.New("mintcondition bucket does not exist")
			}
		}
		err := mintingBucket.Delete(encodeBlockheight(height))
		if err != nil {
			return fmt.Errorf(
				"failed to delete mint condition for block height %d: %v",
				height, err)
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

		err := p.binUnmarshal(b, &mintCondition)
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
		k, b = cursor.Seek(encodeBlockheight(height))
		if len(k) == 0 {
			// could be that we're past the last key, let's try the last key first
			k, b = cursor.Last()
			if len(k) == 0 {
				return errors.New("corrupt transaction DB: no matching mint condition could be found")
			}
			return nil
		}
		foundHeight := decodeBlockheight(k)
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

	err = p.binUnmarshal(b, &mintCondition)
	if err != nil {
		return types.UnlockConditionProxy{}, fmt.Errorf("corrupt transaction DB: failed to decode found mint condition: %v", err)
	}

	return mintCondition, nil
}

// Close unregisters the plugin from the consensus
func (p *Plugin) Close() error {
	return p.storage.Close()
}

// encodeBlockheight encodes the given blockheight as a sortable key
func encodeBlockheight(height types.BlockHeight) []byte {
	key := make([]byte, 8)
	binary.BigEndian.PutUint64(key[:], uint64(height))
	return key
}

// eecodeBlockheight decodes the given sortable key as a blockheight
func decodeBlockheight(key []byte) types.BlockHeight {
	return types.BlockHeight(binary.BigEndian.Uint64(key))
}
