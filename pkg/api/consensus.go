package api

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/threefoldtech/rivine/build"
	"github.com/threefoldtech/rivine/modules"
	"github.com/threefoldtech/rivine/types"

	"github.com/julienschmidt/httprouter"
)

// Go API Errors
var (
	// ErrNotFound is returned when a transaction is not found for a (short) id, but the ID itself is otherwise valid
	ErrNotFound = errors.New("transaction not found")
)

// HTTP API Errors
var (
	// ErrInvalidIDLength is returned when a supposed id does not have the correct length
	ErrInvalidIDLength = errors.New("ID does not have the right length")
)

// GetTransactionByShortID returns a transaction from the given short ID (if one exists)
func GetTransactionByShortID(cs modules.ConsensusSet, shortID string) (types.Transaction, error) {
	var txShortID types.TransactionShortID
	_, err := fmt.Sscan(shortID, &txShortID)
	if err != nil {
		return types.Transaction{}, err
	}

	txn, found := cs.TransactionAtShortID(txShortID)
	if !found {
		err = ErrNotFound
	}
	return txn, err
}

// GetTransactionByLongID returns a transaction from the given full transaction id (if one exists).
// It also returns the short id for future reference
func GetTransactionByLongID(cs modules.ConsensusSet, longid string) (types.Transaction, types.TransactionShortID, error) {
	var txID types.TransactionID
	err := txID.LoadString(longid)
	if err != nil {
		return types.Transaction{}, types.TransactionShortID(0), err
	}

	txn, txShortID, found := cs.TransactionAtID(txID)
	if !found {
		err = ErrNotFound
	}
	return txn, txShortID, err
}

type (
	// ConsensusGET contains general information about the consensus set, with tags
	// to support idiomatic json encodings.
	ConsensusGET struct {
		Synced       bool              `json:"synced"`
		Height       types.BlockHeight `json:"height"`
		CurrentBlock types.BlockID     `json:"currentblock"`
		Target       types.Target      `json:"target"`
	}

	// ConsensusGetTransaction is the object returned by a GET request to
	// /consensus/transaction/:id
	ConsensusGetTransaction struct {
		types.Transaction
		TxShortID types.TransactionShortID `json:"shortid,omitempty"`
	}

	// ConsensusGetUnspentCoinOutput is the object returned by a GET request to
	// /consensus/unspent/coinoutput/:id
	ConsensusGetUnspentCoinOutput struct {
		Output types.CoinOutput `json:"output"`
	}

	// ConsensusGetUnspentBlockstakeOutput is the object returned by a GET request to
	// /consensus/unspent/blockstakeoutput/:id
	ConsensusGetUnspentBlockstakeOutput struct {
		Output types.BlockStakeOutput `json:"output"`
	}
)

// RegisterConsensusHTTPHandlers registers the default Rivine handlers for all default Rivine Consensus HTTP endpoints.
func RegisterConsensusHTTPHandlers(router Router, cs modules.ConsensusSet) {
	if cs == nil {
		build.Critical("no consensus module given")
	}
	if router == nil {
		build.Critical("no httprouter Router given")
	}

	router.GET("/consensus", NewConsensusRootHandler(cs))
	router.GET("/consensus/transactions/:id", NewConsensusGetTransactionHandler(cs))
	router.GET("/consensus/unspent/coinoutputs/:id", NewConsensusGetUnspentCoinOutputHandler(cs))
	router.GET("/consensus/unspent/blockstakeoutputs/:id", NewConsensusGetUnspentBlockstakeOutputHandler(cs))
}

// NewConsensusRootHandler creates a handler to handle the API calls to /consensus.
func NewConsensusRootHandler(cs modules.ConsensusSet) httprouter.Handle {
	return func(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
		cbid := cs.CurrentBlock().ID()
		currentTarget, _ := cs.ChildTarget(cbid)
		WriteJSON(w, ConsensusGET{
			Synced:       cs.Synced(),
			Height:       cs.Height(),
			CurrentBlock: cbid,
			Target:       currentTarget,
		})
	}
}

// NewConsensusGetTransactionHandler creates a handler to handle lookups of a transaction based on a short or long ID.
func NewConsensusGetTransactionHandler(cs modules.ConsensusSet) httprouter.Handle {
	return func(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
		id := ps.ByName("id")
		idLen := len(id)

		var cgt ConsensusGetTransaction
		var err error

		switch {
		// Check if this is a short id
		case idLen <= 19:
			cgt.Transaction, err = GetTransactionByShortID(cs, id)
		// regular id
		case idLen == 64:
			cgt.Transaction, cgt.TxShortID, err = GetTransactionByLongID(cs, id)
		default:
			err = ErrInvalidIDLength
		}

		if err != nil {
			if err == ErrNotFound {
				WriteError(w, Error{err.Error()}, http.StatusNoContent)
				return
			}
			WriteError(w, Error{err.Error()}, http.StatusBadRequest)
			return
		}

		WriteJSON(w, cgt)
	}
}

// NewConsensusGetUnspentCoinOutputHandler creates a handler to handle lookups of unspent coin outputs.
func NewConsensusGetUnspentCoinOutputHandler(cs modules.ConsensusSet) httprouter.Handle {
	return func(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
		var (
			outputID types.CoinOutputID
			id       = ps.ByName("id")
		)

		if len(id) != len(outputID)*2 {
			WriteError(w, Error{ErrInvalidIDLength.Error()}, http.StatusBadRequest)
			return
		}

		err := outputID.LoadString(id)
		if err != nil {
			WriteError(w, Error{err.Error()}, http.StatusBadRequest)
			return
		}

		output, err := cs.GetCoinOutput(outputID)
		if err != nil {
			WriteError(w, Error{err.Error()}, http.StatusNoContent)
			return
		}
		WriteJSON(w, ConsensusGetUnspentCoinOutput{Output: output})
	}
}

// NewConsensusGetUnspentBlockstakeOutputHandler creates a handler to handle lookups of unspent blockstake outputs
func NewConsensusGetUnspentBlockstakeOutputHandler(cs modules.ConsensusSet) httprouter.Handle {
	return func(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
		var (
			outputID types.BlockStakeOutputID
			id       = ps.ByName("id")
		)

		if len(id) != len(outputID)*2 {
			WriteError(w, Error{ErrInvalidIDLength.Error()}, http.StatusBadRequest)
			return
		}

		err := outputID.LoadString(id)
		if err != nil {
			WriteError(w, Error{err.Error()}, http.StatusBadRequest)
			return
		}

		output, err := cs.GetBlockStakeOutput(outputID)
		if err != nil {
			WriteError(w, Error{err.Error()}, http.StatusNoContent)
			return
		}
		WriteJSON(w, ConsensusGetUnspentBlockstakeOutput{Output: output})
	}
}
