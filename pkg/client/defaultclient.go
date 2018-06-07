package client

import (
	"fmt"
	"log"
	"os"
	"reflect"
	"strings"

	"github.com/rivine/rivine/build"
	"github.com/rivine/rivine/pkg/daemon"
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

// Config defines the configuration for the default (CLI) client.
type Config struct {
	ChainName    string
	NetworkName  string
	ChainVersion build.ProtocolVersion

	CurrencyUnits             types.CurrencyUnits
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
	println(fmt.Sprintf("%s Client v", strings.Title(_DefaultClient.name)) + _DefaultClient.version.String())
}

// hidden globals :()
var (
	_DefaultClient struct {
		name       string
		version    build.ProtocolVersion
		httpClient HTTPClient
	}

	_CurrencyUnits             types.CurrencyUnits
	_CurrencyCoinUnit          string
	_CurrencyConvertor         CurrencyConvertor
	_MinimumTransactionFee     types.Currency
	_DefaultTransactionVersion types.TransactionVersion

	_BlockFrequencyInSeconds int64
	_GenesisBlockTimestamp   types.Timestamp
)

func fetchConfigFromDaemon() *Config {
	var constants daemon.SiaConstants
	err := _DefaultClient.httpClient.GetAPI("/daemon/constants", &constants)
	if err != nil {
		log.Println("[ERROR] Failed to load constants from daemon: ", err)
		return nil
	}
	return &Config{
		ChainName:    constants.ChainInfo.Name,
		NetworkName:  constants.ChainInfo.NetworkName,
		ChainVersion: constants.ChainInfo.ChainVersion,
		CurrencyUnits: types.CurrencyUnits{
			OneCoin: constants.OneCoin,
		},
		MinimumTransactionFee:     constants.MinimumTransactionFee,
		DefaultTransactionVersion: constants.DefaultTransactionVersion,
		BlockFrequencyInSeconds:   int64(constants.BlockFrequency),
		GenesisBlockTimestamp:     constants.GenesisTimestamp,
	}
}

// DefaultCLIClient creates a new client for the given address.
// The given address is used to connect to the client daemon, using its REST API.
// configFunc has to be given and has two functions:
// most importantly it serves to provide a default config, should the daemon/constants endpoint not be available,
// and secondly it allows the CLi client to overwrite one or multiple properties.
// Address input parameter defaults to "http://localhost:23110" if none is given.
func DefaultCLIClient(address string, configFunc func(*Config) Config) {
	if address == "" {
		address = "http://localhost:23110"
	}
	_DefaultClient.httpClient.RootURL = address
	_DefaultClient.name, _DefaultClient.version = "R?v?ne", build.Version // defaults for now, until we loaded real values

	root := &cobra.Command{
		Use:   os.Args[0],
		Short: fmt.Sprintf("%s Client v", strings.Title(_DefaultClient.name)) + _DefaultClient.version.String(),
		Long:  fmt.Sprintf("%s Client v", strings.Title(_DefaultClient.name)) + _DefaultClient.version.String(),
		Run:   Wrap(consensuscmd),
		PersistentPreRun: func(*cobra.Command, []string) {
			// sanituze root URL
			url, err := sanitizeURL(_DefaultClient.httpClient.RootURL)
			if err != nil {
				Die("invalid", strings.Title(_DefaultClient.name), "daemon RPC address", _DefaultClient.httpClient.RootURL, ":", err)
			}
			_DefaultClient.httpClient.RootURL = url

			// configure client
			cfg := configFunc(fetchConfigFromDaemon())
			_DefaultClient.name = cfg.ChainName
			_DefaultClient.version = cfg.ChainVersion
			_CurrencyUnits = cfg.CurrencyUnits
			_CurrencyConvertor = NewCurrencyConvertor(_CurrencyUnits)
			_MinimumTransactionFee = cfg.MinimumTransactionFee
			_DefaultTransactionVersion = cfg.DefaultTransactionVersion
			_BlockFrequencyInSeconds = cfg.BlockFrequencyInSeconds
			_GenesisBlockTimestamp = cfg.GenesisBlockTimestamp
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
		Short: fmt.Sprintf("Stop the %s daemon", _DefaultClient.name),
		Long:  fmt.Sprintf("Stop the %s daemon.", _DefaultClient.name),
		Run:   Wrap(stopcmd),
	}

	root.AddCommand(stopCmd)

	createWalletCommands()
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
			_DefaultClient.name))

	if err := root.Execute(); err != nil {
		// Since no commands return errors (all commands set Command.Run instead of
		// Command.RunE), Command.Execute() should only return an error on an
		// invalid command or flag. Therefore Command.Usage() was called (assuming
		// Command.SilenceUsage is false) and we should exit with exitCodeUsage.
		os.Exit(ExitCodeUsage)
	}
}

func init() {
	_CurrencyUnits = types.DefaultCurrencyUnits()
}
