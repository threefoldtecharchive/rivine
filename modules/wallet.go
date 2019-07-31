package modules

import (
	"encoding/hex"
	"errors"

	"github.com/threefoldtech/rivine/build"
	"github.com/threefoldtech/rivine/crypto"
	"github.com/threefoldtech/rivine/types"
	bip39 "github.com/tyler-smith/go-bip39"
)

const (
	// WalletDir is the directory that contains the wallet persistence.
	WalletDir = "wallet"

	// SeedChecksumSize is the number of bytes that are used to checksum
	// addresses to prevent accidental spending.
	SeedChecksumSize = 6

	// PublicKeysPerSeed define the number of public keys that get pregenerated
	// for a seed at startup when searching for balances in the blockchain.
	PublicKeysPerSeed = 2500

	// WalletSeedPreloadDepth is the number of addresses that get automatically
	// loaded by the wallet at startup.
	WalletSeedPreloadDepth = 25
)

var (
	// ErrBadEncryptionKey is returned if the incorrect encryption key to a
	// file is provided.
	ErrBadEncryptionKey = errors.New("provided encryption key is incorrect")

	// ErrLowBalance is returned if the wallet does not have enough funds to
	// complete the desired action.
	ErrLowBalance = errors.New("insufficient balance")

	// ErrIncompleteTransactions is returned if the wallet has incomplete
	// transactions being built that are using all of the current outputs, and
	// therefore the wallet is unable to spend money despite it not technically
	// being 'unconfirmed' yet.
	ErrIncompleteTransactions = errors.New("wallet has coins spent in incomplete transactions - not enough remaining coins")

	// ErrLockedWallet is returned when an action cannot be performed due to
	// the wallet being locked.
	ErrLockedWallet = errors.New("wallet must be unlocked before it can be used")

	// ErrEncryptedWallet is returned in case the wallet is encrypted, preventing it from being
	// used for plain purposes.
	ErrEncryptedWallet = errors.New("wallet is encrypted and cannot use plain functionality")
)

