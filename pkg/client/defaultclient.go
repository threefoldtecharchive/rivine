package client

import (
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/rivine/rivine/build"
	"github.com/rivine/rivine/types"
	"github.com/spf13/cobra"
)

// exit codes
// inspired by sysexits.h
const (
	ExitCodeGeneral  = 1 // Not in sysexits.h, but is standard practice.
	ExitCodeNotFound = 2
	ExitCodeUsage    = 64 // EX_USAGE in sysexits.h
)

// Config defines the configuration for the default (CLI) client.
type Config struct {
	Address               string
	Name                  string
	Version               build.ProtocolVersion // blockchain version (is NOT Rivine Protocol Version)
	CurrencyCoinUnit      string
	CurrencyUnits         types.CurrencyUnits
	MinimumTransactionFee types.Currency

	// version used to create new transactions
	DefaultTransactionVersion types.TransactionVersion
}

// DefaultConfig creates the default configuration for the default (CLI) client.
func DefaultConfig() Config {
	chainConstants := types.DefaultChainConstants()
	return Config{
		Address:                   "http://localhost:23110",
		Name:                      "Rivine",
		Version:                   build.Version,
		CurrencyCoinUnit:          "ROC",
		CurrencyUnits:             types.DefaultCurrencyUnits(),
		MinimumTransactionFee:     chainConstants.MinimumTransactionFee,
		DefaultTransactionVersion: chainConstants.DefaultTransactionVersion,
	}
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
)

// DefaultCLIClient creates a new client using the given params as the default config,
// and an optional flag-based system to overrride some.
func DefaultCLIClient(cfg Config) {
	_DefaultClient.name = cfg.Name
	_DefaultClient.httpClient.RootURL = cfg.Address
	_DefaultClient.version = cfg.Version
	_CurrencyCoinUnit = cfg.CurrencyCoinUnit
	_CurrencyUnits = cfg.CurrencyUnits
	_MinimumTransactionFee = cfg.MinimumTransactionFee
	_DefaultTransactionVersion = cfg.DefaultTransactionVersion

	var err error
	_CurrencyConvertor, err = NewCurrencyConvertor(_CurrencyUnits)
	if err != nil {
		Die("couldn't create currency convertor:", err)
	}

	root := &cobra.Command{
		Use:   os.Args[0],
		Short: fmt.Sprintf("%s Client v", strings.Title(_DefaultClient.name)) + _DefaultClient.version.String(),
		Long:  fmt.Sprintf("%s Client v", strings.Title(_DefaultClient.name)) + _DefaultClient.version.String(),
		Run:   Wrap(consensuscmd),
		PersistentPreRun: func(*cobra.Command, []string) {
			url, err := sanitizeURL(_DefaultClient.httpClient.RootURL)
			if err != nil {
				Die("invalid", strings.Title(_DefaultClient.name), "daemon RPC address", _DefaultClient.httpClient.RootURL, ":", err)
			}
			_DefaultClient.httpClient.RootURL = url
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
