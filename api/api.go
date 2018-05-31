package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/rivine/rivine/modules"

	"github.com/julienschmidt/httprouter"
)

// Error is a type that is encoded as JSON and returned in an API response in
// the event of an error. Only the Message field is required. More fields may
// be added to this struct in the future for better error reporting.
type Error struct {
	// Message describes the error in English. Typically it is set to
	// `err.Error()`. This field is required.
	Message string `json:"message"`

	// TODO: add a Param field with the (omitempty option in the json tag)
	// to indicate that the error was caused by an invalid, missing, or
	// incorrect parameter. This is not trivial as the API does not
	// currently do parameter validation itself. For example, the
	// /gateway/connect endpoint relies on the gateway.Connect method to
	// validate the netaddress. However, this prevents the API from knowing
	// whether an error returned by gateway.Connect is because of a
	// connection error or an invalid netaddress parameter. Validating
	// parameters in the API is not sufficient, as a parameter's value may
	// be valid or invalid depending on the current state of a module.
}

// Error implements the error interface for the Error type. It returns only the
// Message field.
func (err Error) Error() string {
	return err.Message
}

// HttpGET is a utility function for making http get requests to sia with a
// whitelisted user-agent. A non-2xx response does not return an error.
func HttpGET(url string) (resp *http.Response, err error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Rivine-Agent")
	return http.DefaultClient.Do(req)
}

// HttpGETAuthenticated is a utility function for making authenticated http get
// requests to sia with a whitelisted user-agent and the supplied password. A
// non-2xx response does not return an error.
func HttpGETAuthenticated(url string, password string) (resp *http.Response, err error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Rivine-Agent")
	req.SetBasicAuth("", password)
	return http.DefaultClient.Do(req)
}

// HttpPOST is a utility function for making post requests to sia with a
// whitelisted user-agent. A non-2xx response does not return an error.
func HttpPOST(url string, data string) (resp *http.Response, err error) {
	var req *http.Request
	if data != "" {
		req, err = http.NewRequest("POST", url, strings.NewReader(data))
	} else {
		req, err = http.NewRequest("POST", url, nil)
	}
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Rivine-Agent")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return http.DefaultClient.Do(req)
}

// HttpPOSTAuthenticated is a utility function for making authenticated http
// post requests to sia with a whitelisted user-agent and the supplied
// password. A non-2xx response does not return an error.
func HttpPOSTAuthenticated(url string, data string, password string) (resp *http.Response, err error) {
	req, err := http.NewRequest("POST", url, strings.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Rivine-Agent")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth("", password)
	return http.DefaultClient.Do(req)
}

// RequireUserAgent is middleware that requires all requests to set a
// UserAgent that contains the specified string.
func RequireUserAgent(h http.Handler, ua string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if !strings.Contains(req.UserAgent(), ua) {
			WriteError(w, Error{"Browser access disabled due to security vulnerability. Use rivinec."}, http.StatusBadRequest)
			return
		}
		h.ServeHTTP(w, req)
	})
}

// RequirePassword is middleware that requires a request to authenticate with a
// password using HTTP basic auth. Usernames are ignored. Empty passwords
// indicate no authentication is required.
func RequirePassword(h httprouter.Handle, password string) httprouter.Handle {
	// An empty password is equivalent to no password.
	if password == "" {
		return h
	}
	return func(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
		_, pass, ok := req.BasicAuth()
		if !ok || pass != password {
			w.Header().Set("WWW-Authenticate", "Basic realm=\"SiaAPI\"")
			WriteError(w, Error{"API authentication failed."}, http.StatusUnauthorized)
			return
		}
		h(w, req, ps)
	}
}

// API encapsulates a collection of modules and implements a http.Handler
// to access their methods.
type API struct {
	cs       modules.ConsensusSet
	explorer modules.Explorer
	gateway  modules.Gateway
	tpool    modules.TransactionPool
	wallet   modules.Wallet

	router http.Handler
}

// api.ServeHTTP implements the http.Handler interface.
func (api *API) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	api.router.ServeHTTP(w, r)
}

