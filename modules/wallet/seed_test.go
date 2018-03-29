package wallet

import (
	"bytes"
	"path/filepath"
	"testing"

	"github.com/rivine/rivine/build"
	"github.com/rivine/rivine/crypto"
	"github.com/rivine/rivine/modules"
	"github.com/rivine/rivine/types"
)

// TestPrimarySeed checks that the correct seed is returned when calling
// PrimarySeed.
func TestPrimarySeed(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	// Start with a blank wallet tester.
	wt, err := createBlankWalletTester(t.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer wt.closeWt()

	// Create a seed and unlock the wallet.
	encryptionKey := crypto.TwofishKey(crypto.HashObject("TREZOR"))
	seed, err := wt.wallet.Encrypt(encryptionKey)
	if err != nil {
		t.Fatal(err)
	}
	err = wt.wallet.Unlock(encryptionKey)
	if err != nil {
		t.Fatal(err)
	}

	// Try getting an address, see that the seed advances correctly.
	primarySeed, progress, err := wt.wallet.PrimarySeed()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(primarySeed[:], seed[:]) {
		t.Error("PrimarySeed is returning a value inconsitent with the seed returned by Encrypt")
	}
	if progress != 0 {
		t.Error("primary seed is returning the wrong progress")
	}
	_, err = wt.wallet.NextAddress()
	if err != nil {
		t.Fatal(err)
	}
	_, progress, err = wt.wallet.PrimarySeed()
	if err != nil {
		t.Fatal(err)
	}
	if progress != 1 {
		t.Error("primary seed is returning the wrong progress")
	}

	// Lock then unlock the wallet and check the responses.
	err = wt.wallet.Lock()
	if err != nil {
		t.Fatal(err)
	}
	_, _, err = wt.wallet.PrimarySeed()
	if err != modules.ErrLockedWallet {
		t.Error("unexpected err:", err)
	}
	err = wt.wallet.Unlock(encryptionKey)
	if err != nil {
		t.Fatal(err)
	}
	primarySeed, progress, err = wt.wallet.PrimarySeed()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(primarySeed[:], seed[:]) {
		t.Error("PrimarySeed is returning a value inconsitent with the seed returned by Encrypt")
	}
	if progress != 1 {
		t.Error("progress reporting an unexpected value")
	}
}

// TestLoadSeed checks that a seed can be successfully recovered from a wallet,
// and then remain available on subsequent loads of the wallet.
func TestLoadSeed(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	t.Parallel()

	cs := newConsensusSetStub()

	wt, err := createWalletTesterWithStubCS(t.Name(), cs)
	if err != nil {
		t.Fatal(err)
	}
	defer wt.closeWt()
	seed, _, err := wt.wallet.PrimarySeed()
	if err != nil {
		t.Fatal(err)
	}
	allSeeds, err := wt.wallet.AllSeeds()
	if err != nil {
		t.Fatal(err)
	}
	if len(allSeeds) != 1 {
		t.Error("AllSeeds should be returning the primary seed.")
	}
	if allSeeds[0] != seed {
		t.Error("AllSeeds returned the wrong seed")
	}

	addr, err := wt.wallet.NextAddress()
	if err != nil {
		t.Errorf("next address couldn't be created: %v", err)
	}

	c, _ := wt.wallet.ConfirmedBalance()
	if !c.Equals64(0) {
		t.Error("fresh wallet should not have a balance")
	}

	err = cs.addTransactionAsBlock(addr, types.NewCurrency64(1000))
	if err != nil {
		t.Errorf("failed to add transaction as block: %v", err)
	}

	c, _ = wt.wallet.ConfirmedBalance()
	if !c.Equals64(1000) {
		t.Error("wallet requires 1000 coins at this point")
	}

	dir := filepath.Join(build.TempDir(modules.WalletDir, t.Name()+"1"), modules.WalletDir)
	w, err := New(wt.cs, wt.tpool, dir, types.DefaultBlockchainInfo(), types.DefaultChainConstants())
	if err != nil {
		t.Fatal(err)
	}
	encryptionKey := crypto.TwofishKey(crypto.HashObject("TREZOR"))
	newSeed, err := w.Encrypt(encryptionKey)
	if err != nil {
		t.Fatal(err)
	}
	err = w.Unlock(encryptionKey)
	if err != nil {
		t.Fatal(err)
	}
	// Balance of wallet should be 0.
	siacoinBal, _ := w.ConfirmedBalance()
	if !siacoinBal.Equals64(0) {
		t.Error("fresh wallet should not have a balance")
	}
	err = w.LoadSeed(encryptionKey, seed)
	if err != nil {
		t.Fatal(err)
	}
	allSeeds, err = w.AllSeeds()
	if err != nil {
		t.Fatal(err)
	}
	if len(allSeeds) != 2 {
		t.Error("AllSeeds should be returning the primary seed with the recovery seed.")
	}
	if !bytes.Equal(allSeeds[0][:], newSeed[:]) {
		t.Error("AllSeeds returned the wrong seed")
	}
	if !bytes.Equal(allSeeds[1][:], seed[:]) {
		t.Error("AllSeeds returned the wrong seed")
	}

	// rescan
	cs.Unsubscribe(w)
	err = cs.ConsensusSetSubscribe(w, modules.ConsensusChangeID{})
	if err != nil {
		t.Errorf("Couldn't rescan by unsubscribing and subscribing again: %v", err)
	}

	// Balance of wallet should be 0.
	siacoinBal, _ = w.ConfirmedBalance()
	if !siacoinBal.Equals64(1000) {
		t.Errorf("wallet with loaded key should have old balance but has: %v", siacoinBal)
	}

	// Rather than worry about a rescan, which isn't implemented and has
	// synchronization difficulties, just load a new wallet from the same
	// settings file - the same effect is achieved without the difficulties.
	w2, err := New(wt.cs, wt.tpool, dir, types.DefaultBlockchainInfo(), types.DefaultChainConstants())
	if err != nil {
		t.Fatal(err)
	}
	err = w2.Unlock(encryptionKey)
	if err != nil {
		t.Fatal(err)
	}
	siacoinBal, _ = w2.ConfirmedBalance()
	if !siacoinBal.Equals64(1000) {
		t.Errorf("wallet with loaded key should have old balance but has: %v", siacoinBal)
	}
	allSeeds, err = w2.AllSeeds()
	if err != nil {
		t.Fatal(err)
	}
	if len(allSeeds) != 2 {
		t.Error("AllSeeds should be returning the primary seed with the recovery seed.")
	}
	if !bytes.Equal(allSeeds[0][:], newSeed[:]) {
		t.Error("AllSeeds returned the wrong seed")
	}
	if !bytes.Equal(allSeeds[1][:], seed[:]) {
		t.Error("AllSeeds returned the wrong seed")
	}
}
