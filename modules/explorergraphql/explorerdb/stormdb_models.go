package explorerdb

import (
	"fmt"

	"github.com/asdine/storm"
	"github.com/asdine/storm/q"

	"github.com/threefoldtech/rivine/crypto"
	"github.com/threefoldtech/rivine/types"
)

// TODO: link DataID instead of ObjectID

type (
	stormDataID uint64

	stormObject struct {
		ObjectID      ObjectID    `storm:"id"`
		ObjectType    ObjectType  `storm:"index"`
		ObjectVersion ByteVersion `storm:"index"`
		DataID        stormDataID `storm:"unique"`
	}

	stormBlock struct {
		DataID stormDataID `storm:"id"`

		ParentID  types.BlockID
		Height    types.BlockHeight
		Timestamp types.Timestamp
		Payouts   []types.OutputID

		Transactions []types.TransactionID
	}

	stormBlockFacts struct {
		DataID stormDataID `storm:"id"`

		Target     types.Target     `storm:"index"`
		Difficulty types.Difficulty `storm:"index"`

		AggregatedTotalCoins                 types.Currency `storm:"index"`
		AggregatedTotalLockedCoins           types.Currency `storm:"index"`
		AggregatedTotalBlockStakes           types.Currency `storm:"index"`
		AggregatedTotalLockedBlockStakes     types.Currency `storm:"index"`
		AggregatedEstimatedActiveBlockStakes types.Currency `storm:"index"`
	}

	stormTransaction struct {
		DataID stormDataID `storm:"id"`

		ParentBlock types.BlockID
		Version     types.TransactionVersion `storm:"index"`

		CoinInputs  []types.OutputID
		CoinOutputs []types.OutputID

		BlockStakeInputs  []types.OutputID
		BlockStakeOutputs []types.OutputID

		FeePayout TransactionFeePayoutInfo

		ArbitraryData        []byte
		EncodedExtensionData []byte
	}

	stormOutput struct {
		DataID stormDataID `storm:"id"`

		ParentID crypto.Hash

		Type                 OutputType `storm:"index"`
		Value                types.Currency
		Condition            types.UnlockConditionProxy
		UnlockReferencePoint ReferencePoint `storm:"index"`

		SpenditureData *OutputSpenditureData `storm:"index"`
	}

	stormBaseWalletData struct {
		DataID stormDataID `storm:"id"`

		CoinOutputs       []types.OutputID
		BlockStakeOutputs []types.OutputID
		Blocks            []types.BlockID
		Transactions      []types.TransactionID

		CoinsUnlocked types.Currency `storm:"index"`
		CoinsLocked   types.Currency `storm:"index"`

		BlockStakesUnlocked types.Currency `storm:"index"`
		BlockStakesLocked   types.Currency `storm:"index"`
	}

	stormFreeForAllWalletData struct {
		stormBaseWalletData `storm:"inline"`
	}

	stormSingleSignatureWalletData struct {
		stormBaseWalletData `storm:"inline"`

		MultiSignatureWallets []types.UnlockHash
	}

	stormMultiSignatureWalletData struct {
		stormBaseWalletData `storm:"inline"`

		Owners                []types.UnlockHash
		RequiredSgnatureCount int
	}

	stormAtomicSwapContract struct {
		DataID stormDataID `storm:"id"`

		ContractValue     types.Currency
		ContractCondition types.AtomicSwapCondition
		Transactions      []types.TransactionID
		CoinInput         types.CoinOutputID

		SpenditureData *AtomicSwapContractSpenditureData `storm:"index"`
	}
)

func (sblock *stormBlock) AsBlock(blockID types.BlockID) Block {
	return Block{
		ID:        blockID,
		ParentID:  sblock.ParentID,
		Height:    sblock.Height,
		Timestamp: sblock.Timestamp,
		Payouts:   sblock.Payouts,

		Transactions: sblock.Transactions,
	}
}

