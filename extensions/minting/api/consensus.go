package api

import (
	"github.com/threefoldtech/rivine/extensions/minting"
	rapi "github.com/threefoldtech/rivine/pkg/api"
)

// RegisterConsensusMintingHTTPHandlers registers the default Rivine handlers for all default Rivine consensus HTTP endpoints.
func RegisterConsensusMintingHTTPHandlers(router rapi.Router, plugin *minting.Plugin) {
	router.GET("/consensus/mintcondition", NewTransactionDBGetActiveMintConditionHandler(plugin))
	router.GET("/consensus/mintcondition/:height", NewTransactionDBGetMintConditionAtHandler(plugin))
}
