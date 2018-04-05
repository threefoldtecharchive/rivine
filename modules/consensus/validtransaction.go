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
// context of the current consensus set.
func validCoins(tx *bolt.Tx, t types.Transaction) error {
	scoBucket := tx.Bucket(CoinOutputs)
	var inputSum types.Currency
	for _, sci := range t.CoinInputs {
		// Check that the input spends an existing output.
		scoBytes := scoBucket.Get(sci.ParentID[:])
		if scoBytes == nil {
			return errMissingCoinOutput
		}

		// Check that the unlock conditions match the required unlock hash.
		var sco types.CoinOutput
		err := encoding.Unmarshal(scoBytes, &sco)
		if build.DEBUG && err != nil {
			panic(err)
		}
		if sci.Unlocker.UnlockHash() != sco.UnlockHash {
			return errWrongUnlockConditions
		}

		inputSum = inputSum.Add(sco.Value)
	}
	if !inputSum.Equals(t.CoinOutputSum()) {
		return errSiacoinInputOutputMismatch
	}
	return nil
}

// validBlockStakes checks that the blockstake portions of the transaction are valid
// in the context of the consensus set.
func validBlockStakes(tx *bolt.Tx, t types.Transaction) (err error) {
	// Compare the number of input blockstake to the output blockstake.
	var blockstakeInputSum types.Currency
	var blockstakeOutputSum types.Currency
	for _, sfi := range t.BlockStakeInputs {
		sfo, err := getBlockStakeOutput(tx, sfi.ParentID)
		if err != nil {
			return err
		}

		// Check the unlock conditions match the unlock hash.
		if sfi.Unlocker.UnlockHash() != sfo.UnlockHash {
			return errWrongUnlockConditions
		}

		blockstakeInputSum = blockstakeInputSum.Add(sfo.Value)
	}
	for _, sfo := range t.BlockStakeOutputs {
		blockstakeOutputSum = blockstakeOutputSum.Add(sfo.Value)
	}
	if !blockstakeOutputSum.Equals(blockstakeInputSum) {
		return errBlockStakeInputOutputMismatch
	}
	return
}

// validTransaction checks that all fields are valid within the current
// consensus state. If not an error is returned.
func validTransaction(tx *bolt.Tx, t types.Transaction, blockSizeLimit uint64) error {
	// StandaloneValid will check things like signatures and properties that
	// should be inherent to the transaction. (storage proof rules, etc.)
	err := t.StandaloneValid(blockHeight(tx), blockSizeLimit)
	if err != nil {
		return err
	}

	// Check that each portion of the transaction is legal given the current
	// consensus set.
	err = validCoins(tx, t)
	if err != nil {
		return err
	}
	err = validBlockStakes(tx, t)
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
		for _, txn := range txns {
			err := validTransaction(tx, txn, cs.chainCts.BlockSizeLimit)
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
