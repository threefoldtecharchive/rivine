package client

import (
	"testing"
	"time"

	"github.com/threefoldtech/rivine/types"
)

// TestEstimatedHeightAt tests that the expectedHeightAt function correctly
// estimates the blockheight (and rounds to the nearest block).
func TestEstimatedHeightBetween(t *testing.T) {
	tests := []struct {
		From, To       int64 // timestamps in seconds
		BlockFrequency int64 // duration in seconds
		ExpectedHeight types.BlockHeight
	}{
		// 0 or negative
		{
			0, 0,
			1, 0,
		},
		{
			10, 10,
			1, 0,
		},
		{
			10, 5,
			1, 0,
		},
		// 1 (productive) hour
		{
			0, 3600,
			1, 3600,
		},
		{
			0, 3600,
			60, 60,
		},
		{
			0, 3600,
			120, 30,
		},
		// an example with more realistic numbers
		{
			// 7 days and 3 hours
			time.Date(2018, 03, 2, 12, 0, 0, 0, time.Local).Unix(),
			time.Date(2018, 03, 9, 15, 0, 0, 0, time.Local).Unix(),
			// frequency of 10 minutes
			600,
			// expected height: roundUp((171h * 60 * 60) / 600)
			1026,
		},
	}
	for index, tt := range tests {
		h := estimatedHeightBetween(tt.From, tt.To, tt.BlockFrequency)
		if h != tt.ExpectedHeight {
			t.Error(index, h, "!=", tt.ExpectedHeight, tt)
		}
	}
}
