package cli

import (
	"testing"
	"time"

	"github.com/rivine/rivine/types"
)

// TestDateOnlyLayout tests that our custom DateOnly timestamp Layout
// works as expected.
func TestDateOnlyLayout(t *testing.T) {
	testCases := []struct {
		Raw          string
		Day          int
		Month        time.Month
		Year         int
		TimeZoneName string
	}{
		{"01/01/2000 GMT-1", 1, time.January, 2000, "GMT-1"},
		{"11/11/1111 UTC", 11, time.November, 1111, "UTC"},
		{"02/01/2003 GMT+2", 2, time.January, 2003, "GMT+2"},
		{"18/06/2017 PET", 18, time.June, 2017, "PET"},
	}
	for idx, testCase := range testCases {
		ts, err := time.Parse(DateOnlyLayout, testCase.Raw)
		if err != nil {
			t.Error(idx, testCase.Raw, err)
			continue
		}
		if name, _ := ts.Zone(); name != testCase.TimeZoneName {
			t.Error(idx, "timezone", name, "!=", testCase.TimeZoneName)
		}
		before := ts.Unix()
		ts = ts.UTC()
		after := ts.Unix()
		if before != after {
			t.Error(idx, "inconsistent unix epoch:", before, "!=", after)
		}
		if day := ts.Day(); day != testCase.Day {
			t.Error(idx, "day", day, "!=", testCase.Day)
		}
		if month := ts.Month(); month != testCase.Month {
			t.Error(idx, "month", month, "!=", testCase.Month)
		}
		if year := ts.Year(); year != testCase.Year {
			t.Error(idx, "year", year, "!=", testCase.Year)
		}
		timeOnly := ts.Hour() + ts.Minute() + ts.Second()
		if timeOnly != 0 {
			t.Error(idx, "time isn't nil: ", ts.Hour(), ts.Minute(), ts.Second())
		}
	}
}

var lockTimeFlagTestCases = []struct {
	Raw      string
	LockTime uint64
}{
	{"03/08/2018 UTC", 1533254400},
	{"20/08/2018 UTC", 1534723200},
	{"03 Aug 18 12:00 UTC", 1533297600},
	{"20 Sep 18 13:00 GMT+2", 1537448400},
	{"48h", testTimeNow + (48 * 3600)},
	{"+24h", testTimeNow + (24 * 3600)},
	{"-12h", testTimeNow - (12 * 3600)},
	{"12h30m", testTimeNow + (12 * 3600) + (30 * 60)},
	{"42", 42},
	{"25000", 25000},
	{"1525599730", 1525599730},
	{"1533254400", 1533254400},
}

func TestLockTimeFlagSet(t *testing.T) {
	for idx, testCase := range lockTimeFlagTestCases {
		var ltf LockTimeFlag
		err := ltf.Set(testCase.Raw)
		if err != nil {
			t.Error(idx, err)
			continue
		}
		if lt := ltf.LockTime(); lt != testCase.LockTime {
			t.Error(idx, lt, "!=", testCase.LockTime)
		}
	}
}

func TestLockTimeSetStringLoop(t *testing.T) {
	for idx, testCase := range lockTimeFlagTestCases {
		var ltf LockTimeFlag
		err := ltf.Set(testCase.Raw)
		if err != nil {
			t.Error(idx, err)
			continue
		}
		if raw := ltf.String(); raw != testCase.Raw {
			t.Error(idx, raw, "!=", testCase.Raw)
		}
	}
}

func TestStringLoaderFlag(t *testing.T) {
	// test loader->string->loader->string
	stringLoaders := []StringLoader{
		&types.CoinOutputID{},
		&types.CoinOutputID{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0xA, 0xB, 0xC, 0xD, 0xE, 0xF, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0xA, 0xB, 0xC, 0xD, 0xE, 0xF},
		&types.TransactionID{},
		&types.TransactionID{1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2},
	}
	for idx, stringLoader := range stringLoaders {
		str := stringLoader.String()
		err := stringLoader.LoadString(str)
		if err != nil {
			t.Errorf("error while loading string for loader #%d: %v", idx, err)
		}
		str2 := stringLoader.String()
		if str != str2 {
			t.Errorf("loader #%d isn't deterministic: %s != %s", idx, str, str2)
		}
	}

	// test string->loader->string
	testCases := []string{
		"0112210f9efa5441ab705226b0628679ed190eb4588b662991747ea3809d93932c7b41cbe4b732",
		"01450aeb140c58012cb4afb48e068f976272fefa44ffe0991a8a4350a3687558d66c8fc753c37e",
		"01e56d03c7818179c1d21ab1fe99be91ec7fa48a21ca1b0818ad55e0b241d55067e740e11c08f5",
	}
	for idx, testCase := range testCases {
		var uh types.UnlockHash
		err := uh.LoadString(testCase)
		if err != nil {
			t.Errorf("error while loading string for unlockhash #%d: %v", idx, err)
		}
		str := uh.String()
		if testCase != str {
			t.Errorf("unlockhash #%d string loading isn't deterministic: %s != %s", idx, testCase, str)
		}
	}
}

func TestEncodingTypeFlagString(t *testing.T) {
	testCases := []struct {
		f EncodingTypeFlag
		s string
	}{
		{EncodingTypeFlag{et: func() *EncodingType { et := EncodingTypeHuman; return &et }()}, "human"},
		{EncodingTypeFlag{et: func() *EncodingType { et := EncodingTypeJSON; return &et }()}, "json"},
		{EncodingTypeFlag{et: func() *EncodingType { et := EncodingTypeHex; return &et }()}, "hex"},
		{EncodingTypeFlag{et: func() *EncodingType { et := EncodingTypeJSON | EncodingTypeHex; return &et }()}, "human"}, // def = human
		{EncodingTypeFlag{et: func() *EncodingType { et := EncodingType(128); return &et }()}, "human"},                  // def = human
	}
	for idx, testCase := range testCases {
		str := testCase.f.String()
		if str != testCase.s {
			t.Error(idx, "unexpected result", str, "!=", testCase.s)
		}
	}
}

