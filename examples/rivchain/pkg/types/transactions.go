package types

import "github.com/threefoldtech/rivine/types"

const (
	//TransactionVersionMinterDefinition is the transaction version for the   minterdefinition transaction
	TransactionVersionMinterDefinition types.TransactionVersion = 128
	//TransactionVersionCoinCreation is the transaction version for the coin creation transaction
	TransactionVersionCoinCreation types.TransactionVersion = 129
	//TransactionVersionCoinDestruction is the transaction version for the coin destruction transaction
	TransactionVersionCoinDestruction types.TransactionVersion = 130
)

// Auth Coin Tx Extension Transaction Versions
const (
	TransactionVersionAuthAddressUpdate   types.TransactionVersion = 176
	TransactionVersionAuthConditionUpdate types.TransactionVersion = 177
)
