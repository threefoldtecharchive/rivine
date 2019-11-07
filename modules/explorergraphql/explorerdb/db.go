package explorerdb

import (
	"errors"
	"fmt"
	"math"

	"github.com/threefoldtech/rivine/build"
	"github.com/threefoldtech/rivine/crypto"
	"github.com/threefoldtech/rivine/modules"
	"github.com/threefoldtech/rivine/pkg/encoding/rivbin"
	"github.com/threefoldtech/rivine/types"
)

// TODO: integrate context.Context in each call

// TODO: we should not have to rely on CS data for getting the child target

// TODO: keep reference counter for public keys, and delete it in case the reference count is 0 (see TODO (4))

type DB interface {
	// You can run each call in its own R/W  Txn by calling
	// the txn command directly on the DB.
	RWTxn

	// ReadTransaction batches multiple read calls together,
	// to keep the disk I/O to a minimum
	ReadTransaction(func(RTxn) error) error
	// ReadWriteTransaction batches multiple read-write calls together,
	// to keep the disk I/O to a minimum
	ReadWriteTransaction(func(RWTxn) error) error

	// Close the DB
	Close() error
}

type RTxn interface {
	GetChainContext() (ChainContext, error)

	GetChainAggregatedFacts() (ChainAggregatedFacts, error)

	GetObject(ObjectID) (Object, error)
	GetObjectInfo(ObjectID) (ObjectInfo, error)

	GetBlock(types.BlockID) (Block, error)
	GetBlockFacts(types.BlockID) (BlockFacts, error)
	GetBlockAt(types.BlockHeight) (Block, error)
	GetBlockIDAt(types.BlockHeight) (types.BlockID, error)
	GetTransaction(types.TransactionID) (Transaction, error)
	GetOutput(types.OutputID) (Output, error)

	GetBlocks(limit *int, filter *BlocksFilter, cursor *Cursor) ([]Block, *Cursor, error)
	GetBlockIdentifiers(limit *int, filter *BlockIdentifiersFilter, cursor *Cursor) ([]types.BlockID, *Cursor, error)

	GetFreeForAllWallet(types.UnlockHash) (FreeForAllWalletData, error)
	GetSingleSignatureWallet(types.UnlockHash) (SingleSignatureWalletData, error)
	GetMultiSignatureWallet(types.UnlockHash) (MultiSignatureWalletData, error)
	GetAtomicSwapContract(types.UnlockHash) (AtomicSwapContract, error)

	GetPublicKey(types.UnlockHash) (types.PublicKey, error)
}

type RWTxn interface {
	RTxn

	SetChainContext(ChainContext) error

	ApplyBlock(block Block, blockFacts BlockFactsConstants, txs []Transaction, outputs []Output, inputs map[types.OutputID]OutputSpenditureData) error
	// TODO: should we also revert public key (from UH) mapping? (TODO 4)
	RevertBlock(blockContext BlockRevertContext, txs []types.TransactionID, outputs []types.OutputID, inputs []types.OutputID) error

	// commit the work done from Memory to Disk,
	// only required in case you are doing a big amount of calls within a single transaction.
	// If you want to continue using this transaction, you'll have to set final to true
	Commit(final bool) error
}

type (
	TimestampFilter interface {
		TimestampEndpoints() (begin, end *types.Timestamp)
	}
	TimestampFilterRange struct { // the only TimesTamp filter that should ever be MsgPack encoded/decoded
		Begin *types.Timestamp `msgpack:"b"`
		End   *types.Timestamp `msgpack:"e"`
	}
	TimestampFilterBefore types.Timestamp
	TimestampFilterAfter  types.Timestamp

	BlockHeightFilter interface {
		BlockHeightEndpoints() (begin, end *types.BlockHeight)
	}
	BlockHeightFilterRange struct { // the only BlockHeight filter that should ever be MsgPack encoded/decoded
		Begin *types.BlockHeight `msgpack:"b"`
		End   *types.BlockHeight `msgpack:"e"`
	}
	BlockHeightFilterBefore types.BlockHeight
	BlockHeightFilterAfter  types.BlockHeight

	BlocksFilter struct {
		BlockHeight BlockHeightFilter
		Timestamp   TimestampFilter
	}
	// Used for Cursor encoding/decoding (MsgPack) purposes,
	// simplifying the BlocksFilter union typing.
	BlocksFilterCursor struct {
		BlockHeight *BlockHeightFilterRange `msgpack:"h"`
		Timestamp   *TimestampFilterRange   `msgpack:"t"`
	}

	BlockIdentifiersFilter struct {
		BlockHeight BlockHeightFilter
	}
	// Used for Cursor encoding/decoding (MsgPack) purposes,
	// simplifying the BlockIdentifiersFilter union typing.
	BlockIdentifiersFilterCursor struct {
		BlockHeight *BlockHeightFilterRange `msgpack:"h"`
	}
)

