package modules

import (
	"errors"
	"io"
	"time"

	"github.com/threefoldtech/rivine/build"
	"github.com/threefoldtech/rivine/pkg/encoding/siabin"
)

const (
	// AcceptResponse is the response given to an RPC call to indicate
	// acceptance, i.e. that the sender wishes to continue communication.
	AcceptResponse = "accept"

	// StopResponse is the response given to an RPC call to indicate graceful
	// termination, i.e. that the sender wishes to cease communication, but
	// not due to an error.
	StopResponse = "stop"

	// NegotiateDownloadTime defines the amount of time that the renter and
	// host have to negotiate a download request batch. The time is set high
	// enough that two nodes behind Tor have a reasonable chance of completing
	// the negotiation.
	NegotiateDownloadTime = 600 * time.Second

	// NegotiateMaxErrorSize indicates the maximum number of bytes that can be
	// used to encode an error being sent during negotiation.
	NegotiateMaxErrorSize = 256

	// NegotiateMaxSiaPubkeySize defines the maximum size that a SiaPubkey is
	// allowed to be when being sent over the wire during negotiation.
	NegotiateMaxSiaPubkeySize = 1e3

	// NegotiateMaxTransactionSignatureSize defines the maximum size that a
	// transaction signature is allowed to be when being sent over the wire
	// during negotiation.
	NegotiateMaxTransactionSignatureSize = 2e3

	// NegotiateMaxTransactionSignaturesSize defines the maximum size that a
	// transaction signature slice is allowed to be when being sent over the
	// wire during negotiation.
	NegotiateMaxTransactionSignaturesSize = 5e3
)

var (

	// ErrStopResponse is the error returned by ReadNegotiationAcceptance when
	// it reads the StopResponse string.
	ErrStopResponse = errors.New("sender wishes to stop communicating")
)

// ReadNegotiationAcceptance reads an accept/reject response from r (usually a
// net.Conn). If the response is not AcceptResponse, ReadNegotiationAcceptance
// returns the response as an error. If the response is StopResponse,
// ErrStopResponse is returned, allowing for direct error comparison.
//
// Note that since errors returned by ReadNegotiationAcceptance are newly
// allocated, they cannot be compared to other errors in the traditional
// fashion.
func ReadNegotiationAcceptance(r io.Reader) error {
	var resp string
	err := siabin.ReadObject(r, &resp, NegotiateMaxErrorSize)
	if err != nil {
		return err
	}
	switch resp {
	case AcceptResponse:
		return nil
	case StopResponse:
		return ErrStopResponse
	default:
		return errors.New(resp)
	}
}

// WriteNegotiationAcceptance writes the 'accept' response to w (usually a
// net.Conn).
func WriteNegotiationAcceptance(w io.Writer) error {
	return siabin.WriteObject(w, AcceptResponse)
}

// WriteNegotiationRejection will write a rejection response to w (usually a
// net.Conn) and return the input error. If the write fails, the write error
// is joined with the input error.
func WriteNegotiationRejection(w io.Writer, err error) error {
	writeErr := siabin.WriteObject(w, err.Error())
	if writeErr != nil {
		return build.JoinErrors([]error{err, writeErr}, "; ")
	}
	return err
}

// WriteNegotiationStop writes the 'stop' response to w (usually a
// net.Conn).
func WriteNegotiationStop(w io.Writer) error {
	return siabin.WriteObject(w, StopResponse)
}
