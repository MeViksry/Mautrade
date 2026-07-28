package httpapi

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/MeViksry/Mautrade/backend/internal/platform/exchangebalance"
	"github.com/MeViksry/Mautrade/backend/internal/platform/secrets"
	"github.com/MeViksry/Mautrade/backend/internal/store"
)

type bindExchangeRequest struct {
	Exchange        string `json:"exchange"`
	APIKey          string `json:"api_key"`
	APISecret       string `json:"api_secret"`
	APIPassphrase   string `json:"api_passphrase"`
	Passphrase      string `json:"passphrase"`
	AccountMode     string `json:"account_mode"`
	AccountModeAlt  string `json:"accountMode"`
	PermissionScope string `json:"permission_scope"`
}

type updateExchangeBindingStatusRequest struct {
	Status string `json:"status"`
}

type updateExchangeBindingAccountModeRequest struct {
	AccountMode    string `json:"account_mode"`
	AccountModeAlt string `json:"accountMode"`
}

type exchangeBindingCredentialResponse struct {
	ID              string     `json:"id"`
	Exchange        string     `json:"exchange"`
	AccountMode     string     `json:"accountMode"`
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
	accountMode, err := normalizeBindExchangeAccountMode(req)
	if err != nil {
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
		AccountMode:             accountMode,
		PermissionScope:         req.PermissionScope,
		Now:                     time.Now().UTC(),
	})
	if err != nil {
		writeExchangeBindingError(s, w, "bind exchange", err)
		return
	}
	syncCtx, cancel := contextWithExchangeBalanceTimeout(r.Context())
	err = s.syncExchangeBindingBalanceWithPlaintext(
		syncCtx,
		user.ID,
		binding.ID,
		binding.ExchangeName,
		binding.AccountMode,
		strings.TrimSpace(req.APIKey),
		strings.TrimSpace(req.APISecret),
		passphrase,
	)
	cancel()
	if err != nil {
		s.logger.Warn("verify exchange balance during bind", "user_id", user.ID, "binding_id", binding.ID, "exchange", binding.ExchangeName, "account_mode", binding.AccountMode, "error", err)
		if _, statusErr := s.store.UpdateExchangeBindingStatus(r.Context(), user.ID, binding.ExchangeName, "invalid", time.Now().UTC()); statusErr != nil {
			s.logger.Warn("mark exchange binding invalid", "user_id", user.ID, "binding_id", binding.ID, "exchange", binding.ExchangeName, "error", statusErr)
		}
		writeError(w, http.StatusBadGateway, "failed to read USDT spot balance from exchange; check API permissions and account mode")
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

func (s *Server) handleUpdateExchangeBindingAccountMode(w http.ResponseWriter, r *http.Request) {
	if !s.store.Ready() {
		writeError(w, http.StatusServiceUnavailable, "postgres is required to update exchange binding")
		return
	}
	user, err := s.authUserFromRequest(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid or expired session")
		return
	}
	var req updateExchangeBindingAccountModeRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	accountMode, err := normalizeExchangeAccountModeValue(firstNonEmpty(req.AccountMode, req.AccountModeAlt))
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	binding, err := s.store.ExchangeBindingCredential(r.Context(), user.ID, r.PathValue("exchange"))
	if err != nil {
		writeExchangeBindingError(s, w, "read exchange credential for account mode update", err)
		return
	}
	if binding.ExchangeName == "tokocrypto" && accountMode == exchangebalance.AccountModeDemo {
		writeError(w, http.StatusBadRequest, "demo account is not supported for Tokocrypto")
		return
	}
	syncCtx, cancel := contextWithExchangeBalanceTimeout(r.Context())
	err = s.syncExchangeBindingBalance(syncCtx, user.ID, store.ExchangeBindingCredentialCiphertext{
		ID:                      binding.ID,
		ExchangeName:            binding.ExchangeName,
		AccountMode:             accountMode,
		Status:                  binding.Status,
		APIKeyCiphertext:        binding.APIKeyCiphertext,
		APISecretCiphertext:     binding.APISecretCiphertext,
		APIPassphraseCiphertext: binding.APIPassphraseCiphertext,
		PermissionScope:         binding.PermissionScope,
		LastVerifiedAt:          binding.LastVerifiedAt,
		CreatedAt:               binding.CreatedAt,
		UpdatedAt:               binding.UpdatedAt,
	})
	cancel()
	if err != nil {
		s.logger.Warn("update exchange account mode sync failed", "user_id", user.ID, "binding_id", binding.ID, "exchange", binding.ExchangeName, "account_mode", accountMode, "error", err)
		writeError(w, http.StatusBadGateway, "failed to read USDT spot balance with selected account mode")
		return
	}
	updatedBinding, err := s.store.ExchangeBindingCredential(r.Context(), user.ID, binding.ExchangeName)
	if err != nil {
		writeExchangeBindingError(s, w, "read updated exchange credential", err)
		return
	}
	response, err := s.exchangeBindingCredentialResponse(updatedBinding)
	if err != nil {
		s.logger.Error("prepare exchange credential response", "binding_id", updatedBinding.ID, "error", err)
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
	accountMode, err := normalizeExchangeAccountModeValue(firstNonEmpty(req.AccountMode, req.AccountModeAlt))
	if err != nil {
		return err
	}
	switch exchange {
	case "binance", "bybit", "tokocrypto":
		if exchange == "tokocrypto" && accountMode == exchangebalance.AccountModeDemo {
			return fmt.Errorf("demo account is not supported for Tokocrypto")
		}
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

func normalizeBindExchangeAccountMode(req bindExchangeRequest) (string, error) {
	return normalizeExchangeAccountModeValue(firstNonEmpty(req.AccountMode, req.AccountModeAlt))
}

func normalizeExchangeAccountModeValue(value string) (string, error) {
	mode := exchangebalance.NormalizeAccountMode(value)
	switch mode {
	case exchangebalance.AccountModeReal, exchangebalance.AccountModeDemo:
		return mode, nil
	default:
		return "", fmt.Errorf("account_mode must be real or demo")
	}
}

func contextWithExchangeBalanceTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, exchangeBalanceSyncTimeout)
}

func (s *Server) exchangeBindingCredentialResponse(binding store.ExchangeBindingCredentialCiphertext) (exchangeBindingCredentialResponse, error) {
	apiKey, err := s.credentialEncryptor.OpenString(binding.APIKeyCiphertext)
	if err != nil {
		return exchangeBindingCredentialResponse{}, err
	}
	return exchangeBindingCredentialResponse{
		ID:              binding.ID,
		Exchange:        binding.ExchangeName,
		AccountMode:     binding.AccountMode,
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
	case errors.Is(err, store.ErrInvalidExchangeAccountMode):
		writeError(w, http.StatusBadRequest, "invalid exchange account mode")
	case errors.Is(err, store.ErrExchangeBindingNotFound):
		writeError(w, http.StatusNotFound, "exchange binding not found")
	default:
		s.logger.Error(operation, "error", err)
		writeError(w, http.StatusInternalServerError, "exchange binding operation failed")
	}
}
