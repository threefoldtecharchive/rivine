package config

import (
	"math/big"

	"github.com/threefoldtech/rivine/build"
	"github.com/threefoldtech/rivine/modules"
	"github.com/threefoldtech/rivine/types"
)

var (
	rawVersion = "v0.1"
	// Version of the chain binaries.
	//
	// Value is defined by a private build flag,
	// or hardcoded to the latest released tag as fallback.
	Version build.ProtocolVersion
)

const (
	// TokenUnit defines the unit of one Token.
	TokenUnit = "ROC"
	// TokenChainName defines the name of the chain.
	TokenChainName = "rivchain"
)

// chain network names
const (
	NetworkNameDevnet = "devnet"

	NetworkNameStandard = "standard"

	NetworkNameTestnet = "testnet"
)

func GetDefaultGenesis() types.ChainConstants {
	return GetStandardGenesis()
}

// GetBlockchainInfo returns the naming and versioning of tfchain.
func GetBlockchainInfo() types.BlockchainInfo {
	return types.BlockchainInfo{
		Name:            TokenChainName,
		NetworkName:     NetworkNameStandard,
		CoinUnit:        TokenUnit,
		ChainVersion:    Version,       // use our own blockChain/build version
		ProtocolVersion: build.Version, // use latest available rivine protocol version
	}
}

func GetDevnetGenesis() types.ChainConstants {
	cfg := types.DevnetChainConstants()

	// set transaction versions
	cfg.DefaultTransactionVersion = types.TransactionVersion(1)
	cfg.GenesisTransactionVersion = types.TransactionVersion(1)

	// size limits
	cfg.BlockSizeLimit = 2000000
	cfg.ArbitraryDataSizeLimit = 83

	// block time
	cfg.BlockFrequency = 12

	// Time to MaturityDelay
	cfg.MaturityDelay = 10

	// The genesis timestamp
	cfg.GenesisTimestamp = types.Timestamp(1571229355)

	cfg.MedianTimestampWindow = 11

	// block window for difficulty
	cfg.TargetWindow = 20

	cfg.MaxAdjustmentUp = big.NewRat(120, 100)
	cfg.MaxAdjustmentDown = big.NewRat(100, 120)

	cfg.FutureThreshold = 12
	cfg.ExtremeFutureThreshold = 60

	cfg.StakeModifierDelay = 2000

	// Time it takes before transferred blockstakes can be used
	cfg.BlockStakeAging = 1024

	// Coins you receive when you create a block
	cfg.BlockCreatorFee = cfg.CurrencyUnits.OneCoin.Mul64(1) // Minimum transaction fee
	cfg.MinimumTransactionFee = cfg.CurrencyUnits.OneCoin.Div64(10)
	cfg.TransactionFeeCondition = types.NewCondition(types.NewUnlockHashCondition(unlockHashFromHex("015a080a9259b9d4aaa550e2156f49b1a79a64c7ea463d810d4493e8242e6791584fbdac553e6f")))

	// Set Transaction Pool config
	cfg.TransactionPool = types.TransactionPoolConstants{
		TransactionSizeLimit:    16000,
		TransactionSetSizeLimit: 250000,
		PoolSizeLimit:           19750000,
	}

	// allocate initial coin outputs
	cfg.GenesisCoinDistribution = []types.CoinOutput{
		{
			Value:     cfg.CurrencyUnits.OneCoin.Mul64(500000),
			Condition: types.NewCondition(types.NewUnlockHashCondition(unlockHashFromHex("015a080a9259b9d4aaa550e2156f49b1a79a64c7ea463d810d4493e8242e6791584fbdac553e6f"))),
		},
	}

	// allocate initial block stake outputs
	cfg.GenesisBlockStakeAllocation = []types.BlockStakeOutput{
		{
			Value:     types.NewCurrency64(3000),
			Condition: types.NewCondition(types.NewUnlockHashCondition(unlockHashFromHex("015a080a9259b9d4aaa550e2156f49b1a79a64c7ea463d810d4493e8242e6791584fbdac553e6f"))),
		},
	}

	return cfg
}

func GetDevnetBootstrapPeers() []modules.NetAddress {
	return []modules.NetAddress{
		"localhost:23112",
	}
}

