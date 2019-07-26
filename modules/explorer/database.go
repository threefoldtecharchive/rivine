package explorer

import (
	"errors"
	"fmt"

	bolt "github.com/rivine/bbolt"
	"github.com/threefoldtech/rivine/pkg/encoding/siabin"
	"github.com/threefoldtech/rivine/types"
)

var (

	// database buckets
	bucketBlockFacts       = []byte("BlockFacts")
	bucketBlockIDs         = []byte("BlockIDs")
	bucketBlocksDifficulty = []byte("BlocksDifficulty")
	bucketBlockTargets     = []byte("BlockTargets")
	// bucketInternal is used to store values internal to the explorer
	bucketInternal            = []byte("Internal")
	bucketCoinOutputIDs       = []byte("CoinOutputIDs")
	bucketCoinOutputs         = []byte("CoinOutputs")
	bucketBlockStakeOutputIDs = []byte("BlockStakeOutputIDs")
	bucketBlockStakeOutputs   = []byte("BlockStakeOutputs")
	bucketTransactionIDs      = []byte("TransactionIDs")
	bucketUnlockHashes        = []byte("UnlockHashes")
	// used to map (single-signature) wallet addresses to all the
	// multisig addresses they are part of
	bucketWalletAddressToMultiSigAddressMapping = []byte("WalletAddressToMultiSigAddressMapping")

	errNotExist = errors.New("entry does not exist")

	// keys for bucketInternal
	internalBlockHeight  = []byte("BlockHeight")
	internalRecentChange = []byte("RecentChange")
)

// These functions all return a 'func(*bolt.Tx) error', which, allows them to
// be called concisely with the db.View and db.Update functions, e.g.:
//
//    var height types.BlockHeight
//    db.View(dbGetAndDecode(bucketBlockIDs, id, &height))
//
// Instead of:
//
//   var height types.BlockHeight
//   db.View(func(tx *bolt.Tx) error {
//       idBytes, err := siabin.Marshal(id)
//       if err != nil { return err }
//       bytes := tx.Bucket(bucketBlockIDs).Get()
//       return siabin.Unmarshal(bytes, &height)
//   })

// dbGetAndDecode returns a 'func(*bolt.Tx) error' that retrieves and decodes
// a value from the specified bucket. If the value does not exist,
// dbGetAndDecode returns errNotExist.
func dbGetAndDecode(bucket []byte, key, val interface{}) func(*bolt.Tx) error {
	return func(tx *bolt.Tx) error {
		keyBytes, err := siabin.Marshal(key)
		if err != nil {
			return fmt.Errorf("failed to (siabin) marshal key bytes: %v", err)
		}
		valBytes := tx.Bucket(bucket).Get(keyBytes)
		if valBytes == nil {
			return errNotExist
		}
		return siabin.Unmarshal(valBytes, val)
	}
}

// dbGetTransactionIDSet returns a 'func(*bolt.Tx) error' that decodes a
// bucket of transaction IDs into a slice. If the bucket is nil,
// dbGetTransactionIDSet returns errNotExist.
func dbGetTransactionIDSet(bucket []byte, key interface{}, ids *[]types.TransactionID) func(*bolt.Tx) error {
	return func(tx *bolt.Tx) error {
		keyBytes, err := siabin.Marshal(key)
		if err != nil {
			return fmt.Errorf("failed to (siabin) marshal key bytes: %v", err)
		}
		b := tx.Bucket(bucket).Bucket(keyBytes)
		if b == nil {
			return errNotExist
		}
		// decode into a local slice
		var txids []types.TransactionID
		err = b.ForEach(func(txid, _ []byte) error {
			var id types.TransactionID
			err := siabin.Unmarshal(txid, &id)
			if err != nil {
				return err
			}
			txids = append(txids, id)
			return nil
		})
		if err != nil {
			return err
		}
		// set pointer
		*ids = txids
		return nil
	}
}

// dbGetBlockFacts returns a 'func(*bolt.Tx) error' that decodes
// the block facts for `height` into blockfacts
func (e *Explorer) dbGetBlockFacts(height types.BlockHeight, bf *blockFacts) func(*bolt.Tx) error {
	return func(tx *bolt.Tx) error {
		block, exists := e.cs.BlockAtHeight(height)
		if !exists {
			return errors.New("requested block facts for a block that does not exist")
		}
		return dbGetAndDecode(bucketBlockFacts, block.ID(), bf)(tx)
	}
}

// dbSetInternal sets the specified key of bucketInternal to the encoded value.
func dbSetInternal(key []byte, val interface{}) func(*bolt.Tx) error {
	return func(tx *bolt.Tx) error {
		valBytes, err := siabin.Marshal(val)
		if err != nil {
			return fmt.Errorf("failed to (siabin) marshal value: %v", err)
		}
		return tx.Bucket(bucketInternal).Put(key, valBytes)
	}
}

// dbGetInternal decodes the specified key of bucketInternal into the supplied pointer.
func dbGetInternal(key []byte, val interface{}) func(*bolt.Tx) error {
	return func(tx *bolt.Tx) error {
		return siabin.Unmarshal(tx.Bucket(bucketInternal).Get(key), val)
	}
}
