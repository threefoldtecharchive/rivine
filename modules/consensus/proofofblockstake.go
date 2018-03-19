package consensus

import (
	"bytes"
	"crypto/sha256"
	"math/big"
	"strconv"

	"github.com/rivine/rivine/types"
)

// CalculateStakeModifier calculates the stakemodifier from the blockchain.
func (cs *ConsensusSet) CalculateStakeModifier(height types.BlockHeight) *big.Int {
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

	// now signedHeight points to the sfirst block to use for the stakemodifier
	// calculation, we count down 256 blocks and use 1 bit of each blocks ID to
	// calculate the stakemodifier
	for i := 0; i < 256; i++ {
		if signedHeight >= 0 {
			// If the genesis block is not yet reached use the ID of the current block
			BlockID, exist := cs.BlockAtHeight(types.BlockHeight(signedHeight))
			if !exist {
				panic("block to be used for stakemodifier does not yet exist")
			}
			hashof := BlockID.ID()
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
