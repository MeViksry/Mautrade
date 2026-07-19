package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

var (
	ErrUnsupportedExchange     = errors.New("store: unsupported exchange")
	ErrExchangeBindingNotFound = errors.New("store: exchange binding not found")
	ErrInvalidExchangeStatus   = errors.New("store: invalid exchange binding status")
)

type UpsertExchangeBindingParams struct {
	UserID                  string
	ExchangeName            string
	APIKeyCiphertext        []byte
	APISecretCiphertext     []byte
	APIPassphraseCiphertext []byte
	PermissionScope         string
	Now                     time.Time
}

type ExchangeBindingCredentialCiphertext struct {
	ID                      string     `json:"id"`
	ExchangeName            string     `json:"exchange"`
	Status                  string     `json:"status"`
	APIKeyCiphertext        []byte     `json:"-"`
	APISecretCiphertext     []byte     `json:"-"`
	APIPassphraseCiphertext []byte     `json:"-"`
	PermissionScope         string     `json:"permissionScope"`
	LastVerifiedAt          *time.Time `json:"lastVerifiedAt,omitempty"`
	CreatedAt               time.Time  `json:"createdAt"`
	UpdatedAt               time.Time  `json:"updatedAt"`
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
		if err := rows.Scan(&binding.ID, &binding.Name, &binding.Status, &binding.LastSynced, &binding.Balance, &binding.HasAPI); err != nil {
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
  api_passphrase_ciphertext, permission_scope, status, last_verified_at,
  revoked_at, created_at, updated_at
) VALUES (
  $1::uuid, $2::uuid, $3, $4, $5, $6, $7, 'active', $8, NULL, $8, $8
)
ON CONFLICT (user_id, exchange_name) DO UPDATE SET
  api_key_ciphertext = EXCLUDED.api_key_ciphertext,
  api_secret_ciphertext = EXCLUDED.api_secret_ciphertext,
  api_passphrase_ciphertext = EXCLUDED.api_passphrase_ciphertext,
  permission_scope = EXCLUDED.permission_scope,
  status = 'active',
  last_verified_at = EXCLUDED.last_verified_at,
  revoked_at = NULL,
  updated_at = EXCLUDED.updated_at
RETURNING id::text, exchange_name, status, api_key_ciphertext, api_secret_ciphertext,
          api_passphrase_ciphertext, permission_scope, last_verified_at, created_at, updated_at`,
		bindingID,
		params.UserID,
		exchangeName,
		params.APIKeyCiphertext,
		params.APISecretCiphertext,
		nullableBytes(params.APIPassphraseCiphertext),
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
SELECT id::text, exchange_name, status, api_key_ciphertext, api_secret_ciphertext,
       api_passphrase_ciphertext, permission_scope, last_verified_at, created_at, updated_at
FROM exchange_bindings
WHERE user_id = $1::uuid
  AND exchange_name = $2`, userID, normalizedExchange))
	if errors.Is(err, pgx.ErrNoRows) {
		return ExchangeBindingCredentialCiphertext{}, ErrExchangeBindingNotFound
	}
	if err != nil {
		return ExchangeBindingCredentialCiphertext{}, fmt.Errorf("store: exchange binding credential: %w", err)
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
SET status = $3,
    revoked_at = CASE WHEN $3 = 'revoked' THEN $4 ELSE NULL END,
    updated_at = $4
WHERE user_id = $1::uuid
  AND exchange_name = $2
RETURNING id::text, exchange_name, status, api_key_ciphertext, api_secret_ciphertext,
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
	case "revoked", "disconnected", "disconnect", "deleted", "delete":
		return "revoked", nil
	default:
		return "", ErrInvalidExchangeStatus
	}
}

func scanExchangeBindingCredential(row pgx.Row) (ExchangeBindingCredentialCiphertext, error) {
	var binding ExchangeBindingCredentialCiphertext
	if err := row.Scan(
		&binding.ID,
		&binding.ExchangeName,
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
