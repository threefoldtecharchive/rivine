package types

// transaction.go defines the transaction type and all of the sub-fields of the
// transaction, as well as providing helper functions for working with
// transactions. The various IDs are designed such that, in a legal blockchain,
// it is cryptographically unlikely that any two objects would share an id.

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/rivine/rivine/crypto"
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

	ErrTransactionIDWrongLen = errors.New("input has wrong length to be an encoded transaction id")
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
		CoinInputs            []CoinInput            `json:"coininputs"`
		CoinOutputs           []CoinOutput           `json:"coinoutputs"`
		BlockStakeInputs      []BlockStakeInput      `json:"blockstakeinputs"`
		BlockStakeOutputs     []BlockStakeOutput     `json:"blockstakeoutputs"`
		MinerFees             []Currency             `json:"minerfees"`
		ArbitraryData         [][]byte               `json:"arbitrarydata"`
		TransactionSignatures []TransactionSignature `json:"transactionsignatures"`
	}

	// A SiacoinInput consumes a SiacoinOutput and adds the siacoins to the set of
	// siacoins that can be spent in the transaction. The ParentID points to the
	// output that is getting consumed, and the UnlockConditions contain the rules
	// for spending the output. The UnlockConditions must match the UnlockHash of
	// the output.
	CoinInput struct {
		ParentID         CoinOutputID     `json:"parentid"`
		UnlockConditions UnlockConditions `json:"unlockconditions"`
	}

	// A CoinOutput holds a volume of siacoins. Outputs must be spent
	// atomically; that is, they must all be spent in the same transaction. The
	// UnlockHash is the hash of the UnlockConditions that must be fulfilled
	// in order to spend the output.
	CoinOutput struct {
		Value      Currency   `json:"value"`
		UnlockHash UnlockHash `json:"unlockhash"`
	}

	// A BlockStakeInput consumes a BlockStakeOutput and adds the blockstakes to the set of
	// blockstakes that can be spent in the transaction. The ParentID points to the
	// output that is getting consumed, and the UnlockConditions contain the rules
	// for spending the output. The UnlockConditions must match the UnlockHash of
	// the output.
	BlockStakeInput struct {
		ParentID         BlockStakeOutputID `json:"parentid"`
		UnlockConditions UnlockConditions   `json:"unlockconditions"`
	}

	// A BlockStakeOutput holds a volume of blockstakes. Outputs must be spent
	// atomically; that is, they must all be spent in the same transaction. The
	// UnlockHash is the hash of a set of UnlockConditions that must be fulfilled
	// in order to spend the output.
	BlockStakeOutput struct {
		Value      Currency   `json:"value"`
		UnlockHash UnlockHash `json:"unlockhash"`
	}

	// UnspentBlockStakeOutput groups the BlockStakeOutputID, the block height, the transaction index, the output index and the value
	UnspentBlockStakeOutput struct {
		BlockStakeOutputID BlockStakeOutputID
		Indexes            BlockStakeOutputIndexes
		Value              Currency
		UnlockHash         UnlockHash
	}

	// BlockStakeOutputIndexes groups the block height, the transaction index and the output index to uniquely identify a blockstake output.
	// These indexes and the value are required for the POBS protocol.
	BlockStakeOutputIndexes struct {
		BlockHeight      BlockHeight
		TransactionIndex uint64
		OutputIndex      uint64
	}
)

// ID returns the id of a transaction, which is taken by marshalling all of the
// fields except for the signatures and taking the hash of the result.
func (t Transaction) ID() TransactionID {
	return TransactionID(crypto.HashAll(
		t.CoinInputs,
		t.CoinOutputs,
		t.BlockStakeInputs,
		t.BlockStakeOutputs,
		t.MinerFees,
		t.ArbitraryData,
	))
}

// CoinOutputID returns the ID of a coin output at the given index,
// which is calculated by hashing the concatenation of the CoinOutput
// Specifier, all of the fields in the transaction (except the signatures),
// and output index.
func (t Transaction) CoinOutputID(i uint64) CoinOutputID {
	return CoinOutputID(crypto.HashAll(
		SpecifierCoinOutput,
		t.CoinInputs,
		t.CoinOutputs,
		t.BlockStakeInputs,
		t.BlockStakeOutputs,
		t.MinerFees,
		t.ArbitraryData,
		i,
	))
}

