package client

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"
	"sync"

	"github.com/rivine/rivine/build"
	"github.com/rivine/rivine/modules"
	"github.com/rivine/rivine/types"
	"github.com/spf13/cobra"
)

// exit codes
// inspired by sysexits.h
const (
	ExitCodeGeneral   = 1 // Not in sysexits.h, but is standard practice.
	ExitCodeNotFound  = 2
	ExitCodeCancelled = 3
	ExitCodeForbidden = 4
	ExitCodeUsage     = 64 // EX_USAGE in sysexits.h
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

// clientVersion prints the client version and exits
func clientVersion() {
	cfg := _ConfigStorage.Config()
	println(fmt.Sprintf("%s Client v", strings.Title(cfg.ChainName)) + cfg.ChainVersion.String())
}

// hidden globals :()
var (
	_DefaultClient struct {
		httpClient HTTPClient
	}

	_ConfigStorage     *lazyConfigFetcher
	_CurrencyConvertor lazyCurrencyConvertor
)

// DefaultCLIClient creates a new client for the given address.
// The given address is used to connect to the client daemon, using its REST API.
// configFunc allows you to overwrite any config, should some value returned by the daemon not be desired,
// it is however optional and does not have to be given
func DefaultCLIClient(address, name string, configFunc func(*Config) Config) {
	if address == "" {
		address = "http://localhost:23110"
	}
	if name == "" {
		name = "R?v?ne"
	}
	_DefaultClient.httpClient.RootURL = address

	root := &cobra.Command{
		Use:   os.Args[0],
		Short: fmt.Sprintf("%s Client", strings.Title(name)),
		Long:  fmt.Sprintf("%s Client", strings.Title(name)),
		Run:   Wrap(consensuscmd),
		PersistentPreRun: func(*cobra.Command, []string) {
			// sanituze root URL
			url, err := sanitizeURL(_DefaultClient.httpClient.RootURL)
			if err != nil {
				Die("invalid", strings.Title(name), "daemon RPC address", _DefaultClient.httpClient.RootURL, ":", err)
			}
			_DefaultClient.httpClient.RootURL = url
			_ConfigStorage = newLazyConfigFetcher(configFunc)
		},
	}

	// create command tree
	root.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Long:  "Print version information.",
		Run:   Wrap(clientVersion),
	})
	stopCmd := &cobra.Command{
		Use:   "stop",
		Short: fmt.Sprintf("Stop the %s daemon", name),
		Long:  fmt.Sprintf("Stop the %s daemon.", name),
		Run:   Wrap(stopcmd),
	}

	root.AddCommand(stopCmd)

	root.AddCommand(walletCmd)
	walletCmd.AddCommand(
		walletAddressCmd,
		walletAddressesCmd,
		walletInitCmd,
		walletRecoverCmd,
		walletLoadCmd,
		walletLockCmd,
		walletSeedsCmd,
		walletSendCmd,
		walletBalanceCmd,
		walletTransactionsCmd,
		walletUnlockCmd,
		walletBlockStakeStatCmd,
		walletRegisterDataCmd,
		walletListCmd,
		walletCreateCmd,
		walletSignCmd)

	root.AddCommand(atomicSwapCmd)
	atomicSwapCmd.AddCommand(
		atomicSwapParticipateCmd,
		atomicSwapInitiateCmd,
		atomicSwapAuditCmd,
		atomicSwapExtractSecretCmd,
		atomicSwapRedeemCmd,
		atomicSwapRefundCmd,
	)

	walletSendCmd.AddCommand(
		walletSendCoinsCmd,
		walletSendBlockStakesCmd,
		walletSendTxnCmd)

	walletLoadCmd.AddCommand(walletLoadSeedCmd)

	walletListCmd.AddCommand(
		walletListUnlockedCmd,
		walletListLockedCmd)

	walletCreateCmd.AddCommand(
		walletCreateMultisisgAddress,
		walletCreateCoinTxnCmd,
		walletCreateBlockStakeTxnCmd)

	root.AddCommand(gatewayCmd)
	gatewayCmd.AddCommand(
		gatewayConnectCmd,
		gatewayDisconnectCmd,
		gatewayAddressCmd,
		gatewayListCmd)

	root.AddCommand(consensusCmd)
	consensusCmd.AddCommand(
		consensusTransactionCmd,
	)

	root.AddCommand(exploreCmd)
	exploreCmd.AddCommand(
		exploreBlockCmd,
		exploreHashCmd,
	)

	root.AddCommand(mergeCmd)
	mergeCmd.AddCommand(
		mergeTransactionsCmd,
	)

	// parse flags
	root.PersistentFlags().StringVarP(&_DefaultClient.httpClient.RootURL, "addr", "a",
		_DefaultClient.httpClient.RootURL, fmt.Sprintf(
			"which host/port to communicate with (i.e. the host/port %sd is listening on)",
			name))

	if err := root.Execute(); err != nil {
		// Since no commands return errors (all commands set Command.Run instead of
		// Command.RunE), Command.Execute() should only return an error on an
		// invalid command or flag. Therefore Command.Usage() was called (assuming
		// Command.SilenceUsage is false) and we should exit with exitCodeUsage.
		os.Exit(ExitCodeUsage)
	}
}

