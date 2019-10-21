package explorerdb

import (
	"github.com/threefoldtech/rivine/crypto"
	"github.com/threefoldtech/rivine/types"
)

type OutputType uint8

const (
	OutputTypeUndefined OutputType = iota
	OutputTypeCoin
	OutputTypeBlockStake
	OutputTypeBlockCreationReward
	OutputTypeTransactionFee
)

type (
	Output struct {
		ID       types.OutputID `storm:"id"`
		ParentID crypto.Hash

		Type      OutputType
		Value     types.Currency
		Condition types.UnlockConditionProxy

		SpenditureData *OutputSpenditureData
	}

	OutputSpenditureData struct {
		Fulfillment              types.UnlockFulfillmentProxy
		FulfillmentTransactionID types.TransactionID
	}

	TransactionFeePayoutInfo struct {
		PayoutID types.OutputID
		Values   []types.Currency
	}

	Transaction struct {
		ID types.TransactionID `storm:"id"`

		ParentBlock types.BlockID
		Version     types.TransactionVersion

		Inputs    []types.OutputID
		Outputs   []types.OutputID
		FeePayout TransactionFeePayoutInfo

		ExtensionData interface{}
	}

	Block struct {
		ID      types.BlockID `storm:"id"`
		Payouts []types.OutputID

		Transactions []types.TransactionID
	}

	ReferencePoint uint64

	LockedOutputData struct {
		ReferencePoint ReferencePoint
		OutputIndex    int
	}

	Balance struct {
		Unlocked              types.Currency
		Locked                types.Currency
		LockedOutputDataSlice []LockedOutputData

		LastUpdateTimestamp   types.Timestamp
		LastUpdateBlockHeight types.BlockHeight
		LastUpdateTransaction types.TransactionID
	}

	WalletData struct {
		UnlockHash types.UnlockHash `storm:"id"`
		PublicKey  types.PublicKey  `storm:"index"`

		CoinOutputs       []types.OutputID
		BlockStakeOutputs types.OutputID
		Transactions      []types.TransactionID

		CoinBalance       Balance
		BlockStakeBalance Balance
	}

	MultiSignatureWalletOwner struct {
		UnlockHash types.UnlockHash
		PublicKey  types.PublicKey
	}

	MultiSignatureWalletData struct {
		WalletData `storm:"inline"`

		Owners                []MultiSignatureWalletOwner
		RequiredSgnatureCount int
	}

	AtomicSwapContractSpenditureData struct {
		ContractFulfillment types.AtomicSwapFulfillment
		CoinOutput          types.CoinOutputID
	}

	AtomicSwapContract struct {
		UnlockHash types.UnlockHash `storm:"id"`

		ContractValue     types.Currency
		ContractCondition types.AtomicSwapCondition
		Transactions      []types.TransactionID
		CoinInput         types.CoinOutputID

		SpenditureData *AtomicSwapContractSpenditureData
	}
)
