package store

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"regexp"
	"strings"
	"time"

	"github.com/MeViksry/Mautrade/backend/internal/domain/id"
	"github.com/MeViksry/qdecimal"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"
)

const (
	EmailOTPPurposeRegisterVerify = "register_verify"
	EmailOTPPurposeLoginVerify    = "login_verify"
)

var (
	ErrDuplicateAccount  = errors.New("store: account already exists")
	ErrUserNotFound      = errors.New("store: user not found")
	ErrInvalidCredential = errors.New("store: invalid credentials")
	ErrInvalidOTP        = errors.New("store: invalid otp")
	ErrExpiredOTP        = errors.New("store: expired otp")
	ErrInvalidSession    = errors.New("store: invalid session")
)

type AuthUserView struct {
	ID                  string     `json:"id"`
	Email               string     `json:"email"`
	Username            string     `json:"username"`
	DisplayName         string     `json:"displayName"`
	Age                 *int       `json:"age,omitempty"`
	CountryCode         *string    `json:"countryCode,omitempty"`
	Timezone            string     `json:"timezone"`
	ThemePref           string     `json:"themePref"`
	EmailVerified       bool       `json:"emailVerified"`
	OnboardingCompleted bool       `json:"onboardingCompleted"`
	CreatedAt           time.Time  `json:"createdAt"`
	EmailVerifiedAt     *time.Time `json:"emailVerifiedAt,omitempty"`
	OnboardingAt        *time.Time `json:"onboardingCompletedAt,omitempty"`
}

type RegisterUserParams struct {
	Email       string
	Username    string
	DisplayName string
	Password    string
	OTPTTL      time.Duration
	Now         time.Time
}

type RegisterUserResult struct {
	User         AuthUserView `json:"user"`
	OTPRequired  bool         `json:"otpRequired"`
	OTPExpiresAt time.Time    `json:"otpExpiresAt"`
	DevOTP       string       `json:"-"`
}

type LoginUserParams struct {
	Email      string
	Password   string
	SessionTTL time.Duration
	OTPTTL     time.Duration
	UserAgent  string
	IPAddress  string
	Now        time.Time
}

type LoginUserResult struct {
	User         AuthUserView `json:"user"`
	Session      *SessionView `json:"session,omitempty"`
	OTPRequired  bool         `json:"otpRequired"`
	OTPExpiresAt *time.Time   `json:"otpExpiresAt,omitempty"`
	DevOTP       string       `json:"-"`
}

type VerifyEmailOTPParams struct {
	Email      string
	Purpose    string
	Code       string
	SessionTTL time.Duration
	UserAgent  string
	IPAddress  string
	Now        time.Time
}

type VerifyEmailOTPResult struct {
	User    AuthUserView `json:"user"`
	Session SessionView  `json:"session"`
}

type SessionView struct {
	ID        string    `json:"id"`
	Token     string    `json:"token,omitempty"`
	ExpiresAt time.Time `json:"expiresAt"`
	CreatedAt time.Time `json:"createdAt"`
}

type CompleteOnboardingParams struct {
	UserID              string
	Age                 int
	CountryCode         string
	Timezone            string
	ExchangePreferences []string
	GasFeeAmount        string
	GasFeeAsset         string
	GasFeeDepositAddr   string
	TxID                string
	Now                 time.Time
}

type CompleteOnboardingResult struct {
	User          AuthUserView `json:"user"`
	DepositID     string       `json:"depositId"`
	DepositAmount string       `json:"depositAmount"`
	DepositAsset  string       `json:"depositAsset"`
	DepositStatus string       `json:"depositStatus"`
}

type authUserRow struct {
	User            AuthUserView
	PasswordHash    string
	Status          string
	EmailVerifiedAt *time.Time
	OnboardingAt    *time.Time
	RawCountryCode  *string
}

