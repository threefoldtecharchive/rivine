package api

import (
	"fmt"
	"net/http"

	"github.com/rivine/rivine/build"
	"github.com/rivine/rivine/crypto"
	"github.com/rivine/rivine/modules"
	"github.com/rivine/rivine/types"

	"github.com/julienschmidt/httprouter"
)

// hash type string constants
const (
	HashTypeTransactionIDStr      = "transactionid"
	HashTypeCoinOutputIDStr       = "coinoutputid"
	HashTypeBlockStakeOutputIDStr = "blockstakeoutputid"
	HashTypeUnlockHashStr         = "unlockhash"
	HashTypeBlockIDStr            = "blockid"
)

type (
	// ExplorerBlock is a block with some extra information such as the id and
	// height. This information is provided for programs that may not be
	// complex enough to compute the ID on their own.
	ExplorerBlock struct {
		MinerPayoutIDs []types.CoinOutputID  `json:"minerpayoutids"`
		Transactions   []ExplorerTransaction `json:"transactions"`
		RawBlock       types.Block           `json:"rawblock"`

		modules.BlockFacts
	}

	// ExplorerTransaction is a transcation with some extra information such as
	// the parent block. This information is provided for programs that may not
	// be complex enough to compute the extra information on their own.
	ExplorerTransaction struct {
		ID             types.TransactionID `json:"id"`
		Height         types.BlockHeight   `json:"height"`
		Parent         types.BlockID       `json:"parent"`
		RawTransaction types.Transaction   `json:"rawtransaction"`

		CoinInputOutputs             []ExplorerCoinOutput       `json:"coininputoutputs"` // the outputs being spent
		CoinOutputIDs                []types.CoinOutputID       `json:"coinoutputids"`
		CoinOutputUnlockHashes       []types.UnlockHash         `json:"coinoutputunlockhashes"`
		BlockStakeInputOutputs       []ExplorerBlockStakeOutput `json:"blockstakeinputoutputs"` // the outputs being spent
		BlockStakeOutputIDs          []types.BlockStakeOutputID `json:"blockstakeoutputids"`
		BlockStakeOutputUnlockHashes []types.UnlockHash         `json:"blockstakeunlockhashes"`

		Unconfirmed bool `json:"unconfirmed"`
	}
)

// BuildExplorerTransaction takes a transaction and the height + id of the
// block it appears in an uses that to build an explorer transaction.
func BuildExplorerTransaction(explorer modules.Explorer, height types.BlockHeight, parent types.BlockID, txn types.Transaction) (et ExplorerTransaction) {
	// Get the header information for the transaction.
	et.ID = txn.ID()
	et.Height = height
	et.Parent = parent
	et.RawTransaction = txn

	// Add the siacoin outputs that correspond with each siacoin input.
	for _, sci := range txn.CoinInputs {
		sco, exists := explorer.CoinOutput(sci.ParentID)
		if build.DEBUG && !exists {
			panic("could not find corresponding coin output")
		}
		et.CoinInputOutputs = append(et.CoinInputOutputs, ExplorerCoinOutput{
			CoinOutput: sco,
			UnlockHash: sco.Condition.UnlockHash(),
		})
	}

	for i, co := range txn.CoinOutputs {
		et.CoinOutputIDs = append(et.CoinOutputIDs, txn.CoinOutputID(uint64(i)))
		et.CoinOutputUnlockHashes = append(et.CoinOutputUnlockHashes, co.Condition.UnlockHash())
	}

	// Add the siafund outputs that correspond to each siacoin input.
	for _, sci := range txn.BlockStakeInputs {
		sco, exists := explorer.BlockStakeOutput(sci.ParentID)
		if build.DEBUG && !exists {
			panic("could not find corresponding blockstake output")
		}
		et.BlockStakeInputOutputs = append(et.BlockStakeInputOutputs, ExplorerBlockStakeOutput{
			BlockStakeOutput: sco,
			UnlockHash:       sco.Condition.UnlockHash(),
		})
	}

	for i, bso := range txn.BlockStakeOutputs {
		et.BlockStakeOutputIDs = append(et.BlockStakeOutputIDs, txn.BlockStakeOutputID(uint64(i)))
		et.BlockStakeOutputUnlockHashes = append(et.BlockStakeOutputUnlockHashes, bso.Condition.UnlockHash())
	}

	return et
}

