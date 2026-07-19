package httpapi

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/MeViksry/Mautrade/backend/internal/platform/totp"
	"github.com/MeViksry/Mautrade/backend/internal/store"
)

type adminLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	OTPCode  string `json:"otp_code"`
	OtpCode  string `json:"otpCode"`
}

type adminOTPRequest struct {
	Code    string `json:"code"`
	OTPCode string `json:"otp_code"`
	OtpCode string `json:"otpCode"`
}

func (s *Server) handleAdminLogin(w http.ResponseWriter, r *http.Request) {
	if !s.store.Ready() {
		writeError(w, http.StatusServiceUnavailable, "postgres is required for admin login")
		return
	}
	var req adminLoginRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if !looksLikeEmail(req.Email) || strings.TrimSpace(req.Password) == "" {
		writeError(w, http.StatusBadRequest, "email and password are required")
		return
	}

	result, err := s.store.LoginAdmin(r.Context(), store.LoginAdminParams{
		Email:    req.Email,
		Password: req.Password,
		OTPCode:  firstNonEmpty(req.OTPCode, req.OtpCode),
		OTPVerifier: func(ciphertext []byte, code string, now time.Time) bool {
			return s.validateAdminOTPCode(ciphertext, code, now)
		},
		SessionTTL: s.config.AuthSessionTTL,
		UserAgent:  r.UserAgent(),
		IPAddress:  requestIP(r),
		Now:        time.Now().UTC(),
	})
	if err != nil {
		switch {
		case errors.Is(err, store.ErrAdminOTPRequired):
			writeJSON(w, http.StatusAccepted, result)
		case errors.Is(err, store.ErrInvalidAdminCredential), errors.Is(err, store.ErrInvalidOTP):
			writeError(w, http.StatusUnauthorized, "invalid admin credentials")
		default:
			s.logger.Error("admin login", "error", err)
			writeError(w, http.StatusInternalServerError, "failed to login admin")
		}
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) handleAdmin2FASetup(w http.ResponseWriter, r *http.Request) {
	if !s.store.Ready() {
		writeError(w, http.StatusServiceUnavailable, "postgres is required for admin 2fa setup")
		return
	}
	admin, ok := s.requireAdmin(w, r)
	if !ok {
		return
	}
	secret, err := totp.GenerateSecret()
	if err != nil {
		s.logger.Error("generate admin 2fa secret", "admin_id", admin.ID, "error", err)
		writeError(w, http.StatusInternalServerError, "failed to generate 2fa secret")
		return
	}
	ciphertext, err := s.credentialEncryptor.SealString(secret)
	if err != nil {
		s.logger.Error("encrypt admin 2fa secret", "admin_id", admin.ID, "error", err)
		writeError(w, http.StatusInternalServerError, "failed to protect 2fa secret")
		return
	}
	updatedAdmin, err := s.store.StoreAdminOTPSecret(r.Context(), admin.ID, ciphertext, time.Now().UTC())
	if err != nil {
		writeAdminOTPError(s, w, "admin 2fa setup", err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{
		"admin":      updatedAdmin,
		"secret":     secret,
		"otpauthUri": totp.BuildURI("Mautrade Admin", updatedAdmin.Email, secret),
		"status":     "setup_pending_verification",
	})
}

func (s *Server) handleAdmin2FAVerify(w http.ResponseWriter, r *http.Request) {
	if !s.store.Ready() {
		writeError(w, http.StatusServiceUnavailable, "postgres is required for admin 2fa verification")
		return
	}
	admin, ok := s.requireAdmin(w, r)
	if !ok {
		return
	}
	code, ok := decodeAdminOTPRequest(w, r)
	if !ok {
		return
	}
	secretView, err := s.store.AdminOTPSecret(r.Context(), admin.ID)
	if err != nil {
		writeAdminOTPError(s, w, "read admin 2fa secret", err)
		return
	}
	if !s.validateAdminOTPCode(secretView.SecretCiphertext, code, time.Now().UTC()) {
		writeError(w, http.StatusUnauthorized, "invalid otp")
		return
	}
	updatedAdmin, err := s.store.EnableAdminOTP(r.Context(), admin.ID, time.Now().UTC())
	if err != nil {
		writeAdminOTPError(s, w, "enable admin 2fa", err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"admin":  updatedAdmin,
		"status": "2fa_enabled",
	})
}

func (s *Server) handleAdmin2FADisable(w http.ResponseWriter, r *http.Request) {
	if !s.store.Ready() {
		writeError(w, http.StatusServiceUnavailable, "postgres is required for admin 2fa disable")
		return
	}
	admin, ok := s.requireAdmin(w, r)
	if !ok {
		return
	}
	if !admin.OTPEnabled {
		writeError(w, http.StatusConflict, "2fa is not enabled")
		return
	}
	code, ok := decodeAdminOTPRequest(w, r)
	if !ok {
		return
	}
	secretView, err := s.store.AdminOTPSecret(r.Context(), admin.ID)
	if err != nil {
		writeAdminOTPError(s, w, "read admin 2fa secret", err)
		return
	}
	if !s.validateAdminOTPCode(secretView.SecretCiphertext, code, time.Now().UTC()) {
		writeError(w, http.StatusUnauthorized, "invalid otp")
		return
	}
	updatedAdmin, err := s.store.DisableAdminOTP(r.Context(), admin.ID, time.Now().UTC())
	if err != nil {
		writeAdminOTPError(s, w, "disable admin 2fa", err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"admin":  updatedAdmin,
		"status": "2fa_disabled",
	})
}

func (s *Server) handleAdminMe(w http.ResponseWriter, r *http.Request) {
	admin, err := s.adminUserFromRequest(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid or expired admin session")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"admin": admin})
}

func (s *Server) handleAdminLogout(w http.ResponseWriter, r *http.Request) {
	token, err := bearerToken(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "missing bearer token")
		return
	}
	if err := s.store.LogoutAdminSession(r.Context(), token); err != nil {
		if errors.Is(err, store.ErrInvalidSession) {
			writeError(w, http.StatusUnauthorized, "invalid admin session")
			return
		}
		s.logger.Error("admin logout", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to logout admin")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"status": "logged_out"})
}

func (s *Server) adminUserFromRequest(r *http.Request) (store.AdminUserView, error) {
	token, err := bearerToken(r)
	if err != nil {
		return store.AdminUserView{}, err
	}
	return s.store.AuthenticateAdminSession(r.Context(), token)
}

func (s *Server) requireAdmin(w http.ResponseWriter, r *http.Request) (store.AdminUserView, bool) {
	admin, err := s.adminUserFromRequest(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid or expired admin session")
		return store.AdminUserView{}, false
	}
	return admin, true
}

func decodeAdminOTPRequest(w http.ResponseWriter, r *http.Request) (string, bool) {
	var req adminOTPRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return "", false
	}
	code := strings.TrimSpace(firstNonEmpty(req.Code, req.OTPCode, req.OtpCode))
	if code == "" {
		writeError(w, http.StatusBadRequest, "otp code is required")
		return "", false
	}
	return code, true
}

func (s *Server) validateAdminOTPCode(ciphertext []byte, code string, now time.Time) bool {
	secret, err := s.credentialEncryptor.OpenString(ciphertext)
	if err != nil {
		s.logger.Warn("decrypt admin 2fa secret failed", "error", err)
		return false
	}
	return totp.Validate(code, secret, now)
}

func writeAdminOTPError(s *Server, w http.ResponseWriter, operation string, err error) {
	switch {
	case errors.Is(err, store.ErrAdminNotFound):
		writeError(w, http.StatusNotFound, "admin not found")
	case errors.Is(err, store.ErrAdminOTPSecretMissing):
		writeError(w, http.StatusConflict, "2fa setup is required")
	default:
		s.logger.Error(operation, "error", err)
		writeError(w, http.StatusInternalServerError, "admin 2fa operation failed")
	}
}

func adminBodyMustMatchSession(w http.ResponseWriter, provided string, admin store.AdminUserView) bool {
	provided = strings.TrimSpace(provided)
	if provided == "" || provided == admin.ID {
		return true
	}
	writeError(w, http.StatusForbidden, "admin_id does not match session")
	return false
}
