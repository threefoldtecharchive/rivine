package types

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/rivine/rivine/build"
	"github.com/rivine/rivine/crypto"
	"github.com/rivine/rivine/encoding"
)

// The interfaces and input parameter structs that make the unlock conditions tick.
type (
	// UnlockCondition defines the condition that has to be fulfilled
	// using a UnlockFulfillment which supports the implemented unlock condition.
	//
	// An UnlockCondition can be used/supported by multiple UnlockFulfillments.
	//
	// An unlock condtion will be propagated even when it is non-standard,
	// but in that case it will not be minted into a block,
	// unless a (block) creator sees it as standard.
	UnlockCondition interface {
		// Fulfill this condition using the given fulfillment, if supported,
		// within the given (fulfill) context.
		// An error is to be returned in case the given UnlockFulfillment
		// cannot fulfill the UnlockCondition (within the given context).
		Fulfill(fulfillment UnlockFulfillment, ctx FulfillContext) error

		// ConditionType returns the condition type of the UnlockCondtion.
		ConditionType() ConditionType
		// IsStandardCondtion returns if this (unlock) condition is standard.
		// An error result means it is not standard,
		// and it will consequently prevent the transaction, which includes the
		// the output that has this condition, from being minted into a block.
		IsStandardCondition() error

		// UnlockHash returns the deterministic unlock hash of this UnlockCondition.
		// It identifies the owner(s) or contract which own the output,
		// and can spend it, once the conditions becomes `Fulfillable`.
		UnlockHash() UnlockHash

		// Equal returns if the given unlock condition
		// equals the called unlock condition.
		Equal(UnlockCondition) bool

		// Fulfillable returns true if the unlock condition can be fullfilled,
		// should the right fulfillment be given with the same (fulfill)
		// context as given.
		//
		// NOTE: that this method does assume that the condition is a standard condition.
		Fulfillable(FulfillableContext) bool
	}
	// MarshalableUnlockCondition adds binary marshaling as a required interface
	// to the regular unlock condition interface. This allows the condition
	// to be used as an internal condition of an UnlockConditionProxy.
	//
	// NOTE: it is also expected that an implementation implementing this interface,
	// supports JSON marshaling/unmarshaling, whether this is implicit through tags,
	// or explicit by implementing the relevant JSON interfaces is of no importance.
	// This is however not enforced, even though it is expected.
	MarshalableUnlockCondition interface {
		UnlockCondition

		// Marshal this UnlockCondition into a binary format.
		// The returned byte slice should be usable in order
		// to recreate the same unlock condition using the paired
		// Unmarshal method.
		Marshal() []byte
		// Unmarshal this unlock condition from a binary format,
		// the whole byte slice is expected to be used.
		Unmarshal([]byte) error
	}

	// UnlockFulfillment defines the fulfillment that fulfills
	// one or multiple UnlockConditions.
	//
	// An unlock fulfillment will be propagated even when it is non-standard,
	// but in that case it will not be minted into a block,
	// unless a (block) creator sees it as standard.
	UnlockFulfillment interface {
		// Sign the given fulfillment, which is to be done after all properties
		// have been filled of the parent transaction
		// (including the unsigned fulfillments of all inputs).
		//
		// The signing is to be done within the given (fulfillment sign) context.
		Sign(ctx FulfillmentSignContext) error

		// UnlockHash returns the unlock hash of this UnlockFulfillment.
		// It identifies the spender of the output linked by the input's parent ID.
		//
		// The UnlockHash, as returned by the UnlockFulfillment,
		// does not have to be deterministic and can depend on how it fulfills a condition.
		// (e.g. if multiple people can fulfill by means of signing,
		//       than the returned unlock hash might depend upon who signed it)
		UnlockHash() UnlockHash

		// Equal returns if the given unlock fulfillment
		// equals the called unlock fulfillment.
		Equal(UnlockFulfillment) bool

		// FulfillmentType returns the fulfillment type of the UnlockFulfillment.
		FulfillmentType() FulfillmentType
		// IsStandardFulfillment returns if this (unlock) fulfillment is standard.
		// An error result means it is not standard,
		// and it will consequently prevent the transaction, which includes the
		// the input that has this fulfillment, from being minted into a block.
		IsStandardFulfillment() error
	}
	// MarshalableUnlockFulfillment adds binary marshaling as a required interface
	// to the regular unlock fulfillment interface. This allows the fulfillment
	// to be used as an internal fulfillment of an UnlockFulfillmentProxy.
	//
	// NOTE: it is also expected that an implementation implementing this interface,
	// supports JSON marshaling/unmarshaling, whether this is implicit through tags,
	// or explicit by implementing the relevant JSON interfaces is of no importance.
	// This is however not enforced, even though it is expected.
	MarshalableUnlockFulfillment interface {
		UnlockFulfillment

		// Marshal this UnlockFulfillment into a binary format.
		// The returned byte slice should be usable in order
		// to recreate the same unlock fulfillment using the paired
		// Unmarshal method.
		Marshal() []byte
		// Unmarshal this unlock fulfillment from a binary format,
		// the whole byte slice is expected to be used.
		Unmarshal([]byte) error
	}

	// UnlockConditionProxy wraps around a (binary/json) marshalable
	// UnlockCondition, as to allow implicit nil conditions,
	// as well as implicit and universal json/binary unmarshaling
	// and binary marshaling.
	UnlockConditionProxy struct {
		Condition MarshalableUnlockCondition
	}
	// UnlockFulfillmentProxy wraps around a (binary/json) marshalable
	// UnlockFulfillment, as to allow implicit nil fulfillments,
	// as well as implicit and universal json/binary unmarshaling
	// and binary marshaling.
	UnlockFulfillmentProxy struct {
		Fulfillment MarshalableUnlockFulfillment
	}

	// FulfillmentSignContext is given as part of the sign call of an UnlockFullment,
	// as to provide the necessary context required for signing a fulfillment.
	FulfillmentSignContext struct {
		// Index of the input that is to be signed,
		// whether that input is a coin- or blockStake- input is of no importance.
		InputIndex uint64
		// (Parent) transaction the fulfillment belongs to.
		//
		// It is expected that all properties of this transactions
		// are defined at this point, even though it is allowed that one or multiple
		// inputs haven't been signed yet, and this should have no influence on the signature.
		Transaction Transaction
		// (Private) key to be used for signing, what type it is or whether it is defined at all
		// is of no importance, as long as the fulfillment supports its (none) definition.
		Key interface{}
	}

	// FulfillContext is given as part of the fulfill call of an UnlockCondition,
	// as to provide the necessary context required for fulfilling a fulfillment.
	FulfillContext struct {
		// Index of the input that is to be fulfilled,
		// whether that input is a coin- or blockStake- input is of no importance.
		InputIndex uint64
		// BlockHeight of the currently last registered block.
		BlockHeight BlockHeight
		// BlockTime defines the time of the currently last registered block,
		// the transaction belonged to.
		BlockTime Timestamp
		// (Parent) transaction the fulfillment belongs to.
		Transaction Transaction
	}

	// FulfillableContext is given as part of the fulfillable call of an UnlockCondition,
	// as to provide the necessary context required for fulfilling a fulfillment.
	FulfillableContext struct {
		// BlockHeight of the currently last registered block.
		BlockHeight BlockHeight
		// BlockTime defines the time of the currently last registered block,
		// the transaction belonged to.
		BlockTime Timestamp
	}

	// ConditionType defines the type of a condition.
	ConditionType byte
	// FulfillmentType defines the type of a fulfillment.
	FulfillmentType byte
)

// The following enumeration defines the different possible and standard
// unlock conditions. All are defined by an implementation of MarshalableUnlockCondition.
const (
	// ConditionTypeNil defines the nil condition,
	// meaning it explicitly defines that the output can be used by anyone.
	// Whoever claims the output as input does have to sign the input,
	// using SingleSignatureFulfillment.
	//
	// Implemented by the NilCondition type.
	ConditionTypeNil ConditionType = iota
	// ConditionTypeUnlockHash defines a condition which is to be unlocked,
	// by a fulfillment which produces the specified unlock hash.
	//
	// Depending upon the unlock type of the specified unlock hash,
	// a different type of fulfillment can fulfill this condition.
	// If the hash's unlock type is a UnlockTypePubKey,
	// it can be fulfilled by a SingleSignatureFulfillment.
	// If the hash's unlock type is a UnlockTypeAtomicSwap,
	// it can be fulfilled by a LegacyAtomicSwapFulfillment.
	// No other unlock type is supported by this unlock condition.
	//
	// Implemented by the UnlockHashCondition type.
	ConditionTypeUnlockHash
	// ConditionTypeAtomicSwap defines the (new) atomic swap condition,
	// replacing the legacy UnlockHash-based atomic swap condition.
	// It can be fulfilled only by an AtomicSwapFulfillment.
	//
	// Implemented by the AtomicSwapCondition type.
	ConditionTypeAtomicSwap

	// ConditionTypeTimeLock defines an unlock condition
	// which locks another condition with a timestamp.
	// The internal condition has to be one of: [
	// NilCondition,
	// UnlockHashCondition(0x01 unlock hash type is the only standard one at the moment, others are always fulfilled),
	// ]
	ConditionTypeTimeLock
)

