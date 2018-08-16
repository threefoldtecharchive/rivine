package electrum

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"

	"github.com/rivine/rivine/build"
	"github.com/rivine/rivine/types"
)

// AddressStatus generates the status for an address as per the
// Electrum protocol spec.
//
// Generate the status string. The order is:
//
//  - Per block, sorted in height ascending order:
// 		* miner payout
// 		* transactions ordered by appearance in the block
// 	- Transactionpool transactions:
// 		* transactions with at least one unconfirmed input (block height -1)
// 		* transactions where all inputs are already confirmed (block height 0)
//
// The official spec notes that there is no specific ordering for transactions from
// the transaction pool (other than the difference between the ones with unconfirmed or only
// confirmed inputs). As such, it is possible for an address to have a different status when talking
// to different servers, or even in subsequent calls to the same server.
//
// This means that it is possible for an address to have multiple status strings, which resolve to the same
// status, but only if there are at least 2 transactions waiting in the transaction pool with regards to this address.
func (e *Electrum) AddressStatus(address types.UnlockHash) string {
	// Get confirmed transactions
	txns := e.explorer.UnlockHash(address)

	// Get the heights
	// multiple txns could be at the same height

	txnLocation := make(map[int][]types.TransactionID)
	heights := make([]int, len(txns))
	for i, tx := range txns {
		_, height, _ := e.explorer.Transaction(tx)
		heights[i] = int(height)
		txnLocation[i] = append(txnLocation[i], tx)
	}
	// order heights
	sort.Ints(heights)

	var statusString string

	for _, height := range heights {
		// if there is more than 1 txn in the block, fetch the block and and iterate over the txns
		// to put them in the right order
		if len(txnLocation[height]) > 1 {
			block, exists := e.cs.BlockAtHeight(types.BlockHeight(height))
			if build.DEBUG && !exists {
				build.Critical("Block does not exist in consensus set:", height)
			}
			// if the transaction id is the block id (miner payout), place it first in the ordering
			for _, txid := range txnLocation[height] {
				if txid == types.TransactionID(block.ID()) {
					statusString += fmt.Sprintf("%v:%v:", txid.String(), height)
				}
			}
			// iterate through the remaining transactions in order
			for _, blocktx := range block.Transactions {
				for _, txid := range txnLocation[height] {
					if txid == blocktx.ID() {
						statusString += fmt.Sprintf("%v:%v:", txid.String(), height)
						continue
					}
				}
			}
			continue
		}
		// there is only one transaction for this height
		statusString += fmt.Sprintf("%v:%v:", txnLocation[height][0].String(), height)
	}

	// Now fetch all unconfirmed transactions from the transaction pool
	if e.tp != nil {
		unconfirmedTxList := e.tp.TransactionList()

		// Collect all coin and blockstake outputs from the transaction pool, these are unconfirmed
		// also check every input and output to see if its relevant for our address
		ucoids := []types.CoinOutputID{}
		ubsoids := []types.BlockStakeOutputID{}
		ucos := []types.CoinOutput{}
		ubsos := []types.BlockStakeOutput{}

		// list all transactions which have an output with the given address. while we are at it,
		// also list all outputs and inputs
		relevantTpTransactions := make(map[types.TransactionID]bool)

		for _, tx := range unconfirmedTxList {
			for ucoidIdx := 0; ucoidIdx < len(tx.CoinOutputs); ucoidIdx++ {
				ucoids = append(ucoids, tx.CoinOutputID(uint64(ucoidIdx)))
				ucos = append(ucos, tx.CoinOutputs[ucoidIdx])
				if tx.CoinOutputs[ucoidIdx].Condition.UnlockHash() == address {
					relevantTpTransactions[tx.ID()] = true

				}
			}
			for ubsoidIdx := 0; ubsoidIdx < len(tx.BlockStakeOutputs); ubsoidIdx++ {
				ubsoids = append(ubsoids, tx.BlockStakeOutputID(uint64(ubsoidIdx)))
				ubsos = append(ubsos, tx.BlockStakeOutputs[ubsoidIdx])
				if tx.BlockStakeOutputs[ubsoidIdx].Condition.UnlockHash() == address {
					relevantTpTransactions[tx.ID()] = true

				}
			}
		}

		// Now that we have a list of all outputs, loop over all inputs and check if there is
		// a relevant one
		for _, tx := range unconfirmedTxList {
			isRelevantAndUnconfirmed := false

			// For every coin input in the transaction, check if its in the explorer. If its not,
			// it must be in the transactionpool. If its in the transactionpool, look it up in the
			// previously created lists.
			for ciIdx := 0; ciIdx < len(tx.CoinInputs) && !isRelevantAndUnconfirmed; ciIdx++ {
				co, exists := e.explorer.CoinOutput(tx.CoinInputs[ciIdx].ParentID)
				if !exists {
					for ucoidIdx, ucoid := range ucoids {
						if tx.CoinInputs[ciIdx].ParentID == ucoid {
							if ucos[ucoidIdx].Condition.UnlockHash() == address {
								isRelevantAndUnconfirmed = true
								relevantTpTransactions[tx.ID()] = false
								break
							}
						}
					}
					// Input is already confirmed. If we care for it add the transaction.
					// It could be that
				} else if co.Condition.UnlockHash() == address {
					if _, known := relevantTpTransactions[tx.ID()]; !known {
						relevantTpTransactions[tx.ID()] = true
					}
					// Dont break here, but keep looping, it migh be that we find other
					// unconfirmed inputs
				}
			}

			// Same for blockstake inputs
			for bsiIdx := 0; bsiIdx < len(tx.BlockStakeInputs) && !isRelevantAndUnconfirmed; bsiIdx++ {
				co, exists := e.explorer.BlockStakeOutput(tx.BlockStakeInputs[bsiIdx].ParentID)
				if !exists {
					for ubsoidIdx, ubsoid := range ubsoids {
						if tx.BlockStakeInputs[bsiIdx].ParentID == ubsoid {
							if ubsos[ubsoidIdx].Condition.UnlockHash() == address {
								isRelevantAndUnconfirmed = true
								relevantTpTransactions[tx.ID()] = false
								break
							}
						}
					}
				}
				if co.Condition.UnlockHash() == address {
					if _, known := relevantTpTransactions[tx.ID()]; !known {
						relevantTpTransactions[tx.ID()] = true
					}
				}
			}
		}

		// Now we should have all transactions which are in the transaction pool and
		// known if they are using unconfirmed or confirmed inputs
		// First loop for the ones with unconfirmed inputs
		for txid, inputsConfirmed := range relevantTpTransactions {
			if !inputsConfirmed {
				statusString += fmt.Sprintf("%v:%v:", txid.String(), "-1")
			}
			delete(relevantTpTransactions, txid)
		}
		// and a final loop for the ones which are using confirmed inputs
		for txid := range relevantTpTransactions {
			statusString += fmt.Sprintf("%v:%v:", txid.String(), "0")
		}
	}

	statusHash := sha256.Sum256([]byte(statusString))

	return hex.EncodeToString(statusHash[:])
}
