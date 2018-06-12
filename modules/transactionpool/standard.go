package transactionpool

import (
	"github.com/rivine/rivine/encoding"
	"github.com/rivine/rivine/modules"
	"github.com/rivine/rivine/types"
)

// ValidateTransactionSet validates that all transacitons of a set follow the
// defined standards and are valid within its local context, knowing the height and timestamp of the last block.
// It also ensures that the transaction as well as the transaction set,
// are within an acceptable byte size range, when binary encoded.
func (tp *TransactionPool) ValidateTransactionSet(ts []types.Transaction) error {
	totalSize := 0
	ctx := types.ValidationContext{
		Confirmed:   false,
		BlockHeight: tp.consensusSet.Height(),
	}
	//validate each transaction in the transaction set
	var err error
	for _, t := range ts {
		size := len(encoding.Marshal(t))
		if size > tp.chainCts.TransactionPool.TransactionSizeLimit {
			return modules.ErrLargeTransaction
		}
		totalSize += size
		err = t.ValidateTransaction(ctx, types.TransactionValidationConstants{
			BlockSizeLimit:         tp.chainCts.BlockSizeLimit,
			ArbitraryDataSizeLimit: tp.chainCts.ArbitraryDataSizeLimit,
			MinimumMinerFee:        tp.chainCts.MinimumTransactionFee,
		})
		if err != nil {
			return err
		}
	}
	if totalSize > tp.chainCts.TransactionPool.TransactionSetSizeLimit {
		return modules.ErrLargeTransactionSet
	}
	return nil
}
