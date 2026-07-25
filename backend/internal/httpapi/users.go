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
		search := r.URL.Query().Get("search")

		users, err := s.store.AdminListUsers(r.Context(), search, limit, offset)
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

func (s *Server) handleAdminDeleteUser(w http.ResponseWriter, r *http.Request) {
	if s.store.Ready() {
		if _, ok := s.requireAdmin(w, r); !ok {
			return
		}
		
		id := r.PathValue("id")
		if id == "" {
			writeError(w, http.StatusBadRequest, "missing user id")
			return
		}

		err := s.store.AdminDeleteUser(r.Context(), id)
		if err != nil {
			s.logger.Error("delete admin user", "error", err, "id", id)
			writeError(w, http.StatusInternalServerError, "failed to delete user")
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"status": "ok"})
		return
	}
	writeError(w, http.StatusServiceUnavailable, "store not ready")
}
