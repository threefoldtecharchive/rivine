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
		genesisAuthCondition                         types.UnlockConditionProxy
		authAddressUpdateTransactionVersion          types.TransactionVersion
		authConditionUpdateTransactionVersion        types.TransactionVersion
		storage                                      modules.PluginViewStorage
		unregisterCallback                           modules.PluginUnregisterCallback
		unauthorizedCoinTransactionExceptionCallback UnauthorizedCoinTransactionExceptionCallback
		unlockHashFilter                             func(types.UnlockHash) bool
		RequireMinerFees                             bool
		reverse                                      bool
	}

	// PluginOpts are extra optional configurations one can make to the AuthCoin Plugin
	PluginOpts struct {
		// UnauthorizedCoinTransactionExceptionCallback is a callback that can be defined,
		// in case your chain requires custom logic to define what transaction can be considered valid for
		// coin transfers with unauthorized addresses due to whatever rules (e.g. version, pure refund coin flow, ...)
		UnauthorizedCoinTransactionExceptionCallback UnauthorizedCoinTransactionExceptionCallback

		// UnlockHashFilter can be used to filter what unlock hashes require authorization and which not.
		// It is optional, and if none given, all unlockhashes except
		// the NilUnlockHash and AtomicSwap contract addresses require authorization.
		// Returns true in case authorization cheque is required, False otherwise.
		UnlockHashFilter func(types.UnlockHash) bool
		// RequireMinerFees can be used to enable minerfees on authorization transactions.
		RequireMinerFees bool
		// Reverse can be used to reverse the logic of this plugin.
		// This means that by default all addresses are authorized and can send transactions.
		Reverse bool
	}

	// UnauthorizedCoinTransactionExceptionCallback is the function signature for the callback that can be used
	// for chains that requires custom logic to define what transaction can be considered valid for
	// coin transfers with unauthorized addresses due to whatever rules (e.g. version, pure refund coin flow, ...)
	// True is returned in case this tx does not require an authorization check, False otherwise.
	UnauthorizedCoinTransactionExceptionCallback func(tx modules.ConsensusTransaction, dedupAddresses []types.UnlockHash, ctx types.TransactionValidationContext) (bool, error)
)

// DefaultUnauthorizedCoinTransactionExceptionCallback is the default callback that is used in ase the auth coin plugin
// does not define a custom callback.
func DefaultUnauthorizedCoinTransactionExceptionCallback(tx modules.ConsensusTransaction, dedupAddresses []types.UnlockHash, ctx types.TransactionValidationContext) (bool, error) {
	if tx.Version != types.TransactionVersionZero && tx.Version != types.TransactionVersionOne {
		return false, nil
	}
	return (len(dedupAddresses) == 1 && len(tx.CoinOutputs) <= 1), nil
}

