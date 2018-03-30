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

	"github.com/rivine/rivine/crypto"
	"github.com/rivine/rivine/encoding"
)

// Errors returned by input lock types.
var (
	ErrInvalidInputLockType  = errors.New("invalid input lock type")
	ErrUnlockConditionLocked = errors.New("unlock condition is already locked")

	// ErrUnknownUnlockType is an error returned in case
	// one tries to lock an input lock of unknown type
	ErrUnknownUnlockType = errors.New("unknown unlock type")

	// ErrUnknownSignAlgorithmType is an error returned in case
	// one tries to sign using an unknown signing algorithm type.
	//
	// NOTE That verification of unknown signing algorithm types does always succeed!
	ErrUnknownSignAlgorithmType = errors.New("unknown signature algorithm type")
)

// Errors related to atomic swaps
var (
	ErrInvalidPreImageSha256 = errors.New("invalid pre-image sha256")
	ErrInvalidRedeemer       = errors.New("invalid input redeemer")
)

const (
	// AtomicSwapSecretLen is the required/fixed length
	// of an atomic swap secret, the pre-image of an hashed secret.
	AtomicSwapSecretLen = sha256.Size
	// AtomicSwapHashedSecretLen is the required/fixed length
	// of an atomic swap hashed secret, the post-image of a secret.
	AtomicSwapHashedSecretLen = sha256.Size
)

type (
	// RawInputLockFormat defines the binary format of a condition-fullfilment pair,
	// used internally of an input lock.
	RawInputLockFormat struct {
		Condition   []byte
		Fulfillment []byte
	}

	// InputLock is a generic interface which hides the InputLock,
	// which is all serialized data used for generating a determenistic UnlockHash,
	// as well as the input used to unlock the input in the context of the
	// used InputLock, extra serialized input and Transaction it lives in.
	InputLock interface {
		// Lock the current Input within the context of
		// the transaction it exists in and its position in there.
		Lock(inputIndex uint64, tx Transaction, key interface{}) error

		// Unlock checks if the Input it lives in can be used (thus unlocked),
		// within the context of the transaction it lives in, the defined UnlockConditions
		// (defined by its UnlockHash) and the extra input parameters serializd as well.
		Unlock(inputIndex uint64, tx Transaction) error

		// StrictCheck ensures that all conditions and unlock input params,
		// are known and strict. This is useful as to make sure an input (and thus transaction)
		// can be understood by all nodes in the network.
		StrictCheck() error

		// EncodeCondition encodes the unlock conditon part of the InputLock,
		// which is the static part of an input lock, also used to generate the unlock hash,
		// and thus defined by the sender, and redefined by the receiver.
		EncodeCondition() []byte
		// EncodeFulfillment encodes the encode fulfillment part of the InputLock,
		// which is the dynamic part (the input parameters or signature as to speak) of an input lock.
		EncodeFulfillment() []byte

		// Decode the unlockCondition and fulfillment
		// from binary format into a known in-memory format.
		Decode(rf RawInputLockFormat) error
	}

	// InputLockProxy contains either no lock, or it does,
	// when it does it forwwards the functionality to the internal InputLock,
	// otherwise it acts as a nop-InputLock.
	//
	// InputLockProxy serves 2 important purposes:
	//   1. it provides a sane default for the InputLock interface,
	//      which will turn the input lock into a nop-lock, should the input not be defined.
	//   2. it makes it so that all inputlocks are serialized in the same way.
	//      this is important as it means we can ensure that an input's unlock hash,
	//      is the same, no matter if it's a known or unknown unlock type.
	//
	// The latter point (2) is very important,
	// as it is the one requirement that allows soft-forks to safely add new input locks,
	// without having to be afraid that their transactions will screw up older/non-forked nodes.
	InputLockProxy struct {
		t  UnlockType
		il InputLock
	}

	// InputLockConstructor is used to create a fresh internal input lock.
	InputLockConstructor func() InputLock

	// UnknownInputLock is used for all types which are unknown,
	// this allows soft-forks to define their own input lock types,
	// without breaking our code.
	//
	// Unknown types are always Locked and Unlocked,
	// but do not pass the Strict Check.
	UnknownInputLock struct {
		Condition, Fulfillment []byte
	}

	// SingleSignatureInputLock (0x01) is the only and most simplest unlocker.
	// It uses a public key (used as UnlockHash), such that only one public key is expected.
	// The spender will need to proof ownership of that public key by providing a correct signature.
	SingleSignatureInputLock struct {
		PublicKey SiaPublicKey
		Signature ByteSlice
	}
	// single-signature related
	// json util structs
	singleSignatureCondition struct {
		PublicKey SiaPublicKey `json:"publickey"`
	}
	singleSignatureFulfillment struct {
		Signature ByteSlice `json:"signature"`
	}

	// AtomicSwapSecret defines the 256 pre-image byte slice,
	// used as secret within the Atomic Swap protocol/contract.
	AtomicSwapSecret [sha256.Size]byte
	// AtomicSwapHashedSecret defines the 256 image byte slice,
	// used as hashed secret within the Atomic Swap protocol/contract.
	AtomicSwapHashedSecret [sha256.Size]byte
	// AtomicSwapRefundKey defines the refund key pair used by
	// the initiator/sender of such a contract.
	AtomicSwapRefundKey struct {
		PublicKey SiaPublicKey
		SecretKey ByteSlice
	}
	// AtomicSwapClaimKey defines the claim key pair used by
	// the participant/receiver of such a contract,
	// used to lock the contract from their side.
	AtomicSwapClaimKey struct {
		PublicKey SiaPublicKey
		SecretKey ByteSlice
		Secret    AtomicSwapSecret
	}

	// AtomicSwapInputLock (0x02) is a more advanced unlocker,
	// which allows for a more advanced InputLock,
	// where before the TimeLock expired, the output can only go to the receiver,
	// who has to give the secret in order to do so. After the InputLock,
	// the output can only be claimed by the sender, with no deadline in this phase.
	AtomicSwapInputLock struct {
		TimeLock         Timestamp
		Sender, Receiver UnlockHash
		HashedSecret     AtomicSwapHashedSecret
		PublicKey        SiaPublicKey
		Signature        ByteSlice
		Secret           AtomicSwapSecret
	}
	// AtomicSwapCondition defines the condition of an atomic swap contract/input-lock.
	// Only used for encoding purposes.
	AtomicSwapCondition struct {
		Sender       UnlockHash             `json:"sender"`
		Receiver     UnlockHash             `json:"receiver"`
		HashedSecret AtomicSwapHashedSecret `json:"hashedsecret"`
		TimeLock     Timestamp              `json:"timelock"`
	}
	// AtomicSwapFulfillment defines the fulfillment of an atomic swap contract/input-lock.
	// Only used for encoding purposes.
	AtomicSwapFulfillment struct {
		PublicKey SiaPublicKey     `json:"publickey"`
		Signature ByteSlice        `json:"signature"`
		Secret    AtomicSwapSecret `json:"secret"`
	}
)