func (sbfacts *stormBlockFacts) AsBlockFacts() BlockFacts {
	return BlockFacts{
		Constants: BlockFactsConstants{
			Target:     sbfacts.Target,
			Difficulty: sbfacts.Difficulty,
		},
		Aggregated: BlockFactsAggregated{
			TotalCoins:                 sbfacts.AggregatedTotalCoins,
			TotalLockedCoins:           sbfacts.AggregatedTotalLockedCoins,
			TotalBlockStakes:           sbfacts.AggregatedTotalBlockStakes,
			TotalLockedBlockStakes:     sbfacts.AggregatedTotalLockedBlockStakes,
			EstimatedActiveBlockStakes: sbfacts.AggregatedEstimatedActiveBlockStakes,
		},
	}
}

func (stxn *stormTransaction) AsTransaction(txnID types.TransactionID) Transaction {
	return Transaction{
		ID: txnID,

		ParentBlock: stxn.ParentBlock,
		Version:     stxn.Version,

		CoinInputs:  stxn.CoinInputs,
		CoinOutputs: stxn.CoinOutputs,

		BlockStakeInputs:  stxn.BlockStakeInputs,
		BlockStakeOutputs: stxn.BlockStakeOutputs,

		FeePayout: stxn.FeePayout,

		ArbitraryData:        stxn.ArbitraryData,
		EncodedExtensionData: stxn.EncodedExtensionData,
	}
}

func (sout *stormOutput) AsOutput(outputID types.OutputID) Output {
	return Output{
		ID:       outputID,
		ParentID: sout.ParentID,

		Type:                 sout.Type,
		Value:                sout.Value,
		Condition:            sout.Condition,
		UnlockReferencePoint: sout.UnlockReferencePoint,
		SpenditureData:       sout.SpenditureData,
	}
}

func walletDataAsSDB(dataID stormDataID, wallet *WalletData) *stormBaseWalletData {
	return &stormBaseWalletData{
		DataID: dataID,

		CoinOutputs:       wallet.CoinOutputs,
		BlockStakeOutputs: wallet.BlockStakeOutputs,
		Blocks:            wallet.Blocks,
		Transactions:      wallet.Transactions,

		CoinsUnlocked: wallet.CoinBalance.Unlocked,
		CoinsLocked:   wallet.CoinBalance.Locked,

		BlockStakesUnlocked: wallet.BlockStakeBalance.Unlocked,
		BlockStakesLocked:   wallet.BlockStakeBalance.Locked,
	}
}

func (swallet *stormBaseWalletData) AsWalletData(uh types.UnlockHash) WalletData {
	return WalletData{
		UnlockHash: uh,

		CoinOutputs:       swallet.CoinOutputs,
		BlockStakeOutputs: swallet.BlockStakeOutputs,
		Blocks:            swallet.Blocks,
		Transactions:      swallet.Transactions,

		CoinBalance: Balance{
			Unlocked: swallet.CoinsUnlocked,
			Locked:   swallet.CoinsLocked,
		},
		BlockStakeBalance: Balance{
			Unlocked: swallet.BlockStakesUnlocked,
			Locked:   swallet.BlockStakesLocked,
		},
	}
}

func (swallet *stormFreeForAllWalletData) AsFreeForAllWallet(uh types.UnlockHash) FreeForAllWalletData {
	return FreeForAllWalletData{
		WalletData: swallet.AsWalletData(uh),
	}
}

func (swallet *stormSingleSignatureWalletData) AsSingleSignatureWallet(uh types.UnlockHash) SingleSignatureWalletData {
	return SingleSignatureWalletData{
		WalletData:            swallet.AsWalletData(uh),
		MultiSignatureWallets: swallet.MultiSignatureWallets,
	}
}

func (swallet *stormMultiSignatureWalletData) AsMultiSignatureWallet(uh types.UnlockHash) MultiSignatureWalletData {
	return MultiSignatureWalletData{
		WalletData:            swallet.AsWalletData(uh),
		Owners:                swallet.Owners,
		RequiredSgnatureCount: swallet.RequiredSgnatureCount,
	}
}

