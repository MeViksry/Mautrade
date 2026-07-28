package httpapi

import (
	"net/http"

	"github.com/MeViksry/qdecimal"
)

func (s *Server) handleUserStats(w http.ResponseWriter, r *http.Request) {
	user, err := s.authUserFromRequest(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid or expired session")
		return
	}
	s.syncDueUserExchangeBalances(r.Context(), user.ID)
	stats, err := s.store.UserStats(r.Context(), user.ID, s.config.DefaultCurrency)
	if err != nil {
		s.logger.Error("read user stats", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to read user stats")
		return
	}
	writeJSON(w, http.StatusOK, stats)
}

func (s *Server) handleExchangeBindings(w http.ResponseWriter, r *http.Request) {
	user, err := s.authUserFromRequest(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid or expired session")
		return
	}
	s.syncDueUserExchangeBalances(r.Context(), user.ID)
	bindings, err := s.store.UserExchangeBindings(r.Context(), user.ID, s.config.DefaultCurrency)
	if err != nil {
		s.logger.Error("read exchange bindings", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to read exchange bindings")
		return
	}
	writeJSON(w, http.StatusOK, bindings)
}

func (s *Server) handleLayers(w http.ResponseWriter, r *http.Request) {
	user, err := s.authUserFromRequest(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid or expired session")
		return
	}
	layers, err := s.store.ActiveLayers(r.Context(), user.ID)
	if err != nil {
		s.logger.Error("read active layers", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to read active layers")
		return
	}
	writeJSON(w, http.StatusOK, layers)
}

func (s *Server) handleTradeHistory(w http.ResponseWriter, r *http.Request) {
	user, err := s.authUserFromRequest(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid or expired session")
		return
	}
	history, err := s.store.TradeHistory(r.Context(), user.ID)
	if err != nil {
		s.logger.Error("read trade history", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to read trade history")
		return
	}
	writeJSON(w, http.StatusOK, history)
}

func (s *Server) handleAdminOverview(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.requireAdmin(w, r); !ok {
		return
	}
	overview, err := s.store.AdminOverview(r.Context(), s.config.DefaultCurrency)
	if err != nil {
		s.logger.Error("read admin overview", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to read admin overview")
		return
	}
	writeJSON(w, http.StatusOK, overview)
}

type gasFeePreviewRequest struct {
	EntryValue string `json:"entry_value"`
	ExitValue  string `json:"exit_value"`
}

func (s *Server) handleGasFeePreview(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.requireAdmin(w, r); !ok {
		return
	}
	var req gasFeePreviewRequest
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	entry, err := qdecimal.Parse(req.EntryValue)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid entry_value"})
		return
	}
	exit, err := qdecimal.Parse(req.ExitValue)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid exit_value"})
		return
	}
	calc, err := s.getGasFeeCalculator(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to get gas fee settings"})
		return
	}
	writeJSON(w, http.StatusOK, calc.CalculateFromValues(entry, exit))
}
