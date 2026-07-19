package store

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/MeViksry/Mautrade/backend/internal/domain/id"
	"github.com/jackc/pgx/v5"
)

type CreateSignalParams struct {
	AdminID        string
	Type           string
	Symbol         string
	LayerNumber    *int
	AllocationPct  string
	SellPct        string
	IdempotencyKey string
	DefaultAsset   string
	CreatedAt      time.Time
}

type SignalDispatch struct {
	SignalID       string               `json:"signalId"`
	Status         string               `json:"status"`
	IdempotencyKey string               `json:"idempotencyKey"`
	JobsCreated    int                  `json:"jobsCreated"`
	Jobs           []ExecutionJobRecord `json:"jobs"`
}

type ExecutionJobRecord struct {
	ID                string           `json:"id"`
	MasterSignalID    string           `json:"masterSignalId"`
	LayerID           string           `json:"layerId,omitempty"`
	UserID            string           `json:"userId"`
	ExchangeBindingID string           `json:"exchangeBindingId"`
	Exchange          string           `json:"exchange"`
	Subject           string           `json:"subject"`
	Payload           ExecutionPayload `json:"payload"`
	IdempotencyKey    string           `json:"idempotencyKey"`
}

type ExecutionPayload struct {
	ID             string `json:"id"`
	IdempotencyKey string `json:"idempotency_key"`
	MasterSignalID string `json:"master_signal_id"`
	UserID         string `json:"user_id"`
	LayerID        string `json:"layer_id,omitempty"`
	Exchange       string `json:"exchange"`
	Symbol         string `json:"symbol"`
	Side           string `json:"side"`
	Quantity       string `json:"quantity,omitempty"`
	QuoteValue     string `json:"quote_value,omitempty"`
	CreatedAt      string `json:"created_at"`
}

type eligibleBuyBinding struct {
	UserID            string
	ExchangeBindingID string
	Exchange          string
	QuoteValue        string
}

type eligibleSellLayer struct {
	UserID            string
	ExchangeBindingID string
	Exchange          string
	LayerID           string
	Quantity          string
}

