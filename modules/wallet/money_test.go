package wallet

import (
	"sort"
	"testing"

	"github.com/threefoldtech/rivine/crypto"
	"github.com/threefoldtech/rivine/modules"
	"github.com/threefoldtech/rivine/types"
)

// TestSendCoins probes the SendCoins method of the wallet.
func TestSendCoins(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	cs := newConsensusSetStub()
	wt, err := createWalletTesterWithStubCS(t.Name(), cs)
	if err != nil {
		t.Fatal(err)
	}
	defer wt.closeWt()

	// Get the initial balance - should be 1 block. The unconfirmed balances
	// should be 0.
	confirmedBal, _, err := wt.wallet.ConfirmedBalance()
	if err != nil {
		t.Fatal(err)
	}
	unconfirmedOut, unconfirmedIn, err := wt.wallet.UnconfirmedBalance()
	if err != nil {
		t.Fatal(err)
	}
	if !confirmedBal.Equals64(0) {
		t.Error("unexpected confirmed balance")
	}
	if !unconfirmedOut.Equals64(0) {
		t.Error("unconfirmed balance should be 0")
	}
	if !unconfirmedIn.Equals64(0) {
		t.Error("unconfirmed balance should be 0")
	}

	// sending coins requires funds to be send
	_, err = wt.wallet.SendCoins(types.NewCurrency64(5000), types.NewCondition(nil), nil)
	if err != modules.ErrLowBalance {
		t.Fatal(err)
	}

	// give wallet some money to spend
	addr, err := wt.wallet.NextAddress()
	if err != nil {
		t.Fatal(err)
	}
	cs.addTransactionAsBlock(addr,
		wt.wallet.chainCts.MinimumTransactionFee.Mul64(1).Add(types.NewCurrency64(5000)))

	// Send 5000 hastings. The wallet will automatically add a fee. Outgoing
	// unconfirmed siacoins - incoming unconfirmed coins should equal 5000 +
	// fee.
	tpoolFee := wt.wallet.chainCts.MinimumTransactionFee.Mul64(1)
	_, err = wt.wallet.SendCoins(types.NewCurrency64(5000), types.NewCondition(nil), nil)
	if err != nil {
		t.Fatal(err)
	}
	confirmedBal2, _, err := wt.wallet.ConfirmedBalance()
	if err != nil {
		t.Fatal(err)
	}

	if confirmedBal2.Cmp(types.NewCurrency64(5000).Add(tpoolFee)) != 0 {
		t.Error("sending cioins appears to have been send")
	}
}

// TestIntegrationSendOverUnder sends too many coins, resulting in an error,
// followed by sending few enough coins that the send should complete.
//
// This test is here because of a bug found in production where the wallet
// would mark outputs as spent before it knew that there was enough money  to
// complete the transaction. This meant that, after trying to send too many
// coins, all outputs got marked 'sent'. This test reproduces those conditions
// to ensure it does not happen again.
func TestIntegrationSendOverUnder(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	cs := newConsensusSetStub()
	wt, err := createWalletTesterWithStubCS(t.Name(), cs)
	if err != nil {
		t.Fatal(err)
	}
	defer wt.closeWt()

	// Spend too many coins.
	tooManyCoins := wt.wallet.chainCts.CurrencyUnits.OneCoin.Mul64(1e12)
	_, err = wt.wallet.SendCoins(tooManyCoins, types.NewCondition(nil), nil)
	if err != modules.ErrLowBalance {
		t.Error("low balance err not returned after attempting to send too many coins")
	}
	reasonableCoins := wt.wallet.chainCts.CurrencyUnits.OneCoin.Mul64(100e3)

	addr, err := wt.wallet.NextAddress()
	if err != nil {
		t.Fatal(err)
	}
	err = cs.addTransactionAsBlock(addr,
		wt.wallet.chainCts.CurrencyUnits.OneCoin.Mul64(1).Add(reasonableCoins))
	if err != nil {
		t.Fatal(err)
	}

	// Spend a reasonable amount of coins.
	_, err = wt.wallet.SendCoins(reasonableCoins, types.NewCondition(nil), nil)
	if err != nil {
		t.Error("unexpected error: ", err)
	}
}

