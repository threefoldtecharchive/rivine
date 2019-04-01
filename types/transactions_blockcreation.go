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
	SpecifierBlockCreationTransaction = Specifier{'b', 'l', 'o', 'c', 'k', ' ', 'c', 'r', 'e', 'a', 't', 'i', 'o', 'n'}
)

const (
	// TransactionVersionBlockCreation defines the special transaction used for
	// creating blocks by referencing an owned blockstake output
	TransactionVersionBlockCreation TransactionVersion = 2
)

type (
	// BlockCreationTransaction defines the transaction (with version 0x02)
	// used to create a new block by proving ownership of a referenced
	// blockstake output
	BlockCreationTransaction struct {
		// Reference unlocks a blockstake output to prove ownership, but does not consume it
		Reference BlockStakeInput
		// Height for the block this tx is going to create
		Height BlockHeight
	}

	// BlockCreationTransactionExtension defines the BlockCreationTransaction Extension Data
	BlockCreationTransactionExtension struct {
		// Reference unlocks a blockstake output to prove ownership, but does not consume it
		Reference BlockStakeInput
		// Height for the block this tx is going to create
		Height BlockHeight
	}
)

// BlockCreationTransactionFromTransaction creates a BlockCreationTransaction, using a regular in-memory rivine transaction
func BlockCreationTransactionFromTransaction(tx Transaction) (BlockCreationTransaction, error) {
	if tx.Version != TransactionVersionBlockCreation {
		return BlockCreationTransaction{}, fmt.Errorf(
			"block creation transaction requires tx version %d", TransactionVersionBlockCreation,
		)
	}
	return BlockCreationTransactionFromTransactionData(TransactionData{
		CoinInputs:        tx.CoinInputs,
		CoinOutputs:       tx.CoinOutputs,
		BlockStakeInputs:  tx.BlockStakeInputs,
		BlockStakeOutputs: tx.BlockStakeOutputs,
		MinerFees:         tx.MinerFees,
		ArbitraryData:     tx.ArbitraryData,
		Extension:         tx.Extension,
	})
}

// BlockCreationTransactionFromTransactionData creates an BlockCreationTransaction,
// using the TransactionData from a regular in-memory rivine transaction.
func BlockCreationTransactionFromTransactionData(txData TransactionData) (BlockCreationTransaction, error) {
	// validate transaction data

	// no coin inputs or outputs allowed
	if len(txData.CoinInputs) != 0 || len(txData.CoinOutputs) != 0 {
		return BlockCreationTransaction{}, errors.New("no coin input or outputs allowed for a block creation transaction")
	}

	// no miner fee allowed
	if len(txData.MinerFees) != 0 {
		return BlockCreationTransaction{}, errors.New("no transaction fees allowed for a block creation transaction")
	}

	// no arb data allowed
	if len(txData.ArbitraryData) != 0 {
		return BlockCreationTransaction{}, errors.New("no arbitrary data allowed for a block creation transaction")
	}

	// no blockstake input and outputs allowed, can't move block stakes
	if len(txData.BlockStakeOutputs) != 0 || len(txData.BlockStakeInputs) != 0 {
		return BlockCreationTransaction{}, errors.New("no blockstake outputs allowed for a block creation transaction")
	}

	tx := BlockCreationTransaction{}

	if txData.Extension != nil {
		extensionData, ok := txData.Extension.(*BlockCreationTransactionExtension)
		if !ok {
			return tx, errors.New("invalid extension data for a block creation transaction")
		}

		tx.Reference = extensionData.Reference
		tx.Height = extensionData.Height
	}

	return tx, nil
}

// TransactionData returns this BlockCreationTransaction
// as regular rivine transaction data.
func (bctx *BlockCreationTransaction) TransactionData() TransactionData {
	txData := TransactionData{
		Extension: &BlockCreationTransactionExtension{
			Reference: bctx.Reference,
			Height:    bctx.Height,
		},
	}
	return txData
}

