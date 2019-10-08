package minting

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/threefoldtech/rivine/crypto"
	"github.com/threefoldtech/rivine/pkg/encoding/rivbin"
	"github.com/threefoldtech/rivine/pkg/encoding/siabin"
	types "github.com/threefoldtech/rivine/types"
)

// These Specifiers are used internally when calculating a Transaction's ID.
// See Rivine's Specifier for more details.
var (
	SpecifierMintDefinitionTransaction  = types.Specifier{'m', 'i', 'n', 't', 'e', 'r', ' ', 'd', 'e', 'f', 'i', 'n', ' ', 't', 'x'}
	SpecifierCoinCreationTransaction    = types.Specifier{'c', 'o', 'i', 'n', ' ', 'm', 'i', 'n', 't', ' ', 't', 'x'}
	SpecifierCoinDestructionTransaction = types.Specifier{'c', 'o', 'i', 'n', ' ', 'd', 'e', 's', 't', 'r', 'o', 'y', ' ', 't', 'x'}
)

type (
	// MintConditionGetter allows you to get the mint condition at a given block height.
	//
	// For the daemon this interface could be implemented directly by the DB object
	// that keeps track of the mint condition state, while for a client this could
	// come via the REST API from a rivine daemon in a more indirect way.
	MintConditionGetter interface {
		// GetActiveMintCondition returns the active active mint condition.
		GetActiveMintCondition() (types.UnlockConditionProxy, error)
		// GetMintConditionAt returns the mint condition at a given block height.
		GetMintConditionAt(height types.BlockHeight) (types.UnlockConditionProxy, error)
	}
)

type (
	// MintingBaseTransactionController is the base controller for all minting controllers
	MintingBaseTransactionController struct {
		UseLegacySiaEncoding bool
	}

	// MintingMinerFeeBaseTransactionController is the base controller for all minting controllers,
	// that require miner fees.
	MintingMinerFeeBaseTransactionController struct {
		MintingBaseTransactionController

		RequireMinerFees bool
	}
)

type binaryEncoder interface {
	Encode(interface{}) error
	EncodeAll(...interface{}) error
}

type binaryDecoder interface {
	Decode(interface{}) error
	DecodeAll(...interface{}) error
}

func (mbtc MintingBaseTransactionController) newBencoder(w io.Writer) binaryEncoder {
	if mbtc.UseLegacySiaEncoding {
		return siabin.NewEncoder(w)
	}
	return rivbin.NewEncoder(w)
}
func (mbtc MintingBaseTransactionController) newBdecoder(r io.Reader) binaryDecoder {
	if mbtc.UseLegacySiaEncoding {
		return siabin.NewDecoder(r)
	}
	return rivbin.NewDecoder(r)
}

type (
	// CoinCreationTransactionController defines a rivine-specific transaction controller,
	// for a CoinCreation Transaction. It allows for the creation of Coin Outputs,
	// without requiring coin inputs, but can only be used by the defined Coin Minters.
	CoinCreationTransactionController struct {
		MintingMinerFeeBaseTransactionController

		// MintConditionGetter is used to get a mint condition at the context-defined block height.
		//
		// The found MintCondition defines the condition that has to be fulfilled
		// in order to mint new coins into existence (in the form of non-backed coin outputs).
		MintConditionGetter MintConditionGetter

		// TransactionVersion is used to validate/set the transaction version
		// of a coin creation transaction.
		TransactionVersion types.TransactionVersion
	}

	// CoinDestructionTransactionController defines a rivine-specific transaction controller,
	// for a CoinDestruction Transaction. It allows the destruction of coins.
	CoinDestructionTransactionController struct {
		MintingBaseTransactionController

		// TransactionVersion is used to validate/set the transaction version
		// of a coin destruction transaction.
		TransactionVersion types.TransactionVersion
	}

	// MinterDefinitionTransactionController defines a rivine-specific transaction controller,
	// for a MinterDefinition Transaction. It allows the transfer of coin minting powers.
	MinterDefinitionTransactionController struct {
		MintingMinerFeeBaseTransactionController

		// MintConditionGetter is used to get a mint condition at the context-defined block height.
		//
		// The found MintCondition defines the condition that has to be fulfilled
		// in order to mint new coins into existence (in the form of non-backed coin outputs).
		MintConditionGetter MintConditionGetter

		// TransactionVersion is used to validate/set the transaction version
		// of a minter definitiontransaction.
		TransactionVersion types.TransactionVersion
	}
)

