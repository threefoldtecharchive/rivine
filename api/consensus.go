package api

import (
	"fmt"
	"net/http"

	"github.com/rivine/rivine/types"

	"github.com/julienschmidt/httprouter"
)

// ConsensusGET contains general information about the consensus set, with tags
// to support idiomatic json encodings.
type ConsensusGET struct {
	Synced       bool              `json:"synced"`
	Height       types.BlockHeight `json:"height"`
	CurrentBlock types.BlockID     `json:"currentblock"`
	Target       types.Target      `json:"target"`
}

// consensusHandler handles the API calls to /consensus.
func (api *API) consensusHandler(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	cbid := api.cs.CurrentBlock().ID()
	currentTarget, _ := api.cs.ChildTarget(cbid)
	WriteJSON(w, ConsensusGET{
		Synced:       api.cs.Synced(),
		Height:       api.cs.Height(),
		CurrentBlock: cbid,
		Target:       currentTarget,
	})
}

// ConsensusGetTransaction is the object returned by a GET request to
// /consensus/transaction/:shortid.
type ConsensusGetTransaction struct {
	ExplorerTransaction
}

func (api *API) consensusGetTransactionHandler(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	var txShortID types.TransactionShortID
	_, err := fmt.Sscan(ps.ByName("shortid"), &txShortID)
	if err != nil {
		WriteError(w, Error{err.Error()}, http.StatusBadRequest)
		return
	}

	txn, found := api.cs.TransactionAtShortID(txShortID)
	if !found {
		WriteError(w, Error{fmt.Sprintf("no tx found for shortID %v", txShortID)},
			http.StatusNoContent)
		return
	}

	height := txShortID.BlockHeight()
	block, _ := api.cs.BlockAtHeight(height)

	explorerTxn := api.buildExplorerTransaction(height, block.ID(), txn)
	WriteJSON(w, ConsensusGetTransaction{
		ExplorerTransaction: explorerTxn,
	})
}
