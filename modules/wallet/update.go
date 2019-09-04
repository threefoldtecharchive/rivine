package wallet

import (
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/threefoldtech/rivine/build"
	"github.com/threefoldtech/rivine/modules"
	"github.com/threefoldtech/rivine/types"
)

func (w *Wallet) subscribeWallet() error {
	// During rescan, print height every 3 seconds.
	if build.Release != "testing" {
		go func() {
			println("Rescanning consensus set...")
			for range time.Tick(time.Second * 3) {
				w.mu.RLock()
				height := w.consensusSetHeight
				done := w.subscribed
				w.mu.RUnlock()
				if done {
					println("\nDone!")
					break
				}
				print("\rScanned to height ", height, "...")
			}
		}()
	}
	err := w.cs.ConsensusSetSubscribe(w, modules.ConsensusChangeBeginning, w.tg.StopChan())
	if err != nil {
		return errors.New("wallet subscription failed: " + err.Error())
	}
	w.tpool.TransactionPoolSubscribe(w)
	return nil
}

// updateConfirmedSet uses a consensus change to update the confirmed set of
// outputs as understood by the wallet.
func (w *Wallet) updateConfirmedSet(cc modules.ConsensusChange) {
	for _, diff := range cc.CoinOutputDiffs {
		// Verify that the diff is relevant to the wallet.
		if _, exists := w.keys[diff.CoinOutput.Condition.UnlockHash()]; exists {
			_, exists = w.coinOutputs[diff.ID]
			if diff.Direction == modules.DiffApply {
				if exists {
					build.Severe("adding an existing output to wallet")
				}
				w.coinOutputs[diff.ID] = diff.CoinOutput
			} else {
				if !exists {
					build.Severe("deleting nonexisting output from wallet")
				}
				delete(w.coinOutputs, diff.ID)
			}
			continue
		}

		// try to get the unlock hash slice of a multisig
		unlockhashes, _ := getMultisigConditionProperties(diff.CoinOutput.Condition.Condition)
		if len(unlockhashes) == 0 {
			continue
		}
		for _, uh := range unlockhashes {
			if _, exists := w.keys[uh]; exists {
				_, exists = w.multiSigCoinOutputs[diff.ID]
				if diff.Direction == modules.DiffApply {
					if exists {
						build.Severe("adding an existing multisig output to wallet")
					}
					w.multiSigCoinOutputs[diff.ID] = diff.CoinOutput
				} else {
					if !exists {
						build.Severe("deleting nonexisting multisig output from wallet")
					}
					delete(w.multiSigCoinOutputs, diff.ID)
				}
				break
			}
		}
	}

	for _, diff := range cc.BlockStakeOutputDiffs {
		// Verify that the diff is relevant to the wallet.
		if _, exists := w.keys[diff.BlockStakeOutput.Condition.UnlockHash()]; exists {

			_, exists = w.blockstakeOutputs[diff.ID]
			if diff.Direction == modules.DiffApply {
				if exists {
					build.Severe("adding an existing output to wallet")
				}
				w.blockstakeOutputs[diff.ID] = diff.BlockStakeOutput
			} else {
				if !exists {
					build.Severe("deleting an nonexisting output from wallet")
				}
				delete(w.blockstakeOutputs, diff.ID)
			}
			continue
		}

		// try to get the unlock hash slice of a multisig
		unlockhashes, _ := getMultisigConditionProperties(diff.BlockStakeOutput.Condition.Condition)
		if len(unlockhashes) == 0 {
			continue
		}
		for _, uh := range unlockhashes {
			if _, exists := w.keys[uh]; exists {
				_, exists = w.multiSigBlockStakeOutputs[diff.ID]
				if diff.Direction == modules.DiffApply {
					if exists {
						build.Severe("adding an existing multisig output to wallet")
					}
					w.multiSigBlockStakeOutputs[diff.ID] = diff.BlockStakeOutput
				} else {
					if !exists {
						build.Severe("deleting nonexisting multisig output from wallet")
					}
					delete(w.multiSigBlockStakeOutputs, diff.ID)
				}
				break
			}
		}
	}
}

