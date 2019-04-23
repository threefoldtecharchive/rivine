package types

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/threefoldtech/rivine/crypto"
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
		// CoinInputs are used to pay the required fee, the transaction fee
		CoinInputs []CoinInput

		// RefundOutput is an optional coin output used to return the leftover inputs
		// after the txfee has been funded
		RefundOutput *CoinOutput

		// ArbitraryData is optional
		ArbitraryData []byte

		// TransactionFee is the regular tx fee paid for a transaction
		TransactionFee Currency

		// Reference unlocks a blockstake output to prove ownership, but does not consume it
		Reference BlockStakeInput

		// Delegation is the condition which needs to be unlocked to use the delegated blockstakes
		Delegation BlockStakeOutput

		// Fee is the percantage of the block reward the block creator can keep, 0 to 100
		Fee uint8
	}

	// DelegationTransactionExtension defines the DelegationTransaction Extension Data
	DelegationTransactionExtension struct {
		// Reference unlocks a blockstake output to prove ownership, but does not consume it
		Reference BlockStakeInput
		// Delegation is the condition which needs to be unlocked to use the delegated blockstakes
		Delegation BlockStakeOutput
		// Fee is the percantage of the block reward the block creator can keep, 0 to 100
		Fee uint8
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

	// can have at most 1 refund output
	if len(txData.CoinOutputs) > 1 {
		return DelegationTransaction{}, errors.New("can have at most 1 coin output as refund output")
	}

	// need minerfees
	if len(txData.MinerFees) != 1 {
		return DelegationTransaction{}, errors.New("transaction fee must be paid for a delegation transaction")
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
		tx.Fee = extensionData.Fee
	}

	tx.CoinInputs = txData.CoinInputs
	tx.ArbitraryData = txData.ArbitraryData
	tx.TransactionFee = txData.MinerFees[0]

	if len(txData.CoinOutputs) == 1 {
		tx.RefundOutput = &txData.CoinOutputs[0]
	}

	return tx, nil
}

