package config

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"gopkg.in/go-playground/validator.v9"
	"gopkg.in/yaml.v2"
)

type NetworkType int

const (
	Testnet NetworkType = iota
	Devnet
	Standard
)

type (
	Config struct {
		Template   Template   `json:"template,omitempty" yaml:"template,omitempty" toml:"template,omitempty"`
		Blockchain Blockchain `json:"blockchain,omitempty" yaml:"blockchain,omitempty" toml:"blockchain,omitempty" validate:"required"`
	}

	Template struct {
		Repository string `json:"repository,omitempty" yaml:"repository,omitempty" toml:"repository,omitempty"`
		Version    string `json:"version,omitempty" yaml:"version,omitempty" toml:"version,omitempty"`
	}

	Blockchain struct {
		Name         string             `json:"name,omitempty" yaml:"name,omitempty" toml:"name,omitempty" validate:"required"`
		Repository   string             `json:"repository,omitempty" yaml:"repository,omitempty" toml:"repository,omitempty" validate:"required"`
		Currency     Currency           `json:"currency,omitempty" yaml:"currency,omitempty" toml:"currency,omitempty" validate:"required"`
		Ports        Ports              `json:"ports,omitempty" yaml:"ports,omitempty" toml:"ports,omitempty" validate:"required"`
		Binaries     *Binaries          `json:"binaries,omitempty" yaml:"binaries,omitempty" toml:"binaries,omitempty"`
		Transactions Transactions       `json:"transactions,omitempty" yaml:"transactions,omitempty" toml:"transactions,omitempty"`
		Network      map[string]Network `json:"network,omitempty" yaml:"network,omitempty" toml:"network,omitempty"`
	}

	Currency struct {
		Unit      string `json:"unit,omitempty" yaml:"unit,omitempty" toml:"unit,omitempty" validate:"required"`
		Precision uint64 `json:"precision,omitempty" yaml:"precision,omitempty" toml:"precision,omitempty" validate:"required"`
	}

	Ports struct {
		API uint16 `json:"api,omitempty" yaml:"api,omitempty" toml:"api,omitempty" validate:"nefield=RPC"`
		RPC uint16 `json:"rpc,omitempty" yaml:"rpc,omitempty" toml:"rpc,omitempty" validate:"nefield=API"`
	}

	Binaries struct {
		Client string `json:"client,omitempty" yaml:"client,omitempty" toml:"client,omitempty"`
		Deamon string `json:"deamon,omitempty" yaml:"deamon,omitempty" toml:"deamon,omitempty"`
	}

	Transactions struct {
		Default Version  `json:"default,omitempty" yaml:"default,omitempty" toml:"default,omitempty"`
		Minting *Minting `json:"minting,omitempty" yaml:"minting,omitempty" toml:"minting,omitempty"`
	}

	Minting struct {
		ConditionUpdate  Version  `json:"conditionUpdate,omitempty" yaml:"conditionUpdate,omitempty" toml:"conditionUpdate,omitempty" validate:"required"`
		CoinCreation     Version  `json:"coinCreation,omitempty" yaml:"coinCreation,omitempty" toml:"coinCreation,omitempty" validate:"required"`
		CoinDestruction  *Version `json:"coinDestruction,omitempty" yaml:"coinDestruction,omitempty" toml:"coinDestruction,omitempty"`
		RequireMinerFees bool     `json:"requireMinerFees,omitempty" yaml:"requireMinerFees,omitempty" toml:"requireMinerFees,omitempty" validate:"required"`
	}

	Version struct {
		Version uint64 `json:"version,omitempty" yaml:"version,omitempty" toml:"version,omitempty" validate:"required"`
	}

	Genesis struct {
		CoinOutputs       []Output        `json:"coinOuput,omitempty" yaml:"coinOutput,omitempty" toml:"coinOutput,omitempty"`
		BlockStakeOutputs []Output        `json:"blockStakeOutputs,omitempty" yaml:"blockStakeOutputs,omitempty" toml:"blockStakeOutputs,omitempty"`
		Minting           *GenesisMinting `json:"minting,omitempty" yaml:"minting,omitempty" toml:"minting,omitempty"`
	}

	Output struct {
		Value     uint   `json:"value,omitempty" yaml:"value,omitempty" toml:"value,omitempty"`
		Condition string `json:"condition,omitempty" yaml:"condition,omitempty" toml:"condition,omitempty"`
	}

	GenesisMinting struct {
		Condition          Condition `json:"condition,omitempty" yaml:"condition,omitempty" toml:"condition,omitempty"`
		SignaturesRequired uint      `json:"signaturesRequired,omitempty" yaml:"signaturesRequired,omitempty" toml:"signaturesRequired,omitempty"`
	}

	Condition struct {
		Addresses []string `json:"addresses,omitempty"`
	}

	Fee struct {
	}

	Network struct {
		NetworkType            NetworkType     `json:"networkType,omitempty" yaml:"networkType,omitempty" toml:"networkType,omitempty"`
		Genesis                Genesis         `json:"genesis,omitempty" yaml:"genesis,omitempty" toml:"genesis,omitempty,required"`
		TransactionFeePool     string          `json:"transactionFeePool,omitempty" yaml:"transactionFeePool,omitempty" toml:"transactionFeePool,omitempty"`
		BlockSizeLimit         uint64          `json:"blockSizeLimit,omitempty" yaml:"blockSizeLimit,omitempty" toml:"blockSizeLimit,omitempty"`
		ArbitraryDataSizeLimit uint64          `json:"arbitraryDataSizeLimit,omitempty" yaml:"arbitraryDataSizeLimit,omitempty" toml:"arbitraryDataSizeLimit,omitempty"`
		BlockCreatorFee        float32         `json:"blockCreatorFee,omitempty" yaml:"blockCreatorFee,omitempty" toml:"blockCreatorFee,omitempty"`
		MinimumTransactionFee  float32         `json:"minimumTransactionFee,omitempty" yaml:"minimumTransactionFee,omitempty" toml:"minimumTransactionFee,omitempty"`
		BlockFrequency         uint64          `json:"blockFrequency,omitempty" yaml:"blockFrequency,omitempty" toml:"blockFrequency,omitempty"`
		MaturityDelay          uint64          `json:"maturityDelay,omitempty" yaml:"maturityDelay,omitempty" toml:"maturityDelay,omitempty"`
		MedianTimestampWindow  uint64          `json:"medianTimestampWindow,omitempty" yaml:"medianTimestampWindow,omitempty" toml:"medianTimestampWindow,omitempty"`
		TargetWindow           uint64          `json:"targetWindow,omitempty" yaml:"targetWindow,omitempty" toml:"targetWindow,omitempty"`
		MaxAdjustmentUp        Fraction        `json:"maxAdjustmentUp,omitempty" yaml:"maxAdjustmentUp,omitempty" toml:"maxAdjustmentUp,omitempty"`
		MaxAdjustmentDown      Fraction        `json:"maxAdjustmentDown,omitempty" yaml:"maxAdjustmentDown,omitempty" toml:"maxAdjustmentDown,omitempty"`
		FutureThreshold        time.Duration   `json:"futureTreshold,omitempty" yaml:"futureTreshold,omitempty" toml:"futureTreshold,omitempty"`
		ExtremeFutureThreshold time.Duration   `json:"extremeFutureTreshold,omitempty" yaml:"extremeFutureTreshold,omitempty" toml:"extremeFutureTreshold,omitempty"`
		StakeModifierDelay     time.Duration   `json:"stakeModifierDelay,omitempty" yaml:"stakeModifierDelay,omitempty" toml:"stakeModifierDelay,omitempty"`
		BlockStakeAging        time.Duration   `json:"blockStakeAging,omitempty" yaml:"blockStakeAging,omitempty" toml:"blockStakeAging,omitempty"`
		TransactionPool        TransactionPool `json:"transactionPool,omitempty" yaml:"transactionPool,omitempty" toml:"transactionPool,omitempty"`
	}

	// Fraction represents ratio.
	Fraction struct {
		Denominator, Numerator int64
	}

	TransactionPool struct {
		TransactionSizeLimit    uint   `json:"transactionSizeLimit" yaml:"transactionSizeLimit" toml:"transactionSizeLimit"`
		TransactionSetSizeLimit uint   `json:"transactionSetSizeLimit" yaml:"transactionSetSizeLimit" toml:"transactionSetSizeLimit"`
		PoolSizeLimit           uint64 `json:"poolSizeLimit" yaml:"poolSizeLimit" toml:"poolSizeLimit"`
	}
)

