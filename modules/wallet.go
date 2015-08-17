package modules

import (
	"bytes"
	"errors"

	"github.com/NebulousLabs/entropy-mnemonics"

	"github.com/NebulousLabs/Sia/crypto"
	"github.com/NebulousLabs/Sia/types"
)

const (
	WalletDir = "wallet"

	SeedChecksumSize = 6

	PublicKeysPerSeed = 2500
)

var (
	ErrBadEncryptionKey = errors.New("provided encryption key is incorrect")
	ErrLowBalance       = errors.New("insufficient balance")
	ErrLockedWallet     = errors.New("wallet must be unlocked before it can be used")
)

type (
	// AddressSeed is cryptographic entropy that is used to derive spendable
	// wallet addresses.
	Seed [crypto.EntropySize]byte

	// WalletTransactionID is a unique identifier for a wallet transaction.
	WalletTransactionID crypto.Hash

	WalletTransaction struct {
		TransactionID         types.TransactionID `json:"transactionid"`
		ConfirmationHeight    types.BlockHeight   `json:"confirmationheight"`
		ConfirmationTimestamp types.Timestamp     `json:"confirmationtimestamp"`

		FundType       types.Specifier  `json:"fundtype"`
		OutputID       types.OutputID   `json:"outputid"`
		RelatedAddress types.UnlockHash `json:"relatedaddress"`
		Value          types.Currency   `json:"value"`
	}

	// TransactionBuilder is used to construct custom transactions. A transaction
	// builder is intialized via 'RegisterTransaction' and then can be modified by
	// adding funds or other fields. The transaction is completed by calling
	// 'Sign', which will sign all inputs added via the 'FundSiacoins' or
	// 'FundSiafunds' call. All modifications are additive.
	//
	// Parents of the transaction are kept in the transaction builder. A parent is
	// any unconfirmed transaction that is required for the child to be valid.
	//
	// Transaction builders are not thread safe.
	TransactionBuilder interface {
		// FundSiacoins will add a siacoin input of exaclty 'amount' to the
		// transaction. A parent transaction may be needed to achieve an input
		// with the correct value. The siacoin input will not be signed until
		// 'Sign' is called on the transaction builder. The expectation is that
		// the transaction will be completed and broadcast within a few hours.
		// Longer risks double-spends, as the wallet will assume that the
		// transaction failed.
		FundSiacoins(amount types.Currency) error

		// FundSiafunds will add a siafund input of exaclty 'amount' to the
		// transaction. A parent transaction may be needed to achieve an input
		// with the correct value. The siafund input will not be signed until
		// 'Sign' is called on the transaction builder. Any siacoins that are
		// released by spending the siafund outputs will be sent to another
		// address owned by the wallet. The expectation is that the transaction
		// will be completed and broadcast within a few hours. Longer risks
		// double-spends, because the wallet will assume the transcation
		// failed.
		FundSiafunds(amount types.Currency) error

		// AddMinerFee adds a miner fee to the transaction, returning the index
		// of the miner fee within the transaction.
		AddMinerFee(fee types.Currency) uint64

		// AddSiacoinInput adds a siacoin input to the transaction, returning
		// the index of the siacoin input within the transaction. When 'Sign'
		// gets called, this input will be left unsigned.
		AddSiacoinInput(types.SiacoinInput) uint64

		// AddSiacoinOutput adds a siacoin output to the transaction, returning
		// the index of the siacoin output within the transaction.
		AddSiacoinOutput(types.SiacoinOutput) uint64

		// AddFileContract adds a file contract to the transaction, returning
		// the index of the file contract within the transaction.
		AddFileContract(types.FileContract) uint64

		// AddFileContractRevision adds a file contract revision to the
		// transaction, returning the index of the file contract revision
		// within the transaction. When 'Sign' gets called, this revision will
		// be left unsigned.
		AddFileContractRevision(types.FileContractRevision) uint64

		// AddStorageProof adds a storage proof to the transaction, returning
		// the index of the storage proof within the transaction.
		AddStorageProof(types.StorageProof) uint64

		// AddSiafundInput adds a siafund input to the transaction, returning
		// the index of the siafund input within the transaction. When 'Sign'
		// is called, this input will be left unsigned.
		AddSiafundInput(types.SiafundInput) uint64

		// AddSiafundOutput adds a siafund output to the transaction, returning
		// the index of the siafund output within the transaction.
		AddSiafundOutput(types.SiafundOutput) uint64

		// AddArbitraryData adds arbitrary data to the transaction, returning
		// the index of the data within the transaction.
		AddArbitraryData(arb []byte) uint64

		// AddTransactionSignature adds a transaction signature to the
		// transaction, returning the index of the signature within the
		// transaction. The signature should already be valid, and shouldn't
		// sign any of the inputs that were added by calling 'FundSiacoins' or
		// 'FundSiafunds'.
		AddTransactionSignature(types.TransactionSignature) uint64

		// Sign will sign any inputs added by 'FundSiacoins' or 'FundSiafunds'
		// and return a transaction set that contains all parents prepended to
		// the transaction. If more fields need to be added, a new transaction
		// builder will need to be created.
		//
		// If the whole transaction flag is set to true, then the whole
		// transaction flag will be set in the covered fields object. If the
		// whole transaction flag is set to false, then the covered fields
		// object will cover all fields that have already been added to the
		// transaction, but will also leave room for more fields to be added.
		Sign(wholeTransaction bool) ([]types.Transaction, error)

		// View returns the incomplete transaction along with all of its
		// parents.
		View() (txn types.Transaction, parents []types.Transaction)
	}

	// Wallet stores and manages siacoins and siafunds. The wallet file is
	// encrypted using a user-specified password. Common addresses are all
	// dervied from a single address seed.
	Wallet interface {
		// Encrypted returns whether or not the wallet has been encrypted yet.
		// After being encrypted for the first time, the wallet can only be
		// unlocked using the encryption password.
		Encrypted() bool

		// Encrypt will encrypt the wallet using the input key. Upon
		// encryption, a primary seed will be created for the wallet (no seed
		// exists prior to this point). If the key is blank, then the hash of
		// the seed that is generated will be used as the key.
		//
		// Encrypt can only be called once throughout the life of the wallet,
		// and will return an error on subsequent calls (even after restarting
		// the wallet). To reset the wallet, the wallet files must be moved to
		// a different directory or deleted.
		Encrypt(masterKey crypto.TwofishKey) (Seed, error)

		// Unlocked returns true if the wallet is currently unlocked, false
		// otherwise.
		Unlocked() bool

		// Lock deletes all keys in memory and prevents the wallet from being
		// used to spend coins or extract keys until 'Unlock' is called.
		Lock() error

		// Unlock must be called before the wallet is usable. All wallets and
		// wallet seeds are encrypted by default, and the wallet will not know
		// which addresses to watch for on the blockchain until unlock has been
		// called.
		//
		// All items in the wallet are encrypted using different keys which are
		// derived from the master key.
		Unlock(masterKey crypto.TwofishKey) error

		// AllSeeds returns all of the seeds that are being tracked by the
		// wallet, including the primary seed. Only the primary seed is used to
		// generate new addresses, but the wallet can spend funds sent to
		// public keys generated by any of the seeds returned.
		AllSeeds() ([]Seed, error)

		// PrimarySeed returns the current primary seed of the wallet,
		// unencrypted, with an int indicating how many addresses have been
		// consumed.
		PrimarySeed() (Seed, uint64, error)

		// NextAddress returns a new coin addresses generated from the
		// primary seed.
		NextAddress() (types.UnlockConditions, error)

		// RecoverSeed will recreate a wallet file using the recovery phrase.
		// RecoverSeed only needs to be called if the original seed file or
		// encryption password was lost. The master key is used encrypt the
		// recovery seed before saving it to disk.
		RecoverSeed(crypto.TwofishKey, Seed) error

		// RecoverFile will read a file with keys and add them to the wallet.

		// CreateBackup will create a backup of the wallet at the provided
		// filepath. The backup will have all seeds and keys.
		CreateBackup(string) error

		// ConfirmedBalance returns the confirmed balance of the wallet, minus
		// any outgoing transactions. ConfirmedBalance will include unconfirmed
		// refund transacitons.
		ConfirmedBalance() (siacoinBalance types.Currency, siafundBalance types.Currency, siacoinClaimBalance types.Currency)

		// UnconfirmedBalance returns the unconfirmed balance of the wallet.
		// Outgoing funds and incoming funds are reported separately. Refund
		// outputs are included, meaning that a sending a single coin to
		// someone could result in 'outgoing: 12, incoming: 11'. Siafunds are
		// not considered in the unconfirmed balance.
		UnconfirmedBalance() (outgoingSiacoins types.Currency, incomingSiacoins types.Currency)

		// History returns all of the history that was confirmed at heights
		// [startHeight, endHeight]. Unconfirmed history not included.
		History(startHeight types.BlockHeight, endHeight types.BlockHeight) ([]WalletTransaction, error)

		// AddressHistory returns all of the transactions that are related to a
		// given address.
		AddressHistory(types.UnlockHash) ([]WalletTransaction, error)

		// UnconfirmedHistory returns the list of known unconfirmed wallet
		// transactions.
		UnconfirmedHistory() []WalletTransaction

		// AddressUnconfirmedHistory returns all of the wallet transactions
		// related to a given address.
		AddressUnconfirmedHistory(types.UnlockHash) []WalletTransaction

		// Transaction returns the transaction with the given id. The bool
		// indicates whether the transaction is in the wallet database. The
		// wallet only stores transactions that are related to the wallet.
		Transaction(types.TransactionID) (types.Transaction, bool)

		// Transactions returns all transactions that were confirmed from
		// height [startHeight, endHeight] that are relevant to the wallet.
		Transactions(startHeight, endHeight types.BlockHeight) ([]types.Transaction, error)

		// UnconfirmedTransactions returns all unconfirmed transactions.
		UnconfirmedTransactions() []types.Transaction

		// RegisterTransaction takes a transaction and its parents and returns
		// a TransactionBuilder which can be used to expand the transaction.
		// The most typical call is 'RegisterTransaction(types.Transaction{},
		// nil)', which registers a new transaction without parents.
		RegisterTransaction(t types.Transaction, parents []types.Transaction) TransactionBuilder

		// StartTransaction is a convenience method that calls
		// RegisterTransaction(types.Transaction{}, nil)
		StartTransaction() TransactionBuilder

		// SendSiacoins is a tool for sending siacoins from the wallet to an
		// address. Sending money usually results in multiple transactions. The
		// transactions are automatically given to the transaction pool, and
		// are also returned to the caller.
		SendSiacoins(amount types.Currency, dest types.UnlockHash) ([]types.Transaction, error)

		// SendSiafunds is a tool for sending siafunds from the wallet to an
		// address. Sending money usually results in multiple transactions. The
		// transactions are automatically given to the transaction pool, and
		// are also returned to the caller.
		SendSiafunds(amount types.Currency, dest types.UnlockHash) ([]types.Transaction, error)
	}
)

