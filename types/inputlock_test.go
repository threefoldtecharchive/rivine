package types

import (
	"bytes"
	"crypto/sha256"
	"io"
	"reflect"
	"testing"

	"github.com/rivine/rivine/crypto"
	"github.com/rivine/rivine/encoding"
)

func TestSingleSignatureUnlocker(t *testing.T) {
	sk, pk := crypto.GenerateKeyPair()
	ul := NewSingleSignatureInputLock(Ed25519PublicKey(pk))

	err := ul.StrictCheck()
	if err == nil {
		t.Error("error was expected, nil received")
	}

	uh1 := ul.UnlockHash()
	uh2 := ul.UnlockHash()
	if uh1.String() != uh2.String() {
		t.Error("inconsistent unlock hashes:", uh1, uh2)
	}

	err = ul.Unlock(0, Transaction{})
	if err == nil {
		t.Error("error was expected, nil received")
	}

	err = ul.Lock(0, Transaction{}, sk)
	if err != nil {
		t.Errorf("failed to lock transaction: %v", err)
	}

	err = ul.StrictCheck()
	if err != nil {
		t.Errorf("strict check failed while it was expected to succeed: %v", err)
	}
	err = ul.Unlock(0, Transaction{})
	if err != nil {
		t.Errorf("unlock failed while it was expected to succeed: %v", err)
	}

	uh3 := ul.UnlockHash()
	if uh1.String() != uh3.String() {
		t.Error("inconsistent unlock hashes:", uh1, uh3)
	}
}

func TestSingleSignatureUnlockerBadTransaction(t *testing.T) {
	sk, pk := crypto.GenerateKeyPair()
	ul := NewSingleSignatureInputLock(Ed25519PublicKey(pk))

	tx := Transaction{}

	err := ul.Lock(0, tx, sk)
	if err != nil {
		t.Errorf("failed to lock transaction: %v", err)
	}
	err = ul.Unlock(0, tx)
	if err != nil {
		t.Errorf("unlock failed while it was expected to succeed: %v", err)
	}

	tx.CoinInputs = append(tx.CoinInputs, CoinInput{
		Unlocker: NewSingleSignatureInputLock(Ed25519PublicKey(pk)),
	})
	ul.il.(*SingleSignatureInputLock).Signature = nil
	err = ul.Lock(0, tx, sk)
	if err != nil {
		t.Errorf("failed to lock transaction: %v", err)
	}
	err = ul.Unlock(0, tx)
	if err != nil {
		t.Errorf("unlock failed while it was expected to succeed: %v", err)
	}
	err = ul.Unlock(0, Transaction{})
	if err == nil {
		t.Errorf("unlock should fail as transaction is wrong")
	}

	tx.CoinInputs = append(tx.CoinInputs, tx.CoinInputs[0])
	ul.il.(*SingleSignatureInputLock).Signature = nil
	err = ul.Lock(0, tx, sk)
	if err != nil {
		t.Errorf("failed to lock transaction: %v", err)
	}
	err = ul.Unlock(0, tx)
	if err != nil {
		t.Errorf("unlock failed while it was expected to succeed: %v", err)
	}
}

func TestAtomicSwapUnlocker(t *testing.T) {
	sk1, pk1 := crypto.GenerateKeyPair()
	sk2, pk2 := crypto.GenerateKeyPair()

	secret := sha256.Sum256([]byte{1, 2, 3, 4})
	hashedSecret := sha256.Sum256(secret[:])

	ul := NewAtomicSwapInputLock(Ed25519PublicKey(pk1), Ed25519PublicKey(pk2),
		AtomicSwapHashedSecret(hashedSecret), CurrentTimestamp()+50000)

	err := ul.StrictCheck()
	if err == nil {
		t.Error("error was expected, nil received")
	}

	uh1 := ul.UnlockHash()
	uh2 := ul.UnlockHash()
	if uh1.String() != uh2.String() {
		t.Error("inconsistent unlock hashes:", uh1, uh2)
	}

	err = ul.Unlock(0, Transaction{})
	if err == nil {
		t.Error("error was expected, nil received")
	}

	err = ul.Lock(0, Transaction{}, sk2)
	if err == nil {
		t.Errorf("should fail, as reclaim cannot be done, and especially not by sk2")
	}
	ul.il.(*AtomicSwapInputLock).Signature = nil
	err = ul.Lock(0, Transaction{}, sk1)
	if err == nil {
		t.Errorf("should fail, as reclaim cannot be done, and especially not by sk1")
	}
	ul.il.(*AtomicSwapInputLock).Signature = nil
	err = ul.Lock(0, Transaction{}, AtomicSwapClaimKey{
		SecretKey: sk1,
		Secret:    AtomicSwapSecret{},
	})
	if err == nil {
		t.Errorf("should fail, as reclaim cannot be done, and especially not by sk1")
	}
	ul.il.(*AtomicSwapInputLock).Signature = nil
	err = ul.Lock(0, Transaction{}, AtomicSwapClaimKey{
		SecretKey: sk2,
		Secret:    AtomicSwapSecret{},
	})
	if err == nil {
		t.Errorf("should fail, as secret is missing/wrong")
	}
	err = ul.Lock(0, Transaction{}, AtomicSwapClaimKey{
		SecretKey: sk2,
		Secret:    AtomicSwapSecret{},
	})
	ul.il.(*AtomicSwapInputLock).Signature = nil
	if err == nil {
		t.Errorf("should fail, as secret is missing/wrong")
	}

	err = ul.Lock(0, Transaction{}, AtomicSwapClaimKey{
		SecretKey: sk2,
		Secret:    secret,
	})
	if err != nil {
		t.Errorf("should succeed, as secret is correct: %v", err)
	}

	err = ul.StrictCheck()
	if err != nil {
		t.Errorf("strict check failed while it was expected to succeed: %v", err)
	}
	err = ul.Unlock(0, Transaction{})
	if err != nil {
		t.Errorf("unlock failed while it was expected to succeed: %v", err)
	}

	uh3 := ul.UnlockHash()
	if uh1.String() != uh3.String() {
		t.Error("inconsistent unlock hashes:", uh1, uh3)
	}
}

