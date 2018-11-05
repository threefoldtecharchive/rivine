package wallet

import (
	"errors"

	"github.com/threefoldtech/rivine/modules"
	"github.com/threefoldtech/rivine/types"
)

var (
	errOutOfBounds      = errors.New("requesting transactions at unknown confirmation heights")
	errNoHistoryForAddr = errors.New("no history found for provided address")
)

// AddressTransactions returns all of the wallet transactions associated with a
// single unlock hash.
func (w *Wallet) AddressTransactions(uh types.UnlockHash) (pts []modules.ProcessedTransaction, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.unlocked {
		err = modules.ErrLockedWallet
		return
	}

	for _, pt := range w.processedTransactions {
		relevant := false
		for _, input := range pt.Inputs {
			if input.RelatedAddress == uh {
				relevant = true
				break
			}
		}
		for _, output := range pt.Outputs {
			if output.RelatedAddress == uh {
				relevant = true
				break
			}
		}
		if relevant {
			pts = append(pts, pt)
		}
	}
	return
}

// AddressUnconfirmedTransactions returns all of the unconfirmed wallet transactions
// related to a specific address.
func (w *Wallet) AddressUnconfirmedTransactions(uh types.UnlockHash) (pts []modules.ProcessedTransaction, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.unlocked {
		err = modules.ErrLockedWallet
		return
	}

	// Scan the full list of unconfirmed transactions to see if there are any
	// related transactions.
	for _, pt := range w.unconfirmedProcessedTransactions {
		relevant := false
		for _, input := range pt.Inputs {
			if input.RelatedAddress == uh {
				relevant = true
				break
			}
		}
		for _, output := range pt.Outputs {
			if output.RelatedAddress == uh {
				relevant = true
				break
			}
		}
		if relevant {
			pts = append(pts, pt)
		}
	}
	return
}

// Transaction returns the transaction with the given id. 'False' is returned
// if the transaction does not exist.
func (w *Wallet) Transaction(txid types.TransactionID) (modules.ProcessedTransaction, bool, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if !w.unlocked {
		return modules.ProcessedTransaction{}, false, modules.ErrLockedWallet
	}
	pt, exists := w.processedTransactionMap[txid]
	if !exists {
		return modules.ProcessedTransaction{}, exists, nil
	}
	return *pt, exists, nil
}

// Transactions returns all transactions relevant to the wallet that were
// confirmed in the range [startHeight, endHeight].
func (w *Wallet) Transactions(startHeight, endHeight types.BlockHeight) (pts []modules.ProcessedTransaction, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.unlocked {
		err = modules.ErrLockedWallet
		return
	}

	if startHeight > w.consensusSetHeight || startHeight > endHeight {
		return nil, errOutOfBounds
	}
	if len(w.processedTransactions) == 0 {
		return nil, nil
	}

	for _, pt := range w.processedTransactions {
		if pt.ConfirmationHeight > endHeight {
			break
		}
		if pt.ConfirmationHeight >= startHeight {
			pts = append(pts, pt)
		}
	}
	return pts, nil
}

// BlockStakeStats returns the blockstake statistical information of this wallet
func (w *Wallet) BlockStakeStats() (BCcountLast1000 uint64, BCfeeLast1000 types.Currency, BlockCount uint64, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.unlocked {
		err = modules.ErrLockedWallet
		return
	}

	BlockHeightCounter := w.cs.Height()

	for BlockCount = 0; BlockCount < 1000; BlockCount++ {

		block, _ := w.cs.BlockAtHeight(BlockHeightCounter)
		ind := block.POBSOutput
		blockOld, _ := w.cs.BlockAtHeight(ind.BlockHeight)

		bso := blockOld.Transactions[ind.TransactionIndex].BlockStakeOutputs[ind.OutputIndex]

		relevant := false
		switch uh := bso.Condition.UnlockHash(); uh.Type {
		case types.UnlockTypePubKey:
			_, relevant = w.keys[uh]
		case types.UnlockTypeNil:
			relevant = true
		case types.UnlockTypeMultiSig:
			uhs, _ := getMultisigConditionProperties(bso.Condition.Condition)
			if len(uhs) > 0 {
				for _, uh := range uhs {
					_, relevant = w.keys[uh]
					if relevant {
						break
					}
				}
			}
		}

		if relevant {
			BCcountLast1000++
			BCfeeLast1000 = BCfeeLast1000.Add(w.chainCts.BlockCreatorFee)
			if w.chainCts.TransactionFeeCondition.ConditionType() == types.ConditionTypeNil {
				// only when tx fee beneficiary is not defined is the miner fees for the block creator
				BCfeeLast1000 = BCfeeLast1000.Add(block.CalculateTotalMinerFees())
			}
		}
		if BlockHeightCounter == 0 {
			BlockCount++
			break
		}
		BlockHeightCounter--
	}

	return
}

