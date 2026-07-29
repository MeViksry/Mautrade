package httpapi

import (
	"crypto/subtle"
	"net/http"
	"strings"
)

const internalTokenHeader = "X-Mautrade-Internal-Token"

type internalExchangeCredentialResponse struct {
	ID              string `json:"id"`
	Exchange        string `json:"exchange"`
	AccountMode     string `json:"accountMode"`
	Status          string `json:"status"`
	APIKey          string `json:"apiKey"`
	APISecret       string `json:"apiSecret"`
	APIPassphrase   string `json:"apiPassphrase,omitempty"`
	PermissionScope string `json:"permissionScope"`
}

func (s *Server) requireInternalToken(w http.ResponseWriter, r *http.Request) bool {
	expected := strings.TrimSpace(s.config.ExecutionInternalToken)
	if expected == "" {
		writeError(w, http.StatusServiceUnavailable, "internal execution token is not configured")
		return false
	}
	provided := strings.TrimSpace(r.Header.Get(internalTokenHeader))
	if provided == "" {
		if token, err := bearerToken(r); err == nil {
			provided = strings.TrimSpace(token)
		}
	}
	if provided == "" || subtle.ConstantTimeCompare([]byte(provided), []byte(expected)) != 1 {
		writeError(w, http.StatusUnauthorized, "invalid internal execution token")
		return false
	}
	return true
}

func (s *Server) handleInternalExchangeBindingCredential(w http.ResponseWriter, r *http.Request) {
	if !s.requireInternalToken(w, r) {
		return
	}
	if !s.store.Ready() {
		writeError(w, http.StatusServiceUnavailable, "postgres is required to read exchange credential")
		return
	}

	binding, err := s.store.ExchangeBindingCredentialByID(r.Context(), r.PathValue("binding_id"))
	if err != nil {
		writeExchangeBindingError(s, w, "read internal exchange credential", err)
		return
	}

	apiKey, err := s.credentialEncryptor.OpenString(binding.APIKeyCiphertext)
	if err != nil {
		s.logger.Error("decrypt internal exchange api key", "binding_id", binding.ID, "exchange", binding.ExchangeName, "error", err)
		writeError(w, http.StatusInternalServerError, "failed to read protected exchange credential")
		return
	}
	apiSecret, err := s.credentialEncryptor.OpenString(binding.APISecretCiphertext)
	if err != nil {
		s.logger.Error("decrypt internal exchange api secret", "binding_id", binding.ID, "exchange", binding.ExchangeName, "error", err)
		writeError(w, http.StatusInternalServerError, "failed to read protected exchange credential")
		return
	}
	passphrase := ""
	if len(binding.APIPassphraseCiphertext) > 0 {
		passphrase, err = s.credentialEncryptor.OpenString(binding.APIPassphraseCiphertext)
		if err != nil {
			s.logger.Error("decrypt internal exchange api passphrase", "binding_id", binding.ID, "exchange", binding.ExchangeName, "error", err)
			writeError(w, http.StatusInternalServerError, "failed to read protected exchange credential")
			return
		}
	}

	writeJSON(w, http.StatusOK, internalExchangeCredentialResponse{
		ID:              binding.ID,
		Exchange:        binding.ExchangeName,
		AccountMode:     binding.AccountMode,
		Status:          binding.Status,
		APIKey:          apiKey,
		APISecret:       apiSecret,
		APIPassphrase:   passphrase,
		PermissionScope: binding.PermissionScope,
	})
}
