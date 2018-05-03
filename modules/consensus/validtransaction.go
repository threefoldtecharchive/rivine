package consensus

import (
	"errors"

	"github.com/rivine/rivine/build"
	"github.com/rivine/rivine/encoding"
	"github.com/rivine/rivine/modules"
	"github.com/rivine/rivine/types"

	"github.com/rivine/bbolt"
)

var (
	errMissingCoinOutput             = errors.New("transaction spends a nonexisting coin output")
	errMissingBlockStakeOutput       = errors.New("transaction spends a nonexisting blockstake output")
	errSiacoinInputOutputMismatch    = errors.New("coin inputs do not equal coin outputs for transaction")
	errBlockStakeInputOutputMismatch = errors.New("blockstake inputs do not equal blockstake outputs for transaction")
	errWrongUnlockConditions         = errors.New("transaction contains incorrect unlock conditions")
)

// validCoins checks that the coin inputs and outputs are valid in the
// context of the current consensus set, meaning that total coin input sum
// equals the total coin output sum, as well as the fact that all conditions referenced coin outputs,
// have been correctly fulfilled by the child coin inputs.
func validCoins(tx *bolt.Tx, t types.Transaction, blockHeight types.BlockHeight, blockTimestamp types.Timestamp) (err error) {
	scoBucket := tx.Bucket(CoinOutputs)
	var inputSum types.Currency
	for inputIndex, sci := range t.CoinInputs {
		// Check that the input spends an existing output.
		scoBytes := scoBucket.Get(sci.ParentID[:])
		if scoBytes == nil {
			return errMissingCoinOutput
		}

		// unmarshall the output bytes
		var sco types.CoinOutput
		err = encoding.Unmarshal(scoBytes, &sco)
		if build.DEBUG && err != nil {
			panic(err)
		}

		// check if the referenced output's condition has been fulfilled
		err = sco.Condition.Fulfill(sci.Fulfillment, types.FulfillContext{
			InputIndex:  uint64(inputIndex),
			BlockHeight: blockHeight,
			BlockTime:   blockTimestamp,
			Transaction: t,
		})
		if err != nil {
			return
		}

		inputSum = inputSum.Add(sco.Value)
	}
	if !inputSum.Equals(t.CoinOutputSum()) {
		return errSiacoinInputOutputMismatch
	}
	return nil
}

// validBlockStakes checks that the blockstake portions of the transaction are valid
// in the context of the consensus set, meaning that block stake input sum
// equals the block stake output sum, as well as the fact that all conditions
// of referenced block stake outputs, have been correctly fulfilled by the child block stkae inputs.
func validBlockStakes(tx *bolt.Tx, t types.Transaction, blockHeight types.BlockHeight, blockTimestamp types.Timestamp) (err error) {
	// Compare the number of input blockstake to the output blockstake.
	var blockstakeInputSum types.Currency
	var blockstakeOutputSum types.Currency
	var bso types.BlockStakeOutput
	for inputIndex, bsi := range t.BlockStakeInputs {
		bso, err = getBlockStakeOutput(tx, bsi.ParentID)
		if err != nil {
			return
		}

		// check if the referenced output's condition has been fulfilled
		err = bso.Condition.Fulfill(bsi.Fulfillment, types.FulfillContext{
			InputIndex:  uint64(inputIndex),
			BlockHeight: blockHeight,
			BlockTime:   blockTimestamp,
			Transaction: t,
		})
		if err != nil {
			return
		}

		blockstakeInputSum = blockstakeInputSum.Add(bso.Value)
	}
	for _, bso := range t.BlockStakeOutputs {
		blockstakeOutputSum = blockstakeOutputSum.Add(bso.Value)
	}
	if !blockstakeOutputSum.Equals(blockstakeInputSum) {
		return errBlockStakeInputOutputMismatch
	}
	return
}

// validTransaction checks that all fields are valid within the current
// consensus state. If not an error is returned.
func validTransaction(tx *bolt.Tx, t types.Transaction, blockSizeLimit uint64, blockHeight types.BlockHeight, blockTimestamp types.Timestamp) error {
	// StandaloneValid will check things like signatures and properties that
	// should be inherent to the transaction. (storage proof rules, etc.)
	err := t.ValidateTransaction(blockSizeLimit)
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
			err := validTransaction(tx, txn, cs.chainCts.BlockSizeLimit, diffHolder.Height, blockTime)
			if err != nil {
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
