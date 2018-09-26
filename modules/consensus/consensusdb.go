package consensus

// consensusdb.go contains all of the functions related to performing consensus
// related actions on the database, including initializing the consensus
// portions of the database. Many errors cause panics instead of being handled
// gracefully, but only when the debug flag is set. The errors are silently
// ignored otherwise, which is suboptimal.

import (
	"github.com/rivine/rivine/build"
	"github.com/rivine/rivine/encoding"
	"github.com/rivine/rivine/modules"
	"github.com/rivine/rivine/types"

	"github.com/rivine/bbolt"
)

var (
	prefixDCO = []byte("dco_")

	// BlockHeight is a bucket that stores the current block height.
	//
	// Generally we would just look at BlockPath.Stats(), but there is an error
	// in boltdb that prevents the bucket stats from updating until a tx is
	// committed. Wasn't a problem until we started doing the entire block as
	// one tx.
	//
	// DEPRECATED - block.Stats() should be sufficient to determine the block
	// height, but currently stats are only computed after committing a
	// transaction, therefore cannot be assumed reliable.
	BlockHeight = []byte("BlockHeight")

	// BlockMap is a database bucket containing all of the processed blocks,
	// keyed by their id. This includes blocks that are not currently in the
	// consensus set, and blocks that may not have been fully validated yet.
	BlockMap = []byte("BlockMap")

	// BlockPath is a database bucket containing a mapping from the height of a
	// block to the id of the block at that height. BlockPath only includes
	// blocks in the current path.
	BlockPath = []byte("BlockPath")

	// Consistency is a database bucket with a flag indicating whether
	// inconsistencies within the database have been detected.
	Consistency = []byte("Consistency")

	// CoinOutputs is a database bucket that contains all of the unspent
	// coin outputs.
	CoinOutputs = []byte("CoinOutputs")

	// BlockStakeOutputs is a database bucket that contains all of the unspent
	// blockstake outputs.
	BlockStakeOutputs = []byte("BlockStakeOutputs")

	// TransactionIDMap is a database bucket that containsall of the present
	// transaction IDs linked to their short ID
	TransactionIDMap = []byte("TransactionIDMap")
)

// createConsensusObjects initialzes the consensus portions of the database.
func (cs *ConsensusSet) createConsensusDB(tx *bolt.Tx) error {
	// Enumerate and create the database buckets.
	buckets := [][]byte{
		BlockHeight,
		BlockMap,
		BlockPath,
		Consistency,
		CoinOutputs,
		BlockStakeOutputs,
		TransactionIDMap,
	}
	for _, bucket := range buckets {
		_, err := tx.CreateBucket(bucket)
		if err != nil {
			return err
		}
	}

	// Set the block height to -1, so the genesis block is at height 0.
	blockHeight := tx.Bucket(BlockHeight)
	underflow := types.BlockHeight(0)
	err := blockHeight.Put(BlockHeight, encoding.Marshal(underflow-1))
	if err != nil {
		return err
	}

	// Update the blockstake and coin output diffs map for the genesis block on disk. This
	// needs to happen between the database being opened/initilized and the
	// consensus set hash being calculated
	for _, cod := range cs.blockRoot.CoinOutputDiffs {
		commitCoinOutputDiff(tx, cod, modules.DiffApply)
	}
	for _, sfod := range cs.blockRoot.BlockStakeOutputDiffs {
		commitBlockStakeOutputDiff(tx, sfod, modules.DiffApply)
	}

	// Add the genesis block to the block structures - checksum must be taken
	// after pushing the genesis block into the path.
	pushPath(tx, cs.blockRoot.Block.ID())
	if build.DEBUG {
		cs.blockRoot.ConsensusChecksum = consensusChecksum(tx)
	}
	addBlockMap(tx, &cs.blockRoot)
	return nil
}

// blockHeight returns the height of the blockchain.
func blockHeight(tx *bolt.Tx) (height types.BlockHeight) {
	bh := tx.Bucket(BlockHeight)
	err := encoding.Unmarshal(bh.Get(BlockHeight), &height)
	if build.DEBUG && err != nil {
		panic(err)
	}
	return
}

