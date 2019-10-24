package explorergraphql

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/threefoldtech/rivine/crypto"
	"github.com/threefoldtech/rivine/extensions/minting"
	"github.com/threefoldtech/rivine/modules"
	"github.com/threefoldtech/rivine/modules/explorergraphql/explorerdb"
	"github.com/threefoldtech/rivine/pkg/encoding/rivbin"
	"github.com/threefoldtech/rivine/types"
)

// THIS CODE IS A STARTING POINT ONLY. IT WILL NOT BE UPDATED WITH SCHEMA CHANGES.

// TODO: check how to do better error handling

// TODO: ensure we handle errors as gracefully as possible
//      (e.g. do not return errors when a warning is sufficient,
//            and do not stop the world for one property failure)

type Resolver struct {
	db             explorerdb.DB
	cs             modules.ConsensusSet
	chainConstants types.ChainConstants
	blockchainInfo types.BlockchainInfo
}

func (r *Resolver) Block() BlockResolver {
	return &blockResolver{r}
}
func (r *Resolver) BlockFacts() BlockFactsResolver {
	return &blockFactsResolver{r}
}
func (r *Resolver) BlockHeader() BlockHeaderResolver {
	return &blockHeaderResolver{r}
}
func (r *Resolver) ChainFacts() ChainFactsResolver {
	return &chainFactsResolver{r}
}
func (r *Resolver) MintCoinCreationTransaction() MintCoinCreationTransactionResolver {
	return &mintCoinCreationTransactionResolver{r}
}
func (r *Resolver) MintCoinDestructionTransaction() MintCoinDestructionTransactionResolver {
	return &mintCoinDestructionTransactionResolver{r}
}
func (r *Resolver) MintConditionDefinitionTransaction() MintConditionDefinitionTransactionResolver {
	return &mintConditionDefinitionTransactionResolver{r}
}
func (r *Resolver) Output() OutputResolver {
	return &outputResolver{r}
}
func (r *Resolver) QueryRoot() QueryRootResolver {
	return &queryRootResolver{r}
}
func (r *Resolver) StandardTransaction() StandardTransactionResolver {
	return &standardTransactionResolver{r}
}
func (r *Resolver) TransactionFeePayout() TransactionFeePayoutResolver {
	return &transactionFeePayoutResolver{r}
}
func (r *Resolver) TransactionParentInfo() TransactionParentInfoResolver {
	return &transactionParentInfoResolver{r}
}
func (r *Resolver) UnlockHashCondition() UnlockHashConditionResolver {
	return &unlockHashConditionResolver{r}
}
func (r *Resolver) UnlockHashPublicKeyPair() UnlockHashPublicKeyPairResolver {
	return &unlockHashPublicKeyPairResolver{r}
}

type blockResolver struct{ *Resolver }

func (r *blockResolver) Facts(ctx context.Context, obj *Block) (*BlockFacts, error) {
	panic("not implemented")
}
func (r *blockResolver) Transactions(ctx context.Context, obj *Block) ([]Transaction, error) {
	ltxns := len(obj.Transactions)
	txns := make([]Transaction, 0, ltxns)
	txnParents := make([]*TransactionParentInfo, 0, ltxns)
	for idx, txn := range obj.Transactions {
		stxn := txn.(*StubTransaction)
		dbTxn, err := r.db.GetTransaction(types.TransactionID(stxn.ID))
		if err != nil {
			return nil, err
		}
		index := idx
		txnParent := &TransactionParentInfo{
			ID:               obj.Header.ID,
			ParentID:         obj.Header.ParentID,
			Height:           obj.Header.BlockHeight,
			Timestamp:        obj.Header.BlockTime,
			TransactionOrder: &index,
		}
		txn, err := dbTransactionAsGQL(txnParent, &dbTxn)
		if err != nil {
			return nil, err
		}
		txns = append(txns, txn)
		txnParents = append(txnParents, txnParent)
	}
	for idx := range obj.Transactions {
		txnParents[idx].SiblingTransactions = allTransactionsExcept(txns, idx)
	}
	return txns, nil
}

func allTransactionsExcept(txns []Transaction, ignoreIndex int) []Transaction {
	ltxns := len(txns)
	if ltxns == 0 {
		return nil
	}
	ntxns := make([]Transaction, 0, ltxns-1)
	for idx, txn := range txns {
		if idx != ignoreIndex {
			ntxns = append(ntxns, txn)
		}
	}
	return ntxns
}

type blockFactsResolver struct{ *Resolver }

func (r *blockFactsResolver) Aggregated(ctx context.Context, obj *BlockFacts) (*BlockAggregatedFacts, error) {
	panic("not implemented")
}

type blockHeaderResolver struct{ *Resolver }

func (r *blockHeaderResolver) Parent(ctx context.Context, obj *BlockHeader) (*Block, error) {
	if obj.ParentID == nil {
		return nil, nil // no block to return
	}
	dbBlock, err := r.db.GetBlock(types.BlockID(*obj.ParentID))
	if err != nil {
		return nil, fmt.Errorf("internal DB error while fetching parent block %s: %v", obj.ParentID.String(), err)
	}
	return dbBlockAsGQL(ctx, r.db, &dbBlock)
}
func (r *blockHeaderResolver) Child(ctx context.Context, obj *BlockHeader) (*Block, error) {
	if obj.BlockHeight == nil {
		return nil, errors.New("internal server error: Block BlockHeight is not set")
	}
	dbBlock, err := r.db.GetBlockByReferencePoint(explorerdb.ReferencePoint(*obj.BlockHeight) + 1)
	if err != nil {
		// TODO: distinguish between NotFound (possible if last block) and other types of errors
		return nil, nil
	}
	return dbBlockAsGQL(ctx, r.db, &dbBlock)
}
func (r *blockHeaderResolver) Payouts(ctx context.Context, obj *BlockHeader) ([]*BlockPayout, error) {
	payoutLen := len(obj.Payouts)
	if payoutLen == 0 {
		return nil, nil
	}
	payouts := make([]*BlockPayout, 0, payoutLen)
	for _, blockPayout := range obj.Payouts {
		dbOutput, err := r.db.GetOutput(types.OutputID(blockPayout.Output.ID))
		// TODO: is returning errors here desired, or should we handle this more gracefully for individual cases
		if err != nil {
			return nil, err
		}
		var payoutType *BlockPayoutType
		if dbOutput.Type == explorerdb.OutputTypeBlockCreationReward {
			pt := BlockPayoutTypeBlockReward
			payoutType = &pt
		} else if dbOutput.Type == explorerdb.OutputTypeTransactionFee {
			pt := BlockPayoutTypeTransactionFee
			payoutType = &pt
		}
		gqlOutput, err := dbOutputAsGQL(&dbOutput, nil)
		if err != nil {
			return nil, err
		}
		payouts = append(payouts, &BlockPayout{
			Output: gqlOutput,
			Type:   payoutType,
		})
	}
	return payouts, nil
}

