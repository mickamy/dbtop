-- Runs once on a fresh data directory (after postgres starts with
-- shared_preload_libraries=pg_stat_statements), so the Statements screen works
-- out of the box in local and CI databases.
CREATE EXTENSION IF NOT EXISTS pg_stat_statements;
