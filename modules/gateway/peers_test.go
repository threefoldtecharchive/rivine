package gateway

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/NebulousLabs/fastrand"
	"github.com/rivine/rivine/build"
	"github.com/rivine/rivine/modules"
	"github.com/rivine/rivine/types"
)

// dummyConn implements the net.Conn interface, but does not carry any actual
// data. It is passed to muxado, because passing nil results in segfaults.
type dummyConn struct {
	net.Conn
}

func (dc *dummyConn) Read(p []byte) (int, error)       { return len(p), nil }
func (dc *dummyConn) Write(p []byte) (int, error)      { return len(p), nil }
func (dc *dummyConn) Close() error                     { return nil }
func (dc *dummyConn) SetReadDeadline(time.Time) error  { return nil }
func (dc *dummyConn) SetWriteDeadline(time.Time) error { return nil }

// TestAddPeer tries adding a peer to the gateway.
func TestAddPeer(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	t.Parallel()
	g := newTestingGateway(t)
	defer g.Close()

	g.mu.Lock()
	defer g.mu.Unlock()
	g.addPeer(&peer{
		Peer: modules.Peer{
			NetAddress: "foo.com:123",
		},
		sess: newSmuxClient(new(dummyConn)),
	})
	if len(g.peers) != 1 {
		t.Fatal("gateway did not add peer")
	}
}

// TestAcceptPeer tests that acceptPeer does't kick outbound or local peers.
func TestAcceptPeer(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	t.Parallel()
	g := newTestingGateway(t)
	defer g.Close()
	g.mu.Lock()
	defer g.mu.Unlock()

	// Add only unkickable peers.
	var unkickablePeers []*peer
	for i := 0; i < fullyConnectedThreshold+1; i++ {
		addr := modules.NetAddress(fmt.Sprintf("1.2.3.%d", i))
		p := &peer{
			Peer: modules.Peer{
				NetAddress: addr,
				Inbound:    false,
				Local:      false,
			},
			sess: newSmuxClient(new(dummyConn)),
		}
		unkickablePeers = append(unkickablePeers, p)
	}
	for i := 0; i < fullyConnectedThreshold+1; i++ {
		addr := modules.NetAddress(fmt.Sprintf("127.0.0.1:%d", i))
		p := &peer{
			Peer: modules.Peer{
				NetAddress: addr,
				Inbound:    true,
				Local:      true,
			},
			sess: newSmuxClient(new(dummyConn)),
		}
		unkickablePeers = append(unkickablePeers, p)
	}
	for _, p := range unkickablePeers {
		g.addPeer(p)
	}

	// Test that accepting another peer doesn't kick any of the peers.
	g.acceptPeer(&peer{
		Peer: modules.Peer{
			NetAddress: "9.9.9.9",
			Inbound:    true,
		},
		sess: newSmuxClient(new(dummyConn)),
	})
	for _, p := range unkickablePeers {
		if _, exists := g.peers[p.NetAddress]; !exists {
			t.Error("accept peer kicked an outbound or local peer")
		}
	}

	// Add a kickable peer.
	g.addPeer(&peer{
		Peer: modules.Peer{
			NetAddress: "9.9.9.9",
			Inbound:    true,
		},
		sess: newSmuxClient(new(dummyConn)),
	})
	// Test that accepting a local peer will kick a kickable peer.
	g.acceptPeer(&peer{
		Peer: modules.Peer{
			NetAddress: "127.0.0.1:99",
			Inbound:    true,
			Local:      true,
		},
		sess: newSmuxClient(new(dummyConn)),
	})
	if _, exists := g.peers["9.9.9.9"]; exists {
		t.Error("acceptPeer didn't kick a peer to make room for a local peer")
	}
}

// TestRandomInbountPeer checks that randomOutboundPeer returns the correct
// peer.
func TestRandomOutboundPeer(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	t.Parallel()
	g := newTestingGateway(t)
	defer g.Close()
	g.mu.Lock()
	defer g.mu.Unlock()

	_, err := g.randomOutboundPeer()
	if err != errNoPeers {
		t.Fatal("expected errNoPeers, got", err)
	}

	g.addPeer(&peer{
		Peer: modules.Peer{
			NetAddress: "foo.com:123",
			Inbound:    false,
		},
		sess: newSmuxClient(new(dummyConn)),
	})
	if len(g.peers) != 1 {
		t.Fatal("gateway did not add peer")
	}
	addr, err := g.randomOutboundPeer()
	if err != nil || addr != "foo.com:123" {
		t.Fatal("gateway did not select random peer")
	}
}

