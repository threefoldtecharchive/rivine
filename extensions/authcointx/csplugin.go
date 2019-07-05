package authcointx

import (
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/threefoldtech/rivine/modules"
	"github.com/threefoldtech/rivine/persist"
	"github.com/threefoldtech/rivine/pkg/encoding/rivbin"
	"github.com/threefoldtech/rivine/types"

	bolt "github.com/rivine/bbolt"
)

const (
	pluginDBVersion = "1.0.0.0"
	pluginDBHeader  = "AuthCoinTransferPlugin"
)

var (
	bucketAuthConditions = []byte("authconditions")
	bucketAuthAddresses  = []byte("authaddresses")
)

type (
	// Plugin is a struct defines the Auth. Coin Transfer plugin
	Plugin struct {
		genesisAuthCondition                  types.UnlockConditionProxy
		authAddressUpdateTransactionVersion   types.TransactionVersion
		authConditionUpdateTransactionVersion types.TransactionVersion
		storage                               modules.PluginViewStorage
		unregisterCallback                    modules.PluginUnregisterCallback
	}
)

// NewPlugin creates a new Plugin with a genesisAuthCondition and correct transaction versions.
// NOTE: do not register the default rivine transaction versions (0x00 and 0x01) if you use this plugin!!!
func NewPlugin(genesisAuthCondition types.UnlockConditionProxy, authAddressUpdateTransactionVersion, authConditionUpdateTransactionVersion types.TransactionVersion) *Plugin {
	p := &Plugin{
		genesisAuthCondition:                  genesisAuthCondition,
		authAddressUpdateTransactionVersion:   authAddressUpdateTransactionVersion,
		authConditionUpdateTransactionVersion: authConditionUpdateTransactionVersion,
	}
	types.RegisterTransactionVersion(types.TransactionVersionZero, DisabledTransactionController{})
	types.RegisterTransactionVersion(types.TransactionVersionOne, AuthStandardTransferTransactionController{
		AuthInfoGetter: p,
	})
	types.RegisterTransactionVersion(authAddressUpdateTransactionVersion, AuthAddressUpdateTransactionController{
		AuthInfoGetter:     p,
		TransactionVersion: authAddressUpdateTransactionVersion,
	})
	types.RegisterTransactionVersion(authConditionUpdateTransactionVersion, AuthConditionUpdateTransactionController{
		AuthInfoGetter:     p,
		TransactionVersion: authConditionUpdateTransactionVersion,
	})
	return p
}

// InitPlugin initializes the Bucket for the first time
func (p *Plugin) InitPlugin(metadata *persist.Metadata, bucket *bolt.Bucket, storage modules.PluginViewStorage, unregisterCallback modules.PluginUnregisterCallback) (persist.Metadata, error) {
	p.storage = storage
	p.unregisterCallback = unregisterCallback
	if metadata == nil {
		authBucket := bucket.Bucket([]byte(bucketAuthConditions))
		if authBucket == nil {
			var err error
			authBucket, err = bucket.CreateBucket([]byte(bucketAuthConditions))
			if err != nil {
				return persist.Metadata{}, fmt.Errorf("failed to create auth condition bucket: %v", err)
			}
		}

		authAddressesBucket := bucket.Bucket([]byte(bucketAuthAddresses))
		if authAddressesBucket == nil {
			var err error
			authAddressesBucket, err = bucket.CreateBucket([]byte(bucketAuthAddresses))
			if err != nil {
				return persist.Metadata{}, fmt.Errorf("failed to create auth addresses bucket: %v", err)
			}
		}

		mintcond := rivbin.Marshal(p.genesisAuthCondition)
		err := authBucket.Put(encodeBlockheight(0), mintcond)
		if err != nil {
			return persist.Metadata{}, fmt.Errorf("failed to store genesis auth condition: %v", err)
		}
		metadata = &persist.Metadata{
			Version: pluginDBVersion,
			Header:  pluginDBHeader,
		}
	} else if metadata.Version != pluginDBVersion {
		return persist.Metadata{}, errors.New("There is only 1 version of this plugin, version mismatch")
	} else if metadata.Header != pluginDBHeader {
		return persist.Metadata{}, errors.New("There is only 1 header of this plugin, header mismatch")
	}
	return *metadata, nil
}

