package consensus

import (
	"context"
	"errors"
	"fmt"

	"github.com/threefoldtech/rivine/modules"
	"github.com/threefoldtech/rivine/persist"
	"github.com/threefoldtech/rivine/pkg/encoding/rivbin"
	"github.com/threefoldtech/rivine/types"

	bolt "github.com/rivine/bbolt"
)

// plugin user errors
var (
	// ErrPluginNameReserved is returned in case a plugin with that name
	// is not allowed because it is a reserved bucket name.
	ErrPluginNameReserved = errors.New("plugin bucket name is reserved")
	// ErrPluginNameEmpty is returned in case a plugin with that name is empty.
	ErrPluginNameEmpty = errors.New("plugin bucket name cannot be empty")
	// ErrPluginExists is returned in case a plugin with the same name is registered.
	ErrPluginExists = errors.New("a plugin with that name is already registered")
)

// plugin system errors
var (
	// ErrPluginGhostMetadata is returned in case a plugin which wasn't registered yet, does have metadata.
	ErrPluginGhostMetadata = errors.New("a plugin that wasn't registered yet does have metadata")
	// ErrMissingPluginMetadata is returned in case a plugin is missing metadata.
	ErrMissingPluginMetadata = errors.New("the metadata of the plugin is missing")
	// ErrMissingMetadataBucket is returned in case root plugins metadata folder is missing
	ErrMissingMetadataBucket = errors.New("root plugins metadata folder is missing")
)

var (
	reservedPluginNames = []string{"Metadata"}
)

var (
	bucketPluginsMetadata = []byte("Metadata")
)

// pluginMetadata is used to store the metadata for a plugin.
type pluginMetadata struct {
	Version           *persist.Metadata
	ConsensusChangeID modules.ConsensusChangeID
}

// RegisterPlugin registers a plugin to the map of plugins,
// initializes its bucket using the plugin and ensures it receives all
// consensus updates it is missing (as a special case this means anything).
// This initial sync is cancelled if tyhe passed context is cancelled.
func (cs *ConsensusSet) RegisterPlugin(ctx context.Context, name string, plugin modules.ConsensusSetPlugin) error {
	cs.log.Debugln("Registering plugin ", name)
	if name == "" {
		return ErrPluginNameEmpty
	}
	for _, reservedName := range reservedPluginNames {
		if reservedName == name {
			return ErrPluginNameReserved
		}
	}
	err := cs.tg.Add()
	if err != nil {
		return err
	}
	defer cs.tg.Done()
	cs.mu.Lock()
	defer cs.mu.Unlock()

	// ensure the plugin doesn't exist yet
	if _, ok := cs.plugins[name]; ok {
		return ErrPluginExists
	}

	// init the plugin
	var consensusChangeID modules.ConsensusChangeID
	err = cs.db.Update(func(tx *bolt.Tx) (err error) {
		consensusChangeID, err = cs.initConsensusSetPlugin(tx, name, plugin)
		return err
	})
	if err != nil {
		return err
	}

	// init sync the plugin
	initialSyncCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	newConsensusChangeID, err := cs.initPluginSync(initialSyncCtx, name, plugin, consensusChangeID)
	if err != nil {
		return err
	}

	if newConsensusChangeID != consensusChangeID {
		err = cs.db.Update(func(tx *bolt.Tx) error {
			rootbucket := tx.Bucket(BucketPlugins)
			// get the metadata bucket from the rootbucket
			metadataBucket := rootbucket.Bucket(bucketPluginsMetadata)
			if metadataBucket == nil {
				return ErrPluginGhostMetadata
			}
			// get the plugin metadata as-is
			pluginMetadataBytes := metadataBucket.Get([]byte(name))
			if len(pluginMetadataBytes) == 0 {
				return ErrMissingPluginMetadata
			}
			var pluginMetadata pluginMetadata
			err := rivbin.Unmarshal(pluginMetadataBytes, &pluginMetadata)
			if err != nil {
				return err
			}

			// save the new metadata
			pluginMetadata.ConsensusChangeID = newConsensusChangeID
			metadataBytes, err := rivbin.Marshal(pluginMetadata)
			if err != nil {
				return fmt.Errorf("failed to (rivbin) marshal plugin metadata: %v", err)
			}
			return metadataBucket.Put([]byte(name), metadataBytes)
		})
		if err != nil {
			return err
		}
	}

	// Check if map is nil, if nil make one
	if cs.plugins == nil {
		cs.plugins = make(map[string]modules.ConsensusSetPlugin)
	}
	// Add plugin to cs plugins map
	cs.plugins[name] = plugin
	// return without errors
	return nil
}