// BlockStakeOutputID returns the ID of a BlockStakeOutput at the given index, which
// is calculated by hashing the concatenation of the BlockStakeOutput Specifier,
// all of the fields in the transaction (except the signatures), and output
// index.
func (t Transaction) BlockStakeOutputID(i uint64) BlockStakeOutputID {
	return BlockStakeOutputID(crypto.HashAll(
		SpecifierBlockStakeOutput,
		t.CoinInputs,
		t.CoinOutputs,
		t.BlockStakeInputs,
		t.BlockStakeOutputs,
		t.MinerFees,
		t.ArbitraryData,
		i,
	))
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

// Below this point is a bunch of repeated definitions so that all type aliases
// of 'crypto.Hash' are printed as hex strings. A notable exception is
// types.Target, which is still printed as a byte array.

// MarshalJSON marshales a specifier as a hex string.
func (s Specifier) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

// String prints the specifier in hex.
func (s Specifier) String() string {
	var i int
	for i = 0; i < len(s); i++ {
		if s[i] == 0 {
			break
		}
	}
	return string(s[:i])
}

// UnmarshalJSON decodes the json hex string of the id.
func (s *Specifier) UnmarshalJSON(b []byte) error {
	// Copy b into s, minus the json quotation marks.
	copy(s[:], b[1:len(b)-1])
	return nil
}

// MarshalJSON marshales an id as a hex string.
func (tid TransactionID) MarshalJSON() ([]byte, error) {
	return json.Marshal(tid.String())
}

// String prints the id in hex.
func (tid TransactionID) String() string {
	return fmt.Sprintf("%x", tid[:])
}

// UnmarshalJSON decodes the json hex string of the id.
func (tid *TransactionID) UnmarshalJSON(b []byte) error {
	if len(b) != crypto.HashSize*2+2 {
		return crypto.ErrHashWrongLen
	}

	var tidBytes []byte
	_, err := fmt.Sscanf(string(b[1:len(b)-1]), "%x", &tidBytes)
	if err != nil {
		return errors.New("could not unmarshal types.BlockID: " + err.Error())
	}
	copy(tid[:], tidBytes)
	return nil
}

// MarshalJSON marshales an id as a hex string.
func (oid OutputID) MarshalJSON() ([]byte, error) {
	return json.Marshal(oid.String())
}

// String prints the id in hex.
func (oid OutputID) String() string {
	return fmt.Sprintf("%x", oid[:])
}

// UnmarshalJSON decodes the json hex string of the id.
func (oid *OutputID) UnmarshalJSON(b []byte) error {
	if len(b) != crypto.HashSize*2+2 {
		return crypto.ErrHashWrongLen
	}

	var oidBytes []byte
	_, err := fmt.Sscanf(string(b[1:len(b)-1]), "%x", &oidBytes)
	if err != nil {
		return errors.New("could not unmarshal types.BlockID: " + err.Error())
	}
	copy(oid[:], oidBytes)
	return nil
}

// MarshalJSON marshales an id as a hex string.
func (scoid CoinOutputID) MarshalJSON() ([]byte, error) {
	return json.Marshal(scoid.String())
}

// String prints the id in hex.
func (scoid CoinOutputID) String() string {
	return fmt.Sprintf("%x", scoid[:])
}

// UnmarshalJSON decodes the json hex string of the id.
func (scoid *CoinOutputID) UnmarshalJSON(b []byte) error {
	if len(b) != crypto.HashSize*2+2 {
		return crypto.ErrHashWrongLen
	}

	var scoidBytes []byte
	_, err := fmt.Sscanf(string(b[1:len(b)-1]), "%x", &scoidBytes)
	if err != nil {
		return errors.New("could not unmarshal types.BlockID: " + err.Error())
	}
	copy(scoid[:], scoidBytes)
	return nil
}

// MarshalJSON marshales an id as a hex string.
func (bsoid BlockStakeOutputID) MarshalJSON() ([]byte, error) {
	return json.Marshal(bsoid.String())
}

// String prints the id in hex.
func (bsoid BlockStakeOutputID) String() string {
	return fmt.Sprintf("%x", bsoid[:])
}

// UnmarshalJSON decodes the json hex string of the id.
func (bsoid *BlockStakeOutputID) UnmarshalJSON(b []byte) error {
	if len(b) != crypto.HashSize*2+2 {
		return crypto.ErrHashWrongLen
	}

	var bsoidBytes []byte
	_, err := fmt.Sscanf(string(b[1:len(b)-1]), "%x", &bsoidBytes)
	if err != nil {
		return errors.New("could not unmarshal types.BlockID: " + err.Error())
	}
	copy(bsoid[:], bsoidBytes)
	return nil
}
