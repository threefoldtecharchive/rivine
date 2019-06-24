package types

// constants.go contains the Sia constants. Depending on which build tags are
// used, the constants will be initialized to different values.
//
// CONTRIBUTE: We don't have way to check that the non-test constants are all
// sane, plus we have no coverage for them.

import (
	"errors"
	"math/big"

	"github.com/threefoldtech/rivine/build"
	"github.com/threefoldtech/rivine/crypto"
)

// ChainConstants is a utility struct which groups together the chain configuration
type ChainConstants struct {
	// BlockSizeLimit is the maximum size a single block can have, in bytes
	BlockSizeLimit uint64
	// ArbitraryDataSizeLimit is the maximum size an arbitrary data block
	// within a single transaction can have, in bytes
	ArbitraryDataSizeLimit uint64

	RootDepth Target
	// BlockFrequency is the average timespan between blocks, in seconds.
	// I.E.: On average, 1 block will be created every 1 in *BlockFrequency* seconds
	BlockFrequency BlockHeight
	// MaturityDelay is the amount of blocks for which a miner payout must "mature" before it
	// gets added to the consensus set. Until this time has passed, a miner payout cannot be spend
	MaturityDelay BlockHeight

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

	// MinimumTransactionFee is the minimum amount of hastings you need to pay
	// in order to get your transaction to be accepted by block creators.
	MinimumTransactionFee Currency

	// TransactionFeeCondition allows you to define a static unlock hash which collects all transaction fees,
	// by default it is undefined, meaning the transaction fee will go to the creator of the relevant block.
	TransactionFeeCondition UnlockConditionProxy

	// GenesisTimestamp is the unix timestamp of the genesis block
	GenesisTimestamp Timestamp
	// GenesisBlockStakeAllocation are the blockstake outputs of the genesis block
	GenesisBlockStakeAllocation []BlockStakeOutput
	// GenesisCoinDistribution are the coin outputs of the genesis block
	GenesisCoinDistribution []CoinOutput

	// GenesisTransactionVersion defines the transaction versions to be used
	// for the transaction of the genesis block.
	GenesisTransactionVersion TransactionVersion

	// DefaultTransactionVersion defines the default transaction to be used
	// for all to be created transactions. It does not impact how transactions are validated or understood.
	DefaultTransactionVersion TransactionVersion

	CurrencyUnits CurrencyUnits

	TransactionPool TransactionPoolConstants
}

// CurrencyUnits defines the units used for the different kind of currencies.
type CurrencyUnits struct {
	// OneCoin is the size of a "coin", making it possible to split a coin up if wanted
	OneCoin Currency
}

// TransactionPoolConstants defines the constants used by the TransactionPool.
type TransactionPoolConstants struct {
	// TransactionSizeLimit defines the size of the largest transaction that
	// will be accepted by the transaction pool according to the IsStandard
	// rules.
	TransactionSizeLimit int

	// TransactionSetSizeLimit defines the largest set of dependent unconfirmed
	// transactions that will be accepted by the transaction pool.
	TransactionSetSizeLimit int

	// The TransactionPoolSizeLimit is first checked, and then a transaction
	// set is added. The current transaction pool does not do any priority
	// ordering, so the size limit is such that the transaction pool will never
	// exceed the size of a block.
	PoolSizeLimit int
}

// DefaultCurrencyUnits provides sane defaults for currency units
func DefaultCurrencyUnits() CurrencyUnits {
	return CurrencyUnits{
		OneCoin: NewCurrency(new(big.Int).Exp(big.NewInt(10), big.NewInt(9), nil)),
	}
}

// DefaultTransactionPoolConstants provides sane defaults for TransactionPool constants.
func DefaultTransactionPoolConstants() TransactionPoolConstants {
	return TransactionPoolConstants{
		TransactionSizeLimit:    16e3,
		TransactionSetSizeLimit: 250e3,
		PoolSizeLimit:           2e6 - 5e3 - 250e3,
	}
}

