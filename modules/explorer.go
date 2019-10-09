package modules

import (
	"math/big"

	"github.com/threefoldtech/rivine/types"
)

const (
	// ExplorerDir is the name of the directory that is typically used for the
	// explorer.
	ExplorerDir = "explorer"
)

type (
	// BlockFacts returns a bunch of statistics about the consensus set as they
	// were at a specific block.
	BlockFacts struct {
		BlockID                types.BlockID     `json:"blockid"`
		Difficulty             types.Difficulty  `json:"difficulty"`
		EstimatedActiveBS      types.Difficulty  `json:"estimatedactivebs"`
		Height                 types.BlockHeight `json:"height"`
		MaturityTimestamp      types.Timestamp   `json:"maturitytimestamp"`
		Target                 types.Target      `json:"target"`
		TotalCoins             types.Currency    `json:"totalcoins"`
		ArbitraryDataTotalSize uint64            `json:"arbitrarydatatotalsize"`

		// Transaction type counts.
		MinerPayoutCount      uint64 `json:"minerpayoutcount"`
		TransactionCount      uint64 `json:"transactioncount"`
		CoinInputCount        uint64 `json:"coininputcount"`
		CoinOutputCount       uint64 `json:"coinoutputcount"`
		BlockStakeInputCount  uint64 `json:"blockstakeinputcount"`
		BlockStakeOutputCount uint64 `json:"blockstakeoutputcount"`
		MinerFeeCount         uint64 `json:"minerfeecount"`
		ArbitraryDataCount    uint64 `json:"arbitrarydatacount"`
	}

	// ChainStats are data points related to a selection of blocks
	ChainStats struct {
		BlockCount uint32 `json:"blockcount"`
		// The following fields all have the same length,
		// and are in the same order. The length is equal
		// to `BlockCount`.
		BlockHeights           []types.BlockHeight `json:"blockheights"`
		BlockTimeStamps        []types.Timestamp   `json:"blocktimestamps"`
		BlockTimes             []int64             `json:"blocktimes"` // Time to create this block
		EstimatedActiveBS      []types.Difficulty  `json:"estimatedactivebs"`
		BlockTransactionCounts []uint32            `json:"blocktransactioncounts"` // txns in a single block
		Difficulties           []types.Difficulty  `json:"difficulties"`

		// Who created these blocks and how many
		Creators map[string]uint32 `json:"creators"`

		// Some aggregated stats at the time of
		// the respective block
		TransactionCounts      []uint64 `json:"transactioncounts"`
		CoinInputCounts        []uint64 `json:"coininputcounts"`
		CoinOutputCounts       []uint64 `json:"coinoutputcounts"`
		BlockStakeInputCounts  []uint64 `json:"blockstakeinputcounts"`
		BlockStakeOutputCounts []uint64 `json:"blockstakeoutputcounts"`
	}

	// DaemonConstants represent the constants in use by the daemon
	DaemonConstants struct {
		ChainInfo types.BlockchainInfo `json:"chaininfo"`

		// ConsensusPlugins are the plugins loaded in the consensus set module of this daemon
		ConsensusPlugins []string `json:"consensusplugins"`

		GenesisTimestamp       types.Timestamp   `json:"genesistimestamp"`
		BlockSizeLimit         uint64            `json:"blocksizelimit"`
		BlockFrequency         types.BlockHeight `json:"blockfrequency"`
		FutureThreshold        types.Timestamp   `json:"futurethreshold"`
		ExtremeFutureThreshold types.Timestamp   `json:"extremefuturethreshold"`
		BlockStakeCount        types.Currency    `json:"blockstakecount"`

		BlockStakeAging        uint64                     `json:"blockstakeaging"`
		BlockCreatorFee        types.Currency             `json:"blockcreatorfee"`
		MinimumTransactionFee  types.Currency             `json:"minimumtransactionfee"`
		TransactionFeeConition types.UnlockConditionProxy `json:"transactionfeebeneficiary"`

		MaturityDelay         types.BlockHeight `json:"maturitydelay"`
		MedianTimestampWindow uint64            `json:"mediantimestampwindow"`

		RootTarget types.Target `json:"roottarget"`
		RootDepth  types.Target `json:"rootdepth"`

		TargetWindow      types.BlockHeight `json:"targetwindow"`
		MaxAdjustmentUp   *big.Rat          `json:"maxadjustmentup"`
		MaxAdjustmentDown *big.Rat          `json:"maxadjustmentdown"`

		OneCoin types.Currency `json:"onecoin"`

		DefaultTransactionVersion types.TransactionVersion `json:"deftransactionversion"`
	}

	// Explorer tracks the blockchain and provides tools for gathering
	// statistics and finding objects or patterns within the blockchain.
	Explorer interface {
		// Block returns the block that matches the input block id. The bool
		// indicates whether the block appears in the blockchain.
		Block(types.BlockID) (types.Block, types.BlockHeight, bool)

		// BlockFacts returns a set of statistics about the blockchain as they
		// appeared at a given block.
		BlockFacts(types.BlockHeight) (BlockFacts, bool)

		// LatestBlockFacts returns the block facts of the last block
		// in the explorer's database.
		LatestBlockFacts() BlockFacts

		// Transaction returns the block that contains the input transaction
		// id. The transaction itself is either the block (indicating the miner
		// payouts are somehow involved), or it is a transaction inside of the
		// block. The bool indicates whether the transaction is found in the
		// consensus set.
		Transaction(types.TransactionID) (types.Block, types.BlockHeight, bool)

		// UnlockHash returns all of the transaction ids associated with the
		// provided unlock hash.
		UnlockHash(types.UnlockHash) []types.TransactionID

		// MultiSigAddresses returns all multisig addresses this wallet address is involved in.
		MultiSigAddresses(types.UnlockHash) []types.UnlockHash

		// CoinOutput will return the coin output associated with the
		// input id.
		CoinOutput(types.CoinOutputID) (types.CoinOutput, bool)

		// CoinOutputID returns all of the transaction ids associated with
		// the provided coin output id.
		CoinOutputID(types.CoinOutputID) []types.TransactionID

		// BlockStakeOutput will return the blockstake output associated with the
		// input id.
		BlockStakeOutput(types.BlockStakeOutputID) (types.BlockStakeOutput, bool)

		// BlockStakeOutputID returns all of the transaction ids associated with
		// the provided blockstake output id.
		BlockStakeOutputID(types.BlockStakeOutputID) []types.TransactionID

		// HistoryStats return the stats for the last `history` amount of blocks
		HistoryStats(types.BlockHeight) (*ChainStats, error)

		// RangeStats return the stats for the range [`start`, `end`]
		RangeStats(types.BlockHeight, types.BlockHeight) (*ChainStats, error)

		// Constants returns the constants in use by the chain
		Constants() DaemonConstants

		Close() error
	}
)