// MarshalSia implements SiaMarshaler.MarshalSia
func (rf RawInputLockFormat) MarshalSia(w io.Writer) error {
	return encoding.NewEncoder(w).EncodeAll(rf.Condition, rf.Fulfillment)
}

// UnmarshalSia implements SiaUnmarshaler.UnmarshalSia
func (rf *RawInputLockFormat) UnmarshalSia(r io.Reader) error {
	return encoding.NewDecoder(r).DecodeAll(&rf.Condition, &rf.Fulfillment)
}

var (
	_ encoding.SiaMarshaler   = RawInputLockFormat{}
	_ encoding.SiaUnmarshaler = (*RawInputLockFormat)(nil)
)

// NewInputLockProxy creates a new input lock proxy,
// from a type and (existing) input lock.
func NewInputLockProxy(t UnlockType, il InputLock) InputLockProxy {
	if t != UnlockTypeNil && il == nil {
		panic("unexpected nil input lock")
	}
	return InputLockProxy{
		t:  t,
		il: il,
	}
}

// AtomicSwapInputLock casts this input lock proxy into a AtomicSwapInputLock if possible.
func (p InputLockProxy) AtomicSwapInputLock() (*AtomicSwapInputLock, bool) {
	if p.t != UnlockTypeAtomicSwap {
		return nil, false
	}
	il, ok := p.il.(*AtomicSwapInputLock)
	return il, ok
}

// MarshalSia implements SiaMarshaler.MarshalSia
func (p InputLockProxy) MarshalSia(w io.Writer) error {
	encoder := encoding.NewEncoder(w)
	err := encoder.Encode(p.t)
	if err != nil || p.t == UnlockTypeNil {
		return err
	}
	return encoder.Encode(RawInputLockFormat{
		Condition:   p.il.EncodeCondition(),
		Fulfillment: p.il.EncodeFulfillment(),
	})
}

