package wallet

import (
	"errors"

	"github.com/rivine/rivine/modules"
	"github.com/rivine/rivine/types"
)

var (
	errOutOfBounds      = errors.New("requesting transactions at unknown confirmation heights")
	errNoHistoryForAddr = errors.New("no history found for provided address")
)

// AddressTransactions returns all of the wallet transactions associated with a
// single unlock hash.
func (w *Wallet) AddressTransactions(uh types.UnlockHash) (pts []modules.ProcessedTransaction) {
	w.mu.Lock()
	defer w.mu.Unlock()

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
	return pts
}

// AddressUnconfirmedHistory returns all of the unconfirmed wallet transactions
// related to a specific address.
func (w *Wallet) AddressUnconfirmedTransactions(uh types.UnlockHash) (pts []modules.ProcessedTransaction) {
	w.mu.Lock()
	defer w.mu.Unlock()

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
	return pts
}

// Transaction returns the transaction with the given id. 'False' is returned
// if the transaction does not exist.
func (w *Wallet) Transaction(txid types.TransactionID) (modules.ProcessedTransaction, bool) {
	w.mu.Lock()
	defer w.mu.Unlock()
	pt, exists := w.processedTransactionMap[txid]
	if !exists {
		return modules.ProcessedTransaction{}, exists
	}
	return *pt, exists
}

// Transactions returns all transactions relevant to the wallet that were
// confirmed in the range [startHeight, endHeight].
func (w *Wallet) Transactions(startHeight, endHeight types.BlockHeight) (pts []modules.ProcessedTransaction, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

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
func (w *Wallet) BlockStakeStats() (BCcountLast1000 uint64, BCfeeLast1000 types.Currency, BlockCount uint64) {
	w.mu.Lock()
	defer w.mu.Unlock()

	BlockHeightCounter := w.cs.Height()

	for BlockCount = 0; BlockCount < 1000; BlockCount++ {

		block, _ := w.cs.BlockAtHeight(BlockHeightCounter)
		ind := block.POBSOutput
		blockOld, _ := w.cs.BlockAtHeight(ind.BlockHeight)

		bso := blockOld.Transactions[ind.TransactionIndex].BlockStakeOutputs[ind.OutputIndex]

		relevant := false

		_, exists := w.keys[bso.UnlockHash]
		if exists {
			relevant = true
		}

		if relevant {
			BCcountLast1000++
			BCfeeLast1000 = BCfeeLast1000.Add(block.CalculateSubsidy(w.chainCts.BlockCreatorFee))
		}
		if BlockHeightCounter == 0 {
			BlockCount++
			break
		}
		BlockHeightCounter--
	}

	return BCcountLast1000, BCfeeLast1000, BlockCount
}

// UnconfirmedTransactions returns the set of unconfirmed transactions that are
// relevant to the wallet.
func (w *Wallet) UnconfirmedTransactions() []modules.ProcessedTransaction {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.unconfirmedProcessedTransactions
}
