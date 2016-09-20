package explorer

import (
	"fmt"

	"github.com/rivine/rivine/build"
	"github.com/rivine/rivine/encoding"
	"github.com/rivine/rivine/modules"
	"github.com/rivine/rivine/types"

	"github.com/NebulousLabs/bolt"
)

// ProcessConsensusChange follows the most recent changes to the consensus set,
// including parsing new blocks and updating the utxo sets.
func (e *Explorer) ProcessConsensusChange(cc modules.ConsensusChange) {
	if len(cc.AppliedBlocks) == 0 {
		build.Critical("Explorer.ProcessConsensusChange called with a ConsensusChange that has no AppliedBlocks")
	}

	err := e.db.Update(func(tx *bolt.Tx) (err error) {
		// use exception-style error handling to enable more concise update code
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("%v", r)
			}
		}()

		// get starting block height
		var blockheight types.BlockHeight
		err = dbGetInternal(internalBlockHeight, &blockheight)(tx)
		if err != nil {
			return err
		}

		// Update cumulative stats for reverted blocks.
		for _, block := range cc.RevertedBlocks {
			bid := block.ID()
			tbid := types.TransactionID(bid)

			blockheight--
			dbRemoveBlockID(tx, bid)
			dbRemoveTransactionID(tx, tbid) // Miner payouts are a transaction

			target, exists := e.cs.ChildTarget(block.ParentID)
			if !exists {
				target = types.RootTarget
			}
			dbRemoveBlockTarget(tx, bid, target)

			// Remove miner payouts
			for j, payout := range block.MinerPayouts {
				scoid := block.MinerPayoutID(uint64(j))
				dbRemoveSiacoinOutputID(tx, scoid, tbid)
				dbRemoveUnlockHash(tx, payout.UnlockHash, tbid)
			}

			// Remove transactions
			for _, txn := range block.Transactions {
				txid := txn.ID()
				dbRemoveTransactionID(tx, txid)

				for _, sci := range txn.SiacoinInputs {
					dbRemoveSiacoinOutputID(tx, sci.ParentID, txid)
					dbRemoveUnlockHash(tx, sci.UnlockConditions.UnlockHash(), txid)
				}
				for k, sco := range txn.SiacoinOutputs {
					scoid := txn.SiacoinOutputID(uint64(k))
					dbRemoveSiacoinOutputID(tx, scoid, txid)
					dbRemoveUnlockHash(tx, sco.UnlockHash, txid)
					dbRemoveSiacoinOutput(tx, scoid)
				}
				for _, sfi := range txn.SiafundInputs {
					dbRemoveSiafundOutputID(tx, sfi.ParentID, txid)
					dbRemoveUnlockHash(tx, sfi.UnlockConditions.UnlockHash(), txid)
					dbRemoveUnlockHash(tx, sfi.ClaimUnlockHash, txid)
				}
				for k, sfo := range txn.SiafundOutputs {
					sfoid := txn.SiafundOutputID(uint64(k))
					dbRemoveSiafundOutputID(tx, sfoid, txid)
					dbRemoveUnlockHash(tx, sfo.UnlockHash, txid)
				}
			}

			// remove the associated block facts
			dbRemoveBlockFacts(tx, bid)
		}

		// Update cumulative stats for applied blocks.
		for _, block := range cc.AppliedBlocks {
			bid := block.ID()
			tbid := types.TransactionID(bid)

			// special handling for genesis block
			if bid == types.GenesisID {
				dbAddGenesisBlock(tx)
				continue
			}

			blockheight++
			dbAddBlockID(tx, bid, blockheight)
			dbAddTransactionID(tx, tbid, blockheight) // Miner payouts are a transaction

			target, exists := e.cs.ChildTarget(block.ParentID)
			if !exists {
				target = types.RootTarget
			}
			dbAddBlockTarget(tx, bid, target)

			// Catalog the new miner payouts.
			for j, payout := range block.MinerPayouts {
				scoid := block.MinerPayoutID(uint64(j))
				dbAddSiacoinOutputID(tx, scoid, tbid)
				dbAddUnlockHash(tx, payout.UnlockHash, tbid)
			}

			// Update cumulative stats for applied transactions.
			for _, txn := range block.Transactions {
				// Add the transaction to the list of active transactions.
				txid := txn.ID()
				dbAddTransactionID(tx, txid, blockheight)

				for _, sci := range txn.SiacoinInputs {
					dbAddSiacoinOutputID(tx, sci.ParentID, txid)
					dbAddUnlockHash(tx, sci.UnlockConditions.UnlockHash(), txid)
				}
				for j, sco := range txn.SiacoinOutputs {
					scoid := txn.SiacoinOutputID(uint64(j))
					dbAddSiacoinOutputID(tx, scoid, txid)
					dbAddUnlockHash(tx, sco.UnlockHash, txid)
					dbAddSiacoinOutput(tx, scoid, sco)
				}
				for _, sfi := range txn.SiafundInputs {
					dbAddSiafundOutputID(tx, sfi.ParentID, txid)
					dbAddUnlockHash(tx, sfi.UnlockConditions.UnlockHash(), txid)
					dbAddUnlockHash(tx, sfi.ClaimUnlockHash, txid)
				}
				for k, sfo := range txn.SiafundOutputs {
					sfoid := txn.SiafundOutputID(uint64(k))
					dbAddSiafundOutputID(tx, sfoid, txid)
					dbAddUnlockHash(tx, sfo.UnlockHash, txid)
					dbAddSiafundOutput(tx, sfoid, sfo)
				}
			}

			// calculate and add new block facts, if possible
			if tx.Bucket(bucketBlockFacts).Get(encoding.Marshal(block.ParentID)) != nil {
				facts := dbCalculateBlockFacts(tx, e.cs, block)
				dbAddBlockFacts(tx, facts)
			}
		}

		// Compute the changes in the active set. Note, because this is calculated
		// at the end instead of in a loop, the historic facts may contain
		// inaccuracies about the active set. This should not be a problem except
		// for large reorgs.
		// TODO: improve this
		currentBlock, exists := e.cs.BlockAtHeight(blockheight)
		if !exists {
			build.Critical("consensus is missing block", blockheight)
		}
		currentID := currentBlock.ID()
		var facts blockFacts
		err = dbGetAndDecode(bucketBlockFacts, currentID, &facts)(tx)
		if err == nil {
			err = tx.Bucket(bucketBlockFacts).Put(encoding.Marshal(currentID), encoding.Marshal(facts))
			if err != nil {
				return err
			}
		}

		// set final blockheight
		err = dbSetInternal(internalBlockHeight, blockheight)(tx)
		if err != nil {
			return err
		}

		// set change ID
		err = dbSetInternal(internalRecentChange, cc.ID)(tx)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		build.Critical("explorer update failed:", err)
	}
}

