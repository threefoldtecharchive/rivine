package authcointx

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/threefoldtech/rivine/crypto"
	"github.com/threefoldtech/rivine/pkg/encoding/rivbin"
	"github.com/threefoldtech/rivine/types"
)

// These Specifiers are used internally when calculating a Transaction's ID.
// See Rivine's Specifier for more details.
var (
	SpecifierAuthAddressUpdateTransaction   = types.Specifier{'a', 'u', 't', 'h', ' ', 'a', 'd', 'd', 'r', ' ', 'u', 'p', 'd', 'a', 't', 'e'}
	SpecifierAuthConditionUpdateTransaction = types.Specifier{'a', 'u', 't', 'h', ' ', 'c', 'o', 'n', 'd', ' ', 'u', 'p', 'd', 'a', 't', 'e'}
)

type (
	// AuthInfoGetter allows you to check if a list of addresses are authorized,
	// as well as what the auth condition is at a given (as well as current) [block] height.
	AuthInfoGetter interface {
		// GetActiveAuthCondition returns the active auth condition.
		// In other words, the auth condition at the current block height.
		GetActiveAuthCondition() (types.UnlockConditionProxy, error)
		// GetAuthConditionAt returns the auth condition at a given block height.
		GetAuthConditionAt(height types.BlockHeight) (types.UnlockConditionProxy, error)

		// Ensure that the given addresses are authorized at the current block height,
		// returning an address if an address is not authorized at the current block height.
		EnsureAddressesAreAuthNow(addresses ...types.UnlockHash) error
		// Ensure that the given addresses are authorized at the given block height,
		// returning an address if an address is not authorized at the given block height.
		EnsureAddressesAreAuthAt(height types.BlockHeight, addresses ...types.UnlockHash) error
	}
)

///////////////////////////////////////////////////////////////////////////////////
// TRANSACTION		///		Auth Address Update									///
///////////////////////////////////////////////////////////////////////////////////

type (
	// AuthAddressUpdateTransaction is to be used by the owner(s) of the Authorized Condition,
	// as a medium in order to (de)authorize address(es).
	//
	// /!\ This transaction requires NO Miner Fee.
	AuthAddressUpdateTransaction struct {
		// Nonce used to ensure the uniqueness of a AuthAddressUpdateTransaction's ID and signature.
		Nonce types.TransactionNonce `json:"nonce"`
		// AuthAddresses contains a list of addresses to be authorized,
		// it is also considered valid to authorize an address that is already authorized.
		AuthAddresses []types.UnlockHash `json:"authaddresses"`
		// AuthAddresses contains a list of addresses to be authorized,
		// it is also considered valid to deauthorize an address that has no authorization.
		DeauthAddresses []types.UnlockHash `json:"deauthaddresses"`
		// ArbitraryData can be used for any purpose
		ArbitraryData []byte `json:"arbitrarydata,omitempty"`
		// AuthFulfillment defines the fulfillment which is used in order to
		// fulfill the globally defined AuthCondition.
		AuthFulfillment types.UnlockFulfillmentProxy `json:"authfulfillment"`
	}
	// AuthAddressUpdateTransactionExtension defines the AuthAddressUpdateTransaction Extension Data
	AuthAddressUpdateTransactionExtension struct {
		Nonce           types.TransactionNonce
		AuthAddresses   []types.UnlockHash
		DeauthAddresses []types.UnlockHash
		AuthFulfillment types.UnlockFulfillmentProxy
	}
)

// AuthAddressUpdateTransactionFromTransaction creates a AuthAddressUpdateTransaction,
// using a regular in-memory rivine transaction.
//
// Past the (tx) Version validation it piggy-backs onto the
// `AuthAddressUpdateTransactionFromTransactionData` constructor.
func AuthAddressUpdateTransactionFromTransaction(tx types.Transaction, expectedVersion types.TransactionVersion) (AuthAddressUpdateTransaction, error) {
	if tx.Version != expectedVersion {
		return AuthAddressUpdateTransaction{}, fmt.Errorf(
			"an auth address update transaction requires tx version %d",
			expectedVersion)
	}
	return AuthAddressUpdateTransactionFromTransactionData(types.TransactionData{
		CoinInputs:        tx.CoinInputs,
		CoinOutputs:       tx.CoinOutputs,
		BlockStakeInputs:  tx.BlockStakeInputs,
		BlockStakeOutputs: tx.BlockStakeOutputs,
		MinerFees:         tx.MinerFees,
		ArbitraryData:     tx.ArbitraryData,
		Extension:         tx.Extension,
	})
}

