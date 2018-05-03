package explorer

import (
	"os"
	"path/filepath"

	"github.com/rivine/rivine/encoding"
	"github.com/rivine/rivine/modules"
	"github.com/rivine/rivine/persist"
	"github.com/rivine/rivine/types"

	"github.com/rivine/bbolt"
)

var explorerMetadata = persist.Metadata{
	Header:  "Sia Explorer",
	Version: "1.0.5",
}

// initPersist initializes the persistent structures of the explorer module.
func (e *Explorer) initPersist() error {
	// Make the persist directory
	err := os.MkdirAll(e.persistDir, 0700)
	if err != nil {
		return err
	}

	// Open the database
	dbFilPath := filepath.Join(e.persistDir, "explorer.db")
	db, err := persist.OpenDatabase(explorerMetadata, dbFilPath)
	if err != nil {
		if err != persist.ErrBadVersion {
			return err
		}
		db, err = convertLegacyDatabase(dbFilPath)
		if err != nil {
			return err
		}
	}
	e.db = db

	// Initialize the database
	err = e.db.Update(func(tx *bolt.Tx) error {
		buckets := [][]byte{
			bucketBlockFacts,
			bucketBlockIDs,
			bucketBlocksDifficulty,
			bucketBlockTargets,
			bucketInternal,
			bucketCoinOutputIDs,
			bucketCoinOutputs,
			bucketBlockStakeOutputIDs,
			bucketBlockStakeOutputs,
			bucketTransactionIDs,
			bucketUnlockHashes,
		}
		for _, b := range buckets {
			_, err := tx.CreateBucketIfNotExists(b)
			if err != nil {
				return err
			}
		}

		// set default values for the bucketInternal
		internalDefaults := []struct {
			key, val []byte
		}{
			{internalBlockHeight, encoding.Marshal(types.BlockHeight(0))},
			{internalRecentChange, encoding.Marshal(modules.ConsensusChangeID{})},
		}
		b := tx.Bucket(bucketInternal)
		for _, d := range internalDefaults {
			if b.Get(d.key) != nil {
				continue
			}
			err := b.Put(d.key, d.val)
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}
