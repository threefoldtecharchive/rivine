package main

import (
	"github.com/rivine/rivine/build"
	"github.com/rivine/rivine/pkg/client"
	"github.com/rivine/rivine/types"
)

func main() {
	bchainInfo := types.DefaultBlockchainInfo()
	client.DefaultCLIClient("", bchainInfo.Name, func(icfg *client.Config) client.Config {
		var networkName string
		switch build.Release {
		case "dev":
			networkName = "devnet"
		case "testing":
			networkName = "testnet"
		default:
			networkName = "standard"
		}
		if icfg == nil {
			constants := types.DefaultChainConstants()
			return client.Config{
				ChainName:    bchainInfo.Name,
				NetworkName:  networkName,
				ChainVersion: bchainInfo.ChainVersion,

				CurrencyUnits:             constants.CurrencyUnits,
				MinimumTransactionFee:     constants.MinimumTransactionFee,
				DefaultTransactionVersion: constants.DefaultTransactionVersion,

				BlockFrequencyInSeconds: int64(constants.BlockFrequency),
				GenesisBlockTimestamp:   constants.GenesisBlock().Timestamp,
			}
		}
		return *icfg
	})
}
