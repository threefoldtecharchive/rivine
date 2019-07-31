package types

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/threefoldtech/rivine/crypto"
	"github.com/threefoldtech/rivine/pkg/encoding/rivbin"
	"github.com/threefoldtech/rivine/pkg/encoding/siabin"
)

type (
	// TransactionController is the required interface that has to be implemented
	// for each registered transaction version.
	//
	// Besides the required interface,
	// a transaction controller can also implement one or multiple
	// supported extension interfaces supported by this lib.
	TransactionController interface {
		// EncodeTransactionData binary-encodes the transaction data,
		// which is all transaction properties except for the version.
		EncodeTransactionData(io.Writer, TransactionData) error
		// DecodeTransactionData binary-decodes the transaction data,
		// which is all transaction properties except for the version.
		DecodeTransactionData(io.Reader) (TransactionData, error)

		// JSONEncodeTransactionData JSON-encodes the transaction data,
		// which is all transaction properties except for the version.
		JSONEncodeTransactionData(TransactionData) ([]byte, error)
		// JSONDecodeTransactionData JSON-decodes the transaction data,
		// which is all transaction properties except for the version.
		JSONDecodeTransactionData([]byte) (TransactionData, error)

		// ...
		//   any other supported and optional interface method
	}

	// TransactionData contains all the core data of a transaction,
	// which is all data except for the version.
	//
	// Extension is never binary/json encoded/decoded.
	TransactionData struct {
		CoinInputs        []CoinInput        `json:"coininputs"` // required
		CoinOutputs       []CoinOutput       `json:"coinoutputs,omitempty"`
		BlockStakeInputs  []BlockStakeInput  `json:"blockstakeinputs,omitempty"`
		BlockStakeOutputs []BlockStakeOutput `json:"blockstakeoutputs,omitempty"`
		MinerFees         []Currency         `json:"minerfees"` // required
		ArbitraryData     []byte             `json:"arbitrarydata,omitempty"`

		// Extension is an optional field that can be used,
		// in order to attach non-standard state to a transaction.
		// It is used for data only, controller logic is all to be implemented
		// as extended interfaces of the (transaction) controller.
		Extension interface{} `json:"-"` // omitted from JSON
	}

	// TransactionValidationConstants defines the contants that a TransactionValidator
	// can use in order to validate the transaction, within its local scope.
	TransactionValidationConstants struct {
		BlockSizeLimit         uint64
		ArbitraryDataSizeLimit uint64
		MinimumMinerFee        Currency
	}
)

// optional interfaces which a TransactionController can implement as well,
// in order to customize a version even more
type (
	// TransactionSignatureHasher defines the interface a transaction controller
	// can optionally implement, in order to define custom Tx signatures,
	// overwriting the default Tx sig hash logic.
	TransactionSignatureHasher interface {
		SignatureHash(t Transaction, extraObjects ...interface{}) (crypto.Hash, error)
	}

	// TransactionIDEncoder is an optional interface a transaction controller
	// can implement, in order to use a different binary encoding for ID-generation purposes,
	// instead of using the default binary encoding logic for that transaction (version).
	TransactionIDEncoder interface {
		EncodeTransactionIDInput(io.Writer, TransactionData) error
	}

	// TransactionExtensionSigner defines an interface for transactions which have fulfillments in the
	// extension part of the data that have to be signed as well.
	TransactionExtensionSigner interface {
		// SignExtension allows the transaction to sign —using the given sign callback—
		// any fulfillment (giving its condition as reference) that has to be signed.
		SignExtension(extension interface{}, sign func(*UnlockFulfillmentProxy, UnlockConditionProxy, ...interface{}) error) (interface{}, error)
	}

	// TransactionCustomMinerPayoutGetter defines an interface for transactions which have
	// custom MinerPayouts, stored in its extension data, that are not seen as regular Miner Fees.
	TransactionCustomMinerPayoutGetter interface {
		// GetCustomMinerPayouts allows a transaction controller to extract
		// MinerPayouts orginating from the transaction's extension data,
		// for any miner payouts to be added to the parent block,
		// which is not considered as regular MinerFees
		// (and thus is not simply defined in the Transaction's MinerFees property).
		GetCustomMinerPayouts(extension interface{}) ([]MinerPayout, error)
	}

	// TransactionCommonExtensionDataGetter defines an interface for transactions which have
	// common-understood data in the Extension data, allowing Rivine code to extract this generic extension data,
	// without having to know about the actual format/structure of this Tx.
	TransactionCommonExtensionDataGetter interface {
		// GetCommonExtensionData allows a transaction controllor to extract
		// the common-understood data from the Extension data for consumption by the callee.
		GetCommonExtensionData(extension interface{}) (CommonTransactionExtensionData, error)
	}
)

