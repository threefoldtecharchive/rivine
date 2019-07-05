package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/bgentry/speakeasy"
)

var (
	// ErrStatusNotFound is returned when status wasn't found.
	ErrStatusNotFound = errors.New("expecting a response, but API returned status code 204 No Content")
)

// HTTPError is return for HTTP Errors by the HTTPClient
type HTTPError struct {
	internalError error
	statusCode    int
}

// Error implements error.Error
func (e *HTTPError) Error() string {
	return fmt.Sprintf("HTTP %d error: %v", e.statusCode, e.internalError)
}

// HTTPStatusCode returns the internal status code,
// returned by the HTTP client in case of an error.
func (e *HTTPError) HTTPStatusCode() int {
	return e.statusCode
}

// Non2xx returns true for non-success HTTP status codes.
func Non2xx(code int) bool {
	return code < 200 || code > 299
}

// DecodeError returns the Error from a API response. This method should
// only be called if the response's status code is non-2xx. The error returned
// may not be of type Error in the event of an error unmarshalling the
// JSON.
func DecodeError(resp *http.Response) error {
	var apiErr Error
	err := json.NewDecoder(resp.Body).Decode(&apiErr)
	if err != nil {
		return err
	}
	return apiErr
}

// HTTPClient is used to communicate with the Rivine-based daemon,
// using the exposed (local) REST API over HTTP.
type HTTPClient struct {
	RootURL   string
	Password  string
	UserAgent string
}

// PostResp makes a POST API call and decodes the response. An error is
// returned if the response status is not 2xx.
func (c *HTTPClient) PostResp(call, data string, reply interface{}) error {
	resp, err := c.apiPost(call, data)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent {
		return errors.New("expecting a response, but API returned status code 204 No Content")
	}

	err = json.NewDecoder(resp.Body).Decode(&reply)
	if err != nil {
		return err
	}
	return nil
}

// Post makes an API call and discards the response. An error is returned if
// the response status is not 2xx.
func (c *HTTPClient) Post(call, data string) error {
	resp, err := c.apiPost(call, data)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

// GetAPI makes a GET API call and decodes the response. An error is returned
// if the response status is not 2xx.
func (c *HTTPClient) GetAPI(call string, obj interface{}) error {
	resp, err := c.apiGet(call, "")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent {
		return ErrStatusNotFound
	}

	err = json.NewDecoder(resp.Body).Decode(obj)
	if err != nil {
		return err
	}
	return nil
}

// GetWithData makes an API call and discards the response. An error is returned if the
// response status is not 2xx.
func (c *HTTPClient) GetWithData(call, data string) error {
	resp, err := c.apiGet(call, data)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

// GetAPIWithData makes a GET API call and decodes the response. An error is returned
// if the response status is not 2xx.
func (c *HTTPClient) GetAPIWithData(call, data string, obj interface{}) error {
	resp, err := c.apiGet(call, data)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent {
		return ErrStatusNotFound
	}

	err = json.NewDecoder(resp.Body).Decode(obj)
	if err != nil {
		return err
	}
	return nil
}

// Get makes an API call and discards the response. An error is returned if the
// response status is not 2xx.
func (c *HTTPClient) Get(call string) error {
	resp, err := c.apiGet(call, "")
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

// ApiGet wraps a GET request with a status code check, such that if the GET does
// not return 2xx, the error will be read and returned. When no error is returned,
// the response's body isn't closed, otherwise it is.
func (c *HTTPClient) apiGet(call, data string) (*http.Response, error) {
	resp, err := HTTPGet(c.RootURL+call, data, c.UserAgent)
	if err != nil {
		return nil, errors.New("no response from daemon")
	}
	// check error code
	if resp.StatusCode == http.StatusUnauthorized {
		resp.Body.Close()
		// try again using an authenticated HTTP Post call
		password, err := c.apiPassword()
		if err != nil {
			return nil, err
		}
		resp, err = HTTPGETAuthenticated(c.RootURL+call, data, c.UserAgent, password)
		if err != nil {
			return nil, errors.New("no response from daemon - authentication failed")
		}
	}
	if Non2xx(resp.StatusCode) {
		if resp.StatusCode == http.StatusNotFound {
			resp.Body.Close()
			return nil, &HTTPError{
				internalError: errors.New("API call not recognized: " + call),
				statusCode:    resp.StatusCode,
			}
		}
		err := DecodeError(resp)
		resp.Body.Close()
		return nil, &HTTPError{
			internalError: err,
			statusCode:    resp.StatusCode,
		}
	}
	return resp, nil
}

// ApiPost wraps a POST request with a status code check, such that if the POST
// does not return 2xx, the error will be read and returned. When no error is returned,
// the response's body isn't closed, otherwise it is.
func (c *HTTPClient) apiPost(call, data string) (*http.Response, error) {
	resp, err := HTTPPost(c.RootURL+call, data, c.UserAgent)
	if err != nil {
		return nil, errors.New("no response from daemon")
	}
	// check error code
	if resp.StatusCode == http.StatusUnauthorized {
		b, rErr := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if rErr != nil || !c.responseIsAPIPasswordError(b) {
			apiErr, ok := c.responseAsError(b)
			if ok {
				return nil, fmt.Errorf("Unauthorized (401): %s", apiErr.Message)
			}
			return nil, errors.New("API Call failed with the (401) unauthorized status")
		}
		// try again using an authenticated HTTP Post call
		password, err := c.apiPassword()
		if err != nil {
			return nil, err
		}
		resp, err = HTTPPostAuthenticated(c.RootURL+call, data, c.UserAgent, password)
		if err != nil {
			return nil, errors.New("no response from daemon - authentication failed")
		}
	}
	if Non2xx(resp.StatusCode) {
		if resp.StatusCode == http.StatusNotFound {
			resp.Body.Close()
			return nil, &HTTPError{
				internalError: errors.New("API call not recognized: " + call),
				statusCode:    resp.StatusCode,
			}
		}
		err := DecodeError(resp)
		resp.Body.Close()
		return nil, &HTTPError{
			internalError: err,
			statusCode:    resp.StatusCode,
		}
	}
	return resp, nil
}

func (c *HTTPClient) responseIsAPIPasswordError(resp []byte) bool {
	err, ok := c.responseAsError(resp)
	return ok && err.Message == "API Basic authentication failed."
}

func (c *HTTPClient) responseAsError(resp []byte) (Error, bool) {
	var apiError Error
	err := json.Unmarshal(resp, &apiError)
	return apiError, err == nil
}

func (c *HTTPClient) apiPassword() (string, error) {
	if c.Password != "" {
		return c.Password, nil
	}
	var err error
	c.Password, err = speakeasy.Ask("API password: ")
	if err != nil {
		return "", err
	}
	return c.Password, nil
}
