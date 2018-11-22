package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/eth"
	"github.com/ethereum/go-ethereum/eth/downloader"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/ethstats"
	"github.com/ethereum/go-ethereum/les"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/discv5"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/p2p/nat"
	"github.com/ethereum/go-ethereum/params"
)

const (
	// RinkebyNetworkID is the network ID for the rinkeby network
	RinkebyNetworkID = 4
)

var (
	// RinkebyGenesisBlock is the genesis block used by the Rinkeby test network
	RinkebyGenesisBlock = core.DefaultRinkebyGenesisBlock()
)

var (
	genesisFlag = flag.String("genesis", "", "Genesis json file to seed the chain with")
	ethPortFlag = flag.Int("ethport", 30303, "Listener port for the devp2p connection")
	bootFlag    = flag.String("bootnodes", strings.Join(params.RinkebyBootnodes, ","), "Comma separated bootnode enode URLs to seed with") // default to rinkeby boot nodes
	netFlag     = flag.Uint64("network", RinkebyNetworkID, "Network ID to use for the Ethereum protocol")
	statsFlag   = flag.String("ethstats", "", "Ethstats network monitoring auth string")

	accJSONFlag = flag.String("account.json", "", "Key json file to fund user requests with")
	accPassFlag = flag.String("account.pass", "", "Decryption password to access oracle funds")

	logFlag = flag.Int("loglevel", 3, "Log level to use for Ethereum and the oracle")
)

var (
	ether = new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
)

func main() {
	// Parse the flags and set up the logger to print everything requested
	flag.Parse()
	log.Root().SetHandler(log.LvlFilterHandler(log.Lvl(*logFlag), log.StreamHandler(os.Stderr, log.TerminalFormat(true))))

	// Load and parse the genesis block requested by the user
	genesis := new(core.Genesis)
	if *genesisFlag == "" {
		genesis = RinkebyGenesisBlock
	} else {
		blob, err := ioutil.ReadFile(*genesisFlag)
		if err != nil {
			log.Crit("Failed to read genesis block contents", "genesis", *genesisFlag, "err", err)
		}
		if err = json.Unmarshal(blob, genesis); err != nil {
			log.Crit("Failed to parse genesis block json", "err", err)
		}
	}

	// Convert the bootnodes to internal enode representations
	var enodes []*discv5.Node
	for _, boot := range strings.Split(*bootFlag, ",") {
		if url, err := discv5.ParseNode(boot); err == nil {
			enodes = append(enodes, url)
		} else {
			log.Error("Failed to parse bootnode URL", "url", boot, "err", err)
		}
	}

	// Assemble and start the oracle light service
	oracle, err := newOracleProto(genesis, *ethPortFlag, enodes, *netFlag, *statsFlag, nil)
	if err != nil {
		log.Crit("Failed to start oracle", "err", err)
	}
	defer oracle.close()

	// Do your business until we send ctrl-c
	var hold sync.WaitGroup
	hold.Add(1)
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, os.Interrupt)
	go func() {
		<-sc
		hold.Done()
	}()

	hold.Wait()
}

// oracleProto represents a prototype for an oracle, able to call
// contract methods and listen for contract events
type oracleProto struct {
	config *params.ChainConfig // Chain configurations for signing
	stack  *node.Node          // Ethereum protocol stack
	client *ethclient.Client   // Client connection to the Ethereum chain

	head *types.Header // Current head header of the oracle

	nonce uint64   // Current pending nonce of the oracle
	price *big.Int // Current gas price to issue funds with

	lock sync.RWMutex // Lock protecting the oracle's internals
}

func newOracleProto(genesis *core.Genesis, port int, enodes []*discv5.Node, network uint64, stats string, ks *keystore.KeyStore) (*oracleProto, error) {
	// Assemble the raw devp2p protocol stack
	stack, err := node.New(&node.Config{
		Name:    "tfop", //threefold oracle proto
		Version: params.VersionWithMeta,
		DataDir: filepath.Join(os.Getenv("HOME"), ".oracle-proto"),
		P2P: p2p.Config{
			NAT:              nat.Any(),
			NoDiscovery:      true,
			DiscoveryV5:      true,
			ListenAddr:       fmt.Sprintf(":%d", port),
			MaxPeers:         25,
			BootstrapNodesV5: enodes,
		},
	})
	if err != nil {
		return nil, err
	}
	// Assemble the Ethereum light client protocol
	if err := stack.Register(func(ctx *node.ServiceContext) (node.Service, error) {
		cfg := eth.DefaultConfig
		cfg.SyncMode = downloader.LightSync
		cfg.NetworkId = network
		cfg.Genesis = genesis
		return les.New(ctx, &cfg)
	}); err != nil {
		return nil, err
	}
	// Assemble the ethstats monitoring and reporting service'
	if stats != "" {
		if err := stack.Register(func(ctx *node.ServiceContext) (node.Service, error) {
			var serv *les.LightEthereum
			ctx.Service(&serv)
			return ethstats.New(stats, nil, serv)
		}); err != nil {
			return nil, err
		}
	}
	// Boot up the client and ensure it connects to bootnodes
	if err := stack.Start(); err != nil {
		return nil, err
	}
	for _, boot := range enodes {
		old, err := enode.ParseV4(boot.String())
		if err != nil {
			stack.Server().AddPeer(old)
		}
	}
	// Attach to the client and retrieve and interesting metadatas
	api, err := stack.Attach()
	if err != nil {
		stack.Stop()
		return nil, err
	}
	client := ethclient.NewClient(api)

	return &oracleProto{
		config: genesis.Config,
		stack:  stack,
		client: client,
	}, nil
}

// close terminates the Ethereum connection and tears down the oracle proto.
func (f *oracleProto) close() error {
	return f.stack.Stop()
}

// refresh attempts to retrieve the latest header from the chain and extract the
// associated oracle balance and nonce for connectivity caching.
func (f *oracleProto) refresh(head *types.Header) error {
	// Ensure a state update does not run for too long
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// If no header was specified, use the current chain head
	var err error
	if head == nil {
		if head, err = f.client.HeaderByNumber(ctx, nil); err != nil {
			return err
		}
	}
	// Retrieve the balance, nonce and gas price from the current head
	var (
		nonce uint64
		price *big.Int
	)

	if price, err = f.client.SuggestGasPrice(ctx); err != nil {
		return err
	}
	// Everything succeeded, update the cached
	f.lock.Lock()
	f.head = head
	f.price, f.nonce = price, nonce

	f.lock.Unlock()

	return nil
}
