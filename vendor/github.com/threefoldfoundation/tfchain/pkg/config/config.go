package config

import (
	"fmt"
	"math/big"

	"github.com/rivine/rivine/build"
	"github.com/rivine/rivine/modules"
	"github.com/rivine/rivine/types"
)

var (
	rawVersion = "v1.1.1"
	// Version of the tfchain binaries.
	//
	// Value is defined by a private build flag,
	// or hardcoded to the latest released tag as fallback.
	Version build.ProtocolVersion
)

const (
	// ThreeFoldTokenUnit defines the unit of one ThreeFold Token.
	ThreeFoldTokenUnit = "TFT"
	// ThreeFoldTokenChainName defines the name of the ThreeFoldToken chain.
	ThreeFoldTokenChainName = "tfchain"
)

// chain names
const (
	NetworkNameStandard = "standard"
	NetworkNameTest     = "testnet"
	NetworkNameDev      = "devnet"
)

// global network config constants
const (
	StandardNetworkBlockFrequency types.BlockHeight = 120 // 1 block per 2 minutes on average
	TestNetworkBlockFrequency     types.BlockHeight = 120 // 1 block per 2 minutes on average
)

// GetCurrencyUnits returns the currency units used for all ThreeFold networks.
func GetCurrencyUnits() types.CurrencyUnits {
	return types.CurrencyUnits{
		// 1 coin = 1 000 000 000 of the smalles possible units
		OneCoin: types.NewCurrency(new(big.Int).Exp(big.NewInt(10), big.NewInt(9), nil)),
	}
}

// GetBlockchainInfo returns the naming and versioning of tfchain.
func GetBlockchainInfo() types.BlockchainInfo {
	return types.BlockchainInfo{
		Name:            ThreeFoldTokenChainName,
		NetworkName:     NetworkNameStandard,
		CoinUnit:        ThreeFoldTokenUnit,
		ChainVersion:    Version,       // use our own blockChain/build version
		ProtocolVersion: build.Version, // use latest available rivine protocol version
	}
}

// GetStandardnetGenesisMintCondition returns the genesis mint condition used for the standard (prod) net
func GetStandardnetGenesisMintCondition() types.UnlockConditionProxy {
	return types.NewCondition(types.NewMultiSignatureCondition(types.UnlockHashSlice{
		unlockHashFromHex("01434535fd01243c02c277cd58d71423163767a575a8ae44e15807bf545e4a8456a5c4afabad51"),
		unlockHashFromHex("01334cf68f312026ff9df84fc023558db8624bedd717adcc9edc6900488cf6df54ac8e3d1c89a8"),
		unlockHashFromHex("0149a5496fea27315b7db6251e5dfda23bc9d4bf677c5a5c2d70f1382c44357197d8453d9dfa32"),
	}, 2))
}

