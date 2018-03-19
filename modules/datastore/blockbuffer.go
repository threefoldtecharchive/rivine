package datastore

import (
	"fmt"

	"github.com/rivine/rivine/build"
	"github.com/rivine/rivine/modules"
	"github.com/rivine/rivine/types"
)

type (
	// blockBuffer is a ring buffer like implementation of a BlockFrame array.
	blockBuffer struct {
		buffer []*blockFrame
		// size *should* technically be a uint, as it can't be negative,
		// but that would require a lot casting when moving head, so
		// lets go with an int type
		size int32
		// head points to the top element of the buffer
		head int32
		// count keeps track of the amount of elements in the buffer
		// int type to match size
		count int32
	}

	// blockFrame is an element for a block buffer
	blockFrame struct {
		block types.Block
		// ccID is the ID of the consensus change in which the block was received
		ccID modules.ConsensusChangeID
	}
)

// newBlockBuffer creates a new BlockBuffer, with the size equal to the chain maturity delay (types.Maturitydelay).
func (ds *DataStore) newBlockBuffer() *blockBuffer {
	return newSizedBlockBuffer(int32(ds.chainCts.MaturityDelay))
}

// newSizedBlockBuffer creates a new blockBuffer with a configurable size
func newSizedBlockBuffer(size int32) *blockBuffer {
	// Make sure the buffer always has some space
	if size <= 0 {
		build.Critical("Trying to create an invalid sized buffer")
	}
	return &blockBuffer{
		size:   size,
		head:   -1,
		count:  0,
		buffer: make([]*blockFrame, size),
	}
}

// newBlockFrame creates a new block frame for a given block and consensus change id
func newBlockFrame(block types.Block, ccID modules.ConsensusChangeID) *blockFrame {
	return &blockFrame{
		block: block,
		ccID:  ccID,
	}
}

// push pushes a blockFrame into the next element in the buffer. If the block buffer is full, the
// oldest frame is returned and then overwritten
func (bdq *blockBuffer) push(frame *blockFrame) *blockFrame {
	// First increment the head pointer to point to the new location
	bdq.head = (bdq.head + 1) % bdq.size

	// Check if there is already something there
	var lf *blockFrame
	if bdq.count == bdq.size {
		// We are pointing at the oldest element so remove it
		lf = bdq.buffer[bdq.head]
		// Sanity check
		if lf == nil {
			build.Critical("Buffer should be full but we found a nil element")
		}
	} else {
		bdq.count++
	}
	// Add the new frame
	bdq.buffer[bdq.head] = frame
	return lf
}

// pop removes the top (newest) element from the buffer. If the buffer is empty, nil is returned.
// We require passing in the blockID from the block to pop as a safety check, which should never fail
// due to the guarantee of ordering when we receive the consensus changes.
func (bdq *blockBuffer) pop(id types.BlockID) *blockFrame {
	// Check if there are elements to pop
	if bdq.count == 0 {
		return nil
	}

	if bdq.buffer[bdq.head].block.ID() != id {
		build.Critical(fmt.Sprintf("Trying to pop from a block buffer with a wrong block id, trying to pop id %v, but id of head is %v",
			id, bdq.buffer[bdq.head].block.ID()))
	}

	frame := bdq.buffer[bdq.head]
	// Delete the element
	bdq.buffer[bdq.head] = nil
	// Move head to the right index
	bdq.head = (bdq.head - 1 + bdq.size) % bdq.size
	// Acknowledge that we removed an element
	bdq.count--
	return frame
}
