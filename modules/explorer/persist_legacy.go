package explorer

import (
	"fmt"

	"github.com/threefoldtech/rivine/build"
	"github.com/threefoldtech/rivine/persist"
	"github.com/threefoldtech/rivine/pkg/encoding/siabin"
	"github.com/threefoldtech/rivine/types"

	bolt "github.com/rivine/bbolt"
)

func (e *Explorer) convertLegacyDatabase(filePath string) (db *persist.BoltDatabase, err error) {
	var legacyExplorerMetadata = persist.Metadata{
		Header:  "Sia Explorer",
		Version: "1.0.5",
	}
	db, err = persist.OpenDatabase(legacyExplorerMetadata, filePath)
	if err != nil {
		if err != persist.ErrBadVersion {
			return
		}
		db, err = convert052Database(filePath)
		return
	}

	err = db.Update(func(tx *bolt.Tx) error {
		// delete old outputID->condition mapping bucket
		tx.DeleteBucket([]byte("parentIDUnlockHashMapping")) // ignore errors though

		// create the bucketWalletAddressToMultiSigAddressMapping bucket
		_, err := tx.CreateBucket(bucketWalletAddressToMultiSigAddressMapping)
		if err != nil {
			return err
		}

		// fill the mapping
		return tx.Bucket(bucketUnlockHashes).ForEach(func(key, _ []byte) error {
			var uh types.UnlockHash
			err := siabin.Unmarshal(key, &uh)
			if err != nil {
				return fmt.Errorf("failed to unmarshal unlockhash from key in bucketUnlockHashes: %v", err)
			}
			if uh.Type != types.UnlockTypeMultiSig {
				return nil // continue, no need to migrate
			}
			return tx.Bucket(bucketUnlockHashes).Bucket(key).ForEach(func(key, _ []byte) error {
				// get transaction
				var txid types.TransactionID
				copy(txid[:], key[:])
				t, _, ok := e.cs.TransactionAtID(txid)
				if !ok {
					return fmt.Errorf("failed to get tx for id:  %x", txid)
				}

				// for all outputs, check if it is the given unlock hash,
				// and if so, get condition's unlock hashes to do the mapping
				for _, co := range t.CoinOutputs {
					if co.Condition.UnlockHash().Cmp(uh) != 0 || co.Condition.Condition == nil {
						continue
					}
					mapUnlockConditionMultiSigAddress(tx, uh, co.Condition.Condition, txid)
				}
				for _, bso := range t.BlockStakeOutputs {
					if bso.Condition.UnlockHash().Cmp(uh) != 0 || bso.Condition.Condition == nil {
						continue
					}
					mapUnlockConditionMultiSigAddress(tx, uh, bso.Condition.Condition, txid)
				}

				// for all inputs, get the parent output and do the same as we did with outputs
				for _, ci := range t.CoinInputs {
					var co types.CoinOutput
					err := dbGetAndDecode(bucketCoinOutputs, ci.ParentID, &co)(tx)
					if err != nil {
						return fmt.Errorf("failed to get co for parentID %x: %v", txid, err)
					}
					if co.Condition.UnlockHash().Cmp(uh) != 0 || co.Condition.Condition == nil {
						continue
					}
					mapUnlockConditionMultiSigAddress(tx, uh, co.Condition.Condition, txid)
				}
				for _, bsi := range t.BlockStakeInputs {
					var bso types.BlockStakeOutput
					err := dbGetAndDecode(bucketBlockStakeOutputs, bsi.ParentID, &bso)(tx)
					if err != nil {
						return fmt.Errorf("failed to get bso for parentID %x: %v", txid, err)
					}
					if bso.Condition.UnlockHash().Cmp(uh) != 0 || bso.Condition.Condition == nil {
						continue
					}
					mapUnlockConditionMultiSigAddress(tx, uh, bso.Condition.Condition, txid)
				}

				return nil
			})
		})
	})
	// If a bucket exist error is thrown its because we used to have a software version which ran this migration but did not update
	// the database metadata
	if err == nil || err == bolt.ErrBucketExists {
		// set the new metadata, and save it,
		// such that next time we have the new version stored
		db.Header, db.Version = explorerMetadata.Header, explorerMetadata.Version
		err = db.SaveMetadata()
	}
	if err != nil {
		err := db.Close()
		if err != nil {
			build.Severe(err)
		}
	}
	return
}

