package client

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/rivine/rivine/build"
	"github.com/rivine/rivine/modules"
	"github.com/rivine/rivine/types"
	"github.com/spf13/cobra"
)

// exit codes
// inspired by sysexits.h
const (
	ExitCodeGeneral        = 1 // Not in sysexits.h, but is standard practice.
	ExitCodeNotFound       = 2
	ExitCodeCancelled      = 3
	ExitCodeForbidden      = 4
	ExitCodeTemporaryError = 5
	ExitCodeUsage          = 64 // EX_USAGE in sysexits.h
)

// ConfigFromDaemonConstants returns CLI constants using
// a daemon's constants as input.
func ConfigFromDaemonConstants(constants modules.DaemonConstants) Config {
	return Config{
		ChainName:    constants.ChainInfo.Name,
		NetworkName:  constants.ChainInfo.NetworkName,
		ChainVersion: constants.ChainInfo.ChainVersion,
		CurrencyUnits: types.CurrencyUnits{
			OneCoin: constants.OneCoin,
		},
		CurrencyCoinUnit:          constants.ChainInfo.CoinUnit,
		MinimumTransactionFee:     constants.MinimumTransactionFee,
		DefaultTransactionVersion: constants.DefaultTransactionVersion,
		BlockFrequencyInSeconds:   int64(constants.BlockFrequency),
		GenesisBlockTimestamp:     constants.GenesisTimestamp,
	}
}

// Config defines the configuration for the default (CLI) client.
type Config struct {
	ChainName    string
	NetworkName  string
	ChainVersion build.ProtocolVersion

	CurrencyUnits             types.CurrencyUnits
	CurrencyCoinUnit          string
	MinimumTransactionFee     types.Currency
	DefaultTransactionVersion types.TransactionVersion

	// These values aren't used for validation,
	// but only in order to estimate progress with the syncing of your consensus.
	BlockFrequencyInSeconds int64
	GenesisBlockTimestamp   types.Timestamp
}

// Wrap wraps a generic command with a check that the command has been
// passed the correct number of arguments. The command must take only strings
// as arguments.
func Wrap(fn interface{}) func(*cobra.Command, []string) {
	fnVal, fnType := reflect.ValueOf(fn), reflect.TypeOf(fn)
	if fnType.Kind() != reflect.Func {
		panic("wrapped function has wrong type signature")
	}
	for i := 0; i < fnType.NumIn(); i++ {
		if fnType.In(i).Kind() != reflect.String {
			panic("wrapped function has wrong type signature")
		}
	}

	return func(cmd *cobra.Command, args []string) {
		if len(args) != fnType.NumIn() {
			cmd.UsageFunc()(cmd)
			os.Exit(ExitCodeUsage)
		}
		argVals := make([]reflect.Value, fnType.NumIn())
		for i := range args {
			argVals[i] = reflect.ValueOf(args[i])
		}
		fnVal.Call(argVals)
	}
}

// WrapWithConfig wraps a generic command with a check that the command has been
// passed the correct number of arguments, including one extra *Config argument at the front.
// The command must take only strings as arguments, starting from the second position.
func WrapWithConfig(config *Config, fn interface{}) func(*cobra.Command, []string) {
	fnVal, fnType := reflect.ValueOf(fn), reflect.TypeOf(fn)
	if fnType.Kind() != reflect.Func {
		panic("wrapped function has wrong type signature")
	}
	numIn := fnType.NumIn()
	if numIn < 1 {
		panic("wrapped function has insufficient amount of arguments")
	}
	if fnType.In(0).Elem() != reflect.TypeOf(config) {
		panic("wrapped function should have a *Config param as first argument")
	}
	for i := 1; i < numIn; i++ {
		if fnType.In(i).Kind() != reflect.String {
			panic("wrapped function has wrong type signature")
		}
	}

	return func(cmd *cobra.Command, args []string) {
		if len(args) != numIn+1 {
			cmd.UsageFunc()(cmd)
			os.Exit(ExitCodeUsage)
		}
		argVals := make([]reflect.Value, numIn)
		argVals[0] = reflect.ValueOf(config)
		for i := 1; i < numIn; i++ {
			argVals[i] = reflect.ValueOf(args[i])
		}
		fnVal.Call(argVals)
	}
}

// Die prints its arguments to stderr, then exits the program with the default
// error code.
func Die(args ...interface{}) {
	DieWithExitCode(ExitCodeGeneral, args...)
}

// ErrorWithStatusCode couples an exit status code to an error (message)
type ErrorWithStatusCode struct {
	Err    error
	Status int
}

func (ewsc ErrorWithStatusCode) Error() string {
	return ewsc.Err.Error()
}

// DieWithError exits with an error.
// if the error is of the type ErrorWithStatusCode, its status will be used,
// otherwise the General exit code will be used
func DieWithError(description string, err error) {
	if ewsc, ok := err.(ErrorWithStatusCode); ok {
		DieWithExitCode(ewsc.Status, description, ewsc.Err)
		return
	}
	DieWithExitCode(ExitCodeGeneral, description, err)
}

// DieWithExitCode prints its arguments to stderr,
// then exits the program with the given exit code.
func DieWithExitCode(code int, args ...interface{}) {
	fmt.Fprintln(os.Stderr, args...)
	os.Exit(code)
}

