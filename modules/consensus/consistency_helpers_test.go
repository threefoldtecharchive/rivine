package consensus

import (
	bolt "github.com/rivine/bbolt"
	"github.com/threefoldtech/rivine/build"
	"github.com/threefoldtech/rivine/crypto"
)

// dbConsensusChecksum is a convenience function to call consensusChecksum
// without a bolt.Tx.
func (cs *ConsensusSet) dbConsensusChecksum() (checksum crypto.Hash) {
	err := cs.db.Update(func(tx *bolt.Tx) error {
		checksum = consensusChecksum(tx)
		return nil
	})
	if err != nil {
		build.Severe(err)
	}
	return checksum
}
