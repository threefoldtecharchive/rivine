package modules

import (
	"bytes"
	"testing"
)

// TestNegotiationResponses tests the WriteNegotiationAcceptance,
// WriteNegotiationRejection, and ReadNegotiationAcceptance functions.
func TestNegotiationResponses(t *testing.T) {
	// Write/Read acceptance
	buf := new(bytes.Buffer)
	err := WriteNegotiationAcceptance(buf)
	if err != nil {
		t.Fatal(err)
	}
	err = ReadNegotiationAcceptance(buf)
	if err != nil {
		t.Fatal(err)
	}

	// Write/Read rejection
	buf.Reset()
	err = WriteNegotiationRejection(buf, ErrLowBalance)
	if err != ErrLowBalance {
		t.Fatal(err)
	}
	err = ReadNegotiationAcceptance(buf)
	// can't compare to ErrLowBalance directly; contents are the same, but pointer is different
	if err == nil || err.Error() != ErrLowBalance.Error() {
		t.Fatal(err)
	}

	// Write/Read StopResponse
	buf.Reset()
	err = WriteNegotiationStop(buf)
	if err != nil {
		t.Fatal(err)
	}
	err = ReadNegotiationAcceptance(buf)
	if err != ErrStopResponse {
		t.Fatal(err)
	}
}