func getMultisigConditionProperties(condition types.MarshalableUnlockCondition) ([]types.UnlockHash, uint64) {
	ct := condition.ConditionType()
	if ct == types.ConditionTypeTimeLock {
		cg, ok := condition.(types.MarshalableUnlockConditionGetter)
		if !ok {
			build.Severe(fmt.Sprintf("unexpected Go-type for TimeLockCondition: %T", condition))
			return nil, 0
		}
		return getMultisigConditionProperties(cg.GetMarshalableUnlockCondition())
	}
	if ct != types.ConditionTypeMultiSignature {
		return nil, 0
	}
	type multisigCondition interface {
		types.UnlockHashSliceGetter
		GetMinimumSignatureCount() uint64
	}
	switch c := condition.(type) {
	case multisigCondition:
		return c.UnlockHashSlice(), c.GetMinimumSignatureCount()
	default:
		return nil, 0
	}
}

// revertHistory reverts any transaction history that was destroyed by reverted
// blocks in the consensus change.
func (w *Wallet) revertHistory(cc modules.ConsensusChange) {
	for _, block := range cc.RevertedBlocks {
		// Remove any transactions that have been reverted.
		for i := len(block.Transactions) - 1; i >= 0; i-- {
			// If the transaction is relevant to the wallet, it will be the
			// most recent transaction appended to w.processedTransactions.
			// Relevance can be determined just by looking at the last element
			// of w.processedTransactions.
			txn := block.Transactions[i]
			txid := txn.ID()
			if len(w.processedTransactions) > 0 && txid == w.processedTransactions[len(w.processedTransactions)-1].TransactionID {
				w.processedTransactions = w.processedTransactions[:len(w.processedTransactions)-1]
				delete(w.processedTransactionMap, txid)
			}
		}

		// Remove the miner payout transaction if applicable.
		for _, mp := range block.MinerPayouts {
			_, exists := w.keys[mp.UnlockHash]
			if exists {
				w.processedTransactions = w.processedTransactions[:len(w.processedTransactions)-1]
				delete(w.processedTransactionMap, types.TransactionID(block.ID()))
				break
			}
		}
		w.consensusSetHeight--
	}
}

