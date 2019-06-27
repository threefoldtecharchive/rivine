package client

import (
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
	var result api.GetAuthCondition
	err := cli.client.GetAPI(cli.rootEndpoint+"/authcoin/condition", &result)
	if err != nil {
		return types.UnlockConditionProxy{}, fmt.Errorf(
			"failed to get active auth condition from daemon: %v", err)
	}
	return result.AuthCondition, nil
}

// GetAuthConditionAt implements authcointx.AuthInfoGetter.GetAuthConditionAt
func (cli *PluginClient) GetAuthConditionAt(height types.BlockHeight) (types.UnlockConditionProxy, error) {
	var result api.GetAuthCondition
	err := cli.client.GetAPI(fmt.Sprintf("%s/authcoin/condition/%d", cli.rootEndpoint, height), &result)
	if err != nil {
		return types.UnlockConditionProxy{}, fmt.Errorf(
			"failed to get auth condition at height %d from daemon: %v", height, err)
	}
	return result.AuthCondition, nil
}

// EnsureAddressesAreAuthNow implements authcointx.AuthInfoGetter.EnsureAddressesAreAuthNow
func (cli *PluginClient) EnsureAddressesAreAuthNow(addresses ...types.UnlockHash) error {
	var (
		err    error
		result api.GetAddressAuthState
	)
	for _, address := range addresses {
		err = cli.client.GetAPI(fmt.Sprintf("%s/authcoin/authcoin/address/%s", cli.rootEndpoint, address.String()), &result)
		if err != nil {
			err = fmt.Errorf(
				"failed to get address %s auth state from daemon: %v",
				address.String(), err)
			return err
		}
		if result.Address.Cmp(address) != 0 {
			err = fmt.Errorf(
				"failed to get address %s auth state from daemon: invalid address returned: %s",
				address.String(), result.Address.String())
			return err
		}
		if !result.AuthState {
			err = fmt.Errorf("address %s is not authorized", address.String())
			return err
		}
	}
	return nil
}

// EnsureAddressesAreAuthAt implements authcointx.AuthInfoGetter.EnsureAddressesAreAuthAt
func (cli *PluginClient) EnsureAddressesAreAuthAt(height types.BlockHeight, addresses ...types.UnlockHash) error {
	var (
		err    error
		result api.GetAddressAuthState
	)
	for _, address := range addresses {
		err = cli.client.GetAPI(fmt.Sprintf("%s/authcoin/authcoin/address/%s/%d", cli.rootEndpoint, address.String(), height), &result)
		if err != nil {
			err = fmt.Errorf(
				"failed to get address %s auth state at height %d from daemon: %v",
				address.String(), height, err)
			return err
		}
		if result.Address.Cmp(address) != 0 {
			err = fmt.Errorf(
				"failed to get address %s auth state at height %d from daemon: invalid address returned: %s",
				address.String(), height, result.Address.String())
			return err
		}
		if !result.AuthState {
			err = fmt.Errorf("address %s is not authorized at height %d", address.String(), height)
			return err
		}
	}
	return nil
}
