package explorergraphql

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math/big"
	"sync"

	"github.com/threefoldtech/rivine/crypto"
	"github.com/threefoldtech/rivine/extensions/authcointx"
	"github.com/threefoldtech/rivine/extensions/minting"
	"github.com/threefoldtech/rivine/modules/explorergraphql/explorerdb"
	"github.com/threefoldtech/rivine/pkg/encoding/rivbin"
	"github.com/threefoldtech/rivine/types"
)

type (
	Transaction interface {
		Object       // GraphQL Object interface
		OutputParent // GraphQL OutputParent interface

		ID(context.Context) (*crypto.Hash, error)
		Version(context.Context) (*ByteVersion, error)
		ParentBlock(context.Context) (*TransactionParentInfo, error)
		CoinInputs(context.Context) ([]*Input, error)
		CoinOutputs(context.Context) ([]*Output, error)
		BlockStakeInputs(context.Context) ([]*Input, error)
		BlockStakeOutputs(context.Context) ([]*Output, error)
		FeePayouts(context.Context) ([]*TransactionFeePayout, error)
		ArbitraryData(context.Context) (*BinaryData, error)
	}
)

func NewTransaction(txid types.TransactionID, parentInfo *TransactionParentInfo, db explorerdb.DB) (Transaction, error) {
	info, err := db.GetObjectInfo(explorerdb.ObjectID(txid[:]))
	if err != nil {
		return nil, err
	}
	if info.Type != explorerdb.ObjectTypeTransaction {
		return nil, fmt.Errorf(
			"unexpected object type %d for transaction %s (expected %d)",
			info.Type, txid.String(), explorerdb.ObjectTypeTransaction)
	}
	return NewTransactionWithVersion(txid, types.TransactionVersion(info.Version), parentInfo, db)
}

func NewTransactionWithVersion(txid types.TransactionID, version types.TransactionVersion, parentInfo *TransactionParentInfo, db explorerdb.DB) (Transaction, error) {
	switch version {
	case types.TransactionVersionZero, types.TransactionVersionOne:
		return NewStandardTransaction(txid, version, parentInfo, db), nil
	// TODO: handle minting in a kind-off explorer extension manner
	case 128: // Minter Condition Definition Txn
		return NewMintConditionDefinitionTransaction(txid, version, parentInfo, db), nil
	case 129: // Minter Coin Creation Txn
		return NewMintCoinCreationTransaction(txid, version, parentInfo, db), nil
	case 130: // Minter Coin Destruction Txn
		return NewMintCoinDestructionTransaction(txid, version, parentInfo, db), nil
	// TODO: handle auth coin in a kind-off explorer extension manner
	case 176: // Auth Coin Condition Update Txn
		return NewAuthConditionUpdateTransaction(txid, version, parentInfo, db), nil
	case 177: // Auth Coin Address Update Txn
		return NewAuthAddressUpdateTransaction(txid, version, parentInfo, db), nil
	default:
		// TODO: support extensions and render unknowns as an except unknown transaction schema type
		return nil, fmt.Errorf("unsupported transaction version %d", version)
	}
}

type (
	BaseTransactionData struct {
		CoinInputs        []*Input
		CoinOutputs       []*Output
		BlockStakeInputs  []*Input
		BlockStakeOutputs []*Output
		FeePayouts        []*TransactionFeePayout
		ArbitraryData     *BinaryData
	}

	BaseTransaction struct {
		TransactionID      types.TransactionID
		TransactionVersion types.TransactionVersion

		TransactionParentInfo *TransactionParentInfo

		ExtensionDataResolver func([]byte) error

		DB explorerdb.DB

		Data     *BaseTransactionData
		DataOnce sync.Once
		DataErr  error
	}
)

var (
	_ Transaction = (*BaseTransaction)(nil)
)

func (bt *BaseTransaction) IsOutputParent() {}
func (bt *BaseTransaction) IsObject()       {}

func (bt *BaseTransaction) ResolveData() {
	bt.DataOnce.Do(bt._resolveDataOnce)
}

