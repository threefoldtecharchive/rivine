package types

import (
	"encoding/json"
	"errors"
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
		ArbitraryData     ArbitraryData      `json:"arbitrarydata,omitempty"`

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
	// can optionally implement, in order to define custom validation logic
	// for a transaction, overwriting the default validation logic.
	TransactionValidator interface {
		ValidateTransaction(t Transaction, ctx ValidationContext, constants TransactionValidationConstants) error
	}

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

	// CoinOutputValidator defines the interface a transaction controller
	// can optionally implement, in order to define custom validation logic
	// for coin outputs, overwriting the default validation logic.
	//
	// The default validation logic ensures that the total amount of output coins (including fees),
	// equals the total amount of input coins. It also ensures that all coin inputs refer with their given ParentID
	// to an existing unspent coin output.
	CoinOutputValidator interface {
		// ValidateCoinOutputs validates if the coin outputs of a given transaction are valid.
		// What the criteria for this validity are, is up to the CoinOutputValidator.
		// All coin inputs of the transaction are already looked up.
		// Should an input not be found, it will be ignored and not be part of the mapped coin inputs.
		ValidateCoinOutputs(t Transaction, ctx FundValidationContext, coinInputs map[CoinOutputID]CoinOutput) error
	}

	// BlockStakeOutputValidator defines the interface a transaction controller
	// can optionally implement, in order to define custom validation logic
	// for block stake outputs, overwriting the default validation logic.
	//
	// The default validation logic ensures that the total amount of output block stakes,
	// equals the total amount of input block stakes. It also ensures that all block stake inputs refer with their given ParentID
	// to an existing unspent block stake output.
	BlockStakeOutputValidator interface {
		// ValidateBlockStakeOutputs validates if the block stake outputs of a given transaction are valid.
		// What the criteria for this validity are, is up to the BlockStakeOutputValidator.
		// All block stake inputs of the transaction are already looked up.
		// Should an input not be found, it will be ignored and not be part of the mapped block stake inputs.
		ValidateBlockStakeOutputs(t Transaction, ctx FundValidationContext, blockStakeInputs map[BlockStakeOutputID]BlockStakeOutput) error
	}

	// TransactionExtensionSigner defines an interface for transactions which have fulfillments in the
	// extension part of the data that have to be signed as well.
	TransactionExtensionSigner interface {
		// SignExtension allows the transaction to sign —using the given sign callback—
		// any fulfillment (giving its condition as reference) that has to be signed.
		SignExtension(extension interface{}, sign func(*UnlockFulfillmentProxy, UnlockConditionProxy, ...interface{}) error) (interface{}, error)
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

// a structure defining the JSON-structure of the TransactionData,
// defined as to preserve compatibility with the existing JSON-structure,
// prior to that the ArbitraryData received support for optional typing.
type jsonTransactionData struct {
	CoinInputs        []CoinInput        `json:"coininputs"` // required
	CoinOutputs       []CoinOutput       `json:"coinoutputs,omitempty"`
	BlockStakeInputs  []BlockStakeInput  `json:"blockstakeinputs,omitempty"`
	BlockStakeOutputs []BlockStakeOutput `json:"blockstakeoutputs,omitempty"`
	MinerFees         []Currency         `json:"minerfees"` // required
	ArbitraryData     []byte             `json:"arbitrarydata,omitempty"`
	ArbitraryDataType ArbitraryDataType  `json:"arbitrarydatatype,omitempty"`
}

// MarshalJSON implements json.Marshaler.MarshalJSON
func (td TransactionData) MarshalJSON() ([]byte, error) {
	return json.Marshal(jsonTransactionData{
		CoinInputs:        td.CoinInputs,
		CoinOutputs:       td.CoinOutputs,
		BlockStakeInputs:  td.BlockStakeInputs,
		BlockStakeOutputs: td.BlockStakeOutputs,
		MinerFees:         td.MinerFees,
		ArbitraryData:     td.ArbitraryData.Data,
		ArbitraryDataType: td.ArbitraryData.Type,
	})
}

// UnmarshalJSON implements json.Marshaler.UnmarshalJSON
func (td *TransactionData) UnmarshalJSON(b []byte) error {
	var jtd jsonTransactionData
	err := json.Unmarshal(b, &jtd)
	if err != nil {
		return err
	}
	td.CoinInputs = jtd.CoinInputs
	td.CoinOutputs = jtd.CoinOutputs
	td.BlockStakeInputs = jtd.BlockStakeInputs
	td.BlockStakeOutputs = jtd.BlockStakeOutputs
	td.MinerFees = jtd.MinerFees
	td.ArbitraryData.Data = jtd.ArbitraryData
	td.ArbitraryData.Type = jtd.ArbitraryDataType
	return nil
}

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
	b := siabin.Marshal(td)
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

func init() {
	RegisterTransactionVersion(TransactionVersionZero, LegacyTransactionController{})
	RegisterTransactionVersion(TransactionVersionOne, DefaultTransactionController{})
}