// NewPlugin creates a new Plugin with a genesisAuthCondition and correct transaction versions.
// NOTE: do not register the default rivine transaction versions (0x00 and 0x01) if you use this plugin!!!
func NewPlugin(genesisAuthCondition types.UnlockConditionProxy, authAddressUpdateTransactionVersion, authConditionUpdateTransactionVersion types.TransactionVersion, opts *PluginOpts) *Plugin {
	p := &Plugin{
		genesisAuthCondition:                  genesisAuthCondition,
		authAddressUpdateTransactionVersion:   authAddressUpdateTransactionVersion,
		authConditionUpdateTransactionVersion: authConditionUpdateTransactionVersion,
		RequireMinerFees:                      opts != nil && opts.RequireMinerFees,
	}
	if opts != nil && opts.UnauthorizedCoinTransactionExceptionCallback != nil {
		p.unauthorizedCoinTransactionExceptionCallback = opts.UnauthorizedCoinTransactionExceptionCallback
	} else {
		p.unauthorizedCoinTransactionExceptionCallback = DefaultUnauthorizedCoinTransactionExceptionCallback
	}
	if opts != nil && opts.Reverse {
		p.reverse = opts.Reverse
	}
	if opts != nil && opts.UnlockHashFilter != nil {
		p.unlockHashFilter = opts.UnlockHashFilter
	} else {
		p.unlockHashFilter = func(uh types.UnlockHash) bool {
			return uh.Type != types.UnlockTypeNil && uh.Type != types.UnlockTypeAtomicSwap
		}
	}
	types.RegisterTransactionVersion(authAddressUpdateTransactionVersion, AuthAddressUpdateTransactionController{
		AuthAddressBaseTransactionController: AuthAddressBaseTransactionController{
			RequireMinerFees: p.RequireMinerFees,
		},
		AuthInfoGetter:     p,
		TransactionVersion: authAddressUpdateTransactionVersion,
	})
	types.RegisterTransactionVersion(authConditionUpdateTransactionVersion, AuthConditionUpdateTransactionController{
		AuthAddressBaseTransactionController: AuthAddressBaseTransactionController{
			RequireMinerFees: p.RequireMinerFees,
		},
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

		mintcond, err := rivbin.Marshal(p.genesisAuthCondition)
		if err != nil {
			return persist.Metadata{}, fmt.Errorf("failed to (rivbin) marshal genesis auth condition: %v", err)
		}
		err = authBucket.Put(encodeBlockheight(0), mintcond)
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
func (p *Plugin) ApplyBlock(block modules.ConsensusBlock, bucket *persist.LazyBoltBucket) error {
	if bucket == nil {
		return errors.New("plugin bucket does not exist")
	}
	var err error
	for idx, txn := range block.Transactions {
		cTxn := modules.ConsensusTransaction{
			Transaction:            txn,
			BlockHeight:            block.Height,
			BlockTime:              block.Timestamp,
			SequenceID:             uint16(idx),
			SpentCoinOutputs:       block.SpentCoinOutputs,
			SpentBlockStakeOutputs: block.SpentBlockStakeOutputs,
		}
		err = p.ApplyTransaction(cTxn, bucket)
		if err != nil {
			return err
		}
	}
	return nil
}

// ApplyBlockHeader applies nothing and has no effect on this plugin.
func (p *Plugin) ApplyBlockHeader(modules.ConsensusBlockHeader, *persist.LazyBoltBucket) error {
	return nil
}

// ApplyTransaction applies a minting transactions to the minting bucket.
func (p *Plugin) ApplyTransaction(txn modules.ConsensusTransaction, bucket *persist.LazyBoltBucket) error {
	if bucket == nil {
		return errors.New("plugin bucket does not exist")
	}
	// check the version and handle the ones we care about
	switch txn.Version {
	case p.authConditionUpdateTransactionVersion:
		acutx, err := AuthConditionUpdateTransactionFromTransaction(txn.Transaction, p.authConditionUpdateTransactionVersion, p.RequireMinerFees)
		if err != nil {
			return fmt.Errorf("unexpected error while unpacking the auth condition update tx type: %v" + err.Error())
		}
		authBucket, err := bucket.Bucket([]byte(bucketAuthConditions))
		if err != nil {
			return errors.New("auth conditions bucket does not exist")
		}
		authConditionBytes, err := rivbin.Marshal(acutx.AuthCondition)
		if err != nil {
			return fmt.Errorf(
				"failed to (rivbin) marshal auth condition for block height %d: %v",
				txn.BlockHeight, err)
		}
		err = authBucket.Put(encodeBlockheight(txn.BlockHeight), authConditionBytes)
		if err != nil {
			return fmt.Errorf(
				"failed to put auth condition for block height %d: %v",
				txn.BlockHeight, err)
		}

	case p.authAddressUpdateTransactionVersion:
		aautx, err := AuthAddressUpdateTransactionFromTransaction(txn.Transaction, p.authAddressUpdateTransactionVersion, p.RequireMinerFees)
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
			addressBytes, err := rivbin.Marshal(address)
			if err != nil {
				return fmt.Errorf("failed to (rivbin) marshal auth address: %v", err)
			}
			addressAuthBucket, err := authBucket.CreateBucketIfNotExists(addressBytes)
			if err != nil {
				return fmt.Errorf("auth address (%s) condition bucket does not exist and could not be created: %v", address.String(), err)
			}
			stateBytes, err := rivbin.Marshal(true)
			if err != nil {
				return fmt.Errorf("failed to (rivbin) marshal auth address state (true): %v", err)
			}
			err = addressAuthBucket.Put(encodeBlockheight(txn.BlockHeight), stateBytes)
			if err != nil {
				return fmt.Errorf(
					"failed to put auth condition for address %s block height %d: %v",
					address.String(), txn.BlockHeight, err)
			}
		}
		for _, address := range aautx.DeauthAddresses {
			addressBytes, err := rivbin.Marshal(address)
			if err != nil {
				return fmt.Errorf("failed to (rivbin) marshal auth address: %v", err)
			}
			addressAuthBucket, err := authBucket.CreateBucketIfNotExists(addressBytes)
			if err != nil {
				return fmt.Errorf("auth address (%s) condition bucket does not exist and could not be created: %v", address.String(), err)
			}
			stateBytes, err := rivbin.Marshal(false)
			if err != nil {
				return fmt.Errorf("failed to (rivbin) marshal auth address state (false): %v", err)
			}
			err = addressAuthBucket.Put(encodeBlockheight(txn.BlockHeight), stateBytes)
			if err != nil {
				return fmt.Errorf(
					"failed to put deauth condition for address %s block height %d: %v",
					address.String(), txn.BlockHeight, err)
			}
		}
	}
	return nil
}

// RevertBlock reverts a block's minting transaction from the minting bucket
func (p *Plugin) RevertBlock(block modules.ConsensusBlock, bucket *persist.LazyBoltBucket) error {
	if bucket == nil {
		return errors.New("plugin bucket does not exist")
	}
	var err error
	for idx, txn := range block.Transactions {
		cTxn := modules.ConsensusTransaction{
			Transaction:            txn,
			BlockHeight:            block.Height,
			BlockTime:              block.Timestamp,
			SequenceID:             uint16(idx),
			SpentCoinOutputs:       block.SpentCoinOutputs,
			SpentBlockStakeOutputs: block.SpentBlockStakeOutputs,
		}
		err = p.RevertTransaction(cTxn, bucket)
		if err != nil {
			return err
		}
	}
	return nil
}

// RevertBlockHeader reverts nothing and has no effect on this plugin.
func (p *Plugin) RevertBlockHeader(modules.ConsensusBlockHeader, *persist.LazyBoltBucket) error {
	return nil
}

// RevertTransaction reverts a  minting transaction from the minting bucket
func (p *Plugin) RevertTransaction(txn modules.ConsensusTransaction, bucket *persist.LazyBoltBucket) error {
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
		err = authBucket.Delete(encodeBlockheight(txn.BlockHeight))
		if err != nil {
			return fmt.Errorf(
				"failed to delete auth condition for block height %d: %v",
				txn.BlockHeight, err)
		}

	case p.authAddressUpdateTransactionVersion:
		aautx, err := AuthAddressUpdateTransactionFromTransaction(txn.Transaction, p.authAddressUpdateTransactionVersion, p.RequireMinerFees)
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
			addressBytes, err := rivbin.Marshal(address)
			if err != nil {
				return fmt.Errorf("cannot get auth address (%s) condition bucket: failed to (rivbin) marshal auth address: %v", address.String(), err)
			}
			addressAuthBucket := authBucket.Bucket(addressBytes)
			if addressAuthBucket == nil {
				return fmt.Errorf("auth address (%s) condition bucket does not exist", address.String())
			}
			err = addressAuthBucket.Delete(encodeBlockheight(txn.BlockHeight))
			if err != nil {
				return fmt.Errorf(
					"failed to delete auth condition for address %s block height %d: %v",
					address.String(), txn.BlockHeight, err)
			}
		}
		for _, address := range aautx.DeauthAddresses {
			addressBytes, err := rivbin.Marshal(address)
			if err != nil {
				return fmt.Errorf("cannot get auth address (%s) condition bucket: failed to (rivbin) marshal auth address: %v", address.String(), err)
			}
			addressAuthBucket := authBucket.Bucket(addressBytes)
			if addressAuthBucket == nil {
				return fmt.Errorf("auth address (%s) condition bucket does not exist", address.String())
			}
			err = addressAuthBucket.Delete(encodeBlockheight(txn.BlockHeight))
			if err != nil {
				return fmt.Errorf(
					"failed to delete deauth condition for address %s block height %d: %v",
					address.String(), txn.BlockHeight, err)
			}
		}
	}
	return nil
}

