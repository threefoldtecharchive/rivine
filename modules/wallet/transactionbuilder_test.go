package wallet

import (
	"testing"

	"github.com/rivine/rivine/types"
)

// TestDoubleSignError checks that an error is returned if there is a problem
// when trying to call 'Sign' on a transaction twice.
func TestDoubleSignError(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	cs := newConsensusSetStub()
	wt, err := createWalletTesterWithStubCS(t.Name(), cs)
	if err != nil {
		t.Fatal(err)
	}
	defer wt.closeWt()

	uc, err := wt.wallet.NextAddress()
	if err != nil {
		t.Fatal(err)
	}
	txnFund := types.NewCurrency64(100e9)
	cs.addTransactionAsBlock(uc, txnFund.Add(txnFund))

	// Create a transaction, add money to it, and then call sign twice.
	b := wt.wallet.StartTransaction()
	err = b.FundCoins(txnFund)
	if err != nil {
		t.Fatal(err)
	}
	_ = b.AddMinerFee(txnFund)
	txnSet, err := b.Sign()
	if err != nil {
		t.Fatal(err)
	}
	txnSet2, err := b.Sign()
	if err != errBuilderAlreadySigned {
		t.Error("the wrong error is being returned after a double call to sign")
	}
	if err != nil && txnSet2 != nil {
		t.Error("errored call to sign did not return a nil txn set")
	}
	err = wt.tpool.AcceptTransactionSet(txnSet)
	if err != nil {
		t.Fatal(err)
	}
}