// GetStandardnetGenesis explicitly sets all the required constants for the genesis block of the standard (prod) net
func GetStandardnetGenesis() types.ChainConstants {
	cfg := types.StandardnetChainConstants()

	// use the threefold currency units
	cfg.CurrencyUnits = GetCurrencyUnits()

	// set transaction versions
	cfg.DefaultTransactionVersion = types.TransactionVersionOne
	cfg.GenesisTransactionVersion = types.TransactionVersionZero

	// 2 minute block time
	cfg.BlockFrequency = StandardNetworkBlockFrequency

	// Payouts take roughly 1 day to mature.
	cfg.MaturityDelay = 720

	// The genesis timestamp
	cfg.GenesisTimestamp = types.Timestamp(1522501000) // Human time 03/31/2018 @ 1:03pm (UTC)

	// 1000 block window for difficulty
	cfg.TargetWindow = 1e3

	cfg.MaxAdjustmentUp = big.NewRat(25, 10)
	cfg.MaxAdjustmentDown = big.NewRat(10, 25)

	cfg.FutureThreshold = 1 * 60 * 60        // 1 hour.
	cfg.ExtremeFutureThreshold = 2 * 60 * 60 // 2 hours.

	cfg.StakeModifierDelay = 2000

	// Blockstake can be used roughly 1 day after receiving
	cfg.BlockStakeAging = 1 << 17 // 2^16s < 1 day < 2^17s

	// Receive 1 coins when you create a block
	cfg.BlockCreatorFee = cfg.CurrencyUnits.OneCoin.Mul64(1)

	// Use 0.1 coins as minimum transaction fee
	cfg.MinimumTransactionFee = cfg.CurrencyUnits.OneCoin.Div64(10)

	// Threefold Foundation receive all transactions fees in a single pool address,
	// Block Creation Fees do well still go to the block creator creating the block.
	cfg.TransactionFeeCondition = types.NewCondition(types.NewUnlockHashCondition(unlockHashFromHex(
		"017267221ef1947bb18506e390f1f9446b995acfb6d08d8e39508bb974d9830b8cb8fdca788e34")))

	// distribute initial coins
	cfg.GenesisCoinDistribution = []types.CoinOutput{
		{
			// 695M TFT for Capacity availability until 01/04
			Value: cfg.CurrencyUnits.OneCoin.Mul64(695 * 1000 * 1000),
			// temporary pool
			Condition: types.NewCondition(types.NewUnlockHashCondition(unlockHashFromHex("014ab1cf49a331bef9225a51a68623daf7e112fce0e81a91194a7f4fe7af1d9a793bc52d4676d0"))),
		},
		{
			// 3K TFT for dev/test purposes
			Value: cfg.CurrencyUnits.OneCoin.Mul64(3000),
			// @glendc
			Condition: types.NewCondition(types.NewUnlockHashCondition(unlockHashFromHex("01ad4f73417476f8b8350298681dd0fa8640baa53a91915417b1dd8103d118b543c992e6fba1c4"))),
		},
		{
			// 90K TFT for dev/test purposes
			Value: cfg.CurrencyUnits.OneCoin.Mul64(90000),
			// @foundation
			Condition: types.NewCondition(types.NewUnlockHashCondition(unlockHashFromHex("017eb744c7a97443d7927725bfb2a004384e4230386ea0f693f9ce1c1161d771878a1f048c887b"))),
		},
		{
			// 1K TFT for dev/test purposes
			Value: cfg.CurrencyUnits.OneCoin.Mul64(1000),
			// @robvanmieghem
			Condition: types.NewCondition(types.NewUnlockHashCondition(unlockHashFromHex("01cc55df18eb3b86670deb6cfbb9b62b8463b62738426f0c14a7ae8926d6b556fbac3aab17f437"))),
		},
		{
			// 1K TFT for dev/test purposes
			Value: cfg.CurrencyUnits.OneCoin.Mul64(1000),
			// @RubenMattan
			Condition: types.NewCondition(types.NewUnlockHashCondition(unlockHashFromHex("01e2dc01a686fc0ca25612871a6515f2e3b804aa63244bf19449ecb3c9aaaf36f0cc6839b0af60"))),
		},
		{
			// 1K TFT for dev/test purposes
			Value: cfg.CurrencyUnits.OneCoin.Mul64(1000),
			// @FastGeert
			Condition: types.NewCondition(types.NewUnlockHashCondition(unlockHashFromHex("0166da8a4ab39a621637d9e7eb4e1fbaf95f905856af13af1268fc1a79c65b4f6686dec75c9d94"))),
		},
		{
			// 1K TFT for dev/test purposes
			Value: cfg.CurrencyUnits.OneCoin.Mul64(1000),
			// @zaibon
			Condition: types.NewCondition(types.NewUnlockHashCondition(unlockHashFromHex("013dfb1c49e8b9a73bc8b460d9ef20fc1f40e0d034742950f70d983c455342719d1e9f656d002b"))),
		},
		{
			// 2K TFT for dev/test purposes
			Value: cfg.CurrencyUnits.OneCoin.Mul64(2000),
			// @leesmet
			Condition: types.NewCondition(types.NewUnlockHashCondition(unlockHashFromHex("018a28615b277eb7e7a0e6921e85ad5b3ca378ac210b7f258b0b11ef313ea2ce98bd2e2510472d"))),
		},
	}

	// allocate block stakes
	cfg.GenesisBlockStakeAllocation = []types.BlockStakeOutput{
		{
			// 390 BS, initially allocated to a selected group of people (ambassadors)
			Value: types.NewCurrency64(390),
			// Pool address for other people (ambassadors)
			Condition: types.NewCondition(types.NewUnlockHashCondition(unlockHashFromHex("01746677df456546d93729066dd88514e2009930f3eebac3c93d43c88a108f8f9aa9e7c6f58893"))),
		},
		{
			// 100 BS, one BS for each first-generation TFT node
			Value: types.NewCurrency64(100),
			// @glendc (temporary)
			Condition: types.NewCondition(types.NewUnlockHashCondition(unlockHashFromHex("01ad4f73417476f8b8350298681dd0fa8640baa53a91915417b1dd8103d118b543c992e6fba1c4"))),
		},
		{
			// 10 BS, for dev/validation/test purposes
			Value: types.NewCurrency64(10),
			// @foundation @robvanmieghem
			Condition: types.NewCondition(types.NewUnlockHashCondition(unlockHashFromHex("01cc55df18eb3b86670deb6cfbb9b62b8463b62738426f0c14a7ae8926d6b556fbac3aab17f437"))),
		},
	}

	return cfg
}