// StandardnetChainConstants provide sane defaults for a new standard network chain. Not all constants
// are set, since some (e.g. GenesisTimestamp) are chain specific, and this also
// allows some santiy checking later
// GenesisTimestamp, GenesisBlockStakeAllocation, and GenesisCoinDistribution
// aren't set as there is no such thing as a "sane default" for these variables
// since they are really chain specific
func StandardnetChainConstants() ChainConstants {
	currencyUnits := DefaultCurrencyUnits()

	const (
		defaultTxnVersion = TransactionVersionOne
		genesisTxnVersion = TransactionVersionOne
	)

	cts := ChainConstants{
		BlockSizeLimit:            2e6,
		ArbitraryDataSizeLimit:    83,
		RootDepth:                 Target{255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255},
		BlockCreatorFee:           currencyUnits.OneCoin.Mul64(10),
		MinimumTransactionFee:     currencyUnits.OneCoin.Mul64(1),
		BlockFrequency:            600,
		MaturityDelay:             144,
		MedianTimestampWindow:     11,
		TargetWindow:              1e3,
		MaxAdjustmentUp:           big.NewRat(25, 10),
		MaxAdjustmentDown:         big.NewRat(10, 25),
		FutureThreshold:           3 * 60 * 60, // 3 hours.
		ExtremeFutureThreshold:    5 * 60 * 60, // 5 hours.
		StakeModifierDelay:        2000,
		BlockStakeAging:           1 << 17, // 2^16s < 1 day < 2^17s
		GenesisTimestamp:          Timestamp(1496322000),
		GenesisTransactionVersion: genesisTxnVersion,
		DefaultTransactionVersion: defaultTxnVersion,
		CurrencyUnits:             currencyUnits,
		TransactionPool:           DefaultTransactionPoolConstants(),
	}

	cts.GenesisBlockStakeAllocation = append(cts.GenesisBlockStakeAllocation, BlockStakeOutput{
		Value:     NewCurrency64(1000000),
		Condition: NewCondition(NewUnlockHashCondition(unlockHashFromHex("01b5e42056ef394f2ad9b511a61cec874d25bebe2095682dd37455cbafed4bec154e382a23f90e"))),
	})
	cts.GenesisCoinDistribution = append(cts.GenesisCoinDistribution, CoinOutput{
		Value:     currencyUnits.OneCoin.Mul64(100 * 1000 * 1000),
		Condition: NewCondition(NewUnlockHashCondition(unlockHashFromHex("01b5e42056ef394f2ad9b511a61cec874d25bebe2095682dd37455cbafed4bec154e382a23f90e"))),
	})
	return cts
}

