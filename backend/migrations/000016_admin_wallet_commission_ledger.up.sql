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
);

CREATE INDEX IF NOT EXISTS idx_admin_wallet_commission_wallet_created
  ON admin_wallet_commission_ledger (wallet_code, created_at DESC);

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
ON CONFLICT (deposit_id, wallet_code) DO NOTHING;