// NewChainStats initializes a new `ChainStats` object
func NewChainStats(size int) *ChainStats {
	if size <= 0 {
		return nil
	}
	return &ChainStats{
		BlockCount:             uint32(size),
		BlockHeights:           make([]types.BlockHeight, size),
		BlockTimeStamps:        make([]types.Timestamp, size),
		BlockTimes:             make([]int64, size),
		EstimatedActiveBS:      make([]types.Difficulty, size),
		BlockTransactionCounts: make([]uint32, size),
		Difficulties:           make([]types.Difficulty, size),

		Creators: make(map[string]uint32),

		TransactionCounts:      make([]uint64, size),
		CoinInputCounts:        make([]uint64, size),
		CoinOutputCounts:       make([]uint64, size),
		BlockStakeInputCounts:  make([]uint64, size),
		BlockStakeOutputCounts: make([]uint64, size),
	}
}

// NewDaemonConstants returns the Deamon's public constants,
// using the blockchain (network) info and constants used internally as input.
func NewDaemonConstants(info types.BlockchainInfo, constants types.ChainConstants, consensusPlugins []string) DaemonConstants {
	return DaemonConstants{
		ChainInfo: info,

		ConsensusPlugins: consensusPlugins,

		GenesisTimestamp:       constants.GenesisTimestamp,
		BlockSizeLimit:         constants.BlockSizeLimit,
		BlockFrequency:         constants.BlockFrequency,
		FutureThreshold:        constants.FutureThreshold,
		ExtremeFutureThreshold: constants.ExtremeFutureThreshold,
		BlockStakeCount:        constants.GenesisBlockStakeCount(),

		BlockStakeAging:        constants.BlockStakeAging,
		BlockCreatorFee:        constants.BlockCreatorFee,
		MinimumTransactionFee:  constants.MinimumTransactionFee,
		TransactionFeeConition: constants.TransactionFeeCondition,

		MaturityDelay:         constants.MaturityDelay,
		MedianTimestampWindow: constants.MedianTimestampWindow,

		RootTarget: constants.RootTarget(),
		RootDepth:  constants.RootDepth,

		TargetWindow:      constants.TargetWindow,
		MaxAdjustmentUp:   constants.MaxAdjustmentUp,
		MaxAdjustmentDown: constants.MaxAdjustmentDown,

		OneCoin: constants.CurrencyUnits.OneCoin,

		DefaultTransactionVersion: constants.DefaultTransactionVersion,
	}
}