// Transaction returns this BlockCreationTransaction
// as regular rivine transaction, using TransactionVersionBlockCreation as the type.
func (bctx *BlockCreationTransaction) Transaction() Transaction {
	tx := Transaction{
		Version: TransactionVersionBlockCreation,
		Extension: &BlockCreationTransactionExtension{
			Reference: bctx.Reference,
			Height:    bctx.Height,
		},
	}
	return tx
}

// MarshalSia implements SiaMarshaler.MarshalSia,
// alias of MarshalRivine for backwards-compatibility reasons.
func (bctx BlockCreationTransaction) MarshalSia(w io.Writer) error {
	return bctx.MarshalRivine(w)
}

// UnmarshalSia implements SiaUnmarshaler.UnmarshalSia,
// alias of UnmarshalRivine for backwards-compatibility reasons.
func (bctx *BlockCreationTransaction) UnmarshalSia(r io.Reader) error {
	return bctx.UnmarshalRivine(r)
}

// MarshalRivine implements RivineMarshaler.MarshalRivine
func (bctx BlockCreationTransaction) MarshalRivine(w io.Writer) error {
	return rivbin.NewEncoder(w).EncodeAll(
		bctx.Reference,
	)
}

// UnmarshalRivine implements RivineUnmarshaler.UnmarshalRivine
func (bctx *BlockCreationTransaction) UnmarshalRivine(r io.Reader) error {
	return rivbin.NewDecoder(r).DecodeAll(
		&bctx.Reference,
	)
}

type (
	// BlockCreationTransactionController defines a transaction controller for a a transaction type
	// reserved at type 0x02. It allows creation of blocks without blockstake respend
	BlockCreationTransactionController struct {
		bsog BlockStakeOutputGetter
	}

	//BlockStakeOutputGetter allows the retrieval of a blockstake output based on its ID
	BlockStakeOutputGetter interface {
		GetBlockStakeOutput(BlockStakeOutputID) (BlockStakeOutput, error)
	}
)

var (
	// ensure at compile time that BlockCreationTransactionController
	// implements the desired interfaces
	_ TransactionController      = BlockCreationTransactionController{}
	_ TransactionValidator       = BlockCreationTransactionController{}
	_ TransactionSignatureHasher = BlockCreationTransactionController{}
	_ TransactionIDEncoder       = BlockCreationTransactionController{}
	_ TransactionExtensionSigner = BlockCreationTransactionController{}
)

// NewBlockCreationTransactionController creates a new block creation transaction controller
func NewBlockCreationTransactionController(bsog BlockStakeOutputGetter) BlockCreationTransactionController {
	return BlockCreationTransactionController{
		bsog: bsog,
	}
}

// EncodeTransactionData implements TransactionController.EncodeTransactionData
func (bctc BlockCreationTransactionController) EncodeTransactionData(w io.Writer, txData TransactionData) error {
	bctx, err := BlockCreationTransactionFromTransactionData(txData)
	if err != nil {
		return fmt.Errorf("failed to convert txData to a BlockCreationTx: %v", err)
	}
	return rivbin.NewEncoder(w).Encode(bctx)
}

// DecodeTransactionData implements TransactionController.DecodeTransactionData
func (bctc BlockCreationTransactionController) DecodeTransactionData(r io.Reader) (TransactionData, error) {
	var bctx BlockCreationTransaction
	err := rivbin.NewDecoder(r).Decode(&bctx)
	if err != nil {
		return TransactionData{}, fmt.Errorf(
			"failed to binary-decode tx as a BlockCreationTx: %v", err)
	}
	// return block creation tx as regular rivine tx data
	return bctx.TransactionData(), nil
}

// JSONEncodeTransactionData implements TransactionController.JSONEncodeTransactionData
func (bctc BlockCreationTransactionController) JSONEncodeTransactionData(txData TransactionData) ([]byte, error) {
	bctx, err := BlockCreationTransactionFromTransactionData(txData)
	if err != nil {
		return nil, fmt.Errorf("failed to convert txData to a BlockCreationTx: %v", err)
	}
	return json.Marshal(bctx)
}