// lazyConfigFetcher can be used in order to load the config only when needing,
// delaying the fetching of daemon constants until it is needed for the first time
type lazyConfigFetcher struct {
	config                     Config
	configSanitizer            func(*Config) Config
	fetchAndSanitizeConfigOnce sync.Once
}

func newLazyConfigFetcher(f func(*Config) Config) *lazyConfigFetcher {
	if f == nil {
		DieWithError(
			"failed to create lazy config fetcher",
			errors.New("daemon is created without a config sanitization function given to it"))
	}
	return &lazyConfigFetcher{configSanitizer: f}
}

// Config returns the config in a lazy manner
func (lcf *lazyConfigFetcher) Config() Config {
	lcf.fetchAndSanitizeConfigOnce.Do(lcf.fetchAndSanitizeConfig)
	return lcf.config
}

func (lcf *lazyConfigFetcher) fetchAndSanitizeConfig() {
	lcf.config = lcf.configSanitizer(fetchConfigFromDaemon())
}

// fetchConfigFromDaemon fetches constants and creates a config, by fetching the constants from the daemon.
// Can return nil in case the fetching wasn't possible
func fetchConfigFromDaemon() *Config {
	var constants modules.DaemonConstants
	err := _DefaultClient.httpClient.GetAPI("/daemon/constants", &constants)
	if err == nil {
		// returned config from received constants from the daemon's server module
		cfg := ConfigFromDaemonConstants(constants)
		return &cfg
	}
	fmt.Fprintln(os.Stderr,
		"[WARNING] failed to fetch constants from the daemon's server module: ", err)
	err = _DefaultClient.httpClient.GetAPI("/explorer/constants", &constants)
	if err != nil {
		fmt.Fprintln(os.Stderr,
			"[WARNING] "+
				"failed to load constants from daemon's server and explorer modules: ", err)
		return nil
	}
	if constants.ChainInfo == (types.BlockchainInfo{}) {
		// only since 1.0.7 do we support the full set of public daemon constants for both
		// the explorer endpoint as well as the daemon endpoint,
		// so we need to validate this
		fmt.Fprintln(os.Stderr,
			"[WARNING] "+
				"failed to load constants from daemon's server and explorer modules: "+
				"explorer modules does not support the full exposure of public daemon constants")
		return nil
	}
	// returned config from received constants from the daemon's explorer module
	cfg := ConfigFromDaemonConstants(constants)
	return &cfg
}

type lazyCurrencyConvertor struct {
	convertor           CurrencyConvertor
	createConvertorOnce sync.Once
}

// ParseCoinString parses the given string assumed to be in the default unit,
// and parses it into an in-memory currency unit of the smallest unit.
// It will fail if the given string is invalid or too precise.
func (cc *lazyCurrencyConvertor) ParseCoinString(str string) (types.Currency, error) {
	return cc.getConvertor().ParseCoinString(str)
}

// ToCoinString turns the in-memory currency unit,
// into a string version of the default currency unit.
// This can never fail, as the only thing it can do is make a number smaller.
func (cc *lazyCurrencyConvertor) ToCoinString(c types.Currency) string {
	return cc.getConvertor().ToCoinString(c)
}

// ToCoinStringWithUnit turns the in-memory currency unit,
// into a string version of the default currency unit.
// This can never fail, as the only thing it can do is make a number smaller.
// It also adds the unit of the coin behind the coin.
func (cc *lazyCurrencyConvertor) ToCoinStringWithUnit(c types.Currency) string {
	return cc.getConvertor().ToCoinStringWithUnit(c)
}

// CoinArgDescription is used to print a helpful arg description message,
// for this convertor.
func (cc *lazyCurrencyConvertor) CoinArgDescription(argName string) string {
	return cc.getConvertor().CoinArgDescription(argName)
}

func (cc *lazyCurrencyConvertor) getConvertor() CurrencyConvertor {
	cc.createConvertorOnce.Do(cc.createConvertor)
	return cc.convertor
}

// createConvertor creates the currency convertor, by gettng the constants
// and creating the convertor with it. This function should only be executed once,
// and so in a thread-safe manner.
func (cc *lazyCurrencyConvertor) createConvertor() {
	cfg := _ConfigStorage.Config()
	cc.convertor = NewCurrencyConvertor(cfg.CurrencyUnits, cfg.CurrencyCoinUnit)
}
