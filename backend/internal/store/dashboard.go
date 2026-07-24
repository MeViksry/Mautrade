package store

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DashboardStore struct {
	db *pgxpool.Pool
}

func NewDashboardStore(db *pgxpool.Pool) *DashboardStore {
	return &DashboardStore{db: db}
}

func (s *DashboardStore) Ready() bool {
	return s != nil && s.db != nil
}

type UserStats struct {
	TotalBalance      string `json:"totalBalance"`
	RealizedProfit    string `json:"realizedProfit"`
	TotalGasFeePaid   string `json:"totalGasFeePaid"`
	ActiveLayersCount int64  `json:"activeLayersCount"`
	PrecisionPolicy     string  `json:"precisionPolicy"`
	DefaultCurrency     string  `json:"defaultCurrency"`
	GasFeeShareRate     string  `json:"gasFeeShareRate"`
	GasFeeDepositStatus string  `json:"gasFeeDepositStatus"`
	GasFeeDepositTxID   *string `json:"gasFeeDepositTxId,omitempty"`
}

func (s *DashboardStore) UserStats(ctx context.Context, userID, defaultCurrency, gasFeeShareRate string) (UserStats, error) {
	const query = `
WITH latest_balances AS (
  SELECT DISTINCT ON (exchange_binding_id, asset)
    s.exchange_binding_id,
    s.asset,
    s.free_amount,
    s.locked_amount
  FROM exchange_balance_snapshots s
  JOIN exchange_bindings b ON b.id = s.exchange_binding_id
  WHERE b.user_id = $2::uuid
    AND s.asset = $1
  ORDER BY s.exchange_binding_id, s.asset, s.captured_at DESC
),
balance_sum AS (
  SELECT COALESCE(SUM(free_amount + locked_amount), 0)::text AS total_balance
  FROM latest_balances
),
gas_sum AS (
  SELECT
    COALESCE(SUM(gross_pnl), 0)::text AS realized_profit,
    COALESCE(SUM(gas_fee_amount), 0)::text AS total_gas_fee_paid
  FROM gas_fee_ledger
  WHERE user_id = $2::uuid
),
layer_sum AS (
  SELECT COUNT(*) AS active_layers_count
  FROM layers
  WHERE user_id = $2::uuid
    AND status IN ('open', 'partial')
),
latest_deposit AS (
  SELECT status, tx_id
  FROM gas_fee_deposits
  WHERE user_id = $2::uuid
  ORDER BY created_at DESC
  LIMIT 1
)
SELECT balance_sum.total_balance, gas_sum.realized_profit, gas_sum.total_gas_fee_paid, layer_sum.active_layers_count,
       COALESCE((SELECT status FROM latest_deposit), 'none') AS gas_fee_deposit_status,
       (SELECT tx_id FROM latest_deposit) AS gas_fee_deposit_tx_id
FROM balance_sum, gas_sum, layer_sum`

	stats := UserStats{
		PrecisionPolicy: "qdecimal",
		DefaultCurrency: defaultCurrency,
		GasFeeShareRate: gasFeeShareRate,
	}
	if err := s.db.QueryRow(ctx, query, defaultCurrency, userID).Scan(
		&stats.TotalBalance,
		&stats.RealizedProfit,
		&stats.TotalGasFeePaid,
		&stats.ActiveLayersCount,
		&stats.GasFeeDepositStatus,
		&stats.GasFeeDepositTxID,
	); err != nil {
		return UserStats{}, fmt.Errorf("store: user stats: %w", err)
	}
	return stats, nil
}

type ExchangeBindingView struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	Status     string     `json:"status"`
	LastSynced *time.Time `json:"lastSynced"`
	Balance    string     `json:"balance"`
	HasAPI     bool       `json:"hasApi"`
}

