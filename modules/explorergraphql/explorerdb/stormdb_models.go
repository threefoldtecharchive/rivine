package explorerdb

import (
	"fmt"
	"io"

	"github.com/asdine/storm"
	"github.com/asdine/storm/q"

	"github.com/threefoldtech/rivine/pkg/encoding/rivbin"
	"github.com/threefoldtech/rivine/types"
)

// TODO: link DataID instead of ObjectID

// Bucket models
type (
	StormChainContext struct {
		ConsensusChangeID StormHash
		Height            types.BlockHeight
		Timestamp         types.Timestamp
		BlockID           StormHash
	}

	StormBlockFactsContext struct {
		Target    StormHash
		Timestamp types.Timestamp
	}

	StormChainAggregatedFacts struct {
		TotalCoins       StormBigInt
		TotalLockedCoins StormBigInt

		TotalBlockStakes           StormBigInt
		TotalLockedBlockStakes     StormBigInt
		EstimatedActiveBlockStakes StormBigInt

		LastBlocks []StormBlockFactsContext
	}
)

// Node Models
type (
	StormDataID uint64

	StormObject struct {
		ObjectID      ObjectID    `storm:"id", msgpack:"id"`
		ObjectType    ObjectType  `storm:"index", msgpack:"ot"`
		ObjectVersion ByteVersion `storm:"index", msgpack:"ov"`
		DataID        StormDataID `storm:"unique", msgpack:"did"`
	}

	StormBlock struct {
		DataID StormDataID `storm:"id", msgpack:"id"`

		ParentID  StormHash         `msgpack:"pid"`
		Height    types.BlockHeight `msgpack:"hg"`
		Timestamp types.Timestamp   `msgpack:"ts"`
		Payouts   []StormHash       `msgpack:"pas"`

		Transactions []StormHash `msgpack:"txs"`
	}

	StormBlockFacts struct {
		DataID StormDataID `storm:"id", msgpack:"id"`

		Target     StormHash   `storm:"index", msgpack:"bt"`
		Difficulty StormBigInt `storm:"index", msgpack:"bd"`

		AggregatedTotalCoins                 StormBigInt `msgpack:"acou"`
		AggregatedTotalLockedCoins           StormBigInt `msgpack:"acol"`
		AggregatedTotalBlockStakes           StormBigInt `msgpack:"absu"`
		AggregatedTotalLockedBlockStakes     StormBigInt `msgpack:"absl"`
		AggregatedEstimatedActiveBlockStakes StormBigInt `msgpack:"aeabs"`
	}

	StormTransactionFeePayoutInfo struct {
		PayoutID StormHash     `msgpack:"paid"`
		Values   []StormBigInt `msgpack:"vls"`
	}

	StormTransaction struct {
		DataID StormDataID `storm:"id", msgpack:"id"`

		ParentBlock StormHash   `msgpack:"pb"`
		Version     ByteVersion `storm:"index", msgpack:"txv"`

		CoinInputs  []StormHash `msgpack:"cis"`
		CoinOutputs []StormHash `msgpack:"cos"`

		BlockStakeInputs  []StormHash `msgpack:"bsis"`
		BlockStakeOutputs []StormHash `msgpack:"bsos"`

		FeePayout StormTransactionFeePayoutInfo `msgpack:"fp"`

		ArbitraryData        []byte `msgpack:"ad"`
		EncodedExtensionData []byte `msgpack:"eed"`
	}

	StormOutputSpenditureData struct {
		Fulfillment              StormUnlockFulfillment `msgpack:"ff"`
		FulfillmentTransactionID StormHash              `msgpack:"fftid"`
	}

	StormOutput struct {
		DataID StormDataID `storm:"id", msgpack:"id"`

		ParentID StormHash `msgpack:"pid"`

		Type                 OutputType           `storm:"index", msgpack:"t"`
		Value                StormBigInt          `msgpack:"v"`
		Condition            StormUnlockCondition `msgpack:"c"`
		UnlockReferencePoint ReferencePoint       `storm:"index", msgpack:"urp"` // required for reverting outputs

		SpenditureData *StormOutputSpenditureData `storm:"index", msgpack:"sd"`
	}

	StormLockedOutput struct {
		DataID               StormDataID    `storm:"id", msgpack:"id"`
		UnlockReferencePoint ReferencePoint `storm:"index", msgpack:"urp"`
	}

	StormBaseWalletData struct {
		DataID StormDataID `storm:"id", msgpack:"id"`

		CoinOutputs       []StormHash `msgpack:"cos"`
		BlockStakeOutputs []StormHash `msgpack:"bos"`
		Blocks            []StormHash `msgpack:"bls"`
		Transactions      []StormHash `msgpack:"txs"`

		CoinsUnlocked StormBigInt `storm:"index", msgpack:"cou"`
		CoinsLocked   StormBigInt `storm:"index", msgpack:"col"`

		BlockStakesUnlocked StormBigInt `storm:"index", msgpack:"bsou"`
		BlockStakesLocked   StormBigInt `storm:"index", msgpack:"bsol"`
	}

	StormFreeForAllWalletData struct {
		StormBaseWalletData `storm:"inline"`
	}

	StormSingleSignatureWalletData struct {
		StormBaseWalletData `storm:"inline"`

		MultiSignatureWallets []StormUnlockHash `msgpack:"msw"`
	}

	StormMultiSignatureWalletData struct {
		StormBaseWalletData `storm:"inline"`

		Owners                 []StormUnlockHash `msgpack:"ow"`
		RequiredSignatureCount int               `msgpack:"rsc"`
	}

	StormAtomicSwapContractSpenditureData struct {
		ContractFulfillment StormAtomicSwapFulfillment `msgpack:"coff"`
		CoinOutput          StormHash                  `msgpack:"co"`
	}

	StormAtomicSwapContract struct {
		DataID StormDataID `storm:"id", msgpack:"id"`

		ContractValue     StormBigInt              `msgpack:"v"`
		ContractCondition StormAtomicSwapCondition `msgpack:"c"`
		Transactions      []StormHash              `msgpack:"txs"`
		CoinInput         StormHash                `msgpack:"ci"`

		SpenditureData *StormAtomicSwapContractSpenditureData `storm:"index", msgpack:"sd"`
	}
)