func (bt *BaseTransaction) _resolveDataOnce() {
	defer func() {
		if e := recover(); e != nil {
			bt.DataErr = fmt.Errorf("failed to fetch txn %s data from DB: %v", bt.TransactionID.String(), e)
		}
	}()

	txn, err := bt.DB.GetTransaction(bt.TransactionID)
	if err != nil {
		bt.DataErr = err
		return
	}

	var block *Block
	if bt.TransactionParentInfo == nil {
		block = NewBlock(txn.ParentBlock, bt.DB)
		bt.TransactionParentInfo = NewTransactionParentInfoForBlock(bt.TransactionID, block)
	} else {
		block = bt.TransactionParentInfo.parentBlock
	}
	bt.Data = &BaseTransactionData{
		CoinInputs:        make([]*Input, 0, len(txn.CoinInputs)),
		CoinOutputs:       make([]*Output, 0, len(txn.CoinOutputs)),
		BlockStakeInputs:  make([]*Input, 0, len(txn.BlockStakeInputs)),
		BlockStakeOutputs: make([]*Output, 0, len(txn.BlockStakeOutputs)),
		FeePayouts:        dbTxFeePayoutInfoAsGQL(&txn.FeePayout, bt.DB),
		ArbitraryData:     dbByteSliceAsBinaryData(txn.ArbitraryData),
	}
	for _, outputID := range txn.CoinInputs {
		// Note: somehow we might want to just use the lazy resolver this Txn is,
		// as the parent transaction, however this is not as easy,
		// given we would need the Transaction Implementation that wraps the BaseTransaction,
		// not the BaseTransaction itself... Not as easy with the current implementation
		bt.Data.CoinInputs = append(bt.Data.CoinInputs, NewInput(outputID, nil, nil, bt.DB))
	}
	for _, outputID := range txn.CoinOutputs {
		bt.Data.CoinOutputs = append(bt.Data.CoinOutputs, NewOutput(outputID, nil, block, bt.DB))
	}
	for _, outputID := range txn.BlockStakeInputs {
		// Note: somehow we might want to just use the lazy resolver this Txn is,
		// as the parent transaction, however this is not as easy,
		// given we would need the Transaction Implementation that wraps the BaseTransaction,
		// not the BaseTransaction itself... Not as easy with the current implementation
		bt.Data.BlockStakeInputs = append(bt.Data.BlockStakeInputs, NewInput(outputID, nil, nil, bt.DB))
	}
	for _, outputID := range txn.BlockStakeOutputs {
		bt.Data.BlockStakeOutputs = append(bt.Data.BlockStakeOutputs, NewOutput(outputID, nil, block, bt.DB))
	}

	if bt.ExtensionDataResolver != nil {
		err = bt.ExtensionDataResolver(txn.EncodedExtensionData)
		if err != nil {
			bt.DataErr = err
			return
		}
	}
}

func (bt *BaseTransaction) ID(context.Context) (*crypto.Hash, error) {
	h := crypto.Hash(bt.TransactionID)
	return &h, nil
}

func (bt *BaseTransaction) Version(context.Context) (*ByteVersion, error) {
	v := ByteVersion(bt.TransactionVersion)
	return &v, nil
}

func (bt *BaseTransaction) ParentBlock(context.Context) (*TransactionParentInfo, error) {
	bt.ResolveData()
	if bt.DataErr != nil {
		return nil, bt.DataErr
	}
	return bt.TransactionParentInfo, nil
}

func (bt *BaseTransaction) CoinInputs(ctx context.Context) ([]*Input, error) {
	bt.ResolveData()
	if bt.DataErr != nil {
		return nil, bt.DataErr
	}
	return bt.Data.CoinInputs, nil
}
func (bt *BaseTransaction) CoinOutputs(ctx context.Context) ([]*Output, error) {
	bt.ResolveData()
	if bt.DataErr != nil {
		return nil, bt.DataErr
	}
	return bt.Data.CoinOutputs, nil
}
func (bt *BaseTransaction) BlockStakeInputs(ctx context.Context) ([]*Input, error) {
	bt.ResolveData()
	if bt.DataErr != nil {
		return nil, bt.DataErr
	}
	return bt.Data.BlockStakeInputs, nil
}
func (bt *BaseTransaction) BlockStakeOutputs(ctx context.Context) ([]*Output, error) {
	bt.ResolveData()
	if bt.DataErr != nil {
		return nil, bt.DataErr
	}
	return bt.Data.BlockStakeOutputs, nil
}
func (bt *BaseTransaction) FeePayouts(ctx context.Context) ([]*TransactionFeePayout, error) {
	bt.ResolveData()
	if bt.DataErr != nil {
		return nil, bt.DataErr
	}
	return bt.Data.FeePayouts, nil
}
func (bt *BaseTransaction) ArbitraryData(ctx context.Context) (*BinaryData, error) {
	bt.ResolveData()
	if bt.DataErr != nil {
		return nil, bt.DataErr
	}
	return bt.Data.ArbitraryData, nil
}