/* // TODO: enable

// TestConcurrentBuilders checks that multiple transaction builders can safely
// be opened at the same time, and that they will make valid transactions when
// building concurrently.
func TestConcurrentBuilders(t *testing.T) {
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
	funding := types.NewCurrency64(10e3).Mul(chainCts.CurrencyUnits.OneCoin)
	uc, err := wt.wallet.NextAddress()
	if err != nil {
		t.Fatal(err)
	}
	cs.addTransactionAsBlock(uc, funding.Div64(2))
	cs.addTransactionAsBlock(uc, funding.Div64(2))
	cs.addTransactionAsBlock(uc, funding.Div64(2))
	cs.addTransactionAsBlock(uc, funding.Div64(2))
	cs.addTransactionAsBlock(uc, funding.Div64(2))
	cs.addTransactionAsBlock(uc, funding.Div64(2))
	cs.addTransactionAsBlock(uc, funding.Div64(2))
	cs.addTransactionAsBlock(uc, funding.Div64(2))

	// Get a baseline balance for the wallet.
	startingSCConfirmed, _, err := wt.wallet.ConfirmedBalance()
	if err != nil {
		t.Fatal(err)
	}
	startingOutgoing, startingIncoming, err := wt.wallet.UnconfirmedBalance()
	if err != nil {
		t.Fatal(err)
	}
	if !startingOutgoing.IsZero() {
		t.Fatal(startingOutgoing)
	}
	if !startingIncoming.IsZero() {
		t.Fatal(startingIncoming)
	}

	// Create two builders at the same time, then add money to each.
	builder1 := wt.wallet.StartTransaction()
	builder2 := wt.wallet.StartTransaction()
	// Fund each builder with a siacoin output that is smaller than all of the
	// outputs that the wallet should currently have.
	err = builder1.FundCoins(funding)
	if err != nil {
		t.Fatal(err)
	}
	err = builder2.FundCoins(funding)
	if err != nil {
		t.Fatal(err)
	}

	// Get a second reading on the wallet's balance.
	fundedSCConfirmed, _, err := wt.wallet.ConfirmedBalance()
	if err != nil {
		t.Fatal(err)
	}
	if !startingSCConfirmed.Equals(fundedSCConfirmed) {
		t.Fatal("confirmed siacoin balance changed when no blocks have been mined", startingSCConfirmed, fundedSCConfirmed)
	}

	// Spend the transaction funds on miner fees and the void output.
	builder1.AddMinerFee(types.NewCurrency64(25).Mul(chainCts.CurrencyUnits.OneCoin))
	builder2.AddMinerFee(types.NewCurrency64(25).Mul(chainCts.CurrencyUnits.OneCoin))
	// Send the money to the void.
	output := types.CoinOutput{Value: types.NewCurrency64(9975).Mul(chainCts.CurrencyUnits.OneCoin)}
	builder1.AddCoinOutput(output)
	builder2.AddCoinOutput(output)

	// Sign the transactions and verify that both are valid.
	tset1, err := builder1.Sign()
	if err != nil {
		t.Fatal(err)
	}
	tset2, err := builder2.Sign()
	if err != nil {
		t.Fatal(err)
	}
	err = wt.tpool.AcceptTransactionSet(tset1)
	if err != nil {
		t.Fatal(err)
	}
	err = wt.tpool.AcceptTransactionSet(tset2)
	if err != nil {
		t.Fatal(err)
	}
}

// TestConcurrentBuildersSingleOutput probes the behavior when multiple
// builders are created at the same time, but there is only a single wallet
// output that they end up needing to share.
func TestConcurrentBuildersSingleOutput(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	wt, err := createWalletTester(t.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer wt.closeWt()

	// Send all coins to a single confirmed output for the wallet.
	addr, err := wt.wallet.NextAddress()
	if err != nil {
		t.Fatal(err)
	}
	scBal, _, err := wt.wallet.ConfirmedBalance()
	if err != nil {
		t.Fatal(err)
	}
	// Use a custom builder so that there is no transaction fee.
	builder := wt.wallet.StartTransaction()
	err = builder.FundCoins(scBal)
	if err != nil {
		t.Fatal(err)
	}
	output := types.CoinOutput{
		Value:     scBal,
		Condition: types.NewCondition(types.NewUnlockHashCondition(addr)),
	}
	builder.AddCoinOutput(output)
	tSet, err := builder.Sign()
	if err != nil {
		t.Fatal(err)
	}
	err = wt.tpool.AcceptTransactionSet(tSet)
	if err != nil {
		t.Fatal(err)
	}
	//
	//	// Get the transaction into the blockchain without giving a miner payout to
	//	// the wallet.
	//	err = wt.addBlockNoPayout()
	//	if err != nil {
	//		t.Fatal(err)
	//	}
	//

	// Get a baseline balance for the wallet.
	startingSCConfirmed, _, err := wt.wallet.ConfirmedBalance()
	if err != nil {
		t.Fatal(err)
	}
	startingOutgoing, startingIncoming, err := wt.wallet.UnconfirmedBalance()
	if err != nil {
		t.Fatal(err)
	}
	if !startingOutgoing.IsZero() {
		t.Fatal(startingOutgoing)
	}
	if !startingIncoming.IsZero() {
		t.Fatal(startingIncoming)
	}

	chainCts := types.DefaultChainConstants()

	// Create two builders at the same time, then add money to each.
	builder1 := wt.wallet.StartTransaction()
	builder2 := wt.wallet.StartTransaction()
	// Fund each builder with a siacoin output.
	funding := types.NewCurrency64(10e3).Mul(chainCts.CurrencyUnits.OneCoin)
	err = builder1.FundCoins(funding)
	if err != nil {
		t.Fatal(err)
	}
	// This add should fail, blocking the builder from completion.
	err = builder2.FundCoins(funding)
	if err != modules.ErrIncompleteTransactions {
		t.Fatal(err)
	}

	// Get a second reading on the wallet's balance.
	fundedSCConfirmed, _, err := wt.wallet.ConfirmedBalance()
	if err != nil {
		t.Fatal(err)
	}
	if !startingSCConfirmed.Equals(fundedSCConfirmed) {
		t.Fatal("confirmed siacoin balance changed when no blocks have been mined", startingSCConfirmed, fundedSCConfirmed)
	}

	// Spend the transaction funds on miner fees and the void output.
	builder1.AddMinerFee(types.NewCurrency64(25).Mul(chainCts.CurrencyUnits.OneCoin))
	// Send the money to the void.
	output = types.CoinOutput{Value: types.NewCurrency64(9975).Mul(chainCts.CurrencyUnits.OneCoin)}
	builder1.AddCoinOutput(output)

	// Sign the transaction and submit it.
	tset1, err := builder1.Sign()
	if err != nil {
		t.Fatal(err)
	}
	err = wt.tpool.AcceptTransactionSet(tset1)
	if err != nil {
		t.Fatal(err)
	}

	//
	//	// Mine a block to get the transaction sets into the blockchain.
	//	_, err = wt.miner.AddBlock()
	//	if err != nil {
	//		t.Fatal(err)
	//	}
	//
}

// TestParallelBuilders checks that multiple transaction builders can safely be
// opened at the same time, and that they will make valid transactions when
// building concurrently, using multiple gothreads to manage the builders.
func TestParallelBuilders(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	wt, err := createWalletTester(t.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer wt.closeWt()

	outputsDesired := 10
	//
	//	// Mine a few more blocks so that the wallet has lots of outputs to pick
	//	// from.
	//	for i := 0; i < outputsDesired; i++ {
	//		_, err := wt.miner.AddBlock()
	//		if err != nil {
	//			t.Fatal(err)
	//		}
	//	}
	//	// Add MatruityDelay blocks with no payout to make tracking the balance
	//	// easier.
	//	for i := types.BlockHeight(0); i < types.MaturityDelay+1; i++ {
	//		err = wt.addBlockNoPayout()
	//		if err != nil {
	//			t.Fatal(err)
	//		}
	//	}

	// Get a baseline balance for the wallet.
	startingSCConfirmed, _, err := wt.wallet.ConfirmedBalance()
	if err != nil {
		t.Fatal(err)
	}
	startingOutgoing, startingIncoming, err := wt.wallet.UnconfirmedBalance()
	if err != nil {
		t.Fatal(err)
	}
	if !startingOutgoing.IsZero() {
		t.Fatal(startingOutgoing)
	}
	if !startingIncoming.IsZero() {
		t.Fatal(startingIncoming)
	}

	chainCts := types.DefaultChainConstants()

	// Create several builders in parallel.
	var wg sync.WaitGroup
	funding := types.NewCurrency64(10e3).Mul(chainCts.CurrencyUnits.OneCoin)
	for i := 0; i < outputsDesired; i++ {
		wg.Add(1)
		go func() {
			// Create the builder and fund the transaction.
			builder := wt.wallet.StartTransaction()
			err = builder.FundCoins(funding)
			if err != nil {
				t.Fatal(err)
			}

			// Spend the transaction funds on miner fees and the void output.
			builder.AddMinerFee(types.NewCurrency64(25).Mul(chainCts.CurrencyUnits.OneCoin))
			output := types.CoinOutput{Value: types.NewCurrency64(9975).Mul(chainCts.CurrencyUnits.OneCoin)}
			builder.AddCoinOutput(output)
			// Sign the transactions and verify that both are valid.
			tset, err := builder.Sign()
			if err != nil {
				t.Fatal(err)
			}
			err = wt.tpool.AcceptTransactionSet(tset)
			if err != nil {
				t.Fatal(err)
			}
			wg.Done()
		}()
	}
	wg.Wait()

	//
	//	// Mine a block to get the transaction sets into the blockchain.
	//	err = wt.addBlockNoPayout()
	//	if err != nil {
	//		t.Fatal(err)
	//	}
	//

	// Check the final balance.
	endingSCConfirmed, _, err := wt.wallet.ConfirmedBalance()
	if err != nil {
		t.Fatal(err)
	}
	expected := startingSCConfirmed.Sub(funding.Mul(types.NewCurrency64(uint64(outputsDesired))))
	if !expected.Equals(endingSCConfirmed) {
		t.Fatal("did not get the expected ending balance", expected, endingSCConfirmed, startingSCConfirmed)
	}
}

*/
