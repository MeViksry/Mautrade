package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/MeViksry/Mautrade/backend/internal/domain/id"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidAdminCredential = errors.New("store: invalid admin credentials")
	ErrAdminOTPRequired       = errors.New("store: admin otp required")
	ErrAdminOTPSecretMissing  = errors.New("store: admin otp secret missing")
	ErrAdminNotFound          = errors.New("store: admin not found")
	ErrInvalidAdminRole       = errors.New("store: invalid admin role")
)

type AdminUserView struct {
	ID          string     `json:"id"`
	Email       string     `json:"email"`
	DisplayName string     `json:"displayName"`
	Role        string     `json:"role"`
	OTPEnabled  bool       `json:"otpEnabled"`
	Status      string     `json:"status"`
	LastLogin   *time.Time `json:"lastLoginAt,omitempty"`
	CreatedAt   time.Time  `json:"createdAt"`
}

type BootstrapAdminParams struct {
	Email       string
	Password    string
	DisplayName string
	Role        string
	Now         time.Time
}

type BootstrapAdminResult struct {
	Admin   AdminUserView `json:"admin"`
	Created bool          `json:"created"`
}

type LoginAdminParams struct {
	Email       string
	Password    string
	OTPCode     string
	OTPVerifier func(ciphertext []byte, code string, now time.Time) bool
	SessionTTL  time.Duration
	UserAgent   string
	IPAddress   string
	Now         time.Time
}

type LoginAdminResult struct {
	Admin        AdminUserView `json:"admin"`
	Session      *SessionView  `json:"session,omitempty"`
	OTPRequired  bool          `json:"otpRequired"`
	OTPChallenge string        `json:"otpChallenge,omitempty"`
}

type AdminOTPSecretView struct {
	Admin            AdminUserView `json:"admin"`
	SecretCiphertext []byte        `json:"-"`
}

func (s *DashboardStore) BootstrapAdmin(ctx context.Context, params BootstrapAdminParams) (BootstrapAdminResult, error) {
	if !s.Ready() {
		return BootstrapAdminResult{}, fmt.Errorf("store: bootstrap admin requires postgres")
	}
	email := normalizeEmail(params.Email)
	if email == "" {
		return BootstrapAdminResult{}, nil
	}
	password := strings.TrimSpace(params.Password)
	if len(password) < 12 {
		return BootstrapAdminResult{}, fmt.Errorf("store: admin bootstrap password must be at least 12 characters")
	}
	role := normalizeAdminRole(params.Role)
	if role == "" {
		return BootstrapAdminResult{}, ErrInvalidAdminRole
	}
	displayName := strings.TrimSpace(params.DisplayName)
	if displayName == "" {
		displayName = "Mautrade Super Admin"
	}
	now := normalizedNow(params.Now)

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return BootstrapAdminResult{}, fmt.Errorf("store: begin bootstrap admin: %w", err)
	}
	defer tx.Rollback(ctx)

	existing, err := adminByEmailForAuth(ctx, tx, email)
	if err == nil {
		return BootstrapAdminResult{Admin: existing.Admin, Created: false}, tx.Commit(ctx)
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return BootstrapAdminResult{}, err
	}

	adminID, err := id.New()
	if err != nil {
		return BootstrapAdminResult{}, err
	}
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return BootstrapAdminResult{}, fmt.Errorf("store: hash admin password: %w", err)
	}
	if _, err := tx.Exec(ctx, `
INSERT INTO admin_users (
  id, email, display_name, password_hash, role, otp_enabled, status, created_at, updated_at
) VALUES (
  $1::uuid, $2, $3, $4, $5, false, 'active', $6, $6
)`, adminID.String(), email, displayName, string(passwordHash), role, now); err != nil {
		return BootstrapAdminResult{}, fmt.Errorf("store: insert bootstrap admin: %w", err)
	}
	admin, err := adminByIDForAuth(ctx, tx, adminID.String())
	if err != nil {
		return BootstrapAdminResult{}, err
	}
	if err := insertAdminAudit(ctx, tx, "", "admin_bootstrap_created", "admin_user", admin.ID, map[string]any{
		"email": admin.Email,
		"role":  admin.Role,
	}, now); err != nil {
		return BootstrapAdminResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return BootstrapAdminResult{}, fmt.Errorf("store: commit bootstrap admin: %w", err)
	}
	return BootstrapAdminResult{Admin: admin, Created: true}, nil
}

