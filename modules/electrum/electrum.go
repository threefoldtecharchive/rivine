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

	log    *persist.Logger
	bcInfo types.BlockchainInfo

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
	persistDir string, bcInfo types.BlockchainInfo) (*Electrum, error) {

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

		bcInfo: bcInfo,

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

	// TODO: ORDER IN BLOCK + AVOID COLLISIONS

	txnLocation := make(map[int][]types.TransactionID)
	heights := make([]int, len(txns))
	for i, tx := range txns {
		_, height, _ := e.explorer.Transaction(tx)
		heights[i] = int(height)
		txnLocation[i] = append(txnLocation[i], tx)
	}
	// order heights
	sort.Ints(heights)

	var statusString string
	for _, height := range heights {
		for _, tx := range txnLocation[height] {
			statusString += fmt.Sprintf("%v:%v", tx.String(), height)
		}
	}
	statusHash := sha256.Sum256([]byte(statusString))

	return hex.EncodeToString(statusHash[:])
}

// Start all the servers, and accept incomming connections
func (e *Electrum) Start() {
	// Start the http server if one is configured
	if e.httpServer != nil {
		e.log.Println("Starting http server for websocket connections")
		go func() {
			if err := e.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				e.log.Critical("[ERROR] [HTTPSERVER]: error while running http server:", err)
			}
		}()
	}
	// Start listening for raw tcp connections
	if e.tcpServer != nil {
		e.log.Println("Start accepting tcp connections")
		go func() {
			if err := e.listenTCP(); err != nil {
				e.log.Critical("[ERROR] [TCPSERVER]: error while listening for tcp connections:", err)
			}
		}()
	}
}

// Close closses the Electrum instance and every connection it is managing
func (e *Electrum) Close() error {
	e.cs.Unsubscribe(e)

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

	close(e.stopChan)

	e.activeConnections.Wait()

	if err := e.log.Close(); err != nil {
		fmt.Println("Failed to close electrum logger:", err)
	}
	return nil
}
