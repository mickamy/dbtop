package postgres

import (
	"context"
	"fmt"
)

const (
	cancelQuery    = `SELECT pg_cancel_backend($1)`
	terminateQuery = `SELECT pg_terminate_backend($1)`
)

func (d *Driver) Cancel(ctx context.Context, pid int64) error {
	return d.signalBackend(ctx, cancelQuery, pid)
}

func (d *Driver) Terminate(ctx context.Context, pid int64) error {
	return d.signalBackend(ctx, terminateQuery, pid)
}

func (d *Driver) signalBackend(ctx context.Context, query string, pid int64) error {
	var sent bool
	if err := d.pool.QueryRow(ctx, query, pid).Scan(&sent); err != nil {
		return fmt.Errorf("signal backend %d: %w", pid, err)
	}

	// false means there was no such backend to signal.
	if !sent {
		return fmt.Errorf("backend %d not found", pid)
	}

	return nil
}
