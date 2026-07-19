package store

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/MeViksry/Mautrade/backend/internal/domain/gasfee"
	"github.com/MeViksry/Mautrade/backend/internal/domain/id"
	"github.com/MeViksry/qdecimal"
	"github.com/jackc/pgx/v5"
)

type ExecutionResult struct {
	RequestID       string  `json:"request_id"`
	IdempotencyKey  string  `json:"idempotency_key"`
	MasterSignalID  string  `json:"master_signal_id"`
	UserID          string  `json:"user_id"`
	LayerID         *string `json:"layer_id,omitempty"`
	Exchange        string  `json:"exchange"`
	Symbol          string  `json:"symbol"`
	Side            string  `json:"side"`
	Status          string  `json:"status"`
	FilledQuantity  string  `json:"filled_quantity"`
	FillPrice       string  `json:"fill_price"`
	FillValueQuote  string  `json:"fill_value_quote"`
	ExchangeFee     string  `json:"exchange_fee"`
	ExchangeOrderID *string `json:"exchange_order_id,omitempty"`
	ErrorCode       *string `json:"error_code,omitempty"`
	ErrorMessage    *string `json:"error_message,omitempty"`
	ExecutedAt      string  `json:"executed_at"`
}

type ExecutionResultApplySummary struct {
	JobID          string      `json:"jobId"`
	LayerID        string      `json:"layerId,omitempty"`
	ExecutionID    string      `json:"executionId,omitempty"`
	GasFeeLedgerID string      `json:"gasFeeLedgerId,omitempty"`
	Status         string      `json:"status"`
	Side           string      `json:"side"`
	Duplicate      bool        `json:"duplicate"`
	AppliedAt      time.Time   `json:"appliedAt"`
	GasFee         *GasFeeView `json:"gasFee,omitempty"`
}

type GasFeeView struct {
	Type            string `json:"type"`
	GrossPnL        string `json:"grossPnl"`
	GasFeeAmount    string `json:"gasFeeAmount"`
	PlatformRebate  string `json:"platformRebate"`
	NetAmountToUser string `json:"netAmountToUser"`
	NetLossToUser   string `json:"netLossToUser"`
}

type executionJobForResult struct {
	ID                string
	MasterSignalID    string
	LayerID           *string
	UserID            string
	ExchangeBindingID string
	Subject           string
	Status            string
	Payload           []byte
	Request           ExecutionPayload
}

type layerForSettlement struct {
	ID                string
	UserID            string
	ExchangeBindingID string
	MasterSignalID    string
	Symbol            string
	EntryPrice        string
	RemainingQuantity string
}

