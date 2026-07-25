package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/MeViksry/Mautrade/backend/internal/store"
)

func (s *Server) handleExecutionResult(w http.ResponseWriter, r *http.Request) {
	if !s.store.Ready() {
		writeError(w, http.StatusServiceUnavailable, "postgres is required to apply execution results")
		return
	}

	var result store.ExecutionResult
	if err := decodeJSON(r, &result); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	calc, err := s.getGasFeeCalculator(r.Context())
	if err != nil {
		s.logger.Error("get gas fee calculator", "error", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	summary, err := s.store.ApplyExecutionResult(r.Context(), result, calc)
	if err != nil {
		s.logger.Error("apply execution result", "error", err)
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusAccepted, summary)
}

func (s *Server) StartExecutionResultConsumer(ctx context.Context) (func(), error) {
	if s.queue == nil || !s.store.Ready() {
		return func() {}, nil
	}

	return s.queue.ConsumeExecutionResults(ctx, func(ctx context.Context, data []byte) error {
		var result store.ExecutionResult
		if err := json.Unmarshal(data, &result); err != nil {
			return fmt.Errorf("decode execution result: %w", err)
		}

		calc, err := s.getGasFeeCalculator(ctx)
		if err != nil {
			return fmt.Errorf("get gas fee calculator: %w", err)
		}

		summary, err := s.store.ApplyExecutionResult(ctx, result, calc)
		if err != nil {
			return err
		}
		s.logger.Info("execution result applied", "job_id", summary.JobID, "side", summary.Side, "status", summary.Status, "duplicate", summary.Duplicate)
		return nil
	})
}
