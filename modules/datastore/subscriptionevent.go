package datastore

import "github.com/rivine/rivine/types"

// The actions which can be passed through a channel
const (
	// SubStart indicates that a new subscribtion must begin, with an optional
	// starttime indicating the earliest block which must be tracked
	SubStart SubAction = "start"
	// SubEnd indicates that a subscription ends immediatly
	SubEnd SubAction = "end"
)

type (
	// SubAction is the action which must be done for a subscription
	SubAction string

	// SubEvent has some details regarding an event which has been received
	SubEvent struct {
		Action    SubAction
		Namespace Namespace
		Start     types.Timestamp // Optional starttime
	}
)