// AuthAddressUpdateTransactionFromTransactionData creates a AuthAddressUpdateTransaction,
// using the TransactionData from a regular in-memory rivine transaction.
func AuthAddressUpdateTransactionFromTransactionData(txData types.TransactionData) (AuthAddressUpdateTransaction, error) {
	// (tx) extension (data) is expected to be a pointer to a valid AuthAddressUpdateTransactionExtension,
	// which contains all the non-standard information for this transaction type.
	extensionData, ok := txData.Extension.(*AuthAddressUpdateTransactionExtension)
	if !ok {
		return AuthAddressUpdateTransaction{}, errors.New("invalid extension data for a AuthAddressUpdateTransactionExtension")
	}
	// no coin inputs, miner fees, block stake inputs or block stake outputs are allowed
	if len(txData.CoinInputs) != 0 || len(txData.MinerFees) != 0 || len(txData.CoinOutputs) != 0 || len(txData.BlockStakeInputs) != 0 || len(txData.BlockStakeOutputs) != 0 {
		return AuthAddressUpdateTransaction{}, errors.New("no coin/blockstake inputs/outputs or miner fees are allowed in a AuthAddressUpdateTransaction")
	}
	// return the AuthAddressUpdateTransaction, with the data extracted from the TransactionData
	return AuthAddressUpdateTransaction{
		Nonce:           extensionData.Nonce,
		AuthAddresses:   extensionData.AuthAddresses,
		DeauthAddresses: extensionData.DeauthAddresses,
		ArbitraryData:   txData.ArbitraryData,
		AuthFulfillment: extensionData.AuthFulfillment,
	}, nil
}

// TransactionData returns this AuthAddressUpdateTransaction
// as regular rivine transaction data.
func (autx *AuthAddressUpdateTransaction) TransactionData() types.TransactionData {
	return types.TransactionData{
		ArbitraryData: autx.ArbitraryData,
		Extension: &AuthAddressUpdateTransactionExtension{
			Nonce:           autx.Nonce,
			AuthAddresses:   autx.AuthAddresses,
			DeauthAddresses: autx.DeauthAddresses,
			AuthFulfillment: autx.AuthFulfillment,
		},
	}
}

// Transaction returns this AuthAddressUpdateTransaction
// as regular rivine transaction, using AuthAddressUpdateTransaction as the type.
func (autx *AuthAddressUpdateTransaction) Transaction(version types.TransactionVersion) types.Transaction {
	return types.Transaction{
		Version:       version,
		ArbitraryData: autx.ArbitraryData,
		Extension: &AuthAddressUpdateTransactionExtension{
			Nonce:           autx.Nonce,
			AuthAddresses:   autx.AuthAddresses,
			DeauthAddresses: autx.DeauthAddresses,
			AuthFulfillment: autx.AuthFulfillment,
		},
	}
}

///////////////////////////////////////////////////////////////////////////////////
// TRANSACTION CONTROLLER	///		Auth Address Update							///
///////////////////////////////////////////////////////////////////////////////////

// ensures at compile time that the Auth Address Update Transaction Controller implement all desired interfaces
var (
	// ensure at compile time that AuthAddressUpdateTransactionController
	// implements the desired interfaces
	_ types.TransactionController      = AuthAddressUpdateTransactionController{}
	_ types.TransactionExtensionSigner = AuthAddressUpdateTransactionController{}
	_ types.TransactionValidator       = AuthAddressUpdateTransactionController{}
	_ types.CoinOutputValidator        = AuthAddressUpdateTransactionController{}
	_ types.BlockStakeOutputValidator  = AuthAddressUpdateTransactionController{}
	_ types.TransactionSignatureHasher = AuthAddressUpdateTransactionController{}
	_ types.TransactionIDEncoder       = AuthAddressUpdateTransactionController{}
)

