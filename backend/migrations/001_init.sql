CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS users (
  id UUID PRIMARY KEY,
  email TEXT NOT NULL UNIQUE,
  username TEXT NOT NULL UNIQUE,
  password_hash TEXT NOT NULL DEFAULT '',
  display_name TEXT NOT NULL DEFAULT '',
  phone TEXT,
  age INTEGER,
  country_code CHAR(2),
  timezone TEXT NOT NULL DEFAULT 'UTC',
  kyc_status TEXT NOT NULL DEFAULT 'pending',
  biometric_enabled BOOLEAN NOT NULL DEFAULT FALSE,
  theme_pref TEXT NOT NULL DEFAULT 'system',
  email_verified_at TIMESTAMPTZ,
  onboarding_completed_at TIMESTAMPTZ,
  status TEXT NOT NULL DEFAULT 'active',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT users_status_check CHECK (status IN ('active', 'suspended', 'deleted')),
  CONSTRAINT users_age_check CHECK (age IS NULL OR age >= 18),
  CONSTRAINT users_kyc_status_check CHECK (kyc_status IN ('pending', 'verified', 'rejected')),
  CONSTRAINT users_theme_pref_check CHECK (theme_pref IN ('system', 'dark', 'light'))
);

CREATE TABLE IF NOT EXISTS auth_email_otps (
  id UUID PRIMARY KEY,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  email TEXT NOT NULL,
  purpose TEXT NOT NULL,
  code_hash TEXT NOT NULL,
  attempts INTEGER NOT NULL DEFAULT 0,
  expires_at TIMESTAMPTZ NOT NULL,
  consumed_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT auth_email_otps_purpose_check CHECK (purpose IN ('register_verify', 'login_verify', 'password_reset')),
  CONSTRAINT auth_email_otps_attempts_non_negative CHECK (attempts >= 0)
);