func (scontract *stormAtomicSwapContract) AsAtomicSwapContract(uh types.UnlockHash) AtomicSwapContract {
	return AtomicSwapContract{
		UnlockHash: uh,

		ContractValue:     scontract.ContractValue,
		ContractCondition: scontract.ContractCondition,
		Transactions:      scontract.Transactions,
		CoinInput:         scontract.CoinInput,

		SpenditureData: scontract.SpenditureData,
	}
}

const (
	nodeObjectKeyObjectID   = "ObjectID"
	nodeObjectKeyDataID     = "DataID"
	nodeObjectKeyUnlockHash = "UnlockHash"

	nodeObjectOutputKeySpenditureData = "SpenditureData"
)

type (
	stormObjectNodeReader interface {
		GetObject(ObjectID) (Object, error)
		GetObjectInfo(ObjectID) (ObjectInfo, error)

		GetBlock(types.BlockID) (Block, error)
		GetBlockFacts(types.BlockID) (BlockFacts, error)
		GetTransaction(types.TransactionID) (Transaction, error)
		GetOutput(types.OutputID) (Output, error)
		GetFreeForAllWallet(types.UnlockHash) (FreeForAllWalletData, error)
		GetSingleSignatureWallet(types.UnlockHash) (SingleSignatureWalletData, error)
		GetMultiSignatureWallet(types.UnlockHash) (MultiSignatureWalletData, error)
		GetAtomicSwapContract(types.UnlockHash) (AtomicSwapContract, error)

		GetStormOutputsbyUnlockReferencePoint(types.BlockHeight, types.Timestamp) ([]stormOutput, error)
	}

	stormObjectNodeReaderWriter interface {
		stormObjectNodeReader

		GetLastDataID() stormDataID

		SaveBlockWithFacts(*Block, *BlockFacts) error
		SaveTransaction(*Transaction) error
		SaveOutput(*Output) error
		SaveFreeForAllWallet(*FreeForAllWalletData) error
		SaveSingleSignatureWallet(*SingleSignatureWalletData) error
		SaveMultiSignatureWallet(*MultiSignatureWalletData) error
		SaveAtomicSwapContract(*AtomicSwapContract) error

		UpdateOutputSpenditureData(types.OutputID, *OutputSpenditureData) (Output, error)

		DeleteBlock(types.BlockID) (Block, error)
		DeleteTransaction(types.TransactionID) (Transaction, error)
		DeleteOutput(types.OutputID) (Output, error)
		DeleteFreeForAllWallet(types.UnlockHash) error
		DeleteSingleSignatureWallet(types.UnlockHash) error
		DeleteMultiSignatureWallet(types.UnlockHash) error
		DeleteAtomicSwapContract(types.UnlockHash) error
	}
)

const (
	nodeNameObjects = "Objects"
)

type stormObjectNode struct {
	node       storm.Node
	lastDataID stormDataID
}

func newStormObjectNodeReader(db *storm.DB) stormObjectNodeReader {
	return &stormObjectNode{
		node: db.From(nodeNameObjects),
	}
}

func newStormObjectNodeReaderWriter(db *storm.DB, lastDataID stormDataID) stormObjectNodeReaderWriter {
	return &stormObjectNode{
		node:       db.From(nodeNameObjects),
		lastDataID: lastDataID,
	}
}

func (son *stormObjectNode) nextDataID() (dataID stormDataID) {
	son.lastDataID++
	dataID = son.lastDataID
	return
}

func (son *stormObjectNode) GetLastDataID() stormDataID {
	return son.lastDataID
}

