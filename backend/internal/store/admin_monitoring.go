package store

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/MeViksry/Mautrade/backend/internal/domain/id"
	"github.com/jackc/pgx/v5"
)

type AdminSignalExecutionsView struct {
	Signal  AdminSignalExecutionSignal `json:"signal"`
	Summary AdminExecutionJobSummary   `json:"summary"`
	Jobs    []AdminExecutionJobView    `json:"jobs"`
	Limit   int                        `json:"limit"`
	Offset  int                        `json:"offset"`
}

type AdminSignalExecutionSignal struct {
	ID              string     `json:"id"`
	Type            string     `json:"type"`
	Symbol          string     `json:"symbol"`
	LayerNumber     *int       `json:"layerNumber,omitempty"`
	AllocationPct   string     `json:"allocationPct"`
	SellPct         string     `json:"sellPct"`
	Status          string     `json:"status"`
	IdempotencyKey  string     `json:"idempotencyKey"`
	DispatchedAt    *time.Time `json:"dispatchedAt,omitempty"`
	CompletedAt     *time.Time `json:"completedAt,omitempty"`
	CreatedAt       time.Time  `json:"createdAt"`
	PreviewSnapshot string     `json:"previewSnapshot"`
}

type AdminExecutionJobSummary struct {
	Total      int64 `json:"total"`
	Queued     int64 `json:"queued"`
	Published  int64 `json:"published"`
	Running    int64 `json:"running"`
	Success    int64 `json:"success"`
	Failed     int64 `json:"failed"`
	Skipped    int64 `json:"skipped"`
	DeadLetter int64 `json:"deadLetter"`
}

type AdminExecutionJobView struct {
	ID                string                   `json:"id"`
	MasterSignalID    string                   `json:"masterSignalId"`
	LayerID           *string                  `json:"layerId,omitempty"`
	UserID            string                   `json:"userId"`
	UserEmail         string                   `json:"userEmail"`
	ExchangeBindingID string                   `json:"exchangeBindingId"`
	Exchange          string                   `json:"exchange"`
	Subject           string                   `json:"subject"`
	Side              string                   `json:"side"`
	Symbol            string                   `json:"symbol"`
	Status            string                   `json:"status"`
	Attempts          int                      `json:"attempts"`
	LastError         *string                  `json:"lastError,omitempty"`
	IdempotencyKey    string                   `json:"idempotencyKey"`
	CreatedAt         time.Time                `json:"createdAt"`
	UpdatedAt         time.Time                `json:"updatedAt"`
	Execution         *AdminLayerExecutionView `json:"execution,omitempty"`
	GasFee            *AdminGasFeeLedgerView   `json:"gasFee,omitempty"`
}

type AdminLayerExecutionView struct {
	ID              string     `json:"id"`
	LayerID         *string    `json:"layerId,omitempty"`
	Action          string     `json:"action"`
	Status          string     `json:"status"`
	Quantity        string     `json:"quantity"`
	Price           string     `json:"price"`
	ValueQuote      string     `json:"valueQuote"`
	ExchangeFee     string     `json:"exchangeFee"`
	ExchangeOrderID *string    `json:"exchangeOrderId,omitempty"`
	ErrorCode       *string    `json:"errorCode,omitempty"`
	ErrorMessage    *string    `json:"errorMessage,omitempty"`
	ExecutedAt      *time.Time `json:"executedAt,omitempty"`
}

type AdminGasFeeLedgerView struct {
	ID             string    `json:"id"`
	Type           string    `json:"type"`
	GrossPnL       string    `json:"grossPnl"`
	GasFeeAmount   string    `json:"gasFeeAmount"`
	PlatformRebate string    `json:"platformRebate"`
	NetAmountUser  string    `json:"netAmountUser"`
	ShareRate      string    `json:"shareRate"`
	CalculatedAt   time.Time `json:"calculatedAt"`
}

