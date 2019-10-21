package explorerdb

import (
	"fmt"

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
	nodeNameObjects         = "Objects"
	nodeNameUnlockhashes    = "UnlockHashes"
	nodeObjectKeyID         = "ID"
	nodeObjectKeyUnlockHash = "UnlockHash"
)

type rivbinMarshalUnmarshaler struct{}

func (rb rivbinMarshalUnmarshaler) Marshal(v interface{}) ([]byte, error) {
	return rivbin.Marshal(v)
}
func (rb rivbinMarshalUnmarshaler) Unmarshal(b []byte, v interface{}) error {
	return rivbin.Unmarshal(b, v)
}
func (rb rivbinMarshalUnmarshaler) Name() string {
	return "rivbin"
}

func (sdb *StormDB) SetChainContext(chainCtx ChainContext) error {
	b, err := rivbin.Marshal(chainCtx)
	if err != nil {
		return err
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
		return ChainContext{}, err
	}
	return chainCtx, nil
}

func (sdb *StormDB) ApplyBlock(block Block, txs []Transaction, outputs []Output, inputs map[types.OutputID]OutputSpenditureData) error {
	node := sdb.db.From(nodeNameObjects)
	// store block
	err := node.Save(block)
	if err != nil {
		return err
	}
	// store transactions
	for _, tx := range txs {
		err = node.Save(tx)
		if err != nil {
			return err
		}
	}
	// store outputs
	for _, output := range outputs {
		err = node.Save(output)
		if err != nil {
			return err
		}
	}
	// store inputs
	for outputID, spenditureData := range inputs {
		var output Output
		err = node.One(nodeObjectKeyID, outputID, &output)
		if err != nil {
			return err
		}
		if output.SpenditureData != nil {
			return fmt.Errorf("inconsent data stored for output %s: spenditure data should still be nil at the moment", outputID.String())
		}
		err = node.UpdateField(output, "SpenditureData", &spenditureData)
		if err != nil {
			return err
		}
	}
	// all good
	return nil
}

func (sdb *StormDB) RevertBlock(block types.BlockID, txs []types.TransactionID, outputs []types.OutputID, inputs []types.OutputID) error {
	node := sdb.db.From(nodeNameObjects)
	// delete block
	err := deleteFromNodeByID(node, block, new(Block))
	if err != nil {
		return err
	}
	// delete transactions
	for _, tx := range txs {
		err = deleteFromNodeByID(node, tx, new(Transaction))
		if err != nil {
			return err
		}
	}
	// delete outputs
	for _, output := range outputs {
		err = deleteFromNodeByID(node, output, new(Output))
		if err != nil {
			return err
		}
	}
	// delete inputs
	for _, inputID := range inputs {
		var output Output
		err = node.One(nodeObjectKeyID, inputID, &output)
		if err != nil {
			return err
		}
		if output.SpenditureData == nil {
			return fmt.Errorf("inconsent data stored for output %s: spenditure data should not be nil at the moment", inputID.String())
		}
		err = node.UpdateField(output, "SpenditureData", nil)
		if err != nil {
			return err
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

func (sdb *StormDB) Close() error {
	return sdb.db.Close()
}
