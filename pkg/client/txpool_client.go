package client

import (
	"encoding/json"

	rivineapi "github.com/threefoldtech/rivine/pkg/api"
	"github.com/threefoldtech/rivine/types"
)

// TransactionPoolClient is used to easily interact
// with the transaction pool through the HTTP REST API.
type TransactionPoolClient struct {
	bc *BaseClient
}

// NewTransactionPoolClient creates a new TransactionPoolClient,
// that can be used for easy interaction with the TransactionPool API exposed via the HTTP REST API.
func NewTransactionPoolClient(bc *BaseClient) *TransactionPoolClient {
	if bc == nil {
		panic("no BaseClient given")
	}
	return &TransactionPoolClient{
		bc: bc,
	}
}

// AddTransactiom adds the given transaction to the transaction pool, if valid.
func (tpool *TransactionPoolClient) AddTransactiom(t types.Transaction) (types.TransactionID, error) {
	b, err := json.Marshal(t)
	if err != nil {
		return types.TransactionID{}, err
	}
	var resp rivineapi.TransactionPoolPOST
	err = tpool.bc.HTTP().PostWithResponse("/transactionpool/transactions", string(b), &resp)
	if err != nil {
		return types.TransactionID{}, err
	}
	return resp.TransactionID, nil
}
