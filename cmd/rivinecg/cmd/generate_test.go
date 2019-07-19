package cmd

import (
	"testing"
)

func TestMnemonicAndAddressGeneration(t *testing.T) {
	const n = 1
	_, unlockhashes, err := generateMnemonicAndAddresses(n)
	if err != nil {
		t.Errorf("Error occured: %s", err)
	}
	if len(unlockhashes) != 1 {
		t.Errorf("Amount of addresses does not equal 1 %s, received %d amount of addresses", err, len(unlockhashes))
	}
	const mnemonic = "carbon boss inject cover mountain fetch fiber fit tornado cloth wing dinosaur proof joy intact fabric thumb rebel borrow poet chair network expire else"
	addresses, err := generateAddressesFromMnemonic(mnemonic, n)
	if err != nil {
		t.Errorf("Error occured: %s", err)
	}
	if len(addresses) != 1 {
		t.Errorf("Amount of addresses does not equal 1 %s, received %d amount of addresses", err, len(addresses))
	}
	address := addresses[0].String()
	const expectedAddress = "015df22a2e82a3323bc6ffbd1730450ed844feca711c8fe0c15e218c171962fd17b206263220ee"
	if address != expectedAddress {
		t.Errorf("Address generated from devnet mnemonic does not equal what is expected, received %s, expected %s", address, expectedAddress)
	}
}
