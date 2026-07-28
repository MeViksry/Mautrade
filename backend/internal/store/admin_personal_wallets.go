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
)

var (
	ErrInvalidPersonalWalletCode    = errors.New("store: invalid personal wallet code")
	ErrInvalidPersonalWalletAddress = errors.New("store: invalid personal wallet address")
)

type AdminPersonalWalletView struct {
	Code              string     `json:"code"`
	DisplayName       string     `json:"displayName"`
	WalletAddress     string     `json:"walletAddress"`
	ShareRate         string     `json:"shareRate"`
	Balance           string     `json:"balance"`
	CommissionBalance string     `json:"commissionBalance"`
	UpdatedBy         *string    `json:"updatedBy,omitempty"`
	CreatedAt         time.Time  `json:"createdAt"`
	UpdatedAt         *time.Time `json:"updatedAt,omitempty"`
}

type UpdateAdminPersonalWalletParams struct {
	Code          string
	WalletAddress string
	AdminID       string
	Now           time.Time
}

func (s *DashboardStore) AdminPersonalWallets(ctx context.Context) ([]AdminPersonalWalletView, error) {
	if !s.Ready() {
		return nil, fmt.Errorf("store: personal wallets require postgres")
	}

	rows, err := s.db.Query(ctx, `
WITH commission AS (
  SELECT wallet_code, COALESCE(SUM(amount), 0)::text AS balance
  FROM admin_wallet_commission_ledger
  GROUP BY wallet_code
)
SELECT
  w.code,
  w.display_name,
  w.wallet_address,
  CASE w.code WHEN 'viksry' THEN '0.10' WHEN 'aryanto_hong' THEN '0.90' ELSE '0' END AS share_rate,
  COALESCE(c.balance, '0') AS balance,
  w.updated_by::text,
  w.created_at,
  w.updated_at
FROM admin_personal_wallets w
LEFT JOIN commission c ON c.wallet_code = w.code
ORDER BY CASE w.code WHEN 'viksry' THEN 1 WHEN 'aryanto_hong' THEN 2 ELSE 99 END, w.code
`)
	if err != nil {
		return nil, fmt.Errorf("store: list admin personal wallets: %w", err)
	}
	defer rows.Close()

	var wallets []AdminPersonalWalletView
	for rows.Next() {
		var wallet AdminPersonalWalletView
		if err := rows.Scan(
			&wallet.Code,
			&wallet.DisplayName,
			&wallet.WalletAddress,
			&wallet.ShareRate,
			&wallet.Balance,
			&wallet.UpdatedBy,
			&wallet.CreatedAt,
			&wallet.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("store: scan admin personal wallet: %w", err)
		}
		wallet.CommissionBalance = wallet.Balance
		wallets = append(wallets, wallet)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if wallets == nil {
		wallets = []AdminPersonalWalletView{}
	}
	return wallets, nil
}

func (s *DashboardStore) AdminUpdatePersonalWalletAddress(ctx context.Context, params UpdateAdminPersonalWalletParams) (AdminPersonalWalletView, error) {
	if !s.Ready() {
		return AdminPersonalWalletView{}, fmt.Errorf("store: update personal wallet requires postgres")
	}

	code := normalizePersonalWalletCode(params.Code)
	displayName := personalWalletDisplayName(code)
	if displayName == "" {
		return AdminPersonalWalletView{}, ErrInvalidPersonalWalletCode
	}

	address, err := normalizePersonalWalletAddress(params.WalletAddress)
	if err != nil {
		return AdminPersonalWalletView{}, err
	}

	now := normalizedNow(params.Now)
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return AdminPersonalWalletView{}, fmt.Errorf("store: begin update personal wallet: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `
INSERT INTO admin_personal_wallets (code, display_name, created_at, updated_at)
VALUES ($1, $2, $3, $3)
ON CONFLICT (code) DO NOTHING
`, code, displayName, now); err != nil {
		return AdminPersonalWalletView{}, fmt.Errorf("store: ensure personal wallet row: %w", err)
	}

	var previousAddress string
	if err := tx.QueryRow(ctx, `
SELECT wallet_address
FROM admin_personal_wallets
WHERE code = $1
FOR UPDATE
`, code).Scan(&previousAddress); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return AdminPersonalWalletView{}, ErrInvalidPersonalWalletCode
		}
		return AdminPersonalWalletView{}, fmt.Errorf("store: read personal wallet: %w", err)
	}

	var wallet AdminPersonalWalletView
	if err := tx.QueryRow(ctx, `
UPDATE admin_personal_wallets
SET wallet_address = $2,
    updated_by = NULLIF($3, '')::uuid,
    updated_at = $4
WHERE code = $1
RETURNING
  code,
  display_name,
  wallet_address,
  CASE code WHEN 'viksry' THEN '0.10' WHEN 'aryanto_hong' THEN '0.90' ELSE '0' END,
  COALESCE((SELECT SUM(amount)::text FROM admin_wallet_commission_ledger WHERE wallet_code = $1), '0'),
  updated_by::text,
  created_at,
  updated_at
`, code, address, strings.TrimSpace(params.AdminID), now).Scan(
		&wallet.Code,
		&wallet.DisplayName,
		&wallet.WalletAddress,
		&wallet.ShareRate,
		&wallet.Balance,
		&wallet.UpdatedBy,
		&wallet.CreatedAt,
		&wallet.UpdatedAt,
	); err != nil {
		return AdminPersonalWalletView{}, fmt.Errorf("store: update personal wallet: %w", err)
	}
	wallet.CommissionBalance = wallet.Balance

	if err := insertPersonalWalletAudit(ctx, tx, params.AdminID, previousAddress, wallet, now); err != nil {
		return AdminPersonalWalletView{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return AdminPersonalWalletView{}, fmt.Errorf("store: commit update personal wallet: %w", err)
	}
	return wallet, nil
}

func normalizePersonalWalletCode(value string) string {
	code := strings.ToLower(strings.TrimSpace(value))
	code = strings.ReplaceAll(code, "-", "_")
	switch code {
	case "viksry", "aryanto_hong":
		return code
	default:
		return ""
	}
}

func personalWalletDisplayName(code string) string {
	switch code {
	case "viksry":
		return "WALLET VIKSRY"
	case "aryanto_hong":
		return "WALLET ARYANTO HONG"
	default:
		return ""
	}
}

func normalizePersonalWalletAddress(value string) (string, error) {
	address := strings.ToLower(strings.TrimSpace(value))
	if address == "" {
		return "", nil
	}
	if len(address) != 42 || !strings.HasPrefix(address, "0x") {
		return "", ErrInvalidPersonalWalletAddress
	}
	for _, char := range address[2:] {
		if (char < '0' || char > '9') && (char < 'a' || char > 'f') {
			return "", ErrInvalidPersonalWalletAddress
		}
	}
	return address, nil
}

func insertPersonalWalletAudit(ctx context.Context, tx pgx.Tx, adminID string, previousAddress string, wallet AdminPersonalWalletView, now time.Time) error {
	auditID, err := id.New()
	if err != nil {
		return err
	}
	beforeJSON, err := json.Marshal(map[string]any{
		"code":          wallet.Code,
		"walletAddress": previousAddress,
	})
	if err != nil {
		return fmt.Errorf("store: marshal personal wallet audit before: %w", err)
	}
	afterJSON, err := json.Marshal(map[string]any{
		"code":              wallet.Code,
		"displayName":       wallet.DisplayName,
		"walletAddress":     wallet.WalletAddress,
		"shareRate":         wallet.ShareRate,
		"commissionBalance": wallet.CommissionBalance,
	})
	if err != nil {
		return fmt.Errorf("store: marshal personal wallet audit after: %w", err)
	}

	var actorID any
	if strings.TrimSpace(adminID) != "" {
		actorID = strings.TrimSpace(adminID)
	}

	if _, err := tx.Exec(ctx, `
INSERT INTO audit_logs (
  id, actor_type, actor_id, action, entity, before_state, after_state, created_at
) VALUES (
  $1::uuid, 'admin', $2::uuid, 'admin_personal_wallet_address_updated', 'admin_personal_wallet', $3::jsonb, $4::jsonb, $5
)`, auditID.String(), actorID, string(beforeJSON), string(afterJSON), now); err != nil {
		return fmt.Errorf("store: insert personal wallet audit: %w", err)
	}
	return nil
}