func TestNewEncodingTypeFlag(t *testing.T) {
	var et EncodingType
	testPanic(t, "no reference given", func() {
		NewEncodingTypeFlag(0, nil, 0)
	})
	testPanic(t, "no human encoding is allowed", func() {
		NewEncodingTypeFlag(EncodingTypeHuman, &et, EncodingTypeJSON|EncodingTypeHex)
	})
	testPanic(t, "no hex encoding is allowed", func() {
		NewEncodingTypeFlag(EncodingTypeHex, &et, EncodingTypeJSON|EncodingTypeHuman)
	})
	testPanic(t, "no json encoding is allowed", func() {
		NewEncodingTypeFlag(EncodingTypeJSON, &et, EncodingTypeHex|EncodingTypeHuman)
	})

	NewEncodingTypeFlag(EncodingTypeHex, &et, 0)
	if et != EncodingTypeHex {
		t.Error("expected et to be EncodingHEX, but it was instead: ", et)
	}
}

func TestEncodingTypeFlagSet(t *testing.T) {
	// create new encoding type and flag
	var et EncodingType
	f := NewEncodingTypeFlag(0, &et, 0) // mask=0: means all is allowed
	// test if we can set JSON
	err := f.Set("json")
	if err != nil {
		t.Fatal(err)
	}
	if et != EncodingTypeJSON {
		t.Fatal("et was supposed to be EncodingTypeJSON, but was instead:", et)
	}
	// test if we can set JSON using all caps
	err = f.Set("JSON")
	if err != nil {
		t.Fatal(err)
	}
	if et != EncodingTypeJSON {
		t.Fatal("et was supposed to be EncodingTypeJSON, but was instead:", et)
	}
	// test if we can set hex
	err = f.Set("hex")
	if err != nil {
		t.Fatal(err)
	}
	if et != EncodingTypeHex {
		t.Fatal("et was supposed to be EncodingTypeHex, but was instead:", et)
	}
	// test if we can set human explicitly
	err = f.Set("human")
	if err != nil {
		t.Fatal(err)
	}
	if et != EncodingTypeHuman {
		t.Fatal("et was supposed to be EncodingTypeHuman, but was instead:", et)
	}
	// test if we can set a flag in a case insensitive manner
	err = f.Set("HeX")
	if err != nil {
		t.Fatal(err)
	}
	if et != EncodingTypeHex {
		t.Fatal("et was supposed to be EncodingTypeHex, but was instead:", et)
	}
	// set mask to 0, as to allow nothing
	f.mask = 0
	// nothing should be allowed now
	err = f.Set("human")
	if err == nil {
		t.Fatal("setting to human should fail, given that nothing is allowed, but now et is: ", et)
	}
	err = f.Set("json")
	if err == nil {
		t.Fatal("setting to json should fail, given that nothing is allowed, but now et is: ", et)
	}
	err = f.Set("hex")
	if err == nil {
		t.Fatal("setting to hex should fail, given that nothing is allowed, but now et is: ", et)
	}
	// set mask to Human, as to allow only Human
	f.mask = EncodingTypeHuman
	err = f.Set("human")
	if err != nil {
		t.Fatal(err)
	}
	if et != EncodingTypeHuman {
		t.Fatal("et was supposed to be EncodingTypeHuman, but was instead:", et)
	}
	err = f.Set("json")
	if err == nil {
		t.Fatal("setting to json should fail, given that nothing is allowed, but now et is: ", et)
	}
	err = f.Set("hex")
	if err == nil {
		t.Fatal("setting to hex should fail, given that nothing is allowed, but now et is: ", et)
	}
}

func TestInvalidEncodingTypeSetAsFlag(t *testing.T) {
	// create new encoding type and flag
	var et EncodingType
	f := NewEncodingTypeFlag(0, &et, 0) // mask=0: means all is allowed

	// test some invalid options, shoud all fail
	err := f.Set("42")
	if err == nil {
		t.Fatal("setting 42 as a flag was supposed to result in an error, but now et is: ", et)
	}
	if et != EncodingTypeHuman {
		t.Fatal("et was supposed to still be EncodingTypeHuman, but was instead:", et)
	}
	err = f.Set("he")
	if err == nil {
		t.Fatal("setting he as a flag was supposed to result in an error, but now et is: ", et)
	}
	if et != EncodingTypeHuman {
		t.Fatal("et was supposed to still be EncodingTypeHuman, but was instead:", et)
	}
	err = f.Set("yaml")
	if err == nil {
		t.Fatal("setting yaml as a flag was supposed to result in an error, but now et is: ", et)
	}
	if et != EncodingTypeHuman {
		t.Fatal("et was supposed to still be EncodingTypeHuman, but was instead:", et)
	}
	err = f.Set("jon")
	if err == nil {
		t.Fatal("setting jon as a flag was supposed to result in an error, but now et is: ", et)
	}
	if et != EncodingTypeHuman {
		t.Fatal("et was supposed to still be EncodingTypeHuman, but was instead:", et)
	}
}

func testPanic(t *testing.T, label string, f func()) {
	defer func() {
		e := recover()
		if e == nil {
			t.Error(label + ": expected a panic, but received none")
		}
	}()
	f()
}

const testTimeNow = 1525600388

func init() {
	computeTimeNow = func() time.Time {
		return time.Unix(testTimeNow, 0)
	}
}