// UnmarshalSia implements SiaMarshaler.UnmarshalSia
func (p *InputLockProxy) UnmarshalSia(r io.Reader) error {
	decoder := encoding.NewDecoder(r)
	err := decoder.Decode(&p.t)
	if err != nil || p.t == UnlockTypeNil {
		return err
	}
	var rf RawInputLockFormat
	err = decoder.Decode(&rf)
	if err != nil {
		return err
	}
	if c, found := _RegisteredUnlockTypes[p.t]; found {
		p.il = c()
	} else {
		p.il = new(UnknownInputLock)
	}
	return p.il.Decode(rf)
}

var (
	_ encoding.SiaMarshaler   = InputLockProxy{}
	_ encoding.SiaUnmarshaler = (*InputLockProxy)(nil)
)

// jsonInputLockProxy is a util struct
// in order to help us decode (and encode) input locks as JSON
// in a sane manner
//
// IMPORTANT: unknown types's condition and fulfillment have to be base64 byte slices!!!
type jsonInputLockProxy struct {
	Type        UnlockType      `json:"type,omitempty"`
	Condition   json.RawMessage `json:"condition,omitempty"`
	Fulfillment json.RawMessage `json:"fulfillment,omitempty"`
}

// MarshalJSON implements json.Marshaler.MarshalJSON
func (p InputLockProxy) MarshalJSON() ([]byte, error) {
	var (
		err          error
		rawInputLock = jsonInputLockProxy{Type: p.t}
	)

	switch p.t {
	case UnlockTypeNil:
		// nothing to do related to condition and signature

	case UnlockTypeSingleSignature:
		il := p.il.(*SingleSignatureInputLock)
		rawInputLock.Condition, err = json.Marshal(singleSignatureCondition{
			PublicKey: il.PublicKey,
		})
		if err != nil {
			return nil, err
		}
		rawInputLock.Fulfillment, err = json.Marshal(singleSignatureFulfillment{
			Signature: il.Signature,
		})
		if err != nil {
			return nil, err
		}

	case UnlockTypeAtomicSwap:
		il := p.il.(*AtomicSwapInputLock)
		rawInputLock.Condition, err = json.Marshal(AtomicSwapCondition{
			Sender:       il.Sender,
			Receiver:     il.Receiver,
			TimeLock:     il.TimeLock,
			HashedSecret: il.HashedSecret,
		})
		if err != nil {
			return nil, err
		}
		rawInputLock.Fulfillment, err = json.Marshal(AtomicSwapFulfillment{
			PublicKey: il.PublicKey,
			Signature: il.Signature,
			Secret:    il.Secret,
		})
		if err != nil {
			return nil, err
		}

	default:
		// forward-compatibility
		il := p.il.(*UnknownInputLock)
		rawInputLock.Condition, err = json.Marshal(il.Condition)
		if err != nil {
			return nil, err
		}
		rawInputLock.Fulfillment, err = json.Marshal(il.Fulfillment)
		if err != nil {
			return nil, err
		}

	}
	return json.Marshal(rawInputLock)
}

// UnmarshalJSON implements json.Unmarshaler.UnmarshalJSON
func (p *InputLockProxy) UnmarshalJSON(b []byte) error {
	var rawInputLock jsonInputLockProxy
	err := json.Unmarshal(b, &rawInputLock)
	if err != nil {
		return err
	}
	switch rawInputLock.Type {
	case UnlockTypeNil:
		p.il = nil

	case UnlockTypeSingleSignature:
		var (
			condition   singleSignatureCondition
			fulfillment singleSignatureFulfillment
		)
		err = json.Unmarshal(rawInputLock.Condition[:], &condition)
		if err != nil {
			return err
		}
		err = json.Unmarshal(rawInputLock.Fulfillment[:], &fulfillment)
		if err != nil {
			return err
		}
		p.il = &SingleSignatureInputLock{
			PublicKey: condition.PublicKey,
			Signature: fulfillment.Signature,
		}

	case UnlockTypeAtomicSwap:
		var (
			condition   AtomicSwapCondition
			fulfillment AtomicSwapFulfillment
		)
		err = json.Unmarshal(rawInputLock.Condition[:], &condition)
		if err != nil {
			return err
		}
		err = json.Unmarshal(rawInputLock.Fulfillment[:], &fulfillment)
		if err != nil {
			return err
		}
		p.il = &AtomicSwapInputLock{
			TimeLock:     condition.TimeLock,
			Sender:       condition.Sender,
			Receiver:     condition.Receiver,
			HashedSecret: condition.HashedSecret,
			PublicKey:    fulfillment.PublicKey,
			Signature:    fulfillment.Signature,
			Secret:       fulfillment.Secret,
		}

	default:
		var (
			condition, fulfillment []byte
		)
		err = json.Unmarshal(rawInputLock.Condition[:], &condition)
		if err != nil {
			return err
		}
		err = json.Unmarshal(rawInputLock.Fulfillment[:], &fulfillment)
		if err != nil {
			return err
		}
		p.il = &UnknownInputLock{
			Condition:   condition,
			Fulfillment: fulfillment,
		}
	}

	p.t = rawInputLock.Type
	return nil
}