func (cur *BlocksFilterCursor) AsBlocksFilter() *BlocksFilter {
	return &BlocksFilter{
		BlockHeight: cur.BlockHeight,
		Timestamp:   cur.Timestamp,
	}
}

func (cur *BlockIdentifiersFilterCursor) AsBlockIdentifiersFilter() *BlockIdentifiersFilter {
	return &BlockIdentifiersFilter{
		BlockHeight: cur.BlockHeight,
	}
}

func NewTimestampFilterRange(begin, end *types.Timestamp) *TimestampFilterRange {
	if begin == nil && end == nil {
		return nil
	}
	return &TimestampFilterRange{
		Begin: begin,
		End:   end,
	}
}
func NewTimestampFilter(begin, end *types.Timestamp) TimestampFilter {
	return NewTimestampFilterRange(begin, end)
}

func (tsr *TimestampFilterRange) TimestampEndpoints() (begin, end *types.Timestamp) {
	if tsr == nil {
		return nil, nil
	}
	return tsr.Begin, tsr.End
}

func (tsb TimestampFilterBefore) TimestampEndpoints() (begin, end *types.Timestamp) {
	ts := types.Timestamp(tsb)
	if ts > 0 {
		ts--
	}
	return nil, &ts
}
func (tsb TimestampFilterBefore) MarshalMsgpack() ([]byte, error) {
	return nil, errors.New("BUG: TimestampFilterBefore does not support MsgPack encoding, TimestampFilterRange should be used instead")
}
func (tsb *TimestampFilterBefore) UnmarshalMsgpack([]byte) error {
	return errors.New("BUG: TimestampFilterBefore does not support MsgPack decoding, TimestampFilterRange should be used instead")
}

func (tsa TimestampFilterAfter) TimestampEndpoints() (begin, end *types.Timestamp) {
	ts := types.Timestamp(tsa)
	if ts < math.MaxUint64 {
		ts++
	}
	return &ts, nil
}
func (tsa TimestampFilterAfter) MarshalMsgpack() ([]byte, error) {
	return nil, errors.New("BUG: TimestampFilterAfter does not support MsgPack encoding, TimestampFilterRange should be used instead")
}
func (tsa *TimestampFilterAfter) UnmarshalMsgpack([]byte) error {
	return errors.New("BUG: TimestampFilterAfter does not support MsgPack decoding, TimestampFilterRange should be used instead")
}

var (
	_ TimestampFilter = (*TimestampFilterRange)(nil)
	_ TimestampFilter = TimestampFilterBefore(0)
	_ TimestampFilter = TimestampFilterAfter(0)
)

func NewBlockHeightFilterRange(begin, end *types.BlockHeight) *BlockHeightFilterRange {
	if begin == nil && end == nil {
		return nil
	}
	return &BlockHeightFilterRange{
		Begin: begin,
		End:   end,
	}
}

func NewBlockHeightFilter(begin, end *types.BlockHeight) BlockHeightFilter {
	return NewBlockHeightFilterRange(begin, end)
}

func (bhr *BlockHeightFilterRange) BlockHeightEndpoints() (begin, end *types.BlockHeight) {
	if bhr == nil {
		return nil, nil
	}
	return bhr.Begin, bhr.End
}

func (bhb BlockHeightFilterBefore) BlockHeightEndpoints() (begin, end *types.BlockHeight) {

	bh := types.BlockHeight(bhb)
	if bh > 0 {
		bh--
	}
	return nil, &bh
}
func (bhb BlockHeightFilterBefore) MarshalMsgpack() ([]byte, error) {
	return nil, errors.New("BUG: BlockHeightFilterBefore does not support MsgPack encoding, BlockHeightFilterRange should be used instead")
}
func (bhb *BlockHeightFilterBefore) BlockHeightFilterBefore([]byte) error {
	return errors.New("BUG: TimestampFilterBefore does not support MsgPack decoding, BlockHeightFilterRange should be used instead")
}

