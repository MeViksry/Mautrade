package httpapi

import (
	"net/http"
)

func (s *Server) handleAdminListUsers(w http.ResponseWriter, r *http.Request) {
	if s.store.Ready() {
		if _, ok := s.requireAdmin(w, r); !ok {
			return
		}
		limit := positiveIntQuery(r, "limit", 50, 100)
		offset := nonNegativeIntQuery(r, "offset", 0)

		users, err := s.store.AdminListUsers(r.Context(), limit, offset)
		if err != nil {
			s.logger.Error("read admin users", "error", err)
			writeError(w, http.StatusInternalServerError, "failed to read users")
			return
		}
		writeJSON(w, http.StatusOK, users)
		return
	}

	writeJSON(w, http.StatusOK, []map[string]any{})
}
