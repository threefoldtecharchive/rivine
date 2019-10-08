package types

import (
	"github.com/threefoldtech/rivine/types"
)

const (
	// TransactionVersionMinterDefinition defines the Transaction version
	// for a MinterDefinition Transaction.
	//
	// See the `MinterDefinitionTransactionController` and `MinterDefinitionTransaction`
	// types for more information.
	TransactionVersionMinterDefinition types.TransactionVersion = iota + 128
	// TransactionVersionCoinCreation defines the Transaction version
	// for a CoinCreation Transaction.
	//
	// See the `CoinCreationTransactionController` and `CoinCreationTransaction`
	// types for more information.
	TransactionVersionCoinCreation
	// TransactionVersionCoinDestruction defines the Transaction version
	// for a CoinDestruction Transaction.
	//
	// See the `CoinDestructionTransactionController` and `CoinDestructionTransaction`
	// types for more information.
	TransactionVersionCoinDestruction
)

const (
	// TransactionVersionAuthAddressUpdate defines the Transaction version
	// for a AuthAddressUpdate Transaction.
	//
	// See the `AuthAddressUpdateTransactionController` and `AuthAddressUpdateTransaction`
	// types for more information.
	TransactionVersionAuthAddressUpdate types.TransactionVersion = iota + 176
	// TransactionVersionAuthConditionUpdate defines the Transaction version
	// for a AuthConditionUpdat Transaction.
	//
	// See the `AuthConditionUpdateTransactionController` and `AuthConditionUpdateTransaction`
	// types for more information.
	TransactionVersionAuthConditionUpdate
)