// ApplyBlock applies a block's minting transactions to the minting bucket.
func (p *Plugin) ApplyBlock(block types.Block, height types.BlockHeight, bucket *persist.LazyBoltBucket) error {
	if bucket == nil {
		return errors.New("plugin bucket does not exist")
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
		return errors.New("plugin bucket does not exist")
	}
	// check the version and handle the ones we care about
	switch txn.Version {
	case p.authConditionUpdateTransactionVersion:
		acutx, err := AuthConditionUpdateTransactionFromTransaction(txn, p.authConditionUpdateTransactionVersion)
		if err != nil {
			return fmt.Errorf("unexpected error while unpacking the auth condition update tx type: %v" + err.Error())
		}
		authBucket, err := bucket.Bucket([]byte(bucketAuthConditions))
		if err != nil {
			return errors.New("auth conditions bucket does not exist")
		}
		err = authBucket.Put(encodeBlockheight(height), rivbin.Marshal(acutx.AuthCondition))
		if err != nil {
			return fmt.Errorf(
				"failed to put auth condition for block height %d: %v",
				height, err)
		}

	case p.authAddressUpdateTransactionVersion:
		aautx, err := AuthAddressUpdateTransactionFromTransaction(txn, p.authAddressUpdateTransactionVersion)
		if err != nil {
			return fmt.Errorf("unexpected error while unpacking the auth address update tx type: %v" + err.Error())
		}
		authBucket, err := bucket.Bucket([]byte(bucketAuthAddresses))
		if err != nil {
			return errors.New("auth conditions bucket does not exist")
		}
		// store all new (de)auth address info
		// an address can only appear once per tx, so no need to do intermediate bucket caching
		for _, address := range aautx.AuthAddresses {
			addressAuthBucket, err := authBucket.CreateBucketIfNotExists(rivbin.Marshal(address))
			if err != nil {
				return fmt.Errorf("auth address (%s) condition bucket does not exist and could not be created: %v", address.String(), err)
			}
			err = addressAuthBucket.Put(encodeBlockheight(height), rivbin.Marshal(true))
			if err != nil {
				return fmt.Errorf(
					"failed to put auth condition for address %s block height %d: %v",
					address.String(), height, err)
			}
		}
		for _, address := range aautx.DeauthAddresses {
			addressAuthBucket, err := authBucket.CreateBucketIfNotExists(rivbin.Marshal(address))
			if err != nil {
				return fmt.Errorf("auth address (%s) condition bucket does not exist and could not be created: %v", address.String(), err)
			}
			err = addressAuthBucket.Put(encodeBlockheight(height), rivbin.Marshal(false))
			if err != nil {
				return fmt.Errorf(
					"failed to put deauth condition for address %s block height %d: %v",
					address.String(), height, err)
			}
		}
	}
	return nil
}

// RevertBlock reverts a block's minting transaction from the minting bucket
func (p *Plugin) RevertBlock(block types.Block, height types.BlockHeight, bucket *persist.LazyBoltBucket) error {
	if bucket == nil {
		return errors.New("plugin bucket does not exist")
	}
	var err error
	for _, txn := range block.Transactions {
		err = p.RevertTransaction(txn, block, height, bucket)
		if err != nil {
			return err
		}
	}
	return nil
}

