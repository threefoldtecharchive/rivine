package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/threefoldtech/rivine/modules"
	"github.com/threefoldtech/rivine/modules/blockcreator"
	"github.com/threefoldtech/rivine/modules/consensus"
	"github.com/threefoldtech/rivine/modules/explorer"
	"github.com/threefoldtech/rivine/modules/gateway"
	"github.com/threefoldtech/rivine/modules/transactionpool"
	"github.com/threefoldtech/rivine/modules/wallet"
	"github.com/threefoldtech/rivine/pkg/api"
	"github.com/threefoldtech/rivine/pkg/daemon"
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

	// Initialize the Rivine modules
	var g modules.Gateway
	if moduleIdentifiers.Contains(daemon.GatewayModule.Identifier()) {
		printModuleIsLoading("gateway")
		g, err = gateway.New(cfg.RPCaddr, !cfg.NoBootstrap,
			filepath.Join(cfg.RootPersistentDir, modules.GatewayDir),
			cfg.BlockchainInfo, networkCfg.Constants, networkCfg.BootstrapPeers, cfg.VerboseLogging)
		if err != nil {
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
	var cs modules.ConsensusSet
	if moduleIdentifiers.Contains(daemon.ConsensusSetModule.Identifier()) {
		printModuleIsLoading("consensus")
		cs, err = consensus.New(g, !cfg.NoBootstrap,
			filepath.Join(cfg.RootPersistentDir, modules.ConsensusDir),
			cfg.BlockchainInfo, networkCfg.Constants, cfg.VerboseLogging)
		if err != nil {
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
	var w modules.Wallet
	if moduleIdentifiers.Contains(daemon.WalletModule.Identifier()) {
		printModuleIsLoading("wallet")
		w, err = wallet.New(cs, tpool,
			filepath.Join(cfg.RootPersistentDir, modules.WalletDir),
			cfg.BlockchainInfo, networkCfg.Constants, cfg.VerboseLogging)
		if err != nil {
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
			cfg.BlockchainInfo, networkCfg.Constants)
		if err != nil {
			return err
		}
		if cfg.HostExplorer {
			explorerRouter := httprouter.New()
			api.RegisterExplorerHTTPHandlers(explorerRouter, cs, e, tpool)

			// After registering endpoints start explorer frontend
			go func() {
				err = e.ServeFrontend(cfg.StagingCA, cfg.CaddyDomains, cfg.CaddyEmail, explorerRouter)
				if err != nil {
					servErrs <- err
				}
			}()
		} else {
			api.RegisterExplorerHTTPHandlers(router, cs, e, tpool)
		}

		defer func() {
			fmt.Println("Closing explorer...")
			err := e.Close()
			if err != nil {
				fmt.Println("Error during explorer shutdown:", err)
			}
		}()
	}

	fmt.Println("Setting up root HTTP API handler...")

	// register our special daemon HTTP handlers
	router.GET("/daemon/constants", func(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
		constants := modules.NewDaemonConstants(cfg.BlockchainInfo, networkCfg.Constants)
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
			panic("Server does not support flushing")
		}
		f.Flush()

		if err := srv.Close(); err != nil {
			servErrs <- err
		}
	})

	// handle all our endpoints over a router,
	// which requires a user agent should one be configured
	srv.Handle("/", api.RequireUserAgentHandler(router, cfg.RequiredUserAgent))

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
