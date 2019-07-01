package api

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
	"github.com/threefoldtech/rivine/extensions/minting"
	rapi "github.com/threefoldtech/rivine/pkg/api"
	"github.com/threefoldtech/rivine/types"
)

// TransactionDBGetMintCondition contains a requested mint condition,
// either the current active one active for the given blockheight or lower.
type TransactionDBGetMintCondition struct {
	MintCondition types.UnlockConditionProxy `json:"mintcondition"`
}

// RegisterExplorerMintingHTTPHandlers registers the default Rivine handlers for all default Rivine Explprer HTTP endpoints.
func RegisterExplorerMintingHTTPHandlers(router rapi.Router, plugin *minting.Plugin) {
	router.GET("/explorer/mintcondition", NewTransactionDBGetActiveMintConditionHandler(plugin))
	router.GET("/explorer/mintcondition/:height", NewTransactionDBGetMintConditionAtHandler(plugin))
}

// RegisterConsensusMintingHTTPHandlers registers the default Rivine handlers for all default Rivine Explprer HTTP endpoints.
func RegisterConsensusMintingHTTPHandlers(router rapi.Router, plugin *minting.Plugin) {
	router.GET("/consensus/mintcondition", NewTransactionDBGetActiveMintConditionHandler(plugin))
	router.GET("/consensus/mintcondition/:height", NewTransactionDBGetMintConditionAtHandler(plugin))
}

// NewTransactionDBGetActiveMintConditionHandler creates a handler to handle the API calls to /transactiondb/mintcondition.
func NewTransactionDBGetActiveMintConditionHandler(plugin *minting.Plugin) httprouter.Handle {
	return func(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
		mintCondition, err := plugin.GetActiveMintCondition()
		if err != nil {
			rapi.WriteError(w, rapi.Error{Message: err.Error()}, http.StatusInternalServerError)
			return
		}
		rapi.WriteJSON(w, TransactionDBGetMintCondition{
			MintCondition: mintCondition,
		})
	}
}

// NewTransactionDBGetMintConditionAtHandler creates a handler to handle the API calls to /transactiondb/mintcondition/:height.
func NewTransactionDBGetMintConditionAtHandler(plugin *minting.Plugin) httprouter.Handle {
	return func(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
		heightStr := ps.ByName("height")
		height, err := strconv.ParseUint(heightStr, 10, 64)
		if err != nil {
			rapi.WriteError(w, rapi.Error{Message: fmt.Sprintf("invalid block height given: %v", err)}, http.StatusBadRequest)
			return
		}
		mintCondition, err := plugin.GetMintConditionAt(types.BlockHeight(height))
		if err != nil {
			rapi.WriteError(w, rapi.Error{Message: err.Error()}, http.StatusInternalServerError)
			return
		}
		rapi.WriteJSON(w, TransactionDBGetMintCondition{
			MintCondition: mintCondition,
		})
	}
}
