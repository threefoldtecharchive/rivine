package explorerdb

import (
	"fmt"

	"github.com/threefoldtech/rivine/crypto"
	"github.com/threefoldtech/rivine/extensions/minting"
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

		ArbitraryData []byte

		EncodedExtensionData []byte
	}

	TransactionCommonExtensionData struct {
		Fulfillments []types.UnlockFulfillmentProxy
		Conditions   []types.UnlockConditionProxy
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
		Unlocked types.Currency
		Locked   types.Currency
	}

	WalletData struct {
		UnlockHash types.UnlockHash

		// CoinOutputs       []types.OutputID
		// BlockStakeOutputs []types.OutputID

		// Blocks       []types.BlockID
		// Transactions []types.TransactionID

		CoinBalance       Balance
		BlockStakeBalance Balance
	}

	FreeForAllWalletData struct {
		WalletData
	}

	SingleSignatureWalletData struct {
		WalletData

		MultiSignatureWallets []types.UnlockHash
	}

	MultiSignatureWalletData struct {
		WalletData

		Owners                []types.UnlockHash
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

func (txn *Transaction) GetCommonExtensionData() (TransactionCommonExtensionData, error) {
	const ( // TODO: DO NOT HARDCODE THESE VALUES, SHOULD BE PLUGGED IN BY EXTENSIONS SOMEHOW
		txnVersionMinterDefinition types.TransactionVersion = 128
		txnVersionCoinCreation     types.TransactionVersion = 129
	)
	switch txn.Version {
	case txnVersionMinterDefinition:
		var data minting.MinterDefinitionTransactionExtension
		err := rivbin.Unmarshal(txn.EncodedExtensionData, &data)
		if err != nil {
			return TransactionCommonExtensionData{}, fmt.Errorf("failed to unmarshal minter definition ext. data: %v", err)
		}
		return TransactionCommonExtensionData{
			Fulfillments: []types.UnlockFulfillmentProxy{data.MintFulfillment},
			Conditions:   []types.UnlockConditionProxy{data.MintCondition},
		}, nil
	case txnVersionCoinCreation:
		var data minting.CoinCreationTransactionExtension
		err := rivbin.Unmarshal(txn.EncodedExtensionData, &data)
		if err != nil {
			return TransactionCommonExtensionData{}, fmt.Errorf("failed to unmarshal minter definition ext. data: %v", err)
		}
		return TransactionCommonExtensionData{
			Fulfillments: []types.UnlockFulfillmentProxy{data.MintFulfillment},
		}, nil
	default:
		// no extension data to return
		return TransactionCommonExtensionData{}, nil
	}
}

func (output *Output) AsObject() *Object {
	return &Object{
		ID:   ObjectID(output.ID[:]),
		Type: ObjectTypeOutput,
		Data: *output,
	}
}

func (balance *Balance) ApplyInput(input *Output) error {
	if input.SpenditureData == nil {
		return fmt.Errorf("cannot apply output %s as input to balance: nil spenditure data", input.ID.String())
	}
	balance.Unlocked = balance.Unlocked.Sub(input.Value)
	return nil
}

func (balance *Balance) RevertInput(input *Output) error {
	if input.SpenditureData == nil {
		return fmt.Errorf("cannot revert output %s as input to balance: nil spenditure data", input.ID.String())
	}
	balance.Unlocked = balance.Unlocked.Add(input.Value)
	return nil
}

func (balance *Balance) ApplyOutput(height types.BlockHeight, timestamp types.Timestamp, output *Output) error {
	if output.SpenditureData != nil {
		return fmt.Errorf("cannot apply output %s to balance: non-nil spenditure data", output.ID.String())
	}
	if output.UnlockReferencePoint.Reached(height, timestamp) {
		balance.Unlocked = balance.Unlocked.Add(output.Value)
	} else {
		balance.Locked = balance.Locked.Add(output.Value)
	}
	return nil
}

func (balance *Balance) RevertOutput(height types.BlockHeight, timestamp types.Timestamp, output *Output) error {
	if output.SpenditureData != nil {
		return fmt.Errorf("cannot apply output %s to balance: non-nil spenditure data", output.ID.String())
	}
	if output.UnlockReferencePoint.Reached(height, timestamp) {
		balance.Unlocked = balance.Unlocked.Sub(output.Value)
	} else {
		balance.Locked = balance.Locked.Sub(output.Value)
	}
	return nil
}

func (balance *Balance) ApplyUnlockedOutput(height types.BlockHeight, timestamp types.Timestamp, output *Output) error {
	if output.SpenditureData != nil {
		return fmt.Errorf("cannot apply unlocked output %s to balance: non-nil spenditure data", output.ID.String())
	}
	if !output.UnlockReferencePoint.Reached(height, timestamp) {
		return fmt.Errorf(
			"cannot apply output %s to balance as unlocked: height %d and time %d did not reach output reference unlock point %d yet",
			output.ID.String(), height, timestamp, output.UnlockReferencePoint)
	}
	balance.Locked = balance.Locked.Sub(output.Value)
	balance.Unlocked = balance.Unlocked.Add(output.Value)
	return nil
}

func (balance *Balance) RevertUnlockedOutput(height types.BlockHeight, timestamp types.Timestamp, output *Output) error {
	if output.SpenditureData != nil {
		return fmt.Errorf("cannot revert unlocked output %s to balance: non-nil spenditure data", output.ID.String())
	}
	if !output.UnlockReferencePoint.Reached(height, timestamp) {
		return fmt.Errorf(
			"cannot revert output %s to balance as unlocked: height %d and time %d did not reach output reference unlock point %d yet",
			output.ID.String(), height, timestamp, output.UnlockReferencePoint)
	}
	balance.Locked = balance.Locked.Add(output.Value)
	balance.Unlocked = balance.Unlocked.Sub(output.Value)
	return nil
}

// func (wallet *WalletData) RevertBlock(blockID types.BlockID) error {
// 	blockLength := len(wallet.Blocks)
// 	if blockLength == 0 {
// 		return fmt.Errorf("failed to revert block %s from wallet %s: no blocks are registered", blockID.String(), wallet.UnlockHash.String())
// 	}
// 	if bytes.Compare(wallet.Blocks[blockLength-1][:], blockID[:]) != 0 {
// 		return fmt.Errorf("failed to revert block %s from free-for-all wallet %s: unexpected last block of %s", blockID.String(), wallet.UnlockHash.String(), wallet.Blocks[blockLength-1].String())
// 	}
// 	wallet.Blocks = wallet.Blocks[:blockLength-1]
// 	return nil
// }

// func (wallet *WalletData) RevertTransactions(txns ...types.TransactionID) error {
// 	if len(txns) == 0 {
// 		return nil
// 	}
// 	txnMap := make(map[types.TransactionID]int)
// 	for idx, txn := range wallet.Transactions {
// 		txnMap[txn] = idx
// 	}
// 	for _, txn := range txns {
// 		delete(txnMap, txn)
// 	}
// 	wallet.Transactions = make([]types.TransactionID, len(txnMap))
// 	for txnID, idx := range txnMap {
// 		copy(wallet.Transactions[idx][:], txnID[:])
// 	}
// 	return nil
// }

// func (wallet *WalletData) RevertCoinOutput(outputID types.OutputID) error {
// 	indexToDelete := -1
// 	for idx, existingOutputID := range wallet.CoinOutputs {
// 		if bytes.Compare(existingOutputID[:], outputID[:]) == 0 {
// 			indexToDelete = idx
// 			break
// 		}
// 	}
// 	if indexToDelete == -1 {
// 		return fmt.Errorf("could not delete coin output %s from outputs of wallet %s: not found", outputID.String(), wallet.UnlockHash.String())
// 	}
// 	wallet.CoinOutputs = append(wallet.CoinOutputs[:indexToDelete], wallet.CoinOutputs[indexToDelete+1:]...)
// 	return nil
// }

// func (wallet *WalletData) RevertBlockStakeOutput(outputID types.OutputID) error {
// 	indexToDelete := -1
// 	for idx, existingOutputID := range wallet.BlockStakeOutputs {
// 		if bytes.Compare(existingOutputID[:], outputID[:]) == 0 {
// 			indexToDelete = idx
// 			break
// 		}
// 	}
// 	if indexToDelete == -1 {
// 		return fmt.Errorf("could not delete block stake output %s from outputs of wallet %s: not found", outputID.String(), wallet.UnlockHash.String())
// 	}
// 	wallet.BlockStakeOutputs = append(wallet.BlockStakeOutputs[:indexToDelete], wallet.BlockStakeOutputs[indexToDelete+1:]...)
// 	return nil
// }

func (swallet *SingleSignatureWalletData) AddMultiSignatureWallets(uhs ...types.UnlockHash) {
	uhm := make(map[types.UnlockHash]struct{}, len(swallet.MultiSignatureWallets))
	// add already known addresses to dedup map
	for _, uh := range swallet.MultiSignatureWallets {
		uhm[uh] = struct{}{}
	}
	// add possible new addresses to dedup map
	for _, uh := range uhs {
		uhm[uh] = struct{}{}
	}
	swallet.MultiSignatureWallets = make([]types.UnlockHash, 0, len(uhm))
	for uh := range uhm {
		swallet.MultiSignatureWallets = append(swallet.MultiSignatureWallets, uh)
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
	ObjectTypeFreeForAllWallet
	ObjectTypeSingleSignatureWallet
	ObjectTypeMultiSignatureWallet
	ObjectTypeAtomicSwapContract
)

type (
	ObjectID []byte

	ByteVersion uint8

	ObjectInfo struct {
		Type    ObjectType
		Version ByteVersion // used for objects that have individual versions that cannot be interfered from the ID
	}

	Object struct {
		ID      ObjectID
		Type    ObjectType
		Version ByteVersion // used for objects that have individual versions that cannot be interfered from the ID
		Data    interface{}
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

func (obj *Object) AsBlock() (Block, error) {
	if obj.Type != ObjectTypeBlock {
		return Block{}, fmt.Errorf("object of type %d cannot be casted to block (type %d)", obj.Type, ObjectTypeBlock)
	}
	block, ok := obj.Data.(Block)
	if !ok {
		return Block{}, fmt.Errorf("object of type %d has block type but has wrong underlying data type of %T (expected: %T)", obj.Type, obj.Data, Block{})
	}
	if obj.Version != 0 {
		return Block{}, fmt.Errorf("mismatch of object version (%d) and expected unique block version of 0", obj.Version)
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
	if ver := types.TransactionVersion(obj.Version); ver != txn.Version {
		return Transaction{}, fmt.Errorf("mismatch of object version (%d) and transaction version (%d)", ver, txn.Version)
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
	if obj.Version != 0 {
		return Output{}, fmt.Errorf("mismatch of object version (%d) and expected unique output version of 0", obj.Version)
	}
	return output, nil
}

func (obj *Object) AsFreeForAllWallet() (FreeForAllWalletData, error) {
	if obj.Type != ObjectTypeSingleSignatureWallet {
		return FreeForAllWalletData{}, fmt.Errorf("object of type %d cannot be casted to free-for-all wallet (type %d)", obj.Type, ObjectTypeSingleSignatureWallet)
	}
	wd, ok := obj.Data.(FreeForAllWalletData)
	if !ok {
		return FreeForAllWalletData{}, fmt.Errorf("object of type %d has free-for-all wallet type but has wrong underlying data type of %T (expected: %T)", obj.Type, obj.Data, FreeForAllWalletData{})
	}
	if ut := types.UnlockType(obj.Version); ut != types.UnlockTypeNil {
		return FreeForAllWalletData{}, fmt.Errorf("mismatch of object version (%d) and expected unlock type (%d)", ut, types.UnlockTypeNil)
	}
	return wd, nil
}

func (obj *Object) AsSingleSignatureWallet() (SingleSignatureWalletData, error) {
	if obj.Type != ObjectTypeSingleSignatureWallet {
		return SingleSignatureWalletData{}, fmt.Errorf("object of type %d cannot be casted to wallet (type %d)", obj.Type, ObjectTypeSingleSignatureWallet)
	}
	wd, ok := obj.Data.(SingleSignatureWalletData)
	if !ok {
		return SingleSignatureWalletData{}, fmt.Errorf("object of type %d has wallet type but has wrong underlying data type of %T (expected: %T)", obj.Type, obj.Data, SingleSignatureWalletData{})
	}
	if ut := types.UnlockType(obj.Version); ut != types.UnlockTypePubKey {
		return SingleSignatureWalletData{}, fmt.Errorf("mismatch of object version (%d) and expected unlock type (%d)", ut, types.UnlockTypePubKey)
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
	if ut := types.UnlockType(obj.Version); ut != types.UnlockTypeMultiSig {
		return MultiSignatureWalletData{}, fmt.Errorf("mismatch of object version (%d) and expected unlock type (%d)", ut, types.UnlockTypeMultiSig)
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
	if ut := types.UnlockType(obj.Version); ut != types.UnlockTypeAtomicSwap {
		return AtomicSwapContract{}, fmt.Errorf("mismatch of object version (%d) and expected unlock type (%d)", ut, types.UnlockTypeAtomicSwap)
	}
	return asc, nil
}
