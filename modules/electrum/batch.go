package electrum

import (
	"encoding/json"
)

type (
	// BatchRequest is a request send by the client. It can either be a single request,
	// or an array of (multiple) requests (batch request)
	BatchRequest struct {
		requests []*Request
		isBatch  bool
	}

	// BatchResponse is a response send to the client. Depending on the original request
	// it can be either a single response or a batch response
	BatchResponse struct {
		responses []*Response
		isBatch   bool
	}
)

// MarshalJSON produces a valid json value for the protocol version
func (br BatchRequest) MarshalJSON() ([]byte, error) {
	if len(br.requests) == 0 {
		return nil, nil
	}
	if br.isBatch {
		return json.Marshal(br.requests)
	}
	return json.Marshal(br.requests[0])
}

// UnmarshalJSON unmarshals a json value to a protocol version
func (br *BatchRequest) UnmarshalJSON(data []byte) error {
	var request Request
	if err := json.Unmarshal(data, &request); err != nil {
		var requests []*Request
		if err = json.Unmarshal(data, &requests); err != nil {
			return err
		}
		*br = BatchRequest{
			requests: requests,
			isBatch:  true,
		}
		return nil
	}
	*br = BatchRequest{
		requests: []*Request{&request},
		isBatch:  false,
	}

	return nil
}

// NewResponse creates a new BatchResponse from an existing BatchRequest
func (br BatchRequest) NewResponse() BatchResponse {
	return BatchResponse{
		responses: make([]*Response, len(br.requests)),
		isBatch:   br.isBatch,
	}
}

// MarshalJSON produces a valid json value for the protocol version
func (br BatchResponse) MarshalJSON() ([]byte, error) {
	if len(br.responses) == 0 {
		return nil, nil
	}
	if br.isBatch {
		// remove null responses
		for i := len(br.responses) - 1; i >= 0; i-- {
			if br.responses[i] == nil {
				br.responses = append(br.responses[:i], br.responses[i+1:]...)
			}
		}
		return json.Marshal(br.responses)
	}
	return json.Marshal(br.responses[0])
}

// UnmarshalJSON unmarshals a json value to a protocol version
func (br *BatchResponse) UnmarshalJSON(data []byte) error {
	var response Response
	if err := json.Unmarshal(data, &response); err != nil {
		var responses []*Response
		if err = json.Unmarshal(data, &responses); err != nil {
			return err
		}
		*br = BatchResponse{
			responses: responses,
			isBatch:   true,
		}
		return nil
	}
	*br = BatchResponse{
		responses: []*Response{&response},
		isBatch:   false,
	}

	return nil
}

// MustSend checks if the response needs to be send. This avoids sending "null"
// in case of a non-batch notification
func (br BatchResponse) MustSend() bool {
	if !br.isBatch && br.responses[0] == nil {
		return false
	}
	return true
}