func (s *DashboardStore) CreateSignalDispatch(ctx context.Context, params CreateSignalParams) (SignalDispatch, error) {
	if !s.Ready() {
		return SignalDispatch{}, fmt.Errorf("store: signal dispatch requires postgres")
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return SignalDispatch{}, fmt.Errorf("store: begin signal dispatch: %w", err)
	}
	defer tx.Rollback(ctx)

	signalID, err := id.New()
	if err != nil {
		return SignalDispatch{}, err
	}
	signalIDText := signalID.String()

	preview := map[string]any{
		"type":            params.Type,
		"symbol":          params.Symbol,
		"layer_number":    params.LayerNumber,
		"allocation_pct":  params.AllocationPct,
		"sell_pct":        params.SellPct,
		"default_asset":   params.DefaultAsset,
		"requested_at":    params.CreatedAt.UTC().Format(time.RFC3339Nano),
		"execution_stage": "queued",
	}
	previewJSON, err := json.Marshal(preview)
	if err != nil {
		return SignalDispatch{}, fmt.Errorf("store: marshal signal preview: %w", err)
	}

	_, err = tx.Exec(ctx, `
INSERT INTO master_signals (
  id, admin_id, type, symbol, layer_number, allocation_pct, sell_pct,
  status, idempotency_key, preview_snapshot, dispatched_at, created_at
) VALUES (
  $1::uuid, $2::uuid, $3, $4, $5, $6::numeric, $7::numeric,
  'dispatching', $8, $9::jsonb, $10, $10
)`,
		signalIDText,
		params.AdminID,
		params.Type,
		params.Symbol,
		params.LayerNumber,
		params.AllocationPct,
		params.SellPct,
		params.IdempotencyKey,
		string(previewJSON),
		params.CreatedAt.UTC(),
	)
	if err != nil {
		return SignalDispatch{}, fmt.Errorf("store: insert master signal: %w", err)
	}

	var jobs []ExecutionJobRecord
	switch params.Type {
	case "buy":
		jobs, err = s.createBuyJobs(ctx, tx, signalIDText, params)
	case "sell":
		jobs, err = s.createSellJobs(ctx, tx, signalIDText, params)
	default:
		return SignalDispatch{}, fmt.Errorf("store: unsupported signal type %q", params.Type)
	}
	if err != nil {
		return SignalDispatch{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return SignalDispatch{}, fmt.Errorf("store: commit signal dispatch: %w", err)
	}

	return SignalDispatch{
		SignalID:       signalIDText,
		Status:         "dispatching",
		IdempotencyKey: params.IdempotencyKey,
		JobsCreated:    len(jobs),
		Jobs:           jobs,
	}, nil
}

func (s *DashboardStore) createBuyJobs(ctx context.Context, tx pgx.Tx, signalID string, params CreateSignalParams) ([]ExecutionJobRecord, error) {
	const query = `
WITH latest_balances AS (
  SELECT DISTINCT ON (exchange_binding_id, asset)
    exchange_binding_id,
    asset,
    free_amount + locked_amount AS balance
  FROM exchange_balance_snapshots
  WHERE asset = $2
  ORDER BY exchange_binding_id, asset, captured_at DESC
)
SELECT
  b.user_id::text,
  b.id::text,
  b.exchange_name,
  ((COALESCE(lb.balance, 0) * $1::numeric) / 100)::text AS quote_value
FROM exchange_bindings b
JOIN users u ON u.id = b.user_id
LEFT JOIN latest_balances lb ON lb.exchange_binding_id = b.id
WHERE b.status = 'active'
  AND u.status = 'active'
  AND ((COALESCE(lb.balance, 0) * $1::numeric) / 100) > 0
ORDER BY b.created_at ASC`

	rows, err := tx.Query(ctx, query, params.AllocationPct, params.DefaultAsset)
	if err != nil {
		return nil, fmt.Errorf("store: query buy eligible bindings: %w", err)
	}
	defer rows.Close()

	var bindings []eligibleBuyBinding
	for rows.Next() {
		var binding eligibleBuyBinding
		if err := rows.Scan(&binding.UserID, &binding.ExchangeBindingID, &binding.Exchange, &binding.QuoteValue); err != nil {
			return nil, fmt.Errorf("store: scan buy binding: %w", err)
		}
		bindings = append(bindings, binding)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: iterate buy bindings: %w", err)
	}

	jobs := make([]ExecutionJobRecord, 0, len(bindings))
	for _, binding := range bindings {
		job, err := newExecutionJob(signalID, "", binding.UserID, binding.ExchangeBindingID, binding.Exchange, params.Symbol, "buy", "", binding.QuoteValue, params.CreatedAt)
		if err != nil {
			return nil, err
		}
		if err := insertExecutionJob(ctx, tx, job); err != nil {
			return nil, err
		}
		jobs = append(jobs, job)
	}
	return jobs, nil
}

func (s *DashboardStore) createSellJobs(ctx context.Context, tx pgx.Tx, signalID string, params CreateSignalParams) ([]ExecutionJobRecord, error) {
	const query = `
SELECT
  l.user_id::text,
  l.exchange_binding_id::text,
  b.exchange_name,
  l.id::text,
  ((l.remaining_quantity * $3::numeric) / 100)::text AS quantity
FROM layers l
JOIN users u ON u.id = l.user_id
JOIN exchange_bindings b ON b.id = l.exchange_binding_id
WHERE l.symbol = $1
  AND l.layer_number = $2
  AND l.status IN ('open', 'partial')
  AND u.status = 'active'
  AND b.status = 'active'
  AND ((l.remaining_quantity * $3::numeric) / 100) > 0
ORDER BY l.opened_at ASC`

	rows, err := tx.Query(ctx, query, params.Symbol, *params.LayerNumber, params.SellPct)
	if err != nil {
		return nil, fmt.Errorf("store: query sell eligible layers: %w", err)
	}
	defer rows.Close()

	var layers []eligibleSellLayer
	for rows.Next() {
		var layer eligibleSellLayer
		if err := rows.Scan(&layer.UserID, &layer.ExchangeBindingID, &layer.Exchange, &layer.LayerID, &layer.Quantity); err != nil {
			return nil, fmt.Errorf("store: scan sell layer: %w", err)
		}
		layers = append(layers, layer)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: iterate sell layers: %w", err)
	}

	jobs := make([]ExecutionJobRecord, 0, len(layers))
	for _, layer := range layers {
		job, err := newExecutionJob(signalID, layer.LayerID, layer.UserID, layer.ExchangeBindingID, layer.Exchange, params.Symbol, "sell", layer.Quantity, "", params.CreatedAt)
		if err != nil {
			return nil, err
		}
		if err := insertExecutionJob(ctx, tx, job); err != nil {
			return nil, err
		}
		jobs = append(jobs, job)
	}
	return jobs, nil
}

func newExecutionJob(signalID, layerID, userID, bindingID, exchange, symbol, side, quantity, quoteValue string, createdAt time.Time) (ExecutionJobRecord, error) {
	jobID, err := id.New()
	if err != nil {
		return ExecutionJobRecord{}, err
	}
	jobIDText := jobID.String()
	idempotencyKey := fmt.Sprintf("%s:%s:%s:%s", signalID, userID, bindingID, side)
	if layerID != "" {
		idempotencyKey = fmt.Sprintf("%s:%s:%s:%s", signalID, userID, layerID, side)
	}

	subject := "execution.buy.request"
	if side == "sell" {
		subject = "execution.sell.request"
	}

	payload := ExecutionPayload{
		ID:             jobIDText,
		IdempotencyKey: idempotencyKey,
		MasterSignalID: signalID,
		UserID:         userID,
		LayerID:        layerID,
		Exchange:       exchange,
		Symbol:         symbol,
		Side:           side,
		Quantity:       quantity,
		QuoteValue:     quoteValue,
		CreatedAt:      createdAt.UTC().Format(time.RFC3339Nano),
	}

	return ExecutionJobRecord{
		ID:                jobIDText,
		MasterSignalID:    signalID,
		LayerID:           layerID,
		UserID:            userID,
		ExchangeBindingID: bindingID,
		Exchange:          exchange,
		Subject:           subject,
		Payload:           payload,
		IdempotencyKey:    idempotencyKey,
	}, nil
}

func insertExecutionJob(ctx context.Context, tx pgx.Tx, job ExecutionJobRecord) error {
	payloadJSON, err := json.Marshal(job.Payload)
	if err != nil {
		return fmt.Errorf("store: marshal execution job payload: %w", err)
	}

	var layerID any
	if job.LayerID != "" {
		layerID = job.LayerID
	}

	_, err = tx.Exec(ctx, `
INSERT INTO execution_jobs (
  id, master_signal_id, layer_id, user_id, exchange_binding_id, subject, payload, status, idempotency_key
) VALUES (
  $1::uuid, $2::uuid, $3::uuid, $4::uuid, $5::uuid, $6, $7::jsonb, 'queued', $8
)`,
		job.ID,
		job.MasterSignalID,
		layerID,
		job.UserID,
		job.ExchangeBindingID,
		job.Subject,
		string(payloadJSON),
		job.IdempotencyKey,
	)
	if err != nil {
		return fmt.Errorf("store: insert execution job: %w", err)
	}
	return nil
}

func (s *DashboardStore) MarkExecutionJobPublished(ctx context.Context, jobID string) error {
	if !s.Ready() {
		return nil
	}
	_, err := s.db.Exec(ctx, `
UPDATE execution_jobs
SET status = 'published', updated_at = now()
WHERE id = $1::uuid`, jobID)
	if err != nil {
		return fmt.Errorf("store: mark execution job published: %w", err)
	}
	return nil
}

func (s *DashboardStore) MarkExecutionJobPublishFailed(ctx context.Context, jobID, lastError string) error {
	if !s.Ready() {
		return nil
	}
	_, err := s.db.Exec(ctx, `
UPDATE execution_jobs
SET status = 'failed', last_error = $2, updated_at = now()
WHERE id = $1::uuid`, jobID, lastError)
	if err != nil {
		return fmt.Errorf("store: mark execution job publish failed: %w", err)
	}
	return nil
}