// ensure our controllers implement all desired interfaces
var (
	// ensure at compile time that CoinCreationTransactionController
	// implements the desired interfaces
	_ types.TransactionController      = CoinCreationTransactionController{}
	_ types.TransactionExtensionSigner = CoinCreationTransactionController{}
	_ types.TransactionSignatureHasher = CoinCreationTransactionController{}
	_ types.TransactionIDEncoder       = CoinCreationTransactionController{}

	// ensure at compile time that CoinDestructionTransactionController
	// implements the desired interfaces
	_ types.TransactionController      = CoinDestructionTransactionController{}
	_ types.TransactionSignatureHasher = CoinDestructionTransactionController{}
	_ types.TransactionIDEncoder       = CoinDestructionTransactionController{}

	// ensure at compile time that MinterDefinitionTransactionController
	// implements the desired interfaces
	_ types.TransactionController                = MinterDefinitionTransactionController{}
	_ types.TransactionExtensionSigner           = MinterDefinitionTransactionController{}
	_ types.TransactionSignatureHasher           = MinterDefinitionTransactionController{}
	_ types.TransactionIDEncoder                 = MinterDefinitionTransactionController{}
	_ types.TransactionCommonExtensionDataGetter = MinterDefinitionTransactionController{}
)

// CoinCreationTransactionController

// EncodeTransactionData implements TransactionController.EncodeTransactionData
func (cctc CoinCreationTransactionController) EncodeTransactionData(w io.Writer, txData types.TransactionData) error {
	cctx, err := CoinCreationTransactionFromTransactionData(txData, cctc.RequireMinerFees)
	if err != nil {
		return fmt.Errorf("failed to convert txData to a CoinCreationTx: %v", err)
	}
	return cctc.newBencoder(w).Encode(cctx)
}

// DecodeTransactionData implements TransactionController.DecodeTransactionData
func (cctc CoinCreationTransactionController) DecodeTransactionData(r io.Reader) (types.TransactionData, error) {
	var cctx CoinCreationTransaction
	err := cctc.newBdecoder(r).Decode(&cctx)
	if err != nil {
		return types.TransactionData{}, fmt.Errorf(
			"failed to binary-decode tx as a CoinCreationTx: %v", err)
	}
	// return coin creation tx as regular rivine tx data
	return cctx.TransactionData(), nil
}

// JSONEncodeTransactionData implements TransactionController.JSONEncodeTransactionData
func (cctc CoinCreationTransactionController) JSONEncodeTransactionData(txData types.TransactionData) ([]byte, error) {
	cctx, err := CoinCreationTransactionFromTransactionData(txData, cctc.RequireMinerFees)
	if err != nil {
		return nil, fmt.Errorf("failed to convert txData to a CoinCreationTx: %v", err)
	}
	return json.Marshal(cctx)
}

// JSONDecodeTransactionData implements TransactionController.JSONDecodeTransactionData
func (cctc CoinCreationTransactionController) JSONDecodeTransactionData(data []byte) (types.TransactionData, error) {
	var cctx CoinCreationTransaction
	err := json.Unmarshal(data, &cctx)
	if err != nil {
		return types.TransactionData{}, fmt.Errorf(
			"failed to json-decode tx as a CoinCreationTx: %v", err)
	}
	// return coin creation tx as regular rivine tx data
	return cctx.TransactionData(), nil
}

// SignExtension implements TransactionExtensionSigner.SignExtension
func (cctc CoinCreationTransactionController) SignExtension(extension interface{}, sign func(*types.UnlockFulfillmentProxy, types.UnlockConditionProxy, ...interface{}) error) (interface{}, error) {
	// (tx) extension (data) is expected to be a pointer to a valid CoinCreationTransactionExtension,
	// which contains the nonce and the mintFulfillment that can be used to fulfill the globally defined mint condition
	ccTxExtension, ok := extension.(*CoinCreationTransactionExtension)
	if !ok {
		return nil, errors.New("invalid extension data for a CoinCreationTransaction")
	}

	// get the active mint condition and use it to sign
	// NOTE: this does mean that if the mint condition suddenly this transaction will be invalid,
	// however given that only the minters (that create this coin transaction) can change the mint condition,
	// it is unlikely that this ever gives problems
	mintCondition, err := cctc.MintConditionGetter.GetActiveMintCondition()
	if err != nil {
		return nil, fmt.Errorf("failed to get the active mint condition: %v", err)
	}
	err = sign(&ccTxExtension.MintFulfillment, mintCondition)
	if err != nil {
		return nil, fmt.Errorf("failed to sign mint fulfillment of coin creation tx: %v", err)
	}
	return ccTxExtension, nil
}

