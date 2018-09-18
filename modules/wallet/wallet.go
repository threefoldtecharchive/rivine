package wallet

// TODO: Theoretically, the transaction builder in this wallet supports
// multisig, but there are no automated tests to verify that.

import (
	"errors"
	"fmt"
	"sort"
	"sync"

	"github.com/NebulousLabs/threadgroup"
	bolt "github.com/rivine/bbolt"
	"github.com/rivine/rivine/build"
	"github.com/rivine/rivine/crypto"
	"github.com/rivine/rivine/encoding"
	"github.com/rivine/rivine/modules"
	"github.com/rivine/rivine/persist"
	siasync "github.com/rivine/rivine/sync"
	"github.com/rivine/rivine/types"
)

const (
	// RespendTimeout records the number of blocks that the wallet will wait
	// before spending an output that has been spent in the past. If the
	// transaction spending the output has not made it to the transaction pool
	// after the limit, the assumption is that it never will.
	RespendTimeout = 40
)

var (
	errNilConsensusSet = errors.New("wallet cannot initialize with a nil consensus set")
	errNilTpool        = errors.New("wallet cannot initialize with a nil transaction pool")
	errUnknownAddress  = errors.New("given wallet address is not known")
)

// spendableKey is a set of secret keys plus the corresponding unlock
// conditions.  The public key can be derived from the secret key and then
// matched to the corresponding public keys in the unlock conditions. All
// addresses that are to be used in 'FundSiacoins' or 'FundSiafunds' in the
// transaction builder must conform to this form of spendable key.
type spendableKey struct {
	PublicKey crypto.PublicKey
	SecretKey crypto.SecretKey
}

func (sk spendableKey) WipeSecret() spendableKey {
	crypto.SecureWipe(sk.SecretKey[:])
	return sk
}

func (sk spendableKey) UnlockHash() types.UnlockHash {
	return types.NewEd25519PubKeyUnlockHash(sk.PublicKey)
}

// Wallet is an object that tracks balances, creates keys and addresses,
// manages building and sending transactions.
type Wallet struct {
	// encrypted indicates whether the wallet has been encrypted (i.e.
	// initialized). unlocked indicates whether the wallet is currently
	// storing secret keys in memory. subscribed indicates whether the wallet
	// has subscribed to the consensus set yet - the wallet is unable to
	// subscribe to the consensus set until it has been unlocked for the first
	// time. The primary seed is used to generate new addresses for the
	// wallet.
	encrypted   bool
	unlocked    bool
	subscribed  bool
	primarySeed modules.Seed

	cs    modules.ConsensusSet
	tpool modules.TransactionPool

	// The following set of fields are responsible for tracking the confirmed
	// outputs, and for being able to spend them. The seeds are used to derive
	// the keys that are tracked on the blockchain. All keys are pregenerated
	// from the seeds, when checking new outputs or spending outputs, the seeds
	// are not referenced at all. The seeds are only stored so that the user
	// may access them.
	seeds     []modules.Seed
	keys      map[types.UnlockHash]spendableKey
	lookahead map[types.UnlockHash]uint64

	// A separate TryMutex is used to protect against concurrent unlocking or
	// initialization.
	scanLock siasync.TryMutex

	// unconfirmedProcessedTransactions tracks unconfirmed transactions.
	//
	// TODO: Replace this field with a linked list. Currently when a new
	// transaction set diff is provided, the entire array needs to be
	// reallocated. Since this can happen tens of times per second, and the
	// array can have tens of thousands of elements, it's a performance issue.
	unconfirmedProcessedTransactions []modules.ProcessedTransaction

	// The wallet's database tracks its seeds, keys, outputs, and
	// transactions. A global db transaction is maintained in memory to avoid
	// excessive disk writes. Any operations involving dbTx must hold an
	// exclusive lock.
	//
	// If dbRollback is set, then when the database syncs it will perform a
	// rollback instead of a commit. For safety reasons, the db will close and
	// the wallet will close if a rollback is performed.
	db         *persist.BoltDatabase
	dbRollback bool
	dbTx       *bolt.Tx

	persistDir string
	log        *persist.Logger
	mu         sync.RWMutex
	// The wallet's ThreadGroup tells tracked functions to shut down and
	// blocks until they have all exited before returning from Close.
	tg threadgroup.ThreadGroup

	bcInfo   types.BlockchainInfo
	chainCts types.ChainConstants
}

// Height return the internal processed consensus height of the wallet
func (w *Wallet) Height() (types.BlockHeight, error) {
	if err := w.tg.Add(); err != nil {
		return 0, modules.ErrWalletShutdown
	}
	defer w.tg.Done()

	w.mu.Lock()
	defer w.mu.Unlock()

	var height uint64
	err := w.db.View(func(tx *bolt.Tx) error {
		return encoding.Unmarshal(tx.Bucket(bucketWallet).Get(keyConsensusHeight), &height)
	})
	if err != nil {
		return 0, err
	}
	return types.BlockHeight(height), nil
}

