package wallet

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"math/big"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/rivine/rivine/build"
	"github.com/rivine/rivine/crypto"
	"github.com/rivine/rivine/modules"
	"github.com/rivine/rivine/modules/consensus"
	"github.com/rivine/rivine/modules/gateway"
	"github.com/rivine/rivine/modules/transactionpool"
	"github.com/rivine/rivine/types"
)

// A Wallet tester contains a ConsensusTester and has a bunch of helpful
// functions for facilitating wallet integration testing.
type walletTester struct {
	cs      modules.ConsensusSet
	gateway modules.Gateway
	tpool   modules.TransactionPool
	// miner   modules.TestMiner
	wallet *Wallet

	walletMasterKey crypto.TwofishKey

	persistDir string
}

// createWalletTester takes a testing.T and creates a WalletTester.
func createWalletTester(name string) (*walletTester, error) {
	bcInfo := types.DefaultBlockchainInfo()
	chainCts := types.DefaultChainConstants()
	// Create the modules
	testdir := build.TempDir(modules.WalletDir, name)
	g, err := gateway.New("localhost:0", false, filepath.Join(testdir, modules.GatewayDir), bcInfo, chainCts, nil)
	if err != nil {
		return nil, err
	}

	cs, err := consensus.New(g, false, filepath.Join(testdir, modules.ConsensusDir), bcInfo, chainCts)
	if err != nil {
		return nil, err
	}
	tp, err := transactionpool.New(cs, g, filepath.Join(testdir, modules.TransactionPoolDir), bcInfo, chainCts)
	if err != nil {
		return nil, err
	}
	w, err := New(cs, tp, filepath.Join(testdir, modules.WalletDir), bcInfo, chainCts)
	if err != nil {
		return nil, err
	}
	var masterKey crypto.TwofishKey
	_, err = rand.Read(masterKey[:])
	if err != nil {
		return nil, err
	}
	_, err = w.Init(masterKey)
	if err != nil {
		return nil, err
	}
	err = w.Unlock(masterKey)
	if err != nil {
		return nil, err
	}

	// m, err := miner.New(cs, tp, w, filepath.Join(testdir, modules.WalletDir))
	// if err != nil {
	// 	return nil, err
	// }

	// Assemble all components into a wallet tester.
	wt := &walletTester{
		cs:      cs,
		gateway: g,
		tpool:   tp,
		// miner:   m,
		wallet: w,

		walletMasterKey: masterKey,

		persistDir: testdir,
	}
	//
	// // Mine blocks until there is money in the wallet.
	// for i := types.BlockHeight(0); i <= types.MaturityDelay; i++ {
	// 	b, _ := wt.miner.FindBlock()
	// 	err := wt.cs.AcceptBlock(b)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// }
	return wt, nil
}

// createWalletTester takes a testing.T and creates a WalletTester.
// todo rename this and use if for all `createWalletTester` scenarios
func createWalletTesterWithStubCS(name string, cs *consensusSetStub) (*walletTester, error) {
	if cs == nil {
		return nil, errors.New("no stub consensus set given")
	}
	bcInfo := types.DefaultBlockchainInfo()
	chainCts := types.DefaultChainConstants()
	// Create the modules
	testdir := build.TempDir(modules.WalletDir, name)
	g, err := gateway.New("localhost:0", false, filepath.Join(testdir, modules.GatewayDir), bcInfo, chainCts, nil)
	if err != nil {
		return nil, err
	}

	tp, err := transactionpool.New(cs, g, filepath.Join(testdir, modules.TransactionPoolDir), bcInfo, chainCts)
	if err != nil {
		return nil, err
	}
	w, err := New(cs, tp, filepath.Join(testdir, modules.WalletDir), bcInfo, chainCts)
	if err != nil {
		return nil, err
	}
	var masterKey crypto.TwofishKey
	_, err = rand.Read(masterKey[:])
	if err != nil {
		return nil, err
	}
	_, err = w.Init(masterKey)
	if err != nil {
		return nil, err
	}
	err = w.Unlock(masterKey)
	if err != nil {
		return nil, err
	}

	// m, err := miner.New(cs, tp, w, filepath.Join(testdir, modules.WalletDir))
	// if err != nil {
	// 	return nil, err
	// }

	// Assemble all components into a wallet tester.
	wt := &walletTester{
		cs:      cs,
		gateway: g,
		tpool:   tp,
		// miner:   m,
		wallet: w,

		walletMasterKey: masterKey,

		persistDir: testdir,
	}
	//
	// // Mine blocks until there is money in the wallet.
	// for i := types.BlockHeight(0); i <= types.MaturityDelay; i++ {
	// 	b, _ := wt.miner.FindBlock()
	// 	err := wt.cs.AcceptBlock(b)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// }
	return wt, nil
}