func (s *DashboardStore) ApplyExecutionResult(ctx context.Context, result ExecutionResult, calculator gasfee.Calculator) (ExecutionResultApplySummary, error) {
	if !s.Ready() {
		return ExecutionResultApplySummary{}, fmt.Errorf("store: execution result requires postgres")
	}

	normalized, parsed, err := normalizeExecutionResult(result)
	if err != nil {
		return ExecutionResultApplySummary{}, err
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return ExecutionResultApplySummary{}, fmt.Errorf("store: begin execution result: %w", err)
	}
	defer tx.Rollback(ctx)

	job, err := lockExecutionJobForResult(ctx, tx, normalized.IdempotencyKey)
	if err != nil {
		return ExecutionResultApplySummary{}, err
	}
	if err := ensureResultMatchesJob(normalized, job); err != nil {
		return ExecutionResultApplySummary{}, err
	}

	duplicate, summary, err := executionResultDuplicateSummary(ctx, tx, job)
	if err != nil {
		return ExecutionResultApplySummary{}, err
	}
	if duplicate {
		if err := tx.Commit(ctx); err != nil {
			return ExecutionResultApplySummary{}, fmt.Errorf("store: commit duplicate execution result: %w", err)
		}
		return summary, nil
	}

	var applied ExecutionResultApplySummary
	switch normalized.Side {
	case "buy":
		applied, err = applyBuyExecutionResult(ctx, tx, job, normalized, parsed)
	case "sell":
		applied, err = applySellExecutionResult(ctx, tx, job, normalized, parsed, calculator)
	default:
		err = fmt.Errorf("store: unsupported result side %q", normalized.Side)
	}
	if err != nil {
		return ExecutionResultApplySummary{}, err
	}

	if err := refreshMasterSignalCompletion(ctx, tx, job.MasterSignalID); err != nil {
		return ExecutionResultApplySummary{}, err
	}

	if err := insertExecutionResultAuditLog(ctx, tx, normalized, applied); err != nil {
		return ExecutionResultApplySummary{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return ExecutionResultApplySummary{}, fmt.Errorf("store: commit execution result: %w", err)
	}

	return applied, nil
}

type parsedExecutionResult struct {
	FilledQuantity qdecimal.Decimal
	FillPrice      qdecimal.Decimal
	FillValueQuote qdecimal.Decimal
	ExchangeFee    qdecimal.Decimal
	ExecutedAt     time.Time
}

func normalizeExecutionResult(result ExecutionResult) (ExecutionResult, parsedExecutionResult, error) {
	result.RequestID = strings.TrimSpace(result.RequestID)
	result.IdempotencyKey = strings.TrimSpace(result.IdempotencyKey)
	result.MasterSignalID = strings.TrimSpace(result.MasterSignalID)
	result.UserID = strings.TrimSpace(result.UserID)
	result.Exchange = strings.ToLower(strings.TrimSpace(result.Exchange))
	result.Symbol = strings.ToUpper(strings.TrimSpace(result.Symbol))
	result.Side = strings.ToLower(strings.TrimSpace(result.Side))
	result.Status = strings.ToLower(strings.TrimSpace(result.Status))
	result.FilledQuantity = decimalOrZero(result.FilledQuantity)
	result.FillPrice = decimalOrZero(result.FillPrice)
	result.FillValueQuote = decimalOrZero(result.FillValueQuote)
	result.ExchangeFee = decimalOrZero(result.ExchangeFee)
	result.ExecutedAt = strings.TrimSpace(result.ExecutedAt)
	if result.LayerID != nil {
		layerID := strings.TrimSpace(*result.LayerID)
		if layerID == "" {
			result.LayerID = nil
		} else {
			result.LayerID = &layerID
		}
	}

	required := map[string]string{
		"request_id":       result.RequestID,
		"idempotency_key":  result.IdempotencyKey,
		"master_signal_id": result.MasterSignalID,
		"user_id":          result.UserID,
		"exchange":         result.Exchange,
		"symbol":           result.Symbol,
		"side":             result.Side,
		"status":           result.Status,
	}
	for field, value := range required {
		if value == "" {
			return ExecutionResult{}, parsedExecutionResult{}, fmt.Errorf("store: %s is required", field)
		}
	}
	if result.Side != "buy" && result.Side != "sell" {
		return ExecutionResult{}, parsedExecutionResult{}, fmt.Errorf("store: side must be buy or sell")
	}
	if result.Status != "success" && result.Status != "partial" && result.Status != "failed" && result.Status != "skipped" {
		return ExecutionResult{}, parsedExecutionResult{}, fmt.Errorf("store: status must be success, partial, failed, or skipped")
	}

	filledQuantity, err := qdecimal.Parse(result.FilledQuantity)
	if err != nil {
		return ExecutionResult{}, parsedExecutionResult{}, fmt.Errorf("store: filled_quantity must be decimal: %w", err)
	}
	fillPrice, err := qdecimal.Parse(result.FillPrice)
	if err != nil {
		return ExecutionResult{}, parsedExecutionResult{}, fmt.Errorf("store: fill_price must be decimal: %w", err)
	}
	fillValueQuote, err := qdecimal.Parse(result.FillValueQuote)
	if err != nil {
		return ExecutionResult{}, parsedExecutionResult{}, fmt.Errorf("store: fill_value_quote must be decimal: %w", err)
	}
	exchangeFee, err := qdecimal.Parse(result.ExchangeFee)
	if err != nil {
		return ExecutionResult{}, parsedExecutionResult{}, fmt.Errorf("store: exchange_fee must be decimal: %w", err)
	}
	if exchangeFee.Sign() < 0 {
		return ExecutionResult{}, parsedExecutionResult{}, fmt.Errorf("store: exchange_fee cannot be negative")
	}
	if result.Status == "success" || result.Status == "partial" {
		if filledQuantity.Sign() <= 0 {
			return ExecutionResult{}, parsedExecutionResult{}, fmt.Errorf("store: filled_quantity must be greater than zero for filled execution")
		}
		if fillPrice.Sign() <= 0 {
			return ExecutionResult{}, parsedExecutionResult{}, fmt.Errorf("store: fill_price must be greater than zero for filled execution")
		}
		if fillValueQuote.Sign() <= 0 {
			return ExecutionResult{}, parsedExecutionResult{}, fmt.Errorf("store: fill_value_quote must be greater than zero for filled execution")
		}
	}

	executedAt := time.Now().UTC()
	if result.ExecutedAt != "" {
		parsedTime, err := time.Parse(time.RFC3339Nano, result.ExecutedAt)
		if err != nil {
			return ExecutionResult{}, parsedExecutionResult{}, fmt.Errorf("store: executed_at must be RFC3339: %w", err)
		}
		executedAt = parsedTime.UTC()
	}
	result.ExecutedAt = executedAt.Format(time.RFC3339Nano)

	return result, parsedExecutionResult{
		FilledQuantity: filledQuantity,
		FillPrice:      fillPrice,
		FillValueQuote: fillValueQuote,
		ExchangeFee:    exchangeFee,
		ExecutedAt:     executedAt,
	}, nil
}

func decimalOrZero(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "0"
	}
	return value
}