type chainFactsResolver struct{ *Resolver }

func (r *chainFactsResolver) LastBlock(ctx context.Context, obj *ChainFacts) (*Block, error) {
	// default to latest block
	chainCtx, err := r.db.GetChainContext()
	if err != nil {
		return nil, err
	}
	ref := chainCtx.Height
	if ref > 0 {
		ref-- // chainCtx.Height defines the amount of blocks (in other words, the height of the chain), not the height of latest
	}

	dbBlock, err := r.db.GetBlockByReferencePoint(explorerdb.ReferencePoint(ref))
	if err != nil {
		return nil, fmt.Errorf("internal DB error while fetching last block: %v", err)
	}
	return dbBlockAsGQL(ctx, r.db, &dbBlock)
}
func (r *chainFactsResolver) Aggregated(ctx context.Context, obj *ChainFacts) (*ChainAggregatedData, error) {
	panic("not implemented")
}

type mintCoinCreationTransactionResolver struct{ *Resolver }

func (r *mintCoinCreationTransactionResolver) ParentBlock(ctx context.Context, obj *MintCoinCreationTransaction) (*TransactionParentInfo, error) {
	return dbFillTransactionParentInfoForTxn(ctx, r.db, obj.ParentBlock, types.TransactionID(obj.ID))
}
func (r *mintCoinCreationTransactionResolver) CoinInputs(ctx context.Context, obj *MintCoinCreationTransaction) ([]*Input, error) {
	return resolveCoinInputs(ctx, obj.CoinInputs, r.db)
}
func (r *mintCoinCreationTransactionResolver) CoinOutputs(ctx context.Context, obj *MintCoinCreationTransaction) ([]*Output, error) {
	return resolveCoinOutputs(ctx, obj.CoinOutputs, r.db)
}

type mintCoinDestructionTransactionResolver struct{ *Resolver }

func (r *mintCoinDestructionTransactionResolver) ParentBlock(ctx context.Context, obj *MintCoinDestructionTransaction) (*TransactionParentInfo, error) {
	return dbFillTransactionParentInfoForTxn(ctx, r.db, obj.ParentBlock, types.TransactionID(obj.ID))
}
func (r *mintCoinDestructionTransactionResolver) CoinInputs(ctx context.Context, obj *MintCoinDestructionTransaction) ([]*Input, error) {
	return resolveCoinInputs(ctx, obj.CoinInputs, r.db)
}
func (r *mintCoinDestructionTransactionResolver) CoinOutputs(ctx context.Context, obj *MintCoinDestructionTransaction) ([]*Output, error) {
	return resolveCoinOutputs(ctx, obj.CoinOutputs, r.db)
}

type mintConditionDefinitionTransactionResolver struct{ *Resolver }

func (r *mintConditionDefinitionTransactionResolver) ParentBlock(ctx context.Context, obj *MintConditionDefinitionTransaction) (*TransactionParentInfo, error) {
	return dbFillTransactionParentInfoForTxn(ctx, r.db, obj.ParentBlock, types.TransactionID(obj.ID))
}
func (r *mintConditionDefinitionTransactionResolver) CoinInputs(ctx context.Context, obj *MintConditionDefinitionTransaction) ([]*Input, error) {
	return resolveCoinInputs(ctx, obj.CoinInputs, r.db)
}
func (r *mintConditionDefinitionTransactionResolver) CoinOutputs(ctx context.Context, obj *MintConditionDefinitionTransaction) ([]*Output, error) {
	return resolveCoinOutputs(ctx, obj.CoinOutputs, r.db)
}

type outputResolver struct{ *Resolver }

func (r *outputResolver) Parent(ctx context.Context, obj *Output) (OutputParent, error) {
	dbObject, err := r.db.GetObject(explorerdb.ObjectID(obj.ParentID[:]))
	if err != nil {
		return nil, fmt.Errorf("internal DB error while fetching transaction: %v", err)
	}
	switch dbObject.Type {
	case explorerdb.ObjectTypeBlock:
		dbBlock, ok := dbObject.Data.(explorerdb.Block)
		if !ok {
			return nil, fmt.Errorf("internal server error: unexpected type %T for object of type block (%d)", dbObject.Data, dbObject.Type)
		}
		return dbBlockAsGQL(ctx, r.db, &dbBlock)
	case explorerdb.ObjectTypeTransaction:
		dbTransaction, ok := dbObject.Data.(explorerdb.Transaction)
		if !ok {
			return nil, fmt.Errorf("internal server error: unexpected type %T for object of type transaction (%d)", dbObject.Data, dbObject.Type)
		}
		return dbTransactionAsOutputParent(&TransactionParentInfo{
			ID: crypto.Hash(dbTransaction.ParentBlock),
		}, &dbTransaction)
	default:
		return nil, fmt.Errorf("internal server error: unsupported object type %d used as output parent", dbObject.Type)
	}
}

type queryRootResolver struct{ *Resolver }

func (r *queryRootResolver) Chain(ctx context.Context) (*ChainFacts, error) {
	constants, err := rivConstantsAsGQL(r.cs, &r.chainConstants, &r.blockchainInfo)
	if err != nil {
		return nil, fmt.Errorf("internal server error: failed to resolve chain constants: %v", err)
	}
	return &ChainFacts{
		Constants:  constants,
		LastBlock:  nil, // resolved by another lazy resolver
		Aggregated: nil, // TODO: support in explorer DB
	}, nil
}