// CalculateWalletTransactionID is a helper function for determining the id of
// a wallet transaction.
func CalculateWalletTransactionID(tid types.TransactionID, oid types.OutputID) WalletTransactionID {
	return WalletTransactionID(crypto.HashAll(tid, oid))
}

// SeedToString converts a wallet seed to a human friendly string.
func SeedToString(seed Seed, did mnemonics.DictionaryID) (string, error) {
	fullChecksum := crypto.HashObject(seed)
	checksumSeed := append(seed[:], fullChecksum[:SeedChecksumSize]...)
	phrase, err := mnemonics.ToPhrase(checksumSeed, did)
	if err != nil {
		return "", err
	}
	return phrase.String(), nil
}

// StringToSeed converts a string to a wallet seed.
func StringToSeed(str string, did mnemonics.DictionaryID) (Seed, error) {
	// Decode the string into the checksummed byte slice.
	checksumSeedBytes, err := mnemonics.FromString(str, did)
	if err != nil {
		return Seed{}, err
	}

	// Copy the seed from the checksummed slice.
	var seed Seed
	copy(seed[:], checksumSeedBytes)
	fullChecksum := crypto.HashObject(seed)
	if !bytes.Equal(fullChecksum[:SeedChecksumSize], checksumSeedBytes[crypto.EntropySize:]) {
		return Seed{}, errors.New("seed failed checksum verification")
	}
	return seed, nil
}
