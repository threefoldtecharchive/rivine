package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/threefoldtech/rivine/build"
	"github.com/threefoldtech/rivine/crypto"
	"github.com/threefoldtech/rivine/modules"
	"github.com/threefoldtech/rivine/modules/blockcreator"
	"github.com/threefoldtech/rivine/modules/consensus"
	"github.com/threefoldtech/rivine/modules/explorer"
	"github.com/threefoldtech/rivine/modules/gateway"
	"github.com/threefoldtech/rivine/modules/transactionpool"
	"github.com/threefoldtech/rivine/modules/wallet"
	"github.com/threefoldtech/rivine/pkg/api"
	"github.com/threefoldtech/rivine/pkg/daemon"
	"github.com/threefoldtech/rivine/types"

	rivtypes "github.com/threefoldtech/rivine/cmd/rivinec/types"

	"github.com/threefoldtech/rivine/extensions/minting"
	mintingapi "github.com/threefoldtech/rivine/extensions/minting/api"

	"github.com/threefoldtech/rivine/extensions/authcointx"
	authcointxapi "github.com/threefoldtech/rivine/extensions/authcointx/api"
)

const (
	// maxConcurrentRPC is the maximum number of concurrent RPC calls allowed for a single daemon
	maxConcurrentRPC = 1
)

