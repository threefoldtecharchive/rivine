package types

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/rivine/rivine/build"
	"github.com/rivine/rivine/encoding"
)

// legacy transaction-related types,
// as to support the JSON/binary (un)marshalling of transaction
// (related) structs
type (
	legacyTransaction struct {
		Version TransactionVersion    `json:"version"`
		Data    legacyTransactionData `json:"data"`
	}
	legacyTransactionData struct {
		CoinInputs        []legacyTransactionCoinInput        `json:"coininputs"`
		CoinOutputs       []legacyTransactionCoinOutput       `json:"coinoutputs,omitempty"`
		BlockStakeInputs  []legacyTransactionBlockStakeInput  `json:"blockstakeinputs,omitempty"`
		BlockStakeOutputs []legacyTransactionBlockStakeOutput `json:"blockstakeoutputs,omitempty"`
		MinerFees         []Currency                          `json:"minerfees"`
		ArbitraryData     []byte                              `json:"arbitrarydata,omitempty"`
	}
	legacyTransactionCoinInput struct {
		ParentID CoinOutputID                    `json:"parentid"`
		Unlocker legacyTransactionInputLockProxy `json:"unlocker"`
	}
	legacyTransactionCoinOutput struct {
		Value      Currency   `json:"value"`
		UnlockHash UnlockHash `json:"unlockhash"`
	}
	legacyTransactionBlockStakeInput struct {
		ParentID BlockStakeOutputID              `json:"parentid"`
		Unlocker legacyTransactionInputLockProxy `json:"unlocker"`
	}
	legacyTransactionBlockStakeOutput struct {
		Value      Currency   `json:"value"`
		UnlockHash UnlockHash `json:"unlockhash"`
	}
	legacyTransactionInputLockProxy struct {
		Fulfillment MarshalableUnlockFulfillment
	}
)

// newLegacyTransaction creates a legacy transaction (v0 transaction),
// using the Transaction in the newer (in-memory) format, passed as input.
func newLegacyTransaction(t Transaction) (lt legacyTransaction, err error) {
	if build.DEBUG && t.Version != TransactionVersionZero {
		panic(fmt.Sprintf("unexpected transaction version %v", t.Version))
	}
	lt.Version = t.Version

	lt.Data.CoinInputs = make([]legacyTransactionCoinInput, len(t.CoinInputs))
	for i, ci := range t.CoinInputs {
		lt.Data.CoinInputs[i] = legacyTransactionCoinInput{
			ParentID: ci.ParentID,
			Unlocker: legacyTransactionInputLockProxy{
				Fulfillment: ci.Fulfillment.Fulfillment,
			},
		}
	}
	if l := len(t.CoinOutputs); l > 0 {
		lt.Data.CoinOutputs = make([]legacyTransactionCoinOutput, l)
		for i, co := range t.CoinOutputs {
			uhc, ok := co.Condition.Condition.(*UnlockHashCondition)
			if !ok {
				err = errors.New("only unlock hash conditions are supported for legacy transactions")
				return
			}
			lt.Data.CoinOutputs[i] = legacyTransactionCoinOutput{
				Value:      co.Value,
				UnlockHash: uhc.TargetUnlockHash,
			}
		}
	}

	if l := len(t.BlockStakeInputs); l > 0 {
		lt.Data.BlockStakeInputs = make([]legacyTransactionBlockStakeInput, len(t.BlockStakeInputs))
		for i, bsi := range t.BlockStakeInputs {
			lt.Data.BlockStakeInputs[i] = legacyTransactionBlockStakeInput{
				ParentID: bsi.ParentID,
				Unlocker: legacyTransactionInputLockProxy{
					Fulfillment: bsi.Fulfillment.Fulfillment,
				},
			}
		}
	}
	if l := len(t.BlockStakeOutputs); l > 0 {
		lt.Data.BlockStakeOutputs = make([]legacyTransactionBlockStakeOutput, l)
		for i, bso := range t.BlockStakeOutputs {
			uhc, ok := bso.Condition.Condition.(*UnlockHashCondition)
			if !ok {
				err = errors.New("only unlock hash conditions are supported for legacy transactions")
				return
			}
			lt.Data.BlockStakeOutputs[i] = legacyTransactionBlockStakeOutput{
				Value:      bso.Value,
				UnlockHash: uhc.TargetUnlockHash,
			}
		}
	}

	lt.Data.MinerFees, lt.Data.ArbitraryData = t.MinerFees, t.ArbitraryData
	return
}

