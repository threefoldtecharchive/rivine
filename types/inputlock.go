package types

import (
	"bytes"
	"crypto/sha256"
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
)

// Errors related to atomic swaps
var (
	ErrInvalidPreImageSha256 = errors.New("invalid pre-image sha256")
)

type (
	// InputLockType defines the type of an unlock condition-fulfillment pair.
	InputLockType byte

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
		t  InputLockType
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
		Signature []byte
	}

	// AtomicSwapSecret defines the 256 pre-image byte slice,
	// used as secret within the Atomic Swap protocol/contract.
	AtomicSwapSecret [sha256.Size]byte
	// AtomicSwapHashedSecret defines the 256 image byte slice,
	// used as hashed secret within the Atomic Swap protocol/contract.
	AtomicSwapHashedSecret [sha256.Size]byte
	// AtomicSwapClaimKey defines the claim (private) key used by
	// the participant/receiver of such a contract,
	// used to lock the contract from their side.
	AtomicSwapClaimKey struct {
		Secret    AtomicSwapSecret
		SecretKey crypto.SecretKey
	}

	// AtomicSwapInputLock (0x02) is a more advanced unlocker,
	// which allows for a more advanced InputLock,
	// where before the TimeLock expired, the output can only go to the receiver,
	// who has to give the secret in order to do so. After the InputLock,
	// the output can only be claimed by the sender, with no deadline in this phase.
	AtomicSwapInputLock struct {
		Timelock                           Timestamp
		PublicKeySender, PublicKeyReceiver SiaPublicKey
		HashedSecret                       AtomicSwapHashedSecret
		Secret                             AtomicSwapSecret
		Signature                          []byte
	}
	// AtomicSwapCondition defines the condition of an atomic swap contract/input-lock.
	// Only used for encoding purposes.
	AtomicSwapCondition struct {
		PublicKeySender, PublicKeyReceiver SiaPublicKey
		HashedSecret                       AtomicSwapHashedSecret
		Timelock                           Timestamp
	}
	// AtomicSwapFulfillment defines the fulfillment of an atomic swap contract/input-lock.
	// Only used for encoding purposes.
	AtomicSwapFulfillment struct {
		Secret    AtomicSwapSecret
		Signature []byte
	}
)

const (
	// InputLockTypeNil defines a nil (empty) Input Lock and is the default.
	InputLockTypeNil InputLockType = iota

	// InputLockTypeSingleSignature provides the standard and most simple unlock type.
	// In it the sender gives the public key of the intendend receiver.
	// The receiver can redeem the relevant locked input by providing a signature
	// which proofs the ownership of the private key linked to the known public key.
	InputLockTypeSingleSignature

	// InputLockTypeAtomicSwap provides a more advanced unlocker,
	// which allows for a more advanced InputLock,
	// where before the TimeLock expired, the output can only go to the receiver,
	// who has to give the secret in order to do so. After the InputLock,
	// the output can only be claimed by the sender, with no deadline in this phas
	InputLockTypeAtomicSwap

	// MaxStandardInputLockType can be used to define your own
	// InputLockType without having to hardcode the final standard
	// unlock type, while still preventing any possible type overwrite.
	MaxStandardInputLockType = InputLockTypeSingleSignature
)

// MarshalSia implements SiaMarshaler.MarshalSia
func (t InputLockType) MarshalSia(w io.Writer) error {
	_, err := w.Write([]byte{byte(t)})
	return err
}

// UnmarshalSia implements SiaUnmarshaler.UnmarshalSia
func (t *InputLockType) UnmarshalSia(r io.Reader) error {
	var bt [1]byte
	_, err := io.ReadFull(r, bt[:])
	*t = InputLockType(bt[0])
	return err
}

// MarshalSia implements SiaMarshaler.MarshalSia
func (rf RawInputLockFormat) MarshalSia(w io.Writer) error {
	return encoding.NewEncoder(w).EncodeAll(rf.Condition, rf.Fulfillment)
}

// UnmarshalSia implements SiaUnmarshaler.UnmarshalSia
func (rf *RawInputLockFormat) UnmarshalSia(r io.Reader) error {
	return encoding.NewDecoder(r).DecodeAll(&rf.Condition, &rf.Fulfillment)
}

var (
	_ encoding.SiaMarshaler   = InputLockType(0)
	_ encoding.SiaMarshaler   = RawInputLockFormat{}
	_ encoding.SiaUnmarshaler = (*InputLockType)(nil)
	_ encoding.SiaUnmarshaler = (*RawInputLockFormat)(nil)
)

// NewInputLockProxy creates a new input lock proxy,
// from a type and (existing) input lock.
func NewInputLockProxy(t InputLockType, il InputLock) InputLockProxy {
	if t != InputLockTypeNil && il == nil {
		panic("unexpected nil input lock")
	}
	return InputLockProxy{
		t:  t,
		il: il,
	}
}

