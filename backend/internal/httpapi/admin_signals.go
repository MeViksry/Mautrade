package httpapi

import (
	"net/http"
)

func (s *Server) handleAdminListActiveSignals(w http.ResponseWriter, r *http.Request) {
	if !s.store.Ready() {
		writeError(w, http.StatusServiceUnavailable, "postgres is required to read signals")
		return
	}
	if _, ok := s.requireAdmin(w, r); !ok {
		return
	}

	signals, err := s.store.AdminListActiveSignals(
		r.Context(),
		positiveIntQuery(r, "limit", 50, 500),
		nonNegativeIntQuery(r, "offset", 0),
	)
	if err != nil {
		s.logger.Error("list active signals", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to read signals")
		return
	}

	writeJSON(w, http.StatusOK, signals)
}

func (s *Server) handleAdminListOpenOrders(w http.ResponseWriter, r *http.Request) {
	if !s.store.Ready() {
		writeError(w, http.StatusServiceUnavailable, "postgres is required to read open orders")
		return
	}
	if _, ok := s.requireAdmin(w, r); !ok {
		return
	}

	orders, err := s.store.AdminListOpenOrders(
		r.Context(),
		positiveIntQuery(r, "limit", 50, 500),
		nonNegativeIntQuery(r, "offset", 0),
	)
	if err != nil {
		s.logger.Error("list open orders", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to read open orders")
		return
	}

	writeJSON(w, http.StatusOK, orders)
}
