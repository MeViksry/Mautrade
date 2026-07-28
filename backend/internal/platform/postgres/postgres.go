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
	_, _ = pool.Exec(ctx, `CREATE EXTENSION IF NOT EXISTS pgcrypto`)
	_, _ = pool.Exec(ctx, `
CREATE TABLE IF NOT EXISTS admin_wallet_commission_ledger (
  id UUID PRIMARY KEY,
  deposit_id UUID NOT NULL REFERENCES gas_fee_deposits(id) ON DELETE CASCADE,
  wallet_code TEXT NOT NULL REFERENCES admin_personal_wallets(code) ON DELETE RESTRICT,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  asset TEXT NOT NULL DEFAULT 'USDT',
  share_rate NUMERIC(36,18) NOT NULL,
  amount NUMERIC(36,18) NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT admin_wallet_commission_wallet_code_check CHECK (wallet_code IN ('viksry', 'aryanto_hong')),
  CONSTRAINT admin_wallet_commission_share_rate_check CHECK (share_rate >= 0 AND share_rate <= 1),
  CONSTRAINT admin_wallet_commission_amount_check CHECK (amount >= 0),
  CONSTRAINT admin_wallet_commission_deposit_wallet_unique UNIQUE (deposit_id, wallet_code)
)`)
	_, _ = pool.Exec(ctx, `
CREATE INDEX IF NOT EXISTS idx_admin_wallet_commission_wallet_created
  ON admin_wallet_commission_ledger (wallet_code, created_at DESC)`)
	_, _ = pool.Exec(ctx, `
INSERT INTO admin_wallet_commission_ledger (
  id, deposit_id, wallet_code, user_id, asset, share_rate, amount, created_at, updated_at
)
SELECT
  gen_random_uuid(),
  d.id,
  split.wallet_code,
  d.user_id,
  d.asset,
  split.share_rate,
  split.amount,
  COALESCE(d.confirmed_at, d.created_at),
  COALESCE(d.confirmed_at, d.created_at)
FROM gas_fee_deposits d
CROSS JOIN LATERAL (
  SELECT
    TRUNC(d.amount * 0.10, 18)::numeric(36,18) AS viksry_amount
) calc
CROSS JOIN LATERAL (
  VALUES
    ('viksry', 0.10::numeric(36,18), calc.viksry_amount),
    ('aryanto_hong', 0.90::numeric(36,18), (d.amount - calc.viksry_amount)::numeric(36,18))
) AS split(wallet_code, share_rate, amount)
WHERE d.status = 'confirmed'
ON CONFLICT (deposit_id, wallet_code) DO NOTHING`)
	_, _ = pool.Exec(ctx, `
CREATE TABLE IF NOT EXISTS admin_wallet_withdrawals (
  id UUID PRIMARY KEY,
  wallet_code TEXT NOT NULL REFERENCES admin_personal_wallets(code) ON DELETE RESTRICT,
  admin_id UUID REFERENCES admin_users(id) ON DELETE SET NULL,
  destination_address TEXT NOT NULL,
  amount NUMERIC(36,18) NOT NULL,
  asset TEXT NOT NULL DEFAULT 'USDT',
  status TEXT NOT NULL DEFAULT 'pending_signing',
  tx_id TEXT,
  failure_reason TEXT NOT NULL DEFAULT '',
  requested_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  broadcast_at TIMESTAMPTZ,
  confirmed_at TIMESTAMPTZ,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT admin_wallet_withdrawals_wallet_code_check CHECK (wallet_code IN ('viksry', 'aryanto_hong')),
  CONSTRAINT admin_wallet_withdrawals_destination_check CHECK (destination_address ~* '^0x[0-9a-f]{40}$'),
  CONSTRAINT admin_wallet_withdrawals_amount_check CHECK (amount > 0),
  CONSTRAINT admin_wallet_withdrawals_asset_check CHECK (asset = 'USDT'),
  CONSTRAINT admin_wallet_withdrawals_status_check CHECK (status IN ('pending_signing', 'broadcast', 'confirmed', 'failed', 'cancelled'))
)`)
	_, _ = pool.Exec(ctx, `
CREATE INDEX IF NOT EXISTS idx_admin_wallet_withdrawals_wallet_requested
  ON admin_wallet_withdrawals (wallet_code, requested_at DESC)`)
	_, _ = pool.Exec(ctx, `
CREATE UNIQUE INDEX IF NOT EXISTS idx_admin_wallet_withdrawals_tx_id_unique
  ON admin_wallet_withdrawals (tx_id)
  WHERE tx_id IS NOT NULL AND tx_id <> ''`)

	return pool, nil
}
