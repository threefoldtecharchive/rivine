package explorer

import (
	"github.com/rivine/rivine/build"
	"github.com/rivine/rivine/encoding"
	"github.com/rivine/rivine/persist"
	"github.com/rivine/rivine/types"

	"github.com/rivine/bbolt"
)

// convertLegacyDatabase converts a 0.5.2 explorer database,
// to a database of the current version as defined by explorerMetadata.
// It keeps the database open and returns it for further usage.
func convertLegacyDatabase(filePath string) (db *persist.BoltDatabase, err error) {
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
		// TODO: check if we need to do something with bucketUnlockHashes
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
		if build.DEBUG && err != nil {
			panic(err)
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
		err = encoding.Unmarshal(v, &out)
		if err != nil {
			// ensure it is in the new format already
			var co types.CoinOutput
			err = encoding.Unmarshal(v, &co)
			if err != nil {
				return err
			}
		}
		// it's in the legacy format, as expected, we overwrite it using the new format
		err = bucket.Put(k, encoding.Marshal(types.CoinOutput{
			Value: out.Value,
			Condition: types.UnlockConditionProxy{
				Condition: types.NewUnlockHashCondition(out.UnlockHash),
			},
		}))
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
		err = encoding.Unmarshal(v, &out)
		if err != nil {
			// ensure it is in the new format already
			var bso types.BlockStakeOutput
			err = encoding.Unmarshal(v, &bso)
			if err != nil {
				return err
			}
		}
		// it's in the legacy format, as expected, we overwrite it using the new format
		err = bucket.Put(k, encoding.Marshal(types.BlockStakeOutput{
			Value: out.Value,
			Condition: types.UnlockConditionProxy{
				Condition: types.NewUnlockHashCondition(out.UnlockHash),
			},
		}))
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