// initPluginSync ensures the plugin receives all
// consensus updates it is missing (as a special case this means anything).
// This initial sync is  be cancelled if the passes context is closed.
func (cs *ConsensusSet) initPluginSync(ctx context.Context, name string, plugin modules.ConsensusSetPlugin, start modules.ConsensusChangeID) (modules.ConsensusChangeID, error) {
	newChangeID := start
	err := cs.db.View(func(tx *bolt.Tx) error {
		// 'exists' and 'entry' are going to be pointed to the first entry that
		// has not yet been seen by subscriber.
		var exists bool
		var entry changeEntry

		if start == modules.ConsensusChangeBeginning {
			// Special case: for ConsensusChangeBeginning, create an
			// initial node pointing to the genesis block. The subscriber will
			// receive the diffs for all blocks in the consensus set, including
			// the genesis block.
			entry = cs.genesisEntry()
			exists = true
		} else if start == modules.ConsensusChangeRecent {
			// Special case: for ConsensusChangeRecent, set up the
			// subscriber to start receiving only new blocks, but the
			// subscriber does not need to do any catch-up. For this
			// implementation, a no-op will have this effect.
			return nil
		} else {
			// The subscriber has provided an existing consensus change.
			// Because the subscriber already has this consensus change,
			// 'entry' and 'exists' need to be pointed at the next consensus
			// change.
			entry, exists = getEntry(tx, start)
			if !exists {
				// ErrInvalidConsensusChangeID is a named error that
				// signals a break in synchronization between the consensus set
				// persistence and the subscriber persistence. Typically,
				// receiving this error means that the subscriber needs to
				// perform a rescan of the consensus set.
				return modules.ErrInvalidConsensusChangeID
			}
			entry, exists = entry.NextEntry(tx)
		}

		bucket := persist.NewLazyBoltBucket(func() (*bolt.Bucket, error) {
			rootbucket := tx.Bucket(BucketPlugins)
			// get the metadata bucket from the rootbucket
			mdBucket := rootbucket.Bucket(bucketPluginsMetadata)
			if mdBucket == nil {
				return nil, errors.New("metadata plugins bucket is missing, while it should exist at this point")
			}
			b := rootbucket.Bucket([]byte(name))
			if b == nil {
				return nil, fmt.Errorf("bucket %s for plugin does not exist", name)
			}
			return b, nil
		})

		// Send all remaining consensus changes to the subscriber.
		for exists {
			select {
			case <-ctx.Done():
				return errors.New("aborting initPluginSync")
			default:
				cc, err := cs.computeConsensusChange(tx, entry)
				if err != nil {
					return err
				}

				coinOutputDiffs := make(map[types.CoinOutputID]types.CoinOutput)
				blockStakeOutputDiffs := make(map[types.BlockStakeOutputID]types.BlockStakeOutput)
				for _, diff := range cc.CoinOutputDiffs {
					coinOutputDiffs[diff.ID] = diff.CoinOutput
				}
				for _, diff := range cc.BlockStakeOutputDiffs {
					blockStakeOutputDiffs[diff.ID] = diff.BlockStakeOutput
				}

				var ok bool
				for _, block := range cc.RevertedBlocks {
					blockHeight, exists := cs.BlockHeightOfBlock(block)
					if exists {
						cBlock := modules.ConsensusBlock{
							Block:                  block,
							SpentCoinOutputs:       make(map[types.CoinOutputID]types.CoinOutput),
							SpentBlockStakeOutputs: make(map[types.BlockStakeOutputID]types.BlockStakeOutput),
						}
						for _, txn := range block.Transactions {
							for _, ci := range txn.CoinInputs {
								cBlock.SpentCoinOutputs[ci.ParentID], ok = coinOutputDiffs[ci.ParentID]
								if !ok {
									return fmt.Errorf("failed to find reverted coin input %s as coin output in consensus change", ci.ParentID.String())
								}
							}
							for _, bsi := range txn.BlockStakeInputs {
								cBlock.SpentBlockStakeOutputs[bsi.ParentID], ok = blockStakeOutputDiffs[bsi.ParentID]
								if !ok {
									return fmt.Errorf("failed to find reverted block stake input %s as block stake output in consensus change", bsi.ParentID.String())
								}
							}
						}

						err = plugin.RevertBlock(cBlock, blockHeight, bucket)
						if err != nil {
							return err
						}
					}
				}
				for _, block := range cc.AppliedBlocks {
					blockHeight, exists := cs.BlockHeightOfBlock(block)
					if exists {
						cBlock := modules.ConsensusBlock{
							Block:                  block,
							SpentCoinOutputs:       make(map[types.CoinOutputID]types.CoinOutput),
							SpentBlockStakeOutputs: make(map[types.BlockStakeOutputID]types.BlockStakeOutput),
						}
						for _, txn := range block.Transactions {
							for _, ci := range txn.CoinInputs {
								cBlock.SpentCoinOutputs[ci.ParentID], ok = coinOutputDiffs[ci.ParentID]
								if !ok {
									return fmt.Errorf("failed to find applied coin input %s as coin output in consensus change", ci.ParentID.String())
								}
							}
							for _, bsi := range txn.BlockStakeInputs {
								cBlock.SpentBlockStakeOutputs[bsi.ParentID], ok = blockStakeOutputDiffs[bsi.ParentID]
								if !ok {
									return fmt.Errorf("failed to find applied block stake input %s as block stake output in consensus change", bsi.ParentID.String())
								}
							}
						}

						err = plugin.ApplyBlock(cBlock, blockHeight, bucket)
						if err != nil {
							return err
						}
					}
				}
				newChangeID = cc.ID
				entry, exists = entry.NextEntry(tx)
			}
		}
		return nil
	})
	return newChangeID, err
}