func (s *DashboardStore) RegisterUser(ctx context.Context, params RegisterUserParams) (RegisterUserResult, error) {
	if !s.Ready() {
		return RegisterUserResult{}, fmt.Errorf("store: register requires postgres")
	}

	now := normalizedNow(params.Now)
	email := normalizeEmail(params.Email)
	userID, err := id.New()
	if err != nil {
		return RegisterUserResult{}, err
	}
	userIDText := userID.String()
	username := normalizeUsername(params.Username, email, userIDText)
	displayName := strings.TrimSpace(params.DisplayName)
	if displayName == "" {
		displayName = username
	}
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(params.Password), bcrypt.DefaultCost)
	if err != nil {
		return RegisterUserResult{}, fmt.Errorf("store: hash password: %w", err)
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return RegisterUserResult{}, fmt.Errorf("store: begin register: %w", err)
	}
	defer tx.Rollback(ctx)

	row, errFetch := userByEmailForAuth(ctx, tx, email)
	if errFetch == nil {
		if row.OnboardingAt == nil {
			// Allow re-registration by updating password and name for accounts that haven't completed onboarding
			if _, errUpdate := tx.Exec(ctx, `
UPDATE users 
SET password_hash = $2, display_name = $3, updated_at = $4
WHERE id = $1::uuid`, row.User.ID, string(passwordHash), displayName, now); errUpdate != nil {
				return RegisterUserResult{}, fmt.Errorf("store: update unverified user: %w", errUpdate)
			}
			userIDText = row.User.ID
		} else {
			return RegisterUserResult{}, ErrDuplicateAccount
		}
	} else if errors.Is(errFetch, pgx.ErrNoRows) {
		if _, err := tx.Exec(ctx, `
INSERT INTO users (
  id, email, username, password_hash, display_name, timezone, status, created_at, updated_at
) VALUES (
  $1::uuid, $2, $3, $4, $5, 'UTC', 'active', $6, $6
)`,
			userIDText,
			email,
			username,
			string(passwordHash),
			displayName,
			now,
		); err != nil {
			if isUniqueViolation(err) {
				return RegisterUserResult{}, ErrDuplicateAccount
			}
			return RegisterUserResult{}, fmt.Errorf("store: insert user: %w", err)
		}
	} else {
		return RegisterUserResult{}, fmt.Errorf("store: check existing user: %w", errFetch)
	}

	code, expiresAt, err := createEmailOTP(ctx, tx, userIDText, email, EmailOTPPurposeRegisterVerify, params.OTPTTL, now)
	if err != nil {
		return RegisterUserResult{}, err
	}

	user, err := userByIDForAuth(ctx, tx, userIDText)
	if err != nil {
		return RegisterUserResult{}, err
	}
	if err := insertAuthAudit(ctx, tx, "user_register_started", userIDText, map[string]any{"email": email}); err != nil {
		return RegisterUserResult{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return RegisterUserResult{}, fmt.Errorf("store: commit register: %w", err)
	}

	return RegisterUserResult{
		User:         user,
		OTPRequired:  true,
		OTPExpiresAt: expiresAt,
		DevOTP:       code,
	}, nil
}

func (s *DashboardStore) LoginUser(ctx context.Context, params LoginUserParams) (LoginUserResult, error) {
	if !s.Ready() {
		return LoginUserResult{}, fmt.Errorf("store: login requires postgres")
	}

	now := normalizedNow(params.Now)
	email := normalizeEmail(params.Email)

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return LoginUserResult{}, fmt.Errorf("store: begin login: %w", err)
	}
	defer tx.Rollback(ctx)

	row, err := userByEmailForAuth(ctx, tx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return LoginUserResult{}, ErrUserNotFound
		}
		return LoginUserResult{}, err
	}
	if row.Status != "active" {
		return LoginUserResult{}, ErrInvalidCredential
	}
	if err := bcrypt.CompareHashAndPassword([]byte(row.PasswordHash), []byte(params.Password)); err != nil {
		return LoginUserResult{}, ErrInvalidCredential
	}
	code, expiresAt, err := createEmailOTP(ctx, tx, row.User.ID, row.User.Email, EmailOTPPurposeLoginVerify, params.OTPTTL, now)
	if err != nil {
		return LoginUserResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return LoginUserResult{}, fmt.Errorf("store: commit login otp: %w", err)
	}
	return LoginUserResult{
		User:         row.User,
		OTPRequired:  true,
		OTPExpiresAt: &expiresAt,
		DevOTP:       code,
	}, nil
}