// BuildExplorerBlock takes a block and its height and uses it to construct
// an explorer block.
func BuildExplorerBlock(explorer modules.Explorer, height types.BlockHeight, block types.Block) ExplorerBlock {
	var mpoids []types.CoinOutputID
	for i := range block.MinerPayouts {
		mpoids = append(mpoids, block.MinerPayoutID(uint64(i)))
	}

	var etxns []ExplorerTransaction
	for _, txn := range block.Transactions {
		etxns = append(etxns, BuildExplorerTransaction(explorer, height, block.ID(), txn))
	}

	facts, exists := explorer.BlockFacts(height)
	if build.DEBUG && !exists {
		panic("incorrect request to buildExplorerBlock - block does not exist")
	}

	return ExplorerBlock{
		MinerPayoutIDs: mpoids,
		Transactions:   etxns,
		RawBlock:       block,

		BlockFacts: facts,
	}
}

// BuildTransactionSet returns the blocks and transactions that are associated
// with a set of transaction ids.
func BuildTransactionSet(explorer modules.Explorer, txids []types.TransactionID) (txns []ExplorerTransaction, blocks []ExplorerBlock) {
	for _, txid := range txids {
		// Get the block containing the transaction - in the case of miner
		// payouts, the block might be the transaction.
		block, height, exists := explorer.Transaction(txid)
		if !exists && build.DEBUG {
			panic("explorer pointing to nonexistent txn")
		}

		// Check if the block is the transaction.
		if types.TransactionID(block.ID()) == txid {
			blocks = append(blocks, BuildExplorerBlock(explorer, height, block))
		} else {
			// Find the transaction within the block with the correct id.
			for _, t := range block.Transactions {
				if t.ID() == txid {
					txns = append(txns, BuildExplorerTransaction(explorer, height, block.ID(), t))
					break
				}
			}
		}
	}
	return txns, blocks
}

type (
	// ExplorerCoinOutput is the same a regular types.CoinOutput,
	// but with the addition of the pre-computed UnlockHash of its condition.
	ExplorerCoinOutput struct {
		types.CoinOutput
		UnlockHash types.UnlockHash `json:"unlockhash"`
	}

	// ExplorerBlockStakeOutput is the same a regular types.BlockStakeOutput,
	// but with the addition of the pre-computed UnlockHash of its condition.
	ExplorerBlockStakeOutput struct {
		types.BlockStakeOutput
		UnlockHash types.UnlockHash `json:"unlockhash"`
	}

	// ExplorerGET is the object returned as a response to a GET request to
	// /explorer.
	ExplorerGET struct {
		modules.BlockFacts
	}

	// ExplorerBlockGET is the object returned by a GET request to
	// /explorer/block.
	ExplorerBlockGET struct {
		Block ExplorerBlock `json:"block"`
	}

	// ExplorerHashGET is the object returned as a response to a GET request to
	// /explorer/hash. The HashType will indicate whether the hash corresponds
	// to a block id, a transaction id, a siacoin output id, a file contract
	// id, or a siafund output id. In the case of a block id, 'Block' will be
	// filled out and all the rest of the fields will be blank. In the case of
	// a transaction id, 'Transaction' will be filled out and all the rest of
	// the fields will be blank. For everything else, 'Transactions' and
	// 'Blocks' will/may be filled out and everything else will be blank.
	ExplorerHashGET struct {
		HashType          string                `json:"hashtype"`
		Block             ExplorerBlock         `json:"block"`
		Blocks            []ExplorerBlock       `json:"blocks"`
		Transaction       ExplorerTransaction   `json:"transaction"`
		Transactions      []ExplorerTransaction `json:"transactions"`
		MultiSigAddresses []types.UnlockHash    `json:"multisigaddresses"`
		Unconfirmed       bool                  `json:"unconfirmed"`
	}
)

// RegisterExplorerHTTPHandlers registers the default Rivine handlers for all default Rivine Explprer HTTP endpoints.
func RegisterExplorerHTTPHandlers(router Router, cs modules.ConsensusSet, explorer modules.Explorer, tpool modules.TransactionPool) {
	if cs == nil {
		panic("no consensus module given")
	}
	if explorer == nil {
		panic("no explorer module given")
	}
	if router == nil {
		panic("no httprouter Router given")
	}

	router.GET("/explorer", NewExplorerRootHandler(explorer))
	router.GET("/explorer/blocks/:height", NewExplorerBlocksHandler(cs, explorer))
	router.GET("/explorer/hashes/:hash", NewExplorerHashHandler(explorer, tpool))
	router.GET("/explorer/stats/history", NewExplorerHistoryStatsHandler(explorer))
	router.GET("/explorer/stats/range", NewExplorerRangeStatsHandler(explorer))
	router.GET("/explorer/constants", NewExplorerConstantsHandler(explorer))
}