// SignatureHash implements TransactionSignatureHasher.SignatureHash
func (cctc CoinCreationTransactionController) SignatureHash(t types.Transaction, extraObjects ...interface{}) (crypto.Hash, error) {
	cctx, err := CoinCreationTransactionFromTransaction(t, cctc.TransactionVersion, cctc.RequireMinerFees)
	if err != nil {
		return crypto.Hash{}, fmt.Errorf("failed to use tx as a coin creation tx: %v", err)
	}

	h := crypto.NewHash()
	enc := cctc.newBencoder(h)

	enc.EncodeAll(
		t.Version,
		SpecifierCoinCreationTransaction,
		cctx.Nonce,
	)

	if len(extraObjects) > 0 {
		enc.EncodeAll(extraObjects...)
	}

	enc.Encode(cctx.CoinOutputs)
	if cctc.RequireMinerFees {
		enc.Encode(cctx.MinerFees)
	}
	enc.Encode(cctx.ArbitraryData)

	var hash crypto.Hash
	h.Sum(hash[:0])
	return hash, nil
}

// EncodeTransactionIDInput implements TransactionIDEncoder.EncodeTransactionIDInput
func (cctc CoinCreationTransactionController) EncodeTransactionIDInput(w io.Writer, txData types.TransactionData) error {
	cctx, err := CoinCreationTransactionFromTransactionData(txData, cctc.RequireMinerFees)
	if err != nil {
		return fmt.Errorf("failed to convert txData to a CoinCreationTx: %v", err)
	}
	return cctc.newBencoder(w).EncodeAll(SpecifierCoinCreationTransaction, cctx)
}

// CoinDestructionTransactionController

// EncodeTransactionData implements TransactionController.EncodeTransactionData
func (cdtc CoinDestructionTransactionController) EncodeTransactionData(w io.Writer, txData types.TransactionData) error {
	cdtx, err := CoinDestructionTransactionFromTransactionData(txData)
	if err != nil {
		return fmt.Errorf("failed to convert txData to a CoinDestructionTx: %v", err)
	}
	return cdtc.newBencoder(w).Encode(cdtx)
}

// DecodeTransactionData implements TransactionController.DecodeTransactionData
func (cdtc CoinDestructionTransactionController) DecodeTransactionData(r io.Reader) (types.TransactionData, error) {
	var cdtx CoinDestructionTransaction
	err := cdtc.newBdecoder(r).Decode(&cdtx)
	if err != nil {
		return types.TransactionData{}, fmt.Errorf(
			"failed to binary-decode tx as a CoinDestructionTx: %v", err)
	}
	// return coin destruction tx as regular rivine tx data
	return cdtx.TransactionData(), nil
}

// JSONEncodeTransactionData implements TransactionController.JSONEncodeTransactionData
func (cdtc CoinDestructionTransactionController) JSONEncodeTransactionData(txData types.TransactionData) ([]byte, error) {
	cdtx, err := CoinDestructionTransactionFromTransactionData(txData)
	if err != nil {
		return nil, fmt.Errorf("failed to convert txData to a CoinDestructionTx: %v", err)
	}
	return json.Marshal(cdtx)
}

// JSONDecodeTransactionData implements TransactionController.JSONDecodeTransactionData
func (cdtc CoinDestructionTransactionController) JSONDecodeTransactionData(data []byte) (types.TransactionData, error) {
	var cdtx CoinDestructionTransaction
	err := json.Unmarshal(data, &cdtx)
	if err != nil {
		return types.TransactionData{}, fmt.Errorf(
			"failed to json-decode tx as a CoinDestructionTx: %v", err)
	}
	// return coin destruction tx as regular rivine tx data
	return cdtx.TransactionData(), nil
}

// SignatureHash implements TransactionSignatureHasher.SignatureHash
func (cdtc CoinDestructionTransactionController) SignatureHash(t types.Transaction, extraObjects ...interface{}) (crypto.Hash, error) {
	cdtx, err := CoinDestructionTransactionFromTransaction(t, cdtc.TransactionVersion)
	if err != nil {
		return crypto.Hash{}, fmt.Errorf("failed to use tx as a coin desruction tx: %v", err)
	}

	h := crypto.NewHash()
	enc := cdtc.newBencoder(h)

	enc.EncodeAll(
		t.Version,
		SpecifierCoinDestructionTransaction,
	)

	if len(extraObjects) > 0 {
		enc.EncodeAll(extraObjects...)
	}

	parentIDSlice := make([]types.CoinOutputID, 0, len(cdtx.CoinInputs))
	for _, ci := range cdtx.CoinInputs {
		parentIDSlice = append(parentIDSlice, ci.ParentID)
	}

	enc.EncodeAll(
		parentIDSlice,
		cdtx.CoinOutputs,
		cdtx.MinerFees,
		cdtx.ArbitraryData,
	)

	var hash crypto.Hash
	h.Sum(hash[:0])
	return hash, nil
}

