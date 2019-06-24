package types

import (
	"github.com/threefoldtech/rivine/build"
)

// BlockchainInfo contains information about a blockchain.
type BlockchainInfo struct {
	Name            string
	NetworkName     string
	CoinUnit        string
	ChainVersion    build.ProtocolVersion
	ProtocolVersion build.ProtocolVersion
}

// DefaultNetworkName returns a sane default network name,
// based on the build.Release tag (NOTE that in most cases
// you do really want a user-approved default network name
// rather than this static value).
func DefaultNetworkName() string {
	switch build.Release {
	case "testing":
		return "testnet"
	case "dev":
		return "devnet"
	default:
		if build.Release != "standard" {
			err := "unknown build.Release tag: " + build.Release
			build.Critical(err)
		}
		return "standard"
	}
}

// DefaultBlockchainInfo returns the blockchain information
// for the default (Rivine) blockchain, using the version
// which is set as part of the build process.
func DefaultBlockchainInfo() BlockchainInfo {
	return BlockchainInfo{
		Name:            "Rivine",
		NetworkName:     DefaultNetworkName(),
		CoinUnit:        "ROC",
		ChainVersion:    build.Version,
		ProtocolVersion: build.Version,
	}
}