// RegisterTransactionVersion registers or unregisters a given transaction version,
// by attaching a controller to it that helps define the behavior surrounding its state.
//
// NOTE: this function should only be called in the `init` func,
// or at the very least prior to starting to create the daemon server,
// doing it anywhere else can result in undefined behavior,
func RegisterTransactionVersion(v TransactionVersion, c TransactionController) {
	if c == nil {
		delete(_RegisteredTransactionVersions, v)
		return
	}
	_RegisteredTransactionVersions[v] = c
}

var (
	// ErrUnexpectedExtensionType is an error returned by a transaction controller,
	// in case it expects an extension type it didn't expect.
	ErrUnexpectedExtensionType = errors.New("unexpected transaction data extension type")
)

var (
	_RegisteredTransactionVersions = map[TransactionVersion]TransactionController{}
)

// MarshalSia implements siabin.SiaMarshaller.MarshalSia
func (td TransactionData) MarshalSia(w io.Writer) error {
	return siabin.NewEncoder(w).EncodeAll(
		td.CoinInputs, td.CoinOutputs,
		td.BlockStakeInputs, td.BlockStakeOutputs,
		td.MinerFees, td.ArbitraryData)
}

// UnmarshalSia implements siabin.SiaUnmarshaller.UnmarshalSia
func (td *TransactionData) UnmarshalSia(r io.Reader) error {
	return siabin.NewDecoder(r).DecodeAll(
		&td.CoinInputs, &td.CoinOutputs,
		&td.BlockStakeInputs, &td.BlockStakeOutputs,
		&td.MinerFees, &td.ArbitraryData)
}

// MarshalRivine implements rivbin.RivineMarshaler.MarshalRivine
func (td TransactionData) MarshalRivine(w io.Writer) error {
	return rivbin.NewEncoder(w).EncodeAll(
		td.CoinInputs, td.CoinOutputs,
		td.BlockStakeInputs, td.BlockStakeOutputs,
		td.MinerFees, td.ArbitraryData)
}

// UnmarshalRivine implements rivbin.RivineMarshaler.UnmarshalRivine
func (td *TransactionData) UnmarshalRivine(r io.Reader) error {
	return rivbin.NewDecoder(r).DecodeAll(
		&td.CoinInputs, &td.CoinOutputs,
		&td.BlockStakeInputs, &td.BlockStakeOutputs,
		&td.MinerFees, &td.ArbitraryData)
}

// Standard Transaction Controller implementations
type (
	// DefaultTransactionController is the default transaction controller used,
	// and is also by default the controller for the default transaction version 0x01.
	DefaultTransactionController struct{}

	// LegacyTransactionController is a legacy transaction controller,
	// which used to be the default when Rivine launched.
	// It should however not be used any longer, and only exists,
	// as to support chains which launched together with Rivine.
	LegacyTransactionController struct{}
)

// EncodeTransactionData implements TransactionController.EncodeTransactionData
func (dtc DefaultTransactionController) EncodeTransactionData(w io.Writer, td TransactionData) error {
	// encode to a byte slice first
	b, err := siabin.Marshal(td)
	if err != nil {
		return fmt.Errorf("failed to (siabin) marshal transaction data: %v", err)
	}
	// copy those bytes together with its prefixed length, as the final encoding
	return siabin.NewEncoder(w).Encode(b)
}

