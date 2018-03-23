package wallet

import (
	"errors"
	"sort"

	"github.com/rivine/rivine/encoding"
	"github.com/rivine/rivine/modules"
	"github.com/rivine/rivine/types"
)

var (
	// errBuilderAlreadySigned indicates that the transaction builder has
	// already added at least one successful signature to the transaction,
	// meaning that future calls to Sign will result in an invalid transaction.
	errBuilderAlreadySigned = errors.New("sign has already been called on this transaction builder, multiple calls can cause issues")
)

// transactionBuilder allows transactions to be manually constructed, including
// the ability to fund transactions with siacoins and blockstakes from the wallet.
type transactionBuilder struct {
	// 'signed' indicates that at least one transaction signature has been
	// added to the wallet, meaning that future calls to 'Sign' will fail.
	parents     []types.Transaction
	signed      bool
	transaction types.Transaction

	newParents       []int
	coinInputs       []int
	blockstakeInputs []int

	wallet *Wallet
}

// FundCoins will add a siacoin input of exactly 'amount' to the
// transaction. The coin input will not be signed until 'Sign' is called
// on the transaction builder.
func (tb *transactionBuilder) FundCoins(amount types.Currency) error {
	tb.wallet.mu.Lock()
	defer tb.wallet.mu.Unlock()

	// Collect a value-sorted set of siacoin outputs.
	var so sortedOutputs
	for scoid, sco := range tb.wallet.coinOutputs {
		so.ids = append(so.ids, scoid)
		so.outputs = append(so.outputs, sco)
	}
	// Add all of the unconfirmed outputs as well.
	for _, upt := range tb.wallet.unconfirmedProcessedTransactions {
		for i, sco := range upt.Transaction.CoinOutputs {
			// Determine if the output belongs to the wallet.
			_, exists := tb.wallet.keys[sco.UnlockHash]
			if !exists {
				continue
			}
			so.ids = append(so.ids, upt.Transaction.CoinOutputID(uint64(i)))
			so.outputs = append(so.outputs, sco)
		}
	}
	sort.Sort(sort.Reverse(so))

	// Create a transaction that will add the correct amount of siacoins to the
	// transaction.
	var fund types.Currency
	// potentialFund tracks the balance of the wallet including outputs that
	// have been spent in other unconfirmed transactions recently. This is to
	// provide the user with a more useful error message in the event that they
	// are overspending.
	var potentialFund types.Currency
	var spentScoids []types.CoinOutputID
	for i := range so.ids {
		scoid := so.ids[i]
		sco := so.outputs[i]
		// Check that this output has not recently been spent by the wallet.
		spendHeight := tb.wallet.spentOutputs[types.OutputID(scoid)]
		// Prevent an underflow error.
		allowedHeight := tb.wallet.consensusSetHeight - RespendTimeout
		if tb.wallet.consensusSetHeight < RespendTimeout {
			allowedHeight = 0
		}
		if spendHeight > allowedHeight {
			potentialFund = potentialFund.Add(sco.Value)
			continue
		}

		// Add a coin input for this output.
		sci := types.CoinInput{
			ParentID: scoid,
			Unlocker: types.NewSingleSignatureInputLock(
				types.Ed25519PublicKey(tb.wallet.keys[sco.UnlockHash].PublicKey)),
		}

		tb.coinInputs = append(tb.coinInputs, len(tb.transaction.CoinInputs))
		tb.transaction.CoinInputs = append(tb.transaction.CoinInputs, sci)

		spentScoids = append(spentScoids, scoid)

		// Add the output to the total fund
		fund = fund.Add(sco.Value)
		potentialFund = potentialFund.Add(sco.Value)
		if fund.Cmp(amount) >= 0 {
			break
		}
	}
	if potentialFund.Cmp(amount) >= 0 && fund.Cmp(amount) < 0 {
		return modules.ErrIncompleteTransactions
	}
	if fund.Cmp(amount) < 0 {
		return modules.ErrLowBalance
	}

	// Create a refund output if needed.
	if !amount.Equals(fund) {
		refundUnlockHash, err := tb.wallet.nextPrimarySeedAddress()
		if err != nil {
			return err
		}
		refundOutput := types.CoinOutput{
			Value:      fund.Sub(amount),
			UnlockHash: refundUnlockHash,
		}
		tb.transaction.CoinOutputs = append(tb.transaction.CoinOutputs, refundOutput)
	}

	// Mark all outputs that were spent as spent.
	for _, scoid := range spentScoids {
		tb.wallet.spentOutputs[types.OutputID(scoid)] = tb.wallet.consensusSetHeight
	}
	return nil
}