// NewExplorerBlocksHandler creates a handler to handle API calls to /explorer/blocks/:height.
func NewExplorerBlocksHandler(cs modules.ConsensusSet, explorer modules.Explorer) httprouter.Handle {
	return func(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
		// Parse the height that's being requested.
		var height types.BlockHeight
		_, err := fmt.Sscan(ps.ByName("height"), &height)
		if err != nil {
			WriteError(w, Error{err.Error()}, http.StatusBadRequest)
			return
		}

		// Fetch and return the explorer block.
		block, exists := cs.BlockAtHeight(height)
		if !exists {
			WriteError(w, Error{"no block found at input height in call to /explorer/block"}, http.StatusBadRequest)
			return
		}
		WriteJSON(w, ExplorerBlockGET{
			Block: BuildExplorerBlock(explorer, height, block),
		})
	}
}

// NewExplorerHashHandler creates a handler to handle GET requests to /explorer/hash/:hash.
func NewExplorerHashHandler(explorer modules.Explorer, tpool modules.TransactionPool) httprouter.Handle {
	return func(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
		// Scan the hash as a hash. If that fails, try scanning the hash as an
		// address.
		hash, err := ScanHash(ps.ByName("hash"))
		if err != nil {
			addr, err := ScanAddress(ps.ByName("hash"))
			if err != nil {
				WriteError(w, Error{err.Error()}, http.StatusBadRequest)
				return
			}

			// Try the hash as an unlock hash. Unlock hash is checked last because
			// unlock hashes do not have collision-free guarantees. Someone can create
			// an unlock hash that collides with another object id. They will not be
			// able to use the unlock hash, but they can disrupt the explorer. This is
			// handled by checking the unlock hash last. Anyone intentionally creating
			// a colliding unlock hash (such a collision can only happen if done
			// intentionally) will be unable to find their unlock hash in the
			// blockchain through the explorer hash lookup.
			var (
				txns   []ExplorerTransaction
				blocks []ExplorerBlock
			)
			if txids := explorer.UnlockHash(addr); len(txids) != 0 {
				txns, blocks = BuildTransactionSet(explorer, txids)
			}
			txns = append(txns, getUnconfirmedTransactions(explorer, tpool, addr)...)
			multiSigAddresses := explorer.MultiSigAddresses(addr)
			if len(txns) != 0 || len(blocks) != 0 || len(multiSigAddresses) != 0 {
				WriteJSON(w, ExplorerHashGET{
					HashType:          HashTypeUnlockHashStr,
					Blocks:            blocks,
					Transactions:      txns,
					MultiSigAddresses: multiSigAddresses,
				})
				return
			}

			// Hash not found, return an error.
			WriteError(w, Error{"unrecognized hash used as input to /explorer/hash"}, http.StatusBadRequest)
			return
		}

		// TODO: lookups on the zero hash are too expensive to allow. Need a
		// better way to handle this case.
		if hash == (crypto.Hash{}) {
			WriteError(w, Error{"can't lookup the empty unlock hash"}, http.StatusBadRequest)
			return
		}

		// Try the hash as a block id.
		block, height, exists := explorer.Block(types.BlockID(hash))
		if exists {
			WriteJSON(w, ExplorerHashGET{
				HashType: HashTypeBlockIDStr,
				Block:    BuildExplorerBlock(explorer, height, block),
			})
			return
		}

		// Try the hash as a transaction id.
		block, height, exists = explorer.Transaction(types.TransactionID(hash))
		if exists {
			var txn types.Transaction
			for _, t := range block.Transactions {
				if t.ID() == types.TransactionID(hash) {
					txn = t
				}
			}
			WriteJSON(w, ExplorerHashGET{
				HashType:    HashTypeTransactionIDStr,
				Transaction: BuildExplorerTransaction(explorer, height, block.ID(), txn),
			})
			return
		}

		// Try the hash as a siacoin output id.
		txids := explorer.CoinOutputID(types.CoinOutputID(hash))
		if len(txids) != 0 {
			txns, blocks := BuildTransactionSet(explorer, txids)
			WriteJSON(w, ExplorerHashGET{
				HashType:     HashTypeCoinOutputIDStr,
				Blocks:       blocks,
				Transactions: txns,
			})
			return
		}

		// Try the hash as a siafund output id.
		txids = explorer.BlockStakeOutputID(types.BlockStakeOutputID(hash))
		if len(txids) != 0 {
			txns, blocks := BuildTransactionSet(explorer, txids)
			WriteJSON(w, ExplorerHashGET{
				HashType:     HashTypeBlockStakeOutputIDStr,
				Blocks:       blocks,
				Transactions: txns,
			})
			return
		}

		// if the transaction pool is available, try to use it
		if tpool != nil {
			// Try the hash as a transactionID in the transaction pool
			txn, err := tpool.Transaction(types.TransactionID(hash))
			if err == nil {
				WriteJSON(w, ExplorerHashGET{
					HashType:    HashTypeTransactionIDStr,
					Transaction: BuildExplorerTransaction(explorer, 0, types.BlockID{}, txn),
					Unconfirmed: true,
				})
				return
			}
			if err != modules.ErrTransactionNotFound {
				WriteError(w, Error{
					"error during call to /explorer/hash: failed to get txn from transaction pool: " + err.Error()},
					http.StatusInternalServerError)
				return
			}
		}

		// Hash not found, return an error.
		WriteError(w, Error{"unrecognized hash used as input to /explorer/hash"}, http.StatusBadRequest)
	}
}

