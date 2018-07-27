package wallet

import (
	"encoding/binary"
	"errors"
	"fmt"
	"sort"

	"github.com/rivine/rivine/build"
	"github.com/rivine/rivine/encoding"
	"github.com/rivine/rivine/modules"
	"github.com/rivine/rivine/types"
)

var (
	errOutOfBounds = errors.New("requesting transactions at unknown confirmation heights")
)

// AddressTransactions returns all of the wallet transactions associated with a
// single unlock hash.
func (w *Wallet) AddressTransactions(uh types.UnlockHash) (pts []modules.ProcessedTransaction, err error) {
	if err := w.tg.Add(); err != nil {
		return []modules.ProcessedTransaction{}, err
	}
	defer w.tg.Done()
	// ensure durability of reported transactions
	w.mu.Lock()
	defer w.mu.Unlock()
	if err = w.syncDB(); err != nil {
		return
	}

	txnIndices, _ := dbGetAddrTransactions(w.dbTx, uh)
	for _, i := range txnIndices {
		pt, err := dbGetProcessedTransaction(w.dbTx, i)
		if err != nil {
			continue
		}
		pts = append(pts, pt)
	}
	return pts, nil
}

// AddressUnconfirmedTransactions returns all of the unconfirmed wallet transactions
// related to a specific address.
func (w *Wallet) AddressUnconfirmedTransactions(uh types.UnlockHash) (pts []modules.ProcessedTransaction, err error) {
	if err := w.tg.Add(); err != nil {
		return []modules.ProcessedTransaction{}, err
	}
	defer w.tg.Done()
	// ensure durability of reported transactions
	w.mu.Lock()
	defer w.mu.Unlock()
	if err = w.syncDB(); err != nil {
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
	return pts, err
}

// Transaction returns the transaction with the given id. 'False' is returned
// if the transaction does not exist.
func (w *Wallet) Transaction(txid types.TransactionID) (pt modules.ProcessedTransaction, found bool, err error) {
	if err := w.tg.Add(); err != nil {
		return modules.ProcessedTransaction{}, false, err
	}
	defer w.tg.Done()
	// ensure durability of reported transaction
	w.mu.Lock()
	defer w.mu.Unlock()
	if err = w.syncDB(); err != nil {
		return
	}

	// Get the keyBytes for the given txid
	keyBytes, err := dbGetTransactionIndex(w.dbTx, txid)
	if err != nil {
		return modules.ProcessedTransaction{}, false, nil
	}

	// Retrieve the transaction
	found = encoding.Unmarshal(w.dbTx.Bucket(bucketProcessedTransactions).Get(keyBytes), &pt) == nil
	return
}

// Transactions returns all transactions relevant to the wallet that were
// confirmed in the range [startHeight, endHeight].
func (w *Wallet) Transactions(startHeight, endHeight types.BlockHeight) (pts []modules.ProcessedTransaction, err error) {
	if err := w.tg.Add(); err != nil {
		return nil, err
	}
	defer w.tg.Done()
	// ensure durability of reported transactions
	w.mu.Lock()
	defer w.mu.Unlock()
	if err = w.syncDB(); err != nil {
		return
	}

	height, err := dbGetConsensusHeight(w.dbTx)
	if err != nil {
		return
	} else if startHeight > height || startHeight > endHeight {
		return nil, errOutOfBounds
	}

	// Get the bucket, the largest key in it and the cursor
	bucket := w.dbTx.Bucket(bucketProcessedTransactions)
	cursor := bucket.Cursor()
	nextKey := bucket.Sequence() + 1

	// Database is empty
	if nextKey == 1 {
		return
	}

	var pt modules.ProcessedTransaction
	keyBytes := make([]byte, 8)
	var result int
	func() {
		// Recover from possible panic during binary search
		defer func() {
			r := recover()
			if r != nil {
				err = fmt.Errorf("%v", r)
			}
		}()

		// Start binary searching
		result = sort.Search(int(nextKey), func(i int) bool {
			// Create the key for the index
			binary.BigEndian.PutUint64(keyBytes, uint64(i))

			// Retrieve the processed transaction
			key, ptBytes := cursor.Seek(keyBytes)
			if build.DEBUG && key == nil {
				panic("Failed to retrieve processed Transaction by key")
			}

			// Decode the transaction
			if err = decodeProcessedTransaction(ptBytes, &pt); build.DEBUG && err != nil {
				panic(err)
			}

			return pt.ConfirmationHeight >= startHeight
		})
	}()
	if err != nil {
		return
	}

	if uint64(result) == nextKey {
		// No transaction was found
		return
	}

	// Create the key that corresponds to the result of the search
	binary.BigEndian.PutUint64(keyBytes, uint64(result))

	// Get the processed transaction and decode it
	key, ptBytes := cursor.Seek(keyBytes)
	if build.DEBUG && key == nil {
		build.Critical("Couldn't find the processed transaction from the search.")
	}
	if err = decodeProcessedTransaction(ptBytes, &pt); build.DEBUG && err != nil {
		build.Critical(err)
	}

	// Gather all transactions until endHeight is reached
	for pt.ConfirmationHeight <= endHeight {
		if build.DEBUG && pt.ConfirmationHeight < startHeight {
			build.Critical("wallet processed transactions are not sorted")
		}
		pts = append(pts, pt)

		// Get next processed transaction
		key, ptBytes := cursor.Next()
		if key == nil {
			break
		}

		// Decode the transaction
		if err := decodeProcessedTransaction(ptBytes, &pt); build.DEBUG && err != nil {
			panic("Failed to decode the processed transaction")
		}
	}
	return
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
	if err := w.tg.Add(); err != nil {
		return nil, err
	}
	defer w.tg.Done()
	w.mu.RLock()
	defer w.mu.RUnlock()
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
	err := txnBuilder.SignAllPossibleInputs()
	signedTxn, _ := txnBuilder.View()
	return signedTxn, err
}
