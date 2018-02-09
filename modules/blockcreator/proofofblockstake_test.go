package blockcreator

import (
	"math/big"
	"testing"

	"github.com/rivine/rivine/modules"
	"github.com/rivine/rivine/types"
)

type consensusSetStub struct {
}

func (cs consensusSetStub) AcceptBlock(t types.Block) (e error) { return }

// this function will artificially give a block which blockID has a specific bit set.
// bit 255-(blockheight modulo 256) is always 1
// this behavior is made to test the stakemodifier
func (cs consensusSetStub) BlockAtHeight(blockheight types.BlockHeight) (types.Block, bool) {

	bitval := uint32(blockheight) % 8
	byteval := (uint32(blockheight) / 8) % 32

	counter := 0
	var block types.Block
	for { //on average this loop is done 2 times
		block = types.Block{
			Timestamp: types.Timestamp(counter),
		}
		hashof := block.ID()
		if (hashof[byteval] & (1 << (7 - bitval))) > 0 {
			break
		}
		counter++
	}

	return block, true
}

func (cs consensusSetStub) ChildTarget(types.BlockID) (t types.Target, b bool) { return }

func (cs consensusSetStub) Close() (e error) { return }

func (cs consensusSetStub) ConsensusSetSubscribe(modules.ConsensusSetSubscriber, modules.ConsensusChangeID) (e error) {
	return
}

func (cs consensusSetStub) CurrentBlock() (t types.Block) { return }

func (cs consensusSetStub) Flush() (e error) { return }

func (cs consensusSetStub) Height() (h types.BlockHeight) { return }

func (cs consensusSetStub) Synced() bool { return false }

func (cs consensusSetStub) InCurrentPath(types.BlockID) bool { return false }

func (cs consensusSetStub) MinimumValidChildTimestamp(types.BlockID) (t types.Timestamp, b bool) {
	return
}

func (cs consensusSetStub) TryTransactionSet([]types.Transaction) (csc modules.ConsensusChange, e error) {
	return
}

func (cs consensusSetStub) Unsubscribe(modules.ConsensusSetSubscriber) { return }

func (cs consensusSetStub) BlockHeightOfBlock(types.Block) (types.BlockHeight, bool) { return 0, false }

func (cs consensusSetStub) CalculateStakeModifier(height types.BlockHeight) *big.Int { return nil }

// TestCurrencyDiv64 checks that the Div64 function has been correctly implemented.
// TODO: CalculateStakeModifier seems to have been moved to the consensusset
// Moved in: https://github.com/rivine/rivine/commit/473885476ccbcc6c259fa55e3af2aa818fe26db6#diff-4a95df743bd21a9b70774bb3981eac9a
func TestCalculateStakeModifier(t *testing.T) {

	alloneshash := new(big.Int).Sub(new(big.Int).Exp(big.NewInt(2), big.NewInt(256), nil), big.NewInt(1))
	cs := consensusSetStub{}
	types.StakeModifierDelay = 2000

	persist := persistence{Height: types.BlockHeight(2254)} //this blockheight is the current blockheight
	bc := BlockCreator{cs: cs, persist: persist}
	stakemodifier := bc.CalculateStakeModifier()
	if stakemodifier.Cmp(alloneshash) != 0 {
		t.Error("The stakemodifier should be all binary 1's in this specific case")
	}

	persist2 := persistence{Height: types.BlockHeight(2510)} //
	bc2 := BlockCreator{cs: cs, persist: persist2}
	stakemodifier2 := bc2.CalculateStakeModifier()
	if stakemodifier2.Cmp(alloneshash) != 0 {
		t.Error("The stakemodifier should be all binary 1's in this specific case")
	}

	stakemodifierblockheight0 := big.NewInt(0)
	stakemodifierblockheight0.SetString("3C6215E851515D3CBABBBD19002F7E6ABCFCE86CDA4856414165071590AB394", 16)
	persist3 := persistence{Height: types.BlockHeight(0)}
	bc3 := BlockCreator{cs: cs, persist: persist3}
	stakemodifier3 := bc3.CalculateStakeModifier()
	if stakemodifier3.Cmp(stakemodifierblockheight0) != 0 {
		t.Error("The stakemodifier is deterministic for blockheight 0 and StakeModifierDelay 2000 because the blockmodifier is calculated with deterministic blockIDs pre genesis")
	}

	stakemodifierblockheight345 := big.NewInt(0)
	stakemodifierblockheight345.SetString("E7D660EB51B8872A26173D6BC280F8C308E626B1C6CF3F3DBCEF4D9BA6D78AFD", 16)
	persist4 := persistence{Height: types.BlockHeight(345)}
	bc4 := BlockCreator{cs: cs, persist: persist4}
	stakemodifier4 := bc4.CalculateStakeModifier()
	if stakemodifier4.Cmp(stakemodifierblockheight345) != 0 {
		t.Error("The stakemodifier is deterministic for blockheight 345 and StakeModifierDelay 2000 because the blockmodifier is calculated with deterministic blockIDs pre genesis")
	}

	persist5 := persistence{Height: types.BlockHeight(12345)}
	bc5 := BlockCreator{cs: cs, persist: persist5}
	stakemodifier5 := bc5.CalculateStakeModifier()
	persist6 := persistence{Height: types.BlockHeight(12345 + 256)}
	bc6 := BlockCreator{cs: cs, persist: persist6}
	stakemodifier6 := bc6.CalculateStakeModifier()
	if stakemodifier5.Cmp(stakemodifier6) != 0 {
		t.Error("The blockIDs are artificially generated and are the same every 256 blocks -> The stakemodifier is also the same ")
	}

}