func lockExecutionJobForResult(ctx context.Context, tx pgx.Tx, idempotencyKey string) (executionJobForResult, error) {
	var job executionJobForResult
	var payloadText string
	err := tx.QueryRow(ctx, `
SELECT id::text, master_signal_id::text, layer_id::text, user_id::text,
       exchange_binding_id::text, subject, status, payload::text
FROM execution_jobs
WHERE idempotency_key = $1
FOR UPDATE`, idempotencyKey).Scan(
		&job.ID,
		&job.MasterSignalID,
		&job.LayerID,
		&job.UserID,
		&job.ExchangeBindingID,
		&job.Subject,
		&job.Status,
		&payloadText,
	)
	if err != nil {
		return executionJobForResult{}, fmt.Errorf("store: lock execution job by idempotency key: %w", err)
	}
	job.Payload = []byte(payloadText)
	if err := json.Unmarshal(job.Payload, &job.Request); err != nil {
		return executionJobForResult{}, fmt.Errorf("store: decode execution job payload: %w", err)
	}
	if job.Request.ID == "" {
		job.Request.ID = job.ID
	}
	return job, nil
}

func ensureResultMatchesJob(result ExecutionResult, job executionJobForResult) error {
	if result.RequestID != job.Request.ID {
		return fmt.Errorf("store: execution result request_id does not match execution job")
	}
	if result.MasterSignalID != job.MasterSignalID {
		return fmt.Errorf("store: execution result master_signal_id does not match execution job")
	}
	if result.UserID != job.UserID {
		return fmt.Errorf("store: execution result user_id does not match execution job")
	}

	if strings.ToLower(job.Request.Exchange) != result.Exchange {
		return fmt.Errorf("store: execution result exchange does not match execution job")
	}
	if strings.ToUpper(job.Request.Symbol) != result.Symbol {
		return fmt.Errorf("store: execution result symbol does not match execution job")
	}
	if strings.ToLower(job.Request.Side) != result.Side {
		return fmt.Errorf("store: execution result side does not match execution job")
	}
	if result.Side == "sell" {
		expectedLayerID := job.Request.LayerID
		if expectedLayerID == "" && job.LayerID != nil {
			expectedLayerID = *job.LayerID
		}
		if expectedLayerID == "" {
			return fmt.Errorf("store: sell execution job is missing layer_id")
		}
		if result.LayerID != nil && *result.LayerID != expectedLayerID {
			return fmt.Errorf("store: execution result layer_id does not match execution job")
		}
	}
	return nil
}

