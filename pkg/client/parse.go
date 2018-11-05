package client

import (
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/threefoldtech/rivine/types"
)

var errUnableToParseSize = errors.New("unable to parse size")

// PeriodUnits turns a period in terms of blocks to a number of weeks.
func PeriodUnits(blocks types.BlockHeight) string {
	return fmt.Sprint(blocks / 1008) // 1008 blocks per week
}

// ParsePeriod converts a number of weeks to a number of blocks.
func ParsePeriod(period string) (string, error) {
	var weeks float64
	_, err := fmt.Sscan(period, &weeks)
	if err != nil {
		return "", errUnableToParseSize
	}
	blocks := int(weeks * 1008) // 1008 blocks per week
	return fmt.Sprint(blocks), nil
}

// YesNo returns "Yes" if b is true, and "No" if b is false.
func YesNo(b bool) string {
	if b {
		return "Yes"
	}
	return "No"
}

// CurrencyConvertor is used to parse a currency in it's default unit,
// and turn it into its in-memory smallest unit. Simiarly it allow you to
// turn the in-memory smallest unit into a string version of the default init.
type CurrencyConvertor struct {
	scalar    *big.Int
	precision uint // amount of zeros after the comma
	coinUnit  string
}

// NewCurrencyConvertor creates a new currency convertor
// using the given currency units.
//
// See CurrencyConvertor for more information.
func NewCurrencyConvertor(units types.CurrencyUnits, coinUnit string) CurrencyConvertor {
	oneCoinStr := units.OneCoin.String()
	precision := uint(len(oneCoinStr) - 1)
	return CurrencyConvertor{
		scalar:    units.OneCoin.Big(),
		precision: precision,
		coinUnit:  coinUnit,
	}
}

// ParseCoinString parses the given string assumed to be in the default unit,
// and parses it into an in-memory currency unit of the smallest unit.
// It will fail if the given string is invalid or too precise.
func (cc CurrencyConvertor) ParseCoinString(str string) (types.Currency, error) {
	initialParts := strings.SplitN(str, ".", 2)
	if len(initialParts) == 1 {
		// a round number, simply multiply and go
		i, ok := big.NewInt(0).SetString(initialParts[0], 10)
		if !ok {
			return types.Currency{}, errors.New("invalid round currency coin amount")
		}
		if i.Cmp(big.NewInt(0)) == -1 {
			return types.Currency{}, errors.New("invalid round currency coin amount: cannot be negative")
		}
		return types.NewCurrency(i.Mul(i, cc.scalar)), nil
	}

	whole := initialParts[0]
	dac := initialParts[1]
	sn := uint(cc.precision)
	if l := uint(len(dac)); l < sn {
		sn = l
	}
	whole += initialParts[1][:sn]
	dac = dac[sn:]
	for i := range dac {
		if dac[i] != '0' {
			return types.Currency{}, errors.New("invalid or too precise currency coin amount")
		}
	}
	i, ok := big.NewInt(0).SetString(whole, 10)
	if !ok {
		return types.Currency{}, errors.New("invalid currency coin amount")
	}
	if i.Cmp(big.NewInt(0)) == -1 {
		return types.Currency{}, errors.New("invalid round currency coin amount: cannot be negative")
	}
	i.Mul(i, big.NewInt(0).Exp(
		big.NewInt(10), big.NewInt(int64(cc.precision-sn)), nil))
	c := types.NewCurrency(i)
	if c.Cmp64(0) == -1 {
		return types.Currency{}, errors.New("invalid round currency coin amount: cannot be negative")
	}
	return c, nil
}

// ToCoinString turns the in-memory currency unit,
// into a string version of the default currency unit.
// This can never fail, as the only thing it can do is make a number smaller.
func (cc CurrencyConvertor) ToCoinString(c types.Currency) string {
	if c.Equals64(0) {
		return "0"
	}

	str := c.String()
	if cc.precision == 0 {
		return str
	}
	l := uint(len(str))
	if l > cc.precision {
		idx := l - cc.precision
		str = strings.TrimRight(str[:idx]+"."+str[idx:], "0")
		str = strings.TrimRight(str, ".")
		if len(str) == 0 {
			return "0"
		}
		return str
	}
	str = "0." + strings.Repeat("0", int(cc.precision-l)) + str
	str = strings.TrimRight(str, "0")
	str = strings.TrimRight(str, ".")
	return str
}

// ToCoinStringWithUnit turns the in-memory currency unit,
// into a string version of the default currency unit.
// This can never fail, as the only thing it can do is make a number smaller.
// It also adds the unit of the coin behind the coin.
func (cc CurrencyConvertor) ToCoinStringWithUnit(c types.Currency) string {
	return cc.ToCoinString(c) + " " + cc.coinUnit
}

// CoinArgDescription is used to print a helpful arg description message,
// for this convertor.
func (cc CurrencyConvertor) CoinArgDescription(argName string) string {
	if cc.precision < 1 {
		return fmt.Sprintf(
			"argument %s (expressed in default unit %s) has to be a positive natural number (no digits after comma are allowed)",
			argName, cc.coinUnit)
	}
	return fmt.Sprintf(
		"argument %s (expressed in default unit %s) can (only) have up to %d digits after comma and has to be positive",
		argName, cc.coinUnit, cc.precision)
}
