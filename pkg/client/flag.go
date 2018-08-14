package client

import (
	"fmt"
	"os"

	"github.com/rivine/rivine/types"
	"github.com/spf13/pflag"
)

func currencyFlag(ccFactory currencyConvertorFactory, value *types.Currency) *coinFlag {
	if ccFactory == nil {
		panic("no currency convertor factory given")
	}
	if value == nil {
		panic("nil currency value reference given")
	}
	return &coinFlag{
		ccFactory: ccFactory,
		value:     value,
	}
}

type coinFlag struct {
	ccFactory currencyConvertorFactory
	value     *types.Currency
}

type currencyConvertorFactory interface {
	CreateCurrencyConvertor() CurrencyConvertor
}

// String implements pflag.Value.String
func (c *coinFlag) String() string {
	cc := c.ccFactory.CreateCurrencyConvertor()
	return cc.ToCoinString(*c.value)
}

// Set implements pflag.Value.Set
func (c *coinFlag) Set(s string) error {
	cc := c.ccFactory.CreateCurrencyConvertor()
	value, err := cc.ParseCoinString(s)
	if err != nil {
		return err
	}
	*c.value = value
	return nil
}

// Type implements pflag.Value.Type
func (c *coinFlag) Type() string {
	return "Coin"
}

var (
	_ pflag.Value = (*coinFlag)(nil)
)

func parseCoinArg(cc CurrencyConvertor, str string) types.Currency {
	amount, err := cc.ParseCoinString(str)
	if err != nil {
		fmt.Fprintln(os.Stderr, cc.CoinArgDescription("amount"))
		DieWithExitCode(ExitCodeUsage, "failed to parse coin-typed argument: ", err)
		return types.Currency{}
	}
	return amount
}
