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

// Chain configuration variables, these should be injected with the help of the SetChainConfig function
var (
	BlockSizeLimit   = uint64(2e6)
	RootDepth        = Target{255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255}
	BlockFrequency   BlockHeight
	MaturityDelay    BlockHeight
	GenesisTimestamp Timestamp
	RootTarget       Target

	MedianTimestampWindow  = uint64(11)
	TargetWindow           BlockHeight
	MaxAdjustmentUp        *big.Rat
	MaxAdjustmentDown      *big.Rat
	FutureThreshold        Timestamp
	ExtremeFutureThreshold Timestamp

	StakeModifierDelay BlockHeight

	BlockStakeAging uint64

	BlockCreatorFee Currency

	OneCoin = NewCurrency(new(big.Int).Exp(big.NewInt(10), big.NewInt(24), nil))

	GenesisBlockStakeAllocation = []BlockStakeOutput{}
	GenesisBlockStakeCount      Currency
	GenesisCoinDistribution     = []CoinOutput{}
	GenesisCoinCount            Currency

	GenesisBlock Block

	// GenesisID is used in many places. Calculating it once saves lots of
	// redundant computation.
	GenesisID BlockID

	// StartDifficulty is used in many places. Calculate it once.
	StartDifficulty Difficulty
)

// ChainConstants is a utility struct which groups together the chain configuration
type ChainConstants struct {
	BlockSizeLimit   uint64
	RootDepth        Target
	BlockFrequency   BlockHeight
	MaturityDelay    BlockHeight
	GenesisTimestamp Timestamp
	RootTarget       Target

	MedianTimestampWindow uint64

	TargetWindow           BlockHeight
	MaxAdjustmentUp        *big.Rat
	MaxAdjustmentDown      *big.Rat
	FutureThreshold        Timestamp
	ExtremeFutureThreshold Timestamp

	StakeModifierDelay BlockHeight
	BlockStakeAging    uint64
	BlockCreatorFee    Currency

	OneCoin Currency

	GenesisBlockStakeAllocation []BlockStakeOutput
	GenesisBlockStakeCount      Currency
	GenesisCoinDistribution     []CoinOutput
	GenesisCoinCount            Currency

	GenesisBlock Block
	GenesisID    BlockID

	StartDifficulty Difficulty
}

// DefaultChainConstants provide sane defaults for a new chain. Not all constants
// are set, since some (e.g. GenesisTimestamp) are chain specific, and this also
// allows some santiy checking later
func DefaultChainConstants() ChainConstants {
	// GenesisTimestamp, GenesisBlockStakeAllocation, and GenesisCoinDistribution aren't set as there is no such thing as a "sane default" for these variables
	// since they are really chain specific
	// Likewise, don't set RootTarget, GenesisBlockStakeCount, GenesisCoinCount, GenesisBlock, GenesisID, and StartDifficulty as these should be calculated
	// By the Calculate method
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
	c.RootTarget = NewTargetWithDepth(c.StartDifficulty, c.RootDepth)
}

// SetChainConfig injects custom chain constants
func SetChainConfig(cfg ChainConstants) error {
	if err := checkChainConstants(cfg); err != nil {
		return err
	}

	BlockSizeLimit = cfg.BlockSizeLimit
	RootDepth = cfg.RootDepth
	BlockFrequency = cfg.BlockFrequency
	MaturityDelay = cfg.MaturityDelay
	GenesisTimestamp = cfg.GenesisTimestamp
	RootTarget = cfg.RootTarget
	MedianTimestampWindow = cfg.MedianTimestampWindow
	TargetWindow = cfg.TargetWindow
	MaxAdjustmentUp = cfg.MaxAdjustmentUp
	MaxAdjustmentDown = cfg.MaxAdjustmentDown
	FutureThreshold = cfg.FutureThreshold
	ExtremeFutureThreshold = cfg.ExtremeFutureThreshold
	StakeModifierDelay = cfg.StakeModifierDelay
	BlockStakeAging = cfg.BlockStakeAging
	BlockCreatorFee = cfg.BlockCreatorFee
	OneCoin = cfg.OneCoin
	GenesisBlockStakeAllocation = cfg.GenesisBlockStakeAllocation
	GenesisBlockStakeCount = cfg.GenesisBlockStakeCount
	GenesisCoinDistribution = cfg.GenesisCoinDistribution
	GenesisCoinCount = cfg.GenesisCoinCount
	GenesisBlock = cfg.GenesisBlock
	GenesisID = cfg.GenesisID
	StartDifficulty = cfg.StartDifficulty
	return nil
}

// checkChainConstants does a sanity check on some of the constants to see if proper initialization is done
func checkChainConstants(constants ChainConstants) error {
	if constants.GenesisCoinDistribution == nil || len(constants.GenesisCoinDistribution) < 1 {
		return errors.New("Invalid genesis coin distribution")
	}
	if constants.GenesisCoinCount.IsZero() {
		return errors.New("Invalid genesis coin count")
	}
	if constants.GenesisBlockStakeAllocation == nil || len(constants.GenesisBlockStakeAllocation) < 1 {
		return errors.New("Invalid genesis blockstake allocation")
	}
	if constants.GenesisBlockStakeCount.IsZero() {
		return errors.New("Invalid genesis blockstake count")
	}
	// Genesis timestamp should not be too far in the past. The reference timestamp is the timestamp of the bitcoin genesis block,
	// as it's pretty safe to assume no blockchain was created before this (Saturday, January 3, 2009 6:15:05 PM GMT)
	if constants.GenesisTimestamp < Timestamp(1231006505) {
		return errors.New("Invalid genesis timestamp")
	}
	if constants.RootTarget == Target([32]byte{}) {
		return errors.New("Invalid root target")
	}
	return nil
}
