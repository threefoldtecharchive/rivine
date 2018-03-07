package wallet

import (
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
func (w *Wallet) SendCoins(amount types.Currency, dest types.UnlockHash, data []byte) ([]types.Transaction, error) {
	if err := w.tg.Add(); err != nil {
		return nil, err
	}
	defer w.tg.Done()

	tpoolFee := types.OneCoin.Mul64(1) // TODO: better fee algo.
	output := types.CoinOutput{
		Value:      amount,
		UnlockHash: dest,
	}

	txnBuilder := w.StartTransaction()
	err := txnBuilder.FundCoins(amount.Add(tpoolFee))
	if err != nil {
		return nil, err
	}
	txnBuilder.AddMinerFee(tpoolFee)
	txnBuilder.AddCoinOutput(output)
	if data != nil {
		txnBuilder.AddArbitraryData(data)
	}
	txnSet, err := txnBuilder.Sign(true)
	if err != nil {
		return nil, err
	}
	err = w.tpool.AcceptTransactionSet(txnSet)
	if err != nil {
		return nil, err
	}
	return txnSet, nil
}

// SendBlockStakes creates a transaction sending 'amount' to 'dest'. The transaction
// is submitted to the transaction pool and is also returned.
func (w *Wallet) SendBlockStakes(amount types.Currency, dest types.UnlockHash) ([]types.Transaction, error) {
	if err := w.tg.Add(); err != nil {
		return nil, err
	}
	defer w.tg.Done()
	tpoolFee := types.OneCoin.Mul64(1) // TODO: better fee algo.
	output := types.BlockStakeOutput{
		Value:      amount,
		UnlockHash: dest,
	}

	txnBuilder := w.StartTransaction()
	err := txnBuilder.FundCoins(tpoolFee)
	if err != nil {
		return nil, err
	}
	err = txnBuilder.FundBlockStakes(amount)
	if err != nil {
		return nil, err
	}
	txnBuilder.AddMinerFee(tpoolFee)
	txnBuilder.AddBlockStakeOutput(output)
	txnSet, err := txnBuilder.Sign(true)
	if err != nil {
		return nil, err
	}
	err = w.tpool.AcceptTransactionSet(txnSet)
	if err != nil {
		return nil, err
	}
	return txnSet, nil
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
