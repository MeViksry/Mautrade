package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/MeViksry/qdecimal"
	"github.com/jackc/pgx/v5"
)

var (
	ErrUnsupportedExchange        = errors.New("store: unsupported exchange")
	ErrExchangeBindingNotFound    = errors.New("store: exchange binding not found")
	ErrInvalidExchangeStatus      = errors.New("store: invalid exchange binding status")
	ErrInvalidExchangeAccountMode = errors.New("store: invalid exchange account mode")
)

type UpsertExchangeBindingParams struct {
	UserID                  string
	ExchangeName            string
	APIKeyCiphertext        []byte
	APISecretCiphertext     []byte
	APIPassphraseCiphertext []byte
	AccountMode             string
	PermissionScope         string
	Now                     time.Time
}

type ExchangeBindingCredentialCiphertext struct {
	ID                      string     `json:"id"`
	ExchangeName            string     `json:"exchange"`
	AccountMode             string     `json:"accountMode"`
	Status                  string     `json:"status"`
	APIKeyCiphertext        []byte     `json:"-"`
	APISecretCiphertext     []byte     `json:"-"`
	APIPassphraseCiphertext []byte     `json:"-"`
	PermissionScope         string     `json:"permissionScope"`
	LastVerifiedAt          *time.Time `json:"lastVerifiedAt,omitempty"`
	CreatedAt               time.Time  `json:"createdAt"`
	UpdatedAt               time.Time  `json:"updatedAt"`
}

type ExchangeBindingCredentialSyncTarget struct {
	UserID  string
	Binding ExchangeBindingCredentialCiphertext
}

func (s *DashboardStore) UserExchangeBindings(ctx context.Context, userID, defaultCurrency string) ([]ExchangeBindingView, error) {
	const query = `
WITH latest_balances AS (
  SELECT DISTINCT ON (exchange_binding_id, asset)
    exchange_binding_id,
    asset,
    free_amount + locked_amount AS amount
  FROM exchange_balance_snapshots
  WHERE asset = $1
  ORDER BY exchange_binding_id, asset, captured_at DESC
)
SELECT
  b.id::text,
  CASE b.exchange_name
    WHEN 'binance' THEN 'Binance'
    WHEN 'okx' THEN 'OKX'
    WHEN 'bybit' THEN 'Bybit'
    WHEN 'tokocrypto' THEN 'Tokocrypto'
    ELSE b.exchange_name
  END AS name,
  b.account_mode,
  CASE WHEN b.status = 'active' THEN 'connected' ELSE 'disconnected' END AS status,
  b.last_verified_at,
  COALESCE(lb.amount, 0)::text AS balance,
  b.status <> 'revoked' AS has_api
FROM exchange_bindings b
LEFT JOIN latest_balances lb ON lb.exchange_binding_id = b.id
WHERE b.user_id = $2::uuid
ORDER BY b.created_at ASC`

	rows, err := s.db.Query(ctx, query, defaultCurrency, userID)
	if err != nil {
		return nil, fmt.Errorf("store: user exchange bindings: %w", err)
	}
	defer rows.Close()

	var bindings []ExchangeBindingView
	for rows.Next() {
		var binding ExchangeBindingView
		if err := rows.Scan(&binding.ID, &binding.Name, &binding.AccountMode, &binding.Status, &binding.LastSynced, &binding.Balance, &binding.HasAPI); err != nil {
			return nil, fmt.Errorf("store: scan user exchange binding: %w", err)
		}
		bindings = append(bindings, binding)
	}
	if bindings == nil {
		bindings = []ExchangeBindingView{}
	}
	return bindings, rows.Err()
}

