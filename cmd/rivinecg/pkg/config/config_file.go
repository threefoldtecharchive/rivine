package config

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"gopkg.in/go-playground/validator.v9"
	"gopkg.in/yaml.v2"

	"github.com/threefoldtech/rivine/modules"
	"github.com/threefoldtech/rivine/types"
)

type NetworkType int

const (
	Standard NetworkType = iota + 1
	Testnet
	Devnet
)

type (
	Config struct {
		Template   Template    `json:"template,omitempty" yaml:"template,omitempty"`
		Blockchain *Blockchain `json:"blockchain" yaml:"blockchain" validate:"required"`
	}

	Template struct {
		Repository *Repository `json:"repository,omitempty" yaml:"repository,omitempty"`
		Version    string      `json:"version,omitempty" yaml:"version,omitempty"`
	}

	Repository struct {
		Owner string `json:"owner,omitempty" yaml:"owner,omitempty"`
		Repo  string `json:"repo,omitempty" yaml:"repo,omitempty"`
	}

	Blockchain struct {
		Name         string              `json:"name" yaml:"name" validate:"required"`
		Repository   string              `json:"repository" yaml:"repository" validate:"required"`
		Currency     *Currency           `json:"currency" yaml:"currency" validate:"required"`
		Ports        *Ports              `json:"ports" yaml:"ports" validate:"required"`
		Binaries     *Binaries           `json:"binaries,omitempty" yaml:"binaries,omitempty"`
		Transactions *Transactions       `json:"transactions,omitempty" yaml:"transactions,omitempty"`
		Network      map[string]*Network `json:"network,omitempty" yaml:"network,omitempty" validate:"gt=0,dive,required"`
	}

	Currency struct {
		Unit      string `json:"unit" yaml:"unit" validate:"required"`
		Precision uint64 `json:"precision" yaml:"precision" validate:"required"`
	}

	Ports struct {
		API uint16 `json:"api,omitempty" yaml:"api,omitempty" validate:"nefield=RPC"`
		RPC uint16 `json:"rpc,omitempty" yaml:"rpc,omitempty" validate:"nefield=API"`
	}

	Binaries struct {
		Client string `json:"client,omitempty" yaml:"client,omitempty"`
		Daemon string `json:"daemon,omitempty" yaml:"daemon,omitempty"`
	}

	Transactions struct {
		Default *Version `json:"default,omitempty" yaml:"default,omitempty"`
		Minting *Minting `json:"minting,omitempty" yaml:"minting,omitempty"`
	}

	Minting struct {
		ConditionUpdate  Version  `json:"conditionUpdate" yaml:"conditionUpdate" validate:"required"`
		CoinCreation     Version  `json:"coinCreation" yaml:"coinCreation" validate:"required"`
		CoinDestruction  *Version `json:"coinDestruction,omitempty" yaml:"coinDestruction,omitempty"`
		RequireMinerFees bool     `json:"requireMinerFees,omitempty" yaml:"requireMinerFees,omitempty"`
	}

	Version struct {
		Version uint64 `json:"version" yaml:"version" validate:"required"`
	}

	Genesis struct {
		CoinOutputs       []Output   `json:"coinOuput" yaml:"coinOutput" validate:"required"`
		BlockStakeOutputs []Output   `json:"blockStakeOutputs" yaml:"blockStakeOutputs" validate:"required"`
		Minting           *Condition `json:"minting,omitempty" yaml:"minting,omitempty" validate:"required_with=blockchain.transactions.minting"`
	}

	Output struct {
		Value     CurrencyValue `json:"value" yaml:"value" validate:"required"`
		Condition Condition     `json:"condition" yaml:"condition" validate:"required"`
	}

	Condition struct {
		types.UnlockConditionProxy
	}

	multiSignatureCondition struct {
		Addresses          []UnlockHash `json:"addresses" yaml:"addresses"`
		SignaturesRequired uint64       `json:"signaturesRequired,omitempty" yaml:"signaturesRequired,omitempty"`
	}

	UnlockHash struct {
		types.UnlockHash
	}

	CurrencyValue struct {
		types.Currency
	}

	Network struct {
		NetworkType            NetworkType     `json:"networkType" yaml:"networkType" validate:"required"`
		Genesis                *Genesis        `json:"genesis" yaml:"genesis" validate:"required"`
		TransactionFeePool     string          `json:"transactionFeePool,omitempty" yaml:"transactionFeePool,omitempty"`
		BlockSizeLimit         uint64          `json:"blockSizeLimit,omitempty" yaml:"blockSizeLimit,omitempty"`
		ArbitraryDataSizeLimit uint64          `json:"arbitraryDataSizeLimit,omitempty" yaml:"arbitraryDataSizeLimit,omitempty"`
		BlockCreatorFee        string          `json:"blockCreatorFee,omitempty" yaml:"blockCreatorFee,omitempty"`
		MinimumTransactionFee  string          `json:"minimumTransactionFee,omitempty" yaml:"minimumTransactionFee,omitempty"`
		BlockFrequency         uint64          `json:"blockFrequency,omitempty" yaml:"blockFrequency,omitempty"`
		MaturityDelay          uint64          `json:"maturityDelay,omitempty" yaml:"maturityDelay,omitempty"`
		MedianTimestampWindow  uint64          `json:"medianTimestampWindow,omitempty" yaml:"medianTimestampWindow,omitempty"`
		TargetWindow           uint64          `json:"targetWindow,omitempty" yaml:"targetWindow,omitempty"`
		MaxAdjustmentUp        Fraction        `json:"maxAdjustmentUp,omitempty" yaml:"maxAdjustmentUp,omitempty"`
		MaxAdjustmentDown      Fraction        `json:"maxAdjustmentDown,omitempty" yaml:"maxAdjustmentDown,omitempty"`
		FutureThreshold        time.Duration   `json:"futureTreshold,omitempty" yaml:"futureTreshold,omitempty"`
		ExtremeFutureThreshold time.Duration   `json:"extremeFutureTreshold,omitempty" yaml:"extremeFutureTreshold,omitempty"`
		StakeModifierDelay     time.Duration   `json:"stakeModifierDelay,omitempty" yaml:"stakeModifierDelay,omitempty"`
		BlockStakeAging        time.Duration   `json:"blockStakeAging,omitempty" yaml:"blockStakeAging,omitempty"`
		TransactionPool        TransactionPool `json:"transactionPool,omitempty" yaml:"transactionPool,omitempty"`
		BootstapPeers          []*NetAddress   `json:"bootstrapPeers" yaml:"bootstrapPeers" validate:"required"`
	}

	NetAddress struct {
		modules.NetAddress `json:"bootstrapPeer" yaml:"bootstrapPeer" validate:"InvalidAddress"`
	}

	// Fraction represents ratio.
	Fraction struct {
		Denominator, Numerator int64
	}

	TransactionPool struct {
		TransactionSizeLimit    uint   `json:"transactionSizeLimit" yaml:"transactionSizeLimit"`
		TransactionSetSizeLimit uint   `json:"transactionSetSizeLimit" yaml:"transactionSetSizeLimit"`
		PoolSizeLimit           uint64 `json:"poolSizeLimit" yaml:"poolSizeLimit"`
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
	validate.RegisterValidation("InvalidAddress", customNetAddressValidation)
}

func customNetAddressValidation(fl validator.FieldLevel) bool {
	netAddress := modules.NetAddress(fl.Field().String())
	err := netAddress.IsValid()
	if err != nil {
		return false
	}

	return true
}

// MarshalText will marshall JSON/YAML fraction type
func (f Fraction) MarshalText() ([]byte, error) {
	buffer := bytes.NewBuffer(nil)
	buffer.WriteString(fmt.Sprintf("%d/%d", f.Denominator, f.Numerator))
	return buffer.Bytes(), nil
}

// UnmarshalText will unMarshall JSON/YAML fraction type
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

// MarshalText will marshall JSON/YAML CurrencyValue type
func (c CurrencyValue) MarshalText() ([]byte, error) {
	return []byte(c.String()), nil
}

// UnmarshalText will unMarshall JSON/YAML CurrencyValue type
func (c *CurrencyValue) UnmarshalText(text []byte) error {
	str := string(text)
	return c.LoadString(str)
}

// MarshalText will marshall JSON/YAML UnlockHash type
func (u UnlockHash) MarshalText() ([]byte, error) {
	return []byte(u.String()), nil
}

// UnmarshalText will unMarshall JSON/YAML UnlockHash type
func (u *UnlockHash) UnmarshalText(text []byte) error {
	str := string(text)
	err := u.LoadString(strings.TrimSpace(str))
	if err != nil {
		return err
	}
	return nil
}

func (msc multiSignatureCondition) AsCondition() (types.UnlockConditionProxy, error) {
	if len(msc.Addresses) == 0 {
		return types.UnlockConditionProxy{}, errors.New("MultiSig outputs must specify at least a single address which can sign it as an input")
	}
	if msc.SignaturesRequired == 0 {
		msc.SignaturesRequired = uint64(len(msc.Addresses))
	} else {
		if msc.SignaturesRequired == 0 {
			return types.UnlockConditionProxy{}, errors.New("MultiSig outputs must require at least a single signature to unlock")
		}
		if uint64(len(msc.Addresses)) < msc.SignaturesRequired {
			return types.UnlockConditionProxy{}, errors.New("You can't create a multisig which requires more signatures to spent then there are addresses which can sign")
		}
	}
	return types.NewCondition(types.NewMultiSignatureCondition(convertToRivineUnlockHashSlice(msc.Addresses), msc.SignaturesRequired)), nil
}

func NewCondition(mc types.MarshalableUnlockCondition) Condition {
	return Condition{
		UnlockConditionProxy: types.NewCondition(mc),
	}
}

func NewMultisigCondition(minSignatureCount uint64, uhs ...types.UnlockHash) Condition {
	return Condition{
		UnlockConditionProxy: types.NewCondition(types.NewMultiSignatureCondition(uhs, minSignatureCount)),
	}
}

func convertFromRivineUnlockHashSlice(slice types.UnlockHashSlice) []UnlockHash {
	uhs := make([]UnlockHash, len(slice))
	for i, uh := range slice {
		uhs[i] = UnlockHash{uh}
	}
	return uhs
}

func convertToRivineUnlockHashSlice(slice []UnlockHash) types.UnlockHashSlice {
	uhs := make(types.UnlockHashSlice, len(slice))
	for i, uh := range slice {
		uhs[i] = uh.UnlockHash
	}
	return uhs
}

func (c Condition) MarshalJSON() ([]byte, error) {
	ct := c.ConditionType()
	if ct == types.ConditionTypeUnlockHash {
		return json.Marshal(c.UnlockHash().String())
	}
	if ct == types.ConditionTypeMultiSignature {
		msc := c.Condition.(*types.MultiSignatureCondition)
		return json.Marshal(multiSignatureCondition{
			Addresses:          convertFromRivineUnlockHashSlice(msc.UnlockHashes),
			SignaturesRequired: msc.MinimumSignatureCount,
		})
	}
	return nil, fmt.Errorf("cannot marshal unsupported condition of type %d", ct)
}

func (c *Condition) UnmarshalJSON(data []byte) error {
	var str string
	err := json.Unmarshal(data, &str)
	if err == nil {
		// assume it is a plain unlock hash (=address)
		var uh types.UnlockHash
		err := uh.LoadString(str)
		if err != nil {
			return err
		}
		c.UnlockConditionProxy = types.NewCondition(types.NewUnlockHashCondition(uh))
		return nil
	}

	// assume it is a multi signature condition
	var msc multiSignatureCondition
	err = json.Unmarshal(data, &msc)
	if err != nil {
		return err
	}
	c.UnlockConditionProxy, err = msc.AsCondition()
	return err
}

// MarshalYAML will marshal the condition type into our specific format
func (c Condition) MarshalYAML() (interface{}, error) {
	ct := c.ConditionType()
	if ct == types.ConditionTypeUnlockHash {
		uh := c.UnlockHash().String()
		return string(uh), nil
	}
	if ct == types.ConditionTypeMultiSignature {
		msc := c.Condition.(*types.MultiSignatureCondition)
		return multiSignatureCondition{
			Addresses:          convertFromRivineUnlockHashSlice(msc.UnlockHashes),
			SignaturesRequired: msc.MinimumSignatureCount,
		}, nil
	}
	return nil, fmt.Errorf("cannot marshal unsupported condition of type %d", ct)
}

func (c *Condition) UnmarshalYAML(unmarshal func(v interface{}) error) error {
	var str string
	err := unmarshal(&str)
	if err == nil {
		// assume it is a plain unlock hash (=address)
		var uh types.UnlockHash
		err := uh.LoadString(strings.TrimSpace(str))
		if err != nil {
			return err
		}
		c.UnlockConditionProxy = types.NewCondition(types.NewUnlockHashCondition(uh))
		return nil
	}

	// assume it is a multi signature condition
	var msc multiSignatureCondition
	err = unmarshal(&msc)
	if err != nil {
		return err
	}
	c.UnlockConditionProxy, err = msc.AsCondition()
	return err
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

	// Validate if config provided is formatted correctly
	err = validateConfig(config)
	if err != nil {
		return err
	}

	config, err = assignDefaultValues(config)
	if err != nil {
		return err
	}

	// Get the absolute filepath of the config to know where to generate blockchain code
	relativeFilePath, err := filepath.Abs(filePath)
	if err != nil {
		return err
	}
	relativeCodeDestinationDir := filepath.Dir(relativeFilePath) + "/" + config.Template.Repository.Owner + "-" + config.Template.Repository.Repo

	commitHash, err := getTemplateRepo(config.Template.Repository.Owner, config.Template.Repository.Repo, config.Template.Version, filepath.Dir(relativeFilePath))
	if err != nil {
		return err
	}

	err = generateBlockchainTemplate(relativeCodeDestinationDir, commitHash, config)
	if err != nil {
		return err
	}
	fmt.Printf("\nBlockchain code succesfully generated: %s\n", relativeCodeDestinationDir)
	return nil
}

func assignDefaultValues(conf *Config) (*Config, error) {
	// Fill in default values for provided network properties
	for _, network := range conf.Blockchain.Network {
		network = assignDefaultNetworkProps(network)
	}

	// Fill in default values for optional values in provided config
	conf.Template = assignDefaultTemplateValues(conf.Template)
	conf.Blockchain = assignDefaultBlockchainValues(conf.Blockchain)

	// Validate against our new config
	err := validateConfig(conf)
	if err != nil {
		return nil, err
	}
	return conf, nil
}

func assignDefaultTemplateValues(templ Template) Template {
	if templ.Repository == nil {
		templ.Repository = &Repository{
			Owner: "threefoldtech",
			Repo:  "rivine-chain-template",
		}
	}
	if templ.Repository.Owner == "" {
		templ.Repository.Owner = "threefoldtech"
	}
	if templ.Repository.Repo == "" {
		templ.Repository.Repo = "rivine-chain-template"
	}
	if templ.Version == "" {
		templ.Version = "master"
	}
	return templ
}

func assignDefaultBlockchainValues(blockc *Blockchain) *Blockchain {
	if blockc.Binaries == nil {
		blockc.Binaries = &Binaries{
			Client: blockc.Name + "c",
			Daemon: blockc.Name + "d",
		}
	}
	if blockc.Binaries.Client == "" {
		blockc.Binaries.Client = blockc.Name + "c"
	}
	if blockc.Binaries.Daemon == "" {
		blockc.Binaries.Daemon = blockc.Name + "d"
	}
	if blockc.Transactions.Default == nil {
		blockc.Transactions.Default = &Version{
			Version: 1,
		}
	}
	if blockc.Transactions.Default.Version == 0 {
		blockc.Transactions.Default.Version = 1
	}
	return blockc
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
	config := BuildConfigStruct()
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
	default:
		return nil, ErrUnsupportedFileType
	}
	err := dec.Decode(&config)
	return &config, err
}