func rivConstantsAsGQL(cs modules.ConsensusSet, chainConstants *types.ChainConstants, blockchainInfo *types.BlockchainInfo) (*ChainConstants, error) {
	coinPrecision := len(chainConstants.CurrencyUnits.OneCoin.String())
	if coinPrecision > 0 {
		coinPrecision--
	}
	var (
		blockCreatorFee       *BigInt
		minimumTransactionFee *BigInt
	)
	if !chainConstants.BlockCreatorFee.Equals64(0) {
		bi := dbCurrencyAsBigInt(chainConstants.BlockCreatorFee)
		blockCreatorFee = &bi
	}
	if !chainConstants.MinimumTransactionFee.Equals64(0) {
		bi := dbCurrencyAsBigInt(chainConstants.MinimumTransactionFee)
		minimumTransactionFee = &bi
	}
	txFeeBeneficiary, err := dbConditionAsUnlockCondition(chainConstants.TransactionFeeCondition)
	if err != nil {
		return nil, fmt.Errorf("failed to cast transaction fee breneficiary as GQL UnlockCondition: %v", err)
	}
	return &ChainConstants{
		Name:                              blockchainInfo.Name,
		NetworkName:                       blockchainInfo.NetworkName,
		CoinUnit:                          blockchainInfo.CoinUnit,
		CoinPecision:                      coinPrecision,
		ChainVersion:                      blockchainInfo.ChainVersion.String(),
		GatewayProtocolVersion:            blockchainInfo.ProtocolVersion.String(),
		DefaultTransactionVersion:         ByteVersion(chainConstants.DefaultTransactionVersion),
		ConsensusPlugins:                  cs.LoadedPlugins(),
		GenesisTimestamp:                  chainConstants.GenesisTimestamp,
		BlockSizeLimitInBytes:             int(chainConstants.BlockSizeLimit),
		AverageBlockCreationTimeInSeconds: int(chainConstants.BlockFrequency),
		GenesisTotalBlockStakes:           dbCurrencyAsBigInt(chainConstants.GenesisBlockStakeCount()),
		BlockStakeAging:                   int(chainConstants.BlockStakeAging),
		BlockCreatorFee:                   blockCreatorFee,
		MinimumTransactionFee:             minimumTransactionFee,
		TransactionFeeBeneficiary:         txFeeBeneficiary,
		PayoutMaturityDelay:               chainConstants.MaturityDelay,
	}, nil
}

func (r *queryRootResolver) Object(ctx context.Context, id *ObjectID) (Object, error) {
	if id == nil {
		// default to latest block if no ID is given, the only thing that makes sense
		return r.Block(ctx, nil, nil)
	}
	dbObject, err := r.db.GetObject(explorerdb.ObjectID(*id))
	if err != nil {
		return nil, fmt.Errorf("internal DB error while fetching transaction: %v", err)
	}
	switch dbObject.Type {
	case explorerdb.ObjectTypeBlock:
		dbBlock, ok := dbObject.Data.(explorerdb.Block)
		if !ok {
			return nil, fmt.Errorf("internal server error: unexpected type %T for object of type block (%d)", dbObject.Data, dbObject.Type)
		}
		return dbBlockAsGQL(ctx, r.db, &dbBlock)
	case explorerdb.ObjectTypeTransaction:
		dbTransaction, ok := dbObject.Data.(explorerdb.Transaction)
		if !ok {
			return nil, fmt.Errorf("internal server error: unexpected type %T for object of type transaction (%d)", dbObject.Data, dbObject.Type)
		}
		return dbTransactionAsObject(&TransactionParentInfo{
			ID: crypto.Hash(dbTransaction.ParentBlock),
		}, &dbTransaction)
	case explorerdb.ObjectTypeOutput:
		dbOutput, ok := dbObject.Data.(explorerdb.Output)
		if !ok {
			return nil, fmt.Errorf("internal server error: unexpected type %T for object of type output (%d)", dbObject.Data, dbObject.Type)
		}
		return dbOutputAsGQL(&dbOutput, nil)
	case explorerdb.ObjectTypeWallet:
		return nil, fmt.Errorf("internal server error: wallet of object type %d is not yet supported", dbObject.Type)
	case explorerdb.ObjectTypeMultiSignatureWallet:
		return nil, fmt.Errorf("internal server error: multi signature wallet of object type %d is not yet supported", dbObject.Type)
	case explorerdb.ObjectTypeAtomicSwapContract:
		return nil, fmt.Errorf("internal server error: atomic swap contract of object type %d is not yet supported", dbObject.Type)
	default:
		return nil, fmt.Errorf("internal server error: unsupported object type %d", dbObject.Type)
	}
}
func (r *queryRootResolver) Transaction(ctx context.Context, id crypto.Hash) (Transaction, error) {
	transactionID := types.TransactionID(id)
	dbTxn, err := r.db.GetTransaction(transactionID)
	if err != nil {
		return nil, fmt.Errorf("internal DB error while fetching transaction: %v", err)
	}
	return dbTransactionAsGQL(&TransactionParentInfo{
		ID: crypto.Hash(dbTxn.ParentBlock),
	}, &dbTxn)
}
func (r *queryRootResolver) Output(ctx context.Context, id crypto.Hash) (*Output, error) {
	outputID := types.OutputID(id)
	dbOutput, err := r.db.GetOutput(outputID)
	if err != nil {
		return nil, fmt.Errorf("internal DB error while fetching output: %v", err)
	}
	return dbOutputAsGQL(&dbOutput, nil)
}
func (r *queryRootResolver) Block(ctx context.Context, id *crypto.Hash, reference *ReferencePoint) (*Block, error) {
	if id == nil {
		return blockByReferencePoint(ctx, r.db, reference)
	}
	blockID := types.BlockID(*id)
	dbBlock, err := r.db.GetBlock(blockID)
	if err != nil {
		return nil, fmt.Errorf("internal DB error while fetching block: %v", err)
	}
	if reference != nil {
		// TODO: should we handle this as some kind of warning instead of an error???
		if reference.IsBlockHeight() {
			if dbBlock.Height != types.BlockHeight(*reference) {
				return nil, fmt.Errorf("block found but it has an unexpected block height of %d, not %d", dbBlock.Height, *reference)
			}
		} else {
			if dbBlock.Timestamp != types.Timestamp(*reference) {
				return nil, fmt.Errorf("block found but it has an unexpected block timestamp of %d, not %d", dbBlock.Timestamp, *reference)
			}
		}
	}
	return dbBlockAsGQL(ctx, r.db, &dbBlock)
}
func blockByReferencePoint(ctx context.Context, db explorerdb.DB, reference *ReferencePoint) (*Block, error) {
	if reference == nil {
		// default to latest block
		chainCtx, err := db.GetChainContext()
		if err != nil {
			return nil, err
		}
		r := ReferencePoint(chainCtx.Height)
		if r > 0 {
			r-- // chainCtx.Height defines the amount of blocks (in other words, the height of the chain), not the height of latest
		}
		reference = &r
	}
	dbBlock, err := db.GetBlockByReferencePoint(explorerdb.ReferencePoint(*reference))
	if err != nil {
		return nil, fmt.Errorf("internal DB error while fetching block: %v", err)
	}
	return dbBlockAsGQL(ctx, db, &dbBlock)
}

