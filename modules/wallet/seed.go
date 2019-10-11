package wallet

import (
	"bytes"
	"crypto/rand"
	"errors"
	"path/filepath"

	"github.com/threefoldtech/rivine/build"
	"github.com/threefoldtech/rivine/crypto"
	"github.com/threefoldtech/rivine/modules"
	"github.com/threefoldtech/rivine/persist"
	"github.com/threefoldtech/rivine/types"
)

const (
	seedFilePartialPrefix = " Wallet Encrypted Backup Seed - "
	seedFileSuffix        = ".seed"
)

var (
	errAddressExhaustion = errors.New("current seed has used all available addresses")
	errKnownSeed         = errors.New("seed is already known")
)

type (
	// UniqueID is a unique id randomly generated and put at the front of every
	// persistence object. It is used to make sure that a different encryption
	// key can be used for every persistence object.
	UniqueID [crypto.EntropySize]byte

	// SeedFile stores an encrypted wallet seed on disk.
	SeedFile struct {
		UID                    UniqueID
		EncryptionVerification crypto.Ciphertext
		Seed                   crypto.Ciphertext
	}
)

// generateSpendableKey creates the keys and unlock conditions for seed at a
// given index.
func generateSpendableKey(seed modules.Seed, index uint64) (spendableKey, error) {
	h, err := crypto.HashAll(seed, index)
	if err != nil {
		return spendableKey{}, err
	}
	sk, pk := crypto.GenerateKeyPairDeterministic(h)
	return spendableKey{
		PublicKey: pk,
		SecretKey: sk,
		Index:     index,
	}, nil
}

// encryptAndSaveSeedFile encrypts and saves a seed file.
func (w *Wallet) encryptAndSaveSeedFile(masterKey crypto.TwofishKey, seed modules.Seed) (SeedFile, error) {
	var uid UniqueID
	_, err := rand.Read(uid[:])
	if err != nil {
		return SeedFile{}, err
	}
	sek, err := uidEncryptionKey(masterKey, uid)
	if err != nil {
		return SeedFile{}, err
	}
	plaintextVerification := make([]byte, encryptionVerificationLen)
	verification := sek.EncryptBytes(plaintextVerification)
	encryptedSeed := sek.EncryptBytes(seed[:])
	return w.saveSeedFile(uid, encryptedSeed, verification)
}

// savePlainSeedFile saves the seed directly into the file without encrypting it.
func (w *Wallet) savePlainSeedFile(seed modules.Seed) (SeedFile, error) {
	if w.persist.EncryptionVerification != nil {
		return SeedFile{}, modules.ErrEncryptedWallet
	}
	var uid UniqueID
	_, err := rand.Read(uid[:])
	if err != nil {
		return SeedFile{}, err
	}
	return w.saveSeedFile(uid, crypto.Ciphertext(seed[:]), crypto.Ciphertext{})
}

// saveSeedFile defines the common logic to define and JSON-store the seed file.
func (w *Wallet) saveSeedFile(uid UniqueID, seed, verification crypto.Ciphertext) (SeedFile, error) {
	var sf SeedFile
	sf.UID = uid
	sf.Seed = seed
	sf.EncryptionVerification = verification
	seedFilename := filepath.Join(w.persistDir,
		w.bcInfo.Name+seedFilePartialPrefix+persist.RandomSuffix()+seedFileSuffix)
	err := persist.SaveJSON(seedMetadata, sf, seedFilename)
	if err != nil {
		return SeedFile{}, err
	}
	return sf, nil
}

// decryptSeedFile decrypts a seed file using the encryption key.
func decryptSeedFile(masterKey crypto.TwofishKey, sf SeedFile) (seed modules.Seed, err error) {
	// Verify that the provided master key is the correct key.
	decryptionKey, err := uidEncryptionKey(masterKey, sf.UID)
	if err != nil {
		return modules.Seed{}, err
	}
	expectedDecryptedVerification := make([]byte, crypto.EntropySize)
	decryptedVerification, err := decryptionKey.DecryptBytes(sf.EncryptionVerification)
	if err != nil {
		return modules.Seed{}, err
	}
	if !bytes.Equal(expectedDecryptedVerification, decryptedVerification) {
		return modules.Seed{}, modules.ErrBadEncryptionKey
	}

	// Decrypt and return the seed.
	plainSeed, err := decryptionKey.DecryptBytes(sf.Seed)
	if err != nil {
		return modules.Seed{}, err
	}
	copy(seed[:], plainSeed)
	return seed, nil
}

// loadPlainSeedFile loads a plain seed file directly as is.
func loadPlainSeedFile(sf SeedFile) (modules.Seed, error) {
	if len(sf.EncryptionVerification) != 0 {
		return modules.Seed{}, errors.New("unexpected encryption verification in plain seed file")
	}
	var seed modules.Seed
	copy(seed[:], sf.Seed[:])
	return seed, nil
}

