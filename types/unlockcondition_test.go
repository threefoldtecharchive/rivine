package types

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	mrand "math/rand"
	"strings"
	"testing"

	"github.com/threefoldtech/rivine/crypto"
	"github.com/threefoldtech/rivine/pkg/encoding/rivbin"
	"github.com/threefoldtech/rivine/pkg/encoding/siabin"
)

func TestUnlockConditionSiaEncoding(t *testing.T) {
	testCases := []string{
		// nil condition
		`000000000000000000`,
		// unlock hash condition
		`012100000000000000016363636363636363636363636363636363636363636363636363636363636363`,
		// atomic swap condition
		`026a0000000000000001454545454545454545454545454545454545454545454545454545454545454501636363636363636363636363636363636363636363636363636363636363636378787878787878787878787878787878787878787878787878787878787878781234567812345678`,
		// time lock condition
		`030900000000000000111111111111111100`, // using nil condition
		`032a00000000000000111111111111111101016363636363636363636363636363636363636363636363636363636363636363`, // using (pubKey) unlock hash condition
		// MultiSig condition
		`0452000000000000000200000000000000020000000000000001e89843e4b8231a01ba18b254d530110364432aafab8206bea72e5a20eaa55f7001a6a6c5584b2bfbd08738996cd7930831f958b9a5ed1595525236e861c1a0dc35`,
	}
	for idx, testCase := range testCases {
		b, err := hex.DecodeString(testCase)
		if err != nil {
			t.Error(idx, err)
			continue
		}

		var up UnlockConditionProxy
		err = up.UnmarshalSia(bytes.NewReader(b))
		if err != nil {
			t.Error(idx, err)
			continue
		}

		buf := bytes.NewBuffer(nil)
		err = up.MarshalSia(buf)
		if err != nil {
			t.Error(idx, err)
			continue
		}

		out := hex.EncodeToString(buf.Bytes())
		if out != testCase {
			t.Error(idx, out, "!=", testCase)
		}
	}
}

func TestUnlockConditionRivineEncoding(t *testing.T) {
	testCases := []string{
		// nil condition
		`0000`,
		// unlock hash condition
		`0142016363636363636363636363636363636363636363636363636363636363636363`,
		// atomic swap condition
		`02d401454545454545454545454545454545454545454545454545454545454545454501636363636363636363636363636363636363636363636363636363636363636378787878787878787878787878787878787878787878787878787878787878781234567812345678`,
		// time lock condition
		`0312111111111111111100`, // using nil condition
		`0354111111111111111101016363636363636363636363636363636363636363636363636363636363636363`, // using (pubKey) unlock hash condition
		// MultiSig condition
		`049602000000000000000401e89843e4b8231a01ba18b254d530110364432aafab8206bea72e5a20eaa55f7001a6a6c5584b2bfbd08738996cd7930831f958b9a5ed1595525236e861c1a0dc35`,
	}
	for idx, testCase := range testCases {
		b, err := hex.DecodeString(testCase)
		if err != nil {
			t.Error(idx, err)
			continue
		}

		var up UnlockConditionProxy
		err = up.UnmarshalRivine(bytes.NewReader(b))
		if err != nil {
			t.Error(idx, err)
			continue
		}

		buf := bytes.NewBuffer(nil)
		err = up.MarshalRivine(buf)
		if err != nil {
			t.Error(idx, err)
			continue
		}

		out := hex.EncodeToString(buf.Bytes())
		if out != testCase {
			t.Error(idx, out, "!=", testCase)
		}
	}
}

func TestUnknownUnlockConditionSiaEncoding(t *testing.T) {
	testCases := []string{
		// unknown condition
		`ff0c0000000000000048656c6c6f2c205465737421`,
		// time lock condition
		`0315000000000000001111111111111111ff48656c6c6f2c205465737421`, //using unknown condition
	}
	for idx, testCase := range testCases {
		b, err := hex.DecodeString(testCase)
		if err != nil {
			t.Error(idx, err)
			continue
		}

		var up UnlockConditionProxy
		err = up.UnmarshalSia(bytes.NewReader(b))
		if err == nil {
			t.Error(idx, "error was expected, but non received")
		}
	}
}

func TestUnlockFulfillmentSiaEncoding(t *testing.T) {
	testCases := []string{
		// single signature fulfillment
		`01800000000000000065643235353139000000000000000000200000000000000035fffffffffffffffffffffffffffffffffffffffffffffffff46fffffffffff4000000000000000fffffffffffffffffffffffffffff123ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff`,
		// legacy atomic swap fulfillment
		`020a01000000000000011234567891234567891234567891234567891234567891234567891234567891016363636363636363636363636363636363636363636363636363636363636363bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb07edb85a00000000656432353531390000000000000000002000000000000000abababababababababababababababababababababababababababababababab4000000000000000dededededededededededededededededededededededededededededededededededededededededededededededededededededededededededededededededabadabadabadabadabadabadabadabadabadabadabadabadabadabadabadaba`,
		// atomic swap fulfillment
		`02a000000000000000656432353531390000000000000000002000000000000000fffffffffffffffffffffffffffffffff04fffffffffffffffffffffffffffff4000000000000000ffffffffffffffffffffffff56fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff2ffffffffffffffffff123ffffffffffafffffffffffeffffffffffffff`,
		// MultiSig fulfillment
		`0388000000000000000100000000000000656432353531390000000000000000002000000000000000def123def123def123def123def123def123def123def123def123def123def14000000000000000ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef`,
	}
	for idx, testCase := range testCases {
		b, err := hex.DecodeString(testCase)
		if err != nil {
			t.Error(idx, err)
			continue
		}

		var uf UnlockFulfillmentProxy
		err = uf.UnmarshalSia(bytes.NewReader(b))
		if err != nil {
			t.Error(idx, err)
			continue
		}

		buf := bytes.NewBuffer(nil)
		err = uf.MarshalSia(buf)
		if err != nil {
			t.Error(idx, err)
			continue
		}

		out := hex.EncodeToString(buf.Bytes())
		if out != testCase {
			t.Error(idx, out, "!=", testCase)
		}
	}
}

func TestUnlockFulfillmentRivineEncoding(t *testing.T) {
	testCases := []string{
		// single signature fulfillment
		`01c40135fffffffffffffffffffffffffffffffffffffffffffffffff46fffffffffff80fffffffffffffffffffffffffffff123ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff`,
		// legacy atomic swap fulfillment
		`02b103011234567891234567891234567891234567891234567891234567891234567891016363636363636363636363636363636363636363636363636363636363636363bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb07edb85a0000000001abababababababababababababababababababababababababababababababab80dededededededededededededededededededededededededededededededededededededededededededededededededededededededededededededededededabadabadabadabadabadabadabadabadabadabadabadabadabadabadabadaba`,
		// atomic swap fulfillment
		`02090201fffffffffffffffffffffffffffffffff04fffffffffffffffffffffffffffff80ffffffffffffffffffffffff56fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff2ffffffffffffffffff123ffffffffffafffffffffffeffffffffffffff`,
		// MultiSig fulfillment
		`03c60201def123def123def123def123def123def123def123def123def123def123def180ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef12345ef`,
	}
	for idx, testCase := range testCases {
		b, err := hex.DecodeString(testCase)
		if err != nil {
			t.Error(idx, err)
			continue
		}

		var uf UnlockFulfillmentProxy
		err = uf.UnmarshalRivine(bytes.NewReader(b))
		if err != nil {
			t.Error(idx, err)
			continue
		}

		buf := bytes.NewBuffer(nil)
		err = uf.MarshalRivine(buf)
		if err != nil {
			t.Error(idx, err)
			continue
		}

		out := hex.EncodeToString(buf.Bytes())
		if out != testCase {
			t.Error(idx, out, "!=", testCase)
		}
	}
}

func TestUnknownUnlockFulfillmentSiaEncoding(t *testing.T) {
	testCases := []string{
		// unknown fulfillment
		`ff0c0000000000000048656c6c6f2c205465737421`,
		// time lock fulfillment
		`030d00000000000000ff48656c6c6f2c205465737421`, // using unknown fulfillment
	}
	for idx, testCase := range testCases {
		b, err := hex.DecodeString(testCase)
		if err != nil {
			t.Error(idx, err)
			continue
		}

		var uf UnlockFulfillmentProxy
		err = uf.UnmarshalSia(bytes.NewReader(b))
		if err == nil {
			t.Error(idx, "error expected while non received")
		}
	}
}

func TestUnlockConditionJSONEncoding(t *testing.T) {
	testCases := []struct {
		Input  string
		Output string
	}{
		// nil condition
		{`{}`, ``},
		{`{"type":0}`, `{}`},
		{`{"type":0,"data":null}`, `{}`},
		// unlock hash condition
		{`{
	"type":1,
	"data":{ 
		"unlockhash":"01a6a6c5584b2bfbd08738996cd7930831f958b9a5ed1595525236e861c1a0dc353bdcf54be7d8"
	}
}`, ``},
		{`{
	"type":1,
	"data": {
		"unlockhash":"6453402d094ed0f336950c4be0feec37167aaaaf8b974d265900e49ab22773584cfe96393b1360"
	}
}`, ``},
		{`{
	"type": 1,
	"data": {
		"unlockhash": "0101234567890123456789012345678901012345678901234567890123456789018a50e31447b8"
	}
}`, ``},
		// atomic swap condition
		{`{
	"type": 2,
	"data": {
		"sender": "6453402d094ed0f336950c4be0feec37167aaaaf8b974d265900e49ab22773584cfe96393b1360",
		"receiver": "0101234567890123456789012345678901012345678901234567890123456789018a50e31447b8",
		"hashedsecret": "abc543defabc543defabc543defabc543defabc543defabc543defabc543defa",
		"timelock": 1522068743
	}
}`, ``},
		// time lock condition
		{`{
	"type": 3,
	"data": {
		"locktime": 500000000,
		"condition": {}
	}
}`, ``}, // using nil condition
		{`{
	"type": 3,
	"data": {
		"locktime": 500000000,
		"condition": {
			"type": 0,
			"data": {}
		}
	}
}`, `{
	"type": 3,
	"data": {
		"locktime": 500000000,
		"condition": {}
	}
}`}, // using nil condition
		{`{
	"type": 3,
	"data": {
		"locktime": 500000000,
		"condition": {
			"type": 1,
			"data": {
				"unlockhash": "0101234567890123456789012345678901012345678901234567890123456789018a50e31447b8"
			}
		}
	}
}`, ``}, // using unlock hash condition
		// MultiSig condition
		{
			`{
			"type": 4,
			"data": {
				"unlockhashes": [
					"01e89843e4b8231a01ba18b254d530110364432aafab8206bea72e5a20eaa55f70b1ccc65e2105",
					"01a6a6c5584b2bfbd08738996cd7930831f958b9a5ed1595525236e861c1a0dc353bdcf54be7d8"
				],
				"minimumsignaturecount": 2
			}
		}`,
			``,
		},
	}
	for idx, testCase := range testCases {
		var up UnlockConditionProxy
		err := json.Unmarshal([]byte(testCase.Input), &up)
		if err != nil {
			t.Error(idx, err)
			continue
		}

		b, err := json.Marshal(up)
		if err != nil {
			t.Error(idx, err)
			continue
		}

		expected := testCase.Output
		if expected == "" {
			expected = testCase.Input
		}
		expected = strings.Replace(strings.Replace(strings.Replace(
			expected, " ", "", -1), "\t", "", -1), "\n", "", -1)
		out := string(b)
		if out != expected {
			t.Error(idx, out, "!=", expected)
		}
	}
}

func TestUnlockFulfillmentJSONEncoding(t *testing.T) {
	testCases := []struct {
		Input  string
		Output string
	}{
		// nil fulfillment
		{`{}`, ``},
		{`{"type":0}`, `{}`},
		{`{"type":0,"data":null}`, `{}`},
		// single signature fulfillment
		{`{
	"type": 1,
	"data": {
		"publickey": "ed25519:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
		"signature": "abcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefab"
	}
}`, ``},
		// legacy atomic swap fulfillment
		{
			`{
	"type": 2,
	"data": {
		"sender": "6453402d094ed0f336950c4be0feec37167aaaaf8b974d265900e49ab22773584cfe96393b1360",
		"receiver": "0101234567890123456789012345678901012345678901234567890123456789018a50e31447b8",
		"hashedsecret": "abc543defabc543defabc543defabc543defabc543defabc543defabc543defa",
		"timelock": 1522068743,
		"publickey": "ed25519:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
		"signature": "abcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefab"
	}
}`, `{
	"type": 2,
	"data": {
		"sender": "6453402d094ed0f336950c4be0feec37167aaaaf8b974d265900e49ab22773584cfe96393b1360",
		"receiver": "0101234567890123456789012345678901012345678901234567890123456789018a50e31447b8",
		"hashedsecret": "abc543defabc543defabc543defabc543defabc543defabc543defabc543defa",
		"timelock": 1522068743,
		"publickey": "ed25519:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
		"signature": "abcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefab",
		"secret": "0000000000000000000000000000000000000000000000000000000000000000"
	}
}`},
		{
			`{
	"type": 2,
	"data": {
		"sender": "6453402d094ed0f336950c4be0feec37167aaaaf8b974d265900e49ab22773584cfe96393b1360",
		"receiver": "0101234567890123456789012345678901012345678901234567890123456789018a50e31447b8",
		"hashedsecret": "abc543defabc543defabc543defabc543defabc543defabc543defabc543defa",
		"timelock": 1522068743,
		"publickey": "ed25519:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
		"signature": "abcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefab",
		"secret": "def789def789def789def789def789dedef789def789def789def789def789de"
	}
}`, ``},
		// atomic swap fulfillment
		{
			`{
	"type": 2,
	"data": {
		"publickey": "ed25519:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
		"signature": "abcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefab"
	}
}`, `{
	"type": 2,
	"data": {
		"publickey": "ed25519:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
		"signature": "abcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefab",
		"secret": "0000000000000000000000000000000000000000000000000000000000000000"
	}
}`},
		{
			`{
	"type": 2,
	"data": {
		"publickey": "ed25519:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
		"signature": "abcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefab",
		"secret": "def789def789def789def789def789dedef789def789def789def789def789de"
	}
}`, ``},
		{
			`{
			"type": 3,
			"data": {
				"pairs": [
					{
						"publickey": "ed25519:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
						"signature": "abcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefab"
					},
					{
						"publickey": "ed25519:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
						"signature": "abcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefab"
					}
				]
			}
		}`,
			``,
		},
	}
	for idx, testCase := range testCases {
		var fp UnlockFulfillmentProxy
		err := json.Unmarshal([]byte(testCase.Input), &fp)
		if err != nil {
			t.Error(idx, err)
			continue
		}

		b, err := json.Marshal(fp)
		if err != nil {
			t.Error(idx, err)
			continue
		}

		expected := testCase.Output
		if expected == "" {
			expected = testCase.Input
		}
		expected = strings.Replace(strings.Replace(strings.Replace(
			expected, " ", "", -1), "\t", "", -1), "\n", "", -1)
		out := string(b)
		if out != expected {
			t.Error(idx, out, "!=", expected)
		}
	}
}

func TestNilUnlockConditionProxy(t *testing.T) {
	var c UnlockConditionProxy
	if ct := c.ConditionType(); ct != ConditionTypeNil {
		t.Error("ConditionType", ct, "!=", ConditionTypeNil)
	}
	if err := c.IsStandardCondition(ValidationContext{}); err != nil {
		t.Error("IsStandardCondition", err)
	}
	if b, err := c.MarshalJSON(); err != nil || string(b) != "{}" {
		t.Error("MarshalJSON", b, err)
	}
	if b := siabin.Marshal(c); bytes.Compare(b, []byte{0, 0, 0, 0, 0, 0, 0, 0, 0}) != 0 {
		t.Error("MarshalSia", b)
	}
	if b := rivbin.Marshal(c); bytes.Compare(b, []byte{0, 0, 0, 0, 0, 0, 0, 0, 0}) != 0 {
		t.Error("MarshalRivine", b)
	}
	if !c.Equal(nil) {
		t.Error("should equal to nil implicitly")
	}
	if !c.Equal(&NilCondition{}) {
		t.Error("should equal to nil explicitly")
	}
	if !c.Equal(NewCondition(nil)) {
		t.Error("should equal to nil explicitly")
	}
}

