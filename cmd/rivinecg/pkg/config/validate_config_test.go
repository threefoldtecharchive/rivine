package config

import (
	"testing"

	validator "gopkg.in/go-playground/validator.v9"
)

func TestValidateValidConfigFile(t *testing.T) {
	conf := BuildConfigStruct("", nil)
	err := validate.Struct(conf)
	if err != nil {
		t.Errorf("%s", err)
	}

	// this check is only needed when your code could produce
	// an invalid value for validation such as interface with nil
	// value most including myself do not usually have code like this.
	if _, ok := err.(*validator.InvalidValidationError); ok {
		t.Errorf("Something went wrong with validating, %s", err)
	}
}

func TestValidateInvalidConfigFileWithoutBlockchainProperty(t *testing.T) {
	conf := BuildConfigStruct("", nil)
	conf.Blockchain = nil
	err := validate.Struct(conf)
	// this check is only needed when your code could produce
	// an invalid value for validation such as interface with nil
	// value most including myself do not usually have code like this.
	if _, ok := err.(*validator.InvalidValidationError); ok {
		t.Errorf("Something went wrong with validating, %s", err)
	}

	expectedError := "Key: 'Config.Blockchain' Error:Field validation for 'Blockchain' failed on the 'required' tag"
	if err != nil && err.Error() != expectedError {
		t.Errorf("%s", err)
	}
}

func TestValidateInvalidConfigFileWithoutBlockchainNameProperty(t *testing.T) {
	conf := BuildConfigStruct("", nil)
	conf.Blockchain.Name = ""
	err := validate.Struct(conf)
	// this check is only needed when your code could produce
	// an invalid value for validation such as interface with nil
	// value most including myself do not usually have code like this.
	if _, ok := err.(*validator.InvalidValidationError); ok {
		t.Errorf("Something went wrong with validating, %s", err)
	}

	expectedError := "Key: 'Config.Blockchain.Name' Error:Field validation for 'Name' failed on the 'required' tag"
	if err != nil && err.Error() != expectedError {
		t.Errorf("%s", err)
	}
}

func TestValidateInvalidConfigFileWithoutBlockchainRepositoryProperty(t *testing.T) {
	conf := BuildConfigStruct("", nil)
	conf.Blockchain.Repository = ""
	err := validate.Struct(conf)
	// this check is only needed when your code could produce
	// an invalid value for validation such as interface with nil
	// value most including myself do not usually have code like this.
	if _, ok := err.(*validator.InvalidValidationError); ok {
		t.Errorf("Something went wrong with validating, %s", err)
	}

	expectedError := "Key: 'Config.Blockchain.Repository' Error:Field validation for 'Repository' failed on the 'required' tag"
	if err != nil && err.Error() != expectedError {
		t.Errorf("%s", err)
	}
}

func TestValidateInvalidConfigFileWithoutCurrencyProperty(t *testing.T) {
	conf := BuildConfigStruct("", nil)
	conf.Blockchain.Currency = nil
	err := validate.Struct(conf)

	// this check is only needed when your code could produce
	// an invalid value for validation such as interface with nil
	// value most including myself do not usually have code like this.
	if _, ok := err.(*validator.InvalidValidationError); ok {
		t.Errorf("Something went wrong with validating, %s", err)
	}

	expectedError := "Key: 'Config.Blockchain.Currency' Error:Field validation for 'Currency' failed on the 'required' tag"
	if err != nil && err.Error() != expectedError {
		t.Errorf("%s", err.Error())
	}
}

func TestValidateInvalidConfigFileWithoutNetworkProperty(t *testing.T) {
	conf := BuildConfigStruct("", nil)
	conf.Blockchain.Networks = nil
	err := validate.Struct(conf)

	// this check is only needed when your code could produce
	// an invalid value for validation such as interface with nil
	// value most including myself do not usually have code like this.
	if _, ok := err.(*validator.InvalidValidationError); ok {
		t.Errorf("Something went wrong with validating, %s", err)
	}

	expectedError := "Key: 'Config.Blockchain.Networks' Error:Field validation for 'Networks' failed on the 'gt' tag"
	if err != nil && err.Error() != expectedError {
		t.Errorf("%s", err.Error())
	}
}

