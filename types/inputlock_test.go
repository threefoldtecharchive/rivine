package types

import (
	"testing"

	"github.com/rivine/rivine/crypto"
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