func TestNilUnlockFulfillmentProxy(t *testing.T) {
	var f UnlockFulfillmentProxy
	if ft := f.FulfillmentType(); ft != FulfillmentTypeNil {
		t.Error("FulfillmentType", ft, "!=", FulfillmentTypeNil)
	}
	if err := f.IsStandardFulfillment(ValidationContext{}); err == nil {
		t.Error("IsStandardFulfillment should not be standard")
	}
	if b, err := f.MarshalJSON(); err != nil || string(b) != "{}" {
		t.Error("MarshalJSON", b, err)
	}
	if b := siabin.Marshal(f); bytes.Compare(b, []byte{0, 0, 0, 0, 0, 0, 0, 0, 0}) != 0 {
		t.Error("MarshalSia", b)
	}
	if b := rivbin.Marshal(f); bytes.Compare(b, []byte{0, 0, 0, 0, 0, 0, 0, 0, 0}) != 0 {
		t.Error("MarshalRivine", b)
	}
	if !f.Equal(nil) {
		t.Error("should equal to nil implicitly")
	}
	if !f.Equal(&NilFulfillment{}) {
		t.Error("should equal to nil explicitly")
	}
	if !f.Equal(NewFulfillment(nil)) {
		t.Error("should equal to nil explicitly")
	}
}

func TestUnlockConditionEqual(t *testing.T) {
	testCases := []struct {
		A, B            UnlockCondition
		NotEqualMessage string
	}{
		{&NilCondition{}, nil, ""},             // implicit
		{&NilCondition{}, &NilCondition{}, ""}, // explicit
		{&NilCondition{}, &UnlockHashCondition{}, "unequal type"},
		{&UnlockHashCondition{}, nil, "unequal type"},
		{&UnlockHashCondition{}, &NilCondition{}, "unequal type"},
		{&UnlockHashCondition{}, &AtomicSwapCondition{}, "unequal type"},
		{&UnlockHashCondition{}, &UnlockHashCondition{}, ""},
		{
			&UnlockHashCondition{TargetUnlockHash: UnknownUnlockHash},
			&UnlockHashCondition{TargetUnlockHash: NilUnlockHash},
			"unequal crypto hash",
		},
		{
			&UnlockHashCondition{TargetUnlockHash: NewUnlockHash(UnlockTypePubKey, crypto.Hash{})},
			&UnlockHashCondition{TargetUnlockHash: NilUnlockHash},
			"unequal unlock hash type",
		},
		{
			&UnlockHashCondition{TargetUnlockHash: NewUnlockHash(UnlockTypePubKey, crypto.Hash{})},
			&UnlockHashCondition{TargetUnlockHash: NewUnlockHash(UnlockTypePubKey, crypto.Hash{})},
			"",
		},
		{
			&UnlockHashCondition{TargetUnlockHash: unlockHashFromHex("01746677df456546d93729066dd88514e2009930f3eebac3c93d43c88a108f8f9aa9e7c6f58893")},
			&UnlockHashCondition{TargetUnlockHash: unlockHashFromHex("01746677df456546d93729066dd88514e2009930f3eebac3c93d43c88a108f8f9aa9e7c6f58893")},
			"",
		},
		{
			&UnlockHashCondition{TargetUnlockHash: unlockHashFromHex("01746677df456546d93729066dd88514e2009930f3eebac3c93d43c88a108f8f9aa9e7c6f58893")},
			&UnlockHashCondition{TargetUnlockHash: unlockHashFromHex("02a24c97c80eeac111aa4bcbb0ac8ffc364fa9b22da10d3054778d2332f68b365e5e5af8e71541")},
			"unequal unlock hash type",
		},
		{
			&UnlockHashCondition{TargetUnlockHash: unlockHashFromHex("015fe50b9c596d8717e5e7ba79d5a7c9c8b82b1427a04d5c0771268197c90e99dccbcdf0ba9c90")},
			&UnlockHashCondition{TargetUnlockHash: unlockHashFromHex("01746677df456546d93729066dd88514e2009930f3eebac3c93d43c88a108f8f9aa9e7c6f58893")},
			"unequal crypto hash",
		},
		{&AtomicSwapCondition{}, nil, "unequal type"},
		{&AtomicSwapCondition{}, &NilCondition{}, "unequal type"},
		{&AtomicSwapCondition{}, &UnlockHashCondition{}, "unequal type"},
		{&AtomicSwapCondition{}, &AtomicSwapCondition{}, ""},
		{
			&AtomicSwapCondition{
				Sender: unlockHashFromHex("01746677df456546d93729066dd88514e2009930f3eebac3c93d43c88a108f8f9aa9e7c6f58893"),
			},
			&AtomicSwapCondition{}, "unequal atomic swap conditions",
		},
		{
			&AtomicSwapCondition{
				Sender:       unlockHashFromHex("01746677df456546d93729066dd88514e2009930f3eebac3c93d43c88a108f8f9aa9e7c6f58893"),
				Receiver:     unlockHashFromHex("015fe50b9c596d8717e5e7ba79d5a7c9c8b82b1427a04d5c0771268197c90e99dccbcdf0ba9c90"),
				HashedSecret: AtomicSwapHashedSecret{1, 2, 3},
				TimeLock:     TestnetChainConstants().GenesisTimestamp,
			},
			&AtomicSwapCondition{
				Sender:       unlockHashFromHex("01746677df456546d93729066dd88514e2009930f3eebac3c93d43c88a108f8f9aa9e7c6f58893"),
				Receiver:     unlockHashFromHex("015fe50b9c596d8717e5e7ba79d5a7c9c8b82b1427a04d5c0771268197c90e99dccbcdf0ba9c90"),
				HashedSecret: AtomicSwapHashedSecret{1, 2, 3},
				TimeLock:     TestnetChainConstants().GenesisTimestamp,
			},
			"",
		},
		{
			&AtomicSwapCondition{
				Sender:       unlockHashFromHex("01746677df456546d93729066dd88514e2009930f3eebac3c93d43c88a108f8f9aa9e7c6f58893"),
				Receiver:     unlockHashFromHex("015fe50b9c596d8717e5e7ba79d5a7c9c8b82b1427a04d5c0771268197c90e99dccbcdf0ba9c90"),
				HashedSecret: AtomicSwapHashedSecret{1, 2, 4},
				TimeLock:     TestnetChainConstants().GenesisTimestamp,
			},
			&AtomicSwapCondition{
				Sender:       unlockHashFromHex("01746677df456546d93729066dd88514e2009930f3eebac3c93d43c88a108f8f9aa9e7c6f58893"),
				Receiver:     unlockHashFromHex("015fe50b9c596d8717e5e7ba79d5a7c9c8b82b1427a04d5c0771268197c90e99dccbcdf0ba9c90"),
				HashedSecret: AtomicSwapHashedSecret{1, 2, 3},
				TimeLock:     TestnetChainConstants().GenesisTimestamp,
			},
			"unequal hashed secret",
		},
		{
			&AtomicSwapCondition{
				Sender:       unlockHashFromHex("01746677df456546d93729066dd88514e2009930f3eebac3c93d43c88a108f8f9aa9e7c6f58893"),
				Receiver:     unlockHashFromHex("015fe50b9c596d8717e5e7ba79d5a7c9c8b82b1427a04d5c0771268197c90e99dccbcdf0ba9c90"),
				HashedSecret: AtomicSwapHashedSecret{1, 2, 3},
				TimeLock:     TestnetChainConstants().GenesisTimestamp,
			},
			&AtomicSwapCondition{
				Sender:       unlockHashFromHex("01746677df456546d93729066dd88514e2009930f3eebac3c93d43c88a108f8f9aa9e7c6f58893"),
				Receiver:     unlockHashFromHex("01746677df456546d93729066dd88514e2009930f3eebac3c93d43c88a108f8f9aa9e7c6f58893"),
				HashedSecret: AtomicSwapHashedSecret{1, 2, 3},
				TimeLock:     TestnetChainConstants().GenesisTimestamp,
			},
			"unequal receiver",
		},
		{
			&AtomicSwapCondition{
				Sender:       unlockHashFromHex("01746677df456546d93729066dd88514e2009930f3eebac3c93d43c88a108f8f9aa9e7c6f58893"),
				Receiver:     unlockHashFromHex("015fe50b9c596d8717e5e7ba79d5a7c9c8b82b1427a04d5c0771268197c90e99dccbcdf0ba9c90"),
				HashedSecret: AtomicSwapHashedSecret{1, 2, 3},
				TimeLock:     0,
			},
			&AtomicSwapCondition{
				Sender:       unlockHashFromHex("01746677df456546d93729066dd88514e2009930f3eebac3c93d43c88a108f8f9aa9e7c6f58893"),
				Receiver:     unlockHashFromHex("015fe50b9c596d8717e5e7ba79d5a7c9c8b82b1427a04d5c0771268197c90e99dccbcdf0ba9c90"),
				HashedSecret: AtomicSwapHashedSecret{1, 2, 3},
				TimeLock:     TestnetChainConstants().GenesisTimestamp,
			},
			"unequal time lock",
		},
		{
			&AtomicSwapCondition{
				Sender:       unlockHashFromHex("01746677df456546d93729066dd88514e2009930f3eebac3c93d43c88a108f8f9aa9e7c6f58893"),
				Receiver:     unlockHashFromHex("015fe50b9c596d8717e5e7ba79d5a7c9c8b82b1427a04d5c0771268197c90e99dccbcdf0ba9c90"),
				HashedSecret: AtomicSwapHashedSecret{1, 2, 3},
				TimeLock:     TestnetChainConstants().GenesisTimestamp,
			},
			&AtomicSwapCondition{
				Sender:       unlockHashFromHex("015fe50b9c596d8717e5e7ba79d5a7c9c8b82b1427a04d5c0771268197c90e99dccbcdf0ba9c90"),
				Receiver:     unlockHashFromHex("015fe50b9c596d8717e5e7ba79d5a7c9c8b82b1427a04d5c0771268197c90e99dccbcdf0ba9c90"),
				HashedSecret: AtomicSwapHashedSecret{1, 2, 3},
				TimeLock:     TestnetChainConstants().GenesisTimestamp,
			},
			"unequal sender",
		},
		{
			&TimeLockCondition{Condition: &NilCondition{}},
			nil, "unequal type",
		},
		{
			&TimeLockCondition{Condition: &NilCondition{}},
			&NilCondition{}, "unequal type",
		},
		{
			&TimeLockCondition{Condition: &NilCondition{}},
			&UnlockHashCondition{}, "unequal type",
		},
		{
			&TimeLockCondition{Condition: &NilCondition{}},
			&AtomicSwapCondition{}, "unequal type",
		},
		{
			&TimeLockCondition{Condition: &NilCondition{}},
			&TimeLockCondition{}, "",
		},
		{
			&TimeLockCondition{Condition: &NilCondition{}, LockTime: 42},
			&TimeLockCondition{Condition: &NilCondition{}},
			"non-equal lock time",
		},
		{
			&TimeLockCondition{Condition: &NilCondition{}},
			&TimeLockCondition{Condition: &NilCondition{}, LockTime: 42},
			"non-equal lock time",
		},
		{
			&TimeLockCondition{
				Condition: &UnlockHashCondition{
					TargetUnlockHash: unlockHashFromHex("01746677df456546d93729066dd88514e2009930f3eebac3c93d43c88a108f8f9aa9e7c6f58893"),
				},
			},
			&TimeLockCondition{
				Condition: &UnlockHashCondition{
					TargetUnlockHash: unlockHashFromHex("01746677df456546d93729066dd88514e2009930f3eebac3c93d43c88a108f8f9aa9e7c6f58893"),
				},
			},
			"",
		},
		{
			&TimeLockCondition{
				Condition: &UnlockHashCondition{
					TargetUnlockHash: unlockHashFromHex("01746677df456546d93729066dd88514e2009930f3eebac3c93d43c88a108f8f9aa9e7c6f58893"),
				},
				LockTime: 5000,
			},
			&TimeLockCondition{
				Condition: &UnlockHashCondition{
					TargetUnlockHash: unlockHashFromHex("01746677df456546d93729066dd88514e2009930f3eebac3c93d43c88a108f8f9aa9e7c6f58893"),
				},
				LockTime: 5000,
			},
			"",
		},
		{
			&TimeLockCondition{
				Condition: &MultiSignatureCondition{
					UnlockHashes: UnlockHashSlice{
						unlockHashFromHex("01746677df456546d93729066dd88514e2009930f3eebac3c93d43c88a108f8f9aa9e7c6f58893"),
						unlockHashFromHex("01fe92204c6e413c765a39988605a80fe4273eda49b95ff7b6e57d8049afee13574ef6a04985f1"),
					},
					MinimumSignatureCount: 2,
				},
			},
			&TimeLockCondition{
				Condition: &MultiSignatureCondition{
					UnlockHashes: UnlockHashSlice{
						unlockHashFromHex("01746677df456546d93729066dd88514e2009930f3eebac3c93d43c88a108f8f9aa9e7c6f58893"),
						unlockHashFromHex("01fe92204c6e413c765a39988605a80fe4273eda49b95ff7b6e57d8049afee13574ef6a04985f1"),
					},
					MinimumSignatureCount: 2,
				},
			},
			"",
		},
		{
			&MultiSignatureCondition{},
			&MultiSignatureCondition{},
			"",
		},
		{
			&MultiSignatureCondition{},
			&NilCondition{},
			"unequal type",
		},
		{
			&MultiSignatureCondition{},
			nil,
			"unequal type",
		},
		{
			&MultiSignatureCondition{
				MinimumSignatureCount: 2,
				UnlockHashes: UnlockHashSlice{
					unlockHashFromHex("01e89843e4b8231a01ba18b254d530110364432aafab8206bea72e5a20eaa55f70b1ccc65e2105"),
					unlockHashFromHex("01a6a6c5584b2bfbd08738996cd7930831f958b9a5ed1595525236e861c1a0dc353bdcf54be7d8"),
				},
			},
			&MultiSignatureCondition{
				MinimumSignatureCount: 2,
				UnlockHashes: UnlockHashSlice{
					unlockHashFromHex("01e89843e4b8231a01ba18b254d530110364432aafab8206bea72e5a20eaa55f70b1ccc65e2105"),
					unlockHashFromHex("01a6a6c5584b2bfbd08738996cd7930831f958b9a5ed1595525236e861c1a0dc353bdcf54be7d8"),
				},
			},
			"",
		},
		{
			&MultiSignatureCondition{
				MinimumSignatureCount: 2,
				UnlockHashes: UnlockHashSlice{
					unlockHashFromHex("01e89843e4b8231a01ba18b254d530110364432aafab8206bea72e5a20eaa55f70b1ccc65e2105"),
					unlockHashFromHex("01a6a6c5584b2bfbd08738996cd7930831f958b9a5ed1595525236e861c1a0dc353bdcf54be7d8"),
				},
			},
			// swap the unlock hash placement
			&MultiSignatureCondition{
				MinimumSignatureCount: 2,
				UnlockHashes: UnlockHashSlice{
					unlockHashFromHex("01a6a6c5584b2bfbd08738996cd7930831f958b9a5ed1595525236e861c1a0dc353bdcf54be7d8"),
					unlockHashFromHex("01e89843e4b8231a01ba18b254d530110364432aafab8206bea72e5a20eaa55f70b1ccc65e2105"),
				},
			},
			"",
		},
		{
			&MultiSignatureCondition{
				MinimumSignatureCount: 2,
				UnlockHashes: UnlockHashSlice{
					unlockHashFromHex("01e89843e4b8231a01ba18b254d530110364432aafab8206bea72e5a20eaa55f70b1ccc65e2105"),
					unlockHashFromHex("01a6a6c5584b2bfbd08738996cd7930831f958b9a5ed1595525236e861c1a0dc353bdcf54be7d8"),
				},
			},
			&MultiSignatureCondition{
				MinimumSignatureCount: 1,
				UnlockHashes: UnlockHashSlice{
					unlockHashFromHex("01e89843e4b8231a01ba18b254d530110364432aafab8206bea72e5a20eaa55f70b1ccc65e2105"),
					unlockHashFromHex("01a6a6c5584b2bfbd08738996cd7930831f958b9a5ed1595525236e861c1a0dc353bdcf54be7d8"),
				},
			},
			"Different amount of signatures required",
		},
		{
			&MultiSignatureCondition{
				MinimumSignatureCount: 2,
				UnlockHashes: UnlockHashSlice{
					unlockHashFromHex("01e89843e4b8231a01ba18b254d530110364432aafab8206bea72e5a20eaa55f70b1ccc65e2105"),
					unlockHashFromHex("01a6a6c5584b2bfbd08738996cd7930831f958b9a5ed1595525236e861c1a0dc353bdcf54be7d8"),
					unlockHashFromHex("01a6a6c5584b2bfbd08738996cd7930831f958b9a5ed1595525236e861c1a0dc353bdcf54be7d8"),
				},
			},
			&MultiSignatureCondition{
				MinimumSignatureCount: 2,
				UnlockHashes: UnlockHashSlice{
					unlockHashFromHex("01e89843e4b8231a01ba18b254d530110364432aafab8206bea72e5a20eaa55f70b1ccc65e2105"),
					unlockHashFromHex("01a6a6c5584b2bfbd08738996cd7930831f958b9a5ed1595525236e861c1a0dc353bdcf54be7d8"),
				},
			},
			"Unlock hash occurst multiple times",
		},
		{
			&MultiSignatureCondition{
				MinimumSignatureCount: 2,
				UnlockHashes: UnlockHashSlice{
					unlockHashFromHex("01e89843e4b8231a01ba18b254d530110364432aafab8206bea72e5a20eaa55f70b1ccc65e2105"),
					unlockHashFromHex("01a6a6c5584b2bfbd08738996cd7930831f958b9a5ed1595525236e861c1a0dc353bdcf54be7d8"),
					unlockHashFromHex("01a6a6c5584b2bfbd08738996cd7930831f958b9a5ed1595525236e861c1a0dc353bdcf54be7d8"),
				},
			},
			&MultiSignatureCondition{
				MinimumSignatureCount: 2,
				UnlockHashes: UnlockHashSlice{
					unlockHashFromHex("01e89843e4b8231a01ba18b254d530110364432aafab8206bea72e5a20eaa55f70b1ccc65e2105"),
					unlockHashFromHex("01a6a6c5584b2bfbd08738996cd7930831f958b9a5ed1595525236e861c1a0dc353bdcf54be7d8"),
					unlockHashFromHex("01a6a6c5584b2bfbd08738996cd7930831f958b9a5ed1595525236e861c1a0dc353bdcf54be7d8")},
			},
			"",
		},
		{
			&MultiSignatureCondition{
				MinimumSignatureCount: 2,
				UnlockHashes: UnlockHashSlice{
					unlockHashFromHex("01e89843e4b8231a01ba18b254d530110364432aafab8206bea72e5a20eaa55f70b1ccc65e2105"),
				},
			},
			&MultiSignatureCondition{
				MinimumSignatureCount: 2,
				UnlockHashes: UnlockHashSlice{
					unlockHashFromHex("01e89843e4b8231a01ba18b254d530110364432aafab8206bea72e5a20eaa55f70b1ccc65e2105"),
					unlockHashFromHex("01e89843e4b8231a01ba18b254d530110364432aafab8206bea72e5a20eaa55f70b1ccc65e2105"),
				},
			},
			"edge case, where second condition has two times the unlockhashes the first condition has",
		},
		{
			&MultiSignatureCondition{
				MinimumSignatureCount: 2,
				UnlockHashes: UnlockHashSlice{
					unlockHashFromHex("01e89843e4b8231a01ba18b254d530110364432aafab8206bea72e5a20eaa55f70b1ccc65e2105"),
					unlockHashFromHex("01a6a6c5584b2bfbd08738996cd7930831f958b9a5ed1595525236e861c1a0dc353bdcf54be7d8"),
				},
			},
			&MultiSignatureCondition{
				MinimumSignatureCount: 2,
				UnlockHashes: UnlockHashSlice{
					unlockHashFromHex("01e89843e4b8231a01ba18b254d530110364432aafab8206bea72e5a20eaa55f70b1ccc65e2105"),
					unlockHashFromHex("01e89843e4b8231a01ba18b254d530110364432aafab8206bea72e5a20eaa55f70b1ccc65e2105"),
				},
			},
			"edge case, where second condition has two times the unlockhashes the first condition has, but length is equal",
		},
		{
			&MultiSignatureCondition{
				MinimumSignatureCount: 2,
				UnlockHashes: UnlockHashSlice{
					unlockHashFromHex("01e89843e4b8231a01ba18b254d530110364432aafab8206bea72e5a20eaa55f70b1ccc65e2105"),
					unlockHashFromHex("01e89843e4b8231a01ba18b254d530110364432aafab8206bea72e5a20eaa55f70b1ccc65e2105"),
				},
			},
			&MultiSignatureCondition{
				MinimumSignatureCount: 2,
				UnlockHashes: UnlockHashSlice{
					unlockHashFromHex("01e89843e4b8231a01ba18b254d530110364432aafab8206bea72e5a20eaa55f70b1ccc65e2105"),
					unlockHashFromHex("01a6a6c5584b2bfbd08738996cd7930831f958b9a5ed1595525236e861c1a0dc353bdcf54be7d8"),
				},
			},
			"edge case, where first condition has two times the unlockhashes the second condition has, but length is equal",
		},
	}
	for idx, testCase := range testCases {
		equal := testCase.A.Equal(testCase.B)
		if testCase.NotEqualMessage != "" {
			if equal {
				t.Error(idx, "expected not equal, but it's equal:", testCase.NotEqualMessage, testCase.A, testCase.B)
			}
		} else {
			if !equal {
				t.Error(idx, "expected equal, but it's not equal", testCase.A, testCase.B)
			}
		}
	}
}

