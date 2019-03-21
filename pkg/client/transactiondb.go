package client

import (
	"fmt"

	"github.com/threefoldtech/rivine/pkg/api"
	types "github.com/threefoldtech/rivine/types"
)

// TransactionDBClient is used to be able to get the active mint condition,
// the active mint condition at a given block height, as well as any 3bot information
// such that the CLI can also correctly validate a mint-type a 3bot-type transaction,
// without requiring access to the consensus-extended transactiondb,
// normally the validation isn't required on the client side, but it is now possible none the less
type TransactionDBClient struct {
	client       *CommandLineClient
	rootEndpoint string
}

// NewTransactionDBConsensusClient creates a new TransactionDBClient,
// that can be used for easy interaction with the TransactionDB API exposed via the Consensus endpoints
func NewTransactionDBConsensusClient(cli *CommandLineClient) *TransactionDBClient {
	if cli == nil {
		panic("no CommandLineClient given")
	}
	return &TransactionDBClient{
		client:       cli,
		rootEndpoint: "/consensus",
	}
}

// NewTransactionDBExplorerClient creates a new TransactionDBClient,
// that can be used for easy interaction with the TransactionDB API exposed via the Explorer endpoints
func NewTransactionDBExplorerClient(cli *CommandLineClient) *TransactionDBClient {
	if cli == nil {
		panic("no CommandLineClient given")
	}
	return &TransactionDBClient{
		client:       cli,
		rootEndpoint: "/explorer",
	}
}

var (
	// ensure TransactionDBClient implements the MintConditionGetter interface
	_ types.MintConditionGetter = (*TransactionDBClient)(nil)
)

// GetActiveMintCondition implements types.MintConditionGetter.GetActiveMintCondition
func (cli *TransactionDBClient) GetActiveMintCondition() (types.UnlockConditionProxy, error) {
	var result api.TransactionDBGetMintCondition
	err := cli.client.GetAPI(cli.rootEndpoint+"/mintcondition", &result)
	if err != nil {
		return types.UnlockConditionProxy{}, fmt.Errorf(
			"failed to get active mint condition from daemon: %v", err)
	}
	return result.MintCondition, nil
}

// GetMintConditionAt implements types.MintConditionGetter.GetMintConditionAt
func (cli *TransactionDBClient) GetMintConditionAt(height types.BlockHeight) (types.UnlockConditionProxy, error) {
	var result api.TransactionDBGetMintCondition
	err := cli.client.GetAPI(fmt.Sprintf("%s/mintcondition/%d", cli.rootEndpoint, height), &result)
	if err != nil {
		return types.UnlockConditionProxy{}, fmt.Errorf(
			"failed to get mint condition at height %d from daemon: %v", height, err)
	}
	return result.MintCondition, nil
}
