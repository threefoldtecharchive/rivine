package blockcreator

import (
	"github.com/rivine/rivine/modules"
	"github.com/rivine/rivine/types"
)

// ProcessConsensusChange will update the blockcreator's most recent block.
func (b *BlockCreator) ProcessConsensusChange(cc modules.ConsensusChange) {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Update the block creator's understanding of the block height.
	for _, block := range cc.RevertedBlocks {
		// Only doing the block check if the height is above zero saves hashing
		// and saves a nontrivial amount of time during IBD.
		if b.persist.Height > 0 || block.ID() != types.GenesisID {
			b.persist.Height--
		} else if b.persist.Height != 0 {
			// Sanity check - if the current block is the genesis block, the
			// blockcreator height should be set to zero.
			b.log.Critical("BlockCreator has detected a genesis block, but the height of the block creator is set to ", b.persist.Height)
			b.persist.Height = 0
		}
	}
	for _, block := range cc.AppliedBlocks {
		// Only doing the block check if the height is above zero saves hashing
		// and saves a nontrivial amount of time during IBD.
		if b.persist.Height > 0 || block.ID() != types.GenesisID {
			b.persist.Height++
		} else if b.persist.Height != 0 {
			// Sanity check - if the current block is the genesis block, the
			// block creator height should be set to zero.
			b.log.Critical("BlockCreator has detected a genesis block, but the height of the block creator is set to ", b.persist.Height)
			b.persist.Height = 0
		}
	}

	b.persist.RecentChange = cc.ID
	err := b.save()
	if err != nil {
		b.log.Println(err)
	}

	//TODO: modify the block we are trying to create
}

// ReceiveUpdatedUnconfirmedTransactions will replace the current unconfirmed
// set of transactions with the input transactions.
func (b *BlockCreator) ReceiveUpdatedUnconfirmedTransactions(unconfirmedTransactions []types.Transaction, _ modules.ConsensusChange) {
	b.mu.Lock()
	defer b.mu.Unlock()

	//TODO: modify the block we ware trying to create
}