func TestUnlockFulfillmentEqual(t *testing.T) {
	testCases := []struct {
		A, B            UnlockFulfillment
		NotEqualMessage string
	}{
		{&NilFulfillment{}, nil, ""},
		{&NilFulfillment{}, &NilFulfillment{}, ""},
		{&NilFulfillment{}, &SingleSignatureFulfillment{}, "unequal type"},
		{&SingleSignatureFulfillment{}, nil, "unequal type"},
		{&SingleSignatureFulfillment{}, &NilFulfillment{}, "unequal type"},
		{&SingleSignatureFulfillment{}, &AtomicSwapFulfillment{}, "unequal type"},
		{&SingleSignatureFulfillment{}, &LegacyAtomicSwapFulfillment{}, "unequal type"},
		{&SingleSignatureFulfillment{}, &SingleSignatureFulfillment{}, ""},
		{
			&SingleSignatureFulfillment{
				PublicKey: PublicKey{
					Algorithm: SignatureAlgoEd25519,
				},
			},
			&SingleSignatureFulfillment{}, "different pub-key algorithm",
		},
		{
			&SingleSignatureFulfillment{
				PublicKey: PublicKey{
					Algorithm: SignatureAlgoEd25519,
					Key:       ByteSlice{1, 2, 3},
				},
			},
			&SingleSignatureFulfillment{
				PublicKey: PublicKey{
					Algorithm: SignatureAlgoEd25519,
				},
			},
			"different pub-key key",
		},
		{
			&SingleSignatureFulfillment{
				PublicKey: PublicKey{
					Algorithm: SignatureAlgoEd25519,
					Key:       ByteSlice{1, 2, 3},
				},
				Signature: ByteSlice{4, 2},
			},
			&SingleSignatureFulfillment{
				PublicKey: PublicKey{
					Algorithm: SignatureAlgoEd25519,
					Key:       ByteSlice{1, 2, 3},
				},
			},
			"different signature",
		},
		{
			&SingleSignatureFulfillment{
				PublicKey: PublicKey{
					Algorithm: SignatureAlgoEd25519,
					Key:       ByteSlice{1, 2, 3},
				},
				Signature: ByteSlice{4, 2},
			},
			&SingleSignatureFulfillment{
				PublicKey: PublicKey{
					Algorithm: SignatureAlgoEd25519,
					Key:       ByteSlice{1, 2, 3},
				},
				Signature: ByteSlice{4, 2},
			},
			"",
		},
		{&AtomicSwapFulfillment{}, nil, "unequal type"},
		{&AtomicSwapFulfillment{}, &NilFulfillment{}, "unequal type"},
		{&AtomicSwapFulfillment{}, &LegacyAtomicSwapFulfillment{}, "unequal type"},
		{&AtomicSwapFulfillment{}, &SingleSignatureFulfillment{}, "unequal type"},
		{&AtomicSwapFulfillment{}, &AtomicSwapFulfillment{}, ""},
		{
			&AtomicSwapFulfillment{
				PublicKey: PublicKey{
					Algorithm: SignatureAlgoEd25519,
				},
			},
			&AtomicSwapFulfillment{},
			"different pub-key algo",
		},
		{
			&AtomicSwapFulfillment{
				PublicKey: PublicKey{
					Algorithm: SignatureAlgoEd25519,
					Key:       ByteSlice{1, 2, 3},
				},
			},
			&AtomicSwapFulfillment{
				PublicKey: PublicKey{
					Algorithm: SignatureAlgoEd25519,
				},
			},
			"different pub-key key",
		},
		{
			&AtomicSwapFulfillment{
				PublicKey: PublicKey{
					Algorithm: SignatureAlgoEd25519,
					Key:       ByteSlice{1, 2, 3},
				},
				Signature: ByteSlice{4, 5, 6},
			},
			&AtomicSwapFulfillment{
				PublicKey: PublicKey{
					Algorithm: SignatureAlgoEd25519,
					Key:       ByteSlice{1, 2, 3},
				},
			},
			"different signature",
		},
		{
			&AtomicSwapFulfillment{
				PublicKey: PublicKey{
					Algorithm: SignatureAlgoEd25519,
					Key:       ByteSlice{1, 2, 3},
				},
				Signature: ByteSlice{4, 5, 6},
				Secret:    AtomicSwapSecret{7, 8, 9},
			},
			&AtomicSwapFulfillment{
				PublicKey: PublicKey{
					Algorithm: SignatureAlgoEd25519,
					Key:       ByteSlice{1, 2, 3},
				},
				Signature: ByteSlice{4, 5, 6},
			},
			"different secret",
		},
		{
			&AtomicSwapFulfillment{
				PublicKey: PublicKey{
					Algorithm: SignatureAlgoEd25519,
					Key:       ByteSlice{1, 2, 3},
				},
				Signature: ByteSlice{4, 5, 6},
			},
			&AtomicSwapFulfillment{
				PublicKey: PublicKey{
					Algorithm: SignatureAlgoEd25519,
					Key:       ByteSlice{1, 2, 3},
				},
				Signature: ByteSlice{4, 5, 6},
			},
			"",
		},
		{
			&AtomicSwapFulfillment{
				PublicKey: PublicKey{
					Algorithm: SignatureAlgoEd25519,
					Key:       ByteSlice{1, 2, 3},
				},
				Signature: ByteSlice{4, 5, 6},
				Secret:    AtomicSwapSecret{7, 8, 9},
			},
			&AtomicSwapFulfillment{
				PublicKey: PublicKey{
					Algorithm: SignatureAlgoEd25519,
					Key:       ByteSlice{1, 2, 3},
				},
				Signature: ByteSlice{4, 5, 6},
				Secret:    AtomicSwapSecret{7, 8, 9},
			},
			"",
		},
		{&LegacyAtomicSwapFulfillment{}, nil, "unequal type"},
		{&LegacyAtomicSwapFulfillment{}, &NilFulfillment{}, "unequal type"},
		{&LegacyAtomicSwapFulfillment{}, &SingleSignatureFulfillment{}, "unequal type"},
		{&LegacyAtomicSwapFulfillment{}, &AtomicSwapFulfillment{}, "unequal type"},
		{&LegacyAtomicSwapFulfillment{}, &LegacyAtomicSwapFulfillment{}, ""},
		{
			&LegacyAtomicSwapFulfillment{
				PublicKey: PublicKey{
					Algorithm: SignatureAlgoEd25519,
				},
			},
			&LegacyAtomicSwapFulfillment{},
			"different pub-key algo",
		},
		{
			&LegacyAtomicSwapFulfillment{
				PublicKey: PublicKey{
					Algorithm: SignatureAlgoEd25519,
					Key:       ByteSlice{1, 2, 3},
				},
			},
			&LegacyAtomicSwapFulfillment{
				PublicKey: PublicKey{
					Algorithm: SignatureAlgoEd25519,
				},
			},
			"different pub-key key",
		},
		{
			&LegacyAtomicSwapFulfillment{
				PublicKey: PublicKey{
					Algorithm: SignatureAlgoEd25519,
					Key:       ByteSlice{1, 2, 3},
				},
				Signature: ByteSlice{4, 5, 6},
			},
			&LegacyAtomicSwapFulfillment{
				PublicKey: PublicKey{
					Algorithm: SignatureAlgoEd25519,
					Key:       ByteSlice{1, 2, 3},
				},
			},
			"different signature",
		},
		{
			&LegacyAtomicSwapFulfillment{
				PublicKey: PublicKey{
					Algorithm: SignatureAlgoEd25519,
					Key:       ByteSlice{1, 2, 3},
				},
				Signature: ByteSlice{4, 5, 6},
				Secret:    AtomicSwapSecret{7, 8, 9},
			},
			&LegacyAtomicSwapFulfillment{
				PublicKey: PublicKey{
					Algorithm: SignatureAlgoEd25519,
					Key:       ByteSlice{1, 2, 3},
				},
				Signature: ByteSlice{4, 5, 6},
			},
			"different secret",
		},
		{
			&LegacyAtomicSwapFulfillment{
				PublicKey: PublicKey{
					Algorithm: SignatureAlgoEd25519,
					Key:       ByteSlice{1, 2, 3},
				},
				Signature: ByteSlice{4, 5, 6},
			},
			&LegacyAtomicSwapFulfillment{
				PublicKey: PublicKey{
					Algorithm: SignatureAlgoEd25519,
					Key:       ByteSlice{1, 2, 3},
				},
				Signature: ByteSlice{4, 5, 6},
			},
			"",
		},
		{
			&LegacyAtomicSwapFulfillment{
				PublicKey: PublicKey{
					Algorithm: SignatureAlgoEd25519,
					Key:       ByteSlice{1, 2, 3},
				},
				Signature: ByteSlice{4, 5, 6},
				Secret:    AtomicSwapSecret{7, 8, 9},
			},
			&LegacyAtomicSwapFulfillment{
				PublicKey: PublicKey{
					Algorithm: SignatureAlgoEd25519,
					Key:       ByteSlice{1, 2, 3},
				},
				Signature: ByteSlice{4, 5, 6},
				Secret:    AtomicSwapSecret{7, 8, 9},
			},
			"",
		},
		{
			&LegacyAtomicSwapFulfillment{
				Sender:       unlockHashFromHex("01746677df456546d93729066dd88514e2009930f3eebac3c93d43c88a108f8f9aa9e7c6f58893"),
				Receiver:     unlockHashFromHex("015fe50b9c596d8717e5e7ba79d5a7c9c8b82b1427a04d5c0771268197c90e99dccbcdf0ba9c90"),
				HashedSecret: AtomicSwapHashedSecret{1, 2, 3},
				TimeLock:     TestnetChainConstants().GenesisTimestamp,
				PublicKey: PublicKey{
					Algorithm: SignatureAlgoEd25519,
					Key:       ByteSlice{1, 2, 3},
				},
				Signature: ByteSlice{4, 5, 6},
				Secret:    AtomicSwapSecret{7, 8, 9},
			},
			&LegacyAtomicSwapFulfillment{
				Sender:       unlockHashFromHex("01746677df456546d93729066dd88514e2009930f3eebac3c93d43c88a108f8f9aa9e7c6f58893"),
				Receiver:     unlockHashFromHex("015fe50b9c596d8717e5e7ba79d5a7c9c8b82b1427a04d5c0771268197c90e99dccbcdf0ba9c90"),
				HashedSecret: AtomicSwapHashedSecret{1, 2, 3},
				TimeLock:     TestnetChainConstants().GenesisTimestamp,
				PublicKey: PublicKey{
					Algorithm: SignatureAlgoEd25519,
					Key:       ByteSlice{1, 2, 3},
				},
				Signature: ByteSlice{4, 5, 6},
				Secret:    AtomicSwapSecret{7, 8, 9},
			},
			"",
		},
		{
			&LegacyAtomicSwapFulfillment{
				Sender:       unlockHashFromHex("01746677df456546d93729066dd88514e2009930f3eebac3c93d43c88a108f8f9aa9e7c6f58893"),
				Receiver:     unlockHashFromHex("015fe50b9c596d8717e5e7ba79d5a7c9c8b82b1427a04d5c0771268197c90e99dccbcdf0ba9c90"),
				HashedSecret: AtomicSwapHashedSecret{1, 2, 3},
				TimeLock:     TestnetChainConstants().GenesisTimestamp,
				PublicKey: PublicKey{
					Algorithm: SignatureAlgoEd25519,
					Key:       ByteSlice{1, 2, 3},
				},
				Signature: ByteSlice{4, 5, 6},
				Secret:    AtomicSwapSecret{7, 8, 9},
			},
			&LegacyAtomicSwapFulfillment{
				Sender:       unlockHashFromHex("015fe50b9c596d8717e5e7ba79d5a7c9c8b82b1427a04d5c0771268197c90e99dccbcdf0ba9c90"),
				Receiver:     unlockHashFromHex("015fe50b9c596d8717e5e7ba79d5a7c9c8b82b1427a04d5c0771268197c90e99dccbcdf0ba9c90"),
				HashedSecret: AtomicSwapHashedSecret{1, 2, 3},
				TimeLock:     TestnetChainConstants().GenesisTimestamp,
				PublicKey: PublicKey{
					Algorithm: SignatureAlgoEd25519,
					Key:       ByteSlice{1, 2, 3},
				},
				Signature: ByteSlice{4, 5, 6},
				Secret:    AtomicSwapSecret{7, 8, 9},
			},
			"different sender",
		},
		{
			&LegacyAtomicSwapFulfillment{
				Sender:       unlockHashFromHex("01746677df456546d93729066dd88514e2009930f3eebac3c93d43c88a108f8f9aa9e7c6f58893"),
				Receiver:     unlockHashFromHex("01746677df456546d93729066dd88514e2009930f3eebac3c93d43c88a108f8f9aa9e7c6f58893"),
				HashedSecret: AtomicSwapHashedSecret{1, 2, 3},
				TimeLock:     TestnetChainConstants().GenesisTimestamp,
				PublicKey: PublicKey{
					Algorithm: SignatureAlgoEd25519,
					Key:       ByteSlice{1, 2, 3},
				},
				Signature: ByteSlice{4, 5, 6},
				Secret:    AtomicSwapSecret{7, 8, 9},
			},
			&LegacyAtomicSwapFulfillment{
				Sender:       unlockHashFromHex("01746677df456546d93729066dd88514e2009930f3eebac3c93d43c88a108f8f9aa9e7c6f58893"),
				Receiver:     unlockHashFromHex("015fe50b9c596d8717e5e7ba79d5a7c9c8b82b1427a04d5c0771268197c90e99dccbcdf0ba9c90"),
				HashedSecret: AtomicSwapHashedSecret{1, 2, 3},
				TimeLock:     TestnetChainConstants().GenesisTimestamp,
				PublicKey: PublicKey{
					Algorithm: SignatureAlgoEd25519,
					Key:       ByteSlice{1, 2, 3},
				},
				Signature: ByteSlice{4, 5, 6},
				Secret:    AtomicSwapSecret{7, 8, 9},
			},
			"different receiver",
		},
		{
			&LegacyAtomicSwapFulfillment{
				Sender:       unlockHashFromHex("01746677df456546d93729066dd88514e2009930f3eebac3c93d43c88a108f8f9aa9e7c6f58893"),
				Receiver:     unlockHashFromHex("015fe50b9c596d8717e5e7ba79d5a7c9c8b82b1427a04d5c0771268197c90e99dccbcdf0ba9c90"),
				HashedSecret: AtomicSwapHashedSecret{1, 2, 3},
				TimeLock:     TestnetChainConstants().GenesisTimestamp,
				PublicKey: PublicKey{
					Algorithm: SignatureAlgoEd25519,
					Key:       ByteSlice{1, 2, 3},
				},
				Signature: ByteSlice{4, 5, 6},
				Secret:    AtomicSwapSecret{7, 8, 9},
			},
			&LegacyAtomicSwapFulfillment{
				Sender:       unlockHashFromHex("01746677df456546d93729066dd88514e2009930f3eebac3c93d43c88a108f8f9aa9e7c6f58893"),
				Receiver:     unlockHashFromHex("015fe50b9c596d8717e5e7ba79d5a7c9c8b82b1427a04d5c0771268197c90e99dccbcdf0ba9c90"),
				HashedSecret: AtomicSwapHashedSecret{1, 2, 3},
				TimeLock:     0,
				PublicKey: PublicKey{
					Algorithm: SignatureAlgoEd25519,
					Key:       ByteSlice{1, 2, 3},
				},
				Signature: ByteSlice{4, 5, 6},
				Secret:    AtomicSwapSecret{7, 8, 9},
			},
			"different time lock",
		},
		{
			&LegacyAtomicSwapFulfillment{
				Sender:       unlockHashFromHex("01746677df456546d93729066dd88514e2009930f3eebac3c93d43c88a108f8f9aa9e7c6f58893"),
				Receiver:     unlockHashFromHex("015fe50b9c596d8717e5e7ba79d5a7c9c8b82b1427a04d5c0771268197c90e99dccbcdf0ba9c90"),
				HashedSecret: AtomicSwapHashedSecret{4, 2},
				TimeLock:     TestnetChainConstants().GenesisTimestamp,
				PublicKey: PublicKey{
					Algorithm: SignatureAlgoEd25519,
					Key:       ByteSlice{1, 2, 3},
				},
				Signature: ByteSlice{4, 5, 6},
				Secret:    AtomicSwapSecret{7, 8, 9},
			},
			&LegacyAtomicSwapFulfillment{
				Sender:       unlockHashFromHex("01746677df456546d93729066dd88514e2009930f3eebac3c93d43c88a108f8f9aa9e7c6f58893"),
				Receiver:     unlockHashFromHex("015fe50b9c596d8717e5e7ba79d5a7c9c8b82b1427a04d5c0771268197c90e99dccbcdf0ba9c90"),
				HashedSecret: AtomicSwapHashedSecret{1, 2, 3},
				TimeLock:     TestnetChainConstants().GenesisTimestamp,
				PublicKey: PublicKey{
					Algorithm: SignatureAlgoEd25519,
					Key:       ByteSlice{1, 2, 3},
				},
				Signature: ByteSlice{4, 5, 6},
				Secret:    AtomicSwapSecret{7, 8, 9},
			},
			"different hashed secret",
		},
		{&MultiSignatureFulfillment{}, nil, "unequal type"},
		{&MultiSignatureFulfillment{}, &NilFulfillment{}, "unequal type"},
		{&MultiSignatureFulfillment{}, &LegacyAtomicSwapFulfillment{}, "unequal type"},
		{&MultiSignatureFulfillment{}, &SingleSignatureFulfillment{}, "unequal type"},
		{&MultiSignatureFulfillment{}, &AtomicSwapFulfillment{}, "unequal type"},
		{&MultiSignatureFulfillment{}, &MultiSignatureFulfillment{}, ""},
		// case 45
		{
			&MultiSignatureFulfillment{
				Pairs: []PublicKeySignaturePair{
					{
						PublicKey: PublicKey{
							Algorithm: SignatureAlgoEd25519,
							Key:       ByteSlice{1, 2, 3},
						},
						Signature: ByteSlice{4, 5, 6},
					},
					{
						PublicKey: PublicKey{
							Algorithm: SignatureAlgoEd25519,
							Key:       ByteSlice{7, 1, 8},
						},
						Signature: ByteSlice{0, 45, 9},
					},
				},
			},
			&MultiSignatureFulfillment{
				Pairs: []PublicKeySignaturePair{
					{
						PublicKey: PublicKey{
							Algorithm: SignatureAlgoEd25519,
							Key:       ByteSlice{1, 2, 3},
						},
						Signature: ByteSlice{4, 5, 6},
					},
				},
			},
			"different amount of keys/signatures",
		},
		{
			&MultiSignatureFulfillment{
				Pairs: []PublicKeySignaturePair{
					{
						PublicKey: PublicKey{
							Algorithm: SignatureAlgoEd25519,
							Key:       ByteSlice{1, 2, 3},
						},
						Signature: ByteSlice{4, 5, 6},
					},
					{
						PublicKey: PublicKey{
							Algorithm: SignatureAlgoEd25519,
							Key:       ByteSlice{7, 1, 8},
						},
						Signature: ByteSlice{0, 45, 50},
					},
				},
			},
			&MultiSignatureFulfillment{
				Pairs: []PublicKeySignaturePair{
					{
						PublicKey: PublicKey{
							Algorithm: SignatureAlgoEd25519,
							Key:       ByteSlice{1, 2, 3},
						},
						Signature: ByteSlice{4, 5, 6},
					},
					{
						PublicKey: PublicKey{
							Algorithm: SignatureAlgoEd25519,
							Key:       ByteSlice{7, 1, 8},
						},
						Signature: ByteSlice{0, 45, 9},
					},
				},
			},
			"Different signature",
		},
		{
			&MultiSignatureFulfillment{
				Pairs: []PublicKeySignaturePair{
					{
						PublicKey: PublicKey{
							Algorithm: SignatureAlgoEd25519,
							Key:       ByteSlice{1, 2, 3},
						},
						Signature: ByteSlice{4, 5, 6},
					},
					{
						PublicKey: PublicKey{
							Algorithm: SignatureAlgoEd25519,
							Key:       ByteSlice{7, 1, 8},
						},
						Signature: ByteSlice{0, 45, 9},
					},
				},
			},
			&MultiSignatureFulfillment{
				Pairs: []PublicKeySignaturePair{
					{
						PublicKey: PublicKey{
							Algorithm: SignatureAlgoEd25519,
							Key:       ByteSlice{1, 2, 3},
						},
						Signature: ByteSlice{4, 5, 6},
					},
					{
						PublicKey: PublicKey{
							Algorithm: SignatureAlgoEd25519,
							Key:       ByteSlice{7, 10, 8},
						},
						Signature: ByteSlice{0, 45, 9},
					},
				},
			},
			"Different public key",
		},
		{
			&MultiSignatureFulfillment{
				Pairs: []PublicKeySignaturePair{
					{
						PublicKey: PublicKey{
							Algorithm: SignatureAlgoEd25519,
							Key:       ByteSlice{1, 2, 3},
						},
						Signature: ByteSlice{4, 5, 6},
					},
					{
						PublicKey: PublicKey{
							Algorithm: SignatureAlgoEd25519,
							Key:       ByteSlice{7, 1, 8},
						},
						Signature: ByteSlice{0, 45, 9},
					},
				},
			},
			&MultiSignatureFulfillment{
				Pairs: []PublicKeySignaturePair{
					{
						PublicKey: PublicKey{
							Algorithm: SignatureAlgoEd25519,
							Key:       ByteSlice{1, 2, 3},
						},
						Signature: ByteSlice{4, 5, 6},
					},
					{
						PublicKey: PublicKey{
							Algorithm: SignatureAlgoEd25519,
							Key:       ByteSlice{1, 2, 3},
						},
						Signature: ByteSlice{4, 5, 6},
					},
					{
						PublicKey: PublicKey{
							Algorithm: SignatureAlgoEd25519,
							Key:       ByteSlice{7, 1, 8},
						},
						Signature: ByteSlice{0, 45, 9},
					},
				},
			},
			"PublicKeySignaturePair not repeated",
		},
		{
			&MultiSignatureFulfillment{
				Pairs: []PublicKeySignaturePair{
					{
						PublicKey: PublicKey{
							Algorithm: SignatureAlgoEd25519,
							Key:       ByteSlice{1, 2, 3},
						},
						Signature: ByteSlice{4, 5, 6},
					},
					{
						PublicKey: PublicKey{
							Algorithm: SignatureAlgoEd25519,
							Key:       ByteSlice{1, 2, 3},
						},
						Signature: ByteSlice{4, 5, 6},
					},
					{
						PublicKey: PublicKey{
							Algorithm: SignatureAlgoEd25519,
							Key:       ByteSlice{7, 1, 8},
						},
						Signature: ByteSlice{0, 45, 9},
					},
				},
			},
			&MultiSignatureFulfillment{
				Pairs: []PublicKeySignaturePair{
					{
						PublicKey: PublicKey{
							Algorithm: SignatureAlgoEd25519,
							Key:       ByteSlice{1, 2, 3},
						},
						Signature: ByteSlice{4, 5, 6},
					},
					{
						PublicKey: PublicKey{
							Algorithm: SignatureAlgoEd25519,
							Key:       ByteSlice{1, 2, 3},
						},
						Signature: ByteSlice{4, 5, 6},
					},
					{
						PublicKey: PublicKey{
							Algorithm: SignatureAlgoEd25519,
							Key:       ByteSlice{7, 1, 8},
						},
						Signature: ByteSlice{0, 45, 9},
					},
				},
			},
			"",
		},
		// Case 50
		{
			&MultiSignatureFulfillment{
				Pairs: []PublicKeySignaturePair{
					{
						PublicKey: PublicKey{
							Algorithm: SignatureAlgoEd25519,
							Key:       ByteSlice{1, 2, 3},
						},
						Signature: ByteSlice{4, 5, 6},
					},
					{
						PublicKey: PublicKey{
							Algorithm: SignatureAlgoEd25519,
							Key:       ByteSlice{7, 1, 8},
						},
						Signature: ByteSlice{0, 45, 9},
					},
				},
			},
			&MultiSignatureFulfillment{
				Pairs: []PublicKeySignaturePair{
					{
						PublicKey: PublicKey{
							Algorithm: SignatureAlgoEd25519,
							Key:       ByteSlice{1, 2, 3},
						},
						Signature: ByteSlice{4, 5, 6},
					},
					{
						PublicKey: PublicKey{
							Algorithm: SignatureAlgoEd25519,
							Key:       ByteSlice{7, 1, 8},
						},
						Signature: ByteSlice{0, 45, 9},
					},
				},
			},
			"",
		},
		{
			&MultiSignatureFulfillment{
				Pairs: []PublicKeySignaturePair{
					{
						PublicKey: PublicKey{
							Algorithm: SignatureAlgoEd25519,
							Key:       ByteSlice{1, 2, 3},
						},
						Signature: ByteSlice{4, 5, 6},
					},
					{
						PublicKey: PublicKey{
							Algorithm: SignatureAlgoEd25519,
							Key:       ByteSlice{7, 1, 8},
						},
						Signature: ByteSlice{0, 45, 9},
					},
				},
			},
			&MultiSignatureFulfillment{
				Pairs: []PublicKeySignaturePair{
					{
						PublicKey: PublicKey{
							Algorithm: SignatureAlgoEd25519,
							Key:       ByteSlice{7, 1, 8},
						},
						Signature: ByteSlice{0, 45, 9},
					},
					{
						PublicKey: PublicKey{
							Algorithm: SignatureAlgoEd25519,
							Key:       ByteSlice{1, 2, 3},
						},
						Signature: ByteSlice{4, 5, 6},
					},
				},
			},
			"",
		},
		{
			&MultiSignatureFulfillment{
				Pairs: []PublicKeySignaturePair{
					{
						PublicKey: PublicKey{
							Algorithm: SignatureAlgoEd25519,
							Key:       ByteSlice{1, 2, 3},
						},
						Signature: ByteSlice{4, 5, 6},
					},
				},
			},
			&MultiSignatureFulfillment{
				Pairs: []PublicKeySignaturePair{
					{
						PublicKey: PublicKey{
							Algorithm: SignatureAlgoEd25519,
							Key:       ByteSlice{1, 2, 3},
						},
						Signature: ByteSlice{4, 5, 6},
					},
					{
						PublicKey: PublicKey{
							Algorithm: SignatureAlgoEd25519,
							Key:       ByteSlice{1, 2, 3},
						},
						Signature: ByteSlice{4, 5, 6},
					},
				},
			},
			"edge case, where second fulfillment has two times the pair the first fulfillment has",
		},
		{
			&MultiSignatureFulfillment{
				Pairs: []PublicKeySignaturePair{
					{
						PublicKey: PublicKey{
							Algorithm: SignatureAlgoEd25519,
							Key:       ByteSlice{1, 2, 3},
						},
						Signature: ByteSlice{4, 5, 6},
					},
					{
						PublicKey: PublicKey{
							Algorithm: SignatureAlgoEd25519,
							Key:       ByteSlice{7, 1, 8},
						},
						Signature: ByteSlice{0, 45, 9},
					},
				},
			},
			&MultiSignatureFulfillment{
				Pairs: []PublicKeySignaturePair{
					{
						PublicKey: PublicKey{
							Algorithm: SignatureAlgoEd25519,
							Key:       ByteSlice{1, 2, 3},
						},
						Signature: ByteSlice{4, 5, 6},
					},
					{
						PublicKey: PublicKey{
							Algorithm: SignatureAlgoEd25519,
							Key:       ByteSlice{1, 2, 3},
						},
						Signature: ByteSlice{4, 5, 6},
					},
				},
			},
			"edge case, where second fulfillment has two times the pair the first fulfillment has, but lenght is equal",
		},
		{
			&MultiSignatureFulfillment{
				Pairs: []PublicKeySignaturePair{
					{
						PublicKey: PublicKey{
							Algorithm: SignatureAlgoEd25519,
							Key:       ByteSlice{1, 2, 3},
						},
						Signature: ByteSlice{4, 5, 6},
					},
					{
						PublicKey: PublicKey{
							Algorithm: SignatureAlgoEd25519,
							Key:       ByteSlice{1, 2, 3},
						},
						Signature: ByteSlice{4, 5, 6},
					},
				},
			},
			&MultiSignatureFulfillment{
				Pairs: []PublicKeySignaturePair{
					{
						PublicKey: PublicKey{
							Algorithm: SignatureAlgoEd25519,
							Key:       ByteSlice{1, 2, 3},
						},
						Signature: ByteSlice{4, 5, 6},
					},
					{
						PublicKey: PublicKey{
							Algorithm: SignatureAlgoEd25519,
							Key:       ByteSlice{7, 1, 8},
						},
						Signature: ByteSlice{0, 45, 9},
					},
				},
			},
			"edge case, where first fulfillment has two times the pair the second fulfillment has, but lenght is equal",
		},
	}
	for idx, testCase := range testCases {
		equal := testCase.A.Equal(testCase.B)
		if testCase.NotEqualMessage != "" {
			if equal {
				t.Error(idx, "expected not equal, but it's equal:", testCase.NotEqualMessage, testCase.A, testCase.B)
			}
		} else {
			if !equal {
				t.Error(idx, "expected equal, but it's not equal", testCase.A, testCase.B)
			}
		}
	}
}

