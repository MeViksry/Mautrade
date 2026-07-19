package store

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/MeViksry/Mautrade/backend/internal/domain/id"
	"github.com/MeViksry/qdecimal"
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
	JobsSkipped    int                  `json:"jobsSkipped"`
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
	AvailableQuote    string
	QuoteValue        string
}

type eligibleSellLayer struct {
	UserID            string
	ExchangeBindingID string
	Exchange          string
	LayerID           string
	Quantity          string
	AvailableBase     string
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
	var skipped int
	switch params.Type {
	case "buy":
		jobs, skipped, err = s.createBuyJobs(ctx, tx, signalIDText, params)
	case "sell":
		jobs, skipped, err = s.createSellJobs(ctx, tx, signalIDText, params)
	default:
		return SignalDispatch{}, fmt.Errorf("store: unsupported signal type %q", params.Type)
	}
	if err != nil {
		return SignalDispatch{}, err
	}

	status := "dispatching"
	if len(jobs) == 0 {
		status = "completed"
		if _, err := tx.Exec(ctx, `
UPDATE master_signals
SET status = 'completed', completed_at = $2
WHERE id = $1::uuid`, signalIDText, params.CreatedAt.UTC()); err != nil {
			return SignalDispatch{}, fmt.Errorf("store: complete signal without queued jobs: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return SignalDispatch{}, fmt.Errorf("store: commit signal dispatch: %w", err)
	}

	return SignalDispatch{
		SignalID:       signalIDText,
		Status:         status,
		IdempotencyKey: params.IdempotencyKey,
		JobsCreated:    len(jobs) + skipped,
		JobsSkipped:    skipped,
		Jobs:           jobs,
	}, nil
}

func (s *DashboardStore) createBuyJobs(ctx context.Context, tx pgx.Tx, signalID string, params CreateSignalParams) ([]ExecutionJobRecord, int, error) {
	const query = `
WITH latest_balances AS (
  SELECT DISTINCT ON (exchange_binding_id, asset)
    exchange_binding_id,
    asset,
    free_amount AS available_balance
  FROM exchange_balance_snapshots
  WHERE asset = $2
  ORDER BY exchange_binding_id, asset, captured_at DESC
)
SELECT
  b.user_id::text,
  b.id::text,
  b.exchange_name,
  COALESCE(lb.available_balance, 0)::text AS available_quote,
  ((COALESCE(lb.available_balance, 0) * $1::numeric) / 100)::text AS quote_value
FROM exchange_bindings b
JOIN users u ON u.id = b.user_id
LEFT JOIN latest_balances lb ON lb.exchange_binding_id = b.id
WHERE b.status = 'active'
  AND u.status = 'active'
ORDER BY b.created_at ASC`

	rows, err := tx.Query(ctx, query, params.AllocationPct, params.DefaultAsset)
	if err != nil {
		return nil, 0, fmt.Errorf("store: query buy eligible bindings: %w", err)
	}
	defer rows.Close()

	var bindings []eligibleBuyBinding
	for rows.Next() {
		var binding eligibleBuyBinding
		if err := rows.Scan(&binding.UserID, &binding.ExchangeBindingID, &binding.Exchange, &binding.AvailableQuote, &binding.QuoteValue); err != nil {
			return nil, 0, fmt.Errorf("store: scan buy binding: %w", err)
		}
		bindings = append(bindings, binding)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("store: iterate buy bindings: %w", err)
	}

	jobs := make([]ExecutionJobRecord, 0, len(bindings))
	skipped := 0
	for _, binding := range bindings {
		job, err := newExecutionJob(signalID, "", binding.UserID, binding.ExchangeBindingID, binding.Exchange, params.Symbol, "buy", "", binding.QuoteValue, params.CreatedAt)
		if err != nil {
			return nil, skipped, err
		}
		hasQuote, err := decimalGreaterThanZero(binding.QuoteValue)
		if err != nil {
			return nil, skipped, err
		}
		if !hasQuote {
			reason := fmt.Sprintf("reconciliation: insufficient %s balance for buy signal; available=%s", params.DefaultAsset, binding.AvailableQuote)
			if err := insertExecutionJobWithStatus(ctx, tx, job, "skipped", reason); err != nil {
				return nil, skipped, err
			}
			if err := insertReconciliationEvent(ctx, tx, reconciliationEventInput{
				UserID:            binding.UserID,
				ExchangeBindingID: binding.ExchangeBindingID,
				MasterSignalID:    signalID,
				EventType:         "buy_insufficient_quote_balance",
				Asset:             params.DefaultAsset,
				RequiredAmount:    binding.QuoteValue,
				AvailableAmount:   binding.AvailableQuote,
				Reason:            reason,
			}); err != nil {
				return nil, skipped, err
			}
			if err := insertNotification(ctx, tx, binding.UserID, "Buy Signal Skipped", reason); err != nil {
				return nil, skipped, err
			}
			skipped++
			continue
		}
		if err := insertExecutionJob(ctx, tx, job); err != nil {
			return nil, skipped, err
		}
		jobs = append(jobs, job)
	}
	return jobs, skipped, nil
}

func (s *DashboardStore) createSellJobs(ctx context.Context, tx pgx.Tx, signalID string, params CreateSignalParams) ([]ExecutionJobRecord, int, error) {
	baseAsset, err := baseAssetFromSymbol(params.Symbol)
	if err != nil {
		return nil, 0, err
	}

	const query = `
WITH latest_balances AS (
  SELECT DISTINCT ON (exchange_binding_id, asset)
    exchange_binding_id,
    asset,
    free_amount AS available_balance
  FROM exchange_balance_snapshots
  WHERE asset = $4
  ORDER BY exchange_binding_id, asset, captured_at DESC
)
SELECT
  l.user_id::text,
  l.exchange_binding_id::text,
  b.exchange_name,
  l.id::text,
  ((l.remaining_quantity * $3::numeric) / 100)::text AS quantity,
  COALESCE(lb.available_balance, 0)::text AS available_base
FROM layers l
JOIN users u ON u.id = l.user_id
JOIN exchange_bindings b ON b.id = l.exchange_binding_id
LEFT JOIN latest_balances lb ON lb.exchange_binding_id = l.exchange_binding_id
WHERE l.symbol = $1
  AND l.layer_number = $2
  AND l.status IN ('open', 'partial')
  AND u.status = 'active'
  AND b.status = 'active'
  AND ((l.remaining_quantity * $3::numeric) / 100) > 0
ORDER BY l.opened_at ASC`

	rows, err := tx.Query(ctx, query, params.Symbol, *params.LayerNumber, params.SellPct, baseAsset)
	if err != nil {
		return nil, 0, fmt.Errorf("store: query sell eligible layers: %w", err)
	}
	defer rows.Close()

	var layers []eligibleSellLayer
	for rows.Next() {
		var layer eligibleSellLayer
		if err := rows.Scan(&layer.UserID, &layer.ExchangeBindingID, &layer.Exchange, &layer.LayerID, &layer.Quantity, &layer.AvailableBase); err != nil {
			return nil, 0, fmt.Errorf("store: scan sell layer: %w", err)
		}
		layers = append(layers, layer)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("store: iterate sell layers: %w", err)
	}

	jobs := make([]ExecutionJobRecord, 0, len(layers))
	skipped := 0
	for _, layer := range layers {
		if err := acquireExecutionScopeLock(ctx, tx, "sell-layer:"+layer.LayerID); err != nil {
			return nil, skipped, err
		}
		currentQuantity, stillOpen, err := lockSellLayerForDispatch(ctx, tx, layer.LayerID, params.SellPct)
		if err != nil {
			return nil, skipped, err
		}
		if stillOpen {
			layer.Quantity = currentQuantity
		}
		job, err := newExecutionJob(signalID, layer.LayerID, layer.UserID, layer.ExchangeBindingID, layer.Exchange, params.Symbol, "sell", layer.Quantity, "", params.CreatedAt)
		if err != nil {
			return nil, skipped, err
		}
		if !stillOpen {
			reason := "concurrency: layer is no longer open for sell dispatch"
			if err := insertExecutionJobWithStatus(ctx, tx, job, "skipped", reason); err != nil {
				return nil, skipped, err
			}
			if err := insertReconciliationEvent(ctx, tx, reconciliationEventInput{
				UserID:            layer.UserID,
				ExchangeBindingID: layer.ExchangeBindingID,
				MasterSignalID:    signalID,
				LayerID:           layer.LayerID,
				EventType:         "sell_layer_not_open",
				Asset:             baseAsset,
				RequiredAmount:    layer.Quantity,
				AvailableAmount:   layer.AvailableBase,
				Reason:            reason,
			}); err != nil {
				return nil, skipped, err
			}
			skipped++
			continue
		}
		inFlightJobID, inFlight, err := inFlightSellExecutionJob(ctx, tx, layer.LayerID)
		if err != nil {
			return nil, skipped, err
		}
		if inFlight {
			reason := fmt.Sprintf("concurrency: layer already has in-flight sell execution job %s", inFlightJobID)
			if err := insertExecutionJobWithStatus(ctx, tx, job, "skipped", reason); err != nil {
				return nil, skipped, err
			}
			if err := insertReconciliationEvent(ctx, tx, reconciliationEventInput{
				UserID:            layer.UserID,
				ExchangeBindingID: layer.ExchangeBindingID,
				MasterSignalID:    signalID,
				LayerID:           layer.LayerID,
				EventType:         "sell_layer_execution_in_flight",
				Asset:             baseAsset,
				RequiredAmount:    layer.Quantity,
				AvailableAmount:   layer.AvailableBase,
				Reason:            reason,
			}); err != nil {
				return nil, skipped, err
			}
			if err := insertNotification(ctx, tx, layer.UserID, "Layer Sell Skipped", reason); err != nil {
				return nil, skipped, err
			}
			skipped++
			continue
		}
		insufficient, err := decimalLessThan(layer.AvailableBase, layer.Quantity)
		if err != nil {
			return nil, skipped, err
		}
		if insufficient {
			reason := fmt.Sprintf("reconciliation: insufficient %s balance for layer sell; required=%s available=%s", baseAsset, layer.Quantity, layer.AvailableBase)
			if err := insertExecutionJobWithStatus(ctx, tx, job, "skipped", reason); err != nil {
				return nil, skipped, err
			}
			if _, err := tx.Exec(ctx, `
UPDATE layers
SET status = 'orphaned',
    orphan_reason = $2,
    updated_at = now()
WHERE id = $1::uuid
  AND status IN ('open', 'partial')`, layer.LayerID, reason); err != nil {
				return nil, skipped, fmt.Errorf("store: mark layer orphaned after reconciliation failure: %w", err)
			}
			if err := insertReconciliationEvent(ctx, tx, reconciliationEventInput{
				UserID:            layer.UserID,
				ExchangeBindingID: layer.ExchangeBindingID,
				MasterSignalID:    signalID,
				LayerID:           layer.LayerID,
				EventType:         "sell_insufficient_base_balance",
				Asset:             baseAsset,
				RequiredAmount:    layer.Quantity,
				AvailableAmount:   layer.AvailableBase,
				Reason:            reason,
			}); err != nil {
				return nil, skipped, err
			}
			if err := insertNotification(ctx, tx, layer.UserID, "Layer Reconciliation Failed", reason); err != nil {
				return nil, skipped, err
			}
			skipped++
			continue
		}
		if err := insertExecutionJob(ctx, tx, job); err != nil {
			return nil, skipped, err
		}
		jobs = append(jobs, job)
	}
	return jobs, skipped, nil
}

func acquireExecutionScopeLock(ctx context.Context, tx pgx.Tx, scope string) error {
	scope = strings.TrimSpace(scope)
	if scope == "" {
		return fmt.Errorf("store: execution lock scope is required")
	}
	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtext('mautrade-execution'), hashtext($1))`, scope); err != nil {
		return fmt.Errorf("store: acquire execution scope lock: %w", err)
	}
	return nil
}

func lockSellLayerForDispatch(ctx context.Context, tx pgx.Tx, layerID, sellPct string) (string, bool, error) {
	var quantity string
	err := tx.QueryRow(ctx, `
SELECT ((remaining_quantity * $2::numeric) / 100)::text
FROM layers
WHERE id = $1::uuid
  AND status IN ('open', 'partial')
  AND ((remaining_quantity * $2::numeric) / 100) > 0
FOR UPDATE`, layerID, sellPct).Scan(&quantity)
	if err != nil {
		if err == pgx.ErrNoRows {
			return "", false, nil
		}
		return "", false, fmt.Errorf("store: lock sell layer for dispatch: %w", err)
	}
	return quantity, true, nil
}

func inFlightSellExecutionJob(ctx context.Context, tx pgx.Tx, layerID string) (string, bool, error) {
	var jobID string
	err := tx.QueryRow(ctx, `
SELECT id::text
FROM execution_jobs
WHERE layer_id = $1::uuid
  AND subject = 'execution.sell.request'
  AND status IN ('queued', 'published', 'running')
ORDER BY created_at DESC
LIMIT 1`, layerID).Scan(&jobID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return "", false, nil
		}
		return "", false, fmt.Errorf("store: check in-flight sell execution job: %w", err)
	}
	return jobID, true, nil
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
	return insertExecutionJobWithStatus(ctx, tx, job, "queued", "")
}

func insertExecutionJobWithStatus(ctx context.Context, tx pgx.Tx, job ExecutionJobRecord, status, lastError string) error {
	payloadJSON, err := json.Marshal(job.Payload)
	if err != nil {
		return fmt.Errorf("store: marshal execution job payload: %w", err)
	}

	var layerID any
	if job.LayerID != "" {
		layerID = job.LayerID
	}

	var lastErrorValue any
	if lastError != "" {
		lastErrorValue = lastError
	}

	_, err = tx.Exec(ctx, `
INSERT INTO execution_jobs (
  id, master_signal_id, layer_id, user_id, exchange_binding_id, subject, payload, status, idempotency_key, last_error
) VALUES (
  $1::uuid, $2::uuid, $3::uuid, $4::uuid, $5::uuid, $6, $7::jsonb, $8, $9, $10
)`,
		job.ID,
		job.MasterSignalID,
		layerID,
		job.UserID,
		job.ExchangeBindingID,
		job.Subject,
		string(payloadJSON),
		status,
		job.IdempotencyKey,
		lastErrorValue,
	)
	if err != nil {
		return fmt.Errorf("store: insert execution job: %w", err)
	}
	return nil
}

type reconciliationEventInput struct {
	UserID            string
	ExchangeBindingID string
	MasterSignalID    string
	LayerID           string
	EventType         string
	Asset             string
	RequiredAmount    string
	AvailableAmount   string
	Reason            string
}

func insertReconciliationEvent(ctx context.Context, tx pgx.Tx, input reconciliationEventInput) error {
	eventID, err := id.New()
	if err != nil {
		return err
	}

	var layerID any
	if input.LayerID != "" {
		layerID = input.LayerID
	}

	_, err = tx.Exec(ctx, `
INSERT INTO reconciliation_events (
  id, user_id, exchange_binding_id, master_signal_id, layer_id,
  event_type, asset, required_amount, available_amount, reason, status, created_at
) VALUES (
  $1::uuid, $2::uuid, $3::uuid, $4::uuid, $5::uuid,
  $6, $7, $8::numeric, $9::numeric, $10, 'open', now()
)`,
		eventID.String(),
		input.UserID,
		input.ExchangeBindingID,
		input.MasterSignalID,
		layerID,
		input.EventType,
		input.Asset,
		decimalOrZero(input.RequiredAmount),
		decimalOrZero(input.AvailableAmount),
		input.Reason,
	)
	if err != nil {
		return fmt.Errorf("store: insert reconciliation event: %w", err)
	}
	return nil
}

func insertNotification(ctx context.Context, tx pgx.Tx, userID, title, message string) error {
	notificationID, err := id.New()
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `
INSERT INTO notifications (
  id, user_id, type, title, message, created_at
) VALUES (
  $1::uuid, $2::uuid, 'reconciliation', $3, $4, now()
)`,
		notificationID.String(),
		userID,
		title,
		message,
	)
	if err != nil {
		return fmt.Errorf("store: insert reconciliation notification: %w", err)
	}
	return nil
}

func decimalGreaterThanZero(value string) (bool, error) {
	decimal, err := qdecimal.Parse(decimalOrZero(value))
	if err != nil {
		return false, fmt.Errorf("store: parse decimal %q: %w", value, err)
	}
	return decimal.Sign() > 0, nil
}

func decimalLessThan(left, right string) (bool, error) {
	leftDecimal, err := qdecimal.Parse(decimalOrZero(left))
	if err != nil {
		return false, fmt.Errorf("store: parse decimal %q: %w", left, err)
	}
	rightDecimal, err := qdecimal.Parse(decimalOrZero(right))
	if err != nil {
		return false, fmt.Errorf("store: parse decimal %q: %w", right, err)
	}
	return leftDecimal.Cmp(rightDecimal) < 0, nil
}

func baseAssetFromSymbol(symbol string) (string, error) {
	parts := strings.Split(strings.ToUpper(strings.TrimSpace(symbol)), "/")
	if len(parts) != 2 || parts[0] == "" {
		return "", fmt.Errorf("store: symbol must be a spot pair like BTC/USDT")
	}
	return parts[0], nil
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
