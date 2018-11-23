package types

import (
	"errors"
	"io"
	"unicode/utf8"

	"github.com/threefoldtech/rivine/pkg/encoding/rivbin"

	"github.com/threefoldtech/rivine/pkg/encoding/siabin"
)

// ArbitraryDataType defines the optional type of ArbitraryData,
// First 128 values —[0,127]— are reserved by Rivine,
// while the other 128 values —[128,255]— can be used by applications as seen fit.
type ArbitraryDataType uint8

// Rivine-Defined ArbitraryData Types within the reserved [0,127] space.
const (
	ArbitraryDataTypeBinary ArbitraryDataType = iota
	ArbitraryDataTypeUTF8
)

// ArbitraryData represents a strutured pair of Arbitrary Data and its optional Typing.
type ArbitraryData struct {
	Data []byte            `json:"data"`
	Type ArbitraryDataType `json:"type"`
}

// Validate if this arbitrary data is valid according to its type.
func (ad ArbitraryData) Validate() error {
	if len(ad.Data) == 0 {
		return nil
	}
	if ad.Type == ArbitraryDataTypeUTF8 && !utf8.Valid(ad.Data) {
		return errors.New("utf-8 typed arbitrary data is invalid")
	}
	return nil
}

const (
	_siaEncodingArbitraryDataPrefixSize = 8
	_siaEncodingArbitraryDataTypePos    = 3
)

// MarshalSia implements siabin.SiaMarshaler.MarshalSia
func (ad ArbitraryData) MarshalSia(w io.Writer) error {
	// encode data length as prefix and merge optional type into it
	dataLength := len(ad.Data)
	prefix := siabin.EncUint64(uint64(dataLength))
	prefix[_siaEncodingArbitraryDataTypePos] = byte(ad.Type)
	// write prefix
	n, err := w.Write(prefix)
	if err != nil {
		return err
	}
	if n != _siaEncodingArbitraryDataPrefixSize {
		return io.ErrShortWrite
	}
	// write arbitrary data directly
	n, err = w.Write(ad.Data)
	if err != nil {
		return err
	}
	if n != dataLength {
		return io.ErrShortWrite
	}
	// all written as expected
	return nil
}

// UnmarshalSia implements siabin.SiaMarshaler.UnmarshalSia
func (ad *ArbitraryData) UnmarshalSia(r io.Reader) error {
	// read length and optional type
	var prefix [_siaEncodingArbitraryDataPrefixSize]byte
	_, err := io.ReadFull(r, prefix[:])
	if err != nil {
		return err
	}
	// pop optional type
	ad.Type = ArbitraryDataType(prefix[_siaEncodingArbitraryDataTypePos])
	prefix[_siaEncodingArbitraryDataTypePos] = 0
	// decode length
	length := siabin.DecUint64(prefix[:])
	// ensure length is not greater than sia-defined slice length
	if length > siabin.MaxSliceSize {
		return siabin.ErrSliceTooLarge
	}
	if length == 0 {
		ad.Data = nil
		return nil // nothing more to do
	}
	// read arbitrary data
	ad.Data = make([]byte, length)
	_, err = io.ReadFull(r, ad.Data)
	if err != nil {
		return err
	}
	// validate arbitrary data and return
	return ad.Validate()
}

// MarshalRivine implements rivbin.RivineMarshaler.MarshalRivine
func (ad ArbitraryData) MarshalRivine(w io.Writer) error {
	encoder := rivbin.NewEncoder(w)
	err := encoder.Encode(ad.Data)
	if err != nil {
		return err
	}
	if len(ad.Data) == 0 {
		return nil // nothing to do
	}
	// encode optional type only when there is data
	return encoder.Encode(ad.Type)
}

// UnmarshalRivine implements rivbin.RivineMarshaler.UnmarshalRivine
func (ad *ArbitraryData) UnmarshalRivine(r io.Reader) error {
	decoder := rivbin.NewDecoder(r)
	err := decoder.Decode(&ad.Data)
	if err != nil {
		return err
	}
	if len(ad.Data) == 0 {
		return nil // nothing to do
	}
	// decode the optional type
	err = decoder.Decode(&ad.Type)
	if err != nil {
		return err
	}
	// validate arbitrary data and return
	return ad.Validate()
}