func GetDevnetGenesisMintCondition() types.UnlockConditionProxy {
	return types.NewCondition(types.NewUnlockHashCondition(unlockHashFromHex("015a080a9259b9d4aaa550e2156f49b1a79a64c7ea463d810d4493e8242e6791584fbdac553e6f")))
}

func GetDevnetGenesisAuthCoinCondition() types.UnlockConditionProxy {
	return types.NewCondition(types.NewUnlockHashCondition(unlockHashFromHex("015a080a9259b9d4aaa550e2156f49b1a79a64c7ea463d810d4493e8242e6791584fbdac553e6f")))
}

func GetStandardGenesis() types.ChainConstants {
	cfg := types.StandardnetChainConstants()

	// set transaction versions
	cfg.DefaultTransactionVersion = types.TransactionVersion(1)
	cfg.GenesisTransactionVersion = types.TransactionVersion(1)

	// size limits
	cfg.BlockSizeLimit = 2000000
	cfg.ArbitraryDataSizeLimit = 83

	// block time
	cfg.BlockFrequency = 120

	// Time to MaturityDelay
	cfg.MaturityDelay = 144

	// The genesis timestamp
	cfg.GenesisTimestamp = types.Timestamp(1571229355)

	cfg.MedianTimestampWindow = 11

	// block window for difficulty
	cfg.TargetWindow = 1000

	cfg.MaxAdjustmentUp = big.NewRat(25, 10)
	cfg.MaxAdjustmentDown = big.NewRat(10, 25)

	cfg.FutureThreshold = 120
	cfg.ExtremeFutureThreshold = 600

	cfg.StakeModifierDelay = 2000

	// Time it takes before transferred blockstakes can be used
	cfg.BlockStakeAging = 86400

	// Coins you receive when you create a block
	cfg.BlockCreatorFee = cfg.CurrencyUnits.OneCoin.Mul64(1) // Minimum transaction fee
	cfg.MinimumTransactionFee = cfg.CurrencyUnits.OneCoin.Div64(10)
	cfg.TransactionFeeCondition = types.NewCondition(types.NewUnlockHashCondition(unlockHashFromHex("017267221ef1947bb18506e390f1f9446b995acfb6d08d8e39508bb974d9830b8cb8fdca788e34")))

	// Set Transaction Pool config
	cfg.TransactionPool = types.TransactionPoolConstants{
		TransactionSizeLimit:    16000,
		TransactionSetSizeLimit: 250000,
		PoolSizeLimit:           19750000,
	}

	// allocate initial coin outputs
	cfg.GenesisCoinDistribution = []types.CoinOutput{
		{
			Value:     cfg.CurrencyUnits.OneCoin.Mul64(500000),
			Condition: types.NewCondition(types.NewUnlockHashCondition(unlockHashFromHex("01b5e42056ef394f2ad9b511a61cec874d25bebe2095682dd37455cbafed4bec154e382a23f90e"))),
		},
	}

	// allocate initial block stake outputs
	cfg.GenesisBlockStakeAllocation = []types.BlockStakeOutput{
		{
			Value:     types.NewCurrency64(3000),
			Condition: types.NewCondition(types.NewUnlockHashCondition(unlockHashFromHex("01b5e42056ef394f2ad9b511a61cec874d25bebe2095682dd37455cbafed4bec154e382a23f90e"))),
		},
	}

	return cfg
}

func GetStandardBootstrapPeers() []modules.NetAddress {
	return []modules.NetAddress{
		"bootstrap1.rivine.io:23112",
		"bootstrap2.rivine.io:23112",
		"bootstrap3.rivine.io:23112",
	}
}

func GetStandardGenesisMintCondition() types.UnlockConditionProxy {
	return types.NewCondition(types.NewUnlockHashCondition(unlockHashFromHex("01b5e42056ef394f2ad9b511a61cec874d25bebe2095682dd37455cbafed4bec154e382a23f90e")))
}

func GetStandardGenesisAuthCoinCondition() types.UnlockConditionProxy {
	return types.NewCondition(types.NewUnlockHashCondition(unlockHashFromHex("01b5e42056ef394f2ad9b511a61cec874d25bebe2095682dd37455cbafed4bec154e382a23f90e")))
}

