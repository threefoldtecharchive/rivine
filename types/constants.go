package types

// constants.go contains the Sia constants. Depending on which build tags are
// used, the constants will be initialized to different values.
//
// CONTRIBUTE: We don't have way to check that the non-test constants are all
// sane, plus we have no coverage for them.

import (
	"errors"
	"math/big"
)

// The chain constants are a global var here since some of the functions in the types package use them
var (
	cts ChainConstants
)

// ChainConstants is a utility struct which groups together the chain configuration
// RootTarget, GenesisBlockStakeCount, GenesisCoinCount, GenesisBlock, GenesisID, and StartDifficulty, although exposed,
// should not be set manually. Instead, you should rely on the "Calculate()" function to fill
// these in.
type ChainConstants struct {
	// BlockSizeLimit is the maximum size a single block can have, in bytes
	BlockSizeLimit uint64
	RootDepth      Target
	// BlockFrequency is the average timespan between blocks, in seconds.
	// I.E.: On average, 1 block will be created every 1 in *BlockFrequency* seconds
	BlockFrequency BlockHeight
	// MaturityDelay is the amount of blocks for which a miner payout must "mature" before it
	// gets added to the consensus set. Until this time has passed, a miner payout cannot be spend
	MaturityDelay BlockHeight
	// GenesisTimestamp is the unix timestamp of the genesis block
	GenesisTimestamp      Timestamp
	RootTarget            Target
	MedianTimestampWindow uint64

	// TargetWindow is the amount of blocks to go back to adjust the difficulty of the network.
	TargetWindow BlockHeight
	// MaxAdjustmentUp is the maximum multiplier to difficulty over the course of 500 blocks
	MaxAdjustmentUp *big.Rat
	// MaxAdjustmentDown is the minimum multiplier to the difficulty over the course of 500 blocks
	MaxAdjustmentDown *big.Rat
	// FutureThreshold is the amount of seconds that a block timestamp can be "in the future",
	// while stil being accepted by the consensus set. I.E. a block is accepted if:
	// 	block timestamp < current timestamp + future treshold
	// Blocks who's timestamp is bigger than this value will not be accepted, but they might be
	// recondisered as soon as their timestamp is within the future treshold
	FutureThreshold Timestamp
	// ExtremeFutureThreshold is the maximum amount of time a block timstamp can be in the future
	// before sais block is outright rejected. Blocks who's timestamp is between now + FutureThreshold
	// and now + ExtremeFutureThreshold are kept and retried as soon as their timestamp is lower than
	// now + FutureThreshold. In case the block timestamp is higher than now + ExtremeFutureThreshold, we
	// consider that the block will no longer be valid as soon as its timestamp becomes accepteable, the block
	// will no longer be on the longest chain. Also, we can't keep all the blocks to eventually verify this as that
	// opens up a DOS vector
	ExtremeFutureThreshold Timestamp

	// StakeModifierDelay is the amount of blocks to go back to start calculating the Stake Modifier,
	// which is used in the proof of blockstake protoco. The formula for the Stake Modifier is as follows:
	// 	For x = 0 .. 255
	// 	bit x of Stake Modifier = bit x of h(block N-(StakeModifierDelay+x))
	StakeModifierDelay BlockHeight
	// BlockStakeAging is the amount of seconds to wait before a blockstake output
	// which is not on index 0 in the first transaction of a block can be used to
	// participate in the proof of blockstake protocol
	BlockStakeAging uint64
	// BlockCreatorFee is the amount of hastings you get for creating a block on top of
	// all the other rewards such as collected transaction fees.
	BlockCreatorFee Currency

	// OneCoin is the size of a "coin", expressed in hastings (the smallest unit of currency on the chain)
	OneCoin Currency

	// GenesisBlockStakeAllocation are the blockstake outputs of the genesis block
	GenesisBlockStakeAllocation []BlockStakeOutput
	// GenesisBlockStakeCount is the amount of blockstake created in the genesis block,
	// I.E. it is the sum of the Value's of GenesisBlockStakeAllocation
	GenesisBlockStakeCount Currency
	// GenesisCoinDistribution are the coin outputs of the genesis block
	GenesisCoinDistribution []CoinOutput
	// GenesisCoinCount is the amount of hastings created in the genesis block,
	// I.E. it is the sum of the Value's of GenesisCoinDistribution
	GenesisCoinCount Currency

	// GenesisBlock is the first block of the blockchain
	GenesisBlock Block
	// GenesisID is the block ID of the genesis block
	GenesisID BlockID

	// StartDifficulty is the initial difficulty for the proof of blockstake protocol
	// when the chain is started
	StartDifficulty Difficulty
}

// SetConstants sets the chain constants for the types package
func SetConstants(c ChainConstants) {
	cts = c
}

