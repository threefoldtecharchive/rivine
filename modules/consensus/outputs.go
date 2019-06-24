package consensus

import (
	bolt "github.com/rivine/bbolt"
	"github.com/threefoldtech/rivine/build"
	"github.com/threefoldtech/rivine/types"
)

// GetCoinOutput returns the unspent coin output for the given ID
func (cs *ConsensusSet) GetCoinOutput(id types.CoinOutputID) (co types.CoinOutput, err error) {
	dbErr := cs.db.View(func(tx *bolt.Tx) error {
		co, err = getCoinOutput(tx, id)
		return nil
	})
	if dbErr != nil {
		build.Critical(dbErr)
	}
	return co, err
}

// GetBlockStakeOutput returns the unspent blockstake output for the given ID
func (cs *ConsensusSet) GetBlockStakeOutput(id types.BlockStakeOutputID) (bso types.BlockStakeOutput, err error) {
	dbErr := cs.db.View(func(tx *bolt.Tx) error {
		bso, err = getBlockStakeOutput(tx, id)
		return nil
	})
	if dbErr != nil {
		build.Critical(dbErr)
	}
	return bso, err
}