// The following enumeration defines the different possible and standard
// unlock fulfillments. All are defined by an implementation of MarshalableUnlockFulfillment.
const (
	// FulfillmentTypeNil defines the nil fulfillment,
	// and cannot fulfill any condition. It only serves as the nil-fulfillment,
	// which is the default value for a fulfillment type, should none be given.
	// It is considered non-standard.
	//
	// A NilCondition is fulfilled by the SingleSignatureFulfillment,
	// not the NilFulfillment. This as to ensure all fulfillments are secured by a signature,
	// as a protection against tampering.
	//
	// Implemented by the NilFulfillment type.
	FulfillmentTypeNil FulfillmentType = iota
	// FulfillmentTypeSingleSignature defines the single signature fulfillment,
	// and is defined by a public key, matching a private key which is used to
	// produce the given transaction-based signature.
	//
	// It is used to fulfill UnlockHashConditions which are described using an unlock hash
	// of type UnlockTypePubKey, as well as NilConditions.
	//
	// Implemented by the SingleSignatureFulfillment type.
	FulfillmentTypeSingleSignature
	// FulfillmentTypeAtomicSwap defines both the legacy as well as the new
	// AtomicSwapFulfillment. Whether the fulfillemnt is of the legacy or the latter type,
	// depends upon the properties that are defined as part of the fulfillment.
	//
	// In legacy mode it is used to fulfill UnlockHashConditions which are described using
	// an unlock hash of type UnlockTypeAtomicSwap. While the (new) atomic swap fulfillments
	// are used to fulfill AtomicSwapConditions.
	//
	// Implemented by the AtomicSwapFulfillment.
	FulfillmentTypeAtomicSwap
)

// Constants that are used as part of AtomicSwap Conditions/Fulfillments.
const (
	// AtomicSwapSecretLen is the required/fixed length
	// of an atomic swap secret, the pre-image of an hashed secret.
	AtomicSwapSecretLen = sha256.Size
	// AtomicSwapHashedSecretLen is the required/fixed length
	// of an atomic swap hashed secret, the post-image of a secret.
	AtomicSwapHashedSecretLen = sha256.Size
)

// Various errors returned by unlock conditions and fulfillments.
var (
	// ErrUnexpectedUnlockCondition is returnd when a fulfillment is given
	// an UnlockCondition of an unexpected type.
	ErrUnexpectedUnlockCondition = errors.New("unexpected unlock condition")

	// ErrUnexpectedUnlockFulfillment is returnd when an UnlockCondition is given
	// an UnlockFulfillment of an unexpected type.
	ErrUnexpectedUnlockFulfillment = errors.New("unexpected unlock fulfillment")

	// ErrUnexpectedUnlockType is returned when an unlock hash has the wrong type.
	ErrUnexpectedUnlockType = errors.New("unexpected unlock (hash) type")

	// ErrFulfillmentDoubleSign is returned when a fulfillment that is already signed,
	// is attempted to be signed once again.
	ErrFulfillmentDoubleSign = errors.New("cannot sign a fulfillment which is already signed")

	// ErrUnknownConditionType is returned to define the non-standardness
	// of an UnknownUnlockCondition.
	ErrUnknownConditionType = errors.New("unknown condition type")
	// ErrUnknownFulfillmentType is returned to define the non-standardness
	// of an UnknownUnlockFulfillment.
	ErrUnknownFulfillmentType = errors.New("unknown fulfillment type")

	// ErrNilFulfillmentType is returned by pretty much any method of the
	// NilFullfilment type, as it is not to be used for anything.
	ErrNilFulfillmentType = errors.New("nil fulfillment type")

	// ErrUnknownSignAlgorithmType is an error returned in case
	// one tries to sign using an unknown signing algorithm type.
	//
	// NOTE That verification of unknown signing algorithm types does always succeed!
	ErrUnknownSignAlgorithmType = errors.New("unknown signature algorithm type")
)

// RegisterUnlockConditionType is used to register a condition type, by linking it to
// a constructor which constructs a fresh MarshalableUnlockCondition each time it is called.
//
// RegisterUnlockConditionType can also used to unregister a condition type,
// by calling this funciton with nil as the MarshalableUnlockConditionConstructor.
func RegisterUnlockConditionType(ct ConditionType, cc MarshalableUnlockConditionConstructor) {
	if cc == nil {
		delete(_RegisteredUnlockConditionTypes, ct)
		return
	}
	_RegisteredUnlockConditionTypes[ct] = cc
}

// RegisterUnlockFulfillmentType is used to register a fulfillment type, by linking it to
// a constructor which constructs a fresh MarshalableUnlockFulfillment each time it is called.
//
// RegisterUnlockFulfillmentType can also used to unregister a fulfillment type,
// by calling this funciton with nil as the MarshalableUnlockFulfillmentConstructor.
func RegisterUnlockFulfillmentType(ft FulfillmentType, fc MarshalableUnlockFulfillmentConstructor) {
	if fc == nil {
		delete(_RegisteredUnlockFulfillmentTypes, ft)
		return
	}
	_RegisteredUnlockFulfillmentTypes[ft] = fc
}

// Constuctors used to construct marshalable unlock conditions and fulfillments.
type (
	// MarshalableUnlockConditionConstructor defines a constructor type,
	// which is expected to construct a new MarshalableUnlockCondition each time it is called.
	MarshalableUnlockConditionConstructor func() MarshalableUnlockCondition
	// MarshalableUnlockFulfillmentConstructor defines a constructor type,
	// which is expected to construct a new MarshalableUnlockFulfillment each time it is called.
	MarshalableUnlockFulfillmentConstructor func() MarshalableUnlockFulfillment
)

// Hidden globals used to collect the standard as well as the user-defined
// unlock condition/fulfillment types, each linked to a constructor,
// which construct an instance of the type that implement the unlock condition/fulfillment type.
var (
	// Manipulated by the RegisterUnlockConditionType function,
	// and used by the UnlockConditionProxy.
	_RegisteredUnlockConditionTypes = map[ConditionType]MarshalableUnlockConditionConstructor{
		ConditionTypeNil:        func() MarshalableUnlockCondition { return &NilCondition{} },
		ConditionTypeUnlockHash: func() MarshalableUnlockCondition { return &UnlockHashCondition{} },
		ConditionTypeAtomicSwap: func() MarshalableUnlockCondition { return &AtomicSwapCondition{} },
		ConditionTypeTimeLock:   func() MarshalableUnlockCondition { return &TimeLockCondition{} },
	}
	// Manipulated by the RegisterUnlockFulfillmentType function,
	// and used by the UnlockFulfillmentProxy.
	_RegisteredUnlockFulfillmentTypes = map[FulfillmentType]MarshalableUnlockFulfillmentConstructor{
		FulfillmentTypeNil:             func() MarshalableUnlockFulfillment { return &NilFulfillment{} },
		FulfillmentTypeSingleSignature: func() MarshalableUnlockFulfillment { return &SingleSignatureFulfillment{} },
		FulfillmentTypeAtomicSwap:      func() MarshalableUnlockFulfillment { return &anyAtomicSwapFulfillment{} },
	}
)

// NewCondition creates an optional unlock condition,
// using an optionally given MarshalableUnlockCondition.
func NewCondition(c MarshalableUnlockCondition) UnlockConditionProxy {
	return UnlockConditionProxy{Condition: c}
}

// NewFulfillment creates an optional unlock fulfillment,
// using an optionally given MarshalableUnlockFulfillment.
func NewFulfillment(f MarshalableUnlockFulfillment) UnlockFulfillmentProxy {
	return UnlockFulfillmentProxy{Fulfillment: f}
}

