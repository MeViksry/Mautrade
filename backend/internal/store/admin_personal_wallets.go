package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/MeViksry/Mautrade/backend/internal/domain/id"
	"github.com/MeViksry/qdecimal"
	"github.com/jackc/pgx/v5"
)

var (
	ErrInvalidPersonalWalletCode             = errors.New("store: invalid personal wallet code")
	ErrInvalidPersonalWalletAddress          = errors.New("store: invalid personal wallet address")
	ErrInvalidPersonalWalletWithdrawAmount   = errors.New("store: invalid personal wallet withdraw amount")
	ErrPersonalWalletWithdrawAddressRequired = errors.New("store: personal wallet withdraw address required")
	ErrPersonalWalletWithdrawInsufficient    = errors.New("store: personal wallet withdraw insufficient balance")
	ErrPersonalWalletWithdrawalNotFound      = errors.New("store: personal wallet withdrawal not found")
)

type AdminPersonalWalletView struct {
	Code                     string     `json:"code"`
	DisplayName              string     `json:"displayName"`
	WalletAddress            string     `json:"walletAddress"`
	ShareRate                string     `json:"shareRate"`
	Balance                  string     `json:"balance"`
	AvailableBalance         string     `json:"availableBalance"`
	CommissionBalance        string     `json:"commissionBalance"`
	PendingWithdrawalBalance string     `json:"pendingWithdrawalBalance"`
	WithdrawnBalance         string     `json:"withdrawnBalance"`
	DailyInflow              string     `json:"dailyInflow"`
	DailyOutflow             string     `json:"dailyOutflow"`
	CanManage                bool       `json:"canManage"`
	UpdatedBy                *string    `json:"updatedBy,omitempty"`
	CreatedAt                time.Time  `json:"createdAt"`
	UpdatedAt                *time.Time `json:"updatedAt,omitempty"`
}

type AdminPersonalWalletWithdrawalView struct {
	ID                 string    `json:"id"`
	WalletCode         string    `json:"walletCode"`
	DestinationAddress string    `json:"destinationAddress"`
	Amount             string    `json:"amount"`
	Asset              string    `json:"asset"`
	Status             string    `json:"status"`
	TxID               *string   `json:"txId,omitempty"`
	FailureReason      string    `json:"failureReason,omitempty"`
	RequestedAt        time.Time `json:"requestedAt"`
	UpdatedAt          time.Time `json:"updatedAt"`
}

type UpdateAdminPersonalWalletParams struct {
	Code          string
	WalletAddress string
	AdminID       string
	Now           time.Time
}

type CreateAdminPersonalWalletWithdrawalParams struct {
	Code          string
	WalletAddress string
	Amount        string
	AdminID       string
	Now           time.Time
}

type UpdateAdminPersonalWalletWithdrawalStatusParams struct {
	WithdrawalID string
	AdminID      string
	TxID         string
	Reason       string
	Now          time.Time
}