// DecodeTransactionData implements TransactionController.DecodeTransactionData
func (dtc DefaultTransactionController) DecodeTransactionData(r io.Reader) (td TransactionData, err error) {
	// decode it as a byte slice first
	var b []byte
	err = siabin.NewDecoder(r).Decode(&b)
	if err != nil {
		return
	}
	// decode
	err = siabin.Unmarshal(b, &td)
	return
}

// JSONEncodeTransactionData implements TransactionController.JSONEncodeTransactionData
func (dtc DefaultTransactionController) JSONEncodeTransactionData(td TransactionData) ([]byte, error) {
	return json.Marshal(td)
}

// JSONDecodeTransactionData implements TransactionController.JSONDecodeTransactionData
func (dtc DefaultTransactionController) JSONDecodeTransactionData(b []byte) (td TransactionData, err error) {
	err = json.Unmarshal(b, &td)
	return
}

// EncodeTransactionData implements TransactionController.EncodeTransactionData
func (ltc LegacyTransactionController) EncodeTransactionData(w io.Writer, td TransactionData) error {
	// turn the transaction data into the legacy format first
	ltd, err := newLegacyTransactionData(td)
	if err != nil {
		return err
	}
	// and encode its result
	return siabin.NewEncoder(w).Encode(ltd)
}

// DecodeTransactionData implements TransactionController.DecodeTransactionData
func (ltc LegacyTransactionController) DecodeTransactionData(r io.Reader) (TransactionData, error) {
	// first decode it as the legacy format
	var ltd legacyTransactionData
	err := siabin.NewDecoder(r).Decode(&ltd)
	if err != nil {
		return TransactionData{}, err
	}
	// output the decoded legacy data as the new output
	return ltd.TransactionData(), nil
}

// JSONEncodeTransactionData implements TransactionController.JSONEncodeTransactionData
func (ltc LegacyTransactionController) JSONEncodeTransactionData(td TransactionData) ([]byte, error) {
	// turn the transaction data into the legacy format first
	ltd, err := newLegacyTransactionData(td)
	if err != nil {
		return nil, err
	}
	// and only than JSON-encode the legacy data
	return json.Marshal(ltd)
}

// JSONDecodeTransactionData implements TransactionController.JSONDecodeTransactionData
func (ltc LegacyTransactionController) JSONDecodeTransactionData(b []byte) (TransactionData, error) {
	// unmarshal the JSON data as the legacy format first
	var ltd legacyTransactionData
	err := json.Unmarshal(b, &ltd)
	if err != nil {
		return TransactionData{}, err
	}
	return ltd.TransactionData(), nil
}

// SignatureHash implements TransactionSignatureHasher.SignatureHash
func (ltc LegacyTransactionController) SignatureHash(t Transaction, extraObjects ...interface{}) (crypto.Hash, error) {
	h := crypto.NewHash()
	enc := siabin.NewEncoder(h)

	if len(extraObjects) > 0 {
		enc.EncodeAll(extraObjects...)
	}
	for _, ci := range t.CoinInputs {
		enc.EncodeAll(ci.ParentID, legacyUnlockHashFromFulfillment(ci.Fulfillment.Fulfillment))
	}
	// legacy transactions encoded unlock hashes in pure form
	enc.Encode(len(t.CoinOutputs))
	for _, co := range t.CoinOutputs {
		enc.EncodeAll(
			co.Value,
			legacyUnlockHashCondition(co.Condition.Condition),
		)
	}
	for _, bsi := range t.BlockStakeInputs {
		enc.EncodeAll(bsi.ParentID, legacyUnlockHashFromFulfillment(bsi.Fulfillment.Fulfillment))
	}
	// legacy transactions encoded unlock hashes in pure form
	enc.Encode(len(t.BlockStakeOutputs))
	for _, bso := range t.BlockStakeOutputs {
		enc.EncodeAll(
			bso.Value,
			legacyUnlockHashCondition(bso.Condition.Condition),
		)
	}
	enc.EncodeAll(
		t.MinerFees,
		t.ArbitraryData,
	)

	var hash crypto.Hash
	h.Sum(hash[:0])
	return hash, nil
}