func TestValidateInvalidConfigFileWithEmptyNetworkProperty(t *testing.T) {
	conf := BuildConfigStruct("", nil)
	conf.Blockchain.Networks = make(map[string]*Network)
	err := validate.Struct(conf)

	// this check is only needed when your code could produce
	// an invalid value for validation such as interface with nil
	// value most including myself do not usually have code like this.
	if _, ok := err.(*validator.InvalidValidationError); ok {
		t.Errorf("Something went wrong with validating, %s", err)
	}

	expectedError := "Key: 'Config.Blockchain.Networks' Error:Field validation for 'Networks' failed on the 'gt' tag"
	if err != nil && err.Error() != expectedError {
		t.Errorf("%s", err.Error())
	}
}

func TestValidateInvalidConfigFileWithoutGenesisMintingProperty(t *testing.T) {
	conf := BuildConfigStruct("", nil)
	conf.Blockchain.Networks["testnet"].Genesis.Minting = nil
	err := validate.Struct(conf)

	// this check is only needed when your code could produce
	// an invalid value for validation such as interface with nil
	// value most including myself do not usually have code like this.
	if _, ok := err.(*validator.InvalidValidationError); ok {
		t.Errorf("Something went wrong with validating, %s", err)
	}

	expectedError := "Key: 'Config.Blockchain.Networks[testnet].Genesis.Minting' Error:Field validation for 'Minting' failed on the 'required_with' tag"
	if err != nil && err.Error() != expectedError {
		t.Errorf("%s", err.Error())
	}
}

func TestValidateInvalidConfigFileWithoutGenesisProperty(t *testing.T) {
	conf := BuildConfigStruct("", nil)
	conf.Blockchain.Networks["testnet"].Genesis = nil
	err := validate.Struct(conf)

	// this check is only needed when your code could produce
	// an invalid value for validation such as interface with nil
	// value most including myself do not usually have code like this.
	if _, ok := err.(*validator.InvalidValidationError); ok {
		t.Errorf("Something went wrong with validating, %s", err)
	}

	expectedError := "Key: 'Config.Blockchain.Networks[testnet].Genesis' Error:Field validation for 'Genesis' failed on the 'required' tag"
	if err != nil && err.Error() != expectedError {
		t.Errorf("%s", err.Error())
	}
}

func TestValidateInvalidConfigFileWithoutPortsProperty(t *testing.T) {
	conf := BuildConfigStruct("", nil)
	conf.Blockchain.Ports = nil
	err := validate.Struct(conf)

	// this check is only needed when your code could produce
	// an invalid value for validation such as interface with nil
	// value most including myself do not usually have code like this.
	if _, ok := err.(*validator.InvalidValidationError); ok {
		t.Errorf("Something went wrong with validating, %s", err)
	}

	expectedError := "Key: 'Config.Blockchain.Ports' Error:Field validation for 'Ports' failed on the 'required' tag"
	if err != nil && err.Error() != expectedError {
		t.Errorf("%s", err.Error())
	}
}

func TestValidateInvalidConfigFileWithSameAPIAndRPCProperty(t *testing.T) {
	conf := BuildConfigStruct("", nil)
	conf.Blockchain.Ports.API = 1024
	conf.Blockchain.Ports.RPC = 1024
	err := validate.Struct(conf)

	// this check is only needed when your code could produce
	// an invalid value for validation such as interface with nil
	// value most including myself do not usually have code like this.
	if _, ok := err.(*validator.InvalidValidationError); ok {
		t.Errorf("Something went wrong with validating, %s", err)
	}

	expectedError := `Key: 'Config.Blockchain.Ports.API' Error:Field validation for 'API' failed on the 'nefield' tag
Key: 'Config.Blockchain.Ports.RPC' Error:Field validation for 'RPC' failed on the 'nefield' tag`
	if err != nil && err.Error() != expectedError {
		t.Errorf("%s", err.Error())
	}
}

