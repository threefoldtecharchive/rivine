package config

import (
	"fmt"
	"math"
)

const (
	minNetPortValue = 1024
	maxNetPortValue = math.MaxUint16
)

var (
	// ErrAPIPortOutOfRange is an error
	ErrAPIPortOutOfRange = fmt.Errorf("API port range should be a number between %d and %d", minNetPortValue, maxNetPortValue)
)