// GetTestnetGenesisMintCondition returns the genesis mint condition used for the testnet
func GetTestnetGenesisMintCondition() types.UnlockConditionProxy {
	return types.NewCondition(types.NewMultiSignatureCondition(types.UnlockHashSlice{
		unlockHashFromHex("016148ac9b17828e0933796eaca94418a376f2aa3fefa15685cea5fa462093f0150e09067f7512"),
		unlockHashFromHex("01d553fab496f3fd6092e25ce60e6f72e24b57950bffc0d372d659e38e5a95e89fb117b4eb3481"),
		unlockHashFromHex("013a787bf6248c518aee3a040a14b0dd3a029bc8e9b19a1823faf5bcdde397f4201ad01aace4c9"),
	}, 2))
}

// GetTestnetGenesis explicitly sets all the required constants for the genesis block of the testnet
func GetTestnetGenesis() types.ChainConstants {
	cfg := types.TestnetChainConstants()

	// use the threefold currency units
	cfg.CurrencyUnits = GetCurrencyUnits()

	// set transaction versions
	cfg.DefaultTransactionVersion = types.TransactionVersionOne
	cfg.GenesisTransactionVersion = types.TransactionVersionZero

	// 2 minute block time
	cfg.BlockFrequency = TestNetworkBlockFrequency

	// Payouts take rougly 1 day to mature.
	cfg.MaturityDelay = 720

	// The genesis timestamp is set to February 21st, 2018
	cfg.GenesisTimestamp = types.Timestamp(1519200000) // February 21st, 2018 @ 8:00am UTC.

	// 1000 block window for difficulty
	cfg.TargetWindow = 1e3

	cfg.MaxAdjustmentUp = big.NewRat(25, 10)
	cfg.MaxAdjustmentDown = big.NewRat(10, 25)

	cfg.FutureThreshold = 1 * 60 * 60        // 1 hour.
	cfg.ExtremeFutureThreshold = 2 * 60 * 60 // 2 hours.

	cfg.StakeModifierDelay = 2000

	// Blockstake can be used roughly 1 minute after receiving
	cfg.BlockStakeAging = uint64(1 << 6)

	// Receive 10 coins when you create a block
	cfg.BlockCreatorFee = cfg.CurrencyUnits.OneCoin.Mul64(10)

	// Use 0.1 coins as minimum transaction fee
	cfg.MinimumTransactionFee = cfg.CurrencyUnits.OneCoin.Div64(10)

	// distribute initial coins
	cfg.GenesisCoinDistribution = []types.CoinOutput{
		{
			// Create 100M coins
			Value: cfg.CurrencyUnits.OneCoin.Mul64(100 * 1000 * 1000),
			// @leesmet
			Condition: types.NewCondition(types.NewUnlockHashCondition(unlockHashFromHex("01fc8714235d549f890f35e52d745b9eeeee34926f96c4b9ef1689832f338d9349b453898f7e51"))),
		},
	}

	// allocate block stakes
	cfg.GenesisBlockStakeAllocation = []types.BlockStakeOutput{
		{
			// Create 3K blockstakes
			Value: types.NewCurrency64(3000),
			// @leesmet
			Condition: types.NewCondition(types.NewUnlockHashCondition(unlockHashFromHex("01fc8714235d549f890f35e52d745b9eeeee34926f96c4b9ef1689832f338d9349b453898f7e51"))),
		},
	}

	return cfg
}

// GetDevnetGenesisMintCondition returns the genesis mint condition used for the devnet
func GetDevnetGenesisMintCondition() types.UnlockConditionProxy {
	// belongs to wallet with mnemonic:
	// carbon boss inject cover mountain fetch fiber fit tornado cloth wing dinosaur proof joy intact fabric thumb rebel borrow poet chair network expire else
	return types.NewCondition(types.NewUnlockHashCondition(
		unlockHashFromHex("015a080a9259b9d4aaa550e2156f49b1a79a64c7ea463d810d4493e8242e6791584fbdac553e6f")))
}