type (
	// NilCondition implements the ConditionTypeNil (unlock) ConditionType.
	// See ConditionTypeNil for more information.
	NilCondition struct{} // can only be fulfilled by a SingleSignatureFulfillment
	// NilFulfillment implements the FulfillmentTypeNil (unlock) FulfillmentType.
	// See FulfillmentTypeNil for more information.
	NilFulfillment struct{} // invalid fulfillment

	// UnlockHashCondition implements the ConditionTypeUnlockHash (unlock) ConditionType.
	// See ConditionTypeUnlockHash for more information.
	UnlockHashCondition struct {
		TargetUnlockHash UnlockHash `json:"unlockhash"`
	}
	// SingleSignatureFulfillment implements the FulfillmentTypeSingleSignature (unlock) FulfillmentType.
	// See FulfillmentTypeSingleSignature for more information.
	SingleSignatureFulfillment struct {
		PublicKey SiaPublicKey `json:"publickey"`
		Signature ByteSlice    `json:"signature"`
	}

	// AtomicSwapCondition implements the ConditionTypeSingleSignature (unlock) ConditionType.
	// See ConditionTypeSingleSignature for more information.
	AtomicSwapCondition struct {
		Sender       UnlockHash             `json:"sender"`
		Receiver     UnlockHash             `json:"receiver"`
		HashedSecret AtomicSwapHashedSecret `json:"hashedsecret"`
		TimeLock     Timestamp              `json:"timelock"`
	}
	// AtomicSwapFulfillment implements the (new) FulfillmentTypeAtomicSwap (unlock) FulfillmentType.
	// See FulfillmentTypeAtomicSwap for more information.
	AtomicSwapFulfillment struct {
		PublicKey SiaPublicKey     `json:"publickey"`
		Signature ByteSlice        `json:"signature"`
		Secret    AtomicSwapSecret `json:"secret,omitempty"`
	}
	// LegacyAtomicSwapFulfillment implements the (legacy) FulfillmentTypeAtomicSwap (unlock) FulfillmentType.
	// See FulfillmentTypeAtomicSwap for more information.
	LegacyAtomicSwapFulfillment struct { // legacy fulfillment as used in transactions of version 0
		Sender       UnlockHash             `json:"sender"`
		Receiver     UnlockHash             `json:"receiver"`
		HashedSecret AtomicSwapHashedSecret `json:"hashedsecret"`
		TimeLock     Timestamp              `json:"timelock"`
		PublicKey    SiaPublicKey           `json:"publickey"`
		Signature    ByteSlice              `json:"signature"`
		Secret       AtomicSwapSecret       `json:"secret,omitempty"`
	}
	// AtomicSwapSecret defines the 256 pre-image byte slice,
	// used as secret within the Atomic Swap protocol/contract.
	AtomicSwapSecret [sha256.Size]byte
	// AtomicSwapHashedSecret defines the 256 image byte slice,
	// used as hashed secret within the Atomic Swap protocol/contract.
	AtomicSwapHashedSecret [sha256.Size]byte

	// TimeLockCondition defines an unlock condition which require a LockTime
	// to be reached on top of some other defined condition,
	// which both have to be fulfilled in order to unlock/spend/use the unspend output as an input.
	TimeLockCondition struct {
		// LockTime defines either a block height or a timestamp.
		// If the value is less than LockTimeMinTimestampValue it is concidered a lock based on block height,
		// otherwise it is used as a unix epoch value expressed in seconds.
		LockTime uint64
		// Condition defines the condition which has to be fulfilled
		// on top of the LockTime condition defined by this condition.
		// See ConditionTypeTimeLock in order to know which conditions are supported.
		Condition MarshalableUnlockCondition
	}
)

const (
	// LockTimeMinTimestampValue defines the minimum value a LockTime can be
	// in order to be interpreted as a (unix epoch seconds) timestamp,
	// otherwise it is interpreted as the block height instead.
	LockTimeMinTimestampValue = 500 * 1000 * 1000
)

type (
	// anyAtomicSwapFulfillment is used to be able to unmarshal an atomic swap fulfillment,
	// no matter if it's in the legacy format or in the original format.
	anyAtomicSwapFulfillment struct {
		MarshalableUnlockFulfillment
	}
)

// Errors related to atomic swap validation.
var (
	// ErrInvalidPreImageSha256 is returned as the result of a failed fulfillment,
	// in case the condition-defined hashed secret (pre image) does not match
	// the fulfillment-defined secret (image).
	ErrInvalidPreImageSha256 = errors.New("invalid pre-image sha256")
	// ErrInvalidRedeemer is returned in case the redeemer, one of two parties,
	// is the wrong redeemer due to the timelock rule.
	// Prior to the timelock only the receiver can redeem,
	// while after that timelock only the sender can redeem.
	ErrInvalidRedeemer = errors.New("invalid input redeemer")
)

// Fulfill implements UnlockCondition.Fulfill
func (n *NilCondition) Fulfill(fulfillment UnlockFulfillment, ctx FulfillContext) error {
	switch tf := fulfillment.(type) {
	case *SingleSignatureFulfillment:
		return verifyHashUsingSiaPublicKey(tf.PublicKey,
			ctx.InputIndex, ctx.Transaction, tf.Signature)
	default:
		return ErrUnexpectedUnlockFulfillment
	}
}

// ConditionType implements UnlockCondition.ConditionType
func (n *NilCondition) ConditionType() ConditionType { return ConditionTypeNil }

// IsStandardCondition implements UnlockCondition.IsStandardCondition
func (n *NilCondition) IsStandardCondition() error { return nil } // always valid

// UnlockHash implements UnlockCondition.UnlockHash
func (n *NilCondition) UnlockHash() UnlockHash { return NilUnlockHash }

// Equal implements UnlockCondition.Equal
func (n *NilCondition) Equal(c UnlockCondition) bool {
	if c == nil {
		return true // implicit equality
	}
	_, equal := c.(*NilCondition)
	return equal // explicit equality
}

// Fulfillable implements UnlockCondition.Fulfillable
func (n *NilCondition) Fulfillable(FulfillableContext) bool { return true }

// Marshal implements MarshalableUnlockCondition.Marshal
func (n *NilCondition) Marshal() []byte { return nil } // nothing to marshal
// Unmarshal implements MarshalableUnlockCondition.Unmarshal
func (n *NilCondition) Unmarshal(b []byte) error { return nil } // nothing to unmarshal

// Sign implements UnlockFulfillment.Sign
func (n *NilFulfillment) Sign(FulfillmentSignContext) error { return ErrNilFulfillmentType }

// UnlockHash implements UnlockFulfillment.UnlockHash
func (n *NilFulfillment) UnlockHash() UnlockHash { return NilUnlockHash }

// Equal implements UnlockFulfillment.Equal
func (n *NilFulfillment) Equal(f UnlockFulfillment) bool {
	if f == nil {
		return true // implicit equality
	}
	_, equal := f.(*NilFulfillment)
	return equal // explicit equality
}

// FulfillmentType implements UnlockFulfillment.FulfillmentType
func (n *NilFulfillment) FulfillmentType() FulfillmentType { return FulfillmentTypeNil }

// IsStandardFulfillment implements UnlockFulfillment.IsStandardFulfillment
func (n *NilFulfillment) IsStandardFulfillment() error { return ErrNilFulfillmentType } // never valid

// Marshal implements MarshalableUnlockFulfillment.Marshal
func (n *NilFulfillment) Marshal() []byte {
	if build.DEBUG {
		panic(ErrNilFulfillmentType)
	}
	return nil // nothing to marshal
}

// Unmarshal implements MarshalableUnlockFulfillment.Unmarshal
func (n *NilFulfillment) Unmarshal([]byte) error { return ErrNilFulfillmentType } // cannot be unmarshaled

// NewUnlockHashCondition creates a new unlock condition,
// using a (target) unlock hash as the condtion to be fulfilled.
func NewUnlockHashCondition(uh UnlockHash) *UnlockHashCondition {
	return &UnlockHashCondition{TargetUnlockHash: uh}
}

// Fulfill implements UnlockCondition.Fulfill
func (uh *UnlockHashCondition) Fulfill(fulfillment UnlockFulfillment, ctx FulfillContext) error {
	switch tf := fulfillment.(type) {
	case *SingleSignatureFulfillment:
		// only UnlockTypePubKey is supported when fulfilling using a SingleSignatureFulfillment
		if uh.TargetUnlockHash.Type != UnlockTypePubKey {
			return ErrUnexpectedUnlockType
		}

		euh := tf.UnlockHash()
		if euh != uh.TargetUnlockHash {
			return errors.New("single signature fulfillment provides wrong public key")
		}
		return verifyHashUsingSiaPublicKey(tf.PublicKey,
			ctx.InputIndex, ctx.Transaction, tf.Signature)

	case *LegacyAtomicSwapFulfillment:
		// only UnlockTypeAtomicSwap is supported when fulfilling using a LegacyAtomicSwapFulfillment
		if uh.TargetUnlockHash.Type != UnlockTypeAtomicSwap {
			return ErrUnexpectedUnlockType
		}

		// ensure the condition equals the ours
		ourHS := tf.UnlockHash()
		if ourHS.Cmp(uh.TargetUnlockHash) != 0 {
			return errors.New("produced unlock hash doesn't equal the expected unlock hash")
		}

		// create the unlockHash for the given public Key
		unlockHash := NewPubKeyUnlockHash(tf.PublicKey)

		// prior to our timelock, only the receiver can claim the unspend output
		if ctx.BlockTime <= tf.TimeLock {
			// verify that receiver public key was given
			if unlockHash.Cmp(tf.Receiver) != 0 {
				return ErrInvalidRedeemer
			}

			// verify signature
			err := verifyHashUsingSiaPublicKey(
				tf.PublicKey, ctx.InputIndex, ctx.Transaction, tf.Signature,
				tf.PublicKey, tf.Secret)
			if err != nil {
				return err
			}

			// in order for the receiver to spend,
			// the secret has to be known
			hashedSecret := NewAtomicSwapHashedSecret(tf.Secret)
			if bytes.Compare(tf.HashedSecret[:], hashedSecret[:]) != 0 {
				return ErrInvalidPreImageSha256
			}

			return nil
		}

		// verify that sender public key was given
		if unlockHash.Cmp(tf.Sender) != 0 {
			return ErrInvalidRedeemer
		}

		// after the deadline (timelock),
		// only the original sender can reclaim the unspend output
		return verifyHashUsingSiaPublicKey(
			tf.PublicKey, ctx.InputIndex, ctx.Transaction, tf.Signature,
			tf.PublicKey)

	default:
		return ErrUnexpectedUnlockFulfillment
	}
}

// ConditionType implements UnlockCondition.ConditionType
func (uh *UnlockHashCondition) ConditionType() ConditionType { return ConditionTypeUnlockHash }