// helper functions
func assertNil(err error) {
	if err != nil {
		panic(err)
	}
}
func mustPut(bucket *bolt.Bucket, key, val interface{}) {
	assertNil(bucket.Put(encoding.Marshal(key), encoding.Marshal(val)))
}
func mustPutSet(bucket *bolt.Bucket, key interface{}) {
	assertNil(bucket.Put(encoding.Marshal(key), nil))
}
func mustDelete(bucket *bolt.Bucket, key interface{}) {
	assertNil(bucket.Delete(encoding.Marshal(key)))
}

// These functions panic on error. The panic will be caught by
// ProcessConsensusChange.

// Add/Remove block ID
func dbAddBlockID(tx *bolt.Tx, id types.BlockID, height types.BlockHeight) {
	mustPut(tx.Bucket(bucketBlockIDs), id, height)
}
func dbRemoveBlockID(tx *bolt.Tx, id types.BlockID) {
	mustDelete(tx.Bucket(bucketBlockIDs), id)
}

// Add/Remove block facts
func dbAddBlockFacts(tx *bolt.Tx, facts blockFacts) {
	mustPut(tx.Bucket(bucketBlockFacts), facts.BlockID, facts)
}
func dbRemoveBlockFacts(tx *bolt.Tx, id types.BlockID) {
	mustDelete(tx.Bucket(bucketBlockFacts), id)
}

