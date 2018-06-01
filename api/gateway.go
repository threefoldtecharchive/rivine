package api

import (
	"net/http"

	"github.com/rivine/rivine/modules"

	"github.com/julienschmidt/httprouter"
)

// GatewayGET contains the fields returned by a GET call to "/gateway".
type GatewayGET struct {
	NetAddress modules.NetAddress `json:"netaddress"`
	Peers      []modules.Peer     `json:"peers"`
}

// gatewayHandler handles the API call asking for the gatway status.
func (api *API) gatewayHandler(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	peers := api.gateway.Peers()
	// nil slices are marshalled as 'null' in JSON, whereas 0-length slices are
	// marshalled as '[]'. The latter is preferred, indicating that the value
	// exists but contains no elements.
	if peers == nil {
		peers = make([]modules.Peer, 0)
	}
	WriteJSON(w, GatewayGET{api.gateway.Address(), peers})
}

// gatewayConnectHandler handles the API call to add a peer to the gateway.
func (api *API) gatewayConnectHandler(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	addr := modules.NetAddress(ps.ByName("netaddress"))
	// Try to resolve a possible (domain) name
	// Catching an error here is not particularly useful I feel, so ignore it
	addr.TryNameResolution()
	err := api.gateway.Connect(addr)
	if err != nil {
		WriteError(w, Error{err.Error()}, http.StatusBadRequest)
		return
	}

	WriteSuccess(w)
}

// gatewayDisconnectHandler handles the API call to remove a peer from the gateway.
func (api *API) gatewayDisconnectHandler(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	addr := modules.NetAddress(ps.ByName("netaddress"))
	// I don't feel like this is particularly useful here, but I suppose its nice to have nonetheless
	// Handeling a possible error is not really that useful
	addr.TryNameResolution()
	err := api.gateway.Disconnect(addr)
	if err != nil {
		WriteError(w, Error{err.Error()}, http.StatusBadRequest)
		return
	}

	WriteSuccess(w)
}
