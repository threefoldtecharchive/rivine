package gateway

import (
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

// TestListen is a general test probling the connection listener.
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
	var gID gatewayID
	fastrand.Read(gID[:])
	addr := modules.NetAddress(conn.LocalAddr().String())
	ack, err := g.connectHandshake(conn, build.NewVersion(0, 0, 0), gID, true)
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

	// a simple 'conn.Close' would not obey the muxado disconnect protocol
	newSmuxClient(conn).Close()

	// compliant connect with invalid port
	conn, err = net.Dial("tcp", string(g.Address()))
	if err != nil {
		t.Fatal("dial failed:", err)
	}
	addr = modules.NetAddress(conn.LocalAddr().String())
	ack, err = g.connectHandshake(conn, build.Version, gID, true)
	if err != nil {
		t.Fatal(err)
	}
	if ack.Compare(build.Version) != 0 {
		t.Fatal("gateway should have given ack")
	}
	err = connectPortHandshake(conn, "0")
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 10; i++ {
		g.mu.RLock()
		_, ok := g.peers[addr]
		g.mu.RUnlock()
		if ok {
			t.Fatal("gateway should not have added a peer with an invalid port")
		}
		time.Sleep(20 * time.Millisecond)
	}

	// a simple 'conn.Close' would not obey the muxado disconnect protocol
	newSmuxClient(conn).Close()

	// compliant connect
	conn, err = net.Dial("tcp", string(g.Address()))
	if err != nil {
		t.Fatal("dial failed:", err)
	}
	addr = modules.NetAddress(conn.LocalAddr().String())
	ack, err = g.connectHandshake(conn, build.Version, gID, true)
	if err != nil {
		t.Fatal(err)
	}
	if ack.Compare(build.Version) != 0 {
		t.Fatal("gateway should have given ack")
	}

	err = connectPortHandshake(conn, addr.Port())
	if err != nil {
		t.Fatal(err)
	}

	// g should add the peer
	var ok bool
	for !ok {
		g.mu.RLock()
		_, ok = g.peers[addr]
		g.mu.RUnlock()
	}

	newSmuxClient(conn).Close()

	// g should remove the peer
	for ok {
		g.mu.RLock()
		_, ok = g.peers[addr]
		g.mu.RUnlock()
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
		// Test that Connect fails when the remote peer's version is < 0.0.1 (0).
		{
			version: build.NewVersion(0, 0, 0),
			errWant: insufficientVersionError("0.0.0"),
			msg:     "Connect should fail when the remote peer's version is 0.0.0",
		},
		// Test that Connect /could/ succeed when the remote peer's version is >= 0.1.0.
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
			remoteVersion, err := g.acceptConnHandshake(conn, tt.version, tt.uniqueID)
			if err != tt.localErrWant {
				panic(fmt.Sprintf("test #%d failed: %s", testIndex, err))
			} else if err == nil && build.Version.Compare(remoteVersion) != 0 {
				panic(fmt.Sprintf("test #%d failed: %q != %q",
					testIndex, build.Version.String(), remoteVersion.String()))
			}
		}()
		err = g.Connect(modules.NetAddress(listener.Addr().String()))
		if err != tt.errWant {
			t.Fatalf("expected Connect to error with '%v', but got '%v': %s", tt.errWant, err, tt.msg)
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
	g1.nodes = map[modules.NetAddress]struct{}{}
	g1.nodes[g2.Address()] = struct{}{}
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