func TestFulfillLegacyCompatibility(t *testing.T) {
	// utility funcs
	hbs := func(str string) []byte { // hexStr -> byte slice
		bs, _ := hex.DecodeString(str)
		return bs
	}
	hs := func(str string) (hash crypto.Hash) { // hbs -> crypto.Hash
		copy(hash[:], hbs(str))
		return
	}

	testCases := []struct {
		Transaction                      Transaction
		CoinConditions                   []MarshalableUnlockCondition
		ExpectedCoinIdentifiers          []CoinOutputID
		ExpectedCoinInputSigHashes       []ByteSlice
		BlockStakeConditions             []MarshalableUnlockCondition
		ExpectedBlockStakeIdentifiers    []BlockStakeOutputID
		ExpectedBlockStakeInputSigHashes []ByteSlice
		ExpectedTransactionIdentifier    TransactionID
	}{
		{
			Transaction{
				Version: TransactionVersionZero,
				BlockStakeInputs: []BlockStakeInput{
					{
						ParentID: BlockStakeOutputID(hs("a4292b24a9868649efa7ec49221b97043554eefb4be92de8d6ac885c2fa533c4")),
						Fulfillment: NewFulfillment(&SingleSignatureFulfillment{
							PublicKey: PublicKey{
								Algorithm: SignatureAlgoEd25519,
								Key:       hbs("8d368f6c457f1f7f49f4cb32636c1d34197c046f5398ea6661b0b4ecfe36a3cd"),
							},
							Signature: hbs("248fce862f030e5e98962b43cb437a809aa30ba99367db018e410c8a6854be88a03c07c9f788fe75d0f12af9ddc39f9c9508aa55283a6ac02c41e8cc7be8f303"),
						}),
					},
				},
				BlockStakeOutputs: []BlockStakeOutput{
					{
						Value: NewCurrency64(400),
						Condition: NewCondition(NewUnlockHashCondition(
							unlockHashFromHex("01746677df456546d93729066dd88514e2009930f3eebac3c93d43c88a108f8f9aa9e7c6f58893"))),
					},
				},
			},
			nil,
			nil,
			nil,
			[]MarshalableUnlockCondition{
				&UnlockHashCondition{
					TargetUnlockHash: unlockHashFromHex("01746677df456546d93729066dd88514e2009930f3eebac3c93d43c88a108f8f9aa9e7c6f58893"),
				},
			},
			[]BlockStakeOutputID{
				BlockStakeOutputID(hs("03ee547f5efbc60cef3f185471a532faa284f3ec3900da8a929525ba459708d5")),
			},
			[]ByteSlice{
				hbs("798144cd1e876daf6f0d5008547dbda8ae69ef3b7dd94555d7c14e6e5ccdeeda"),
			},
			TransactionID(hs("6255ff840923595598a134795a66814e512395f5c9e96669e7f2c104c98ff090")),
		},
		{
			Transaction{
				Version: TransactionVersionZero,
				CoinInputs: []CoinInput{
					{
						ParentID: CoinOutputID(hs("9a3b7ea912f6438eec826b49b71876e92b09624621a51c8f1ca76645a54cab4a")),
						Fulfillment: NewFulfillment(&SingleSignatureFulfillment{
							PublicKey: PublicKey{
								Algorithm: SignatureAlgoEd25519,
								Key:       hbs("07fa00de51b678926885e96fb1904d3eebca2c283dee40e975871ed6109f7f4b"),
							},
							Signature: hbs("ae8e2891033e260bf35f7c340823818a46cb6240aac8aa4bcdadecf30604b54d339ec9930b8be95a9a779bb48027e6314d8b2f701809cd352b1d14753a145f01"),
						}),
					},
				},
				CoinOutputs: []CoinOutput{
					{
						Value: NewCurrency64(100000000000),
						Condition: NewCondition(NewUnlockHashCondition(
							unlockHashFromHex("015fe50b9c596d8717e5e7ba79d5a7c9c8b82b1427a04d5c0771268197c90e99dccbcdf0ba9c90"))),
					},
					{
						Value: NewCurrency64(694999899800000000),
						Condition: NewCondition(NewUnlockHashCondition(
							unlockHashFromHex("01d3a8d366864f5f368bd73959139c55da5f1f8beaa07cb43519cc87d2a51135ae0b3ba93cf2d9"))),
					},
				},
				MinerFees: []Currency{
					NewCurrency64(100000000),
				},
			},
			[]MarshalableUnlockCondition{
				&UnlockHashCondition{
					TargetUnlockHash: unlockHashFromHex("01437c56286c76dec14e87f5da5e5a436651006e6cd46bee5865c9060ba178f7296ed843b70a57"),
				},
			},
			[]CoinOutputID{
				CoinOutputID(hs("8c193a699d27799efebb52e501ed7fdbc4da38a3cf539c431e9659734e23827d")),
				CoinOutputID(hs("2829711c7dd071d3d3031d30eedbae3d126d62d3ac3369b01cecdda7d2aebfef")),
			},
			[]ByteSlice{
				hbs("b6847f66a5437ef11250eebd0eccb7454dca395e1de68a6f7f86f3c5014a238d"),
			},
			nil,
			nil,
			nil,
			TransactionID(hs("f6f7c6bd071ea9403d07a74c865e5aa2074564cd557e81746a945695c0dcf579")),
		},
	}
	for tidx, testCase := range testCases {
		for idx, ci := range testCase.Transaction.CoinInputs {
			sigHash, err := testCase.Transaction.SignatureHash(uint64(idx))
			if err != nil {
				t.Error(tidx, idx, "unexpected error", err)
			}
			if bytes.Compare(testCase.ExpectedCoinInputSigHashes[idx][:], sigHash[:]) != 0 {
				t.Error(tidx, idx, "invalid coin input sigh hash",
					testCase.ExpectedCoinInputSigHashes[idx], "!=", sigHash)
			}

			err = ci.Fulfillment.IsStandardFulfillment(ValidationContext{})
			if err != nil {
				t.Error(tidx, idx, "unexpected error", err)
			}

			err = (UnlockConditionProxy{testCase.CoinConditions[idx]}).Fulfill(ci.Fulfillment, FulfillContext{
				ExtraObjects: []interface{}{uint64(idx)},
				Transaction:  testCase.Transaction,
			})
			if err != nil {
				t.Error(tidx, idx, err)
			}
		}
		for idx, bsi := range testCase.Transaction.BlockStakeInputs {
			sigHash, err := testCase.Transaction.SignatureHash(uint64(idx))
			if err != nil {
				t.Error(tidx, idx, "unexpected error", err)
			}
			if bytes.Compare(testCase.ExpectedBlockStakeInputSigHashes[idx][:], sigHash[:]) != 0 {
				t.Error(tidx, idx, "invalid bs input sigh hash",
					testCase.ExpectedBlockStakeInputSigHashes[idx], "!=", sigHash)
			}

			err = bsi.Fulfillment.IsStandardFulfillment(ValidationContext{})
			if err != nil {
				t.Error(tidx, idx, "unexpected error", err)
			}

			err = (UnlockConditionProxy{testCase.BlockStakeConditions[idx]}).Fulfill(bsi.Fulfillment, FulfillContext{
				ExtraObjects: []interface{}{uint64(idx)},
				Transaction:  testCase.Transaction,
			})
			if err != nil {
				t.Error(tidx, idx, err)
			}
		}
		for idx, co := range testCase.Transaction.CoinOutputs {
			outputID := testCase.Transaction.CoinOutputID(uint64(idx))
			if bytes.Compare(testCase.ExpectedCoinIdentifiers[idx][:], outputID[:]) != 0 {
				t.Error(tidx, idx, testCase.ExpectedCoinIdentifiers[idx], "!=", outputID)
			}

			err := co.Condition.IsStandardCondition(ValidationContext{})
			if err != nil {
				t.Error(tidx, idx, "unexpected error", err)
			}
		}
		for idx, bso := range testCase.Transaction.BlockStakeOutputs {
			outputID := testCase.Transaction.BlockStakeOutputID(uint64(idx))
			if bytes.Compare(testCase.ExpectedBlockStakeIdentifiers[idx][:], outputID[:]) != 0 {
				t.Error(tidx, idx, testCase.ExpectedBlockStakeIdentifiers[idx], "!=", outputID)
			}

			err := bso.Condition.IsStandardCondition(ValidationContext{})
			if err != nil {
				t.Error(tidx, idx, "unexpected error", err)
			}
		}
		transactionID := testCase.Transaction.ID()
		if bytes.Compare(testCase.ExpectedTransactionIdentifier[:], transactionID[:]) != 0 {
			t.Error(tidx, testCase.ExpectedTransactionIdentifier, "!=", transactionID)
		}
	}
}

