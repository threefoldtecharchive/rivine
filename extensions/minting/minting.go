package minting

import (
	"errors"
	"fmt"

	bolt "github.com/rivine/bbolt"
	persist "github.com/threefoldtech/rivine/persist"
	"github.com/threefoldtech/rivine/pkg/encoding/siabin"
	"github.com/threefoldtech/rivine/types"
)

const (
	pluginDbVersion = "1.0.0.0"
	pluginDbHeader  = "mintingPlugin"
)

type (
	// Plugin is a struct defines the minting plugin
	Plugin struct {
		genesisMintCondition types.UnlockConditionProxy
		name                 string
		database             *persist.BoltDatabase
	}
)

// New creates a new Plugin with a genesisMintCondition
func New(genesisMintCondition types.UnlockConditionProxy) *Plugin {
	p := &Plugin{
		genesisMintCondition: genesisMintCondition,
	}
	types.RegisterTransactionVersion(TransactionVersionMinterDefinition, MinterDefinitionTransactionController{MintConditionGetter: p})
	types.RegisterTransactionVersion(TransactionVersionCoinCreation, CoinCreationTransactionController{MintConditionGetter: p})
	return p
}

// InitBucket initializes the Bucket for the first time
func (p *Plugin) InitBucket(metadata *persist.Metadata, name string, bucket *bolt.Bucket, db *persist.BoltDatabase) (persist.Metadata, error) {
	p.database = db
	p.name = name
	if metadata == nil {
		// TODO setup bucket for first time, store the genesis mint condition
		mintcond := siabin.Marshal(p.genesisMintCondition)
		err := bucket.Put(EncodeBlockheight(0), mintcond)
		if err != nil {
			return persist.Metadata{}, fmt.Errorf("failed to store genesis mint condition: %v", err)
		}
		metadata = &persist.Metadata{
			Version: pluginDbVersion,
			Header:  pluginDbHeader,
		}
	} else if metadata.Version != pluginDbVersion {
		// TODO upgrade and set metadata to new version
		return persist.Metadata{}, errors.New("There is only 1 version of this plugin, version mismatch")
	}
	return *metadata, nil
}

// ApplyBlock applies a block's minting transactions to the minting bucket.
func (p *Plugin) ApplyBlock(block types.Block, height types.BlockHeight, _ *persist.LazyBoltBucket) error {
	err := p.database.Update(func(tx *bolt.Tx) (err error) {
		bucket := tx.Bucket([]byte(p.name))
		if bucket == nil {
			return errors.New("corrupt transaction DB: mint conditions bucket does not exist")
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
				err = bucket.Put(EncodeBlockheight(height), siabin.Marshal(mdtx.MintCondition))
				if err != nil {
					return fmt.Errorf(
						"failed to put mint condition for block height %d: %v",
						height, err)
				}
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

// RevertBlock reverts a block's minting transaction from the minting bucket
func (p *Plugin) RevertBlock(block types.Block, height types.BlockHeight, bucket *persist.LazyBoltBucket) error {
	err := p.database.View(func(tx *bolt.Tx) (err error) {
		bucket := tx.Bucket([]byte(p.name))
		if bucket == nil {
			return errors.New("corrupt transaction DB: mint conditions bucket does not exist")
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
				err := bucket.Delete(EncodeBlockheight(height))
				if err != nil {
					return fmt.Errorf(
						"failed to delete mint condition for block height %d: %v",
						height, err)
				}
				return nil
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

// GetActiveMintCondition implements types.MintConditionGetter.GetActiveMintCondition
func (p *Plugin) GetActiveMintCondition() (types.UnlockConditionProxy, error) {
	var mintCondition types.UnlockConditionProxy
	err := p.database.View(func(tx *bolt.Tx) (err error) {
		bucket := tx.Bucket([]byte(p.name))
		if bucket == nil {
			return errors.New("corrupt transaction DB: mint conditions bucket does not exist")
		}
		var b []byte
		// return the last cursor
		cursor := bucket.Cursor()

		var k []byte
		k, b = cursor.Last()
		if len(k) == 0 {
			return errors.New("corrupt transaction DB: no matching mint condition could be found")
		}

		err = siabin.Unmarshal(b, &mintCondition)
		if err != nil {
			return fmt.Errorf("corrupt transaction DB: failed to decode found mint condition: %v", err)
		}
		return nil
	})
	if err != nil {
		return types.UnlockConditionProxy{}, fmt.Errorf("corrupt transaction DB: failed to decode found mint condition: %v", err)
	}

	// mint condition found, return it
	return mintCondition, nil
}

// GetMintConditionAt implements types.MintConditionGetter.GetMintConditionAt
func (p *Plugin) GetMintConditionAt(height types.BlockHeight) (types.UnlockConditionProxy, error) {
	var mintCondition types.UnlockConditionProxy
	var b []byte
	err := p.database.View(func(tx *bolt.Tx) (err error) {
		bucket := tx.Bucket([]byte(p.name))
		if bucket == nil {
			return errors.New("corrupt transaction DB: mint conditions bucket does not exist")
		}
		cursor := bucket.Cursor()
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
