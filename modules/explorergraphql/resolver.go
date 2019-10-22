package explorergraphql

import (
	"context"
	"errors"
	"fmt"

	"github.com/threefoldtech/rivine/crypto"
	"github.com/threefoldtech/rivine/modules/explorergraphql/explorerdb"
	"github.com/threefoldtech/rivine/types"
)

// THIS CODE IS A STARTING POINT ONLY. IT WILL NOT BE UPDATED WITH SCHEMA CHANGES.

type Resolver struct {
	db explorerdb.DB
}

func (r *Resolver) QueryRoot() QueryRootResolver {
	return &queryRootResolver{r}
}

type queryRootResolver struct{ *Resolver }

func (r *queryRootResolver) Object(ctx context.Context, id *BinaryData) (Object, error) {
	panic("not implemented")
}
func (r *queryRootResolver) Transaction(ctx context.Context, id *crypto.Hash) (Transaction, error) {
	panic("not implemented")
}
func (r *queryRootResolver) Transactions(ctx context.Context, after *ReferencePoint, first *int, before *ReferencePoint, last *int) (Transaction, error) {
	panic("not implemented")
}
func (r *queryRootResolver) Block(ctx context.Context, id *crypto.Hash, reference *ReferencePoint) (*Block, error) {
	if id == nil {
		return r.blockByReferencePoint(ctx, reference)
	}
	blockID := types.BlockID(*id)
	dbBlock, err := r.db.GetBlock(blockID)
	if err != nil {
		return nil, fmt.Errorf("internal DB error while fetching block: %v", err)
	}
	// assemble block header
	header := &BlockHeader{
		ID:          *id,
		BlockTime:   &dbBlock.Timestamp,
		BlockHeight: &dbBlock.Height,
	}
	if dbBlock.ParentID != (types.BlockID{}) {
		h := crypto.Hash(dbBlock.ParentID)
		header.ParentID = &h
	}
	header.Payouts = nil // TODO
	/*make([]*BlockPayout, 0, len(dbBlock.Payouts))
	for _, payout := range dbBlock.Payouts {
		var pt BlockPayoutType
		header.Payouts = append(header.Payouts, &BlockPayout{

		})
	}*/
	// return the fully assembled block
	return &Block{
		Header:       header,
		Transactions: nil, // TODO
	}, nil
}
func (r *queryRootResolver) blockByReferencePoint(ctx context.Context, reference *ReferencePoint) (*Block, error) {
	if reference == nil {
		return nil, errors.New("no reference point given for block, while this is required if no ID is given")
	}
	return nil, errors.New("blockByReferencePoint is not yet implemented")
}
func (r *queryRootResolver) Blocks(ctx context.Context, after *ReferencePoint, first *int, before *ReferencePoint, last *int) (Transaction, error) {
	panic("not implemented")
}
func (r *queryRootResolver) Wallet(ctx context.Context, unlockhash *types.UnlockHash) (Wallet, error) {
	panic("not implemented")
}
func (r *queryRootResolver) Contract(ctx context.Context, unlockhash *types.UnlockHash) (Contract, error) {
	panic("not implemented")
}
