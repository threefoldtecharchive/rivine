package daemon

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math/big"
	"net"
	"net/http"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/rivine/rivine/api"
	"github.com/rivine/rivine/build"
	"github.com/rivine/rivine/types"

	"github.com/inconshreveable/go-update"
	"github.com/julienschmidt/httprouter"
	"github.com/kardianos/osext"
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
	}

	// SiaConstants is a struct listing all of the constants in use.
	SiaConstants struct {
		GenesisTimestamp       types.Timestamp   `json:"genesistimestamp"`
		BlockSizeLimit         uint64            `json:"blocksizelimit"`
		BlockFrequency         types.BlockHeight `json:"blockfrequency"`
		FutureThreshold        types.Timestamp   `json:"futurethreshold"`
		ExtremeFutureThreshold types.Timestamp   `json:"extremefuturethreshold"`
		BlockStakeCount        types.Currency    `json:"blockstakecount"`

		BlockStakeAging       uint64         `json:"blockstakeaging"`
		BlockCreatorFee       types.Currency `json:"blockcreatorfee"`
		MinimumTransactionFee types.Currency `json:"minimumtransactionfee"`

		MaturityDelay         types.BlockHeight `json:"maturitydelay"`
		MedianTimestampWindow uint64            `json:"mediantimestampwindow"`

		RootTarget types.Target `json:"roottarget"`
		RootDepth  types.Target `json:"rootdepth"`

		TargetWindow      types.BlockHeight `json:"targetwindow"`
		MaxAdjustmentUp   *big.Rat          `json:"maxadjustmentup"`
		MaxAdjustmentDown *big.Rat          `json:"maxadjustmentdown"`

		OneCoin types.Currency `json:"onecoin"`
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

const (
	// The developer key is used to sign updates and other important Sia-
	// related information.
	developerKey = `-----BEGIN PUBLIC KEY-----
NOT USED FOR NOW
-----END PUBLIC KEY-----`
)

// fetchLatestRelease returns metadata about the most recent GitHub release.
func fetchLatestRelease() (githubRelease, error) {
	req, err := http.NewRequest("GET", "https://api.github.com/repos/rivine/rivine/releases/latest", nil)
	if err != nil {
		return githubRelease{}, err
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return githubRelease{}, err
	}
	defer resp.Body.Close()
	var release githubRelease
	err = json.NewDecoder(resp.Body).Decode(&release)
	if err != nil {
		return githubRelease{}, err
	}
	if release.TagName == "" && len(release.Assets) == 0 {
		return githubRelease{}, errEmptyUpdateResponse
	}
	return release, nil
}

// updateToRelease updates siad and siac to the release specified. siac is
// assumed to be in the same folder as siad.
func updateToRelease(release githubRelease) error {
	updateOpts := update.Options{
		Verifier: update.NewRSAVerifier(),
	}
	err := updateOpts.SetPublicKeyPEM([]byte(developerKey))
	if err != nil {
		// should never happen
		return err
	}

	binaryFolder, err := osext.ExecutableFolder()
	if err != nil {
		return err
	}

	// construct release filename
	releaseName := fmt.Sprintf("Rivine-%s-%s-%s.zip", release.TagName, runtime.GOOS, runtime.GOARCH)

	// find release
	var downloadURL string
	for _, asset := range release.Assets {
		if asset.Name == releaseName {
			downloadURL = asset.DownloadURL
			break
		}
	}
	if downloadURL == "" {
		return errors.New("couldn't find download URL for " + releaseName)
	}

	// download release archive
	resp, err := http.Get(downloadURL)
	if err != nil {
		return err
	}
	// release should be small enough to store in memory (<10 MiB); use
	// LimitReader to ensure we don't download more than 32 MiB
	content, err := ioutil.ReadAll(io.LimitReader(resp.Body, 1<<25))
	resp.Body.Close()
	if err != nil {
		return err
	}
	r := bytes.NewReader(content)
	z, err := zip.NewReader(r, r.Size())
	if err != nil {
		return err
	}

	// process zip, finding siad/siac binaries and signatures
	for _, binary := range []string{"rivined", "rivinec"} {
		var binData io.ReadCloser
		var signature []byte
		var binaryName string // needed for TargetPath below
		for _, zf := range z.File {
			switch base := path.Base(zf.Name); base {
			case binary, binary + ".exe":
				binaryName = base
				binData, err = zf.Open()
				if err != nil {
					return err
				}
				defer binData.Close()
			case binary + ".sig", binary + ".exe.sig":
				sigFile, err := zf.Open()
				if err != nil {
					return err
				}
				defer sigFile.Close()
				signature, err = ioutil.ReadAll(sigFile)
				if err != nil {
					return err
				}
			}
		}
		if binData == nil {
			return errors.New("could not find " + binary + " binary")
		} else if signature == nil {
			return errors.New("could not find " + binary + " signature")
		}

		// apply update
		updateOpts.Signature = signature
		updateOpts.TargetMode = 0775 // executable
		updateOpts.TargetPath = filepath.Join(binaryFolder, binaryName)
		err = update.Apply(binData, updateOpts)
		if err != nil {
			return err
		}
	}

	return nil
}

// daemonUpdateHandlerGET handles the API call that checks for an update.
func (srv *Server) daemonUpdateHandlerGET(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
	release, err := fetchLatestRelease()
	if err != nil {
		api.WriteError(w, api.Error{Message: "Failed to fetch latest release: " + err.Error()}, http.StatusInternalServerError)
		return
	}
	latestVersion, err := build.Parse(release.TagName)
	if err != nil {
		api.WriteError(w, api.Error{Message: "Failed to parse latest release: " + err.Error()}, http.StatusInternalServerError)
		return
	}
	api.WriteJSON(w, UpdateInfo{
		Available: latestVersion.Compare(build.Version) > 0,
		Version:   latestVersion.String(),
	})
}

// daemonUpdateHandlerPOST handles the API call that updates siad and siac.
// There is no safeguard to prevent "updating" to the same release, so callers
// should always check the latest version via daemonUpdateHandlerGET first.
// TODO: add support for specifying version to update to.
func (srv *Server) daemonUpdateHandlerPOST(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
	release, err := fetchLatestRelease()
	if err != nil {
		api.WriteError(w, api.Error{Message: "Failed to fetch latest release: " + err.Error()}, http.StatusInternalServerError)
		return
	}
	err = updateToRelease(release)
	if err != nil {
		if rerr := update.RollbackError(err); rerr != nil {
			api.WriteError(w, api.Error{Message: "Serious error: Failed to rollback from bad update: " + rerr.Error()}, http.StatusInternalServerError)
		} else {
			api.WriteError(w, api.Error{Message: "Failed to apply update: " + err.Error()}, http.StatusInternalServerError)
		}
		return
	}
	api.WriteSuccess(w)
}

// debugConstantsHandler prints a json file containing all of the constants.
func (srv *Server) daemonConstantsHandler(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
	sc := SiaConstants{
		GenesisTimestamp:       srv.chainCts.GenesisTimestamp,
		BlockSizeLimit:         srv.chainCts.BlockSizeLimit,
		BlockFrequency:         srv.chainCts.BlockFrequency,
		FutureThreshold:        srv.chainCts.FutureThreshold,
		ExtremeFutureThreshold: srv.chainCts.ExtremeFutureThreshold,
		BlockStakeCount:        srv.chainCts.GenesisBlockStakeCount(),

		BlockStakeAging:       srv.chainCts.BlockStakeAging,
		BlockCreatorFee:       srv.chainCts.BlockCreatorFee,
		MinimumTransactionFee: srv.chainCts.MinimumTransactionFee,

		MaturityDelay:         srv.chainCts.MaturityDelay,
		MedianTimestampWindow: srv.chainCts.MedianTimestampWindow,

		RootTarget: srv.chainCts.RootTarget(),
		RootDepth:  srv.chainCts.RootDepth,

		TargetWindow:      srv.chainCts.TargetWindow,
		MaxAdjustmentUp:   srv.chainCts.MaxAdjustmentUp,
		MaxAdjustmentDown: srv.chainCts.MaxAdjustmentDown,

		OneCoin: srv.chainCts.CurrencyUnits.OneCoin,
	}

	api.WriteJSON(w, sc)
}

// daemonVersionHandler handles the API call that requests the daemon's version.
func (srv *Server) daemonVersionHandler(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
	api.WriteJSON(w, DaemonVersion{Version: build.Version.String()})
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
	router.GET("/daemon/update", srv.daemonUpdateHandlerGET)
	router.POST("/daemon/update", srv.daemonUpdateHandlerPOST)
	router.POST("/daemon/stop", api.RequirePassword(srv.daemonStopHandler, password))

	return router
}

// NewServer creates a new net.http server listening on bindAddr.  Only the
// /daemon/ routes are registered by this func, additional routes can be
// registered later by calling serv.mux.Handle.
func NewServer(bindAddr, requiredUserAgent, requiredPassword string, chainCts types.ChainConstants) (*Server, error) {
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
