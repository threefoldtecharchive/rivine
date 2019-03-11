package consensus

// database_test.go contains a bunch of legacy functions to preserve
// compatibility with the test suite.

import (
	bolt "github.com/rivine/bbolt"
	"github.com/threefoldtech/rivine/build"
	"github.com/threefoldtech/rivine/pkg/encoding/siabin"
	"github.com/threefoldtech/rivine/types"
)

// dbBlockHeight is a convenience function allowing blockHeight to be called
// without a bolt.Tx.
func (cs *ConsensusSet) dbBlockHeight() (bh types.BlockHeight) {
	dbErr := cs.db.View(func(tx *bolt.Tx) error {
		bh = blockHeight(tx)
		return nil
	})
	if dbErr != nil {
		build.Severe(dbErr)
	}
	return bh
}

// dbCurrentBlockID is a convenience function allowing currentBlockID to be
// called without a bolt.Tx.
func (cs *ConsensusSet) dbCurrentBlockID() (id types.BlockID) {
	dbErr := cs.db.View(func(tx *bolt.Tx) error {
		id = currentBlockID(tx)
		return nil
	})
	if dbErr != nil {
		build.Severe(dbErr)
	}
	return id
}

// dbCurrentProcessedBlock is a convenience function allowing
// currentProcessedBlock to be called without a bolt.Tx.
func (cs *ConsensusSet) dbCurrentProcessedBlock() (pb *processedBlock) {
	dbErr := cs.db.View(func(tx *bolt.Tx) error {
		pb = currentProcessedBlock(tx)
		return nil
	})
	if dbErr != nil {
		build.Severe(dbErr)
	}
	return pb
}

// dbGetPath is a convenience function allowing getPath to be called without a
// bolt.Tx.
func (cs *ConsensusSet) dbGetPath(bh types.BlockHeight) (id types.BlockID, err error) {
	dbErr := cs.db.View(func(tx *bolt.Tx) error {
		id, err = getPath(tx, bh)
		return nil
	})
	if dbErr != nil {
		build.Severe(dbErr)
	}
	return id, err
}

// dbPushPath is a convenience function allowing pushPath to be called without a
// bolt.Tx.
func (cs *ConsensusSet) dbPushPath(bid types.BlockID) {
	dbErr := cs.db.Update(func(tx *bolt.Tx) error {
		pushPath(tx, bid)
		return nil
	})
	if dbErr != nil {
		build.Severe(dbErr)
	}
}

// dbGetBlockMap is a convenience function allowing getBlockMap to be called
// without a bolt.Tx.
func (cs *ConsensusSet) dbGetBlockMap(id types.BlockID) (pb *processedBlock, err error) {
	dbErr := cs.db.View(func(tx *bolt.Tx) error {
		pb, err = getBlockMap(tx, id)
		return nil
	})
	if dbErr != nil {
		build.Severe(dbErr)
	}
	return pb, err
}

// dbGetCoinOutput is a convenience function allowing getCoinOutput to be
// called without a bolt.Tx.
func (cs *ConsensusSet) dbGetCoinOutput(id types.CoinOutputID) (sco types.CoinOutput, err error) {
	dbErr := cs.db.View(func(tx *bolt.Tx) error {
		sco, err = getCoinOutput(tx, id)
		return nil
	})
	if dbErr != nil {
		build.Severe(dbErr)
	}
	return sco, err
}

// getArbCoinOutput is a convenience function fetching a single random
// coin output from the database.
func (cs *ConsensusSet) getArbCoinOutput() (scoid types.CoinOutputID, sco types.CoinOutput, err error) {
	dbErr := cs.db.View(func(tx *bolt.Tx) error {
		cursor := tx.Bucket(CoinOutputs).Cursor()
		scoidBytes, scoBytes := cursor.First()
		copy(scoid[:], scoidBytes)
		return siabin.Unmarshal(scoBytes, &sco)
	})
	if dbErr != nil {
		build.Severe(dbErr)
	}
	if err != nil {
		return types.CoinOutputID{}, types.CoinOutput{}, err
	}
	return scoid, sco, nil
}

// dbGetBlockStakeOutput is a convenience function allowing getSiafundOutput to be
// called without a bolt.Tx.
func (cs *ConsensusSet) dbGetBlockStakeOutput(id types.BlockStakeOutputID) (sfo types.BlockStakeOutput, err error) {
	dbErr := cs.db.View(func(tx *bolt.Tx) error {
		sfo, err = getBlockStakeOutput(tx, id)
		return nil
	})
	if dbErr != nil {
		build.Severe(dbErr)
	}
	return sfo, err
}

// dbAddBlockStakeOutput is a convenience function allowing addBlockStakeOutput to be
// called without a bolt.Tx.
func (cs *ConsensusSet) dbAddBlockStakeOutput(id types.BlockStakeOutputID, sfo types.BlockStakeOutput) {
	dbErr := cs.db.Update(func(tx *bolt.Tx) error {
		addBlockStakeOutput(tx, id, sfo)
		return nil
	})
	if dbErr != nil {
		build.Severe(dbErr)
	}
}
