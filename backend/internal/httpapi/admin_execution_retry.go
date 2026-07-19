package httpapi

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/MeViksry/Mautrade/backend/internal/domain/id"
	"github.com/MeViksry/Mautrade/backend/internal/platform/queue"
	"github.com/MeViksry/Mautrade/backend/internal/store"
)

type retryExecutionRequest struct {
	AdminID string `json:"admin_id"`
}

type retryExecutionResponse struct {
	JobID          string `json:"jobId"`
	MasterSignalID string `json:"masterSignalId"`
	PreviousStatus string `json:"previousStatus"`
	Status         string `json:"status"`
	Attempt        int    `json:"attempt"`
	RetryRequestID string `json:"retryRequestId"`
	MessageID      string `json:"messageId"`
	QueueState     string `json:"queueState"`
}

func (s *Server) handleRetryAdminExecution(w http.ResponseWriter, r *http.Request) {
	if !s.store.Ready() {
		writeError(w, http.StatusServiceUnavailable, "postgres is required to retry execution jobs")
		return
	}
	if s.queue == nil {
		writeError(w, http.StatusServiceUnavailable, "execution queue is unavailable")
		return
	}
	admin, ok := s.requireAdmin(w, r)
	if !ok {
		return
	}

	jobID := strings.TrimSpace(r.PathValue("job_id"))
	if _, err := id.Parse(jobID); err != nil {
		writeError(w, http.StatusBadRequest, "job_id must be a canonical UUID")
		return
	}

	var req retryExecutionRequest
	if err := decodeOptionalJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if !adminBodyMustMatchSession(w, req.AdminID, admin) {
		return
	}

	retry, err := s.store.PrepareExecutionJobRetry(r.Context(), jobID, admin.ID, time.Now().UTC())
	if err != nil {
		switch {
		case errors.Is(err, store.ErrExecutionJobNotFound):
			writeError(w, http.StatusNotFound, "execution job not found")
		case errors.Is(err, store.ErrExecutionRetryNotAllowed):
			writeError(w, http.StatusConflict, "execution job status cannot be retried")
		default:
			s.logger.Error("prepare execution retry", "job_id", jobID, "error", err)
			writeError(w, http.StatusInternalServerError, "failed to prepare execution retry")
		}
		return
	}

	publishRequest := queue.ExecutionRequest{
		ID:             retry.Payload.ID,
		IdempotencyKey: retry.Payload.IdempotencyKey,
		MasterSignalID: retry.Payload.MasterSignalID,
		UserID:         retry.Payload.UserID,
		LayerID:        retry.Payload.LayerID,
		Exchange:       retry.Payload.Exchange,
		Symbol:         retry.Payload.Symbol,
		Side:           retry.Payload.Side,
		Quantity:       retry.Payload.Quantity,
		QuoteValue:     retry.Payload.QuoteValue,
		CreatedAt:      retry.Payload.CreatedAt,
	}
	if err := s.queue.PublishExecutionRequestWithMsgID(r.Context(), publishRequest, retry.MessageID); err != nil {
		s.logger.Error("publish execution retry", "job_id", retry.JobID, "attempt", retry.Attempt, "error", err)
		if markErr := s.store.MarkExecutionJobPublishFailed(r.Context(), retry.JobID, err.Error()); markErr != nil {
			s.logger.Error("mark execution retry publish failure", "job_id", retry.JobID, "error", markErr)
		}
		writeError(w, http.StatusBadGateway, "failed to publish execution retry")
		return
	}
	if err := s.store.MarkExecutionJobPublished(r.Context(), retry.JobID); err != nil {
		s.logger.Error("mark execution retry published", "job_id", retry.JobID, "error", err)
		writeError(w, http.StatusInternalServerError, "execution retry published but status update failed")
		return
	}

	writeJSON(w, http.StatusAccepted, retryExecutionResponse{
		JobID:          retry.JobID,
		MasterSignalID: retry.MasterSignalID,
		PreviousStatus: retry.PreviousStatus,
		Status:         "published",
		Attempt:        retry.Attempt,
		RetryRequestID: retry.RetryRequestID,
		MessageID:      retry.MessageID,
		QueueState:     "ready",
	})
}
