package client

import (
	"fmt"
	"math/big"
	"strings"
	"testing"

	"github.com/threefoldtech/rivine/types"
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
		"123,456,789.987654321",
		"1.1234567890",
		"1.123456789000",
		"12.345678900000",
		"123.4567",
		"1,234.8945",
		"12,345.4944100",
		"166,198,297,161.1100",
		"1,234,567,890.00000",
		"12,987,654,321.002650",
	}
	for idx, testCase := range testCases {
		x, err := cc.ParseCoinString(testCase)
		if err != nil {
			t.Error(idx, "expected to parse, but it didn't", err)
			continue
		}

		str := cc.ToCoinString(x)
		strippedTestCase := strings.TrimRight(strings.TrimRight(testCase, "0"), ".")
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
		// make sure to insert proper formatting
		formatterStartingOffset := int((i + 1) % 3)
		for j := 1; j < int(i)+1; j++ {
			if (j-formatterStartingOffset)%3 == 0 {
				fmt.Println(j, formatterStartingOffset, j-formatterStartingOffset)
				str += ","
			}
			str += "0"
		}

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
