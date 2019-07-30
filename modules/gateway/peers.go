package gateway

import (
	"errors"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/NebulousLabs/fastrand"
	"github.com/threefoldtech/rivine/build"
	"github.com/threefoldtech/rivine/modules"
	"github.com/threefoldtech/rivine/pkg/encoding/siabin"
	"github.com/threefoldtech/rivine/types"
)

var (
	errPeerExists       = errors.New("already connected to this peer")
	errPeerRejectedConn = errors.New("peer rejected connection")
	errPeerNoConnWanted = errors.New("peer did not want a connection")
)

var (
	// constant to explicitly indicate a reject,
	// new since v1.0.2
	rejectedVersion = build.NewPrereleaseVersion(0, 0, 0, 0, "reject")
)

// insufficientVersionError indicates a peer's version is insufficient.
type insufficientVersionError string

// Error implements the error interface for insufficientVersionError.
func (s insufficientVersionError) Error() string {
	return "unacceptable version: " + string(s)
}

type peer struct {
	modules.Peer
	sess streamSession
	// rate limiting channel
	token chan struct{}
}

// sessionHeader is sent as the initial exchange between peers.
// It prevents peers on different blockchains from connecting to each other,
// and prevents the gateway from connecting to itself.
// The receiving peer can set WantConn to false to refuse the connection,
// and the initiating peer van can set WantConn to false
// if they merely want to confirm that a node is online.
//
// The version is send prior to any session header,
// a handshake consists out of a Version + Session Header,
// which session header depends upon the version header.
type sessionHeader struct {
	GenesisID types.BlockID
	UniqueID  gatewayID
	WantConn  bool
}

func (p *peer) open() (modules.PeerConn, error) {
	conn, err := p.sess.Open()
	if err != nil {
		return nil, err
	}
	return &peerConn{conn, p.NetAddress}, nil
}

func (p *peer) accept() (modules.PeerConn, error) {
	conn, err := p.sess.Accept()
	if err != nil {
		return nil, err
	}
	return &peerConn{conn, p.NetAddress}, nil
}

// addPeer adds a peer to the Gateway's peer list and spawns a listener thread
// to handle its requests and increments the remotePeers accordingly
func (g *Gateway) addPeer(p *peer) {
	g.peers[p.NetAddress] = p
	g.log.Debugln("Added peer on address", p.NetAddress)
	go g.threadedListenPeer(p)
}

// randomOutboundPeer returns a random outbound peer.
func (g *Gateway) randomOutboundPeer() (modules.NetAddress, error) {
	// Get the list of outbound peers.
	var addrs []modules.NetAddress
	for addr, peer := range g.peers {
		if peer.Inbound {
			continue
		}
		addrs = append(addrs, addr)
	}
	if len(addrs) == 0 {
		return "", errNoPeers
	}

	// Of the remaining options, select one at random.
	return addrs[fastrand.Intn(len(addrs))], nil
}

// permanentListen handles incoming connection requests. If the connection is
// accepted, the peer will be added to the Gateway's peer list.
func (g *Gateway) permanentListen(closeChan chan struct{}) {
	// Signal that the permanentListen thread has completed upon returning.
	defer close(closeChan)

	for {
		conn, err := g.listener.Accept()
		if err != nil {
			g.log.Debugln("[PL] Closing permanentListen:", err)
			return
		}

		go g.threadedAcceptConn(conn)

		// Sleep after each accept. This limits the rate at which the Gateway
		// will accept new connections. The intent here is to prevent new
		// incoming connections from kicking out old ones before they have a
		// chance to request additional nodes.
		select {
		case <-time.After(acceptInterval):
		case <-g.threads.StopChan():
			return
		}
	}
}

