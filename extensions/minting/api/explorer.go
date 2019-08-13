package api

import (
	"github.com/threefoldtech/rivine/extensions/minting"
	rapi "github.com/threefoldtech/rivine/pkg/api"
)

// RegisterExplorerMintingHTTPHandlers registers the default Rivine handlers for all default Rivine explorer endpoints.
func RegisterExplorerMintingHTTPHandlers(router rapi.Router, plugin *minting.Plugin) {
	router.GET("/explorer/mintcondition", NewTransactionDBGetActiveMintConditionHandler(plugin))
	router.GET("/explorer/mintcondition/:height", NewTransactionDBGetMintConditionAtHandler(plugin))
}