func (s *DashboardStore) AdminSignalExecutions(ctx context.Context, signalID string, limit, offset int) (AdminSignalExecutionsView, error) {
	if !s.Ready() {
		return AdminSignalExecutionsView{}, fmt.Errorf("store: admin signal executions requires postgres")
	}

	signal, err := s.adminSignalExecutionSignal(ctx, signalID)
	if err != nil {
		return AdminSignalExecutionsView{}, err
	}

	summary, err := s.adminExecutionJobSummary(ctx, signalID)
	if err != nil {
		return AdminSignalExecutionsView{}, err
	}

	jobs, err := s.adminExecutionJobs(ctx, signalID, limit, offset)
	if err != nil {
		return AdminSignalExecutionsView{}, err
	}

	return AdminSignalExecutionsView{
		Signal:  signal,
		Summary: summary,
		Jobs:    jobs,
		Limit:   limit,
		Offset:  offset,
	}, nil
}

func (s *DashboardStore) adminSignalExecutionSignal(ctx context.Context, signalID string) (AdminSignalExecutionSignal, error) {
	var signal AdminSignalExecutionSignal
	var previewSnapshot string
	if err := s.db.QueryRow(ctx, `
SELECT id::text, type, symbol, layer_number, allocation_pct::text, sell_pct::text,
       status, idempotency_key, dispatched_at, completed_at, created_at, preview_snapshot::text
FROM master_signals
WHERE id = $1::uuid`, signalID).Scan(
		&signal.ID,
		&signal.Type,
		&signal.Symbol,
		&signal.LayerNumber,
		&signal.AllocationPct,
		&signal.SellPct,
		&signal.Status,
		&signal.IdempotencyKey,
		&signal.DispatchedAt,
		&signal.CompletedAt,
		&signal.CreatedAt,
		&previewSnapshot,
	); err != nil {
		return AdminSignalExecutionSignal{}, fmt.Errorf("store: admin signal execution signal: %w", err)
	}
	signal.PreviewSnapshot = previewSnapshot
	return signal, nil
}

func (s *DashboardStore) adminExecutionJobSummary(ctx context.Context, signalID string) (AdminExecutionJobSummary, error) {
	var summary AdminExecutionJobSummary
	if err := s.db.QueryRow(ctx, `
SELECT
  COUNT(*)::bigint,
  COUNT(*) FILTER (WHERE status = 'queued')::bigint,
  COUNT(*) FILTER (WHERE status = 'published')::bigint,
  COUNT(*) FILTER (WHERE status = 'running')::bigint,
  COUNT(*) FILTER (WHERE status = 'success')::bigint,
  COUNT(*) FILTER (WHERE status = 'failed')::bigint,
  COUNT(*) FILTER (WHERE status = 'skipped')::bigint,
  COUNT(*) FILTER (WHERE status = 'dead_letter')::bigint
FROM execution_jobs
WHERE master_signal_id = $1::uuid`, signalID).Scan(
		&summary.Total,
		&summary.Queued,
		&summary.Published,
		&summary.Running,
		&summary.Success,
		&summary.Failed,
		&summary.Skipped,
		&summary.DeadLetter,
	); err != nil {
		return AdminExecutionJobSummary{}, fmt.Errorf("store: admin execution job summary: %w", err)
	}
	return summary, nil
}