// threadedAcceptConn adds a connecting node as a peer.
func (g *Gateway) threadedAcceptConn(conn net.Conn) {
	if g.threads.Add() != nil {
		conn.Close()
		return
	}
	defer g.threads.Done()
	conn.SetDeadline(time.Now().Add(connStdDeadline))

	addr := modules.NetAddress(conn.RemoteAddr().String())
	g.log.Debugf("INFO: %v wants to connect", addr)

	remoteInfo, err := g.acceptConnHandshake(conn, g.bcInfo.ProtocolVersion, g.id)
	if err != nil {
		g.log.Debugf("INFO: %v wanted to connect but handshake failed: %v", addr, err)
		conn.Close()
		return
	}

	err = g.managedAcceptConnPeer(conn, remoteInfo)
	if err != nil {
		g.log.Debugf("INFO: %v wanted to connect, but failed: %v", addr, err)
		conn.Close()
		return
	}
	// Handshake successful, remove the deadline.
	conn.SetDeadline(time.Time{})

	g.log.Debugf("INFO: accepted connection from new peer '%v -> %v' (v%s)",
		addr, remoteInfo.NetAddress, remoteInfo.Version.String())
}

// managedAcceptConnPeer accepts connection requests from peers.
// The requesting peer is added as a node and a peer. The peer is only added if
// a nil error is returned.
func (g *Gateway) managedAcceptConnPeer(conn net.Conn, remoteInfo remoteInfo) error {
	// Get the remote address on which the connecting peer is listening on.
	// This means we need to combine the incoming connections ip address with
	// the announced open port of the peer.
	remoteIP := modules.NetAddress(conn.RemoteAddr().String()).Host()
	remotePort := remoteInfo.NetAddress.Port()
	remoteAddr := modules.NetAddress(net.JoinHostPort(remoteIP, remotePort))

	// Accept the peer.
	peer := &peer{
		Peer: modules.Peer{
			Inbound: true,
			// NOTE: local may be true even if the supplied remoteAddr is not
			// actually reachable.
			Local: remoteAddr.IsLocal(),
			// Ignoring claimed IP address (which should be == to the socket address)
			// by the host but keeping note of the port number so we can call back
			NetAddress: remoteAddr,
			Version:    remoteInfo.Version,
		},
		sess:  newSmuxServer(conn),
		token: make(chan struct{}, g.concurrentRPCPerPeer),
	}
	for i := 0; uint64(i) < g.concurrentRPCPerPeer; i++ {
		// Fill the channel wit htokens
		peer.token <- struct{}{}
	}

	g.mu.Lock()
	g.acceptPeer(peer)
	g.mu.Unlock()

	// Attempt to ping the supplied address. If successful, we will add
	// remoteInfo.NetAddress to our node list after accepting the peer. We do this in a
	// goroutine so that we can start communicating with the peer immediately.
	if g.peerTG.Add() == nil {
		// adding it to peer's Thread Group,
		// as to ensure that the logger doesn't get closed,
		// prior to this ping-pong gorutine being finished.
		go func() {
			defer g.peerTG.Done()
			err := g.pingNode(remoteAddr)
			if err == nil {
				g.mu.Lock()
				g.addNode(remoteAddr)
				g.mu.Unlock()
			}
		}()
	}

	return nil
}

// acceptPeer makes room for the peer if necessary by kicking out existing
// peers, then adds the peer to the peer list.
func (g *Gateway) acceptPeer(p *peer) {
	// If we are not fully connected, add the peer without kicking any out.
	if len(g.peers) < fullyConnectedThreshold {
		g.addPeer(p)
		return
	}

	// Select a peer to kick. Outbound peers and local peers are not
	// available to be kicked.
	var addrs []modules.NetAddress
	for addr, peer := range g.peers {
		// Do not kick outbound peers or local peers.
		if !peer.Inbound || peer.Local {
			continue
		}

		// Prefer kicking a peer with the same hostname.
		if addr.Host() == p.NetAddress.Host() {
			addrs = []modules.NetAddress{addr}
			break
		}
		addrs = append(addrs, addr)
	}
	if len(addrs) == 0 {
		// There is nobody suitable to kick, therefore do not kick anyone.
		g.addPeer(p)
		return
	}

	// Of the remaining options, select one at random.
	kick := addrs[fastrand.Intn(len(addrs))]

	g.peers[kick].sess.Close()
	delete(g.peers, kick)
	g.log.Printf("INFO: disconnected from %v to make room for %v\n", kick, p.NetAddress)
	g.addPeer(p)
}

// remoteInfo is the info we care about about our remote connection,
// after a successful handshake
type remoteInfo struct {
	Version    build.ProtocolVersion
	NetAddress modules.NetAddress
}

