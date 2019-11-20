package explorerdb

import (
	"fmt"

	"github.com/threefoldtech/rivine/build"
	"github.com/threefoldtech/rivine/modules"
	"github.com/threefoldtech/rivine/types"

	"github.com/threefoldtech/rivine/modules/explorergraphql/explorerdb/basedb"
	"github.com/threefoldtech/rivine/modules/explorergraphql/explorerdb/stormdb"
)

// TODO: integrate context.Context in each call

// TODO: we should not have to rely on CS data for getting the child target

// TODO: keep reference counter for public keys, and delete it in case the reference count is 0 (see TODO (4))

// TODO: handle also chain-specific stuff, such as chains that do not have block rewards

func NewStormDB(path string, bcInfo types.BlockchainInfo, chainCts types.ChainConstants, verbose bool) (*stormdb.StormDB, error) {
	return stormdb.New(path, bcInfo, chainCts, verbose)
}

func ApplyConsensusChangeWithChannel(db basedb.DB, cs modules.ConsensusSet, ch <-chan modules.ConsensusChange, chainCts *types.ChainConstants) error {
	const (
		minBlocksPerCommit = 1000
	)
	var blockCount = 0
	return db.ReadWriteTransaction(func(db basedb.RWTxn) error {
		var err error
		for csc := range ch {
			if len(csc.AppliedBlocks) == 0 {
				build.Critical("Explorer.ProcessConsensusChange called with a ConsensusChange that has no AppliedBlocks")
			}
			err = applyConsensusChangeForRWTxn(db, cs, csc, chainCts)
			if err != nil {
				return err
			}
			blockCount -= len(csc.RevertedBlocks)
			blockCount += len(csc.AppliedBlocks)
			if blockCount >= minBlocksPerCommit {
				err = db.Commit(false)
				if err != nil {
					return fmt.Errorf("failed to commit last (net) %d blocks: %v", blockCount, err)
				}
				blockCount = 0
			}
		}
		return nil
	})
}

func ApplyConsensusChange(db basedb.DB, cs modules.ConsensusSet, csc modules.ConsensusChange, chainCts *types.ChainConstants) error {
	return db.ReadWriteTransaction(func(db basedb.RWTxn) error {
		return applyConsensusChangeForRWTxn(db, cs, csc, chainCts)
	})
}

func applyConsensusChangeForRWTxn(db basedb.RWTxn, cs modules.ConsensusSet, csc modules.ConsensusChange, chainCts *types.ChainConstants) error {
	chainCtx, err := db.GetChainContext()
	if err != nil {
		return err
	}

	for _, revertedBlock := range csc.RevertedBlocks {
		// TODO: verify if this is correct, or if it should be done after
		chainCtx.Height--
		chainCtx.Timestamp = revertedBlock.Timestamp

		block := basedb.RivineBlockAsExplorerBlock(chainCtx.Height, revertedBlock)
		chainCtx.BlockID = block.ID

		outputs := make([]types.OutputID, 0, len(revertedBlock.MinerPayouts))
		for idx := range revertedBlock.MinerPayouts {
			outputs = append(outputs, block.Payouts[idx])
		}

		var inputs []types.OutputID
		transactions := make([]types.TransactionID, 0, len(revertedBlock.Transactions))
		for idx, txn := range revertedBlock.Transactions {
			// add txn
			transactions = append(transactions, block.Transactions[idx])
			// add inputs
			for _, input := range txn.CoinInputs {
				inputs = append(inputs, types.OutputID(input.ParentID))
			}
			for _, input := range txn.BlockStakeInputs {
				inputs = append(inputs, types.OutputID(input.ParentID))
			}
			// add outputs
			for cidx := range txn.CoinOutputs {
				outputs = append(outputs, types.OutputID(txn.CoinOutputID(uint64(cidx))))
			}
			for bsidx := range txn.BlockStakeOutputs {
				outputs = append(outputs, types.OutputID(txn.BlockStakeOutputID(uint64(bsidx))))
			}
		}

		err = db.RevertBlock(basedb.BlockRevertContext{
			ID:        block.ID,
			Height:    chainCtx.Height,
			Timestamp: chainCtx.Timestamp,
		}, transactions, outputs, inputs)
		if err != nil {
			return err
		}
	}

	for _, appliedBlock := range csc.AppliedBlocks {
		var target types.Target
		if chainCtx.Height > 0 {
			// TODO: find a better way than having to get this target from the consensusSet DB
			var ok bool
			target, ok = cs.ChildTarget(appliedBlock.ParentID)
			if !ok {
				return fmt.Errorf("failed to look up child target for parent block %s", appliedBlock.ParentID.String())
			}
		} else {
			target = chainCts.RootTarget()
		}
		blockFacts := basedb.BlockFactsConstants{
			Target:     target,
			Difficulty: target.Difficulty(chainCts.RootDepth),
		}

		block := basedb.RivineBlockAsExplorerBlock(chainCtx.Height, appliedBlock)

		outputs := make([]basedb.Output, 0, len(appliedBlock.MinerPayouts))
		for idx, mp := range appliedBlock.MinerPayouts {
			outputs = append(outputs, basedb.RivineMinerPayoutAsOutput(
				block.ID,
				types.CoinOutputID(block.Payouts[idx]),
				mp,
				// TODO: customize this per chain network (behaviour and constants)
				idx == 0,
				chainCtx.Height,
				chainCts.MaturityDelay,
			))
		}
		// TODO: customize this per chain network
		var feePayoutID types.OutputID
		if len(block.Payouts) > 1 {
			feePayoutID = block.Payouts[1]
		}

		inputs := make(map[types.OutputID]basedb.OutputSpenditureData)
		transactions := make([]basedb.Transaction, 0, len(appliedBlock.Transactions))
		for txidx, txn := range appliedBlock.Transactions {
			transaction := basedb.RivineTransactionAsTransaction(
				block.ID,
				block.Transactions[txidx],
				txn,
				feePayoutID,
			)
			transactions = append(transactions, transaction)
			// add inputs
			for _, input := range txn.CoinInputs {
				inputs[types.OutputID(input.ParentID)] = basedb.OutputSpenditureData{
					Fulfillment:              input.Fulfillment,
					FulfillmentTransactionID: block.Transactions[txidx],
				}
			}
			for _, input := range txn.BlockStakeInputs {
				inputs[types.OutputID(input.ParentID)] = basedb.OutputSpenditureData{
					Fulfillment:              input.Fulfillment,
					FulfillmentTransactionID: block.Transactions[txidx],
				}
			}
			// add outputs
			for coidx, output := range txn.CoinOutputs {
				outputs = append(outputs, basedb.RivineCoinOutputAsOutput(
					block.Transactions[txidx],
					types.CoinOutputID(transaction.CoinOutputs[coidx]),
					output,
				))
			}
			for bsidx, output := range txn.BlockStakeOutputs {
				outputs = append(outputs, basedb.RivineBlockStakeOutputAsOutput(
					block.Transactions[txidx],
					types.BlockStakeOutputID(transaction.BlockStakeOutputs[bsidx]),
					output,
				))
			}
		}

		err = db.ApplyBlock(block, blockFacts, transactions, outputs, inputs)
		if err != nil {
			return err
		}

		// TODO: verify if this is correct, or if it should be done before
		chainCtx.Height++
		chainCtx.Timestamp = appliedBlock.Timestamp
		chainCtx.BlockID = block.ID
	}

	chainCtx.ConsensusChangeID = csc.ID
	err = db.SetChainContext(chainCtx)
	return err
}
