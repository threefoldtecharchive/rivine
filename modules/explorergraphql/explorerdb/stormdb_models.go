package explorerdb

import (
	"fmt"

	"github.com/asdine/storm"
	"github.com/asdine/storm/q"

	"github.com/threefoldtech/rivine/persist"
	"github.com/threefoldtech/rivine/types"
)

// TODO: link DataID instead of ObjectID

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

		ArbitraryData        []byte `storm:"index", msgpack:"ad"`
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

	StormLockedOutputsByHeight struct {
		Height  types.BlockHeight `storm:"id", msgpack:"h"`
		DataIDs []StormDataID     `msgpack:"ids"`
	}

	StormDataIDTimestampPair struct {
		DataID StormDataID           `msgpack:"id"`
		Offset StormTimeBucketOffset `msgpack:"o,omitempty"`
	}
	StormLockedOutputsByBucketTimestamp struct {
		BucketID StormTimeBucketID          `storm:"id", msgpack:"t"`
		Pairs    []StormDataIDTimestampPair `msgpack:"p"`
	}

	StormBaseWalletData struct {
		DataID StormDataID `storm:"id", msgpack:"id"`

		// CoinOutputs       []StormHash `msgpack:"cos"`
		// BlockStakeOutputs []StormHash `msgpack:"bos"`
		// Blocks            []StormHash `msgpack:"bls"`
		// Transactions      []StormHash `msgpack:"txs"`

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

type (
	StormTimeBucketID     uint64
	StormTimeBucketOffset uint8
)

const (
	// needs to fit in an uint8,
	// and want to leave some room for possible extension flags in future
	tsBasedBucketRange = 240
)

func GetTimestampBucketAndOffset(ts types.Timestamp) (StormTimeBucketID, StormTimeBucketOffset) {
	return StormTimeBucketID(ts / tsBasedBucketRange), StormTimeBucketOffset(ts % tsBasedBucketRange)
}

func GetTimestampBucketIdentifiersForTimestampRange(startExclusive, endInclusive types.Timestamp) (buckets []StormTimeBucketID) {
	if startExclusive >= endInclusive {
		panic(fmt.Sprintf(
			"invalid startExclusive %d or endInclusive %d values for getting time-based buckets for locked outputs",
			startExclusive, endInclusive))
	}

	dur := uint64(endInclusive - startExclusive)

	bucket, bucketOffset := GetTimestampBucketAndOffset(startExclusive + 1)
	buckets = append(buckets, bucket)
	interval := uint64(tsBasedBucketRange - bucketOffset)

	for dur > interval {
		dur -= interval
		bucket++
		buckets = append(buckets, bucket)
		interval = tsBasedBucketRange
	}

	return
}

func GetStormDataIDTimestampPairsForDataIDTimestampPair(dataID StormDataID, ts types.Timestamp) (StormTimeBucketID, StormDataIDTimestampPair) {
	bucket, bucketOffset := GetTimestampBucketAndOffset(ts)
	return bucket, StormDataIDTimestampPair{
		DataID: dataID,
		Offset: bucketOffset,
	}
}

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

func StormBlockSliceAsBlockSlice(sblocks []StormBlock, blockIdentifiers []types.BlockID) (blocks []Block) {
	blocks = make([]Block, 0, len(sblocks))
	for idx, block := range sblocks {
		blocks = append(blocks, block.AsBlock(blockIdentifiers[idx]))
	}
	return
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

		// CoinOutputs:       OutputIDSliceAsStormHashSlice(wallet.CoinOutputs),
		// BlockStakeOutputs: OutputIDSliceAsStormHashSlice(wallet.BlockStakeOutputs),
		// Blocks:            BlockIDSliceAsStormHashSlice(wallet.Blocks),
		// Transactions:      TransactionIDSliceAsStormHashSlice(wallet.Transactions),

		CoinsUnlocked: StormBigIntFromCurrency(wallet.CoinBalance.Unlocked),
		CoinsLocked:   StormBigIntFromCurrency(wallet.CoinBalance.Locked),

		BlockStakesUnlocked: StormBigIntFromCurrency(wallet.BlockStakeBalance.Unlocked),
		BlockStakesLocked:   StormBigIntFromCurrency(wallet.BlockStakeBalance.Locked),
	}
}

