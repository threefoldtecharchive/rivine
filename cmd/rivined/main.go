package main

import (
	"math/big"

	"github.com/rivine/rivine/modules"
	"github.com/rivine/rivine/pkg/daemon"

	"github.com/rivine/rivine/build"
	"github.com/rivine/rivine/types"
)

// main establishes a set of commands and flags using the cobra package.
func main() {
	// Use the default daemon configuration
	cfg := daemon.DefaultConfig()
	// register the default networks we use
	cfg.Networks = registerNetworks()
	// Try to connect to the right network by default without having to pass flags around
	cfg.NetworkName = getDefaultNetworkForBuild()
	daemon.SetupDefaultDaemon(cfg)
}

// getDefaultNetworkForBuild returns the default network name based on the build tags
func getDefaultNetworkForBuild() string {
	return build.Release
}

// registerNetworks registers all the networks
func registerNetworks() map[string]daemon.NetworkConfig {
	netCfgs := make(map[string]daemon.NetworkConfig)
	netCfgs["dev"] = registerDevNet()
	netCfgs["testing"] = registerTestingNet()
	netCfgs["standard"] = registerStandardNet()

	return netCfgs
}

// registerDevNet registers the "dev" network
func registerDevNet() daemon.NetworkConfig {
	// 'dev' settings are for small developer testnets, usually on the same
	// computer. Settings are slow enough that a small team of developers
	// can coordinate their actions over a the developer testnets, but fast
	// enough that there isn't much time wasted on waiting for things to
	// happen.
	cts := types.DefaultChainConstants()
	cts.BlockFrequency = 12 // 12 seconds: slow enough for developers to see ~each block, fast enough that blocks don't waste time.
	cts.MaturityDelay = 10  // 120 seconds before a delayed output matures.

	// Change as necessary. If not changed, the first few difficulty addaptions
	// will be wrong, but after some new difficulty calculations the error will
	// fade out.
	cts.GenesisTimestamp = types.Timestamp(1424139000)

	cts.TargetWindow = 20                        // Difficulty is adjusted based on prior 20 blocks.
	cts.MaxAdjustmentUp = big.NewRat(120, 100)   // Difficulty adjusts quickly.
	cts.MaxAdjustmentDown = big.NewRat(100, 120) // Difficulty adjusts quickly.
	cts.FutureThreshold = 2 * 60                 // 2 minutes.
	cts.ExtremeFutureThreshold = 4 * 60          // 4 minutes.
	cts.StakeModifierDelay = 2000                // Number of blocks to take in history to calculate the stakemodifier

	cts.BlockStakeAging = uint64(1 << 10) // Block stake aging if unspent block stake is not at index 0

	// OneCoin must reference the OneCoin variable set in the config or it'll resolve against the OneCoin before it is updated
	cts.BlockCreatorFee = cts.OneCoin.Mul64(100)

	bso := types.BlockStakeOutput{
		Value:      types.NewCurrency64(1000000),
		UnlockHash: types.UnlockHash{},
	}

	co := types.CoinOutput{
		Value: cts.OneCoin.Mul64(1000),
	}

	// Seed for this address:
	// across knife thirsty puck itches hazard enmity fainted pebbles unzip echo queen rarest aphid bugs yanks okay abbey eskimos dove orange nouns august ailments inline rebel glass tyrant acumen
	bso.UnlockHash.LoadString("e66bbe9638ae0e998641dc9faa0180c15a1071b1767784cdda11ad3c1d309fa692667931be66")
	cts.GenesisBlockStakeAllocation = append(cts.GenesisBlockStakeAllocation, bso)
	co.UnlockHash.LoadString("e66bbe9638ae0e998641dc9faa0180c15a1071b1767784cdda11ad3c1d309fa692667931be66")
	cts.GenesisCoinDistribution = append(cts.GenesisCoinDistribution, co)

	// Calculate the remaining params
	cts.Calculate()

	devNet := daemon.NetworkConfig{
		BootstrapPeers: nil,
		Constants:      cts,
	}
	// register the network
	return devNet
}

