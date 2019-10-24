package explorerdb

import (
	"fmt"
	"io"

	"github.com/threefoldtech/rivine/crypto"
	"github.com/threefoldtech/rivine/pkg/encoding/rivbin"
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
		ID       types.OutputID
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
		ID types.TransactionID

		ParentBlock types.BlockID
		Version     types.TransactionVersion

		CoinInputs  []types.OutputID
		CoinOutputs []types.OutputID

		BlockStakeInputs  []types.OutputID
		BlockStakeOutputs []types.OutputID

		FeePayout TransactionFeePayoutInfo

		ArbitraryData        []byte
		EncodedExtensionData []byte
	}

	Block struct {
		ID        types.BlockID
		ParentID  types.BlockID
		Height    types.BlockHeight
		Timestamp types.Timestamp
		Payouts   []types.OutputID

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
		UnlockHash types.UnlockHash
		PublicKey  types.PublicKey `storm:"index"`

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
		UnlockHash types.UnlockHash

		ContractValue     types.Currency
		ContractCondition types.AtomicSwapCondition
		Transactions      []types.TransactionID
		CoinInput         types.CoinOutputID

		SpenditureData *AtomicSwapContractSpenditureData
	}
)

func (block *Block) AsObject() *Object {
	return &Object{
		ID:   ObjectID(block.ID[:]),
		Type: ObjectTypeBlock,
		Data: *block,
	}
}

func (txn *Transaction) AsObject() *Object {
	return &Object{
		ID:   ObjectID(txn.ID[:]),
		Type: ObjectTypeTransaction,
		Data: *txn,
	}
}

func (output *Output) AsObject() *Object {
	return &Object{
		ID:   ObjectID(output.ID[:]),
		Type: ObjectTypeOutput,
		Data: *output,
	}
}

func (wallet *WalletData) AsObject() *Object {
	return &Object{
		ID:   ObjectIDFromUnlockHash(wallet.UnlockHash),
		Type: ObjectTypeWallet,
		Data: *wallet,
	}
}

func (mswallet *MultiSignatureWalletData) AsObject() *Object {
	return &Object{
		ID:   ObjectIDFromUnlockHash(mswallet.UnlockHash),
		Type: ObjectTypeMultiSignatureWallet,
		Data: *mswallet,
	}
}

func (contract *AtomicSwapContract) AsObject() *Object {
	return &Object{
		ID:   ObjectIDFromUnlockHash(contract.UnlockHash),
		Type: ObjectTypeAtomicSwapContract,
		Data: *contract,
	}
}

func (rp ReferencePoint) IsTimestamp() bool {
	return rp >= types.LockTimeMinTimestampValue
}

func (rp ReferencePoint) IsBlockHeight() bool {
	return rp < types.LockTimeMinTimestampValue
}

type ObjectType uint8

const (
	ObjectTypeUndefined ObjectType = iota
	ObjectTypeBlock
	ObjectTypeTransaction
	ObjectTypeOutput
	ObjectTypeWallet
	ObjectTypeMultiSignatureWallet
	ObjectTypeAtomicSwapContract
)

type (
	ObjectID []byte

	Object struct {
		ID   ObjectID `storm:"id"`
		Type ObjectType
		Data interface{}
	}
)

func ObjectIDFromUnlockHash(uh types.UnlockHash) ObjectID {
	id := make(ObjectID, crypto.HashSize+1)
	id[0] = byte(uh.Type)
	copy(id[1:], uh.Hash[:])
	return id
}

func (obj Object) MarshalRivine(w io.Writer) error {
	enc := rivbin.NewEncoder(w)
	err := enc.EncodeAll(obj.ID, obj.Type)
	if err != nil {
		return err
	}
	if obj.Type == ObjectTypeUndefined || obj.Data == nil {
		return enc.Encode(false)
	}
	return enc.EncodeAll(true, obj.Data)
}

func (obj *Object) UnmarshalRivine(r io.Reader) error {
	dec := rivbin.NewDecoder(r)
	var hasData bool
	err := dec.DecodeAll(&obj.ID, &obj.Type, &hasData)
	if err != nil {
		return err
	}
	if !hasData {
		obj.Data = nil
		return nil // nothing more to do
	}
	switch obj.Type {
	case ObjectTypeBlock:
		var block Block
		err = dec.Decode(&block)
		obj.Data = block
	case ObjectTypeTransaction:
		var txn Transaction
		err = dec.Decode(&txn)
		obj.Data = txn
	case ObjectTypeOutput:
		var output Output
		err = dec.Decode(&output)
		obj.Data = output
	case ObjectTypeWallet:
		var wd WalletData
		err = dec.Decode(&wd)
		obj.Data = wd
	case ObjectTypeMultiSignatureWallet:
		var mswd MultiSignatureWalletData
		err = dec.Decode(&mswd)
		obj.Data = mswd
	case ObjectTypeAtomicSwapContract:
		var asc AtomicSwapContract
		err = dec.Decode(&asc)
		obj.Data = asc
	default:
		return fmt.Errorf("object type %d is unknown or is not expected to have data", obj.Type)
	}
	return err
}