func TestInputLockTypeMarshaling(t *testing.T) {
	var ut InputLockType
	buffer := bytes.NewBuffer(nil)
	err := ut.MarshalSia(buffer)
	if err != nil {
		t.Errorf("error wile marshalling: %v", err)
	}
	if bytes.Compare([]byte{0}, buffer.Bytes()) != 0 {
		t.Errorf("invalid buffer: %v", buffer.Bytes())
	}
	err = ut.UnmarshalSia(buffer)
	if err != nil {
		t.Errorf("error wile unmarshalling: %v", err)
	}

	ut = InputLockTypeAtomicSwap
	err = ut.MarshalSia(buffer)
	if err != nil {
		t.Errorf("error wile marshalling: %v", err)
	}
	if bytes.Compare([]byte{byte(InputLockTypeAtomicSwap)}, buffer.Bytes()) != 0 {
		t.Errorf("invalid buffer: %v", buffer.Bytes())
	}
	err = ut.UnmarshalSia(buffer)
	if err != nil {
		t.Errorf("error wile unmarshalling: %v", err)
	}

	err = ut.UnmarshalSia(buffer)
	if err != io.EOF {
		t.Errorf("expected EOF but received: %v", err)
	}
}

func TestInputLockProxyEncoding(t *testing.T) {
	var ul InputLockProxy
	err := encoding.NewDecoder(bytes.NewReader([]byte{0})).Decode(&ul)
	if err != nil {
		t.Errorf("error while decoding a nil-input-porxy: %v", err)
	}
	if ul.t != InputLockTypeNil {
		t.Errorf("wrong inputLock type: %v", ul.t)
	}

	var entropy [crypto.EntropySize]byte
	sk, pk := crypto.GenerateKeyPairDeterministic(entropy)
	ul = NewSingleSignatureInputLock(Ed25519PublicKey(pk))
	err = ul.Lock(0, Transaction{}, sk)
	if err != nil {
		t.Errorf("error while locking InputLock: %v", err)
	}

	buf := bytes.NewBuffer(nil)
	err = ul.MarshalSia(buf)
	if err != nil {
		t.Errorf("error while marshaling InputLock: %v", err)
	}

	expectedBytes := []byte{1, 101, 100, 50, 53, 53, 49, 57, 0, 0, 0, 0, 0, 0, 0, 0, 0, 32, 0, 0, 0, 0, 0, 0, 0, 59, 106, 39, 188, 206, 182, 164, 45, 98, 163, 168, 208, 42, 111, 13, 115, 101, 50, 21, 119, 29, 226, 67, 166, 58, 192, 72, 161, 139, 89, 218, 41, 64, 0, 0, 0, 0, 0, 0, 0, 66, 25, 213, 18, 135, 159, 37, 24, 230, 55, 197, 175, 69, 230, 179, 61, 126, 220, 255, 213, 2, 30, 181, 202, 91, 195, 38, 154, 136, 120, 151, 38, 228, 28, 153, 72, 152, 85, 249, 167, 220, 161, 15, 61, 4, 55, 21, 134, 208, 142, 59, 180, 111, 50, 50, 126, 97, 12, 116, 246, 162, 235, 93, 9}
	if bytes.Compare(expectedBytes, buf.Bytes()) != 0 {
		t.Errorf("wrong marshaling: %v", buf.Bytes())
	}

	var ul2 InputLockProxy
	err = ul2.UnmarshalSia(buf)
	if err != nil {
		t.Errorf("error while unmarshaling InputLock: %v", err)
	}
	if !reflect.DeepEqual(ul, ul2) {
		t.Errorf("%v != %v", ul, ul2)
	}

	var secret, hashedSecret [sha256.Size]byte
	hashedSecret = sha256.Sum256(secret[:])

	var entropy2 [crypto.EntropySize]byte
	for i := range entropy2 {
		entropy2[i]++
	}
	sk2, pk2 := crypto.GenerateKeyPairDeterministic(entropy2)
	ul = NewAtomicSwapInputLock(
		Ed25519PublicKey(pk), Ed25519PublicKey(pk2),
		AtomicSwapHashedSecret(hashedSecret), CurrentTimestamp()+500000)
	err = ul.Lock(0, Transaction{}, AtomicSwapClaimKey{
		Secret:    secret,
		SecretKey: sk2,
	})
	if err != nil {
		t.Errorf("error while locking InputLock: %v", err)
	}

	err = ul.MarshalSia(buf)
	if err != nil {
		t.Errorf("error while marshaling InputLock: %v", err)
	}
	var ul3 InputLockProxy
	err = ul3.UnmarshalSia(buf)
	if err != nil {
		t.Errorf("error while unmarshaling InputLock: %v", err)
	}
	if !reflect.DeepEqual(ul, ul3) {
		t.Errorf("%v != %v", ul, ul3)
	}
}