// IsStandardCondition implements UnlockCondition.IsStandardCondition
func (uh *UnlockHashCondition) IsStandardCondition() error {
	if uh.TargetUnlockHash.Type != UnlockTypePubKey && uh.TargetUnlockHash.Type != UnlockTypeAtomicSwap {
		return fmt.Errorf("unsupported unlock type '%d' by unlock hash condition", uh.TargetUnlockHash.Type)
	}
	if uh.TargetUnlockHash.Hash == (crypto.Hash{}) {
		return errors.New("nil crypto hash cannot be used as unlock hash")
	}
	return nil
}

// UnlockHash implements UnlockCondition.UnlockHash
func (uh *UnlockHashCondition) UnlockHash() UnlockHash {
	return uh.TargetUnlockHash
}

// Equal implements UnlockCondition.Equal
func (uh *UnlockHashCondition) Equal(c UnlockCondition) bool {
	ouh, ok := c.(*UnlockHashCondition)
	if !ok {
		return false
	}
	return uh.TargetUnlockHash.Cmp(ouh.TargetUnlockHash) == 0
}

// Fulfillable implements UnlockCondition.Fulfillable
func (uh *UnlockHashCondition) Fulfillable(FulfillableContext) bool { return true }

// Marshal implements MarshalableUnlockCondition.Marshal
func (uh *UnlockHashCondition) Marshal() []byte {
	return encoding.Marshal(uh.TargetUnlockHash)
}

// Unmarshal implements MarshalableUnlockCondition.Unmarshal
func (uh *UnlockHashCondition) Unmarshal(b []byte) error {
	return encoding.Unmarshal(b, &uh.TargetUnlockHash)
}

// NewSingleSignatureFulfillment creates an unsigned SingleSignatureFulfillment,
// using the given Public Key, which is to be matched with the private key given
// as part of the later sign call to the returned instance.
func NewSingleSignatureFulfillment(pk SiaPublicKey) *SingleSignatureFulfillment {
	return &SingleSignatureFulfillment{PublicKey: pk}
}

// Sign implements UnlockFulfillment.Sign
func (ss *SingleSignatureFulfillment) Sign(ctx FulfillmentSignContext) (err error) {
	if len(ss.Signature) != 0 {
		return ErrFulfillmentDoubleSign
	}

	ss.Signature, err = signHashUsingSiaPublicKey(
		ss.PublicKey, ctx.InputIndex, ctx.Transaction, ctx.Key)
	return
}

// UnlockHash implements UnlockFulfillment.UnlockHash
func (ss *SingleSignatureFulfillment) UnlockHash() UnlockHash {
	return NewPubKeyUnlockHash(ss.PublicKey)
}

// FulfillmentType implements UnlockFulfillment.FulfillmentType
func (ss *SingleSignatureFulfillment) FulfillmentType() FulfillmentType {
	return FulfillmentTypeSingleSignature
}

// IsStandardFulfillment implements UnlockFulfillment.IsStandardFulfillment
func (ss *SingleSignatureFulfillment) IsStandardFulfillment() error {
	return strictSignatureCheck(ss.PublicKey, ss.Signature)
}

// Equal implements UnlockFulfillment.Equal
func (ss *SingleSignatureFulfillment) Equal(f UnlockFulfillment) bool {
	oss, ok := f.(*SingleSignatureFulfillment)
	if !ok {
		return false
	}
	if ss.PublicKey.Algorithm != oss.PublicKey.Algorithm {
		return false
	}
	if bytes.Compare(ss.PublicKey.Key[:], oss.PublicKey.Key[:]) != 0 {
		return false
	}
	return bytes.Compare(ss.Signature[:], oss.Signature[:]) == 0
}

// Marshal implements MarshalableUnlockFulfillment.Marshal
func (ss *SingleSignatureFulfillment) Marshal() []byte {
	return encoding.MarshalAll(ss.PublicKey, ss.Signature)
}

// Unmarshal implements MarshalableUnlockFulfillment.Unmarshal
func (ss *SingleSignatureFulfillment) Unmarshal(b []byte) error {
	return encoding.UnmarshalAll(b, &ss.PublicKey, &ss.Signature)
}

// Fulfill implements UnlockCondition.Fulfill
func (as *AtomicSwapCondition) Fulfill(fulfillment UnlockFulfillment, ctx FulfillContext) error {
	switch tf := fulfillment.(type) {
	case *AtomicSwapFulfillment:
		// create the unlockHash for the given public Ke
		unlockHash := NewUnlockHash(UnlockTypePubKey,
			crypto.HashObject(encoding.Marshal(tf.PublicKey)))
		// prior to our timelock, only the receiver can claim the unspend output
		if ctx.BlockTime <= as.TimeLock {
			// verify that receiver public key was given
			if unlockHash.Cmp(as.Receiver) != 0 {
				return ErrInvalidRedeemer
			}

			// verify signature
			err := verifyHashUsingSiaPublicKey(
				tf.PublicKey, ctx.InputIndex, ctx.Transaction, tf.Signature,
				tf.PublicKey, tf.Secret)
			if err != nil {
				return err
			}

			// in order for the receiver to spend,
			// the secret has to be known
			hashedSecret := NewAtomicSwapHashedSecret(tf.Secret)
			if bytes.Compare(as.HashedSecret[:], hashedSecret[:]) != 0 {
				return ErrInvalidPreImageSha256
			}

			return nil
		}

		// verify that sender public key was given
		if unlockHash.Cmp(as.Sender) != 0 {
			return ErrInvalidRedeemer
		}

		// after the deadline (timelock),
		// only the original sender can reclaim the unspend output
		return verifyHashUsingSiaPublicKey(
			tf.PublicKey, ctx.InputIndex, ctx.Transaction, tf.Signature,
			tf.PublicKey)

	case *LegacyAtomicSwapFulfillment:
		// it's perfectly fine to unlock an atomic swap condition
		// using an atomic swap format in the legacy format,
		// as long as all properties check out
		if as.Sender.Cmp(tf.Sender) != 0 {
			return errors.New("legacy atomic swap fulfillment defines an incorrect sender")
		}
		if as.Receiver.Cmp(tf.Receiver) != 0 {
			return errors.New("legacy atomic swap fulfillment defines an incorrect receiver")
		}
		if as.TimeLock != tf.TimeLock {
			return errors.New("legacy atomic swap fulfillment defines an incorrect time lock")
		}
		if bytes.Compare(as.HashedSecret[:], tf.HashedSecret[:]) != 0 {
			return errors.New("legacy atomic swap fulfillment defines an incorrect hashed secret")
		}
		// delegate logic to the fulfillment in the new format,
		// by calling this method once again
		return as.Fulfill(&AtomicSwapFulfillment{
			PublicKey: tf.PublicKey,
			Signature: tf.Signature,
			Secret:    tf.Secret,
		}, ctx)

	default:
		return ErrUnexpectedUnlockFulfillment
	}
}

// ConditionType implements UnlockCondition.ConditionType
func (as *AtomicSwapCondition) ConditionType() ConditionType { return ConditionTypeAtomicSwap }

// IsStandardCondition implements UnlockCondition.IsStandardCondition
func (as *AtomicSwapCondition) IsStandardCondition() error {
	if as.Sender.Type != UnlockTypePubKey {
		return fmt.Errorf("unsupported unlock hash sender type: %d", as.Sender.Type)
	}
	if as.Receiver.Type != UnlockTypePubKey {
		return fmt.Errorf("unsupported unlock hash receiver type: %d", as.Receiver.Type)
	}
	if as.Sender.Hash == (crypto.Hash{}) || as.Receiver.Hash == (crypto.Hash{}) {
		return errors.New("nil crypto hash cannot be used as unlock hash")
	}
	if as.HashedSecret == (AtomicSwapHashedSecret{}) {
		return errors.New("nil hashed secret not allowed")
	}
	return nil
}

// UnlockHash implements UnlockCondition.UnlockHash
func (as *AtomicSwapCondition) UnlockHash() UnlockHash {
	return NewUnlockHash(UnlockTypeAtomicSwap, crypto.HashObject(as.Marshal()))
}

// Equal implements UnlockCondition.Equal
func (as *AtomicSwapCondition) Equal(c UnlockCondition) bool {
	oas, ok := c.(*AtomicSwapCondition)
	if !ok {
		return false
	}
	if as.TimeLock != oas.TimeLock {
		return false
	}
	if bytes.Compare(as.HashedSecret[:], oas.HashedSecret[:]) != 0 {
		return false
	}
	if as.Sender.Cmp(oas.Sender) != 0 {
		return false
	}
	return as.Receiver.Cmp(oas.Receiver) == 0
}

// Fulfillable implements UnlockCondition.Fulfillable
func (as *AtomicSwapCondition) Fulfillable(FulfillableContext) bool { return true }

// Marshal implements MarshalableUnlockCondition.Marshal
func (as *AtomicSwapCondition) Marshal() []byte {
	return encoding.MarshalAll(as.Sender, as.Receiver, as.HashedSecret, as.TimeLock)
}

// Unmarshal implements MarshalableUnlockCondition.Unmarshal
func (as *AtomicSwapCondition) Unmarshal(b []byte) error {
	return encoding.UnmarshalAll(b, &as.Sender, &as.Receiver, &as.HashedSecret, &as.TimeLock)
}

