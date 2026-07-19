package httpapi

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/MeViksry/Mautrade/backend/internal/platform/secrets"
	"github.com/MeViksry/Mautrade/backend/internal/store"
)

type bindExchangeRequest struct {
	Exchange        string `json:"exchange"`
	APIKey          string `json:"api_key"`
	APISecret       string `json:"api_secret"`
	APIPassphrase   string `json:"api_passphrase"`
	Passphrase      string `json:"passphrase"`
	PermissionScope string `json:"permission_scope"`
}

type updateExchangeBindingStatusRequest struct {
	Status string `json:"status"`
}

type exchangeBindingCredentialResponse struct {
	ID              string     `json:"id"`
	Exchange        string     `json:"exchange"`
	Status          string     `json:"status"`
	MaskedAPIKey    string     `json:"maskedApiKey"`
	HasAPISecret    bool       `json:"hasApiSecret"`
	HasPassphrase   bool       `json:"hasPassphrase"`
	PermissionScope string     `json:"permissionScope"`
	LastVerifiedAt  *time.Time `json:"lastVerifiedAt,omitempty"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
}

func (s *Server) handleBindExchange(w http.ResponseWriter, r *http.Request) {
	if !s.store.Ready() {
		writeError(w, http.StatusServiceUnavailable, "postgres is required to bind exchange")
		return
	}
	user, err := s.authUserFromRequest(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid or expired session")
		return
	}

	var req bindExchangeRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := validateBindExchangeRequest(req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	apiKeyCiphertext, err := s.credentialEncryptor.SealString(strings.TrimSpace(req.APIKey))
	if err != nil {
		s.logger.Error("encrypt exchange api key", "exchange", strings.ToLower(strings.TrimSpace(req.Exchange)), "error", err)
		writeError(w, http.StatusInternalServerError, "failed to protect api key")
		return
	}
	apiSecretCiphertext, err := s.credentialEncryptor.SealString(strings.TrimSpace(req.APISecret))
	if err != nil {
		s.logger.Error("encrypt exchange api secret", "exchange", strings.ToLower(strings.TrimSpace(req.Exchange)), "error", err)
		writeError(w, http.StatusInternalServerError, "failed to protect api secret")
		return
	}
	var passphraseCiphertext []byte
	passphrase := strings.TrimSpace(firstNonEmpty(req.APIPassphrase, req.Passphrase))
	if passphrase != "" {
		passphraseCiphertext, err = s.credentialEncryptor.SealString(passphrase)
		if err != nil {
			s.logger.Error("encrypt exchange api passphrase", "exchange", strings.ToLower(strings.TrimSpace(req.Exchange)), "error", err)
			writeError(w, http.StatusInternalServerError, "failed to protect api passphrase")
			return
		}
	}

	binding, err := s.store.UpsertExchangeBinding(r.Context(), store.UpsertExchangeBindingParams{
		UserID:                  user.ID,
		ExchangeName:            req.Exchange,
		APIKeyCiphertext:        apiKeyCiphertext,
		APISecretCiphertext:     apiSecretCiphertext,
		APIPassphraseCiphertext: passphraseCiphertext,
		PermissionScope:         req.PermissionScope,
		Now:                     time.Now().UTC(),
	})
	if err != nil {
		writeExchangeBindingError(s, w, "bind exchange", err)
		return
	}

	response, err := s.exchangeBindingCredentialResponse(binding)
	if err != nil {
		s.logger.Error("prepare exchange credential response", "binding_id", binding.ID, "error", err)
		writeError(w, http.StatusInternalServerError, "failed to read protected exchange credential")
		return
	}
	writeJSON(w, http.StatusCreated, response)
}

func (s *Server) handleExchangeBindingCredentials(w http.ResponseWriter, r *http.Request) {
	if !s.store.Ready() {
		writeError(w, http.StatusServiceUnavailable, "postgres is required to read exchange credentials")
		return
	}
	user, err := s.authUserFromRequest(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid or expired session")
		return
	}
	binding, err := s.store.ExchangeBindingCredential(r.Context(), user.ID, r.PathValue("exchange"))
	if err != nil {
		writeExchangeBindingError(s, w, "read exchange credential", err)
		return
	}
	response, err := s.exchangeBindingCredentialResponse(binding)
	if err != nil {
		s.logger.Error("prepare exchange credential response", "binding_id", binding.ID, "error", err)
		writeError(w, http.StatusInternalServerError, "failed to read protected exchange credential")
		return
	}
	writeJSON(w, http.StatusOK, response)
}

func (s *Server) handleUpdateExchangeBindingStatus(w http.ResponseWriter, r *http.Request) {
	if !s.store.Ready() {
		writeError(w, http.StatusServiceUnavailable, "postgres is required to update exchange binding")
		return
	}
	user, err := s.authUserFromRequest(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid or expired session")
		return
	}
	var req updateExchangeBindingStatusRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if strings.TrimSpace(req.Status) == "" {
		writeError(w, http.StatusBadRequest, "status is required")
		return
	}
	binding, err := s.store.UpdateExchangeBindingStatus(r.Context(), user.ID, r.PathValue("exchange"), req.Status, time.Now().UTC())
	if err != nil {
		writeExchangeBindingError(s, w, "update exchange binding status", err)
		return
	}
	response, err := s.exchangeBindingCredentialResponse(binding)
	if err != nil {
		s.logger.Error("prepare exchange credential response", "binding_id", binding.ID, "error", err)
		writeError(w, http.StatusInternalServerError, "failed to read protected exchange credential")
		return
	}
	writeJSON(w, http.StatusOK, response)
}

func (s *Server) handleDeleteExchangeBinding(w http.ResponseWriter, r *http.Request) {
	if !s.store.Ready() {
		writeError(w, http.StatusServiceUnavailable, "postgres is required to delete exchange binding")
		return
	}
	user, err := s.authUserFromRequest(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid or expired session")
		return
	}
	binding, err := s.store.UpdateExchangeBindingStatus(r.Context(), user.ID, r.PathValue("exchange"), "revoked", time.Now().UTC())
	if err != nil {
		writeExchangeBindingError(s, w, "delete exchange binding", err)
		return
	}
	response, err := s.exchangeBindingCredentialResponse(binding)
	if err != nil {
		s.logger.Error("prepare exchange credential response", "binding_id", binding.ID, "error", err)
		writeError(w, http.StatusInternalServerError, "failed to read protected exchange credential")
		return
	}
	writeJSON(w, http.StatusOK, response)
}

func validateBindExchangeRequest(req bindExchangeRequest) error {
	exchange := strings.ToLower(strings.TrimSpace(req.Exchange))
	switch exchange {
	case "binance", "bybit", "tokocrypto":
	case "okx":
		if strings.TrimSpace(firstNonEmpty(req.APIPassphrase, req.Passphrase)) == "" {
			return fmt.Errorf("api_passphrase is required for OKX")
		}
	default:
		return fmt.Errorf("exchange must be binance, okx, bybit, or tokocrypto")
	}
	if strings.TrimSpace(req.APIKey) == "" {
		return fmt.Errorf("api_key is required")
	}
	if strings.TrimSpace(req.APISecret) == "" {
		return fmt.Errorf("api_secret is required")
	}
	return nil
}

func (s *Server) exchangeBindingCredentialResponse(binding store.ExchangeBindingCredentialCiphertext) (exchangeBindingCredentialResponse, error) {
	apiKey, err := s.credentialEncryptor.OpenString(binding.APIKeyCiphertext)
	if err != nil {
		return exchangeBindingCredentialResponse{}, err
	}
	return exchangeBindingCredentialResponse{
		ID:              binding.ID,
		Exchange:        binding.ExchangeName,
		Status:          binding.Status,
		MaskedAPIKey:    secrets.Mask(apiKey),
		HasAPISecret:    len(binding.APISecretCiphertext) > 0,
		HasPassphrase:   len(binding.APIPassphraseCiphertext) > 0,
		PermissionScope: binding.PermissionScope,
		LastVerifiedAt:  binding.LastVerifiedAt,
		CreatedAt:       binding.CreatedAt,
		UpdatedAt:       binding.UpdatedAt,
	}, nil
}

func writeExchangeBindingError(s *Server, w http.ResponseWriter, operation string, err error) {
	switch {
	case errors.Is(err, store.ErrUnsupportedExchange):
		writeError(w, http.StatusBadRequest, "unsupported exchange")
	case errors.Is(err, store.ErrInvalidExchangeStatus):
		writeError(w, http.StatusBadRequest, "invalid exchange status")
	case errors.Is(err, store.ErrExchangeBindingNotFound):
		writeError(w, http.StatusNotFound, "exchange binding not found")
	default:
		s.logger.Error(operation, "error", err)
		writeError(w, http.StatusInternalServerError, "exchange binding operation failed")
	}
}
