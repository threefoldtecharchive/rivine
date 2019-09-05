package cmd

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/threefoldtech/rivine/modules"

	"github.com/spf13/cobra"
	"github.com/threefoldtech/rivine/cmd/rivinecg/pkg/config"
	"github.com/threefoldtech/rivine/crypto"
	"github.com/threefoldtech/rivine/types"
	bip39 "github.com/tyler-smith/go-bip39"
)

// root generate command
var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate blockchains and related content",
}

// sub generate seed command
var generateSeedCmd = &cobra.Command{
	Use:   "seed",
	Short: "Generate a seed and one or multiple addresses",
	Args:  cobra.ExactArgs(0),
	RunE:  generateSeed,
}

var generateConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Generate blockchain config file",
	Args:  cobra.ExactArgs(0),
	RunE:  generateConfigFile,
}

var generateBlockchainCmd = &cobra.Command{
	Use:   "blockchain",
	Short: "Generate blockchain from a config file",
	Long:  "Generate a blockchain from a config file, this blockchain will be stored in your GOPATH",
	Args:  cobra.ExactArgs(0),
	RunE:  generateBlockchain,
}

// generateSeed generates amount of mnemonic and amount of addresses based on provided amount and outputs this to the cli
func generateSeed(cmd *cobra.Command, args []string) error {
	if numberOfAddresses == 0 {
		return errors.New("Amount of addresses cannot be below 1")
	}
	mnemonic, addresses, err := generateMnemonicAndAddresses(numberOfAddresses)
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
	generateSeedCmd.Flags().Uint64VarP(&numberOfAddresses, "address-amount", "n", 1, "amount of generated addresses")

	generateConfigCmd.Flags().BoolVar(&pluginMintingEnabled, "minting", true, "enable minting plugin")
	generateConfigCmd.Flags().BoolVar(&pluginAuthcoinEnabled, "authcoin", true, "enable minting plugin")

	for _, cmd := range []*cobra.Command{generateConfigCmd, generateBlockchainCmd} {
		cmd.Flags().StringVarP(
			&filePath, "config", "c", "blockchaincfg.yaml",
			"file path of the config, ecoding is based on the file extension, can be yaml or json")
	}
	generateBlockchainCmd.Flags().StringVarP(
		&outputDir, "output", "o", "",
		"file path where the blockchain will be generated to")

	// adds generateSeedCmd to rootCmd
	generateCmd.AddCommand(
		generateSeedCmd,
		generateConfigCmd,
		generateBlockchainCmd,
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
	seed, err := modules.InitialSeedFromMnemonic(mnemonic)
	for index := uint64(0); index < n; index++ {
		if err != nil {
			return nil, err
		}
		h, err := crypto.HashAll(seed, index)
		if err != nil {
			return nil, err
		}
		_, pkey := crypto.GenerateKeyPairDeterministic(h)
		uh, err := types.NewPubKeyUnlockHash(types.Ed25519PublicKey(pkey))
		if err != nil {
			return nil, err
		}
		unlockhashes = append(unlockhashes, uh)
	}
	return unlockhashes, nil
}

func generateConfigFile(cmd *cobra.Command, args []string) error {
	err := config.GenerateConfigFile(filePath, &config.ConfigGenerationOpts{
		PluginMintingEnabled:  pluginMintingEnabled,
		PluginAuthcoinEnabled: pluginAuthcoinEnabled,
	})
	if err != nil {
		return err
	}
	fmt.Printf("Config written in: %s\n", filePath)
	return nil
}

func generateBlockchain(cmd *cobra.Command, args []string) error {
	dir := outputDir
	if dir == "" {
		dir = filepath.Dir(filePath)
	}
	err := config.GenerateBlockchain(filePath, dir)
	if err != nil {
		return err
	}
	return nil
}
