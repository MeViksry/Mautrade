# Mautrade Backend

Go control-plane backend for Mautrade copy trading.

## What is here now

- Go HTTP API skeleton under `cmd/api`.
- `quuid` UUIDv7 ID package usage through `internal/domain/id`.
- `qdecimal` Gas Fee engine under `internal/domain/gasfee`.
- PostgreSQL schema in `migrations/001_init.sql`.
- NATS JetStream stream setup for distributed Rust execution workers.
- PostgreSQL-backed user/admin dashboard reads with mock fallback only when `DATABASE_URL` is empty.
- User auth/onboarding flow: email/password register, hashed email OTP, bearer sessions, country/timezone/age exchange preferences, and minimum 500 USDT gas-fee deposit intent.
- Admin auth flow: bootstrap first admin from env, bcrypt password hashes, separate admin bearer sessions, and admin endpoint gating.
- Admin signal creation endpoint that writes `master_signals`, creates per-user `execution_jobs`, and publishes idempotent jobs to NATS when available.
- Execution result settlement from Rust reports: buy fills create independent Layers; sell fills update the targeted Layer and write Gas Fee Ledger rows with `qdecimal`.
- DB-backed reconciliation guard before queue publish: users with insufficient latest balance snapshots are written as skipped jobs, reconciliation events, and user notifications instead of being sent to Rust.
- Docker Compose stack for API, PostgreSQL, NATS, KeyDB, and QuestDB.

## Local

```bash
go test ./...
go run ./cmd/api
```

With infrastructure:

```bash
docker compose up --build
```

The API listens on `:8080` by default.

## HTTP API

- `GET /healthz`
- `GET /readyz`
- `POST /api/v1/auth/register`
- `POST /api/v1/auth/login`
- `POST /api/v1/auth/verify-otp`
- `GET /api/v1/auth/me`
- `POST /api/v1/auth/logout`
- `POST /api/v1/onboarding/complete`
- `GET /api/v1/user/stats`
- `GET /api/v1/user/exchange-bindings`
- `POST /api/v1/user/exchange-bindings`
- `GET /api/v1/user/exchange-bindings/{exchange}/credentials`
- `PATCH /api/v1/user/exchange-bindings/{exchange}/status`
- `DELETE /api/v1/user/exchange-bindings/{exchange}`
- `GET /api/v1/user/gas-fee`
- `POST /api/v1/user/gas-fee/deposits`
- `GET /api/v1/user/layers`
- `GET /api/v1/user/history/trades`
- `POST /api/v1/admin/auth/login`
- `GET /api/v1/admin/auth/me`
- `POST /api/v1/admin/auth/logout`
- `POST /api/v1/admin/auth/2fa/setup`
- `POST /api/v1/admin/auth/2fa/verify`
- `POST /api/v1/admin/auth/2fa/disable`
- `GET /api/v1/admin/overview`
- `POST /api/v1/admin/gas-fee/preview`
- `GET /api/v1/admin/gas-fee/deposits?status=pending`
- `PATCH /api/v1/admin/gas-fee/deposits/{deposit_id}/status`
- `POST /api/v1/admin/signals`
- `POST /api/v1/admin/executions/{job_id}/retry`
- `GET /api/v1/admin/signals/{signal_id}/executions`
- `GET /api/v1/admin/reconciliation-events?status=open`
- `PATCH /api/v1/admin/reconciliation-events/{event_id}/resolve`
- `POST /api/v1/internal/execution-results`

## Auth and onboarding flow

Register creates an inactive email verification challenge. Passwords are stored as bcrypt hashes, OTP codes are stored as bcrypt hashes, and session tokens are returned once while only their hashes are stored.

Example register:

```json
{
  "name": "Mautrade User",
  "email": "user@example.com",
  "password": "strong-password",
  "confirm_password": "strong-password"
}
```

Response includes `otpRequired: true`. In non-production environments only, `devOtp` is returned to unblock local frontend development until SMTP delivery is wired.

Example OTP verification:

```json
{
  "email": "user@example.com",
  "purpose": "register_verify",
  "code": "123456"
}
```

Use the returned `session.token` as a bearer token:

```txt
Authorization: Bearer <session token>
```

