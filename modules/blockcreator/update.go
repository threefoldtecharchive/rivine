package blockcreator

import (
	"github.com/rivine/rivine/encoding"
	"github.com/rivine/rivine/modules"
	"github.com/rivine/rivine/types"
)

// ProcessConsensusChange will update the blockcreator's most recent block.
func (bc *BlockCreator) ProcessConsensusChange(cc modules.ConsensusChange) {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	// Update the block creator's understanding of the block height.
	for _, block := range cc.RevertedBlocks {
		// Only doing the block check if the height is above zero saves hashing
		// and saves a nontrivial amount of time during IBD.
		if bc.persist.Height > 0 || block.ID() != types.GenesisID {
			bc.persist.Height--
		} else if bc.persist.Height != 0 {
			// Sanity check - if the current block is the genesis block, the
			// blockcreator height should be set to zero.
			bc.log.Critical("BlockCreator has detected a genesis block, but the height of the block creator is set to ", bc.persist.Height)
			bc.persist.Height = 0
		}
	}
	for _, block := range cc.AppliedBlocks {
		// Only doing the block check if the height is above zero saves hashing
		// and saves a nontrivial amount of time during IBD.
		if bc.persist.Height > 0 || block.ID() != types.GenesisID {
			bc.persist.Height++
		} else if bc.persist.Height != 0 {
			// Sanity check - if the current block is the genesis block, the
			// block creator height should be set to zero.
			bc.log.Critical("BlockCreator has detected a genesis block, but the height of the block creator is set to ", bc.persist.Height)
			bc.persist.Height = 0
		}
	}

	// Update the unsolved block.
	bc.unsolvedBlock.ParentID = cc.AppliedBlocks[len(cc.AppliedBlocks)-1].ID()

	bc.persist.RecentChange = cc.ID
	bc.persist.ParentID = bc.unsolvedBlock.ParentID
	err := bc.save()
	if err != nil {
		bc.log.Println(err)
	}

}

// ReceiveUpdatedUnconfirmedTransactions will replace the current unconfirmed
// set of transactions with the input transactions.
func (bc *BlockCreator) ReceiveUpdatedUnconfirmedTransactions(unconfirmedTransactions []types.Transaction, _ modules.ConsensusChange) {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	// Edge case - if there are no transactions, set the block's transactions
	// to nil and return.
	if len(unconfirmedTransactions) == 0 {
		bc.unsolvedBlock.Transactions = nil
		return
	}

	// Add transactions to the block until the block size limit is reached.
	// Transactions are assumed to be in a sensible order.
	var i int
	remainingSize := int(types.BlockSizeLimit - 5e3) //check this 5k for the first extra
	for i = range unconfirmedTransactions {
		remainingSize -= len(encoding.Marshal(unconfirmedTransactions[i]))
		if remainingSize < 0 {
			break
		}
	}
	bc.unsolvedBlock.Transactions = unconfirmedTransactions[:i+1]
}
