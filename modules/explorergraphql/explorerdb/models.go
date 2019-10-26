package explorerdb

import (
	"fmt"
	"io"

	"github.com/threefoldtech/rivine/crypto"
	"github.com/threefoldtech/rivine/modules"
	"github.com/threefoldtech/rivine/pkg/encoding/rivbin"
	"github.com/threefoldtech/rivine/types"
)

const (
	// ActiveBSEstimationBlocks is the number of blocks that are used to
	// estimate the active block stake used to generate blocks.
	ActiveBSEstimationBlocks = 500
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
	ChainContext struct {
		ConsensusChangeID modules.ConsensusChangeID
		Height            types.BlockHeight
		Timestamp         types.Timestamp
		BlockID           types.BlockID
	}

	ChainAggregatedFacts struct {
		TotalCoins       types.Currency
		TotalLockedCoins types.Currency

		TotalBlockStakes           types.Currency
		TotalLockedBlockStakes     types.Currency
		EstimatedActiveBlockStakes types.Currency

		LastBlocks []BlockFactsContext
	}

	BlockFactsContext struct {
		Target    types.Target
		Timestamp types.Timestamp
	}

	BlockRevertContext struct {
		ID        types.BlockID
		Height    types.BlockHeight
		Timestamp types.Timestamp
	}

	Output struct {
		ID       types.OutputID
		ParentID crypto.Hash

		Type                 OutputType
		Value                types.Currency
		Condition            types.UnlockConditionProxy
		UnlockReferencePoint ReferencePoint

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

	BlockFacts struct {
		Constants  BlockFactsConstants
		Aggregated BlockFactsAggregated
	}

	BlockFactsConstants struct {
		Difficulty types.Difficulty
		Target     types.Target
	}

	BlockFactsAggregated struct {
		TotalCoins                 types.Currency
		TotalLockedCoins           types.Currency
		TotalBlockStakes           types.Currency
		TotalLockedBlockStakes     types.Currency
		EstimatedActiveBlockStakes types.Currency
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

		CoinOutputs       []types.OutputID
		BlockStakeOutputs types.OutputID
		Transactions      []types.TransactionID

		CoinBalance       Balance
		BlockStakeBalance Balance
	}

	SingleSignatureWalletData struct {
		WalletData

		PublicKey types.PublicKey
	}

	MultiSignatureWalletOwner struct {
		UnlockHash types.UnlockHash
		PublicKey  types.PublicKey
	}

	MultiSignatureWalletData struct {
		WalletData

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

func (facts *ChainAggregatedFacts) AddLastBlockContext(ctx BlockFactsContext) {
	// assemble block fact ctx
	facts.LastBlocks = append(facts.LastBlocks, ctx)
	// trim if needed
	blockLength := len(facts.LastBlocks)
	if blockLength > ActiveBSEstimationBlocks {
		facts.LastBlocks = facts.LastBlocks[blockLength-ActiveBSEstimationBlocks:]
	}
}

func (facts *ChainAggregatedFacts) RemoveLastBlockContext() {
	blockLength := len(facts.LastBlocks)
	if blockLength == 0 {
		return // nothing to do
	}
	facts.LastBlocks = facts.LastBlocks[:blockLength-1]
}

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
		Type: ObjectTypeSingleSignatureWallet,
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

func (rp ReferencePoint) Reached(height types.BlockHeight, timestamp types.Timestamp) bool {
	if rp < types.LockTimeMinTimestampValue {
		return types.BlockHeight(rp) <= height
	}
	return types.Timestamp(rp) <= timestamp
}

func (rp ReferencePoint) Overreached(height types.BlockHeight, timestamp types.Timestamp) bool {
	if rp < types.LockTimeMinTimestampValue {
		return types.BlockHeight(rp) < height
	}
	return types.Timestamp(rp) < timestamp
}

func UnlockReferencePointFromCondition(condition types.UnlockConditionProxy) (ReferencePoint, bool) {
	if condition.ConditionType() != types.ConditionTypeTimeLock {
		return 0, false
	}
	tlc, ok := condition.Condition.(*types.TimeLockCondition)
	if !ok {
		return 0, false
	}
	return ReferencePoint(tlc.LockTime), true
}

type ObjectType uint8

const (
	ObjectTypeUndefined ObjectType = iota
	ObjectTypeBlock
	ObjectTypeTransaction
	ObjectTypeOutput
	ObjectTypeSingleSignatureWallet
	ObjectTypeMultiSignatureWallet
	ObjectTypeAtomicSwapContract
)

type (
	ObjectID []byte

	Object struct {
		ID   ObjectID
		Type ObjectType
		Data interface{}
	}
)

const (
	BinaryHashDataSize       = crypto.HashSize
	BinaryUnlockHashDataSize = crypto.HashSize + 1
)

func ObjectIDFromUnlockHash(uh types.UnlockHash) ObjectID {
	id := make(ObjectID, BinaryUnlockHashDataSize)
	id[0] = byte(uh.Type)
	copy(id[1:], uh.Hash[:])
	return id
}

func (objectID ObjectID) AsUnlockHash() (types.UnlockHash, error) {
	if length := len(objectID); length != BinaryUnlockHashDataSize {
		return types.UnlockHash{}, fmt.Errorf("objectID %x cannot be casted as UnlockHash: expected byte size %d, but the size is %d", objectID, length, BinaryUnlockHashDataSize)
	}
	uh := types.UnlockHash{
		Type: types.UnlockType(objectID[0]),
	}
	copy(uh.Hash[:], objectID[1:])
	return uh, nil
}

func (objectID ObjectID) AsHash() (crypto.Hash, error) {
	if length := len(objectID); length != BinaryHashDataSize {
		return crypto.Hash{}, fmt.Errorf("objectID %x cannot be casted as Hash: expected byte size %d, but the size is %d", objectID, length, BinaryHashDataSize)
	}
	var hash crypto.Hash
	copy(hash[:], objectID[:])
	return hash, nil
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
	case ObjectTypeSingleSignatureWallet:
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

func (obj *Object) AsBlock() (Block, error) {
	if obj.Type != ObjectTypeBlock {
		return Block{}, fmt.Errorf("object of type %d cannot be casted to block (type %d)", obj.Type, ObjectTypeBlock)
	}
	block, ok := obj.Data.(Block)
	if !ok {
		return Block{}, fmt.Errorf("object of type %d has block type but has wrong underlying data type of %T (expected: %T)", obj.Type, obj.Data, Block{})
	}
	return block, nil
}

func (obj *Object) AsTransaction() (Transaction, error) {
	if obj.Type != ObjectTypeTransaction {
		return Transaction{}, fmt.Errorf("object of type %d cannot be casted to transaction (type %d)", obj.Type, ObjectTypeTransaction)
	}
	txn, ok := obj.Data.(Transaction)
	if !ok {
		return Transaction{}, fmt.Errorf("object of type %d has transaction type but has wrong underlying data type of %T (expected: %T)", obj.Type, obj.Data, Transaction{})
	}
	return txn, nil
}

func (obj *Object) AsOutput() (Output, error) {
	if obj.Type != ObjectTypeOutput {
		return Output{}, fmt.Errorf("object of type %d cannot be casted to output (type %d)", obj.Type, ObjectTypeOutput)
	}
	output, ok := obj.Data.(Output)
	if !ok {
		return Output{}, fmt.Errorf("object of type %d has output type but has wrong underlying data type of %T (expected: %T)", obj.Type, obj.Data, Output{})
	}
	return output, nil
}

func (obj *Object) AsSingleSignatureWallet() (WalletData, error) {
	if obj.Type != ObjectTypeSingleSignatureWallet {
		return WalletData{}, fmt.Errorf("object of type %d cannot be casted to wallet (type %d)", obj.Type, ObjectTypeSingleSignatureWallet)
	}
	wd, ok := obj.Data.(WalletData)
	if !ok {
		return WalletData{}, fmt.Errorf("object of type %d has wallet type but has wrong underlying data type of %T (expected: %T)", obj.Type, obj.Data, WalletData{})
	}
	return wd, nil
}

func (obj *Object) AsMultiSignatureWallet() (MultiSignatureWalletData, error) {
	if obj.Type != ObjectTypeMultiSignatureWallet {
		return MultiSignatureWalletData{}, fmt.Errorf("object of type %d cannot be casted to multi signature wallet (type %d)", obj.Type, ObjectTypeMultiSignatureWallet)
	}
	mswd, ok := obj.Data.(MultiSignatureWalletData)
	if !ok {
		return MultiSignatureWalletData{}, fmt.Errorf("object of type %d has multi signature wallet type but has wrong underlying data type of %T (expected: %T)", obj.Type, obj.Data, MultiSignatureWalletData{})
	}
	return mswd, nil
}

func (obj *Object) AsAtomicSwapContract() (AtomicSwapContract, error) {
	if obj.Type != ObjectTypeAtomicSwapContract {
		return AtomicSwapContract{}, fmt.Errorf("object of type %d cannot be casted to atomic swap contract (type %d)", obj.Type, ObjectTypeAtomicSwapContract)
	}
	asc, ok := obj.Data.(AtomicSwapContract)
	if !ok {
		return AtomicSwapContract{}, fmt.Errorf("object of type %d has atomic swap contract type but has wrong underlying data type of %T (expected: %T)", obj.Type, obj.Data, AtomicSwapContract{})
	}
	return asc, nil
}