func executionResultDuplicateSummary(ctx context.Context, tx pgx.Tx, job executionJobForResult) (bool, ExecutionResultApplySummary, error) {
	var summary ExecutionResultApplySummary
	var layerID *string
	var executionID string
	err := tx.QueryRow(ctx, `
SELECT id::text, layer_id::text, status
FROM layer_executions
WHERE idempotency_key = $1`, job.Request.ID).Scan(&executionID, &layerID, &summary.Status)
	if err == nil {
		summary.JobID = job.ID
		summary.ExecutionID = executionID
		summary.Duplicate = true
		summary.AppliedAt = time.Now().UTC()
		if layerID != nil {
			summary.LayerID = *layerID
		}
		if job.Subject == "execution.sell.request" {
			summary.Side = "sell"
			_ = tx.QueryRow(ctx, `
SELECT id::text
FROM gas_fee_ledger
WHERE execution_id = $1::uuid`, executionID).Scan(&summary.GasFeeLedgerID)
		} else {
			summary.Side = "buy"
		}
		return true, summary, nil
	}
	if err != pgx.ErrNoRows {
		return false, ExecutionResultApplySummary{}, fmt.Errorf("store: check duplicate execution result: %w", err)
	}
	if job.Status == "success" {
		return true, ExecutionResultApplySummary{
			JobID:     job.ID,
			Status:    "success",
			Duplicate: true,
			AppliedAt: time.Now().UTC(),
		}, nil
	}
	return false, ExecutionResultApplySummary{}, nil
}

func applyBuyExecutionResult(ctx context.Context, tx pgx.Tx, job executionJobForResult, result ExecutionResult, parsed parsedExecutionResult) (ExecutionResultApplySummary, error) {
	executionID, err := newUUIDText()
	if err != nil {
		return ExecutionResultApplySummary{}, err
	}

	if result.Status == "failed" || result.Status == "skipped" {
		if err := insertLayerExecution(ctx, tx, layerExecutionInsert{
			ID:                executionID,
			LayerID:           nil,
			MasterSignalID:    job.MasterSignalID,
			UserID:            job.UserID,
			ExchangeBindingID: job.ExchangeBindingID,
			Action:            "buy",
			Symbol:            result.Symbol,
			Quantity:          parsed.FilledQuantity.String(),
			Price:             parsed.FillPrice.String(),
			ValueQuote:        parsed.FillValueQuote.String(),
			ExchangeFee:       parsed.ExchangeFee.String(),
			ExchangeOrderID:   result.ExchangeOrderID,
			Status:            result.Status,
			ErrorCode:         result.ErrorCode,
			ErrorMessage:      result.ErrorMessage,
			IdempotencyKey:    job.Request.ID,
			ExecutedAt:        parsed.ExecutedAt,
		}); err != nil {
			return ExecutionResultApplySummary{}, err
		}
		if err := markExecutionJobTerminal(ctx, tx, job.ID, "failed", result.ErrorMessage); err != nil {
			return ExecutionResultApplySummary{}, err
		}
		return ExecutionResultApplySummary{
			JobID:       job.ID,
			ExecutionID: executionID,
			Status:      result.Status,
			Side:        "buy",
			AppliedAt:   time.Now().UTC(),
		}, nil
	}

	layerID, err := newUUIDText()
	if err != nil {
		return ExecutionResultApplySummary{}, err
	}

	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtext($1)::bigint)`, "layer-number:"+job.UserID+":"+result.Symbol); err != nil {
		return ExecutionResultApplySummary{}, fmt.Errorf("store: acquire layer-number lock: %w", err)
	}

	var layerNumber int
	if err := tx.QueryRow(ctx, `