func (s *DashboardStore) adminExecutionJobs(ctx context.Context, signalID string, limit, offset int) ([]AdminExecutionJobView, error) {
	rows, err := s.db.Query(ctx, `
SELECT
  j.id::text,
  j.master_signal_id::text,
  j.layer_id::text,
  j.user_id::text,
  u.email,
  j.exchange_binding_id::text,
  b.exchange_name,
  j.subject,
  COALESCE(j.payload->>'side', '') AS side,
  COALESCE(j.payload->>'symbol', '') AS symbol,
  j.status,
  j.attempts,
  j.last_error,
  j.idempotency_key,
  j.created_at,
  j.updated_at,
  e.id::text,
  e.layer_id::text,
  e.action,
  e.status,
  e.quantity::text,
  e.price::text,
  e.value_quote::text,
  e.exchange_fee::text,
  e.exchange_order_id,
  e.error_code,
  e.error_message,
  e.executed_at,
  g.id::text,
  g.type,
  g.gross_pnl::text,
  g.gas_fee_amount::text,
  g.platform_rebate::text,
  g.net_amount_user::text,
  g.share_rate::text,
  g.calculated_at
FROM execution_jobs j
JOIN users u ON u.id = j.user_id
JOIN exchange_bindings b ON b.id = j.exchange_binding_id
LEFT JOIN layer_executions e ON e.idempotency_key = j.id::text
LEFT JOIN gas_fee_ledger g ON g.execution_id = e.id
WHERE j.master_signal_id = $1::uuid
ORDER BY j.created_at ASC
LIMIT $2 OFFSET $3`, signalID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("store: admin execution jobs: %w", err)
	}
	defer rows.Close()

	var jobs []AdminExecutionJobView
	for rows.Next() {
		var job AdminExecutionJobView
		var executionID *string
		var executionLayerID *string
		var executionAction *string
		var executionStatus *string
		var executionQuantity *string
		var executionPrice *string
		var executionValueQuote *string
		var executionExchangeFee *string
		var executionExchangeOrderID *string
		var executionErrorCode *string
		var executionErrorMessage *string
		var executionExecutedAt *time.Time
		var gasFeeID *string
		var gasFeeType *string
		var gasFeeGrossPnL *string
		var gasFeeAmount *string
		var gasFeePlatformRebate *string
		var gasFeeNetAmountUser *string
		var gasFeeShareRate *string
		var gasFeeCalculatedAt *time.Time

		if err := rows.Scan(
			&job.ID,
			&job.MasterSignalID,
			&job.LayerID,
			&job.UserID,
			&job.UserEmail,
			&job.ExchangeBindingID,
			&job.Exchange,
			&job.Subject,
			&job.Side,
			&job.Symbol,
			&job.Status,
			&job.Attempts,
			&job.LastError,
			&job.IdempotencyKey,
			&job.CreatedAt,
			&job.UpdatedAt,
			&executionID,
			&executionLayerID,
			&executionAction,
			&executionStatus,
			&executionQuantity,
			&executionPrice,
			&executionValueQuote,
			&executionExchangeFee,
			&executionExchangeOrderID,
			&executionErrorCode,
			&executionErrorMessage,
			&executionExecutedAt,
			&gasFeeID,
			&gasFeeType,
			&gasFeeGrossPnL,
			&gasFeeAmount,
			&gasFeePlatformRebate,
			&gasFeeNetAmountUser,
			&gasFeeShareRate,
			&gasFeeCalculatedAt,
		); err != nil {
			return nil, fmt.Errorf("store: scan admin execution job: %w", err)
		}

		if executionID != nil {
			job.Execution = &AdminLayerExecutionView{
				ID:              *executionID,
				LayerID:         executionLayerID,
				Action:          valueOrEmpty(executionAction),
				Status:          valueOrEmpty(executionStatus),
				Quantity:        valueOrEmpty(executionQuantity),
				Price:           valueOrEmpty(executionPrice),
				ValueQuote:      valueOrEmpty(executionValueQuote),
				ExchangeFee:     valueOrEmpty(executionExchangeFee),
				ExchangeOrderID: executionExchangeOrderID,
				ErrorCode:       executionErrorCode,
				ErrorMessage:    executionErrorMessage,
				ExecutedAt:      executionExecutedAt,
			}
		}

		if gasFeeID != nil && gasFeeCalculatedAt != nil {
			job.GasFee = &AdminGasFeeLedgerView{
				ID:             *gasFeeID,
				Type:           valueOrEmpty(gasFeeType),
				GrossPnL:       valueOrEmpty(gasFeeGrossPnL),
				GasFeeAmount:   valueOrEmpty(gasFeeAmount),
				PlatformRebate: valueOrEmpty(gasFeePlatformRebate),
				NetAmountUser:  valueOrEmpty(gasFeeNetAmountUser),
				ShareRate:      valueOrEmpty(gasFeeShareRate),
				CalculatedAt:   *gasFeeCalculatedAt,
			}
		}

		jobs = append(jobs, job)
	}
	if jobs == nil {
		jobs = []AdminExecutionJobView{}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: iterate admin execution jobs: %w", err)
	}
	return jobs, nil
}

type AdminReconciliationEventsParams struct {
	Status string
	Limit  int
	Offset int
}

type AdminReconciliationEventsView struct {
	Status string                         `json:"status"`
	Limit  int                            `json:"limit"`
	Offset int                            `json:"offset"`
	Events []AdminReconciliationEventView `json:"events"`
}