// New creates a new Sia API from the provided modules.  The API will require
// authentication using HTTP basic auth for certain endpoints of the supplied
// password is not the empty string.  Usernames are ignored for authentication.
func New(requiredUserAgent string, requiredPassword string, cs modules.ConsensusSet, e modules.Explorer, g modules.Gateway, tp modules.TransactionPool, w modules.Wallet) *API {
	api := &API{
		cs:       cs,
		explorer: e,
		gateway:  g,
		tpool:    tp,
		wallet:   w,
	}

	// Register API handlers
	router := httprouter.New()
	router.NotFound = http.HandlerFunc(UnrecognizedCallHandler)

	// Consensus API Calls
	if api.cs != nil {
		router.GET("/consensus", api.consensusHandler)
		router.GET("/consensus/transactions/:id", api.consensusGetTransactionHandler)
		router.GET("/consensus/unspent/coinoutputs/:id", api.consensusGetUnspentCoinOutputHandler)
		router.GET("/consensus/unspent/blockstakeoutputs/:id", api.consensusGetUnspentBlockstakeOutputHandler)
	}

	// Explorer API Calls
	if api.explorer != nil {
		router.GET("/explorer", api.explorerHandler)
		router.GET("/explorer/blocks/:height", api.explorerBlocksHandler)
		router.GET("/explorer/hashes/:hash", api.explorerHashHandler)
		router.GET("/explorer/stats/history", api.historyStatsHandler)
		router.GET("/explorer/stats/range", api.rangeStatsHandler)
		router.GET("/explorer/constants", api.constantsHandler)
	}

	// Gateway API Calls
	if api.gateway != nil {
		router.GET("/gateway", api.gatewayHandler)
		router.POST("/gateway/connect/:netaddress", RequirePassword(api.gatewayConnectHandler, requiredPassword))
		router.POST("/gateway/disconnect/:netaddress", RequirePassword(api.gatewayDisconnectHandler, requiredPassword))
	}

	// TransactionPool API Calls
	if api.tpool != nil {
		// TODO: re-enable this route once the transaction pool API has been finalized
		router.GET("/transactionpool/transactions", api.transactionpoolTransactionsHandler)
		router.POST("/transactionpool/transactions", RequirePassword(api.transactionpoolPostTransactionHandler, requiredPassword))
	}

	// Wallet API Calls
	if api.wallet != nil {
		router.GET("/wallet", api.walletHandler)
		router.GET("/wallet/blockstakestats", RequirePassword(api.walletBlockStakeStats, requiredPassword))
		router.GET("/wallet/address", RequirePassword(api.walletAddressHandler, requiredPassword))
		router.GET("/wallet/addresses", api.walletAddressesHandler)
		router.GET("/wallet/backup", RequirePassword(api.walletBackupHandler, requiredPassword))
		router.POST("/wallet/init", RequirePassword(api.walletInitHandler, requiredPassword))
		router.POST("/wallet/lock", RequirePassword(api.walletLockHandler, requiredPassword))
		router.POST("/wallet/seed", RequirePassword(api.walletSeedHandler, requiredPassword))
		router.GET("/wallet/seeds", RequirePassword(api.walletSeedsHandler, requiredPassword))
		router.GET("/wallet/key/:unlockhash", RequirePassword(api.walletKeyHandler, requiredPassword))
		router.POST("/wallet/transaction", RequirePassword(api.walletTransactionCreateHandler, requiredPassword))
		router.POST("/wallet/coins", RequirePassword(api.walletCoinsHandler, requiredPassword))
		router.POST("/wallet/blockstakes", RequirePassword(api.walletBlockStakesHandler, requiredPassword))
		router.POST("/wallet/data", RequirePassword(api.walletDataHandler, requiredPassword))
		router.GET("/wallet/transaction/:id", api.walletTransactionHandler)
		router.GET("/wallet/transactions", api.walletTransactionsHandler)
		router.GET("/wallet/transactions/:addr", api.walletTransactionsAddrHandler)
		router.POST("/wallet/unlock", RequirePassword(api.walletUnlockHandler, requiredPassword))
		router.GET("/wallet/unlocked", RequirePassword(api.walletListUnlockedHandler, requiredPassword))
		router.GET("/wallet/locked", RequirePassword(api.walletListLockedHandler, requiredPassword))
		router.POST("/wallet/create/transaction", RequirePassword(api.walletCreateTransactionHandler, requiredPassword))
		router.POST("/wallet/sign", RequirePassword(api.walletSignHandler, requiredPassword))
	}

	// Apply UserAgent middleware and return the API
	api.router = RequireUserAgent(router, requiredUserAgent)
	return api
}

// UnrecognizedCallHandler handles calls to unknown pages (404).
func UnrecognizedCallHandler(w http.ResponseWriter, req *http.Request) {
	WriteError(w, Error{"404 - Refer to API.md"}, http.StatusNotFound)
}

// WriteError an error to the API caller.
func WriteError(w http.ResponseWriter, err Error, code int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(err) // ignore error, as it probably means that the status code does not allow a body
}

// WriteJSON writes the object to the ResponseWriter. If the encoding fails, an
// error is written instead. The Content-Type of the response header is set
// accordingly.
func WriteJSON(w http.ResponseWriter, obj interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if json.NewEncoder(w).Encode(obj) != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// WriteSuccess writes the HTTP header with status 204 No Content to the
// ResponseWriter. WriteSuccess should only be used to indicate that the
// requested action succeeded AND there is no data to return.
func WriteSuccess(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}