type (
	TransactionParentInfo struct {
		txnID       types.TransactionID
		parentBlock *Block
	}
)

func NewTransactionParentInfo(txnID types.TransactionID, id types.BlockID, db explorerdb.DB) *TransactionParentInfo {
	return NewTransactionParentInfoForBlock(txnID, NewBlock(id, db))
}

func NewTransactionParentInfoForBlock(txnID types.TransactionID, parentBlock *Block) *TransactionParentInfo {
	return &TransactionParentInfo{
		txnID:       txnID,
		parentBlock: parentBlock,
	}
}

func (info *TransactionParentInfo) ID(ctx context.Context) (crypto.Hash, error) {
	header, err := info.parentBlock.Header(ctx)
	if err != nil {
		return crypto.Hash{}, err
	}
	return header.ID, nil
}
func (info *TransactionParentInfo) ParentID(ctx context.Context) (*crypto.Hash, error) {
	header, err := info.parentBlock.Header(ctx)
	if err != nil {
		return nil, err
	}
	return header.ParentID, nil
}
func (info *TransactionParentInfo) Height(ctx context.Context) (*types.BlockHeight, error) {
	header, err := info.parentBlock.Header(ctx)
	if err != nil {
		return nil, err
	}
	return header.BlockHeight, nil
}
func (info *TransactionParentInfo) Timestamp(ctx context.Context) (*types.Timestamp, error) {
	header, err := info.parentBlock.Header(ctx)
	if err != nil {
		return nil, err
	}
	return header.BlockTime, nil
}
func (info *TransactionParentInfo) TransactionOrder(ctx context.Context) (*int, error) {
	transactions, err := info.parentBlock.Transactions(ctx, nil)
	if err != nil {
		return nil, err
	}
	for idx, txn := range transactions {
		hash, err := txn.ID(ctx)
		if err != nil {
			return nil, err
		}
		if bytes.Compare((*hash)[:], info.txnID[:]) == 0 {
			index := idx
			return &index, nil
		}
	}
	return nil, fmt.Errorf("transaction %s was not found in fetched parent block", info.txnID.String())
}
func (info *TransactionParentInfo) SiblingTransactions(ctx context.Context, filter *TransactionsFilter) ([]Transaction, error) {
	order, err := info.TransactionOrder(ctx)
	if err != nil {
		return nil, err
	}
	transactions, err := info.parentBlock.Transactions(ctx, nil)
	if err != nil {
		return nil, err
	}
	outTransactions := make([]Transaction, 0, len(transactions)-1)
	ignoreIdx := *order
	for idx, txn := range transactions {
		if idx != ignoreIdx {
			outTransactions = append(outTransactions, txn)
		}
	}
	return FilterTransactions(ctx, outTransactions, filter)
}

type (
	StandardTransaction struct {
		BaseTransaction
	}
)

func NewStandardTransaction(txid types.TransactionID, version types.TransactionVersion, parentInfo *TransactionParentInfo, db explorerdb.DB) *StandardTransaction {
	return &StandardTransaction{
		BaseTransaction: BaseTransaction{
			TransactionID:         txid,
			TransactionVersion:    version,
			TransactionParentInfo: parentInfo,
			ExtensionDataResolver: nil, // no extension needed here
			DB:                    db,
			// data will be resolved in a lazy manner
		},
	}
}

type (
	mintConditionDefinitionTransactionData struct {
		Nonce           BinaryData
		MintCondition   UnlockCondition
		MintFulfillment UnlockFulfillment
	}
	MintConditionDefinitionTransaction struct {
		BaseTransaction
		mintData *mintConditionDefinitionTransactionData
	}

	mintCoinCreationTransactionData struct {
		Nonce           BinaryData
		MintCondition   UnlockCondition
		MintFulfillment UnlockFulfillment
	}
	MintCoinCreationTransaction struct {
		BaseTransaction
		mintData *mintCoinCreationTransactionData
	}

	MintCoinDestructionTransaction struct {
		BaseTransaction
	}
)