// UnconfirmedTransactions returns the set of unconfirmed transactions that are
// relevant to the wallet.
func (w *Wallet) UnconfirmedTransactions() ([]modules.ProcessedTransaction, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if !w.unlocked {
		return nil, modules.ErrLockedWallet
	}
	return w.unconfirmedProcessedTransactions, nil
}

// CreateRawTransaction with the given inputs and outputs
func (w *Wallet) CreateRawTransaction(coids []types.CoinOutputID, bsoids []types.BlockStakeOutputID,
	cos []types.CoinOutput, bsos []types.BlockStakeOutput, arb []byte) (types.Transaction, error) {

	coinInputCount := types.ZeroCurrency
	blockStakeInputCount := types.ZeroCurrency

	// Make sure coin inputs and outputs + txnfee match
	for _, id := range coids {
		co, err := w.cs.GetCoinOutput(id)
		if err != nil {
			return types.Transaction{}, err
		}
		coinInputCount = coinInputCount.Add(co.Value)
	}
	requiredCoins := w.chainCts.MinimumTransactionFee
	for _, co := range cos {
		requiredCoins = requiredCoins.Add(co.Value)
	}

	if requiredCoins.Cmp(coinInputCount) != 0 {
		return types.Transaction{}, errors.New("Mismatched coin input - output count")
	}

	for _, id := range bsoids {
		bso, err := w.cs.GetBlockStakeOutput(id)
		if err != nil {
			return types.Transaction{}, err
		}
		blockStakeInputCount = blockStakeInputCount.Add(bso.Value)
	}
	requiredBlockStakes := types.ZeroCurrency
	for _, bso := range bsos {
		requiredBlockStakes = requiredBlockStakes.Add(bso.Value)
	}

	if requiredBlockStakes.Cmp(blockStakeInputCount) != 0 {
		return types.Transaction{}, errors.New("Mismatched blockstake input - output count")
	}

	if uint64(len(arb)) > w.chainCts.ArbitraryDataSizeLimit {
		return types.Transaction{}, errors.New("Arbitrary data too large")
	}

	txnBuilder := w.StartTransaction()
	// No need to drop the builder since we manually added inputs,
	// so they won't be consumed from the wallet

	for _, ci := range coids {
		txnBuilder.AddCoinInput(types.CoinInput{ParentID: ci})
	}
	for _, bsi := range bsoids {
		txnBuilder.AddBlockStakeInput(types.BlockStakeInput{ParentID: bsi})
	}
	for _, co := range cos {
		txnBuilder.AddCoinOutput(co)
	}
	for _, bso := range bsos {
		txnBuilder.AddBlockStakeOutput(bso)
	}
	txnBuilder.AddMinerFee(w.chainCts.MinimumTransactionFee)
	txnBuilder.SetArbitraryData(arb)

	txn, _ := txnBuilder.View()
	return txn, nil
}

// GreedySign attempts to sign every input in the transaction that can be signed
// using the keys loaded in this wallet. The transaction is assumed to be valid
func (w *Wallet) GreedySign(txn types.Transaction) (types.Transaction, error) {
	txnBuilder := w.RegisterTransaction(txn, nil)
	err := txnBuilder.SignAllPossible()
	signedTxn, _ := txnBuilder.View()
	return signedTxn, err
}
