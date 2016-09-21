package main

import (
	"math/big"
	"testing"

	"github.com/rivine/rivine/types"
)

func TestCurrencyUnits(t *testing.T) {
	tests := []struct {
		in, out string
	}{
		{"1", "1 H"},
		{"1000", "1000 H"},
		{"100000000000", "100000000000 H"},
		{"1000000000000", "1 pS"},
		{"1234560000000", "1.235 pS"},
		{"12345600000000", "12.35 pS"},
		{"123456000000000", "123.5 pS"},
		{"1000000000000000", "1 nS"},
		{"1000000000000000000", "1 uS"},
		{"1000000000000000000000", "1 mS"},
		{"1000000000000000000000000", "1 SC"},
		{"1000000000000000000000000000", "1 KS"},
		{"1000000000000000000000000000000", "1 MS"},
		{"1000000000000000000000000000000000", "1 GS"},
		{"1000000000000000000000000000000000000", "1 TS"},
		{"1234560000000000000000000000000000000", "1.235 TS"},
		{"1234560000000000000000000000000000000000", "1235 TS"},
	}
	for _, test := range tests {
		i, _ := new(big.Int).SetString(test.in, 10)
		out := currencyUnits(types.NewCurrency(i))
		if out != test.out {
			t.Errorf("currencyUnits(%v): expected %v, got %v", test.in, test.out, out)
		}
	}
}