func dbBlockAsGQL(ctx context.Context, db explorerdb.DB, dbBlock *explorerdb.Block) (*Block, error) {
	// assemble block header
	header := &BlockHeader{
		ID:          crypto.Hash(dbBlock.ID),
		BlockTime:   &dbBlock.Timestamp,
		BlockHeight: &dbBlock.Height,
	}
	if dbBlock.ParentID != (types.BlockID{}) {
		h := crypto.Hash(dbBlock.ParentID)
		header.ParentID = &h
	}
	header.Payouts = make([]*BlockPayout, 0, len(dbBlock.Payouts))
	for _, payoutID := range dbBlock.Payouts {
		header.Payouts = append(header.Payouts, &BlockPayout{
			Output: &Output{
				ID: crypto.Hash(payoutID),
				// rest of fields are populated using the separate blockHeaderResolver::Payouts resolver
			},
			Type: nil, // populated using the separate blockHeaderResolver::Payouts resolver
		})
	}

	// nil // populated using the separate blockHeaderResolver::Payouts resolver
	// assemble transactions with ID only
	transactions := make([]Transaction, 0, len(dbBlock.Transactions))
	for _, txnID := range dbBlock.Transactions {
		// populated with real types using resolver
		transactions = append(transactions, &StubTransaction{
			ID: crypto.Hash(txnID),
		})
	}
	return &Block{
		Header:       header,
		Transactions: transactions,
	}, nil
}

func (r *queryRootResolver) Blocks(ctx context.Context, after *ReferencePoint, first *int, before *ReferencePoint, last *int) (Transaction, error) {
	panic("not implemented")
}
func (r *queryRootResolver) Wallet(ctx context.Context, unlockhash types.UnlockHash) (Wallet, error) {
	panic("not implemented")
}
func (r *queryRootResolver) Contract(ctx context.Context, unlockhash types.UnlockHash) (Contract, error) {
	panic("not implemented")
}

type standardTransactionResolver struct{ *Resolver }

func (r *standardTransactionResolver) ParentBlock(ctx context.Context, obj *StandardTransaction) (*TransactionParentInfo, error) {
	return dbFillTransactionParentInfoForTxn(ctx, r.db, obj.ParentBlock, types.TransactionID(obj.ID))
}

func dbFillTransactionParentInfoForTxn(ctx context.Context, db explorerdb.DB, existingParentInfo *TransactionParentInfo, ownerTxnID types.TransactionID) (*TransactionParentInfo, error) {
	if existingParentInfo == nil {
		return nil, nil
	}
	if existingParentInfo.Height != nil && existingParentInfo.Timestamp != nil && existingParentInfo.TransactionOrder != nil {
		return existingParentInfo, nil
	}
	parentBlock, err := db.GetBlock(types.BlockID(existingParentInfo.ID))
	if err != nil {
		return nil, err
	}
	parentInfo := &TransactionParentInfo{
		ID:        existingParentInfo.ID,
		ParentID:  dbBlockIDAsHash(parentBlock.ParentID),
		Height:    &parentBlock.Height,
		Timestamp: &parentBlock.Timestamp,
	}
	existingParentInfo.Height = &parentBlock.Height
	existingParentInfo.Timestamp = &parentBlock.Timestamp
	for idx, txnID := range parentBlock.Transactions {
		if bytes.Compare(txnID[:], ownerTxnID[:]) == 0 {
			index := idx
			parentInfo.TransactionOrder = &index
			break
		}
	}
	if parentInfo.TransactionOrder == nil {
		return nil, fmt.Errorf("failed to find transaction order for txn %s, while it does have a parent", ownerTxnID.String())
	}
	parentInfo.SiblingTransactions = allTransactionIdentifiersExceptAsGQLTransactions(parentBlock.Transactions, *parentInfo.TransactionOrder)
	return parentInfo, nil
}

func allTransactionIdentifiersExceptAsGQLTransactions(txns []types.TransactionID, ignoreIndex int) []Transaction {
	ltxns := len(txns)
	if ltxns == 0 {
		return nil
	}
	ntxns := make([]Transaction, 0, ltxns-1)
	for idx, txnID := range txns {
		if idx != ignoreIndex {
			ntxns = append(ntxns, &StubTransaction{
				ID: crypto.Hash(txnID),
			})
		}
	}
	return ntxns
}

func (r *standardTransactionResolver) CoinInputs(ctx context.Context, obj *StandardTransaction) ([]*Input, error) {
	return resolveCoinInputs(ctx, obj.CoinInputs, r.db)
}
func (r *standardTransactionResolver) CoinOutputs(ctx context.Context, obj *StandardTransaction) ([]*Output, error) {
	return resolveCoinOutputs(ctx, obj.CoinOutputs, r.db)
}
func (r *standardTransactionResolver) BlockStakeInputs(ctx context.Context, obj *StandardTransaction) ([]*Input, error) {
	return resolveBlockStakeInputs(ctx, obj.BlockStakeInputs, r.db)
}
func (r *standardTransactionResolver) BlockStakeOutputs(ctx context.Context, obj *StandardTransaction) ([]*Output, error) {
	return resolveBlockStakeOutputs(ctx, obj.BlockStakeOutputs, r.db)
}

func resolveCoinOutputs(ctx context.Context, outputIDs []*Output, db explorerdb.DB) ([]*Output, error) {
	return resolveOutputs(ctx, outputIDs, db, explorerdb.OutputTypeCoin)
}