SELECT COALESCE(MAX(layer_number), 0) + 1
FROM layers
WHERE user_id = $1::uuid
  AND symbol = $2`, job.UserID, result.Symbol).Scan(&layerNumber); err != nil {
		return ExecutionResultApplySummary{}, fmt.Errorf("store: next layer number: %w", err)
	}

	var allocationPct string
	if err := tx.QueryRow(ctx, `
SELECT allocation_pct::text
FROM master_signals
WHERE id = $1::uuid`, job.MasterSignalID).Scan(&allocationPct); err != nil {
		return ExecutionResultApplySummary{}, fmt.Errorf("store: buy allocation pct: %w", err)
	}

	if _, err := tx.Exec(ctx, `
INSERT INTO layers (
  id, user_id, exchange_binding_id, master_signal_id, layer_number, symbol,
  entry_price, entry_quantity, entry_value_quote, remaining_quantity, allocation_pct,
  status, opened_at, updated_at
) VALUES (
  $1::uuid, $2::uuid, $3::uuid, $4::uuid, $5, $6,
  $7::numeric, $8::numeric, $9::numeric, $8::numeric, $10::numeric,
  'open', $11, $11
)`,
		layerID,
		job.UserID,
		job.ExchangeBindingID,
		job.MasterSignalID,
		layerNumber,
		result.Symbol,
		parsed.FillPrice.String(),
		parsed.FilledQuantity.String(),
		parsed.FillValueQuote.String(),
		allocationPct,
		parsed.ExecutedAt,
	); err != nil {
		return ExecutionResultApplySummary{}, fmt.Errorf("store: insert buy layer: %w", err)
	}

	if err := insertLayerExecution(ctx, tx, layerExecutionInsert{
		ID:                executionID,
		LayerID:           &layerID,
		MasterSignalID:    job.MasterSignalID,
		UserID:            job.UserID,
		ExchangeBindingID: job.ExchangeBindingID,
		Action:            "buy",
		Symbol:            result.Symbol,
		Quantity:          parsed.FilledQuantity.String(),
		Price:             parsed.FillPrice.String(),
		ValueQuote:        parsed.FillValueQuote.String(),
		ExchangeFee:       parsed.ExchangeFee.String(),
		ExchangeOrderID:   result.ExchangeOrderID,
		Status:            result.Status,
		ErrorCode:         result.ErrorCode,
		ErrorMessage:      result.ErrorMessage,
		IdempotencyKey:    job.Request.ID,
		ExecutedAt:        parsed.ExecutedAt,
	}); err != nil {
		return ExecutionResultApplySummary{}, err
	}

	if err := markExecutionJobTerminal(ctx, tx, job.ID, "success", nil); err != nil {
		return ExecutionResultApplySummary{}, err
	}

	return ExecutionResultApplySummary{
		JobID:       job.ID,
		LayerID:     layerID,
		ExecutionID: executionID,
		Status:      result.Status,
		Side:        "buy",
		AppliedAt:   time.Now().UTC(),
	}, nil
}

func applySellExecutionResult(ctx context.Context, tx pgx.Tx, job executionJobForResult, result ExecutionResult, parsed parsedExecutionResult, calculator gasfee.Calculator) (ExecutionResultApplySummary, error) {
	layerID := ""
	if result.LayerID != nil {
		layerID = *result.LayerID
	} else if job.LayerID != nil {
		layerID = *job.LayerID
	}
	if layerID == "" {
		return ExecutionResultApplySummary{}, fmt.Errorf("store: sell result requires layer_id")
	}

	layer, err := lockLayerForSettlement(ctx, tx, layerID)
	if err != nil {
		return ExecutionResultApplySummary{}, err
	}
	if layer.UserID != job.UserID || layer.ExchangeBindingID != job.ExchangeBindingID || layer.Symbol != result.Symbol {
		return ExecutionResultApplySummary{}, fmt.Errorf("store: sell result does not match locked layer")
	}

	executionID, err := newUUIDText()
	if err != nil {
		return ExecutionResultApplySummary{}, err
	}

	if result.Status == "failed" || result.Status == "skipped" {
		if err := insertLayerExecution(ctx, tx, layerExecutionInsert{
			ID:                executionID,
			LayerID:           &layerID,
			MasterSignalID:    job.MasterSignalID,
			UserID:            job.UserID,
			ExchangeBindingID: job.ExchangeBindingID,
			Action:            "sell",
			Symbol:            result.Symbol,
			Quantity:          parsed.FilledQuantity.String(),
			Price:             parsed.FillPrice.String(),
			ValueQuote:        parsed.FillValueQuote.String(),
			ExchangeFee:       parsed.ExchangeFee.String(),
			ExchangeOrderID:   result.ExchangeOrderID,
			Status:            result.Status,
			ErrorCode:         result.ErrorCode,
			ErrorMessage:      result.ErrorMessage,
			IdempotencyKey:    job.Request.ID,
			ExecutedAt:        parsed.ExecutedAt,
		}); err != nil {
			return ExecutionResultApplySummary{}, err
		}
		if err := markExecutionJobTerminal(ctx, tx, job.ID, "failed", result.ErrorMessage); err != nil {
			return ExecutionResultApplySummary{}, err
		}
		return ExecutionResultApplySummary{
			JobID:       job.ID,
			LayerID:     layerID,
			ExecutionID: executionID,
			Status:      result.Status,
			Side:        "sell",
			AppliedAt:   time.Now().UTC(),
		}, nil
	}

	entryPrice, err := qdecimal.Parse(layer.EntryPrice)
	if err != nil {
		return ExecutionResultApplySummary{}, fmt.Errorf("store: parse layer entry price: %w", err)
	}
	remainingQuantity, err := qdecimal.Parse(layer.RemainingQuantity)
	if err != nil {
		return ExecutionResultApplySummary{}, fmt.Errorf("store: parse layer remaining quantity: %w", err)
	}
	if parsed.FilledQuantity.Cmp(remainingQuantity) > 0 {
		return ExecutionResultApplySummary{}, fmt.Errorf("store: filled quantity exceeds layer remaining quantity")
	}
	nextRemaining := remainingQuantity.Sub(parsed.FilledQuantity)
	nextStatus := "partial"
	var closedAt any
	if nextRemaining.Sign() == 0 {
		nextStatus = "closed"
		closedAt = parsed.ExecutedAt
	}

	entryValue := entryPrice.Mul(parsed.FilledQuantity)
	gasResult := calculator.CalculateFromValues(entryValue, parsed.FillValueQuote)
	gasFeeLedgerID, err := newUUIDText()
	if err != nil {
		return ExecutionResultApplySummary{}, err
	}
	netAmountUser := gasResult.NetAmountToUser
	if gasResult.Type == gasfee.TypeLossRebate {
		netAmountUser = gasResult.NetLossToUser.Neg()
	}

	if err := insertLayerExecution(ctx, tx, layerExecutionInsert{
		ID:                executionID,
		LayerID:           &layerID,
		MasterSignalID:    job.MasterSignalID,
		UserID:            job.UserID,
		ExchangeBindingID: job.ExchangeBindingID,
		Action:            "sell",
		Symbol:            result.Symbol,
		Quantity:          parsed.FilledQuantity.String(),
		Price:             parsed.FillPrice.String(),
		ValueQuote:        parsed.FillValueQuote.String(),
		ExchangeFee:       parsed.ExchangeFee.String(),
		ExchangeOrderID:   result.ExchangeOrderID,
		Status:            result.Status,
		ErrorCode:         result.ErrorCode,
		ErrorMessage:      result.ErrorMessage,
		IdempotencyKey:    job.Request.ID,
		ExecutedAt:        parsed.ExecutedAt,
	}); err != nil {
		return ExecutionResultApplySummary{}, err
	}

	if _, err := tx.Exec(ctx, `