func TestValidateValidConfigWithoutOptionalParameters(t *testing.T) {
	conf := BuildConfigStruct("", nil)
	conf.Blockchain.Binaries = nil
	conf.Blockchain.Transactions.Default = nil
	conf.Blockchain.Networks["testnet"].TransactionFeePool = ""
	err := validate.Struct(conf)

	// this check is only needed when your code could produce
	// an invalid value for validation such as interface with nil
	// value most including myself do not usually have code like this.
	if _, ok := err.(*validator.InvalidValidationError); ok {
		t.Errorf("Something went wrong with validating, %s", err)
	}

	if err != nil {
		t.Errorf("%s", err.Error())
	}
}

func TestValidateConfigWithLeavingOutAllOptionalParametersShouldFillInAllParamsDefaults(t *testing.T) {
	var err error
	conf := BuildConfigStruct("", nil)
	conf.Template.Version = ""
	conf.Template.Repository = ""
	conf.Blockchain.Binaries.Daemon = ""
	conf.Blockchain.Binaries.Client = ""

	conf, err = assignDefaultValues(conf)
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	if conf.Template.Version != "master" {
		t.Errorf("Something went wrong with setting default value for template version")
	}
	templOwner, templRepo, err := githubOwnerAndRepoFromString(conf.Template.Repository)
	if err != nil {
		t.Error(err)
	}
	if templOwner != "threefoldtech" {
		t.Errorf("Something went wrong with setting default value for template repository owner")
	}
	if templRepo != "rivine-chain-template" {
		t.Errorf("Something went wrong with setting default value for template repository repo")
	}
	if conf.Blockchain.Binaries.Daemon != "pkgd" {
		t.Errorf("%s", conf.Blockchain.Binaries.Daemon)
		t.Errorf("Something went wrong with setting default value for blockchain binaries daemon")
	}
	if conf.Blockchain.Binaries.Client != "pkgc" {
		t.Errorf("Something went wrong with setting default value for blockchain binaries client")
	}
}

func TestValidateConfigWithLeavingOutTransactionsDefaultShouldFillItIn(t *testing.T) {
	var err error
	conf := BuildConfigStruct("", nil)
	conf.Blockchain.Transactions.Default = nil

	conf, err = assignDefaultValues(conf)
	if err != nil {
		t.Errorf("%s", err.Error())
	}
	if conf.Blockchain.Transactions.Default.Version != 1 {
		t.Errorf("Something went wrong with setting default value for blockchain transaction default version")
	}
}

func TestValidateConfigWithLeavingOutBinariesShouldFillItIn(t *testing.T) {
	var err error
	conf := BuildConfigStruct("", nil)
	conf.Blockchain.Binaries = nil

	conf, err = assignDefaultValues(conf)
	if err != nil {
		t.Errorf("%s", err.Error())
	}
	if conf.Blockchain.Binaries.Daemon != "pkgd" {
		t.Errorf("Something went wrong with setting default value for blockchain binaries daemon")
	}
	if conf.Blockchain.Binaries.Client != "pkgc" {
		t.Errorf("Something went wrong with setting default value for blockchain binaries client")
	}
}

func TestValidateConfigWithLeavingOutNetworkBootstrapPeersShouldThrowError(t *testing.T) {
	var err error
	conf := BuildConfigStruct("", nil)
	conf.Blockchain.Networks["testnet"].BootstrapPeers = nil

	err = validateConfig(conf)
	expectedError := "Key: 'Config.Blockchain.Networks[testnet].BootstrapPeers' Error:Field validation for 'BootstrapPeers' failed on the 'required' tag"
	if err != nil && err.Error() != expectedError {
		t.Errorf("%s", err.Error())
	}
}

func TestValidateConfigWithFaultyNetworkBootstrapPeersShouldThrowError(t *testing.T) {
	var err error
	conf := BuildConfigStruct("", nil)
	conf.Blockchain.Networks["testnet"].BootstrapPeers = []*BootstrapPeer{
		&BootstrapPeer{"invalid"},
	}

	delete(conf.Blockchain.Networks, "devnet")
	delete(conf.Blockchain.Networks, "standard")

	err = validateConfig(conf)
	expectedError := "address invalid: missing port in address"
	if err == nil {
		t.Errorf("no error, expected: %s", expectedError)
	}
	if err.Error() != expectedError {
		t.Errorf("%s", err.Error())
	}
}
