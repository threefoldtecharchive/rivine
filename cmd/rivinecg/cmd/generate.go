package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/threefoldtech/rivine/modules"

	"github.com/spf13/cobra"
	"github.com/threefoldtech/rivine/cmd/rivinecg/pkg/config"
	"github.com/threefoldtech/rivine/crypto"
	"github.com/threefoldtech/rivine/types"
	"github.com/tyler-smith/go-bip39"
)

// root generate command
var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "generate blockchains and related content",
	Args:  cobra.ExactArgs(0),
}

// sub generate seed command
var generateSeedCmd = &cobra.Command{
	Use:   "seed",
	Short: "generate a seed and one or multiple addresses",
	Args:  cobra.ExactArgs(0),
	RunE:  generateSeed,
}

var generateConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "generate blockchain config file",
	Args:  cobra.ExactArgs(0),
	RunE:  generateConfigFile,
}

func generateConfigFile(cmd *cobra.Command, args []string) error {
	return config.GenerateConfigFileYaml(generateConfigCfg.filePath)
}

var generateConfigCfg struct {
	fileType string
	filePath string
}

// sub generate seed command config
var generateSeedCfg struct {
	NumberOfAddresses uint64
}

// generateSeed generates amount of mnemonic and amount of addresses based on provided amount and outputs this to the cli
func generateSeed(cmd *cobra.Command, args []string) error {
	if generateSeedCfg.NumberOfAddresses == 0 {
		return errors.New("Amount of addresses cannot be below 1")
	}
	mnemonic, addresses, err := generateMnemonicAndAddresses(generateSeedCfg.NumberOfAddresses)
	if err != nil {
		return fmt.Errorf("Error when generating mnemonic and addresses: %v", err)
	}

	fmt.Println(mnemonic)
	for _, addr := range addresses {
		fmt.Println(addr)
	}
	return nil
}

func init() {
	// adds a address-amount flag to generate seed command
	generateSeedCmd.Flags().Uint64VarP(&generateSeedCfg.NumberOfAddresses, "address-amount", "n", 1, "amount of generated addresses")

	generateConfigCmd.Flags().StringVarP(&generateConfigCfg.fileType, "file-type", "f", "", "file type can be (YAML, TOML, JSON)")
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	generateConfigCmd.Flags().StringVarP(&generateConfigCfg.filePath, "file-path", "p", home, "file path where to config file will be stored")

	// adds generateSeedCmd to rootCmd
	generateCmd.AddCommand(
		generateSeedCmd,
		generateConfigCmd,
	)
}

// generateMnemonicAndAddresses generates mnemonic and amount of addresses based on provided amount
func generateMnemonicAndAddresses(n uint64) (string, []types.UnlockHash, error) {
	// generate entropy
	// Generate a mnemonic for memorization or user-friendly seeds
	entropy, err := bip39.NewEntropy(256)
	if err != nil {
		return "", nil, err
	}
	// generate mnemonic
	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		return "", nil, err
	}
	// call generateAddressesFromMnemonic and return mnemonic, unlockhashes and error
	unlockHashes, err := generateAddressesFromMnemonic(mnemonic, n)
	if err != nil {
		return "", nil, err
	}
	return mnemonic, unlockHashes, nil
}

// generateAddressesFromMnemonic generates amount of addresses based on mnemonic and amount provided
func generateAddressesFromMnemonic(mnemonic string, n uint64) ([]types.UnlockHash, error) {
	unlockhashes := make([]types.UnlockHash, 0, n)
	for index := uint64(0); index < n; index++ {
		seed, err := modules.InitialSeedFromMnemonic(mnemonic)
		if err != nil {
			return nil, err
		}
		_, pkey := crypto.GenerateKeyPairDeterministic(crypto.HashAll(seed, index))
		unlockhashes = append(unlockhashes, types.NewPubKeyUnlockHash(types.Ed25519PublicKey(pkey)))
	}
	return unlockhashes, nil
}
