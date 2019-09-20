package consensus

import (
	"sort"

	"github.com/threefoldtech/rivine/pkg/encoding/siabin"
	"github.com/threefoldtech/rivine/types"
)

// blockRuleHelper assists with block validity checks by calculating values
// on blocks that are relevant to validity rules.
type blockRuleHelper interface {
	minimumValidChildTimestamp(dbBucket, *processedBlock) types.Timestamp
}

// stdBlockRuleHelper is the standard implementation of blockRuleHelper.
type stdBlockRuleHelper struct {
	chainCts types.ChainConstants

	headerCache *ringBuf
}

func newStdBlockRuleHelper(chainCts types.ChainConstants) stdBlockRuleHelper {
	return stdBlockRuleHelper{
		chainCts:    chainCts,
		headerCache: newRingBuf(chainCts.MedianTimestampWindow * 2),
	}
}

// minimumValidChildTimestamp returns the earliest timestamp that a child node
// can have while still being valid. See section 'Block Timestamps' in
// Consensus.md.
//
// To boost performance, minimumValidChildTimestamp is passed a bucket that it
// can use from inside of a boltdb transaction.
func (rh stdBlockRuleHelper) minimumValidChildTimestamp(blockMap dbBucket, pb *processedBlock) types.Timestamp {
	// Get the previous MedianTimestampWindow timestamps.
	windowTimes := make(types.TimestampSlice, rh.chainCts.MedianTimestampWindow)
	windowTimes[0] = pb.Block.Timestamp
	parent := pb.Block.ParentID
	var timestamp types.Timestamp
	for i := uint64(1); i < rh.chainCts.MedianTimestampWindow; i++ {
		// If the genesis block is 'parent', use the genesis block timestamp
		// for all remaining times.
		if parent == (types.BlockID{}) {
			windowTimes[i] = windowTimes[i-1]
			continue
		}

		parentHeader, exists := rh.headerCache.Get(parent)
		if !exists {
			parentBytes := blockMap.Get(parent[:])
			// remember id of current block
			id := parent
			// Get the next parent's bytes. Because the ordering is specific, the
			// parent does not need to be decoded entirely to get the desired
			// information. This provides a performance boost. The id of the next
			// parent lies at the first 32 bytes, and the timestamp of the block
			// lies at bytes 32-40.
			copy(parent[:], parentBytes[:32])
			timestamp = types.Timestamp(siabin.DecUint64(parentBytes[32:40]))
			header := cachedHeaderTimestamp{id: id, parent: parent, timestamp: windowTimes[i]}
			rh.headerCache.Push(header)
		} else {
			timestamp = parentHeader.timestamp
			parent = parentHeader.parent
		}
		windowTimes[i] = timestamp
	}
	sort.Sort(windowTimes)

	// Return the median of the sorted timestamps.
	return windowTimes[len(windowTimes)/2]
}

type ringBuf struct {
	buf           []cachedHeaderTimestamp
	newHead, size uint64
}

// cachedHeaderTimestamp contains the important info for calculating the minimum valid
// timestamp of a block, without having to reload entire blocks from disk. It keeps
// track of the timestamp of a header, the parent ID of the associated header (so it
// can be used to traverse the chain), as well as the ID of this header.
//
// We can't wrap the regular header, since the header is only an in-memory structure,
// so it can't be trivially loaded/decoded from the db, and it contains a merkleroot,
// which means we need to decode the whole block to calculate it. This is inefficient
// for our purposes
type cachedHeaderTimestamp struct {
	parent    types.BlockID
	timestamp types.Timestamp
	id        types.BlockID
}

func newRingBuf(size uint64) *ringBuf {
	return &ringBuf{
		buf: make([]cachedHeaderTimestamp, size),
		// newHead points to the slot in the buffer that
		// will contain the next header
		newHead: 0,
		size:    size,
	}
}

// Push a new element into the buffer
func (rb *ringBuf) Push(h cachedHeaderTimestamp) {
	rb.buf[rb.newHead] = h
	rb.newHead = (rb.newHead + 1) % rb.size
}

// Get a header based on the id.
func (rb *ringBuf) Get(id types.BlockID) (cachedHeaderTimestamp, bool) {
	for i := range rb.buf {
		if rb.buf[i].id == id {
			return rb.buf[i], true
		}
	}
	return cachedHeaderTimestamp{}, false
}