func NewMintConditionDefinitionTransaction(txid types.TransactionID, version types.TransactionVersion, parentInfo *TransactionParentInfo, db explorerdb.DB) *MintConditionDefinitionTransaction {
	txn := &MintConditionDefinitionTransaction{
		BaseTransaction: BaseTransaction{
			TransactionID:         txid,
			TransactionVersion:    version,
			TransactionParentInfo: parentInfo,
			ExtensionDataResolver: nil, // assigned below
			DB:                    db,
			// data will be resolved in a lazy manner
		},
	}
	txn.ExtensionDataResolver = txn._resolveExtensionData
	return txn
}

func (txn *MintConditionDefinitionTransaction) Nonce(ctx context.Context) (BinaryData, error) {
	txn.ResolveData() // resolve base transaction, also resolves our extension data
	if txn.DataErr != nil {
		return nil, txn.DataErr
	}
	return txn.mintData.Nonce, nil
}

func (txn *MintConditionDefinitionTransaction) NewMintCondition(ctx context.Context) (UnlockCondition, error) {
	txn.ResolveData() // resolve base transaction, also resolves our extension data
	if txn.DataErr != nil {
		return nil, txn.DataErr
	}
	return txn.mintData.MintCondition, nil
}

func (txn *MintConditionDefinitionTransaction) MintFulfillment(ctx context.Context) (UnlockFulfillment, error) {
	txn.ResolveData() // resolve base transaction, also resolves our extension data
	if txn.DataErr != nil {
		return nil, txn.DataErr
	}
	return txn.mintData.MintFulfillment, nil
}

func (txn *MintConditionDefinitionTransaction) _resolveExtensionData(encodedExtensionData []byte) error {
	var mcdtxExtensionData *minting.MinterDefinitionTransactionExtension
	err := rivbin.Unmarshal(encodedExtensionData, &mcdtxExtensionData)
	if err != nil {
		return fmt.Errorf("failed to rivbin unmarshal extension-encoded Minter Condition Definition data: %v", err)
	}
	if mcdtxExtensionData == nil {
		return errors.New("Minter Condition Definition Data is nil")
	}
	mintCondition, err := dbConditionAsUnlockCondition(mcdtxExtensionData.MintCondition)
	if err != nil {
		return fmt.Errorf("failed to convert new mint condition as GQL unlock condition: %v", err)
	}
	mintFulfillment, err := dbFulfillmentAsUnlockFulfillment(mcdtxExtensionData.MintFulfillment, nil)
	if err != nil {
		return fmt.Errorf("failed to convert used mint fulfillment as GQL unlock fulfillment: %v", err)
	}
	txn.mintData = &mintConditionDefinitionTransactionData{
		Nonce:           *dbByteSliceAsBinaryData(mcdtxExtensionData.Nonce[:]),
		MintCondition:   mintCondition,
		MintFulfillment: mintFulfillment,
	}
	return nil
}

func NewMintCoinCreationTransaction(txid types.TransactionID, version types.TransactionVersion, parentInfo *TransactionParentInfo, db explorerdb.DB) *MintCoinCreationTransaction {
	txn := &MintCoinCreationTransaction{
		BaseTransaction: BaseTransaction{
			TransactionID:         txid,
			TransactionVersion:    version,
			TransactionParentInfo: parentInfo,
			ExtensionDataResolver: nil, // assigned below
			DB:                    db,
			// data will be resolved in a lazy manner
		},
	}
	txn.ExtensionDataResolver = txn._resolveExtensionData
	return txn
}

func (txn *MintCoinCreationTransaction) Nonce(ctx context.Context) (BinaryData, error) {
	txn.ResolveData() // resolve base transaction, also resolves our extension data
	if txn.DataErr != nil {
		return nil, txn.DataErr
	}
	return txn.mintData.Nonce, nil
}

func (txn *MintCoinCreationTransaction) MintFulfillment(ctx context.Context) (UnlockFulfillment, error) {
	txn.ResolveData() // resolve base transaction, also resolves our extension data
	if txn.DataErr != nil {
		return nil, txn.DataErr
	}
	return txn.mintData.MintFulfillment, nil
}

