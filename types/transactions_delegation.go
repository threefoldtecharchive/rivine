package types

// These Specifiers are used internally when calculating a type's ID. See
// Specifier for more details.
var (
	SpecifierDelegationTransaction = Specifier{'d', 'e', 'l', 'e', 'g', 'a', 't', 'i', 'o', 'n'}
)

const (
	// TransactionVersionDelegation defines the special transaction used for
	// temporarily delegating a blockstake output to a third party
	TransactionVersionDelegation TransactionVersion = 3
)

type (
	// DelegationTransaction defines the transaction (with version 0x03)
	// used to allow a third party to use these blockstakes to create new blocks
	DelegationTransaction struct {
		// Reference unlocks a blockstake output to prove ownership, but does not consume it
		Reference BlockStakeInput
		// Delegation is the condition which needs to be unlocked to use the delegated blockstakes
		Delegation BlockStakeOutput
	}

	// DelegationTransactionExtension defines the DelegationTransaction Extension Data
	DelegationTransactionExtension struct {
		// Reference unlocks a blockstake output to prove ownership, but does not consume it
		Reference BlockStakeInput
		// Delegation is the condition which needs to be unlocked to use the delegated blockstakes
		Delegation BlockStakeOutput
	}
)
