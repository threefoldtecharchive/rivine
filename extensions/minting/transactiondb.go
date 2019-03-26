package minting

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"path"

	types "github.com/threefoldtech/rivine/types"

	bolt "github.com/rivine/bbolt"
	"github.com/threefoldtech/rivine/build"
	"github.com/threefoldtech/rivine/modules"
	persist "github.com/threefoldtech/rivine/persist"
	"github.com/threefoldtech/rivine/pkg/encoding/siabin"
	rivinesync "github.com/threefoldtech/rivine/sync"
)

// TransactionDB I/O constants
const (
	TransactionDBDir      = "transactiondb"
	TransactionDBFilename = TransactionDBDir + ".db"
)

// TODO:
//   - modify (function godoc) comments to take into account that now we also store/delete/manage 3bots, not just mintconditions

// TODO:
//   add an in-memory cache (layer), such that we do not constantly have to look up the same
//   names/records/identifiers/publickeys
//   (for mint condition I do not think it is required, as interaction with minting is minimal)

// internal bucket database keys used for the transactionDB
var (
	bucketInternal         = []byte("internal")
	bucketInternalKeyStats = []byte("stats") // stored as a single struct, see `transactionDBStats`

	// getBucketMintConditionPerHeightRangeKey is used to compute the keys
	// of the values in this bucket
	bucketMintConditions = []byte("mintconditions")
)

type (
	// TransactionDB extends Rivine's ConsensusSet module,
	// allowing us to track transactions (and specifically parts of it) that we care about,
	// and for which Rivine does not implement any logic.
	//
	// The initial motivation (and currently only use case) is to track MintConditions,
	// as to be able to know for any given block height what the active MintCondition is,
	// but other use cases can be supported in future updates should they appear.
	TransactionDB struct {
		// The DB's ThreadGroup tells tracked functions to shut down and
		// blocks until they have all exited before returning from Close.
		tg rivinesync.ThreadGroup

		db    *persist.BoltDatabase
		stats transactionDBStats

		subscriber *transactionDBCSSubscriber
	}

	// implements modules.ConsensusSetSubscriber,
	// such that the TransactionDB does not have to publicly implement
	// the ConsensusSetSubscriber interface, allowing us to "force"
	// the user to register to the consensus set using our provided
	// (*TransactionDB).SubscribeToConsensusSet method
	transactionDBCSSubscriber struct {
		txdb *TransactionDB
		cs   modules.ConsensusSet
	}
	transactionDBStats struct {
		ConsensusChangeID modules.ConsensusChangeID
		BlockHeight       types.BlockHeight
		ChainTime         types.Timestamp
		Synced            bool
	}
)

var (
	// ensure TransactionDB implements the MintConditionGetter interface
	_ MintConditionGetter = (*TransactionDB)(nil)
)

// NewTransactionDB creates a new TransactionDB, using the given file (path) to store the (single) persistent BoltDB file.
// A new db will be created if it doesn't exist yet, if it does exist it should be ensured that the given genesis mint condition
// equals the already stored genesis mint condition.
func NewTransactionDB(rootDir string, genesisMintCondition types.UnlockConditionProxy) (*TransactionDB, error) {
	persistDir := path.Join(rootDir, TransactionDBDir)
	// Create the directory if it doesn't exist.
	err := os.MkdirAll(persistDir, 0700)
	if err != nil {
		return nil, err
	}

	txdb := new(TransactionDB)
	err = txdb.openDB(path.Join(persistDir, TransactionDBFilename), genesisMintCondition)
	if err != nil {
		return nil, fmt.Errorf("failed to open the transaction DB: %v", err)
	}
	return txdb, nil
}

// Retrieves the Last ConsensusChangeID stored.
func (txdb *TransactionDB) GetLastConsensusChangeID() modules.ConsensusChangeID {
	return txdb.stats.ConsensusChangeID
}

