package main

import (
	"github.com/rivine/rivine/pkg/client"
	"github.com/rivine/rivine/types"
	"github.com/threefoldfoundation/tfchain/pkg/config"
)

func main() {
	client.DefaultCLIClient("", func(icfg *client.Config) client.Config {
		if icfg == nil {
			bchainInfo := types.DefaultBlockchainInfo()
			constants := types.DefaultChainConstants()
			return client.Config{
				ChainName:    bchainInfo.Name,
				NetworkName:  config.NetworkNameStandard,
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