func resolveBlockStakeOutputs(ctx context.Context, outputIDs []*Output, db explorerdb.DB) ([]*Output, error) {
	return resolveOutputs(ctx, outputIDs, db, explorerdb.OutputTypeBlockStake)
}

func resolveOutputs(ctx context.Context, outputIDs []*Output, db explorerdb.DB, expectedOutputType explorerdb.OutputType) ([]*Output, error) {
	outputs := make([]*Output, 0, len(outputIDs))
	var (
		err       error
		output    explorerdb.Output
		outputGQL *Output
	)
	for cidx, ci := range outputIDs {
		output, err = db.GetOutput(types.OutputID(ci.ID))
		if err != nil {
			return nil, err
		}
		if output.Type != expectedOutputType {
			return nil, fmt.Errorf("output #%d %s is of an unexpected type %d, expected type %d", cidx+1, ci.ID.String(), output.Type, expectedOutputType)
		}
		outputGQL, err = dbOutputAsGQL(&output, nil)
		if err != nil {
			return nil, err
		}
		outputs = append(outputs, outputGQL)
	}
	return outputs, nil
}

func resolveCoinInputs(ctx context.Context, inputIDs []*Input, db explorerdb.DB) ([]*Input, error) {
	return resolveInputs(ctx, inputIDs, db, explorerdb.OutputTypeCoin)
}

func resolveBlockStakeInputs(ctx context.Context, inputIDs []*Input, db explorerdb.DB) ([]*Input, error) {
	return resolveInputs(ctx, inputIDs, db, explorerdb.OutputTypeBlockStake)
}

func resolveInputs(ctx context.Context, inputIDs []*Input, db explorerdb.DB, expectedOutputType explorerdb.OutputType) ([]*Input, error) {
	inputs := make([]*Input, 0, len(inputIDs))
	var (
		err    error
		output explorerdb.Output
		input  *Input
	)
	for cidx, ci := range inputIDs {
		output, err = db.GetOutput(types.OutputID(ci.ID))
		if err != nil {
			return nil, err
		}
		if output.Type != expectedOutputType {
			return nil, fmt.Errorf("output #%d %s (used as input) is of an unexpected type %d, expected type %d", cidx+1, ci.ID.String(), output.Type, expectedOutputType)
		}
		input, err = dbOutputAsInputGQL(&output, nil)
		if err != nil {
			return nil, err
		}
		inputs = append(inputs, input)
	}
	return inputs, nil
}

type transactionFeePayoutResolver struct{ *Resolver }

func (r *transactionFeePayoutResolver) BlockPayout(ctx context.Context, obj *TransactionFeePayout) (*BlockPayout, error) {
	dbBlockPayout, err := r.db.GetOutput(types.OutputID(obj.BlockPayout.Output.ID))
	if err != nil {
		return nil, err
	}
	output, err := dbOutputAsGQL(&dbBlockPayout, nil)
	if err != nil {
		return nil, err
	}
	return &BlockPayout{
		Output: output,
		Type:   dbOutputTypeAsBlockPayoutType(dbBlockPayout.Type),
	}, nil
}

type transactionParentInfoResolver struct{ *Resolver }

func (r *transactionParentInfoResolver) SiblingTransactions(ctx context.Context, obj *TransactionParentInfo) ([]Transaction, error) {
	ltxns := len(obj.SiblingTransactions)
	txns := make([]Transaction, 0, ltxns)
	txnParents := make([]*TransactionParentInfo, 0, ltxns)
	for _, txn := range obj.SiblingTransactions {
		stxn := txn.(*StubTransaction)
		dbTxn, err := r.db.GetTransaction(types.TransactionID(stxn.ID))
		if err != nil {
			return nil, err
		}
		txnParent := &TransactionParentInfo{
			ID:               obj.ID,
			ParentID:         obj.ParentID,
			Height:           obj.Height,
			Timestamp:        obj.Timestamp,
			TransactionOrder: obj.TransactionOrder,
		}
		txn, err := dbTransactionAsGQL(txnParent, &dbTxn)
		if err != nil {
			return nil, err
		}
		txns = append(txns, txn)
		txnParents = append(txnParents, txnParent)
	}
	for idx := range obj.SiblingTransactions {
		txnParents[idx].SiblingTransactions = allTransactionsExcept(txns, idx)
	}
	return txns, nil
}

type unlockHashConditionResolver struct{ *Resolver }

func (r *unlockHashConditionResolver) PublicKey(ctx context.Context, obj *UnlockHashCondition) (*types.PublicKey, error) {
	pk, err := r.db.GetPublicKey(obj.UnlockHash)
	if err != nil {
		return nil, err
	}
	return &pk, nil
}

type unlockHashPublicKeyPairResolver struct{ *Resolver }

func (r *unlockHashPublicKeyPairResolver) PublicKey(ctx context.Context, obj *UnlockHashPublicKeyPair) (*types.PublicKey, error) {
	pk, err := r.db.GetPublicKey(obj.UnlockHash)
	if err != nil {
		return nil, err
	}
	return &pk, nil
}

type StubTransaction struct {
	ID crypto.Hash `json:"ID"`
}

func (st *StubTransaction) IsTransaction() {}

func dbTransactionAsGQL(parentBlockInfo *TransactionParentInfo, dbTxn *explorerdb.Transaction) (Transaction, error) {
	switch dbTxn.Version {
	case types.TransactionVersionOne, types.TransactionVersionZero:
		return dbTransactionAsStandardTransaction(parentBlockInfo, dbTxn)
	// TODO: do not hardcode these versions
	// TODO: how to support non standard transactions???
	case 128:
		return dbMintConditionDefinitionTransactionAsGQL(parentBlockInfo, dbTxn)
	case 129:
		return dbMintCoinCreationTransactionAsGQL(parentBlockInfo, dbTxn)
	case 130:
		return dbMintCoinDestructionTransactionAsGQL(parentBlockInfo, dbTxn)
	default:
		// TODO: support extensions and render unknowns as an except unknown transaction schema type
		return nil, fmt.Errorf("unsupported transaction version %d", dbTxn.Version)
	}
}

