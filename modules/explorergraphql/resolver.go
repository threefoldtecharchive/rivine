package explorergraphql

import (
	"context"
) // THIS CODE IS A STARTING POINT ONLY. IT WILL NOT BE UPDATED WITH SCHEMA CHANGES.

type Resolver struct{}

func (r *Resolver) QueryRoot() QueryRootResolver {
	return &queryRootResolver{r}
}

type queryRootResolver struct{ *Resolver }

func (r *queryRootResolver) Object(ctx context.Context, id *string) (Object, error) {
	panic("not implemented")
}
func (r *queryRootResolver) Transaction(ctx context.Context, id *string) (Transaction, error) {
	panic("not implemented")
}
func (r *queryRootResolver) Transactions(ctx context.Context, after *string, first *int, before *string, last *int) (Transaction, error) {
	panic("not implemented")
}
func (r *queryRootResolver) Block(ctx context.Context, id *string) (*Block, error) {
	panic("not implemented")
}
func (r *queryRootResolver) Blocks(ctx context.Context, after *string, first *int, before *string, last *int) (Transaction, error) {
	panic("not implemented")
}
func (r *queryRootResolver) Wallet(ctx context.Context, unlockhash *string) (Wallet, error) {
	panic("not implemented")
}
func (r *queryRootResolver) Contract(ctx context.Context, unlockhash *string) (Contract, error) {
	panic("not implemented")
}
