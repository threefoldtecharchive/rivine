package api

import (
	"encoding/json"
	"net/http"

	"github.com/rivine/rivine/types"

	"github.com/julienschmidt/httprouter"
)

type TransactionPoolGET struct {
	Transactions []types.Transaction `json:"transactions"`
}

// transactionpoolTransactionsHandler handles the API call to get the
// transaction pool transactions.
func (api *API) transactionpoolTransactionsHandler(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	WriteJSON(w, TransactionPoolGET{Transactions: api.tpool.TransactionList()})
}

// TransactionPoolPOST is the success response for a POST to /transactionpool/transactions.
// It is the ID of the newly posted transaction
type TransactionPoolPOST struct {
	TransactionID types.TransactionID `json:"transactionid"`
}

// transactionpoolPostTransactionHandler handles the API call to post a complete transaction on /transactionpool/transactions
func (api *API) transactionpoolPostTransactionHandler(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	tx := types.Transaction{}

	if err := json.NewDecoder(req.Body).Decode(&tx); err != nil {
		WriteError(w, Error{"error decoding the supplied transaction: " + err.Error()}, http.StatusBadRequest)
		return
	}
	if err := api.tpool.AcceptTransactionSet([]types.Transaction{tx}); err != nil {
		WriteError(w, Error{"error after call to /wallet/transactions: " + err.Error()}, http.StatusBadRequest)
		return
	}
	WriteJSON(w, TransactionPoolPOST{TransactionID: tx.ID()})
}
