package driver

import "time"

// Statement is one normalized query in the cumulative digest ranking
// (pg_stat_statements / events_statements_summary_by_digest).
type Statement struct {
	// queryid (PostgreSQL) or digest (MySQL).
	ID    string
	Query string

	Calls int64
	Total time.Duration // total execution time; default sort key
	Mean  time.Duration // mean execution time per call
	Rows  int64         // total rows; per-call is Rows / Calls

	// Full scan without an index (MySQL SUM_NO_INDEX_USED); always false on PostgreSQL.
	NoIndex bool

	Min    time.Duration
	Max    time.Duration
	Stddev time.Duration

	// PostgreSQL-only.
	SharedBlocksHit  int64
	SharedBlocksRead int64

	// MySQL-only.
	RowsExamined int64
}