CREATE TABLE IF NOT EXISTS auth_sessions (
  id UUID PRIMARY KEY,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  token_hash TEXT NOT NULL UNIQUE,
  user_agent TEXT,
  ip_address TEXT,
  expires_at TIMESTAMPTZ NOT NULL,
  revoked_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  last_seen_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS user_exchange_preferences (
  id UUID PRIMARY KEY,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  exchange_name TEXT NOT NULL,
  selected_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT user_exchange_preferences_exchange_check CHECK (exchange_name IN ('binance', 'okx', 'bybit', 'tokocrypto', 'kucoin', 'coinbase')),
  CONSTRAINT user_exchange_preferences_unique UNIQUE (user_id, exchange_name)
);

CREATE TABLE IF NOT EXISTS admin_users (
  id UUID PRIMARY KEY,
  email TEXT NOT NULL UNIQUE,
  display_name TEXT NOT NULL,
  password_hash TEXT NOT NULL,
  role TEXT NOT NULL,
  otp_secret_ciphertext BYTEA,
  otp_enabled BOOLEAN NOT NULL DEFAULT FALSE,
  status TEXT NOT NULL DEFAULT 'active',
  last_login_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT admin_users_role_check CHECK (role IN ('super_admin', 'admin', 'ops', 'auditor')),
  CONSTRAINT admin_users_status_check CHECK (status IN ('active', 'suspended', 'revoked'))
);

CREATE TABLE IF NOT EXISTS admin_auth_sessions (
  id UUID PRIMARY KEY,
  admin_id UUID NOT NULL REFERENCES admin_users(id) ON DELETE CASCADE,
  token_hash TEXT NOT NULL UNIQUE,
  user_agent TEXT,
  ip_address TEXT,
  expires_at TIMESTAMPTZ NOT NULL,
  revoked_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  last_seen_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS exchange_bindings (
  id UUID PRIMARY KEY,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  exchange_name TEXT NOT NULL,
  api_key_ciphertext BYTEA NOT NULL,
  api_secret_ciphertext BYTEA NOT NULL,
  api_passphrase_ciphertext BYTEA,
  permission_scope TEXT NOT NULL DEFAULT 'trade_only',
  status TEXT NOT NULL DEFAULT 'active',
  last_verified_at TIMESTAMPTZ,
  revoked_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT exchange_bindings_exchange_check CHECK (exchange_name IN ('binance', 'okx', 'bybit', 'tokocrypto')),
  CONSTRAINT exchange_bindings_status_check CHECK (status IN ('active', 'invalid', 'revoked')),
  CONSTRAINT exchange_bindings_one_active_per_exchange UNIQUE (user_id, exchange_name)
);

CREATE TABLE IF NOT EXISTS exchange_balance_snapshots (
  id UUID PRIMARY KEY,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  exchange_binding_id UUID NOT NULL REFERENCES exchange_bindings(id) ON DELETE CASCADE,
  asset TEXT NOT NULL,
  free_amount NUMERIC(36,18) NOT NULL DEFAULT 0,
  locked_amount NUMERIC(36,18) NOT NULL DEFAULT 0,
  captured_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT exchange_balance_snapshots_amount_non_negative CHECK (free_amount >= 0 AND locked_amount >= 0)
);

CREATE TABLE IF NOT EXISTS market_prices (
  id UUID PRIMARY KEY,
  symbol TEXT NOT NULL,
  price_quote NUMERIC(36,18) NOT NULL,
  source TEXT NOT NULL DEFAULT 'internal',
  captured_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT market_prices_price_positive CHECK (price_quote > 0)
);

CREATE TABLE IF NOT EXISTS master_signals (
  id UUID PRIMARY KEY,
  admin_id UUID NOT NULL REFERENCES admin_users(id),
  type TEXT NOT NULL,
  symbol TEXT NOT NULL,
  layer_number INTEGER,
  allocation_pct NUMERIC(36,18) NOT NULL DEFAULT 0,
  sell_pct NUMERIC(36,18) NOT NULL DEFAULT 0,
  status TEXT NOT NULL DEFAULT 'draft',
  idempotency_key TEXT NOT NULL UNIQUE,
  preview_snapshot JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  dispatched_at TIMESTAMPTZ,
  completed_at TIMESTAMPTZ,
  CONSTRAINT master_signals_type_check CHECK (type IN ('buy', 'sell')),
  CONSTRAINT master_signals_status_check CHECK (status IN ('draft', 'dispatching', 'completed', 'failed', 'cancelled')),
  CONSTRAINT master_signals_sell_requires_layer CHECK (type = 'buy' OR layer_number IS NOT NULL)
);

CREATE TABLE IF NOT EXISTS layers (
  id UUID PRIMARY KEY,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  exchange_binding_id UUID NOT NULL REFERENCES exchange_bindings(id),
  master_signal_id UUID NOT NULL REFERENCES master_signals(id),
  layer_number INTEGER NOT NULL,
  symbol TEXT NOT NULL,
  entry_price NUMERIC(36,18) NOT NULL,
  entry_quantity NUMERIC(36,18) NOT NULL,
  entry_value_quote NUMERIC(36,18) NOT NULL,
  remaining_quantity NUMERIC(36,18) NOT NULL,
  allocation_pct NUMERIC(36,18) NOT NULL,
  status TEXT NOT NULL DEFAULT 'open',
  orphan_reason TEXT,
  opened_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  closed_at TIMESTAMPTZ,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT layers_status_check CHECK (status IN ('open', 'partial', 'closed', 'orphaned')),
  CONSTRAINT layers_quantity_non_negative CHECK (entry_quantity >= 0 AND remaining_quantity >= 0),
  CONSTRAINT layers_layer_number_unique UNIQUE (user_id, symbol, layer_number)
);

CREATE TABLE IF NOT EXISTS layer_executions (
  id UUID PRIMARY KEY,
  layer_id UUID REFERENCES layers(id) ON DELETE SET NULL,
  master_signal_id UUID NOT NULL REFERENCES master_signals(id),
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  exchange_binding_id UUID NOT NULL REFERENCES exchange_bindings(id),
  action TEXT NOT NULL,
  symbol TEXT NOT NULL,
  quantity NUMERIC(36,18) NOT NULL,
  price NUMERIC(36,18) NOT NULL,
  value_quote NUMERIC(36,18) NOT NULL,
  exchange_fee NUMERIC(36,18) NOT NULL DEFAULT 0,
  exchange_order_id TEXT,
  status TEXT NOT NULL DEFAULT 'pending',
  error_code TEXT,
  error_message TEXT,
  idempotency_key TEXT NOT NULL UNIQUE,
  executed_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT layer_executions_action_check CHECK (action IN ('buy', 'sell')),
  CONSTRAINT layer_executions_status_check CHECK (status IN ('pending', 'success', 'failed', 'skipped', 'partial'))
);

CREATE TABLE IF NOT EXISTS gas_fee_ledger (
  id UUID PRIMARY KEY,
  layer_id UUID NOT NULL REFERENCES layers(id),
  execution_id UUID NOT NULL REFERENCES layer_executions(id),
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  type TEXT NOT NULL,
  gross_pnl NUMERIC(36,18) NOT NULL,
  gas_fee_amount NUMERIC(36,18) NOT NULL,
  platform_rebate NUMERIC(36,18) NOT NULL DEFAULT 0,
  net_amount_user NUMERIC(36,18) NOT NULL,
  share_rate NUMERIC(36,18) NOT NULL,
  calculated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT gas_fee_ledger_type_check CHECK (type IN ('profit_share', 'loss_rebate', 'breakeven'))
);

CREATE TABLE IF NOT EXISTS gas_fee_deposits (
  id UUID PRIMARY KEY,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  amount NUMERIC(36,18) NOT NULL,
  asset TEXT NOT NULL DEFAULT 'USDT',
  deposit_address TEXT NOT NULL,
  tx_id TEXT,
  status TEXT NOT NULL DEFAULT 'pending',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  confirmed_at TIMESTAMPTZ,
  CONSTRAINT gas_fee_deposits_amount_min CHECK (amount >= 500),
  CONSTRAINT gas_fee_deposits_status_check CHECK (status IN ('pending', 'confirmed', 'rejected'))
);

CREATE TABLE IF NOT EXISTS execution_jobs (
  id UUID PRIMARY KEY,
  master_signal_id UUID NOT NULL REFERENCES master_signals(id),
  layer_id UUID REFERENCES layers(id),
  user_id UUID NOT NULL REFERENCES users(id),
  exchange_binding_id UUID NOT NULL REFERENCES exchange_bindings(id),
  subject TEXT NOT NULL,
  payload JSONB NOT NULL,
  status TEXT NOT NULL DEFAULT 'queued',
  attempts INTEGER NOT NULL DEFAULT 0,
  idempotency_key TEXT NOT NULL UNIQUE,
  last_error TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT execution_jobs_status_check CHECK (status IN ('queued', 'published', 'running', 'success', 'failed', 'skipped', 'dead_letter'))
);

CREATE TABLE IF NOT EXISTS reconciliation_events (
  id UUID PRIMARY KEY,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  exchange_binding_id UUID NOT NULL REFERENCES exchange_bindings(id),
  master_signal_id UUID REFERENCES master_signals(id),
  layer_id UUID REFERENCES layers(id),
  event_type TEXT NOT NULL,
  asset TEXT NOT NULL,
  required_amount NUMERIC(36,18) NOT NULL DEFAULT 0,
  available_amount NUMERIC(36,18) NOT NULL DEFAULT 0,
  reason TEXT NOT NULL,
  status TEXT NOT NULL DEFAULT 'open',
  resolved_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT reconciliation_events_amount_non_negative CHECK (required_amount >= 0 AND available_amount >= 0),
  CONSTRAINT reconciliation_events_status_check CHECK (status IN ('open', 'resolved', 'ignored'))
);

CREATE TABLE IF NOT EXISTS notifications (
  id UUID PRIMARY KEY,
  user_id UUID REFERENCES users(id) ON DELETE CASCADE,
  type TEXT NOT NULL,
  title TEXT NOT NULL,
  message TEXT NOT NULL,
  read_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS audit_logs (
  id UUID PRIMARY KEY,
  actor_type TEXT NOT NULL,
  actor_id UUID,
  action TEXT NOT NULL,
  entity TEXT NOT NULL,
  entity_id UUID,
  before_state JSONB,
  after_state JSONB,
  ip_address INET,
  user_agent TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_exchange_bindings_user_status ON exchange_bindings(user_id, status);
CREATE INDEX IF NOT EXISTS idx_auth_email_otps_user_purpose_created ON auth_email_otps(user_id, purpose, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_auth_sessions_user_expires ON auth_sessions(user_id, expires_at DESC);
CREATE INDEX IF NOT EXISTS idx_admin_auth_sessions_admin_expires ON admin_auth_sessions(admin_id, expires_at DESC);
CREATE INDEX IF NOT EXISTS idx_user_exchange_preferences_user ON user_exchange_preferences(user_id, selected_at DESC);
CREATE INDEX IF NOT EXISTS idx_exchange_balance_snapshots_latest ON exchange_balance_snapshots(exchange_binding_id, asset, captured_at DESC);
CREATE INDEX IF NOT EXISTS idx_market_prices_latest ON market_prices(symbol, captured_at DESC);
CREATE INDEX IF NOT EXISTS idx_master_signals_status_created ON master_signals(status, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_layers_user_status ON layers(user_id, status);
CREATE INDEX IF NOT EXISTS idx_layers_symbol_status ON layers(symbol, status);
CREATE INDEX IF NOT EXISTS idx_layer_executions_signal_status ON layer_executions(master_signal_id, status);
CREATE INDEX IF NOT EXISTS idx_gas_fee_ledger_user_calculated ON gas_fee_ledger(user_id, calculated_at DESC);
CREATE INDEX IF NOT EXISTS idx_gas_fee_deposits_user_created ON gas_fee_deposits(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_gas_fee_deposits_status_created ON gas_fee_deposits(status, created_at DESC);
CREATE UNIQUE INDEX IF NOT EXISTS idx_gas_fee_deposits_asset_tx_id_unique ON gas_fee_deposits(asset, tx_id) WHERE tx_id IS NOT NULL AND tx_id <> '';
CREATE INDEX IF NOT EXISTS idx_execution_jobs_status_created ON execution_jobs(status, created_at);
CREATE UNIQUE INDEX IF NOT EXISTS idx_execution_jobs_one_inflight_sell_per_layer ON execution_jobs(layer_id)
  WHERE layer_id IS NOT NULL
    AND subject = 'execution.sell.request'
    AND status IN ('queued', 'published', 'running');
CREATE INDEX IF NOT EXISTS idx_reconciliation_events_status_created ON reconciliation_events(status, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_reconciliation_events_layer ON reconciliation_events(layer_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_logs_entity_created ON audit_logs(entity, entity_id, created_at DESC);
