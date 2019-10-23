package explorerdb

import (
	"encoding/hex"
	"fmt"
	"reflect"

	"github.com/threefoldtech/rivine/modules"
	"github.com/threefoldtech/rivine/pkg/encoding/rivbin"
	"github.com/threefoldtech/rivine/types"

	"github.com/asdine/storm"
	bolt "go.etcd.io/bbolt"
)

// TODO: store wallet updates

type StormDB struct {
	db *storm.DB
}

func NewStormDB(path string) (*StormDB, error) {
	db, err := storm.Open(path, storm.Codec(rivbinMarshalUnmarshaler{}))
	if err != nil {
		return nil, err
	}
	return &StormDB{
		db: db,
	}, nil
}

var (
	bucketNameMetadata      = []byte("Metadata")
	MetadataKeyChainContext = []byte("ChainContext")
)

const (
	nodeNameObjects      = "Objects"
	nodeNameUnlockhashes = "UnlockHashes"
	nodeNamePublicKeys   = "PublicKeys"
	nodeNameBlocks       = "Blocks"

	nodeObjectKeyID         = "ID"
	nodeObjectKeyUnlockHash = "UnlockHash"

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
		return bucket.Put(MetadataKeyChainContext, b)
	})
}

func (sdb *StormDB) GetChainContext() (ChainContext, error) {
	var chainCtxBytes []byte
	err := sdb.db.Bolt.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(bucketNameMetadata)
		if bucket == nil {
			return nil
		}
		chainCtxBytes = bucket.Get(MetadataKeyChainContext)
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
			"failed to unmarshal chain context %s: %v",
			hex.EncodeToString(chainCtxBytes), err)
	}
	return chainCtx, nil
}

func (sdb *StormDB) ApplyBlock(block Block, txs []Transaction, outputs []Output, inputs map[types.OutputID]OutputSpenditureData, publicKeys map[types.UnlockHash]types.PublicKey) error {
	node := sdb.db.From(nodeNameObjects)
	blockNode := sdb.db.From(nodeNameBlocks)
	publicKeysNode := sdb.db.From(nodeNamePublicKeys)

	// store block
	err := node.Save(&block)
	if err != nil {
		return fmt.Errorf("failed to apply block: failed to save block %s: %v", block.ID.String(), err)
	}
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
		err = node.Save(&tx)
		if err != nil {
			return fmt.Errorf(
				"failed to apply block: failed to save txn %s (#%d) of block %s: %v",
				tx.ID.String(), idx+1, block.ID.String(), err)
		}
	}
	// store outputs
	for idx, output := range outputs {
		err = node.Save(&output)
		if err != nil {
			return fmt.Errorf(
				"failed to apply block: failed to save output %s (#%d) of parent %s: %v",
				output.ID.String(), idx+1, output.ParentID.String(), err)
		}
	}
	// store inputs
	for outputID, spenditureData := range inputs {
		var output Output
		err = node.One(nodeObjectKeyID, outputID, &output)
		if err != nil {
			return fmt.Errorf(
				"failed to apply block %s: failed to update output (as spent): failed to fetch existing output %s: %v",
				block.ID.String(), outputID.String(), err)
		}
		if output.SpenditureData != nil {
			return fmt.Errorf("failed to apply block %s: failed to update output (as spent): inconsent data stored for output %s: spenditure data should still be nil at the moment",
				block.ID.String(), outputID.String())
		}
		err = node.UpdateField(&output, "SpenditureData", &spenditureData)
		if err != nil {
			return fmt.Errorf(
				"failed to apply block %s: failed to update output (as spent) from parent %s: failed to update field 'SpenditureData' of existing output %s: %v",
				block.ID.String(), outputID.String(), output.ParentID.String(), err)
		}
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
	// all good
	return nil
}

func (sdb *StormDB) RevertBlock(blockContext BlockRevertContext, txs []types.TransactionID, outputs []types.OutputID, inputs []types.OutputID) error {
	node := sdb.db.From(nodeNameObjects)
	blockNode := sdb.db.From(nodeNameBlocks)

	// delete block
	err := deleteFromNodeByID(node, blockContext.ID, new(Block))
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
		err = deleteFromNodeByID(node, tx, new(Transaction))
		if err != nil {
			return fmt.Errorf(
				"failed to revert block: failed to delete txn %s (#%d) of block %s by ID: %v",
				tx.String(), idx+1, blockContext.ID.String(), err)
		}
	}
	// delete outputs
	for _, output := range outputs {
		err = deleteFromNodeByID(node, output, new(Output))
		if err != nil {
			return fmt.Errorf(
				"failed to revert block: failed to delete unspent output %s of block %s by ID: %v",
				output.String(), blockContext.ID.String(), err)
		}
	}
	// delete inputs
	for _, inputID := range inputs {
		var output Output
		err = node.One(nodeObjectKeyID, inputID, &output)
		if err != nil {
			return fmt.Errorf(
				"failed to revert block: failed to unmark spent output %s of block %s: failed to fetch by ID: %v",
				inputID.String(), blockContext.ID.String(), err)
		}
		if output.SpenditureData == nil {
			return fmt.Errorf("failed to revert block: failed to unmark spent output %s of block %s: inconsent data stored for output: spenditure data should not be nil at the moment",
				inputID.String(), blockContext.ID.String())
		}
		err = node.UpdateField(&output, "SpenditureData", nil)
		if err != nil {
			return fmt.Errorf(
				"failed to revert block: failed to unmark spent output %s from parent %s of block %s: failed to write field 'SpenditureData' as nil: %v",
				inputID.String(), output.ParentID.String(), blockContext.ID.String(), err)
		}
	}
	// all good
	return nil
}

func deleteFromNodeByID(node storm.Node, ID interface{}, value interface{}) error {
	err := node.One(nodeObjectKeyID, ID, value)
	if err != nil {
		return err
	}
	return node.DeleteStruct(value)
}

func (sdb *StormDB) GetBlock(id types.BlockID) (block Block, err error) {
	node := sdb.db.From(nodeNameObjects)
	err = node.One(nodeObjectKeyID, id, &block)
	return
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

func (sdb *StormDB) GetTransaction(id types.TransactionID) (txn Transaction, err error) {
	node := sdb.db.From(nodeNameObjects)
	err = node.One(nodeObjectKeyID, id, &txn)
	return
}

func (sdb *StormDB) GetOutput(id types.OutputID) (output Output, err error) {
	node := sdb.db.From(nodeNameObjects)
	err = node.One(nodeObjectKeyID, id, &output)
	return
}

func (sdb *StormDB) GetWallet(uh types.UnlockHash) (wallet WalletData, err error) {
	node := sdb.db.From(nodeNameUnlockhashes)
	err = node.One(nodeObjectKeyUnlockHash, uh, &wallet)
	return
}

func (sdb *StormDB) GetMultiSignatureWallet(uh types.UnlockHash) (wallet MultiSignatureWalletData, err error) {
	node := sdb.db.From(nodeNameUnlockhashes)
	err = node.One(nodeObjectKeyUnlockHash, uh, &wallet)
	return
}

func (sdb *StormDB) GetAtomicSwapContract(uh types.UnlockHash) (contract AtomicSwapContract, err error) {
	node := sdb.db.From(nodeNameUnlockhashes)
	err = node.One(nodeObjectKeyUnlockHash, uh, &contract)
	return
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
