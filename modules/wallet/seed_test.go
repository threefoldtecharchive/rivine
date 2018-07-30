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
	t.Parallel()
	// Start with a blank wallet tester.
	wt, err := createBlankWalletTester(t.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer wt.closeWt()

	// Create a seed and unlock the wallet.
	seed, err := wt.wallet.Init(crypto.TwofishKey{})
	if err != nil {
		t.Fatal(err)
	}
	err = wt.wallet.Unlock(crypto.TwofishKey(crypto.HashObject(seed)))
	if err != nil {
		t.Fatal(err)
	}

	// Try getting an address, see that the seed advances correctly.
	primarySeed, remaining, err := wt.wallet.PrimarySeed()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(primarySeed[:], seed[:]) {
		t.Error("PrimarySeed is returning a value inconsitent with the seed returned by Encrypt")
	}
	if remaining != maxScanKeys {
		t.Error("primary seed is returning the wrong number of remaining addresses")
	}
	_, err = wt.wallet.NextAddress()
	if err != nil {
		t.Fatal(err)
	}
	_, remaining, err = wt.wallet.PrimarySeed()
	if err != nil {
		t.Fatal(err)
	}
	if remaining != maxScanKeys-1 {
		t.Error("primary seed is returning the wrong number of remaining addresses")
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
	err = wt.wallet.Unlock(crypto.TwofishKey(crypto.HashObject(seed)))
	if err != nil {
		t.Fatal(err)
	}
	primarySeed, remaining, err = wt.wallet.PrimarySeed()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(primarySeed[:], seed[:]) {
		t.Error("PrimarySeed is returning a value inconsitent with the seed returned by Encrypt")
	}
	if remaining != maxScanKeys-1 {
		t.Error("primary seed is returning the wrong number of remaining addresses")
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
		t.Fatal("AllSeeds should be returning the primary seed.")
	} else if allSeeds[0] != seed {
		t.Fatal("AllSeeds returned the wrong seed")
	}
	wt.wallet.Close()

	chainCts := types.DefaultChainConstants()
	bcInfo := types.DefaultBlockchainInfo()

	dir := filepath.Join(build.TempDir(modules.WalletDir, t.Name()+"1"), modules.WalletDir)
	w, err := New(wt.cs, wt.tpool, dir, bcInfo, chainCts)
	if err != nil {
		t.Fatal(err)
	}
	newSeed, err := w.Init(crypto.TwofishKey{})
	if err != nil {
		t.Fatal(err)
	}
	err = w.Unlock(crypto.TwofishKey(crypto.HashObject(newSeed)))
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
	err = w.LoadSeed(crypto.TwofishKey(crypto.HashObject(newSeed)), seed)
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
	if allSeeds[0] != newSeed {
		t.Error("AllSeeds returned the wrong seed")
	}
	if !bytes.Equal(allSeeds[1][:], seed[:]) {
		t.Error("AllSeeds returned the wrong seed")
	}
	err = cs.addTransactionAsBlock(
		types.NewEd25519PubKeyUnlockHash(generateKeys(newSeed, 1, 1)[0].PublicKey),
		types.NewCurrency64(1000))
	if err != nil {
		t.Errorf("failed to add transaction as block: %v", err)
	}

	siacoinBal2, _, err := w.ConfirmedBalance()
	if err != nil {
		t.Fatal(err)
	}
	if siacoinBal2.Cmp64(0) <= 0 {
		t.Error("wallet failed to load a seed with money in it")
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
}

// TestSweepSeedCoins tests that sweeping a seed results in the transfer of
// its siacoin outputs to the wallet.
func TestSweepSeedCoins(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	t.Parallel()
	cs := newConsensusSetStub()
	wt, err := createWalletTesterWithStubCS("TestSweepSeedCoins0", cs)
	if err != nil {
		t.Fatal(err)
	}
	defer wt.closeWt()
	seed, _, err := wt.wallet.PrimarySeed()
	if err != nil {
		t.Fatal(err)
	}
	// send money to ourselves, so that we sweep a real output (instead of
	// just a miner payout)
	uc, err := wt.wallet.NextAddress()
	if err != nil {
		t.Fatal(err)
	}
	chainCts := types.DefaultChainConstants()
	cs.addTransactionAsBlock(uc, chainCts.MinimumTransactionFee.Add(chainCts.CurrencyUnits.OneCoin))
	_, err = wt.wallet.SendCoins(chainCts.CurrencyUnits.OneCoin,
		types.NewCondition(types.NewUnlockHashCondition(uc)), nil)
	if err != nil {
		t.Fatal(err)
	}

	bcInfo := types.DefaultBlockchainInfo()
	// create a blank wallet
	dir := filepath.Join(build.TempDir(modules.WalletDir, "TestSweepSeedCoins1"), modules.WalletDir)
	w, err := New(wt.cs, wt.tpool, dir, bcInfo, chainCts)
	if err != nil {
		t.Fatal(err)
	}
	newSeed, err := w.Init(crypto.TwofishKey{})
	if err != nil {
		t.Fatal(err)
	}
	err = w.Unlock(crypto.TwofishKey(crypto.HashObject(newSeed)))
	if err != nil {
		t.Fatal(err)
	}
	// starting balance should be 0.
	siacoinBal, _, err := w.ConfirmedBalance()
	if err != nil {
		t.Fatal(err)
	}
	if !siacoinBal.IsZero() {
		t.Error("fresh wallet should not have a balance")
	}

	// sweep the seed of the first wallet into the second
	sweptCoins, _, err := w.SweepSeed(seed)
	if err != nil {
		t.Fatal(err)
	}

	// new wallet should have exactly 'sweptCoins' coins
	_, incoming, err := w.UnconfirmedBalance()
	if err != nil {
		t.Fatal(err)
	}
	if incoming.Cmp(sweptCoins) != 0 {
		t.Fatalf("wallet should have correct balance after sweeping seed: wanted %v, got %v", sweptCoins, incoming)
	}
}

// TODO: fix and enable
/*
// TestSweepSeedBlockStakes tests that sweeping a seed results in the transfer of
// its block stakes outputs to the wallet.
func TestSweepSeedBlockStakes(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	t.Parallel()
	cs := newConsensusSetStub()
	wt, err := createWalletTesterWithStubCS("TestSweepSeedCoins0", cs)
	if err != nil {
		t.Fatal(err)
	}
	defer wt.closeWt()

	_, siafundBal, _, err := wt.wallet.ConfirmedBalance()
	if err != nil {
		t.Fatal(err)
	}
	if siafundBal.Cmp(types.NewCurrency64(2000)) != 0 {
		t.Error("expecting a siafund balance of 2000 from the 1of1 key")
	}
	// need to reset the miner as well, since it depends on the wallet
	wt.miner, err = miner.New(wt.cs, wt.tpool, wt.wallet, wt.wallet.persistDir)
	if err != nil {
		t.Fatal(err)
	}

	// Create a seed and generate an address to send money to.
	seed := modules.Seed{1, 2, 3}
	sk := generateSpendableKey(seed, 1)

	// Send some siafunds to the address.
	_, err = wt.wallet.SendSiafunds(types.NewCurrency64(12), sk.UnlockConditions.UnlockHash())
	if err != nil {
		t.Fatal(err)
	}
	// Send some siacoins to the address, but not enough to cover the
	// transaction fee.
	_, err = wt.wallet.SendSiacoins(types.NewCurrency64(1), sk.UnlockConditions.UnlockHash())
	if err != nil {
		t.Fatal(err)
	}
	// mine blocks without earning payout until our balance is stable
	for i := types.BlockHeight(0); i < types.MaturityDelay; i++ {
		wt.addBlockNoPayout()
	}
	oldCoinBalance, siafundBal, _, err := wt.wallet.ConfirmedBalance()
	if err != nil {
		t.Fatal(err)
	}
	if siafundBal.Cmp(types.NewCurrency64(1988)) != 0 {
		t.Errorf("expecting balance of %v after sending siafunds to the seed, got %v", 1988, siafundBal)
	}

	// Sweep the seed.
	coins, funds, err := wt.wallet.SweepSeed(seed)
	if err != nil {
		t.Fatal(err)
	}
	if !coins.IsZero() {
		t.Error("expected to sweep 0 coins, got", coins)
	}
	if funds.Cmp(types.NewCurrency64(12)) != 0 {
		t.Errorf("expected to sweep %v funds, got %v", 12, funds)
	}
	// add a block without earning its payout
	wt.addBlockNoPayout()

	// Wallet balance should have decreased to pay for the sweep transaction.
	newCoinBalance, _, _, err := wt.wallet.ConfirmedBalance()
	if err != nil {
		t.Fatal(err)
	}
	if newCoinBalance.Cmp(oldCoinBalance) >= 0 {
		t.Error("expecting balance to go down; instead, increased by", newCoinBalance.Sub(oldCoinBalance))
	}
}
*/

// TestSweepSeedCoinsAndBlockStakes tests that sweeping a seed results in the transfer of
// its coins and block stakes outputs to the wallet.
func TestSweepSeedCoinsAndBlockStakes(t *testing.T) {
	// TODO
}

// TestSweepSeedNothing tests that sweeping a seed results in the transfer of
// its coins and block stakes outputs to the wallet, and that it simply returns an error with nothing changed,
// in case nothing can be sweeped.
func TestSweepSeedNothing(t *testing.T) {
	// TODO
}

// TestGenerateKeys tests that the generateKeys function correctly generates a
// key for every index specified.
func TestGenerateKeys(t *testing.T) {
	for i, k := range generateKeys(modules.Seed{}, 1000, 4000) {
		if len(k.UnlockHash().Hash) == 0 {
			t.Errorf("index %v was skipped", i)
		}
	}
}