func (son *stormObjectNode) SaveBlockWithFacts(block *Block, facts *BlockFacts) error {
	obj := stormObject{
		ObjectID:      ObjectID(block.ID[:]),
		ObjectType:    ObjectTypeBlock,
		ObjectVersion: 0, // there is only one version of blocks
		DataID:        son.nextDataID(),
	}
	err := son.node.Save(&obj)
	if err != nil {
		return fmt.Errorf("failed to save object type info for block %s: %v", block.ID.String(), err)
	}
	err = son.node.Save(&stormBlock{
		DataID: obj.DataID, // automatically incremented in previous save call

		ParentID:  block.ParentID,
		Height:    block.Height,
		Timestamp: block.Timestamp,
		Payouts:   block.Payouts,

		Transactions: block.Transactions,
	})
	if err != nil {
		return fmt.Errorf("failed to save block %s by (object) data ID %d: %v", block.ID.String(), obj.DataID, err)
	}
	err = son.node.Save(&stormBlockFacts{
		DataID: obj.DataID, // automatically incremented in previous save call

		Target:     facts.Constants.Target,
		Difficulty: facts.Constants.Difficulty,

		AggregatedTotalCoins:                 facts.Aggregated.TotalCoins,
		AggregatedTotalLockedCoins:           facts.Aggregated.TotalLockedCoins,
		AggregatedTotalBlockStakes:           facts.Aggregated.TotalBlockStakes,
		AggregatedTotalLockedBlockStakes:     facts.Aggregated.TotalLockedBlockStakes,
		AggregatedEstimatedActiveBlockStakes: facts.Aggregated.EstimatedActiveBlockStakes,
	})
	if err != nil {
		return fmt.Errorf("failed to save facts for block %s by (object) data ID %d: %v", block.ID.String(), obj.DataID, err)
	}
	return nil
}

func (son *stormObjectNode) SaveTransaction(txn *Transaction) error {
	obj := stormObject{
		ObjectID:      ObjectID(txn.ID[:]),
		ObjectType:    ObjectTypeTransaction,
		ObjectVersion: ByteVersion(txn.Version),
		DataID:        son.nextDataID(),
	}
	err := son.node.Save(&obj)
	if err != nil {
		return fmt.Errorf("failed to save object type info for transaction %s: %v", txn.ID.String(), err)
	}
	err = son.node.Save(&stormTransaction{
		DataID: obj.DataID, // automatically incremented in previous save call

		ParentBlock: txn.ParentBlock,
		Version:     txn.Version,

		CoinInputs:  txn.CoinInputs,
		CoinOutputs: txn.CoinOutputs,

		BlockStakeInputs:  txn.BlockStakeInputs,
		BlockStakeOutputs: txn.BlockStakeOutputs,

		FeePayout: txn.FeePayout,

		ArbitraryData:        txn.ArbitraryData,
		EncodedExtensionData: txn.EncodedExtensionData,
	})
	if err != nil {
		return fmt.Errorf("failed to save transaction %s by (object) data ID %d: %v", txn.ID.String(), obj.DataID, err)
	}
	return nil
}

func (son *stormObjectNode) SaveOutput(output *Output) error {
	obj := stormObject{
		ObjectID:      ObjectID(output.ID[:]),
		ObjectType:    ObjectTypeOutput,
		ObjectVersion: 0, // there is only one version of outputs
		DataID:        son.nextDataID(),
	}
	err := son.node.Save(&obj)
	if err != nil {
		return fmt.Errorf("failed to save object type info for output %s: %v", output.ID.String(), err)
	}
	err = son.node.Save(&stormOutput{
		DataID: obj.DataID, // automatically incremented in previous save call

		ParentID: output.ParentID,

		Type:                 output.Type,
		Value:                output.Value,
		Condition:            output.Condition,
		UnlockReferencePoint: output.UnlockReferencePoint,

		SpenditureData: output.SpenditureData,
	})
	if err != nil {
		return fmt.Errorf("failed to save output %s by (object) data ID %d: %v", output.ID.String(), obj.DataID, err)
	}
	return nil
}