// EncodeTransactionIDInput implements TransactionIDEncoder.EncodeTransactionIDInput
func (ltc LegacyTransactionController) EncodeTransactionIDInput(w io.Writer, td TransactionData) error {
	return ltc.EncodeTransactionData(w, td)
}

// ensures at compile time that the Transaction Controller implement all desired interfaces
var (
	_ TransactionController = DisabledTransactionController{}
)

// DisabledTransactionController is used for transaction versions that are disabled but still need to be JSON decodable.
type DisabledTransactionController struct {
	DefaultTransactionController
}

// EncodeTransactionData implements TransactionController.EncodeTransactionData
func (dtc DisabledTransactionController) EncodeTransactionData(w io.Writer, td TransactionData) error {
	err := dtc.validateTransactionData(td) // ensure txdata is undefined
	if err != nil {
		return err
	}
	return dtc.DefaultTransactionController.EncodeTransactionData(w, td)
}

// DecodeTransactionData implements TransactionController.DecodeTransactionData
func (dtc DisabledTransactionController) DecodeTransactionData(r io.Reader) (TransactionData, error) {
	td, err := dtc.DefaultTransactionController.DecodeTransactionData(r)
	if err != nil {
		return td, err
	}
	return td, dtc.validateTransactionData(td) // ensure txdata is undefined
}

// JSONEncodeTransactionData implements TransactionController.JSONEncodeTransactionData
func (dtc DisabledTransactionController) JSONEncodeTransactionData(td TransactionData) ([]byte, error) {
	err := dtc.validateTransactionData(td) // ensure txdata is undefined
	if err != nil {
		return nil, err
	}
	return dtc.DefaultTransactionController.JSONEncodeTransactionData(td)
}

// JSONDecodeTransactionData implements TransactionController.JSONDecodeTransactionData
func (dtc DisabledTransactionController) JSONDecodeTransactionData(b []byte) (TransactionData, error) {
	var td TransactionData
	err := json.Unmarshal(b, &td)
	if err != nil {
		return td, err
	}
	return td, dtc.validateTransactionData(td) // ensure txdata is undefined
}

// EncodeTransactionData imple

func (dtc DisabledTransactionController) validateTransaction(t Transaction) error {
	if t.Version != 0 {
		return fmt.Errorf("DisabledTransactionController allows only empty (nil) transactions: invalid %d tx version", t.Version)
	}
	return dtc.validateTransactionData(TransactionData{
		CoinInputs:        t.CoinInputs,
		CoinOutputs:       t.CoinOutputs,
		BlockStakeInputs:  t.BlockStakeInputs,
		BlockStakeOutputs: t.BlockStakeOutputs,
		MinerFees:         t.MinerFees,
		ArbitraryData:     t.ArbitraryData,
		Extension:         t.Extension,
	})
}

func (dtc DisabledTransactionController) validateTransactionData(t TransactionData) error {
	if len(t.CoinInputs) != 0 {
		return errors.New("DisabledTransactionController allows only empty (nil) transactions: coin inputs not allowed")
	}
	if len(t.CoinOutputs) != 0 {
		return errors.New("DisabledTransactionController allows only empty (nil) transactions: coin outputs not allowed")
	}
	if len(t.BlockStakeInputs) != 0 {
		return errors.New("DisabledTransactionController allows only empty (nil) transactions: block stake inputs not allowed")
	}
	if len(t.BlockStakeOutputs) != 0 {
		return errors.New("DisabledTransactionController allows only empty (nil) transactions: block stake outputs not allowed")
	}
	if len(t.MinerFees) != 0 {
		return errors.New("DisabledTransactionController allows only empty (nil) transactions: miner fees not allowed")
	}
	if len(t.ArbitraryData) != 0 {
		return errors.New("DisabledTransactionController allows only empty (nil) transactions: arbitrary data not allowed")
	}
	if t.Extension != nil {
		return errors.New("DisabledTransactionController allows only empty (nil) transactions: extension data not allowed")
	}
	return nil
}

func init() {
	RegisterTransactionVersion(TransactionVersionZero, LegacyTransactionController{})
	RegisterTransactionVersion(TransactionVersionOne, DefaultTransactionController{})
}
