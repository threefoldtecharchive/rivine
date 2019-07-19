package config

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"path"
	"time"

	"github.com/pelletier/go-toml"
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
	Denominator, Numerator int
}

// MarshalJSON will marshall Fraction struct into our specific format
func (f Fraction) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBuffer(nil)
	buffer.WriteString(fmt.Sprintf("\"%d/%d\"", f.Denominator, f.Numerator))
	return buffer.Bytes(), nil
}

// MarshalTOML will marshall Fraction struct into our specific format
func (f Fraction) MarshalTOML() ([]byte, error) {
	buffer := bytes.NewBuffer(nil)
	buffer.WriteString(fmt.Sprintf("%d/%d", f.Denominator, f.Numerator))
	return buffer.Bytes(), nil
}

// MarshalYAML will marshall Fraction struct into our specific format
func (f Fraction) MarshalYAML() (interface{}, error) {
	return fmt.Sprintf("%d/%d", f.Denominator, f.Numerator), nil
}

type TransactionPool struct {
	TransactionSizeLimit    uint
	TransactionSetSizeLimit uint
	PoolSizeLimit           uint64
}

// GenerateConfigFile generates a blockchain config file
func GenerateConfigFile(filepath string, filetype string) error {
	marshalledConfig, err := generateMarshalledConfig(filetype)
	if err != nil {
		return err
	}

	viper.SetConfigType(filetype) // or viper.SetConfigType("YAML")

	err = viper.ReadConfig(bytes.NewBuffer(marshalledConfig))
	if err != nil {
		return err
	}

	filepath = path.Join(filepath, "blockchaincfg."+filetype)
	fmt.Println(filepath)
	return viper.WriteConfigAs(filepath)
}

// generateMarshalledConfig generates a marshalledConfig with default values
func generateMarshalledConfig(filetype string) ([]byte, error) {
	config := buildConfigStruct()
	var marshaller func(interface{}) ([]byte, error)
	switch filetype {
	case "yaml":
		marshaller = yaml.Marshal
	case "json":
		marshaller = json.Marshal
	case "toml":
		marshaller = toml.Marshal
	default:
		return nil, errors.New("Filetype not supported")
	}
	return marshaller(config)
}

// buildConfigStruct builds to default values in our config struct
func buildConfigStruct() *Config {
	return &Config{
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
						Denominator: 10,
						Numerator:   25,
					},
					MaxAdjustmentDown: Fraction{
						Denominator: 10,
						Numerator:   25,
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
}
