package client

import (
	"github.com/rivine/rivine/types"
	"github.com/spf13/pflag"
)

type coinFlag struct {
	Amount types.Currency
}

// String implements pflag.Value.String,
// printing this LockTime either as a timestamp in DateOnlyLayout or RFC822 layout,
// a duration or as an uint64.
func (c coinFlag) String() string {
	return _CurrencyConvertor.ToCoinString(c.Amount)
}

// Set implements pflag.Value.Set,
// which parses the given string either as a timestamp in DateOnlyLayout or RFC822 layout,
// a duration or as an uint64.
func (c *coinFlag) Set(s string) (err error) {
	c.Amount, err = _CurrencyConvertor.ParseCoinString(s)
	return
}

// Type implements pflag.Value.Type
func (c coinFlag) Type() string {
	return "Coin"
}

var (
	_ pflag.Value = (*coinFlag)(nil)
)
