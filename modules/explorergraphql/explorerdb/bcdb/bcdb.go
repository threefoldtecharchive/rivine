package bcdb

import (
	"encoding/hex"
	"fmt"
	"path/filepath"

	"github.com/threefoldtech/rivine/crypto"
	"github.com/threefoldtech/rivine/modules"
	"github.com/threefoldtech/rivine/persist"
	"github.com/threefoldtech/rivine/types"

	"github.com/threefoldtech/rivine/modules/explorergraphql/explorerdb/basedb"
	"github.com/threefoldtech/rivine/modules/explorergraphql/explorerdb/bcdb/tftexplorer"
	"github.com/threefoldtech/rivine/modules/explorergraphql/explorerdb/bcdb/tftexplorer/tftschema"
)

const (
	bcDBName = "BCDB"
)

type DB struct {
	client   *tftexplorer.Client
	logger   *persist.Logger
	bcInfo   *types.BlockchainInfo
	chainCts *types.ChainConstants
}

var (
	_ basedb.DB = (*DB)(nil)
)

func New(addr, path string, bcInfo types.BlockchainInfo, chainCts types.ChainConstants, verbose bool) (*DB, error) {
	client := tftexplorer.NewClient(addr)
	// Initialize the logger.
	logFilePath := filepath.Join(path, "explorer.log")
	logger, err := persist.NewFileLogger(bcInfo, logFilePath, verbose)
	if err != nil {
		return nil, err
	}
	return &DB{
		client:   client,
		logger:   logger,
		chainCts: &chainCts,
		bcInfo:   &bcInfo,
	}, nil
}

func (bcdb *DB) SetChainContext(chainCtx basedb.ChainContext) error {
	schemaChainCtx := tftschema.TftExplorerChainContext1{
		ConsensusChangeId: hex.EncodeToString(chainCtx.ConsensusChangeID[:]),
		Height:            int64(chainCtx.Height),
		Timestamp:         tftschema.DateFromTimestamp(int64(chainCtx.Timestamp)),
		BlockId:           chainCtx.BlockID.String(),
	}
	return bcdb.client.Set("set_chain_context", schemaChainCtx, nil)
}

func (bcdb *DB) GetChainContext() (basedb.ChainContext, error) {
	var schemaChainCtx tftschema.TftExplorerChainContext1
	err := bcdb.client.Get("get_chain_context", nil, &schemaChainCtx)
	if err != nil {
		return basedb.ChainContext{}, err
	}
	var (
		ccid modules.ConsensusChangeID
		bid  crypto.Hash
	)
	if schemaChainCtx.ConsensusChangeId != "" {
		var ccidHash crypto.Hash
		err = ccidHash.LoadString(schemaChainCtx.ConsensusChangeId)
		if err != nil {
			return basedb.ChainContext{}, fmt.Errorf("failed to parse received ccid '%s': %v", schemaChainCtx.ConsensusChangeId, err)
		}
		ccid = modules.ConsensusChangeID(ccidHash)
	} else {
		ccid = modules.ConsensusChangeBeginning
	}
	if schemaChainCtx.BlockId != "" {
		err = bid.LoadString(schemaChainCtx.BlockId)
		if err != nil {
			return basedb.ChainContext{}, fmt.Errorf("failed to parse received block ID '%s': %v", schemaChainCtx.BlockId, err)
		}
	}
	return basedb.ChainContext{
		ConsensusChangeID: ccid,
		Height:            types.BlockHeight(schemaChainCtx.Height),
		Timestamp:         types.Timestamp(schemaChainCtx.Timestamp.Unix()),
		BlockID:           types.BlockID(bid),
	}, nil
}

func (bcdb *DB) GetChainAggregatedFacts() (basedb.ChainAggregatedFacts, error) {
	return basedb.ChainAggregatedFacts{}, basedb.ErrNotFound // TODO
}

func (bcdb *DB) ApplyBlock(block basedb.Block, blockFacts basedb.BlockFactsConstants, txs []basedb.Transaction, outputs []basedb.Output, inputs map[types.OutputID]basedb.OutputSpenditureData) error {
	return nil // TODO
}