func (s *DashboardStore) UpsertExchangeBinding(ctx context.Context, params UpsertExchangeBindingParams) (ExchangeBindingCredentialCiphertext, error) {
	if !s.Ready() {
		return ExchangeBindingCredentialCiphertext{}, fmt.Errorf("store: exchange binding requires postgres")
	}
	exchangeName, err := normalizeSupportedExchange(params.ExchangeName)
	if err != nil {
		return ExchangeBindingCredentialCiphertext{}, err
	}
	accountMode, err := normalizeExchangeAccountMode(params.AccountMode)
	if err != nil {
		return ExchangeBindingCredentialCiphertext{}, err
	}
	if len(params.APIKeyCiphertext) == 0 || len(params.APISecretCiphertext) == 0 {
		return ExchangeBindingCredentialCiphertext{}, fmt.Errorf("store: encrypted api key and secret are required")
	}
	permissionScope := strings.TrimSpace(params.PermissionScope)
	if permissionScope == "" {
		permissionScope = "trade_only"
	}
	now := normalizedNow(params.Now)
	bindingID, err := newUUIDText()
	if err != nil {
		return ExchangeBindingCredentialCiphertext{}, err
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return ExchangeBindingCredentialCiphertext{}, fmt.Errorf("store: begin exchange binding: %w", err)
	}
	defer tx.Rollback(ctx)

	row := tx.QueryRow(ctx, `
INSERT INTO exchange_bindings (
  id, user_id, exchange_name, api_key_ciphertext, api_secret_ciphertext,
  api_passphrase_ciphertext, account_mode, permission_scope, status, last_verified_at,
  revoked_at, created_at, updated_at
) VALUES (
  $1::uuid, $2::uuid, $3, $4, $5, $6, $7, $8, 'active', $9, NULL, $9, $9
)
ON CONFLICT (user_id, exchange_name) DO UPDATE SET
  api_key_ciphertext = EXCLUDED.api_key_ciphertext,
  api_secret_ciphertext = EXCLUDED.api_secret_ciphertext,
  api_passphrase_ciphertext = EXCLUDED.api_passphrase_ciphertext,
  account_mode = EXCLUDED.account_mode,
  permission_scope = EXCLUDED.permission_scope,
  status = 'active',
  last_verified_at = EXCLUDED.last_verified_at,
  revoked_at = NULL,
  updated_at = EXCLUDED.updated_at
RETURNING id::text, exchange_name, account_mode, status, api_key_ciphertext, api_secret_ciphertext,
          api_passphrase_ciphertext, permission_scope, last_verified_at, created_at, updated_at`,
		bindingID,
		params.UserID,
		exchangeName,
		params.APIKeyCiphertext,
		params.APISecretCiphertext,
		nullableBytes(params.APIPassphraseCiphertext),
		accountMode,
		permissionScope,
		now,
	)
	binding, err := scanExchangeBindingCredential(row)
	if err != nil {
		return ExchangeBindingCredentialCiphertext{}, fmt.Errorf("store: upsert exchange binding: %w", err)
	}
	if err := insertExchangeBindingAudit(ctx, tx, params.UserID, "exchange_binding_upserted", binding); err != nil {
		return ExchangeBindingCredentialCiphertext{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return ExchangeBindingCredentialCiphertext{}, fmt.Errorf("store: commit exchange binding: %w", err)
	}
	return binding, nil
}

func (s *DashboardStore) ExchangeBindingCredential(ctx context.Context, userID, exchangeName string) (ExchangeBindingCredentialCiphertext, error) {
	if !s.Ready() {
		return ExchangeBindingCredentialCiphertext{}, fmt.Errorf("store: exchange binding requires postgres")
	}
	normalizedExchange, err := normalizeSupportedExchange(exchangeName)
	if err != nil {
		return ExchangeBindingCredentialCiphertext{}, err
	}
	binding, err := scanExchangeBindingCredential(s.db.QueryRow(ctx, `
SELECT id::text, exchange_name, account_mode, status, api_key_ciphertext, api_secret_ciphertext,
       api_passphrase_ciphertext, permission_scope, last_verified_at, created_at, updated_at
FROM exchange_bindings
WHERE user_id = $1::uuid
  AND exchange_name = $2
  AND status <> 'revoked'`, userID, normalizedExchange))
	if errors.Is(err, pgx.ErrNoRows) {
		return ExchangeBindingCredentialCiphertext{}, ErrExchangeBindingNotFound
	}
	if err != nil {
		return ExchangeBindingCredentialCiphertext{}, fmt.Errorf("store: exchange binding credential: %w", err)
	}
	return binding, nil
}

func (s *DashboardStore) ExchangeBindingCredentialByID(ctx context.Context, bindingID string) (ExchangeBindingCredentialCiphertext, error) {
	if !s.Ready() {
		return ExchangeBindingCredentialCiphertext{}, fmt.Errorf("store: exchange binding requires postgres")
	}
	bindingID = strings.TrimSpace(bindingID)
	if bindingID == "" {
		return ExchangeBindingCredentialCiphertext{}, ErrExchangeBindingNotFound
	}
	binding, err := scanExchangeBindingCredential(s.db.QueryRow(ctx, `
SELECT id::text, exchange_name, account_mode, status, api_key_ciphertext, api_secret_ciphertext,
       api_passphrase_ciphertext, permission_scope, last_verified_at, created_at, updated_at
FROM exchange_bindings
WHERE id = $1::uuid
  AND status IN ('active', 'closing_only')`, bindingID))
	if errors.Is(err, pgx.ErrNoRows) {
		return ExchangeBindingCredentialCiphertext{}, ErrExchangeBindingNotFound
	}
	if err != nil {
		return ExchangeBindingCredentialCiphertext{}, fmt.Errorf("store: exchange binding credential by id: %w", err)
	}
	return binding, nil
}

func (s *DashboardStore) UpdateExchangeBindingStatus(ctx context.Context, userID, exchangeName, status string, now time.Time) (ExchangeBindingCredentialCiphertext, error) {
	if !s.Ready() {
		return ExchangeBindingCredentialCiphertext{}, fmt.Errorf("store: exchange binding requires postgres")
	}
	normalizedExchange, err := normalizeSupportedExchange(exchangeName)
	if err != nil {
		return ExchangeBindingCredentialCiphertext{}, err
	}
	normalizedStatus, err := normalizeBindingStatus(status)
	if err != nil {
		return ExchangeBindingCredentialCiphertext{}, err
	}
	now = normalizedNow(now)

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return ExchangeBindingCredentialCiphertext{}, fmt.Errorf("store: begin exchange binding status: %w", err)
	}
	defer tx.Rollback(ctx)

	binding, err := scanExchangeBindingCredential(tx.QueryRow(ctx, `
UPDATE exchange_bindings
SET status = CASE
      WHEN $3 = 'revoked' AND EXISTS (
        SELECT 1
        FROM layers l
        WHERE l.exchange_binding_id = exchange_bindings.id
          AND l.status IN ('open', 'partial')
      ) THEN 'closing_only'
      ELSE $3
    END,
    revoked_at = CASE WHEN $3 = 'revoked' THEN $4 ELSE NULL END,
    updated_at = $4
WHERE user_id = $1::uuid
  AND exchange_name = $2
RETURNING id::text, exchange_name, account_mode, status, api_key_ciphertext, api_secret_ciphertext,
          api_passphrase_ciphertext, permission_scope, last_verified_at, created_at, updated_at`,
		userID,
		normalizedExchange,
		normalizedStatus,
		now,
	))
	if errors.Is(err, pgx.ErrNoRows) {
		return ExchangeBindingCredentialCiphertext{}, ErrExchangeBindingNotFound
	}
	if err != nil {
		return ExchangeBindingCredentialCiphertext{}, fmt.Errorf("store: update exchange binding status: %w", err)
	}
	if err := insertExchangeBindingAudit(ctx, tx, userID, "exchange_binding_status_updated", binding); err != nil {
		return ExchangeBindingCredentialCiphertext{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return ExchangeBindingCredentialCiphertext{}, fmt.Errorf("store: commit exchange binding status: %w", err)
	}
	return binding, nil
}

func (s *DashboardStore) DeleteExchangeBinding(ctx context.Context, userID, exchangeName string, now time.Time) (ExchangeBindingCredentialCiphertext, error) {
	if !s.Ready() {
		return ExchangeBindingCredentialCiphertext{}, fmt.Errorf("store: exchange binding requires postgres")
	}
	normalizedExchange, err := normalizeSupportedExchange(exchangeName)
	if err != nil {
		return ExchangeBindingCredentialCiphertext{}, err
	}
	now = normalizedNow(now)

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return ExchangeBindingCredentialCiphertext{}, fmt.Errorf("store: begin exchange binding delete: %w", err)
	}
	defer tx.Rollback(ctx)

	binding, err := scanExchangeBindingCredential(tx.QueryRow(ctx, `
UPDATE exchange_bindings
SET status = CASE
      WHEN EXISTS (
        SELECT 1
        FROM layers l
        WHERE l.exchange_binding_id = exchange_bindings.id
          AND l.status IN ('open', 'partial')
      ) THEN 'closing_only'
      ELSE 'revoked'
    END,
    api_key_ciphertext = CASE
      WHEN EXISTS (
        SELECT 1
        FROM layers l
        WHERE l.exchange_binding_id = exchange_bindings.id
          AND l.status IN ('open', 'partial')
      ) THEN api_key_ciphertext
      ELSE ''::bytea
    END,
    api_secret_ciphertext = CASE
      WHEN EXISTS (
        SELECT 1
        FROM layers l
        WHERE l.exchange_binding_id = exchange_bindings.id
          AND l.status IN ('open', 'partial')
      ) THEN api_secret_ciphertext
      ELSE ''::bytea
    END,
    api_passphrase_ciphertext = CASE
      WHEN EXISTS (
        SELECT 1
        FROM layers l
        WHERE l.exchange_binding_id = exchange_bindings.id
          AND l.status IN ('open', 'partial')
      ) THEN api_passphrase_ciphertext
      ELSE NULL
    END,
    last_verified_at = CASE
      WHEN EXISTS (
        SELECT 1
        FROM layers l
        WHERE l.exchange_binding_id = exchange_bindings.id
          AND l.status IN ('open', 'partial')
      ) THEN last_verified_at
      ELSE NULL
    END,
    revoked_at = $3,
    updated_at = $3
WHERE user_id = $1::uuid
  AND exchange_name = $2
RETURNING id::text, exchange_name, account_mode, status, api_key_ciphertext, api_secret_ciphertext,
          api_passphrase_ciphertext, permission_scope, last_verified_at, created_at, updated_at`,
		userID,
		normalizedExchange,
		now,
	))
	if errors.Is(err, pgx.ErrNoRows) {
		return ExchangeBindingCredentialCiphertext{}, ErrExchangeBindingNotFound
	}
	if err != nil {
		return ExchangeBindingCredentialCiphertext{}, fmt.Errorf("store: delete exchange binding: %w", err)
	}
	if err := insertExchangeBindingAudit(ctx, tx, userID, "exchange_binding_deleted", binding); err != nil {
		return ExchangeBindingCredentialCiphertext{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return ExchangeBindingCredentialCiphertext{}, fmt.Errorf("store: commit exchange binding delete: %w", err)
	}
	return binding, nil
}

func (s *DashboardStore) DueActiveExchangeBindingCredentials(ctx context.Context, userID, asset string, staleBefore time.Time) ([]ExchangeBindingCredentialCiphertext, error) {
	if !s.Ready() {
		return nil, fmt.Errorf("store: exchange binding requires postgres")
	}
	asset = strings.ToUpper(strings.TrimSpace(asset))
	if asset == "" {
		asset = "USDT"
	}
	rows, err := s.db.Query(ctx, `
SELECT id::text, exchange_name, account_mode, status, api_key_ciphertext, api_secret_ciphertext,
       api_passphrase_ciphertext, permission_scope, last_verified_at, created_at, updated_at
FROM exchange_bindings
WHERE user_id = $1::uuid
  AND status = 'active'
  AND NOT EXISTS (
    SELECT 1
    FROM exchange_balance_snapshots s
    WHERE s.exchange_binding_id = exchange_bindings.id
      AND s.asset = $2
      AND s.captured_at >= $3
  )
ORDER BY updated_at ASC`, userID, asset, staleBefore)
	if err != nil {
		return nil, fmt.Errorf("store: due exchange binding credentials: %w", err)
	}
	defer rows.Close()

	var bindings []ExchangeBindingCredentialCiphertext
	for rows.Next() {
		binding, err := scanExchangeBindingCredential(rows)
		if err != nil {
			return nil, fmt.Errorf("store: scan due exchange binding credential: %w", err)
		}
		bindings = append(bindings, binding)
	}
	if bindings == nil {
		bindings = []ExchangeBindingCredentialCiphertext{}
	}
	return bindings, rows.Err()
}

func (s *DashboardStore) DueAllActiveExchangeBindingCredentials(ctx context.Context, asset string, staleBefore time.Time, includeClosingOnly bool) ([]ExchangeBindingCredentialSyncTarget, error) {
	if !s.Ready() {
		return nil, fmt.Errorf("store: exchange binding requires postgres")
	}
	asset = strings.ToUpper(strings.TrimSpace(asset))
	if asset == "" {
		asset = "USDT"
	}
	statusFilter := "b.status = 'active'"
	if includeClosingOnly {
		statusFilter = "b.status IN ('active', 'closing_only')"
	}
	rows, err := s.db.Query(ctx, fmt.Sprintf(`
SELECT b.user_id::text, b.id::text, b.exchange_name, b.account_mode, b.status, b.api_key_ciphertext, b.api_secret_ciphertext,
       b.api_passphrase_ciphertext, b.permission_scope, b.last_verified_at, b.created_at, b.updated_at
FROM exchange_bindings b
JOIN users u ON u.id = b.user_id
WHERE %s
  AND u.status = 'active'
  AND NOT EXISTS (
    SELECT 1
    FROM exchange_balance_snapshots s
    WHERE s.exchange_binding_id = b.id
      AND s.asset = $1
      AND s.captured_at >= $2
  )
ORDER BY b.updated_at ASC`, statusFilter), asset, staleBefore)
	if err != nil {
		return nil, fmt.Errorf("store: due all exchange binding credentials: %w", err)
	}
	defer rows.Close()

	var targets []ExchangeBindingCredentialSyncTarget
	for rows.Next() {
		var target ExchangeBindingCredentialSyncTarget
		if err := rows.Scan(
			&target.UserID,
			&target.Binding.ID,
			&target.Binding.ExchangeName,
			&target.Binding.AccountMode,
			&target.Binding.Status,
			&target.Binding.APIKeyCiphertext,
			&target.Binding.APISecretCiphertext,
			&target.Binding.APIPassphraseCiphertext,
			&target.Binding.PermissionScope,
			&target.Binding.LastVerifiedAt,
			&target.Binding.CreatedAt,
			&target.Binding.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("store: scan due all exchange binding credential: %w", err)
		}
		targets = append(targets, target)
	}
	if targets == nil {
		targets = []ExchangeBindingCredentialSyncTarget{}
	}
	return targets, rows.Err()
}

type ExchangeBalanceSnapshotParams struct {
	UserID            string
	ExchangeBindingID string
	AccountMode       string
	Asset             string
	FreeAmount        string
	LockedAmount      string
	CapturedAt        time.Time
}

func (s *DashboardStore) RecordExchangeBalanceSnapshot(ctx context.Context, params ExchangeBalanceSnapshotParams) error {
	if !s.Ready() {
		return fmt.Errorf("store: exchange balance snapshot requires postgres")
	}
	asset := strings.ToUpper(strings.TrimSpace(params.Asset))
	if asset == "" {
		asset = "USDT"
	}
	accountMode := ""
	if strings.TrimSpace(params.AccountMode) != "" {
		var err error
		accountMode, err = normalizeExchangeAccountMode(params.AccountMode)
		if err != nil {
			return err
		}
	}
	freeAmount, err := normalizeSnapshotAmount(params.FreeAmount)
	if err != nil {
		return fmt.Errorf("store: invalid free balance amount: %w", err)
	}
	lockedAmount, err := normalizeSnapshotAmount(params.LockedAmount)
	if err != nil {
		return fmt.Errorf("store: invalid locked balance amount: %w", err)
	}
	now := normalizedNow(params.CapturedAt)
	snapshotID, err := newUUIDText()
	if err != nil {
		return err
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("store: begin exchange balance snapshot: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `
INSERT INTO exchange_balance_snapshots (
  id, user_id, exchange_binding_id, asset, free_amount, locked_amount, captured_at
) VALUES (
  $1::uuid, $2::uuid, $3::uuid, $4, $5, $6, $7
)`,
		snapshotID,
		params.UserID,
		params.ExchangeBindingID,
		asset,
		freeAmount,
		lockedAmount,
		now,
	); err != nil {
		return fmt.Errorf("store: insert exchange balance snapshot: %w", err)
	}

	updateQuery := `
UPDATE exchange_bindings
SET status = 'active',
    last_verified_at = $3,
    updated_at = $3
WHERE user_id = $1::uuid
  AND id = $2::uuid`
	updateArgs := []any{params.UserID, params.ExchangeBindingID, now}
	if accountMode != "" {
		updateQuery = `
UPDATE exchange_bindings
SET status = 'active',
    account_mode = $4,
    last_verified_at = $3,
    updated_at = $3
WHERE user_id = $1::uuid
  AND id = $2::uuid`
		updateArgs = append(updateArgs, accountMode)
	}
	tag, err := tx.Exec(ctx, updateQuery, updateArgs...)
	if err != nil {
		return fmt.Errorf("store: mark exchange binding verified: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrExchangeBindingNotFound
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("store: commit exchange balance snapshot: %w", err)
	}
	return nil
}

func normalizeSupportedExchange(exchangeName string) (string, error) {
	exchangeName = strings.ToLower(strings.TrimSpace(exchangeName))
	switch exchangeName {
	case "binance", "okx", "bybit", "tokocrypto":
		return exchangeName, nil
	default:
		return "", ErrUnsupportedExchange
	}
}

func normalizeBindingStatus(status string) (string, error) {
	status = strings.ToLower(strings.TrimSpace(status))
	switch status {
	case "active", "connected", "connect":
		return "active", nil
	case "invalid":
		return "invalid", nil
	case "closing_only", "closing-only", "sell_only", "sell-only":
		return "closing_only", nil
	case "revoked", "disconnected", "disconnect", "deleted", "delete":
		return "revoked", nil
	default:
		return "", ErrInvalidExchangeStatus
	}
}

func normalizeExchangeAccountMode(accountMode string) (string, error) {
	accountMode = strings.ToLower(strings.TrimSpace(accountMode))
	switch accountMode {
	case "", "real", "live", "production", "prod":
		return "real", nil
	case "demo", "paper", "simulated", "simulation":
		return "demo", nil
	case "testnet", "sandbox":
		return "testnet", nil
	default:
		return "", ErrInvalidExchangeAccountMode
	}
}

func normalizeSnapshotAmount(amount string) (string, error) {
	amount = strings.TrimSpace(amount)
	if amount == "" {
		amount = "0"
	}
	parsed, err := qdecimal.Parse(amount)
	if err != nil {
		return "", err
	}
	if parsed.Sign() < 0 {
		return "", fmt.Errorf("amount must be non-negative")
	}
	return parsed.String(), nil
}

func scanExchangeBindingCredential(row pgx.Row) (ExchangeBindingCredentialCiphertext, error) {
	var binding ExchangeBindingCredentialCiphertext
	if err := row.Scan(
		&binding.ID,
		&binding.ExchangeName,
		&binding.AccountMode,
		&binding.Status,
		&binding.APIKeyCiphertext,
		&binding.APISecretCiphertext,
		&binding.APIPassphraseCiphertext,
		&binding.PermissionScope,
		&binding.LastVerifiedAt,
		&binding.CreatedAt,
		&binding.UpdatedAt,
	); err != nil {
		return ExchangeBindingCredentialCiphertext{}, err
	}
	return binding, nil
}

func nullableBytes(value []byte) any {
	if len(value) == 0 {
		return nil
	}
	return value
}

func insertExchangeBindingAudit(ctx context.Context, tx pgx.Tx, userID, action string, binding ExchangeBindingCredentialCiphertext) error {
	auditID, err := newUUIDText()
	if err != nil {
		return err
	}
	afterJSON, err := json.Marshal(map[string]any{
		"exchange":         binding.ExchangeName,
		"account_mode":     binding.AccountMode,
		"status":           binding.Status,
		"permission_scope": binding.PermissionScope,
		"has_api_key":      len(binding.APIKeyCiphertext) > 0,
		"has_api_secret":   len(binding.APISecretCiphertext) > 0,
		"has_passphrase":   len(binding.APIPassphraseCiphertext) > 0,
	})
	if err != nil {
		return fmt.Errorf("store: marshal exchange binding audit: %w", err)
	}
	if _, err := tx.Exec(ctx, `
INSERT INTO audit_logs (
  id, actor_type, actor_id, action, entity, entity_id, after_state, created_at
) VALUES (
  $1::uuid, 'user', $2::uuid, $3, 'exchange_binding', $4::uuid, $5::jsonb, now()
)`, auditID, userID, action, binding.ID, string(afterJSON)); err != nil {
		return fmt.Errorf("store: insert exchange binding audit: %w", err)
	}
	return nil
}
