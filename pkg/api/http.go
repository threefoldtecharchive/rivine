package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/julienschmidt/httprouter"
)

// client util functions

// HTTPGet is a utility function for making http get requests to sia with a
// whitelisted user-agent. A non-2xx response does not return an error.
func HTTPGet(url, data, userAgent string) (resp *http.Response, err error) {
	var req *http.Request
	if data != "" {
		req, err = http.NewRequest("GET", url, strings.NewReader(data))
	} else {
		req, err = http.NewRequest("GET", url, nil)
	}
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	return http.DefaultClient.Do(req)
}

// HTTPGETAuthenticated is a utility function for making authenticated http get
// requests to sia with a whitelisted user-agent and the supplied password. A
// non-2xx response does not return an error.
func HTTPGETAuthenticated(url, data, userAgent, password string) (resp *http.Response, err error) {
	var req *http.Request
	if data != "" {
		req, err = http.NewRequest("GET", url, strings.NewReader(data))
	} else {
		req, err = http.NewRequest("GET", url, nil)
	}
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	req.SetBasicAuth("", password)
	return http.DefaultClient.Do(req)
}

// HTTPPost is a utility function for making post requests to sia with a
// whitelisted user-agent. A non-2xx response does not return an error.
func HTTPPost(url, data, userAgent string) (resp *http.Response, err error) {
	var req *http.Request
	if data != "" {
		req, err = http.NewRequest("POST", url, strings.NewReader(data))
	} else {
		req, err = http.NewRequest("POST", url, nil)
	}
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return http.DefaultClient.Do(req)
}

// HTTPPostAuthenticated is a utility function for making authenticated http
// post requests to sia with a whitelisted user-agent and the supplied
// password. A non-2xx response does not return an error.
func HTTPPostAuthenticated(url, data, userAgent, password string) (resp *http.Response, err error) {
	req, err := http.NewRequest("POST", url, strings.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth("", password)
	return http.DefaultClient.Do(req)
}

// server middleware: handler->handler

// RequireUserAgentHandler is middleware that requires all requests to set a
// UserAgent that contains the specified string.
func RequireUserAgentHandler(h http.Handler, userAgent string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if !strings.Contains(req.UserAgent(), userAgent) {
			WriteError(w, Error{"Browser access disabled due to security vulnerability. Use an official client."}, http.StatusBadRequest)
			return
		}
		h.ServeHTTP(w, req)
	})
}

// RequirePasswordHandler is middleware that requires a request to authenticate with a
// password using HTTP basic auth. Usernames are ignored. Empty passwords
// indicate no authentication is required.
func RequirePasswordHandler(h httprouter.Handle, password string) httprouter.Handle {
	// An empty password is equivalent to no password.
	if password == "" {
		return h
	}
	return func(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
		_, pass, ok := req.BasicAuth()
		if !ok || pass != password {
			w.Header().Set("WWW-Authenticate", "Basic realm=\"SiaAPI\"")
			WriteError(w, Error{"API Basic authentication failed."}, http.StatusUnauthorized)
			return
		}
		h(w, req, ps)
	}
}

// server util functions to write errors and JSON-encoded bodies

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