func (sblock *StormBlock) AsBlock(blockID types.BlockID) Block {
	return Block{
		ID:        blockID,
		ParentID:  sblock.ParentID.AsBlockID(),
		Height:    sblock.Height,
		Timestamp: sblock.Timestamp,
		Payouts:   StormHashSliceAsOutputIDSlice(sblock.Payouts),

		Transactions: StormHashSliceAsTransactionIDSlice(sblock.Transactions),
	}
}

func (sbfacts *StormBlockFacts) AsBlockFacts() BlockFacts {
	return BlockFacts{
		Constants: BlockFactsConstants{
			Target:     sbfacts.Target.AsTarget(),
			Difficulty: sbfacts.Difficulty.AsDifficulty(),
		},
		Aggregated: BlockFactsAggregated{
			TotalCoins:                 sbfacts.AggregatedTotalCoins.AsCurrency(),
			TotalLockedCoins:           sbfacts.AggregatedTotalLockedCoins.AsCurrency(),
			TotalBlockStakes:           sbfacts.AggregatedTotalBlockStakes.AsCurrency(),
			TotalLockedBlockStakes:     sbfacts.AggregatedTotalLockedBlockStakes.AsCurrency(),
			EstimatedActiveBlockStakes: sbfacts.AggregatedEstimatedActiveBlockStakes.AsCurrency(),
		},
	}
}

func (stxn *StormTransaction) AsTransaction(txnID types.TransactionID) Transaction {
	return Transaction{
		ID: txnID,

		ParentBlock: stxn.ParentBlock.AsBlockID(),
		Version:     types.TransactionVersion(stxn.Version),

		CoinInputs:  StormHashSliceAsOutputIDSlice(stxn.CoinInputs),
		CoinOutputs: StormHashSliceAsOutputIDSlice(stxn.CoinOutputs),

		BlockStakeInputs:  StormHashSliceAsOutputIDSlice(stxn.BlockStakeInputs),
		BlockStakeOutputs: StormHashSliceAsOutputIDSlice(stxn.BlockStakeOutputs),

		FeePayout: TransactionFeePayoutInfo{
			PayoutID: stxn.FeePayout.PayoutID.AsOutputID(),
			Values:   StormBigIntSliceAsCurrencies(stxn.FeePayout.Values),
		},

		ArbitraryData:        stxn.ArbitraryData,
		EncodedExtensionData: stxn.EncodedExtensionData,
	}
}

func (sout *StormOutput) AsOutput(outputID types.OutputID) Output {
	out := Output{
		ID:       outputID,
		ParentID: sout.ParentID.AsCryptoHash(),

		Type:                 sout.Type,
		Value:                sout.Value.AsCurrency(),
		Condition:            sout.Condition.AsUnlockConditionProxy(),
		UnlockReferencePoint: sout.UnlockReferencePoint,
	}
	if sout.SpenditureData != nil {
		out.SpenditureData = &OutputSpenditureData{
			Fulfillment:              sout.SpenditureData.Fulfillment.AsUnlockFulfillmentProxy(),
			FulfillmentTransactionID: sout.SpenditureData.FulfillmentTransactionID.AsTransactionID(),
		}
	}
	return out
}

func walletDataAsSDB(dataID StormDataID, wallet *WalletData) *StormBaseWalletData {
	return &StormBaseWalletData{
		DataID: dataID,

		CoinOutputs:       OutputIDSliceAsStormHashSlice(wallet.CoinOutputs),
		BlockStakeOutputs: OutputIDSliceAsStormHashSlice(wallet.BlockStakeOutputs),
		Blocks:            BlockIDSliceAsStormHashSlice(wallet.Blocks),
		Transactions:      TransactionIDSliceAsStormHashSlice(wallet.Transactions),

		CoinsUnlocked: StormBigIntFromCurrency(wallet.CoinBalance.Unlocked),
		CoinsLocked:   StormBigIntFromCurrency(wallet.CoinBalance.Locked),

		BlockStakesUnlocked: StormBigIntFromCurrency(wallet.BlockStakeBalance.Unlocked),
		BlockStakesLocked:   StormBigIntFromCurrency(wallet.BlockStakeBalance.Locked),
	}
}