func (s *DashboardStore) VerifyEmailOTP(ctx context.Context, params VerifyEmailOTPParams) (VerifyEmailOTPResult, error) {
	if !s.Ready() {
		return VerifyEmailOTPResult{}, fmt.Errorf("store: verify otp requires postgres")
	}

	now := normalizedNow(params.Now)
	email := normalizeEmail(params.Email)
	purpose := strings.TrimSpace(params.Purpose)
	if purpose == "" {
		purpose = EmailOTPPurposeRegisterVerify
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return VerifyEmailOTPResult{}, fmt.Errorf("store: begin verify otp: %w", err)
	}
	defer tx.Rollback(ctx)

	var otpID string
	var userID string
	var codeHash string
	var attempts int
	var expiresAt time.Time
	if err := tx.QueryRow(ctx, `
SELECT id::text, user_id::text, code_hash, attempts, expires_at
FROM auth_email_otps
WHERE lower(email) = $1
  AND purpose = $2
  AND consumed_at IS NULL
ORDER BY created_at DESC
LIMIT 1
FOR UPDATE`, email, purpose).Scan(&otpID, &userID, &codeHash, &attempts, &expiresAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return VerifyEmailOTPResult{}, ErrInvalidOTP
		}
		return VerifyEmailOTPResult{}, fmt.Errorf("store: lock otp: %w", err)
	}
	if attempts >= 5 {
		return VerifyEmailOTPResult{}, ErrInvalidOTP
	}
	if !now.Before(expiresAt) {
		return VerifyEmailOTPResult{}, ErrExpiredOTP
	}
	if err := bcrypt.CompareHashAndPassword([]byte(codeHash), []byte(strings.TrimSpace(params.Code))); err != nil {
		if _, updateErr := tx.Exec(ctx, `
UPDATE auth_email_otps
SET attempts = attempts + 1
WHERE id = $1::uuid`, otpID); updateErr != nil {
			return VerifyEmailOTPResult{}, fmt.Errorf("store: count otp attempt: %w", updateErr)
		}
		return VerifyEmailOTPResult{}, ErrInvalidOTP
	}

	if _, err := tx.Exec(ctx, `
UPDATE auth_email_otps
SET consumed_at = $2
WHERE id = $1::uuid`, otpID, now); err != nil {
		return VerifyEmailOTPResult{}, fmt.Errorf("store: consume otp: %w", err)
	}
	if purpose == EmailOTPPurposeRegisterVerify || purpose == EmailOTPPurposeLoginVerify {
		if _, err := tx.Exec(ctx, `
UPDATE users
SET email_verified_at = COALESCE(email_verified_at, $2),
    updated_at = $2
WHERE id = $1::uuid`, userID, now); err != nil {
			return VerifyEmailOTPResult{}, fmt.Errorf("store: mark email verified: %w", err)
		}
	}

	user, err := userByIDForAuth(ctx, tx, userID)
	if err != nil {
		return VerifyEmailOTPResult{}, err
	}
	session, err := createSession(ctx, tx, user.ID, params.SessionTTL, params.UserAgent, params.IPAddress, now)
	if err != nil {
		return VerifyEmailOTPResult{}, err
	}
	if err := insertAuthAudit(ctx, tx, "user_email_otp_verified", user.ID, map[string]any{"purpose": purpose, "session_id": session.ID}); err != nil {
		return VerifyEmailOTPResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return VerifyEmailOTPResult{}, fmt.Errorf("store: commit verify otp: %w", err)
	}

	return VerifyEmailOTPResult{User: user, Session: session}, nil
}

func (s *DashboardStore) AuthenticateSession(ctx context.Context, rawToken string) (AuthUserView, error) {
	if !s.Ready() {
		return AuthUserView{}, fmt.Errorf("store: authenticate session requires postgres")
	}
	tokenHash := hashSessionToken(rawToken)
	if tokenHash == "" {
		return AuthUserView{}, ErrInvalidSession
	}

	var userID string
	var sessionID string
	if err := s.db.QueryRow(ctx, `
SELECT s.id::text, u.id::text
FROM auth_sessions s
JOIN users u ON u.id = s.user_id
WHERE s.token_hash = $1
  AND s.revoked_at IS NULL
  AND s.expires_at > now()
  AND u.status = 'active'`, tokenHash).Scan(&sessionID, &userID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return AuthUserView{}, ErrInvalidSession
		}
		return AuthUserView{}, fmt.Errorf("store: authenticate session: %w", err)
	}

	_, _ = s.db.Exec(ctx, `
UPDATE auth_sessions
SET last_seen_at = now()
WHERE id = $1::uuid`, sessionID)

	return s.AuthUser(ctx, userID)
}

