package explorergraphql

import (
	"bytes"
	"errors"
	"math/big"
)

func filterFunctionForBinaryDataFilter(binaryDataFilter *BinaryDataFilter) (func([]byte) bool, error) {
	if binaryDataFilter.StartsWith != nil {
		if binaryDataFilter.Contains != nil || binaryDataFilter.EndsWith != nil || binaryDataFilter.EqualTo != nil {
			return nil, errors.New("binary data filter can only use one option, multiple cannot be combined")
		}
		prefix := *binaryDataFilter.StartsWith
		return func(b []byte) bool {
			return bytes.HasPrefix(b, prefix)
		}, nil
	}
	if binaryDataFilter.Contains != nil {
		// StartsWith is already confirmed by previous check to be nil
		if binaryDataFilter.EndsWith != nil || binaryDataFilter.EqualTo != nil {
			return nil, errors.New("binary data filter can only use one option, multiple cannot be combined")
		}
		subslice := *binaryDataFilter.Contains
		return func(b []byte) bool {
			return bytes.Contains(b, subslice)
		}, nil
	}
	if binaryDataFilter.EndsWith != nil {
		// StartsWith and Contains are already confirmed by previous checks to be nil
		if binaryDataFilter.EqualTo != nil {
			return nil, errors.New("binary data filter can only use one option, multiple cannot be combined")
		}
		suffix := *binaryDataFilter.EndsWith
		return func(b []byte) bool {
			return bytes.HasSuffix(b, suffix)
		}, nil
	}
	if binaryDataFilter.EqualTo != nil {
		// All others are already confirmed by previous checks to be nil
		a := *binaryDataFilter.EqualTo
		return func(b []byte) bool {
			return bytes.Equal(b, a)
		}, nil
	}
	return nil, nil
}

func filterFunctionForBigIntFilter(bigIntFilter *BigIntFilter) (func(BigInt) bool, error) {
	// define min-max range
	var (
		min, max *BigInt
	)
	if bigIntFilter.LessThan != nil {
		if bigIntFilter.LessThanOrEqualTo != nil || bigIntFilter.EqualTo != nil || bigIntFilter.GreaterThanOrEqualTo != nil || bigIntFilter.GreaterThan != nil {
			return nil, errors.New("big int filter can only use one option, multiple cannot be combined")
		}
		max = bigIntFilter.LessThan
		max.Sub(max.Int, big.NewInt(1))
	} else if bigIntFilter.LessThanOrEqualTo != nil {
		// LessThan is already confirmed by previous check to be nil
		if bigIntFilter.EqualTo != nil || bigIntFilter.GreaterThanOrEqualTo != nil || bigIntFilter.GreaterThan != nil {
			return nil, errors.New("big int filter can only use one option, multiple cannot be combined")
		}
		max = bigIntFilter.LessThanOrEqualTo
	} else if bigIntFilter.EqualTo != nil {
		// LessThan and LessThanOrEqualTo are already confirmed by previous check to be nil
		if bigIntFilter.GreaterThanOrEqualTo != nil || bigIntFilter.GreaterThan != nil {
			return nil, errors.New("big int filter can only use one option, multiple cannot be combined")
		}
		min = bigIntFilter.EqualTo
		max = bigIntFilter.EqualTo
	} else if bigIntFilter.GreaterThanOrEqualTo != nil {
		// LessThan, LessThanOrEqualTo and EqualTo are already confirmed by previous check to be nil
		if bigIntFilter.GreaterThan != nil {
			return nil, errors.New("big int filter can only use one option, multiple cannot be combined")
		}
		min = bigIntFilter.GreaterThanOrEqualTo
	} else if bigIntFilter.GreaterThan != nil {
		// all others are already confirmed by previous check to be nil
		min = bigIntFilter.GreaterThan
		min.Add(min.Int, big.NewInt(1))
	}

	// return filter function based on min/max range
	if min == nil {
		if max == nil {
			return nil, nil // nothing to filter
		}
		return func(bi BigInt) bool {
			return bi.Cmp(max.Int) <= 0
		}, nil
	}
	// min != nil
	if max == nil {
		return func(bi BigInt) bool {
			return bi.Cmp(min.Int) >= 0
		}, nil
	}
	if min.Cmp(max.Int) == 0 {
		return func(bi BigInt) bool {
			return bi.Cmp(min.Int) == 0
		}, nil
	}
	return func(bi BigInt) bool {
		return bi.Cmp(max.Int) <= 0 && bi.Cmp(min.Int) >= 0
	}, nil
}
