package wallet

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/threefoldtech/rivine/build"
	"github.com/threefoldtech/rivine/modules"
	"github.com/threefoldtech/rivine/types"
)

// various errors returned by the wallet
var (
	ErrNilOutputs = errors.New("nil outputs cannot be send")
)

// sortedOutputs is a struct containing a slice of siacoin outputs and their
// corresponding ids. sortedOutputs can be sorted using the sort package.
type sortedOutputs struct {
	ids     []types.CoinOutputID
	outputs []types.CoinOutput
}

// ConfirmedBalance returns the balance of the wallet according to all of the
// confirmed transactions.
func (w *Wallet) ConfirmedBalance() (coinBalance types.Currency, blockstakeBalance types.Currency, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.unlocked {
		err = modules.ErrLockedWallet
		return
	}

	// prepare fulfillable context
	ctx := w.getFulfillableContextForLatestBlock()

	// get all coin and block stake stum
	for _, sco := range w.coinOutputs {
		if sco.Condition.Fulfillable(ctx) {
			coinBalance = coinBalance.Add(sco.Value)
		}
	}
	for _, sfo := range w.blockstakeOutputs {
		if sfo.Condition.Fulfillable(ctx) {
			blockstakeBalance = blockstakeBalance.Add(sfo.Value)
		}
	}
	return
}

// ConfirmedLockedBalance returns the locked balance of the wallet according to all of the
// confirmed transactions, which have locked outputs.
func (w *Wallet) ConfirmedLockedBalance() (coinBalance types.Currency, blockstakeBalance types.Currency, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.unlocked {
		err = modules.ErrLockedWallet
		return
	}

	// prepare fulfillable context
	ctx := w.getFulfillableContextForLatestBlock()

	// get all coin and block stake stum
	for _, sco := range w.coinOutputs {
		if !sco.Condition.Fulfillable(ctx) {
			coinBalance = coinBalance.Add(sco.Value)
		}
	}
	for _, sfo := range w.blockstakeOutputs {
		if !sfo.Condition.Fulfillable(ctx) {
			blockstakeBalance = blockstakeBalance.Add(sfo.Value)
		}
	}
	return
}

// UnspentBlockStakeOutputs returns the blockstake outputs where the beneficiary is an
// address this wallet has an unlockhash for.
func (w *Wallet) UnspentBlockStakeOutputs() (map[types.BlockStakeOutputID]types.BlockStakeOutput, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.unlocked {
		return nil, modules.ErrLockedWallet
	}

	// prepare fulfillable context
	ctx := w.getFulfillableContextForLatestBlock()

	// get all unspend block stake outputs, which are fulfillable
	outputs := make(map[types.BlockStakeOutputID]types.BlockStakeOutput, 0)
	for id := range w.blockstakeOutputs {
		output := w.blockstakeOutputs[id]
		if output.Condition.Fulfillable(ctx) {
			outputs[id] = output
		}
	}
	return outputs, nil
}

