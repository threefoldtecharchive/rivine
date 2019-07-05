package api

import (
	"encoding/json"
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

// GetAddressAuthStateResponse contains a requested auth state for the address,
// either the current active one active for the given blockheight or lower.
type GetAddressAuthStateResponse struct {
	AuthState bool `json:"auth"`
}

// GetAddressesAuthState contains addresses used as input to request
// for all the given addresses the current active or if given a block height
// at that given block height or lower.
type GetAddressesAuthState struct {
	Addresses []types.UnlockHash `json:"addresses"`
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
	router.GET(root+"/authcoin/address/:address", NewGetAddressAuthStateNowHandler(plugin))
	router.GET(root+"/authcoin/address/:address/:height", NewGetAddressAuthStateAtHeightHandler(plugin))
	router.GET(root+"/authcoin/addresses", NewGetAddressesAuthStateNowHandler(plugin))
	router.GET(root+"/authcoin/addresses/:height", NewGetAddressesAuthStateAtHeightHandler(plugin))
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

// NewGetAddressAuthStateNowHandler creates a handler to handle the API calls to /explorer/authcoin/address/:address.
func NewGetAddressAuthStateNowHandler(plugin *authcointx.Plugin) httprouter.Handle {
	return func(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
		address := ps.ByName("address")
		var uh types.UnlockHash
		err := uh.LoadString(address)
		if err != nil {
			rapi.WriteError(w, rapi.Error{Message: err.Error()}, http.StatusBadRequest)
			return
		}
		stateSlice, err := plugin.GetAddressesAuthStateNow([]types.UnlockHash{uh}, nil)
		if err != nil {
			rapi.WriteError(w, rapi.Error{Message: err.Error()}, http.StatusInternalServerError)
			return
		}
		rapi.WriteJSON(w, GetAddressAuthStateResponse{
			AuthState: stateSlice[0],
		})
	}
}

// NewGetAddressAuthStateAtHeightHandler creates a handler to handle the API calls to /explorer/authcoin/address/:address/:height.
func NewGetAddressAuthStateAtHeightHandler(plugin *authcointx.Plugin) httprouter.Handle {
	return func(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
		heightStr := ps.ByName("height")
		height, err := strconv.ParseUint(heightStr, 10, 64)
		if err != nil {
			rapi.WriteError(w, rapi.Error{Message: fmt.Sprintf("invalid block height given: %v", err)}, http.StatusBadRequest)
			return
		}

		address := ps.ByName("address")
		var uh types.UnlockHash
		err = uh.LoadString(address)
		if err != nil {
			rapi.WriteError(w, rapi.Error{Message: err.Error()}, http.StatusBadRequest)
			return
		}
		stateSlice, err := plugin.GetAddressesAuthStateAt(types.BlockHeight(height), []types.UnlockHash{uh}, nil)
		if err != nil {
			rapi.WriteError(w, rapi.Error{Message: err.Error()}, http.StatusInternalServerError)
			return
		}
		rapi.WriteJSON(w, GetAddressAuthStateResponse{
			AuthState: stateSlice[0],
		})
	}
}

// NewGetAddressesAuthStateNowHandler creates a handler to handle the API calls to /explorer/authcoin/addresses
func NewGetAddressesAuthStateNowHandler(plugin *authcointx.Plugin) httprouter.Handle {
	return func(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
		var body GetAddressesAuthState
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			rapi.WriteError(w, rapi.Error{Message: "error decoding the supplied addresses: " + err.Error()}, http.StatusBadRequest)
			return
		}
		stateSlice, err := plugin.GetAddressesAuthStateNow(body.Addresses, nil)
		if err != nil {
			rapi.WriteError(w, rapi.Error{Message: err.Error()}, http.StatusInternalServerError)
			return
		}
		rapi.WriteJSON(w, GetAddressesAuthStateResponse{
			AuthStates: stateSlice,
		})
	}
}

// NewGetAddressesAuthStateAtHeightHandler creates a handler to handle the API calls to /explorer/authcoin/addresses/:height.
func NewGetAddressesAuthStateAtHeightHandler(plugin *authcointx.Plugin) httprouter.Handle {
	return func(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
		heightStr := ps.ByName("height")
		height, err := strconv.ParseUint(heightStr, 10, 64)
		if err != nil {
			rapi.WriteError(w, rapi.Error{Message: fmt.Sprintf("invalid block height given: %v", err)}, http.StatusBadRequest)
			return
		}

		var body GetAddressesAuthState
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			rapi.WriteError(w, rapi.Error{Message: "error decoding the supplied addresses: " + err.Error()}, http.StatusBadRequest)
			return
		}
		stateSlice, err := plugin.GetAddressesAuthStateAt(types.BlockHeight(height), body.Addresses, nil)
		if err != nil {
			rapi.WriteError(w, rapi.Error{Message: err.Error()}, http.StatusInternalServerError)
			return
		}
		rapi.WriteJSON(w, GetAddressesAuthStateResponse{
			AuthStates: stateSlice,
		})
	}
}