// EncodeTransactionIDInput implements TransactionIDEncoder.EncodeTransactionIDInput
func (cdtc CoinDestructionTransactionController) EncodeTransactionIDInput(w io.Writer, txData types.TransactionData) error {
	cdtx, err := CoinDestructionTransactionFromTransactionData(txData)
	if err != nil {
		return fmt.Errorf("failed to convert txData to a CoinDestructionTx: %v", err)
	}
	return cdtc.newBencoder(w).EncodeAll(SpecifierCoinDestructionTransaction, cdtx)
}

// MinterDefinitionTransactionController

// EncodeTransactionData implements TransactionController.EncodeTransactionData
func (mdtc MinterDefinitionTransactionController) EncodeTransactionData(w io.Writer, txData types.TransactionData) error {
	mdtx, err := MinterDefinitionTransactionFromTransactionData(txData, mdtc.RequireMinerFees)
	if err != nil {
		return fmt.Errorf("failed to convert txData to a MinterDefinitionTx: %v", err)
	}
	return mdtc.newBencoder(w).Encode(mdtx)
}

// DecodeTransactionData implements TransactionController.DecodeTransactionData
func (mdtc MinterDefinitionTransactionController) DecodeTransactionData(r io.Reader) (types.TransactionData, error) {
	var mdtx MinterDefinitionTransaction
	err := mdtc.newBdecoder(r).Decode(&mdtx)
	if err != nil {
		return types.TransactionData{}, fmt.Errorf(
			"failed to binary-decode tx as a MinterDefinitionTx: %v", err)
	}
	// return minter definition tx as regular rivine tx data
	return mdtx.TransactionData(), nil
}

// JSONEncodeTransactionData implements TransactionController.JSONEncodeTransactionData
func (mdtc MinterDefinitionTransactionController) JSONEncodeTransactionData(txData types.TransactionData) ([]byte, error) {
	mdtx, err := MinterDefinitionTransactionFromTransactionData(txData, mdtc.RequireMinerFees)
	if err != nil {
		return nil, fmt.Errorf("failed to convert txData to a MinterDefinitionTx: %v", err)
	}
	return json.Marshal(mdtx)
}

// JSONDecodeTransactionData implements TransactionController.JSONDecodeTransactionData
func (mdtc MinterDefinitionTransactionController) JSONDecodeTransactionData(data []byte) (types.TransactionData, error) {
	var mdtx MinterDefinitionTransaction
	err := json.Unmarshal(data, &mdtx)
	if err != nil {
		return types.TransactionData{}, fmt.Errorf(
			"failed to json-decode tx as a MinterDefinitionTx: %v", err)
	}
	// return minter definition tx as regular rivine tx data
	return mdtx.TransactionData(), nil
}

// SignExtension implements TransactionExtensionSigner.SignExtension
func (mdtc MinterDefinitionTransactionController) SignExtension(extension interface{}, sign func(*types.UnlockFulfillmentProxy, types.UnlockConditionProxy, ...interface{}) error) (interface{}, error) {
	// (tx) extension (data) is expected to be a pointer to a valid MinterDefinitionTransactionExtension,
	// which contains the nonce and the mintFulfillment that can be used to fulfill the globally defined mint condition
	mdTxExtension, ok := extension.(*MinterDefinitionTransactionExtension)
	if !ok {
		return nil, errors.New("invalid extension data for a MinterDefinitionTx")
	}

	// get the active mint condition and use it to sign
	// NOTE: this does mean that if the mint condition suddenly this transaction will be invalid,
	// however given that only the minters (that create this coin transaction) can change the mint condition,
	// it is unlikely that this ever gives problems
	mintCondition, err := mdtc.MintConditionGetter.GetActiveMintCondition()
	if err != nil {
		return nil, fmt.Errorf("failed to get the active mint condition: %v", err)
	}
	err = sign(&mdTxExtension.MintFulfillment, mintCondition)
	if err != nil {
		return nil, fmt.Errorf("failed to sign mint fulfillment of MinterDefinitionTx: %v", err)
	}
	return mdTxExtension, nil
}