// MarshalSia implements SiaMarshaler.MarshalSia
func (p InputLockProxy) MarshalSia(w io.Writer) error {
	encoder := encoding.NewEncoder(w)
	err := encoder.Encode(p.t)
	if err != nil || p.t == InputLockTypeNil {
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
	if err != nil || p.t == InputLockTypeNil {
		return err
	}
	var rf RawInputLockFormat
	err = decoder.Decode(&rf)
	if err != nil {
		return err
	}
	if c, found := _RegisteredInputLocks[p.t]; found {
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

// UnlockHash implements InputLock.UnlockHash
func (p InputLockProxy) UnlockHash() UnlockHash {
	if p.t == InputLockTypeNil {
		return UnlockHash{}
	}
	return UnlockHash(crypto.HashObject(p.il.EncodeCondition()))
}

// Lock implements InputLock.Lock
func (p InputLockProxy) Lock(inputIndex uint64, tx Transaction, key interface{}) error {
	if p.t == InputLockTypeNil {
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
	if p.t == InputLockTypeNil {
		return errors.New("nil input lock cannot be unlocked")
	}
	return p.il.Unlock(inputIndex, tx)
}

// StrictCheck implements InputLock.StrictCheck
func (p InputLockProxy) StrictCheck() error {
	if p.t == InputLockTypeNil {
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
	return nil // locking does nothing for an unknown type
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
	return NewInputLockProxy(InputLockTypeSingleSignature,
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

// NewAtomicSwapInputLock creates a new input lock as part of an atomic swap,
// using the given public keys, timelock and timestamp.
// Prior to the timestamp only the receiver can claim, using the required secret,
// after te deadline only the sender can claim a fund.
func NewAtomicSwapInputLock(s, r SiaPublicKey, hs AtomicSwapHashedSecret, tl Timestamp) InputLockProxy {
	return NewInputLockProxy(InputLockTypeAtomicSwap,
		&AtomicSwapInputLock{
			Timelock:          tl,
			PublicKeySender:   s,
			PublicKeyReceiver: r,
			HashedSecret:      hs,
		})
}

// Lock implements InputLock.Lock
func (as *AtomicSwapInputLock) Lock(inputIndex uint64, tx Transaction, key interface{}) error {
	if len(as.Signature) != 0 {
		return ErrUnlockConditionLocked
	}

	switch v := key.(type) {
	case AtomicSwapClaimKey: // claim
		if CurrentTimestamp() > as.Timelock {
			return errors.New("atomic swap contract expired")
		}

		as.Secret = v.Secret
		hashedSecret := sha256.Sum256(as.Secret[:])
		if bytes.Compare(as.HashedSecret[:], hashedSecret[:]) != 0 {
			return ErrInvalidPreImageSha256
		}

		var err error
		as.Signature, err = signHashUsingSiaPublicKey(as.PublicKeyReceiver, inputIndex, tx, v.SecretKey, as.Secret)
		return err

	case crypto.SecretKey: // refund
		if CurrentTimestamp() <= as.Timelock {
			return errors.New("atomic swap contract not yet expired")
		}

		var err error
		as.Signature, err = signHashUsingSiaPublicKey(as.PublicKeyReceiver, inputIndex, tx, v)
		return err

	default:
		return fmt.Errorf("cannot atomic-swap-lock using %T key", key)
	}
}

// Unlock implements InputLock.Unlock
func (as *AtomicSwapInputLock) Unlock(inputIndex uint64, tx Transaction) error {
	// prior to our timelock, only the receiver can claim the unspend output
	if CurrentTimestamp() <= as.Timelock {
		err := verifyHashUsingSiaPublicKey(as.PublicKeyReceiver, inputIndex, tx, as.Signature, as.Secret)
		if err != nil {
			return err
		}

		// in order for the receiver to spend,
		// the secret has to be known
		hashedSecret := sha256.Sum256(as.Secret[:])
		if bytes.Compare(as.HashedSecret[:], hashedSecret[:]) != 0 {
			return ErrInvalidPreImageSha256
		}

		return nil
	}

	// after the deadline (timelock),
	// only the original sender can reclaim the unspend output
	return verifyHashUsingSiaPublicKey(as.PublicKeySender, inputIndex, tx, as.Signature)
}

// EncodeCondition implements InputLock.EncodeCondition
func (as *AtomicSwapInputLock) EncodeCondition() []byte {
	return encoding.Marshal(AtomicSwapCondition{
		PublicKeySender:   as.PublicKeySender,
		PublicKeyReceiver: as.PublicKeyReceiver,
		Timelock:          as.Timelock,
		HashedSecret:      as.HashedSecret,
	})
}

// EncodeFulfillment implements InputLock.EncodeFulfillment
func (as *AtomicSwapInputLock) EncodeFulfillment() []byte {
	return encoding.Marshal(AtomicSwapFulfillment{
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
	as.Timelock = condition.Timelock
	as.PublicKeySender = condition.PublicKeySender
	as.PublicKeyReceiver = condition.PublicKeyReceiver
	as.HashedSecret = condition.HashedSecret

	var fulfillment AtomicSwapFulfillment
	err = encoding.Unmarshal(rf.Fulfillment, &fulfillment)
	if err != nil {
		return err
	}
	as.Signature = fulfillment.Signature
	as.Secret = fulfillment.Secret
	return nil
}

// MarshalSia implements SiaMarshaler.MarshalSia
func (ac *AtomicSwapCondition) MarshalSia(w io.Writer) error {
	return encoding.NewEncoder(w).EncodeAll(
		ac.PublicKeySender, ac.PublicKeyReceiver,
		ac.HashedSecret, ac.Timelock)
}

// UnmarshalSia implements SiaUnmarshaler.UnmarshalSia
func (ac *AtomicSwapCondition) UnmarshalSia(r io.Reader) error {
	return encoding.NewDecoder(r).DecodeAll(
		&ac.PublicKeySender, &ac.PublicKeyReceiver,
		&ac.HashedSecret, &ac.Timelock)
}

// MarshalSia implements SiaMarshaler.MarshalSia
func (af *AtomicSwapFulfillment) MarshalSia(w io.Writer) error {
	return encoding.NewEncoder(w).EncodeAll(
		af.Secret, af.Signature)
}

// UnmarshalSia implements SiaUnmarshaler.UnmarshalSia
func (af *AtomicSwapFulfillment) UnmarshalSia(r io.Reader) error {
	return encoding.NewDecoder(r).DecodeAll(
		&af.Secret, &af.Signature)
}

var (
	_ encoding.SiaMarshaler   = (*AtomicSwapCondition)(nil)
	_ encoding.SiaMarshaler   = (*AtomicSwapFulfillment)(nil)
	_ encoding.SiaUnmarshaler = (*AtomicSwapCondition)(nil)
	_ encoding.SiaUnmarshaler = (*AtomicSwapFulfillment)(nil)
)

// StrictCheck implements InputLock.StrictCheck
func (as *AtomicSwapInputLock) StrictCheck() error {
	// prior to our timelock, only the receiver can claim the unspend output
	if CurrentTimestamp() <= as.Timelock {
		return strictSignatureCheck(as.PublicKeyReceiver, as.Signature)
	}
	// after the deadline (timelock),
	// only the original sender can reclaim the unspend output
	return strictSignatureCheck(as.PublicKeySender, as.Signature)
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
		sigHash := tx.InputSigHash(inputIndex, extraObjects...)
		sig := crypto.SignHash(sigHash, key.(crypto.SecretKey))
		return sig[:], nil

	default:
		// If the identifier is not recognized, assume that the signature
		// is valid. This allows more signature types to be added via soft
		// forking.
	}

	return nil, nil
}

func verifyHashUsingSiaPublicKey(pk SiaPublicKey, inputIndex uint64, tx Transaction, sig []byte, extraObjects ...interface{}) error {
	switch pk.Algorithm {
	case SignatureEntropy:
		// Entropy cannot ever be used to sign a transaction.
		return ErrEntropyKey

	case SignatureEd25519:
		// Decode the public key and signature.
		var edPK crypto.PublicKey
		err := encoding.Unmarshal([]byte(pk.Key), &edPK)
		if err != nil {
			return err
		}
		var edSig [crypto.SignatureSize]byte
		err = encoding.Unmarshal(sig, &edSig)
		if err != nil {
			return err
		}
		cryptoSig := crypto.Signature(edSig)
		sigHash := tx.InputSigHash(inputIndex, extraObjects...)
		err = crypto.VerifyHash(sigHash, edPK, cryptoSig)
		if err != nil {
			return err
		}

	default:
		// If the identifier is not recognized, assume that the signature
		// is valid. This allows more signature types to be added via soft
		// forking.
	}

	return nil
}

// _RegisteredInputLocks contains all known/registered unlockers constructors.
var _RegisteredInputLocks = map[InputLockType]InputLockConstructor{}

// RegisterInputLockType registers the given non-nil locker constructor,
// for the given InputLockType, essentially linking the given locker constructor to the given type.
func RegisterInputLockType(t InputLockType, ilc InputLockConstructor) {
	if ilc == nil {
		panic("cannot register nil input locker")
	}
	_RegisteredInputLocks[t] = ilc
}

// UnregisterInputLockType unregisters the given InputLockType,
// meaning the given InputLockType will no longer have a matching unlocker constructor.
func UnregisterInputLockType(t InputLockType) {
	delete(_RegisteredInputLocks, t)
}

func init() {
	// standard non-nil input locks
	RegisterInputLockType(InputLockTypeSingleSignature, func() InputLock {
		return new(SingleSignatureInputLock)
	})
	RegisterInputLockType(InputLockTypeAtomicSwap, func() InputLock {
		return new(AtomicSwapInputLock)
	})
}
