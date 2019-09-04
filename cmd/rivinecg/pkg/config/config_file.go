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

	validator "gopkg.in/go-playground/validator.v9"
	yaml "gopkg.in/yaml.v2"

	"github.com/threefoldtech/rivine/modules"
	"github.com/threefoldtech/rivine/types"
)

// NetworkType is the type of the network
type NetworkType int

// NetworkType start with 1 because 0 breaks tests and serialization
const (
	NetworkTypeStandard NetworkType = iota + 1
	NetworkTypeTestnet
	NetworkTypeDevnet
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
		Default  *Version  `json:"default,omitempty" yaml:"default,omitempty"`
		Minting  *Minting  `json:"minting,omitempty" yaml:"minting,omitempty"`
		Authcoin *Authcoin `json:"authcoin,omitempty" yaml:"authcoin,omitempty"`
	}

	Minting struct {
		ConditionUpdate  Version  `json:"conditionUpdate" yaml:"conditionUpdate" validate:"required"`
		CoinCreation     Version  `json:"coinCreation" yaml:"coinCreation" validate:"required"`
		CoinDestruction  *Version `json:"coinDestruction,omitempty" yaml:"coinDestruction,omitempty"`
		RequireMinerFees bool     `json:"requireMinerFees,omitempty" yaml:"requireMinerFees,omitempty"`
	}

	Authcoin struct {
		TransactionVersionAuthAddressUpdate   Version `json:"transactionVersionAuthAddressUpdate" yaml:"transactionVersionAuthAddressUpdate" validate:"required"`
		TransactionVersionAuthConditionUpdate Version `json:"transactionVersionAuthConditionUpdate" yaml:"transactionVersionAuthConditionUpdate" validate:"required"`
	}

	Version struct {
		Version uint64 `json:"version" yaml:"version" validate:"required"`
	}

	Genesis struct {
		CoinOutputs           []Output   `json:"coinOuputs" yaml:"coinOutputs" validate:"required"`
		BlockStakeOutputs     []Output   `json:"blockStakeOutputs" yaml:"blockStakeOutputs" validate:"required"`
		Minting               *Condition `json:"minting,omitempty" yaml:"minting,omitempty"`
		Authcoin              *Condition `json:"authcoin,omitempty" yaml:"authcoin,omitempty"`
		GenesisBlockTimestamp int64      `json:"genesisBlockTimestamp" yaml:"genesisBlockTimestamp" validate:"required"`
	}

	Output struct {
		Value     string    `json:"value" yaml:"value" validate:"required"`
		Condition Condition `json:"condition" yaml:"condition" validate:"required"`
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

	Network struct {
		NetworkType            NetworkType      `json:"networkType" yaml:"networkType" validate:"required"`
		Genesis                *Genesis         `json:"genesis" yaml:"genesis" validate:"required"`
		TransactionFeePool     string           `json:"transactionFeePool,omitempty" yaml:"transactionFeePool,omitempty"`
		BlockSizeLimit         uint64           `json:"blockSizeLimit,omitempty" yaml:"blockSizeLimit,omitempty"`
		ArbitraryDataSizeLimit uint64           `json:"arbitraryDataSizeLimit,omitempty" yaml:"arbitraryDataSizeLimit,omitempty"`
		BlockCreatorFee        string           `json:"blockCreatorFee,omitempty" yaml:"blockCreatorFee,omitempty"`
		MinimumTransactionFee  string           `json:"minimumTransactionFee,omitempty" yaml:"minimumTransactionFee,omitempty"`
		BlockFrequency         uint64           `json:"blockFrequency,omitempty" yaml:"blockFrequency,omitempty"`
		MaturityDelay          uint64           `json:"maturityDelay,omitempty" yaml:"maturityDelay,omitempty"`
		MedianTimestampWindow  uint64           `json:"medianTimestampWindow,omitempty" yaml:"medianTimestampWindow,omitempty"`
		TargetWindow           uint64           `json:"targetWindow,omitempty" yaml:"targetWindow,omitempty"`
		MaxAdjustmentUp        Fraction         `json:"maxAdjustmentUp,omitempty" yaml:"maxAdjustmentUp,omitempty"`
		MaxAdjustmentDown      Fraction         `json:"maxAdjustmentDown,omitempty" yaml:"maxAdjustmentDown,omitempty"`
		FutureThreshold        uint64           `json:"futureTreshold,omitempty" yaml:"futureTreshold,omitempty"`
		ExtremeFutureThreshold uint64           `json:"extremeFutureTreshold,omitempty" yaml:"extremeFutureTreshold,omitempty"`
		StakeModifierDelay     uint64           `json:"stakeModifierDelay,omitempty" yaml:"stakeModifierDelay,omitempty"`
		BlockStakeAging        uint64           `json:"blockStakeAging,omitempty" yaml:"blockStakeAging,omitempty"`
		TransactionPool        TransactionPool  `json:"transactionPool,omitempty" yaml:"transactionPool,omitempty"`
		BootstrapPeers         []*BootstrapPeer `json:"bootstrapPeers" yaml:"bootstrapPeers" validate:"required"`
	}

	BootstrapPeer struct {
		modules.NetAddress
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

// MarshalText will marshall JSON/YAML BootstrapPeer type
func (bp BootstrapPeer) MarshalText() ([]byte, error) {
	return []byte(string(bp.NetAddress)), nil
}

// UnmarshalText will unMarshall JSON/YAML BootstrapPeer type
func (bp *BootstrapPeer) UnmarshalText(text []byte) error {
	bp.NetAddress = modules.NetAddress(text)
	return nil
}

// GenerateBlockchain imports a config file and uses it to generate a blockchain
func GenerateBlockchain(configFilePath, outputDir string) error {
	typ := path.Ext(configFilePath)
	file, err := os.Open(configFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	config, err := decodeConfig(typ, file)
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

	commitHash, err := getTemplateRepo(config.Template.Repository.Owner, config.Template.Repository.Repo, config.Template.Version, outputDir)
	if err != nil {
		return err
	}

	err = generateBlockchainTemplate(outputDir, commitHash, config)
	if err != nil {
		return err
	}

	printSteps(outputDir, config.Blockchain.Name)
	return nil
}

// assignDefaultValues assign sane default values to missing parameters in config
func assignDefaultValues(conf *Config) (*Config, error) {
	// Fill in default values for provided network properties
	for _, network := range conf.Blockchain.Network {
		assignDefaultNetworkProps(network)
	}

	// Fill in default values for optional values in provided config
	conf.Template = assignDefaultTemplateValues(conf.Template)
	conf.Blockchain = assignDefaultBlockchainValues(conf.Blockchain)

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
		if blockc.Transactions.Default.Version == 0 {
			blockc.Transactions.Default.Version = 1
		}
	}
	return blockc
}

func validateConfig(conf *Config) error {
	// returns nil or ValidationErrors ( []FieldError )
	err := validate.Struct(conf)
	if err != nil {
		return err
	}

	// validates if a bootstrapPeer is formatted correctly
	for _, network := range conf.Blockchain.Network {
		for _, peer := range network.BootstrapPeers {
			// allow loopback addresses in devnet network
			if peer.IsLoopback() && network.NetworkType == NetworkTypeDevnet {
				return peer.IsStdValid()
			}
			err := peer.IsValid()
			if err != nil {
				return err
			}
		}
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

	return encodeConfig(typ, file)
}

// encodeConfig generates a default config, encodes it according
// to the given type, and writes it to the provided writer
func encodeConfig(typ string, w io.Writer) error {
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

// decodeConfig decodes the provided Reader into a config struct
func decodeConfig(typ string, r io.Reader) (*Config, error) {
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
		NetworkType: NetworkTypeTestnet,
		Genesis: &Genesis{
			CoinOutputs: []Output{
				{
					Value:     "500000",
					Condition: uhsc("01434535fd01243c02c277cd58d71423163767a575a8ae44e15807bf545e4a8456a5c4afabad51"),
				},
				{
					Value: "500000",
					Condition: NewMultisigCondition(
						2,
						uhs("01434535fd01243c02c277cd58d71423163767a575a8ae44e15807bf545e4a8456a5c4afabad51"),
						uhs("01434535fd01243c02c277cd58d71423163767a575a8ae44e15807bf545e4a8456a5c4afabad51"),
					),
				},
			},
			BlockStakeOutputs: []Output{
				{
					Value:     "3000",
					Condition: uhsc("01434535fd01243c02c277cd58d71423163767a575a8ae44e15807bf545e4a8456a5c4afabad51"),
				},
			},
			Minting:               &mintCondition,
			GenesisBlockTimestamp: 1524168391,
		},
		TransactionFeePool: "01434535fd01243c02c277cd58d71423163767a575a8ae44e15807bf545e4a8456a5c4afabad51",
		BootstrapPeers: []*BootstrapPeer{
			&BootstrapPeer{"bootstrap1.testnet.example.com:23112"},
			&BootstrapPeer{"bootstrap2.testnet.example.com:23112"},
			&BootstrapPeer{"bootstrap3.testnet.example.com:23112"},
			&BootstrapPeer{"bootstrap4.testnet.example.com:23112"},
			&BootstrapPeer{"bootstrap5.testnet.example.com:23112"},
		},
	}

	networks["standard"] = &Network{
		NetworkType: NetworkTypeStandard,
		Genesis: &Genesis{
			CoinOutputs: []Output{
				{
					Value:     "500000",
					Condition: uhsc("01b5e42056ef394f2ad9b511a61cec874d25bebe2095682dd37455cbafed4bec154e382a23f90e"),
				},
			},
			BlockStakeOutputs: []Output{
				{
					Value:     "3000",
					Condition: uhsc("01b5e42056ef394f2ad9b511a61cec874d25bebe2095682dd37455cbafed4bec154e382a23f90e"),
				},
			},
			Minting:               &mintCondition,
			GenesisBlockTimestamp: 1524168391,
		},
		TransactionFeePool: "017267221ef1947bb18506e390f1f9446b995acfb6d08d8e39508bb974d9830b8cb8fdca788e34",
		BootstrapPeers: []*BootstrapPeer{
			&BootstrapPeer{"bootstrap1.example.com:23112"},
			&BootstrapPeer{"bootstrap2.example.com:23112"},
			&BootstrapPeer{"bootstrap3.example.com:23112"},
			&BootstrapPeer{"bootstrap4.example.com:23112"},
			&BootstrapPeer{"bootstrap5.example.com:23112"},
		},
	}

	networks["devnet"] = &Network{
		NetworkType: NetworkTypeDevnet,
		Genesis: &Genesis{
			CoinOutputs: []Output{
				{
					Value:     "500000",
					Condition: uhsc("015a080a9259b9d4aaa550e2156f49b1a79a64c7ea463d810d4493e8242e6791584fbdac553e6f"),
				},
			},
			BlockStakeOutputs: []Output{
				{
					Value:     "3000",
					Condition: uhsc("015a080a9259b9d4aaa550e2156f49b1a79a64c7ea463d810d4493e8242e6791584fbdac553e6f"),
				},
			},
			Minting:               &mintCondition,
			GenesisBlockTimestamp: 1524168391,
		},
		TransactionFeePool: "015a080a9259b9d4aaa550e2156f49b1a79a64c7ea463d810d4493e8242e6791584fbdac553e6f",
		BootstrapPeers: []*BootstrapPeer{
			&BootstrapPeer{"localhost:23112"},
		},
	}

	for _, cfg := range networks {
		assignDefaultNetworkProps(cfg)
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
			Name:       "bctempl",
			Repository: "github.com/somebody/bctempl",
			Currency: &Currency{
				Unit:      "ROC",
				Precision: 9,
			},
			Ports: &Ports{
				API: 23111,
				RPC: 23112,
			},
			Binaries: &Binaries{
				Client: "bctemplc",
				Daemon: "bctempld",
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

func printSteps(destinationDir, blockchainName string) {
	fmt.Printf("Your blockchain code is now available for usage. In order to publish it to GitHub follow these steps: \n\n")
	fmt.Printf("1. Change directory into: %s \n", destinationDir)
	fmt.Println("2. Create a repository on github.com")
	fmt.Println("3. Folow steps on github.com to upload this code to your github repository")
	fmt.Println(`4. Create a tag for this repository example: git tag -a v0.1 -m "my version 0.1" `)
	fmt.Println("5. Push your tags: git push --tags")
	fmt.Println("6. Fetch dependencies for your repository: dep ensure (in root of project)")
	fmt.Println("7. Create the binaries: make install-std")
	fmt.Printf("8. Launch your blockchain localy: %s \n", blockchainName)
}
