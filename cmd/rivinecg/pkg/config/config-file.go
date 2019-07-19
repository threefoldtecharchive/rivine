package config

import (
	"bytes"
	"fmt"
	"math/big"
	"time"

	"os"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

type NetworkType int

const (
	Testnet NetworkType = iota
	Devnet
	Standard
)

type Config struct {
	Template   Template
	Blockchain Blockchain
}

type Template struct {
	Repository string
	Version    string
}

type Blockchain struct {
	Name         string
	Repository   string
	Currency     Currency
	Ports        Ports
	Binaries     Binaries
	Transactions Transactions
	Network      []Network
}

type Currency struct {
	Unit      string
	Precision uint
}

type Ports struct {
	Api uint
	Rpc uint
}

type Binaries struct {
	Client string
	Deamon string
}

type Transactions struct {
	Default Version
	Minting Minting
}

type Minting struct {
	MintConditionUpdate Version
	CoinCreation        Version
	CoinDestruction     Version
	RequireMinerFees    bool
}

type Version struct {
	Version uint
}

type Genesis struct {
	CoinOutputs       []Output
	BlockStakeOutputs []Output
	Minting           GenesisMinting
}

type Output struct {
	Value     uint
	Condition string
}

type GenesisMinting struct {
	Condition          Condition
	SignaturesRequired uint
}

type Condition struct {
	Addresses []string
}

type Network struct {
	NetworkType            NetworkType
	Genesis                Genesis
	TransactionFeePool     string
	BlockSizeLimit         uint64
	ArbitraryDataSizeLimit uint
	BlockCreatorFee        float32
	MinimumTransactionFee  float32
	BlockFrequency         uint
	MaturityDelay          uint
	MedianTimestampWindow  uint
	TargetWindow           uint
	MaxAdjustmentUp        Fraction
	MaxAdjustmentDown      Fraction
	FutureThreshold        time.Duration
	ExtremeFutureThreshold time.Duration
	StakeModifierDelay     time.Duration
	BlockStakeAging        time.Duration
	TransactionPool        TransactionPool
}

// Fraction represents ratio.
type Fraction struct {
	Numerator, Denominator int
}

type TransactionPool struct {
	TransactionSizeLimit    uint
	TransactionSetSizeLimit uint
	PoolSizeLimit           uint64
}

type Rational struct {
	Value *big.Rat
}

func GenerateConfigFileYaml(filepath string) error {
	fmt.Println(filepath)
	config := &Config{
		Template{
			Repository: "https://github.com/threefoldtech/rivine-chain-template",
			Version:    "Master",
		},
		Blockchain{
			Name:       "rivine",
			Repository: "github.com/threefoldtech/rivine",
			Currency: Currency{
				Unit:      "ROC",
				Precision: uint(9),
			},
			Ports: Ports{
				Api: uint(23110),
				Rpc: uint(23112),
			},
			Binaries: Binaries{
				Client: "rivinec",
				Deamon: "rivined",
			},
			Transactions: Transactions{
				Default: Version{
					Version: uint(1),
				},
				Minting: Minting{
					MintConditionUpdate: Version{
						Version: uint(128),
					},
					CoinCreation: Version{
						Version: uint(129),
					},
					CoinDestruction: Version{
						Version: uint(130),
					},
					RequireMinerFees: false,
				},
			},
			Network: []Network{
				{
					Genesis: Genesis{
						CoinOutputs: []Output{
							{
								Value:     uint(100000000),
								Condition: "01b5e42056ef394f2ad9b511a61cec874d25bebe2095682dd37455cbafed4bec154e382a23f90e",
							},
						},
						BlockStakeOutputs: []Output{
							{
								Value:     uint(100000000),
								Condition: "01b5e42056ef394f2ad9b511a61cec874d25bebe2095682dd37455cbafed4bec154e382a23f90e",
							},
						},
						Minting: GenesisMinting{
							Condition: Condition{
								Addresses: []string{
									"01434535fd01243c02c277cd58d71423163767a575a8ae44e15807bf545e4a8456a5c4afabad51",
								},
							},
							SignaturesRequired: uint(2),
						},
					},
					TransactionFeePool:     "017267221ef1947bb18506e390f1f9446b995acfb6d08d8e39508bb974d9830b8cb8fdca788e34",
					BlockSizeLimit:         uint64(2e6),
					ArbitraryDataSizeLimit: uint(83),
					BlockCreatorFee:        float32(1.0),
					MinimumTransactionFee:  float32(0.1),
					BlockFrequency:         uint(120),
					MaturityDelay:          uint(144),
					MedianTimestampWindow:  uint(11),
					TargetWindow:           uint(1e3),
					MaxAdjustmentUp: Fraction{
						Numerator:   25,
						Denominator: 10,
					},
					MaxAdjustmentDown: Fraction{
						Numerator:   25,
						Denominator: 10,
					},
					FutureThreshold:        time.Hour,
					ExtremeFutureThreshold: time.Hour * 2,
					StakeModifierDelay:     time.Second * 2000,
					BlockStakeAging:        time.Hour * 24,
					TransactionPool: TransactionPool{
						TransactionSizeLimit:    uint(16e3),
						TransactionSetSizeLimit: uint(250e3),
						PoolSizeLimit:           uint64(2e6 - 5e3 - 250e3),
					},
				},
			},
		},
	}
	y, err := yaml.Marshal(config)
	if err != nil {
		return err
	}
	viper.SetConfigType("yaml") // or viper.SetConfigType("YAML")
	viper.SetConfigName("blockchainConfig")

	err = viper.ReadConfig(bytes.NewBuffer(y))
	if err != nil {
		return err
	}
	f, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer f.Close()

	return viper.WriteConfigAs(filepath)
}