When PostgreSQL is configured, these user dashboard endpoints are scoped to the bearer session user and return `401` without a valid token:

- `GET /api/v1/user/stats`
- `GET /api/v1/user/exchange-bindings`
- `GET /api/v1/user/gas-fee`
- `GET /api/v1/user/layers`
- `GET /api/v1/user/history/trades`

Example onboarding completion:

```json
{
  "country_code": "ID",
  "timezone": "Asia/Jakarta",
  "age": 24,
  "exchange_preferences": ["binance", "okx", "bybit", "tokocrypto"],
  "gas_fee_deposit_amount": "500",
  "gas_fee_asset": "USDT",
  "tx_id": "local-demo-tx"
}
```

Onboarding writes user profile data, selected exchange preferences, and a pending Gas Fee deposit record. The database enforces the 500 USDT minimum.

## Admin auth flow

Admin dashboard endpoints require a separate admin bearer token when PostgreSQL is configured. Bootstrap the first admin account by setting these environment variables before startup:

```txt
ADMIN_BOOTSTRAP_EMAIL=<admin email>
ADMIN_BOOTSTRAP_PASSWORD=<strong password, at least 12 chars>
ADMIN_BOOTSTRAP_NAME=Mautrade Super Admin
ADMIN_BOOTSTRAP_ROLE=super_admin
```

The bootstrap is idempotent: if the email already exists, the API keeps the existing admin and does not overwrite the password.

Admin login:

```json
{
  "email": "<admin email>",
  "password": "<admin password>"
}
```

Use the returned `session.token` on admin dashboard calls:

```txt
Authorization: Bearer <admin session token>
```

Admin 2FA is TOTP-based. Setup requires an existing admin session and returns the secret plus `otpauthUri` once so the admin UI can render a QR code locally:

```txt
POST /api/v1/admin/auth/2fa/setup
```

Verify the authenticator code to enable 2FA:

```json
{
  "code": "123456"
}
```

After 2FA is enabled, admin login with only email/password returns `202 Accepted` with `otpRequired: true`; submit the same login request with `otp_code` or `otpCode` to receive a session token:

```json
{
  "email": "<admin email>",
  "password": "<admin password>",
  "otp_code": "123456"
}
```

Disabling 2FA also requires a valid admin session plus a current TOTP code:

```txt
POST /api/v1/admin/auth/2fa/disable
```

```json
{
  "code": "123456"
}
```

TOTP secrets are encrypted at rest with the same AES-GCM envelope used for exchange credentials. Do not log or persist the plaintext setup secret in the frontend.

Protected admin endpoints include overview, gas fee preview/deposit approval, signal creation, execution retry, signal execution monitoring, and reconciliation actions. Older request bodies may still include `admin_id`, but if present it must match the authenticated admin session.

## Gas Fee account flow

`GET /api/v1/user/gas-fee` returns the authenticated user's confirmed gas fee balance, pending deposits, net gas fee movement, fees used, rebates, and recent mixed history from `gas_fee_deposits` plus `gas_fee_ledger`.

Balance is derived from source-of-truth rows:

```txt
confirmed deposits - net gas_fee_ledger movement
```

That means a profit-share gas fee decreases balance, while a loss rebate increases balance because the ledger amount is negative.

Example user deposit request:

```json
{
  "amount": "500",
  "asset": "USDT",
  "tx_id": "user-chain-tx-id"
}
```

`POST /api/v1/user/gas-fee/deposits` always creates a pending deposit and requires at least `500` USDT plus a TX ID. The server supplies the configured deposit address through `GAS_FEE_DEPOSIT_ADDRESS`.

Admin deposit queue:

- `GET /api/v1/admin/gas-fee/deposits?status=pending` lists pending deposits; use `status=all`, `confirmed`, or `rejected` to change the filter.
- `PATCH /api/v1/admin/gas-fee/deposits/{deposit_id}/status` confirms or rejects a pending deposit and writes an audit log.

Example admin confirmation:

```json
{
  "status": "confirmed",
  "resolution_note": "TX confirmed on-chain."
}
```

## Exchange binding flow

Exchange API credentials are encrypted with AES-GCM before being stored in PostgreSQL. API secret and passphrase are never returned by the API; user-facing responses only include a masked API key plus metadata. One user can only have one binding per exchange. Rebinding the same exchange replaces the encrypted credential on that binding instead of creating duplicates.

