package main

import (
	"github.com/rivine/rivine/modules"
	"github.com/rivine/rivine/pkg/client"
	"github.com/rivine/rivine/types"
)

func main() {
	bchainInfo := types.DefaultBlockchainInfo()
	client.DefaultCLIClient("", bchainInfo.Name, func(icfg *client.Config) client.Config {
		if icfg != nil {
			return *icfg
		}
		chainConstants := types.DefaultChainConstants()
		daemonConstants := modules.NewDaemonConstants(bchainInfo, chainConstants)
		return client.ConfigFromDaemonConstants(daemonConstants)
	})
}