func (s *DashboardStore) ExchangeBindings(ctx context.Context, defaultCurrency string) ([]ExchangeBindingView, error) {
	const query = `
WITH latest_balances AS (
  SELECT DISTINCT ON (exchange_binding_id, asset)
    exchange_binding_id,
    asset,
    free_amount + locked_amount AS amount
  FROM exchange_balance_snapshots
  WHERE asset = $1
  ORDER BY exchange_binding_id, asset, captured_at DESC
)
SELECT
  b.id::text,
  CASE b.exchange_name
    WHEN 'binance' THEN 'Binance'
    WHEN 'okx' THEN 'OKX'
    WHEN 'bybit' THEN 'Bybit'
    WHEN 'tokocrypto' THEN 'Tokocrypto'
    ELSE b.exchange_name
  END AS name,
  CASE WHEN b.status = 'active' THEN 'connected' ELSE 'disconnected' END AS status,
  b.last_verified_at,
  COALESCE(lb.amount, 0)::text AS balance,
  b.status <> 'revoked' AS has_api
FROM exchange_bindings b
LEFT JOIN latest_balances lb ON lb.exchange_binding_id = b.id
ORDER BY b.created_at ASC`

	rows, err := s.db.Query(ctx, query, defaultCurrency)
	if err != nil {
		return nil, fmt.Errorf("store: exchange bindings: %w", err)
	}
	defer rows.Close()

	var bindings []ExchangeBindingView
	for rows.Next() {
		var binding ExchangeBindingView
		if err := rows.Scan(&binding.ID, &binding.Name, &binding.Status, &binding.LastSynced, &binding.Balance, &binding.HasAPI); err != nil {
			return nil, fmt.Errorf("store: scan exchange binding: %w", err)
		}
		bindings = append(bindings, binding)
	}
	if bindings == nil {
		bindings = []ExchangeBindingView{}
	}
	return bindings, rows.Err()
}

type LayerView struct {
	ID               string `json:"id"`
	Pair             string `json:"pair"`
	EntryPrice       string `json:"entryPrice"`
	CurrentPrice     string `json:"currentPrice"`
	AllocationPct    string `json:"allocationPct"`
	AllocatedUSDT    string `json:"allocatedUsdt"`
	UnrealizedPnL    string `json:"unrealizedPnl"`
	UnrealizedPnLPct string `json:"unrealizedPnlPct"`
	Status           string `json:"status"`
	OpenedAt         string `json:"openedAt"`
}

func (s *DashboardStore) ActiveLayers(ctx context.Context, userID string) ([]LayerView, error) {
	const query = `
WITH latest_prices AS (
  SELECT DISTINCT ON (symbol)
    symbol,
    price_quote
  FROM market_prices
  ORDER BY symbol, captured_at DESC
)
SELECT
  l.id::text,
  l.symbol,
  l.entry_price::text,
  COALESCE(mp.price_quote, l.entry_price)::text AS current_price,
  l.allocation_pct::text,
  l.entry_value_quote::text,
  ((COALESCE(mp.price_quote, l.entry_price) - l.entry_price) * l.remaining_quantity)::text AS unrealized_pnl,
  CASE
    WHEN l.entry_value_quote = 0 THEN '0'
    ELSE ((((COALESCE(mp.price_quote, l.entry_price) - l.entry_price) * l.remaining_quantity) / l.entry_value_quote) * 100)::text
  END AS unrealized_pnl_pct,
  l.status,
  l.opened_at::text
FROM layers l
LEFT JOIN latest_prices mp ON mp.symbol = l.symbol
WHERE l.user_id = $1::uuid
  AND l.status IN ('open', 'partial')
ORDER BY l.opened_at DESC
LIMIT 100`

	rows, err := s.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("store: active layers: %w", err)
	}
	defer rows.Close()

	var layers []LayerView
	for rows.Next() {
		var layer LayerView
		if err := rows.Scan(
			&layer.ID,
			&layer.Pair,
			&layer.EntryPrice,
			&layer.CurrentPrice,
			&layer.AllocationPct,
			&layer.AllocatedUSDT,
			&layer.UnrealizedPnL,
			&layer.UnrealizedPnLPct,
			&layer.Status,
			&layer.OpenedAt,
		); err != nil {
			return nil, fmt.Errorf("store: scan active layer: %w", err)
		}
		layers = append(layers, layer)
	}
	if layers == nil {
		layers = []LayerView{}
	}
	return layers, rows.Err()
}

