package client

import (
	"math/big"
	"strings"
	"testing"

	"github.com/rivine/rivine/types"
)

func TestNewCurrencyConvertor(t *testing.T) {
	testCases := []struct {
		Input  types.CurrencyUnits
		Output *CurrencyConvertor
	}{
		{
			types.CurrencyUnits{
				OneCoin: types.NewCurrency(new(big.Int).Exp(big.NewInt(10), big.NewInt(9), nil)),
			},
			&CurrencyConvertor{
				scalar:    big.NewInt(1000000000),
				precision: 9,
			},
		},
		{
			types.CurrencyUnits{
				OneCoin: types.NewCurrency(new(big.Int).Exp(big.NewInt(10), big.NewInt(0), nil)),
			},
			&CurrencyConvertor{
				scalar:    big.NewInt(1),
				precision: 0,
			},
		},
		{
			types.CurrencyUnits{
				OneCoin: types.NewCurrency(new(big.Int).Exp(big.NewInt(10), big.NewInt(24), nil)),
			},
			&CurrencyConvertor{
				scalar:    new(big.Int).Exp(big.NewInt(10), big.NewInt(24), nil),
				precision: 24,
			},
		},
		{
			types.CurrencyUnits{
				OneCoin: types.NewCurrency(big.NewInt(1)),
			},
			&CurrencyConvertor{
				scalar:    big.NewInt(1),
				precision: 0,
			},
		},
	}
	for idx, testCase := range testCases {
		cc, err := NewCurrencyConvertor(testCase.Input)
		if testCase.Output == nil {
			if err == nil {
				t.Errorf("expected error for testCase #%d, but nothing received", idx)
			}
			continue
		}

		if testCase.Output.Cmp(cc) != 0 {
			t.Errorf("#%d: %v != %v", idx, testCase.Output, cc)
		}
	}
}

func TestParseCoinStringInvalidStrings(t *testing.T) {
	cc, err := NewCurrencyConvertor(types.DefaultCurrencyUnits())
	if err != nil {
		t.Fatal(err)
	}

	testCases := []string{
		"-1",
		"",
		"1..0",
		"1.-1",
	}
	for idx, testCase := range testCases {
		x, err := cc.ParseCoinString(testCase)
		if err == nil {
			t.Error(idx, "expected to not parse, but it did", x)
		}
	}
}

func TestParseCoinStringToCoinString_E0(t *testing.T) {
	cc, err := NewCurrencyConvertor(types.CurrencyUnits{
		OneCoin: types.NewCurrency(new(big.Int).Exp(big.NewInt(10), big.NewInt(0), nil)),
	})
	if err != nil {
		t.Fatal(err)
	}
	testParseCoinStringToCoinString(t, cc)
}

func TestParseCoinStringToCoinString_E9(t *testing.T) {
	cc, err := NewCurrencyConvertor(types.CurrencyUnits{
		OneCoin: types.NewCurrency(new(big.Int).Exp(big.NewInt(10), big.NewInt(9), nil)),
	})
	if err != nil {
		t.Fatal(err)
	}
	testParseCoinStringToCoinString(t, cc)
}

func TestParseCoinStringToCoinString_E24(t *testing.T) {
	cc, err := NewCurrencyConvertor(types.CurrencyUnits{
		OneCoin: types.NewCurrency(new(big.Int).Exp(big.NewInt(10), big.NewInt(24), nil)),
	})
	if err != nil {
		t.Fatal(err)
	}
	testParseCoinStringToCoinString(t, cc)
}

func testParseCoinStringToCoinString(t *testing.T, cc CurrencyConvertor) {
	for i := uint(0); i < cc.precision; i++ {
		var str string
		if i == 0 {
			str = "1"
		} else {
			str = "0."
			str += strings.Repeat("0", int(i-1))
			str += "1"
		}

		c, err := cc.ParseCoinString(str)
		if err != nil {
			t.Error(i, err)
			continue
		}
		expected := types.NewCurrency(big.NewInt(10).Add(new(big.Int),
			big.NewInt(10).Exp(big.NewInt(10), big.NewInt(int64(cc.precision-i)), nil)))
		if expected.Cmp(c) != 0 {
			t.Errorf("#%d: %v != %v", i, expected, c)
			continue
		}

		outStr := cc.ToCoinString(c)
		outStr2 := cc.ToCoinString(expected)
		if outStr != outStr2 {
			t.Errorf("#%d: %v != %v", i, outStr, outStr2)
			continue
		}
		if outStr != str {
			t.Errorf("#%d: %v != %v", i, outStr, str)
		}
	}
}

func (cc CurrencyConvertor) Cmp(other CurrencyConvertor) int {
	if cc.precision < other.precision {
		return -1
	}
	if cc.precision > other.precision {
		return 1
	}
	return strings.Compare(cc.scalar.String(), other.scalar.String())
}
