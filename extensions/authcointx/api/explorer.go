package api

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/julienschmidt/httprouter"
	"github.com/threefoldtech/rivine/extensions/authcointx"
	"github.com/threefoldtech/rivine/modules"
	rapi "github.com/threefoldtech/rivine/pkg/api"
	"github.com/threefoldtech/rivine/types"
)

// GetAuthConditionResponse contains a requested auth condition,
// either the current active one active for the given blockheight or lower.
type GetAuthConditionResponse struct {
	AuthCondition types.UnlockConditionProxy `json:"authcondition"`
}

// GetAddressesAuthStateResponse contains a requested auth state for the requested addresses,
// either the current active one active for the given blockheight or lower.
type GetAddressesAuthStateResponse struct {
	AuthStates []bool `json:"auths"`
}

// RegisterConsensusAuthCoinHTTPHandlers registers the default Rivine handlers for all default Rivine Consensus HTTP endpoints.
func RegisterConsensusAuthCoinHTTPHandlers(router rapi.Router, plugin *authcointx.Plugin, txpool modules.TransactionPool, authConditionTxVersion, authAddressTxVersion types.TransactionVersion) {
	registerAuthCoinHTTPHandlers(router, "/consensus", plugin, txpool, authConditionTxVersion, authAddressTxVersion)
}

// RegisterExplorerAuthCoinHTTPHandlers registers the default Rivine handlers for all default Rivine Explorer HTTP endpoints.
func RegisterExplorerAuthCoinHTTPHandlers(router rapi.Router, plugin *authcointx.Plugin, txpool modules.TransactionPool, authConditionTxVersion, authAddressTxVersion types.TransactionVersion) {
	registerAuthCoinHTTPHandlers(router, "/explorer", plugin, txpool, authConditionTxVersion, authAddressTxVersion)
}

func registerAuthCoinHTTPHandlers(router rapi.Router, root string, plugin *authcointx.Plugin, txpool modules.TransactionPool, authConditionTxVersion, authAddressTxVersion types.TransactionVersion) {
	router.GET(root+"/authcoin/condition", NewGetActiveAuthConditionHandler(plugin, txpool, authConditionTxVersion))
	router.GET(root+"/authcoin/condition/:height", NewGetAuthConditionAtHandler(plugin))
	router.GET(root+"/authcoin/status", NewGetAddressesAuthStateHandler(plugin, txpool, authAddressTxVersion))
}

// NewGetActiveAuthConditionHandler creates a handler to handle the API calls to /explorer/authcoin/condition.
func NewGetActiveAuthConditionHandler(plugin *authcointx.Plugin, txpool modules.TransactionPool, authConditionTxVersion types.TransactionVersion) httprouter.Handle {
	return func(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
		// do we accept txpool
		q := req.URL.Query()
		txPoolAllowed := true
		txPoolAllowedStr := q.Get("tzxpool")
		if txpool == nil || (txPoolAllowedStr != "" && (txPoolAllowedStr == "0" || strings.ToLower(txPoolAllowedStr) == "false")) {
			txPoolAllowed = false
		}
		if txPoolAllowed {
			// look through txpool first
			txns := txpool.TransactionList()
			txnsOffset := len(txns) - 1
			var txn *types.Transaction
			for i := range txns {
				txn = &txns[txnsOffset-i]
				if txn.Version != authConditionTxVersion {
					continue
				}
				atxn, err := authcointx.AuthConditionUpdateTransactionFromTransaction(*txn, authConditionTxVersion, plugin.RequireMinerFees)
				if err != nil {
					rapi.WriteError(w, rapi.Error{Message: err.Error()}, http.StatusInternalServerError)
					return
				}
				rapi.WriteJSON(w, GetAuthConditionResponse{
					AuthCondition: atxn.AuthCondition,
				})
				return
			}
		}
		// use confirmed later
		authCondition, err := plugin.GetActiveAuthCondition()
		if err != nil {
			rapi.WriteError(w, rapi.Error{Message: err.Error()}, http.StatusInternalServerError)
			return
		}
		rapi.WriteJSON(w, GetAuthConditionResponse{
			AuthCondition: authCondition,
		})
	}
}

