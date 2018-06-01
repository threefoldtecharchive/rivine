package transactionpool

import (
	"github.com/rivine/rivine/encoding"
	"github.com/rivine/rivine/modules"
	"github.com/rivine/rivine/types"
)

// standard.go adds extra rules to transactions which help preserve network
// health and provides flexibility for future soft forks and tweaks to the
// network.
//
// Rule: Transaction size is limited
//		There is a DoS vector where large transactions can both contain many
//		signatures, and have each signature's CoveredFields object cover a
//		unique but large portion of the transaction. A 1mb transaction could
//		force a verifier to hash very large volumes of data, which takes a long
//		time on nonspecialized hardware.
//
// Rule: Foreign signature algorithms are rejected.
//		There are plans to add newer, faster signature algorithms to Sia as the
//		project matures and the need for increased verification speed grows.
//		Foreign signatures are allowed into the blockchain, where they are
//		accepted as valid. Hoewver, if there has been a soft-fork, the foreign
//		signatures might actually be invalid. This rule protects legacy miners
//		from including potentially invalid transactions in their blocks.
//
// Rule: The types of allowed arbitrary data are limited
//		The arbitrary data field can be used to orchestrate soft-forks to Sia
//		that add features. Legacy miners are at risk of creating invalid blocks
//		if they include arbitrary data which has meanings that the legacy miner
//		doesn't understand.
//
// Rule: The transaction set size is limited.
//		A group of dependent transactions cannot exceed 100kb to limit how
//		quickly the transaction pool can be filled with new transactions.

// IsStandardTransaction enforces extra rules such as a transaction size limit.
// These rules can be altered without disrupting consensus.
func (tp *TransactionPool) IsStandardTransaction(t types.Transaction) error {
	// check if the transaction is standard (e.g. based on its version)
	err := t.IsStandardTransaction()
	if err != nil {
		return err
	}

	// Check that the size of the transaction does not exceed the standard
	// established in Standard.md. Larger transactions are a DOS vector,
	// because someone can fill a large transaction with a bunch of signatures
	// that require hashing the entire transaction. Several hundred megabytes
	// of hashing can be required of a verifier. Enforcing this rule makes it
	// more difficult for attackers to exploid this DOS vector, though a miner
	// with sufficient power could still create unfriendly blocks.
	if len(encoding.Marshal(t)) > modules.TransactionSizeLimit {
		return modules.ErrLargeTransaction
	}

	// check if all condtions are standard
	for _, sco := range t.CoinOutputs {
		err = sco.Condition.IsStandardCondition()
		if err != nil {
			return err
		}
	}
	for _, sfo := range t.BlockStakeOutputs {
		err = sfo.Condition.IsStandardCondition()
		if err != nil {
			return err
		}
	}
	// check if all fulfillments are standard
	for _, sci := range t.CoinInputs {
		err = sci.Fulfillment.IsStandardFulfillment()
		if err != nil {
			return err
		}
	}
	for _, sfi := range t.BlockStakeInputs {
		err = sfi.Fulfillment.IsStandardFulfillment()
		if err != nil {
			return err
		}
	}

	return nil
}

// IsStandardTransactionSet checks that all transacitons of a set follow the
// IsStandard guidelines, and that the set as a whole follows the guidelines as
// well.
func (tp *TransactionPool) IsStandardTransactionSet(ts []types.Transaction) error {
	// Check that the set is a reasonable size.
	totalSize := 0
	for i := range ts {
		totalSize += len(encoding.Marshal(ts[i]))
		if totalSize > modules.TransactionSetSizeLimit {
			return modules.ErrLargeTransactionSet
		}
	}

	// Check that each transaction is acceptable.
	for i := range ts {
		err := tp.IsStandardTransaction(ts[i])
		if err != nil {
			return err
		}
	}
	return nil
}