type AdminReconciliationEventView struct {
	ID                string     `json:"id"`
	UserID            string     `json:"userId"`
	UserEmail         string     `json:"userEmail"`
	Username          string     `json:"username"`
	ExchangeBindingID string     `json:"exchangeBindingId"`
	Exchange          string     `json:"exchange"`
	MasterSignalID    *string    `json:"masterSignalId,omitempty"`
	SignalType        *string    `json:"signalType,omitempty"`
	Symbol            *string    `json:"symbol,omitempty"`
	LayerID           *string    `json:"layerId,omitempty"`
	LayerNumber       *int       `json:"layerNumber,omitempty"`
	LayerStatus       *string    `json:"layerStatus,omitempty"`
	EventType         string     `json:"eventType"`
	Asset             string     `json:"asset"`
	RequiredAmount    string     `json:"requiredAmount"`
	AvailableAmount   string     `json:"availableAmount"`
	Reason            string     `json:"reason"`
	Status            string     `json:"status"`
	ResolvedAt        *time.Time `json:"resolvedAt,omitempty"`
	CreatedAt         time.Time  `json:"createdAt"`
}

func (s *DashboardStore) AdminReconciliationEvents(ctx context.Context, params AdminReconciliationEventsParams) (AdminReconciliationEventsView, error) {
	if !s.Ready() {
		return AdminReconciliationEventsView{}, fmt.Errorf("store: admin reconciliation events requires postgres")
	}

	rows, err := s.db.Query(ctx, adminReconciliationEventsQuery()+`
WHERE ($1 = '' OR re.status = $1)
ORDER BY re.created_at DESC
LIMIT $2 OFFSET $3`, params.Status, params.Limit, params.Offset)
	if err != nil {
		return AdminReconciliationEventsView{}, fmt.Errorf("store: admin reconciliation events: %w", err)
	}
	defer rows.Close()

	events, err := scanAdminReconciliationEvents(rows)
	if err != nil {
		return AdminReconciliationEventsView{}, err
	}
	return AdminReconciliationEventsView{
		Status: params.Status,
		Limit:  params.Limit,
		Offset: params.Offset,
		Events: events,
	}, nil
}

type ResolveReconciliationEventParams struct {
	EventID        string
	ActorID        string
	ResolutionNote string
}

func (s *DashboardStore) ResolveReconciliationEvent(ctx context.Context, params ResolveReconciliationEventParams) (AdminReconciliationEventView, error) {
	if !s.Ready() {
		return AdminReconciliationEventView{}, fmt.Errorf("store: resolve reconciliation event requires postgres")
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return AdminReconciliationEventView{}, fmt.Errorf("store: begin resolve reconciliation event: %w", err)
	}
	defer tx.Rollback(ctx)

	var beforeStatus string
	var beforeResolvedAt *time.Time
	var beforeReason string
	if err := tx.QueryRow(ctx, `
SELECT status, resolved_at, reason
FROM reconciliation_events
WHERE id = $1::uuid
FOR UPDATE`, params.EventID).Scan(&beforeStatus, &beforeResolvedAt, &beforeReason); err != nil {
		return AdminReconciliationEventView{}, fmt.Errorf("store: lock reconciliation event: %w", err)
	}

	if _, err := tx.Exec(ctx, `
UPDATE reconciliation_events
SET status = 'resolved',
    resolved_at = COALESCE(resolved_at, now())
WHERE id = $1::uuid`, params.EventID); err != nil {
		return AdminReconciliationEventView{}, fmt.Errorf("store: update reconciliation event resolved: %w", err)
	}

	afterEvent, err := adminReconciliationEventByID(ctx, tx, params.EventID)
	if err != nil {
		return AdminReconciliationEventView{}, err
	}

	if err := insertReconciliationResolveAudit(ctx, tx, params, beforeStatus, beforeResolvedAt, beforeReason, afterEvent); err != nil {
		return AdminReconciliationEventView{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return AdminReconciliationEventView{}, fmt.Errorf("store: commit resolve reconciliation event: %w", err)
	}

	return afterEvent, nil
}

type reconciliationQueryer interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}

