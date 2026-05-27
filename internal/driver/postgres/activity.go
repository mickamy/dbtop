package postgres

import (
	"context"
	"fmt"

	"github.com/mickamy/dbtop/internal/driver"
	"github.com/mickamy/dbtop/internal/durations"
	"github.com/mickamy/dbtop/internal/ptr"
)

const activityQuery = `
SELECT a.pid, a.usename, a.datname, a.state,
       a.wait_event_type, a.wait_event, a.query, a.backend_type,
       pg_blocking_pids(a.pid)::bigint[]                 AS blocked_by,
       EXTRACT(EPOCH FROM now() - a.query_start)::float8 AS query_age,
       EXTRACT(EPOCH FROM now() - a.xact_start)::float8  AS xact_age
FROM pg_stat_activity a
WHERE a.pid <> pg_backend_pid()
ORDER BY CASE
           WHEN a.state = 'active'
             THEN EXTRACT(EPOCH FROM now() - a.query_start)
           WHEN a.state IN ('idle in transaction', 'idle in transaction (aborted)')
             THEN EXTRACT(EPOCH FROM now() - a.xact_start)
           ELSE NULL
         END DESC NULLS LAST
`

func (d *Driver) Activity(ctx context.Context) ([]driver.Backend, error) {
	rows, err := d.pool.Query(ctx, activityQuery)
	if err != nil {
		return nil, fmt.Errorf("query activity: %w", err)
	}
	defer rows.Close()

	var backends []driver.Backend

	for rows.Next() {
		var (
			pid                 int64
			user, db, state     *string
			waitType, waitEvent *string
			query, backendType  *string
			blockedBy           []int64
			queryAge, xactAge   *float64
		)

		if err := rows.Scan(
			&pid, &user, &db, &state, &waitType, &waitEvent, &query, &backendType, &blockedBy, &queryAge, &xactAge,
		); err != nil {
			return nil, fmt.Errorf("scan activity: %w", err)
		}

		backends = append(backends, driver.Backend{
			PID:           pid,
			User:          ptr.OrZero(user),
			DB:            ptr.OrZero(db),
			State:         parseState(ptr.OrZero(state)),
			WaitType:      ptr.OrZero(waitType),
			WaitEvent:     ptr.OrZero(waitEvent),
			Query:         ptr.OrZero(query),
			BlockedBy:     blockedBy,
			QueryAge:      ptr.Map(queryAge, durations.FromSeconds),
			XactAge:       ptr.Map(xactAge, durations.FromSeconds),
			SystemBackend: ptr.OrZero(backendType) != "client backend",
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate activity: %w", err)
	}

	return backends, nil
}

func parseState(s string) driver.State {
	switch s {
	case "active":
		return driver.StateActive
	case "idle":
		return driver.StateIdle
	case "idle in transaction":
		return driver.StateIdleInTx
	case "idle in transaction (aborted)":
		return driver.StateIdleInTxAborted
	default:
		return driver.StateUnknown
	}
}