type (
	// AuthAddressUpdateTransactionController defines a custom transaction controller,
	// for an Auth AddressUpdate Transaction. It allows the modification of the set
	// of addresses that are authorizes in order to receive or send coins.
	AuthAddressUpdateTransactionController struct {
		// AuthInfoGetter is used to get (coin) authorization information.
		AuthInfoGetter AuthInfoGetter

		// TransactionVersion is used to validate/set the transaction version
		// of an auth address update transaction.
		TransactionVersion types.TransactionVersion
	}
)

// EncodeTransactionData implements TransactionController.EncodeTransactionData
func (autc AuthAddressUpdateTransactionController) EncodeTransactionData(w io.Writer, txData types.TransactionData) error {
	autx, err := AuthAddressUpdateTransactionFromTransactionData(txData)
	if err != nil {
		return fmt.Errorf("failed to convert txData to a AuthAddressUpdateTx: %v", err)
	}
	return rivbin.NewEncoder(w).Encode(autx)
}

// DecodeTransactionData implements TransactionController.DecodeTransactionData
func (autc AuthAddressUpdateTransactionController) DecodeTransactionData(r io.Reader) (types.TransactionData, error) {
	var autx AuthAddressUpdateTransaction
	err := rivbin.NewDecoder(r).Decode(&autx)
	if err != nil {
		return types.TransactionData{}, fmt.Errorf(
			"failed to binary-decode tx as a AuthAddressUpdateTx: %v", err)
	}
	// return coin creation tx as regular rivine tx data
	return autx.TransactionData(), nil
}

// JSONEncodeTransactionData implements TransactionController.JSONEncodeTransactionData
func (autc AuthAddressUpdateTransactionController) JSONEncodeTransactionData(txData types.TransactionData) ([]byte, error) {
	autx, err := AuthAddressUpdateTransactionFromTransactionData(txData)
	if err != nil {
		return nil, fmt.Errorf("failed to convert txData to a AuthAddressUpdateTx: %v", err)
	}
	return json.Marshal(autx)
}

// JSONDecodeTransactionData implements TransactionController.JSONDecodeTransactionData
func (autc AuthAddressUpdateTransactionController) JSONDecodeTransactionData(data []byte) (types.TransactionData, error) {
	var autx AuthAddressUpdateTransaction
	err := json.Unmarshal(data, &autx)
	if err != nil {
		return types.TransactionData{}, fmt.Errorf(
			"failed to json-decode tx as a AuthAddressUpdateTx: %v", err)
	}
	// return coin creation tx as regular rivine tx data
	return autx.TransactionData(), nil
}

// SignExtension implements TransactionExtensionSigner.SignExtension
func (autc AuthAddressUpdateTransactionController) SignExtension(extension interface{}, sign func(*types.UnlockFulfillmentProxy, types.UnlockConditionProxy, ...interface{}) error) (interface{}, error) {
	// (tx) extension (data) is expected to be a pointer to a valid AuthAddressUpdateTransaction,
	// which contains the nonce and the mintFulfillment that can be used to fulfill the globally defined auth condition
	auTxExtension, ok := extension.(*AuthAddressUpdateTransaction)
	if !ok {
		return nil, errors.New("invalid extension data for a AuthAddressUpdateTransaction")
	}

	// get the active auth condition and use it to sign
	// NOTE: this does mean that if the mint condition suddenly this transaction will be invalid
	authCondition, err := autc.AuthInfoGetter.GetActiveAuthCondition()
	if err != nil {
		return nil, fmt.Errorf("failed to get the active auth condition: %v", err)
	}
	err = sign(&auTxExtension.AuthFulfillment, authCondition)
	if err != nil {
		return nil, fmt.Errorf("failed to sign auth fulfillment of auth address update tx: %v", err)
	}
	return auTxExtension, nil
}

