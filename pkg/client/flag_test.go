package client

import (
	"testing"

	"github.com/rivine/rivine/types"
)

func TestCurrencyFlagPanics(t *testing.T) {
	assertPanic(t, func() {
		currencyFlag(nil, new(types.Currency))
	})
	assertPanic(t, func() {
		currencyFlag(new(CommandLineClient), nil)
	})
	assertNoPanic(t, func() {
		currencyFlag(new(CommandLineClient), new(types.Currency))
	})
}

func TestCurrencyFlag(t *testing.T) {
	var value types.Currency
	flag := currencyFlag(defaultCurrencyConvertorFactory{}, &value)
	if !value.IsZero() {
		t.Fatal("value should still be zero but is:", value.String())
	}
	str := flag.String()
	if str != "0" {
		t.Fatal(`stringified value should equal "0" but equals instead:`, str)
	}
	err := flag.Set("1")
	if err != nil {
		t.Fatal("failed to set value to the value of one coin:", err)
	}
	expected := types.DefaultCurrencyUnits().OneCoin
	if expected.IsZero() || expected.Equals64(1) {
		t.Fatal(
			"one coin isn't expected to equal zero or one, but some power —greater than 1— of 10, instead is: ",
			expected.String())
	}
	if !value.Equals(expected) {
		t.Fatal("unexpected value:", value.String(), "!=", expected.String())
	}
	str = flag.String()
	expectedStr := createDefaultCurrencyConvertor().ToCoinString(expected)
	if str != expectedStr {
		t.Fatal("unexpected stringified value: `" + str + `" != "` + expectedStr + `"`)
	}
}

type defaultCurrencyConvertorFactory struct{}

func (dccf defaultCurrencyConvertorFactory) CreateCurrencyConvertor() CurrencyConvertor {
	return createDefaultCurrencyConvertor()
}

func assertPanic(t *testing.T, f func()) {
	defer func() {
		p := recover()
		if p == nil {
			t.Error("expected panic but none received")
		}
	}()
	f()
}

func assertNoPanic(t *testing.T, f func()) {
	defer func() {
		p := recover()
		if p != nil {
			t.Errorf("expected no panic but received one: %v", p)
		}
	}()
	f()
}