func (txn *MintCoinCreationTransaction) _resolveExtensionData(encodedExtensionData []byte) error {
	var mcctxExtensionData *minting.CoinCreationTransactionExtension
	err := rivbin.Unmarshal(encodedExtensionData, &mcctxExtensionData)
	if err != nil {
		return fmt.Errorf("failed to rivbin unmarshal extension-encoded Minter Coin Creation data: %v", err)
	}
	if mcctxExtensionData == nil {
		return errors.New("Minter Coin Creation Data is nil")
	}
	mintFulfillment, err := dbFulfillmentAsUnlockFulfillment(mcctxExtensionData.MintFulfillment, nil)
	if err != nil {
		return fmt.Errorf("failed to convert used mint fulfillment as GQL unlock fulfillment: %v", err)
	}
	txn.mintData = &mintCoinCreationTransactionData{
		Nonce:           *dbByteSliceAsBinaryData(mcctxExtensionData.Nonce[:]),
		MintFulfillment: mintFulfillment,
	}
	return nil
}

func NewMintCoinDestructionTransaction(txid types.TransactionID, version types.TransactionVersion, parentInfo *TransactionParentInfo, db explorerdb.DB) *MintCoinDestructionTransaction {
	return &MintCoinDestructionTransaction{
		BaseTransaction: BaseTransaction{
			TransactionID:         txid,
			TransactionVersion:    version,
			TransactionParentInfo: parentInfo,
			ExtensionDataResolver: nil, // no extension needed here
			DB:                    db,
			// data will be resolved in a lazy manner
		},
	}
}

type (
	authAddressUpdateTransactionData struct {
		Nonce           BinaryData
		AuthAddresses   []*UnlockHashPublicKeyPair
		DeauthAddresses []*UnlockHashPublicKeyPair
		AuthFulfillment UnlockFulfillment
	}
	AuthAddressUpdateTransaction struct {
		BaseTransaction
		authData *authAddressUpdateTransactionData
	}

	authConditionUpdateTransactionData struct {
		Nonce            BinaryData
		NewAuthCondition UnlockCondition
		AuthFulfillment  UnlockFulfillment
	}
	AuthConditionUpdateTransaction struct {
		BaseTransaction
		authData *authConditionUpdateTransactionData
	}
)

func NewAuthAddressUpdateTransaction(txid types.TransactionID, version types.TransactionVersion, parentInfo *TransactionParentInfo, db explorerdb.DB) *AuthAddressUpdateTransaction {
	txn := &AuthAddressUpdateTransaction{
		BaseTransaction: BaseTransaction{
			TransactionID:         txid,
			TransactionVersion:    version,
			TransactionParentInfo: parentInfo,
			ExtensionDataResolver: nil, // assigned below
			DB:                    db,
			// data will be resolved in a lazy manner
		},
	}
	txn.ExtensionDataResolver = txn._resolveExtensionData
	return txn
}

func (txn *AuthAddressUpdateTransaction) Nonce(ctx context.Context) (BinaryData, error) {
	txn.ResolveData() // resolve base transaction, also resolves our extension data
	if txn.DataErr != nil {
		return nil, txn.DataErr
	}
	return txn.authData.Nonce, nil
}

func (txn *AuthAddressUpdateTransaction) AuthAddresses(ctx context.Context) ([]*UnlockHashPublicKeyPair, error) {
	txn.ResolveData() // resolve base transaction, also resolves our extension data
	if txn.DataErr != nil {
		return nil, txn.DataErr
	}
	return txn.authData.AuthAddresses, nil
}

func (txn *AuthAddressUpdateTransaction) DeauthAddresses(ctx context.Context) ([]*UnlockHashPublicKeyPair, error) {
	txn.ResolveData() // resolve base transaction, also resolves our extension data
	if txn.DataErr != nil {
		return nil, txn.DataErr
	}
	return txn.authData.DeauthAddresses, nil
}

func (txn *AuthAddressUpdateTransaction) AuthFulfillment(ctx context.Context) (UnlockFulfillment, error) {
	txn.ResolveData() // resolve base transaction, also resolves our extension data
	if txn.DataErr != nil {
		return nil, txn.DataErr
	}
	return txn.authData.AuthFulfillment, nil
}