// Transaction returns this legacy Transaction,
// in the new Transaction format.
func (lt legacyTransaction) Transaction() (t Transaction) {
	if build.DEBUG && lt.Version != TransactionVersionZero {
		panic(fmt.Sprintf("unexpected transaction version %v", lt.Version))
	}
	t.Version = lt.Version

	t.CoinInputs = make([]CoinInput, len(lt.Data.CoinInputs))
	for i, lci := range lt.Data.CoinInputs {
		t.CoinInputs[i] = CoinInput{
			ParentID:    lci.ParentID,
			Fulfillment: NewFulfillment(lci.Unlocker.Fulfillment),
		}
	}
	if l := len(lt.Data.CoinOutputs); l > 0 {
		t.CoinOutputs = make([]CoinOutput, l)
		for i, lco := range lt.Data.CoinOutputs {
			t.CoinOutputs[i] = CoinOutput{
				Value:     lco.Value,
				Condition: NewCondition(NewUnlockHashCondition(lco.UnlockHash)),
			}
		}
	}

	if l := len(lt.Data.BlockStakeInputs); l > 0 {
		t.BlockStakeInputs = make([]BlockStakeInput, l)
		for i, lci := range lt.Data.BlockStakeInputs {
			t.BlockStakeInputs[i] = BlockStakeInput{
				ParentID:    lci.ParentID,
				Fulfillment: NewFulfillment(lci.Unlocker.Fulfillment),
			}
		}
	}
	if l := len(lt.Data.BlockStakeOutputs); l > 0 {
		t.BlockStakeOutputs = make([]BlockStakeOutput, l)
		for i, lco := range lt.Data.BlockStakeOutputs {
			t.BlockStakeOutputs[i] = BlockStakeOutput{
				Value:     lco.Value,
				Condition: NewCondition(NewUnlockHashCondition(lco.UnlockHash)),
			}
		}
	}

	t.MinerFees, t.ArbitraryData = lt.Data.MinerFees, lt.Data.ArbitraryData
	return
}

// MarshalSia implements encoding.SiaMarshaller.MarshalSia
//
// Encodes the legacy fulfillment as it used to be done,
// as a so-called legacy input lock.
func (ilp legacyTransactionInputLockProxy) MarshalSia(w io.Writer) error {
	switch tc := ilp.Fulfillment.(type) {
	case *SingleSignatureFulfillment:
		return encoding.NewEncoder(w).EncodeAll(FulfillmentTypeSingleSignature,
			encoding.Marshal(tc.PublicKey), tc.Signature)
	case *LegacyAtomicSwapFulfillment:
		return encoding.NewEncoder(w).EncodeAll(FulfillmentTypeAtomicSwap,
			encoding.MarshalAll(tc.Sender, tc.Receiver, tc.HashedSecret, tc.TimeLock),
			encoding.MarshalAll(tc.PublicKey, tc.Signature, tc.Secret))
	default:
		return errors.New("unlock type is invalid in a v0 transaction")
	}
}

// UnmarshalSia implements encoding.SiaUnmarshaller.UnmarshalSia
//
// Decodes the legacy fulfillment as it used to be done,
// from a so-called legacy input lock structure.
func (ilp *legacyTransactionInputLockProxy) UnmarshalSia(r io.Reader) (err error) {
	var (
		unlockType UnlockType
		decoder    = encoding.NewDecoder(r)
	)
	err = decoder.Decode(&unlockType)
	if err != nil {
		return
	}
	switch unlockType {
	case UnlockTypePubKey:
		var cb []byte
		err = decoder.Decode(&cb)
		if err != nil {
			return
		}

		ss := new(SingleSignatureFulfillment)
		err = encoding.Unmarshal(cb, &ss.PublicKey)
		if err != nil {
			return err
		}
		err = decoder.Decode(&ss.Signature)
		ilp.Fulfillment = ss
		return

	case UnlockTypeAtomicSwap:
		var cb, fb []byte
		err = decoder.DecodeAll(&cb, &fb)
		if err != nil {
			return
		}
		as := new(LegacyAtomicSwapFulfillment)
		err = encoding.UnmarshalAll(cb, &as.Sender, &as.Receiver, &as.HashedSecret, &as.TimeLock)
		if err != nil {
			return
		}
		err = encoding.UnmarshalAll(fb, &as.PublicKey, &as.Signature, &as.Secret)
		ilp.Fulfillment = as
		return

	default:
		err = errors.New("v0 transactions only support single-signature and atomic-swap unlock conditions")
		return
	}
}

