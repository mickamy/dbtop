package durations_test

import (
	"testing"
	"time"

	"github.com/mickamy/dbtop/internal/durations"
)

func TestFromSeconds(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in   float64
		want time.Duration
	}{
		{1.5, 1500 * time.Millisecond},
		{0, 0},
		{90, 90 * time.Second},
	}

	for _, tt := range tests {
		t.Run(tt.want.String(), func(t *testing.T) {
			t.Parallel()

			if got := durations.FromSeconds(tt.in); got != tt.want {
				t.Errorf("FromSeconds(%v) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}
