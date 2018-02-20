package types

// constants.go contains the Sia constants. Depending on which build tags are
// used, the constants will be initialized to different values.
//
// CONTRIBUTE: We don't have way to check that the non-test constants are all
// sane, plus we have no coverage for them.

import (
	"math/big"

	"github.com/rivine/rivine/build"
)

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

// init checks which build constant is in place and initializes the variables
// accordingly.
func init() {

	if build.Release == "dev" {
		// 'dev' settings are for small developer testnets, usually on the same
		// computer. Settings are slow enough that a small team of developers
		// can coordinate their actions over a the developer testnets, but fast
		// enough that there isn't much time wasted on waiting for things to
		// happen.
		BlockFrequency = 12 // 12 seconds: slow enough for developers to see ~each block, fast enough that blocks don't waste time.
		MaturityDelay = 10  // 120 seconds before a delayed output matures.

		// Change as necessary. If not changed, the first few difficulty addaptions
		// will be wrong, but after some new difficulty calculations the error will
		// fade out.
		GenesisTimestamp = Timestamp(1424139000)

		TargetWindow = 20                        // Difficulty is adjusted based on prior 20 blocks.
		MaxAdjustmentUp = big.NewRat(120, 100)   // Difficulty adjusts quickly.
		MaxAdjustmentDown = big.NewRat(100, 120) // Difficulty adjusts quickly.
		FutureThreshold = 2 * 60                 // 2 minutes.
		ExtremeFutureThreshold = 4 * 60          // 4 minutes.
		StakeModifierDelay = 2000                // Number of blocks to take in history to calculate the stakemodifier

		BlockStakeAging = uint64(1 << 10) // Block stake aging if unspent block stake is not at index 0

		BlockCreatorFee = OneCoin.Mul64(100)

		bso := BlockStakeOutput{
			Value:      NewCurrency64(1000000),
			UnlockHash: UnlockHash{},
		}

		co := CoinOutput{
			Value: OneCoin.Mul64(1000),
		}

		// Seed for this address:
		// across knife thirsty puck itches hazard enmity fainted pebbles unzip echo queen rarest aphid bugs yanks okay abbey eskimos dove orange nouns august ailments inline rebel glass tyrant acumen
		bso.UnlockHash.LoadString("e66bbe9638ae0e998641dc9faa0180c15a1071b1767784cdda11ad3c1d309fa692667931be66")
		GenesisBlockStakeAllocation = append(GenesisBlockStakeAllocation, bso)
		co.UnlockHash.LoadString("e66bbe9638ae0e998641dc9faa0180c15a1071b1767784cdda11ad3c1d309fa692667931be66")
		GenesisCoinDistribution = append(GenesisCoinDistribution, co)

	} else if build.Release == "testing" {
		// 'testing' settings are for automatic testing, and create much faster
		// environments than a human can interact with.
		BlockFrequency = 1 // As fast as possible
		MaturityDelay = 3
		GenesisTimestamp = CurrentTimestamp() - 1e6
		RootTarget = Target{128} // Takes an expected 2 hashes; very fast for testing but still probes 'bad hash' code.

		// A restrictive difficulty clamp prevents the difficulty from climbing
		// during testing, as the resolution on the difficulty adjustment is
		// only 1 second and testing mining should be happening substantially
		// faster than that.
		TargetWindow = 200
		MaxAdjustmentUp = big.NewRat(10001, 10000)
		MaxAdjustmentDown = big.NewRat(9999, 10000)
		FutureThreshold = 3        // 3 seconds
		ExtremeFutureThreshold = 6 // 6 seconds
		StakeModifierDelay = 20

		BlockStakeAging = uint64(1 << 10)

		BlockCreatorFee = OneCoin.Mul64(100)

		GenesisBlockStakeAllocation = []BlockStakeOutput{
			{
				Value:      NewCurrency64(2000),
				UnlockHash: UnlockHash{214, 166, 197, 164, 29, 201, 53, 236, 106, 239, 10, 158, 127, 131, 20, 138, 63, 221, 230, 16, 98, 247, 32, 77, 210, 68, 116, 12, 241, 89, 27, 223},
			},
			{
				Value:      NewCurrency64(7000),
				UnlockHash: UnlockHash{209, 246, 228, 60, 248, 78, 242, 110, 9, 8, 227, 248, 225, 216, 163, 52, 142, 93, 47, 176, 103, 41, 137, 80, 212, 8, 132, 58, 241, 189, 2, 17},
			},
			{
				Value:      NewCurrency64(1000),
				UnlockHash: UnlockConditions{}.UnlockHash(),
			},
		}
	} else if build.Release == "standard" {
		// 'standard' settings are for the full network. They are slow enough
		// that the network is secure in a real-world byzantine environment.

		// A block time of 1 block per 10 minutes is chosen to follow Bitcoin's
		// example. The security lost by lowering the block time is not
		// insignificant, and the convenience gained by lowering the blocktime
		// even down to 90 seconds is not significant. I do feel that 10
		// minutes could even be too short, but it has worked well for Bitcoin.
		BlockFrequency = 600

		// Payouts take 1 day to mature. This is to prevent a class of double
		// spending attacks parties unintentionally spend coins that will stop
		// existing after a blockchain reorganization. There are multiple
		// classes of payouts in Sia that depend on a previous block - if that
		// block changes, then the output changes and the previously existing
		// output ceases to exist. This delay stops both unintentional double
		// spending and stops a small set of long-range mining attacks.
		MaturityDelay = 144

		// The genesis timestamp is set to June 1st, 2017
		GenesisTimestamp = Timestamp(1496322000) // June 2nd, 2017 @ 1:00pm UTC.

		// The RootTarget was set such that the developers could reasonable
		// premine 100 blocks in a day. It was known to the developrs at launch
		// this this was at least one and perhaps two orders of magnitude too
		// small.
		RootTarget = Target{0, 0, 0, 0, 32}

		// When the difficulty is adjusted, it is adjusted by looking at the
		// timestamp of the 1000th previous block. This minimizes the abilities
		// of miners to attack the network using rogue timestamps.
		TargetWindow = 1e3

		// The difficutly adjustment is clamped to 2.5x every 500 blocks. This
		// corresponds to 6.25x every 2 weeks, which can be compared to
		// Bitcoin's clamp of 4x every 2 weeks. The difficulty clamp is
		// primarily to stop difficulty raising attacks. Sia's safety margin is
		// similar to Bitcoin's despite the looser clamp because Sia's
		// difficulty is adjusted four times as often. This does result in
		// greater difficulty oscillation, a tradeoff that was chosen to be
		// acceptable due to Sia's more vulnerable position as an altcoin.
		MaxAdjustmentUp = big.NewRat(25, 10)
		MaxAdjustmentDown = big.NewRat(10, 25)

		// Blocks will not be accepted if their timestamp is more than 3 hours
		// into the future, but will be accepted as soon as they are no longer
		// 3 hours into the future. Blocks that are greater than 5 hours into
		// the future are rejected outright, as it is assumed that by the time
		// 2 hours have passed, those blocks will no longer be on the longest
		// chain. Blocks cannot be kept forever because this opens a DoS
		// vector.
		FutureThreshold = 3 * 60 * 60        // 3 hours.
		ExtremeFutureThreshold = 5 * 60 * 60 // 5 hours.

		// The stakemodifier is calculated from blocks in history. The stakemodifier
		// is calculated as: For x = 0 to 255
		// bit x of Stake Modifier = bit x of h(block N-(StakeModifierDelay+x))
		StakeModifierDelay = 2000

		// Blockstakeaging is the number of seconds to wait before blockstake can be
		// used to solve blocks. But only when the block stake output is not the
		// first transaction with the first index. (2^16s < 1 day < 2^17s)
		BlockStakeAging = uint64(1 << 17)

		// BlockCreatorFee is the asset you get when creating a block on top of the
		// other fee.
		BlockCreatorFee = OneCoin.Mul64(10)

		bso := BlockStakeOutput{
			Value:      NewCurrency64(1000000),
			UnlockHash: UnlockHash{},
		}

		co := CoinOutput{
			Value: OneCoin.Mul64(100 * 1000 * 1000),
		}

		bso.UnlockHash.LoadString("b5e42056ef394f2ad9b511a61cec874d25bebe2095682dd37455cbafed4bec15c28ee7d7ed1d")
		GenesisBlockStakeAllocation = append(GenesisBlockStakeAllocation, bso)
		co.UnlockHash.LoadString("b5e42056ef394f2ad9b511a61cec874d25bebe2095682dd37455cbafed4bec15c28ee7d7ed1d")
		GenesisCoinDistribution = append(GenesisCoinDistribution, co)
	}

	CalculateGenesis()
}

// CalculateGenesis fills in the genesis block variables which are computed based on
// variables that have been set earlier
func CalculateGenesis() {
	// Create the genesis block.
	GenesisBlock = Block{
		Timestamp: GenesisTimestamp,
		Transactions: []Transaction{
			{
				BlockStakeOutputs: GenesisBlockStakeAllocation,
				CoinOutputs:       GenesisCoinDistribution,
			},
		},
	}
	// Calculate the genesis ID.
	GenesisID = GenesisBlock.ID()

	for _, bso := range GenesisBlockStakeAllocation {
		GenesisBlockStakeCount = GenesisBlockStakeCount.Add(bso.Value)
	}
	for _, co := range GenesisCoinDistribution {
		GenesisCoinCount = GenesisCoinCount.Add(co.Value)
	}

	//Calculate start difficulty
	StartDifficulty = NewDifficulty(big.NewInt(0).Mul(big.NewInt(int64(BlockFrequency)), GenesisBlockStakeCount.Big()))
	RootTarget = NewTarget(StartDifficulty)

}
