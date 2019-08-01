package minting

import (
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/threefoldtech/rivine/build"
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
		coinCreationTransactionVersion     types.TransactionVersion
		coinDestructionTransactionVersion  *types.TransactionVersion
		storage                            modules.PluginViewStorage
		unregisterCallback                 modules.PluginUnregisterCallback

		binMarshal   func(v interface{}) ([]byte, error)
		binUnmarshal func(b []byte, v interface{}) error

		requireMinerFees bool
	}
)

type (
	// PluginOptions allows optional parameters to be defined for the minting plugin.
	PluginOptions struct {
		CoinDestructionTransactionVersion types.TransactionVersion
		UseLegacySiaEncoding              bool
		RequireMinerFees                  bool
	}
)

// NewMintingPlugin creates a new Plugin with a genesisMintCondition and correct transaction versions
func NewMintingPlugin(genesisMintCondition types.UnlockConditionProxy, minterDefinitionTransactionVersion, coinCreationTransactionVersion types.TransactionVersion, opts *PluginOptions) *Plugin {
	p := &Plugin{
		genesisMintCondition:               genesisMintCondition,
		minterDefinitionTransactionVersion: minterDefinitionTransactionVersion,
		coinCreationTransactionVersion:     coinCreationTransactionVersion,
		requireMinerFees:                   opts != nil && opts.RequireMinerFees,
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
		p.coinDestructionTransactionVersion = &opts.CoinDestructionTransactionVersion
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

		mintcond, err := p.binMarshal(p.genesisMintCondition)
		if err != nil {
			return persist.Metadata{}, fmt.Errorf("failed to marshal genesis mint condition: %v", err)
		}
		err = mintingBucket.Put(encodeBlockheight(0), mintcond)
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
func (p *Plugin) ApplyBlock(block modules.ConsensusBlock, height types.BlockHeight, bucket *persist.LazyBoltBucket) error {
	if bucket == nil {
		return errors.New("minting bucket does not exist")
	}
	var err error
	for _, txn := range block.Transactions {
		cTxn := modules.ConsensusTransaction{
			Transaction:            txn,
			SpentCoinOutputs:       block.SpentCoinOutputs,
			SpentBlockStakeOutputs: block.SpentBlockStakeOutputs,
		}
		err = p.ApplyTransaction(cTxn, height, bucket)
		if err != nil {
			return err
		}
	}
	return nil
}

// ApplyTransaction applies a minting transactions to the minting bucket.
func (p *Plugin) ApplyTransaction(txn modules.ConsensusTransaction, height types.BlockHeight, bucket *persist.LazyBoltBucket) error {
	if bucket == nil {
		return errors.New("minting bucket does not exist")
	}
	var (
		mintingBucket *bolt.Bucket
	)
	// check the version and handle the ones we care about
	switch txn.Version {
	case p.minterDefinitionTransactionVersion:
		mdtx, err := MinterDefinitionTransactionFromTransaction(txn.Transaction, p.minterDefinitionTransactionVersion, p.requireMinerFees)
		if err != nil {
			return fmt.Errorf("unexpected error while unpacking the minter def. tx type: %v" + err.Error())
		}
		if mintingBucket == nil {
			mintingBucket, err = bucket.Bucket([]byte(bucketMintConditions))
			if err != nil {
				return errors.New("mintcondition bucket does not exist")
			}
		}
		mintcond, err := p.binMarshal(mdtx.MintCondition)
		if err != nil {
			return fmt.Errorf("failed to marshal mint condition: %v", err)
		}
		err = mintingBucket.Put(encodeBlockheight(height), mintcond)
		if err != nil {
			return fmt.Errorf(
				"failed to put mint condition for block height %d: %v",
				height, err)
		}
	}
	return nil
}

// RevertBlock reverts a block's minting transaction from the minting bucket
func (p *Plugin) RevertBlock(block modules.ConsensusBlock, height types.BlockHeight, bucket *persist.LazyBoltBucket) error {
	if bucket == nil {
		return errors.New("mint conditions bucket does not exist")
	}
	// collect all one-per-block mint conditions
	var err error
	for _, txn := range block.Transactions {
		cTxn := modules.ConsensusTransaction{
			Transaction:            txn,
			SpentCoinOutputs:       block.SpentCoinOutputs,
			SpentBlockStakeOutputs: block.SpentBlockStakeOutputs,
		}
		err = p.RevertTransaction(cTxn, height, bucket)
		if err != nil {
			return err
		}
	}
	return nil
}

// RevertTransaction reverts a minting transactions to the minting bucket.
func (p *Plugin) RevertTransaction(txn modules.ConsensusTransaction, height types.BlockHeight, bucket *persist.LazyBoltBucket) error {
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
		mintingBucket := bucket.Bucket([]byte(bucketMintConditions))
		if mintingBucket == nil {
			return errors.New("no minting condition bucket found")
		}
		var err error
		mintCondition, err = p.getMintConditionFromBucket(mintingBucket)
		return err

	})
	return mintCondition, err
}

