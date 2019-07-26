package transactionpool

import (
	"github.com/threefoldtech/rivine/modules"
	"github.com/threefoldtech/rivine/pkg/encoding/siabin"
	"github.com/threefoldtech/rivine/types"
)

// ValidateTransactionSetSize validates that all transacitons are within the size limits
// as defined by the transaction pool constants.
func (tp *TransactionPool) ValidateTransactionSetSize(ts []types.Transaction) error {
	totalSize := 0
	//validate size of individual and all transactions
	for _, t := range ts {
		size := len(siabin.Marshal(t))
		if size > tp.chainCts.TransactionPool.TransactionSizeLimit {
			return modules.ErrLargeTransaction
		}
		totalSize += size
	}
	if totalSize > tp.chainCts.TransactionPool.TransactionSetSizeLimit {
		return modules.ErrLargeTransactionSet
	}
	return nil
}