// SubscribeToConsensusSet subscribes the TransactionDB to the given ConsensusSet,
// allowing it to stay in sync with the blockchain, and also making it automatically unsubscribe
// from the consensus set when the TransactionDB is closed (using (*TransactionDB).Close).
func (txdb *TransactionDB) SubscribeToConsensusSet(cs modules.ConsensusSet) error {
	if txdb.subscriber != nil {
		return errors.New("transactionDB is already subscribed to a consensus set")
	}

	subscriber := &transactionDBCSSubscriber{txdb: txdb, cs: cs}
	err := cs.ConsensusSetSubscribe(
		subscriber,
		txdb.stats.ConsensusChangeID,
		txdb.tg.StopChan(),
	)
	if err != nil {
		return fmt.Errorf("failed to subscribe to consensus set: %v", err)
	}
	txdb.subscriber = subscriber
	return nil
}

// GetActiveMintCondition implements types.MintConditionGetter.GetActiveMintCondition
func (txdb *TransactionDB) GetActiveMintCondition() (types.UnlockConditionProxy, error) {
	var b []byte
	err := txdb.db.View(func(tx *bolt.Tx) (err error) {
		mintConditionsBucket := tx.Bucket(bucketMintConditions)
		if mintConditionsBucket == nil {
			return errors.New("corrupt transaction DB: mint conditions bucket does not exist")
		}

		// return the last cursor
		cursor := mintConditionsBucket.Cursor()

		var k []byte
		k, b = cursor.Last()
		if len(k) == 0 {
			return errors.New("corrupt transaction DB: no matching mint condition could be found")
		}
		return nil
	})
	if err != nil {
		return types.UnlockConditionProxy{}, err
	}

	var mintCondition types.UnlockConditionProxy
	err = siabin.Unmarshal(b, &mintCondition)
	if err != nil {
		return types.UnlockConditionProxy{}, fmt.Errorf("corrupt transaction DB: failed to decode found mint condition: %v", err)
	}
	// mint condition found, return it
	return mintCondition, nil
}

// GetMintConditionAt implements types.MintConditionGetter.GetMintConditionAt
func (txdb *TransactionDB) GetMintConditionAt(height types.BlockHeight) (types.UnlockConditionProxy, error) {
	var b []byte
	err := txdb.db.View(func(tx *bolt.Tx) (err error) {
		mintConditionsBucket := tx.Bucket(bucketMintConditions)
		if mintConditionsBucket == nil {
			return errors.New("corrupt transaction DB: mint conditions bucket does not exist")
		}

		cursor := mintConditionsBucket.Cursor()

		var k []byte
		k, b = cursor.Seek(EncodeBlockheight(height))
		if len(k) == 0 {
			// could be that we're past the last key, let's try the last key first
			k, b = cursor.Last()
			if len(k) == 0 {
				return errors.New("corrupt transaction DB: no matching mint condition could be found")
			}
			return nil
		}
		foundHeight := DecodeBlockheight(k)
		if foundHeight <= height {
			return nil
		}
		k, b = cursor.Prev()
		if len(k) == 0 {
			return errors.New("corrupt transaction DB: no matching mint condition could be found")
		}
		return nil

	})
	if err != nil {
		return types.UnlockConditionProxy{}, err
	}

	var mintCondition types.UnlockConditionProxy
	err = siabin.Unmarshal(b, &mintCondition)
	if err != nil {
		return types.UnlockConditionProxy{}, fmt.Errorf("corrupt transaction DB: failed to decode found mint condition: %v", err)
	}
	// mint condition found, return it
	return mintCondition, nil
}

// Close the transaction DB,
// meaning the db will be unsubscribed from the consensus set,
// as well the threadgroup will be stopped and the internal bolt db will be closed.
func (txdb *TransactionDB) Close() error {
	if txdb.db == nil {
		return errors.New("transactionDB is already closed or was never created")
	}

	// unsubscribe from the consensus set, if subscribed at all
	if txdb.subscriber != nil {
		txdb.subscriber.unsubscribe()
		txdb.subscriber = nil
	}
	// stop thread group
	tgErr := txdb.tg.Stop()
	if tgErr != nil {
		tgErr = fmt.Errorf("failed to stop the threadgroup of TransactionDB: %v", tgErr)
	}
	// close database
	dbErr := txdb.db.Close()
	if dbErr != nil {
		dbErr = fmt.Errorf("failed to close the internal bolt db of TransactionDB: %v", dbErr)
	}
	txdb.db = nil

	return build.ComposeErrors(tgErr, dbErr)
}