// RevertTransaction reverts a  minting transaction from the minting bucket
func (p *Plugin) RevertTransaction(txn types.Transaction, block types.Block, height types.BlockHeight, bucket *persist.LazyBoltBucket) error {
	if bucket == nil {
		return errors.New("plugin bucket does not exist")
	}
	// check the version and handle the ones we care about
	switch txn.Version {
	case p.authConditionUpdateTransactionVersion:
		authBucket, err := bucket.Bucket([]byte(bucketAuthConditions))
		if err != nil {
			return errors.New("auth conditions bucket does not exist")
		}
		err = authBucket.Delete(encodeBlockheight(height))
		if err != nil {
			return fmt.Errorf(
				"failed to delete auth condition for block height %d: %v",
				height, err)
		}

	case p.authAddressUpdateTransactionVersion:
		aautx, err := AuthAddressUpdateTransactionFromTransaction(txn, p.authAddressUpdateTransactionVersion)
		if err != nil {
			return fmt.Errorf("unexpected error while unpacking the auth address update tx type: %v" + err.Error())
		}
		authBucket, err := bucket.Bucket([]byte(bucketAuthAddresses))
		if err != nil {
			return errors.New("auth conditions bucket does not exist")
		}
		// delete all reverted (de)auth address info
		// an address can only appear once per tx, so no need to do intermediate bucket caching
		for _, address := range aautx.AuthAddresses {
			addressAuthBucket := authBucket.Bucket(rivbin.Marshal(address))
			if addressAuthBucket == nil {
				return fmt.Errorf("auth address (%s) condition bucket does not exist", address.String())
			}
			err = addressAuthBucket.Delete(encodeBlockheight(height))
			if err != nil {
				return fmt.Errorf(
					"failed to delete auth condition for address %s block height %d: %v",
					address.String(), height, err)
			}
		}
		for _, address := range aautx.DeauthAddresses {
			addressAuthBucket := authBucket.Bucket(rivbin.Marshal(address))
			if addressAuthBucket == nil {
				return fmt.Errorf("auth address (%s) condition bucket does not exist", address.String())
			}
			err = addressAuthBucket.Delete(encodeBlockheight(height))
			if err != nil {
				return fmt.Errorf(
					"failed to delete deauth condition for address %s block height %d: %v",
					address.String(), height, err)
			}
		}
	}
	return nil
}

// GetActiveAuthCondition implements types.AuthInfoGetter.GetActiveAuthCondition
func (p *Plugin) GetActiveAuthCondition() (types.UnlockConditionProxy, error) {
	var authCondition types.UnlockConditionProxy
	err := p.storage.View(func(bucket *bolt.Bucket) error {
		var b []byte
		authBucket := bucket.Bucket([]byte(bucketAuthConditions))
		// return the last cursor
		cursor := authBucket.Cursor()

		var k []byte
		k, b = cursor.Last()
		if len(k) == 0 {
			return errors.New("no matching auth condition could be found")
		}

		err := rivbin.Unmarshal(b, &authCondition)
		if err != nil {
			return fmt.Errorf("failed to decode found auth condition: %v", err)
		}
		return nil
	})

	return authCondition, err
}

// GetAuthConditionAt implements types.AuthInfoGetter.GetAuthConditionAt
func (p *Plugin) GetAuthConditionAt(height types.BlockHeight) (types.UnlockConditionProxy, error) {
	var authCondition types.UnlockConditionProxy
	var b []byte
	err := p.storage.View(func(bucket *bolt.Bucket) error {
		authBucket := bucket.Bucket([]byte(bucketAuthConditions))
		cursor := authBucket.Cursor()
		var k []byte
		k, b = cursor.Seek(encodeBlockheight(height))
		if len(k) == 0 {
			// could be that we're past the last key, let's try the last key first
			k, b = cursor.Last()
			if len(k) == 0 {
				return errors.New("corrupt plugin DB: no matching auth condition could be found")
			}
			return nil
		}
		foundHeight := decodeBlockheight(k)
		if foundHeight <= height {
			return nil
		}
		k, b = cursor.Prev()
		if len(k) == 0 {
			return errors.New("corrupt plugin DB: no matching auth condition could be found")
		}
		return nil
	})
	if err != nil {
		return types.UnlockConditionProxy{}, err
	}

	err = rivbin.Unmarshal(b, &authCondition)
	if err != nil {
		return types.UnlockConditionProxy{}, fmt.Errorf("corrupt plugin DB: failed to decode found mint condition: %v", err)
	}

	return authCondition, nil
}

