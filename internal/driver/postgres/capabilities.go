package postgres

import (
	"context"
	"fmt"

	"github.com/mickamy/dbtop/internal/driver"
)

const capabilitiesQuery = `
SELECT
  current_setting('is_superuser')::bool,
  pg_has_role(current_user, 'pg_monitor', 'MEMBER'),
  pg_has_role(current_user, 'pg_signal_backend', 'MEMBER'),
  EXISTS (SELECT 1 FROM pg_extension WHERE extname = 'pg_stat_statements')
`

func (d *Driver) Capabilities(ctx context.Context) (driver.Capabilities, error) {
	var capabilities driver.Capabilities

	err := d.pool.QueryRow(ctx, capabilitiesQuery).Scan(
		&capabilities.Superuser,
		&capabilities.Monitor,
		&capabilities.Kill,
		&capabilities.Statements,
	)
	if err != nil {
		return driver.Capabilities{}, fmt.Errorf("query capabilities: %w", err)
	}

	// A superuser can always monitor and kill, regardless of role membership.
	if capabilities.Superuser {
		capabilities.Monitor = true
		capabilities.Kill = true
	}

	return capabilities, nil
}
