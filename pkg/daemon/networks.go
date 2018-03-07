package daemon

import (
	"fmt"

	"github.com/rivine/rivine/modules"
	"github.com/rivine/rivine/types"
)

var (
	// networks is the mapping of names to the parameters required to initialize them
	networks = make(map[string]Network)
)

// Network are variables for a particular chain. Currently, these are genesis constants and bootstrap peers
type Network struct {
	// Constants for this network
	Constants types.ChainConstants
	// BootstrapPeers for this network
	BootstrapPeers []modules.NetAddress
}

// RegisterNetwork registers a new network for a given name. A single name can only be used once
func RegisterNetwork(name string, network Network) error {
	if _, exists := networks[name]; exists {
		return fmt.Errorf("Network with name %s already exists", name)
	}
	networks[name] = network
	return nil
}

// setupNetwork takes a network and injects all its variables into the respective packages
func setupNetwork(name string) error {
	nw, exists := networks[name]
	if !exists {
		return fmt.Errorf("Network with name %s does not exist", name)
	}
	err := types.SetChainConfig(nw.Constants)
	if err != nil {
		return err
	}
	modules.BootstrapPeers = nw.BootstrapPeers
	return nil
}