var (
	_ json.Marshaler   = InputLockProxy{}
	_ json.Unmarshaler = (*InputLockProxy)(nil)
)

// UnlockHash implements InputLock.UnlockHash
func (p InputLockProxy) UnlockHash() UnlockHash {
	if p.t == UnlockTypeNil {
		return UnlockHash{}
	}
	return NewUnlockHash(p.t, crypto.HashObject(p.il.EncodeCondition()))
}

// Lock implements InputLock.Lock
func (p InputLockProxy) Lock(inputIndex uint64, tx Transaction, key interface{}) error {
	if p.t == UnlockTypeNil {
		return errors.New("nil input lock cannot be locked")
	}
	err := p.il.Lock(inputIndex, tx, key)
	if err != nil {
		return err
	}
	// validate the locking was done correctly
	return p.il.Unlock(inputIndex, tx)
}

// Unlock implements InputLock.Unlock
func (p InputLockProxy) Unlock(inputIndex uint64, tx Transaction) error {
	if p.t == UnlockTypeNil {
		return errors.New("nil input lock cannot be unlocked")
	}
	return p.il.Unlock(inputIndex, tx)
}

// StrictCheck implements InputLock.StrictCheck
func (p InputLockProxy) StrictCheck() error {
	if p.t == UnlockTypeNil {
		return errors.New("nil input lock")
	}
	return p.il.StrictCheck()
}

// EncodeCondition implements InputLock.EncodeCondition
func (u *UnknownInputLock) EncodeCondition() []byte {
	return u.Condition
}

// EncodeFulfillment implements InputLock.EncodeFulfillment
func (u *UnknownInputLock) EncodeFulfillment() []byte {
	return u.Fulfillment
}

// Decode implements InputLock.Decode
func (u *UnknownInputLock) Decode(rf RawInputLockFormat) error {
	u.Condition, u.Fulfillment = rf.Condition, rf.Fulfillment
	return nil
}

// Lock implements InputLock.Lock
func (u *UnknownInputLock) Lock(_ uint64, _ Transaction, _ interface{}) error {
	return ErrUnknownUnlockType // locking an unkown type is never valid
}

// Unlock implements InputLock.Unlock
func (u *UnknownInputLock) Unlock(_ uint64, _ Transaction) error {
	return nil // unlocking always passes for an unknown type
}

// StrictCheck implements InputLock.StrictCheck
func (u *UnknownInputLock) StrictCheck() error {
	return errors.New("unknown input lock")
}

var (
	_ InputLock = (*UnknownInputLock)(nil)
)

// NewSingleSignatureInputLock creates a new input lock,
// using the given public key and signature.
func NewSingleSignatureInputLock(pk SiaPublicKey) InputLockProxy {
	return NewInputLockProxy(UnlockTypeSingleSignature,
		&SingleSignatureInputLock{PublicKey: pk})
}

// MarshalSia implements SiaMarshaler.MarshalSia
func (ss *SingleSignatureInputLock) MarshalSia(w io.Writer) error {
	return encoding.NewEncoder(w).EncodeAll(ss.PublicKey, ss.Signature)
}

// UnmarshalSia implements SiaUnmarshaler.UnmarshalSia
func (ss *SingleSignatureInputLock) UnmarshalSia(r io.Reader) error {
	return encoding.NewDecoder(r).DecodeAll(&ss.PublicKey, &ss.Signature)
}

var (
	_ encoding.SiaMarshaler   = (*SingleSignatureInputLock)(nil)
	_ encoding.SiaUnmarshaler = (*SingleSignatureInputLock)(nil)
)