// ValidateTransaction implements TransactionValidator.ValidateTransaction
func (autc AuthAddressUpdateTransactionController) ValidateTransaction(t types.Transaction, ctx types.ValidationContext, constants types.TransactionValidationConstants) (err error) {
	err = types.TransactionFitsInABlock(t, constants.BlockSizeLimit)
	if err != nil {
		return err
	}

	// get AuthAddressUpdateTx
	autx, err := AuthAddressUpdateTransactionFromTransaction(t, autc.TransactionVersion)
	if err != nil {
		// this check also fails if the tx contains coin/blockstake inputs/outputs or miner fees
		return fmt.Errorf("failed to use tx as a auth address update tx: %v", err)
	}

	// get AuthCondition
	authCondition, err := autc.AuthInfoGetter.GetAuthConditionAt(ctx.BlockHeight)
	if err != nil {
		return fmt.Errorf("failed to get auth condition at block height %d: %v", ctx.BlockHeight, err)
	}

	// check if AuthFulfillment fulfills the Globally defined AuthCondition for the context-defined block height
	err = authCondition.Fulfill(autx.AuthFulfillment, types.FulfillContext{
		BlockHeight: ctx.BlockHeight,
		BlockTime:   ctx.BlockTime,
		Transaction: t,
	})
	if err != nil {
		return fmt.Errorf("failed to fulfill mint condition: %v", err)
	}
	// ensure the Nonce is not Nil
	if autx.Nonce == (types.TransactionNonce{}) {
		return errors.New("nil nonce is not allowed for a auth address update transaction")
	}

	// validate the rest of the content
	err = types.ArbitraryDataFits(autx.ArbitraryData, constants.ArbitraryDataSizeLimit)
	if err != nil {
		return
	}

	// ensure we have at least one address to (de)authorize
	if len(autx.AuthAddresses) == 0 && len(autx.DeauthAddresses) == 0 {
		err = errors.New("at least one address is required to be authorized or deauthorized")
	}
	// ensure all addresses are unique and no address is both authorized and deauthorized
	addressesSeen := map[types.UnlockHash]struct{}{}
	var ok bool
	for _, address := range autx.AuthAddresses {
		if _, ok = addressesSeen[address]; ok {
			err = fmt.Errorf("an address can only be defined once per AuthAddressUpdate transaction: %s was seen twice", address.String())
			return
		}
		addressesSeen[address] = struct{}{}
	}
	for _, address := range autx.DeauthAddresses {
		if _, ok = addressesSeen[address]; ok {
			err = fmt.Errorf("an address can only be defined once per AuthAddressUpdate transaction: %s was seen twice", address.String())
			return
		}
		addressesSeen[address] = struct{}{}
	}

	// no error to return
	return nil
}

// ValidateCoinOutputs implements CoinOutputValidator.ValidateCoinOutputs
func (autc AuthAddressUpdateTransactionController) ValidateCoinOutputs(t types.Transaction, ctx types.FundValidationContext, coinInputs map[types.CoinOutputID]types.CoinOutput) (err error) {
	return nil // always valid, no coin inputs/outputs exist within a coin creation transaction
}

// ValidateBlockStakeOutputs implements BlockStakeOutputValidator.ValidateBlockStakeOutputs
func (autc AuthAddressUpdateTransactionController) ValidateBlockStakeOutputs(t types.Transaction, ctx types.FundValidationContext, blockStakeInputs map[types.BlockStakeOutputID]types.BlockStakeOutput) (err error) {
	return nil // always valid, no block stake inputs/outputs exist within a coin creation transaction
}

// SignatureHash implements TransactionSignatureHasher.SignatureHash
func (autc AuthAddressUpdateTransactionController) SignatureHash(t types.Transaction, extraObjects ...interface{}) (crypto.Hash, error) {
	autx, err := AuthAddressUpdateTransactionFromTransaction(t, autc.TransactionVersion)
	if err != nil {
		return crypto.Hash{}, fmt.Errorf("failed to use tx as an auth update tx: %v", err)
	}

	h := crypto.NewHash()
	enc := rivbin.NewEncoder(h)

	enc.EncodeAll(
		t.Version,
		SpecifierAuthAddressUpdateTransaction,
		autx.Nonce,
	)

	if len(extraObjects) > 0 {
		enc.EncodeAll(extraObjects...)
	}

	enc.EncodeAll(
		autx.AuthAddresses,
		autx.DeauthAddresses,
		autx.ArbitraryData,
	)

	var hash crypto.Hash
	h.Sum(hash[:0])
	return hash, nil
}

// EncodeTransactionIDInput implements TransactionIDEncoder.EncodeTransactionIDInput
func (autc AuthAddressUpdateTransactionController) EncodeTransactionIDInput(w io.Writer, txData types.TransactionData) error {
	autx, err := AuthAddressUpdateTransactionFromTransactionData(txData)
	if err != nil {
		return fmt.Errorf("failed to convert txData to a AuthAddressUpdateTx: %v", err)
	}
	return rivbin.NewEncoder(w).EncodeAll(SpecifierAuthAddressUpdateTransaction, autx)
}