func (swallet *StormBaseWalletData) AsWalletData(uh types.UnlockHash) WalletData {
	return WalletData{
		UnlockHash: uh,

		CoinOutputs:       StormHashSliceAsOutputIDSlice(swallet.CoinOutputs),
		BlockStakeOutputs: StormHashSliceAsOutputIDSlice(swallet.BlockStakeOutputs),
		Blocks:            StormHashSliceAsBlockIDSlice(swallet.Blocks),
		Transactions:      StormHashSliceAsTransactionIDSlice(swallet.Transactions),

		CoinBalance: Balance{
			Unlocked: swallet.CoinsUnlocked.AsCurrency(),
			Locked:   swallet.CoinsLocked.AsCurrency(),
		},
		BlockStakeBalance: Balance{
			Unlocked: swallet.BlockStakesUnlocked.AsCurrency(),
			Locked:   swallet.BlockStakesLocked.AsCurrency(),
		},
	}
}

func (swallet *StormFreeForAllWalletData) AsFreeForAllWallet(uh types.UnlockHash) FreeForAllWalletData {
	return FreeForAllWalletData{
		WalletData: swallet.AsWalletData(uh),
	}
}

// Required to write these functions ourselves,
// as Rivine Encoding ignores anonymous fields (embedded fields are anonymous).
// Not sure why Rivine Encoding (inspired by Sia Encoding) choose to ignore this,
// but it does, that is what matters.
func (swallet StormFreeForAllWalletData) MarshalRivine(w io.Writer) error {
	return rivbin.NewEncoder(w).Encode(swallet.StormBaseWalletData)
}
func (swallet *StormFreeForAllWalletData) UnmarshalRivine(r io.Reader) error {
	return rivbin.NewDecoder(r).Decode(&swallet.StormBaseWalletData)
}

func (swallet *StormSingleSignatureWalletData) AsSingleSignatureWallet(uh types.UnlockHash) SingleSignatureWalletData {
	return SingleSignatureWalletData{
		WalletData:            swallet.AsWalletData(uh),
		MultiSignatureWallets: StormUnlockHashSliceAsUnlockHashSlice(swallet.MultiSignatureWallets),
	}
}

// Required to write these functions ourselves,
// as Rivine Encoding ignores anonymous fields (embedded fields are anonymous).
// Not sure why Rivine Encoding (inspired by Sia Encoding) choose to ignore this,
// but it does, that is what matters.
func (swallet StormSingleSignatureWalletData) MarshalRivine(w io.Writer) error {
	return rivbin.NewEncoder(w).EncodeAll(
		swallet.StormBaseWalletData, swallet.MultiSignatureWallets)
}

func (swallet *StormSingleSignatureWalletData) UnmarshalRivine(r io.Reader) error {
	return rivbin.NewDecoder(r).DecodeAll(
		&swallet.StormBaseWalletData, &swallet.MultiSignatureWallets)
}

func (swallet *StormMultiSignatureWalletData) AsMultiSignatureWallet(uh types.UnlockHash) MultiSignatureWalletData {
	return MultiSignatureWalletData{
		WalletData:            swallet.AsWalletData(uh),
		Owners:                StormUnlockHashSliceAsUnlockHashSlice(swallet.Owners),
		RequiredSgnatureCount: swallet.RequiredSignatureCount,
	}
}

// Required to write these functions ourselves,
// as Rivine Encoding ignores anonymous fields (embedded fields are anonymous).
// Not sure why Rivine Encoding (inspired by Sia Encoding) choose to ignore this,
// but it does, that is what matters.
func (swallet StormMultiSignatureWalletData) MarshalRivine(w io.Writer) error {
	return rivbin.NewEncoder(w).EncodeAll(
		swallet.StormBaseWalletData, swallet.Owners, swallet.RequiredSignatureCount)
}

func (swallet *StormMultiSignatureWalletData) UnmarshalRivine(r io.Reader) error {
	return rivbin.NewDecoder(r).DecodeAll(
		&swallet.StormBaseWalletData, &swallet.Owners, &swallet.RequiredSignatureCount)
}

func (scontract *StormAtomicSwapContract) AsAtomicSwapContract(uh types.UnlockHash) AtomicSwapContract {
	asc := AtomicSwapContract{
		UnlockHash: uh,

		ContractValue:     scontract.ContractValue.AsCurrency(),
		ContractCondition: scontract.ContractCondition.AsAtomicSwapCondition(),
		Transactions:      StormHashSliceAsTransactionIDSlice(scontract.Transactions),
		CoinInput:         scontract.CoinInput.AsCoinOutputID(),
	}
	if scontract.SpenditureData != nil {
		asc.SpenditureData = &AtomicSwapContractSpenditureData{
			ContractFulfillment: scontract.SpenditureData.ContractFulfillment.AsAtomicSwapFulfillment(),
			CoinOutput:          scontract.SpenditureData.CoinOutput.AsCoinOutputID(),
		}
	}
	return asc
}