// UnconfirmedBalance returns the number of outgoing and incoming coins in
// the unconfirmed transaction set. Refund outputs are included in this
// reporting.
func (w *Wallet) UnconfirmedBalance() (outgoingCoins types.Currency, incomingCoins types.Currency, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.unlocked {
		err = modules.ErrLockedWallet
		return
	}

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

// MultiSigWallets returns all multisig wallets which contain at least one unlock hash owned by this wallet.
func (w *Wallet) MultiSigWallets() ([]modules.MultiSigWallet, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.unlocked {
		return nil, modules.ErrLockedWallet
	}

	wallets := make(map[types.UnlockHash]*modules.MultiSigWallet)

	ctx := w.getFulfillableContextForLatestBlock()

	var wallet *modules.MultiSigWallet
	var exists bool
	for id, co := range w.multiSigCoinOutputs {
		address := co.Condition.UnlockHash()
		// Check if the wallet exists
		if wallet, exists = wallets[address]; !exists {
			// get the internal multisig unlock condition
			unlockhashes, minSignatureCount := getMultisigConditionProperties(co.Condition.Condition)
			if len(unlockhashes) == 0 {
				w.log.Printf("[ERROR] failed to convert output to multisig condition: type=%T conditionType=%d",
					co.Condition.Condition, co.Condition.ConditionType())
				build.Critical("Failed to convert output to multisig condition")
				continue
			}
			// Create a new wallet for this address
			wallet = &modules.MultiSigWallet{
				Address: address,
				Owners:  unlockhashes,
				MinSigs: minSignatureCount,
			}
			wallets[address] = wallet
		}
		if !co.Condition.Fulfillable(ctx) {
			// Add the locked coins if applicable
			wallet.ConfirmedLockedCoinBalance = wallet.ConfirmedLockedCoinBalance.Add(co.Value)
		} else {
			// Add the coins to the unlocked balance
			wallet.ConfirmedCoinBalance = wallet.ConfirmedCoinBalance.Add(co.Value)
		}
		// Add the output ID
		wallet.CoinOutputIDs = append(wallet.CoinOutputIDs, id)
	}

	for id, bso := range w.multiSigBlockStakeOutputs {
		address := bso.Condition.UnlockHash()
		// Check if the wallet exists
		if wallet, exists = wallets[address]; !exists {
			// get the internal multisig unlock condition
			unlockhashes, minSignatureCount := getMultisigConditionProperties(bso.Condition.Condition)
			if len(unlockhashes) == 0 {
				w.log.Printf("[ERROR] failed to convert output to multisig condition: type=%T conditionType=%d",
					bso.Condition.Condition, bso.Condition.ConditionType())
				build.Severe("Failed to convert output to multisig condition")
				continue
			}
			// Create a new wallet for this address
			wallet = &modules.MultiSigWallet{
				Address: address,
				Owners:  unlockhashes,
				MinSigs: minSignatureCount,
			}
			wallets[address] = wallet
		}
		if !bso.Condition.Fulfillable(ctx) {
			// Add the locked block stakes if applicable
			wallet.ConfirmedLockedBlockStakeBalance = wallet.ConfirmedLockedBlockStakeBalance.Add(bso.Value)
		} else {
			// Add the block stakes to the confirmed balance
			wallet.ConfirmedBlockStakeBalance = wallet.ConfirmedBlockStakeBalance.Add(bso.Value)
		}
		// Add the output ID
		wallet.BlockStakeOutputIDs = append(wallet.BlockStakeOutputIDs, id)
	}

	// Check unconfrimed transactions
	for _, upt := range w.unconfirmedProcessedTransactions {
		for _, input := range upt.Inputs {
			if wallet, exists = wallets[input.RelatedAddress]; exists && input.FundType == types.SpecifierCoinInput {
				wallet.UnconfirmedOutgoingCoins = wallet.UnconfirmedOutgoingCoins.Add(input.Value)
			} else if exists && input.FundType == types.SpecifierBlockStakeInput {
				wallet.UnconfirmedOutgoingBlockStakes = wallet.UnconfirmedOutgoingBlockStakes.Add(input.Value)
			}
		}
		for _, output := range upt.Outputs {
			if wallet, exists = wallets[output.RelatedAddress]; exists && output.FundType == types.SpecifierCoinOutput {
				wallet.UnconfirmedIncomingCoins = wallet.UnconfirmedIncomingCoins.Add(output.Value)
			} else if exists && output.FundType == types.SpecifierBlockStakeOutput {
				wallet.UnconfirmedIncomingBlockStakes = wallet.UnconfirmedIncomingBlockStakes.Add(output.Value)
			}
		}
	}

	msws := make([]modules.MultiSigWallet, 0, len(wallets))
	for _, wallet := range wallets {
		msws = append(msws, *wallet)
	}

	return msws, nil
}

// SendCoins creates a transaction sending 'amount' to whoever can fulfill the condition. If data is provided,
// it is added as arbitrary data to the transaction. The transaction
// is submitted to the transaction pool and is also returned.
func (w *Wallet) SendCoins(amount types.Currency, cond types.UnlockConditionProxy, data []byte) (types.Transaction, error) {
	return w.SendOutputs([]types.CoinOutput{
		{
			Condition: cond,
			Value:     amount,
		},
	}, nil, nil, nil, false)
}

// SendBlockStakes creates a transaction sending 'amount' to whoever can fulfill the condition. The transaction
// is submitted to the transaction pool and is also returned.
func (w *Wallet) SendBlockStakes(amount types.Currency, cond types.UnlockConditionProxy) (types.Transaction, error) {
	return w.SendOutputs(nil, []types.BlockStakeOutput{
		{
			Condition: cond,
			Value:     amount,
		},
	}, nil, nil, false)
}

// SendOutputs is a tool for sending coins and block stakes from the wallet, to one or multiple addreses.
// The transaction is automatically given to the transaction pool, and is also returned to the caller.
func (w *Wallet) SendOutputs(coinOutputs []types.CoinOutput, blockstakeOutputs []types.BlockStakeOutput, data []byte, refundAddress *types.UnlockHash, reuseRefundAddress bool) (types.Transaction, error) {
	if len(coinOutputs) == 0 && len(blockstakeOutputs) == 0 {
		// at least one coin output OR one block stake output has to be send
		return types.Transaction{}, ErrNilOutputs
	}

	if err := w.tg.Add(); err != nil {
		return types.Transaction{}, err
	}
	defer w.tg.Done()

	tpoolFee := w.chainCts.MinimumTransactionFee.Mul64(1) // TODO better fee algo
	totalAmount := types.NewCurrency64(0).Add(tpoolFee)
	var err error
	txnBuilder := w.StartTransaction()
	// Make sure to release inputs in case of an error
	defer func() {
		if err != nil {
			txnBuilder.Drop()
		}
	}()
	for _, co := range coinOutputs {
		txnBuilder.AddCoinOutput(co)
		totalAmount = totalAmount.Add(co.Value)
	}
	err = txnBuilder.FundCoins(totalAmount, refundAddress, reuseRefundAddress)
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
		err = txnBuilder.FundBlockStakes(totalAmount, refundAddress, reuseRefundAddress)
		if err != nil {
			return types.Transaction{}, err
		}
	}
	if len(data) != 0 {
		txnBuilder.SetArbitraryData(data)
	}
	var txnSet []types.Transaction
	txnSet, err = txnBuilder.Sign()
	if err != nil {
		return types.Transaction{}, err
	}
	if len(txnSet) == 0 {
		build.Severe(fmt.Errorf("unexpected txnSet length: " + strconv.Itoa(len(txnSet))))
	}
	err = w.tpool.AcceptTransactionSet(txnSet)
	if err != nil {
		return types.Transaction{}, err
	}
	return txnSet[0], nil
}

// Len returns the number of elements in the sortedOutputs struct.
func (so sortedOutputs) Len() int {
	if len(so.ids) != len(so.outputs) {
		build.Severe("sortedOutputs object is corrupt")
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