///////////////////////////////////////////////////////////////////////////////////
// TRANSACTION		///		Auth Condition Update								///
///////////////////////////////////////////////////////////////////////////////////

type (
	// AuthConditionUpdateTransaction is to be used by the owner(s) of the Authorized Condition,
	// as a medium in order transfer the authorization power to a new condition.
	//
	// /!\ This transaction requires NO Miner Fee.
	AuthConditionUpdateTransaction struct {
		// Nonce used to ensure the uniqueness of a AuthConditionUpdateTransaction's ID and signature.
		Nonce types.TransactionNonce `json:"nonce"`
		// ArbitraryData can be used for any purpose
		ArbitraryData []byte `json:"arbitrarydata,omitempty"`
		// AuthCondition defines the condition which will have to fulfilled
		// in order to prove powers of authority when changing power (again) or using those powers to update
		// the set of authorized addresses.
		AuthCondition types.UnlockConditionProxy `json:"authcondition"`
		// AuthFulfillment defines the fulfillment which is used in order to
		// fulfill the globally defined AuthCondition.
		AuthFulfillment types.UnlockFulfillmentProxy `json:"authfulfillment"`
	}
	// AuthConditionUpdateTransactionExtension defines the AuthConditionUpdateTransaction Extension Data
	AuthConditionUpdateTransactionExtension struct {
		Nonce           types.TransactionNonce
		AuthCondition   types.UnlockConditionProxy
		AuthFulfillment types.UnlockFulfillmentProxy
	}
)

// AuthConditionUpdateTransactionFromTransaction creates a AuthConditionUpdateTransaction,
// using a regular in-memory rivine transaction.
//
// Past the (tx) Version validation it piggy-backs onto the
// `AuthAddressUpdateTransactionFromTransactionData` constructor.
func AuthConditionUpdateTransactionFromTransaction(tx types.Transaction, expectedVersion types.TransactionVersion) (AuthConditionUpdateTransaction, error) {
	if tx.Version != expectedVersion {
		return AuthConditionUpdateTransaction{}, fmt.Errorf(
			"an auth condition update transaction requires tx version %d",
			expectedVersion)
	}
	return AuthConditionUpdateTransactionFromTransactionData(types.TransactionData{
		CoinInputs:        tx.CoinInputs,
		CoinOutputs:       tx.CoinOutputs,
		BlockStakeInputs:  tx.BlockStakeInputs,
		BlockStakeOutputs: tx.BlockStakeOutputs,
		MinerFees:         tx.MinerFees,
		ArbitraryData:     tx.ArbitraryData,
		Extension:         tx.Extension,
	})
}

// AuthConditionUpdateTransactionFromTransactionData creates a AuthConditionUpdateTransaction,
// using the TransactionData from a regular in-memory rivine transaction.
func AuthConditionUpdateTransactionFromTransactionData(txData types.TransactionData) (AuthConditionUpdateTransaction, error) {
	// (tx) extension (data) is expected to be a pointer to a valid AuthAddressUpdateTransactionExtension,
	// which contains all the non-standard information for this transaction type.
	extensionData, ok := txData.Extension.(*AuthConditionUpdateTransactionExtension)
	if !ok {
		return AuthConditionUpdateTransaction{}, errors.New("invalid extension data for a AuthConditionUpdateTransactionExtension")
	}
	// no coin inputs, miner fees, block stake inputs or block stake outputs are allowed
	if len(txData.CoinInputs) != 0 || len(txData.MinerFees) != 0 || len(txData.CoinOutputs) != 0 || len(txData.BlockStakeInputs) != 0 || len(txData.BlockStakeOutputs) != 0 {
		return AuthConditionUpdateTransaction{}, errors.New("no coin/blockstake inputs/outputs or miner fees are allowed in a AuthConditionUpdateTransaction")
	}
	// return the AuthConditionUpdateTransaction, with the data extracted from the TransactionData
	return AuthConditionUpdateTransaction{
		Nonce:           extensionData.Nonce,
		ArbitraryData:   txData.ArbitraryData,
		AuthCondition:   extensionData.AuthCondition,
		AuthFulfillment: extensionData.AuthFulfillment,
	}, nil
}

