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
	types.RegisterTransactionVersion(types.TransactionVersionZero, nil) // not supported
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
	for i := range block.Transactions {
		rtx := &block.Transactions[i]

		// check the version and handle the ones we care about
		switch rtx.Version {
		case p.authConditionUpdateTransactionVersion:
			acutx, err := AuthConditionUpdateTransactionFromTransaction(*rtx, p.authConditionUpdateTransactionVersion)
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
			aautx, err := AuthAddressUpdateTransactionFromTransaction(*rtx, p.authAddressUpdateTransactionVersion)
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
	}
	return nil
}

// RevertBlock reverts a block's minting transaction from the minting bucket
func (p *Plugin) RevertBlock(block types.Block, height types.BlockHeight, bucket *persist.LazyBoltBucket) error {
	if bucket == nil {
		return errors.New("plugin bucket does not exist")
	}
	for i := range block.Transactions {
		rtx := &block.Transactions[i]

		// check the version and handle the ones we care about
		switch rtx.Version {
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
			aautx, err := AuthAddressUpdateTransactionFromTransaction(*rtx, p.authAddressUpdateTransactionVersion)
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

// EnsureAddressesAreAuthNow implements types.AuthInfoGetter.EnsureAddressesAreAuthNow
func (p *Plugin) EnsureAddressesAreAuthNow(addresses ...types.UnlockHash) error {
	return p.storage.View(func(bucket *bolt.Bucket) error {
		authBucket := bucket.Bucket([]byte(bucketAuthAddresses))
		var b []byte
		for _, address := range addresses {
			addressAuthBucket := authBucket.Bucket(rivbin.Marshal(address))
			if addressAuthBucket == nil {
				return fmt.Errorf("auth address (%s) condition bucket does not exist: address was never authorized", address.String())
			}

			// return the last cursor
			cursor := addressAuthBucket.Cursor()
			var k []byte
			k, b = cursor.Last()
			if len(k) == 0 {
				return fmt.Errorf("no matching auth condition for address could be found: address %s was never authorized", address.String())
			}

			var authState bool
			err := rivbin.Unmarshal(b, &authState)
			if err != nil {
				return fmt.Errorf("failed to decode found auth condition for address %s: %v", address.String(), err)
			}

			if !authState {
				return fmt.Errorf("address %s is not authorized", address.String())
			}
		}
		return nil
	})
}

// EnsureAddressesAreAuthAt implements types.AuthInfoGetter.EnsureAddressesAreAuthAt
func (p *Plugin) EnsureAddressesAreAuthAt(height types.BlockHeight, addresses ...types.UnlockHash) error {
	return p.storage.View(func(bucket *bolt.Bucket) error {
		authBucket := bucket.Bucket([]byte(bucketAuthAddresses))
		var b []byte
		for _, address := range addresses {
			addressAuthBucket := authBucket.Bucket(rivbin.Marshal(address))
			if addressAuthBucket == nil {
				return fmt.Errorf("auth address (%s) condition bucket does not exist: address was never authorized", address.String())
			}
			cursor := addressAuthBucket.Cursor()

			var k []byte
			k, b = cursor.Seek(encodeBlockheight(height))
			if len(k) == 0 {
				// could be that we're past the last key, let's try the last key first
				k, b = cursor.Last()
				if len(k) == 0 {
					return fmt.Errorf("corrupt plugin DB: no matching address (%s) auth condition could be found", address.String())
				}
				return nil
			}
			foundHeight := decodeBlockheight(k)
			if foundHeight <= height {
				return nil
			}
			k, b = cursor.Prev()
			if len(k) == 0 {
				return fmt.Errorf("corrupt plugin DB: no matching address (%s) auth condition could be found", address.String())
			}
			var authState bool
			err := rivbin.Unmarshal(b, &authState)
			if err != nil {
				return fmt.Errorf("failed to decode found address (%s) auth condition for address: %v", address.String(), err)
			}

			if !authState {
				return fmt.Errorf("address %s is not authorized at heigh %d", address.String(), height)
			}
		}
		return nil
	})
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