type TradeHistoryView struct {
	ID        string `json:"id"`
	Pair      string `json:"pair"`
	ExitPrice string `json:"exitPrice"`
	PnL       string `json:"pnl"`
	GasFee    string `json:"gasFee"`
	ClosedAt  string `json:"closedAt"`
}

func (s *DashboardStore) TradeHistory(ctx context.Context, userID string) ([]TradeHistoryView, error) {
	const query = `
SELECT
  l.id::text,
  l.symbol,
  e.price::text,
  g.gross_pnl::text,
  g.gas_fee_amount::text,
  COALESCE(e.executed_at, g.calculated_at)::text
FROM gas_fee_ledger g
JOIN layers l ON l.id = g.layer_id
JOIN layer_executions e ON e.id = g.execution_id
WHERE g.user_id = $1::uuid
ORDER BY COALESCE(e.executed_at, g.calculated_at) DESC
LIMIT 100`

	rows, err := s.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("store: trade history: %w", err)
	}
	defer rows.Close()

	var history []TradeHistoryView
	for rows.Next() {
		var item TradeHistoryView
		if err := rows.Scan(&item.ID, &item.Pair, &item.ExitPrice, &item.PnL, &item.GasFee, &item.ClosedAt); err != nil {
			return nil, fmt.Errorf("store: scan trade history: %w", err)
		}
		history = append(history, item)
	}
	if history == nil {
		history = []TradeHistoryView{}
	}
	return history, rows.Err()
}

type AdminOverview struct {
	RegisteredUsers     int64  `json:"registeredUsers"`
	ActiveUsers         int64  `json:"activeUsers"`
	OpenLayers          int64  `json:"openLayers"`
	EstimatedAUM        string `json:"estimatedAUM"`
	GasFeeRevenueToday  string `json:"gasFeeRevenueToday"`
	OrphanedLayers      int64  `json:"orphanedLayers"`
	ExecutionQueueState string `json:"executionQueueState"`
}

func (s *DashboardStore) AdminOverview(ctx context.Context, defaultCurrency string) (AdminOverview, error) {
	const query = `
WITH latest_balances AS (
  SELECT DISTINCT ON (exchange_binding_id, asset)
    exchange_binding_id,
    asset,
    free_amount,
    locked_amount
  FROM exchange_balance_snapshots
  ORDER BY exchange_binding_id, asset, captured_at DESC
)
SELECT
  (SELECT COUNT(*) FROM users)::bigint,
  (SELECT COUNT(*) FROM users WHERE status = 'active')::bigint,
  (SELECT COUNT(*) FROM layers WHERE status IN ('open', 'partial'))::bigint,
  (SELECT COALESCE(SUM(free_amount + locked_amount) FILTER (WHERE asset = $1), 0)::text FROM latest_balances),
  (SELECT COALESCE(SUM(gas_fee_amount), 0)::text FROM gas_fee_ledger WHERE calculated_at >= date_trunc('day', now())),
  (SELECT COUNT(*) FROM layers WHERE status = 'orphaned')::bigint`

	overview := AdminOverview{ExecutionQueueState: "ready"}
	if err := s.db.QueryRow(ctx, query, defaultCurrency).Scan(
		&overview.RegisteredUsers,
		&overview.ActiveUsers,
		&overview.OpenLayers,
		&overview.EstimatedAUM,
		&overview.GasFeeRevenueToday,
		&overview.OrphanedLayers,
	); err != nil && err != pgx.ErrNoRows {
		return AdminOverview{}, fmt.Errorf("store: admin overview: %w", err)
	}
	return overview, nil
}
