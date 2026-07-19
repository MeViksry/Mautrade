package httpapi

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/MeViksry/Mautrade/backend/internal/domain/id"
	"github.com/MeViksry/Mautrade/backend/internal/platform/queue"
	"github.com/MeViksry/Mautrade/backend/internal/store"
	"github.com/MeViksry/qdecimal"
)

type createAdminSignalRequest struct {
	AdminID        string `json:"admin_id"`
	Type           string `json:"type"`
	Symbol         string `json:"symbol"`
	LayerNumber    *int   `json:"layer_number"`
	AllocationPct  string `json:"allocation_pct"`
	Percentage     string `json:"percentage"`
	SellPct        string `json:"sell_pct"`
	SellPercentage string `json:"sell_percentage"`
	IdempotencyKey string `json:"idempotency_key"`
}

type createAdminSignalResponse struct {
	SignalID        string                 `json:"signalId"`
	Status          string                 `json:"status"`
	IdempotencyKey  string                 `json:"idempotencyKey"`
	JobsCreated     int                    `json:"jobsCreated"`
	JobsSkipped     int                    `json:"jobsSkipped"`
	JobsPublished   int                    `json:"jobsPublished"`
	QueueState      string                 `json:"queueState"`
	PublishFailures []signalPublishFailure `json:"publishFailures,omitempty"`
}

type signalPublishFailure struct {
	JobID string `json:"jobId"`
	Error string `json:"error"`
}

func (s *Server) handleCreateAdminSignal(w http.ResponseWriter, r *http.Request) {
	if !s.store.Ready() {
		writeError(w, http.StatusServiceUnavailable, "postgres is required to create admin signals")
		return
	}
	admin, ok := s.requireAdmin(w, r)
	if !ok {
		return
	}

	var req createAdminSignalRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if !adminBodyMustMatchSession(w, req.AdminID, admin) {
		return
	}

	params, err := s.validateCreateAdminSignalRequest(r, req, admin.ID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	dispatch, err := s.store.CreateSignalDispatch(r.Context(), params)
	if err != nil {
		s.logger.Error("create signal dispatch", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to create signal dispatch")
		return
	}

	published, failures := s.publishSignalJobs(r, dispatch)
	queueState := "ready"
	if s.queue == nil {
		queueState = "unavailable"
	}
	if len(failures) > 0 {
		queueState = "degraded"
	}

	writeJSON(w, http.StatusAccepted, createAdminSignalResponse{
		SignalID:        dispatch.SignalID,
		Status:          dispatch.Status,
		IdempotencyKey:  dispatch.IdempotencyKey,
		JobsCreated:     dispatch.JobsCreated,
		JobsSkipped:     dispatch.JobsSkipped,
		JobsPublished:   published,
		QueueState:      queueState,
		PublishFailures: failures,
	})
}

func (s *Server) validateCreateAdminSignalRequest(r *http.Request, req createAdminSignalRequest, adminID string) (store.CreateSignalParams, error) {
	if _, err := id.Parse(adminID); err != nil {
		return store.CreateSignalParams{}, fmt.Errorf("admin_id must be a canonical UUID")
	}

	signalType := strings.ToLower(strings.TrimSpace(req.Type))
	if signalType != "buy" && signalType != "sell" {
		return store.CreateSignalParams{}, fmt.Errorf("type must be buy or sell")
	}

	symbol, err := normalizeSpotSymbol(req.Symbol)
	if err != nil {
		return store.CreateSignalParams{}, err
	}

	idempotencyKey := strings.TrimSpace(req.IdempotencyKey)
	if idempotencyKey == "" {
		idempotencyKey = strings.TrimSpace(r.Header.Get("Idempotency-Key"))
	}
	if idempotencyKey == "" {
		return store.CreateSignalParams{}, fmt.Errorf("idempotency_key or Idempotency-Key header is required")
	}
	if len(idempotencyKey) > 180 {
		return store.CreateSignalParams{}, fmt.Errorf("idempotency_key is too long")
	}

	allocationPct := firstNonEmpty(req.AllocationPct, req.Percentage)
	sellPct := firstNonEmpty(req.SellPct, req.SellPercentage)
	if signalType == "buy" {
		if _, err := validatePercent(allocationPct, "allocation_pct"); err != nil {
			return store.CreateSignalParams{}, err
		}
		sellPct = "0"
	}
	if signalType == "sell" {
		if req.LayerNumber == nil || *req.LayerNumber <= 0 {
			return store.CreateSignalParams{}, fmt.Errorf("layer_number is required for sell signals")
		}
		if _, err := validatePercent(sellPct, "sell_pct"); err != nil {
			return store.CreateSignalParams{}, err
		}
		allocationPct = "0"
	}

	return store.CreateSignalParams{
		AdminID:        adminID,
		Type:           signalType,
		Symbol:         symbol,
		LayerNumber:    req.LayerNumber,
		AllocationPct:  allocationPct,
		SellPct:        sellPct,
		IdempotencyKey: idempotencyKey,
		DefaultAsset:   s.config.DefaultCurrency,
		CreatedAt:      time.Now().UTC(),
	}, nil
}

func (s *Server) publishSignalJobs(r *http.Request, dispatch store.SignalDispatch) (int, []signalPublishFailure) {
	if s.queue == nil {
		return 0, nil
	}

	var failures []signalPublishFailure
	published := 0
	for _, job := range dispatch.Jobs {
		req := queue.ExecutionRequest{
			ID:             job.Payload.ID,
			IdempotencyKey: job.Payload.IdempotencyKey,
			MasterSignalID: job.Payload.MasterSignalID,
			UserID:         job.Payload.UserID,
			LayerID:        job.Payload.LayerID,
			Exchange:       job.Payload.Exchange,
			Symbol:         job.Payload.Symbol,
			Side:           job.Payload.Side,
			Quantity:       job.Payload.Quantity,
			QuoteValue:     job.Payload.QuoteValue,
			CreatedAt:      job.Payload.CreatedAt,
		}
		if err := s.queue.PublishExecutionRequest(r.Context(), req); err != nil {
			s.logger.Error("publish execution job", "job_id", job.ID, "error", err)
			failures = append(failures, signalPublishFailure{JobID: job.ID, Error: "failed to publish execution job"})
			if markErr := s.store.MarkExecutionJobPublishFailed(r.Context(), job.ID, err.Error()); markErr != nil {
				s.logger.Error("mark execution job publish failure", "job_id", job.ID, "error", markErr)
			}
			continue
		}

		published++
		if err := s.store.MarkExecutionJobPublished(r.Context(), job.ID); err != nil {
			s.logger.Error("mark execution job published", "job_id", job.ID, "error", err)
		}
	}
	return published, failures
}

func normalizeSpotSymbol(value string) (string, error) {
	symbol := strings.ToUpper(strings.TrimSpace(value))
	if symbol == "" {
		return "", fmt.Errorf("symbol is required")
	}
	parts := strings.Split(symbol, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", fmt.Errorf("symbol must be a spot pair like BTC/USDT")
	}
	return symbol, nil
}

func validatePercent(value, field string) (qdecimal.Decimal, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return qdecimal.Decimal{}, fmt.Errorf("%s is required", field)
	}
	percent, err := qdecimal.Parse(value)
	if err != nil {
		return qdecimal.Decimal{}, fmt.Errorf("%s must be a decimal percentage", field)
	}
	if percent.Sign() <= 0 || percent.Cmp(qdecimal.MustParse("100")) > 0 {
		return qdecimal.Decimal{}, fmt.Errorf("%s must be greater than 0 and at most 100", field)
	}
	return percent, nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
