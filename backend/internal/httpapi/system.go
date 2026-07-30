package httpapi

import (
	"io"
	"net/http"
	"strings"
)

var cachedServerIP string

func (s *Server) handleGetServerIP(w http.ResponseWriter, r *http.Request) {
	if cachedServerIP == "" {
		cachedServerIP = "109.123.233.56" // Default VPS IP

		resp, err := http.Get("https://api.ipify.org")
		if err == nil {
			defer resp.Body.Close()
			body, err := io.ReadAll(resp.Body)
			if err == nil {
				fetchedIP := strings.TrimSpace(string(body))
				if fetchedIP != "" {
					cachedServerIP = fetchedIP
				}
			}
		}
	}

	writeJSON(w, http.StatusOK, map[string]string{"ip": cachedServerIP})
}
