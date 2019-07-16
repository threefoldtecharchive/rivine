package api

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
	"github.com/threefoldtech/rivine/extensions/authcointx"
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

// RegisterConsensusuthCoinHTTPHandlers registers the default Rivine handlers for all default Rivine Explprer HTTP endpoints.
func RegisterConsensusuthCoinHTTPHandlers(router rapi.Router, plugin *authcointx.Plugin) {
	registerAuthCoinHTTPHandlers(router, "/consensus", plugin)
}

// RegisterExplorerAuthCoinHTTPHandlers registers the default Rivine handlers for all default Rivine Explprer HTTP endpoints.
func RegisterExplorerAuthCoinHTTPHandlers(router rapi.Router, plugin *authcointx.Plugin) {
	registerAuthCoinHTTPHandlers(router, "/explorer", plugin)
}

func registerAuthCoinHTTPHandlers(router rapi.Router, root string, plugin *authcointx.Plugin) {
	router.GET(root+"/authcoin/condition", NewGetActiveAuthConditionHandler(plugin))
	router.GET(root+"/authcoin/condition/:height", NewGetAuthConditionAtHandler(plugin))
	router.GET(root+"/authcoin/status", NewGetAddressesAuthStateHandler(plugin))
}

// NewGetActiveAuthConditionHandler creates a handler to handle the API calls to /explorer/authcoin/condition.
func NewGetActiveAuthConditionHandler(plugin *authcointx.Plugin) httprouter.Handle {
	return func(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
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
func NewGetAddressesAuthStateHandler(plugin *authcointx.Plugin) httprouter.Handle {
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
			resp.AuthStates, err = plugin.GetAddressesAuthStateNow(addresses, nil)
			if err != nil {
				rapi.WriteError(w, rapi.Error{Message: err.Error()}, http.StatusInternalServerError)
				return
			}
		}

		// return successfull response
		rapi.WriteJSON(w, resp)
	}
}
