package postgres_test

import (
	"testing"

	"github.com/mickamy/dbtop/internal/driver"
	"github.com/mickamy/dbtop/internal/driver/postgres"
)

func TestParseState(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in   string
		want driver.State
	}{
		{"active", driver.StateActive},
		{"idle", driver.StateIdle},
		{"idle in transaction", driver.StateIdleInTx},
		{"idle in transaction (aborted)", driver.StateIdleInTxAborted},
		{"", driver.StateUnknown},
		{"fastpath function call", driver.StateUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			t.Parallel()

			if got := postgres.ParseState(tt.in); got != tt.want {
				t.Errorf("ParseState(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}