func (son *stormObjectNode) SaveFreeForAllWallet(wallet *FreeForAllWalletData) error {
	obj := stormObject{
		ObjectID:      ObjectIDFromUnlockHash(wallet.UnlockHash),
		ObjectType:    ObjectTypeSingleSignatureWallet,
		ObjectVersion: ByteVersion(wallet.UnlockHash.Type),
		DataID:        son.nextDataID(),
	}
	err := son.node.Save(&obj)
	if err != nil {
		return fmt.Errorf("failed to save object type info for wallet %s: %v", wallet.UnlockHash.String(), err)
	}
	err = son.node.Save(&stormFreeForAllWalletData{
		// DataID was automatically incremented in previous save call
		stormBaseWalletData: *walletDataAsSDB(obj.DataID, &wallet.WalletData),
	})
	if err != nil {
		return fmt.Errorf("failed to save wallet %s by (object) data ID %d: %v", wallet.UnlockHash.String(), obj.DataID, err)
	}
	return nil
}

func (son *stormObjectNode) SaveSingleSignatureWallet(wallet *SingleSignatureWalletData) error {
	obj := stormObject{
		ObjectID:      ObjectIDFromUnlockHash(wallet.UnlockHash),
		ObjectType:    ObjectTypeSingleSignatureWallet,
		ObjectVersion: ByteVersion(wallet.UnlockHash.Type),
		DataID:        son.nextDataID(),
	}
	err := son.node.Save(&obj)
	if err != nil {
		return fmt.Errorf("failed to save object type info for wallet %s: %v", wallet.UnlockHash.String(), err)
	}
	err = son.node.Save(&stormSingleSignatureWalletData{
		// DataID was automatically incremented in previous save call
		stormBaseWalletData:   *walletDataAsSDB(obj.DataID, &wallet.WalletData),
		MultiSignatureWallets: wallet.MultiSignatureWallets,
	})
	if err != nil {
		return fmt.Errorf("failed to save wallet %s by (object) data ID %d: %v", wallet.UnlockHash.String(), obj.DataID, err)
	}
	return nil
}

func (son *stormObjectNode) SaveMultiSignatureWallet(wallet *MultiSignatureWalletData) error {
	obj := stormObject{
		ObjectID:      ObjectIDFromUnlockHash(wallet.UnlockHash),
		ObjectType:    ObjectTypeMultiSignatureWallet,
		ObjectVersion: ByteVersion(wallet.UnlockHash.Type),
		DataID:        son.nextDataID(),
	}
	err := son.node.Save(&obj)
	if err != nil {
		return fmt.Errorf("failed to save object type info for wallet %s: %v", wallet.UnlockHash.String(), err)
	}
	err = son.node.Save(&stormMultiSignatureWalletData{
		// DataID was automatically incremented in previous save call
		stormBaseWalletData:   *walletDataAsSDB(obj.DataID, &wallet.WalletData),
		Owners:                wallet.Owners,
		RequiredSgnatureCount: wallet.RequiredSgnatureCount,
	})
	if err != nil {
		return fmt.Errorf("failed to save wallet %s by (object) data ID %d: %v", wallet.UnlockHash.String(), obj.DataID, err)
	}
	return nil
}

func (son *stormObjectNode) SaveAtomicSwapContract(contract *AtomicSwapContract) error {
	obj := stormObject{
		ObjectID:      ObjectIDFromUnlockHash(contract.UnlockHash),
		ObjectType:    ObjectTypeAtomicSwapContract,
		ObjectVersion: ByteVersion(contract.UnlockHash.Type),
		DataID:        son.nextDataID(),
	}
	err := son.node.Save(&obj)
	if err != nil {
		return fmt.Errorf("failed to save object type info for contract %s: %v", contract.UnlockHash.String(), err)
	}
	err = son.node.Save(&stormAtomicSwapContract{
		DataID: obj.DataID, // automatically incremented in previous save call

		ContractValue:     contract.ContractValue,
		ContractCondition: contract.ContractCondition,
		Transactions:      contract.Transactions,
		CoinInput:         contract.CoinInput,

		SpenditureData: contract.SpenditureData,
	})
	if err != nil {
		return fmt.Errorf("failed to save contract %s by (object) data ID %d: %v", contract.UnlockHash.String(), obj.DataID, err)
	}
	return nil
}