func adminReconciliationEventByID(ctx context.Context, queryer reconciliationQueryer, eventID string) (AdminReconciliationEventView, error) {
	row := queryer.QueryRow(ctx, adminReconciliationEventsQuery()+`
WHERE re.id = $1::uuid`, eventID)

	var event AdminReconciliationEventView
	if err := row.Scan(
		&event.ID,
		&event.UserID,
		&event.UserEmail,
		&event.Username,
		&event.ExchangeBindingID,
		&event.Exchange,
		&event.MasterSignalID,
		&event.SignalType,
		&event.Symbol,
		&event.LayerID,
		&event.LayerNumber,
		&event.LayerStatus,
		&event.EventType,
		&event.Asset,
		&event.RequiredAmount,
		&event.AvailableAmount,
		&event.Reason,
		&event.Status,
		&event.ResolvedAt,
		&event.CreatedAt,
	); err != nil {
		return AdminReconciliationEventView{}, fmt.Errorf("store: reconciliation event by id: %w", err)
	}
	return event, nil
}

func adminReconciliationEventsQuery() string {
	return `
SELECT
  re.id::text,
  re.user_id::text,
  u.email,
  u.username,
  re.exchange_binding_id::text,
  b.exchange_name,
  re.master_signal_id::text,
  ms.type,
  COALESCE(ms.symbol, l.symbol),
  re.layer_id::text,
  l.layer_number,
  l.status,
  re.event_type,
  re.asset,
  re.required_amount::text,
  re.available_amount::text,
  re.reason,
  re.status,
  re.resolved_at,
  re.created_at
FROM reconciliation_events re
JOIN users u ON u.id = re.user_id
JOIN exchange_bindings b ON b.id = re.exchange_binding_id
LEFT JOIN master_signals ms ON ms.id = re.master_signal_id
LEFT JOIN layers l ON l.id = re.layer_id
`
}

func scanAdminReconciliationEvents(rows pgx.Rows) ([]AdminReconciliationEventView, error) {
	var events []AdminReconciliationEventView
	for rows.Next() {
		var event AdminReconciliationEventView
		if err := rows.Scan(
			&event.ID,
			&event.UserID,
			&event.UserEmail,
			&event.Username,
			&event.ExchangeBindingID,
			&event.Exchange,
			&event.MasterSignalID,
			&event.SignalType,
			&event.Symbol,
			&event.LayerID,
			&event.LayerNumber,
			&event.LayerStatus,
			&event.EventType,
			&event.Asset,
			&event.RequiredAmount,
			&event.AvailableAmount,
			&event.Reason,
			&event.Status,
			&event.ResolvedAt,
			&event.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("store: scan admin reconciliation event: %w", err)
		}
		events = append(events, event)
	}
	if events == nil {
		events = []AdminReconciliationEventView{}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: iterate admin reconciliation events: %w", err)
	}
	return events, nil
}

func insertReconciliationResolveAudit(ctx context.Context, tx pgx.Tx, params ResolveReconciliationEventParams, beforeStatus string, beforeResolvedAt *time.Time, beforeReason string, afterEvent AdminReconciliationEventView) error {
	auditID, err := id.New()
	if err != nil {
		return err
	}

	beforeJSON, err := json.Marshal(map[string]any{
		"status":      beforeStatus,
		"resolved_at": beforeResolvedAt,
		"reason":      beforeReason,
	})
	if err != nil {
		return fmt.Errorf("store: marshal reconciliation before audit: %w", err)
	}

	afterJSON, err := json.Marshal(map[string]any{
		"event":           afterEvent,
		"resolution_note": params.ResolutionNote,
	})
	if err != nil {
		return fmt.Errorf("store: marshal reconciliation after audit: %w", err)
	}

	var actorID any
	if params.ActorID != "" {
		actorID = params.ActorID
	}

	_, err = tx.Exec(ctx, `
INSERT INTO audit_logs (
  id, actor_type, actor_id, action, entity, entity_id, before_state, after_state, created_at
) VALUES (
  $1::uuid, 'admin', $2::uuid, 'reconciliation_event_resolved', 'reconciliation_event',
  $3::uuid, $4::jsonb, $5::jsonb, now()
)`,
		auditID.String(),
		actorID,
		params.EventID,
		string(beforeJSON),
		string(afterJSON),
	)
	if err != nil {
		return fmt.Errorf("store: insert reconciliation resolve audit: %w", err)
	}
	return nil
}

func valueOrEmpty(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