// NewAtomicSwapClaimFulfillment creates an unsigned atomic swap fulfillment,
// as to spend an output as a claim (meaning redeeming the money as the receiver).
//
// Returned fulfillment still has to be signed, as to add the signature,
// with the parent transaction as the input as well as the matching private key.
//
// Note that this fulfillment will fail if the current time is
// equal to or past the timestamp specified as time lock in the parent output.
func NewAtomicSwapClaimFulfillment(pk SiaPublicKey, secret AtomicSwapSecret) *AtomicSwapFulfillment {
	return &AtomicSwapFulfillment{
		PublicKey: pk,
		Secret:    secret,
	}
}

// NewAtomicSwapRefundFulfillment creates an unsigned atomic swap fulfillment,
// as to get a refund (meaning redeeming the money as the sender).
//
// Returned fulfillment still has to be signed, as to add the signature,
// with the parent transaction as the input as well as the matching private key.
//
// Note that this fulfillment will fail if the current time is
// prior to the timestamp specified as time lock in the parent output.
func NewAtomicSwapRefundFulfillment(pk SiaPublicKey) *AtomicSwapFulfillment {
	return &AtomicSwapFulfillment{PublicKey: pk}
}

// Sign implements UnlockFulfillment.Sign
func (as *AtomicSwapFulfillment) Sign(ctx FulfillmentSignContext) error {
	if len(as.Signature) != 0 {
		return ErrFulfillmentDoubleSign
	}

	if as.Secret != (AtomicSwapSecret{}) {
		// sign as claimer
		var err error
		as.Signature, err = signHashUsingSiaPublicKey(
			as.PublicKey, ctx.InputIndex, ctx.Transaction, ctx.Key,
			as.PublicKey, as.Secret)
		return err
	}

	// sign as refunder
	var err error
	as.Signature, err = signHashUsingSiaPublicKey(
		as.PublicKey, ctx.InputIndex, ctx.Transaction, ctx.Key,
		as.PublicKey)
	return err
}

// UnlockHash implements UnlockFulfillment.UnlockHash
func (as *AtomicSwapFulfillment) UnlockHash() UnlockHash {
	return NewUnlockHash(UnlockTypeAtomicSwap,
		crypto.HashObject(encoding.Marshal(as.PublicKey)))
}

// FulfillmentType implements UnlockFulfillment.FulfillmentType
func (as *AtomicSwapFulfillment) FulfillmentType() FulfillmentType { return FulfillmentTypeAtomicSwap }

// IsStandardFulfillment implements UnlockFulfillment.IsStandardFulfillment
func (as *AtomicSwapFulfillment) IsStandardFulfillment() error {
	return strictSignatureCheck(as.PublicKey, as.Signature)
}

// Equal implements UnlockFulfillment.Equal
func (as *AtomicSwapFulfillment) Equal(f UnlockFulfillment) bool {
	oas, ok := f.(*AtomicSwapFulfillment)
	if !ok {
		return false
	}
	if as.PublicKey.Algorithm != oas.PublicKey.Algorithm {
		return false
	}
	if bytes.Compare(as.PublicKey.Key[:], oas.PublicKey.Key[:]) != 0 {
		return false
	}
	if bytes.Compare(as.Signature[:], oas.Signature[:]) != 0 {
		return false
	}
	return bytes.Compare(as.Secret[:], oas.Secret[:]) == 0
}

// Marshal implements MarshalableUnlockFulfillment.Marshal
func (as *AtomicSwapFulfillment) Marshal() []byte {
	return encoding.MarshalAll(as.PublicKey, as.Signature, as.Secret)
}

// Unmarshal implements MarshalableUnlockFulfillment.Unmarshal
func (as *AtomicSwapFulfillment) Unmarshal(b []byte) error {
	return encoding.UnmarshalAll(b, &as.PublicKey, &as.Signature, &as.Secret)
}

// AtomicSwapSecret returns the AtomicSwapSecret defined in this legacy fulfillmen
func (as *AtomicSwapFulfillment) AtomicSwapSecret() AtomicSwapSecret {
	return as.Secret
}

// Sign implements UnlockFulfillment.Sign
func (as *LegacyAtomicSwapFulfillment) Sign(ctx FulfillmentSignContext) error {
	if len(as.Signature) != 0 {
		return ErrFulfillmentDoubleSign
	}
	if as.Secret != (AtomicSwapSecret{}) {
		// sign as claimer
		var err error
		as.Signature, err = signHashUsingSiaPublicKey(
			as.PublicKey, ctx.InputIndex, ctx.Transaction, ctx.Key,
			as.PublicKey, as.Secret)
		return err
	}

	// sign as refunder
	var err error
	as.Signature, err = signHashUsingSiaPublicKey(
		as.PublicKey, ctx.InputIndex, ctx.Transaction, ctx.Key,
		as.PublicKey)
	return err
}

// UnlockHash implements UnlockFulfillment.UnlockHash
func (as *LegacyAtomicSwapFulfillment) UnlockHash() UnlockHash {
	return NewUnlockHash(UnlockTypeAtomicSwap, crypto.HashObject(encoding.MarshalAll(
		as.Sender,
		as.Receiver,
		as.HashedSecret,
		as.TimeLock,
	)))
}

// FulfillmentType implements UnlockFulfillment.FulfillmentType
func (as *LegacyAtomicSwapFulfillment) FulfillmentType() FulfillmentType {
	return FulfillmentTypeAtomicSwap
}

// IsStandardFulfillment implements UnlockFulfillment.IsStandardFulfillment
func (as *LegacyAtomicSwapFulfillment) IsStandardFulfillment() error {
	if as.Sender.Type != UnlockTypePubKey || as.Receiver.Type != UnlockTypePubKey {
		return errors.New("unsupported unlock hash type")
	}
	if as.Sender.Hash == (crypto.Hash{}) || as.Receiver.Hash == (crypto.Hash{}) {
		return errors.New("nil crypto hash cannot be used as unlock hash")
	}
	if as.HashedSecret == (AtomicSwapHashedSecret{}) {
		return errors.New("nil hashed secret not allowed")
	}
	return strictSignatureCheck(as.PublicKey, as.Signature)
}

// Equal implements UnlockFulfillment.Equal
func (as *LegacyAtomicSwapFulfillment) Equal(f UnlockFulfillment) bool {
	olas, ok := f.(*LegacyAtomicSwapFulfillment)
	if !ok {
		return false
	}
	if as.TimeLock != olas.TimeLock {
		return false
	}
	if bytes.Compare(as.HashedSecret[:], olas.HashedSecret[:]) != 0 {
		return false
	}
	if as.Sender.Cmp(olas.Sender) != 0 {
		return false
	}
	if as.Receiver.Cmp(olas.Receiver) != 0 {
		return false
	}
	if as.PublicKey.Algorithm != olas.PublicKey.Algorithm {
		return false
	}
	if bytes.Compare(as.PublicKey.Key[:], olas.PublicKey.Key[:]) != 0 {
		return false
	}
	if bytes.Compare(as.Signature[:], olas.Signature[:]) != 0 {
		return false
	}
	return bytes.Compare(as.Secret[:], olas.Secret[:]) == 0
}

// Marshal implements MarshalableUnlockFulfillment.Marshal
func (as *LegacyAtomicSwapFulfillment) Marshal() []byte {
	return encoding.MarshalAll(
		as.Sender, as.Receiver, as.HashedSecret, as.TimeLock,
		as.PublicKey, as.Signature, as.Secret)
}

// Unmarshal implements MarshalableUnlockFulfillment.Unmarshal
func (as *LegacyAtomicSwapFulfillment) Unmarshal(b []byte) error {
	return encoding.UnmarshalAll(b,
		&as.Sender, &as.Receiver, &as.HashedSecret, &as.TimeLock,
		&as.PublicKey, &as.Signature, &as.Secret)
}

// AtomicSwapSecret returns the AtomicSwapSecret defined in this legacy fulfillment.
func (as *LegacyAtomicSwapFulfillment) AtomicSwapSecret() AtomicSwapSecret {
	return as.Secret
}

var (
	_ MarshalableUnlockCondition = (*NilCondition)(nil)
	_ MarshalableUnlockCondition = (*UnlockHashCondition)(nil)
	_ MarshalableUnlockCondition = (*AtomicSwapCondition)(nil)

	_ MarshalableUnlockFulfillment = (*NilFulfillment)(nil)
	_ MarshalableUnlockFulfillment = (*SingleSignatureFulfillment)(nil)
	_ MarshalableUnlockFulfillment = (*AtomicSwapFulfillment)(nil)
	_ MarshalableUnlockFulfillment = (*LegacyAtomicSwapFulfillment)(nil)
)

// NewAtomicSwapHashedSecret creates a new atomic swap hashed secret,
// using a pre-generated atomic swap secret.
func NewAtomicSwapHashedSecret(secret AtomicSwapSecret) AtomicSwapHashedSecret {
	return AtomicSwapHashedSecret(sha256.Sum256(secret[:]))
}

// String turns this hashed secret into a hex-formatted string.
func (hs AtomicSwapHashedSecret) String() string {
	return hex.EncodeToString(hs[:])
}

// LoadString loads a hashed secret from a hex-formatted string.
func (hs *AtomicSwapHashedSecret) LoadString(str string) error {
	n, err := hex.Decode(hs[:], []byte(str))
	if err != nil {
		return err
	}
	if n != AtomicSwapHashedSecretLen {
		return errors.New("invalid (atomic-swap) hashed secret length")
	}
	return nil
}