// TestIntegrationSpendHalfHalf spends more than half of the coins, and then
// more than half of the coins again, to make sure that the wallet is not
// reusing outputs that it has already spent.
func TestIntegrationSpendHalfHalf(t *testing.T) {
	//if testing.Short() {
	t.SkipNow()
	//}
	wt, err := createWalletTester(t.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer wt.closeWt()

	// Spend more than half of the coins twice.
	halfPlus := wt.wallet.chainCts.CurrencyUnits.OneCoin.Mul64(200e3)
	_, err = wt.wallet.SendCoins(halfPlus, types.NewCondition(nil), nil)
	if err != nil {
		t.Error("unexpected error: ", err)
	}
	_, err = wt.wallet.SendCoins(halfPlus,
		types.NewCondition(types.NewUnlockHashCondition(types.NewUnlockHash(0, crypto.Hash{1}))),
		nil)
	if err != modules.ErrIncompleteTransactions {
		t.Error("wallet appears to be reusing outputs when building transactions: ", err)
	}
}

// TestIntegrationSpendUnconfirmed spends an unconfirmed siacoin output.
func TestIntegrationSpendUnconfirmed(t *testing.T) {
	//if testing.Short() {
	t.SkipNow()
	//}
	wt, err := createWalletTester(t.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer wt.closeWt()

	// Spend the only output.
	halfPlus := wt.wallet.chainCts.CurrencyUnits.OneCoin.Mul64(200e3)
	_, err = wt.wallet.SendCoins(halfPlus, types.NewCondition(nil), nil)
	if err != nil {
		t.Error("unexpected error: ", err)
	}
	someMore := wt.wallet.chainCts.CurrencyUnits.OneCoin.Mul64(75e3)
	_, err = wt.wallet.SendCoins(someMore,
		types.NewCondition(types.NewUnlockHashCondition(types.NewUnlockHash(0, crypto.Hash{1}))),
		nil)
	if err != nil {
		t.Error("wallet appears to be struggling to spend unconfirmed outputs")
	}
}

// TestIntegrationSortedOutputsSorting checks that the outputs are being correctly sorted
// by the currency value.
func TestIntegrationSortedOutputsSorting(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	so := sortedOutputs{
		ids: []types.CoinOutputID{{0}, {1}, {2}, {3}, {4}, {5}, {6}, {7}},
		outputs: []types.CoinOutput{
			{Value: types.NewCurrency64(2)},
			{Value: types.NewCurrency64(3)},
			{Value: types.NewCurrency64(4)},
			{Value: types.NewCurrency64(7)},
			{Value: types.NewCurrency64(6)},
			{Value: types.NewCurrency64(0)},
			{Value: types.NewCurrency64(1)},
			{Value: types.NewCurrency64(5)},
		},
	}
	sort.Sort(so)

	expectedIDSorting := []types.CoinOutputID{{5}, {6}, {0}, {1}, {2}, {7}, {4}, {3}}
	for i := uint64(0); i < 8; i++ {
		if so.ids[i] != expectedIDSorting[i] {
			t.Error("an id is out of place: ", i)
		}
		if !so.outputs[i].Value.Equals64(i) {
			t.Error("a value is out of place: ", i)
		}
	}
}

// Test to confirm that a call with nil outputs (no outputs),
// results in the ErrNilOutputs error.
// Tests the solution for: https://github.com/threefoldtech/rivine/issues/327
func TestNilOutputs(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	cs := newConsensusSetStub()
	wt, err := createWalletTesterWithStubCS(t.Name(), cs)
	if err != nil {
		t.Fatal(err)
	}
	defer wt.closeWt()

	_, err = wt.wallet.SendOutputs(nil, nil, []byte("data"), nil, false)
	if err != ErrNilOutputs {
		t.Fatal("expected ErrNilOutput, but receiver: ", err)
	}
}
