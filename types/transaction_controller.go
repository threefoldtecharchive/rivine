package types

import (
	"encoding/json"
	"errors"
	"io"

	"github.com/rivine/rivine/crypto"
	"github.com/rivine/rivine/encoding"
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
	// TransactionValidator defines the interface a transaction controller
	// can optonally implement, in order to define custom validation logic
	// for a transaction, overwriting the default validation logic.
	TransactionValidator interface {
		ValidateTransaction(t Transaction, ctx ValidationContext, constants TransactionValidationConstants) error
	}

	// InputSigHasher defines the interface a transaction controller
	// can optionally implement, in order to define custom Input signatures,
	// overwriting the default input sig hash logic.
	InputSigHasher interface {
		InputSigHash(t Transaction, inputIndex uint64, extraObjects ...interface{}) (crypto.Hash, error)
	}

	// TransactionIDEncoder is an optional interface a transaction controller
	// can implement, in order to use a different binary encoding for ID-generation purposes,
	// instead of using the default binary encoding logic for that transaction (version).
	TransactionIDEncoder interface {
		EncodeTransactionIDInput(io.Writer, TransactionData) error
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
	_RegisteredTransactionVersions = map[TransactionVersion]TransactionController{
		TransactionVersionZero: LegacyTransactionController{},
		TransactionVersionOne:  DefaultTransactionController{},
	}
)

// MarshalSia implements encoding.SiaMarshaller.MarshalSia
func (td TransactionData) MarshalSia(w io.Writer) error {
	return encoding.NewEncoder(w).EncodeAll(
		td.CoinInputs, td.CoinOutputs,
		td.BlockStakeInputs, td.BlockStakeOutputs,
		td.MinerFees, td.ArbitraryData)
}

// UnmarshalSia implements encoding.SiaUnmarshaller.UnmarshalSia
func (td *TransactionData) UnmarshalSia(r io.Reader) error {
	return encoding.NewDecoder(r).DecodeAll(
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
	b := encoding.Marshal(td)
	// copy those bytes together with its prefixed length, as the final encoding
	return encoding.NewEncoder(w).Encode(b)
}

// DecodeTransactionData implements TransactionController.DecodeTransactionData
func (dtc DefaultTransactionController) DecodeTransactionData(r io.Reader) (td TransactionData, err error) {
	// decode it as a byte slice first
	var b []byte
	err = encoding.NewDecoder(r).Decode(&b)
	if err != nil {
		return
	}
	// decode
	err = encoding.Unmarshal(b, &td)
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
	return encoding.NewEncoder(w).Encode(ltd)
}

// DecodeTransactionData implements TransactionController.DecodeTransactionData
func (ltc LegacyTransactionController) DecodeTransactionData(r io.Reader) (TransactionData, error) {
	// first decode it as the legacy format
	var ltd legacyTransactionData
	err := encoding.NewDecoder(r).Decode(&ltd)
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

// InputSigHash implements InputSigHasher.InputSigHash
func (ltc LegacyTransactionController) InputSigHash(t Transaction, inputIndex uint64, extraObjects ...interface{}) (crypto.Hash, error) {
	h := crypto.NewHash()
	enc := encoding.NewEncoder(h)

	enc.Encode(inputIndex)
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
