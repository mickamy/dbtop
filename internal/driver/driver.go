package driver

import "context"

// Driver abstracts a database backend behind the screens the TUI renders. The
// TUI is driver-independent: it only consumes []Backend, MetricSample, and
// []Statement.
type Driver interface {
	Activity(ctx context.Context) ([]Backend, error)
	Metrics(ctx context.Context) (MetricSample, error)
	Statements(ctx context.Context) ([]Statement, error)
	ResetStatements(ctx context.Context) error

	// Cancel gracefully stops the running query (pg_cancel_backend / KILL QUERY).
	Cancel(ctx context.Context, pid int64) error

	// Terminate forcibly closes the connection (pg_terminate_backend / KILL CONNECTION).
	Terminate(ctx context.Context, pid int64) error

	Capabilities(ctx context.Context) (Caps, error)
	Close() error
}

// Caps reports the privileges and features available to the connected user,
// used to degrade gracefully and to render the startup banner.
type Caps struct {
	Superuser bool

	// Can read other backends' full query text (pg_monitor / PROCESS).
	Monitor bool

	// Can cancel/terminate backends (pg_signal_backend / CONNECTION_ADMIN).
	Kill bool

	// Digest source available (pg_stat_statements / performance_schema digest).
	Statements bool
}