// GetAddressesAuthStateNow rerturns for each requested address, in order as given,
// the current auth state for that address as a boolean: true if authed, false otherwise.
// If exitEarlyFn is given GetAddressesAuthStateNow can stop earlier in case exitEarlyFn returns true for an iteration.
func (p *Plugin) GetAddressesAuthStateNow(addresses []types.UnlockHash, exitEarlyFn func(index int, state bool) bool) ([]bool, error) {
	l := len(addresses)
	if l == 0 {
		return nil, errors.New("no addresses given to check the current auth state for")
	}
	stateSlice := make([]bool, l)
	err := p.storage.View(func(bucket *bolt.Bucket) error {
		authBucket := bucket.Bucket([]byte(bucketAuthAddresses))
		var b []byte
		for index, address := range addresses {
			addressAuthBucket := authBucket.Bucket(rivbin.Marshal(address))
			if addressAuthBucket != nil {
				// return the last cursor
				cursor := addressAuthBucket.Cursor()
				var k []byte
				k, b = cursor.Last()
				if len(k) != 0 {
					var authState bool
					err := rivbin.Unmarshal(b, &authState)
					if err != nil {
						return fmt.Errorf("failed to decode found auth condition for address %s: %v", address.String(), err)
					}
					stateSlice[index] = authState
				}
			}
			if exitEarlyFn != nil && exitEarlyFn(index, stateSlice[index]) {
				return nil
			}
		}
		return nil
	})
	return stateSlice, err
}

// GetAddressesAuthStateAt rerturns for each requested address, in order as given,
// the auth state at the given height for that address as a boolean: true if authed, false otherwise.
// If exitEarlyFn is given GetAddressesAuthStateNow can stop earlier in case exitEarlyFn returns true for an iteration.
func (p *Plugin) GetAddressesAuthStateAt(height types.BlockHeight, addresses []types.UnlockHash, exitEarlyFn func(index int, state bool) bool) ([]bool, error) {
	l := len(addresses)
	if l == 0 {
		return nil, errors.New("no addresses given to check the current auth state for")
	}
	stateSlice := make([]bool, l)
	err := p.storage.View(func(bucket *bolt.Bucket) error {
		authBucket := bucket.Bucket([]byte(bucketAuthAddresses))
		var err error
		for index, address := range addresses {
			addressAuthBucket := authBucket.Bucket(rivbin.Marshal(address))
			if addressAuthBucket != nil {
				stateSlice[index], err = func() (bool, error) {
					cursor := addressAuthBucket.Cursor()
					var k, b []byte
					k, b = cursor.Seek(encodeBlockheight(height))
					if len(k) == 0 {
						// could be that we're past the last key, let's try the last key first
						k, b = cursor.Last()
						if len(k) == 0 {
							return false, nil
						}
					}
					foundHeight := decodeBlockheight(k)
					if foundHeight > height {
						k, b = cursor.Prev()
						if len(k) == 0 {
							return false, nil
						}
					}
					var authState bool
					err = rivbin.Unmarshal(b, &authState)
					if err != nil {
						return false, fmt.Errorf("failed to decode found address (%s) auth condition for address: %v", address.String(), err)
					}
					return authState, nil
				}()
			}
			if exitEarlyFn != nil && exitEarlyFn(index, stateSlice[index]) {
				return nil
			}
		}
		return nil
	})
	return stateSlice, err
}

// Close unregisters the plugin from the consensus
func (p *Plugin) Close() error {
	p.unregisterCallback(p)
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