const (
	nodeObjectKeyObjectID   = "ObjectID"
	nodeObjectKeyDataID     = "DataID"
	nodeObjectKeyUnlockHash = "UnlockHash"

	nodeObjectOutputKeySpenditureData       = "SpenditureData"
	nodeObjectOutputKeyUnlockReferencePoint = "UnlockReferencePoint"
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
	}

	stormObjectNodeReaderWriter interface {
		stormObjectNodeReader

		GetLastDataID() StormDataID

		SaveBlockWithFacts(*Block, *BlockFacts) error
		SaveTransaction(*Transaction) error
		SaveOutput(*Output, types.BlockHeight, types.Timestamp) error
		SaveFreeForAllWallet(*FreeForAllWalletData) error
		SaveSingleSignatureWallet(*SingleSignatureWalletData) error
		SaveMultiSignatureWallet(*MultiSignatureWalletData) error
		SaveAtomicSwapContract(*AtomicSwapContract) error

		UpdateOutputSpenditureData(types.OutputID, *OutputSpenditureData) (Output, error)
		UnlockLockedOutputs(types.BlockHeight, types.Timestamp, types.Timestamp) ([]StormOutput, error)
		RelockLockedOutputs(types.BlockHeight, types.Timestamp, types.Timestamp) ([]StormOutput, error)

		DeleteBlock(types.BlockID) (Block, error)
		DeleteTransaction(types.TransactionID) (Transaction, error)
		DeleteOutput(types.OutputID, types.BlockHeight, types.Timestamp) (Output, error)
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
	lastDataID StormDataID
}

func newStormObjectNodeReader(db *StormDB) stormObjectNodeReader {
	return &stormObjectNode{
		node: db.rootNode(nodeNameObjects),
	}
}

func newStormObjectNodeReaderWriter(db *StormDB, lastDataID StormDataID) stormObjectNodeReaderWriter {
	return &stormObjectNode{
		node:       db.rootNode(nodeNameObjects),
		lastDataID: lastDataID,
	}
}

func (son *stormObjectNode) nextDataID() (dataID StormDataID) {
	son.lastDataID++
	dataID = son.lastDataID
	return
}

func (son *stormObjectNode) GetLastDataID() StormDataID {
	return son.lastDataID
}

func (son *stormObjectNode) SaveBlockWithFacts(block *Block, facts *BlockFacts) error {
	obj := StormObject{
		ObjectID:      ObjectID(block.ID[:]),
		ObjectType:    ObjectTypeBlock,
		ObjectVersion: 0, // there is only one version of blocks
		DataID:        son.nextDataID(),
	}
	err := son.node.Save(&obj)
	if err != nil {
		return fmt.Errorf("failed to save object type info for block %s: %v", block.ID.String(), err)
	}
	err = son.node.Save(&StormBlock{
		DataID: obj.DataID, // automatically incremented in previous save call

		ParentID:  StormHashFromBlockID(block.ParentID),
		Height:    block.Height,
		Timestamp: block.Timestamp,
		Payouts:   OutputIDSliceAsStormHashSlice(block.Payouts),

		Transactions: TransactionIDSliceAsStormHashSlice(block.Transactions),
	})
	if err != nil {
		return fmt.Errorf("failed to save block %s by (object) data ID %d: %v", block.ID.String(), obj.DataID, err)
	}
	err = son.node.Save(&StormBlockFacts{
		DataID: obj.DataID, // automatically incremented in previous save call

		Target:     StormHashFromTarget(facts.Constants.Target),
		Difficulty: StormBigIntFromDifficulty(facts.Constants.Difficulty),

		AggregatedTotalCoins:                 StormBigIntFromCurrency(facts.Aggregated.TotalCoins),
		AggregatedTotalLockedCoins:           StormBigIntFromCurrency(facts.Aggregated.TotalLockedCoins),
		AggregatedTotalBlockStakes:           StormBigIntFromCurrency(facts.Aggregated.TotalBlockStakes),
		AggregatedTotalLockedBlockStakes:     StormBigIntFromCurrency(facts.Aggregated.TotalLockedBlockStakes),
		AggregatedEstimatedActiveBlockStakes: StormBigIntFromCurrency(facts.Aggregated.EstimatedActiveBlockStakes),
	})
	if err != nil {
		return fmt.Errorf("failed to save facts for block %s by (object) data ID %d: %v", block.ID.String(), obj.DataID, err)
	}
	return nil
}

