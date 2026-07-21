package store

import (
	"context"
	"fmt"
	"time"
)

type AdminActiveSignalView struct {
	ID               string    `json:"id"`
	Symbol           string    `json:"symbol"`
	Type             string    `json:"type"`
	AllocationPct    float64   `json:"allocationPct"`
	Status           string    `json:"status"`
	CreatedAt        time.Time `json:"createdAt"`
	TotalLayers      int       `json:"totalLayers"`
	TotalVolumeQuote float64   `json:"totalVolumeQuote"`
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
			ms.id, ms.symbol, ms.type, ms.allocation_pct, ms.status, ms.created_at,
			COUNT(l.id) as total_layers,
			COALESCE(SUM(l.entry_value_quote), 0) as total_volume_quote
		FROM master_signals ms
		LEFT JOIN layers l ON l.master_signal_id = ms.id
		WHERE ms.status IN ('draft', 'dispatching', 'completed')
		GROUP BY ms.id
		ORDER BY ms.created_at DESC
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
			&sig.ID, &sig.Symbol, &sig.Type, &sig.AllocationPct, &sig.Status, &sig.CreatedAt,
			&sig.TotalLayers, &sig.TotalVolumeQuote,
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
