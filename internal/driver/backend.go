package driver

import "time"

// State is a backend's connection state, normalized across drivers.
type State uint8

const (
	StateUnknown State = iota
	StateActive
	StateIdle
	StateIdleInTx
	StateIdleInTxAborted
)

func (s State) String() string {
	switch s {
	case StateUnknown:
		return "unknown"
	case StateActive:
		return "active"
	case StateIdle:
		return "idle"
	case StateIdleInTx:
		return "idle-tx"
	case StateIdleInTxAborted:
		return "idle-tx-aborted"
	}

	return "unknown"
}

// Backend is one server-side connection in an Activity snapshot.
type Backend struct {
	PID   int64
	User  string
	DB    string
	State State

	// Empty when the backend is not waiting.
	WaitType  string
	WaitEvent string

	Query string

	// PIDs blocking this backend; the ▲/⊘ markers are derived across rows.
	BlockedBy []int64

	// now() - query_start and now() - xact_start; nil when the timestamp is NULL.
	QueryAge *time.Duration
	XactAge  *time.Duration

	// Non-client backend, hidden by default.
	BackgroundWorker bool
}

// Duration returns the value shown in the DURATION column. Idle and unknown
// backends have none: their query_start is stale from the last query.
func (b Backend) Duration() (time.Duration, bool) {
	switch b.State {
	case StateActive:
		if b.QueryAge != nil {
			return *b.QueryAge, true
		}
	case StateIdleInTx, StateIdleInTxAborted:
		if b.XactAge != nil {
			return *b.XactAge, true
		}
	case StateUnknown, StateIdle:
	}

	return 0, false
}

func (b Backend) Wait() string {
	if b.WaitType == "" {
		return ""
	}

	return b.WaitType + ":" + b.WaitEvent
}
