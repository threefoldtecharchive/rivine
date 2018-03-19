package datastore

import "errors"

// Namespace is an identifier used to group data written on the blockchain when it is
// replicated to external storage
type Namespace [NamespaceBytes]byte

// NamespaceBytes is the byte length of a namespace
const NamespaceBytes = 4

// ErrInvalidNamespaceLength indicates that a namespace identifier is used which does not have the
// correct length according to the NamespaceBytes
var ErrInvalidNamespaceLength = errors.New("Namespace identifier does not have the right length")

// String convets the Namespace to a string representation
func (ns *Namespace) String() string {
	return string(ns[:])
}

// LoadString tries to convert a string representation of a namespace into the corresponding bytes
func (ns *Namespace) LoadString(nsString string) error {
	if len(nsString) != NamespaceBytes {
		return ErrInvalidNamespaceLength
	}
	copy(ns[:], []byte(nsString))
	return nil
}