func TestValidFulFill(t *testing.T) {
	// test public/private key pair
	sk, pk := crypto.GenerateKeyPair()
	ed25519pk := Ed25519PublicKey(pk)

	// Second keypair
	sk2, pk2 := crypto.GenerateKeyPair()
	ed25519pk2 := Ed25519PublicKey(pk2)

	// future time stamp
	futureTimeStamp := CurrentTimestamp() + 123456

	testCases := []signAndFulfillInput{
		{ // nil -> single signature
			&NilCondition{},
			func() MarshalableUnlockFulfillment {
				return &SingleSignatureFulfillment{
					PublicKey: ed25519pk,
				}
			},
			sk,
		},
		{ // unlock hash -> single signature
			&UnlockHashCondition{
				TargetUnlockHash: NewUnlockHash(UnlockTypePubKey, crypto.HashObject(siabin.Marshal(ed25519pk))),
			},
			func() MarshalableUnlockFulfillment {
				return &SingleSignatureFulfillment{
					PublicKey: ed25519pk,
				}
			},
			sk,
		},
		{ // [LEGACY] unlock hash -> atomic swap (refund)
			&UnlockHashCondition{
				TargetUnlockHash: (&AtomicSwapCondition{
					Sender:       NewUnlockHash(UnlockTypePubKey, crypto.HashObject(siabin.Marshal(ed25519pk))),
					Receiver:     unlockHashFromHex("01437c56286c76dec14e87f5da5e5a436651006e6cd46bee5865c9060ba178f7296ed843b70a57"),
					HashedSecret: NewAtomicSwapHashedSecret(AtomicSwapSecret{4, 2}),
					TimeLock:     42, // we are waaaaay beyond this
				}).UnlockHash(),
			},
			func() MarshalableUnlockFulfillment {
				return &LegacyAtomicSwapFulfillment{
					Sender:       NewUnlockHash(UnlockTypePubKey, crypto.HashObject(siabin.Marshal(ed25519pk))),
					Receiver:     unlockHashFromHex("01437c56286c76dec14e87f5da5e5a436651006e6cd46bee5865c9060ba178f7296ed843b70a57"),
					HashedSecret: NewAtomicSwapHashedSecret(AtomicSwapSecret{4, 2}),
					TimeLock:     42, // we are waaaaay beyond this
					PublicKey:    ed25519pk,
					Secret:       AtomicSwapSecret{}, // refund as claiming is impossible due to time lock
					// Signature is set at signing step
				}
			},
			sk,
		},
		{ // [LEGACY] unlock hash -> atomic swap (refund)
			&UnlockHashCondition{
				TargetUnlockHash: (&AtomicSwapCondition{
					Sender:       NewUnlockHash(UnlockTypePubKey, crypto.HashObject(siabin.Marshal(ed25519pk))),
					Receiver:     unlockHashFromHex("01437c56286c76dec14e87f5da5e5a436651006e6cd46bee5865c9060ba178f7296ed843b70a57"),
					HashedSecret: NewAtomicSwapHashedSecret(AtomicSwapSecret{4, 2}),
					TimeLock:     42, // we are waaaaay beyond this
				}).UnlockHash(),
			},
			func() MarshalableUnlockFulfillment {
				return &anyAtomicSwapFulfillment{
					&LegacyAtomicSwapFulfillment{
						Sender:       NewUnlockHash(UnlockTypePubKey, crypto.HashObject(siabin.Marshal(ed25519pk))),
						Receiver:     unlockHashFromHex("01437c56286c76dec14e87f5da5e5a436651006e6cd46bee5865c9060ba178f7296ed843b70a57"),
						HashedSecret: NewAtomicSwapHashedSecret(AtomicSwapSecret{4, 2}),
						TimeLock:     42, // we are waaaaay beyond this
						PublicKey:    ed25519pk,
						Secret:       AtomicSwapSecret{}, // refund as claiming is impossible due to time lock
						// Signature is set at signing step
					},
				}
			},
			sk,
		},
		{ // [LEGACY] unlock hash -> atomic swap (claim)
			&UnlockHashCondition{
				TargetUnlockHash: (&AtomicSwapCondition{
					Receiver:     NewUnlockHash(UnlockTypePubKey, crypto.HashObject(siabin.Marshal(ed25519pk))),
					Sender:       unlockHashFromHex("01437c56286c76dec14e87f5da5e5a436651006e6cd46bee5865c9060ba178f7296ed843b70a57"),
					HashedSecret: NewAtomicSwapHashedSecret(AtomicSwapSecret{4, 2}),
					TimeLock:     futureTimeStamp,
				}).UnlockHash(),
			},
			func() MarshalableUnlockFulfillment {
				return &LegacyAtomicSwapFulfillment{
					Receiver:     NewUnlockHash(UnlockTypePubKey, crypto.HashObject(siabin.Marshal(ed25519pk))),
					Sender:       unlockHashFromHex("01437c56286c76dec14e87f5da5e5a436651006e6cd46bee5865c9060ba178f7296ed843b70a57"),
					HashedSecret: NewAtomicSwapHashedSecret(AtomicSwapSecret{4, 2}),
					TimeLock:     futureTimeStamp,
					PublicKey:    ed25519pk,
					Secret:       AtomicSwapSecret{4, 2},
					// Signature is set at signing step
				}
			},
			sk,
		},
		{ // atomic swap -> atomic swap (refund)
			&AtomicSwapCondition{
				Sender:       NewUnlockHash(UnlockTypePubKey, crypto.HashObject(siabin.Marshal(ed25519pk))),
				Receiver:     unlockHashFromHex("01437c56286c76dec14e87f5da5e5a436651006e6cd46bee5865c9060ba178f7296ed843b70a57"),
				HashedSecret: NewAtomicSwapHashedSecret(AtomicSwapSecret{4, 2}),
				TimeLock:     42, // we are waaaaay beyond this
			},
			func() MarshalableUnlockFulfillment {
				return &AtomicSwapFulfillment{
					PublicKey: ed25519pk,
					Secret:    AtomicSwapSecret{}, // refund as claiming is impossible due to time lock
					// Signature is set at signing step
				}
			},
			sk,
		},
		{ // atomic swap -> atomic swap (claim)
			&AtomicSwapCondition{
				Receiver:     NewUnlockHash(UnlockTypePubKey, crypto.HashObject(siabin.Marshal(ed25519pk))),
				Sender:       unlockHashFromHex("01437c56286c76dec14e87f5da5e5a436651006e6cd46bee5865c9060ba178f7296ed843b70a57"),
				HashedSecret: NewAtomicSwapHashedSecret(AtomicSwapSecret{4, 2}),
				TimeLock:     futureTimeStamp,
			},
			func() MarshalableUnlockFulfillment {
				return &AtomicSwapFulfillment{
					PublicKey: ed25519pk,
					Secret:    AtomicSwapSecret{4, 2},
					// Signature is set at signing step
				}
			},
			sk,
		},
		{ // atomic swap -> atomic swap (claim, possible even when expired)
			&AtomicSwapCondition{
				Receiver:     NewUnlockHash(UnlockTypePubKey, crypto.HashObject(siabin.Marshal(ed25519pk))),
				Sender:       unlockHashFromHex("01437c56286c76dec14e87f5da5e5a436651006e6cd46bee5865c9060ba178f7296ed843b70a57"),
				HashedSecret: NewAtomicSwapHashedSecret(AtomicSwapSecret{4, 2}),
				TimeLock:     42,
			},
			func() MarshalableUnlockFulfillment {
				return &AtomicSwapFulfillment{
					PublicKey: ed25519pk,
					Secret:    AtomicSwapSecret{4, 2},
					// Signature is set at signing step
				}
			},
			sk,
		},
		{ // atomic swap -> atomic swap (refund)
			&AtomicSwapCondition{
				Sender:       NewUnlockHash(UnlockTypePubKey, crypto.HashObject(siabin.Marshal(ed25519pk))),
				Receiver:     unlockHashFromHex("01437c56286c76dec14e87f5da5e5a436651006e6cd46bee5865c9060ba178f7296ed843b70a57"),
				HashedSecret: NewAtomicSwapHashedSecret(AtomicSwapSecret{4, 2}),
				TimeLock:     42, // we are waaaaay beyond this
			},
			func() MarshalableUnlockFulfillment {
				return &anyAtomicSwapFulfillment{
					&AtomicSwapFulfillment{
						PublicKey: ed25519pk,
						Secret:    AtomicSwapSecret{}, // refund as claiming is impossible due to time lock
						// Signature is set at signing step
					},
				}
			},
			sk,
		},
		{ // atomic swap -> atomic swap (claim)
			&AtomicSwapCondition{
				Receiver:     NewUnlockHash(UnlockTypePubKey, crypto.HashObject(siabin.Marshal(ed25519pk))),
				Sender:       unlockHashFromHex("01437c56286c76dec14e87f5da5e5a436651006e6cd46bee5865c9060ba178f7296ed843b70a57"),
				HashedSecret: NewAtomicSwapHashedSecret(AtomicSwapSecret{4, 2}),
				TimeLock:     futureTimeStamp,
			},
			func() MarshalableUnlockFulfillment {
				return &anyAtomicSwapFulfillment{
					&AtomicSwapFulfillment{
						PublicKey: ed25519pk,
						Secret:    AtomicSwapSecret{4, 2},
						// Signature is set at signing step
					},
				}
			},
			sk,
		},
		{ // TimeLockedCondition (Nil) -> SingleSignature
			&TimeLockCondition{
				LockTime:  uint64(CurrentTimestamp()),
				Condition: &NilCondition{},
			},
			func() MarshalableUnlockFulfillment {
				return &SingleSignatureFulfillment{
					PublicKey: ed25519pk,
				}
			},
			sk,
		},
		{ // TimeLockedCondition (UnlockHash) -> TimeLockedFulfillment (SingleSignature)
			&TimeLockCondition{
				LockTime: uint64(CurrentTimestamp()),
				Condition: &UnlockHashCondition{
					TargetUnlockHash: NewUnlockHash(UnlockTypePubKey, crypto.HashObject(siabin.Marshal(ed25519pk))),
				},
			},
			func() MarshalableUnlockFulfillment {
				return &SingleSignatureFulfillment{
					PublicKey: ed25519pk,
				}
			},
			sk,
		},
		{ // TimeLockedCondition (MultiSignatureCondition) -> MultiSignatureFulfillment
			&TimeLockCondition{
				LockTime: uint64(CurrentTimestamp()),
				Condition: &MultiSignatureCondition{
					UnlockHashes: UnlockHashSlice{
						NewUnlockHash(UnlockTypePubKey, crypto.HashObject(siabin.Marshal(ed25519pk))),
						NewUnlockHash(UnlockTypePubKey, crypto.HashObject(siabin.Marshal(ed25519pk2))),
					},
					MinimumSignatureCount: 1,
				},
			},
			func() MarshalableUnlockFulfillment {
				return &MultiSignatureFulfillment{}
			},
			KeyPair{
				PublicKey:  ed25519pk2,
				PrivateKey: sk2[:],
			},
		},
		{
			&MultiSignatureCondition{
				UnlockHashes: UnlockHashSlice{
					NewUnlockHash(UnlockTypePubKey, crypto.HashObject(siabin.Marshal(ed25519pk))),
					NewUnlockHash(UnlockTypePubKey, crypto.HashObject(siabin.Marshal(ed25519pk2))),
				},
				MinimumSignatureCount: 1,
			},
			func() MarshalableUnlockFulfillment {
				return &MultiSignatureFulfillment{}
			},
			KeyPair{
				PublicKey:  ed25519pk2,
				PrivateKey: sk2[:],
			},
		},
	}
	for idx, testCase := range testCases {
		// test each testcase separately
		testValidSignAndFulfill(t, idx, []signAndFulfillInput{testCase})
	}
	// test all test cases at once
	testValidSignAndFulfill(t, len(testCases), testCases)
}

