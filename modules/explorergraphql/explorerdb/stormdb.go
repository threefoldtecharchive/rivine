package explorerdb

import (
	"fmt"
	"reflect"

	"github.com/threefoldtech/rivine/modules"
	"github.com/threefoldtech/rivine/pkg/encoding/rivbin"
	"github.com/threefoldtech/rivine/types"

	"github.com/asdine/storm"
	bolt "go.etcd.io/bbolt"
)

// TODO: store atomic swap contract updates

// TODO: store wallet updates

type StormDB struct {
	db       *storm.DB
	chainCts types.ChainConstants
}

func NewStormDB(path string, chainCts types.ChainConstants) (*StormDB, error) {
	db, err := storm.Open(path, storm.Codec(rivbinMarshalUnmarshaler{}))
	if err != nil {
		return nil, err
	}
	return &StormDB{
		db:       db,
		chainCts: chainCts,
	}, nil
}

var (
	bucketNameMetadata              = []byte("Metadata")
	metadataKeyChainContext         = []byte("ChainContext")
	metadataKeyInternalData         = []byte("InternalData")
	metadataKeyChainAggregatedFacts = []byte("ChainAggregatedFacts")
)

const (
	nodeNamePublicKeys = "PublicKeys"
	nodeNameBlocks     = "Blocks"

	nodePublicKeysKeyUnlockHash = "UnlockHash"

	nodeBlocksKeyReference = "Reference"
)

type rivbinMarshalUnmarshaler struct{}

func (rb rivbinMarshalUnmarshaler) Marshal(v interface{}) ([]byte, error) {
	// do not marshal ptrs, this function is only called on root objects, so should be fine
	// ... last famous words o.o
	val := reflect.ValueOf(v)
	if val.IsValid() && val.Kind() == reflect.Ptr && !val.IsNil() {
		val = val.Elem()
	}
	v = val.Interface()
	return rivbin.Marshal(v)
}
func (rb rivbinMarshalUnmarshaler) Unmarshal(b []byte, v interface{}) error {
	return rivbin.Unmarshal(b, v)
}
func (rb rivbinMarshalUnmarshaler) Name() string {
	return "rivbin"
}

// used for block rp -> bID node
type blockReferencePointIDPair struct {
	Reference ReferencePoint `storm:"id"`
	BlockID   types.BlockID
}

// used for unlockHash -> PublicKey node
type unlockHashPublicKeyPair struct {
	UnlockHash types.UnlockHash `storm:"id"`
	PublicKey  types.PublicKey
}

type stormDBInternalData struct {
	LastDataID stormDataID
}

// TODO: where we use manual bolt calls, re-use existing bolt.Tx, once we do start using those somehow

func (sdb *StormDB) saveInternalData(internalData stormDBInternalData) error {
	b, err := rivbin.Marshal(internalData)
	if err != nil {
		return fmt.Errorf(
			"failed to marshal internal stormDB data %v: %v", internalData, err)
	}
	return sdb.db.Bolt.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists(bucketNameMetadata)
		if err != nil {
			return fmt.Errorf("bucket %s is not created while it is expected to be: %v", string(bucketNameMetadata), err)
		}
		return bucket.Put(metadataKeyInternalData, b)
	})
}

func (sdb *StormDB) getInternalData() (stormDBInternalData, error) {
	var internalDataBytes []byte
	err := sdb.db.Bolt.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(bucketNameMetadata)
		if bucket == nil {
			return nil
		}
		internalDataBytes = bucket.Get(metadataKeyInternalData)
		return nil
	})
	if err != nil {
		return stormDBInternalData{}, err
	}
	if len(internalDataBytes) == 0 { // start from zero
		return stormDBInternalData{
			LastDataID: 1, // start at index 1, as stormDB doesn't allow a 0-index
		}, nil
	}
	var internalData stormDBInternalData
	err = rivbin.Unmarshal(internalDataBytes, &internalData)
	if err != nil {
		return stormDBInternalData{}, fmt.Errorf(
			"failed to unmarshal internal stormDB data %x: %v", internalDataBytes, err)
	}
	return internalData, nil
}

