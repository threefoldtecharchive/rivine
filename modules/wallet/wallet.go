package wallet

// TODO: Theoretically, the transaction builder in this wallet supports
// multisig, but there are no automated tests to verify that.

import (
	"errors"
	"fmt"
	"sort"
	"sync"

	"github.com/threefoldtech/rivine/build"
	"github.com/threefoldtech/rivine/crypto"
	"github.com/threefoldtech/rivine/modules"
	"github.com/threefoldtech/rivine/persist"
	siasync "github.com/threefoldtech/rivine/sync"
	"github.com/threefoldtech/rivine/types"
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

func (sk spendableKey) UnlockHash() (types.UnlockHash, error) {
	return types.NewEd25519PubKeyUnlockHash(sk.PublicKey)
}

// Wallet is an object that tracks balances, creates keys and addresses,
// manages building and sending transactions.
type Wallet struct {
	// unlocked indicates whether the wallet is currently storing secret keys
	// in memory. subscribed indicates whether the wallet has subscribed to the
	// consensus set yet - the wallet is unable to subscribe to the consensus
	// set until it has been unlocked for the first time. The primary seed is
	// used to generate new addresses for the wallet.
	unlocked    bool
	subscribed  bool
	persist     WalletPersist
	primarySeed modules.Seed

	// The wallet's dependencies. The items 'consensusSetHeight' and
	// 'siafundPool' are tracked separately from the consensus set to minimize
	// the number of queries that the wallet needs to make to the consensus
	// set; queries to the consensus set are very slow.
	cs                 modules.ConsensusSet
	tpool              modules.TransactionPool
	consensusSetHeight types.BlockHeight

	// The following set of fields are responsible for tracking the confirmed
	// outputs, and for being able to spend them. The seeds are used to derive
	// the keys that are tracked on the blockchain. All keys are pregenerated
	// from the seeds, when checking new outputs or spending outputs, the seeds
	// are not referenced at all. The seeds are only stored so that the user
	// may access them.
	//
	// coinOutputs, blockstakeOutputs, and spentOutputs are kept so that they
	// can be scanned when trying to fund transactions.
	seeds                    []modules.Seed
	keys                     map[types.UnlockHash]spendableKey
	coinOutputs              map[types.CoinOutputID]types.CoinOutput
	blockstakeOutputs        map[types.BlockStakeOutputID]types.BlockStakeOutput
	unspentblockstakeoutputs map[types.BlockStakeOutputID]types.UnspentBlockStakeOutput
	spentOutputs             map[types.OutputID]types.BlockHeight

	// multiSigOutputs holds all the multisig addresses this wallet is part of
	multiSigCoinOutputs       map[types.CoinOutputID]types.CoinOutput
	multiSigBlockStakeOutputs map[types.BlockStakeOutputID]types.BlockStakeOutput

	// The following fields are kept to track transaction history.
	// processedTransactions are stored in chronological order, and have a map for
	// constant time random access. The set of full transactions is kept as
	// well, ordering can be determined by the processedTransactions slice.
	//
	// The unconfirmed transactions are kept the same way, except without the
	// random access. It is assumed that the list of unconfirmed transactions
	// will be small enough that this will not be a problem.
	//
	// historicOutputs is kept so that the values and addresses of transaction inputs can be
	// determined. historicOutputs is never cleared, but in general should be
	// small compared to the list of transactions.
	processedTransactions            []modules.ProcessedTransaction
	processedTransactionMap          map[types.TransactionID]*modules.ProcessedTransaction
	unconfirmedProcessedTransactions []modules.ProcessedTransaction

	// TODO: Storing the whole set of historic outputs is expensive and
	// unnecessary. There's a better way to do it.
	historicOutputs map[types.OutputID]historicOutput

	persistDir string
	log        *persist.Logger
	mu         sync.RWMutex
	// The wallet's ThreadGroup tells tracked functions to shut down and
	// blocks until they have all exited before returning from Close.
	tg siasync.ThreadGroup

	bcInfo   types.BlockchainInfo
	chainCts types.ChainConstants
}

type historicOutput struct {
	UnlockHash types.UnlockHash
	Value      types.Currency
}

// New creates a new wallet, loading any known addresses from the input file
// name and then using the file to save in the future. Keys and addresses are
// not loaded into the wallet during the call to 'new', but rather during the
// call to 'Unlock'.
func New(cs modules.ConsensusSet, tpool modules.TransactionPool, persistDir string, bcInfo types.BlockchainInfo, chainCts types.ChainConstants, verboseLogging bool) (*Wallet, error) {
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

		keys:                      make(map[types.UnlockHash]spendableKey),
		coinOutputs:               make(map[types.CoinOutputID]types.CoinOutput),
		blockstakeOutputs:         make(map[types.BlockStakeOutputID]types.BlockStakeOutput),
		spentOutputs:              make(map[types.OutputID]types.BlockHeight),
		unspentblockstakeoutputs:  make(map[types.BlockStakeOutputID]types.UnspentBlockStakeOutput),
		multiSigCoinOutputs:       make(map[types.CoinOutputID]types.CoinOutput),
		multiSigBlockStakeOutputs: make(map[types.BlockStakeOutputID]types.BlockStakeOutput),

		processedTransactionMap: make(map[types.TransactionID]*modules.ProcessedTransaction),

		historicOutputs: make(map[types.OutputID]historicOutput),

		persistDir: persistDir,

		bcInfo:   bcInfo,
		chainCts: chainCts,
	}
	err := w.initPersist(verboseLogging)
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
	if w.Unlocked() {
		w.mu.RLock()
		encrypted := w.persist.EncryptionVerification != nil
		w.mu.RUnlock()
		if encrypted {
			if err := w.Lock(); err != nil {
				errs = append(errs, err)
			}
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
	w.mu.RLock()
	defer w.mu.RUnlock()

	if !w.unlocked {
		return nil, modules.ErrLockedWallet
	}

	addrs := make(types.UnlockHashSlice, 0, len(w.keys))
	for addr := range w.keys {
		addrs = append(addrs, addr)
	}
	sort.Sort(addrs)
	return addrs, nil
}

// GetKey gets the pub/priv key pair,
// which is linked to the given unlock hash (address).
func (w *Wallet) GetKey(address types.UnlockHash) (pk types.PublicKey, sk types.ByteSlice, err error) {
	w.mu.RLock()
	pk, sk, err = w.getKey(address)
	w.mu.RUnlock()
	return
}
func (w *Wallet) getKey(address types.UnlockHash) (types.PublicKey, types.ByteSlice, error) {
	if !w.unlocked {
		return types.PublicKey{}, types.ByteSlice{}, modules.ErrLockedWallet
	}
	sp, found := w.keys[address]
	if !found {
		return types.PublicKey{}, types.ByteSlice{}, errUnknownAddress
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

// GetUnspentBlockStakeOutputs returns the blockstake outputs where the beneficiary is an
// address this wallet has an unlockhash for.
func (w *Wallet) GetUnspentBlockStakeOutputs() (unspent []types.UnspentBlockStakeOutput, err error) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	if !w.unlocked {
		err = modules.ErrLockedWallet
		return
	}

	unspent = make([]types.UnspentBlockStakeOutput, 0)

	// prepare fulfillable context
	ctx := w.getFulfillableContextForLatestBlock()

	// collect all fulfillable block stake outputs
	for usbsoid, output := range w.blockstakeOutputs {
		if output.Condition.Fulfillable(ctx) {
			unspent = append(unspent, w.unspentblockstakeoutputs[usbsoid])
		}
	}
	return
}

func (w *Wallet) getFulfillableContextForLatestBlock() types.FulfillableContext {
	height := w.cs.Height()
	block, _ := w.cs.BlockAtHeight(height)
	return types.FulfillableContext{
		BlockHeight: height,
		BlockTime:   block.Timestamp,
	}
}
