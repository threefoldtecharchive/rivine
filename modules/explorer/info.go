package explorer

import (
	"github.com/rivine/rivine/build"
	"github.com/rivine/rivine/modules"
	"github.com/rivine/rivine/types"

	"github.com/rivine/bbolt"
)

// Block takes a block ID and finds the corresponding block, provided that the
// block is in the consensus set.
func (e *Explorer) Block(id types.BlockID) (types.Block, types.BlockHeight, bool) {
	var height types.BlockHeight
	err := e.db.View(dbGetAndDecode(bucketBlockIDs, id, &height))
	if err != nil {
		return types.Block{}, 0, false
	}
	block, exists := e.cs.BlockAtHeight(height)
	if !exists {
		return types.Block{}, 0, false
	}
	return block, height, true
}

// BlockFacts returns a set of statistics about the blockchain as they appeared
// at a given block height, and a bool indicating whether facts exist for the
// given height.
func (e *Explorer) BlockFacts(height types.BlockHeight) (modules.BlockFacts, bool) {
	var bf blockFacts
	err := e.db.View(e.dbGetBlockFacts(height, &bf))
	if err != nil {
		return modules.BlockFacts{}, false
	}

	return bf.BlockFacts, true
}

// LatestBlockFacts returns a set of statistics about the blockchain as they appeared
// at the latest block height in the explorer's consensus set.
func (e *Explorer) LatestBlockFacts() modules.BlockFacts {
	var bf blockFacts
	err := e.db.View(func(tx *bolt.Tx) error {
		var height types.BlockHeight
		err := dbGetInternal(internalBlockHeight, &height)(tx)
		if err != nil {
			return err
		}
		return e.dbGetBlockFacts(height, &bf)(tx)
	})
	if err != nil {
		build.Critical(err)
	}
	return bf.BlockFacts
}

// Transaction takes a transaction ID and finds the block containing the
// transaction. Because of the miner payouts, the transaction ID might be a
// block ID. To find the transaction, iterate through the block.
func (e *Explorer) Transaction(id types.TransactionID) (types.Block, types.BlockHeight, bool) {
	var height types.BlockHeight
	err := e.db.View(dbGetAndDecode(bucketTransactionIDs, id, &height))
	if err != nil {
		return types.Block{}, 0, false
	}
	block, exists := e.cs.BlockAtHeight(height)
	if !exists {
		return types.Block{}, 0, false
	}
	return block, height, true
}

// UnlockHash returns the IDs of all the transactions that contain the unlock
// hash. An empty set indicates that the unlock hash does not appear in the
// blockchain.
func (e *Explorer) UnlockHash(uh types.UnlockHash) []types.TransactionID {
	var ids []types.TransactionID
	err := e.db.View(dbGetTransactionIDSet(bucketUnlockHashes, uh, &ids))
	if err != nil {
		ids = nil
	}
	return ids
}

// CoinOutput returns the coin output associated with the specified ID.
func (e *Explorer) CoinOutput(id types.CoinOutputID) (types.CoinOutput, bool) {
	var sco types.CoinOutput
	err := e.db.View(dbGetAndDecode(bucketCoinOutputs, id, &sco))
	if err != nil {
		return types.CoinOutput{}, false
	}
	return sco, true
}

// CoinOutputID returns all of the transactions that contain the specified
// coin output ID. An empty set indicates that the siacoin output ID does
// not appear in the blockchain.
func (e *Explorer) CoinOutputID(id types.CoinOutputID) []types.TransactionID {
	var ids []types.TransactionID
	err := e.db.View(dbGetTransactionIDSet(bucketCoinOutputIDs, id, &ids))
	if err != nil {
		ids = nil
	}
	return ids
}

// BlockStakeOutput returns the blockstake output associated with the specified ID.
func (e *Explorer) BlockStakeOutput(id types.BlockStakeOutputID) (types.BlockStakeOutput, bool) {
	var sco types.BlockStakeOutput
	err := e.db.View(dbGetAndDecode(bucketBlockStakeOutputs, id, &sco))
	if err != nil {
		return types.BlockStakeOutput{}, false
	}
	return sco, true
}

// BlockStakeOutputID returns all of the transactions that contain the specified
// blockstake output ID. An empty set indicates that the blockstake output ID does
// not appear in the blockchain.
func (e *Explorer) BlockStakeOutputID(id types.BlockStakeOutputID) []types.TransactionID {
	var ids []types.TransactionID
	err := e.db.View(dbGetTransactionIDSet(bucketBlockStakeOutputIDs, id, &ids))
	if err != nil {
		ids = nil
	}
	return ids
}