// TODO: is there a way to cast a Transaction directly do an object, without this duplication???
func dbTransactionAsObject(parentBlockInfo *TransactionParentInfo, dbTxn *explorerdb.Transaction) (Object, error) {
	switch dbTxn.Version {
	case types.TransactionVersionOne, types.TransactionVersionZero:
		return dbTransactionAsStandardTransaction(parentBlockInfo, dbTxn)
	// TODO: do not hardcode these versions
	// TODO: how to support non standard transactions???
	case 128:
		return dbMintConditionDefinitionTransactionAsGQL(parentBlockInfo, dbTxn)
	case 129:
		return dbMintCoinCreationTransactionAsGQL(parentBlockInfo, dbTxn)
	case 130:
		return dbMintCoinDestructionTransactionAsGQL(parentBlockInfo, dbTxn)
	default:
		// TODO: support extensions and render unknowns as an except unknown transaction schema type
		return nil, fmt.Errorf("unsupported transaction version %d", dbTxn.Version)
	}
}

// TODO: is there a way to cast a Transaction directly do an outputParent, without this duplication???
func dbTransactionAsOutputParent(parentBlockInfo *TransactionParentInfo, dbTxn *explorerdb.Transaction) (OutputParent, error) {
	switch dbTxn.Version {
	case types.TransactionVersionOne, types.TransactionVersionZero:
		return dbTransactionAsStandardTransaction(parentBlockInfo, dbTxn)
	// TODO: do not hardcode these versions
	// TODO: how to support non standard transactions???
	case 128:
		return dbMintConditionDefinitionTransactionAsGQL(parentBlockInfo, dbTxn)
	case 129:
		return dbMintCoinCreationTransactionAsGQL(parentBlockInfo, dbTxn)
	case 130:
		return dbMintCoinDestructionTransactionAsGQL(parentBlockInfo, dbTxn)
	default:
		// TODO: support extensions and render unknowns as an except unknown transaction schema type
		return nil, fmt.Errorf("unsupported transaction version %d", dbTxn.Version)
	}
}

func dbTransactionAsStandardTransaction(txParentInfo *TransactionParentInfo, dbTxn *explorerdb.Transaction) (*StandardTransaction, error) {
	var (
		outputID          types.OutputID
		coinInputs        []*Input
		coinOutputs       []*Output
		blockStakeInputs  []*Input
		blockStakeOutputs []*Output
	)
	for _, outputID = range dbTxn.CoinInputs {
		coinInputs = append(coinInputs, &Input{
			ID: crypto.Hash(outputID),
			// other fields are done by another lazy resolver
		})
	}
	for _, outputID = range dbTxn.CoinOutputs {
		coinOutputs = append(coinOutputs, &Output{
			ID: crypto.Hash(outputID),
			// other fields are done by another lazy resolver
		})
	}
	for _, outputID = range dbTxn.BlockStakeInputs {
		blockStakeInputs = append(blockStakeInputs, &Input{
			ID: crypto.Hash(outputID),
			// other fields are done by another lazy resolver
		})
	}
	for _, outputID = range dbTxn.BlockStakeOutputs {
		blockStakeOutputs = append(blockStakeOutputs, &Output{
			ID: crypto.Hash(outputID),
			// other fields are done by another lazy resolver
		})
	}

	return &StandardTransaction{
		ID:                crypto.Hash(dbTxn.ID),
		Version:           ByteVersion(dbTxn.Version),
		ParentBlock:       txParentInfo,
		CoinInputs:        coinInputs,
		CoinOutputs:       coinOutputs,
		BlockStakeInputs:  blockStakeInputs,
		BlockStakeOutputs: blockStakeOutputs,
		FeePayouts:        dbTxFeePayoutInfoAsGQL(&dbTxn.FeePayout),
		ArbitraryData:     dbByteSliceAsBinaryData(dbTxn.ArbitraryData),
	}, nil
}

func dbMintConditionDefinitionTransactionAsGQL(txParentInfo *TransactionParentInfo, dbTxn *explorerdb.Transaction) (*MintConditionDefinitionTransaction, error) {
	var (
		outputID    types.OutputID
		coinInputs  []*Input
		coinOutputs []*Output
	)
	for _, outputID = range dbTxn.CoinInputs {
		coinInputs = append(coinInputs, &Input{
			ID: crypto.Hash(outputID),
			// other fields are done by another lazy resolver
		})
	}
	for _, outputID = range dbTxn.CoinOutputs {
		coinOutputs = append(coinOutputs, &Output{
			ID: crypto.Hash(outputID),
			// other fields are done by another lazy resolver
		})
	}
	var mcdtxExtensionData minting.MinterDefinitionTransactionExtension
	err := rivbin.Unmarshal(dbTxn.EncodedExtensionData, &mcdtxExtensionData)
	if err != nil {
		return nil, err
	}
	mintCondition, err := dbConditionAsUnlockCondition(mcdtxExtensionData.MintCondition)
	if err != nil {
		return nil, err
	}
	mintFulfillment, err := dbFulfillmentAsUnlockFulfillment(mcdtxExtensionData.MintFulfillment, nil)
	if err != nil {
		return nil, err
	}
	return &MintConditionDefinitionTransaction{
		ID:               crypto.Hash(dbTxn.ID),
		Version:          ByteVersion(dbTxn.Version),
		ParentBlock:      txParentInfo,
		Nonce:            *dbByteSliceAsBinaryData(mcdtxExtensionData.Nonce[:]),
		NewMintCondition: mintCondition,
		MintFulfillment:  mintFulfillment,
		CoinInputs:       coinInputs,
		CoinOutputs:      coinOutputs,
		FeePayouts:       dbTxFeePayoutInfoAsGQL(&dbTxn.FeePayout),
		ArbitraryData:    dbByteSliceAsBinaryData(dbTxn.ArbitraryData),
	}, nil
}

