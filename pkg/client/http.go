package client

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/bgentry/speakeasy"
	"github.com/rivine/rivine/api"
)

// Non2xx returns true for non-success HTTP status codes.
func Non2xx(code int) bool {
	return code < 200 || code > 299
}

// DecodeError returns the api.Error from a API response. This method should
// only be called if the response's status code is non-2xx. The error returned
// may not be of type api.Error in the event of an error unmarshalling the
// JSON.
func DecodeError(resp *http.Response) error {
	var apiErr api.Error
	err := json.NewDecoder(resp.Body).Decode(&apiErr)
	if err != nil {
		return err
	}
	return apiErr
}

// HTTPClient is used to communicate with the Rivine-based daemon,
// using the exposed (local) REST API over HTTP.
type HTTPClient struct {
	RootURL string
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
	resp, err := c.apiGet(call)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent {
		return errors.New("expecting a response, but API returned status code 204 No Content")
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
	resp, err := c.apiGet(call)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

// ApiGet wraps a GET request with a status code check, such that if the GET does
// not return 2xx, the error will be read and returned. When no error is returned,
// the response's body isn't closed, otherwise it is.
func (c *HTTPClient) apiGet(call string) (*http.Response, error) {
	resp, err := api.HttpGET("http://" + c.RootURL + call)
	if err != nil {
		return nil, errors.New("no response from daemon")
	}
	// check error code
	if resp.StatusCode == http.StatusUnauthorized {
		resp.Body.Close()
		// Prompt for password and retry request with authentication.
		password, err := speakeasy.Ask("API password: ")
		if err != nil {
			return nil, err
		}
		resp, err = api.HttpGETAuthenticated("http://"+c.RootURL+call, password)
		if err != nil {
			return nil, errors.New("no response from daemon - authentication failed")
		}
	}
	if resp.StatusCode == http.StatusNotFound {
		resp.Body.Close()
		return nil, errors.New("API call not recognized: " + call)
	}
	if Non2xx(resp.StatusCode) {
		err := DecodeError(resp)
		resp.Body.Close()
		return nil, err
	}
	return resp, nil
}

// ApiPost wraps a POST request with a status code check, such that if the POST
// does not return 2xx, the error will be read and returned. When no error is returned,
// the response's body isn't closed, otherwise it is.
func (c *HTTPClient) apiPost(call, data string) (*http.Response, error) {
	resp, err := api.HttpPOST("http://"+c.RootURL+call, data)
	if err != nil {
		return nil, errors.New("no response from daemon")
	}
	// check error code
	if resp.StatusCode == http.StatusUnauthorized {
		resp.Body.Close()
		// Prompt for password and retry request with authentication.
		password, err := speakeasy.Ask("API password: ")
		if err != nil {
			return nil, err
		}
		resp, err = api.HttpPOSTAuthenticated("http://"+c.RootURL+call, data, password)
		if err != nil {
			return nil, errors.New("no response from daemon - authentication failed")
		}
	}
	if resp.StatusCode == http.StatusNotFound {
		resp.Body.Close()
		return nil, errors.New("API call not recognized: " + call)
	}
	if Non2xx(resp.StatusCode) {
		err := DecodeError(resp)
		resp.Body.Close()
		return nil, err
	}
	return resp, nil
}