// DefaultChainConstants provide sane defaults for a new chain. Not all constants
// are set, since some (e.g. GenesisTimestamp) are chain specific, and this also
// allows some santiy checking later
// GenesisTimestamp, GenesisBlockStakeAllocation, and GenesisCoinDistribution aren't set as there is no such thing as a "sane default" for these variables
// since they are really chain specific
// Likewise, don't set RootTarget, GenesisBlockStakeCount, GenesisCoinCount, GenesisBlock, GenesisID, and StartDifficulty as these should be calculated
// By the Calculate method
func DefaultChainConstants() ChainConstants {
	cts := ChainConstants{
		BlockSizeLimit:         2e6,
		RootDepth:              Target{255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255},
		BlockFrequency:         600,
		MaturityDelay:          144,
		MedianTimestampWindow:  11,
		TargetWindow:           1e3,
		MaxAdjustmentUp:        big.NewRat(25, 10),
		MaxAdjustmentDown:      big.NewRat(10, 25),
		FutureThreshold:        3 * 60 * 60, // 3 hours.
		ExtremeFutureThreshold: 5 * 60 * 60, // 5 hours.
		StakeModifierDelay:     2000,
		BlockStakeAging:        1 << 17, // 2^16s < 1 day < 2^17s
		OneCoin:                NewCurrency(new(big.Int).Exp(big.NewInt(10), big.NewInt(24), nil)),
	}
	// BlockCreatorFee must be set here since it references the onecoin variable
	cts.BlockCreatorFee = cts.OneCoin.Mul64(10)
	return cts
}

// Calculate sets the GenesisBlock, GenesisID, GenesisBlockStakeCount and GenesisCoinCount.
// StartDifficulty and RootTarget are also set. If they need to be changed to be different (e.g.) for
// "pre-mining", this must be done after invoking this method
func (c *ChainConstants) Calculate() {
	// Create the genesis block.
	c.GenesisBlock = Block{
		Timestamp: c.GenesisTimestamp,
		Transactions: []Transaction{
			{
				BlockStakeOutputs: c.GenesisBlockStakeAllocation,
				CoinOutputs:       c.GenesisCoinDistribution,
			},
		},
	}
	// Calculate the genesis ID.
	c.GenesisID = c.GenesisBlock.ID()

	// Reset blockstake and currency count to avoid issues if this function is called twice by accident
	c.GenesisBlockStakeCount = ZeroCurrency
	c.GenesisCoinCount = ZeroCurrency
	for _, bso := range c.GenesisBlockStakeAllocation {
		c.GenesisBlockStakeCount = c.GenesisBlockStakeCount.Add(bso.Value)
	}
	for _, co := range c.GenesisCoinDistribution {
		c.GenesisCoinCount = c.GenesisCoinCount.Add(co.Value)
	}

	//Calculate start difficulty
	c.StartDifficulty = NewDifficulty(big.NewInt(0).Mul(big.NewInt(int64(c.BlockFrequency)), c.GenesisBlockStakeCount.Big()))
	// Add a check for a zero difficulty to avoid zero division. If the startDifficulty is zero, just
	// set it to something positive. It doesn't really matter what as there can be no block creation anyway
	// due to the lack of blockstake.
	if c.StartDifficulty.Cmp(Difficulty{}) == 0 {
		c.StartDifficulty = Difficulty{i: *big.NewInt(1)}
	}
	c.RootTarget = NewTargetWithDepth(c.StartDifficulty, c.RootDepth)
}

// CheckChainConstants does a sanity check on some of the constants to see if proper initialization is done
func (c *ChainConstants) CheckChainConstants() error {
	if c.GenesisCoinDistribution == nil || len(c.GenesisCoinDistribution) < 1 {
		return errors.New("Invalid genesis coin distribution")
	}
	if c.GenesisCoinCount.IsZero() {
		return errors.New("Invalid genesis coin count")
	}
	if c.GenesisBlockStakeAllocation == nil || len(c.GenesisBlockStakeAllocation) < 1 {
		return errors.New("Invalid genesis blockstake allocation")
	}
	if c.GenesisBlockStakeCount.IsZero() {
		return errors.New("Invalid genesis blockstake count")
	}
	// Genesis timestamp should not be too far in the past. The reference timestamp is the timestamp of the bitcoin genesis block,
	// as it's pretty safe to assume no blockchain was created before this (Saturday, January 3, 2009 6:15:05 PM GMT)
	if c.GenesisTimestamp < Timestamp(1231006505) {
		return errors.New("Invalid genesis timestamp")
	}
	if c.RootTarget == Target([32]byte{}) {
		return errors.New("Invalid root target")
	}
	return nil
}
