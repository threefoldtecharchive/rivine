package explorergraphql

import (
	"context"
	"fmt"
	"sync"

	"github.com/threefoldtech/rivine/crypto"
	"github.com/threefoldtech/rivine/modules/explorergraphql/explorerdb"
	"github.com/threefoldtech/rivine/types"
)

type (
	blockData struct {
		Header       *BlockHeader
		Facts        *BlockFacts
		Transactions []Transaction
	}

	Block struct {
		id types.BlockID
		db explorerdb.DB

		onceData sync.Once
		data     *blockData
		dataErr  error
	}
)

// compile-time interface check
var (
	_ Object       = (*Block)(nil)
	_ OutputParent = (*Block)(nil)
)

func NewBlock(id types.BlockID, db explorerdb.DB) *Block {
	return &Block{
		id: id,
		db: db,
	}
}

func (block *Block) blockData(ctx context.Context) (*blockData, error) {
	block.onceData.Do(block._blockDataOnce)
	return block.data, block.dataErr
}

func (block *Block) _blockDataOnce() {
	data, err := block.db.GetBlock(block.id)
	if err != nil {
		block.dataErr = fmt.Errorf("failed to fetch block %s data from DB: %v", block.id.String(), err)
		return
	}

	// restructure all fetched data...

	// ... assemble block header
	header := &BlockHeader{
		ID:          crypto.Hash(data.ID),
		BlockTime:   &data.Timestamp,
		BlockHeight: &data.Height,
	}
	if data.ParentID != (types.BlockID{}) {
		h := crypto.Hash(data.ParentID)
		header.ParentID = &h
		header.Parent = NewBlock(data.ParentID, block.db)
	}
	header.Payouts = make([]*BlockPayout, 0, len(data.Payouts))
	for _, payoutID := range data.Payouts {
		header.Payouts = append(header.Payouts, NewBlockPayout(payoutID, block, block.db))
	}

	// ... assemble transactions
	transactions := make([]Transaction, 0, len(data.Transactions))
	for _, txnID := range data.Transactions {
		txn, err := NewTransaction(txnID, NewTransactionParentInfoForBlock(txnID, block), block.db)
		if err != nil {
			block.dataErr = fmt.Errorf("failed to convert block %s data from DB: failed to create txn %s resolver: %v", block.id.String(), err)
			return
		}
		transactions = append(transactions, txn)
	}

	// ... finally we can put it all together
	block.data = &blockData{
		Header:       header,
		Facts:        nil, // facts are resolved in a lazy manner when required
		Transactions: transactions,
	}
}

// IsObject implements the GraphQL Object interface
func (block *Block) IsObject() {}

// IsOutputParent implements the GraphQL OutputParent interface
func (block *Block) IsOutputParent() {}

func (block *Block) Header(ctx context.Context) (*BlockHeader, error) {
	data, err := block.blockData(ctx)
	if err != nil {
		return nil, err
	}
	return data.Header, nil
}

func (block *Block) Facts(ctx context.Context) (*BlockFacts, error) {
	dbBlockFacts, err := block.db.GetBlockFacts(block.id)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch block %s facts data from DB: %v", block.id.String(), err)
	}
	return dbBlockFactsAsGQL(&dbBlockFacts), nil
}

func (block *Block) Transactions(ctx context.Context) ([]Transaction, error) {
	data, err := block.blockData(ctx)
	if err != nil {
		return nil, err
	}
	return data.Transactions, nil
}

func dbBlockFactsAsGQL(dbBlockFacts *explorerdb.BlockFacts) *BlockFacts {
	return &BlockFacts{
		Difficulty: dbBigIntAsGQLRef(dbBlockFacts.Constants.Difficulty.Big()),
		Target:     dbTargetAsHash(dbBlockFacts.Constants.Target),
		ChainSnapshot: &BlockChainSnapshotFacts{
			TotalCoins:                 dbCurrencyAsBigIntRef(dbBlockFacts.Aggregated.TotalCoins),
			TotalLockedCoins:           dbCurrencyAsBigIntRef(dbBlockFacts.Aggregated.TotalLockedCoins),
			TotalBlockStakes:           dbCurrencyAsBigIntRef(dbBlockFacts.Aggregated.TotalBlockStakes),
			TotalLockedBlockStakes:     dbCurrencyAsBigIntRef(dbBlockFacts.Aggregated.TotalLockedBlockStakes),
			EstimatedActiveBlockStakes: dbCurrencyAsBigIntRef(dbBlockFacts.Aggregated.EstimatedActiveBlockStakes),
		},
	}
}

type (
	BlockPayout struct {
		output *Output
	}
)

func NewBlockPayout(id types.OutputID, parent OutputParent, db explorerdb.DB) *BlockPayout {
	return &BlockPayout{
		output: NewOutput(id, nil, parent, db),
	}
}

func (payout *BlockPayout) Output(context.Context) (*Output, error) {
	return payout.output, nil
}

func (payout *BlockPayout) Type(ctx context.Context) (*BlockPayoutType, error) {
	ot, err := payout.output.Type(ctx)
	if err != nil {
		return nil, err
	}
	return outputTypeAsBlockPayoutType(ot), nil
}

func outputTypeAsBlockPayoutType(outputType *OutputType) *BlockPayoutType {
	switch *outputType {
	case OutputTypeBlockCreationReward:
		bpt := BlockPayoutTypeBlockReward
		return &bpt
	case OutputTypeTransactionFee:
		bpt := BlockPayoutTypeTransactionFee
		return &bpt
	default:
		return nil // == unknown
	}
}