// applyHistory applies any transaction history that was introduced by the
// applied blocks.
func (w *Wallet) applyHistory(cc modules.ConsensusChange) {
	for _, block := range cc.AppliedBlocks {
		w.consensusSetHeight++
		// Apply the miner payout transaction if applicable.
		minerPT := modules.ProcessedTransaction{
			Transaction:           types.Transaction{},
			TransactionID:         types.TransactionID(block.ID()),
			ConfirmationHeight:    w.consensusSetHeight,
			ConfirmationTimestamp: block.Timestamp,
		}
		relevant := false
		for i, mp := range block.MinerPayouts {
			_, exists := w.keys[mp.UnlockHash]
			if exists {
				relevant = true
			}
			minerPT.Outputs = append(minerPT.Outputs, modules.ProcessedOutput{
				FundType:       types.SpecifierMinerPayout,
				MaturityHeight: w.consensusSetHeight + w.chainCts.MaturityDelay,
				WalletAddress:  exists,
				RelatedAddress: mp.UnlockHash,
				Value:          mp.Value,
			})
			w.historicOutputs[types.OutputID(block.MinerPayoutID(uint64(i)))] = historicOutput{
				UnlockHash: mp.UnlockHash,
				Value:      mp.Value,
			}
		}
		if relevant {
			w.processedTransactions = append(w.processedTransactions, minerPT)
			w.processedTransactionMap[minerPT.TransactionID] = &w.processedTransactions[len(w.processedTransactions)-1]
		}

		blockheight, blockexists := w.cs.BlockHeightOfBlock(block)
		if !blockexists {
			build.Critical("Block wherer ubs is used to respent, does not yet exist as processedblock")
		}

		for ti, txn := range block.Transactions {
			relevant := false
			pt := modules.ProcessedTransaction{
				Transaction:           txn,
				TransactionID:         txn.ID(),
				ConfirmationHeight:    w.consensusSetHeight,
				ConfirmationTimestamp: block.Timestamp,
			}
			for _, sci := range txn.CoinInputs {
				output := w.historicOutputs[types.OutputID(sci.ParentID)]
				_, exists := w.keys[output.UnlockHash]
				if exists {
					relevant = true
				} else if _, exists = w.multiSigCoinOutputs[sci.ParentID]; exists {
					// Since we know about every multisig output that is still open and releated,
					// any relevant multisig input must have a parent ID present in the multisig
					// output map.
					relevant = true
					// set "exists" to false since the output is not owned by the wallet.
					exists = false
				}
				pt.Inputs = append(pt.Inputs, modules.ProcessedInput{
					FundType:       types.SpecifierCoinInput,
					WalletAddress:  exists,
					RelatedAddress: output.UnlockHash,
					Value:          output.Value,
				})
			}
			for i, sco := range txn.CoinOutputs {
				_, exists := w.keys[sco.Condition.UnlockHash()]
				if exists {
					relevant = true
				} else if _, exists = w.multiSigCoinOutputs[txn.CoinOutputID(uint64(i))]; exists {
					// If the coin output is a relevant multisig output, it's ID will already
					// be present in the multisigCoinOutputs map
					relevant = true
					// set "exists" to false since the output is not owned by the wallet.
					exists = false
				}
				uh := sco.Condition.UnlockHash()
				pt.Outputs = append(pt.Outputs, modules.ProcessedOutput{
					FundType:       types.SpecifierCoinOutput,
					MaturityHeight: w.consensusSetHeight,
					WalletAddress:  exists,
					RelatedAddress: uh,
					Value:          sco.Value,
				})
				w.historicOutputs[types.OutputID(txn.CoinOutputID(uint64(i)))] = historicOutput{
					UnlockHash: uh,
					Value:      sco.Value,
				}
			}
			for _, sfi := range txn.BlockStakeInputs {
				output := w.historicOutputs[types.OutputID(sfi.ParentID)]
				_, exists := w.keys[output.UnlockHash]
				if exists {
					relevant = true
				} else if _, exists = w.multiSigBlockStakeOutputs[sfi.ParentID]; exists {
					// Since we know about every multisig output that is still open and releated,
					// any relevant multisig input must have a parent ID present in the multisig
					// output map.
					relevant = true
					// set "exists" to false since the output is not owned by the wallet.
					exists = false
				}
				pt.Inputs = append(pt.Inputs, modules.ProcessedInput{
					FundType:       types.SpecifierBlockStakeInput,
					WalletAddress:  exists,
					RelatedAddress: output.UnlockHash,
					Value:          output.Value,
				})
			}
			for i, sfo := range txn.BlockStakeOutputs {
				_, exists := w.keys[sfo.Condition.UnlockHash()]
				if exists {
					relevant = true
				} else if _, exists = w.multiSigBlockStakeOutputs[txn.BlockStakeOutputID(uint64(i))]; exists {
					// If the block stake output is a relevant multisig output, it's ID will already
					// be present in the multisigBlockStakeOutputs map
					relevant = true
					// set "exists" to false since the output is not owned by the wallet.
					exists = false
				}
				uh := sfo.Condition.UnlockHash()
				pt.Outputs = append(pt.Outputs, modules.ProcessedOutput{
					FundType:       types.SpecifierBlockStakeOutput,
					MaturityHeight: w.consensusSetHeight,
					WalletAddress:  exists,
					RelatedAddress: uh,
					Value:          sfo.Value,
				})
				bsoid := txn.BlockStakeOutputID(uint64(i))
				_, exists = w.blockstakeOutputs[bsoid]
				if exists {
					w.unspentblockstakeoutputs[bsoid] = types.UnspentBlockStakeOutput{
						BlockStakeOutputID: bsoid,
						Indexes: types.BlockStakeOutputIndexes{
							BlockHeight:      blockheight,
							TransactionIndex: uint64(ti),
							OutputIndex:      uint64(i),
						},
						Value:     sfo.Value,
						Condition: sfo.Condition,
					}
				}
				w.historicOutputs[types.OutputID(bsoid)] = historicOutput{
					UnlockHash: uh,
					Value:      sfo.Value,
				}
			}
			if relevant {
				w.processedTransactions = append(w.processedTransactions, pt)
				w.processedTransactionMap[pt.TransactionID] = &w.processedTransactions[len(w.processedTransactions)-1]
			}
		}
	}
}

