package types

import (
	"errors"
	"fmt"
	"io"

	"github.com/threefoldtech/rivine/pkg/encoding/rivbin"
)

// These Specifiers are used internally when calculating a type's ID. See
// Specifier for more details.
var (
	SpecifierDelegationTransaction = Specifier{'d', 'e', 'l', 'e', 'g', 'a', 't', 'i', 'o', 'n'}
)

const (
	// TransactionVersionDelegation defines the special transaction used for
	// temporarily delegating a blockstake output to a third party
	TransactionVersionDelegation TransactionVersion = 3
)

type (
	// DelegationTransaction defines the transaction (with version 0x03)
	// used to allow a third party to use these blockstakes to create new blocks
	DelegationTransaction struct {
		CoinInputs    []CoinInput
		CoinOutputs   []CoinOutput
		ArbitraryData []byte
		// Reference unlocks a blockstake output to prove ownership, but does not consume it
		Reference BlockStakeInput
		// Delegation is the condition which needs to be unlocked to use the delegated blockstakes
		Delegation BlockStakeOutput
	}

	// DelegationTransactionExtension defines the DelegationTransaction Extension Data
	DelegationTransactionExtension struct {
		// Reference unlocks a blockstake output to prove ownership, but does not consume it
		Reference BlockStakeInput
		// Delegation is the condition which needs to be unlocked to use the delegated blockstakes
		Delegation BlockStakeOutput
	}
)

// DelegationTransactionFromTransaction creates a DelegationTransaction, using a regular in-memory rivine transaction
func DelegationTransactionFromTransaction(tx Transaction) (DelegationTransaction, error) {
	if tx.Version != TransactionVersionDelegation {
		return DelegationTransaction{}, fmt.Errorf(
			"delegation transaction requires tx version %d", TransactionVersionDelegation,
		)
	}
	return DelegationTransactionFromTransactionData(TransactionData{
		CoinInputs:        tx.CoinInputs,
		CoinOutputs:       tx.CoinOutputs,
		BlockStakeInputs:  tx.BlockStakeInputs,
		BlockStakeOutputs: tx.BlockStakeOutputs,
		MinerFees:         tx.MinerFees,
		ArbitraryData:     tx.ArbitraryData,
		Extension:         tx.Extension,
	})
}

// DelegationTransactionFromTransactionData creates a DelegationTransaction,
// using the TransactionData from a regular in-memory rivine transaction.
func DelegationTransactionFromTransactionData(txData TransactionData) (DelegationTransaction, error) {
	// validate transaction data

	// need at least 1 coin input to fund the transaction
	if len(txData.CoinInputs) == 0 {
		return DelegationTransaction{}, errors.New("need at least one coin input for a delegate  transaction")
	}

	// need minerfees
	if len(txData.MinerFees) == 0 {
		return DelegationTransaction{}, errors.New("transaction fees must be paid for a delegation transaction")
	}

	// no blockstake input and outputs allowed, can't move block stakes. The block stakes to delegate are identified
	// in the transaction data
	if len(txData.BlockStakeOutputs) != 0 || len(txData.BlockStakeInputs) != 0 {
		return DelegationTransaction{}, errors.New("no blockstake outputs allowed for a delegate transaction")
	}

	tx := DelegationTransaction{}

	if txData.Extension != nil {
		extensionData, ok := txData.Extension.(*DelegationTransactionExtension)
		if !ok {
			return tx, errors.New("invalid extension data for a block creation transaction")
		}

		tx.Reference = extensionData.Reference
		tx.Delegation = extensionData.Delegation
	}

	return tx, nil
}

// TransactionData returns this DelegationTransaction
// as regular rivine transaction data.
func (dtx *DelegationTransaction) TransactionData() TransactionData {
	txData := TransactionData{
		ArbitraryData: dtx.ArbitraryData,
		CoinInputs:    dtx.CoinInputs,
		CoinOutputs:   dtx.CoinOutputs,
		Extension: &DelegationTransactionExtension{
			Reference:  dtx.Reference,
			Delegation: dtx.Delegation,
		},
	}
	return txData
}

// Transaction returns this DelegationTransaction
// as regular rivine transaction, using TransactionVersionDelegation as the type.
func (dtx *DelegationTransaction) Transaction() Transaction {
	tx := Transaction{
		Version:       TransactionVersionDelegation,
		ArbitraryData: dtx.ArbitraryData,
		CoinInputs:    dtx.CoinInputs,
		CoinOutputs:   dtx.CoinOutputs,
		Extension: &DelegationTransaction{
			Reference:  dtx.Reference,
			Delegation: dtx.Delegation,
		},
	}
	return tx
}

// MarshalSia implements SiaMarshaler.MarshalSia,
// alias of MarshalRivine for backwards-compatibility reasons.
func (dtx DelegationTransaction) MarshalSia(w io.Writer) error {
	return dtx.MarshalRivine(w)
}

// UnmarshalSia implements SiaUnmarshaler.UnmarshalSia,
// alias of UnmarshalRivine for backwards-compatibility reasons.
func (dtx *DelegationTransaction) UnmarshalSia(r io.Reader) error {
	return dtx.UnmarshalRivine(r)
}

// MarshalRivine implements RivineMarshaler.MarshalRivine
func (dtx DelegationTransaction) MarshalRivine(w io.Writer) error {
	return rivbin.NewEncoder(w).EncodeAll(
		dtx.CoinInputs,
		dtx.CoinOutputs,
		dtx.ArbitraryData,
		dtx.Reference,
		dtx.Delegation,
	)
}

// UnmarshalRivine implements RivineUnmarshaler.UnmarshalRivine
func (dtx *DelegationTransaction) UnmarshalRivine(r io.Reader) error {
	return rivbin.NewDecoder(r).DecodeAll(
		&dtx.CoinInputs,
		&dtx.CoinOutputs,
		&dtx.ArbitraryData,
		&dtx.Reference,
		&dtx.Delegation,
	)
}