func (s *DashboardStore) LoginAdmin(ctx context.Context, params LoginAdminParams) (LoginAdminResult, error) {
	if !s.Ready() {
		return LoginAdminResult{}, fmt.Errorf("store: admin login requires postgres")
	}
	now := normalizedNow(params.Now)
	email := normalizeEmail(params.Email)

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return LoginAdminResult{}, fmt.Errorf("store: begin admin login: %w", err)
	}
	defer tx.Rollback(ctx)

	row, err := adminByEmailForAuth(ctx, tx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return LoginAdminResult{}, ErrInvalidAdminCredential
		}
		return LoginAdminResult{}, err
	}
	if row.Admin.Status != "active" {
		return LoginAdminResult{}, ErrInvalidAdminCredential
	}
	if err := bcrypt.CompareHashAndPassword([]byte(row.PasswordHash), []byte(params.Password)); err != nil {
		return LoginAdminResult{}, ErrInvalidAdminCredential
	}
	if row.Admin.OTPEnabled {
		if strings.TrimSpace(params.OTPCode) == "" {
			return LoginAdminResult{Admin: row.Admin, OTPRequired: true, OTPChallenge: "totp"}, ErrAdminOTPRequired
		}
		if len(row.OTPSecretCiphertext) == 0 || params.OTPVerifier == nil || !params.OTPVerifier(row.OTPSecretCiphertext, params.OTPCode, now) {
			return LoginAdminResult{}, ErrInvalidOTP
		}
	}

	session, err := createAdminSession(ctx, tx, row.Admin.ID, params.SessionTTL, params.UserAgent, params.IPAddress, now)
	if err != nil {
		return LoginAdminResult{}, err
	}
	if _, err := tx.Exec(ctx, `
UPDATE admin_users
SET last_login_at = $2,
    updated_at = $2
WHERE id = $1::uuid`, row.Admin.ID, now); err != nil {
		return LoginAdminResult{}, fmt.Errorf("store: update admin last login: %w", err)
	}
	row.Admin.LastLogin = &now
	if err := insertAdminAudit(ctx, tx, row.Admin.ID, "admin_login", "admin_user", row.Admin.ID, map[string]any{
		"session_id": session.ID,
	}, now); err != nil {
		return LoginAdminResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return LoginAdminResult{}, fmt.Errorf("store: commit admin login: %w", err)
	}
	return LoginAdminResult{Admin: row.Admin, Session: &session}, nil
}

func (s *DashboardStore) StoreAdminOTPSecret(ctx context.Context, adminID string, secretCiphertext []byte, now time.Time) (AdminUserView, error) {
	if !s.Ready() {
		return AdminUserView{}, fmt.Errorf("store: admin otp setup requires postgres")
	}
	if strings.TrimSpace(adminID) == "" {
		return AdminUserView{}, ErrAdminNotFound
	}
	if len(secretCiphertext) == 0 {
		return AdminUserView{}, ErrAdminOTPSecretMissing
	}
	now = normalizedNow(now)

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return AdminUserView{}, fmt.Errorf("store: begin admin otp setup: %w", err)
	}
	defer tx.Rollback(ctx)

	result, err := tx.Exec(ctx, `
UPDATE admin_users
SET otp_secret_ciphertext = $2,
    otp_enabled = false,
    updated_at = $3
WHERE id = $1::uuid
  AND status = 'active'`, adminID, secretCiphertext, now)
	if err != nil {
		return AdminUserView{}, fmt.Errorf("store: store admin otp secret: %w", err)
	}
	if result.RowsAffected() == 0 {
		return AdminUserView{}, ErrAdminNotFound
	}
	admin, err := adminByIDForAuth(ctx, tx, adminID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return AdminUserView{}, ErrAdminNotFound
		}
		return AdminUserView{}, err
	}
	if err := insertAdminAudit(ctx, tx, admin.ID, "admin_2fa_setup_started", "admin_user", admin.ID, map[string]any{
		"otp_enabled": admin.OTPEnabled,
	}, now); err != nil {
		return AdminUserView{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return AdminUserView{}, fmt.Errorf("store: commit admin otp setup: %w", err)
	}
	return admin, nil
}

