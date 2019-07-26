package blockcreator

import (
	"encoding/json"
	"math/big"
	"time"

	"github.com/threefoldtech/rivine/crypto"
	"github.com/threefoldtech/rivine/types"
)

// SolveBlocks participates in the Proof Of Block Stake protocol by continuously checking if
// unspent block stake outputs make a solution for the current unsolved block.
// If a match is found, the block is submitted to the consensus set.
// This function does not return until the blockcreator threadgroup is stopped.
func (bc *BlockCreator) SolveBlocks() {
	for {

		// Bail if 'Stop' has been called.
		select {
		case <-bc.tg.StopChan():
			return
		default:
		}

		// This is mainly here to avoid the creation of useless blocks during IBD and when a node comes back online
		// after some downtime
		if !bc.csSynced {
			if !bc.cs.Synced() {
				bc.log.Debugln("Consensus set is not synced, don't create blocks yet")
				time.Sleep(8 * time.Second)
				continue
			}
			bc.csSynced = true
		}

		// Try to solve a block for blocktimes of the next 10 seconds
		now := time.Now().Unix()
		bc.log.Debugln("[BC] Attempting to solve blocks")
		b := bc.solveBlock(uint64(now), 10)
		if b != nil {
			bjson, err := json.Marshal(b)
			if err != nil {
				bc.log.Println("Solved block but failed to JSON-marshal it for logging purposes:", err)
			} else {
				bc.log.Println("Solved block:", string(bjson))
			}

			err = bc.submitBlock(*b)
			if err != nil {
				bc.log.Println("ERROR: An error occurred while submitting a solved block:", err)
			}
		}
		//sleep a while before recalculating
		time.Sleep(8 * time.Second)
	}
}

