package electrum

import (
	"bytes"
	"encoding/json"
	"reflect"
	"testing"
)

func TestBatchRequestDecode(t *testing.T) {
	t.Parallel()

	type testCase struct {
		json   string
		result BatchRequest
	}

	// Keep a seperate list of the params as constants can't be addresses
	params := []json.RawMessage{
		json.RawMessage([]byte(`{"protocol_version":["1.0","1.0"]}`)),
		json.RawMessage([]byte(`{"address":"015a080a9259b9d4aaa550e2156f49b1a79a64c7ea463d810d4493e8242e6791584fbdac553e6f"}`)),
	}

	// Note that integers in json are actualy NUMBER's, which mean they are float64's (as ID is an interface{} field)
	testCases := []testCase{
		{
			json:   `{"id":1,"jsonrpc":"2.0","method":"server.ping"}`,
			result: BatchRequest{isBatch: false, requests: []*Request{&Request{ID: 1., JSONRPC: jsonRPCVersion, Method: "server.ping"}}},
		},
		{
			json:   `{"id":2,"jsonrpc":"2.0","method":"server.version","params":{"protocol_version":["1.0","1.0"]}}`,
			result: BatchRequest{isBatch: false, requests: []*Request{&Request{ID: 2., JSONRPC: jsonRPCVersion, Method: "server.version", Params: &params[0]}}},
		},
		{
			json:   `{"id":3,"jsonrpc":"2.0","method":"blockchain.address.subscribe","params":{"address":"015a080a9259b9d4aaa550e2156f49b1a79a64c7ea463d810d4493e8242e6791584fbdac553e6f"}}`,
			result: BatchRequest{isBatch: false, requests: []*Request{&Request{ID: 3., JSONRPC: jsonRPCVersion, Method: "blockchain.address.subscribe", Params: &params[1]}}},
		},
		{
			json:   `[{"id":4,"jsonrpc":"2.0","method":"server.ping"}]`,
			result: BatchRequest{isBatch: true, requests: []*Request{&Request{ID: 4., JSONRPC: jsonRPCVersion, Method: "server.ping"}}},
		},
		{
			json:   `[{"id":5,"jsonrpc":"2.0","method":"server.version","params":{"protocol_version":["1.0","1.0"]}}]`,
			result: BatchRequest{isBatch: true, requests: []*Request{&Request{ID: 5., JSONRPC: jsonRPCVersion, Method: "server.version", Params: &params[0]}}},
		},
		{
			json:   `[{"id":6,"jsonrpc":"2.0","method":"blockchain.address.subscribe","params":{"address":"015a080a9259b9d4aaa550e2156f49b1a79a64c7ea463d810d4493e8242e6791584fbdac553e6f"}}]`,
			result: BatchRequest{isBatch: true, requests: []*Request{&Request{ID: 6., JSONRPC: jsonRPCVersion, Method: "blockchain.address.subscribe", Params: &params[1]}}},
		},
		{
			json: `[{"id":7,"jsonrpc":"2.0","method":"server.ping"}, {"id":8,"jsonrpc":"2.0","method":"server.version","params":{"protocol_version":["1.0","1.0"]}}]`,
			result: BatchRequest{isBatch: true, requests: []*Request{
				&Request{ID: 7., JSONRPC: jsonRPCVersion, Method: "server.ping"},
				&Request{ID: 8., JSONRPC: jsonRPCVersion, Method: "server.version", Params: &params[0]},
			}},
		},
	}

	for i, test := range testCases {
		var result BatchRequest
		err := json.NewDecoder(bytes.NewBufferString(test.json)).Decode(&result)
		if err != nil {
			t.Error("Decoding batch request", i, "failed:", err)
		}

		if result.isBatch != test.result.isBatch {
			t.Error("Test isBatch failed for case", i)
		}
		if len(result.requests) != len(test.result.requests) {
			t.Error("Actual amount of requests differs from expected amount of requests")
		}
		for j := range result.requests {
			if !reflect.DeepEqual(test.result.requests[j], result.requests[j]) {
				t.Error("Request", i, ", Request with index", j, "does not match, got", *test.result.requests[j], "expected", *result.requests[j])
			}
		}
	}

}

func TestBatchResponseEncodeDecode(t *testing.T) {
	t.Parallel()

	testCases := []BatchResponse{
		BatchResponse{isBatch: false, responses: []*Response{&Response{ID: 1., JSONRPC: jsonRPCVersion, Result: "test"}}},
		BatchResponse{isBatch: true, responses: []*Response{&Response{ID: 1., JSONRPC: jsonRPCVersion, Result: "test"}}},
	}

	for i, test := range testCases {
		buf := bytes.NewBuffer(nil)
		err := json.NewEncoder(buf).Encode(test)
		if err != nil {
			t.Error("Failed to encode batch response", i, ", error:", err)
		}
		result := BatchResponse{}
		err = json.NewDecoder(buf).Decode(&result)
		if err != nil {
			t.Error("Failed to decode batch response", i, ", error:", err)
		}
		if result.isBatch != test.isBatch {
			t.Error("test isBatch failed for case", i)
		}
		if len(result.responses) != len(test.responses) {
			t.Error("Actual amount of responses differs from expected amount of responses")
		}
		for j := range result.responses {
			if !reflect.DeepEqual(test.responses[j], result.responses[j]) {
				t.Error("Response", i, ", Response with index", j, "does not match, got", *result.responses[j], "expected", *test.responses[j])
			}
		}
	}
}