// MarshalJSON marshals a hashed secret as a hex string.
func (hs AtomicSwapHashedSecret) MarshalJSON() ([]byte, error) {
	return json.Marshal(hs.String())
}

// UnmarshalJSON decodes the json string of the hashed secret.
func (hs *AtomicSwapHashedSecret) UnmarshalJSON(b []byte) error {
	var str string
	if err := json.Unmarshal(b, &str); err != nil {
		return err
	}
	return hs.LoadString(str)
}

var (
	_ json.Marshaler   = AtomicSwapHashedSecret{}
	_ json.Unmarshaler = (*AtomicSwapHashedSecret)(nil)
)

// NewAtomicSwapSecret creates a new cryptographically secure
// atomic swap secret
func NewAtomicSwapSecret() (ass AtomicSwapSecret, err error) {
	_, err = rand.Read(ass[:])
	return
}

// String turns this secret into a hex-formatted string.
func (s AtomicSwapSecret) String() string {
	return hex.EncodeToString(s[:])
}

// LoadString loads a secret from a hex-formatted string.
func (s *AtomicSwapSecret) LoadString(str string) error {
	n, err := hex.Decode(s[:], []byte(str))
	if err != nil {
		return err
	}
	if n != AtomicSwapSecretLen {
		return errors.New("invalid (atomic-swap) secret length")
	}
	return nil
}

// MarshalJSON marshals a secret as a hex string.
func (s AtomicSwapSecret) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

// UnmarshalJSON decodes the json string of the secret.
func (s *AtomicSwapSecret) UnmarshalJSON(b []byte) error {
	var str string
	if err := json.Unmarshal(b, &str); err != nil {
		return err
	}
	return s.LoadString(str)
}

var (
	_ json.Marshaler   = AtomicSwapSecret{}
	_ json.Unmarshaler = (*AtomicSwapSecret)(nil)
)

// Unmarshal implements MarshalableUnlockFulfillment.Unmarshal
//
// It ensures we can marshal an atomic swap fulfillment,
// no matter if it is in the legacy format or the new format.
//
// No need to support the paired marshal interface,
// as that is done implicitly using the child fulfillment.
func (as *anyAtomicSwapFulfillment) Unmarshal(b []byte) error {
	asf := new(AtomicSwapFulfillment)
	// be positive, first try the new format
	err := encoding.Unmarshal(b, asf)
	if err == nil {
		as.MarshalableUnlockFulfillment = asf
		return nil
	}

	// didn't work out, let's try the legacy atomic swap fulfillment
	lasf := new(LegacyAtomicSwapFulfillment)
	err = encoding.Unmarshal(b, lasf)
	as.MarshalableUnlockFulfillment = lasf
	return err
}

// MarshalJSON implements json.Marshaler.MarshalJSON
//
// This function is required, as to ensure the underlying/child
// atomic swap fulfillment is marshaled on its own,
// rather than as part of this in-memory structure,
// as it is not supposed to be visible from an encoded perspective.
func (as *anyAtomicSwapFulfillment) MarshalJSON() ([]byte, error) {
	return json.Marshal(as.MarshalableUnlockFulfillment)
}

// UnmarshalJSON implements json.Unmarshaler.UnmarshalJSON
//
// It ensures we can marshal an atomic swap fulfillment,
// no matter if it is in the legacy format or the new format.
//
// No need to support the paired marshal interface,
// as that is done implicitly using the child fulfillment.
func (as *anyAtomicSwapFulfillment) UnmarshalJSON(b []byte) error {
	lasf := new(LegacyAtomicSwapFulfillment)
	err := json.Unmarshal(b, lasf)
	if err != nil {
		return err
	}
	var undefOptArgCount uint8
	if lasf.Sender.Cmp(UnlockHash{}) == 0 {
		undefOptArgCount++
	}
	if lasf.Receiver.Cmp(UnlockHash{}) == 0 {
		undefOptArgCount++
	}
	if lasf.TimeLock == 0 {
		undefOptArgCount++
	}
	if nilHS := (AtomicSwapHashedSecret{}); bytes.Compare(lasf.HashedSecret[:], nilHS[:]) == 0 {
		undefOptArgCount++
	}
	switch undefOptArgCount {
	case 0:
		as.MarshalableUnlockFulfillment = lasf
	case 4:
		as.MarshalableUnlockFulfillment = &AtomicSwapFulfillment{
			PublicKey: lasf.PublicKey,
			Signature: lasf.Signature,
			Secret:    lasf.Secret,
		}
	default:
		return errors.New("when an atomic swap fulfillment defines any of the legacy properties, all of them have to be given")
	}
	return nil
}

var (
	_ MarshalableUnlockFulfillment = (*anyAtomicSwapFulfillment)(nil)
)

// NewTimeLockCondition creates a new TimeLockCondition.
// If no MarshalableUnlockCondition is given, the NilCondition is assumed.
func NewTimeLockCondition(lockTime uint64, condition MarshalableUnlockCondition) *TimeLockCondition {
	if build.DEBUG && lockTime == 0 {
		panic("lock time is required")
	}
	if condition == nil {
		condition = &NilCondition{}
	}
	return &TimeLockCondition{
		LockTime:  lockTime,
		Condition: condition,
	}
}

// Fulfill implements UnlockFulfillment.Fulfill
//
// The TimeLockFulfillment can only be used to fulfill a TimeLockCondition.
func (tl *TimeLockCondition) Fulfill(fulfillment UnlockFulfillment, ctx FulfillContext) error {
	if !tl.Fulfillable(FulfillableContext{BlockHeight: ctx.BlockHeight, BlockTime: ctx.BlockTime}) {
		return errors.New("time lock has not yet been reached")
	}

	// time lock hash been reached,
	// delegate the actual fulfillment to the given fulfillment, if supported
	switch tf := fulfillment.(type) {
	case *SingleSignatureFulfillment:
		return tl.Condition.Fulfill(tf, ctx)
	default:
		return ErrUnexpectedUnlockFulfillment
	}
}

// ConditionType implements UnlockCondition.ConditionType
func (tl *TimeLockCondition) ConditionType() ConditionType { return ConditionTypeTimeLock }

// IsStandardCondition implements UnlockCondition.IsStandardCondition
func (tl *TimeLockCondition) IsStandardCondition() error {
	if tl.LockTime == 0 {
		return errors.New("lock time has to be defined")
	}
	switch tc := tl.Condition.(type) {
	case *UnlockHashCondition:
		if tc.TargetUnlockHash.Type != UnlockTypePubKey {
			return fmt.Errorf("unsupported unlock type '%d' by unlock hash condition", tc.TargetUnlockHash.Type)
		}
		if tc.TargetUnlockHash.Hash == (crypto.Hash{}) {
			return errors.New("nil crypto hash cannot be used as unlock hash")
		}
		return nil
	case *NilCondition:
		return nil
	default:
		return errors.New("unexpected internal unlock condition used as part of time lock condition")
	}
}

// UnlockHash implements UnlockCondition.UnlockHash
func (tl *TimeLockCondition) UnlockHash() UnlockHash {
	return tl.Condition.UnlockHash()
}

// Equal implements UnlockCondition.Equal
func (tl *TimeLockCondition) Equal(c UnlockCondition) bool {
	otl, ok := c.(*TimeLockCondition)
	if !ok {
		return false
	}
	return tl.LockTime == otl.LockTime && tl.Condition.Equal(otl.Condition)
}

// Fulfillable implements UnlockCondition.Fulfillable
func (tl *TimeLockCondition) Fulfillable(ctx FulfillableContext) bool {
	if tl.LockTime < LockTimeMinTimestampValue {
		return BlockHeight(tl.LockTime) <= ctx.BlockHeight
	}
	return Timestamp(tl.LockTime) <= ctx.BlockTime
}

// Marshal implements MarshalableUnlockCondition.Marshal
func (tl *TimeLockCondition) Marshal() []byte {
	return append(
		encoding.MarshalAll(tl.LockTime, tl.Condition.ConditionType()),
		tl.Condition.Marshal()...)
}

// Unmarshal implements MarshalableUnlockCondition.Unmarshal
func (tl *TimeLockCondition) Unmarshal(b []byte) error {
	if len(b) < 9 {
		// at least 9 bytes are required (lock time (8) + condition type (1)),
		// as to enforce we can decode the time lock condition's properties,
		// whether or not the internal condition requires bytes is of no concern of us.
		return io.ErrUnexpectedEOF
	}
	// unmarshal the lock time
	err := encoding.Unmarshal(b[:8], &tl.LockTime)
	if err != nil {
		return err
	}
	// interpret the condition type, and continue decoding based on that,
	// by getting the correct constructor from the registration mapping
	ct := ConditionType(b[8])
	cc, ok := _RegisteredUnlockConditionTypes[ct]
	if !ok {
		return ErrUnknownConditionType
	}
	// known condition type, create and decode it
	tl.Condition = cc()
	return tl.Condition.Unmarshal(b[9:])
}

type jsonTimeLockCondition struct {
	LockTime  uint64               `json:"locktime"`
	Condition UnlockConditionProxy `json:"condition"`
}

