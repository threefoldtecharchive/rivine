package datastore

import (
	"fmt"

	"github.com/rivine/rivine/build"
	"github.com/rivine/rivine/modules"
	"github.com/rivine/rivine/types"
)

type (
	// BlockBuffer is a ring buffer like implementation of a BlockFrame array.
	BlockBuffer struct {
		buffer []*BlockFrame
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

	// BlockFrame is an element for a block buffer
	BlockFrame struct {
		Block types.Block
		// CCID is the ID of the consensus change in which the block was received
		CCID modules.ConsensusChangeID
	}
)

// NewBlockBuffer creates a new BlockBuffer, with the size equal to the chain maturity delay (types.Maturitydelay).
func NewBlockBuffer() *BlockBuffer {
	return NewSizedBlockBuffer(int32(types.MaturityDelay))
}

// NewSizedBlockBuffer creates a new BlockBuffer with a configurable size
func NewSizedBlockBuffer(size int32) *BlockBuffer {
	// Make sure the buffer always has some space
	if size <= 0 {
		build.Critical("Trying to create an invalid sized buffer")
	}
	return &BlockBuffer{
		size:   size,
		head:   -1,
		count:  0,
		buffer: make([]*BlockFrame, size),
	}
}

// NewBlockFrame creates a new block frame for a given block and consensus change id
func NewBlockFrame(block types.Block, ccID modules.ConsensusChangeID) *BlockFrame {
	return &BlockFrame{
		Block: block,
		CCID:  ccID,
	}
}

// Push pushes a BlockFrame into the next element in the buffer. If the block buffer is full, the
// oldest frame is returned and then overwritten
func (bdq *BlockBuffer) Push(frame *BlockFrame) *BlockFrame {
	// First increment the head pointer to point to the new location
	bdq.head = (bdq.head + 1) % bdq.size

	// Check if there is already something there
	var lf *BlockFrame
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

// Pop removes the top (newest) element from the buffer. If the buffer is empty, nil is returned.
// We require passing in the blockID from the block to pop as a safety check, which should never fail
// due to the guarantee of ordering when we receive the consensus changes.
func (bdq *BlockBuffer) Pop(id types.BlockID) *BlockFrame {
	// Check if there are elements to pop
	if bdq.count == 0 {
		return nil
	}

	if bdq.buffer[bdq.head].Block.ID() != id {
		build.Critical(fmt.Sprintf("Trying to pop from a block buffer with a wrong block id, trying to pop id %v, but id of head is %v",
			id, bdq.buffer[bdq.head].Block.ID()))
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