func (bha BlockHeightFilterAfter) BlockHeightEndpoints() (begin, end *types.BlockHeight) {
	bh := types.BlockHeight(bha)
	if bh < math.MaxUint64 {
		bh++
	}
	return &bh, nil
}
func (bha BlockHeightFilterAfter) MarshalMsgpack() ([]byte, error) {
	return nil, errors.New("BUG: BlockHeightFilterAfter does not support MsgPack encoding, BlockHeightFilterRange should be used instead")
}
func (bha *BlockHeightFilterAfter) BlockHeightFilterAfter([]byte) error {
	return errors.New("BUG: BlockHeightFilterAfter does not support MsgPack decoding, BlockHeightFilterRange should be used instead")
}

var (
	_ TimestampFilter = (*TimestampFilterRange)(nil)
	_ TimestampFilter = TimestampFilterBefore(0)
	_ TimestampFilter = TimestampFilterAfter(0)
)

// TODO: handle also chain-specific stuff, such as chains that do not have block rewards

func ApplyConsensusChangeWithChannel(db DB, cs modules.ConsensusSet, ch <-chan modules.ConsensusChange, chainCts *types.ChainConstants) error {
	const (
		minBlocksPerCommit = 1000
	)
	var blockCount = 0
	return db.ReadWriteTransaction(func(db RWTxn) error {
		var err error
		for csc := range ch {
			if len(csc.AppliedBlocks) == 0 {
				build.Critical("Explorer.ProcessConsensusChange called with a ConsensusChange that has no AppliedBlocks")
			}
			err = applyConsensusChangeForRWTxn(db, cs, csc, chainCts)
			if err != nil {
				return err
			}
			blockCount -= len(csc.RevertedBlocks)
			blockCount += len(csc.AppliedBlocks)
			if blockCount >= minBlocksPerCommit {
				err = db.Commit(false)
				if err != nil {
					return fmt.Errorf("failed to commit last (net) %d blocks: %v", blockCount, err)
				}
				blockCount = 0
			}
		}
		return nil
	})
}

func ApplyConsensusChange(db DB, cs modules.ConsensusSet, csc modules.ConsensusChange, chainCts *types.ChainConstants) error {
	return db.ReadWriteTransaction(func(db RWTxn) error {
		return applyConsensusChangeForRWTxn(db, cs, csc, chainCts)
	})
}