// Add/Remove block target
func dbAddBlockTarget(tx *bolt.Tx, id types.BlockID, target types.Target) {
	mustPut(tx.Bucket(bucketBlockTargets), id, target)
}
func dbRemoveBlockTarget(tx *bolt.Tx, id types.BlockID, target types.Target) {
	mustDelete(tx.Bucket(bucketBlockTargets), id)
}

// Add/Remove siacoin output
func dbAddSiacoinOutput(tx *bolt.Tx, id types.SiacoinOutputID, output types.SiacoinOutput) {
	mustPut(tx.Bucket(bucketSiacoinOutputs), id, output)
}
func dbRemoveSiacoinOutput(tx *bolt.Tx, id types.SiacoinOutputID) {
	mustDelete(tx.Bucket(bucketSiacoinOutputs), id)
}

// Add/Remove txid from siacoin output ID bucket
func dbAddSiacoinOutputID(tx *bolt.Tx, id types.SiacoinOutputID, txid types.TransactionID) {
	b, err := tx.Bucket(bucketSiacoinOutputIDs).CreateBucketIfNotExists(encoding.Marshal(id))
	assertNil(err)
	mustPutSet(b, txid)
}
func dbRemoveSiacoinOutputID(tx *bolt.Tx, id types.SiacoinOutputID, txid types.TransactionID) {
	// TODO: delete bucket when it becomes empty
	mustDelete(tx.Bucket(bucketSiacoinOutputIDs).Bucket(encoding.Marshal(id)), txid)
}

// Add/Remove siafund output
func dbAddSiafundOutput(tx *bolt.Tx, id types.SiafundOutputID, output types.SiafundOutput) {
	mustPut(tx.Bucket(bucketSiafundOutputs), id, output)
}
func dbRemoveSiafundOutput(tx *bolt.Tx, id types.SiafundOutputID) {
	mustDelete(tx.Bucket(bucketSiafundOutputs), id)
}

// Add/Remove txid from siafund output ID bucket
func dbAddSiafundOutputID(tx *bolt.Tx, id types.SiafundOutputID, txid types.TransactionID) {
	b, err := tx.Bucket(bucketSiafundOutputIDs).CreateBucketIfNotExists(encoding.Marshal(id))
	assertNil(err)
	mustPutSet(b, txid)
}
func dbRemoveSiafundOutputID(tx *bolt.Tx, id types.SiafundOutputID, txid types.TransactionID) {
	// TODO: delete bucket when it becomes empty
	mustDelete(tx.Bucket(bucketSiafundOutputIDs).Bucket(encoding.Marshal(id)), txid)
}

// Add/Remove transaction ID
func dbAddTransactionID(tx *bolt.Tx, id types.TransactionID, height types.BlockHeight) {
	mustPut(tx.Bucket(bucketTransactionIDs), id, height)
}
func dbRemoveTransactionID(tx *bolt.Tx, id types.TransactionID) {
	mustDelete(tx.Bucket(bucketTransactionIDs), id)
}

// Add/Remove txid from unlock hash bucket
func dbAddUnlockHash(tx *bolt.Tx, uh types.UnlockHash, txid types.TransactionID) {
	b, err := tx.Bucket(bucketUnlockHashes).CreateBucketIfNotExists(encoding.Marshal(uh))
	assertNil(err)
	mustPutSet(b, txid)
}
func dbRemoveUnlockHash(tx *bolt.Tx, uh types.UnlockHash, txid types.TransactionID) {
	// TODO: delete bucket when it becomes empty
	mustDelete(tx.Bucket(bucketUnlockHashes).Bucket(encoding.Marshal(uh)), txid)
}

