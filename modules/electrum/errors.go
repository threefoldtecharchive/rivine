package electrum

import "errors"

// JSONRPC 2.0 errors as defined by the spec, see
// https://www.jsonrpc.org/specification
var (
	ErrParse = RPCError{
		Code:    -32700,
		Message: "Error parsing the request",
	}
	ErrInvalidRequest = RPCError{
		Code:    -32600,
		Message: "Invalid request",
	}
	ErrMethodNotFound = RPCError{
		Code:    -32601,
		Message: "Method not found",
	}
	ErrInvalidParams = RPCError{
		Code:    -32602,
		Message: "Invalid parameters",
	}
	ErrInternal = RPCError{
		Code:    -32603,
		Message: "Internal error",
	}
)

// internal error conditions, so rpc calls can report
// what went wrong. This way the actual call doesn't
// need to be aware of what is happening
var (
	errFatal  = errors.New("Connection needs to be closed")
	errClient = errors.New("Client error")
)
