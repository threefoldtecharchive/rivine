package api

import (
	"encoding/json"
	"net/http"

	"github.com/threefoldtech/rivine/build"
	"github.com/threefoldtech/rivine/modules"
	"github.com/threefoldtech/rivine/types"

	"github.com/julienschmidt/httprouter"
)

type (
	// TransactionPoolGET contains the fields returned by a GET call to "/transactionpool/transactions".
	TransactionPoolGET struct {
		Transactions []types.Transaction `json:"transactions"`
	}

	// TransactionPoolPOST is the success response for a POST to "/transactionpool/transactions".
	// It contains the the ID of the newly posted transaction.
	TransactionPoolPOST struct {
		TransactionID types.TransactionID `json:"transactionid"`
	}
)

// RegisterTransactionPoolHTTPHandlers registers the default Rivine handlers for all default Rivine TransactionPool HTTP endpoints.
func RegisterTransactionPoolHTTPHandlers(router Router, cs modules.ConsensusSet, tpool modules.TransactionPool, requiredPassword string) {
	if cs == nil {
		build.Critical("no consensus set module given")
	}
	if tpool == nil {
		build.Critical("no transaction pool module given")
	}
	if router == nil {
		build.Critical("no httprouter Router given")
	}
	router.GET("/transactionpool/transactions", NewTransactionPoolGetTransactionsHandler(cs, tpool))
	router.POST("/transactionpool/transactions", RequirePasswordHandler(NewTransactionPoolPostTransactionHandler(tpool), requiredPassword))
	router.OPTIONS("/transactionpool/transactions", RequirePasswordHandler(NewTransactionPoolOptionsTransactionHandler(), requiredPassword))
}

// NewTransactionPoolGetTransactionsHandler creates a handler
// to handle the API call to get the transaction pool transactions, filtered or not.
func NewTransactionPoolGetTransactionsHandler(cs modules.ConsensusSet, tpool modules.TransactionPool) httprouter.Handle {
	return func(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
		// get transactions
		txns := tpool.TransactionList()

		// get optional filter parameters
		q := req.URL.Query()
		str := q.Get("unlockhash")
		if str == "" {
			// if this parameter is not given, simply return all transactions
			WriteJSON(w, TransactionPoolGET{Transactions: txns})
			return
		}

		// parse unlockhash param as an actual UnlockHash
		var uh types.UnlockHash
		err := uh.LoadString(str)
		if err != nil {
			WriteError(w, Error{err.Error()}, http.StatusBadRequest)
			return
		}

		// filter based on unlock hash
	txnLoop:
		for i := 0; i < len(txns); {
			txn := txns[i]
			// try to find it either as the condition's unlockhash,
			// or as an unlockhash-property of a condition,
			// in other words where the unlockhash is the target
			for _, co := range txn.CoinOutputs {
				if isUnlockHashInCondition(uh, co.Condition) {
					i++
					continue txnLoop
				}
			}
			for _, bso := range txn.BlockStakeOutputs {
				if isUnlockHashInCondition(uh, bso.Condition) {
					i++
					continue txnLoop
				}
			}
			// try to find it the parent-condition's unlockhash,
			// or as an unlockhash-property of that parent-condition,
			// in other words where the unlockhash is the source
			for _, ci := range txn.CoinInputs {
				co, err := cs.GetCoinOutput(ci.ParentID)
				if err != nil {
					continue
				}
				if isUnlockHashInCondition(uh, co.Condition) {
					i++
					continue txnLoop
				}
			}
			for _, bsi := range txn.BlockStakeInputs {
				bso, err := cs.GetBlockStakeOutput(bsi.ParentID)
				if err != nil {
					continue
				}
				if isUnlockHashInCondition(uh, bso.Condition) {
					i++
					continue txnLoop
				}
			}
			// txn doesn't reference unlock hash
			txns = append(txns[:i], txns[i+1:]...)
		}

		// return filtered transactions
		WriteJSON(w, TransactionPoolGET{Transactions: txns})
		return
	}
}

func isUnlockHashInCondition(uh types.UnlockHash, co types.UnlockConditionProxy) bool {
	if uh == co.UnlockHash() {
		return true
	}
	switch tco := co.Condition.(type) {
	case *types.MultiSignatureCondition:
		for _, muh := range tco.UnlockHashes {
			if uh == muh {
				return true
			}
		}
	case *types.NilCondition, nil:
		return true
	}
	return false
}

// NewTransactionPoolPostTransactionHandler creates a handler to handle
// the API call to post a complete/valid transaction on /transactionpool/transactions
func NewTransactionPoolPostTransactionHandler(tpool modules.TransactionPool) httprouter.Handle {
	return func(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
		tx := types.Transaction{}

		if err := json.NewDecoder(req.Body).Decode(&tx); err != nil {
			WriteError(w, Error{"error decoding the supplied transaction: " + err.Error()}, http.StatusBadRequest)
			return
		}
		if err := tpool.AcceptTransactionSet([]types.Transaction{tx}); err != nil {
			WriteError(w, Error{"error after call to /wallet/transactions: " + err.Error()}, transactionPoolErrorToHTTPStatus(err))
			return
		}
		WriteJSON(w, TransactionPoolPOST{TransactionID: tx.ID()})
	}
}

// NewTransactionPoolOptionsTransactionHandler creates a handler to handle OPTIONS calls
func NewTransactionPoolOptionsTransactionHandler() httprouter.Handle {
	return func(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
		w.Header().Set("Access-Control-Allow-Methods", "*")
	}
}

func transactionPoolErrorToHTTPStatus(err error) int {
	if cErr, ok := err.(types.ClientError); ok {
		return cErr.Kind.AsHTTPStatusCode()
	}
	return http.StatusBadRequest
}
