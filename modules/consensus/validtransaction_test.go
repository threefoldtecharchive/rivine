package consensus

import (
	"testing"

	"github.com/rivine/rivine/types"

	"github.com/NebulousLabs/bolt"
)

// TestTryValidTransactionSet submits a valid transaction set to the
// TryTransactionSet method.
func TestTryValidTransactionSet(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	t.Parallel()
	cst, err := createConsensusSetTester("TestTryValidTransactionSet")
	if err != nil {
		t.Fatal(err)
	}
	defer cst.Close()
	initialHash := cst.cs.dbConsensusChecksum()

	// Try a valid transaction.
	_, err = cst.wallet.SendCoins(types.NewCurrency64(1), types.UnlockHash{})
	if err != nil {
		t.Fatal(err)
	}
	txns := cst.tpool.TransactionList()
	cc, err := cst.cs.TryTransactionSet(txns)
	if err != nil {
		t.Error(err)
	}
	if cst.cs.dbConsensusChecksum() != initialHash {
		t.Error("TryTransactionSet did not resotre order")
	}
	if len(cc.CoinOutputDiffs) == 0 {
		t.Error("consensus change is missing diffs after verifying a transction clump")
	}
}

// TestTryInvalidTransactionSet submits an invalid transaction set to the
// TryTransaction method.
func TestTryInvalidTransactionSet(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	t.Parallel()
	cst, err := createConsensusSetTester("TestTryInvalidTransactionSet")
	if err != nil {
		t.Fatal(err)
	}
	defer cst.Close()
	initialHash := cst.cs.dbConsensusChecksum()

	// Try a valid transaction followed by an invalid transaction.
	_, err = cst.wallet.SendCoins(types.NewCurrency64(1), types.UnlockHash{})
	if err != nil {
		t.Fatal(err)
	}
	txns := cst.tpool.TransactionList()
	txn := types.Transaction{
		CoinInputs: []types.CoinInput{{}},
	}
	txns = append(txns, txn)
	cc, err := cst.cs.TryTransactionSet(txns)
	if err == nil {
		t.Error("bad transaction survived filter")
	}
	if cst.cs.dbConsensusChecksum() != initialHash {
		t.Error("TryTransactionSet did not restore order")
	}
	if len(cc.CoinOutputDiffs) != 0 {
		t.Error("consensus change was not empty despite an error being returned")
	}
}

// TestValidSiacoins probes the validSiacoins method of the consensus set.
func TestValidSiacoins(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	t.Parallel()
	cst, err := createConsensusSetTester("TestValidSiacoins")
	if err != nil {
		t.Fatal(err)
	}
	defer cst.Close()

	// Create a transaction pointing to a nonexistent siacoin output.
	txn := types.Transaction{
		CoinInputs: []types.CoinInput{{}},
	}
	err = cst.cs.db.View(func(tx *bolt.Tx) error {
		err := validCoins(tx, txn)
		if err != errMissingCoinOutput {
			t.Fatal(err)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	// Create a transaction with invalid unlock conditions.
	scoid, _, err := cst.cs.getArbCoinOutput()
	if err != nil {
		t.Fatal(err)
	}
	txn = types.Transaction{
		CoinInputs: []types.CoinInput{{
			ParentID: scoid,
		}},
	}
	err = cst.cs.db.View(func(tx *bolt.Tx) error {
		err := validCoins(tx, txn)
		if err != errWrongUnlockConditions {
			t.Fatal(err)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	// Create a txn with more outputs than inputs.
	txn = types.Transaction{
		CoinOutputs: []types.CoinOutput{{
			Value: types.NewCurrency64(1),
		}},
	}
	err = cst.cs.db.View(func(tx *bolt.Tx) error {
		err := validCoins(tx, txn)
		if err != errSiacoinInputOutputMismatch {
			t.Fatal(err)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

/*
// TestValidSiafunds probes the validSiafunds mthod of the consensus set.
func TestValidSiafunds(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	cst, err := createConsensusSetTester("TestValidSiafunds")
	if err != nil {
		t.Fatal(err)
	}
	defer cst.closeCst()

	// Create a transaction pointing to a nonexistent siafund output.
	txn := types.Transaction{
		SiafundInputs: []types.SiafundInput{{}},
	}
	err = cst.cs.validSiafunds(txn)
	if err != ErrMissingSiafundOutput {
		t.Error(err)
	}

	// Create a transaction with invalid unlock conditions.
	var sfoid types.SiafundOutputID
	cst.cs.db.forEachSiafundOutputs(func(mapSfoid types.SiafundOutputID, sfo types.SiafundOutput) {
		sfoid = mapSfoid
		// pointless to do this but I can't think of a better way.
	})
	txn = types.Transaction{
		SiafundInputs: []types.SiafundInput{{
			ParentID:         sfoid,
			UnlockConditions: types.UnlockConditions{Timelock: 12345}, // avoid collisions with existing outputs
		}},
	}
	err = cst.cs.validSiafunds(txn)
	if err != ErrWrongUnlockConditions {
		t.Error(err)
	}

	// Create a transaction with more outputs than inputs.
	txn = types.Transaction{
		SiafundOutputs: []types.SiafundOutput{{
			Value: types.NewCurrency64(1),
		}},
	}
	err = cst.cs.validSiafunds(txn)
	if err != ErrSiafundInputOutputMismatch {
		t.Error(err)
	}
}

// TestValidTransaction probes the validTransaction method of the consensus
// set.
func TestValidTransaction(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	cst, err := createConsensusSetTester("TestValidTransaction")
	if err != nil {
		t.Fatal(err)
	}
	defer cst.closeCst()

	// Create a transaction that is not standalone valid.
	txn := types.Transaction{
		FileContracts: []types.FileContract{{
			WindowStart: 0,
		}},
	}
	err = cst.cs.validTransaction(txn)
	if err == nil {
		t.Error("transaction is valid")
	}

	// Create a transaction with invalid siacoins.
	txn = types.Transaction{
		SiacoinInputs: []types.SiacoinInput{{}},
	}
	err = cst.cs.validTransaction(txn)
	if err == nil {
		t.Error("transaction is valid")
	}

	// Create a transaction with invalid storage proofs.
	txn = types.Transaction{
		StorageProofs: []types.StorageProof{{}},
	}
	err = cst.cs.validTransaction(txn)
	if err == nil {
		t.Error("transaction is valid")
	}

	// Create a transaction with invalid file contract revisions.
	txn = types.Transaction{
		FileContractRevisions: []types.FileContractRevision{{
			NewWindowStart: 5000,
			NewWindowEnd:   5005,
			ParentID:       types.FileContractID{},
		}},
	}
	err = cst.cs.validTransaction(txn)
	if err == nil {
		t.Error("transaction is valid")
	}

	// Create a transaction with invalid siafunds.
	txn = types.Transaction{
		SiafundInputs: []types.SiafundInput{{}},
	}
	err = cst.cs.validTransaction(txn)
	if err == nil {
		t.Error("transaction is valid")
	}
}
*/