// GetMintConditionAt implements types.MintConditionGetter.GetMintConditionAt
func (p *Plugin) GetMintConditionAt(height types.BlockHeight) (types.UnlockConditionProxy, error) {
	var mintCondition types.UnlockConditionProxy
	err := p.storage.View(func(bucket *bolt.Bucket) error {
		mintingBucket := bucket.Bucket([]byte(bucketMintConditions))
		if mintingBucket == nil {
			return errors.New("no minting condition bucket found")
		}
		var err error
		mintCondition, err = p.getMintConditionFromBucketAt(mintingBucket, height)
		return err

	})
	return mintCondition, err
}

func (p *Plugin) getMintConditionFromBucketAt(mintConditionBucket *bolt.Bucket, height types.BlockHeight) (types.UnlockConditionProxy, error) {
	var b []byte
	err := func() error {
		cursor := mintConditionBucket.Cursor()
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
	}()
	if err != nil {
		return types.UnlockConditionProxy{}, err
	}
	var mintCondition types.UnlockConditionProxy
	err = p.binUnmarshal(b, &mintCondition)
	if err != nil {
		return types.UnlockConditionProxy{}, fmt.Errorf("corrupt transaction DB: failed to decode found mint condition: %v", err)
	}
	return mintCondition, nil
}

func (p *Plugin) getMintConditionFromBucket(mintConditionBucket *bolt.Bucket) (types.UnlockConditionProxy, error) {
	cursor := mintConditionBucket.Cursor()
	k, b := cursor.Last()
	if len(k) == 0 {
		return types.UnlockConditionProxy{}, errors.New("no matching mint condition could be found")
	}
	var mintCondition types.UnlockConditionProxy
	err := p.binUnmarshal(b, &mintCondition)
	if err != nil {
		return types.UnlockConditionProxy{}, fmt.Errorf("failed to decode found mint condition: %v", err)
	}
	return mintCondition, nil
}

func (p *Plugin) getMintConditionFromBucketWithContextInfo(mintConditionBucket *bolt.Bucket, confirmed bool, blockHeight types.BlockHeight) (types.UnlockConditionProxy, error) {
	if confirmed || blockHeight > 0 {
		mintCondition, err := p.getMintConditionFromBucketAt(mintConditionBucket, blockHeight)
		if err != nil {
			return types.UnlockConditionProxy{}, fmt.Errorf("failed to get mint condition at block height %d", blockHeight)
		}
		return mintCondition, nil
	}
	mintCondition, err := p.getMintConditionFromBucket(mintConditionBucket)
	if err != nil {
		return types.UnlockConditionProxy{}, errors.New("failed to get the latest mint condition")
	}
	return mintCondition, nil
}

// TransactionValidatorVersionFunctionMapping returns all tx validators linked to this plugin
func (p *Plugin) TransactionValidatorVersionFunctionMapping() map[types.TransactionVersion][]modules.PluginTransactionValidationFunction {
	m := map[types.TransactionVersion][]modules.PluginTransactionValidationFunction{
		p.minterDefinitionTransactionVersion: []modules.PluginTransactionValidationFunction{
			p.validateMinterDefinitionTx,
		},
		p.coinCreationTransactionVersion: []modules.PluginTransactionValidationFunction{
			p.validateCoinCreationTx,
		},
	}
	if p.coinDestructionTransactionVersion != nil {
		m[*p.coinDestructionTransactionVersion] = []modules.PluginTransactionValidationFunction{
			p.validateCoinDestructionTxCreation,
		}
	}
	return m
}

// TransactionValidators returns all tx validators linked to this plugin
func (p *Plugin) TransactionValidators() []modules.PluginTransactionValidationFunction {
	return nil
}

func (p *Plugin) validateMinterDefinitionTx(tx modules.ConsensusTransaction, ctx types.TransactionValidationContext, bucket *persist.LazyBoltBucket) error {
	mdtx, err := MinterDefinitionTransactionFromTransaction(tx.Transaction, p.minterDefinitionTransactionVersion, p.requireMinerFees)
	if err != nil {
		return fmt.Errorf("failed to use tx as a minter definition tx: %v", err)
	}

	// ensure the Nonce is not Nil
	if mdtx.Nonce == (types.TransactionNonce{}) {
		return errors.New("nil nonce is not allowed for a minter definition transaction")
	}

	// check if the MintCondition is valid
	err = mdtx.MintCondition.IsStandardCondition(ctx.ValidationContext)
	if err != nil {
		return fmt.Errorf("defined mint condition is not standard within the given blockchain context: %v", err)
	}
	// check if the valid mint condition has a type we want to support, one of:
	//   * PubKey-UnlockHashCondtion
	//   * MultiSigConditions
	//   * TimeLockConditions (if the internal condition type is supported)
	err = validateMintCondition(mdtx.MintCondition)
	if err != nil {
		return err
	}

	// get MintCondition
	mintConditionBucket, err := bucket.Bucket([]byte(bucketMintConditions))
	if err != nil {
		return err
	}
	mintCondition, err := p.getMintConditionFromBucketWithContextInfo(mintConditionBucket, ctx.Confirmed, ctx.BlockHeight)
	if err != nil {
		return err
	}

	// check if MintFulfillment fulfills the Globally defined MintCondition for the context-defined block height
	err = mintCondition.Fulfill(mdtx.MintFulfillment, types.FulfillContext{
		BlockHeight: ctx.BlockHeight,
		BlockTime:   ctx.BlockTime,
		Transaction: tx.Transaction,
	})
	if err != nil {
		return fmt.Errorf("failed to fulfill mint condition for minter definition transaction: %v", err)
	}

	return nil // valid what this validator concerns
}

