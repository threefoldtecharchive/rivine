package wallet

import (
	"strconv"

	"github.com/rivine/rivine/build"
	"github.com/rivine/rivine/types"
)

// sortedOutputs is a struct containing a slice of siacoin outputs and their
// corresponding ids. sortedOutputs can be sorted using the sort package.
type sortedOutputs struct {
	ids     []types.CoinOutputID
	outputs []types.CoinOutput
}

// ConfirmedBalance returns the balance of the wallet according to all of the
// confirmed transactions.
func (w *Wallet) ConfirmedBalance() (coinBalance types.Currency, blockstakeBalance types.Currency) {
	w.mu.Lock()
	defer w.mu.Unlock()

	for _, sco := range w.coinOutputs {
		coinBalance = coinBalance.Add(sco.Value)
	}
	for _, sfo := range w.blockstakeOutputs {
		blockstakeBalance = blockstakeBalance.Add(sfo.Value)
	}
	return
}

// UnspentBlockStakeOutputs returns the blockstake outputs where the beneficiary is an
// address this wallet has an unlockhash for.
func (w *Wallet) UnspentBlockStakeOutputs() map[types.BlockStakeOutputID]types.BlockStakeOutput {
	//TODO: think about returning a copy
	return w.blockstakeOutputs
}

// UnconfirmedBalance returns the number of outgoing and incoming coins in
// the unconfirmed transaction set. Refund outputs are included in this
// reporting.
func (w *Wallet) UnconfirmedBalance() (outgoingCoins types.Currency, incomingCoins types.Currency) {
	w.mu.Lock()
	defer w.mu.Unlock()

	for _, upt := range w.unconfirmedProcessedTransactions {
		for _, input := range upt.Inputs {
			if input.FundType == types.SpecifierCoinInput && input.WalletAddress {
				outgoingCoins = outgoingCoins.Add(input.Value)
			}
		}
		for _, output := range upt.Outputs {
			if output.FundType == types.SpecifierCoinOutput && output.WalletAddress {
				incomingCoins = incomingCoins.Add(output.Value)
			}
		}
	}
	return
}

// SendCoins creates a transaction sending 'amount' to 'dest'. If data is provided,
// it is added as arbitrary data to the transaction. The transaction
// is submitted to the transaction pool and is also returned.
func (w *Wallet) SendCoins(amount types.Currency, dest types.UnlockHash, data []byte) (types.Transaction, error) {
	return w.SendOutputs([]types.CoinOutput{
		{
			UnlockHash: dest,
			Value:      amount,
		},
	}, nil, data)
}

// SendBlockStakes creates a transaction sending 'amount' to 'dest'. The transaction
// is submitted to the transaction pool and is also returned.
func (w *Wallet) SendBlockStakes(amount types.Currency, dest types.UnlockHash) (types.Transaction, error) {
	return w.SendOutputs(nil, []types.BlockStakeOutput{
		{
			UnlockHash: dest,
			Value:      amount,
		},
	}, nil)
}

// SendOutputs is a tool for sending coins and block stakes from the wallet, to one or multiple addreses.
// The transaction is automatically given to the transaction pool, and is also returned to the caller.
func (w *Wallet) SendOutputs(coinOutputs []types.CoinOutput, blockstakeOutputs []types.BlockStakeOutput, data []byte) (types.Transaction, error) {
	if err := w.tg.Add(); err != nil {
		return types.Transaction{}, err
	}
	defer w.tg.Done()

	tpoolFee := w.chainCts.MinimumTransactionFee.Mul64(1) // TODO better fee algo
	totalAmount := types.NewCurrency64(0).Add(tpoolFee)
	txnBuilder := w.StartTransaction()
	for _, co := range coinOutputs {
		txnBuilder.AddCoinOutput(co)
		totalAmount = totalAmount.Add(co.Value)
	}
	err := txnBuilder.FundCoins(totalAmount)
	if err != nil {
		return types.Transaction{}, err
	}
	txnBuilder.AddMinerFee(tpoolFee)
	totalAmount = types.NewCurrency64(0)
	for _, bso := range blockstakeOutputs {
		txnBuilder.AddBlockStakeOutput(bso)
		totalAmount = totalAmount.Add(bso.Value)
	}
	if !totalAmount.Equals64(0) {
		err = txnBuilder.FundBlockStakes(totalAmount)
		if err != nil {
			return types.Transaction{}, err
		}
	}
	if len(data) != 0 {
		txnBuilder.SetArbitraryData(data)
	}
	txnSet, err := txnBuilder.Sign()
	if err != nil {
		return types.Transaction{}, err
	}
	if len(txnSet) == 0 {
		panic("unexpected txnSet length: " + strconv.Itoa(len(txnSet)))
	}
	err = w.tpool.AcceptTransactionSet(txnSet)
	if err != nil {
		return types.Transaction{}, err
	}
	return txnSet[0], nil
}

// Len returns the number of elements in the sortedOutputs struct.
func (so sortedOutputs) Len() int {
	if build.DEBUG && len(so.ids) != len(so.outputs) {
		panic("sortedOutputs object is corrupt")
	}
	return len(so.ids)
}

// Less returns whether element 'i' is less than element 'j'. The currency
// value of each output is used for comparison.
func (so sortedOutputs) Less(i, j int) bool {
	return so.outputs[i].Value.Cmp(so.outputs[j].Value) < 0
}

// Swap swaps two elements in the sortedOutputs set.
func (so sortedOutputs) Swap(i, j int) {
	so.ids[i], so.ids[j] = so.ids[j], so.ids[i]
	so.outputs[i], so.outputs[j] = so.outputs[j], so.outputs[i]
}
