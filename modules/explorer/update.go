package explorer

import (
	"fmt"

	"github.com/threefoldtech/rivine/build"
	"github.com/threefoldtech/rivine/encoding"
	"github.com/threefoldtech/rivine/modules"
	"github.com/threefoldtech/rivine/types"

	"github.com/rivine/bbolt"
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
				target = e.rootTarget
			}
			dbRemoveBlockTarget(tx, bid, target)

			// Remove miner payouts
			for j, payout := range block.MinerPayouts {
				scoid := block.MinerPayoutID(uint64(j))
				dbRemoveCoinOutputID(tx, scoid, tbid)
				dbRemoveUnlockHash(tx, payout.UnlockHash, tbid)
				dbRemoveCoinOutput(tx, scoid)
			}

			// Remove transactions
			for _, txn := range block.Transactions {
				txid := txn.ID()
				dbRemoveTransactionID(tx, txid)

				for _, sci := range txn.CoinInputs {
					dbRemoveCoinOutputID(tx, sci.ParentID, txid)
					unmapParentUnlockConditionHash(tx, sci.ParentID, txid)
				}
				for k, sco := range txn.CoinOutputs {
					scoid := txn.CoinOutputID(uint64(k))
					dbRemoveCoinOutputID(tx, scoid, txid)
					dbRemoveCoinOutput(tx, scoid)
					unmapUnlockConditionHash(tx, sco.Condition, txid)
				}
				for _, sfi := range txn.BlockStakeInputs {
					dbRemoveBlockStakeOutputID(tx, sfi.ParentID, txid)
					unmapParentUnlockConditionHash(tx, sfi.ParentID, txid)
				}
				for k, sfo := range txn.BlockStakeOutputs {
					sfoid := txn.BlockStakeOutputID(uint64(k))
					dbRemoveBlockStakeOutputID(tx, sfoid, txid)
					unmapUnlockConditionHash(tx, sfo.Condition, txid)
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
			if bid == e.genesisBlockID {
				e.dbAddGenesisBlock(tx)
				continue
			}

			blockheight++
			dbAddBlockID(tx, bid, blockheight)
			dbAddTransactionID(tx, tbid, blockheight) // Miner payouts are a transaction

			target, exists := e.cs.ChildTarget(block.ParentID)
			if !exists {
				target = e.rootTarget
			}
			dbAddBlockTarget(tx, bid, target)

			// Catalog the new miner payouts.
			for j, payout := range block.MinerPayouts {
				scoid := block.MinerPayoutID(uint64(j))
				dbAddCoinOutputID(tx, scoid, tbid)
				dbAddUnlockHash(tx, payout.UnlockHash, tbid)
				dbAddCoinOutput(tx, scoid, types.CoinOutput{
					Value: payout.Value,
					Condition: types.UnlockConditionProxy{
						Condition: types.NewUnlockHashCondition(payout.UnlockHash),
					},
				})
			}

			// Update cumulative stats for applied transactions.
			for _, txn := range block.Transactions {
				// Add the transaction to the list of active transactions.
				txid := txn.ID()
				dbAddTransactionID(tx, txid, blockheight)

				for _, sci := range txn.CoinInputs {
					dbAddCoinOutputID(tx, sci.ParentID, txid)
					mapParentUnlockConditionHash(tx, sci.ParentID, txid)
				}
				for j, sco := range txn.CoinOutputs {
					scoid := txn.CoinOutputID(uint64(j))
					dbAddCoinOutputID(tx, scoid, txid)
					mapUnlockConditionHash(tx, sco.Condition, txid)
				}
				for _, sfi := range txn.BlockStakeInputs {
					dbAddBlockStakeOutputID(tx, sfi.ParentID, txid)
					mapParentUnlockConditionHash(tx, sfi.ParentID, txid)
				}
				for k, sfo := range txn.BlockStakeOutputs {
					sfoid := txn.BlockStakeOutputID(uint64(k))
					dbAddBlockStakeOutputID(tx, sfoid, txid)
					mapUnlockConditionHash(tx, sfo.Condition, txid)
				}
			}

			// calculate and add new block facts, if possible
			if tx.Bucket(bucketBlockFacts).Get(encoding.Marshal(block.ParentID)) != nil {
				facts := e.dbCalculateBlockFacts(tx, block)
				dbAddBlockFacts(tx, facts)
			}
		}

		// Update stats according to CoinOutputDiffs
		for _, scod := range cc.CoinOutputDiffs {
			if scod.Direction == modules.DiffApply {
				dbAddCoinOutput(tx, scod.ID, scod.CoinOutput)
			}
		}

		// Update stats according to BlockStakeOutputDiffs
		for _, sfod := range cc.BlockStakeOutputDiffs {
			if sfod.Direction == modules.DiffApply {
				dbAddBlockStakeOutput(tx, sfod.ID, sfod.BlockStakeOutput)
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

func (e *Explorer) dbCalculateBlockFacts(tx *bolt.Tx, block types.Block) blockFacts {
	// get the parent block facts
	var bf blockFacts
	err := dbGetAndDecode(bucketBlockFacts, block.ParentID, &bf)(tx)
	assertNil(err)

	// get target
	target, exists := e.cs.ChildTarget(block.ParentID)
	if !exists {
		panic(fmt.Sprint("ConsensusSet is missing target of known block", block.ParentID))
	}

	// update fields
	bf.BlockID = block.ID()
	bf.Height++
	bf.Difficulty = target.Difficulty(e.chainCts.RootDepth)
	bf.Target = target
	bf.Timestamp = block.Timestamp
	//TODO rivine
	bf.TotalCoins = types.NewCurrency64(0)

	// calculate maturity timestamp
	var maturityTimestamp types.Timestamp
	if bf.Height > e.chainCts.MaturityDelay {
		oldBlock, exists := e.cs.BlockAtHeight(bf.Height - e.chainCts.MaturityDelay)
		if !exists {
			panic(fmt.Sprint("ConsensusSet is missing block at height", bf.Height-e.chainCts.MaturityDelay))
		}
		maturityTimestamp = oldBlock.Timestamp
	}
	bf.MaturityTimestamp = maturityTimestamp

	// calculate hashrate by averaging last 'ActiveBSEstimationBlocks' blocks
	var EstimatedActiveBS types.Difficulty
	if bf.Height > ActiveBSEstimationBlocks {
		var totalDifficulty = bf.Target
		var oldestTimestamp types.Timestamp
		for i := types.BlockHeight(1); i < ActiveBSEstimationBlocks; i++ {
			b, exists := e.cs.BlockAtHeight(bf.Height - i)
			if !exists {
				panic(fmt.Sprint("ConsensusSet is missing block at height", bf.Height-i))
			}
			target, exists := e.cs.ChildTarget(b.ParentID)
			if !exists {
				panic(fmt.Sprint("ConsensusSet is missing target of known block", b.ParentID))
			}
			totalDifficulty = totalDifficulty.AddDifficulties(
				target, e.chainCts.RootDepth)
			oldestTimestamp = b.Timestamp
		}
		secondsPassed := bf.Timestamp - oldestTimestamp
		EstimatedActiveBS = totalDifficulty.Difficulty(
			e.chainCts.RootDepth).Div64(uint64(secondsPassed))
	}
	bf.EstimatedActiveBS = EstimatedActiveBS

	bf.MinerPayoutCount += uint64(len(block.MinerPayouts))
	bf.TransactionCount += uint64(len(block.Transactions))
	for _, txn := range block.Transactions {
		bf.CoinInputCount += uint64(len(txn.CoinInputs))
		bf.CoinOutputCount += uint64(len(txn.CoinOutputs))
		bf.BlockStakeInputCount += uint64(len(txn.BlockStakeInputs))
		bf.BlockStakeOutputCount += uint64(len(txn.BlockStakeOutputs))
		bf.MinerFeeCount += uint64(len(txn.MinerFees))
		if size := len(txn.ArbitraryData); size > 0 {
			bf.ArbitraryDataTotalSize += uint64(size)
			bf.ArbitraryDataCount++
		}
	}

	return bf
}

// Special handling for the genesis block. No other functions are called on it.
func (e *Explorer) dbAddGenesisBlock(tx *bolt.Tx) {
	id := e.genesisBlockID
	dbAddBlockID(tx, id, 0)
	txid := e.genesisBlock.Transactions[0].ID()
	dbAddTransactionID(tx, txid, 0)
	for i, sco := range e.chainCts.GenesisCoinDistribution {
		scoid := e.genesisBlock.Transactions[0].CoinOutputID(uint64(i))
		dbAddCoinOutputID(tx, scoid, txid)
		mapUnlockConditionHash(tx, sco.Condition, txid)
		dbAddCoinOutput(tx, scoid, sco)
	}
	for i, sfo := range e.chainCts.GenesisBlockStakeAllocation {
		sfoid := e.genesisBlock.Transactions[0].BlockStakeOutputID(uint64(i))
		dbAddBlockStakeOutputID(tx, sfoid, txid)
		mapUnlockConditionHash(tx, sfo.Condition, txid)
		dbAddBlockStakeOutput(tx, sfoid, sfo)
	}
	dbAddBlockFacts(tx, blockFacts{
		BlockFacts: modules.BlockFacts{
			BlockID:               id,
			Height:                0,
			Difficulty:            e.rootTarget.Difficulty(e.rootTarget),
			Target:                e.rootTarget,
			TotalCoins:            types.NewCurrency64(0), //TODO rivine
			TransactionCount:      1,
			BlockStakeOutputCount: uint64(len(e.chainCts.GenesisBlockStakeAllocation)),
			CoinOutputCount:       uint64(len(e.chainCts.GenesisCoinDistribution)),
		},
		Timestamp: e.genesisBlock.Timestamp,
	})
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
func bucketIsEmpty(bucket *bolt.Bucket) bool {
	k, _ := bucket.Cursor().First()
	return k == nil
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
func dbAddCoinOutput(tx *bolt.Tx, id types.CoinOutputID, output types.CoinOutput) {
	mustPut(tx.Bucket(bucketCoinOutputs), id, output)
}
func dbRemoveCoinOutput(tx *bolt.Tx, id types.CoinOutputID) {
	mustDelete(tx.Bucket(bucketCoinOutputs), id)
}

// Add/Remove txid from siacoin output ID bucket
func dbAddCoinOutputID(tx *bolt.Tx, id types.CoinOutputID, txid types.TransactionID) {
	b, err := tx.Bucket(bucketCoinOutputIDs).CreateBucketIfNotExists(encoding.Marshal(id))
	assertNil(err)
	mustPutSet(b, txid)
}

func dbRemoveCoinOutputID(tx *bolt.Tx, id types.CoinOutputID, txid types.TransactionID) {
	bucket := tx.Bucket(bucketCoinOutputIDs).Bucket(encoding.Marshal(id))
	mustDelete(bucket, txid)
	if bucketIsEmpty(bucket) {
		tx.Bucket(bucketCoinOutputIDs).DeleteBucket(encoding.Marshal(id))
	}
}

// Add/Remove blockstake output
func dbAddBlockStakeOutput(tx *bolt.Tx, id types.BlockStakeOutputID, output types.BlockStakeOutput) {
	mustPut(tx.Bucket(bucketBlockStakeOutputs), id, output)
}
func dbRemoveBlockStakeOutput(tx *bolt.Tx, id types.BlockStakeOutputID) {
	mustDelete(tx.Bucket(bucketBlockStakeOutputs), id)
}

// Add/Remove txid from blockstake output ID bucket
func dbAddBlockStakeOutputID(tx *bolt.Tx, id types.BlockStakeOutputID, txid types.TransactionID) {
	b, err := tx.Bucket(bucketBlockStakeOutputIDs).CreateBucketIfNotExists(encoding.Marshal(id))
	assertNil(err)
	mustPutSet(b, txid)
}

func dbRemoveBlockStakeOutputID(tx *bolt.Tx, id types.BlockStakeOutputID, txid types.TransactionID) {
	bucket := tx.Bucket(bucketBlockStakeOutputIDs).Bucket(encoding.Marshal(id))
	mustDelete(bucket, txid)
	if bucketIsEmpty(bucket) {
		tx.Bucket(bucketBlockStakeOutputIDs).DeleteBucket(encoding.Marshal(id))
	}
}

// Add/Remove transaction ID
func dbAddTransactionID(tx *bolt.Tx, id types.TransactionID, height types.BlockHeight) {
	mustPut(tx.Bucket(bucketTransactionIDs), id, height)
}
func dbRemoveTransactionID(tx *bolt.Tx, id types.TransactionID) {
	mustDelete(tx.Bucket(bucketTransactionIDs), id)
}

func mapParentUnlockConditionHash(tx *bolt.Tx, parentID interface{}, txid types.TransactionID) {
	switch id := parentID.(type) {
	case types.CoinOutputID:
		var sco types.CoinOutput
		err := dbGetAndDecode(bucketCoinOutputs, id, &sco)(tx)
		if err != nil {
			return
		}
		mapUnlockConditionHash(tx, sco.Condition, txid)

	case types.BlockStakeOutputID:
		var bso types.BlockStakeOutput
		err := dbGetAndDecode(bucketBlockStakeOutputs, id, &bso)(tx)
		if err != nil {
			return
		}
		mapUnlockConditionHash(tx, bso.Condition, txid)

	default:
		panic(fmt.Sprintf("unexpected output ID type: %T", parentID))
	}
}
func unmapParentUnlockConditionHash(tx *bolt.Tx, parentID interface{}, txid types.TransactionID) {
	switch id := parentID.(type) {
	case types.CoinOutputID:
		var sco types.CoinOutput
		err := dbGetAndDecode(bucketCoinOutputs, id, &sco)(tx)
		if err != nil {
			return
		}
		unmapUnlockConditionHash(tx, sco.Condition, txid)

	case types.BlockStakeOutputID:
		var bso types.BlockStakeOutput
		err := dbGetAndDecode(bucketBlockStakeOutputs, id, &bso)(tx)
		if err != nil {
			return
		}
		unmapUnlockConditionHash(tx, bso.Condition, txid)

	default:
		panic(fmt.Sprintf("unexpected output ID type: %T", parentID))
	}
}
func mapUnlockConditionHash(tx *bolt.Tx, ucp types.UnlockConditionProxy, txid types.TransactionID) {
	muh := ucp.UnlockHash()
	dbAddUnlockHash(tx, muh, txid)
	if ucp.ConditionType() == types.ConditionTypeNil {
		return // nothing to do
	}
	mapUnlockConditionMultiSigAddress(tx, muh, ucp.Condition, txid)
}
func unmapUnlockConditionHash(tx *bolt.Tx, ucp types.UnlockConditionProxy, txid types.TransactionID) {
	muh := ucp.UnlockHash()
	dbRemoveUnlockHash(tx, muh, txid)
	if ucp.ConditionType() == types.ConditionTypeNil {
		return // nothing to do
	}
	unmapUnlockConditionMultiSigAddress(tx, muh, ucp.Condition, txid)
}
func mapUnlockConditionMultiSigAddress(tx *bolt.Tx, muh types.UnlockHash, cond types.MarshalableUnlockCondition, txid types.TransactionID) {
	switch cond.ConditionType() {
	case types.ConditionTypeTimeLock:
		cg, ok := cond.(types.MarshalableUnlockConditionGetter)
		if !ok {
			if build.DEBUG {
				panic(fmt.Sprintf("unexpected Go-type for TimeLockCondition: %T", cond))
			}
			return
		}
		cond = cg.GetMarshalableUnlockCondition()
		if cond == nil {
			if build.DEBUG {
				panic("unexpected nil-type for Internal condition of TimeLockCondition")
			}
			return
		}
		mapUnlockConditionMultiSigAddress(tx, muh, cond, txid)

	case types.ConditionTypeMultiSignature:
		mcond, ok := cond.(types.UnlockHashSliceGetter)
		if !ok {
			if build.DEBUG {
				panic(fmt.Sprintf("unexpected Go-type for MultiSignatureCondition: %T", cond))
			}
			return
		}
		// map the multisig address to all internal addresses
		for _, uh := range mcond.UnlockHashSlice() {
			dbAddWalletAddressToMultiSigAddressMapping(tx, uh, muh, txid)
		}
	}
}
func unmapUnlockConditionMultiSigAddress(tx *bolt.Tx, muh types.UnlockHash, cond types.MarshalableUnlockCondition, txid types.TransactionID) {
	switch cond.ConditionType() {
	case types.ConditionTypeTimeLock:
		cg, ok := cond.(types.MarshalableUnlockConditionGetter)
		if !ok {
			if build.DEBUG {
				panic(fmt.Sprintf("unexpected Go-type for TimeLockCondition: %T", cond))
			}
			return
		}
		cond = cg.GetMarshalableUnlockCondition()
		if cond == nil {
			if build.DEBUG {
				panic("unexpected nil-type for Internal condition of TimeLockCondition")
			}
			return
		}
		unmapUnlockConditionMultiSigAddress(tx, muh, cond, txid)

	case types.ConditionTypeMultiSignature:
		mcond, ok := cond.(types.UnlockHashSliceGetter)
		if !ok {
			if build.DEBUG {
				panic(fmt.Sprintf("unexpected Go-type for MultiSignatureCondition: %T", cond))
			}
			return
		}
		// unmap the multisig address to all internal addresses
		for _, uh := range mcond.UnlockHashSlice() {
			dbRemoveWalletAddressToMultiSigAddressMapping(tx, uh, muh, txid)
		}
	}
}

// Add/Remove txid from unlock hash bucket
func dbAddUnlockHash(tx *bolt.Tx, uh types.UnlockHash, txid types.TransactionID) {
	b, err := tx.Bucket(bucketUnlockHashes).CreateBucketIfNotExists(encoding.Marshal(uh))
	assertNil(err)
	mustPutSet(b, txid)
}
func dbRemoveUnlockHash(tx *bolt.Tx, uh types.UnlockHash, txid types.TransactionID) {
	uhb := tx.Bucket(bucketUnlockHashes)
	muh := encoding.Marshal(uh)
	b := uhb.Bucket(muh)
	mustDelete(b, txid)
	if bucketIsEmpty(b) {
		uhb.DeleteBucket(muh)
	}
}

// add/remove multisig address from wallet address mapping bucket
func dbAddWalletAddressToMultiSigAddressMapping(tx *bolt.Tx, walletAddress, multiSigAddress types.UnlockHash, txid types.TransactionID) {
	if build.DEBUG {
		if walletAddress.Type != types.UnlockTypePubKey {
			panic(fmt.Sprintf("wallet address has wrong type: %d", walletAddress.Type))
		}
		if multiSigAddress.Type != types.UnlockTypeMultiSig {
			panic(fmt.Sprintf("multisig address has wrong type: %d", multiSigAddress.Type))
		}
	}
	wab, err := tx.Bucket(bucketWalletAddressToMultiSigAddressMapping).CreateBucketIfNotExists(encoding.Marshal(walletAddress))
	assertNil(err)
	b, err := wab.CreateBucketIfNotExists(encoding.Marshal(multiSigAddress))
	assertNil(err)
	mustPutSet(b, txid)
}
func dbRemoveWalletAddressToMultiSigAddressMapping(tx *bolt.Tx, walletAddress, multiSigAddress types.UnlockHash, txid types.TransactionID) {
	if build.DEBUG {
		if walletAddress.Type != types.UnlockTypePubKey {
			panic(fmt.Sprintf("wallet address has wrong type: %d", walletAddress.Type))
		}
		if multiSigAddress.Type != types.UnlockTypeMultiSig {
			panic(fmt.Sprintf("multisig address has wrong type: %d", multiSigAddress.Type))
		}
	}
	mb := tx.Bucket(bucketWalletAddressToMultiSigAddressMapping)
	wa := encoding.Marshal(walletAddress)
	wb := mb.Bucket(wa)
	msa := encoding.Marshal(multiSigAddress)
	msb := wb.Bucket(msa)
	mustDelete(msb, txid)
	if bucketIsEmpty(msb) {
		wb.DeleteBucket(msa)
		if bucketIsEmpty(wb) {
			mb.DeleteBucket(wa)
		}
	}
}
