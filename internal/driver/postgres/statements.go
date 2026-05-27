package postgres

import (
	"context"
	"fmt"
	"strconv"

	"github.com/mickamy/dbtop/internal/driver"
	"github.com/mickamy/dbtop/internal/durations"
	"github.com/mickamy/dbtop/internal/ptr"
)

const statementsQuery = `
SELECT
  queryid,
  query,
  calls,
  total_exec_time,
  mean_exec_time,
  min_exec_time,
  max_exec_time,
  stddev_exec_time,
  rows,
  shared_blks_hit,
  shared_blks_read
FROM pg_stat_statements
WHERE dbid = (SELECT oid FROM pg_database WHERE datname = current_database())
ORDER BY total_exec_time DESC
LIMIT 100
`

const resetStatementsQuery = `SELECT pg_stat_statements_reset()`

func (d *Driver) Statements(ctx context.Context) ([]driver.Statement, error) {
	rows, err := d.pool.Query(ctx, statementsQuery)
	if err != nil {
		return nil, fmt.Errorf("query statements: %w", err)
	}
	defer rows.Close()

	var statements []driver.Statement

	for rows.Next() {
		var (
			queryID                      *int64
			query                        *string
			calls                        int64
			total, mean                  float64
			minTime, maxTime, stddevTime float64
			rowCount                     int64
			blocksHit, blocksRead        int64
		)

		if err := rows.Scan(
			&queryID,
			&query,
			&calls,
			&total,
			&mean,
			&minTime,
			&maxTime,
			&stddevTime,
			&rowCount,
			&blocksHit,
			&blocksRead,
		); err != nil {
			return nil, fmt.Errorf("scan statement: %w", err)
		}

		statements = append(statements, driver.Statement{
			ID:               queryIDString(queryID),
			Query:            ptr.OrZero(query),
			Calls:            calls,
			Total:            durations.FromMillis(total),
			Mean:             durations.FromMillis(mean),
			Rows:             rowCount,
			Min:              durations.FromMillis(minTime),
			Max:              durations.FromMillis(maxTime),
			Stddev:           durations.FromMillis(stddevTime),
			SharedBlocksHit:  blocksHit,
			SharedBlocksRead: blocksRead,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate statements: %w", err)
	}

	return statements, nil
}

func (d *Driver) ResetStatements(ctx context.Context) error {
	if _, err := d.pool.Exec(ctx, resetStatementsQuery); err != nil {
		return fmt.Errorf("reset statements: %w", err)
	}

	return nil
}

func queryIDString(id *int64) string {
	if id == nil {
		return ""
	}

	return strconv.FormatInt(*id, 10)
}