func (txn *AuthAddressUpdateTransaction) _resolveExtensionData(encodedExtensionData []byte) error {
	var aadtxExtensionData *authcointx.AuthAddressUpdateTransactionExtension
	err := rivbin.Unmarshal(encodedExtensionData, &aadtxExtensionData)
	if err != nil {
		return fmt.Errorf("failed to rivbin unmarshal extension-encoded Auth Address Update data: %v", err)
	}
	if aadtxExtensionData == nil {
		return errors.New("Auth Address Update Data is nil")
	}
	authFulfillment, err := dbFulfillmentAsUnlockFulfillment(aadtxExtensionData.AuthFulfillment, nil)
	if err != nil {
		return fmt.Errorf("failed to convert used auth fulfillment as GQL unlock fulfillment: %v", err)
	}
	authAddresses := make([]*UnlockHashPublicKeyPair, 0, len(aadtxExtensionData.AuthAddresses))
	for idx := range aadtxExtensionData.AuthAddresses {
		addr := aadtxExtensionData.AuthAddresses[idx]
		authAddresses = append(authAddresses, dbUnlockHashAsUnlockHashPublicKeyPair(addr))
	}
	deauthAddresses := make([]*UnlockHashPublicKeyPair, 0, len(aadtxExtensionData.DeauthAddresses))
	for idx := range aadtxExtensionData.DeauthAddresses {
		addr := aadtxExtensionData.AuthAddresses[idx]
		deauthAddresses = append(deauthAddresses, dbUnlockHashAsUnlockHashPublicKeyPair(addr))
	}
	txn.authData = &authAddressUpdateTransactionData{
		Nonce:           *dbByteSliceAsBinaryData(aadtxExtensionData.Nonce[:]),
		AuthAddresses:   authAddresses,
		DeauthAddresses: deauthAddresses,
		AuthFulfillment: authFulfillment,
	}
	return nil
}

func NewAuthConditionUpdateTransaction(txid types.TransactionID, version types.TransactionVersion, parentInfo *TransactionParentInfo, db explorerdb.DB) *AuthConditionUpdateTransaction {
	txn := &AuthConditionUpdateTransaction{
		BaseTransaction: BaseTransaction{
			TransactionID:         txid,
			TransactionVersion:    version,
			TransactionParentInfo: parentInfo,
			ExtensionDataResolver: nil, // assigned below
			DB:                    db,
			// data will be resolved in a lazy manner
		},
	}
	txn.ExtensionDataResolver = txn._resolveExtensionData
	return txn
}

func (txn *AuthConditionUpdateTransaction) Nonce(ctx context.Context) (BinaryData, error) {
	txn.ResolveData() // resolve base transaction, also resolves our extension data
	if txn.DataErr != nil {
		return nil, txn.DataErr
	}
	return txn.authData.Nonce, nil
}

func (txn *AuthConditionUpdateTransaction) NewAuthCondition(ctx context.Context) (UnlockCondition, error) {
	txn.ResolveData() // resolve base transaction, also resolves our extension data
	if txn.DataErr != nil {
		return nil, txn.DataErr
	}
	return txn.authData.NewAuthCondition, nil
}

func (txn *AuthConditionUpdateTransaction) AuthFulfillment(ctx context.Context) (UnlockFulfillment, error) {
	txn.ResolveData() // resolve base transaction, also resolves our extension data
	if txn.DataErr != nil {
		return nil, txn.DataErr
	}
	return txn.authData.AuthFulfillment, nil
}

func (txn *AuthConditionUpdateTransaction) _resolveExtensionData(encodedExtensionData []byte) error {
	var acdtxExtensionData *authcointx.AuthConditionUpdateTransactionExtension
	err := rivbin.Unmarshal(encodedExtensionData, &acdtxExtensionData)
	if err != nil {
		return fmt.Errorf("failed to rivbin unmarshal extension-encoded Auth Condition Update data: %v", err)
	}
	if acdtxExtensionData == nil {
		return errors.New("Auth Condition Update Data is nil")
	}
	newAuthCondition, err := dbConditionAsUnlockCondition(acdtxExtensionData.AuthCondition)
	if err != nil {
		return fmt.Errorf("failed to convert (new) auth condition as GQL unlock condition: %v", err)
	}
	authFulfillment, err := dbFulfillmentAsUnlockFulfillment(acdtxExtensionData.AuthFulfillment, nil)
	if err != nil {
		return fmt.Errorf("failed to convert used auth fulfillment as GQL unlock fulfillment: %v", err)
	}
	txn.authData = &authConditionUpdateTransactionData{
		Nonce:            *dbByteSliceAsBinaryData(acdtxExtensionData.Nonce[:]),
		NewAuthCondition: newAuthCondition,
		AuthFulfillment:  authFulfillment,
	}
	return nil
}

