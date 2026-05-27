//go:build integration

package postgres_test

import (
	"os"
	"testing"

	"github.com/mickamy/dbtop/internal/driver/postgres"
)

// Run with: DBTOP_TEST_DSN=postgres://... go test -tags integration ./internal/driver/postgres/
func openTestDriver(t *testing.T) *postgres.Driver {
	t.Helper()

	dsn := os.Getenv("DBTOP_TEST_DSN")
	if dsn == "" {
		t.Skip("set DBTOP_TEST_DSN to run integration tests")
	}

	d, err := postgres.Open(t.Context(), dsn)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	t.Cleanup(func() {
		_ = d.Close()
	})

	return d
}

func TestIntegrationActivity(t *testing.T) {
	t.Parallel()

	d := openTestDriver(t)

	backends, err := d.Activity(t.Context())
	if err != nil {
		t.Fatalf("Activity: %v", err)
	}

	for _, b := range backends {
		if b.PID == 0 {
			t.Errorf("backend has zero PID: %+v", b)
		}
	}
}

func TestIntegrationMetrics(t *testing.T) {
	t.Parallel()

	d := openTestDriver(t)

	m, err := d.Metrics(t.Context())
	if err != nil {
		t.Fatalf("Metrics: %v", err)
	}

	if m.MaxConnections <= 0 {
		t.Errorf("MaxConnections = %d, want > 0", m.MaxConnections)
	}
	if m.At.IsZero() {
		t.Error("At is zero")
	}
}

func TestIntegrationCapabilities(t *testing.T) {
	t.Parallel()

	d := openTestDriver(t)

	// Smoke test: the probe must succeed; values depend on the connected role.
	if _, err := d.Capabilities(t.Context()); err != nil {
		t.Fatalf("Capabilities: %v", err)
	}
}

func TestIntegrationStatements(t *testing.T) {
	t.Parallel()

	d := openTestDriver(t)
	ctx := t.Context()

	caps, err := d.Capabilities(ctx)
	if err != nil {
		t.Fatalf("Capabilities: %v", err)
	}
	if !caps.Statements {
		t.Skip("pg_stat_statements not available")
	}

	if _, err := d.Statements(ctx); err != nil {
		t.Fatalf("Statements: %v", err)
	}
}
