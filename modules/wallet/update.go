package wallet

import (
	"math"

	"github.com/NebulousLabs/errors"
	bolt "github.com/rivine/bbolt"
	"github.com/rivine/rivine/modules"
	"github.com/rivine/rivine/types"
)

type (
	spentCoinOutputSet       map[types.CoinOutputID]types.CoinOutput
	spentBlockStakeOutputSet map[types.BlockStakeOutputID]types.BlockStakeOutput
)

// threadedResetSubscriptions unsubscribes the wallet from the consensus set and transaction pool
// and subscribes again.
func (w *Wallet) threadedResetSubscriptions() error {
	if !w.scanLock.TryLock() {
		return errScanInProgress
	}
	defer w.scanLock.Unlock()

	w.cs.Unsubscribe(w)
	w.tpool.Unsubscribe(w)

	err := w.cs.ConsensusSetSubscribe(w, modules.ConsensusChangeBeginning, w.tg.StopChan())
	if err != nil {
		return err
	}
	w.tpool.TransactionPoolSubscribe(w)
	return nil
}

// advanceSeedLookahead generates all keys from the current primary seed progress up to index
// and adds them to the set of spendable keys.  Therefore the new primary seed progress will
// be index+1 and new lookahead keys will be generated starting from index+1
// Returns true if a blockchain rescan is required
func (w *Wallet) advanceSeedLookahead(index uint64) (bool, error) {
	progress, err := dbGetPrimarySeedProgress(w.dbTx)
	if err != nil {
		return false, err
	}
	newProgress := index + 1

	// Add spendable keys and remove them from lookahead
	spendableKeys := generateKeys(w.primarySeed, progress, newProgress-progress)
	for _, key := range spendableKeys {
		uh := key.UnlockHash()
		w.keys[uh] = key
		delete(w.lookahead, uh)
	}

	// Update the primarySeedProgress
	dbPutPrimarySeedProgress(w.dbTx, newProgress)
	if err != nil {
		return false, err
	}

	// Regenerate lookahead
	w.regenerateLookahead(newProgress)

	// If more than lookaheadRescanThreshold keys were generated
	// also initialize a rescan just to be safe.
	if uint64(len(spendableKeys)) > lookaheadRescanThreshold {
		return true, nil
	}

	return false, nil
}

// getOutputRelevance is a helper function that checks if a condition is
// relevant for one of the wallet's spendable keys or future keys
func (w *Wallet) getOutputRelevance(cond types.UnlockConditionProxy) outputRelevance {
	if cond.ConditionType() == types.ConditionTypeNil {
		return outputRelevanceNone
	}
	if w.isWalletAddress(cond.UnlockHash()) {
		return outputRelevanceWallet
	}
	uhs, _ := getMultisigConditionProperties(cond.Condition)
	for _, uh := range uhs {
		if w.isWalletAddress(uh) {
			return outputRelevanceMultisigWallet
		}
	}
	return outputRelevanceNone
}

func (w *Wallet) isWalletAddress(uh types.UnlockHash) bool {
	_, exists := w.keys[uh]
	return exists
}

type outputRelevance uint8

const (
	outputRelevanceNone outputRelevance = iota
	outputRelevanceWallet
	outputRelevanceMultisigWallet
)

// updateLookahead uses a consensus change to update the seed progress if one of the outputs
// contains an unlock hash of the lookahead set. Returns true if a blockchain rescan is required
func (w *Wallet) updateLookahead(tx *bolt.Tx, cc modules.ConsensusChange) (bool, error) {
	var largestIndex uint64
	for _, diff := range cc.CoinOutputDiffs {
		if index, ok := w.lookahead[diff.CoinOutput.Condition.UnlockHash()]; ok {
			if index > largestIndex {
				largestIndex = index
			}
		}
	}
	for _, diff := range cc.BlockStakeOutputDiffs {
		if index, ok := w.lookahead[diff.BlockStakeOutput.Condition.UnlockHash()]; ok {
			if index > largestIndex {
				largestIndex = index
			}
		}
	}
	if largestIndex > 0 {
		return w.advanceSeedLookahead(largestIndex)
	}

	return false, nil
}