func (son *stormObjectNode) GetObject(objectID ObjectID) (Object, error) {
	var object stormObject
	err := son.node.One(nodeObjectKeyObjectID, objectID, &object)
	if err != nil {
		return Object{}, err
	}
	obj := Object{
		ID:      objectID,
		Type:    object.ObjectType,
		Version: object.ObjectVersion,
	}
	switch object.ObjectType {
	case ObjectTypeBlock:
		blockHash, err := objectID.AsHash()
		if err != nil {
			return Object{}, err
		}
		obj.Data, err = son.getBlockByDataID(types.BlockID(blockHash), object.DataID)
	case ObjectTypeTransaction:
		txnHash, err := objectID.AsHash()
		if err != nil {
			return Object{}, err
		}
		var txn Transaction
		txn, err = son.getTransactionByDataID(types.TransactionID(txnHash), object.DataID)
		obj.Data = txn
	case ObjectTypeOutput:
		outputHash, err := objectID.AsHash()
		if err != nil {
			return Object{}, err
		}
		obj.Data, err = son.getOutputByDataID(types.OutputID(outputHash), object.DataID)
	case ObjectTypeFreeForAllWallet:
		uh, err := objectID.AsUnlockHash()
		if err != nil {
			return Object{}, err
		}
		obj.Data, err = son.getFreeForAllWalletByDataID(uh, object.DataID)
	case ObjectTypeSingleSignatureWallet:
		uh, err := objectID.AsUnlockHash()
		if err != nil {
			return Object{}, err
		}
		obj.Data, err = son.getSingleSignatureWalletByDataID(uh, object.DataID)
	case ObjectTypeMultiSignatureWallet:
		uh, err := objectID.AsUnlockHash()
		if err != nil {
			return Object{}, err
		}
		obj.Data, err = son.getMultiSignatureWalletByDataID(uh, object.DataID)
	case ObjectTypeAtomicSwapContract:
		uh, err := objectID.AsUnlockHash()
		if err != nil {
			return Object{}, err
		}
		obj.Data, err = son.getAtomicSwapContractByDataID(uh, object.DataID)
	default:
		err = fmt.Errorf("object type %d is unknown or is not expected to have data", object.ObjectType)
	}
	return obj, err
}

func (son *stormObjectNode) GetObjectInfo(objectID ObjectID) (ObjectInfo, error) {
	var object stormObject
	err := son.node.One(nodeObjectKeyObjectID, objectID, &object)
	if err != nil {
		return ObjectInfo{}, err
	}
	return ObjectInfo{
		Type:    object.ObjectType,
		Version: object.ObjectVersion,
	}, nil
}

func (son *stormObjectNode) getTypedObject(objectID ObjectID, value interface{}) error {
	var object stormObject
	err := son.node.One(nodeObjectKeyObjectID, objectID, &object)
	if err != nil {
		return err
	}
	return son.getTypedObjectByDataID(object.DataID, value)
}

func (son *stormObjectNode) getTypedObjectByDataID(dataID stormDataID, value interface{}) error {
	return son.node.One(nodeObjectKeyDataID, dataID, value)
}

func (son *stormObjectNode) GetBlock(blockID types.BlockID) (Block, error) {
	var sblock stormBlock
	err := son.getTypedObject(ObjectID(blockID[:]), &sblock)
	if err != nil {
		return Block{}, err
	}
	return sblock.AsBlock(blockID), nil
}

func (son *stormObjectNode) GetBlockFacts(blockID types.BlockID) (BlockFacts, error) {
	var sbfacts stormBlockFacts
	err := son.getTypedObject(ObjectID(blockID[:]), &sbfacts)
	if err != nil {
		return BlockFacts{}, err
	}
	return sbfacts.AsBlockFacts(), nil
}

func (son *stormObjectNode) getBlockByDataID(blockID types.BlockID, dataID stormDataID) (Block, error) {
	var sblock stormBlock
	err := son.getTypedObjectByDataID(dataID, &sblock)
	if err != nil {
		return Block{}, err
	}
	return sblock.AsBlock(blockID), nil
}

