package explorergraphql

import (
	"context"
	"fmt"
	"sync"

	"github.com/threefoldtech/rivine/modules/explorergraphql/explorerdb"

	"github.com/threefoldtech/rivine/types"
)

type (
	atomicSwapContractData struct {
		ContractCondition   *AtomicSwapCondition
		ContractFulfillment *AtomicSwapFulfillment
		ContractValue       BigInt

		Transactions []Transaction

		CoinOutput *Output
	}

	AtomicSwapContract struct {
		uh types.UnlockHash
		db explorerdb.DB

		onceData sync.Once
		data     *atomicSwapContractData
		dataErr  error
	}
)

// compile-time interface checkers
var (
	_ Object   = (*AtomicSwapContract)(nil)
	_ Contract = (*AtomicSwapContract)(nil)
)

func NewAtomicSwapContract(uh types.UnlockHash, db explorerdb.DB) *AtomicSwapContract {
	return &AtomicSwapContract{
		uh: uh,
		db: db,
	}
}

func (contract *AtomicSwapContract) contractData(ctx context.Context) (*atomicSwapContractData, error) {
	contract.onceData.Do(contract._contractDataOnce)
	return contract.data, contract.dataErr
}

func (contract *AtomicSwapContract) _contractDataOnce() {
	defer func() {
		if e := recover(); e != nil {
			contract.dataErr = fmt.Errorf("failed to fetch atomic swap contract %s data from DB: %v", contract.uh.String(), e)
		}
	}()

	dbContract, err := contract.db.GetAtomicSwapContract(contract.uh)
	if err != nil {
		if err != explorerdb.ErrNotFound {
			contract.dataErr = fmt.Errorf("failed to fetch atomic swap contract %s data from DB: %v", contract.uh.String(), err)
			return
		}
		dbContract.UnlockHash = contract.uh
	}

	// restructure all fetched data
	contractCondition := dbAtomicSwapConditionAsGQL(&dbContract.ContractCondition)
	contract.data = &atomicSwapContractData{
		ContractCondition:   contractCondition,
		ContractFulfillment: nil, // optionally set later in this function, iff possible
		ContractValue:       dbCurrencyAsBigInt(dbContract.ContractValue),
		Transactions:        make([]Transaction, 0, len(dbContract.Transactions)),
	}
	for _, txnID := range dbContract.Transactions {
		txn, err := NewTransaction(txnID, nil, contract.db)
		if err != nil {
			contract.dataErr = fmt.Errorf(
				"failed to fetch atomic swap contract %s: failed to fetch referenced transaction %s data from DB: %v",
				contract.uh.String(), txnID.String(), err)
			return
		}
		contract.data.Transactions = append(contract.data.Transactions, txn)
	}
	// given the contract exist, we can assume at least one tx is linked,
	// and the transactions are in order of defined, where the first one is the created
	// one by convention
	contract.data.CoinOutput = NewOutput(
		types.OutputID(dbContract.CoinInput), nil,
		contract.data.Transactions[0], contract.db)
	if dbContract.SpenditureData != nil {
		contract.data.ContractFulfillment = dbAtomicSwapFulfillmentAsGQL(&dbContract.SpenditureData.ContractFulfillment, contractCondition)
	}
}

// IsObject implements the GraphQL Contract interface
func (contract *AtomicSwapContract) IsContract() {}

// IsObject implements the GraphQL Object interface
func (contract *AtomicSwapContract) IsObject() {}

func (contract *AtomicSwapContract) UnlockHash(ctx context.Context) (types.UnlockHash, error) {
	return contract.uh, nil
}
func (contract *AtomicSwapContract) ContractCondition(ctx context.Context) (*AtomicSwapCondition, error) {
	data, err := contract.contractData(ctx)
	if err != nil {
		return nil, err
	}
	return data.ContractCondition, nil
}
func (contract *AtomicSwapContract) ContractFulfillment(ctx context.Context) (*AtomicSwapFulfillment, error) {
	data, err := contract.contractData(ctx)
	if err != nil {
		return nil, err
	}
	return data.ContractFulfillment, nil
}
func (contract *AtomicSwapContract) ContractValue(ctx context.Context) (BigInt, error) {
	data, err := contract.contractData(ctx)
	if err != nil {
		return BigInt{}, err
	}
	return data.ContractValue, nil
}
func (contract *AtomicSwapContract) Transactions(ctx context.Context) ([]Transaction, error) {
	data, err := contract.contractData(ctx)
	if err != nil {
		return nil, err
	}
	return data.Transactions, nil
}
func (contract *AtomicSwapContract) CoinOutput(ctx context.Context) (*Output, error) {
	data, err := contract.contractData(ctx)
	if err != nil {
		return nil, err
	}
	return data.CoinOutput, nil
}
