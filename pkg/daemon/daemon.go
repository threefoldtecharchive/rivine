package daemon

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/rivine/rivine/api"
	"github.com/rivine/rivine/build"
	"github.com/rivine/rivine/modules"
	"github.com/rivine/rivine/modules/blockcreator"
	"github.com/rivine/rivine/modules/consensus"
	"github.com/rivine/rivine/modules/explorer"
	"github.com/rivine/rivine/modules/gateway"
	"github.com/rivine/rivine/modules/transactionpool"
	"github.com/rivine/rivine/modules/wallet"
	"github.com/rivine/rivine/types"

	"github.com/bgentry/speakeasy"
)

// Config contains all configurable variables for rivined.
type Config struct {
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
	//  If this is true, then the AuthenticateAPI parameter
	// must also be true
	AllowAPIBind bool

	// the modules to enable, this string must contain one letter
	// for each module (order does not matter).
	// All required modules must be specified,
	// if a required (parent) module is not present,
	// an error is returned
	Modules string
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

	// Network defines the network config to use
	NetworkName string
	// optional network config constructor,
	// if you're implementing your own rivine-based blockchain,
	// you'll probably want to define this one,
	// as otherwise a pure rivine blockchain config will be created,
	// within this function you should also Register any specific
	// transaction, conditions and fulfillments
	CreateNetworkConfig func(name string) (NetworkConfig, error)
}

// DefaultConfig returns the default daemon configuration
func DefaultConfig() Config {
	return Config{
		BlockchainInfo: types.DefaultBlockchainInfo(),

		APIPassword: "",

		APIaddr:      "localhost:23110",
		RPCaddr:      ":23112",
		AllowAPIBind: false,

		Modules:           "cgtwb",
		NoBootstrap:       false,
		RequiredUserAgent: "Rivine-Agent",
		AuthenticateAPI:   false,

		Profile:           false,
		ProfileDir:        "profiles",
		RootPersistentDir: "",

		NetworkName: build.Release,
	}
}

func (cfg *Config) createConfiguredNetworkConfig() (NetworkConfig, error) {
	if cfg.NetworkName == "" {
		// default to build.Release as network name
		cfg.NetworkName = build.Release
	}
	if cfg.CreateNetworkConfig != nil {
		// use custom network config creator
		return cfg.CreateNetworkConfig(cfg.NetworkName)
	}

	// use default network config creator
	networkCfg := NetworkConfig{
		Constants: types.DefaultChainConstants(),
	}
	if cfg.NetworkName == "standard" {
		networkCfg.BootstrapPeers = []modules.NetAddress{
			"136.243.144.132:23112",
			"[2a01:4f8:171:1303::2]:23112",
			"bootstrap2.rivine.io:23112",
			"bootstrap3.rivine.io:23112",
		}
	}
	return networkCfg, nil
}