// blockTimeStamp returns the timestamp of the block on the given height.
func blockTimeStamp(tx *bolt.Tx, height types.BlockHeight) (types.Timestamp, error) {
	id, err := getPath(tx, height)
	if err != nil {
		return 0, err
	}
	pb, err := getBlockMap(tx, id)
	if err != nil {
		return 0, err
	}
	return pb.Block.Timestamp, nil
}

// currentBlockID returns the id of the most recent block in the consensus set.
func currentBlockID(tx *bolt.Tx) types.BlockID {
	id, err := getPath(tx, blockHeight(tx))
	if build.DEBUG && err != nil {
		panic(err)
	}
	return id
}

// currentProcessedBlock returns the most recent block in the consensus set.
func currentProcessedBlock(tx *bolt.Tx) *processedBlock {
	pb, err := getBlockMap(tx, currentBlockID(tx))
	if build.DEBUG && err != nil {
		panic(err)
	}
	return pb
}

// getBlockMap returns a processed block with the input id.
func getBlockMap(tx *bolt.Tx, id types.BlockID) (*processedBlock, error) {
	// Look up the encoded block.
	pbBytes := tx.Bucket(BlockMap).Get(id[:])
	if pbBytes == nil {
		return nil, errNilItem
	}

	// Decode the block - should never fail.
	var pb processedBlock
	err := encoding.Unmarshal(pbBytes, &pb)
	if build.DEBUG && err != nil {
		panic(err)
	}
	return &pb, nil
}

// addBlockMap adds a processed block to the block map.
func addBlockMap(tx *bolt.Tx, pb *processedBlock) {
	id := pb.Block.ID()
	err := tx.Bucket(BlockMap).Put(id[:], encoding.Marshal(*pb))
	if build.DEBUG && err != nil {
		panic(err)
	}
}

// getPath returns the block id at 'height' in the block path.
func getPath(tx *bolt.Tx, height types.BlockHeight) (id types.BlockID, err error) {
	idBytes := tx.Bucket(BlockPath).Get(encoding.Marshal(height))
	if idBytes == nil {
		return types.BlockID{}, errNilItem
	}

	err = encoding.Unmarshal(idBytes, &id)
	if build.DEBUG && err != nil {
		panic(err)
	}
	return id, nil
}

// pushPath adds a block to the BlockPath at current height + 1.
func pushPath(tx *bolt.Tx, bid types.BlockID) {
	// Fetch and update the block height.
	bh := tx.Bucket(BlockHeight)
	heightBytes := bh.Get(BlockHeight)
	var oldHeight types.BlockHeight
	err := encoding.Unmarshal(heightBytes, &oldHeight)
	if build.DEBUG && err != nil {
		panic(err)
	}
	newHeightBytes := encoding.Marshal(oldHeight + 1)
	err = bh.Put(BlockHeight, newHeightBytes)
	if build.DEBUG && err != nil {
		panic(err)
	}

	// Add the block to the block path.
	bp := tx.Bucket(BlockPath)
	err = bp.Put(newHeightBytes, bid[:])
	if build.DEBUG && err != nil {
		panic(err)
	}
}

// popPath removes a block from the "end" of the chain, i.e. the block
// with the largest height.
func popPath(tx *bolt.Tx) {
	// Fetch and update the block height.
	bh := tx.Bucket(BlockHeight)
	oldHeightBytes := bh.Get(BlockHeight)
	var oldHeight types.BlockHeight
	err := encoding.Unmarshal(oldHeightBytes, &oldHeight)
	if build.DEBUG && err != nil {
		panic(err)
	}
	newHeightBytes := encoding.Marshal(oldHeight - 1)
	err = bh.Put(BlockHeight, newHeightBytes)
	if build.DEBUG && err != nil {
		panic(err)
	}

	// Remove the block from the path - make sure to remove the block at
	// oldHeight.
	bp := tx.Bucket(BlockPath)
	err = bp.Delete(oldHeightBytes)
	if build.DEBUG && err != nil {
		panic(err)
	}
}

// isCoinOutput returns true if there is a coin output of that id in the
// database.
func isCoinOutput(tx *bolt.Tx, id types.CoinOutputID) bool {
	bucket := tx.Bucket(CoinOutputs)
	sco := bucket.Get(id[:])
	return sco != nil
}