// New creates a new wallet, loading any known addresses from the input file
// name and then using the file to save in the future. Keys and addresses are
// not loaded into the wallet during the call to 'new', but rather during the
// call to 'Unlock'.
func New(cs modules.ConsensusSet, tpool modules.TransactionPool, persistDir string, bcInfo types.BlockchainInfo, chainCts types.ChainConstants) (*Wallet, error) {
	// Check for nil dependencies.
	if cs == nil {
		return nil, errNilConsensusSet
	}
	if tpool == nil {
		return nil, errNilTpool
	}

	// Initialize the data structure.
	w := &Wallet{
		cs:    cs,
		tpool: tpool,

		keys:      make(map[types.UnlockHash]spendableKey),
		lookahead: make(map[types.UnlockHash]uint64),

		persistDir: persistDir,

		bcInfo:   bcInfo,
		chainCts: chainCts,
	}
	err := w.initPersist()
	if err != nil {
		return nil, err
	}
	return w, nil
}

// Close terminates all ongoing processes involving the wallet, enabling
// garbage collection.
func (w *Wallet) Close() error {
	if err := w.tg.Stop(); err != nil {
		return err
	}
	var errs []error
	// Lock the wallet outside of mu.Lock because Lock uses its own mu.Lock.
	// Once the wallet is locked it cannot be unlocked except using the
	// unexported unlock method (w.Unlock returns an error if the wallet's
	// ThreadGroup is stopped).
	unlocked := w.managedUnlocked()
	if unlocked {
		if err := w.managedLock(); err != nil {
			errs = append(errs, err)
		}
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	w.cs.Unsubscribe(w)
	w.tpool.Unsubscribe(w)

	if err := w.log.Close(); err != nil {
		errs = append(errs, fmt.Errorf("log.Close failed: %v", err))
	}
	return build.JoinErrors(errs, "; ")
}

// AllAddresses returns all addresses that the wallet is able to spend from.
// Addresses are returned sorted in byte-order.
func (w *Wallet) AllAddresses() ([]types.UnlockHash, error) {
	if err := w.tg.Add(); err != nil {
		return []types.UnlockHash{}, modules.ErrWalletShutdown
	}
	defer w.tg.Done()

	w.mu.RLock()
	defer w.mu.RUnlock()

	if !w.unlocked {
		return nil, modules.ErrLockedWallet
	}

	addrs := make(types.UnlockHashSlice, 0, len(w.keys))
	for addr := range w.keys {
		addrs = append(addrs, addr)
	}
	sort.Slice(addrs, func(i, j int) bool {
		return addrs[i].Cmp(addrs[j]) < 0
	})
	return addrs, nil
}

// GetKey gets the pub/priv key pair,
// which is linked to the given unlock hash (address).
func (w *Wallet) GetKey(address types.UnlockHash) (pk types.SiaPublicKey, sk types.ByteSlice, err error) {
	w.mu.RLock()
	pk, sk, err = w.getKey(address)
	w.mu.RUnlock()
	return
}
func (w *Wallet) getKey(address types.UnlockHash) (types.SiaPublicKey, types.ByteSlice, error) {
	if !w.unlocked {
		return types.SiaPublicKey{}, types.ByteSlice{}, modules.ErrLockedWallet
	}
	sp, found := w.keys[address]
	if !found {
		return types.SiaPublicKey{}, types.ByteSlice{}, errUnknownAddress
	}
	return types.Ed25519PublicKey(sp.PublicKey), types.ByteSlice(sp.SecretKey[:]), nil
}
func (w *Wallet) keyExists(address types.UnlockHash) (bool, error) {
	if !w.unlocked {
		return false, modules.ErrLockedWallet
	}
	_, exists := w.keys[address]
	return exists, nil
}

// Rescanning reports whether the wallet is currently rescanning the
// blockchain.
func (w *Wallet) Rescanning() (bool, error) {
	if err := w.tg.Add(); err != nil {
		return false, modules.ErrWalletShutdown
	}
	defer w.tg.Done()

	rescanning := !w.scanLock.TryLock()
	if !rescanning {
		w.scanLock.Unlock()
	}
	return rescanning, nil
}

// GetUnspentBlockStakeOutputs returns the blockstake outputs where the beneficiary is an
// address this wallet has an unlockhash for.
func (w *Wallet) GetUnspentBlockStakeOutputs() ([]types.UnspentBlockStakeOutput, error) {
	if err := w.tg.Add(); err != nil {
		return nil, modules.ErrWalletShutdown
	}
	defer w.tg.Done()

	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.unlocked {
		return nil, modules.ErrLockedWallet
	}

	// ensure durability of reported balance
	if err := w.syncDB(); err != nil {
		return nil, err
	}

	unspent := make([]types.UnspentBlockStakeOutput, 0)

	// prepare fulfillable context
	ctx := w.getFulfillableContextForLatestBlock()

	// collect all fulfillable block stake outputs
	dbForEachUnspentBlockStakeOutput(w.dbTx, func(_ types.BlockStakeOutputID, ubso types.UnspentBlockStakeOutput) {
		if ubso.Condition.Fulfillable(ctx) {
			unspent = append(unspent, ubso)
		}
	})
	return unspent, nil
}

func (w *Wallet) getFulfillableContextForLatestBlock() types.FulfillableContext {
	height := w.cs.Height()
	block, _ := w.cs.BlockAtHeight(height)
	return types.FulfillableContext{
		BlockHeight: height,
		BlockTime:   block.Timestamp,
	}
}