// verifyAPISecurity checks that the security values are consistent with a
// sane, secure system.
func verifyAPISecurity(cfg Config) error {
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

// processModules makes the modules string lowercase to make checking if a
// module in the string easier, and returns an error if the string contains an
// invalid module character.
func processModules(modules string) (string, error) {
	modules = strings.ToLower(modules)
	validModules := "cgtwebd"
	invalidModules := modules
	for _, m := range validModules {
		invalidModules = strings.Replace(invalidModules, string(m), "", 1)
	}
	if len(invalidModules) > 0 {
		return "", errors.New("Unable to parse --modules flag, unrecognized or duplicate modules: " + invalidModules)
	}
	return modules, nil
}

// processConfig checks the configuration values and performs cleanup on
// incorrect-but-allowed values.
func processConfig(config *Config) error {
	var err1 error
	config.APIaddr = processNetAddr(config.APIaddr)
	config.RPCaddr = processNetAddr(config.RPCaddr)
	config.Modules, err1 = processModules(config.Modules)
	err2 := verifyAPISecurity(*config)
	return build.JoinErrors([]error{err1, err2}, ", and ")
}

// StartDaemon uses the config parameters
// to initialize Rivine modules and start
func StartDaemon(cfg Config) (err error) {
	networkConfig, err := cfg.createConfiguredNetworkConfig()
	if err != nil {
		return
	}

	err = networkConfig.Constants.Validate()
	if err != nil {
		return err
	}
	// Silently append a subdirectory for storage with the name of the network so we don't create conflicts
	cfg.RootPersistentDir = filepath.Join(cfg.RootPersistentDir, cfg.NetworkName)
	// Check if we require an api password
	if cfg.AuthenticateAPI {
		// if its not set, ask one now
		if cfg.APIPassword == "" {
			// Prompt user for API password.
			cfg.APIPassword, err = speakeasy.Ask("Enter API password: ")
			if err != nil {
				return err
			}
		}
		if cfg.APIPassword == "" {
			return errors.New("password cannot be blank")
		}
	} else {
		// If authenticateAPI is not set, explicitly set the password to the empty string.
		// This way the api server maintains consistency with the authenticateAPI var, even if apiPassword is set (possibly by mistake)
		cfg.APIPassword = ""
	}

	// Process the config variables
	// If there is an error or inconsistency in the config, we return, so there is no need to correct any values
	err = processConfig(&cfg)
	if err != nil {
		return err
	}

	// Print a startup message.
	fmt.Println("Loading...")
	loadStart := time.Now()

	// Create the server and start serving daemon routes immediately.
	fmt.Printf("(0/%d) Loading daemon of "+cfg.BlockchainInfo.Name+"...\n", len(cfg.Modules))
	srv, err := NewServer(cfg.APIaddr, cfg.RequiredUserAgent, cfg.APIPassword,
		networkConfig.Constants, cfg.BlockchainInfo)
	if err != nil {
		return err
	}

	servErrs := make(chan error)
	go func() {
		servErrs <- srv.Serve()
	}()

	// Initialize the Rivine modules
	i := 0
	var g modules.Gateway
	if strings.Contains(cfg.Modules, "g") {
		i++
		fmt.Printf("(%d/%d) Loading gateway...\n", i, len(cfg.Modules))
		g, err = gateway.New(cfg.RPCaddr, !cfg.NoBootstrap,
			filepath.Join(cfg.RootPersistentDir, modules.GatewayDir),
			cfg.BlockchainInfo, networkConfig.Constants, networkConfig.BootstrapPeers)
		if err != nil {
			return err
		}
		defer func() {
			fmt.Println("Closing gateway...")
			err := g.Close()
			if err != nil {
				fmt.Println("Error during gateway shutdown:", err)
			}
		}()

	}
	var cs modules.ConsensusSet
	if strings.Contains(cfg.Modules, "c") {
		i++
		fmt.Printf("(%d/%d) Loading consensus...\n", i, len(cfg.Modules))
		cs, err = consensus.New(g, !cfg.NoBootstrap,
			filepath.Join(cfg.RootPersistentDir, modules.ConsensusDir),
			cfg.BlockchainInfo, networkConfig.Constants)
		if err != nil {
			return err
		}
		defer func() {
			fmt.Println("Closing consensus set...")
			err := cs.Close()
			if err != nil {
				fmt.Println("Error during consensus set shutdown:", err)
			}
		}()

	}
	var e modules.Explorer
	if strings.Contains(cfg.Modules, "e") {
		i++
		fmt.Printf("(%d/%d) Loading explorer...\n", i, len(cfg.Modules))
		e, err = explorer.New(cs,
			filepath.Join(cfg.RootPersistentDir, modules.ExplorerDir),
			cfg.BlockchainInfo, networkConfig.Constants)
		if err != nil {
			return err
		}
		defer func() {
			fmt.Println("Closing explorer...")
			err := e.Close()
			if err != nil {
				fmt.Println("Error during explorer shutdown:", err)
			}
		}()

	}
	var tpool modules.TransactionPool
	if strings.Contains(cfg.Modules, "t") {
		i++
		fmt.Printf("(%d/%d) Loading transaction pool...\n", i, len(cfg.Modules))
		tpool, err = transactionpool.New(cs, g,
			filepath.Join(cfg.RootPersistentDir, modules.TransactionPoolDir),
			cfg.BlockchainInfo, networkConfig.Constants)
		if err != nil {
			return err
		}
		defer func() {
			fmt.Println("Closing transaction pool...")
			err := tpool.Close()
			if err != nil {
				fmt.Println("Error during transaction pool shutdown:", err)
			}
		}()
	}
	var w modules.Wallet
	if strings.Contains(cfg.Modules, "w") {
		i++
		fmt.Printf("(%d/%d) Loading wallet...\n", i, len(cfg.Modules))
		w, err = wallet.New(cs, tpool,
			filepath.Join(cfg.RootPersistentDir, modules.WalletDir),
			cfg.BlockchainInfo, networkConfig.Constants)
		if err != nil {
			return err
		}
		defer func() {
			fmt.Println("Closing wallet...")
			err := w.Close()
			if err != nil {
				fmt.Println("Error during wallet shutdown:", err)
			}
		}()

	}
	var b modules.BlockCreator
	if strings.Contains(cfg.Modules, "b") {
		i++
		fmt.Printf("(%d/%d) Loading block creator...\n", i, len(cfg.Modules))
		b, err = blockcreator.New(cs, tpool, w,
			filepath.Join(cfg.RootPersistentDir, modules.BlockCreatorDir),
			cfg.BlockchainInfo, networkConfig.Constants)
		if err != nil {
			return err
		}
		defer func() {
			fmt.Println("Closing block creator...")
			err := b.Close()
			if err != nil {
				fmt.Println("Error during block creator shutdown:", err)
			}
		}()
	}

	// Create the Rivine API
	a := api.New(
		cfg.RequiredUserAgent,
		cfg.APIPassword,
		cs,
		e,
		g,
		tpool,
		w,
	)

	// connect the API to the server
	srv.mux.Handle("/", a)

	// stop the server if a kill signal is caught
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, os.Kill)
	go func() {
		<-sigChan
		fmt.Println("\rCaught stop signal, quitting...")
		srv.Close()
	}()

	// Print a 'startup complete' message.
	startupTime := time.Since(loadStart)
	fmt.Println("Finished loading in", startupTime.Seconds(), "seconds")

	err = <-servErrs
	if err != nil {
		build.Critical(err)
	}

	return nil
}
