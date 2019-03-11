package consensus

import (
	"bytes"
	"crypto/sha256"
	"math/big"
	"strconv"

	"github.com/threefoldtech/rivine/build"
	"github.com/threefoldtech/rivine/types"
)

// CalculateStakeModifier calculates the stakemodifier from the blockchain.
// Height is the height for which to calculate the stakemodifier. This is needed in case height - StakeModifierDelay goes
// 	sub-genesis
// Block is a block on the correct fork, so we can roll back from it.
// Delay is the amount of blocks we roll back from `block`.
// Block and delay must be chosen so that rolling back `delay` blocks from `block` reaches first block with which to calculate
// the stakemodifier.
func (cs *ConsensusSet) CalculateStakeModifier(height types.BlockHeight, block types.Block, delay types.BlockHeight) *big.Int {
	//TODO: check if a new Stakemodifier needs to be calculated. The stakemodifier
	// only change when a new block is created, and this calculation is also needed
	// to validate an incomming new block

	// make a signed version of the current height because sub genesis block is
	// possible here.
	signedHeight := int64(height)
	signedHeight -= int64(cs.chainCts.StakeModifierDelay)

	mask := big.NewInt(1)
	var BlockIDHash *big.Int
	stakemodifier := big.NewInt(0)
	var buffer bytes.Buffer

	// Rollback the required amount of blocks, minus 1. This way we end up at the direct child of the
	// block we use to calculate the stakemodifer, rather than the actual first block. Simplifies
	// the main loop a bit
	block, _ = cs.FindParentBlock(block, delay-1)

	// We have the direct child of the first block used in the stake modifier calculation. As such
	// we can follow the parentID in the block to retrieve all the blocks required, using 1 bit
	// of each blocks ID to calculate the stake modifier
	for i := 0; i < 256; i++ {
		if signedHeight >= 0 {
			var exist bool
			block, exist = cs.FindParentBlock(block, 1)
			if !exist {
				build.Severe("block to be used for stakemodifier does not yet exist")
			}
			hashof := block.ID()
			BlockIDHash = big.NewInt(0).SetBytes(hashof[:])
		} else {
			// if the counter goes sub genesis block , calculate a predefined hash
			// from the sub genesis distance.
			buffer.WriteString("genesis" + strconv.FormatInt(signedHeight, 10))
			hashof := sha256.Sum256(buffer.Bytes())
			BlockIDHash = big.NewInt(0).SetBytes(hashof[:])
		}

		stakemodifier.Or(stakemodifier, big.NewInt(0).And(BlockIDHash, mask))
		mask.Mul(mask, big.NewInt(2)) //shift 1 bit to left (more close to msb)
		signedHeight--
	}
	return stakemodifier
}
