package cli

import (
	"testing"
	"time"
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

const testTimeNow = 1525600388

func init() {
	computeTimeNow = func() time.Time {
		return time.Unix(testTimeNow, 0)
	}
}
