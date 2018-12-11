package main

import (
	"context"
	"encoding/json"
	"errors"
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

	"github.com/ethereum/go-ethereum/accounts/abi/bind"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"

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
	// ContractAddressString is the hex string address of the contract to monitor
	ContractAddressString = "0xd4466DAb724DD9Dc860e08E48a78D20e75361A25"
)

var (
	// RinkebyGenesisBlock is the genesis block used by the Rinkeby test network
	RinkebyGenesisBlock = core.DefaultRinkebyGenesisBlock()
	// ContractAddress is the address of the contract to monitor
	ContractAddress = common.HexToAddress(ContractAddressString)
	// OneToken is the exact value of one token
	OneToken = new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
)

var (
	genesisFlag = flag.String("genesis", "", "Genesis json file to seed the chain with")
	ethPortFlag = flag.Int("ethport", 30303, "Listener port for the devp2p connection")
	bootFlag    = flag.String("bootnodes", strings.Join(params.RinkebyBootnodes, ","), "Comma separated bootnode enode URLs to seed with") // default to rinkeby boot nodes
	netFlag     = flag.Uint64("network", RinkebyNetworkID, "Network ID to use for the Ethereum protocol")
	statsFlag   = flag.String("ethstats", "", "Ethstats network monitoring auth string")

	transferFlag = flag.String("transferTo", "", "Address in hex form to transfer 10 Tokens to, if not set then no transfer is done")

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

	if accPassFlag == nil || *accPassFlag == "" {
		log.Crit("Account password must be provided")
	}
	pass := *accPassFlag

	var acc accounts.Account
	// load account
	ks := keystore.NewKeyStore(filepath.Join(os.Getenv("HOME"), ".oracle-proto", "keys"), keystore.StandardScryptN, keystore.StandardScryptP)

	// If the accJSONFlag is provided try to import the account
	if accJSONFlag != nil && *accJSONFlag != "" {
		blob, err := ioutil.ReadFile(*accJSONFlag)
		if err != nil {
			log.Crit("Failed to read account key contents", "file", *accJSONFlag, "err", err)
		}
		// Import the account
		acc, err = ks.Import(blob, pass, pass)
		if err != nil {
			log.Crit("Failed to import faucet signer account", "err", err)
		}
	} else {
		// check if there are accounts already loaded
		if len(ks.Accounts()) == 0 {
			log.Crit("Failed to find any existing account")
		}
		acc = ks.Accounts()[0]
	}
	ks.Unlock(acc, pass)

	// Assemble and start the oracle light service
	oracle, err := newOracleProto(genesis, *ethPortFlag, enodes, *netFlag, *statsFlag, ks)
	if err != nil {
		log.Crit("Failed to start oracle", "err", err)
	}
	defer oracle.close()

	closeChan := make(chan struct{})

	// Do your business until we send ctrl-c
	var hold sync.WaitGroup
	hold.Add(1)
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, os.Interrupt)
	go func() {
		<-sc
		close(closeChan)
		hold.Done()
	}()

	go oracle.loop()
	go oracle.SubscribeTransfers(ContractAddress)

	if *transferFlag != "" {
		toAddr := common.HexToAddress(*transferFlag)
		go func() {
			// sleep 30 secs then transfer
			time.Sleep(time.Second * 30)

			if err := oracle.TransferFunds(ContractAddress, toAddr, new(big.Int).Mul(big.NewInt(10), OneToken)); err != nil {
				log.Warn("Failed to send tokens", "err", err)
				return
			}
			log.Warn("tranfered tokens successfully", "recipient", *transferFlag)
		}()

	}

	hold.Wait()
}

// oracleProto represents a prototype for an oracle, able to call
// contract methods and listen for contract events
type oracleProto struct {
	config *params.ChainConfig // Chain configurations for signing
	stack  *node.Node          // Ethereum protocol stack
	client *ethclient.Client   // Client connection to the Ethereum chain

	keystore *keystore.KeyStore // Keystore containing the signing info
	account  accounts.Account   // Account funding the oracle requests
	head     *types.Header      // Current head header of the oracle
	balance  *big.Int           // The current balance of the oracle (note: ethers only!)

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
		cfg.Ethash.DatasetDir = filepath.Join(os.Getenv("HOME"), ".oracle-proto", "ethash")
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
		config:   genesis.Config,
		stack:    stack,
		client:   client,
		keystore: ks,
		account:  ks.Accounts()[0],
	}, nil
}

