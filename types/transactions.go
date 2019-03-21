package types

// transaction.go defines the transaction type and all of the sub-fields of the
// transaction, as well as providing helper functions for working with
// transactions. The various IDs are designed such that, in a legal blockchain,
// it is cryptographically unlikely that any two objects would share an id.

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/threefoldtech/rivine/build"
	"github.com/threefoldtech/rivine/crypto"
	"github.com/threefoldtech/rivine/pkg/encoding/rivbin"
	"github.com/threefoldtech/rivine/pkg/encoding/siabin"
)

const (
	SpecifierLen = 16
)

// These Specifiers are used internally when calculating a type's ID. See
// Specifier for more details.
var (
	SpecifierMinerPayout      = Specifier{'m', 'i', 'n', 'e', 'r', ' ', 'p', 'a', 'y', 'o', 'u', 't'}
	SpecifierCoinInput        = Specifier{'c', 'o', 'i', 'n', ' ', 'i', 'n', 'p', 'u', 't'}
	SpecifierCoinOutput       = Specifier{'c', 'o', 'i', 'n', ' ', 'o', 'u', 't', 'p', 'u', 't'}
	SpecifierBlockStakeInput  = Specifier{'b', 'l', 's', 't', 'a', 'k', 'e', ' ', 'i', 'n', 'p', 'u', 't'}
	SpecifierBlockStakeOutput = Specifier{'b', 'l', 's', 't', 'a', 'k', 'e', ' ', 'o', 'u', 't', 'p', 'u', 't'}
	SpecifierMinerFee         = Specifier{'m', 'i', 'n', 'e', 'r', ' ', 'f', 'e', 'e'}

	ErrInvalidTransactionVersion = errors.New("invalid transaction version")
	ErrTransactionIDWrongLen     = errors.New("input has wrong length to be an encoded transaction id")
)

const (
	// TransactionVersionZero defines the initial (and currently only)
	// version format. Any other version number is considered invalid.
	TransactionVersionZero TransactionVersion = iota
	// TransactionVersionOne defines the new (default) transaction version,
	// which deprecates and is based upon TransactionVersionZero.
	TransactionVersionOne
)

type (
	// A Specifier is a fixed-length byte-array that serves two purposes. In
	// the wire protocol, they are used to identify a particular encoding
	// algorithm, signature algorithm, etc. This allows nodes to communicate on
	// their own terms; for example, to reduce bandwidth costs, a node might
	// only accept compressed messages.
	//
	// Internally, Specifiers are used to guarantee unique IDs. Various
	// consensus types have an associated ID, calculated by hashing the data
	// contained in the type. By prepending the data with Specifier, we can
	// guarantee that distinct types will never produce the same hash.
	Specifier [SpecifierLen]byte

	// TransactionVersion defines the format version of a transaction.
	// However in the future we might wish to support one or multiple new formats,
	// which will be identifable during encoding/decoding by this version number.
	TransactionVersion byte

	// IDs are used to refer to a type without revealing its contents. They
	// are constructed by hashing specific fields of the type, along with a
	// Specifier. While all of these types are hashes, defining type aliases
	// gives us type safety and makes the code more readable.
	TransactionID      crypto.Hash
	CoinOutputID       crypto.Hash
	BlockStakeOutputID crypto.Hash
	OutputID           crypto.Hash

	// A Transaction is an atomic component of a block. Transactions can contain
	// inputs and outputs and even arbitrary
	// data. They can also contain signatures to prove that a given party has
	// approved the transaction, or at least a particular subset of it.
	//
	// Transactions can depend on other previous transactions in the same block,
	// but transactions cannot spend outputs that they create or otherwise be
	// self-dependent.
	Transaction struct {
		// Version of the transaction.
		Version TransactionVersion

		// Core data of a transaction,
		// as expected by the rivine protocol,
		// and will always be available, defined or not.
		CoinInputs        []CoinInput
		CoinOutputs       []CoinOutput
		BlockStakeInputs  []BlockStakeInput
		BlockStakeOutputs []BlockStakeOutput
		MinerFees         []Currency
		ArbitraryData     []byte

		// can adhere any (at once) of {TransactionDataEncoder, TransactionValidator, InputSigHasher},
		// or simply be nil.
		//
		// It is to be used to allow the transactions to take whatever logic and shape
		// as it requires to be, without the rest of the code having to wory about that.
		Extension interface{}
	}

	// A CoinInput consumes a CoinInput and adds the coins to the set of
	// coins that can be spent in the transaction. The ParentID points to the
	// output that is getting consumed, and the UnlockConditions contain the rules
	// for spending the output. The UnlockConditions must match the UnlockHash of
	// the output.
	CoinInput struct {
		ParentID    CoinOutputID           `json:"parentid"`
		Fulfillment UnlockFulfillmentProxy `json:"fulfillment"`
	}

	// A CoinOutput holds a volume of siacoins. Outputs must be spent
	// atomically; that is, they must all be spent in the same transaction. The
	// UnlockHash is the hash of the UnlockConditions that must be fulfilled
	// in order to spend the output.
	CoinOutput struct {
		Value     Currency             `json:"value"`
		Condition UnlockConditionProxy `json:"condition"`
	}

	// A BlockStakeInput consumes a BlockStakeOutput and adds the blockstakes to the set of
	// blockstakes that can be spent in the transaction. The ParentID points to the
	// output that is getting consumed, and the UnlockConditions contain the rules
	// for spending the output. The UnlockConditions must match the UnlockHash of
	// the output.
	BlockStakeInput struct {
		ParentID    BlockStakeOutputID     `json:"parentid"`
		Fulfillment UnlockFulfillmentProxy `json:"fulfillment"`
	}

	// A BlockStakeOutput holds a volume of blockstakes. Outputs must be spent
	// atomically; that is, they must all be spent in the same transaction. The
	// UnlockHash is the hash of a set of UnlockConditions that must be fulfilled
	// in order to spend the output.
	BlockStakeOutput struct {
		Value     Currency             `json:"value"`
		Condition UnlockConditionProxy `json:"condition"`
	}

	// UnspentBlockStakeOutput groups the BlockStakeOutputID, the block height, the transaction index, the output index and the value
	UnspentBlockStakeOutput struct {
		BlockStakeOutputID BlockStakeOutputID
		Indexes            BlockStakeOutputIndexes
		Value              Currency
		Condition          UnlockConditionProxy
	}

	// BlockStakeOutputIndexes groups the block height, the transaction index and the output index to uniquely identify a blockstake output.
	// These indexes and the value are required for the POBS protocol.
	BlockStakeOutputIndexes struct {
		BlockHeight      BlockHeight
		TransactionIndex uint64
		OutputIndex      uint64
	}
)

var (
	// ErrUnknownTransactionType is returned when an unknown transaction version/type was encountered.
	ErrUnknownTransactionType = errors.New("unknown transaction type")
)

// ID returns the id of a transaction, which is taken by marshalling all of the
// fields except for the signatures and taking the hash of the result.
func (t Transaction) ID() (id TransactionID) {
	h := crypto.NewHash()
	t.encodeTransactionDataAsIDInput(h)
	h.Sum(id[:0])
	return
}

