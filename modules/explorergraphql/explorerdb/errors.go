package explorerdb

import (
	"errors"
	"fmt"
)

var (
	// ErrNotFound is returned in case a requested object is not found in the DB.
	ErrNotFound = errors.New("not found")
)

type (
	// InternalError is a generic error returned for unexpected DB errors
	InternalError interface {
		error

		InternalError() error
	}

	internalDBWrappedError struct {
		name string
		err  error
	}
)

var (
	_ InternalError = (*internalDBWrappedError)(nil)
)

// NewInternalError returns a new internal explorer DB error
func NewInternalError(dbName string, err error) InternalError {
	return &internalDBWrappedError{
		name: dbName,
		err:  err,
	}
}

func (idberr *internalDBWrappedError) Error() string {
	return fmt.Sprintf("internal %s DB Error: %v", idberr.name, idberr.err)
}

func (idberr *internalDBWrappedError) InternalError() error {
	return idberr.err
}
