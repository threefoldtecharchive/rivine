package explorergraphql

import (
	"context"
	"fmt"
	"sync"

	"github.com/threefoldtech/rivine/modules/explorergraphql/explorerdb"

	"github.com/threefoldtech/rivine/types"
)

type (
	baseWalletData struct {
		CoinOutputs       []*Output
		BlockStakeOutputs []*Output

		CoinBalance       *Balance
		BlockStakeBalance *Balance
	}

	FreeForAllWallet struct {
		uh types.UnlockHash
		db explorerdb.DB

		onceData sync.Once
		data     *baseWalletData
		dataErr  error
	}

	singleSignatureWalletData struct {
		baseWalletData

		PublicKey             *types.PublicKey
		MultiSignatureWallets []*MultiSignatureWallet
	}

	SingleSignatureWallet struct {
		uh types.UnlockHash
		db explorerdb.DB

		onceData sync.Once
		data     *singleSignatureWalletData
		dataErr  error
	}

	multiSignatureWalletData struct {
		baseWalletData

		Owners                 []*UnlockHashPublicKeyPair
		RequiredSignatureCount int
	}

	MultiSignatureWallet struct {
		uh types.UnlockHash
		db explorerdb.DB

		onceData sync.Once
		data     *multiSignatureWalletData
		dataErr  error
	}
)

// compile time interface check
type _baseWallet interface {
	Object
	Wallet
}

var (
	_ _baseWallet = (*FreeForAllWallet)(nil)
	_ _baseWallet = (*SingleSignatureWallet)(nil)
	_ _baseWallet = (*MultiSignatureWallet)(nil)
)

func newBaseWalletDataFromEDB(data *explorerdb.WalletData, db explorerdb.DB) (wdata *baseWalletData) {
	wdata = &baseWalletData{
		CoinOutputs:       make([]*Output, 0, len(data.CoinOutputs)),
		BlockStakeOutputs: make([]*Output, 0, len(data.BlockStakeOutputs)),
		CoinBalance:       dbBalanceAsGQL(&data.CoinBalance),
		BlockStakeBalance: dbBalanceAsGQL(&data.BlockStakeBalance),
	}
	for _, coid := range data.CoinOutputs {
		output := NewOutput(coid, nil, nil, db)
		wdata.CoinOutputs = append(wdata.CoinOutputs, output)
	}
	for _, bsoid := range data.BlockStakeOutputs {
		output := NewOutput(bsoid, nil, nil, db)
		wdata.BlockStakeOutputs = append(wdata.BlockStakeOutputs, output)
	}
	return
}

func NewFreeForAllWallet(uh types.UnlockHash, db explorerdb.DB) *FreeForAllWallet {
	return &FreeForAllWallet{
		uh: uh,
		db: db,
	}
}

func (wallet *FreeForAllWallet) walletData(ctx context.Context) (*baseWalletData, error) {
	wallet.onceData.Do(wallet._walletDataOnce)
	return wallet.data, wallet.dataErr
}

func (wallet *FreeForAllWallet) _walletDataOnce() {
	data, err := wallet.db.GetFreeForAllWallet(wallet.uh)
	if err != nil {
		if err != explorerdb.ErrNotFound {
			wallet.dataErr = fmt.Errorf("failed to fetch free-for-all wallet %s data from DB: %v", wallet.uh.String(), err)
			return
		}
		data.UnlockHash = wallet.uh
	}

	// restructure all fetched data
	wallet.data = newBaseWalletDataFromEDB(&data.WalletData, wallet.db)
}

// IsObject implements the generated Object GraphQL interface
func (wallet *FreeForAllWallet) IsObject() {}

// IsWallet implements the generated Wallet GraphQL interface
func (wallet *FreeForAllWallet) IsWallet() {}