UPDATE layers
SET remaining_quantity = $2::numeric,
    status = $3,
    closed_at = COALESCE($4::timestamptz, closed_at),
    updated_at = now()
WHERE id = $1::uuid`,
		layerID,
		nextRemaining.String(),
		nextStatus,
		closedAt,
	); err != nil {
		return ExecutionResultApplySummary{}, fmt.Errorf("store: update sell layer: %w", err)
	}

	if _, err := tx.Exec(ctx, `
INSERT INTO gas_fee_ledger (
  id, layer_id, execution_id, user_id, type, gross_pnl, gas_fee_amount,
  platform_rebate, net_amount_user, share_rate, calculated_at
) VALUES (
  $1::uuid, $2::uuid, $3::uuid, $4::uuid, $5, $6::numeric, $7::numeric,
  $8::numeric, $9::numeric, $10::numeric, $11
)`,
		gasFeeLedgerID,
		layerID,
		executionID,
		job.UserID,
		string(gasResult.Type),
		gasResult.GrossPnL.String(),
		gasResult.GasFeeAmount.String(),
		gasResult.PlatformRebate.String(),
		netAmountUser.String(),
		calculator.ShareRate.String(),
		parsed.ExecutedAt,
	); err != nil {
		return ExecutionResultApplySummary{}, fmt.Errorf("store: insert gas fee ledger: %w", err)
	}

	if err := markExecutionJobTerminal(ctx, tx, job.ID, "success", nil); err != nil {
		return ExecutionResultApplySummary{}, err
	}

	return ExecutionResultApplySummary{
		JobID:          job.ID,
		LayerID:        layerID,
		ExecutionID:    executionID,
		GasFeeLedgerID: gasFeeLedgerID,
		Status:         result.Status,
		Side:           "sell",
		AppliedAt:      time.Now().UTC(),
		GasFee: &GasFeeView{
			Type:            string(gasResult.Type),
			GrossPnL:        gasResult.GrossPnL.String(),
			GasFeeAmount:    gasResult.GasFeeAmount.String(),
			PlatformRebate:  gasResult.PlatformRebate.String(),
			NetAmountToUser: gasResult.NetAmountToUser.String(),
			NetLossToUser:   gasResult.NetLossToUser.String(),
		},
	}, nil
}

func lockLayerForSettlement(ctx context.Context, tx pgx.Tx, layerID string) (layerForSettlement, error) {
	var layer layerForSettlement
	err := tx.QueryRow(ctx, `