func (s *DashboardStore) AdminOTPSecret(ctx context.Context, adminID string) (AdminOTPSecretView, error) {
	if !s.Ready() {
		return AdminOTPSecretView{}, fmt.Errorf("store: admin otp secret requires postgres")
	}
	row, err := scanAuthAdminRow(s.db.QueryRow(ctx, adminSelectSQL()+`
WHERE id = $1::uuid
  AND status = 'active'`, adminID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return AdminOTPSecretView{}, ErrAdminNotFound
		}
		return AdminOTPSecretView{}, err
	}
	if len(row.OTPSecretCiphertext) == 0 {
		return AdminOTPSecretView{}, ErrAdminOTPSecretMissing
	}
	return AdminOTPSecretView{Admin: row.Admin, SecretCiphertext: row.OTPSecretCiphertext}, nil
}

func (s *DashboardStore) EnableAdminOTP(ctx context.Context, adminID string, now time.Time) (AdminUserView, error) {
	return s.setAdminOTPEnabled(ctx, adminID, true, now)
}

func (s *DashboardStore) DisableAdminOTP(ctx context.Context, adminID string, now time.Time) (AdminUserView, error) {
	return s.setAdminOTPEnabled(ctx, adminID, false, now)
}

func (s *DashboardStore) AuthenticateAdminSession(ctx context.Context, rawToken string) (AdminUserView, error) {
	if !s.Ready() {
		return AdminUserView{}, fmt.Errorf("store: authenticate admin session requires postgres")
	}
	tokenHash := hashSessionToken(rawToken)
	if tokenHash == "" {
		return AdminUserView{}, ErrInvalidSession
	}

	var sessionID string
	var adminID string
	if err := s.db.QueryRow(ctx, `
SELECT s.id::text, a.id::text
FROM admin_auth_sessions s
JOIN admin_users a ON a.id = s.admin_id
WHERE s.token_hash = $1
  AND s.revoked_at IS NULL
  AND s.expires_at > now()
  AND a.status = 'active'`, tokenHash).Scan(&sessionID, &adminID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return AdminUserView{}, ErrInvalidSession
		}
		return AdminUserView{}, fmt.Errorf("store: authenticate admin session: %w", err)
	}
	_, _ = s.db.Exec(ctx, `
UPDATE admin_auth_sessions
SET last_seen_at = now()
WHERE id = $1::uuid`, sessionID)
	return s.AdminUser(ctx, adminID)
}

func (s *DashboardStore) LogoutAdminSession(ctx context.Context, rawToken string) error {
	if !s.Ready() {
		return fmt.Errorf("store: admin logout requires postgres")
	}
	tokenHash := hashSessionToken(rawToken)
	if tokenHash == "" {
		return ErrInvalidSession
	}
	if _, err := s.db.Exec(ctx, `
UPDATE admin_auth_sessions
SET revoked_at = COALESCE(revoked_at, now())
WHERE token_hash = $1`, tokenHash); err != nil {
		return fmt.Errorf("store: revoke admin session: %w", err)
	}
	return nil
}

func (s *DashboardStore) AdminUser(ctx context.Context, adminID string) (AdminUserView, error) {
	if !s.Ready() {
		return AdminUserView{}, fmt.Errorf("store: admin user requires postgres")
	}
	return adminByIDForAuth(ctx, s.db, adminID)
}

type authAdminRow struct {
	Admin               AdminUserView
	PasswordHash        string
	OTPSecretCiphertext []byte
}

func (s *DashboardStore) setAdminOTPEnabled(ctx context.Context, adminID string, enabled bool, now time.Time) (AdminUserView, error) {
	if !s.Ready() {
		return AdminUserView{}, fmt.Errorf("store: admin otp state requires postgres")
	}
	now = normalizedNow(now)

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return AdminUserView{}, fmt.Errorf("store: begin admin otp state: %w", err)
	}
	defer tx.Rollback(ctx)

	var result pgconn.CommandTag
	if enabled {
		result, err = tx.Exec(ctx, `
UPDATE admin_users
SET otp_enabled = true,
    updated_at = $2
WHERE id = $1::uuid
  AND status = 'active'
  AND otp_secret_ciphertext IS NOT NULL`, adminID, now)
	} else {
		result, err = tx.Exec(ctx, `
UPDATE admin_users
SET otp_enabled = false,
    otp_secret_ciphertext = NULL,
    updated_at = $2
WHERE id = $1::uuid
  AND status = 'active'`, adminID, now)
	}
	if err != nil {
		return AdminUserView{}, fmt.Errorf("store: update admin otp state: %w", err)
	}
	if result.RowsAffected() == 0 {
		if _, lookupErr := adminByIDForAuth(ctx, tx, adminID); errors.Is(lookupErr, pgx.ErrNoRows) {
			return AdminUserView{}, ErrAdminNotFound
		}
		if enabled {
			return AdminUserView{}, ErrAdminOTPSecretMissing
		}
		return AdminUserView{}, ErrAdminNotFound
	}

	admin, err := adminByIDForAuth(ctx, tx, adminID)
	if err != nil {
		return AdminUserView{}, err
	}
	action := "admin_2fa_disabled"
	if enabled {
		action = "admin_2fa_enabled"
	}
	if err := insertAdminAudit(ctx, tx, admin.ID, action, "admin_user", admin.ID, map[string]any{
		"otp_enabled": admin.OTPEnabled,
	}, now); err != nil {
		return AdminUserView{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return AdminUserView{}, fmt.Errorf("store: commit admin otp state: %w", err)
	}
	return admin, nil
}

