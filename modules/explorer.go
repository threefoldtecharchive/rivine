package modules

import (
	"github.com/rivine/rivine/types"
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
		BlockID           types.BlockID     `json:"blockid"`
		Difficulty        types.Difficulty  `json:"difficulty"`
		EstimatedActiveBS types.Difficulty  `json:"estimatedactivebs"`
		Height            types.BlockHeight `json:"height"`
		MaturityTimestamp types.Timestamp   `json:"maturitytimestamp"`
		Target            types.Target      `json:"target"`
		TotalCoins        types.Currency    `json:"totalcoins"`

		// Transaction type counts.
		MinerPayoutCount          uint64 `json:"minerpayoutcount"`
		TransactionCount          uint64 `json:"transactioncount"`
		CoinInputCount            uint64 `json:"coininputcount"`
		CoinOutputCount           uint64 `json:"coinoutputcount"`
		BlockStakeInputCount      uint64 `json:"blockstakeinputcount"`
		BlockStakeOutputCount     uint64 `json:"blockstakeoutputcount"`
		MinerFeeCount             uint64 `json:"minerfeecount"`
		ArbitraryDataCount        uint64 `json:"arbitrarydatacount"`
		TransactionSignatureCount uint64 `json:"transactionsignaturecount"`
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

		Close() error
	}
)
