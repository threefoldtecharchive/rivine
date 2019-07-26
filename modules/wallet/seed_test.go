package wallet

import (
	"bytes"
	"path/filepath"
	"testing"

	"github.com/threefoldtech/rivine/build"
	"github.com/threefoldtech/rivine/crypto"
	"github.com/threefoldtech/rivine/modules"
	"github.com/threefoldtech/rivine/types"
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
	h, err := crypto.HashObject("TREZOR")
	if err != nil {
		t.Fatal(err)
	}
	encryptionKey := crypto.TwofishKey(h)
	seed, err := wt.wallet.Encrypt(encryptionKey, modules.Seed{})
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

	// recover first wallet into second wallet

	wt2, err := createBlankWalletTester(t.Name() + "2")
	if err != nil {
		t.Fatal(err)
	}
	defer wt2.closeWt()

	h2, err := crypto.HashObject("TREZOR 2")
	if err != nil {
		t.Fatal(err)
	}
	encryptionKey2 := crypto.TwofishKey(h2)
	seedDup, err := wt2.wallet.Encrypt(encryptionKey2, seed)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Compare(seed[:], seedDup[:]) != 0 {
		t.Fatal(seed, "!=", seedDup)
	}

	err = wt2.wallet.Unlock(encryptionKey2)
	if err != nil {
		t.Fatal(err)
	}

	primarySeed, progress, err = wt2.wallet.PrimarySeed()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(primarySeed[:], seed[:]) {
		t.Error("PrimarySeed is returning a value inconsitent with the seed returned by Encrypt",
			primarySeed, "!=", seed)
	}
	expectedProgress := uint64(modules.PublicKeysPerSeed - modules.WalletSeedPreloadDepth)
	if progress != expectedProgress {
		t.Error("progress reporting an unexpected value", progress, "!=", expectedProgress)
	}

	_, err = wt2.wallet.NextAddress()
	if err != nil {
		t.Fatal(err)
	}
	_, progress, err = wt2.wallet.PrimarySeed()
	if err != nil {
		t.Fatal(err)
	}
	expectedProgress++
	if progress != expectedProgress {
		t.Error("primary seed is returning the wrong progress", progress, "!=", expectedProgress)
	}

	// ensure we have all addresses of original wallet in the new wallet
	uhs1, err := wt.wallet.AllAddresses()
	if err != nil {
		t.Fatal(err)
	}
	uhs2, err := wt2.wallet.AllAddresses()
	if err != nil {
		t.Fatal(err)
	}
	il := len(uhs2)
	for _, uh1 := range uhs1 {
		l := len(uhs2)
		for i, uh2 := range uhs2 {
			if uh1.Cmp(uh2) == 0 {
				uhs2 = append(uhs2[:i], uhs2[i+1:]...)
				break
			}
		}
		if l == len(uhs1) {
			t.Error("couldn't find ", uh1, " in new wallet")
		}
	}
	if il-len(uhs1) != len(uhs2) {
		t.Error("couldn't find all original hashes in the new wallet")
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

	c, _, err := wt.wallet.ConfirmedBalance()
	if err != nil {
		t.Fatal(err)
	}
	if !c.Equals64(0) {
		t.Error("fresh wallet should not have a balance")
	}

	err = cs.addTransactionAsBlock(addr, types.NewCurrency64(1000))
	if err != nil {
		t.Errorf("failed to add transaction as block: %v", err)
	}

	c, _, err = wt.wallet.ConfirmedBalance()
	if err != nil {
		t.Fatal(err)
	}
	if !c.Equals64(1000) {
		t.Error("wallet requires 1000 coins at this point")
	}

	dir := filepath.Join(build.TempDir(modules.WalletDir, t.Name()+"1"), modules.WalletDir)
	w, err := New(wt.cs, wt.tpool, dir, types.DefaultBlockchainInfo(), types.TestnetChainConstants(), false)
	if err != nil {
		t.Fatal(err)
	}

	h, err := crypto.HashObject("TREZOR")
	if err != nil {
		t.Fatal(err)
	}
	encryptionKey := crypto.TwofishKey(h)
	newSeed, err := w.Encrypt(encryptionKey, modules.Seed{})
	if err != nil {
		t.Fatal(err)
	}
	err = w.Unlock(encryptionKey)
	if err != nil {
		t.Fatal(err)
	}
	// Balance of wallet should be 0.
	siacoinBal, _, err := w.ConfirmedBalance()
	if err != nil {
		t.Fatal(err)
	}
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
	err = cs.ConsensusSetSubscribe(w, modules.ConsensusChangeID{}, w.tg.StopChan())
	if err != nil {
		t.Errorf("Couldn't rescan by unsubscribing and subscribing again: %v", err)
	}

	// Balance of wallet should be 0.
	siacoinBal, _, err = w.ConfirmedBalance()
	if err != nil {
		t.Fatal(err)
	}
	if !siacoinBal.Equals64(1000) {
		t.Errorf("wallet with loaded key should have old balance but has: %v", siacoinBal)
	}

	// Rather than worry about a rescan, which isn't implemented and has
	// synchronization difficulties, just load a new wallet from the same
	// settings file - the same effect is achieved without the difficulties.
	w2, err := New(wt.cs, wt.tpool, dir, types.DefaultBlockchainInfo(), types.TestnetChainConstants(), false)
	if err != nil {
		t.Fatal(err)
	}
	err = w2.Unlock(encryptionKey)
	if err != nil {
		t.Fatal(err)
	}
	siacoinBal, _, err = w2.ConfirmedBalance()
	if err != nil {
		t.Fatal(err)
	}
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