// createBlankWalletTester creates a wallet tester that has not mined any
// blocks or encrypted the wallet.
func createBlankWalletTester(name string) (*walletTester, error) {
	chainCts := types.DefaultChainConstants()
	bcInfo := types.DefaultBlockchainInfo()

	// Create the modules
	testdir := build.TempDir(modules.WalletDir, name)
	g, err := gateway.New("localhost:0", false, filepath.Join(testdir, modules.GatewayDir), bcInfo, chainCts, nil)
	if err != nil {
		return nil, err
	}
	cs, err := consensus.New(g, false, filepath.Join(testdir, modules.ConsensusDir), bcInfo, chainCts)
	if err != nil {
		return nil, err
	}
	tp, err := transactionpool.New(cs, g, filepath.Join(testdir, modules.TransactionPoolDir), bcInfo, chainCts)
	if err != nil {
		return nil, err
	}
	w, err := New(cs, tp, filepath.Join(testdir, modules.WalletDir), bcInfo, chainCts)
	if err != nil {
		return nil, err
	}

	// Assemble all components into a wallet tester.
	wt := &walletTester{
		gateway: g,
		cs:      cs,
		tpool:   tp,
		wallet:  w,

		persistDir: testdir,
	}
	return wt, nil
}

// closeWt closes all of the modules in the wallet tester.
func (wt *walletTester) closeWt() error {
	errs := []error{
		wt.gateway.Close(),
		wt.cs.Close(),
		wt.tpool.Close(),
		wt.wallet.Close(),
	}
	return build.JoinErrors(errs, "; ")
}

// TestNilInputs tries starting the wallet using nil inputs.
func TestNilInputs(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	chainCts := types.DefaultChainConstants()
	bcInfo := types.DefaultBlockchainInfo()

	testdir := build.TempDir(modules.WalletDir, t.Name())
	g, err := gateway.New("localhost:0", false, filepath.Join(testdir, modules.GatewayDir), bcInfo, chainCts, nil)
	if err != nil {
		t.Fatal(err)
	}
	cs, err := consensus.New(g, false, filepath.Join(testdir, modules.ConsensusDir), bcInfo, chainCts)
	if err != nil {
		t.Fatal(err)
	}
	tp, err := transactionpool.New(cs, g, filepath.Join(testdir, modules.TransactionPoolDir), bcInfo, chainCts)
	if err != nil {
		t.Fatal(err)
	}

	wdir := filepath.Join(testdir, modules.WalletDir)
	_, err = New(cs, nil, wdir, bcInfo, chainCts)
	if err != errNilTpool {
		t.Error(err)
	}
	_, err = New(nil, tp, wdir, bcInfo, chainCts)
	if err != errNilConsensusSet {
		t.Error(err)
	}
	_, err = New(nil, nil, wdir, bcInfo, chainCts)
	if err != errNilConsensusSet {
		t.Error(err)
	}
}