// connectHandshake performs the version handshake and should be called
// on the side making the connection request.
func (g *Gateway) connectHandshake(conn net.Conn, version build.ProtocolVersion, uniqueID gatewayID, netAddress modules.NetAddress, wantConn bool) (remoteInfo remoteInfo, err error) {
	// Send our version header.
	if err = siabin.WriteObject(conn, version); err != nil {
		err = fmt.Errorf("failed to write version header: %v", err)
		return
	}
	// Send our session header.
	ourSessionHeader := sessionHeader{
		GenesisID: g.genesisBlockID,
		UniqueID:  uniqueID,
		WantConn:  wantConn,
	}
	if err = siabin.WriteObject(conn, ourSessionHeader); err != nil {
		err = fmt.Errorf("failed to write session header: %v", err)
		return
	}

	// read their version
	if err = siabin.ReadObject(conn, &remoteInfo.Version, build.EncodedVersionLength); err != nil {
		err = fmt.Errorf("failed to read remote version header: %v", err)
		return
	}

	// check if we've been rejected
	if remoteInfo.Version.Compare(rejectedVersion) == 0 {
		err = errPeerRejectedConn
		return
	}

	// check if remote version is acceptable
	if remoteInfo.Version.Compare(MinAcceptableVersion) < 0 {
		// invalid version
		err = insufficientVersionError(remoteInfo.Version.String())
		return
	}

	// read their session header.
	var theirs sessionHeader
	if err = siabin.ReadObject(conn, &theirs, EncodedSessionHeaderLength); err != nil {
		err = fmt.Errorf("failed to read session header: %v", err)
		return
	}
	// check content
	if theirs.GenesisID != g.genesisBlockID {
		err = errPeerGenesisID
		return
	}
	if theirs.UniqueID == uniqueID {
		err = errOurAddress
		return
	}

	// continue handshake based on lowest version
	lowestVersion := version // be positive, assume ours is lowest
	if remoteInfo.Version.Compare(lowestVersion) < 0 {
		// theirs is lower, use that one
		lowestVersion = remoteInfo.Version
	}
	// now compare the version, as based on that we might want something verry different
	if lowestVersion.Compare(HandshakNetAddressUpgrade) >= 0 {
		// v1.0.2+
		remoteInfo.NetAddress, err = g.connectSessionHandshakeV102(conn, theirs, netAddress)
	} else {
		// v1.0.0 and v1.0.1 (launch version)
		remoteInfo.NetAddress, err = g.connectSessionHandshakeV100(conn, theirs)
	}
	if err == nil && !theirs.WantConn {
		err = errPeerNoConnWanted
	}
	return
}

func (g *Gateway) connectSessionHandshakeV100(conn net.Conn, theirs sessionHeader) (remoteAddress modules.NetAddress, err error) {
	// send our port, so they can dial us back, should we d/c due to an error
	g.mu.RLock()
	port := g.port
	g.mu.RUnlock()
	err = siabin.WriteObject(conn, port)
	if err != nil {
		err = errors.New("could not write port #: " + err.Error())
		return
	}
	// simply use the conn's remote address as their address
	remoteAddress = modules.NetAddress(conn.RemoteAddr().String())
	return
}

func (g *Gateway) connectSessionHandshakeV102(conn net.Conn, theirs sessionHeader, netAddress modules.NetAddress) (remoteAddress modules.NetAddress, err error) {
	// send our net address first, as we are the one wanting to connect
	g.log.Debugln("accept: sending our netaddr:", netAddress, netAddress.IsLocal())
	err = siabin.WriteObject(conn, netAddress)
	if err != nil {
		err = errors.New("could not write address: " + err.Error())
		return
	}
	// receive their address
	err = siabin.ReadObject(conn, &remoteAddress, modules.MaxEncodedNetAddressLength)
	if err != nil {
		err = errors.New("connect: could not read remote address: " + err.Error())
		return
	}
	if remoteAddress == "reject" {
		// remote address is rejected
		err = errPeerRejectedConn
		return
	}
	// standard check their addr
	if err = remoteAddress.IsStdValid(); err != nil {
		err = fmt.Errorf("invalid remote address: %v", err)
		return
	}
	return
}

