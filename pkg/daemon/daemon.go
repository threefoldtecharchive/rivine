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

	"github.com/bgentry/speakeasy"
)

var (
	// globalConfig is used by the cobra package to fill out the configuration
	// variables.
	globalConfig Config
)

// The Config struct contains all configurable variables for siad. It is
// compatible with gcfg.
type Config struct {
	APIPassword string

	APIaddr      string
	RPCaddr      string
	HostAddr     string
	AllowAPIBind bool

	Modules           string
	NoBootstrap       bool
	RequiredUserAgent string
	AuthenticateAPI   bool

	Profile    bool
	ProfileDir string
	RivineDir  string
}

// DefaultConfig returns the default daemon configuration
func DefaultConfig() Config {
	return Config{
		APIPassword: "",

		APIaddr:      "localhost:23110",
		RPCaddr:      ":23112",
		HostAddr:     "",
		AllowAPIBind: false,

		Modules:           "cgtwb",
		NoBootstrap:       false,
		RequiredUserAgent: "Rivine-Agent",
		AuthenticateAPI:   false,

		Profile:    false,
		ProfileDir: "profiles",
		RivineDir:  "",
	}
}

// verifyAPISecurity checks that the security values are consistent with a
// sane, secure system.
func verifyAPISecurity(config Config) error {
	// Make sure that only the loopback address is allowed unless the
	// --disable-api-security flag has been used.
	if !config.AllowAPIBind {
		addr := modules.NetAddress(config.APIaddr)
		if !addr.IsLoopback() {
			if addr.Host() == "" {
				return fmt.Errorf("a blank host will listen on all interfaces, did you mean localhost:%v?\nyou must pass --disable-api-security to bind %s to a non-localhost address", addr.Port(), DaemonName)
			}
			return fmt.Errorf("you must pass --disable-api-security to bind %s to a non-localhost address", DaemonName)
		}
		return nil
	}

	// If the --disable-api-security flag is used, enforce that
	// --authenticate-api must also be used.
	if config.AllowAPIBind && !config.AuthenticateAPI {
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
	validModules := "cgtweb"
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
func processConfig(config Config) (Config, error) {
	var err1 error
	config.APIaddr = processNetAddr(config.APIaddr)
	config.RPCaddr = processNetAddr(config.RPCaddr)
	config.HostAddr = processNetAddr(config.HostAddr)
	config.Modules, err1 = processModules(config.Modules)
	err2 := verifyAPISecurity(config)
	err := build.JoinErrors([]error{err1, err2}, ", and ")
	if err != nil {
		return Config{}, err
	}
	return config, nil
}

// StartDaemon uses the config parameters to initialize Rivine modules and start
// rivined. Required parameters are, in order of appearance:
// 	- authenticateAPI (bool): indicates if the http api is password protected
// 	- apiPassword (string): the password required to use the http api, if `authenticateAPI` is true. If `authenticateAPI` is true, and the password is the empty string, a password will be prompted when the daemon starts
// 	- allowAPIBind (bool): indicates that the http API can listen on a non localhost address. If this is true, then the authenticateAPI parameter must also be true
// 	- apiAddr (string): the host:port for the http api to listen on. If allowAPIBind is false, only localhost hosts are allowed
// 	- rpcAddr (string): the host:port to listen for rpc calls
// 	- moduleString (string): the modules to enable, this string must contain one letter for each module (order does not matter). All required modules must be specified, if a required (parent) module is not present, an error is returned
// 	 	Allowed letters: g(gateway), c(consensus), w(wallet), t(transactionpool), b(blockcreator), e(explorer)
// 	- noBootstrap (bool): indicates that the daemon should not try to connect to the bootstrap nodes
// 	- requiredUserAgent (string): the user agent required to connect to the http api.
// 	- profile (bool): indicates if profile info should be collected while the daemon is running
// 	- profileDir (string): name of the directory to store the profile info, should this be collected
// 	- rivineDir (string): the parent directory where the individual module directories will be created
func StartDaemon(authenticateAPI bool, apiPassword string, allowAPIBind bool, apiAddr string, rpcAddr string, moduleString string, noBootstrap bool, requiredUserAgent string, profile bool, profileDir string, rivineDir string) (err error) {
	// Check if we require an api password
	if authenticateAPI {
		// if its not set, ask one now
		if apiPassword == "" {
			// Prompt user for API password.
			apiPassword, err = speakeasy.Ask("Enter API password: ")
			if err != nil {
				return err
			}
		}
		if apiPassword == "" {
			return errors.New("password cannot be blank")
		}
	} else {
		// If authenticateAPI is not set, explicitly set the password to the empty string.
		// This way the api server maintains consistency with the authenticateAPI var, even if apiPassword is set (possibly by mistake)
		apiPassword = ""
	}

	// create a config to validate the variables
	config := Config{
		APIPassword:       apiPassword,
		APIaddr:           apiAddr,
		RPCaddr:           rpcAddr,
		HostAddr:          "",
		AllowAPIBind:      allowAPIBind,
		Modules:           moduleString,
		NoBootstrap:       noBootstrap,
		RequiredUserAgent: requiredUserAgent,
		AuthenticateAPI:   authenticateAPI,
		Profile:           profile,
		ProfileDir:        profileDir,
		RivineDir:         rivineDir,
	}

	// Process the config variables
	// If there is an error or inconsistency in the config, we return, so there is no need to correct any values
	config, err = processConfig(config)
	if err != nil {
		return err
	}

	// Print a startup message.
	fmt.Println("Loading...")
	loadStart := time.Now()

	// Create the server and start serving daemon routes immediately.
	fmt.Printf("(0/%d) Loading "+DaemonName+"...\n", len(moduleString))
	srv, err := NewServer(apiAddr, requiredUserAgent, apiPassword)
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
	if strings.Contains(moduleString, "g") {
		i++
		fmt.Printf("(%d/%d) Loading gateway...\n", i, len(moduleString))
		g, err = gateway.New(rpcAddr, !noBootstrap, filepath.Join(rivineDir, modules.GatewayDir))
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
	if strings.Contains(moduleString, "c") {
		i++
		fmt.Printf("(%d/%d) Loading consensus...\n", i, len(moduleString))
		cs, err = consensus.New(g, !noBootstrap, filepath.Join(rivineDir, modules.ConsensusDir))
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
	if strings.Contains(moduleString, "e") {
		i++
		fmt.Printf("(%d/%d) Loading explorer...\n", i, len(moduleString))
		e, err = explorer.New(cs, filepath.Join(rivineDir, modules.ExplorerDir))
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
	if strings.Contains(moduleString, "t") {
		i++
		fmt.Printf("(%d/%d) Loading transaction pool...\n", i, len(moduleString))
		tpool, err = transactionpool.New(cs, g, filepath.Join(rivineDir, modules.TransactionPoolDir))
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
	if strings.Contains(moduleString, "w") {
		i++
		fmt.Printf("(%d/%d) Loading wallet...\n", i, len(moduleString))
		w, err = wallet.New(cs, tpool, filepath.Join(rivineDir, modules.WalletDir))
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
	if strings.Contains(moduleString, "b") {
		i++
		fmt.Printf("(%d/%d) Loading block creator...\n", i, len(moduleString))
		b, err = blockcreator.New(cs, tpool, w, filepath.Join(rivineDir, modules.BlockCreatorDir))
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

	// Create the Sia API
	a := api.New(
		requiredUserAgent,
		apiPassword,
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
