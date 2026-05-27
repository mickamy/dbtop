package postgres_test

import (
	"testing"

	"github.com/mickamy/dbtop/internal/driver/postgres"
)

func TestGuardVersion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		num     int
		wantErr bool
	}{
		{"pg 13 rejected", 130010, true},
		{"pg 14.0 accepted", 140000, false},
		{"pg 16.2 accepted", 160002, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := postgres.GuardVersion(tt.num)
			if (err != nil) != tt.wantErr {
				t.Errorf("GuardVersion(%d) error = %v, wantErr %v", tt.num, err, tt.wantErr)
			}
		})
	}
}