type signAndFulfillInput struct {
	Condition   MarshalableUnlockCondition
	FulFillment func() MarshalableUnlockFulfillment
	SignKey     interface{}
}

func testValidSignAndFulfill(t *testing.T, testIndex int, inputs []signAndFulfillInput) { // utility funcs
	rhs := func() (hash crypto.Hash) { // random crypto hash
		rand.Read(hash[:])
		return
	}

	// create transaction
	txn := Transaction{
		Version:   TransactionVersionOne,
		MinerFees: []Currency{NewCurrency64(1000 * 1000 * 100 * uint64(mrand.Int31n(100)+20))},
		ArbitraryData: func() []byte {
			b := make([]byte, mrand.Int31n(242)+14)
			rand.Read(b[:])
			return b
		}(),
	}

	// add one coin input and blockstake input per input param
	for idx, input := range inputs {
		txn.CoinInputs = append(txn.CoinInputs, CoinInput{
			ParentID:    CoinOutputID(rhs()),
			Fulfillment: NewFulfillment(input.FulFillment()),
		})
		txn.CoinOutputs = append(txn.CoinOutputs, CoinOutput{
			Value:     NewCurrency64(1000 * 1000 * 1000 * uint64(mrand.Int31n(100)+20)),
			Condition: NewCondition(inputs[(idx+1)%len(inputs)].Condition),
		})
		txn.BlockStakeInputs = append(txn.BlockStakeInputs, BlockStakeInput{
			ParentID:    BlockStakeOutputID(rhs()),
			Fulfillment: NewFulfillment(input.FulFillment()),
		})
		txn.BlockStakeOutputs = append(txn.BlockStakeOutputs, BlockStakeOutput{
			Value:     NewCurrency64(uint64(mrand.Int31n(100) + 20)),
			Condition: NewCondition(inputs[(idx+2)%len(inputs)].Condition),
		})
	}

	// sign all inputs (unless we are playing with unknown fulfillments)
	for idx, input := range inputs {
		if fulfillmentCanBeSigned(txn.CoinInputs[idx].Fulfillment.Fulfillment) {
			signContext := FulfillmentSignContext{
				ExtraObjects: []interface{}{uint64(idx)},
				Transaction:  txn,
				Key:          input.SignKey,
			}
			err := txn.CoinInputs[idx].Fulfillment.Sign(signContext)
			if err != nil {
				t.Error(testIndex, idx, "signing coin input failed", err)
			}
			err = txn.BlockStakeInputs[idx].Fulfillment.Sign(signContext)
			if err != nil {
				t.Error(testIndex, idx, "signing block stake input failed", err)
			}
		}
	}

	// fulfill all inputs
	for idx := range inputs {
		fulfillContext := FulfillContext{
			ExtraObjects: []interface{}{uint64(idx)},
			BlockHeight:  0,                  // not important for now
			BlockTime:    CurrentTimestamp(), // not important for now,
			Transaction:  txn,
		}
		err := (UnlockConditionProxy{inputs[idx].Condition}).Fulfill(txn.CoinInputs[idx].Fulfillment, fulfillContext)
		if err != nil {
			t.Error(testIndex, idx, "fulfilling coin input failed", err)
		}
		err = (UnlockConditionProxy{inputs[idx].Condition}).Fulfill(txn.BlockStakeInputs[idx].Fulfillment, fulfillContext)
		if err != nil {
			t.Error(testIndex, idx, "fulfilling block stake input failed", err)
		}
	}
}

func fulfillmentCanBeSigned(fulfillment MarshalableUnlockFulfillment) bool {
	return fulfillment != nil
}

func TestIsStandardCondition(t *testing.T) {
	testCases := []struct {
		Condition          UnlockCondition
		NotStandardMessage string
	}{
		// nil conditions
		{NewCondition(nil), ""},
		{NewCondition(&NilCondition{}), ""},
		{&NilCondition{}, ""},
		// unlock hash condition
		{&UnlockHashCondition{}, "nil unlock type not allowed"},
		{&UnlockHashCondition{TargetUnlockHash: UnlockHash{Type: 255}}, "non-standard unlock type not allowed"},
		{&UnlockHashCondition{TargetUnlockHash: UnlockHash{Type: UnlockTypePubKey}}, "nil crypto hash not allowed"},
		{&UnlockHashCondition{TargetUnlockHash: UnlockHash{Type: UnlockTypeAtomicSwap}}, "nil crypto hash not allowed"},
		{&UnlockHashCondition{TargetUnlockHash: UnlockHash{Type: UnlockTypePubKey, Hash: crypto.Hash{1}}}, ""},
		{&UnlockHashCondition{TargetUnlockHash: UnlockHash{Type: UnlockTypeAtomicSwap, Hash: crypto.Hash{1}}}, ""},
		// atomic swap condition
		{&AtomicSwapCondition{}, "nil atomic swap condition not allowed"},
		{
			&AtomicSwapCondition{
				Sender:       unlockHashFromHex("01746677df456546d93729066dd88514e2009930f3eebac3c93d43c88a108f8f9aa9e7c6f58893"),
				Receiver:     unlockHashFromHex("02a24c97c80eeac111aa4bcbb0ac8ffc364fa9b22da10d3054778d2332f68b365e5e5af8e71541"),
				HashedSecret: AtomicSwapHashedSecret{4, 5, 6},
				TimeLock:     TestnetChainConstants().GenesisTimestamp,
			}, "receiver has unsupported unlock hash type",
		},
		{
			&AtomicSwapCondition{
				Sender:       unlockHashFromHex("02a24c97c80eeac111aa4bcbb0ac8ffc364fa9b22da10d3054778d2332f68b365e5e5af8e71541"),
				Receiver:     unlockHashFromHex("01746677df456546d93729066dd88514e2009930f3eebac3c93d43c88a108f8f9aa9e7c6f58893"),
				HashedSecret: AtomicSwapHashedSecret{4, 5, 6},
				TimeLock:     TestnetChainConstants().GenesisTimestamp,
			}, "sender has unsupported unlock hash type",
		},
		{
			&AtomicSwapCondition{
				Sender:       unlockHashFromHex("01746677df456546d93729066dd88514e2009930f3eebac3c93d43c88a108f8f9aa9e7c6f58893"),
				Receiver:     unlockHashFromHex("015fe50b9c596d8717e5e7ba79d5a7c9c8b82b1427a04d5c0771268197c90e99dccbcdf0ba9c90"),
				HashedSecret: AtomicSwapHashedSecret{},
				TimeLock:     TestnetChainConstants().GenesisTimestamp,
			}, "empty/nil hashed secret not allowed",
		},
		{
			&AtomicSwapCondition{
				Sender:       unlockHashFromHex("01746677df456546d93729066dd88514e2009930f3eebac3c93d43c88a108f8f9aa9e7c6f58893"),
				Receiver:     unlockHashFromHex("015fe50b9c596d8717e5e7ba79d5a7c9c8b82b1427a04d5c0771268197c90e99dccbcdf0ba9c90"),
				HashedSecret: AtomicSwapHashedSecret{1, 2, 3},
				TimeLock:     0,
			}, "",
		},
		{
			&AtomicSwapCondition{
				Sender:       unlockHashFromHex("01746677df456546d93729066dd88514e2009930f3eebac3c93d43c88a108f8f9aa9e7c6f58893"),
				Receiver:     unlockHashFromHex("015fe50b9c596d8717e5e7ba79d5a7c9c8b82b1427a04d5c0771268197c90e99dccbcdf0ba9c90"),
				HashedSecret: AtomicSwapHashedSecret{4, 5, 6},
				TimeLock:     TestnetChainConstants().GenesisTimestamp,
			}, "",
		},
		// time lock condition
		{
			&TimeLockCondition{
				LockTime:  42,
				Condition: &NilCondition{},
			}, "",
		},
		{
			&TimeLockCondition{
				Condition: &NilCondition{},
			}, "lock time has to be defined",
		},
		{
			&TimeLockCondition{
				LockTime: 1,
				Condition: &UnlockHashCondition{
					TargetUnlockHash: unlockHashFromHex("015fe50b9c596d8717e5e7ba79d5a7c9c8b82b1427a04d5c0771268197c90e99dccbcdf0ba9c90"),
				},
			}, "",
		},
		{
			&TimeLockCondition{
				LockTime: 1,
				Condition: &UnlockHashCondition{
					TargetUnlockHash: unlockHashFromHex("02a24c97c80eeac111aa4bcbb0ac8ffc364fa9b22da10d3054778d2332f68b365e5e5af8e71541"),
				},
			}, "non-standard unlock hash type",
		},
		{
			&TimeLockCondition{
				LockTime: 0,
				Condition: &UnlockHashCondition{
					TargetUnlockHash: unlockHashFromHex("015fe50b9c596d8717e5e7ba79d5a7c9c8b82b1427a04d5c0771268197c90e99dccbcdf0ba9c90"),
				},
			}, "no lock time provided",
		},
		{
			&TimeLockCondition{
				LockTime: 1,
				Condition: &MultiSignatureCondition{
					MinimumSignatureCount: 2,
					UnlockHashes: UnlockHashSlice{
						unlockHashFromHex("015fe50b9c596d8717e5e7ba79d5a7c9c8b82b1427a04d5c0771268197c90e99dccbcdf0ba9c90"),
						unlockHashFromHex("01fc8714235d549f890f35e52d745b9eeeee34926f96c4b9ef1689832f338d9349b453898f7e51"),
						unlockHashFromHex("01746677df456546d93729066dd88514e2009930f3eebac3c93d43c88a108f8f9aa9e7c6f58893"),
					},
				},
			}, "",
		},
		{
			&TimeLockCondition{
				LockTime: 0,
				Condition: &MultiSignatureCondition{
					MinimumSignatureCount: 2,
					UnlockHashes: UnlockHashSlice{
						unlockHashFromHex("015fe50b9c596d8717e5e7ba79d5a7c9c8b82b1427a04d5c0771268197c90e99dccbcdf0ba9c90"),
						unlockHashFromHex("01fc8714235d549f890f35e52d745b9eeeee34926f96c4b9ef1689832f338d9349b453898f7e51"),
						unlockHashFromHex("01746677df456546d93729066dd88514e2009930f3eebac3c93d43c88a108f8f9aa9e7c6f58893"),
					},
				},
			}, "no lock time provided",
		},
		{
			&TimeLockCondition{
				LockTime: 1,
				Condition: &TimeLockCondition{
					LockTime:  4,
					Condition: &NilCondition{},
				},
			}, "no time lock condition can be used as the internal condition of antoher time lock condition",
		},
		{
			&TimeLockCondition{
				LockTime: 1,
				Condition: &TimeLockCondition{
					LockTime: 4,
					Condition: &UnlockHashCondition{
						TargetUnlockHash: unlockHashFromHex("015fe50b9c596d8717e5e7ba79d5a7c9c8b82b1427a04d5c0771268197c90e99dccbcdf0ba9c90"),
					},
				},
			}, "no time lock condition can be used as the internal condition of antoher time lock condition",
		},
		{
			&TimeLockCondition{
				LockTime: 1,
				Condition: &AtomicSwapCondition{
					Sender:       unlockHashFromHex("01746677df456546d93729066dd88514e2009930f3eebac3c93d43c88a108f8f9aa9e7c6f58893"),
					Receiver:     unlockHashFromHex("015fe50b9c596d8717e5e7ba79d5a7c9c8b82b1427a04d5c0771268197c90e99dccbcdf0ba9c90"),
					HashedSecret: AtomicSwapHashedSecret{4, 5, 6},
					TimeLock:     TestnetChainConstants().GenesisTimestamp,
				},
			}, "no atomic swap condition can be used as the internal condition of a time lock condition",
		},
		{
			&MultiSignatureCondition{},
			"amount of required signatures must be greater than one",
		},
		{
			&MultiSignatureCondition{
				MinimumSignatureCount: 1,
				UnlockHashes: UnlockHashSlice{
					unlockHashFromHex("015fe50b9c596d8717e5e7ba79d5a7c9c8b82b1427a04d5c0771268197c90e99dccbcdf0ba9c90"),
				},
			},
			"at least one unlock hash is required",
		},
		{
			&MultiSignatureCondition{
				MinimumSignatureCount: 2,
				UnlockHashes: UnlockHashSlice{
					unlockHashFromHex("015fe50b9c596d8717e5e7ba79d5a7c9c8b82b1427a04d5c0771268197c90e99dccbcdf0ba9c90"),
				},
			},
			"amount of required signatures must be less then or equal to the amount of provided unlock hashes",
		},
		{
			&MultiSignatureCondition{
				MinimumSignatureCount: 2,
				UnlockHashes: UnlockHashSlice{
					unlockHashFromHex("015fe50b9c596d8717e5e7ba79d5a7c9c8b82b1427a04d5c0771268197c90e99dccbcdf0ba9c90"),
					unlockHashFromHex("01fc8714235d549f890f35e52d745b9eeeee34926f96c4b9ef1689832f338d9349b453898f7e51"),
					unlockHashFromHex("01746677df456546d93729066dd88514e2009930f3eebac3c93d43c88a108f8f9aa9e7c6f58893"),
				},
			},
			"",
		},
		{
			&MultiSignatureCondition{
				MinimumSignatureCount: 2,
				UnlockHashes: UnlockHashSlice{
					unlockHashFromHex("015fe50b9c596d8717e5e7ba79d5a7c9c8b82b1427a04d5c0771268197c90e99dccbcdf0ba9c90"),
					unlockHashFromHex("02a24c97c80eeac111aa4bcbb0ac8ffc364fa9b22da10d3054778d2332f68b365e5e5af8e71541"),
					unlockHashFromHex("01746677df456546d93729066dd88514e2009930f3eebac3c93d43c88a108f8f9aa9e7c6f58893"),
				},
			},
			"only pubKey unlockhashes are allowed",
		},
	}
	for idx, testCase := range testCases {
		err := testCase.Condition.IsStandardCondition(ValidationContext{})
		if testCase.NotStandardMessage != "" {
			if err == nil {
				t.Error(idx, "expected error, but none received:", testCase.NotStandardMessage, testCase.Condition)
			}
		} else {
			if err != nil {
				t.Error(idx, "expected no error but received one:", err, testCase.Condition)
			}
		}
	}
}