// TransactionData returns this AuthAddressUpdateTransaction
// as regular rivine transaction data.
func (autx *AuthConditionUpdateTransaction) TransactionData() types.TransactionData {
	return types.TransactionData{
		ArbitraryData: autx.ArbitraryData,
		Extension: &AuthConditionUpdateTransactionExtension{
			Nonce:           autx.Nonce,
			AuthCondition:   autx.AuthCondition,
			AuthFulfillment: autx.AuthFulfillment,
		},
	}
}

// Transaction returns this AuthAddressUpdateTransaction
// as regular rivine transaction, using AuthAddressUpdateTransaction as the type.
func (autx *AuthConditionUpdateTransaction) Transaction(version types.TransactionVersion) types.Transaction {
	return types.Transaction{
		Version:       version,
		ArbitraryData: autx.ArbitraryData,
		Extension: &AuthConditionUpdateTransactionExtension{
			Nonce:           autx.Nonce,
			AuthCondition:   autx.AuthCondition,
			AuthFulfillment: autx.AuthFulfillment,
		},
	}
}

///////////////////////////////////////////////////////////////////////////////////
// TRANSACTION CONTROLLER	///		Auth Condition Update						///
///////////////////////////////////////////////////////////////////////////////////

// ensures at compile time that the Auth Condition Update Transaction Controller implement all desired interfaces
var (
	// ensure at compile time that AuthConditionUpdateTransactionController
	// implements the desired interfaces
	_ types.TransactionController      = AuthConditionUpdateTransactionController{}
	_ types.TransactionExtensionSigner = AuthConditionUpdateTransactionController{}
	_ types.TransactionValidator       = AuthConditionUpdateTransactionController{}
	_ types.CoinOutputValidator        = AuthConditionUpdateTransactionController{}
	_ types.BlockStakeOutputValidator  = AuthConditionUpdateTransactionController{}
	_ types.TransactionSignatureHasher = AuthConditionUpdateTransactionController{}
	_ types.TransactionIDEncoder       = AuthConditionUpdateTransactionController{}
)

type (
	// AuthConditionUpdateTransactionController defines a custom transaction controller,
	// for an Auth ConditionUpdate Transaction. It allows the modification of the set
	// of addresses that are authorizes in order to receive or send coins.
	AuthConditionUpdateTransactionController struct {
		// AuthInfoGetter is used to get (coin) authorization information.
		AuthInfoGetter AuthInfoGetter

		// TransactionVersion is used to validate/set the transaction version
		// of an auth address update transaction.
		TransactionVersion types.TransactionVersion
	}
)

// EncodeTransactionData implements TransactionController.EncodeTransactionData
func (cutc AuthConditionUpdateTransactionController) EncodeTransactionData(w io.Writer, txData types.TransactionData) error {
	cutx, err := AuthConditionUpdateTransactionFromTransactionData(txData)
	if err != nil {
		return fmt.Errorf("failed to convert txData to a AuthConditionUpdateTx: %v", err)
	}
	return rivbin.NewEncoder(w).Encode(cutx)
}

// DecodeTransactionData implements TransactionController.DecodeTransactionData
func (cutc AuthConditionUpdateTransactionController) DecodeTransactionData(r io.Reader) (types.TransactionData, error) {
	var cutx AuthConditionUpdateTransaction
	err := rivbin.NewDecoder(r).Decode(&cutx)
	if err != nil {
		return types.TransactionData{}, fmt.Errorf(
			"failed to binary-decode tx as a AuthConditionUpdateTx: %v", err)
	}
	// return auth condition update tx as regular rivine tx data
	return cutx.TransactionData(), nil
}

// JSONEncodeTransactionData implements TransactionController.JSONEncodeTransactionData
func (cutc AuthConditionUpdateTransactionController) JSONEncodeTransactionData(txData types.TransactionData) ([]byte, error) {
	autx, err := AuthAddressUpdateTransactionFromTransactionData(txData)
	if err != nil {
		return nil, fmt.Errorf("failed to convert txData to a AuthAddressUpdateTx: %v", err)
	}
	return json.Marshal(autx)
}

