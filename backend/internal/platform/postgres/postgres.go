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
	_, _ = pool.Exec(ctx, `
CREATE TABLE IF NOT EXISTS admin_personal_wallets (
  code TEXT PRIMARY KEY,
  display_name TEXT NOT NULL,
  wallet_address TEXT NOT NULL DEFAULT '',
  updated_by UUID REFERENCES admin_users(id) ON DELETE SET NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT admin_personal_wallets_code_check CHECK (code IN ('viksry', 'aryanto_hong')),
  CONSTRAINT admin_personal_wallets_address_check CHECK (wallet_address = '' OR wallet_address ~* '^0x[0-9a-f]{40}$')
)`)
	_, _ = pool.Exec(ctx, `
INSERT INTO admin_personal_wallets (code, display_name)
VALUES
  ('viksry', 'WALLET VIKSRY'),
  ('aryanto_hong', 'WALLET ARYANTO HONG')
ON CONFLICT (code) DO NOTHING`)

	return pool, nil
}