// openDB loads the set database and populates it with the necessary buckets
func (txdb *TransactionDB) openDB(filename string, genesisMintCondition types.UnlockConditionProxy) (err error) {
	var (
		dbMetadata = persist.Metadata{
			Header:  "Rivine Transaction Database",
			Version: "1.1.2.1",
		}
	)

	txdb.db, err = persist.OpenDatabase(dbMetadata, filename)
	if err != nil {
		if err != persist.ErrBadVersion {
			return fmt.Errorf("error opening rivine transaction database: %v", err)
		}
		// try to migrate the DB
		err = txdb.migrateDB(filename)
		if err != nil {
			return err
		}
		// save the new metadata
		txdb.db.Metadata = dbMetadata
		err = txdb.db.SaveMetadata()
		if err != nil {
			return fmt.Errorf("error while saving the v1.1.2 metadata in the rivine transaction database: %v", err)
		}
	}
	return txdb.db.Update(func(tx *bolt.Tx) (err error) {
		if txdb.dbInitialized(tx) {
			// db is already created, get the stored stats
			internalBucket := tx.Bucket(bucketInternal)
			b := internalBucket.Get(bucketInternalKeyStats)
			if len(b) == 0 {
				return errors.New("structured stats value could not be found in existing transaction db")
			}
			err = siabin.Unmarshal(b, &txdb.stats)
			if err != nil {
				return fmt.Errorf("failed to unmarshal structured stats value from existing transaction db: %v", err)
			}

			// and ensure the genesis mint condition is the same as the given one
			mintConditionsBucket := tx.Bucket(bucketMintConditions)
			b = mintConditionsBucket.Get(EncodeBlockheight(0))
			if len(b) == 0 {
				return errors.New("genesis mint condition could not be found in existing transaction db")
			}
			var storedMintCondition types.UnlockConditionProxy
			err = siabin.Unmarshal(b, &storedMintCondition)
			if err != nil {
				return fmt.Errorf("failed to unmarshal genesis mint condition from existing transaction db: %v", err)
			}
			if !storedMintCondition.Equal(genesisMintCondition) {
				return errors.New("stored genesis mint condition is different from the given genesis mint condition")
			}

			return nil // nothing to do
		}

		// successfully create the DB
		err = txdb.createDB(tx, genesisMintCondition)
		if err != nil {
			return fmt.Errorf("failed to create transactionDB: %v", err)
		}
		return nil
	})
}

func (txdb *TransactionDB) migrateDB(filename string) error {
	// try to open the DB using the original version
	dbMetadata := persist.Metadata{
		Header:  "Rivine Transaction Database",
		Version: "1.1.0",
	}
	var err error
	txdb.db, err = persist.OpenDatabase(dbMetadata, filename)
	// if err == nil {
	// 	// migrate from a v1.1.0 DB
	// 	return txdb.db.Update(txdb.migrateV110DB)
	// }
	if err != persist.ErrBadVersion {
		return fmt.Errorf("error opening rivine transaction v1.1.0 database: %v", err)
	}

	// try to open the initial v1.2.0 DB (never released, but already out in field for dev purposes)
	dbMetadata.Version = "1.2.0"
	txdb.db, err = persist.OpenDatabase(dbMetadata, filename)
	// if err == nil {
	// 	// migrate from a v1.2.0 DB
	// 	return txdb.db.Update(txdb.migrateV120DB)
	// }
	if err == persist.ErrBadVersion {
		return fmt.Errorf("error opening rivine transaction database with unknown version: %v", err)
	}
	return fmt.Errorf("error opening rivine transaction v1.2.0 database: %v", err)
}

