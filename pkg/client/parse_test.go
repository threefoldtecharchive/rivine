package client

import (
	"math/big"
	"strings"
	"testing"

	"github.com/rivine/rivine/types"
)

func TestParseCoinStringInvalidStrings(t *testing.T) {
	bchainInfo := types.DefaultBlockchainInfo()
	cc := NewCurrencyConvertor(types.DefaultCurrencyUnits(), bchainInfo.CoinUnit)

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

func TestParseCoinStringValidStrings(t *testing.T) {
	bchainInfo := types.DefaultBlockchainInfo()
	cc := NewCurrencyConvertor(types.DefaultCurrencyUnits(), bchainInfo.CoinUnit)

	testCases := []string{
		"1",
		"1.1",
		"1.123",
		"1.123456",
		"1.123456789",
		"123456789.987654321",
		"1.1234567890",
		"1.123456789000",
	}
	for idx, testCase := range testCases {
		x, err := cc.ParseCoinString(testCase)
		if err != nil {
			t.Error(idx, "expected to parse, but it didn't", err)
			continue
		}

		str := cc.ToCoinString(x)
		strippedTestCase := strings.TrimRight(testCase, "0")
		if str != strippedTestCase {
			t.Error(idx, str, "!=", strippedTestCase)
		}
	}
}

func TestParseCoinStringToCoinSmallValueString_E0(t *testing.T) {
	bchainInfo := types.DefaultBlockchainInfo()
	cc := NewCurrencyConvertor(types.CurrencyUnits{
		OneCoin: types.NewCurrency(new(big.Int).Exp(big.NewInt(10), big.NewInt(0), nil)),
	}, bchainInfo.CoinUnit)
	testParseCoinStringToCoinSmallValueString(t, cc)
}

func TestParseCoinStringToCoinSmallValueString_E9(t *testing.T) {
	bchainInfo := types.DefaultBlockchainInfo()
	cc := NewCurrencyConvertor(types.CurrencyUnits{
		OneCoin: types.NewCurrency(new(big.Int).Exp(big.NewInt(10), big.NewInt(9), nil)),
	}, bchainInfo.CoinUnit)
	testParseCoinStringToCoinSmallValueString(t, cc)
}

func TestParseCoinStringToCoinSmallValueString_E24(t *testing.T) {
	bchainInfo := types.DefaultBlockchainInfo()
	cc := NewCurrencyConvertor(types.CurrencyUnits{
		OneCoin: types.NewCurrency(new(big.Int).Exp(big.NewInt(10), big.NewInt(24), nil)),
	}, bchainInfo.CoinUnit)
	testParseCoinStringToCoinSmallValueString(t, cc)
}

func testParseCoinStringToCoinSmallValueString(t *testing.T, cc CurrencyConvertor) {
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

func TestParseCoinStringToCoinBigValueString_E0(t *testing.T) {
	bchainInfo := types.DefaultBlockchainInfo()
	cc := NewCurrencyConvertor(types.CurrencyUnits{
		OneCoin: types.NewCurrency(new(big.Int).Exp(big.NewInt(10), big.NewInt(0), nil)),
	}, bchainInfo.CoinUnit)
	testParseCoinStringToCoinBigValueString(t, cc)
}

func TestParseCoinStringToCoinBigValueString_E9(t *testing.T) {
	bchainInfo := types.DefaultBlockchainInfo()
	cc := NewCurrencyConvertor(types.CurrencyUnits{
		OneCoin: types.NewCurrency(new(big.Int).Exp(big.NewInt(10), big.NewInt(9), nil)),
	}, bchainInfo.CoinUnit)
	testParseCoinStringToCoinBigValueString(t, cc)
}

func TestParseCoinStringToCoinBigValueString_E24(t *testing.T) {
	bchainInfo := types.DefaultBlockchainInfo()
	cc := NewCurrencyConvertor(types.CurrencyUnits{
		OneCoin: types.NewCurrency(new(big.Int).Exp(big.NewInt(10), big.NewInt(24), nil)),
	}, bchainInfo.CoinUnit)
	testParseCoinStringToCoinBigValueString(t, cc)
}

func testParseCoinStringToCoinBigValueString(t *testing.T, cc CurrencyConvertor) {
	for i := uint(0); i < cc.precision; i++ {
		str := "1"
		str += strings.Repeat("0", int(i))

		c, err := cc.ParseCoinString(str)
		if err != nil {
			t.Error(i, err)
			continue
		}
		expected := types.NewCurrency(big.NewInt(10).Add(new(big.Int),
			big.NewInt(10).Exp(big.NewInt(10), big.NewInt(int64(cc.precision+i)), nil)))
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