// TestnetChainConstants provide sane defaults for a new testnet chain. Not all constants
// are set, since some (e.g. GenesisTimestamp) are chain specific, and this also
// allows some santiy checking later
// GenesisTimestamp, GenesisBlockStakeAllocation, and GenesisCoinDistribution
// aren't set as there is no such thing as a "sane default" for these variables
// since they are really chain specific
func TestnetChainConstants() ChainConstants {
	currencyUnits := DefaultCurrencyUnits()

	const (
		defaultTxnVersion = TransactionVersionOne
		genesisTxnVersion = TransactionVersionOne
	)

	// 'testing' settings are for automatic testing, and create much faster
	// environments than a human can interact with.
	return ChainConstants{
		BlockSizeLimit:            2e6,
		ArbitraryDataSizeLimit:    83,
		RootDepth:                 Target{255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255},
		BlockCreatorFee:           currencyUnits.OneCoin.Mul64(100),
		MinimumTransactionFee:     currencyUnits.OneCoin.Mul64(1),
		BlockFrequency:            1, // ASFAP
		MaturityDelay:             3,
		MedianTimestampWindow:     11,
		GenesisTimestamp:          testGenesisTimestamp,
		TargetWindow:              200,
		MaxAdjustmentUp:           big.NewRat(10001, 10000),
		MaxAdjustmentDown:         big.NewRat(9999, 10000),
		FutureThreshold:           3, // 3 seconds
		ExtremeFutureThreshold:    6, // seconds
		StakeModifierDelay:        20,
		BlockStakeAging:           uint64(1 << 10),
		GenesisTransactionVersion: genesisTxnVersion,
		DefaultTransactionVersion: defaultTxnVersion,
		GenesisBlockStakeAllocation: []BlockStakeOutput{
			{
				Value: NewCurrency64(2000),
				Condition: NewCondition(NewUnlockHashCondition(UnlockHash{
					Type: UnlockTypePubKey,
					Hash: crypto.Hash{214, 166, 197, 164, 29, 201, 53, 236, 106, 239, 10, 158, 127, 131, 20, 138, 63, 221, 230, 16, 98, 247, 32, 77, 210, 68, 116, 12, 241, 89, 27, 223},
				})),
			},
			{
				Value: NewCurrency64(7000),
				Condition: NewCondition(NewUnlockHashCondition(UnlockHash{
					Type: UnlockTypePubKey,
					Hash: crypto.Hash{209, 246, 228, 60, 248, 78, 242, 110, 9, 8, 227, 248, 225, 216, 163, 52, 142, 93, 47, 176, 103, 41, 137, 80, 212, 8, 132, 58, 241, 189, 2, 17},
				})),
			},
			{
				Value:     NewCurrency64(1000),
				Condition: NewCondition(nil),
			},
		},
		GenesisCoinDistribution: []CoinOutput{
			{
				Value: currencyUnits.OneCoin.Mul64(1000),
				Condition: NewCondition(NewUnlockHashCondition(UnlockHash{
					Type: UnlockTypePubKey,
					Hash: crypto.Hash{214, 166, 197, 164, 29, 201, 53, 236, 106, 239, 10, 158, 127, 131, 20, 138, 63, 221, 230, 16, 98, 247, 32, 77, 210, 68, 116, 12, 241, 89, 27, 223},
				})),
			},
		},
		CurrencyUnits:   currencyUnits,
		TransactionPool: DefaultTransactionPoolConstants(),
	}
}

// DevnetChainConstants provide sane defaults for a new devnet chain. Not all constants
// are set, since some (e.g. GenesisTimestamp) are chain specific, and this also
// allows some santiy checking later
// GenesisTimestamp, GenesisBlockStakeAllocation, and GenesisCoinDistribution
// aren't set as there is no such thing as a "sane default" for these variables
// since they are really chain specific
func DevnetChainConstants() ChainConstants {
	currencyUnits := DefaultCurrencyUnits()

	const (
		defaultTxnVersion = TransactionVersionOne
		genesisTxnVersion = TransactionVersionOne
	)

	// 'dev' settings are for small developer testnets, usually on the same
	// computer. Settings are slow enough that a small team of developers
	// can coordinate their actions over a the developer testnets, but fast
	// enough that there isn't much time wasted on waiting for things to
	// happen.
	cts := ChainConstants{
		BlockSizeLimit:         2e6,
		ArbitraryDataSizeLimit: 83,
		RootDepth:              Target{255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255},
		BlockCreatorFee:        currencyUnits.OneCoin.Mul64(10),
		MinimumTransactionFee:  currencyUnits.OneCoin.Mul64(1),
		// 12 seconds, slow enough for developers to see
		// ~each block, fast enough that blocks don't waste time
		BlockFrequency: 12,
		// 120 seconds before a delayed output matters
		// as it's expressed in units of blocks
		MaturityDelay:         10,
		MedianTimestampWindow: 11,
		// difficulty is adjusted based on prior 20 blocks
		TargetWindow: 20,
		// Difficulty adjusts quickly.
		MaxAdjustmentUp: big.NewRat(120, 100),
		// Difficulty adjusts quickly.
		MaxAdjustmentDown:      big.NewRat(100, 120),
		FutureThreshold:        2 * 60, // 2 minutes
		ExtremeFutureThreshold: 4 * 60, // 4 minutees
		// Number of blocks to take in history to calculate the stakemodifier
		StakeModifierDelay: 2000,
		// Block stake aging if unspent block stake is not at index 0
		BlockStakeAging:           uint64(1 << 10),
		GenesisTransactionVersion: genesisTxnVersion,
		DefaultTransactionVersion: defaultTxnVersion,
		GenesisTimestamp:          Timestamp(1424139000),
		CurrencyUnits:             currencyUnits,
		TransactionPool:           DefaultTransactionPoolConstants(),
	}
	// Seed for the address given below twice:
	// carbon boss inject cover mountain fetch fiber fit tornado cloth wing dinosaur proof joy intact fabric thumb rebel borrow poet chair network expire else
	cts.GenesisBlockStakeAllocation = append(cts.GenesisBlockStakeAllocation, BlockStakeOutput{
		Value:     NewCurrency64(1000000),
		Condition: NewCondition(NewUnlockHashCondition(unlockHashFromHex("015a080a9259b9d4aaa550e2156f49b1a79a64c7ea463d810d4493e8242e6791584fbdac553e6f"))),
	})
	cts.GenesisCoinDistribution = append(cts.GenesisCoinDistribution, CoinOutput{
		Value:     currencyUnits.OneCoin.Mul64(1000),
		Condition: NewCondition(NewUnlockHashCondition(unlockHashFromHex("015a080a9259b9d4aaa550e2156f49b1a79a64c7ea463d810d4493e8242e6791584fbdac553e6f"))),
	})

	return cts
}

