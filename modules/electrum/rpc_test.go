package electrum

import (
	"encoding/json"
	"errors"
	"net"
	"path/filepath"
	"reflect"
	"sync"
	"testing"

	"github.com/rivine/rivine/build"
	"github.com/rivine/rivine/modules"
	"github.com/rivine/rivine/modules/consensus"
	"github.com/rivine/rivine/modules/explorer"
	"github.com/rivine/rivine/modules/gateway"
	"github.com/rivine/rivine/types"
)

type testConn struct {
	errorChan   chan error
	requestChan chan *BatchRequest
	stopChan    chan struct{}
	response    interface{}
	sendChan    chan interface{}
}

func (c *testConn) Close(wg *sync.WaitGroup) error {
	close(c.stopChan)
	close(c.requestChan)
	close(c.errorChan)
	close(c.sendChan)

	wg.Wait()
	return nil
}

func (c *testConn) GetError() <-chan error {
	return c.errorChan
}

func (c *testConn) GetRequest() <-chan *BatchRequest {
	return c.requestChan
}

func (c *testConn) Send(msg interface{}) error {
	c.response = msg
	c.sendChan <- msg
	return nil
}

func (c *testConn) RemoteAddr() net.Addr {
	return nil
}

func (c *testConn) IsClosed() <-chan struct{} {
	return c.stopChan
}

func TestElectrum_ServeRPC(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	testDir := build.TempDir(modules.ElectrumDir, "electrumtest")
	bcInfo := types.DefaultBlockchainInfo()
	chainCts := types.DefaultChainConstants()
	// Create the modules
	g, err := gateway.New("localhost:0", false, filepath.Join(testDir, modules.GatewayDir), bcInfo, chainCts, nil)
	if err != nil {
		t.Fatal("Failed to initialize gateway:", err)
	}

	cs, err := consensus.New(g, false, filepath.Join(testDir, modules.ConsensusDir), bcInfo, chainCts)
	if err != nil {
		t.Fatal("Failed to intitialize cs:", err)
	}

	explorer, err := explorer.New(cs, filepath.Join(testDir, modules.ExplorerDir), bcInfo, chainCts)
	if err != nil {
		t.Fatal("Failed to intitialize explorer:", err)
	}

	electrum, err := New(cs, nil, explorer, "", "", filepath.Join(testDir, modules.ElectrumDir), bcInfo, chainCts)
	if err != nil {
		t.Fatal("Failed to intitialize electrum:", err)
	}

	testConn := createTestConn()
	go electrum.ServeRPC(testConn)

	// Test pinging the connection
	expectedMessage := BatchResponse{isBatch: false, responses: []*Response{&Response{ID: 1, JSONRPC: jsonRPCVersion}}}
	testConn.requestChan <- &BatchRequest{isBatch: false, requests: []*Request{&Request{ID: 1, Method: "server.ping", JSONRPC: jsonRPCVersion}}}
	<-testConn.sendChan
	if !reflect.DeepEqual(testConn.response, expectedMessage) {
		t.Fatal("Electrum returned wrong response for server ping")
	}

	// Set version 1.0, no name
	params := json.RawMessage([]byte(`{"protocol_version": "1.0"}`))

	testConn.requestChan <- &BatchRequest{isBatch: false, requests: []*Request{&Request{ID: 2, Method: "server.version", JSONRPC: jsonRPCVersion, Params: &params}}}
	<-testConn.sendChan
	batchResponse, ok := testConn.response.(BatchResponse)
	if !ok {
		t.Fatal("Electrum  returned wrong response type")
	}
	if batchResponse.responses[0].Result == nil {
		t.Fatal("Electrum did not return a real response value for server version")
	}
	if batchResponse.responses[0].Error != nil {
		t.Fatal("Electrum returned an unexpected error for server version")
	}

	// subscribe to an address
	params = json.RawMessage([]byte(`{"address":"015a080a9259b9d4aaa550e2156f49b1a79a64c7ea463d810d4493e8242e6791584fbdac553e6f"}`))
	testConn.requestChan <- &BatchRequest{isBatch: false, requests: []*Request{&Request{ID: 3, Method: "blockchain.address.subscribe", JSONRPC: jsonRPCVersion, Params: &params}}}
	<-testConn.sendChan
	batchResponse, ok = testConn.response.(BatchResponse)
	if !ok {
		t.Fatal("Electrum  returned wrong response type")
	}
	if batchResponse.responses[0].Result == nil {
		t.Fatal("Electrum did not return a real response value for address subscription")
	}
	if batchResponse.responses[0].Error != nil {
		t.Fatal("Electrum returned an unexpected error for address subscription")
	}

	// check errors from the transport
	testConn.errorChan <- errors.New("failed to decode request")
	<-testConn.sendChan
	resp, ok := testConn.response.(*Response)
	if !ok {
		t.Fatal("Electrum  returned wrong response type")
	}
	if resp.Result != nil {
		t.Fatal("Electrum returned a response instead of an error")
	}
	if resp.Error != &ErrParse {
		t.Fatal("Electrum returned wrong error type for invalid request")
	}
	if resp.ID != nil {
		t.Fatal("Electrum returned unexpected ID in parse error")
	}

	electrum.Close()
}

func createTestConn() *testConn {
	return &testConn{
		errorChan:   make(chan error),
		requestChan: make(chan *BatchRequest),
		stopChan:    make(chan struct{}),
		sendChan:    make(chan interface{}),
	}
}