// BuildConfigStruct builds to default values in our config struct
func BuildConfigStruct() *Config {
	uhs := func(str string) (uh types.UnlockHash) {
		err := uh.LoadString(str)
		if err != nil {
			panic(err)
		}
		return
	}
	uhsc := func(str string) Condition {
		return NewCondition(types.NewUnlockHashCondition(uhs(str)))
	}

	mintCondition := uhsc("01434535fd01243c02c277cd58d71423163767a575a8ae44e15807bf545e4a8456a5c4afabad51")

	networks := make(map[string]*Network)
	networks["testnet"] = &Network{
		NetworkType: Testnet,
		Genesis: &Genesis{
			CoinOutputs: []Output{
				{
					Value: CurrencyValue{
						types.NewCurrency64(5e6 * 1e9),
					},
					Condition: uhsc("01434535fd01243c02c277cd58d71423163767a575a8ae44e15807bf545e4a8456a5c4afabad51"),
				},
				{
					Value: CurrencyValue{
						types.NewCurrency64(5e6 * 1e9),
					},
					Condition: NewMultisigCondition(
						2,
						uhs("01434535fd01243c02c277cd58d71423163767a575a8ae44e15807bf545e4a8456a5c4afabad51"),
						uhs("01434535fd01243c02c277cd58d71423163767a575a8ae44e15807bf545e4a8456a5c4afabad51"),
					),
					// Condition: uhs("01b5e42056ef394f2ad9b511a61cec874d25bebe2095682dd37455cbafed4bec154e382a23f90e"),
				},
			},
			BlockStakeOutputs: []Output{
				{
					Value: CurrencyValue{
						types.NewCurrency64(3000),
					},
					Condition: uhsc("01b5e42056ef394f2ad9b511a61cec874d25bebe2095682dd37455cbafed4bec154e382a23f90e"),
				},
			},
			Minting: &mintCondition,
		},
		TransactionFeePool:     "017267221ef1947bb18506e390f1f9446b995acfb6d08d8e39508bb974d9830b8cb8fdca788e34",
		BlockSizeLimit:         uint64(2e6),
		ArbitraryDataSizeLimit: 83,
		BlockCreatorFee:        "1.0",
		MinimumTransactionFee:  "0.1",
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
		BootstapPeers: []*NetAddress{
			&NetAddress{"bootstrap1.testnet.threefoldtoken.com:23112"},
			&NetAddress{"bootstrap2.testnet.threefoldtoken.com:23112"},
			&NetAddress{"bootstrap3.testnet.threefoldtoken.com:23112"},
			&NetAddress{"bootstrap4.testnet.threefoldtoken.com:24112"},
			&NetAddress{"bootstrap5.testnet.threefoldtoken.com:24112"},
		},
	}

	return &Config{
		Template{
			Repository: &Repository{
				Owner: "threefoldtech",
				Repo:  "rivine-chain-template",
			},
			Version: "master",
		},
		&Blockchain{
			Name:       "rivine",
			Repository: "github.com/threefoldtech/rivine",
			Currency: &Currency{
				Unit:      "ROC",
				Precision: 9,
			},
			Ports: &Ports{
				API: 23112,
				RPC: 23111,
			},
			Binaries: &Binaries{
				Client: "rivinec",
				Daemon: "rivined",
			},
			Transactions: &Transactions{
				Default: &Version{
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