var (
	// ErrUnsupportedFileType is returned in case generation of a config is requested for
	// an unsupported file type
	ErrUnsupportedFileType = errors.New("file type not supported")
	// use a single instance of Validate, it caches struct info
	validate *validator.Validate
)

func init() {
	validate = validator.New()
}

// MarshalText will marshall JSON/YAML/TOML fraction type
func (f Fraction) MarshalText() ([]byte, error) {
	buffer := bytes.NewBuffer(nil)
	buffer.WriteString(fmt.Sprintf("%d/%d", f.Denominator, f.Numerator))
	return buffer.Bytes(), nil
}

// UnmarshalText will unMarshall JSON/YAML/TOML fraction type
func (f *Fraction) UnmarshalText(text []byte) error {
	fractions := strings.Split(string(text), "/")
	var err error
	var denominator, numerator int64
	denominator, err = strconv.ParseInt(strings.Trim(fractions[0], "\""), 10, 64)
	if err != nil {
		return err
	}
	numerator, err = strconv.ParseInt(strings.Trim(fractions[1], "\""), 10, 64)
	if err != nil {
		return err
	}
	f.Numerator = numerator
	f.Denominator = denominator
	return nil
}

// LoadConfigFile loads a config file from a filepath and deserialize it into our Config struct
func LoadConfigFile(filePath string) error {
	typ := path.Ext(filePath)
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	config, err := loadConfig(typ, file)
	if err != nil {
		return err
	}
	err = validateConfig(config)
	if err != nil {
		return err
	}
	fmt.Println(config.Blockchain.Network)
	return nil
}