Set `EXCHANGE_CREDENTIAL_KEY` in production. Local development can run without it, but production startup rejects a missing key.

Example bind request:

```json
{
  "exchange": "okx",
  "api_key": "okx-api-key",
  "api_secret": "okx-api-secret",
  "api_passphrase": "okx-passphrase",
  "permission_scope": "trade_only"
}
```

Supported exchanges for this v1 backend surface are `binance`, `okx`, `bybit`, and `tokocrypto`. OKX requires `api_passphrase`.

Status update example:

```json
{
  "status": "revoked"
}
```

`DELETE /api/v1/user/exchange-bindings/{exchange}` is a soft revoke, not a hard delete, so old Layer Ledger and audit relationships stay intact.

Example buy signal:

```json
{
  "type": "buy",
  "symbol": "BTC/USDT",
  "allocation_pct": "10",
  "idempotency_key": "admin-demo-btc-buy-20260719-001"
}
```

Example sell signal:

```json
{
  "type": "sell",
  "symbol": "BTC/USDT",
  "layer_number": 1,
  "sell_pct": "100",
  "idempotency_key": "admin-demo-btc-sell-layer-1-20260719-001"
}
```

## Core subjects

- `execution.buy.request`
- `execution.sell.request`
- `execution.result`
- `execution.dlq`

`execution.buy.request` and `execution.sell.request` live in the `EXECUTION` JetStream work queue. `execution.result` and `execution.dlq` live in the `EXECUTION_RESULTS` stream so the Go API can consume Rust execution reports durably.

Before a request reaches `EXECUTION`, the Go API checks the latest `exchange_balance_snapshots` row:

- Buy checks free quote balance (`USDT` by default). If there is no usable quote balance, the execution job is saved as `skipped`.
- Sell checks free base balance for the targeted Layer. If the user has less base asset than the Layer quantity being sold, the Layer is marked `orphaned`, a `reconciliation_events` row is opened, and no Rust order is published.
- Sell dispatch locks the target Layer with `FOR UPDATE`, takes a PostgreSQL advisory transaction lock for that Layer, and checks existing in-flight sell jobs before creating a new one. A partial unique index on `execution_jobs(layer_id)` blocks more than one queued/published/running sell execution for the same Layer.

Admin monitoring endpoints:

- `GET /api/v1/admin/signals/{signal_id}/executions` returns the master signal, execution job status counts, each per-user job, any normalized execution report, and gas fee ledger details.
- `GET /api/v1/admin/reconciliation-events?status=open` returns open balance/orphan reconciliation events. Use `status=all`, `resolved`, or `ignored` to change the filter.
- `POST /api/v1/admin/executions/{job_id}/retry` retries `failed` or `dead_letter` execution jobs by creating a new request attempt id while keeping the trading idempotency key unchanged. Body is optional because the admin session supplies the actor id.

```json
{}
```

- `PATCH /api/v1/admin/reconciliation-events/{event_id}/resolve` marks an event as resolved and writes an admin audit log. Optional body:

```json
{
  "resolution_note": "Balance manually reconciled after exchange sync."
}
```

Example execution result:

```json
{
  "request_id": "018fbd2e-7b46-7cc0-98c4-89e6f6dc0c22",
  "idempotency_key": "signal:user:binding:buy",
  "master_signal_id": "018fbd2e-7b46-7cc0-98c4-89e6f6dc0c22",
  "user_id": "018fbd2e-7b46-7cc0-98c4-89e6f6dc0c22",
  "exchange": "binance",
  "symbol": "BTC/USDT",
  "side": "buy",
  "status": "success",
  "filled_quantity": "0.015000000000000000",
  "fill_price": "62450.000000000000000000",
  "fill_value_quote": "936.750000000000000000",
  "exchange_fee": "0.000015000000000000",
  "exchange_order_id": "paper-binance-001",
  "executed_at": "2026-07-19T07:00:00Z"
}
```

## Precision rule

Financial values use `github.com/MeViksry/qdecimal` in Go and `NUMERIC(36,18)` in PostgreSQL. Do not use floats for trading, fee, quantity, or price calculations.
