package stormdb

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/threefoldtech/rivine/types"
)

func TestGetTimestampBucketAndOffset(t *testing.T) {
	testCases := []struct {
		Input               types.Timestamp
		ExpectedBucketID    StormTimeBucketID
		ExpectedBucktOffset StormTimeBucketOffset
	}{
		{0, 0, 0},
		{239, 0, 239},
		{240, 1, 0},
		{42, 0, 42},
		{282, 1, 42},
		{1573027885, 6554282, 205},
	}
	for idx, testCase := range testCases {
		testCase := testCase
		t.Run(fmt.Sprintf("test #%d", idx+1), func(t *testing.T) {
			bucketID, bucketOffset := GetTimestampBucketAndOffset(testCase.Input)
			if bucketID != testCase.ExpectedBucketID {
				t.Errorf("unexpected bucketID: expected %d, not %d", testCase.ExpectedBucketID, bucketID)
			}
			if bucketOffset != testCase.ExpectedBucktOffset {
				t.Errorf("unexpected buckt offset: expected %d, not %d", testCase.ExpectedBucktOffset, bucketOffset)
			}
		})
	}
}

func TestGetTimestampBucketIdentifiersForTimestampRange(t *testing.T) {
	type sids []StormTimeBucketID
	testCases := []struct {
		StartExclusive  types.Timestamp
		EndInclusive    types.Timestamp
		ExpectedBuckets sids
	}{
		{0, 1, sids{0}},
		{0, 240, sids{0, 1}},
		{238, 240, sids{0, 1}},
		{238, 479, sids{0, 1}},
		{1573028665, 1573028825, sids{6554286}},
	}
	for idx, testCase := range testCases {
		testCase := testCase
		t.Run(fmt.Sprintf("test #%d", idx+1), func(t *testing.T) {
			buckets := GetTimestampBucketIdentifiersForTimestampRange(testCase.StartExclusive, testCase.EndInclusive)
			if !reflect.DeepEqual(buckets, []StormTimeBucketID(testCase.ExpectedBuckets)) {
				t.Errorf("unexpected buckets: expected %d, not %d", testCase.ExpectedBuckets, buckets)
			}
		})
	}
}
