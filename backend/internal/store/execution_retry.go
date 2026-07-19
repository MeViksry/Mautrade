package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

var (
	ErrExecutionJobNotFound     = errors.New("store: execution job not found")
	ErrExecutionRetryNotAllowed = errors.New("store: execution retry not allowed")
)

type ExecutionJobRetry struct {
	JobID          string           `json:"jobId"`
	MasterSignalID string           `json:"masterSignalId"`
	PreviousStatus string           `json:"previousStatus"`
	Status         string           `json:"status"`
	Attempt        int              `json:"attempt"`
	RetryRequestID string           `json:"retryRequestId"`
	MessageID      string           `json:"messageId"`
	Payload        ExecutionPayload `json:"payload"`
}

func (s *DashboardStore) PrepareExecutionJobRetry(ctx context.Context, jobID, adminID string, now time.Time) (ExecutionJobRetry, error) {
	if !s.Ready() {
		return ExecutionJobRetry{}, fmt.Errorf("store: retry execution job requires postgres")
	}
	now = normalizedNow(now)

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return ExecutionJobRetry{}, fmt.Errorf("store: begin retry execution job: %w", err)
	}
	defer tx.Rollback(ctx)

	var retry ExecutionJobRetry
	var payloadText string
	var idempotencyKey string
	if err := tx.QueryRow(ctx, `
SELECT id::text, master_signal_id::text, status, attempts, payload::text, idempotency_key
FROM execution_jobs
WHERE id = $1::uuid
FOR UPDATE`, jobID).Scan(
		&retry.JobID,
		&retry.MasterSignalID,
		&retry.PreviousStatus,
		&retry.Attempt,
		&payloadText,
		&idempotencyKey,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ExecutionJobRetry{}, ErrExecutionJobNotFound
		}
		return ExecutionJobRetry{}, fmt.Errorf("store: lock retry execution job: %w", err)
	}

	switch retry.PreviousStatus {
	case "failed", "dead_letter":
	default:
		return ExecutionJobRetry{}, ErrExecutionRetryNotAllowed
	}

	if err := json.Unmarshal([]byte(payloadText), &retry.Payload); err != nil {
		return ExecutionJobRetry{}, fmt.Errorf("store: decode retry execution payload: %w", err)
	}
	retryRequestID, err := newUUIDText()
	if err != nil {
		return ExecutionJobRetry{}, err
	}
	retry.Attempt++
	retry.RetryRequestID = retryRequestID
	retry.MessageID = fmt.Sprintf("%s:retry:%d", retry.JobID, retry.Attempt)
	retry.Status = "queued"
	retry.Payload.ID = retryRequestID
	retry.Payload.IdempotencyKey = idempotencyKey
	retry.Payload.CreatedAt = now.Format(time.RFC3339Nano)

	payloadJSON, err := json.Marshal(retry.Payload)
	if err != nil {
		return ExecutionJobRetry{}, fmt.Errorf("store: marshal retry execution payload: %w", err)
	}
	if _, err := tx.Exec(ctx, `
UPDATE execution_jobs
SET payload = $2::jsonb,
    status = 'queued',
    attempts = $3,
    last_error = NULL,
    updated_at = $4
WHERE id = $1::uuid`,
		retry.JobID,
		string(payloadJSON),
		retry.Attempt,
		now,
	); err != nil {
		return ExecutionJobRetry{}, fmt.Errorf("store: update retry execution job: %w", err)
	}
	if err := insertExecutionRetryAudit(ctx, tx, adminID, retry, now); err != nil {
		return ExecutionJobRetry{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return ExecutionJobRetry{}, fmt.Errorf("store: commit retry execution job: %w", err)
	}
	return retry, nil
}

func insertExecutionRetryAudit(ctx context.Context, tx pgx.Tx, adminID string, retry ExecutionJobRetry, now time.Time) error {
	auditID, err := newUUIDText()
	if err != nil {
		return err
	}
	afterJSON, err := json.Marshal(map[string]any{
		"previous_status":  retry.PreviousStatus,
		"status":           retry.Status,
		"attempt":          retry.Attempt,
		"retry_request_id": retry.RetryRequestID,
		"message_id":       retry.MessageID,
		"master_signal_id": retry.MasterSignalID,
	})
	if err != nil {
		return fmt.Errorf("store: marshal execution retry audit: %w", err)
	}
	if _, err := tx.Exec(ctx, `
INSERT INTO audit_logs (
  id, actor_type, actor_id, action, entity, entity_id, after_state, created_at
) VALUES (
  $1::uuid, 'admin', $2::uuid, 'execution_job_retry_requested', 'execution_job', $3::uuid, $4::jsonb, $5
)`, auditID, adminID, retry.JobID, string(afterJSON), now); err != nil {
		return fmt.Errorf("store: insert execution retry audit: %w", err)
	}
	return nil
}