// MarshalJSON implements json.Marshaler.MarshalJSON
//
// This function is required, as to ensure
// the underlying properties are properly serialized,
// including the type of the internal condition.
func (tl *TimeLockCondition) MarshalJSON() ([]byte, error) {
	return json.Marshal(jsonTimeLockCondition{
		LockTime:  tl.LockTime,
		Condition: UnlockConditionProxy{Condition: tl.Condition},
	})
}

// UnmarshalJSON implements json.Unmarshaler.UnmarshalJSON
//
// This function is required, as to be able to unmarshal
// the internal condition based on the encoded condition type.
func (tl *TimeLockCondition) UnmarshalJSON(b []byte) error {
	// first unmarshal the top-layered time lock condition
	var jtl jsonTimeLockCondition
	err := json.Unmarshal(b, &jtl)
	if err != nil {
		return err
	}
	// move over the JSON-structured properties to the in-memory struct
	tl.LockTime = jtl.LockTime
	if jtl.Condition.Condition == nil {
		tl.Condition = &NilCondition{}
	} else {
		tl.Condition = jtl.Condition.Condition
	}
	return nil
}

// MarshalSia implements encoding.SiaMarshaler.MarshalSia
//
// Marshals this ConditionType as a single byte.
func (ct ConditionType) MarshalSia(w io.Writer) error {
	_, err := w.Write([]byte{byte(ct)})
	return err
}

// UnmarshalSia implements encoding.SiaUnmarshaler.UnmarshalSia
//
// Unmarshals this ConditionType from a single byte.
func (ct *ConditionType) UnmarshalSia(r io.Reader) error {
	var b [1]byte
	_, err := r.Read(b[:])
	*ct = ConditionType(b[0])
	return err
}

// MarshalSia implements encoding.Marshaler.MarshalSia
//
// Marshals this FulfillmentType as a single byte.
func (ft FulfillmentType) MarshalSia(w io.Writer) error {
	_, err := w.Write([]byte{byte(ft)})
	return err
}

// UnmarshalSia implements encoding.Unmarshaler.UnmarshalSia
//
// Unmarshals this FulfillmentType from a single byte.
func (ft *FulfillmentType) UnmarshalSia(r io.Reader) error {
	var b [1]byte
	_, err := r.Read(b[:])
	*ft = FulfillmentType(b[0])
	return err
}

// Fulfill implements UnlockFulfillment.Fulfill
//
// If no child is defined, an error will be returned,
// otherwise the child condition will be attempted to be fulfilled
// using the given fulfillment within the given (fulfill) context.
func (up UnlockConditionProxy) Fulfill(fulfillment UnlockFulfillment, ctx FulfillContext) error {
	condition := up.Condition
	if condition == nil {
		condition = &NilCondition{}
	}
	if p, ok := fulfillment.(UnlockFulfillmentProxy); ok {
		fulfillment = p.Fulfillment
		if fulfillment == nil {
			fulfillment = &NilFulfillment{}
		}
	}
	return condition.Fulfill(fulfillment, ctx)
}

// ConditionType implements UnlockCondition.ConditionType
//
// ConditionType returns the ConditionType of the child UnlockCondition.
// If no child is defined, a NilCondition is assumed.
func (up UnlockConditionProxy) ConditionType() ConditionType {
	condition := up.Condition
	if condition == nil {
		condition = &NilCondition{}
	}
	return condition.ConditionType()
}

// IsStandardCondition implements UnlockCondition.IsStandardCondition
//
// If no child is defined, nil will be returned,
// otherwise the question will be delegated to the child condition.
func (up UnlockConditionProxy) IsStandardCondition() error {
	condition := up.Condition
	if condition == nil {
		condition = &NilCondition{}
	}
	return condition.IsStandardCondition()
}

// UnlockHash implements UnlockCondition.UnlockHash
//
// If no child is defined, a nil hash will be returned,
// otherwise the child condition's unlock hash will be returned.
func (up UnlockConditionProxy) UnlockHash() UnlockHash {
	condition := up.Condition
	if condition == nil {
		condition = &NilCondition{}
	}
	return condition.UnlockHash()
}

// Equal implements UnlockCondition.Equal
//
// If no child is defined, the given UnlockCondition will be compared to the NilCondition,
// otherwise the child condition will be returned with the given UnlockCondition.
func (up UnlockConditionProxy) Equal(o UnlockCondition) bool {
	condition := up.Condition
	if condition == nil {
		condition = &NilCondition{}
	}
	if p, ok := o.(UnlockConditionProxy); ok {
		o = p.Condition
	}
	return condition.Equal(o)
}

// Fulfillable implements UnlockCondition.Fulfillable
func (up UnlockConditionProxy) Fulfillable(ctx FulfillableContext) bool {
	condition := up.Condition
	if condition == nil {
		condition = &NilCondition{}
	}
	return condition.Fulfillable(ctx)
}

// Sign implements UnlockFulfillment.Sign
//
// If no child is defined, an error will be returned,
// otherwise the child fulfillment will be signed within the given (fulfill sign) context.
func (fp UnlockFulfillmentProxy) Sign(ctx FulfillmentSignContext) error {
	fulfillment := fp.Fulfillment
	if fulfillment == nil {
		fulfillment = &NilFulfillment{}
	}
	return fulfillment.Sign(ctx)
}

// UnlockHash implements UnlockFulfillment.UnlockHash
//
// If no child is defined, the Nil UnlockHash will be returned,
// otherwise the child fulfillment's unlock hash will be returned.
func (fp UnlockFulfillmentProxy) UnlockHash() UnlockHash {
	fulfillment := fp.Fulfillment
	if fulfillment == nil {
		fulfillment = &NilFulfillment{}
	}
	return fulfillment.UnlockHash()
}

// FulfillmentType implements UnlockFulfillment.FulfillmentType
//
// If no child is defined, the Nil Fulfillment Type will be returned,
// otherwise the child fulfillment's type will be returned.
func (fp UnlockFulfillmentProxy) FulfillmentType() FulfillmentType {
	fulfillment := fp.Fulfillment
	if fulfillment == nil {
		fulfillment = &NilFulfillment{}
	}
	return fulfillment.FulfillmentType()
}

// IsStandardFulfillment implements UnlockFulfillment.IsStandardFulfillment
//
// If no child is defined, an error will be returned,
// otherwise the question will be delegated to the child fulfillmment.
func (fp UnlockFulfillmentProxy) IsStandardFulfillment() error {
	fulfillment := fp.Fulfillment
	if fulfillment == nil {
		fulfillment = &NilFulfillment{}
	}
	return fulfillment.IsStandardFulfillment()
}

// Equal implements UnlockFulfillment.Equal
//
// If no child is defined, the given unlock fulfillment will be compared to the NilFulfillment,
// otherwise the given fulfillment will be compared to the child fulfillment
func (fp UnlockFulfillmentProxy) Equal(f UnlockFulfillment) bool {
	fulfillment := fp.Fulfillment
	if fulfillment == nil {
		fulfillment = &NilFulfillment{}
	}
	if p, ok := f.(UnlockFulfillmentProxy); ok {
		f = p.Fulfillment
	}
	return fulfillment.Equal(f)
}

// MarshalSia implements encoding.SiaMarshaler.MarshalSia
//
// If no child is defined, the nil condition will be marshaled,
// otherwise the child condition will be marshaled using the
// MarshalableUnlockCondition.Marshal's method, appending the result
// after the binary-marshaled version of its type.
func (up UnlockConditionProxy) MarshalSia(w io.Writer) error {
	encoder := encoding.NewEncoder(w)
	if up.Condition == nil {
		return encoder.EncodeAll(ConditionTypeNil, 0) // type + nil-slice
	}
	return encoding.NewEncoder(w).EncodeAll(
		up.Condition.ConditionType(), up.Condition.Marshal())
}

// UnmarshalSia implements encoding.SiaUnmarshaler.UnmarshalSia
//
// First the ConditionType is unmarshaled, using that type,
// the correct UnlockCondition constructor is used to create
// an unlock condition instance, as to be able to (binary) unmarshal
// the child UnlockCondition.
//
// If the decoded type is unknown, the condition will not be attempted to be decoded,
// and instead the raw bytes will be kept in-memory as to be able to write it directly,
// when it is required to (binary) marshal this condition once again.
func (up *UnlockConditionProxy) UnmarshalSia(r io.Reader) error {
	var (
		t  ConditionType
		rc []byte
	)
	err := encoding.NewDecoder(r).DecodeAll(&t, &rc)
	if err != nil {
		return err
	}
	cc, ok := _RegisteredUnlockConditionTypes[t]
	if !ok {
		return ErrUnknownConditionType
	}
	c := cc()
	err = c.Unmarshal(rc)
	up.Condition = c
	return err
}

// MarshalSia implements encoding.SiaMarshaler.MarshalSia
//
// If no child is defined, the nil fulfillment will be marshaled,
// otherwise the child fulfillment will be marshaled using the
// MarshalableUnlockFulfillment.Marshal's method, appending the result
// after the binary-marshaled version of its type.
func (fp UnlockFulfillmentProxy) MarshalSia(w io.Writer) error {
	encoder := encoding.NewEncoder(w)
	if fp.Fulfillment == nil {
		return encoder.EncodeAll(FulfillmentTypeNil, 0) // type + nil-slice
	}
	return encoding.NewEncoder(w).EncodeAll(
		fp.Fulfillment.FulfillmentType(), fp.Fulfillment.Marshal())
}

