package electrum

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"net/http"
	"sort"
	"sync"

	"github.com/rivine/rivine/build"
	"github.com/rivine/rivine/modules"
	"github.com/rivine/rivine/persist"
	"github.com/rivine/rivine/types"
)

// Electrum managas connections on which the electrum
// protocol is served
type Electrum struct {
	cs       modules.ConsensusSet
	tp       modules.TransactionPool
	explorer modules.Explorer

	mu          sync.Mutex
	connections map[*Client]chan<- *Update

	log      *persist.Logger
	bcInfo   types.BlockchainInfo
	chainCts types.ChainConstants

	availableVersions []ProtocolVersion

	// http server for websocket connections
	httpServer *http.Server
	// tcp listener for tcp connections
	tcpServer net.Listener

	// Make sure we can wait for all active connections to be closed when stopping
	activeConnections sync.WaitGroup
	stopChan          chan struct{}
}

// New creates a new Electrum instance using the provided consensus set, transactionpool,
// and explorer. The consensus set is a mandatory dependancy, the transactionpool and explorer are
// optional. wsAddress is the host:port to be used for the http server which will handle the websocket
// connection. If the string is empty, no http server is configured
//
// Listerners are not started yet, this is done through the Start method
func New(cs modules.ConsensusSet, tp modules.TransactionPool,
	explorer modules.Explorer, tcpAddress string, wsAddress string,
	persistDir string, bcInfo types.BlockchainInfo, chainCts types.ChainConstants) (*Electrum, error) {

	if cs == nil {
		return nil, errors.New("Consensus set is required for the Electrum module")
	}

	if explorer == nil {
		return nil, errors.New("Explorer is required for the Electrum module")
	}

	e := &Electrum{
		cs:       cs,
		tp:       tp,
		explorer: explorer,

		connections: make(map[*Client]chan<- *Update), // add update channel value

		bcInfo:   bcInfo,
		chainCts: chainCts,

		// Support only "v1.0.0"
		availableVersions: []ProtocolVersion{ProtocolVersion{1, 0, 0}},

		stopChan: make(chan struct{}),
	}

	if err := e.initLogger(persistDir); err != nil {
		return nil, errors.New("Failed to initialize electrum logger: " + err.Error())
	}

	// Create the http server for websocket connections
	var httpServer *http.Server
	if wsAddress != "" {
		mux := http.NewServeMux()
		mux.HandleFunc("/", e.handleWs)
		httpServer = &http.Server{Addr: wsAddress, Handler: mux}
	}
	e.httpServer = httpServer

	var tcpServer net.Listener
	var err error
	if tcpAddress != "" {
		tcpServer, err = net.Listen("tcp", tcpAddress)
		if err != nil {
			return nil, err
		}
	}
	e.tcpServer = tcpServer

	err = cs.ConsensusSetSubscribe(e, modules.ConsensusChangeRecent)

	return e, err
}

