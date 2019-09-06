package config

import (
	"testing"
)

func TestValidateConfigWithAssigningDefaultNetworkValues(t *testing.T) {
	conf := BuildConfigStruct("", nil)
	network := conf.Blockchain.Networks["testnet"]
	network.ArbitraryDataSizeLimit = 0
	network.BlockCreatorFee = ""
	network.BlockSizeLimit = 0
	network.MaxAdjustmentDown = Fraction{Denominator: 0, Numerator: 0}
	network.TransactionPool.PoolSizeLimit = 0

	assignDefaultNetworkProps(network)
	if network.ArbitraryDataSizeLimit != 83 {
		t.Errorf("Something went wrong with setting default value for ArbitraryDataSizeLimit")
	}
	if network.BlockCreatorFee != "1.0" {
		t.Errorf("Something went wrong with setting default value for BlockCreatorFee")
	}
	if network.BlockSizeLimit != 2e6 {
		t.Errorf("Something went wrong with setting default value for BlockSizeLimit")
	}
	if network.MaxAdjustmentDown.Numerator == 0 {
		t.Errorf("Something went wrong with setting default value for MaxAdjustmentDown")
	}
	if network.TransactionPool.PoolSizeLimit != ActualPoolSize {
		t.Errorf("Something went wrong with setting default value for TransactionPool PoolSizeLimit")
	}
}
