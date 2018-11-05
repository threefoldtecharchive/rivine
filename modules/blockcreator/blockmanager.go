package blockcreator

import (
	"github.com/threefoldtech/rivine/modules"
	"github.com/threefoldtech/rivine/types"
)

// submitBlock accepts a block.
func (m *BlockCreator) submitBlock(blockToSubmit types.Block) error {
	if err := m.tg.Add(); err != nil {
		return err
	}
	defer m.tg.Done()

	// Give the block to the consensus set.
	err := m.cs.AcceptBlock(blockToSubmit)

	if err == modules.ErrNonExtendingBlock {
		m.log.Println("Created a stale block - block appears valid but does not extend the blockchain")
		return err
	}
	if err == modules.ErrBlockUnsolved {
		m.log.Println("Created an unsolved block - submission appears to be incorrect")
		return err
	}
	if err == modules.ErrBlockKnown {
		m.log.Println("Created a block that is already known in the consensusset")
		return err
	}
	if err != nil {
		m.tpool.PurgeTransactionPool()
		m.log.Critical("ERROR: an invalid block was submitted:", err)
		return err
	}
	return nil
}