// SignatureHash implements TransactionSignatureHasher.SignatureHash
func (mdtc MinterDefinitionTransactionController) SignatureHash(t types.Transaction, extraObjects ...interface{}) (crypto.Hash, error) {
	mdtx, err := MinterDefinitionTransactionFromTransaction(t, mdtc.TransactionVersion, mdtc.RequireMinerFees)
	if err != nil {
		return crypto.Hash{}, fmt.Errorf("failed to use tx as a MinterDefinitionTx: %v", err)
	}

	h := crypto.NewHash()
	enc := mdtc.newBencoder(h)

	enc.EncodeAll(
		t.Version,
		SpecifierMintDefinitionTransaction,
		mdtx.Nonce,
	)

	if len(extraObjects) > 0 {
		enc.EncodeAll(extraObjects...)
	}

	enc.EncodeAll(
		mdtx.MintCondition,
		mdtx.MinerFees,
		mdtx.ArbitraryData,
	)

	var hash crypto.Hash
	h.Sum(hash[:0])
	return hash, nil
}

// EncodeTransactionIDInput implements TransactionIDEncoder.EncodeTransactionIDInput
func (mdtc MinterDefinitionTransactionController) EncodeTransactionIDInput(w io.Writer, txData types.TransactionData) error {
	mdtx, err := MinterDefinitionTransactionFromTransactionData(txData, mdtc.RequireMinerFees)
	if err != nil {
		return fmt.Errorf("failed to convert txData to a MinterDefinitionTx: %v", err)
	}
	return mdtc.newBencoder(w).EncodeAll(SpecifierMintDefinitionTransaction, mdtx)
}

// GetCommonExtensionData implements TransactionCommonExtensionDataGetter.GetCommonExtensionData
func (mdtc MinterDefinitionTransactionController) GetCommonExtensionData(extension interface{}) (types.CommonTransactionExtensionData, error) {
	mdext, ok := extension.(*MinterDefinitionTransactionExtension)
	if !ok {
		return types.CommonTransactionExtensionData{}, errors.New("invalid extension data for a MinterDefinitionTx")
	}
	return types.CommonTransactionExtensionData{
		UnlockConditions: []types.UnlockConditionProxy{mdext.MintCondition},
	}, nil
}

type (
	// CoinCreationTransaction is to be created only by the defined Coin Minters,
	// as a medium in order to create coins (coin outputs), without backing them
	// (so without having to spend previously unspend coin outputs, see: coin inputs).
	CoinCreationTransaction struct {
		// Nonce used to ensure the uniqueness of a CoinCreationTransaction's ID and signature.
		Nonce types.TransactionNonce `json:"nonce"`
		// MintFulfillment defines the fulfillment which is used in order to
		// fulfill the globally defined MintCondition.
		MintFulfillment types.UnlockFulfillmentProxy `json:"mintfulfillment"`
		// CoinOutputs defines the coin outputs,
		// which contain the freshly created coins, adding to the total pool of coins
		// available in the rivine network.
		CoinOutputs []types.CoinOutput `json:"coinoutputs"`
		// Minerfees, a fee paid for this coin creation transaction.
		MinerFees []types.Currency `json:"minerfees,omitempty"`
		// ArbitraryData can be used for any purpose,
		// but is mostly to be used in order to define the reason/origins
		// of the coin creation.
		ArbitraryData []byte `json:"arbitrarydata,omitempty"`
	}
	// CoinCreationTransactionExtension defines the CoinCreationTx Extension Data
	CoinCreationTransactionExtension struct {
		Nonce           types.TransactionNonce
		MintFulfillment types.UnlockFulfillmentProxy
	}
)

// CoinCreationTransactionFromTransaction creates a CoinCreationTransaction,
// using a regular in-memory rivine transaction.
//
// Past the (tx) Version validation it piggy-backs onto the
// `CoinCreationTransactionFromTransactionData` constructor.
func CoinCreationTransactionFromTransaction(tx types.Transaction, expectedVersion types.TransactionVersion, requireMinerFees bool) (CoinCreationTransaction, error) {
	if tx.Version != expectedVersion {
		return CoinCreationTransaction{}, fmt.Errorf(
			"a coin creation transaction requires tx version %d",
			expectedVersion)
	}
	return CoinCreationTransactionFromTransactionData(types.TransactionData{
		CoinInputs:        tx.CoinInputs,
		CoinOutputs:       tx.CoinOutputs,
		BlockStakeInputs:  tx.BlockStakeInputs,
		BlockStakeOutputs: tx.BlockStakeOutputs,
		MinerFees:         tx.MinerFees,
		ArbitraryData:     tx.ArbitraryData,
		Extension:         tx.Extension,
	}, requireMinerFees)
}