// Validate does a sanity check on some of the constants to see if proper initialization is done
func (c *ChainConstants) Validate() error {
	if len(c.GenesisCoinDistribution) == 0 {
		return errors.New("Invalid genesis coin distribution")
	}
	if len(c.GenesisBlockStakeAllocation) == 0 {
		return errors.New("Invalid genesis blockstake allocation")
	}
	// Genesis timestamp should not be too far in the past. The reference timestamp is the timestamp of the bitcoin genesis block,
	// as it's pretty safe to assume no blockchain was created before this (Saturday, January 3, 2009 6:15:05 PM GMT)
	if c.GenesisTimestamp < Timestamp(1231006505) {
		return errors.New("Invalid genesis timestamp")
	}
	return nil
}

// GenesisBlock returns the genesis block based on the blockchain config
func (c *ChainConstants) GenesisBlock() Block {
	if err := c.GenesisTransactionVersion.IsValidTransactionVersion(); err != nil {
		build.Critical(err)
	}
	return Block{
		Timestamp: c.GenesisTimestamp,
		Transactions: []Transaction{
			{
				Version:           c.GenesisTransactionVersion,
				BlockStakeOutputs: c.GenesisBlockStakeAllocation,
				CoinOutputs:       c.GenesisCoinDistribution,
			},
		},
	}
}

// GenesisBlockID returns the ID of the genesis Block
func (c *ChainConstants) GenesisBlockID() BlockID {
	return c.GenesisBlock().ID()
}

// GenesisBlockStakeCount computes and returns the total amount of
// block stakes allocated in the genesis block.
func (c *ChainConstants) GenesisBlockStakeCount() (bsc Currency) {
	for _, bs := range c.GenesisBlockStakeAllocation {
		bsc = bsc.Add(bs.Value)
	}
	return
}

// GenesisCoinCount computes and returns the total amount of coins
// distributed in the genesis block.
func (c *ChainConstants) GenesisCoinCount() (cc Currency) {
	for _, coin := range c.GenesisCoinDistribution {
		cc = cc.Add(coin.Value)
	}
	return
}

// StartDifficulty computes the start difficulty based on the set block frequency,
// and the computer genesis block stake count.
func (c *ChainConstants) StartDifficulty() Difficulty {
	startDifficulty := NewDifficulty(
		big.NewInt(0).Mul(big.NewInt(int64(c.BlockFrequency)),
			c.GenesisBlockStakeCount().Big()))
	// Add a check for a zero difficulty to avoid zero division. If the startDifficulty is zero, just
	// set it to something positive. It doesn't really matter what as there can be no block creation anyway
	// due to the lack of blockstake.
	if startDifficulty.Cmp(Difficulty{}) == 0 {
		return Difficulty{i: *big.NewInt(1)}
	}
	return startDifficulty
}

// RootTarget computes the new target, based on the root depth and
// the computed start difficulty
func (c *ChainConstants) RootTarget() Target {
	return NewTarget(c.StartDifficulty(), c.RootDepth)
}

// testGenesisTimestamp is computed only once, and reused always afters
var testGenesisTimestamp = CurrentTimestamp() - 1e6