// registerTestingNet registers the "testing" network
func registerTestingNet() daemon.NetworkConfig {
	// 'testing' settings are for automatic testing, and create much faster
	//  environments than a human can interact with.
	cts := types.DefaultChainConstants()
	cts.BlockFrequency = 1 // As fast as possible
	cts.MaturityDelay = 3
	cts.GenesisTimestamp = types.CurrentTimestamp() - 1e6

	// A restrictive difficulty clamp prevents the difficulty from climbing
	// during testing, as the resolution on the difficulty adjustment is
	// only 1 second and testing mining should be happening substantially
	// faster than that.
	cts.TargetWindow = 200
	cts.MaxAdjustmentUp = big.NewRat(10001, 10000)
	cts.MaxAdjustmentDown = big.NewRat(9999, 10000)
	cts.FutureThreshold = 3        // 3 seconds
	cts.ExtremeFutureThreshold = 6 // 6 seconds
	cts.StakeModifierDelay = 20

	cts.BlockStakeAging = uint64(1 << 10)

	cts.BlockCreatorFee = cts.OneCoin.Mul64(100)

	cts.GenesisBlockStakeAllocation = []types.BlockStakeOutput{
		{
			Value:      types.NewCurrency64(2000),
			UnlockHash: types.UnlockHash{214, 166, 197, 164, 29, 201, 53, 236, 106, 239, 10, 158, 127, 131, 20, 138, 63, 221, 230, 16, 98, 247, 32, 77, 210, 68, 116, 12, 241, 89, 27, 223},
		},
		{
			Value:      types.NewCurrency64(7000),
			UnlockHash: types.UnlockHash{209, 246, 228, 60, 248, 78, 242, 110, 9, 8, 227, 248, 225, 216, 163, 52, 142, 93, 47, 176, 103, 41, 137, 80, 212, 8, 132, 58, 241, 189, 2, 17},
		},
		{
			Value:      types.NewCurrency64(1000),
			UnlockHash: types.UnlockConditions{}.UnlockHash(),
		},
	}

	cts.GenesisCoinDistribution = []types.CoinOutput{
		{
			Value:      cts.OneCoin.Mul64(1000),
			UnlockHash: types.UnlockHash{214, 166, 197, 164, 29, 201, 53, 236, 106, 239, 10, 158, 127, 131, 20, 138, 63, 221, 230, 16, 98, 247, 32, 77, 210, 68, 116, 12, 241, 89, 27, 223},
		},
	}

	cts.Calculate()

	// Set the root target after calculating the remainder of the constants so its not overwritten
	cts.RootTarget = types.Target{128} // Takes an expected 2 hashes; very fast for testing but still probes 'bad hash' code.

	testingNet := daemon.NetworkConfig{
		BootstrapPeers: nil,
		Constants:      cts,
	}
	// register the network
	return testingNet
}