// NewCommandLineClient creates a new CLI client, which can be run as it is,
// or be extended/modified to fit your needs.
// If a config is not loaded automatically, it will be tried to be loaded from the daemon
// automatically, this might however fail.
func NewCommandLineClient(address, name string) (*CommandLineClient, error) {
	if address == "" {
		address = "http://localhost:23110"
	}
	if name == "" {
		name = "R?v?ne"
	}
	cli := new(CommandLineClient)
	cli.HTTPClient = &HTTPClient{
		RootURL: address,
	}

	var consensusCmd *consensusCmd
	consensusCmd, cli.ConsensusCmd = createConsensusCmd(cli)
	cli.RootCmd = &cobra.Command{
		Use:               os.Args[0],
		Short:             fmt.Sprintf("%s Client", strings.Title(name)),
		Long:              fmt.Sprintf("%s Client", strings.Title(name)),
		Run:               Wrap(consensusCmd.rootCmd),
		PersistentPreRunE: cli.preRunE,
	}

	// create command tree
	cli.RootCmd.AddCommand(cli.ConsensusCmd)
	cli.RootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Long:  "Print version information.",
		Run: Wrap(func() {
			fmt.Printf("%s Client v%s\r\n",
				strings.Title(cli.Config.ChainName),
				cli.Config.ChainVersion.String())
		}),
	})
	cli.RootCmd.AddCommand(&cobra.Command{
		Use:   "stop",
		Short: fmt.Sprintf("Stop the %s daemon", name),
		Long:  fmt.Sprintf("Stop the %s daemon.", name),
		Run: Wrap(func() {
			err := cli.Post("/daemon/stop", "")
			if err != nil {
				Die("Could not stop daemon:", err)
			}
			fmt.Printf("%s daemon stopped.\n", cli.Config.ChainName)
		}),
	})

	cli.WalletCmd = createWalletCmd(cli)
	cli.RootCmd.AddCommand(cli.WalletCmd.Command)

	cli.AtomicSwapCmd = createAtomicSwapCmd(cli)
	cli.RootCmd.AddCommand(cli.AtomicSwapCmd)

	cli.GatewayCmd = createGatewayCmd(cli)
	cli.RootCmd.AddCommand(cli.GatewayCmd)

	cli.ExploreCmd = createExploreCmd(cli)
	cli.RootCmd.AddCommand(cli.ExploreCmd)

	cli.MergeCmd = createMergeCmd(cli)
	cli.RootCmd.AddCommand(cli.MergeCmd)

	// parse flags
	cli.RootCmd.PersistentFlags().StringVarP(&cli.HTTPClient.RootURL, "addr", "a",
		cli.HTTPClient.RootURL, fmt.Sprintf(
			"which host/port to communicate with (i.e. the host/port %sd is listening on)",
			name))

	// return cli
	return cli, nil
}

// CommandLineClient represents the Rivine Reference CLI Client implementation, which can be run as it is,
// or be extended/modified to fit your needs.
//
// If a config is not loaded automatically, it will be tried to be loaded from the daemon
// automatically, this might however fail, in which case a DefaultConfigFunc can be registered as callback to fallback to.
type CommandLineClient struct {
	*HTTPClient

	Config *Config

	PreRunE func(*Config) (*Config, error)

	RootCmd       *cobra.Command
	WalletCmd     *WalletCommand
	ConsensusCmd  *cobra.Command
	AtomicSwapCmd *cobra.Command
	GatewayCmd    *cobra.Command
	ExploreCmd    *cobra.Command
	MergeCmd      *cobra.Command
}

// preRunE checks that all preConditions match
func (cli *CommandLineClient) preRunE(*cobra.Command, []string) error {
	address, err := sanitizeURL(cli.HTTPClient.RootURL)
	if err != nil {
		return fmt.Errorf("invalid daemon RPC address %q: %v", cli.HTTPClient.RootURL, err)
	}
	cli.HTTPClient.RootURL = address

	if cli.Config == nil {
		var err error
		cli.Config, err = FetchConfigFromDaemon(cli.HTTPClient)
		if err != nil {
			fmt.Fprintf(os.Stderr, "etching config from daemon failed: %v", err)
		}
	}
	if cli.PreRunE != nil {
		cli.Config, err = cli.PreRunE(cli.Config)
		if err != nil {
			return fmt.Errorf("user-defined pre-run callback failed: %v", err)
		}
	}
	if cli.Config == nil {
		return errors.New("cannot run command line client: no config is defined")
	}
	return nil
}

// Run the CLI, logic dependend upon the command the user used.
func (cli *CommandLineClient) Run() error {
	return cli.RootCmd.Execute()
}

// CreateCurrencyConvertor creates a currency convertor using the internally stored Config.
func (cli *CommandLineClient) CreateCurrencyConvertor() CurrencyConvertor {
	return NewCurrencyConvertor(
		cli.Config.CurrencyUnits,
		cli.Config.CurrencyCoinUnit,
	)
}

// FetchConfigFromDaemon fetches constants and creates a config, by fetching the constants from the daemon.
// Returns an error in case the fetching wasn't possible.
func FetchConfigFromDaemon(httpClient *HTTPClient) (*Config, error) {
	var constants modules.DaemonConstants
	err := httpClient.GetAPI("/daemon/constants", &constants)
	if err == nil {
		// returned config from received constants from the daemon's server module
		cfg := ConfigFromDaemonConstants(constants)
		return &cfg, nil
	}
	err = httpClient.GetAPI("/explorer/constants", &constants)
	if err != nil {
		return nil, fmt.Errorf("failed to load constants from daemon's server and explorer modules: %v", err)
	}
	if constants.ChainInfo == (types.BlockchainInfo{}) {
		// only since 1.0.7 do we support the full set of public daemon constants for both
		// the explorer endpoint as well as the daemon endpoint,
		// so we need to validate this
		return nil, errors.New("failed to load constants from daemon's server and explorer modules: " +
			"explorer modules does not support the full exposure of public daemon constants")
	}
	// returned config from received constants from the daemon's explorer module
	cfg := ConfigFromDaemonConstants(constants)
	return &cfg, nil
}