// func (txdb *TransactionDB) migrateV110DB(tx *bolt.Tx) error {
// 	// Enumerate and create the new database buckets.
// 	buckets := [][]byte{
// 		bucketBotRecords,
// 		bucketBotKeyToIDMapping,
// 		bucketBotNameToIDMapping,
// 		bucketBotRecordImplicitUpdates,
// 		bucketBotTransactions,
// 	}
// 	var err error
// 	for _, bucket := range buckets {
// 		_, err = tx.CreateBucket(bucket)
// 		if err != nil {
// 			return err
// 		}
// 	}
// 	// update the stats bucket
// 	var oldStats struct {
// 		ConsensusChangeID modules.ConsensusChangeID
// 		BlockHeight       types.BlockHeight
// 		Synced            bool
// 	}
// 	internalBucket := tx.Bucket(bucketInternal)
// 	b := internalBucket.Get(bucketInternalKeyStats)
// 	if len(b) == 0 {
// 		return errors.New("structured stats value could not be found in existing transaction db")
// 	}
// 	err = siabin.Unmarshal(b, &oldStats)
// 	if err != nil {
// 		return fmt.Errorf("failed to unmarshal structured stats value from existing transaction db: %v", err)
// 	}
// 	err = internalBucket.Put(bucketInternalKeyStats, siabin.Marshal(transactionDBStats{
// 		ConsensusChangeID: oldStats.ConsensusChangeID,
// 		BlockHeight:       oldStats.BlockHeight,
// 		ChainTime:         0, // will fix itself on the first block it receives
// 		Synced:            oldStats.Synced,
// 	}))
// 	if err != nil {
// 		return err
// 	}

// 	// Continue the migration process towards the newest version
// 	return txdb.migrateV120DB(tx)
// }

// func (txdb *TransactionDB) migrateV120DB(tx *bolt.Tx) error {
// 	// Enumerate and create the new database buckets.
// 	buckets := [][]byte{
// 		bucketERC20ToTFTAddresses,
// 		bucketTFTToERC20Addresses,
// 		bucketERC20TransactionIDs,
// 	}
// 	var err error
// 	for _, bucket := range buckets {
// 		_, err = tx.CreateBucket(bucket)
// 		if err != nil {
// 			return err
// 		}
// 	}

// 	// migration process is finished
// 	return nil
// }

// dbInitialized returns true if the database appears to be initialized, false
// if not. Checking for the existence of the siafund pool bucket is typically
// sufficient to determine whether the database has gone through the
// initialization process.
func (txdb *TransactionDB) dbInitialized(tx *bolt.Tx) bool {
	return tx.Bucket(bucketInternal) != nil
}

// createConsensusObjects initialzes the consensus portions of the database.
func (txdb *TransactionDB) createDB(tx *bolt.Tx, genesisMintCondition types.UnlockConditionProxy) (err error) {
	// Enumerate and create the database buckets.
	buckets := [][]byte{
		bucketInternal,
		bucketMintConditions,
	}
	for _, bucket := range buckets {
		_, err = tx.CreateBucket(bucket)
		if err != nil {
			return err
		}
	}

	// set the initial block height and initial consensus change iD
	txdb.stats.BlockHeight = 0
	txdb.stats.ConsensusChangeID = modules.ConsensusChangeBeginning
	internalBucket := tx.Bucket(bucketInternal)
	err = internalBucket.Put(bucketInternalKeyStats, siabin.Marshal(txdb.stats))
	if err != nil {
		return fmt.Errorf("failed to store transaction db (height=%d; changeID=%x) as a stat: %v",
			txdb.stats.BlockHeight, txdb.stats.ConsensusChangeID, err)
	}

	// store the genesis mint condition
	mintConditionsBucket := tx.Bucket(bucketMintConditions)
	err = mintConditionsBucket.Put(EncodeBlockheight(0), siabin.Marshal(genesisMintCondition))
	if err != nil {
		return fmt.Errorf("failed to store genesis mint condition: %v", err)
	}

	// all buckets created, and populated with initial content
	return nil
}