func dbCalculateBlockFacts(tx *bolt.Tx, cs modules.ConsensusSet, block types.Block) blockFacts {
	// get the parent block facts
	var bf blockFacts
	err := dbGetAndDecode(bucketBlockFacts, block.ParentID, &bf)(tx)
	assertNil(err)

	// get target
	target, exists := cs.ChildTarget(block.ParentID)
	if !exists {
		panic(fmt.Sprint("ConsensusSet is missing target of known block", block.ParentID))
	}

	// update fields
	bf.BlockID = block.ID()
	bf.Height++
	bf.Difficulty = target.Difficulty()
	bf.Target = target
	bf.Timestamp = block.Timestamp
	bf.TotalCoins = types.CalculateNumSiacoins(bf.Height)

	// calculate maturity timestamp
	var maturityTimestamp types.Timestamp
	if bf.Height > types.MaturityDelay {
		oldBlock, exists := cs.BlockAtHeight(bf.Height - types.MaturityDelay)
		if !exists {
			panic(fmt.Sprint("ConsensusSet is missing block at height", bf.Height-types.MaturityDelay))
		}
		maturityTimestamp = oldBlock.Timestamp
	}
	bf.MaturityTimestamp = maturityTimestamp

	// calculate hashrate by averaging last 'hashrateEstimationBlocks' blocks
	var estimatedHashrate types.Currency
	if bf.Height > hashrateEstimationBlocks {
		var totalDifficulty = bf.Target
		var oldestTimestamp types.Timestamp
		for i := types.BlockHeight(1); i < hashrateEstimationBlocks; i++ {
			b, exists := cs.BlockAtHeight(bf.Height - i)
			if !exists {
				panic(fmt.Sprint("ConsensusSet is missing block at height", bf.Height-hashrateEstimationBlocks))
			}
			target, exists := cs.ChildTarget(b.ParentID)
			if !exists {
				panic(fmt.Sprint("ConsensusSet is missing target of known block", b.ParentID))
			}
			totalDifficulty = totalDifficulty.AddDifficulties(target)
			oldestTimestamp = b.Timestamp
		}
		secondsPassed := bf.Timestamp - oldestTimestamp
		estimatedHashrate = totalDifficulty.Difficulty().Div64(uint64(secondsPassed))
	}
	bf.EstimatedHashrate = estimatedHashrate

	bf.MinerPayoutCount += uint64(len(block.MinerPayouts))
	bf.TransactionCount += uint64(len(block.Transactions))
	for _, txn := range block.Transactions {
		bf.SiacoinInputCount += uint64(len(txn.SiacoinInputs))
		bf.SiacoinOutputCount += uint64(len(txn.SiacoinOutputs))
		bf.SiafundInputCount += uint64(len(txn.SiafundInputs))
		bf.SiafundOutputCount += uint64(len(txn.SiafundOutputs))
		bf.MinerFeeCount += uint64(len(txn.MinerFees))
		bf.ArbitraryDataCount += uint64(len(txn.ArbitraryData))
		bf.TransactionSignatureCount += uint64(len(txn.TransactionSignatures))

	}

	return bf
}

// Special handling for the genesis block. No other functions are called on it.
func dbAddGenesisBlock(tx *bolt.Tx) {
	id := types.GenesisID
	dbAddBlockID(tx, id, 0)
	txid := types.GenesisBlock.Transactions[0].ID()
	dbAddTransactionID(tx, txid, 0)
	for i, sfo := range types.GenesisSiafundAllocation {
		sfoid := types.GenesisBlock.Transactions[0].SiafundOutputID(uint64(i))
		dbAddSiafundOutputID(tx, sfoid, txid)
		dbAddUnlockHash(tx, sfo.UnlockHash, txid)
		dbAddSiafundOutput(tx, sfoid, sfo)
	}
	dbAddBlockFacts(tx, blockFacts{
		BlockFacts: modules.BlockFacts{
			BlockID:            id,
			Height:             0,
			Difficulty:         types.RootTarget.Difficulty(),
			Target:             types.RootTarget,
			TotalCoins:         types.CalculateCoinbase(0),
			TransactionCount:   1,
			SiafundOutputCount: uint64(len(types.GenesisSiafundAllocation)),
		},
		Timestamp: types.GenesisBlock.Timestamp,
	})
}
