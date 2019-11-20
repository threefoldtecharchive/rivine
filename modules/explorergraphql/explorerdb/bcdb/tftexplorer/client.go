package tftexplorer

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Client allows you to communicate with the TFT Explorer,
// communicating over the Gedis HTTP interface.
type Client struct {
	addr string
	http *http.Client
}

// NewClient creates a new TFT Explore Client.
// See `Client` for more information.
func NewClient(addr string) *Client {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	return &Client{
		addr: fmt.Sprintf("%s/web/gedis/http/tft_explorer/", addr),
		http: client,
	}
}

// Get data from the TFT Explorer actor.
func (cl *Client) Get(method string, arguments interface{}, response interface{}) error {
	return cl.httpDo("GET", method, arguments, response)
}

// Set data in the TFT Explorer actor.
func (cl *Client) Set(method string, arguments interface{}, response interface{}) error {
	return cl.httpDo("POST", method, arguments, response)
}

func (cl *Client) httpDo(httpMethod, method string, arguments interface{}, response interface{}) error {
	// define body reader if needed
	var br io.Reader
	if arguments != nil {
		b, err := json.Marshal(requestBody{Arguments: arguments})
		if err != nil {
			return fmt.Errorf("failed to JSON marshal request body for method %s: %v", method, err)
		}
		br = bytes.NewReader(b)
	}

	// create GET HTTP Request
	req, err := http.NewRequest(httpMethod, cl.addr+method, br)
	if err != nil {
		return fmt.Errorf("failed to create %s request for method %s: %v", httpMethod, method, err)
	}

	// add HTTP Headers
	req.Header.Add("Content-Type", "application/json")

	// Execute the Request
	resp, err := cl.http.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute %s request for method %s: %v", httpMethod, method, err)
	}
	defer resp.Body.Close()

	// if a response is given to decode too, do so
	if response != nil {
		err = json.NewDecoder(resp.Body).Decode(response)
		if err != nil {
			return fmt.Errorf("failed to decode response returned by %s request for method %s: %v", httpMethod, method, err)
		}
	}

	// all good
	return nil
}

type (
	requestBody struct {
		Arguments interface{} `json:"args"`
	}
)