package client

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
	"github.com/threefoldtech/rivine/build"
	"github.com/threefoldtech/rivine/modules"
	"github.com/threefoldtech/rivine/pkg/api"
	"github.com/threefoldtech/rivine/pkg/cli"
	"github.com/threefoldtech/rivine/types"
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
		build.Critical("wrapped function has wrong type signature")
	}
	for i := 0; i < fnType.NumIn(); i++ {
		if fnType.In(i).Kind() != reflect.String {
			build.Critical("wrapped function has wrong type signature")
		}
	}

	return func(cmd *cobra.Command, args []string) {
		if len(args) != fnType.NumIn() {
			cmd.UsageFunc()(cmd)
			os.Exit(cli.ExitCodeUsage)
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
		build.Critical("wrapped function has wrong type signature")
	}
	numIn := fnType.NumIn()
	if numIn < 1 {
		build.Critical("wrapped function has insufficient amount of arguments")
	}
	if fnType.In(0).Elem() != reflect.TypeOf(config) {
		build.Critical("wrapped function should have a *Config param as first argument")
	}
	for i := 1; i < numIn; i++ {
		if fnType.In(i).Kind() != reflect.String {
			build.Critical("wrapped function has wrong type signature")
		}
	}

	return func(cmd *cobra.Command, args []string) {
		if len(args) != numIn+1 {
			cmd.UsageFunc()(cmd)
			os.Exit(cli.ExitCodeUsage)
		}
		argVals := make([]reflect.Value, numIn)
		argVals[0] = reflect.ValueOf(config)
		for i := 1; i < numIn; i++ {
			argVals[i] = reflect.ValueOf(args[i])
		}
		fnVal.Call(argVals)
	}
}

// NewCommandLineClient creates a new CLI client, which can be run as it is,
// or be extended/modified to fit your needs.
// If a config is not loaded automatically, it will be tried to be loaded from the daemon
// automatically, this might however fail.
func NewCommandLineClient(address, name, userAgent string) (*CommandLineClient, error) {
	if address == "" {
		address = "http://localhost:23110"
	}
	if name == "" {
		name = "R?v?ne"
	}
	client := new(CommandLineClient)
	client.HTTPClient = &api.HTTPClient{
		RootURL:   address,
		UserAgent: userAgent,
	}

	var consensusCmd *consensusCmd
	consensusCmd, client.ConsensusCmd = createConsensusCmd(client)
	client.RootCmd = &cobra.Command{
		Use:               os.Args[0],
		Short:             fmt.Sprintf("%s Client", strings.Title(name)),
		Long:              fmt.Sprintf("%s Client", strings.Title(name)),
		Run:               Wrap(consensusCmd.rootCmd),
		PersistentPreRunE: client.preRunE,
	}

	// create command tree
	client.RootCmd.AddCommand(client.ConsensusCmd)
	client.RootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Long:  "Print version information.",
		Run: Wrap(func() {
			fmt.Printf("%s Client v%s\r\n",
				strings.Title(client.Config.ChainName),
				client.Config.ChainVersion.String())

			fmt.Println()
			fmt.Printf("Go Version   v%s\r\n", runtime.Version()[2:])
			fmt.Printf("GOOS         %s\r\n", runtime.GOOS)
			fmt.Printf("GOARCH       %s\r\n", runtime.GOARCH)
		}),
	})
	client.RootCmd.AddCommand(&cobra.Command{
		Use:   "stop",
		Short: fmt.Sprintf("Stop the %s daemon", name),
		Long:  fmt.Sprintf("Stop the %s daemon.", name),
		Run: Wrap(func() {
			err := client.Post("/daemon/stop", "")
			if err != nil {
				cli.Die("Could not stop daemon:", err)
			}
			fmt.Printf("%s daemon stopped.\n", client.Config.ChainName)
		}),
	})

	client.WalletCmd = createWalletCmd(client)
	client.RootCmd.AddCommand(client.WalletCmd.Command)

	client.AtomicSwapCmd = createAtomicSwapCmd(client)
	client.RootCmd.AddCommand(client.AtomicSwapCmd)

	client.GatewayCmd = createGatewayCmd(client)
	client.RootCmd.AddCommand(client.GatewayCmd)

	client.ExploreCmd = createExploreCmd(client)
	client.RootCmd.AddCommand(client.ExploreCmd)

	client.MergeCmd = createMergeCmd(client)
	client.RootCmd.AddCommand(client.MergeCmd)

	// parse flags
	client.RootCmd.PersistentFlags().StringVarP(&client.HTTPClient.RootURL, "addr", "a",
		client.HTTPClient.RootURL, fmt.Sprintf(
			"which host/port to communicate with (i.e. the host/port %sd is listening on)",
			name))

	// return client
	return client, nil
}

// CommandLineClient represents the Rivine Reference CLI Client implementation, which can be run as it is,
// or be extended/modified to fit your needs.
//
// If a config is not loaded automatically, it will be tried to be loaded from the daemon
// automatically, this might however fail, in which case a DefaultConfigFunc can be registered as callback to fallback to.
type CommandLineClient struct {
	*api.HTTPClient

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
			fmt.Fprintf(os.Stderr, "fetching config from daemon failed: %v\r\n", err)
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
