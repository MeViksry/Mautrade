package httpapi

import (
	"net/http"

	"time"

	"github.com/MeViksry/Mautrade/backend/internal/store"
)

func (s *Server) handleGetGlobalSettings(w http.ResponseWriter, r *http.Request) {
	settings, err := s.store.GlobalSettings(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, settings)
}

func (s *Server) handleAdminUpdateGlobalSettings(w http.ResponseWriter, r *http.Request) {
	_, ok := s.requireAdmin(w, r)
	if !ok {
		return
	}

	var req store.UpdateGlobalSettingsParams
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	req.Now = time.Now()

	settings, err := s.store.UpdateGlobalSettings(r.Context(), req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, settings)
}
