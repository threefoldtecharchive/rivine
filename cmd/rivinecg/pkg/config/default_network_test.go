package config

import (
	"fmt"
	"testing"
)

func TestValidateConfigWithAssigningDefaultNetworkValues(t *testing.T) {
	conf := BuildConfigStruct()
	network := conf.Blockchain.Network["testnet"]
	network.ArbitraryDataSizeLimit = 0
	network.BlockCreatorFee = ""
	network.BlockSizeLimit = 0
	network.MaxAdjustmentDown = Fraction{Denominator: 0, Numerator: 0}
	network.TransactionPool.PoolSizeLimit = 0

	network = assignDefaultNetworkProps(network)
	if network.ArbitraryDataSizeLimit == 0 && network.ArbitraryDataSizeLimit != 83 {
		t.Errorf("Something went wrong with setting default value for ArbitraryDataSizeLimit")
	}
	if network.BlockCreatorFee == "" && network.BlockCreatorFee != "0.0" {
		t.Errorf("Something went wrong with setting default value for BlockCreatorFee")
	}
	if network.BlockSizeLimit == 0 && network.BlockSizeLimit != 2e6 {
		t.Errorf("Something went wrong with setting default value for BlockSizeLimit")
	}
	if network.MaxAdjustmentDown.Denominator == 0 && network.MaxAdjustmentDown.Numerator == 0 {
		t.Errorf("Something went wrong with setting default value for MaxAdjustmentDown")
	}
	if network.TransactionPool.PoolSizeLimit == 0 && network.TransactionPool.PoolSizeLimit != 2e6-5e3-250e3 {
		t.Errorf("Something went wrong with setting default value for TransactionPool PoolSizeLimit")
	}
	fmt.Println(network.MaxAdjustmentDown)
}