// FundBlockStakes will add a blockstake input of exaclty 'amount' to the
// transaction. The blockstake input will not be signed until 'Sign' is called
// on the transaction builder.
func (tb *transactionBuilder) FundBlockStakes(amount types.Currency) error {
	tb.wallet.mu.Lock()
	defer tb.wallet.mu.Unlock()

	// Create a transaction that will add the correct amount of siafunds to the
	// transaction.
	var fund types.Currency
	var potentialFund types.Currency
	var spentSfoids []types.BlockStakeOutputID
	for sfoid, sfo := range tb.wallet.blockstakeOutputs {
		// Check that this output has not recently been spent by the wallet.
		spendHeight := tb.wallet.spentOutputs[types.OutputID(sfoid)]
		// Prevent an underflow error.
		allowedHeight := tb.wallet.consensusSetHeight - RespendTimeout
		if tb.wallet.consensusSetHeight < RespendTimeout {
			allowedHeight = 0
		}
		if spendHeight > allowedHeight {
			potentialFund = potentialFund.Add(sfo.Value)
			continue
		}

		sfi := types.BlockStakeInput{
			ParentID: sfoid,
			Unlocker: types.NewSingleSignatureInputLock(
				types.Ed25519PublicKey(tb.wallet.keys[sfo.UnlockHash].PublicKey)),
		}
		tb.blockstakeInputs = append(tb.blockstakeInputs, len(tb.transaction.BlockStakeInputs))
		tb.transaction.BlockStakeInputs = append(tb.transaction.BlockStakeInputs, sfi)

		spentSfoids = append(spentSfoids, sfoid)

		// Add the output to the total fund
		fund = fund.Add(sfo.Value)
		potentialFund = potentialFund.Add(sfo.Value)
		if fund.Cmp(amount) >= 0 {
			break
		}
	}
	if potentialFund.Cmp(amount) >= 0 && fund.Cmp(amount) < 0 {
		return modules.ErrIncompleteTransactions
	}
	if fund.Cmp(amount) < 0 {
		return modules.ErrLowBalance
	}

	// Create a refund output if needed.
	if !amount.Equals(fund) {
		refundUnlockHash, err := tb.wallet.nextPrimarySeedAddress()
		if err != nil {
			return err
		}
		refundOutput := types.BlockStakeOutput{
			Value:      fund.Sub(amount),
			UnlockHash: refundUnlockHash,
		}
		tb.transaction.BlockStakeOutputs = append(tb.transaction.BlockStakeOutputs, refundOutput)
	}

	// Mark all outputs that were spent as spent.
	for _, sfoid := range spentSfoids {
		tb.wallet.spentOutputs[types.OutputID(sfoid)] = tb.wallet.consensusSetHeight
	}
	return nil
}

// AddParents adds a set of parents to the transaction.
func (tb *transactionBuilder) AddParents(newParents []types.Transaction) {
	tb.parents = append(tb.parents, newParents...)
}

// AddMinerFee adds a miner fee to the transaction, returning the index of the
// miner fee within the transaction.
func (tb *transactionBuilder) AddMinerFee(fee types.Currency) uint64 {
	tb.transaction.MinerFees = append(tb.transaction.MinerFees, fee)
	return uint64(len(tb.transaction.MinerFees) - 1)
}