// ProcessConsensusChange implements modules.ConsensusSetSubscriber,
// calling txdb.processConsensusChange, so that the TransactionDB
// does not expose its interface implementation outside this package,
// given that we want the user to subscribe using the (*TransactionDB).SubscribeToConsensusSet method.
func (sub *transactionDBCSSubscriber) ProcessConsensusChange(css modules.ConsensusChange) {
	sub.txdb.processConsensusChange(css)
}

func (sub *transactionDBCSSubscriber) unsubscribe() {
	sub.cs.Unsubscribe(sub)
}

// processConsensusChange implements modules.ConsensusSetSubscriber,
// used to apply/revert transactions we care about in the internal persistent storage.
func (txdb *TransactionDB) processConsensusChange(css modules.ConsensusChange) {
	if err := txdb.tg.Add(); err != nil {
		// The TransactionDB should gracefully reject updates from the consensus set
		// that are sent after the wallet's Close method has closed the wallet's ThreadGroup.
		return
	}
	defer txdb.tg.Done()

	err := txdb.db.Update(func(tx *bolt.Tx) (err error) {
		// update reverted transactions in a block-defined order
		err = txdb.revertBlocks(tx, css.RevertedBlocks)
		if err != nil {
			return fmt.Errorf("failed to revert blocks: %v", err)
		}

		// update applied transactions in a block-defined order
		err = txdb.applyBlocks(tx, css.AppliedBlocks)
		if err != nil {
			return fmt.Errorf("failed to apply blocks: %v", err)
		}

		// update the consensus change ID and synced status
		txdb.stats.ConsensusChangeID, txdb.stats.Synced = css.ID, css.Synced

		// store stats
		internalBucket := tx.Bucket(bucketInternal)
		err = internalBucket.Put(bucketInternalKeyStats, siabin.Marshal(txdb.stats))
		if err != nil {
			return fmt.Errorf("failed to store transaction db (height=%d; changeID=%x; synced=%v) as a stat: %v",
				txdb.stats.BlockHeight, txdb.stats.ConsensusChangeID, txdb.stats.Synced, err)
		}

		return nil // all good
	})
	if err != nil {
		build.Critical("transactionDB update failed:", err)
	}
}

// revert all the given blocks using the given writable bolt Transaction,
// meaning the block height will be decreased per reverted block and
// all reverted mint conditions will be deleted as well
func (txdb *TransactionDB) revertBlocks(tx *bolt.Tx, blocks []types.Block) error {
	var (
		err error
		rtx *types.Transaction
	)

	mintConditionsBucket := tx.Bucket(bucketMintConditions)
	if mintConditionsBucket == nil {
		return errors.New("corrupt transaction DB: mint conditions bucket does not exist")
	}

	// collect all one-per-block mint conditions
	for _, block := range blocks {
		for i := range block.Transactions {
			rtx = &block.Transactions[i]
			if rtx.Version == types.TransactionVersionOne {
				continue // ignore most common Tx
			}

			// check the version and handle the ones we care about
			switch rtx.Version {
			case TransactionVersionMinterDefinition:
				err = txdb.revertMintConditionTx(tx, rtx)
			}
			if err != nil {
				return err
			}
		}

		// decrease block height (store later)
		txdb.stats.BlockHeight--
		// not super accurate, should be accurate enough and will fix itself when new blocks get applied
		txdb.stats.ChainTime = block.Timestamp
	}

	// all good
	return nil
}

// apply all the given blocks using the given writable bolt Transaction,
// meaning the block height will be increased per applied block and
// all applied mint conditions will be stored linked to their block height as well
//
// if a block contains multiple transactions with a mint condition,
// only the mint condition of the last transaction in the block's transaction list will be stored
func (txdb *TransactionDB) applyBlocks(tx *bolt.Tx, blocks []types.Block) error {
	var (
		err error
		rtx *types.Transaction
	)

	// collect all one-per-block mint conditions
	for _, block := range blocks {
		// increase block height (store later)
		txdb.stats.BlockHeight++
		txdb.stats.ChainTime = block.Timestamp

		for i := range block.Transactions {
			rtx = &block.Transactions[i]
			if rtx.Version == types.TransactionVersionOne {
				continue // ignore most common Tx
			}
			// check the version and handle the ones we care about
			switch rtx.Version {
			case TransactionVersionMinterDefinition:
				err = txdb.applyMintConditionTx(tx, rtx)
			}
			if err != nil {
				return err
			}
		}
	}

	// all good
	return nil
}