func (son *stormObjectNode) SaveTransaction(txn *Transaction) error {
	obj := StormObject{
		ObjectID:      ObjectID(txn.ID[:]),
		ObjectType:    ObjectTypeTransaction,
		ObjectVersion: ByteVersion(txn.Version),
		DataID:        son.nextDataID(),
	}
	err := son.node.Save(&obj)
	if err != nil {
		return fmt.Errorf("failed to save object type info for transaction %s: %v", txn.ID.String(), err)
	}
	err = son.node.Save(&StormTransaction{
		DataID: obj.DataID, // automatically incremented in previous save call

		ParentBlock: StormHashFromBlockID(txn.ParentBlock),
		Version:     ByteVersion(txn.Version),

		CoinInputs:  OutputIDSliceAsStormHashSlice(txn.CoinInputs),
		CoinOutputs: OutputIDSliceAsStormHashSlice(txn.CoinOutputs),

		BlockStakeInputs:  OutputIDSliceAsStormHashSlice(txn.BlockStakeInputs),
		BlockStakeOutputs: OutputIDSliceAsStormHashSlice(txn.BlockStakeOutputs),

		FeePayout: StormTransactionFeePayoutInfo{
			PayoutID: StormHashFromOutputID(txn.FeePayout.PayoutID),
			Values:   CurrencySliceAsStormBigIntSlice(txn.FeePayout.Values),
		},

		ArbitraryData:        txn.ArbitraryData,
		EncodedExtensionData: txn.EncodedExtensionData,
	})
	if err != nil {
		return fmt.Errorf("failed to save transaction %s by (object) data ID %d: %v", txn.ID.String(), obj.DataID, err)
	}
	return nil
}

func (son *stormObjectNode) SaveOutput(output *Output, height types.BlockHeight, timestamp types.Timestamp) error {
	obj := StormObject{
		ObjectID:      ObjectID(output.ID[:]),
		ObjectType:    ObjectTypeOutput,
		ObjectVersion: 0, // there is only one version of outputs
		DataID:        son.nextDataID(),
	}
	err := son.node.Save(&obj)
	if err != nil {
		return fmt.Errorf("failed to save object type info for output %s: %v", output.ID.String(), err)
	}
	sout := StormOutput{
		DataID: obj.DataID, // automatically incremented in previous save call

		ParentID: StormHashFromHash(output.ParentID),

		Type:                 output.Type,
		Value:                StormBigIntFromCurrency(output.Value),
		Condition:            StormUnlockConditionFromUnlockCondition(output.Condition),
		UnlockReferencePoint: output.UnlockReferencePoint,
	}
	if output.SpenditureData != nil {
		sout.SpenditureData = &StormOutputSpenditureData{
			Fulfillment:              StormUnlockFulfillmentFromUnlockFulfillment(output.SpenditureData.Fulfillment),
			FulfillmentTransactionID: StormHashFromTransactionID(output.SpenditureData.FulfillmentTransactionID),
		}
	}
	err = son.node.Save(&sout)
	if err != nil {
		return fmt.Errorf("failed to save output %s by (object) data ID %d: %v", output.ID.String(), obj.DataID, err)
	}
	if !output.UnlockReferencePoint.Reached(height, timestamp) {
		err = son.node.Save(&StormLockedOutput{
			DataID:               obj.DataID, // automatically incremented in previous save call
			UnlockReferencePoint: output.UnlockReferencePoint,
		})
		if err != nil {
			return fmt.Errorf("failed to save output %s by (object) data ID %d as locked: %v", output.ID.String(), obj.DataID, err)
		}
	}
	return nil
}

func (son *stormObjectNode) SaveFreeForAllWallet(wallet *FreeForAllWalletData) error {
	obj := StormObject{
		ObjectID:      ObjectIDFromUnlockHash(wallet.UnlockHash),
		ObjectType:    ObjectTypeSingleSignatureWallet,
		ObjectVersion: ByteVersion(wallet.UnlockHash.Type),
		DataID:        son.nextDataID(),
	}
	err := son.node.Save(&obj)
	if err != nil {
		return fmt.Errorf("failed to save object type info for wallet %s: %v", wallet.UnlockHash.String(), err)
	}
	err = son.node.Save(&StormFreeForAllWalletData{
		// DataID was automatically incremented in previous save call
		StormBaseWalletData: *walletDataAsSDB(obj.DataID, &wallet.WalletData),
	})
	if err != nil {
		return fmt.Errorf("failed to save wallet %s by (object) data ID %d: %v", wallet.UnlockHash.String(), obj.DataID, err)
	}
	return nil
}

func (son *stormObjectNode) SaveSingleSignatureWallet(wallet *SingleSignatureWalletData) error {
	obj := StormObject{
		ObjectID:      ObjectIDFromUnlockHash(wallet.UnlockHash),
		ObjectType:    ObjectTypeSingleSignatureWallet,
		ObjectVersion: ByteVersion(wallet.UnlockHash.Type),
		DataID:        son.nextDataID(),
	}
	err := son.node.Save(&obj)
	if err != nil {
		return fmt.Errorf("failed to save object type info for wallet %s: %v", wallet.UnlockHash.String(), err)
	}

	err = son.node.Save(&StormSingleSignatureWalletData{
		// DataID was automatically incremented in previous save call
		StormBaseWalletData:   *walletDataAsSDB(obj.DataID, &wallet.WalletData),
		MultiSignatureWallets: UnlockHashSliceAsStormUnlockHashSlice(wallet.MultiSignatureWallets),
	})
	if err != nil {
		return fmt.Errorf("failed to save wallet %s by (object) data ID %d: %v", wallet.UnlockHash.String(), obj.DataID, err)
	}
	return nil
}