func (bc *BlockCreator) solveBlock(startTime uint64, secondsInTheFuture uint64) (b *types.Block) {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	currentBlock := bc.cs.CurrentBlock()
	stakemodifier := bc.cs.CalculateStakeModifier(bc.persist.Height+1, currentBlock, bc.chainCts.StakeModifierDelay-1)
	cbid := bc.cs.CurrentBlock().ID()
	target, _ := bc.cs.ChildTarget(cbid)

	// Try all unspent blockstake outputs
	unspentBlockStakeOutputs, err := bc.wallet.GetUnspentBlockStakeOutputs()
	if err != nil {
		bc.log.Printf("failed to start solving block stakes: %v", err)
		return nil
	}
	for _, ubso := range unspentBlockStakeOutputs {
		BlockStakeAge := types.Timestamp(0)
		// Filter all unspent block stakes for aging. If the index of the unspent
		// block stake output is not the first transaction with the first index,
		// then block stake can only be used to solve blocks after its aging is
		// older than types.BlockStakeAging (more than 1 day)
		if ubso.Indexes.TransactionIndex != 0 || ubso.Indexes.OutputIndex != 0 {
			blockatheigh, _ := bc.cs.BlockAtHeight(ubso.Indexes.BlockHeight)
			BlockStakeAge = blockatheigh.Header().Timestamp + types.Timestamp(bc.chainCts.BlockStakeAging)
		}
		// Try all timestamps for this timerange
		for blocktime := startTime; blocktime < startTime+secondsInTheFuture; blocktime++ {
			if BlockStakeAge > types.Timestamp(blocktime) {
				continue
			}
			// Calculate the hash for the given unspent output and timestamp
			pobshash, err := crypto.HashAll(stakemodifier.Bytes(), ubso.Indexes.BlockHeight, ubso.Indexes.TransactionIndex, ubso.Indexes.OutputIndex, blocktime)
			if err != nil {
				bc.log.Printf("solveBlock failed due to failed crypto hash All %q: %v", ubso.BlockStakeOutputID.String(), err)
				return nil
			}
			// Check if it meets the difficulty
			pobshashvalue := big.NewInt(0).SetBytes(pobshash[:])
			pobshashvalue.Div(pobshashvalue, ubso.Value.Big()) //TODO rivine : this div can be mul on the other side of the compare

			if pobshashvalue.Cmp(target.Int()) == -1 {
				err := bc.RespentBlockStake(ubso)
				if err != nil {
					bc.log.Printf("failed to respond block stake %q: %v", ubso.BlockStakeOutputID.String(), err)
					return nil
				}

				bc.log.Debugln("\nSolved block with target", target)
				blockToSubmit := types.Block{
					ParentID:   bc.unsolvedBlock.ParentID,
					Timestamp:  types.Timestamp(blocktime),
					POBSOutput: ubso.Indexes,
				}

				// Block is going to be passed to external memory, but the memory pointed
				// to by the transactions slice is still being modified - needs to be
				// copied.
				txns := make([]types.Transaction, len(bc.unsolvedBlock.Transactions))
				copy(txns, bc.unsolvedBlock.Transactions)
				blockToSubmit.Transactions = txns
				// Collect the block creation fee
				if !bc.chainCts.BlockCreatorFee.IsZero() {
					blockToSubmit.MinerPayouts = append(blockToSubmit.MinerPayouts, types.MinerPayout{
						Value: bc.chainCts.BlockCreatorFee, UnlockHash: ubso.Condition.UnlockHash()})
				}
				// Collect the summed miner fee of all transactions
				collectedMinerFees := blockToSubmit.CalculateTotalMinerFees()
				if !collectedMinerFees.IsZero() {
					condition := bc.chainCts.TransactionFeeCondition
					if condition.ConditionType() == types.ConditionTypeNil {
						condition = ubso.Condition
					}
					blockToSubmit.MinerPayouts = append(blockToSubmit.MinerPayouts, types.MinerPayout{
						Value: collectedMinerFees, UnlockHash: condition.UnlockHash()})
				}
				// Add any transaction-specific Custom "Miner" payouts
				var mps []types.MinerPayout
				for _, txn := range blockToSubmit.Transactions {
					mps, err = txn.CustomMinerPayouts()
					if err != nil {
						// ignore here, not critical, but do log
						bc.log.Printf("error occured while fetching custom miner payouts from txn v%v: %v", txn.Version, err)
						continue
					}
					blockToSubmit.MinerPayouts = append(blockToSubmit.MinerPayouts, mps...)
				}

				return &blockToSubmit
			}
		}
	}
	return
}

// RespentBlockStake will spent the unspent block stake output which is needed
// for the POBS algorithm. The transaction created will be the first transaction
// in the block to avoid the BlockStakeAging for later use of this block stake.
func (bc *BlockCreator) RespentBlockStake(ubso types.UnspentBlockStakeOutput) error {
	// There is a special case: When the unspent block stake output is already
	// used in another transaction in this unsolved block, this extra transaction
	// is obsolete
	for _, ubstr := range bc.unsolvedBlock.Transactions {
		for _, ubstrinput := range ubstr.BlockStakeInputs {
			if ubstrinput.ParentID == ubso.BlockStakeOutputID {
				return nil
			}
		}
	}

	//otherwise the blockstake is not yet spent in this block, spent it now
	t := bc.wallet.StartTransaction()
	err := t.SpendBlockStake(ubso.BlockStakeOutputID) // link the input of this transaction
	// to the used BlockStake output
	if err != nil {
		return err
	}

	bso := types.BlockStakeOutput{
		Value:     ubso.Value,     //use the same amount of BlockStake
		Condition: ubso.Condition, //use the same condition.
	}
	t.AddBlockStakeOutput(bso)
	txnSet, err := t.Sign()
	if err != nil {
		return err
	}
	//add this transaction in front of the list of unsolved block transactions
	bc.unsolvedBlock.Transactions = append(txnSet, bc.unsolvedBlock.Transactions...)
	return nil
}