func validateConfig(conf *Config) error {
	// returns nil or ValidationErrors ( []FieldError )
	err := validate.Struct(conf)
	if err == nil {
		return nil
	}
	// this check is only needed when your code could produce
	// an invalid value for validation such as interface with nil
	// value most including myself do not usually have code like this.
	if _, ok := err.(*validator.InvalidValidationError); ok {
		fmt.Println(err)
		return nil
	}

	// for _, err := range err.(validator.ValidationErrors) {
	// 	switch err.Field() {
	// 	case "API":
	// 		return ErrAPIPortOutOfRange
	// 	}
	// }
	return err
}

// GenerateConfigFile generates a blockchain config file
func GenerateConfigFile(filepath string) error {
	typ := path.Ext(filepath)
	file, err := os.Create(path.Join(filepath))
	if err != nil {
		return err
	}
	defer file.Close()

	fmt.Println(filepath)
	return generateConfig(typ, file)
}

// generatConfig generates a default config, encodes it according
// to the given type, and writes it to the provided writer
func generateConfig(typ string, w io.Writer) error {
	config := buildConfigStruct()
	var enc interface {
		Encode(interface{}) error
	}
	switch typ {
	case ".yaml":
		enc = yaml.NewEncoder(w)
	case ".json":
		// properly format json file
		t := json.NewEncoder(w)
		t.SetIndent("", "    ")
		enc = t
	case ".toml":
		enc = toml.NewEncoder(w)
	default:
		return ErrUnsupportedFileType
	}
	return enc.Encode(config)
}

// generatConfig generates a default config, encodes it according
// to the given type, and writes it to the provided reader
func loadConfig(typ string, r io.Reader) (*Config, error) {
	var config Config
	var dec interface {
		Decode(interface{}) error
	}
	switch typ {
	case ".yaml":
		dec = yaml.NewDecoder(r)
	case ".json":
		dec = json.NewDecoder(r)
	case ".toml":
		_, err := toml.DecodeReader(r, &config)
		return &config, err
	default:
		return nil, ErrUnsupportedFileType
	}
	err := dec.Decode(&config)
	return &config, err
}

// buildConfigStruct builds to default values in our config struct
func buildConfigStruct() *Config {
	networks := make(map[string]Network)
	networks["testnet"] = Network{
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
			Minting: &GenesisMinting{
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
		ArbitraryDataSizeLimit: 83,
		BlockCreatorFee:        1.0,
		MinimumTransactionFee:  0.1,
		BlockFrequency:         120,
		MaturityDelay:          144,
		MedianTimestampWindow:  11,
		TargetWindow:           1e3,
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
	}

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
				Precision: 9,
			},
			Ports: Ports{
				API: 23112,
				RPC: 23111,
			},
			Binaries: &Binaries{
				Client: "rivinec",
				Deamon: "rivined",
			},
			Transactions: Transactions{
				Default: Version{
					Version: 1,
				},
				Minting: &Minting{
					ConditionUpdate: Version{
						Version: 128,
					},
					CoinCreation: Version{
						Version: 129,
					},
					CoinDestruction: &Version{
						Version: 130,
					},
					RequireMinerFees: false,
				},
			},
			Network: networks,
		},
	}
}