func (son *stormObjectNode) SaveMultiSignatureWallet(wallet *MultiSignatureWalletData) error {
	obj := StormObject{
		ObjectID:      ObjectIDFromUnlockHash(wallet.UnlockHash),
		ObjectType:    ObjectTypeMultiSignatureWallet,
		ObjectVersion: ByteVersion(wallet.UnlockHash.Type),
		DataID:        son.nextDataID(),
	}
	err := son.node.Save(&obj)
	if err != nil {
		return fmt.Errorf("failed to save object type info for wallet %s: %v", wallet.UnlockHash.String(), err)
	}
	err = son.node.Save(&StormMultiSignatureWalletData{
		// DataID was automatically incremented in previous save call
		StormBaseWalletData:    *walletDataAsSDB(obj.DataID, &wallet.WalletData),
		Owners:                 UnlockHashSliceAsStormUnlockHashSlice(wallet.Owners),
		RequiredSignatureCount: wallet.RequiredSgnatureCount,
	})
	if err != nil {
		return fmt.Errorf("failed to save wallet %s by (object) data ID %d: %v", wallet.UnlockHash.String(), obj.DataID, err)
	}
	return nil
}

func (son *stormObjectNode) SaveAtomicSwapContract(contract *AtomicSwapContract) error {
	obj := StormObject{
		ObjectID:      ObjectIDFromUnlockHash(contract.UnlockHash),
		ObjectType:    ObjectTypeAtomicSwapContract,
		ObjectVersion: ByteVersion(contract.UnlockHash.Type),
		DataID:        son.nextDataID(),
	}
	err := son.node.Save(&obj)
	if err != nil {
		return fmt.Errorf("failed to save object type info for contract %s: %v", contract.UnlockHash.String(), err)
	}
	scontract := StormAtomicSwapContract{
		DataID: obj.DataID, // automatically incremented in previous save call

		ContractValue:     StormBigIntFromCurrency(contract.ContractValue),
		ContractCondition: StormAtomicSwapConditionFromAtomicSwapCondition(contract.ContractCondition),
		Transactions:      TransactionIDSliceAsStormHashSlice(contract.Transactions),
		CoinInput:         StormHashFromCoinOutputID(contract.CoinInput),
	}
	if contract.SpenditureData != nil {
		scontract.SpenditureData = &StormAtomicSwapContractSpenditureData{
			ContractFulfillment: StormAtomicSwapFulfillmentFromAtomicSwapFulfillment(contract.SpenditureData.ContractFulfillment),
			CoinOutput:          StormHashFromCoinOutputID(contract.SpenditureData.CoinOutput),
		}
	}
	err = son.node.Save(&scontract)
	if err != nil {
		return fmt.Errorf("failed to save contract %s by (object) data ID %d: %v", contract.UnlockHash.String(), obj.DataID, err)
	}
	return nil
}

func (son *stormObjectNode) GetObject(objectID ObjectID) (Object, error) {
	var object StormObject
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
	var object StormObject
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
	var object StormObject
	err := son.node.One(nodeObjectKeyObjectID, objectID, &object)
	if err != nil {
		return err
	}
	return son.getTypedObjectByDataID(object.DataID, value)
}

func (son *stormObjectNode) getTypedObjectByDataID(dataID StormDataID, value interface{}) error {
	return son.node.One(nodeObjectKeyDataID, dataID, value)
}

func (son *stormObjectNode) GetBlock(blockID types.BlockID) (Block, error) {
	var sblock StormBlock
	err := son.getTypedObject(ObjectID(blockID[:]), &sblock)
	if err != nil {
		return Block{}, err
	}
	return sblock.AsBlock(blockID), nil
}

func (son *stormObjectNode) GetBlockFacts(blockID types.BlockID) (BlockFacts, error) {
	var sbfacts StormBlockFacts
	err := son.getTypedObject(ObjectID(blockID[:]), &sbfacts)
	if err != nil {
		return BlockFacts{}, err
	}
	return sbfacts.AsBlockFacts(), nil
}

func (son *stormObjectNode) getBlockByDataID(blockID types.BlockID, dataID StormDataID) (Block, error) {
	var sblock StormBlock
	err := son.getTypedObjectByDataID(dataID, &sblock)
	if err != nil {
		return Block{}, err
	}
	return sblock.AsBlock(blockID), nil
}

func (son *stormObjectNode) GetTransaction(transactionID types.TransactionID) (Transaction, error) {
	var stxn StormTransaction
	err := son.getTypedObject(ObjectID(transactionID[:]), &stxn)
	if err != nil {
		return Transaction{}, err
	}
	return stxn.AsTransaction(transactionID), nil
}

func (son *stormObjectNode) getTransactionByDataID(transactionID types.TransactionID, dataID StormDataID) (Transaction, error) {
	var stxn StormTransaction
	err := son.getTypedObjectByDataID(dataID, &stxn)
	if err != nil {
		return Transaction{}, err
	}
	return stxn.AsTransaction(transactionID), nil
}

