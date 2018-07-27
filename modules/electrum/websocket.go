package electrum

import (
	"io"
	"net"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

type websocketConn struct {
	conn        *websocket.Conn
	requestChan chan *Request
	errorChan   chan error
	stopChan    chan struct{}

	// Synchronize writes
	mu sync.Mutex
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func (e *Electrum) handleWs(w http.ResponseWriter, r *http.Request) {
	// Accept any origin
	upgrader.CheckOrigin = func(r *http.Request) bool {
		return true
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		e.log.Println("Error while creating websocket connection:", err)
		return
	}
	// wrap the connection, also starting to monitor for incomming
	// requests.
	wsConn := createWebsocketConn(conn)
	e.log.Debugln("Opened websocket connection to", conn.RemoteAddr())
	go e.ServeRPC(wsConn)
}

func createWebsocketConn(conn *websocket.Conn) *websocketConn {
	c := &websocketConn{
		conn:        conn,
		requestChan: make(chan *Request),
		errorChan:   make(chan error),
		stopChan:    make(chan struct{}),
	}

	// Start goroutine which reads on the connection
	go func() {
		for {
			// Since we only use jsonrpc, we can use the
			// ReadJSON convenience method here
			req := &Request{}
			err := c.conn.ReadJSON(req)
			select {
			case <-c.stopChan:
				return
			default:
			}
			if err != nil {
				// Since we don't expect the client to close the connection,
				// everything is unexpected
				if websocket.IsUnexpectedCloseError(err) || err == io.ErrUnexpectedEOF {
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

func (c *websocketConn) Close(wg *sync.WaitGroup) error {
	close(c.stopChan)
	close(c.requestChan)
	close(c.errorChan)
	// Wait untill all active calls have finished before actually closing
	// the connection
	wg.Wait()
	return c.conn.Close()
}

func (c *websocketConn) GetError() <-chan error {
	return c.errorChan
}

func (c *websocketConn) GetRequest() <-chan *Request {
	return c.requestChan
}

func (c *websocketConn) Send(msg interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.conn.WriteJSON(msg)
}

func (c *websocketConn) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

func (c *websocketConn) IsClosed() <-chan struct{} {
	return c.stopChan
}
