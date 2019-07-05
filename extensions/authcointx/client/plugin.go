package client

import (
	"encoding/json"
	"fmt"

	"github.com/threefoldtech/rivine/extensions/authcointx"
	"github.com/threefoldtech/rivine/extensions/authcointx/api"
	"github.com/threefoldtech/rivine/pkg/client"
	"github.com/threefoldtech/rivine/types"
)

// PluginClient is used to be able to get auth information from
// a daemon that has the authcointx extension enabled and running.
type PluginClient struct {
	client       *client.CommandLineClient
	rootEndpoint string
}

// NewPluginConsensusClient creates a new PluginClient,
// that can be used for easy interaction with the API exposed via the Consensus endpoints
func NewPluginConsensusClient(cli *client.CommandLineClient) *PluginClient {
	if cli == nil {
		panic("no CommandLineClient given")
	}
	return &PluginClient{
		client:       cli,
		rootEndpoint: "/consensus",
	}
}

// NewPluginExplorerClient creates a new PluginClient,
// that can be used for easy interaction with the API exposed via the Explorer endpoints
func NewPluginExplorerClient(cli *client.CommandLineClient) *PluginClient {
	if cli == nil {
		panic("no CommandLineClient given")
	}
	return &PluginClient{
		client:       cli,
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
	err := cli.client.GetAPI(cli.rootEndpoint+"/authcoin/condition", &result)
	if err != nil {
		return types.UnlockConditionProxy{}, fmt.Errorf(
			"failed to get active auth condition from daemon: %v", err)
	}
	return result.AuthCondition, nil
}

// GetAuthConditionAt implements authcointx.AuthInfoGetter.GetAuthConditionAt
func (cli *PluginClient) GetAuthConditionAt(height types.BlockHeight) (types.UnlockConditionProxy, error) {
	var result api.GetAuthConditionResponse
	err := cli.client.GetAPI(fmt.Sprintf("%s/authcoin/condition/%d", cli.rootEndpoint, height), &result)
	if err != nil {
		return types.UnlockConditionProxy{}, fmt.Errorf(
			"failed to get auth condition at height %d from daemon: %v", height, err)
	}
	return result.AuthCondition, nil
}

// GetAddresAuthStateNow provides functionality now required for the AuthInfoGetter,
// allowing you to request it for a single address
func (cli *PluginClient) GetAddresAuthStateNow(address types.UnlockHash) (bool, error) {
	var result api.GetAddressAuthStateResponse
	err := cli.client.GetAPI(fmt.Sprintf("%s/authcoin/address/%s", cli.rootEndpoint, address.String()), &result)
	return result.AuthState, err
}

// GetAddressAuthStateAt provides functionality now required for the AuthInfoGetter,
// allowing you to request it for a single address
func (cli *PluginClient) GetAddressAuthStateAt(height types.BlockHeight, address types.UnlockHash) (bool, error) {
	var result api.GetAddressAuthStateResponse
	err := cli.client.GetAPI(fmt.Sprintf("%s/authcoin/address/%s/%d", cli.rootEndpoint, address.String(), height), &result)
	return result.AuthState, err
}

// GetAddressesAuthStateNow implements authcointx.AuthInfoGetter.GetAddressesAuthStateNow
func (cli *PluginClient) GetAddressesAuthStateNow(addresses []types.UnlockHash, _ func(index int, state bool) bool) ([]bool, error) {
	requestData, err := json.Marshal(api.GetAddressesAuthState{
		Addresses: addresses,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal addresses as input addresses for the request: %v", err)
	}
	var result api.GetAddressesAuthStateResponse
	err = cli.client.GetAPIWithData(fmt.Sprintf("%s/authcoin/addresses", cli.rootEndpoint), string(requestData), &result)
	return result.AuthStates, err
}

// GetAddressesAuthStateAt implements authcointx.AuthInfoGetter.GetAddressesAuthStateAt
func (cli *PluginClient) GetAddressesAuthStateAt(height types.BlockHeight, addresses []types.UnlockHash, _ func(index int, state bool) bool) ([]bool, error) {
	requestData, err := json.Marshal(api.GetAddressesAuthState{
		Addresses: addresses,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal addresses as input addresses for the request: %v", err)
	}
	var result api.GetAddressesAuthStateResponse
	err = cli.client.GetAPIWithData(fmt.Sprintf("%s/authcoin/addresses/%d", cli.rootEndpoint, height), string(requestData), &result)
	return result.AuthStates, err
}
