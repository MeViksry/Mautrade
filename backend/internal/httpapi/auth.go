package httpapi

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/MeViksry/Mautrade/backend/internal/store"
	"github.com/MeViksry/qdecimal"
)

type registerRequest struct {
	Email           string `json:"email"`
	Username        string `json:"username"`
	DisplayName     string `json:"display_name"`
	Name            string `json:"name"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirm_password"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type verifyEmailOTPRequest struct {
	Email   string `json:"email"`
	Purpose string `json:"purpose"`
	Code    string `json:"code"`
}

type onboardingRequest struct {
	CountryCode         string   `json:"country_code"`
	Timezone            string   `json:"timezone"`
	Age                 int      `json:"age"`
	ExchangePreferences []string `json:"exchange_preferences"`
	Exchanges           []string `json:"exchanges"`
	GasFeeDepositAmount string   `json:"gas_fee_deposit_amount"`
	Amount              string   `json:"amount"`
	GasFeeAsset         string   `json:"gas_fee_asset"`
	TxID                string   `json:"tx_id"`
}

func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	if !s.store.Ready() {
		writeError(w, http.StatusServiceUnavailable, "postgres is required for register")
		return
	}

	var req registerRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := validateRegisterRequest(req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	displayName := strings.TrimSpace(firstNonEmpty(req.DisplayName, req.Name))
	result, err := s.store.RegisterUser(r.Context(), store.RegisterUserParams{
		Email:       req.Email,
		Username:    req.Username,
		DisplayName: displayName,
		Password:    req.Password,
		OTPTTL:      s.config.EmailOTPTTL,
		Now:         time.Now().UTC(),
	})
	if err != nil {
		if errors.Is(err, store.ErrDuplicateAccount) {
			writeError(w, http.StatusConflict, "account already exists")
			return
		}
		s.logger.Error("register user", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to register account")
		return
	}
	if result.DevOTP != "" {
		go func(toEmail, name, otp string) {
			if err := s.mailer.SendOTP(toEmail, name, otp, "register"); err != nil {
				s.logger.Error("failed to send register otp email", "error", err, "email", toEmail)
			}
		}(req.Email, result.User.DisplayName, result.DevOTP)
	}
	maskDevOTP(&result.DevOTP, s.config.Environment)
	writeJSON(w, http.StatusCreated, result)
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	if !s.store.Ready() {
		writeError(w, http.StatusServiceUnavailable, "postgres is required for login")
		return
	}

	var req loginRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := validateLoginRequest(req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	result, err := s.store.LoginUser(r.Context(), store.LoginUserParams{
		Email:      req.Email,
		Password:   req.Password,
		SessionTTL: s.config.AuthSessionTTL,
		OTPTTL:     s.config.EmailOTPTTL,
		UserAgent:  r.UserAgent(),
		IPAddress:  requestIP(r),
		Now:        time.Now().UTC(),
	})
	if err != nil {
		if errors.Is(err, store.ErrInvalidCredential) {
			writeError(w, http.StatusUnauthorized, "invalid email or password")
			return
		}
		s.logger.Error("login user", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to login")
		return
	}
	if result.OTPRequired && result.DevOTP != "" {
		go func(toEmail, name, otp string) {
			if err := s.mailer.SendOTP(toEmail, name, otp, "login"); err != nil {
				s.logger.Error("failed to send login otp email", "error", err, "email", toEmail)
			}
		}(req.Email, result.User.DisplayName, result.DevOTP)
	}
	maskDevOTP(&result.DevOTP, s.config.Environment)
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) handleVerifyEmailOTP(w http.ResponseWriter, r *http.Request) {
	if !s.store.Ready() {
		writeError(w, http.StatusServiceUnavailable, "postgres is required for otp verification")
		return
	}

	var req verifyEmailOTPRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := validateVerifyEmailOTPRequest(req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	result, err := s.store.VerifyEmailOTP(r.Context(), store.VerifyEmailOTPParams{
		Email:      req.Email,
		Purpose:    req.Purpose,
		Code:       req.Code,
		SessionTTL: s.config.AuthSessionTTL,
		UserAgent:  r.UserAgent(),
		IPAddress:  requestIP(r),
		Now:        time.Now().UTC(),
	})
	if err != nil {
		switch {
		case errors.Is(err, store.ErrInvalidOTP):
			writeError(w, http.StatusUnauthorized, "invalid otp")
		case errors.Is(err, store.ErrExpiredOTP):
			writeError(w, http.StatusGone, "otp expired")
		default:
			s.logger.Error("verify email otp", "error", err)
			writeError(w, http.StatusInternalServerError, "failed to verify otp")
		}
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func (s *Server) handleMe(w http.ResponseWriter, r *http.Request) {
	user, err := s.authUserFromRequest(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid or expired session")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"user": user})
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	token, err := bearerToken(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "missing bearer token")
		return
	}
	if err := s.store.LogoutSession(r.Context(), token); err != nil {
		if errors.Is(err, store.ErrInvalidSession) {
			writeError(w, http.StatusUnauthorized, "invalid session")
			return
		}
		s.logger.Error("logout user", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to logout")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"status": "logged_out"})
}

func (s *Server) handleCompleteOnboarding(w http.ResponseWriter, r *http.Request) {
	user, err := s.authUserFromRequest(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid or expired session")
		return
	}

	var req onboardingRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := validateOnboardingRequest(req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	amount := strings.TrimSpace(firstNonEmpty(req.GasFeeDepositAmount, req.Amount))
	result, err := s.store.CompleteOnboarding(r.Context(), store.CompleteOnboardingParams{
		UserID:              user.ID,
		Age:                 req.Age,
		CountryCode:         strings.ToUpper(strings.TrimSpace(req.CountryCode)),
		Timezone:            strings.TrimSpace(req.Timezone),
		ExchangePreferences: firstNonEmptyStringSlice(req.ExchangePreferences, req.Exchanges),
		GasFeeAmount:        amount,
		GasFeeAsset:         firstNonEmpty(req.GasFeeAsset, s.config.DefaultCurrency),
		GasFeeDepositAddr:   s.config.GasFeeDepositAddress,
		TxID:                req.TxID,
		Now:                 time.Now().UTC(),
	})
	if err != nil {
		s.logger.Error("complete onboarding", "user_id", user.ID, "error", err)
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func (s *Server) authUserFromRequest(r *http.Request) (store.AuthUserView, error) {
	token, err := bearerToken(r)
	if err != nil {
		return store.AuthUserView{}, err
	}
	return s.store.AuthenticateSession(r.Context(), token)
}

func bearerToken(r *http.Request) (string, error) {
	header := strings.TrimSpace(r.Header.Get("Authorization"))
	if header == "" {
		return "", fmt.Errorf("missing authorization header")
	}
	parts := strings.Fields(header)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", fmt.Errorf("authorization must be bearer token")
	}
	return parts[1], nil
}

func validateRegisterRequest(req registerRequest) error {
	if !looksLikeEmail(req.Email) {
		return fmt.Errorf("email is required")
	}
	if strings.TrimSpace(firstNonEmpty(req.DisplayName, req.Name)) == "" {
		return fmt.Errorf("name is required")
	}
	if len(req.Password) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}
	if req.ConfirmPassword != "" && req.Password != req.ConfirmPassword {
		return fmt.Errorf("confirm_password does not match password")
	}
	return nil
}

func validateLoginRequest(req loginRequest) error {
	if !looksLikeEmail(req.Email) {
		return fmt.Errorf("email is required")
	}
	if strings.TrimSpace(req.Password) == "" {
		return fmt.Errorf("password is required")
	}
	return nil
}

func validateVerifyEmailOTPRequest(req verifyEmailOTPRequest) error {
	if !looksLikeEmail(req.Email) {
		return fmt.Errorf("email is required")
	}
	if strings.TrimSpace(req.Code) == "" {
		return fmt.Errorf("code is required")
	}
	if req.Purpose != "" && req.Purpose != store.EmailOTPPurposeRegisterVerify && req.Purpose != store.EmailOTPPurposeLoginVerify {
		return fmt.Errorf("purpose must be register_verify or login_verify")
	}
	return nil
}

func validateOnboardingRequest(req onboardingRequest) error {
	countryCode := strings.ToUpper(strings.TrimSpace(req.CountryCode))
	if len(countryCode) != 2 {
		return fmt.Errorf("country_code must be ISO-3166 alpha-2")
	}
	if req.Age < 18 {
		return fmt.Errorf("age must be at least 18")
	}
	if strings.TrimSpace(req.Timezone) == "" {
		return fmt.Errorf("timezone is required")
	}
	if len(firstNonEmptyStringSlice(req.ExchangePreferences, req.Exchanges)) == 0 {
		return fmt.Errorf("at least one exchange preference is required")
	}
	amount := strings.TrimSpace(firstNonEmpty(req.GasFeeDepositAmount, req.Amount))
	parsed, err := qdecimal.Parse(amount)
	if err != nil {
		return fmt.Errorf("gas_fee_deposit_amount must be decimal")
	}
	if parsed.Cmp(qdecimal.MustParse("500")) < 0 {
		return fmt.Errorf("gas_fee_deposit_amount must be at least 500 USDT")
	}
	return nil
}

func looksLikeEmail(value string) bool {
	value = strings.TrimSpace(value)
	return strings.Contains(value, "@") && strings.Contains(value, ".")
}

func requestIP(r *http.Request) string {
	for _, header := range []string{"X-Forwarded-For", "X-Real-IP"} {
		value := strings.TrimSpace(r.Header.Get(header))
		if value != "" {
			return strings.TrimSpace(strings.Split(value, ",")[0])
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	return r.RemoteAddr
}

func firstNonEmptyStringSlice(values ...[]string) []string {
	for _, value := range values {
		if len(value) > 0 {
			return value
		}
	}
	return nil
}

func maskDevOTP(value *string, environment string) {
	if strings.EqualFold(environment, "production") {
		*value = ""
	}
}