func (bcdb *DB) RevertBlock(blockContext basedb.BlockRevertContext, txs []types.TransactionID, outputs []types.OutputID, inputs []types.OutputID) error {
	return nil // TODO
}

func (bcdb *DB) GetObject(id basedb.ObjectID) (basedb.Object, error) {
	return basedb.Object{}, basedb.ErrNotFound // TODO
}

func (bcdb *DB) GetObjectInfo(id basedb.ObjectID) (basedb.ObjectInfo, error) {
	return basedb.ObjectInfo{}, basedb.ErrNotFound // TODO
}

func (bcdb *DB) GetBlock(id types.BlockID) (basedb.Block, error) {
	return basedb.Block{}, basedb.ErrNotFound // TODO
}

func (bcdb *DB) GetBlockAt(height types.BlockHeight) (basedb.Block, error) {
	return basedb.Block{}, basedb.ErrNotFound // TODO
}

func (bcdb *DB) GetBlockFacts(id types.BlockID) (basedb.BlockFacts, error) {
	return basedb.BlockFacts{}, basedb.ErrNotFound // TODO
}

func (bcdb *DB) GetBlockIDAt(height types.BlockHeight) (types.BlockID, error) {
	return types.BlockID{}, basedb.ErrNotFound // TODO
}

func (bcdb *DB) GetTransaction(id types.TransactionID) (basedb.Transaction, error) {
	return basedb.Transaction{}, basedb.ErrNotFound // TODO
}

func (bcdb *DB) GetOutput(id types.OutputID) (basedb.Output, error) {
	return basedb.Output{}, basedb.ErrNotFound // TODO
}

const (
	DefaultLimitBlocks = 10
	UpperLimitBlocks   = 100
)

func (bcdb *DB) GetBlocks(limit *int, filter *basedb.BlocksFilter, cursor *basedb.Cursor) ([]basedb.Block, *basedb.Cursor, error) {
	return nil, nil, basedb.ErrNotFound // TODO
}

func (bcdb *DB) GetFreeForAllWallet(uh types.UnlockHash) (basedb.FreeForAllWalletData, error) {
	return basedb.FreeForAllWalletData{}, basedb.ErrNotFound // TODO
}

func (bcdb *DB) GetSingleSignatureWallet(uh types.UnlockHash) (basedb.SingleSignatureWalletData, error) {
	return basedb.SingleSignatureWalletData{}, basedb.ErrNotFound // TODO
}

func (bcdb *DB) GetMultiSignatureWallet(uh types.UnlockHash) (basedb.MultiSignatureWalletData, error) {
	return basedb.MultiSignatureWalletData{}, basedb.ErrNotFound // TODO
}

func (bcdb *DB) GetAtomicSwapContract(uh types.UnlockHash) (basedb.AtomicSwapContract, error) {
	return basedb.AtomicSwapContract{}, basedb.ErrNotFound // TODO
}

func (bcdb *DB) GetPublicKey(uh types.UnlockHash) (types.PublicKey, error) {
	return types.PublicKey{}, basedb.ErrNotFound // TODO
}

func (bcdb *DB) Commit(final bool) error {
	return nil // nothing to be done here, every bcdb call is immediately commited.
}

// ReadTransaction batches multiple read calls together,
// to keep the disk I/O to a minimum
func (bcdb *DB) ReadTransaction(f func(basedb.RTxn) error) error {
	return f(bcdb) // there is no Tx we can make for BCDB AFAIK, so use the bcdb directly
}

// ReadWriteTransaction batches multiple read-write calls together,
// to keep the disk I/O to a minimum
func (bcdb *DB) ReadWriteTransaction(f func(basedb.RWTxn) error) error {
	return f(bcdb) // there is no Tx we can make for BCDB AFAIK, so use the bcdb directly
}

func (bcdb *DB) Close() error {
	return nil // nothing to do
}