// updateConfirmedSet uses a consensus change to update the confirmed set of
// outputs as understood by the wallet.
func (w *Wallet) updateConfirmedSet(tx *bolt.Tx, cc modules.ConsensusChange) error {
	for _, diff := range cc.CoinOutputDiffs {
		switch w.getOutputRelevance(diff.CoinOutput.Condition) {
		case outputRelevanceNone:
			continue

		case outputRelevanceWallet:
			var err error
			if diff.Direction == modules.DiffApply {
				w.log.Println("Wallet has gained a spendable coin output:", diff.ID, "::", diff.CoinOutput.Value.String())
				err = dbPutCoinOutput(tx, diff.ID, diff.CoinOutput)
			} else {
				w.log.Println("Wallet has lost a spendable coin output:", diff.ID, "::", diff.CoinOutput.Value.String())
				err = dbDeleteCoinOutput(tx, diff.ID)
			}
			if err != nil {
				w.log.Severe("Could not update coin output:", err)
				return err
			}

		case outputRelevanceMultisigWallet:
			var err error
			if diff.Direction == modules.DiffApply {
				w.log.Println("Multisig Wallet ", diff.CoinOutput.Condition.UnlockHash().String(),
					" has gained a spendable coin output:", diff.ID, "::", diff.CoinOutput.Value.String())
				err = dbPutMultisigCoinOutput(tx, diff.ID, diff.CoinOutput)
			} else {
				w.log.Println("Multisig Wallet ", diff.CoinOutput.Condition.UnlockHash().String(),
					" has lost a spendable coin output:", diff.ID, "::", diff.CoinOutput.Value.String())
				err = dbDeleteMultisigCoinOutput(tx, diff.ID)
			}
			if err != nil {
				w.log.Severe("Could not update multisig coin output:", err)
				return err
			}
		}
	}
	for _, diff := range cc.BlockStakeOutputDiffs {
		switch w.getOutputRelevance(diff.BlockStakeOutput.Condition) {
		case outputRelevanceNone:
			continue

		case outputRelevanceWallet:
			var err error
			if diff.Direction == modules.DiffApply {
				w.log.Println("Wallet has gained a spendable block stake output:", diff.ID, "::", diff.BlockStakeOutput.Value)
				err = dbPutBlockStakeOutput(tx, diff.ID, diff.BlockStakeOutput)
			} else {
				w.log.Println("Wallet has lost a spendable block stake output:", diff.ID, "::", diff.BlockStakeOutput.Value)
				err = dbDeleteBlockStakeOutput(tx, diff.ID)
				if err == nil {
					err = dbDeleteUnspentBlockstakeOutput(tx, diff.ID)
				}
			}
			if err != nil {
				w.log.Severe("Could not update block stake output:", err)
				return err
			}

		case outputRelevanceMultisigWallet:
			var err error
			if diff.Direction == modules.DiffApply {
				w.log.Println("Multisig Wallet ", diff.BlockStakeOutput.Condition.UnlockHash().String(),
					"has gained a spendable block stake output:", diff.ID, "::", diff.BlockStakeOutput.Value)
				err = dbPutMultisigBlockStakeOutput(tx, diff.ID, diff.BlockStakeOutput)
			} else {
				w.log.Println("Multisig Wallet ", diff.BlockStakeOutput.Condition.UnlockHash().String(),
					" has lost a spendable block stake output:", diff.ID, "::", diff.BlockStakeOutput.Value)
				err = dbDeleteMultisigBlockStakeOutput(tx, diff.ID)
			}
			if err != nil {
				w.log.Severe("Could not update multisig block stake output:", err)
				return err
			}
		}
	}
	return nil
}