func (sdb *StormDB) SetChainContext(chainCtx ChainContext) error {
	b, err := rivbin.Marshal(chainCtx)
	if err != nil {
		return fmt.Errorf(
			"failed to marshal chain context %v: %v", chainCtx, err)
	}
	return sdb.db.Bolt.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists(bucketNameMetadata)
		if err != nil {
			return fmt.Errorf("bucket %s is not created while it is expected to be: %v", string(bucketNameMetadata), err)
		}
		return bucket.Put(metadataKeyChainContext, b)
	})
}

func (sdb *StormDB) GetChainContext() (ChainContext, error) {
	var chainCtxBytes []byte
	err := sdb.db.Bolt.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(bucketNameMetadata)
		if bucket == nil {
			return nil
		}
		chainCtxBytes = bucket.Get(metadataKeyChainContext)
		return nil
	})
	if err != nil {
		return ChainContext{}, err
	}
	if len(chainCtxBytes) == 0 { // start from zero
		return ChainContext{
			ConsensusChangeID: modules.ConsensusChangeBeginning,
		}, nil
	}
	var chainCtx ChainContext
	err = rivbin.Unmarshal(chainCtxBytes, &chainCtx)
	if err != nil {
		return ChainContext{}, fmt.Errorf(
			"failed to unmarshal chain context %x: %v", chainCtxBytes, err)
	}
	return chainCtx, nil
}

func (sdb *StormDB) GetChainAggregatedFacts() (ChainAggregatedFacts, error) {
	var factsBytes []byte
	err := sdb.db.Bolt.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(bucketNameMetadata)
		if bucket == nil {
			return nil
		}
		factsBytes = bucket.Get(metadataKeyChainAggregatedFacts)
		return nil
	})
	if err != nil {
		return ChainAggregatedFacts{}, err
	}
	if len(factsBytes) == 0 { // start from zero
		return ChainAggregatedFacts{}, nil
	}
	var facts ChainAggregatedFacts
	err = rivbin.Unmarshal(factsBytes, &facts)
	if err != nil {
		return ChainAggregatedFacts{}, fmt.Errorf(
			"failed to unmarshal chain aggregated facts %x: %v", factsBytes, err)
	}
	return facts, nil
}

func (sdb *StormDB) updateChainAggregatedFacts(cb func(facts *ChainAggregatedFacts) error) error {
	return sdb.db.Bolt.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists(bucketNameMetadata)
		if err != nil {
			return err
		}

		var (
			facts      ChainAggregatedFacts
			factsBytes []byte
		)
		if factsBytes = bucket.Get(metadataKeyChainAggregatedFacts); len(factsBytes) != 0 {
			err = rivbin.Unmarshal(factsBytes, &facts)
			if err != nil {
				return fmt.Errorf(
					"failed to unmarshal chain aggregated facts %x: %v", factsBytes, err)
			}
		}

		err = cb(&facts)
		if err != nil {
			return err
		}

		factsBytes, err = rivbin.Marshal(facts)
		if err != nil {
			return err
		}

		return bucket.Put(metadataKeyChainAggregatedFacts, factsBytes)
	})
}

