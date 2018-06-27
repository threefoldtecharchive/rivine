package api

import (
	"encoding/hex"
	"fmt"
	"net/http"

	"github.com/rivine/rivine/build"
	"github.com/rivine/rivine/crypto"
	"github.com/rivine/rivine/encoding"
	"github.com/rivine/rivine/modules"
	"github.com/rivine/rivine/types"

	"github.com/julienschmidt/httprouter"
)

type (
	// ExplorerBlock is a block with some extra information such as the id and
	// height. This information is provided for programs that may not be
	// complex enough to compute the ID on their own.
	ExplorerBlock struct {
		MinerPayoutIDs []types.CoinOutputID  `json:"minerpayoutids"`
		Transactions   []ExplorerTransaction `json:"transactions"`
		RawBlock       types.Block           `json:"rawblock"`
		HexBlock       string                `json:"hexblock"`

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
		HexTransaction string              `json:"hextransaction"`

		CoinInputOutputs             []ExplorerCoinOutput       `json:"coininputoutputs"` // the outputs being spent
		CoinOutputIDs                []types.CoinOutputID       `json:"coinoutputids"`
		CoinOutputUnlockHashes       []types.UnlockHash         `json:"coinoutputunlockhashes"`
		BlockStakeInputOutputs       []ExplorerBlockStakeOutput `json:"blockstakeinputoutputs"` // the outputs being spent
		BlockStakeOutputIDs          []types.BlockStakeOutputID `json:"blockstakeoutputids"`
		BlockStakeOutputUnlockHashes []types.UnlockHash         `json:"blockstakeunlockhashes"`
	}

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
		HashType     string                `json:"hashtype"`
		Block        ExplorerBlock         `json:"block"`
		Blocks       []ExplorerBlock       `json:"blocks"`
		Transaction  ExplorerTransaction   `json:"transaction"`
		Transactions []ExplorerTransaction `json:"transactions"`
		Unconfirmed  bool                  `json:"unconfirmed"`
	}
)