func (s *DashboardStore) AdminPersonalWallets(ctx context.Context) ([]AdminPersonalWalletView, error) {
	if !s.Ready() {
		return nil, fmt.Errorf("store: personal wallets require postgres")
	}

	rows, err := s.db.Query(ctx, `
WITH day_window AS (
  SELECT
    (((now() AT TIME ZONE 'Asia/Jakarta')::date)::timestamp AT TIME ZONE 'Asia/Jakarta') AS starts_at,
    ((((now() AT TIME ZONE 'Asia/Jakarta')::date + 1))::timestamp AT TIME ZONE 'Asia/Jakarta') AS ends_at
),
commission AS (
  SELECT wallet_code, COALESCE(SUM(amount), 0)::numeric(36,18) AS balance
  FROM admin_wallet_commission_ledger
  GROUP BY wallet_code
),
withdrawals AS (
  SELECT
    wallet_code,
    COALESCE(SUM(amount) FILTER (WHERE status IN ('pending_signing', 'broadcast')), 0)::numeric(36,18) AS pending_balance,
    COALESCE(SUM(amount) FILTER (WHERE status = 'confirmed'), 0)::numeric(36,18) AS withdrawn_balance
  FROM admin_wallet_withdrawals
  GROUP BY wallet_code
),
today_commission AS (
  SELECT wallet_code, COALESCE(SUM(amount), 0)::numeric(36,18) AS daily_inflow
  FROM admin_wallet_commission_ledger, day_window
  WHERE created_at >= day_window.starts_at
    AND created_at < day_window.ends_at
  GROUP BY wallet_code
),
today_withdrawals AS (
  SELECT wallet_code, COALESCE(SUM(amount), 0)::numeric(36,18) AS daily_outflow
  FROM admin_wallet_withdrawals, day_window
  WHERE status IN ('pending_signing', 'broadcast', 'confirmed')
    AND requested_at >= day_window.starts_at
    AND requested_at < day_window.ends_at
  GROUP BY wallet_code
)
SELECT
  w.code,
  w.display_name,
  w.wallet_address,
  CASE w.code WHEN 'viksry' THEN '0.10' WHEN 'aryanto_hong' THEN '0.90' ELSE '0' END AS share_rate,
  GREATEST(COALESCE(c.balance, 0) - COALESCE(x.pending_balance, 0) - COALESCE(x.withdrawn_balance, 0), 0)::text AS available_balance,
  COALESCE(c.balance, 0)::text AS commission_balance,
  COALESCE(x.pending_balance, 0)::text AS pending_withdrawal_balance,
  COALESCE(x.withdrawn_balance, 0)::text AS withdrawn_balance,
  COALESCE(tc.daily_inflow, 0)::text AS daily_inflow,
  COALESCE(tw.daily_outflow, 0)::text AS daily_outflow,
  w.updated_by::text,
  w.created_at,
  w.updated_at
FROM admin_personal_wallets w
LEFT JOIN commission c ON c.wallet_code = w.code
LEFT JOIN withdrawals x ON x.wallet_code = w.code
LEFT JOIN today_commission tc ON tc.wallet_code = w.code
LEFT JOIN today_withdrawals tw ON tw.wallet_code = w.code
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
			&wallet.CommissionBalance,
			&wallet.PendingWithdrawalBalance,
			&wallet.WithdrawnBalance,
			&wallet.DailyInflow,
			&wallet.DailyOutflow,
			&wallet.UpdatedBy,
			&wallet.CreatedAt,
			&wallet.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("store: scan admin personal wallet: %w", err)
		}
		wallet.AvailableBalance = wallet.Balance
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
  GREATEST(
    COALESCE((SELECT SUM(amount) FROM admin_wallet_commission_ledger WHERE wallet_code = $1), 0)
    - COALESCE((SELECT SUM(amount) FROM admin_wallet_withdrawals WHERE wallet_code = $1 AND status IN ('pending_signing', 'broadcast')), 0)
    - COALESCE((SELECT SUM(amount) FROM admin_wallet_withdrawals WHERE wallet_code = $1 AND status = 'confirmed'), 0),
    0
  )::text,
  COALESCE((SELECT SUM(amount)::text FROM admin_wallet_commission_ledger WHERE wallet_code = $1), '0'),
  COALESCE((SELECT SUM(amount)::text FROM admin_wallet_withdrawals WHERE wallet_code = $1 AND status IN ('pending_signing', 'broadcast')), '0'),
  COALESCE((SELECT SUM(amount)::text FROM admin_wallet_withdrawals WHERE wallet_code = $1 AND status = 'confirmed'), '0'),
  COALESCE((
    SELECT SUM(amount)::text
    FROM admin_wallet_commission_ledger
    WHERE wallet_code = $1
      AND created_at >= (((now() AT TIME ZONE 'Asia/Jakarta')::date)::timestamp AT TIME ZONE 'Asia/Jakarta')
      AND created_at < ((((now() AT TIME ZONE 'Asia/Jakarta')::date + 1))::timestamp AT TIME ZONE 'Asia/Jakarta')
  ), '0'),
  COALESCE((
    SELECT SUM(amount)::text
    FROM admin_wallet_withdrawals
    WHERE wallet_code = $1
      AND status IN ('pending_signing', 'broadcast', 'confirmed')
      AND requested_at >= (((now() AT TIME ZONE 'Asia/Jakarta')::date)::timestamp AT TIME ZONE 'Asia/Jakarta')
      AND requested_at < ((((now() AT TIME ZONE 'Asia/Jakarta')::date + 1))::timestamp AT TIME ZONE 'Asia/Jakarta')
  ), '0'),
  updated_by::text,
  created_at,
  updated_at
`, code, address, strings.TrimSpace(params.AdminID), now).Scan(
		&wallet.Code,
		&wallet.DisplayName,
		&wallet.WalletAddress,
		&wallet.ShareRate,
		&wallet.Balance,
		&wallet.CommissionBalance,
		&wallet.PendingWithdrawalBalance,
		&wallet.WithdrawnBalance,
		&wallet.DailyInflow,
		&wallet.DailyOutflow,
		&wallet.UpdatedBy,
		&wallet.CreatedAt,
		&wallet.UpdatedAt,
	); err != nil {
		return AdminPersonalWalletView{}, fmt.Errorf("store: update personal wallet: %w", err)
	}
	wallet.AvailableBalance = wallet.Balance

	if err := insertPersonalWalletAudit(ctx, tx, params.AdminID, previousAddress, wallet, now); err != nil {
		return AdminPersonalWalletView{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return AdminPersonalWalletView{}, fmt.Errorf("store: commit update personal wallet: %w", err)
	}
	return wallet, nil
}

func (s *DashboardStore) AdminCreatePersonalWalletWithdrawal(ctx context.Context, params CreateAdminPersonalWalletWithdrawalParams) (AdminPersonalWalletWithdrawalView, error) {
	if !s.Ready() {
		return AdminPersonalWalletWithdrawalView{}, fmt.Errorf("store: create personal wallet withdrawal requires postgres")
	}

	code := normalizePersonalWalletCode(params.Code)
	displayName := personalWalletDisplayName(code)
	if displayName == "" {
		return AdminPersonalWalletWithdrawalView{}, ErrInvalidPersonalWalletCode
	}

	amount, err := parsePersonalWalletWithdrawalAmount(params.Amount)
	if err != nil {
		return AdminPersonalWalletWithdrawalView{}, err
	}

	requestedAddress, err := normalizePersonalWalletAddress(params.WalletAddress)
	if err != nil {
		return AdminPersonalWalletWithdrawalView{}, err
	}

	now := normalizedNow(params.Now)
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return AdminPersonalWalletWithdrawalView{}, fmt.Errorf("store: begin personal wallet withdrawal: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `
INSERT INTO admin_personal_wallets (code, display_name, created_at, updated_at)
VALUES ($1, $2, $3, $3)
ON CONFLICT (code) DO NOTHING
`, code, displayName, now); err != nil {
		return AdminPersonalWalletWithdrawalView{}, fmt.Errorf("store: ensure personal wallet row: %w", err)
	}

	var linkedAddress string
	var availableBalanceText string
	if err := tx.QueryRow(ctx, `
WITH locked_wallet AS (
  SELECT code, wallet_address
  FROM admin_personal_wallets
  WHERE code = $1
  FOR UPDATE
),
commission AS (
  SELECT COALESCE(SUM(amount), 0)::numeric(36,18) AS balance
  FROM admin_wallet_commission_ledger
  WHERE wallet_code = $1
),
withdrawals AS (
  SELECT
    COALESCE(SUM(amount) FILTER (WHERE status IN ('pending_signing', 'broadcast')), 0)::numeric(36,18) AS pending_balance,
    COALESCE(SUM(amount) FILTER (WHERE status = 'confirmed'), 0)::numeric(36,18) AS withdrawn_balance
  FROM admin_wallet_withdrawals
  WHERE wallet_code = $1
)
SELECT
  w.wallet_address,
  GREATEST(c.balance - x.pending_balance - x.withdrawn_balance, 0)::text AS available_balance
FROM locked_wallet w
CROSS JOIN commission c
CROSS JOIN withdrawals x
`, code).Scan(&linkedAddress, &availableBalanceText); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return AdminPersonalWalletWithdrawalView{}, ErrInvalidPersonalWalletCode
		}
		return AdminPersonalWalletWithdrawalView{}, fmt.Errorf("store: read personal wallet withdrawal balance: %w", err)
	}

	destinationAddress := requestedAddress
	if destinationAddress == "" {
		destinationAddress = linkedAddress
	}
	if destinationAddress == "" {
		return AdminPersonalWalletWithdrawalView{}, ErrPersonalWalletWithdrawAddressRequired
	}

	availableBalance, err := qdecimal.Parse(decimalOrZero(availableBalanceText))
	if err != nil {
		return AdminPersonalWalletWithdrawalView{}, fmt.Errorf("store: parse personal wallet available balance: %w", err)
	}
	if amount.Cmp(availableBalance) > 0 {
		return AdminPersonalWalletWithdrawalView{}, ErrPersonalWalletWithdrawInsufficient
	}

	withdrawalID, err := newUUIDText()
	if err != nil {
		return AdminPersonalWalletWithdrawalView{}, err
	}

	var withdrawal AdminPersonalWalletWithdrawalView
	if err := tx.QueryRow(ctx, `
INSERT INTO admin_wallet_withdrawals (
  id, wallet_code, admin_id, destination_address, amount, asset, status, requested_at, updated_at
) VALUES (
  $1::uuid, $2, NULLIF($3, '')::uuid, $4, $5::numeric, 'USDT', 'pending_signing', $6, $6
)
RETURNING id::text, wallet_code, destination_address, amount::text, asset, status, tx_id, failure_reason, requested_at, updated_at
`, withdrawalID, code, strings.TrimSpace(params.AdminID), destinationAddress, amount.String(), now).Scan(
		&withdrawal.ID,
		&withdrawal.WalletCode,
		&withdrawal.DestinationAddress,
		&withdrawal.Amount,
		&withdrawal.Asset,
		&withdrawal.Status,
		&withdrawal.TxID,
		&withdrawal.FailureReason,
		&withdrawal.RequestedAt,
		&withdrawal.UpdatedAt,
	); err != nil {
		return AdminPersonalWalletWithdrawalView{}, fmt.Errorf("store: insert personal wallet withdrawal: %w", err)
	}

	if err := insertPersonalWalletWithdrawalAudit(ctx, tx, params.AdminID, withdrawal, availableBalanceText, now); err != nil {
		return AdminPersonalWalletWithdrawalView{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return AdminPersonalWalletWithdrawalView{}, fmt.Errorf("store: commit personal wallet withdrawal: %w", err)
	}
	return withdrawal, nil
}

func (s *DashboardStore) AdminMarkPersonalWalletWithdrawalBroadcast(ctx context.Context, params UpdateAdminPersonalWalletWithdrawalStatusParams) (AdminPersonalWalletWithdrawalView, error) {
	if !s.Ready() {
		return AdminPersonalWalletWithdrawalView{}, fmt.Errorf("store: update personal wallet withdrawal requires postgres")
	}

	txID, err := normalizeWithdrawalTxID(params.TxID)
	if err != nil {
		return AdminPersonalWalletWithdrawalView{}, err
	}

	now := normalizedNow(params.Now)
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return AdminPersonalWalletWithdrawalView{}, fmt.Errorf("store: begin broadcast personal wallet withdrawal: %w", err)
	}
	defer tx.Rollback(ctx)

	withdrawal, err := updatePersonalWalletWithdrawalStatus(ctx, tx, params.WithdrawalID, "broadcast", txID, "", now)
	if err != nil {
		return AdminPersonalWalletWithdrawalView{}, err
	}
	if err := insertPersonalWalletWithdrawalStatusAudit(ctx, tx, params.AdminID, "admin_personal_wallet_withdraw_broadcast", withdrawal, now); err != nil {
		return AdminPersonalWalletWithdrawalView{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return AdminPersonalWalletWithdrawalView{}, fmt.Errorf("store: commit broadcast personal wallet withdrawal: %w", err)
	}
	return withdrawal, nil
}

func (s *DashboardStore) AdminMarkPersonalWalletWithdrawalFailed(ctx context.Context, params UpdateAdminPersonalWalletWithdrawalStatusParams) (AdminPersonalWalletWithdrawalView, error) {
	if !s.Ready() {
		return AdminPersonalWalletWithdrawalView{}, fmt.Errorf("store: update personal wallet withdrawal requires postgres")
	}

	now := normalizedNow(params.Now)
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return AdminPersonalWalletWithdrawalView{}, fmt.Errorf("store: begin fail personal wallet withdrawal: %w", err)
	}
	defer tx.Rollback(ctx)

	withdrawal, err := updatePersonalWalletWithdrawalStatus(ctx, tx, params.WithdrawalID, "failed", "", safeWithdrawalFailureReason(params.Reason), now)
	if err != nil {
		return AdminPersonalWalletWithdrawalView{}, err
	}
	if err := insertPersonalWalletWithdrawalStatusAudit(ctx, tx, params.AdminID, "admin_personal_wallet_withdraw_failed", withdrawal, now); err != nil {
		return AdminPersonalWalletWithdrawalView{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return AdminPersonalWalletWithdrawalView{}, fmt.Errorf("store: commit fail personal wallet withdrawal: %w", err)
	}
	return withdrawal, nil
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

func parsePersonalWalletWithdrawalAmount(value string) (qdecimal.Decimal, error) {
	amountText := strings.TrimSpace(value)
	if amountText == "" {
		return qdecimal.Decimal{}, ErrInvalidPersonalWalletWithdrawAmount
	}
	amountText = strings.TrimPrefix(amountText, "+")

	parts := strings.Split(amountText, ".")
	if len(parts) > 2 || parts[0] == "" {
		return qdecimal.Decimal{}, ErrInvalidPersonalWalletWithdrawAmount
	}
	for _, char := range parts[0] {
		if char < '0' || char > '9' {
			return qdecimal.Decimal{}, ErrInvalidPersonalWalletWithdrawAmount
		}
	}
	if len(parts) == 2 {
		if len(parts[1]) == 0 || len(parts[1]) > 18 {
			return qdecimal.Decimal{}, ErrInvalidPersonalWalletWithdrawAmount
		}
		for _, char := range parts[1] {
			if char < '0' || char > '9' {
				return qdecimal.Decimal{}, ErrInvalidPersonalWalletWithdrawAmount
			}
		}
	}

	amount, err := qdecimal.Parse(amountText)
	if err != nil || amount.Sign() <= 0 {
		return qdecimal.Decimal{}, ErrInvalidPersonalWalletWithdrawAmount
	}
	return amount, nil
}

func normalizeWithdrawalTxID(value string) (string, error) {
	txID := strings.ToLower(strings.TrimSpace(value))
	if txID == "" {
		return "", fmt.Errorf("%w: tx_id is required", ErrInvalidPersonalWalletWithdrawAmount)
	}
	if !strings.HasPrefix(txID, "0x") {
		txID = "0x" + txID
	}
	if len(txID) != 66 {
		return "", fmt.Errorf("%w: tx_id must be 32-byte hex hash", ErrInvalidPersonalWalletWithdrawAmount)
	}
	for _, char := range txID[2:] {
		if !isHexRune(char) {
			return "", fmt.Errorf("%w: tx_id must be hex", ErrInvalidPersonalWalletWithdrawAmount)
		}
	}
	return txID, nil
}

func safeWithdrawalFailureReason(value string) string {
	reason := strings.TrimSpace(value)
	if reason == "" {
		return "broadcast failed"
	}
	if len(reason) > 700 {
		return reason[:700]
	}
	return reason
}

func updatePersonalWalletWithdrawalStatus(ctx context.Context, tx pgx.Tx, withdrawalID string, status string, txID string, reason string, now time.Time) (AdminPersonalWalletWithdrawalView, error) {
	var withdrawal AdminPersonalWalletWithdrawalView
	if err := tx.QueryRow(ctx, `
UPDATE admin_wallet_withdrawals
SET status = $2,
    tx_id = NULLIF($3, ''),
    failure_reason = $4,
    broadcast_at = CASE WHEN $2 = 'broadcast' THEN $5 ELSE broadcast_at END,
    updated_at = $5
WHERE id = $1::uuid
  AND status = 'pending_signing'
RETURNING id::text, wallet_code, destination_address, amount::text, asset, status, tx_id, failure_reason, requested_at, updated_at
`, strings.TrimSpace(withdrawalID), status, txID, reason, now).Scan(
		&withdrawal.ID,
		&withdrawal.WalletCode,
		&withdrawal.DestinationAddress,
		&withdrawal.Amount,
		&withdrawal.Asset,
		&withdrawal.Status,
		&withdrawal.TxID,
		&withdrawal.FailureReason,
		&withdrawal.RequestedAt,
		&withdrawal.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return AdminPersonalWalletWithdrawalView{}, ErrPersonalWalletWithdrawalNotFound
		}
		return AdminPersonalWalletWithdrawalView{}, fmt.Errorf("store: update personal wallet withdrawal status: %w", err)
	}
	return withdrawal, nil
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

func insertPersonalWalletWithdrawalStatusAudit(ctx context.Context, tx pgx.Tx, adminID string, action string, withdrawal AdminPersonalWalletWithdrawalView, now time.Time) error {
	auditID, err := newUUIDText()
	if err != nil {
		return err
	}
	afterJSON, err := json.Marshal(map[string]any{
		"id":                 withdrawal.ID,
		"walletCode":         withdrawal.WalletCode,
		"destinationAddress": withdrawal.DestinationAddress,
		"amount":             withdrawal.Amount,
		"asset":              withdrawal.Asset,
		"status":             withdrawal.Status,
		"txId":               withdrawal.TxID,
		"failureReason":      withdrawal.FailureReason,
	})
	if err != nil {
		return fmt.Errorf("store: marshal personal wallet withdrawal status audit: %w", err)
	}
	var actorID any
	if strings.TrimSpace(adminID) != "" {
		actorID = strings.TrimSpace(adminID)
	}

	if _, err := tx.Exec(ctx, `
INSERT INTO audit_logs (
  id, actor_type, actor_id, action, entity, entity_id, after_state, created_at
) VALUES (
  $1::uuid, 'admin', $2::uuid, $3, 'admin_personal_wallet_withdrawal', $4::uuid, $5::jsonb, $6
)`, auditID, actorID, action, withdrawal.ID, string(afterJSON), now); err != nil {
		return fmt.Errorf("store: insert personal wallet withdrawal status audit: %w", err)
	}
	return nil
}

func insertPersonalWalletWithdrawalAudit(ctx context.Context, tx pgx.Tx, adminID string, withdrawal AdminPersonalWalletWithdrawalView, availableBefore string, now time.Time) error {
	auditID, err := newUUIDText()
	if err != nil {
		return err
	}
	afterJSON, err := json.Marshal(map[string]any{
		"id":                 withdrawal.ID,
		"walletCode":         withdrawal.WalletCode,
		"destinationAddress": withdrawal.DestinationAddress,
		"amount":             withdrawal.Amount,
		"asset":              withdrawal.Asset,
		"status":             withdrawal.Status,
		"availableBefore":    availableBefore,
	})
	if err != nil {
		return fmt.Errorf("store: marshal personal wallet withdrawal audit: %w", err)
	}
	var actorID any
	if strings.TrimSpace(adminID) != "" {
		actorID = strings.TrimSpace(adminID)
	}

	if _, err := tx.Exec(ctx, `
INSERT INTO audit_logs (
  id, actor_type, actor_id, action, entity, entity_id, after_state, created_at
) VALUES (
  $1::uuid, 'admin', $2::uuid, 'admin_personal_wallet_withdraw_requested', 'admin_personal_wallet_withdrawal', $3::uuid, $4::jsonb, $5
)`, auditID, actorID, withdrawal.ID, string(afterJSON), now); err != nil {
		return fmt.Errorf("store: insert personal wallet withdrawal audit: %w", err)
	}
	return nil
}
