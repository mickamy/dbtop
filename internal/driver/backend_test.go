package driver_test

import (
	"testing"
	"time"

	"github.com/mickamy/dbtop/internal/driver"
)

func TestStateString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		state driver.State
		want  string
	}{
		{driver.StateActive, "active"},
		{driver.StateIdle, "idle"},
		{driver.StateIdleInTx, "idle-tx"},
		{driver.StateIdleInTxAborted, "idle-tx-aborted"},
		{driver.StateUnknown, "unknown"},
	}

	for _, tt := range tests {
		if got := tt.state.String(); got != tt.want {
			t.Errorf("State(%d).String() = %q, want %q", tt.state, got, tt.want)
		}
	}
}

func TestBackendDuration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		backend driver.Backend
		want    time.Duration
		wantOK  bool
	}{
		{
			name:    "active uses query age",
			backend: driver.Backend{State: driver.StateActive, QueryAge: new(3 * time.Second), XactAge: new(time.Minute)},
			want:    3 * time.Second,
			wantOK:  true,
		},
		{
			name:    "idle-tx uses xact age",
			backend: driver.Backend{State: driver.StateIdleInTx, QueryAge: new(time.Second), XactAge: new(90 * time.Second)},
			want:    90 * time.Second,
			wantOK:  true,
		},
		{
			name:    "idle-tx-aborted uses xact age",
			backend: driver.Backend{State: driver.StateIdleInTxAborted, XactAge: new(30 * time.Second)},
			want:    30 * time.Second,
			wantOK:  true,
		},
		{
			name:    "active without query age is undefined",
			backend: driver.Backend{State: driver.StateActive},
			wantOK:  false,
		},
		{
			name:    "idle ignores stale query age",
			backend: driver.Backend{State: driver.StateIdle, QueryAge: new(2 * time.Hour)},
			wantOK:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, ok := tt.backend.Duration()
			if ok != tt.wantOK || got != tt.want {
				t.Errorf("Duration() = (%v, %v), want (%v, %v)", got, ok, tt.want, tt.wantOK)
			}
		})
	}
}

func TestBackendWait(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		backend driver.Backend
		want    string
	}{
		{"lock wait", driver.Backend{WaitType: "Lock", WaitEvent: "tuple"}, "Lock:tuple"},
		{"not waiting", driver.Backend{}, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.backend.Wait(); got != tt.want {
				t.Errorf("Wait() = %q, want %q", got, tt.want)
			}
		})
	}
}