func (p *Plugin) validateCoinCreationTx(tx modules.ConsensusTransaction, ctx types.TransactionValidationContext, bucket *persist.LazyBoltBucket) error {
	cctx, err := CoinCreationTransactionFromTransaction(tx.Transaction, p.coinCreationTransactionVersion, p.requireMinerFees)
	if err != nil {
		return fmt.Errorf("failed to use tx as a coin creation tx: %v", err)
	}

	// ensure the Nonce is not Nil
	if cctx.Nonce == (types.TransactionNonce{}) {
		return errors.New("nil nonce is not allowed for a coin creation transaction")
	}

	// get MintCondition
	mintConditionBucket, err := bucket.Bucket([]byte(bucketMintConditions))
	if err != nil {
		return err
	}
	mintCondition, err := p.getMintConditionFromBucketWithContextInfo(mintConditionBucket, ctx.Confirmed, ctx.BlockHeight)
	if err != nil {
		return err
	}

	// check if MintFulfillment fulfills the Globally defined MintCondition for the context-defined block height
	err = mintCondition.Fulfill(cctx.MintFulfillment, types.FulfillContext{
		BlockHeight: ctx.BlockHeight,
		BlockTime:   ctx.BlockTime,
		Transaction: tx.Transaction,
	})
	if err != nil {
		return fmt.Errorf("failed to fulfill mint condition for coin creation transaction: %v", err)
	}

	return nil // valid what this validator concerns
}

func validateMintCondition(condition types.UnlockCondition) error {
	switch ct := condition.ConditionType(); ct {
	case types.ConditionTypeMultiSignature:
		// always valid
		return nil

	case types.ConditionTypeUnlockHash:
		// only valid for unlock hash type 1 (PubKey)
		if condition.UnlockHash().Type == types.UnlockTypePubKey {
			return nil
		}
		return errors.New("unlockHash conditions can be used as mint conditions, if the unlock hash type is PubKey")

	case types.ConditionTypeTimeLock:
		// ensure to unpack a proxy condition first
		if cp, ok := condition.(types.UnlockConditionProxy); ok {
			condition = cp.Condition
		}
		// time lock conditions are allowed as long as the internal condition is allowed
		cg, ok := condition.(types.MarshalableUnlockConditionGetter)
		if !ok {
			err := fmt.Errorf("unexpected Go-type for TimeLockCondition: %T", condition)
			if build.DEBUG {
				panic(err)
			}
			return err
		}
		return validateMintCondition(cg.GetMarshalableUnlockCondition())

	default:
		// all other types aren't allowed
		return fmt.Errorf("condition type %d cannot be used as a mint condition", ct)
	}
}

func (p *Plugin) validateCoinDestructionTxCreation(tx modules.ConsensusTransaction, ctx types.TransactionValidationContext, bucket *persist.LazyBoltBucket) error {
	// collect the coin input sum
	var coinInputSum types.Currency
	for _, ci := range tx.CoinInputs {
		co, ok := tx.SpentCoinOutputs[ci.ParentID]
		if !ok {
			return fmt.Errorf(
				"unable to find parent ID %s as an unspent coin output in the current consensus transaction at block height %d",
				ci.ParentID.String(), ctx.BlockHeight)
		}
		coinInputSum = coinInputSum.Add(co.Value)
	}

	// compute the lower bound
	lowerBound := tx.CoinOutputSum()

	// ensure input sum is above lowerBound
	rcmp := lowerBound.Cmp(coinInputSum)
	if rcmp == 0 {
		return fmt.Errorf(
			"all coin outputs (minus miner fees) (%s) are refunded for tx %s, this is not allowed for a coin destruction transaction",
			lowerBound.String(), tx.ID().String(),
		)
	}
	if rcmp > 0 {
		return fmt.Errorf(
			"more coin outputs (including miner fees) (%s) are refunded for tx %s than there are coin inputs (%s), this is not allowed",
			lowerBound.String(), tx.ID().String(), coinInputSum.String(),
		)
	}
	return nil
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