func (cs *ConsensusSet) initConsensusSetPlugin(tx *bolt.Tx, name string, plugin modules.ConsensusSetPlugin) (modules.ConsensusChangeID, error) {
	// get the root plugins bucket
	rootbucket := tx.Bucket([]byte(BucketPlugins))
	if rootbucket == nil {
		// create the root plugins bucket
		var err error
		rootbucket, err = tx.CreateBucket([]byte(BucketPlugins))
		if err != nil {
			return modules.ConsensusChangeID{}, err
		}
	}

	// get the plugin bucket
	bucket := rootbucket.Bucket([]byte(name))
	if bucket == nil {
		// create the metadata
		metadataBucket, err := rootbucket.CreateBucketIfNotExists(bucketPluginsMetadata)
		if err != nil {
			return modules.ConsensusChangeID{}, err
		}
		data := metadataBucket.Get([]byte(name))
		if len(data) != 0 {
			return modules.ConsensusChangeID{}, ErrPluginGhostMetadata
		}
		pluginMetadataBytes, err := rivbin.Marshal(pluginMetadata{})
		if err != nil {
			return modules.ConsensusChangeID{}, fmt.Errorf("failed to rivbin marshal nil plugin metadata: %v", err)
		}
		err = metadataBucket.Put([]byte(name), pluginMetadataBytes)
		if err != nil {
			return modules.ConsensusChangeID{}, err
		}

		// create the plugin bucket in the rootbucket
		bucket, err = rootbucket.CreateBucket([]byte(name))
		if err != nil {
			return modules.ConsensusChangeID{}, err
		}
	}

	// get the metadata
	metadataBucket := rootbucket.Bucket(bucketPluginsMetadata)
	if metadataBucket == nil {
		return modules.ConsensusChangeID{}, errors.New("metadata bucket should always exist at this point")
	}
	pluginMetadataBytes := metadataBucket.Get([]byte(name))
	if len(pluginMetadataBytes) == 0 {
		return modules.ConsensusChangeID{}, ErrMissingPluginMetadata
	}
	var pluginMetadata pluginMetadata
	err := rivbin.Unmarshal(pluginMetadataBytes, &pluginMetadata)
	if err != nil {
		return modules.ConsensusChangeID{}, err
	}

	var pluginStorage modules.PluginViewStorage
	cs.log.Debugln("Creating new pluginstorage for ", name)
	pluginStorage = NewPluginStorage(cs.db, name, &cs.pluginsWaitGroup)
	// init plugin
	pluginVersion, err := plugin.InitPlugin(pluginMetadata.Version, bucket, pluginStorage, func(plugin modules.ConsensusSetPlugin) {
		cs.UnregisterPlugin(name, plugin)
	})
	if err != nil {
		return modules.ConsensusChangeID{}, err
	}
	// save the new metadata
	pluginMetadata.Version = &pluginVersion
	pluginMetadataBytes, err = rivbin.Marshal(pluginMetadata)
	if err != nil {
		return modules.ConsensusChangeID{}, fmt.Errorf("failed to rivbin marshal updated plugin metadata: %v", err)
	}
	err = metadataBucket.Put([]byte(name), pluginMetadataBytes)
	if err != nil {
		return modules.ConsensusChangeID{}, err
	}

	// return the consensus change ID that we already have, for further usage
	return pluginMetadata.ConsensusChangeID, nil
}

