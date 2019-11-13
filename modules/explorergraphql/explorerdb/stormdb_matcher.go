package explorerdb

import (
	"fmt"

	"github.com/asdine/storm/q"
)

type transactionIDLengthMatcher struct {
	matcher func(int) bool
}

var (
	_ q.Matcher = (*transactionIDLengthMatcher)(nil)
)

func newTransactionIDLengthMatcher(filter *IntFilterRange) *transactionIDLengthMatcher {
	if filter == nil {
		return nil
	}
	var matcher func(int) bool
	if filter.Min == nil {
		if filter.Max == nil {
			return nil
		}
		matcher = func(l int) bool {
			return l <= *filter.Max
		}
	} else if filter.Max == nil {
		matcher = func(l int) bool {
			return l >= *filter.Min
		}
	} else {
		matcher = func(l int) bool {
			return l >= *filter.Min && l <= *filter.Max
		}
	}
	return &transactionIDLengthMatcher{matcher: matcher}
}

// Match is used to test the criteria against a structure.
func (m *transactionIDLengthMatcher) Match(v interface{}) (bool, error) {
	block, ok := v.(StormBlock)
	if !ok {
		return false, fmt.Errorf("transactionIDLengthMatcher: unexpected value %[1]v (%[1]T) to match against", v)
	}
	return m.matcher(len(block.Transactions)), nil
}