func (txdb *TransactionDB) applyMintConditionTx(tx *bolt.Tx, rtx *types.Transaction) error {
	mintConditionsBucket := tx.Bucket(bucketMintConditions)
	if mintConditionsBucket == nil {
		return errors.New("corrupt transaction DB: mint conditions bucket does not exist")
	}
	mdtx, err := MinterDefinitionTransactionFromTransaction(*rtx)
	if err != nil {
		return fmt.Errorf("unexpected error while unpacking the minter def. tx type: %v" + err.Error())
	}
	err = mintConditionsBucket.Put(EncodeBlockheight(txdb.stats.BlockHeight), siabin.Marshal(mdtx.MintCondition))
	if err != nil {
		return fmt.Errorf(
			"failed to put mint condition for block height %d: %v",
			txdb.stats.BlockHeight, err)
	}
	return nil
}

func (txdb *TransactionDB) revertMintConditionTx(tx *bolt.Tx, rtx *types.Transaction) error {
	mintConditionsBucket := tx.Bucket(bucketMintConditions)
	if mintConditionsBucket == nil {
		return errors.New("corrupt transaction DB: mint conditions bucket does not exist")
	}
	err := mintConditionsBucket.Delete(EncodeBlockheight(txdb.stats.BlockHeight))
	if err != nil {
		return fmt.Errorf(
			"failed to delete mint condition for block height %d: %v",
			txdb.stats.BlockHeight, err)
	}
	return nil
}

type transactionContext struct {
	BlockHeight  types.BlockHeight
	BlockTime    types.Timestamp
	TxSequenceID uint16
}

func (tctx transactionContext) TransactionShortID() sortableTransactionShortID {
	return newSortableTransactionShortID(tctx.BlockHeight, tctx.TxSequenceID)
}

// sortableTransactionShortID wraps around the types.TransactionShortID,
// as to ensure it is encoded in a way that allows boltdb use it for natural ordering.
type sortableTransactionShortID types.TransactionShortID

func newSortableTransactionShortID(height types.BlockHeight, txSequenceID uint16) sortableTransactionShortID {
	return sortableTransactionShortID(types.NewTransactionShortID(height, txSequenceID))
}

// MarshalSia implements SiaMarshaler.MarshalSia,
// alias of MarshalRivine for backwards-compatibility reasons.
func (sid sortableTransactionShortID) MarshalSia(w io.Writer) error {
	return sid.MarshalRivine(w)
}

// UnmarshalSia implements SiaMarshaler.UnmarshalSia,
// alias of UnmarshalRivine for backwards-compatibility reasons.
func (sid *sortableTransactionShortID) UnmarshalSia(r io.Reader) error {
	return sid.UnmarshalRivine(r)
}

// MarshalRivine implements RivineMarshaler.MarshalRivine
func (sid sortableTransactionShortID) MarshalRivine(w io.Writer) error {
	var b [8]byte
	binary.BigEndian.PutUint64(b[:], uint64(sid))
	n, err := w.Write(b[:])
	if err != nil {
		return err
	}
	if n != 8 {
		return io.ErrShortWrite
	}
	return nil
}

// UnmarshalRivine implements RivineUnmarshaler.UnmarshalRivine
func (sid *sortableTransactionShortID) UnmarshalRivine(r io.Reader) error {
	var b [8]byte
	n, err := r.Read(b[:])
	if err != nil {
		return err
	}
	if n != 8 {
		return io.ErrUnexpectedEOF
	}
	*sid = sortableTransactionShortID(binary.BigEndian.Uint64(b[:]))
	return nil
}