func dbMintCoinCreationTransactionAsGQL(txParentInfo *TransactionParentInfo, dbTxn *explorerdb.Transaction) (*MintCoinCreationTransaction, error) {
	var (
		outputID    types.OutputID
		coinInputs  []*Input
		coinOutputs []*Output
	)
	for _, outputID = range dbTxn.CoinInputs {
		coinInputs = append(coinInputs, &Input{
			ID: crypto.Hash(outputID),
			// other fields are done by another lazy resolver
		})
	}
	for _, outputID = range dbTxn.CoinOutputs {
		coinOutputs = append(coinOutputs, &Output{
			ID: crypto.Hash(outputID),
			// other fields are done by another lazy resolver
		})
	}
	var mcctxExtensionData minting.CoinCreationTransactionExtension
	err := rivbin.Unmarshal(dbTxn.EncodedExtensionData, &mcctxExtensionData)
	if err != nil {
		return nil, err
	}
	mintFulfillment, err := dbFulfillmentAsUnlockFulfillment(mcctxExtensionData.MintFulfillment, nil)
	if err != nil {
		return nil, err
	}
	return &MintCoinCreationTransaction{
		ID:              crypto.Hash(dbTxn.ID),
		Version:         ByteVersion(dbTxn.Version),
		ParentBlock:     txParentInfo,
		Nonce:           *dbByteSliceAsBinaryData(mcctxExtensionData.Nonce[:]),
		MintFulfillment: mintFulfillment,
		CoinInputs:      coinInputs,
		CoinOutputs:     coinOutputs,
		FeePayouts:      dbTxFeePayoutInfoAsGQL(&dbTxn.FeePayout),
		ArbitraryData:   dbByteSliceAsBinaryData(dbTxn.ArbitraryData),
	}, nil
}

func dbMintCoinDestructionTransactionAsGQL(txParentInfo *TransactionParentInfo, dbTxn *explorerdb.Transaction) (*MintCoinDestructionTransaction, error) {
	var (
		outputID    types.OutputID
		coinInputs  []*Input
		coinOutputs []*Output
	)
	for _, outputID = range dbTxn.CoinInputs {
		coinInputs = append(coinInputs, &Input{
			ID: crypto.Hash(outputID),
			// other fields are done by another lazy resolver
		})
	}
	for _, outputID = range dbTxn.CoinOutputs {
		coinOutputs = append(coinOutputs, &Output{
			ID: crypto.Hash(outputID),
			// other fields are done by another lazy resolver
		})
	}
	return &MintCoinDestructionTransaction{
		ID:            crypto.Hash(dbTxn.ID),
		Version:       ByteVersion(dbTxn.Version),
		ParentBlock:   txParentInfo,
		CoinInputs:    coinInputs,
		CoinOutputs:   coinOutputs,
		FeePayouts:    dbTxFeePayoutInfoAsGQL(&dbTxn.FeePayout),
		ArbitraryData: dbByteSliceAsBinaryData(dbTxn.ArbitraryData),
	}, nil
}

func dbTxFeePayoutInfoAsGQL(fpInfo *explorerdb.TransactionFeePayoutInfo) []*TransactionFeePayout {
	payouts := make([]*TransactionFeePayout, 0, len(fpInfo.Values))
	blockPayoutType := BlockPayoutTypeTransactionFee
	blockPayout := &BlockPayout{
		Output: &Output{
			ID: crypto.Hash(fpInfo.PayoutID),
			// other fields are done by another lazy resolver
		},
		Type: &blockPayoutType,
	}
	for _, v := range fpInfo.Values {
		payouts = append(payouts, &TransactionFeePayout{
			BlockPayout: blockPayout,
			Value:       dbCurrencyAsBigInt(v),
		})
	}
	return payouts
}

// TODO: Support ParentID and FulfillmentTxID in GraphQL schema

func dbOutputAsGQL(output *explorerdb.Output, sibling *Input) (*Output, error) {
	gqlCondition, err := dbConditionAsUnlockCondition(output.Condition)
	if err != nil {
		return nil, err
	}
	gqlOutput := &Output{
		ID:        crypto.Hash(output.ID),
		Type:      dbOutputTypeAsGQL(output.Type),
		Value:     dbCurrencyAsBigInt(output.Value),
		Condition: gqlCondition,
		ParentID:  output.ParentID,
	}
	if sibling != nil {
		gqlOutput.ChildInput = sibling
	} else if output.SpenditureData != nil {
		ff, err := dbFulfillmentAsUnlockFulfillment(output.SpenditureData.Fulfillment, gqlCondition)
		if err != nil {
			return nil, err
		}
		gqlOutput.ChildInput = &Input{
			ID:           gqlOutput.ID,
			Value:        gqlOutput.Value,
			Fulfillment:  ff,
			ParentOutput: gqlOutput,
		}
	}
	return gqlOutput, nil
}

func dbOutputAsInputGQL(output *explorerdb.Output, parent *Output) (*Input, error) {
	if output.SpenditureData == nil {
		return nil, fmt.Errorf("spenditure data of output %s is not defined and can as such not be used as a GQL input", output.ID.String())
	}
	var parentCondition UnlockCondition
	if parent != nil {
		parentCondition = parent.Condition
	}
	ff, err := dbFulfillmentAsUnlockFulfillment(output.SpenditureData.Fulfillment, parentCondition)
	if err != nil {
		return nil, err
	}
	glqInput := &Input{
		ID:          crypto.Hash(output.ID),
		Type:        dbOutputTypeAsGQL(output.Type),
		Value:       dbCurrencyAsBigInt(output.Value),
		Fulfillment: ff,
	}
	if parent == nil {
		glqInput.ParentOutput, err = dbOutputAsGQL(output, glqInput)
		if err != nil {
			return nil, err
		}
	} else {
		parent.ChildInput = glqInput
		glqInput.ParentOutput = parent
	}
	return glqInput, nil
}

func dbOutputTypeAsGQL(outputType explorerdb.OutputType) *OutputType {
	switch outputType {
	case explorerdb.OutputTypeCoin:
		ot := OutputTypeCoin
		return &ot
	case explorerdb.OutputTypeBlockStake:
		ot := OutputTypeBlockStake
		return &ot
	case explorerdb.OutputTypeBlockCreationReward:
		ot := OutputTypeBlockCreationReward
		return &ot
	case explorerdb.OutputTypeTransactionFee:
		ot := OutputTypeTransactionFee
		return &ot
	default:
		return nil
	}
}

