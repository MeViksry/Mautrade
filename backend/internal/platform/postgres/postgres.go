package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Connect(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	if databaseURL == "" {
		return nil, nil
	}

	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("postgres: parse database url: %w", err)
	}
	config.MaxConns = 16
	config.MinConns = 2
	config.MaxConnLifetime = 30 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("postgres: connect: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("postgres: ping: %w", err)
	}

	// HOTFIX: Update gas_fee_deposits constraint for existing production databases automatically on startup.
	// This allows the CI/CD deployment to fix the constraint without manual database intervention.
	_, _ = pool.Exec(ctx, `ALTER TABLE gas_fee_deposits DROP CONSTRAINT IF EXISTS gas_fee_deposits_amount_min`)
	_, _ = pool.Exec(ctx, `ALTER TABLE gas_fee_deposits ADD CONSTRAINT gas_fee_deposits_amount_min CHECK (amount > 0)`)
	_, _ = pool.Exec(ctx, `ALTER TABLE exchange_bindings ADD COLUMN IF NOT EXISTS revoked_at TIMESTAMPTZ`)
	_, _ = pool.Exec(ctx, `
UPDATE exchange_bindings
SET status = CASE
  WHEN lower(status) IN ('active', 'connected', 'connect') THEN 'active'
  WHEN lower(status) IN ('revoked', 'disconnected', 'disconnect', 'deleted', 'delete') THEN 'revoked'
  WHEN lower(status) = 'invalid' THEN 'invalid'
  ELSE 'invalid'
END`)
	_, _ = pool.Exec(ctx, `ALTER TABLE exchange_bindings DROP CONSTRAINT IF EXISTS exchange_bindings_status_check`)
	_, _ = pool.Exec(ctx, `ALTER TABLE exchange_bindings ADD CONSTRAINT exchange_bindings_status_check CHECK (status IN ('active', 'invalid', 'revoked'))`)

	return pool, nil
}
