package store

import (
	"context"
	"fmt"
	"time"
)

type AdminActiveSignalView struct {
	ID                  string    `json:"id"`
	Symbol              string    `json:"symbol"`
	Type                string    `json:"type"`
	LayerNumber         int       `json:"layerNumber"`
	AllocationPct       float64   `json:"allocationPct"`
	Status              string    `json:"status"`
	CreatedAt           time.Time `json:"createdAt"`
	TotalLayers         int       `json:"totalLayers"`
	ActiveUsers         int       `json:"activeUsers"`
	TotalVolumeQuote    float64   `json:"totalVolumeQuote"`
	RemainingQuantity   float64   `json:"remainingQuantity"`
	RemainingValueQuote float64   `json:"remainingValueQuote"`
}

type AdminOpenOrderView struct {
	ID        string    `json:"id"`
	Symbol    string    `json:"symbol"`
	Action    string    `json:"action"`
	Quantity  float64   `json:"quantity"`
	Price     float64   `json:"price"`
	Status    string    `json:"status"`
	Exchange  string    `json:"exchange"`
	CreatedAt time.Time `json:"createdAt"`
}

func (s *DashboardStore) AdminListActiveSignals(ctx context.Context, limit, offset int) ([]AdminActiveSignalView, error) {
	const query = `
		SELECT
			MIN(l.id::text) AS id,
			l.symbol,
			'buy' AS type,
			l.layer_number,
			COALESCE(MAX(l.allocation_pct), 0)::float8 AS allocation_pct,
			CASE
				WHEN BOOL_OR(l.status = 'partial') THEN 'partial'
				ELSE 'open'
			END AS status,
			MAX(l.opened_at) AS created_at,
			COUNT(l.id)::int AS total_layers,
			COUNT(DISTINCT l.user_id)::int AS active_users,
			COALESCE(SUM(l.entry_value_quote), 0)::float8 AS total_volume_quote,
			COALESCE(SUM(l.remaining_quantity), 0)::float8 AS remaining_quantity,
			COALESCE(SUM(l.remaining_quantity * l.entry_price), 0)::float8 AS remaining_value_quote
		FROM layers l
		JOIN users u ON u.id = l.user_id
		JOIN exchange_bindings b ON b.id = l.exchange_binding_id
		WHERE l.status IN ('open', 'partial')
			AND u.status = 'active'
			AND b.status = 'active'
		GROUP BY l.symbol, l.layer_number
		ORDER BY l.symbol ASC, l.layer_number ASC
		LIMIT $1 OFFSET $2
	`
	rows, err := s.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("store: admin list active signals: %w", err)
	}
	defer rows.Close()

	var signals []AdminActiveSignalView
	for rows.Next() {
		var sig AdminActiveSignalView
		if err := rows.Scan(
			&sig.ID, &sig.Symbol, &sig.Type, &sig.LayerNumber, &sig.AllocationPct, &sig.Status, &sig.CreatedAt,
			&sig.TotalLayers, &sig.ActiveUsers, &sig.TotalVolumeQuote, &sig.RemainingQuantity, &sig.RemainingValueQuote,
		); err != nil {
			return nil, fmt.Errorf("store: scan active signal: %w", err)
		}
		signals = append(signals, sig)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if signals == nil {
		signals = []AdminActiveSignalView{}
	}
	return signals, nil
}

func (s *DashboardStore) AdminListOpenOrders(ctx context.Context, limit, offset int) ([]AdminOpenOrderView, error) {
	const query = `
		SELECT
			le.id, le.symbol, le.action, le.quantity, le.price, le.status, eb.exchange_name, le.created_at
		FROM layer_executions le
		JOIN exchange_bindings eb ON eb.id = le.exchange_binding_id
		WHERE le.status IN ('pending', 'partial')
		ORDER BY le.created_at DESC
		LIMIT $1 OFFSET $2
	`
	rows, err := s.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("store: admin list open orders: %w", err)
	}
	defer rows.Close()

	var orders []AdminOpenOrderView
	for rows.Next() {
		var o AdminOpenOrderView
		if err := rows.Scan(
			&o.ID, &o.Symbol, &o.Action, &o.Quantity, &o.Price, &o.Status, &o.Exchange, &o.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("store: scan open order: %w", err)
		}
		orders = append(orders, o)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if orders == nil {
		orders = []AdminOpenOrderView{}
	}
	return orders, nil
}
