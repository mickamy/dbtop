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

func TestFromMillis(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in   float64
		want time.Duration
	}{
		{0.26, 260 * time.Microsecond},
		{0, 0},
		{1500, 1500 * time.Millisecond},
	}

	for _, tt := range tests {
		t.Run(tt.want.String(), func(t *testing.T) {
			t.Parallel()

			if got := durations.FromMillis(tt.in); got != tt.want {
				t.Errorf("FromMillis(%v) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}
