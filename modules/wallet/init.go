package wallet

import (
	"crypto/rand"
	"errors"

	"github.com/threefoldtech/rivine/modules"
)

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
func (w *Wallet) Init(primarySeed modules.Seed) (modules.Seed, error) {
	if err := w.tg.Add(); err != nil {
		return modules.Seed{}, err
	}
	defer w.tg.Done()
	w.mu.Lock()
	subscribed := w.subscribed
	seed, err := w.initPlainWallet(primarySeed)
	w.mu.Unlock()
	if err != nil {
		return modules.Seed{}, err
	}
	if subscribed {
		return seed, nil
	}
	// subscribe wallet immediate if not yet subscribed
	err = w.subscribeWallet()
	if err != nil {
		return seed, err
	}
	w.mu.Lock()
	w.subscribed = true
	w.mu.Unlock()
	return seed, nil
}

// initPlainWallet creates the wallet,
// The primary seed can be given if it is known upfront,
// but if a nil primary seed is given, a random one will be generated instead.
func (w *Wallet) initPlainWallet(seed modules.Seed) (modules.Seed, error) {
	// Check if the wallet encryption key has already been set.
	if len(w.persist.EncryptionVerification) != 0 {
		return modules.Seed{}, errReencrypt
	}
	// Check if the wallet has already been created as a plain wallet.
	if w.persist.PrimarySeedFile.UID != (UniqueID{}) {
		return modules.Seed{}, errors.New("wallet has already been created as a plain wallet")
	}

	// If no primary seed is given, create a random seed insead.
	// Existing seeds get the full initial seed depth (PublicKeysPerSeed) (resulting in more addresses up front),
	// compared to a new seed. This because an existing seed probably might have already addresses,
	// outside the limited depth as identified by WalletSeedPreloadDepth.
	preloadDepth := uint64(modules.PublicKeysPerSeed)
	if seed == (modules.Seed{}) {
		_, err := rand.Read(seed[:])
		if err != nil {
			return modules.Seed{}, err
		}
		preloadDepth = modules.WalletSeedPreloadDepth
	}

	// If the input key is blank, use the seed to create the master key.
	// Otherwise, use the input key.
	err := w.createPlainSeed(seed, preloadDepth)
	if err != nil {
		return modules.Seed{}, err
	}

	// ensure the verification is nil
	w.persist.EncryptionVerification = nil

	// save the wallet
	err = w.saveSettings()
	if err != nil {
		return modules.Seed{}, err
	}

	// Load the wallet seed that is used to generate new addresses.
	err = w.initPlainPrimarySeed()
	if err != nil {
		return modules.Seed{}, err
	}

	// Load all wallet seeds that are not used to generate new addresses.
	err = w.initPlainAuxiliarySeeds()
	if err != nil {
		return modules.Seed{}, err
	}

	// unlock the wallet by default
	w.unlocked = true

	// return the primary seed
	return seed, nil
}
