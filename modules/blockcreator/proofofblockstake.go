package blockcreator

import (
	"encoding/json"
	"fmt"
	"time"

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
		unspentBlockStakeOutputs := bc.wallet.UnspentBlockStakeOutputs()
		for outputID, ubso := range unspentBlockStakeOutputs {
			ubsobytes, _ := json.Marshal(ubso)
			fmt.Println("ROB-solving block for", outputID, string(ubsobytes))
			// TODO: Take a copy here instead of in submitBlock?
			// Solve the block.
			// Try to solve a block for blocktimes of the next 10 seconds
			now := time.Now().Unix()
			for blocktime := now; blocktime < now+10; blocktime++ {

				b := bc.solveBlock(outputID, blocktime)
				if b != nil {
					err := bc.submitBlock(*b)
					if err != nil {
						bc.log.Println("ERROR: An error occurred while submitting a solved block:", err)
					}
				}
			}
		}

		//sleep a while before recalculating
		time.Sleep(8 * time.Second)
	}
}

func (bc *BlockCreator) solveBlock(outputID types.BlockStakeOutputID, blocktime int64) (b *types.Block) {
	return
}