// registerStandardNet registers the "standard" network
func registerStandardNet() daemon.NetworkConfig {
	// 'standard' settings are for the full network. They are slow enough
	// that the network is secure in a real-world byzantine environment.
	cts := types.DefaultChainConstants()
	// A block time of 1 block per 10 minutes is chosen to follow Bitcoin's
	// example. The security lost by lowering the block time is not
	// insignificant, and the convenience gained by lowering the blocktime
	// even down to 90 seconds is not significant. I do feel that 10
	// minutes could even be too short, but it has worked well for Bitcoin.
	cts.BlockFrequency = 600

	// Payouts take 1 day to mature. This is to prevent a class of double
	// spending attacks parties unintentionally spend coins that will stop
	// existing after a blockchain reorganization. There are multiple
	// classes of payouts in Sia that depend on a previous block - if that
	// block changes, then the output changes and the previously existing
	// output ceases to exist. This delay stops both unintentional double
	// spending and stops a small set of long-range mining attacks.
	cts.MaturityDelay = 144

	// The genesis timestamp is set to June 1st, 2017
	cts.GenesisTimestamp = types.Timestamp(1496322000) // June 2nd, 2017 @ 1:00pm UTC.

	// When the difficulty is adjusted, it is adjusted by looking at the
	// timestamp of the 1000th previous block. This minimizes the abilities
	// of miners to attack the network using rogue timestamps.
	cts.TargetWindow = 1e3

	// The difficutly adjustment is clamped to 2.5x every 500 blocks. This
	// corresponds to 6.25x every 2 weeks, which can be compared to
	// Bitcoin's clamp of 4x every 2 weeks. The difficulty clamp is
	// primarily to stop difficulty raising attacks. Sia's safety margin is
	// similar to Bitcoin's despite the looser clamp because Sia's
	// difficulty is adjusted four times as often. This does result in
	// greater difficulty oscillation, a tradeoff that was chosen to be
	// acceptable due to Sia's more vulnerable position as an altcoin.
	cts.MaxAdjustmentUp = big.NewRat(25, 10)
	cts.MaxAdjustmentDown = big.NewRat(10, 25)

	// Blocks will not be accepted if their timestamp is more than 3 hours
	// into the future, but will be accepted as soon as they are no longer
	// 3 hours into the future. Blocks that are greater than 5 hours into
	// the future are rejected outright, as it is assumed that by the time
	// 2 hours have passed, those blocks will no longer be on the longest
	// chain. Blocks cannot be kept forever because this opens a DoS
	// vector.
	cts.FutureThreshold = 3 * 60 * 60        // 3 hours.
	cts.ExtremeFutureThreshold = 5 * 60 * 60 // 5 hours.

	// The stakemodifier is calculated from blocks in history. The stakemodifier
	// is calculated as: For x = 0 to 255
	// bit x of Stake Modifier = bit x of h(block N-(StakeModifierDelay+x))
	cts.StakeModifierDelay = 2000

	// Blockstakeaging is the number of seconds to wait before blockstake can be
	// used to solve blocks. But only when the block stake output is not the
	// first transaction with the first index. (2^16s < 1 day < 2^17s)
	cts.BlockStakeAging = uint64(1 << 17)

	// BlockCreatorFee is the asset you get when creating a block on top of the
	// other fee.
	cts.BlockCreatorFee = cts.OneCoin.Mul64(10)

	bso := types.BlockStakeOutput{
		Value:      types.NewCurrency64(1000000),
		UnlockHash: types.UnlockHash{},
	}

	co := types.CoinOutput{
		Value: cts.OneCoin.Mul64(100 * 1000 * 1000),
	}

	cts.GenesisBlockStakeAllocation = []types.BlockStakeOutput{}
	cts.GenesisCoinDistribution = []types.CoinOutput{}

	bso.UnlockHash.LoadString("b5e42056ef394f2ad9b511a61cec874d25bebe2095682dd37455cbafed4bec15c28ee7d7ed1d")
	cts.GenesisBlockStakeAllocation = append(cts.GenesisBlockStakeAllocation, bso)
	co.UnlockHash.LoadString("b5e42056ef394f2ad9b511a61cec874d25bebe2095682dd37455cbafed4bec15c28ee7d7ed1d")
	cts.GenesisCoinDistribution = append(cts.GenesisCoinDistribution, co)

	cts.Calculate()
	// The RootTarget was set such that the developers could reasonable
	// premine 100 blocks in a day. It was known to the developrs at launch
	// this this was at least one and perhaps two orders of magnitude too
	// small.
	cts.RootTarget = types.Target{0, 0, 0, 0, 32}

	standardNet := daemon.NetworkConfig{
		BootstrapPeers: []modules.NetAddress{
			"136.243.144.132:23112",
			"[2a01:4f8:171:1303::2]:23112",
			"bootstrap2.rivine.io:23112",
			"bootstrap3.rivine.io:23112",
		},
		Constants: cts,
	}
	// register the network
	return standardNet
}