SELECT id::text, user_id::text, exchange_binding_id::text, master_signal_id::text,
       symbol, entry_price::text, remaining_quantity::text
FROM layers
WHERE id = $1::uuid
  AND status IN ('open', 'partial')
FOR UPDATE`, layerID).Scan(
		&layer.ID,
		&layer.UserID,
		&layer.ExchangeBindingID,
		&layer.MasterSignalID,
		&layer.Symbol,
		&layer.EntryPrice,
		&layer.RemainingQuantity,
	)
	if err != nil {
		return layerForSettlement{}, fmt.Errorf("store: lock layer for settlement: %w", err)
	}
	return layer, nil
}

type layerExecutionInsert struct {
	ID                string
	LayerID           *string
	MasterSignalID    string
	UserID            string
	ExchangeBindingID string
	Action            string
	Symbol            string
	Quantity          string
	Price             string
	ValueQuote        string
	ExchangeFee       string
	ExchangeOrderID   *string
	Status            string
	ErrorCode         *string
	ErrorMessage      *string
	IdempotencyKey    string
	ExecutedAt        time.Time
}

func insertLayerExecution(ctx context.Context, tx pgx.Tx, input layerExecutionInsert) error {
	var layerID any
	if input.LayerID != nil {
		layerID = *input.LayerID
	}

	_, err := tx.Exec(ctx, `