// NewGetAuthConditionAtHandler creates a handler to handle the API calls to /explorer/authcoin/condition/:height.
func NewGetAuthConditionAtHandler(plugin *authcointx.Plugin) httprouter.Handle {
	return func(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
		heightStr := ps.ByName("height")
		height, err := strconv.ParseUint(heightStr, 10, 64)
		if err != nil {
			rapi.WriteError(w, rapi.Error{Message: fmt.Sprintf("invalid block height given: %v", err)}, http.StatusBadRequest)
			return
		}
		authCondition, err := plugin.GetAuthConditionAt(types.BlockHeight(height))
		if err != nil {
			rapi.WriteError(w, rapi.Error{Message: err.Error()}, http.StatusInternalServerError)
			return
		}
		rapi.WriteJSON(w, GetAuthConditionResponse{
			AuthCondition: authCondition,
		})
	}
}

// NewGetAddressesAuthStateHandler creates a handler to handle the API calls to /<root>/authcoin/status.
func NewGetAddressesAuthStateHandler(plugin *authcointx.Plugin, txpool modules.TransactionPool, authAddressTxVersion types.TransactionVersion) httprouter.Handle {
	return func(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
		q := req.URL.Query()

		var (
			err       error
			addresses []types.UnlockHash
			resp      GetAddressesAuthStateResponse
		)

		// parse all addresses
		addressStrings, ok := q["addr"]
		if !ok {
			rapi.WriteError(w, rapi.Error{Message: "no address given as query parameter while at least one is required"}, http.StatusBadRequest)
			return
		}
		addresses = make([]types.UnlockHash, len(addressStrings))
		for idx, addressStr := range addressStrings {
			err = addresses[idx].LoadString(addressStr)
			if err != nil {
				rapi.WriteError(w, rapi.Error{Message: fmt.Sprintf("invalid address %s (q#%d) given: %v", addressStr, idx, err)}, http.StatusBadRequest)
				return
			}
		}

		// get state slice now or in the past
		if heightStr := q.Get("height"); heightStr != "" {
			height, err := strconv.ParseUint(heightStr, 10, 64)
			if err != nil {
				rapi.WriteError(w, rapi.Error{Message: fmt.Sprintf("invalid block height given: %v", err)}, http.StatusBadRequest)
				return
			}

			resp.AuthStates, err = plugin.GetAddressesAuthStateAt(types.BlockHeight(height), addresses, nil)
			if err != nil {
				rapi.WriteError(w, rapi.Error{Message: err.Error()}, http.StatusInternalServerError)
				return
			}
		} else {
			// do we accept txpool
			q := req.URL.Query()
			txPoolAllowed := true
			txPoolAllowedStr := q.Get("txpool")
			if txpool == nil || (txPoolAllowedStr != "" && (txPoolAllowedStr == "0" || strings.ToLower(txPoolAllowedStr) == "false")) {
				txPoolAllowed = false
			}
			if txPoolAllowed {
				resp.AuthStates, err = getAddressesAuthStateNowAcceptingTxPool(addresses, plugin, txpool, authAddressTxVersion)
			} else {
				resp.AuthStates, err = plugin.GetAddressesAuthStateNow(addresses, nil)
			}
			if err != nil {
				rapi.WriteError(w, rapi.Error{Message: err.Error()}, http.StatusInternalServerError)
				return
			}
		}

		// return successfull response
		rapi.WriteJSON(w, resp)
	}
}

func getAddressesAuthStateNowAcceptingTxPool(addresses []types.UnlockHash, plugin *authcointx.Plugin, txpool modules.TransactionPool, authAddressTxVersion types.TransactionVersion) ([]bool, error) {
	states, err := plugin.GetAddressesAuthStateNow(addresses, nil)
	if err != nil {
		return nil, err
	}
	// look through txpool first
	txns := txpool.TransactionList()
	addressMapping := map[types.UnlockHash]int{}
	for idx, address := range addresses {
		addressMapping[address] = idx
	}
	for _, txn := range txns {
		if txn.Version != authAddressTxVersion {
			continue
		}
		atxn, err := authcointx.AuthAddressUpdateTransactionFromTransaction(txn, authAddressTxVersion, plugin.RequireMinerFees)
		if err != nil {
			return nil, err
		}
		for _, uh := range atxn.AuthAddresses {
			if idx, ok := addressMapping[uh]; ok {
				states[idx] = true
			}
		}
		for _, uh := range atxn.DeauthAddresses {
			if idx, ok := addressMapping[uh]; ok {
				states[idx] = false
			}
		}
	}
	return states, nil
}
