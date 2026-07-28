CREATE TABLE IF NOT EXISTS admin_personal_wallets (
  code TEXT PRIMARY KEY,
  display_name TEXT NOT NULL,
  wallet_address TEXT NOT NULL DEFAULT '',
  updated_by UUID REFERENCES admin_users(id) ON DELETE SET NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT admin_personal_wallets_code_check CHECK (code IN ('viksry', 'aryanto_hong')),
  CONSTRAINT admin_personal_wallets_address_check CHECK (wallet_address = '' OR wallet_address ~* '^0x[0-9a-f]{40}$')
);

INSERT INTO admin_personal_wallets (code, display_name)
VALUES
  ('viksry', 'WALLET VIKSRY'),
  ('aryanto_hong', 'WALLET ARYANTO HONG')
ON CONFLICT (code) DO NOTHING;