func (wallet *FreeForAllWallet) UnlockHash(ctx context.Context) (types.UnlockHash, error) {
	return wallet.uh, nil
}
func (wallet *FreeForAllWallet) CoinOutputs(ctx context.Context) ([]*Output, error) {
	data, err := wallet.walletData(ctx)
	if err != nil {
		return nil, err
	}
	return data.CoinOutputs, nil
}
func (wallet *FreeForAllWallet) BlockStakeOutputs(ctx context.Context) ([]*Output, error) {
	data, err := wallet.walletData(ctx)
	if err != nil {
		return nil, err
	}
	return data.BlockStakeOutputs, nil
}
func (wallet *FreeForAllWallet) CoinBalance(ctx context.Context) (*Balance, error) {
	data, err := wallet.walletData(ctx)
	if err != nil {
		return nil, err
	}
	return data.CoinBalance, nil
}
func (wallet *FreeForAllWallet) BlockStakeBalance(ctx context.Context) (*Balance, error) {
	data, err := wallet.walletData(ctx)
	if err != nil {
		return nil, err
	}
	return data.BlockStakeBalance, nil
}

func NewSingleSignatureWallet(uh types.UnlockHash, db explorerdb.DB) *SingleSignatureWallet {
	return &SingleSignatureWallet{
		uh: uh,
		db: db,
	}
}
func (wallet *SingleSignatureWallet) walletData(ctx context.Context) (*singleSignatureWalletData, error) {
	wallet.onceData.Do(wallet._walletDataOnce)
	return wallet.data, wallet.dataErr
}

func (wallet *SingleSignatureWallet) _walletDataOnce() {
	data, err := wallet.db.GetSingleSignatureWallet(wallet.uh)
	if err != nil {
		if err != explorerdb.ErrNotFound {
			wallet.dataErr = fmt.Errorf("failed to fetch single signature wallet %s data from DB: %v", wallet.uh.String(), err)
			return
		}
		data.UnlockHash = wallet.uh
	}

	// restructure all fetched data
	wallet.data = &singleSignatureWalletData{
		baseWalletData:        *newBaseWalletDataFromEDB(&data.WalletData, wallet.db),
		PublicKey:             nil, // resolved by another sibling resolver
		MultiSignatureWallets: make([]*MultiSignatureWallet, 0, len(data.MultiSignatureWallets)),
	}
	for _, uh := range data.MultiSignatureWallets {
		// each multi signature wallet can be resolved as far as it has to be resolved,
		// if only the unlockhashes are desired for example, no additional database fetches are required
		wallet.data.MultiSignatureWallets = append(
			wallet.data.MultiSignatureWallets, NewMultiSignatureWallet(uh, wallet.db))
	}
}

// IsObject implements the generated Object GraphQL interface
func (wallet *SingleSignatureWallet) IsObject() {}

// IsWallet implements the generated Wallet GraphQL interface
func (wallet *SingleSignatureWallet) IsWallet() {}

func (wallet *SingleSignatureWallet) UnlockHash(ctx context.Context) (types.UnlockHash, error) {
	return wallet.uh, nil
}
func (wallet *SingleSignatureWallet) CoinOutputs(ctx context.Context) ([]*Output, error) {
	data, err := wallet.walletData(ctx)
	if err != nil {
		return nil, err
	}
	return data.CoinOutputs, nil
}
func (wallet *SingleSignatureWallet) BlockStakeOutputs(ctx context.Context) ([]*Output, error) {
	data, err := wallet.walletData(ctx)
	if err != nil {
		return nil, err
	}
	return data.BlockStakeOutputs, nil
}
func (wallet *SingleSignatureWallet) CoinBalance(ctx context.Context) (*Balance, error) {
	data, err := wallet.walletData(ctx)
	if err != nil {
		return nil, err
	}
	return data.CoinBalance, nil
}
func (wallet *SingleSignatureWallet) BlockStakeBalance(ctx context.Context) (*Balance, error) {
	data, err := wallet.walletData(ctx)
	if err != nil {
		return nil, err
	}
	return data.BlockStakeBalance, nil
}
func (wallet *SingleSignatureWallet) PublicKey(ctx context.Context) (*types.PublicKey, error) {
	pk, err := wallet.db.GetPublicKey(wallet.uh)
	if err != nil {
		if err == explorerdb.ErrNotFound {
			return nil, nil // no error
		}
		return nil, err
	}
	return &pk, err
}
func (wallet *SingleSignatureWallet) MultiSignatureWallets(ctx context.Context) ([]*MultiSignatureWallet, error) {
	data, err := wallet.walletData(ctx)
	if err != nil {
		return nil, err
	}
	return data.MultiSignatureWallets, nil
}

