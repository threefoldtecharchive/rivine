package types

import (
	"encoding/json"
	"errors"

	"github.com/rivine/rivine/build"
)

// RegisterTransactionDecoder registers or unregisters a given decoder,
// linked to a given version.
//
// NOTE: this function should only be called in the `init` func,
// doing it anywhere else can result in undefined behavior.
//
// This function will panic if transaction version 0x00 is tried to be overridden.hehe
func RegisterTransactionDecoder(v TransactionVersion, d TransactionDecoder) {
	// version 0x00 is off limits,
	// as it's decoding logic is now concidered non-standard
	if v == TransactionVersionZero {
		panic("transaction version 0x00 cannot be overriden")
	}
	if d == nil {
		delete(_RegisteredTransactionDecoders, v)
	}
	_RegisteredTransactionDecoders[v] = d
}

var (
	_RegisteredTransactionDecoders = map[TransactionVersion]TransactionDecoder{
	// legacy (0x00) and unknown (0x??) transaction versions are
	// decoded without being registered
	}
)

type (
	// unknownTransactionDecoder is the decoder used for all transactions,
	// which have an unknown version at runtime.
	//
	// This is a special decoder as it is not registered to any specific transaction version,
	// and instead is used by the runtime system for all unknown versions (see: not registered).
	unknownTransactionDecoder struct{}

	// unknownTransactionExtension is the extension
	// used for unknown transactions, as to implement custom encoding and validation logic,
	// specific to transactions which have unknown versions.
	unknownTransactionExtension struct {
		rawData []byte
	}
)

// DecodeTransactionData implements TransactionDecoder.DecodeTransactionData
func (d unknownTransactionDecoder) DecodeTransactionData(version TransactionVersion, b []byte) (t Transaction, err error) {
	t.Version, t.Extension = version, unknownTransactionExtension{rawData: b}
	return
}

// JSONDecodeTransactionData implements TransactionDecoder.JSONDecodeTransactionData
func (d unknownTransactionDecoder) JSONDecodeTransactionData(version TransactionVersion, b []byte) (Transaction, error) {
	return Transaction{}, ErrInvalidTransactionVersion // json-decoding is never allowed for transactions with unknown versions
}

var (
	_ TransactionDecoder = unknownTransactionDecoder{}
)

// EncodeTransaction implements TransactionEncoder.EncodeTransaction
func (ext unknownTransactionExtension) EncodeTransactionData(t Transaction) ([]byte, error) {
	if build.DEBUG {
		err := ext.debugEncodeTransactionCheck(t)
		if err != nil {
			return nil, err
		}
	}
	// all transaction properties get ignored,
	// as we only care about the original raw data.
	return ext.rawData, nil
}

// JSONEncodeTransaction implements TransactionEncoder.JSONEncodeTransaction
func (ext unknownTransactionExtension) JSONEncodeTransactionData(t Transaction) ([]byte, error) {
	if build.DEBUG {
		err := ext.debugEncodeTransactionCheck(t)
		if err != nil {
			return nil, err
		}
	}
	return json.Marshal(ext.rawData) // base64 encoding
}

func (ext unknownTransactionExtension) debugEncodeTransactionCheck(t Transaction) error {
	if len(t.CoinInputs) != 0 {
		return errors.New("no coin inputs should be defined for a transaction with an unknown version number")
	}
	if len(t.CoinOutputs) != 0 {
		return errors.New("no coin outputs should be defined for a transaction with an unknown version number")
	}
	if len(t.BlockStakeInputs) != 0 {
		return errors.New("no block stake inputs should be defined for a transaction with an unknown version number")
	}
	if len(t.BlockStakeOutputs) != 0 {
		return errors.New("no block stake outputs should be defined for a transaction with an unknown version number")
	}
	if len(t.MinerFees) != 0 {
		return errors.New("no miner fees should be defined for a transaction with an unknown version number")
	}
	if len(t.ArbitraryData) != 0 {
		return errors.New("no arbitrary data should be defined for a transaction with an unknown version number")
	}
	return nil
}

// ValidateTransaction implements TransactionValidator.ValidateTransaction
func (ext unknownTransactionExtension) ValidateTransaction(ctx TransactionValidationContext, _ Transaction) error {
	return nil // always valid, as there is no way to really validate the transaction, given its never unmarshalled
}

var (
	_ TransactionDataEncoder = unknownTransactionExtension{}
	_ TransactionValidator   = unknownTransactionExtension{}
)