// revertHistory reverts any transaction history that was destroyed by reverted
// blocks in the consensus change.
func (w *Wallet) revertHistory(tx *bolt.Tx, reverted []types.Block) error {
	for _, block := range reverted {
		// Remove any transactions that have been reverted.
		for i := len(block.Transactions) - 1; i >= 0; i-- {
			// If the transaction is relevant to the wallet, it will be the
			// most recent transaction in bucketProcessedTransactions.
			txid := block.Transactions[i].ID()
			pt, err := dbGetLastProcessedTransaction(tx)
			if err != nil {
				break // bucket is empty
			}
			if txid == pt.TransactionID {
				w.log.Println("A wallet transaction has been reverted due to a reorg:", txid)
				if err := dbDeleteLastProcessedTransaction(tx); err != nil {
					w.log.Severe("Could not revert transaction:", err)
					return err
				}
			}
		}

		// Remove the miner payout transaction if applicable.
		for i, mp := range block.MinerPayouts {
			// If the transaction is relevant to the wallet, it will be the
			// most recent transaction in bucketProcessedTransactions.
			pt, err := dbGetLastProcessedTransaction(tx)
			if err != nil {
				break // bucket is empty
			}
			if types.TransactionID(block.ID()) == pt.TransactionID {
				w.log.Println("Miner payout has been reverted due to a reorg:", block.MinerPayoutID(uint64(i)), "::", mp.Value.String())
				if err := dbDeleteLastProcessedTransaction(tx); err != nil {
					w.log.Severe("Could not revert transaction:", err)
					return err
				}
				break // there will only ever be one miner transaction
			}
		}

		// decrement the consensus height
		if block.ID() != w.chainCts.GenesisBlockID() {
			consensusHeight, err := dbGetConsensusHeight(tx)
			if err != nil {
				return err
			}
			err = dbPutConsensusHeight(tx, consensusHeight-1)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// outputs and collects them in a map of CoinOutputID -> CoinOutput.
func computeSpentCoinOutputSet(diffs []modules.CoinOutputDiff) spentCoinOutputSet {
	outputs := make(spentCoinOutputSet)
	for _, diff := range diffs {
		if diff.Direction == modules.DiffRevert {
			// DiffRevert means spent.
			outputs[diff.ID] = diff.CoinOutput
		}
	}
	return outputs
}

// computeSpentBlockStakeOutputSet scans a slice of block stake output diffs for spent
// outputs and collects them in a map of BlockStakeOutputID -> BlockStakeOutput.
func computeSpentBlockStakeOutputSet(diffs []modules.BlockStakeOutputDiff) spentBlockStakeOutputSet {
	outputs := make(spentBlockStakeOutputSet)
	for _, diff := range diffs {
		if diff.Direction == modules.DiffRevert {
			// DiffRevert means spent.
			outputs[diff.ID] = diff.BlockStakeOutput
		}
	}
	return outputs
}

// computeProcessedTransactionsFromBlock searches all the miner payouts and
// transactions in a block and computes a ProcessedTransaction slice containing
// all of the transactions processed for the given block.
func (w *Wallet) computeProcessedTransactionsFromBlock(tx *bolt.Tx, block types.Block, spentCoinOutputs spentCoinOutputSet, spentBlockStakeOutputs spentBlockStakeOutputSet, consensusHeight types.BlockHeight) []modules.ProcessedTransaction {
	var pts []modules.ProcessedTransaction

	// Find ProcessedTransactions from miner payouts.
	var relevant bool
	for _, mp := range block.MinerPayouts {
		relevant = relevant || w.isWalletAddress(mp.UnlockHash)
	}
	if relevant {
		w.log.Println("Wallet has received new miner payouts:", block.ID())
		// Apply the miner payout transaction if applicable.
		minerPT := modules.ProcessedTransaction{
			Transaction:           types.Transaction{},
			TransactionID:         types.TransactionID(block.ID()),
			ConfirmationHeight:    consensusHeight,
			ConfirmationTimestamp: block.Timestamp,
		}
		for i, mp := range block.MinerPayouts {
			w.log.Println("\tminer payout:", block.MinerPayoutID(uint64(i)), "::", mp.Value.String())
			minerPT.Outputs = append(minerPT.Outputs, modules.ProcessedOutput{
				ID:             types.OutputID(block.MinerPayoutID(uint64(i))),
				FundType:       types.SpecifierMinerPayout,
				MaturityHeight: consensusHeight + w.chainCts.MaturityDelay,
				WalletAddress:  w.isWalletAddress(mp.UnlockHash),
				RelatedAddress: mp.UnlockHash,
				Value:          mp.Value,
			})
		}
		pts = append(pts, minerPT)
	}

	// Find ProcessedTransactions from transactions.
	for ti, txn := range block.Transactions {
		// Determine if transaction is relevant.
		relevant := false
		for _, co := range txn.CoinOutputs {
			relevant = relevant || w.getOutputRelevance(co.Condition) != outputRelevanceNone
		}
		for _, bso := range txn.BlockStakeOutputs {
			relevant = relevant || w.getOutputRelevance(bso.Condition) != outputRelevanceNone
		}
		for _, ci := range txn.CoinInputs {
			relevant = relevant || w.getOutputRelevance(spentCoinOutputs[ci.ParentID].Condition) != outputRelevanceNone
		}
		for _, bsi := range txn.BlockStakeInputs {
			relevant = relevant || w.getOutputRelevance(spentBlockStakeOutputs[bsi.ParentID].Condition) != outputRelevanceNone
		}

		// Only create a ProcessedTransaction if transaction is relevant.
		if !relevant {
			continue
		}
		w.log.Println("A transaction has been confirmed on the blockchain:", txn.ID())

		blockheight, blockexists := w.cs.BlockHeightOfBlock(block)
		if !blockexists {
			panic("Block wherer ubs is used to respent, does not yet exist as processedblock")
		}

		pt := modules.ProcessedTransaction{
			Transaction:           txn,
			TransactionID:         txn.ID(),
			ConfirmationHeight:    consensusHeight,
			ConfirmationTimestamp: block.Timestamp,
		}

		for _, sci := range txn.CoinInputs {
			pco := spentCoinOutputs[sci.ParentID]
			uh := pco.Condition.UnlockHash()
			pi := modules.ProcessedInput{
				ParentID:       types.OutputID(sci.ParentID),
				FundType:       types.SpecifierCoinInput,
				WalletAddress:  w.isWalletAddress(uh),
				RelatedAddress: uh,
				Value:          pco.Value,
			}
			pt.Inputs = append(pt.Inputs, pi)

			// Log any wallet-relevant inputs.
			if pi.WalletAddress {
				w.log.Println("\tCoin Input:", pi.ParentID, "::", pi.Value.String())
			}
		}

		for i, sco := range txn.CoinOutputs {
			uh := sco.Condition.UnlockHash()
			po := modules.ProcessedOutput{
				ID:             types.OutputID(txn.CoinOutputID(uint64(i))),
				FundType:       types.SpecifierCoinOutput,
				MaturityHeight: consensusHeight,
				WalletAddress:  w.isWalletAddress(uh),
				RelatedAddress: uh,
				Value:          sco.Value,
			}
			pt.Outputs = append(pt.Outputs, po)

			// Log any wallet-relevant outputs.
			if po.WalletAddress {
				w.log.Println("\tCoin Output:", po.ID, "::", po.Value.String())
			}
		}

		for _, bsi := range txn.BlockStakeInputs {
			pbso := spentBlockStakeOutputs[bsi.ParentID]
			uh := pbso.Condition.UnlockHash()
			pi := modules.ProcessedInput{
				ParentID:       types.OutputID(bsi.ParentID),
				FundType:       types.SpecifierBlockStakeInput,
				WalletAddress:  w.isWalletAddress(uh),
				RelatedAddress: uh,
				Value:          pbso.Value,
			}
			pt.Inputs = append(pt.Inputs, pi)
			// Log any wallet-relevant inputs.
			if pi.WalletAddress {
				w.log.Println("\tBlockStake Input:", pi.ParentID, "::", pi.Value.String())
			}
		}

		for i, bso := range txn.BlockStakeOutputs {
			uh := bso.Condition.UnlockHash()
			po := modules.ProcessedOutput{
				ID:             types.OutputID(txn.BlockStakeOutputID(uint64(i))),
				FundType:       types.SpecifierBlockStakeOutput,
				MaturityHeight: consensusHeight,
				WalletAddress:  w.isWalletAddress(uh),
				RelatedAddress: uh,
				Value:          bso.Value,
			}
			pt.Outputs = append(pt.Outputs, po)
			// Log any wallet-relevant outputs.
			if po.WalletAddress {
				w.log.Println("\tBlockStake Output:", po.ID, "::", po.Value.String())
			}
			bsoid := txn.BlockStakeOutputID(uint64(i))
			if _, err := dbGetBlockStakeOutput(tx, bsoid); err == nil {
				dbPutUnspentBlockstakeOutput(tx, bsoid, types.UnspentBlockStakeOutput{
					BlockStakeOutputID: bsoid,
					Indexes: types.BlockStakeOutputIndexes{
						BlockHeight:      blockheight,
						TransactionIndex: uint64(ti),
						OutputIndex:      uint64(i),
					},
					Value:     bso.Value,
					Condition: bso.Condition,
				})
			}
		}

		for _, fee := range txn.MinerFees {
			pt.Outputs = append(pt.Outputs, modules.ProcessedOutput{
				FundType:       types.SpecifierMinerFee,
				MaturityHeight: consensusHeight + w.chainCts.MaturityDelay,
				Value:          fee,
			})
		}
		pts = append(pts, pt)
	}
	return pts
}

// applyHistory applies any transaction history that the applied blocks
// introduced.
func (w *Wallet) applyHistory(tx *bolt.Tx, cc modules.ConsensusChange) error {
	spentCoinOutputs := computeSpentCoinOutputSet(cc.CoinOutputDiffs)
	spentBlockStakeOutputs := computeSpentBlockStakeOutputSet(cc.BlockStakeOutputDiffs)
	genesisID := w.chainCts.GenesisBlockID()

	for _, block := range cc.AppliedBlocks {
		consensusHeight, err := dbGetConsensusHeight(tx)
		if err != nil {
			return errors.AddContext(err, "failed to get consensus height")
		}
		// Increment the consensus height.
		if block.ID() != genesisID {
			consensusHeight++
			err = dbPutConsensusHeight(tx, consensusHeight)
			if err != nil {
				return errors.AddContext(err, "failed to store consensus height in database")
			}
		}

		pts := w.computeProcessedTransactionsFromBlock(tx, block, spentCoinOutputs, spentBlockStakeOutputs, consensusHeight)
		for _, pt := range pts {
			err := dbAppendProcessedTransaction(tx, pt)
			if err != nil {
				return errors.AddContext(err, "could not put processed transaction")
			}
		}
	}

	return nil
}

// ProcessConsensusChange parses a consensus change to update the set of
// confirmed outputs known to the wallet.
func (w *Wallet) ProcessConsensusChange(cc modules.ConsensusChange) {
	if err := w.tg.Add(); err != nil {
		return
	}
	defer w.tg.Done()

	w.mu.Lock()
	defer w.mu.Unlock()

	if needRescan, err := w.updateLookahead(w.dbTx, cc); err != nil {
		w.log.Severe("ERROR: failed to update lookahead:", err)
		w.dbRollback = true
	} else if needRescan {
		go w.threadedResetSubscriptions()
	}
	if err := w.updateConfirmedSet(w.dbTx, cc); err != nil {
		w.log.Severe("ERROR: failed to update confirmed set:", err)
		w.dbRollback = true
	}
	if err := w.revertHistory(w.dbTx, cc.RevertedBlocks); err != nil {
		w.log.Severe("ERROR: failed to revert consensus change:", err)
		w.dbRollback = true
	}
	if err := w.applyHistory(w.dbTx, cc); err != nil {
		w.log.Severe("ERROR: failed to apply consensus change:", err)
		w.dbRollback = true
	}
	if err := dbPutConsensusChangeID(w.dbTx, cc.ID); err != nil {
		w.log.Severe("ERROR: failed to update consensus change ID:", err)
		w.dbRollback = true
	}

	/*
		if cc.Synced {
			go w.threadedDefragWallet()
		}
	*/
}

// ReceiveUpdatedUnconfirmedTransactions updates the wallet's unconfirmed
// transaction set.
func (w *Wallet) ReceiveUpdatedUnconfirmedTransactions(txns []types.Transaction, _ modules.ConsensusChange) {
	if err := w.tg.Add(); err != nil {
		// Gracefully reject transactions if the wallet's Close method has
		// closed the wallet's ThreadGroup already.
		return
	}
	defer w.tg.Done()
	w.mu.Lock()
	defer w.mu.Unlock()

	w.unconfirmedProcessedTransactions = nil
	for _, txn := range txns {
		// To save on code complexity, relevancy is determined while building
		// up the wallet transaction.
		var relevant bool
		pt := modules.ProcessedTransaction{
			Transaction:           txn,
			TransactionID:         txn.ID(),
			ConfirmationHeight:    types.BlockHeight(math.MaxUint64),
			ConfirmationTimestamp: types.Timestamp(math.MaxUint64),
		}
		for _, sci := range txn.CoinInputs {
			if output, err := dbGetCoinOutput(w.dbTx, sci.ParentID); err == nil {
				relevant = true
				pt.Inputs = append(pt.Inputs, modules.ProcessedInput{
					FundType:       types.SpecifierCoinInput,
					WalletAddress:  true,
					RelatedAddress: output.Condition.UnlockHash(),
					Value:          output.Value,
				})
			} else if output, err := dbGetMultisigCoinOutput(w.dbTx, sci.ParentID); err == nil {
				pt.Inputs = append(pt.Inputs, modules.ProcessedInput{
					FundType:       types.SpecifierCoinInput,
					WalletAddress:  false,
					RelatedAddress: output.Condition.UnlockHash(),
					Value:          output.Value,
				})
			}

		}
		for _, sco := range txn.CoinOutputs {
			switch w.getOutputRelevance(sco.Condition) {
			case outputRelevanceNone:
				continue

			case outputRelevanceWallet:
				relevant = true
				pt.Outputs = append(pt.Outputs, modules.ProcessedOutput{
					FundType:       types.SpecifierCoinOutput,
					MaturityHeight: types.BlockHeight(math.MaxUint64),
					WalletAddress:  true,
					RelatedAddress: sco.Condition.UnlockHash(),
					Value:          sco.Value,
				})

			case outputRelevanceMultisigWallet:
				relevant = true
				pt.Outputs = append(pt.Outputs, modules.ProcessedOutput{
					FundType:       types.SpecifierCoinOutput,
					MaturityHeight: types.BlockHeight(math.MaxUint64),
					WalletAddress:  false,
					RelatedAddress: sco.Condition.UnlockHash(),
					Value:          sco.Value,
				})
			}
		}
		if relevant {
			w.unconfirmedProcessedTransactions = append(w.unconfirmedProcessedTransactions, pt)
		}
	}
}