// CoinCreationTransactionFromTransactionData creates a CoinCreationTransaction,
// using the TransactionData from a regular in-memory rivine transaction.
func CoinCreationTransactionFromTransactionData(txData types.TransactionData, requireMinerFees bool) (CoinCreationTransaction, error) {
	// (tx) extension (data) is expected to be a pointer to a valid CoinCreationTransactionExtension,
	// which contains the nonce and the mintFulfillment that can be used to fulfill the globally defined mint condition
	extensionData, ok := txData.Extension.(*CoinCreationTransactionExtension)
	if !ok {
		return CoinCreationTransaction{}, errors.New("invalid extension data for a CoinCreationTransaction")
	}
	// at least one coin output as well as one miner fee is required
	if len(txData.CoinOutputs) == 0 {
		return CoinCreationTransaction{}, errors.New("at least one coin output is required for a CoinCreationTransaction")
	}
	if requireMinerFees && len(txData.MinerFees) == 0 {
		return CoinCreationTransaction{}, errors.New("at least one miner fee is required for a CoinCreationTransaction")
	} else if !requireMinerFees && len(txData.MinerFees) != 0 {
		return CoinCreationTransaction{}, errors.New("undesired miner fees: no miner fees are required, yet are defined")
	}
	// no coin inputs, block stake inputs or block stake outputs are allowed
	if len(txData.CoinInputs) != 0 || len(txData.BlockStakeInputs) != 0 || len(txData.BlockStakeOutputs) != 0 {
		return CoinCreationTransaction{}, errors.New("no coin inputs and block stake inputs/outputs are allowed in a CoinCreationTransaction")
	}
	// return the CoinCreationTransaction, with the data extracted from the TransactionData
	return CoinCreationTransaction{
		Nonce:           extensionData.Nonce,
		MintFulfillment: extensionData.MintFulfillment,
		CoinOutputs:     txData.CoinOutputs,
		MinerFees:       txData.MinerFees,
		// ArbitraryData is optional
		ArbitraryData: txData.ArbitraryData,
	}, nil
}

// TransactionData returns this CoinCreationTransaction
// as regular rivine transaction data.
func (cctx *CoinCreationTransaction) TransactionData() types.TransactionData {
	return types.TransactionData{
		CoinOutputs:   cctx.CoinOutputs,
		MinerFees:     cctx.MinerFees,
		ArbitraryData: cctx.ArbitraryData,
		Extension: &CoinCreationTransactionExtension{
			Nonce:           cctx.Nonce,
			MintFulfillment: cctx.MintFulfillment,
		},
	}
}

// Transaction returns this CoinCreationTransaction
// as regular rivine transaction, using TransactionVersionCoinCreation as the type.
func (cctx *CoinCreationTransaction) Transaction(version types.TransactionVersion) types.Transaction {
	return types.Transaction{
		Version:       version,
		CoinOutputs:   cctx.CoinOutputs,
		MinerFees:     cctx.MinerFees,
		ArbitraryData: cctx.ArbitraryData,
		Extension: &CoinCreationTransactionExtension{
			Nonce:           cctx.Nonce,
			MintFulfillment: cctx.MintFulfillment,
		},
	}
}

type (
	// CoinDestructionTransaction is to to be used by anyone
	// as a medium in order to destroy coins (coin outputs), partially or complete.
	CoinDestructionTransaction struct {
		// CoinInputs defines the coin outputs that are being spent.
		CoinInputs []types.CoinInput `json:"coininputs"`
		// CoinOutputs defines the coin outputs,
		// which contain the freshly created coins, adding to the total pool of coins
		// available in the rivine network.
		CoinOutputs []types.CoinOutput `json:"coinoutputs"`
		// Minerfees, a fee paid for this coin destruction transaction.
		MinerFees []types.Currency `json:"minerfees,omitempty"`
		// ArbitraryData can be used for any purpose,
		// but is mostly to be used in order to define the reason/origins
		// of the coin creation.
		ArbitraryData []byte `json:"arbitrarydata,omitempty"`
	}
)

