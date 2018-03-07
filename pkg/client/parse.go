package client

import (
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/rivine/rivine/types"
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

// CurrencyUnits converts a types.Currency to a string with human-readable
// units. The unit used will be the largest unit that results in a value
// greater than 1. The value is rounded to 4 significant digits.
func CurrencyUnits(c types.Currency) string {
	pico := types.OneCoin.Div64(1e12)
	if c.Cmp(pico) < 0 {
		return c.String() + " H"
	}

	// iterate until we find a unit greater than c
	mag := pico
	unit := ""
	for _, unit = range []string{"p", "n", "u", "m", "C", "K", "M", "G", "T"} {
		if c.Cmp(mag.Mul64(1e3)) < 0 {
			break
		} else if unit != "T" {
			// don't want to perform this multiply on the last iter; that
			// would give us 1.235 TS instead of 1235 TS
			mag = mag.Mul64(1e3)
		}
	}

	num := new(big.Rat).SetInt(c.Big())
	denom := new(big.Rat).SetInt(mag.Big())
	res, _ := new(big.Rat).Mul(num, denom.Inv(denom)).Float64()

	return fmt.Sprintf("%.4g %s", res, unit)
}

// ParseCurrency converts a siacoin amount to base units.
func ParseCurrency(amount string) (string, error) {
	units := []string{"p", "n", "u", "m", "C", "K", "M", "G", "T"}
	for i, unit := range units {
		if strings.HasSuffix(amount, unit) {
			// scan into big.Rat
			r, ok := new(big.Rat).SetString(strings.TrimSuffix(amount, unit))
			if !ok {
				return "", errors.New("malformed amount")
			}
			// convert units
			exp := 24 + 3*(int64(i)-4)
			mag := new(big.Int).Exp(big.NewInt(10), big.NewInt(exp), nil)
			r.Mul(r, new(big.Rat).SetInt(mag))
			// r must be an integer at this point
			if !r.IsInt() {
				return "", errors.New("non-integer number of hastings")
			}
			return r.RatString(), nil
		}
	}
	// check for hastings separately
	if strings.HasSuffix(amount, "H") {
		return strings.TrimSuffix(amount, "H"), nil
	}

	return "", errors.New("amount is missing units; run 'wallet --help' for a list of units")
}

// YesNo returns "Yes" if b is true, and "No" if b is false.
func YesNo(b bool) string {
	if b {
		return "Yes"
	}
	return "No"
}
