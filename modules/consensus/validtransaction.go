package consensus

import (
	"errors"

	"github.com/threefoldtech/rivine/build"
	"github.com/threefoldtech/rivine/modules"
	"github.com/threefoldtech/rivine/pkg/encoding/siabin"
	"github.com/threefoldtech/rivine/types"

	"github.com/rivine/bbolt"
)

// validCoins checks that the coin inputs and outputs are valid in the
// context of the current consensus set, meaning that total coin input sum
// equals the total coin output sum, as well as the fact that all conditions referenced coin outputs,
// have been correctly fulfilled by the child coin inputs.
func validCoins(tx *bolt.Tx, t types.Transaction, blockHeight types.BlockHeight, blockTimestamp types.Timestamp) (err error) {
	coinInputs := make(map[types.CoinOutputID]types.CoinOutput, len(t.CoinInputs))
	for _, sci := range t.CoinInputs {
		// Check that the input spends an existing output.
		scoBytes := tx.Bucket(CoinOutputs).Get(sci.ParentID[:])
		if scoBytes == nil {
			continue // ignore, up to transaction to define if this is invalid
		}
		// unmarshall the output bytes
		var sco types.CoinOutput
		err = siabin.Unmarshal(scoBytes, &sco)
		if build.DEBUG && err != nil {
			panic(err)
		}
		coinInputs[sci.ParentID] = sco
	}
	return t.ValidateCoinOutputs(types.FundValidationContext{
		BlockHeight: blockHeight,
		BlockTime:   blockTimestamp,
	}, coinInputs)
}

// validBlockStakes checks that the blockstake portions of the transaction are valid
// in the context of the consensus set, meaning that block stake input sum
// equals the block stake output sum, as well as the fact that all conditions
// of referenced block stake outputs, have been correctly fulfilled by the child block stkae inputs.
func validBlockStakes(tx *bolt.Tx, t types.Transaction, blockHeight types.BlockHeight, blockTimestamp types.Timestamp) (err error) {
	blockStakeInputs := make(map[types.BlockStakeOutputID]types.BlockStakeOutput, len(t.BlockStakeInputs))
	for _, bsi := range t.BlockStakeInputs {
		// Check that the input spends an existing output.
		bso, err := getBlockStakeOutput(tx, bsi.ParentID)
		if err == errNilItem {
			continue // ignore, up to transaction to define if this is invalid
		}
		blockStakeInputs[bsi.ParentID] = bso
	}
	return t.ValidateBlockStakeOutputs(types.FundValidationContext{
		BlockHeight: blockHeight,
		BlockTime:   blockTimestamp,
	}, blockStakeInputs)
}

// validTransaction checks that all fields are valid within the current
// consensus state. If not an error is returned.
func validTransaction(tx *bolt.Tx, t types.Transaction, constants types.TransactionValidationConstants, blockHeight types.BlockHeight, blockTimestamp types.Timestamp, isBlockCreatingTx bool) error {
	// StandaloneValid will check things like signatures and properties that
	// should be inherent to the transaction. (storage proof rules, etc.)
	err := t.ValidateTransaction(types.ValidationContext{
		Confirmed:         true,
		BlockHeight:       blockHeight,
		BlockTime:         blockTimestamp,
		IsBlockCreatingTx: isBlockCreatingTx,
	}, constants)
	if err != nil {
		return err
	}

	// Check that each portion of the transaction is legal given the current
	// consensus set.
	err = validCoins(tx, t, blockHeight, blockTimestamp)
	if err != nil {
		return err
	}
	err = validBlockStakes(tx, t, blockHeight, blockTimestamp)
	if err != nil {
		return err
	}
	return nil
}

// TryTransactionSet applies the input transactions to the consensus set to
// determine if they are valid. An error is returned IFF they are not a valid
// set in the current consensus set. The size of the transactions and the set
// is not checked. After the transactions have been validated, a consensus
// change is returned detailing the diffs that the transaction set would have.
func (cs *ConsensusSet) TryTransactionSet(txns []types.Transaction) (modules.ConsensusChange, error) {
	err := cs.tg.Add()
	if err != nil {
		return modules.ConsensusChange{}, err
	}
	defer cs.tg.Done()
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	// applyTransaction will apply the diffs from a transaction and store them
	// in a block node. diffHolder is the blockNode that tracks the temporary
	// changes. At the end of the function, all changes that were made to the
	// consensus set get reverted.
	diffHolder := new(processedBlock)

	// Boltdb will only roll back a tx if an error is returned. In the case of
	// TryTransactionSet, we want to roll back the tx even if there is no
	// error. So errSuccess is returned. An alternate method would be to
	// manually manage the tx instead of using 'Update', but that has safety
	// concerns and is more difficult to implement correctly.
	errSuccess := errors.New("success")
	err = cs.db.Update(func(tx *bolt.Tx) error {
		diffHolder.Height = blockHeight(tx)
		blockTime, err := blockTimeStamp(tx, diffHolder.Height)
		if err != nil {
			return err
		}
		for _, txn := range txns {
			// a transaction can only be "block creating" in the context of a block,
			// which we don't have here, so just pass in false for the "isBlockCreatingTx"
			// argument. In other words, a block creating transaction can never be part
			// of a transaction pool and must be inserted when the block is actually created
			err := validTransaction(tx, txn, types.TransactionValidationConstants{
				BlockSizeLimit:         cs.chainCts.BlockSizeLimit,
				ArbitraryDataSizeLimit: cs.chainCts.ArbitraryDataSizeLimit,
				MinimumMinerFee:        cs.chainCts.MinimumTransactionFee,
			}, diffHolder.Height, blockTime, false)
			if err != nil {
				cs.log.Printf("WARN: try-out tx %v is invalid: %v", txn.ID(), err)
				return err
			}
			applyTransaction(tx, diffHolder, txn)
		}
		return errSuccess
	})
	if err != errSuccess {
		return modules.ConsensusChange{}, err
	}
	cc := modules.ConsensusChange{
		CoinOutputDiffs:       diffHolder.CoinOutputDiffs,
		BlockStakeOutputDiffs: diffHolder.BlockStakeOutputDiffs,
	}
	return cc, nil
}
