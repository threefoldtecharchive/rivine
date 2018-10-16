package electrum

import "errors"

// JSONRPC 2.0 error codes as defined by the spec.
// These are only the error codes, not the full error
// objects returned.
const (
	ErrCodeParse          int64 = -32700
	ErrCodeInvalidRequest       = -32600
	ErrCodeMethodNotFound       = -32601
	ErrCodeInvalidParams        = -32602
	ErrCodeInternalError        = -32603
)

// Error codes we define ourselves. These error codes
// indicate an error which does not map cleanely to the
// predefined JSONRPC errors.
const (
	ErrCodeVersionAlreadySet = 101
)

// JSONRPC 2.0 errors as defined by the spec, see
// https://www.jsonrpc.org/specification
var (
	ErrParse = RPCError{
		Code:    ErrCodeParse,
		Message: "Error parsing the request",
	}
	ErrInvalidRequest = RPCError{
		Code:    ErrCodeInvalidRequest,
		Message: "Invalid request",
	}
	ErrMethodNotFound = RPCError{
		Code:    ErrCodeMethodNotFound,
		Message: "Method not found",
	}
	ErrInvalidParams = RPCError{
		Code:    ErrCodeInvalidParams,
		Message: "Invalid parameters",
	}
	ErrInternal = RPCError{
		Code:    ErrCodeInternalError,
		Message: "Internal error",
	}
)

// internal error conditions, so rpc calls can report
// what went wrong. This way the actual call doesn't
// need to be aware of what is happening
var (
	errFatal             = errors.New("Connection needs to be closed")
	errClient            = errors.New("Client error")
	errAlreadySubscribed = errors.New("Already subscribed to this address")
)
