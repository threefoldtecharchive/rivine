package gateway

import (
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/NebulousLabs/fastrand"
	"github.com/rivine/rivine/build"
	"github.com/rivine/rivine/encoding"
	"github.com/rivine/rivine/modules"
	"github.com/rivine/rivine/types"
)

var (
	errPeerExists       = errors.New("already connected to this peer")
	errPeerRejectedConn = errors.New("peer rejected connection")
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
}

// versionHeader is sent as the initial exchange between peers.
// It ensures that peers take eachothers protocol version into account.
// It also prevents peers on different blockchains from connecting to each other,
// and prevents the gateway from connecting to itself.
// The receiving peer can set WantConn to false to refuse the connection,
// and the initiating peer van can set WantConn to false
// if they merely want to confirm that a node is online.
type versionHeader struct {
	Version   build.ProtocolVersion
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
// to handle its requests.
func (g *Gateway) addPeer(p *peer) {
	g.peers[p.NetAddress] = p
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

	remoteVersion, err := g.acceptConnHandshake(conn, build.Version, g.id)
	if err != nil {
		g.log.Debugf("INFO: %v wanted to connect but version handshake failed: %v", addr, err)
		conn.Close()
		return
	}

	err = g.managedAcceptConnPeer(conn, remoteVersion)
	if err != nil {
		g.log.Debugf("INFO: %v wanted to connect, but failed: %v", addr, err)
		conn.Close()
		return
	}
	// Handshake successful, remove the deadline.
	conn.SetDeadline(time.Time{})

	g.log.Debugf("INFO: accepted connection from new peer %v (v%v)", addr, remoteVersion)
}

// managedAcceptConnPeer accepts connection requests from peers.
// The requesting peer is added as a node and a peer. The peer is only added if
// a nil error is returned.
func (g *Gateway) managedAcceptConnPeer(conn net.Conn, remoteVersion build.ProtocolVersion) error {
	// Learn the peer's dialback address. Peers older than v1.0.0 will only be
	// able to be discovered by newer peers via the ShareNodes RPC.
	remoteAddr, err := acceptConnPortHandshake(conn)
	if err != nil {
		return err
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	// Don't accept a connection from a peer we're already connected to.
	if _, exists := g.peers[remoteAddr]; exists {
		return fmt.Errorf("already connected to a peer on that address: %v", remoteAddr)
	}
	// Accept the peer.
	g.acceptPeer(&peer{
		Peer: modules.Peer{
			Inbound: true,
			// NOTE: local may be true even if the supplied remoteAddr is not
			// actually reachable.
			Local:      remoteAddr.IsLocal(),
			NetAddress: remoteAddr,
			Version:    remoteVersion,
		},
		sess: newSmuxServer(conn),
	})

	// Attempt to ping the supplied address. If successful, and a connection is wanted,
	// we will add remoteAddr to our node list after accepting the peer. We do this in a
	// goroutine so that we can start communicating with the peer immediately.
	go func() {
		err := g.pingNode(remoteAddr)
		if err == nil {
			g.mu.Lock()
			g.addNode(remoteAddr)
			g.save()
			g.mu.Unlock()
		}
	}()

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
	for addr := range g.peers {
		// Do not kick outbound peers or local peers.
		if !p.Inbound || p.Local {
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

// acceptConnPortHandshake performs the port handshake and should be called on
// the side accepting a connection request. The remote address is only returned
// if err == nil.
func acceptConnPortHandshake(conn net.Conn) (remoteAddr modules.NetAddress, err error) {
	host, _, err := net.SplitHostPort(conn.RemoteAddr().String())
	if err != nil {
		return "", err
	}

	// Read the peer's port that we can dial them back on.
	var dialbackPort string
	err = encoding.ReadObject(conn, &dialbackPort, 13) // Max port # is 65535 (5 digits long) + 8 byte string length prefix
	if err != nil {
		return "", fmt.Errorf("could not read remote peer's port: %v", err)
	}
	remoteAddr = modules.NetAddress(net.JoinHostPort(host, dialbackPort))
	if err := remoteAddr.IsStdValid(); err != nil {
		return "", fmt.Errorf("peer's address (%v) is invalid: %v", remoteAddr, err)
	}
	// Sanity check to ensure that appending the port string to the host didn't
	// change the host. Only necessary because the peer sends the port as a string
	// instead of an integer.
	if remoteAddr.Host() != host {
		return "", fmt.Errorf("peer sent a port which modified the host")
	}
	return remoteAddr, nil
}

// connectPortHandshake performs the port handshake and should be called on the
// side initiating the connection request. This shares our port with the peer
// so they can connect to us in the future.
func connectPortHandshake(conn net.Conn, port string) error {
	err := encoding.WriteObject(conn, port)
	if err != nil {
		return errors.New("could not write port #: " + err.Error())
	}
	return nil
}

// acceptableVersionHeader returns an error if the version header is unacceptable.
func acceptableVersionHeader(ours, theirs versionHeader) error {
	if theirs.Version.Compare(minAcceptableVersion) < 0 {
		return insufficientVersionError(theirs.Version.String())
	} else if theirs.GenesisID != ours.GenesisID {
		return errPeerGenesisID
	} else if theirs.UniqueID == ours.UniqueID {
		return errOurAddress
	}

	return nil
}

// connectHandshake performs the version handshake and should be called
// on the side making the connection request. The remote version is only
// returned if err == nil.
func (g *Gateway) connectHandshake(conn net.Conn, version build.ProtocolVersion, uniqueID gatewayID, wantConn bool) (remoteVersion build.ProtocolVersion, err error) {
	ours := versionHeader{
		Version:   version,
		GenesisID: g.genesisBlockID,
		UniqueID:  uniqueID,
		WantConn:  wantConn,
	}

	// Send our version header.
	if err = encoding.WriteObject(conn, ours); err != nil {
		err = fmt.Errorf("failed to write version header: %v", err)
		return
	}

	var theirs versionHeader
	// Read remote version.
	if err = encoding.ReadObject(conn, &theirs, EncodedVersionHeaderLength); err != nil {
		err = fmt.Errorf("failed to read remote version header: %v", err)
		return
	}

	// validate if their header checks out against ours
	if err = acceptableVersionHeader(ours, theirs); err != nil {
		return
	}

	// checks if they want a connection or not
	if !theirs.WantConn {
		err = errPeerRejectedConn
		return
	}

	// all good, pass on the remote version to the caller
	remoteVersion = theirs.Version
	return
}

// acceptConnHandshake performs the version header handshake and should be
// called on the side accepting a connection request. The remote version is
// only returned if err == nil.
func (g *Gateway) acceptConnHandshake(conn net.Conn, version build.ProtocolVersion, uniqueID gatewayID) (remoteVersion build.ProtocolVersion, err error) {
	var theirs versionHeader
	// Read remote version.
	if err = encoding.ReadObject(conn, &theirs, EncodedVersionHeaderLength); err != nil {
		err = fmt.Errorf("failed to read remote version header: %v", err)
		return
	}

	ours := versionHeader{
		Version:   version,
		GenesisID: g.genesisBlockID,
		UniqueID:  uniqueID,
		WantConn:  true,
	}

	// validate if their header checks out against ours
	err = acceptableVersionHeader(ours, theirs)
	ours.WantConn = err == nil

	if err := encoding.WriteObject(conn, ours); err != nil {
		return remoteVersion, fmt.Errorf("failed to write version header: %v", err)
	}

	// if err was non-nil because of validation, return now
	if err != nil {
		return
	}

	// checks if they want a connection or not
	if !theirs.WantConn {
		err = errPeerRejectedConn
		return
	}

	// all good, pass on the remote version to the caller
	remoteVersion = theirs.Version
	return
}

// managedConnectPeer connects to peers. The peer is added as a
// node and a peer. The peer is only added if a nil error is returned.
func (g *Gateway) managedConnectPeer(conn net.Conn, remoteVersion build.ProtocolVersion, remoteAddr modules.NetAddress) error {
	g.mu.RLock()
	port := g.port
	g.mu.RUnlock()
	// Send our dialable address to the peer
	// so they can dial us back should we disconnect.
	if err := connectPortHandshake(conn, port); err != nil {
		return err
	}

	g.mu.Lock()
	defer g.mu.Unlock()
	g.addPeer(&peer{
		Peer: modules.Peer{
			Inbound:    false,
			Local:      remoteAddr.IsLocal(),
			NetAddress: remoteAddr,
			Version:    remoteVersion,
		},
		sess: newSmuxClient(conn),
	})
	// Add the peer to the node list. We can ignore the error: addNode
	// validates the address and checks for duplicates, but we don't care
	// about duplicates and we have already validated the address by
	// connecting to it.
	g.addNode(remoteAddr)
	return g.save()
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
		return errors.New("can't connect to invalid address")
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
	remoteVersion, err := g.connectHandshake(conn, build.Version, g.id, true)
	if err != nil {
		conn.Close()
		return err
	}

	err = g.managedConnectPeer(conn, remoteVersion, addr)
	if err != nil {
		conn.Close()
		return err
	}
	g.log.Debugln("INFO: connected to new peer", addr)

	// Connection successful, clear the timeout as to maintain a persistent
	// connection to this peer.
	conn.SetDeadline(time.Time{})

	// call initRPCs
	g.mu.RLock()
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
	g.mu.RUnlock()

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
	delete(g.peers, addr)
	g.mu.Unlock()
	if err := p.sess.Close(); err != nil {
		return err
	}

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