func (sdb *StormDB) ApplyBlock(block Block, blockFacts BlockFactsConstants, txs []Transaction, outputs []Output, inputs map[types.OutputID]OutputSpenditureData, publicKeys map[types.UnlockHash]types.PublicKey) error {
	sdbInternalData, err := sdb.getInternalData()
	if err != nil {
		return err
	}

	node := newStormObjectNodeReaderWriter(sdb.db, sdbInternalData.LastDataID)
	blockNode := sdb.db.From(nodeNameBlocks)
	publicKeysNode := sdb.db.From(nodeNamePublicKeys)

	// store block reference point parings
	err = blockNode.Save(&blockReferencePointIDPair{
		Reference: ReferencePoint(block.Height) + 1, // require +1, as a storm identifier cannot be 0
		BlockID:   block.ID,
	})
	if err != nil {
		return fmt.Errorf("failed to apply block: failed to save block %s's height %d as reference point: %v", block.ID.String(), block.Height, err)
	}
	err = blockNode.Save(&blockReferencePointIDPair{
		Reference: ReferencePoint(block.Timestamp),
		BlockID:   block.ID,
	})
	if err != nil {
		return fmt.Errorf("failed to apply block: failed to save block %s's timestamp %d as reference point: %v", block.ID.String(), block.Timestamp, err)
	}
	// store transactions
	for idx, tx := range txs {
		err = node.SaveTransaction(&tx)
		if err != nil {
			return fmt.Errorf(
				"failed to apply block: failed to save txn %s (#%d) of block %s: %v",
				tx.ID.String(), idx+1, block.ID.String(), err)
		}
	}
	// store outputs
	for idx, output := range outputs {
		err = node.SaveOutput(&output)
		if err != nil {
			return fmt.Errorf(
				"failed to apply block: failed to save output %s (#%d) of parent %s: %v",
				output.ID.String(), idx+1, output.ParentID.String(), err)
		}
	}
	// store inputs
	inputOutputSlice := make([]*Output, 0, len(inputs))
	for outputID, spenditureData := range inputs {
		output, err := node.UpdateOutputSpenditureData(outputID, &spenditureData)
		if err != nil {
			return fmt.Errorf(
				"failed to apply block %s: failed to update output (as spent): failed to update existing output %s: %v",
				block.ID.String(), outputID.String(), err)
		}
		inputOutputSlice = append(inputOutputSlice, &output)
	}
	// store public keys
	for uh, pk := range publicKeys {
		err = publicKeysNode.Save(&unlockHashPublicKeyPair{
			UnlockHash: uh,
			PublicKey:  pk,
		})
		if err != nil {
			return fmt.Errorf("failed to apply block: failed to save block %s's unlock hash %s mapped to public key %s: %v", block.ID.String(), uh.String(), pk.String(), err)
		}
	}
	// update aggregated facts
	facts, err := sdb.applyBlockToAggregatedFacts(block.Height, block.Timestamp, node, outputs, inputOutputSlice, blockFacts.Target)
	if err != nil {
		return fmt.Errorf("failed to apply block: failed to save aggregated chain facts at block %s (height: %d): %v", block.ID.String(), block.Height, err)
	}

	// store block with facts at the end, now that we have the final chain facts, after applying this block
	err = node.SaveBlockWithFacts(&block, &BlockFacts{
		Constants: blockFacts,
		Aggregated: BlockFactsAggregated{
			TotalCoins:                 facts.TotalCoins,
			TotalLockedCoins:           facts.TotalLockedCoins,
			TotalBlockStakes:           facts.TotalBlockStakes,
			TotalLockedBlockStakes:     facts.TotalLockedBlockStakes,
			EstimatedActiveBlockStakes: facts.EstimatedActiveBlockStakes,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to apply block: failed to save block %s with facts: %v", block.ID.String(), err)
	}

	// finally store the internal stormDB data
	sdbInternalData.LastDataID = node.GetLastDataID()
	return sdb.saveInternalData(sdbInternalData)
}

func (sdb *StormDB) applyBlockToAggregatedFacts(height types.BlockHeight, timestamp types.Timestamp, objectNode stormObjectNodeReaderWriter, outputs []Output, inputs []*Output, target types.Target) (ChainAggregatedFacts, error) {
	var (
		err             error
		isLocked        bool
		outputsUnlocked []stormOutput
	)
	// get outputs unlocked by height and timestamp
	if height != 0 { // do not do it for block 0, as it will return all outputs that do not have a lock
		outputsUnlocked, err = objectNode.GetStormOutputsbyUnlockReferencePoint(height, timestamp)
		if err != nil {
			if err == storm.ErrNotFound {
				// ignore not found
				outputsUnlocked = nil
			} else {
				return ChainAggregatedFacts{}, fmt.Errorf("failed to get outputs unlocked by height %d or timestamp %d: %v", height, timestamp, err)
			}
		}
	}
	// get outputs unlocked by timestamp
	var resultFacts ChainAggregatedFacts
	err = sdb.updateChainAggregatedFacts(func(facts *ChainAggregatedFacts) error {
		// discount all new inputs from total coins/blockstakes
		for _, input := range inputs {
			if input.Type == OutputTypeBlockStake {
				facts.TotalBlockStakes = facts.TotalBlockStakes.Sub(input.Value)
			} else {
				facts.TotalCoins = facts.TotalCoins.Sub(input.Value)
			}
		}
		// count all new outputs, and if locked also add them to the total locked assets
		for _, output := range outputs {
			isLocked = output.UnlockReferencePoint > 0 && !output.UnlockReferencePoint.Overreached(height, timestamp)
			if output.Type == OutputTypeBlockStake {
				facts.TotalBlockStakes = facts.TotalBlockStakes.Add(output.Value)
				if isLocked {
					facts.TotalLockedBlockStakes = facts.TotalLockedBlockStakes.Add(output.Value)
				}
			} else {
				facts.TotalCoins = facts.TotalCoins.Add(output.Value)
				if isLocked {
					facts.TotalLockedCoins = facts.TotalLockedCoins.Add(output.Value)
				}
			}
		}
		// discount all unlocked outputs from total locked coins/blockstakes
		for _, output := range outputsUnlocked {
			if output.Type == OutputTypeBlockStake {
				facts.TotalLockedBlockStakes = facts.TotalLockedBlockStakes.Sub(output.Value)
			} else {
				facts.TotalLockedCoins = facts.TotalLockedCoins.Sub(output.Value)
			}
		}
		// update estimated active block stakes
		facts.AddLastBlockContext(BlockFactsContext{
			Target:    target,
			Timestamp: timestamp,
		})
		facts.EstimatedActiveBlockStakes = sdb.estimateActiveBS(height, timestamp, facts.LastBlocks)

		// keep a shallow copy of the facts
		resultFacts = *facts
		return nil
	})
	return resultFacts, err
}

func (sdb *StormDB) estimateActiveBS(height types.BlockHeight, timestamp types.Timestamp, blocks []BlockFactsContext) types.Currency {
	if len(blocks) == 0 {
		return types.Currency{}
	}
	var (
		estimatedActiveBS types.Difficulty
		block             BlockFactsContext

		lBlockOffset    = len(blocks) - 1
		oldestTimestamp = blocks[lBlockOffset].Timestamp
		totalDifficulty = blocks[lBlockOffset].Target
	)
	for i := range blocks[:lBlockOffset] {
		block = blocks[lBlockOffset-i]
		totalDifficulty = totalDifficulty.AddDifficulties(block.Target, sdb.chainCts.RootDepth)
		oldestTimestamp = block.Timestamp
	}
	secondsPassed := timestamp - oldestTimestamp
	estimatedActiveBS = totalDifficulty.Difficulty(sdb.chainCts.RootDepth)
	if secondsPassed > 0 {
		estimatedActiveBS = estimatedActiveBS.Div64(uint64(secondsPassed))
	}
	return types.NewCurrency(estimatedActiveBS.Big())
}

func (sdb *StormDB) RevertBlock(blockContext BlockRevertContext, txs []types.TransactionID, outputs []types.OutputID, inputs []types.OutputID) error {
	sdbInternalData, err := sdb.getInternalData()
	if err != nil {
		return err
	}

	node := newStormObjectNodeReaderWriter(sdb.db, sdbInternalData.LastDataID)
	blockNode := sdb.db.From(nodeNameBlocks)

	// delete block
	err = node.DeleteBlock(blockContext.ID)
	if err != nil {
		return fmt.Errorf("failed to revert block: failed to delete block %s by ID: %v", blockContext.ID.String(), err)
	}
	// delete block reference point parings
	err = blockNode.DeleteStruct(&blockReferencePointIDPair{
		Reference: ReferencePoint(blockContext.Height) + 1, // require +1, as a storm identifier cannot be 0
		BlockID:   blockContext.ID,
	})
	if err != nil {
		return fmt.Errorf("failed to revert block: failed to delete block %s's height %d by reference point: %v", blockContext.ID.String(), blockContext.Height, err)
	}
	err = blockNode.Save(&blockReferencePointIDPair{
		Reference: ReferencePoint(blockContext.Timestamp),
		BlockID:   blockContext.ID,
	})
	if err != nil {
		return fmt.Errorf("failed to revert block: failed to delete block %s's timestamp %d by reference point: %v", blockContext.ID.String(), blockContext.Timestamp, err)
	}
	// delete transactions
	for idx, tx := range txs {
		err = node.DeleteTransaction(tx)
		if err != nil {
			return fmt.Errorf(
				"failed to revert block: failed to delete txn %s (#%d) of block %s by ID: %v",
				tx.String(), idx+1, blockContext.ID.String(), err)
		}
	}
	// delete outputs
	outputSlice := make([]*Output, 0, len(outputs))
	for _, output := range outputs {
		var obj Object
		err = node.DeleteOutput(output)
		if err != nil {
			return fmt.Errorf(
				"failed to revert block: failed to delete unspent output %s of block %s by ID: %v",
				output.String(), blockContext.ID.String(), err)
		}
		if obj.Type != ObjectTypeOutput {
			return fmt.Errorf(
				"failed to revert block: unspent output %s of block %s was stored as unexpected object type %d (expected %d)",
				output.String(), blockContext.ID.String(), obj.Type, ObjectTypeOutput)
		}
		outputObj, ok := obj.Data.(Output)
		if !ok {
			return fmt.Errorf(
				"failed to revert block: unspent output %s of block %s was stored as unexpected object data type %T (expected %T)",
				output.String(), blockContext.ID.String(), obj.Data, Output{})
		}
		outputSlice = append(outputSlice, &outputObj)
	}
	// delete inputs
	inputOutputSlice := make([]*Output, 0, len(inputs))
	for _, inputID := range inputs {
		output, err := node.UpdateOutputSpenditureData(inputID, nil)
		if err != nil {
			return fmt.Errorf(
				"failed to revert block: failed to unmark spent output %s of block %s: failed to update existing output: %v",
				inputID.String(), blockContext.ID.String(), err)
		}
		inputOutputSlice = append(inputOutputSlice, &output)
	}
	// update aggregated facts
	err = sdb.revertBlockToAggregatedFacts(blockContext.Height, blockContext.Timestamp, node, outputSlice, inputOutputSlice)
	if err != nil {
		return fmt.Errorf("failed to apply block: failed to save aggregated chain facts at block %s (height: %d): %v", blockContext.ID.String(), blockContext.Height, err)
	}

	// finally store the internal stormDB data
	sdbInternalData.LastDataID = node.GetLastDataID()
	return sdb.saveInternalData(sdbInternalData)
}

func (sdb *StormDB) revertBlockToAggregatedFacts(height types.BlockHeight, timestamp types.Timestamp, objectNode stormObjectNodeReaderWriter, outputs []*Output, inputs []*Output) error {
	var (
		err           error
		isLocked      bool
		outputsLocked []stormOutput
	)
	// get outputs unlocked by height and timestamp
	if height != 0 { // do not do it for block 0, as it will return all outputs that do not have a lock
		outputsLocked, err = objectNode.GetStormOutputsbyUnlockReferencePoint(height, timestamp)
		if err != nil {
			if err == storm.ErrNotFound {
				// ignore not found
				outputsLocked = nil
			} else {
				return fmt.Errorf("failed to get outputs locked until height %d or timestamp %d: %v", height, timestamp, err)
			}
		}
	}
	// get outputs unlocked by timestamp
	return sdb.updateChainAggregatedFacts(func(facts *ChainAggregatedFacts) error {
		// re-apply all reverted inputs to total coins/blockstakes
		for _, input := range inputs {
			if input.Type == OutputTypeBlockStake {
				facts.TotalBlockStakes = facts.TotalBlockStakes.Add(input.Value)
			} else {
				facts.TotalCoins = facts.TotalCoins.Add(input.Value)
			}
		}
		// re-apply all locked outputs to total locked coins/blockstakes
		for _, output := range outputsLocked {
			if output.Type == OutputTypeBlockStake {
				facts.TotalLockedBlockStakes = facts.TotalLockedBlockStakes.Add(output.Value)
			} else {
				facts.TotalLockedCoins = facts.TotalLockedCoins.Add(output.Value)
			}
		}
		// discount all reverted outputs, and if locked also subtract them to the total locked assets
		for _, output := range outputs {
			isLocked = output.UnlockReferencePoint > 0 && !output.UnlockReferencePoint.Overreached(height, timestamp)
			if output.Type == OutputTypeBlockStake {
				facts.TotalBlockStakes = facts.TotalBlockStakes.Sub(output.Value)
				if isLocked {
					facts.TotalLockedBlockStakes = facts.TotalLockedBlockStakes.Sub(output.Value)
				}
			} else {
				facts.TotalCoins = facts.TotalCoins.Sub(output.Value)
				if isLocked {
					facts.TotalLockedCoins = facts.TotalLockedCoins.Sub(output.Value)
				}
			}
		}
		facts.RemoveLastBlockContext()
		facts.EstimatedActiveBlockStakes = sdb.estimateActiveBS(height, timestamp, facts.LastBlocks)
		// all good
		return nil
	})
}

func (sdb *StormDB) GetObject(id ObjectID) (Object, error) {
	node := newStormObjectNodeReader(sdb.db)
	return node.GetObject(id)
}

func (sdb *StormDB) GetBlock(id types.BlockID) (Block, error) {
	node := newStormObjectNodeReader(sdb.db)
	return node.GetBlock(id)
}

func (sdb *StormDB) GetBlockFacts(id types.BlockID) (BlockFacts, error) {
	node := newStormObjectNodeReader(sdb.db)
	return node.GetBlockFacts(id)
}

func (sdb *StormDB) GetBlockByReferencePoint(rp ReferencePoint) (Block, error) {
	node := sdb.db.From(nodeNameBlocks)
	var pair blockReferencePointIDPair
	if rp.IsBlockHeight() {
		rp++
	}
	err := node.One(nodeBlocksKeyReference, rp, &pair)
	if err != nil {
		return Block{}, err
	}
	return sdb.GetBlock(pair.BlockID)
}

func (sdb *StormDB) GetBlockFactsByReferencePoint(rp ReferencePoint) (BlockFacts, error) {
	node := sdb.db.From(nodeNameBlocks)
	var pair blockReferencePointIDPair
	if rp.IsBlockHeight() {
		rp++
	}
	err := node.One(nodeBlocksKeyReference, rp, &pair)
	if err != nil {
		return BlockFacts{}, err
	}
	return sdb.GetBlockFacts(pair.BlockID)
}

func (sdb *StormDB) GetTransaction(id types.TransactionID) (Transaction, error) {
	node := newStormObjectNodeReader(sdb.db)
	return node.GetTransaction(id)
}

func (sdb *StormDB) GetOutput(id types.OutputID) (Output, error) {
	node := newStormObjectNodeReader(sdb.db)
	return node.GetOutput(id)
}

func (sdb *StormDB) GetSingleSignatureWallet(uh types.UnlockHash) (SingleSignatureWalletData, error) {
	node := newStormObjectNodeReader(sdb.db)
	return node.GetSingleSignatureWallet(uh)
}

func (sdb *StormDB) GetMultiSignatureWallet(uh types.UnlockHash) (MultiSignatureWalletData, error) {
	node := newStormObjectNodeReader(sdb.db)
	return node.GetMultiSignatureWallet(uh)
}

func (sdb *StormDB) GetAtomicSwapContract(uh types.UnlockHash) (AtomicSwapContract, error) {
	node := newStormObjectNodeReader(sdb.db)
	return node.GetAtomicSwapContract(uh)
}

func (sdb *StormDB) GetPublicKey(uh types.UnlockHash) (types.PublicKey, error) {
	node := sdb.db.From(nodeNamePublicKeys)
	var pair unlockHashPublicKeyPair
	err := node.One(nodePublicKeysKeyUnlockHash, uh, &pair)
	if err != nil {
		return types.PublicKey{}, err
	}
	return pair.PublicKey, nil
}

func (sdb *StormDB) Close() error {
	return sdb.db.Close()
}
