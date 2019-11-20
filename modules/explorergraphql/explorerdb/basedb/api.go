package basedb

import (
	"github.com/threefoldtech/rivine/types"
)

type (
	DB interface {
		// You can run each call in its own R/W  Txn by calling
		// the txn command directly on the DB.
		RWTxn

		// ReadTransaction batches multiple read calls together,
		// to keep the disk I/O to a minimum
		ReadTransaction(func(RTxn) error) error
		// ReadWriteTransaction batches multiple read-write calls together,
		// to keep the disk I/O to a minimum
		ReadWriteTransaction(func(RWTxn) error) error

		// Close the DB
		Close() error
	}

	RTxn interface {
		GetChainContext() (ChainContext, error)

		GetChainAggregatedFacts() (ChainAggregatedFacts, error)

		GetObject(ObjectID) (Object, error)
		GetObjectInfo(ObjectID) (ObjectInfo, error)

		GetBlock(types.BlockID) (Block, error)
		GetBlockFacts(types.BlockID) (BlockFacts, error)
		GetBlockAt(types.BlockHeight) (Block, error)
		GetBlockIDAt(types.BlockHeight) (types.BlockID, error)
		GetTransaction(types.TransactionID) (Transaction, error)
		GetOutput(types.OutputID) (Output, error)

		GetBlocks(limit *int, filter *BlocksFilter, cursor *Cursor) ([]Block, *Cursor, error)

		GetFreeForAllWallet(types.UnlockHash) (FreeForAllWalletData, error)
		GetSingleSignatureWallet(types.UnlockHash) (SingleSignatureWalletData, error)
		GetMultiSignatureWallet(types.UnlockHash) (MultiSignatureWalletData, error)
		GetAtomicSwapContract(types.UnlockHash) (AtomicSwapContract, error)

		GetPublicKey(types.UnlockHash) (types.PublicKey, error)
	}

	RWTxn interface {
		RTxn

		SetChainContext(ChainContext) error

		ApplyBlock(block Block, blockFacts BlockFactsConstants, txs []Transaction, outputs []Output, inputs map[types.OutputID]OutputSpenditureData) error
		// TODO: should we also revert public key (from UH) mapping? (TODO 4)
		RevertBlock(blockContext BlockRevertContext, txs []types.TransactionID, outputs []types.OutputID, inputs []types.OutputID) error

		// commit the work done from Memory to Disk,
		// only required in case you are doing a big amount of calls within a single transaction.
		// If you want to continue using this transaction, you'll have to set final to true
		Commit(final bool) error
	}
)
