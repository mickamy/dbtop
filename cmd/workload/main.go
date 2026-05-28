package main

import (
	"context"
	"fmt"
	"log"
	"math/rand/v2"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	// defaultDSN points at the local compose database; password is the dev one.
	defaultDSN    = "postgres://postgres:pass@postgres:5432/dev" //nolint:gosec // dev-only default DSN
	poolSize      = 20
	seedRows      = 1000
	normalWorkers = 8
	extraWorkers  = 4 // nPlusOne, longQueries, blocker, contender
	hotRowID      = 1 // the single row the blocker and contenders fight over
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("workload: %v", err)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	dsn := os.Getenv("WORKLOAD_DSN")
	if dsn == "" {
		dsn = defaultDSN
	}

	pool, err := connect(ctx, dsn)
	if err != nil {
		return err
	}
	defer pool.Close()

	if err := setup(ctx, pool); err != nil {
		return err
	}

	log.Println("workload: generating traffic (ctrl-c to stop)")

	workers := make([]func(context.Context, *pgxpool.Pool), 0, normalWorkers+extraWorkers)
	for range normalWorkers {
		workers = append(workers, normalLoad)
	}

	workers = append(workers, nPlusOne, longQueries, blocker, contender)

	var wg sync.WaitGroup

	for _, w := range workers {
		wg.Go(func() {
			w(ctx, pool)
		})
	}

	<-ctx.Done()
	wg.Wait()
	log.Println("workload: stopped")

	return nil
}

func connect(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse dsn: %w", err)
	}

	cfg.MaxConns = poolSize

	for {
		pool, err := pgxpool.NewWithConfig(ctx, cfg)
		if err == nil {
			if pingErr := pool.Ping(ctx); pingErr == nil {
				return pool, nil
			} else {
				pool.Close()
				err = pingErr
			}
		}

		log.Printf("workload: waiting for db: %v", err)

		if !wait(ctx, time.Second) {
			return nil, fmt.Errorf("canceled while connecting: %w", ctx.Err())
		}
	}
}

func setup(ctx context.Context, pool *pgxpool.Pool) error {
	// pg_stat_statements needs shared_preload_libraries; tolerate its absence.
	if _, err := pool.Exec(ctx, `CREATE EXTENSION IF NOT EXISTS pg_stat_statements`); err != nil {
		log.Printf("workload: pg_stat_statements unavailable: %v", err)
	}

	schema := []string{
		`CREATE TABLE IF NOT EXISTS items (
			id    int PRIMARY KEY,
			name  text NOT NULL,
			value int NOT NULL
		)`,
		fmt.Sprintf(`INSERT INTO items
			SELECT g, 'item-' || g, g
			FROM generate_series(1, %d) g
			ON CONFLICT (id) DO NOTHING`, seedRows),
	}

	for _, stmt := range schema {
		if _, err := pool.Exec(ctx, stmt); err != nil {
			return fmt.Errorf("setup: %w", err)
		}
	}

	return nil
}

// normalLoad does steady point reads and writes across random rows.
func normalLoad(ctx context.Context, pool *pgxpool.Pool) {
	for wait(ctx, jitter(5*time.Millisecond, 45*time.Millisecond)) {
		id := rand.IntN(seedRows) + 1 //nolint:gosec // weak RNG is fine for a load generator

		var (
			name  string
			value int
		)

		_ = pool.QueryRow(ctx, `SELECT name, value FROM items WHERE id = $1`, id).Scan(&name, &value)
		_, _ = pool.Exec(ctx, `UPDATE items SET value = value + 1 WHERE id = $1`, id)
	}
}

// nPlusOne fetches a batch of ids then queries each one separately, the classic
// pattern that is cheap per call but dominates pg_stat_statements by count.
func nPlusOne(ctx context.Context, pool *pgxpool.Pool) {
	for wait(ctx, 500*time.Millisecond) {
		rows, err := pool.Query(ctx, `SELECT id FROM items ORDER BY random() LIMIT 50`)
		if err != nil {
			continue
		}

		var ids []int

		for rows.Next() {
			var id int
			if err := rows.Scan(&id); err == nil {
				ids = append(ids, id)
			}
		}

		rows.Close()

		for _, id := range ids {
			var name string

			_ = pool.QueryRow(ctx, `SELECT name FROM items WHERE id = $1`, id).Scan(&name)
		}
	}
}

// longQueries runs an occasional multi-second query so the Activity screen has
// a long DURATION to surface.
func longQueries(ctx context.Context, pool *pgxpool.Pool) {
	for wait(ctx, 2*time.Second) {
		seconds := rand.IntN(8) + 3 //nolint:gosec // weak RNG is fine for a load generator

		_, _ = pool.Exec(ctx, `SELECT pg_sleep($1)`, seconds)
	}
}

// blocker holds a row lock inside an open transaction, then sits idle, creating
// an idle-in-transaction backend that blocks the contender.
func blocker(ctx context.Context, pool *pgxpool.Pool) {
	for wait(ctx, 3*time.Second) {
		holdLock(ctx, pool)
	}
}

func holdLock(ctx context.Context, pool *pgxpool.Pool) {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx, `UPDATE items SET value = value WHERE id = $1`, hotRowID); err != nil {
		return
	}

	wait(ctx, 5*time.Second)

	_ = tx.Commit(ctx)
}

// contender repeatedly updates the hot row, so it blocks whenever the blocker
// is holding the lock.
func contender(ctx context.Context, pool *pgxpool.Pool) {
	for wait(ctx, time.Second) {
		_, _ = pool.Exec(ctx, `UPDATE items SET value = value + 1 WHERE id = $1`, hotRowID)
	}
}

// wait sleeps for d, returning false if the context is canceled first.
func wait(ctx context.Context, d time.Duration) bool {
	select {
	case <-ctx.Done():
		return false
	case <-time.After(d):
		return true
	}
}

func jitter(minimum, maximum time.Duration) time.Duration {
	return minimum + rand.N(maximum-minimum) //nolint:gosec // weak RNG is fine for a load generator
}
