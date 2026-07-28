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
);

CREATE INDEX IF NOT EXISTS idx_admin_wallet_withdrawals_wallet_requested
  ON admin_wallet_withdrawals (wallet_code, requested_at DESC);

CREATE UNIQUE INDEX IF NOT EXISTS idx_admin_wallet_withdrawals_tx_id_unique
  ON admin_wallet_withdrawals (tx_id)
  WHERE tx_id IS NOT NULL AND tx_id <> '';