func applyConsensusChangeForRWTxn(db RWTxn, cs modules.ConsensusSet, csc modules.ConsensusChange, chainCts *types.ChainConstants) error {
	chainCtx, err := db.GetChainContext()
	if err != nil {
		return err
	}

	for _, revertedBlock := range csc.RevertedBlocks {
		// TODO: verify if this is correct, or if it should be done after
		chainCtx.Height--
		chainCtx.Timestamp = revertedBlock.Timestamp

		block := RivineBlockAsExplorerBlock(chainCtx.Height, revertedBlock)
		chainCtx.BlockID = block.ID

		outputs := make([]types.OutputID, 0, len(revertedBlock.MinerPayouts))
		for idx := range revertedBlock.MinerPayouts {
			outputs = append(outputs, block.Payouts[idx])
		}

		var inputs []types.OutputID
		transactions := make([]types.TransactionID, 0, len(revertedBlock.Transactions))
		for idx, txn := range revertedBlock.Transactions {
			// add txn
			transactions = append(transactions, block.Transactions[idx])
			// add inputs
			for _, input := range txn.CoinInputs {
				inputs = append(inputs, types.OutputID(input.ParentID))
			}
			for _, input := range txn.BlockStakeInputs {
				inputs = append(inputs, types.OutputID(input.ParentID))
			}
			// add outputs
			for cidx := range txn.CoinOutputs {
				outputs = append(outputs, types.OutputID(txn.CoinOutputID(uint64(cidx))))
			}
			for bsidx := range txn.BlockStakeOutputs {
				outputs = append(outputs, types.OutputID(txn.BlockStakeOutputID(uint64(bsidx))))
			}
		}

		err = db.RevertBlock(BlockRevertContext{
			ID:        block.ID,
			Height:    chainCtx.Height,
			Timestamp: chainCtx.Timestamp,
		}, transactions, outputs, inputs)
		if err != nil {
			return err
		}
	}

	for _, appliedBlock := range csc.AppliedBlocks {
		var target types.Target
		if chainCtx.Height > 0 {
			// TODO: find a better way than having to get this target from the consensusSet DB
			var ok bool
			target, ok = cs.ChildTarget(appliedBlock.ParentID)
			if !ok {
				return fmt.Errorf("failed to look up child target for parent block %s", appliedBlock.ParentID.String())
			}
		} else {
			target = chainCts.RootTarget()
		}
		blockFacts := BlockFactsConstants{
			Target:     target,
			Difficulty: target.Difficulty(chainCts.RootDepth),
		}

		block := RivineBlockAsExplorerBlock(chainCtx.Height, appliedBlock)

		outputs := make([]Output, 0, len(appliedBlock.MinerPayouts))
		for idx, mp := range appliedBlock.MinerPayouts {
			outputs = append(outputs, RivineMinerPayoutAsOutput(
				block.ID,
				types.CoinOutputID(block.Payouts[idx]),
				mp,
				// TODO: customize this per chain network (behaviour and constants)
				idx == 0,
				chainCtx.Height,
				chainCts.MaturityDelay,
			))
		}
		// TODO: customize this per chain network
		var feePayoutID types.OutputID
		if len(block.Payouts) > 1 {
			feePayoutID = block.Payouts[1]
		}

		inputs := make(map[types.OutputID]OutputSpenditureData)
		transactions := make([]Transaction, 0, len(appliedBlock.Transactions))
		for txidx, txn := range appliedBlock.Transactions {
			transaction := RivineTransactionAsTransaction(
				block.ID,
				block.Transactions[txidx],
				txn,
				feePayoutID,
			)
			transactions = append(transactions, transaction)
			// add inputs
			for _, input := range txn.CoinInputs {
				inputs[types.OutputID(input.ParentID)] = OutputSpenditureData{
					Fulfillment:              input.Fulfillment,
					FulfillmentTransactionID: block.Transactions[txidx],
				}
			}
			for _, input := range txn.BlockStakeInputs {
				inputs[types.OutputID(input.ParentID)] = OutputSpenditureData{
					Fulfillment:              input.Fulfillment,
					FulfillmentTransactionID: block.Transactions[txidx],
				}
			}
			// add outputs
			for coidx, output := range txn.CoinOutputs {
				outputs = append(outputs, RivineCoinOutputAsOutput(
					block.Transactions[txidx],
					types.CoinOutputID(transaction.CoinOutputs[coidx]),
					output,
				))
			}
			for bsidx, output := range txn.BlockStakeOutputs {
				outputs = append(outputs, RivineBlockStakeOutputAsOutput(
					block.Transactions[txidx],
					types.BlockStakeOutputID(transaction.BlockStakeOutputs[bsidx]),
					output,
				))
			}
		}

		err = db.ApplyBlock(block, blockFacts, transactions, outputs, inputs)
		if err != nil {
			return err
		}

		// TODO: verify if this is correct, or if it should be done before
		chainCtx.Height++
		chainCtx.Timestamp = appliedBlock.Timestamp
		chainCtx.BlockID = block.ID
	}

	chainCtx.ConsensusChangeID = csc.ID
	err = db.SetChainContext(chainCtx)
	return err
}

func RivineBlockAsExplorerBlock(height types.BlockHeight, block types.Block) Block {
	// aggregate payouts (as a list of identifiers)
	payouts := make([]types.OutputID, 0, len(block.MinerPayouts))
	for idx := range block.MinerPayouts {
		payouts = append(payouts, types.OutputID(block.MinerPayoutID(uint64(idx))))
	}
	// aggregate transactions (as a list of identifiers)
	transactions := make([]types.TransactionID, 0, len(block.Transactions))
	for _, txn := range block.Transactions {
		transactions = append(transactions, txn.ID())
	}
	// return the block
	return Block{
		ID:           block.ID(),
		ParentID:     block.ParentID,
		Height:       height,
		Timestamp:    block.Timestamp,
		Payouts:      payouts,
		Transactions: transactions,
	}
}

func RivineTransactionAsTransaction(parent types.BlockID, id types.TransactionID, rtxn types.Transaction, feePayoutID types.OutputID) Transaction {
	// aggregate inputs (as a list of identifiers)
	coinInputs := make([]types.OutputID, 0, len(rtxn.CoinInputs))
	for _, input := range rtxn.CoinInputs {
		coinInputs = append(coinInputs, types.OutputID(input.ParentID))
	}
	blockStakeInputs := make([]types.OutputID, 0, len(rtxn.BlockStakeInputs))
	for _, input := range rtxn.BlockStakeInputs {
		blockStakeInputs = append(blockStakeInputs, types.OutputID(input.ParentID))
	}
	// aggregate outputs (as a list of identifiers)
	coinOutputs := make([]types.OutputID, 0, len(rtxn.CoinOutputs))
	for idx := range rtxn.CoinOutputs {
		coinOutputs = append(coinOutputs, types.OutputID(rtxn.CoinOutputID(uint64(idx))))
	}
	blockStakeOutputs := make([]types.OutputID, 0, len(rtxn.BlockStakeOutputs))
	for idx := range rtxn.BlockStakeOutputs {
		blockStakeOutputs = append(blockStakeOutputs, types.OutputID(rtxn.BlockStakeOutputID(uint64(idx))))
	}
	// encode extension data
	var encodedExtensionData []byte
	if rtxn.Extension != nil {
		var err error
		encodedExtensionData, err = rivbin.Marshal(rtxn.Extension)
		if err != nil {
			build.Severe("failed to encode rivine txn extension data", err)
		}
	}
	// return transaction
	return Transaction{
		ID: id,

		ParentBlock: parent,
		Version:     rtxn.Version,

		CoinInputs:  coinInputs,
		CoinOutputs: coinOutputs,

		BlockStakeInputs:  blockStakeInputs,
		BlockStakeOutputs: blockStakeOutputs,

		FeePayout: TransactionFeePayoutInfo{
			PayoutID: feePayoutID,
			Values:   rtxn.MinerFees,
		},

		ArbitraryData:        rtxn.ArbitraryData,
		EncodedExtensionData: encodedExtensionData,
	}
}