// acceptConnHandshake performs the version header handshake and should be
// called on the side accepting a connection request.
// Incoming version dicates which handshake version to use,
// meaning we'll use an older handshake protocol, even if we support a newer one.
func (g *Gateway) acceptConnHandshake(conn net.Conn, version build.ProtocolVersion, uniqueID gatewayID) (remoteInfo remoteInfo, err error) {
	var (
		theirs sessionHeader
		legacy bool
	)
	remoteInfo.Version, theirs, legacy, err = g.readRemoteHeaders(conn)
	if err != nil {
		return
	}
	if legacy {
		// 2nd part of legacy logic,
		// as to be able to receive incoming connections
		// from peers who are still on 1.0.0 or 1.0.1,
		// we'll drop support for them in about two months as well,
		// as it's still beta after all,
		// so no need to keep this in forever.
		if !theirs.WantConn {
			err = errPeerRejectedConn
			return
		}

		if remoteInfo.Version.Compare(MinAcceptableVersion) < 0 { // already drop support for peers < 1.0.0, as that's prior to any public launch
			err = insufficientVersionError(remoteInfo.Version.String())
		} else if remoteInfo.Version.Compare(HandshakNetAddressUpgrade) >= 0 {
			err = insufficientVersionError(remoteInfo.Version.String())
		} else if theirs.GenesisID != g.genesisBlockID {
			err = errPeerGenesisID
		} else if theirs.UniqueID == uniqueID {
			err = errOurAddress
		}
		var legacyErr error
		remoteInfo.NetAddress, legacyErr = g.legacyAcceptConnectHandshake(conn, version, uniqueID, err == nil)
		if legacyErr != nil {
			err = fmt.Errorf("failed to write version header: %v", legacyErr)
		}
		return
	}

	// based on the incoming remote version we'll handle the next steps a bit different
	// check version, and handle based on that from here on out

	// reject too low versions
	if remoteInfo.Version.Compare(MinAcceptableVersion) < 0 {
		// return invalid version
		err = insufficientVersionError(remoteInfo.Version.String())
		g.writeRejectVersionHeader(conn, err)
		return
	}

	// write our version, as their info checks out
	err = siabin.WriteObject(conn, version)
	if err != nil {
		return
	}

	// compare this received information
	if theirs.GenesisID != g.genesisBlockID {
		err = errPeerGenesisID
	}
	if err == nil && theirs.UniqueID == uniqueID {
		err = errOurAddress
	}

	// write our header
	ours := sessionHeader{
		GenesisID: g.genesisBlockID,
		UniqueID:  uniqueID,
		WantConn:  err == nil,
	}
	if e := siabin.WriteObject(conn, ours); e != nil {
		err = e
	}
	if err != nil {
		return
	}

	// continue handshake based on lowest version
	lowestVersion := version // be positive, assume ours is lowest
	if remoteInfo.Version.Compare(lowestVersion) < 0 {
		// theirs is lower, use that one
		lowestVersion = remoteInfo.Version
	}
	if lowestVersion.Compare(HandshakNetAddressUpgrade) >= 0 {
		// v1.0.2+
		remoteInfo.NetAddress, err = g.acceptConnSessionHandshakeV102(conn)
	} else {
		// v1.0.0 and v1.0.1 (launch version)
		remoteInfo.NetAddress, err = g.acceptConnSessionHandshakeV100(conn)
	}
	if err == nil && !theirs.WantConn {
		err = errPeerNoConnWanted
	}
	return
}

func (g *Gateway) readRemoteHeaders(conn net.Conn) (remoteVersion build.ProtocolVersion, remoteHeader sessionHeader, legacy bool, err error) {
	prefix := make([]byte, 8)
	if _, err = io.ReadFull(conn, prefix); err != nil {
		return
	}
	dataLen := siabin.DecUint64(prefix)
	if dataLen == legacyEncodedVersionHeaderLength {
		legacy = true

		data := make([]byte, dataLen)
		_, err = io.ReadFull(conn, data)
		if err != nil {
			return
		}

		// assume legacy peer
		var legacyVersionHeader legacyVersionHeader
		err = siabin.Unmarshal(data, &legacyVersionHeader)
		if err != nil {
			return
		}

		remoteVersion = legacyVersionHeader.Version
		remoteHeader.GenesisID = legacyVersionHeader.GenesisID
		remoteHeader.UniqueID = legacyVersionHeader.UniqueID
		remoteHeader.WantConn = legacyVersionHeader.WantConn
		return
	}

	// should equal version size
	if dataLen != build.EncodedVersionLength {
		err = errors.New("invalid data len received, cannot proceed with accept handshake")
		return
	}

	data := make([]byte, dataLen)
	_, err = io.ReadFull(conn, data)
	if err != nil {
		return
	}
	// read remote version
	if err = siabin.Unmarshal(data, &remoteVersion); err != nil {
		err = fmt.Errorf("failed to read remote version header: %v", err)
		g.writeRejectVersionHeader(conn, err)
		return
	}

	// read remote session header,
	err = siabin.ReadObject(conn, &remoteHeader, EncodedSessionHeaderLength)
	return
}