func (son *stormObjectNode) GetTransaction(transactionID types.TransactionID) (Transaction, error) {
	var stxn stormTransaction
	err := son.getTypedObject(ObjectID(transactionID[:]), &stxn)
	if err != nil {
		return Transaction{}, err
	}
	return stxn.AsTransaction(transactionID), nil
}

func (son *stormObjectNode) getTransactionByDataID(transactionID types.TransactionID, dataID stormDataID) (Transaction, error) {
	var stxn stormTransaction
	err := son.getTypedObjectByDataID(dataID, &stxn)
	if err != nil {
		return Transaction{}, err
	}
	return stxn.AsTransaction(transactionID), nil
}

func (son *stormObjectNode) GetOutput(outputID types.OutputID) (Output, error) {
	var sout stormOutput
	err := son.getTypedObject(ObjectID(outputID[:]), &sout)
	if err != nil {
		return Output{}, err
	}
	return sout.AsOutput(outputID), nil
}

func (son *stormObjectNode) getOutputByDataID(outputID types.OutputID, dataID stormDataID) (Output, error) {
	var sout stormOutput
	err := son.getTypedObjectByDataID(dataID, &sout)
	if err != nil {
		return Output{}, err
	}
	return sout.AsOutput(outputID), nil
}

func (son *stormObjectNode) GetFreeForAllWallet(uh types.UnlockHash) (FreeForAllWalletData, error) {
	var swallet stormFreeForAllWalletData
	err := son.getTypedObject(ObjectIDFromUnlockHash(uh), &swallet)
	if err != nil {
		return FreeForAllWalletData{}, err
	}
	return swallet.AsFreeForAllWallet(uh), nil
}

func (son *stormObjectNode) getFreeForAllWalletByDataID(uh types.UnlockHash, dataID stormDataID) (FreeForAllWalletData, error) {
	var swallet stormFreeForAllWalletData
	err := son.getTypedObjectByDataID(dataID, &swallet)
	if err != nil {
		return FreeForAllWalletData{}, err
	}
	return swallet.AsFreeForAllWallet(uh), nil
}

func (son *stormObjectNode) GetSingleSignatureWallet(uh types.UnlockHash) (SingleSignatureWalletData, error) {
	var swallet stormSingleSignatureWalletData
	err := son.getTypedObject(ObjectIDFromUnlockHash(uh), &swallet)
	if err != nil {
		return SingleSignatureWalletData{}, err
	}
	return swallet.AsSingleSignatureWallet(uh), nil
}

func (son *stormObjectNode) getSingleSignatureWalletByDataID(uh types.UnlockHash, dataID stormDataID) (SingleSignatureWalletData, error) {
	var swallet stormSingleSignatureWalletData
	err := son.getTypedObjectByDataID(dataID, &swallet)
	if err != nil {
		return SingleSignatureWalletData{}, err
	}
	return swallet.AsSingleSignatureWallet(uh), nil
}

func (son *stormObjectNode) GetMultiSignatureWallet(uh types.UnlockHash) (MultiSignatureWalletData, error) {
	var swallet stormMultiSignatureWalletData
	err := son.getTypedObject(ObjectIDFromUnlockHash(uh), &swallet)
	if err != nil {
		return MultiSignatureWalletData{}, err
	}
	return swallet.AsMultiSignatureWallet(uh), nil
}

func (son *stormObjectNode) getMultiSignatureWalletByDataID(uh types.UnlockHash, dataID stormDataID) (MultiSignatureWalletData, error) {
	var swallet stormMultiSignatureWalletData
	err := son.getTypedObjectByDataID(dataID, &swallet)
	if err != nil {
		return MultiSignatureWalletData{}, err
	}
	return swallet.AsMultiSignatureWallet(uh), nil
}

func (son *stormObjectNode) GetAtomicSwapContract(uh types.UnlockHash) (AtomicSwapContract, error) {
	var scontract stormAtomicSwapContract
	err := son.getTypedObject(ObjectIDFromUnlockHash(uh), &scontract)
	if err != nil {
		return AtomicSwapContract{}, err
	}
	return scontract.AsAtomicSwapContract(uh), nil
}