// ProcessConsensusChange parses a consensus change to update the set of
// confirmed outputs known to the wallet.
func (w *Wallet) ProcessConsensusChange(cc modules.ConsensusChange) {
	if err := w.tg.Add(); err != nil {
		// The wallet should gracefully reject updates from the consensus set
		// or transaction pool that are sent after the wallet's Close method
		// has closed the wallet's ThreadGroup.
		return
	}
	defer w.tg.Done()
	w.mu.Lock()
	defer w.mu.Unlock()
	w.updateConfirmedSet(cc)
	w.revertHistory(cc)
	w.applyHistory(cc)
}

// ReceiveUpdatedUnconfirmedTransactions updates the wallet's unconfirmed
// transaction set.
func (w *Wallet) ReceiveUpdatedUnconfirmedTransactions(txns []types.Transaction, _ modules.ConsensusChange) error {
	if err := w.tg.Add(); err != nil {
		// Gracefully reject transactions if the wallet's Close method has
		// closed the wallet's ThreadGroup already.
		return err
	}
	defer w.tg.Done()
	w.mu.Lock()
	defer w.mu.Unlock()

	w.unconfirmedProcessedTransactions = nil
	for _, txn := range txns {
		// To save on code complexity, relevancy is determined while building
		// up the wallet transaction.
		relevant := false
		pt := modules.ProcessedTransaction{
			Transaction:           txn,
			TransactionID:         txn.ID(),
			ConfirmationHeight:    types.BlockHeight(math.MaxUint64),
			ConfirmationTimestamp: types.Timestamp(math.MaxUint64),
		}
		for _, sci := range txn.CoinInputs {
			output := w.historicOutputs[types.OutputID(sci.ParentID)]
			_, exists := w.keys[output.UnlockHash]
			if exists {
				relevant = true
			} else if _, exists = w.multiSigCoinOutputs[sci.ParentID]; exists {
				// Since we know about every multisig output that is still open and releated,
				// any relevant multisig input must have a parent ID present in the multisig
				// output map.
				relevant = true
				// set "exists" to false since the output is not owned by the wallet.
				exists = false
			}
			pt.Inputs = append(pt.Inputs, modules.ProcessedInput{
				FundType:       types.SpecifierCoinInput,
				WalletAddress:  exists,
				RelatedAddress: output.UnlockHash,
				Value:          output.Value,
			})
		}
		for i, sco := range txn.CoinOutputs {
			uh := sco.Condition.UnlockHash()
			_, exists := w.keys[uh]
			if exists {
				relevant = true
			} else if _, exists = w.multiSigCoinOutputs[txn.CoinOutputID(uint64(i))]; exists {
				// If the coin output is a relevant multisig output, it's ID will already
				// be present in the multisigCoinOutputs map
				relevant = true
				// set "exists" to false since the output is not owned by the wallet.
				exists = false
			}
			pt.Outputs = append(pt.Outputs, modules.ProcessedOutput{
				FundType:       types.SpecifierCoinOutput,
				MaturityHeight: types.BlockHeight(math.MaxUint64),
				WalletAddress:  exists,
				RelatedAddress: uh,
				Value:          sco.Value,
			})
			w.historicOutputs[types.OutputID(txn.CoinOutputID(uint64(i)))] = historicOutput{
				UnlockHash: uh,
				Value:      sco.Value,
			}
		}
		if relevant {
			w.unconfirmedProcessedTransactions = append(w.unconfirmedProcessedTransactions, pt)
		}
	}
	return nil
}