func TestIsStandardFulfillment(t *testing.T) {
	testCases := []struct {
		Fulfillment        UnlockFulfillment
		NotStandardMessage string
	}{
		// nil fulfillment
		{NewFulfillment(nil), "nil fulfillment is never allowed"},
		{NewFulfillment(&NilFulfillment{}), "nil fulfillment is never allowed"},
		{&NilFulfillment{}, "nil fulfillment is never allowed"},
		// Single Signature
		{&SingleSignatureFulfillment{}, "nil single signature fulfillment is not allowed"},
		{
			&SingleSignatureFulfillment{
				PublicKey: PublicKey{
					Algorithm: SignatureAlgoEd25519,
				},
			},
			"nil pub-key + signature is not allowed for single-signature fulfilment",
		},
		{
			&SingleSignatureFulfillment{
				PublicKey: PublicKey{
					Algorithm: SignatureAlgoEd25519,
				},
				Signature: ByteSlice{
					1, 2, 3, 4, 5, 6, 7, 8,
					1, 2, 3, 4, 5, 6, 7, 8,
					1, 2, 3, 4, 5, 6, 7, 8,
					1, 2, 3, 4, 5, 6, 7, 8,
					1, 2, 3, 4, 5, 6, 7, 8,
					1, 2, 3, 4, 5, 6, 7, 8,
					1, 2, 3, 4, 5, 6, 7, 8,
					1, 2, 3, 4, 5, 6, 7, 8,
				},
			},
			"nil pub-key is not allowed for single-signature fulfilment",
		},
		{
			&SingleSignatureFulfillment{
				PublicKey: PublicKey{
					Algorithm: SignatureAlgoEd25519,
					Key: ByteSlice{
						1, 2, 3, 4, 5, 6, 7, 8,
						1, 2, 3, 4, 5, 6, 7, 8,
						1, 2, 3, 4, 5, 6, 7, 8,
						1, 2, 3, 4, 5, 6, 7, 8,
					},
				},
			},
			"nil signature is not allowed for single-signature fulfilment",
		},
		{
			&SingleSignatureFulfillment{
				PublicKey: PublicKey{
					Algorithm: SignatureAlgoEd25519,
					Key: ByteSlice{
						1, 2, 3, 4, 5, 6, 7, 8,
						1, 2, 3, 4, 5, 6, 7, 8,
						1, 2, 3, 4, 5, 6, 7, 8,
						1, 2, 3, 4, 5, 6, 7, 8,
					},
				},
				Signature: ByteSlice{
					1, 2, 3, 4, 5, 6, 7, 8,
					1, 2, 3, 4, 5, 6, 7, 8,
					1, 2, 3, 4, 5, 6, 7, 8,
					1, 2, 3, 4, 5, 6, 7, 8,
					1, 2, 3, 4, 5, 6, 7, 8,
					1, 2, 3, 4, 5, 6, 7, 8,
					1, 2, 3, 4, 5, 6, 7, 8,
					1, 2, 3, 4, 5, 6, 7, 8,
				},
			},
			"",
		},
		{
			&SingleSignatureFulfillment{
				PublicKey: PublicKey{
					Algorithm: SignatureAlgoEd25519,
					Key: ByteSlice{
						1, 2, 3, 4, 5, 6, 7, 8,
						1, 2, 3, 4, 5, 6, 7, 8,
						1, 2, 3, 4, 5, 6, 7, 8,
						1, 2, 3, 4, 5, 6, 7, 8,
					},
				},
				Signature: ByteSlice{1, 2, 3},
			},
			"wrong signature size",
		},
		{
			&SingleSignatureFulfillment{
				PublicKey: PublicKey{
					Algorithm: SignatureAlgoEd25519,
					Key:       ByteSlice{1, 2, 3},
				},
				Signature: ByteSlice{
					1, 2, 3, 4, 5, 6, 7, 8,
					1, 2, 3, 4, 5, 6, 7, 8,
					1, 2, 3, 4, 5, 6, 7, 8,
					1, 2, 3, 4, 5, 6, 7, 8,
					1, 2, 3, 4, 5, 6, 7, 8,
					1, 2, 3, 4, 5, 6, 7, 8,
					1, 2, 3, 4, 5, 6, 7, 8,
					1, 2, 3, 4, 5, 6, 7, 8,
				},
			},
			"wrong key size",
		},
		// Atomic Swap Fulfillment
		{&AtomicSwapFulfillment{}, "nil atomic swap fulfillment not allowed"},
		{
			&AtomicSwapFulfillment{
				PublicKey: PublicKey{
					Algorithm: SignatureAlgoEd25519,
					Key:       ByteSlice{1, 2, 3},
				},
				Signature: ByteSlice{
					1, 2, 3, 4, 5, 6, 7, 8,
					1, 2, 3, 4, 5, 6, 7, 8,
					1, 2, 3, 4, 5, 6, 7, 8,
					1, 2, 3, 4, 5, 6, 7, 8,
					1, 2, 3, 4, 5, 6, 7, 8,
					1, 2, 3, 4, 5, 6, 7, 8,
					1, 2, 3, 4, 5, 6, 7, 8,
					1, 2, 3, 4, 5, 6, 7, 8,
				},
			},
			"wrong pub key size",
		},
		{
			&AtomicSwapFulfillment{
				PublicKey: PublicKey{
					Algorithm: SignatureAlgoEd25519,
					Key: ByteSlice{
						1, 2, 3, 4, 5, 6, 7, 8,
						1, 2, 3, 4, 5, 6, 7, 8,
						1, 2, 3, 4, 5, 6, 7, 8,
						1, 2, 3, 4, 5, 6, 7, 8,
					},
				},
				Signature: ByteSlice{1, 2, 3},
			},
			"wrong signature size",
		},
		{
			&AtomicSwapFulfillment{
				PublicKey: PublicKey{
					Algorithm: SignatureAlgoEd25519,
					Key: ByteSlice{
						1, 2, 3, 4, 5, 6, 7, 8,
						1, 2, 3, 4, 5, 6, 7, 8,
						1, 2, 3, 4, 5, 6, 7, 8,
						1, 2, 3, 4, 5, 6, 7, 8,
					},
				},
				Signature: ByteSlice{
					1, 2, 3, 4, 5, 6, 7, 8,
					1, 2, 3, 4, 5, 6, 7, 8,
					1, 2, 3, 4, 5, 6, 7, 8,
					1, 2, 3, 4, 5, 6, 7, 8,
					1, 2, 3, 4, 5, 6, 7, 8,
					1, 2, 3, 4, 5, 6, 7, 8,
					1, 2, 3, 4, 5, 6, 7, 8,
					1, 2, 3, 4, 5, 6, 7, 8,
				},
			},
			"",
		},
		{
			&AtomicSwapFulfillment{
				PublicKey: PublicKey{
					Algorithm: SignatureAlgoEd25519,
					Key: ByteSlice{
						1, 2, 3, 4, 5, 6, 7, 8,
						1, 2, 3, 4, 5, 6, 7, 8,
						1, 2, 3, 4, 5, 6, 7, 8,
						1, 2, 3, 4, 5, 6, 7, 8,
					},
				},
				Signature: ByteSlice{
					1, 2, 3, 4, 5, 6, 7, 8,
					1, 2, 3, 4, 5, 6, 7, 8,
					1, 2, 3, 4, 5, 6, 7, 8,
					1, 2, 3, 4, 5, 6, 7, 8,
					1, 2, 3, 4, 5, 6, 7, 8,
					1, 2, 3, 4, 5, 6, 7, 8,
					1, 2, 3, 4, 5, 6, 7, 8,
					1, 2, 3, 4, 5, 6, 7, 8,
				},
				Secret: AtomicSwapSecret{1, 2, 3},
			},
			"",
		},
		// Legacy Atomic Swap Fulfillment
		{&LegacyAtomicSwapFulfillment{}, "nil legacy atomic swap fulfillment not allowed"},
		{
			&LegacyAtomicSwapFulfillment{
				Sender:       UnlockHash{},
				Receiver:     unlockHashFromHex("015fe50b9c596d8717e5e7ba79d5a7c9c8b82b1427a04d5c0771268197c90e99dccbcdf0ba9c90"),
				HashedSecret: AtomicSwapHashedSecret{4, 5, 6},
				TimeLock:     TestnetChainConstants().GenesisTimestamp,
				PublicKey: PublicKey{
					Algorithm: SignatureAlgoEd25519,
					Key: ByteSlice{
						1, 2, 3, 4, 5, 6, 7, 8,
						1, 2, 3, 4, 5, 6, 7, 8,
						1, 2, 3, 4, 5, 6, 7, 8,
						1, 2, 3, 4, 5, 6, 7, 8,
					},
				},
				Signature: ByteSlice{
					1, 2, 3, 4, 5, 6, 7, 8,
					1, 2, 3, 4, 5, 6, 7, 8,
					1, 2, 3, 4, 5, 6, 7, 8,
					1, 2, 3, 4, 5, 6, 7, 8,
					1, 2, 3, 4, 5, 6, 7, 8,
					1, 2, 3, 4, 5, 6, 7, 8,
					1, 2, 3, 4, 5, 6, 7, 8,
					1, 2, 3, 4, 5, 6, 7, 8,
				},
			},
			"nil sender",
		},
		{
			&LegacyAtomicSwapFulfillment{
				Sender:       unlockHashFromHex("01746677df456546d93729066dd88514e2009930f3eebac3c93d43c88a108f8f9aa9e7c6f58893"),
				Receiver:     UnlockHash{},
				HashedSecret: AtomicSwapHashedSecret{4, 5, 6},
				TimeLock:     TestnetChainConstants().GenesisTimestamp,
				PublicKey: PublicKey{
					Algorithm: SignatureAlgoEd25519,
					Key: ByteSlice{
						1, 2, 3, 4, 5, 6, 7, 8,
						1, 2, 3, 4, 5, 6, 7, 8,
						1, 2, 3, 4, 5, 6, 7, 8,
						1, 2, 3, 4, 5, 6, 7, 8,
					},
				},
				Signature: ByteSlice{
					1, 2, 3, 4, 5, 6, 7, 8,
					1, 2, 3, 4, 5, 6, 7, 8,
					1, 2, 3, 4, 5, 6, 7, 8,
					1, 2, 3, 4, 5, 6, 7, 8,
					1, 2, 3, 4, 5, 6, 7, 8,
					1, 2, 3, 4, 5, 6, 7, 8,
					1, 2, 3, 4, 5, 6, 7, 8,
					1, 2, 3, 4, 5, 6, 7, 8,
				},
			},
			"nil receiver",
		},
		{
			&LegacyAtomicSwapFulfillment{
				Sender:       unlockHashFromHex("01746677df456546d93729066dd88514e2009930f3eebac3c93d43c88a108f8f9aa9e7c6f58893"),
				Receiver:     unlockHashFromHex("015fe50b9c596d8717e5e7ba79d5a7c9c8b82b1427a04d5c0771268197c90e99dccbcdf0ba9c90"),
				HashedSecret: AtomicSwapHashedSecret{},
				TimeLock:     TestnetChainConstants().GenesisTimestamp,
				PublicKey: PublicKey{
					Algorithm: SignatureAlgoEd25519,
					Key: ByteSlice{
						1, 2, 3, 4, 5, 6, 7, 8,
						1, 2, 3, 4, 5, 6, 7, 8,
						1, 2, 3, 4, 5, 6, 7, 8,
						1, 2, 3, 4, 5, 6, 7, 8,
					},
				},
				Signature: ByteSlice{
					1, 2, 3, 4, 5, 6, 7, 8,
					1, 2, 3, 4, 5, 6, 7, 8,
					1, 2, 3, 4, 5, 6, 7, 8,
					1, 2, 3, 4, 5, 6, 7, 8,
					1, 2, 3, 4, 5, 6, 7, 8,
					1, 2, 3, 4, 5, 6, 7, 8,
					1, 2, 3, 4, 5, 6, 7, 8,
					1, 2, 3, 4, 5, 6, 7, 8,
				},
			},
			"nil hashed secret",
		},
		{
			&LegacyAtomicSwapFulfillment{
				Sender:       unlockHashFromHex("01746677df456546d93729066dd88514e2009930f3eebac3c93d43c88a108f8f9aa9e7c6f58893"),
				Receiver:     unlockHashFromHex("015fe50b9c596d8717e5e7ba79d5a7c9c8b82b1427a04d5c0771268197c90e99dccbcdf0ba9c90"),
				HashedSecret: AtomicSwapHashedSecret{4, 5, 6},
				TimeLock:     TestnetChainConstants().GenesisTimestamp,
				PublicKey: PublicKey{
					Algorithm: SignatureAlgoEd25519,
					Key:       ByteSlice{1, 2, 3},
				},
				Signature: ByteSlice{
					1, 2, 3, 4, 5, 6, 7, 8,
					1, 2, 3, 4, 5, 6, 7, 8,
					1, 2, 3, 4, 5, 6, 7, 8,
					1, 2, 3, 4, 5, 6, 7, 8,
					1, 2, 3, 4, 5, 6, 7, 8,
					1, 2, 3, 4, 5, 6, 7, 8,
					1, 2, 3, 4, 5, 6, 7, 8,
					1, 2, 3, 4, 5, 6, 7, 8,
				},
			},
			"wrong pub key size",
		},
		{
			&LegacyAtomicSwapFulfillment{
				Sender:       unlockHashFromHex("01746677df456546d93729066dd88514e2009930f3eebac3c93d43c88a108f8f9aa9e7c6f58893"),
				Receiver:     unlockHashFromHex("015fe50b9c596d8717e5e7ba79d5a7c9c8b82b1427a04d5c0771268197c90e99dccbcdf0ba9c90"),
				HashedSecret: AtomicSwapHashedSecret{4, 5, 6},
				TimeLock:     TestnetChainConstants().GenesisTimestamp,
				PublicKey: PublicKey{
					Algorithm: SignatureAlgoEd25519,
					Key: ByteSlice{
						1, 2, 3, 4, 5, 6, 7, 8,
						1, 2, 3, 4, 5, 6, 7, 8,
						1, 2, 3, 4, 5, 6, 7, 8,
						1, 2, 3, 4, 5, 6, 7, 8,
					},
				},
				Signature: ByteSlice{1, 2, 3},
			},
			"wrong signature size",
		},
		{
			&LegacyAtomicSwapFulfillment{
				Sender:       unlockHashFromHex("01746677df456546d93729066dd88514e2009930f3eebac3c93d43c88a108f8f9aa9e7c6f58893"),
				Receiver:     unlockHashFromHex("015fe50b9c596d8717e5e7ba79d5a7c9c8b82b1427a04d5c0771268197c90e99dccbcdf0ba9c90"),
				HashedSecret: AtomicSwapHashedSecret{4, 5, 6},
				TimeLock:     TestnetChainConstants().GenesisTimestamp,
				PublicKey: PublicKey{
					Algorithm: SignatureAlgoEd25519,
					Key: ByteSlice{
						1, 2, 3, 4, 5, 6, 7, 8,
						1, 2, 3, 4, 5, 6, 7, 8,
						1, 2, 3, 4, 5, 6, 7, 8,
						1, 2, 3, 4, 5, 6, 7, 8,
					},
				},
				Signature: ByteSlice{
					1, 2, 3, 4, 5, 6, 7, 8,
					1, 2, 3, 4, 5, 6, 7, 8,
					1, 2, 3, 4, 5, 6, 7, 8,
					1, 2, 3, 4, 5, 6, 7, 8,
					1, 2, 3, 4, 5, 6, 7, 8,
					1, 2, 3, 4, 5, 6, 7, 8,
					1, 2, 3, 4, 5, 6, 7, 8,
					1, 2, 3, 4, 5, 6, 7, 8,
				},
			},
			"",
		},
		{
			&LegacyAtomicSwapFulfillment{
				Sender:       unlockHashFromHex("01746677df456546d93729066dd88514e2009930f3eebac3c93d43c88a108f8f9aa9e7c6f58893"),
				Receiver:     unlockHashFromHex("015fe50b9c596d8717e5e7ba79d5a7c9c8b82b1427a04d5c0771268197c90e99dccbcdf0ba9c90"),
				HashedSecret: AtomicSwapHashedSecret{4, 5, 6},
				TimeLock:     TestnetChainConstants().GenesisTimestamp,
				PublicKey: PublicKey{
					Algorithm: SignatureAlgoEd25519,
					Key: ByteSlice{
						1, 2, 3, 4, 5, 6, 7, 8,
						1, 2, 3, 4, 5, 6, 7, 8,
						1, 2, 3, 4, 5, 6, 7, 8,
						1, 2, 3, 4, 5, 6, 7, 8,
					},
				},
				Signature: ByteSlice{
					1, 2, 3, 4, 5, 6, 7, 8,
					1, 2, 3, 4, 5, 6, 7, 8,
					1, 2, 3, 4, 5, 6, 7, 8,
					1, 2, 3, 4, 5, 6, 7, 8,
					1, 2, 3, 4, 5, 6, 7, 8,
					1, 2, 3, 4, 5, 6, 7, 8,
					1, 2, 3, 4, 5, 6, 7, 8,
					1, 2, 3, 4, 5, 6, 7, 8,
				},
				Secret: AtomicSwapSecret{1, 2, 3},
			},
			"",
		},
		{
			new(MultiSignatureFulfillment),
			"no pairs given",
		},
		{
			&MultiSignatureFulfillment{
				Pairs: []PublicKeySignaturePair{
					{
						PublicKey: PublicKey{
							Algorithm: SignatureAlgoEd25519,
							Key: ByteSlice{
								1, 2, 3, 4, 5, 6, 7, 8,
								1, 2, 3, 4, 5, 6, 7, 8,
								1, 2, 3, 4, 5, 6, 7, 8,
								1, 2, 3, 4, 5, 6, 7, 8,
							},
						},
						Signature: ByteSlice{
							1, 2, 3, 4, 5, 6, 7, 8,
							1, 2, 3, 4, 5, 6, 7, 8,
							1, 2, 3, 4, 5, 6, 7, 8,
							1, 2, 3, 4, 5, 6, 7, 8,
							1, 2, 3, 4, 5, 6, 7, 8,
							1, 2, 3, 4, 5, 6, 7, 8,
							1, 2, 3, 4, 5, 6, 7, 8,
							1, 2, 3, 4, 5, 6, 7, 8,
						},
					},
					{
						PublicKey: PublicKey{
							Algorithm: SignatureAlgoEd25519,
							Key: ByteSlice{
								1, 2, 3, 4, 5, 6, 7, 8,
								1, 2, 3, 4, 5, 6, 7, 8,
								1, 2, 3, 4, 5, 6, 7, 8,
								1, 2, 3, 4, 5, 6, 7, 8,
							},
						},
						Signature: ByteSlice{
							1, 2, 3, 4, 5, 6, 7, 8,
							1, 2, 3, 4, 5, 6, 7, 8,
							1, 2, 3, 4, 5, 6, 7, 8,
							1, 2, 3, 4, 5, 6, 7, 8,
							1, 2, 3, 4, 5, 6, 7, 8,
							1, 2, 3, 4, 5, 6, 7, 8,
							1, 2, 3, 4, 5, 6, 7, 8,
							1, 2, 3, 4, 5, 6, 7, 8,
						},
					},
				},
			},
			"",
		},
		{
			&MultiSignatureFulfillment{
				Pairs: []PublicKeySignaturePair{
					{
						PublicKey: PublicKey{
							Algorithm: SignatureAlgoEd25519,
							Key: ByteSlice{
								1, 2, 3, 4, 5, 6, 7, 8,
								1, 2, 3, 4, 5, 6, 7, 8,
								1, 2, 3, 4, 5, 6, 7, 8,
								1, 2, 3, 4, 5, 6, 7, 8,
							},
						},
						Signature: ByteSlice{ // only 56 bytes
							1, 2, 3, 4, 5, 6, 7, 8,
							1, 2, 3, 4, 5, 6, 7, 8,
							1, 2, 3, 4, 5, 6, 7, 8,
							1, 2, 3, 4, 5, 6, 7, 8,
							1, 2, 3, 4, 5, 6, 7, 8,
							1, 2, 3, 4, 5, 6, 7, 8,
							1, 2, 3, 4, 5, 6, 7, 8,
						},
					},
					{
						PublicKey: PublicKey{
							Algorithm: SignatureAlgoEd25519,
							Key: ByteSlice{
								1, 2, 3, 4, 5, 6, 7, 8,
								1, 2, 3, 4, 5, 6, 7, 8,
								1, 2, 3, 4, 5, 6, 7, 8,
								1, 2, 3, 4, 5, 6, 7, 8,
							},
						},
						Signature: ByteSlice{
							1, 2, 3, 4, 5, 6, 7, 8,
							1, 2, 3, 4, 5, 6, 7, 8,
							1, 2, 3, 4, 5, 6, 7, 8,
							1, 2, 3, 4, 5, 6, 7, 8,
							1, 2, 3, 4, 5, 6, 7, 8,
							1, 2, 3, 4, 5, 6, 7, 8,
							1, 2, 3, 4, 5, 6, 7, 8,
							1, 2, 3, 4, 5, 6, 7, 8,
						},
					},
				},
			},
			"Invalid signature",
		},
		{
			&MultiSignatureFulfillment{
				Pairs: []PublicKeySignaturePair{
					{
						PublicKey: PublicKey{
							Algorithm: SignatureAlgoEd25519,
							Key: ByteSlice{
								1, 2, 3, 4, 5, 6, 7, 8,
								1, 2, 3, 4, 5, 6, 7, 8,
								1, 2, 3, 4, 5, 6, 7, 8,
								1, 2, 3, 4, 5, 6, 7, 8,
							},
						},
						Signature: ByteSlice{
							1, 2, 3, 4, 5, 6, 7, 8,
							1, 2, 3, 4, 5, 6, 7, 8,
							1, 2, 3, 4, 5, 6, 7, 8,
							1, 2, 3, 4, 5, 6, 7, 8,
							1, 2, 3, 4, 5, 6, 7, 8,
							1, 2, 3, 4, 5, 6, 7, 8,
							1, 2, 3, 4, 5, 6, 7, 8,
							1, 2, 3, 4, 5, 6, 7, 8,
						},
					},
					{
						PublicKey: PublicKey{
							Algorithm: SignatureAlgoEd25519,
							Key: ByteSlice{
								1, 2, 3, 4, 5, 6, 7, 8,
								1, 2, 3, 4, 5, 6, 7, 8,
								1, 2, 3, 4, 5, 6, 7, 8,
								1, 2, 3, 4, 5, 6, 7, 8,
							},
						},
					},
				},
			},
			"Empty signature",
		},
	}
	for idx, testCase := range testCases {
		err := testCase.Fulfillment.IsStandardFulfillment(ValidationContext{})
		if testCase.NotStandardMessage != "" {
			if err == nil {
				t.Error(idx, "expected error, but none received:", testCase.NotStandardMessage, testCase.Fulfillment)
			}
		} else {
			if err != nil {
				t.Error(idx, "expected no error but received one:", err, testCase.Fulfillment)
			}
		}
	}
}

