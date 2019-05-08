package explorer

import (
	"errors"

	"github.com/tarantool/go-tarantool"

	"github.com/threefoldtech/rivine/build"
	"github.com/threefoldtech/rivine/modules"
	"github.com/threefoldtech/rivine/pkg/encoding/siabin"
	"github.com/threefoldtech/rivine/types"
)

// Block takes a block ID and finds the corresponding block, provided that the
// block is in the consensus set.
func (e *Explorer) Block(id types.BlockID) (types.Block, types.BlockHeight, bool) {
	var height types.BlockHeight
	err := dbGetAndDecode("get_block", id, &height, e.client)
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
	err := e.dbGetBlockFacts(height, &bf)
	if err != nil {
		return modules.BlockFacts{}, false
	}

	return bf.BlockFacts, true
}

// LatestBlockFacts returns a set of statistics about the blockchain as they appeared
// at the latest block height in the explorer's consensus set.
func (e *Explorer) LatestBlockFacts() (modules.BlockFacts, error) {
	var bf blockFacts
	var height types.BlockHeight
	err := dbGetBlockheight(&height, e.client)
	if err != nil {
		return modules.BlockFacts{}, err
	}

	err = e.dbGetBlockFacts(height, &bf)
	if err != nil {
		build.Critical(err)
	}
	return bf.BlockFacts, nil
}

// Transaction takes a transaction ID and finds the block containing the
// transaction. Because of the miner payouts, the transaction ID might be a
// block ID. To find the transaction, iterate through the block.
func (e *Explorer) Transaction(id types.TransactionID) (types.Block, types.BlockHeight, bool) {
	var height types.BlockHeight
	err := dbGetAndDecode("get_transaction_blockheight", id, &height, e.client)
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
	err := dbGetTransactionIDSet("get_unlockhash_txid_set", uh, &ids, e.client)
	if err != nil {
		ids = nil
	}
	return ids
}

// MultiSigAddresses returns all multisig addresses this wallet address is involved in.
func (e *Explorer) MultiSigAddresses(uh types.UnlockHash) (uhs []types.UnlockHash, err error) {
	if uh.Type != types.UnlockTypePubKey {
		return nil, nil
	}

	valBytes, err := e.client.Get(UnlockConditionSpace, "primary", 0, 0, tarantool.IterAll, uh)
	if valBytes == nil {
		return nil, errors.New("not found")
	}
	for _, interf := range valBytes {
		bytes, ok := interf.([]byte)
		var uh types.UnlockHash
		if ok {
			return nil, siabin.Unmarshal(bytes, &uh)
		}
		uhs = append(uhs, uh)
	}

	return nil, nil
}

// CoinOutput returns the coin output associated with the specified ID.
func (e *Explorer) CoinOutput(id types.CoinOutputID) (types.CoinOutput, bool) {
	var sco types.CoinOutput
	err := dbGetAndDecode("get_coinoutput", id, &sco, e.client)
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
	err := dbGetTransactionIDSet("get_coinoutput_txid_set", id, &ids, e.client)
	if err != nil {
		ids = nil
	}
	return ids
}

// BlockStakeOutput returns the blockstake output associated with the specified ID.
func (e *Explorer) BlockStakeOutput(id types.BlockStakeOutputID) (types.BlockStakeOutput, bool) {
	var sco types.BlockStakeOutput
	err := dbGetAndDecode("get_blockstakeoutput", id, &sco, e.client)
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
	err := dbGetTransactionIDSet("get_bsoutput_txid_set", id, &ids, e.client)
	if err != nil {
		ids = nil
	}
	return ids
}

// HistoryStats return the stats for the last `history` amount of blocks
func (e *Explorer) HistoryStats(history types.BlockHeight) (*modules.ChainStats, error) {
	if history == 0 {
		return nil, errors.New("No history to show for 0 blocks")
	}
	// Get the current height
	var height types.BlockHeight
	err := dbGetBlockheight(&height, e.client)

	if err != nil {
		return nil, err
	}
	start := height - history + 1
	// Since blockheight is an uint64, we can't just check for a negative blockheight
	if history > height {
		start = 0
	}
	return e.getStats(start, height)
}

// RangeStats return the stats for the range [`start`, `end`]
func (e *Explorer) RangeStats(start types.BlockHeight, end types.BlockHeight) (*modules.ChainStats, error) {
	if start > end {
		return nil, errors.New("Invalid range")
	}
	// Get the current height
	var height types.BlockHeight
	err := dbGetBlockheight(&height, e.client)

	if err != nil || height < start {
		return nil, nil
	}
	if height < end {
		end = height
	}
	return e.getStats(start, end)
}

// getRangeStats fills in some stats from the blockfacts and the actual blocks in a specified range
func (e *Explorer) getStats(start types.BlockHeight, end types.BlockHeight) (*modules.ChainStats, error) {
	stats := modules.NewChainStats(int(end-start) + 1)

	for height := start; height <= end; height++ {

		// Load the block from the consensus set first so we have the ID, this saves a DB call to the
		// consensus set later on
		block, exists := e.cs.BlockAtHeight(height)
		if !exists {
			return nil, errors.New("Block does not exist in consensus set")
		}

		var facts blockFacts
		err := dbGetAndDecode("get_block", block.ID(), &facts, e.client)
		if err != nil {
			return nil, err
		}
		// Calculate the block index
		i := height - start

		stats.BlockHeights[i] = facts.Height
		stats.BlockTimeStamps[i] = facts.Timestamp
		if i > 0 {
			stats.BlockTimes[i] = int64(stats.BlockTimeStamps[i] - stats.BlockTimeStamps[i-1])
		}
		stats.EstimatedActiveBS[i] = facts.EstimatedActiveBS
		stats.Difficulties[i] = facts.Difficulty

		stats.TransactionCounts[i] = facts.TransactionCount
		stats.CoinInputCounts[i] = facts.CoinInputCount
		stats.CoinOutputCounts[i] = facts.CoinOutputCount
		stats.BlockStakeInputCounts[i] = facts.BlockStakeInputCount
		stats.BlockStakeOutputCounts[i] = facts.BlockStakeOutputCount

		stats.BlockTransactionCounts[i] = uint32(len(block.Transactions))
		// Don't count the transaction to respent the blockstake. However, it is possible
		// that this is an actual transaction (e.g. send blockstakes to someone else), and
		// is at the same time used to create the block.
		//
		// So we assume that:
		// 1. The block creating transaction is in index 0
		// 2. The block creating transaction does not pay a miner fee if the transaction is
		// 		created for the sole purpose of respending the BS so the block can be created
		if len(block.Transactions) > 0 && block.Transactions[0].MinerFees == nil {
			stats.BlockTransactionCounts[i]--
		}

		// Add the block creator to the node
		// Also genesis wan't created
		if height != 0 {
			creator := block.MinerPayouts[0].UnlockHash.String()
			_, exists = stats.Creators[creator]
			if !exists {
				stats.Creators[creator] = 1
			} else {
				stats.Creators[creator]++
			}
		}
	}
	// Set the creation time for the first block
	if start > 0 && stats.BlockCount > 0 {
		block, exists := e.cs.BlockAtHeight(start - 1)
		if !exists {
			return nil, errors.New("Block does not exist in consensus set")
		}
		stats.BlockTimes[0] = int64(stats.BlockTimeStamps[0] - block.Timestamp)
	}

	return stats, nil
}

// Constants returns all of the constants in use by the chain
func (e *Explorer) Constants() modules.DaemonConstants {
	return modules.NewDaemonConstants(e.bcInfo, e.chainCts)
}