func (cs *ConsensusSet) validateTransactionUsingPlugins(tx modules.ConsensusTransaction, ctx types.TransactionValidationContext, btx *bolt.Tx) error {
	if len(cs.plugins) == 0 {
		return nil // if there are no errors, there is nothing to validate using plugins
	}
	var err error
	for name, plugin := range cs.plugins {
		validators := plugin.TransactionValidators()
		validatorMapping := plugin.TransactionValidatorVersionFunctionMapping()
		if len(validators) == 0 && len(validatorMapping) == 0 {
			continue // no validators attached to this plugin
		}
		txValidators, ok := validatorMapping[tx.Version]
		if !ok && len(validators) == 0 {
			continue // no validators attached to this plugin for this tx's version
		}
		// return the first error that occurs
		bucket := cs.bucketForPlugin(btx, name)
		for _, txValidator := range txValidators {
			err = txValidator(tx, ctx, bucket)
			if err != nil {
				return err
			}
		}
		for _, txValidator := range validators {
			err = txValidator(tx, ctx, bucket)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// return the root bucket for a plugin using name in the form of a LazyBoltBucket
func (cs *ConsensusSet) bucketForPlugin(tx *bolt.Tx, name string) *persist.LazyBoltBucket {
	return persist.NewLazyBoltBucket(func() (*bolt.Bucket, error) {
		mdBucket := tx.Bucket(BucketPlugins)
		if mdBucket == nil {
			return nil, errors.New("metadata plugins bucket is missing, while it should exist at this point")
		}
		b := mdBucket.Bucket([]byte(name))
		if b == nil {
			return nil, fmt.Errorf("bucket %s for plugin does not exist", name)
		}
		return b, nil
	})
}

//closePlugins calls Close on all registered plugins
func (cs *ConsensusSet) closePlugins() error {
	if err := cs.tg.Add(); err != nil {
		return err
	}
	defer cs.tg.Done()
	cs.log.Debugln("Number of plugins to close:", len(cs.plugins))

	for name, plugin := range cs.plugins {
		cs.log.Debugln("Closing plugin ", name)
		if err := plugin.Close(); err != nil {
			return err
		}
	}
	return nil
}

// UnregisterPlugin removes a plugin from the map of plugins
func (cs *ConsensusSet) UnregisterPlugin(name string, plugin modules.ConsensusSetPlugin) {
	if cs.tg.Add() != nil {
		return
	}
	defer cs.tg.Done()
	cs.mu.Lock()
	defer cs.mu.Unlock()

	if existingPlugin, ok := cs.plugins[name]; ok && existingPlugin == plugin {
		plugin.Close()
		delete(cs.plugins, name)
	} else {
		fmt.Printf("try to delete plugin %s, plugin does not exist", name)
	}
}
