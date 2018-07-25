package electrum

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"sync"
)

type tcpConn struct {
	conn        net.Conn
	dec         *json.Decoder
	enc         *json.Encoder
	requestChan chan *Request
	errorChan   chan error

	// Synchronize writes
	mu sync.Mutex
}

// listenTCP listens for new tcp connections, and registers them as new rpc connections
func (e *Electrum) listenTCP() error {
	for {
		conn, err := e.tcpServer.Accept()
		if err != nil {
			// TODO: Ignore server close errors
			return err
		}
		tcpConn := createTCPConn(conn)
		e.log.Debugln("Opened tcp connection to", conn.RemoteAddr())
		go e.ServeRPC(tcpConn)
	}
}

func createTCPConn(conn net.Conn) *tcpConn {
	c := &tcpConn{
		conn:        conn,
		dec:         json.NewDecoder(conn),
		enc:         json.NewEncoder(conn),
		requestChan: make(chan *Request),
		errorChan:   make(chan error),
	}

	// Start goroutine which reads on the connection
	go func() {
		for {
			// Since we only use jsonrpc, we can use the
			// ReadJSON convenience method here
			req := &Request{}
			if err := c.dec.Decode(req); err != nil {
				if err == io.EOF || err == io.ErrUnexpectedEOF {
					fmt.Println("The transport is closed")
					c.errorChan <- errTransportClosed{err}
					// No need to keep on reading here
					return
				}
				c.errorChan <- err
				continue
			}
			c.requestChan <- req
		}
	}()

	return c
}

func (c *tcpConn) Close() error {
	return c.conn.Close()
}

func (c *tcpConn) GetError() <-chan error {
	return c.errorChan
}

func (c *tcpConn) GetRequest() <-chan *Request {
	return c.requestChan
}

func (c *tcpConn) Send(msg interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.enc.Encode(msg)
}

func (c *tcpConn) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}