func GetTestnetGenesis() types.ChainConstants {
	cfg := types.TestnetChainConstants()

	// set transaction versions
	cfg.DefaultTransactionVersion = types.TransactionVersion(1)
	cfg.GenesisTransactionVersion = types.TransactionVersion(1)

	// size limits
	cfg.BlockSizeLimit = 2000000
	cfg.ArbitraryDataSizeLimit = 83

	// block time
	cfg.BlockFrequency = 120

	// Time to MaturityDelay
	cfg.MaturityDelay = 720

	// The genesis timestamp
	cfg.GenesisTimestamp = types.Timestamp(1571229355)

	cfg.MedianTimestampWindow = 11

	// block window for difficulty
	cfg.TargetWindow = 1000

	cfg.MaxAdjustmentUp = big.NewRat(25, 10)
	cfg.MaxAdjustmentDown = big.NewRat(10, 25)

	cfg.FutureThreshold = 120
	cfg.ExtremeFutureThreshold = 600

	cfg.StakeModifierDelay = 2000

	// Time it takes before transferred blockstakes can be used
	cfg.BlockStakeAging = 64

	// Coins you receive when you create a block
	cfg.BlockCreatorFee = cfg.CurrencyUnits.OneCoin.Mul64(1) // Minimum transaction fee
	cfg.MinimumTransactionFee = cfg.CurrencyUnits.OneCoin.Div64(10)
	cfg.TransactionFeeCondition = types.NewCondition(types.NewUnlockHashCondition(unlockHashFromHex("01434535fd01243c02c277cd58d71423163767a575a8ae44e15807bf545e4a8456a5c4afabad51")))

	// Set Transaction Pool config
	cfg.TransactionPool = types.TransactionPoolConstants{
		TransactionSizeLimit:    16000,
		TransactionSetSizeLimit: 250000,
		PoolSizeLimit:           19750000,
	}

	// allocate initial coin outputs
	cfg.GenesisCoinDistribution = []types.CoinOutput{
		{
			Value:     cfg.CurrencyUnits.OneCoin.Mul64(500000),
			Condition: types.NewCondition(types.NewUnlockHashCondition(unlockHashFromHex("01434535fd01243c02c277cd58d71423163767a575a8ae44e15807bf545e4a8456a5c4afabad51"))),
		},
		{
			Value:     cfg.CurrencyUnits.OneCoin.Mul64(500000),
			Condition: types.NewCondition(types.NewMultiSignatureCondition(types.UnlockHashSlice{unlockHashFromHex("01434535fd01243c02c277cd58d71423163767a575a8ae44e15807bf545e4a8456a5c4afabad51"), unlockHashFromHex("01434535fd01243c02c277cd58d71423163767a575a8ae44e15807bf545e4a8456a5c4afabad51")}, 2)),
		},
	}

	// allocate initial block stake outputs
	cfg.GenesisBlockStakeAllocation = []types.BlockStakeOutput{
		{
			Value:     types.NewCurrency64(3000),
			Condition: types.NewCondition(types.NewUnlockHashCondition(unlockHashFromHex("01434535fd01243c02c277cd58d71423163767a575a8ae44e15807bf545e4a8456a5c4afabad51"))),
		},
	}

	return cfg
}

func GetTestnetBootstrapPeers() []modules.NetAddress {
	return []modules.NetAddress{
		"bootstrap1.testnet.rivine.io:23112",
		"bootstrap2.testnet.rivine.io:23112",
		"bootstrap3.testnet.rivine.io:23112",
	}
}

func GetTestnetGenesisMintCondition() types.UnlockConditionProxy {
	return types.NewCondition(types.NewUnlockHashCondition(unlockHashFromHex("01434535fd01243c02c277cd58d71423163767a575a8ae44e15807bf545e4a8456a5c4afabad51")))
}

func GetTestnetGenesisAuthCoinCondition() types.UnlockConditionProxy {
	return types.NewCondition(types.NewUnlockHashCondition(unlockHashFromHex("01434535fd01243c02c277cd58d71423163767a575a8ae44e15807bf545e4a8456a5c4afabad51")))
}

func init() {
	Version = build.MustParse(rawVersion)
}