// EncodeCondition implements InputLock.EncodeCondition
func (ss *SingleSignatureInputLock) EncodeCondition() []byte {
	return encoding.Marshal(ss.PublicKey)
}

// EncodeFulfillment implements InputLock.EncodeFulfillment
func (ss *SingleSignatureInputLock) EncodeFulfillment() []byte {
	return ss.Signature
}

// Decode implements InputLock.Decode
func (ss *SingleSignatureInputLock) Decode(rf RawInputLockFormat) error {
	ss.Signature = rf.Fulfillment
	return encoding.Unmarshal(rf.Condition, &ss.PublicKey)
}

// Lock implements InputLock.Lock
func (ss *SingleSignatureInputLock) Lock(inputIndex uint64, tx Transaction, key interface{}) error {
	if len(ss.Signature) != 0 {
		return ErrUnlockConditionLocked
	}

	var err error
	ss.Signature, err = signHashUsingSiaPublicKey(ss.PublicKey, inputIndex, tx, key)
	return err
}

// Unlock implements InputLock.Unlock
func (ss *SingleSignatureInputLock) Unlock(inputIndex uint64, tx Transaction) error {
	return verifyHashUsingSiaPublicKey(ss.PublicKey, inputIndex, tx, ss.Signature)
}

// StrictCheck implements InputLock.StrictCheck
func (ss *SingleSignatureInputLock) StrictCheck() error {
	return strictSignatureCheck(ss.PublicKey, ss.Signature)
}

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

// NewAtomicSwapInputLock creates a new input lock as part of an atomic swap,
// using the given public keys, timelock and timestamp.
// Prior to the timestamp only the receiver can claim, using the required secret,
// after te deadline only the sender can claim a fund.
func NewAtomicSwapInputLock(condition AtomicSwapCondition) InputLockProxy {
	return NewInputLockProxy(UnlockTypeAtomicSwap,
		&AtomicSwapInputLock{
			TimeLock:     condition.TimeLock,
			Sender:       condition.Sender,
			Receiver:     condition.Receiver,
			HashedSecret: condition.HashedSecret,
		})
}

// Lock implements InputLock.Lock
func (as *AtomicSwapInputLock) Lock(inputIndex uint64, tx Transaction, key interface{}) error {
	if len(as.Signature) != 0 {
		return ErrUnlockConditionLocked
	}

	switch v := key.(type) {
	case AtomicSwapClaimKey: // claim
		if CurrentTimestamp() > as.TimeLock {
			return errors.New("atomic swap contract expired already")
		}

		as.Secret = v.Secret
		as.PublicKey = v.PublicKey
		hashedSecret := NewAtomicSwapHashedSecret(as.Secret)
		if bytes.Compare(as.HashedSecret[:], hashedSecret[:]) != 0 {
			return ErrInvalidPreImageSha256
		}

		var err error
		as.Signature, err = signHashUsingSiaPublicKey(v.PublicKey, inputIndex, tx, v.SecretKey, as.PublicKey, as.Secret)
		return err

	case AtomicSwapRefundKey: // refund
		if CurrentTimestamp() <= as.TimeLock {
			return errors.New("atomic swap contract not yet expired")
		}
		as.PublicKey = v.PublicKey

		var err error
		as.Signature, err = signHashUsingSiaPublicKey(v.PublicKey, inputIndex, tx, v.SecretKey, as.PublicKey)
		return err

	default:
		return fmt.Errorf("cannot atomic-swap-lock using %T key", key)
	}
}

// Unlock implements InputLock.Unlock
func (as *AtomicSwapInputLock) Unlock(inputIndex uint64, tx Transaction) error {
	// create the unlockHash for the given public Key
	unlockHash := NewSingleSignatureInputLock(as.PublicKey).UnlockHash()

	// prior to our timelock, only the receiver can claim the unspend output
	if CurrentTimestamp() <= as.TimeLock {
		// verify that receiver public key was given
		if unlockHash.Type != as.Receiver.Type ||
			bytes.Compare(unlockHash.Hash[:], as.Receiver.Hash[:]) != 0 {
			return ErrInvalidRedeemer
		}

		// verify signature
		err := verifyHashUsingSiaPublicKey(as.PublicKey, inputIndex, tx, as.Signature, as.PublicKey, as.Secret)
		if err != nil {
			return err
		}

		// in order for the receiver to spend,
		// the secret has to be known
		hashedSecret := NewAtomicSwapHashedSecret(as.Secret)
		if bytes.Compare(as.HashedSecret[:], hashedSecret[:]) != 0 {
			return ErrInvalidPreImageSha256
		}

		return nil
	}

	// verify that sender public key was given
	if unlockHash.Type != as.Sender.Type ||
		bytes.Compare(unlockHash.Hash[:], as.Sender.Hash[:]) != 0 {
		return ErrInvalidRedeemer
	}

	// after the deadline (timelock),
	// only the original sender can reclaim the unspend output
	return verifyHashUsingSiaPublicKey(as.PublicKey, inputIndex, tx, as.Signature, as.PublicKey)
}