// getCoinOutput fetches a coin output from the database. An error is
// returned if the siacoin output does not exist.
func getCoinOutput(tx *bolt.Tx, id types.CoinOutputID) (types.CoinOutput, error) {
	scoBytes := tx.Bucket(CoinOutputs).Get(id[:])
	if scoBytes == nil {
		return types.CoinOutput{}, errNilItem
	}
	var sco types.CoinOutput
	err := encoding.Unmarshal(scoBytes, &sco)
	if err != nil {
		return types.CoinOutput{}, err
	}
	return sco, nil
}

// addCoinOutput adds a coin output to the database. An error is returned
// if the coin output is already in the database.
func addCoinOutput(tx *bolt.Tx, id types.CoinOutputID, sco types.CoinOutput) {
	// While this is not supposed to be allowed, there's a bug in the consensus
	// code which means that earlier versions have accetped 0-value outputs
	// onto the blockchain. A hardfork to remove 0-value outputs will fix this,
	// and that hardfork is planned, but not yet.
	/*
		if build.DEBUG && sco.Value.IsZero() {
			panic("discovered a zero value siacoin output")
		}
	*/
	coinOutputs := tx.Bucket(CoinOutputs)
	// Sanity check - should not be adding an item that exists.
	if build.DEBUG && coinOutputs.Get(id[:]) != nil {
		panic("repeat siacoin output")
	}
	err := coinOutputs.Put(id[:], encoding.Marshal(sco))
	if build.DEBUG && err != nil {
		panic(err)
	}
}

// removeCoinOutput removes a coin output from the database. An error is
// returned if the coin output is not in the database prior to removal.
func removeCoinOutput(tx *bolt.Tx, id types.CoinOutputID) {
	scoBucket := tx.Bucket(CoinOutputs)
	// Sanity check - should not be removing an item that is not in the db.
	if build.DEBUG && scoBucket.Get(id[:]) == nil {
		panic("nil siacoin output")
	}
	err := scoBucket.Delete(id[:])
	if build.DEBUG && err != nil {
		panic(err)
	}
}

// getBlockStakeOutput fetches a blockstake output from the database. An error is
// returned if the blockstake output does not exist.
func getBlockStakeOutput(tx *bolt.Tx, id types.BlockStakeOutputID) (types.BlockStakeOutput, error) {
	sfoBytes := tx.Bucket(BlockStakeOutputs).Get(id[:])
	if sfoBytes == nil {
		return types.BlockStakeOutput{}, errNilItem
	}
	var sfo types.BlockStakeOutput
	err := encoding.Unmarshal(sfoBytes, &sfo)
	if err != nil {
		return types.BlockStakeOutput{}, err
	}
	return sfo, nil
}

// addBlockStakeOutput adds a blockstake output to the database. An error is returned
// if the blockstake output is already in the database.
func addBlockStakeOutput(tx *bolt.Tx, id types.BlockStakeOutputID, sfo types.BlockStakeOutput) {
	blockstakeOutputs := tx.Bucket(BlockStakeOutputs)
	// Sanity check - should not be adding a blockstake output with a value of
	// zero.
	if build.DEBUG && sfo.Value.IsZero() {
		panic("zero value blockstake being added")
	}
	// Sanity check - should not be adding an item already in the db.
	if build.DEBUG && blockstakeOutputs.Get(id[:]) != nil {
		panic("repeat blockstake output")
	}
	err := blockstakeOutputs.Put(id[:], encoding.Marshal(sfo))
	if build.DEBUG && err != nil {
		panic(err)
	}
}

// removeBlockStakeOutput removes a blockstake output from the database. An error is
// returned if the blockstake output is not in the database prior to removal.
func removeBlockStakeOutput(tx *bolt.Tx, id types.BlockStakeOutputID) {
	sfoBucket := tx.Bucket(BlockStakeOutputs)
	if build.DEBUG && sfoBucket.Get(id[:]) == nil {
		panic("nil blockstake output")
	}
	err := sfoBucket.Delete(id[:])
	if build.DEBUG && err != nil {
		panic(err)
	}
}