// JSONDecodeTransactionData implements TransactionController.JSONDecodeTransactionData
func (cutc AuthConditionUpdateTransactionController) JSONDecodeTransactionData(data []byte) (types.TransactionData, error) {
	var cutx AuthConditionUpdateTransaction
	err := json.Unmarshal(data, &cutx)
	if err != nil {
		return types.TransactionData{}, fmt.Errorf(
			"failed to json-decode tx as a AuthConditionUpdateTransaction: %v", err)
	}
	// return coin creation tx as regular rivine tx data
	return cutx.TransactionData(), nil
}

// SignExtension implements TransactionExtensionSigner.SignExtension
func (cutc AuthConditionUpdateTransactionController) SignExtension(extension interface{}, sign func(*types.UnlockFulfillmentProxy, types.UnlockConditionProxy, ...interface{}) error) (interface{}, error) {
	// (tx) extension (data) is expected to be a pointer to a valid AuthConditionUpdateTransaction,
	// which contains the nonce and the mintFulfillment that can be used to fulfill the globally defined auth condition
	cuTxExtension, ok := extension.(*AuthConditionUpdateTransactionExtension)
	if !ok {
		return nil, errors.New("invalid extension data for a AuthConditionUpdateTransaction")
	}

	// get the active auth condition and use it to sign
	// NOTE: this does mean that if the mint condition suddenly this transaction will be invalid
	authCondition, err := cutc.AuthInfoGetter.GetActiveAuthCondition()
	if err != nil {
		return nil, fmt.Errorf("failed to get the active auth condition: %v", err)
	}
	err = sign(&cuTxExtension.AuthFulfillment, authCondition)
	if err != nil {
		return nil, fmt.Errorf("failed to sign auth fulfillment of auth address update tx: %v", err)
	}
	return cuTxExtension, nil
}

// ValidateTransaction implements TransactionValidator.ValidateTransaction
func (cutc AuthConditionUpdateTransactionController) ValidateTransaction(t types.Transaction, ctx types.ValidationContext, constants types.TransactionValidationConstants) (err error) {
	err = types.TransactionFitsInABlock(t, constants.BlockSizeLimit)
	if err != nil {
		return err
	}

	// get AuthConditionUpdateTx
	cutx, err := AuthConditionUpdateTransactionFromTransaction(t, cutc.TransactionVersion)
	if err != nil {
		// this check also fails if the tx contains coin/blockstake inputs/outputs or miner fees
		return fmt.Errorf("failed to use tx as a auth condition update tx: %v", err)
	}

	// get AuthCondition
	authCondition, err := cutc.AuthInfoGetter.GetAuthConditionAt(ctx.BlockHeight)
	if err != nil {
		return fmt.Errorf("failed to get auth condition at block height %d: %v", ctx.BlockHeight, err)
	}

	// check if AuthFulfillment fulfills the Globally defined AuthCondition for the context-defined block height
	err = authCondition.Fulfill(cutx.AuthFulfillment, types.FulfillContext{
		BlockHeight: ctx.BlockHeight,
		BlockTime:   ctx.BlockTime,
		Transaction: t,
	})
	if err != nil {
		return fmt.Errorf("failed to fulfill mint condition: %v", err)
	}
	// ensure the Nonce is not Nil
	if cutx.Nonce == (types.TransactionNonce{}) {
		return errors.New("nil nonce is not allowed for a auth address update transaction")
	}

	// validate the rest of the content
	err = types.ArbitraryDataFits(cutx.ArbitraryData, constants.ArbitraryDataSizeLimit)
	if err != nil {
		return
	}

	// ensure the defined condition is not equal to the current active auth condition
	if authCondition.Equal(cutx.AuthCondition) {
		return errors.New("defined condition is already used as the currently active auth condition (nop update not allowed)")
	}
	// ensure the defined condition maps to an acceptable uh
	uh := cutx.AuthCondition.UnlockHash()
	if uh.Type != types.UnlockTypePubKey && uh.Type != types.UnlockTypeMultiSig {
		return fmt.Errorf("defined condition maps to an invalid unlock hash type %d", uh.Type)
	}

	// no error to return
	return nil
}

// ValidateCoinOutputs implements CoinOutputValidator.ValidateCoinOutputs
func (cutc AuthConditionUpdateTransactionController) ValidateCoinOutputs(t types.Transaction, ctx types.FundValidationContext, coinInputs map[types.CoinOutputID]types.CoinOutput) (err error) {
	return nil // always valid, no coin inputs/outputs exist within a coin creation transaction
}