func FilterTransactions(ctx context.Context, txns []Transaction, filter *TransactionsFilter) ([]Transaction, error) {
	if filter == nil {
		return txns, nil // nothing to filter
	}

	filterOut := make([]bool, len(txns))

	// filter on versions if defined
	if vl := len(filter.Versions); vl > 0 {
		vm := make(map[ByteVersion]struct{}, vl)
		for _, v := range filter.Versions {
			vm[v] = struct{}{}
		}
		var (
			ok  bool
			err error
			bv  *ByteVersion
		)
		for idx, txn := range txns {
			bv, err = txn.Version(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to get version of txn %d: %v", idx+1, err)
			}
			if bv == nil {
				filterOut[idx] = true
			} else if _, ok = vm[*bv]; !ok {
				filterOut[idx] = true
			}
		}
	}

	// filter on arbitrary data if defined
	if filter.ArbitraryData != nil {
		filterArbitraryData, err := filterFunctionForBinaryDataFilter(filter.ArbitraryData)
		if err != nil {
			return nil, fmt.Errorf("failed to process arbitrary data filter: %v", err)
		}

		if filterArbitraryData != nil {
			var bd *BinaryData
			for idx, txn := range txns {
				bd, err = txn.ArbitraryData(ctx)
				if err != nil {
					return nil, fmt.Errorf("failed to get arbitrary data of txn %d: %v", idx+1, err)
				}
				var b []byte
				if bd != nil {
					b = (*bd)[:]
				}
				if !filterArbitraryData(b) {
					filterOut[idx] = true
				}
			}
		}
	}

	// filter on (total) coin input value if defined
	if filter.CoinInputValue != nil {
		filterFunc, err := filterFunctionForBigIntFilter(filter.CoinInputValue)
		if err != nil {
			return nil, fmt.Errorf("failed to process coin input value filter: %v", err)
		}

		if filterFunc != nil {
			var (
				cis   []*Input
				cidx  int
				ci    *Input
				value *BigInt
			)
			for idx, txn := range txns {
				cis, err = txn.CoinInputs(ctx)
				if err != nil {
					return nil, fmt.Errorf("failed to get coin inputs of txn %d: %v", idx+1, err)
				}
				// aggregate total input value
				totalInputValue := BigInt{Int: new(big.Int)}
				for cidx, ci = range cis {
					value, err = ci.Value(ctx)
					if err != nil {
						return nil, fmt.Errorf("failed to get value of coin input %d of txn %d: %v", cidx+1, idx+1, err)
					}
					totalInputValue.Add(totalInputValue.Int, value.Int)
				}
				// filter txn based on input value
				if !filterFunc(totalInputValue) {
					filterOut[idx] = true
				}
			}
		}
	}

	// filter on (total) coin output value if defined
	if filter.CoinOutputValue != nil {
		filterFunc, err := filterFunctionForBigIntFilter(filter.CoinOutputValue)
		if err != nil {
			return nil, fmt.Errorf("failed to process coin output value filter: %v", err)
		}

		if filterFunc != nil {
			var (
				cos   []*Output
				cidx  int
				co    *Output
				value *BigInt
			)
			for idx, txn := range txns {
				cos, err = txn.CoinOutputs(ctx)
				if err != nil {
					return nil, fmt.Errorf("failed to get coin outputs of txn %d: %v", idx+1, err)
				}
				// aggregate total output value
				totalOutputValue := BigInt{Int: new(big.Int)}
				for cidx, co = range cos {
					value, err = co.Value(ctx)
					if err != nil {
						return nil, fmt.Errorf("failed to get value of coin output %d of txn %d: %v", cidx+1, idx+1, err)
					}
					totalOutputValue.Add(totalOutputValue.Int, value.Int)
				}
				// filter txn based on output value
				if !filterFunc(totalOutputValue) {
					filterOut[idx] = true
				}
			}
		}
	}

	// assemble new transaction list, and return it
	var filteredTxns []Transaction
	for idx, txn := range txns {
		if !filterOut[idx] { // if not filtered out, keep it
			filteredTxns = append(filteredTxns, txn)
		}
	}
	return filteredTxns, nil
}