type (
	// Seed is cryptographic entropy that is used to derive spendable wallet
	// addresses.
	Seed [crypto.EntropySize]byte

	// WalletTransactionID is a unique identifier for a wallet transaction.
	WalletTransactionID crypto.Hash

	// A ProcessedInput represents funding to a transaction. The input is
	// coming from an address and going to the outputs. The fund types are
	// 'SiacoinInput', 'SiafundInput'.
	ProcessedInput struct {
		FundType types.Specifier `json:"fundtype"`
		// WalletAddress indicates it's an address owned by this wallet
		WalletAddress  bool             `json:"walletaddress"`
		RelatedAddress types.UnlockHash `json:"relatedaddress"`
		Value          types.Currency   `json:"value"`
	}

	// A ProcessedOutput is a coin output that appears in a transaction.
	// Some outputs mature immediately, some are delayed.
	//
	// Fund type can either be 'CoinOutput', 'BlockStakeOutput'
	// or 'MinerFee'. All outputs except the miner fee create
	// outputs accessible to an address. Miner fees are not spendable, and
	// instead contribute to the block subsidy.
	//
	// MaturityHeight indicates at what block height the output becomes
	// available. CoinInputs and BlockStakeInputs become available immediately.
	// MinerPayouts become available after 144 confirmations.
	ProcessedOutput struct {
		FundType       types.Specifier   `json:"fundtype"`
		MaturityHeight types.BlockHeight `json:"maturityheight"`
		// WalletAddress indicates it's an address owned by this wallet
		WalletAddress  bool             `json:"walletaddress"`
		RelatedAddress types.UnlockHash `json:"relatedaddress"`
		Value          types.Currency   `json:"value"`
	}

	// A ProcessedTransaction is a transaction that has been processed into
	// explicit inputs and outputs and tagged with some header data such as
	// confirmation height + timestamp.
	//
	// Because of the block subsidy, a block is considered as a transaction.
	// Since there is technically no transaction id for the block subsidy, the
	// block id is used instead.
	ProcessedTransaction struct {
		Transaction           types.Transaction   `json:"transaction"`
		TransactionID         types.TransactionID `json:"transactionid"`
		ConfirmationHeight    types.BlockHeight   `json:"confirmationheight"`
		ConfirmationTimestamp types.Timestamp     `json:"confirmationtimestamp"`

		Inputs  []ProcessedInput  `json:"inputs"`
		Outputs []ProcessedOutput `json:"outputs"`
	}

	// MultiSigWallet is a collection of coin and blockstake outputs, which have the same
	// unlockhash.
	MultiSigWallet struct {
		Address             types.UnlockHash           `json:"address"`
		CoinOutputIDs       []types.CoinOutputID       `json:"coinoutputids"`
		BlockStakeOutputIDs []types.BlockStakeOutputID `json:"blockstakeoutputids"`

		ConfirmedCoinBalance       types.Currency `json:"confirmedcoinbalance"`
		ConfirmedLockedCoinBalance types.Currency `json:"confirmedlockedcoinbalance"`
		UnconfirmedOutgoingCoins   types.Currency `json:"unconfirmedoutgoingcoins"`
		UnconfirmedIncomingCoins   types.Currency `json:"unconfirmedincomingcoins"`

		ConfirmedBlockStakeBalance       types.Currency `json:"confirmedblockstakebalance"`
		ConfirmedLockedBlockStakeBalance types.Currency `json:"confirmedlockedblockstakebalance"`
		UnconfirmedOutgoingBlockStakes   types.Currency `json:"unconfirmedoutgoingblockstakes"`
		UnconfirmedIncomingBlockStakes   types.Currency `json:"unconfirmedincomingblockstakes"`

		Owners  []types.UnlockHash `json:"owners"`
		MinSigs uint64             `json:"minsigs"`
	}

	// TransactionBuilder is used to construct custom transactions. A transaction
	// builder is initialized via 'RegisterTransaction' and then can be modified by
	// adding funds or other fields. The transaction is completed by calling
	// 'Sign', which will sign all inputs added via the 'FundSiacoins' or
	// 'FundSiafunds' call. All modifications are additive.
	//
	// Parents of the transaction are kept in the transaction builder. A parent is
	// any unconfirmed transaction that is required for the child to be valid.
	//
	// Transaction builders are not thread safe.
	TransactionBuilder interface {
		// Fundcoins will add a siacoin input of exactly 'amount' to the
		// transaction. A parent transaction may be needed to achieve an input
		// with the correct value. The siacoin input will not be signed until
		// 'Sign' is called on the transaction builder. The expectation is that
		// the transaction will be completed and broadcast within a few hours.
		// Longer risks double-spends, as the wallet will assume that the
		// transaction failed.
		FundCoins(amount types.Currency, refundAddress *types.UnlockHash, reuseRefundAddress bool) error

		// FundBlockStakes will add a siafund input of exactly 'amount' to the
		// transaction. A parent transaction may be needed to achieve an input
		// with the correct value. The siafund input will not be signed until
		// 'Sign' is called on the transaction builder. Any siacoins that are
		// released by spending the siafund outputs will be sent to another
		// address owned by the wallet. The expectation is that the transaction
		// will be completed and broadcast within a few hours. Longer risks
		// double-spends, because the wallet will assume the transaction
		// failed.
		FundBlockStakes(amount types.Currency, refundAddress *types.UnlockHash, reuseRefundAddress bool) error

		// SpendBlockStake will link the unspent block stake to the transaction as an input.
		// In contrast with FundBlockStakes, this function will not loop over all unspent
		// block stake output. the ubsoid is an argument. The blockstake input will not be
		// signed until 'Sign' is called on the transaction builder.
		SpendBlockStake(ubsoid types.BlockStakeOutputID) error

		// AddParents adds a set of parents to the transaction.
		AddParents([]types.Transaction)

		// AddMinerFee adds a miner fee to the transaction, returning the index
		// of the miner fee within the transaction.
		AddMinerFee(fee types.Currency) uint64

		// AddCoinInput adds a coin input to the transaction, returning
		// the index of the coin input within the transaction. When 'Sign'
		// gets called, this input will be left unsigned.
		AddCoinInput(types.CoinInput) uint64

		// AddCoinOutput adds a coin output to the transaction, returning
		// the index of the coin output within the transaction.
		AddCoinOutput(types.CoinOutput) uint64

		// AddBlockStakeInput adds a blockstake input to the transaction, returning
		// the index of the blockstake input within the transaction. When 'Sign'
		// is called, this input will be left unsigned.
		AddBlockStakeInput(types.BlockStakeInput) uint64

		// AddBlockStakeOutput adds a blockstake output to the transaction, returning
		// the index of the blockstake output within the transaction.
		AddBlockStakeOutput(types.BlockStakeOutput) uint64

		// AddArbitraryData sets the arbitrary data of the transaction.
		SetArbitraryData(arb []byte)

		// Sign will sign any inputs added by 'FundCoins' or 'FundBlockStakes'
		// and return a transaction set that contains all parents prepended to
		// the transaction. If more fields need to be added, a new transaction
		// builder will need to be created.
		//
		// An error will be returned if there are multiple calls to 'Sign',
		// sometimes even if the first call to Sign has failed. Sign should
		// only ever be called once, and if the first signing fails, the
		// transaction should be dropped.
		Sign() ([]types.Transaction, error)

		// View returns the incomplete transaction along with all of its
		// parents.
		View() (txn types.Transaction, parents []types.Transaction)

		// ViewAdded returns all of the siacoin inputs, siafund inputs, and
		// parent transactions that have been automatically added by the
		// builder. Items are returned by index.
		ViewAdded() (newParents, siacoinInputs, siafundInputs []int)

		// Drop indicates that a transaction is no longer useful and will not be
		// broadcast, and that all of the outputs can be reclaimed. 'Drop'
		// should only be used before signatures are added.
		Drop()

		// SignAllPossible tries to sign as much of the inputs —and extension if required—
		// in the tranaction using the keys loaded in the wallet
		SignAllPossible() error
	}

	// EncryptionManager can encrypt, lock, unlock, and indicate the current
	// status of the EncryptionManager.
	EncryptionManager interface {
		// Encrypt will encrypt the wallet using the input key. Upon
		// encryption, a primary seed will be created for the wallet (no seed
		// exists prior to this point). If the key is blank, then the hash of
		// the seed that is generated will be used as the key.
		//
		// Encrypt can only be called once throughout the life of the wallet
		// and will return an error on subsequent calls (even after restarting
		// the wallet). To reset the wallet, the wallet files must be moved to
		// a different directory or deleted.
		Encrypt(masterKey crypto.TwofishKey, primarySeed Seed) (Seed, error)

		// Encrypted returns whether or not the wallet has been encrypted yet.
		// After being encrypted for the first time, the wallet can only be
		// unlocked using the encryption password.
		Encrypted() bool

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

		// Unlocked returns true if the wallet is currently unlocked, false
		// otherwise.
		Unlocked() bool
	}

	// KeyManager manages wallet keys, including the use of seeds, creating and
	// loading backups, and providing a layer of compatibility for older wallet
	// files.
	KeyManager interface {
		// AllAddresses returns all addresses that the wallet is able to spend
		// from, including unseeded addresses. Addresses are returned sorted in
		// byte-order.
		AllAddresses() ([]types.UnlockHash, error)

		// AllSeeds returns all of the seeds that are being tracked by the
		// wallet, including the primary seed. Only the primary seed is used to
		// generate new addresses, but the wallet can spend funds sent to
		// public keys generated by any of the seeds returned.
		AllSeeds() ([]Seed, error)

		// GetKey allows you to fetch the Public/Private key pair,
		// which is linked to the given unlock hash (assumed to be the address a user).
		GetKey(address types.UnlockHash) (types.PublicKey, types.ByteSlice, error)

		// PrimarySeed returns the current primary seed of the wallet,
		// unencrypted, with an int indicating how many addresses have been
		// consumed.
		PrimarySeed() (Seed, uint64, error)

		// NextAddress returns a new coin addresses generated from the
		// primary seed.
		NextAddress() (types.UnlockHash, error)

		// CreateBackup will create a backup of the wallet at the provided
		// filepath. The backup will have all seeds and keys.
		CreateBackup(string) error

		// LoadBackup will load a backup of the wallet from the provided
		// address. The backup wallet will be added as an auxiliary seed, not
		// as a primary seed.
		// LoadBackup(masterKey, backupMasterKey crypto.TwofishKey, string) error

		// LoadSeed will recreate a wallet file using the recovery phrase.
		// LoadSeed only needs to be called if the original seed file or
		// encryption password was lost. The master key is used to encrypt the
		// recovery seed before saving it to disk.
		LoadSeed(crypto.TwofishKey, Seed) error

		// LoadPlainSeed will recreate a wallet file using the recovery phrase.
		// LoadPlainSeed only needs to be called if the original seed file was lost.
		LoadPlainSeed(Seed) error
	}

	// Wallet stores and manages siacoins and siafunds. The wallet file is
	// encrypted using a user-specified password. Common addresses are all
	// derived from a single address seed.
	Wallet interface {
		EncryptionManager
		KeyManager

		// Init will create the PLAIN wallet using the primary seed,
		// or creating a seed for you if none is given.
		//
		// Init can only be called once throughout the life of the wallet
		// and will return an error on subsequent calls (even after restarting
		// the wallet). To reset the wallet, the wallet files must be moved to
		// a different directory or deleted.
		//
		// NOTE: the seed is stored in a plain text file on the local FS.
		// This is not recommended, unless you are working in an isolated and controlled environment.
		// Please use the Encrypt method instead, in order to create the wallet using a master key,
		// the generated adddresses will be the same either way, but the seed will be encrypted
		// prior to being stored on the local FS.
		Init(primarySeed Seed) (Seed, error)

		// Close permits clean shutdown during testing and serving.
		Close() error

		// ConfirmedBalance returns the confirmed balance of the wallet, minus
		// any outgoing transactions. ConfirmedBalance will include unconfirmed
		// refund transactions.
		ConfirmedBalance() (siacoinBalance types.Currency, blockstakeBalance types.Currency, err error)

		// ConfirmedLockedBalance returns the confirmed balance of the wallet, which is locked,
		// minus any outgoing transactions. ConfirmedLockedBalance will include unconfirmed
		// refund transactions which are locked as well.
		ConfirmedLockedBalance() (siacoinBalance types.Currency, blockstakeBalance types.Currency, err error)

		// GetUnspentBlockStakeOutputs returns the blockstake outputs where the beneficiary is an
		// address this wallet has an unlockhash for.
		GetUnspentBlockStakeOutputs() ([]types.UnspentBlockStakeOutput, error)

		// UnconfirmedBalance returns the unconfirmed balance of the wallet.
		// Outgoing funds and incoming funds are reported separately. Refund
		// outputs are included, meaning that sending a single coin to
		// someone could result in 'outgoing: 12, incoming: 11'. Siafunds are
		// not considered in the unconfirmed balance.
		UnconfirmedBalance() (outgoingSiacoins types.Currency, incomingSiacoins types.Currency, err error)

		// AddressTransactions returns all of the transactions that are related
		// to a given address.
		AddressTransactions(types.UnlockHash) ([]ProcessedTransaction, error)

		// AddressUnconfirmedHistory returns all of the unconfirmed
		// transactions related to a given address.
		AddressUnconfirmedTransactions(types.UnlockHash) ([]ProcessedTransaction, error)

		// Transaction returns the transaction with the given id. The bool
		// indicates whether the transaction is in the wallet database. The
		// wallet only stores transactions that are related to the wallet.
		Transaction(types.TransactionID) (ProcessedTransaction, bool, error)

		// Transactions returns all of the transactions that were confirmed at
		// heights [startHeight, endHeight]. Unconfirmed transactions are not
		// included.
		Transactions(startHeight types.BlockHeight, endHeight types.BlockHeight) ([]ProcessedTransaction, error)

		// UnconfirmedTransactions returns all unconfirmed transactions
		// relative to the wallet.
		UnconfirmedTransactions() ([]ProcessedTransaction, error)

		// MultiSigWallets returns all multisig wallets which contain at least one unlock hash owned by this wallet.
		// A multisig wallet is in this context defined as a (group of) coin and or blockstake outputs, where the unlockhash
		// of these outputs are exactly the same. In practice, this means that the collection of unlock hashes in the condition,
		// as well as the minimum amount of signatures required, must match
		MultiSigWallets() ([]MultiSigWallet, error)

		// RegisterTransaction takes a transaction and its parents and returns
		// a TransactionBuilder which can be used to expand the transaction.
		RegisterTransaction(t types.Transaction, parents []types.Transaction) TransactionBuilder

		// StartTransaction is a convenience method that calls
		// RegisterTransaction(types.Transaction{}, nil)
		StartTransaction() TransactionBuilder

		// SendCoins is a tool for sending coins from the wallet to anyone who can fulfill the
		// given condition (can be nil). The transaction is automatically given to the transaction pool, and
		// are also returned to the caller.
		SendCoins(amount types.Currency, cond types.UnlockConditionProxy, data []byte) (types.Transaction, error)

		// SendBlockStakes is a tool for sending blockstakes from the wallet to anyone who can fulfill the
		// given condition (can be nil). Sending money usually results in multiple transactions. The
		// transactions are automatically given to the transaction pool, and
		// are also returned to the caller.
		SendBlockStakes(amount types.Currency, cond types.UnlockConditionProxy) (types.Transaction, error)

		// SendOutputs is a tool for sending coins and/or block stakes from the wallet, to one or multiple addreses.
		// The transaction is automatically given to the transaction pool, and is also returned to the caller.
		SendOutputs(coinOutputs []types.CoinOutput, blockstakeOutputs []types.BlockStakeOutput, data []byte, refundAddress *types.UnlockHash, reuseRefundAddress bool) (types.Transaction, error)

		// BlockStakeStats returns the blockstake statistical information of
		// this wallet of the last 1000 blocks. If the blockcount is less than
		// 1000 blocks, BlockCount will be the number available.
		BlockStakeStats() (BCcountLast1000 uint64, BCfeeLast1000 types.Currency, BlockCount uint64, err error)

		// UnlockedUnspendOutputs returns all unlocked and unspend coin and blockstake outputs
		// owned by this wallet
		UnlockedUnspendOutputs() (map[types.CoinOutputID]types.CoinOutput, map[types.BlockStakeOutputID]types.BlockStakeOutput, error)

		// LockedUnspendOutputs returns all locked and unspend coin and blockstake outputs owned
		// by this wallet
		LockedUnspendOutputs() (map[types.CoinOutputID]types.CoinOutput, map[types.BlockStakeOutputID]types.BlockStakeOutput, error)

		// CreateRawTransaction creates a new transaction with the given inputs and outputs.
		// All inputs must exist in the consensus set at the time this method is called. The total
		// value of the inputs must match the sum of all respective outputs and the transaction fee.
		CreateRawTransaction([]types.CoinOutputID, []types.BlockStakeOutputID, []types.CoinOutput, []types.BlockStakeOutput, []byte) (types.Transaction, error)

		// GreedySign attempts to sign every input which can be signed by the keys loaded
		// in this wallet.
		GreedySign(types.Transaction) (types.Transaction, error)
	}
)

