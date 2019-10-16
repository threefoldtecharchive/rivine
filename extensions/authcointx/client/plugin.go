package client

import (
	"errors"
	"fmt"

	"github.com/threefoldtech/rivine/extensions/authcointx"
	"github.com/threefoldtech/rivine/extensions/authcointx/api"
	"github.com/threefoldtech/rivine/pkg/client"
	"github.com/threefoldtech/rivine/types"
)

// PluginClient is used to be able to get auth information from
// a daemon that has the authcointx extension enabled and running.
type PluginClient struct {
	bc           client.BaseClient
	rootEndpoint string
}

// NewPluginConsensusClient creates a new PluginClient,
// that can be used for easy interaction with the API exposed via the Consensus endpoints
func NewPluginConsensusClient(bc client.BaseClient) *PluginClient {
	if bc == nil {
		panic("no BaseClient given")
	}
	return &PluginClient{
		bc:           bc,
		rootEndpoint: "/consensus",
	}
}

// NewPluginExplorerClient creates a new PluginClient,
// that can be used for easy interaction with the API exposed via the Explorer endpoints
func NewPluginExplorerClient(bc client.BaseClient) *PluginClient {
	if bc == nil {
		panic("no BaseClient given")
	}
	return &PluginClient{
		bc:           bc,
		rootEndpoint: "/explorer",
	}
}

var (
	// ensure PluginClient implements the AuthInfoGetter interface
	_ authcointx.AuthInfoGetter = (*PluginClient)(nil)
)

// GetActiveAuthCondition implements authcointx.AuthInfoGetter.GetActiveAuthCondition
func (cli *PluginClient) GetActiveAuthCondition() (types.UnlockConditionProxy, error) {
	var result api.GetAuthConditionResponse
	err := cli.bc.HTTP().GetWithResponse(cli.rootEndpoint+"/authcoin/condition", &result)
	if err != nil {
		return types.UnlockConditionProxy{}, fmt.Errorf(
			"failed to get active auth condition from daemon: %v", err)
	}
	return result.AuthCondition, nil
}

// GetAuthConditionAt implements authcointx.AuthInfoGetter.GetAuthConditionAt
func (cli *PluginClient) GetAuthConditionAt(height types.BlockHeight) (types.UnlockConditionProxy, error) {
	var result api.GetAuthConditionResponse
	err := cli.bc.HTTP().GetWithResponse(fmt.Sprintf("%s/authcoin/condition/%d", cli.rootEndpoint, height), &result)
	if err != nil {
		return types.UnlockConditionProxy{}, fmt.Errorf(
			"failed to get auth condition at height %d from daemon: %v", height, err)
	}
	return result.AuthCondition, nil
}

// GetAddresAuthStateNow provides functionality now required for the AuthInfoGetter,
// allowing you to request it for a single address
func (cli *PluginClient) GetAddresAuthStateNow(address types.UnlockHash) (bool, error) {
	states, err := cli.GetAddressesAuthStateNow([]types.UnlockHash{address}, nil)
	if err != nil {
		return false, err
	}
	return states[0], nil
}

// GetAddressAuthStateAt provides functionality now required for the AuthInfoGetter,
// allowing you to request it for a single address
func (cli *PluginClient) GetAddressAuthStateAt(height types.BlockHeight, address types.UnlockHash) (bool, error) {
	states, err := cli.GetAddressesAuthStateAt(height, []types.UnlockHash{address}, nil)
	if err != nil {
		return false, err
	}
	return states[0], nil
}

// GetAddressesAuthStateNow implements authcointx.AuthInfoGetter.GetAddressesAuthStateNow
func (cli *PluginClient) GetAddressesAuthStateNow(addresses []types.UnlockHash, _ func(index int, state bool) bool) ([]bool, error) {
	if len(addresses) == 0 {
		return nil, errors.New("no addresses are defined while this is required")
	}

	resource := fmt.Sprintf("%s/authcoin/status?", cli.rootEndpoint)
	for _, addr := range addresses {
		resource += fmt.Sprintf("addr=%s&", addr.String())
	}
	resource = resource[:len(resource)-1] // remove trailing '&'

	var result api.GetAddressesAuthStateResponse
	err := cli.bc.HTTP().GetWithResponse(resource, &result)
	return result.AuthStates, err
}

// GetAddressesAuthStateAt implements authcointx.AuthInfoGetter.GetAddressesAuthStateAt
func (cli *PluginClient) GetAddressesAuthStateAt(height types.BlockHeight, addresses []types.UnlockHash, _ func(index int, state bool) bool) ([]bool, error) {
	if len(addresses) == 0 {
		return nil, errors.New("no addresses are defined while this is required")
	}

	resource := fmt.Sprintf("%s/authcoin/status?height=%d&", cli.rootEndpoint, height)
	for _, addr := range addresses {
		resource += fmt.Sprintf("addr=%s&", addr.String())
	}
	resource = resource[:len(resource)-1] // remove trailing '&'

	var result api.GetAddressesAuthStateResponse
	err := cli.bc.HTTP().GetWithResponse(resource, &result)
	return result.AuthStates, err
}
