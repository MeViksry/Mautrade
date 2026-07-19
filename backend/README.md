# Mautrade Backend

Go control-plane backend for Mautrade copy trading.

## What is here now

- Go HTTP API skeleton under `cmd/api`.
- `quuid` UUIDv7 ID package usage through `internal/domain/id`.
- `qdecimal` Gas Fee engine under `internal/domain/gasfee`.
- PostgreSQL schema in `migrations/001_init.sql`.
- NATS JetStream stream setup for distributed Rust execution workers.
- PostgreSQL-backed user/admin dashboard reads with mock fallback only when `DATABASE_URL` is empty.
- Admin signal creation endpoint that writes `master_signals`, creates per-user `execution_jobs`, and publishes idempotent jobs to NATS when available.
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
- `GET /api/v1/user/stats`
- `GET /api/v1/user/exchange-bindings`
- `GET /api/v1/user/layers`
- `GET /api/v1/user/history/trades`
- `GET /api/v1/admin/overview`
- `POST /api/v1/admin/gas-fee/preview`
- `POST /api/v1/admin/signals`

Example buy signal:

```json
{
  "admin_id": "018fbd2e-7b46-7cc0-98c4-89e6f6dc0c22",
  "type": "buy",
  "symbol": "BTC/USDT",
  "allocation_pct": "10",
  "idempotency_key": "admin-demo-btc-buy-20260719-001"
}
```

Example sell signal:

```json
{
  "admin_id": "018fbd2e-7b46-7cc0-98c4-89e6f6dc0c22",
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

## Precision rule

Financial values use `github.com/MeViksry/qdecimal` in Go and `NUMERIC(36,18)` in PostgreSQL. Do not use floats for trading, fee, quantity, or price calculations.