// close terminates the Ethereum connection and tears down the oracle proto.
func (op *oracleProto) close() error {
	return op.stack.Stop()
}

// refresh attempts to retrieve the latest header from the chain and extract the
// associated oracle balance and nonce for connectivity caching.
func (op *oracleProto) refresh(head *types.Header) error {
	// Ensure a state update does not run for too long
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// If no header was specified, use the current chain head
	var err error
	if head == nil {
		if head, err = op.client.HeaderByNumber(ctx, nil); err != nil {
			return err
		}
	}
	// Retrieve the balance, nonce and gas price from the current head
	var (
		nonce   uint64
		price   *big.Int
		balance *big.Int
	)

	if price, err = op.client.SuggestGasPrice(ctx); err != nil {
		return err
	}
	if balance, err = op.client.BalanceAt(ctx, op.account.Address, head.Number); err != nil {
		return err
	}

	// Everything succeeded, update the cached stats
	op.lock.Lock()
	op.head, op.balance = head, balance
	op.price, op.nonce = price, nonce

	op.lock.Unlock()

	return nil
}

func (op *oracleProto) loop() {
	// channel to receive head updates from client on
	heads := make(chan *types.Header, 16)
	// subscribe to head upates
	sub, err := op.client.SubscribeNewHead(context.Background(), heads)
	if err != nil {
		log.Crit("Failed to subscribe to head events", "err", err)
	}
	defer sub.Unsubscribe()

	// channel so we can update the internal state from the heads
	update := make(chan *types.Header)

	go func() {
		for head := range update {
			// old heads should be ignored during a chain sync after some downtime
			if err := op.refresh(head); err != nil {
				log.Warn("Failed to update state", "block", head.Number, "err", err)
			}
			log.Info("Internal stats updated", "block", head.Number, "account balance", op.balance, "gas price", op.price, "nonce", op.nonce)
		}
	}()

	for head := range heads {
		select {
		// only process new head if another isn't being processed yet
		case update <- head:
		default:
		}
	}
}

// SubscribeTransfers subscribes to new Transfer events on the given contract. This call blocks
// and prints out info about any transfer as it happened
func (op *oracleProto) SubscribeTransfers(contractAddress common.Address) error {
	filter, err := NewTTFT20Filterer(contractAddress, op.client)
	if err != nil {
		return err
	}
	sink := make(chan *TTFT20Transfer)
	opts := &bind.WatchOpts{Context: context.Background(), Start: nil}
	sub, err := filter.WatchTransfer(opts, sink, nil, nil)
	if err != nil {
		return err
	}
	defer sub.Unsubscribe()
	for {
		select {
		case err = <-sub.Err():
			return err
		case transfer := <-sink:
			log.Info("Noticed transfer event", "from", transfer.From, "to", transfer.To, "amount", transfer.Tokens)
		}
	}
}

//
func (op *oracleProto) TransferFunds(contractAddress common.Address, recipient common.Address, amount *big.Int) error {
	if amount == nil {
		return errors.New("invalid amount")
	}

	tr, err := NewTTFT20Transactor(contractAddress, op.client)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	opts := &bind.TransactOpts{Context: ctx, From: op.account.Address, Signer: op.GetSignerFunc(), Value: nil, Nonce: nil, GasLimit: 0, GasPrice: nil}

	_, err = tr.Transfer(opts, recipient, amount)
	if err != nil {
		return err
	}

	return nil
}

func (op *oracleProto) GetSignerFunc() bind.SignerFn {
	return func(signer types.Signer, address common.Address, tx *types.Transaction) (*types.Transaction, error) {
		if address != op.account.Address {
			return nil, errors.New("not authorized to sign this account")
		}
		return op.keystore.SignTx(op.account, tx, big.NewInt(RinkebyNetworkID))

	}
}
