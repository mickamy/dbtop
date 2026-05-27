package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/mickamy/dbtop/internal/driver"
)

var _ driver.Driver = (*Driver)(nil)

// PostgreSQL 14.0 (server_version_num); 14+ gives both total_exec_time and pg_blocking_pids.
const minServerVersionNum = 140000

// Small pool: a poll and a kill may run at once, nothing more.
const maxConns = 3

// appName tags dbtop's own connections so they can be excluded from metrics.
const appName = "dbtop"

type Driver struct {
	pool *pgxpool.Pool
}

// Open connects to dsn and verifies the server is PostgreSQL 14+. The caller
// must Close the returned Driver.
func Open(ctx context.Context, dsn string) (*Driver, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse dsn: %w", err)
	}

	cfg.MaxConns = maxConns
	cfg.ConnConfig.RuntimeParams["application_name"] = appName

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("connect: %w", err)
	}

	var version int
	if err := pool.QueryRow(ctx, "SELECT current_setting('server_version_num')::int").Scan(&version); err != nil {
		pool.Close()

		return nil, fmt.Errorf("read server version: %w", err)
	}

	if err := guardVersion(version); err != nil {
		pool.Close()

		return nil, err
	}

	return &Driver{pool: pool}, nil
}

func (d *Driver) Close() error {
	d.pool.Close()

	return nil
}

func guardVersion(num int) error {
	if num < minServerVersionNum {
		return fmt.Errorf("dbtop requires PostgreSQL 14+ (server_version_num=%d)", num)
	}

	return nil
}