// CoinOutputID returns the ID of a coin output at the given index,
// which is calculated by hashing the concatenation of the CoinOutput
// Specifier, all of the fields in the transaction (except the signatures),
// and output index.
func (t Transaction) CoinOutputID(i uint64) (id CoinOutputID) {
	h := crypto.NewHash()
	e := siabin.NewEncoder(h)
	e.Encode(SpecifierCoinOutput)
	t.encodeTransactionDataAsIDInput(h)
	e.Encode(i)
	h.Sum(id[:0])
	return
}

// BlockStakeOutputID returns the ID of a BlockStakeOutput at the given index, which
// is calculated by hashing the concatenation of the BlockStakeOutput Specifier,
// all of the fields in the transaction (except the signatures), and output
// index.
func (t Transaction) BlockStakeOutputID(i uint64) (id BlockStakeOutputID) {
	h := crypto.NewHash()
	e := siabin.NewEncoder(h)
	e.Encode(SpecifierBlockStakeOutput)
	t.encodeTransactionDataAsIDInput(h)
	e.Encode(i)
	h.Sum(id[:0])
	return
}

func (t Transaction) encodeTransactionDataAsIDInput(w io.Writer) error {
	// get a controller registered or unknown controller
	controller, exists := _RegisteredTransactionVersions[t.Version]
	if !exists {
		return ErrUnknownTransactionType
	}
	if transactionIDEncoder, ok := controller.(TransactionIDEncoder); ok {
		td := TransactionData{
			CoinInputs:        t.CoinInputs,
			CoinOutputs:       t.CoinOutputs,
			BlockStakeInputs:  t.BlockStakeInputs,
			BlockStakeOutputs: t.BlockStakeOutputs,
			MinerFees:         t.MinerFees,
			ArbitraryData:     t.ArbitraryData,
			Extension:         t.Extension,
		}
		// use binary encoded specialized for ID Input
		return transactionIDEncoder.EncodeTransactionIDInput(w, td)
	}
	// use the default binary encoding, if the controller does not require specialized
	// encoding logic for ID purposes
	// TODO: (optionally) do this using the RivineEncoder?!
	return t.MarshalSia(w)
}

// CoinOutputSum returns the sum of all the coin outputs in the
// transaction, which must match the sum of all the coin inputs.
func (t Transaction) CoinOutputSum() (sum Currency) {
	// Add the siacoin outputs.
	for _, sco := range t.CoinOutputs {
		sum = sum.Add(sco.Value)
	}

	// Add the miner fees.
	for _, fee := range t.MinerFees {
		sum = sum.Add(fee)
	}

	// add any custom miner payouts
	mps, _ := t.CustomMinerPayouts()
	for _, mp := range mps {
		sum = sum.Add(mp.Value)
	}

	return
}

// CustomMinerPayouts returns any miner payouts originating from this transaction,
// that are not registered as regular MinerFees.
func (t Transaction) CustomMinerPayouts() ([]MinerPayout, error) {
	// get a controller registered or unknown controller
	controller, exists := _RegisteredTransactionVersions[t.Version]
	if !exists {
		return nil, ErrUnknownTransactionType
	}
	if cmpGetter, ok := controller.(TransactionCustomMinerPayoutGetter); ok {
		return cmpGetter.GetCustomMinerPayouts(t.Extension)
	}
	// nothing to do
	return nil, nil
}

// CommonTransactionExtensionData collects the common-understood
// Tx Extension data as a single struct.
type CommonTransactionExtensionData struct {
	UnlockConditions []UnlockConditionProxy
}

// CommonExtensionData returns the common-understood Extension data.
func (t Transaction) CommonExtensionData() (CommonTransactionExtensionData, error) {
	// get a controller registered or unknown controller
	controller, exists := _RegisteredTransactionVersions[t.Version]
	if !exists {
		return CommonTransactionExtensionData{}, ErrUnknownTransactionType
	}
	if cedGetter, ok := controller.(TransactionCommonExtensionDataGetter); ok {
		return cedGetter.GetCommonExtensionData(t.Extension)
	}
	// nothing to do
	return CommonTransactionExtensionData{}, nil
}

// MarshalSia implements the siabin.SiaMarshaler interface.
func (t Transaction) MarshalSia(w io.Writer) error {
	// get a controller registered or unknown controller
	controller, exists := _RegisteredTransactionVersions[t.Version]
	if !exists {
		return ErrUnknownTransactionType
	}

	// encode the version already
	err := siabin.NewEncoder(w).Encode(t.Version)
	if err != nil {
		return err
	}
	// encode the data itself using the controller
	return controller.EncodeTransactionData(w, TransactionData{
		CoinInputs:        t.CoinInputs,
		CoinOutputs:       t.CoinOutputs,
		BlockStakeInputs:  t.BlockStakeInputs,
		BlockStakeOutputs: t.BlockStakeOutputs,
		MinerFees:         t.MinerFees,
		ArbitraryData:     t.ArbitraryData,
		Extension:         t.Extension,
	})
}

// UnmarshalSia implements the siabin.SiaUnmarshaler interface.
func (t *Transaction) UnmarshalSia(r io.Reader) error {
	decoder := siabin.NewDecoder(r)
	err := decoder.Decode(&t.Version)
	if err != nil {
		return err
	}
	// decode the data using the version's controller
	controller, exists := _RegisteredTransactionVersions[t.Version]
	if !exists {
		// a controller is required for each version
		return ErrUnknownTransactionType
	}
	td, err := controller.DecodeTransactionData(r)
	if err != nil {
		return err
	}

	// assign all our data variables directly, zero copy
	t.CoinInputs, t.CoinOutputs = td.CoinInputs, td.CoinOutputs
	t.BlockStakeInputs = td.BlockStakeInputs
	t.BlockStakeOutputs = td.BlockStakeOutputs
	t.MinerFees, t.ArbitraryData = td.MinerFees, td.ArbitraryData
	t.Extension = td.Extension
	return nil
}

// TODO:
// For now the MarshalRivine/UnmarshalRivine methods
// of the Transaction are not used, and even if they are,
// it does not matter as the transaction controllers,
// still define the encoder themselves.
// For now this is OK. Later, when we simplified the
// Rivine codebase, and hopefully got rid of the transaction controller,
// we can probably fix this, by again letting the transaction be in full control of its encoding.

// MarshalRivine implements the rivbin.RivineMarshaler interface.
func (t Transaction) MarshalRivine(w io.Writer) error {
	// get a controller registered or unknown controller
	controller, exists := _RegisteredTransactionVersions[t.Version]
	if !exists {
		return ErrUnknownTransactionType
	}

	// encode the version already
	err := rivbin.NewEncoder(w).Encode(t.Version)
	if err != nil {
		return err
	}
	// encode the data itself using the controller
	return controller.EncodeTransactionData(w, TransactionData{
		CoinInputs:        t.CoinInputs,
		CoinOutputs:       t.CoinOutputs,
		BlockStakeInputs:  t.BlockStakeInputs,
		BlockStakeOutputs: t.BlockStakeOutputs,
		MinerFees:         t.MinerFees,
		ArbitraryData:     t.ArbitraryData,
		Extension:         t.Extension,
	})
}