// buildExplorerTransaction takes a transaction and the height + id of the
// block it appears in an uses that to build an explorer transaction.
func (api *API) buildExplorerTransaction(height types.BlockHeight, parent types.BlockID, txn types.Transaction) (et ExplorerTransaction) {
	// Get the header information for the transaction.
	et.ID = txn.ID()
	et.Height = height
	et.Parent = parent
	et.RawTransaction = txn
	et.HexTransaction = hex.EncodeToString(encoding.Marshal(txn))

	// Add the siacoin outputs that correspond with each siacoin input.
	for _, sci := range txn.CoinInputs {
		sco, exists := api.explorer.CoinOutput(sci.ParentID)
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
		sco, exists := api.explorer.BlockStakeOutput(sci.ParentID)
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

// buildExplorerBlock takes a block and its height and uses it to construct an
// explorer block.
func (api *API) buildExplorerBlock(height types.BlockHeight, block types.Block) ExplorerBlock {
	var mpoids []types.CoinOutputID
	for i := range block.MinerPayouts {
		mpoids = append(mpoids, block.MinerPayoutID(uint64(i)))
	}

	var etxns []ExplorerTransaction
	for _, txn := range block.Transactions {
		etxns = append(etxns, api.buildExplorerTransaction(height, block.ID(), txn))
	}

	facts, exists := api.explorer.BlockFacts(height)
	if build.DEBUG && !exists {
		panic("incorrect request to buildExplorerBlock - block does not exist")
	}

	return ExplorerBlock{
		MinerPayoutIDs: mpoids,
		Transactions:   etxns,
		RawBlock:       block,
		HexBlock:       hex.EncodeToString(encoding.Marshal(block)),

		BlockFacts: facts,
	}
}

// explorerHandler handles API calls to /explorer/blocks/:height.
func (api *API) explorerBlocksHandler(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	// Parse the height that's being requested.
	var height types.BlockHeight
	_, err := fmt.Sscan(ps.ByName("height"), &height)
	if err != nil {
		WriteError(w, Error{err.Error()}, http.StatusBadRequest)
		return
	}

	// Fetch and return the explorer block.
	block, exists := api.cs.BlockAtHeight(height)
	if !exists {
		WriteError(w, Error{"no block found at input height in call to /explorer/block"}, http.StatusBadRequest)
		return
	}
	WriteJSON(w, ExplorerBlockGET{
		Block: api.buildExplorerBlock(height, block),
	})
}

// buildTransactionSet returns the blocks and transactions that are associated
// with a set of transaction ids.
func (api *API) buildTransactionSet(txids []types.TransactionID) (txns []ExplorerTransaction, blocks []ExplorerBlock) {
	for _, txid := range txids {
		// Get the block containing the transaction - in the case of miner
		// payouts, the block might be the transaction.
		block, height, exists := api.explorer.Transaction(txid)
		if !exists && build.DEBUG {
			panic("explorer pointing to nonexistent txn")
		}

		// Check if the block is the transaction.
		if types.TransactionID(block.ID()) == txid {
			blocks = append(blocks, api.buildExplorerBlock(height, block))
		} else {
			// Find the transaction within the block with the correct id.
			for _, t := range block.Transactions {
				if t.ID() == txid {
					txns = append(txns, api.buildExplorerTransaction(height, block.ID(), t))
					break
				}
			}
		}
	}
	return txns, blocks
}

// explorerHashHandler handles GET requests to /explorer/hash/:hash.
func (api *API) explorerHashHandler(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	// Scan the hash as a hash. If that fails, try scanning the hash as an
	// address.
	hash, err := scanHash(ps.ByName("hash"))
	if err != nil {
		addr, err := scanAddress(ps.ByName("hash"))
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
		txids := api.explorer.UnlockHash(addr)
		if len(txids) != 0 {
			txns, blocks := api.buildTransactionSet(txids)
			WriteJSON(w, ExplorerHashGET{
				HashType:     "unlockhash",
				Blocks:       blocks,
				Transactions: txns,
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
	block, height, exists := api.explorer.Block(types.BlockID(hash))
	if exists {
		WriteJSON(w, ExplorerHashGET{
			HashType: "blockid",
			Block:    api.buildExplorerBlock(height, block),
		})
		return
	}

	// Try the hash as a transaction id.
	block, height, exists = api.explorer.Transaction(types.TransactionID(hash))
	if exists {
		var txn types.Transaction
		for _, t := range block.Transactions {
			if t.ID() == types.TransactionID(hash) {
				txn = t
			}
		}
		WriteJSON(w, ExplorerHashGET{
			HashType:    HashTypeTransactionIDStr,
			Transaction: api.buildExplorerTransaction(height, block.ID(), txn),
		})
		return
	}

	// Try the hash as a siacoin output id.
	txids := api.explorer.CoinOutputID(types.CoinOutputID(hash))
	if len(txids) != 0 {
		txns, blocks := api.buildTransactionSet(txids)
		WriteJSON(w, ExplorerHashGET{
			HashType:     HashTypeCoinOutputIDStr,
			Blocks:       blocks,
			Transactions: txns,
		})
		return
	}

	// Try the hash as a siafund output id.
	txids = api.explorer.BlockStakeOutputID(types.BlockStakeOutputID(hash))
	if len(txids) != 0 {
		txns, blocks := api.buildTransactionSet(txids)
		WriteJSON(w, ExplorerHashGET{
			HashType:     HashTypeBlockStakeOutputIDStr,
			Blocks:       blocks,
			Transactions: txns,
		})
		return
	}

	// if the transaction pool is available, try to use it
	if api.tpool != nil {
		// Try the hash as a transactionID in the transaction pool
		txn, err := api.tpool.Transaction(types.TransactionID(hash))
		if err == nil {
			WriteJSON(w, ExplorerHashGET{
				HashType:    HashTypeTransactionIDStr,
				Transaction: api.buildExplorerTransaction(0, types.BlockID{}, txn),
				Unconfirmed: true,
			})
			return
		}
		if err != modules.ErrTransactionNotFound {
			WriteError(w, Error{
				"error during call to /explorer/hash: failed to get txn from transaction pool: " + err.Error()},
				http.StatusInternalServerError)
		}
	}

	// Hash not found, return an error.
	WriteError(w, Error{"unrecognized hash used as input to /explorer/hash"}, http.StatusBadRequest)
}

// hash type string constants
const (
	HashTypeTransactionIDStr      = "transactionid"
	HashTypeCoinOutputIDStr       = "coinoutputid"
	HashTypeBlockStakeOutputIDStr = "blockstakeoutputid"
)

// explorerHandler handles API calls to /explorer
func (api *API) explorerHandler(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	facts := api.explorer.LatestBlockFacts()
	WriteJSON(w, ExplorerGET{
		BlockFacts: facts,
	})
}

func (api *API) constantsHandler(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
	WriteJSON(w, api.explorer.Constants())
}

func (api *API) historyStatsHandler(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	var history types.BlockHeight
	// GET request so the only place the vars can be is the queryparams
	q := req.URL.Query()
	_, err := fmt.Sscan(q.Get("history"), &history)
	if err != nil {
		WriteError(w, Error{err.Error()}, http.StatusBadRequest)
		return
	}
	stats, err := api.explorer.HistoryStats(history)
	if err != nil {
		WriteError(w, Error{err.Error()}, http.StatusBadRequest)
		return
	}
	WriteJSON(w, stats)
}

func (api *API) rangeStatsHandler(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
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
	stats, err := api.explorer.RangeStats(start, end)
	if err != nil {
		WriteError(w, Error{err.Error()}, http.StatusBadRequest)
		return
	}
	WriteJSON(w, stats)
}
