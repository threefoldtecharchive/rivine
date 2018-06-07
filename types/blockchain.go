package types

import (
	"github.com/rivine/rivine/build"
)

// BlockchainInfo contains information about a blockchain.
type BlockchainInfo struct {
	Name            string
	NetworkName     string
	CoinUnit        string
	ChainVersion    build.ProtocolVersion
	ProtocolVersion build.ProtocolVersion
}

// DefaultBlockchainInfo returns the blockchain information
// for the default (Rivine) blockchain, using the version
// which is set as part of the build process.
func DefaultBlockchainInfo() BlockchainInfo {
	var networkName string
	switch build.Release {
	case "dev":
		networkName = "devnet"
	case "testing":
		networkName = "testnet"
	default:
		networkName = "standard"
	}
	return BlockchainInfo{
		Name:            "Rivine",
		NetworkName:     networkName,
		CoinUnit:        "ROC",
		ChainVersion:    build.Version,
		ProtocolVersion: build.Version,
	}
}
