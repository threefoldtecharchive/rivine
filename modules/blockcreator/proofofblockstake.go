package blockcreator

import (
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	"github.com/rivine/rivine/crypto"
	"github.com/rivine/rivine/types"
)

// SolveBlocks participates in the Proof Of Block Stake protocol by continously checking if
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

		// TODO: where to put the lock exactly
		// Try to solve a block for blocktimes of the next 10 seconds
		now := time.Now().Unix()
		b := bc.solveBlock(now, 10)
		if b != nil {
			bjson, _ := json.Marshal(b)
			fmt.Println("Solved block:", string(bjson))

			err := bc.submitBlock(*b)
			if err != nil {
				bc.log.Println("ERROR: An error occurred while submitting a solved block:", err)
			}
		}
		//sleep a while before recalculating
		time.Sleep(8 * time.Second)
	}
}

func (bc *BlockCreator) solveBlock(startTime int64, secondsInTheFuture int64) (b *types.Block) {
	//height := bc.persist.Height + 1
	//TODO: properly calculate stakemodifier
	stakemodifier := big.NewInt(0)
	//TODO: sliding difficulty
	difficulty := types.StartDifficulty
	unspentBlockStakeOutputs := bc.wallet.GetUnspentBlockStakeOutputs()
	for _, ubso := range unspentBlockStakeOutputs {
		// TODO: look up the blockheight and transaction index of the unspent block stake output
		for blocktime := startTime; blocktime < startTime+secondsInTheFuture; blocktime++ {
			pobshash := crypto.HashAll(stakemodifier, ubso.BlockHeight, ubso.TransactionIndex, ubso.OutputIndex, blocktime)
			pobshashvalue := big.NewInt(0).SetBytes(pobshash[:])
			if pobshashvalue.Div(pobshashvalue, ubso.Value.Big()).Cmp(difficulty) == -1 {
				blockToSubmit := types.Block{
					ParentID:  bc.unsolvedBlock.ParentID,
					Timestamp: types.Timestamp(blocktime),
				}
				// Block is going to be passed to external memory, but the memory pointed
				// to by the transactions slice is still being modified - needs to be
				// copied.
				txns := make([]types.Transaction, len(bc.unsolvedBlock.Transactions))
				copy(txns, bc.unsolvedBlock.Transactions)
				blockToSubmit.Transactions = txns

				// TODO: add blockcreator payouts
				// TODO: use the unspent block stake output and send it to ourselves
				return &blockToSubmit

			}

		}

	}
	return
}