// EncodeCondition implements InputLock.EncodeCondition
func (as *AtomicSwapInputLock) EncodeCondition() []byte {
	return encoding.Marshal(AtomicSwapCondition{
		Sender:       as.Sender,
		Receiver:     as.Receiver,
		TimeLock:     as.TimeLock,
		HashedSecret: as.HashedSecret,
	})
}

// EncodeFulfillment implements InputLock.EncodeFulfillment
func (as *AtomicSwapInputLock) EncodeFulfillment() []byte {
	return encoding.Marshal(AtomicSwapFulfillment{
		PublicKey: as.PublicKey,
		Signature: as.Signature,
		Secret:    as.Secret,
	})
}

// Decode implements InputLock.Decode
func (as *AtomicSwapInputLock) Decode(rf RawInputLockFormat) error {
	var condition AtomicSwapCondition
	err := encoding.Unmarshal(rf.Condition, &condition)
	if err != nil {
		return err
	}
	as.TimeLock = condition.TimeLock
	as.Sender = condition.Sender
	as.Receiver = condition.Receiver
	as.HashedSecret = condition.HashedSecret

	var fulfillment AtomicSwapFulfillment
	err = encoding.Unmarshal(rf.Fulfillment, &fulfillment)
	if err != nil {
		return err
	}
	as.PublicKey = fulfillment.PublicKey
	as.Signature = fulfillment.Signature
	as.Secret = fulfillment.Secret
	return nil
}

// StrictCheck implements InputLock.StrictCheck
func (as *AtomicSwapInputLock) StrictCheck() error {
	return strictSignatureCheck(as.PublicKey, as.Signature)
}

func strictSignatureCheck(pk SiaPublicKey, signature []byte) error {
	switch pk.Algorithm {
	case SignatureEntropy:
		return nil
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

func signHashUsingSiaPublicKey(pk SiaPublicKey, inputIndex uint64, tx Transaction, key interface{}, extraObjects ...interface{}) ([]byte, error) {
	switch pk.Algorithm {
	case SignatureEntropy:
		// Entropy cannot ever be used to sign a transaction.
		return nil, ErrEntropyKey

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

func verifyHashUsingSiaPublicKey(pk SiaPublicKey, inputIndex uint64, tx Transaction, sig []byte, extraObjects ...interface{}) (err error) {
	switch pk.Algorithm {
	case SignatureEntropy:
		// Entropy cannot ever be used to sign a transaction.
		err = ErrEntropyKey

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
		// If the identifier is not recognized, assume that the signature
		// is valid. This allows more signature types to be added via soft
		// forking.
	}

	return
}

// _RegisteredInputLocks contains all known/registered unlockers constructors.
var _RegisteredUnlockTypes = map[UnlockType]InputLockConstructor{}

// RegisterUnlockType registers the given non-nil locker constructor,
// for the given UnlockType, essentially linking the given locker constructor to the given type.
func RegisterUnlockType(t UnlockType, ilc InputLockConstructor) {
	if ilc == nil {
		panic("cannot register nil input locker")
	}
	_RegisteredUnlockTypes[t] = ilc
}

// UnregisterUnlockType unregisters the given UnlockType,
// meaning the given UnlockType will no longer have a matching unlocker constructor.
func UnregisterUnlockType(t UnlockType) {
	delete(_RegisteredUnlockTypes, t)
}

func init() {
	// standard non-nil input locks
	RegisterUnlockType(UnlockTypeSingleSignature, func() InputLock {
		return new(SingleSignatureInputLock)
	})
	RegisterUnlockType(UnlockTypeAtomicSwap, func() InputLock {
		return new(AtomicSwapInputLock)
	})
}