// CoinDestructionTransactionFromTransaction creates a CoinDestructionTransaction,
// using a regular in-memory rivine transaction.
//
// Past the (tx) Version validation it piggy-backs onto the
// `CoinCreationTransactionFromTransactionData` constructor.
func CoinDestructionTransactionFromTransaction(tx types.Transaction, expectedVersion types.TransactionVersion) (CoinDestructionTransaction, error) {
	if tx.Version != expectedVersion {
		return CoinDestructionTransaction{}, fmt.Errorf(
			"a coin destruction transaction requires tx version %d",
			expectedVersion)
	}
	return CoinDestructionTransactionFromTransactionData(types.TransactionData{
		CoinInputs:        tx.CoinInputs,
		CoinOutputs:       tx.CoinOutputs,
		BlockStakeInputs:  tx.BlockStakeInputs,
		BlockStakeOutputs: tx.BlockStakeOutputs,
		MinerFees:         tx.MinerFees,
		ArbitraryData:     tx.ArbitraryData,
		Extension:         tx.Extension,
	})
}

// CoinDestructionTransactionFromTransactionData creates a CoinDestructionTransaction,
// using the TransactionData from a regular in-memory rivine transaction.
func CoinDestructionTransactionFromTransactionData(txData types.TransactionData) (CoinDestructionTransaction, error) {
	// no restrictions are there for coin outputs, but at least one miner fee is required
	if len(txData.MinerFees) == 0 {
		return CoinDestructionTransaction{}, errors.New("at least one miner fee is required for a CoinDestructionTransaction")
	}
	// at least one coin input is required
	if len(txData.CoinInputs) == 0 {
		return CoinDestructionTransaction{}, errors.New("at least one coin input is required for a CoinDestructionTransaction")
	}
	// no block stake inputs or block stake outputs are allowed
	if len(txData.BlockStakeInputs) != 0 || len(txData.BlockStakeOutputs) != 0 {
		return CoinDestructionTransaction{}, errors.New("no block stake inputs/outputs are allowed in a CoinDestructionTransaction")
	}
	// return the CoinDestructionTransaction, with the data extracted from the TransactionData
	return CoinDestructionTransaction{
		CoinInputs:    txData.CoinInputs,
		CoinOutputs:   txData.CoinOutputs,
		MinerFees:     txData.MinerFees,
		ArbitraryData: txData.ArbitraryData,
	}, nil
}

// TransactionData returns this CoinDestructionTransaction
// as regular rivine transaction data.
func (cdtx *CoinDestructionTransaction) TransactionData() types.TransactionData {
	return types.TransactionData{
		CoinInputs:    cdtx.CoinInputs,
		CoinOutputs:   cdtx.CoinOutputs,
		MinerFees:     cdtx.MinerFees,
		ArbitraryData: cdtx.ArbitraryData,
	}
}

// Transaction returns this CoinDestructionTransaction
// as regular rivine transaction, using TransactionVersionCoinCreation as the type.
func (cdtx *CoinDestructionTransaction) Transaction(version types.TransactionVersion) types.Transaction {
	return types.Transaction{
		Version:       version,
		CoinInputs:    cdtx.CoinInputs,
		CoinOutputs:   cdtx.CoinOutputs,
		MinerFees:     cdtx.MinerFees,
		ArbitraryData: cdtx.ArbitraryData,
	}
}

type (
	// MinterDefinitionTransaction is to be created only by the defined Coin Minters,
	// as a medium in order to transfer minting powers.
	MinterDefinitionTransaction struct {
		// Nonce used to ensure the uniqueness of a MinterDefinitionTransaction's ID and signature.
		Nonce types.TransactionNonce `json:"nonce"`
		// MintFulfillment defines the fulfillment which is used in order to
		// fulfill the globally defined MintCondition.
		MintFulfillment types.UnlockFulfillmentProxy `json:"mintfulfillment"`
		// MintCondition defines a new condition that defines who become(s) the new minter(s),
		// and thus defines who can create coins as well as update who is/are the current minter(s)
		//
		// UnlockHash (unlockhash type 1) and MultiSigConditions are allowed,
		// as well as TimeLocked conditions which have UnlockHash- and MultiSigConditions as
		// internal condition.
		MintCondition types.UnlockConditionProxy `json:"mintcondition"`
		// Minerfees, a fee paid for this minter definition transaction.
		MinerFees []types.Currency `json:"minerfees,omitempty"`
		// ArbitraryData can be used for any purpose,
		// but is mostly to be used in order to define the reason/origins
		// of the transfer of minting power.
		ArbitraryData []byte `json:"arbitrarydata,omitempty"`
	}
	// MinterDefinitionTransactionExtension defines the MinterDefinitionTx Extension Data
	MinterDefinitionTransactionExtension struct {
		Nonce           types.TransactionNonce
		MintFulfillment types.UnlockFulfillmentProxy
		MintCondition   types.UnlockConditionProxy
	}
)