func (son *stormObjectNode) GetOutput(outputID types.OutputID) (Output, error) {
	var sout StormOutput
	err := son.getTypedObject(ObjectID(outputID[:]), &sout)
	if err != nil {
		return Output{}, err
	}
	return sout.AsOutput(outputID), nil
}

func (son *stormObjectNode) getOutputByDataID(outputID types.OutputID, dataID StormDataID) (Output, error) {
	var sout StormOutput
	err := son.getTypedObjectByDataID(dataID, &sout)
	if err != nil {
		return Output{}, err
	}
	return sout.AsOutput(outputID), nil
}

func (son *stormObjectNode) GetFreeForAllWallet(uh types.UnlockHash) (FreeForAllWalletData, error) {
	var swallet StormFreeForAllWalletData
	err := son.getTypedObject(ObjectIDFromUnlockHash(uh), &swallet)
	if err != nil {
		return FreeForAllWalletData{}, err
	}
	return swallet.AsFreeForAllWallet(uh), nil
}

func (son *stormObjectNode) getFreeForAllWalletByDataID(uh types.UnlockHash, dataID StormDataID) (FreeForAllWalletData, error) {
	var swallet StormFreeForAllWalletData
	err := son.getTypedObjectByDataID(dataID, &swallet)
	if err != nil {
		return FreeForAllWalletData{}, err
	}
	return swallet.AsFreeForAllWallet(uh), nil
}

func (son *stormObjectNode) GetSingleSignatureWallet(uh types.UnlockHash) (SingleSignatureWalletData, error) {
	var swallet StormSingleSignatureWalletData
	err := son.getTypedObject(ObjectIDFromUnlockHash(uh), &swallet)
	if err != nil {
		return SingleSignatureWalletData{}, err
	}
	return swallet.AsSingleSignatureWallet(uh), nil
}

func (son *stormObjectNode) getSingleSignatureWalletByDataID(uh types.UnlockHash, dataID StormDataID) (SingleSignatureWalletData, error) {
	var swallet StormSingleSignatureWalletData
	err := son.getTypedObjectByDataID(dataID, &swallet)
	if err != nil {
		return SingleSignatureWalletData{}, err
	}
	return swallet.AsSingleSignatureWallet(uh), nil
}

func (son *stormObjectNode) GetMultiSignatureWallet(uh types.UnlockHash) (MultiSignatureWalletData, error) {
	var swallet StormMultiSignatureWalletData
	err := son.getTypedObject(ObjectIDFromUnlockHash(uh), &swallet)
	if err != nil {
		return MultiSignatureWalletData{}, err
	}
	return swallet.AsMultiSignatureWallet(uh), nil
}

func (son *stormObjectNode) getMultiSignatureWalletByDataID(uh types.UnlockHash, dataID StormDataID) (MultiSignatureWalletData, error) {
	var swallet StormMultiSignatureWalletData
	err := son.getTypedObjectByDataID(dataID, &swallet)
	if err != nil {
		return MultiSignatureWalletData{}, err
	}
	return swallet.AsMultiSignatureWallet(uh), nil
}

func (son *stormObjectNode) GetAtomicSwapContract(uh types.UnlockHash) (AtomicSwapContract, error) {
	var scontract StormAtomicSwapContract
	err := son.getTypedObject(ObjectIDFromUnlockHash(uh), &scontract)
	if err != nil {
		return AtomicSwapContract{}, err
	}
	return scontract.AsAtomicSwapContract(uh), nil
}

func (son *stormObjectNode) getAtomicSwapContractByDataID(uh types.UnlockHash, dataID StormDataID) (AtomicSwapContract, error) {
	var scontract StormAtomicSwapContract
	err := son.getTypedObjectByDataID(dataID, &scontract)
	if err != nil {
		return AtomicSwapContract{}, err
	}
	return scontract.AsAtomicSwapContract(uh), nil
}

func (son *stormObjectNode) UpdateOutputSpenditureData(outputID types.OutputID, spenditureData *OutputSpenditureData) (Output, error) {
	var object StormObject
	err := son.node.One(nodeObjectKeyObjectID, ObjectID(outputID[:]), &object)
	if err != nil {
		return Output{}, err
	}
	var output StormOutput
	err = son.node.One(nodeObjectKeyDataID, object.DataID, &output)
	if err != nil {
		return Output{}, err
	}
	if spenditureData == nil {
		output.SpenditureData = nil
	} else {
		output.SpenditureData = &StormOutputSpenditureData{
			Fulfillment:              StormUnlockFulfillmentFromUnlockFulfillment(spenditureData.Fulfillment),
			FulfillmentTransactionID: StormHashFromTransactionID(spenditureData.FulfillmentTransactionID),
		}
	}
	err = son.node.Update(&output)
	if err != nil {
		return Output{}, err
	}
	return output.AsOutput(outputID), nil
}

