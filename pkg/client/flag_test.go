package client

import (
	"testing"

	"github.com/rivine/rivine/types"
)

func TestCoinFlag(t *testing.T) {
	// test coin->string->coin->string
	coins := []coinFlag{
		{},
		{types.NewCurrency64(1)},
		{types.NewCurrency64(42)},
		{types.NewCurrency64(1234567890123456789)},
	}
	for idx, coin := range coins {
		str := coin.String()
		err := coin.Set(str)
		if err != nil {
			t.Errorf("error while setting string for coin #%d: %v (input: '%s')", idx, err, str)
		}
		str2 := coin.String()
		if str != str2 {
			t.Errorf("coin #%d string loading isn't deterministic: %s != %s", idx, str, str2)
		}
	}

	// test string->loader->string
	testCases := []string{
		"0.00001",
		"1",
		"0",
		"123456",
		"42",
		"12345.67889",
	}
	for idx, testCase := range testCases {
		var cf coinFlag
		err := cf.Set(testCase)
		if err != nil {
			t.Errorf("error while loading string for coin flag #%d: %v", idx, err)
		}
		str := cf.String()
		if testCase != str {
			t.Errorf("coin flag #%d string loading isn't deterministic: %s != %s", idx, testCase, str)
		}
	}
}