func (swallet *StormBaseWalletData) AsWalletData(uh types.UnlockHash) WalletData {
	return WalletData{
		UnlockHash: uh,

		// CoinOutputs:       StormHashSliceAsOutputIDSlice(swallet.CoinOutputs),
		// BlockStakeOutputs: StormHashSliceAsOutputIDSlice(swallet.BlockStakeOutputs),
		// Blocks:            StormHashSliceAsBlockIDSlice(swallet.Blocks),
		// Transactions:      StormHashSliceAsTransactionIDSlice(swallet.Transactions),

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

func (swallet *StormSingleSignatureWalletData) AsSingleSignatureWallet(uh types.UnlockHash) SingleSignatureWalletData {
	return SingleSignatureWalletData{
		WalletData:            swallet.AsWalletData(uh),
		MultiSignatureWallets: StormUnlockHashSliceAsUnlockHashSlice(swallet.MultiSignatureWallets),
	}
}

func (swallet *StormMultiSignatureWalletData) AsMultiSignatureWallet(uh types.UnlockHash) MultiSignatureWalletData {
	return MultiSignatureWalletData{
		WalletData:            swallet.AsWalletData(uh),
		Owners:                StormUnlockHashSliceAsUnlockHashSlice(swallet.Owners),
		RequiredSgnatureCount: swallet.RequiredSignatureCount,
	}
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
	nodeObjectOutputKeyHeight               = "Height"
	nodeObjectOutputKeyBucketID             = "BucketID"

	nodeObjectBlockFieldHeight    = "Height"
	nodeObjectBlockFieldTimestamp = "Timestamp"
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

		GetBlocks(int, *BlocksFilter, *Cursor) ([]Block, *Cursor, error)
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
	logger     *persist.Logger
}

func newStormObjectNodeReader(db *StormDB, logger *persist.Logger) stormObjectNodeReader {
	return &stormObjectNode{
		node:       db.rootNode(nodeNameObjects),
		lastDataID: 0, // not given, nor needed
		logger:     logger,
	}
}

func newStormObjectNodeReaderWriter(db *StormDB, lastDataID StormDataID, logger *persist.Logger) stormObjectNodeReaderWriter {
	return &stormObjectNode{
		node:       db.rootNode(nodeNameObjects),
		lastDataID: lastDataID,
		logger:     logger,
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
		err = son.saveLockedOutput(obj.DataID, output)
		if err != nil {
			return fmt.Errorf("failed to save output %s by (object) data ID %d as locked: %v", output.ID.String(), obj.DataID, err)
		}
	}
	return nil
}

func (son *stormObjectNode) saveLockedOutput(dataID StormDataID, output *Output) error {
	var err error
	if output.UnlockReferencePoint.IsBlockHeight() {
		height := types.BlockHeight(output.UnlockReferencePoint)
		var lockedOutputs StormLockedOutputsByHeight
		err = son.node.One(nodeObjectOutputKeyHeight, height, &lockedOutputs)
		if err != nil {
			if err != storm.ErrNotFound {
				return fmt.Errorf("failed to get locked outputs (by height: %d) collection: %v", height, err)
			}
			lockedOutputs.Height = height
			lockedOutputs.DataIDs = []StormDataID{dataID}
		} else {
			lockedOutputs.DataIDs = append(lockedOutputs.DataIDs, dataID)
		}
		err = son.node.Save(&lockedOutputs)
		if err != nil {
			return fmt.Errorf("by height %d: %v", height, err)
		}
	} else { // by timestamp
		blockTime := types.Timestamp(output.UnlockReferencePoint)
		bucketID, pair := GetStormDataIDTimestampPairsForDataIDTimestampPair(dataID, blockTime)
		var lockedOutputs StormLockedOutputsByBucketTimestamp
		err = son.node.One(nodeObjectOutputKeyBucketID, bucketID, &lockedOutputs)
		if err != nil {
			if err != storm.ErrNotFound {
				return fmt.Errorf("failed to get locked outputs (by time %d, bucket %d) collection: %v", blockTime, bucketID, err)
			}
			lockedOutputs.BucketID = bucketID
			lockedOutputs.Pairs = []StormDataIDTimestampPair{pair}
		} else {
			lockedOutputs.Pairs = append(lockedOutputs.Pairs, pair)
		}
		err = son.node.Save(&lockedOutputs)
		if err != nil {
			return fmt.Errorf("by time %d (bucket %d): %v", blockTime, bucketID, err)
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

func (son *stormObjectNode) GetBlocks(limit int, filter *BlocksFilter, cursor *Cursor) ([]Block, *Cursor, error) {
	// unpack cursor (from a previous GetBlocks) query if defined
	if cursor != nil {
		var cursorFilter types.BlockHeight
		err := cursor.UnpackValue(&cursorFilter)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to unpack cursor")
		}
		if filter == nil {
			filter = &BlocksFilter{
				BlockHeight: NewBlockHeightFilterRange(&cursorFilter, nil),
				Timestamp:   nil, // not used by cursor
			}
		} else if filter.BlockHeight == nil {
			filter.BlockHeight = NewBlockHeightFilterRange(&cursorFilter, nil)
		} else {
			filter.BlockHeight.Begin = &cursorFilter
			if filter.BlockHeight.End != nil && *filter.BlockHeight.End < *filter.BlockHeight.Begin {
				return nil, nil, nil // nothing to do
			}
		}
	}
	// define the StormDB Query Matchers based on the used BlocksFilter,
	// unrelated to the fact that it might be defined/enforced by a cursor from a previous call
	var matchers []q.Matcher
	if filter != nil {
		if filter.BlockHeight != nil {
			if filter.BlockHeight.Begin != nil {
				matchers = append(matchers, q.Gte(nodeObjectBlockFieldHeight, *filter.BlockHeight.Begin))
			}
			if filter.BlockHeight.End != nil {
				matchers = append(matchers, q.Lte(nodeObjectBlockFieldHeight, *filter.BlockHeight.End))
			}
		}
		if filter.Timestamp != nil {
			if filter.Timestamp.Begin != nil {
				matchers = append(matchers, q.Gte(nodeObjectBlockFieldTimestamp, *filter.Timestamp.Begin))
			}
			if filter.Timestamp.End != nil {
				matchers = append(matchers, q.Lte(nodeObjectBlockFieldTimestamp, *filter.Timestamp.End))
			}
		}
		if filter.TransactionLength != nil {
			matchers = append(matchers, newTransactionIDLengthMatcher(filter.TransactionLength))
		}
	}
	// look up all blocks, optionally using matchers, but defintely with a limit
	// we allow one more than the limit, such that we can define a cursor if needed,
	// based on the used filter and the extra result (defining the next one to start from)
	var blocks []StormBlock
	err := son.node.Select(matchers...).Limit(limit + 1).Find(&blocks)
	if err != nil {
		return nil, nil, err
	}
	if len(blocks) <= limit {
		// if blocks were found, but not more than the defined limit,
		// we can simply return without the need for a new cursor
		apiBlocks, err := son.stormBlockSliceAsBlockSlice(blocks)
		if err != nil {
			return nil, nil, err
		}
		return apiBlocks, nil, nil
	}
	// create the next filter, such that we can turn it into a cursor,
	// and return it with the found blocks (minus the last block, as that one was only used to define the next cursor)
	nextFilter := blocks[len(blocks)-1].Height
	nextCursor, err := NewCursor(nextFilter)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create cursor from composed next filter: %v", err)
	}
	// all good, return the limited results, as well as
	// the cursor that can be used by the callee to continue this query where it left off
	apiBlocks, err := son.stormBlockSliceAsBlockSlice(blocks[:limit])
	if err != nil {
		return nil, nil, err
	}
	return apiBlocks, &nextCursor, nil

}

func (son *stormObjectNode) stormBlockSliceAsBlockSlice(sblocks []StormBlock) ([]Block, error) {
	// TODO: check if we should do this in a cheaper way
	dataIdentifiers := make([]StormDataID, 0, len(sblocks))
	for _, sblock := range sblocks {
		dataIdentifiers = append(dataIdentifiers, sblock.DataID)
	}
	var objects []StormObject
	err := son.node.Select(q.In(nodeObjectKeyDataID, dataIdentifiers)).Find(&objects)
	if err != nil {
		return nil, err
	}
	dataKeyObjectIDMapping := make(map[StormDataID]types.BlockID, len(objects))
	for _, obj := range objects {
		h, err := obj.ObjectID.AsHash()
		if err != nil {
			return nil, fmt.Errorf("failed to convert objectID from (block) object with dataID %d: %v", obj.DataID, err)
		}
		dataKeyObjectIDMapping[obj.DataID] = types.BlockID(h)
	}
	blockIdentifiers := make([]types.BlockID, 0, len(sblocks))
	for _, sblock := range sblocks {
		blockID, ok := dataKeyObjectIDMapping[sblock.DataID]
		if !ok {
			return nil, fmt.Errorf("failed to find ID for (block) object with dataID %d", sblock.DataID)
		}
		blockIdentifiers = append(blockIdentifiers, blockID)
	}
	return StormBlockSliceAsBlockSlice(sblocks, blockIdentifiers), nil
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
	// fetch all unlocked locked outputs by height first
	unlockedOutputs, err := son.unlockLockedOutputsByHeight(height)
	if err != nil {
		return nil, err
	}
	// fetch all unlocked locked outputs by time secondly and merge
	unlockedOutputsByTime, err := son.unlockLockedOutputsByTimeRange(minTimestamp, maxInclusiveTimestamp)
	return append(unlockedOutputs, unlockedOutputsByTime...), err
}

func (son *stormObjectNode) unlockLockedOutputsByHeight(height types.BlockHeight) ([]StormOutput, error) {
	var lockedOutputs StormLockedOutputsByHeight
	err := son.deleteByID(nodeObjectOutputKeyHeight, height, &lockedOutputs)
	if err != nil {
		if err == storm.ErrNotFound {
			return nil, nil // no error
		}
		return nil, fmt.Errorf("failed to get locked outputs (by height: %d) collection (and delete it): %v", height, err)
	}
	stormOutputs := make([]StormOutput, len(lockedOutputs.DataIDs))
	for idx, dataID := range lockedOutputs.DataIDs {
		err = son.node.One(nodeObjectKeyDataID, dataID, &stormOutputs[idx])
		if err != nil {
			return nil, fmt.Errorf("failed to get locked (by height: %d) output #%d (data id: %d): %v", height, idx+1, dataID, err)
		}
	}
	return stormOutputs, nil
}

func (son *stormObjectNode) unlockLockedOutputsByTimeRange(minTimestamp, maxInclusiveTimestamp types.Timestamp) ([]StormOutput, error) {
	if maxInclusiveTimestamp <= minTimestamp {
		// NOTE: this is only possible due to the fact that in a hacky way
		// rivine allows this to facilitate the POBS algorithm as well
		// as block creators with skewed clocks.
		return nil, nil // nothing to do
	}
	buckets := GetTimestampBucketIdentifiersForTimestampRange(minTimestamp, maxInclusiveTimestamp)
	bucketCollections := make([]StormLockedOutputsByBucketTimestamp, len(buckets))
	var err error
	// fetch all collections
	for idx, bucketID := range buckets {
		err = son.node.One(nodeObjectOutputKeyBucketID, bucketID, &bucketCollections[idx])
		if err != nil {
			if err != storm.ErrNotFound {
				return nil, fmt.Errorf("failed to fetch existing locked output collection for time-based bucket %d: %v", bucketID, err)
			}
			bucketCollections[idx].BucketID = bucketID
		}
	}

	// first data ID's are collected, afterwards this list
	// will be populated with the full data, if nothing goes wrong in between
	var stormOutputs []StormOutput

	// clear all outputs within range from the buckets, and delete the empty ones
	for _, bucketCollection := range bucketCollections {
		if len(bucketCollection.Pairs) == 0 {
			continue // ignore, can be assumed that bucket wasn't found (as we keep no empty buckets)
		}
		// if the entire bucket falls within the maxInclusive range we can simply delete it
		bucketTimestamp := types.Timestamp(bucketCollection.BucketID) * tsBasedBucketRange
		if bucketTimestamp <= maxInclusiveTimestamp {
			for _, pair := range bucketCollection.Pairs {
				stormOutputs = append(stormOutputs, StormOutput{
					DataID: pair.DataID,
				}) // keep track for later, so we can fetch it at the end
			}
			err = son.node.DeleteStruct(&bucketCollection)
			if err != nil {
				return nil, fmt.Errorf(
					"failed to delete (by range) collection for time-based bucket %d: %v",
					bucketCollection.BucketID, err)
			}
			continue // work is done for this collection
		}
		// otherwise we have to go through the list and manually delete those that fall within range,
		// and only delete the bucket if it is empty
		pairs := make([]StormDataIDTimestampPair, 0, len(bucketCollection.Pairs))
		for _, pair := range bucketCollection.Pairs {
			outputTimestamp := bucketTimestamp + types.Timestamp(pair.Offset)
			if outputTimestamp > minTimestamp && outputTimestamp <= maxInclusiveTimestamp {
				pairs = append(pairs, pair)
			} else {
				stormOutputs = append(stormOutputs, StormOutput{
					DataID: pair.DataID,
				}) // keep track for later, so we can fetch it at the end
			}
		}
		bucketCollection.Pairs = pairs
		if len(bucketCollection.Pairs) == 0 {
			err = son.node.DeleteStruct(&bucketCollection)
			if err != nil {
				return nil, fmt.Errorf(
					"failed to delete (by iteration) collection for time-based bucket %d: %v",
					bucketCollection.BucketID, err)
			}
		} else {
			err = son.node.Save(&bucketCollection)
			if err != nil {
				return nil, fmt.Errorf(
					"failed to save (after iteration update) collection for time-based bucket %d: %v",
					bucketCollection.BucketID, err)
			}
		}
	}

	// collect all gathered outputs
	for idx := range stormOutputs {
		err = son.node.One(nodeObjectKeyDataID, stormOutputs[idx].DataID, &stormOutputs[idx])
		if err != nil {
			return nil, fmt.Errorf(
				"failed to fetch existing relocked (by time ]%d,%d]) output %d: %v",
				minTimestamp, maxInclusiveTimestamp, stormOutputs[idx].DataID, err)
		}
	}

	// all good, return the unlocked outputs
	return stormOutputs, nil
}

func (son *stormObjectNode) RelockLockedOutputs(height types.BlockHeight, minTimestamp, maxInclusiveTimestamp types.Timestamp) ([]StormOutput, error) {
	// fetch all unlocked outputs
	var (
		err             error
		relockTimeBased bool
		relockedOutputs []StormOutput
	)
	if maxInclusiveTimestamp <= minTimestamp {
		// NOTE: this is only possible due to the fact that in a hacky way
		// rivine allows this to facilitate the POBS algorithm as well
		// as block creators with skewed clocks.
		err = son.node.Select(q.Eq(nodeObjectOutputKeyUnlockReferencePoint, ReferencePoint(height))).Find(&relockedOutputs)
	} else {
		relockTimeBased = true
		err = son.node.Select(q.Or(
			q.Eq(nodeObjectOutputKeyUnlockReferencePoint, ReferencePoint(height)),
			q.And(
				q.Gt(nodeObjectOutputKeyUnlockReferencePoint, ReferencePoint(minTimestamp)),
				q.Lte(nodeObjectOutputKeyUnlockReferencePoint, ReferencePoint(maxInclusiveTimestamp)),
			))).Find(&relockedOutputs)
	}
	if err != nil {
		return nil, err
	}
	if len(relockedOutputs) == 0 {
		return nil, nil // nothing to do
	}

	// cache the fetched collections, so we don't have to fetch for no reason
	var heightCollection *StormLockedOutputsByHeight // only one can be possible by height
	relockByHeight := func(dataID StormDataID) error {
		if heightCollection == nil {
			heightCollection = new(StormLockedOutputsByHeight)
			err := son.node.One(nodeObjectOutputKeyHeight, height, heightCollection)
			if err != nil {
				return err
			}
		}
		for _, knownDataID := range heightCollection.DataIDs {
			if knownDataID == dataID {
				return fmt.Errorf("output with data id %d is already referenced in the locked output (by height %d) collection", dataID, height)
			}
		}
		heightCollection.DataIDs = append(heightCollection.DataIDs, dataID)
		return nil
	}
	var (
		buckets               []StormTimeBucketID
		relockByTimestamp     func(dataID StormDataID, ts types.Timestamp) error
		timeBucketCollections map[StormTimeBucketID]*StormLockedOutputsByBucketTimestamp
	)
	if relockTimeBased {
		buckets = GetTimestampBucketIdentifiersForTimestampRange(minTimestamp, maxInclusiveTimestamp)
		timeBucketCollections = make(map[StormTimeBucketID]*StormLockedOutputsByBucketTimestamp, len(buckets)) // multiple can be possible by timestamp (unless the chain is sick it shouldn't be more then 2)
		relockByTimestamp = func(dataID StormDataID, ts types.Timestamp) error {
			bucketID, bucketOffset := GetTimestampBucketAndOffset(ts)
			timeCollection, ok := timeBucketCollections[bucketID]
			if !ok {
				timeCollection = new(StormLockedOutputsByBucketTimestamp)
				err := son.node.One(nodeObjectOutputKeyBucketID, bucketID, timeCollection)
				if err != nil {
					return err
				}
				timeBucketCollections[bucketID] = timeCollection
			}
			for _, knownPair := range timeCollection.Pairs {
				if knownPair.DataID == dataID {
					return fmt.Errorf(
						"output with data id %d is already referenced in the locked output (by timestamp %d, bucket %d) collection",
						dataID, ts, bucketID)
				}
			}
			timeCollection.Pairs = append(timeCollection.Pairs, StormDataIDTimestampPair{
				DataID: dataID,
				Offset: bucketOffset,
			})
			return nil
		}
	}
	// go through each unlocked output and store a lockd by reference once again
	for _, uo := range relockedOutputs {
		if uo.UnlockReferencePoint.IsBlockHeight() {
			err = relockByHeight(uo.DataID)
		} else { // by timestamp
			err = relockByTimestamp(uo.DataID, types.Timestamp(uo.UnlockReferencePoint))
		}
		if err != nil {
			return nil, err
		}
	}
	// save all cached collections
	if heightCollection != nil {
		err = son.node.Save(heightCollection)
		if err != nil {
			return nil, fmt.Errorf("failed to save updated height (%d) collection: %v", height, err)
		}
	}
	for bucketID, timeCollection := range timeBucketCollections {
		err = son.node.Save(timeCollection)
		if err != nil {
			return nil, fmt.Errorf("failed to save updated time-based bucket (%d) collection: %v", bucketID, err)
		}
	}

	// return all relocked outputs
	return relockedOutputs, err
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
		delErr := son.unreferenceLockedOutput(output.DataID, output.UnlockReferencePoint)
		if delErr != nil {
			son.logger.Printf(
				"[ERR] failed to unreference locked output %s (dataID %d, ref. point: %d) as part of delete process: %v",
				outputID.String(), output.DataID, output.UnlockReferencePoint, delErr)
		}
	}
	return output.AsOutput(outputID), nil
}

func (son *stormObjectNode) unreferenceLockedOutput(dataID StormDataID, refPoint ReferencePoint) error {
	if refPoint.IsBlockHeight() {
		// by height
		height := types.BlockHeight(refPoint)
		var lockedOutputs StormLockedOutputsByHeight
		err := son.node.One(nodeObjectOutputKeyHeight, height, &lockedOutputs)
		if err != nil {
			return fmt.Errorf(
				"failed to find output by data ID %d in found locked outputs by height %d collection: %v",
				dataID, height, err)
		}
		for idx, potDataID := range lockedOutputs.DataIDs {
			if dataID == potDataID {
				lockedOutputs.DataIDs = append(lockedOutputs.DataIDs[:idx], lockedOutputs.DataIDs[idx+1:]...)
				if len(lockedOutputs.DataIDs) != 0 {
					// save remaining identifiers
					return son.node.Save(&lockedOutputs)
				}
				// delete empty collection
				return son.node.DeleteStruct(&lockedOutputs)
			}
		}
		return fmt.Errorf(
			"failed to find output by data ID %d in found locked outputs by height %d collection, did find: %v",
			dataID, height, lockedOutputs.DataIDs)
	}

	// by timestamp
	timestamp := types.Timestamp(refPoint)
	bucketID, _ := GetTimestampBucketAndOffset(timestamp)
	var lockedOutputs StormLockedOutputsByBucketTimestamp
	err := son.node.One(nodeObjectOutputKeyHeight, bucketID, &lockedOutputs)
	if err != nil {
		return fmt.Errorf(
			"failed to find output by data ID %d in found locked outputs by timestamp %d collection: %v",
			dataID, timestamp, err)
	}
	for idx, pair := range lockedOutputs.Pairs {
		if dataID == pair.DataID {
			lockedOutputs.Pairs = append(lockedOutputs.Pairs[:idx], lockedOutputs.Pairs[idx+1:]...)
			if len(lockedOutputs.Pairs) != 0 {
				// save remaining pairs
				return son.node.Save(&lockedOutputs)
			}
			// delete empty collection
			return son.node.DeleteStruct(&lockedOutputs)
		}
	}
	return fmt.Errorf(
		"failed to find output by data ID %d in found locked outputs by timestamp %d collection, did find: %v",
		dataID, timestamp, lockedOutputs.Pairs)
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
