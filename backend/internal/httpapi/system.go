package httpapi

import (
	"io"
	"net/http"
	"strings"
)

var cachedServerIP string

func (s *Server) handleGetServerIP(w http.ResponseWriter, r *http.Request) {
	if cachedServerIP == "" {
		resp, err := http.Get("https://api.ipify.org")
		if err == nil {
			defer resp.Body.Close()
			body, err := io.ReadAll(resp.Body)
			if err == nil {
				cachedServerIP = strings.TrimSpace(string(body))
			}
		}
	}

	if cachedServerIP == "" {
		writeError(w, http.StatusInternalServerError, "Failed to get server IP")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"ip": cachedServerIP})
}