// integrateSeed takes an address seed as input and from that generates
// 'publicKeysPerSeed' addresses that the wallet is able to spend.
// integrateSeed should not be called with the primary seed.
func (w *Wallet) integrateSeed(seed modules.Seed) error {
	for i := uint64(0); i < modules.PublicKeysPerSeed; i++ {
		// Generate the key and check it is new to the wallet.
		spendableKey, err := generateSpendableKey(seed, i)
		if err != nil {
			return err
		}
		uh, err := spendableKey.UnlockHash()
		if err != nil {
			return err
		}
		w.keys[uh] = spendableKey
	}
	w.seeds = append(w.seeds, seed)
	return nil
}

// recoverSeed integrates a recovery seed into the wallet.
func (w *Wallet) recoverEncryptedSeed(masterKey crypto.TwofishKey, seed modules.Seed) error {
	return w.recoverSeed(seed, func(modules.Seed) (SeedFile, error) {
		return w.encryptAndSaveSeedFile(masterKey, seed)
	})
}

// recoverPlainSeed integrates a recovery seed into the wallet, without encrypting it.
func (w *Wallet) recoverPlainSeed(seed modules.Seed) error {
	if w.persist.EncryptionVerification != nil {
		return modules.ErrEncryptedWallet
	}
	return w.recoverSeed(seed, w.savePlainSeedFile)
}

func (w *Wallet) recoverSeed(seed modules.Seed, fs func(modules.Seed) (SeedFile, error)) error {
	// Because the recovery seed does not have a UID, duplication must be
	// prevented by comparing with the list of decrypted seeds. This can only
	// occur while the wallet is unlocked.
	if !w.unlocked {
		return modules.ErrLockedWallet
	}

	// Check that the seed is not already known.
	for _, wSeed := range w.seeds {
		if seed == wSeed {
			return errKnownSeed
		}
	}
	if seed == w.primarySeed {
		return errKnownSeed
	}
	seedFile, err := fs(seed)
	if err != nil {
		return err
	}

	// Add the seed file to the wallet's set of tracked seeds and save the
	// wallet settings.
	w.persist.AuxiliarySeedFiles = append(w.persist.AuxiliarySeedFiles, seedFile)
	err = w.saveSettingsSync()
	if err != nil {
		return err
	}
	return w.integrateSeed(seed)
}

// createEncryptedSeed creates a wallet seed and encrypts it using a key derived from
// the master key, then addds it to the wallet as the primary seed, while
// making a disk backup.
func (w *Wallet) createEncryptedSeed(masterKey crypto.TwofishKey, seed modules.Seed, depth uint64) error {
	return w.createSeed(seed, depth, func(seed modules.Seed) (SeedFile, error) {
		return w.encryptAndSaveSeedFile(masterKey, seed)
	})
}

// createPlainSeed creates a wallet seed,
// then addds it to the wallet as the primary seed, while making a disk backup.
func (w *Wallet) createPlainSeed(seed modules.Seed, depth uint64) error {
	if w.persist.EncryptionVerification != nil {
		return modules.ErrEncryptedWallet
	}
	return w.createSeed(seed, depth, w.savePlainSeedFile)
}

func (w *Wallet) createSeed(seed modules.Seed, depth uint64, fs func(modules.Seed) (SeedFile, error)) error {
	seedFile, err := fs(seed)
	if err != nil {
		return err
	}
	w.primarySeed = seed
	w.persist.PrimarySeedFile = seedFile
	w.persist.PrimarySeedProgress = depth - modules.WalletSeedPreloadDepth
	// The wallet preloads keys to prevent confusion for people using the same
	// seed/wallet file in multiple places.
	for i := uint64(0); i < depth; i++ {
		spendableKey, err := generateSpendableKey(seed, i)
		if err != nil {
			return err
		}
		uh, err := spendableKey.UnlockHash()
		if err != nil {
			return err
		}
		w.keys[uh] = spendableKey
	}
	return w.saveSettingsSync()
}

// initEncryptedPrimarySeed loads the primary seed into the wallet.
func (w *Wallet) initEncryptedPrimarySeed(masterKey crypto.TwofishKey) error {
	return w.initPrimarySeed(func(file SeedFile) (modules.Seed, error) {
		return decryptSeedFile(masterKey, file)
	})
}

// initPlainPrimarySeed loads the primary seed into the wallet.
func (w *Wallet) initPlainPrimarySeed() error {
	if w.persist.EncryptionVerification != nil {
		return modules.ErrEncryptedWallet
	}
	return w.initPrimarySeed(loadPlainSeedFile)
}