// AddressStatus generates the status for an address as per the
// Electrum protocol spec.
func (e *Electrum) AddressStatus(address types.UnlockHash) string {
	// Get confirmed transactions
	txns := e.explorer.UnlockHash(address)

	if len(txns) == 0 {
		return ""
	}
	// Get the heights
	// multiple txns could be at the same height

	txnLocation := make(map[int][]types.TransactionID)
	heights := make([]int, len(txns))
	for i, tx := range txns {
		_, height, _ := e.explorer.Transaction(tx)
		heights[i] = int(height)
		txnLocation[i] = append(txnLocation[i], tx)
	}
	// order heights
	sort.Ints(heights)

	// Generate the status string. The order is:
	// 	- Transactionpool transactions
	//  - Per block, sordted in height ascending order:
	// 		* miner payout
	// 		* transactions ordered by appearance in the block
	//
	// TODO: if a transactionpool transaction is depending on another unconfirmed
	// (transactionpool) transaction, its height must be -1.
	// Enfore deterministic ordering of transactionpool transactions in some way. Right now
	// we depend on the explorer module keeping the ordering deterministic. Note that this may
	// cause a discrepancy between multiple different servers, if they received the transactions
	// in a different order. Hence we should fix this somehow eventually
	var statusString string
	for _, height := range heights {
		if height == 0 && len(txnLocation[height]) > 0 {
			// first load the genesis block and sort the transactions from that block
			var genesisStatusString string
			genesis, exists := e.cs.BlockAtHeight(types.BlockHeight(height))
			if build.DEBUG && !exists {
				build.Critical("Genesis block does not exist")
			}
			for _, genesisTx := range genesis.Transactions {
				// Loop backward through the transactions so we can slice out the ones we find in genesis
				for i := len(txnLocation[height]) - 1; i >= 0; i-- {
					if txnLocation[height][i] == genesisTx.ID() {
						genesisStatusString += fmt.Sprintf("%v:%v", genesisTx.ID().String(), height)
						txnLocation[height] = append(txnLocation[height][:i], txnLocation[height][i+1:]...)
						continue
					}
				}
			}
			// now add transactions from the transactionpool to the status string
			for _, txid := range txnLocation[height] {
				statusString += fmt.Sprintf("%v:%v", txid.String(), height)
			}
			// finally add the status from the genesis block to the transaction ID
			statusString += genesisStatusString
			continue
		}
		// if there is more than 1 txn in the block, fetch the block and and iterate over the txns
		// to put them in the right order
		// height must be bigger than zero because tp transactions are reported at height 0
		if len(txnLocation[height]) > 1 {
			block, exists := e.cs.BlockAtHeight(types.BlockHeight(height))
			if build.DEBUG && !exists {
				build.Critical("Block does not exist in consensus set:", height)
			}
			// if the transaction id is the block id (miner payout), place it first in the ordering
			for _, txid := range txnLocation[height] {
				if txid == types.TransactionID(block.ID()) {
					statusString += fmt.Sprintf("%v:%v", txid.String(), height)
				}
			}
			// iterate through the remaining transactions in order
			for _, blocktx := range block.Transactions {
				for _, txid := range txnLocation[height] {
					if txid == blocktx.ID() {
						statusString += fmt.Sprintf("%v:%v", txid.String(), height)
						continue
					}
				}
			}
			continue
		}
		// there is only one transaction for this height
		statusString += fmt.Sprintf("%v:%v", txnLocation[height][0].String(), height)
	}
	statusHash := sha256.Sum256([]byte(statusString))

	return hex.EncodeToString(statusHash[:])
}

// Start all the servers, and accept incomming connections
func (e *Electrum) Start() {
	// Start the http server if one is configured
	if e.httpServer != nil {
		e.log.Println("Starting http server for websocket connections on", e.httpServer.Addr)
		go func() {
			if err := e.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				e.log.Critical("[ERROR] [HTTPSERVER]: error while running http server:", err)
			}
		}()
	}
	// Start listening for raw tcp connections
	if e.tcpServer != nil {
		e.log.Println("Start accepting tcp connections on", e.tcpServer.Addr())
		go func() {
			if err := e.listenTCP(); err != nil {
				// Error for listenTCP on closed connection is not exported, so we can't reliably match it.
				// As a workaround, check if the server is stopping, and if so, ignore the error
				select {
				case <-e.stopChan:
					return
				default:
				}
				e.log.Critical("[ERROR] [TCPSERVER]: error while listening for tcp connections:", err)
			}
		}()
	}
}

// Close closses the Electrum instance and every connection it is managing
func (e *Electrum) Close() error {
	e.cs.Unsubscribe(e)

	close(e.stopChan)

	if e.httpServer != nil {
		if err := e.httpServer.Shutdown(nil); err != nil {
			e.log.Println("[ERROR] [HTTPSERVER] Error while closing http server:", err)
			return err
		}
	}

	if e.tcpServer != nil {
		if err := e.tcpServer.Close(); err != nil {
			e.log.Println("[ERROR] [TCPSERVER]: Error while closing tcp listener:", err)
			return err
		}
	}

	e.activeConnections.Wait()

	if err := e.log.Close(); err != nil {
		fmt.Println("Failed to close electrum logger:", err)
	}
	return nil
}