// NewExplorerRootHandler creates a handler to handle API calls to /explorer
func NewExplorerRootHandler(explorer modules.Explorer) httprouter.Handle {
	return func(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
		facts := explorer.LatestBlockFacts()
		WriteJSON(w, ExplorerGET{
			BlockFacts: facts,
		})
	}
}

// NewExplorerConstantsHandler creates a handler to handle API calls to /explorer/constants
func NewExplorerConstantsHandler(explorer modules.Explorer) httprouter.Handle {
	return func(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
		WriteJSON(w, explorer.Constants())
	}
}

// NewExplorerHistoryStatsHandler creates a handler to handle API calls to /explorer/stats/history
func NewExplorerHistoryStatsHandler(explorer modules.Explorer) httprouter.Handle {
	return func(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
		var history types.BlockHeight
		// GET request so the only place the vars can be is the queryparams
		q := req.URL.Query()
		_, err := fmt.Sscan(q.Get("history"), &history)
		if err != nil {
			WriteError(w, Error{err.Error()}, http.StatusBadRequest)
			return
		}
		stats, err := explorer.HistoryStats(history)
		if err != nil {
			WriteError(w, Error{err.Error()}, http.StatusBadRequest)
			return
		}
		WriteJSON(w, stats)
	}
}

// NewExplorerRangeStatsHandler creates a handler to handle API calls to /explorer/stats/range
func NewExplorerRangeStatsHandler(explorer modules.Explorer) httprouter.Handle {
	return func(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
		var start, end types.BlockHeight
		// GET request so the only place the vars can be is the queryparams
		q := req.URL.Query()
		_, err := fmt.Sscan(q.Get("start"), &start)
		if err != nil {
			WriteError(w, Error{err.Error()}, http.StatusBadRequest)
			return
		}
		_, err = fmt.Sscan(q.Get("end"), &end)
		if err != nil {
			WriteError(w, Error{err.Error()}, http.StatusBadRequest)
			return
		}
		stats, err := explorer.RangeStats(start, end)
		if err != nil {
			WriteError(w, Error{err.Error()}, http.StatusBadRequest)
			return
		}
		WriteJSON(w, stats)
	}
}

// getUnconfirmedTransactions returns a list of all transactions which are unconfirmed and related to the given unlock hash from the transactionpool
func getUnconfirmedTransactions(explorer modules.Explorer, tpool modules.TransactionPool, addr types.UnlockHash) []ExplorerTransaction {
	if tpool == nil {
		return nil
	}
	relatedTxns := []types.Transaction{}
	unconfirmedTxns := tpool.TransactionList()
	for _, txn := range unconfirmedTxns {
		related := false
		// Check if any coin output is related to the hash we currently have
		for _, co := range txn.CoinOutputs {
			if co.Condition.UnlockHash() == addr {
				related = true
				relatedTxns = append(relatedTxns, txn)
				break
			}
		}
		if related {
			continue
		}
		// Check if any blockstake output is related
		for _, bso := range txn.BlockStakeOutputs {
			if bso.Condition.UnlockHash() == addr {
				related = true
				relatedTxns = append(relatedTxns, txn)
				break
			}
		}
		if related {
			continue
		}
		// Check the coin inputs
		for _, ci := range txn.CoinInputs {
			co, _ := explorer.CoinOutput(ci.ParentID)
			if co.Condition.UnlockHash() == addr {
				related = true
				relatedTxns = append(relatedTxns, txn)
				break
			}
		}
		if related {
			continue
		}
		// Check blockstake inputs
		for _, bsi := range txn.BlockStakeInputs {
			bsi, _ := explorer.BlockStakeOutput(bsi.ParentID)
			if bsi.Condition.UnlockHash() == addr {
				related = true
				relatedTxns = append(relatedTxns, txn)
				break
			}
		}
	}
	explorerTxns := make([]ExplorerTransaction, len(relatedTxns))
	for i := range relatedTxns {
		explorerTxns[i] = BuildExplorerTransaction(explorer, 0, types.BlockID{}, relatedTxns[i])
		explorerTxns[i].Unconfirmed = true
	}
	return explorerTxns
}