// convert052Database converts a 0.5.2 explorer database,
// to a database of the current version as defined by explorerMetadata.
// It keeps the database open and returns it for further usage.
func convert052Database(filePath string) (db *persist.BoltDatabase, err error) {
	var legacyExplorerMetadata = persist.Metadata{
		Header:  "Sia Explorer",
		Version: "0.5.2",
	}
	db, err = persist.OpenDatabase(legacyExplorerMetadata, filePath)
	if err != nil {
		return
	}

	err = db.Update(func(tx *bolt.Tx) error {
		if bucket := tx.Bucket(bucketCoinOutputs); bucket != nil {
			if err := updateLegacyCoinOutputBucket(bucket); err != nil {
				return err
			}
		}
		if bucket := tx.Bucket(bucketBlockStakeOutputs); bucket != nil {
			return updateLegacyBlockstakeOutputBucket(bucket)
		}
		return nil
	})
	if err == nil {
		// set the new metadata, and save it,
		// such that next time we have the new version stored
		db.Header, db.Version = explorerMetadata.Header, explorerMetadata.Version
		err = db.SaveMetadata()
	}
	if err != nil {
		err := db.Close()
		if err != nil {
			build.Severe(err)
		}
	}
	return
}

func updateLegacyCoinOutputBucket(bucket *bolt.Bucket) error {
	var (
		err    error
		cursor = bucket.Cursor()
	)
	for k, v := cursor.First(); len(k) != 0; k, v = cursor.Next() {
		// try to decode the legacy format
		var out legacyOutput
		err = siabin.Unmarshal(v, &out)
		if err != nil {
			// ensure it is in the new format already
			var co types.CoinOutput
			err = siabin.Unmarshal(v, &co)
			if err != nil {
				return err
			}
		}
		// it's in the legacy format, as expected, we overwrite it using the new format
		cob, err := siabin.Marshal(types.CoinOutput{
			Value: out.Value,
			Condition: types.UnlockConditionProxy{
				Condition: types.NewUnlockHashCondition(out.UnlockHash),
			},
		})
		if err != nil {
			return fmt.Errorf("failed to (siabin) marshal coin output: %v", err)
		}
		err = bucket.Put(k, cob)
		if err != nil {
			return err
		}
	}
	return nil
}

func updateLegacyBlockstakeOutputBucket(bucket *bolt.Bucket) error {
	var (
		err    error
		cursor = bucket.Cursor()
	)
	for k, v := cursor.First(); len(k) != 0; k, v = cursor.Next() {
		// try to decode the legacy format
		var out legacyOutput
		err = siabin.Unmarshal(v, &out)
		if err != nil {
			// ensure it is in the new format already
			var bso types.BlockStakeOutput
			err = siabin.Unmarshal(v, &bso)
			if err != nil {
				return err
			}
		}
		// it's in the legacy format, as expected, we overwrite it using the new format
		bsoBytes, err := siabin.Marshal(types.BlockStakeOutput{
			Value: out.Value,
			Condition: types.UnlockConditionProxy{
				Condition: types.NewUnlockHashCondition(out.UnlockHash),
			},
		})
		if err != nil {
			return fmt.Errorf("failed to (siabin) marshal block stake output: %v", err)
		}
		err = bucket.Put(k, bsoBytes)
		if err != nil {
			return err
		}
	}
	return nil
}

type legacyOutput struct {
	Value      types.Currency
	UnlockHash types.UnlockHash
}
