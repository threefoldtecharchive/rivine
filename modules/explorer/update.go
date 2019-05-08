package explorer

import (
	"fmt"

	"github.com/threefoldtech/rivine/build"
	"github.com/threefoldtech/rivine/modules"
	"github.com/threefoldtech/rivine/pkg/encoding/siabin"
	persist "github.com/threefoldtech/rivine/tarantool-persist"
	"github.com/threefoldtech/rivine/types"

	bolt "github.com/rivine/bbolt"
)

// ProcessConsensusChange follows the most recent changes to the consensus set,
// including parsing new blocks and updating the utxo sets.
func (e *Explorer) ProcessConsensusChange(cc modules.ConsensusChange) {
	if len(cc.AppliedBlocks) == 0 {
		build.Critical("Explorer.ProcessConsensusChange called with a ConsensusChange that has no AppliedBlocks")
	}

	// use exception-style error handling to enable more concise update code
	defer func() {
		if r := recover(); r != nil {
			err := fmt.Errorf("%v", r)
			panic(err)
		}
	}()

	// get starting block height
	var blockheight types.BlockHeight
	err := dbGetStartingBlockheight(&blockheight, e.client)
	if err != nil {
		return
	}

	// Update cumulative stats for reverted blocks.
	for _, block := range cc.RevertedBlocks {
		bid := block.ID()
		tbid := types.TransactionID(bid)

		blockheight--
		dbRemoveBlockID(e.client, bid)
		dbRemoveTransactionID(e.client, tbid) // Miner payouts are a transaction

		target, exists := e.cs.ChildTarget(block.ParentID)
		if !exists {
			target = e.rootTarget
		}
		dbRemoveBlockTarget(e.client, bid, target)

		// Remove miner payouts
		for j, payout := range block.MinerPayouts {
			scoid := block.MinerPayoutID(uint64(j))
			dbRemoveCoinOutputID(e.client, scoid, tbid)
			dbRemoveUnlockHash(e.client, payout.UnlockHash, tbid)
			dbRemoveCoinOutput(e.client, scoid)
		}

		// Remove transactions
		for _, txn := range block.Transactions {
			txid := txn.ID()
			dbRemoveTransactionID(e.client, txid)

			for _, sci := range txn.CoinInputs {
				dbRemoveCoinOutputID(e.client, sci.ParentID, txid)
				unmapParentUnlockConditionHash(e.client, sci.ParentID, txid)
			}
			for k, sco := range txn.CoinOutputs {
				scoid := txn.CoinOutputID(uint64(k))
				dbRemoveCoinOutputID(e.client, scoid, txid)
				dbRemoveCoinOutput(e.client, scoid)
				unmapUnlockConditionHash(e.client, sco.Condition, txid)
			}
			for _, sfi := range txn.BlockStakeInputs {
				dbRemoveBlockStakeOutputID(e.client, sfi.ParentID, txid)
				unmapParentUnlockConditionHash(e.client, sfi.ParentID, txid)
			}
			for k, sfo := range txn.BlockStakeOutputs {
				sfoid := txn.BlockStakeOutputID(uint64(k))
				dbRemoveBlockStakeOutputID(e.client, sfoid, txid)
				dbRemoveBlockStakeOutput(e.client, sfoid)
				unmapUnlockConditionHash(e.client, sfo.Condition, txid)
			}

			// remove any common extension data, should the txn have it
			exData, _ := txn.CommonExtensionData()
			for _, condition := range exData.UnlockConditions {
				unmapUnlockConditionHash(e.client, condition, txid)
			}
		}

		// remove the associated block facts
		dbRemoveBlockFacts(e.client, bid)
	}

	// Update cumulative stats for applied blocks.
	for _, block := range cc.AppliedBlocks {
		bid := block.ID()
		tbid := types.TransactionID(bid)

		// special handling for genesis block
		if bid == e.genesisBlockID {
			e.dbAddGenesisBlock()
			continue
		}

		blockheight++
		dbAddBlockID(e.client, bid, blockheight)
		dbAddTransactionID(e.client, tbid, blockheight) // Miner payouts are a transaction

		target, exists := e.cs.ChildTarget(block.ParentID)
		if !exists {
			target = e.rootTarget
		}
		dbAddBlockTarget(e.client, bid, target)

		// Catalog the new miner payouts.
		for j, payout := range block.MinerPayouts {
			scoid := block.MinerPayoutID(uint64(j))
			dbAddCoinOutputID(e.client, scoid, tbid)
			dbAddUnlockHash(e.client, payout.UnlockHash, tbid)
			dbAddCoinOutput(e.client, scoid, types.CoinOutput{
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
			dbAddTransactionID(e.client, txid, blockheight)

			for j, sco := range txn.CoinOutputs {
				scoid := txn.CoinOutputID(uint64(j))
				dbAddCoinOutputID(e.client, scoid, txid)
				dbAddCoinOutput(e.client, scoid, sco)
				mapUnlockConditionHash(e.client, sco.Condition, txid)
			}
			for _, sci := range txn.CoinInputs {
				dbAddCoinOutputID(e.client, sci.ParentID, txid)
				mapParentUnlockConditionHash(e.client, sci.ParentID, txid)
			}
			for k, sfo := range txn.BlockStakeOutputs {
				sfoid := txn.BlockStakeOutputID(uint64(k))
				dbAddBlockStakeOutputID(e.client, sfoid, txid)
				dbAddBlockStakeOutput(e.client, sfoid, sfo)
				mapUnlockConditionHash(e.client, sfo.Condition, txid)
			}
			for _, sfi := range txn.BlockStakeInputs {
				dbAddBlockStakeOutputID(e.client, sfi.ParentID, txid)
				mapParentUnlockConditionHash(e.client, sfi.ParentID, txid)
			}

			// add any common extension data, should the txn have it
			exData, _ := txn.CommonExtensionData()
			for _, condition := range exData.UnlockConditions {
				mapUnlockConditionHash(e.client, condition, txid)
			}
		}

		// calculate and add new block facts, if possible
		// if tx.Bucket(bucketBlockFacts).Get(siabin.Marshal(block.ParentID)) != nil {
		// 	facts := e.dbCalculateBlockFacts(tx, block)
		// 	dbAddBlockFacts(tx, facts)
		// }
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
	currentID := currentBlock.ID().String()
	var facts blockFacts
	err = dbGetAndDecode(BlockSpace, currentID, &facts, e.client)
	if err == nil {
		// space := e.client.Schema.Spaces[BlockSpace]
		e.client.Insert(BlockSpace, []interface{}{siabin.Marshal(currentID), siabin.Marshal(facts)})
		// err = tx.Bucket(spaceBlockFacts).Put(siabin.Marshal(currentID), siabin.Marshal(facts))
		// if err != nil {
		// 	return err
		// }
	}

	// set final blockheight
	// err = dbSetInternal("blockheight", blockheight, e.client)
	// if err != nil {
	// 	return
	// }

	// set change ID
	err = dbSetConsensusChangeID(cc.ID, e.client)
	if err != nil {
		return
	}

	return

	if err != nil {
		build.Critical("explorer update failed:", err)
	}
}

func (e *Explorer) dbCalculateBlockFacts(tx *bolt.Tx, block types.Block) blockFacts {
	// get the parent block facts
	var bf blockFacts
	err := dbGetAndDecode(BlockSpace, block.ParentID, &bf, e.client)
	if err != nil {
		return blockFacts{}
	}
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
func (e *Explorer) dbAddGenesisBlock() {
	id := e.genesisBlockID
	dbAddBlockID(e.client, id, 0)
	txid := e.genesisBlock.Transactions[0].ID()
	dbAddTransactionID(e.client, txid, 0)
	for i, sco := range e.chainCts.GenesisCoinDistribution {
		scoid := e.genesisBlock.Transactions[0].CoinOutputID(uint64(i))
		dbAddCoinOutputID(e.client, scoid, txid)
		mapUnlockConditionHash(e.client, sco.Condition, txid)
		dbAddCoinOutput(e.client, scoid, sco)
	}
	for i, sfo := range e.chainCts.GenesisBlockStakeAllocation {
		sfoid := e.genesisBlock.Transactions[0].BlockStakeOutputID(uint64(i))
		dbAddBlockStakeOutputID(e.client, sfoid, txid)
		mapUnlockConditionHash(e.client, sfo.Condition, txid)
		dbAddBlockStakeOutput(e.client, sfoid, sfo)
	}
	dbAddBlockFacts(e.client, blockFacts{
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
func assertNil(a []interface{}, err error) {
	if err != nil {
		panic(err)
	}
}

// func mustPut(bucket *bolt.Bucket, key, val interface{}) {
// 	assertNil(bucket.Put(siabin.Marshal(key), siabin.Marshal(val)))
// }
// func mustPutSet(bucket *bolt.Bucket, key interface{}) {
// 	assertNil(bucket.Put(siabin.Marshal(key), nil))
// }
// func mustDelete(bucket *bolt.Bucket, key interface{}) {
// 	assertNil(bucket.Delete(siabin.Marshal(key)))
// }
func bucketIsEmpty(bucket *bolt.Bucket) bool {
	k, _ := bucket.Cursor().First()
	return k == nil
}

// These functions panic on error. The panic will be caught by
// ProcessConsensusChange.

// Add/Remove block ID
func dbAddBlockID(client *persist.TarantoolClient, id types.BlockID, height types.BlockHeight) {
	res, err := client.Insert(BlockSpace, []interface{}{id.String(), height})
	if err == persist.ErrDuplicateKey {
		return
	}
	assertNil(res, err)
}
func dbRemoveBlockID(client *persist.TarantoolClient, id types.BlockID) {
	assertNil(client.Delete(BlockSpace, BlockHeightIndex, id))
}

// Add/Remove block facts
func dbAddBlockFacts(client *persist.TarantoolClient, facts blockFacts) {
	res, err := client.Insert(BlockSpace, []interface{}{facts.BlockID.String(), facts})
	if err == persist.ErrDuplicateKey {
		return
	}
	assertNil(res, err)
}
func dbRemoveBlockFacts(client *persist.TarantoolClient, id types.BlockID) {
	assertNil(client.Delete(BlockSpace, BlockIDIndex, id))
}

// Add/Remove block target
func dbAddBlockTarget(client *persist.TarantoolClient, id types.BlockID, target types.Target) {
	res, err := client.Insert(TransactionSpace, []interface{}{id.String(), target})
	if err == persist.ErrDuplicateKey {
		return
	}
	assertNil(res, err)
}
func dbRemoveBlockTarget(client *persist.TarantoolClient, id types.BlockID, target types.Target) {
	assertNil(client.Delete(TransactionSpace, TransactionIDIndex, id))
}

// Add/Remove siacoin output
func dbAddCoinOutput(client *persist.TarantoolClient, id types.CoinOutputID, output types.CoinOutput) {
	res, err := client.Insert(TransactionSpace, []interface{}{id.String(), output})
	if err == persist.ErrDuplicateKey {
		return
	}
	assertNil(res, err)
}
func dbRemoveCoinOutput(client *persist.TarantoolClient, id types.CoinOutputID) {
	assertNil(client.Delete(TransactionSpace, TransactionIDIndex, id))
}

// Add/Remove txid from siacoin output ID bucket
func dbAddCoinOutputID(client *persist.TarantoolClient, id types.CoinOutputID, txid types.TransactionID) {
	res, err := client.Insert(TransactionSpace, []interface{}{id.String(), txid})
	if err == persist.ErrDuplicateKey {
		return
	}
	assertNil(res, err)
}

func dbRemoveCoinOutputID(client *persist.TarantoolClient, id types.CoinOutputID, txid types.TransactionID) {
	assertNil(client.Delete(TransactionSpace, TransactionIDIndex, id))

	// bucket := tx.Bucket(bucketCoinOutputIDs).Bucket(siabin.Marshal(id))
	// mustDelete(bucket, txid)
	// if bucketIsEmpty(bucket) {
	// 	tx.Bucket(bucketCoinOutputIDs).DeleteBucket(siabin.Marshal(id))
	// }
}

// Add/Remove blockstake output
func dbAddBlockStakeOutput(client *persist.TarantoolClient, id types.BlockStakeOutputID, output types.BlockStakeOutput) {
	res, err := client.Insert(TransactionSpace, []interface{}{id.String(), output})
	if err == persist.ErrDuplicateKey {
		return
	}
	assertNil(res, err)
}
func dbRemoveBlockStakeOutput(client *persist.TarantoolClient, id types.BlockStakeOutputID) {

	client.Delete(TransactionSpace, TransactionIDIndex, id)

}

// Add/Remove txid from blockstake output ID bucket
func dbAddBlockStakeOutputID(client *persist.TarantoolClient, id types.BlockStakeOutputID, txid types.TransactionID) {
	res, err := client.Insert(TransactionSpace, []interface{}{id.String(), txid})
	if err == persist.ErrDuplicateKey {
		return
	}
	assertNil(res, err)
}

func dbRemoveBlockStakeOutputID(client *persist.TarantoolClient, id types.BlockStakeOutputID, txid types.TransactionID) {
	client.Delete(TransactionSpace, TransactionIDIndex, txid)
	//bucket := tx.Bucket(bucketBlockStakeOutputIDs).Bucket(siabin.Marshal(id))
	// mustDelete(bucket, txid)
	// if bucketIsEmpty(bucket) {
	// 	tx.Bucket(bucketBlockStakeOutputIDs).DeleteBucket(siabin.Marshal(id))
	// }
}

// Add/Remove transaction ID
func dbAddTransactionID(client *persist.TarantoolClient, id types.TransactionID, height types.BlockHeight) {
	res, err := client.Insert(TransactionSpace, []interface{}{id.String(), height})
	if err == persist.ErrDuplicateKey {
		return
	}
	assertNil(res, err)
}
func dbRemoveTransactionID(client *persist.TarantoolClient, id types.TransactionID) {
	// space := client.Schema.Spaces[TransactionSpace]
	assertNil(client.Delete(TransactionSpace, TransactionIDIndex, id.String()))
	// mustDelete(tx.Bucket(bucketTransactionIDs), id)
}

func mapParentUnlockConditionHash(client *persist.TarantoolClient, parentID interface{}, txid types.TransactionID) {
	switch id := parentID.(type) {
	case types.CoinOutputID:
		var sco types.CoinOutput
		err := dbGetAndDecode(TransactionSpace, id, &sco, client)
		if err != nil {
			return
		}
		mapUnlockConditionHash(client, sco.Condition, txid)

	case types.BlockStakeOutputID:
		var bso types.BlockStakeOutput
		err := dbGetAndDecode(TransactionSpace, id, &bso, client)
		if err != nil {
			return
		}
		mapUnlockConditionHash(client, bso.Condition, txid)

	default:
		panic(fmt.Sprintf("unexpected output ID type: %T", parentID))
	}
}
func unmapParentUnlockConditionHash(client *persist.TarantoolClient, parentID interface{}, txid types.TransactionID) {
	switch id := parentID.(type) {
	case types.CoinOutputID:
		var sco types.CoinOutput
		err := dbGetAndDecode(TransactionSpace, id, &sco, client)
		if err != nil {
			return
		}
		unmapUnlockConditionHash(client, sco.Condition, txid)

	case types.BlockStakeOutputID:
		var bso types.BlockStakeOutput
		err := dbGetAndDecode(TransactionSpace, id, &bso, client)
		if err != nil {
			return
		}
		unmapUnlockConditionHash(client, bso.Condition, txid)

	default:
		panic(fmt.Sprintf("unexpected output ID type: %T", parentID))
	}
}
func mapUnlockConditionHash(client *persist.TarantoolClient, ucp types.UnlockConditionProxy, txid types.TransactionID) {
	muh := ucp.UnlockHash()
	dbAddUnlockHash(client, muh, txid)
	if ucp.ConditionType() == types.ConditionTypeNil {
		return // nothing to do
	}
	mapUnlockConditionMultiSigAddress(client, muh, ucp.Condition, txid)
}
func unmapUnlockConditionHash(client *persist.TarantoolClient, ucp types.UnlockConditionProxy, txid types.TransactionID) {
	muh := ucp.UnlockHash()
	dbRemoveUnlockHash(client, muh, txid)
	if ucp.ConditionType() == types.ConditionTypeNil {
		return // nothing to do
	}
	unmapUnlockConditionMultiSigAddress(client, muh, ucp.Condition, txid)
}
func mapUnlockConditionMultiSigAddress(client *persist.TarantoolClient, muh types.UnlockHash, cond types.MarshalableUnlockCondition, txid types.TransactionID) {
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
		mapUnlockConditionMultiSigAddress(client, muh, cond, txid)

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
			dbAddWalletAddressToMultiSigAddressMapping(client, uh, muh, txid)
		}
	}
}
func unmapUnlockConditionMultiSigAddress(client *persist.TarantoolClient, muh types.UnlockHash, cond types.MarshalableUnlockCondition, txid types.TransactionID) {
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
		unmapUnlockConditionMultiSigAddress(client, muh, cond, txid)

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
			dbRemoveWalletAddressToMultiSigAddressMapping(client, uh, muh, txid)
		}
	}
}

// Add/Remove txid from unlock hash bucket
func dbAddUnlockHash(client *persist.TarantoolClient, uh types.UnlockHash, txid types.TransactionID) {
	res, err := client.Insert(UnlockConditionSpace, []interface{}{txid.String(), siabin.Marshal(uh)})
	if err == persist.ErrDuplicateKey {
		return
	}
	assertNil(res, err)

}
func dbRemoveUnlockHash(client *persist.TarantoolClient, uh types.UnlockHash, txid types.TransactionID) {
	// space := client.Schema.Spaces[UnlockConditionSpace]

	// uhb := tx.Bucket(bucketUnlockHashes)
	// muh := siabin.Marshal(uh)
	// b := uhb.Bucket(muh)
	// mustDelete(b, txid)
	assertNil(client.Delete(UnlockConditionSpace, UnlockConditionIDIndex, txid.String()))
	// if bucketIsEmpty(b) {
	// 	uhb.DeleteBucket(muh)
	// }
}

// add/remove multisig address from wallet address mapping bucket
func dbAddWalletAddressToMultiSigAddressMapping(client *persist.TarantoolClient, walletAddress, multiSigAddress types.UnlockHash, txid types.TransactionID) {
	if build.DEBUG {
		if walletAddress.Type != types.UnlockTypePubKey {
			panic(fmt.Sprintf("wallet address has wrong type: %d", walletAddress.Type))
		}
		if multiSigAddress.Type != types.UnlockTypeMultiSig {
			panic(fmt.Sprintf("multisig address has wrong type: %d", multiSigAddress.Type))
		}
	}
	// space := client.Schema.Spaces[UnlockConditionSpace]

	// wab, err := tx.Bucket(bucketWalletAddressToMultiSigAddressMapping).CreateBucketIfNotExists(siabin.Marshal(walletAddress))
	// assertNil(err)
	// b, err := wab.CreateBucketIfNotExists(siabin.Marshal(multiSigAddress))
	// assertNil(err)
	// mustPutSet(b, txid)
	res, err := client.Insert(UnlockConditionSpace, []interface{}{txid.String(), siabin.Marshal(multiSigAddress)})
	if err == persist.ErrDuplicateKey {
		return
	}
	assertNil(res, err)
}
func dbRemoveWalletAddressToMultiSigAddressMapping(client *persist.TarantoolClient, walletAddress, multiSigAddress types.UnlockHash, txid types.TransactionID) {
	if build.DEBUG {
		if walletAddress.Type != types.UnlockTypePubKey {
			panic(fmt.Sprintf("wallet address has wrong type: %d", walletAddress.Type))
		}
		if multiSigAddress.Type != types.UnlockTypeMultiSig {
			panic(fmt.Sprintf("multisig address has wrong type: %d", multiSigAddress.Type))
		}
	}
	// space := client.Schema.Spaces[UnlockConditionSpace]

	// mb := tx.Bucket(bucketWalletAddressToMultiSigAddressMapping)
	// wa := siabin.Marshal(walletAddress)
	// wb := mb.Bucket(wa)
	// msa := siabin.Marshal(multiSigAddress)
	// msb := wb.Bucket(msa)
	// mustDelete(msb, txid)
	client.Delete(UnlockConditionSpace, UnlockConditionIDIndex, txid.String())
	// if bucketIsEmpty(msb) {
	// 	wb.DeleteBucket(msa)
	// 	if bucketIsEmpty(wb) {
	// 		mb.DeleteBucket(wa)
	// 	}
	// }
}
