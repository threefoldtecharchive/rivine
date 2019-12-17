package blockcreator

import (
	"github.com/threefoldtech/rivine/modules"
	"github.com/threefoldtech/rivine/types"
)

// submitBlock accepts a block.
func (b *BlockCreator) submitBlock(blockToSubmit types.Block) error {
	if err := b.tg.Add(); err != nil {
		return err
	}
	defer b.tg.Done()

	// Give the block to the consensus set.
	err := b.cs.AcceptBlock(blockToSubmit)

	if err == modules.ErrNonExtendingBlock {
		b.log.Println("Created a stale block - block appears valid but does not extend the blockchain")
		return err
	}
	if err == modules.ErrBlockUnsolved {
		b.log.Println("Created an unsolved block - submission appears to be incorrect")
		return err
	}
	if err == modules.ErrBlockKnown {
		b.log.Println("Created a block that is already known in the consensusset")
		return err
	}
	if err != nil {
		b.tpool.PurgeTransactionPool()
		b.log.Critical("ERROR: an invalid block was submitted:", err)
		return err
	}
	return nil
}
