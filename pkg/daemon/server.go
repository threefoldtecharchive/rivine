package daemon

import (
	"errors"
	"net"
	"net/http"
	"strings"

	"github.com/rivine/rivine/api"
	"github.com/rivine/rivine/build"
	"github.com/rivine/rivine/modules"
	"github.com/rivine/rivine/types"

	"github.com/julienschmidt/httprouter"
)

var errEmptyUpdateResponse = errors.New("API call to https://api.github.com/repos/rivine/rivine/releases/latest is returning an empty response")

type (
	// Server creates and serves a HTTP server that offers communication with a
	// Sia API.
	Server struct {
		httpServer *http.Server
		mux        *http.ServeMux
		listener   net.Listener
		chainCts   types.ChainConstants
		bcInfo     types.BlockchainInfo
	}

	DaemonVersion struct {
		Version string `json:"version"`
	}
	// UpdateInfo indicates whether an update is available, and to what
	// version.
	UpdateInfo struct {
		Available bool   `json:"available"`
		Version   string `json:"version"`
	}
	// githubRelease represents some of the JSON returned by the GitHub release API
	// endpoint. Only the fields relevant to updating are included.
	githubRelease struct {
		TagName string `json:"tag_name"`
		Assets  []struct {
			Name        string `json:"name"`
			DownloadURL string `json:"browser_download_url"`
		} `json:"assets"`
	}
)

// debugConstantsHandler prints a json file containing all of the constants.
func (srv *Server) daemonConstantsHandler(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
	constants := modules.NewDaemonConstants(srv.bcInfo, srv.chainCts)
	api.WriteJSON(w, constants)
}

// daemonVersionHandler handles the API call that requests the daemon's version.
func (srv *Server) daemonVersionHandler(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
	api.WriteJSON(w, DaemonVersion{Version: srv.bcInfo.ChainVersion.String()})
}

// daemonStopHandler handles the API call to stop the daemon cleanly.
func (srv *Server) daemonStopHandler(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
	// can't write after we stop the server, so lie a bit.
	api.WriteSuccess(w)

	// need to flush the response before shutting down the server
	f, ok := w.(http.Flusher)
	if !ok {
		panic("Server does not support flushing")
	}
	f.Flush()

	if err := srv.Close(); err != nil {
		build.Critical(err)
	}
}

func (srv *Server) daemonHandler(password string) http.Handler {
	router := httprouter.New()

	router.GET("/daemon/constants", srv.daemonConstantsHandler)
	router.GET("/daemon/version", srv.daemonVersionHandler)
	router.POST("/daemon/stop", api.RequirePassword(srv.daemonStopHandler, password))

	return router
}

// NewServer creates a new net.http server listening on bindAddr.  Only the
// /daemon/ routes are registered by this func, additional routes can be
// registered later by calling serv.mux.Handle.
func NewServer(bindAddr, requiredUserAgent, requiredPassword string, chainCts types.ChainConstants, bcInfo types.BlockchainInfo) (*Server, error) {
	// Create the listener for the server
	l, err := net.Listen("tcp", bindAddr)
	if err != nil {
		return nil, err
	}

	// Create the Server
	mux := http.NewServeMux()
	srv := &Server{
		mux:      mux,
		listener: l,
		httpServer: &http.Server{
			Handler: mux,
		},
		chainCts: chainCts,
		bcInfo:   bcInfo,
	}

	// Register siad routes
	srv.mux.Handle("/daemon/", api.RequireUserAgent(srv.daemonHandler(requiredPassword), requiredUserAgent))

	return srv, nil
}

func (srv *Server) Serve() error {
	// The server will run until an error is encountered or the listener is
	// closed, via either the Close method or the signal handling above.
	// Closing the listener will result in the benign error handled below.
	err := srv.httpServer.Serve(srv.listener)
	if err != nil && !strings.HasSuffix(err.Error(), "use of closed network connection") {
		return err
	}
	return nil
}

// Close closes the Server's listener, causing the HTTP server to shut down.
func (srv *Server) Close() error {
	// Close the listener, which will cause Server.Serve() to return.
	return srv.listener.Close()
}
