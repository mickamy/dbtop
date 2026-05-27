package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/mickamy/dbtop/internal/driver"
	"github.com/mickamy/dbtop/internal/durations"
	"github.com/mickamy/dbtop/internal/ptr"
)

const metricsQuery = `
SELECT
  current_setting('max_connections')::int,
  (SELECT count(*) FROM pg_stat_activity WHERE backend_type = 'client backend'),
  (SELECT count(*) FROM pg_stat_activity WHERE backend_type = 'client backend' AND state = 'active'),
  (SELECT count(*) FROM pg_stat_activity WHERE backend_type = 'client backend' AND state = 'idle'),
  (SELECT count(*) FROM pg_stat_activity
   WHERE backend_type = 'client backend'
     AND state IN ('idle in transaction', 'idle in transaction (aborted)')),
  (SELECT count(*) FROM pg_locks WHERE NOT granted),
  d.xact_commit,
  d.xact_rollback,
  d.blks_hit,
  d.blks_read,
  d.tup_inserted,
  d.tup_updated,
  d.tup_deleted,
  d.tup_returned,
  d.tup_fetched,
  d.temp_files,
  d.temp_bytes
FROM pg_stat_database d
WHERE d.datname = current_database()
`

const replicaLagQuery = `
SELECT
  coalesce(nullif(client_addr::text, ''), application_name),
  EXTRACT(EPOCH FROM replay_lag)::float8
FROM pg_stat_replication
`

func (d *Driver) Metrics(ctx context.Context) (driver.MetricSample, error) {
	sample := driver.MetricSample{At: time.Now()}

	err := d.pool.QueryRow(ctx, metricsQuery).Scan(
		&sample.MaxConnections,
		&sample.Conns.Total,
		&sample.Conns.Active,
		&sample.Conns.Idle,
		&sample.Conns.IdleInTx,
		&sample.WaitingLocks,
		&sample.Commits,
		&sample.Rollbacks,
		&sample.BlocksHit,
		&sample.BlocksRead,
		&sample.TuplesInserted,
		&sample.TuplesUpdated,
		&sample.TuplesDeleted,
		&sample.TuplesReturned,
		&sample.TuplesFetched,
		&sample.TempFiles,
		&sample.TempBytes,
	)
	if err != nil {
		return driver.MetricSample{}, fmt.Errorf("query metrics: %w", err)
	}

	replicas, err := d.replicaLag(ctx)
	if err != nil {
		return driver.MetricSample{}, err
	}

	sample.Replicas = replicas

	return sample, nil
}

func (d *Driver) replicaLag(ctx context.Context) ([]driver.ReplicaLag, error) {
	rows, err := d.pool.Query(ctx, replicaLagQuery)
	if err != nil {
		return nil, fmt.Errorf("query replica lag: %w", err)
	}
	defer rows.Close()

	var replicas []driver.ReplicaLag

	for rows.Next() {
		var (
			client *string
			lagSec *float64
		)

		if err := rows.Scan(&client, &lagSec); err != nil {
			return nil, fmt.Errorf("scan replica lag: %w", err)
		}

		replicas = append(replicas, driver.ReplicaLag{
			Client: ptr.OrZero(client),
			Lag:    durations.FromSeconds(ptr.OrZero(lagSec)),
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate replica lag: %w", err)
	}

	return replicas, nil
}