// AddCoinInput adds a siacoin input to the transaction, returning the index
// of the coin input within the transaction. When 'Sign' gets called, this
// input will be left unsigned.
func (tb *transactionBuilder) AddCoinInput(input types.CoinInput) uint64 {
	tb.transaction.CoinInputs = append(tb.transaction.CoinInputs, input)
	return uint64(len(tb.transaction.CoinInputs) - 1)
}

// AddCoinOutput adds a siacoin output to the transaction, returning the
// index of the siacoin output within the transaction.
func (tb *transactionBuilder) AddCoinOutput(output types.CoinOutput) uint64 {
	tb.transaction.CoinOutputs = append(tb.transaction.CoinOutputs, output)
	return uint64(len(tb.transaction.CoinOutputs) - 1)
}

// AddBlockStakeInput adds a blockstake input to the transaction, returning the index
// of the blockstake input within the transaction. When 'Sign' is called, this
// input will be left unsigned.
func (tb *transactionBuilder) AddBlockStakeInput(input types.BlockStakeInput) uint64 {
	tb.transaction.BlockStakeInputs = append(tb.transaction.BlockStakeInputs, input)
	return uint64(len(tb.transaction.BlockStakeInputs) - 1)
}

// SpendBlockStake will link the unspent block stake to the transaction as an input.
// In contrast with FundBlockStakes, this function will not loop over all unspent
// block stake output. the ubsoid is an argument. The blockstake input will not be
// signed until 'Sign' is called on the transaction builder.
func (tb *transactionBuilder) SpendBlockStake(ubsoid types.BlockStakeOutputID) error {
	tb.wallet.mu.Lock()
	defer tb.wallet.mu.Unlock()

	ubso, ok := tb.wallet.unspentblockstakeoutputs[ubsoid]
	if !ok {
		return modules.ErrIncompleteTransactions //TODO: not right error
	}

	bsi := types.BlockStakeInput{
		ParentID: ubsoid,
		Unlocker: types.NewSingleSignatureInputLock(
			types.Ed25519PublicKey(tb.wallet.keys[ubso.UnlockHash].PublicKey)),
	}
	tb.blockstakeInputs = append(tb.blockstakeInputs, len(tb.transaction.BlockStakeInputs))
	tb.transaction.BlockStakeInputs = append(tb.transaction.BlockStakeInputs, bsi)

	// Mark output as spent.
	tb.wallet.spentOutputs[types.OutputID(ubsoid)] = tb.wallet.consensusSetHeight
	return nil
}

// AddBlockStakeOutput adds a blockstake output to the transaction, returning the
// index of the blockstake output within the transaction.
func (tb *transactionBuilder) AddBlockStakeOutput(output types.BlockStakeOutput) uint64 {
	tb.transaction.BlockStakeOutputs = append(tb.transaction.BlockStakeOutputs, output)
	return uint64(len(tb.transaction.BlockStakeOutputs) - 1)
}

// AddArbitraryData sets the arbitrary data of the transaction.
func (tb *transactionBuilder) SetArbitraryData(arb []byte) {
	tb.transaction.ArbitraryData = arb
}

// Drop discards all of the outputs in a transaction, returning them to the
// pool so that other transactions may use them. 'Drop' should only be called
// if a transaction is both unsigned and will not be used any further.
func (tb *transactionBuilder) Drop() {
	tb.wallet.mu.Lock()
	defer tb.wallet.mu.Unlock()

	// Iterate through all parents and the transaction itself and restore all
	// outputs to the list of available outputs.
	txns := append(tb.parents, tb.transaction)
	for _, txn := range txns {
		for _, sci := range txn.CoinInputs {
			delete(tb.wallet.spentOutputs, types.OutputID(sci.ParentID))
		}
	}

	tb.parents = nil
	tb.signed = false
	tb.transaction = types.Transaction{}

	tb.newParents = nil
	tb.coinInputs = nil
	tb.blockstakeInputs = nil
}

