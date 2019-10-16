package client

import (
	"fmt"

	minting "github.com/threefoldtech/rivine/extensions/minting"
	"github.com/threefoldtech/rivine/extensions/minting/api"
	client "github.com/threefoldtech/rivine/pkg/client"
	types "github.com/threefoldtech/rivine/types"
)

// PluginClient is used to be able to get the active mint condition,
// and the active mint condition at a given block height,
// such that the CLI can correctly validate a mint condition,
// without requiring access to the consensus-extended transactiondb,
// normally the validation isn't required on the client side, but it is possible none the less.
type PluginClient struct {
	client       client.BaseClient
	rootEndpoint string
}

// NewPluginConsensusClient creates a new PluginClient,
// that can be used for easy interaction with the TransactionDB API exposed via the Consensus endpoints
func NewPluginConsensusClient(cli client.BaseClient) *PluginClient {
	if cli == nil {
		panic("no BaseClient given")
	}
	return &PluginClient{
		client:       cli,
		rootEndpoint: "/consensus",
	}
}

// NewPluginExplorerClient creates a new PluginClient,
// that can be used for easy interaction with the TransactionDB API exposed via the Explorer endpoints
func NewPluginExplorerClient(cli client.BaseClient) *PluginClient {
	if cli == nil {
		panic("no BaseClient given")
	}
	return &PluginClient{
		client:       cli,
		rootEndpoint: "/explorer",
	}
}

var (
	// ensure PluginClient implements the MintConditionGetter interface
	_ minting.MintConditionGetter = (*PluginClient)(nil)
)

// GetActiveMintCondition implements minting.MintConditionGetter.GetActiveMintCondition
func (cli *PluginClient) GetActiveMintCondition() (types.UnlockConditionProxy, error) {
	var result api.TransactionDBGetMintCondition
	err := cli.client.HTTP().GetWithResponse(cli.rootEndpoint+"/mintcondition", &result)
	if err != nil {
		return types.UnlockConditionProxy{}, fmt.Errorf(
			"failed to get active mint condition from daemon: %v", err)
	}
	return result.MintCondition, nil
}

// GetMintConditionAt implements minting.MintConditionGetter.GetMintConditionAt
func (cli *PluginClient) GetMintConditionAt(height types.BlockHeight) (types.UnlockConditionProxy, error) {
	var result api.TransactionDBGetMintCondition
	err := cli.client.HTTP().GetWithResponse(fmt.Sprintf("%s/mintcondition/%d", cli.rootEndpoint, height), &result)
	if err != nil {
		return types.UnlockConditionProxy{}, fmt.Errorf(
			"failed to get mint condition at height %d from daemon: %v", height, err)
	}
	return result.MintCondition, nil
}
