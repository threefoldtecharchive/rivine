package consensus

import (
	"errors"

	"github.com/rivine/rivine/build"
	"github.com/rivine/rivine/encoding"
	"github.com/rivine/rivine/modules"
	"github.com/rivine/rivine/types"

	"github.com/NebulousLabs/bolt"
)

var (
	errMissingSiacoinOutput       = errors.New("transaction spends a nonexisting siacoin output")
	errMissingSiafundOutput       = errors.New("transaction spends a nonexisting siafund output")
	errSiacoinInputOutputMismatch = errors.New("siacoin inputs do not equal siacoin outputs for transaction")
	errSiafundInputOutputMismatch = errors.New("siafund inputs do not equal siafund outputs for transaction")
	errWrongUnlockConditions      = errors.New("transaction contains incorrect unlock conditions")
)

// validSiacoins checks that the siacoin inputs and outputs are valid in the
// context of the current consensus set.
func validSiacoins(tx *bolt.Tx, t types.Transaction) error {
	scoBucket := tx.Bucket(SiacoinOutputs)
	var inputSum types.Currency
	for _, sci := range t.SiacoinInputs {
		// Check that the input spends an existing output.
		scoBytes := scoBucket.Get(sci.ParentID[:])
		if scoBytes == nil {
			return errMissingSiacoinOutput
		}

		// Check that the unlock conditions match the required unlock hash.
		var sco types.SiacoinOutput
		err := encoding.Unmarshal(scoBytes, &sco)
		if build.DEBUG && err != nil {
			panic(err)
		}
		if sci.UnlockConditions.UnlockHash() != sco.UnlockHash {
			return errWrongUnlockConditions
		}

		inputSum = inputSum.Add(sco.Value)
	}
	if inputSum.Cmp(t.SiacoinOutputSum()) != 0 {
		return errSiacoinInputOutputMismatch
	}
	return nil
}

// validSiafunds checks that the siafund portions of the transaction are valid
// in the context of the consensus set.
func validSiafunds(tx *bolt.Tx, t types.Transaction) (err error) {
	// Compare the number of input siafunds to the output siafunds.
	var siafundInputSum types.Currency
	var siafundOutputSum types.Currency
	for _, sfi := range t.SiafundInputs {
		sfo, err := getSiafundOutput(tx, sfi.ParentID)
		if err != nil {
			return err
		}

		// Check the unlock conditions match the unlock hash.
		if sfi.UnlockConditions.UnlockHash() != sfo.UnlockHash {
			return errWrongUnlockConditions
		}

		siafundInputSum = siafundInputSum.Add(sfo.Value)
	}
	for _, sfo := range t.SiafundOutputs {
		siafundOutputSum = siafundOutputSum.Add(sfo.Value)
	}
	if siafundOutputSum.Cmp(siafundInputSum) != 0 {
		return errSiafundInputOutputMismatch
	}
	return
}

// validTransaction checks that all fields are valid within the current
// consensus state. If not an error is returned.
func validTransaction(tx *bolt.Tx, t types.Transaction) error {
	// StandaloneValid will check things like signatures and properties that
	// should be inherent to the transaction. (storage proof rules, etc.)
	err := t.StandaloneValid(blockHeight(tx))
	if err != nil {
		return err
	}

	// Check that each portion of the transaction is legal given the current
	// consensus set.
	err = validSiacoins(tx, t)
	if err != nil {
		return err
	}
	err = validSiafunds(tx, t)
	if err != nil {
		return err
	}
	return nil
}

// TryTransactionSet applies the input transactions to the consensus set to
// determine if they are valid. An error is returned IFF they are not a valid
// set in the current consensus set. The size of the transactions and the set
// is not checked. After the transactions have been validated, a consensus
// change is returned detailing the diffs that the transaciton set would have.
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
			err := validTransaction(tx, txn)
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
		SiacoinOutputDiffs:        diffHolder.SiacoinOutputDiffs,
		SiafundOutputDiffs:        diffHolder.SiafundOutputDiffs,
		DelayedSiacoinOutputDiffs: diffHolder.DelayedSiacoinOutputDiffs,
		SiafundPoolDiffs:          diffHolder.SiafundPoolDiffs,
	}
	return cc, nil
}
