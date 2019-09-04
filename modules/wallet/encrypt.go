package wallet

import (
	"bytes"
	"crypto/rand"
	"errors"

	"github.com/threefoldtech/rivine/crypto"
	"github.com/threefoldtech/rivine/modules"
	"github.com/threefoldtech/rivine/types"
)

var (
	errAlreadyUnlocked   = errors.New("wallet has already been unlocked")
	errReencrypt         = errors.New("wallet is already encrypted, cannot encrypt again")
	errUnencryptedWallet = errors.New("wallet has not been encrypted")

	unlockModifier = types.Specifier{'u', 'n', 'l', 'o', 'c', 'k'}
)

// uidEncryptionKey creates an encryption key that is used to decrypt a
// specific key file.
func uidEncryptionKey(masterKey crypto.TwofishKey, uid UniqueID) (crypto.TwofishKey, error) {
	h, err := crypto.HashAll(masterKey, uid)
	if err != nil {
		return crypto.TwofishKey{}, err
	}
	return crypto.TwofishKey(h), nil
}

// checkMasterKey verifies that the master key is correct.
func (w *Wallet) checkMasterKey(masterKey crypto.TwofishKey) error {
	// ensure if crypto key is given
	if masterKey == (crypto.TwofishKey{}) {
		return modules.ErrBadEncryptionKey
	}

	uk, err := uidEncryptionKey(masterKey, w.persist.UID)
	if err != nil {
		return err
	}
	verification, err := uk.DecryptBytes(w.persist.EncryptionVerification)
	if err != nil {
		// Most of the time, the failure is an authentication failure.
		return modules.ErrBadEncryptionKey
	}
	expected := make([]byte, encryptionVerificationLen)
	if !bytes.Equal(expected, verification) {
		return modules.ErrBadEncryptionKey
	}
	return nil
}

// initEncryption checks that the provided encryption key is the valid
// encryption key for the wallet. If encryption has not yet been established
// for the wallet, an encryption key is created.
// The primary seed can be given if it is known upfront,
// but if a nil primary seed is given, a random one will be generated instead.
func (w *Wallet) initEncryption(masterKey crypto.TwofishKey, seed modules.Seed) (modules.Seed, error) {
	// ensure if crypto key is given
	if masterKey == (crypto.TwofishKey{}) {
		return modules.Seed{}, modules.ErrBadEncryptionKey
	}

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
	err := w.createEncryptedSeed(masterKey, seed, preloadDepth)
	if err != nil {
		return modules.Seed{}, err
	}

	// Establish the encryption verification using the masterKey. After this
	// point, the wallet is encrypted.
	uk, err := uidEncryptionKey(masterKey, w.persist.UID)
	if err != nil {
		return modules.Seed{}, err
	}
	encryptionBase := make([]byte, encryptionVerificationLen)
	w.persist.EncryptionVerification = uk.EncryptBytes(encryptionBase)
	err = w.saveSettings()
	if err != nil {
		return modules.Seed{}, err
	}
	return seed, nil
}

// managedUnlock loads all of the encrypted file structures into wallet memory. Even
// after loading, the structures are kept encrypted, but some data such as
// addresses are decrypted so that the wallet knows what to track.
func (w *Wallet) managedUnlock(masterKey crypto.TwofishKey) error {
	var subscribed bool
	err := func() error {
		w.mu.Lock()
		defer w.mu.Unlock()
		subscribed = w.subscribed

		// Wallet should only be unlocked once.
		if w.unlocked {
			return errAlreadyUnlocked
		}

		// Check if the wallet encryption key has already been set.
		if len(w.persist.EncryptionVerification) == 0 {
			return errUnencryptedWallet
		}

		// Initialize the encryption of the wallet.
		err := w.checkMasterKey(masterKey)
		if err != nil {
			return err
		}

		// Load the wallet seed that is used to generate new addresses.
		err = w.initEncryptedPrimarySeed(masterKey)
		if err != nil {
			return err
		}

		// Load all wallet seeds that are not used to generate new addresses.
		return w.initEncryptedAuxiliarySeeds(masterKey)
	}()
	if err != nil {
		return err
	}

	// Subscribe to the consensus set if this is the first unlock for the
	// wallet object.
	if !subscribed {
		err = w.subscribeWallet()
		if err != nil {
			return err
		}
		w.mu.Lock()
		w.subscribed = true
		w.mu.Unlock()
	}

	w.mu.Lock()
	w.unlocked = true
	w.mu.Unlock()
	return nil
}

// wipeSecrets erases all of the seeds and secret keys in the wallet.
func (w *Wallet) wipeSecrets() {
	// 'for i := range' must be used to prevent copies of secret data from
	// being made.
	for i := range w.keys {
		w.keys[i] = w.keys[i].WipeSecret()
	}
	for i := range w.seeds {
		crypto.SecureWipe(w.seeds[i][:])
	}
	crypto.SecureWipe(w.primarySeed[:])
	w.seeds = w.seeds[:0]
}

// Encrypted returns whether or not the wallet has been encrypted.
func (w *Wallet) Encrypted() bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	return len(w.persist.EncryptionVerification) != 0
}

// Encrypt will encrypt the wallet using the input key. Upon encryption, a
// primary seed will be created for the wallet (no seed exists prior to this
// point). If the key is blank, then the hash of the seed that is generated
// will be used as the key. The wallet will still be locked after encryption.
//
// Encrypt can only be called once throughout the life of the wallet, and will
// return an error on subsequent calls (even after restarting the wallet). To
// reset the wallet, the wallet files must be moved to a different directory or
// deleted.
//
// If no primary seed is given (which is possible if a nil seed is passed as primary seed),
// a random one will be generated for you.
func (w *Wallet) Encrypt(masterKey crypto.TwofishKey, primarySeed modules.Seed) (modules.Seed, error) {
	if err := w.tg.Add(); err != nil {
		return modules.Seed{}, err
	}
	defer w.tg.Done()
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.initEncryption(masterKey, primarySeed)
}

// Unlocked indicates whether the wallet is locked or unlocked.
func (w *Wallet) Unlocked() bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.unlocked
}

// Lock will erase all keys from memory and prevent the wallet from spending
// coins until it is unlocked.
func (w *Wallet) Lock() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if !w.unlocked {
		return modules.ErrLockedWallet
	}
	if w.persist.EncryptionVerification == nil {
		return errors.New("cannot lock an unencrypted wallet")
	}
	w.log.Println("INFO: Locking wallet.")

	// Wipe all of the seeds and secret keys, they will be replaced upon
	// calling 'Unlock' again.
	w.wipeSecrets()
	w.unlocked = false
	return nil
}

// Unlock will decrypt the wallet seed and load all of the addresses into
// memory.
func (w *Wallet) Unlock(masterKey crypto.TwofishKey) error {
	// By having the wallet's ThreadGroup track the Unlock method, we ensure
	// that Unlock will never unlock the wallet once the ThreadGroup has been
	// stopped. Without this precaution, the wallet's Close method would be
	// unsafe because it would theoretically be possible for another function
	// to Unlock the wallet in the short interval after Close calls w.Lock
	// and before Close calls w.mu.Lock.
	if err := w.tg.Add(); err != nil {
		return err
	}
	defer w.tg.Done()
	w.log.Println("INFO: Unlocking wallet.")

	// Initialize all of the keys in the wallet under a lock. While holding the
	// lock, also grab the subscriber status.
	return w.managedUnlock(masterKey)
}
