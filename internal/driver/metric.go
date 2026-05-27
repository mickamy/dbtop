package driver

import "time"

// MetricSample is one polling sample of server-wide health counters. Drivers
// return raw cumulative counters and gauges; per-second rates are derived
// downstream from the delta between successive samples.
type MetricSample struct {
	At time.Time

	MaxConnections int
	Conns          ConnCounts

	// Monotonic counters; rate = delta / dt.
	Commits        int64
	Rollbacks      int64
	BlocksHit      int64
	BlocksRead     int64
	TuplesInserted int64
	TuplesUpdated  int64
	TuplesDeleted  int64
	TuplesReturned int64
	TuplesFetched  int64
	TempFiles      int64
	TempBytes      int64

	// Point-in-time value; use as-is, not as a rate.
	WaitingLocks int

	Replicas []ReplicaLag
}

type ConnCounts struct {
	Total    int
	Active   int
	Idle     int
	IdleInTx int
}

type ReplicaLag struct {
	Client string
	Lag    time.Duration
}