func NewMultiSignatureWallet(uh types.UnlockHash, db explorerdb.DB) *MultiSignatureWallet {
	return &MultiSignatureWallet{
		uh: uh,
		db: db,
	}
}

func (wallet *MultiSignatureWallet) walletData(ctx context.Context) (*multiSignatureWalletData, error) {
	wallet.onceData.Do(wallet._walletDataOnce)
	return wallet.data, wallet.dataErr
}

func (wallet *MultiSignatureWallet) _walletDataOnce() {
	data, err := wallet.db.GetMultiSignatureWallet(wallet.uh)
	if err != nil {
		if err != explorerdb.ErrNotFound {
			wallet.dataErr = fmt.Errorf("failed to fetch multi signature wallet %s data from DB: %v", wallet.uh.String(), err)
			return
		}
		data.UnlockHash = wallet.uh
	}

	// restructure all fetched data
	wallet.data = &multiSignatureWalletData{
		baseWalletData:         *newBaseWalletDataFromEDB(&data.WalletData, wallet.db),
		Owners:                 make([]*UnlockHashPublicKeyPair, 0, len(data.Owners)),
		RequiredSignatureCount: data.RequiredSgnatureCount,
	}
	for _, uh := range data.Owners {
		// for each pair, the public key can be resolved in a lazy manner
		// automatically by the pair type's lazy resolver
		wallet.data.Owners = append(wallet.data.Owners, &UnlockHashPublicKeyPair{
			UnlockHash: uh,
			// PublicKey is resolved in a lazy manner when desired
		})
	}
}

// IsObject implements the generated Object GraphQL interface
func (wallet *MultiSignatureWallet) IsObject() {}

// IsWallet implements the generated Wallet GraphQL interface
func (wallet *MultiSignatureWallet) IsWallet() {}

func (wallet *MultiSignatureWallet) UnlockHash(ctx context.Context) (types.UnlockHash, error) {
	return wallet.uh, nil
}
func (wallet *MultiSignatureWallet) CoinOutputs(ctx context.Context) ([]*Output, error) {
	data, err := wallet.walletData(ctx)
	if err != nil {
		return nil, err
	}
	return data.CoinOutputs, nil
}
func (wallet *MultiSignatureWallet) BlockStakeOutputs(ctx context.Context) ([]*Output, error) {
	data, err := wallet.walletData(ctx)
	if err != nil {
		return nil, err
	}
	return data.BlockStakeOutputs, nil
}
func (wallet *MultiSignatureWallet) CoinBalance(ctx context.Context) (*Balance, error) {
	data, err := wallet.walletData(ctx)
	if err != nil {
		return nil, err
	}
	return data.CoinBalance, nil
}
func (wallet *MultiSignatureWallet) BlockStakeBalance(ctx context.Context) (*Balance, error) {
	data, err := wallet.walletData(ctx)
	if err != nil {
		return nil, err
	}
	return data.BlockStakeBalance, nil
}
func (wallet *MultiSignatureWallet) Owners(ctx context.Context) ([]*UnlockHashPublicKeyPair, error) {
	data, err := wallet.walletData(ctx)
	if err != nil {
		return nil, err
	}
	return data.Owners, nil
}
func (wallet *MultiSignatureWallet) RequiredSignatureCount(ctx context.Context) (int, error) {
	data, err := wallet.walletData(ctx)
	if err != nil {
		return 0, err
	}
	return data.RequiredSignatureCount, nil
}
