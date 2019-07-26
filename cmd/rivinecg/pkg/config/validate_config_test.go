package config

import (
	"testing"

	"gopkg.in/go-playground/validator.v9"
)

func TestValidateValidConfigFile(t *testing.T) {
	conf := BuildConfigStruct()
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
	conf := BuildConfigStruct()
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
	conf := BuildConfigStruct()
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
	conf := BuildConfigStruct()
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
	conf := BuildConfigStruct()
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
	conf := BuildConfigStruct()
	conf.Blockchain.Network = nil
	err := validate.Struct(conf)

	// this check is only needed when your code could produce
	// an invalid value for validation such as interface with nil
	// value most including myself do not usually have code like this.
	if _, ok := err.(*validator.InvalidValidationError); ok {
		t.Errorf("Something went wrong with validating, %s", err)
	}

	expectedError := "Key: 'Config.Blockchain.Network' Error:Field validation for 'Network' failed on the 'gt' tag"
	if err != nil && err.Error() != expectedError {
		t.Errorf("%s", err.Error())
	}
}

func TestValidateInvalidConfigFileWithEmptyNetworkProperty(t *testing.T) {
	conf := BuildConfigStruct()
	conf.Blockchain.Network = make(map[string]*Network)
	err := validate.Struct(conf)

	// this check is only needed when your code could produce
	// an invalid value for validation such as interface with nil
	// value most including myself do not usually have code like this.
	if _, ok := err.(*validator.InvalidValidationError); ok {
		t.Errorf("Something went wrong with validating, %s", err)
	}

	expectedError := "Key: 'Config.Blockchain.Network' Error:Field validation for 'Network' failed on the 'gt' tag"
	if err != nil && err.Error() != expectedError {
		t.Errorf("%s", err.Error())
	}
}

func TestValidateInvalidConfigFileWithoutGenesisMintingProperty(t *testing.T) {
	conf := BuildConfigStruct()
	conf.Blockchain.Network["testnet"].Genesis.Minting = nil
	err := validate.Struct(conf)

	// this check is only needed when your code could produce
	// an invalid value for validation such as interface with nil
	// value most including myself do not usually have code like this.
	if _, ok := err.(*validator.InvalidValidationError); ok {
		t.Errorf("Something went wrong with validating, %s", err)
	}

	expectedError := "Key: 'Config.Blockchain.Network[testnet].Genesis.Minting' Error:Field validation for 'Minting' failed on the 'required_with' tag"
	if err != nil && err.Error() != expectedError {
		t.Errorf("%s", err.Error())
	}
}

func TestValidateInvalidConfigFileWithoutGenesisProperty(t *testing.T) {
	conf := BuildConfigStruct()
	conf.Blockchain.Network["testnet"].Genesis = nil
	err := validate.Struct(conf)

	// this check is only needed when your code could produce
	// an invalid value for validation such as interface with nil
	// value most including myself do not usually have code like this.
	if _, ok := err.(*validator.InvalidValidationError); ok {
		t.Errorf("Something went wrong with validating, %s", err)
	}

	expectedError := "Key: 'Config.Blockchain.Network[testnet].Genesis' Error:Field validation for 'Genesis' failed on the 'required' tag"
	if err != nil && err.Error() != expectedError {
		t.Errorf("%s", err.Error())
	}
}

func TestValidateInvalidConfigFileWithoutPortsProperty(t *testing.T) {
	conf := BuildConfigStruct()
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
	conf := BuildConfigStruct()
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
	conf := BuildConfigStruct()
	conf.Blockchain.Binaries = nil
	conf.Blockchain.Transactions.Default = nil
	conf.Blockchain.Network["testnet"].TransactionFeePool = ""
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
	conf := BuildConfigStruct()
	conf.Template.Version = ""
	conf.Template.Repository = nil
	conf.Blockchain.Binaries.Daemon = ""
	conf.Blockchain.Binaries.Client = ""
	conf.Blockchain.Transactions.Default.Version = 0

	conf, err = assignDefaultValues(conf)
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	if conf.Template.Version != "master" {
		t.Errorf("Something went wrong with setting default value for template version")
	}
	if conf.Template.Repository.Owner != "threefoldtech" {
		t.Errorf("Something went wrong with setting default value for template repository owner")
	}
	if conf.Template.Repository.Repo != "rivine-chain-template" {
		t.Errorf("Something went wrong with setting default value for template repository repo")
	}
	if conf.Blockchain.Binaries.Daemon != "rivined" {
		t.Errorf("Something went wrong with setting default value for blockchain binaries daemon")
	}
	if conf.Blockchain.Binaries.Client != "rivinec" {
		t.Errorf("Something went wrong with setting default value for blockchain binaries client")
	}
	if conf.Blockchain.Transactions.Default.Version != 1 {
		t.Errorf("Something went wrong with setting default value for blockchain transaction default version")
	}
}

func TestValidateConfigWithLeavingOutTransactionsDefaultShouldFillItIn(t *testing.T) {
	var err error
	conf := BuildConfigStruct()
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
	conf := BuildConfigStruct()
	conf.Blockchain.Binaries = nil

	conf, err = assignDefaultValues(conf)
	if err != nil {
		t.Errorf("%s", err.Error())
	}
	if conf.Blockchain.Binaries.Daemon != "rivined" {
		t.Errorf("Something went wrong with setting default value for blockchain binaries daemon")
	}
	if conf.Blockchain.Binaries.Client != "rivinec" {
		t.Errorf("Something went wrong with setting default value for blockchain binaries client")
	}
}