// Sign will sign any inputs added by 'FundSiacoins' or 'FundSiafunds' and
// return a transaction set that contains all parents prepended to the
// transaction. If more fields need to be added, a new transaction builder will
// need to be created.
//
// If the whole transaction flag is set to true, then the whole transaction
// flag will be set in the covered fields object. If the whole transaction flag
// is set to false, then the covered fields object will cover all fields that
// have already been added to the transaction, but will also leave room for
// more fields to be added.
//
// Sign should not be called more than once. If, for some reason, there is an
// error while calling Sign, the builder should be dropped.
func (tb *transactionBuilder) Sign() ([]types.Transaction, error) {
	if tb.signed {
		return nil, errBuilderAlreadySigned
	}

	// For each siacoin input in the transaction that we added, provide a
	// signature.
	tb.wallet.mu.Lock()
	defer tb.wallet.mu.Unlock()
	for _, inputIndex := range tb.coinInputs {
		input := tb.transaction.CoinInputs[inputIndex]
		key := tb.wallet.keys[input.Unlocker.UnlockHash()]
		err := input.Unlocker.Lock(uint64(inputIndex), tb.transaction, key.SecretKey)
		if err != nil {
			return nil, err
		}
		tb.signed = true // Signed is set to true after one successful signature to indicate that future signings can cause issues.
	}
	for _, inputIndex := range tb.blockstakeInputs {
		input := tb.transaction.BlockStakeInputs[inputIndex]
		key := tb.wallet.keys[input.Unlocker.UnlockHash()]
		err := input.Unlocker.Lock(uint64(inputIndex), tb.transaction, key.SecretKey)
		if err != nil {
			return nil, err
		}
		tb.signed = true // Signed is set to true after one successful signature to indicate that future signings can cause issues.
	}

	// Get the transaction set and delete the transaction from the registry.
	txnSet := append(tb.parents, tb.transaction)
	return txnSet, nil
}

// ViewTransaction returns a transaction-in-progress along with all of its
// parents, specified by id. An error is returned if the id is invalid.  Note
// that ids become invalid for a transaction after 'SignTransaction' has been
// called because the transaction gets deleted.
func (tb *transactionBuilder) View() (types.Transaction, []types.Transaction) {
	return tb.transaction, tb.parents
}

// ViewAdded returns all of the siacoin inputs, siafund inputs, and parent
// transactions that have been automatically added by the builder.
func (tb *transactionBuilder) ViewAdded() (newParents, coinInputs, blockstakeInputs []int) {
	return tb.newParents, tb.coinInputs, tb.blockstakeInputs
}

// RegisterTransaction takes a transaction and its parents and returns a
// transactionBuilder which can be used to expand the transaction. The most
// typical call is 'RegisterTransaction(types.Transaction{}, nil)', which
// registers a new transaction without parents.
func (w *Wallet) RegisterTransaction(t types.Transaction, parents []types.Transaction) modules.TransactionBuilder {
	// Create a deep copy of the transaction and parents by encoding them. A
	// deep copy ensures that there are no pointer or slice related errors -
	// the builder will be working directly on the transaction, and the
	// transaction may be in use elsewhere (in this case, the host is using the
	// transaction.
	pBytes := encoding.Marshal(parents)
	var pCopy []types.Transaction
	err := encoding.Unmarshal(pBytes, &pCopy)
	if err != nil {
		panic(err)
	}
	tBytes := encoding.Marshal(t)
	var tCopy types.Transaction
	err = encoding.Unmarshal(tBytes, &tCopy)
	if err != nil {
		panic(err)
	}
	return &transactionBuilder{
		parents:     pCopy,
		transaction: tCopy,
		wallet:      w,
	}
}

// StartTransaction is a convenience function that calls
// RegisterTransaction(types.Transaction{}, nil).
func (w *Wallet) StartTransaction() modules.TransactionBuilder {
	return w.RegisterTransaction(types.Transaction{}, nil)
}
