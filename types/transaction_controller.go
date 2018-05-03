package types

import (
	"encoding/json"
	"errors"
	"io"

	"github.com/rivine/rivine/build"
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
		EncodeTransactionData(TransactionData) ([]byte, error)
		// DecodeTransactionData binary-decodes the transaction data,
		// which is all transaction properties except for the version.
		DecodeTransactionData([]byte) (TransactionData, error)

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
		Extension interface{} `json:"-"` // omited from JSON
	}
)

// optional interfaces which a TransactionController can implement as well,
// in order to customize a version even more
type (
	// TransactionValidator defines the interface a transaction controller
	// can optonally implement, in order to define custom validation logic
	// for a transaction, overwriting the default validation logic.
	TransactionValidator interface {
		ValidateTransaction(t Transaction, blockSizeLimit uint64) error
	}

	// InputSigHasher defines the interface a transaction controller
	// can optionally implement, in order to define custom Input signatures,
	// overwriting the default input sig hash logic.
	InputSigHasher interface {
		InputSigHash(t Transaction, inputIndex uint64, extraObjects ...interface{}) crypto.Hash
	}

	// TransactionIsStandardChecker defines the interface a transaction controller
	// can optionally implement, in order to define logic,
	// which defines if a transaction is to be concidered standard.
	TransactionIsStandardChecker interface {
		IsStandardTransaction(t Transaction) error
	}
)

// RegisterTransactionVersion registers or unregisters a given transaction version,
// by attaching a controller to it that helps define the behavior surrounding its state.
//
// NOTE: this function should only be called in the `init` func,
// doing it anywhere else can result in undefined behavior.
func RegisterTransactionVersion(v TransactionVersion, c TransactionController) {
	// version 0x00 is off limits,
	// as it's decoding logic is now concidered non-standard,
	// which doesn't work with the current expected format,
	// in which the data slice is encoded as a raw slice,
	// as to support unknown versions properly.
	if v == TransactionVersionZero {
		panic("transaction version 0x00 (legacy) cannot be overriden")
	}
	if c == nil {
		delete(_RegisteredTransactionVersions, v)
		return
	}
	_RegisteredTransactionVersions[v] = c
}

var (
	// ErrUnexpectedExtensionType is an error returned by a transaction controller,
	// in case it expects an extention type it didn't expect.
	ErrUnexpectedExtensionType = errors.New("unexpected transaction data extension type")
)

var (
	_RegisteredTransactionVersions = map[TransactionVersion]TransactionController{
		TransactionVersionOne: DefaultTransactionController{},
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
)

// EncodeTransactionData implements TransactionController.EncodeTransactionData
func (dtc DefaultTransactionController) EncodeTransactionData(td TransactionData) ([]byte, error) {
	if build.DEBUG && td.Extension != nil {
		// for default transactions the extension is always expected to be nil
		return nil, ErrUnexpectedExtensionType
	}
	return encoding.Marshal(td), nil
}

// DecodeTransactionData implements TransactionController.DecodeTransactionData
func (dtc DefaultTransactionController) DecodeTransactionData(b []byte) (td TransactionData, err error) {
	err = encoding.Unmarshal(b, &td)
	return
}

// JSONEncodeTransactionData implements TransactionController.JSONEncodeTransactionData
func (dtc DefaultTransactionController) JSONEncodeTransactionData(td TransactionData) ([]byte, error) {
	if build.DEBUG && td.Extension != nil {
		// for default transactions the extension is always expected to be nil
		return nil, ErrUnexpectedExtensionType
	}
	return json.Marshal(td)
}

// JSONDecodeTransactionData implements TransactionController.JSONDecodeTransactionData
func (dtc DefaultTransactionController) JSONDecodeTransactionData(b []byte) (td TransactionData, err error) {
	err = json.Unmarshal(b, &td)
	return
}