type UnlockHashPublicKeyPair struct {
	UnlockHash types.UnlockHash
	PublicKey  types.PublicKey
}

func RivineUnlockHashPublicKeyPairsFromFulfillment(fulfillment types.UnlockFulfillmentProxy) []UnlockHashPublicKeyPair {
	switch ft := fulfillment.FulfillmentType(); ft {
	case types.FulfillmentTypeSingleSignature:
		ssft := fulfillment.Fulfillment.(*types.SingleSignatureFulfillment)
		return []UnlockHashPublicKeyPair{
			{
				UnlockHash: RivineUnlockHashFromPublicKey(ssft.PublicKey),
				PublicKey:  ssft.PublicKey,
			},
		}
	case types.FulfillmentTypeMultiSignature:
		msft := fulfillment.Fulfillment.(*types.MultiSignatureFulfillment)
		var pairs []UnlockHashPublicKeyPair
		for _, pair := range msft.Pairs {
			pairs = append(pairs, UnlockHashPublicKeyPair{
				UnlockHash: RivineUnlockHashFromPublicKey(pair.PublicKey),
				PublicKey:  pair.PublicKey,
			})
		}
		return pairs
	default:
		build.Critical(fmt.Sprintf("unsupported fulfillment type %d: %v", ft, fulfillment))
	}
	// should never reach here
	return nil
}

func RivineUnlockHashFromPublicKey(pk types.PublicKey) types.UnlockHash {
	uh, err := types.NewPubKeyUnlockHash(pk)
	if err != nil {
		build.Severe("failed to convert unlock hash to public key", pk, err)
	}
	return uh
}

func RivineMinerPayoutAsOutput(parent types.BlockID, id types.CoinOutputID, payout types.MinerPayout, reward bool, height types.BlockHeight, delay types.BlockHeight) Output {
	// define output type
	var ot OutputType
	if reward {
		ot = OutputTypeBlockCreationReward
	} else {
		ot = OutputTypeTransactionFee
	}

	condition := types.NewCondition(types.NewTimeLockCondition(
		uint64(height+delay),
		types.NewUnlockHashCondition(payout.UnlockHash)))
	unlockReferencePoint, _ := UnlockReferencePointFromCondition(condition)

	// return output
	return Output{
		ID:                   types.OutputID(id),
		ParentID:             crypto.Hash(parent),
		Type:                 ot,
		Value:                payout.Value,
		Condition:            condition,
		UnlockReferencePoint: unlockReferencePoint,
		SpenditureData:       nil,
	}
}

func RivineCoinOutputAsOutput(parent types.TransactionID, id types.CoinOutputID, output types.CoinOutput) Output {
	unlockReferencePoint, _ := UnlockReferencePointFromCondition(output.Condition)
	return Output{
		ID:                   types.OutputID(id),
		ParentID:             crypto.Hash(parent),
		Type:                 OutputTypeCoin,
		Value:                output.Value,
		Condition:            output.Condition,
		UnlockReferencePoint: unlockReferencePoint,
	}
}

func RivineBlockStakeOutputAsOutput(parent types.TransactionID, id types.BlockStakeOutputID, output types.BlockStakeOutput) Output {
	unlockReferencePoint, _ := UnlockReferencePointFromCondition(output.Condition)
	return Output{
		ID:                   types.OutputID(id),
		ParentID:             crypto.Hash(parent),
		Type:                 OutputTypeBlockStake,
		Value:                output.Value,
		Condition:            output.Condition,
		UnlockReferencePoint: unlockReferencePoint,
	}
}