func (w *Wallet) initPrimarySeed(sf func(SeedFile) (modules.Seed, error)) error {
	seed, err := sf(w.persist.PrimarySeedFile)
	if err != nil {
		return err
	}
	// The wallet preloads keys to prevent confusion when using the same wallet
	// in multiple places.
	for i := uint64(0); i < w.persist.PrimarySeedProgress+modules.WalletSeedPreloadDepth; i++ {
		spendableKey, err := generateSpendableKey(seed, i)
		if err != nil {
			return err
		}
		uh, err := spendableKey.UnlockHash()
		if err != nil {
			return err
		}
		w.keys[uh] = spendableKey
	}
	w.primarySeed = seed
	w.seeds = append(w.seeds, seed)
	return nil
}

// initEncryptedAuxiliarySeeds scans the wallet folder for wallet seeds.
func (w *Wallet) initEncryptedAuxiliarySeeds(masterKey crypto.TwofishKey) error {
	return w.initAuxiliarySeeds(func(file SeedFile) (modules.Seed, error) {
		return decryptSeedFile(masterKey, file)
	})
}

// initPlainAuxiliarySeeds scans the wallet folder for wallet seeds.
func (w *Wallet) initPlainAuxiliarySeeds() error {
	if w.persist.EncryptionVerification != nil {
		return modules.ErrEncryptedWallet
	}
	return w.initAuxiliarySeeds(loadPlainSeedFile)
}

func (w *Wallet) initAuxiliarySeeds(sf func(SeedFile) (modules.Seed, error)) error {
	for _, seedFile := range w.persist.AuxiliarySeedFiles {
		seed, err := sf(seedFile)
		if err != nil {
			build.Severe(err)
		}
		if err != nil {
			w.log.Println("UNLOCK: failed to load an auxiliary seed:", err)
			continue
		}
		err = w.integrateSeed(seed)
		if err != nil {
			return err
		}
	}
	return nil
}

// nextPrimarySeedAddress fetches the next address from the primary seed.
func (w *Wallet) nextPrimarySeedAddress() (types.UnlockHash, error) {
	// Check that the wallet has been unlocked.
	if !w.unlocked {
		return types.UnlockHash{}, modules.ErrLockedWallet
	}

	// Integrate the next key into the wallet, and return the unlock
	// conditions. Because the wallet preloads keys, the progress used is
	// 'PrimarySeedProgress+modules.WalletSeedPreloadDepth'.
	spendableKey, err := generateSpendableKey(w.primarySeed, w.persist.PrimarySeedProgress+modules.WalletSeedPreloadDepth)
	if err != nil {
		return types.UnlockHash{}, err
	}
	uh, err := spendableKey.UnlockHash()
	if err != nil {
		return types.UnlockHash{}, err
	}
	w.keys[uh] = spendableKey
	w.persist.PrimarySeedProgress++
	err = w.saveSettingsSync()
	if err != nil {
		return types.UnlockHash{}, err
	}
	return spendableKey.UnlockHash()
}

// AllSeeds returns a list of all seeds known to and used by the wallet.
func (w *Wallet) AllSeeds() ([]modules.Seed, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if !w.unlocked {
		return nil, modules.ErrLockedWallet
	}
	return w.seeds, nil
}

// PrimarySeed returns the decrypted primary seed of the wallet.
func (w *Wallet) PrimarySeed() (modules.Seed, uint64, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if !w.unlocked {
		return modules.Seed{}, 0, modules.ErrLockedWallet
	}
	return w.primarySeed, w.persist.PrimarySeedProgress, nil
}

// NextAddress returns an unlock hash that is ready to receive siacoins or
// siafunds. The address is generated using the primary address seed.
func (w *Wallet) NextAddress() (types.UnlockHash, error) {
	if err := w.tg.Add(); err != nil {
		return types.UnlockHash{}, err
	}
	defer w.tg.Done()
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.nextPrimarySeedAddress()
}

// LoadSeed will track all of the addresses generated by the input seed,
// reclaiming any funds that were lost due to a deleted file or lost encryption
// key. An error will be returned if the seed has already been integrated with
// the wallet.
func (w *Wallet) LoadSeed(masterKey crypto.TwofishKey, seed modules.Seed) error {
	if err := w.tg.Add(); err != nil {
		return err
	}
	defer w.tg.Done()
	w.mu.Lock()
	defer w.mu.Unlock()
	err := w.checkMasterKey(masterKey)
	if err != nil {
		return err
	}
	return w.recoverEncryptedSeed(masterKey, seed)
}

// LoadPlainSeed will track all of the addresses generated by the input seed,
// reclaiming any funds that were lost due to a deleted/lost file.
// An error will be returned if the seed has already been integrated with the wallet.
func (w *Wallet) LoadPlainSeed(seed modules.Seed) error {
	if err := w.tg.Add(); err != nil {
		return err
	}
	defer w.tg.Done()
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.recoverPlainSeed(seed)
}