func (s *DashboardStore) LogoutSession(ctx context.Context, rawToken string) error {
	if !s.Ready() {
		return fmt.Errorf("store: logout requires postgres")
	}
	tokenHash := hashSessionToken(rawToken)
	if tokenHash == "" {
		return ErrInvalidSession
	}
	_, err := s.db.Exec(ctx, `
UPDATE auth_sessions
SET revoked_at = COALESCE(revoked_at, now())
WHERE token_hash = $1`, tokenHash)
	if err != nil {
		return fmt.Errorf("store: revoke session: %w", err)
	}
	return nil
}

func (s *DashboardStore) AuthUser(ctx context.Context, userID string) (AuthUserView, error) {
	if !s.Ready() {
		return AuthUserView{}, fmt.Errorf("store: auth user requires postgres")
	}
	return userByIDForAuth(ctx, s.db, userID)
}

func (s *DashboardStore) CompleteOnboarding(ctx context.Context, params CompleteOnboardingParams) (CompleteOnboardingResult, error) {
	if !s.Ready() {
		return CompleteOnboardingResult{}, fmt.Errorf("store: onboarding requires postgres")
	}

	now := normalizedNow(params.Now)
	asset := strings.ToUpper(strings.TrimSpace(params.GasFeeAsset))
	if asset == "" {
		asset = "USDT"
	}
	amount := decimalOrZero(params.GasFeeAmount)
	settings, err := s.GlobalSettings(ctx)
	if err != nil {
		return CompleteOnboardingResult{}, fmt.Errorf("store: fetch settings for onboarding: %w", err)
	}

	if parsed, err := qdecimal.Parse(amount); err != nil {
		return CompleteOnboardingResult{}, fmt.Errorf("store: gas fee amount must be decimal: %w", err)
	} else if parsed.Cmp(settings.MinDepositUsdt) < 0 {
		return CompleteOnboardingResult{}, fmt.Errorf("store: gas fee amount must be at least %s %s", settings.MinDepositUsdt.String(), asset)
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return CompleteOnboardingResult{}, fmt.Errorf("store: begin onboarding: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `
UPDATE users
SET age = $2,
    country_code = $3,
    timezone = $4,
    onboarding_completed_at = COALESCE(onboarding_completed_at, $5),
    updated_at = $5
WHERE id = $1::uuid`, params.UserID, params.Age, strings.ToUpper(params.CountryCode), params.Timezone, now); err != nil {
		return CompleteOnboardingResult{}, fmt.Errorf("store: update onboarding profile: %w", err)
	}

	if _, err := tx.Exec(ctx, `DELETE FROM user_exchange_preferences WHERE user_id = $1::uuid`, params.UserID); err != nil {
		return CompleteOnboardingResult{}, fmt.Errorf("store: clear exchange preferences: %w", err)
	}
	seen := map[string]bool{}
	for _, exchange := range params.ExchangePreferences {
		exchangeName := strings.ToLower(strings.TrimSpace(exchange))
		if exchangeName == "" || seen[exchangeName] {
			continue
		}
		seen[exchangeName] = true
		prefID, err := id.New()
		if err != nil {
			return CompleteOnboardingResult{}, err
		}
		if _, err := tx.Exec(ctx, `
INSERT INTO user_exchange_preferences (
  id, user_id, exchange_name, selected_at
) VALUES (
  $1::uuid, $2::uuid, $3, $4
)`, prefID.String(), params.UserID, exchangeName, now); err != nil {
			return CompleteOnboardingResult{}, fmt.Errorf("store: insert exchange preference: %w", err)
		}
	}

	depositID, err := id.New()
	if err != nil {
		return CompleteOnboardingResult{}, err
	}
	var txID any
	if strings.TrimSpace(params.TxID) != "" {
		txID = strings.TrimSpace(params.TxID)
	}
	status := "pending"
	if _, err := tx.Exec(ctx, `
INSERT INTO gas_fee_deposits (
  id, user_id, amount, asset, deposit_address, tx_id, status, created_at
) VALUES (
  $1::uuid, $2::uuid, $3::numeric, $4, $5, $6, $7, $8
)`,
		depositID.String(),
		params.UserID,
		amount,
		asset,
		params.GasFeeDepositAddr,
		txID,
		status,
		now,
	); err != nil {
		return CompleteOnboardingResult{}, fmt.Errorf("store: insert onboarding gas fee deposit: %w", err)
	}

	user, err := userByIDForAuth(ctx, tx, params.UserID)
	if err != nil {
		return CompleteOnboardingResult{}, err
	}
	if err := insertAuthAudit(ctx, tx, "user_onboarding_completed", params.UserID, map[string]any{
		"country_code": strings.ToUpper(params.CountryCode),
		"age":          params.Age,
		"timezone":     params.Timezone,
		"deposit_id":   depositID.String(),
	}); err != nil {
		return CompleteOnboardingResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return CompleteOnboardingResult{}, fmt.Errorf("store: commit onboarding: %w", err)
	}

	return CompleteOnboardingResult{
		User:          user,
		DepositID:     depositID.String(),
		DepositAmount: amount,
		DepositAsset:  asset,
		DepositStatus: status,
	}, nil
}

func createEmailOTP(ctx context.Context, tx pgx.Tx, userID, email, purpose string, ttl time.Duration, now time.Time) (string, time.Time, error) {
	if ttl <= 0 {
		ttl = 10 * time.Minute
	}
	if _, err := tx.Exec(ctx, `
UPDATE auth_email_otps
SET consumed_at = COALESCE(consumed_at, $4)
WHERE user_id = $1::uuid
  AND email = $2
  AND purpose = $3
  AND consumed_at IS NULL`, userID, email, purpose, now); err != nil {
		return "", time.Time{}, fmt.Errorf("store: invalidate previous otp: %w", err)
	}

	code, err := randomNumericCode(6)
	if err != nil {
		return "", time.Time{}, err
	}
	codeHash, err := bcrypt.GenerateFromPassword([]byte(code), bcrypt.DefaultCost)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("store: hash otp: %w", err)
	}
	otpID, err := id.New()
	if err != nil {
		return "", time.Time{}, err
	}
	expiresAt := now.Add(ttl)
	if _, err := tx.Exec(ctx, `
INSERT INTO auth_email_otps (
  id, user_id, email, purpose, code_hash, expires_at, created_at
) VALUES (
  $1::uuid, $2::uuid, $3, $4, $5, $6, $7
)`, otpID.String(), userID, email, purpose, string(codeHash), expiresAt, now); err != nil {
		return "", time.Time{}, fmt.Errorf("store: insert otp: %w", err)
	}
	return code, expiresAt, nil
}

func createSession(ctx context.Context, tx pgx.Tx, userID string, ttl time.Duration, userAgent, ipAddress string, now time.Time) (SessionView, error) {
	if ttl <= 0 {
		ttl = 720 * time.Hour
	}
	sessionID, err := id.New()
	if err != nil {
		return SessionView{}, err
	}
	rawToken, tokenHash, err := newSessionToken()
	if err != nil {
		return SessionView{}, err
	}
	expiresAt := now.Add(ttl)
	if _, err := tx.Exec(ctx, `
INSERT INTO auth_sessions (
  id, user_id, token_hash, user_agent, ip_address, expires_at, created_at, last_seen_at
) VALUES (
  $1::uuid, $2::uuid, $3, $4, $5, $6, $7, $7
)`, sessionID.String(), userID, tokenHash, userAgent, ipAddress, expiresAt, now); err != nil {
		return SessionView{}, fmt.Errorf("store: insert session: %w", err)
	}
	return SessionView{
		ID:        sessionID.String(),
		Token:     rawToken,
		ExpiresAt: expiresAt,
		CreatedAt: now,
	}, nil
}

type authQueryer interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}

func userByEmailForAuth(ctx context.Context, queryer authQueryer, email string) (authUserRow, error) {
	return scanAuthUserRow(queryer.QueryRow(ctx, authUserSelectSQL()+`
WHERE lower(email) = $1`, email))
}

func userByIDForAuth(ctx context.Context, queryer authQueryer, userID string) (AuthUserView, error) {
	row, err := scanAuthUserRow(queryer.QueryRow(ctx, authUserSelectSQL()+`
WHERE id = $1::uuid`, userID))
	if err != nil {
		return AuthUserView{}, err
	}
	return row.User, nil
}

func authUserSelectSQL() string {
	return `
SELECT id::text, email, username, display_name, country_code::text, timezone, theme_pref,
       email_verified_at, onboarding_completed_at, password_hash, status, created_at, age
FROM users
`
}

func scanAuthUserRow(row pgx.Row) (authUserRow, error) {
	var result authUserRow
	if err := row.Scan(
		&result.User.ID,
		&result.User.Email,
		&result.User.Username,
		&result.User.DisplayName,
		&result.RawCountryCode,
		&result.User.Timezone,
		&result.User.ThemePref,
		&result.EmailVerifiedAt,
		&result.OnboardingAt,
		&result.PasswordHash,
		&result.Status,
		&result.User.CreatedAt,
		&result.User.Age,
	); err != nil {
		return authUserRow{}, err
	}
	if result.RawCountryCode != nil {
		countryCode := strings.TrimSpace(*result.RawCountryCode)
		result.User.CountryCode = &countryCode
	}
	result.User.EmailVerifiedAt = result.EmailVerifiedAt
	result.User.OnboardingAt = result.OnboardingAt
	result.User.EmailVerified = result.EmailVerifiedAt != nil
	result.User.OnboardingCompleted = result.OnboardingAt != nil
	return result, nil
}

func insertAuthAudit(ctx context.Context, tx pgx.Tx, action, userID string, after map[string]any) error {
	auditID, err := id.New()
	if err != nil {
		return err
	}
	afterJSON, err := json.Marshal(after)
	if err != nil {
		return fmt.Errorf("store: marshal auth audit: %w", err)
	}
	_, err = tx.Exec(ctx, `
INSERT INTO audit_logs (
  id, actor_type, actor_id, action, entity, entity_id, after_state, created_at
) VALUES (
  $1::uuid, 'user', $2::uuid, $3, 'user', $2::uuid, $4::jsonb, now()
)`, auditID.String(), userID, action, string(afterJSON))
	if err != nil {
		return fmt.Errorf("store: insert auth audit: %w", err)
	}
	return nil
}

func randomNumericCode(length int) (string, error) {
	if length <= 0 {
		length = 6
	}
	max := big.NewInt(1)
	ten := big.NewInt(10)
	for range length {
		max.Mul(max, ten)
	}
	value, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", fmt.Errorf("store: generate otp: %w", err)
	}
	return fmt.Sprintf("%0*d", length, value.Int64()), nil
}

func newSessionToken() (string, string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", "", fmt.Errorf("store: generate session token: %w", err)
	}
	rawToken := base64.RawURLEncoding.EncodeToString(bytes)
	return rawToken, hashSessionToken(rawToken), nil
}

func hashSessionToken(rawToken string) string {
	rawToken = strings.TrimSpace(rawToken)
	if rawToken == "" {
		return ""
	}
	sum := sha256.Sum256([]byte(rawToken))
	return hex.EncodeToString(sum[:])
}

func normalizeEmail(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

var usernameCleaner = regexp.MustCompile(`[^a-z0-9_]+`)

func normalizeUsername(value, email, userID string) string {
	username := strings.ToLower(strings.TrimSpace(value))
	if username == "" {
		username = strings.Split(email, "@")[0]
	}
	username = usernameCleaner.ReplaceAllString(username, "_")
	username = strings.Trim(username, "_")
	if username == "" {
		username = "user"
	}
	if len(username) > 32 {
		username = username[:32]
	}
	if value == "" && len(userID) >= 8 {
		username = fmt.Sprintf("%s_%s", username, strings.ReplaceAll(userID[:8], "-", ""))
	}
	return username
}

func normalizedNow(value time.Time) time.Time {
	if value.IsZero() {
		return time.Now().UTC()
	}
	return value.UTC()
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
