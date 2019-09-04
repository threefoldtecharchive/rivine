package consensus

import (
	"bytes"
	"encoding/hex"
	"fmt"

	bolt "github.com/rivine/bbolt"
	"github.com/threefoldtech/rivine/build"
	"github.com/threefoldtech/rivine/crypto"
	"github.com/threefoldtech/rivine/modules"
	"github.com/threefoldtech/rivine/persist"
	"github.com/threefoldtech/rivine/pkg/encoding/siabin"
	"github.com/threefoldtech/rivine/types"
)

func convertLegacyDatabase(filepath string, log *persist.Logger) (*persist.BoltDatabase, error) {
	return convertLegacyOneZeroFiveDatabase(filepath, log)
}

// convertLegacyOneZeroFiveDatabase converts a 1.0.5 consensus database,
// to a database of the current version as defined by dbMetadata.
// It keeps the database open and returns it for further usage.
func convertLegacyOneZeroFiveDatabase(filepath string, log *persist.Logger) (db *persist.BoltDatabase, err error) {
	var legacyDBMetadata = persist.Metadata{
		Header:  "Consensus Set Database",
		Version: "1.0.5",
	}
	db, err = persist.OpenDatabase(legacyDBMetadata, filepath)
	if err != nil {
		if err == persist.ErrBadVersion {
			db, err = convertLegacyZeroFiveZeroDatabase(filepath, log, legacyDBMetadata)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucket(BucketPlugins)
		return err
	})
	if err == nil {
		// set the new metadata, and save it,
		// such that next time we have the new version stored
		db.Header, db.Version = dbMetadata.Header, dbMetadata.Version
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

// convertLegacyDatabase converts a 0.5.0 consensus database,
// to a database of the current version as defined by dbMetadata.
// It keeps the database open and returns it for further usage.
func convertLegacyZeroFiveZeroDatabase(filePath string, log *persist.Logger, desiredMetadata persist.Metadata) (db *persist.BoltDatabase, err error) {
	var legacyDBMetadata = persist.Metadata{
		Header:  "Consensus Set Database",
		Version: "0.5.0",
	}
	db, err = persist.OpenDatabase(legacyDBMetadata, filePath)
	if err != nil {
		return nil, err
	}
	err = db.Update(func(tx *bolt.Tx) error {
		if bucket := tx.Bucket(BlockMap); bucket != nil {
			log.Printf("upgrading bucket %s format from 0.5.0 to %v...\n", string(BlockMap), dbMetadata.Version)
			if err := updateLegacyBlockMapBucket(bucket, log); err != nil {
				return err
			}
		}
		if bucket := tx.Bucket(CoinOutputs); bucket != nil {
			log.Printf("upgrading bucket %s format from 0.5.0 to %v...\n", string(CoinOutputs), dbMetadata.Version)
			if err := updateLegacyCoinOutputBucket(bucket, log); err != nil {
				return err
			}
		}
		if bucket := tx.Bucket(BlockStakeOutputs); bucket != nil {
			log.Printf("upgrading bucket %s format from 0.5.0 to %v...\n", string(BlockStakeOutputs), dbMetadata.Version)
			if err := updateLegacyBlockstakeOutputBucket(bucket, log); err != nil {
				return err
			}
		}
		return tx.ForEach(func(name []byte, bucket *bolt.Bucket) error {
			if !bytes.HasPrefix(name, prefixDCO) {
				return nil
			}
			log.Printf("upgrading legacy DCO bucket 0x%s format from 0.5.0 to %v...\n", hex.EncodeToString(name), dbMetadata.Version)
			return updateLegacyCoinOutputBucket(bucket, log)
		})
	})
	if err == nil {
		// set the new metadata, and save it,
		// such that next time we have the new version stored
		db.Header, db.Version = desiredMetadata.Header, desiredMetadata.Version
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

func updateLegacyBlockMapBucket(bucket *bolt.Bucket, log *persist.Logger) error {
	var (
		err    error
		cursor = bucket.Cursor()
	)
	for k, v := cursor.First(); len(k) != 0; k, v = cursor.Next() {
		// try to decode the legacy format
		var legacyBlock legacyProcessedBlock
		err = siabin.Unmarshal(v, &legacyBlock)
		if err != nil {
			// ensure it is in the new format already
			var block processedBlock
			err = siabin.Unmarshal(v, &block)
			if err != nil {
				return err
			}
		}
		log.Printf("overwriting legacy block #%d with binary ID 0x%s created at %d to format as known since %v...\n",
			legacyBlock.Height, hex.EncodeToString(k), legacyBlock.Block.Timestamp, dbMetadata.Version)
		// it's in the legacy format, as expected, we overwrite it using the new format
		err = legacyBlock.storeAsNewFormat(bucket, k)
		if err != nil {
			return err
		}
	}
	return nil
}

func updateLegacyCoinOutputBucket(bucket *bolt.Bucket, log *persist.Logger) error {
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
		log.Printf("overwriting legacy coin output (%s, %s) with binary ID 0x%s to format as known since %v...\n",
			out.UnlockHash.String(), out.Value.String(), hex.EncodeToString(k), dbMetadata.Version)
		// it's in the legacy format, as expected, we overwrite it using the new format
		coBytes, err := siabin.Marshal(types.CoinOutput{
			Value: out.Value,
			Condition: types.UnlockConditionProxy{
				Condition: types.NewUnlockHashCondition(out.UnlockHash),
			},
		})
		if err != nil {
			return fmt.Errorf("failed to (siabin) marshal coin output: %v", err)
		}
		err = bucket.Put(k, coBytes)
		if err != nil {
			return err
		}
	}
	return nil
}

func updateLegacyBlockstakeOutputBucket(bucket *bolt.Bucket, log *persist.Logger) error {
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
		log.Printf("overwriting legacy block stake output (%s, %s) with binary ID 0x%s to format as known since %v...\n",
			out.UnlockHash.String(), out.Value.String(), hex.EncodeToString(k), dbMetadata.Version)
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

// legacyProcessedBlock defines the legacy version of what used to be the processedBlock,
// as serialized in the consensus database
type (
	legacyProcessedBlock struct {
		Block       legacyBlock
		Height      types.BlockHeight
		Depth       types.Target
		ChildTarget types.Target

		DiffsGenerated         bool
		CoinOutputDiffs        []legacyCoinOutputDiff
		BlockStakeOutputDiffs  []legacyBlockStakeOutputDiff
		DelayedCoinOutputDiffs []legacyDelayedCoinOutputDiff
		TxIDDiffs              []modules.TransactionIDDiff

		ConsensusChecksum crypto.Hash
	}
	legacyBlock struct {
		ParentID     types.BlockID
		Timestamp    types.Timestamp
		POBSOutput   types.BlockStakeOutputIndexes
		MinerPayouts []types.MinerPayout
		Transactions []types.Transaction
	}
	legacyCoinOutputDiff struct {
		Direction  modules.DiffDirection
		ID         types.CoinOutputID
		CoinOutput legacyOutput
	}
	legacyBlockStakeOutputDiff struct {
		Direction        modules.DiffDirection
		ID               types.BlockStakeOutputID
		BlockStakeOutput legacyOutput
	}
	legacyDelayedCoinOutputDiff struct {
		Direction      modules.DiffDirection
		ID             types.CoinOutputID
		CoinOutput     legacyOutput
		MaturityHeight types.BlockHeight
	}
	legacyOutput struct {
		Value      types.Currency
		UnlockHash types.UnlockHash
	}
)

func (lpb *legacyProcessedBlock) storeAsNewFormat(bucket *bolt.Bucket, key []byte) error {
	block := processedBlock{
		Block: types.Block{
			ParentID:     lpb.Block.ParentID,
			Timestamp:    lpb.Block.Timestamp,
			POBSOutput:   lpb.Block.POBSOutput,
			MinerPayouts: lpb.Block.MinerPayouts,
			Transactions: lpb.Block.Transactions,
		},
		Height:      lpb.Height,
		Depth:       lpb.Depth,
		ChildTarget: lpb.ChildTarget,

		DiffsGenerated: lpb.DiffsGenerated,
		TxIDDiffs:      lpb.TxIDDiffs,

		ConsensusChecksum: lpb.ConsensusChecksum,
	}

	block.CoinOutputDiffs = make([]modules.CoinOutputDiff, len(lpb.CoinOutputDiffs))
	for i, od := range lpb.CoinOutputDiffs {
		block.CoinOutputDiffs[i] = modules.CoinOutputDiff{
			Direction: od.Direction,
			ID:        od.ID,
			CoinOutput: types.CoinOutput{
				Value: od.CoinOutput.Value,
				Condition: types.UnlockConditionProxy{
					Condition: types.NewUnlockHashCondition(od.CoinOutput.UnlockHash),
				},
			},
		}
	}

	block.BlockStakeOutputDiffs = make([]modules.BlockStakeOutputDiff, len(lpb.BlockStakeOutputDiffs))
	for i, od := range lpb.BlockStakeOutputDiffs {
		block.BlockStakeOutputDiffs[i] = modules.BlockStakeOutputDiff{
			Direction: od.Direction,
			ID:        od.ID,
			BlockStakeOutput: types.BlockStakeOutput{
				Value: od.BlockStakeOutput.Value,
				Condition: types.UnlockConditionProxy{
					Condition: types.NewUnlockHashCondition(od.BlockStakeOutput.UnlockHash),
				},
			},
		}
	}

	block.DelayedCoinOutputDiffs = make([]modules.DelayedCoinOutputDiff, len(lpb.DelayedCoinOutputDiffs))
	for i, od := range lpb.DelayedCoinOutputDiffs {
		block.DelayedCoinOutputDiffs[i] = modules.DelayedCoinOutputDiff{
			Direction: od.Direction,
			ID:        od.ID,
			CoinOutput: types.CoinOutput{
				Value: od.CoinOutput.Value,
				Condition: types.UnlockConditionProxy{
					Condition: types.NewUnlockHashCondition(od.CoinOutput.UnlockHash),
				},
			},
			MaturityHeight: od.MaturityHeight,
		}
	}
	blockBytes, err := siabin.Marshal(block)
	if err != nil {
		return fmt.Errorf("failed to (siabin) marshal block: %v", err)
	}
	return bucket.Put(key, blockBytes)
}
