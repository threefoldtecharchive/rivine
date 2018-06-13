package types

// transaction.go defines the transaction type and all of the sub-fields of the
// transaction, as well as providing helper functions for working with
// transactions. The various IDs are designed such that, in a legal blockchain,
// it is cryptographically unlikely that any two objects would share an id.

import (
	"encoding/json"
	"errors"
	"io"

	"github.com/rivine/rivine/crypto"
	"github.com/rivine/rivine/encoding"
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
	e := encoding.NewEncoder(h)
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
	e := encoding.NewEncoder(h)
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

	return
}

// MarshalSia implements the encoding.SiaMarshaler interface.
func (t Transaction) MarshalSia(w io.Writer) error {
	// get a controller registered or unknown controller
	controller, exists := _RegisteredTransactionVersions[t.Version]
	if !exists {
		return ErrUnknownTransactionType
	}

	// encode the version already
	err := encoding.NewEncoder(w).Encode(t.Version)
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

// UnmarshalSia implements the encoding.SiaUnmarshaler interface.
func (t *Transaction) UnmarshalSia(r io.Reader) error {
	decoder := encoding.NewDecoder(r)
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
		// this will also trigger for v0 transactions
		return DefaultTransactionValidation(t, ctx, constants)
	}
	validator, ok := controller.(TransactionValidator)
	if !ok {
		return DefaultTransactionValidation(t, ctx, constants)
	}
	return validator.ValidateTransaction(t, ctx, constants)
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

var (
	_ encoding.SiaMarshaler   = TransactionVersion(0)
	_ encoding.SiaUnmarshaler = (*TransactionVersion)(nil)
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
	b := encoding.EncUint64(uint64(txsid))
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

	*txsid = TransactionShortID(encoding.DecUint64(b))
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
