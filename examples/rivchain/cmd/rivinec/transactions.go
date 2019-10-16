package main

import (
	rivchaintypes "github.com/threefoldtech/rivine/examples/rivchain/pkg/types"
	"github.com/threefoldtech/rivine/extensions/authcointx"
	authcointxcli "github.com/threefoldtech/rivine/extensions/authcointx/client"
	"github.com/threefoldtech/rivine/extensions/minting"
	mintingcli "github.com/threefoldtech/rivine/extensions/minting/client"
	"github.com/threefoldtech/rivine/types"

	"github.com/threefoldtech/rivine/pkg/client"
)

func RegisterDevnetTransactions(bc client.BaseClient) {
	registerTransactions(bc)
}

func RegisterStandardTransactions(bc client.BaseClient) {
	registerTransactions(bc)
}

func RegisterTestnetTransactions(bc client.BaseClient) {
	registerTransactions(bc)
}

func registerTransactions(bc client.BaseClient) {
	// create minting plugin client...
	mintingCLI := mintingcli.NewPluginConsensusClient(bc)
	// ...and register minting types
	types.RegisterTransactionVersion(rivchaintypes.TransactionVersionMinterDefinition, minting.MinterDefinitionTransactionController{
		MintConditionGetter: mintingCLI,
		TransactionVersion:  rivchaintypes.TransactionVersionMinterDefinition,
	})
	types.RegisterTransactionVersion(rivchaintypes.TransactionVersionCoinCreation, minting.CoinCreationTransactionController{
		MintConditionGetter: mintingCLI,
		TransactionVersion:  rivchaintypes.TransactionVersionCoinCreation,
	})
	types.RegisterTransactionVersion(rivchaintypes.TransactionVersionCoinDestruction, minting.CoinDestructionTransactionController{
		TransactionVersion: rivchaintypes.TransactionVersionCoinDestruction,
	})

	// create coin auth tx plugin client...
	authCoinTxCLI := authcointxcli.NewPluginConsensusClient(bc)
	// ...and register coin auth tx types
	types.RegisterTransactionVersion(rivchaintypes.TransactionVersionAuthConditionUpdate, authcointx.AuthConditionUpdateTransactionController{
		AuthInfoGetter:     authCoinTxCLI,
		TransactionVersion: rivchaintypes.TransactionVersionAuthConditionUpdate,
	})
	types.RegisterTransactionVersion(rivchaintypes.TransactionVersionAuthAddressUpdate, authcointx.AuthAddressUpdateTransactionController{
		AuthInfoGetter:     authCoinTxCLI,
		TransactionVersion: rivchaintypes.TransactionVersionAuthAddressUpdate,
	})
}