func TestAtomicSwapHashedSecretStringify(t *testing.T) {
	hexASHS := "000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f"
	var hs AtomicSwapHashedSecret
	err := hs.LoadString(hexASHS)
	if err != nil {
		t.Error("failed to load hashed secret in hash format:", err)
	}
	if bytes.Compare([]byte{
		0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
		16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31}, hs[:]) != 0 {
		t.Error("unexpected loaded hashed secret:", hs)
	}

	str := hs.String()
	if str != hexASHS {
		t.Error(str, "!=", hexASHS)
	}

	for i := 0; i < 64; i++ {
		_, err := rand.Read(hs[:])
		if err != nil {
			t.Error(i, err)
		}

		str := hs.String()
		if str == "" {
			t.Error(i, "hex-encoded string is empty")
		}

		var hs2 AtomicSwapHashedSecret
		err = hs2.LoadString(str)
		if err != nil {
			t.Error(i, "failed to load hashed secret in hash format:", err)
		}

		if bytes.Compare(hs2[:], hs[:]) != 0 {
			t.Error(i, "unexpected loaded hashed secret:", hs2, "!=", hs)
		}
	}
}

func TestAtomicSwapHashedSecretJSON(t *testing.T) {
	hexASHS := `"000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f"`
	var hs AtomicSwapHashedSecret
	err := hs.UnmarshalJSON([]byte(hexASHS))
	if err != nil {
		t.Error("failed to load hashed secret in hash format:", err)
	}
	if bytes.Compare([]byte{
		0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
		16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31}, hs[:]) != 0 {
		t.Error("unexpected loaded hashed secret:", hs)
	}

	b, err := hs.MarshalJSON()
	if err != nil {
		t.Error(err, b)
	}
	str := string(b)
	if str != hexASHS {
		t.Error(str, "!=", hexASHS)
	}

	for i := 0; i < 64; i++ {
		_, err := rand.Read(hs[:])
		if err != nil {
			t.Error(i, err)
		}

		b, err := hs.MarshalJSON()
		if err != nil || len(b) == 0 {
			t.Error(i, "hex-encoded string is empty or an error occurred", err, b)
		}

		var hs2 AtomicSwapHashedSecret
		err = hs2.UnmarshalJSON(b)
		if err != nil {
			t.Error(i, "failed to load hashed secret in hex format:", err)
		}

		if bytes.Compare(hs2[:], hs[:]) != 0 {
			t.Error(i, "unexpected loaded hashed secret:", hs2, "!=", hs)
		}
	}
}

func TestAtomicSwapSecretStringify(t *testing.T) {
	hexASS := "000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f"
	var s AtomicSwapSecret
	err := s.LoadString(hexASS)
	if err != nil {
		t.Error("failed to load secret in hex format:", err)
	}
	if bytes.Compare([]byte{
		0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
		16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31}, s[:]) != 0 {
		t.Error("unexpected loaded secret:", s)
	}

	str := s.String()
	if str != hexASS {
		t.Error(str, "!=", hexASS)
	}

	for i := 0; i < 64; i++ {
		_, err := rand.Read(s[:])
		if err != nil {
			t.Error(i, err)
		}

		str := s.String()
		if str == "" {
			t.Error(i, "hex-encoded string is empty")
		}

		var s2 AtomicSwapSecret
		err = s2.LoadString(str)
		if err != nil {
			t.Error(i, "failed to load secret in hex format:", err)
		}

		if bytes.Compare(s2[:], s[:]) != 0 {
			t.Error(i, "unexpected loaded secret:", s2, "!=", s)
		}
	}
}

func TestAtomicSwapSecretJSON(t *testing.T) {
	hexASS := `"000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f"`
	var s AtomicSwapSecret
	err := s.UnmarshalJSON([]byte(hexASS))
	if err != nil {
		t.Error("failed to load secret in hash format:", err)
	}
	if bytes.Compare([]byte{
		0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
		16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31}, s[:]) != 0 {
		t.Error("unexpected loaded secret:", s)
	}

	b, err := s.MarshalJSON()
	if err != nil {
		t.Error(err, b)
	}
	str := string(b)
	if str != hexASS {
		t.Error(str, "!=", hexASS)
	}

	for i := 0; i < 64; i++ {
		_, err := rand.Read(s[:])
		if err != nil {
			t.Error(i, err)
		}

		b, err := s.MarshalJSON()
		if err != nil || len(b) == 0 {
			t.Error(i, "hex-encoded string is empty or an error occurred", err, b)
		}

		var s2 AtomicSwapHashedSecret
		err = s2.UnmarshalJSON(b)
		if err != nil {
			t.Error(i, "failed to load secret in hex format:", err)
		}

		if bytes.Compare(s2[:], s[:]) != 0 {
			t.Error(i, "unexpected loaded secret:", s2, "!=", s)
		}
	}
}

func TestUnlockConditionFulfillable(t *testing.T) {
	testCases := []struct {
		Condition   UnlockCondition
		Context     FulfillableContext
		Fulfillable bool
	}{
		{UnlockConditionProxy{}, FulfillableContext{}, true},
		{UnlockConditionProxy{&NilCondition{}}, FulfillableContext{}, true},
		{&NilCondition{}, FulfillableContext{}, true},
		{&UnlockHashCondition{}, FulfillableContext{}, true},
		{&AtomicSwapCondition{}, FulfillableContext{}, true},
		{&TimeLockCondition{}, FulfillableContext{}, true},
		{
			&TimeLockCondition{LockTime: 1},
			FulfillableContext{},
			false,
		},
		{
			&TimeLockCondition{LockTime: 1},
			FulfillableContext{BlockHeight: 1},
			true,
		},
		{
			&TimeLockCondition{},
			FulfillableContext{BlockHeight: 1},
			true,
		},
		{
			&TimeLockCondition{LockTime: 500000001},
			FulfillableContext{BlockHeight: 500000002, BlockTime: 50000000},
			false,
		},
		{
			&TimeLockCondition{LockTime: 500000001},
			FulfillableContext{BlockHeight: 500000002, BlockTime: 500000001},
			true,
		},
		{
			&TimeLockCondition{LockTime: 500000003},
			FulfillableContext{BlockHeight: 500000002, BlockTime: 500000003},
			true,
		},
		{
			&TimeLockCondition{LockTime: 500000004},
			FulfillableContext{BlockHeight: 500000002, BlockTime: 500000005},
			true,
		},
		{
			&MultiSignatureCondition{},
			FulfillableContext{},
			true,
		},
	}
	for idx, testCase := range testCases {
		fulfillable := testCase.Condition.Fulfillable(testCase.Context)
		if fulfillable != testCase.Fulfillable {
			t.Error(idx, fulfillable, "!=", testCase.Fulfillable)
		}
	}
}

// Test to ensure that multisig wallet addresses are different
// depending upon all properties, including the MinimumSignatureCount property,
// but are the same even if the address order is different.
func TestMultisigAddresses(t *testing.T) {
	// create a multisig condition and take the unlock hash
	msc := NewMultiSignatureCondition(UnlockHashSlice{
		unlockHashFromHex("015fe50b9c596d8717e5e7ba79d5a7c9c8b82b1427a04d5c0771268197c90e99dccbcdf0ba9c90"),
		unlockHashFromHex("01d3a8d366864f5f368bd73959139c55da5f1f8beaa07cb43519cc87d2a51135ae0b3ba93cf2d9"),
		unlockHashFromHex("01437c56286c76dec14e87f5da5e5a436651006e6cd46bee5865c9060ba178f7296ed843b70a57"),
	}, 3)
	uh1 := msc.UnlockHash()

	// change the count property of it, and take another unlock hash,
	// the new unlockhash should be different from the previous one
	msc.MinimumSignatureCount = 2
	uh2 := msc.UnlockHash()
	if uh1.Cmp(uh2) == 0 {
		t.Fatal("expected uh1 to be not equal to uh2, but the ywere both: ", uh1.String())
	}

	// switch two unlock hashes in the slice of the multisig condition,
	// this should remain the unlock hash generated from it identical to the previous one
	var uh UnlockHash
	uh, msc.UnlockHashes[0] = msc.UnlockHashes[0], msc.UnlockHashes[2]
	msc.UnlockHashes[2] = uh
	uh3 := msc.UnlockHash()
	if uh2.Cmp(uh3) != 0 {
		t.Fatalf("expected uh2 (%s) to be equal to uh3 (%s), but the they weren't", uh2.String(), uh3.String())
	}
}
