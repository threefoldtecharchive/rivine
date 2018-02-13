package wallet

import (
	"bytes"
	"errors"
	"sort"

	"github.com/rivine/rivine/crypto"
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

	newParents            []int
	coinInputs            []int
	blockstakeInputs      []int
	transactionSignatures []int

	wallet *Wallet
}

// addSignatures will sign a transaction using a spendable key, with support
// for multisig spendable keys. Because of the restricted input, the function
// is compatible with both coin inputs and blockstake inputs.
func addSignatures(txn *types.Transaction, cf types.CoveredFields, uc types.UnlockConditions, parentID crypto.Hash, spendKey spendableKey) (newSigIndices []int, err error) {
	// Try to find the matching secret key for each public key - some public
	// keys may not have a match. Some secret keys may be used multiple times,
	// which is why public keys are used as the outer loop.
	totalSignatures := uint64(0)
	for i, siaPubKey := range uc.PublicKeys {
		// Search for the matching secret key to the public key.
		for j := range spendKey.SecretKeys {
			pubKey := spendKey.SecretKeys[j].PublicKey()
			if bytes.Compare(siaPubKey.Key, pubKey[:]) != 0 {
				continue
			}

			// Found the right secret key, add a signature.
			sig := types.TransactionSignature{
				ParentID:       parentID,
				CoveredFields:  cf,
				PublicKeyIndex: uint64(i),
			}
			newSigIndices = append(newSigIndices, len(txn.TransactionSignatures))
			txn.TransactionSignatures = append(txn.TransactionSignatures, sig)
			sigIndex := len(txn.TransactionSignatures) - 1
			sigHash := txn.SigHash(sigIndex)
			encodedSig := crypto.SignHash(sigHash, spendKey.SecretKeys[j])
			txn.TransactionSignatures[sigIndex].Signature = encodedSig[:]

			// Count that the signature has been added, and break out of the
			// secret key loop.
			totalSignatures++
			break
		}

		// If there are enough signatures to satisfy the unlock conditions,
		// break out of the outer loop.
		if totalSignatures == uc.SignaturesRequired {
			break
		}
	}
	return newSigIndices, nil
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
		outputUnlockConditions := tb.wallet.keys[sco.UnlockHash].UnlockConditions
		if tb.wallet.consensusSetHeight < outputUnlockConditions.Timelock {
			continue
		}

		// Add a coin input for this output.
		sci := types.CoinInput{
			ParentID:         scoid,
			UnlockConditions: outputUnlockConditions,
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
	if amount.Cmp(fund) != 0 {
		refundUnlockConditions, err := tb.wallet.nextPrimarySeedAddress()
		if err != nil {
			return err
		}
		refundOutput := types.CoinOutput{
			Value:      fund.Sub(amount),
			UnlockHash: refundUnlockConditions.UnlockHash(),
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
		outputUnlockConditions := tb.wallet.keys[sfo.UnlockHash].UnlockConditions
		if tb.wallet.consensusSetHeight < outputUnlockConditions.Timelock {
			continue
		}

		sfi := types.BlockStakeInput{
			ParentID:         sfoid,
			UnlockConditions: outputUnlockConditions,
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
	if amount.Cmp(fund) != 0 {
		refundUnlockConditions, err := tb.wallet.nextPrimarySeedAddress()
		if err != nil {
			return err
		}
		refundOutput := types.BlockStakeOutput{
			Value:      fund.Sub(amount),
			UnlockHash: refundUnlockConditions.UnlockHash(),
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

	// Check that this output has no Timelock
	outputUnlockConditions := tb.wallet.keys[ubso.UnlockHash].UnlockConditions
	if tb.wallet.consensusSetHeight < outputUnlockConditions.Timelock {
		return modules.ErrIncompleteTransactions //TODO: not right error
	}

	bsi := types.BlockStakeInput{
		ParentID:         ubsoid,
		UnlockConditions: outputUnlockConditions,
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

// AddArbitraryData adds arbitrary data to the transaction, returning the index
// of the data within the transaction.
func (tb *transactionBuilder) AddArbitraryData(arb []byte) uint64 {
	tb.transaction.ArbitraryData = append(tb.transaction.ArbitraryData, arb)
	return uint64(len(tb.transaction.ArbitraryData) - 1)
}

// AddTransactionSignature adds a transaction signature to the transaction,
// returning the index of the signature within the transaction. The signature
// should already be valid, and shouldn't sign any of the inputs that were
// added by calling 'FundSiacoins' or 'FundSiafunds'.
func (tb *transactionBuilder) AddTransactionSignature(sig types.TransactionSignature) uint64 {
	tb.transaction.TransactionSignatures = append(tb.transaction.TransactionSignatures, sig)
	return uint64(len(tb.transaction.TransactionSignatures) - 1)
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
	tb.transactionSignatures = nil
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
func (tb *transactionBuilder) Sign(wholeTransaction bool) ([]types.Transaction, error) {
	if tb.signed {
		return nil, errBuilderAlreadySigned
	}

	// Create the coveredfields struct.
	var coveredFields types.CoveredFields
	if wholeTransaction {
		coveredFields = types.CoveredFields{WholeTransaction: true}
	} else {
		for i := range tb.transaction.MinerFees {
			coveredFields.MinerFees = append(coveredFields.MinerFees, uint64(i))
		}
		for i := range tb.transaction.CoinInputs {
			coveredFields.CoinInputs = append(coveredFields.CoinInputs, uint64(i))
		}
		for i := range tb.transaction.CoinOutputs {
			coveredFields.CoinOutputs = append(coveredFields.CoinOutputs, uint64(i))
		}
		for i := range tb.transaction.BlockStakeInputs {
			coveredFields.BlockStakeInputs = append(coveredFields.BlockStakeInputs, uint64(i))
		}
		for i := range tb.transaction.BlockStakeOutputs {
			coveredFields.BlockStakeOutputs = append(coveredFields.BlockStakeOutputs, uint64(i))
		}
		for i := range tb.transaction.ArbitraryData {
			coveredFields.ArbitraryData = append(coveredFields.ArbitraryData, uint64(i))
		}
	}
	// TransactionSignatures don't get covered by the 'WholeTransaction' flag,
	// and must be covered manually.
	for i := range tb.transaction.TransactionSignatures {
		coveredFields.TransactionSignatures = append(coveredFields.TransactionSignatures, uint64(i))
	}

	// For each siacoin input in the transaction that we added, provide a
	// signature.
	tb.wallet.mu.Lock()
	defer tb.wallet.mu.Unlock()
	for _, inputIndex := range tb.coinInputs {
		input := tb.transaction.CoinInputs[inputIndex]
		key := tb.wallet.keys[input.UnlockConditions.UnlockHash()]
		newSigIndices, err := addSignatures(&tb.transaction, coveredFields, input.UnlockConditions, crypto.Hash(input.ParentID), key)
		if err != nil {
			return nil, err
		}
		tb.transactionSignatures = append(tb.transactionSignatures, newSigIndices...)
		tb.signed = true // Signed is set to true after one successful signature to indicate that future signings can cause issues.
	}
	for _, inputIndex := range tb.blockstakeInputs {
		input := tb.transaction.BlockStakeInputs[inputIndex]
		key := tb.wallet.keys[input.UnlockConditions.UnlockHash()]
		newSigIndices, err := addSignatures(&tb.transaction, coveredFields, input.UnlockConditions, crypto.Hash(input.ParentID), key)
		if err != nil {
			return nil, err
		}
		tb.transactionSignatures = append(tb.transactionSignatures, newSigIndices...)
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
func (tb *transactionBuilder) ViewAdded() (newParents, coinInputs, blockstakeInputs, transactionSignatures []int) {
	return tb.newParents, tb.coinInputs, tb.blockstakeInputs, tb.transactionSignatures
}

// RegisterTransaction takes a transaction and its parents and returns a
// TransactionBuilder which can be used to expand the transaction. The most
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