// TransactionData returns this DelegationTransaction
// as regular rivine transaction data.
func (dtx *DelegationTransaction) TransactionData() TransactionData {
	txData := TransactionData{
		ArbitraryData: dtx.ArbitraryData,
		CoinInputs:    dtx.CoinInputs,
		CoinOutputs:   []CoinOutput{*dtx.RefundOutput},
		MinerFees:     []Currency{dtx.TransactionFee},
		Extension: &DelegationTransactionExtension{
			Reference:  dtx.Reference,
			Delegation: dtx.Delegation,
			Fee:        dtx.Fee,
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
		CoinOutputs:   []CoinOutput{*dtx.RefundOutput},
		MinerFees:     []Currency{dtx.TransactionFee},
		Extension: &DelegationTransaction{
			Reference:  dtx.Reference,
			Delegation: dtx.Delegation,
			Fee:        dtx.Fee,
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
		dtx.RefundOutput,
		dtx.TransactionFee,
		dtx.ArbitraryData,
		dtx.Reference,
		dtx.Delegation,
		dtx.Fee,
	)
}

// UnmarshalRivine implements RivineUnmarshaler.UnmarshalRivine
func (dtx *DelegationTransaction) UnmarshalRivine(r io.Reader) error {
	return rivbin.NewDecoder(r).DecodeAll(
		&dtx.CoinInputs,
		&dtx.RefundOutput,
		&dtx.TransactionFee,
		&dtx.ArbitraryData,
		&dtx.Reference,
		&dtx.Delegation,
		&dtx.Fee,
	)
}

type (
	// DelegationTransactionController defines a transaction controller for a a transaction type
	// reserved at type 0x03. It allows delegation of block stakes to a third party so they can use
	// them to create blocks
	DelegationTransactionController struct {
		bsog BlockStakeOutputGetter
	}
)

var (
	// ensure at compile time that BlockCreationTransactionController
	// implements the desired interfaces
	_ TransactionController      = DelegationTransactionController{}
	_ TransactionValidator       = DelegationTransactionController{}
	_ TransactionSignatureHasher = DelegationTransactionController{}
	_ TransactionIDEncoder       = DelegationTransactionController{}
	_ TransactionExtensionSigner = DelegationTransactionController{}
)

// NewDelegationTransactionController creates a new block creation transaction controller
func NewDelegationTransactionController(bsog BlockStakeOutputGetter) DelegationTransactionController {
	return DelegationTransactionController{
		bsog: bsog,
	}
}

// EncodeTransactionData implements TransactionController.EncodeTransactionData
func (dtc DelegationTransactionController) EncodeTransactionData(w io.Writer, txData TransactionData) error {
	dtx, err := DelegationTransactionFromTransactionData(txData)
	if err != nil {
		return fmt.Errorf("failed to convert txData to a DelegationTx: %v", err)
	}
	return rivbin.NewEncoder(w).Encode(dtx)
}

// DecodeTransactionData implements TransactionController.DecodeTransactionData
func (dtc DelegationTransactionController) DecodeTransactionData(r io.Reader) (TransactionData, error) {
	var dtx DelegationTransaction
	err := rivbin.NewDecoder(r).Decode(&dtx)
	if err != nil {
		return TransactionData{}, fmt.Errorf(
			"failed to binary-decode tx as a DelegationTx: %v", err)
	}
	// return block creation tx as regular rivine tx data
	return dtx.TransactionData(), nil
}

// JSONEncodeTransactionData implements TransactionController.JSONEncodeTransactionData
func (dtc DelegationTransactionController) JSONEncodeTransactionData(txData TransactionData) ([]byte, error) {
	dtx, err := DelegationTransactionFromTransactionData(txData)
	if err != nil {
		return nil, fmt.Errorf("failed to convert txData to a DelegationTx: %v", err)
	}
	return json.Marshal(dtx)
}

// JSONDecodeTransactionData implements TransactionController.JSONDecodeTransactionData
func (dtc DelegationTransactionController) JSONDecodeTransactionData(data []byte) (TransactionData, error) {
	var dtx DelegationTransaction
	err := json.Unmarshal(data, &dtx)
	if err != nil {
		return TransactionData{}, fmt.Errorf(
			"failed to json-decode tx as a DelegationTx: %v", err)
	}
	// return block creation tx as regular rivine tx data
	return dtx.TransactionData(), nil
}

// ValidateTransaction implements TransactionValidator.ValidateTransaction
func (dtc DelegationTransactionController) ValidateTransaction(t Transaction, ctx ValidationContext, constants TransactionValidationConstants) error {
	// check tx fits within a block
	err := TransactionFitsInABlock(t, constants.BlockSizeLimit)
	if err != nil {
		return err
	}

	// get DelegationTx
	dtx, err := DelegationTransactionFromTransaction(t)
	if err != nil {
		return fmt.Errorf("failed to use tx as a DelegationTx: %v", err)
	}

	// validate minerfee
	if dtx.TransactionFee.Cmp(constants.MinimumMinerFee) < 0 {
		return ErrTooSmallMinerFee
	}

	// validate fee is in acceptable range
	if dtx.Fee > 100 {
		return fmt.Errorf("fee too large: %v", err)
	}

	// prevent double spending
	spendCoins := make(map[CoinOutputID]struct{})
	for _, ci := range dtx.CoinInputs {
		if _, found := spendCoins[ci.ParentID]; found {
			return ErrDoubleSpend
		}
		spendCoins[ci.ParentID] = struct{}{}
	}

	// check if the reference is a standard fulfillment
	if err = dtx.Reference.Fulfillment.IsStandardFulfillment(ctx); err != nil {
		return err
	}

	bso, err := dtc.bsog.GetBlockStakeOutput(dtx.Reference.ParentID)
	if err != nil {
		return fmt.Errorf("failed to get the referenced blockstake output condition condition: %v", err)
	}

	// check that the amount of bs in the delegation is also the amount in the input
	if !dtx.Delegation.Value.Equals(bso.Value) {
		return fmt.Errorf("transaction does not delegate all blockstakes")
	}

	// Make sure we can unlock the delegation condition
	if err = dtx.Delegation.Condition.IsStandardCondition(ctx); err != nil {
		return fmt.Errorf("delegation condition is not standard: %v", err)
	}

	// Validate that the condition of the blockstake which is being delegated is succesfully fulfilled
	if err = bso.Condition.Fulfill(dtx.Reference.Fulfillment, FulfillContext{BlockHeight: ctx.BlockHeight, BlockTime: ctx.BlockTime, Transaction: t, ExtraObjects: nil}); err != nil {
		return err
	}

	// Tx is valid
	return nil
}

// SignatureHash implements TransactionSignatureHasher.SignatureHash
func (dtc DelegationTransactionController) SignatureHash(t Transaction, extraObjects ...interface{}) (crypto.Hash, error) {
	dtx, err := DelegationTransactionFromTransaction(t)
	if err != nil {
		return crypto.Hash{}, fmt.Errorf("failed to use tx as a DelegationTx: %v", err)
	}

	h := crypto.NewHash()
	enc := rivbin.NewEncoder(h)

	enc.EncodeAll(
		t.Version,
		SpecifierDelegationTransaction,
		dtx.CoinInputs,
		dtx.RefundOutput,
		dtx.TransactionFee,
		dtx.ArbitraryData,
		dtx.Reference.ParentID,
		dtx.Delegation,
		dtx.Fee,
	)

	if len(extraObjects) > 0 {
		enc.EncodeAll(extraObjects...)
	}

	var hash crypto.Hash
	h.Sum(hash[:0])
	return hash, nil
}

// EncodeTransactionIDInput implements TransactionIDEncoder.EncodeTransactionIDInput
func (dtc DelegationTransactionController) EncodeTransactionIDInput(w io.Writer, txData TransactionData) error {
	dtx, err := DelegationTransactionFromTransactionData(txData)
	if err != nil {
		return fmt.Errorf("failed to convert txData to a DelegationTx: %v", err)
	}
	return rivbin.NewEncoder(w).EncodeAll(SpecifierDelegationTransaction, dtx)
}

// SignExtension implements TransactionExtensionSigner.SignExtension
func (dtc DelegationTransactionController) SignExtension(extension interface{}, sign func(*UnlockFulfillmentProxy, UnlockConditionProxy, ...interface{}) error) (interface{}, error) {
	dtxExtension, ok := extension.(*DelegationTransactionExtension)
	if !ok {
		return nil, errors.New("Invalid extension data for a delegation transaction")
	}

	bso, err := dtc.bsog.GetBlockStakeOutput(dtxExtension.Reference.ParentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get the referenced blockstake output condition condition: %v", err)
	}
	err = sign(&dtxExtension.Reference.Fulfillment, bso.Condition)
	if err != nil {
		return nil, fmt.Errorf("failed to sign delegation tx extension: %v", err)
	}
	return dtxExtension, nil
}
