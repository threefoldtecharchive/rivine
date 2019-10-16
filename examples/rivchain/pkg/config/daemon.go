package config

import (
	"github.com/threefoldtech/rivine/types"
)

// DaemonNetworkConfig defines network-specific constants.
type DaemonNetworkConfig struct {
	FoundationPoolAddress types.UnlockHash
}

func GetDevnetDaemonNetworkConfig() DaemonNetworkConfig {
	return DaemonNetworkConfig{
		FoundationPoolAddress: unlockHashFromHex("015a080a9259b9d4aaa550e2156f49b1a79a64c7ea463d810d4493e8242e6791584fbdac553e6f"),
	}
}

func GetStandardDaemonNetworkConfig() DaemonNetworkConfig {
	return DaemonNetworkConfig{
		FoundationPoolAddress: unlockHashFromHex("017267221ef1947bb18506e390f1f9446b995acfb6d08d8e39508bb974d9830b8cb8fdca788e34"),
	}
}

func GetTestnetDaemonNetworkConfig() DaemonNetworkConfig {
	return DaemonNetworkConfig{
		FoundationPoolAddress: unlockHashFromHex("01434535fd01243c02c277cd58d71423163767a575a8ae44e15807bf545e4a8456a5c4afabad51"),
	}
}