// ValidateBlockStakeOutputs implements BlockStakeOutputValidator.ValidateBlockStakeOutputs
func (cutc AuthConditionUpdateTransactionController) ValidateBlockStakeOutputs(t types.Transaction, ctx types.FundValidationContext, blockStakeInputs map[types.BlockStakeOutputID]types.BlockStakeOutput) (err error) {
	return nil // always valid, no block stake inputs/outputs exist within a coin creation transaction
}

// SignatureHash implements TransactionSignatureHasher.SignatureHash
func (cutc AuthConditionUpdateTransactionController) SignatureHash(t types.Transaction, extraObjects ...interface{}) (crypto.Hash, error) {
	cutx, err := AuthConditionUpdateTransactionFromTransaction(t, cutc.TransactionVersion)
	if err != nil {
		return crypto.Hash{}, fmt.Errorf("failed to use tx as an auth condition update tx: %v", err)
	}

	h := crypto.NewHash()
	enc := rivbin.NewEncoder(h)

	enc.EncodeAll(
		t.Version,
		SpecifierAuthConditionUpdateTransaction,
		cutx.Nonce,
	)

	if len(extraObjects) > 0 {
		enc.EncodeAll(extraObjects...)
	}

	enc.EncodeAll(
		cutx.AuthCondition,
		cutx.ArbitraryData,
	)

	var hash crypto.Hash
	h.Sum(hash[:0])
	return hash, nil
}

// EncodeTransactionIDInput implements TransactionIDEncoder.EncodeTransactionIDInput
func (cutc AuthConditionUpdateTransactionController) EncodeTransactionIDInput(w io.Writer, txData types.TransactionData) error {
	autx, err := AuthConditionUpdateTransactionFromTransactionData(txData)
	if err != nil {
		return fmt.Errorf("failed to convert txData to a AuthConditionUpdateTx: %v", err)
	}
	return rivbin.NewEncoder(w).EncodeAll(SpecifierAuthConditionUpdateTransaction, autx)
}

///////////////////////////////////////////////////////////////////////////////////
// TRANSACTION	CONTROLLER	///		Auth Standard Transaction					///
///////////////////////////////////////////////////////////////////////////////////

// ensures at compile time that the Auth Condition Update Transaction Controller implement all desired interfaces
var (
	// ensure at compile time that AuthConditionUpdateTransactionController
	// implements the desired interfaces
	_ types.CoinOutputValidator = AuthStandardTransferTransactionController{}
)

type (
	// AuthStandardTransferTransactionController defines a custom transaction controller,
	// for the regular Rivine (0x01) transaction. It adds validation to the transaction,
	// as well as coin outputs to ensure that all coin inputs/outputs are from/to authorized addresses.
	AuthStandardTransferTransactionController struct {
		types.DefaultTransactionController

		// AuthInfoGetter is used to get (coin) authorization information.
		AuthInfoGetter AuthInfoGetter
	}
)

// ValidateCoinOutputs implements CoinOutputValidator.ValidateCoinOutputs
func (sttc AuthStandardTransferTransactionController) ValidateCoinOutputs(t types.Transaction, ctx types.FundValidationContext, coinInputs map[types.CoinOutputID]types.CoinOutput) error {
	// collect all dedupAddresses
	dedupAddresses := map[types.UnlockHash]struct{}{}
	for _, co := range t.CoinOutputs {
		dedupAddresses[co.Condition.UnlockHash()] = struct{}{}
	}
	for _, co := range coinInputs {
		dedupAddresses[co.Condition.UnlockHash()] = struct{}{}
	}
	addressLength := len(dedupAddresses)
	if addressLength == 0 {
		return nil // early return if nothing to do
	}
	addresses := make([]types.UnlockHash, 0, len(dedupAddresses))
	for uh := range dedupAddresses {
		addresses = append(addresses, uh)
	}

	// validate them all at once
	err := sttc.AuthInfoGetter.EnsureAddressesAreAuthAt(ctx.BlockHeight, addresses...)
	if err != nil {
		return fmt.Errorf("unauthorized address(s) cannot participate in a coin transfer: %v", err)
	}
	return nil
}