// CalculateWalletTransactionID is a helper function for determining the id of
// a wallet transaction.
func CalculateWalletTransactionID(tid types.TransactionID, oid types.OutputID) WalletTransactionID {
	h, err := crypto.HashAll(tid, oid)
	if err != nil {
		build.Severe("failed to crypto hash transaction id and output id as a single wallet tx id", err)
	}
	return WalletTransactionID(h)
}

// NewMnemonic converts a wallet seed to a mnemonic, a human friendly string.
func NewMnemonic(seed Seed) (string, error) {
	return bip39.NewMnemonic(seed[:])
}

// InitialSeedFromMnemonic converts the mnemonic into the initial seed,
// also called entropy, that was used to create the given mnemonic initially.
func InitialSeedFromMnemonic(mnemonic string) (out Seed, err error) {
	seed, err := bip39.EntropyFromMnemonic(mnemonic)
	if err != nil {
		return
	}
	copy(out[:], seed[:])
	return
}

// String returns this seed as a hex-encoded string.
func (s Seed) String() string {
	return hex.EncodeToString(s[:])
}

// LoadString loads a hex-encoded string into this seed.
func (s *Seed) LoadString(str string) error {
	b, err := hex.DecodeString(str)
	if err != nil {
		return err
	}
	if len(b) != crypto.EntropySize {
		return errors.New("seed has invalid size, should equal to crypto entropy size")
	}
	copy(s[:], b[:])
	return nil
}