// We support accepting connections from legacy peers,
// HOWEVER, we do not accept connecting to a legacy peer

func (g *Gateway) legacyAcceptConnectHandshake(conn net.Conn, version build.ProtocolVersion, uniqueID gatewayID, wantConn bool) (remoteAddr modules.NetAddress, err error) {
	ours := legacyVersionHeader{
		Version:   version,
		GenesisID: g.genesisBlockID,
		UniqueID:  uniqueID,
		WantConn:  wantConn,
	}
	err = siabin.WriteObject(conn, ours)
	if err != nil {
		return
	}

	remoteAddr, err = g.acceptConnSessionHandshakeV100(conn)
	return
}

var legacyMinAcceptableVersion = build.NewVersion(0, 0, 1, 0)

const legacyEncodedVersionHeaderLength = 57

type legacyVersionHeader struct {
	Version   build.ProtocolVersion
	GenesisID types.BlockID
	UniqueID  gatewayID
	WantConn  bool
}

func (g *Gateway) acceptConnSessionHandshakeV100(conn net.Conn) (remoteAddress modules.NetAddress, err error) {
	// continue with net-port handshake
	host, _, err := net.SplitHostPort(conn.RemoteAddr().String())
	if err != nil {
		return "", err
	}

	// Read the peer's port that we can dial them back on.
	var dialbackPort string
	err = siabin.ReadObject(conn, &dialbackPort, 13) // Max port # is 65535 (5 digits long) + 8 byte string length prefix
	if err != nil {
		return "", fmt.Errorf("could not read remote peer's port: %v", err)
	}
	remoteAddress = modules.NetAddress(net.JoinHostPort(host, dialbackPort))
	err = remoteAddress.IsStdValid()
	if err != nil {
		err = fmt.Errorf("peer's address (%v) is invalid: %v", remoteAddress, err)
		return
	}
	// Sanity check to ensure that appending the port string to the host didn't
	// change the host. Only necessary because the peer sends the port as a string
	// instead of an integer.
	if remoteAddress.Host() != host {
		err = fmt.Errorf("peer sent a port which modified the host")
		return
	}

	// all good, return
	return
}

func (g *Gateway) acceptConnSessionHandshakeV102(conn net.Conn) (remoteAddress modules.NetAddress, err error) {
	// receive their address first, as we accept
	err = siabin.ReadObject(conn, &remoteAddress, modules.MaxEncodedNetAddressLength)
	if err != nil {
		err = errors.New("accept: could not read remote address: " + err.Error())
	}
	// validate the address
	err = remoteAddress.IsStdValid()
	if err != nil {
		if err := siabin.WriteObject(conn, "reject"); err != nil {
			g.log.Println("WARN: failed to write reject address:", err)
		}
		err = fmt.Errorf("peer's address (%v) is invalid: %v", remoteAddress, err)
	}
	// write now our net address
	g.mu.RLock()
	gaddr := g.myAddr
	g.mu.RUnlock()
	g.log.Debugln("accept: sending our netaddr:", gaddr, gaddr.IsLocal())
	err = siabin.WriteObject(conn, gaddr)
	if err != nil {
		err = errors.New("could not write address: " + err.Error())
		return
	}
	return
}

func (g *Gateway) writeRejectVersionHeader(conn net.Conn, reason error) {
	// write our version
	err := siabin.WriteObject(conn, rejectedVersion)
	if err != nil {
		g.log.Printf(`WARN: failed to write rejected version for "%v": %v`, reason, err)
	}
}