// JSONDecodeTransactionData implements TransactionController.JSONDecodeTransactionData
func (bctc BlockCreationTransactionController) JSONDecodeTransactionData(data []byte) (TransactionData, error) {
	var bctx BlockCreationTransaction
	err := json.Unmarshal(data, &bctx)
	if err != nil {
		return TransactionData{}, fmt.Errorf(
			"failed to json-decode tx as a BlockCreationTx:: %v", err)
	}
	// return block creation tx as regular rivine tx data
	return bctx.TransactionData(), nil
}

// ValidateTransaction implements TransactionValidator.ValidateTransaction
func (bctc BlockCreationTransactionController) ValidateTransaction(t Transaction, ctx ValidationContext, constants TransactionValidationConstants) error {
	// check tx fits within a block
	err := TransactionFitsInABlock(t, constants.BlockSizeLimit)
	if err != nil {
		return err
	}

	// get BlockCreationTx
	bctx, err := BlockCreationTransactionFromTransaction(t)
	if err != nil {
		return fmt.Errorf("failed to use tx as a BlockCreationTx: %v", err)
	}

	// check if the reference is a standard fulfillment
	if err = bctx.Reference.Fulfillment.IsStandardFulfillment(ctx); err != nil {
		return err
	}

	bso, err := bctc.bsog.GetBlockStakeOutput(bctx.Reference.ParentID)
	if err != nil {
		return fmt.Errorf("failed to get the referenced blockstake output condition condition: %v", err)
	}

	if err = bso.Condition.Fulfill(bctx.Reference.Fulfillment, FulfillContext{BlockHeight: ctx.BlockHeight, BlockTime: ctx.BlockTime, Transaction: t, ExtraObjects: nil}); err != nil {
		return err
	}

	// check block height in tx
	if bctx.Height != ctx.BlockHeight {
		return fmt.Errorf("tx is supposed to create block %v, but is in block %v", bctx.Height, ctx.BlockHeight)
	}

	// Tx is valid
	return nil
}

// SignatureHash implements TransactionSignatureHasher.SignatureHash
func (bctc BlockCreationTransactionController) SignatureHash(t Transaction, extraObjects ...interface{}) (crypto.Hash, error) {
	bctx, err := BlockCreationTransactionFromTransaction(t)
	if err != nil {
		return crypto.Hash{}, fmt.Errorf("failed to use tx as a BlockCreationTx: %v", err)
	}

	h := crypto.NewHash()
	enc := rivbin.NewEncoder(h)

	enc.EncodeAll(
		t.Version,
		SpecifierBlockCreationTransaction,
		bctx.Reference.ParentID,
		bctx.Height,
	)

	if len(extraObjects) > 0 {
		enc.EncodeAll(extraObjects...)
	}

	var hash crypto.Hash
	h.Sum(hash[:0])
	return hash, nil
}

// EncodeTransactionIDInput implements TransactionIDEncoder.EncodeTransactionIDInput
func (bctc BlockCreationTransactionController) EncodeTransactionIDInput(w io.Writer, txData TransactionData) error {
	bctx, err := BlockCreationTransactionFromTransactionData(txData)
	if err != nil {
		return fmt.Errorf("failed to convert txData to a BlockCreationTx: %v", err)
	}
	return rivbin.NewEncoder(w).EncodeAll(SpecifierBlockCreationTransaction, bctx)
}

// SignExtension implements TransactionExtensionSigner.SignExtension
func (bctc BlockCreationTransactionController) SignExtension(extension interface{}, sign func(*UnlockFulfillmentProxy, UnlockConditionProxy, ...interface{}) error) (interface{}, error) {
	bctxExtension, ok := extension.(*BlockCreationTransactionExtension)
	if !ok {
		return nil, errors.New("Invalid extension data for a block creation transaction")
	}

	bso, err := bctc.bsog.GetBlockStakeOutput(bctxExtension.Reference.ParentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get the referenced blockstake output condition condition: %v", err)
	}
	err = sign(&bctxExtension.Reference.Fulfillment, bso.Condition)
	if err != nil {
		return nil, fmt.Errorf("failed to sign block creation tx extension: %v", err)
	}
	return bctxExtension, nil
}
