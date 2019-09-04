package types

import (
	"encoding/json"
	"errors"
	"io"

	"github.com/threefoldtech/rivine/pkg/encoding/rivbin"
	"github.com/threefoldtech/rivine/pkg/encoding/siabin"
)

// legacy transaction-related types,
// as to support the JSON/binary (un)marshalling of transaction
// (related) structs
type (
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

// newLegacyTransactionDataFromTransaction creates a legacy transaction from an
// in-memory Transaction. It does so by exposing the data to the newLegacyTransactionData
// constructor, and thus piggy-backing for the actual logic on that one.
func newLegacyTransactionDataFromTransaction(t Transaction) (legacyTransactionData, error) {
	return newLegacyTransactionData(TransactionData{
		CoinInputs:        t.CoinInputs,
		CoinOutputs:       t.CoinOutputs,
		BlockStakeInputs:  t.BlockStakeInputs,
		BlockStakeOutputs: t.BlockStakeOutputs,
		MinerFees:         t.MinerFees,
		ArbitraryData:     t.ArbitraryData,
	})
}

// newLegacyTransactionData creates legacy transaction data (as part of v0 transactions),
// using the given transaction data in the newer (in-memory) format, passed as input.
func newLegacyTransactionData(data TransactionData) (ltd legacyTransactionData, err error) {
	ltd.CoinInputs = make([]legacyTransactionCoinInput, len(data.CoinInputs))
	for i, ci := range data.CoinInputs {
		ltd.CoinInputs[i] = legacyTransactionCoinInput{
			ParentID: ci.ParentID,
			Unlocker: legacyTransactionInputLockProxy{
				Fulfillment: ci.Fulfillment.Fulfillment,
			},
		}
	}
	if l := len(data.CoinOutputs); l > 0 {
		ltd.CoinOutputs = make([]legacyTransactionCoinOutput, l)
		for i, co := range data.CoinOutputs {
			uhc, ok := co.Condition.Condition.(*UnlockHashCondition)
			if !ok {
				err = errors.New("only unlock hash conditions are supported for legacy transactions")
				return
			}
			ltd.CoinOutputs[i] = legacyTransactionCoinOutput{
				Value:      co.Value,
				UnlockHash: uhc.TargetUnlockHash,
			}
		}
	}

	if l := len(data.BlockStakeInputs); l > 0 {
		ltd.BlockStakeInputs = make([]legacyTransactionBlockStakeInput, len(data.BlockStakeInputs))
		for i, bsi := range data.BlockStakeInputs {
			ltd.BlockStakeInputs[i] = legacyTransactionBlockStakeInput{
				ParentID: bsi.ParentID,
				Unlocker: legacyTransactionInputLockProxy{
					Fulfillment: bsi.Fulfillment.Fulfillment,
				},
			}
		}
	}
	if l := len(data.BlockStakeOutputs); l > 0 {
		ltd.BlockStakeOutputs = make([]legacyTransactionBlockStakeOutput, l)
		for i, bso := range data.BlockStakeOutputs {
			uhc, ok := bso.Condition.Condition.(*UnlockHashCondition)
			if !ok {
				err = errors.New("only unlock hash conditions are supported for legacy transactions")
				return
			}
			ltd.BlockStakeOutputs[i] = legacyTransactionBlockStakeOutput{
				Value:      bso.Value,
				UnlockHash: uhc.TargetUnlockHash,
			}
		}
	}

	ltd.MinerFees, ltd.ArbitraryData = data.MinerFees, data.ArbitraryData
	return
}

// TransactionData returns this legacy TransactionData,
// in the new TransactionData format.
func (ltd legacyTransactionData) TransactionData() (data TransactionData) {
	data.CoinInputs = make([]CoinInput, len(ltd.CoinInputs))
	for i, lci := range ltd.CoinInputs {
		data.CoinInputs[i] = CoinInput{
			ParentID:    lci.ParentID,
			Fulfillment: NewFulfillment(lci.Unlocker.Fulfillment),
		}
	}
	if l := len(ltd.CoinOutputs); l > 0 {
		data.CoinOutputs = make([]CoinOutput, l)
		for i, lco := range ltd.CoinOutputs {
			data.CoinOutputs[i] = CoinOutput{
				Value:     lco.Value,
				Condition: NewCondition(NewUnlockHashCondition(lco.UnlockHash)),
			}
		}
	}

	if l := len(ltd.BlockStakeInputs); l > 0 {
		data.BlockStakeInputs = make([]BlockStakeInput, l)
		for i, lci := range ltd.BlockStakeInputs {
			data.BlockStakeInputs[i] = BlockStakeInput{
				ParentID:    lci.ParentID,
				Fulfillment: NewFulfillment(lci.Unlocker.Fulfillment),
			}
		}
	}
	if l := len(ltd.BlockStakeOutputs); l > 0 {
		data.BlockStakeOutputs = make([]BlockStakeOutput, l)
		for i, lco := range ltd.BlockStakeOutputs {
			data.BlockStakeOutputs[i] = BlockStakeOutput{
				Value:     lco.Value,
				Condition: NewCondition(NewUnlockHashCondition(lco.UnlockHash)),
			}
		}
	}

	data.MinerFees, data.ArbitraryData = ltd.MinerFees, ltd.ArbitraryData
	return
}

// MarshalSia implements siabin.SiaMarshaller.MarshalSia
//
// Encodes the legacy fulfillment as it used to be done,
// as a so-called legacy input lock.
func (ilp legacyTransactionInputLockProxy) MarshalSia(w io.Writer) error {
	switch tc := ilp.Fulfillment.(type) {
	case *SingleSignatureFulfillment:
		pkBytes, err := siabin.Marshal(tc.PublicKey)
		if err != nil {
			return err
		}
		return siabin.NewEncoder(w).EncodeAll(FulfillmentTypeSingleSignature, pkBytes, tc.Signature)
	case *LegacyAtomicSwapFulfillment:
		conditionBytes, err := siabin.MarshalAll(tc.Sender, tc.Receiver, tc.HashedSecret, tc.TimeLock)
		if err != nil {
			return err
		}
		fulfillmentBytes, err := siabin.MarshalAll(tc.PublicKey, tc.Signature, tc.Secret)
		if err != nil {
			return err
		}
		return siabin.NewEncoder(w).EncodeAll(FulfillmentTypeAtomicSwap, conditionBytes, fulfillmentBytes)
	default:
		return errors.New("unlock type is invalid in a v0 transaction")
	}
}

// UnmarshalSia implements siabin.SiaUnmarshaller.UnmarshalSia
//
// Decodes the legacy fulfillment as it used to be done,
// from a so-called legacy input lock structure.
func (ilp *legacyTransactionInputLockProxy) UnmarshalSia(r io.Reader) (err error) {
	var (
		unlockType UnlockType
		decoder    = siabin.NewDecoder(r)
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
		err = siabin.Unmarshal(cb, &ss.PublicKey)
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
		err = siabin.UnmarshalAll(cb, &as.Sender, &as.Receiver, &as.HashedSecret, &as.TimeLock)
		if err != nil {
			return
		}
		err = siabin.UnmarshalAll(fb, &as.PublicKey, &as.Signature, &as.Secret)
		ilp.Fulfillment = as
		return

	default:
		err = errors.New("v0 transactions only support single-signature and atomic-swap unlock conditions")
		return
	}
}

// MarshalRivine implements siabin.RivineMarshaler.MarshalRivine
//
// Encodes the legacy fulfillment as it used to be done,
// as a so-called legacy input lock.
func (ilp legacyTransactionInputLockProxy) MarshalRivine(w io.Writer) error {
	switch tc := ilp.Fulfillment.(type) {
	case *SingleSignatureFulfillment:
		pkBytes, err := rivbin.Marshal(tc.PublicKey)
		if err != nil {
			return err
		}
		return rivbin.NewEncoder(w).EncodeAll(FulfillmentTypeSingleSignature, pkBytes, tc.Signature)
	case *LegacyAtomicSwapFulfillment:
		conditionBytes, err := rivbin.MarshalAll(tc.Sender, tc.Receiver, tc.HashedSecret, tc.TimeLock)
		if err != nil {
			return err
		}
		fulfillmentBytes, err := rivbin.MarshalAll(tc.PublicKey, tc.Signature, tc.Secret)
		if err != nil {
			return err
		}
		return rivbin.NewEncoder(w).EncodeAll(FulfillmentTypeAtomicSwap, conditionBytes, fulfillmentBytes)
	default:
		return errors.New("unlock type is invalid in a v0 transaction")
	}
}

// UnmarshalRivine implements rivbin.RivineUnmarshaler.UnmarshalRivine
//
// Decodes the legacy fulfillment as it used to be done,
// from a so-called legacy input lock structure.
func (ilp *legacyTransactionInputLockProxy) UnmarshalRivine(r io.Reader) (err error) {
	var (
		unlockType UnlockType
		decoder    = rivbin.NewDecoder(r)
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
		err = rivbin.Unmarshal(cb, &ss.PublicKey)
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
		err = rivbin.UnmarshalAll(cb, &as.Sender, &as.Receiver, &as.HashedSecret, &as.TimeLock)
		if err != nil {
			return
		}
		err = rivbin.UnmarshalAll(fb, &as.PublicKey, &as.Signature, &as.Secret)
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
		PublicKey PublicKey `json:"publickey"`
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