func runDaemon(cfg daemon.Config, networkCfg daemon.NetworkConfig, moduleIdentifiers daemon.ModuleIdentifierSet) error {
	// Print a startup message.
	fmt.Println("Loading...")
	loadStart := time.Now()

	if len(cfg.BootstrapPeers) > 0 {
		networkCfg.BootstrapPeers = cfg.BootstrapPeers
	}

	var (
		i             = 1
		modulesToLoad = moduleIdentifiers.Len()
	)
	printModuleIsLoading := func(name string) {
		fmt.Printf("Loading %s (%d/%d)...\r\n", name, i, modulesToLoad)
		i++
	}

	// create our server already, this way we can fail early if the API addr is already bound
	fmt.Println("Binding API Address and serving the API...")
	srv, err := daemon.NewHTTPServer(cfg.APIaddr)
	if err != nil {
		return err
	}
	servErrs := make(chan error)
	go func() {
		servErrs <- srv.Serve()
	}()

	// router to register all endpoints to
	router := httprouter.New()

	fmt.Println("Setting up root HTTP API handler...")

	// handle all our endpoints over a router,
	// which requires a user agent should one be configured
	srv.Handle("/", api.RequireUserAgentHandler(router, cfg.RequiredUserAgent))

	// consensus set
	var cs modules.ConsensusSet

	// register our special daemon HTTP handlers
	router.GET("/daemon/constants", func(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
		var pluginNames []string
		if cs != nil {
			pluginNames = cs.LoadedPlugins()
		}
		constants := modules.NewDaemonConstants(cfg.BlockchainInfo, networkCfg.Constants, pluginNames)
		api.WriteJSON(w, constants)
	})
	router.GET("/daemon/version", func(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
		api.WriteJSON(w, daemon.Version{
			ChainVersion:    cfg.BlockchainInfo.ChainVersion,
			ProtocolVersion: cfg.BlockchainInfo.ProtocolVersion,
		})
	})
	router.POST("/daemon/stop", func(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
		// can't write after we stop the server, so lie a bit.
		api.WriteSuccess(w)

		// need to flush the response before shutting down the server
		f, ok := w.(http.Flusher)
		if !ok {
			err := errors.New("Server does not support flushing")
			build.Severe(err)
		}
		f.Flush()

		if err := srv.Close(); err != nil {
			servErrs <- err
		}
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize the Rivine modules
	var g modules.Gateway
	if moduleIdentifiers.Contains(daemon.GatewayModule.Identifier()) {
		printModuleIsLoading("gateway")
		g, err = gateway.New(cfg.RPCaddr, !cfg.NoBootstrap, maxConcurrentRPC,
			filepath.Join(cfg.RootPersistentDir, modules.GatewayDir),
			cfg.BlockchainInfo, networkCfg.Constants, networkCfg.BootstrapPeers, cfg.VerboseLogging)
		if err != nil {
			cancel()
			return err
		}
		api.RegisterGatewayHTTPHandlers(router, g, cfg.APIPassword)
		defer func() {
			fmt.Println("Closing gateway...")
			err := g.Close()
			if err != nil {
				fmt.Println("Error during gateway shutdown:", err)
			}
		}()

	}
	var (
		mintingPlugin    *minting.Plugin
		authCoinTxPlugin *authcointx.Plugin
	)
	if moduleIdentifiers.Contains(daemon.ConsensusSetModule.Identifier()) {
		printModuleIsLoading("consensus")
		cs, err = consensus.New(g, !cfg.NoBootstrap,
			filepath.Join(cfg.RootPersistentDir, modules.ConsensusDir),
			cfg.BlockchainInfo, networkCfg.Constants, cfg.VerboseLogging, cfg.DebugConsensusDB)
		if err != nil {
			cancel()
			return err
		}
		api.RegisterConsensusHTTPHandlers(router, cs)
		defer func() {
			fmt.Println("Closing consensus set...")
			err := cs.Close()
			if err != nil {
				fmt.Println("Error during consensus set shutdown:", err)
			}
		}()
	}
	var tpool modules.TransactionPool
	if moduleIdentifiers.Contains(daemon.TransactionPoolModule.Identifier()) {
		printModuleIsLoading("transaction pool")
		tpool, err = transactionpool.New(cs, g,
			filepath.Join(cfg.RootPersistentDir, modules.TransactionPoolDir),
			cfg.BlockchainInfo, networkCfg.Constants, cfg.VerboseLogging)
		if err != nil {
			cancel()
			return err
		}
		api.RegisterTransactionPoolHTTPHandlers(router, cs, tpool, cfg.APIPassword)
		defer func() {
			fmt.Println("Closing transaction pool...")
			err := tpool.Close()
			if err != nil {
				fmt.Println("Error during transaction pool shutdown:", err)
			}
		}()
	}

	if cs != nil {
		rivCfg, err := newRivineNetworkConfig(cfg)
		if err != nil {
			cancel()
			return err
		}

		// create the minting extension plugin
		mintingPlugin = minting.NewMintingPlugin(
			rivCfg.GenesisMintCondition,
			rivtypes.TransactionVersionMinterDefinition,
			rivtypes.TransactionVersionCoinCreation,
			&minting.PluginOptions{
				RequireMinerFees:                  false,
				CoinDestructionTransactionVersion: rivtypes.TransactionVersionCoinDestruction,
			},
		)
		// add the HTTP handlers for the auth coin tx extension as well
		mintingapi.RegisterConsensusMintingHTTPHandlers(router, mintingPlugin)

		// create the auth coin tx plugin
		// > NOTE: this also overwrites the standard tx controllers!!!!
		authCoinTxPlugin = authcointx.NewPlugin(
			rivCfg.GenesisAuthCondition,
			rivtypes.TransactionVersionAuthAddressUpdate,
			rivtypes.TransactionVersionAuthConditionUpdate,
			&authcointx.PluginOpts{
				UnauthorizedCoinTransactionExceptionCallback: nil, // no callback required
				UnlockHashFilter: nil, // use defualt filter
			},
		)
		// add the HTTP handlers for the auth coin tx extension as well
		if tpool != nil {
			authcointxapi.RegisterConsensusAuthCoinHTTPHandlers(
				router, authCoinTxPlugin,
				tpool, rivtypes.TransactionVersionAuthConditionUpdate,
				rivtypes.TransactionVersionAuthAddressUpdate)
		} else {
			authcointxapi.RegisterConsensusAuthCoinHTTPHandlers(
				router, authCoinTxPlugin,
				nil, rivtypes.TransactionVersionAuthConditionUpdate,
				rivtypes.TransactionVersionAuthAddressUpdate)
		}

		// register the minting extension plugin
		err = cs.RegisterPlugin(ctx, "minting", mintingPlugin)
		if err != nil {
			servErrs <- fmt.Errorf("failed to register the minting extension: %v", err)
			err = mintingPlugin.Close() //make sure any resources are released
			if err != nil {
				fmt.Println("Error during closing of the mintingPlugin :", err)
			}
			cancel()
			return err
		}

		// register the AuthCoin extension plugin
		err = cs.RegisterPlugin(ctx, "authcointx", authCoinTxPlugin)
		if err != nil {
			servErrs <- fmt.Errorf("failed to register the auth coin tx extension: %v", err)
			err = authCoinTxPlugin.Close() //make sure any resources are released
			if err != nil {
				fmt.Println("Error during closing of the authCoinTxPlugin :", err)
			}
			cancel()
			return err
		}
	}

	var w modules.Wallet
	if moduleIdentifiers.Contains(daemon.WalletModule.Identifier()) {
		printModuleIsLoading("wallet")
		w, err = wallet.New(cs, tpool,
			filepath.Join(cfg.RootPersistentDir, modules.WalletDir),
			cfg.BlockchainInfo, networkCfg.Constants, cfg.VerboseLogging)
		if err != nil {
			cancel()
			return err
		}
		api.RegisterWalletHTTPHandlers(router, w, cfg.APIPassword)
		defer func() {
			fmt.Println("Closing wallet...")
			err := w.Close()
			if err != nil {
				fmt.Println("Error during wallet shutdown:", err)
			}
		}()

	}
	var b modules.BlockCreator
	if moduleIdentifiers.Contains(daemon.BlockCreatorModule.Identifier()) {
		printModuleIsLoading("block creator")
		b, err = blockcreator.New(cs, tpool, w,
			filepath.Join(cfg.RootPersistentDir, modules.BlockCreatorDir),
			cfg.BlockchainInfo, networkCfg.Constants, cfg.VerboseLogging)
		if err != nil {
			cancel()
			return err
		}
		// block creator has no API endpoints to register
		defer func() {
			fmt.Println("Closing block creator...")
			err := b.Close()
			if err != nil {
				fmt.Println("Error during block creator shutdown:", err)
			}
		}()
	}
	var e modules.Explorer
	if moduleIdentifiers.Contains(daemon.ExplorerModule.Identifier()) {
		printModuleIsLoading("creator")
		e, err = explorer.New(cs,
			filepath.Join(cfg.RootPersistentDir, modules.ExplorerDir),
			cfg.BlockchainInfo, networkCfg.Constants, cfg.VerboseLogging)
		if err != nil {
			cancel()
			return err
		}
		api.RegisterExplorerHTTPHandlers(router, cs, e, tpool)
		defer func() {
			fmt.Println("Closing explorer...")
			err := e.Close()
			if err != nil {
				fmt.Println("Error during explorer shutdown:", err)
			}
		}()

		mintingapi.RegisterExplorerMintingHTTPHandlers(router, mintingPlugin)
		if tpool != nil {
			authcointxapi.RegisterExplorerAuthCoinHTTPHandlers(
				router, authCoinTxPlugin,
				tpool, rivtypes.TransactionVersionAuthConditionUpdate,
				rivtypes.TransactionVersionAuthAddressUpdate)
		} else {
			authcointxapi.RegisterExplorerAuthCoinHTTPHandlers(
				router, authCoinTxPlugin,
				nil, rivtypes.TransactionVersionAuthConditionUpdate,
				rivtypes.TransactionVersionAuthAddressUpdate)
		}
	}

	// If there are any long running operations that need to happen first (e.g. for some extension code)
	// You can do that first before starting the cs syncing
	if cs != nil {
		cs.Start()
	}

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

	// return the first error which is returned
	return <-servErrs
}

type rivineNetworkConfig struct {
	GenesisMintCondition types.UnlockConditionProxy
	GenesisAuthCondition types.UnlockConditionProxy
}

func newRivineNetworkConfig(cfg daemon.Config) (rivineNetworkConfig, error) {
	switch cfg.BlockchainInfo.NetworkName {
	case "standard":
		condition := types.NewCondition(types.NewUnlockHashCondition(unlockHashFromHex("01b5e42056ef394f2ad9b511a61cec874d25bebe2095682dd37455cbafed4bec154e382a23f90e")))
		return rivineNetworkConfig{
			GenesisMintCondition: condition,
			GenesisAuthCondition: condition,
		}, nil
	case "testnet":
		condition := types.NewCondition(types.NewUnlockHashCondition(types.UnlockHash{
			Type: types.UnlockTypePubKey,
			Hash: crypto.Hash{214, 166, 197, 164, 29, 201, 53, 236, 106, 239, 10, 158, 127, 131, 20, 138, 63, 221, 230, 16, 98, 247, 32, 77, 210, 68, 116, 12, 241, 89, 27, 223},
		}))
		return rivineNetworkConfig{
			GenesisMintCondition: condition,
			GenesisAuthCondition: condition,
		}, nil
	case "devnet":
		// Seed for the address given below twice:
		// carbon boss inject cover mountain fetch fiber fit tornado cloth wing dinosaur proof joy intact fabric thumb rebel borrow poet chair network expire else
		condition := types.NewCondition(types.NewUnlockHashCondition(unlockHashFromHex("015a080a9259b9d4aaa550e2156f49b1a79a64c7ea463d810d4493e8242e6791584fbdac553e6f")))
		return rivineNetworkConfig{
			GenesisMintCondition: condition,
			GenesisAuthCondition: condition,
		}, nil
	default:
		// network isn't recognised
		return rivineNetworkConfig{}, fmt.Errorf(
			"Netork name %q not recognized", cfg.BlockchainInfo.NetworkName)
	}
}

func unlockHashFromHex(hstr string) (uh types.UnlockHash) {
	err := uh.LoadString(hstr)
	if err != nil {
		build.Critical(fmt.Sprintf("func unlockHashFromHex(%s) failed: %v", hstr, err))
	}
	return
}