// GetDevnetGenesis explicitly sets all the required constants for the genesis block of the devnet
func GetDevnetGenesis() types.ChainConstants {
	cfg := types.DevnetChainConstants()

	// use the threefold currency units
	cfg.CurrencyUnits = GetCurrencyUnits()

	// set transaction versions
	cfg.DefaultTransactionVersion = types.TransactionVersionOne
	// no need to keep v0 as genesis transaction version for the dev network
	cfg.GenesisTransactionVersion = types.TransactionVersionOne

	// 12 seconds, slow enough for developers to see
	// ~each block, fast enough that blocks don't waste time
	cfg.BlockFrequency = 12

	// 120 seconds before a delayed output matters
	// as it's expressed in units of blocks
	cfg.MaturityDelay = 10
	cfg.MedianTimestampWindow = 11

	// The genesis timestamp is set to February 21st, 2018
	cfg.GenesisTimestamp = types.Timestamp(1519200000) // February 21st, 2018 @ 8:00am UTC.

	// difficulity is adjusted based on prior 20 blocks
	cfg.TargetWindow = 20

	// Difficulty adjusts quickly.
	cfg.MaxAdjustmentUp = big.NewRat(120, 100)
	cfg.MaxAdjustmentDown = big.NewRat(100, 120)

	cfg.FutureThreshold = 2 * 60        // 2 minutes
	cfg.ExtremeFutureThreshold = 4 * 60 // 4 minutes

	cfg.StakeModifierDelay = 2000

	// Blockstake can be used roughly 1 minute after receiving
	cfg.BlockStakeAging = uint64(1 << 6)

	// Receive 10 coins when you create a block
	cfg.BlockCreatorFee = cfg.CurrencyUnits.OneCoin.Mul64(10)

	// Use 0.1 coins as minimum transaction fee
	cfg.MinimumTransactionFee = cfg.CurrencyUnits.OneCoin.Mul64(1)

	// distribute initial coins
	cfg.GenesisCoinDistribution = []types.CoinOutput{
		{
			// Create 100M coins
			Value: cfg.CurrencyUnits.OneCoin.Mul64(100 * 1000 * 1000),
			// belong to wallet with mnemonic:
			// carbon boss inject cover mountain fetch fiber fit tornado cloth wing dinosaur proof joy intact fabric thumb rebel borrow poet chair network expire else
			Condition: types.NewCondition(types.NewUnlockHashCondition(unlockHashFromHex("015a080a9259b9d4aaa550e2156f49b1a79a64c7ea463d810d4493e8242e6791584fbdac553e6f"))),
		},
	}

	// allocate block stakes
	cfg.GenesisBlockStakeAllocation = []types.BlockStakeOutput{
		{
			// Create 3K blockstakes
			Value: types.NewCurrency64(3000),
			// belongs to wallet with mnemonic:
			// carbon boss inject cover mountain fetch fiber fit tornado cloth wing dinosaur proof joy intact fabric thumb rebel borrow poet chair network expire else
			Condition: types.NewCondition(types.NewUnlockHashCondition(unlockHashFromHex("015a080a9259b9d4aaa550e2156f49b1a79a64c7ea463d810d4493e8242e6791584fbdac553e6f"))),
		},
	}

	return cfg
}

// GetStandardnetBootstrapPeers sets the standard bootstrap node addresses
func GetStandardnetBootstrapPeers() []modules.NetAddress {
	return []modules.NetAddress{
		"bootstrap1.threefoldtoken.com:23112",
		"bootstrap2.threefoldtoken.com:23112",
		"bootstrap3.threefoldtoken.com:23112",
		"bootstrap4.threefoldtoken.com:23112",
		"bootstrap5.threefoldtoken.com:23112",
	}
}

// GetTestnetBootstrapPeers sets the testnet bootstrap node addresses
func GetTestnetBootstrapPeers() []modules.NetAddress {
	return []modules.NetAddress{
		"bootstrap1.testnet.threefoldtoken.com:23112",
		"bootstrap2.testnet.threefoldtoken.com:23112",
		"bootstrap3.testnet.threefoldtoken.com:23112",
		"bootstrap4.testnet.threefoldtoken.com:24112",
		"bootstrap5.testnet.threefoldtoken.com:23112",
	}
}

func unlockHashFromHex(hstr string) (uh types.UnlockHash) {
	err := uh.LoadString(hstr)
	if err != nil {
		panic(fmt.Sprintf("func unlockHashFromHex(%s) failed: %v", hstr, err))
	}
	return
}

func init() {
	Version = build.MustParse(rawVersion)
}
