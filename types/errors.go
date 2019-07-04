package types

import (
	"fmt"
	"net/http"

	"github.com/threefoldtech/rivine/build"
)

type ClientErrorKind uint32

const (
	ClientErrorBadRequest      ClientErrorKind = 0
	ClientErrorUnauthorized    ClientErrorKind = 1
	ClientErrorPaymentRequired ClientErrorKind = 2
	ClientErrorForbidden       ClientErrorKind = 3
	ClientErrorNotFound        ClientErrorKind = 4
	ClientErrorTimeout         ClientErrorKind = 8

	maxClientError = ClientErrorTimeout
)

func (kind ClientErrorKind) String() string {
	switch kind {
	case ClientErrorBadRequest:
		return "bad request"
	case ClientErrorUnauthorized:
		return "unauthorized"
	case ClientErrorPaymentRequired:
		return "payment required"
	case ClientErrorForbidden:
		return "forbidden"
	case ClientErrorNotFound:
		return "not found"
	case ClientErrorTimeout:
		return "time out"
	default:
		return "???"
	}
}

func (kind ClientErrorKind) AsHTTPStatusCode() int {
	switch kind {
	case ClientErrorBadRequest:
		return http.StatusBadRequest
	case ClientErrorUnauthorized:
		return http.StatusUnauthorized
	case ClientErrorPaymentRequired:
		return http.StatusPaymentRequired
	case ClientErrorForbidden:
		return http.StatusForbidden
	case ClientErrorNotFound:
		return http.StatusNotFound
	case ClientErrorTimeout:
		return http.StatusRequestTimeout
	default:
		return http.StatusInternalServerError
	}
}

type ClientError struct {
	Err  error
	Kind ClientErrorKind
}

func NewClientError(err error, kind ClientErrorKind) ClientError {
	if kind > maxClientError {
		build.Severe("invalid client error kind", kind, err)
		kind = ClientErrorBadRequest
	}
	return ClientError{
		Err:  err,
		Kind: kind,
	}
}

func (ce ClientError) Error() string {
	return fmt.Sprintf("User Error %s: %v", ce.Kind, ce.Err)
}
