package transactionpool

import (
	"github.com/threefoldtech/rivine/modules"
	"github.com/threefoldtech/rivine/types"
)

type (
	transactionCache struct {
		cache map[TransactionSetID]*broadcastInfo
	}

	broadcastInfo struct {
		// originalSubmit is the original block height in which a transacton was submitted
		// in our cache
		originalSubmit types.BlockHeight
		// broadcasts is the amount of times we have rebroadcast a transaction
		broadcasts uint32
	}
)

func newTransactionCache() transactionCache {
	return transactionCache{
		cache: make(map[TransactionSetID]*broadcastInfo),
	}
}

// add a new transaction ID to the cache
func (tc *transactionCache) add(id TransactionSetID, currentHeight types.BlockHeight) {
	// Because the update method of the transaction pool purges the tp, and then adds all the
	// transactions again which it finds are still unconfirmed, we need to check if the id we
	// add is not already known to us
	if _, exists := tc.cache[id]; !exists {
		tc.cache[id] = &broadcastInfo{originalSubmit: currentHeight, broadcasts: 0}
	}
}

// delete ensures the transaction ID is no longer present in the cache. If it is not present in the first place,
// no action is taken.
func (tc *transactionCache) delete(id TransactionSetID) {
	delete(tc.cache, id)
}

func (tc *transactionCache) getTransactionsToBroadcast(height types.BlockHeight) []TransactionSetID {
	txnIDs := []TransactionSetID{}
	for id, info := range tc.cache {
		if info.shouldRebroadcast(height) {
			txnIDs = append(txnIDs, id)
		}
	}
	return txnIDs
}

// shouldRebroadcast decides if the txninfo indicates that the transaction should
// be rebroadcast. If this should be the case, the broadcasts coutner will be updated.
func (bci *broadcastInfo) shouldRebroadcast(height types.BlockHeight) bool {
	if (height-bci.originalSubmit)%modules.TransactionPoolRebroadcastDelay == 0 &&
		bci.broadcasts <= modules.TransactionPoolMaxRebroadcasts {
		bci.broadcasts++
		return true
	}
	return false
}
