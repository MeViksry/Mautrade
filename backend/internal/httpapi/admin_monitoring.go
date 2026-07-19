package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/MeViksry/Mautrade/backend/internal/domain/id"
	"github.com/MeViksry/Mautrade/backend/internal/store"
	"github.com/jackc/pgx/v5"
)

type resolveReconciliationEventRequest struct {
	AdminID        string `json:"admin_id"`
	ActorID        string `json:"actor_id"`
	ResolutionNote string `json:"resolution_note"`
}

func (s *Server) handleAdminSignalExecutions(w http.ResponseWriter, r *http.Request) {
	if !s.store.Ready() {
		writeError(w, http.StatusServiceUnavailable, "postgres is required to read signal executions")
		return
	}
	if _, ok := s.requireAdmin(w, r); !ok {
		return
	}

	signalID := strings.TrimSpace(r.PathValue("signal_id"))
	if _, err := id.Parse(signalID); err != nil {
		writeError(w, http.StatusBadRequest, "signal_id must be a canonical UUID")
		return
	}

	view, err := s.store.AdminSignalExecutions(
		r.Context(),
		signalID,
		positiveIntQuery(r, "limit", 100, 500),
		nonNegativeIntQuery(r, "offset", 0),
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "signal not found")
			return
		}
		s.logger.Error("read admin signal executions", "signal_id", signalID, "error", err)
		writeError(w, http.StatusInternalServerError, "failed to read signal executions")
		return
	}

	writeJSON(w, http.StatusOK, view)
}

func (s *Server) handleAdminReconciliationEvents(w http.ResponseWriter, r *http.Request) {
	if !s.store.Ready() {
		writeError(w, http.StatusServiceUnavailable, "postgres is required to read reconciliation events")
		return
	}
	if _, ok := s.requireAdmin(w, r); !ok {
		return
	}

	status, err := normalizeReconciliationStatus(r.URL.Query().Get("status"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	view, err := s.store.AdminReconciliationEvents(r.Context(), store.AdminReconciliationEventsParams{
		Status: status,
		Limit:  positiveIntQuery(r, "limit", 100, 500),
		Offset: nonNegativeIntQuery(r, "offset", 0),
	})
	if err != nil {
		s.logger.Error("read reconciliation events", "status", status, "error", err)
		writeError(w, http.StatusInternalServerError, "failed to read reconciliation events")
		return
	}

	writeJSON(w, http.StatusOK, view)
}

func (s *Server) handleResolveReconciliationEvent(w http.ResponseWriter, r *http.Request) {
	if !s.store.Ready() {
		writeError(w, http.StatusServiceUnavailable, "postgres is required to resolve reconciliation events")
		return
	}
	admin, ok := s.requireAdmin(w, r)
	if !ok {
		return
	}

	eventID := strings.TrimSpace(r.PathValue("event_id"))
	if _, err := id.Parse(eventID); err != nil {
		writeError(w, http.StatusBadRequest, "event_id must be a canonical UUID")
		return
	}

	var req resolveReconciliationEventRequest
	if err := decodeOptionalJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	actorID := strings.TrimSpace(firstNonEmpty(req.AdminID, req.ActorID))
	if !adminBodyMustMatchSession(w, actorID, admin) {
		return
	}

	event, err := s.store.ResolveReconciliationEvent(r.Context(), store.ResolveReconciliationEventParams{
		EventID:        eventID,
		ActorID:        admin.ID,
		ResolutionNote: strings.TrimSpace(req.ResolutionNote),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "reconciliation event not found")
			return
		}
		s.logger.Error("resolve reconciliation event", "event_id", eventID, "error", err)
		writeError(w, http.StatusInternalServerError, "failed to resolve reconciliation event")
		return
	}

	writeJSON(w, http.StatusOK, event)
}

func positiveIntQuery(r *http.Request, name string, fallback, max int) int {
	value := nonNegativeIntQuery(r, name, fallback)
	if value <= 0 {
		return fallback
	}
	if value > max {
		return max
	}
	return value
}

func nonNegativeIntQuery(r *http.Request, name string, fallback int) int {
	raw := strings.TrimSpace(r.URL.Query().Get(name))
	if raw == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 0 {
		return fallback
	}
	return value
}

func normalizeReconciliationStatus(value string) (string, error) {
	status := strings.ToLower(strings.TrimSpace(value))
	if status == "" {
		return "open", nil
	}
	if status == "all" {
		return "", nil
	}
	if status != "open" && status != "resolved" && status != "ignored" {
		return "", errors.New("status must be open, resolved, ignored, or all")
	}
	return status, nil
}

func decodeOptionalJSON(r *http.Request, target any) error {
	if r.Body == nil {
		return nil
	}
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(target); err != nil && !errors.Is(err, io.EOF) {
		return err
	}
	return nil
}