func (son *stormObjectNode) UnlockLockedOutputs(height types.BlockHeight, minTimestamp, maxInclusiveTimestamp types.Timestamp) ([]StormOutput, error) {
	// fetch all unlocked outputs
	var lockedOutputs []StormLockedOutput
	err := son.node.Select(q.Or(
		q.Eq(nodeObjectOutputKeyUnlockReferencePoint, ReferencePoint(height)),
		q.And(
			q.Gt(nodeObjectOutputKeyUnlockReferencePoint, ReferencePoint(minTimestamp)),
			q.Lte(nodeObjectOutputKeyUnlockReferencePoint, ReferencePoint(maxInclusiveTimestamp)),
		))).Find(&lockedOutputs)
	if err != nil {
		return nil, err
	}
	if len(lockedOutputs) == 0 {
		return nil, nil // nothing to do
	}
	// gather the data identifiers of the locked outputs
	dataIdentifiers := make([]StormDataID, 0, len(lockedOutputs))
	for _, lo := range lockedOutputs {
		dataIdentifiers = append(dataIdentifiers, lo.DataID)
		// delete unlocked output as well
		err = son.node.DeleteStruct(&lo)
		if err != nil {
			return nil, fmt.Errorf("failed to delete unlocked (previously) locked output by data ID %d: %v", lo.DataID, err)
		}
	}
	// fetch all unlocked storm outputs, so callee is aware of their details
	var stormOutputs []StormOutput
	err = son.node.Select(q.In(nodeObjectKeyDataID, dataIdentifiers)).Find(&stormOutputs)
	return stormOutputs, err
}

func (son *stormObjectNode) RelockLockedOutputs(height types.BlockHeight, minTimestamp, maxInclusiveTimestamp types.Timestamp) ([]StormOutput, error) {
	// fetch all unlocked outputs
	var unlockedOutputs []StormOutput
	err := son.node.Select(q.Or(
		q.Eq(nodeObjectOutputKeyUnlockReferencePoint, ReferencePoint(height)),
		q.And(
			q.Gt(nodeObjectOutputKeyUnlockReferencePoint, ReferencePoint(minTimestamp)),
			q.Lte(nodeObjectOutputKeyUnlockReferencePoint, ReferencePoint(maxInclusiveTimestamp)),
		))).Find(&unlockedOutputs)
	if err != nil {
		return nil, err
	}
	if len(unlockedOutputs) == 0 {
		return nil, nil // nothing to do
	}
	// gather the data identifiers of the locked outputs
	dataIdentifiers := make([]StormDataID, 0, len(unlockedOutputs))
	for _, uo := range unlockedOutputs {
		dataIdentifiers = append(dataIdentifiers, uo.DataID)
		// revert unlocked output as well (locking it again)
		err = son.node.Save(&StormLockedOutput{
			DataID:               uo.DataID, // automatically incremented in previous save call
			UnlockReferencePoint: uo.UnlockReferencePoint,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to save output by (object) data ID %d as locked: %v", uo.DataID, err)
		}
	}
	return unlockedOutputs, err
}

func (son *stormObjectNode) DeleteBlock(blockID types.BlockID) (Block, error) {
	var block StormBlock
	err := son.deleteObject(ObjectID(blockID[:]), &block)
	if err != nil {
		return Block{}, err
	}
	return block.AsBlock(blockID), nil
}

func (son *stormObjectNode) DeleteTransaction(txnID types.TransactionID) (Transaction, error) {
	var txn StormTransaction
	err := son.deleteObject(ObjectID(txnID[:]), &txn)
	if err != nil {
		return Transaction{}, err
	}
	return txn.AsTransaction(txnID), nil
}

func (son *stormObjectNode) DeleteOutput(outputID types.OutputID, height types.BlockHeight, timestamp types.Timestamp) (Output, error) {
	var output StormOutput
	err := son.deleteObject(ObjectID(outputID[:]), &output)
	if err != nil {
		return Output{}, err
	}
	if !output.UnlockReferencePoint.Reached(height, timestamp) {
		// TODO: log error??
		son.deleteObject(ObjectID(outputID[:]), new(StormLockedOutput))
	}
	return output.AsOutput(outputID), nil
}

func (son *stormObjectNode) DeleteFreeForAllWallet(uh types.UnlockHash) error {
	return son.deleteObject(ObjectIDFromUnlockHash(uh), new(StormFreeForAllWalletData))
}

func (son *stormObjectNode) DeleteSingleSignatureWallet(uh types.UnlockHash) error {
	return son.deleteObject(ObjectIDFromUnlockHash(uh), new(StormSingleSignatureWalletData))
}

func (son *stormObjectNode) DeleteMultiSignatureWallet(uh types.UnlockHash) error {
	return son.deleteObject(ObjectIDFromUnlockHash(uh), new(StormMultiSignatureWalletData))
}

func (son *stormObjectNode) DeleteAtomicSwapContract(uh types.UnlockHash) error {
	return son.deleteObject(ObjectIDFromUnlockHash(uh), new(StormAtomicSwapContract))
}

func (son *stormObjectNode) deleteObject(objectID ObjectID, value interface{}) error {
	// delete object metadata
	var object StormObject
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