// UnmarshalRivine implements the rivbin.RivineUnmarshaler interface.
func (t *Transaction) UnmarshalRivine(r io.Reader) error {
	decoder := rivbin.NewDecoder(r)
	err := decoder.Decode(&t.Version)
	if err != nil {
		return err
	}
	// decode the data using the version's controller
	controller, exists := _RegisteredTransactionVersions[t.Version]
	if !exists {
		// a controller is required for each version
		return ErrUnknownTransactionType
	}
	td, err := controller.DecodeTransactionData(r)
	if err != nil {
		return err
	}

	// assign all our data variables directly, zero copy
	t.CoinInputs, t.CoinOutputs = td.CoinInputs, td.CoinOutputs
	t.BlockStakeInputs = td.BlockStakeInputs
	t.BlockStakeOutputs = td.BlockStakeOutputs
	t.MinerFees, t.ArbitraryData = td.MinerFees, td.ArbitraryData
	t.Extension = td.Extension
	return nil
}

// util structs to support some kind of json OneOf feature
// as to make sure our data can support whatever versions we support
type (
	jsonTransaction struct {
		Version TransactionVersion `json:"version"`
		Data    json.RawMessage    `json:"data"`
	}
)

// MarshalJSON implements the json.Marshaler interface.
func (t Transaction) MarshalJSON() ([]byte, error) {
	// get a controller registered or unknown controller
	controller, exists := _RegisteredTransactionVersions[t.Version]
	if !exists {
		return nil, ErrUnknownTransactionType
	}

	// json-encode the data using the controller,
	// to than json-encode the transaction in its totality,
	// using the version and previously encoded data
	data, err := controller.JSONEncodeTransactionData(TransactionData{
		CoinInputs:        t.CoinInputs,
		CoinOutputs:       t.CoinOutputs,
		BlockStakeInputs:  t.BlockStakeInputs,
		BlockStakeOutputs: t.BlockStakeOutputs,
		MinerFees:         t.MinerFees,
		ArbitraryData:     t.ArbitraryData,
		Extension:         t.Extension,
	})
	if err != nil {
		return nil, err
	}
	return json.Marshal(jsonTransaction{
		Version: t.Version,
		Data:    data,
	})
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (t *Transaction) UnmarshalJSON(b []byte) error {
	var txn jsonTransaction
	err := json.Unmarshal(b, &txn)
	if err != nil {
		return err
	}
	controller, exists := _RegisteredTransactionVersions[txn.Version]
	if !exists {
		return ErrUnknownTransactionType
	}
	td, err := controller.JSONDecodeTransactionData(txn.Data)
	if err != nil {
		return err
	}

	// assign all our data variables directly, zero copy
	t.Version = txn.Version
	t.CoinInputs, t.CoinOutputs = td.CoinInputs, td.CoinOutputs
	t.BlockStakeInputs = td.BlockStakeInputs
	t.BlockStakeOutputs = td.BlockStakeOutputs
	t.MinerFees, t.ArbitraryData = td.MinerFees, td.ArbitraryData
	t.Extension = td.Extension
	return nil
}

var (
	_ json.Marshaler   = Transaction{}
	_ json.Unmarshaler = (*Transaction)(nil)
)

// ValidateTransaction validates this transaction in the given context.
//
// By default it checks for a transaction whether the transaction fits within a block,
// that the arbitrary data is within a limited size, there are no outputs
// spent multiple times, all defined values adhere to a network-constant defined
// minimum value (e.g. miner fees having to be at least the MinimumMinerFee),
// and also validates that all used Conditions and Fulfillments are standard.
//
// Each transaction Version however can also choose to overwrite this logic,
// and implement none, some or all of these default rules, optionally
// adding some version-specific rules to it.
func (t Transaction) ValidateTransaction(ctx ValidationContext, constants TransactionValidationConstants) error {
	controller, exists := _RegisteredTransactionVersions[t.Version]
	if !exists {
		return ErrUnknownTransactionType
	}
	validator, ok := controller.(TransactionValidator)
	if !ok {
		return DefaultTransactionValidation(t, ctx, constants)
	}
	return validator.ValidateTransaction(t, ctx, constants)
}

// ValidateCoinOutputs validates the coin outputs within the current blockchain context.
// This behaviour can be overwritten by the used Transaction Controller.
//
// The default validation logic ensures that the total amount of output coins (including fees),
// equals the total amount of input coins. It also ensures that all coin inputs refer with their given ParentID
// to an existing unspent coin output.
func (t Transaction) ValidateCoinOutputs(ctx FundValidationContext, coinInputs map[CoinOutputID]CoinOutput) error {
	controller, exists := _RegisteredTransactionVersions[t.Version]
	if !exists {
		return ErrUnknownTransactionType
	}
	validator, ok := controller.(CoinOutputValidator)
	if !ok {
		return DefaultCoinOutputValidation(t, ctx, coinInputs)
	}
	return validator.ValidateCoinOutputs(t, ctx, coinInputs)
}

// ValidateBlockStakeOutputs validates the blockstake outputs within the current blockchain context.
// This behaviour can be overwritten by the used Transaction Controller.
//
// The default validation logic ensures that the total amount of output blockstakes,
// equals the total amount of input blockstakes. It also ensures that all blockstake inputs refer with their given ParentID
// to an existing unspent blockstake output.
func (t Transaction) ValidateBlockStakeOutputs(ctx FundValidationContext, blockStakeInputs map[BlockStakeOutputID]BlockStakeOutput) error {
	controller, exists := _RegisteredTransactionVersions[t.Version]
	if !exists {
		return ErrUnknownTransactionType
	}
	validator, ok := controller.(BlockStakeOutputValidator)
	if !ok {
		return DefaultBlockStakeOutputValidation(t, ctx, blockStakeInputs)
	}
	return validator.ValidateBlockStakeOutputs(t, ctx, blockStakeInputs)
}

// SignExtension allows the transaction to sign —using the given sign callback—
// any fulfillment defined within the extension data of the transaction that has to be signed.
func (t *Transaction) SignExtension(sign func(*UnlockFulfillmentProxy, UnlockConditionProxy, ...interface{}) error) error {
	controller, exists := _RegisteredTransactionVersions[t.Version]
	if !exists {
		return ErrUnknownTransactionType
	}
	signer, ok := controller.(TransactionExtensionSigner)
	if !ok {
		return nil // nothing to do
	}
	extension, err := signer.SignExtension(t.Extension, sign)
	if err != nil {
		return err
	}
	t.Extension = extension
	return nil
}

// MarshalSia implements SiaMarshaler.MarshalSia
func (v TransactionVersion) MarshalSia(w io.Writer) error {
	_, err := w.Write([]byte{byte(v)})
	return err
}

// UnmarshalSia implements SiaUnmarshaler.UnmarshalSia
func (v *TransactionVersion) UnmarshalSia(r io.Reader) error {
	var bv [1]byte
	_, err := io.ReadFull(r, bv[:])
	*v = TransactionVersion(bv[0])
	return err
}

// MarshalRivine implements RivineMarshaler.MarshalRivine
func (v TransactionVersion) MarshalRivine(w io.Writer) error {
	return rivbin.MarshalUint8(w, uint8(v))
}

// UnmarshalRivine implements RivineUnmarshaler.UnmarshalRivine
func (v *TransactionVersion) UnmarshalRivine(r io.Reader) error {
	x, err := rivbin.UnmarshalUint8(r)
	if err != nil {
		return err
	}
	*v = TransactionVersion(x)
	return err
}

var (
	_ siabin.SiaMarshaler   = TransactionVersion(0)
	_ siabin.SiaUnmarshaler = (*TransactionVersion)(nil)

	_ rivbin.RivineMarshaler   = TransactionVersion(0)
	_ rivbin.RivineUnmarshaler = (*TransactionVersion)(nil)
)

// IsValidTransactionVersion returns an error in case the
// transaction version is not 0, and isn't registered either.
func (v TransactionVersion) IsValidTransactionVersion() error {
	if _, ok := _RegisteredTransactionVersions[v]; ok {
		return nil
	}
	return ErrInvalidTransactionVersion
}

// NewTransactionShortID creates a new Transaction ShortID,
// combining a blockheight together with a transaction index.
// See the TransactionShortID type for more information.
func NewTransactionShortID(height BlockHeight, txSequenceID uint16) TransactionShortID {
	if (height & blockHeightOOBMask) > 0 {
		panic("block height out of bounds")
	}
	if (txSequenceID & txSeqIndexOOBMask) > 0 {
		panic("transaction sequence ID out of bounds")
	}

	return TransactionShortID(height<<txShortIDBlockHeightShift) |
		TransactionShortID(txSequenceID&txSeqIndexMaxMask)
}

// BlockHeight returns the block height part of the transacton short ID.
func (txsid TransactionShortID) BlockHeight() BlockHeight {
	return BlockHeight(txsid >> txShortIDBlockHeightShift)
}

// TransactionSequenceIndex returns the transaction sequence index,
// which is the local (sequence) index of the transaction within a block,
// of the transacton short ID.
func (txsid TransactionShortID) TransactionSequenceIndex() uint16 {
	return uint16(txsid & txSeqIndexMaxMask)
}

// MarshalSia implements SiaMarshaler.SiaMarshaler
func (txsid TransactionShortID) MarshalSia(w io.Writer) error {
	b := siabin.EncUint64(uint64(txsid))
	_, err := w.Write(b)
	return err
}

// UnmarshalSia implements SiaMarshaler.UnmarshalSia
func (txsid *TransactionShortID) UnmarshalSia(r io.Reader) error {
	b := make([]byte, 8)
	_, err := r.Read(b)
	if err != nil {
		return err
	}

	*txsid = TransactionShortID(siabin.DecUint64(b))
	return nil
}

// MarshalRivine implements RivineMarshaler.MarshalRivine
func (txsid TransactionShortID) MarshalRivine(w io.Writer) error {
	return rivbin.MarshalUint64(w, uint64(txsid))
}

// UnmarshalRivine implements RivineUnmarshaler.UnmarshalRivine
func (txsid *TransactionShortID) UnmarshalRivine(r io.Reader) error {
	x, err := rivbin.UnmarshalUint64(r)
	if err != nil {
		return err
	}

	*txsid = TransactionShortID(x)
	return nil
}

// masking and shifting constants used to (de)compose a short transaction ID,
// see the TransactionShortID type for more information.
const (
	// used to protect against a given block height which goes out of
	// the bit range of the available 50 bits, panicing if we're OOB
	blockHeightOOBMask        = 0xFFFC000000000000
	txShortIDBlockHeightShift = 14 // amount of bits reserved for tx index

	txSeqIndexOOBMask = 0xC000
	txSeqIndexMaxMask = 0x3FFF
)

// TransactionShortID is another way to uniquely identify a transaction,
// just as the default hash-based (32-byte) ID uniquely identifies a transaction as well.
// The differences with the default/long ID is that it is 4 times smaller (only 8 bytes),
// and is not just unique, but also ordered. Meaning that byte-wise,
// this short ID informs about its position within the blockchain,
// on such a precise level that you not only to which block it belongs,
// but also its position within that transaction.
//
// The position (indicated by the transaction index),
// is obviously not as important as it is more of a client-side choice,
// rather something agreed upon by consensus.
//
// In memory the transaction is used and manipulated as a uint64,
// where the first 50 bits (going from left to right),
// define the block height, which can have a maximum of about 1.126e+15 (2^50) blocks,
// and the last 14 bits (again going from left to right),
// define the transaction sequence ID, or in other words,
// its unique and shorted position within a given block.
// When serialized into a binary (byte slice) format, is done so using LittleEndian,
// as to correctly preserve the sorted property in all cases.
// Meaning that the ID can be represented in memory and in serialized form as follows:
//
//    [ blockHeight: 50 bits | txSequenceID: 14 bits ]
type TransactionShortID uint64

// MarshalJSON marshals a specifier as a string.
func (s Specifier) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

// String returns the specifier as a string, trimming any trailing zeros.
func (s Specifier) String() string {
	var i int
	for i = range s {
		if s[i] == 0 {
			break
		}
	}
	return string(s[:i])
}

// LoadString loads a stringified specifier into the specifier type
func (s *Specifier) LoadString(str string) error {
	if len(str) > SpecifierLen {
		return errors.New("invalid specifier")
	}
	copy(s[:], str[:])
	return nil
}

// UnmarshalJSON decodes the json string of the specifier.
func (s *Specifier) UnmarshalJSON(b []byte) error {
	var str string
	if err := json.Unmarshal(b, &str); err != nil {
		return err
	}
	return s.LoadString(str)
}

// String prints the id in hex.
func (tid TransactionID) String() string {
	return crypto.Hash(tid).String()
}

// LoadString loads the given transaction ID from a hex string
func (tid *TransactionID) LoadString(str string) error {
	return (*crypto.Hash)(tid).LoadString(str)
}

// MarshalJSON marshals an id as a hex string.
func (tid TransactionID) MarshalJSON() ([]byte, error) {
	return crypto.Hash(tid).MarshalJSON()
}

// UnmarshalJSON decodes the json hex string of the id.
func (tid *TransactionID) UnmarshalJSON(b []byte) error {
	return (*crypto.Hash)(tid).UnmarshalJSON(b)
}

// String prints the output id in hex.
func (oid OutputID) String() string {
	return crypto.Hash(oid).String()
}

// LoadString loads the given output id from a hex string
func (oid *OutputID) LoadString(str string) error {
	return (*crypto.Hash)(oid).LoadString(str)
}

// MarshalJSON marshals an output id as a hex string.
func (oid OutputID) MarshalJSON() ([]byte, error) {
	return crypto.Hash(oid).MarshalJSON()
}

// UnmarshalJSON decodes the json hex string of the output id.
func (oid *OutputID) UnmarshalJSON(b []byte) error {
	return (*crypto.Hash)(oid).UnmarshalJSON(b)
}

// String prints the coin output id in hex.
func (coid CoinOutputID) String() string {
	return crypto.Hash(coid).String()
}

// LoadString loads the given coin output id from a hex string
func (coid *CoinOutputID) LoadString(str string) error {
	return (*crypto.Hash)(coid).LoadString(str)
}

// MarshalJSON marshals an coin output id as a hex string.
func (coid CoinOutputID) MarshalJSON() ([]byte, error) {
	return crypto.Hash(coid).MarshalJSON()
}

// UnmarshalJSON decodes the json hex string of the coin output id.
func (coid *CoinOutputID) UnmarshalJSON(b []byte) error {
	return (*crypto.Hash)(coid).UnmarshalJSON(b)
}

// String prints the blockstake output id in hex.
func (bsoid BlockStakeOutputID) String() string {
	return crypto.Hash(bsoid).String()
}

// LoadString loads the given blockstake output id from a hex string
func (bsoid *BlockStakeOutputID) LoadString(str string) error {
	return (*crypto.Hash)(bsoid).LoadString(str)
}

// MarshalJSON marshals an blockstake output id as a hex string.
func (bsoid BlockStakeOutputID) MarshalJSON() ([]byte, error) {
	return crypto.Hash(bsoid).MarshalJSON()
}

// UnmarshalJSON decodes the json hex string of the blockstake output id.
func (bsoid *BlockStakeOutputID) UnmarshalJSON(b []byte) error {
	return (*crypto.Hash)(bsoid).UnmarshalJSON(b)
}

const (
	// TransactionVersionMinterDefinition defines the Transaction version
	// for a MinterDefinition Transaction.
	//
	// See the `MinterDefinitionTransactionController` and `MinterDefinitionTransaction`
	// types for more information.
	TransactionVersionMinterDefinition TransactionVersion = iota + 128
	// TransactionVersionCoinCreation defines the Transaction version
	// for a CoinCreation Transaction.
	//
	// See the `CoinCreationTransactionController` and `CoinCreationTransaction`
	// types for more information.
	TransactionVersionCoinCreation
)

// These Specifiers are used internally when calculating a Transaction's ID.
// See Rivine's Specifier for more details.
var (
	SpecifierMintDefinitionTransaction = Specifier{'m', 'i', 'n', 't', 'e', 'r', ' ', 'd', 'e', 'f', 'i', 'n', ' ', 't', 'x'}
	SpecifierCoinCreationTransaction   = Specifier{'c', 'o', 'i', 'n', ' ', 'm', 'i', 'n', 't', ' ', 't', 'x'}
)

type (
	// MintConditionGetter allows you to get the mint condition at a given block height.
	//
	// For the daemon this interface could be implemented directly by the DB object
	// that keeps track of the mint condition state, while for a client this could
	// come via the REST API from a rivine daemon in a more indirect way.
	MintConditionGetter interface {
		// GetActiveMintCondition returns the active active mint condition.
		GetActiveMintCondition() (UnlockConditionProxy, error)
		// GetMintConditionAt returns the mint condition at a given block height.
		GetMintConditionAt(height BlockHeight) (UnlockConditionProxy, error)
	}
)

type (
	// CoinCreationTransactionController defines a rivine-specific transaction controller,
	// for a transaction type reserved at type 129. It allows for the creation of Coin Outputs,
	// without requiring coin inputs, but can only be used by the defined Coin Minters.
	CoinCreationTransactionController struct {
		// MintConditionGetter is used to get a mint condition at the context-defined block height.
		//
		// The found MintCondition defines the condition that has to be fulfilled
		// in order to mint new coins into existence (in the form of non-backed coin outputs).
		MintConditionGetter MintConditionGetter
	}

	// MinterDefinitionTransactionController defines a rivine-specific transaction controller,
	// for a transaction type reserved at type 128. It allows the transfer of coin minting powers.
	MinterDefinitionTransactionController struct {
		// MintConditionGetter is used to get a mint condition at the context-defined block height.
		//
		// The found MintCondition defines the condition that has to be fulfilled
		// in order to mint new coins into existence (in the form of non-backed coin outputs).
		MintConditionGetter MintConditionGetter
	}
)

// ensure our controllers implement all desired interfaces
var (
	// ensure at compile time that DefaultTransactionController
	// implements the desired interfaces
	_ TransactionController = DefaultTransactionController{}
	_ TransactionValidator  = DefaultTransactionController{}

	// ensure at compile time that LegacyTransactionController
	// implements the desired interfaces
	_ TransactionController      = LegacyTransactionController{}
	_ TransactionValidator       = LegacyTransactionController{}
	_ TransactionSignatureHasher = LegacyTransactionController{}
	_ TransactionIDEncoder       = LegacyTransactionController{}

	// ensure at compile time that CoinCreationTransactionController
	// implements the desired interfaces
	_ TransactionController      = CoinCreationTransactionController{}
	_ TransactionExtensionSigner = CoinCreationTransactionController{}
	_ TransactionValidator       = CoinCreationTransactionController{}
	_ CoinOutputValidator        = CoinCreationTransactionController{}
	_ BlockStakeOutputValidator  = CoinCreationTransactionController{}
	_ TransactionSignatureHasher = CoinCreationTransactionController{}
	_ TransactionIDEncoder       = CoinCreationTransactionController{}

	// ensure at compile time that MinterDefinitionTransactionController
	// implements the desired interfaces
	_ TransactionController                = MinterDefinitionTransactionController{}
	_ TransactionExtensionSigner           = MinterDefinitionTransactionController{}
	_ TransactionValidator                 = MinterDefinitionTransactionController{}
	_ CoinOutputValidator                  = MinterDefinitionTransactionController{}
	_ BlockStakeOutputValidator            = MinterDefinitionTransactionController{}
	_ TransactionSignatureHasher           = MinterDefinitionTransactionController{}
	_ TransactionIDEncoder                 = MinterDefinitionTransactionController{}
	_ TransactionCommonExtensionDataGetter = MinterDefinitionTransactionController{}
)

// DefaultTransactionController

// ValidateTransaction implements TransactionValidator.ValidateTransaction
func (dtc DefaultTransactionController) ValidateTransaction(t Transaction, ctx ValidationContext, constants TransactionValidationConstants) error {
	if ctx.Confirmed && ctx.BlockHeight < dtc.TransactionFeeCheckBlockHeight {
		// as to ensure the miner fee is at least bigger than 0,
		// we however only want to put this restriction within the consensus set,
		// the stricter miner fee checks should apply immediately to the transaction pool logic
		constants.MinimumMinerFee = NewCurrency64(1)
	}
	return DefaultTransactionValidation(t, ctx, constants)
}

// LegacyTransactionController

// ValidateTransaction implements TransactionValidator.ValidateTransaction
func (ltc LegacyTransactionController) ValidateTransaction(t Transaction, ctx ValidationContext, constants TransactionValidationConstants) error {
	if ctx.Confirmed && ctx.BlockHeight < ltc.TransactionFeeCheckBlockHeight {
		// as to ensure the miner fee is at least bigger than 0,
		// we however only want to put this restriction within the consensus set,
		// the stricter miner fee checks should apply immediately to the transaction pool logic
		constants.MinimumMinerFee = NewCurrency64(1)
	}
	return DefaultTransactionValidation(t, ctx, constants)
}

// CoinCreationTransactionController

// EncodeTransactionData implements TransactionController.EncodeTransactionData
func (cctc CoinCreationTransactionController) EncodeTransactionData(w io.Writer, txData TransactionData) error {
	cctx, err := CoinCreationTransactionFromTransactionData(txData)
	if err != nil {
		return fmt.Errorf("failed to convert txData to a CoinCreationTx: %v", err)
	}
	return siabin.NewEncoder(w).Encode(cctx)
}

// DecodeTransactionData implements TransactionController.DecodeTransactionData
func (cctc CoinCreationTransactionController) DecodeTransactionData(r io.Reader) (TransactionData, error) {
	var cctx CoinCreationTransaction
	err := siabin.NewDecoder(r).Decode(&cctx)
	if err != nil {
		return TransactionData{}, fmt.Errorf(
			"failed to binary-decode tx as a CoinCreationTx: %v", err)
	}
	// return coin creation tx as regular rivine tx data
	return cctx.TransactionData(), nil
}

// JSONEncodeTransactionData implements TransactionController.JSONEncodeTransactionData
func (cctc CoinCreationTransactionController) JSONEncodeTransactionData(txData TransactionData) ([]byte, error) {
	cctx, err := CoinCreationTransactionFromTransactionData(txData)
	if err != nil {
		return nil, fmt.Errorf("failed to convert txData to a CoinCreationTx: %v", err)
	}
	return json.Marshal(cctx)
}

// JSONDecodeTransactionData implements TransactionController.JSONDecodeTransactionData
func (cctc CoinCreationTransactionController) JSONDecodeTransactionData(data []byte) (TransactionData, error) {
	var cctx CoinCreationTransaction
	err := json.Unmarshal(data, &cctx)
	if err != nil {
		return TransactionData{}, fmt.Errorf(
			"failed to json-decode tx as a CoinCreationTx: %v", err)
	}
	// return coin creation tx as regular rivine tx data
	return cctx.TransactionData(), nil
}

// SignExtension implements TransactionExtensionSigner.SignExtension
func (cctc CoinCreationTransactionController) SignExtension(extension interface{}, sign func(*UnlockFulfillmentProxy, UnlockConditionProxy, ...interface{}) error) (interface{}, error) {
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

// ValidateTransaction implements TransactionValidator.ValidateTransaction
func (cctc CoinCreationTransactionController) ValidateTransaction(t Transaction, ctx ValidationContext, constants TransactionValidationConstants) (err error) {
	err = TransactionFitsInABlock(t, constants.BlockSizeLimit)
	if err != nil {
		return err
	}

	// get CoinCreationTxn
	cctx, err := CoinCreationTransactionFromTransaction(t)
	if err != nil {
		return fmt.Errorf("failed to use tx as a coin creation tx: %v", err)
	}

	// get MintCondition
	mintCondition, err := cctc.MintConditionGetter.GetMintConditionAt(ctx.BlockHeight)
	if err != nil {
		return fmt.Errorf("failed to get mint condition at block height %d: %v", ctx.BlockHeight, err)
	}

	// check if MintFulfillment fulfills the Globally defined MintCondition for the context-defined block height
	err = mintCondition.Fulfill(cctx.MintFulfillment, FulfillContext{
		BlockHeight: ctx.BlockHeight,
		BlockTime:   ctx.BlockTime,
		Transaction: t,
	})
	if err != nil {
		return fmt.Errorf("failed to fulfill mint condition: %v", err)
	}
	// ensure the Nonce is not Nil
	if cctx.Nonce == (TransactionNonce{}) {
		return errors.New("nil nonce is not allowed for a coin creation transaction")
	}

	// validate the rest of the content
	err = ArbitraryDataFits(cctx.ArbitraryData, constants.ArbitraryDataSizeLimit)
	if err != nil {
		return
	}
	for _, fee := range cctx.MinerFees {
		if fee.Cmp(constants.MinimumMinerFee) == -1 {
			return ErrTooSmallMinerFee
		}
	}
	// check if all condtions are standard and that the parent outputs have non-zero values
	for _, sco := range cctx.CoinOutputs {
		if sco.Value.IsZero() {
			return ErrZeroOutput
		}
		err = sco.Condition.IsStandardCondition(ctx)
		if err != nil {
			return err
		}
	}
	return
}

// ValidateCoinOutputs implements CoinOutputValidator.ValidateCoinOutputs
func (cctc CoinCreationTransactionController) ValidateCoinOutputs(t Transaction, ctx FundValidationContext, coinInputs map[CoinOutputID]CoinOutput) (err error) {
	return nil // always valid, coin outputs are created not backed
}

// ValidateBlockStakeOutputs implements BlockStakeOutputValidator.ValidateBlockStakeOutputs
func (cctc CoinCreationTransactionController) ValidateBlockStakeOutputs(t Transaction, ctx FundValidationContext, blockStakeInputs map[BlockStakeOutputID]BlockStakeOutput) (err error) {
	return nil // always valid, no block stake inputs/outputs exist within a coin creation transaction
}

// SignatureHash implements TransactionSignatureHasher.SignatureHash
func (cctc CoinCreationTransactionController) SignatureHash(t Transaction, extraObjects ...interface{}) (crypto.Hash, error) {
	cctx, err := CoinCreationTransactionFromTransaction(t)
	if err != nil {
		return crypto.Hash{}, fmt.Errorf("failed to use tx as a coin creation tx: %v", err)
	}

	h := crypto.NewHash()
	enc := siabin.NewEncoder(h)

	enc.EncodeAll(
		t.Version,
		SpecifierCoinCreationTransaction,
		cctx.Nonce,
	)

	if len(extraObjects) > 0 {
		enc.EncodeAll(extraObjects...)
	}

	enc.EncodeAll(
		cctx.CoinOutputs,
		cctx.MinerFees,
		cctx.ArbitraryData,
	)

	var hash crypto.Hash
	h.Sum(hash[:0])
	return hash, nil
}

// EncodeTransactionIDInput implements TransactionIDEncoder.EncodeTransactionIDInput
func (cctc CoinCreationTransactionController) EncodeTransactionIDInput(w io.Writer, txData TransactionData) error {
	cctx, err := CoinCreationTransactionFromTransactionData(txData)
	if err != nil {
		return fmt.Errorf("failed to convert txData to a CoinCreationTx: %v", err)
	}
	return siabin.NewEncoder(w).EncodeAll(SpecifierCoinCreationTransaction, cctx)
}

// MinterDefinitionTransactionController

// EncodeTransactionData implements TransactionController.EncodeTransactionData
func (mdtc MinterDefinitionTransactionController) EncodeTransactionData(w io.Writer, txData TransactionData) error {
	mdtx, err := MinterDefinitionTransactionFromTransactionData(txData)
	if err != nil {
		return fmt.Errorf("failed to convert txData to a MinterDefinitionTx: %v", err)
	}
	return siabin.NewEncoder(w).Encode(mdtx)
}

// DecodeTransactionData implements TransactionController.DecodeTransactionData
func (mdtc MinterDefinitionTransactionController) DecodeTransactionData(r io.Reader) (TransactionData, error) {
	var mdtx MinterDefinitionTransaction
	err := siabin.NewDecoder(r).Decode(&mdtx)
	if err != nil {
		return TransactionData{}, fmt.Errorf(
			"failed to binary-decode tx as a MinterDefinitionTx: %v", err)
	}
	// return minter definition tx as regular rivine tx data
	return mdtx.TransactionData(), nil
}

// JSONEncodeTransactionData implements TransactionController.JSONEncodeTransactionData
func (mdtc MinterDefinitionTransactionController) JSONEncodeTransactionData(txData TransactionData) ([]byte, error) {
	mdtx, err := MinterDefinitionTransactionFromTransactionData(txData)
	if err != nil {
		return nil, fmt.Errorf("failed to convert txData to a MinterDefinitionTx: %v", err)
	}
	return json.Marshal(mdtx)
}

// JSONDecodeTransactionData implements TransactionController.JSONDecodeTransactionData
func (mdtc MinterDefinitionTransactionController) JSONDecodeTransactionData(data []byte) (TransactionData, error) {
	var mdtx MinterDefinitionTransaction
	err := json.Unmarshal(data, &mdtx)
	if err != nil {
		return TransactionData{}, fmt.Errorf(
			"failed to json-decode tx as a MinterDefinitionTx: %v", err)
	}
	// return minter definition tx as regular rivine tx data
	return mdtx.TransactionData(), nil
}

// SignExtension implements TransactionExtensionSigner.SignExtension
func (mdtc MinterDefinitionTransactionController) SignExtension(extension interface{}, sign func(*UnlockFulfillmentProxy, UnlockConditionProxy, ...interface{}) error) (interface{}, error) {
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

// ValidateTransaction implements TransactionValidator.ValidateTransaction
func (mdtc MinterDefinitionTransactionController) ValidateTransaction(t Transaction, ctx ValidationContext, constants TransactionValidationConstants) (err error) {
	err = TransactionFitsInABlock(t, constants.BlockSizeLimit)
	if err != nil {
		return err
	}

	// get MinterDefinitionTx
	mdtx, err := MinterDefinitionTransactionFromTransaction(t)
	if err != nil {
		return fmt.Errorf("failed to use tx as a coin creation tx: %v", err)
	}

	// check if the MintCondition is valid
	err = mdtx.MintCondition.IsStandardCondition(ctx)
	if err != nil {
		return fmt.Errorf("defined mint condition is not standard within the given blockchain context: %v", err)
	}
	// check if the valid mint condition has a type we want to support, one of:
	//   * PubKey-UnlockHashCondtion
	//   * MultiSigConditions
	//   * TimeLockConditions (if the internal condition type is supported)
	err = validateMintCondition(mdtx.MintCondition)
	if err != nil {
		return err
	}

	// get MintCondition
	mintCondition, err := mdtc.MintConditionGetter.GetMintConditionAt(ctx.BlockHeight)
	if err != nil {
		return fmt.Errorf("failed to get mint condition at block height %d: %v", ctx.BlockHeight, err)
	}

	// check if MintFulfillment fulfills the Globally defined MintCondition for the context-defined block height
	err = mintCondition.Fulfill(mdtx.MintFulfillment, FulfillContext{
		BlockHeight: ctx.BlockHeight,
		BlockTime:   ctx.BlockTime,
		Transaction: t,
	})
	if err != nil {
		return fmt.Errorf("failed to fulfill mint condition: %v", err)
	}
	// ensure the Nonce is not Nil
	if mdtx.Nonce == (TransactionNonce{}) {
		return errors.New("nil nonce is not allowed for a mint condition transaction")
	}

	// validate the rest of the content
	err = ArbitraryDataFits(mdtx.ArbitraryData, constants.ArbitraryDataSizeLimit)
	if err != nil {
		return
	}
	for _, fee := range mdtx.MinerFees {
		if fee.Cmp(constants.MinimumMinerFee) == -1 {
			return ErrTooSmallMinerFee
		}
	}
	return
}

func validateMintCondition(condition UnlockCondition) error {
	switch ct := condition.ConditionType(); ct {
	case ConditionTypeMultiSignature:
		// always valid
		return nil

	case ConditionTypeUnlockHash:
		// only valid for unlock hash type 1 (PubKey)
		if condition.UnlockHash().Type == UnlockTypePubKey {
			return nil
		}
		return errors.New("unlockHash conditions can be used as mint conditions, if the unlock hash type is PubKey")

	case ConditionTypeTimeLock:
		// ensure to unpack a proxy condition first
		if cp, ok := condition.(UnlockConditionProxy); ok {
			condition = cp.Condition
		}
		// time lock conditions are allowed as long as the internal condition is allowed
		cg, ok := condition.(MarshalableUnlockConditionGetter)
		if !ok {
			err := fmt.Errorf("unexpected Go-type for TimeLockCondition: %T", condition)
			if build.DEBUG {
				panic(err)
			}
			return err
		}
		return validateMintCondition(cg.GetMarshalableUnlockCondition())

	default:
		// all other types aren't allowed
		return fmt.Errorf("condition type %d cannot be used as a mint condition", ct)
	}
}

// ValidateCoinOutputs implements CoinOutputValidator.ValidateCoinOutputs
func (mdtc MinterDefinitionTransactionController) ValidateCoinOutputs(t Transaction, ctx FundValidationContext, coinInputs map[CoinOutputID]CoinOutput) (err error) {
	return nil // always valid, no block stake inputs/outputs exist within a minter definition transaction
}

// ValidateBlockStakeOutputs implements BlockStakeOutputValidator.ValidateBlockStakeOutputs
func (mdtc MinterDefinitionTransactionController) ValidateBlockStakeOutputs(t Transaction, ctx FundValidationContext, blockStakeInputs map[BlockStakeOutputID]BlockStakeOutput) (err error) {
	return nil // always valid, no block stake inputs/outputs exist within a minter definition transaction
}

// SignatureHash implements TransactionSignatureHasher.SignatureHash
func (mdtc MinterDefinitionTransactionController) SignatureHash(t Transaction, extraObjects ...interface{}) (crypto.Hash, error) {
	mdtx, err := MinterDefinitionTransactionFromTransaction(t)
	if err != nil {
		return crypto.Hash{}, fmt.Errorf("failed to use tx as a MinterDefinitionTx: %v", err)
	}

	h := crypto.NewHash()
	enc := siabin.NewEncoder(h)

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
func (mdtc MinterDefinitionTransactionController) EncodeTransactionIDInput(w io.Writer, txData TransactionData) error {
	mdtx, err := MinterDefinitionTransactionFromTransactionData(txData)
	if err != nil {
		return fmt.Errorf("failed to convert txData to a MinterDefinitionTx: %v", err)
	}
	return siabin.NewEncoder(w).EncodeAll(SpecifierMintDefinitionTransaction, mdtx)
}

// GetCommonExtensionData implements TransactionCommonExtensionDataGetter.GetCommonExtensionData
func (mdtc MinterDefinitionTransactionController) GetCommonExtensionData(extension interface{}) (CommonTransactionExtensionData, error) {
	mdext, ok := extension.(*MinterDefinitionTransactionExtension)
	if !ok {
		return CommonTransactionExtensionData{}, errors.New("invalid extension data for a MinterDefinitionTx")
	}
	return CommonTransactionExtensionData{
		UnlockConditions: []UnlockConditionProxy{mdext.MintCondition},
	}, nil
}

// TransactionNonce is a nonce
// used to ensure the uniqueness of an otherwise potentially non-unique Tx
type TransactionNonce [TransactionNonceLength]byte

// TransactionNonceLength defines the length of a TransactionNonce
const TransactionNonceLength = 8

type (
	// CoinCreationTransaction is to be created only by the defined Coin Minters,
	// as a medium in order to create coins (coin outputs), without backing them
	// (so without having to spend previously unspend coin outputs, see: coin inputs).
	CoinCreationTransaction struct {
		// Nonce used to ensure the uniqueness of a CoinCreationTransaction's ID and signature.
		Nonce TransactionNonce `json:"nonce"`
		// MintFulfillment defines the fulfillment which is used in order to
		// fulfill the globally defined MintCondition.
		MintFulfillment UnlockFulfillmentProxy `json:"mintfulfillment"`
		// CoinOutputs defines the coin outputs,
		// which contain the freshly created coins, adding to the total pool of coins
		// available in the rivine network.
		CoinOutputs []CoinOutput `json:"coinoutputs"`
		// Minerfees, a fee paid for this coin creation transaction.
		MinerFees []Currency `json:"minerfees"`
		// ArbitraryData can be used for any purpose,
		// but is mostly to be used in order to define the reason/origins
		// of the coin creation.
		ArbitraryData []byte `json:"arbitrarydata,omitempty"`
	}
	// CoinCreationTransactionExtension defines the CoinCreationTx Extension Data
	CoinCreationTransactionExtension struct {
		Nonce           TransactionNonce
		MintFulfillment UnlockFulfillmentProxy
	}
)

// CoinCreationTransactionFromTransaction creates a CoinCreationTransaction,
// using a regular in-memory rivine transaction.
//
// Past the (tx) Version validation it piggy-backs onto the
// `CoinCreationTransactionFromTransactionData` constructor.
func CoinCreationTransactionFromTransaction(tx Transaction) (CoinCreationTransaction, error) {
	if tx.Version != TransactionVersionCoinCreation {
		return CoinCreationTransaction{}, fmt.Errorf(
			"a coin creation transaction requires tx version %d",
			TransactionVersionCoinCreation)
	}
	return CoinCreationTransactionFromTransactionData(TransactionData{
		CoinInputs:        tx.CoinInputs,
		CoinOutputs:       tx.CoinOutputs,
		BlockStakeInputs:  tx.BlockStakeInputs,
		BlockStakeOutputs: tx.BlockStakeOutputs,
		MinerFees:         tx.MinerFees,
		ArbitraryData:     tx.ArbitraryData,
		Extension:         tx.Extension,
	})
}

// CoinCreationTransactionFromTransactionData creates a CoinCreationTransaction,
// using the TransactionData from a regular in-memory rivine transaction.
func CoinCreationTransactionFromTransactionData(txData TransactionData) (CoinCreationTransaction, error) {
	// (tx) extension (data) is expected to be a pointer to a valid CoinCreationTransactionExtension,
	// which contains the nonce and the mintFulfillment that can be used to fulfill the globally defined mint condition
	extensionData, ok := txData.Extension.(*CoinCreationTransactionExtension)
	if !ok {
		return CoinCreationTransaction{}, errors.New("invalid extension data for a CoinCreationTransaction")
	}
	// at least one coin output as well as one miner fee is required
	if len(txData.CoinOutputs) == 0 || len(txData.MinerFees) == 0 {
		return CoinCreationTransaction{}, errors.New("at least one coin output and miner fee is required for a CoinCreationTransaction")
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
func (cctx *CoinCreationTransaction) TransactionData() TransactionData {
	return TransactionData{
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
func (cctx *CoinCreationTransaction) Transaction() Transaction {
	return Transaction{
		Version:       TransactionVersionCoinCreation,
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
	// MinterDefinitionTransaction is to be created only by the defined Coin Minters,
	// as a medium in order to transfer minting powers.
	MinterDefinitionTransaction struct {
		// Nonce used to ensure the uniqueness of a MinterDefinitionTransaction's ID and signature.
		Nonce TransactionNonce `json:"nonce"`
		// MintFulfillment defines the fulfillment which is used in order to
		// fulfill the globally defined MintCondition.
		MintFulfillment UnlockFulfillmentProxy `json:"mintfulfillment"`
		// MintCondition defines a new condition that defines who become(s) the new minter(s),
		// and thus defines who can create coins as well as update who is/are the current minter(s)
		//
		// UnlockHash (unlockhash type 1) and MultiSigConditions are allowed,
		// as well as TimeLocked conditions which have UnlockHash- and MultiSigConditions as
		// internal condition.
		MintCondition UnlockConditionProxy `json:"mintcondition"`
		// Minerfees, a fee paid for this minter definition transaction.
		MinerFees []Currency `json:"minerfees"`
		// ArbitraryData can be used for any purpose,
		// but is mostly to be used in order to define the reason/origins
		// of the transfer of minting power.
		ArbitraryData []byte `json:"arbitrarydata,omitempty"`
	}
	// MinterDefinitionTransactionExtension defines the MinterDefinitionTx Extension Data
	MinterDefinitionTransactionExtension struct {
		Nonce           TransactionNonce
		MintFulfillment UnlockFulfillmentProxy
		MintCondition   UnlockConditionProxy
	}
)

// MinterDefinitionTransactionFromTransaction creates a MinterDefinitionTransaction,
// using a regular in-memory rivine transaction.
//
// Past the (tx) Version validation it piggy-backs onto the
// `MinterDefinitionTransactionFromTransactionData` constructor.
func MinterDefinitionTransactionFromTransaction(tx Transaction) (MinterDefinitionTransaction, error) {
	if tx.Version != TransactionVersionMinterDefinition {
		return MinterDefinitionTransaction{}, fmt.Errorf(
			"a minter definition transaction requires tx version %d",
			TransactionVersionCoinCreation)
	}
	return MinterDefinitionTransactionFromTransactionData(TransactionData{
		CoinInputs:        tx.CoinInputs,
		CoinOutputs:       tx.CoinOutputs,
		BlockStakeInputs:  tx.BlockStakeInputs,
		BlockStakeOutputs: tx.BlockStakeOutputs,
		MinerFees:         tx.MinerFees,
		ArbitraryData:     tx.ArbitraryData,
		Extension:         tx.Extension,
	})
}

// MinterDefinitionTransactionFromTransactionData creates a MinterDefinitionTransaction,
// using the TransactionData from a regular in-memory rivine transaction.
func MinterDefinitionTransactionFromTransactionData(txData TransactionData) (MinterDefinitionTransaction, error) {
	// (tx) extension (data) is expected to be a pointer to a valid MinterDefinitionTransactionExtension,
	// which contains the nonce, the mintFulfillment that can be used to fulfill the currently globally defined mint condition,
	// as well as a mintCondition to replace the current in-place mintCondition.
	extensionData, ok := txData.Extension.(*MinterDefinitionTransactionExtension)
	if !ok {
		return MinterDefinitionTransaction{}, errors.New("invalid extension data for a MinterDefinitionTransaction")
	}
	// at least one miner fee is required
	if len(txData.MinerFees) == 0 {
		return MinterDefinitionTransaction{}, errors.New("at least one miner fee is required for a MinterDefinitionTransaction")
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
func (cctx *MinterDefinitionTransaction) TransactionData() TransactionData {
	return TransactionData{
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
func (cctx *MinterDefinitionTransaction) Transaction() Transaction {
	return Transaction{
		Version:       TransactionVersionMinterDefinition,
		MinerFees:     cctx.MinerFees,
		ArbitraryData: cctx.ArbitraryData,
		Extension: &MinterDefinitionTransactionExtension{
			Nonce:           cctx.Nonce,
			MintFulfillment: cctx.MintFulfillment,
			MintCondition:   cctx.MintCondition,
		},
	}
}