// TestAllAddresses checks that AllAddresses returns all of the wallet's
// addresses in sorted order.
func TestAllAddresses(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	wt, err := createWalletTester(t.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer wt.closeWt()

	wt.wallet.keys = map[types.UnlockHash]spendableKey{}
	wt.wallet.keys[types.NewUnlockHash(1, crypto.Hash{0})] = spendableKey{}
	wt.wallet.keys[types.NewUnlockHash(0, crypto.Hash{1})] = spendableKey{}
	wt.wallet.keys[types.NewUnlockHash(1, crypto.Hash{5})] = spendableKey{}
	wt.wallet.keys[types.NewUnlockHash(0, crypto.Hash{5})] = spendableKey{}
	wt.wallet.keys[types.NewUnlockHash(0, crypto.Hash{0})] = spendableKey{}
	wt.wallet.keys[types.NewUnlockHash(1, crypto.Hash{1})] = spendableKey{}
	wt.wallet.keys[types.NewUnlockHash(1, crypto.Hash{2})] = spendableKey{}
	wt.wallet.keys[types.NewUnlockHash(0, crypto.Hash{3})] = spendableKey{}
	wt.wallet.keys[types.NewUnlockHash(0, crypto.Hash{2})] = spendableKey{}
	wt.wallet.keys[types.NewUnlockHash(0, crypto.Hash{4})] = spendableKey{}
	wt.wallet.keys[types.NewUnlockHash(1, crypto.Hash{3})] = spendableKey{}
	wt.wallet.keys[types.NewUnlockHash(1, crypto.Hash{4})] = spendableKey{}
	addrs, err := wt.wallet.AllAddresses()
	if err != nil {
		t.Fatal(err)
	}
	for i := range addrs[:5] {
		if addrs[i].Hash[0] != byte(i) {
			t.Error("address sorting failed:", i, addrs[i].Hash[0])
		}
	}
}

// TestCloseWallet tries to close the wallet.
func TestCloseWallet(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	chainCts := types.DefaultChainConstants()
	bcInfo := types.DefaultBlockchainInfo()

	testdir := build.TempDir(modules.WalletDir, t.Name())
	g, err := gateway.New("localhost:0", false, filepath.Join(testdir, modules.GatewayDir), bcInfo, chainCts, nil)
	if err != nil {
		t.Fatal(err)
	}
	cs, err := consensus.New(g, false, filepath.Join(testdir, modules.ConsensusDir), bcInfo, chainCts)
	if err != nil {
		t.Fatal(err)
	}
	tp, err := transactionpool.New(cs, g, filepath.Join(testdir, modules.TransactionPoolDir), bcInfo, chainCts)
	if err != nil {
		t.Fatal(err)
	}
	wdir := filepath.Join(testdir, modules.WalletDir)
	w, err := New(cs, tp, wdir, bcInfo, chainCts)
	if err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
}

// TestRescanning verifies that calling Rescanning during a scan operation
// returns true, and false otherwise.
func TestRescanning(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	wt, err := createWalletTester(t.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer wt.closeWt()

	// A fresh wallet should not be rescanning.
	rescanning, err := wt.wallet.Rescanning()
	if err != nil {
		t.Fatal(err)
	}
	if rescanning {
		t.Fatal("fresh wallet should not report that a scan is underway")
	}

	// lock the wallet
	wt.wallet.Lock()

	// spawn an unlock goroutine
	errChan := make(chan error)
	go func() {
		// acquire the write lock so that Unlock acquires the trymutex, but
		// cannot proceed further
		wt.wallet.mu.Lock()
		errChan <- wt.wallet.Unlock(wt.walletMasterKey)
	}()

	// wait for goroutine to start, after which Rescanning should return true
	time.Sleep(time.Millisecond * 10)
	rescanning, err = wt.wallet.Rescanning()
	if err != nil {
		t.Fatal(err)
	}
	if !rescanning {
		t.Fatal("wallet should report that a scan is underway")
	}

	// release the mutex and allow the call to complete
	wt.wallet.mu.Unlock()
	if err := <-errChan; err != nil {
		t.Fatal("unlock failed:", err)
	}

	// Rescanning should now return false again
	rescanning, err = wt.wallet.Rescanning()
	if err != nil {
		t.Fatal(err)
	}
	if rescanning {
		t.Fatal("wallet should not report that a scan is underway")
	}
}

// TestFutureAddressGeneration checks if the right amount of future addresses
// is generated after calling NextAddress() or locking + unlocking the wallet.
func TestLookaheadGeneration(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	wt, err := createWalletTester(t.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer wt.closeWt()

	// Check if number of future keys is correct
	wt.wallet.mu.RLock()
	progress, err := dbGetPrimarySeedProgress(wt.wallet.dbTx)
	wt.wallet.mu.RUnlock()
	if err != nil {
		t.Fatal("Couldn't fetch primary seed from db")
	}

	actualKeys := uint64(len(wt.wallet.lookahead))
	expectedKeys := maxLookahead(progress)
	if actualKeys != expectedKeys {
		t.Errorf("expected len(lookahead) == %d but was %d", actualKeys, expectedKeys)
	}

	// Generate some more keys
	for i := 0; i < 100; i++ {
		wt.wallet.NextAddress()
	}

	// Lock and unlock
	wt.wallet.Lock()
	wt.wallet.Unlock(wt.walletMasterKey)

	wt.wallet.mu.RLock()
	progress, err = dbGetPrimarySeedProgress(wt.wallet.dbTx)
	wt.wallet.mu.RUnlock()
	if err != nil {
		t.Fatal("Couldn't fetch primary seed from db")
	}

	actualKeys = uint64(len(wt.wallet.lookahead))
	expectedKeys = maxLookahead(progress)
	if actualKeys != expectedKeys {
		t.Errorf("expected len(lookahead) == %d but was %d", actualKeys, expectedKeys)
	}

	wt.wallet.mu.RLock()
	defer wt.wallet.mu.RUnlock()
	for i := range wt.wallet.keys {
		_, exists := wt.wallet.lookahead[i]
		if exists {
			t.Fatal("wallet keys contained a key which is also present in lookahead")
		}
	}
}

// TODO: fix tests
/*
// TestAdvanceLookaheadNoRescan tests if a transaction to multiple lookahead addresses
// is handled correctly without forcing a wallet rescan.
func TestAdvanceLookaheadNoRescan(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	cs := newConsensusSetStub()
	wt, err := createWalletTesterWithStubCS(t.Name(), cs)
	if err != nil {
		t.Fatal(err)
	}
	defer wt.closeWt()

	builder := wt.wallet.StartTransaction()
	if err != nil {
		t.Fatal(err)
	}
	payout := types.ZeroCurrency

	// Get the current progress
	wt.wallet.mu.RLock()
	progress, err := dbGetPrimarySeedProgress(wt.wallet.dbTx)
	wt.wallet.mu.RUnlock()
	if err != nil {
		t.Fatal("Couldn't fetch primary seed from db")
	}

	// choose 10 keys in the lookahead and remember them
	var receivingAddresses []types.UnlockHash
	for _, sk := range generateKeys(wt.wallet.primarySeed, progress, 10) {
		sco := types.CoinOutput{
			Condition: types.NewCondition(types.NewUnlockHashCondition(sk.UnlockHash())),
			Value:     types.NewCurrency64(1e3),
		}

		builder.AddCoinOutput(sco)
		payout = payout.Add(sco.Value)
		receivingAddresses = append(receivingAddresses, sk.UnlockHash())
	}

	uc, err := wt.wallet.NextAddress()
	if err != nil {
		t.Fatal(err)
	}
	cs.addTransactionAsBlock(uc, wt.wallet.chainCts.MinimumTransactionFee.Add(payout))

	err = builder.FundCoins(payout)
	if err != nil {
		t.Fatal(err)
	}

	tSet, err := builder.Sign()
	if err != nil {
		t.Fatal(err)
	}

	err = wt.tpool.AcceptTransactionSet(tSet)
	if err != nil {
		t.Fatal(err)
	}

	// Check if the receiving addresses were moved from future keys to keys
	wt.wallet.mu.RLock()
	defer wt.wallet.mu.RUnlock()
	for _, uh := range receivingAddresses {
		_, exists := wt.wallet.lookahead[uh]
		if exists {
			t.Fatal("UnlockHash still exists in wallet lookahead")
		}

		_, exists = wt.wallet.keys[uh]
		if !exists {
			t.Fatal("UnlockHash not in map of spendable keys")
		}
	}
}

// TestAdvanceLookaheadNoRescan tests if a transaction to multiple lookahead addresses
// is handled correctly forcing a wallet rescan.
func TestAdvanceLookaheadForceRescan(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	cs := newConsensusSetStub()
	wt, err := createWalletTesterWithStubCS(t.Name(), cs)
	if err != nil {
		t.Fatal(err)
	}
	defer wt.closeWt()

	// Get the current progress and balance
	wt.wallet.mu.RLock()
	progress, err := dbGetPrimarySeedProgress(wt.wallet.dbTx)
	wt.wallet.mu.RUnlock()
	if err != nil {
		t.Fatal("Couldn't fetch primary seed from db")
	}
	startBal, _, err := wt.wallet.ConfirmedBalance()
	if err != nil {
		t.Fatal(err)
	}

	// Send coins to an address with a high seed index, just outside the
	// lookahead range. It will not be initially detected, but later the
	// rescan should find it.
	highIndex := progress + uint64(len(wt.wallet.lookahead)) + 5
	farAddr := generateSpendableKey(wt.wallet.primarySeed, highIndex).UnlockHash()
	farPayout := types.DefaultChainConstants().CurrencyUnits.OneCoin.Mul64(8888)

	builder := wt.wallet.StartTransaction()
	builder.AddCoinOutput(types.CoinOutput{
		Condition: types.NewCondition(types.NewUnlockHashCondition(farAddr)),
		Value:     farPayout,
	})
	uc, err := wt.wallet.NextAddress()
	if err != nil {
		t.Fatal(err)
	}
	cs.addTransactionAsBlock(uc, wt.wallet.chainCts.MinimumTransactionFee.Add(farPayout))
	err = builder.FundCoins(farPayout)
	if err != nil {
		t.Fatal(err)
	}

	txnSet, err := builder.Sign()
	if err != nil {
		t.Fatal(err)
	}

	err = wt.tpool.AcceptTransactionSet(txnSet)
	if err != nil {
		t.Fatal(err)
	}

	newBal, _, err := wt.wallet.ConfirmedBalance()
	if err != nil {
		t.Fatal(err)
	}
	if !startBal.Sub(newBal).Equals(farPayout) {
		t.Fatal("wallet should not recognize coins sent to very high seed index")
	}

	builder = wt.wallet.StartTransaction()
	var payout types.Currency

	// choose 10 keys in the lookahead and remember them
	var receivingAddresses []types.UnlockHash
	for uh, index := range wt.wallet.lookahead {
		// Only choose keys that force a rescan
		if index < progress+lookaheadRescanThreshold {
			continue
		}
		sco := types.CoinOutput{
			Condition: types.NewCondition(types.NewUnlockHashCondition(uh)),
			Value:     types.DefaultChainConstants().CurrencyUnits.OneCoin.Mul64(1000),
		}
		builder.AddCoinOutput(sco)
		payout = payout.Add(sco.Value)
		receivingAddresses = append(receivingAddresses, uh)

		if len(receivingAddresses) >= 10 {
			break
		}
	}

	cs.addTransactionAsBlock(uc, wt.wallet.chainCts.MinimumTransactionFee.Add(payout))
	err = builder.FundCoins(payout)
	if err != nil {
		t.Fatal(err)
	}

	txnSet, err = builder.Sign()
	if err != nil {
		t.Fatal(err)
	}

	err = wt.tpool.AcceptTransactionSet(txnSet)
	if err != nil {
		t.Fatal(err)
	}

	// Allow the wallet rescan to finish
	time.Sleep(time.Second * 2)

	// Check that high seed index txn was discovered in the rescan
	rescanBal, _, err := wt.wallet.ConfirmedBalance()
	if err != nil {
		t.Fatal(err)
	}
	if !rescanBal.Equals(startBal) {
		t.Fatal("wallet did not discover txn after rescan")
	}

	// Check if the receiving addresses were moved from future keys to keys
	wt.wallet.mu.RLock()
	defer wt.wallet.mu.RUnlock()
	for _, uh := range receivingAddresses {
		_, exists := wt.wallet.lookahead[uh]
		if exists {
			t.Fatal("UnlockHash still exists in wallet lookahead")
		}

		_, exists = wt.wallet.keys[uh]
		if !exists {
			t.Fatal("UnlockHash not in map of spendable keys")
		}
	}
}
*/

/* // TODO: fix broken test
// TestDistantWallets tests if two wallets that use the same seed stay
// synchronized.
func TestDistantWallets(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	cs := newConsensusSetStub()
	wt, err := createWalletTesterWithStubCS(t.Name(), cs)
	if err != nil {
		t.Fatal(err)
	}
	defer wt.closeWt()

	chainCts := types.DefaultChainConstants()
	bcInfo := types.DefaultBlockchainInfo()

	// Create another wallet with the same seed.
	w2, err := New(wt.cs, wt.tpool, build.TempDir(modules.WalletDir, t.Name()+"2", modules.WalletDir), bcInfo, chainCts)
	if err != nil {
		t.Fatal(err)
	}
	err = w2.InitFromSeed(crypto.TwofishKey{}, wt.wallet.primarySeed)
	if err != nil {
		t.Fatal(err)
	}
	err = w2.Unlock(crypto.TwofishKey(crypto.HashObject(wt.wallet.primarySeed)))
	if err != nil {
		t.Fatal(err)
	}

	uc1, err := wt.wallet.NextAddress()
	if err != nil {
		t.Fatal(err)
	}

	cs.addTransactionAsBlock(
		uc1,
		chainCts.MinimumTransactionFee.Add(chainCts.CurrencyUnits.OneCoin).
			Mul64(uint64(lookaheadBuffer/2)),
	)

	// Use the first wallet.
	for i := uint64(0); i < lookaheadBuffer/2; i++ {
		_, err = wt.wallet.SendCoins(chainCts.CurrencyUnits.OneCoin, types.NewCondition(&types.NilCondition{}), nil)
		if err != nil {
			t.Fatal(err)
		}
	}

	// The second wallet's balance should update accordingly.
	w1bal, _, err := wt.wallet.ConfirmedBalance()
	if err != nil {
		t.Fatal(err)
	}
	w2bal, _, err := w2.ConfirmedBalance()
	if err != nil {
		t.Fatal(err)
	}

	if !w1bal.Equals(w2bal) {
		t.Fatal("balances do not match:", w1bal, w2bal)
	}

	// Send coins to an address with a very high seed index, outside the
	// lookahead range. w2 should not detect it.
	tbuilder := wt.wallet.StartTransaction()
	farAddr := generateSpendableKey(wt.wallet.primarySeed, lookaheadBuffer*10).UnlockHash()
	value := chainCts.CurrencyUnits.OneCoin.Mul64(1e3)
	tbuilder.AddCoinOutput(types.CoinOutput{
		Condition: types.NewCondition(types.NewUnlockHashCondition(farAddr)),
		Value:     value,
	})
	cs.addTransactionAsBlock(
		uc1,
		chainCts.MinimumTransactionFee.Add(value))
	err = tbuilder.FundCoins(value)
	if err != nil {
		t.Fatal(err)
	}
	txnSet, err := tbuilder.Sign()
	if err != nil {
		t.Fatal(err)
	}
	err = wt.tpool.AcceptTransactionSet(txnSet)
	if err != nil {
		t.Fatal(err)
	}

	if newBal, _, err := w2.ConfirmedBalance(); !newBal.Equals(w2bal.Sub(value)) {
		if err != nil {
			t.Fatal(err)
		}
		// TODO: FAILS HERE!!
		t.Fatal("wallet should not recognize coins sent to very high seed index")
	}
}*/

func newConsensusSetStub() *consensusSetStub {
	chainCts := types.DefaultChainConstants()
	return &consensusSetStub{
		blocks: []types.Block{
			chainCts.GenesisBlock(),
		},
		subscribers: make(map[modules.ConsensusSetSubscriber]struct{}),
	}
}

type consensusSetStub struct {
	blocks      []types.Block
	subscribers map[modules.ConsensusSetSubscriber]struct{}
}

func (css *consensusSetStub) addTransactionAsBlock(unlockHash types.UnlockHash, value types.Currency) error {
	l := len(css.blocks)
	if l == 0 {
		return errors.New("invalid block list in consensus set")
	}
	return css.AcceptBlock(types.Block{
		ParentID:  css.blocks[l-1].ID(),
		Timestamp: types.CurrentTimestamp(),
		Transactions: []types.Transaction{
			{
				Version: types.DefaultChainConstants().DefaultTransactionVersion,
				CoinOutputs: []types.CoinOutput{
					{
						Value:     value,
						Condition: types.NewCondition(types.NewUnlockHashCondition(unlockHash)),
					},
				},
			},
		},
	})
}

func (css *consensusSetStub) revertBlock() error {
	li := len(css.blocks) - 1
	if li < 0 {
		return errors.New("there are no blocks to revert")
	}
	block := css.blocks[li]
	css.blocks = css.blocks[:li]
	for subscriber := range css.subscribers {
		processRevertedBlock(block, subscriber)
	}

	return nil
}

func (css *consensusSetStub) AcceptBlock(block types.Block) error {
	id := block.ID()
	for _, b := range css.blocks {
		if b.ID() == id {
			return errors.New("block seen before")
		}
	}
	css.blocks = append(css.blocks, block)

	for subscriber := range css.subscribers {
		processAppliedBlock(block, subscriber)
	}

	return nil
}

func processAppliedBlock(block types.Block, subscriber modules.ConsensusSetSubscriber) {
	cc := modules.ConsensusChange{
		ID:            modules.ConsensusChangeID(crypto.HashObject(block)),
		AppliedBlocks: []types.Block{block},
	}
	for _, tx := range block.Transactions {
		for _, co := range tx.CoinOutputs {
			cc.CoinOutputDiffs = append(cc.CoinOutputDiffs, modules.CoinOutputDiff{
				Direction:  modules.DiffApply,
				ID:         types.CoinOutputID(crypto.HashObject(co)),
				CoinOutput: co,
			})
		}
	}
	subscriber.ProcessConsensusChange(cc)
}

func processRevertedBlock(block types.Block, subscriber modules.ConsensusSetSubscriber) {
	cc := modules.ConsensusChange{
		ID:             modules.ConsensusChangeID(crypto.HashObject(block)),
		RevertedBlocks: []types.Block{block},
	}
	for _, tx := range block.Transactions {
		for _, co := range tx.CoinOutputs {
			cc.CoinOutputDiffs = append(cc.CoinOutputDiffs, modules.CoinOutputDiff{
				Direction:  modules.DiffRevert,
				ID:         types.CoinOutputID(crypto.HashObject(co)),
				CoinOutput: co,
			})
		}
	}
	subscriber.ProcessConsensusChange(cc)
}

func (css *consensusSetStub) BlockAtHeight(height types.BlockHeight) (types.Block, bool) {
	if height >= types.BlockHeight(len(css.blocks)) {
		return types.Block{}, false
	}
	return css.blocks[height], true
}

func (css *consensusSetStub) BlockHeightOfBlock(block types.Block) (types.BlockHeight, bool) {
	id := block.ID()
	for height, b := range css.blocks {
		if b.ID() == id {
			return types.BlockHeight(height), true
		}
	}
	return 0, false
}

func (css *consensusSetStub) TransactionAtShortID(shortID types.TransactionShortID) (types.Transaction, bool) {
	height := shortID.BlockHeight()
	block, found := css.BlockAtHeight(height)
	if !found {
		return types.Transaction{}, false
	}

	txSeqID := int(shortID.TransactionSequenceIndex())
	if len(block.Transactions) <= txSeqID {
		return types.Transaction{}, false
	}

	return block.Transactions[txSeqID], true
}

func (css *consensusSetStub) TransactionAtID(id types.TransactionID) (types.Transaction, types.TransactionShortID, bool) {
	for i, b := range css.blocks {
		for j, t := range b.Transactions {
			if t.ID() == id {
				return t, types.NewTransactionShortID(types.BlockHeight(i), uint16(j)), true
			}
		}
	}
	return types.Transaction{}, 0, false
}

func (css *consensusSetStub) FindParentBlock(b types.Block, depth types.BlockHeight) (block types.Block, exists bool) {
	var blockIndex int
	for i, block := range css.blocks {
		if block.Header().ID() == b.Header().ID() {
			blockIndex = i
			break
		}
	}
	if int(depth) > blockIndex {
		return types.Block{}, false
	}
	return css.blocks[blockIndex-int(depth)], true
}

func (css *consensusSetStub) ChildTarget(id types.BlockID) (types.Target, bool) {
	// TODO: return a more sensible value if required
	return types.Target{}, false
}

func (css *consensusSetStub) Close() error {
	return nil
}

func (css *consensusSetStub) CurrentBlock() types.Block {
	l := len(css.blocks)
	if l == 0 {
		return types.Block{}
	}
	return css.blocks[l-1]
}

func (css *consensusSetStub) Flush() error {
	return nil
}

func (css *consensusSetStub) Height() types.BlockHeight {
	return types.BlockHeight(len(css.blocks))
}

func (css *consensusSetStub) Synced() bool {
	return true
}

func (css *consensusSetStub) InCurrentPath(id types.BlockID) bool {
	for _, b := range css.blocks {
		if b.ID() == id {
			return true
		}
	}
	return false
}

func (css *consensusSetStub) MinimumValidChildTimestamp(id types.BlockID) (types.Timestamp, bool) {
	if len(css.blocks) == 0 {
		return 0, false
	}
	return css.blocks[0].Timestamp, true
}

func (css *consensusSetStub) CalculateStakeModifier(height types.BlockHeight, block types.Block, delay types.BlockHeight) *big.Int {
	//TODO: check if a new Stakemodifier needs to be calculated. The stakemodifier
	// only change when a new block is created, and this calculation is also needed
	// to validate an incomming new block

	// make a signed version of the current height because sub genesis block is
	// possible here.
	signedHeight := int64(height)
	signedHeight -= int64(types.DefaultChainConstants().StakeModifierDelay)

	mask := big.NewInt(1)
	var BlockIDHash *big.Int
	stakemodifier := big.NewInt(0)
	var buffer bytes.Buffer

	// Rollback the required amount of blocks, minus 1. This way we end up at the direct child of the
	// block we use to calculate the stakemodifer, rather than the actual first block. Simplifies
	// the main loop a bit
	block, _ = css.FindParentBlock(block, delay-1)

	// We have the direct child of the first block used in the stake modifier calculation. As such
	// we can follow the parentID in the block to retrieve all the blocks required, using 1 bit
	// of each blocks ID to calculate the stake modifier
	for i := 0; i < 256; i++ {
		if signedHeight >= 0 {
			var exist bool
			block, exist = css.FindParentBlock(block, 1)
			if build.DEBUG && !exist {
				panic("block to be used for stakemodifier does not yet exist")
			}
			hashof := block.ID()
			BlockIDHash = big.NewInt(0).SetBytes(hashof[:])
		} else {
			// if the counter goes sub genesis block , calculate a predefined hash
			// from the sub genesis distance.
			buffer.WriteString("genesis" + strconv.FormatInt(signedHeight, 10))
			hashof := sha256.Sum256(buffer.Bytes())
			BlockIDHash = big.NewInt(0).SetBytes(hashof[:])
		}

		stakemodifier.Or(stakemodifier, big.NewInt(0).And(BlockIDHash, mask))
		mask.Mul(mask, big.NewInt(2)) //shift 1 bit to left (more close to msb)
		signedHeight--
	}
	return stakemodifier
}

func (css *consensusSetStub) TryTransactionSet(txs []types.Transaction) (change modules.ConsensusChange, err error) {
	l := len(css.blocks)
	if l == 0 {
		return modules.ConsensusChange{}, errors.New("invalid block list in consensus set")
	}
	block := types.Block{
		ParentID:     css.blocks[l-1].ID(),
		Timestamp:    types.CurrentTimestamp(),
		Transactions: txs,
	}
	cc := modules.ConsensusChange{
		ID: modules.ConsensusChangeID(crypto.HashObject(block)),
	}
	for _, tx := range block.Transactions {
		for _, co := range tx.CoinOutputs {
			cc.CoinOutputDiffs = append(cc.CoinOutputDiffs, modules.CoinOutputDiff{
				Direction:  modules.DiffApply,
				ID:         types.CoinOutputID(crypto.HashObject(co)),
				CoinOutput: co,
			})
		}
	}
	return cc, nil
}

func (css *consensusSetStub) ConsensusSetSubscribe(subscriber modules.ConsensusSetSubscriber, changeID modules.ConsensusChangeID) error {
	if _, ok := css.subscribers[subscriber]; ok {
		return errors.New("subscriber already registered to stub consensus set")
	}
	css.subscribers[subscriber] = struct{}{}

	var i int
	if changeID != modules.ConsensusChangeID(crypto.Hash{}) {
		for ; i < len(css.blocks); i++ {
			if modules.ConsensusChangeID(crypto.HashObject(css.blocks[i])) == changeID {
				break
			}
		}
	}
	for _, block := range css.blocks[i:] {
		processAppliedBlock(block, subscriber)
	}
	return nil
}

func (css *consensusSetStub) Unsubscribe(subscriber modules.ConsensusSetSubscriber) {
	delete(css.subscribers, subscriber)
}

func (css *consensusSetStub) GetCoinOutput(id types.CoinOutputID) (co types.CoinOutput, err error) {
	for _, block := range css.blocks {
		for _, txn := range block.Transactions {
			for i, co := range txn.CoinOutputs {
				if txn.CoinOutputID(uint64(i)) == id {
					return co, nil
				}
			}
		}
	}
	return types.CoinOutput{}, errors.New("Coin output not found in database")
}

func (css *consensusSetStub) GetBlockStakeOutput(id types.BlockStakeOutputID) (bso types.BlockStakeOutput, err error) {
	for _, block := range css.blocks {
		for _, txn := range block.Transactions {
			for i, bso := range txn.BlockStakeOutputs {
				if txn.BlockStakeOutputID(uint64(i)) == id {
					return bso, nil
				}
			}
		}
	}
	return types.BlockStakeOutput{}, errors.New("BlockStake output not found in database")
}