func dbFulfillmentAsUnlockFulfillment(fulfillment types.UnlockFulfillmentProxy, parentCondition UnlockCondition) (UnlockFulfillment, error) {
	switch ft := fulfillment.FulfillmentType(); ft {
	case types.FulfillmentTypeSingleSignature:
		sft := fulfillment.Fulfillment.(*types.SingleSignatureFulfillment)
		return &SingleSignatureFulfillment{
			Version:         ByteVersion(ft),
			ParentCondition: parentCondition,
			PublicKey:       sft.PublicKey,
			Signature:       Signature(sft.Signature[:]),
		}, nil
	case types.FulfillmentTypeMultiSignature:
		msft := fulfillment.Fulfillment.(*types.MultiSignatureFulfillment)
		pairs := make([]*PublicKeySignaturePair, 0, len(msft.Pairs))
		for _, pair := range msft.Pairs {
			pairs = append(pairs, &PublicKeySignaturePair{
				PublicKey: pair.PublicKey,
				Signature: Signature(pair.Signature[:]),
			})
		}
		return &MultiSignatureFulfillment{
			Version:         ByteVersion(ft),
			ParentCondition: parentCondition,
			Pairs:           pairs,
		}, nil
	case types.FulfillmentTypeAtomicSwap:
		asft := fulfillment.Fulfillment.(*types.AtomicSwapFulfillment)
		asfGQL := &AtomicSwapFulfillment{
			Version:         ByteVersion(ft),
			ParentCondition: parentCondition,
			PublicKey:       asft.PublicKey,
			Signature:       Signature(asft.Signature[:]),
		}
		if asft.Secret != (types.AtomicSwapSecret{}) {
			asfGQL.Secret = dbByteSliceAsBinaryData(asft.Secret[:])
		}
		return asfGQL, nil
	default:
		return nil, fmt.Errorf("unsupported fulfillment type %d: %v", ft, fulfillment)
	}
}

func dbConditionAsUnlockCondition(condition types.UnlockConditionProxy) (UnlockCondition, error) {
	switch ct := condition.ConditionType(); ct {
	case types.ConditionTypeNil:
		return dbNilConditionAsGQL(condition.Condition.(*types.NilCondition)), nil
	case types.ConditionTypeUnlockHash:
		return dbUnlockHashConditionAsGQL(condition.Condition.(*types.UnlockHashCondition)), nil
	case types.ConditionTypeAtomicSwap:
		asc := condition.Condition.(*types.AtomicSwapCondition)
		return &AtomicSwapCondition{
			Version:      ByteVersion(ct),
			UnlockHash:   condition.UnlockHash(),
			Sender:       dbUnlockHashAsUnlockHashPublicKeyPair(asc.Sender),
			Receiver:     dbUnlockHashAsUnlockHashPublicKeyPair(asc.Receiver),
			HashedSecret: *dbByteSliceAsBinaryData(asc.HashedSecret[:]),
			TimeLock:     LockTime(asc.TimeLock),
		}, nil
	case types.ConditionTypeTimeLock:
		tlc := condition.Condition.(*types.TimeLockCondition)
		lt := LockTypeTimestamp
		if tlc.LockTime < types.LockTimeMinTimestampValue {
			lt = LockTypeBlockHeight
		}
		uh := condition.UnlockHash()
		ltc := &LockTimeCondition{
			Version:    ByteVersion(ct),
			UnlockHash: &uh,
			LockValue:  LockTime(tlc.LockTime),
			LockType:   lt,
		}
		switch ict := tlc.Condition.ConditionType(); ict {
		case types.ConditionTypeNil:
			ltc.Condition = dbNilConditionAsGQL(tlc.Condition.(*types.NilCondition))
		case types.ConditionTypeUnlockHash:
			ltc.Condition = dbUnlockHashConditionAsGQL(tlc.Condition.(*types.UnlockHashCondition))
		case types.ConditionTypeMultiSignature:
			ltc.Condition = dbMultiSignatureConditionAsGQL(tlc.Condition.(*types.MultiSignatureCondition))
		default:
			return nil, fmt.Errorf("unsupported inner LockTime condition type %d: %v", ict, tlc.Condition)
		}
		return ltc, nil
	case types.ConditionTypeMultiSignature:
		return dbMultiSignatureConditionAsGQL(condition.Condition.(*types.MultiSignatureCondition)), nil
	default:
		return nil, fmt.Errorf("unsupported condition type %d: %v", ct, condition)
	}
}

func dbNilConditionAsGQL(condition *types.NilCondition) *NilCondition {
	return &NilCondition{
		Version:    ByteVersion(condition.ConditionType()),
		UnlockHash: types.NilUnlockHash,
	}
}

func dbUnlockHashConditionAsGQL(condition *types.UnlockHashCondition) *UnlockHashCondition {
	return &UnlockHashCondition{
		Version:    ByteVersion(condition.ConditionType()),
		UnlockHash: condition.UnlockHash(),
	}
}

func dbMultiSignatureConditionAsGQL(condition *types.MultiSignatureCondition) *MultiSignatureCondition {
	owners := make([]*UnlockHashPublicKeyPair, 0, len(condition.UnlockHashes))
	for i := range condition.UnlockHashes {
		owners = append(owners, dbUnlockHashAsUnlockHashPublicKeyPair(condition.UnlockHashes[i]))
	}
	return &MultiSignatureCondition{
		Version:                ByteVersion(condition.ConditionType()),
		UnlockHash:             condition.UnlockHash(),
		Owners:                 owners,
		RequiredSignatureCount: int(condition.MinimumSignatureCount),
	}
}

func dbUnlockHashAsUnlockHashPublicKeyPair(uh types.UnlockHash) *UnlockHashPublicKeyPair {
	return &UnlockHashPublicKeyPair{
		UnlockHash: uh,
		PublicKey:  nil, // field is populated by another lazy resolver
	}
}

func dbOutputTypeAsBlockPayoutType(outputType explorerdb.OutputType) *BlockPayoutType {
	switch outputType {
	case explorerdb.OutputTypeBlockCreationReward:
		bpt := BlockPayoutTypeBlockReward
		return &bpt
	case explorerdb.OutputTypeTransactionFee:
		bpt := BlockPayoutTypeTransactionFee
		return &bpt
	default:
		return nil // == unknown
	}
}

func dbOutputIDAsHash(outputID types.OutputID) *crypto.Hash {
	h := crypto.Hash(outputID)
	return &h
}

func dbBlockIDAsHash(blockID types.BlockID) *crypto.Hash {
	h := crypto.Hash(blockID)
	return &h
}

func dbCurrencyAsBigInt(c types.Currency) BigInt {
	return BigInt{
		Int: c.Big(),
	}
}

func dbByteSliceAsBinaryData(b []byte) *BinaryData {
	bd := BinaryData(b)
	return &bd
}