// TestListen is a general test probing the connection listener.
func TestListen(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	t.Parallel()
	g := newTestingGateway(t)
	defer g.Close()

	// compliant connect with old version
	conn, err := net.Dial("tcp", string(g.Address()))
	if err != nil {
		t.Fatal("dial failed:", err)
	}
	addr := modules.NetAddress(conn.LocalAddr().String())
	ack, err := g.connectHandshake(conn, build.NewVersion(0, 1, 0), gatewayID{}, g.myAddr, true)
	if err != errPeerRejectedConn {
		t.Fatal(err)
	}

	for i := 0; i < 10; i++ {
		g.mu.RLock()
		_, ok := g.peers[addr]
		g.mu.RUnlock()
		if ok {
			t.Fatal("gateway should not have added an old peer")
		}
		time.Sleep(20 * time.Millisecond)
	}

	// compliant connect with fake netAddress
	conn, err = net.Dial("tcp", string(g.Address()))
	if err != nil {
		t.Fatal("dial failed:", err)
	}
	ack, err = g.connectHandshake(conn, build.Version, gatewayID{}, "fake", true)
	if err != errPeerRejectedConn {
		t.Fatal(err)
	}

	// a simple 'conn.Close' would not obey the muxado disconnect protocol
	newSmuxClient(conn).Close()

	// compliant connect
	conn, err = net.Dial("tcp", string(g.Address()))
	if err != nil {
		t.Fatal("dial failed:", err)
	}
	addr = modules.NetAddress(conn.LocalAddr().String())
	ack, err = g.connectHandshake(conn, build.Version, gatewayID{}, addr, true)
	if err != nil {
		t.Fatal(err)
	}
	if ack.Version.Compare(build.Version) != 0 {
		t.Fatal("gateway should have given ack")
	}

	// g should add the peer
	err = build.Retry(50, 100*time.Millisecond, func() error {
		g.mu.RLock()
		_, ok := g.peers[addr]
		g.mu.RUnlock()
		if !ok {
			return errors.New("g should have added the peer")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	// Disconnect. Now that connection has been established, need to shutdown
	// via the stream multiplexer.
	newSmuxClient(conn).Close()

	// g should remove the peer
	err = build.Retry(50, 100*time.Millisecond, func() error {
		g.mu.RLock()
		_, ok := g.peers[addr]
		g.mu.RUnlock()
		if ok {
			return errors.New("g should have removed the peer")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	// uncompliant connect
	conn, err = net.Dial("tcp", string(g.Address()))
	if err != nil {
		t.Fatal("dial failed:", err)
	}
	if _, err := conn.Write([]byte("missing length prefix")); err != nil {
		t.Fatal("couldn't write malformed header")
	}
	// g should have closed the connection
	if n, err := conn.Write([]byte("closed")); err != nil && n > 0 {
		t.Error("write succeeded after closed connection")
	}
}

// TestConnect verifies that connecting peers will add peer relationships to
// the gateway, and that certain edge cases are properly handled.
func TestConnect(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	t.Parallel()
	// create bootstrap peer
	bootstrap := newNamedTestingGateway(t, "1")
	defer bootstrap.Close()

	// give it a node
	bootstrap.mu.Lock()
	bootstrap.addNode(dummyNode)
	bootstrap.mu.Unlock()

	// create peer who will connect to bootstrap
	g := newNamedTestingGateway(t, "2")
	defer g.Close()

	// first simulate a "bad" connect, where bootstrap won't share its nodes
	bootstrap.mu.Lock()
	bootstrap.handlers[handlerName("ShareNodes")] = func(modules.PeerConn) error {
		return nil
	}
	bootstrap.mu.Unlock()
	// connect
	err := g.Connect(bootstrap.Address())
	if err != nil {
		t.Fatal(err)
	}
	// g should not have the node
	if g.removeNode(dummyNode) == nil {
		t.Fatal("bootstrapper should not have received dummyNode:", g.nodes)
	}

	// split 'em up
	g.Disconnect(bootstrap.Address())
	bootstrap.Disconnect(g.Address())

	// now restore the correct ShareNodes RPC and try again
	bootstrap.mu.Lock()
	bootstrap.handlers[handlerName("ShareNodes")] = bootstrap.shareNodes
	bootstrap.mu.Unlock()
	err = g.Connect(bootstrap.Address())
	if err != nil {
		t.Fatal(err)
	}
	// g should have the node
	time.Sleep(200 * time.Millisecond)
	g.mu.RLock()
	if _, ok := g.nodes[dummyNode]; !ok {
		g.mu.RUnlock() // Needed to prevent a deadlock if this error condition is reached.
		t.Fatal("bootstrapper should have received dummyNode:", g.nodes)
	}
	g.mu.RUnlock()
}

// TestConnectRejectsInvalidAddrs tests that Connect only connects to valid IP
// addresses.
func TestConnectRejectsInvalidAddrs(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	t.Parallel()
	g := newNamedTestingGateway(t, "1")
	defer g.Close()

	g2 := newNamedTestingGateway(t, "2")
	defer g2.Close()

	_, g2Port, err := net.SplitHostPort(string(g2.Address()))
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		addr    modules.NetAddress
		wantErr bool
		msg     string
	}{
		{
			addr:    "127.0.0.1:123",
			wantErr: true,
			msg:     "Connect should reject unreachable addresses",
		},
		{
			addr:    "111.111.111.111:0",
			wantErr: true,
			msg:     "Connect should reject invalid NetAddresses",
		},
		{
			addr:    modules.NetAddress(net.JoinHostPort("localhost", g2Port)),
			wantErr: true,
			msg:     "Connect should reject non-IP addresses",
		},
		{
			addr: g2.Address(),
			msg:  "Connect failed to connect to another gateway",
		},
		{
			addr:    g2.Address(),
			wantErr: true,
			msg:     "Connect should reject an address it's already connected to",
		},
	}
	for _, tt := range tests {
		err := g.Connect(tt.addr)
		if tt.wantErr != (err != nil) {
			t.Errorf("%v, wantErr: %v, err: %v", tt.msg, tt.wantErr, err)
		}
	}
}

// TestConnectRejectsVersions tests that Gateway.Connect only accepts peers
// with sufficient and valid versions.
func TestConnectRejectsVersions(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	cts := types.DefaultChainConstants()
	g := newTestingGateway(t)
	defer g.Close()
	// Setup a listener that mocks Gateway.acceptConn, but sends the
	// version sent over mockVersionChan instead of build.Version.
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()

	tests := []struct {
		version         build.ProtocolVersion
		errWant         error
		localErrWant    error
		msg             string
		versionRequired string
		genesisID       types.BlockID
		uniqueID        gatewayID
	}{
		// Test that Connect fails when the remote peer's version is < 1.0.0 (0).
		{
			version: build.NewVersion(0, 0, 0),
			errWant: insufficientVersionError("0.0.0"),
			msg:     "Connect should fail when the remote peer's version is 0.0.0",
		},
		// Test that Connect /could/ succeed when the remote peer's version is = minAcceptableVersion
		{
			version:   minAcceptableVersion,
			msg:       "Connect should succeed when the remote peer's versionHeader checks out",
			uniqueID:  func() (id gatewayID) { fastrand.Read(id[:]); return }(),
			genesisID: cts.GenesisBlockID(),
		},
		{
			version:      minAcceptableVersion,
			msg:          "Connect should not succeed when peer is connecting to itself",
			uniqueID:     g.id,
			genesisID:    cts.GenesisBlockID(),
			errWant:      errOurAddress,
			localErrWant: errOurAddress,
		},
		// Test that Connect /could/ succeed when the remote peer's version is = Gateway NetAddress Update Version
		{
			version:   handshakNetAddressUpgrade,
			msg:       "Connect should succeed when the remote peer's versionHeader checks out",
			uniqueID:  func() (id gatewayID) { fastrand.Read(id[:]); return }(),
			genesisID: cts.GenesisBlockID(),
		},
		{
			version:      handshakNetAddressUpgrade,
			msg:          "Connect should not succeed when peer is connecting to itself",
			uniqueID:     g.id,
			genesisID:    cts.GenesisBlockID(),
			errWant:      errOurAddress,
			localErrWant: errOurAddress,
		},
		// Test that Connect /could/ succeed when the remote peer's version is = current version
		{
			version:   build.Version,
			msg:       "Connect should succeed when the remote peer's versionHeader checks out",
			uniqueID:  func() (id gatewayID) { fastrand.Read(id[:]); return }(),
			genesisID: cts.GenesisBlockID(),
		},
		{
			version:      build.Version,
			msg:          "Connect should not succeed when peer is connecting to itself",
			uniqueID:     g.id,
			genesisID:    cts.GenesisBlockID(),
			errWant:      errOurAddress,
			localErrWant: errOurAddress,
		},
	}
	for testIndex, tt := range tests {
		doneChan := make(chan struct{})
		go func() {
			defer close(doneChan)
			conn, err := listener.Accept()
			if err != nil {
				panic(fmt.Sprintf("test #%d failed: %s", testIndex, err))
			}
			remoteInfo, err := g.acceptConnHandshake(conn, tt.version, tt.uniqueID)
			if tt.localErrWant != nil && err != tt.localErrWant {
				panic(fmt.Sprintf("test #%d failed: %s", testIndex, err))
			} else if err == nil && build.Version.Compare(remoteInfo.Version) != 0 {
				panic(fmt.Sprintf("test #%d failed: %q != %q",
					testIndex, build.Version.String(), remoteInfo.Version.String()))
			} else if err != nil && tt.errWant == nil {
				panic(fmt.Sprintf("test #%d failed: %q != %q",
					testIndex, build.Version.String(), remoteInfo.Version.String()))
			}
		}()
		err = g.Connect(modules.NetAddress(listener.Addr().String()))
		if err != tt.errWant {
			t.Fatalf("test #%d failed: expected Connect to error with '%v', but got '%v': %s", testIndex, tt.errWant, err, tt.msg)
		}
		<-doneChan
		g.Disconnect(modules.NetAddress(listener.Addr().String()))
	}
}

// TestDisconnect checks that calls to gateway.Disconnect correctly disconnect
// and remove peers from the gateway.
func TestDisconnect(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	t.Parallel()
	g := newTestingGateway(t)
	defer g.Close()

	if err := g.Disconnect("bar.com:123"); err == nil {
		t.Fatal("disconnect removed unconnected peer")
	}

	// dummy listener to accept connection
	l, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatal("couldn't start listener:", err)
	}
	go func() {
		_, err := l.Accept()
		if err != nil {
			panic(err)
		}
	}()
	// skip standard connection protocol
	conn, err := net.Dial("tcp", l.Addr().String())
	if err != nil {
		t.Fatal("dial failed:", err)
	}
	g.mu.Lock()
	g.addPeer(&peer{
		Peer: modules.Peer{
			NetAddress: "foo.com:123",
		},
		sess: newSmuxClient(conn),
	})
	g.mu.Unlock()
	if err := g.Disconnect("foo.com:123"); err != nil {
		t.Fatal("disconnect failed:", err)
	}
}

// TestPeerManager checks that the peer manager is properly spacing out peer
// connection requests.
func TestPeerManager(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	t.Parallel()
	g1 := newNamedTestingGateway(t, "1")
	defer g1.Close()

	// create a valid node to connect to
	g2 := newNamedTestingGateway(t, "2")
	defer g2.Close()

	// g1's node list should only contain g2
	g1.mu.Lock()
	g1.nodes = map[modules.NetAddress]*node{}
	g1.nodes[g2.Address()] = &node{NetAddress: g2.Address()}
	g1.mu.Unlock()

	// when peerManager wakes up, it should connect to g2.
	time.Sleep(time.Second + noNodesDelay)

	g1.mu.RLock()
	defer g1.mu.RUnlock()
	if len(g1.peers) != 1 || g1.peers[g2.Address()] == nil {
		t.Fatal("gateway did not connect to g2:", g1.peers)
	}
}

// TestOverloadedBootstrap creates a bunch of gateways and connects all of them
// to the first gateway, the bootstrap gateway. More gateways will be created
// than is allowed by the bootstrap for the total number of connections. After
// waiting, all peers should eventually get to the full number of outbound
// peers.
func TestOverloadedBootstrap(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	t.Parallel()

	// Create fullyConnectedThreshold*2 peers and connect them all to only the
	// first node.
	var gs []*Gateway
	for i := 0; i < fullyConnectedThreshold*2; i++ {
		gs = append(gs, newNamedTestingGateway(t, strconv.Itoa(i)))
		// Connect this gateway to the first gateway.
		if i == 0 {
			continue
		}
		err := gs[i].Connect(gs[0].myAddr)
		for j := 0; j < 100 && err != nil; j++ {
			time.Sleep(time.Millisecond * 250)
			err = gs[i].Connect(gs[0].myAddr)
		}
		if err != nil {
			panic(err)
		}
	}

	// Spin until all gateways have a complete number of outbound peers.
	success := false
	for i := 0; i < 100; i++ {
		success = true
		for _, g := range gs {
			outboundPeers := 0
			g.mu.RLock()
			for _, p := range g.peers {
				if !p.Inbound {
					outboundPeers++
				}
			}
			g.mu.RUnlock()

			if outboundPeers < wellConnectedThreshold {
				success = false
				break
			}
		}
		if !success {
			time.Sleep(time.Second)
		}
	}
	if !success {
		for i, g := range gs {
			outboundPeers := 0
			g.mu.RLock()
			for _, p := range g.peers {
				if !p.Inbound {
					outboundPeers++
				}
			}
			g.mu.RUnlock()
			t.Log("Gateway", i, ":", outboundPeers)
		}
		t.Fatal("after 100 seconds not all gateways able to become well connected")
	}

	// Randomly close many of the peers. For many peers, this should put them
	// below the well connected threshold, but there are still enough nodes on
	// the network that no partitions should occur.
	var newGS []*Gateway
	for _, i := range fastrand.Perm(len(gs)) {
		newGS = append(newGS, gs[i])
	}
	cutSize := len(newGS) / 4
	// Close the first many of the now-randomly-sorted gateways.
	for _, g := range newGS[:cutSize] {
		err := g.Close()
		if err != nil {
			t.Fatal(err)
		}
	}
	// Set 'gs' equal to the remaining gateways.
	gs = newGS[cutSize:]

	// Spin until all gateways have a complete number of outbound peers. The
	// test can fail if there are network partitions, however not a huge
	// magnitude of nodes are being removed, and they all started with 4
	// connections. A partition is unlikely.
	success = false
	for i := 0; i < 100; i++ {
		success = true
		for _, g := range gs {
			outboundPeers := 0
			g.mu.RLock()
			for _, p := range g.peers {
				if !p.Inbound {
					outboundPeers++
				}
			}
			g.mu.RUnlock()

			if outboundPeers < wellConnectedThreshold {
				success = false
				break
			}
		}
		if !success {
			time.Sleep(time.Second)
		}
	}
	if !success {
		t.Fatal("after 100 seconds not all gateways able to become well connected")
	}

	// Close all remaining gateways.
	for _, g := range gs {
		err := g.Close()
		if err != nil {
			t.Error(err)
		}
	}
}

// TestPeerManagerPriority tests that the peer manager will prioritize
// connecting to previous outbound peers before inbound peers.
func TestPeerManagerPriority(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	t.Parallel()

	g1 := newNamedTestingGateway(t, "1")
	defer g1.Close()
	g2 := newNamedTestingGateway(t, "2")
	defer g2.Close()
	g3 := newNamedTestingGateway(t, "3")
	defer g3.Close()

	// Connect g1 to g2. This will cause g2 to be saved as an outbound peer in
	// g1's node list.
	if err := g1.Connect(g2.Address()); err != nil {
		t.Fatal(err)
	}
	// Connect g3 to g1. This will cause g3 to be added to g1's node list, but
	// not as an outbound peer.
	if err := g3.Connect(g1.Address()); err != nil {
		t.Fatal(err)
	}

	// Spin until the connections succeeded.
	for i := 0; i < 50; i++ {
		g1.mu.RLock()
		_, exists2 := g1.nodes[g2.Address()]
		_, exists3 := g1.nodes[g3.Address()]
		g1.mu.RUnlock()
		if exists2 && exists3 {
			break
		}
		time.Sleep(time.Millisecond * 100)
	}
	g1.mu.RLock()
	peer2, exists2 := g1.nodes[g2.Address()]
	peer3, exists3 := g1.nodes[g3.Address()]
	g1.mu.RUnlock()
	if !exists2 {
		t.Fatal("peer 2 not in gateway")
	}
	if !exists3 {
		t.Fatal("peer 3 not found") // ERRORS
	}
	// Verify assumptions about node list.
	g1.mu.RLock()
	g2isOutbound := peer2.WasOutboundPeer
	g3isOutbound := peer3.WasOutboundPeer
	g1.mu.RUnlock()
	if !g2isOutbound {
		t.Fatal("g2 should be an outbound node")
	}
	if g3isOutbound {
		t.Fatal("g3 should not be an outbound node")
	}

	// Disconnect everyone.
	g2.Disconnect(g1.Address())
	g3.Disconnect(g1.Address())

	// Shutdown g1.
	err := g1.Close()
	if err != nil {
		t.Fatal(err)
	}

	// Restart g1. It should immediately reconnect to g2, and then g3 after a
	// delay.
	g1, err = New(string(g1.myAddr), false, g1.persistDir,
		types.DefaultBlockchainInfo(), types.DefaultChainConstants(), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer g1.Close()

	// Wait until g1 connects to g2.
	for i := 0; i < 100; i++ {
		if peers := g1.Peers(); len(peers) == 0 {
			time.Sleep(10 * time.Millisecond)
		} else if len(peers) == 1 && peers[0].NetAddress == g2.Address() {
			break
		} else {
			t.Fatal("something wrong with the peer list:", peers)
		}
	}
	// Wait until g1 connects to g3.
	for i := 0; i < 100; i++ {
		if peers := g1.Peers(); len(peers) == 1 {
			time.Sleep(10 * time.Millisecond)
		} else if len(peers) == 2 {
			break
		} else {
			t.Fatal("something wrong with the peer list:", peers)
		}
	}
}

// TestPeerManagerOutboundSave sets up an island of nodes and checks that they
// can all connect to eachother, and that the all add eachother as
// 'WasOutboundPeer'.
func TestPeerManagerOutboundSave(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	t.Parallel()

	// Create enough gateways so that every gateway should automatically end up
	// with every other gateway as an outbound peer.
	var gs []*Gateway
	for i := 0; i < wellConnectedThreshold+1; i++ {
		gs = append(gs, newNamedTestingGateway(t, strconv.Itoa(i)))
	}
	// Connect g1 to each peer. This should be enough that every peer eventually
	// has the full set of outbound peers.
	for _, g := range gs[1:] {
		if err := gs[0].Connect(g.Address()); err != nil {
			t.Fatal(err)
		}
	}

	// Block until every peer has wellConnectedThreshold outbound peers.
	err := build.Retry(100, time.Millisecond*200, func() error {
		for _, g := range gs {
			var outboundNodes, outboundPeers int
			g.mu.RLock()
			for _, node := range g.nodes {
				if node.WasOutboundPeer {
					outboundNodes++
				}
			}
			for _, peer := range g.peers {
				if !peer.Inbound {
					outboundPeers++
				}
			}
			g.mu.RUnlock()
			if outboundNodes < wellConnectedThreshold {
				return errors.New("not enough outbound nodes: " + strconv.Itoa(outboundNodes))
			}
			if outboundPeers < wellConnectedThreshold {
				return errors.New("not enough outbound peers: " + strconv.Itoa(outboundPeers))
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

// TestBuildPeerManagerNodeList tests the buildPeerManagerNodeList method.
func TestBuildPeerManagerNodeList(t *testing.T) {
	g := &Gateway{
		nodes: map[modules.NetAddress]*node{
			"foo":  {NetAddress: "foo", WasOutboundPeer: true},
			"bar":  {NetAddress: "bar", WasOutboundPeer: false},
			"baz":  {NetAddress: "baz", WasOutboundPeer: true},
			"quux": {NetAddress: "quux", WasOutboundPeer: false},
		},
	}
	nodelist := g.buildPeerManagerNodeList()
	// all outbound nodes should be at the front of the list
	var i int
	for i < len(nodelist) && g.nodes[nodelist[i]].WasOutboundPeer {
		i++
	}
	for i < len(nodelist) && !g.nodes[nodelist[i]].WasOutboundPeer {
		i++
	}
	if i != len(nodelist) {
		t.Fatal("bad nodelist:", nodelist)
	}
}
