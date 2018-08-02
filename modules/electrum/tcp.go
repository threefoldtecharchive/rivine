package electrum

import (
	"encoding/json"
	"io"
	"net"
	"sync"
)

type tcpConn struct {
	conn        net.Conn
	dec         *json.Decoder
	enc         *json.Encoder
	requestChan chan *BatchRequest
	errorChan   chan error
	stopChan    chan struct{}

	// Synchronize writes
	mu sync.Mutex
}

// listenTCP listens for new tcp connections, and registers them as new rpc connections
func (e *Electrum) listenTCP() error {
	for {
		conn, err := e.tcpServer.Accept()
		if err != nil {
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
		requestChan: make(chan *BatchRequest),
		errorChan:   make(chan error),
		stopChan:    make(chan struct{}),
	}

	// Start goroutine which reads on the connection
	go func() {
		for {
			req := &BatchRequest{}
			err := c.dec.Decode(req)

			select {
			case <-c.stopChan:
				return
			default:
			}
			if err != nil {
				if err == io.EOF || err == io.ErrUnexpectedEOF {
					c.errorChan <- errTransportClosed{err}
					// No need to keep on reading here
					return
				}
				c.errorChan <- err
				// decoder errored, replace with new one
				c.dec = json.NewDecoder(c.conn)
				continue
			}
			c.requestChan <- req
		}
	}()

	return c
}

func (c *tcpConn) Close(wg *sync.WaitGroup) error {
	close(c.stopChan)
	close(c.requestChan)
	close(c.errorChan)
	// Wait untill all active calls have finished before actually closing
	// the connection
	wg.Wait()
	return c.conn.Close()
}

func (c *tcpConn) GetError() <-chan error {
	return c.errorChan
}

func (c *tcpConn) GetRequest() <-chan *BatchRequest {
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

func (c *tcpConn) IsClosed() <-chan struct{} {
	return c.stopChan
}
