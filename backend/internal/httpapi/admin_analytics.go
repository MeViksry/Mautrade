package httpapi

import (
	"net/http"
)

func (s *Server) handleAdminGetAnalytics(w http.ResponseWriter, r *http.Request) {
	data, err := s.store.AdminGetAnalytics(r.Context())
	if err != nil {
		s.logger.Error("failed to get admin analytics", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to get analytics")
		return
	}

	writeJSON(w, http.StatusOK, data)
}
