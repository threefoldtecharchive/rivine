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

	// InputLock is a generic interface which hides the InputLock,
	// which is all serialized data used for generating a determenistic UnlockHash,
	// as well as the input used to unlock the input in the context of the
	// used InputLock, extra serialized input and Transaction it lives in.
	InputLock interface {
		// UnlockHash generates a deterministic UnlockHash,
		// and should be related only to the static unlock conditions,
		// without any possible influence of other input parameters and other conditions.
		UnlockHash() UnlockHash

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
	}

	// InputLockProxy contains either no lock, or it does,
	// when it does it forwwards the functionality to the internal InputLock,
	// otherwise it acts as a nop-InputLock
	InputLockProxy struct {
		t  InputLockType
		il InputLock
	}

	// UnlockerConstructor is used to create a fresh unlocker.
	UnlockerConstructor func() InputLock

	// SingleSignatureInputLock (0x01) is the only and most simplest unlocker.
	// It uses a public key (used as UnlockHash), such that only one public key is expected.
	// The spender will need to proof ownership of that public key by providing a correct signature.
	SingleSignatureInputLock struct {
		PublicKey SiaPublicKey
		Signature []byte
	}

	AtomicSwapSecret       [sha256.Size]byte
	AtomicSwapHashedSecret [sha256.Size]byte
	AtomicSwapClaimKey     struct {
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
	err := encoding.NewEncoder(w).Encode(p.t)
	if err != nil || p.t == InputLockTypeNil {
		return err
	}
	return encoding.NewEncoder(w).Encode(p.il)
}

// UnmarshalSia implements SiaMarshaler.UnmarshalSia
func (p InputLockProxy) UnmarshalSia(r io.Reader) error {
	err := encoding.NewDecoder(r).Decode(&p.t)
	if err != nil || p.t == InputLockTypeNil {
		return err
	}

	c, found := _RegisteredInputLocks[p.t]
	if !found {
		return ErrInvalidInputLockType
	}
	p.il = c()
	return encoding.NewDecoder(r).Decode(p.il)
}

// UnlockHash implements InputLock.UnlockHash
func (p InputLockProxy) UnlockHash() UnlockHash {
	if p.t == InputLockTypeNil {
		return UnlockHash{}
	}
	return p.il.UnlockHash()
}

// Lock implements InputLock.Lock
func (p InputLockProxy) Lock(inputIndex uint64, tx Transaction, key interface{}) error {
	if p.t == InputLockTypeNil {
		return nil
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
		return nil
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

// NewSingleSignatureInputLock creates a new input lock,
// using the given public key and signature.
func NewSingleSignatureInputLock(pk SiaPublicKey) InputLockProxy {
	return NewInputLockProxy(InputLockTypeSingleSignature,
		&SingleSignatureInputLock{PublicKey: pk})
}

// UnlockHash implements InputLock.UnlockHash
func (ss *SingleSignatureInputLock) UnlockHash() UnlockHash {
	return UnlockHash(crypto.HashObject(ss.PublicKey))
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
			PublicKeySender:   s,
			PublicKeyReceiver: r,
			HashedSecret:      hs,
			Timelock:          tl,
		})
}

// UnlockHash implements InputLock.UnlockHash
func (as *AtomicSwapInputLock) UnlockHash() UnlockHash {
	return UnlockHash(crypto.HashAll(
		as.Timelock, as.PublicKeyReceiver,
		as.PublicKeySender, as.HashedSecret))
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

		// TODO: integrate secret in sigHash
		var err error
		as.Signature, err = signHashUsingSiaPublicKey(as.PublicKeyReceiver, inputIndex, tx, v.SecretKey)
		return err

	case crypto.SecretKey: // refund
		if CurrentTimestamp() <= as.Timelock {
			return errors.New("atomic swap contract not yet expired")
		}

		// TODO: integrate secret in sigHash
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
		// TODO: integrate secret in sigHash
		err := verifyHashUsingSiaPublicKey(as.PublicKeyReceiver, inputIndex, tx, as.Signature)
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
	// TODO: integrate secret in sigHash
	return verifyHashUsingSiaPublicKey(as.PublicKeySender, inputIndex, tx, as.Signature)
}

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

func signHashUsingSiaPublicKey(pk SiaPublicKey, inputIndex uint64, tx Transaction, key interface{}) ([]byte, error) {
	switch pk.Algorithm {
	case SignatureEntropy:
		// Entropy cannot ever be used to sign a transaction.
		return nil, ErrEntropyKey

	case SignatureEd25519:
		sigHash := tx.InputSigHash(inputIndex)
		sig := crypto.SignHash(sigHash, key.(crypto.SecretKey))
		return sig[:], nil

	default:
		// If the identifier is not recognized, assume that the signature
		// is valid. This allows more signature types to be added via soft
		// forking.
	}

	return nil, nil
}

func verifyHashUsingSiaPublicKey(pk SiaPublicKey, inputIndex uint64, tx Transaction, sig []byte) error {
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
		sigHash := tx.InputSigHash(inputIndex)
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
var _RegisteredInputLocks = map[InputLockType]UnlockerConstructor{}

// RegisterInputLockType registers the given non-nil locker constructor,
// for the given InputLockType, essentially linking the given locker constructor to the given type.
func RegisterInputLockType(t InputLockType, uc UnlockerConstructor) {
	if uc == nil {
		panic("cannot register nil input locker")
	}
	_RegisteredInputLocks[t] = uc
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
