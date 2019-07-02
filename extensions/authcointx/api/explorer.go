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

// GetAuthCondition contains a requested auth condition,
// either the current active one active for the given blockheight or lower.
type GetAuthCondition struct {
	AuthCondition types.UnlockConditionProxy `json:"authcondition"`
}

// GetAddressAuthState contains a requested auth state for the address,
// either the current active one active for the given blockheight or lower.
type GetAddressAuthState struct {
	Address   types.UnlockHash `json:"unlockhash"`
	AuthState bool             `json:"auth"`
}

// RegisterConsensusuthCoinHTTPHandlers registers the default Rivine handlers for all default Rivine Explprer HTTP endpoints.
func RegisterConsensusuthCoinHTTPHandlers(router rapi.Router, plugin *authcointx.Plugin) {
	router.GET("/consensus/authcoin/condition", NewGetActiveAuthConditionHandler(plugin))
	router.GET("/consensus/authcoin/condition/:height", NewGetAuthConditionAtHandler(plugin))
	router.GET("/consensus/authcoin/address/:address", NewGetAddressAuthStateNowHandler(plugin))
	router.GET("/consensus/authcoin/address/:address/:height", NewGetAddressAuthStateAtHeightHandler(plugin))
}

// RegisterExplorerAuthCoinHTTPHandlers registers the default Rivine handlers for all default Rivine Explprer HTTP endpoints.
func RegisterExplorerAuthCoinHTTPHandlers(router rapi.Router, plugin *authcointx.Plugin) {
	router.GET("/explorer/authcoin/condition", NewGetActiveAuthConditionHandler(plugin))
	router.GET("/explorer/authcoin/condition/:height", NewGetAuthConditionAtHandler(plugin))
	router.GET("/explorer/authcoin/address/:address", NewGetAddressAuthStateNowHandler(plugin))
	router.GET("/explorer/authcoin/address/:address/:height", NewGetAddressAuthStateAtHeightHandler(plugin))
}

// NewGetActiveAuthConditionHandler creates a handler to handle the API calls to /explorer/authcoin/condition.
func NewGetActiveAuthConditionHandler(plugin *authcointx.Plugin) httprouter.Handle {
	return func(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
		authCondition, err := plugin.GetActiveAuthCondition()
		if err != nil {
			rapi.WriteError(w, rapi.Error{Message: err.Error()}, http.StatusInternalServerError)
			return
		}
		rapi.WriteJSON(w, GetAuthCondition{
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
		rapi.WriteJSON(w, GetAuthCondition{
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
		err = plugin.EnsureAddressesAreAuthNow(uh)
		rapi.WriteJSON(w, GetAddressAuthState{
			Address:   uh,
			AuthState: err == nil,
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

		err = plugin.EnsureAddressesAreAuthAt(types.BlockHeight(height), uh)
		rapi.WriteJSON(w, GetAddressAuthState{
			Address:   uh,
			AuthState: err == nil,
		})
	}
}
