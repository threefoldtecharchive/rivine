package explorergraphql

import (
	"context"

	"github.com/threefoldtech/rivine/crypto"
	"github.com/threefoldtech/rivine/types"
)

// THIS CODE IS A STARTING POINT ONLY. IT WILL NOT BE UPDATED WITH SCHEMA CHANGES.

type Resolver struct{}

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
func (r *queryRootResolver) Block(ctx context.Context, id *crypto.Hash) (*Block, error) {
	panic("not implemented")
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