// addTxnIDMapping adds a transaction ID mapping to the database.
func addTxnIDMapping(tx *bolt.Tx, longID types.TransactionID, shortID types.TransactionShortID) {
	txIDMapBucket := tx.Bucket(TransactionIDMap)
	// Sanity check - should not be adding an item already in the db.
	if build.DEBUG && txIDMapBucket.Get(longID[:]) != nil {
		panic("repeat transaction id mapping")
	}
	err := txIDMapBucket.Put(longID[:], encoding.Marshal(shortID))
	if build.DEBUG && err != nil {
		panic(err)
	}
}

// removeTxnIDMappng removes a transaction ID mapping from the database.
func removeTxnIDMapping(tx *bolt.Tx, longID types.TransactionID) {
	txIDMapBucket := tx.Bucket(TransactionIDMap)
	if build.DEBUG && txIDMapBucket.Get(longID[:]) == nil {
		panic("nil txID mapping")
	}
	err := txIDMapBucket.Delete(longID[:])
	if build.DEBUG && err != nil {
		panic(err)
	}
}

// getTransactionShortID returns a transaction short ID from
// a regular transaction ID
func getTransactionShortID(tx *bolt.Tx, id types.TransactionID) (types.TransactionShortID, error) {
	shortIDBytes := tx.Bucket(TransactionIDMap).Get(id[:])
	if shortIDBytes == nil {
		return types.TransactionShortID(0), errNilItem
	}
	var shortID types.TransactionShortID
	err := encoding.Unmarshal(shortIDBytes, &shortID)
	return shortID, err
}

// addDCO adds a delayed coin output to the consensus set.
func addDCO(tx *bolt.Tx, bh types.BlockHeight, id types.CoinOutputID, sco types.CoinOutput) {
	// Sanity check - dco should never have a value of zero.
	if build.DEBUG && sco.Value.IsZero() {
		panic("zero-value dco being added")
	}
	// Sanity check - output should not already be in the full set of outputs.
	if build.DEBUG && tx.Bucket(CoinOutputs).Get(id[:]) != nil {
		panic("dco already in output set")
	}
	dscoBucketID := append(prefixDCO, encoding.EncUint64(uint64(bh))...)
	dscoBucket := tx.Bucket(dscoBucketID)
	// Sanity check - should not be adding an item already in the db.
	if build.DEBUG && dscoBucket.Get(id[:]) != nil {
		panic(errRepeatInsert)
	}
	err := dscoBucket.Put(id[:], encoding.Marshal(sco))
	if build.DEBUG && err != nil {
		panic(err)
	}
}

// removeDCO removes a delayed siacoin output from the consensus set.
func removeDCO(tx *bolt.Tx, bh types.BlockHeight, id types.CoinOutputID) {
	bucketID := append(prefixDCO, encoding.Marshal(bh)...)
	// Sanity check - should not remove an item not in the db.
	dscoBucket := tx.Bucket(bucketID)
	if build.DEBUG && dscoBucket.Get(id[:]) == nil {
		panic("nil dco")
	}
	err := dscoBucket.Delete(id[:])
	if build.DEBUG && err != nil {
		panic(err)
	}
}

// createDCOBucket creates a bucket for the delayed coin outputs at the
// input height.
func createDCOBucket(tx *bolt.Tx, bh types.BlockHeight) {
	bucketID := append(prefixDCO, encoding.Marshal(bh)...)
	_, err := tx.CreateBucket(bucketID)
	if build.DEBUG && err != nil {
		panic(err)
	}
}

// deleteDCOBucket deletes the bucket that held a set of delayed coin outputs.
func deleteDCOBucket(tx *bolt.Tx, bh types.BlockHeight) {
	// Delete the bucket.
	bucketID := append(prefixDCO, encoding.Marshal(bh)...)
	bucket := tx.Bucket(bucketID)
	if build.DEBUG && bucket == nil {
		panic(errNilBucket)
	}

	// TODO: Check that the bucket is empty. Using Stats() does not work at the
	// moment, as there is an error in the boltdb code.

	err := tx.DeleteBucket(bucketID)
	if build.DEBUG && err != nil {
		panic(err)
	}
}