func (son *stormObjectNode) getAtomicSwapContractByDataID(uh types.UnlockHash, dataID stormDataID) (AtomicSwapContract, error) {
	var scontract stormAtomicSwapContract
	err := son.getTypedObjectByDataID(dataID, &scontract)
	if err != nil {
		return AtomicSwapContract{}, err
	}
	return scontract.AsAtomicSwapContract(uh), nil
}

func (son *stormObjectNode) GetStormOutputsbyUnlockReferencePoint(height types.BlockHeight, timestamp types.Timestamp) (outputs []stormOutput, err error) {
	err = son.node.Select(q.Or(
		q.Eq("UnlockReferencePoint", ReferencePoint(height)),
		q.Eq("UnlockReferencePoint", ReferencePoint(timestamp)))).Find(&outputs)
	return
}

func (son *stormObjectNode) UpdateOutputSpenditureData(outputID types.OutputID, spenditureData *OutputSpenditureData) (Output, error) {
	var object stormObject
	err := son.node.One(nodeObjectKeyObjectID, ObjectID(outputID[:]), &object)
	if err != nil {
		return Output{}, err
	}
	var output stormOutput
	err = son.node.One(nodeObjectKeyDataID, object.DataID, &output)
	if err != nil {
		return Output{}, err
	}
	output.SpenditureData = spenditureData
	err = son.node.Update(&output)
	if err != nil {
		return Output{}, err
	}
	return output.AsOutput(outputID), nil
}

func (son *stormObjectNode) DeleteBlock(blockID types.BlockID) (Block, error) {
	var block stormBlock
	err := son.deleteObject(ObjectID(blockID[:]), &block)
	if err != nil {
		return Block{}, err
	}
	return block.AsBlock(blockID), nil
}

func (son *stormObjectNode) DeleteTransaction(txnID types.TransactionID) (Transaction, error) {
	var txn stormTransaction
	err := son.deleteObject(ObjectID(txnID[:]), &txn)
	if err != nil {
		return Transaction{}, err
	}
	return txn.AsTransaction(txnID), nil
}

func (son *stormObjectNode) DeleteOutput(outputID types.OutputID) (Output, error) {
	var output stormOutput
	err := son.deleteObject(ObjectID(outputID[:]), &output)
	if err != nil {
		return Output{}, err
	}
	return output.AsOutput(outputID), nil
}

func (son *stormObjectNode) DeleteFreeForAllWallet(uh types.UnlockHash) error {
	return son.deleteObject(ObjectIDFromUnlockHash(uh), new(stormFreeForAllWalletData))
}

func (son *stormObjectNode) DeleteSingleSignatureWallet(uh types.UnlockHash) error {
	return son.deleteObject(ObjectIDFromUnlockHash(uh), new(stormSingleSignatureWalletData))
}

func (son *stormObjectNode) DeleteMultiSignatureWallet(uh types.UnlockHash) error {
	return son.deleteObject(ObjectIDFromUnlockHash(uh), new(stormMultiSignatureWalletData))
}

func (son *stormObjectNode) DeleteAtomicSwapContract(uh types.UnlockHash) error {
	return son.deleteObject(ObjectIDFromUnlockHash(uh), new(stormAtomicSwapContract))
}

func (son *stormObjectNode) deleteObject(objectID ObjectID, value interface{}) error {
	// delete object metadata
	var object stormObject
	err := son.deleteByID(nodeObjectKeyObjectID, objectID, &object)
	if err != nil {
		return fmt.Errorf("failed to delete object metadata for %x: %v", objectID[:], err)
	}
	// delete object (output) data
	return son.deleteByID(nodeObjectKeyDataID, object.DataID, value)
}

func (son *stormObjectNode) deleteByID(fieldName string, ID interface{}, value interface{}) error {
	err := son.node.One(fieldName, ID, value)
	if err != nil {
		return err
	}
	return son.node.DeleteStruct(value)
}
