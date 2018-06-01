package datastore

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/rivine/rivine/modules"

	"github.com/rivine/rivine/types"
)

func TestBlockBuffer(t *testing.T) {
	// Generate some random blocks
	// The block is likely invalid but that doesn't really matter, since block validity is a concept
	// which only holds in a consensus set which is not used in these tests
	testBlocks := [10][]types.Block{}
	for i := 0; i < len(testBlocks); i++ {
		testBlocks[i] = make([]types.Block, (i+1)*(i+1))
		for j := 0; j < len(testBlocks[i]); j++ {
			testBlocks[i][j] = generateRandomBlock()
		}
	}

	bufferSizes := []int{1, 5, 10, 25, 50}

	t.Parallel()
	for _, bs := range bufferSizes {
		for _, tb := range testBlocks {
			t.Run(fmt.Sprintf("Test push %d-%d", bs, len(tb)), getBlockBufferPushTest(bs, tb))
			t.Run(fmt.Sprintf("Test pop %d-%d", bs, len(tb)), getBlockBufferPopTest(bs, tb))
		}
	}

}

func getBlockBufferPushTest(bufferSize int, blocks []types.Block) func(*testing.T) {
	return func(t *testing.T) {
		buf := newSizedBlockBuffer(int32(bufferSize))
		blockCount := len(blocks)
		for i := 0; i < blockCount; i++ {
			// Set a random consensus change id, does not really matter here
			oldFrame := buf.push(newBlockFrame(blocks[i], modules.ConsensusChangeID{byte(i)}))
			// Buffer is not at capicity yet so it should not have returned a frame
			if i < bufferSize && oldFrame != nil {
				t.Error("Buffer should not be at capacity but it returned a frame")
			}
			if i > bufferSize && oldFrame.block.ID() != blocks[i-bufferSize].ID() {
				t.Error("Wrong block returned when pushingn new block on a full buffer")
			}
		}
	}
}

func getBlockBufferPopTest(bufferSize int, blocks []types.Block) func(*testing.T) {
	return func(t *testing.T) {
		// First push all blocks, then remove them again and check if results are as expected
		buf := newSizedBlockBuffer(int32(bufferSize))
		blockCount := len(blocks)
		for i := 0; i < blockCount; i++ {
			buf.push(newBlockFrame(blocks[i], modules.ConsensusChangeID{byte(i)}))
		}
		for i := blockCount - 1; i >= 0; i-- {
			fr := buf.pop(blocks[i].ID())
			if i >= blockCount-bufferSize && fr == nil {
				t.Error("Buffer should not be empty but no frame was popped")
			}
			if i < blockCount-bufferSize && fr != nil {
				t.Error("Buffer should be empty but pop returned a frame")
			}
		}

		// Push 2, then pop one until all blocks are in the buffer. Then check if the correct blocks are poped again
		buf = newSizedBlockBuffer(int32(bufferSize))
		for i := 0; i < blockCount; i++ {
			// First pop the previous element.
			// We could also pops on the empty buffer, since thats a NOOP, but that requires a fake blockID
			if i%2 == 0 && i > 0 {
				buf.pop(blocks[i-1].ID())
			}
			// Add the current element
			buf.push(newBlockFrame(blocks[i], modules.ConsensusChangeID{byte(i)}))
		}

		// Pop all the frames, ensure that they are as expected
		// All blocks on even index, + the last index will be in the buffer (older blocks could be gone if the buffer is too small)
		count := bufferSize
		for i := blockCount - 1; i >= 0; i-- {
			if i%2 != 0 && i != blockCount-1 {
				continue
			}
			fr := buf.pop(blocks[i].ID())
			// Check if the frame's block ID is as expected. We could skip this if we always build the
			// tests in debug mode since then the wrong ID will panic inside the Pop function, but lets not rely
			// on that
			if fr != nil && fr.block.ID() != blocks[i].ID() {
				t.Errorf("Block in pop'ed framed has wrong ID, expected %v, but got %v", blocks[i].ID(), fr.block.ID())
			}
			if count <= 0 && fr != nil {
				t.Error("Got a frame while there should not be any left")
			}
			if count > 0 {
				if fr == nil {
					t.Error("Failed to retrieve frame while buffer should not be empty")
				}
				count--
			}
		}
	}
}

// Generate a random block. Only set a random timestamp as this is enough to change the block id
func generateRandomBlock() types.Block {
	block := types.Block{
		Timestamp: types.Timestamp(rand.Uint64()),
	}
	return block
}
