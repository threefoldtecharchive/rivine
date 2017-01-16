package blockcreator

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"math/big"
	"strconv"
	"time"

	"github.com/rivine/rivine/crypto"
	"github.com/rivine/rivine/types"
)

// SolveBlocks participates in the Proof Of Block Stake protocol by continously checking if
// unspent block stake outputs make a solution for the current unsolved block.
// If a match is found, the block is submitted to the consensus set.
// This function does not return until the blockcreator threadgroup is stopped.
func (bc *BlockCreator) SolveBlocks() {
	for {

		// Bail if 'Stop' has been called.
		select {
		case <-bc.tg.StopChan():
			return
		default:
		}

		// TODO: where to put the lock exactly
		// Try to solve a block for blocktimes of the next 10 seconds
		now := time.Now().Unix()
		b := bc.solveBlock(uint64(now), 10)
		if b != nil {
			bjson, _ := json.Marshal(b)
			bc.log.Debugln("Solved block:", string(bjson))

			err := bc.submitBlock(*b)
			if err != nil {
				bc.log.Println("ERROR: An error occurred while submitting a solved block:", err)
			}
		}
		//sleep a while before recalculating
		time.Sleep(8 * time.Second)
	}
}

// CalculateStakeModifier calculates the stakemodifier from the blockchain.
func (bc *BlockCreator) CalculateStakeModifier() *big.Int {
	//TODO: check if a new Stakemodifier needs to be calculated. The stakemodifier
	// only change when a new block is created

	// make a signed version of the current height because sub genesis block is
	// possible here.
	signedHeight := int64(bc.persist.Height) + 1
	signedHeight -= int64(types.StakeModifierDelay)

	mask := big.NewInt(1)
	BlockIDHash := big.NewInt(0)
	stakemodifier := big.NewInt(0)
	var buffer bytes.Buffer

	// now signedHeight points to the first block to use for the stakemodifier
	// calculation, we count down 256 blocks and use 1 bit of each blocks ID to
	// calculate the stakemodifier
	for i := 0; i < 256; i++ {
		if signedHeight >= 0 {
			// If the genesis block is not yet reached use the ID of the current block
			BlockID, _ := bc.cs.BlockAtHeight(types.BlockHeight(signedHeight))
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

func (bc *BlockCreator) solveBlock(startTime uint64, secondsInTheFuture uint64) (b *types.Block) {

	stakemodifier := bc.CalculateStakeModifier()
	cbid := bc.cs.CurrentBlock().ID()
	target, _ := bc.cs.ChildTarget(cbid)

	// Try all unspent blockstake outputs
	unspentBlockStakeOutputs := bc.wallet.GetUnspentBlockStakeOutputs()
	for _, ubso := range unspentBlockStakeOutputs {
		BlockStakeAge := types.Timestamp(0)
		// Filter all unspent block stakes for aging. If the index of the unspent
		// block stake output is not the first transaction with the first index,
		// then block stake can only be used to solve blocks after its aging is
		// older than types.BlockStakeAging (more than 1 day)
		if ubso.Indexes.TransactionIndex != 0 || ubso.Indexes.OutputIndex != 0 {
			blockatheigh, _ := bc.cs.BlockAtHeight(ubso.Indexes.BlockHeight)
			BlockStakeAge = blockatheigh.Header().Timestamp + types.Timestamp(types.BlockStakeAging)
		}
		// Try all timestamps for this timerange
		for blocktime := startTime; blocktime < startTime+secondsInTheFuture; blocktime++ {
			if BlockStakeAge > types.Timestamp(blocktime) {
				continue
			}
			// Calculate the hash for the given unspent output and timestamp
			pobshash := crypto.HashAll(stakemodifier.Bytes(), ubso.Indexes.BlockHeight, ubso.Indexes.TransactionIndex, ubso.Indexes.OutputIndex, blocktime)
			// Check if it meets the difficulty
			pobshashvalue := big.NewInt(0).SetBytes(pobshash[:])

			if pobshashvalue.Div(pobshashvalue, ubso.Value.Big()).Cmp(target.Int()) == -1 {
				bc.log.Debugln("\nSolved block with target", target)
				blockToSubmit := types.Block{
					ParentID:   bc.unsolvedBlock.ParentID,
					Timestamp:  types.Timestamp(blocktime),
					POBSOutput: ubso.Indexes,
				}

				bc.RespentBlockStake(ubso)

				// Block is going to be passed to external memory, but the memory pointed
				// to by the transactions slice is still being modified - needs to be
				// copied.
				txns := make([]types.Transaction, len(bc.unsolvedBlock.Transactions))
				copy(txns, bc.unsolvedBlock.Transactions)
				blockToSubmit.Transactions = txns
				// Collect the transaction fees
				collectedMinerFees := blockToSubmit.CalculateSubsidy()
				if collectedMinerFees.Cmp(types.ZeroCurrency) != 0 {
					blockToSubmit.MinerPayouts = []types.CoinOutput{{Value: collectedMinerFees, UnlockHash: ubso.UnlockHash}}
				}

				return &blockToSubmit
			}
		}
	}
	return
}

// RespentBlockStake will spent the unspent block stake output which is needed
// for the POBS algorithm. The transaction created will be the first transaction
// in the block to avoid the BlockStakeAging for later use of this block stake.
func (bc *BlockCreator) RespentBlockStake(ubso types.UnspentBlockStakeOutput) {

	// There is a special case: When the unspent block stake output is allready
	// used in another transaction in this unsolved block, this extra transaction
	// is obsolete
	for _, ubstr := range bc.unsolvedBlock.Transactions {
		for _, ubstrinput := range ubstr.BlockStakeInputs {
			if ubstrinput.ParentID == ubso.BlockStakeOutputID {
				return
			}
		}
	}

	//otherwise the blockstake is not yet spent in this block, spent it now
	t := bc.wallet.StartTransaction()
	t.SpendBlockStake(ubso.BlockStakeOutputID) // link the input of this transaction
	// to the used BlockStake output

	bso := types.BlockStakeOutput{
		Value:      ubso.Value,      //use the same amount of BlockStake
		UnlockHash: ubso.UnlockHash, //use the same unlockhash.
	}
	ind := t.AddBlockStakeOutput(bso)
	if ind != 0 {
		// should not happen //TODO: not right error
	}
	txnSet, err := t.Sign(true)
	if err != nil {
		// should not happen //TODO: not right error
	}
	//Only one transaction is generated for this.
	if len(txnSet) > 1 {
		// should not happen //TODO: not right error
	}
	//add this transaction in front of the list of unsolved block transactions
	bc.unsolvedBlock.Transactions = append(txnSet, bc.unsolvedBlock.Transactions...)

	return
}