// GetActiveAuthCondition implements types.AuthInfoGetter.GetActiveAuthCondition
func (p *Plugin) GetActiveAuthCondition() (types.UnlockConditionProxy, error) {
	var authCondition types.UnlockConditionProxy
	err := p.storage.View(func(bucket *bolt.Bucket) (err error) {
		authBucket := bucket.Bucket([]byte(bucketAuthConditions))
		if authBucket == nil {
			return errors.New("auth condition bucket could not be found")
		}
		authCondition, err = p.getAuthConditionFromBucket(authBucket)
		return err
	})
	if err != nil {
		return types.UnlockConditionProxy{}, err
	}
	return authCondition, nil
}

// GetAuthConditionAt implements types.AuthInfoGetter.GetAuthConditionAt
func (p *Plugin) GetAuthConditionAt(height types.BlockHeight) (types.UnlockConditionProxy, error) {
	var authCondition types.UnlockConditionProxy
	err := p.storage.View(func(bucket *bolt.Bucket) (err error) {
		authBucket := bucket.Bucket([]byte(bucketAuthConditions))
		if authBucket == nil {
			return errors.New("auth condition bucket could not be found")
		}
		authCondition, err = p.getAuthConditionFromBucketAt(authBucket, height)
		return err
	})
	if err != nil {
		return types.UnlockConditionProxy{}, err
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
		var err error
		for index, address := range addresses {
			if !p.unlockHashFilter(address) {
				stateSlice[index] = true
			} else {
				stateSlice[index], err = p.getAuthAddressStateFromBucket(authBucket, address)
				if err != nil {
					return err
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
			if !p.unlockHashFilter(address) {
				stateSlice[index] = true
			} else {
				stateSlice[index], err = p.getAuthAddressStateFromBucket(authBucket, address)
				if err != nil {
					return err
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

func (p *Plugin) getAuthConditionFromBucketAt(authConditionBucket *bolt.Bucket, height types.BlockHeight) (types.UnlockConditionProxy, error) {
	var b []byte
	err := func() error {
		cursor := authConditionBucket.Cursor()
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
	}()
	if err != nil {
		return types.UnlockConditionProxy{}, err
	}
	var authCondition types.UnlockConditionProxy
	err = rivbin.Unmarshal(b, &authCondition)
	if err != nil {
		return types.UnlockConditionProxy{}, fmt.Errorf("corrupt transaction DB: failed to decode found auth condition: %v", err)
	}
	return authCondition, nil
}

func (p *Plugin) getAuthConditionFromBucket(authConditionBucket *bolt.Bucket) (types.UnlockConditionProxy, error) {
	cursor := authConditionBucket.Cursor()
	k, b := cursor.Last()
	if len(k) == 0 {
		return types.UnlockConditionProxy{}, errors.New("no matching auth condition could be found")
	}
	var authCondition types.UnlockConditionProxy
	err := rivbin.Unmarshal(b, &authCondition)
	if err != nil {
		return types.UnlockConditionProxy{}, fmt.Errorf("failed to decode found auth condition: %v", err)
	}
	return authCondition, nil
}

func (p *Plugin) getAuthConditionFromBucketWithContextInfo(authConditionBucket *bolt.Bucket, confirmed bool, blockHeight types.BlockHeight) (types.UnlockConditionProxy, error) {
	if confirmed || blockHeight > 0 {
		mintCondition, err := p.getAuthConditionFromBucketAt(authConditionBucket, blockHeight)
		if err != nil {
			return types.UnlockConditionProxy{}, fmt.Errorf("failed to get auth condition at block height %d", blockHeight)
		}
		return mintCondition, nil
	}
	mintCondition, err := p.getAuthConditionFromBucket(authConditionBucket)
	if err != nil {
		return types.UnlockConditionProxy{}, errors.New("failed to get the latest auth condition")
	}
	return mintCondition, nil
}

func (p *Plugin) getAuthAddressStateFromBucketAt(authAddressBucket *bolt.Bucket, uh types.UnlockHash, height types.BlockHeight) (bool, error) {
	var b []byte
	err := func() error {
		uhBytes, err := rivbin.Marshal(uh)
		if err != nil {
			return fmt.Errorf("failed to (rivbin) marshal unlockhash %s: %v", uh.String(), err)
		}
		addressAuthBucket := authAddressBucket.Bucket(uhBytes)
		if addressAuthBucket == nil {
			return nil // nothing to do, state will be fals by default
		}
		cursor := addressAuthBucket.Cursor()
		var k []byte
		k, b = cursor.Seek(encodeBlockheight(height))
		if len(k) == 0 {
			// could be that we're past the last key, let's try the last key first
			k, b = cursor.Last()
			if len(k) == 0 {
				b = nil
				return nil
			}
		}
		foundHeight := decodeBlockheight(k)
		if foundHeight > height {
			k, b = cursor.Prev()
			if len(k) == 0 {
				b = nil
				return nil
			}
		}
		return nil
	}()
	if err != nil {
		return false, err
	}
	if b == nil {
		return false, nil
	}
	var state bool
	err = rivbin.Unmarshal(b, &state)
	if err != nil {
		return false, fmt.Errorf("corrupt transaction DB: failed to decode found address auth state for address %s at height %d: %v", uh.String(), height, err)
	}
	if p.reverse {
		return !state, nil
	}
	return state, nil
}

func (p *Plugin) getAuthAddressStateFromBucket(authAddressBucket *bolt.Bucket, uh types.UnlockHash) (bool, error) {
	var b []byte
	err := func() error {
		uhBytes, err := rivbin.Marshal(uh)
		if err != nil {
			return fmt.Errorf("failed to (rivbin) marshal unlockhash %s: %v", uh.String(), err)
		}
		addressAuthBucket := authAddressBucket.Bucket(uhBytes)
		if addressAuthBucket == nil {
			return nil
		}
		// return the last cursor
		cursor := addressAuthBucket.Cursor()
		var k []byte
		k, b = cursor.Last()
		if len(k) == 0 {
			b = nil
		}
		return nil
	}()
	if err != nil {
		return false, err
	}
	if b == nil {
		return false, nil
	}
	var state bool
	err = rivbin.Unmarshal(b, &state)
	if err != nil {
		return false, fmt.Errorf("corrupt transaction DB: failed to decode found address auth state for address %s: %v", uh.String(), err)
	}

	if p.reverse {
		return !state, nil
	}
	return state, nil
}

func (p *Plugin) getAuthAddressStateFromBucketWithContextInfo(authAddressBucket *bolt.Bucket, uh types.UnlockHash, confirmed bool, blockHeight types.BlockHeight) (bool, error) {
	if confirmed || blockHeight > 0 {
		state, err := p.getAuthAddressStateFromBucketAt(authAddressBucket, uh, blockHeight)
		if err != nil {
			return false, fmt.Errorf("failed to get auth address state for address %s at block height %d", uh.String(), blockHeight)
		}
		if p.reverse {
			return !state, nil
		}
		return state, nil
	}
	state, err := p.getAuthAddressStateFromBucket(authAddressBucket, uh)
	if err != nil {
		return false, fmt.Errorf("failed to get the latest auth address state for address %s", uh.String())
	}
	if p.reverse {
		return !state, nil
	}
	return state, nil
}

// TransactionValidatorVersionFunctionMapping returns all tx validators for specific tx versions linked to this plugin
func (p *Plugin) TransactionValidatorVersionFunctionMapping() map[types.TransactionVersion][]modules.PluginTransactionValidationFunction {
	return map[types.TransactionVersion][]modules.PluginTransactionValidationFunction{
		p.authAddressUpdateTransactionVersion: []modules.PluginTransactionValidationFunction{
			p.validateAuthAddressUpdateTx,
		},
		p.authConditionUpdateTransactionVersion: []modules.PluginTransactionValidationFunction{
			p.validateAuthConditionUpdateTx,
		},
	}
}

// TransactionValidators returns all tx validators linked to this plugin
func (p *Plugin) TransactionValidators() []modules.PluginTransactionValidationFunction {
	return []modules.PluginTransactionValidationFunction{
		p.validateAuthorizedCoinFlowForAllTxs,
	}
}

func (p *Plugin) validateAuthorizedCoinFlowForAllTxs(tx modules.ConsensusTransaction, ctx types.TransactionValidationContext, bucket *persist.LazyBoltBucket) error {
	// collect all dedupAddresses
	dedupAddresses := map[types.UnlockHash]struct{}{}
	for _, co := range tx.CoinOutputs {
		dedupAddresses[co.Condition.UnlockHash()] = struct{}{}
	}
	for _, ci := range tx.CoinInputs {
		co, ok := tx.SpentCoinOutputs[ci.ParentID]
		if !ok {
			return fmt.Errorf(
				"unable to find parent ID %s as an unspent coin output in the current consensus transaction at block height %d",
				ci.ParentID.String(), ctx.BlockHeight)
		}
		dedupAddresses[co.Condition.UnlockHash()] = struct{}{}
	}

	addressLength := len(dedupAddresses)
	if addressLength == 0 {
		return nil // nothing to do
	}
	dedupAddressesSlice := make([]types.UnlockHash, 0, len(dedupAddresses))
	for uh := range dedupAddresses {
		dedupAddressesSlice = append(dedupAddressesSlice, uh)
	}
	allowedToBeUnauthorized, err := p.unauthorizedCoinTransactionExceptionCallback(tx, dedupAddressesSlice, ctx)
	if err != nil {
		return fmt.Errorf("failed to check if transaction is allowed to be a potential unauthorized coin transfer: %v", err)
	}
	if allowedToBeUnauthorized {
		return nil // nothing to validate, whether it is authorized or not is no longer important
	}

	// get authAddressBucket from plugin bucket, so we can check the state of addresses
	authAddressBucket, err := bucket.Bucket(bucketAuthAddresses)
	if err != nil {
		return err
	}
	// validate that all used addresses are authorized
	for addr := range dedupAddresses {
		if !p.unlockHashFilter(addr) {
			continue
		}
		state, err := p.getAuthAddressStateFromBucketWithContextInfo(authAddressBucket, addr, ctx.Confirmed, ctx.BlockHeight)
		if err != nil {
			return fmt.Errorf("failed to check if address %s is authorized at the moment: %v", addr.String(), err)
		}
		if !state {
			return types.NewClientError(fmt.Errorf("address %s is not authorized", addr), types.ClientErrorForbidden)
		}
	}
	return nil
}

func (p *Plugin) validateAuthAddressUpdateTx(tx modules.ConsensusTransaction, ctx types.TransactionValidationContext, bucket *persist.LazyBoltBucket) error {
	// get AuthAddressUpdateTx
	autx, err := AuthAddressUpdateTransactionFromTransaction(tx.Transaction, p.authAddressUpdateTransactionVersion, p.RequireMinerFees)
	if err != nil {
		// this check also fails if the tx contains coin/blockstake inputs/outputs or miner fees
		return fmt.Errorf("failed to use tx as a auth address update tx: %v", err)
	}

	// ensure the Nonce is not Nil
	if autx.Nonce == (types.TransactionNonce{}) {
		return errors.New("nil nonce is not allowed for a auth address update transaction")
	}

	// get AuthCondition from auth bucket
	authConditionBucket, err := bucket.Bucket(bucketAuthConditions)
	if err != nil {
		return err
	}
	authCondition, err := p.getAuthConditionFromBucketWithContextInfo(authConditionBucket, ctx.Confirmed, ctx.BlockHeight)
	if err != nil {
		return fmt.Errorf("failed to get auth condition at block height %d: %v", ctx.BlockHeight, err)
	}
	// check if AuthFulfillment fulfills the Globally defined AuthCondition for the context-defined block height
	err = authCondition.Fulfill(autx.AuthFulfillment, types.FulfillContext{
		BlockHeight: ctx.BlockHeight,
		BlockTime:   ctx.BlockTime,
		Transaction: tx.Transaction,
	})
	if err != nil {
		return types.NewClientError(fmt.Errorf("cannot update address states: failed to fulfill auth condition: %v", err), types.ClientErrorUnauthorized)
	}

	// ensure we have at least one address to (de)authorize
	lAuthAddresses := len(autx.AuthAddresses)
	lDeauthAddresses := len(autx.DeauthAddresses)
	if lAuthAddresses == 0 && lDeauthAddresses == 0 {
		return errors.New("at least one address is required to be authorized or deauthorized")
	}
	// ensure all addresses are unique and thus also that no address is both authorized and deauthorized
	addressesSeen := map[types.UnlockHash]struct{}{}
	var ok bool
	for _, address := range autx.AuthAddresses {
		if _, ok = addressesSeen[address]; ok {
			return fmt.Errorf("an address can only be defined once per AuthAddressUpdate transaction: %s was seen twice", address.String())
		}
		addressesSeen[address] = struct{}{}
	}
	for _, address := range autx.DeauthAddresses {
		if _, ok = addressesSeen[address]; ok {
			return fmt.Errorf("an address can only be defined once per AuthAddressUpdate transaction: %s was seen twice", address.String())
		}
		addressesSeen[address] = struct{}{}
	}
	// get authAddressBucket from plugin bucket, so we can check the state of addresses
	authAddressBucket, err := bucket.Bucket(bucketAuthAddresses)
	if err != nil {
		return err
	}
	// ensure all address to be authorized are currently deauthorized
	for _, addr := range autx.AuthAddresses {
		state, err := p.getAuthAddressStateFromBucketWithContextInfo(authAddressBucket, addr, ctx.Confirmed, ctx.BlockHeight)
		if err != nil {
			return fmt.Errorf("failed to check if address %s is deauthorized at the moment: %v", addr.String(), err)
		}
		if state {
			return types.NewClientError(fmt.Errorf("address %s (to auth) is already authorized", addr), types.ClientErrorForbidden)
		}
	}
	// ensure all address to be deauthroized are currently authorized
	for _, addr := range autx.DeauthAddresses {
		state, err := p.getAuthAddressStateFromBucketWithContextInfo(authAddressBucket, addr, ctx.Confirmed, ctx.BlockHeight)
		if err != nil {
			return fmt.Errorf("failed to check if address %s is authorized at the moment: %v", addr.String(), err)
		}
		if !state {
			return types.NewClientError(fmt.Errorf("address %s (to auth) is already deauthorized", addr), types.ClientErrorForbidden)
		}
	}

	return nil // tx is valid according to this tx validator
}

func (p *Plugin) validateAuthConditionUpdateTx(tx modules.ConsensusTransaction, ctx types.TransactionValidationContext, bucket *persist.LazyBoltBucket) error {
	// get AuthConditionUpdateTx
	cutx, err := AuthConditionUpdateTransactionFromTransaction(tx.Transaction, p.authConditionUpdateTransactionVersion, p.RequireMinerFees)
	if err != nil {
		// this check also fails if the tx contains coin/blockstake inputs/outputs or miner fees
		return fmt.Errorf("failed to use tx as a auth condition update tx: %v", err)
	}

	// ensure the Nonce is not Nil
	if cutx.Nonce == (types.TransactionNonce{}) {
		return errors.New("nil nonce is not allowed for a auth condition update transaction")
	}

	// get AuthCondition from auth bucket
	authConditionBucket, err := bucket.Bucket(bucketAuthConditions)
	if err != nil {
		return err
	}
	authCondition, err := p.getAuthConditionFromBucketWithContextInfo(authConditionBucket, ctx.Confirmed, ctx.BlockHeight)
	if err != nil {
		return fmt.Errorf("failed to get auth condition at block height %d: %v", ctx.BlockHeight, err)
	}

	// ensure the defined condition is not equal to the current active auth condition
	if authCondition.Equal(cutx.AuthCondition) {
		return errors.New("defined condition is already used as the currently active auth condition (nop update not allowed)")
	}
	// ensure the defined condition maps to an acceptable uh
	uh := cutx.AuthCondition.UnlockHash()
	if uh.Type != types.UnlockTypePubKey && uh.Type != types.UnlockTypeMultiSig {
		return fmt.Errorf("defined condition maps to an invalid unlock hash type %d", uh.Type)
	}

	// check if AuthFulfillment fulfills the Globally defined AuthCondition for the context-defined block height
	err = authCondition.Fulfill(cutx.AuthFulfillment, types.FulfillContext{
		BlockHeight: ctx.BlockHeight,
		BlockTime:   ctx.BlockTime,
		Transaction: tx.Transaction,
	})
	if err != nil {
		return types.NewClientError(fmt.Errorf("failed to fulfill auth condition: %v", err), types.ClientErrorUnauthorized)
	}

	return nil // tx is valid according to this tx validator
}

// Close unregisters the plugin from the consensus
func (p *Plugin) Close() error {
	if p.storage == nil {
		return nil
	}
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