// managedConnect establishes a persistent connection to a peer, and adds it to
// the Gateway's peer list.
func (g *Gateway) managedConnect(addr modules.NetAddress) error {
	// Perform verification on the input address.
	g.mu.RLock()
	gaddr := g.myAddr
	g.mu.RUnlock()
	if addr == gaddr {
		return errors.New("can't connect to our own address")
	}
	if err := addr.IsStdValid(); err != nil {
		return errors.New("can't connect to invalid address: " + err.Error())
	}
	if net.ParseIP(addr.Host()) == nil {
		return errors.New("address must be an IP address")
	}
	g.mu.RLock()
	_, exists := g.peers[addr]
	g.mu.RUnlock()
	if exists {
		return errPeerExists
	}

	// Dial the peer and perform peer initialization.
	conn, err := g.dial(addr)
	if err != nil {
		return err
	}

	// Perform peer initialization.
	remoteInfo, err := g.connectHandshake(conn, g.bcInfo.ProtocolVersion, g.id, gaddr, true)
	if err != nil {
		conn.Close()
		return err
	}

	// Connection successful, clear the timeout as to maintain a persistent
	// connection to this peer.
	conn.SetDeadline(time.Time{})

	// Add the peer.
	g.mu.Lock()
	defer g.mu.Unlock()

	peer := &peer{
		Peer: modules.Peer{
			Inbound:    false,
			Local:      addr.IsLocal(),
			NetAddress: addr,
			Version:    remoteInfo.Version,
		},
		sess:  newSmuxClient(conn),
		token: make(chan struct{}, g.concurrentRPCPerPeer),
	}
	for i := 0; uint64(i) < g.concurrentRPCPerPeer; i++ {
		// Fill the channel with tokens
		peer.token <- struct{}{}
	}

	g.addPeer(peer)
	g.addNode(addr)
	g.nodes[addr].WasOutboundPeer = true

	if err := g.saveSync(); err != nil {
		g.log.Println("ERROR: Unable to save new outbound peer to gateway:", err)
	}

	g.log.Debugln("INFO: connected to new peer", addr)

	// call initRPCs
	for name, fn := range g.initRPCs {
		go func(name string, fn modules.RPCFunc) {
			if g.threads.Add() != nil {
				return
			}
			defer g.threads.Done()

			err := g.managedRPC(addr, name, fn)
			if err != nil {
				g.log.Debugf("INFO: RPC %q on peer %q failed: %v", name, addr, err)
			}
		}(name, fn)
	}

	return nil
}

// Connect establishes a persistent connection to a peer, and adds it to the
// Gateway's peer list.
func (g *Gateway) Connect(addr modules.NetAddress) error {
	if err := g.threads.Add(); err != nil {
		return err
	}
	defer g.threads.Done()
	return g.managedConnect(addr)
}

// Disconnect terminates a connection to a peer and removes it from the
// Gateway's peer list. The peer's address remains in the node list.
func (g *Gateway) Disconnect(addr modules.NetAddress) error {
	if err := g.threads.Add(); err != nil {
		return err
	}
	defer g.threads.Done()

	g.mu.RLock()
	p, exists := g.peers[addr]
	g.mu.RUnlock()
	if !exists {
		return errors.New("not connected to that node")
	}

	p.sess.Close()
	g.mu.Lock()
	// Peer is removed from the peer list as well as the node list, to prevent
	// the node from being re-connected while looking for a replacement peer.
	delete(g.peers, addr)
	delete(g.nodes, addr)
	g.mu.Unlock()

	g.log.Debugln("Disconnected from peer:", addr)
	g.log.Println("INFO: disconnected from peer", addr)
	return nil
}

// Peers returns the addresses currently connected to the Gateway.
func (g *Gateway) Peers() []modules.Peer {
	g.mu.RLock()
	defer g.mu.RUnlock()
	var peers []modules.Peer
	for _, p := range g.peers {
		peers = append(peers, p.Peer)
	}
	return peers
}

// Online returns true if the node is connected to the internet.
// During testing and dev we always assume that the node is online.
func (g *Gateway) Online() bool {
	if build.Release == "dev" || build.Release == "testing" {
		return true
	}

	g.mu.RLock()
	defer g.mu.RUnlock()
	for _, p := range g.peers {
		if !p.Local {
			return true
		}
	}
	return false
}