INSERT INTO layer_executions (
  id, layer_id, master_signal_id, user_id, exchange_binding_id, action, symbol,
  quantity, price, value_quote, exchange_fee, exchange_order_id, status,
  error_code, error_message, idempotency_key, executed_at, updated_at
) VALUES (
  $1::uuid, $2::uuid, $3::uuid, $4::uuid, $5::uuid, $6, $7,
  $8::numeric, $9::numeric, $10::numeric, $11::numeric, $12, $13,
  $14, $15, $16, $17, now()
)`,
		input.ID,
		layerID,
		input.MasterSignalID,
		input.UserID,
		input.ExchangeBindingID,
		input.Action,
		input.Symbol,
		input.Quantity,
		input.Price,
		input.ValueQuote,
		input.ExchangeFee,
		input.ExchangeOrderID,
		input.Status,
		input.ErrorCode,
		input.ErrorMessage,
		input.IdempotencyKey,
		input.ExecutedAt,
	)
	if err != nil {
		return fmt.Errorf("store: insert layer execution: %w", err)
	}
	return nil
}

func markExecutionJobTerminal(ctx context.Context, tx pgx.Tx, jobID, status string, maybeError *string) error {
	var lastError any
	if maybeError != nil {
		lastError = *maybeError
	}
	_, err := tx.Exec(ctx, `
UPDATE execution_jobs
SET status = $2,
    last_error = COALESCE($3, last_error),
    updated_at = now()
WHERE id = $1::uuid`, jobID, status, lastError)
	if err != nil {
		return fmt.Errorf("store: mark execution job terminal: %w", err)
	}
	return nil
}

func refreshMasterSignalCompletion(ctx context.Context, tx pgx.Tx, masterSignalID string) error {
	var pendingCount int
	var failedCount int
	var successCount int
	if err := tx.QueryRow(ctx, `
SELECT
  COUNT(*) FILTER (WHERE status IN ('queued', 'published', 'running'))::int,
  COUNT(*) FILTER (WHERE status IN ('failed', 'dead_letter'))::int,
  COUNT(*) FILTER (WHERE status = 'success')::int
FROM execution_jobs
WHERE master_signal_id = $1::uuid`, masterSignalID).Scan(&pendingCount, &failedCount, &successCount); err != nil {
		return fmt.Errorf("store: count execution job terminal state: %w", err)
	}
	if pendingCount > 0 {
		return nil
	}

	status := "completed"
	if successCount == 0 && failedCount > 0 {
		status = "failed"
	}
	_, err := tx.Exec(ctx, `
UPDATE master_signals
SET status = $2,
    completed_at = COALESCE(completed_at, now())
WHERE id = $1::uuid
  AND status = 'dispatching'`, masterSignalID, status)
	if err != nil {
		return fmt.Errorf("store: update master signal completion: %w", err)
	}
	return nil
}

func insertExecutionResultAuditLog(ctx context.Context, tx pgx.Tx, result ExecutionResult, summary ExecutionResultApplySummary) error {
	auditID, err := newUUIDText()
	if err != nil {
		return err
	}
	afterJSON, err := json.Marshal(map[string]any{
		"result":  result,
		"summary": summary,
	})
	if err != nil {
		return fmt.Errorf("store: marshal execution audit: %w", err)
	}
	var entityID any = result.RequestID
	_, err = tx.Exec(ctx, `
INSERT INTO audit_logs (
  id, actor_type, action, entity, entity_id, after_state, created_at
) VALUES (
  $1::uuid, 'system', 'execution_result_applied', 'execution_job', $2::uuid, $3::jsonb, now()
)`,
		auditID,
		entityID,
		string(afterJSON),
	)
	if err != nil {
		return fmt.Errorf("store: insert execution audit log: %w", err)
	}
	return nil
}

func newUUIDText() (string, error) {
	value, err := id.New()
	if err != nil {
		return "", err
	}
	return value.String(), nil
}