// UnmarshalSia implements encoding.SiaUnmarshaler.UnmarshalSia
//
// First the FulfillmentType is unmarshaled, using that type,
// the correct UnlockFulfillment constructor is used to create
// an unlock fulfillment instance, as to be able to (binary) unmarshal
// the child UnlockFulfillment.
//
// If the decoded type is unknown, the fulfillment will not be attempted to be decoded,
// and instead the raw bytes will be kept in-memory as to be able to write it directly,
// when it is required to (binary) marshal this fulfillment once again.
func (fp *UnlockFulfillmentProxy) UnmarshalSia(r io.Reader) error {
	var (
		t  FulfillmentType
		rf []byte
	)
	err := encoding.NewDecoder(r).DecodeAll(&t, &rf)
	if err != nil {
		return err
	}
	fc, ok := _RegisteredUnlockFulfillmentTypes[t]
	if !ok {
		return ErrUnknownFulfillmentType
	}
	f := fc()
	err = f.Unmarshal(rf)
	fp.Fulfillment = f
	return err
}

var (
	_ encoding.SiaMarshaler   = UnlockConditionProxy{}
	_ encoding.SiaUnmarshaler = (*UnlockConditionProxy)(nil)

	_ encoding.SiaMarshaler   = UnlockFulfillmentProxy{}
	_ encoding.SiaUnmarshaler = (*UnlockFulfillmentProxy)(nil)
)

type (
	unlockConditionJSONFormat struct {
		Type ConditionType   `json:"type,omitempty"`
		Data json.RawMessage `json:"data,omitempty"`
	}
	unlockConditionJSONFormatWithNilData struct {
		Type ConditionType `json:"type,omitempty"`
	}
	unlockFulfillmentJSONFormat struct {
		Type FulfillmentType `json:"type,omitempty"`
		Data json.RawMessage `json:"data,omitempty"`
	}
	unlockFulfillmentJSONFormatWithNilData struct {
		Type FulfillmentType `json:"type,omitempty"`
	}
)

// MarshalJSON implements json.Marshaler.MarshalJSON
//
// If no child is defined, the nil condition will be marshaled,
// otherwise the child condition will be marshaled either implicitly
// or explicitly, depending on the child condition.
func (up UnlockConditionProxy) MarshalJSON() ([]byte, error) {
	if up.Condition == nil {
		return json.Marshal(unlockConditionJSONFormat{
			Type: ConditionTypeNil,
			Data: nil,
		})
	}
	data, err := json.Marshal(up.Condition)
	if err != nil {
		return nil, err
	}
	if string(data) == "{}" {
		return json.Marshal(unlockConditionJSONFormatWithNilData{
			Type: up.Condition.ConditionType(),
		})
	}
	return json.Marshal(unlockConditionJSONFormat{
		Type: up.Condition.ConditionType(),
		Data: data,
	})
}

// UnmarshalJSON implements json.Unmarshaler.UnmarshalJSON
//
// First the top-level condition structure is unmarshaled,
// resulting in the ConditionType property and a dynamic Data property.
// Using the now known ConditionType,
// the correct UnlockCondition constructor is used to create
// an unlock condition instance, as to be able to (binary) unmarshal
// the child UnlockCondition.
//
// If the unmarshaled ConditionType is unknown,
// an error will be returned.
func (up *UnlockConditionProxy) UnmarshalJSON(b []byte) error {
	var rf unlockConditionJSONFormat
	err := json.Unmarshal(b, &rf)
	if err != nil {
		return err
	}
	cc, ok := _RegisteredUnlockConditionTypes[rf.Type]
	if !ok {
		return ErrUnknownConditionType
	}
	c := cc()
	if rf.Data != nil {
		err = json.Unmarshal(rf.Data, &c)
	}
	up.Condition = c
	return err
}

// MarshalJSON implements json.Marshaler.MarshalJSON
//
// If no child is defined, the nil fulfillment will be marshaled,
// otherwise the child fulfillment will be marshaled either implicitly
// or explicitly, depending on the child fulfillment.
func (fp UnlockFulfillmentProxy) MarshalJSON() ([]byte, error) {
	if fp.Fulfillment == nil {
		return json.Marshal(unlockFulfillmentJSONFormat{
			Type: FulfillmentTypeNil,
			Data: nil,
		})
	}
	data, err := json.Marshal(fp.Fulfillment)
	if err != nil {
		return nil, err
	}
	if string(data) == "{}" {
		return json.Marshal(unlockFulfillmentJSONFormatWithNilData{
			Type: fp.Fulfillment.FulfillmentType(),
		})
	}
	return json.Marshal(unlockFulfillmentJSONFormat{
		Type: fp.Fulfillment.FulfillmentType(),
		Data: data,
	})
}

// UnmarshalJSON implements json.Unmarshaler.UnmarshalJSON
//
// First the top-level fulfillment structure is unmarshaled,
// resulting in the FulfillmentType property and a dynamic Data property.
// Using the now known FulfillmentType,
// the correct UnlockFulfillment constructor is used to create
// an unlock condition instance, as to be able to (binary) unmarshal
// the child UnlockFulfillment.
//
// If the unmarshaled FulfillmentType is unknown,
// an error will be returned.
func (fp *UnlockFulfillmentProxy) UnmarshalJSON(b []byte) error {
	var rf unlockFulfillmentJSONFormat
	err := json.Unmarshal(b, &rf)
	if err != nil {
		return err
	}
	fc, ok := _RegisteredUnlockFulfillmentTypes[rf.Type]
	if !ok {
		return ErrUnknownFulfillmentType
	}
	f := fc()
	if rf.Data != nil {
		err = json.Unmarshal(rf.Data, &f)
	}
	fp.Fulfillment = f
	return err
}

var (
	_ json.Marshaler   = UnlockConditionProxy{}
	_ json.Unmarshaler = (*UnlockConditionProxy)(nil)

	_ json.Marshaler   = UnlockFulfillmentProxy{}
	_ json.Unmarshaler = (*UnlockFulfillmentProxy)(nil)
)

var (
	_ UnlockCondition   = UnlockConditionProxy{}
	_ UnlockFulfillment = UnlockFulfillmentProxy{}
)

// strictSignatureCheck is used as part of the IsStandardFulfillment
// check of any Fulfillment which has a signature as part of its body.
// It ensures that the given public key and signature are a valid pair.
func strictSignatureCheck(pk SiaPublicKey, signature ByteSlice) error {
	switch pk.Algorithm {
	case SignatureEd25519:
		if len(pk.Key) != crypto.PublicKeySize {
			return errors.New("invalid public key size in transaction")
		}
		if len(signature) != crypto.SignatureSize {
			return errors.New("invalid signature size in transaction")
		}
		return nil
	default:
		return errors.New("unrecognized public key type in transaction")
	}
}

// signHashUsingSiaPublicKey produces a signature,
// for a given input, which is located within the given (parent) transaction,
// using the given (optional private) key, and using any extra objects (on top of the normal properties).
// The public key is to be given, as based on that the function can figure out what algorithm to use,
// and this also allows the function to know how to interpret the given (private) key.
func signHashUsingSiaPublicKey(pk SiaPublicKey, inputIndex uint64, tx Transaction, key interface{}, extraObjects ...interface{}) ([]byte, error) {
	switch pk.Algorithm {
	case SignatureEd25519:
		// decode the ed-secretKey
		var edSK crypto.SecretKey
		switch k := key.(type) {
		case crypto.SecretKey:
			edSK = k
		case ByteSlice:
			if len(k) != crypto.SecretKeySize {
				return nil, errors.New("invalid secret key size")
			}
			copy(edSK[:], k)
		case []byte:
			if len(k) != crypto.SecretKeySize {
				return nil, errors.New("invalid secret key size")
			}
			copy(edSK[:], k)
		default:
			return nil, fmt.Errorf("%T is an unknown secret key size", key)
		}
		sigHash := tx.InputSigHash(inputIndex, extraObjects...)
		sig := crypto.SignHash(sigHash, edSK)
		return sig[:], nil

	default:
		return nil, ErrUnknownSignAlgorithmType
	}
}

// verifyHashUsingSiaPublicKey verfies the given signature.
// It does so by:
//
// 1. producing the hash used to create the signature,
//    using the given inputIndex, (parent) transaction and any extra Objects to include
//    together with the normal transaction properties;
// 2. using the algorithm type of the given public key,
//    as to figure out what signature algorithm is used,
//    and thus being able to know how to verify the given signature;
func verifyHashUsingSiaPublicKey(pk SiaPublicKey, inputIndex uint64, tx Transaction, sig []byte, extraObjects ...interface{}) (err error) {
	switch pk.Algorithm {
	case SignatureEd25519:
		// Decode the public key and signature.
		var (
			edPK  crypto.PublicKey
			edSig crypto.Signature
		)
		copy(edPK[:], pk.Key)
		copy(edSig[:], sig)
		cryptoSig := crypto.Signature(edSig)
		sigHash := tx.InputSigHash(inputIndex, extraObjects...)
		err = crypto.VerifyHash(sigHash, edPK, cryptoSig)

	default:
		err = ErrUnknownSignAlgorithmType
	}
	return
}