type (
	legacyJSONInputLockProxy struct {
		Type        UnlockType      `json:"type,omitempty"`
		Condition   json.RawMessage `json:"condition,omitempty"`
		Fulfillment json.RawMessage `json:"fulfillment,omitempty"`
	}
	legacyJSONSingleSignatureCondition struct {
		PublicKey SiaPublicKey `json:"publickey"`
	}
	legacyJSONSingleSignatureFulfillment struct {
		Signature ByteSlice `json:"signature"`
	}
)

// MarshalJSON implements json.Marshaller.MarshalJSON
//
// Encodes the legacy fulfillment as it used to be done,
// as a so-called legacy input lock.
func (ilp legacyTransactionInputLockProxy) MarshalJSON() ([]byte, error) {
	var (
		err error
		out legacyJSONInputLockProxy
	)
	switch tc := ilp.Fulfillment.(type) {
	case *SingleSignatureFulfillment:
		out.Type = UnlockTypePubKey
		out.Condition, err = json.Marshal(legacyJSONSingleSignatureCondition{tc.PublicKey})
		if err != nil {
			return nil, err
		}
		out.Fulfillment, err = json.Marshal(legacyJSONSingleSignatureFulfillment{tc.Signature})
		if err != nil {
			return nil, err
		}

	case *LegacyAtomicSwapFulfillment:
		out.Type = UnlockTypeAtomicSwap
		out.Condition, err = json.Marshal(AtomicSwapCondition{
			Sender:       tc.Sender,
			Receiver:     tc.Receiver,
			HashedSecret: tc.HashedSecret,
			TimeLock:     tc.TimeLock,
		})
		if err != nil {
			return nil, err
		}
		out.Fulfillment, err = json.Marshal(AtomicSwapFulfillment{
			PublicKey: tc.PublicKey,
			Signature: tc.Signature,
			Secret:    tc.Secret,
		})
		if err != nil {
			return nil, err
		}
	default:
		return nil, errors.New("unlock type is invalid in a v0 transaction")
	}
	return json.Marshal(out)
}

// UnmarshalJSON implements json.Unmarshaller.UnmarshalJSON
//
// Decodes the legacy fulfillment as it used to be done,
// from a so-called legacy input lock structure.
func (ilp *legacyTransactionInputLockProxy) UnmarshalJSON(b []byte) error {
	var in legacyJSONInputLockProxy
	err := json.Unmarshal(b, &in)
	if err != nil {
		return err
	}
	switch in.Type {
	case UnlockTypePubKey:
		var (
			jc legacyJSONSingleSignatureCondition
			jf legacyJSONSingleSignatureFulfillment
		)
		err = json.Unmarshal(in.Condition, &jc)
		if err != nil {
			return err
		}
		err = json.Unmarshal(in.Fulfillment, &jf)
		if err != nil {
			return err
		}
		ilp.Fulfillment = &SingleSignatureFulfillment{
			PublicKey: jc.PublicKey,
			Signature: jf.Signature,
		}
		return nil

	case UnlockTypeAtomicSwap:
		var (
			jc AtomicSwapCondition
			jf AtomicSwapFulfillment
		)
		err = json.Unmarshal(in.Condition, &jc)
		if err != nil {
			return err
		}
		err = json.Unmarshal(in.Fulfillment, &jf)
		if err != nil {
			return err
		}
		ilp.Fulfillment = &LegacyAtomicSwapFulfillment{
			Sender:       jc.Sender,
			Receiver:     jc.Receiver,
			HashedSecret: jc.HashedSecret,
			TimeLock:     jc.TimeLock,
			PublicKey:    jf.PublicKey,
			Signature:    jf.Signature,
			Secret:       jf.Secret,
		}
		return nil

	default:
		return errors.New("unlock type is invalid in a v0 transaction")
	}
}