func createAdminSession(ctx context.Context, tx pgx.Tx, adminID string, ttl time.Duration, userAgent, ipAddress string, now time.Time) (SessionView, error) {
	if ttl <= 0 {
		ttl = 12 * time.Hour
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
INSERT INTO admin_auth_sessions (
  id, admin_id, token_hash, user_agent, ip_address, expires_at, created_at, last_seen_at
) VALUES (
  $1::uuid, $2::uuid, $3, $4, $5, $6, $7, $7
)`, sessionID.String(), adminID, tokenHash, userAgent, ipAddress, expiresAt, now); err != nil {
		return SessionView{}, fmt.Errorf("store: insert admin session: %w", err)
	}
	return SessionView{
		ID:        sessionID.String(),
		Token:     rawToken,
		ExpiresAt: expiresAt,
		CreatedAt: now,
	}, nil
}

func adminByEmailForAuth(ctx context.Context, queryer authQueryer, email string) (authAdminRow, error) {
	return scanAuthAdminRow(queryer.QueryRow(ctx, adminSelectSQL()+`
WHERE lower(email) = $1`, email))
}

func adminByIDForAuth(ctx context.Context, queryer authQueryer, adminID string) (AdminUserView, error) {
	row, err := scanAuthAdminRow(queryer.QueryRow(ctx, adminSelectSQL()+`
WHERE id = $1::uuid`, adminID))
	if err != nil {
		return AdminUserView{}, err
	}
	return row.Admin, nil
}

func adminSelectSQL() string {
	return `
SELECT id::text, email, display_name, role, otp_enabled, status, last_login_at, created_at, password_hash, otp_secret_ciphertext
FROM admin_users
`
}

func scanAuthAdminRow(row pgx.Row) (authAdminRow, error) {
	var result authAdminRow
	if err := row.Scan(
		&result.Admin.ID,
		&result.Admin.Email,
		&result.Admin.DisplayName,
		&result.Admin.Role,
		&result.Admin.OTPEnabled,
		&result.Admin.Status,
		&result.Admin.LastLogin,
		&result.Admin.CreatedAt,
		&result.PasswordHash,
		&result.OTPSecretCiphertext,
	); err != nil {
		return authAdminRow{}, err
	}
	return result, nil
}

func insertAdminAudit(ctx context.Context, tx pgx.Tx, adminID, action, entity, entityID string, after map[string]any, now time.Time) error {
	auditID, err := id.New()
	if err != nil {
		return err
	}
	afterJSON, err := json.Marshal(after)
	if err != nil {
		return fmt.Errorf("store: marshal admin audit: %w", err)
	}
	var actorID any
	if strings.TrimSpace(adminID) != "" {
		actorID = strings.TrimSpace(adminID)
	}
	if _, err := tx.Exec(ctx, `
INSERT INTO audit_logs (
  id, actor_type, actor_id, action, entity, entity_id, after_state, created_at
) VALUES (
  $1::uuid, 'admin', $2::uuid, $3, $4, $5::uuid, $6::jsonb, $7
)`, auditID.String(), actorID, action, entity, entityID, string(afterJSON), now); err != nil {
		return fmt.Errorf("store: insert admin audit: %w", err)
	}
	return nil
}

func normalizeAdminRole(value string) string {
	role := strings.ToLower(strings.TrimSpace(value))
	if role == "" {
		role = "super_admin"
	}
	switch role {
	case "super_admin", "admin", "ops", "auditor":
		return role
	default:
		return ""
	}
}
