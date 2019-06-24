package api

import (
	"net/http"

	"github.com/threefoldtech/rivine/build"
	"github.com/threefoldtech/rivine/modules"

	"github.com/julienschmidt/httprouter"
)

// GatewayGET contains the fields returned by a GET call to "/gateway".
type GatewayGET struct {
	NetAddress modules.NetAddress `json:"netaddress"`
	Peers      []modules.Peer     `json:"peers"`
}

// RegisterGatewayHTTPHandlers registers the default Rivine handlers for all default Rivine Gateway HTTP endpoints.
func RegisterGatewayHTTPHandlers(router Router, gateway modules.Gateway, requiredPassword string) {
	if gateway == nil {
		build.Critical("no gateway module given")
	}
	if router == nil {
		build.Critical("no httprouter Router given")
	}
	router.GET("/gateway", NewGatewayRootHandler(gateway))
	router.POST("/gateway/connect/:netaddress", RequirePasswordHandler(NewGatewayConnectHandler(gateway), requiredPassword))
	router.POST("/gateway/disconnect/:netaddress", RequirePasswordHandler(NewGatewayDisconnectHandler(gateway), requiredPassword))
}

// NewGatewayRootHandler creates a handler to handle the API call asking for the gatway status.
func NewGatewayRootHandler(gateway modules.Gateway) httprouter.Handle {
	return func(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
		peers := gateway.Peers()
		// nil slices are marshalled as 'null' in JSON, whereas 0-length slices are
		// marshalled as '[]'. The latter is preferred, indicating that the value
		// exists but contains no elements.
		if peers == nil {
			peers = make([]modules.Peer, 0)
		}
		WriteJSON(w, GatewayGET{gateway.Address(), peers})
	}
}

// NewGatewayConnectHandler creates a handler to handle the API call to add a peer to the gateway.
func NewGatewayConnectHandler(gateway modules.Gateway) httprouter.Handle {
	return func(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
		addr := modules.NetAddress(ps.ByName("netaddress"))
		// Try to resolve a possible (domain) name
		// Catching an error here is not particularly useful I feel, so ignore it
		addr.TryNameResolution()
		err := gateway.Connect(addr)
		if err != nil {
			WriteError(w, Error{err.Error()}, http.StatusBadRequest)
			return
		}
		WriteSuccess(w)
	}
}

// NewGatewayDisconnectHandler creates a handler to handle the API call to remove a peer from the gateway.
func NewGatewayDisconnectHandler(gateway modules.Gateway) httprouter.Handle {
	return func(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
		addr := modules.NetAddress(ps.ByName("netaddress"))
		// I don't feel like this is particularly useful here, but I suppose its nice to have nonetheless
		// Handeling a possible error is not really that useful
		addr.TryNameResolution()
		err := gateway.Disconnect(addr)
		if err != nil {
			WriteError(w, Error{err.Error()}, http.StatusBadRequest)
			return
		}
		WriteSuccess(w)
	}
}
