package daemon

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/spf13/pflag"

	"github.com/threefoldtech/rivine/modules"
	"github.com/threefoldtech/rivine/pkg/cli"
	"github.com/threefoldtech/rivine/types"
)

const (
	// RivineUserAgent is the user agent used by Rivine by default.
	RivineUserAgent = "Rivine-Agent"
)

type (
	// Config contains all configurable variables for rivined.
	Config struct {
		BlockchainInfo types.BlockchainInfo

		// the password required to use the http api,
		// if `AuthenticateAPI` is true, and the password is the empty string,
		// a password will be prompted when the daemon starts
		APIPassword string

		// the host:port for the HTTP API to listen on.
		// If `AllowAPIBind` is false, only localhost hosts are allowed
		APIaddr string
		// the host:port to listen for RPC calls
		RPCaddr string
		// indicates that the http API can listen on a non localhost address.
		// If this is true, then the AuthenticateAPI parameter
		// must also be true
		AllowAPIBind bool

		// indicates that the daemon should not try to connect to
		// the bootstrap nodes
		NoBootstrap bool
		// the user agent required to connect to the http api.
		RequiredUserAgent string
		// indicates if the http api is password protected
		AuthenticateAPI bool

		// indicates if profile info should be collected while
		// the daemon is running
		Profile bool
		// name of the directory to store the profile info,
		// should this be collected
		ProfileDir string
		// the parent directory where the individual module
		// directories will be created
		RootPersistentDir string

		//Verbose (debug) logging
		VerboseLogging bool

		// Optional BootstrapPeers we want to use instead of the default NetworkConfigs.
		BootstrapPeers []modules.NetAddress
	}

	// NetworkConfig are variables for a particular chain. Currently, these are genesis constants and bootstrap peers
	NetworkConfig struct {
		// Blockchain Constants for this network
		Constants types.ChainConstants
		// BootstrapPeers for this network
		BootstrapPeers []modules.NetAddress
	}
)

// DefaultConfig returns the default daemon configuration
func DefaultConfig() Config {
	return Config{
		BlockchainInfo: types.DefaultBlockchainInfo(),

		APIPassword: "",

		APIaddr:      "localhost:23110",
		RPCaddr:      ":23112",
		AllowAPIBind: false,

		NoBootstrap:       false,
		RequiredUserAgent: RivineUserAgent,
		AuthenticateAPI:   false,

		Profile:           false,
		ProfileDir:        "profiles",
		RootPersistentDir: "",
		VerboseLogging:    false,

		BootstrapPeers: nil,
	}
}

// RegisterAsFlags registers all properties —for which it makes sense— as a flag.
func (cfg *Config) RegisterAsFlags(flagSet *pflag.FlagSet) {
	flagSet.StringVarP(&cfg.RequiredUserAgent, "agent", "", cfg.RequiredUserAgent, "required substring for the user agent")
	flagSet.StringVarP(&cfg.ProfileDir, "profile-directory", "", cfg.ProfileDir, "location of the profiling directory")
	flagSet.StringVarP(&cfg.APIaddr, "api-addr", "", cfg.APIaddr, "which host:port the API server listens on")
	flagSet.StringVarP(&cfg.RootPersistentDir, "persistent-directory", "d", cfg.RootPersistentDir,
		"location of the root directory used to store persistent data of the daemon of "+
			cfg.BlockchainInfo.Name)

	flagSet.BoolVarP(&cfg.VerboseLogging, "verboselogging", "v", false, "enable logging of debug information in the logfiles of the modules")
	flagSet.BoolVarP(&cfg.NoBootstrap, "no-bootstrap", "", cfg.NoBootstrap, "disable bootstrapping on this run")
	flagSet.BoolVarP(&cfg.Profile, "profile", "", cfg.Profile, "enable profiling")
	flagSet.StringVarP(&cfg.RPCaddr, "rpc-addr", "", cfg.RPCaddr, "which port the gateway listens on")
	flagSet.BoolVarP(&cfg.AuthenticateAPI, "authenticate-api", "", cfg.AuthenticateAPI, "enable API password protection")
	flagSet.BoolVarP(&cfg.AllowAPIBind, "disable-api-security", "", cfg.AllowAPIBind, fmt.Sprintf("allow the daemon of %s to listen on a non-localhost address (DANGEROUS)", cfg.BlockchainInfo.Name))
	flagSet.StringVarP(&cfg.BlockchainInfo.NetworkName, "network", "n", cfg.BlockchainInfo.NetworkName, "the name of the network to which the daemon connects")

	cli.NetAddressArrayFlagVar(flagSet, &cfg.BootstrapPeers, "bootstrap-peers",
		"overwrite the bootstrap peers to use, instead of using the default bootstrap peers")
}

// ProcessConfig checks the configuration values and performs cleanup on
// incorrect-but-allowed values.
func ProcessConfig(config Config) Config {
	config.APIaddr = processNetAddr(config.APIaddr)
	config.RPCaddr = processNetAddr(config.RPCaddr)
	return config
}

// VerifyAPISecurity checks that the security values are consistent with a
// sane, secure system.
func VerifyAPISecurity(cfg Config) error {
	// Make sure that only the loopback address is allowed unless the
	// --disable-api-security flag has been used.
	if !cfg.AllowAPIBind {
		addr := modules.NetAddress(cfg.APIaddr)
		if !addr.IsLoopback() {
			if addr.Host() == "" {
				return fmt.Errorf("a blank host will listen on all interfaces, did you mean localhost:%v?\nyou must pass --disable-api-security to bind daemon of %s to a non-localhost address", addr.Port(), cfg.BlockchainInfo.Name)
			}
			return fmt.Errorf("you must pass --disable-api-security to bind daemon of %s to a non-localhost address", cfg.BlockchainInfo.Name)
		}
		return nil
	}

	// If the --disable-api-security flag is used, enforce that
	// --authenticate-api must also be used.
	if cfg.AllowAPIBind && !cfg.AuthenticateAPI {
		return errors.New("cannot use --disable-api-security without setting an api password")
	}
	return nil
}

// processNetAddr adds a ':' to a bare integer, so that it is a proper port
// number.
func processNetAddr(addr string) string {
	_, err := strconv.Atoi(addr)
	if err == nil {
		return ":" + addr
	}
	return addr
}

// DefaultNetworkConfig returns the default network config based on a given network name.
func DefaultNetworkConfig(networkName string) (NetworkConfig, error) {
	if networkName == "" {
		networkName = types.DefaultNetworkName()
	}

	var networkCfg NetworkConfig
	switch networkName {
	case "standard":
		networkCfg.Constants = types.StandardnetChainConstants()
	case "testnet":
		networkCfg.Constants = types.TestnetChainConstants()
	case "devnet":
		networkCfg.Constants = types.DevnetChainConstants()
	default:
		return NetworkConfig{}, fmt.Errorf("unknown network name %s", networkName)
	}
	return networkCfg, nil
}
