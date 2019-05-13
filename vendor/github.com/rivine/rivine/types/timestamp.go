package types

// timestamp.go defines the timestamp type and implements sort.Interface
// interface for slices of timestamps.

import (
	"fmt"
	"time"
)

type (
	Timestamp      uint64
	TimestampSlice []Timestamp
)

func (t Timestamp) String() string {
	return time.Unix(int64(t), 0).String()
}

func (t *Timestamp) LoadString(str string) error {
	_, err := fmt.Sscan(str, t)
	return err
}

// CurrentTimestamp returns the offset based on the current time as a Timestamp.
func OffsetTimestamp(offset time.Duration) Timestamp {
	return Timestamp(time.Now().Add(offset).Unix())
}

// CurrentTimestamp returns the current time as a Timestamp.
func CurrentTimestamp() Timestamp {
	return Timestamp(time.Now().Unix())
}

// Len is part of sort.Interface
func (ts TimestampSlice) Len() int {
	return len(ts)
}

// Less is part of sort.Interface
func (ts TimestampSlice) Less(i, j int) bool {
	return ts[i] < ts[j]
}

// Swap is part of sort.Interface
func (ts TimestampSlice) Swap(i, j int) {
	ts[i], ts[j] = ts[j], ts[i]
}

// Clock allows clients to retrieve the current time.
type Clock interface {
	Now() Timestamp
}

// StdClock is an implementation of Clock that retrieves the current time using
// the system time.
type StdClock struct{}

// Now retrieves the current timestamp.
func (c StdClock) Now() Timestamp {
	return Timestamp(time.Now().Unix())
}
