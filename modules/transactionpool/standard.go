package transactionpool

import (
	"fmt"

	"github.com/threefoldtech/rivine/encoding"
	"github.com/threefoldtech/rivine/modules"
	"github.com/threefoldtech/rivine/types"
)

// ValidateTransactionSet validates that all transacitons of a set follow the
// defined standards and are valid within its local context, knowing the height and timestamp of the last block.
// It also ensures that the transaction as well as the transaction set,
// are within an acceptable byte size range, when binary encoded.
func (tp *TransactionPool) ValidateTransactionSet(ts []types.Transaction) error {
	totalSize := 0
	blockHeight := tp.consensusSet.Height()
	block, ok := tp.consensusSet.BlockAtHeight(blockHeight)
	if !ok {
		return fmt.Errorf("failed to fetch block at height %d", blockHeight)
	}
	ctx := types.ValidationContext{
		Confirmed:   false,
		BlockHeight: blockHeight,
		BlockTime:   block.Timestamp,
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
