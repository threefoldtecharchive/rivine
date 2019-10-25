package explorerdb

import (
	"fmt"

	"github.com/asdine/storm"
	"github.com/asdine/storm/q"

	"github.com/threefoldtech/rivine/crypto"
	"github.com/threefoldtech/rivine/types"
)

type (
	stormDataID uint64

	stormObject struct {
		ObjectID   ObjectID    `storm:"id"`
		ObjectType ObjectType  `storm:"index"`
		DataID     stormDataID `storm:"unique"`
	}

	stormBlock struct {
		DataID stormDataID `storm:"id"`

		ParentID  types.BlockID
		Target    types.Target
		Height    types.BlockHeight
		Timestamp types.Timestamp
		Payouts   []types.OutputID

		Transactions []types.TransactionID
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

	stormSingleSignatureWalletData struct {
		DataID stormDataID `storm:"id"`

		PublicKey types.PublicKey `storm:"unique"`

		CoinOutputs       []types.OutputID
		BlockStakeOutputs types.OutputID
		Transactions      []types.TransactionID

		CoinBalance       Balance
		BlockStakeBalance Balance
	}

	stormMultiSignatureWalletData struct {
		DataID stormDataID `storm:"id"`

		Owners                []MultiSignatureWalletOwner
		RequiredSgnatureCount int

		CoinOutputs       []types.OutputID
		BlockStakeOutputs types.OutputID
		Transactions      []types.TransactionID

		CoinBalance       Balance
		BlockStakeBalance Balance
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
		Target:    sblock.Target,
		Height:    sblock.Height,
		Timestamp: sblock.Timestamp,
		Payouts:   sblock.Payouts,

		Transactions: sblock.Transactions,
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

func (swallet *stormSingleSignatureWalletData) AsSingleSignatureWallet(uh types.UnlockHash) SingleSignatureWalletData {
	return SingleSignatureWalletData{
		WalletData: WalletData{
			UnlockHash: uh,

			CoinOutputs:       swallet.CoinOutputs,
			BlockStakeOutputs: swallet.BlockStakeOutputs,
			Transactions:      swallet.Transactions,

			CoinBalance:       swallet.CoinBalance,
			BlockStakeBalance: swallet.BlockStakeBalance,
		},
		PublicKey: swallet.PublicKey,
	}
}

func (swallet *stormMultiSignatureWalletData) AsMultiSignatureWallet(uh types.UnlockHash) MultiSignatureWalletData {
	return MultiSignatureWalletData{
		WalletData: WalletData{
			UnlockHash: uh,

			CoinOutputs:       swallet.CoinOutputs,
			BlockStakeOutputs: swallet.BlockStakeOutputs,
			Transactions:      swallet.Transactions,

			CoinBalance:       swallet.CoinBalance,
			BlockStakeBalance: swallet.BlockStakeBalance,
		},
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

type stormObjectNode struct {
	node       storm.Node
	lastDataID stormDataID
}

func (son *stormObjectNode) NextDataID() (dataID stormDataID) {
	son.lastDataID++
	dataID = son.lastDataID
	return
}

func (son *stormObjectNode) SaveBlock(block *Block) error {
	obj := stormObject{
		ObjectID:   ObjectID(block.ID[:]),
		ObjectType: ObjectTypeBlock,
		DataID:     son.NextDataID(),
	}
	err := son.node.Save(&obj)
	if err != nil {
		return fmt.Errorf("failed to save object type info for block %s: %v", block.ID.String(), err)
	}
	err = son.node.Save(&stormBlock{
		DataID: obj.DataID, // automatically incremented in previous save call

		ParentID:  block.ParentID,
		Target:    block.Target,
		Height:    block.Height,
		Timestamp: block.Timestamp,
		Payouts:   block.Payouts,

		Transactions: block.Transactions,
	})
	if err != nil {
		return fmt.Errorf("failed to save block %s by (object) data ID %d: %v", block.ID.String(), obj.DataID, err)
	}
	return nil
}

func (son *stormObjectNode) SaveTransaction(txn *Transaction) error {
	obj := stormObject{
		ObjectID:   ObjectID(txn.ID[:]),
		ObjectType: ObjectTypeTransaction,
		DataID:     son.NextDataID(),
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
		ObjectID:   ObjectID(output.ID[:]),
		ObjectType: ObjectTypeOutput,
		DataID:     son.NextDataID(),
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

func (son *stormObjectNode) SaveSingleSignatureWallet(wallet *SingleSignatureWalletData) error {
	obj := stormObject{
		ObjectID:   ObjectIDFromUnlockHash(wallet.UnlockHash),
		ObjectType: ObjectTypeSingleSignatureWallet,
		DataID:     son.NextDataID(),
	}
	err := son.node.Save(&obj)
	if err != nil {
		return fmt.Errorf("failed to save object type info for wallet %s: %v", wallet.UnlockHash.String(), err)
	}
	err = son.node.Save(&stormSingleSignatureWalletData{
		DataID: obj.DataID, // automatically incremented in previous save call

		PublicKey: wallet.PublicKey,

		CoinOutputs:       wallet.CoinOutputs,
		BlockStakeOutputs: wallet.BlockStakeOutputs,
		Transactions:      wallet.Transactions,

		CoinBalance:       wallet.CoinBalance,
		BlockStakeBalance: wallet.BlockStakeBalance,
	})
	if err != nil {
		return fmt.Errorf("failed to save wallet %s by (object) data ID %d: %v", wallet.UnlockHash.String(), obj.DataID, err)
	}
	return nil
}

func (son *stormObjectNode) SaveMultiSignatureWallet(wallet *MultiSignatureWalletData) error {
	obj := stormObject{
		ObjectID:   ObjectIDFromUnlockHash(wallet.UnlockHash),
		ObjectType: ObjectTypeMultiSignatureWallet,
		DataID:     son.NextDataID(),
	}
	err := son.node.Save(&obj)
	if err != nil {
		return fmt.Errorf("failed to save object type info for wallet %s: %v", wallet.UnlockHash.String(), err)
	}
	err = son.node.Save(&stormMultiSignatureWalletData{
		DataID: obj.DataID, // automatically incremented in previous save call

		Owners:                wallet.Owners,
		RequiredSgnatureCount: wallet.RequiredSgnatureCount,

		CoinOutputs:       wallet.CoinOutputs,
		BlockStakeOutputs: wallet.BlockStakeOutputs,
		Transactions:      wallet.Transactions,

		CoinBalance:       wallet.CoinBalance,
		BlockStakeBalance: wallet.BlockStakeBalance,
	})
	if err != nil {
		return fmt.Errorf("failed to save wallet %s by (object) data ID %d: %v", wallet.UnlockHash.String(), obj.DataID, err)
	}
	return nil
}

func (son *stormObjectNode) SaveAtomicSwapContract(contract *AtomicSwapContract) error {
	obj := stormObject{
		ObjectID:   ObjectIDFromUnlockHash(contract.UnlockHash),
		ObjectType: ObjectTypeAtomicSwapContract,
		DataID:     son.NextDataID(),
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
		ID:   objectID,
		Type: object.ObjectType,
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
		obj.Data, err = son.getTransactionByDataID(types.TransactionID(txnHash), object.DataID)
	case ObjectTypeOutput:
		outputHash, err := objectID.AsHash()
		if err != nil {
			return Object{}, err
		}
		obj.Data, err = son.getOutputByDataID(types.OutputID(outputHash), object.DataID)
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

func (son *stormObjectNode) DeleteBlock(blockID types.BlockID) error {
	return son.deleteObject(ObjectID(blockID[:]), new(stormBlock))
}

func (son *stormObjectNode) DeleteTransaction(txnID types.TransactionID) error {
	return son.deleteObject(ObjectID(txnID[:]), new(stormTransaction))
}

func (son *stormObjectNode) DeleteOutput(outputID types.OutputID) error {
	return son.deleteObject(ObjectID(outputID[:]), new(stormOutput))
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
