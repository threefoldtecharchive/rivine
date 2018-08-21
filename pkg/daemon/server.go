package daemon

import (
	"net"
	"net/http"
	"strings"
)

// HTTPServer creates and serves a HTTP server that offers communication using a REST API.
type HTTPServer struct {
	httpServer *http.Server
	mux        *http.ServeMux
	listener   net.Listener
}

// NewHTTPServer creates a new net.http server listening on bindAddr.
func NewHTTPServer(bindAddr string) (*HTTPServer, error) {
	l, err := net.Listen("tcp", bindAddr)
	if err != nil {
		return nil, err
	}
	mux := http.NewServeMux()
	return &HTTPServer{
		mux:      mux,
		listener: l,
		httpServer: &http.Server{
			Handler: mux,
		},
	}, nil
}

// Handle the given pattern using the given handler.
func (srv *HTTPServer) Handle(pattern string, handler http.Handler) {
	srv.mux.Handle(pattern, handler)
}

// Serve all registered endpoins as a REST API over HTTP endpoints.
func (srv *HTTPServer) Serve() error {
	// The server will run until an error is encountered or the listener is
	// closed, via the Close method. Closing the listener will result in the benign error handled below.
	err := srv.httpServer.Serve(srv.listener)
	if err != nil && !strings.HasSuffix(err.Error(), "use of closed network connection") {
		return err
	}
	return nil
}

// Close closes the Server's listener, causing the HTTP server to shut down.
func (srv *HTTPServer) Close() error {
	// Close the listener, which will cause Server.Serve() to return.
	return srv.listener.Close()
}
