package api

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/rivine/rivine/types"

	"github.com/julienschmidt/httprouter"
)

var (
	// errNotFound is returned when a transaction is not found for a (short) id, but the ID itself is otherwise valid
	errNotFound = errors.New("Transaction not found")
	// errInvalidIDLength is returned when a supposed transaction id does not have the length to be either a short or standard transaction id
	errInvalidIDLength = errors.New("ID does not have the right length")
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
// /consensus/transaction/:id
type ConsensusGetTransaction struct {
	types.Transaction
	TxShortID types.TransactionShortID `json:"shortid,omitempty"`
}

// consensusGetTransactionHandler handles lookups
func (api *API) consensusGetTransactionHandler(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	id := ps.ByName("id")
	idLen := len(id)

	var cgt ConsensusGetTransaction
	var err error

	switch {
	// Check if this is a short id
	case idLen <= 19:
		cgt.Transaction, err = api.getTransactionByShortID(id)
	// regular id
	case idLen == 64:
		cgt.Transaction, cgt.TxShortID, err = api.getTransactionByLongID(id)
	default:
		err = errInvalidIDLength
	}

	if err != nil {
		if err == errNotFound {
			WriteError(w, Error{err.Error()}, http.StatusNoContent)
			return
		}
		WriteError(w, Error{err.Error()}, http.StatusBadRequest)
		return
	}

	WriteJSON(w, cgt)
}

// getTransactionByShortID returns a transaction from the given short ID (if one exists)
func (api *API) getTransactionByShortID(shortID string) (types.Transaction, error) {
	var txShortID types.TransactionShortID
	_, err := fmt.Sscan(shortID, &txShortID)
	if err != nil {
		return types.Transaction{}, err
	}

	txn, found := api.cs.TransactionAtShortID(txShortID)
	if !found {
		err = errNotFound
	}
	return txn, err
}

// getTransactionByLongID returns a transaction from the given full transaction id (if one exists).
// It also returns the short id for future reference
func (api *API) getTransactionByLongID(longid string) (types.Transaction, types.TransactionShortID, error) {
	var txID types.TransactionID
	err := txID.LoadString(longid)
	if err != nil {
		return types.Transaction{}, types.TransactionShortID(0), err
	}

	txn, txShortID, found := api.cs.TransactionAtID(txID)
	if !found {
		err = errNotFound
	}
	return txn, txShortID, err
}