// MinterDefinitionTransactionFromTransaction creates a MinterDefinitionTransaction,
// using a regular in-memory rivine transaction.
//
// Past the (tx) Version validation it piggy-backs onto the
// `MinterDefinitionTransactionFromTransactionData` constructor.
func MinterDefinitionTransactionFromTransaction(tx types.Transaction, expectedVersion types.TransactionVersion, requireMinerFees bool) (MinterDefinitionTransaction, error) {
	if tx.Version != expectedVersion {
		return MinterDefinitionTransaction{}, fmt.Errorf(
			"a minter definition transaction requires tx version %d",
			expectedVersion)
	}
	return MinterDefinitionTransactionFromTransactionData(types.TransactionData{
		CoinInputs:        tx.CoinInputs,
		CoinOutputs:       tx.CoinOutputs,
		BlockStakeInputs:  tx.BlockStakeInputs,
		BlockStakeOutputs: tx.BlockStakeOutputs,
		MinerFees:         tx.MinerFees,
		ArbitraryData:     tx.ArbitraryData,
		Extension:         tx.Extension,
	}, requireMinerFees)
}

// MinterDefinitionTransactionFromTransactionData creates a MinterDefinitionTransaction,
// using the TransactionData from a regular in-memory rivine transaction.
func MinterDefinitionTransactionFromTransactionData(txData types.TransactionData, requireMinerFees bool) (MinterDefinitionTransaction, error) {
	// (tx) extension (data) is expected to be a pointer to a valid MinterDefinitionTransactionExtension,
	// which contains the nonce, the mintFulfillment that can be used to fulfill the currently globally defined mint condition,
	// as well as a mintCondition to replace the current in-place mintCondition.
	extensionData, ok := txData.Extension.(*MinterDefinitionTransactionExtension)
	if !ok {
		return MinterDefinitionTransaction{}, errors.New("invalid extension data for a MinterDefinitionTransaction")
	}
	if requireMinerFees && len(txData.MinerFees) == 0 {
		return MinterDefinitionTransaction{}, errors.New("at least one miner fee is required for a MinterDefinitionTransaction")
	} else if !requireMinerFees && len(txData.MinerFees) != 0 {
		return MinterDefinitionTransaction{}, errors.New("undesired miner fees: no miner fees are required, yet are defined")
	}
	// no coin inputs, block stake inputs or block stake outputs are allowed
	if len(txData.CoinInputs) != 0 || len(txData.CoinOutputs) != 0 || len(txData.BlockStakeInputs) != 0 || len(txData.BlockStakeOutputs) != 0 {
		return MinterDefinitionTransaction{}, errors.New(
			"no coin inputs/outputs and block stake inputs/outputs are allowed in a MinterDefinitionTransaction")
	}
	// return the MinterDefinitionTransaction, with the data extracted from the TransactionData
	return MinterDefinitionTransaction{
		Nonce:           extensionData.Nonce,
		MintFulfillment: extensionData.MintFulfillment,
		MintCondition:   extensionData.MintCondition,
		MinerFees:       txData.MinerFees,
		// ArbitraryData is optional
		ArbitraryData: txData.ArbitraryData,
	}, nil
}

// TransactionData returns this CoinCreationTransaction
// as regular rivine transaction data.
func (cctx *MinterDefinitionTransaction) TransactionData() types.TransactionData {
	return types.TransactionData{
		MinerFees:     cctx.MinerFees,
		ArbitraryData: cctx.ArbitraryData,
		Extension: &MinterDefinitionTransactionExtension{
			Nonce:           cctx.Nonce,
			MintFulfillment: cctx.MintFulfillment,
			MintCondition:   cctx.MintCondition,
		},
	}
}

// Transaction returns this CoinCreationTransaction
// as regular rivine transaction, using TransactionVersionCoinCreation as the type.
func (cctx *MinterDefinitionTransaction) Transaction(version types.TransactionVersion) types.Transaction {
	return types.Transaction{
		Version:       version,
		MinerFees:     cctx.MinerFees,
		ArbitraryData: cctx.ArbitraryData,
		Extension: &MinterDefinitionTransactionExtension{
			Nonce:           cctx.Nonce,
			MintFulfillment: cctx.MintFulfillment,
			MintCondition:   cctx.MintCondition,
		},
	}
}
